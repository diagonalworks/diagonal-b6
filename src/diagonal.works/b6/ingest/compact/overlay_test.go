package compact

import (
	"context"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/osm"
	"diagonal.works/b6/test"
	"diagonal.works/b6/test/camden"
)

func TestOverlayPathOnExistingWorld(t *testing.T) {
	nodes, ways, relations, err := osm.ReadWholePBF(test.Data(test.GranarySquarePBF))
	if err != nil {
		t.Fatalf("Failed to build granary square: %s", err)
	}

	osmSource := ingest.MemoryOSMSource{Nodes: nodes, Ways: ways, Relations: relations}
	source, err := ingest.NewFeatureSourceFromPBF(&osmSource, &ingest.BuildOptions{Cores: 2}, context.Background())
	options := Options{Goroutines: 2, PointsScratchOutputType: OutputTypeMemory}
	index, err := BuildInMemory(source, &options)
	if err != nil {
		t.Fatalf("Failed to build base index: %s", err)
	}

	path := &ingest.GenericFeature{}
	path.SetFeatureID(b6.FeatureID{b6.FeatureTypePath, b6.NamespaceDiagonalAccessPoints, 42})
	path.AddTag(b6.Tag{Key: "#highway", Value: b6.StringExpression("cycleway")})
	path.ModifyOrAddTag(b6.Tag{b6.PathTag, b6.Values([]b6.Value{ingest.FromOSMNodeID(camden.LightermanEntranceNode), ingest.FromOSMNodeID(camden.StableStreetBridgeNorthEndNode)})})

	w, err := NewWorldFromData(index)
	if err != nil {
		t.Fatalf("Failed to create base world: %s", err)
	}
	source = ingest.MemoryFeatureSource([]ingest.Feature{path})
	overlay, err := BuildOverlayInMemory(source, &options, w)
	if err != nil {
		t.Fatalf("Failed to build overlay index: %s", err)
	}
	if err := w.Merge(overlay); err != nil {
		t.Fatalf("Failed to merge overlay index: %s", err)
	}

	highways := w.FindFeatures(b6.Keyed{"#highway"})
	found := false
	for highways.Next() {
		if highways.FeatureID() == path.FeatureID() {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Expected to find overlaid path via FindFeatures")
	}

	ps := w.FindReferences(ingest.FromOSMNodeID(camden.LightermanEntranceNode), b6.FeatureTypePath)
	found = false
	for ps.Next() {
		if ps.FeatureID() == path.FeatureID() {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find overlaid path via FindPathsByPoint")
	}
}
