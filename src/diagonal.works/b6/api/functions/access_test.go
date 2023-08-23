package functions

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/testcamden"
)

func TestBuildingAccessibility(t *testing.T) {
	w := testcamden.BuildGranarySquare(t)
	m := ingest.NewMutableOverlayWorld(w)

	origin := b6.FindAreaByID(testcamden.LightermanID, m)
	if origin == nil {
		t.Fatalf("FindAreaByID(LightermanID) failed to find origin")
	}

	origins := &api.ArrayFeatureCollection{
		Features: []b6.Feature{origin},
	}

	c := api.Context{World: m}
	accessible, err := buildingAccess(&c, origins, 1000, "walking")
	if err != nil {
		t.Fatalf("buildingAccess() failed with: %v", err)
	}

	count := 0
	i := accessible.Begin()
	for {
		if ok, err := i.Next(); err != nil {
			t.Fatalf("i.Next() failed with: %v", err)
		} else if !ok {
			break
		}
		if k := i.Key().(b6.FeatureID); k != origin.FeatureID() {
			t.Errorf("i.Key() got %s, want %s", k, origin.FeatureID())
		}
		f := m.FindFeatureByID(i.Value().(b6.FeatureID))
		if b := f.Get("#building"); !b.IsValid() {
			t.Errorf("Expected a building")
		}
		count++
	}
	if count < 2 {
		t.Errorf("got at %d accessible buildings, want at least two", count)
	}
}
