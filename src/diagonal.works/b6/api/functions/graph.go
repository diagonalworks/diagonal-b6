package functions

import (
	"fmt"
	"math"
	"sort"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/graph"
	"diagonal.works/b6/ingest"
	"golang.org/x/sync/errgroup"

	"github.com/golang/geo/s2"
)

func newShortestPathSearch(origin b6.Feature, mode string, distance float64, features graph.ShortestPathFeatures, w b6.World) (*graph.ShortestPathSearch, error) {
	var weights graph.Weights
	switch mode {
	case "bus":
		weights = graph.BusWeights{}
	case "car":
		weights = graph.CarWeights{}
	case "walk":
		weights = graph.SimpleHighwayWeights{}
	default:
		return nil, fmt.Errorf("Unknown travel mode %q", mode)
	}

	var s *graph.ShortestPathSearch
	switch origin := origin.(type) {
	case b6.PointFeature:
		s = graph.NewShortestPathSearchFromPoint(origin.PointID())
	case b6.AreaFeature:
		s = graph.NewShortestPathSearchFromBuilding(origin, weights, w)
	default:
		return nil, fmt.Errorf("Can't find paths from feature type %s", origin.FeatureID().Type)
	}
	s.ExpandSearch(distance, weights, features, w)
	return s, nil
}

func reachablePoints(context *api.Context, origin b6.Feature, mode string, distance float64, query b6.Query) (b6.Collection[b6.FeatureID, b6.PointFeature], error) {
	points := b6.ArrayFeatureCollection[b6.PointFeature](make([]b6.PointFeature, 0))
	s, err := newShortestPathSearch(origin, mode, distance, graph.Points, context.World)
	if err == nil {
		for id := range s.PointDistances() {
			if point := b6.FindPointByID(id, context.World); point != nil {
				if query.Matches(point, context.World) {
					points = append(points, point)
				}
			}
		}
	}
	return points.Collection(), err
}

func FindReachableFeaturesWithPathStates(context *api.Context, origin b6.Feature, mode string, distance float64, query b6.Query, pathStates *geojson.FeatureCollection) (b6.Collection[b6.FeatureID, b6.Feature], error) {
	features := b6.ArrayFeatureCollection[b6.Feature](make([]b6.Feature, 0))
	s, err := newShortestPathSearch(origin, mode, distance, graph.PointsAndAreas, context.World)
	if err == nil {
		for id := range s.PointDistances() {
			if point := b6.FindPointByID(id, context.World); point != nil {
				if query.Matches(point, context.World) {
					features = append(features, point)
				}
			}
		}
		for id := range s.AreaDistances() {
			if area := b6.FindAreaByID(id, context.World); area != nil {
				if query.Matches(area, context.World) {
					features = append(features, area)
				}
			}
		}
		if pathStates != nil {
			for key, state := range s.PathStates() {
				segment := b6.FindPathSegmentByKey(key, context.World)
				polyline := segment.Polyline()
				geometry := geojson.GeometryFromLineString(geojson.FromPolyline(polyline))
				shape := geojson.NewFeatureWithGeometry(geometry)
				pathStates.AddFeature(shape)
				label := geojson.NewFeatureFromS2Point(polyline.Centroid())
				switch state {
				case graph.PathStateTraversed:
					shape.Properties["colour"] = "#00ff00"
				case graph.PathStateTooFar:
					shape.Properties["colour"] = "#ff0000"
					label.Properties["label"] = "Too far"
					pathStates.AddFeature(label)
				case graph.PathStateNotUseable:
					shape.Properties["colour"] = "#ff0000"
					label.Properties["label"] = "Not useable"
					pathStates.AddFeature(label)
				}
				pathStates.AddFeature(shape)
			}
		}
	}
	return features.Collection(), err
}

func reachable(context *api.Context, origin b6.Feature, mode string, distance float64, query b6.Query) (b6.Collection[b6.FeatureID, b6.Feature], error) {
	return FindReachableFeaturesWithPathStates(context, origin, mode, distance, query, nil)
}

type odCollection struct {
	origins      []b6.Identifiable
	destinations [][]b6.FeatureID

	i int
	j int
}

func (o *odCollection) IsCountable() bool { return true }

func (o *odCollection) Count() (int, bool) {
	n := 0
	for _, ds := range o.destinations {
		n += len(ds)
	}
	return n, true
}

func (o *odCollection) Begin() b6.Iterator[b6.Identifiable, b6.FeatureID] {
	return &odCollection{origins: o.origins, destinations: o.destinations}
}

func (o *odCollection) Key() b6.Identifiable {
	return o.origins[o.i-1]
}

func (o *odCollection) Value() b6.FeatureID {
	return o.destinations[o.i-1][o.j-1]
}

func (o *odCollection) Next() (bool, error) {
	o.j++
	if o.i < 1 || o.j > len(o.destinations[o.i-1]) {
		o.j = 1
		for {
			o.i++
			if o.i > len(o.origins) || o.j <= len(o.destinations[o.i-1]) {
				break
			}
		}
	}
	return o.i <= len(o.origins), nil
}

func (o *odCollection) Flip() {
	flipped := make(map[b6.FeatureID][]b6.FeatureID)
	for i, origin := range o.origins {
		for _, destination := range o.destinations[i] {
			flipped[destination] = append(flipped[destination], origin.FeatureID())
		}
	}
	o.origins = o.origins[0:0]
	o.destinations = o.destinations[0:0]
	for origin, destinations := range flipped {
		o.origins = append(o.origins, origin)
		o.destinations = append(o.destinations, destinations)
	}
}

func (o *odCollection) Len() int { return len(o.origins) }

func (o *odCollection) Swap(i, j int) {
	o.origins[i], o.origins[j] = o.origins[j], o.origins[i]
	o.destinations[i], o.destinations[j] = o.destinations[j], o.destinations[i]
}

func (o *odCollection) Less(i, j int) bool {
	return o.origins[i].FeatureID().Less(o.origins[j].FeatureID())
}

func accessible(context *api.Context, origins b6.Collection[any, b6.Identifiable], destinations b6.Query, distance float64, options b6.UntypedCollection) (b6.Collection[b6.Identifiable, b6.FeatureID], error) {
	tags, err := api.CollectionToTags(options)
	if err != nil {
		return b6.Collection[b6.Identifiable, b6.FeatureID]{}, err
	}

	os := make([]b6.Identifiable, 0)
	i := origins.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return b6.Collection[b6.Identifiable, b6.FeatureID]{}, err
		} else if !ok {
			break
		}
		os = append(os, i.Value())
	}

	mode := "walk"
	if m := tags.Get("mode"); m.IsValid() {
		mode = m.Value
	}
	weights, err := weightsFromMode(mode)
	if err != nil {
		return b6.Collection[b6.Identifiable, b6.FeatureID]{}, err
	}

	ds := make([][]b6.FeatureID, len(os))
	c := make(chan int)
	g, gc := errgroup.WithContext(context.Context)
	for i := 0; i < context.Cores; i++ {
		g.Go(func() error {
			for j := range c {
				if origin := api.Resolve(os[j], context.World); origin != nil {
					ds[j] = accessibleFromOrigin(ds[j], origin, destinations, weights, distance, context.World)
				}
			}
			return nil
		})
	}

done:
	for i := range os {
		select {
		case <-gc.Done():
			break done
		case c <- i:
		}
	}
	close(c)
	err = g.Wait()
	ods := &odCollection{origins: os, destinations: ds}
	if flip := tags.Get("flip"); flip.Value == "yes" {
		ods.Flip()
	} else {
		for i := range ods.destinations {
			if len(ods.destinations[i]) == 0 {
				ods.destinations[i] = append(ods.destinations[i], b6.FeatureIDInvalid)
			}
		}
	}
	sort.Sort(ods)
	return b6.Collection[b6.Identifiable, b6.FeatureID]{
		AnyCollection: ods,
	}, err
}

func accessibleFromOrigin(ds []b6.FeatureID, origin b6.Feature, destinations b6.Query, weights graph.Weights, distance float64, w b6.World) []b6.FeatureID {
	s := graph.NewShortestPathSearchFromFeature(origin, weights, w)
	s.ExpandSearch(distance, weights, graph.PointsAndAreas, w)
	for id := range s.AreaDistances() {
		if id.FeatureID() == origin.FeatureID() {
			continue
		}
		if area := w.FindFeatureByID(id.FeatureID()); area != nil {
			if destinations.Matches(area, w) {
				ds = append(ds, area.FeatureID())
			}
		}
	}
	for id := range s.PointDistances() {
		if id.FeatureID() == origin.FeatureID() {
			continue
		}
		if point := w.FindFeatureByID(id.FeatureID()); point != nil {
			if destinations.Matches(point, w) {
				ds = append(ds, point.FeatureID())
			}
		}
	}
	return ds
}

func weightsFromMode(mode string) (graph.Weights, error) {
	switch mode {
	case "bus":
		return graph.BusWeights{}, nil
	case "car":
		return graph.CarWeights{}, nil
	case "walk":
		return graph.SimpleHighwayWeights{}, nil
	}
	return nil, fmt.Errorf("Unknown travel mode %q", mode)
}

func closestFeature(context *api.Context, origin b6.Feature, mode string, distance float64, query b6.Query) (b6.Feature, error) {
	feature, _, err := findClosest(context, origin, mode, distance, query)
	return feature, err
}

// findClosestFeatureDistance returns the distance to the closest matching feature.
// Ideally, we'd either return the distance along with the feature as a pair from, or
// return a new primitive Route instance that described the route to that feature,
// allowing distance to be derived. Neither are possible right now, so this is a
// stopgap. TODO: Improve this API.
func closestFeatureDistance(context *api.Context, origin b6.Feature, mode string, distance float64, query b6.Query) (float64, error) {
	_, distance, err := findClosest(context, origin, mode, distance, query)
	return distance, err
}

func findClosest(context *api.Context, origin b6.Feature, mode string, distance float64, query b6.Query) (b6.Feature, float64, error) {
	s, err := newShortestPathSearch(origin, mode, distance, graph.PointsAndAreas, context.World)
	if err == nil {
		// TODO: This expands the search everywhere up to the maximum distance, and we
		// can actually stop early.
		distance := math.Inf(1)
		var closest b6.Feature
		for id, d := range s.PointDistances() {
			if d < distance {
				if point := b6.FindPointByID(id, context.World); point != nil {
					if query.Matches(point, context.World) {
						distance = d
						closest = point
					}
				}
			}
		}
		for id, d := range s.AreaDistances() {
			if d < distance {
				if area := b6.FindAreaByID(id, context.World); area != nil {
					if query.Matches(area, context.World) {
						distance = d
						closest = area
					}
				}
			}
		}
		if closest != nil {
			return closest, distance, nil
		}
	}
	return nil, 0.0, err
}

func pathsToReachFeatures(context *api.Context, origin b6.Feature, mode string, distance float64, query b6.Query) (b6.Collection[b6.FeatureID, int], error) {
	features := &b6.ArrayCollection[b6.FeatureID, int]{
		Keys:   make([]b6.FeatureID, 0),
		Values: make([]int, 0),
	}
	s, err := newShortestPathSearch(origin, mode, distance, graph.PointsAndAreas, context.World)
	if err == nil {
		points := 0
		counts := make(map[b6.PathID]int)
		for id := range s.PointDistances() {
			if point := b6.FindPointByID(id, context.World); point != nil {
				if query.Matches(point, context.World) {
					points++
					last := b6.PathIDInvalid
					for _, segment := range s.BuildPath(id) {
						if segment.Feature.PathID() != last {
							counts[segment.Feature.PathID()]++
							last = segment.Feature.PathID()
						}
					}
				}
			}
		}

		areas := 0
		for areaID, pointID := range s.AreaEntrances() {
			if area := b6.FindAreaByID(areaID, context.World); area != nil {
				if query.Matches(area, context.World) {
					areas++
					if point := b6.FindPointByID(pointID, context.World); point != nil {
						last := b6.PathIDInvalid
						for _, segment := range s.BuildPath(pointID) {
							if segment.Feature.PathID() != last {
								counts[segment.Feature.PathID()]++
								last = segment.Feature.PathID()
							}
						}
					}
				}
			}
		}

		for id, count := range counts {
			features.Keys = append(features.Keys, id.FeatureID())
			features.Values = append(features.Values, count)
		}
	}
	return features.Collection(), err
}

func reachableArea(context *api.Context, origin b6.Feature, mode string, distance float64) (float64, error) {
	area := 0.0
	s, err := newShortestPathSearch(origin, mode, distance, graph.Points, context.World)
	if err == nil {
		distances := s.PointDistances()
		query := s2.NewConvexHullQuery()
		for id := range distances {
			if point := b6.FindPointByID(id, context.World); point != nil {
				query.AddPoint(point.Point())
			}
		}
		area = query.ConvexHull().Area()
	}
	return area, err
}

func connect(c *api.Context, a b6.PointFeature, b b6.PointFeature) (ingest.Change, error) {
	add := &ingest.AddFeatures{
		IDsToReplace: map[b6.Namespace]b6.Namespace{b6.NamespacePrivate: b6.NamespaceDiagonalAccessPoints},
	}
	segments := c.World.Traverse(a.PointID())
	connected := false
	for segments.Next() {
		segment := segments.Segment()
		if segment.LastFeature().PointID() == b.PointID() {
			connected = true
			break
		}
	}
	if !connected {
		path := ingest.NewPathFeature(2)
		path.PathID = b6.MakePathID(b6.NamespacePrivate, 1)
		path.SetPointID(0, a.PointID())
		path.SetPointID(1, b.PointID())
		add.Paths = append(add.Paths, path)
	}
	return add, nil
}

func connectToNetwork(c *api.Context, feature b6.Feature) (ingest.Change, error) {
	highways := b6.FindPaths(b6.Keyed{Key: "#highway"}, c.World)
	network := graph.BuildStreetNetwork(highways, b6.MetersToAngle(500.0), graph.SimpleHighwayWeights{}, nil, c.World)
	connections := graph.NewConnections()
	strategy := graph.InsertNewPointsIntoPaths{
		Connections:      connections,
		World:            c.World,
		ClusterThreshold: b6.MetersToAngle(4.0),
	}
	graph.ConnectFeature(feature, network, b6.MetersToAngle(500.0), c.World, strategy)
	strategy.Finish()
	return strategy.Connections.Change(c.World), nil
}
