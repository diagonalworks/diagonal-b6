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
		basic, err := buildBasicWorld([]osm.Node{}, []osm.Way{}, []osm.Relation{}, &BuildOptions{Cores: 2})
		if err != nil {
			return nil, err
		}
		return NewMutableOverlayWorld(basic), nil
	}},
}

func osmPoint(id osm.NodeID, lat float64, lng float64) Feature {
	return &GenericFeature{
		ID:   FromOSMNodeID(id).FeatureID(),
		Tags: []b6.Tag{{Key: b6.PointTag, Value: b6.PointExpression(s2.LatLngFromDegrees(lat, lng))}}}
}

func osmPath(id osm.WayID, points []Feature) *GenericFeature {
	path := GenericFeature{}
	path.SetFeatureID(FromOSMWayID(id))
	for i, point := range points {
		path.ModifyOrAddTagAt(b6.Tag{b6.PathTag, point.FeatureID()}, i)
	}
	return &path
}

func osmSimpleArea(id osm.WayID) *AreaFeature {
	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(id)
	area.SetPathIDs(0, []b6.FeatureID{FromOSMWayID(id)})
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

func simpleCollection(id b6.CollectionID, key, value string) *CollectionFeature {
	return &CollectionFeature{
		CollectionID: id,
		Keys:         []interface{}{key},
		Values:       []interface{}{value},
	}
}

func addFeatures(w MutableWorld, features ...Feature) error {
	for _, f := range features {
		if err := w.AddFeature(f); err != nil {
			return err
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
		{"UpdateCollectionsByFeatureWhenChangingCollections", ValidateUpdateCollectionsByFeatureWhenChangingCollections},
		{"UpdatingPathUpdatesS2CellIndex", ValidateUpdatingPathUpdatesS2CellIndex},
		{"UpdatingPointLocationsUpdatesS2CellIndex", ValidateUpdatingPointLocationsUpdatesS2CellIndex},
		{"UpdatingPointLocationsWillFailIfAreasAreInvalidated", ValidateUpdatingPointLocationsWillFailIfAreasAreInvalidated},
		{"UpdatingPathWillFailIfAreasAreInvalidated", ValidateUpdatingPathWillFailIfAreasAreInvalidated},
		{"RepeatedModification", ValidateRepeatedModification},
		{"AddingFeaturesWithNoIDFails", ValidateAddingFeaturesWithNoIDFails},
		{"AddTagToExistingFeature", ValidateAddTagToExistingFeature},
		{"AddSearchableTagToExistingFeature", ValidateAddSearchableTagToExistingFeature},
		{"ChangeSearchableTagOnExistingFeature", ValidateChangeSearchableTagOnExistingFeature},
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
	path := osmPath(558345071, []Feature{start, end})
	path.ModifyOrAddTag(b6.Tag{Key: "#highway", Value: b6.StringExpression("footway")})

	if err := addFeatures(w, start, end, path); err != nil {
		t.Fatal(err)
	}

	paths := b6.AllFeatures(w.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.Tagged{Key: "#highway", Value: b6.StringExpression("footway")}}))
	if len(paths) != 1 || paths[0].FeatureID() != path.FeatureID() {
		t.Error("Expected to find one path")
	}

	paths = b6.AllFeatures(w.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.Tagged{Key: "#bridge", Value: b6.StringExpression("yes")}}))
	if len(paths) != 0 {
		t.Error("Didn't expect to find any bridges")
	}

	path.Tags = append(path.Tags, b6.Tag{Key: "#bridge", Value: b6.StringExpression("yes")})
	w.AddFeature(path)
	paths = b6.AllFeatures(w.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.Intersection{b6.Tagged{Key: "#highway", Value: b6.StringExpression("footway")}, b6.Tagged{Key: "#bridge", Value: b6.StringExpression("yes")}}}))
	if len(paths) != 1 || paths[0].FeatureID() != path.FeatureID() {
		t.Errorf("Expected to find one path, found %d", len(paths))
	}
}

func ValidateUpdatePathConnectivity(w MutableWorld, t *testing.T) {
	a := osmPoint(5384190463, 51.5358664, -0.1272493)
	b := osmPoint(5384190494, 51.5362126, -0.1270125)
	c := osmPoint(5384190476, 51.5367563, -0.1266297)

	ab := osmPath(558345071, []Feature{a, b})
	ca := osmPath(558345054, []Feature{c, a})

	if err := addFeatures(w, a, b, c, ab, ca); err != nil {
		t.Fatal(err)
	}

	segments := b6.AllSegments(w.Traverse(c.FeatureID()))
	if len(segments) != 1 || segments[0].LastFeatureID() != a.FeatureID() {
		t.Error("Expected to find a connection to point a")
	}

	// Swap pathC from c -> a to c -> b
	path := NewFeatureFromWorld(w.FindFeatureByID(ca.FeatureID())).(Feature)
	path.ModifyOrAddTagAt(b6.Tag{b6.PathTag, b.FeatureID()}, 1)
	if w.AddFeature(path) != nil {
		t.Error("Failed to swap path c -> b")
	}

	segments = b6.AllSegments(w.Traverse(c.FeatureID()))
	if len(segments) != 1 || segments[0].LastFeatureID() != b.FeatureID() {
		t.Errorf("Expected to find a connection to point b, found none (%d segments)", len(segments))
	}
}

func ValidateUpdateAreasByPointWhenChangingPointsOnAPath(w MutableWorld, t *testing.T) {
	a := osmPoint(2309943870, 51.5371371, -0.1240464) // Top left
	b := osmPoint(2309943835, 51.5355393, -0.1247150)
	c := osmPoint(2309943825, 51.5354848, -0.1243698)

	d := osmPoint(2309943868, 51.5370710, -0.1240744)

	path := osmPath(222021570, []Feature{a, b, c, a})

	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(222021570)
	area.SetPathIDs(0, []b6.FeatureID{path.FeatureID()})

	if err := addFeatures(w, a, b, c, d, path, area); err != nil {
		t.Fatal(err)
	}

	areas := b6.AllAreas(w.FindAreasByPoint(c.FeatureID()))
	if len(areas) != 1 || areas[0].AreaID() != area.AreaID {
		t.Error("Expected to find an area for point c")
	}

	areas = b6.AllAreas(w.FindAreasByPoint(d.FeatureID()))
	if len(areas) != 0 {
		t.Error("Didn't expect to find an area for point d")
	}

	// Replace a point on the path
	modifiedPath := NewFeatureFromWorld(w.FindFeatureByID(path.FeatureID())).(Feature)
	modifiedPath.ModifyOrAddTagAt(b6.Tag{b6.PathTag, d.FeatureID()}, 1)
	if err := addFeatures(w, modifiedPath); err != nil {
		t.Fatal(err)
	}

	areas = b6.AllAreas(w.FindAreasByPoint(d.FeatureID()))
	if len(areas) != 1 || areas[0].AreaID() != area.AreaID {
		t.Errorf("Expected to find an area for point d, found: %d", len(areas))
	}

	areas = b6.AllAreas(w.FindAreasByPoint(b.FeatureID()))
	if len(areas) != 0 {
		t.Errorf("Expected to find no areas for point b, found: %d", len(areas))
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

	pathA := osmPath(222021570, []Feature{a, b, c, a})
	areaA := NewAreaFeature(1)
	areaA.AreaID = AreaIDFromOSMWayID(222021570)
	areaA.SetPathIDs(0, []b6.FeatureID{pathA.FeatureID()})

	pathB := osmPath(222021576, []Feature{a, d, e, a})
	areaB := NewAreaFeature(1)
	areaB.AreaID = AreaIDFromOSMWayID(222021576)
	areaB.SetPathIDs(0, []b6.FeatureID{pathB.FeatureID()})

	pathC := osmPath(222021578, []Feature{a, f, g, a})
	areaC := NewAreaFeature(1)
	areaC.AreaID = AreaIDFromOSMWayID(222021578)
	areaC.SetPathIDs(0, []b6.FeatureID{pathC.FeatureID()})

	if err := addFeatures(w, a, b, c, d, e, f, g, h, pathA, pathB, pathC, areaA, areaB, areaC); err != nil {
		t.Fatal(err)
	}

	// Replace points on pathA
	modifiedPath := NewFeatureFromWorld(w.FindFeatureByID(pathA.FeatureID())).(Feature)
	modifiedPath.ModifyOrAddTagAt(b6.Tag{b6.PathTag, h.FeatureID()}, 0)
	modifiedPath.ModifyOrAddTagAt(b6.Tag{b6.PathTag, h.FeatureID()}, 3)
	if err := addFeatures(w, modifiedPath); err != nil {
		t.Fatal(err)
	}

	areas := b6.AllAreas(w.FindAreasByPoint(h.FeatureID()))
	if len(areas) != 1 || areas[0].AreaID() != areaA.AreaID {
		t.Errorf("Expected to find an area for point h, found: %d", len(areas))
	}

	areas = b6.AllAreas(w.FindAreasByPoint(a.FeatureID()))
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

	pathA := osmPath(222021570, []Feature{a, b, c, a})
	areaA := NewAreaFeature(1)
	areaA.AreaID = AreaIDFromOSMWayID(222021570)
	areaA.SetPathIDs(0, []b6.FeatureID{pathA.FeatureID()})

	pathB := osmPath(222021576, []Feature{b, d, e, b})
	areaB := NewAreaFeature(1)
	areaB.AreaID = AreaIDFromOSMWayID(222021576)
	areaB.SetPathIDs(0, []b6.FeatureID{pathB.FeatureID()})

	if err := addFeatures(w, a, b, c, d, e, pathA, pathB, areaA, areaB); err != nil {
		t.Fatal(err)
	}

	// Move the shared point
	modifiedPoint := osmPoint(2309943835, 51.5355185, -0.1254976)
	if err := addFeatures(w, modifiedPoint); err != nil {
		t.Fatal(err)
	}

	point := w.FindFeatureByID(modifiedPoint.FeatureID())
	if point == nil {
		t.Fatal("Expected to find point")
	}

	if point.(b6.Geometry).Point().Distance(modifiedPoint.(b6.Geometry).Point()) > 0.000001 {
		t.Error("Expected modified point to have an updated location")
	}
}

func ValidateUpdateAreasByPointWhenChangingPathsForAnArea(w MutableWorld, t *testing.T) {
	a := osmPoint(2309943835, 51.5355393, -0.1247150)
	b := osmPoint(2309943825, 51.5354848, -0.1243698)
	c := osmPoint(2309943870, 51.5371371, -0.1240464)
	d := osmPoint(598093309, 51.5321649, -0.1269834)

	path := osmPath(222021570, []Feature{a, b, c, a})

	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(222021570)
	area.SetPathIDs(0, []b6.FeatureID{path.FeatureID()})

	if err := addFeatures(w, a, b, c, d, path, area); err != nil {
		t.Fatal(err)
	}

	newPath := osmPath(222021577, []Feature{b, c, d, b})
	modifiedArea := NewAreaFeatureFromWorld(b6.FindAreaByID(area.AreaID, w))
	modifiedArea.SetPathIDs(0, []b6.FeatureID{newPath.FeatureID()})

	if err := addFeatures(w, newPath); err != nil {
		t.Fatal(err)
	}

	if err := w.AddFeature(modifiedArea); err != nil {
		t.Errorf("Failed to add modified area: %s", err)
	}

	areas := b6.AllAreas(w.FindAreasByPoint(d.FeatureID()))
	if len(areas) != 1 || areas[0].AreaID() != modifiedArea.AreaID {
		t.Error("Expected to find an area for point d")
	}

	areas = b6.AllAreas(w.FindAreasByPoint(a.FeatureID()))
	if len(areas) != 0 {
		t.Error("Didn't expect to find an area for point a")
	}
}

func ValidateUpdateRelationsByFeatureWhenChangingRelations(w MutableWorld, t *testing.T) {
	a := osmPoint(5378333625, 51.5352195, -0.1254286)
	b := osmPoint(5384190491, 51.5352339, -0.1255240)
	c := osmPoint(4966136655, 51.5349570, -0.1256696)

	ab := &GenericFeature{ID: FromOSMWayID(807925586)}
	ab.ModifyOrAddTag(b6.Tag{b6.PathTag, b6.Values([]b6.Value{a.FeatureID(), b.FeatureID()})})

	bc := &GenericFeature{ID: FromOSMWayID(558345068)}
	bc.ModifyOrAddTag(b6.Tag{b6.PathTag, b6.Values([]b6.Value{b.FeatureID(), c.FeatureID()})})

	relation := NewRelationFeature(1)
	relation.RelationID = FromOSMRelationID(11139964)
	relation.Members[0] = b6.RelationMember{ID: ab.FeatureID()}
	relation.Tags = []b6.Tag{{Key: "type", Value: b6.StringExpression("route")}}

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
	if err := w.AddFeature(modifiedRelation); err != nil {
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

func ValidateUpdateCollectionsByFeatureWhenChangingCollections(w MutableWorld, t *testing.T) {
	c := simpleCollection(b6.MakeCollectionID(b6.Namespace("test"), 1), "repetition", "difference")

	if err := addFeatures(w, c); err != nil {
		t.Fatal(err)
	}

	original := w.FindFeatureByID(c.FeatureID())
	if original.FeatureID().ToCollectionID() != c.CollectionID {
		t.Error("Expected to find a collection")
	}

	modifiedCollection := NewCollectionFeatureFromWorld(b6.FindCollectionByID(c.CollectionID, w))
	modifiedCollection.Keys = append(modifiedCollection.Keys, "extension")
	modifiedCollection.Values = append(modifiedCollection.Values, "intensity")

	if err := w.AddFeature(modifiedCollection); err != nil {
		t.Errorf("Failed to add collection: %s", err)
	}

	updated := b6.FindCollectionByID(c.CollectionID, w)

	i := updated.BeginUntyped()
	if ok, err := i.Next(); err != nil || !ok {
		t.Error("Expected to find two key / value pairs")
	}
	if i.Key() != "repetition" {
		t.Errorf("Expected to find initial key, got %v", i.Key())
	}

	if ok, err := i.Next(); err != nil || !ok {
		t.Error("Expected to find two key / value pairs")
	}
	if i.Value() != "intensity" {
		t.Errorf("Expected to find added value, got %v", i.Value())
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

	path := osmPath(222021577, []Feature{a, b, c, d, a})
	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(222021577)
	area.SetPathIDs(0, []b6.FeatureID{path.FeatureID()})

	if err := addFeatures(w, a, b, c, d, path, area); err != nil {
		t.Fatal(err)
	}

	// A cap covering only part of the Eastern Shed
	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.5370349, -0.1232719)), b6.MetersToAngle(10))
	areas := b6.AllAreas(b6.FindAreas(b6.MightIntersect{cap}, w))
	if len(areas) != 0 {
		t.Errorf("Didn't expect to find an area %s", areas[0].FeatureID())
	}

	modifiedPath := NewFeatureFromWorld(w.FindFeatureByID(path.FeatureID())).(Feature)
	modifiedPath.ModifyOrAddTagAt(b6.Tag{b6.PathTag, e.FeatureID()}, 2)
	modifiedPath.ModifyOrAddTagAt(b6.Tag{b6.PathTag, f.FeatureID()}, 3)
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

	path := osmPath(222021577, []Feature{a, b, c, d, a})
	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(222021577)
	area.SetPathIDs(0, []b6.FeatureID{path.FeatureID()})

	if err := addFeatures(w, a, b, c, d, path, area); err != nil {
		t.Fatal(err)
	}

	// A cap covering only part of the Eastern Shed
	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.5370349, -0.1232719)), b6.MetersToAngle(10))
	areas := b6.AllAreas(b6.FindAreas(b6.NewIntersectsCap(cap), w))
	if len(areas) != 0 {
		t.Errorf("Didn't expect to find an area %s", areas[0].FeatureID())
	}

	// Eastern Handyside Canopy, East edge
	newC := &GenericFeature{ID: c.FeatureID(), Tags: []b6.Tag{{Key: b6.PointTag, Value: b6.PointExpression(s2.LatLngFromDegrees(51.5358965, -0.1230551))}}}
	newD := &GenericFeature{ID: d.FeatureID(), Tags: []b6.Tag{{Key: b6.PointTag, Value: b6.PointExpression(s2.LatLngFromDegrees(51.5370349, -0.1232719))}}}

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

	path := osmPath(222021577, []Feature{a, b, c, d, a})
	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(222021577)
	area.SetPathIDs(0, []b6.FeatureID{path.FeatureID()})

	if err := addFeatures(w, a, b, c, d, path, area); err != nil {
		t.Fatal(err)
	}

	modifiedPoint := &GenericFeature{ID: c.FeatureID(), Tags: []b6.Tag{{Key: b6.PointTag, Value: b6.PointExpression(s2.LatLngFromDegrees(51.5368549, -0.1256275))}}}

	if addFeatures(w, modifiedPoint) == nil {
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

	path := osmPath(222021577, []Feature{a, b, c, d, a})
	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(222021577)
	area.SetPathIDs(0, []b6.FeatureID{path.FeatureID()})

	if err := addFeatures(w, a, b, c, d, e, path, area); err != nil {
		t.Fatal(err)
	}

	modifiedPath := NewFeatureFromWorld(w.FindFeatureByID(path.FeatureID())).(Feature)
	modifiedPath.ModifyOrAddTagAt(b6.Tag{b6.PathTag, e.FeatureID()}, 2)

	if w.AddFeature(modifiedPath) == nil {
		t.Error("Expected adding path to fail, as it invalidates an area")
	}
}

func ValidateAddingFeaturesWithNoIDFails(w MutableWorld, t *testing.T) {
	if err := w.AddFeature(&GenericFeature{}); err == nil {
		t.Error("Expected adding a feature with no ID to fail")
	}
	if err := w.AddFeature(&GenericFeature{}); err == nil {
		t.Error("Expected adding a path with no ID to fail")
	}
	if err := w.AddFeature(&AreaFeature{}); err == nil {
		t.Error("Expected adding an area with no ID to fail")
	}
	if err := w.AddFeature(&RelationFeature{}); err == nil {
		t.Error("Expected adding a relation with no ID to fail")
	}
	if err := w.AddFeature(&CollectionFeature{}); err == nil {
		t.Error("Expected adding a collection with no ID to fail")
	}
}

func ValidateAddTagToExistingFeature(w MutableWorld, t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Caravan")})
	if err := addFeatures(w, caravan); err != nil {
		t.Fatal(err)
	}

	if err := w.AddTag(caravan.FeatureID(), b6.Tag{Key: "amenity", Value: b6.StringExpression("restaurant")}); err != nil {
		t.Fatalf("Failed to add tag: %s", err)
	}

	found := w.FindFeatureByID(caravan.FeatureID())
	if found == nil {
		t.Fatal("Failed to find feature")
	}
	if found.Get("amenity").Value.String() != "restaurant" {
		t.Error("Failed to find expected tag value")
	}
}

func ValidateRepeatedModification(w MutableWorld, t *testing.T) {
	id := osm.WayID(222021577)
	finders := []struct {
		name string
		f    func(w b6.World) b6.Feature
	}{
		{"ByID", func(w b6.World) b6.Feature {
			return w.FindFeatureByID(FromOSMWayID(id))
		}},
		{"BySearch", func(w b6.World) b6.Feature {
			highways := b6.AllFeatures(w.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.Keyed{"#highway"}}))
			if len(highways) != 1 {
				t.Fatalf("Expected to find 1 path, found %d", len(highways))
			}
			return highways[0]
		}},
	}

	tests := []struct {
		name string
		f    func(id osm.WayID, find func(b6.World) b6.Feature, w MutableWorld, t *testing.T)
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

func ValidateRepeatedModificationChangingNodes(id osm.WayID, find func(b6.World) b6.Feature, w MutableWorld, t *testing.T) {
	// Start with area abcda, repeatedly modify it to increment the ID
	// of each point by 1
	a := osmPoint(2309943873, 51.5373249, -0.1251784)
	b := osmPoint(2309943847, 51.5357239, -0.1258568)
	c := osmPoint(2309943846, 51.5356657, -0.1254957)
	d := osmPoint(2309943872, 51.5372656, -0.1248160)

	path := osmPath(id, []Feature{a, b, c, d, a})
	path.ModifyOrAddTag(b6.Tag{Key: "#highway", Value: b6.StringExpression("primary")})
	if err := addFeatures(w, a, b, c, d, path); err != nil {
		t.Fatal(err)
	}

	points := []Feature{a, b, c, d}
	for i := 0; i < len(points); i++ {
		modifiedPoint := points[i].Clone()
		modifiedPoint.SetFeatureID(b6.FeatureID{b6.FeatureTypePoint, points[i].FeatureID().Namespace, points[i].FeatureID().Value + 1})
		if err := w.AddFeature(modifiedPoint); err != nil {
			t.Fatal(err)
		}
		found := find(w)
		if found == nil {
			// TODO Do we need to fail or return?
			t.Fatal("Failed to find feature")
		}
		modifiedPath := NewFeatureFromWorld(found).(Feature)
		modifiedPath.ModifyOrAddTagAt(b6.Tag{b6.PathTag, modifiedPoint.FeatureID()}, i)
		if err := w.AddFeature(modifiedPath); err != nil {
			t.Fatal(err)
		}
	}

	found := w.FindFeatureByID(path.FeatureID())
	for i := 0; i < len(points); i++ {
		id := found.Reference(i).Source()
		if id.Value != points[i].FeatureID().Value+1 {
			t.Errorf("Expected to find ID %d for point %d, found %d", points[i].FeatureID().Value+1, i, id.Value)
		}
	}
}

func ValidateRepeatedModificationAddingNodes(id osm.WayID, find func(b6.World) b6.Feature, w MutableWorld, t *testing.T) {
	// Start with path ab, repeatedly add points to it to turn it into aedcb
	a := osmPoint(2309943873, 51.5373249, -0.1251784)
	b := osmPoint(2309943847, 51.5357239, -0.1258568)
	c := osmPoint(2309943846, 51.5356657, -0.1254957)
	d := osmPoint(2309943872, 51.5372656, -0.1248160)
	e := osmPoint(4031177264, 51.5368549, -0.1256275)

	path := osmPath(id, []Feature{a, b})
	path.ModifyOrAddTag(b6.Tag{Key: "#highway", Value: b6.StringExpression("primary")})
	if err := addFeatures(w, a, b, c, d, e, path); err != nil {
		t.Fatal(err)
	}

	newPoints := []Feature{c, d, e}
	for i := 0; i < len(newPoints); i++ {
		found := find(w)
		if found == nil {
			// TODO Do we need to fail or return?
			t.Fatal("Failed to find feature")
		}
		newPath := &GenericFeature{ID: found.FeatureID()}
		newPath.Tags = found.AllTags().Clone()
		insert := 0
		for j := 0; j < len(found.References()); j++ {
			if j == 1 {
				newPath.ModifyOrAddTagAt(b6.Tag{b6.PathTag, newPoints[i].FeatureID()}, insert)
				insert++
			}
			id := found.Reference(j).Source()
			newPath.ModifyOrAddTagAt(b6.Tag{b6.PathTag, id}, insert)
			insert++
		}
		if err := w.AddFeature(newPath); err != nil {
			t.Fatal(err)
		}
	}

	found := w.FindFeatureByID(path.FeatureID())
	var gotPointIDs []b6.FeatureID
	for i := 0; i < len(found.References()); i++ {
		id := found.Reference(i).Source()
		gotPointIDs = append(gotPointIDs, id)
	}
	expected := []b6.FeatureID{a.FeatureID(), e.FeatureID(), d.FeatureID(), c.FeatureID(), b.FeatureID()}
	if diff := cmp.Diff(expected, gotPointIDs); diff != "" {
		t.Errorf("Unexpected point IDs (-want +got):\n%s", diff)
	}
}

func ValidateAddSearchableTagToExistingFeature(w MutableWorld, t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Caravan")})

	lighterman := osmPoint(427900370, 51.5353986, -0.1243711)
	lighterman.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("The Lighterman")})
	lighterman.AddTag(b6.Tag{Key: "#amenity", Value: b6.StringExpression("restaurant")})

	if err := addFeatures(w, caravan, lighterman); err != nil {
		t.Fatal(err)
	}

	points := w.FindFeatures(b6.Typed{b6.FeatureTypePoint, b6.Tagged{Key: "#amenity", Value: b6.StringExpression("restaurant")}})
	if points.Next() != true && points.Next() != false {
		t.Fatal("Expected to find 1 point")
	}

	if err := w.AddTag(caravan.FeatureID(), b6.Tag{Key: "#amenity", Value: b6.StringExpression("restaurant")}); err != nil {
		t.Fatalf("Failed to add tag: %s", err)
	}

	foundPoint := w.FindFeatureByID(caravan.FeatureID())
	if foundPoint == nil {
		t.Fatal("Failed to find feature")
	}
	if foundPoint.Get("#amenity").Value.String() != "restaurant" {
		t.Errorf("Failed to find expected tag value")
	}

	points = w.FindFeatures(b6.Typed{b6.FeatureTypePoint, b6.Tagged{Key: "#amenity", Value: b6.StringExpression("restaurant")}})
	pointLen := 0
	found := false
	for points.Next() {
		pointLen++
		if points.FeatureID() == caravan.FeatureID() {
			found = true
		}
	}

	if pointLen != 2 {
		t.Fatalf("Expected to find 2 points, found %d", pointLen)
	}

	if !found {
		t.Error("Expected to find Caravan")
	}
}

func ValidateChangeSearchableTagOnExistingFeature(w MutableWorld, t *testing.T) {
	lighterman := osmPoint(427900370, 51.5353986, -0.1243711)
	lighterman.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("The Lighterman")})
	lighterman.AddTag(b6.Tag{Key: "#amenity", Value: b6.StringExpression("restaurant")})

	if err := addFeatures(w, lighterman); err != nil {
		t.Fatal(err)
	}

	points := w.FindFeatures(b6.Typed{b6.FeatureTypePoint, b6.Tagged{Key: "#amenity", Value: b6.StringExpression("restaurant")}})
	if points.Next() != true && points.Next() != false {
		t.Fatal("Expected to find 1 point")
	}

	if err := w.AddTag(lighterman.FeatureID(), b6.Tag{Key: "#amenity", Value: b6.StringExpression("pub")}); err != nil {
		t.Fatalf("Failed to add tag: %s", err)
	}

	points = w.FindFeatures(b6.Typed{b6.FeatureTypePoint, b6.Tagged{Key: "#amenity", Value: b6.StringExpression("restaurant")}})
	if points.Next() != false {
		t.Fatal("Expected to find 0 points")
	}
}

func ValidateAddTagToNonExistingFeature(w MutableWorld, t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	if err := w.AddTag(caravan.FeatureID(), b6.Tag{Key: "#amenity", Value: b6.StringExpression("restaurant")}); err == nil {
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

	path := osmPath(222021577, []Feature{a, b, c, d, a})
	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(222021577)
	area.SetPathIDs(0, []b6.FeatureID{path.FeatureID()})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, c, d, g, path, area); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	if fountain := overlay.FindFeatureByID(e.FeatureID()); fountain != nil {
		t.Error("Expected to find fountain in base via overlay")
	}

	// A cap covering only part of the Eastern Shed
	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.5370349, -0.1232719)), b6.MetersToAngle(10))
	areas := b6.AllAreas(b6.FindAreas(b6.NewIntersectsCap(cap), overlay))
	if len(areas) != 0 {
		t.Errorf("Didn't expect to find an area %s", areas[0].FeatureID())
	}

	modifiedPath := NewFeatureFromWorld(overlay.FindFeatureByID(path.FeatureID())).(Feature)
	modifiedPath.ModifyOrAddTagAt(b6.Tag{b6.PathTag, e.FeatureID()}, 2)
	modifiedPath.ModifyOrAddTagAt(b6.Tag{b6.PathTag, f.FeatureID()}, 3)
	if err := addFeatures(overlay, e, f, modifiedPath); err != nil {
		t.Fatal(err)
	}

	areas = b6.AllAreas(b6.FindAreas(b6.NewIntersectsCap(cap), overlay))
	if len(areas) != 1 || areas[0].AreaID() != area.AreaID {
		t.Errorf("Expected to find area within the region (found %d areas)", len(areas))
	}

	paths := b6.AllFeatures(overlay.FindReferences(a.FeatureID(), b6.FeatureTypePath))
	if len(paths) != 1 || paths[0].FeatureID().Value != path.FeatureID().Value {
		t.Errorf("Expected to find 1 path by point a, found %d", len(paths))
	}
}

func TestModifyPointsOnPathInExistingWorld(t *testing.T) {
	// Move the Stable Street bridge to somewhere around Bank, by relocating
	// its points, to ensure indices are properly updated.
	a := osmPoint(1447052073, 51.5350350, -0.1247934)
	b := osmPoint(1540349979, 51.5348204, -0.1246405)

	path := osmPath(140633010, []Feature{a, b})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, path); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)

	granarySquareCap := s2.CapFromCenterAngle(s2.Interpolate(0.5, a.(b6.Geometry).Point(), b.(b6.Geometry).Point()), b6.MetersToAngle(10))
	paths := b6.AllFeatures(overlay.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.NewIntersectsCap(granarySquareCap)}))
	if len(paths) != 1 {
		t.Errorf("Expected to find 1 path around Granary Square, found %d", len(paths))
	}

	aPrime := osmPoint(1447052073, 51.5132689, -0.0988335)
	bPrime := osmPoint(1540349979, 51.5129188, -0.0985641)
	if err := addFeatures(base, aPrime, bPrime); err != nil {
		t.Fatal(err)
	}

	bankCap := s2.CapFromCenterAngle(s2.Interpolate(0.5, aPrime.(b6.Geometry).Point(), bPrime.(b6.Geometry).Point()), b6.MetersToAngle(10))
	paths = b6.AllFeatures(overlay.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.NewIntersectsCap(bankCap)}))
	if len(paths) != 1 {
		t.Errorf("Expected to find 1 path around Bank, found %d", len(paths))
	}

	paths = b6.AllFeatures(overlay.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.NewIntersectsCap(granarySquareCap)}))
	if len(paths) != 0 {
		t.Errorf("Expected to find no paths around Granary Square, found %d", len(paths))
	}

	paths = b6.AllFeatures(overlay.FindReferences(a.FeatureID(), b6.FeatureTypePath))
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

	path := osmPath(140633010, []Feature{a, b, c, a})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, c, path); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)

	granarySquareCap := s2.CapFromCenterAngle(s2.Interpolate(0.5, a.(b6.Geometry).Point(), b.(b6.Geometry).Point()), b6.MetersToAngle(10))
	paths := b6.AllFeatures(overlay.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.NewIntersectsCap(granarySquareCap)}))
	if len(paths) != 1 {
		t.Errorf("Expected to find 1 path around Granary Square, found %d", len(paths))
	}

	aPrime := osmPoint(4270651271, 51.5137306, -0.0905139)
	bPrime := osmPoint(5693730033, 51.5134981, -0.0898162)
	cPrime := osmPoint(4270651273, 51.5138208, -0.0896115)

	if err := addFeatures(base, aPrime, bPrime, cPrime); err != nil {
		t.Fatal(err)
	}

	bankCap := s2.CapFromCenterAngle(s2.Interpolate(0.5, aPrime.(b6.Geometry).Point(), bPrime.(b6.Geometry).Point()), b6.MetersToAngle(10))
	paths = b6.AllFeatures(overlay.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.NewIntersectsCap(bankCap)}))
	if len(paths) != 1 {
		t.Errorf("Expected to find 1 path around Bank, found %d", len(paths))
	}

	paths = b6.AllFeatures(overlay.FindFeatures(b6.Typed{b6.FeatureTypePath, b6.NewIntersectsCap(granarySquareCap)}))
	if len(paths) != 0 {
		t.Errorf("Expected to find no paths around Granary Square, found %d", len(paths))
	}

	paths = b6.AllFeatures(overlay.FindReferences(a.FeatureID(), b6.FeatureTypePath))
	if len(paths) != 1 || paths[0].FeatureID().Value != path.FeatureID().Value {
		t.Errorf("Expected 1 path, found: %d", len(paths))
		for _, p := range paths {
			log.Printf("  * %s", p.FeatureID())
		}
	}
}

func TestModifyPathWithIntersectionsInExistingWorld(t *testing.T) {
	a := osmPoint(6083741698, 51.5352814, -0.1266217)
	b := osmPoint(7787634237, 51.5354236, -0.1267632)
	c := osmPoint(6083735356, 51.5355776, -0.1268618)
	d := osmPoint(6083735379, 51.5361482, -0.1264835)

	e := osmPoint(7787634210, 51.5355869, -0.1269299)

	ad := osmPath(647895239, []Feature{a, b, c, d})
	ec := osmPath(647895212, []Feature{e, c})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, c, d, e, ad, ec); err != nil {
		t.Fatal(err)
	}
	overlay := NewMutableOverlayWorld(base)

	reachable := make(map[b6.FeatureID]bool)
	segments := overlay.Traverse(c.FeatureID())
	for segments.Next() {
		reachable[segments.Segment().Feature.FeatureID()] = true
	}

	if len(reachable) != 2 || !reachable[ad.FeatureID()] || !reachable[ec.FeatureID()] {
		t.Error("Didn't find expected initial connections")
	}

	newAD := osmPath(osm.WayID(ad.FeatureID().Value), []Feature{a, c, d})
	if err := addFeatures(overlay, newAD); err != nil {
		t.Fatal(err)
	}

	reachable = make(map[b6.FeatureID]bool)
	segments = overlay.Traverse(c.FeatureID())
	for segments.Next() {
		reachable[segments.Segment().Feature.FeatureID()] = true
	}

	if len(reachable) != 2 || !reachable[ad.FeatureID()] || !reachable[ec.FeatureID()] {
		t.Error("Expected modification to retain existing connections")
	}
}

func validateTags(tagged b6.Taggable, expected []b6.Tag) error {
	for _, tag := range expected {
		if found := tagged.Get(tag.Key); !found.IsValid() || found.Value.String() != tag.Value.String() {
			return fmt.Errorf("Expected value %q for tag %q, found %q", tag.Value, tag.Key, found.Value)
		}
	}

	if allTags := tagged.AllTags(); len(allTags) != len(expected) {
		return fmt.Errorf("Expected %d tags from AllTags(), found %d", len(expected), len(allTags))
	}

	tags := make(map[string]string)
	for _, tag := range tagged.AllTags() {
		tags[tag.Key] = tag.Value.String()
	}

	for _, tag := range expected {
		if value, ok := tags[tag.Key]; !ok || value != tag.Value.String() {
			return fmt.Errorf("Expected value %q for tag %q, found %q", tag.Value, tag.Key, value)
		}
	}
	return nil
}

func TestChangeTagsOnExistingPoint(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Caravan")})
	caravan.AddTag(b6.Tag{Key: "wheelchair", Value: b6.StringExpression("no")})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	// TODO: Updating the amenity actually needs to update the search index too.
	overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "amenity", Value: b6.StringExpression("restaurant")})
	overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "wheelchair", Value: b6.StringExpression("yes")})

	point := overlay.FindFeatureByID(caravan.FeatureID())

	expected := []b6.Tag{
		{Key: "name", Value: b6.StringExpression("Caravan")},
		{Key: "wheelchair", Value: b6.StringExpression("yes")},
		{Key: "amenity", Value: b6.StringExpression("restaurant")},
		{Key: b6.PointTag, Value: b6.PointExpression(s2.LatLngFromDegrees(51.5357237, -0.1253052))},
	}

	if err := validateTags(point, expected); err != nil {
		t.Error(err)
	}
}

func TestChangeTagsOnExistingPath(t *testing.T) {
	a := osmPoint(7555161584, 51.5345488, -0.1251005)
	b := osmPoint(6384669830, 51.5342291, -0.1262792)

	// Goods Way, South of Granary Square
	ab := osmPath(807924986, []Feature{a, b})
	ab.AddTag(b6.Tag{Key: "highway", Value: b6.StringExpression("tertiary")})
	ab.AddTag(b6.Tag{Key: "lit", Value: b6.StringExpression("no")})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, ab); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	overlay.AddTag(ab.FeatureID(), b6.Tag{Key: "cycleway:left", Value: b6.StringExpression("track")})
	overlay.AddTag(ab.FeatureID(), b6.Tag{Key: "lit", Value: b6.StringExpression("yes")})

	path := overlay.FindFeatureByID(ab.FeatureID())

	expected := []b6.Tag{
		{Key: "highway", Value: b6.StringExpression("tertiary")},
		{Key: "lit", Value: b6.StringExpression("yes")},
		{Key: "cycleway:left", Value: b6.StringExpression("track")},
		{Key: b6.PathTag, Value: b6.Values([]b6.Value{a.FeatureID(), b.FeatureID()})},
	}

	if err := validateTags(path, expected); err != nil {
		t.Error(err)
	}
}

func TestChangeTagsOnExistingArea(t *testing.T) {
	a := osmPoint(5693730034, 51.5352979, -0.1244842)
	b := osmPoint(5693730033, 51.5352871, -0.1244193)
	c := osmPoint(4270651271, 51.5353986, -0.1243711)
	abc := osmPath(427900370, []Feature{a, b, c, a})

	lighterman := osmSimpleArea(427900370)
	lighterman.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("The Lighterman")})
	lighterman.AddTag(b6.Tag{Key: "wheelchair", Value: b6.StringExpression("no")})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, c, abc, lighterman); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	overlay.AddTag(lighterman.FeatureID(), b6.Tag{Key: "amenity", Value: b6.StringExpression("pub")})
	overlay.AddTag(lighterman.FeatureID(), b6.Tag{Key: "wheelchair", Value: b6.StringExpression("yes")})

	area := b6.FindAreaByID(lighterman.AreaID, overlay)

	expected := []b6.Tag{
		{Key: "name", Value: b6.StringExpression("The Lighterman")},
		{Key: "amenity", Value: b6.StringExpression("pub")},
		{Key: "wheelchair", Value: b6.StringExpression("yes")},
	}

	if err := validateTags(area, expected); err != nil {
		t.Error(err)
	}
}

func TestAddSearchableTagTagToExistingArea(t *testing.T) {
	a := osmPoint(5693730034, 51.5352979, -0.1244842)
	b := osmPoint(5693730033, 51.5352871, -0.1244193)
	c := osmPoint(4270651271, 51.5353986, -0.1243711)
	abc := osmPath(427900370, []Feature{a, b, c, a})

	lighterman := osmSimpleArea(427900370)
	lighterman.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("The Lighterman")})
	lighterman.AddTag(b6.Tag{Key: "wheelchair", Value: b6.StringExpression("no")})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, c, abc, lighterman); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	overlay.AddTag(lighterman.FeatureID(), b6.Tag{Key: "#reachable", Value: b6.StringExpression("yes")})

	areas := b6.AllAreas(b6.FindAreas(b6.Tagged{Key: "#reachable", Value: b6.StringExpression("yes")}, overlay))
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
	ab := osmPath(807924986, []Feature{a, b})
	// Part of C6
	c6 := osmSimpleRelation(10341051, 807924986)
	c6.AddTag(b6.Tag{Key: "type", Value: b6.StringExpression("route")})
	c6.AddTag(b6.Tag{Key: "network", Value: b6.StringExpression("lcn")})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, a, b, ab, c6); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	overlay.AddTag(c6.FeatureID(), b6.Tag{Key: "route", Value: b6.StringExpression("bicycle")})
	overlay.AddTag(c6.FeatureID(), b6.Tag{Key: "network", Value: b6.StringExpression("rcn")})

	relation := b6.FindRelationByID(c6.RelationID, overlay)

	expected := []b6.Tag{
		{Key: "type", Value: b6.StringExpression("route")},
		{Key: "route", Value: b6.StringExpression("bicycle")},
		{Key: "network", Value: b6.StringExpression("rcn")},
	}

	if err := validateTags(relation, expected); err != nil {
		t.Error(err)
	}
}

func TestChangeTagsOnExistingCollection(t *testing.T) {
	c := simpleCollection(b6.MakeCollectionID(b6.Namespace("test"), 1), "sappho", "fragments")
	f13 := b6.Tag{Key: "13", Value: b6.StringExpression("Of all the stars")}
	f31 := b6.Tag{Key: "31", Value: b6.StringExpression("That man seems to me to be equal to the gods, sitting opposite of you..")}
	c.AddTag(f13)
	c.AddTag(f31)

	base := NewBasicMutableWorld()
	if err := addFeatures(base, c); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	overlay.AddTag(c.FeatureID(), b6.Tag{Key: "13", Value: b6.StringExpression("Of all the stars, the loveliest")})

	collection := b6.FindCollectionByID(c.CollectionID, overlay)

	expected := []b6.Tag{
		{Key: "31", Value: b6.StringExpression("That man seems to me to be equal to the gods, sitting opposite of you..")},
		{Key: "13", Value: b6.StringExpression("Of all the stars, the loveliest")},
	}

	if err := validateTags(collection, expected); err != nil {
		t.Error(err)
	}
}

func TestReturnModifiedTagsFromSearch(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Caravan")})
	caravan.AddTag(b6.Tag{Key: "#amenity", Value: b6.StringExpression("restaurant")})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "wheelchair", Value: b6.StringExpression("yes")})

	points := overlay.FindFeatures(b6.Typed{b6.FeatureTypePoint, b6.Tagged{Key: "#amenity", Value: b6.StringExpression("restaurant")}})
	if !points.Next() {
		t.Fatal("Expected to find 1 point")
	}

	point := points.Feature()
	if points.Next() {
		t.Fatal("Expected to find 1 point only")
	}

	expected := []b6.Tag{
		{Key: "name", Value: b6.StringExpression("Caravan")},
		{Key: "#amenity", Value: b6.StringExpression("restaurant")},
		{Key: "wheelchair", Value: b6.StringExpression("yes")},
		{Key: b6.PointTag, Value: b6.PointExpression(s2.LatLngFromDegrees(51.5357237, -0.1253052))},
	}

	if err := validateTags(point, expected); err != nil {
		t.Error(err)
	}
}

func TestSettingTheSameTagMultipleTimesChangesTheValue(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Caravan")})
	caravan.AddTag(b6.Tag{Key: "#amenity", Value: b6.StringExpression("restaurant")})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	for i := 0; i < 4; i++ {
		overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "100m", Value: b6.StringExpression("yes")})
		overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "#200m", Value: b6.StringExpression("yes")})
	}

	expected := []b6.Tag{
		{Key: "name", Value: b6.StringExpression("Caravan")},
		{Key: "#amenity", Value: b6.StringExpression("restaurant")},
		{Key: "100m", Value: b6.StringExpression("yes")},
		{Key: "#200m", Value: b6.StringExpression("yes")},
		{Key: b6.PointTag, Value: b6.PointExpression(s2.LatLngFromDegrees(51.5357237, -0.1253052))},
	}

	if err := validateTags(overlay.FindFeatureByID(caravan.FeatureID()), expected); err != nil {
		t.Error(err)
	}
}

func TestRemovedTag(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Caravan")})
	caravan.AddTag(b6.Tag{Key: "#amenity", Value: b6.StringExpression("restaurant")})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	for i := 0; i < 4; i++ {
		overlay.RemoveTag(caravan.FeatureID(), "#amenity")
		overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "#shop", Value: b6.StringExpression("supermarket")})
	}

	expected := []b6.Tag{
		{Key: "name", Value: b6.StringExpression("Caravan")},
		{Key: "#shop", Value: b6.StringExpression("supermarket")},
		{Key: b6.PointTag, Value: b6.PointExpression(s2.LatLngFromDegrees(51.5357237, -0.1253052))},
	}

	if err := validateTags(overlay.FindFeatureByID(caravan.FeatureID()), expected); err != nil {
		t.Error(err)
	}
}

func TestConnectivityUnchangedFollowingTagModification(t *testing.T) {
	w, err := NewWorldFromPBFFile(test.Data(test.CamdenPBF), &BuildOptions{Cores: 2})
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
		m.AddTag(a.FeatureID(), b6.Tag{Key: "#reachable", Value: b6.StringExpression("yes")})
	}

	after := b6.AllAreas(m.FindAreasByPoint(entrance))
	if len(before) != len(after) {
		t.Errorf("Expected %d areas, found %d", len(before), len(after))
	}
}

func TestWatchModifiedTags(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Caravan")})
	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableTagsOverlayWorld(base)
	c, cancel := overlay.Watch()

	overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "wheelchair", Value: b6.StringExpression("yes")})
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
	overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "wheelchair", Value: b6.StringExpression("no")})
	expected := ModifiedTag{ID: caravan.FeatureID(), Tag: b6.Tag{Key: "wheelchair", Value: b6.StringExpression("yes")}, Deleted: false}
	if m != expected {
		t.Errorf("Expected %v, found %v", expected, m)
	}
}

func TestPropagateModifiedTags(t *testing.T) {
	base := NewBasicMutableWorld()

	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Caravan")})
	if err := base.AddFeature(caravan); err != nil {
		t.Fatal(err)
	}

	yumchaa := osmPoint(3790640853, 51.5355955, -0.1250640)
	if err := base.AddFeature(yumchaa); err != nil {
		t.Fatal(err)
	}

	granary := osmPath(222021576, []Feature{caravan, yumchaa})
	if err := base.AddFeature(granary); err != nil {
		t.Fatal(err)
	}

	w := NewMutableTagsOverlayWorld(base)
	w.AddTag(caravan.FeatureID(), b6.Tag{Key: "amenity", Value: b6.StringExpression("restaurant")})

	found := w.FindFeatureByID(caravan.FeatureID())
	if found == nil {
		t.Fatal("Failed to find feature")
	}
	if found.Get("amenity").Value.String() != "restaurant" {
		t.Error("Failed to find expected tag value")
	}

	path := w.FindFeatureByID(granary.FeatureID()).(b6.PhysicalFeature)
	if path == nil {
		t.Fatal("Failed to find feature")
	}
	for i := 0; i < path.GeometryLen(); i++ {
		if found := w.FindFeatureByID(path.Reference(i).Source()); found == nil || found.FeatureID() == caravan.FeatureID() && found.Get("amenity").Value.String() != "restaurant" {
			t.Error("Failed to find expected tag value")
		}
	}

	ss := w.Traverse(caravan.FeatureID())
	for ss.Next() {
		segment := ss.Segment()
		for i := 0; i < segment.Feature.GeometryLen(); i++ {
			if found := segment.Feature.Reference(i).Source(); found == caravan.FeatureID() {
				if w.FindFeatureByID(found).Get("amenity").Value.String() != "restaurant" {
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

	for _, p := range []Feature{caravan, yumchaa} {
		if err := base.AddFeature(p); err != nil {
			t.Fatal(err)
		}
	}

	path := &GenericFeature{ID: FromOSMWayID(222021576)}
	path.ModifyOrAddTag(b6.Tag{b6.PathTag, b6.Values([]b6.Value{caravan.FeatureID(), b6.PointExpression(s2.LatLngFromDegrees(51.535490, -0.125167)), yumchaa.FeatureID()})})
	if err := base.AddFeature(path); err != nil {
		t.Fatal(err)
	}

	w := NewMutableTagsOverlayWorld(base)
	w.AddTag(path.FeatureID(), b6.Tag{Key: "#highway", Value: b6.StringExpression("path")})

	path2 := w.FindFeatureByID(path.FeatureID()).(b6.PhysicalFeature)
	if path2 == nil {
		t.Fatal("Failed to find feature")
	}
	if id := path2.Reference(0).Source(); id != caravan.FeatureID() {
		t.Errorf("Expected ID %s, found %s", caravan.FeatureID(), id)
	}
	if w.FindFeatureByID(path2.Reference(1).Source()) != nil {
		t.Error("Expected nil feature for lat, lng")
	}
}

func TestModifiedTagsFeatureFromArea(t *testing.T) {
	path := &GenericFeature{ID: FromOSMWayID(265714033)}
	path.ModifyOrAddTag(b6.Tag{b6.PathTag, b6.Values([]b6.Value{
		b6.PointExpression(s2.LatLngFromDegrees(51.5369431, -0.1231868)),
		b6.PointExpression(s2.LatLngFromDegrees(51.5365692, -0.1230608)),
		b6.PointExpression(s2.LatLngFromDegrees(51.5365536, -0.1229421)),
		b6.PointExpression(s2.LatLngFromDegrees(51.5367378, -0.1229110)),
		b6.PointExpression(s2.LatLngFromDegrees(51.5369431, -0.1231868)),
	})})

	area := NewAreaFeature(1)
	area.AreaID = AreaIDFromOSMWayID(265714033)
	area.SetPathIDs(0, []b6.FeatureID{path.FeatureID()})
	area.AddTag(b6.Tag{Key: "#leisure", Value: b6.StringExpression("playground")})

	base := NewBasicMutableWorld()
	if err := base.AddFeature(path); err != nil {
		t.Fatalf("Failed to add path: %s", err)
	}
	if err := base.AddFeature(area); err != nil {
		t.Fatalf("Failed to add area: %s", err)
	}

	overlay := NewMutableOverlayWorld(base)
	overlay.AddTag(path.FeatureID(), b6.Tag{Key: "barrier", Value: b6.StringExpression("fence")})

	found := b6.FindAreaByID(area.AreaID, overlay)
	if found == nil {
		t.Fatal("Expected to find feature")
	}

	if b := found.Feature(0)[0].Get("barrier"); !b.IsValid() || b.Value.String() != "fence" {
		t.Error("Expected to find added tag value")
	}
}

func TestSnapshot(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Caravan")})
	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(base)
	overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "wheelchair", Value: b6.StringExpression("yes")})
	snapshot := overlay.Snapshot()
	overlay.AddTag(caravan.FeatureID(), b6.Tag{Key: "wheelchair", Value: b6.StringExpression("no")})

	if wheelchair := overlay.FindFeatureByID(caravan.FeatureID()).Get("wheelchair"); wheelchair.Value.String() != "no" {
		t.Errorf("Expected wheelchair=no in overlay, found %s", wheelchair.Value)
	}

	if wheelchair := snapshot.FindFeatureByID(caravan.FeatureID()).Get("wheelchair"); wheelchair.Value.String() != "yes" {
		t.Errorf("Expected wheelchair=yes in snapshot, found %s", wheelchair.Value)
	}
}

func TestChangeTagsOnNonExistantPoint(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Caravan")})

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
	lighterman.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("The Lighterman")})
	lighterman.AddTag(b6.Tag{Key: "wheelchair", Value: b6.StringExpression("no")})

	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Caravan")})
	caravan.AddTag(b6.Tag{Key: "amenity", Value: b6.StringExpression("restaurant")})

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
		overlay.AddFeature(caravan)
	}
}

func TestEachFeatureWithAPathDependingOnTheBaseWorld(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Caravan")})

	dishoom := osmPoint(3501612811, 51.536454, -0.126826)
	dishoom.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Dishoom")})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan, dishoom); err != nil {
		t.Fatal(err)
	}

	footway := osmPath(558345071, []Feature{caravan, dishoom})
	footway.AddTag(b6.Tag{Key: "highway", Value: b6.StringExpression("footway")})
	m := NewMutableOverlayWorld(base)
	if err := m.AddFeature(footway); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	found := false
	each := func(f b6.Feature, goroutine int) error {
		if f.FeatureID() == footway.FeatureID() {
			found = true
			path := f.(b6.PhysicalFeature)
			for i := 0; i < path.GeometryLen(); i++ {
				// Ensure features are wrapped correctly to be able to
				// lookup points in the base world.
				path.PointAt(i)
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
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Caravan")})
	caravan.AddTag(b6.Tag{Key: "cuisine", Value: b6.StringExpression("coffee_shop")})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, caravan); err != nil {
		t.Fatal(err)
	}

	m := NewMutableOverlayWorld(base)
	m.AddTag(caravan.FeatureID(), b6.Tag{Key: "wheelchair", Value: b6.StringExpression("yes")})
	m.RemoveTag(caravan.FeatureID(), "cuisine")

	found := false
	each := func(f b6.Feature, goroutine int) error {
		if f.FeatureID() == caravan.FeatureID() {
			found = true
			if wheelchair := f.Get("wheelchair"); wheelchair.Value.String() != "yes" {
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
	lighterman.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("The Lighterman")})
	lighterman.AddTag(b6.Tag{Key: "wheelchair", Value: b6.StringExpression("no")})

	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Caravan")})
	caravan.AddTag(b6.Tag{Key: "amenity", Value: b6.StringExpression("restaurant")})

	base := NewBasicMutableWorld()
	if err := addFeatures(base, lighterman); err != nil {
		t.Fatal(err)
	}

	lower := NewMutableOverlayWorld(base)
	upper := NewMutableOverlayWorld(lower)
	i := lower.FindFeatures(b6.All{})
	for i.Next() {
		upper.AddFeature(caravan)
	}

	if err := upper.MergeInto(lower); err != nil {
		t.Errorf("Expected no error from merge, found: %s", err)
	}

	if point := lower.FindFeatureByID(caravan.FeatureID()); point == nil {
		t.Error("Expected to find caravan in the lower world.")
	}
}

func TestChangeSearchableTagOnFeatureInBaseWorld(t *testing.T) {
	lighterman := osmPoint(427900370, 51.5353986, -0.1243711)
	lighterman.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("The Lighterman")})
	lighterman.AddTag(b6.Tag{Key: "#amenity", Value: b6.StringExpression("restaurant")})

	// Recreate a bug that occured where modified features where incorrectly
	// returning from a search when embeded between other matching features
	// that weren't modified.
	id := lighterman.FeatureID()
	id.Value--
	before := lighterman.Clone()
	before.SetFeatureID(id)
	id.Value += 2
	after := lighterman.Clone()
	after.SetFeatureID(id)

	w := NewBasicMutableWorld()
	if err := addFeatures(w, lighterman, before, after); err != nil {
		t.Fatal(err)
	}

	overlay := NewMutableOverlayWorld(w)

	q := b6.Tagged{Key: "#amenity", Value: b6.StringExpression("restaurant")}
	found := b6.AllFeatures(overlay.FindFeatures(q))
	if len(found) != 3 {
		t.Fatalf("Expected to find 3 features, found %d", len(found))
	}

	if err := overlay.AddTag(lighterman.FeatureID(), b6.Tag{Key: "#amenity", Value: b6.StringExpression("pub")}); err != nil {
		t.Fatalf("Failed to add tag: %s", err)
	}

	found = b6.AllFeatures(overlay.FindFeatures(q))
	if len(found) != 2 {
		t.Fatalf("Expected to find 2 features, found %d", len(found))
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
