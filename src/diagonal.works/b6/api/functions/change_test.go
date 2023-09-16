package functions

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
)

func TestIDToRelationID(t *testing.T) {
	id := idToRelationID(nil, "diagonal.works/accessability", ingest.FromOSMNodeID(6082053666))
	if id.Type != b6.FeatureTypeRelation {
		t.Errorf("Expected FeatureTypeRelation, found %s", id.Type)
	}
	if id.Namespace != b6.Namespace("diagonal.works/accessability") {
		t.Errorf("Unexpected namespace: %s", id.Namespace)
	}
}

func TestAddRelation(t *testing.T) {
	m := ingest.NewMutableOverlayWorld(b6.EmptyWorld{})
	id := b6.MakeRelationID("diagonal.works/test", 1)

	tags := api.ArrayTagCollection{
		Tags: []b6.Tag{{Key: "#route", Value: "bicycle"}},
	}

	members := api.ArrayAnyCollection{
		Keys:   []interface{}{ingest.FromOSMWayID(4262451)},
		Values: []interface{}{"forward"},
	}

	change, err := addRelation(nil, id, &tags, &members)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	if _, err := change.Apply(m); err != nil {
		t.Fatalf("Expected no error applying change, found: %s", err)
	}
	f := b6.FindRelationByID(id, m)
	if f == nil {
		t.Errorf("Expected to find added relation, found none")
	}
}
