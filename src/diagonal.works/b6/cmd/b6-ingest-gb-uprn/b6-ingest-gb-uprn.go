package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"
	"diagonal.works/b6/ingest/gb/uprn"

	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/gcs"
	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/local"
)

func main() {
	inputFlag := flag.String("input", "", "Input open UPRN CSV, gzipped")
	outputFlag := flag.String("output", "", "Output index")
	boundingBoxFlag := flag.String("bounding-box", "", "lat,lng,lat,lng bounding box to crop points outside")
	filterFlag := flag.String("filter", "", "A b6 shell expression for a function taking a point feature and returning a boolean")
	joinFlag := flag.String("join", "", "Join tag values from a CSV")
	cores := flag.Int("cores", runtime.NumCPU(), "Available cores")
	flag.Parse()

	crop, err := ingest.ParseBoundingBox(*boundingBoxFlag)
	if err != nil {
		log.Fatal(err)
	}

	var filter func(g b6.Feature, c *api.Context) (bool, error)
	apiContext := functions.NewContext(b6.EmptyWorld{})
	if *filterFlag != "" {
		expression, err := api.ParseExpression(*filterFlag)
		if err != nil {
			log.Fatal(err)
		}
		err = api.EvaluateAndFill(expression, apiContext, &filter)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		filter = func(f b6.Feature, c *api.Context) (bool, error) {
			if p, ok := f.(b6.PointFeature); ok {
				return crop.ContainsPoint(p.Point()), nil
			}
			return true, nil
		}
	}

	source := &uprn.Source{
		Filename: *inputFlag,
		Filter:   filter,
		Context:  apiContext,
	}

	if *joinFlag != "" {
		var err error
		patterns := strings.Split(*joinFlag, ",")
		source.JoinTags, err = ingest.NewJoinTagsFromPatterns(patterns, context.Background())
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	config := compact.Options{
		OutputFilename:       *outputFlag,
		Cores:                *cores,
		WorkDirectory:        "",
		PointsWorkOutputType: compact.OutputTypeMemory,
	}
	if err := compact.Build(source, &config); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
