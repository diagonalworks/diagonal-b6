package encoding

import (
	"os"
	"testing"

	"diagonal.works/b6/osm"
)

func loadGranarySquareForTests(t *testing.T) (map[osm.NodeID]string, map[osm.NodeID]string) {
	t.Helper()
	names := make(map[osm.NodeID]string)
	amenities := make(map[osm.NodeID]string)
	var maxNodeID osm.NodeID
	emit := func(element osm.Element) error {
		if node, ok := element.(*osm.Node); ok {
			if node.ID > maxNodeID {
				maxNodeID = node.ID
			}
			if name, ok := node.Tag("name"); ok {
				names[node.ID] = name
			}
			if amenity, ok := node.Tag("amenity"); ok {
				amenities[node.ID] = amenity
			}
		}
		return nil
	}
	input, err := os.Open("../../../../data/tests/granary-square.osm.pbf")
	if err != nil {
		t.Fatalf("Failed to open test data: %s", err)
	}
	defer input.Close()
	if err = osm.ReadPBF(input, emit); err != nil {
		t.Fatalf("Failed to read test data: %s", err)
	}
	return names, amenities
}
