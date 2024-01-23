package ingest

import (
	"math"
	"testing"

	"diagonal.works/b6"
	"github.com/golang/geo/s2"
)

func TestAddPoints(t *testing.T) {
	point1 := &GenericFeature{ID: FromOSMNodeID(6082053666).FeatureID(), Tags: []b6.Tag{{Key: "latlng", Value: b6.LatLng(s2.LatLngFromDegrees(51.5366467, -0.1263796))}}}
	point2 := &GenericFeature{ID: b6.FeatureID{Type: b6.FeatureTypePoint, Namespace: b6.NamespacePrivate, Value: 1}, Tags: []b6.Tag{{Key: "latlng", Value: b6.LatLng(s2.LatLngFromDegrees(51.5351906, -0.1245464))}}}
	add := AddFeatures([]Feature{point1, point2.Clone()})

	w := NewBasicMutableWorld()
	applied, err := add.Apply(w)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	ids := make(map[b6.FeatureID]b6.FeatureID)
	b6.FillMap(applied, ids)

	added := w.FindFeatureByID(point1.FeatureID())
	if added == nil || added.(b6.Geometry).Location().Distance(point1.Location()) > b6.MetersToAngle(1.0) {
		t.Error("Expected to find p2 under its given ID")
	}

	allocated, ok := ids[point2.FeatureID()]
	if !ok {
		t.Fatalf("Expected a new ID for %s", point2.FeatureID())
	}

	added = w.FindFeatureByID(allocated)
	if added == nil || added.(b6.Geometry).Location().Distance(point2.Location()) > b6.MetersToAngle(1.0) {
		t.Error("Expected to find p2 under a new ID")
	}
}

func TestAddPaths(t *testing.T) {
	point1 := &GenericFeature{ID: FromOSMNodeID(6082053666).FeatureID(), Tags: []b6.Tag{{Key: "latlng", Value: b6.LatLng(s2.LatLngFromDegrees(51.5366467, -0.1263796))}}}
	point2 := &GenericFeature{ID: b6.FeatureID{Type: b6.FeatureTypePoint, Namespace: b6.NamespacePrivate, Value: 1}, Tags: []b6.Tag{{Key: "latlng", Value: b6.LatLng(s2.LatLngFromDegrees(51.5351906, -0.1245464))}}}

	path := NewPathFeature(2)
	path.PathID = b6.MakePathID(b6.NamespacePrivate+"/1", 1)
	path.SetPointID(0, point1.FeatureID())
	path.SetPointID(1, point2.FeatureID())

	add := AddFeatures([]Feature{point2.Clone(), path.ClonePathFeature()})

	w := NewBasicMutableWorld()
	w.AddFeature(point1)

	applied, err := add.Apply(w)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	ids := make(map[b6.FeatureID]b6.FeatureID)
	b6.FillMap(applied, ids)

	allocated, ok := ids[point2.FeatureID()]
	if !ok {
		t.Fatalf("Expected a new ID for %s", point2.FeatureID())
	}

	addedPoint := w.FindFeatureByID(allocated)
	if addedPoint == nil || addedPoint.(b6.Geometry).Location().Distance(point2.Location()) > b6.MetersToAngle(1.0) {
		t.Error("Expected to find p2 under a new ID")
	}

	allocated, ok = ids[path.FeatureID()]
	if !ok {
		t.Fatalf("Expected a new ID for %s", path.FeatureID())
	}
	addedPath := b6.FindPathByID(allocated.ToPathID(), w)
	if addedPath == nil {
		t.Fatal("Expected to find path under a new ID")
	}
	if p := addedPath.Feature(1); p == nil || p.FeatureID() != ids[point2.FeatureID()] {
		t.Error("Expected path to reference newly generated ID")
	}
}

func TestAddAreas(t *testing.T) {
	p1 := &GenericFeature{ID: FromOSMNodeID(4270651271).FeatureID(), Tags: []b6.Tag{{Key: "latlng", Value: b6.LatLng(s2.LatLngFromDegrees(51.5354124, -0.1243817))}}}
	p2 := &GenericFeature{ID: FromOSMNodeID(5693730034).FeatureID(), Tags: []b6.Tag{{Key: "latlng", Value: b6.LatLng(s2.LatLngFromDegrees(51.5353117, -0.1244943))}}}
	p3 := &GenericFeature{ID: b6.FeatureID{Type: b6.FeatureTypePoint, Namespace: b6.NamespacePrivate, Value: 1}, Tags: []b6.Tag{{Key: "latlng", Value: b6.LatLng(s2.LatLngFromDegrees(51.5353736, -0.1242415))}}}

	path := NewPathFeature(4)
	path.PathID = b6.MakePathID(b6.NamespacePrivate+"/1", 1)
	path.SetPointID(0, p1.FeatureID())
	path.SetPointID(1, p2.FeatureID())
	path.SetPointID(2, p3.FeatureID())
	path.SetPointID(3, p1.FeatureID())

	area := NewAreaFeature(1)
	area.AreaID = b6.MakeAreaID(b6.NamespacePrivate+"/2", 1)
	area.SetPathIDs(0, []b6.PathID{path.PathID})

	add := AddFeatures([]Feature{p2.Clone(), p3.Clone(), path.ClonePathFeature(), area.CloneAreaFeature()})

	w := NewBasicMutableWorld()
	w.AddFeature(p1)

	applied, err := add.Apply(w)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	ids := make(map[b6.FeatureID]b6.FeatureID)
	b6.FillMap(applied, ids)

	allocated, ok := ids[area.FeatureID()]
	if !ok {
		t.Fatalf("Expected a new ID for %s", area.FeatureID())
	}
	added := b6.FindAreaByID(allocated.ToAreaID(), w)
	if added == nil {
		t.Fatal("Expected to find area under a new ID")
	}
	if added.Feature(0)[0].PathID().FeatureID() != ids[path.FeatureID()] {
		t.Error("Expected path to reference newly generated ID")
	}
}

func TestAddRelations(t *testing.T) {
	p1 := &GenericFeature{ID: FromOSMNodeID(6082053666).FeatureID(), Tags: []b6.Tag{{Key: "latlng", Value: b6.LatLng(s2.LatLngFromDegrees(51.5366467, -0.1263796))}}}
	p2 := &GenericFeature{ID: b6.FeatureID{Type: b6.FeatureTypePoint, Namespace: b6.NamespacePrivate, Value: 1}, Tags: []b6.Tag{{Key: "latlng", Value: b6.LatLng(s2.LatLngFromDegrees(51.5351906, -0.1245464))}}}

	relation := NewRelationFeature(2)
	relation.RelationID = b6.MakeRelationID(b6.NamespaceDiagonalAccessPoints, 1)
	relation.Members[0] = b6.RelationMember{ID: p1.FeatureID()}
	relation.Members[1] = b6.RelationMember{ID: p2.FeatureID()}

	add := AddFeatures([]Feature{p1.Clone(), p2.Clone(), relation.CloneRelationFeature()})

	w := NewBasicMutableWorld()
	w.AddFeature(p1)

	applied, err := add.Apply(w)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	ids := make(map[b6.FeatureID]b6.FeatureID)
	b6.FillMap(applied, ids)

	added := b6.FindRelationByID(relation.RelationID, w)
	if added == nil {
		t.Fatal("Expected to find relation under a new ID")
	}
	if added.Member(1).ID != ids[p2.FeatureID()] {
		t.Error("Expected relation member to reference newly generated ID")
	}
}

func TestAddCollections(t *testing.T) {
	collection := CollectionFeature{
		CollectionID: b6.MakeCollectionID(b6.NamespacePrivate, 1),
		Keys:         []interface{}{b6.PathID{b6.NamespaceDiagonalEntrances, 777}},
		Values:       []interface{}{"i dont need to be humble"},
	}

	add := AddFeatures([]Feature{&collection})

	w := NewBasicMutableWorld()

	_, err := add.Apply(w)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	added := b6.FindCollectionByID(collection.CollectionID, w)
	if added == nil {
		t.Fatalf("Expected to find collection under %s", collection.CollectionID.String())
	}
	if collection.Values[0].(string) != "i dont need to be humble" {
		t.Error("Expected collection member value to match")
	}
}

func TestMergeChanges(t *testing.T) {
	ns := b6.Namespace("diagonal.works/test")
	p1 := &GenericFeature{ID: b6.FeatureID{b6.FeatureTypePoint, ns, 1}, Tags: []b6.Tag{{Key: "latlng", Value: b6.LatLng(s2.LatLngFromDegrees(51.5366467, -0.1263796))}}}
	p2 := &GenericFeature{ID: b6.FeatureID{b6.FeatureTypePoint, ns, 2}, Tags: []b6.Tag{{Key: "latlng", Value: b6.LatLng(s2.LatLngFromDegrees(51.5351906, -0.1245464))}}}

	add1 := AddFeatures([]Feature{p1, p2})

	path := NewPathFeature(2)
	path.PathID = b6.MakePathID(ns, 3)
	path.SetPointID(0, p1.FeatureID())
	path.SetPointID(1, p2.FeatureID())
	add2 := AddFeatures([]Feature{path})

	merged := MergedChange{&add1, &add2}
	w := NewBasicMutableWorld()
	_, err := merged.Apply(w)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	found := b6.FindPathByID(path.PathID, w)
	if found == nil {
		t.Fatal("Expected to find added path")
	}

	expected := 200.0
	actual := b6.AngleToMeters(found.Polyline().Length())
	delta := math.Abs((actual - expected) / expected)
	if delta > 0.1 {
		t.Errorf("Length of added path outside expected bounds")
	}
}

func TestMergeChangesLeavesWorldUnmodfiedFollowingError(t *testing.T) {
	ns := b6.Namespace("diagonal.works/test")
	point := &GenericFeature{ID: b6.FeatureID{b6.FeatureTypePoint, ns, 1}, Tags: []b6.Tag{{Key: "latlng", Value: b6.LatLng(s2.LatLngFromDegrees(51.5366467, -0.1263796))}}}
	add1 := AddFeatures([]Feature{point})

	path := NewPathFeature(2)
	path.PathID = b6.MakePathID(ns, 3)
	path.SetPointID(0, b6.FeatureID{b6.FeatureTypePoint, b6.Namespace("nonexistant"), 0})
	path.SetPointID(1, b6.FeatureID{b6.FeatureTypePoint, b6.Namespace("nonexistant"), 1})
	add2 := AddFeatures([]Feature{path})

	merged := MergedChange{&add1, &add2}
	w := NewBasicMutableWorld()
	_, err := merged.Apply(w)
	if err == nil {
		t.Fatal("Expected an error, found none")
	}

	found := w.FindFeatureByID(point.FeatureID())
	if found != nil {
		t.Error("Expected world to be unchanged following failure")
	}
}
