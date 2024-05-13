package functions

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/graph"
	"diagonal.works/b6/ingest"
	"golang.org/x/sync/errgroup"

	"github.com/golang/geo/s2"
)

func newShortestPathSearch(origin b6.Feature, options b6.UntypedCollection, distance float64, features graph.ShortestPathFeatures, w b6.World) (*graph.ShortestPathSearch, error) {
	weights, err := WeightsFromOptions(options)
	if err != nil {
		return nil, err
	}

	var s *graph.ShortestPathSearch
	if origin, ok := origin.(b6.PhysicalFeature); ok {
		switch origin.GeometryType() {
		case b6.GeometryTypePoint:
			s = graph.NewShortestPathSearchFromPoint(origin.FeatureID())
		case b6.GeometryTypeArea:
			s = graph.NewShortestPathSearchFromBuilding(origin.(b6.AreaFeature), weights, w)
		}

		s.ExpandSearch(distance, weights, features, w)
		return s, nil
	}

	return nil, fmt.Errorf("Can't find paths from feature type %s", origin.FeatureID().Type)
}

func FindReachableFeaturesWithPathStates(context *api.Context, origin b6.Feature, options b6.UntypedCollection, distance float64, query b6.Query, pathStates *geojson.FeatureCollection) (b6.Collection[b6.FeatureID, b6.Feature], error) {
	features := b6.ArrayFeatureCollection[b6.Feature](make([]b6.Feature, 0))
	s, err := newShortestPathSearch(origin, options, distance, graph.PointsAndAreas, context.World)
	if err == nil {
		for id := range s.PointDistances() {
			if point := context.World.FindFeatureByID(id); point != nil {
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

// Return the a collection of the features reachable from the given origin via the given mode, within the given distance in meters, that match the given query.
// See accessible-all for options values.
// Deprecated. Use accessible-all.
func reachable(context *api.Context, origin b6.Feature, options b6.UntypedCollection, distance float64, query b6.Query) (b6.Collection[b6.FeatureID, b6.Feature], error) {
	return FindReachableFeaturesWithPathStates(context, origin, options, distance, query, nil)
}

type odCollection struct {
	origins      []b6.FeatureID
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

func (o *odCollection) Begin() b6.Iterator[b6.FeatureID, b6.FeatureID] {
	return &odCollection{origins: o.origins, destinations: o.destinations}
}

func (o *odCollection) Key() b6.FeatureID {
	return o.origins[o.i-1].FeatureID()
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

func (o *odCollection) KeyExpression() b6.Expression {
	return b6.NewFeatureIDExpression(o.Key())
}

func (o *odCollection) ValueExpression() b6.Expression {
	return b6.NewFeatureIDExpression(o.Value())
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

// Return the a collection of the features reachable from the given origins, within the given duration in seconds, that match the given query.
// Keys of the collection are origins, values are reachable destinations.
// Options are passed as tags containing the mode, and mode specific values. Examples include:
// Walking, with the default speed of 4.5km/h:
// mode=walk
// Walking, a speed of 3km/h:
// mode=walk, walking speed=3.0
// Transit at peak times:
// mode=transit
// Transit at off-peak times:
// mode=transit, peak=no
// Walking, accounting for elevation:
// elevation=true (optional: uphill=hard downhill=hard)
// Walking, with the resulting collection flipped such that keys are
// destinations and values are origins. Useful for efficiency if you assume
// symmetry, and the number of destinations is considerably smaller than the
// number of origins:
// mode=walk, flip=yes
func accessibleAll(context *api.Context, origins b6.Collection[any, b6.Identifiable], destinations b6.Query, duration float64, options b6.UntypedCollection) (b6.Collection[b6.FeatureID, b6.FeatureID], error) {
	tags, err := api.CollectionToTags(options)
	if err != nil {
		return b6.Collection[b6.FeatureID, b6.FeatureID]{}, err
	}

	weights, err := WeightsFromOptions(options)
	if err != nil {
		return b6.Collection[b6.FeatureID, b6.FeatureID]{}, err
	}

	os := make([]b6.FeatureID, 0)
	i := origins.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return b6.Collection[b6.FeatureID, b6.FeatureID]{}, err
		} else if !ok {
			break
		}
		os = append(os, i.Value().FeatureID())
	}

	ds := make([][]b6.FeatureID, len(os))
	c := make(chan int)
	g, gc := errgroup.WithContext(context.Context)
	for i := 0; i < context.Cores; i++ {
		g.Go(func() error {
			for j := range c {
				if origin := context.World.FindFeatureByID(os[j]); origin != nil {
					ds[j] = accessibleFromOrigin(ds[j], origin, destinations, weights, duration, context.World)
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
	if flip := tags.Get("flip"); flip.Value.String() == "yes" {
		ods.Flip()
	} else {
		for i := range ods.destinations {
			if len(ods.destinations[i]) == 0 {
				ods.destinations[i] = append(ods.destinations[i], b6.FeatureIDInvalid)
			}
		}
	}
	sort.Sort(ods)
	return b6.Collection[b6.FeatureID, b6.FeatureID]{
		AnyCollection: ods,
	}, err
}

func WeightsFromOptions(options b6.UntypedCollection) (graph.Weights, error) {
	opts, err := api.CollectionToTags(options)
	if err != nil {
		return nil, err
	}

	var weights graph.Weights

	if opts.Get("elevation").IsValid() {
		elevation := graph.ElevationWeights{}
		if upHill := opts.Get("uphill"); upHill.IsValid() && upHill.Value.String() == "hard" {
			elevation.UpHillHard = true
		}
		if downHill := opts.Get("downhill"); downHill.IsValid() && downHill.Value.String() == "hard" {
			elevation.DownHillHard = true
		}
		weights = elevation
	} else {
		walking := graph.WalkingTimeWeights{
			Speed: graph.WalkingMetersPerSecond,
		}

		if speed := opts.Get("walking speed"); speed.IsValid() {
			if f, err := strconv.ParseFloat(speed.Value.String(), 64); err == nil {
				walking.Speed = f
			}
		}
		weights = walking
	}

	switch m := opts.Get("mode").Value.String(); m {
	case "", "walk":
	case "transit":
		if p := opts.Get("peak"); p.Value.String() == "no" {
			weights = graph.TransitTimeWeights{PeakTraffic: false, Weights: weights}
		} else {
			weights = graph.TransitTimeWeights{PeakTraffic: true, Weights: weights}
		}
	default:
		return nil, fmt.Errorf("Expected mode=walk or mode=transit, found %s", m)
	}

	return weights, nil
}

func accessibleRoutes(context *api.Context, origin b6.Identifiable, destinations b6.Query, duration float64, options b6.UntypedCollection) (b6.Collection[b6.FeatureID, b6.Route], error) {
	f := api.Resolve(origin, context.World)
	if f == nil {
		return b6.Collection[b6.FeatureID, b6.Route]{}, nil
	}

	weights, err := WeightsFromOptions(options)
	if err != nil {
		return b6.Collection[b6.FeatureID, b6.Route]{}, err
	}

	s := graph.NewShortestPathSearchFromFeature(f, weights, context.World)
	s.ExpandSearch(duration, weights, graph.PointsAndAreas, context.World)
	routes := s.AllRoutes()
	c := b6.ArrayCollection[b6.FeatureID, b6.Route]{}
	for id, route := range routes {
		// TODO: This can be more efficient by only building routes for
		// features that match a query.
		if f := context.World.FindFeatureByID(id); f != nil {
			if destinations.Matches(f, context.World) {
				c.Keys = append(c.Keys, id)
				c.Values = append(c.Values, route)
			}
		}

	}
	return c.Collection(), nil
}

// Return a collection containing only the values of the given collection that match the given query.
// If no values for a key match the query, emit a single invalid feature ID
// for that key, allowing callers to count the number of keys with no valid
// values.
// Keys are taken from the given collection.
func filterAccessible(context *api.Context, collection b6.Collection[b6.FeatureID, b6.FeatureID], filter b6.Query) (b6.Collection[b6.FeatureID, b6.FeatureID], error) {
	filtered := b6.ArrayCollection[b6.FeatureID, b6.FeatureID]{}
	i := collection.Begin()
	last := b6.FeatureIDInvalid
	empty := true
	for {
		ok, err := i.Next()
		if err != nil || !ok {
			return filtered.Collection(), err
		}
		if last != i.Key() {
			if last.IsValid() && empty {
				filtered.Keys = append(filtered.Keys, last)
				filtered.Values = append(filtered.Values, b6.FeatureIDInvalid)
			}
			last = i.Key()
			empty = true
		}
		if f := context.World.FindFeatureByID(i.Value()); f != nil {
			if filter.Matches(f, context.World) {
				empty = false
				filtered.Keys = append(filtered.Keys, i.Key())
				filtered.Values = append(filtered.Values, i.Value())
			}
		}
	}
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
		if id == origin.FeatureID() {
			continue
		}
		if point := w.FindFeatureByID(id); point != nil {
			if destinations.Matches(point, w) {
				ds = append(ds, point.FeatureID())
			}
		}
	}
	return ds
}

// Return the closest feature from the given origin via the given mode, within the given distance in meters, matching the given query.
// See accessible-all for options values.
func closestFeature(context *api.Context, origin b6.Feature, options b6.UntypedCollection, distance float64, query b6.Query) (b6.Feature, error) {
	feature, _, err := findClosest(context, origin, options, distance, query)
	return feature, err
}

// Return the distance through the graph of the closest feature from the given origin via the given mode, within the given distance in meters, matching the given query.
// See accessible-all for options values.
func closestFeatureDistance(context *api.Context, origin b6.Feature, options b6.UntypedCollection, distance float64, query b6.Query) (float64, error) {
	_, distance, err := findClosest(context, origin, options, distance, query)
	return distance, err
}

func findClosest(context *api.Context, origin b6.Feature, options b6.UntypedCollection, distance float64, query b6.Query) (b6.Feature, float64, error) {
	s, err := newShortestPathSearch(origin, options, distance, graph.PointsAndAreas, context.World)
	if err == nil {
		// TODO: This expands the search everywhere up to the maximum distance, and we
		// can actually stop early.
		distance := math.Inf(1)
		var closest b6.Feature
		for id, d := range s.PointDistances() {
			if d < distance {
				if point := context.World.FindFeatureByID(id); point != nil {
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

// Return a collection of the paths used to reach all features matching the given query from the given origin via the given mode, within the given distance in meters.
// Keys are the paths used, values are the number of times that path was used during traversal.
// See accessible-all for options values.
func pathsToReachFeatures(context *api.Context, origin b6.Feature, options b6.UntypedCollection, distance float64, query b6.Query) (b6.Collection[b6.FeatureID, int], error) {
	features := &b6.ArrayCollection[b6.FeatureID, int]{
		Keys:   make([]b6.FeatureID, 0),
		Values: make([]int, 0),
	}
	s, err := newShortestPathSearch(origin, options, distance, graph.PointsAndAreas, context.World)
	if err == nil {
		points := 0
		counts := make(map[b6.FeatureID]int)
		for id := range s.PointDistances() {
			if point := context.World.FindFeatureByID(id); point != nil {
				if query.Matches(point, context.World) {
					points++
					last := b6.FeatureIDInvalid
					for _, segment := range s.BuildPath(id) {
						if segment.Feature.FeatureID() != last {
							counts[segment.Feature.FeatureID()]++
							last = segment.Feature.FeatureID()
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
					if point := context.World.FindFeatureByID(pointID); point != nil {
						last := b6.FeatureIDInvalid
						for _, segment := range s.BuildPath(pointID) {
							if segment.Feature.FeatureID() != last {
								counts[segment.Feature.FeatureID()]++
								last = segment.Feature.FeatureID()
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

// Return the area formed by the convex hull of the features matching the given query reachable from the given origin via the given mode specified in options, within the given distance in meters.
// See accessible-all for options values.
func reachableArea(context *api.Context, origin b6.Feature, options b6.UntypedCollection, distance float64) (float64, error) {
	area := 0.0
	s, err := newShortestPathSearch(origin, options, distance, graph.Points, context.World)
	if err == nil {
		distances := s.PointDistances()
		query := s2.NewConvexHullQuery()
		for id := range distances {
			if point := context.World.FindFeatureByID(id); point != nil {
				if p, ok := point.(b6.Geometry); ok && p.GeometryType() == b6.GeometryTypePoint {
					query.AddPoint(p.Point())
				}
			}
		}
		area = query.ConvexHull().Area()
	}
	return area, err
}

// Add a path that connects the two given points, if they're not already directly connected.
func connect(c *api.Context, a b6.Feature, b b6.Feature) (ingest.Change, error) {
	add := &ingest.AddFeatures{}
	segments := c.World.Traverse(a.FeatureID())
	connected := false
	for segments.Next() {
		segment := segments.Segment()
		if segment.LastFeatureID() == b.FeatureID() {
			connected = true
			break
		}
	}
	if !connected {
		path := ingest.GenericFeature{}
		path.SetFeatureID(b6.FeatureID{b6.FeatureTypePath, b6.NamespaceDiagonalAccessPoints, 1})
		path.ModifyOrAddTag(b6.Tag{b6.PathTag, b6.Values([]b6.Value{a.FeatureID(), b.FeatureID()})})
		*add = append(*add, &path)
	}
	return add, nil
}

// Add a path and point to connect given feature to the street network.
// The street network is defined at the set of paths tagged #highway that
// allow traversal of more than 500m. A point is added to the closest
// network path at the projection of the origin point on that path, unless
// that point is within 4m of an existing path point.
func connectToNetwork(c *api.Context, feature b6.Feature) (ingest.Change, error) {
	q := b6.Intersection{b6.Keyed{Key: "#highway"}}
	if p, ok := feature.(b6.PhysicalFeature); ok {
		if centroid, ok := b6.Centroid(p); ok {
			q = append(q, b6.NewIntersectsCap(s2.CapFromCenterAngle(centroid, b6.MetersToAngle(500.0))))
		}
	} else {
		return nil, fmt.Errorf("expected a PhysicalFeature, found: %T", feature)
	}
	highways := c.World.FindFeatures(b6.Typed{b6.FeatureTypePath, q})
	network := graph.BuildStreetNetwork(highways, b6.MetersToAngle(500.0), graph.SimpleHighwayWeights{}, nil, c.World)
	strategy := graph.UseExisitingPoints{Connections: graph.NewConnections()}
	graph.ConnectFeature(feature, network, b6.MetersToAngle(500.0), c.World, strategy)
	strategy.Finish()
	return strategy.Connections.Change(c.World), nil
}

// Add paths and points to connect the given collection of features to the
// network. See connect-to-network for connection details.
// More efficient than using map with connect-to-network, as the street
// network is only computed once.
func connectToNetworkAll(c *api.Context, features b6.Collection[any, b6.FeatureID]) (ingest.Change, error) {
	highways := c.World.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.Keyed{Key: "#highway"}})
	network := graph.BuildStreetNetwork(highways, b6.MetersToAngle(500.0), graph.SimpleHighwayWeights{}, nil, c.World)
	strategy := graph.UseExisitingPoints{Connections: graph.NewConnections()}
	i := features.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		if f := c.World.FindFeatureByID(i.Value()); f != nil {
			graph.ConnectFeature(f, network, b6.MetersToAngle(500.0), c.World, strategy)
		}
	}
	strategy.Finish()
	return strategy.Connections.Change(c.World), nil
}
