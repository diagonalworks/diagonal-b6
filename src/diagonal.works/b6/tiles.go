package b6

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"

	pb "diagonal.works/b6/proto"
	"github.com/golang/geo/r2"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

type Tile struct {
	X uint
	Y uint
	Z uint
}

func (t Tile) String() string {
	return fmt.Sprintf("%d/%d/%d", t.Z, t.X, t.Y)
}

func (t Tile) Parent() Tile {
	return Tile{X: t.X / 2, Y: t.Y / 2, Z: t.Z - 1}
}

func (t Tile) Children() []Tile {
	children := make([]Tile, 4)
	children[0] = Tile{X: t.X * 2, Y: t.Y * 2, Z: t.Z + 1}
	children[1] = Tile{X: t.X*2 + 1, Y: t.Y * 2, Z: t.Z + 1}
	children[2] = Tile{X: t.X * 2, Y: t.Y*2 + 1, Z: t.Z + 1}
	children[3] = Tile{X: t.X*2 + 1, Y: t.Y*2 + 1, Z: t.Z + 1}
	return children
}

func (t Tile) ToID() TileID {
	return TileIDFromXYZ(t.X, t.Y, t.Z)
}

func (t Tile) ToToken() string {
	return t.ToID().ToToken()
}

var InvalidRect = s2.RectFromLatLng(s2.LatLng{s1.InfAngle(), s1.InfAngle()})

// RectBound returns an s2.Rect that marks the boundary of this tile. As
// s2.Rect operates in lat,lng space, it's guaranteed to be exactly match
// the tile.
func (t *Tile) RectBound() s2.Rect {
	if t.Z > 31 {
		return InvalidRect
	}
	n := uint(1) << t.Z
	if t.X >= n || t.Y >= n {
		return InvalidRect
	}
	projection := NewTileMercatorProjection(t.Z)
	rect := s2.RectFromLatLng(projection.ToLatLng(r2.Point{float64(t.X), float64(t.Y)}))
	return rect.AddPoint(projection.ToLatLng(r2.Point{float64(t.X + 1), float64(t.Y + 1)}))
}

// PolygonBound returns an s2.Polygon that marks the boundary of this tile. As
// tiles operate in lat,lng space, the bounds aren't guaranteed to exactly
// match the tile, but will be reasonable enough for rendering use cases.
func (t *Tile) PolygonBound() *s2.Polygon {
	return t.PolygonBoundWithEpsilon(0.001)
}

// PolygonBoundWithEpsilon returns an s2.Polygon that marks the boundary of
// this tile. As tiles operate in lat,lng space, the bounds aren't guaranteed
// to exactly match the tile, but will be within the given epsilon of the
// diagonal distance.
func (t *Tile) PolygonBoundWithEpsilon(epsilon float64) *s2.Polygon {
	vertices := []r2.Point{
		{float64(t.X), float64(t.Y)},
		{float64(t.X), float64(t.Y + 1)},
		{float64(t.X + 1), float64(t.Y + 1)},
		{float64(t.X + 1), float64(t.Y)},
	}
	projection := NewTileMercatorProjection(t.Z)
	diagonal := s2.PointFromLatLng(projection.ToLatLng(vertices[0])).Distance(s2.PointFromLatLng(projection.ToLatLng(vertices[2])))
	tesselator := s2.NewEdgeTessellator(projection, diagonal*s1.Angle(epsilon))
	points := make([]s2.Point, 0, 4)
	for i := 0; i < 4; i++ {
		edge := tesselator.AppendUnprojected(vertices[i], vertices[(i+1)%4], make([]s2.Point, 0, 2))
		points = append(points, edge[0:len(edge)-1]...)
	}
	return s2.PolygonFromLoops([]*s2.Loop{s2.LoopFromPoints(points)})
}

var tileURLPattern = regexp.MustCompile(`/(\d+)/(\d+)/(\d+).[a-z]{3}$`)

// TileFromURL returns a tile initialised from a URL path ending with tile
// coordinates in z/x/y order, eg /tiles/earth/17/65490/43568.mvt.
// See https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames
func TileFromURLPath(path string) (Tile, error) {
	var tile Tile
	match := tileURLPattern.FindAllStringSubmatch(path, 1)
	if match == nil || len(match) != 1 {
		return tile, fmt.Errorf("Failed to parse tile URL: %s", path)
	}
	var coordinates [3]int
	for i := 0; i < 3; i++ {
		var err error
		coordinates[i], err = strconv.Atoi(match[0][i+1])
		if err != nil {
			ordinates := [3]string{"z", "x", "y"}
			return tile, fmt.Errorf("Failed to parse %s ordinate: %s", ordinates[i], path)
		}
	}
	tile.Z, tile.X, tile.Y = uint(coordinates[0]), uint(coordinates[1]), uint(coordinates[2])
	return tile, nil
}

// TileID represents a single map tile are a 64 integer. The top 5 bits represents the zoom level,
// the following the XY coordinates. By construction, the ID of a tile's parent is always smaller
// than the ID of a tile. This property is useful for binary searching for the parents of a tile
// ID in a sorted list.
type TileID uint64

const tileIDZBits = uint64(5)

func TileIDFromXYZ(x uint, y uint, z uint) TileID {
	tileIDYBits := uint64(z)
	return TileID((uint64(z) << (64 - tileIDZBits)) | (uint64(y) << tileIDYBits) | uint64(x))
}

func (t TileID) Parent() TileID {
	tile := t.ToTile()
	tile = tile.Parent()
	return tile.ToID()
}

func (t TileID) ToXYZ() (uint, uint, uint) {
	z := uint64(t) >> (64 - tileIDZBits)
	tileIDYBits := uint64(z)
	y := (uint64(t) >> tileIDYBits) & (1<<tileIDYBits - 1)
	x := uint64(t) & (1<<tileIDYBits - 1)
	return uint(x), uint(y), uint(z)
}

func (t TileID) ToTile() Tile {
	x, y, z := t.ToXYZ()
	return Tile{X: x, Y: y, Z: z}
}

func (t TileID) ToToken() string {
	return strconv.FormatUint(uint64(t), 32)
}

func TileIDFromToken(token string) TileID {
	n, _ := strconv.ParseUint(token, 32, 64)
	// TODO: Would be nice to have an invalid tile ID if err != nil
	return TileID(n)
}

type TileIDs []TileID

func (t TileIDs) Len() int           { return len(t) }
func (t TileIDs) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t TileIDs) Less(i, j int) bool { return t[i] < t[j] }

// Contains returns true if either 'other', or one of it's ancestors, is in the list.
// Relies on t being sorted.
func (t TileIDs) Contains(other Tile) bool {
	n := len(t)
	for {
		id := other.ToID()
		i := sort.Search(n, func(j int) bool {
			return t[j] >= id
		})
		if i < n && t[i] == id {
			return true
		}
		if other.Z == 0 {
			break
		} else {
			other = other.Parent()
		}
	}
	return false
}

type ZoomRange struct {
	Min uint
	Max uint
}

func (z *ZoomRange) IsEmpty() bool {
	return z.Min > z.Max
}

var EmptyZoomRange = ZoomRange{Min: 2, Max: 1}

type TileMercatorProjection struct {
	level    uint
	extent   float64
	mercator s2.Projection
}

func MakeTileMercatorProjection(level uint) TileMercatorProjection {
	extent := float64(uint64(1 << (level - 1)))
	mercator := s2.NewMercatorProjection(extent)
	return TileMercatorProjection{level: level, extent: extent, mercator: mercator}
}

func NewTileMercatorProjection(level uint) *TileMercatorProjection {
	p := MakeTileMercatorProjection(level)
	return &p
}

func (p *TileMercatorProjection) Project(point s2.Point) r2.Point {
	projected := p.mercator.Project(point)
	return r2.Point{p.extent + projected.X, p.extent - projected.Y}
}

func (p *TileMercatorProjection) Unproject(point r2.Point) s2.Point {
	point = r2.Point{point.X - p.extent, -point.Y + p.extent}
	return p.mercator.Unproject(point)
}

func (p *TileMercatorProjection) FromLatLng(ll s2.LatLng) r2.Point {
	projected := p.mercator.FromLatLng(ll)
	return r2.Point{p.extent + projected.X, p.extent - projected.Y}
}

func (p *TileMercatorProjection) ToLatLng(projected r2.Point) s2.LatLng {
	point := r2.Point{projected.X - p.extent, -projected.Y + p.extent}
	return p.mercator.ToLatLng(point)
}

func (p *TileMercatorProjection) TileFromLatLng(ll s2.LatLng) Tile {
	projected := p.FromLatLng(ll)
	return Tile{X: uint(projected.X), Y: uint(projected.Y), Z: p.level}
}

func (p *TileMercatorProjection) Interpolate(f float64, a, b r2.Point) r2.Point {
	a = r2.Point{a.X - p.extent, -a.Y + p.extent}
	b = r2.Point{b.X - p.extent, -b.Y + p.extent}
	interpolated := p.mercator.Interpolate(f, a, b)
	return r2.Point{p.extent + interpolated.X, p.extent - interpolated.Y}
}

func (p *TileMercatorProjection) WrapDistance() r2.Point {
	return r2.Point{p.extent, p.extent}
}

func (p *TileMercatorProjection) FromPointProto(point *pb.PointProto) r2.Point {
	return p.FromLatLng(s2.LatLng{s1.Angle(point.LatE7) * s1.E7, s1.Angle(point.LngE7) * s1.E7})
}

func CoverCellIDWithTiles(cellID s2.CellID, zoom uint) []Tile {
	// TODO: This is only an approximation near the poles
	cell := s2.CellFromCellID(cellID)
	projection := NewTileMercatorProjection(zoom)
	bottomLeft := projection.Project(cell.Vertex(0))
	bottomRight := projection.Project(cell.Vertex(1))
	topRight := projection.Project(cell.Vertex(2))
	topLeft := projection.Project(cell.Vertex(3))

	top := uint(math.Min(topLeft.Y, topRight.Y))
	bottom := uint(math.Max(bottomLeft.Y, bottomRight.Y))
	left := uint(math.Min(topLeft.X, bottomLeft.X))
	right := uint(math.Max(topRight.X, bottomRight.X))

	tiles := make([]Tile, 0, 1)
	for y := top; y <= bottom; y++ {
		for x := left; x <= right; x++ {
			tiles = append(tiles, Tile{X: x, Y: y, Z: zoom})
		}
	}
	return tiles
}

func CoverCellUnionWithTiles(union s2.CellUnion, zoom uint) []Tile {
	tiles := make([]Tile, 0, 1)
	seen := make(map[Tile]struct{})
	for _, id := range union {
		for _, tile := range CoverCellIDWithTiles(id, zoom) {
			if _, ok := seen[tile]; !ok {
				tiles = append(tiles, tile)
				seen[tile] = struct{}{}
			}
		}
	}
	return tiles
}

func CoverCellUnionWithTilesAcrossZooms(union s2.CellUnion, zooms ZoomRange) []Tile {
	if zooms.IsEmpty() {
		return []Tile{}
	}
	tileSet := make(map[TileID]struct{})
	for _, cellID := range union {
		coverCellIDWithTiles(tileSet, cellID, zooms.Max)
	}
	tiles := make([]Tile, 0, len(tileSet)*2)
	for tileID := range tileSet {
		tiles = append(tiles, tileID.ToTile())
	}
	lastZoomBegin := 0
	lastZoomEnd := len(tiles)
	for zoom := zooms.Max - 1; zoom >= zooms.Min; zoom-- {
		for i := lastZoomBegin; i < lastZoomEnd; i++ {
			parent := tiles[i].Parent()
			if _, ok := tileSet[parent.ToID()]; !ok {
				tiles = append(tiles, parent)
				tileSet[parent.ToID()] = struct{}{}
			}
		}
		lastZoomBegin = lastZoomEnd
		lastZoomEnd = len(tiles)
	}
	return tiles
}

func coverCellIDWithTiles(tileSet map[TileID]struct{}, cellID s2.CellID, zoom uint) {
	cell := s2.CellFromCellID(cellID)
	projection := NewTileMercatorProjection(zoom)
	bottomLeft := projection.Project(cell.Vertex(0))
	bottomRight := projection.Project(cell.Vertex(1))
	topRight := projection.Project(cell.Vertex(2))
	topLeft := projection.Project(cell.Vertex(3))

	top := uint(math.Min(topLeft.Y, topRight.Y))
	bottom := uint(math.Max(bottomLeft.Y, bottomRight.Y))
	left := uint(math.Min(topLeft.X, bottomLeft.X))
	right := uint(math.Max(topRight.X, bottomRight.X))

	for y := top; y <= bottom; y++ {
		for x := left; x <= right; x++ {
			tile := Tile{X: x, Y: y, Z: zoom}
			tileSet[tile.ToID()] = struct{}{}
		}
	}
}
