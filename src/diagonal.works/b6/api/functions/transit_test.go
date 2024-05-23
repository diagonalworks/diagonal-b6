package functions

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"
)

func TestFindReachableFeatures(t *testing.T) {
	w := camden.BuildCamdenForTests(t)
	m := ingest.NewMutableOverlayWorld(w)

	origin := m.FindFeatureByID(camden.StableStreetBridgeSouthEndID)
	if origin == nil {
		t.Fatal("Failed to find origin")
	}

	context := api.Context{
		World: m,
	}
	collection, err := reachable(&context, origin, b6.ArrayValuesCollection[b6.Tag]([]b6.Tag{{Key: "mode", Value: b6.String("walk")}}).Collection(), 1000.0, b6.Keyed{Key: "#amenity"})
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	amenities := make(map[b6.FeatureID]bool)
	i := collection.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			t.Fatalf("Expected no error, found: %s", err)
		}
		if !ok {
			break
		}
		amenities[i.Value().(b6.Feature).FeatureID()] = true
	}

	if _, ok := amenities[camden.LightermanID.FeatureID()]; !ok {
		t.Errorf("Expected to find %s", camden.LightermanID)
	}
}

func TestPathsToReachFeatures(t *testing.T) {
	w := camden.BuildCamdenForTests(t)
	m := ingest.NewMutableOverlayWorld(w)

	origin := m.FindFeatureByID(camden.StableStreetBridgeSouthEndID)
	if origin == nil {
		t.Errorf("Failed to find origin")
	}

	context := api.Context{
		World: m,
	}
	collection, err := pathsToReachFeatures(&context, origin, b6.ArrayValuesCollection[b6.Tag]([]b6.Tag{{Key: "mode", Value: b6.String("walk")}}).Collection(), 1000.0, b6.Keyed{"#amenity"})
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	paths := make(map[b6.FeatureID]int)
	if err := api.FillMap(collection, paths); err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

	if len(paths) < 60 {
		t.Errorf("Expected counts for more than 60 paths, found %d", len(paths))
	}
	if count := paths[ingest.FromOSMWayID(camden.StableStreetBridgeWay).FeatureID()]; count < 2 {
		t.Errorf("Expected more than 2 routes to use bridge, found %d", count)
	}
}
