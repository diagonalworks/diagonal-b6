package functions

import (
	"fmt"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/search"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

func findFeature(id b6.FeatureID, context *api.Context) (b6.Feature, error) {
	return context.World.FindFeatureByID(id), nil
}

func findPointFeature(id b6.FeatureID, context *api.Context) (b6.PointFeature, error) {
	if id.Type == b6.FeatureTypePoint {
		return b6.FindPointByID(id.ToPointID(), context.World), nil
	}
	return nil, fmt.Errorf("%s isn't a point", id)
}

func findPathFeature(id b6.FeatureID, context *api.Context) (b6.PathFeature, error) {
	if id.Type == b6.FeatureTypePath {
		return b6.FindPathByID(id.ToPathID(), context.World), nil
	}
	return nil, fmt.Errorf("%s isn't a path", id)
}

func findAreaFeature(id b6.FeatureID, context *api.Context) (b6.AreaFeature, error) {
	if id.Type == b6.FeatureTypeArea {
		return b6.FindAreaByID(id.ToAreaID(), context.World), nil
	}
	return nil, fmt.Errorf("%s isn't a area", id)
}

func findRelationFeature(id b6.FeatureID, context *api.Context) (b6.RelationFeature, error) {
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

type arrayAreaFeatureCollection struct {
	features []b6.AreaFeature
	i        int
}

func (a *arrayAreaFeatureCollection) Count() int { return len(a.features) }

func (a *arrayAreaFeatureCollection) Begin() api.CollectionIterator {
	return &arrayAreaFeatureCollection{
		features: a.features,
	}
}

func (a *arrayAreaFeatureCollection) Key() interface{} {
	return a.FeatureIDKey()
}

func (a *arrayAreaFeatureCollection) Value() interface{} {
	return a.AreaFeatureValue()
}

func (a *arrayAreaFeatureCollection) FeatureIDKey() b6.FeatureID {
	return a.features[a.i-1].FeatureID()
}

func (a *arrayAreaFeatureCollection) AreaFeatureValue() b6.AreaFeature {
	return a.features[a.i-1]
}

func (a *arrayAreaFeatureCollection) Next() (bool, error) {
	a.i++
	return a.i <= len(a.features), nil
}

var _ api.Collection = &arrayAreaFeatureCollection{}
var _ api.Countable = &arrayAreaFeatureCollection{}

func findAreasContainingPoints(points api.PointFeatureCollection, q b6.Query, context *api.Context) (api.AreaFeatureCollection, error) {
	cells := make(map[s2.CellID][]s2.Point)
	i := points.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
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
		areas := b6.FindAreas(b6.Intersection{q, b6.MightIntersect{&region}}, context.World)
		for areas.Next() {
			id := areas.Feature().AreaID()
			if _, ok := matched[id]; !ok {
				if _, ok := areaContainsAnyPoint(areas.Feature(), points); ok {
					matched[id] = areas.Feature()
				}
			}
		}
	}
	collection := &arrayAreaFeatureCollection{
		features: make([]b6.AreaFeature, 0, len(matched)),
	}
	for _, feature := range matched {
		collection.features = append(collection.features, feature)
	}
	return collection, nil
}

func tag(key string, value string, context *api.Context) (b6.Tag, error) {
	return b6.Tag{Key: key, Value: value}, nil
}

func value(tag b6.Tag, context *api.Context) (string, error) {
	return tag.Value, nil
}

func intValue(tag b6.Tag, context *api.Context) (int, error) {
	i, _ := tag.IntValue()
	return i, nil
}

func floatValue(tag b6.Tag, context *api.Context) (float64, error) {
	f, _ := tag.FloatValue()
	return f, nil
}

func get(id b6.Identifiable, key string, context *api.Context) (b6.Tag, error) {
	if feature := api.Resolve(id, context.World); feature != nil {
		return feature.Get(key), nil
	}
	return b6.InvalidTag(), nil
}

func getString(id b6.Identifiable, key string, context *api.Context) (string, error) {
	if feature := api.Resolve(id, context.World); feature != nil {
		return feature.Get(key).Value, nil
	}
	return "", nil
}

func getInt(id b6.Identifiable, key string, context *api.Context) (int, error) {
	if feature := api.Resolve(id, context.World); feature != nil {
		if i, ok := feature.Get(key).IntValue(); ok {
			return i, nil
		}
	}
	return 0, nil
}

func getFloat(id b6.Identifiable, key string, context *api.Context) (float64, error) {
	if feature := api.Resolve(id, context.World); feature != nil {
		if f, ok := feature.Get(key).FloatValue(); ok {
			return f, nil
		}
	}
	return 0.0, nil
}

func hasKey(id b6.Feature, key string, context *api.Context) (bool, error) {
	if feature := api.Resolve(id, context.World); feature != nil {
		return feature.Get(key).IsValid(), nil
	}
	return false, nil
}

func countTagValue(id b6.Identifiable, key string, context *api.Context) (api.Collection, error) {
	c := &api.ArrayAnyIntCollection{
		Keys:   make([]interface{}, 0, 1),
		Values: make([]int, 0, 1),
	}
	if feature := api.Resolve(id, context.World); feature != nil {
		if tag := feature.Get(key); tag.IsValid() {
			c.Keys = append(c.Keys, api.StringStringPair{key, tag.Value})
			c.Values = append(c.Values, 1)
		}
	}
	return c, nil
}

func allTags(id b6.Identifiable, c *api.Context) (api.IntTagCollection, error) {
	var tags []b6.Tag
	if f := api.Resolve(id, c.World); f != nil {
		tags = f.AllTags()
	}
	return &api.ArrayTagCollection{Tags: tags}, nil
}

func pointDegree(point b6.PointFeature, context *api.Context) (int, error) {
	paths := context.World.FindPathsByPoint(point.PointID())
	n := 0
	for paths.Next() {
		n++
	}
	return n, nil
}

func pathLengthMeters(path b6.PathFeature, context *api.Context) (float64, error) {
	return b6.AngleToMeters(path.Polyline().Length()), nil
}

type pathPointCollection struct {
	path b6.Path
	i    int
}

func (p *pathPointCollection) Begin() api.CollectionIterator {
	return &pathPointCollection{path: p.path}
}

func (p *pathPointCollection) Count() int {
	return p.path.Len()
}

func (p *pathPointCollection) Next() (bool, error) {
	if p.i >= p.path.Len() {
		return false, nil
	}
	p.i++
	return true, nil
}

func (p *pathPointCollection) Key() interface{} {
	return p.i - 1
}

func (p *pathPointCollection) Value() interface{} {
	return b6.PointFromS2Point(p.path.Point(p.i - 1))
}

var _ api.Collection = &pathPointCollection{}
var _ api.Countable = &pathPointCollection{}

type areaPointCollection struct {
	area    b6.Area
	polygon *s2.Polygon
	loop    *s2.Loop
	i       int
	j       int
	k       int
	n       int
}

func (a *areaPointCollection) Begin() api.CollectionIterator {
	return &areaPointCollection{area: a.area}
}

func (a *areaPointCollection) Count() int {
	count := 0
	for i := 0; i < a.area.Len(); i++ {
		// TODO: Add a more efficient interface to Area() that takes
		// the indices directly?
		polygon := a.area.Polygon(i)
		for j := 0; j < polygon.NumLoops(); j++ {
			count += polygon.Loop(j).NumVertices()
		}
	}
	return count
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

func (a *areaPointCollection) Key() interface{} {
	return a.n - 1
}

func (a *areaPointCollection) Value() interface{} {
	return b6.PointFromS2Point(a.loop.Vertex(a.k - 1))
}

var _ api.Collection = &areaPointCollection{}

func points(g b6.Geometry, context *api.Context) (api.PointCollection, error) {
	switch g := g.(type) {
	case b6.Point:
		return &singletonCollection{k: 0, v: g}, nil
	case b6.Path:
		return &pathPointCollection{path: g}, nil
	case b6.Area:
		return &areaPointCollection{area: g}, nil
	}
	return &api.ArrayPointCollection{}, nil
}

func pointFeatures(f b6.Feature, context *api.Context) (api.PointFeatureCollection, error) {
	points := &api.ArrayPointFeatureCollection{Features: make([]b6.PointFeature, 0)}
	switch f := f.(type) {
	case b6.PointFeature:
		points.Features = append(points.Features, f)
	case b6.PathFeature:
		for i := 0; i < f.Len(); i++ {
			if p := f.Feature(i); p != nil {
				points.Features = append(points.Features, p)
			}
		}
	case b6.AreaFeature:
		for i := 0; i < f.Len(); i++ {
			for _, path := range f.Feature(i) {
				for j := 0; j < path.Len(); j++ {
					if p := path.Feature(j); p != nil {
						points.Features = append(points.Features, p)
					}
				}
			}
		}
	}
	return points, nil
}

func pointPaths(id b6.IdentifiablePoint, context *api.Context) (api.PathFeatureCollection, error) {
	p := api.ResolvePoint(id, context.World)
	if p == nil {
		return nil, fmt.Errorf("No point with id %s", id)
	}
	paths := &api.ArrayPathFeatureCollection{Features: make([]b6.PathFeature, 0)}
	segments := context.World.FindPathsByPoint(p.PointID())
	seen := make(map[b6.FeatureID]struct{})
	for segments.Next() {
		id := segments.PathSegment().PathFeature.FeatureID()
		if _, ok := seen[id]; !ok {
			paths.Features = append(paths.Features, segments.PathSegment().PathFeature)
			seen[id] = struct{}{}
		}
	}
	return paths, nil
}

func samplePointsAlongPaths(paths api.PathCollection, distanceMeters float64, context *api.Context) (api.PointCollection, error) {
	// TODO: We shouldn't need to special case this: we should be able to flattern the results of sample_points
	// on a collection of paths.
	seen := make(map[s2.Point]struct{})
	points := make([]s2.Point, 0, 16)
	i := paths.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		points = appendUnseenSampledPoints(i.Value().(b6.Path), distanceMeters, seen, points)
	}
	return pointsToCollection(points), nil
}

func samplePoints(path b6.Path, distanceMeters float64, context *api.Context) (api.StringPointCollection, error) {
	points := appendUnseenSampledPoints(path, distanceMeters, make(map[s2.Point]struct{}), make([]s2.Point, 0, 16))
	return pointsToCollection(points), nil
}

func appendUnseenSampledPoints(p b6.Path, distanceMeters float64, seen map[s2.Point]struct{}, points []s2.Point) []s2.Point {
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
			points = append(points, p)
			seen[p] = struct{}{}
		}
		j += step
	}
	return points
}

// pointsToCollection wraps the slice of points into a PointCollection, reusing,
// rather than copying, the slice
func pointsToCollection(points []s2.Point) api.StringPointCollection {
	keys := make([]string, len(points))
	for i, p := range points {
		ll := s2.LatLngFromPoint(p)
		keys[i] = fmt.Sprintf("%f,%f", ll.Lat.Degrees(), ll.Lng.Degrees())
	}
	return &api.ArrayPointCollection{Keys: keys, Values: points}
}

func join(a b6.Path, b b6.Path, context *api.Context) (b6.Path, error) {
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
func orderedJoin(a b6.Path, b b6.Path, context *api.Context) (b6.Path, error) {
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
