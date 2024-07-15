package compact

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"sort"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/encoding"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/geometry"
	"diagonal.works/b6/ingest"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/search"
	"github.com/golang/geo/s2"
	"github.com/golang/groupcache/lru"
	"golang.org/x/mod/semver"
)

type featureBlock struct {
	FeatureBlock
	Strings        *encoding.StringTable
	NamespaceTable *NamespaceTable
}

const FeaturesByIDCacheSize = 4000 // Empirically the working-set size of Galashiels

type FeaturesByID struct {
	features [b6.FeatureTypeEnd][]*featureBlock
	base     b6.FeaturesByID
	cache    *lru.Cache
	lock     sync.Mutex
}

func NewFeaturesByID(base b6.FeaturesByID) *FeaturesByID {
	f := &FeaturesByID{base: base, cache: lru.New(FeaturesByIDCacheSize)}
	for i := range f.features {
		f.features[i] = make([]*featureBlock, 0, 1)
	}
	return f
}

func NewFeaturesByIDFromData(data []byte, base b6.FeaturesByID) (*FeaturesByID, error) {
	f := NewFeaturesByID(base)
	return f, f.Merge(data)
}

func verifyVersion(header *Header, buffer []byte) error {
	v := header.UnmarshalVersion(buffer)
	if !semver.IsValid("v" + v) {
		return fmt.Errorf("invalid index version %s", v)
	}
	if semver.Major("v"+v) != semver.Major("v"+Version) {
		return fmt.Errorf("need index version %s, found %s", Version, v)
	}
	return nil
}

func (f *FeaturesByID) Merge(data []byte) error {
	var header Header
	header.Unmarshal(data)
	if header.Magic != HeaderMagic {
		return fmt.Errorf("Bad header magic")
	}
	if err := verifyVersion(&header, data); err != nil {
		return err
	}
	strings := encoding.NewStringTable(data[header.StringsOffset:])

	var hp pb.CompactHeaderProto
	if err := UnmarshalProto(data[header.HeaderProtoOffset:], &hp); err != nil {
		return err
	}
	var nt NamespaceTable
	nt.FillFromProto(&hp)

	offset := header.BlockOffset
	var block BlockHeader
	for offset < encoding.Offset(len(data)) {
		offset += encoding.Offset(block.Unmarshal(data[offset:]))
		if block.Type == BlockTypeFeatures {
			fb := &featureBlock{Strings: strings, NamespaceTable: &nt}
			fb.Unmarshal(data[offset:])
			f.features[fb.FeatureType] = append(f.features[fb.FeatureType], fb)
		}
		offset += encoding.Offset(block.Length)
	}
	return nil
}

func (f *FeaturesByID) FindFeatureByID(id b6.FeatureID) b6.Feature {
	if int(id.Type) >= len(f.features) {
		return nil
	}
	cacheable := id.Type == b6.FeatureTypePath || id.Type == b6.FeatureTypeArea
	if cacheable {
		f.lock.Lock()
		if hit, ok := f.cache.Get(id); ok {
			f.lock.Unlock()
			return hit.(b6.Feature)
		} else {
			f.lock.Unlock()
		}
	}
	feature := f.findWithoutCache(id)
	if feature != nil && cacheable {
		f.lock.Lock()
		f.cache.Add(id, feature)
		f.lock.Unlock()
	}
	return feature
}

func (f *FeaturesByID) findWithoutCache(id b6.FeatureID) b6.Feature {
	if int(id.Type) >= len(f.features) {
		return nil
	}
	for _, fb := range f.features[id.Type] {
		if ns, ok := fb.NamespaceTable.MaybeEncode(id.Namespace); ok && ns == fb.Namespaces[id.Type] {
			switch id.Type {
			case b6.FeatureTypePoint:
				if p := f.newPhysicalFeature(fb, b6.FeatureTypePoint, id.Value); p != nil {
					return p
				}
			case b6.FeatureTypePath:
				if p := f.newWrappedPhysicalFeature(fb, id.Value); p != nil {
					return p
				}
			case b6.FeatureTypeArea:
				if a := f.newArea(fb, id.Value); a != nil {
					return a
				}
			case b6.FeatureTypeRelation:
				if r := f.newRelation(fb, id.Value); r != nil {
					return r
				}
			}
		}
	}
	return f.base.FindFeatureByID(id)
}

func (f *FeaturesByID) HasFeatureWithID(id b6.FeatureID) bool {
	if int(id.Type) >= len(f.features) {
		return false
	}
	return hasFeatureWithID(id, f.features[id.Type]) || f.base.HasFeatureWithID(id)
}

func hasFeatureWithID(id b6.FeatureID, fbs []*featureBlock) bool {
	for _, fb := range fbs {
		ns, ok := fb.NamespaceTable.MaybeEncode(id.Namespace)
		if ok && ns == fb.Namespaces[id.Type] {
			_, ok := fb.Map.FindFirst(id.Value)
			return ok
		}
	}
	return false
}

func (f *FeaturesByID) FindLocationByID(id b6.FeatureID) (s2.LatLng, error) {
	for _, fb := range f.features[b6.FeatureTypePoint] {
		if ns, ok := fb.NamespaceTable.MaybeEncode(id.Namespace); ok && ns == fb.Namespaces[b6.FeatureTypePoint] {
			if t, ok := fb.Map.FindFirst(id.Value); ok {
				if t.Tag != PointTagReferencesOnly {
					point := MarshalledTags{t.Data, fb.Strings, fb.NamespaceTable, TypeAndNamespaceInvalid}.Point()
					if point.Norm() != 0 {
						return s2.LatLngFromPoint(point), nil
					}
				}
			}
		}
	}
	return f.base.FindLocationByID(id)
}

func (f *FeaturesByID) newPhysicalFeature(fb *featureBlock, typ b6.FeatureType, id uint64) b6.Feature {
	t, ok := fb.Map.FindFirst(id)
	if ok {
		return f.newPhysicalFeatureFromTagged(fb, b6.FeatureID{typ, fb.NamespaceTable.Decode(fb.Namespaces[typ]), id}, t)
	}
	return nil
}

func (f *FeaturesByID) Namespaces() []b6.Namespace {
	nss := make(map[b6.Namespace]struct{})
	for _, fbs := range f.features {
		for _, fb := range fbs {
			for _, ns := range fb.NamespaceTable.FromEncoded {
				nss[ns] = struct{}{}
			}
		}
	}
	sorted := make(b6.Namespaces, 0, len(nss))
	for ns := range nss {
		sorted = append(sorted, ns)
	}
	sort.Sort(sorted)
	return sorted
}

func (f *FeaturesByID) FillNamespaceTable(nt *NamespaceTable) error {
	if nt.ToEncoded == nil {
		nt.ToEncoded = make(map[b6.Namespace]Namespace)
	}
	for _, fbs := range f.features {
		for _, fb := range fbs {
			for i, ns := range fb.NamespaceTable.FromEncoded {
				for len(nt.FromEncoded) < i+1 {
					nt.FromEncoded = append(nt.FromEncoded, b6.NamespaceInvalid)
				}
				if nt.FromEncoded[i] == b6.NamespaceInvalid {
					nt.FromEncoded[i] = ns
					nt.ToEncoded[ns] = Namespace(i)
				} else if nt.FromEncoded[i] != ns {
					return fmt.Errorf("duplicate namespace for namespace %d: %s vs %s", i, nt.FromEncoded[i], ns)
				}
			}
		}
	}
	return nil
}

type marshalledPhysicalFeature struct {
	b6.ID
	MarshalledTags
}

func (m marshalledPhysicalFeature) FeatureID() b6.FeatureID {
	return m.ID
}

func (m marshalledPhysicalFeature) ToGeoJSON() geojson.GeoJSON {
	return b6.PhysicalFeatureToGeoJSON(m)
}

func (f *FeaturesByID) newPhysicalFeatureFromTagged(fb *featureBlock, id b6.FeatureID, t encoding.Tagged) b6.PhysicalFeature {
	if t.Tag != PointTagReferencesOnly {
		return marshalledPhysicalFeature{id, MarshalledTags{Tags: t.Data, Strings: fb.Strings}}
	}
	return nil
}

type wrappedMarshalledPhysicalFeature struct {
	marshalledPhysicalFeature
	byID *FeaturesByID

	lock     sync.Mutex
	polyline s2.Polyline // TODO: avoid caching
}

func (m *wrappedMarshalledPhysicalFeature) PointAt(i int) s2.Point {
	if id := m.Reference(i).Source(); id.IsValid() {
		var err error
		ll, err := m.byID.FindLocationByID(id)
		if err != nil {
			panic(fmt.Sprintf("Missing point %s", id))
		}
		return s2.PointFromLatLng(ll)
	} else {
		return m.MarshalledTags.PointAt(i)
	}
}

func (m *wrappedMarshalledPhysicalFeature) Polyline() *s2.Polyline {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.polyline == nil {
		m.polyline = make(s2.Polyline, m.GeometryLen())
		for i := 0; i < m.GeometryLen(); i++ {
			m.polyline[i] = m.PointAt(i)
		}
	}
	return &m.polyline
}

func (m *wrappedMarshalledPhysicalFeature) Feature(i int) b6.PhysicalFeature {
	id := m.Reference(i).Source()
	if p, ok := m.byID.FindFeatureByID(id).(b6.PhysicalFeature); ok {
		return p
	}
	panic("not a physical feature")
}

func (f *FeaturesByID) newWrappedPhysicalFeature(fb *featureBlock, id uint64) b6.PhysicalFeature {
	b := fb.Map.FindFirstWithTag(id, encoding.NoTag)
	if b != nil {
		return f.newWrappedPhysicalFeatureFromBuffer(fb, id, b)
	}
	return nil
}

func (f *FeaturesByID) newWrappedPhysicalFeatureFromBuffer(fb *featureBlock, id uint64, buffer []byte) b6.PhysicalFeature {
	return &wrappedMarshalledPhysicalFeature{
		marshalledPhysicalFeature: marshalledPhysicalFeature{
			ID:             b6.FeatureID{b6.FeatureTypePath, fb.NamespaceTable.Decode(fb.Namespaces[b6.FeatureTypePath]), id},
			MarshalledTags: MarshalledTags{buffer, fb.Strings, fb.NamespaceTable, CombineTypeAndNamespace(b6.FeatureTypePoint, fb.NamespaceTable.Encode(b6.NamespaceOSMNode))},
		},
		byID: f,
	}
}

func (f *FeaturesByID) newPathFromEncodedPath(fb *featureBlock, id uint64, p *Path) *ingest.GenericFeature {
	return &ingest.GenericFeature{
		ID:   b6.FeatureID{b6.FeatureTypePath, fb.NamespaceTable.Decode(fb.Namespaces[b6.FeatureTypePath]), id},
		Tags: toTags(p.Tags, fb.Strings, fb.NamespaceTable),
	}
}

type marshalledArea struct {
	id       b6.AreaID
	area     MarshalledArea
	geometry AreaGeometry
	polygons []*s2.Polygon
	fb       *featureBlock
	byID     *FeaturesByID
	lock     sync.Mutex
}

func (m *marshalledArea) FeatureID() b6.FeatureID {
	return m.id.FeatureID()
}

func (m *marshalledArea) AreaID() b6.AreaID {
	return m.id
}

func (m *marshalledArea) AllTags() b6.Tags {
	return m.area.Tags(m.fb.Strings).AllTags()
}

func (m *marshalledArea) Get(key string) b6.Tag {
	return m.area.Tags(m.fb.Strings).Get(key)
}

func (m *marshalledArea) Len() int {
	return m.area.Len()
}

func (m *marshalledArea) Polygon(i int) *s2.Polygon {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.fillGeometry()
	if m.polygons[i] == nil {
		if p, ok := m.geometry.Polygon(i); ok {
			m.polygons[i] = p
		} else {
			paths := m.featureWithLock(i)
			loops := make([]*s2.Loop, len(paths))
			for j, path := range paths {
				points := *path.Polyline()
				loops[j] = s2.LoopFromPoints(points[0 : len(points)-1])
			}
			m.polygons[i] = s2.PolygonFromLoops(loops)
		}
	}
	return m.polygons[i]
}

func (m *marshalledArea) MultiPolygon() geometry.MultiPolygon {
	mp := make(geometry.MultiPolygon, m.Len())
	for i := 0; i < m.Len(); i++ {
		mp[i] = m.Polygon(i)
	}
	return mp
}

func (m *marshalledArea) Feature(i int) []b6.PhysicalFeature {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.featureWithLock(i)
}

func (m *marshalledArea) featureWithLock(i int) []b6.PhysicalFeature {
	m.fillGeometry()
	var paths []b6.PhysicalFeature
	if ids, ok := m.geometry.PathIDs(i); ok {
		paths = make([]b6.PhysicalFeature, len(ids))
		for i := range ids {
			typ, ns := ids[i].TypeAndNamespace.Split()
			id := b6.FeatureID{typ, m.fb.NamespaceTable.Decode(ns), ids[i].Value}
			if path := m.byID.FindFeatureByID(id).(b6.PhysicalFeature); path != nil {
				paths[i] = path
			} else {
				panic(fmt.Sprintf("Missing path %s for %s", id, m.id))
			}
		}
	}
	return paths
}

func (m *marshalledArea) fillGeometry() {
	if m.geometry == nil {
		m.geometry = m.area.UnmarshalPolygons(CombineTypeAndNamespace(b6.FeatureTypePath, m.fb.Namespaces[b6.FeatureTypePath]))
		m.polygons = make([]*s2.Polygon, m.geometry.Len())
	}
}

func (m *marshalledArea) ToGeoJSON() geojson.GeoJSON {
	return b6.AreaFeatureToGeoJSON(m)
}

func (m marshalledArea) References() []b6.Reference {
	panic("not implemented")
}

func (m marshalledArea) Reference(i int) b6.Reference {
	panic("not implemented")
}

func (m marshalledArea) Polyline() *s2.Polyline {
	panic("not implemented")
}

func (m marshalledArea) GeometryLen() int {
	panic("not implemented")
}

func (m marshalledArea) PointAt(i int) s2.Point {
	panic("not implemented")
}

func (m *marshalledArea) GeometryType() b6.GeometryType {
	return b6.GeometryTypeArea
}

func (m *marshalledArea) Point() s2.Point {
	return s2.Point{}
}

func (f *FeaturesByID) newArea(fb *featureBlock, id uint64) b6.AreaFeature {
	if b := fb.Map.FindFirstWithTag(id, encoding.NoTag); len(b) > 0 {
		return f.newAreaFromBuffer(fb, id, b)
	}
	return nil
}

func (f *FeaturesByID) newAreaFromBuffer(fb *featureBlock, id uint64, b []byte) b6.AreaFeature {
	return &marshalledArea{
		id:   b6.MakeAreaID(fb.NamespaceTable.Decode(fb.Namespaces[b6.FeatureTypeArea]), id),
		area: MarshalledArea(b),
		fb:   fb,
		byID: f,
	}
}

type marshalledRelation struct {
	id       b6.RelationID
	relation MarshalledRelation
	members  Members
	fb       *featureBlock
	byID     *FeaturesByID
}

func (m *marshalledRelation) FeatureID() b6.FeatureID {
	return m.id.FeatureID()
}

func (m *marshalledRelation) RelationID() b6.RelationID {
	return m.id
}

func (m *marshalledRelation) AllTags() b6.Tags {
	return m.relation.Tags(m.fb.Strings).AllTags()
}

func (m *marshalledRelation) Get(key string) b6.Tag {
	return m.relation.Tags(m.fb.Strings).Get(key)
}

func (m *marshalledRelation) Len() int {
	return m.relation.Len()
}

func (m *marshalledRelation) Member(i int) b6.RelationMember {
	m.fillMembers()
	typ, ns := m.members[i].ID.TypeAndNamespace.Split()
	return b6.RelationMember{
		Role: m.fb.Strings.Lookup(m.members[i].Role),
		ID:   b6.FeatureID{Type: typ, Namespace: m.fb.NamespaceTable.Decode(ns), Value: m.members[i].ID.Value},
	}
}

func (m *marshalledRelation) fillMembers() {
	if len(m.members) == 0 {
		m.relation.UnmarshalMembers(b6.FeatureTypePath, &m.fb.Namespaces, &m.members)
	}
}

func (m *marshalledRelation) ToGeoJSON() geojson.GeoJSON {
	return b6.RelationFeatureToGeoJSON(m, m.byID)
}

func (m marshalledRelation) References() []b6.Reference {
	panic("not implemented")
}

func (m marshalledRelation) Reference(i int) b6.Reference {
	panic("not implemented")
}

func (f *FeaturesByID) newRelation(fb *featureBlock, id uint64) b6.RelationFeature {
	b := fb.Map.FindFirstWithTag(id, encoding.NoTag)
	if b != nil {
		return f.newRelationFromBuffer(fb, id, b)
	}
	return nil
}

func (f *FeaturesByID) newRelationFromBuffer(fb *featureBlock, id uint64, b []byte) b6.RelationFeature {
	return &marshalledRelation{
		id:       b6.MakeRelationID(fb.NamespaceTable.Decode(fb.Namespaces[b6.FeatureTypeRelation]), id),
		relation: MarshalledRelation(b),
		fb:       fb,
		byID:     f,
	}
}

func (f *FeaturesByID) EachFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	goroutines := options.Goroutines
	if goroutines < 1 {
		goroutines = 1
	}

	if !options.SkipPoints {
		for _, fb := range f.features[b6.FeatureTypePoint] {
			emit := func(id uint64, tagged []encoding.Tagged, g int) error {
				if point := f.newPhysicalFeatureFromTagged(fb, b6.FeatureID{b6.FeatureTypePoint, fb.NamespaceTable.Decode(fb.Namespaces[b6.FeatureTypePoint]), id}, tagged[0]); point != nil {
					return each(point, g)
				}
				return nil
			}
			if err := fb.Map.EachItem(emit, goroutines); err != nil {
				return err
			}
		}
	}

	if !options.SkipPaths {
		for _, fb := range f.features[b6.FeatureTypePath] {
			emit := func(id uint64, tagged []encoding.Tagged, g int) error {
				return each(f.newWrappedPhysicalFeatureFromBuffer(fb, id, tagged[0].Data), g)
			}
			if err := fb.Map.EachItem(emit, goroutines); err != nil {
				return err
			}
		}
	}

	if !options.SkipAreas {
		for _, fb := range f.features[b6.FeatureTypeArea] {
			emit := func(id uint64, tagged []encoding.Tagged, g int) error {
				return each(f.newAreaFromBuffer(fb, id, tagged[0].Data), g)
			}
			if err := fb.Map.EachItem(emit, goroutines); err != nil {
				return err
			}
		}
	}

	if !options.SkipRelations {
		for _, fb := range f.features[b6.FeatureTypeRelation] {
			emit := func(id uint64, tagged []encoding.Tagged, g int) error {
				return each(f.newRelationFromBuffer(fb, id, tagged[0].Data), g)
			}
			if err := fb.Map.EachItem(emit, goroutines); err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *FeaturesByID) FindReferences(id b6.FeatureID, typed ...b6.FeatureType) b6.Features {
	// TODO(mari): provide an implementation that handles collections etc
	// when we move to a unified way of storing feature references in
	// the compact index
	features := make([]b6.Feature, 0)
	if id.Type == b6.FeatureTypePoint && (len(typed) == 0 || slices.Contains(typed, b6.FeatureTypePath)) {
		ids := f.findPathsByPoint(id, make([]b6.FeatureID, 0, 2))
		paths := make([]b6.Feature, 0, len(ids))
		for _, id := range ids {
			if path, ok := f.FindFeatureByID(id).(b6.Feature); ok {
				paths = append(paths, path)
			}
		}
		i := b6.NewFeatureIterator(paths)

		for i.Next() {
			features = append(features, i.Feature())
		}
	}
	if len(typed) == 0 || slices.Contains(typed, b6.FeatureTypeRelation) {
		i := f.FindRelationsByFeature(id)
		for i.Next() {
			features = append(features, i.Feature())
		}
	}
	if id.Type == b6.FeatureTypePoint && (len(typed) == 0 || slices.Contains(typed, b6.FeatureTypeArea)) {
		i := f.FindAreasByPoint(id)
		for i.Next() {
			features = append(features, i.Feature())
		}
	}
	return b6.NewFeatureIterator(features)
}

func (f *FeaturesByID) findPathsByPoint(id b6.FeatureID, paths []b6.FeatureID) []b6.FeatureID {
	for _, fb := range f.features[b6.FeatureTypePoint] {
		if ns, ok := fb.NamespaceTable.MaybeEncode(id.Namespace); ok && ns == fb.Namespaces[b6.FeatureTypePoint] {
			t, ok := fb.Map.FindFirst(id.Value)
			if ok {
				switch t.Tag {
				case PointTagCommon:
					var p CommonPoint
					p.Unmarshal(&fb.Namespaces, t.Data)
					_, ns := p.Path.TypeAndNamespace.Split()
					paths = append(paths, b6.FeatureID{Type: b6.FeatureTypePath, Namespace: fb.NamespaceTable.Decode(ns), Value: p.Path.Value})
				case PointTagFull:
					var p FullPoint
					p.Unmarshal(&fb.Namespaces, t.Data)
					for _, path := range p.Paths {
						_, ns := path.TypeAndNamespace.Split()
						paths = append(paths, b6.FeatureID{Type: b6.FeatureTypePath, Namespace: fb.NamespaceTable.Decode(ns), Value: path.Value})
					}
				case PointTagReferencesOnly:
					var r PointReferences
					r.Unmarshal(&fb.Namespaces, t.Data)
					for _, path := range r.Paths {
						_, ns := path.TypeAndNamespace.Split()
						paths = append(paths, b6.FeatureID{Type: b6.FeatureTypePath, Namespace: fb.NamespaceTable.Decode(ns), Value: path.Value})
					}
				}
			}
		}
	}
	return paths
}

func (f *FeaturesByID) FindAreasByPoint(id b6.FeatureID) b6.AreaFeatures {
	features := make([]b6.AreaFeature, 0, 1)
	for _, fb := range f.features[b6.FeatureTypePoint] {
		if ns, ok := fb.NamespaceTable.MaybeEncode(id.Namespace); ok && fb.Namespaces[b6.FeatureTypePoint] == ns {
			t, ok := fb.Map.FindFirst(id.Value)
			var paths []Reference
			if ok {
				switch t.Tag {
				case PointTagCommon:
					var p CommonPoint
					p.Unmarshal(&fb.Namespaces, t.Data)
					paths = []Reference{p.Path}
				case PointTagFull:
					var p FullPoint
					p.Unmarshal(&fb.Namespaces, t.Data)
					paths = p.Paths
				}
			}
			areas := make(map[Reference]struct{})
			var p Path
			for _, path := range paths {
				for _, pm := range f.features[b6.FeatureTypePath] {
					_, ns := path.TypeAndNamespace.Split()
					if pm.Namespaces[b6.FeatureTypePath] == ns {
						if b := pm.Map.FindFirstWithTag(path.Value, encoding.NoTag); len(b) > 0 {
							p.Unmarshal(&pm.Namespaces, b)
							for _, area := range p.Areas {
								areas[area] = struct{}{}
							}
							break
						}
					}
				}
			}
			for area := range areas {
				for _, am := range f.features[b6.FeatureTypeArea] {
					_, ns := area.TypeAndNamespace.Split()
					if am.Namespaces[b6.FeatureTypeArea] == ns {
						if a := f.newArea(am, area.Value); a != nil {
							features = append(features, a)
							break
						}
					}
				}
			}

		}
	}
	return ingest.NewAreaFeatureIterator(features)
}

func (f *FeaturesByID) Traverse(id b6.FeatureID) b6.Segments {
	pids := f.findPathsByPoint(id, make([]b6.FeatureID, 0, 2))
	segments := make([]b6.Segment, 0, len(pids)*2)
	for _, pid := range pids {
		segments = f.fillPathSegments(id, pid, segments)
	}
	return ingest.NewSegmentIterator(segments)
}

func (f *FeaturesByID) fillPathSegments(point b6.FeatureID, path b6.FeatureID, segments []b6.Segment) []b6.Segment {
	for _, fb := range f.features[b6.FeatureTypePath] {
		if ns, ok := fb.NamespaceTable.MaybeEncode(path.Namespace); ok && ns == fb.Namespaces[b6.FeatureTypePath] {
			b := fb.Map.FindFirstWithTag(path.Value, encoding.NoTag)
			if b == nil {
				continue
			}
			var p Path
			p.Unmarshal(&fb.Namespaces, b)
			previous := 0
			var position int
			next := p.PathLen(fb.Strings) - 1
			var pf b6.PhysicalFeature
			for i := 0; i < p.PathLen(fb.Strings); i++ {
				if id, ok := p.Reference(i, fb.Strings); ok {
					if pf == nil {
						_, ns := id.TypeAndNamespace.Split()
						if id.Value == point.Value && fb.NamespaceTable.Decode(ns) == point.Namespace {
							pf = b6.WrapPhysicalFeature(f.newPathFromEncodedPath(fb, path.Value, &p), f)
							position = i
						} else if f.isGraphNode(id) {
							previous = i
						}
					} else if f.isGraphNode(id) {
						next = i
						break
					}
				}
			}
			if pf != nil {
				if previous != position {
					segments = append(segments, b6.Segment{Feature: pf, First: position, Last: previous})
				}
				if next != position {
					segments = append(segments, b6.Segment{Feature: pf, First: position, Last: next})
				}
			}
			break
		}
	}
	return segments
}

// isGraphNode returns true if this point should be a node in the
// network graph. We currently consider intersections and points with tags
// as nodes.
func (f *FeaturesByID) isGraphNode(point Reference) bool {
	paths := 0
	for _, fb := range f.features[b6.FeatureTypePoint] {
		_, ns := point.TypeAndNamespace.Split()
		if fb.Namespaces[b6.FeatureTypePoint] == ns {
			t, ok := fb.Map.FindFirst(point.Value)
			if ok {
				switch t.Tag {
				case PointTagCommon:
					var p CommonPoint
					p.Unmarshal(&fb.Namespaces, t.Data)
					if len(p.Tags) > 0 {
						return true
					}
					paths++
				case PointTagFull:
					var p FullPoint
					p.Unmarshal(&fb.Namespaces, t.Data)
					if len(p.Tags) > 0 {
						return true
					}
					paths += len(p.Paths)
				case PointTagReferencesOnly:
					var r PointReferences
					r.Unmarshal(&fb.Namespaces, t.Data)
					paths += len(r.Paths)
				}
				if paths > 1 {
					return true
				}
			}
		}
	}
	return false
}

func (f *FeaturesByID) FindRelationsByFeature(id b6.FeatureID) b6.RelationFeatures {
	relations := make([]b6.RelationFeature, 0, 2)
	if int(id.Type) >= len(f.features) {
		return ingest.NewRelationFeatureIterator(relations)
	}
	for _, fb := range f.features[id.Type] {
		if ns, ok := fb.NamespaceTable.MaybeEncode(id.Namespace); ok && ns == fb.Namespaces[id.Type] {
			switch id.Type {
			case b6.FeatureTypePoint:
				relations = f.fillRelationsFromPoint(fb, id.Value, relations)
			case b6.FeatureTypePath:
				relations = f.fillRelationsFromPath(fb, id.Value, relations)
			case b6.FeatureTypeArea:
				relations = f.fillRelationsFromArea(fb, id.Value, relations)
			case b6.FeatureTypeRelation:
				relations = f.fillRelationsFromRelation(fb, id.Value, relations)
			}
			break
		}
	}
	return ingest.NewRelationFeatureIterator(relations)
}

func (f *FeaturesByID) fillRelationsFromPoint(fb *featureBlock, id uint64, relations []b6.RelationFeature) []b6.RelationFeature {
	t, ok := fb.Map.FindFirst(id)
	if ok && t.Tag == PointTagFull {
		var p FullPoint
		// TODO: don't need to unmarshal everything
		p.Unmarshal(&fb.Namespaces, t.Data)
		for _, r := range p.Relations {
			for _, rm := range f.features[b6.FeatureTypeRelation] {
				if _, ns := r.TypeAndNamespace.Split(); ns == rm.Namespaces[b6.FeatureTypeRelation] {
					relations = append(relations, f.newRelation(rm, r.Value))
					break
				}
			}
		}
	}
	return relations
}

func (f *FeaturesByID) fillRelationsFromPath(fb *featureBlock, id uint64, relations []b6.RelationFeature) []b6.RelationFeature {
	b := fb.Map.FindFirstWithTag(id, encoding.NoTag)
	if b != nil {
		var p Path
		p.Unmarshal(&fb.Namespaces, b)
		for _, r := range p.Relations {
			for _, rm := range f.features[b6.FeatureTypeRelation] {
				if _, ns := r.TypeAndNamespace.Split(); ns == rm.Namespaces[b6.FeatureTypeRelation] {
					relations = append(relations, f.newRelation(rm, r.Value))
					break
				}
			}
		}
	}
	return relations
}

func (f *FeaturesByID) fillRelationsFromArea(fb *featureBlock, id uint64, relations []b6.RelationFeature) []b6.RelationFeature {
	b := fb.Map.FindFirstWithTag(id, encoding.NoTag)
	if b != nil {
		var a Area
		a.Unmarshal(&fb.Namespaces, b)
		for _, r := range a.Relations {
			for _, rm := range f.features[b6.FeatureTypeRelation] {
				if _, ns := r.TypeAndNamespace.Split(); ns == rm.Namespaces[b6.FeatureTypeRelation] {
					relations = append(relations, f.newRelation(rm, r.Value))
					break
				}
			}
		}
	}
	return relations
}

func (f *FeaturesByID) fillRelationsFromRelation(fb *featureBlock, id uint64, relations []b6.RelationFeature) []b6.RelationFeature {
	b := fb.Map.FindFirstWithTag(id, encoding.NoTag)
	if b != nil {
		var r Relation
		r.Unmarshal(b6.FeatureTypePath, &fb.Namespaces, b)
		for _, rr := range r.Relations {
			for _, rm := range f.features[b6.FeatureTypeRelation] {
				if _, ns := rr.TypeAndNamespace.Split(); ns == rm.Namespaces[b6.FeatureTypeRelation] {
					relations = append(relations, f.newRelation(rm, rr.Value))
					break
				}
			}
		}
	}
	return relations
}

// TODO(andrew): implement binary encoding for expressions and collections.
func (f *FeaturesByID) FindCollectionsByFeature(id b6.FeatureID) b6.CollectionFeatures {
	collections := make([]b6.CollectionFeature, 0, 2)
	return ingest.NewCollectionFeatureIterator(collections)
}

func (f *FeaturesByID) LogSummary() {
	names := make([]string, 0, 8)
	maps := make([]*encoding.Uint64Map, 0, 8)

	for i, fbs := range f.features {
		for j, fb := range fbs {
			names = append(names, fmt.Sprintf("%s[%d]", b6.FeatureType(i), j))
			maps = append(maps, fb.Map)
		}
	}

	var histogram [16]int
	for i, m := range maps {
		if err := m.ComputeHistogram(histogram[0:]); err == nil {
			log.Printf("%s:", names[i])
			for i, count := range histogram {
				log.Printf("%4d: %d", i, count)
			}
		} else {
			log.Printf("%s: %s", names[i], err)
		}
	}
}

func (f *FeaturesByID) StatusText() string {
	status := ""
	for _, fbs := range f.features {
		for _, fb := range fbs {
			status += fmt.Sprintf("%s: %d: %d bytes\n", fb.FeatureType, fb.Namespaces[fb.FeatureType], fb.Map.Length())
		}
	}
	return status
}

type World struct {
	byID    *FeaturesByID
	indices []*Index
	status  string
	lock    sync.Mutex
}

func (w *World) FindFeatureByID(id b6.FeatureID) b6.Feature {
	return w.byID.FindFeatureByID(id)
}

func (w *World) HasFeatureWithID(id b6.FeatureID) bool {
	return w.FindFeatureByID(id) != nil
}

func (w *World) FindLocationByID(id b6.FeatureID) (s2.LatLng, error) {
	return w.byID.FindLocationByID(id)
}

func (w *World) FindFeatures(q b6.Query) b6.Features {
	features := make([]b6.Features, len(w.indices))
	for i, index := range w.indices {
		features[i] = b6.NewSearchFeatureIterator(q.Compile(index, w), index)
	}
	return b6.MergeFeatures(features...)
}

func (w *World) FindAreasByPoint(id b6.FeatureID) b6.AreaFeatures {
	return w.byID.FindAreasByPoint(id)
}

func (w *World) FindReferences(id b6.FeatureID, typed ...b6.FeatureType) b6.Features {
	return w.byID.FindReferences(id, typed...)
}

func (w *World) FindRelationsByFeature(id b6.FeatureID) b6.RelationFeatures {
	return w.byID.FindRelationsByFeature(id)
}

func (w *World) FindCollectionsByFeature(id b6.FeatureID) b6.CollectionFeatures {
	return w.byID.FindCollectionsByFeature(id)
}

func (w *World) Traverse(id b6.FeatureID) b6.Segments {
	return w.byID.Traverse(id)
}

func (w *World) EachFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	return w.byID.EachFeature(each, options)
}

func (w *World) Tokens() []string {
	seen := make(map[string]struct{})
	tokens := make([]string, 0)
	for _, index := range w.indices {
		t := index.Tokens()
		for t.Next() {
			if _, ok := seen[t.Token()]; !ok {
				tokens = append(tokens, t.Token())
				seen[t.Token()] = struct{}{}
			}
		}
	}
	return tokens
}

func toTags(tags Tags, s *encoding.StringTable, nt *NamespaceTable) []b6.Tag { // TODO(mari): see if we can get away with not decoding this at all
	// TODO: we don't need to decode everything given the interface
	ts := make([]b6.Tag, len(tags))
	for i, t := range tags {
		ts[i].Key = s.Lookup(t.Key)
		ts[i].Value = fromCompactValue(t.Value, s, nt)
	}
	return ts
}

func MergeFromFile(filename string, w *World) error {
	m, err := encoding.Mmap(filename)
	if err != nil {
		return err
	}
	return w.Merge(m.Data)
}

func NewWorldFromFile(filename string) (*World, error) {
	m, err := encoding.Mmap(filename)
	if err != nil {
		return nil, err
	}
	return NewWorldFromData(m.Data)
}

type emptyFeaturesByID struct{}

func (_ emptyFeaturesByID) FindFeatureByID(id b6.FeatureID) b6.Feature {
	return nil
}

func (_ emptyFeaturesByID) FindLocationByID(id b6.FeatureID) (s2.LatLng, error) {
	return s2.LatLng{}, fmt.Errorf("empty features")
}

func (_ emptyFeaturesByID) HasFeatureWithID(id b6.FeatureID) bool {
	return false
}

func NewWorld() *World {
	return &World{byID: NewFeaturesByID(emptyFeaturesByID{}), indices: make([]*Index, 0)}
}

func NewWorldWithBase(base b6.FeaturesByID) *World {
	return &World{byID: NewFeaturesByID(base), indices: make([]*Index, 0)}
}

func NewWorldFromData(data []byte) (*World, error) {
	w := NewWorld()
	return w, w.Merge(data)
}

func (w *World) Merge(data []byte) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	var header Header
	header.Unmarshal(data)
	if header.Magic != HeaderMagic {
		return fmt.Errorf("Bad header magic: expected %x, found %x", uint64(HeaderMagic), header.Magic)
	}
	if err := verifyVersion(&header, data); err != nil {
		return err
	}
	if err := w.byID.Merge(data); err != nil {
		return err
	}

	w.status += fmt.Sprintf("version %s\n", header.UnmarshalVersion(data))

	var hp pb.CompactHeaderProto
	if err := UnmarshalProto(data[header.HeaderProtoOffset:], &hp); err != nil {
		return err
	}
	var nt NamespaceTable
	nt.FillFromProto(&hp)

	offset := header.BlockOffset
	for offset < encoding.Offset(len(data)) {
		var header BlockHeader
		offset += encoding.Offset(header.Unmarshal(data[offset:]))
		if header.Type == BlockTypeSearchIndex {
			index, err := NewIndex(data[offset:], &nt, w)
			if err != nil {
				return fmt.Errorf("Failed to create search index: %s", err)
			}
			w.indices = append(w.indices, index)
		}
		offset += encoding.Offset(header.Length)
	}
	return nil
}

func (w *World) ServeHTTP(rw http.ResponseWriter, rr *http.Request) {
	output := w.status
	output += w.byID.StatusText()
	for _, index := range w.indices {
		output += search.TokensToHTML(index.Tokens())
	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.Write([]byte(output))
}

type Index struct {
	w  *World
	t  TokenMap
	b  *encoding.ByteArrays
	nt *NamespaceTable
}

func NewIndex(data []byte, nt *NamespaceTable, w *World) (*Index, error) {
	i := &Index{w: w, nt: nt}
	n := i.t.Unmarshal(data)
	i.b = encoding.NewByteArrays(data[n:])
	return i, nil
}

func (i *Index) Begin(token string) search.Iterator {
	j := i.t.FindPossibleIndices(token)
	for {
		index, ok := j.Next()
		if !ok {
			break
		}
		item := i.b.Item(index)
		if PostingListHeaderTokenEquals(item, token) {
			return NewIterator(item, i.nt)
		}
	}
	return search.NewEmptyIterator()
}

func (i *Index) Feature(v search.Value) b6.Feature {
	return i.w.FindFeatureByID(v.(b6.FeatureID))
}

func (i *Index) ID(v search.Value) b6.FeatureID {
	return v.(b6.FeatureID)
}

func (i *Index) NumTokens() int {
	return i.b.NumItems()
}

type tokenIterator struct {
	b *encoding.ByteArrays
	i int
}

func (t *tokenIterator) Next() bool {
	t.i++
	return t.i <= t.b.NumItems()
}

func (t *tokenIterator) Token() string {
	return PostingListHeaderToken(t.b.Item(t.i - 1))
}

func (t *tokenIterator) Advance(token string) bool {
	start := t.i - 1
	if start < 0 {
		start = 0
	}
	i := sort.Search(t.b.NumItems()-start, func(j int) bool {
		return PostingListHeaderToken(t.b.Item(start+j)) >= token
	})
	t.i = start + i + 1
	return t.i <= t.b.NumItems()
}

func (i *Index) Tokens() search.TokenIterator {
	return &tokenIterator{b: i.b}
}

func (i *Index) Values() search.Values {
	return i
}

func (i *Index) Compare(a search.Value, b search.Value) search.Comparison {
	ida, idb := a.(b6.FeatureID), b.(b6.FeatureID)
	if ida.Less(idb) {
		return search.ComparisonLess
	} else if ida == idb {
		return search.ComparisonEqual
	}
	return search.ComparisonGreater
}

func (i *Index) CompareKey(v search.Value, k search.Key) search.Comparison {
	return i.Compare(v, k)
}

func (i *Index) Key(v search.Value) search.Key {
	return v
}
