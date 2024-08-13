package gdal

import (
	"context"
	"fmt"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
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
