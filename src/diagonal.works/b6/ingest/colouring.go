package ingest

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"

	"diagonal.works/b6"
	"github.com/golang/geo/s2"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/coloring"
)

type FeatureNode int

func (f FeatureNode) ID() int64 {
	return int64(f)
}

type FeatureEdge struct {
	from int
	to   int
}

func (f FeatureEdge) From() graph.Node {
	return FeatureNode(f.from)
}

func (f FeatureEdge) ReversedEdge() graph.Edge {
	return FeatureEdge{from: f.to, to: f.from}
}

func (f FeatureEdge) To() graph.Node {
	return FeatureNode(f.to)
}

type AllFeatureNodes struct {
	n int
	i int
}

func (a *AllFeatureNodes) Node() graph.Node {
	return FeatureNode(a.i - 1)
}

func (a *AllFeatureNodes) Next() bool {
	a.i++
	return a.i <= a.n
}

func (a *AllFeatureNodes) Len() int {
	return a.n
}

func (a *AllFeatureNodes) Reset() {
	a.i = 0
}

type FeatureNodes struct {
	nodes []int
	i     int
}

func (f *FeatureNodes) Node() graph.Node {
	return FeatureNode(f.nodes[f.i-1])
}

func (f *FeatureNodes) Next() bool {
	f.i++
	return f.i <= len(f.nodes)
}

func (f *FeatureNodes) Len() int {
	return len(f.nodes)
}

func (f *FeatureNodes) Reset() {
	f.i = 0
}

type FeatureGraph struct {
	byID    map[b6.FeatureID]int
	byIndex []b6.FeatureID
	graph   [][]int
}

func (f *FeatureGraph) Node(id int64) graph.Node {
	return FeatureNode(id)
}

func (f *FeatureGraph) ID(id b6.FeatureID) (int, bool) {
	i, ok := f.byID[id]
	return i, ok
}

func (f *FeatureGraph) Nodes() graph.Nodes {
	return &AllFeatureNodes{n: len(f.byIndex)}
}

func (f *FeatureGraph) From(id int64) graph.Nodes {
	return &FeatureNodes{nodes: f.graph[id]}
}

func (f *FeatureGraph) HasEdgeBetween(xid int64, yid int64) bool {
	if int(xid) >= len(f.graph) {
		return false
	}
	for _, id := range f.graph[xid] {
		if id == int(yid) {
			return true
		}
	}
	return false
}

func (f *FeatureGraph) Edge(uid int64, vid int64) graph.Edge {
	if f.HasEdgeBetween(uid, vid) {
		return FeatureEdge{from: int(uid), to: int(vid)}
	}
	return nil
}

func (f *FeatureGraph) EdgeBetween(xid int64, yid int64) graph.Edge {
	return f.Edge(xid, yid)
}

func (f *FeatureGraph) AddNeighbours(a b6.FeatureID, b b6.FeatureID) {
	f.addNeighbour(a, b)
	f.addNeighbour(b, a)
}

func (f *FeatureGraph) addNeighbour(a b6.FeatureID, b b6.FeatureID) {
	if a == b {
		return
	}
	ai := f.lookupIndex(a)
	bi := f.lookupIndex(b)
	for _, ni := range f.graph[ai] {
		if ni == bi {
			return
		}
	}
	f.graph[ai] = append(f.graph[ai], bi)
}

func (f *FeatureGraph) lookupIndex(id b6.FeatureID) int {
	if f.byID == nil {
		f.byID = make(map[b6.FeatureID]int)
	}
	i, ok := f.byID[id]
	if !ok {
		f.byIndex = append(f.byIndex, id)
		i = len(f.byIndex) - 1
		f.byID[id] = i
	}
	for len(f.graph) <= i {
		f.graph = append(f.graph, nil)
	}
	return i
}

func (f *FeatureGraph) Log() {
	for i, ns := range f.graph {
		if len(ns) > 0 {
			s := fmt.Sprintf("%s: ", f.byIndex[i])
			for j, n := range ns {
				if j > 0 {
					s += fmt.Sprintf(", %s", f.byIndex[n])
				} else {
					s += f.byIndex[n].String()
				}
			}
			log.Printf("%s", s)
		}
	}
}

type AreaColourer struct {
	Cores   int
	colours map[int64]int
}

const colouringS2Level = 21 // Roughly 3m sides

const AreaColourTag = "b6:colour"

func ColourAreas(source FeatureSource, cores int) (FeatureSource, error) {
	var lock sync.Mutex
	idsByCell := make(map[s2.CellID][]b6.FeatureID)
	emit := func(f Feature, goroutine int) error {
		if f, ok := f.(*AreaFeature); ok {
			for i := 0; i < f.Len(); i++ {
				if p, ok := f.Polygon(i); ok {
					for j := 0; j < p.NumLoops(); j++ {
						loop := p.Loop(j)
						for k := 0; k < loop.NumVertices(); k++ {
							cell := s2.CellIDFromLatLng(s2.LatLngFromPoint(loop.Vertex(k))).Parent(colouringS2Level)
							var ids []b6.FeatureID
							var ok bool
							lock.Lock()
							if ids, ok = idsByCell[cell]; !ok {
								ids = make([]b6.FeatureID, 0, 1)
							}
							idsByCell[cell] = append(ids, f.FeatureID())
							lock.Unlock()
						}
					}
				}
			}
		}
		return nil
	}

	options := ReadOptions{
		SkipPoints:      true,
		SkipPaths:       true,
		SkipAreas:       false,
		SkipRelations:   true,
		SkipCollections: true,
		SkipTags:        true,
		Goroutines:      cores,
	}
	if err := source.Read(options, emit, context.Background()); err != nil {
		return nil, err
	}
	var g FeatureGraph
	for _, ids := range idsByCell {
		for i := 0; i < len(ids); i++ {
			for j := i + 1; j < len(ids); j++ {
				g.AddNeighbours(ids[i], ids[j])
			}
		}
	}
	_, colours, err := coloring.Dsatur(&g, nil)
	return &areaColouringSource{
		source:  source,
		graph:   &g,
		colours: colours,
	}, err

}

type areaColouringSource struct {
	source  FeatureSource
	colours map[int64]int
	graph   *FeatureGraph
}

func (a *areaColouringSource) Read(options ReadOptions, emit Emit, ctx context.Context) error {
	addColour := func(f Feature, goroutine int) error {
		if f.FeatureID().Type == b6.FeatureTypeArea {
			if id, ok := a.graph.ID(f.FeatureID()); ok {
				v := b6.NewStringExpression(strconv.Itoa(a.colours[int64(id)]))
				f.AddTag(b6.Tag{Key: AreaColourTag, Value: v})
			} else {
				f.AddTag(b6.Tag{Key: AreaColourTag, Value: b6.NewStringExpression("0")})
			}
		}
		return emit(f, goroutine)
	}
	return a.source.Read(options, addColour, ctx)
}
