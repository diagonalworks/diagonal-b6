package functions

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"
)

func TestBuildingAccessibility(t *testing.T) {
	w := camden.BuildGranarySquareForTests(t)
	m := ingest.NewMutableOverlayWorld(w)

	origin := b6.FindAreaByID(camden.LightermanID, m)
	if origin == nil {
		t.Fatal("Failed to find origin")
	}

	origins := &api.ArrayFeatureCollection{
		Features: []b6.Feature{origin},
	}

	c := api.Context{World: m}
	accessible, err := buildingAccess(&c, origins, 1000, "walking")
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	count := 0
	i := accessible.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			t.Fatalf("Expected no error, found: %s", err)
		}
		if !ok {
			break
		}
		if k := i.Key().(b6.FeatureID); k != origin.FeatureID() {
			t.Errorf("Expected origin %s, found %s", origin.FeatureID(), k)
		}
		f := m.FindFeatureByID(i.Value().(b6.FeatureID))
		if b := f.Get("#building"); !b.IsValid() {
			t.Error("Expected a building")
		}
		count++
	}
	if count < 2 {
		t.Error("Expected at least two accessible buildings")
	}
}
