package renderer

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/test/testcamden"
	"github.com/golang/geo/s2"
)

func TestQueryRenderer(t *testing.T) {
	granarySquare := testcamden.BuildGranarySquare(t)
	if granarySquare == nil {
		return
	}

	projection := b6.NewTileMercatorProjection(16)
	r := NewQueryRenderer(granarySquare, 2)
	args := &TileArgs{Q: "[#amenity=cafe]"}
	tile, err := r.Render(projection.TileFromLatLng(s2.LatLngFromDegrees(51.53531, -0.12434)), args)
	if err == nil {
		if len(tile.Layers) == 1 && tile.Layers[0].Name == "query" {
			if len(tile.Layers[0].Features) < 4 {
				t.Errorf("Expected features in layer")
			}
		} else {
			t.Errorf("Expected one layer named 'query', found %d", len(tile.Layers))
		}
	} else {
		t.Errorf("Expected no error, found: %s", err)
	}
}
