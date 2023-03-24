package ingest

import (
	"context"
	"fmt"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/osm"
	"github.com/golang/geo/s2"
)

type Tags []b6.Tag

func (t Tags) Get(key string) b6.Tag {
	for _, tag := range t {
		if tag.Key == key {
			return tag
		}
	}
	return b6.InvalidTag()
}

func (t Tags) TagOrFallback(key string, fallback string) b6.Tag {
	if value := t.Get(key); value.IsValid() {
		return value
	}
	return b6.Tag{Key: key, Value: fallback}
}

func (t *Tags) AddTag(tag b6.Tag) {
	*t = append(*t, tag)
}

// Modifies an existing tag value, or add it if it doesn't exist.
// Returns (true, old value) if it modifies, or (false, undefined) if
// added.
func (t *Tags) ModifyOrAddTag(tag b6.Tag) (bool, string) {
	for i := range *t {
		if (*t)[i].Key == tag.Key {
			old := (*t)[i].Value
			(*t)[i].Value = tag.Value
			return true, old
		}
	}
	t.AddTag(tag)
	return false, ""
}

func (t *Tags) RemoveTag(key string) {
	w := 0
	for r := range *t {
		if (*t)[r].Key != key {
			if w != r {
				(*t)[w] = (*t)[r]
			}
			w++
		}
	}
	*t = (*t)[0:w]
}

func (t Tags) AllTags() []b6.Tag {
	return t
}

func (t Tags) Clone() Tags {
	clone := make([]b6.Tag, len(t))
	for i, tag := range t {
		clone[i] = tag
	}
	return clone
}

func (t *Tags) MergeFrom(other Tags) {
	i := copy(*t, other)
	if i < len(other) {
		*t = append(*t, other[i:]...)
	} else {
		*t = (*t)[0:len(other)]
	}
}

func (t *Tags) FillFromOSM(tags osm.Tags) {
	*t = (*t)[0:0]
	for _, tag := range tags {
		key := tag.Key
		if mapped, ok := osmTagMapping[tag.Key]; ok {
			key = mapped
		}
		*t = append(*t, b6.Tag{Key: key, Value: tag.Value})
	}
}

type ByTagKey Tags

func (t ByTagKey) Len() int           { return len(t) }
func (t ByTagKey) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t ByTagKey) Less(i, j int) bool { return t[i].Key < t[j].Key }

func NewTagsFromWorld(t b6.Taggable) Tags {
	all := t.AllTags()
	tags := make([]b6.Tag, len(all))
	for i := range all {
		tags[i] = all[i]
	}
	return tags
}

type Feature interface {
	FeatureID() b6.FeatureID
	b6.Taggable
	AddTag(tag b6.Tag)
	ModifyOrAddTag(tag b6.Tag) (bool, string)
	RemoveTag(key string)
	Clone() Feature
	MergeFrom(other Feature)
}

func NewFeatureFromWorld(f b6.Feature) Feature {
	switch f := f.(type) {
	case b6.PointFeature:
		return NewPointFeatureFromWorld(f)
	case b6.PathFeature:
		return NewPathFeatureFromWorld(f)
	case b6.AreaFeature:
		return NewAreaFeatureFromWorld(f)
	case b6.RelationFeature:
		return NewRelationFeatureFromWorld(f)
	}
	panic(fmt.Sprintf("Can't handle feature type: %T", f))
}

func WrapFeature(f Feature, byID b6.FeaturesByID) b6.Feature {
	switch f := f.(type) {
	case *PointFeature:
		return WrapPointFeature(f)
	case *PathFeature:
		return WrapPathFeature(f, byID)
	case *AreaFeature:
		return WrapAreaFeature(f, byID)
	case *RelationFeature:
		return WrapRelationFeature(f, byID)
	}
	panic(fmt.Sprintf("Can't handle feature type: %T", f))
}

type PointFeature struct {
	b6.PointID
	Tags
	Location s2.LatLng
}

func NewPointFeature(id b6.PointID, location s2.LatLng) *PointFeature {
	return &PointFeature{
		PointID:  id,
		Tags:     []b6.Tag{},
		Location: location,
	}
}

func NewPointFeatureFromOSM(o *osm.Node) *PointFeature {
	p := &PointFeature{}
	p.FillFromOSM(o)
	return p
}

func NewPointFeatureFromWorld(p b6.PointFeature) *PointFeature {
	return &PointFeature{
		PointID:  p.PointID(),
		Tags:     NewTagsFromWorld(p),
		Location: s2.LatLngFromPoint(p.Point()),
	}
}

func (p *PointFeature) FillFromOSM(node *osm.Node) {
	p.PointID = FromOSMNodeID(node.ID)
	p.Tags.FillFromOSM(node.Tags)
	p.Location = node.Location.ToS2LatLng()
}

func (p *PointFeature) Point() s2.Point {
	return s2.PointFromLatLng(p.Location)
}

func (p *PointFeature) CellID() s2.CellID {
	return s2.CellIDFromLatLng(p.Location)
}

func (p *PointFeature) Clone() Feature {
	return p.ClonePointFeature()
}

func (p *PointFeature) ClonePointFeature() *PointFeature {
	return &PointFeature{
		PointID:  p.PointID,
		Tags:     p.Tags.Clone(),
		Location: p.Location,
	}
}

func (p *PointFeature) MergeFrom(other Feature) {
	if point, ok := other.(*PointFeature); ok {
		p.MergeFromPointFeature(point)
	} else {
		panic(fmt.Sprintf("Expected a PointFeature, found %T", other))
	}
}

func (p *PointFeature) MergeFromPointFeature(other *PointFeature) {
	p.PointID = other.PointID
	p.Location = other.Location
	p.Tags.MergeFrom(other.Tags)
}

type PathMembers struct {
	ids []b6.PointID
}

func NewPathMembers(n int) PathMembers {
	return PathMembers{ids: make([]b6.PointID, n)}
}

func (p *PathMembers) FillFromOSM(way *osm.Way) {
	p.ids = p.ids[0:0]
	for _, id := range way.Nodes {
		p.ids = append(p.ids, FromOSMNodeID(id))
	}
}

func (p *PathMembers) Len() int {
	return len(p.ids)
}

func (p *PathMembers) SetPointID(i int, id b6.PointID) {
	p.ids[i] = id
}

func (p *PathMembers) PointID(i int) (b6.PointID, bool) {
	if p.ids[i].Namespace != b6.NamespaceLatLng {
		return p.ids[i], true
	}
	return b6.PointID{}, false
}

func (p *PathMembers) SetLatLng(i int, ll s2.LatLng) {
	p.ids[i] = NewLatLngID(ll)
}

func (p *PathMembers) LatLng(i int) (s2.LatLng, bool) {
	return LatLngFromID(p.ids[i])
}

func (p *PathMembers) Point(i int) (s2.Point, bool) {
	if ll, ok := p.LatLng(i); ok {
		return s2.PointFromLatLng(ll), true
	}
	return s2.Point{}, false
}

func (p *PathMembers) IsClosed() bool {
	if begin, ok := p.PointID(0); ok {
		if end, ok := p.PointID(p.Len() - 1); ok {
			return begin == end
		}
	}
	return false
}

func (p *PathMembers) Invert() {
	n := len(p.ids)
	for i := 0; i < n/2; i++ {
		p.ids[i], p.ids[n-i-1] = p.ids[n-i-1], p.ids[i]
	}
}

func (p *PathMembers) Clone() PathMembers {
	clone := PathMembers{ids: make([]b6.PointID, len(p.ids))}
	for i, id := range p.ids {
		clone.ids[i] = id
	}
	return clone
}

func (p *PathMembers) MergeFrom(other PathMembers) {
	i := copy(p.ids, other.ids)
	if i < len(other.ids) {
		p.ids = append(p.ids, other.ids[i:]...)
	} else {
		p.ids = p.ids[0:len(other.ids)]
	}
}

type PathFeature struct {
	b6.PathID
	Tags
	PathMembers
}

func NewPathFeature(n int) *PathFeature {
	return &PathFeature{PathMembers: NewPathMembers(n)}
}

func NewPathFeatureFromOSM(way *osm.Way) *PathFeature {
	path := &PathFeature{}
	path.FillFromOSM(way)
	return path
}

func NewPathFeatureFromWorld(p b6.PathFeature) *PathFeature {
	path := NewPathFeature(p.Len())
	path.PathID = p.PathID()
	path.Tags = NewTagsFromWorld(p)
	for i := 0; i < p.Len(); i++ {
		if point := p.Feature(i); point != nil {
			path.SetPointID(i, point.PointID())
		} else {
			path.SetLatLng(i, s2.LatLngFromPoint(p.Point(i)))
		}
	}
	return path
}

func (p *PathFeature) FillFromOSM(way *osm.Way) {
	p.PathID = FromOSMWayID(way.ID)
	p.Tags.FillFromOSM(way.Tags)
	p.PathMembers.FillFromOSM(way)
}

func (p *PathFeature) FillFromOSMForArea(way *osm.Way) {
	p.PathID = FromOSMWayID(way.ID)
	p.Tags = p.Tags[0:0]
	p.PathMembers.FillFromOSM(way)
}

// AllPoints returns a list of s2.Points for the
// points along the path, or nil if at least one point is missing,
// together with an error
func (p *PathFeature) AllPoints(byID b6.LocationsByID) ([]s2.Point, error) {
	points := make([]s2.Point, p.Len())
	for i := 0; i < p.Len(); i++ {
		if point, ok := p.Point(i); ok {
			points[i] = point
		} else {
			id, _ := p.PointID(i)
			if ll, ok := byID.FindLocationByID(id); ok {
				points[i] = s2.PointFromLatLng(ll)
			} else {
				return nil, fmt.Errorf("Path %s missing point %s", p.PathID, id)
			}
		}
	}
	return points, nil
}

func (p *PathFeature) Clone() Feature {
	return p.ClonePathFeature()
}

func (p *PathFeature) ClonePathFeature() *PathFeature {
	return &PathFeature{
		PathID:      p.PathID,
		Tags:        p.Tags.Clone(),
		PathMembers: p.PathMembers.Clone(),
	}
}

func (p *PathFeature) MergeFrom(other Feature) {
	if path, ok := other.(*PathFeature); ok {
		p.MergeFromPathFeature(path)
	} else {
		panic(fmt.Sprintf("Expected a PathFeature, found %T", other))
	}
}

func (p *PathFeature) MergeFromPathFeature(other *PathFeature) {
	p.PathID = other.PathID
	p.Tags.MergeFrom(other.Tags)
	p.PathMembers.MergeFrom(other.PathMembers)
}

type AreaMembers struct {
	ids      [][]b6.PathID
	polygons []*s2.Polygon
}

func NewAreaMembers(n int) AreaMembers {
	return AreaMembers{
		ids:      make([][]b6.PathID, n),
		polygons: make([]*s2.Polygon, n),
	}
}

func (a *AreaMembers) Len() int {
	return len(a.ids)
}

func (a *AreaMembers) SetPathIDs(i int, ids []b6.PathID) {
	a.ids[i] = ids
	a.polygons[i] = nil
}

func (a *AreaMembers) SetPathID(i int, j int, id b6.PathID) {
	for len(a.ids[i]) <= j {
		a.ids[i] = append(a.ids[i], b6.PathIDInvalid)
	}
	a.ids[i][j] = id
	a.polygons[i] = nil
}

func (a *AreaMembers) PathIDs(i int) ([]b6.PathID, bool) {
	if a.ids[i] != nil {
		return a.ids[i], true
	}
	return nil, false
}

func (a *AreaMembers) SetPolygon(i int, polygon *s2.Polygon) {
	a.ids[i] = nil
	a.polygons[i] = polygon
}

func (a *AreaMembers) Polygon(i int) (*s2.Polygon, bool) {
	if a.polygons[i] != nil {
		return a.polygons[i], true
	}
	return nil, false
}

func (a *AreaMembers) Clone() AreaMembers {
	clone := AreaMembers{
		ids:      make([][]b6.PathID, len(a.ids)),
		polygons: make([]*s2.Polygon, len(a.polygons)),
	}
	copy(clone.ids, a.ids)
	copy(clone.polygons, a.polygons)
	return clone
}

func (a *AreaMembers) MergeFrom(other AreaMembers) {
	if len(a.ids) < len(other.ids) {
		for i := len(a.ids); i < len(other.ids); i++ {
			a.ids = append(a.ids, make([]b6.PathID, len(other.ids[i])))
		}
	} else {
		a.ids = a.ids[0:len(other.ids)]
	}
	for i, ids := range other.ids {
		j := copy(a.ids[i], ids)
		if j < len(ids) {
			a.ids[i] = append(a.ids[i], ids[j:]...)
		} else {
			a.ids[i] = a.ids[i][0:len(ids)]
		}
	}
	i := copy(a.polygons, other.polygons)
	if i < len(other.polygons) {
		a.polygons = append(a.polygons, other.polygons[i:]...)
	} else {
		a.polygons = a.polygons[0:len(other.polygons)]
	}
}

func (a *AreaMembers) FillFromOSMWay(way *osm.Way) {
	a.ids = a.ids[0:0]
	a.ids = append(a.ids, []b6.PathID{FromOSMWayID(way.ID)})
	a.polygons = a.polygons[0:0]
	a.polygons = append(a.polygons, nil)
}

type AreaFeature struct {
	b6.AreaID
	Tags
	AreaMembers
}

func NewAreaFeature(n int) *AreaFeature {
	return &AreaFeature{AreaMembers: NewAreaMembers(n)}
}

func NewAreaFeatureFromOSMWay(way *osm.Way) (*AreaFeature, *PathFeature) {
	path := NewPathFeatureFromOSM(way)
	path.Tags = make(Tags, 0) // Tags only exist on the area feature
	area := &AreaFeature{}
	area.FillFromOSMWay(way)
	area.SetPathIDs(0, []b6.PathID{path.PathID})
	return area, path
}

func NewAreaFeatureFromWorld(a b6.AreaFeature) *AreaFeature {
	area := NewAreaFeature(a.Len())
	area.AreaID = a.AreaID()
	area.Tags = NewTagsFromWorld(a)
	for i := 0; i < a.Len(); i++ {
		if paths := a.Feature(i); paths != nil {
			ids := make([]b6.PathID, len(paths))
			for j, path := range paths {
				ids[j] = path.PathID()
			}
			area.SetPathIDs(i, ids)
		} else {
			area.SetPolygon(i, a.Polygon(i))
		}
	}
	return area
}

func NewAreaFeatureFromS2Cell(id b6.AreaID, cell s2.CellID) *AreaFeature {
	area := NewAreaFeature(1)
	area.AreaID = id
	area.SetPolygon(0, s2.PolygonFromLoops([]*s2.Loop{s2.LoopFromCell(s2.CellFromCellID(cell))}))
	return area
}

func (a *AreaFeature) FillFromOSMWay(way *osm.Way) {
	a.AreaID = AreaIDFromOSMWayID(way.ID)
	a.Tags.FillFromOSM(way.Tags)
	a.AreaMembers.FillFromOSMWay(way)
}

func (a *AreaFeature) Clone() Feature {
	return a.CloneAreaFeature()
}

func (a *AreaFeature) CloneAreaFeature() *AreaFeature {
	return &AreaFeature{
		AreaID:      a.AreaID,
		Tags:        a.Tags.Clone(),
		AreaMembers: a.AreaMembers.Clone(),
	}
}

func (a *AreaFeature) MergeFrom(other Feature) {
	if area, ok := other.(*AreaFeature); ok {
		a.MergeFromAreaFeature(area)
	} else {
		panic(fmt.Sprintf("Expected a AreaFeature, found %T", other))
	}
}

func (a *AreaFeature) MergeFromAreaFeature(other *AreaFeature) {
	a.AreaID = other.AreaID
	a.Tags.MergeFrom(other.Tags)
	a.AreaMembers.MergeFrom(other.AreaMembers)
}

type RelationFeature struct {
	b6.RelationID
	Tags
	Members []b6.RelationMember
}

func NewRelationFeature(n int) *RelationFeature {
	return &RelationFeature{Members: make([]b6.RelationMember, n)}
}

func NewRelationFeatureFromWorld(r b6.RelationFeature) *RelationFeature {
	relation := NewRelationFeature(r.Len())
	relation.RelationID = r.RelationID()
	relation.Tags = NewTagsFromWorld(r)
	for i := 0; i < r.Len(); i++ {
		relation.Members[i] = r.Member(i)
	}
	return relation
}

func (r *RelationFeature) Len() int {
	return len(r.Members)
}

func (r *RelationFeature) Member(i int) b6.RelationMember {
	return r.Members[i]
}

func (r *RelationFeature) Clone() Feature {
	return r.CloneRelationFeature()
}

func (r *RelationFeature) CloneRelationFeature() *RelationFeature {
	clone := &RelationFeature{
		RelationID: r.RelationID,
		Tags:       r.Tags.Clone(),
		Members:    make([]b6.RelationMember, len(r.Members)),
	}
	for i, member := range r.Members {
		clone.Members[i] = member
	}
	return clone
}

func (r *RelationFeature) MergeFrom(other Feature) {
	if relation, ok := other.(*RelationFeature); ok {
		r.MergeFromRelationFeature(relation)
	} else {
		panic(fmt.Sprintf("Expected a RelationFeature, found %T", other))
	}
}

func (r *RelationFeature) MergeFromRelationFeature(other *RelationFeature) {
	r.RelationID = other.RelationID
	r.Tags.MergeFrom(other.Tags)
	i := copy(r.Members, other.Members)
	if i < len(other.Members) {
		r.Members = append(r.Members, other.Members[i:]...)
	} else {
		r.Members = r.Members[0:len(other.Members)]
	}
}

type PathPosition struct {
	Path     *PathFeature
	Position int
}

func (p *PathPosition) IsFirstOrLast() bool {
	return p.Position == 0 || p.Position == p.Path.Len()-1
}

type FeaturesByID struct {
	Points    map[b6.PointID]*PointFeature
	Paths     map[b6.PathID]*PathFeature
	Areas     map[b6.AreaID]*AreaFeature
	Relations map[b6.RelationID]*RelationFeature

	PointsFromOSMNodes map[uint64]s2.LatLng
}

func NewFeaturesByID() *FeaturesByID {
	f := &FeaturesByID{
		Points:    make(map[b6.PointID]*PointFeature),
		Paths:     make(map[b6.PathID]*PathFeature),
		Areas:     make(map[b6.AreaID]*AreaFeature),
		Relations: make(map[b6.RelationID]*RelationFeature),
	}
	f.PointsFromOSMNodes = make(map[uint64]s2.LatLng)
	return f
}

func (f *FeaturesByID) AddSimplePoint(id b6.PointID, ll s2.LatLng) {
	if id.Namespace == b6.NamespaceOSMNode {
		f.PointsFromOSMNodes[id.Value] = ll
		delete(f.Points, id)
	} else {
		f.Points[id] = NewPointFeature(id, ll)
	}
}

func (f *FeaturesByID) AddFeature(feature Feature) {
	switch feature := feature.(type) {
	case *PointFeature:
		if feature.FeatureID().Namespace == b6.NamespaceOSMNode {
			delete(f.PointsFromOSMNodes, feature.FeatureID().Value)
		}
		f.Points[feature.PointID] = feature
	case *PathFeature:
		f.Paths[feature.PathID] = feature
	case *AreaFeature:
		f.Areas[feature.AreaID] = feature
	case *RelationFeature:
		f.Relations[feature.RelationID] = feature
	}
}

func (f *FeaturesByID) findSimplePointByID(id b6.FeatureID) (s2.LatLng, bool) {
	if id.Namespace == b6.NamespaceOSMNode {
		ll, ok := f.PointsFromOSMNodes[id.Value]
		return ll, ok
	}
	return s2.LatLng{}, false
}

func (f *FeaturesByID) FindFeatureByID(id b6.FeatureID) b6.Feature {
	switch id.Type {
	case b6.FeatureTypePoint:
		if ll, ok := f.findSimplePointByID(id); ok {
			return newPointFeature(NewPointFeature(id.ToPointID(), ll))
		}
		if p, ok := f.Points[id.ToPointID()]; ok {
			return newPointFeature(p)
		}
	case b6.FeatureTypePath:
		if p, ok := f.Paths[id.ToPathID()]; ok {
			return newPathFeature(p, f)
		}
	case b6.FeatureTypeArea:
		if a, ok := f.Areas[id.ToAreaID()]; ok {
			return newAreaFeature(a, f)
		}
	case b6.FeatureTypeRelation:
		if r, ok := f.Relations[id.ToRelationID()]; ok {
			return newRelationFeature(r, f)
		}
	}
	return nil
}

func (f *FeaturesByID) HasFeatureWithID(id b6.FeatureID) bool {
	switch id.Type {
	case b6.FeatureTypePoint:
		if _, ok := f.findSimplePointByID(id); ok {
			return true
		}
		_, ok := f.Points[id.ToPointID()]
		return ok
	case b6.FeatureTypePath:
		_, ok := f.Paths[id.ToPathID()]
		return ok
	case b6.FeatureTypeArea:
		_, ok := f.Areas[id.ToAreaID()]
		return ok
	case b6.FeatureTypeRelation:
		_, ok := f.Relations[id.ToRelationID()]
		return ok
	}
	return false
}

func (f *FeaturesByID) FindMutableFeatureByID(id b6.FeatureID) Feature {
	switch id.Type {
	case b6.FeatureTypePoint:
		if ll, ok := f.findSimplePointByID(id); ok {
			mutable := NewPointFeature(id.ToPointID(), ll)
			f.Points[id.ToPointID()] = mutable
			delete(f.PointsFromOSMNodes, id.Value)
			return mutable
		}
		if p, ok := f.Points[id.ToPointID()]; ok {
			return p
		}
	case b6.FeatureTypePath:
		if p, ok := f.Paths[id.ToPathID()]; ok {
			return p
		}
	case b6.FeatureTypeArea:
		if a, ok := f.Areas[id.ToAreaID()]; ok {
			return a
		}
	case b6.FeatureTypeRelation:
		if r, ok := f.Relations[id.ToRelationID()]; ok {
			return r
		}
	}
	return nil
}

func (f *FeaturesByID) FindLocationByID(id b6.PointID) (s2.LatLng, bool) {
	if ll, ok := f.findSimplePointByID(id.FeatureID()); ok {
		return ll, true
	}
	if p, ok := f.Points[id]; ok {
		return p.Location, true
	}
	return s2.LatLng{}, false
}

func (f *FeaturesByID) EachFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	cores := options.Cores
	if cores < 1 {
		cores = 1
	}

	var cause error
	var wg sync.WaitGroup
	features := make(chan b6.Feature, cores)
	ctx, cancel := context.WithCancel(context.Background())
	feed := func(goroutine int) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case feature, ok := <-features:
				if ok {
					if err := each(feature, goroutine); err != nil {
						cause = err
						cancel()
					}
				} else {
					return
				}
			}
		}
	}

	wg.Add(cores)
	for i := 0; i < cores; i++ {
		go feed(i)
	}

	if !options.SkipPoints {
		for _, point := range f.Points {
			features <- newPointFeature(point)
		}
	}
	if !options.SkipPaths {
		for _, path := range f.Paths {
			features <- newPathFeature(path, f)
		}
	}
	if !options.SkipAreas {
		for _, area := range f.Areas {
			features <- newAreaFeature(area, f)
		}
	}
	if !options.SkipRelations {
		for _, relation := range f.Relations {
			features <- newRelationFeature(relation, f)
		}
	}
	close(features)
	wg.Wait()
	return cause
}

type FeatureReferences struct {
	RelationsByFeature map[b6.FeatureID][]*RelationFeature
	PathsByPoint       map[b6.PointID][]PathPosition
	AreasByPath        map[b6.PathID][]*AreaFeature
}

func NewFeatureReferences() *FeatureReferences {
	return &FeatureReferences{
		RelationsByFeature: make(map[b6.FeatureID][]*RelationFeature),
		PathsByPoint:       make(map[b6.PointID][]PathPosition),
		AreasByPath:        make(map[b6.PathID][]*AreaFeature),
	}
}

func NewFilledFeatureReferences(byID *FeaturesByID) *FeatureReferences {
	f := NewFeatureReferences()
	for _, path := range byID.Paths {
		f.AddPointsForPath(path)
	}

	for _, area := range byID.Areas {
		f.AddPathsForArea(area)
	}

	for _, relation := range byID.Relations {
		f.AddRelationsForFeature(relation)
	}
	return f
}

func (f *FeatureReferences) AddFeature(feature Feature, byID b6.FeaturesByID) {
	switch feature := feature.(type) {
	case *PathFeature:
		f.AddPointsForPath(feature)
	case *AreaFeature:
		f.AddPathsForArea(feature)
	case *RelationFeature:
		f.AddRelationsForFeature(feature)
	}
}

func (f *FeatureReferences) RemoveFeature(feature Feature, byID b6.FeaturesByID) {
	switch feature := feature.(type) {
	case *PathFeature:
		f.RemovePointsForPath(feature)
	case *AreaFeature:
		f.RemovePathsForArea(feature)
	case *RelationFeature:
		f.RemoveRelationsForFeature(feature)
	}
}

// AreasForPoint returns the areas that reference the point id. The world w
// is used to find the paths that reference the point, before finding the
// areas that reference those paths.
func (f *FeatureReferences) AreasForPoint(id b6.PointID, w b6.World) []*AreaFeature {
	areasSeen := make(map[b6.AreaID]*AreaFeature)
	paths := w.FindPathsByPoint(id)
	for paths.Next() {
		for _, a := range f.AreasForPath(paths.FeatureID().ToPathID()) {
			areasSeen[a.AreaID] = a
		}
	}
	areas := make([]*AreaFeature, 0, len(areasSeen))
	for _, a := range areasSeen {
		areas = append(areas, a)
	}
	return areas
}

func (f *FeatureReferences) AreasForPath(id b6.PathID) []*AreaFeature {
	if areas, ok := f.AreasByPath[id]; ok {
		return areas
	}
	return []*AreaFeature{}
}

func (f *FeatureReferences) RemovePointsForPath(p *PathFeature) {
	for i := 0; i < p.Len(); i++ {
		if id, ok := p.PointID(i); ok {
			if paths := f.PathsByPoint[id]; paths != nil && len(paths) > 0 {
				read := 0
				write := 0
				for read < len(paths) {
					if paths[read].Path == p {
						read++
						continue
					}
					if read != write {
						paths[write] = paths[read]
					}
					read++
					write++
				}
				f.PathsByPoint[id] = paths[0:write]
			}
		}
	}
}

func (f *FeatureReferences) AddPointsForPath(p *PathFeature) {
	for i := 0; i < p.Len(); i++ {
		if id, ok := p.PointID(i); ok {
			paths, ok := f.PathsByPoint[id]
			if !ok {
				paths = make([]PathPosition, 0, 1)
			}
			seen := false
			for j := 0; j < len(paths); j++ {
				if paths[j].Path == p {
					paths[j].Position = i
					seen = true
					break
				}
			}
			if !seen {
				f.PathsByPoint[id] = append(paths, PathPosition{Path: p, Position: i})
			}
		}
	}
}

func (f *FeatureReferences) RemovePathsForArea(a *AreaFeature) {
	for i := 0; i < a.Len(); i++ {
		if paths, ok := a.PathIDs(i); ok {
			for _, id := range paths {
				if areas := f.AreasByPath[id]; areas != nil && len(areas) > 0 {
					read := 0
					write := 0
					for read < len(areas) {
						if areas[read] == a {
							read++
							continue
						}
						if read != write {
							areas[write] = areas[read]
						}
						read++
						write++
					}
					f.AreasByPath[id] = areas[0:write]
				}
			}
		}
	}
}

func (f *FeatureReferences) AddPathsForArea(a *AreaFeature) {
	for i := 0; i < a.Len(); i++ {
		if ids, ok := a.PathIDs(i); ok {
			for _, id := range ids {
				areas, ok := f.AreasByPath[id]
				if !ok {
					areas = make([]*AreaFeature, 0, 1)
				}
				f.AreasByPath[id] = append(areas, a)
			}
		}
	}
}

func (f *FeatureReferences) RemoveRelationsForFeature(r *RelationFeature) {
	for _, member := range r.Members {
		ids := f.RelationsByFeature[member.ID]
		filtered := make([]*RelationFeature, 0, len(ids)-1)
		for _, relation := range ids {
			if relation != r {
				filtered = append(filtered, relation)
			}
		}
		f.RelationsByFeature[member.ID] = filtered
	}
}

func (f *FeatureReferences) AddRelationsForFeature(r *RelationFeature) {
	for _, member := range r.Members {
		ids, ok := f.RelationsByFeature[member.ID]
		if !ok {
			ids = make([]*RelationFeature, 0, 1)
		}
		f.RelationsByFeature[member.ID] = append(ids, r)
	}
}

type ReadOptions struct {
	SkipPoints    bool
	SkipPaths     bool
	SkipAreas     bool
	SkipRelations bool
	SkipTags      bool
	Cores         int
}

type Emit func(f Feature, goroutine int) error

type FeatureSource interface {
	Read(options ReadOptions, emit Emit, ctx context.Context) error
}

type WorldFeatureSource struct {
	World b6.World
}

func (w WorldFeatureSource) Read(options ReadOptions, emit Emit, ctx context.Context) error {
	o := b6.EachFeatureOptions{
		SkipPoints:    options.SkipPoints,
		SkipPaths:     options.SkipPaths,
		SkipAreas:     options.SkipAreas,
		SkipRelations: options.SkipRelations,
		Cores:         options.Cores,
	}
	f := func(f b6.Feature, goroutine int) error {
		return emit(NewFeatureFromWorld(f), goroutine)
	}
	return w.World.EachFeature(f, &o)
}

type MemoryFeatureSource []Feature

func (m MemoryFeatureSource) Read(options ReadOptions, emit Emit, ctx context.Context) error {
	cores := options.Cores
	if cores < 1 {
		cores = 1
	}
	c := make(chan Feature, cores)
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	var cause error
	feed := func(goroutine int) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case f, ok := <-c:
				if ok {
					if err := emit(f, goroutine); err != nil {
						cause = err
						cancel()
					}
				} else {
					return
				}
			}
		}
	}

	wg.Add(cores)
	for i := 0; i < cores; i++ {
		go feed(i)
	}
	for _, f := range m {
		c <- f
	}
	close(c)
	wg.Wait()
	return cause
}
