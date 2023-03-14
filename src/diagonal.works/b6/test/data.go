package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"diagonal.works/b6"
)

const (
	GranarySquarePBF = "granary-square.osm.pbf"
	CamdenPBF        = "camden.osm.pbf"
)

func testDataDirectory() string {
	directory, err := os.Getwd()
	if err == nil {
		if index := strings.Index(directory, "src/diagonal.works/"); index > 0 {
			return filepath.Join(directory[0:index], "data/tests/")
		}
	}
	return ""
}

func Data(filename string) string {
	return filepath.Join(testDataDirectory(), filename)
}

func FindFeatureByID(id b6.FeatureID, w b6.World, t *testing.T) b6.Feature {
	feature := w.FindFeatureByID(id)
	if feature == nil {
		t.Errorf("Failed to find feature expected to be present in test data: %s", id)
	}
	return feature
}
