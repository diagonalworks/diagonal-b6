package test

import (
	"os"
	"path/filepath"
	"strings"
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
