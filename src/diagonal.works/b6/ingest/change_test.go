package ingest

import (
	"testing"

	"diagonal.works/b6"
	"github.com/golang/geo/s2"
)

func TestAddPoints(t *testing.T) {
	p1 := NewPointFeature(FromOSMNodeID(6082053666), s2.LatLngFromDegrees(51.5366467, -0.1263796))
	p2 := NewPointFeature(b6.MakePointID(b6.NamespacePrivate, 1), s2.LatLngFromDegrees(51.5351906, -0.1245464))

	add := AddFeatures{
		Points: []*PointFeature{p1, p2.ClonePointFeature()}, // Clone p2 as the ID is overwritten
		IDsToReplace: map[b6.Namespace]b6.Namespace{
			b6.NamespacePrivate: b6.NamespaceDiagonalEntrances,
		},
	}

	w := NewBasicMutableWorld()
	ids, err := add.Apply(w)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	added := b6.FindPointByID(p1.PointID, w)
	if added == nil || added.Point().Distance(s2.PointFromLatLng(p1.Location)) > b6.MetersToAngle(1.0) {
		t.Errorf("Expected to find p2 under its given ID")
	}

	if allocated, ok := ids[p2.FeatureID()]; ok {
		added := b6.FindPointByID(allocated.ToPointID(), w)
		if added == nil || added.Point().Distance(s2.PointFromLatLng(p2.Location)) > b6.MetersToAngle(1.0) {
			t.Errorf("Expected to find p2 under a new ID")
		}
	} else {
		t.Errorf("Expected a new ID for %s", p2.FeatureID())
		return
	}

}

func TestAddPaths(t *testing.T) {
	p1 := NewPointFeature(FromOSMNodeID(6082053666), s2.LatLngFromDegrees(51.5366467, -0.1263796))
	p2 := NewPointFeature(b6.MakePointID(b6.NamespacePrivate, 1), s2.LatLngFromDegrees(51.5351906, -0.1245464))

	path := NewPathFeature(2)
	path.PathID = b6.MakePathID(b6.NamespacePrivate+"/1", 1)
	path.SetPointID(0, p1.PointID)
	path.SetPointID(1, p2.PointID)

	add := AddFeatures{
		Points: []*PointFeature{p2.ClonePointFeature()}, // Clone features as the ID is overwritten
		Paths:  []*PathFeature{path.ClonePathFeature()},
		IDsToReplace: map[b6.Namespace]b6.Namespace{
			b6.NamespacePrivate:        b6.NamespaceDiagonalEntrances,
			b6.NamespacePrivate + "/1": b6.NamespaceDiagonalAccessPoints,
		},
	}

	w := NewBasicMutableWorld()
	w.AddPoint(p1)

	ids, err := add.Apply(w)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	if allocated, ok := ids[p2.FeatureID()]; ok {
		added := b6.FindPointByID(allocated.ToPointID(), w)
		if added == nil || added.Point().Distance(s2.PointFromLatLng(p2.Location)) > b6.MetersToAngle(1.0) {
			t.Errorf("Expected to find p2 under a new ID")
		}
	} else {
		t.Errorf("Expected a new ID for %s", p2.FeatureID())
		return
	}

	if allocated, ok := ids[path.FeatureID()]; ok {
		added := b6.FindPathByID(allocated.ToPathID(), w)
		if added == nil {
			t.Errorf("Expected to find path under a new ID")
		} else {
			if p := added.Feature(1); p == nil || p.FeatureID() != ids[p2.FeatureID()] {
				t.Errorf("Expected path to reference newly generated ID")
			}
		}
	} else {
		t.Errorf("Expected a new ID for %s", path.FeatureID())
		return
	}
}

func TestAddAreas(t *testing.T) {
	p1 := NewPointFeature(FromOSMNodeID(4270651271), s2.LatLngFromDegrees(51.5354124, -0.1243817))
	p2 := NewPointFeature(FromOSMNodeID(5693730034), s2.LatLngFromDegrees(51.5353117, -0.1244943))
	p3 := NewPointFeature(b6.MakePointID(b6.NamespacePrivate, 1), s2.LatLngFromDegrees(51.5353736, -0.1242415))

	path := NewPathFeature(4)
	path.PathID = b6.MakePathID(b6.NamespacePrivate+"/1", 1)
	path.SetPointID(0, p1.PointID)
	path.SetPointID(1, p2.PointID)
	path.SetPointID(2, p3.PointID)
	path.SetPointID(3, p1.PointID)

	area := NewAreaFeature(1)
	area.AreaID = b6.MakeAreaID(b6.NamespacePrivate+"/2", 1)
	area.SetPathIDs(0, []b6.PathID{path.PathID})

	add := AddFeatures{
		Points: []*PointFeature{p2.ClonePointFeature(), p3.ClonePointFeature()}, // Clone features as the ID is overwritten
		Paths:  []*PathFeature{path.ClonePathFeature()},
		Areas:  []*AreaFeature{area.CloneAreaFeature()},
		IDsToReplace: map[b6.Namespace]b6.Namespace{
			b6.NamespacePrivate:        b6.NamespaceDiagonalEntrances,
			b6.NamespacePrivate + "/1": b6.NamespaceDiagonalAccessPoints,
			b6.NamespacePrivate + "/2": b6.NamespaceUKONSBoundaries,
		},
	}

	w := NewBasicMutableWorld()
	w.AddPoint(p1)

	ids, err := add.Apply(w)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	if allocated, ok := ids[area.FeatureID()]; ok {
		added := b6.FindAreaByID(allocated.ToAreaID(), w)
		if added == nil {
			t.Errorf("Expected to find area under a new ID")
		} else {
			if added.Feature(0)[0].PathID().FeatureID() != ids[path.FeatureID()] {
				t.Errorf("Expected path to reference newly generated ID")
			}
		}
	} else {
		t.Errorf("Expected a new ID for %s", area.FeatureID())
		return
	}
}

func TestAddRelations(t *testing.T) {
	p1 := NewPointFeature(FromOSMNodeID(6082053666), s2.LatLngFromDegrees(51.5366467, -0.1263796))
	p2 := NewPointFeature(b6.MakePointID(b6.NamespacePrivate, 1), s2.LatLngFromDegrees(51.5351906, -0.1245464))

	relation := NewRelationFeature(2)
	relation.RelationID = b6.MakeRelationID(b6.NamespaceDiagonalAccessPoints, 1)
	relation.Members[0] = b6.RelationMember{ID: p1.FeatureID()}
	relation.Members[1] = b6.RelationMember{ID: p2.FeatureID()}

	add := AddFeatures{
		Points:    []*PointFeature{p1.ClonePointFeature(), p2.ClonePointFeature()}, // Clone features as the ID is overwritten
		Relations: []*RelationFeature{relation.CloneRelationFeature()},
		IDsToReplace: map[b6.Namespace]b6.Namespace{
			b6.NamespacePrivate: b6.NamespaceDiagonalEntrances,
		},
	}

	w := NewBasicMutableWorld()
	w.AddPoint(p1)

	ids, err := add.Apply(w)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	added := b6.FindRelationByID(relation.RelationID, w)
	if added == nil {
		t.Errorf("Expected to find relation under a new ID")
	} else {
		if added.Member(1).ID != ids[p2.FeatureID()] {
			t.Errorf("Expected relation member to reference newly generated ID")
		}
	}
}
