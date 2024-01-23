package functions

import (
	"fmt"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/test/camden"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

func TestSightlineFunction(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	context := &api.Context{
		World: granarySquare,
	}

	from := b6.GeometryFromLatLng(s2.LatLngFromDegrees(51.53545, -0.12561))
	area, err := sightline(context, from, 250.0)
	if err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

	if area.Len() != 1 {
		t.Errorf("Expected 1 area, found %d", area.Len())
	}
}

func TestSightline(t *testing.T) {
	sightlines := []struct {
		name string
		f    func(from s2.Point, radius s1.Angle, w b6.World) *s2.Polygon
	}{
		{"TestSightlineDefault", Sightline},
		{"TestSightlineUsingPolygonIntersection", SightlineUsingPolygonIntersection},
		{"TestSightlineUsingPolarCoordinates", SightlineUsingPolarCoordinates},
		{"TestSightlineUsingPolarCoordinates2", SightlineUsingPolarCoordinates2},
	}

	for _, s := range sightlines {
		t.Run(fmt.Sprintf("%s/Sightline", s.name), func(t *testing.T) { ValidateSightline(s.f, s.name, t) })
		t.Run(fmt.Sprintf("%s/SightlineInsideBuilding", s.name), func(t *testing.T) { ValidateSightlineInsideBuilding(s.f, s.name, t) })
	}
}

func ValidateSightline(computeSightline func(from s2.Point, radius s1.Angle, w b6.World) *s2.Polygon, name string, t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	radius := 250.0
	from := s2.PointFromLatLng(s2.LatLngFromDegrees(51.53545, -0.12561))
	sightline := computeSightline(from, b6.MetersToAngle(radius), granarySquare)

	cap := s2.CapFromCenterAngle(from, b6.MetersToAngle(radius))
	ratio := sightline.Area() / cap.Area()
	if ratio < 0.2 || ratio > 0.3 {
		t.Errorf("Unexpected ratio between sightline area and cap area: %f", ratio)
	}

	points := []struct {
		lat     float64
		lng     float64
		visible bool
	}{
		{51.53525, -0.12502, true},  // Inside Granary Square
		{51.53576, -0.12516, false}, // Within Central St Martins
		{51.53578, -0.12587, false}, // In the shadow of the West side of Central St Martins
	}

	actual := make([]bool, len(points))
	for i, p := range points {
		if actual[i] = sightline.ContainsPoint(s2.PointFromLatLng(s2.LatLngFromDegrees(p.lat, p.lng))); actual[i] != p.visible {
			t.Errorf("Unexpected visibility for (%f,%f)", p.lat, p.lng)
		}
	}

	if t.Failed() {
		g := geojson.NewFeatureCollection()
		g.AddFeature(geojson.NewFeatureFromS2Polygon(sightline))
		for i, p := range points {
			f := geojson.NewFeatureFromS2LatLng(s2.LatLngFromDegrees(p.lat, p.lng))
			f.Properties["expected"] = fmt.Sprintf("%v", p.visible)
			f.Properties["actual"] = fmt.Sprintf("%v", actual[i])
			g.AddFeature(f)
		}
		g.WriteToFile(fmt.Sprintf("test-%s.geojson", name))
	}
}

func ValidateSightlineInsideBuilding(computeSightline func(from s2.Point, radius s1.Angle, w b6.World) *s2.Polygon, name string, t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	radius := 250.0
	from := s2.PointFromLatLng(s2.LatLngFromDegrees(51.535280, -0.124357))
	computeSightline(from, b6.MetersToAngle(radius), granarySquare)
}

func TestSightlineDoesntHaveSpikes(t *testing.T) {
	// 'spikes' occur in the sighline polygon when numerical accuracy issues
	// lead to a very thin field of visibility incorrectly appearing in the
	// join between two edges of a building. We filter these out.
	granarySquare := camden.BuildGranarySquareForTests(t)

	// This location exhibits a spike - detmined visually though Atlas, and
	// the original implementation of the algorithm
	ll := s2.LatLngFromDegrees(51.536703, -0.126709)

	radius := 250.0
	sightline := Sightline(s2.PointFromLatLng(ll), b6.MetersToAngle(radius), granarySquare)
	boundary := s2.RegularLoop(s2.PointFromLatLng(ll), b6.MetersToAngle(100), 128)
	intersections := countIntersections(sightline.Loop(0), boundary)
	expected := 2 // There were 6 intersections before spike removal.
	if intersections != expected {
		t.Errorf("Expected %d intersections, found %d", expected, intersections)
	}
}

func countIntersections(a *s2.Loop, b *s2.Loop) int {
	intersections := 0
	for i := 0; i < a.NumEdges(); i++ {
		for j := 0; j < b.NumEdges(); j++ {
			if s2.CrossingSign(a.Vertex(i), a.Vertex(i+1), b.Vertex(j), b.Vertex(j+1)) != s2.DoNotCross {
				intersections++
			}
		}
	}
	return intersections
}

func TestOccludeWithCenterCloseToEdge(t *testing.T) {
	// Tests the situation in which the boundary arc section of an occlusion wraps around
	// the first vertex, which practically occurs when the the center is close to an edge.
	a := s2.PointFromLatLng(s2.LatLngFromDegrees(51.51898, -0.09662))
	b := s2.PointFromLatLng(s2.LatLngFromDegrees(51.51869, -0.09539))
	center := s2.PointFromLatLng(s2.LatLngFromDegrees(51.51891, -0.09657))
	o := Occlude(a, b, center, b6.MetersToAngle(250.0))
	if o == nil {
		t.Error("Expected an occlusion, found nil")
	} else if !o.ContainsPoint(s2.PointFromLatLng(s2.LatLngFromDegrees(51.51957, -0.09439))) {
		t.Error("Occlusion didn't contain expected point")
	}
}

func BenchmarkOcclude(b *testing.B) {
	center := s2.PointFromLatLng(s2.LatLngFromDegrees(51.53545, -0.12561))
	radius := b6.MetersToAngle(275.0)
	p0, p1 := s2.LatLngFromDegrees(51.5366452, -0.1270338), s2.LatLngFromDegrees(51.5366355, -0.1269504)
	e := [2]s2.Point{s2.PointFromLatLng(p0), s2.PointFromLatLng(p1)}
	for n := 0; n < b.N; n++ {
		for i := 0; i < 200; i++ {
			// Occlude is typically called ~200 times during a single sightline calculation
			occlude(e, center, radius)
		}
	}
}

func BenchmarkOccludeWithIndex(b *testing.B) {
	center := s2.PointFromLatLng(s2.LatLngFromDegrees(51.53545, -0.12561))
	radius := b6.MetersToAngle(275.0)
	boundary := s2.RegularLoop(center, radius, 128)
	index := s2.NewShapeIndex()
	index.Add(boundary)
	query := s2.NewContainsPointQuery(index, s2.VertexModelOpen)
	p0, p1 := s2.LatLngFromDegrees(51.5366452, -0.1270338), s2.LatLngFromDegrees(51.5366355, -0.1269504)
	e := [2]s2.Point{s2.PointFromLatLng(p0), s2.PointFromLatLng(p1)}
	for n := 0; n < b.N; n++ {
		for i := 0; i < 200; i++ {
			// Occlude is typically called ~200 times during a single sightline calculation
			occludeWithIndex(e, center, radius, boundary, query)
		}
	}
}
