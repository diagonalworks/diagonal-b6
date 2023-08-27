package b6

import (
	"math"
	"sort"
	"testing"

	"github.com/golang/geo/r2"
	"github.com/golang/geo/s2"
)

func TestBadTilesReturnInvalidBounds(t *testing.T) {
	tiles := []Tile{
		{1, 2, 32}, // z is too large
		{5, 1, 2},  // x is too large
		{1, 5, 2},  // y is too large
	}
	for _, tile := range tiles {
		bounds := tile.RectBound()
		if bounds.IsValid() {
			t.Errorf("Expected invalid bounds for tile %s", tile.String())
		}
	}
}

func TestTileBoundsHaveCorrectGeometry(t *testing.T) {
	tile := Tile{X: 130980, Y: 87135, Z: 18}
	bounds := tile.RectBound()
	lls := []s2.LatLng{
		s2.LatLngFromDegrees(51.536933, -0.126037), // Top leftish
		s2.LatLngFromDegrees(51.536543, -0.125624), // Center
		s2.LatLngFromDegrees(51.536112, -0.125007), // Bottom right
	}
	for _, ll := range lls {
		if !bounds.ContainsLatLng(ll) {
			t.Errorf("Bounds does not contain expected point: %s", ll)
		}
	}
	size := bounds.Size()
	area := AngleToMeters(size.Lat) * AngleToMeters(size.Lng)
	if area < 14000.0 || area > 15000.0 {
		t.Errorf("Unexected tile bounds area: %fm2", area)
	}
}

func TestPolygonBound(t *testing.T) {
	tile := Tile{X: 8, Y: 2, Z: 4}
	rect := tile.RectBound()
	projection := NewTileMercatorProjection(tile.Z)
	p := projection.Unproject(r2.Point{X: 8.5, Y: 2.99})

	points := make([]s2.Point, 4)
	for i := 0; i < 4; i++ {
		points[i] = s2.PointFromLatLng(rect.Vertex(i))
	}
	square := s2.PolygonFromLoops([]*s2.Loop{s2.LoopFromPoints(points)})
	if !rect.ContainsPoint(p) || square.ContainsPoint(p) {
		t.Errorf("Test data failure: square shouldn't contain p")
	}

	polygon := tile.PolygonBound()
	if !polygon.ContainsPoint(p) {
		t.Errorf("Expected TesselatedBounds() to contain p")
	}

	ratio := math.Abs((rect.Area() - polygon.Area()) / polygon.Area())
	if ratio > 0.001 {
		t.Errorf("Expected rect area and polygon area to be similar, found %f", ratio)
	}
}

func TestTileFromURL(t *testing.T) {
	tests := []struct {
		path string
		ok   bool
		z    uint
		x    uint
		y    uint
	}{
		{"/tiles/earth/17/65490/43568.mvt", true, 17, 65490, 43568},
		{"/17/65490/43568.mvt", true, 17, 65490, 43568},
		{"/tiles/earth/17/65490x/43568.mvt", false, 0, 0, 0},
		{"/tiles/earth/17/43568.mvt", false, 0, 0, 0},
		{"/17/-65490/43568.mvt", false, 0, 0, 0},
	}
	for _, test := range tests {
		tile, err := TileFromURLPath(test.path)
		if test.ok {
			if err != nil {
				t.Errorf("Expected no error for path %q, found: %s", test.path, err)
			} else {
				if tile.Z != test.z {
					t.Errorf("Expected z ordinate %d, found %d in path %q", test.z, tile.Z, test.path)
				}
				if tile.X != test.x {
					t.Errorf("Expected x ordinate %d, found %d in path %q", test.x, tile.X, test.path)
				}
				if tile.Y != test.y {
					t.Errorf("Expected y ordinate %d, found %d in path %q", test.y, tile.Y, test.path)
				}
			}
		} else {
			if err == nil {
				t.Errorf("Expected error for path %q, found none", test.path)
			}
		}
	}
}

func TestTestID(t *testing.T) {
	// A tile in Granary Square, eg https://tile.openstreetmap.org/17/65490/43568.png
	tileID := TileIDFromXYZ(65490, 43568, 17)

	x, y, z := tileID.ToXYZ()
	expectedX, expectedY, expectedZ := uint(65490), uint(43568), uint(17)
	if x != expectedX || y != expectedY || z != expectedZ {
		t.Errorf("Expected %d,%d,%d, found %d,%d,%d", expectedX, expectedY, expectedZ, x, y, z)
	}

	token := tileID.ToToken()
	expectedToken := "8g00005a61vui"
	if token != expectedToken {
		t.Errorf("Expected token %q, found %q", expectedToken, token)
	}

	if TileIDFromToken(token) != tileID {
		t.Errorf("Expected tile ID %d, found %d", uint64(tileID), uint64(TileIDFromToken(token)))
	}
}

func TestTileChildren(t *testing.T) {
	tile := Tile{X: 65490, Y: 43568, Z: 17}
	children := tile.Children()

	expected := 4
	if len(children) != expected {
		t.Fatalf("Expected %d children, found %d", expected, len(children))
	}

	expectedTiles := []Tile{
		{X: 130980, Y: 87136, Z: 18},
		{X: 130981, Y: 87136, Z: 18},
		{X: 130980, Y: 87137, Z: 18},
		{X: 130981, Y: 87137, Z: 18},
	}
	for i, expectedTile := range expectedTiles {
		if children[i] != expectedTile {
			t.Errorf("Expected tile %v, found %v", expectedTile, children[i])
		}
	}
}

func TestTileIDContains(t *testing.T) {
	tile := Tile{X: 65490, Y: 43568, Z: 17}
	children := tile.Children()

	missing := 2
	ids := make(TileIDs, 0, 3)
	for i := 0; i < 4; i++ {
		if i != missing {
			ids = append(ids, children[i].ToID())
		}
	}

	sort.Sort(ids)
	contained := []Tile{children[3], children[1].Children()[0]}
	for i, tile := range contained {
		if !ids.Contains(tile) {
			t.Errorf("Expected ids to contain tile %d: %q", i, tile.ToToken())
		}
	}
	notContained := []Tile{children[missing], children[missing].Children()[3]}
	for i, tile := range notContained {
		if ids.Contains(tile) {
			t.Errorf("Expected ids to not contain tile %d: %q", i, tile.ToToken())
		}
	}
}

func TestTileMercatorProjection(t *testing.T) {
	coordinates := []struct {
		lat float64
		lng float64
		x   float64
		y   float64
		z   uint
	}{
		{51.53560, -0.12683, 130979.64, 87136.56, 18},
		{51.53671, -0.12618, 130980.12, 87135.27, 18},
	}
	for _, c := range coordinates {
		projection := NewTileMercatorProjection(c.z)
		p := projection.FromLatLng(s2.LatLngFromDegrees(c.lat, c.lng))
		if math.Abs(p.X-c.x) > 1.0 || math.Abs(p.Y-c.y) > 1.0 {
			t.Errorf("Unexpected projected values: %v", p)
		}
		ll := projection.ToLatLng(p)
		if math.Abs(ll.Lat.Degrees()-c.lat) > 0.1 || math.Abs(ll.Lng.Degrees()-c.lng) > 0.1 {
			t.Errorf("Unexpected unprojected values: %v", ll)
		}
	}
}

func TestTileFromLatLng(t *testing.T) {
	projection := NewTileMercatorProjection(16)
	tile := projection.TileFromLatLng(s2.LatLngFromDegrees(51.53531, -0.12434))

	expected := Tile{Z: 16, X: 32745, Y: 21784}
	if tile != expected {
		t.Errorf("Expected %+v, found %+v", expected, tile)
	}
}

func TestCoverCellIDWithTiles(t *testing.T) {
	granarySquare := s2.CellIDFromToken("48761b3dc")
	tiles := CoverCellIDWithTiles(granarySquare, 16)

	expectedTiles := []Tile{{X: 32744, Y: 21784, Z: 16}, {X: 32745, Y: 21784, Z: 16}}
	if len(tiles) != len(expectedTiles) {
		t.Errorf("Expected length %d, found %d", len(expectedTiles), len(tiles))
	}

	for i := 0; i < len(expectedTiles); i++ {
		if tiles[i] != expectedTiles[i] {
			t.Errorf("Expected tile %v, found %v", expectedTiles[i], tiles[i])
		}
	}
}

func TestCoverCellUnionWithTiles(t *testing.T) {
	granarySquare := s2.CellIDFromToken("48761b3dc")
	kingsCross := s2.CellIDFromToken("48761b3c4")
	union := s2.CellUnion{granarySquare, kingsCross}
	tiles := CoverCellUnionWithTiles(union, 16)

	expectedTiles := []Tile{{X: 32744, Y: 21784, Z: 16}, {X: 32745, Y: 21784, Z: 16}, {X: 32745, Y: 21785, Z: 16}}
	if len(tiles) != len(expectedTiles) {
		t.Errorf("Expected length %d, found %d", len(expectedTiles), len(tiles))
	}

	for i := 0; i < len(expectedTiles); i++ {
		if tiles[i] != expectedTiles[i] {
			t.Errorf("Expected tile %v, found %v", expectedTiles[i], tiles[i])
		}
	}
}

func TestCoverCellUnionWithTilesAcrossZooms(t *testing.T) {
	tokens := []string{"48760da19", "48760da1f", "48760da23", "48760da25", "48760da31", "48760da33", "48760da3b"}
	chiswickFlyover := make(s2.CellUnion, len(tokens))
	for i, token := range tokens {
		chiswickFlyover[i] = s2.CellIDFromToken(token)
	}

	tiles := CoverCellUnionWithTilesAcrossZooms(chiswickFlyover, ZoomRange{Max: 18, Min: 8})
	expected := 48
	if len(tiles) != 48 {
		t.Errorf("Expected covering of %d tiles, found %d", expected, len(tiles))
	}
}
