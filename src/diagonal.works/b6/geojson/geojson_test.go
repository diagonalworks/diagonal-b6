package geojson

import (
	"encoding/json"
	"reflect"
	"testing"

	pb "diagonal.works/b6/proto"
	"github.com/golang/geo/s2"
)

func TestMarshalPointFeature(t *testing.T) {
	feature := NewFeatureFromS2LatLng(s2.LatLngFromDegrees(51.5353602, -0.1260527))
	actual, err := json.Marshal(feature)
	if err != nil {
		t.Errorf("json.Marshal failed: %s", err)
		return
	}

	expected := `{"type":"Feature","geometry":{"type":"Point","coordinates":[-0.1260527,51.5353602]},"properties":{}}`
	if string(actual) != expected {
		t.Errorf("Unexpected GeoJSON encoding. Expected %s, found %s", expected, actual)
	}
}

func TestUnmarshalPointFeature(t *testing.T) {
	marshalled := []byte(`{"type":"Feature","geometry":{"type":"Point","coordinates":[-0.1260527,51.5353602]},"properties":{}}`)

	var feature Feature
	json.Unmarshal(marshalled, &feature)
	if point, ok := feature.Geometry.Coordinates.(Point); ok {
		high, low := s2.LatLngFromDegrees(51.5353603, -0.1260528), s2.LatLngFromDegrees(51.5353601, -0.1260526)
		rect := s2.EmptyRect().AddPoint(high).AddPoint(low)
		if !rect.ContainsLatLng(point.ToS2LatLng()) {
			t.Errorf("Point not within expected bounds")
		}
	} else {
		t.Errorf("Unexpected geometry type")
	}
}

func TestMarshalPolygon(t *testing.T) {
	loops := []*s2.Loop{
		s2.LoopFromPoints([]s2.Point{
			s2.PointFromLatLng(s2.LatLngFromDegrees(0.0, 0.0)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(0.0, 3.0)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(4.0, 3.0)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(4.0, 0.0)),
		}),
	}
	polygon := s2.PolygonFromLoops(loops)
	j := NewFeatureFromS2Polygon(polygon)
	coordinates, ok := j.Geometry.Coordinates.(Polygon)
	if !ok {
		t.Errorf("Wrong geometry type")
		return
	}
	if len(coordinates[0]) != polygon.Loop(0).NumVertices()+1 {
		t.Errorf("Expected GeoJSON linear ring to be explicitly closed")
	}

	points := make([]s2.Point, 3)
	for i := range points {
		points[i] = coordinates[0][i].ToS2Point()
	}
	center := s2.PointFromLatLng(s2.LatLngFromDegrees(2.0, 1.5))
	if !s2.OrderedCCW(points[0], points[1], points[2], center) {
		t.Errorf("Expected exterior loop to be orderer counterclockwise, per RFC7946")
	}
}

func TestMarshalPolygonWithHole(t *testing.T) {
	loops := []*s2.Loop{
		s2.LoopFromPoints([]s2.Point{
			s2.PointFromLatLng(s2.LatLngFromDegrees(0.0, 0.0)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(0.0, 3.0)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(4.0, 3.0)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(4.0, 0.0)),
		}),
		s2.LoopFromPoints([]s2.Point{
			s2.PointFromLatLng(s2.LatLngFromDegrees(2.0, 1.0)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(2.0, 2.0)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(3.0, 2.0)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(3.0, 1.0)),
		}),
	}
	polygon := s2.PolygonFromLoops(loops)
	j := NewFeatureFromS2Polygon(polygon)
	coordinates, ok := j.Geometry.Coordinates.(Polygon)
	if !ok {
		t.Errorf("Wrong geometry type")
		return
	}
	if len(coordinates) != 2 {
		t.Errorf("Expected GeoJSON polygon to have two linear rings")
		return
	}

	points := make([]s2.Point, 3)
	for i := range points {
		points[i] = coordinates[0][i].ToS2Point()
	}
	center := s2.PointFromLatLng(s2.LatLngFromDegrees(2.5, 1.5))
	if !s2.OrderedCCW(points[0], points[1], points[2], center) {
		t.Errorf("Expected exterior loop to be ordered counterclockwise, per RFC7946")
	}
	for i := range points {
		points[i] = coordinates[1][i].ToS2Point()
	}
	if s2.OrderedCCW(points[0], points[1], points[2], center) {
		t.Errorf("Expected interor loop to be ordered clockwise, per RFC7946")
	}
}

func TestMarshalPolygonProto(t *testing.T) {
	polygon := &pb.PolygonProto{
		Loops: []*pb.LoopProto{
			&pb.LoopProto{
				Points: []*pb.PointProto{
					&pb.PointProto{LatE7: 515356929, LngE7: -1276359},
					&pb.PointProto{LatE7: 515353710, LngE7: -1273698},
					&pb.PointProto{LatE7: 515350393, LngE7: -1270861},
				},
			},
		},
	}
	j := PolygonProtoToGeoJSON(polygon)
	coordinates, ok := j.Geometry.Coordinates.(Polygon)
	if !ok {
		t.Errorf("Wrong geometry type")
		return
	}
	if len(coordinates[0]) != len(polygon.Loops[0].Points)+1 {
		t.Errorf("Expected GeoJSON linear ring to be explicitly closed")
	}

	actual, err := json.Marshal(j)
	if err != nil {
		t.Errorf("json.Marshal failed: %s", err)
		return
	}
	expected := `{"type":"Feature","geometry":{"type":"Polygon","coordinates":[[[-0.1276359,51.5356929],[-0.1273698,51.535371],[-0.1270861,51.5350393],[-0.1276359,51.5356929]]]},"properties":{}}`
	if string(actual) != expected {
		t.Errorf("Unexpected GeoJSON encoding. Expected %s, found %s", expected, actual)
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		buffer string
		t      GeoJSON
	}{
		{
			`{"type":"Feature","geometry":{"type":"Point","coordinates":[-0.1260527,51.5353602]},"properties":{}}`,
			&Feature{},
		},
		{
			`{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"Point","coordinates":[-0.1260527,51.5353602]},"properties":{}}]}`,
			&FeatureCollection{},
		},
		{
			`{"type":"Point","coordinates":[-0.1260527,51.5353602]}`,
			&Geometry{},
		},
	}

	for _, test := range tests {
		if g, err := Unmarshal([]byte(test.buffer)); err == nil {
			if reflect.TypeOf(g) != reflect.TypeOf(test.t) {
				t.Errorf("Expected %s, found %s unmarshalling %s", reflect.TypeOf(test.t), reflect.TypeOf(g), test.buffer)
			}
		} else {
			t.Errorf("Expected no error, found %s unmarshalling %s", err, test.buffer)
		}
	}
}
