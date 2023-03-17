package functions

import (
	"diagonal.works/b6"
	"diagonal.works/b6/api"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

func ll(lat float64, lng float64, context *api.Context) (b6.Point, error) {
	return b6.PointFromLatLng(s2.LatLngFromDegrees(lat, lng)), nil
}

func distanceMeters(a b6.Point, b b6.Point, context *api.Context) (float64, error) {
	return b6.AngleToMeters(a.Point().Distance(b.Point())), nil
}

func distanceToPointMeters(path b6.Path, point b6.Point, context *api.Context) (float64, error) {
	polyline := *path.Polyline()
	projection, vertex := polyline.Project(point.Point())
	distance := polyline[vertex-1].Distance(projection)
	if vertex > 1 {
		p := polyline[0:vertex]
		distance += p.Length()
	}
	return b6.AngleToMeters(distance), nil
}

func centroid(geometry b6.Geometry, context *api.Context) (b6.Point, error) {
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

func interpolate(p b6.Path, fraction float64, context *api.Context) (b6.Point, error) {
	polyline := p.Polyline()
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

func areaArea(area b6.Area, context *api.Context) (float64, error) {
	m2 := 0.0
	for i := 0; i < area.Len(); i++ {
		polygon := area.Polygon(i)
		if validatePolygon(polygon) {
			m2 += b6.AreaToMeters2(polygon.Area())
		}
	}
	return m2, nil
}

func rectanglePolygon(p0 b6.Point, p1 b6.Point, context *api.Context) (b6.Area, error) {
	r := s2.EmptyRect().AddPoint(s2.LatLngFromPoint(p0.Point()))
	r = r.AddPoint(s2.LatLngFromPoint(p1.Point()))
	points := make([]s2.Point, 4)
	for i := range points {
		points[i] = s2.PointFromLatLng(r.Vertex(i))
	}
	return b6.AreaFromS2Loop(s2.LoopFromPoints(points)), nil
}

func capPolygon(center b6.Point, radius float64, context *api.Context) (b6.Area, error) {
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

func snapAreaEdges(g b6.Area, query b6.Query, threshold float64, context *api.Context) (b6.Area, error) {
	thresholdAngle := b6.MetersToAngle(threshold)
	snapped := make([]*s2.Polygon, 0, g.Len())
	for i := 0; i < g.Len(); i++ {
		polygon := g.Polygon(i)
		cap := polygon.CapBound()
		buffered := s2.CapFromCenterAngle(cap.Center(), cap.Radius()+b6.MetersToAngle(threshold))
		polylines := make([]*s2.Polyline, 0, 4)
		paths := b6.FindPaths(b6.Intersection{query, b6.MightIntersect{buffered}}, context.World)
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
