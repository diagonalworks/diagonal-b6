package osm

import (
	"math"
	"os"
	"testing"

	"diagonal.works/b6/units"
	"github.com/golang/geo/s2"
)

type nodeMap map[NodeID]Node

func (n nodeMap) FindLocation(id NodeID) (s2.Point, bool) {
	if node, ok := n[id]; ok {
		return node.Location.ToS2Point(), true
	}
	return s2.Point{}, false
}

type relationMap map[RelationID]Relation

func readPBF(filename string) (nodeMap, WayMap, relationMap, error) {
	nodes := make(nodeMap)
	ways := make(WayMap)
	relations := make(relationMap)

	input, err := os.Open(filename)
	if err != nil {
		return nodes, ways, relations, err
	}
	defer input.Close()

	f := func(e Element) error {
		switch e := e.(type) {
		case *Node:
			nodes[e.ID] = e.Clone()
		case *Way:
			ways[e.ID] = e.Clone()
		case *Relation:
			relations[e.ID] = e.Clone()
		}
		return nil
	}
	ReadPBF(input, f)
	return nodes, ways, relations, nil
}

// TODO: somehow class this as a larger style test? Load the test data once across all tests?
func TestBoundaryRelationToPolygon(t *testing.T) {
	nodes, ways, relations, err := readPBF("../../../../data/tests/london-boundaries.osm.pbf")
	if err != nil {
		t.Errorf("Failed to read test data: %s", err)
		return
	}

	const londonID RelationID = 65606
	london, ok := relations[londonID]
	if !ok {
		t.Errorf("Failed to find feature for London")
		return
	}

	polygon, err := RelationToPolygon(&london, nodes, ways)
	if err != nil {
		t.Errorf("Expected no error, found %v", err)
		return
	}
	if polygon == nil || polygon.NumLoops() != 2 {
		t.Errorf("Expected a polygon with 2 loops")
		return
	}
	expectedArea := 1500.0 * 1000.0 * 1000.0
	actualArea := units.AreaToMeters2(polygon.Area())
	if math.Abs(actualArea-expectedArea)/expectedArea > 0.1 {
		t.Errorf("Expected area around %fm2, found %fm2", expectedArea, actualArea)
	}
}

type nodes []Node

func (ns nodes) FindLocation(id NodeID) (s2.Point, bool) {
	for _, node := range ns {
		if node.ID == id {
			return node.Location.ToS2Point(), true
		}
	}
	return s2.Point{}, false
}

type ways []Way

func (ws ways) FindWay(id WayID) (Way, bool) {
	for i, way := range ws {
		if way.ID == id {
			return ws[i], true
		}
	}
	return Way{}, false
}

func TestRelationToPolygonWithMultipleWays(t *testing.T) {
	ns := nodes{
		Node{ID: 6327612098, Location: LatLng{Lat: 51.2867601, Lng: -0.1243196}},
		Node{ID: 6327612080, Location: LatLng{Lat: 51.2911891, Lng: -0.1281379}},
		Node{ID: 6327611986, Location: LatLng{Lat: 51.3002878, Lng: -0.1500199}},
		Node{ID: 6327611916, Location: LatLng{Lat: 51.3168747, Lng: -0.1605073}},
	}

	ws := ways{
		Way{
			ID:    675674748,
			Nodes: []NodeID{6327612098, 6327612080, 6327611986},
		},
		Way{
			ID:    675674743,
			Nodes: []NodeID{6327611986, 6327611916, 6327612098},
		},
	}

	relation := Relation{
		ID: 65606,
		Members: []Member{
			{Type: ElementTypeWay, ID: 675674748},
			{Type: ElementTypeWay, ID: 675674743},
		},
	}

	polygon, err := RelationToPolygon(&relation, ns, ws)
	if err != nil || polygon == nil || polygon.NumLoops() != 1 {
		t.Errorf("Expected a polygon with a single loop %v", err)
	}
}

func TestRelationToPolygonWithMultipleWaysThatDoNotClose(t *testing.T) {
	ns := nodes{
		Node{ID: 6327612098, Location: LatLng{Lat: 51.2867601, Lng: -0.1243196}},
		Node{ID: 6327612080, Location: LatLng{Lat: 51.2911891, Lng: -0.1281379}},
		Node{ID: 6327611986, Location: LatLng{Lat: 51.3002878, Lng: -0.1500199}},
		Node{ID: 6327611916, Location: LatLng{Lat: 51.3168747, Lng: -0.1605073}},
	}

	ws := ways{
		Way{
			ID:    675674748,
			Nodes: []NodeID{6327612098, 6327612080, 6327611986},
		},
		Way{
			ID:    675674743,
			Nodes: []NodeID{6327611986, 6327611916},
		},
	}

	relation := Relation{
		ID: 65606,
		Members: []Member{
			{Type: ElementTypeWay, ID: 675674748},
			{Type: ElementTypeWay, ID: 675674743},
		},
	}

	polygon, err := RelationToPolygon(&relation, ns, ws)
	if polygon != nil || err == nil {
		t.Errorf("Expected error, found none")
	}
}

func TestRelationToPolygonWithMissingNodes(t *testing.T) {
	ns := nodes{
		Node{ID: 6327612098, Location: LatLng{Lat: 51.2867601, Lng: -0.1243196}},
		Node{ID: 6327612080, Location: LatLng{Lat: 51.2911891, Lng: -0.1281379}},
		// 6327611986 is missing
		Node{ID: 6327611916, Location: LatLng{Lat: 51.3168747, Lng: -0.1605073}},
	}

	ws := ways{
		Way{
			ID:    675674748,
			Nodes: []NodeID{6327612098, 6327612080, 6327611986},
		},
		Way{
			ID:    675674743,
			Nodes: []NodeID{6327611986, 6327611916, 6327612098},
		},
	}

	relation := Relation{
		ID: 65606,
		Members: []Member{
			{Type: ElementTypeWay, ID: 675674748},
			{Type: ElementTypeWay, ID: 675674743},
		},
	}

	polygon, err := RelationToPolygon(&relation, ns, ws)
	if polygon != nil || err == nil {
		t.Errorf("Expected error, found none")
	}
}

func TestRelationToPolygonWithMissingWays(t *testing.T) {
	ns := nodes{
		Node{ID: 6327612098, Location: LatLng{Lat: 51.2867601, Lng: -0.1243196}},
		Node{ID: 6327612080, Location: LatLng{Lat: 51.2911891, Lng: -0.1281379}},
		Node{ID: 6327611986, Location: LatLng{Lat: 51.3002878, Lng: -0.1500199}},
		Node{ID: 6327611916, Location: LatLng{Lat: 51.3168747, Lng: -0.1605073}},
	}

	ws := ways{
		Way{
			ID:    675674748,
			Nodes: []NodeID{6327612098, 6327612080, 6327611986},
		},
		// 675674743 is missing
	}

	relation := Relation{
		ID: 65606,
		Members: []Member{
			{Type: ElementTypeWay, ID: 675674748},
			{Type: ElementTypeWay, ID: 675674743},
		},
	}

	polygon, err := RelationToPolygon(&relation, ns, ws)
	if polygon != nil || err == nil {
		t.Errorf("Expected error, found none")
	}
}

func TestRelationToPolygonWithASingleOpenWay(t *testing.T) {
	ns := nodes{
		Node{ID: 5378333625, Location: LatLng{Lat: 51.5352195, Lng: -0.1254286}},
		Node{ID: 1715968743, Location: LatLng{Lat: 51.5351547, Lng: -0.1250628}},
		Node{ID: 1715968738, Location: LatLng{Lat: 51.5351015, Lng: -0.1248611}},
	}

	ws := ways{
		Way{
			ID:    642639444,
			Nodes: []NodeID{5378333625, 1715968743, 1715968738},
		},
	}

	relation := Relation{
		ID: 1,
		Members: []Member{
			{Type: ElementTypeWay, ID: AnyID(ws[0].ID)},
		},
	}

	polygon, err := RelationToPolygon(&relation, ns, ws)
	if polygon != nil || err == nil {
		t.Errorf("Expected error, found none")
	}
}

func TestRelationToPolygonWithASingleClosedWay(t *testing.T) {
	ns := nodes{
		Node{ID: 5378333625, Location: LatLng{Lat: 51.5352195, Lng: -0.1254286}},
		Node{ID: 1715968743, Location: LatLng{Lat: 51.5351547, Lng: -0.1250628}},
		Node{ID: 1715968738, Location: LatLng{Lat: 51.5351015, Lng: -0.1248611}},
	}

	ws := ways{
		Way{
			ID:    642639444,
			Nodes: []NodeID{5378333625, 1715968743, 1715968738, 5378333625},
		},
	}

	relation := Relation{
		ID: 1,
		Members: []Member{
			{Type: ElementTypeWay, ID: AnyID(ws[0].ID)},
		},
	}

	polygon, err := RelationToPolygon(&relation, ns, ws)
	if err != nil || polygon == nil || polygon.NumLoops() != 1 {
		t.Errorf("Expected a polygon with a single loop")
	}
}

// TODO: Move to geo? Needs test data too.
func TestSimplifyBoundaryPolygon(t *testing.T) {
	nodes, ways, relations, err := readPBF("../../../../data/tests/london-boundaries.osm.pbf")
	if err != nil {
		t.Errorf("Failed to read test data: %s", err)
		return
	}

	const londonID RelationID = 65606
	london, ok := relations[londonID]
	if !ok {
		t.Errorf("Failed to find feature for London")
		return
	}

	polygon, err := RelationToPolygon(&london, nodes, ways)
	if err != nil {
		t.Errorf("Failed to convert London to polygon: %s", err)
		return
	}

	simplified := SimplifyPolygon(polygon, units.Meters2ToArea(100))
	if polygon == nil {
		t.Errorf("Expected a polygon, found nil")
	}

	if math.Abs(simplified.Area()-polygon.Area())/polygon.Area() > 0.01 {
		t.Errorf("Expected areas to be similar")
	}

	for i := 0; i < polygon.NumLoops(); i++ {
		less := simplified.Loop(i).NumVertices()
		more := polygon.Loop(i).NumVertices()
		if float32(less)/float32(more) > 0.5 {
			t.Errorf("Expected simplified polygon loop %d to have many fewer vertices (%d vs %d)", i, less, more)
		}
	}
}
