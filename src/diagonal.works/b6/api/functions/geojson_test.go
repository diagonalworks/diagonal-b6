package functions

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/test/camden"
)

func TestGeoJSON(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	context := &api.Context{
		World: granarySquare,
	}
	features, err := find(context, b6.Keyed{Key: "#building"})
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	renderables := b6.AdaptCollection[interface{}, b6.Renderable](features)
	g, err := toGeoJSONCollection(context, renderables)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
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
	areas, err := geojsonAreas(&api.Context{}, g)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	i := areas.Begin()
	ok, err := i.Next()
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	if !ok {
		t.Fatal("Expected at least one area")
	}
	p2 := i.Value().(b6.Area).Polygon(0)
	if p2.Area() >= p1.Area() {
		t.Error("Expected clockwise GeoJSON polygons to be inverted")
	}
}
