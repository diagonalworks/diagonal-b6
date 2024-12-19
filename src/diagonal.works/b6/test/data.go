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

func findModuleRoot(dir string) string {
	if dir == "" {
		panic("dir not set")
	}
	dir = filepath.Clean(dir)
	// Look for enclosing go.mod.
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return filepath.Join(dir, "data/tests")
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}
	return ""
}

func testDataDirectory() string {
	directory, err := os.Getwd()
	if err == nil {
		if index := strings.Index(directory, "src/diagonal.works/"); index > 0 {
			return filepath.Join(directory[0:index], "data/tests/")
		}
		return findModuleRoot(directory)
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
