package ingest

import (
	"fmt"
	"log"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/osm"
	"diagonal.works/b6/test"
	"github.com/golang/geo/s2"
	"github.com/google/go-cmp/cmp"
)

type mutableWorldCreator func() (MutableWorld, error)

var mutableWorldCreators = []struct {
	name   string
	create mutableWorldCreator
}{
	{"BasicMutable", func() (MutableWorld, error) {
		return NewBasicMutableWorld(), nil
	}},
	{"MutableOverlay", func() (MutableWorld, error) {
		basic, err := buildBasicWorld([]osm.Node{}, []osm.Way{}, []osm.Relation{})
		if err != nil {
			return nil, err
		}
		return NewMutableOverlayWorld(basic), nil
	}},
}

func osmPoint(id osm.NodeID, lat float64, lng float64) *PointFeature {
	return &PointFeature{
		PointID:  FromOSMNodeID(id),
		Location: s2.LatLngFromDegrees(lat, lng),
	}
}

func osmPath(id osm.WayID, points []*PointFeature) *PathFeature {
	path := NewPathFeature(len(points))
	path.PathID = FromOSMWayID(id)
	for i, point := range points {
		path.SetPointID(i, point.PointID)
	}
	return path
}

func osmSimpleArea(id osm.WayID) *AreaFeature {
	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(id)
	area.SetPathIDs(0, []b6.PathID{FromOSMWayID(id)})
	return area
}

// osmSimpleRelation returns a relation with one Path member
func osmSimpleRelation(id osm.RelationID, member osm.WayID) *RelationFeature {
	relation := NewRelationFeature(1)
	relation.RelationID = FromOSMRelationID(id)
	relation.Members = []b6.RelationMember{
		{ID: FromOSMWayID(member).FeatureID()},
	}
	return relation
}

func addFeatures(w MutableWorld, features ...Feature) error {
	for _, f := range features {
		switch f := f.(type) {
		case *PointFeature:
			if err := w.AddPoint(f); err != nil {
				return err
			}
		case *PathFeature:
			if err := w.AddPath(f); err != nil {
				return err
			}
		case *AreaFeature:
			if err := w.AddArea(f); err != nil {
				return err
			}
		case *RelationFeature:
			if err := w.AddRelation(f); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Can't add feature type %T", f)
		}
	}
	return nil
}

func TestMutableWorlds(t *testing.T) {
	tests := []struct {
		name string
		f    func(MutableWorld, *testing.T)
	}{
		{"ReplaceFeatureWithAdditionalTag", ValidateReplaceFeatureWithAdditionalTag},
		{"UpdatePathConnectivity", ValidateUpdatePathConnectivity},
		{"UpdateAreasByPointWhenChangingPointsOnAPath", ValidateUpdateAreasByPointWhenChangingPointsOnAPath},
		{"UpdateAreasByPointWhenChangingPathsForAnArea", ValidateUpdateAreasByPointWhenChangingPathsForAnArea},
		{"UpdateAreasSharingAPoint", ValidateUpdateAreasSharingAPoint},
		{"UpdateAreasSharingAPointOnAPath", ValidateUpdateAreasSharingAPointOnAPath},
		{"UpdateRelationsByFeatureWhenChangingRelations", ValidateUpdateRelationsByFeatureWhenChangingRelations},
		{"UpdatingPathUpdatesS2CellIndex", ValidateUpdatingPathUpdatesS2CellIndex},
		{"UpdatingPointLocationsUpdatesS2CellIndex", ValidateUpdatingPointLocationsUpdatesS2CellIndex},
		{"UpdatingPointLocationsWillFailIfAreasAreInvalidated", ValidateUpdatingPointLocationsWillFailIfAreasAreInvalidated},
		{"UpdatingPathWillFailIfAreasAreInvalidated", ValidateUpdatingPathWillFailIfAreasAreInvalidated},
		{"RepeatedModification", ValidateRepeatedModification},
		{"AddingFeaturesWithNoIDFails", ValidateAddingFeaturesWithNoIDFails},
		{"AddTagToExistingFeature", ValidateAddTagToExistingFeature},
		{"AddSearchableTagToExistingFeature", ValidateAddSearchableTagToExistingFeature},
		{"AddTagToNonExistingFeature", ValidateAddTagToNonExistingFeature},
	}

	for _, creator := range mutableWorldCreators {
		for _, test := range tests {
			t.Run(fmt.Sprintf("%s/%s", creator.name, test.name), func(t *testing.T) {
				w, err := creator.create()
				if err != nil {
					t.Fatalf("Failed to build world: %v", err)
				}
				test.f(w, t)
			})
		}
	}
}

func ValidateReplaceFeatureWithAdditionalTag(w MutableWorld, t *testing.T) {
	start := osmPoint(5384190463, 51.5358664, -0.1272493)
	end := osmPoint(5384190494, 51.5362126, -0.1270125)
	path := osmPath(558345071, []*PointFeature{start, end})
	path.Tags = []b6.Tag{{Key: "#highway", Value: "footway"}}

	if err := addFeatures(w, start, end, path); err != nil {
		t.Fatal(err)
	}

	paths := b6.AllPaths(b6.FindPaths(b6.Tagged{Key: "#highway", Value: "footway"}, w))
	if len(paths) != 1 || paths[0].PathID() != path.PathID {
		t.Error("Expected to find one path")
	}

	paths = b6.AllPaths(b6.FindPaths(b6.Tagged{Key: "#bridge", Value: "yes"}, w))
	if len(paths) != 0 {
		t.Error("Didn't expect to find any bridges")
	}

	path.Tags = append(path.Tags, b6.Tag{Key: "#bridge", Value: "yes"})
	w.AddPath(path)
	paths = b6.AllPaths(b6.FindPaths(b6.Intersection{b6.Tagged{Key: "#highway", Value: "footway"}, b6.Tagged{Key: "#bridge", Value: "yes"}}, w))
	if len(paths) != 1 || paths[0].PathID() != path.PathID {
		t.Errorf("Expected to find one path, found %d", len(paths))
	}
}

func ValidateUpdatePathConnectivity(w MutableWorld, t *testing.T) {
	a := osmPoint(5384190463, 51.5358664, -0.1272493)
	b := osmPoint(5384190494, 51.5362126, -0.1270125)
	c := osmPoint(5384190476, 51.5367563, -0.1266297)

	ab := osmPath(558345071, []*PointFeature{a, b})
	ca := osmPath(558345054, []*PointFeature{c, a})

	if err := addFeatures(w, a, b, c, ab, ca); err != nil {
		t.Fatal(err)
	}

	segments := b6.AllSegments(w.Traverse(c.PointID))
	if len(segments) != 1 || segments[0].LastFeature().PointID() != a.PointID {
		t.Error("Expected to find a connection to point a")
	}

	// Swap pathC from c -> a to c -> b
	path := NewPathFeatureFromWorld(b6.FindPathByID(ca.PathID, w))
	path.SetPointID(1, b.PointID)
	if w.AddPath(path) != nil {
		t.Error("Failed to swap path c -> b")
	}

	segments = b6.AllSegments(w.Traverse(c.PointID))
	if len(segments) != 1 || segments[0].LastFeature().PointID() != b.PointID {
		t.Errorf("Expected to find a connection to point b, found none (%d segments)", len(segments))
	}
}

func ValidateUpdateAreasByPointWhenChangingPointsOnAPath(w MutableWorld, t *testing.T) {
	a := osmPoint(2309943870, 51.5371371, -0.1240464) // Top left
	b := osmPoint(2309943835, 51.5355393, -0.1247150)
	c := osmPoint(2309943825, 51.5354848, -0.1243698)

	d := osmPoint(2309943868, 51.5370710, -0.1240744)

	path := osmPath(222021570, []*PointFeature{a, b, c, a})

	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(222021570)
	area.SetPathIDs(0, []b6.PathID{path.PathID})

	if err := addFeatures(w, a, b, c, d, path, area); err != nil {
		t.Fatal(err)
	}

	areas := b6.AllAreas(w.FindAreasByPoint(c.PointID))
	if len(areas) != 1 || areas[0].AreaID() != area.AreaID {
		t.Error("Expected to find an area for point c")
	}

	areas = b6.AllAreas(w.FindAreasByPoint(d.PointID))
	if len(areas) != 0 {
		t.Error("Didn't expect to find an area for point d")
	}

	// Replace a point on the path
	modifiedPath := NewPathFeatureFromWorld(b6.FindPathByID(path.PathID, w))
	modifiedPath.SetPointID(1, d.PointID)
	if err := addFeatures(w, modifiedPath); err != nil {
		t.Fatal(err)
	}

	areas = b6.AllAreas(w.FindAreasByPoint(d.PointID))
	if len(areas) != 1 || areas[0].AreaID() != area.AreaID {
		t.Errorf("Expected to find an area for point d, found: %d", len(areas))
	}

	areas = b6.AllAreas(w.FindAreasByPoint(b.PointID))
	if len(areas) != 0 {
		t.Errorf("Expected to find no areas for point a, found: %d", len(areas))
	}
}

func ValidateUpdateAreasSharingAPointOnAPath(w MutableWorld, t *testing.T) {
	// Setup 3 areas that share a point. We specifically chose 3 to trigger
	// a bug in handling dependencies
	a := osmPoint(2309943870, 51.5371371, -0.1240464) // Top left of Eastern Transit Shed
	b := osmPoint(2309943835, 51.5355393, -0.1247150)
	c := osmPoint(2309943825, 51.5354848, -0.1243698)

	d := osmPoint(2309943853, 51.5359057, -0.1253953) // The Granary
	e := osmPoint(2309943849, 51.5357793, -0.1246146)

	f := osmPoint(2309943871, 51.5371983, -0.1248445) // UAL
	g := osmPoint(2309943853, 51.5359057, -0.1253953)

	h := osmPoint(2309943869, 51.5370778, -0.1236840) // Used to modify

	pathA := osmPath(222021570, []*PointFeature{a, b, c, a})
	areaA := NewAreaFeature(1)
	areaA.AreaID = AreaIDFromOSMWayID(222021570)
	areaA.SetPathIDs(0, []b6.PathID{pathA.PathID})

	pathB := osmPath(222021576, []*PointFeature{a, d, e, a})
	areaB := NewAreaFeature(1)
	areaB.AreaID = AreaIDFromOSMWayID(222021576)
	areaB.SetPathIDs(0, []b6.PathID{pathB.PathID})

	pathC := osmPath(222021578, []*PointFeature{a, f, g, a})
	areaC := NewAreaFeature(1)
	areaC.AreaID = AreaIDFromOSMWayID(222021578)
	areaC.SetPathIDs(0, []b6.PathID{pathC.PathID})

	if err := addFeatures(w, a, b, c, d, e, f, g, h, pathA, pathB, pathC, areaA, areaB, areaC); err != nil {
		t.Fatal(err)
	}

	// Replace points on pathA
	modifiedPath := NewPathFeatureFromWorld(b6.FindPathByID(pathA.PathID, w))
	modifiedPath.SetPointID(0, h.PointID)
	modifiedPath.SetPointID(3, h.PointID)
	if err := addFeatures(w, modifiedPath); err != nil {
		t.Fatal(err)
	}

	areas := b6.AllAreas(w.FindAreasByPoint(h.PointID))
	if len(areas) != 1 || areas[0].AreaID() != areaA.AreaID {
		t.Errorf("Expected to find an area for point h, found: %d", len(areas))
	}

	areas = b6.AllAreas(w.FindAreasByPoint(a.PointID))
	if len(areas) != 2 || areas[0].AreaID() == areaA.AreaID || areas[1].AreaID() == areaA.AreaID {
		t.Errorf("Expected to find 2 areas for point a, none areaA, found: %d", len(areas))
	}
}

func ValidateUpdateAreasSharingAPoint(w MutableWorld, t *testing.T) {
	a := osmPoint(2309943870, 51.5371371, -0.1240464) // Top left of Eastern Transit Shed
	b := osmPoint(2309943835, 51.5355393, -0.1247150) // Shared with other way
	c := osmPoint(2309943825, 51.5354848, -0.1243698)

	d := osmPoint(2309943849, 51.5357793, -0.1246146) // The Granary
	e := osmPoint(2309943853, 51.5359057, -0.1253953)

	pathA := osmPath(222021570, []*PointFeature{a, b, c, a})
	areaA := NewAreaFeature(1)
	areaA.AreaID = AreaIDFromOSMWayID(222021570)
	areaA.SetPathIDs(0, []b6.PathID{pathA.PathID})

	pathB := osmPath(222021576, []*PointFeature{b, d, e, b})
	areaB := NewAreaFeature(1)
	areaB.AreaID = AreaIDFromOSMWayID(222021576)
	areaB.SetPathIDs(0, []b6.PathID{pathB.PathID})

	if err := addFeatures(w, a, b, c, d, e, pathA, pathB, areaA, areaB); err != nil {
		t.Fatal(err)
	}

	// Move the shared point
	modifiedPoint := osmPoint(2309943835, 51.5355185, -0.1254976)
	if err := addFeatures(w, modifiedPoint); err != nil {
		t.Fatal(err)
	}

	point := b6.FindPointByID(modifiedPoint.PointID, w)
	if point == nil {
		t.Fatal("Expected to find point")
	}

	if point.Point().Distance(modifiedPoint.Point()) > 0.000001 {
		t.Error("Expected modified point to have an updated location")
	}
}

func ValidateUpdateAreasByPointWhenChangingPathsForAnArea(w MutableWorld, t *testing.T) {
	a := osmPoint(2309943835, 51.5355393, -0.1247150)
	b := osmPoint(2309943825, 51.5354848, -0.1243698)
	c := osmPoint(2309943870, 51.5371371, -0.1240464)
	d := osmPoint(598093309, 51.5321649, -0.1269834)

	path := osmPath(222021570, []*PointFeature{a, b, c, a})

	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(222021570)
	area.SetPathIDs(0, []b6.PathID{path.PathID})

	if err := addFeatures(w, a, b, c, d, path, area); err != nil {
		t.Fatal(err)
	}

	newPath := osmPath(222021577, []*PointFeature{b, c, d, b})
	modifiedArea := NewAreaFeatureFromWorld(b6.FindAreaByID(area.AreaID, w))
	modifiedArea.SetPathIDs(0, []b6.PathID{newPath.PathID})

	if err := addFeatures(w, newPath); err != nil {
		t.Fatal(err)
	}

	if err := w.AddArea(modifiedArea); err != nil {
		t.Errorf("Failed to add modified area: %s", err)
	}

	areas := b6.AllAreas(w.FindAreasByPoint(d.PointID))
	if len(areas) != 1 || areas[0].AreaID() != modifiedArea.AreaID {
		t.Error("Expected to find an area for point d")
	}

	areas = b6.AllAreas(w.FindAreasByPoint(a.PointID))
	if len(areas) != 0 {
		t.Error("Didn't expect to find an area for point a")
	}
}

func ValidateUpdateRelationsByFeatureWhenChangingRelations(w MutableWorld, t *testing.T) {
	a := osmPoint(5378333625, 51.5352195, -0.1254286)
	b := osmPoint(5384190491, 51.5352339, -0.1255240)
	c := osmPoint(4966136655, 51.5349570, -0.1256696)

	ab := NewPathFeature(2)
	ab.PathID = FromOSMWayID(807925586)
	ab.SetPointID(0, a.PointID)
	ab.SetPointID(1, b.PointID)

	bc := NewPathFeature(2)
	bc.PathID = FromOSMWayID(558345068)
	bc.SetPointID(0, b.PointID)
	bc.SetPointID(1, c.PointID)

	relation := NewRelationFeature(1)
	relation.RelationID = FromOSMRelationID(11139964)
	relation.Members[0] = b6.RelationMember{ID: ab.FeatureID()}
	relation.Tags = []b6.Tag{{Key: "type", Value: "route"}}

	if err := addFeatures(w, a, b, c, ab, bc, relation); err != nil {
		t.Fatal(err)
	}

	relations := b6.AllRelations(w.FindRelationsByFeature(ab.FeatureID()))
	if len(relations) != 1 || relations[0].RelationID() != relation.RelationID {
		t.Error("Expected to find a relation for path AB")
	}

	relations = b6.AllRelations(w.FindRelationsByFeature(bc.FeatureID()))
	if len(relations) != 0 {
		t.Error("Didn't expect any relations for path BC")
	}

	modifiedRelation := NewRelationFeatureFromWorld(b6.FindRelationByID(relation.RelationID, w))
	modifiedRelation.Members[0] = b6.RelationMember{ID: bc.FeatureID()}
	if err := w.AddRelation(modifiedRelation); err != nil {
		t.Errorf("Failed to add relation: %s", err)
	}

	relations = b6.AllRelations(w.FindRelationsByFeature(bc.FeatureID()))
	if len(relations) != 1 || relations[0].RelationID() != relation.RelationID {
		t.Error("Expected to find a relation for path BC")
	}

	relations = b6.AllRelations(w.FindRelationsByFeature(ab.FeatureID()))
	if len(relations) != 0 {
		t.Error("Didn't expect any relations for path AB")
	}
}

func ValidateUpdatingPathUpdatesS2CellIndex(w MutableWorld, t *testing.T) {
	// Extend the Western Transit Shed in Granary Square to cover the Eastern
	// Handyside Canopy, and ensure we can find the resulting area

	// Western Shed points, counterclockwise
	a := osmPoint(2309943873, 51.5373249, -0.1251784) // Top-left
	b := osmPoint(2309943847, 51.5357239, -0.1258568)
	c := osmPoint(2309943846, 51.5356657, -0.1254957)
	d := osmPoint(2309943872, 51.5372656, -0.1248160)

	// Eastern Handyside Canopy, East edge
	e := osmPoint(2309943852, 51.5358965, -0.1230551)
	f := osmPoint(2309943867, 51.5370349, -0.1232719) // Top right

	path := osmPath(222021577, []*PointFeature{a, b, c, d, a})
	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(222021577)
	area.SetPathIDs(0, []b6.PathID{path.PathID})

	if err := addFeatures(w, a, b, c, d, path, area); err != nil {
		t.Fatal(err)
	}

	// A cap covering only part of the Eastern Shed
	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.5370349, -0.1232719)), b6.MetersToAngle(10))
	areas := b6.AllAreas(b6.FindAreas(b6.MightIntersect{cap}, w))
	if len(areas) != 0 {
		t.Errorf("Didn't expect to find an area %s", areas[0].FeatureID())
	}

	modifiedPath := NewPathFeatureFromWorld(b6.FindPathByID(path.PathID, w))
	modifiedPath.SetPointID(2, e.PointID)
	modifiedPath.SetPointID(3, f.PointID)
	if err := addFeatures(w, e, f, modifiedPath); err != nil {
		t.Fatal(err)
	}

	areas = b6.AllAreas(b6.FindAreas(b6.NewIntersectsCap(cap), w))
	if len(areas) != 1 || areas[0].AreaID() != area.AreaID {
		t.Errorf("Expected to find 1 matching area within the region (found %d)", len(areas))
	}
}

func ValidateUpdatingPointLocationsUpdatesS2CellIndex(w MutableWorld, t *testing.T) {
	// Extend the Western Transit Shed in Granary Square to cover the Eastern
	// Handyside Canopy, and ensure we can find the resulting area

	// Western Shed points, counterclockwise
	a := osmPoint(2309943873, 51.5373249, -0.1251784) // Top-left
	b := osmPoint(2309943847, 51.5357239, -0.1258568)
	c := osmPoint(2309943846, 51.5356657, -0.1254957)
	d := osmPoint(2309943872, 51.5372656, -0.1248160)

	// Eastern Handyside Canopy, East edge
	e := osmPoint(2309943852, 51.5358965, -0.1230551)
	f := osmPoint(2309943867, 51.5370349, -0.1232719) // Top right

	path := osmPath(222021577, []*PointFeature{a, b, c, d, a})
	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(222021577)
	area.SetPathIDs(0, []b6.PathID{path.PathID})

	if err := addFeatures(w, a, b, c, d, path, area); err != nil {
		t.Fatal(err)
	}

	// A cap covering only part of the Eastern Shed
	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.5370349, -0.1232719)), b6.MetersToAngle(10))
	areas := b6.AllAreas(b6.FindAreas(b6.NewIntersectsCap(cap), w))
	if len(areas) != 0 {
		t.Errorf("Didn't expect to find an area %s", areas[0].FeatureID())
	}

	newC := NewPointFeatureFromWorld(b6.FindPointByID(c.PointID, w))
	newC.Location = e.Location
	newD := NewPointFeatureFromWorld(b6.FindPointByID(d.PointID, w))
	newD.Location = f.Location
	if err := addFeatures(w, newC, newD); err != nil {
		t.Fatal(err)
	}

	areas = b6.AllAreas(b6.FindAreas(b6.NewIntersectsCap(cap), w))
	if len(areas) != 1 || areas[0].AreaID() != area.AreaID {
		t.Error("Expected to find an area within the region")
	}
}

func ValidateUpdatingPointLocationsWillFailIfAreasAreInvalidated(w MutableWorld, t *testing.T) {
	// Move a point of the Western transit shed way out West, such that the
	// polygon self-intersects.

	// Western Shed points, counterclockwise
	a := osmPoint(2309943873, 51.5373249, -0.1251784) // Top-left
	b := osmPoint(2309943847, 51.5357239, -0.1258568)
	c := osmPoint(2309943846, 51.5356657, -0.1254957)
	d := osmPoint(2309943872, 51.5372656, -0.1248160)

	path := osmPath(222021577, []*PointFeature{a, b, c, d, a})
	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(222021577)
	area.SetPathIDs(0, []b6.PathID{path.PathID})

	if err := addFeatures(w, a, b, c, d, path, area); err != nil {
		t.Fatal(err)
	}

	modifiedPoint := NewPointFeatureFromWorld(b6.FindPointByID(c.PointID, w))
	modifiedPoint.Location = s2.LatLngFromDegrees(51.5368549, -0.1256275)
	if w.AddPoint(modifiedPoint) == nil {
		t.Error("Expected adding point to fail, as it invalidates an area")
	}
}

func ValidateUpdatingPathWillFailIfAreasAreInvalidated(w MutableWorld, t *testing.T) {
	// Replace a path used by an area with one which self-intersects.

	// Western Shed points, counterclockwise
	a := osmPoint(2309943873, 51.5373249, -0.1251784) // Top-left
	b := osmPoint(2309943847, 51.5357239, -0.1258568)
	c := osmPoint(2309943846, 51.5356657, -0.1254957)
	d := osmPoint(2309943872, 51.5372656, -0.1248160)

	// A fountain in Lewis Cubitt Square
	e := osmPoint(4031177264, 51.5368549, -0.1256275)

	path := osmPath(222021577, []*PointFeature{a, b, c, d, a})
	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(222021577)
	area.SetPathIDs(0, []b6.PathID{path.PathID})

	if err := addFeatures(w, a, b, c, d, e, path, area); err != nil {
		t.Fatal(err)
	}

	modifiedPath := NewPathFeatureFromWorld(b6.FindPathByID(path.PathID, w))
	modifiedPath.SetPointID(2, e.PointID)
	if w.AddPath(modifiedPath) == nil {
		t.Error("Expected adding path to fail, as it invalidates an area")
	}
}

func ValidateAddingFeaturesWithNoIDFails(w MutableWorld, t *testing.T) {
	if err := w.AddPoint(&PointFeature{}); err == nil {
		t.Error("Expected adding a point with no ID to fail")
	}
	if err := w.AddPath(&PathFeature{}); err == nil {
		t.Error("Expected adding a path with no ID to fail")
	}
	if err := w.AddArea(&AreaFeature{}); err == nil {
		t.Error("Expected adding an area with no ID to fail")
	}
	if err := w.AddRelation(&RelationFeature{}); err == nil {
		t.Error("Expected adding a realtion with no ID to fail")
	}
}

func ValidateAddTagToExistingFeature(w MutableWorld, t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})
	if err := addFeatures(w, caravan); err != nil {
		t.Fatal(err)
	}

	if err := w.AddTag(caravan.FeatureID(), b6.Tag{Key: "amenity", Value: "restaurant"}); err != nil {
		t.Fatalf("Failed to add tag: %s", err)
	}

	found := b6.FindPointByID(caravan.PointID, w)
	if found == nil {
		t.Fatal("Failed to find feature")
	}
	if found.Get("amenity").Value != "restaurant" {
		t.Error("Failed to find expected tag value")
	}
}

func ValidateRepeatedModification(w MutableWorld, t *testing.T) {
	id := osm.WayID(222021577)
	finders := []struct {
		name string
		f    func(w b6.World) b6.PathFeature
	}{
		{"ByID", func(w b6.World) b6.PathFeature {
			return b6.FindPathByID(FromOSMWayID(id), w)
		}},
		{"BySearch", func(w b6.World) b6.PathFeature {
			highways := b6.AllPaths(b6.FindPaths(b6.Keyed{"#highway"}, w))
			if len(highways) != 1 {
				t.Fatalf("Expected to find 1 path, found %d", len(highways))
			}
			return highways[0]
		}},
	}

	tests := []struct {
		name string
		f    func(id osm.WayID, find func(b6.World) b6.PathFeature, w MutableWorld, t *testing.T)
	}{
		{"ChangingNodes", ValidateRepeatedModificationChangingNodes},
		{"AddingNodes", ValidateRepeatedModificationAddingNodes},
	}

	for _, find := range finders {
		for _, test := range tests {
			t.Run(fmt.Sprintf("%s/%s", find.name, test.name), func(t *testing.T) { test.f(id, find.f, w, t) })
		}
	}
}

func ValidateRepeatedModificationChangingNodes(id osm.WayID, find func(b6.World) b6.PathFeature, w MutableWorld, t *testing.T) {
	// Start with area abcda, repeatedly modify it to increment the ID
	// of each point by 1
	a := osmPoint(2309943873, 51.5373249, -0.1251784)
	b := osmPoint(2309943847, 51.5357239, -0.1258568)
	c := osmPoint(2309943846, 51.5356657, -0.1254957)
	d := osmPoint(2309943872, 51.5372656, -0.1248160)

	path := osmPath(id, []*PointFeature{a, b, c, d, a})
	path.Tags = []b6.Tag{{Key: "#highway", Value: "primary"}}
	if err := addFeatures(w, a, b, c, d, path); err != nil {
		t.Fatal(err)
	}

	points := []*PointFeature{a, b, c, d}
	for i := 0; i < len(points); i++ {
		modifiedPoint := points[i].ClonePointFeature()
		modifiedPoint.PointID = b6.MakePointID(points[i].Namespace, points[i].Value+1)
		if err := w.AddPoint(modifiedPoint); err != nil {
			t.Fatal(err)
		}
		found := find(w)
		if found == nil {
			// TODO Do we need to fail or return?
			t.Fatal("Failed to find feature")
		}
		modifiedPath := NewPathFeatureFromWorld(found)
		modifiedPath.SetPointID(i, modifiedPoint.PointID)
		if err := w.AddPath(modifiedPath); err != nil {
			t.Fatal(err)
		}
	}

	found := b6.FindPathByID(path.PathID, w)
	for i := 0; i < len(points); i++ {
		if found.Feature(i).FeatureID().Value != points[i].FeatureID().Value+1 {
			t.Errorf("Expected to find ID %d for point %d, found %d", points[i].FeatureID().Value+1, i, found.Feature(i).FeatureID().Value)
		}
	}
}

func ValidateRepeatedModificationAddingNodes(id osm.WayID, find func(b6.World) b6.PathFeature, w MutableWorld, t *testing.T) {
	// Start with path ab, repeatedly add points to it to turn it into aedcb
	a := osmPoint(2309943873, 51.5373249, -0.1251784)
	b := osmPoint(2309943847, 51.5357239, -0.1258568)
	c := osmPoint(2309943846, 51.5356657, -0.1254957)
	d := osmPoint(2309943872, 51.5372656, -0.1248160)
	e := osmPoint(4031177264, 51.5368549, -0.1256275)

	path := osmPath(id, []*PointFeature{a, b})
	path.Tags = []b6.Tag{{Key: "#highway", Value: "primary"}}
	if err := addFeatures(w, a, b, c, d, e, path); err != nil {
		t.Fatal(err)
	}

	newPoints := []*PointFeature{c, d, e}
	for i := 0; i < len(newPoints); i++ {
		found := find(w)
		if found == nil {
			// TODO Do we need to fail or return?
			t.Fatal("Failed to find feature")
		}
		newPath := NewPathFeature(found.Len() + 1)
		newPath.PathID = found.PathID()
		newPath.Tags = NewTagsFromWorld(found)
		insert := 0
		for j := 0; j < found.Len(); j++ {
			if j == 1 {
				newPath.SetPointID(insert, newPoints[i].PointID)
				insert++
			}
			newPath.SetPointID(insert, found.Feature(j).PointID())
			insert++
		}
		if err := w.AddPath(newPath); err != nil {
			t.Fatal(err)
		}
	}

	found := b6.FindPathByID(path.PathID, w)
	var gotPointIDs []b6.PointID
	for i := 0; i < found.Len(); i++ {
		gotPointIDs = append(gotPointIDs, found.Feature(i).PointID())
	}
	expected := []b6.PointID{a.PointID, e.PointID, d.PointID, c.PointID, b.PointID}
	if diff := cmp.Diff(expected, gotPointIDs); diff != "" {
		t.Errorf("Unexpected point IDs (-want +got):\n%s", diff)
	}
}

func ValidateAddSearchableTagToExistingFeature(w MutableWorld, t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})

	lighterman := osmPoint(427900370, 51.5353986, -0.1243711)
	lighterman.AddTag(b6.Tag{Key: "name", Value: "The Lighterman"})
	lighterman.AddTag(b6.Tag{Key: "#amenity", Value: "restaurant"})

	if err := addFeatures(w, caravan, lighterman); err != nil {
		t.Fatal(err)
	}

	points := b6.AllPoints(b6.FindPoints(b6.Tagged{Key: "#amenity", Value: "restaurant"}, w))
	if len(points) != 1 {
		t.Fatalf("Expected to find 1 point, found %d", len(points))
	}

	if err := w.AddTag(caravan.FeatureID(), b6.Tag{Key: "#amenity", Value: "restaurant"}); err != nil {
		t.Fatalf("Failed to add tag: %s", err)
	}

	foundPoint := b6.FindPointByID(caravan.PointID, w)
	if foundPoint == nil {
		t.Fatal("Failed to find feature")
	}
	if foundPoint.Get("#amenity").Value != "restaurant" {
		t.Errorf("Failed to find expected tag value")
	}

	points = b6.AllPoints(b6.FindPoints(b6.Tagged{Key: "#amenity", Value: "restaurant"}, w))
	if len(points) != 2 {
		t.Fatalf("Expected to find 2 points, found %d", len(points))
	}

	found := false
	for _, point := range points {
		if point.FeatureID() == caravan.FeatureID() {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find Caravan")
	}
}

func ValidateAddTagToNonExistingFeature(w MutableWorld, t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	if err := w.AddTag(caravan.FeatureID(), b6.Tag{Key: "#amenity", Value: "restaurant"}); err == nil {
		t.Error("Expected an error, found none")
	}
}

func TestModifyPathInExistingWorld(t *testing.T) {
	// Extend the Western Transit Shed in Granary Square to cover the Eastern
	// Handyside Canopy, by switching out points in the path, and ensure we
	// can find the resulting area Western Shed points
	a := osmPoint(2309943873, 51.5373249, -0.1251784) // Top-left
	b := osmPoint(2309943847, 51.5357239, -0.1258568)
	c := osmPoint(2309943846, 51.5356657, -0.1254957)
	d := osmPoint(2309943872, 51.5372656, -0.1248160)

	// Eastern Handyside Canopy, East edge
	e := osmPoint(2309943852, 51.5358965, -0.1230551)
	f := osmPoint(2309943867, 51.5370349, -0.1232719) // Top right

	// A fountain in Lewis Cubitt Square
	g := osmPoint(4031177264, 51.5368549, -0.1256275)

	path := osmPath(222021577, []*PointFeature{a, b, c, d, a})
	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(222021577)
	area.SetPathIDs(0, []b6.PathID{path.PathID})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, c, d, g, path, area); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	if fountain := b6.FindPointByID(e.PointID, overlay); fountain != nil {
		t.Error("Expected to find fountain in base via overlay")
	}

	// A cap covering only part of the Eastern Shed
	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.5370349, -0.1232719)), b6.MetersToAngle(10))
	areas := b6.AllAreas(b6.FindAreas(b6.NewIntersectsCap(cap), overlay))
	if len(areas) != 0 {
		t.Errorf("Didn't expect to find an area %s", areas[0].FeatureID())
	}

	modifiedPath := NewPathFeatureFromWorld(b6.FindPathByID(path.PathID, overlay))
	modifiedPath.SetPointID(2, e.PointID)
	modifiedPath.SetPointID(3, f.PointID)
	if err := addFeatures(overlay, e, f, modifiedPath); err != nil {
		t.Fatal(err)
	}

	areas = b6.AllAreas(b6.FindAreas(b6.NewIntersectsCap(cap), overlay))
	if len(areas) != 1 || areas[0].AreaID() != area.AreaID {
		t.Errorf("Expected to area within the region (found %d areas)", len(areas))
	}

	paths := b6.AllPaths(overlay.FindPathsByPoint(a.PointID))
	if len(paths) != 1 || paths[0].PathID().Value != path.PathID.Value {
		t.Errorf("Expected to find 1 path by point a, found %d", len(paths))
	}
}

func TestModifyPointsOnPathInExistingWorld(t *testing.T) {
	// Move the Stable Street bridge to somewhere around Bank, by relocating
	// its points, to ensure indices are properly updated.
	a := osmPoint(1447052073, 51.5350350, -0.1247934)
	b := osmPoint(1540349979, 51.5348204, -0.1246405)

	path := osmPath(140633010, []*PointFeature{a, b})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, path); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	granarySquareCap := s2.CapFromCenterAngle(s2.Interpolate(0.5, a.Point(), b.Point()), b6.MetersToAngle(10))
	paths := b6.AllPaths(b6.FindPaths(b6.NewIntersectsCap(granarySquareCap), overlay))
	if len(paths) != 1 {
		t.Errorf("Expected to find 1 path around Granary Square, found %d", len(paths))
	}

	aPrime := osmPoint(1447052073, 51.5132689, -0.0988335)
	bPrime := osmPoint(1540349979, 51.5129188, -0.0985641)
	if err := addFeatures(base, aPrime, bPrime); err != nil {
		t.Fatal(err)
	}

	bankCap := s2.CapFromCenterAngle(s2.Interpolate(0.5, aPrime.Point(), bPrime.Point()), b6.MetersToAngle(10))
	paths = b6.AllPaths(b6.FindPaths(b6.NewIntersectsCap(bankCap), overlay))
	if len(paths) != 1 {
		t.Errorf("Expected to find 1 path around Bank, found %d", len(paths))
	}

	paths = b6.AllPaths(b6.FindPaths(b6.NewIntersectsCap(granarySquareCap), overlay))
	if len(paths) != 0 {
		t.Errorf("Expected to find no paths around Granary Square, found %d", len(paths))
	}

	paths = b6.AllPaths(overlay.FindPathsByPoint(a.PointID))
	if len(paths) != 1 {
		t.Errorf("Expected 1 path, found: %d", len(paths))
	}
}

func TestModifyPointsOnClosedPathInExistingWorld(t *testing.T) {
	// Move the path representing the Lighterman to Bank, by relocating
	// its points, to ensure indices are properly updated.
	a := osmPoint(4270651271, 51.5353986, -0.1243711)
	b := osmPoint(5693730033, 51.5352871, -0.1244193)
	c := osmPoint(4270651273, 51.5351278, -0.1243315)

	path := osmPath(140633010, []*PointFeature{a, b, c, a})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, c, path); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	granarySquareCap := s2.CapFromCenterAngle(s2.Interpolate(0.5, a.Point(), b.Point()), b6.MetersToAngle(10))
	paths := b6.AllPaths(b6.FindPaths(b6.NewIntersectsCap(granarySquareCap), overlay))
	if len(paths) != 1 {
		t.Errorf("Expected to find 1 path around Granary Square, found %d", len(paths))
	}

	aPrime := osmPoint(4270651271, 51.5137306, -0.0905139)
	bPrime := osmPoint(5693730033, 51.5134981, -0.0898162)
	cPrime := osmPoint(4270651273, 51.5138208, -0.0896115)

	if err := addFeatures(base, aPrime, bPrime, cPrime); err != nil {
		t.Fatal(err)
	}

	bankCap := s2.CapFromCenterAngle(s2.Interpolate(0.5, aPrime.Point(), bPrime.Point()), b6.MetersToAngle(10))
	paths = b6.AllPaths(b6.FindPaths(b6.NewIntersectsCap(bankCap), overlay))
	if len(paths) != 1 {
		t.Errorf("Expected to find 1 path around Bank, found %d", len(paths))
	}

	paths = b6.AllPaths(b6.FindPaths(b6.NewIntersectsCap(granarySquareCap), overlay))
	if len(paths) != 0 {
		t.Errorf("Expected to find no paths around Granary Square, found %d", len(paths))
	}

	paths = b6.AllPaths(overlay.FindPathsByPoint(a.PointID))
	if len(paths) != 1 || paths[0].PathID().Value != path.PathID.Value {
		t.Errorf("Expected 1 path, found: %d", len(paths))
		for _, p := range paths {
			log.Printf("  * %s", p.PathID())
		}
	}
}

func TestModifyPathWithIntersectionsInExistingWorld(t *testing.T) {
	a := osmPoint(6083741698, 51.5352814, -0.1266217)
	b := osmPoint(7787634237, 51.5354236, -0.1267632)
	c := osmPoint(6083735356, 51.5355776, -0.1268618)
	d := osmPoint(6083735379, 51.5361482, -0.1264835)

	e := osmPoint(7787634210, 51.5355869, -0.1269299)

	ad := osmPath(647895239, []*PointFeature{a, b, c, d})
	ec := osmPath(647895212, []*PointFeature{e, c})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, c, d, e, ad, ec); err != nil {
		t.Fatal(err)
	}
	overlay := NewMutableOverlayWorld(base)

	reachable := make(map[b6.PathID]bool)
	segments := overlay.Traverse(c.PointID)
	for segments.Next() {
		reachable[segments.Segment().Feature.PathID()] = true
	}

	if len(reachable) != 2 || !reachable[ad.PathID] || !reachable[ec.PathID] {
		t.Error("Didn't find expected initial connections")
	}

	newAD := osmPath(osm.WayID(ad.PathID.Value), []*PointFeature{a, c, d})
	if err := addFeatures(overlay, newAD); err != nil {
		t.Fatal(err)
	}

	reachable = make(map[b6.PathID]bool)
	segments = overlay.Traverse(c.PointID)
	for segments.Next() {
		reachable[segments.Segment().Feature.PathID()] = true
	}

	if len(reachable) != 2 || !reachable[ad.PathID] || !reachable[ec.PathID] {
		t.Error("Expected modification to retain existing connections")
	}
}

func validateTags(tagged b6.Taggable, expected []b6.Tag) error {
	for _, tag := range expected {
		if found := tagged.Get(tag.Key); !found.IsValid() || found.Value != tag.Value {
			return fmt.Errorf("Expected value %q for tag %q, found %q", tag.Value, tag.Key, found.Value)
		}
	}

	if allTags := tagged.AllTags(); len(allTags) != len(expected) {
		return fmt.Errorf("Expected %d tags from AllTags(), found %d", len(expected), len(allTags))
	}

	tags := make(map[string]string)
	for _, tag := range tagged.AllTags() {
		tags[tag.Key] = tag.Value
	}

	for _, tag := range expected {
		if value, ok := tags[tag.Key]; !ok || value != tag.Value {
			return fmt.Errorf("Expected value %q for tag %q, found %q", tag.Value, tag.Key, value)
		}
	}
	return nil
}

func TestChangeTagsOnExistingPoint(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})
	caravan.AddTag(b6.Tag{Key: "wheelchair", Value: "no"})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	// TODO: Updating the amenity actually needs to update the search index too.
	overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "amenity", Value: "restaurant"})
	overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "wheelchair", Value: "yes"})

	point := b6.FindPointByID(caravan.PointID, overlay)

	expected := []b6.Tag{
		{Key: "name", Value: "Caravan"},
		{Key: "wheelchair", Value: "yes"},
		{Key: "amenity", Value: "restaurant"},
	}

	if err := validateTags(point, expected); err != nil {
		t.Error(err)
	}
}

func TestChangeTagsOnExistingPath(t *testing.T) {
	a := osmPoint(7555161584, 51.5345488, -0.1251005)
	b := osmPoint(6384669830, 51.5342291, -0.1262792)

	// Goods Way, South of Granary Square
	ab := osmPath(807924986, []*PointFeature{a, b})
	ab.AddTag(b6.Tag{Key: "highway", Value: "tertiary"})
	ab.AddTag(b6.Tag{Key: "lit", Value: "no"})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, ab); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	overlay.AddTag(ab.FeatureID(), b6.Tag{Key: "cycleway:left", Value: "track"})
	overlay.AddTag(ab.FeatureID(), b6.Tag{Key: "lit", Value: "yes"})

	path := b6.FindPathByID(ab.PathID, overlay)

	expected := []b6.Tag{
		{Key: "highway", Value: "tertiary"},
		{Key: "lit", Value: "yes"},
		{Key: "cycleway:left", Value: "track"},
	}

	if err := validateTags(path, expected); err != nil {
		t.Error(err)
	}
}

func TestChangeTagsOnExistingArea(t *testing.T) {
	a := osmPoint(5693730034, 51.5352979, -0.1244842)
	b := osmPoint(5693730033, 51.5352871, -0.1244193)
	c := osmPoint(4270651271, 51.5353986, -0.1243711)
	abc := osmPath(427900370, []*PointFeature{a, b, c, a})

	lighterman := osmSimpleArea(427900370)
	lighterman.AddTag(b6.Tag{Key: "name", Value: "The Lighterman"})
	lighterman.AddTag(b6.Tag{Key: "wheelchair", Value: "no"})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, c, abc, lighterman); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	overlay.AddTag(lighterman.FeatureID(), b6.Tag{Key: "amenity", Value: "pub"})
	overlay.AddTag(lighterman.FeatureID(), b6.Tag{Key: "wheelchair", Value: "yes"})

	area := b6.FindAreaByID(lighterman.AreaID, overlay)

	expected := []b6.Tag{
		{Key: "name", Value: "The Lighterman"},
		{Key: "amenity", Value: "pub"},
		{Key: "wheelchair", Value: "yes"},
	}

	if err := validateTags(area, expected); err != nil {
		t.Error(err)
	}
}

func TestAddSearchableTagTagToExistingArea(t *testing.T) {
	a := osmPoint(5693730034, 51.5352979, -0.1244842)
	b := osmPoint(5693730033, 51.5352871, -0.1244193)
	c := osmPoint(4270651271, 51.5353986, -0.1243711)
	abc := osmPath(427900370, []*PointFeature{a, b, c, a})

	lighterman := osmSimpleArea(427900370)
	lighterman.AddTag(b6.Tag{Key: "name", Value: "The Lighterman"})
	lighterman.AddTag(b6.Tag{Key: "wheelchair", Value: "no"})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, c, abc, lighterman); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	overlay.AddTag(lighterman.FeatureID(), b6.Tag{Key: "#reachable", Value: "yes"})

	areas := b6.AllAreas(b6.FindAreas(b6.Tagged{Key: "#reachable", Value: "yes"}, overlay))
	if len(areas) != 1 {
		t.Fatalf("Expected to find 1 area, found %d", len(areas))
	}
	if areas[0].FeatureID().Value != lighterman.FeatureID().Value {
		t.Errorf("Expected to find ID %s, found %s", lighterman.FeatureID(), areas[0].FeatureID())
	}
}

func TestChangeTagsOnExistingRelation(t *testing.T) {
	a := osmPoint(7555161584, 51.5345488, -0.1251005)
	b := osmPoint(6384669830, 51.5342291, -0.1262792)
	// Goods Way, South of Granary Square
	ab := osmPath(807924986, []*PointFeature{a, b})
	// Part of C6
	c6 := osmSimpleRelation(10341051, 807924986)
	c6.AddTag(b6.Tag{Key: "type", Value: "route"})
	c6.AddTag(b6.Tag{Key: "network", Value: "lcn"})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, ab, c6); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	overlay.AddTag(c6.FeatureID(), b6.Tag{Key: "route", Value: "bicycle"})
	overlay.AddTag(c6.FeatureID(), b6.Tag{Key: "network", Value: "rcn"})

	relation := b6.FindRelationByID(c6.RelationID, overlay)

	expected := []b6.Tag{
		{Key: "type", Value: "route"},
		{Key: "route", Value: "bicycle"},
		{Key: "network", Value: "rcn"},
	}

	if err := validateTags(relation, expected); err != nil {
		t.Error(err)
	}
}

func TestReturnModifiedTagsFromSearch(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})
	caravan.AddTag(b6.Tag{Key: "#amenity", Value: "restaurant"})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "wheelchair", Value: "yes"})

	points := b6.AllPoints(b6.FindPoints(b6.Tagged{Key: "#amenity", Value: "restaurant"}, overlay))
	if len(points) != 1 {
		t.Fatal("Expected to find a point")
	}

	expected := []b6.Tag{
		{Key: "name", Value: "Caravan"},
		{Key: "#amenity", Value: "restaurant"},
		{Key: "wheelchair", Value: "yes"},
	}

	if err := validateTags(points[0], expected); err != nil {
		t.Error(err)
	}
}

func TestSettingTheSameTagMultipleTimesChangesTheValue(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})
	caravan.AddTag(b6.Tag{Key: "#amenity", Value: "restaurant"})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	for i := 0; i < 4; i++ {
		overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "100m", Value: "yes"})
		overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "#200m", Value: "yes"})
	}

	expected := []b6.Tag{
		{Key: "name", Value: "Caravan"},
		{Key: "#amenity", Value: "restaurant"},
		{Key: "100m", Value: "yes"},
		{Key: "#200m", Value: "yes"},
	}

	if err := validateTags(overlay.FindFeatureByID(caravan.FeatureID()), expected); err != nil {
		t.Error(err)
	}
}

func TestRemoveTag(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})
	caravan.AddTag(b6.Tag{Key: "#amenity", Value: "restaurant"})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	for i := 0; i < 4; i++ {
		overlay.RemoveTag(caravan.FeatureID(), "#amenity")
		overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "#shop", Value: "supermarket"})
	}

	expected := []b6.Tag{
		{Key: "name", Value: "Caravan"},
		{Key: "#shop", Value: "supermarket"},
	}

	if err := validateTags(overlay.FindFeatureByID(caravan.FeatureID()), expected); err != nil {
		t.Error(err)
	}
}

func TestConnectivityUnchangedFollowingTagModification(t *testing.T) {
	o := BuildOptions{Cores: 2}
	w, err := NewWorldFromPBFFile(test.Data(test.CamdenPBF), &o)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	m := NewMutableOverlayWorld(w)

	entrance := FromOSMNodeID(osm.NodeID(4270651271))
	before := b6.AllAreas(m.FindAreasByPoint(entrance))
	if len(before) == 0 {
		t.Fatal("Expected entrace to be part of at least one area")
	}

	for _, a := range before {
		m.AddTag(a.FeatureID(), b6.Tag{Key: "#reachable", Value: "yes"})
	}

	after := b6.AllAreas(m.FindAreasByPoint(entrance))
	if len(before) != len(after) {
		t.Errorf("Expected %d areas, found %d", len(before), len(after))
	}
}

func TestWatchModifiedTags(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})
	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableTagsOverlayWorld(base)
	c, cancel := overlay.Watch()

	overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "wheelchair", Value: "yes"})
	var m ModifiedTag
	ok := true
	for ok {
		select {
		case m = <-c:
		default:
			ok = false
		}
	}
	cancel()
	overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "wheelchair", Value: "no"})
	expected := ModifiedTag{ID: caravan.FeatureID(), Tag: b6.Tag{Key: "wheelchair", Value: "yes"}, Deleted: false}
	if m != expected {
		t.Errorf("Expected %v, found %v", expected, m)
	}
}

func TestPropagateModifiedTags(t *testing.T) {
	base := NewBasicMutableWorld()

	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})
	if err := base.AddPoint(caravan); err != nil {
		t.Fatal(err)
	}

	yumchaa := osmPoint(3790640853, 51.5355955, -0.1250640)
	if err := base.AddPoint(yumchaa); err != nil {
		t.Fatal(err)
	}

	granary := osmPath(222021576, []*PointFeature{caravan, yumchaa})
	if err := base.AddPath(granary); err != nil {
		t.Fatal(err)
	}

	w := NewMutableTagsOverlayWorld(base)
	w.AddTag(caravan.FeatureID(), b6.Tag{Key: "amenity", Value: "restaurant"})

	found := b6.FindPointByID(caravan.PointID, w)
	if found == nil {
		t.Fatal("Failed to find feature")
	}
	if found.Get("amenity").Value != "restaurant" {
		t.Error("Failed to find expected tag value")
	}

	path := b6.FindPathByID(granary.PathID, w)
	if path == nil {
		t.Fatal("Failed to find feature")
	}
	for i := 0; i < path.Len(); i++ {
		if found := path.Feature(i); found.PointID() == caravan.PointID && found.Get("amenity").Value != "restaurant" {
			t.Error("Failed to find expected tag value")
		}
	}

	ss := w.Traverse(caravan.PointID)
	for ss.Next() {
		segment := ss.Segment()
		for i := 0; i < segment.Feature.Len(); i++ {
			if found := segment.Feature.Feature(i); found.PointID() == caravan.PointID {
				if found.Get("amenity").Value != "restaurant" {
					t.Errorf("Failed to find expected tag value")
				}
			}
		}
	}
}

func TestModifiedTagsOnPathWithPlainLatLngs(t *testing.T) {
	base := NewBasicMutableWorld()

	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	yumchaa := osmPoint(3790640853, 51.5355955, -0.1250640)

	for _, p := range []*PointFeature{caravan, yumchaa} {
		if err := base.AddPoint(p); err != nil {
			t.Fatal(err)
		}
	}

	path := NewPathFeature(3)
	path.PathID = FromOSMWayID(222021576)
	path.SetPointID(0, caravan.PointID)
	path.SetLatLng(1, s2.LatLngFromDegrees(51.535490, -0.125167))
	path.SetPointID(2, yumchaa.PointID)

	if err := base.AddPath(path); err != nil {
		t.Fatal(err)
	}

	w := NewMutableTagsOverlayWorld(base)
	w.AddTag(path.FeatureID(), b6.Tag{Key: "#highway", Value: "path"})

	path2 := b6.FindPathByID(path.PathID, w)
	if path2 == nil {
		t.Fatal("Failed to find feature")
	}
	if path2.Feature(0).FeatureID() != caravan.FeatureID() {
		t.Errorf("Expected ID %s, found %s", caravan.FeatureID(), path2.Feature(0).FeatureID())
	}
	if path2.Feature(1) != nil {
		t.Error("Expected nil feature for lat, lng")
	}

}

func TestModifiedTagsFeatureFromArea(t *testing.T) {
	path := NewPathFeature(5)
	path.PathID = FromOSMWayID(265714033)
	path.SetLatLng(0, s2.LatLngFromDegrees(51.5369431, -0.1231868))
	path.SetLatLng(1, s2.LatLngFromDegrees(51.5365692, -0.1230608))
	path.SetLatLng(2, s2.LatLngFromDegrees(51.5365536, -0.1229421))
	path.SetLatLng(3, s2.LatLngFromDegrees(51.5367378, -0.1229110))
	path.SetLatLng(4, s2.LatLngFromDegrees(51.5369431, -0.1231868))

	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(265714033)
	area.SetPathIDs(0, []b6.PathID{path.PathID})
	area.AddTag(b6.Tag{Key: "#leisure", Value: "playground"})

	base := NewBasicMutableWorld()
	if err := base.AddPath(path); err != nil {
		t.Fatalf("Failed to add path: %s", err)
	}
	if err := base.AddArea(area); err != nil {
		t.Fatalf("Failed to add area: %s", err)
	}

	overlay := NewMutableOverlayWorld(base)
	overlay.AddTag(path.FeatureID(), b6.Tag{Key: "barrier", Value: "fence"})

	found := b6.FindAreaByID(area.AreaID, overlay)
	if found == nil {
		t.Fatal("Expected to find feature")
	}

	if b := found.Feature(0)[0].Get("barrier"); !b.IsValid() || b.Value != "fence" {
		t.Error("Expected to find added tag value")
	}
}

func TestSnapshot(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})
	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "wheelchair", Value: "yes"})
	snapshot := overlay.Snapshot()
	overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "wheelchair", Value: "no"})

	if wheelchair := overlay.FindFeatureByID(caravan.FeatureID()).Get("wheelchair"); wheelchair.Value != "no" {
		t.Errorf("Expected wheelchair=no in overlay, found %s", wheelchair.Value)
	}

	if wheelchair := snapshot.FindFeatureByID(caravan.FeatureID()).Get("wheelchair"); wheelchair.Value != "yes" {
		t.Errorf("Expected wheelchair=yes in snapshot, found %s", wheelchair.Value)
	}
}

func TestChangeTagsOnNonExistantPoint(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	if overlay.FindFeatureByID(FromOSMNodeID(427900370).FeatureID()) != nil {
		t.Error("Expected no features to be returned")
	}
}

func TestModifyingFeaturesWhileQueryingPanics(t *testing.T) {
	lighterman := osmPoint(427900370, 51.535242, -0.124388)
	lighterman.AddTag(b6.Tag{Key: "name", Value: "The Lighterman"})
	lighterman.AddTag(b6.Tag{Key: "wheelchair", Value: "no"})

	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})
	caravan.AddTag(b6.Tag{Key: "amenity", Value: "restaurant"})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan, lighterman); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic during query")
		}
	}()

	i := overlay.FindFeatures(b6.All{})
	for i.Next() {
		overlay.AddPoint(caravan)
	}
}

func TestEachFeatureWithAPathDependingOnTheBaseWorld(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})

	dishoom := osmPoint(3501612811, 51.536454, -0.126826)
	dishoom.AddTag(b6.Tag{Key: "name", Value: "Dishoom"})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan, dishoom); err != nil {
		t.Fatal(err)
	}

	footway := osmPath(558345071, []*PointFeature{caravan, dishoom})
	footway.AddTag(b6.Tag{Key: "highway", Value: "footway"})
	m := NewMutableOverlayWorld(base)
	if err := m.AddPath(footway); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	found := false
	each := func(f b6.Feature, goroutine int) error {
		if f.FeatureID() == footway.FeatureID() {
			found = true
			path := f.(b6.PathFeature)
			for i := 0; i < path.Len(); i++ {
				// Ensure features are wrapped correctly to be able to
				// lookup points in the base world.
				path.Point(i)
			}
		}
		return nil
	}
	m.EachFeature(each, &b6.EachFeatureOptions{})
	if !found {
		t.Errorf("Expected to find added path")
	}
}

func TestEachFeatureWithModifiedTags(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})
	caravan.AddTag(b6.Tag{Key: "cuisine", Value: "coffee_shop"})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan); err != nil {
		t.Fatal(err)
	}

	m := NewMutableOverlayWorld(base)
	m.AddTag(caravan.FeatureID(), b6.Tag{Key: "wheelchair", Value: "yes"})
	m.RemoveTag(caravan.FeatureID(), "cuisine")

	found := false
	each := func(f b6.Feature, goroutine int) error {
		if f.FeatureID() == caravan.FeatureID() {
			found = true
			if wheelchair := f.Get("wheelchair"); wheelchair.Value != "yes" {
				t.Errorf("Expected to find wheelchair=yes, found %s", wheelchair)
			}
			if cuisine := f.Get("cuisine"); cuisine.IsValid() {
				t.Errorf("Expected to cuisine to be removed, found %s", cuisine)
			}
		}
		return nil
	}
	m.EachFeature(each, &b6.EachFeatureOptions{})
	if !found {
		t.Errorf("Expected to find caravan")
	}
}

func TestMergeWorlds(t *testing.T) {
	lighterman := osmPoint(427900370, 51.535242, -0.124388)
	lighterman.AddTag(b6.Tag{Key: "name", Value: "The Lighterman"})
	lighterman.AddTag(b6.Tag{Key: "wheelchair", Value: "no"})

	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})
	caravan.AddTag(b6.Tag{Key: "amenity", Value: "restaurant"})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, lighterman); err != nil {
		t.Fatal(err)
	}

	lower := NewMutableOverlayWorld(base)
	upper := NewMutableOverlayWorld(lower)
	i := lower.FindFeatures(b6.All{})
	for i.Next() {
		upper.AddPoint(caravan)
	}

	if err := upper.MergeInto(lower); err != nil {
		t.Errorf("Expected no error from merge, found: %s", err)
	}

	if point := lower.FindFeatureByID(caravan.FeatureID()); point == nil {
		t.Error("Expected to find caravan in the lower world.")
	}
}

func TestSortAndDiffTokens(t *testing.T) {
	before := []string{"amenity=pub", "building=yes", "addr:city=London", "building:levels=1"}
	after := []string{"amenity=pub", "building=yes", "building:levels=3", "addr:housenumber=3"}

	added, removed := sortAndDiffTokens(before, after)

	if diff := cmp.Diff([]string{"addr:housenumber=3", "building:levels=3"}, added); diff != "" {
		t.Errorf("Got diff in 'added':\n%s", diff)
	}

	if diff := cmp.Diff([]string{"addr:city=London", "building:levels=1"}, removed); diff != "" {
		t.Errorf("Got diff in 'removed':\n%s", diff)
	}
}
