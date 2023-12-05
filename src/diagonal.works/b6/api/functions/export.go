package functions

import (
	"fmt"

	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"
)

// Write the current world to the given filename in the b6 compact index format.
// As the file is written by the b6 server process, the filename it relative
// to the filesystems it sees. Writing files to cloud storage is
// supported.
func exportWorld(c *api.Context, filename string) (int, error) {
	// TODO: Shouldn't return anything
	if !c.FileIOAllowed {
		return 0, fmt.Errorf("File IO is not allowed")
	}

	source := ingest.WorldFeatureSource{World: c.World}
	options := compact.Options{
		OutputFilename:       filename,
		Goroutines:           c.Cores,
		WorkDirectory:        "",
		PointsWorkOutputType: compact.OutputTypeMemory,
	}
	return 0, compact.Build(source, &options)
}
