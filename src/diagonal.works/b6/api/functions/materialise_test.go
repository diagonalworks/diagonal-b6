package functions

import (
	"fmt"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"
	"github.com/google/go-cmp/cmp"
)

func TestMaterialiseFeatureIntCollection(t *testing.T) {
	expected := b6.ArrayCollection[b6.FeatureID, int]{
		Keys: []b6.FeatureID{
			camden.VermuteriaID,
			camden.LightermanID.FeatureID(),
		},
		Values: []int{
			36,
			42,
		},
	}

	id := b6.MakeCollectionID("diagonal.works/test", 1)
	lambda := b6.NewLambdaExpression(
		[]string{},
		b6.NewCollectionExpression(expected.Collection()),
	)

	e := b6.NewCallExpression(
		b6.NewSymbolExpression("materialise"),
		[]b6.Expression{
			b6.NewFeatureIDExpression(id.FeatureID()),
			lambda,
		},
	)

	w := ingest.NewBasicMutableWorld()
	result, err := api.Evaluate(e, NewContext(w))
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	if change, ok := result.(ingest.Change); ok {
		if _, err := change.Apply(w); err != nil {
			t.Fatalf("Expected no error, found: %s", err)
		}
	} else {
		t.Fatalf("Expected an ingest.Change, found: %T", change)
	}

	if cc := b6.FindCollectionByID(id, w); cc != nil {
		if diff := DiffCollections(expected.Collection(), cc); diff != "" {
			t.Errorf(diff)
		}
	} else {
		t.Errorf("Expected to find a collection")
	}

	if ee := b6.FindExpressionByID(b6.MakeExpressionID(id.Namespace, id.Value), w); ee != nil {
		if diff := cmp.Diff(lambda, ee.Expression()); diff != "" {
			t.Errorf(diff)
		}
	} else {
		t.Errorf("Expected to find an expression")
	}
}

func DiffCollections(a b6.UntypedCollection, b b6.UntypedCollection) string {
	ai := a.BeginUntyped()
	bi := b.BeginUntyped()

	diffs := ""
	for {
		ok, err := ai.Next()
		if err != nil {
			return fmt.Sprintf("expected no error from a, found: %s", err)
		}
		if !ok {
			ok, _ = bi.Next()
			if ok {
				diffs += "expected collection to end\n"
			}
			return diffs
		}
		ok, err = bi.Next()
		if err != nil {
			return fmt.Sprintf("expected no error from b, found: %s", err)
		}
		if !ok {
			diffs += "expected collection to continue\n"
			return diffs
		}
		diffs += cmp.Diff(ai.Key(), bi.Key())
		diffs += cmp.Diff(ai.Value(), bi.Value())
	}
}

func TestMaterialiseMap(t *testing.T) {
	w := camden.BuildGranarySquareForTests(t)

	id := b6.MakeCollectionID("diagonal.works/test", 0)
	e := b6.NewCallExpression(
		b6.NewSymbolExpression("materialise-map"),
		[]b6.Expression{
			b6.NewCallExpression(
				b6.NewSymbolExpression("find"),
				[]b6.Expression{b6.NewQueryExpression(b6.Keyed{Key: "#building"})},
			),
			b6.NewFeatureIDExpression(id.FeatureID()),
			b6.NewSymbolExpression("all-tags"),
		},
	)

	change, err := api.Evaluate(e, NewContext(w))
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	mutable := ingest.NewMutableOverlayWorld(w)
	if _, err := change.(ingest.Change).Apply(mutable); err != nil {
		t.Fatalf("Expected no error applying change, found: %s", err)
	}

	materialised := b6.FindCollectionByID(id, mutable)
	if materialised == nil {
		t.Fatalf("Failed to find materialsed collection")
	}

	entries := make(map[b6.FeatureID]b6.CollectionID)
	if err := api.FillMap(materialised, entries); err != nil {
		t.Fatalf("Failed to fill map: %s", err)
	}

	if lighterman, ok := entries[camden.LightermanID.FeatureID()]; ok {
		c := b6.FindCollectionByID(lighterman, mutable)
		if c == nil {
			t.Fatalf("No collection for id %s", lighterman)
		}
		if tags, err := b6.AdaptCollection[int, b6.Tag](c).AllValues(nil); err != nil {
			t.Fatalf("Failed to fill values: %s", err)
		} else {
			if website := b6.Tags(tags).Get("website"); website.Value.String() != "https://thelighterman.co.uk/" {
				t.Errorf("Failed to find expected website tag in %s", tags)
			}
		}
	} else {
		t.Fatalf("No materialised collection for %s", camden.LightermanID)
	}
}

func TestMaterialiseMapMergesExistingCollectionItems(t *testing.T) {
	w := camden.BuildGranarySquareForTests(t)

	mutable := ingest.NewMutableOverlayWorld(w)

	id := b6.MakeCollectionID("diagonal.works/test", 0)
	collection := &ingest.CollectionFeature{
		CollectionID: id,
		Keys:         []interface{}{camden.StableStreetBridgeID.FeatureID()},
		Values:       []interface{}{b6.MakeCollectionID("diagonal.works/test", 1)},
	}
	if err := mutable.AddFeature(collection); err != nil {
		t.Fatalf("Failed to add feature: %s", err)
	}

	e := b6.NewCallExpression(
		b6.NewSymbolExpression("materialise-map"),
		[]b6.Expression{
			b6.NewCallExpression(
				b6.NewSymbolExpression("find"),
				[]b6.Expression{b6.NewQueryExpression(b6.Keyed{Key: "#building"})},
			),
			b6.NewFeatureIDExpression(id.FeatureID()),
			b6.NewSymbolExpression("all-tags"),
		},
	)

	change, err := api.Evaluate(e, NewContext(mutable))
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	if _, err := change.(ingest.Change).Apply(mutable); err != nil {
		t.Fatalf("Expected no error applying change, found: %s", err)
	}
	materialised := b6.FindCollectionByID(id, mutable)
	if materialised == nil {
		t.Fatalf("Failed to find materialsed collection")
	}

	entries := make(map[b6.FeatureID]b6.CollectionID)
	if err := api.FillMap(materialised, entries); err != nil {
		t.Fatalf("Failed to fill map: %s", err)
	}
	// The materialised collection should contain Stable Street,
	// because although it's not a building (and so wan't matched by)
	// the expression we evaluated, it was present in the existing
	// collection.
	if _, ok := entries[camden.StableStreetBridgeID.FeatureID()]; !ok {
		t.Errorf("Failed to find entry for %s", camden.StableStreetBridgeID.FeatureID())
	}
}
