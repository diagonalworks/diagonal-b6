package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"

	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/gcs"
	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/local"
)

func main() {
	input := flag.String("input", "", "Input filename, OSM PBF format")
	output := flag.String("output", "tmp/out", "Output filename")
	cores := flag.Int("cores", runtime.NumCPU(), "Available cores")
	memory := flag.Bool("memory", true, "Use memory for intermediate data")
	scratch := flag.String("scratch", ".", "Directory for temporary files, for --memory=false")
	flag.Parse()

	var err error
	if *input == "" || *output == "" {
		err = fmt.Errorf("Must specify --input and --output")
	} else {
		t := compact.OutputTypeMemory
		if !*memory {
			t = compact.OutputTypeDisk
		}
		options := compact.Options{
			OutputFilename:       *output,
			Cores:                *cores,
			WorkDirectory:        *scratch,
			PointsWorkOutputType: t,
		}
		osmSource := ingest.PBFFilesOSMSource{Glob: *input}
		var source ingest.FeatureSource
		source, err = ingest.NewFeatureSourceFromPBF(&osmSource, &ingest.BuildOptions{Cores: *cores}, context.Background())
		if err == nil {
			err = compact.Build(source, &options)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
