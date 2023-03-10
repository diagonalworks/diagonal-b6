package functions

import (
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/region"
)

func exportWorld(filename string, c *api.Context) (int, error) {
	// TODO: Shouldn't return anything
	source := ingest.WorldFeatureSource{World: c.World}
	config := region.Config{
		OutputFilename:       filename,
		Cores:                c.Cores,
		WorkDirectory:        "",
		PointsWorkOutputType: region.OutputTypeMemory,
	}
	return 0, region.BuildRegionFromPBF(source, &config)
}
