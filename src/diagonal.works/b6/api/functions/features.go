package functions

import (
	"fmt"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/search"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

func findFeature(context *api.Context, id b6.FeatureID) (b6.Feature, error) {
	return context.World.FindFeatureByID(id), nil
}

func findPointFeature(context *api.Context, id b6.FeatureID) (b6.PointFeature, error) {
	if id.Type == b6.FeatureTypePoint {
		return b6.FindPointByID(id.ToPointID(), context.World), nil
	}
	return nil, fmt.Errorf("%s isn't a point", id)
}

func findPathFeature(context *api.Context, id b6.FeatureID) (b6.PathFeature, error) {
	if id.Type == b6.FeatureTypePath {
		return b6.FindPathByID(id.ToPathID(), context.World), nil
	}
	return nil, fmt.Errorf("%s isn't a path", id)
}

func findAreaFeature(context *api.Context, id b6.FeatureID) (b6.AreaFeature, error) {
	if id.Type == b6.FeatureTypeArea {
		return b6.FindAreaByID(id.ToAreaID(), context.World), nil
	}
	return nil, fmt.Errorf("%s isn't a area", id)
}

func findRelationFeature(context *api.Context, id b6.FeatureID) (b6.RelationFeature, error) {
	if id.Type == b6.FeatureTypeRelation {
		return b6.FindRelationByID(id.ToRelationID(), context.World), nil
	}
	return nil, fmt.Errorf("%s isn't a relation", id)
}

func areaContainsAnyPoint(area b6.AreaFeature, points []s2.Point) (s2.Point, bool) {
	for i := 0; i < area.Len(); i++ {
		polygon := area.Polygon(i)
		for _, point := range points {
			if polygon.ContainsPoint(point) {
				return point, true
			}
		}
	}
	return s2.Point{}, false
}

func findAreasContainingPoints(context *api.Context, points b6.Collection[any, b6.Point], q b6.Query) (b6.Collection[b6.FeatureID, b6.AreaFeature], error) {
	cells := make(map[s2.CellID][]s2.Point)
	i := points.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return b6.Collection[b6.FeatureID, b6.AreaFeature]{}, err
		}
		if !ok {
			break
		}
		p := i.Value().(b6.PointFeature).Point()
		id := s2.CellFromPoint(p).ID().Parent(search.MaxIndexedCellLevel)
		points, ok := cells[id]
		if !ok {
			points = make([]s2.Point, 0, 2)
		}
		cells[id] = append(points, p)
	}

	matched := make(map[b6.AreaID]b6.AreaFeature)
	for cell, points := range cells {
		region := s2.CellUnion{cell}
		areas := b6.FindAreas(b6.Intersection{q, b6.MightIntersect{Region: &region}}, context.World)
		for areas.Next() {
			id := areas.Feature().AreaID()
			if _, ok := matched[id]; !ok {
				if _, ok := areaContainsAnyPoint(areas.Feature(), points); ok {
					matched[id] = areas.Feature()
				}
			}
		}
	}
	collection := b6.ArrayFeatureCollection[b6.AreaFeature](make([]b6.AreaFeature, 0, len(matched)))
	for _, feature := range matched {
		collection = append(collection, feature)
	}
	return collection.Collection(), nil
}

func tag(context *api.Context, key string, value string) (b6.Tag, error) {
	return b6.Tag{Key: key, Value: value}, nil
}

func value(context *api.Context, tag b6.Tag) (string, error) {
	return tag.Value, nil
}

func intValue(context *api.Context, tag b6.Tag) (int, error) {
	i, _ := tag.IntValue()
	return i, nil
}

func floatValue(context *api.Context, tag b6.Tag) (float64, error) {
	f, _ := tag.FloatValue()
	return f, nil
}

func get(context *api.Context, id b6.Identifiable, key string) (b6.Tag, error) {
	if feature := api.Resolve(id, context.World); feature != nil {
		return feature.Get(key), nil
	}
	return b6.InvalidTag(), nil
}

func getString(context *api.Context, id b6.Identifiable, key string) (string, error) {
	if feature := api.Resolve(id, context.World); feature != nil {
		return feature.Get(key).Value, nil
	}
	return "", nil
}

func getInt(context *api.Context, id b6.Identifiable, key string) (int, error) {
	if feature := api.Resolve(id, context.World); feature != nil {
		if i, ok := feature.Get(key).IntValue(); ok {
			return i, nil
		}
	}
	return 0, nil
}

func getFloat(context *api.Context, id b6.Identifiable, key string) (float64, error) {
	if feature := api.Resolve(id, context.World); feature != nil {
		if f, ok := feature.Get(key).FloatValue(); ok {
			return f, nil
		}
	}
	return 0.0, nil
}

func countTagValue(context *api.Context, id b6.Identifiable, key string) (b6.Collection[interface{}, int], error) {
	c := &b6.ArrayCollection[interface{}, int]{
		Keys:   make([]interface{}, 0, 1),
		Values: make([]int, 0, 1),
	}
	if feature := api.Resolve(id, context.World); feature != nil {
		if tag := feature.Get(key); tag.IsValid() {
			c.Keys = append(c.Keys, api.StringStringPair{key, tag.Value})
			c.Values = append(c.Values, 1)
		}
	}
	return c.Collection(), nil
}

func allTags(c *api.Context, id b6.Identifiable) (b6.Collection[int, b6.Tag], error) {
	var tags []b6.Tag
	if f := api.Resolve(id, c.World); f != nil {
		tags = f.AllTags()
	}
	return b6.ArrayValuesCollection[b6.Tag](tags).Collection(), nil
}

func matches(c *api.Context, id b6.Identifiable, query b6.Query) (bool, error) {
	if f := api.Resolve(id, c.World); f != nil {
		return query.Matches(f, c.World), nil
	}
	return false, nil
}

func pointDegree(context *api.Context, point b6.PointFeature) (int, error) {
	segments := context.World.Traverse(point.PointID())
	n := 0
	for segments.Next() {
		n++
	}
	return n, nil
}

func pathLengthMeters(context *api.Context, path b6.PathFeature) (float64, error) {
	return b6.AngleToMeters(path.Polyline().Length()), nil
}

type pathPointCollection struct {
	path b6.Path
	i    int
}

func (p *pathPointCollection) Begin() b6.Iterator[int, b6.Point] {
	return &pathPointCollection{path: p.path}
}

func (p *pathPointCollection) Count() (int, bool) {
	return p.path.Len(), true
}

func (p *pathPointCollection) Next() (bool, error) {
	if p.i >= p.path.Len() {
		return false, nil
	}
	p.i++
	return true, nil
}

func (p *pathPointCollection) Key() int {
	return p.i - 1
}

func (p *pathPointCollection) Value() b6.Point {
	return b6.PointFromS2Point(p.path.Point(p.i - 1))
}

var _ b6.AnyCollection[int, b6.Point] = &pathPointCollection{}

type areaPointCollection struct {
	area    b6.Area
	polygon *s2.Polygon
	loop    *s2.Loop
	i       int
	j       int
	k       int
	n       int
}

func (a *areaPointCollection) Begin() b6.Iterator[int, b6.Point] {
	return &areaPointCollection{area: a.area}
}

func (a *areaPointCollection) Count() (int, bool) {
	n := 0
	for i := 0; i < a.area.Len(); i++ {
		// TODO: Add a more efficient interface to Area() that takes
		// the indices directly?
		polygon := a.area.Polygon(i)
		for j := 0; j < polygon.NumLoops(); j++ {
			n += polygon.Loop(j).NumVertices()
		}
	}
	return n, true
}

func (a *areaPointCollection) Next() (bool, error) {
	for {
		if a.polygon == nil {
			if a.i >= a.area.Len() {
				return false, nil
			}
			a.polygon = a.area.Polygon(a.i)
			a.loop = nil
			a.i++
			a.j = 0
		}
		if a.loop == nil {
			if a.j >= a.polygon.NumLoops() {
				a.polygon = nil
				a.loop = nil
				continue
			}
			a.loop = a.polygon.Loop(a.j)
			a.j++
			a.k = 0
		}
		if a.k >= a.loop.NumVertices() {
			a.loop = nil
		} else {
			a.k++
			a.n++
			return true, nil
		}
	}
}

func (a *areaPointCollection) Key() int {
	return a.n - 1
}

func (a *areaPointCollection) Value() b6.Point {
	return b6.PointFromS2Point(a.loop.Vertex(a.k - 1))
}

var _ b6.AnyCollection[int, b6.Point] = &areaPointCollection{}

func points(context *api.Context, g b6.Geometry) (b6.Collection[int, b6.Point], error) {
	switch g := g.(type) {
	case b6.Point:
		return b6.ArrayValuesCollection[b6.Point]([]b6.Point{g}).Collection(), nil
	case b6.Path:
		return b6.Collection[int, b6.Point]{
			AnyCollection: &pathPointCollection{path: g},
		}, nil
	case b6.Area:
		return b6.Collection[int, b6.Point]{
			AnyCollection: &areaPointCollection{area: g},
		}, nil
	}
	return b6.ArrayValuesCollection[b6.Point]([]b6.Point{}).Collection(), nil
}

func pointFeatures(context *api.Context, f b6.Feature) (b6.Collection[b6.FeatureID, b6.PointFeature], error) {
	points := b6.ArrayFeatureCollection[b6.PointFeature](make([]b6.PointFeature, 0))
	switch f := f.(type) {
	case b6.PointFeature:
		points = append(points, f)
	case b6.PathFeature:
		for i := 0; i < f.Len(); i++ {
			if p := f.Feature(i); p != nil {
				points = append(points, p)
			}
		}
	case b6.AreaFeature:
		for i := 0; i < f.Len(); i++ {
			for _, path := range f.Feature(i) {
				for j := 0; j < path.Len(); j++ {
					if p := path.Feature(j); p != nil {
						points = append(points, p)
					}
				}
			}
		}
	}
	return points.Collection(), nil
}

func pointPaths(context *api.Context, id b6.IdentifiablePoint) (b6.Collection[b6.FeatureID, b6.PathFeature], error) {
	p := api.ResolvePoint(id, context.World)
	if p == nil {
		return b6.Collection[b6.FeatureID, b6.PathFeature]{}, fmt.Errorf("No point with id %s", id)
	}
	collection := b6.ArrayFeatureCollection[b6.PathFeature](make([]b6.PathFeature, 0))
	paths := context.World.FindPathsByPoint(p.PointID())
	for paths.Next() {
		collection = append(collection, paths.Feature())
	}
	return collection.Collection(), nil
}

func samplePointsAlongPaths(context *api.Context, paths b6.Collection[b6.FeatureID, b6.Path], distanceMeters float64) (b6.Collection[int, b6.Point], error) {
	// TODO: We shouldn't need to special case this: we should be able to flatten the results of sample_points
	// on a collection of paths.
	seen := make(map[s2.Point]struct{})
	points := make([]b6.Point, 0, 16)
	i := paths.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return b6.Collection[int, b6.Point]{}, err
		}
		if !ok {
			break
		}
		points = appendUnseenSampledPoints(i.Value(), distanceMeters, seen, points)
	}
	return b6.ArrayValuesCollection[b6.Point](points).Collection(), nil
}

func samplePoints(context *api.Context, path b6.Path, distanceMeters float64) (b6.Collection[int, b6.Point], error) {
	points := appendUnseenSampledPoints(path, distanceMeters, make(map[s2.Point]struct{}), make([]b6.Point, 0, 16))
	return b6.ArrayValuesCollection[b6.Point](points).Collection(), nil
}

func appendUnseenSampledPoints(p b6.Path, distanceMeters float64, seen map[s2.Point]struct{}, points []b6.Point) []b6.Point {
	const epsilon s1.Angle = 1.6e-09 // Roughly 1cm
	polyline := p.Polyline()
	var step float64
	if polyline.Length() > epsilon {
		step = float64(b6.MetersToAngle(distanceMeters) / polyline.Length())
	} else {
		step = 1.0
	}
	j := 0.0
	done := false
	for !done {
		if j >= 1.0 {
			j = 1.0
			done = true
		}
		p, _ := polyline.Interpolate(j)
		if _, ok := seen[p]; !ok {
			points = append(points, b6.PointFromS2Point(p))
			seen[p] = struct{}{}
		}
		j += step
	}
	return points
}

func join(context *api.Context, a b6.Path, b b6.Path) (b6.Path, error) {
	points := make([]s2.Point, 0, a.Len()+b.Len())
	i := 0
	for i < a.Len() {
		points = append(points, a.Point(i))
		i++
	}
	i = 0
	if b.Point(0) == a.Point(a.Len()-1) {
		i++
	}
	for i < b.Len() {
		points = append(points, b.Point(i))
		i++
	}
	return b6.PathFromS2Points(points), nil
}

// orderedJoinPaths returns a new path formed by joining a and b, in that order, reversing
// the order of the points to maintain a consistent order, determined by which points of
// the paths are shared. Returns an error if the paths don't share an end point.
func orderedJoin(context *api.Context, a b6.Path, b b6.Path) (b6.Path, error) {
	var reverseA, reverseB bool
	if a.Point(a.Len()-1) == b.Point(0) {
		reverseA, reverseB = false, false
	} else if a.Point(a.Len()-1) == b.Point(b.Len()-1) {
		reverseA, reverseB = false, true
	} else if a.Point(0) == b.Point(0) {
		reverseA, reverseB = true, false
	} else if a.Point(0) == b.Point(b.Len()-1) {
		reverseA, reverseB = true, true
	} else {
		return nil, fmt.Errorf("Paths don't share an end vertex")
	}
	points := make([]s2.Point, 0, a.Len()+b.Len())
	if reverseA {
		for i := a.Len() - 1; i >= 0; i-- {
			points = append(points, a.Point(i))
		}
	} else {
		for i := 0; i < a.Len(); i++ {
			points = append(points, a.Point(i))
		}
	}
	if reverseB {
		for i := b.Len() - 2; i >= 0; i-- {
			points = append(points, b.Point(i))
		}
	} else {
		for i := 1; i < b.Len(); i++ {
			points = append(points, b.Point(i))
		}
	}
	return b6.PathFromS2Points(points), nil
}
