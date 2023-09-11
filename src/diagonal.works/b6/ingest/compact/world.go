package compact

import (
	"fmt"
	"log"
	"net/http"
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
	for _, fb := range f.features[id.Type] {
		if ns, ok := fb.NamespaceTable.MaybeEncode(id.Namespace); ok && ns == fb.Namespaces[id.Type] {
			switch id.Type {
			case b6.FeatureTypePoint:
				if p := f.newPoint(fb, id.Value); p != nil {
					return p
				}
			case b6.FeatureTypePath:
				if p := f.newPath(fb, id.Value); p != nil {
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

func (f *FeaturesByID) FindLocationByID(id b6.PointID) (s2.LatLng, bool) {
	for _, fb := range f.features[b6.FeatureTypePoint] {
		if ns, ok := fb.NamespaceTable.MaybeEncode(id.Namespace); ok && ns == fb.Namespaces[b6.FeatureTypePoint] {
			if t, ok := fb.Map.FindFirst(id.Value); ok {
				if t.Tag != PointTagReferencesOnly {
					return MarshalledPoint(t.Data).Location(), true
				}
			}
		}
	}
	return f.base.FindLocationByID(id)
}

func (f *FeaturesByID) newPoint(fb *featureBlock, id uint64) b6.Feature {
	t, ok := fb.Map.FindFirst(id)
	if ok {
		return f.newPointFromTagged(fb, id, t)
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

type marshalledPoint struct {
	id      b6.PointID
	point   MarshalledPoint
	strings encoding.Strings
}

func (m marshalledPoint) FeatureID() b6.FeatureID {
	return m.id.FeatureID()
}

func (m marshalledPoint) PointID() b6.PointID {
	return m.id
}

func (m marshalledPoint) AllTags() []b6.Tag {
	return m.point.Tags(m.strings).AllTags()
}

func (m marshalledPoint) Get(key string) b6.Tag {
	return m.point.Tags(m.strings).Get(key)
}

func (m marshalledPoint) Point() s2.Point {
	return s2.PointFromLatLng(m.point.Location())
}

func (m marshalledPoint) CellID() s2.CellID {
	return s2.CellIDFromLatLng(m.point.Location())
}

func (m marshalledPoint) Covering(coverer s2.RegionCoverer) s2.CellUnion {
	return coverer.Covering(m.Point())
}

func (m marshalledPoint) ToGeoJSON() geojson.GeoJSON {
	return b6.PointFeatureToGeoJSON(m)
}

func (f *FeaturesByID) newPointFromTagged(fb *featureBlock, id uint64, t encoding.Tagged) b6.PointFeature {
	if t.Tag != PointTagReferencesOnly {
		return marshalledPoint{
			id:      b6.MakePointID(fb.NamespaceTable.Decode(fb.Namespaces[b6.FeatureTypePoint]), id),
			point:   MarshalledPoint(t.Data),
			strings: fb.Strings,
		}
	}
	return nil
}

type marshalledPath struct {
	id       b6.PathID
	path     MarshalledPath
	geometry PathGeometry
	polyline s2.Polyline
	fb       *featureBlock
	byID     *FeaturesByID
	lock     sync.Mutex
}

func (m *marshalledPath) FeatureID() b6.FeatureID {
	return m.id.FeatureID()
}

func (m *marshalledPath) PathID() b6.PathID {
	return m.id
}

func (m *marshalledPath) AllTags() []b6.Tag {
	return m.path.Tags(m.fb.Strings).AllTags()
}

func (m *marshalledPath) Get(key string) b6.Tag {
	return m.path.Tags(m.fb.Strings).Get(key)
}

func (m *marshalledPath) Len() int {
	return m.path.Len()
}

func (m *marshalledPath) Point(i int) s2.Point {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.fillGeometry()
	var ll s2.LatLng
	if r, ok := m.geometry.PointID(i); ok {
		id := b6.MakePointID(m.fb.NamespaceTable.Decode(r.Namespace), r.Value)
		var ok bool
		ll, ok = m.byID.FindLocationByID(id)
		if !ok {
			panic(fmt.Sprintf("Missing point %s", id))
		}
	} else {
		cll, ok := m.geometry.LatLng(i)
		if !ok {
			panic("Expected a latlng")
		}
		ll = cll.ToS2LatLng()
	}
	return s2.PointFromLatLng(ll)
}

func (m *marshalledPath) Polyline() *s2.Polyline {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.fillGeometry()
	if m.polyline == nil {
		m.polyline = make(s2.Polyline, m.geometry.Len())
		for i := 0; i < m.geometry.Len(); i++ {
			if id, ok := m.geometry.PointID(i); ok {
				id := b6.MakePointID(m.fb.NamespaceTable.Decode(id.Namespace), id.Value)
				ll, ok := m.byID.FindLocationByID(id)
				if !ok {
					panic(fmt.Sprintf("Missing point %s", id))
				}
				m.polyline[i] = s2.PointFromLatLng(ll)
			} else {
				ll, _ := m.geometry.LatLng(i)
				m.polyline[i] = s2.PointFromLatLng(ll.ToS2LatLng())
			}
		}
	}
	return &m.polyline
}

func (m *marshalledPath) Feature(i int) b6.PointFeature {
	id := m.feature(i)
	return b6.FindPointByID(id, m.byID)
}

func (m *marshalledPath) Covering(coverer s2.RegionCoverer) s2.CellUnion {
	return coverer.Covering(m.Polyline())
}

func (m *marshalledPath) feature(i int) b6.PointID {
	m.lock.Lock()
	m.fillGeometry()
	m.lock.Unlock()
	if r, ok := m.geometry.PointID(i); ok {
		return b6.MakePointID(m.fb.NamespaceTable.Decode(r.Namespace), r.Value)
	}
	return b6.PointIDInvalid
}

func (m *marshalledPath) fillGeometry() {
	if m.geometry == nil {
		m.geometry, _ = m.path.UnmarshalPoints(&m.fb.Namespaces)
	}
}

func (m *marshalledPath) ToGeoJSON() geojson.GeoJSON {
	return b6.PathFeatureToGeoJSON(m)
}

func (f *FeaturesByID) newPath(fb *featureBlock, id uint64) b6.PathFeature {
	b := fb.Map.FindFirstWithTag(id, encoding.NoTag)
	if b != nil {
		return f.newPathFromBuffer(fb, id, b)
	}
	return nil
}

func (f *FeaturesByID) newPathFromBuffer(fb *featureBlock, id uint64, buffer []byte) b6.PathFeature {
	return &marshalledPath{
		id:   b6.MakePathID(fb.NamespaceTable.Decode(fb.Namespaces[b6.FeatureTypePath]), id),
		path: MarshalledPath(buffer),
		fb:   fb,
		byID: f,
	}
}

func (f *FeaturesByID) newPathFromEncodedPath(fb *featureBlock, id uint64, p *Path) *ingest.PathFeature {
	pp := ingest.NewPathFeature(p.Points.Len())
	pp.PathID = b6.MakePathID(fb.NamespaceTable.Decode(fb.Namespaces[b6.FeatureTypePath]), id)
	// TODO: avoid the creation of tags in the encoded path?
	pp.Tags = toTags(p.Tags, fb.Strings)
	for i := 0; i < p.Points.Len(); i++ {
		if id, ok := p.PointID(i); ok {
			pp.SetPointID(i, b6.MakePointID(fb.NamespaceTable.Decode(id.Namespace), id.Value))
		} else {
			ll, _ := p.LatLng(i)
			pp.SetLatLng(i, ll.ToS2LatLng())
		}
	}
	return pp
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

func (m *marshalledArea) AllTags() []b6.Tag {
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

func (m *marshalledArea) Feature(i int) []b6.PathFeature {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.featureWithLock(i)
}

func (m *marshalledArea) featureWithLock(i int) []b6.PathFeature {
	m.fillGeometry()
	var paths []b6.PathFeature
	if ids, ok := m.geometry.PathIDs(i); ok {
		paths = make([]b6.PathFeature, len(ids))
		for i := range ids {
			id := b6.MakePathID(m.fb.NamespaceTable.Decode(ids[i].Namespace), ids[i].Value)
			if path := b6.FindPathByID(id, m.byID); path != nil {
				paths[i] = path
			} else {
				panic(fmt.Sprintf("Missing path %s for %s", id, m.id))
			}
		}
	}
	return paths
}

func (m *marshalledArea) Covering(coverer s2.RegionCoverer) s2.CellUnion {
	covering := s2.CellUnion([]s2.CellID{})
	for i := 0; i < m.Len(); i++ {
		if polygon := m.Polygon(i); polygon != nil {
			covering = s2.CellUnionFromUnion(covering, coverer.Covering(polygon))
		}
	}
	return covering
}

func (m *marshalledArea) fillGeometry() {
	if m.geometry == nil {
		m.geometry = m.area.UnmarshalPolygons(m.fb.Namespaces[b6.FeatureTypePath])
		m.polygons = make([]*s2.Polygon, m.geometry.Len())
	}
}

func (m *marshalledArea) ToGeoJSON() geojson.GeoJSON {
	return b6.AreaFeatureToGeoJSON(m)
}

func (f *FeaturesByID) newArea(fb *featureBlock, id uint64) b6.AreaFeature {
	b := fb.Map.FindFirstWithTag(id, encoding.NoTag)
	if b != nil {
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

func (m *marshalledRelation) AllTags() []b6.Tag {
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
	return b6.RelationMember{
		Role: m.fb.Strings.Lookup(m.members[i].Role),
		ID:   b6.FeatureID{Type: m.members[i].Type, Namespace: m.fb.NamespaceTable.Decode(m.members[i].ID.Namespace), Value: m.members[i].ID.Value},
	}
}

func (m *marshalledRelation) Covering(coverer s2.RegionCoverer) s2.CellUnion {
	return s2.CellUnion{}
}

func (m *marshalledRelation) fillMembers() {
	if len(m.members) == 0 {
		m.relation.UnmarshalMembers(b6.FeatureTypePath, &m.fb.Namespaces, &m.members)
	}
}

func (m *marshalledRelation) ToGeoJSON() geojson.GeoJSON {
	return b6.RelationFeatureToGeoJSON(m, m.byID)
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
				if point := f.newPointFromTagged(fb, id, tagged[0]); point != nil {
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
				return each(f.newPathFromBuffer(fb, id, tagged[0].Data), g)
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

func (f *FeaturesByID) FindPathsByPoint(id b6.PointID) b6.PathFeatures {
	pids := f.findPathsByPoint(id, make([]b6.PathID, 0, 2))
	paths := make([]b6.PathFeature, len(pids))
	for i, pid := range pids {
		paths[i] = f.FindFeatureByID(pid.FeatureID()).(b6.PathFeature)
	}
	return ingest.NewPathFeatureIterator(paths)
}

func (f *FeaturesByID) findPathsByPoint(id b6.PointID, paths []b6.PathID) []b6.PathID {
	for _, fb := range f.features[b6.FeatureTypePoint] {
		if ns, ok := fb.NamespaceTable.MaybeEncode(id.Namespace); ok && ns == fb.Namespaces[b6.FeatureTypePoint] {
			t, ok := fb.Map.FindFirst(id.Value)
			if ok {
				switch t.Tag {
				case PointTagCommon:
					var p CommonPoint
					p.Unmarshal(&fb.Namespaces, t.Data)
					paths = append(paths, b6.MakePathID(fb.NamespaceTable.Decode(p.Path.Namespace), p.Path.Value))
				case PointTagFull:
					var p FullPoint
					p.Unmarshal(&fb.Namespaces, t.Data)
					for _, path := range p.Paths {
						paths = append(paths, b6.MakePathID(fb.NamespaceTable.Decode(path.Namespace), path.Value))
					}
				case PointTagReferencesOnly:
					var r PointReferences
					r.Unmarshal(&fb.Namespaces, t.Data)
					for _, path := range r.Paths {
						paths = append(paths, b6.MakePathID(fb.NamespaceTable.Decode(path.Namespace), path.Value))
					}
				}
			}
		}
	}
	return paths
}

func (f *FeaturesByID) FindAreasByPoint(id b6.PointID) b6.AreaFeatures {
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
					if pm.Namespaces[b6.FeatureTypePath] == path.Namespace {
						b := pm.Map.FindFirstWithTag(path.Value, encoding.NoTag)
						p.Unmarshal(&pm.Namespaces, b)
						for _, area := range p.Areas {
							areas[area] = struct{}{}
						}
						break
					}
				}
			}
			for area := range areas {
				for _, am := range f.features[b6.FeatureTypeArea] {
					if am.Namespaces[b6.FeatureTypeArea] == area.Namespace {
						features = append(features, f.newArea(am, area.Value))
						break
					}
				}
			}

		}
	}
	return ingest.NewAreaFeatureIterator(features)
}

func (f *FeaturesByID) Traverse(id b6.PointID) b6.Segments {
	pids := f.findPathsByPoint(id, make([]b6.PathID, 0, 2))
	segments := make([]b6.Segment, 0, len(pids)*2)
	for _, pid := range pids {
		segments = f.fillPathSegments(id, pid, segments)
	}
	return ingest.NewSegmentIterator(segments)
}

func (f *FeaturesByID) fillPathSegments(point b6.PointID, path b6.PathID, segments []b6.Segment) []b6.Segment {
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
			next := p.Points.Len() - 1
			var pf b6.PathFeature
			for i := 0; i < p.Points.Len(); i++ {
				if id, ok := p.Points.PointID(i); ok {
					if pf == nil {
						if id.Value == point.Value && fb.NamespaceTable.Decode(id.Namespace) == point.Namespace {
							pf = ingest.WrapPathFeature(f.newPathFromEncodedPath(fb, path.Value, &p), f)
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
		if fb.Namespaces[b6.FeatureTypePoint] == point.Namespace {
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
				if r.Namespace == rm.Namespaces[b6.FeatureTypeRelation] {
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
				if r.Namespace == rm.Namespaces[b6.FeatureTypeRelation] {
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
				if r.Namespace == rm.Namespaces[b6.FeatureTypeRelation] {
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
				if rr.Namespace == rm.Namespaces[b6.FeatureTypeRelation] {
					relations = append(relations, f.newRelation(rm, rr.Value))
					break
				}
			}
		}
	}
	return relations
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

func (w *World) FindLocationByID(id b6.PointID) (s2.LatLng, bool) {
	return w.byID.FindLocationByID(id)
}

func (w *World) FindFeatures(q b6.Query) b6.Features {
	features := make([]b6.Features, len(w.indices))
	for i, index := range w.indices {
		features[i] = b6.NewFeatureIterator(q.Compile(index, w), index)
	}
	return b6.MergeFeatures(features...)
}

func (w *World) FindPathsByPoint(id b6.PointID) b6.PathFeatures {
	return w.byID.FindPathsByPoint(id)
}

func (w *World) FindAreasByPoint(id b6.PointID) b6.AreaFeatures {
	return w.byID.FindAreasByPoint(id)
}

func (w *World) FindRelationsByFeature(id b6.FeatureID) b6.RelationFeatures {
	return w.byID.FindRelationsByFeature(id)
}

func (w *World) Traverse(id b6.PointID) b6.Segments {
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

func toTags(tags Tags, s *encoding.StringTable) []b6.Tag {
	// TODO: we don't need to decode everything given the interface
	ts := make([]b6.Tag, len(tags))
	for i, t := range tags {
		ts[i].Key = s.Lookup(t.Key)
		ts[i].Value = s.Lookup(t.Value)
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

func (_ emptyFeaturesByID) FindLocationByID(id b6.PointID) (s2.LatLng, bool) {
	return s2.LatLng{}, false
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
