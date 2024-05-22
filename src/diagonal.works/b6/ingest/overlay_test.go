package ingest

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/osm"
	"github.com/golang/geo/s2"
)

func TestOverlayWorldReturnsPathsFromAllIndices(t *testing.T) {
	nodes := []osm.Node{
		{ID: 5378333625, Location: osm.LatLng{Lat: 51.5352195, Lng: -0.1254286}},
		{ID: 1715968739, Location: osm.LatLng{Lat: 51.5351398, Lng: -0.1249654}},
		{ID: 1715968738, Location: osm.LatLng{Lat: 51.5351015, Lng: -0.1248611}},
		{ID: 4966136648, Location: osm.LatLng{Lat: 51.5348874, Lng: -0.1260855}},
		{ID: 5378333638, Location: osm.LatLng{Lat: 51.5367686, Lng: -0.1282862}},
		{ID: 7555184307, Location: osm.LatLng{Lat: 51.5373281, Lng: -0.1252851}},
		{ID: 1715968755, Location: osm.LatLng{Lat: 51.5354037, Lng: -0.1260829}},
		{ID: 1447052073, Location: osm.LatLng{Lat: 51.5350326, Lng: -0.1247915}},
		{ID: 1540349979, Location: osm.LatLng{Lat: 51.5348204, Lng: -0.1246405}},
	}

	ways1 := []osm.Way{
		{ID: 642639444, Nodes: []osm.NodeID{5378333625, 1715968739, 1715968738}},
		{ID: 557698825, Nodes: []osm.NodeID{5378333625, 4966136648, 5378333638}},
	}

	ways2 := []osm.Way{
		{ID: 807925586, Nodes: []osm.NodeID{7555184307, 1715968755, 5378333625}},
		{ID: 140633010, Nodes: []osm.NodeID{1447052073, 1540349979}},
	}

	ways := [][]osm.Way{ways1, ways2}
	worlds := make([]b6.World, len(ways))
	for i, w := range ways {
		var err error
		if worlds[i], err = BuildWorldFromOSM(nodes, w, []osm.Relation{}, &BuildOptions{Cores: 2}); err != nil {
			t.Fatalf("Failed to build world: %s", err)
		}
	}
	overlay := NewOverlayWorld(worlds[0], worlds[1])

	cap := s2.CapFromCenterAngle(nodes[0].Location.ToS2Point(), b6.MetersToAngle(500))
	found := b6.AllFeatures(overlay.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.NewIntersectsCap(cap)}))

	expected := []osm.WayID{140633010, 557698825, 642639444, 807925586}
	if len(found) != len(expected) {
		t.Fatalf("Expected length %d, found %d", len(expected), len(found))
	}

	for i, id := range expected {
		if found[i].FeatureID().Value != uint64(id) {
			t.Errorf("Expected ID %d, found %s at index %d", id, found[i].FeatureID(), i)
		}
	}
}

func TestOverlayWorldReplacesPathsFromOneIndexWithAnother(t *testing.T) {
	nodes := []osm.Node{
		{ID: 5378333625, Location: osm.LatLng{Lat: 51.5352195, Lng: -0.1254286}},
		{ID: 1715968739, Location: osm.LatLng{Lat: 51.5351398, Lng: -0.1249654}},
		{ID: 1715968738, Location: osm.LatLng{Lat: 51.5351015, Lng: -0.1248611}},
		{ID: 4966136648, Location: osm.LatLng{Lat: 51.5348874, Lng: -0.1260855}},
		{ID: 5378333638, Location: osm.LatLng{Lat: 51.5367686, Lng: -0.1282862}},
		{ID: 7555184307, Location: osm.LatLng{Lat: 51.5373281, Lng: -0.1252851}},
		{ID: 1715968755, Location: osm.LatLng{Lat: 51.5354037, Lng: -0.1260829}},
		{ID: 1447052073, Location: osm.LatLng{Lat: 51.5350326, Lng: -0.1247915}},
		{ID: 1540349979, Location: osm.LatLng{Lat: 51.5348204, Lng: -0.1246405}},
	}

	ways1 := []osm.Way{
		{
			ID:    642639444,
			Nodes: []osm.NodeID{5378333625, 1715968739, 1715968738},
			Tags:  []osm.Tag{{Key: "highway", Value: "path"}},
		},
		{ID: 557698825, Nodes: []osm.NodeID{5378333625, 4966136648, 5378333638}},
	}

	ways2 := []osm.Way{
		{
			ID:    642639444,
			Nodes: []osm.NodeID{5378333625, 1715968738},
			Tags:  []osm.Tag{{Key: "highway", Value: "cycleway"}},
		},
		{ID: 557698825, Nodes: []osm.NodeID{5378333625, 5378333638}},
	}

	ways := [][]osm.Way{ways1, ways2}
	worlds := make([]b6.World, len(ways))
	for i, w := range ways {
		var err error
		if worlds[i], err = BuildWorldFromOSM(nodes, w, []osm.Relation{}, &BuildOptions{Cores: 2}); err != nil {
			t.Fatalf("Failed to build world: %s", err)
		}
	}
	overlay := NewOverlayWorld(worlds[1], worlds[0])

	paths := b6.AllFeatures(overlay.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.Tagged{Key: "#highway", Value: b6.StringExpression("path")}}))
	if len(paths) > 0 {
		t.Errorf("Expected to find 0 paths, found %d", len(paths))
	}

	paths = b6.AllFeatures(overlay.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.Tagged{Key: "#highway", Value: b6.StringExpression("cycleway")}}))
	if len(paths) == 1 {
		expectedValue := "cycleway"
		if highway := paths[0].Get("#highway"); highway.Value.String() != expectedValue {
			t.Errorf("Expected to find highway tag value %q, found %q", expectedValue, highway.Value)
		}
		expectedLength := 2
		if len := paths[0].(b6.PhysicalFeature).GeometryLen(); len != expectedLength {
			t.Errorf("Expected to find %d points, found %d", expectedLength, len)
		}
	} else {
		t.Errorf("Expected to find 1 path, found %d", len(paths))
	}
}
