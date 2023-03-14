package functions

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/geojson"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/test/camden"
)

func TestGeoJSON(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	context := &api.Context{
		World: granarySquare,
	}

	q := &pb.QueryProto{
		Query: &pb.QueryProto_Key{
			Key: &pb.KeyQueryProto{
				Key: "#building",
			},
		},
	}
	qq, _ := api.NewQueryFromProto(q, granarySquare)

	features, err := Find(qq, context)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	g, err := toGeoJSONCollection(features, context)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	if collection, ok := g.(*geojson.FeatureCollection); ok {
		if len(collection.Features) < 4 {
			t.Errorf("Unexpected number of features: %d", len(collection.Features))
		}
		for _, f := range collection.Features {
			if _, ok := f.Properties["#building"]; !ok {
				t.Errorf("Expected a building property, found: %v", f.Properties)
			}
		}
	} else {
		t.Errorf("Expected a FeatureCollection, found %T", g)
	}
}

func TestGeoJSONAreasInvertsLargePolygons(t *testing.T) {
	cs := geojson.Polygon{
		{ // Ordered clockwise
			geojson.Coordinate{Lat: 51.5371371, Lng: -0.1240464},
			geojson.Coordinate{Lat: 51.5370778, Lng: -0.1236840},
			geojson.Coordinate{Lat: 51.5354848, Lng: -0.1243698},
			geojson.Coordinate{Lat: 51.5355393, Lng: -0.1247150},
		},
	}
	g := geojson.NewFeatureWithGeometry(geojson.GeometryFromCoordinates(cs))
	p1 := cs.ToS2Polygon()
	areas, err := geojsonAreas(g, &api.Context{})
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}
	i := areas.Begin()
	if ok, err := i.Next(); !ok || err != nil {
		t.Errorf("Expected at least one area")
		return
	}
	p2 := i.Value().(b6.Area).Polygon(0)
	if p2.Area() >= p1.Area() {
		t.Errorf("Expected clockwise GeoJSON polygons to be inverted")
	}
}
