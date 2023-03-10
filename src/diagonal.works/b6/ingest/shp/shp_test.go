package shp

import (
	"testing"

	"diagonal.works/b6/test"

	"github.com/golang/geo/s2"
)

type onsLSOA struct {
	Code string `shp:"code"`
	Name string `shp:"name"`
}

func TestReadFeaturesFromLSOABoundaries(t *testing.T) {
	attributes := make([]onsLSOA, 0)
	geometries, err := ReadFeatures(test.Data("census/lsoa-camden.shp"), &attributes)
	if err != nil {
		t.Error(err)
		return
	}

	expectedName, foundName := "Camden 018B", false
	expectedCode, foundCode := "E01000858", false

	for i, a := range attributes {
		if a.Name == expectedName {
			foundName = true
		}
		if a.Code == expectedCode {
			foundCode = true
		}
		if _, ok := geometries[i].(*s2.Polygon); !ok {
			t.Errorf("Expected a *s2.Polygon, found %T", geometries[i])
		}
	}

	if !foundName {
		t.Errorf("Failed to find expected name %q", expectedName)
	}
	if !foundCode {
		t.Errorf("Failed to find expected code %q", expectedCode)
	}
}
