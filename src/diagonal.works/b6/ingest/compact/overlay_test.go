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
		t.Errorf("Failed to build granary square: %s", err)
	}

	osmSource := ingest.MemoryOSMSource{Nodes: nodes, Ways: ways, Relations: relations}
	o := ingest.BuildOptions{Cores: 2}
	source, err := ingest.NewFeatureSourceFromPBF(&osmSource, &o, context.Background())
	options := Options{Cores: 2, PointsWorkOutputType: OutputTypeMemory}
	index, err := BuildInMemory(source, &options)
	if err != nil {
		t.Errorf("Failed to build base index: %s", err)
		return
	}

	path := ingest.NewPathFeature(2)
	path.PathID = b6.MakePathID(b6.NamespaceDiagonalAccessPoints, 42)
	path.Tags.AddTag(b6.Tag{Key: "#highway", Value: "cycleway"})
	path.SetPointID(0, ingest.FromOSMNodeID(camden.LightermanEntranceNode))
	path.SetPointID(1, ingest.FromOSMNodeID(camden.StableStreetBridgeNorthEndNode))

	w, err := NewWorldFromData(index)
	if err != nil {
		t.Errorf("Failed to create base world: %s", err)
	}
	source = ingest.MemoryFeatureSource([]ingest.Feature{path})
	overlay, err := BuildOverlayInMemory(source, &options, w)
	if err != nil {
		t.Errorf("Failed to build overlay index: %s", err)
		return
	}
	if err := w.Merge(overlay); err != nil {
		t.Errorf("Failed to merge overlay index: %s", err)
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
		t.Errorf("Expected to find overlaid path via FindFeatures")
	}

	ps := w.FindPathsByPoint(ingest.FromOSMNodeID(camden.LightermanEntranceNode))
	found = false
	for ps.Next() {
		if ps.PathSegment().FeatureID() == path.FeatureID() {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find overlaid path via FindPathsByPoint")
	}
}
