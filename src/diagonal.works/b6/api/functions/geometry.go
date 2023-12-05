package functions

import (
	"diagonal.works/b6"
	"diagonal.works/b6/api"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

// Return a point at the given latitude and longitude, specified in degrees.
func ll(context *api.Context, lat float64, lng float64) (b6.Point, error) {
	return b6.PointFromLatLng(s2.LatLngFromDegrees(lat, lng)), nil
}

// Return a single area containing all areas from the given collection.
// If areas in the collection overlap, loops within the returned area
// will overlap, which will likely cause undefined behaviour in many
// functions.
func collectAreas(context *api.Context, areas b6.Collection[any, b6.Area]) (b6.Area, error) {
	i := areas.Begin()
	ps := make([]*s2.Polygon, 0)
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		if area, ok := i.Value().(b6.Area); ok {
			ps = append(ps, area.MultiPolygon()...)
		}
	}
	return b6.AreaFromS2Polygons(ps), nil
}

// Return the distance in meters between the given points.
func distanceMeters(context *api.Context, a b6.Point, b b6.Point) (float64, error) {
	return b6.AngleToMeters(a.Point().Distance(b.Point())), nil
}

// Return the distance in meters between the given path, and the project of the give point onto it.
func distanceToPointMeters(context *api.Context, path b6.Path, point b6.Point) (float64, error) {
	polyline := *path.Polyline()
	projection, vertex := polyline.Project(point.Point())
	distance := polyline[vertex-1].Distance(projection)
	if vertex > 1 {
		p := polyline[0:vertex]
		distance += p.Length()
	}
	return b6.AngleToMeters(distance), nil
}

// Return the centroid of the given geometry.
// For multipolygons, we return the centroid of the convex hull formed from
// the points of those polygons.
func centroid(context *api.Context, geometry b6.Geometry) (b6.Point, error) {
	switch g := geometry.(type) {
	case b6.Point:
		return g, nil
	case b6.Path:
		return b6.PointFromS2Point(g.Polyline().Centroid()), nil
	case b6.Area:
		query := s2.NewConvexHullQuery()
		for i := 0; i < g.Len(); i++ {
			polygon := g.Polygon(i)
			query.AddPolygon(polygon)
		}
		return b6.PointFromS2Point(query.ConvexHull().Centroid()), nil
	}
	return b6.PointFromS2Point(s2.Point{}), nil
}

// Return the point at the given fraction along the given path.
func interpolate(context *api.Context, path b6.Path, fraction float64) (b6.Point, error) {
	polyline := path.Polyline()
	point, _ := polyline.Interpolate(fraction)
	return b6.PointFromS2Point(point), nil
}

func validatePolygon(polygon *s2.Polygon) bool {
	if polygon.NumLoops() == 0 {
		return false
	}
	for i := 0; i < polygon.NumLoops(); i++ {
		if polygon.Loop(i).NumVertices() < 3 {
			return false
		}
	}
	return true
}

// Return the area of the given polygon in m².
func areaArea(context *api.Context, area b6.Area) (float64, error) {
	m2 := 0.0
	for i := 0; i < area.Len(); i++ {
		polygon := area.Polygon(i)
		if validatePolygon(polygon) {
			m2 += b6.AreaToMeters2(polygon.Area())
		}
	}
	return m2, nil
}

// Return a rectangle polygon with the given top left and bottom right points.
func rectanglePolygon(context *api.Context, a b6.Point, b b6.Point) (b6.Area, error) {
	r := s2.EmptyRect().AddPoint(s2.LatLngFromPoint(a.Point()))
	r = r.AddPoint(s2.LatLngFromPoint(b.Point()))
	points := make([]s2.Point, 4)
	for i := range points {
		points[i] = s2.PointFromLatLng(r.Vertex(i))
	}
	return b6.AreaFromS2Loop(s2.LoopFromPoints(points)), nil
}

// Return a polygon approximating a spherical cap with the given center and radius in meters.
func capPolygon(context *api.Context, center b6.Point, radius float64) (b6.Area, error) {
	return b6.AreaFromS2Loop(s2.RegularLoop(center.Point(), b6.MetersToAngle(radius), 128)), nil
}

func filterShortEdges(edges []s2.Edge, threshold s1.Angle) []s2.Edge {
	filtered := make([]s2.Edge, 0, len(edges))
	for _, edge := range edges {
		if edge.V0.Distance(edge.V1) > threshold {
			filtered = append(filtered, edge)
		}
	}
	return filtered
}

func projectEdgesOntoPolylines(loop *s2.Loop, polylines []*s2.Polyline, threshold s1.Angle) []s2.Edge {
	edges := make([]s2.Edge, loop.NumVertices())
	for i := 0; i < loop.NumVertices(); i++ {
		edges[i].V0 = loop.Vertex(i)
		edges[i].V1 = loop.Vertex(i + 1)
		d := s1.InfAngle()
		for _, polyline := range polylines {
			v0, _ := polyline.Project(loop.Vertex(i))
			d0 := v0.Distance(loop.Vertex(i))
			v1, _ := polyline.Project(loop.Vertex(i + 1))
			d1 := v1.Distance(loop.Vertex(i + 1))
			if d1 > d0 {
				d0, d1 = d1, d0
			}
			if d0 < d && d0 < threshold {
				d = d0
				edges[i].V0 = v0
				edges[i].V1 = v1
			}
		}
	}
	return edges
}

// Return an area formed by projecting the edges of the given polygon onto the paths present in the world matching the given query.
// Paths beyond the given threshold in meters are ignored.
func snapAreaEdges(context *api.Context, area b6.Area, query b6.Query, threshold float64) (b6.Area, error) {
	thresholdAngle := b6.MetersToAngle(threshold)
	snapped := make([]*s2.Polygon, 0, area.Len())
	for i := 0; i < area.Len(); i++ {
		polygon := area.Polygon(i)
		cap := polygon.CapBound()
		buffered := s2.CapFromCenterAngle(cap.Center(), cap.Radius()+b6.MetersToAngle(threshold))
		polylines := make([]*s2.Polyline, 0, 4)
		paths := b6.FindPaths(b6.Intersection{query, b6.MightIntersect{Region: buffered}}, context.World)
		for paths.Next() {
			polylines = append(polylines, paths.Feature().Polyline())
		}
		loops := make([]*s2.Loop, 0, polygon.NumLoops())
		for j := 0; j < polygon.NumLoops(); j++ {
			joinThreshold := b6.MetersToAngle(0.1)
			edges := filterShortEdges(projectEdgesOntoPolylines(polygon.Loop(j), polylines, thresholdAngle), joinThreshold)
			points := make([]s2.Point, 0, len(edges))
			for k := range edges {
				next := (k + 1) % len(edges)
				points = append(points, edges[k].V0)
				if edges[k].V1.Distance(edges[next].V0) > joinThreshold {
					points = append(points, edges[k].V1)
					a := s2.InterpolateAtDistance(-thresholdAngle, edges[k].V0, edges[k].V1)
					b := s2.InterpolateAtDistance(-thresholdAngle, edges[k].V1, edges[k].V0)
					c := s2.InterpolateAtDistance(-thresholdAngle, edges[next].V0, edges[next].V1)
					d := s2.InterpolateAtDistance(-thresholdAngle, edges[next].V1, edges[next].V0)
					if s2.CrossingSign(a, b, c, d) == s2.Cross {
						intersection := s2.Intersection(a, b, c, d)
						if intersection.Distance(edges[k].V1) > joinThreshold && intersection.Distance(edges[next].V0) > joinThreshold {
							points = append(points, intersection)
						}
					}
				}
			}
			if l := s2.LoopFromPoints(points); l.Validate() == nil {
				loops = append(loops, l)
			} else {
				loops = append(loops, polygon.Loop(j))
			}
			snapped = append(snapped, s2.PolygonFromLoops(loops))
		}
	}
	return b6.AreaFromS2Polygons(snapped), nil
}

// Return the convex hull of the given geometries.
func convexHull(context *api.Context, c b6.Collection[any, b6.Geometry]) (b6.Area, error) {
	query := s2.NewConvexHullQuery()
	i := c.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		switch g := i.Value().(type) {
		case b6.Point:
			query.AddPoint(g.Point())
		case b6.Path:
			for i := 0; i < g.Len(); i++ {
				query.AddPoint(g.Point(i))
			}
		case b6.Area:
			for i := 0; i < g.Len(); i++ {
				query.AddPolygon(g.Polygon(i))
			}
		}
	}
	return b6.AreaFromS2Loop(query.ConvexHull()), nil
}
