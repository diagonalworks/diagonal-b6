package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

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
	scratch := flag.String("scratch", ".", "Directory for temporary files, for --memory=false  or writing to cloud")
	flag.Parse()

	var err error
	if *input == "" || *output == "" {
		err = fmt.Errorf("must specify --input and --output")
	} else {
		t := compact.OutputTypeMemory
		if !*memory {
			t = compact.OutputTypeDisk
		}
		options := compact.Options{
			OutputFilename:          *output,
			Goroutines:              *cores,
			ScratchDirectory:        *scratch,
			PointsScratchOutputType: t,
		}
		var finish func() error
		finish, err = compact.MaybeWriteToCloud(&options)
		if err == nil {
			osmSource := ingest.PBFFilesOSMSource{Glob: *input}
			var source ingest.FeatureSource
			source, err = ingest.NewFeatureSourceFromPBF(&osmSource, &ingest.BuildOptions{Cores: *cores}, context.Background())
			start := time.Now()
			if err == nil {
				err = compact.Build(source, &options)
			}

			log.Printf("index build time: %s", time.Since(start).Truncate(time.Second))
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			log.Printf("total alloc MB: %d", m.TotalAlloc/(1024*1024))

			if err == nil {
				err = finish()
			}
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
