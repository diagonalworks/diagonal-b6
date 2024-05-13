package functions

import (
	"fmt"
	"strconv"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/search"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

// Return the feature with the given ID.
func findFeature(context *api.Context, id b6.FeatureID) (b6.Feature, error) {
	if f := context.World.FindFeatureByID(id); f != nil {
		return f, nil
	} else {
		return nil, fmt.Errorf("feature with %s id not found", id.String())
	}
}

// Return the area feature with the given ID.
func findAreaFeature(context *api.Context, id b6.FeatureID) (b6.AreaFeature, error) {
	if id.Type == b6.FeatureTypeArea {
		return b6.FindAreaByID(id.ToAreaID(), context.World), nil
	}
	return nil, fmt.Errorf("%s isn't a area", id)
}

// Return the relation feature with the given ID.
func findRelationFeature(context *api.Context, id b6.FeatureID) (b6.RelationFeature, error) {
	if id.Type == b6.FeatureTypeRelation {
		return b6.FindRelationByID(id.ToRelationID(), context.World), nil
	}
	return nil, fmt.Errorf("%s isn't a relation", id)
}

// Return the collection feature with the given ID.
func findCollectionFeature(context *api.Context, id b6.FeatureID) (b6.CollectionFeature, error) {
	if id.Type == b6.FeatureTypeCollection {
		return b6.FindCollectionByID(id.ToCollectionID(), context.World), nil
	}
	return nil, fmt.Errorf("%s isn't a collection", id)
}

// Return the expression feature with the given ID.
func findExpressionFeature(context *api.Context, id b6.FeatureID) (b6.ExpressionFeature, error) {
	if id.Type == b6.FeatureTypeCollection {
		return b6.FindExpressionByID(id.ToExpressionID(), context.World), nil
	}
	return nil, fmt.Errorf("%s isn't an expression", id)
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

func findAreasContainingPoints(context *api.Context, points b6.Collection[any, b6.Feature], q b6.Query) (b6.Collection[b6.FeatureID, b6.AreaFeature], error) {
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

		v, ok := i.Value().(b6.Geometry)
		if !ok {
			return b6.Collection[b6.FeatureID, b6.AreaFeature]{}, fmt.Errorf("features don't implement geometry")
		}

		p := v.Point()
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

// Return a tag with the given key and value.
func tag(context *api.Context, key string, value string) (b6.Tag, error) {
	return b6.Tag{Key: key, Value: b6.String(value)}, nil
}

// Return the value of the given tag as a string.
func value(context *api.Context, tag b6.Tag) (string, error) {
	return tag.Value.String(), nil
}

// Return the value of the given tag as an integer.
// Propagates error if the value isn't a valid integer.
func intValue(context *api.Context, tag b6.Tag) (int, error) {
	return strconv.Atoi(tag.Value.String())
}

// Return the value of the given tag as a float.
// Propagates error if the value isn't a valid float.
func floatValue(context *api.Context, tag b6.Tag) (float64, error) {
	return strconv.ParseFloat(tag.Value.String(), 64)
}

// Return the tag with the given key on the given feature.
// Returns a tag. To return the string value of a tag, use get-string.
func get(context *api.Context, id b6.Identifiable, key string) (b6.Tag, error) {
	if feature := api.Resolve(id, context.World); feature != nil {
		return feature.Get(key), nil
	}
	return b6.InvalidTag(), nil
}

// Return the value of tag with the given key on the given feature as a string.
// Returns an empty string if there isn't a tag with that key.
func getString(context *api.Context, id b6.Identifiable, key string) (string, error) {
	if feature := api.Resolve(id, context.World); feature != nil {
		return feature.Get(key).Value.String(), nil
	}
	return "", nil
}

// Return the value of tag with the given key on the given feature as an integer.
// Returns error if there isn't a feature with that id, a tag with that key, or if the value isn't a valid integer.
func getInt(context *api.Context, id b6.Identifiable, key string) (int, error) {
	if feature := api.Resolve(id, context.World); feature != nil {
		return strconv.Atoi(feature.Get(key).Value.String())
	}
	return 0, fmt.Errorf("could not find feature with %s id", id.FeatureID().String())
}

// Return the value of tag with the given key on the given feature as a float.
// Returns error if there isn't a feature with that id, a tag with that key, or if the value isn't a valid float.
func getFloat(context *api.Context, id b6.Identifiable, key string) (float64, error) {
	if feature := api.Resolve(id, context.World); feature != nil {
		return strconv.ParseFloat(feature.Get(key).Value.String(), 64)
	}
	return 0.0, fmt.Errorf("could not find feature with %s id", id.FeatureID().String())
}

// Deprecated.
func countTagValue(context *api.Context, id b6.Identifiable, key string) (b6.Collection[interface{}, int], error) {
	c := &b6.ArrayCollection[interface{}, int]{
		Keys:   make([]interface{}, 0, 1),
		Values: make([]int, 0, 1),
	}
	if feature := api.Resolve(id, context.World); feature != nil {
		if tag := feature.Get(key); tag.IsValid() {
			c.Keys = append(c.Keys, api.StringStringPair{key, tag.Value.String()})
			c.Values = append(c.Values, 1)
		}
	}
	return c.Collection(), nil
}

// Return a collection of all the tags on the given feature.
// Keys are ordered integers from 0, values are tags.
func allTags(c *api.Context, id b6.Identifiable) (b6.Collection[int, b6.Tag], error) {
	var tags []b6.Tag
	if f := api.Resolve(id, c.World); f != nil {
		tags = f.AllTags()
	}
	return b6.ArrayValuesCollection[b6.Tag](tags).Collection(), nil
}

// Return true if the given feature matches the given query.
func matches(c *api.Context, id b6.Identifiable, query b6.Query) (bool, error) {
	if f := api.Resolve(id, c.World); f != nil {
		return query.Matches(f, c.World), nil
	}
	return false, nil
}

// Return the number of paths connected to the given point.
// A single path will be counted twice if the point isn't at one of its
// two ends - once in one direction, and once in the other.
func pointDegree(context *api.Context, point b6.Feature) (int, error) {
	segments := context.World.Traverse(point.FeatureID())
	n := 0
	for segments.Next() {
		n++
	}
	return n, nil
}

// Return the length of the given path in meters.
func pathLengthMeters(context *api.Context, path b6.Geometry) (float64, error) {
	return b6.AngleToMeters(path.Polyline().Length()), nil
}

type pathPointCollection struct {
	path b6.Geometry
	i    int
}

func (p *pathPointCollection) Begin() b6.Iterator[int, b6.Geometry] {
	return &pathPointCollection{path: p.path}
}

func (p *pathPointCollection) Count() (int, bool) {
	return p.path.GeometryLen(), true
}

func (p *pathPointCollection) Next() (bool, error) {
	if p.i >= p.path.GeometryLen() {
		return false, nil
	}
	p.i++
	return true, nil
}

func (p *pathPointCollection) Key() int {
	return p.i - 1
}

func (p *pathPointCollection) Value() b6.Geometry {
	return b6.GeometryFromLatLng(s2.LatLngFromPoint(p.path.PointAt(p.i - 1)))
}

func (p *pathPointCollection) KeyExpression() b6.Expression {
	return b6.NewIntExpression(p.Key())
}

func (p *pathPointCollection) ValueExpression() b6.Expression {
	return b6.NewPointExpressionFromLatLng(s2.LatLngFromPoint(p.Value().Point()))
}

var _ b6.AnyCollection[int, b6.Geometry] = &pathPointCollection{}

type areaPointCollection struct {
	area    b6.Area
	polygon *s2.Polygon
	loop    *s2.Loop
	i       int
	j       int
	k       int
	n       int
}

func (a *areaPointCollection) Begin() b6.Iterator[int, b6.Geometry] {
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

func (a *areaPointCollection) Value() b6.Geometry {
	return b6.GeometryFromLatLng(s2.LatLngFromPoint(a.loop.Vertex(a.k - 1)))
}

func (a *areaPointCollection) KeyExpression() b6.Expression {
	return b6.NewIntExpression(a.Key())
}

func (a *areaPointCollection) ValueExpression() b6.Expression {
	return b6.NewPointExpressionFromLatLng(s2.LatLngFromPoint(a.Value().Point()))
}

var _ b6.AnyCollection[int, b6.Geometry] = &areaPointCollection{}

// Return a collection of the points of the given geometry.
// Keys are ordered integers from 0, values are points.
func points(context *api.Context, geometry b6.Geometry) (b6.Collection[int, b6.Geometry], error) {
	switch geometry.GeometryType() {
	case b6.GeometryTypePoint:
		return b6.ArrayValuesCollection[b6.Geometry]([]b6.Geometry{geometry}).Collection(), nil
	case b6.GeometryTypePath:
		return b6.Collection[int, b6.Geometry]{
			AnyCollection: &pathPointCollection{path: geometry},
		}, nil
	case b6.GeometryTypeArea:
		return b6.Collection[int, b6.Geometry]{
			AnyCollection: &areaPointCollection{area: geometry.(b6.Area)},
		}, nil
	}
	return b6.ArrayValuesCollection[b6.Geometry]([]b6.Geometry{}).Collection(), nil
}

// Return a collection of the point features referenced by the given feature.
// Keys are ids of the respective value, values are point features. Area
// features return the points referenced by their path features.
func pointFeatures(context *api.Context, f b6.Feature) (b6.Collection[b6.FeatureID, b6.PhysicalFeature], error) {
	points := b6.ArrayFeatureCollection[b6.PhysicalFeature](make([]b6.PhysicalFeature, 0))
	if f, ok := f.(b6.PhysicalFeature); ok {
		switch f.GeometryType() {
		case b6.GeometryTypePoint:
			points = append(points, f)
		case b6.GeometryTypePath:
			for i := 0; i < f.GeometryLen(); i++ {
				if p := f.(b6.NestedPhysicalFeature).Feature(i); p != nil {
					points = append(points, p)
				}
			}
		case b6.GeometryTypeArea:
			for i := 0; i < f.(b6.AreaFeature).Len(); i++ {
				for _, path := range f.(b6.AreaFeature).Feature(i) {
					for j := 0; j < path.GeometryLen(); j++ {
						if p := path.Feature(j); p != nil {
							points = append(points, p)
						}
					}
				}
			}
		}
	}
	return points.Collection(), nil
}

// Return a collection of the path features referencing the given point.
// Keys are the ids of the respective paths.
func pointPaths(context *api.Context, id b6.Identifiable) (b6.Collection[b6.FeatureID, b6.PhysicalFeature], error) {
	p := context.World.FindFeatureByID(id.FeatureID())
	if p == nil {
		return b6.Collection[b6.FeatureID, b6.PhysicalFeature]{}, fmt.Errorf("No point with id %s", id)
	}
	collection := b6.ArrayFeatureCollection[b6.PhysicalFeature](make([]b6.PhysicalFeature, 0))
	paths := context.World.FindReferences(p.FeatureID(), b6.FeatureTypePath)
	for paths.Next() {
		if path, ok := paths.Feature().(b6.PhysicalFeature); ok {
			collection = append(collection, path)
		}
	}
	return collection.Collection(), nil
}

// Return a collection of points along the given paths, with the given distance in meters between them.
// Keys are the id of the respective path, values are points.
func samplePointsAlongPaths(context *api.Context, paths b6.Collection[b6.FeatureID, b6.Geometry], distanceMeters float64) (b6.Collection[int, b6.Geometry], error) {
	// TODO: We shouldn't need to special case this: we should be able to flatten the results of sample_points
	// on a collection of paths.
	seen := make(map[s2.Point]struct{})
	points := make([]b6.Geometry, 0, 16)
	i := paths.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return b6.Collection[int, b6.Geometry]{}, err
		}
		if !ok {
			break
		}
		points = appendUnseenSampledPoints(i.Value(), distanceMeters, seen, points)
	}
	return b6.ArrayValuesCollection[b6.Geometry](points).Collection(), nil
}

// Return a collection of points along the given path, with the given distance in meters between them.
// Keys are ordered integers from 0, values are points.
func samplePoints(context *api.Context, path b6.Geometry, distanceMeters float64) (b6.Collection[int, b6.Geometry], error) {
	points := appendUnseenSampledPoints(path, distanceMeters, make(map[s2.Point]struct{}), make([]b6.Geometry, 0, 16))
	return b6.ArrayValuesCollection[b6.Geometry](points).Collection(), nil
}

func appendUnseenSampledPoints(path b6.Geometry, distanceMeters float64, seen map[s2.Point]struct{}, points []b6.Geometry) []b6.Geometry {
	if path.GeometryType() == b6.GeometryTypePath {
		const epsilon s1.Angle = 1.6e-09 // Roughly 1cm
		polyline := path.Polyline()
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
				points = append(points, b6.GeometryFromLatLng(s2.LatLngFromPoint(p)))
				seen[p] = struct{}{}
			}
			j += step
		}
	}
	return points
}

// Return a path formed from the points of the two given paths, in the order they occur in those paths.
func join(context *api.Context, pathA b6.Geometry, pathB b6.Geometry) (b6.Geometry, error) {
	points := make([]s2.Point, 0, pathA.GeometryLen()+pathB.GeometryLen())
	i := 0
	for i < pathA.GeometryLen() {
		points = append(points, pathA.PointAt(i))
		i++
	}
	i = 0
	if pathB.PointAt(0) == pathA.PointAt(pathA.GeometryLen()-1) {
		i++
	}
	for i < pathB.GeometryLen() {
		points = append(points, pathB.PointAt(i))
		i++
	}
	return b6.GeometryFromPoints(points), nil
}

// Returns a path formed by joining the two given paths.
// If necessary to maintain consistency, the order of points is reversed,
// determined by which points are shared between the paths. Returns an error
// if no endpoints are shared.
func orderedJoin(context *api.Context, pathA b6.Geometry, pathB b6.Geometry) (b6.Geometry, error) {
	var reverseA, reverseB bool
	if pathA.PointAt(pathA.GeometryLen()-1) == pathB.PointAt(0) {
		reverseA, reverseB = false, false
	} else if pathA.PointAt(pathA.GeometryLen()-1) == pathB.PointAt(pathB.GeometryLen()-1) {
		reverseA, reverseB = false, true
	} else if pathA.PointAt(0) == pathB.PointAt(0) {
		reverseA, reverseB = true, false
	} else if pathA.PointAt(0) == pathB.PointAt(pathB.GeometryLen()-1) {
		reverseA, reverseB = true, true
	} else {
		return nil, fmt.Errorf("Paths don't share an end vertex")
	}
	points := make([]s2.Point, 0, pathA.GeometryLen()+pathB.GeometryLen())
	if reverseA {
		for i := pathA.GeometryLen() - 1; i >= 0; i-- {
			points = append(points, pathA.PointAt(i))
		}
	} else {
		for i := 0; i < pathA.GeometryLen(); i++ {
			points = append(points, pathA.PointAt(i))
		}
	}
	if reverseB {
		for i := pathB.GeometryLen() - 2; i >= 0; i-- {
			points = append(points, pathB.PointAt(i))
		}
	} else {
		for i := 1; i < pathB.GeometryLen(); i++ {
			points = append(points, pathB.PointAt(i))
		}
	}
	return b6.GeometryFromPoints(points), nil
}
