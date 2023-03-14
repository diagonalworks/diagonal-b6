package functions

import (
	"sort"

	"diagonal.works/b6"
	"diagonal.works/b6/api"

	"github.com/golang/geo/s2"
)

// compactArrayStringCollection is a StringCollection where the value for each key
// is the same as that key
type compactArrayStringCollection struct {
	strings []string
	i       int
}

func (c *compactArrayStringCollection) Count() int { return len(c.strings) }

func (c *compactArrayStringCollection) Begin() api.CollectionIterator {
	return &compactArrayStringCollection{
		strings: c.strings,
		i:       0,
	}
}

func (c *compactArrayStringCollection) Key() interface{} {
	return c.StringKey()
}

func (c *compactArrayStringCollection) Value() interface{} {
	return c.StringValue()
}

func (c *compactArrayStringCollection) StringKey() string {
	return c.strings[c.i-1]
}

func (c *compactArrayStringCollection) StringValue() string {
	return c.strings[c.i-1]
}

func (c *compactArrayStringCollection) Next() (bool, error) {
	c.i++
	return c.i <= len(c.strings), nil
}

var _ api.Collection = &compactArrayStringCollection{}
var _ api.Countable = &compactArrayStringCollection{}

func s2Points(area b6.Area, minLevel int, maxLevel int, context *api.Context) (api.StringPointCollection, error) {
	coverer := s2.RegionCoverer{MinLevel: minLevel, MaxLevel: maxLevel}
	cells := make(map[s2.CellID]struct{})
	for i := 0; i < area.Len(); i++ {
		for _, id := range coverer.Covering(area.Polygon(i)) {
			cells[id] = struct{}{}
		}
	}
	keys := make([]string, 0, len(cells))
	values := make([]s2.Point, 0, len(cells))
	for cell := range cells {
		keys = append(keys, cell.ToToken())
		values = append(values, cell.Point())
	}
	return &api.ArrayPointCollection{Keys: keys, Values: values}, nil
}

func s2Grid(area b6.Area, level int, context *api.Context) (api.StringStringCollection, error) {
	coverer := s2.RegionCoverer{MinLevel: level, MaxLevel: level}
	cells := make(map[s2.CellID]struct{})
	for i := 0; i < area.Len(); i++ {
		for _, id := range coverer.Covering(area.Polygon(i)) {
			cells[id] = struct{}{}
		}
	}
	tokens := make([]string, 0, len(cells))
	for cell := range cells {
		tokens = append(tokens, cell.ToToken())
	}
	sort.Strings(tokens)
	return &compactArrayStringCollection{strings: tokens}, nil
}

func s2Covering(area b6.Area, minLevel int, maxLevel int, context *api.Context) (api.StringStringCollection, error) {
	coverer := s2.RegionCoverer{MinLevel: minLevel, MaxLevel: maxLevel}
	cells := make(s2.CellUnion, 0, 4)
	for i := 0; i < area.Len(); i++ {
		cells = s2.CellUnionFromUnion(cells, coverer.CellUnion(area.Polygon(i)))
	}
	tokens := make([]string, 0, len(cells))
	for _, cell := range cells {
		tokens = append(tokens, cell.ToToken())
	}
	return &compactArrayStringCollection{strings: tokens}, nil
}

func s2Center(token string, context *api.Context) (b6.Point, error) {
	return b6.PointFromS2Point(s2.CellIDFromToken(token).Point()), nil
}

func s2Polygon(token string, context *api.Context) (b6.Area, error) {
	cell := s2.CellFromCellID(s2.CellIDFromToken(token))
	points := make([]s2.Point, 4)
	for i := range points {
		points[i] = cell.Vertex(i)
	}
	return b6.AreaFromS2Loop(s2.LoopFromPoints(points)), nil
}
