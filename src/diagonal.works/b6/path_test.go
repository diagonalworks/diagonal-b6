package b6

import (
	"math"
	"testing"

	"diagonal.works/b6/units"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

func TestSamplePoints(t *testing.T) {
	path := PathFromS2Points([]s2.Point{
		s2.PointFromLatLng(s2.LatLngFromDegrees(51.535317, -0.125961)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(51.535364, -0.1260701)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(51.535407, -0.126080)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(51.537327, -0.125291)),
	})
	spacing := units.MetersToAngle(1.0)
	sampled := SamplePoints(path, spacing, []s2.Point{})

	l := units.AngleToMeters(path.Polyline().Length())
	if expected := int(math.Floor(l)) + 1; len(sampled) != expected {
		t.Errorf("Expected %d points, found %d", expected, len(sampled))
	}

	// Roughly 10cm. The test path isn't straight, so we need some tolerance.
	const epsilon s1.Angle = 1.6e-08
	for i := 0; i < len(sampled)-2; i++ {
		if d := sampled[i].Distance(sampled[i+1]); math.Abs(float64(d-spacing)) > float64(epsilon) {
			t.Errorf("Expected distance %fm, found %fm at point %d", AngleToMeters(spacing), AngleToMeters(d), i)
		}
	}
}
