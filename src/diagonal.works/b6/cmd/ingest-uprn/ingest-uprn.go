package main

import (
	"compress/gzip"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/region"

	"github.com/golang/geo/s2"
)

type UPRNSource struct {
	Filename string
	Crop     s2.Rect
	Join     map[uint64][]b6.Tag
}

func (s *UPRNSource) Read(options ingest.ReadOptions, emit ingest.Emit, ctx context.Context) error {
	f, err := os.OpenFile(s.Filename, os.O_RDONLY, 0)
	defer f.Close()
	if err != nil {
		return err
	}

	if options.SkipPoints {
		return nil
	}

	d, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	r := csv.NewReader(d)
	header, err := r.Read()
	if err != nil {
		return err
	}
	fields := []string{"UPRN", "LATITUDE", "LONGITUDE"}
	indicies := make([]int, len(fields))
	for i, f := range fields {
		indicies[i] = -1
		for j, column := range header {
			if f == strings.Trim(column, "\ufeff") { // Remove byte order mark
				indicies[i] = j
				break
			}
		}
		if indicies[i] < 0 {
			return fmt.Errorf("Failed to find field %q", f)
		}
	}

	var point ingest.PointFeature
	point.PointID.Namespace = b6.NamespaceGBUPRN
	point.Tags = []b6.Tag{{Key: "#place", Value: "uprn"}}
	uprns := 0
	emits := 0
	joins := 0
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		uprns++
		point.PointID.Value, err = strconv.ParseUint(row[indicies[0]], 10, 64)
		if err != nil {
			return fmt.Errorf("Parsing id for %q: %s", row, err)
		}
		lat, err := strconv.ParseFloat(row[indicies[1]], 64)
		if err != nil {
			return fmt.Errorf("Parsing latitude for %q: %s", row, err)
		}
		lng, err := strconv.ParseFloat(row[indicies[2]], 64)
		if err != nil {
			return fmt.Errorf("Parsing longitude for %q: %s", row, err)
		}
		point.Location = s2.LatLngFromDegrees(lat, lng)
		if s.Crop.ContainsLatLng(point.Location) {
			point.Tags = point.Tags[0:1] // Keep #place=uprn
			for _, t := range s.Join[point.PointID.Value] {
				point.Tags = append(point.Tags, t)
			}
			if len(point.Tags) > 1 {
				joins++
			}
			emits++
			if err := emit(&point, 0); err != nil {
				return err
			}
		}
	}
	log.Printf("uprns: %d emits: %d joined: %d", uprns, emits, joins)
	return nil
}

func parseJoinedCSV(filename string) (map[uint64][]b6.Tag, error) {
	if filename == "" {
		return map[uint64][]b6.Tag{}, nil
	}

	f, err := os.OpenFile(filename, os.O_RDONLY, 0)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(f)
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	if len(header) < 2 {
		return nil, fmt.Errorf("Expected at least 2 columns in %s, found %s", filename, header)
	}

	join := make(map[uint64][]b6.Tag)
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if len(row) == len(header) {
			id, err := strconv.ParseUint(row[0], 10, 64)
			if err == nil {
				tags := join[id]
				for i := 1; i < len(row); i++ {
					tags = append(tags, b6.Tag{Key: header[i], Value: row[i]})
				}
				join[id] = tags
			}
		}
	}
	return join, nil
}

func main() {
	inputFlag := flag.String("input", "", "Input open UPRN CSV, gzipped")
	outputFlag := flag.String("output", "", "Output index")
	boundingBoxFlag := flag.String("bounding-box", "", "lat,lng,lat,lng bounding box to crop points outside")
	joinFlag := flag.String("join", "", "Join tag values from a CSV")
	flag.Parse()

	crop, err := ingest.ParseBoundingBox(*boundingBoxFlag)
	if err != nil {
		log.Fatal(err)
	}

	join, err := parseJoinedCSV(*joinFlag)
	if err != nil {
		log.Fatal(err)
	}

	source := &UPRNSource{
		Filename: *inputFlag,
		Crop:     crop,
		Join:     join,
	}

	config := region.Config{
		OutputFilename:       *outputFlag,
		Cores:                1,
		WorkDirectory:        "",
		PointsWorkOutputType: region.OutputTypeMemory,
	}
	// TODO: rename, it's not a PBF
	if err := region.BuildRegionFromPBF(source, &config); err != nil {
		log.Fatal(err)
	}
}
