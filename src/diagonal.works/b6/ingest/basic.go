package ingest

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/geojson"
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
	byID b6.FeaturesByID
}

func NewFeatureIndex(byID b6.FeaturesByID) *FeatureIndex {
	return &FeatureIndex{ArrayIndex: *search.NewArrayIndex(featureValues{}), byID: byID}
}

func (f *FeatureIndex) RewriteSpatialQuery(query search.Spatial) search.Query {
	return RewriteSpatialQuery(query)
}

func (f *FeatureIndex) RewriteFeatureTypeQuery(q b6.FeatureTypeQuery) search.Query {
	return RewriteFeatureTypeQuery(q)
}

func (f *FeatureIndex) Feature(v search.Value) b6.Feature {
	switch feature := v.(type) {
	case *PointFeature:
		return newPointFeature(feature)
	case *PathFeature:
		return newPathFeature(feature, f.byID)
	case *AreaFeature:
		return newAreaFeature(feature, f.byID)
	case *RelationFeature:
		return newRelationFeature(feature, f.byID)
	}
	panic(fmt.Sprintf("Bad feature type: %T", v))
}

func (f *FeatureIndex) ID(v search.Value) b6.FeatureID {
	if i, ok := v.(b6.Identifiable); ok {
		return i.FeatureID()
	}
	panic(fmt.Sprintf("Bad feature type: %T", v))
}

type pointFeature struct {
	*PointFeature
}

func newPointFeature(p *PointFeature) pointFeature {
	return pointFeature{p}
}

func (p pointFeature) PointID() b6.PointID {
	return p.PointFeature.PointID
}

func (p pointFeature) Covering(coverer s2.RegionCoverer) s2.CellUnion {
	return s2.CellUnion([]s2.CellID{p.CellID().Parent(coverer.MaxLevel)})
}

func (p pointFeature) ToGeoJSON() geojson.GeoJSON {
	return b6.PointFeatureToGeoJSON(p)
}

var _ b6.PointFeature = (*pointFeature)(nil)

func WrapPointFeature(p *PointFeature) b6.PointFeature {
	return newPointFeature(p)
}

type pathFeature struct {
	*PathFeature
	features b6.FeaturesByID
}

func newPathFeature(p *PathFeature, features b6.FeaturesByID) pathFeature {
	return pathFeature{p, features}
}

func (p pathFeature) PathID() b6.PathID {
	return p.PathFeature.PathID
}

func (p pathFeature) Point(i int) s2.Point {
	if point, ok := p.PathFeature.Point(i); ok {
		return point
	}
	id, _ := p.PointID(i)
	if ll, ok := p.features.FindLocationByID(id); ok {
		return s2.PointFromLatLng(ll)
	}
	panic(fmt.Sprintf("No point with ID %s", id))
}

func (p pathFeature) Polyline() *s2.Polyline {
	points := make([]s2.Point, 0, p.Len())
	for i := 0; i < p.Len(); i++ {
		points = append(points, p.Point(i))
	}
	return (*s2.Polyline)(&points)
}

func (p pathFeature) Covering(coverer s2.RegionCoverer) s2.CellUnion {
	return coverer.Covering(p.Polyline())
}

func (p pathFeature) Feature(i int) b6.PointFeature {
	if id, ok := p.PointID(i); ok {
		if point := b6.FindPointByID(id, p.features); point != nil {
			return point
		}
		panic(fmt.Sprintf("No point with ID %s", id))
	}
	return nil
}

func (p pathFeature) ToGeoJSON() geojson.GeoJSON {
	return b6.PathFeatureToGeoJSON(p)
}

var _ b6.PathFeature = (*pathFeature)(nil)

func WrapPathFeature(p *PathFeature, features b6.FeaturesByID) b6.PathFeature {
	return newPathFeature(p, features)
}

type areaFeature struct {
	*AreaFeature
	features b6.FeaturesByID
}

func newAreaFeature(a *AreaFeature, features b6.FeaturesByID) areaFeature {
	return areaFeature{a, features}
}

func WrapAreaFeature(a *AreaFeature, byID b6.FeaturesByID) b6.AreaFeature {
	return newAreaFeature(a, byID)
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
		if path := b6.FindPathByID(id, a.features); path != nil {
			loop := make([]s2.Point, path.Len()-1) // Paths are explicitly closed, so drop the duplicate last point
			for j := range loop {
				loop[j] = path.Point(j)
			}
			loops = append(loops, s2.LoopFromPoints(loop))
		}
	}
	return s2.PolygonFromLoops(loops)
}

func (a areaFeature) Covering(coverer s2.RegionCoverer) s2.CellUnion {
	covering := s2.CellUnion([]s2.CellID{})
	for i := 0; i < a.Len(); i++ {
		if polygon := a.Polygon(i); polygon != nil {
			covering = s2.CellUnionFromUnion(covering, coverer.Covering(polygon))
		}
	}
	return covering
}

func (a areaFeature) Feature(i int) []b6.PathFeature {
	if ids, ok := a.PathIDs(i); ok {
		paths := make([]b6.PathFeature, 0, len(ids))
		for _, id := range ids {
			if path := b6.FindPathByID(id, a.features); path != nil {
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

func newRelationFeature(r *RelationFeature, byID b6.FeaturesByID) relationFeature {
	return relationFeature{r, byID}
}

func WrapRelationFeature(r *RelationFeature, byID b6.FeaturesByID) b6.RelationFeature {
	return newRelationFeature(r, byID)
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

func newFeature(feature Feature, byID b6.FeaturesByID) b6.PhysicalFeature {
	switch f := feature.(type) {
	case *PointFeature:
		return newPointFeature(f)
	case *PathFeature:
		return newPathFeature(f, byID)
	case *AreaFeature:
		return newAreaFeature(f, byID)
	case *RelationFeature:
		return newRelationFeature(f, byID)
	}
	panic(fmt.Sprintf("Unknown feature: %T", feature))
}

func (r relationFeature) Covering(coverer s2.RegionCoverer) s2.CellUnion {
	return s2.CellUnion([]s2.CellID{})
}

type pathSegments struct {
	pathSegments []b6.PathSegment
	i            int
}

func NewPathSegmentIterator(s []b6.PathSegment) b6.PathSegments {
	return &pathSegments{pathSegments: s, i: -1}
}

func (p *pathSegments) Next() bool {
	p.i++
	return p.i < len(p.pathSegments)
}

func (p *pathSegments) PathSegment() b6.PathSegment {
	return p.pathSegments[p.i]
}

func findPathsByPoint(byID b6.FeaturesByID, r *FeatureReferences, origin b6.PointID, w b6.World) b6.PathSegments {
	pathPoints, ok := r.PathsByPoint[origin]
	if !ok {
		return b6.EmptyPathSegments{}
	}
	segments := make([]b6.PathSegment, 0)
	for _, p := range pathPoints {
		traversals := []struct {
			start int
			delta int
		}{
			{p.Position + 1, 1},
			{p.Position - 1, -1},
		}
		for _, traversal := range traversals {
			for i := traversal.start; i >= 0 && i < p.Path.Len(); i += traversal.delta {
				if id, ok := p.Path.PointID(i); ok {
					isNode := i == 0 || i == p.Path.Len()-1 || len(r.PathsByPoint[id]) > 1
					if !isNode {
						if point := byID.FindFeatureByID(id.FeatureID()); point != nil {
							isNode = len(point.AllTags()) > 0
						}
					}
					if isNode {
						segments = append(segments, b6.PathSegment{
							PathFeature: pathFeature{p.Path, w},
							First:       p.Position,
							Last:        i,
						})
						break
					}
				}
			}
		}
	}
	return NewPathSegmentIterator(segments)
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

func findAreasByPoint(r *FeatureReferences, p b6.PointID, w b6.World) b6.AreaFeatures {
	areas := r.AreasForPoint(p, w)
	features := make([]b6.AreaFeature, len(areas))
	for i, area := range areas {
		features[i] = newAreaFeature(area, w)
	}
	return NewAreaFeatureIterator(features)
}

func findRelationsByFeature(r *FeatureReferences, id b6.FeatureID, w b6.World) b6.RelationFeatures {
	if relations, ok := r.RelationsByFeature[id]; ok {
		wrapped := make([]b6.RelationFeature, len(relations))
		for i, r := range relations {
			wrapped[i] = relationFeature{r, w}
		}
		return NewRelationFeatureIterator(wrapped)
	}
	return NewRelationFeatureIterator([]b6.RelationFeature{})
}

type basicWorld struct {
	byID       *FeaturesByID
	references *FeatureReferences
	index      *FeatureIndex
}

func (b *basicWorld) FindFeatureByID(id b6.FeatureID) b6.Feature {
	return b.byID.FindFeatureByID(id)
}

func (b *basicWorld) HasFeatureWithID(id b6.FeatureID) bool {
	return b.byID.HasFeatureWithID(id)
}

func (b *basicWorld) FindLocationByID(id b6.PointID) (s2.LatLng, bool) {
	return b.byID.FindLocationByID(id)
}

func (b *basicWorld) FindFeatures(q search.Query) b6.Features {
	return b6.NewFeatureIterator(q.Compile(b.index), b.index)
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
	return findRelationsByFeature(b.references, id, b)
}

func (b *basicWorld) FindPathsByPoint(origin b6.PointID) b6.PathSegments {
	return findPathsByPoint(b.byID, b.references, origin, b)
}

func (b *basicWorld) FindAreasByPoint(p b6.PointID) b6.AreaFeatures {
	return findAreasByPoint(b.references, p, b)
}

func (b *basicWorld) EachFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	return b.byID.EachFeature(each, options)
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
	byID *FeaturesByID
}

func NewBasicWorldBuilder() *BasicWorldBuilder {
	b := &BasicWorldBuilder{byID: NewFeaturesByID()}
	return b
}

func (b *BasicWorldBuilder) AddSimplePoint(id b6.PointID, ll s2.LatLng) {
	b.byID.AddSimplePoint(id, ll)
}

func (b *BasicWorldBuilder) AddPoint(p *PointFeature) {
	b.byID.Points[p.PointID] = p
}

func (b *BasicWorldBuilder) AddPath(p *PathFeature) {
	b.byID.Paths[p.PathID] = p
}

func (b *BasicWorldBuilder) AddArea(a *AreaFeature) {
	b.byID.Areas[a.AreaID] = a
}

func (b *BasicWorldBuilder) AddRelation(r *RelationFeature) {
	b.byID.Relations[r.RelationID] = r
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
		func(c chan<- Feature, byID *FeaturesByID) {
			for _, point := range byID.Points {
				c <- point
			}
		},
		func(c chan<- Feature, byID *FeaturesByID) {
			for _, path := range byID.Paths {
				c <- path
			}
		},
		func(c chan<- Feature, byID *FeaturesByID) {
			for _, area := range byID.Areas {
				c <- area
			}
		},
		func(c chan<- Feature, byID *FeaturesByID) {
			for _, relation := range byID.Relations {
				c <- relation
			}
		},
	}

	var wg sync.WaitGroup
	var lock sync.Mutex
	broken := make(BrokenFeatures, 0)
	validate := func(c <-chan Feature) {
		defer wg.Done()
		for feature := range c {
			var err error
			switch f := feature.(type) {
			case *PointFeature:
				err = ValidatePoint(f)
			case *PathFeature:
				vo := ValidateOptions{InvertClockwisePaths: !o.FailClockwisePaths}
				err = ValidatePath(f, &vo, b.byID)
			case *AreaFeature:
				err = ValidateArea(f, b.byID)
			case *RelationFeature:
				err = ValidateRelation(f)
			}
			if err != nil {
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
		feed(toValidate, b.byID)
		close(toValidate)
		wg.Wait()
		if len(broken) > 0 {
			if o.FailInvalidFeatures {
				return nil, broken
			}
			for _, br := range broken {
				switch br.ID.Type {
				case b6.FeatureTypePoint:
					delete(b.byID.Points, br.ID.ToPointID())
				case b6.FeatureTypePath:
					delete(b.byID.Paths, br.ID.ToPathID())
				case b6.FeatureTypeArea:
					delete(b.byID.Areas, br.ID.ToAreaID())
				case b6.FeatureTypeRelation:
					delete(b.byID.Relations, br.ID.ToRelationID())
				}
			}
		}
	}

	w := &basicWorld{
		byID:       b.byID,
		references: NewFilledFeatureReferences(b.byID),
		index:      NewFeatureIndex(b.byID),
	}

	index := func(toIndex <-chan Feature) {
		defer wg.Done()
		for feature := range toIndex {
			tokens := TokensForFeature(newFeature(feature, w))
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
		feed(toIndex, w.byID)
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
	b := NewBasicWorldBuilder()
	var lock sync.Mutex
	f := func(feature Feature, g int) error {
		feature = feature.Clone()
		lock.Lock()
		switch f := feature.(type) {
		case *PointFeature:
			b.AddPoint(f)
		case *PathFeature:
			b.AddPath(f)
		case *AreaFeature:
			b.AddArea(f)
		case *RelationFeature:
			b.AddRelation(f)
		}
		lock.Unlock()
		return nil
	}
	options := ReadOptions{Cores: o.Cores}
	source.Read(options, f, context.Background())
	return b.Finish(o)
}
