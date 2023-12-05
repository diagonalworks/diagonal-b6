package functions

import (
	"sort"

	"diagonal.works/b6"
	"diagonal.works/b6/api"

	"github.com/golang/geo/s2"
)

// Return a collection of points representing the centroids of s2 cells that cover the given area between the given levels.
func s2Points(context *api.Context, area b6.Area, minLevel int, maxLevel int) (b6.Collection[string, b6.Point], error) {
	coverer := s2.RegionCoverer{MinLevel: minLevel, MaxLevel: maxLevel}
	cells := make(map[s2.CellID]struct{})
	for i := 0; i < area.Len(); i++ {
		for _, id := range coverer.Covering(area.Polygon(i)) {
			cells[id] = struct{}{}
		}
	}
	keys := make([]string, 0, len(cells))
	values := make([]b6.Point, 0, len(cells))
	for cell := range cells {
		keys = append(keys, cell.ToToken())
		values = append(values, b6.PointFromS2Point(cell.Point()))
	}
	return b6.ArrayCollection[string, b6.Point]{Keys: keys, Values: values}.Collection(), nil
}

// Return a collection of points representing the centroids of s2 cells that cover the given area at the given level.
func s2Grid(context *api.Context, area b6.Area, level int) (b6.Collection[int, string], error) {
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
	return b6.ArrayValuesCollection[string](tokens).Collection(), nil
}

// Return a collection of of s2 cells tokens that cover the given area at the given level.
func s2Covering(context *api.Context, area b6.Area, minLevel int, maxLevel int) (b6.Collection[int, string], error) {
	coverer := s2.RegionCoverer{MinLevel: minLevel, MaxLevel: maxLevel}
	cells := make(s2.CellUnion, 0, 4)
	for i := 0; i < area.Len(); i++ {
		cells = s2.CellUnionFromUnion(cells, coverer.CellUnion(area.Polygon(i)))
	}
	tokens := make([]string, 0, len(cells))
	for _, cell := range cells {
		tokens = append(tokens, cell.ToToken())
	}
	return b6.ArrayValuesCollection[string](tokens).Collection(), nil
}

// Return a collection the center of the s2 cell with the given token.
func s2Center(context *api.Context, token string) (b6.Point, error) {
	return b6.PointFromS2Point(s2.CellIDFromToken(token).Point()), nil
}

// Return the bounding area of the s2 cell with the given token.
func s2Polygon(context *api.Context, token string) (b6.Area, error) {
	cell := s2.CellFromCellID(s2.CellIDFromToken(token))
	points := make([]s2.Point, 4)
	for i := range points {
		points[i] = cell.Vertex(i)
	}
	return b6.AreaFromS2Loop(s2.LoopFromPoints(points)), nil
}
