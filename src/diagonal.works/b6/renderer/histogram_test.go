package renderer

import (
	"log"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"
	"github.com/golang/geo/s2"
)

func TestHistogramWithBucketedFeature(t *testing.T) {
	worlds := &ingest.MutableWorlds{
		Base: camden.BuildGranarySquareForTests(t),
	}

	worldID := b6.CollectionID{Namespace: "diagonal.works/test/world", Value: 0}
	histogramID := b6.CollectionID{Namespace: "diagonal.works/test/histogram", Value: 0}

	c := b6.ArrayCollection[b6.FeatureID, b6.StringExpression]{
		Keys:   []b6.FeatureID{camden.LightermanID.FeatureID()},
		Values: []b6.StringExpression{"pub"},
	}
	histogram, err := api.NewHistogramFromCollection(c.Collection(), histogramID)
	if err != nil {
		log.Fatalf("Expected no error when creating histogram, found: %s", err)
	}
	w := worlds.FindOrCreateWorld(worldID.FeatureID())
	if err := w.AddFeature(histogram); err != nil {
		t.Fatalf("Failed to add collection: %s", err)
	}

	r := NewHistogramRenderer(BasemapRenderRules, worlds)
	args := &TileArgs{Q: histogramID.String(), R: worldID.FeatureID()}
	projection := b6.NewTileMercatorProjection(16)
	tile, err := r.Render(projection.TileFromLatLng(s2.LatLngFromDegrees(51.53531, -0.12434)), args)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	if l := len(tile.Layers); l != 1 {
		t.Fatalf("Expected one layer, found %d", l)
	}
	if l := len(tile.Layers[0].Features); l != 1 {
		t.Fatalf("Expected one feature, found %d", l)
	}
	if b, ok := tile.Layers[0].Features[0].Tags["bucket"]; !ok || b != "0" {
		t.Errorf("Expected bucket 2, found %s", b)
	}
}

func TestHistogramWithInvalidCollection(t *testing.T) {
	worlds := &ingest.MutableWorlds{
		Base: camden.BuildGranarySquareForTests(t),
	}

	worldID := b6.CollectionID{Namespace: "diagonal.works/test/world", Value: 0}
	histogramID := b6.CollectionID{Namespace: "diagonal.works/test/histogram", Value: 0}

	histogram := ingest.CollectionFeature{
		CollectionID: histogramID,
		Keys:         []interface{}{"not a feature"},
		Values:       []interface{}{42},
	}
	w := worlds.FindOrCreateWorld(worldID.FeatureID())
	if err := w.AddFeature(&histogram); err != nil {
		t.Fatalf("Failed to add collection: %s", err)
	}

	r := NewHistogramRenderer(BasemapRenderRules, worlds)
	args := &TileArgs{Q: histogramID.String(), R: worldID.FeatureID()}
	projection := b6.NewTileMercatorProjection(16)
	tile, err := r.Render(projection.TileFromLatLng(s2.LatLngFromDegrees(51.53531, -0.12434)), args)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	if l := len(tile.Layers); l != 1 {
		t.Fatalf("Expected one layer, found %d", l)
	}
	if l := len(tile.Layers[0].Features); l != 0 {
		t.Fatalf("Expected no features, found %d", l)
	}
}
