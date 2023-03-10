package ingest

import (
	"testing"

	"diagonal.works/b6"
	"github.com/golang/geo/s2"
)

func TestIDFromLatLng(t *testing.T) {
	expected := s2.LatLngFromDegrees(51.5371371, -0.1240464)
	actual, ok := LatLngFromID(NewLatLngID(expected))

	if !ok {
		t.Errorf("Failed to convert world.FeatureID to s2.LatLng")
		return
	}
	if s2.PointFromLatLng(expected).Distance(s2.PointFromLatLng(actual)) > b6.MetersToAngle(0.01) {
		t.Errorf("Expected %v to be close to %v", expected, actual)
	}

	actual, ok = LatLngFromID(FromOSMNodeID(2309943870))
	if ok {
		t.Errorf("Incorrectly converted world.FeatureID to s2.LatLng")
	}
}
