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
	if w == nil {
		return
	}
	m := ingest.NewMutableOverlayWorld(w)

	origin := b6.FindAreaByID(camden.LightermanID, m)
	if origin == nil {
		t.Errorf("Failed to find origin")
		return
	}

	origins := &api.ArrayFeatureCollection{
		Features: []b6.Feature{origin},
	}

	c := api.Context{World: m}
	accessible, err := buildingAccess(origins, 1000, "walking", &c)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	count := 0
	i := accessible.Begin()
	for {
		if ok, err := i.Next(); err != nil {
			t.Errorf("Expected no error, found: %s", err)
			return
		} else if !ok {
			break
		}
		if k := i.Key().(b6.FeatureID); k != origin.FeatureID() {
			t.Errorf("Expected origin %s, found %s", origin.FeatureID(), k)
		}
		f := m.FindFeatureByID(i.Value().(b6.FeatureID))
		if b := f.Get("#building"); !b.IsValid() {
			t.Errorf("Expected a building")
		}
		count++
	}
	if count < 2 {
		t.Errorf("Expected at least two accessible buildings")
	}
}
