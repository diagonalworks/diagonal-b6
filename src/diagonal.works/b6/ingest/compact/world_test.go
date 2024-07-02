package compact

import (
	"context"
	"slices"
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

func TestPointPathExpressionTypesCorrectlyInferred(t *testing.T) {
	nodes := []osm.Node{
		{
			ID:       9663680708,
			Location: osm.LatLng{Lat: 58.2859244, Lng: -6.7920486},
			Tags:     []osm.Tag{},
		},
		{
			ID:       9663680707,
			Location: osm.LatLng{Lat: 58.2856653, Lng: -6.7915911},
			Tags:     []osm.Tag{},
		},
	}

	ways := [1][]osm.Way{
		{
			osm.Way{
				ID:    1051579980,
				Nodes: []osm.NodeID{9663680708, 9663680707},
				Tags:  []osm.Tag{{Key: "highway", Value: "path"}},
			},
		},
	}

	w := NewWorld()
	if err := mergeOSM(nodes, ways[0], []osm.Relation{}, nil, w, &ingest.BuildOptions{Cores: 2}); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	point := w.FindFeatureByID(ingest.FromOSMNodeID(9663680708))
	if point == nil {
		t.Fatalf("Expected to find point feature")
	}

	tags := point.AllTags()
	if len(tags) != 1 {
		t.Fatalf("Expected one tag got %d", len(tags))
	}
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
	for _, s := range b6.AllFeatures(w.byID.FindReferences(id, b6.FeatureTypePath)) {
		seen[osm.WayID(s.FeatureID().Value)] = struct{}{}
	}

	for _, ws := range ways {
		for _, w := range ws {
			if _, ok := seen[w.ID]; !ok {
				t.Errorf("Expected to find way %d", w.ID)
			}
		}
	}
}

func TestAreasForPointWithSameNamespaceInMultipleBlocks(t *testing.T) {
	londonNodes := []osm.Node{
		{
			ID:       4270651271,
			Location: osm.LatLng{Lat: 51.5354065, Lng: -0.1243874},
		},
		{
			ID:       1715968760,
			Location: osm.LatLng{Lat: 51.5353655, Lng: -0.1244049},
		},
		{
			ID:       2309943806,
			Location: osm.LatLng{Lat: 51.5352053, Lng: -0.1245396},
		},
	}

	londonWays := []osm.Way{
		{
			ID:    427900370,
			Nodes: []osm.NodeID{4270651271, 1715968760, 2309943806, 4270651271},
			Tags:  []osm.Tag{{Key: "building", Value: "yes"}, {Key: "name", Value: "The Lighterman"}},
		},
	}

	manchesterNodes := []osm.Node{
		{
			ID:       1492899494,
			Location: osm.LatLng{Lat: 53.4792602, Lng: -2.2316735},
		},
		{
			ID:       1492899361,
			Location: osm.LatLng{Lat: 53.4788548, Lng: -2.2312303},
		},
		{
			ID:       1492899446,
			Location: osm.LatLng{Lat: 53.4791165, Lng: -2.2305658},
		},
	}

	manchesterWays := []osm.Way{
		{
			ID:    136038212,
			Nodes: []osm.NodeID{1492899494, 1492899361, 1492899446, 1492899494},
			Tags:  []osm.Tag{{Key: "building", Value: "yes"}, {Key: "name", Value: "Ducie Street Warehouse"}},
		},
	}

	w := NewWorld()
	if err := mergeOSM(londonNodes, londonWays, []osm.Relation{}, nil, w, &ingest.BuildOptions{Cores: 2}); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	if err := mergeOSM(manchesterNodes, manchesterWays, []osm.Relation{}, nil, w, &ingest.BuildOptions{Cores: 2}); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	areas := w.FindAreasByPoint(ingest.FromOSMNodeID(londonNodes[0].ID))
	ids := make([]b6.FeatureID, 0)
	for areas.Next() {
		ids = append(ids, areas.FeatureID())
	}
	if !slices.Equal(ids, []b6.FeatureID{ingest.AreaIDFromOSMWayID(londonWays[0].ID).FeatureID()}) {
		t.Errorf("Expected to find London area")
	}

	areas = w.FindAreasByPoint(ingest.FromOSMNodeID(manchesterNodes[0].ID))
	ids = ids[0:0]
	for areas.Next() {
		ids = append(ids, areas.FeatureID())
	}
	if !slices.Equal(ids, []b6.FeatureID{ingest.AreaIDFromOSMWayID(manchesterWays[0].ID).FeatureID()}) {
		t.Errorf("Expected to find Manchester area")
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
		mutable.AddTag(id, b6.Tag{Key: "#100m", Value: b6.NewStringExpression("yes")})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		features := mutable.FindFeatures(b6.Keyed{"#building"})
		for features.Next() {
		}
	}
}
