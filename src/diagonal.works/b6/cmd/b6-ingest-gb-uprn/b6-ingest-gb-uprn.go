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
	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"

	"github.com/golang/geo/s2"

	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/gcs"
	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/local"
)

type UPRNSource struct {
	Filename string
	Filter   func(p b6.PointFeature, c *api.Context) (bool, error)
	JoinTags ingest.JoinTags
}

type PointWrapper struct {
	*ingest.PointFeature
}

func (p PointWrapper) PointID() b6.PointID {
	return p.PointFeature.PointID
}

func (p PointWrapper) Covering(coverer s2.RegionCoverer) s2.CellUnion {
	return s2.CellUnion([]s2.CellID{p.CellID().Parent(coverer.MaxLevel)})
}

func (p PointWrapper) ToGeoJSON() geojson.GeoJSON {
	return b6.PointFeatureToGeoJSON(p)
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

	point := ingest.PointFeature{
		PointID: b6.PointID{
			Namespace: b6.NamespaceGBUPRN,
		},
		Tags: []b6.Tag{{Key: "#place", Value: "uprn"}},
	}
	context := api.Context{World: b6.EmptyWorld{}}
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
		point.Tags = point.Tags[0:1] // Keep #place=uprn
		s.JoinTags.AddTags(row[indicies[0]], &point)
		if len(point.Tags) > 1 {
			joins++
		}
		if ok, err := s.Filter(PointWrapper{PointFeature: &point}, &context); ok {
			emits++
			if err := emit(&point, 0); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	log.Printf("uprns: %d emits: %d joined: %d", uprns, emits, joins)
	return nil
}

func main() {
	inputFlag := flag.String("input", "", "Input open UPRN CSV, gzipped")
	outputFlag := flag.String("output", "", "Output index")
	boundingBoxFlag := flag.String("bounding-box", "", "lat,lng,lat,lng bounding box to crop points outside")
	filterFlag := flag.String("filter", "", "A b6 shell expression for a function taking a point feature and returning a boolean")
	joinFlag := flag.String("join", "", "Join tag values from a CSV")
	flag.Parse()

	crop, err := ingest.ParseBoundingBox(*boundingBoxFlag)
	if err != nil {
		log.Fatal(err)
	}

	var filter func(b6.PointFeature, *api.Context) (bool, error)
	if *filterFlag != "" {
		expression, err := api.ParseExpression(*filterFlag)
		if err != nil {
			log.Fatal(err)
		}
		err = api.EvaluateAndFill(expression, b6.EmptyWorld{}, functions.Functions(), functions.FunctionConvertors(), &filter)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		filter = func(p b6.PointFeature, c *api.Context) (bool, error) {
			return crop.ContainsPoint(p.Point()), nil
		}
	}

	source := &UPRNSource{
		Filename: *inputFlag,
		Filter:   filter,
	}

	if *joinFlag != "" {
		var err error
		source.JoinTags, err = ingest.NewJoinTagsFromCSV(*joinFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	config := compact.Options{
		OutputFilename:       *outputFlag,
		Cores:                1,
		WorkDirectory:        "",
		PointsWorkOutputType: compact.OutputTypeMemory,
	}
	if err := compact.Build(source, &config); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
