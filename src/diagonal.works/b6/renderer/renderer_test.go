package renderer

import (
	"strconv"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/osm"
	"diagonal.works/b6/test/camden"
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
		tags := b6.Tags{{Key: "diagonal:colour", Value: b6.String(test.featureColour)}}
		feature := NewFeature(&Point{})
		fillColourFromFeature(feature, tags)
		tileColour, ok := feature.Tags["colour"]
		if ok != test.ok {
			t.Errorf("Expected ok %v, found %v", test.ok, ok)
		} else if ok && tileColour != test.tileColour {
			t.Errorf("Expected tile colour %q, found %q", test.tileColour, tileColour)
		}
	}
}

func TestFeaturesHaveTagsForNamespaceAndID(t *testing.T) {
	w := &ingest.MutableWorlds{
		Base: camden.BuildGranarySquareForTests(t),
	}

	projection := b6.NewTileMercatorProjection(16)
	r := BasemapRenderer{RenderRules: BasemapRenderRules, Worlds: w}
	tile, err := r.Render(projection.TileFromLatLng(s2.LatLngFromDegrees(51.53531, -0.12434)), &TileArgs{})
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	expected := "19813dd2"
	if id, _ := strconv.ParseUint(expected, 16, 64); osm.WayID(id) != camden.LightermanWay {
		t.Fatal("Unexpected hex ID for LightermanWay. Test data changed?")
	}

	layer := tile.FindLayer("building")
	if layer == nil {
		t.Fatal("Expected to find building layer")
	}
	found := false
	for _, f := range layer.Features {
		if ns, ok := f.Tags["ns"]; ok && ns == b6.NamespaceOSMWay.String() {
			if id, ok := f.Tags["id"]; ok && id == expected {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("Expected Lighterman feature to be tagged with an ID")
	}
}

func TestFeaturesAreOrderedByLayerTag(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	w := &ingest.MutableWorlds{
		Base: ingest.NewMutableOverlayWorld(granarySquare),
	}

	mutable := w.FindOrCreateWorld(ingest.DefaultWorldFeatureID)

	// Add a roof terrace and second floor to the Lighterman
	lighterman := b6.FindAreaByID(camden.LightermanID, mutable)
	roof := ingest.NewAreaFeatureFromWorld(lighterman)
	roof.AreaID = b6.MakeAreaID(b6.NamespacePrivate, 1)
	roof.AddTag(b6.Tag{Key: "layer", Value: b6.String("2")})
	if err := mutable.AddFeature(roof); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	basement := ingest.NewAreaFeatureFromWorld(lighterman)
	basement.AreaID = b6.MakeAreaID(b6.NamespacePrivate, 2)
	basement.AddTag(b6.Tag{Key: "layer", Value: b6.String("-1")})
	if err := mutable.AddFeature(basement); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	projection := b6.NewTileMercatorProjection(16)
	r := BasemapRenderer{RenderRules: BasemapRenderRules, Worlds: w}
	tile, err := r.Render(projection.TileFromLatLng(s2.LatLngFromDegrees(51.53531, -0.12434)), &TileArgs{})
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	order := []b6.FeatureID{basement.FeatureID(), lighterman.FeatureID(), roof.FeatureID()}
	layer := tile.FindLayer("building")
	if layer == nil {
		t.Fatal("Expected to find building layer")
	}
	next := 0
	for _, f := range layer.Features {
		if f.ID == api.TileFeatureIDForPolygon(order[next], 0) {
			next++
		}
	}
	if next != len(order) {
		t.Errorf("Unexpected feature order while searching for %s", order[next])
	}
}

func TestRulesThatMatchAllTagValues(t *testing.T) {
	rules := []RenderRule{
		{
			Tag: b6.Tag{
				Key:   "#building",
				Value: b6.String(""),
			},
		},
		{
			Tag: b6.Tag{
				Key:   "#building",
				Value: nil,
			},
		},
	}
	for _, r := range rules {
		if !(RenderRules{r}).IsRendered(b6.Tag{Key: "#building", Value: b6.String("yes")}) {
			t.Errorf("Expected building to be rendered with %+v", r)
		}
		if (RenderRules{r}).IsRendered(b6.Tag{Key: "#amenity", Value: b6.String("cafe")}) {
			t.Errorf("Expected cafe to not be rendered with %+v", r)
		}
	}
}
