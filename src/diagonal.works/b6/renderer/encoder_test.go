package renderer

import (
	"reflect"
	"testing"

	"diagonal.works/b6"
	pb "diagonal.works/b6/proto"
	"github.com/golang/geo/s2"
	"google.golang.org/protobuf/proto"
)

func ll(lat float64, lng float64) s2.Point {
	return s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng))
}

func TestEncodeTile(t *testing.T) {
	location := b6.Tile{Z: 18, X: 130980, Y: 87134}

	polygon := s2.PolygonFromLoops([]*s2.Loop{
		s2.LoopFromPoints([]s2.Point{
			ll(51.5374960, -0.1256576),
			ll(51.5367729, -0.1259940),
			ll(51.5367018, -0.1256288),
			ll(51.5374191, -0.1253019),
		}),
	})
	content := &Tile{
		Layers: []*Layer{
			&Layer{
				Name: "landuse",
				Features: []*Feature{
					&Feature{
						Geometry: NewPolygon(polygon),
						Tags:     map[string]string{"class": "pedestrian"},
					},
				},
			},
			&Layer{
				Name: "poi_label",
				Features: []*Feature{
					&Feature{
						Geometry: NewPoint(ll(51.5373032, -0.1254470)),
						Tags:     map[string]string{"class": "fountain"},
					},
					&Feature{
						Geometry: NewPoint(ll(51.5371185, -0.1255186)),
						Tags:     map[string]string{"class": "fountain"},
					},
				},
			},
		},
	}

	encoded := EncodeTile(location, content)

	expectedLayers := []string{"background", "landuse", "poi_label"}
	if len(encoded.Layers) != len(expectedLayers) {
		t.Errorf("Expected %d layers, found %d", len(expectedLayers), len(encoded.Layers))
		return
	}
	for i, layer := range encoded.Layers {
		if layer.GetName() != expectedLayers[i] {
			t.Errorf("Expected layer %q, found %q", expectedLayers[i], layer.GetName())
		}
	}

	expectedCommands := [][]int{[]int{11}, []int{11}, []int{3, 3}}
	for i, layer := range encoded.Layers {
		if len(layer.Features) != len(expectedCommands[i]) {
			t.Errorf("Expected %d features, found %d", len(expectedCommands[i]), len(layer.Features))
			return
		}
		for j, feature := range layer.Features {
			if len(feature.Geometry) != expectedCommands[i][j] {
				t.Errorf("Expected %d commands, found %d", expectedCommands[i][j], len(feature.Geometry))
			}
		}
	}
}

func TestEncodeVectorTileGeometry(t *testing.T) {
	// Test cases taken from the examples in the Mapbox vector
	// tile specification.
	// See: https://github.com/mapbox/vector-tile-spec/tree/master/2.1

	// 4.3.5.1. Example Point
	e := NewEncoder(0, 0, "test", 1<<TileExtent)
	feature := e.StartFeature()
	e.MoveTo(1)
	e.XY(25, 17)
	if !reflect.DeepEqual(feature.Geometry, []uint32{9, 50, 34}) {
		t.Errorf("Unexpected point encoding: %+v", feature.Geometry)
	}

	// 4.3.5.2. Example Multi Point
	e = NewEncoder(0, 0, "test", 1<<TileExtent)
	feature = e.StartFeature()
	e.MoveTo(2)
	e.XY(5, 7)
	e.XY(3, 2)
	if !reflect.DeepEqual(feature.Geometry, []uint32{17, 10, 14, 3, 9}) {
		t.Errorf("Unexpected multi point encoding: %+v", feature.Geometry)
	}

	// 4.3.5.2. Example Linestring
	e = NewEncoder(0, 0, "test", 1<<TileExtent)
	feature = e.StartFeature()
	e.MoveTo(1)
	e.XY(2, 2)
	e.LineTo(2)
	e.XY(2, 10)
	e.XY(10, 10)
	if !reflect.DeepEqual(feature.Geometry, []uint32{9, 4, 4, 18, 0, 16, 16, 0}) {
		t.Errorf("Unexpected linestring encoding: %+v", feature.Geometry)
	}

	// 4.3.5.2. Example Multi Linestring
	e = NewEncoder(0, 0, "test", 1<<TileExtent)
	feature = e.StartFeature()
	e.MoveTo(1)
	e.XY(2, 2)
	e.LineTo(2)
	e.XY(2, 10)
	e.XY(10, 10)
	e.MoveTo(1)
	e.XY(1, 1)
	e.LineTo(1)
	e.XY(3, 5)
	if !reflect.DeepEqual(feature.Geometry, []uint32{9, 4, 4, 18, 0, 16, 16, 0, 9, 17, 17, 10, 4, 8}) {
		t.Errorf("Unexpected multi linestring encoding: %+v", feature.Geometry)
	}
}

func TestEncodeVectorTileGeometryRelativeToOrigin(t *testing.T) {
	// (51.53560, -0.12683) lies within tile 16/32744/21784
	ll1 := s2.LatLngFromDegrees(51.53560, -0.12683)
	ll2 := s2.LatLngFromDegrees(51.53671, -0.12618)
	zoom := uint(16)
	p := b6.NewTileMercatorProjection(zoom)
	tile := p.FromLatLng(ll1)
	originX, originY := int(tile.X)<<TileExtent, int(tile.Y)<<TileExtent

	p = b6.NewTileMercatorProjection(zoom + TileExtent)

	e := NewEncoder(originX, originY, "test", 1<<TileExtent)
	// Add two features on the same layer, to ensure that the origin is reset
	// between features.
	for i := 0; i < 2; i++ {
		e.StartFeature()
		e.MoveTo(2)
		e.Point(p.FromLatLng(ll1))
		e.Point(p.FromLatLng(ll2))
	}

	expected := []uint32{17, 7464, 1164, 970, 2661}
	for i := 0; i < 2; i++ {
		if !reflect.DeepEqual(e.Layer().Features[i].Geometry, expected) {
			t.Errorf("Unexpected geometry encoding: %+v", e.Layer().Features[i].Geometry)
		}
	}
}

func TestEncodeVectorTileTags(t *testing.T) {
	e := NewEncoder(0, 0, "test", 1<<TileExtent)
	feature := e.StartFeature()
	e.Tag("amenity", "bicycle_parking")
	e.Tag("capacity", 12)
	e.Tag("amenity", "bicycle_parking")
	e.Tag("capacity", 16)

	if !reflect.DeepEqual(e.Layer().Keys, []string{"amenity", "capacity"}) {
		t.Errorf("Unexpected layer keys: %+v", e.Layer().Keys)
	}

	values := []*pb.TileProto_Value{
		&pb.TileProto_Value{StringValue: proto.String("bicycle_parking")},
		&pb.TileProto_Value{IntValue: proto.Int64(12)},
		&pb.TileProto_Value{IntValue: proto.Int64(16)},
	}
	if !reflect.DeepEqual(e.Layer().Values, values) {
		t.Errorf("Unexpected layer values: %+v", e.Layer().Values)
	}

	if !reflect.DeepEqual(feature.Tags, []uint32{0, 0, 1, 1, 0, 0, 1, 2}) {
		t.Errorf("Unexpected feature tags: %+v", feature.Tags)
	}
}
