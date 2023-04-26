package gdal

import (
	"context"
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
		AddTags:    []b6.Tag{{Key: "#boundary", Value: "lsoa"}},
		IDField:    "LSOA11CD",
		IDStrategy: GBONS2011IDStrategy,
		Bounds:     s2.FullRect(),
	}

	var found ingest.Feature
	emit := func(f ingest.Feature, goroutine int) error {
		if f.FeatureID() == b6.FeatureIDFromGBONSCode("E01000858", 2011, b6.FeatureTypeArea) {
			found = f
		}
		return nil
	}
	err := source.Read(ingest.ReadOptions{}, emit, context.Background())
	if err != nil {
		t.Errorf("Expected no error, found %s", err)
		return
	}

	if found != nil {
		expected := []b6.Tag{{Key: "#boundary", Value: "lsoa"}, {Key: "name", Value: "Camden 018B"}}
		for _, tag := range expected {
			if found.Get(tag.Key) != tag {
				t.Errorf("Expected to find %s", tag)
			}
		}
	} else {
		t.Errorf("Expected to find boundary")
	}
}

func TestReadFeaturesFromLSOABoundariesCopyingAllFields(t *testing.T) {
	source := Source{
		Filename:      test.Data("lsoa-camden.shp"),
		CopyAllFields: true,
		CopyTags:      []CopyTag{{Field: "LSOA11CD", Key: "code"}},
		AddTags:       []b6.Tag{{Key: "#boundary", Value: "lsoa"}},
		IDField:       "LSOA11CD",
		IDStrategy:    GBONS2011IDStrategy,
		Bounds:        s2.FullRect(),
	}

	var found ingest.Feature
	emit := func(f ingest.Feature, goroutine int) error {
		if f.FeatureID() == b6.FeatureIDFromGBONSCode("E01000858", 2011, b6.FeatureTypeArea) {
			found = f
		}
		return nil
	}
	err := source.Read(ingest.ReadOptions{}, emit, context.Background())
	if err != nil {
		t.Errorf("Expected no error, found %s", err)
		return
	}
	if found != nil {
		expected := []b6.Tag{{Key: "code", Value: "E01000858"}, {Key: "LSOA11NM", Value: "Camden 018B"}}
		for _, tag := range expected {
			if found.Get(tag.Key) != tag {
				t.Errorf("Expected to find %s", tag)
			}
		}
	} else {
		t.Errorf("Expected to find boundary")
	}
}
