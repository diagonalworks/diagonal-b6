package osm

import (
	"container/heap"
	"fmt"
	"math"

	"github.com/golang/geo/s2"
)

func RelationToPolygon(relation *Relation, nodes NodeLocations, ways Ways) (*s2.Polygon, error) {
	waysByNode, err := indexWaysByEndNodes(relation, nodes, ways)
	if err != nil {
		return nil, err
	}

	loops, err := groupWaysIntoLoops(relation, waysByNode, ways)
	if err != nil {
		return nil, err
	}

	s2loops := make([]*s2.Loop, 0, len(loops))
	for _, loop := range loops {
		s2loop, err := waysToS2Loop(loop, ways, nodes)
		if err != nil {
			return nil, err
		}
		s2loops = append(s2loops, s2loop)
	}
	return s2.PolygonFromLoops(s2loops), nil
}

func indexWaysByEndNodes(relation *Relation, nodes NodeLocations, ways Ways) (map[NodeID][]WayID, error) {
	waysByNode := make(map[NodeID][]WayID)
	for _, member := range relation.Members {
		if member.Type == ElementTypeWay {
			way, ok := ways.FindWay(WayID(member.ID))
			if !ok {
				return nil, fmt.Errorf("Failed to find way %d", member.ID)
			}
			for _, node := range []NodeID{way.FirstNode(), way.LastNode()} {
				ws, ok := waysByNode[node]
				if !ok {
					ws = make([]WayID, 0, 2)
				}
				waysByNode[node] = append(ws, way.ID)
			}
		}
	}
	return waysByNode, nil
}

func groupWaysIntoLoops(relation *Relation, waysByNode map[NodeID][]WayID, ways Ways) ([][]WayID, error) {
	seen := make(map[WayID]struct{})
	loops := make([][]WayID, 0)
	for _, member := range relation.Members {
		if member.Type == ElementTypeWay {
			if _, ok := seen[WayID(member.ID)]; ok {
				continue
			}
			seen[WayID(member.ID)] = struct{}{}
			way, _ := ways.FindWay(WayID(member.ID))
			if way.FirstNode() == way.LastNode() {
				loops = append(loops, []WayID{way.ID})
				continue
			}
			loop := make([]WayID, 0, len(relation.Members))
			jointNode := way.LastNode()
			for {
				loop = append(loop, way.ID)
				found := false
				for _, next := range waysByNode[jointNode] {
					nextWay, _ := ways.FindWay(next)
					if next != way.ID {
						found = true
						if jointNode == nextWay.FirstNode() {
							jointNode = nextWay.LastNode()
						} else {
							jointNode = nextWay.FirstNode()
						}
						way = nextWay
						break
					}
				}
				if !found {
					return nil, fmt.Errorf("Failed to find next for node %d", way.LastNode())
				}
				if _, ok := seen[WayID(way.ID)]; ok {
					loops = append(loops, loop)
					break
				}
				seen[WayID(way.ID)] = struct{}{}
			}
		}
	}
	return loops, nil
}

func waysToS2Loop(ids []WayID, ways Ways, nodes NodeLocations) (*s2.Loop, error) {
	points := make([]s2.Point, 0, 3)
	way, _ := ways.FindWay(ids[0])
	jointNode := way.FirstNode()
	// Each way could be joined to the next by either it's first or last node.
	// If it's the first, we add points forwards, if not, we add the backwards.
	for _, id := range ids {
		way, _ = ways.FindWay(id)
		if way.FirstNode() == jointNode { // Add points forwards
			for i := 0; i < len(way.Nodes); i++ {
				point, ok := nodes.FindLocation(way.Nodes[i])
				if !ok {
					return nil, fmt.Errorf("Failed to find location for node %d", way.Nodes[i])
				}
				points = append(points, point)
			}
			jointNode = way.Nodes[len(way.Nodes)-1]
		} else { // Backwards
			for i := len(way.Nodes) - 1; i >= 0; i-- {
				point, ok := nodes.FindLocation(way.Nodes[i])
				if !ok {
					return nil, fmt.Errorf("Failed to find location for node %d", way.Nodes[i])
				}
				points = append(points, point)
			}
			jointNode = way.Nodes[0]
		}
	}
	s2loop := s2.LoopFromPoints(points)
	if s2loop.Area() > 2*math.Pi {
		s2loop.Invert()
	}
	return s2loop, nil
}

// svertex is a doubly-linked list representation of a list of vertices
// TODO: factor out this code with that in clipping.
type svertex struct {
	point    s2.Point
	previous *svertex
	next     *svertex
	// The index of the entry in the heap representing the simplification triange
	// starting at this vertex
	heapIndex int
}

// Delete and return the next vertex from the list
func (s *svertex) deleteNext() *svertex {
	deleted := s.next
	s.next.next.previous = s
	s.next = s.next.next
	return deleted
}

// simplification represents a change that could be made to a polygon:
// the removal of vertex.next. It's used as an entry in a priority queue,
// ordered by area.
type simplification struct {
	area float64
	// The vertex at which this simplification triangle starts. If it's at the
	// top of the heap, vertex.next will be deleted.
	vertex *svertex
}

func (s *simplification) updateArea() {
	points := make([]s2.Point, 3)
	points[0] = s.vertex.point
	points[1] = s.vertex.next.point
	points[2] = s.vertex.next.next.point
	loop := s2.LoopFromPoints(points)
	s.area = loop.Area()
	if s.area > 2*math.Pi {
		loop.Invert()
		s.area = loop.Area()
	}
}

type simplifications []simplification

func (s simplifications) Len() int { return len(s) }

func (s simplifications) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
	if s[i].vertex != nil {
		s[i].vertex.heapIndex = i
	}
	if s[j].vertex != nil {
		s[j].vertex.heapIndex = j
	}
}

func (s simplifications) Less(i, j int) bool {
	return s[i].area < s[j].area
}

func (s *simplifications) Push(e interface{}) {
	new := e.(simplification)
	new.vertex.heapIndex = len(*s)
	*s = append(*s, new)
}

func (s *simplifications) Pop() interface{} {
	old := *s
	n := len(old)
	head := old[n-1]
	*s = old[0 : n-1]
	return head
}

// SimplifyPolygon simplifies all loops of a polygon by calling SimplifyLoop
// idependently. It makes no attempt to detect or correct for the simplification
// causing the loops of the polygon to intersect.
func SimplifyPolygon(polygon *s2.Polygon, maxAreaError float64) *s2.Polygon {
	loops := make([]*s2.Loop, polygon.NumLoops())
	for i, loop := range polygon.Loops() {
		loops[i] = SimplifyLoop(loop, maxAreaError)
	}
	return s2.PolygonFromLoops(loops)
}

// SimplifyLoop implements the Visvalingam line simplification algorithm,
// iteratively removing points from a loop that cause the smallest change to
// its area, stopping when maxAreaError is reached. See:
// https://bost.ocks.org/mike/simplify/
func SimplifyLoop(loop *s2.Loop, maxAreaError float64) *s2.Loop {
	var start *svertex
	var previous *svertex
	s := make(simplifications, loop.NumVertices())
	for i := 0; i < loop.NumVertices(); i++ {
		v := &svertex{point: loop.Vertex(i), previous: previous}
		if previous != nil {
			previous.next = v
		} else {
			start = v
		}
		previous = v
		s[i].vertex = v
		s[i].vertex.heapIndex = i
	}
	start.previous = previous
	previous.next = start

	for i := range s {
		s[i].updateArea()
	}

	heap.Init(&s)
	for len(s) > 3 {
		// Prevent head from changing when we fix the dependent triangle
		if s[0].area > maxAreaError {
			break
		}
		s[0].area = -1.0

		deleted := s[0].vertex.deleteNext()
		if start == deleted {
			start = deleted.previous
		}
		// Resuse the deleted vertex's heap entry for the triangle adjusted
		// by deleting the vertex
		s[deleted.heapIndex].vertex = s[0].vertex
		s[deleted.heapIndex].vertex.heapIndex = deleted.heapIndex
		s[deleted.heapIndex].updateArea()
		s[0].vertex = nil
		heap.Fix(&s, deleted.heapIndex)
		heap.Pop(&s)
	}
	points := make([]s2.Point, 0, len(s))
	for v := start; v.next != start; v = v.next {
		points = append(points, v.point)
	}
	return s2.LoopFromPoints(points)
}
