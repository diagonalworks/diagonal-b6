package search

import (
	"fmt"
	"strings"

	"github.com/golang/geo/s2"
)

const MaxIndexedCellLevel = 16

const s2CellIDTokenPrefix = "s2:"
const s2AncestorCellIDTokenPrefix = "a2:"

func cellIDToToken(cell s2.CellID) string {
	return fmt.Sprintf("%s%s", s2CellIDTokenPrefix, cell.ToToken())
}

func tokenToCellID(token string) (s2.CellID, bool) {
	if strings.HasPrefix(token, s2CellIDTokenPrefix) {
		return s2.CellIDFromToken(token[len(s2CellIDTokenPrefix):]), true
	}
	return s2.CellID(0), false
}

func ancestorCellIDToToken(cell s2.CellID) string {
	return fmt.Sprintf("%s%s", s2AncestorCellIDTokenPrefix, cell.ToToken())
}

func MakeCoverer() s2.RegionCoverer {
	return s2.RegionCoverer{MaxLevel: 16, MaxCells: 5}
}

type Spatial s2.CellUnion

type SpatialIndex interface {
	RewriteSpatialQuery(query Spatial) Query
}

func NewSpatialFromRegion(region s2.Region) Spatial {
	coverer := MakeCoverer()
	return Spatial(coverer.Covering(region))
}

func NewSpatialFromCellUnion(union s2.CellUnion) Spatial {
	return Spatial(union)
}

func (s Spatial) String() string {
	joined := "(spatial"
	for _, cellID := range s2.CellUnion(s) {
		joined = joined + " " + cellID.ToToken()
	}
	return joined + ")"
}

func (s Spatial) Covering() s2.CellUnion {
	return s2.CellUnion(s)
}

func (s Spatial) Compile(index Index) Iterator {
	return RewriteSpatialQuery(s).Compile(index)
}

func RewriteSpatialQuery(q Spatial) Query {
	covering := q.Covering()
	rewritten := make(Union, 0, len(covering)*3)
	ids := make(map[s2.CellID]struct{})
	for _, id := range covering {
		rewritten = append(rewritten, All{Token: ancestorCellIDToToken(id)})
		for {
			ids[id] = struct{}{}
			if id.Level() == 0 {
				break
			}
			id = id.Parent(id.Level() - 1)
		}
	}
	for id := range ids {
		rewritten = append(rewritten, All{Token: cellIDToToken(id)})
	}
	return rewritten
}

func TokensForCovering(covering s2.CellUnion, tokens []string) []string {
	for _, cell := range covering {
		if cell.Level() == 0 {
			continue
		}
		tokens = append(tokens, cellIDToToken(cell))
	}
	return cellIDAncestorTokens(covering, tokens)
}

func cellIDAncestorTokens(covering s2.CellUnion, tokens []string) []string {
	cells := make(map[s2.CellID]struct{})
	for _, cell := range covering {
		cells[cell] = struct{}{}
	}
	for len(cells) > 0 {
		parents := make(map[s2.CellID]struct{})
		for id := range cells {
			if id.Level() != 0 {
				parents[id.Parent(id.Level()-1)] = struct{}{}
			}
		}
		for id := range parents {
			tokens = append(tokens, ancestorCellIDToToken(id))
		}
		cells = parents
	}
	return tokens
}
