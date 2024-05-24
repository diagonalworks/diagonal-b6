package graph

import (
	"math"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/osm"
	"diagonal.works/b6/test/camden"
)

func TestShortestPath(t *testing.T) {
	camden := camden.BuildCamdenForTests(t)

	fromPath := camden.FindFeatureByID(ingest.FromOSMWayID(687471322))
	if fromPath == nil {
		t.Fatal("Failed to find way")
	}
	from := fromPath.Reference(0).Source()
	if !from.IsValid() {
		t.Fatal("Expected a point reference")
	}

	toPath := camden.FindFeatureByID(ingest.FromOSMWayID(367808662))
	if toPath == nil {
		t.Fatal("Failed to find way")
	}
	to := toPath.Reference(0).Source()
	if !to.IsValid() {
		t.Fatal("Expected a point reference")
	}

	path := ComputeShortestPath(from, to, 1000.0, BusWeights{}, camden)
	wayIDs := make(map[osm.WayID]bool)
	for _, segment := range path {
		wayIDs[osm.WayID(segment.Feature.FeatureID().Value)] = true
	}

	expected := []osm.WayID{673733343, 207107599}
	for _, wayID := range expected {
		if !wayIDs[wayID] {
			t.Errorf("Expected to find way %d", wayID)
		}
	}

	notExpected := []osm.WayID{
		681764413, // Cycleway along Midland Road
		673447483, // Highway that's not on the shortest path
	}
	for _, wayID := range notExpected {
		if wayIDs[wayID] {
			t.Errorf("Didn't expect to find way %d", wayID)
		}
	}
}

func TestShortestPathWithOverriddenWeight(t *testing.T) {
	nodes := []osm.Node{
		// Intersections between Royal College Street and the cyclepath
		{ID: 7799663850, Location: osm.LatLng{Lat: 51.5409703, Lng: -0.1376308}},
		{ID: 5336117979, Location: osm.LatLng{Lat: 51.5416858, Lng: -0.1382541}},

		// Part of the cyclepath, but not the road
		{ID: 4931754288, Location: osm.LatLng{Lat: 51.5416379, Lng: -0.1382604}},
	}

	ways := []osm.Way{
		// Royal College Street
		{ID: 835622320, Nodes: []osm.NodeID{7799663850, 5336117979}},
		// Cycleway along Royal College Street
		{ID: 835622319, Nodes: []osm.NodeID{7799663850, 4931754288, 5336117979}},
	}

	w, err := ingest.BuildWorldFromOSM(nodes, ways, []osm.Relation{}, &ingest.BuildOptions{Cores: 2})
	if err != nil {
		t.Fatal(err)
	}

	from := ingest.FromOSMNodeID(7799663850)
	to := ingest.FromOSMNodeID(5336117979)
	path := ComputeShortestPath(from, to, 500.0, SimpleWeights{}, w)
	if len(path) != 1 || path[0].Feature.FeatureID().Value != uint64(ways[0].ID) {
		t.Error("Expected shortest path to use road")
	}

	// Override the weight of the cyclepath, and ensure we're routed down it
	ways[1].Tags = []osm.Tag{{Key: "diagonal:weight", Value: "0.1"}}
	w, err = ingest.BuildWorldFromOSM(nodes, ways, []osm.Relation{}, &ingest.BuildOptions{Cores: 2})
	if err != nil {
		t.Fatal(err)
	}
	path = ComputeShortestPath(from, to, 500.0, SimpleWeights{}, w)
	if len(path) != 1 || path[0].Feature.FeatureID().Value != uint64(ways[1].ID) {
		t.Error("Expected shortest path to use cycleway")
	}
}

func TestShortestPathWithTwoJoinedPaths(t *testing.T) {
	nodes := []osm.Node{
		{ID: 5384190463, Location: osm.LatLng{Lat: 51.5358664, Lng: -0.1272493}},
		{ID: 5384190494, Location: osm.LatLng{Lat: 51.5362126, Lng: -0.1270125}},
		{ID: 5384190476, Location: osm.LatLng{Lat: 51.5367563, Lng: -0.1266297}},
	}

	// Two ways, one ending with node 5384190494, one starting with it.
	ways := []osm.Way{
		{ID: 558345071, Nodes: []osm.NodeID{5384190463, 5384190494}},
		{ID: 558345054, Nodes: []osm.NodeID{5384190494, 5384190476}},
	}

	w, err := ingest.BuildWorldFromOSM(nodes, ways, []osm.Relation{}, &ingest.BuildOptions{Cores: 2})
	if err != nil {
		t.Fatal(err)
	}
	from := ingest.FromOSMNodeID(5384190494)
	to := ingest.FromOSMNodeID(5384190463)
	path := ComputeShortestPath(from, to, 500.0, SimpleWeights{}, w)

	expected := 1
	if len(path) == expected {
		if path[0].Feature.FeatureID().Value != uint64(ways[0].ID) {
			t.Errorf("Expected way %d, found %d", ways[0].ID, path[0].Feature.FeatureID().Value)
		}
	} else {
		t.Errorf("Expected path with %d segments, found %d", expected, len(path))
	}
}

func TestAccessibilityWithTwoJoinedPaths(t *testing.T) {
	nodes := []osm.Node{
		{ID: 5384190463, Location: osm.LatLng{Lat: 51.5358664, Lng: -0.1272493}},
		{ID: 5384190494, Location: osm.LatLng{Lat: 51.5362126, Lng: -0.1270125}},
		{ID: 5384190476, Location: osm.LatLng{Lat: 51.5367563, Lng: -0.1266297}},
	}

	// Two ways, one ending with node 5384190494, one starting with it.
	ways := []osm.Way{
		{ID: 558345071, Nodes: []osm.NodeID{5384190463, 5384190494}},
		{ID: 558345054, Nodes: []osm.NodeID{5384190494, 5384190476}},
	}

	w, err := ingest.BuildWorldFromOSM(nodes, ways, []osm.Relation{}, &ingest.BuildOptions{Cores: 2})
	if err != nil {
		t.Fatal(err)
	}
	from := ingest.FromOSMNodeID(5384190494)

	_, counts := ComputeAccessibility(from, 500.0, SimpleWeights{}, w)
	for _, way := range ways {
		id := ingest.FromOSMWayID(way.ID)
		key := b6.SegmentKey{ID: id, First: 0, Last: 1}
		expected := 1
		if count, ok := counts[key]; !ok || count != expected {
			t.Errorf("Expected count of %d on %s, found %d", expected, id, count)
		}
	}
}

func TestShortestPathTakesIntoAccountOneWayStreets(t *testing.T) {
	// Tests that shortest path routing doesn't take you down one way streets. The
	// test case is the road junction here: 51.5452312, -0.1415558, where the West
	// hand fork is shorter when heading South, but oneway in the wrong direction.
	camden := camden.BuildCamdenForTests(t)

	from := camden.FindFeatureByID(ingest.FromOSMNodeID(33000703))
	if from == nil {
		t.Fatal("Failed to find from node")
	}

	to := camden.FindFeatureByID(ingest.FromOSMNodeID(970237231))
	if to == nil {
		t.Fatal("Failed to find to node")
	}

	path := ComputeShortestPath(from.FeatureID(), to.FeatureID(), 500.0, BusWeights{}, camden)
	wayIDs := make(map[osm.WayID]bool)
	for _, segment := range path {
		wayIDs[osm.WayID(segment.Feature.FeatureID().Value)] = true
	}

	expected := []osm.WayID{
		835618252, // Oneway, heading in the corect direction
	}
	for _, wayID := range expected {
		if !wayIDs[wayID] {
			t.Errorf("Expected to find way %d", wayID)
		}
	}

	notExpected := []osm.WayID{
		502802551, // Oneway, heading in the wrong direction
	}
	for _, wayID := range notExpected {
		if wayIDs[wayID] {
			t.Errorf("Didn't expect to find way %d", wayID)
		}
	}
}

func TestInterpolateShortestPathDistances(t *testing.T) {
	nodes := []osm.Node{
		{ID: 5384190463, Location: osm.LatLng{Lat: 51.5358664, Lng: -0.1272493}},
		{ID: 5384190445, Location: osm.LatLng{Lat: 51.5359780, Lng: -0.1271810}},
		{ID: 7788210688, Location: osm.LatLng{Lat: 51.5360033, Lng: -0.1271628}},
		{ID: 5384190494, Location: osm.LatLng{Lat: 51.5362126, Lng: -0.1270125}},
	}

	// Two ways, one ending with node 5384190494, one starting with it.
	ways := []osm.Way{
		{ID: 558345071, Nodes: []osm.NodeID{5384190463, 5384190445, 7788210688, 5384190494}},
	}

	w, err := ingest.BuildWorldFromOSM(nodes, ways, []osm.Relation{}, &ingest.BuildOptions{Cores: 2})
	if err != nil {
		t.Fatal(err)
	}
	path := w.FindFeatureByID(ingest.FromOSMWayID(558345071)).(b6.PhysicalFeature)

	cases := []struct {
		first         int
		last          int
		firstDistance float64
		lastDistance  float64
		expected      []float64
	}{
		{0, path.GeometryLen() - 1, 100.0, 200.0, []float64{100.0, 113.0, 116.0, 141.0}},
		{0, path.GeometryLen() - 1, 100.0, 50.0, []float64{91.0, 78.0, 75.0, 50.0}},
		{path.GeometryLen() - 1, 0, 200.0, 100.0, []float64{141.0, 116.0, 113.0, 100.0}},
		{0, path.GeometryLen() - 1, 100.0, math.Inf(1), []float64{100.0, 113.0, 116.0, 141.0}},
	}

	for _, c := range cases {
		segment := b6.Segment{Feature: path, First: c.first, Last: c.last}
		distances := interpolateShortestPathDistances(segment, b6.MetersToAngle(c.firstDistance), b6.MetersToAngle(c.lastDistance))

		if len(distances) == len(c.expected) {
			for i := range c.expected {
				if math.Abs(c.expected[i]-b6.AngleToMeters(distances[i])) > 1.0 {
					t.Errorf("Expected distance of %.0fm, found %0fm", c.expected[i], b6.AngleToMeters(distances[i]))
				}
			}
		} else {
			t.Errorf("Expected %d distances, found %d", len(c.expected), len(distances))
		}
	}
}

func TestAccessibility(t *testing.T) {
	camden := camden.BuildCamdenForTests(t)

	// Generate routes from the South end of Coal Drops Yard
	from := camden.FindFeatureByID(ingest.FromOSMNodeID(6083735356))
	if from == nil {
		t.Fatal("Failed to find from node")
	}
	distances, counts := ComputeAccessibility(from.FeatureID(), 500.0, SimpleWeights{}, camden)

	bridgeEnd := ingest.FromOSMNodeID(1540349979)
	expectedDistance := 210.0
	if math.Abs(distances[bridgeEnd]-expectedDistance) > 20.0 {
		t.Errorf("Expected distance of around %fm, found %fm", expectedDistance, distances[bridgeEnd])
	}

	bridge := ingest.FromOSMWayID(140633010)
	key := b6.SegmentKey{ID: bridge, First: 0, Last: 1}
	expectedCount := 132
	if counts[key] != expectedCount {
		t.Errorf("Expected count of around %d, found %d", expectedCount, counts[key])
	}

	footpath := ingest.FromOSMWayID(278159862)
	key = b6.SegmentKey{ID: footpath, First: 9, Last: 11}
	expectedCount = 1
	if counts[key] != expectedCount {
		t.Errorf("Expected count of around %d, found %d", expectedCount, counts[key])
	}

	foundNonIntersectionWithDistance := false
	for id := range distances {
		if n := len(b6.AllSegments(camden.Traverse(id))); n == 2 {
			foundNonIntersectionWithDistance = true
			break
		}
	}
	if !foundNonIntersectionWithDistance {
		t.Error("Expected non-intersection points to also have a distances")
	}
}

func TestBusWeights(t *testing.T) {
	camden := camden.BuildCamdenForTests(t)

	tests := []struct {
		id      osm.WayID
		useable bool
	}{
		{673447744, true},  // Midland road
		{681764413, false}, // Cycleway along Midland Road
	}

	weights := BusWeights{}
	for _, test := range tests {
		path := camden.FindFeatureByID(ingest.FromOSMWayID(test.id)).(b6.PhysicalFeature)
		if path == nil {
			t.Errorf("Failed to find way %d", test.id)
			continue
		}
		useable := weights.IsUseable(b6.ToSegment(path))
		if useable != test.useable {
			t.Errorf("Expected useable=%v for way %d, found %v", test.useable, test.id, useable)
		}
	}
}

func TestShortestPathFromConnectedBuildingWithNoEntrance(t *testing.T) {
	w := camden.BuildCamdenForTests(t)

	lighterman := b6.FindAreaByID(ingest.AreaIDFromOSMWayID(camden.LightermanWay), w)
	if lighterman == nil {
		t.Fatal("Expected to find The Lighterman")
	}

	entrances := 0
	for i := 0; i < lighterman.Len(); i++ {
		for _, path := range lighterman.Feature(i) {
			for _, r := range path.References() {
				if w.FindFeatureByID(r.Source()).Get("entrance").IsValid() {
					entrances++
				}
			}
		}
	}

	if entrances > 0 {
		t.Fatalf("Expected The Lightman to have no entrances, found %d", entrances)
	}

	weights := SimpleHighwayWeights{}
	search := NewShortestPathSearchFromBuilding(lighterman, weights, w)
	search.ExpandSearch(100.0, weights, Points, w)
	distances := search.PointDistances()

	if _, ok := distances[camden.StableStreetBridgeNorthEndID]; !ok {
		t.Error("Expected to find a route to Stable Street bridge")
	}
}

func TestShortestPathFromBuildingWithMoreThanOneEntrance(t *testing.T) {
	camden := camden.BuildCamdenForTests(t)

	stPancras := b6.FindAreaByID(ingest.AreaIDFromOSMWayID(4256246), camden)
	if stPancras == nil {
		t.Fatal("Expected to find St Pancras")
	}

	entrances := 0
	for i := 0; i < stPancras.Len(); i++ {
		for _, path := range stPancras.Feature(i) {
			for _, r := range path.References() {
				if camden.FindFeatureByID(r.Source()).Get("entrance").IsValid() {
					entrances++
				}
			}
		}
	}

	if entrances < 2 {
		t.Fatalf("Expected St Pancras to have many entrances, found %d", entrances)
	}

	weights := SimpleHighwayWeights{}
	search := NewShortestPathSearchFromBuilding(stPancras, weights, camden)
	search.ExpandSearch(500.0, weights, Points, camden)
	distances := search.PointDistances()

	// Check the distances to nodes on either side of the station - they should both be small, as the shortest
	// path should use the closest entrance.
	nodes := []osm.NodeID{
		6481824008, // Junction of Midland Road and Dangoor walk, accessible from entrance 1492770154 on the West side
		1237701825, // Junction of Pancras Road and the taxi rank, accessible from entrance 4360568915 on the East side
	}

	maxDistance := 30.0
	for _, node := range nodes {
		point := ingest.FromOSMNodeID(node)
		if distance, ok := distances[point]; !ok || distance > maxDistance {
			t.Errorf("Expected a short distance from St Pancras to Node %d, found %.2fm", node, distance)
		}
	}
}

func TestShortestPathReturnsBuildings(t *testing.T) {
	w := camden.BuildCamdenForTests(t)

	// Generate routes from the South end of Coal Drops Yard
	from := w.FindFeatureByID(ingest.FromOSMNodeID(6083735356))
	if from == nil {
		t.Fatal("Failed to find from node")
	}
	s := NewShortestPathSearchFromPoint(from.FeatureID())
	s.ExpandSearch(500.0, SimpleWeights{}, PointsAndAreas, w)

	expected := ingest.AreaIDFromOSMWayID(camden.CoalDropsYardWestBuildingWay)
	if _, ok := s.AreaDistances()[expected]; !ok {
		t.Errorf("Expected to find building when searching from a point")
	}

	theGranary := b6.FindAreaByID(ingest.AreaIDFromOSMWayID(222021576), w)
	if theGranary == nil {
		t.Error("Expected to find The Granary")
	}
	weights := SimpleHighwayWeights{}
	s = NewShortestPathSearchFromBuilding(theGranary, weights, w)
	s.ExpandSearch(500.0, weights, PointsAndAreas, w)

	expected = ingest.AreaIDFromOSMWayID(camden.LightermanWay)
	if _, ok := s.AreaDistances()[expected]; !ok {
		t.Errorf("Expected to find building when searching from a building")
	}
}

func TestElevationWeights(t *testing.T) {
	camden := camden.BuildCamdenForTests(t)

	camdenWithHill := ingest.NewMutableTagsOverlayWorld(camden)
	camdenWithHill.AddTag(ingest.FromOSMNodeID(4931754283).FeatureID(), b6.Tag{Key: "ele", Value: b6.String("100")})
	camdenWithHill.AddTag(ingest.FromOSMNodeID(6773349520).FeatureID(), b6.Tag{Key: "ele", Value: b6.String("200")})

	path := ComputeShortestPath(
		ingest.FromOSMNodeID(33000703),
		ingest.FromOSMNodeID(970237231),
		500.0, ElevationWeights{UpHillHard: true, w: camdenWithHill}, camdenWithHill)
	wayIDs := make(map[osm.WayID]bool)
	for _, segment := range path {
		wayIDs[osm.WayID(segment.Feature.FeatureID().Value)] = true
	}

	if !wayIDs[835618252] {
		t.Errorf("Expected to find way %d", 835618252) // Longer way.
	}

	if wayIDs[502802551] {
		t.Errorf("Didn't expect to find way %d", 502802551) // Shorter, but elevated way.
	}
}

func TestBuildRoute(t *testing.T) {
	camden := camden.BuildCamdenForTests(t)

	fromPath := camden.FindFeatureByID(ingest.FromOSMWayID(687471322))
	if fromPath == nil {
		t.Fatal("Failed to find way")
	}
	from := fromPath.Reference(0).Source()
	if !from.IsValid() {
		t.Fatal("Expected a point reference")
	}

	toPath := camden.FindFeatureByID(ingest.FromOSMWayID(367808662))
	if toPath == nil {
		t.Fatal("Failed to find way")
	}
	to := toPath.Reference(0).Source()
	if !to.IsValid() {
		t.Fatal("Expected a point reference")
	}

	s := NewShortestPathSearchFromPoint(from)
	s.ExpandSearchTo(to, 1000.0, WalkingTimeWeights{Speed: WalkingMetersPerSecond}, camden)
	route := s.BuildRoute(to)

	if steps := len(route.Steps); steps < 35 || steps > 45 {
		t.Errorf("Unexpected number of route steps %d", steps)
	}

	if cost := route.Steps[len(route.Steps)-1].Cost; cost < 850.0 || cost > 950.0 {
		t.Errorf("Unexpected route cost %f", cost)
	}
}
