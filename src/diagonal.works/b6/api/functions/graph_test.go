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

	options := []b6.Tag{{Key: "mode", Value: b6.StringExpression("walk")}}
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
			Origin:      camden.StableStreetBridgeNorthEndID,
			Destination: ingest.FromOSMNodeID(3790640851),
		},
		{
			Origin:      camden.VermuteriaID,
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
			Origin:      camden.StableStreetBridgeNorthEndID,
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

	options := []b6.Tag{{Key: "flip", Value: b6.StringExpression("yes")}, {Key: "mode", Value: b6.StringExpression("walk")}}
	od, err := accessibilityForGranarySquare(options, w)
	if err != nil {
		t.Fatal(err)
	}

	seen := make(map[graph.OD]struct{})
	if err := fillODsFromCollection(seen, od); err != nil {
		t.Fatal(err)
	}

	expected := graph.OD{
		Origin:      ingest.FromOSMNodeID(3790640851),
		Destination: ingest.FromOSMNodeID(1447052073),
	}
	if _, ok := seen[expected]; !ok {
		t.Errorf("Failed to find expected origin destination pair")
	}
}

func TestWeightsFromOptions(t *testing.T) {
	w := camden.BuildGranarySquareForTests(t)
	if w == nil {
		return
	}

	options := []b6.Tag{
		{Key: "mode", Value: b6.StringExpression("transit")},
		{Key: "walk:speed", Value: b6.StringExpression("7.6")},
	}
	weights, err := WeightsFromOptions(b6.ArrayValuesCollection[b6.Tag](options).Collection(), w)
	if err != nil {
		t.Errorf("expected no error, found: %s", err)
	}

	expected := graph.TransitTimeWeights{PeakTraffic: true, Weights: graph.WalkingTimeWeights{Speed: 7.6}}
	if weights != expected {
		t.Errorf("expected %+v, found %+v", expected, weights)
	}

	options = []b6.Tag{
		{Key: "mode", Value: b6.StringExpression("transit")},
		{Key: "elevation", Value: b6.StringExpression("true")},
		{Key: "elevation:downhill", Value: b6.StringExpression("1.2")},
		{Key: "walk:speed", Value: b6.StringExpression("8.7")},
	}
	weights, err = WeightsFromOptions(b6.ArrayValuesCollection[b6.Tag](options).Collection(), w)
	if err != nil {
		t.Errorf("expected no error, found: %s", err)
	}

	expected = graph.TransitTimeWeights{
		PeakTraffic: true,
		Weights: graph.ElevationWeights{
			UpHillPenalty:   1.0,
			DownHillPenalty: 1.2,
			Weights: graph.WalkingTimeWeights{
				Speed: 8.7,
			},
			W: w,
		},
	}
	if weights != expected {
		t.Errorf("expected %+v, found %+v", expected, weights)
	}
}

func accessibilityForGranarySquare(options []b6.Tag, w b6.World) (b6.Collection[b6.FeatureID, b6.FeatureID], error) {
	context := &api.Context{
		World:   w,
		Cores:   2,
		Context: context.Background(),
	}
	origins := b6.ArrayFeatureCollection[b6.Feature]{
		w.FindFeatureByID(camden.StableStreetBridgeNorthEndID),
		w.FindFeatureByID(camden.VermuteriaID),
	}
	ids := b6.AdaptCollection[any, b6.Identifiable](origins.Collection())
	return accessibleAll(context, ids, b6.Keyed{Key: "entrance"}, 500, b6.ArrayValuesCollection[b6.Tag](options).Collection())
}

func fillODsFromCollection(ods map[graph.OD]struct{}, c b6.Collection[b6.FeatureID, b6.FeatureID]) error {
	i := c.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return fmt.Errorf("Expected no error from Next, found: %s", err)
		} else if !ok {
			break
		}
		ods[graph.OD{Origin: i.Key(), Destination: i.Value()}] = struct{}{}
	}
	return nil
}
