package ingest

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/geometry"
	"diagonal.works/b6/search"
	"github.com/golang/geo/s2"
)

type featureValues struct{}

func (f featureValues) Compare(a search.Value, b search.Value) search.Comparison {
	ida, idb := a.(Feature).FeatureID(), b.(Feature).FeatureID()
	if ida.Less(idb) {
		return search.ComparisonLess
	} else if ida == idb {
		return search.ComparisonEqual
	}
	return search.ComparisonGreater
}

func (f featureValues) CompareKey(v search.Value, k search.Key) search.Comparison {
	ida, idb := v.(Feature).FeatureID(), k.(b6.FeatureID)
	if ida.Less(idb) {
		return search.ComparisonLess
	} else if ida == idb {
		return search.ComparisonEqual
	}
	return search.ComparisonGreater
}

func (f featureValues) Key(v search.Value) search.Key {
	return v.(Feature).FeatureID()
}

type FeatureIndex struct {
	search.ArrayIndex
	features b6.FeaturesByID
}

func NewFeatureIndex(features b6.FeaturesByID) *FeatureIndex {
	return &FeatureIndex{ArrayIndex: *search.NewArrayIndex(featureValues{}), features: features}
}

func (f *FeatureIndex) Feature(v search.Value) b6.Feature {
	if feature, ok := v.(Feature); ok {
		return WrapFeature(feature, f.features)
	}
	panic("Not a feature")
}

func (f *FeatureIndex) ID(v search.Value) b6.FeatureID {
	if i, ok := v.(b6.Identifiable); ok {
		return i.FeatureID()
	}
	panic(fmt.Sprintf("Bad feature type: %T", v))
}

type areaFeature struct {
	*AreaFeature
	features b6.FeaturesByID
}

func WrapAreaFeature(a *AreaFeature, byID b6.FeaturesByID) b6.AreaFeature {
	return areaFeature{a, byID}
}

func (a areaFeature) AreaID() b6.AreaID {
	return a.AreaFeature.AreaID
}

func (a areaFeature) Polygon(i int) *s2.Polygon {
	if polygon, ok := a.AreaFeature.Polygon(i); ok {
		return polygon
	}
	ids, _ := a.PathIDs(i)
	loops := make([]*s2.Loop, 0, len(ids))
	for _, id := range ids {
		if path, ok := a.features.FindFeatureByID(id).(b6.PhysicalFeature); path != nil && ok {
			loop := make([]s2.Point, path.GeometryLen()-1) // Paths are explicitly closed, so drop the duplicate last point
			for j := range loop {
				loop[j] = path.PointAt(j)
			}
			loops = append(loops, s2.LoopFromPoints(loop))
		}
	}
	return s2.PolygonFromLoops(loops)
}

func (a areaFeature) MultiPolygon() geometry.MultiPolygon {
	m := make(geometry.MultiPolygon, a.Len())
	for i := 0; i < a.Len(); i++ {
		m[i] = a.Polygon(i)
	}
	return m
}

func (a areaFeature) Feature(i int) []b6.PhysicalFeature {
	if ids, ok := a.PathIDs(i); ok {
		paths := make([]b6.PhysicalFeature, 0, len(ids))
		for _, id := range ids {
			if path := a.features.FindFeatureByID(id).(b6.PhysicalFeature); path != nil {
				paths = append(paths, path)
			}
		}
		return paths
	}
	return nil
}

func (a areaFeature) ToGeoJSON() geojson.GeoJSON {
	return b6.AreaFeatureToGeoJSON(a)
}

var _ b6.AreaFeature = (*areaFeature)(nil)

type relationFeature struct {
	*RelationFeature
	features b6.FeaturesByID
}

func WrapRelationFeature(r *RelationFeature, byID b6.FeaturesByID) b6.RelationFeature {
	return relationFeature{r, byID}
}

func (r relationFeature) RelationID() b6.RelationID {
	return r.RelationFeature.RelationID
}

func (r relationFeature) Feature(i int) b6.Feature {
	if feature := r.features.FindFeatureByID(r.Members[i].ID); feature != nil {
		return feature
	}
	return nil
}

func (r relationFeature) ToGeoJSON() geojson.GeoJSON {
	return b6.RelationFeatureToGeoJSON(r, r.features)
}

type collectionFeature struct {
	*CollectionFeature
	features b6.FeaturesByID
}

func WrapCollectionFeature(c *CollectionFeature, byID b6.FeaturesByID) b6.CollectionFeature {
	return collectionFeature{c, byID}
}

func (c collectionFeature) ToGeoJSON() geojson.GeoJSON {
	return b6.CollectionFeatureToGeoJSON(c, c.features)
}

func (c collectionFeature) Feature() b6.CollectionFeature {
	return c
}

func (c collectionFeature) CollectionID() b6.CollectionID {
	return c.CollectionFeature.CollectionID
}

func (c collectionFeature) BeginUntyped() b6.Iterator[any, any] {
	return &collectionFeatureIterator{c: c.CollectionFeature}
}

func (c collectionFeature) Count() (int, bool) {
	return len(c.Keys), true
}

type collectionFeatureIterator struct {
	c *CollectionFeature
	i int
}

func (c *collectionFeatureIterator) Next() (bool, error) {
	c.i++
	return c.i <= len(c.c.Keys), nil
}

func (c *collectionFeatureIterator) Key() interface{} {
	return c.c.Keys[c.i-1]
}

func (c collectionFeatureIterator) Value() interface{} {
	return c.c.Values[c.i-1]
}

func (c *collectionFeatureIterator) KeyExpression() b6.Expression {
	if l, err := b6.FromLiteral(c.Key()); err == nil {
		return b6.Expression{AnyExpression: l.AnyLiteral}
	} else {
		panic(err.Error())
	}
}

func (c collectionFeatureIterator) ValueExpression() b6.Expression {
	if l, err := b6.FromLiteral(c.Value()); err == nil {
		return b6.Expression{AnyExpression: l.AnyLiteral}
	} else {
		panic(err.Error())
	}
}

type segmentIterator struct {
	segments []b6.Segment
	i        int
}

func NewSegmentIterator(segments []b6.Segment) b6.Segments {
	return &segmentIterator{segments: segments}
}

func (s *segmentIterator) Next() bool {
	s.i++
	return s.i <= len(s.segments)
}

func (s *segmentIterator) Segment() b6.Segment {
	return s.segments[s.i-1]
}

func traverse(originID b6.FeatureID, f b6.FeaturesByID, r *FeatureReferencesByID) []b6.Segment {
	segments := make([]b6.Segment, 0, 2)
	origin := f.FindFeatureByID(originID)
	if origin == nil {
		return segments
	}

	references := r.FindReferences(originID, b6.FeatureTypePath)
	for _, reference := range references {
		path := f.FindFeatureByID(reference.Source()).(b6.PhysicalFeature)

		reference, ok := reference.(b6.IndexedReference)
		if !ok {
			continue
		}
		index := reference.Index()

		traversals := []struct {
			start int
			delta int
		}{
			{index + 1, 1},
			{index - 1, -1},
		}
		for _, traversal := range traversals {
			for i := traversal.start; i >= 0 && i < path.GeometryLen(); i += traversal.delta {
				if pointID := path.Reference(i).Source(); pointID.IsValid() {
					isNode := i == 0 || i == path.GeometryLen()-1 || len(r.FindReferences(pointID, b6.FeatureTypePath)) > 1
					if !isNode {
						if point := f.FindFeatureByID(pointID); point != nil {
							isNode = len(point.AllTags()) > 1
						}
					}
					if isNode {
						segments = append(segments, b6.Segment{Feature: path, First: index, Last: i})
						break
					}
				}
			}
		}
	}

	return segments
}

type areaFeatures struct {
	features []b6.AreaFeature
	i        int
}

func NewAreaFeatureIterator(features []b6.AreaFeature) b6.AreaFeatures {
	return &areaFeatures{features: features, i: -1}
}

func (a *areaFeatures) Feature() b6.AreaFeature {
	return a.features[a.i]
}

func (a *areaFeatures) FeatureID() b6.FeatureID {
	return a.features[a.i].FeatureID()
}

func (a *areaFeatures) Next() bool {
	if a.i < len(a.features)-1 {
		a.i++
		return true
	}
	return false
}

type basicWorld struct {
	features   *FeaturesByID
	references *FeatureReferencesByID
	index      *FeatureIndex
}

func (b *basicWorld) FindFeatureByID(id b6.FeatureID) b6.Feature {
	return b.features.FindFeatureByID(id)
}

func (b *basicWorld) HasFeatureWithID(id b6.FeatureID) bool {
	return b.features.HasFeatureWithID(id)
}

func (b *basicWorld) FindLocationByID(id b6.FeatureID) (s2.LatLng, error) {
	return b.features.FindLocationByID(id)
}

func (b *basicWorld) FindFeatures(q b6.Query) b6.Features {
	return b6.NewSearchFeatureIterator(q.Compile(b.index, b), b.index)
}

type relationFeatures struct {
	relations []b6.RelationFeature
	i         int
}

func NewRelationFeatureIterator(relations []b6.RelationFeature) b6.RelationFeatures {
	return &relationFeatures{relations: relations, i: -1}
}

func (r *relationFeatures) Next() bool {
	if r.i < len(r.relations)-1 {
		r.i++
		return true
	}
	return false
}

func (r *relationFeatures) Feature() b6.RelationFeature {
	return r.relations[r.i]
}

func (r *relationFeatures) FeatureID() b6.FeatureID {
	return r.relations[r.i].FeatureID()
}

func (r *relationFeatures) RelationID() b6.RelationID {
	return r.relations[r.i].RelationID()
}

func (b *basicWorld) FindRelationsByFeature(id b6.FeatureID) b6.RelationFeatures {
	references := b.FindReferences(id, b6.FeatureTypeRelation)
	var features []b6.RelationFeature
	for references.Next() {
		features = append(features, references.Feature().(b6.RelationFeature))
	}

	return NewRelationFeatureIterator(features)
}

type collectionFeatures struct {
	collections []b6.CollectionFeature
	i           int
}

func NewCollectionFeatureIterator(collections []b6.CollectionFeature) b6.CollectionFeatures {
	return &collectionFeatures{collections: collections, i: -1}
}

func (c *collectionFeatures) Next() bool {
	if c.i < len(c.collections)-1 {
		c.i++
		return true
	}

	return false
}

func (c *collectionFeatures) Feature() b6.CollectionFeature {
	return c.collections[c.i]
}

func (c *collectionFeatures) FeatureID() b6.FeatureID {
	return c.collections[c.i].FeatureID()
}

func (c *collectionFeatures) CollectionID() b6.CollectionID {
	return c.collections[c.i].CollectionID()
}

func (b *basicWorld) FindCollectionsByFeature(id b6.FeatureID) b6.CollectionFeatures {
	references := b.FindReferences(id, b6.FeatureTypeCollection)
	var features []b6.CollectionFeature
	for references.Next() {
		features = append(features, references.Feature().(b6.CollectionFeature))
	}

	return NewCollectionFeatureIterator(features)
}

func (b *basicWorld) FindAreasByPoint(p b6.FeatureID) b6.AreaFeatures {
	references := b.FindReferences(p.FeatureID(), b6.FeatureTypeArea)
	var features []b6.AreaFeature
	for references.Next() {
		features = append(features, references.Feature().(b6.AreaFeature))
	}

	return NewAreaFeatureIterator(features)
}

func (b *basicWorld) FindReferences(id b6.FeatureID, typed ...b6.FeatureType) b6.Features {
	references := b.references.FindReferences(id, typed...)

	features := make([]b6.Feature, 0, len(references))
	for _, reference := range references {
		if feature := b.FindFeatureByID(reference.Source()); feature != nil {
			if !slices.Contains(features, feature) {
				features = append(features, feature)
			}
		}
	}

	return b6.NewFeatureIterator(features)
}

func (b *basicWorld) Traverse(origin b6.FeatureID) b6.Segments {
	return NewSegmentIterator(traverse(origin, b.features, b.references))
}

func (b *basicWorld) EachFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	return EachFeature(each, b.features, b.references, options)
}

func (b *basicWorld) Tokens() []string {
	return search.AllTokens(b.index.Tokens())
}

type BuildOptions struct {
	// Return an error when paths have points ordered clockwise if true, otherwise,
	// invert them
	FailClockwisePaths bool
	// Return an error when featues are invalid, otherwise, delete them
	FailInvalidFeatures bool
	Cores               int
}

type BasicWorldBuilder struct {
	features *FeaturesByID
}

func NewBasicWorldBuilder(o *BuildOptions) *BasicWorldBuilder {
	return &BasicWorldBuilder{features: NewFeaturesByID()}
}

func (b *BasicWorldBuilder) AddFeature(f Feature) {
	b.features.AddFeature(f)
}

type BrokenFeatures []BrokenFeature

func (b BrokenFeatures) Error() string {
	messages := make([]string, len(b))
	for i, broken := range b {
		messages[i] = fmt.Sprintf("%s: %s", broken.ID, broken.Err.Error())
	}
	return strings.Join(messages, "\n")
}

func (b *BasicWorldBuilder) Finish(o *BuildOptions) (b6.World, error) {
	stages := []func(toIndex chan<- Feature, byID *FeaturesByID){
		func(c chan<- Feature, features *FeaturesByID) {
			for _, feature := range *features {
				c <- feature
			}
		},
	}

	var wg sync.WaitGroup
	var lock sync.Mutex
	broken := make(BrokenFeatures, 0)
	validate := func(c <-chan Feature) {
		defer wg.Done()
		for feature := range c {
			if err := ValidateFeature(feature, &ValidateOptions{InvertClockwisePaths: !o.FailClockwisePaths}, b.features); err != nil {
				lock.Lock()
				broken = append(broken, BrokenFeature{ID: feature.FeatureID(), Err: err})
				lock.Unlock()
			}
		}
	}

	cores := o.Cores
	if cores < 1 {
		cores = 1
	}
	for _, feed := range stages {
		toValidate := make(chan Feature, cores)
		wg.Add(cores)
		for i := 0; i < cores; i++ {
			go validate(toValidate)
		}
		feed(toValidate, b.features)
		close(toValidate)
		wg.Wait()
		if len(broken) > 0 {
			if o.FailInvalidFeatures {
				return nil, broken
			}
			for _, br := range broken {
				delete(*b.features, br.ID)
			}
		}
	}

	w := &basicWorld{
		features:   b.features,
		references: NewFilledFeatureReferences(b.features),
		index:      NewFeatureIndex(b.features),
	}

	index := func(toIndex <-chan Feature) {
		defer wg.Done()
		for feature := range toIndex {
			tokens := TokensForFeature(WrapFeature(feature, w))
			lock.Lock()
			w.index.Add(feature, tokens)
			lock.Unlock()
		}
	}

	for _, feed := range stages {
		wg.Add(cores)
		toIndex := make(chan Feature, cores)
		for i := 0; i < cores; i++ {
			go index(toIndex)
		}
		feed(toIndex, w.features)
		close(toIndex)
		wg.Wait()
	}

	w.index.Finish(cores)
	return w, nil
}

type BrokenFeature struct {
	ID  b6.FeatureID
	Err error
}

func NewWorldFromSource(source FeatureSource, o *BuildOptions) (b6.World, error) {
	b := NewBasicWorldBuilder(o)
	var lock sync.Mutex
	f := func(feature Feature, g int) error {
		feature = feature.Clone()
		lock.Lock()
		b.AddFeature(feature)
		lock.Unlock()
		return nil
	}
	options := ReadOptions{Goroutines: o.Cores}
	if err := source.Read(options, f, context.Background()); err != nil {
		return nil, err
	}

	return b.Finish(o)
}
