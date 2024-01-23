package graph

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/osm"
	"diagonal.works/b6/test/camden"
)

func TestBuildStreetNetwork(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	highways := b6.FindPaths(b6.Keyed{"#highway"}, granarySquare)
	network := BuildStreetNetwork(highways, b6.MetersToAngle(100.0), SimpleHighwayWeights{}, nil, granarySquare)

	if len(network) < 100 {
		t.Errorf("Expected at least 100 paths on the network, found %d", len(network))
	}

	disconnected := ingest.FromOSMWayID(23211356)
	if _, ok := network[disconnected]; ok {
		t.Errorf("Didn't expect %s to be part of the street network", disconnected)
	}
}

func TestMergeInsertions(t *testing.T) {
	nodes := []osm.Node{
		{ID: 7555184307, Location: osm.LatLng{Lat: 51.5373281, Lng: -0.1252851}},
		{ID: 7555184302, Location: osm.LatLng{Lat: 51.5366646, Lng: -0.1255689}},
		{ID: 7787634260, Location: osm.LatLng{Lat: 51.5363904, Lng: -0.1256803}},
	}

	// Points are spaced 0m, 76m and 108m
	ways := []osm.Way{
		{
			ID:    807925586, // Stable Street
			Nodes: []osm.NodeID{7555184307, 7555184302, 7787634260},
		},
	}

	w, err := ingest.BuildWorldFromOSM(nodes, ways, []osm.Relation{}, &ingest.BuildOptions{Cores: 2})
	if err != nil {
		t.Fatal(err)
	}

	ids := []b6.PointID{
		accessID(ingest.AreaIDFromOSMWayID(222021577).FeatureID(), 0),
		accessID(ingest.AreaIDFromOSMWayID(222021573).FeatureID(), 0),
		accessID(ingest.AreaIDFromOSMWayID(532767917).FeatureID(), 0),
	}

	c := NewConnections()
	c.InsertPoint(ingest.FromOSMWayID(ways[0].ID), b6.MetersToAngle(10), ids[0])
	c.InsertPoint(ingest.FromOSMWayID(ways[0].ID), b6.MetersToAngle(85), ids[1])
	c.InsertPoint(ingest.FromOSMWayID(557519243), b6.MetersToAngle(10), ids[2]) // Not Stable Street

	applied := c.ApplyToPath(b6.FindPathByID(ingest.FromOSMWayID(ways[0].ID), w))
	if applied == nil {
		t.Fatal("Expected a path, found nil")
	}
	if applied.Len() != 5 {
		t.Fatalf("Expected 5 points, found %d", applied.Len())
	}
	expected := []b6.PointID{
		ingest.FromOSMNodeID(nodes[0].ID), ids[0], ingest.FromOSMNodeID(nodes[1].ID), ids[1], ingest.FromOSMNodeID(nodes[2].ID),
	}
	for i, e := range expected {
		if p, ok := applied.PointID(i); !ok || p != e {
			t.Errorf("Expected %s, found %s", e, p)
		}
	}
}

func TestClusterCloseInsertions(t *testing.T) {
	nodes := []osm.Node{
		{ID: 7555184307, Location: osm.LatLng{Lat: 51.5373281, Lng: -0.1252851}},
		{ID: 7555184302, Location: osm.LatLng{Lat: 51.5366646, Lng: -0.1255689}},
		{ID: 7787634260, Location: osm.LatLng{Lat: 51.5363904, Lng: -0.1256803}},
	}

	// Points are spaced 0m, 76m and 108m
	ways := []osm.Way{
		{
			ID:    807925586, // Stable Street
			Nodes: []osm.NodeID{7555184307, 7555184302, 7787634260},
		},
	}

	w, err := ingest.BuildWorldFromOSM(nodes, ways, []osm.Relation{}, &ingest.BuildOptions{Cores: 2})
	if err != nil {
		t.Fatal(err)
	}

	ids := []b6.PointID{
		accessID(ingest.AreaIDFromOSMWayID(222021577).FeatureID(), 0),
		accessID(ingest.AreaIDFromOSMWayID(222021573).FeatureID(), 0),
		accessID(ingest.AreaIDFromOSMWayID(532767917).FeatureID(), 0),
		accessID(ingest.AreaIDFromOSMWayID(222021572).FeatureID(), 0),
	}

	c := NewConnections()
	c.InsertPoint(ingest.FromOSMWayID(ways[0].ID), b6.MetersToAngle(10), ids[0])
	c.InsertPoint(ingest.FromOSMWayID(ways[0].ID), b6.MetersToAngle(13), ids[1])
	c.InsertPoint(ingest.FromOSMWayID(ways[0].ID), b6.MetersToAngle(85), ids[2])
	c.InsertPoint(ingest.FromOSMWayID(557519243), b6.MetersToAngle(10), ids[3]) // Not Stable Street
	c.Cluster(b6.MetersToAngle(4.0), w)

	applied := c.ApplyToPath(b6.FindPathByID(ingest.FromOSMWayID(ways[0].ID), w))
	if applied == nil {
		t.Fatal("Expected a path, found nil")
	}
	if applied.Len() != 5 {
		t.Fatalf("Expected 5 points, found %d", applied.Len())
	}
	// ids[1] should have been been clusters with ids[0]
	expected := []b6.PointID{
		ingest.FromOSMNodeID(nodes[0].ID), ids[0], ingest.FromOSMNodeID(nodes[1].ID), ids[2], ingest.FromOSMNodeID(nodes[2].ID),
	}
	for i, e := range expected {
		if p, ok := applied.PointID(i); !ok || p != e {
			t.Errorf("Expected %s, found %s", e, p)
		}
	}
}

func TestClusterInsertionsOntoExistingPoints(t *testing.T) {
	nodes := []osm.Node{
		{ID: 7555184307, Location: osm.LatLng{Lat: 51.5373281, Lng: -0.1252851}},
		{ID: 7555184302, Location: osm.LatLng{Lat: 51.5366646, Lng: -0.1255689}},
		{ID: 7787634260, Location: osm.LatLng{Lat: 51.5363904, Lng: -0.1256803}},

		{ID: 8742459182, Location: osm.LatLng{Lat: 51.5349113, Lng: -0.1265422}},
		{ID: 87470440, Location: osm.LatLng{Lat: 51.5348674, Lng: -0.1264565}},
		{ID: 2512646902, Location: osm.LatLng{Lat: 51.5348441, Lng: -0.1263883}},

		{ID: 1715968780, Location: osm.LatLng{Lat: 51.5371597, Lng: -0.1235755}},
		{ID: 1715968788, Location: osm.LatLng{Lat: 51.5374419, Lng: -0.1252412}},

		// An arbitrary entrance node to test point clustering
		{ID: 3790640854, Location: osm.LatLng{Lat: 51.5355958, Lng: -0.1250639}},
	}

	ways := []osm.Way{
		{
			// Points are spaced 0m, 7m and 13m
			ID:    642639441, // Regent's Canal Towpath
			Nodes: []osm.NodeID{8742459182, 87470440, 2512646902},
		},
		{
			ID: 807925586, // Stable Street
			// Points are spaced 0m, 76m and 108m
			Nodes: []osm.NodeID{7555184307, 7555184302, 7787634260},
		},
		{
			ID:    557519243, // Handyside Street
			Nodes: []osm.NodeID{1715968780, 1715968788},
		},
	}

	w, err := ingest.BuildWorldFromOSM(nodes, ways, []osm.Relation{}, &ingest.BuildOptions{Cores: 2})
	if err != nil {
		t.Fatal(err)
	}

	ids := []b6.PointID{
		// Near the towpath
		accessID(ingest.AreaIDFromOSMWayID(834276095).FeatureID(), 0),

		// Near Stable Street
		accessID(ingest.AreaIDFromOSMWayID(222021577).FeatureID(), 0),
		accessID(ingest.AreaIDFromOSMWayID(222021573).FeatureID(), 0),
		accessID(ingest.AreaIDFromOSMWayID(532767917).FeatureID(), 0),
		accessID(ingest.AreaIDFromOSMWayID(222021572).FeatureID(), 0),
	}

	c := NewConnections()
	c.InsertPoint(ingest.FromOSMWayID(ways[0].ID), b6.MetersToAngle(4), ids[0])
	// The following two will be clustered together, then clustered onto a point
	// in Stable Street
	c.InsertPoint(ingest.FromOSMWayID(ways[1].ID), b6.MetersToAngle(78), ids[1])
	c.InsertPoint(ingest.FromOSMWayID(ways[1].ID), b6.MetersToAngle(80), ids[2])
	c.InsertPoint(ingest.FromOSMWayID(ways[1].ID), b6.MetersToAngle(85), ids[3])
	c.InsertPoint(ingest.FromOSMWayID(ways[2].ID), b6.MetersToAngle(10), ids[4])

	c.AddPath(ids[2], ingest.FromOSMNodeID(nodes[len(nodes)-1].ID))
	c.Cluster(b6.MetersToAngle(4.0), w)

	applied := c.ApplyToPath(b6.FindPathByID(ingest.FromOSMWayID(ways[1].ID), w))
	if applied == nil {
		t.Fatal("Expected a path, found nil")
	}
	if applied.Len() != 4 {
		t.Fatalf("Expected 4 points, found %d", applied.Len())
	}
	// ids[1] and ids[2] should have been been clustered onto point[1]
	expected := []b6.PointID{
		ingest.FromOSMNodeID(nodes[0].ID), ingest.FromOSMNodeID(nodes[1].ID), ids[3], ingest.FromOSMNodeID(nodes[2].ID),
	}
	for i, e := range expected {
		if p, ok := applied.PointID(i); !ok || p != e {
			t.Errorf("Expected %s, found %s", e, p)
		}
	}

	source := c.ModifyWorld(w)
	connected, err := ingest.NewWorldFromSource(source, &ingest.BuildOptions{Cores: 2})
	if err != nil {
		t.Fatal(err)
	}

	// The access path from ids[2] should start from point[1] of ways[1], as
	// ids[1] and ids[2] have been clustered onto it
	paths := b6.AllPaths(connected.FindPathsByPoint(ingest.FromOSMNodeID(nodes[len(nodes)-1].ID)))
	if len(paths) != 1 {
		t.Fatalf("Expected 1 path, found %d", len(paths))
	}

	access := paths[0]
	if access.Len() != 2 {
		t.Fatalf("Expected 2 points, found %d", access.Len())
	}
	expected = []b6.PointID{
		ingest.FromOSMNodeID(ways[1].Nodes[1]),
		ingest.FromOSMNodeID(nodes[len(nodes)-1].ID),
	}
	for i, e := range expected {
		if p := access.Feature(i); p.PointID() != e {
			t.Errorf("Expected %s, found %s", e, p)
		}
	}
}

func countAccessibleBuildings(from b6.PointID, maxDistance float64, w b6.World) int {
	distances, _ := ComputeAccessibility(from, maxDistance, SimpleHighwayWeights{}, w)
	buildings := make(map[b6.AreaID]struct{})
	for point := range distances {
		areas := w.FindAreasByPoint(point)
		for areas.Next() {
			area := areas.Feature()
			if area.Get("#building").IsValid() {
				buildings[area.AreaID()] = struct{}{}
			}
		}
	}
	return len(buildings)
}

func countAccessibleAmenities(from b6.PointID, maxDistance float64, w b6.World) int {
	distances, _ := ComputeAccessibility(from, maxDistance, SimpleHighwayWeights{}, w)
	amenities := make(map[b6.PointID]struct{})
	for id := range distances {
		point := b6.FindPointByID(id, w)
		if point.Get("#amenity").IsValid() {
			amenities[id] = struct{}{}
		}
	}
	return len(amenities)
}

func TestConnectGranarySquare(t *testing.T) {
	tests := []struct {
		name string
		f    func(features b6.Features, network PathIDSet, w b6.World, t *testing.T) b6.World
	}{
		{"ConnectInsertingNewPoints", ValidateConnectInsertingNewPoints},
		{"ConnectUsingExistingPoints", ValidateConnectUsingExistingPoints},
	}

	granarySquare := camden.BuildGranarySquareForTests(t)

	highways := b6.FindPaths(b6.Keyed{Key: "#highway"}, granarySquare)
	weights := SimpleHighwayWeights{}
	network := BuildStreetNetwork(highways, b6.MetersToAngle(100), weights, nil, granarySquare)

	for _, test := range tests {
		features := granarySquare.FindFeatures(b6.Union{b6.Keyed{Key: "#building"}, b6.Keyed{Key: "#amenity"}})
		t.Run(test.name, func(t *testing.T) {
			connected := test.f(features, network, granarySquare, t)
			if connected != nil {
				origin := b6.FindPointByID(ingest.FromOSMNodeID(6083735356), granarySquare) // South end of the Coal Drops Yard footway
				before := countAccessibleBuildings(origin.PointID(), 1000.0, granarySquare)
				after := countAccessibleBuildings(origin.PointID(), 1000.0, connected)
				if after <= before {
					t.Errorf("Expected more buildings to be connected, before: %d after: %d", before, after)
				}

				before = countAccessibleAmenities(origin.PointID(), 1000.0, granarySquare)
				after = countAccessibleAmenities(origin.PointID(), 1000.0, connected)
				if after <= before {
					t.Errorf("Expected more amenities to be connected, before: %d after: %d", before, after)
				}
			}
		})
	}
}

func ValidateConnectInsertingNewPoints(features b6.Features, network PathIDSet, w b6.World, t *testing.T) b6.World {
	s := InsertNewPointsIntoPaths{
		Connections:      NewConnections(),
		World:            w,
		ClusterThreshold: b6.MetersToAngle(4.0),
	}
	for features.Next() {
		ConnectFeature(features.Feature(), network, b6.MetersToAngle(100), w, s)
	}
	s.Finish()
	source := s.Output()

	connected, err := ingest.NewWorldFromSource(source, &ingest.BuildOptions{Cores: 2})
	if err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}
	return connected
}

func ValidateConnectUsingExistingPoints(features b6.Features, network PathIDSet, w b6.World, t *testing.T) b6.World {
	s := UseExisitingPoints{
		Connections: NewConnections(),
	}
	for features.Next() {
		ConnectFeature(features.Feature(), network, b6.MetersToAngle(100), w, s)
	}
	s.Finish()
	connected := ingest.NewMutableOverlayWorld(w)
	if err := connected.MergeSource(s.Output()); err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}
	return connected
}
