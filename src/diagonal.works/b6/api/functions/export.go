package functions

import (
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"
)

func exportWorld(filename string, c *api.Context) (int, error) {
	// TODO: Shouldn't return anything
	source := ingest.WorldFeatureSource{World: c.World}
	options := compact.Options{
		OutputFilename:       filename,
		Goroutines:           c.Cores,
		WorkDirectory:        "",
		PointsWorkOutputType: compact.OutputTypeMemory,
	}
	return 0, compact.Build(source, &options)
}
