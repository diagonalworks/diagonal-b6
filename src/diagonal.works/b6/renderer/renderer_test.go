package renderer

import (
	"strconv"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/osm"
	"diagonal.works/b6/test/testcamden"
	"github.com/golang/geo/s2"
)

func TestFillColourFromFeature(t *testing.T) {
	tests := []struct {
		featureColour string
		ok            bool
		tileColour    string
	}{
		{"#ff0000", true, "#ff0000"},
		{"#ff000011", false, ""},
		{"#gg000011", false, ""},
		{"0.75", true, "#f87f51"},
		{"red", false, ""},
	}
	for _, test := range tests {
		tags := ingest.Tags{{Key: "diagonal:colour", Value: test.featureColour}}
		feature := NewFeature(&Point{})
		fillColourFromFeature(feature, tags)
		tileColour, ok := feature.Tags["colour"]
		if ok != test.ok {
			t.Errorf("Expected ok %v, found %v", test.ok, ok)
		} else if ok {
			if tileColour != test.tileColour {
				t.Errorf("Expected tile colour %q, found %q", test.tileColour, tileColour)
			}
		}
	}
}

func TestFeaturesHaveTagsForNamespaceAndID(t *testing.T) {
	granarySquare := testcamden.BuildGranarySquare(t)
	if granarySquare == nil {
		return
	}

	projection := b6.NewTileMercatorProjection(16)
	r := BasemapRenderer{RenderRules: BasemapRenderRules, World: granarySquare}
	tile, err := r.Render(projection.TileFromLatLng(s2.LatLngFromDegrees(51.53531, -0.12434)), &TileArgs{})
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
	}

	expected := "19813dd2"
	if id, _ := strconv.ParseUint(expected, 16, 64); osm.WayID(id) != testcamden.LightermanWay {
		t.Errorf("Unexpected hex ID for LightermanWay. Test data changed?")
		return
	}

	found := false
	if layer := tile.FindLayer("building"); layer != nil {
		for _, f := range layer.Features {
			if ns, ok := f.Tags["ns"]; ok && ns == b6.NamespaceOSMWay.String() {
				if id, ok := f.Tags["id"]; ok && id == expected {
					found = true
					break
				}
			}
		}
	} else {
		t.Errorf("Expected to find building layer")
	}
	if !found {
		t.Errorf("Expected Lighterman feature to be tagged with an ID")
	}
}

func TestFeaturesAreOrderedByLayerTag(t *testing.T) {
	granarySquare := testcamden.BuildGranarySquare(t)
	if granarySquare == nil {
		return
	}
	mutable := ingest.NewMutableOverlayWorld(granarySquare)

	// Add a roof terrace and second floor to the Lighterman
	lighterman := b6.FindAreaByID(testcamden.LightermanID, mutable)
	roof := ingest.NewAreaFeatureFromWorld(lighterman)
	roof.AreaID = b6.MakeAreaID(b6.NamespacePrivate, 1)
	roof.AddTag(b6.Tag{Key: "layer", Value: "2"})
	if err := mutable.AddArea(roof); err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}
	basement := ingest.NewAreaFeatureFromWorld(lighterman)
	basement.AreaID = b6.MakeAreaID(b6.NamespacePrivate, 2)
	basement.AddTag(b6.Tag{Key: "layer", Value: "-1"})
	if err := mutable.AddArea(basement); err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	projection := b6.NewTileMercatorProjection(16)
	r := BasemapRenderer{RenderRules: BasemapRenderRules, World: mutable}
	tile, err := r.Render(projection.TileFromLatLng(s2.LatLngFromDegrees(51.53531, -0.12434)), &TileArgs{})
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
	}

	order := []b6.FeatureID{basement.FeatureID(), lighterman.FeatureID(), roof.FeatureID()}
	if layer := tile.FindLayer("building"); layer != nil {
		next := 0
		for _, f := range layer.Features {
			if f.ID == api.TileFeatureIDForPolygon(order[next], 0) {
				next++
			}
		}
		if next != len(order) {
			t.Errorf("Unexpected feature order while searching for %s", order[next])
		}
	} else {
		t.Errorf("Expected to find building layer")
	}
}
