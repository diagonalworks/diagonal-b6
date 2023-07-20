package functions

import (
	"math"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"

	"github.com/golang/geo/s2"
)

func TestAllTags(t *testing.T) {
	w := camden.BuildCamdenForTests(t)
	if w == nil {
		return
	}

	vermuteria := b6.FindPointByID(camden.VermuteriaID, w)
	if vermuteria == nil {
		t.Errorf("Failed to find expected test point")
		return
	}

	all, err := allTags(&api.Context{World: w}, vermuteria)
	if err != nil {
		t.Errorf("Expected no error, found %s", err)
		return
	}

	filled := make([]b6.Tag, 0)
	err = api.FillSliceFromValues(all, &filled)
	if err != nil {
		t.Errorf("Expected no error, found %s", err)
		return
	}

	if len(filled) < 2 {
		t.Errorf("Expected at least two tags")
	}
	found := false
	for _, tag := range filled {
		if tag.Key == "#amenity" {
			if tag.Value != "cafe" {
				t.Errorf("Expected #amenity=cafe, found %+v", tag)
			}
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find #amenity tag")
	}
}

func TestFindAreasContainingPoints(t *testing.T) {
	w := camden.BuildCamdenForTests(t)
	if w == nil {
		return
	}
	m := ingest.NewMutableOverlayWorld(w)

	vermuteria := b6.FindPointByID(camden.VermuteriaID, m)
	if vermuteria == nil {
		t.Errorf("Failed to find expected test point")
		return
	}

	points := &api.ArrayPointFeatureCollection{Features: []b6.PointFeature{vermuteria}}
	context := api.Context{
		World: m,
	}
	found, err := findAreasContainingPoints(&context, points, b6.Keyed{"#shop"})
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
	}

	areas := make(map[b6.AreaID]b6.AreaFeature)
	if err := api.FillMap(found, areas); err != nil {
		t.Errorf("Expected no error, found %s", err)
		return
	}

	if _, ok := areas[camden.CoalDropsYardEnclosureID]; !ok {
		t.Errorf("Expected points to be contained within %s", camden.CoalDropsYardEnclosureID.FeatureID())
	}
}

func TestPoints(t *testing.T) {
	granarySquare := []s2.Point{
		s2.PointFromLatLng(s2.LatLngFromDegrees(51.5357019, -0.1260475)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(51.5355674, -0.1261001)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(51.5350372, -0.1255004)),
	}
	lighterman := []s2.Point{
		s2.PointFromLatLng(s2.LatLngFromDegrees(51.5354124, -0.1243817)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(51.5353117, -0.1244943)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(51.5353594, -0.1242476)),
	}

	polygons := []*s2.Polygon{
		s2.PolygonFromLoops([]*s2.Loop{s2.LoopFromPoints(granarySquare)}),
		s2.PolygonFromLoops([]*s2.Loop{s2.LoopFromPoints(lighterman)}),
	}
	var c api.Context
	ps, err := points(&c, b6.AreaFromS2Polygons(polygons))
	if err != nil {
		t.Errorf("Expected no error, found %s", err)
		return
	}

	points := make(map[int]b6.Point)
	if err := api.FillMap(ps, points); err != nil {
		t.Errorf("Expected no error, found %s", err)
		return
	}

	if len(points) != len(granarySquare)+len(lighterman) {
		t.Errorf("Expected %d points, found %d", len(granarySquare)+len(lighterman), len(points))
		return
	}

	center := s2.PointFromLatLng(s2.LatLngFromDegrees(51.53541, -0.12530))
	for _, v := range points {
		if d := b6.AngleToMeters(v.Point().Distance(center)); d > 100.0 {
			t.Errorf("Point too far away: expected <= 100.0m, found %fm", d)
		}
	}
}

func TestSamplePointsAlongPaths(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	context := &api.Context{
		World: granarySquare,
	}

	paths, err := findPathFeatures(context, b6.Keyed{"#highway"})
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	sampled, err := samplePointsAlongPaths(context, paths, 20.0)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	points := make(map[interface{}]b6.Point)
	if err := api.FillMap(sampled, points); err != nil {
		t.Errorf("Expected no error, found %s", err)
		return
	}

	if len(points) < 300 || len(points) > 350 {
		t.Errorf("Number of sampled points outside expected bounds: %d", len(points))
	}

	center := s2.PointFromLatLng(s2.LatLngFromDegrees(51.53539, -0.12537))
	for _, v := range points {
		if v.Point().Distance(center) > b6.MetersToAngle(500) {
			t.Errorf("Point too far away from the center of the test data area")
		}

	}
}

func TestSamplePointsAlongPathsIsConsistentAcrossRuns(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	context := &api.Context{
		World: granarySquare,
	}

	paths, err := findPathFeatures(context, b6.Keyed{"#highway"})
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	runs := make([][]s2.Point, 4)
	for run := range runs {
		points, err := samplePointsAlongPaths(context, paths, 20.0)
		if err != nil {
			t.Errorf("Expected no error on run %d, found: %s", run, err)
			return
		}
		runs[run] = make([]s2.Point, 0, 2)
		i := points.Begin()
		for {
			ok, err := i.Next()
			if err != nil {
				t.Errorf("Expected no error, found: %s", err)
				return
			}
			if !ok {
				break
			}
			runs[run] = append(runs[run], i.Value().(b6.Point).Point())
		}
	}

	for run := 1; run < len(runs); run++ {
		if len(runs[run]) != len(runs[0]) {
			t.Errorf("Run %d length %d, expected %d", run, len(runs[run]), len(runs[0]))
			return
		}
		for i := range runs[run] {
			if runs[run][i] != runs[0][i] {
				t.Errorf("Sample points results were not consistent between runs")
				return
			}
		}
	}
}

func TestJoin(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	context := &api.Context{
		World: granarySquare,
	}

	// Two connected paths
	a := b6.FindPathByID(ingest.FromOSMWayID(377974549), granarySquare)
	b := b6.FindPathByID(ingest.FromOSMWayID(834245629), granarySquare)

	if a == nil || b == nil {
		t.Errorf("Failed to find expected paths")
		return
	}

	joined, err := join(context, a, b)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	if d := math.Abs((joined.Polyline().Length() / (a.Polyline().Length() + b.Polyline().Length())).Radians() - 1.0); d > 0.0001 {
		t.Errorf("Expected delta to be small, found: %f", d)
	}
}

func TestOrderedJoin(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	path := b6.FindPathByID(ingest.FromOSMWayID(377974549), granarySquare)
	midVertex := path.Len() / 2

	aPoints := make([]s2.Point, 0, path.Len()/2)
	for i := midVertex; i >= 0; i-- {
		aPoints = append(aPoints, path.Point(i))
	}
	a := b6.PathFromS2Points(aPoints)

	bPoints := make([]s2.Point, 0, path.Len()/2)
	for i := midVertex; i < path.Len(); i++ {
		bPoints = append(bPoints, path.Point(i))
	}
	b := b6.PathFromS2Points(bPoints)

	joined, err := orderedJoin(&api.Context{}, a, b)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	midpoint, _ := interpolate(&api.Context{}, joined, 0.5)
	expected, _ := interpolate(&api.Context{}, path, 0.5)

	if midpoint.Point().Distance(expected.Point()) > 0.000001 {
		t.Errorf("Midpoint of joined paths too far from expected point")
	}
}

func TestInterpolate(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	context := &api.Context{
		World: granarySquare,
	}
	path := b6.FindPathByID(ingest.FromOSMWayID(377974549), granarySquare)
	if path == nil {
		t.Errorf("Failed to find expected path")
	}

	interpolated, err := interpolate(context, path, 0.5)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	expected := s2.LatLngFromDegrees(51.5361869, -0.1258445)
	if d := b6.AngleToMeters(s2.PointFromLatLng(expected).Distance(interpolated.Point())); d > 0.1 {
		t.Errorf("Interpolated point not close to expected location: %fm", d)
	}
}

func TestOrderedJoinPathsWithNoSharedPoint(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	path := b6.FindPathByID(ingest.FromOSMWayID(377974549), granarySquare)
	midVertex := path.Len() / 2

	aPoints := make([]s2.Point, 0, path.Len()/2)
	for i := midVertex; i >= 0; i-- {
		aPoints = append(aPoints, path.Point(i))
	}
	a := b6.PathFromS2Points(aPoints)

	bPoints := make([]s2.Point, 0, path.Len()/2)
	for i := midVertex + 1; i < path.Len(); i++ { // Don't add the shared point
		bPoints = append(bPoints, path.Point(i))
	}
	b := b6.PathFromS2Points(bPoints)

	_, err := orderedJoin(&api.Context{}, a, b)
	if err == nil {
		t.Errorf("Expected an error, found nil")
	}
}
