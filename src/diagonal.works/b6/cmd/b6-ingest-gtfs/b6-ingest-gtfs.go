package main

import (
	"flag"
	"log"
	"runtime"

	"diagonal.works/b6/ingest/compact"
	"diagonal.works/b6/ingest/gtfs"

	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/local"
)

// TODO: add monitoring
func main() {
	input := flag.String("input", "", "Input directory, with GTFS data")
	output := flag.String("output", "", "Output index filename")
	cores := flag.Int("cores", runtime.NumCPU(), "Available cores")
	operator := flag.String("operator", "", "GTFS operator, useful if overlaying multiple GTFS data sources; optional")
	scratch := flag.String("scratch", ".", "Directory for temporary files")
	flag.Parse()

	if err := compact.Build(
		&gtfs.TXTFilesGTFSSource{Directory: *input, Operator: *operator, FailWhenNoFiles: true},
		&compact.Options{
			OutputFilename:          *output,
			Goroutines:              *cores,
			ScratchDirectory:        *scratch,
			PointsScratchOutputType: compact.OutputTypeMemory,
		}); err != nil {
		log.Fatal(err)
	}
}
