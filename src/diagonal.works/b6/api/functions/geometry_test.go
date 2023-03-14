package functions

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"
)

func TestDistanceToPointMeters(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	context := &api.Context{
		World: granarySquare,
	}
	path := b6.FindPathByID(ingest.FromOSMWayID(377974549), granarySquare)
	if path == nil {
		t.Errorf("Failed to find expected path")
	}

	point := b6.PointFromLatLngDegrees(51.53586, -0.12564)
	distance, err := distanceToPointMeters(path, point, context)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
	}

	baseline := b6.AngleToMeters(path.Point(0).Distance(point.Point()))
	if baseline/distance > 1.5 {
		t.Errorf("Distances aren't similar enough; ratio: %f", baseline/distance)
	}
}
