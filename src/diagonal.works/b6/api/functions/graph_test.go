package functions

import (
	"context"
	"fmt"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/graph"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"
)

func TestAccessibility(t *testing.T) {
	w := camden.BuildGranarySquareForTests(t)
	if w == nil {
		return
	}

	options := []b6.Tag{{Key: "mode", Value: "walk"}}
	od, err := accessibilityForGranarySquare(options, w)
	if err != nil {
		t.Fatal(err)
	}

	seen := make(map[graph.OD]struct{})
	if err := fillODsFromCollection(seen, od); err != nil {
		t.Fatal(err)
	}

	expected := []graph.OD{
		{
			Origin:      camden.StableStreetBridgeNorthEndID.FeatureID(),
			Destination: ingest.FromOSMNodeID(3790640851).FeatureID(),
		},
		{
			Origin:      camden.VermuteriaID.FeatureID(),
			Destination: b6.FeatureIDInvalid,
		},
	}
	for _, od := range expected {
		if _, ok := seen[od]; !ok {
			t.Errorf("Failed to find expected origin destination pair: %+v", od)
		}
	}

	unexpected := []graph.OD{
		{
			Origin:      camden.StableStreetBridgeNorthEndID.FeatureID(),
			Destination: b6.FeatureIDInvalid,
		},
	}
	for _, od := range unexpected {
		if _, ok := seen[od]; ok {
			t.Errorf("Found unexpected origin destination pair: %+v", od)
		}
	}
}

func TestAccessibilityFlipped(t *testing.T) {
	w := camden.BuildGranarySquareForTests(t)
	if w == nil {
		return
	}

	options := []b6.Tag{{Key: "flip", Value: "yes"}, {Key: "mode", Value: "walk"}}
	od, err := accessibilityForGranarySquare(options, w)
	if err != nil {
		t.Fatal(err)
	}

	seen := make(map[graph.OD]struct{})
	if err := fillODsFromCollection(seen, od); err != nil {
		t.Fatal(err)
	}

	expected := graph.OD{
		Origin:      ingest.FromOSMNodeID(3790640851).FeatureID(),
		Destination: ingest.FromOSMNodeID(1447052073).FeatureID(),
	}
	if _, ok := seen[expected]; !ok {
		t.Errorf("Failed to find expected origin destination pair")
	}
}

func accessibilityForGranarySquare(options []b6.Tag, w b6.World) (b6.Collection[b6.Identifiable, b6.FeatureID], error) {
	context := &api.Context{
		World:   w,
		Cores:   2,
		Context: context.Background(),
	}
	origins := b6.ArrayFeatureCollection[b6.Feature]{
		w.FindFeatureByID(camden.StableStreetBridgeNorthEndID.FeatureID()),
		w.FindFeatureByID(camden.VermuteriaID.FeatureID()),
	}
	ids := b6.AdaptCollection[any, b6.Identifiable](origins.Collection())
	return accessible(context, ids, b6.Keyed{Key: "entrance"}, 500, b6.ArrayValuesCollection[b6.Tag](options).Collection())
}

func fillODsFromCollection(ods map[graph.OD]struct{}, c b6.Collection[b6.Identifiable, b6.FeatureID]) error {
	i := c.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return fmt.Errorf("Expected no error from Next, found: %s", err)
		} else if !ok {
			break
		}
		ods[graph.OD{Origin: i.Key().FeatureID(), Destination: i.Value().FeatureID()}] = struct{}{}
	}
	return nil
}
