package functions

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"
)

func TestFindReachablePoints(t *testing.T) {
	w := camden.BuildCamdenForTests(t)
	if w == nil {
		return
	}
	m := ingest.NewMutableOverlayWorld(w)

	origin := b6.FindPointByID(camden.StableStreetBridgeSouthEndID, m)
	if origin == nil {
		t.Errorf("Failed to find origin")
	}

	context := api.Context{
		World: m,
	}

	query := b6.Tagged{Key: "#barrier", Value: "gate"}
	collection, err := ReachablePoints(origin, "walk", 1000.0, query, &context)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	barriers := make(map[b6.PointID]bool)
	i := collection.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			t.Errorf("Expected no error, found: %s", err)
		}
		if !ok {
			break
		}
		barriers[i.Value().(b6.PointFeature).PointID()] = true
	}

	if _, ok := barriers[camden.SomersTownBridgeEastGateID]; !ok {
		t.Errorf("Expected to find %s", camden.SomersTownBridgeEastGateID)
	}
}

func TestFindReachableFeatures(t *testing.T) {
	w := camden.BuildCamdenForTests(t)
	if w == nil {
		return
	}
	m := ingest.NewMutableOverlayWorld(w)

	origin := b6.FindPointByID(camden.StableStreetBridgeSouthEndID, m)
	if origin == nil {
		t.Errorf("Failed to find origin")
	}

	context := api.Context{
		World: m,
	}
	collection, err := ReachableFeatures(origin, "walk", 1000.0, b6.Keyed{"#amenity"}, &context)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	amenities := make(map[b6.FeatureID]bool)
	i := collection.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			t.Errorf("Expected no error, found: %s", err)
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
	if w == nil {
		return
	}
	m := ingest.NewMutableOverlayWorld(w)

	origin := b6.FindPointByID(camden.StableStreetBridgeSouthEndID, m)
	if origin == nil {
		t.Errorf("Failed to find origin")
	}

	context := api.Context{
		World: m,
	}
	collection, err := PathsToReachFeatures(origin, "walk", 1000.0, b6.Keyed{"#amenity"}, &context)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	paths := make(map[b6.FeatureID]int)
	if err := api.FillMap(collection, paths); err != nil {
		t.Errorf("Expected no error, found %s", err)
		return
	}

	if len(paths) < 60 {
		t.Errorf("Expected counts for more than 60 paths, found %d", len(paths))
	}
	if count := paths[ingest.FromOSMWayID(camden.StableStreetBridgeWay).FeatureID()]; count < 2 {
		t.Errorf("Expected more than 2 routes to use bridge, found %d", count)
	}
}
