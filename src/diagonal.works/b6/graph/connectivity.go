package graph

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sort"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/ingest"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

type ConnectResult int

const (
	ConnectResultAlreadyConnected ConnectResult = iota
	ConnectResultConnected
	ConnectResultDisconnected
	ConnectResultImpossible
)

type PathIDSet map[b6.PathID]struct{}

func (p PathIDSet) Contains(id b6.PathID) bool {
	_, connected := p[id]
	return connected
}

// BuildStreetNetwork returns the IDs of paths classified as being part of the street network.
// A path is classed as being part of the network if it's possible to traverse more than
// the given threshold distance away from the start of the path, using the paths allowed
// by the given Weights.
func BuildStreetNetwork(paths b6.PathFeatures, threshold s1.Angle, weights Weights, g *geojson.FeatureCollection, w b6.World) PathIDSet {
	network := make(PathIDSet)
	stack := make([]b6.PhysicalFeature, 0, 2)
	for paths.Next() {
		path := paths.Feature()
		if network.Contains(path.PathID()) || !weights.IsUseable(b6.ToSegment(path)) {
			continue
		}
		stack = stack[0:0]
		seen := make(map[b6.SegmentKey]struct{})
		first := path.Feature(0)
		if first == nil {
			continue
		}
		segments := w.Traverse(first.FeatureID())
		var origin s2.Point
		for segments.Next() {
			segment := segments.Segment()
			if segment.Feature.FeatureID() == path.FeatureID() {
				seen[segment.ToKey()] = struct{}{}
				if first := segment.FirstFeature(); first != nil {
					origin = s2.PointFromLatLng(first.Location())
					if last := segment.LastFeature(); last != nil {
						stack = append(stack, last)
						break
					}
				}
			}
		}

		connected := false
		for len(stack) > 0 {
			point := stack[len(stack)-1]
			stack = stack[0 : len(stack)-1]
			if origin.Distance(s2.PointFromLatLng(point.Location())) > threshold {
				connected = true
			} else {
				segments := w.Traverse(point.FeatureID())
				for segments.Next() {
					segment := segments.Segment()
					if !weights.IsUseable(segment) {
						continue
					}
					if _, ok := seen[segment.ToKey()]; !ok {
						if network.Contains(segment.Feature.PathID()) {
							connected = true
							break
						} else {
							if last := segment.LastFeature(); last != nil {
								stack = append(stack, segment.LastFeature())
							}
							seen[segment.ToKey()] = struct{}{}
						}
					}
				}
			}
			if connected {
				for key := range seen {
					network[key.ID] = struct{}{}
				}
				break
			}
		}
	}

	if g != nil {
		for id := range network {
			path := b6.FindPathByID(id, w)
			polyline := path.Polyline()
			shape := geojson.NewFeatureFromS2Polyline(*polyline)
			shape.Properties["colour"] = "#00ff00"
			g.AddFeature(shape)
		}
	}
	return network
}

func IsPointConnected(id b6.FeatureID, network PathIDSet, w b6.World) bool {
	paths := w.FindPathsByPoint(id)
	for paths.Next() {
		if _, ok := network[paths.FeatureID().ToPathID()]; ok {
			return true
		}
	}
	return false
}

func pathDistances(path b6.PathFeature, distances []s1.Angle) []s1.Angle {
	distances = append(distances[0:0], s1.Angle(0))
	previous := path.Point(0)
	for i := 1; i < path.Len(); i++ {
		p := path.Point(i)
		distances = append(distances, distances[i-1]+previous.Distance(p))
		previous = p
	}
	return distances
}

func hash(a uint64, b uint64) uint64 {
	h := fnv.New64()
	var buffer [8]byte
	binary.LittleEndian.PutUint64(buffer[0:], a)
	h.Write(buffer[0:])
	binary.LittleEndian.PutUint64(buffer[0:], b)
	h.Write(buffer[0:])
	return h.Sum64()
}

func hashIDs(points [2]b6.FeatureID) uint64 {
	h := fnv.New64()
	var buffer [8]byte
	for _, p := range points {
		h.Write([]byte(p.Namespace))
		binary.LittleEndian.PutUint64(buffer[0:], p.Value)
		h.Write(buffer[0:])
	}
	return h.Sum64()
}

func entranceID(f b6.FeatureID, entrance int) b6.FeatureID {
	ns := b6.NamespaceDiagonalEntrances.String() + "/" + f.Namespace.String()
	return b6.FeatureID{b6.FeatureTypePoint, b6.Namespace(ns), hash(f.Value, uint64(entrance))}
}

func accessID(id b6.FeatureID, entrance int) b6.FeatureID {
	ns := b6.NamespaceDiagonalAccessPoints.String() + "/" + id.Namespace.String()
	return b6.FeatureID{b6.FeatureTypePoint, b6.Namespace(ns), hash(id.Value, uint64(entrance))}
}

type insertion struct {
	PathID   b6.PathID
	Distance s1.Angle
	PointID  b6.FeatureID
}

type insertions []insertion

func (is insertions) Len() int      { return len(is) }
func (is insertions) Swap(i, j int) { is[i], is[j] = is[j], is[i] }
func (is insertions) Less(i, j int) bool {
	if is[i].PathID == is[j].PathID {
		return is[i].Distance < is[j].Distance
	}
	return is[i].PathID.Less(is[j].PathID)
}

type extent struct {
	From b6.FeatureID
	To   b6.FeatureID
}

type additions [][2]b6.FeatureID

func (as additions) Len() int      { return len(as) }
func (as additions) Swap(i, j int) { as[i], as[j] = as[j], as[i] }
func (as additions) Less(i, j int) bool {
	if as[i][0] == as[j][0] {
		return as[i][1].Less(as[j][1])
	}
	return as[i][0].Less(as[j][0])
}

type Connections struct {
	insertions insertions
	additions  additions
	clustered  map[b6.FeatureID]b6.FeatureID
	lock       sync.Mutex
}

func NewConnections() *Connections {
	return &Connections{clustered: make(map[b6.FeatureID]b6.FeatureID)}
}

func (c *Connections) String() string {
	return fmt.Sprintf("%d insertions, %d additions, %d clustered", len(c.insertions), len(c.additions), len(c.clustered))
}

func (c *Connections) InsertPoint(path b6.PathID, distance s1.Angle, point b6.FeatureID) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.insertions = append(c.insertions, insertion{PathID: path, Distance: distance, PointID: point})
}

func (c *Connections) AddPath(from b6.FeatureID, to b6.FeatureID) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.additions = append(c.additions, [2]b6.FeatureID{from, to})
}

func (c *Connections) Cluster(threshold s1.Angle, w b6.World) {
	sort.Sort(c.insertions)
	c.clusterCloseInsertions(threshold)
	c.clusterInsertionsOntoExistingPoints(threshold, w)
	for i := range c.additions {
		for j := 0; j < 2; j++ {
			for {
				if next, ok := c.clustered[c.additions[i][j]]; ok {
					c.additions[i][j] = next
				} else {
					break
				}
			}
		}
	}
	sort.Sort(c.additions)
}

func (c *Connections) clusterCloseInsertions(threshold s1.Angle) {
	id := b6.PathIDInvalid
	d := s1.InfAngle()
	last := -1
	for i, insertion := range c.insertions {
		if insertion.PathID != id {
			id = insertion.PathID
			d = insertion.Distance
			last = i
		} else {
			if insertion.Distance-d < threshold {
				d = (d + insertion.Distance) / 2.0
				c.insertions[last].Distance = d
				c.clustered[insertion.PointID] = c.insertions[last].PointID
				c.insertions[i].Distance = s1.InfAngle()
			} else {
				d = insertion.Distance
				last = i
			}
		}
	}
}

func (c *Connections) clusterInsertionsOntoExistingPoints(threshold s1.Angle, w b6.World) {
	id := b6.PathIDInvalid
	var path b6.PathFeature
	var distances []s1.Angle
	p := 0
	for i, insertion := range c.insertions {
		if insertion.PathID != id {
			id = insertion.PathID
			path = b6.FindPathByID(id, w)
			if path == nil {
				continue
			}
			distances = pathDistances(path, distances)
			p = 0
		}
		if insertion.Distance == s1.InfAngle() { // Already clustered
			continue
		}
		for p < len(distances) && distances[p] <= insertion.Distance {
			p++
		}
		previous := insertion.Distance - distances[p-1]
		next := s1.InfAngle()
		if p < len(distances) {
			next = distances[p] - insertion.Distance
		}
		if previous < next {
			if previous < threshold {
				c.clustered[insertion.PointID] = path.Feature(p - 1).FeatureID()
				c.insertions[i].Distance = s1.InfAngle()
			}
		} else {
			if next < threshold {
				c.clustered[insertion.PointID] = path.Feature(p).FeatureID()
				c.insertions[i].Distance = s1.InfAngle()
			}
		}
	}
}

func (c *Connections) ApplyToPath(path b6.PathFeature) *ingest.PathFeature {
	id := path.PathID()
	i := sort.Search(len(c.insertions), func(j int) bool {
		return !c.insertions[j].PathID.Less(id)
	})
	if i >= len(c.insertions) {
		return ingest.NewPathFeatureFromWorld(path)
	}
	distances := pathDistances(path, []s1.Angle{})
	ids := make([]b6.FeatureID, 0, path.Len())
	points := make([]s2.Point, 0, path.Len())
	next := 0
	for i < len(c.insertions) && c.insertions[i].PathID == path.PathID() {
		if c.insertions[i].Distance != s1.InfAngle() {
			for distances[next] < c.insertions[i].Distance {
				ids = append(ids, path.Feature(next).FeatureID())
				points = append(points, path.Point(next))
				next++
			}
			ids = append(ids, c.insertions[i].PointID)
			points = append(points, s2.Point{})
		}
		i++
	}
	for next < path.Len() {
		ids = append(ids, path.Feature(next).FeatureID())
		points = append(points, path.Point(next))
		next++
	}
	applied := ingest.NewPathFeature(len(ids))
	applied.PathID = path.PathID()
	applied.Tags = ingest.NewTagsFromWorld(path)
	for i, p := range ids {
		if p != b6.FeatureIDInvalid {
			applied.SetPointID(i, p)
		} else {
			applied.SetLatLng(i, s2.LatLngFromPoint(points[i]))
		}
	}
	return applied
}

func (c *Connections) EachInsertedPoint(f func(id b6.FeatureID, ll s2.LatLng) error, w b6.World) error {
	id := b6.PathIDInvalid
	var polyline *s2.Polyline
	length := s1.InfAngle()
	for _, insertion := range c.insertions {
		if insertion.PathID != id {
			id = insertion.PathID
			path := b6.FindPathByID(id, w)
			if path == nil {
				continue
			}
			polyline = path.Polyline()
			length = polyline.Length()
		}
		if insertion.Distance == s1.InfAngle() { // Clustered
			continue
		}
		p, _ := polyline.Interpolate(float64(insertion.Distance / length))
		if err := f(insertion.PointID, s2.LatLngFromPoint(p)); err != nil {
			return err
		}
	}
	return nil
}

func (c *Connections) EachAddedPath(emit ingest.Emit) error {
	last := [2]b6.FeatureID{b6.FeatureIDInvalid, b6.FeatureIDInvalid}
	path := ingest.NewPathFeature(2)
	path.Tags.AddTag(b6.Tag{Key: "diagonal", Value: b6.String("connection")})
	for _, a := range c.additions {
		if a != last {
			last = a
			path.PathID = b6.MakePathID(b6.NamespaceDiagonalAccessPaths, hashIDs(a))
			path.SetPointID(0, a[0])
			path.SetPointID(1, a[1])
			if err := emit(path, 0); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Connections) Change(w b6.World) ingest.Change {
	change := &ingest.AddFeatures{}
	id := b6.PathIDInvalid
	for _, insertion := range c.insertions {
		if insertion.PathID != id {
			id = insertion.PathID
			if existing := b6.FindPathByID(id, w); existing != nil {
				*change = append(*change, c.ApplyToPath(existing))
			}
		}
	}
	f := func(id b6.FeatureID, ll s2.LatLng) error {
		*change = append(*change, &ingest.GenericFeature{ID: id.FeatureID(), Tags: []b6.Tag{{Key: b6.LatLngTag, Value: b6.LatLng(ll)}}})
		return nil
	}
	c.EachInsertedPoint(f, w)
	ff := func(f ingest.Feature, _ int) error {
		*change = append(*change, f.(*ingest.PathFeature).ClonePathFeature())
		return nil
	}
	c.EachAddedPath(ff)
	return change
}

type modifyWorldSource struct {
	World       b6.World
	Connections *Connections
}

func (m *modifyWorldSource) Read(options ingest.ReadOptions, emit ingest.Emit, ctx context.Context) error {
	o := b6.EachFeatureOptions{
		SkipPoints:    options.SkipPoints,
		SkipPaths:     options.SkipPaths,
		SkipAreas:     options.SkipAreas,
		SkipRelations: options.SkipRelations,
		Goroutines:    options.Goroutines,
	}
	each := func(f b6.Feature, goroutine int) error {
		switch f := f.(type) {
		case b6.PathFeature:
			return emit(m.Connections.ApplyToPath(f), goroutine)
		default:
			return emit(ingest.NewFeatureFromWorld(f), goroutine)
		}
	}
	if err := m.World.EachFeature(each, &o); err != nil {
		return nil
	}
	if !o.SkipPoints {
		var point ingest.Feature
		f := func(id b6.FeatureID, ll s2.LatLng) error {
			point = &ingest.GenericFeature{ID: id.FeatureID(), Tags: []b6.Tag{{Key: b6.LatLngTag, Value: b6.LatLng(ll)}}}
			return emit(point, 0)
		}
		if err := m.Connections.EachInsertedPoint(f, m.World); err != nil {
			return err
		}
	}
	if !o.SkipPaths {
		if err := m.Connections.EachAddedPath(emit); err != nil {
			return err
		}
	}
	return nil
}

func (c *Connections) ModifyWorld(w b6.World) ingest.FeatureSource {
	return &modifyWorldSource{World: w, Connections: c}
}

type addedPathsSource struct {
	Connections *Connections
}

func (a *addedPathsSource) Read(options ingest.ReadOptions, emit ingest.Emit, ctx context.Context) error {
	if !options.SkipPaths {
		if err := a.Connections.EachAddedPath(emit); err != nil {
			return err
		}
	}
	return nil
}

func (c *Connections) AddedPaths() ingest.FeatureSource {
	return &addedPathsSource{Connections: c}
}

type candidate struct {
	Feature  b6.PathFeature
	Polyline *s2.Polyline
}

type Projection struct {
	Feature    b6.PathFeature
	Polyline   *s2.Polyline
	Point      s2.Point
	Distance   s1.Angle
	NextVertex int
}

func newProjection(point s2.Point, path b6.PathFeature, polyline *s2.Polyline) *Projection {
	p, v := polyline.Project(point)
	return &Projection{
		Feature:    path,
		Polyline:   polyline,
		Point:      p,
		Distance:   point.Distance(p),
		NextVertex: v,
	}
}

func (p *Projection) InsertPoint(id b6.FeatureID, connections *Connections) {
	begin := s2.Polyline((*p.Polyline)[0 : p.NextVertex-1])
	distance := begin.Length() + (*p.Polyline)[p.NextVertex-1].Distance(p.Point)
	connections.InsertPoint(p.Feature.PathID(), distance, id)
}

func (p *Projection) Points() (s2.Point, s2.Point) {
	before, after := p.Indices()
	return p.Feature.Point(before), p.Feature.Point(after)
}

func (p *Projection) PointIDs() (b6.FeatureID, b6.FeatureID) {
	// TODO: The projection code doesn't handle the case of polygons with
	// some points that have IDs, and some that don't. Improve when
	// necessary
	before, after := p.Indices()
	return p.Feature.Feature(before).FeatureID(), p.Feature.Feature(after).FeatureID()
}

func (p *Projection) Indices() (int, int) {
	if p.NextVertex >= p.Feature.Len()-1 {
		return p.Feature.Len() - 2, p.Feature.Len() - 1
	}
	if p.NextVertex == 0 {
		return 0, 1
	}
	return p.NextVertex - 1, p.NextVertex
}

func closestCandidate(point s2.Point, candidates []candidate) *Projection {
	p := &Projection{Distance: s1.InfAngle()}
	for _, c := range candidates {
		if pp := newProjection(point, c.Feature, c.Polyline); pp.Distance < p.Distance {
			p = pp
		}
	}
	return p
}

func isAreaConnected(area b6.AreaFeature, i int, network PathIDSet, w b6.World) ConnectResult {
	if paths := area.Feature(i); paths != nil {
		for j := 0; j < paths[0].Len(); j++ {
			if point := paths[0].Feature(j); point != nil {
				if IsPointConnected(point.FeatureID(), network, w) {
					return ConnectResultAlreadyConnected
				}
			}
		}
		return ConnectResultDisconnected
	} else {
		return ConnectResultImpossible
	}

}

func ConnectArea(area b6.AreaFeature, network PathIDSet, threshold s1.Angle, w b6.World, s ConnectionStrategy) {
	n := 0
	for i := 0; i < area.Len(); i++ {
		if isAreaConnected(area, i, network, w) != ConnectResultDisconnected {
			continue
		}
		boundary := area.Feature(i)[0]
		entrances := make([]b6.PhysicalFeature, 0)
		for j := 0; j < boundary.Len(); j++ {
			point := boundary.Feature(j)
			if entrance := point.Get("entrance"); entrance.IsValid() {
				entrances = append(entrances, point)
			}
		}
		cap := area.Polygon(i).CapBound().Expanded(threshold)
		highways := b6.FindPaths(b6.Intersection{b6.Keyed{"#highway"}, b6.MightIntersect{cap}}, w)
		candidates := make([]candidate, 0, 16)
		for highways.Next() {
			if _, ok := network[highways.FeatureID().ToPathID()]; ok {
				f := highways.Feature()
				candidates = append(candidates, candidate{Feature: f, Polyline: f.Polyline()})
			}
		}
		if len(entrances) == 0 {
			// Check all building sides
			access := &Projection{Distance: s1.InfAngle()}
			entrance := &Projection{Feature: boundary, Polyline: boundary.Polyline()}
			for j := 0; j < boundary.Len()-1; j++ {
				m := s2.Interpolate(0.5, boundary.Point(j), boundary.Point(j+1))
				if p := closestCandidate(m, candidates); p.Distance < access.Distance {
					access = p
					entrance.Distance = p.Distance
					entrance.NextVertex = j + 1
				}
			}
			if access.Distance < threshold {
				s.ConnectProjection(area.FeatureID(), entrance, access, n)
				n++
			}
		} else {
			for _, entrance := range entrances {
				if access := closestCandidate(s2.PointFromLatLng(entrance.Location()), candidates); access.Distance < threshold {
					s.ConnectPoint(area.FeatureID(), entrance.FeatureID(), access, n)
					n++
				}
			}
		}
	}
}

func ConnectPoint(point b6.PhysicalFeature, network PathIDSet, threshold s1.Angle, w b6.World, s ConnectionStrategy) {
	if IsPointConnected(point.FeatureID(), network, w) {
		return
	}
	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(point.Location()), threshold)
	highways := b6.FindPaths(b6.Intersection{b6.Keyed{"#highway"}, b6.MightIntersect{cap}}, w)
	candidates := make([]candidate, 0, 16)
	for highways.Next() {
		if _, ok := network[highways.FeatureID().ToPathID()]; ok {
			f := highways.Feature()
			candidates = append(candidates, candidate{Feature: f, Polyline: f.Polyline()})
		}
	}
	if access := closestCandidate(s2.PointFromLatLng(point.Location()), candidates); access.Distance < threshold {
		s.ConnectPoint(point.FeatureID(), point.FeatureID(), access, 0)
	}
}

func ConnectFeature(f b6.Feature, network PathIDSet, threshold s1.Angle, w b6.World, s ConnectionStrategy) {
	if f, ok := f.(b6.PhysicalFeature); ok {
		switch f.GeometryType() {
		case b6.GeometryTypePoint:
			ConnectPoint(f, network, threshold, w, s)
		case b6.GeometryTypeArea:
			ConnectArea(f.(b6.AreaFeature), network, threshold, w, s)
		}
	}
}

type ConnectionStrategy interface {
	ConnectProjection(f b6.FeatureID, entrance *Projection, access *Projection, n int)
	ConnectPoint(f b6.FeatureID, entrance b6.FeatureID, access *Projection, n int)
	Finish()
	Output() ingest.FeatureSource
}

type InsertNewPointsIntoPaths struct {
	Connections      *Connections
	World            b6.World
	ClusterThreshold s1.Angle
}

func (s InsertNewPointsIntoPaths) ConnectProjection(f b6.FeatureID, entrance *Projection, access *Projection, n int) {
	from := entranceID(f, n)
	to := accessID(f, n)
	entrance.InsertPoint(from, s.Connections)
	access.InsertPoint(to, s.Connections)
	s.Connections.AddPath(from, to)
}

func (s InsertNewPointsIntoPaths) ConnectPoint(f b6.FeatureID, entrance b6.FeatureID, access *Projection, n int) {
	to := accessID(f, n)
	access.InsertPoint(to, s.Connections)
	s.Connections.AddPath(entrance, to)
}

func (s InsertNewPointsIntoPaths) Finish() {
	s.Connections.Cluster(s.ClusterThreshold, s.World)
}

func (s InsertNewPointsIntoPaths) Output() ingest.FeatureSource {
	return s.Connections.ModifyWorld(s.World)
}

type UseExisitingPoints struct {
	Connections *Connections
}

func (s UseExisitingPoints) ConnectProjection(f b6.FeatureID, entrance *Projection, access *Projection, n int) {
	be, ae := entrance.Points()
	ba, aa := access.Points()
	// Create two access paths, one to each of the points before and after the
	// access projection, choosing the entrance point that minimises the
	// distance of addedd paths
	var from b6.FeatureID
	if be.Distance(ba)+be.Distance(aa) < ae.Distance(ba)+aa.Distance(aa) {
		from, _ = entrance.PointIDs()
	} else {
		_, from = entrance.PointIDs()
	}
	bid, aid := access.PointIDs()
	s.Connections.AddPath(from, bid)
	s.Connections.AddPath(from, aid)
}

func (s UseExisitingPoints) ConnectPoint(f b6.FeatureID, entrance b6.FeatureID, access *Projection, n int) {
	bid, aid := access.PointIDs()
	s.Connections.AddPath(entrance, bid)
	s.Connections.AddPath(entrance, aid)
}

func (s UseExisitingPoints) Finish() {}

func (s UseExisitingPoints) Output() ingest.FeatureSource {
	return s.Connections.AddedPaths()
}
