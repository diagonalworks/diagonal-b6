package functions

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/test/camden"

	"github.com/golang/geo/s2"
)

func TestS2Points(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	context := &api.Context{
		World: granarySquare,
	}

	area := b6.FindAreaByID(camden.GranarySquareID, granarySquare)
	if area == nil {
		t.Fatal("Failed to find Granary Square")
	}

	center := s2.PointFromLatLng(s2.LatLngFromDegrees(51.53536, -0.12539))
	points, err := s2Points(context, area, 21, 21)
	if err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

	count := 0
	maxDistance := 0.0
	i := points.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			t.Fatalf("Expected no error, found %s", err)
		}
		if !ok {
			break
		}
		count++
		if d := b6.AngleToMeters(center.Distance(i.Value().(b6.Point).Point())); d > maxDistance {
			maxDistance = d
		}
	}

	if count < 400 || count > 500 {
		t.Errorf("Number of points outside expected range: %d", count)
	}
	if maxDistance < 50.0 || maxDistance > 70.0 {
		t.Errorf("Maximum point distance from the center of Granary Square outside expected range: %fm", maxDistance)
	}
}

func TestS2Grid(t *testing.T) {
	context := &api.Context{}
	topLeft := b6.PointFromLatLngDegrees(51.5146, -0.1140)
	bottomRight := b6.PointFromLatLngDegrees(51.5124, -0.0951)
	rectangle, _ := rectanglePolygon(context, topLeft, bottomRight)

	grid, err := s2Grid(context, rectangle, 21)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	bounds := rectangle.Polygon(0).RectBound()

	i := grid.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			t.Fatalf("Expected no error, found: %s", err)
		}
		if !ok {
			break
		}
		cellID := s2.CellIDFromToken(i.Value().(string))
		if cellID.Level() != 21 {
			t.Fatalf("Expected cell level 21, found: %d", cellID.Level())
		}
		if !s2.CellFromCellID(cellID).RectBound().Intersects(bounds) {
			t.Fatal("Expected cell to intersect rectangle bounds")
		}
	}
}

func TestS2Covering(t *testing.T) {
	context := &api.Context{}
	topLeft := b6.PointFromLatLngDegrees(51.5146, -0.1140)
	bottomRight := b6.PointFromLatLngDegrees(51.5124, -0.0951)
	rectangle, _ := rectanglePolygon(context, topLeft, bottomRight)

	covering, err := s2Covering(context, rectangle, 1, 21)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	bounds := rectangle.Polygon(0).RectBound()

	i := covering.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			t.Fatalf("Expected no error, found: %s", err)
		}
		if !ok {
			break
		}
		cellID := s2.CellIDFromToken(i.Value().(string))
		if !s2.CellFromCellID(cellID).RectBound().Intersects(bounds) {
			t.Fatal("Expected cell to intersect rectangle bounds")
		}
	}
}
