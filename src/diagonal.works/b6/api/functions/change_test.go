package functions

import (
	"testing"

	"diagonal.works/b6"
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

	tags := b6.ArrayValuesCollection[b6.Tag]{
		{Key: "#route", Value: b6.StringExpression("bicycle")},
	}

	members := &b6.ArrayCollection[b6.Identifiable, string]{
		Keys:   []b6.Identifiable{ingest.FromOSMWayID(4262451).FeatureID()},
		Values: []string{"forward"},
	}

	change, err := addRelation(nil, id, tags.Collection().Values(), members.Collection())
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

func TestAddCollection(t *testing.T) {
	m := ingest.NewMutableOverlayWorld(b6.EmptyWorld{})
	id := b6.MakeCollectionID("diagonal.works/test", 1)

	tags := b6.AdaptCollection[any, b6.Tag](
		b6.ArrayValuesCollection[b6.Tag]{
			{Key: "#route", Value: b6.StringExpression("bicycle")},
		}.Collection(),
	)

	collection := &b6.ArrayCollection[any, any]{
		Keys:   []interface{}{ingest.FromOSMWayID(4262451)},
		Values: []interface{}{"forward"},
	}

	change, err := addCollection(nil, id, tags, collection.Collection())
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	if _, err := change.Apply(m); err != nil {
		t.Fatalf("Expected no error applying change, found: %s", err)
	}

	if collection := b6.FindCollectionByID(id, m); collection == nil {
		t.Errorf("Expected to find added collection, found none")
	}
}

func TestAddExpression(t *testing.T) {
	m := ingest.NewMutableOverlayWorld(b6.EmptyWorld{})
	id := b6.MakeExpressionID("diagonal.works/test", 1)

	tags := b6.AdaptCollection[any, b6.Tag](
		b6.ArrayValuesCollection[b6.Tag]{
			{Key: "b6", Value: b6.StringExpression("docked")},
		}.Collection(),
	)

	expression := b6.NewCallExpression(
		b6.NewSymbolExpression("add"),
		[]b6.Expression{
			b6.NewIntExpression(10),
			b6.NewIntExpression(20),
		},
	)

	change, err := addExpression(nil, id, tags, expression)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	if _, err := change.Apply(m); err != nil {
		t.Fatalf("Expected no error applying change, found: %s", err)
	}

	if collection := b6.FindExpressionByID(id, m); collection == nil {
		t.Errorf("Expected to find added expression, found none")
	}
}
