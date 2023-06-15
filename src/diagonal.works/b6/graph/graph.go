package graph

import (
	"container/heap"
	"math"

	"diagonal.works/b6"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

const WalkingMetersPerSecond = 5000.0 / (60.0 * 60.0)

func weightFromSegment(segment b6.Segment) float64 {
	weight := b6.AngleToMeters(segment.Polyline().Length())
	if factor := segment.Feature.Get("diagonal:weight"); factor.IsValid() {
		if f, ok := factor.FloatValue(); ok {
			return weight * f
		}
	}
	return weight
}

type Weights interface {
	IsUseable(segment b6.Segment) bool
	Weight(segment b6.Segment) float64
}

type SimpleWeights struct{}

func (SimpleWeights) IsUseable(segment b6.Segment) bool {
	return true
}

func (SimpleWeights) Weight(segment b6.Segment) float64 {
	return weightFromSegment(segment)
}

type SimpleHighwayWeights struct{}

func (SimpleHighwayWeights) IsUseable(segment b6.Segment) bool {
	if highway := segment.Feature.Get("#highway"); highway.IsValid() {
		return true
	}
	return segment.Feature.Get("diagonal").Value == "connection"
}

func (SimpleHighwayWeights) Weight(segment b6.Segment) float64 {
	return weightFromSegment(segment)
}

func IsPathUsableByBus(path b6.PathFeature) bool {
	if path.Get("diagonal").Value == "connection" {
		return true
	}
	if highway := path.Get("#highway"); highway.IsValid() {
		// TODO: Make this an accept list, rather than a reject list
		isPath := highway.Value == "footway" || highway.Value == "steps" || highway.Value == "corridor" || highway.Value == "path" || highway.Value == "pedestrian"
		if isPath {
			return false
		}
		if highway.Value == "cycleway" || highway.Value == "bridleway" || highway.Value == "escape" {
			return false
		}
		if highway.Value == "proposed" || highway.Value == "construction" {
			return false
		}
		if access := path.Get("access"); access.Value == "no" {
			return path.Get("bus").Value == "yes"
		}
		return true
	}
	return false
}

func IsPathPreferredByBus(path b6.PathFeature) bool {
	highway := path.Get("#highway").Value
	return highway == "primary" || highway == "secondary" || highway == "trunk"
}

func IsSegmentUseableInThisDirectionByBus(segment b6.Segment) bool {
	if oneway := segment.Feature.Get("oneway"); oneway.Value != "yes" {
		return true
	}
	if oneway := segment.Feature.Get("oneway:bus"); oneway.Value == "no" {
		return true
	}
	return segment.Last > segment.First
}

type BusWeights struct{}

func (BusWeights) IsUseable(segment b6.Segment) bool {
	return IsSegmentUseableInThisDirectionByBus(segment) && IsPathUsableByBus(segment.Feature)
}

func (BusWeights) Weight(segment b6.Segment) float64 {
	return weightFromSegment(segment)
}

func IsPathUsableByCar(path b6.PathFeature) bool {
	if path.Get("diagonal").Value == "connection" {
		return true
	}
	if highway := path.Get("#highway"); highway.IsValid() {
		// TODO: Make this an accept list, rather than a reject list
		isPath := highway.Value == "footway" || highway.Value == "steps" || highway.Value == "corridor" || highway.Value == "path" || highway.Value == "pedestrian"
		if isPath {
			return false
		}
		if highway.Value == "cycleway" || highway.Value == "bridleway" || highway.Value == "escape" {
			return false
		}
		if highway.Value == "proposed" || highway.Value == "construction" {
			return false
		}
		return true
	}
	return false
}

func IsSegmentUseableInThisDirectionByCar(segment b6.Segment) bool {
	if oneway := segment.Feature.Get("oneway"); oneway.Value != "yes" {
		return true
	}
	return segment.Last > segment.First
}

type CarWeights struct{}

func (CarWeights) IsUseable(segment b6.Segment) bool {
	return IsSegmentUseableInThisDirectionByCar(segment) && IsPathUsableByCar(segment.Feature)
}

func (CarWeights) Weight(segment b6.Segment) float64 {
	return weightFromSegment(segment)
}

func IsPathUsableByPedestrian(path b6.PathFeature) bool {
	// Taken from the table here:
	// https://wiki.openstreetmap.org/wiki/OSM_tags_for_routing/Access_restrictions#United_Kingdom
	// TODO: Factor out the logic here and IsPathUsableByBus
	// TODO: Also take into account access tags
	if path.Get("diagonal").Value == "connection" {
		return true
	}
	if highway := path.Get("#highway"); highway.IsValid() {
		return highway.Value != "motorway"
	}
	return false
}

type ElevationWeights struct {
	UpHillHard   bool
	DownHillHard bool
}

func (ElevationWeights) IsUseable(segment b6.Segment) bool {
	return SimpleHighwayWeights{}.IsUseable(segment)
}

func (e ElevationWeights) Weight(segment b6.Segment) float64 {
	var weight float64

	elevation, fromMemory := 0.0, false

	first, last := segment.First, segment.Last
	if first > last {
		first, last = last, first
	}

	path := segment.Feature
	for i := first; i < last; i++ {
		start := path.Feature(i)
		stop := path.Feature(i + 1)
		w := b6.AngleToMeters((*s2.Polyline)(&[]s2.Point{start.Point(), stop.Point()}).Length())

		startElevation, ok := start.Get("ele").FloatValue()
		if ok {
			elevation, fromMemory = startElevation, ok
		} else {
			startElevation, ok = elevation, fromMemory
		}

		stopElevation, ok := stop.Get("ele").FloatValue()

		if fromMemory && ok {

			if stopElevation > startElevation { // Ascending.
				// Naismith’s Rule adds ~6s/m of elevation,
				// which we're normalizing against 1.38m/s avg. walking speed.
				w += (stopElevation-startElevation) * 6 * WalkingMetersPerSecond
			}

			if (e.UpHillHard && stopElevation > startElevation) ||
			   (e.DownHillHard && stopElevation < startElevation) {
				w *= 1.2  // Arbitrary coefficient.
			}
		}

		weight += w
	}

	return weight
}

func interpolateShortestPathDistances(segment b6.Segment, firstDistance s1.Angle, lastDistance s1.Angle) []s1.Angle {
	distances := make([]s1.Angle, segment.Len())
	distances[0] = firstDistance
	distances[len(distances)-1] = lastDistance
	for i := 1; i < len(distances)-1; i++ {
		distances[i] = s1.InfAngle()
	}

	points := make([]s2.Point, segment.Len())
	for i := 0; i < segment.Len(); i++ {
		points[i] = segment.SegmentPoint(i)
	}

	for i := 1; i < len(distances); i++ {
		distance := distances[i-1] + points[i-1].Distance(points[i])
		if distance < distances[i] {
			distances[i] = distance
		} else {
			break
		}
	}

	for i := len(distances) - 2; i >= 0; i-- {
		distance := distances[i+1] + points[i+1].Distance(points[i])
		if distance < distances[i] {
			distances[i] = distance
		} else {
			break
		}
	}

	return distances
}

type reachable struct {
	point    b6.PointID
	visited  bool
	distance float64
	segment  b6.Segment
	index    int // Index of this entry within the heap entries, negative if removed
}

type PathState int

const (
	PathStateTraversed PathState = iota
	PathStateNotUseable
	PathStateTooFar
)

type ShortestPathSearch struct {
	queue      []*reachable
	byPoint    map[b6.PointID]*reachable
	byArea     map[b6.AreaID]*reachable // The reachable instance for the entrance used to enter the area
	pathStates map[b6.SegmentKey]PathState
}

func NewShortestPathSearchFromFeature(f b6.Feature, weights Weights, w b6.World) *ShortestPathSearch {
	switch f := f.(type) {
	case b6.PointFeature:
		return NewShortestPathSearchFromPoint(f.PointID())
	case b6.AreaFeature:
		return NewShortestPathSearchFromBuilding(f, weights, w)
	}
	return newShortestPathSearch()
}

func NewShortestPathSearchFromPoint(from b6.PointID) *ShortestPathSearch {
	s := newShortestPathSearch()
	s.queue = append(s.queue, &reachable{point: from, visited: false, distance: 0.0, segment: b6.SegmentInvalid, index: 0})
	s.byPoint[from] = s.queue[0]
	return s
}

func isConnected(p b6.PointID, weights Weights, w b6.World) bool {
	ps := w.FindPathsByPoint(p)
	for ps.Next() {
		if weights.IsUseable(b6.ToSegment(ps.Feature())) {
			return true
		}
	}
	return false
}

func NewShortestPathSearchFromBuilding(area b6.AreaFeature, weights Weights, w b6.World) *ShortestPathSearch {
	s := newShortestPathSearch()
	for i := 0; i < area.Len(); i++ {
		for _, path := range area.Feature(i) {
			for j := 0; j < path.Len(); j++ {
				if point := path.Feature(j); point != nil {
					if isConnected(point.PointID(), weights, w) {
						r := &reachable{point: point.PointID(), visited: false, distance: 0.0, segment: b6.SegmentInvalid, index: len(s.queue)}
						s.queue = append(s.queue, r)
					}
				}
			}
		}
	}
	for _, r := range s.queue {
		s.byPoint[r.point] = r
	}
	return s
}

func newShortestPathSearch() *ShortestPathSearch {
	return &ShortestPathSearch{
		queue:      make([]*reachable, 0, 64),
		byPoint:    make(map[b6.PointID]*reachable),
		byArea:     make(map[b6.AreaID]*reachable),
		pathStates: make(map[b6.SegmentKey]PathState),
	}
}

func (s *ShortestPathSearch) Len() int { return len(s.queue) }

func (s *ShortestPathSearch) Less(i, j int) bool {
	return s.queue[i].distance < s.queue[j].distance
}

func (s *ShortestPathSearch) Swap(i, j int) {
	s.queue[i], s.queue[j] = s.queue[j], s.queue[i]
	s.queue[i].index = i
	s.queue[j].index = j
}

func (s *ShortestPathSearch) Push(x interface{}) {
	r := x.(*reachable)
	s.queue = append(s.queue, r)
	r.index = len(s.queue) - 1
	s.byPoint[r.point] = r
}

func (s *ShortestPathSearch) Pop() interface{} {
	old := s.queue
	n := len(old)
	r := old[n-1]
	r.index = -1
	s.queue = old[0 : n-1]
	return r
}

func (s *ShortestPathSearch) AddOrUpdate(segment b6.Segment, distance float64, features ShortestPathFeatures, w b6.World) {
	point := segment.LastFeature().PointID()
	updated := false
	var r *reachable
	var ok bool
	if r, ok = s.byPoint[point]; ok {
		if r.distance > distance {
			r.distance = distance
			r.segment = segment
			heap.Fix(s, r.index)
			updated = true
		}
	} else {
		r = &reachable{point: point, distance: distance, segment: segment, index: -1}
		heap.Push(s, r)
		updated = true
	}
	if updated && features == PointsAndAreas {
		i := w.FindAreasByPoint(point)
		for i.Next() {
			area := i.Feature().AreaID()
			if current, ok := s.byArea[area]; !ok || current.distance > distance {
				// TODO: Maybe we should also keep track of the node we used reach the area
				s.byArea[area] = r
			}
		}
	}
}

func (s *ShortestPathSearch) CurrentDistance(point b6.PointID) float64 {
	if r, ok := s.byPoint[point]; ok {
		return r.distance
	}
	return math.Inf(1)
}

type ShortestPathFeatures bool

const (
	Points         ShortestPathFeatures = false
	PointsAndAreas                      = true
)

func (s *ShortestPathSearch) ExpandSearchTo(to b6.PointID, maxDistance float64, weights Weights, w b6.World) {
	destination := &reachable{point: to, distance: math.Inf(1)}
	s.byPoint[to] = destination
	heap.Push(s, destination)
	for s.Len() > 0 {
		r := heap.Pop(s).(*reachable)
		s.byPoint[r.point].visited = true
		if r.point == to || destination.distance < r.distance {
			break
		}
		ss := w.Traverse(r.point)
		for ss.Next() {
			segment := ss.Segment()
			point := segment.LastFeature()
			if next, ok := s.byPoint[point.PointID()]; !ok || !next.visited {
				if weights.IsUseable(segment) {
					weight := weights.Weight(segment)
					if r.distance+weight < maxDistance {
						s.pathStates[segment.ToKey()] = PathStateTraversed
						s.AddOrUpdate(segment, r.distance+weight, false, w)
					} else {
						s.pathStates[segment.ToKey()] = PathStateTooFar
					}
				} else {
					s.pathStates[segment.ToKey()] = PathStateNotUseable
				}
			}
		}
	}
}

func (s *ShortestPathSearch) ExpandSearch(maxDistance float64, weights Weights, features ShortestPathFeatures, w b6.World) {
	for s.Len() > 0 {
		r := heap.Pop(s).(*reachable)
		s.byPoint[r.point].visited = true
		ss := w.Traverse(r.point)
		for ss.Next() {
			segment := ss.Segment()
			point := segment.LastFeature()
			if next, ok := s.byPoint[point.PointID()]; !ok || !next.visited {
				if weights.IsUseable(segment) {
					weight := weights.Weight(segment)
					if r.distance+weight < maxDistance {
						s.pathStates[segment.ToKey()] = PathStateTraversed
						s.AddOrUpdate(segment, r.distance+weight, features, w)
					} else {
						s.pathStates[segment.ToKey()] = PathStateTooFar
					}
				} else {
					s.pathStates[segment.ToKey()] = PathStateNotUseable
				}
			}
		}
	}
}

func (s *ShortestPathSearch) BuildPath(destination b6.PointID) []b6.Segment {
	segments := make([]b6.Segment, 0, 16)
	point := destination
	for {
		if r, ok := s.byPoint[point]; ok && r.segment != b6.SegmentInvalid {
			segments = append(segments, r.segment)
			point = r.segment.FirstFeature().PointID()
		} else {
			break
		}
	}

	for i := 0; i < len(segments)/2; i++ {
		j := len(segments) - 1 - i
		segments[i], segments[j] = segments[j], segments[i]
	}
	return segments
}

func (s *ShortestPathSearch) PointDistances() map[b6.PointID]float64 {
	distances := make(map[b6.PointID]float64, len(s.byPoint))
	for id, r := range s.byPoint {
		distances[id] = r.distance
	}
	return distances
}

func (s *ShortestPathSearch) AreaDistances() map[b6.AreaID]float64 {
	distances := make(map[b6.AreaID]float64, len(s.byArea))
	for id, r := range s.byArea {
		distances[id] = r.distance
	}
	return distances
}

func (s *ShortestPathSearch) AreaEntrances() map[b6.AreaID]b6.PointID {
	entrances := make(map[b6.AreaID]b6.PointID, len(s.byArea))
	for id, r := range s.byArea {
		entrances[id] = r.point
	}
	return entrances
}

func (s *ShortestPathSearch) FillCountsAndDistancesFromPaths(counts map[b6.SegmentKey]int, distances map[b6.PointID]float64) {
	for id, _ := range s.byPoint {
		for _, segment := range s.BuildPath(id) {
			first, last := segment.First, segment.Last
			// Deliberately throw away direction
			if last < first {
				first, last = last, first
			}
			key := b6.SegmentKey{ID: segment.Feature.PathID(), First: first, Last: last}
			if count, ok := counts[key]; ok {
				counts[key] = count + 1
			} else {
				counts[key] = 1
			}
			// TODO: We're confusing weights with physical distances here. It should be made more generic,
			// maybe by introducting Weights to calculate the distance along the segment when
			// interpolating
			firstDistance, lastDistance := s1.InfAngle(), s1.InfAngle()
			if d, ok := distances[segment.FirstFeature().PointID()]; ok {
				firstDistance = b6.MetersToAngle(d)
			}
			if d, ok := distances[segment.LastFeature().PointID()]; ok {
				lastDistance = b6.MetersToAngle(d)
			}
			ds := interpolateShortestPathDistances(segment, firstDistance, lastDistance)
			for i := 0; i < segment.Len(); i++ {
				distances[segment.SegmentFeature(i).PointID()] = b6.AngleToMeters(ds[i])
			}
		}
	}
}

func (s *ShortestPathSearch) PathStates() map[b6.SegmentKey]PathState {
	return s.pathStates
}

func ComputeShortestPath(from b6.PointID, to b6.PointID, maxDistance float64, weights Weights, w b6.World) []b6.Segment {
	s := NewShortestPathSearchFromPoint(from)
	s.ExpandSearchTo(to, maxDistance, weights, w)
	return s.BuildPath(to)
}

func ComputeAccessibility(from b6.PointID, maxDistance float64, weights Weights, w b6.World) (map[b6.PointID]float64, map[b6.SegmentKey]int) {
	s := NewShortestPathSearchFromPoint(from)
	s.ExpandSearch(maxDistance, weights, Points, w)
	// TODO: Rework this API: We shouldn't have to do PointDistances() and the Fill()
	distances := s.PointDistances()
	counts := make(map[b6.SegmentKey]int, len(distances))
	s.FillCountsAndDistancesFromPaths(counts, distances)
	return distances, counts
}
