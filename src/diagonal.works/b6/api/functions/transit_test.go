package functions

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/testcamden"
)

func TestFindReachablePoints(t *testing.T) {
	w := testcamden.BuildCamden(t)
	m := ingest.NewMutableOverlayWorld(w)
	origin := b6.FindPointByID(testcamden.StableStreetBridgeSouthEndID, m)
	if origin == nil {
		t.Fatal("Failed to find origin")
	}

	context := api.Context{
		World: m,
	}

	query := b6.Tagged{Key: "#barrier", Value: "gate"}
	collection, err := reachablePoints(&context, origin, "walk", 1000.0, query)
	if err != nil {
		t.Fatalf("reachablePoints() failed with: %v", err)
	}

	barriers := make(map[b6.PointID]bool)
	i := collection.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			t.Fatalf("i.Next() failed with: %v", err)
		}
		if !ok {
			break
		}
		barriers[i.Value().(b6.PointFeature).PointID()] = true
	}

	if _, ok := barriers[testcamden.SomersTownBridgeEastGateID]; !ok {
		t.Errorf("Expected to find %s in barriers, but got: %v", testcamden.SomersTownBridgeEastGateID, barriers)
	}
}

func TestFindReachableFeatures(t *testing.T) {
	w := testcamden.BuildCamden(t)
	m := ingest.NewMutableOverlayWorld(w)
	origin := b6.FindPointByID(testcamden.StableStreetBridgeSouthEndID, m)
	if origin == nil {
		t.Fatal("Failed to find origin")
	}

	context := api.Context{
		World: m,
	}
	collection, err := reachableFeatures(&context, origin, "walk", 1000.0, b6.Keyed{Key: "#amenity"})
	if err != nil {
		t.Fatalf("reachableFeatures() failed with: %v", err)
	}

	amenities := make(map[b6.FeatureID]bool)
	i := collection.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			t.Fatalf("i.Next() failed with: %v", err)
		}
		if !ok {
			break
		}
		amenities[i.Value().(b6.Feature).FeatureID()] = true
	}

	if _, ok := amenities[testcamden.LightermanID.FeatureID()]; !ok {
		t.Errorf("Expected to find %s in amenities, but got: %v", testcamden.LightermanID, amenities)
	}
}

func TestPathsToReachFeatures(t *testing.T) {
	w := testcamden.BuildCamden(t)
	m := ingest.NewMutableOverlayWorld(w)
	origin := b6.FindPointByID(testcamden.StableStreetBridgeSouthEndID, m)
	if origin == nil {
		t.Fatal("Failed to find origin")
	}

	context := api.Context{
		World: m,
	}
	collection, err := pathsToReachFeatures(&context, origin, "walk", 1000.0, b6.Keyed{Key: "#amenity"})
	if err != nil {
		t.Fatalf("pathsToReachFeatures() failed with: %v", err)
	}

	paths := make(map[b6.FeatureID]int)
	if err := api.FillMap(collection, paths); err != nil {
		t.Fatalf("api.FillMap() failed with: %v", err)
	}

	if len(paths) < 60 {
		t.Errorf("Expected counts for more than 60 paths, found %d", len(paths))
	}
	if count := paths[ingest.FromOSMWayID(testcamden.StableStreetBridgeWay).FeatureID()]; count < 2 {
		t.Errorf("Expected more than 2 routes to use bridge, found %d", count)
	}
}
