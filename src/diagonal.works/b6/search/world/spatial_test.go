package world

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/search"
	"diagonal.works/b6/test/camden"

	"github.com/golang/geo/s2"
)

func makeAreaMap(areas b6.AreaFeatures) map[b6.AreaID]b6.AreaFeature {
	m := make(map[b6.AreaID]b6.AreaFeature)
	for areas.Next() {
		m[areas.Feature().AreaID()] = areas.Feature()
	}
	return m
}

func makePointMap(points b6.PointFeatures) map[b6.PointID]b6.PointFeature {
	m := make(map[b6.PointID]b6.PointFeature)
	for points.Next() {
		m[points.Feature().PointID()] = points.Feature()
	}
	return m
}

func TestIntersectsWithCap(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.53617, -0.12582)), b6.MetersToAngle(100))
	roughAreas := makeAreaMap(b6.FindAreas(search.NewSpatialFromRegion(cap), granarySquare))
	exactAreas := makeAreaMap(b6.FindAreas(NewIntersectsCap(cap), granarySquare))

	if len(roughAreas) <= len(exactAreas) || len(exactAreas) == 0 {
		t.Errorf("Expected there to be less areas by exact match than rough match")
	}

	if _, ok := roughAreas[b6.MakeAreaID(b6.NamespaceOSMWay, uint64(camden.LightermanWay))]; !ok {
		t.Errorf("Expected rough areas to contain the Lighterman")
	}

	if _, ok := exactAreas[b6.MakeAreaID(b6.NamespaceOSMWay, uint64(camden.LightermanWay))]; ok {
		t.Errorf("Expected exact areas to not contain the Lighterman")
	}
}

func TestIntersectsWithAreaFeature(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	coalDropsYard := b6.FindAreaByID(b6.MakeAreaID(b6.NamespaceOSMWay, 222021572), granarySquare)
	if coalDropsYard == nil {
		t.Errorf("Failed to find Coal Drops Yard")
		return
	}

	points := makePointMap(b6.FindPoints(NewIntersectsFeature(coalDropsYard), granarySquare))
	if _, ok := points[b6.MakePointID(b6.NamespaceOSMNode, 6082053669)]; !ok {
		t.Errorf("Expected to find Outsiders Store")
	}

	if _, ok := points[b6.MakePointID(b6.NamespaceOSMNode, 6082053666)]; ok {
		t.Errorf("Didn't expect to find Vermuteria")
	}
}
