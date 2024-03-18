package compact

import (
	"context"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/osm"
	"diagonal.works/b6/test"

	"github.com/golang/geo/s2"
)

func mergeOSM(nodes []osm.Node, ways []osm.Way, relations []osm.Relation, base b6.World, w *World, o *ingest.BuildOptions) error {
	osmSource := ingest.MemoryOSMSource{Nodes: nodes, Ways: ways, Relations: relations}
	source, err := ingest.NewFeatureSourceFromPBF(&osmSource, o, context.Background())
	if err != nil {
		return err
	}
	options := Options{Goroutines: 2, PointsScratchOutputType: OutputTypeMemory}
	var index []byte
	if base == nil {
		if index, err = BuildInMemory(source, &options); err != nil {
			return err
		}
	} else {
		if index, err = BuildOverlayInMemory(source, &options, base); err != nil {
			return err
		}
	}
	return w.Merge(index)
}

func TestValidateWorld(t *testing.T) {
	build := func(nodes []osm.Node, ways []osm.Way, relations []osm.Relation, o *ingest.BuildOptions) (b6.World, error) {
		w := NewWorld()
		return w, mergeOSM(nodes, ways, relations, nil, w, o)
	}
	ingest.ValidateWorld("Compact", build, t)
}

func TestPathSegmentsWithSameNamespaceInMultipleBlocks(t *testing.T) {
	nodes := []osm.Node{
		{
			ID:       1715968751,
			Location: osm.LatLng{Lat: 51.5354079, Lng: -0.1244521},
			Tags:     []osm.Tag{{Key: "entrance", Value: "yes"}},
		},
		{
			ID:       2309943825,
			Location: osm.LatLng{Lat: 51.5354848, Lng: -0.1243698},
			Tags:     []osm.Tag{{Key: "entrance", Value: "yes"}},
		},
		{
			ID:       2309943870,
			Location: osm.LatLng{Lat: 51.5355393, Lng: -0.1247150},
			Tags:     []osm.Tag{{Key: "entrance", Value: "yes"}},
		},
	}

	ways := [2][]osm.Way{
		{
			osm.Way{
				ID:    159483602,
				Nodes: []osm.NodeID{1715968751, 2309943825},
				Tags:  []osm.Tag{{Key: "highway", Value: "path"}},
			},
		},
		{
			osm.Way{
				ID:    581001306,
				Nodes: []osm.NodeID{1715968751, 2309943870},
				Tags:  []osm.Tag{{Key: "highway", Value: "path"}},
			},
		},
	}

	w := NewWorld()
	if err := mergeOSM(nodes, ways[0], []osm.Relation{}, nil, w, &ingest.BuildOptions{Cores: 2}); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	if err := mergeOSM([]osm.Node{}, ways[1], []osm.Relation{}, w, w, &ingest.BuildOptions{Cores: 2}); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	id := ingest.FromOSMNodeID(nodes[0].ID)
	seen := make(map[osm.WayID]struct{})
	for _, s := range b6.AllPaths(w.byID.FindPathsByPoint(id)) {
		seen[osm.WayID(s.PathID().Value)] = struct{}{}
	}

	for _, ws := range ways {
		for _, w := range ws {
			if _, ok := seen[w.ID]; !ok {
				t.Errorf("Expected to find way %d", w.ID)
			}
		}
	}
}

// Compare the time taken to search a Flat world, and a mutable overlay world
// based on it

func mustBuildCamdenForBenchmarks() b6.World {
	pbf := ingest.PBFFilesOSMSource{Glob: test.Data(test.CamdenPBF), FailWhenNoFiles: true}
	source, err := ingest.NewFeatureSourceFromPBF(&pbf, &ingest.BuildOptions{Cores: 2}, context.Background())
	if err != nil {
		panic(err)
	}

	options := Options{Goroutines: 2, PointsScratchOutputType: OutputTypeMemory}
	var index []byte
	if index, err = BuildInMemory(source, &options); err != nil {
		panic(err)
	}

	w := NewWorld()
	if err := w.Merge(index); err != nil {
		panic(err)
	}
	return w
}

var benchmarkSearchQuery = b6.Intersection{
	b6.Keyed{Key: "#building"},
	b6.NewIntersectsCap(s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.5305, -0.1232)), b6.MetersToAngle(1000.0))),
}

func BenchmarkSearchWorld(b *testing.B) {
	w := mustBuildCamdenForBenchmarks()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		features := w.FindFeatures(benchmarkSearchQuery)
		for features.Next() {
		}
	}
}

func BenchmarkSearchModifiedWorld(b *testing.B) {
	w := mustBuildCamdenForBenchmarks()
	ids := make([]b6.FeatureID, 0)
	features := w.FindFeatures(benchmarkSearchQuery)
	for features.Next() {
		ids = append(ids, features.Feature().FeatureID())
	}
	mutable := ingest.NewMutableOverlayWorld(w)
	for _, id := range ids {
		mutable.AddTag(id, b6.Tag{Key: "#100m", Value: b6.String("yes")})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		features := mutable.FindFeatures(b6.Keyed{"#building"})
		for features.Next() {
		}
	}
}
