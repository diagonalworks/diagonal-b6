package renderer

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/test/camden"
	"github.com/golang/geo/s2"
)

func TestQueryRenderer(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	projection := b6.NewTileMercatorProjection(16)
	r := NewQueryRenderer(granarySquare, 2)
	args := &TileArgs{Q: "[#amenity=cafe]"}
	tile, err := r.Render(projection.TileFromLatLng(s2.LatLngFromDegrees(51.53531, -0.12434)), args)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	if len(tile.Layers) != 1 || tile.Layers[0].Name != "query" {
		t.Fatalf("Expected one layer named 'query', found %d", len(tile.Layers))
	}
	if len(tile.Layers[0].Features) < 4 {
		t.Errorf("Expected at least 4 features in layer, got %d", len(tile.Layers[0].Features))
	}
}
