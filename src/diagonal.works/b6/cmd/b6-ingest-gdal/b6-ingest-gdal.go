package main

import (
	"archive/zip"
	"context"
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"path"
	"runtime"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"
	"diagonal.works/b6/ingest/gdal"
	"golang.org/x/sync/errgroup"

	"github.com/golang/geo/s2"

	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/gcs"
	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/local"
)

var idStrategies = map[string]gdal.IDStrategy{
	"":            gdal.IndexIDStrategy,
	"strip":       gdal.StripNonDigitsIDStrategy,
	"hash":        gdal.HashIDStrategy,
	"gb-ons-2011": gdal.GBONS2011IDStrategy,
}

type input interface {
	FilenameForGDAL() string
}

type fileInput struct {
	Filename string
}

func (f fileInput) FilenameForGDAL() string {
	return f.Filename
}

type zipFileInput struct {
	ZipFilename string
	Filename    string
}

func (z zipFileInput) FilenameForGDAL() string {
	return fmt.Sprintf("/vsizip/%s/%s", z.ZipFilename, z.Filename)
}

func isIngestable(filename string) bool {
	return strings.HasSuffix(filename, ".shp") || strings.HasSuffix(filename, ".geojson")
}

func findInputs(filename string, zipped bool, recurse bool, inputs []input) ([]input, error) {
	s, err := os.Stat(filename)
	if err != nil {
		// If we can't stat the file, it may be because it's a gdal virtual
		// path, like /vsizip/....
		return append(inputs, fileInput{Filename: filename}), nil
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
					inputs = append(inputs, zipFileInput{ZipFilename: filename, Filename: zf.Name})
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
			inputs, err = findInputs(path.Join(filename, entry.Name()), zipped, recurse, inputs)
			if err != nil {
				return nil, err
			}
		}
	} else if isIngestable(filename) {
		inputs = append(inputs, fileInput{Filename: filename})
	}
	return inputs, nil
}

type mergedSource []*gdal.Source

func (m mergedSource) Read(options ingest.ReadOptions, emit ingest.Emit, ctx context.Context) error {
	if options.Goroutines < 1 {
		options.Goroutines = 1
	}
	perFile := options
	perFile.Goroutines /= len(m)
	if perFile.Goroutines < 1 {
		perFile.Goroutines = 1
	}
	g, gc := errgroup.WithContext(ctx)
	c := make(chan *gdal.Source)
	for i := 0; i < options.Goroutines/perFile.Goroutines; i++ {
		i := i
		g.Go(func() error {
			offset := func(f ingest.Feature, g int) error {
				return emit(f, g+(i*perFile.Goroutines))
			}
			for {
				select {
				case <-gc.Done():
					return nil
				case s, ok := <-c:
					if ok {
						if err := s.Read(perFile, offset, ctx); err != nil {
							return err
						}
					} else {
						return nil
					}
				}
			}
		})
	}
	for _, s := range m {
		select {
		case <-gc.Done():
			return g.Wait()
		case c <- s:
		}
	}
	close(c)
	return g.Wait()
}

func main() {
	inputFlag := flag.String("input", "", "Input shapefile")
	outputFlag := flag.String("output", "", "Output index")
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
	coresFlag := flag.Int("cores", runtime.NumCPU(), "Number of cores available")
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

	inputs, err := findInputs(*inputFlag, *zippedFlag, *recurseFlag, []input{})
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
			addTags = append(addTags, b6.Tag{Key: parts[0], Value: parts[1]})
		}
	}

	source := make(mergedSource, len(inputs))
	for i, ii := range inputs {
		source[i] = &gdal.Source{
			Filename:      ii.FilenameForGDAL(),
			Namespace:     b6.Namespace(*namespaceFlag),
			IDField:       *idFlag,
			IDStrategy:    strategy,
			CopyAllFields: *copyAllFieldsFlag,
			CopyTags:      copyTags,
			AddTags:       addTags,
			JoinTags:      joinTags,
			Bounds:        bounds,
		}
	}

	options := compact.Options{
		OutputFilename:       *outputFlag,
		Goroutines:           *coresFlag,
		WorkDirectory:        "",
		PointsWorkOutputType: compact.OutputTypeMemory,
	}

	if err := compact.Build(source, &options); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
