package b6

import (
	"math"
	"testing"

	"diagonal.works/b6/units"
	"github.com/golang/geo/s2"
)

func TestElevationField(t *testing.T) {
	e := ElevationField{Radius: units.MetersToAngle(500)}
	e.Add(s2.LatLngFromDegrees(55.9913178, -3.4858895), 28.9)
	e.Add(s2.LatLngFromDegrees(55.9895487, -3.4857895), 39.8)
	e.Add(s2.LatLngFromDegrees(55.9895694, -3.4825841), 50.8)
	e.Add(s2.LatLngFromDegrees(55.9914559, -3.4826242), 34.3)
	e.Finish()

	actual := 54.0
	estimated, ok := e.Elevation(s2.PointFromLatLng(s2.LatLngFromDegrees(55.9905299, -3.4841599)))
	if ok {
		if d := math.Abs((estimated - actual) / actual); d > 0.3 {
			t.Errorf("Expected a delta of less than 0.3, found %f", d)
		}
	} else {
		t.Errorf("Expected to find an elevation")
	}
}
