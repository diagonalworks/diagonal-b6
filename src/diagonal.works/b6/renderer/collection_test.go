package renderer

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"
	"github.com/golang/geo/s2"
)

func TestCollectionWithBoundaries(t *testing.T) {

	w := ingest.NewMutableOverlayWorld(camden.BuildGranarySquareForTests(t))
	collection := ingest.CollectionFeature{
		CollectionID: b6.MakeCollectionID(b6.NamespacePrivate, 1),
		Keys:         []interface{}{0, 1},
		Values: []interface{}{
			camden.LightermanID,
			camden.GranarySquareID,
		},
	}
	if err := w.AddFeature(&collection); err != nil {
		t.Fatalf("failed to add collection")
	}

	worlds := &ingest.MutableWorlds{Base: w}
	r := NewCollectionRenderer(BasemapRenderRules, worlds)
	args := &TileArgs{Q: collection.FeatureID().String()}
	projection := b6.NewTileMercatorProjection(16)
	tile, err := r.Render(projection.TileFromLatLng(s2.LatLngFromDegrees(51.535268, -0.124603)), args)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	if l := len(tile.Layers); l != 1 {
		t.Fatalf("Expected one layer, found %d", l)
	}
	if l := len(tile.Layers[0].Features); l != 2 {
		t.Fatalf("Expected two features, found %d", l)
	}
}

func TestCollectionWithFeatureKeysAndValues(t *testing.T) {
	w := ingest.NewMutableOverlayWorld(camden.BuildGranarySquareForTests(t))
	collection := ingest.CollectionFeature{
		CollectionID: b6.MakeCollectionID(b6.NamespacePrivate, 1),
		Keys: []interface{}{
			camden.StableStreetBridgeNorthEndID,
			camden.StableStreetBridgeNorthEndID,
		},
		Values: []interface{}{
			camden.LightermanID,
			camden.GranarySquareID,
		},
	}
	if err := w.AddFeature(&collection); err != nil {
		t.Fatalf("failed to add collection")
	}

	worlds := &ingest.MutableWorlds{Base: w}
	r := NewCollectionRenderer(BasemapRenderRules, worlds)
	args := &TileArgs{Q: collection.FeatureID().String()}
	projection := b6.NewTileMercatorProjection(16)
	tile, err := r.Render(projection.TileFromLatLng(s2.LatLngFromDegrees(51.535268, -0.124603)), args)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	if l := len(tile.Layers); l != 1 {
		t.Fatalf("Expected one layer, found %d", l)
	}
	if l := len(tile.Layers[0].Features); l != 3 {
		t.Fatalf("Expected three features, found %d", l)
	}

}
