package gdal

import (
	"archive/zip"
	"fmt"
	"os"
	"path"
	"strings"
)

func isIngestable(filename string) bool {
	return strings.HasSuffix(filename, ".shp") || strings.HasSuffix(filename, ".geojson")
}

func FindInputs(filename string, zipped bool, recurse bool, inputs []string) ([]string, error) {
	s, err := os.Stat(filename)
	if err != nil {
		// If we can't stat the file, it may be because it's a gdal virtual
		// path, like /vsizip/....
		return append(inputs, filename), nil
	}
	if strings.HasSuffix(filename, ".zip") {
		if zipped {
			f, err := os.Open(filename)
			if err != nil {
				return nil, fmt.Errorf("can't open %s: %s", filename, err)
			}
			z, err := zip.NewReader(f, s.Size())
			for _, zf := range z.File {
				if isIngestable(zf.Name) {
					inputs = append(inputs, fmt.Sprintf("/vsizip/%s/%s", filename, zf.Name))
				}
			}
			f.Close()
		}
	} else if s.IsDir() {
		entries, err := os.ReadDir(filename)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", filename, err)
		}
		for _, entry := range entries {
			var err error
			inputs, err = FindInputs(path.Join(filename, entry.Name()), zipped, recurse, inputs)
			if err != nil {
				return nil, err
			}
		}
	} else if isIngestable(filename) {
		inputs = append(inputs, filename)
	}
	return inputs, nil
}
