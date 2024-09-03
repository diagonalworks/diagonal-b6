package main

import (
	"context"
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"
	"diagonal.works/b6/ingest/gdal"

	"github.com/golang/geo/s2"

	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/gcs"
	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/local"
)

var idStrategies = map[string]gdal.IDStrategy{
	"":            gdal.IndexIDStrategy,
	"strip":       gdal.StripNonDigitsIDStrategy,
	"hash":        gdal.HashIDStrategy,
	"uk-ons-2011": gdal.UKONS2011IDStrategy,
	"uk-ons-2021": gdal.UKONS2021IDStrategy,
	"uk-ons-2022": gdal.UKONS2022IDStrategy,
}

func main() {
	inputFlag := flag.String("input", "", "Input shapefile")
	outputFlag := flag.String("output", "", "Output index")
	layerFlag := flag.String("layer", "", "Name of layer to ingest, empty for all")
	namespaceFlag := flag.String("namespace", "", "Namespace for features")
	idFlag := flag.String("id", "", "Field to use for ID generation")
	idStategyFlag := flag.String("id-strategy", "", "Strategy to use for ID generation")
	copyAllFieldsFlag := flag.Bool("copy-all-fields", false, "Copy all fields as tags with the same name, unless mentioned in --copy-tags")
	copyTagsFlag := flag.String("copy-tags", "", "Attributes to copy from underlying data, eg name=LSOA11NM")
	addTagsFlag := flag.String("add-tags", "", "Tags to add to imported data, eg #boundary=datazone,year=2011")
	recurseFlag := flag.Bool("recurse", false, "Recurse into directories")
	zippedFlag := flag.Bool("zipped", false, "Read shapefiles within zipfiles")
	boundingBoxFlag := flag.String("bounding-box", "", "lat,lng,lat,lng bounding box to crop points outside")
	joinFlag := flag.String("join", "", "Join tag values from a CSV")
	keepLargeLoopsFlag := flag.Bool("keep-large-loops", false, "Keep loops that cover more than half of the earth's surface (ie don't invert large loops)")
	colourAreasFlag := flag.Bool("colour-areas", false, "Add a b6:colour tag to areas, such that adjacent areas don't have the same colour")
	coresFlag := flag.Int("cores", runtime.NumCPU(), "Number of cores available")
	scratch := flag.String("scratch", ".", "Directory for temporary files, for --memory=false or writing to cloud")
	flag.Parse()

	if *inputFlag == "" || *outputFlag == "" {
		fmt.Fprintln(os.Stderr, "Must specify --input and --output")
		os.Exit(1)
	}

	strategy, ok := idStrategies[*idStategyFlag]
	if !ok {
		ss := make([]string, 0, len(idStrategies))
		for key := range idStrategies {
			if key != "" {
				ss = append(ss, key)
			}
		}
		fmt.Fprintf(os.Stderr, "No ID strategy %q - try one of %v", *idStategyFlag, ss)
		os.Exit(1)
	}

	bounds := s2.FullRect()
	if *boundingBoxFlag != "" {
		var err error
		bounds, err = ingest.ParseBoundingBox(*boundingBoxFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	inputs, err := gdal.FindInputs(*inputFlag, *zippedFlag, *recurseFlag, []string{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	var joinTags ingest.JoinTags
	if *joinFlag != "" {
		var err error
		patterns := strings.Split(*joinFlag, ",")
		joinTags, err = ingest.NewJoinTagsFromPatterns(patterns, context.Background())
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	var copyTags []gdal.CopyTag
	if *copyTagsFlag != "" {
		for _, field := range strings.Split(*copyTagsFlag, ",") {
			if index := strings.Index(field, "="); index > 0 {
				copyTags = append(copyTags, gdal.CopyTag{Key: field[0:index], Field: field[index+1:]})
			} else {
				copyTags = append(copyTags, gdal.CopyTag{Key: field, Field: field})
			}
		}
	}

	var addTags []b6.Tag
	for _, tag := range strings.Split(*addTagsFlag, ",") {
		parts := strings.Split(tag, "=")
		if len(parts) == 2 {
			addTags = append(addTags, b6.Tag{Key: parts[0], Value: b6.NewStringExpression(parts[1])})
		}
	}

	merged := make(ingest.MergedFeatureSource, len(inputs))
	for i, ii := range inputs {
		merged[i] = &gdal.Source{
			Filename:       ii,
			Layer:          *layerFlag,
			Namespace:      b6.Namespace(*namespaceFlag),
			IDField:        *idFlag,
			IDStrategy:     strategy,
			CopyAllFields:  *copyAllFieldsFlag,
			CopyTags:       copyTags,
			AddTags:        addTags,
			JoinTags:       joinTags,
			Bounds:         bounds,
			KeepLargeLoops: *keepLargeLoopsFlag,
		}
	}
	source := ingest.FeatureSource(merged)

	if *colourAreasFlag {
		if source, err = ingest.ColourAreas(source, *coresFlag); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	options := compact.Options{
		OutputFilename:          *outputFlag,
		Goroutines:              *coresFlag,
		ScratchDirectory:        *scratch,
		PointsScratchOutputType: compact.OutputTypeMemory,
	}
	finish, err := compact.MaybeWriteToCloud(&options)
	if err == nil {
		if err = compact.Build(source, &options); err == nil {
			err = finish()
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
