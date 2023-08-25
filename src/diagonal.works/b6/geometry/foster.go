package geometry

import (
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

const (
	verbose                  = false
	angleEpsilon    s1.Angle = 0.001   // Roughly equivalent to 1m error in 1km
	distanceEpsilon s1.Angle = 1.5e-10 // Roughly 10mm in the UK, slighly larger than the width of the smallest s2 cell
	ratioEpsilon    float64  = 1.5e-10
	epsilon                  = 2e-16
)

func PolygonClipFoster(polygon *s2.Polygon, mask *s2.Polygon) MultiPolygon {
	return polygonOperationFoster(polygon, mask, clip)
}

func PolygonDifferenceFoster(pp *s2.Polygon, qq *s2.Polygon) MultiPolygon {
	return polygonOperationFoster(pp, qq, difference)
}

type intersectionType uint8

const (
	noIntersection intersectionType = iota
	xIntersection
	tIntersectionOnP
	tIntersectionOnQ
	vIntersection
	xOverlap
	tOverlapOnP
	tOverlapOnQ
	vOverlap
)

func (i intersectionType) Intersects() bool {
	return i >= xIntersection && i <= vIntersection
}

func (i intersectionType) Overlaps() bool {
	return i >= xOverlap && i <= vOverlap
}

func (t intersectionType) String() string {
	switch t {
	case noIntersection:
		return "noIntersection"
	case xIntersection:
		return "xIntersection"
	case tIntersectionOnP:
		return "tIntersectionOnP"
	case tIntersectionOnQ:
		return "tIntersectionOnQ"
	case vIntersection:
		return "vIntersection"
	case xOverlap:
		return "xOverlap"
	case tOverlapOnP:
		return "tOverlapOnP"
	case tOverlapOnQ:
		return "tOverlapOnQ"
	case vOverlap:
		return "vOverlap"
	}
	return strconv.Itoa(int(t))
}

func fraction2(q s2.Point, p1 s2.Point, p2 s2.Point) float64 {
	p1p2 := p1.Distance(p2)
	p1q := p1.Distance(q)
	p2q := p2.Distance(q)
	alpha := p1q.Radians() / p1p2.Radians()
	if p2q > p1p2 && p1q < p2q {
		alpha = -alpha
	}
	return alpha
}

func intersects2(p1 s2.Point, p2 s2.Point, q1 s2.Point, q2 s2.Point) intersectionType {
	if q1.Distance(p1) < distanceEpsilon {
		a := math.Abs(s2.TurnAngle(p1, p2, q2).Radians())
		if a_ := math.Abs(a - math.Pi); a_ < a {
			a = a_
		}
		if a < angleEpsilon.Radians() {
			return vOverlap
		}
		return vIntersection
	}

	a1 := math.Abs(s2.TurnAngle(p2, p1, q1).Radians())
	a2 := math.Abs(s2.TurnAngle(q1, q2, p2).Radians())
	if a := math.Abs(a1 - math.Pi); a < a1 {
		a1 = a
	}
	if a := math.Abs(a2 - math.Pi); a < a2 {
		a2 = a
	}
	if a1 < angleEpsilon.Radians() && a2 < angleEpsilon.Radians() {
		alpha := fraction2(q1, p1, p2)
		beta := fraction2(p1, q1, q2)
		if alpha > ratioEpsilon && alpha < 1.0-ratioEpsilon && beta > ratioEpsilon && beta < 1.0-ratioEpsilon {
			return xOverlap
		} else if beta > ratioEpsilon && beta < 1.0-ratioEpsilon && (alpha < -ratioEpsilon || alpha > 1.0+ratioEpsilon) {
			return tOverlapOnQ
		} else if alpha > ratioEpsilon && alpha < 1.0-ratioEpsilon && (beta < -ratioEpsilon || beta > 1.0+ratioEpsilon) {
			return tOverlapOnP
		}
		return noIntersection
	} else if x := s2.Project(p1, q1, q2); x.Distance(p1) < distanceEpsilon {
		if p1.Distance(q1) < q1.Distance(q2)-distanceEpsilon {
			return tIntersectionOnQ
		}
		return noIntersection // No intersection at the end of edges
	} else if x := s2.Project(q1, p1, p2); x.Distance(q1) < distanceEpsilon {
		if p1.Distance(q1) < p1.Distance(p2)-distanceEpsilon {
			return tIntersectionOnP
		}
		return noIntersection // No intersection at the end of edges
	} else if s2.CrossingSign(p1, p2, q1, q2) == s2.Cross {
		x := s2.Intersection(p1, p2, q1, q2)
		if x.Distance(p2) < distanceEpsilon || x.Distance(q2) < distanceEpsilon {
			return noIntersection // No intersection at the end of edges
		}
		if x.Distance(p1) < distanceEpsilon || x.Distance(q1) < distanceEpsilon {
			debug("too close to p1/p2")
		}
		return xIntersection
	}
	return noIntersection
}

type intersectionLabel uint8

const (
	noLabel intersectionLabel = iota
	crossing
	bouncing
	leftOn
	rightOn
	onOn
	onLeft
	onRight
	delayedCrossing
	delayedBouncing
)

func (t intersectionLabel) String() string {
	switch t {
	case noLabel:
		return "noLabel"
	case crossing:
		return "crossing"
	case bouncing:
		return "bouncing"
	case leftOn:
		return "leftOn"
	case rightOn:
		return "rightOn"
	case onOn:
		return "onOn"
	case onLeft:
		return "onLeft"
	case onRight:
		return "onRight"
	case delayedCrossing:
		return "delayedCrossing"
	case delayedBouncing:
		return "delayedBouncing"
	}
	return strconv.Itoa(int(t))
}

type entryExitLabel uint8

const (
	neither entryExitLabel = iota
	enters
	exits
)

func (e entryExitLabel) String() string {
	switch e {
	case neither:
		return "neither"
	case enters:
		return "enters"
	case exits:
		return "exits"
	}
	return strconv.Itoa(int(e))
}

func (e entryExitLabel) Toggle() entryExitLabel {
	switch e {
	case enters:
		return exits
	case exits:
		return enters
	}
	return neither
}

type vertex2 struct {
	p         s2.Point
	previous  *vertex2
	next      *vertex2
	type_     intersectionType
	label     intersectionLabel
	enters    entryExitLabel
	neighbour *vertex2
	debug     string
}

func (v *vertex2) Insert(new *vertex2) {
	new.next = v.next
	new.previous = v
	v.next = new
	new.next.previous = new
}

func toVertices(p *s2.Polygon) []*vertex2 {
	l := make([]*vertex2, p.NumLoops())
	for i := 0; i < p.NumLoops(); i++ {
		loop := p.Loop(i)
		var previous *vertex2
		for j := 0; j < loop.NumVertices(); j++ {
			v := &vertex2{p: loop.Vertex(j), debug: fmt.Sprintf("%d/%d", i, j)}
			if previous != nil {
				previous.next = v
				v.previous = previous
			} else {
				l[i] = v
			}
			previous = v
		}
		l[i].previous = previous
		l[i].previous.next = l[i]
	}
	return l
}

func toPolygon(p []*vertex2) *s2.Polygon {
	loops := make([]*s2.Loop, len(p))
	for i, pv := range p {
		loops[i] = toLoop(pv)
	}
	return s2.PolygonFromLoops(loops)
}

func toLoop(start *vertex2) *s2.Loop {
	points := make([]s2.Point, 0, 3)
	pv := start
	for {
		points = append(points, pv.p)
		pv = pv.next
		if pv == start {
			break
		}
	}
	return s2.LoopFromPoints(points)
}

type operation int

const (
	clip operation = iota
	difference
)

func polygonOperationFoster(pp *s2.Polygon, qq *s2.Polygon, op operation) MultiPolygon {
	p := toVertices(pp)
	q := toVertices(qq)
	debug("p points")
	logVertices(p)
	debug("q points")
	logVertices(q)
	addIntersections(p, q)
	labelIntersections(p)
	labelDelayedIntersections(p)
	copyLabelsToNeighbours(p)
	debug("p intersections")
	logVertices(p)
	debug("q intersections")
	logVertices(q)
	var err error
	var pCandidates, qCandidates map[*vertex2]struct{}
	if pCandidates, err = labelEntersAndExits(p, q, op); err != nil {
		panic(err)
	}
	if qCandidates, err = labelEntersAndExits(q, p, op); err != nil {
		panic(err)
	}
	labelCrossingCandidates(pCandidates, qCandidates)
	debug("p labels")
	logVertices(p)
	debug("q labels")
	logVertices(q)
	return traceResult(p, q, op)
}

func addIntersections(p []*vertex2, q []*vertex2) {
	for i := 0; i < len(p); i++ {
		debug("loop %d", i)
		pv := p[i]
		for {
			for j := 0; j < len(q); j++ {
				qv := q[j]
				for {
					if pv.neighbour != qv {
						// TODO: What happens if there are multiple tIntersections at the same point. We don't insert a new
						// vertex, so maybe we overwrite an old intersection. Is there a valid polygon for which this is true?
						// Possibly one with loops that intersect a line at the same place.
						t := intersects2(pv.p, pv.next.p, qv.p, qv.next.p)
						debug("* %p %p %s %s t: %s", pv, qv, pv.debug, qv.debug, t)
						if t == xIntersection {
							x := s2.Intersection(pv.p, pv.next.p, qv.p, qv.next.p)
							newP := &vertex2{p: x, type_: t, debug: fmt.Sprintf("%sx%s", pv.debug, qv.debug)}
							newQ := &vertex2{p: x, type_: t, debug: fmt.Sprintf("%sx%s", qv.debug, pv.debug)}
							newP.neighbour = newQ
							newQ.neighbour = newP
							pv.Insert(newP)
							qv.Insert(newQ)
							qv = newQ
						} else if t == xOverlap {
							newP := &vertex2{p: qv.p, type_: t, debug: fmt.Sprintf("%sx%s", pv.debug, qv.debug)}
							newQ := &vertex2{p: pv.p, type_: t, debug: fmt.Sprintf("%sx%s", qv.debug, pv.debug)}
							newP.neighbour = qv
							qv.neighbour = newP
							newQ.neighbour = pv
							pv.neighbour = newQ
							pv.Insert(newP)
							qv.Insert(newQ)
							qv = newQ
						} else if t == tIntersectionOnP || t == tOverlapOnP {
							newP := &vertex2{p: qv.p, type_: t, debug: fmt.Sprintf("%sx%s", pv.debug, qv.debug)}
							if qv.type_ != noIntersection {
								panic("Reset intersection")
							}
							newP.neighbour = qv
							qv.type_ = t
							qv.neighbour = newP
							pv.Insert(newP)
						} else if t == tIntersectionOnQ || t == tOverlapOnQ {
							newQ := &vertex2{p: pv.p, type_: t, debug: fmt.Sprintf("%sx%s", qv.debug, pv.debug)}
							newQ.neighbour = pv
							pv.type_ = t
							pv.neighbour = newQ
							qv.Insert(newQ)
							qv = newQ
						} else if t == vIntersection || t == vOverlap {
							if pv.type_ != noIntersection {
								panic(fmt.Sprintf("Reset intersection pv: %s -> %s at %s", pv.type_, t, pv.debug))
							}
							pv.type_ = t
							pv.neighbour = qv
							if qv.type_ != noIntersection {
								panic(fmt.Sprintf("Reset intersection qv: %s -> %s at %s", qv.type_, t, qv.debug))
							}
							qv.type_ = t
							qv.neighbour = pv
						}
					}
					qv = qv.next
					if qv == q[j] {
						break
					}
				}
			}
			pv = pv.next
			if pv == p[i] {
				break
			}
		}
	}
}

type position uint8

const (
	left position = iota
	right
	isPMinus
	isPPlus
)

func (p position) String() string {
	switch p {
	case left:
		return "left"
	case right:
		return "right"
	case isPMinus:
		return "isPMinus"
	case isPPlus:
		return "isPPlus"
	}
	return strconv.Itoa(int(p))
}

func relate(q *vertex2, p1 *vertex2, p2 *vertex2, p3 *vertex2) position {
	if q.neighbour == p1 {
		return isPMinus
	}
	if q.neighbour == p3 {
		return isPPlus
	}

	s1_ := s2.SignedArea(q.p, p1.p, p2.p)
	s2_ := s2.SignedArea(q.p, p2.p, p3.p)
	s3_ := s2.SignedArea(p1.p, p2.p, p3.p)
	if s3_ > epsilon {
		// p1 p2 p3 takes a left turn
		if s1_ > 0.0 && s2_ > 0.0 {
			return left
		}
		return right

	} else if math.Abs(s3_) < epsilon {
		// p1 p2 p3 is a straight line
		if s1_ > 0.0 {
			return left
		} else {
			return right
		}

	} else {
		// p1 p2 p3 takes a right turn
		if s1_ > 0 || s2_ > 0 {
			return left
		}
		return right
	}
}

func labelIntersections(p []*vertex2) {
	for i := 0; i < len(p); i++ {
		pv := p[i]
		for {
			if pv.type_ == xIntersection {
				pv.label = crossing
			} else if pv.neighbour != nil {
				qMinus := relate(pv.neighbour.previous, pv.previous, pv, pv.next)
				qPlus := relate(pv.neighbour.next, pv.previous, pv, pv.next)
				if (qMinus == left && qPlus == right) || (qMinus == right && qPlus == left) {
					pv.label = crossing
				} else if (qMinus == left && qPlus == left) || (qMinus == right && qPlus == right) {
					pv.label = bouncing
				} else if (qMinus == right && qPlus == isPPlus) || (qMinus == isPPlus && qPlus == right) {
					pv.label = leftOn
				} else if (qMinus == left && qPlus == isPPlus) || (qMinus == isPPlus && qPlus == left) {
					pv.label = rightOn
				} else if (qMinus == right && qPlus == isPMinus) || (qMinus == isPMinus && qPlus == right) {
					pv.label = onLeft
				} else if (qMinus == left && qPlus == isPMinus) || (qMinus == isPMinus && qPlus == left) {
					pv.label = onRight
				} else {
					debug("no label for intersection with neighbour %s %s %s", pv.type_, qMinus, qPlus)
				}
			}
			pv = pv.next
			if pv == p[i] {
				break
			}
		}
	}
}

func labelDelayedIntersections(p []*vertex2) {
	for i := 0; i < len(p); i++ {
		pv := p[i]
		var start *vertex2
		done := false
		// We may start traversal in the middle of a chain, in which case we need to
		// wrap around visit the end
		for !done || start != nil {
			if start == nil {
				if pv.label == leftOn || pv.label == rightOn {
					start = pv
				}
			} else {
				if pv.label == onOn {
					pv.label = bouncing
				} else if pv.label == onLeft || pv.label == onRight {
					bouncing := (pv.label == onLeft && start.label == leftOn) || (pv.label == onRight && start.label == rightOn)
					if bouncing {
						start.label = delayedBouncing
						pv.label = delayedBouncing
					} else {
						start.label = delayedCrossing
						pv.label = delayedCrossing
					}
					start = nil
				}
			}
			pv = pv.next
			if pv == p[i] {
				done = true
			}
		}
	}
}

func copyLabelsToNeighbours(p []*vertex2) {
	for i, pv := range p {
		for {
			if pv.neighbour != nil {
				pv.neighbour.label = pv.label
			}
			pv = pv.next
			if pv == p[i] {
				break
			}
		}
	}
}

func findStartVertexAndStatus(p *vertex2, q *s2.Polygon) (*vertex2, entryExitLabel) {
	pv := p
	var start *vertex2
	var point s2.Point
	var notColinear *vertex2
	for {
		if pv.type_ == noIntersection {
			start = pv
			point = start.p
			break
		}
		if pv.next.neighbour != pv.neighbour.previous && pv.next.neighbour != pv.neighbour.next {
			notColinear = pv
		}
		pv = pv.next
		if pv == p {
			break
		}
	}
	if start == nil {
		if notColinear != nil {
			start = notColinear.next
			point = s2.Interpolate(0.5, notColinear.p, notColinear.next.p)
		} else {
			panic("loops identical")
		}
	}
	var status entryExitLabel
	if q.ContainsPoint(point) {
		status = exits
	} else {
		status = enters
	}
	return start, status
}

func labelEntersAndExits(p []*vertex2, q []*vertex2, op operation) (map[*vertex2]struct{}, error) {
	debug("labelEntersAndExits")
	qp := toPolygon(q)
	crossingCandidates := make(map[*vertex2]struct{})
	for i := 0; i < len(p); i++ {
		start, status := findStartVertexAndStatus(p[i], qp)
		inChain := false
		debug("start %p %s", start, status)
		pv := start
		for {
			if pv.label == crossing {
				debug("  crossing %p %s", pv, status)
				pv.enters = status
				status = status.Toggle()
			} else if pv.label == delayedCrossing {
				pv.enters = status
				if !inChain {
					debug("  start delayed crossing %p %s", pv, status)
					if (op == clip && status == exits) || (op == difference && status == enters) {
						pv.label = crossing
					}
				} else {
					debug("  end delayed crossing %p %s", pv, status)
					if (op == clip && status == enters) || (op == difference && status == exits) {
						pv.label = crossing
					}
					status = status.Toggle()
				}
				inChain = !inChain
			} else if pv.label == delayedBouncing {
				pv.enters = status
				if !inChain {
					debug("  start delayed bouncing %p %s", pv, status)
					if (op == clip && status == exits) || op == difference {
						debug("    add crossing candidate: %p/%p", pv, pv.neighbour)
						crossingCandidates[pv] = struct{}{}
					}
				} else {
					debug("  end delayed bouncing %p %s", pv, status)
					if (op == clip && status == enters) || op == difference {
						debug("    add crossing candidate: %p/%p", pv, pv.neighbour)
						crossingCandidates[pv] = struct{}{}
					}
				}
				status = status.Toggle()
				inChain = !inChain
			}
			pv = pv.next
			if pv == start {
				break
			}
		}
	}
	return crossingCandidates, nil
}

func labelCrossingCandidates(p map[*vertex2]struct{}, q map[*vertex2]struct{}) {
	for pv := range p {
		if _, ok := q[pv.neighbour]; ok {
			debug("  crossing candidate -> crossing %p/%p", pv, pv.neighbour)
			pv.label = crossing
			pv.neighbour.label = crossing
		}
	}
}

func sxy(p s2.Point) string {
	ll := s2.LatLngFromPoint(p)
	return fmt.Sprintf("%.1f,%.1f", ll.Lng.Degrees(), ll.Lat.Degrees())
}

func traceResult(p []*vertex2, q []*vertex2, op operation) MultiPolygon {
	debug("traceResult")
	loops := make([]*s2.Loop, 0)

	if op == clip {
		loops = appendLoopsWithoutCrossings(loops, p, q, true)
		loops = appendLoopsWithoutCrossings(loops, q, p, true)
	} else if op == difference {
		loops = appendLoopsWithoutCrossings(loops, p, q, false)
		loops = appendLoopsWithoutCrossings(loops, q, p, true)
	}

	for i, cv := range p {
		for {
			// When clipping, start at crossings that enter, so that traversal starts forwards,
			// resulting in correctly counter-clockwise ordered loops. When differencing,
			// start at exist
			var start entryExitLabel
			if op == clip {
				start = enters
			} else if op == difference {
				start = exits
			}
			onPnotQ := true
			if cv.label == crossing && cv.enters == start {
				points := make([]s2.Point, 0)
				pv := cv
				status := pv.enters
				debug("start: %p %v", pv, pv.enters)
				seen := make(map[*vertex2]struct{})
				for {
					if _, ok := seen[pv]; ok {
						return MultiPolygon{} // There's a loop, which shouldn't happen
					}
					seen[pv] = struct{}{}
					for {
						if op == clip {
							if status == enters {
								pv = pv.next
							} else {
								pv = pv.previous
							}
						} else if op == difference {
							if onPnotQ {
								if status == enters {
									pv = pv.previous
								} else {
									pv = pv.next
								}
							} else {
								if status == enters {
									pv = pv.next
								} else {
									pv = pv.previous
								}
							}
						}
						if pv.label == crossing {
							debug("    crossing %v", pv.enters)
						}
						debug("  add %p", pv)
						points = append(points, pv.p)
						pv.label = noLabel
						if pv.enters == status.Toggle() || pv == cv || pv.neighbour == cv {
							break
						}
					}
					if pv == cv || pv.neighbour == cv {
						debug("  end %p", pv)
						break
					}
					debug("  switch %p %s %s", pv, pv.enters, pv.neighbour.enters)
					pv = pv.neighbour
					pv.label = noLabel
					onPnotQ = !onPnotQ
					status = pv.enters
				}
				loops = append(loops, s2.LoopFromPoints(points))
			}
			cv = cv.next
			if cv == p[i] {
				break
			}
		}
	}
	return NewMultiPolygonFromLoops(loops)
}

func appendLoopsWithoutCrossings(loops []*s2.Loop, p []*vertex2, q []*vertex2, appendInside bool) []*s2.Loop {
	qp := toPolygon(q)
	for i, pv := range p {
		crossings := false
		for {
			if pv.label == crossing {
				crossings = true
				break
			}
			pv = pv.next
			if pv == p[i] {
				break
			}
		}
		if !crossings {
			_, status := findStartVertexAndStatus(p[i], qp)
			inside := status == exits
			if (inside && appendInside) || (!inside && !appendInside) {
				loops = append(loops, toLoop(p[i]))
			}
		}
	}
	return loops
}

func debug(format string, values ...interface{}) {
	if verbose {
		log.Printf(format, values...)
	}
}

func logVertices(vs []*vertex2) {
	for i := 0; i < len(vs); i++ {
		debug("loop %d", i)
		v := vs[i]
		for {
			debug("  %p %s %s %v n: %p", v, v.type_, v.label, v.enters, v.neighbour)
			v = v.next
			if v == vs[i] {
				break
			}
		}
	}
}
