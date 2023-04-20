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
		CopyTags:   []CopyTag{{Field: "code", Key: "code"}, {Field: "name", Key: "name"}},
		AddTags:    []b6.Tag{{Key: "#boundary", Value: "lsoa"}},
		IDField:    "code",
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
		boundary := b6.Tag{Key: "#boundary", Value: "lsoa"}
		if found.Get("#boundary") != boundary {
			t.Errorf("Expected to find %s", boundary)
		}
		name := b6.Tag{Key: "name", Value: "Camden 018B"}
		if found.Get("name") != name {
			t.Errorf("Expected to find %s", name)
		}
	} else {
		t.Errorf("Expected to find boundary")
	}
}
