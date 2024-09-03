package ingest

import (
	"encoding/json"
	"os"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/test"
)

func TestColourAreas(t *testing.T) {
	// colour-areas.geojson defines 4 features, such that:
	// - feature 0 shares vertices with feature 1
	// - feature 1 shares vertices with feature 2
	// - feature 3 shares no vertices with any shape
	f, err := os.Open(test.Data("colour-areas.geojson"))
	if err != nil {
		t.Fatalf("expected no error, found: %s", err)
	}

	var collection geojson.FeatureCollection
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&collection); err != nil {
		t.Fatalf("expected no error, found: %s", err)
	}
	f.Close()

	add := &AddFeatures{}
	ns := b6.Namespace("diagonal.works/test")
	add.FillFromGeoJSON(&collection, ns)
	w := NewBasicMutableWorld()
	add.Apply(w)

	source, err := ColourAreas(WorldFeatureSource{World: w}, 2)
	if err != nil {
		t.Fatalf("expected no error, found: %s", err)
	}

	coloured, err := NewMutableWorldFromSource(&BuildOptions{Cores: 2}, source)
	if err != nil {
		t.Fatalf("expected no error, found: %s", err)
	}

	colours := make([]string, 4)
	for i := range colours {
		a := b6.FindAreaByID(b6.AreaID{Namespace: ns, Value: uint64(i)}, coloured)
		if v := a.Get("b6:colour"); v.IsValid() {
			if s, ok := v.Value.AnyExpression.(b6.StringExpression); ok {
				colours[i] = s.String()
			} else {
				t.Errorf("expected a string value for the b6:colour tag")
			}
		} else {
			t.Errorf("expected to find b6:colour tag")
		}
	}

	if colours[0] == colours[1] || colours[1] == colours[2] {
		t.Errorf("expected neighbouring areas to be coloured differently")
	}
	if colours[3] != "0" {
		t.Errorf("expected disconnected areas to have colour index 0")
	}
}
