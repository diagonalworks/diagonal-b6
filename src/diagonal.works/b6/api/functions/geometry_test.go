package functions

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"
	"github.com/golang/geo/s2"
)

func TestCollectPolygons(t *testing.T) {
	r := b6.MetersToAngle(300)
	ll1 := s2.LatLngFromDegrees(51.535239, -0.124416)
	p1 := b6.AreaFromS2Loop(s2.RegularLoop(s2.PointFromLatLng(ll1), r, 128))
	ll2 := s2.LatLngFromDegrees(51.536631, -0.126495)
	p2 := b6.AreaFromS2Loop(s2.RegularLoop(s2.PointFromLatLng(ll2), r, 128))
	c := b6.ArrayValuesCollection[b6.Area]{p1, p2}

	collected, err := collectAreas(nil, c.Collection().Values())
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	ll3 := s2.LatLngFromDegrees(51.536255, -0.126154)
	for _, ll := range []s2.LatLng{ll1, ll2, ll3} {
		if !collected.MultiPolygon().ContainsPoint(s2.PointFromLatLng(ll)) {
			t.Errorf("Expected collected areas to contain %s", ll)
		}
	}
}

func TestDistanceToPointMeters(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	context := &api.Context{
		World: granarySquare,
	}
	path := b6.FindPathByID(ingest.FromOSMWayID(377974549), granarySquare)
	if path == nil {
		t.Fatal("Failed to find expected path")
	}

	point := b6.GeometryFromLatLng(s2.LatLngFromDegrees(51.53586, -0.12564))
	distance, err := distanceToPointMeters(context, path, point)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	baseline := b6.AngleToMeters(path.Point(0).Distance(s2.PointFromLatLng(point.Location())))
	if baseline/distance > 1.5 {
		t.Errorf("Distances aren't similar enough; ratio: %f", baseline/distance)
	}
}
