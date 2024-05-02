package ingest

import (
	"context"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/osm"
)

func buildBasicWorld(nodes []osm.Node, ways []osm.Way, relations []osm.Relation, o *BuildOptions) (b6.World, error) {
	return BuildWorldFromOSM(nodes, ways, relations, o)
}

func buildOverlayWorld(nodes []osm.Node, ways []osm.Way, relations []osm.Relation, o *BuildOptions) (b6.World, error) {
	// TODO: Interleave the elements to make two different worlds?
	emptyWorld, err := buildBasicWorld([]osm.Node{}, []osm.Way{}, []osm.Relation{}, o)
	if err != nil {
		return nil, err
	}
	basic, err := buildBasicWorld(nodes, ways, relations, o)
	if err != nil {
		return nil, err
	}
	return NewOverlayWorld(emptyWorld, basic), nil
}

func buildBasicMutableWorld(nodes []osm.Node, ways []osm.Way, relations []osm.Relation, o *BuildOptions) (b6.World, error) {
	return BuildMutableWorldFromOSM(nodes, ways, relations, o)
}

func buildMutableOverlayWorld(nodes []osm.Node, ways []osm.Way, relations []osm.Relation, o *BuildOptions) (b6.World, error) {
	// TODO: Interleave the elements to make two different worlds?
	basic, err := buildBasicWorld(nodes, ways, relations, o)
	if err != nil {
		return nil, err
	}
	return NewMutableOverlayWorld(basic), nil
}

func buildMutableOverlayWorldOnBasic(nodes []osm.Node, ways []osm.Way, relations []osm.Relation, o *BuildOptions) (b6.World, error) {
	// TODO: Interleave the elements to make two different worlds?
	basic, err := buildBasicWorld(nodes, ways, relations, o)
	if err != nil {
		return nil, err
	}
	w := NewMutableOverlayWorld(basic)

	osmSource := MemoryOSMSource{Nodes: nodes, Ways: ways, Relations: relations}
	source, err := NewFeatureSourceFromPBF(&osmSource, o, context.Background())
	if err != nil {
		return nil, err
	}

	emit := func(feature Feature, g int) error {
		w.AddFeature(feature)
		return nil
	}
	options := ReadOptions{Goroutines: 2}
	if err := source.Read(options, emit, context.Background()); err != nil {
		return nil, err
	}
	return w, nil
}

var worldBuilders = []struct {
	name    string
	builder BuildOSMWorld
}{
	{"Basic", buildBasicWorld},
	{"Overlay", buildOverlayWorld},
	{"MutableOverlay", buildMutableOverlayWorld},
	// Fill a basic world with the given features, then add them all again in an overlay
	{"MutableOverlayOnBasic", buildMutableOverlayWorldOnBasic},
}

func TestWorlds(t *testing.T) {
	for _, builder := range worldBuilders {
		ValidateWorld(builder.name, builder.builder, t)
	}
}
