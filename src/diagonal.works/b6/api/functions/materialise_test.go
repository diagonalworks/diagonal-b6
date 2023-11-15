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
			camden.VermuteriaID.FeatureID(),
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
