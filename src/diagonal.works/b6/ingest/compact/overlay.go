package compact

import (
	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
)

func BuildOverlayInMemory(source ingest.FeatureSource, o *Options, base b6.World) ([]byte, error) {
	var output MemoryOutput
	if err := build(source, base, o, &output); err != nil {
		return nil, err
	}
	bytes, _, _ := output.Bytes()
	return bytes, nil
}

func BuildOverlay(source ingest.FeatureSource, o *Options, base b6.World) error {
	return build(source, base, o, o.Output())
}
