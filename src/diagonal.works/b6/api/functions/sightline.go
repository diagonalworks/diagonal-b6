package functions

import (
	"fmt"
	"log"
	"math"
	"sort"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/geometry"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

const verbose = false
const verboseGeoJSON = false

func sightline(context *api.Context, from b6.Geometry, radius float64) (b6.Area, error) {
	if centroid, ok := b6.Centroid(from); ok {
		return b6.AreaFromS2Polygon(Sightline(centroid, b6.MetersToAngle(radius), context.World)), nil
	}
	return nil, fmt.Errorf("invalid starting point for sightline")
}

func Sightline(center s2.Point, radius s1.Angle, w b6.World) *s2.Polygon {
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Sprintf("panic in Sightline %s: %v", s2.LatLngFromPoint(center).String(), r))
		}
	}()
	return SightlineUsingPolarCoordinates2(center, radius, w)
}

func loopIntersects(a s2.Point, b s2.Point, loop *s2.Loop, exceptEdge int) bool {
	for i := 0; i < loop.NumVertices(); i++ {
		if i != exceptEdge {
			c := loop.Vertex(i)
			d := loop.Vertex((i + 1) % loop.NumVertices())
			if s2.CrossingSign(a, b, c, d) == s2.Cross {
				return true
			}
		}
	}
	return false
}

// Find the point at which the line ab intersects loop. In the case of multiple
// intersections, return the intersection that's closest to point a.
// Returns:
// - the intersection point
// - the index within the loop of the first vertex on which the intersection point lies
// - a boolean, true if there is an intersection
func loopIntersection(a s2.Point, b s2.Point, loop *s2.Loop) (s2.Point, int, bool) {
	angle := s1.InfAngle()
	intersects := false
	closest := s2.Point{}
	edge := 0
	for i := 0; i < loop.NumVertices(); i++ {
		c := loop.Vertex(i)
		d := loop.Vertex((i + 1) % loop.NumVertices())
		if s2.CrossingSign(a, b, c, d) == s2.Cross {
			intersection := s2.Intersection(a, b, c, d)
			if a.Distance(intersection) < angle {
				intersects = true
				closest = intersection
				edge = i
			}
		}
	}
	return closest, edge, intersects
}

type byAngle struct {
	center    s2.Point
	reference s2.Point
	edges     [][2]s2.Point
	midpoints []s2.Point
}

func (b byAngle) Len() int { return len(b.edges) }
func (b byAngle) Swap(i, j int) {
	b.edges[i], b.edges[j] = b.edges[j], b.edges[i]
	b.midpoints[i], b.midpoints[j] = b.midpoints[j], b.midpoints[i]
}
func (b byAngle) Less(i, j int) bool {
	return s2.TurnAngle(b.reference, b.center, b.midpoints[i]) < s2.TurnAngle(b.reference, b.center, b.midpoints[j])
}

// Compute the area visible from a given point by computing the occlusion
// polygons formed by each building edge, and subtracting these polygons
// from each building and sightline boundary edge. We then sort the
// remaining edges by the angle with resprect to an arbitrary reference point
// on the sightline, and linking them up in a counter-clockwise order.
// Can return an empty polygon if the sightline would be made entirely
// of edges below a tolerance (approximately 1cm), or if the center point
// lies within a building.
func SightlineUsingPolarCoordinates(center s2.Point, radius s1.Angle, w b6.World) *s2.Polygon {
	sightline := s2.RegularLoop(center, radius, 128)
	boundary := s2.RegularLoop(center, radius*1.1, 128)
	index := s2.NewShapeIndex()
	index.Add(boundary)
	containsPoint := s2.NewContainsPointQuery(index, s2.VertexModelOpen)
	tolerance := b6.MetersToAngle(0.01)
	cap := s2.CapFromCenterAngle(center, radius)
	features := b6.FindAreas(b6.Intersection{b6.NewIntersectsCap(cap), b6.Keyed{Key: "#building"}}, w)
	edges := make([][2]s2.Point, 0, 8)
	occlusions := make([]*s2.Loop, 0, 2)
	for features.Next() {
		area := features.Feature()
		for i := 0; i < area.Len(); i++ {
			loop := area.Polygon(i).Loop(0)
			for j := 0; j < loop.NumVertices(); j++ {
				e := [2]s2.Point{loop.Vertex(j), loop.Vertex((j + 1) % loop.NumVertices())}
				midpoint := s2.Interpolate(0.5, e[0], e[1])
				// Only consider edges at the "front" of the feature, with respect to the center point
				if !loopIntersects(center, midpoint, loop, j) {
					edges = append(edges, intersectEdge(e, sightline)...)
					if o := occludeWithIndex(e, center, radius*1.1, boundary, containsPoint); o != nil {
						occlusions = append(occlusions, o)
					}
				}
			}
		}
	}

	for i := 0; i < sightline.NumVertices(); i++ {
		edges = append(edges, [2]s2.Point{sightline.Vertex(i), sightline.Vertex(i + 1)})
	}

	for _, o := range occlusions {
		intersected := make([][2]s2.Point, 0, len(edges))
		for _, e := range edges {
			if e[0].Distance(e[1]) > tolerance {
				intersected = append(intersected, differenceEdge(e, o)...)
			}
		}
		edges = intersected
	}

	midpoints := make([]s2.Point, len(edges))
	for i, e := range edges {
		midpoints[i] = s2.Interpolate(0.5, e[0], e[1])
	}

	if len(edges) < 2 {
		return s2.PolygonFromLoops([]*s2.Loop{})
	}

	// Use the North pole as a fixed reference point when sorting edges to form
	// the sightline, to ensure that the first vertex on each sightline is in a
	// similar logical position to others nearby. This is useful for animating
	// the transitions between sightlines.
	reference := s2.PointFromCoords(0.0, 0.0, 1.0)
	sort.Sort(byAngle{center, reference, edges, midpoints})

	points := make([]s2.Point, 0, len(edges)*2)
	for i, e := range edges {
		if s2.OrderedCCW(e[0], midpoints[i], e[1], center) {
			points = append(points, e[0], e[1])
		} else {
			points = append(points, e[1], e[0])
		}
	}

	// Remove very thin spikes from the sightline, caused due to numerical accuracy
	// issues when computing the occlusion from two joined edges - the spike appears
	// at the joint. We match these by looking for two sightline vertices that are
	// close to each other, but the path along the loop between them is long. We
	// test for runs of up to 5 points, as the worst case is:
	// A, boundary intersection, boundary vertex, boundary intersection, B
	// where A and B are close to each other, and the points between cause the spike.
	// As the spike is 'thin', it can at worst traverse one boundary vertex.
	cleaned := make([]s2.Point, 0, len(points))
	for i := 1; i < len(points)+1; i++ {
		if points[i%len(points)].Distance(points[(i+1)%len(points)]) < tolerance {
			continue
		}
		var d1 s1.Angle
		removed := false
		for j := i + 1; j < i+4; j++ {
			d1 += points[(j-1)%len(points)].Distance(points[j%len(points)])
			d2 := points[i%len(points)].Distance(points[j%len(points)])
			if d2 < tolerance || d1/d2 > 10 {
				i = j
				removed = true
				break
			}
		}
		if removed {
			continue
		}
		if len(cleaned) == 0 || points[i%len(points)].Distance(cleaned[len(cleaned)-1]) > tolerance {
			cleaned = append(cleaned, points[i%len(points)])
		}
	}

	if len(cleaned) < 3 {
		return s2.PolygonFromLoops([]*s2.Loop{})
	}

	if cleaned[len(cleaned)-1].Distance(cleaned[0]) < tolerance {
		cleaned = cleaned[0 : len(cleaned)-1]
	}

	loop := s2.LoopFromPoints(cleaned)
	if err := loop.Validate(); err != nil {
		panic(fmt.Sprintf("Invalid sightline from %s: %s", s2.LatLngFromPoint(center), err))
	}
	return s2.PolygonFromLoops([]*s2.Loop{loop})
}

// This now unused implementation of occlude() serves as a reference to the optimised
// occludeWithIndex
func occlude(e [2]s2.Point, center s2.Point, radius s1.Angle) *s2.Loop {
	boundary := s2.RegularLoop(center, radius, 128)
	contained := [2]bool{boundary.ContainsPoint(e[0]), boundary.ContainsPoint(e[1])}
	if !contained[0] && !contained[1] {
		return nil
	}
	return occludeContainedEdge(e, contained, center, radius, boundary)
}

// Occlude computes the loop represeting the area of the cap with the given center and
// radius that's occluded by the edge ab when videwed from the center. Exposed for testing.
func Occlude(a s2.Point, b s2.Point, center s2.Point, radius s1.Angle) *s2.Loop {
	boundary := s2.RegularLoop(center, radius*1.1, 128)
	index := s2.NewShapeIndex()
	index.Add(boundary)
	containsPoint := s2.NewContainsPointQuery(index, s2.VertexModelOpen)
	return occludeWithIndex([2]s2.Point{a, b}, center, radius*1.1, boundary, containsPoint)
}

func occludeWithIndex(e [2]s2.Point, center s2.Point, radius s1.Angle, boundary *s2.Loop, query *s2.ContainsPointQuery) *s2.Loop {
	contained := [2]bool{query.Contains(e[0]), query.Contains(e[1])}
	if !contained[0] && !contained[1] {
		return nil
	}
	return occludeContainedEdge(e, contained, center, radius, boundary)
}

func occludeContainedEdge(e [2]s2.Point, contained [2]bool, center s2.Point, radius s1.Angle, boundary *s2.Loop) *s2.Loop {
	points := make([]s2.Point, 0, 4)
	var v0, v1 int
	if !contained[0] {
		var p s2.Point
		p, v0, _ = loopIntersection(e[0], e[1], boundary)
		points = append(points, p)
	} else {
		var p s2.Point
		p, v0, _ = loopIntersection(center, s2.InterpolateAtDistance(radius*1.1, center, e[0]), boundary)
		points = append(points, p, e[0])
	}
	if !contained[1] {
		var p s2.Point
		p, v1, _ = loopIntersection(e[0], e[1], boundary)
		points = append(points, p)
	} else {
		var p s2.Point
		p, v1, _ = loopIntersection(center, s2.InterpolateAtDistance(radius*1.1, center, e[1]), boundary)
		points = append(points, e[1], p)
	}

	i := v1 + 1
	for {
		points = append(points, boundary.Vertex(i))
		if i == v0 {
			break
		}
		i = (i + 1) % boundary.NumVertices()
	}
	return s2.LoopFromPoints(points)
}

type edgeModification int

const (
	edgeIntersection edgeModification = iota
	edgeDifference
)

func intersectEdge(e [2]s2.Point, l *s2.Loop) [][2]s2.Point {
	return modifyEdge(e, l, edgeIntersection)
}

func differenceEdge(e [2]s2.Point, l *s2.Loop) [][2]s2.Point {
	return modifyEdge(e, l, edgeDifference)
}

func modifyEdge(e [2]s2.Point, l *s2.Loop, m edgeModification) [][2]s2.Point {
	points := []s2.Point{e[0], e[1]}

	for i := 0; i < l.NumVertices(); i++ {
		a := l.Vertex(i)
		b := l.Vertex((i + 1) % l.NumVertices())
		for j := 0; j < len(points)-1; j++ {
			if s2.CrossingSign(a, b, points[j], points[j+1]) == s2.Cross {
				points = append(points, s2.Point{})
				for k := len(points) - 1; k > j+1; k-- {
					points[k] = points[k-1]
				}
				points[j+1] = s2.Intersection(a, b, points[j], points[j+1])
				// ab can only intersect the line formed by points once, since all points are colinear
				break
			}
		}
	}

	// Use the midpoint of the first two points for the containment test, since
	// either a or b could be a vertex of the loop
	midpoint := s2.Interpolate(0.5, points[0], points[1])
	contained := l.ContainsPoint(midpoint)
	if contained {
		min := s1.InfAngle()
		for i := 0; i < l.NumVertices(); i++ {
			d := s2.Project(midpoint, l.Vertex(i), l.Vertex((i+1)%l.NumVertices())).Distance(midpoint)
			if d < min {
				min = d
			}
		}
		if min < b6.MetersToAngle(0.01) {
			contained = false
		}
	}
	edges := make([][2]s2.Point, 0, len(points)-1)
	for i := 0; i < len(points)-1; i++ {
		if (m == edgeDifference && !contained) || (m == edgeIntersection && contained) {
			edges = append(edges, [2]s2.Point{points[i], points[i+1]})
		}
		contained = !contained
	}
	return edges
}

// Compute the area visible from a given point by repeatedly subtracting
// polygons representing the area occluded by each edge of each building
// from a cap representing the maximum visible area.
// This implementation places a significant amount of stress on our
// implementation of polygon operations, as it repeatedly subtracts
// small polygons with coincident vertices and colinear edges. In
// practice, this leads to accuracy issues that makes this implementation
// unsuitable for production use.
func SightlineUsingPolygonIntersection(from s2.Point, radius s1.Angle, w b6.World) *s2.Polygon {
	sightline := s2.PolygonFromLoops([]*s2.Loop{s2.RegularLoop(from, radius, 128)})

	cap := s2.CapFromCenterAngle(from, radius)
	features := b6.FindAreas(b6.Intersection{b6.MightIntersect{cap}, b6.Keyed{"#building"}}, w)
	for features.Next() {
		area := features.Feature()
		for i := 0; i < area.Len(); i++ {
			loop := area.Polygon(i).Loop(0)
			for j := 0; j < loop.NumVertices(); j++ {
				a := loop.Vertex(j)
				b := loop.Vertex((j + 1) % loop.NumVertices())
				midpoint := s2.Interpolate(0.5, a, b)
				if !loopIntersects(from, midpoint, loop, j) {
					// Offset the occlusion slightly from a and b, so that when we compute the
					// difference between the current sightline and consecutive occclusions,
					// we're not left with a thin sliver of incorrect visibility between them
					// due to numerical precision.
					tolerance := b6.MetersToAngle(0.5)
					ea := s2.InterpolateAtDistance(-tolerance, a, b)
					eb := s2.InterpolateAtDistance(a.Distance(b)+tolerance, a, b)
					if o := occlude([2]s2.Point{ea, eb}, from, radius*1.1); o != nil {
						if difference := geometry.PolygonDifferenceFoster(sightline, s2.PolygonFromLoops([]*s2.Loop{o})); len(difference) == 1 {
							sightline = difference[0]
						}
					}
				}
			}
		}
	}
	return sightline
}

type edgeEvent struct {
	x     float64
	begin bool
	edge  int
}

type edgeEvents []edgeEvent

func (e edgeEvents) Len() int           { return len(e) }
func (e edgeEvents) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e edgeEvents) Less(i, j int) bool { return e[i].x < e[j].x }

func (e edgeEvents) AppendEdge(v0 float64, v1 float64, i int) edgeEvents {
	if v1 < v0 {
		v0, v1 = v1, v0
	}
	return append(e, edgeEvent{x: v0, begin: true, edge: i}, edgeEvent{x: v1, begin: false, edge: i})
}

func crossingCandidates(es0 edgeEvents, es1 edgeEvents) [][2]int {
	candidates := make([][2]int, 0)
	if len(es0) == 0 || len(es1) == 0 {
		return [][2]int{}
	}
	es := [2]edgeEvents{es0, es1}
	i := [2]int{0, 0}
	live := [2]map[int]struct{}{make(map[int]struct{}), make(map[int]struct{})}
	for i[0] < len(es[0]) && i[1] < len(es[1]) {
		var sweep int
		if es[0][i[0]].x < es[1][i[1]].x {
			sweep = 0
		} else {
			sweep = 1
		}
		if es[sweep][i[sweep]].begin {
			edge := es[sweep][i[sweep]].edge
			live[sweep][edge] = struct{}{}
			other := [2]int{1, 0}[sweep]
			var candidate [2]int
			for eo := range live[other] {
				candidate[sweep] = edge
				candidate[other] = eo
				candidates = append(candidates, candidate)
			}
		} else {
			delete(live[sweep], es[sweep][i[sweep]].edge)
		}
		i[sweep]++
	}
	return candidates
}

func eventOrdinates(v0 s2.Point, v1 s2.Point) (float64, float64) {
	// As S2 points and edges lie on the surface of a unit sphere, if the vertices of an edge
	// cross the Y=0 or Z=0 plane, the maximum X value will be 1.0, rather than the maximum
	// value for the vertices.
	if v0.Y < 0 && v1.Y > 0 || v0.Y > 0 && v1.Y < 0 || v0.Z < 0 && v1.Z > 0 || v0.Z > 0 && v1.Z < 0 {
		if v0.X < v1.X {
			return v0.X, 1.0
		} else {
			return v1.X, 1.0
		}
	}
	return v0.X, v1.X
}

func buildEdgeEvents(es []s2.Edge) edgeEvents {
	events := make(edgeEvents, 0, 2*len(es))
	for i, e := range es {
		o0, o1 := eventOrdinates(e.V0, e.V1)
		events = events.AppendEdge(o0, o1, i)
	}
	sort.Sort(events)
	return events
}

type polarEvent struct {
	begin bool
	edge  int
}

type polarEvents []polarEvent

type byEventAngle struct {
	events []polarEvent
	edges  []polarEdge
}

func (b byEventAngle) Len() int      { return len(b.events) }
func (b byEventAngle) Swap(i, j int) { b.events[i], b.events[j] = b.events[j], b.events[i] }
func (b byEventAngle) Less(i, j int) bool {
	// Order events by the angle at which they occur. For events at the same angle,
	// we ensure the begin events occur before end events, and that begin events are ordered
	// by distance (closest to center to furthest), and end events are reverse ordered by
	// distance.
	ie, je := b.edges[b.events[i].edge], b.edges[b.events[j].edge]
	var ip, jp polarPoint
	if b.events[i].begin {
		ip = ie.V0
	} else {
		ip = ie.V1
	}
	if b.events[j].begin {
		jp = je.V0
	} else {
		jp = je.V1
	}
	if ip.T == jp.T {
		if b.events[i].begin == b.events[j].begin {
			if b.events[i].begin {
				if ip.R == jp.R {
					if ie.V1.R == je.V1.R {
						// Edges are duplicate
						return b.events[i].edge < b.events[j].edge
					}
					return ie.V1.R < je.V1.R
				}
				return ip.R < jp.R
			}
			if ip.R == jp.R {
				if ie.V0.R == je.V0.R {
					// Edges are duplicate
					return b.events[i].edge < b.events[j].edge
				}
				return ie.V0.R > je.V0.R
			}
			return ip.R > jp.R
		}
		return b.events[i].begin
	}
	return ip.T < jp.T
}

var reference = s2.PointFromCoords(0.0, 0.0, 1.0)

type polarPoint struct {
	R s1.Angle // Distance from center
	T s1.Angle // Angle counterclockwise from reference line from the center to the North pole
}

type polarEdge struct {
	V0 polarPoint
	V1 polarPoint
}

func (p *polarEdge) Interpolate(t s1.Angle, edge s2.Edge) s2.Point {
	return s2.Interpolate(((t - p.V0.T) / (p.V1.T - p.V0.T)).Radians(), edge.V0, edge.V1)
}

func (p *polarEdge) Split(edge s2.Edge, center s2.Point) ([2]polarEdge, [2]s2.Edge) {
	if !(p.V0.T < 2*math.Pi && p.V1.T > 2*math.Pi) {
		panic("Can't split edges that don't cross the reference")
	}
	var ps [2]polarEdge
	var es [2]s2.Edge
	x := p.Interpolate(2*math.Pi, edge)
	d := center.Distance(x)
	es[0].V0 = edge.V0
	es[0].V1 = x
	es[1].V0 = es[0].V1
	es[1].V1 = edge.V1

	ps[0].V0 = p.V0
	ps[0].V1 = polarPoint{R: d, T: 2 * math.Pi}
	ps[1].V0 = polarPoint{R: d, T: 0.0}
	// Recalculate the angle for V1, to ensure it exactly matches the result of TurnAngle for
	// the point (in case it was previously calculated as a delta from the V0). This ensures
	// it will be equal to the angle associated with a following edge connected to V1.
	ps[1].V1 = polarPoint{R: p.V1.R, T: math.Pi - s2.TurnAngle(es[1].V1, center, reference)}
	return ps, es
}

func outputSightlineEvents(barriers []s2.Edge, polar []polarEdge, events []polarEvent, points []s2.Point, filename string) {
	g := geojson.NewFeatureCollection()
	fs := make(map[int]*geojson.Feature)
	for i, e := range events {
		if e.begin {
			f := geojson.NewFeatureFromS2Edge(barriers[e.edge])
			f.Properties["edge"] = fmt.Sprintf("%d", e.edge)
			f.Properties["a"] = fmt.Sprintf("%g,%g", polar[e.edge].V0.T.Radians(), polar[e.edge].V1.T.Radians())
			f.Properties["begin"] = fmt.Sprintf("%d", i)
			g.AddFeature(f)
			fs[e.edge] = f
		} else {
			if f, ok := fs[e.edge]; ok {
				f.Properties["end"] = fmt.Sprintf("%d", i)
			}
		}
	}
	seen := make(map[s2.Point]int)
	for i, p := range points {
		if d, ok := seen[p]; ok {
			f := geojson.NewFeatureFromS2Point(p)
			f.Properties["vertex"] = fmt.Sprintf("%d", i)
			f.Properties["duplicate"] = fmt.Sprintf("%d", d)
			g.AddFeature(f)
		} else {
			seen[p] = i
		}
	}
	g.WriteToFile(filename)
}

func SightlineUsingPolarCoordinates2(center s2.Point, radius s1.Angle, w b6.World) *s2.Polygon {
	boundary := s2.RegularLoop(center, radius, 128)
	boundaryEvents := make(edgeEvents, 0, 2*boundary.NumVertices())
	for i := 0; i < boundary.NumVertices(); i++ {
		v0 := boundary.Vertex(i)
		v1 := boundary.Vertex((i + 1) % boundary.NumVertices())
		o0, o1 := eventOrdinates(v0, v1)
		boundaryEvents = boundaryEvents.AppendEdge(o0, o1, i)
	}
	sort.Sort(boundaryEvents)

	barriers := make([]s2.Edge, 0, 64)
	cap := s2.CapFromCenterAngle(center, radius)
	features := b6.FindAreas(b6.Intersection{b6.MightIntersect{Region: cap}, b6.Keyed{Key: "#building"}}, w)
	for features.Next() {
		area := features.Feature()
		for i := 0; i < area.Len(); i++ {
			loop := area.Polygon(i).Loop(0)
			for j := 0; j < loop.NumVertices(); j++ {
				v0 := loop.Vertex(j)
				v1 := loop.Vertex((j + 1) % loop.NumVertices())
				if center.Distance(v0) > radius && center.Distance(v1) > radius {
					continue
				}
				barriers = append(barriers, s2.Edge{V0: v0, V1: v1})
			}
		}
	}
	barrierEvents := buildEdgeEvents(barriers)

	candidates := crossingCandidates(boundaryEvents, barrierEvents)
	for _, c := range candidates {
		boundary0 := boundary.Vertex(c[0])
		boundary1 := boundary.Vertex((c[0] + 1) % boundary.NumVertices())
		if s2.CrossingSign(boundary0, boundary1, barriers[c[1]].V0, barriers[c[1]].V1) == s2.Cross {
			intersection := s2.Intersection(boundary0, boundary1, barriers[c[1]].V0, barriers[c[1]].V1)
			if center.Distance(barriers[c[1]].V0) < center.Distance(barriers[c[1]].V1) {
				barriers[c[1]] = s2.Edge{V0: barriers[c[1]].V0, V1: intersection}
			} else {
				barriers[c[1]] = s2.Edge{V0: intersection, V1: barriers[c[1]].V1}
			}
		}
	}

	// Rebuild barrier events, and the pass above has changed the set of barriers
	barrierEvents = buildEdgeEvents(barriers)
	candidates = crossingCandidates(barrierEvents, barrierEvents)
	continuations := make([]int, len(barriers), len(barriers)+16)
	for _, c := range candidates {
		c0 := c[0]
		for {
			c1 := c[1]
			for {
				if s2.CrossingSign(barriers[c0].V0, barriers[c0].V1, barriers[c1].V0, barriers[c1].V1) == s2.Cross {
					x := s2.Intersection(barriers[c0].V0, barriers[c0].V1, barriers[c1].V0, barriers[c1].V1)
					barriers = append(barriers, s2.Edge{V0: x, V1: barriers[c0].V1})
					continuations = append(continuations, 0)
					barriers[c0].V1 = x
					continuations[c0] = len(barriers) - 1
					barriers = append(barriers, s2.Edge{V0: x, V1: barriers[c1].V1})
					continuations = append(continuations, 0)
					barriers[c1].V1 = x
					continuations[c1] = len(barriers) - 1
				}
				if continuations[c1] != 0 {
					c1 = continuations[c1]
				} else {
					break
				}
			}
			if continuations[c0] != 0 {
				c0 = continuations[c0]
			} else {
				break
			}
		}
	}

	for i := 0; i < boundary.NumEdges(); i++ {
		barriers = append(barriers, boundary.Edge(i))
	}

	polar := make([]polarEdge, len(barriers), len(barriers)+16) // Slightly larger to handle split edges crossing North
	n := len(barriers)
	for i := 0; i < n; i++ {
		v0, v1 := barriers[i].V0, barriers[i].V1
		v0v1 := s2.TurnAngle(v1, center, v0)
		// Convert to v0v1 from and exterior to an interior angle,
		// swapping the order of v0 and v1 to ensure all edges are
		// ordered counterclockwise with respect to the center.
		if v0v1 < 0 {
			v0, v1 = v1, v0
			v0v1 = math.Pi + v0v1
			barriers[i] = s2.Edge{V0: v0, V1: v1}
		} else {
			v0v1 = math.Pi - v0v1
		}
		rv0 := math.Pi - s2.TurnAngle(v0, center, reference)
		if rv0+v0v1 <= 2*math.Pi {
			rv1 := math.Pi - s2.TurnAngle(v1, center, reference)
			if rv1 < rv0 { // true if rv0 ~= 2*math.Pi, and so is rounded to 0.
				rv1 += 2 * math.Pi
			}
			polar[i] = polarEdge{V0: polarPoint{R: center.Distance(v0), T: rv0}, V1: polarPoint{R: center.Distance(v1), T: rv1}}
		} else {
			p := polarEdge{V0: polarPoint{R: center.Distance(v0), T: rv0}, V1: polarPoint{R: center.Distance(v1), T: rv0 + v0v1}}
			ps, bs := p.Split(barriers[i], center)
			polar[i] = ps[0]
			barriers[i] = bs[0]
			polar = append(polar, ps[1])
			barriers = append(barriers, bs[1])
		}
	}
	pevs := make(polarEvents, len(polar)*2)
	for i := range polar {
		pevs[2*i] = polarEvent{edge: i, begin: true}
		pevs[(2*i)+1] = polarEvent{edge: i, begin: false}
	}
	sort.Sort(byEventAngle{events: pevs, edges: polar})

	points := make([]s2.Point, 0)
	live := make(map[int]int)
	current := -1
	for i := 0; i < len(pevs); i++ {
		if pevs[i].begin {
			if current < 0 {
				if verbose {
					log.Printf("reset: begin %d", pevs[i].edge)
				}
				current = pevs[i].edge
				if len(points) == 0 || barriers[current].V0 != points[len(points)-1] {
					points = append(points, barriers[current].V0)
				}
			} else {
				x := polar[current].Interpolate(polar[pevs[i].edge].V0.T, barriers[current])
				if polar[pevs[i].edge].V0.R < center.Distance(x) {
					if verbose {
						log.Printf("begin: %d -> %d", current, pevs[i].edge)
					}
					current = pevs[i].edge
					if x != points[len(points)-1] {
						points = append(points, x)
					}
					if barriers[current].V0 != points[len(points)-1] {
						points = append(points, barriers[current].V0)
					}
				}
			}
			live[pevs[i].edge] = i
		} else {
			if _, ok := live[pevs[i].edge]; !ok {
				panic(fmt.Sprintf("End of dead edge %d", pevs[i].edge))
			}
			delete(live, pevs[i].edge)
			if pevs[i].edge == current {
				if barriers[current].V1 != points[len(points)-1] {
					points = append(points, barriers[current].V1)
				}
				if len(live) == 0 {
					// This will happen towards the end of iteration, as we approach the reference
					// point
					current = -1
				} else {
					distance := s1.InfAngle()
					next := -1
					order := -1
					projection := s2.Point{}
					for edge, begin := range live {
						var x s2.Point
						var d s1.Angle
						if barriers[edge].V0 == barriers[current].V1 {
							x = barriers[edge].V0
							d = polar[edge].V0.R
						} else {
							x = polar[edge].Interpolate(polar[current].V1.T, barriers[edge])
							d = center.Distance(x)
						}
						if d < distance || d == distance && begin < order {
							next = edge
							distance = d
							order = begin
							projection = x
						}
					}
					if verbose {
						log.Printf("end: %d -> %d", current, next)
					}
					current = next
					if projection != points[len(points)-1] {
						points = append(points, projection)
					}
				}
			}
		}
	}

	if points[0] == points[len(points)-1] {
		points = points[0 : len(points)-1]
	}
	loop := s2.LoopFromPoints(points)

	if verboseGeoJSON {
		outputSightlineEvents(barriers, polar, pevs, points, "sightline-events.geojson")
		g := geojson.NewFeatureCollection()
		g.AddFeature(geojson.NewFeatureFromS2Loop(loop))
		g.AddFeature(geojson.NewFeatureFromS2Point(center))
		g.WriteToFile("sightline.geojson")
	}

	if err := loop.Validate(); err != nil {
		g := geojson.NewFeatureCollection()
		g.AddFeature(geojson.NewFeatureFromS2Loop(loop))
		g.AddFeature(geojson.NewFeatureFromS2Point(center))
		g.WriteToFile("sightline-panic.geojson")
		outputSightlineEvents(barriers, polar, pevs, points, "sightline-panic-events.geojson")
		panic(fmt.Sprintf("bad loop: %s", err))
	}
	return s2.PolygonFromLoops([]*s2.Loop{loop})
}

const entranceApproachDistanceMeters = 4.0

func pointApproach(entrance b6.Feature, area geometry.MultiPolygon, w b6.World) (s2.Point, bool) {
	segments := w.Traverse(entrance.FeatureID())
	for segments.Next() {
		s := segments.Segment()
		if h := s.Feature.Get("#highway"); h.IsValid() {
			p := s.Polyline()
			var approach s2.Point
			if l := p.Length(); l > b6.MetersToAngle(entranceApproachDistanceMeters) {
				approach, _ = p.Interpolate(float64(b6.MetersToAngle(entranceApproachDistanceMeters) / l))
			} else {
				approach, _ = p.Interpolate(0.5)
			}
			if !area.ContainsPoint(approach) {
				return approach, true
			}
		}
	}
	return s2.Point{}, false
}

func possibleEntraces(c *api.Context, area b6.AreaFeature) []b6.Feature {
	all := make([]b6.Feature, 0)
	entrances := make([]b6.Feature, 0)
	for i := 0; i < area.Len(); i++ {
		boundary := area.Feature(i)[0]
		for _, r := range boundary.References() {
			if point := c.World.FindFeatureByID(r.Source()); point != nil {
				all = append(all, point)
				if entrance := point.Get("entrance"); entrance.IsValid() {
					entrances = append(entrances, point)
				}
			}
		}
	}
	if len(entrances) > 0 {
		return entrances
	} else {
		return all
	}
}

func entranceApproach(c *api.Context, area b6.AreaFeature) (b6.Geometry, error) {
	if area != nil && area.Len() > 0 {
		m := area.MultiPolygon()
		for _, entrance := range possibleEntraces(c, area) {
			if approach, ok := pointApproach(entrance, m, c.World); ok {
				return b6.GeometryFromPoint(approach), nil
			}
		}
	}
	return nil, fmt.Errorf("no entrance found")
}
