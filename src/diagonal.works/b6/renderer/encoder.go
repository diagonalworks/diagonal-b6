package renderer

import (
	"fmt"
	"log"

	"diagonal.works/b6"
	pb "diagonal.works/b6/proto"

	"github.com/golang/geo/r2"
	"github.com/golang/geo/s2"
	"google.golang.org/protobuf/proto"
)

const (
	TileExtent           = 12 // Implies 1 << 12, ie 4096, units per tile
	TileCommandMoveTo    = 1
	TileCommandLineTo    = 2
	TileCommandClosePath = 7
)

// See https://github.com/mapbox/vector-tile-spec/tree/master/2.1
func zigzagEncode(value int) uint32 {
	return uint32(int32(value)<<1) ^ uint32(int32(value)>>31)
}

func zigzagDecode(value uint32) int {
	return int((int32(value) >> 1) ^ (-(int32(value) & 1)))
}

func EncodeTile(location b6.Tile, content *Tile) *pb.TileProto {
	projection := b6.NewTileMercatorProjection(location.Z + TileExtent)
	encoded := &pb.TileProto{Layers: make([]*pb.TileProto_Layer, 0, len(content.Layers)+1)}
	encoded.Layers = append(encoded.Layers, newBackgroundLayer())
	for _, layer := range content.Layers {
		if len(layer.Features) == 0 {
			continue
		}
		encoder := NewEncoder(int(location.X<<TileExtent), int(location.Y<<TileExtent), layer.Name, 1<<TileExtent)
		encoded.Layers = append(encoded.Layers, encoder.Layer())
		for _, feature := range layer.Features {
			switch g := feature.Geometry.(type) {
			case *Polygon:
				simplifyAndEncodePolygon(g.ToS2Polygon(), encoder, projection)
			case *LineString:
				simplifyAndEncodeLineString(g.ToS2Polyline(), encoder, projection)
			case *Point:
				encodePoint(g.ToS2Point(), encoder, projection)
			default:
				panic(fmt.Sprintf("Can't encode %T", g))
			}
			if feature.ID != 0 {
				encoder.ID(feature.ID)
			}
			for key, value := range feature.Tags {
				encoder.Tag(key, value)
			}
		}
	}
	return encoded
}

func newBackgroundLayer() *pb.TileProto_Layer {
	e := NewEncoder(0, 0, "background", 1<<TileExtent)
	f := e.StartFeature()
	f.Type = pb.TileProto_POLYGON.Enum()
	e.MoveTo(1)
	e.XY(0, 0)
	e.LineTo(3)
	e.XY(4095, 0)
	e.XY(4095, 4095)
	e.XY(0, 4095)
	e.ClosePath()
	return e.Layer()
}

func simplifyAndEncodePolygon(polygon *s2.Polygon, e *Encoder, projection *b6.TileMercatorProjection) {
	feature := e.StartFeature()
	feature.Type = pb.TileProto_POLYGON.Enum()
	for _, loop := range polygon.Loops() {
		points := projectLoop(loop, projection)
		if len(points) > 1000 {
			points = Simplify(points, 5.0)
		}
		if len(points) > 1 {
			e.MoveTo(1)
			e.Point(points[0])
			e.LineTo(len(points) - 1)
			if !loop.IsHole() {
				// S2 polygons have counterclockwise outer loops, while mapbox tiles
				// expect clockwise. However, the tile coordinate system has y
				// increasing downwards, effectively reversing the direction of
				// polygons, making the S2 native counterclockwise order appropriate.
				for i := 1; i < len(points); i++ {
					e.Point(points[i])
				}
			} else {
				for i := len(points) - 1; i > 0; i-- {
					e.Point(points[i])
				}
			}
			e.ClosePath()
		}
	}
}

func projectLoop(loop *s2.Loop, projection *b6.TileMercatorProjection) []r2.Point {
	points := make([]r2.Point, loop.NumVertices())
	for i := 0; i < loop.NumVertices(); i++ {
		points[i] = projection.Project(loop.Vertex(i))
	}
	return points
}

func simplifyAndEncodeLineString(line *s2.Polyline, e *Encoder, projection *b6.TileMercatorProjection) {
	tf := e.StartFeature()
	tf.Type = pb.TileProto_LINESTRING.Enum()
	e.MoveTo(1)
	e.Point(projection.Project((*line)[0]))
	e.LineTo(len(*line) - 1)
	for i := 1; i < len(*line); i++ {
		e.Point(projection.Project((*line)[i]))
	}
}

func encodePoint(point s2.Point, e *Encoder, projection *b6.TileMercatorProjection) {
	tf := e.StartFeature()
	tf.Type = pb.TileProto_POINT.Enum()
	e.MoveTo(1)
	e.Point(projection.Project(point))
}

type Encoder struct {
	originX      int
	originY      int
	layer        *pb.TileProto_Layer
	feature      *pb.TileProto_Feature
	cursorX      int
	cursorY      int
	keys         map[string]uint32
	stringValues map[string]uint32
	int64Values  map[int64]uint32
}

func NewEncoder(originX int, originY int, name string, extent int) *Encoder {
	return &Encoder{
		originX: originX,
		originY: originY,
		layer: &pb.TileProto_Layer{
			Name:     proto.String(name),
			Version:  proto.Uint32(2),
			Extent:   proto.Uint32(uint32(extent)),
			Keys:     make([]string, 0, 0),
			Values:   make([]*pb.TileProto_Value, 0, 0),
			Features: make([]*pb.TileProto_Feature, 0, 0),
		},
		keys:         make(map[string]uint32),
		stringValues: make(map[string]uint32),
		int64Values:  make(map[int64]uint32),
	}
}

func (e *Encoder) Layer() *pb.TileProto_Layer {
	return e.layer
}

func (e *Encoder) StartFeature() *pb.TileProto_Feature {
	e.cursorX = e.originX
	e.cursorY = e.originY
	e.feature = &pb.TileProto_Feature{
		Tags:     make([]uint32, 0, 0),
		Geometry: make([]uint32, 0, 0),
	}
	e.layer.Features = append(e.layer.Features, e.feature)
	return e.feature
}

func (e *Encoder) MoveTo(count int) {
	e.feature.Geometry = append(e.feature.Geometry, (TileCommandMoveTo&0x7)|(uint32(count)<<3))
}

func (e *Encoder) LineTo(count int) {
	e.feature.Geometry = append(e.feature.Geometry, (TileCommandLineTo&0x7)|(uint32(count)<<3))
}

func (e *Encoder) ClosePath() {
	count := 1
	e.feature.Geometry = append(e.feature.Geometry, (TileCommandClosePath&0x7)|(uint32(count)<<3))
}

func (e *Encoder) XY(x int, y int) {
	e.feature.Geometry = append(e.feature.Geometry, zigzagEncode(x-e.cursorX), zigzagEncode(y-e.cursorY))
	e.cursorX = x
	e.cursorY = y
}

func (e *Encoder) Point(p r2.Point) {
	e.XY(int(p.X), int(p.Y))
}

func (e *Encoder) ID(id uint64) {
	e.feature.Id = proto.Uint64(id)
}

func (e *Encoder) Tag(key string, value interface{}) {
	switch v := value.(type) {
	case string:
		index, ok := e.stringValues[v]
		if !ok {
			protoValue := &pb.TileProto_Value{StringValue: proto.String(v)}
			e.layer.Values = append(e.layer.Values, protoValue)
			index = uint32(len(e.layer.Values) - 1)
			e.stringValues[v] = index
		}
		e.feature.Tags = append(append(e.feature.Tags, e.key(key)), index)
	case int64:
		index, ok := e.int64Values[v]
		if !ok {
			protoValue := &pb.TileProto_Value{IntValue: proto.Int64(v)}
			e.layer.Values = append(e.layer.Values, protoValue)
			index = uint32(len(e.layer.Values) - 1)
			e.int64Values[v] = index
		}
		e.feature.Tags = append(append(e.feature.Tags, e.key(key)), index)
	case int:
		e.Tag(key, int64(v))
	default:
		log.Printf("Can't encode tag: key=%s, value=%v, type=%T", key, value, value)
	}
}

func (e *Encoder) key(key string) uint32 {
	index, ok := e.keys[key]
	if !ok {
		e.layer.Keys = append(e.layer.Keys, key)
		index = uint32(len(e.layer.Keys) - 1)
		e.keys[key] = index
	}
	return index
}
