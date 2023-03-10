package region

import (
	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
)

func BuildOverlayRegionInMemory(source ingest.FeatureSource, c *Config, base b6.World) ([]byte, error) {
	var output MemoryOutput
	if err := buildRegionFromPBF(source, base, c, &output); err != nil {
		return nil, err
	}
	bytes, _, _ := output.Bytes()
	return bytes, nil
}

func BuildOverlayRegion(source ingest.FeatureSource, c *Config, base b6.World) error {
	// TODO: It's not actually from a PBF anymore.
	return buildRegionFromPBF(source, base, c, c.Output())
}
