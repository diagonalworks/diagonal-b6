package osm

import (
	math "math"

	"github.com/golang/geo/s2"
)

type LatLng struct {
	Lat float64
	Lng float64
}

func (ll *LatLng) ToS2LatLng() s2.LatLng {
	return s2.LatLngFromDegrees(ll.Lat, ll.Lng)
}

func (ll *LatLng) ToS2CellID() s2.CellID {
	return s2.CellIDFromLatLng(s2.LatLngFromDegrees(ll.Lat, ll.Lng))
}

func (ll *LatLng) ToS2Point() s2.Point {
	return s2.PointFromLatLng(ll.ToS2LatLng())
}

func FromS2LatLng(ll s2.LatLng) LatLng {
	return LatLng{Lat: ll.Lat.Degrees(), Lng: ll.Lng.Degrees()}
}

func FromS2CellID(cell s2.CellID) LatLng {
	return FromS2LatLng(cell.LatLng())
}

func FromS2Point(point s2.Point) LatLng {
	return FromS2LatLng(s2.LatLngFromPoint(point))
}

type Tag struct {
	Key   string
	Value string
}

type Tags []Tag

func (t Tags) GetTags() []Tag {
	return []Tag(t)
}

func (t Tags) HasTags() bool {
	return len(t) > 0
}

func (t Tags) HasTag(key string) bool {
	_, ok := t.Tag(key)
	return ok
}

func (t Tags) Tag(key string) (string, bool) {
	for _, tag := range t {
		if tag.Key == key {
			return tag.Value, true
		}
	}
	return "", false
}

func (t Tags) TagDefault(key string, d string) string {
	value, ok := t.Tag(key)
	if ok {
		return value
	}
	return d
}

func (t *Tags) AddTag(key string, value string) {
	*t = append(*t, Tag{Key: key, Value: value})
}

func (t *Tags) Clone() Tags {
	tags := make(Tags, len(*t))
	copy(tags, *t)
	return tags
}

type AnyID int64

type Element interface {
	GetID() AnyID
	GetTags() []Tag
	Tag(key string) (string, bool)
	HasTags() bool
	AddTag(key string, value string)
}

type NodeID int64

type NodeIDs []NodeID

func (n NodeIDs) Len() int           { return len(n) }
func (n NodeIDs) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n NodeIDs) Less(i, j int) bool { return n[i] < n[j] }

type Node struct {
	ID       NodeID
	Location LatLng
	Tags
}

func (n *Node) GetID() AnyID {
	return AnyID(n.ID)
}

func (n *Node) Clone() Node {
	return Node{ID: n.ID, Location: n.Location, Tags: n.Tags.Clone()}
}

type NodeIDSet map[NodeID]struct{}

func (n NodeIDSet) Add(id NodeID) {
	n[id] = struct{}{}
}

func (n NodeIDSet) Has(id NodeID) bool {
	_, ok := n[id]
	return ok
}

type WayID int64

type Way struct {
	ID    WayID
	Nodes []NodeID
	Tags
}

func (w *Way) GetID() AnyID {
	return AnyID(w.ID)
}

func (w *Way) FirstNode() NodeID {
	return w.Nodes[0]
}

func (w *Way) LastNode() NodeID {
	return w.Nodes[len(w.Nodes)-1]
}

func (w *Way) Clone() Way {
	nodes := make([]NodeID, len(w.Nodes))
	copy(nodes, w.Nodes)
	return Way{ID: w.ID, Nodes: nodes, Tags: w.Tags.Clone()}
}

type ElementType byte

const (
	ElementTypeNode ElementType = iota
	ElementTypeWay
	ElementTypeRelation
)

type Member struct {
	Type ElementType
	ID   AnyID
	Role string
}

func (m *Member) NodeID() NodeID {
	if m.Type == ElementTypeNode {
		return NodeID(m.ID)
	}
	panic("Not a node")
}

func (m *Member) WayID() WayID {
	if m.Type == ElementTypeWay {
		return WayID(m.ID)
	}
	panic("Not a way")
}

func (m *Member) RelationID() RelationID {
	if m.Type == ElementTypeRelation {
		return RelationID(m.ID)
	}
	panic("Not a relation")
}

type RelationID int64

type Relation struct {
	ID      RelationID
	Members []Member
	Tags
}

func (r *Relation) GetID() AnyID {
	return AnyID(r.ID)
}

func (r *Relation) Clone() Relation {
	members := make([]Member, len(r.Members))
	copy(members, r.Members)
	return Relation{ID: r.ID, Members: members, Tags: r.Tags.Clone()}
}

type RelationIDSet map[RelationID]struct{}

func (r RelationIDSet) Add(id RelationID) {
	r[id] = struct{}{}
}

func (r RelationIDSet) Has(id RelationID) bool {
	_, ok := r[id]
	return ok
}

type NodeLocations interface {
	FindLocation(id NodeID) (s2.Point, bool)
}

type LocationMap map[NodeID]s2.Point

func (l LocationMap) FindLocation(id NodeID) (s2.Point, bool) {
	point, ok := l[id]
	return point, ok
}

type Ways interface {
	FindWay(id WayID) (Way, bool)
}

type WayMap map[WayID]Way

func (w WayMap) FindWay(id WayID) (Way, bool) {
	way, ok := w[id]
	return way, ok
}

type RelationSlice []*Relation

func (r RelationSlice) Len() int           { return len(r) }
func (r RelationSlice) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r RelationSlice) Less(i, j int) bool { return r[i].ID < r[j].ID }

type Relations interface {
	FindRelation(id RelationID) *Relation
}

type InvertedRelations interface {
	FindRelationsForWay(way WayID) []*Relation
}

type RelationMap map[RelationID]*Relation

func (r RelationMap) FindRelation(id RelationID) *Relation {
	if relation, ok := r[id]; ok {
		return relation
	}
	return nil
}

type GeometryState int

const (
	GeometryStateOK GeometryState = iota
	GeometryStateReversed
	GeometryStateMissingNodes
	GeometryStateInvalid
)

// Return the region representing a way, together with whether the way needed to
// be reversed to be counter-clockwise. Should move into geo/osm, if we have an interface to find nodes
func WayToRegion(way *Way, locations NodeLocations) (s2.Region, GeometryState) {
	if way.Nodes[0] != way.Nodes[len(way.Nodes)-1] {
		polyline := make(s2.Polyline, len(way.Nodes))
		for i, id := range way.Nodes {
			if location, ok := locations.FindLocation(id); ok {
				polyline[i] = location
			} else {
				return nil, GeometryStateMissingNodes
			}
		}
		return &polyline, GeometryStateOK
	} else {
		points := make([]s2.Point, len(way.Nodes)-1)
		for i := 0; i < len(way.Nodes)-1; i++ {
			if location, ok := locations.FindLocation(way.Nodes[i]); ok {
				points[i] = location
			} else {
				return nil, GeometryStateMissingNodes
			}
		}
		loop := s2.LoopFromPoints(points)
		if err := loop.Validate(); err != nil {
			return nil, GeometryStateInvalid
		}
		state := GeometryStateOK
		// Ensure loop vertices are ordered counterclockwise
		if loop.Area() > 2.0*math.Pi {
			loop.Invert()
			state = GeometryStateReversed
		}
		return loop, state
	}
}

// ClipEdges returns both vertices of every edge (partially) within bounds.
// TODO: Why do we need to add IsFull() here to prevent crashes??
func ClipEdges(lls []s2.LatLng, bounds *s2.Polygon) []s2.LatLng {
	if len(lls) < 2 {
		return []s2.LatLng{}
	} else if bounds.IsFull() {
		return lls
	}
	clipped := make([]s2.LatLng, 0, len(lls))
	previous := bounds.ContainsPoint(s2.PointFromLatLng(lls[len(lls)-1]))
	this := bounds.ContainsPoint(s2.PointFromLatLng(lls[0]))

	for i, ll := range lls {
		next := bounds.ContainsPoint(s2.PointFromLatLng(lls[(i+1)%len(lls)]))
		if previous || this || next {
			clipped = append(clipped, ll)
		}
		previous = this
		this = next
	}
	return clipped
}

func IsPolygonSelfIntersecting(polygon *s2.Polygon) bool {
	if polygon == nil {
		return false
	}
	edges := make([]s2.Edge, 0, 0)
	for _, loop := range polygon.Loops() {
		for i := 0; i < loop.NumEdges(); i++ {
			edges = append(edges, loop.Edge(i))
		}
	}
	for i := range edges {
		crosser := s2.NewEdgeCrosser(edges[i].V0, edges[i].V1)
		for j := i + 1; j < len(edges); j++ {
			if crosser.CrossingSign(edges[j].V0, edges[j].V1) == s2.Cross {
				return true
			}
		}
	}
	return false
}
