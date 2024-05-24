package ingest

import (
	"fmt"
	"log"
	"sort"
	"sync/atomic"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/osm"
	"diagonal.works/b6/test"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

type BuildOSMWorld func(nodes []osm.Node, ways []osm.Way, relations []osm.Relation, o *BuildOptions) (b6.World, error)

func ValidateWorld(name string, buildWorld BuildOSMWorld, t *testing.T) {
	tests := []struct {
		name string
		f    func(BuildOSMWorld, *testing.T)
	}{
		{"GranarySquareSeemsReasonable", ValidateGranarySquareSeemsReasonable},
		{"WaysWithMissingNodesArentIndexed", ValidateWaysWithMissingNodesArentIndexed},
		{"WaysFallingExactlyWithinSearchCellAreFound", ValidateWaysFallingExactlyWithinSearchCellAreFound},
		{"ClockwisePolygonsAreIndexedCorrectly", ValidateClockwisePolygonsAreIndexedCorrectly},
		{"PolygonsWithInvalidGeometryAreSkipped", ValidatePolygonsWithInvalidGeometryAreSkipped},
		{"WaysAsAreas", ValidateWaysAsAreas},
		{"RelationsAsAreas", ValidateRelationsAsAreas},
		{"RelationsAsAreasWithMultipleLoopsAreArrangedCorrectly", ValidateRelationsAsAreasWithMultipleLoopsAreArrangedCorrectly},
		{"AllQueryOnATokenThatDoesntExistReturnsNothing", ValidateAllQueryOnATokenThatDoesntExistReturnsNothing},
		{"PointsNotOnAPathDoesntReturnPaths", ValidatePointsNotOnAPathDoesntReturnPaths},
		{"FindFeatureWithInvalidIDReturnsNil", ValidateFindFeatureWithInvalidIDReturnsNil},
		{"FindWithIntersectionQuery", ValidateFindWithIntersectionQuery},
		{"TraverseReturnsSegmentsWithCorrectOrigin", ValidateTraverseReturnsSegmentsWithCorrectOrigin},
		{"TraverseAlongWaysThatHaveBeenInverted", ValidateTraverseAlongWaysThatHaveBeenInverted},
		{"FindPathsWithTwoJoinedPaths", ValidateFindPathsWithTwoJoinedPaths},
		{"TraverseByIntersectionsAtEndNodes", ValidateTraverseByIntersectionsAtEndNodes},
		{"TraverseWithoutIntersectionsAtEndNodes", ValidateTraverseWithoutIntersectionsAtEndNodes},
		{"TraverseByIntersectionsBetweenEndNodes", ValidateTraverseByIntersectionsBetweenEndNodes},
		{"SegmentsAreOnlyReturnedForWaysWithAllNodesPresent", ValidateSegmentsAreOnlyReturnedForWaysWithAllNodesPresent},
		{"TraversalCanEndAtTaggedNodes", ValidateTraversalCanEndAtTaggedNodes},
		{"FindAreasByPoint", ValidateFindAreasByPoint},
		{"ValidateFindRelationsByFeature", ValidateFindRelationsByFeature},
		{"SpatialQueriesOnAnEmptyIndexReturnNothing", ValidateSpatialQueriesOnAnEmptyIndexReturnNothing},
		{"SpatialQueriesRecallParentCells", ValidateSpatialQueriesRecallParentCells},
		{"FindRelationWithMissingWay", ValidateFindRelationWithMissingWay},
		{"PathsAreExplicityClosedLoopsArent", ValidatePathsAreExplicityClosedLoopsArent},
		{"PolygonForAreaWithAHole", ValidatePolygonForAreaWithAHole},
		{"TagsAreSearchable", ValidateTagsAreSearchable},
		{"MultipolygonsWithLoopsThatSharePoints", ValidateMultipolygonsWithLoopsThatSharePoints},
		{"ThinBuilding", ValidateThinBuilding},
		{"BrokenOSMWayForArea", ValidateBrokenOSMWayForArea},
		{"SearchableRelationsNotReferencedByAFeature", ValidateSearchableRelationsNotReferencedByAFeature},
		{"EachFeature", ValidateEachFeature},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s/%s", name, test.name), func(t *testing.T) { test.f(buildWorld, t) })
	}
}

func ValidateGranarySquareSeemsReasonable(buildWorld BuildOSMWorld, t *testing.T) {
	nodes, ways, relations, err := osm.ReadWholePBF(test.Data(test.GranarySquarePBF))
	if err != nil {
		t.Errorf("Failed to read world: %s", err)
		return
	}
	w, err := buildWorld(nodes, ways, relations, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.53534, -0.12447)), b6.MetersToAngle(100))
	found := w.FindFeatures(b6.MightIntersect{cap})
	buildings := 0
	streets := 0

	features := make(map[b6.FeatureID]b6.Feature)
	for found.Next() {
		feature := found.Feature()
		features[feature.FeatureID()] = feature
		if tag := feature.Get("#building"); tag.IsValid() {
			buildings++
		} else if tag := feature.Get("#highway"); tag.IsValid() {
			streets++
		}
	}

	if feature, ok := features[AreaIDFromOSMWayID(222021576).FeatureID()]; ok {
		if area, ok := feature.(b6.AreaFeature); ok {
			if area.Len() == 1 {
				paths := area.Feature(0)
				if len(paths) == 1 {
					expectedPoints := 6
					if paths[0].GeometryLen() != expectedPoints {
						t.Errorf("Expected %d nodes for The Granary, found %d", expectedPoints, paths[0].GeometryLen())
					}
					for i, r := range paths[0].References() {
						if w.FindFeatureByID(r.Source()) == nil {
							t.Errorf("Expected a PointFeature at index %d, found nil", i)
						}
					}
				} else {
					t.Errorf("Expected 1 path, found %d", len(paths))
				}
			} else {
				t.Errorf("Expected 1 polygon, found %d", area.Len())
			}
		} else {
			t.Errorf("Expected an AreaFeature, found %T", area)
		}
	} else {
		t.Errorf("Expected to find The Granary, openstreetmap.org/way/222021576")
	}

	expectedBuildings := 9
	expectedStreets := 76
	if buildings != expectedBuildings {
		t.Errorf("Expected %d buildings, found %d", expectedBuildings, buildings)
	}
	if streets != expectedStreets {
		t.Errorf("Expected %d streets, found %d", expectedStreets, streets)
	}

	wayID := AreaIDFromOSMWayID(222021576).FeatureID()
	way := w.FindFeatureByID(wayID)
	if way == nil {
		t.Errorf("Expected to find way %s", wayID)
	} else {
		if way.FeatureID() != wayID {
			t.Errorf("Expected ID %s, found %s", wayID, way.FeatureID())
		}
	}

	wayID = AreaIDFromOSMWayID(42).FeatureID() // Doesn't exist
	way = w.FindFeatureByID(wayID)
	if way != nil {
		t.Errorf("Expected not to find any way, found %v", way)
	}

	relationID := FromOSMRelationID(8905542).FeatureID()
	relation := w.FindFeatureByID(relationID)
	if relation == nil {
		t.Errorf("Expected to find relation %s", relationID)
	} else {
		if relation.FeatureID() != relationID {
			t.Errorf("Expected ID %s, found %s", relationID, relation.FeatureID())
		}
	}

	relationID = FromOSMRelationID(42).FeatureID() // Doesn't exist
	relation = w.FindFeatureByID(relationID)
	if relation != nil {
		t.Errorf("Expected not to find any relation, found %s", relation.FeatureID())

	}

	expectedTrees := 26
	expectedBenches := 12
	found = w.FindFeatures(b6.MightIntersect{cap})

	benches := 0
	trees := 0
	for found.Next() {
		feature := found.Feature()
		if feature.Get("#amenity").Value.String() == "bench" {
			benches++
		}
		if feature.Get("#natural").Value.String() == "tree" {
			trees++
		}
	}

	if trees != expectedTrees {
		t.Errorf("Expected %d trees, found %d", expectedTrees, trees)
	}
	if benches != expectedBenches {
		t.Errorf("Expected %d benches, found %d", expectedBenches, benches)
	}

	seen := make(map[string]struct{})
	for _, token := range w.Tokens() {
		if _, ok := seen[token]; !ok {
			seen[token] = struct{}{}
		} else {
			t.Errorf("Duplicate token %q", token)
		}
	}

	found = w.FindFeatures(b6.Intersection{b6.Keyed{"#highway"}, b6.MightIntersect{cap}})
	if n := len(b6.AllFeatures(found)); n != streets {
		t.Errorf("Expected streets for intersection search to equal streets from filter (%d vs %d)", n, streets)
	}
}

func ValidateWaysWithMissingNodesArentIndexed(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		osm.Node{ID: 5378333625, Location: osm.LatLng{Lat: 51.5352195, Lng: -0.1254286}},
		osm.Node{ID: 1715968743, Location: osm.LatLng{Lat: 51.5351547, Lng: -0.1250628}},
		// Missing: 1715968739
		osm.Node{ID: 1715968738, Location: osm.LatLng{Lat: 51.5351015, Lng: -0.1248611}},

		osm.Node{ID: 1447052073, Location: osm.LatLng{Lat: 51.5350326, Lng: -0.1247915}},
		osm.Node{ID: 1540349979, Location: osm.LatLng{Lat: 51.5348204, Lng: -0.1246405}},
	}

	ways := []osm.Way{
		osm.Way{
			ID:    642639444,
			Nodes: []osm.NodeID{5378333625, 1715968743, 1715968739, 1715968738},
		},
		osm.Way{
			ID:    140633010,
			Nodes: []osm.NodeID{1447052073, 1540349979},
		},
	}
	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}
	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.53534, -0.12447)), b6.MetersToAngle(100))
	paths := b6.AllFeatures(w.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.NewIntersectsCap(cap)}))

	expected := 1
	if len(paths) != expected {
		t.Errorf("Expected %d ways, found %d", expected, len(paths))
	}
}

func ValidatePointsWithoutTagsArentIndexed(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		osm.Node{
			ID:       5378333625,
			Location: osm.LatLng{Lat: 51.5352195, Lng: -0.1254286},
			Tags:     []osm.Tag{{Key: "barrier", Value: "bollard"}},
		},
		osm.Node{ID: 1715968743, Location: osm.LatLng{Lat: 51.5351547, Lng: -0.1250628}},
	}

	w, err := buildWorld(nodes, []osm.Way{}, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}
	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.53534, -0.12447)), b6.MetersToAngle(100))
	points := w.FindFeatures(b6.NewIntersectsCap(cap))

	if !points.Next() || points.FeatureID().Value != uint64(nodes[0].ID) {
		t.Errorf("Expected ID %d, found %d", nodes[0].ID, points.FeatureID().Value)
	}

	if points.Next() {
		t.Errorf("Expected 1 point, found more")
	}
}

func ValidateWaysFallingExactlyWithinSearchCellAreFound(buildWorld BuildOSMWorld, t *testing.T) {
	cell := s2.CellFromCellID(s2.CellIDFromToken("48761b3dec")) // Part of granary square
	nodes := make([]osm.Node, 5)
	way := osm.Way{ID: 1, Nodes: make([]osm.NodeID, 5)}
	for i := 0; i < 5; i++ {
		ll := s2.LatLngFromPoint(cell.Vertex(i % 4))
		nodes[i] = osm.Node{ID: osm.NodeID(i), Location: osm.LatLng{Lat: ll.Lat.Degrees(), Lng: ll.Lng.Degrees()}}
		way.Nodes[i] = osm.NodeID(i)
	}

	w, err := buildWorld(nodes, []osm.Way{way}, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}
	paths := b6.AllFeatures(w.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.NewIntersectsCell(cell)}))

	expected := 1
	if len(paths) != expected {
		t.Errorf("Expected %d ways, found %d", expected, len(paths))
	}
}

func ValidateClockwisePolygonsAreIndexedCorrectly(buildWorld BuildOSMWorld, t *testing.T) {
	// https://www.openstreetmap.org/way/222021570 is an example of a way
	// representing a polygon that is ordered clockwise, and needs to be
	// reordered counterclockwise.
	nodes := []osm.Node{
		osm.Node{ID: 2309943870, Location: osm.LatLng{Lat: 51.5371371, Lng: -0.1240464}},
		osm.Node{ID: 2309943869, Location: osm.LatLng{Lat: 51.5370778, Lng: -0.1236840}},
		osm.Node{ID: 2309943825, Location: osm.LatLng{Lat: 51.5354848, Lng: -0.1243698}},
		osm.Node{ID: 2309943835, Location: osm.LatLng{Lat: 51.5355393, Lng: -0.1247150}},
	}

	way := osm.Way{
		ID:    222021570,
		Nodes: []osm.NodeID{2309943870, 2309943869, 2309943825, 2309943835, 2309943870},
	}

	w, err := buildWorld(nodes, []osm.Way{way}, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}
	paths := b6.AllFeatures(w.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.MightIntersect{s2.CellFromLatLng(s2.LatLngFromDegrees(51.53634, -0.12422))}}))

	if len(paths) == 1 {
		if paths[0].(b6.PhysicalFeature).PointAt(1).Distance(s2.PointFromLatLng(nodes[1].Location.ToS2LatLng())) < s1.Angle(0.00001) {
			t.Errorf("Expected nodes to be reversed")
		}
	} else {
		t.Errorf("Expected to find 1 path, found %d", len(paths))
	}
}

func ValidatePolygonsWithInvalidGeometryAreSkipped(buildWorld BuildOSMWorld, t *testing.T) {
	// This (now deleted) way caused indexing to crash, as the geometry is invalid.
	nodes := []osm.Node{
		osm.Node{ID: 5790639824, Location: osm.LatLng{Lat: 33.5526155, Lng: -0.2648717}},
		osm.Node{ID: 5790639825, Location: osm.LatLng{Lat: 33.5525395, Lng: -0.2649040}},
	}

	way := osm.Way{
		ID:    611666432,
		Nodes: []osm.NodeID{5790639824, 5790639825, 5790639824},
		Tags:  []osm.Tag{{Key: "building", Value: "yes"}},
	}

	w, err := buildWorld(nodes, []osm.Way{way}, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}
	paths := b6.AllAreas(b6.FindAreas(b6.NewIntersectsCell(s2.CellFromLatLng(s2.LatLngFromDegrees(33.5526155, -0.2648717))), w))

	if len(paths) != 0 {
		t.Errorf("Expected to find no paths, found %d", len(paths))
	}
}

func ValidateWaysAsAreas(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		osm.Node{ID: 2309943870, Location: osm.LatLng{Lat: 51.5371371, Lng: -0.1240464}},
		osm.Node{ID: 2309943825, Location: osm.LatLng{Lat: 51.5354848, Lng: -0.1243698}},
		osm.Node{ID: 2309943835, Location: osm.LatLng{Lat: 51.5355393, Lng: -0.1247150}},

		osm.Node{ID: 598093309, Location: osm.LatLng{Lat: 51.5321649, Lng: -0.1269834}},
		osm.Node{ID: 598093325, Location: osm.LatLng{Lat: 51.5343194, Lng: -0.1287286}},
		osm.Node{ID: 4765024403, Location: osm.LatLng{Lat: 51.5321777, Lng: -0.1269398}},
	}

	ways := []osm.Way{
		// Has a building tag, is closed
		osm.Way{
			ID:    222021570,
			Nodes: []osm.NodeID{2309943870, 2309943825, 2309943835, 2309943870},
			Tags:  []osm.Tag{{Key: "building", Value: "yes"}},
		},

		// Has an area tag, is closed
		osm.Way{
			ID:    109793493,
			Nodes: []osm.NodeID{598093309, 598093325, 4765024403, 598093309},
			Tags:  []osm.Tag{{Key: "area", Value: "yes"}},
		},
	}

	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	for _, way := range ways {
		area := b6.FindAreaByID(AreaIDFromOSMWayID(way.ID), w)
		if area == nil {
			t.Errorf("Expected to find an area for way %d", way.ID)
		}
	}
}

func ValidateRelationsAsAreas(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		osm.Node{ID: 5266979317, Location: osm.LatLng{Lat: 51.5366775, Lng: -0.1270333}},
		osm.Node{ID: 5266979315, Location: osm.LatLng{Lat: 51.5368503, Lng: -0.1267743}},
		osm.Node{ID: 5266979313, Location: osm.LatLng{Lat: 51.5366313, Lng: -0.1269323}},
	}

	ways := []osm.Way{
		osm.Way{
			ID:    544908185,
			Nodes: []osm.NodeID{5266979317, 5266979315, 5266979313, 5266979317},
		},
	}

	relations := []osm.Relation{
		osm.Relation{
			ID: 7972217,
			Members: []osm.Member{
				osm.Member{Type: osm.ElementTypeWay, ID: osm.AnyID(ways[0].ID)},
			},
			Tags: []osm.Tag{{Key: "building", Value: "yes"}, {Key: "type", Value: "multipolygon"}},
		},
	}

	w, err := buildWorld(nodes, ways, relations, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.53534, -0.12447)), b6.MetersToAngle(500))

	expectedAreas := 1
	q := b6.Intersection{b6.NewIntersectsCap(cap), b6.Tagged{Key: "#building", Value: b6.String("yes")}}
	if areas := b6.AllAreas(b6.FindAreas(q, w)); len(areas) != expectedAreas {
		t.Errorf("Expected %d area, found %d", expectedAreas, len(areas))
	} else {
		expectedPolygons := 1
		if areas[0].Len() != expectedPolygons {
			t.Errorf("Expected an areas with %d paths, found %d", expectedPolygons, areas[0].Len())
		} else {
			expectedLoops := 1
			if areas[0].Polygon(0).NumLoops() != expectedLoops {
				t.Errorf("Expected %d loop, found %d", expectedLoops, areas[0].Polygon(0).NumLoops())
			}
			areaM2 := b6.AreaToMeters2(areas[0].Polygon(0).Area())
			if areaM2 < 0 || areaM2 > 1000 {
				t.Errorf("Expected area between 0m2 and 1000m2, found %fm2", areaM2)
			}
		}

		relations := b6.AllRelations(w.FindRelationsByFeature(areas[0].FeatureID()))
		if len(relations) != 0 {
			t.Errorf("Expected no relations for feature, found %d", len(relations))
		}
	}
}

func ValidateRelationsAsAreasWithMultipleLoopsAreArrangedCorrectly(buildWorld BuildOSMWorld, t *testing.T) {
	nodes, ways, relations, err := osm.ReadWholePBF(test.Data(test.GranarySquarePBF))
	if err != nil {
		t.Errorf("Failed to read world: %s", err)
		return
	}
	w, err := buildWorld(nodes, ways, relations, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	area := b6.FindAreaByID(AreaIDFromOSMRelationID(7972217), w) // Gasholder apartments
	if area == nil {
		t.Errorf("Failed to find Gasholder apartments")
		return
	}

	expectedLength := 3
	if area.Len() != expectedLength {
		t.Errorf("Expected area of length %d, found %d", expectedLength, area.Len())
		return
	}

	for i := 0; i < area.Len(); i++ {
		if paths := area.Feature(i); len(paths) != 1 {
			t.Errorf("Expected feature %d to have 1 path, found %d", i, len(paths))
		}
	}
}

func ValidateAllQueryOnATokenThatDoesntExistReturnsNothing(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		{ID: 5266979317, Location: osm.LatLng{Lat: 51.5366775, Lng: -0.1270333}},
		{ID: 5266979315, Location: osm.LatLng{Lat: 51.5368503, Lng: -0.1267743}},
		{ID: 5266979313, Location: osm.LatLng{Lat: 51.5366313, Lng: -0.1269323}},
	}

	ways := []osm.Way{
		{
			ID:    544908185,
			Nodes: []osm.NodeID{5266979317, 5266979315, 5266979313, 5266979317},
		},
	}
	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	paths := b6.AllFeatures(w.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.Keyed{"#missing"}}))
	if len(paths) != 0 {
		t.Errorf("Expected to not find any paths, found %d", len(paths))
	}
}

func ValidatePointsNotOnAPathDoesntReturnPaths(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		osm.Node{ID: 2309943870, Location: osm.LatLng{Lat: 51.5371371, Lng: -0.1240464}},
		osm.Node{ID: 2309943825, Location: osm.LatLng{Lat: 51.5354848, Lng: -0.1243698}},
		osm.Node{ID: 2309943835, Location: osm.LatLng{Lat: 51.5355393, Lng: -0.1247150}},

		osm.Node{ID: 598093309, Location: osm.LatLng{Lat: 51.5321649, Lng: -0.1269834}},
		osm.Node{
			ID:       3838023409,
			Location: osm.LatLng{Lat: 51.5321649, Lng: -0.1269834},
			Tags: []osm.Tag{
				{Key: "shop", Value: "supermarket"},
				{Key: "name", Value: "Waitrose"},
			},
		},
	}

	ways := []osm.Way{
		osm.Way{
			ID:    222021570,
			Nodes: []osm.NodeID{2309943870, 2309943825, 2309943835, 2309943870},
			Tags:  []osm.Tag{{Key: "building", Value: "yes"}},
		},
	}

	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	for _, id := range []osm.NodeID{598093309, 3838023409} {
		if paths := b6.AllFeatures(w.FindReferences(FromOSMNodeID(id), b6.FeatureTypePath)); len(paths) != 0 {
			t.Errorf("Expected no paths for point %d, found %d", id, len(paths))
		}
		if segments := b6.AllSegments(w.Traverse(FromOSMNodeID(id))); len(segments) != 0 {
			t.Errorf("Expected no segments for point %d, found %d", id, len(segments))
		}
	}
}

func ValidateFindFeatureWithInvalidIDReturnsNil(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		osm.Node{
			ID:       3838023409,
			Location: osm.LatLng{Lat: 51.5321649, Lng: -0.1269834},
			Tags: []osm.Tag{
				{Key: "shop", Value: "supermarket"},
				{Key: "name", Value: "Waitrose"},
			},
		},
	}

	w, err := buildWorld(nodes, []osm.Way{}, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	f := w.FindFeatureByID(b6.FeatureID{
		Type:      b6.FeatureTypeInvalid,
		Namespace: b6.NamespaceOSMNode,
		Value:     uint64(nodes[0].ID),
	})

	if f != nil {
		t.Errorf("Expected nil, found %v", f)
	}
}

func ValidateFindWithIntersectionQuery(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		{
			ID:       1447052000,
			Location: osm.LatLng{Lat: 51.5351015, Lng: -0.1248611},
			Tags:     osm.Tags{{"building", "yes"}},
		},
		{
			ID:       1447052073,
			Location: osm.LatLng{Lat: 51.5350326, Lng: -0.1247915},
			Tags:     osm.Tags{{"amenity", "school"}},
		},
		{
			ID:       1540349979,
			Location: osm.LatLng{Lat: 51.5348204, Lng: -0.1246405},
			Tags:     osm.Tags{{"amenity", "school"}},
		},
		{
			ID:       1715968739,
			Location: osm.LatLng{Lat: 51.5351398, Lng: -0.1249654},
			Tags:     osm.Tags{{"amenity", "school"}, {"building", "yes"}},
		},
		{
			ID:       1715968755,
			Location: osm.LatLng{Lat: 51.5354037, Lng: -0.1260829},
			Tags:     osm.Tags{{"amenity", "school"}},
		},
		{
			ID:       4966136648,
			Location: osm.LatLng{Lat: 51.5348874, Lng: -0.1260855},
			Tags:     osm.Tags{{"building", "yes"}},
		},
		{
			ID:       5378333625,
			Location: osm.LatLng{Lat: 51.5352195, Lng: -0.1254286},
			Tags:     osm.Tags{{"building", "yes"}},
		},
	}

	w, err := buildWorld(nodes, []osm.Way{}, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	q := b6.Intersection{b6.Tagged{Key: "#amenity", Value: b6.String("school")}, b6.Tagged{Key: "#building", Value: b6.String("yes")}}
	fs := b6.AllFeatures(w.FindFeatures(q))
	if len(fs) != 1 {
		t.Errorf("Expected one feature, found %d", len(fs))
		return
	}

	expected := uint64(1715968739)
	if v := fs[0].FeatureID().Value; v != expected {
		t.Errorf("Expected node %d, found %d", expected, v)
		return
	}
}

func ValidateTraverseReturnsSegmentsWithCorrectOrigin(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		{ID: 5378333625, Location: osm.LatLng{Lat: 51.5352195, Lng: -0.1254286}},
		{ID: 1715968739, Location: osm.LatLng{Lat: 51.5351398, Lng: -0.1249654}},
		{ID: 1715968738, Location: osm.LatLng{Lat: 51.5351015, Lng: -0.1248611}},
		{ID: 4966136648, Location: osm.LatLng{Lat: 51.5348874, Lng: -0.1260855}},
		{ID: 5378333638, Location: osm.LatLng{Lat: 51.5367686, Lng: -0.1282862}},
		{ID: 7555184307, Location: osm.LatLng{Lat: 51.5373281, Lng: -0.1252851}},
		{ID: 1715968755, Location: osm.LatLng{Lat: 51.5354037, Lng: -0.1260829}},
		{ID: 1447052073, Location: osm.LatLng{Lat: 51.5350326, Lng: -0.1247915}},
		{ID: 1540349979, Location: osm.LatLng{Lat: 51.5348204, Lng: -0.1246405}},
	}

	ways := []osm.Way{
		{ID: 642639444, Nodes: []osm.NodeID{5378333625, 1715968739, 1715968738}},
		{ID: 557698825, Nodes: []osm.NodeID{5378333625, 4966136648, 5378333638}},
		{ID: 807925586, Nodes: []osm.NodeID{7555184307, 1715968755, 5378333625}},
	}

	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	origin := FromOSMNodeID(5378333625)
	segments := b6.AllSegments(w.Traverse(origin))

	expected := 3
	if len(segments) != expected {
		t.Errorf("Expected %d segments, found %d", expected, len(segments))
	}

	for _, segment := range segments {
		if segment.FirstFeatureID() != origin {
			t.Errorf("Expected first point to be %s, found %s", origin, segment.FirstFeatureID())
		}
	}
}

func ValidateTraverseAlongWaysThatHaveBeenInverted(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		osm.Node{ID: 2309943870, Location: osm.LatLng{Lat: 51.5371371, Lng: -0.1240464}},
		osm.Node{ID: 2309943825, Location: osm.LatLng{Lat: 51.5354848, Lng: -0.1243698}},
		osm.Node{ID: 2309943835, Location: osm.LatLng{Lat: 51.5355393, Lng: -0.1247150}},

		osm.Node{ID: 3790640851, Location: osm.LatLng{Lat: 51.5360342, Lng: -0.1257243}},
		osm.Node{ID: 3790640850, Location: osm.LatLng{Lat: 51.5358205, Lng: -0.1242138}},
	}

	ways := []osm.Way{
		osm.Way{
			ID:    222021570,
			Nodes: []osm.NodeID{2309943870, 2309943825, 2309943835, 2309943870},
			Tags:  []osm.Tag{{Key: "building", Value: "yes"}},
		},

		osm.Way{
			ID:    870512564,
			Nodes: []osm.NodeID{3790640851, 2309943825, 3790640850},
		},
	}

	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	origin := FromOSMNodeID(2309943825)
	for _, segment := range b6.AllSegments(w.Traverse(origin)) {
		if segment.FirstFeatureID() != origin {
			t.Errorf("Expected point %s, found %s", origin, segment.FirstFeatureID())
		}
	}
}

func ValidateFindPathsWithTwoJoinedPaths(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		{ID: 5384190463, Location: osm.LatLng{Lat: 51.5358664, Lng: -0.1272493}},
		{ID: 5384190494, Location: osm.LatLng{Lat: 51.5362126, Lng: -0.1270125}},
		{ID: 5384190476, Location: osm.LatLng{Lat: 51.5367563, Lng: -0.1266297}},
	}

	// Two ways, one ending with node 5384190494, one starting with it.
	ways := []osm.Way{
		{ID: 558345071, Nodes: []osm.NodeID{5384190463, 5384190494}},
		{ID: 558345054, Nodes: []osm.NodeID{5384190494, 5384190476}},
	}

	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	paths := b6.AllFeatures(w.FindReferences(FromOSMNodeID(5384190494), b6.FeatureTypePath))
	expected := 2
	if len(paths) != expected {
		t.Errorf("Expected %d segments, found %d", expected, len(paths))
	}
}

func ValidateTraverseByIntersectionsAtEndNodes(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		{ID: 5378333625, Location: osm.LatLng{Lat: 51.5352195, Lng: -0.1254286}},
		{ID: 1715968739, Location: osm.LatLng{Lat: 51.5351398, Lng: -0.1249654}},
		{ID: 1715968738, Location: osm.LatLng{Lat: 51.5351015, Lng: -0.1248611}},
		{ID: 4966136648, Location: osm.LatLng{Lat: 51.5348874, Lng: -0.1260855}},
		{ID: 5378333638, Location: osm.LatLng{Lat: 51.5367686, Lng: -0.1282862}},
		{ID: 7555184307, Location: osm.LatLng{Lat: 51.5373281, Lng: -0.1252851}},
		{ID: 1715968755, Location: osm.LatLng{Lat: 51.5354037, Lng: -0.1260829}},
		{ID: 1447052073, Location: osm.LatLng{Lat: 51.5350326, Lng: -0.1247915}},
		{ID: 1540349979, Location: osm.LatLng{Lat: 51.5348204, Lng: -0.1246405}},
	}

	ways := []osm.Way{
		{ID: 642639444, Nodes: []osm.NodeID{5378333625, 1715968739, 1715968738}},
		{ID: 557698825, Nodes: []osm.NodeID{5378333625, 4966136648, 5378333638}},
		{ID: 807925586, Nodes: []osm.NodeID{7555184307, 1715968755, 5378333625}},
		// Not connected to the above ways
		{ID: 140633010, Nodes: []osm.NodeID{1447052073, 1540349979}},
	}

	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	path := w.FindFeatureByID(FromOSMWayID(557698825))
	if path == nil {
		t.Errorf("Test data incorrect: failed to find way")
		return
	}

	point := path.Reference(0).Source()
	if !point.IsValid() {
		t.Errorf("Expected a feature at index 0")
	}
	found := b6.AllSegments(w.Traverse(point))
	expected := []uint64{557698825, 642639444, 807925586}
	if len(found) != len(expected) {
		t.Errorf("Expected %d paths, found %d", len(expected), len(found))
		return
	}

	ids := make(b6.FeatureIDs, len(found))
	for i, route := range found {
		ids[i] = route.Feature.FeatureID()
	}
	sort.Sort(ids)
	for i := range ids {
		if ids[i].Value != expected[i] {
			t.Errorf("Expected path %d, found %d", expected[i], ids[i].Value)
		}
	}

	point = path.Reference(1).Source()
	if !point.IsValid() {
		t.Errorf("Expected a PointFeature at index 1")
	}
	found = b6.AllSegments(w.Traverse(point))
	if len(found) != 2 {
		t.Errorf("Expected 2 path segments, found %d", len(found))
	}
}

func ValidateTraverseWithoutIntersectionsAtEndNodes(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		{ID: 1447052073, Location: osm.LatLng{Lat: 51.5350326, Lng: -0.1247915}},
		{ID: 1540349979, Location: osm.LatLng{Lat: 51.5348204, Lng: -0.1246405}},
	}

	ways := []osm.Way{
		{ID: 140633010, Nodes: []osm.NodeID{1447052073, 1540349979}},
	}

	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	path := w.FindFeatureByID(FromOSMWayID(140633010))
	if path == nil {
		t.Errorf("Test data incorrect: failed to find way")
		return
	}

	point := path.Reference(0).Source()
	if !point.IsValid() {
		t.Errorf("Expected a feature at index 0")
	}
	found := b6.AllSegments(w.Traverse(point))
	if len(found) != 1 {
		t.Errorf("Expected 1 segment, found %d", len(found))
		return
	}
	expectedID := uint64(140633010)
	if found[0].Feature.FeatureID().Value != expectedID {
		t.Errorf("Expected %d, found %d", expectedID, found[0].Feature.FeatureID().Value)
	}
}

type bySegmentKey []b6.Segment

func (p bySegmentKey) Len() int      { return len(p) }
func (p bySegmentKey) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p bySegmentKey) Less(i, j int) bool {
	return p[i].ToKey().Less(p[j].ToKey())
}

func ValidateTraverseByIntersectionsBetweenEndNodes(buildWorld BuildOSMWorld, t *testing.T) {
	// TODO: expand this test once we read real source data in from a geojson file,
	// specifically, include tests in which an edge is returned that terminates at
	// another intersection, rather than the end of the way
	nodes := []osm.Node{
		{ID: 5378333625, Location: osm.LatLng{Lat: 51.5352195, Lng: -0.1254286}},
		{ID: 5384190491, Location: osm.LatLng{Lat: 51.5352339, Lng: -0.1255240}},
		{ID: 7555184305, Location: osm.LatLng{Lat: 51.5352897, Lng: -0.1258430}},

		{ID: 4966136655, Location: osm.LatLng{Lat: 51.5349570, Lng: -0.1256696}},
	}

	ways := []osm.Way{
		{ID: 807925586, Nodes: []osm.NodeID{5378333625, 5384190491, 7555184305}},
		{ID: 558345068, Nodes: []osm.NodeID{5384190491, 4966136655}},
	}

	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	path := w.FindFeatureByID(FromOSMWayID(807925586))
	if path == nil {
		t.Errorf("Test data incorrect: failed to find way")
		return
	}

	point := path.Reference(1).Source()
	if !point.IsValid() {
		t.Errorf("Expected a feature at index 1")
	}

	found := b6.AllSegments(w.Traverse(point))
	sort.Sort(bySegmentKey(found))
	expected := []struct {
		way   osm.WayID
		first int
		last  int
	}{
		{558345068, 0, 1},
		{807925586, 1, 0},
		{807925586, 1, 2},
	}

	if len(found) != len(expected) {
		t.Errorf("Expected %d segments, found %d", len(expected), len(found))
		return
	}
	for i := range expected {
		if FromOSMWayID(expected[i].way) != found[i].Feature.FeatureID() {
			t.Errorf("Expected %d, found %s", expected[i].way, found[i].Feature.FeatureID())
		}
		if expected[i].first != found[i].First {
			t.Errorf("Expected first index %d on way %d, found %d", expected[i].first, expected[i].way, found[i].First)
		}
		if expected[i].last != found[i].Last {
			t.Errorf("Expected last index %d on way %d, found %d", expected[i].last, expected[i].way, found[i].Last)
		}
	}
}

func ValidateSegmentsAreOnlyReturnedForWaysWithAllNodesPresent(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		{ID: 5378333625, Location: osm.LatLng{Lat: 51.5352195, Lng: -0.1254286}},
		{ID: 5384190491, Location: osm.LatLng{Lat: 51.5352339, Lng: -0.1255240}},

		{ID: 4966136655, Location: osm.LatLng{Lat: 51.5349570, Lng: -0.1256696}},
	}

	ways := []osm.Way{
		{ID: 807925586, Nodes: []osm.NodeID{5378333625, 5384190491, 7555184305}},
		{ID: 558345068, Nodes: []osm.NodeID{5384190491, 4966136655}},
	}

	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	path := w.FindFeatureByID(FromOSMWayID(558345068))
	if path == nil {
		t.Errorf("Test data incorrect: failed to find way")
		return
	}

	point := path.Reference(0).Source() // This node is an intersection with way 807925586, which is missing a node
	if !point.IsValid() {
		t.Errorf("Expected a feature at index 0")
	}

	found := b6.AllSegments(w.Traverse(point))
	if len(found) != 1 || found[0].Feature.FeatureID().Value != uint64(558345068) {
		t.Errorf("Expected to find a single segment")
	}
}

func ValidateTraversalCanEndAtTaggedNodes(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		{ID: 4966135914, Location: osm.LatLng{Lat: 51.5355578, Lng: -0.1282138}},
		{
			ID:       4966135915,
			Location: osm.LatLng{Lat: 51.5355795, Lng: -0.1281742},
			Tags:     []osm.Tag{{Key: "barrier", Value: "gate"}},
		},
		{ID: 9306328667, Location: osm.LatLng{Lat: 51.5356298, Lng: -0.1282285}},
	}

	ways := []osm.Way{
		osm.Way{
			ID:    507018236,
			Nodes: []osm.NodeID{4966135914, 4966135915, 9306328667},
			Tags:  []osm.Tag{{Key: "bridge", Value: "yes"}},
		},
	}

	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	origin := FromOSMNodeID(nodes[0].ID)
	segments := b6.AllSegments(w.Traverse(origin))
	if len(segments) == 1 {
		if id := segments[0].LastFeatureID().Value; id != uint64(nodes[1].ID) {
			t.Errorf("Expected segment to end at a barrier, not node %d", id)
		}
	} else {
		t.Errorf("Expected to find a single segment, found %d", len(segments))
	}
}

func ValidateFindAreasByPoint(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		osm.Node{ID: 2309943870, Location: osm.LatLng{Lat: 51.5371371, Lng: -0.1240464}},
		osm.Node{ID: 2309943825, Location: osm.LatLng{Lat: 51.5354848, Lng: -0.1243698}},
		osm.Node{ID: 2309943835, Location: osm.LatLng{Lat: 51.5355393, Lng: -0.1247150}},

		osm.Node{ID: 598093309, Location: osm.LatLng{Lat: 51.5321649, Lng: -0.1269834}},
	}

	ways := []osm.Way{
		osm.Way{
			ID:    222021570,
			Nodes: []osm.NodeID{2309943870, 2309943825, 2309943835, 2309943870},
			Tags:  []osm.Tag{{Key: "building", Value: "yes"}},
		},
	}

	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	areas := b6.AllAreas(w.FindAreasByPoint(FromOSMNodeID(2309943835)))
	if len(areas) == 1 {
		if areas[0].FeatureID().Value != uint64(ways[0].ID) {
			t.Errorf("Expected area %d, found %d", ways[0].ID, areas[0].FeatureID().Value)
		}
	} else {
		t.Errorf("Expected %d areas, found %d", 1, len(areas))
	}

	areas = b6.AllAreas(w.FindAreasByPoint(FromOSMNodeID(598093309)))
	if len(areas) != 0 {
		log.Printf("bad: %s", areas[0].FeatureID())
		t.Errorf("Expected no areas, found %d", len(areas))
	}
}

func ValidateFindRelationsByFeature(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		{ID: 5378333625, Location: osm.LatLng{Lat: 51.5352195, Lng: -0.1254286}},
		{ID: 5384190491, Location: osm.LatLng{Lat: 51.5352339, Lng: -0.1255240}},

		{ID: 4966136655, Location: osm.LatLng{Lat: 51.5349570, Lng: -0.1256696}},
	}
	ways := []osm.Way{
		{ID: 807925586, Nodes: []osm.NodeID{5378333625, 5384190491, 4966136655}},
		{ID: 558345068, Nodes: []osm.NodeID{5384190491, 4966136655}},
	}
	relations := []osm.Relation{
		{ID: 11139964,
			Members: []osm.Member{
				{Type: osm.ElementTypeWay, ID: 807925586},
			},
			Tags: []osm.Tag{
				{Key: "type", Value: "route"},
				{Key: "route", Value: "bicycle"},
				{Key: "network", Value: "lcn"},
			},
		},
	}

	w, err := buildWorld(nodes, ways, relations, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	found := b6.AllRelations(w.FindRelationsByFeature(FromOSMWayID(807925586).FeatureID()))
	if len(found) == 1 {
		expected := b6.Tag{Key: "#route", Value: b6.String("bicycle")}
		if route := found[0].Get("#route"); route != expected {
			t.Errorf("Expected tag value %q, found %q", expected, route)
		}
	} else {
		t.Errorf("Expected %d relations, found %d", 1, len(found))
	}

	found = b6.AllRelations(w.FindRelationsByFeature(FromOSMWayID(558345068).FeatureID()))
	if len(found) != 0 {
		t.Errorf("Expected 0 relations, found %d", len(found))
	}
}

func ValidateFindRelationWithMissingWay(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		{ID: 5378333625, Location: osm.LatLng{Lat: 51.5352195, Lng: -0.1254286}},
		{ID: 5384190491, Location: osm.LatLng{Lat: 51.5352339, Lng: -0.1255240}},

		{ID: 4966136655, Location: osm.LatLng{Lat: 51.5349570, Lng: -0.1256696}},
	}
	ways := []osm.Way{
		{ID: 807925586, Nodes: []osm.NodeID{5378333625, 5384190491, 4966136655}},
	}
	relations := []osm.Relation{
		{ID: 11139964,
			Members: []osm.Member{
				{Type: osm.ElementTypeWay, ID: 807925586},
				{Type: osm.ElementTypeWay, ID: 558345068}, // Missing
			},
			Tags: []osm.Tag{
				{Key: "type", Value: "route"},
				{Key: "route", Value: "bicycle"},
				{Key: "network", Value: "lcn"},
			},
		},
	}

	w, err := buildWorld(nodes, ways, relations, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	if relation := b6.FindRelationByID(FromOSMRelationID(11139964), w); relation != nil {
		expectedLength := 2
		if relation.Len() == expectedLength {
			expectedIDs := []osm.WayID{807925586, 558345068}
			for i, id := range expectedIDs {
				expected := FromOSMWayID(id).FeatureID()
				if actual := relation.Member(i).ID; actual != expected {
					t.Errorf("Expected %s, found %s at index %d", expected, actual, i)
				}
			}
		} else {
			t.Errorf("Expected relation length %d, found %d", expectedLength, relation.Len())
		}
	} else {
		t.Errorf("Expected to find a relation")
	}
}

func ValidateSpatialQueriesOnAnEmptyIndexReturnNothing(buildWorld BuildOSMWorld, t *testing.T) {
	w, err := buildWorld([]osm.Node{}, []osm.Way{}, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.53534, -0.12447)), b6.MetersToAngle(500))
	query := b6.NewIntersectsCap(cap)
	if paths := b6.AllFeatures(w.FindFeatures(b6.Typed{b6.FeatureTypePath, query})); len(paths) != 0 {
		// The most likely failure mode is a panic()/nil pointer, rather than
		// imaginary ways, but both are covered.
		t.Errorf("Didn't expect to find any paths")
	}
	if points := w.FindFeatures(b6.Typed{b6.FeatureTypePoint, query}); points.Next() != false {
		t.Errorf("Didn't expect to find any points")
	}
}

func ValidateSpatialQueriesRecallParentCells(buildWorld BuildOSMWorld, t *testing.T) {
	// A bug in our generation of ancestor cell tokens prevented us from
	// adding a token for a cell when all children of that cell already had
	// tokens added. Test the fix that that here by indexing an area that exhibits
	// the problem, created from two arbitrary neighbouring S2 cells, one on top
	// of the other.
	top := s2.CellIDFromToken("48761af")
	topChildren := top.Children()
	bottom := top.EdgeNeighbors()[0]
	bottomChildren := bottom.Children()

	nodes := []osm.Node{
		{ID: 1, Location: osm.FromS2Point(s2.CellFromCellID(topChildren[3]).Center())},
		{ID: 2, Location: osm.FromS2Point(s2.CellFromCellID(topChildren[2]).Center())},
		{ID: 3, Location: osm.FromS2Point(s2.CellFromCellID(bottomChildren[2]).Center())},
		{ID: 4, Location: osm.FromS2Point(s2.CellFromCellID(bottomChildren[0]).Center())},
	}
	ways := []osm.Way{
		{ID: 1, Nodes: []osm.NodeID{1, 2, 3, 4, 1}, Tags: osm.Tags{{Key: "building", Value: "yes"}}},
	}
	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	searchRegion := s2.CellIDFromToken("48761ac")
	if !searchRegion.Contains(top) && !searchRegion.Contains(bottom) {
		t.Errorf("Expected search region to contain our test area")
		return
	}
	areas := b6.AllAreas(b6.FindAreas(b6.NewIntersectsCellID(searchRegion), w))
	if len(areas) != 1 || areas[0].AreaID().Value != uint64(ways[0].ID) {
		t.Errorf("Expected to find area")
	}
}

func ValidatePathsAreExplicityClosedLoopsArent(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		// Exterior loop, counterclockwise
		{ID: 1, Location: osm.LatLng{Lng: 0.0, Lat: 0.0}},
		{ID: 2, Location: osm.LatLng{Lng: 3.0, Lat: 0.0}},
		{ID: 3, Location: osm.LatLng{Lng: 3.0, Lat: 4.0}},
		{ID: 4, Location: osm.LatLng{Lng: 0.0, Lat: 4.0}},

		// Interior loop, counterclockwise
		{ID: 5, Location: osm.LatLng{Lng: 1.0, Lat: 2.0}},
		{ID: 6, Location: osm.LatLng{Lng: 2.0, Lat: 2.0}},
		{ID: 7, Location: osm.LatLng{Lng: 2.0, Lat: 3.0}},
		{ID: 8, Location: osm.LatLng{Lng: 1.0, Lat: 3.0}},
	}

	ways := []osm.Way{
		{ID: 1, Nodes: []osm.NodeID{1, 2, 3, 4, 1}},
		{ID: 2, Nodes: []osm.NodeID{5, 6, 7, 8, 5}},
	}

	relations := []osm.Relation{
		{ID: 1,
			Members: []osm.Member{
				{Type: osm.ElementTypeWay, ID: 1, Role: "outer"},
				{Type: osm.ElementTypeWay, ID: 2, Role: "inner"},
			},
			Tags: []osm.Tag{
				{Key: "type", Value: "multipolygon"},
			},
		},
	}

	w, err := buildWorld(nodes, ways, relations, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	area := b6.FindAreaByID(AreaIDFromOSMRelationID(1), w)
	if area == nil {
		t.Errorf("Expected to find an area, found none")
		return
	} else if area.Len() != 1 {
		t.Errorf("Expected to find area with 1 polygon, found %d", area.Len())
		return
	}

	// Paths should be explicitly closed, and their polylines too
	paths := area.Feature(0)
	if len(paths) != 2 {
		t.Errorf("Expected a polygon with 2 loops")
		return
	}
	for i, path := range paths {
		if path.GeometryLen() != 5 {
			t.Errorf("Expected each path to have 5 points, path %d has %d", i, path.GeometryLen())
		}
		if path.PointAt(0) != path.PointAt(path.GeometryLen()-1) {
			t.Errorf("Expected path to be explicitly closed")
		}
		points := *path.Polyline()
		if len(points) != 5 || points[0] != points[len(points)-1] {
			if path.PointAt(0) != path.PointAt(path.GeometryLen()-1) {
				t.Errorf("Expected polyline to have 5 points and be explicitly closed")
			}
		}
	}

	// Loops of the polygon shouldn't be explicitly closed
	polygon := area.Polygon(0)
	if polygon.NumLoops() != 2 {
		t.Errorf("Expected a polygon with 2 loops")
		return
	}
	for i, loop := range polygon.Loops() {
		if loop.NumVertices() != paths[i].GeometryLen()-1 {
			t.Errorf("Expected each loop to be implicited closed, loop %d wasn't", i)
		}
	}
}

func ValidatePolygonForAreaWithAHole(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		// Exterior loop, counterclockwise
		{ID: 1, Location: osm.LatLng{Lng: 0.0, Lat: 0.0}},
		{ID: 2, Location: osm.LatLng{Lng: 3.0, Lat: 0.0}},
		{ID: 3, Location: osm.LatLng{Lng: 3.0, Lat: 4.0}},
		{ID: 4, Location: osm.LatLng{Lng: 0.0, Lat: 4.0}},

		// Interior loop, counterclockwise
		{ID: 5, Location: osm.LatLng{Lng: 1.0, Lat: 2.0}},
		{ID: 6, Location: osm.LatLng{Lng: 2.0, Lat: 2.0}},
		{ID: 7, Location: osm.LatLng{Lng: 2.0, Lat: 3.0}},
		{ID: 8, Location: osm.LatLng{Lng: 1.0, Lat: 3.0}},
	}

	ways := []osm.Way{
		{ID: 1, Nodes: []osm.NodeID{1, 2, 3, 4, 1}},
		{ID: 2, Nodes: []osm.NodeID{5, 6, 7, 8, 5}},
	}

	relations := []osm.Relation{
		{ID: 1,
			Members: []osm.Member{
				{Type: osm.ElementTypeWay, ID: 1, Role: "outer"},
				{Type: osm.ElementTypeWay, ID: 2, Role: "inner"},
			},
			Tags: []osm.Tag{
				{Key: "type", Value: "multipolygon"},
			},
		},
	}

	w, err := buildWorld(nodes, ways, relations, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	area := b6.FindAreaByID(AreaIDFromOSMRelationID(1), w)
	if area == nil || area.Len() != 1 {
		t.Errorf("Expected to find area with 1 polygon")
		return
	}

	polygon := area.Polygon(0)
	if polygon.NumLoops() != 2 {
		t.Errorf("Expected polygon to have 2 loops")
		return
	}
	ll := s2.LatLngFromDegrees(1.0, 1.0)
	if !polygon.ContainsPoint(s2.PointFromLatLng(ll)) {
		t.Errorf("Expected %s to be inside the polygon", ll)
	}
	ll = s2.LatLngFromDegrees(2.5, 1.5)
	if polygon.ContainsPoint(s2.PointFromLatLng(ll)) {
		t.Errorf("Expected %s to be outside the polygon", ll)
	}
	if !polygon.Loop(1).IsHole() {
		t.Errorf("Expected loop 1 to be a hole")
	}
}

func ValidateTagsAreSearchable(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		{
			ID:       2300722786,
			Location: osm.LatLng{Lat: 51.5357237, Lng: -0.1253052},
			Tags: []osm.Tag{
				{Key: "name", Value: "Caravan"},
				{Key: "amenity", Value: "Restaurant"},
			},
		},
	}

	w, err := buildWorld(nodes, []osm.Way{}, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.5357237, -0.1253052)), b6.MetersToAngle(100))
	points := w.FindFeatures(b6.Typed{b6.FeatureTypePoint, b6.NewIntersectsCap(cap)})
	pointsLen := 0

	for points.Next() {
		pointsLen++

		if amenity := points.Feature().Get("#amenity"); !amenity.IsValid() {
			t.Errorf("Expected amenity tag to be searchable")
		}
		if amenity := points.Feature().Get("amenity"); amenity.IsValid() {
			t.Errorf("Expected amenity tag to be searchable")
		}
		if name := points.Feature().Get("name"); !name.IsValid() {
			t.Errorf("Didn't expect name tag to be searchable")
		}
		if name := points.Feature().Get("#name"); name.IsValid() {
			t.Errorf("Didn't expect name tag to be searchable")
		}
	}

	if pointsLen != 1 {
		t.Errorf("Expected to find 1 point, found %d", pointsLen)
	}
}

func ValidateMultipolygonsWithLoopsThatSharePoints(buildWorld BuildOSMWorld, t *testing.T) {
	relations := []osm.Relation{
		{
			ID: 10959745,
			Tags: []osm.Tag{
				{Key: "type", Value: "multipolygon"},
				{Key: "building", Value: "apartments"},
			},
			Members: []osm.Member{
				{ID: 788077724, Type: osm.ElementTypeWay, Role: "outer"},
				{ID: 788077709, Type: osm.ElementTypeWay, Role: "outer"},
				{ID: 788077705, Type: osm.ElementTypeWay, Role: "outer"},
			},
		},
	}

	// Node 7367847732 is common to all ways
	ways := []osm.Way{
		{ID: 788077724, Nodes: []osm.NodeID{7367847765, 7367847732, 7367847763, 7367847765}},
		{ID: 788077709, Nodes: []osm.NodeID{7367847759, 7367847732, 7367847765, 7367847759}},
		{ID: 788077705, Nodes: []osm.NodeID{7367847763, 7367847732, 7367847759, 7367847763}},
	}

	nodes := []osm.Node{
		{ID: 7367847765, Location: osm.LatLng{Lat: 41.3653514, Lng: 2.1141034}},
		{ID: 7367847732, Location: osm.LatLng{Lat: 41.3654027, Lng: 2.1140818}},
		{ID: 7367847763, Location: osm.LatLng{Lat: 41.3654481, Lng: 2.1140626}},
		{ID: 7367847759, Location: osm.LatLng{Lat: 41.3653772, Lng: 2.1139749}},
	}

	w, err := buildWorld(nodes, ways, relations, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	as := w.FindAreasByPoint(FromOSMNodeID(7367847732))
	seen := make(map[b6.FeatureID]struct{})
	for as.Next() {
		if _, ok := seen[as.Feature().FeatureID()]; ok {
			t.Errorf("Duplicate area %s returned from FindAreasByPoint", as.Feature().FeatureID())
		}
		seen[as.Feature().FeatureID()] = struct{}{}
	}
	expected := 4
	if len(seen) != expected {
		t.Errorf("Expected %d areas, found %d", expected, len(seen))
	}
}

func ValidateThinBuilding(buildWorld BuildOSMWorld, t *testing.T) {
	// This (likely incorrectly modelled) part of a building in Barcelona
	// is so thin that the order of it's nodes flips when they're naievely
	// snapped to an S2 cell grid. Check that this case is handled without
	// error - currently be silently dropping the feature during flattening.
	nodes := []osm.Node{
		// Way 940281419
		{ID: 7080511338, Location: osm.LatLng{Lat: 41.4389541, Lng: 2.2160516}},
		{ID: 7080511178, Location: osm.LatLng{Lat: 41.4389176, Lng: 2.2160032}},
		{ID: 7080484861, Location: osm.LatLng{Lat: 41.4390138, Lng: 2.2161307}},

		// Way 363809733
		{ID: 3679380362, Location: osm.LatLng{Lat: 41.4389069, Lng: 2.2160086}},
		{ID: 3679380364, Location: osm.LatLng{Lat: 41.4389149, Lng: 2.21601936}},
		{ID: 3679380370, Location: osm.LatLng{Lat: 41.4391610, Lng: 2.2163477}},
	}

	ways := []osm.Way{
		{
			ID:    940281419,
			Nodes: []osm.NodeID{7080511338, 7080511178, 7080484861, 7080511338},
			Tags:  []osm.Tag{{Key: "#building", Value: "yes"}},
		},
		{
			ID:    363809733,
			Nodes: []osm.NodeID{3679380362, 3679380364, 3679380370},
			Tags:  []osm.Tag{{Key: "#highway", Value: "footway"}, {Key: "footway", Value: "sidewalk"}},
		},
	}

	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	if a := b6.FindAreaByID(AreaIDFromOSMWayID(940281419), w); a != nil {
		copy := NewAreaFeatureFromWorld(a)
		if err := ValidateArea(copy, w); err != nil {
			t.Errorf("Expected no error, found: %s", err)
		}
	}

	// Ensure the invalid features are correctly filtered from all indices,
	// and that attempting to inspect the feature doesn't crash.
	fs := w.FindFeatures(b6.All{})
	for fs.Next() {
		fs.Feature().FeatureID()
	}

	ps := w.FindReferences(FromOSMNodeID(7080511178), b6.FeatureTypePath)
	for ps.Next() {
		ps.Feature().FeatureID()
	}

	as := w.FindAreasByPoint(FromOSMNodeID(7080511178))
	for as.Next() {
		as.Feature().FeatureID()
	}
}

func ValidateBrokenOSMWayForArea(buildWorld BuildOSMWorld, t *testing.T) {
	// While we delete invalid ways before adding them to the world,
	// we still attempt to calculate their area to decide whether we
	// need to inverse them - without guards, this caused a panic
	// from S2.
	nodes := []osm.Node{
		osm.Node{ID: 2309943870, Location: osm.LatLng{Lat: 51.5371371, Lng: -0.1240464}},
	}

	ways := []osm.Way{
		osm.Way{
			ID:    222021570,
			Nodes: []osm.NodeID{2309943870},
			Tags:  []osm.Tag{{Key: "area", Value: "yes"}},
		},
	}

	w, err := buildWorld(nodes, ways, []osm.Relation{}, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	if a := b6.FindAreaByID(AreaIDFromOSMWayID(222021570), w); a != nil {
		t.Errorf("Didn't expect to find a feature")
	}
}

func ValidateSearchableRelationsNotReferencedByAFeature(buildWorld BuildOSMWorld, t *testing.T) {
	nodes := []osm.Node{
		{ID: 7080511338, Location: osm.LatLng{Lat: 41.4389541, Lng: 2.2160516}},
		{ID: 7080511178, Location: osm.LatLng{Lat: 41.4389176, Lng: 2.2160032}},
	}

	relations := []osm.Relation{
		{ID: 11139964,
			Members: []osm.Member{
				{Type: osm.ElementTypeWay, ID: 807925586}, // Missing
				{Type: osm.ElementTypeWay, ID: 558345068}, // Missing
			},
			Tags: []osm.Tag{
				{Key: "#type", Value: "route"},
				{Key: "route", Value: "bicycle"},
				{Key: "network", Value: "lcn"},
			},
		},
	}

	_, err := buildWorld(nodes, []osm.Way{}, relations, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}
}

func ValidateEachFeature(buildWorld BuildOSMWorld, t *testing.T) {
	nodes, ways, relations, err := osm.ReadWholePBF(test.Data(test.GranarySquarePBF))
	if err != nil {
		t.Errorf("Failed to read world: %s", err)
		return
	}
	w, err := buildWorld(nodes, ways, relations, &BuildOptions{Cores: 2})
	if err != nil {
		t.Errorf("Failed to build world: %s", err)
		return
	}

	areas := uint64(0)
	each := func(f b6.Feature, goroutine int) error {
		atomic.AddUint64(&areas, 1)
		return nil
	}

	options := b6.EachFeatureOptions{
		SkipPoints:    true,
		SkipPaths:     true,
		SkipAreas:     false,
		SkipRelations: true,
		Goroutines:    2,
	}
	if err := w.EachFeature(each, &options); err != nil {
		t.Errorf("Expected no error from EachFeature, found: %s", err)
	}

	expectedAreas := uint64(54)
	if areas != expectedAreas {
		t.Errorf("Expected %d areas, found %d", expectedAreas, areas)
	}
}
