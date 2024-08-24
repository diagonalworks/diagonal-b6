package gdal

import (
	"context"
	"fmt"
	"math"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"
	"diagonal.works/b6/test"
	"github.com/golang/geo/s2"
)

func TestReadFeaturesFromLSOABoundaries(t *testing.T) {
	source := Source{
		Filename:   test.Data("lsoa-camden.shp"),
		CopyTags:   []CopyTag{{Field: "LSOA11CD", Key: "code"}, {Field: "LSOA11NM", Key: "name"}, {Field: "POPULATION", Key: "population"}},
		AddTags:    []b6.Tag{{Key: "#boundary", Value: b6.NewStringExpression("lsoa")}},
		IDField:    "LSOA11CD",
		IDStrategy: UKONS2011IDStrategy,
		Bounds:     s2.FullRect(),
	}

	var found ingest.Feature
	emit := func(f ingest.Feature, goroutine int) error {
		if f.FeatureID() == b6.FeatureIDFromUKONSCode("E01000858", 2011, b6.FeatureTypeArea) {
			found = f
		}
		return nil
	}
	err := source.Read(ingest.ReadOptions{}, emit, context.Background())
	if err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

	if found == nil {
		t.Fatal("Expected to find boundary")
	}
	expected := []b6.Tag{{Key: "#boundary", Value: b6.NewStringExpression("lsoa")}, {Key: "name", Value: b6.NewStringExpression("Camden 018B")}}
	for _, tag := range expected {
		if found.Get(tag.Key) != tag {
			t.Errorf("Expected to find %s", tag)
		}
	}
}

func TestReadFeaturesFromLSOABoundariesCopyingAllFields(t *testing.T) {
	source := Source{
		Filename:      test.Data("lsoa-camden.shp"),
		CopyAllFields: true,
		CopyTags:      []CopyTag{{Field: "LSOA11CD", Key: "code"}},
		AddTags:       []b6.Tag{{Key: "#boundary", Value: b6.NewStringExpression("lsoa")}},
		IDField:       "LSOA11CD",
		IDStrategy:    UKONS2011IDStrategy,
		Bounds:        s2.FullRect(),
	}

	var found ingest.Feature
	emit := func(f ingest.Feature, goroutine int) error {
		if f.FeatureID() == b6.FeatureIDFromUKONSCode("E01000858", 2011, b6.FeatureTypeArea) {
			found = f
		}
		return nil
	}
	err := source.Read(ingest.ReadOptions{}, emit, context.Background())
	if err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}
	if found == nil {
		t.Fatal("Expected to find boundary")
	}
	expected := []b6.Tag{{Key: "code", Value: b6.NewStringExpression("E01000858")}, {Key: "LSOA11NM", Value: b6.NewStringExpression("Camden 018B")}}
	for _, tag := range expected {
		if found.Get(tag.Key) != tag {
			t.Errorf("Expected to find %s", tag)
		}
	}
}

func TestReadFeatureFromWardBoundaryWithHole(t *testing.T) {
	source := Source{
		Filename:   test.Data("ward-hole.shp"),
		AddTags:    []b6.Tag{{Key: "#boundary", Value: b6.NewStringExpression("ward")}},
		IDField:    "WD21CD",
		IDStrategy: UKONS2021IDStrategy,
		Bounds:     s2.FullRect(),
	}
	var found *ingest.AreaFeature
	emit := func(f ingest.Feature, goroutine int) error {
		if f.FeatureID() == b6.FeatureIDFromUKONSCode("E05003517", 2021, b6.FeatureTypeArea) {
			if a, ok := f.(*ingest.AreaFeature); ok {
				found = a
			} else {
				return fmt.Errorf("expected an area feature, found %T", f)
			}
		}
		return nil
	}
	err := source.Read(ingest.ReadOptions{}, emit, context.Background())
	if err != nil {
		t.Fatalf("expected no error, found %s", err)
	}
	if found == nil {
		t.Fatal("expected to find boundary")
	}

	if l := found.Len(); l != 1 {
		t.Fatalf("expected one polygon, found %d", l)
	}

	polygon, ok := found.Polygon(0)
	if !ok {
		t.Fatalf("expected a polygon from points, not path IDs")
	}

	inside := s2.LatLngFromDegrees(50.85720, -3.40841)
	if !polygon.ContainsPoint(s2.PointFromLatLng(inside)) {
		t.Fatalf("expected polygon to contain point")
	}
}

func TestReadFeatureFromWardBoundaryWithInvertedLoop(t *testing.T) {
	// Boundary E05004196 contains both polygons with incorrectly wound
	// boundaries, but also geometry that becomes invalid during the
	// compact marshalling process. This test ensures that incorrectly
	// wound loops are inverted, and the loops that become invalid during
	// marshalling are dropped.
	source := Source{
		Filename:   test.Data("ward-inverted.shp"),
		AddTags:    []b6.Tag{{Key: "#boundary", Value: b6.NewStringExpression("ward")}},
		IDField:    "WD22CD",
		IDStrategy: UKONS2022IDStrategy,
		Bounds:     s2.FullRect(),
	}

	options := compact.Options{Goroutines: 2, PointsScratchOutputType: compact.OutputTypeMemory}
	index, err := compact.BuildInMemory(&source, &options)
	if err != nil {
		t.Fatalf("expected no error, found: %s", err)
	}
	w := compact.NewWorld()
	w.Merge(index)

	id := b6.FeatureIDFromUKONSCode("E05004196", 2022, b6.FeatureTypeArea)
	a := b6.FindAreaByID(id.ToAreaID(), w)
	if a == nil {
		t.Fatalf("failed to find expected area")
	}

	inside := s2.PointFromLatLng(s2.LatLngFromDegrees(51.720359, 0.694137))
	outside := s2.PointFromLatLng(s2.LatLngFromDegrees(51.722034, 0.707170))

	insideOK := false
	outsideOK := true

	var total float64
	for i := 0; i < a.Len(); i++ {
		p := a.Polygon(i)
		total += b6.AreaToMeters2(p.Area())
		insideOK = insideOK || p.ContainsPoint(inside)
		outsideOK = outsideOK && !p.ContainsPoint(outside)
	}

	expected := 3450000.0
	if r := math.Abs(1.0 - (total / expected)); r > 0.01 {
		t.Errorf("total area outside expected range: ratio %f %f %f", r, total, expected)
	}

	if !insideOK {
		t.Errorf("point not found found to be inside area")
	}

	if !outsideOK {
		t.Errorf("point not found found to be outside area")
	}
}
