package ingest

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/geometry"
	"diagonal.works/b6/osm"
	"github.com/golang/geo/s2"
	"golang.org/x/sync/errgroup"
)

type ByTagKey b6.Tags

func (t ByTagKey) Len() int           { return len(t) }
func (t ByTagKey) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t ByTagKey) Less(i, j int) bool { return t[i].Key < t[j].Key }

type Feature interface {
	b6.Feature

	SetFeatureID(id b6.FeatureID)
	SetTags(tags []b6.Tag)
	AddTag(tag b6.Tag)
	ModifyOrAddTag(tag b6.Tag) (bool, b6.Expression)
	ModifyOrAddTagAt(tag b6.Tag, index int) (bool, b6.Expression)
	RemoveTag(key string)
	RemoveTags(keys []string)
	RemoveAllTags()

	Clone() Feature
	MergeFrom(other Feature)
}

func NewFeatureFromWorld(f b6.Feature) Feature {
	switch f.FeatureID().Type {
	case b6.FeatureTypeArea:
		return NewAreaFeatureFromWorld(f.(b6.AreaFeature))
	case b6.FeatureTypeRelation:
		return NewRelationFeatureFromWorld(f.(b6.RelationFeature))
	case b6.FeatureTypeCollection:
		return NewCollectionFeatureFromWorld(f.(b6.CollectionFeature))
	case b6.FeatureTypeExpression:
		return NewExpressionFeatureFromWorld(f.(b6.ExpressionFeature))
	default:
		return NewGenericFeatureFromWorld(f)
	}
	panic(fmt.Sprintf("Can't handle feature type: %T", f))
}

func WrapFeature(f Feature, byID b6.FeaturesByID) b6.Feature {
	switch f.FeatureID().Type {
	case b6.FeatureTypePoint, b6.FeatureTypePath:
		return b6.WrapPhysicalFeature(f.(b6.PhysicalFeature), byID)
	case b6.FeatureTypeArea:
		return WrapAreaFeature(f.(*AreaFeature), byID)
	case b6.FeatureTypeRelation:
		return WrapRelationFeature(f.(*RelationFeature), byID)
	case b6.FeatureTypeCollection:
		return WrapCollectionFeature(f.(*CollectionFeature), byID)
	case b6.FeatureTypeExpression:
		return WrapExpressionFeature(f.(*ExpressionFeature))
	}
	panic(fmt.Sprintf("Can't wrap feature type: %T", f))
}

type GenericFeature struct {
	b6.ID
	b6.Tags
}

func (f *GenericFeature) Clone() Feature {
	return &GenericFeature{
		ID:   f.ID.FeatureID(),
		Tags: f.Tags.Clone(),
	}
}

func (f *GenericFeature) MergeFrom(other Feature) {
	f.SetFeatureID(other.FeatureID())
	f.Tags.MergeFrom(other.AllTags().Clone())
}

func (f *GenericFeature) MarshalYAML() (interface{}, error) {
	y := map[string]interface{}{
		"id": f.ID,
	}

	if len(f.Tags) > 0 {
		y["tags"] = f.Tags
	}
	return y, nil
}

type OSMFeature struct {
	*osm.Node
	*osm.Way
	ClosedWay bool
}

func (f *GenericFeature) FillFromOSM(o OSMFeature) {
	if o.Node != nil {
		f.SetFeatureID(FromOSMNodeID(o.Node.ID))
		FillTagsFromOSM(&f.Tags, o.Node.Tags)

		f.ModifyOrAddTag(b6.Tag{Key: b6.PointTag, Value: b6.NewPointExpressionFromLatLng(o.Node.Location.ToS2LatLng())})
	} else if o.Way != nil {
		f.SetFeatureID(FromOSMWayID(o.Way.ID))
		FillTagsFromOSM(&f.Tags, o.Way.Tags)
		if o.ClosedWay {
			f.Tags = f.Tags[0:0]
		}

		points := make([]b6.AnyExpression, 0, len(o.Way.Nodes))
		for _, id := range o.Way.Nodes {
			points = append(points, b6.FeatureIDExpression(FromOSMNodeID(id)))
		}
		f.ModifyOrAddTag(b6.Tag{Key: b6.PathTag, Value: b6.NewExpressions(points)})
	}
}

func NewGenericFeatureFromWorld(f b6.Feature) *GenericFeature {
	return &GenericFeature{
		ID:   f.FeatureID(),
		Tags: f.AllTags().Clone(),
	}
}

func (f *GenericFeature) ToGeoJSON() geojson.GeoJSON {
	return b6.PhysicalFeatureToGeoJSON(f)
}

type AreaMembers struct {
	ids      [][]b6.FeatureID
	polygons []*s2.Polygon
}

func NewAreaMembers(n int) AreaMembers {
	return AreaMembers{
		ids:      make([][]b6.FeatureID, n),
		polygons: make([]*s2.Polygon, n),
	}
}

func (a *AreaMembers) Len() int {
	return len(a.ids)
}

func (a *AreaMembers) SetPathIDs(i int, ids []b6.FeatureID) {
	a.ids[i] = ids
	a.polygons[i] = nil
}

func (a *AreaMembers) SetPathID(i int, j int, id b6.FeatureID) {
	for len(a.ids[i]) <= j {
		a.ids[i] = append(a.ids[i], b6.FeatureIDInvalid)
	}
	a.ids[i][j] = id
	a.polygons[i] = nil
}

func (a *AreaMembers) PathIDs(i int) ([]b6.FeatureID, bool) {
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
		ids:      make([][]b6.FeatureID, len(a.ids)),
		polygons: make([]*s2.Polygon, len(a.polygons)),
	}
	copy(clone.ids, a.ids)
	copy(clone.polygons, a.polygons)
	return clone
}

func (a *AreaMembers) MergeFrom(other AreaMembers) {
	if len(a.ids) < len(other.ids) {
		for i := len(a.ids); i < len(other.ids); i++ {
			a.ids = append(a.ids, make([]b6.FeatureID, len(other.ids[i])))
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
	a.ids = append(a.ids, []b6.FeatureID{FromOSMWayID(way.ID)})
	a.polygons = a.polygons[0:0]
	a.polygons = append(a.polygons, nil)
}

type AreaFeature struct {
	b6.AreaID
	b6.Tags
	AreaMembers
}

func NewAreaFeature(n int) *AreaFeature {
	return &AreaFeature{AreaMembers: NewAreaMembers(n)}
}

func NewAreaFeatureFromOSMWay(way *osm.Way) (*AreaFeature, *GenericFeature) {
	path := GenericFeature{}
	path.FillFromOSM(OSMFeature{Way: way})
	path.Tags = make(b6.Tags, 0) // Tags only exist on the area feature
	area := &AreaFeature{}
	area.FillFromOSMWay(way)
	area.SetPathIDs(0, []b6.FeatureID{path.ID})
	return area, &path
}

func NewAreaFeatureFromWorld(a b6.AreaFeature) *AreaFeature {
	area := NewAreaFeature(a.Len())
	area.AreaID = a.AreaID()
	area.Tags = a.AllTags().Clone()
	for i := 0; i < a.Len(); i++ {
		if paths := a.Feature(i); paths != nil {
			ids := make([]b6.FeatureID, len(paths))
			for j, path := range paths {
				ids[j] = path.FeatureID()
			}
			area.SetPathIDs(i, ids)
		} else {
			area.SetPolygon(i, a.Polygon(i))
		}
	}
	return area
}

func NewAreaFeatureFromRegion(g s2.Region) (*AreaFeature, error) {
	switch g := g.(type) {
	case *s2.Polygon:
		area := NewAreaFeature(1)
		area.SetPolygon(0, g)
		return area, nil
	case geometry.MultiPolygon:
		area := NewAreaFeature(len(g))
		for i, polygon := range g {
			area.SetPolygon(i, polygon)
		}
		return area, nil
	default:
		return nil, fmt.Errorf("Can't convert geometry type %T", g)
	}
}

func NewAreaFeatureFromS2Cell(id b6.AreaID, cell s2.CellID) *AreaFeature {
	area := NewAreaFeature(1)
	area.AreaID = id
	area.SetPolygon(0, s2.PolygonFromLoops([]*s2.Loop{s2.LoopFromCell(s2.CellFromCellID(cell))}))
	return area
}

func (a *AreaFeature) SetFeatureID(id b6.FeatureID) {
	a.AreaID = id.ToAreaID()
}

func (a *AreaFeature) FillFromOSMWay(way *osm.Way) {
	a.AreaID = AreaIDFromOSMWayID(way.ID)
	FillTagsFromOSM(&a.Tags, way.Tags)
	a.AreaMembers.FillFromOSMWay(way)
}

func (a *AreaFeature) References() []b6.Reference {
	references := make([]b6.Reference, 0, a.Len())
	for i := 0; i < a.Len(); i++ {
		if ids, ok := a.PathIDs(i); ok {
			for _, id := range ids {
				references = append(references, id)
			}
		}
	}

	return references
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

func (a *AreaFeature) MarshalYAML() (interface{}, error) {
	polygonsYAML := make([]interface{}, a.Len())
	for i := range polygonsYAML {
		if p, ok := a.AreaMembers.Polygon(i); ok {
			loopsYAML := make([]interface{}, p.NumLoops())
			for j := range loopsYAML {
				loop := p.Loop(j)
				loopYAML := make([]interface{}, loop.NumVertices())
				for k := range loopYAML {
					loopYAML[k] = LatLngYAML{LatLng: s2.LatLngFromPoint(loop.Vertex(k))}
				}
				loopsYAML[j] = loopYAML
			}
			polygonsYAML[i] = loopsYAML
		} else {
			pathIDs, _ := a.AreaMembers.PathIDs(i)
			loopsYAML := make([]interface{}, len(pathIDs))
			for j := range loopsYAML {
				loopsYAML[j] = pathIDs[j]
			}
			polygonsYAML[i] = loopsYAML
		}
	}
	y := map[string]interface{}{
		"id":   a.AreaID,
		"area": polygonsYAML,
	}
	if len(a.Tags) > 0 {
		y["tags"] = a.Tags
	}
	return y, nil
}

func (a *AreaFeature) GeometryType() b6.GeometryType {
	return b6.GeometryTypeArea
}

type RelationFeature struct {
	b6.RelationID
	b6.Tags
	Members []b6.RelationMember
}

func NewRelationFeature(n int) *RelationFeature {
	return &RelationFeature{Members: make([]b6.RelationMember, n)}
}

func NewRelationFeatureFromWorld(r b6.RelationFeature) *RelationFeature {
	relation := NewRelationFeature(r.Len())
	relation.RelationID = r.RelationID()
	relation.Tags = r.AllTags().Clone()
	for i := 0; i < r.Len(); i++ {
		relation.Members[i] = r.Member(i)
	}
	return relation
}

func (r *RelationFeature) SetFeatureID(id b6.FeatureID) {
	r.RelationID = id.ToRelationID()
}

func (r *RelationFeature) Len() int {
	return len(r.Members)
}

func (r *RelationFeature) Member(i int) b6.RelationMember {
	return r.Members[i]
}

func (r *RelationFeature) References() []b6.Reference {
	references := make([]b6.Reference, 0, len(r.Members))
	for _, member := range r.Members {
		references = append(references, member.ID)
	}

	return references
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

func (r *RelationFeature) MarshalYAML() (interface{}, error) {
	y := map[string]interface{}{
		"id":       r.RelationID,
		"relation": r.Members,
	}
	if len(r.Tags) > 0 {
		y["tags"] = r.Tags
	}
	return y, nil
}

type CollectionFeature struct {
	b6.CollectionID
	b6.Tags

	Keys   []interface{}
	Values []interface{}
	sorted bool
}

func (c *CollectionFeature) References() []b6.Reference {
	references := make([]b6.Reference, 0, len(c.Keys))
	for _, key := range c.Keys {
		if id, ok := key.(b6.Identifiable); ok {
			references = append(references, id.FeatureID())
		}
	}
	return references
}

func (c *CollectionFeature) Clone() Feature {
	return &CollectionFeature{
		CollectionID: c.CollectionID,
		Tags:         c.Tags.Clone(),
		Keys:         c.Keys,
		Values:       c.Values,
		sorted:       c.sorted,
	}
}

func (c *CollectionFeature) MergeFrom(other Feature) {
	if collection, ok := other.(*CollectionFeature); ok {
		c.MergeFromCollectionFeature(collection)
	} else {
		panic(fmt.Sprintf("Expected a CollectionFeature, found %T", other))
	}
}

func (c *CollectionFeature) MergeFromCollectionFeature(other *CollectionFeature) {
	c.CollectionID = other.CollectionID
	c.Tags = other.Tags
	c.Keys = other.Keys
	c.Values = other.Values
	c.sorted = other.sorted
}

func (c *CollectionFeature) SetFeatureID(id b6.FeatureID) {
	c.CollectionID = id.ToCollectionID()
}

func NewCollectionFeatureFromWorld(c b6.CollectionFeature) *CollectionFeature {
	feature := &CollectionFeature{
		CollectionID: c.CollectionID(),
		Tags:         c.AllTags().Clone(),
		sorted:       c.IsSortedByKey(),
	}

	i := c.BeginUntyped()
	for {
		ok, err := i.Next()
		if !ok || err != nil {
			break
		}

		feature.Keys = append(feature.Keys, i.Key())
		feature.Values = append(feature.Values, i.Value())
	}

	return feature
}

func (c *CollectionFeature) IsSortedByKey() bool {
	return c.sorted
}

func (c *CollectionFeature) FindValue(key any) (any, bool) {
	// TODO: we shouldn't need to cast the keys and values
	// on every comparison, since they're all the same time.
	// We could replace Keys and Values with a generic type,
	// building from, eg, b6.ArrayCollection
	if c.sorted {
		i := sort.Search(len(c.Keys), func(i int) bool {
			greater, _ := b6.Less(c.Keys[i], key)
			return !greater
		})
		if i < len(c.Keys) {
			if equal, _ := b6.Equal(c.Keys[i], key); equal {
				return c.Values[i], true
			}
		}
	} else {
		for i := range c.Keys {
			if equal, _ := b6.Equal(c.Keys[i], key); equal {
				return c.Values[i], true
			}
		}
	}
	return nil, false
}

func (c *CollectionFeature) FindValues(key any, values []any) []any {
	if c.sorted {
		i := sort.Search(len(c.Keys), func(i int) bool {
			greater, _ := b6.Less(c.Keys[i], key)
			return !greater
		})
		for i < len(c.Keys) {
			if equal, _ := b6.Equal(c.Keys[i], key); equal {
				values = append(values, c.Values[i])
			} else {
				break
			}
			i++
		}
	} else {
		for i := range c.Keys {
			if equal, _ := b6.Equal(c.Keys[i], key); equal {
				values = append(values, c.Values[i])
			}
		}
	}
	return values
}

func (c *CollectionFeature) Sort() {
	sort.Sort(byCollectionFeatureKey(*c))
	c.sorted = true
}

type byCollectionFeatureKey CollectionFeature

func (b byCollectionFeatureKey) Len() int {
	return len(b.Keys)
}

func (b byCollectionFeatureKey) Swap(i, j int) {
	b.Keys[i], b.Keys[j] = b.Keys[j], b.Keys[i]
	b.Values[i], b.Values[j] = b.Values[j], b.Values[i]
}

func (b byCollectionFeatureKey) Less(i, j int) bool {
	less, _ := b6.Less(b.Keys[i], b.Keys[j])
	return less
}

func (c *CollectionFeature) MarshalYAML() (interface{}, error) {
	e := b6.CollectionExpression{
		UntypedCollection: b6.ArrayCollection[any, any]{
			Keys:   c.Keys,
			Values: c.Values,
		}.Collection(),
	}

	y := map[string]interface{}{
		"id":         c.CollectionID,
		"collection": e,
	}

	if len(c.Tags) > 0 {
		y["tags"] = c.Tags
	}

	return y, nil
}

type ExpressionFeature struct {
	b6.ExpressionID
	b6.Tags
	b6.Expression
}

func (c *ExpressionFeature) References() []b6.Reference {
	return []b6.Reference{}
}

func (e *ExpressionFeature) Clone() Feature {
	return &ExpressionFeature{
		ExpressionID: e.ExpressionID,
		Tags:         e.Tags.Clone(),
		Expression:   e.Expression.Clone(),
	}
}

func (e *ExpressionFeature) SetFeatureID(id b6.FeatureID) {
	e.ExpressionID = id.ToExpressionID()
}

func (e *ExpressionFeature) MergeFrom(other Feature) {
	if expression, ok := other.(*ExpressionFeature); ok {
		e.MergeFromExpressionFeature(expression)
	} else {
		panic(fmt.Sprintf("Expected an ExpressionFeature, found %T", other))
	}
}

func (e *ExpressionFeature) MergeFromExpressionFeature(other *ExpressionFeature) {
	e.ExpressionID = other.ExpressionID
	e.Tags.MergeFrom(other.Tags)
	e.Expression = other.Expression.Clone()
}

func (e *ExpressionFeature) MarshalYAML() (interface{}, error) {
	y := map[string]interface{}{
		"id":         e.ExpressionID,
		"expression": e.Expression,
	}
	if len(e.Tags) > 0 {
		y["tags"] = e.Tags
	}
	return y, nil
}

func (e *ExpressionFeature) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return e.Expression.UnmarshalYAML(unmarshal)
}

func NewExpressionFeatureFromWorld(e b6.ExpressionFeature) *ExpressionFeature {
	return &ExpressionFeature{
		ExpressionID: e.ExpressionID(),
		Tags:         e.AllTags().Clone(),
		Expression:   e.Expression().Clone(),
	}
}

type FeaturesByID map[b6.FeatureID]Feature

func NewFeaturesByID() *FeaturesByID {
	f := FeaturesByID(make(map[b6.FeatureID]Feature))
	return &f
}

func (f *FeaturesByID) AddFeature(feature Feature) {
	(*f)[feature.FeatureID()] = feature
}

func (f *FeaturesByID) FindFeatureByID(id b6.FeatureID) b6.Feature {
	if feature, ok := (*f)[id]; ok {
		return WrapFeature(feature, f)
	}

	return nil
}

func (f *FeaturesByID) HasFeatureWithID(id b6.FeatureID) bool {
	_, ok := (*f)[id]
	return ok
}

func (f *FeaturesByID) FindMutableFeatureByID(id b6.FeatureID) Feature {
	if feature, ok := (*f)[id]; ok {
		return feature
	}

	return nil
}

func (f *FeaturesByID) FindLocationByID(id b6.FeatureID) (s2.LatLng, error) {
	if feature, ok := (*f)[id.FeatureID()]; ok {
		if f, ok := feature.(b6.Geometry); ok && f.GeometryType() == b6.GeometryTypePoint {
			return s2.LatLngFromPoint(f.Point()), nil
		}
	}

	return s2.LatLng{}, fmt.Errorf("location for feature with %s id not found", id.String())
}

func EachFeature(each func(f b6.Feature, goroutine int) error, f *FeaturesByID, r *FeatureReferencesByID, options *b6.EachFeatureOptions) error {
	wrap := func(feature Feature, goroutine int) error {
		return each(WrapFeature(feature, f), goroutine)
	}
	return eachIngestFeature(wrap, f, r, options)
}

func eachIngestFeature(each func(f Feature, goroutine int) error, f *FeaturesByID, r *FeatureReferencesByID, options *b6.EachFeatureOptions) error {
	goroutines := options.Goroutines
	if goroutines < 1 {
		goroutines = 1
	}

	c := make(chan Feature, goroutines)
	g, gc := errgroup.WithContext(context.Background())
	for i := 0; i < goroutines; i++ {
		g.Go(func() error {
			for tag := range c {
				if err := each(tag, i); err != nil {
					return err
				}
			}
			return nil
		})
	}

	feedFeatures(c, gc.Done(), f, r, options)
	close(c)
	return g.Wait()
}

func feedFeatures(c chan<- Feature, done <-chan struct{}, f *FeaturesByID, r *FeatureReferencesByID, options *b6.EachFeatureOptions) {
	features := make([]Feature, 0, len(*f))
	for _, feature := range *f {
		features = append(features, feature)
	}

	if options.FeedReferencesFirst {
		sort.Slice(features, func(i, j int) bool {
			return len(r.FindReferences(features[i].FeatureID())) > len(r.FindReferences(features[j].FeatureID()))
		})
	}

	for _, feature := range features {
		if options.SkipPoints && feature.FeatureID().Type == b6.FeatureTypePoint || // TODO(mari): make this nicer / like opts r types so we can one-line this
			options.SkipPaths && feature.FeatureID().Type == b6.FeatureTypePath ||
			options.SkipAreas && feature.FeatureID().Type == b6.FeatureTypeArea ||
			options.SkipRelations && feature.FeatureID().Type == b6.FeatureTypeRelation ||
			options.SkipCollections && feature.FeatureID().Type == b6.FeatureTypeCollection ||
			options.SkipExpressions && feature.FeatureID().Type == b6.FeatureTypeExpression {
			continue
		}

		select {
		case <-done:
			return
		case c <- feature:
		}
	}
}

type FeatureReferencesByID map[b6.FeatureID][]b6.Reference

func NewFeatureReferences() *FeatureReferencesByID {
	f := FeatureReferencesByID(make(map[b6.FeatureID][]b6.Reference))
	return &f
}

func NewFilledFeatureReferences(byID *FeaturesByID) *FeatureReferencesByID {
	f := NewFeatureReferences()

	for _, feature := range *byID {
		f.AddFeature(feature)
	}

	return f
}

func (f *FeatureReferencesByID) findReferences(id b6.FeatureID, m *map[b6.Reference]bool) {
	if references, ok := (*f)[id]; ok {
		for _, reference := range references {
			(*m)[reference] = true
			f.findReferences(reference.Source(), m)
		}
	}
}

func (f *FeatureReferencesByID) FindReferences(id b6.FeatureID, typed ...b6.FeatureType) []b6.Reference {
	m := make(map[b6.Reference]bool)
	f.findReferences(id, &m)

	var references []b6.Reference
	for reference := range m {
		if len(typed) == 0 || slices.Contains(typed, reference.Source().Type) {
			references = append(references, reference)
		}
	}

	return references
}

func (f *FeatureReferencesByID) AddFeature(feature Feature) {
	for index, reference := range feature.References() {
		references, ok := (*f)[reference.Source()]
		if !ok {
			references = make([]b6.Reference, 0, 1)
		}

		referenced := false
		for i := range references {
			if references[i].Source() == feature.FeatureID() {
				referenced = true
				if reference, ok := (*f)[reference.Source()][i].(b6.IndexedReference); ok { // Overwrite the index to deal with paths, which have duplicate start/end point.
					reference.SetIndex(index) // Only applies to paths atm; we may need to limit if we introduce other indexed references. TODO: consider moving this logic further down.
				}
			}
		}

		if !referenced {
			if _, ok := reference.(b6.IndexedReference); ok {
				ref := b6.IndexedFeatureID{FeatureID: feature.FeatureID()}
				ref.SetIndex(index)
				(*f)[reference.Source()] = append(references, &ref)
			} else {
				(*f)[reference.Source()] = append(references, feature.FeatureID())
			}
		}
	}
}

func (f *FeatureReferencesByID) RemoveFeature(feature Feature) {
	for _, reference := range feature.References() {
		if references, ok := (*f)[reference.Source()]; ok {
			for i := range references {
				if references[i].Source() == feature.FeatureID() {
					if i == len(references)-1 {
						(*f)[reference.Source()] = (*f)[reference.Source()][:i]
					} else {
						(*f)[reference.Source()] = append((*f)[reference.Source()][:i], (*f)[reference.Source()][i+1:]...)
					}
				}
			}
		}
	}
}

type ReadOptions struct {
	SkipPoints      bool
	SkipPaths       bool
	SkipAreas       bool
	SkipRelations   bool
	SkipCollections bool
	SkipTags        bool
	Goroutines      int
}

type WorldFeatureSource struct {
	World b6.World
}

func (w WorldFeatureSource) Read(options ReadOptions, emit Emit, ctx context.Context) error {
	o := b6.EachFeatureOptions{
		SkipPoints:      options.SkipPoints,
		SkipPaths:       options.SkipPaths,
		SkipAreas:       options.SkipAreas,
		SkipRelations:   options.SkipRelations,
		SkipCollections: options.SkipCollections,
		Goroutines:      options.Goroutines,
	}
	f := func(f b6.Feature, goroutine int) error {
		return emit(NewFeatureFromWorld(f), goroutine)
	}
	return w.World.EachFeature(f, &o)
}

type MemoryFeatureSource []Feature

func (m MemoryFeatureSource) Read(options ReadOptions, emit Emit, ctx context.Context) error {
	cores := options.Goroutines
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
