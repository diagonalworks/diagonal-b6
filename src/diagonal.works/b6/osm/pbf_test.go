package osm

import (
	"bytes"
	"os"
	"testing"

	"github.com/golang/geo/s2"
)

const GranarySquarePBF = "../../../../data/tests/granary-square.osm.pbf"

func TestParsePBF(t *testing.T) {
	f, err := os.Open(GranarySquarePBF)
	if err != nil {
		t.Errorf("Failed to open test data: %s", err)
		return
	}
	nodes := make(map[NodeID]Node, 0)
	ways := make(map[WayID]Way, 0)
	relations := make(map[RelationID]Relation, 0)
	emit := func(e Element) error {
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
	if err := ReadPBF(f, emit); err != nil {
		t.Errorf("Expected no error from ParsePBF, found: %s", err)
	}
	// TODO: Find some interesting statistics about node, way and relation
	// densities that can be used for aggregate testing, and are relatively
	// stable over the world.
	// Could make it dependent on the geographic extents of the area being
	// processed, which could remove some dependencies.

	tags := 0
	for _, node := range nodes {
		tags += len(node.Tags)
	}
	for _, way := range ways {
		tags += len(way.Tags)
	}
	for _, relation := range relations {
		tags += len(relation.Tags)
	}

	if len(nodes) < 1500 || len(nodes) > 1600 {
		t.Errorf("Unexpected number of nodes: %d", len(nodes))
	}

	if len(ways) < 160 || len(ways) > 180 {
		t.Errorf("Unexpected number of ways: %d", len(ways))
	}

	if len(relations) < 15 || len(relations) > 20 {
		t.Errorf("Unexpected number of relations: %d", len(relations))
	}

	if tags < 1000 || tags > 1100 {
		t.Errorf("Unexpected number of tags: %d", tags)
	}

	granarySquareID := RelationID(5735955)
	if relation, ok := relations[granarySquareID]; ok {
		found := false
		fountainWayID := WayID(167318943)
		for _, member := range relation.Members {
			if member.ID == AnyID(fountainWayID) {
				found = true
				if member.Role != "inner" {
					t.Errorf("Expected role inner, found %q", member.Role)
				}
			}
		}
		if !found {
			t.Errorf("Expected to find way %d as member", fountainWayID)
		}
	} else {
		t.Errorf("Expected to find relation %d", granarySquareID)
	}
}

func TestParsePBFSkippingTags(t *testing.T) {
	f, err := os.Open(GranarySquarePBF)
	if err != nil {
		t.Errorf("Failed to open test data: %s", err)
		return
	}
	emit := func(e Element, g int) error {
		switch e := e.(type) {
		case *Node:
			if len(e.Tags) > 0 {
				t.Errorf("Expected tags on nodes to be skipped")
			}
		case *Way:
			if len(e.Tags) > 0 {
				t.Errorf("Expected tags on ways to be skipped")
			}
		case *Relation:
			if len(e.Tags) > 0 {
				t.Errorf("Expected tags on relations to be skipped")
			}
		}
		return nil
	}
	if err := ReadPBFWithOptions(f, emit, ReadOptions{SkipTags: true}); err != nil {
		t.Errorf("Expected no error from ParsePBF, found: %s", err)
	}
}

func TestWritePBF(t *testing.T) {
	f, err := os.Open(GranarySquarePBF)
	if err != nil {
		t.Errorf("Failed to open test data: %s", err)
		return
	}
	defer f.Close()

	expectedNodes := 0
	expectedWays := 0
	expectedRelations := 0

	var buffer bytes.Buffer
	writer, err := NewWriter(&buffer)
	if err != nil {
		t.Errorf("Unexpected error creating Writer: %s", err)
		return
	}
	rect := s2.EmptyRect()
	emit := func(e Element) error {
		switch e := e.(type) {
		case *Node:
			expectedNodes++
			rect = rect.AddPoint(s2.LatLngFromDegrees(e.Location.Lat, e.Location.Lng))
			return writer.WriteNode(e)
		case *Way:
			expectedWays++
			return writer.WriteWay(e)
		case *Relation:
			expectedRelations++
			return writer.WriteRelation(e)
		}
		return nil
	}
	if err := ReadPBF(f, emit); err != nil {
		t.Errorf("Failed to write PBF: %s", err)
		return
	}
	if err := writer.Flush(); err != nil {
		t.Errorf("writer.Flush() failed: %s", err)
		return
	}

	nodes := 0
	ways := 0
	relations := 0

	// TODO: Do some statistical validation on coordinates by binning them by
	// S2 cell. Maybe the densities are scale free and fit a well known
	// distribution?

	// Slightly increase the size of the rectangle to allow some tolerance
	size := rect.Size()
	center := rect.Center()
	rect = s2.RectFromCenterSize(center, s2.LatLng{Lat: size.Lat * 1.1, Lng: size.Lng * 1.1})

	emit = func(e Element) error {
		switch e := e.(type) {
		case *Node:
			if !rect.ContainsLatLng(s2.LatLngFromDegrees(e.Location.Lat, e.Location.Lng)) {
				t.Errorf("Unexpected point: %v", e.Location)
			}
			nodes++
		case *Way:
			ways++
		case *Relation:
			relations++
		}
		return nil
	}
	if err := ReadPBF(&buffer, emit); err != nil {
		t.Errorf("Failed to read back PBF: %s", err)
		return
	}

	if nodes != expectedNodes {
		t.Errorf("Expected %d nodes, found %d", expectedNodes, nodes)
	}
	if ways != expectedWays {
		t.Errorf("Expected %d ways, found %d", expectedWays, ways)
	}
	if relations != expectedRelations {
		t.Errorf("Expected %d relations, found %d", expectedRelations, relations)
	}
}

func TestWritePBFWithManyBlocks(t *testing.T) {
	const elementsPerRun = 9000 // Just more than a 8000 element block
	const runs = 3
	const expectedElements = elementsPerRun * runs

	var buffer bytes.Buffer
	writer, err := NewWriter(&buffer)
	if err != nil {
		t.Errorf("NewWriter(): %s", err)
		return
	}

	node := &Node{
		ID: NodeID(42),
		Location: LatLng{
			Lat: 51.5354932,
			Lng: -0.1258180,
		},
		Tags: []Tag{
			Tag{
				Key:   "natural",
				Value: "tree",
			},
		},
	}
	way := &Way{
		ID:    WayID(42),
		Nodes: []NodeID{1},
		Tags: []Tag{
			Tag{
				Key:   "highway",
				Value: "service",
			},
		},
	}
	relation := &Relation{
		ID: RelationID(42),
		Members: []Member{
			Member{
				ID:   1,
				Type: ElementTypeNode,
				Role: "forward",
			},
		},
		Tags: []Tag{
			Tag{
				Key:   "type",
				Value: "route",
			},
		},
	}

	for run := 0; run < runs; run++ {
		for i := 0; i < elementsPerRun; i++ {
			if err := writer.WriteNode(node); err != nil {
				t.Errorf("WriteNode(): %s", err)
				return
			}
		}
		for i := 0; i < elementsPerRun; i++ {
			if err := writer.WriteWay(way); err != nil {
				t.Errorf("WriteWay(): %s", err)
				return
			}
		}
		for i := 0; i < elementsPerRun; i++ {
			if err := writer.WriteRelation(relation); err != nil {
				t.Errorf("WriteRelation(): %s", err)
				return
			}
		}
	}
	writer.Flush()

	nodes := 0
	ways := 0
	relations := 0
	emit := func(e Element) error {
		switch e := e.(type) {
		case *Node:
			if len(e.Tags) != len(node.Tags) {
				t.Errorf("Expected %d tags, found %d", len(node.Tags), len(e.Tags))
			}
			nodes++
		case *Way:
			if len(e.Tags) != len(way.Tags) {
				t.Errorf("Expected %d tags, found %d", len(way.Tags), len(e.Tags))
			}
			if len(e.Nodes) != len(way.Nodes) {
				t.Errorf("Expected %d nodes, found %d", len(way.Nodes), len(e.Nodes))
			}
			ways++
		case *Relation:
			if len(e.Tags) != len(relation.Tags) {
				t.Errorf("Expected %d tags, found %d", len(relation.Tags), len(e.Tags))
			}
			if len(e.Members) != len(relation.Members) {
				t.Errorf("Expected %d members, found %d", len(relation.Members), len(e.Members))
			}
			relations++
		}
		return nil
	}
	if err := ReadPBF(&buffer, emit); err != nil {
		t.Errorf("ReadPBF(): %s", err)
		return
	}

	if nodes != expectedElements {
		t.Errorf("Expected %d nodes, found %d", expectedElements, nodes)
	}
	if ways != expectedElements {
		t.Errorf("Expected %d ways, found %d", expectedElements, ways)
	}
	if relations != expectedElements {
		t.Errorf("Expected %d relations, found %d", expectedElements, relations)
	}
}

func TestUpdateBoundingBox(t *testing.T) {
	box := EmptyBoundingBox
	box.Include(LatLng{Lat: 51.5357019, Lng: -0.1260475})
	box.Include(LatLng{Lat: 51.5350667, Lng: -0.1255202})

	if box.TopLeft.Lat < box.BottomRight.Lat {
		t.Errorf("Expected Top > Bottom: %v", box)
	}
	if box.TopLeft.Lng > box.BottomRight.Lng {
		t.Errorf("Expected Left < Right: %v", box)
	}

	size := 0.1
	if box.TopLeft.Lat-box.BottomRight.Lat > size {
		t.Errorf("Expected lat size < %.2f: %v", size, box)
	}
	if box.BottomRight.Lng-box.TopLeft.Lng > size {
		t.Errorf("Expected lng size < %.2f: %v", size, box)
	}
}
