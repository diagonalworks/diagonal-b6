package transit

import (
	"testing"

	"github.com/golang/geo/s2"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/osm"
	"diagonal.works/b6/test/camden"
)

func TestWayHeadsTowardsNextStop(t *testing.T) {
	nodes := []osm.Node{
		osm.Node{ID: 1, Location: osm.LatLng{Lat: 0.0, Lng: 0.0}},
		osm.Node{ID: 2, Location: osm.LatLng{Lat: 1.0, Lng: 0.0}},
	}

	ways := []osm.Way{
		osm.Way{
			ID:    1,
			Nodes: []osm.NodeID{1, 2},
		},
		osm.Way{
			ID:    2,
			Nodes: []osm.NodeID{2, 1},
		},
		osm.Way{
			ID:    3,
			Nodes: []osm.NodeID{1, 2},
			Tags:  []osm.Tag{{Key: "oneway", Value: "yes"}},
		},
		osm.Way{
			ID:    4,
			Nodes: []osm.NodeID{2, 1},
			Tags:  []osm.Tag{{Key: "oneway", Value: "yes"}},
		},
	}

	tests := []struct {
		id       osm.WayID
		expected bool
	}{
		{1, true},
		{2, true},
		{3, true},
		{4, false},
	}

	w, err := ingest.BuildWorldFromOSM(nodes, ways, []osm.Relation{}, &ingest.BuildOptions{FailInvalidFeatures: true})
	if err != nil {
		t.Error(err)
		return
	}
	point := s2.PointFromLatLng(s2.LatLngFromDegrees(2.0, 0.0))
	for _, test := range tests {
		path := w.FindFeatureByID(ingest.FromOSMWayID(test.id))
		if p, ok := path.(b6.PhysicalFeature); ok {
			if actual := isPathHeadingTowardsPoint(p, point); actual != test.expected {
				t.Errorf("Expected %v for way %d, found %v", test.expected, test.id, actual)
			}
		} else {
			t.Fatalf("Expected a NestedPhysicalFeature")
		}
	}
}

func TestLookNaptanStreet(t *testing.T) {
	// Royal college street on 214, OSM node 469780522, naptan 490011756E,
	// lat,lng 51.5354776,-0.1338760
	camden := camden.BuildCamdenForTests(t)
	if camden == nil {
		return
	}
	stop := Stop{
		ID:       StopID("490011756E"),
		Location: s2.LatLngFromDegrees(51.5354776, -0.1338760),
		AlternateIDs: []AlternateID{
			{Namespace: NaptanAtcoNamespace, ID: "490011756E"},
		},
	}

	correctWay, incorrectWay := osm.WayID(39025893), osm.WayID(8386760)
	path := camden.FindFeatureByID(ingest.FromOSMWayID(correctWay)).(b6.PhysicalFeature)
	if path == nil {
		t.Errorf("Failed to find way %d: test data error", correctWay)
		return
	}

	matched, ok := isPathNamedTheSameAsNaptanStreet(path, &stop, camden)
	if !ok || !matched {
		t.Errorf("Expected way %d to match stop", correctWay)
	}

	path = camden.FindFeatureByID(ingest.FromOSMWayID(incorrectWay)).(b6.PhysicalFeature)
	if path == nil {
		t.Errorf("Failed to find way %d: test data error", incorrectWay)
		return
	}

	matched, ok = isPathNamedTheSameAsNaptanStreet(path, &stop, camden)
	if !ok || matched {
		t.Errorf("Expected way %d to not match stop", incorrectWay)
	}
}
