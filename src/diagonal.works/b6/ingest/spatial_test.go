package ingest

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/osm"
	"diagonal.works/b6/test"

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
	nodes, ways, relations, err := osm.ReadWholePBF(test.Data(test.GranarySquarePBF))
	o := BuildOptions{Cores: 2}
	w, err := BuildWorldFromOSM(nodes, ways, relations, &o)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.53617, -0.12582)), b6.MetersToAngle(100))
	roughAreas := makeAreaMap(b6.FindAreas(b6.MightIntersect{cap}, w))
	exactAreas := makeAreaMap(b6.FindAreas(b6.NewIntersectsCap(cap), w))

	if len(roughAreas) <= len(exactAreas) || len(exactAreas) == 0 {
		t.Errorf("Expected there to be less areas by exact match than rough match")
	}

	lighterman := osm.WayID(427900370)
	if _, ok := roughAreas[b6.MakeAreaID(b6.NamespaceOSMWay, uint64(lighterman))]; !ok {
		t.Errorf("Expected rough areas to contain the Lighterman")
	}

	if _, ok := exactAreas[b6.MakeAreaID(b6.NamespaceOSMWay, uint64(lighterman))]; ok {
		t.Errorf("Expected exact areas to not contain the Lighterman")
	}
}

func TestIntersectsWithAreaFeature(t *testing.T) {
	nodes, ways, relations, err := osm.ReadWholePBF(test.Data(test.GranarySquarePBF))
	o := BuildOptions{Cores: 2}
	w, err := BuildWorldFromOSM(nodes, ways, relations, &o)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	coalDropsYard := b6.FindAreaByID(b6.MakeAreaID(b6.NamespaceOSMWay, 222021572), w)
	if coalDropsYard == nil {
		t.Errorf("Failed to find Coal Drops Yard")
		return
	}

	points := makePointMap(b6.FindPoints(b6.IntersectsFeature{coalDropsYard.FeatureID()}, w))
	if _, ok := points[b6.MakePointID(b6.NamespaceOSMNode, 6082053669)]; !ok {
		t.Errorf("Expected to find Outsiders Store")
	}

	if _, ok := points[b6.MakePointID(b6.NamespaceOSMNode, 6082053666)]; ok {
		t.Errorf("Didn't expect to find Vermuteria")
	}
}
