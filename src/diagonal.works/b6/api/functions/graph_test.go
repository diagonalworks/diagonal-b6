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

	expected := graph.OD{
		Origin:      ingest.FromOSMNodeID(1447052073).FeatureID(),
		Destination: ingest.FromOSMNodeID(3790640851).FeatureID(),
	}
	if _, ok := seen[expected]; !ok {
		t.Errorf("Failed to find expected origin destination pair")
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

func accessibilityForGranarySquare(options []b6.Tag, w b6.World) (api.Collection, error) {
	context := &api.Context{
		World:   w,
		Cores:   2,
		Context: context.Background(),
	}
	origins := &api.ArrayFeatureCollection{
		Features: []b6.Feature{w.FindFeatureByID(camden.StableStreetBridgeNorthEndID.FeatureID())},
	}
	return accessible(context, origins, b6.Keyed{Key: "entrance"}, 500, &api.ArrayTagCollection{Tags: options})
}

func fillODsFromCollection(ods map[graph.OD]struct{}, c api.Collection) error {
	i := c.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return fmt.Errorf("Expected no error from Next, found: %s", err)
		} else if !ok {
			break
		}
		if o, ok := i.Key().(b6.Identifiable); ok {
			if d, ok := i.Value().(b6.Identifiable); ok {
				ods[graph.OD{Origin: o.FeatureID(), Destination: d.FeatureID()}] = struct{}{}
			}
		} else {
			return fmt.Errorf("Expected b6.Identifiable for key, found %T", i.Key())
		}
	}
	return nil
}
