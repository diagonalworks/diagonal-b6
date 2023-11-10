package api

import (
	"fmt"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"
	"github.com/google/go-cmp/cmp"
)

func TestMaterialiseFeatureIntCollection(t *testing.T) {
	c := b6.ArrayCollection[b6.FeatureID, int]{
		Keys: []b6.FeatureID{
			camden.VermuteriaID.FeatureID(),
			camden.LightermanID.FeatureID(),
		},
		Values: []int{
			36,
			42,
		},
	}
	r, err := Materialise(b6.MakeRelationID(b6.Namespace("diagonal.works/test"), 1), c.Collection())
	if err != nil {
		t.Fatalf("Expected no error from Materialise, found: %s", err)
	}
	dematerialised, err := Dematerialise(ingest.WrapRelationFeature(r, b6.EmptyWorld{}))
	if err != nil {
		t.Fatalf("Expected no error from Dematerialise, found: %s", err)
	}

	if cc, ok := dematerialised.(b6.UntypedCollection); ok {
		if diff := DiffCollections(c.Collection(), cc); diff != "" {
			t.Errorf(diff)
		}
	} else {
		t.Errorf("Expected Dematerialise to return a collection, found: %T", dematerialised)
	}
}

func TestMaterialiseFeatureFloatCollection(t *testing.T) {
	c := b6.ArrayCollection[b6.FeatureID, float64]{
		Keys: []b6.FeatureID{
			camden.VermuteriaID.FeatureID(),
			camden.LightermanID.FeatureID(),
		},
		Values: []float64{
			36.0,
			42.0,
		},
	}
	r, err := Materialise(b6.MakeRelationID(b6.Namespace("diagonal.works/test"), 1), c.Collection())
	if err != nil {
		t.Fatalf("Expected no error from Materialise, found: %s", err)
	}
	dematerialised, err := Dematerialise(ingest.WrapRelationFeature(r, b6.EmptyWorld{}))
	if err != nil {
		t.Fatalf("Expected no error from Dematerialise, found: %s", err)
	}

	if cc, ok := dematerialised.(b6.UntypedCollection); ok {
		if diff := DiffCollections(c.Collection(), cc); diff != "" {
			t.Errorf(diff)
		}
	} else {
		t.Errorf("Expected Dematerialise to return a collection, found: %T", dematerialised)
	}
}

func TestMaterialiseFeatureFeatureCollection(t *testing.T) {
	c := b6.ArrayCollection[b6.FeatureID, b6.FeatureID]{
		Keys: []b6.FeatureID{
			camden.VermuteriaID.FeatureID(),
			camden.LightermanID.FeatureID(),
		},
		Values: []b6.FeatureID{
			camden.GranarySquareID.FeatureID(),
			camden.StableStreetBridgeID.FeatureID(),
		},
	}
	r, err := Materialise(b6.MakeRelationID(b6.Namespace("diagonal.works/test"), 1), c.Collection())
	if err != nil {
		t.Fatalf("Expected no error from Materialise, found: %s", err)
	}
	dematerialised, err := Dematerialise(ingest.WrapRelationFeature(r, b6.EmptyWorld{}))
	if err != nil {
		t.Fatalf("Expected no error from Dematerialise, found: %s", err)
	}

	if len(r.Members) != 4 {
		t.Errorf("Expected both keys and values to be encoded as members")
	}

	if cc, ok := dematerialised.(b6.UntypedCollection); ok {
		if diff := DiffCollections(c.Collection(), cc); diff != "" {
			t.Errorf(diff)
		}
	} else {
		t.Errorf("Expected Dematerialise to return a collection, found: %T", dematerialised)
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
