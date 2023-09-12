package b6

import (
	"testing"

	"github.com/golang/geo/s2"
)

// Additional tests that depend on real data are in
// ingest/spatial_test.go

func TestCapIntersectsPolygon(t *testing.T) {
	f := func(cap s2.Cap, p *s2.Polygon) bool {
		return CapIntersectsPolygon(cap, p)
	}
	ValidateCapPolygonIntersection(f, t)
}

func TestCapIntersectsPolygonViaQuery(t *testing.T) {
	f := func(cap s2.Cap, p *s2.Polygon) bool {
		q := NewIntersectsCap(cap)
		return q.IntersectsPolygon(p)
	}
	ValidateCapPolygonIntersection(f, t)
}

func ValidateCapPolygonIntersection(f func(cap s2.Cap, p *s2.Polygon) bool, t *testing.T) {
	loop := s2.LoopFromPoints([]s2.Point{
		s2.PointFromLatLng(s2.LatLngFromDegrees(51.535623, -0.125801)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(51.535401, -0.125887)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(51.535245, -0.124957)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(51.535447, -0.124826)),
	})
	p := s2.PolygonFromLoops([]*s2.Loop{loop})

	tests := []struct {
		Center     s2.LatLng
		Radius     float64
		Intersects bool
	}{
		{s2.LatLngFromDegrees(51.535437, -0.125363), 15.0, true},
		{s2.LatLngFromDegrees(51.535437, -0.125363), 100.0, true},
		{s2.LatLngFromDegrees(51.535437, -0.125363), 1.0, true},
		{s2.LatLngFromDegrees(51.535269, -0.124520), 1.0, false},
		{s2.LatLngFromDegrees(51.535269, -0.124520), 30.0, true},
	}
	for _, test := range tests {
		cap := s2.CapFromCenterAngle(s2.PointFromLatLng(test.Center), MetersToAngle(test.Radius))
		if CapIntersectsPolygon(cap, p) != test.Intersects {
			t.Errorf("Unexpected intersection for cap center %s radius %.2f", test.Center, test.Radius)
		}
	}
}

func benchmarkCapIntersectsPolygon(points int, b *testing.B) {
	center := s2.LatLngFromDegrees(51.535437, -0.125363)
	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(center), MetersToAngle(10))
	p := s2.PolygonFromLoops([]*s2.Loop{s2.RegularLoop(s2.PointFromLatLng(center), MetersToAngle(20), points)})
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		CapIntersectsPolygon(cap, p)
	}
}

func BenchmarkCapIntersectsPolygon8(b *testing.B) {
	benchmarkCapIntersectsPolygon(8, b)
}

func BenchmarkCapIntersectsPolygon16(b *testing.B) {
	benchmarkCapIntersectsPolygon(16, b)
}

func BenchmarkCapIntersectsPolygon32(b *testing.B) {
	benchmarkCapIntersectsPolygon(32, b)
}

func BenchmarkCapIntersectsPolygon256(b *testing.B) {
	benchmarkCapIntersectsPolygon(256, b)
}

func benchmarkCapIntersectsPolygonViaQuery(points int, b *testing.B) {
	center := s2.LatLngFromDegrees(51.535437, -0.125363)
	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(center), MetersToAngle(10))
	q := NewIntersectsCap(cap)
	p := s2.PolygonFromLoops([]*s2.Loop{s2.RegularLoop(s2.PointFromLatLng(center), MetersToAngle(20), points)})
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		q.IntersectsPolygon(p)
	}
}

func BenchmarkCapIntersectsPolygonViaQuery8(b *testing.B) {
	benchmarkCapIntersectsPolygonViaQuery(16, b)
}

func BenchmarkCapIntersectsPolygonViaQuery16(b *testing.B) {
	benchmarkCapIntersectsPolygonViaQuery(16, b)
}

func BenchmarkCapIntersectsPolygonViaQuery32(b *testing.B) {
	benchmarkCapIntersectsPolygonViaQuery(32, b)
}

func BenchmarkCapIntersectsPolygonViaQuery256(b *testing.B) {
	benchmarkCapIntersectsPolygonViaQuery(256, b)
}
