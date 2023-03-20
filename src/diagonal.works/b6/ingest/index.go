package ingest

import (
	"sort"

	"github.com/golang/geo/s2"
)

type S2Range struct {
	Min s2.CellID
	Max s2.CellID
}

type S2Ranges []S2Range

func (r S2Ranges) Len() int           { return len(r) }
func (r S2Ranges) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r S2Ranges) Less(i, j int) bool { return r[i].Min < r[j].Min }

func (r *S2Ranges) AddRange(min s2.CellID, max s2.CellID) {
	if min <= max {
		*r = append(*r, S2Range{Min: min, Max: max})
	} else {
		*r = append(*r, S2Range{Min: max, Max: min})
	}
}

func (r *S2Ranges) AddCell(cell s2.CellID) {
	*r = append(*r, S2Range{Min: cell, Max: cell})
}

func (r *S2Ranges) Normalise() {
	sort.Sort(*r)
	for i := 0; i < len(*r)-1; i++ {
		if (*r)[i+1].Min <= (*r)[i].Max {
			if (*r)[i+1].Max > (*r)[i].Max {
				(*r)[i].Max = (*r)[i+1].Max
			}
			copy((*r)[i+1:], (*r)[i+2:])
			*r = (*r)[0 : len(*r)-1]
			i--
		}
	}
}

// TODO: Add an explit test now we've factored this out?
func S2RangesForSearch(union s2.CellUnion) S2Ranges {
	ranges := make(S2Ranges, 0)
	for _, cell := range union {
		ranges.AddRange(cell.RangeMin(), cell.RangeMax())
		for {
			ranges.AddCell(cell)
			if cell.Level() == 0 {
				break
			}
			cell = cell.Parent(cell.Level() - 1)
		}
	}
	ranges.Normalise()
	return ranges
}

// TODO: Move somewhere else - like search/, which can move out of ingest?
type cellIndex interface {
	Len() int
	Cell(i int) s2.CellID
}

type searcher struct {
	coverer s2.RegionCoverer
}

func (s *searcher) Search(region s2.Region, cells cellIndex, f func(i int)) {
	covering := s.coverer.Covering(region)
	s.searchRanges(S2RangesForSearch(covering), cells, f)
}

func (s *searcher) searchRanges(ranges S2Ranges, cells cellIndex, f func(i int)) {
	offset := 0
	for i := 0; i < len(ranges); i++ {
		if offset >= cells.Len() {
			return
		}
		j := sort.Search(cells.Len()-offset, func(k int) bool {
			return cells.Cell(offset+k) >= ranges[i].Min
		})
		for ; j+offset < cells.Len() && cells.Cell(j+offset) <= ranges[i].Max; j++ {
			f(j + offset)
		}
		offset = j
	}
}

func (s *searcher) searchRangesBackwards(ranges S2Ranges, cells cellIndex, f func(i int)) {
	n := cells.Len()
	for i := len(ranges) - 1; i > 0; i-- {
		if n == 0 {
			return
		}
		j := sort.Search(n, func(k int) bool {
			return cells.Cell(k) > ranges[i].Max
		})
		j--
		for ; j >= 0 && cells.Cell(j) >= ranges[i].Min; j-- {
			f(j)
		}
		n = j + 1
	}
}

func (s *searcher) SearchCellUnion(union s2.CellUnion, cells cellIndex, f func(i int)) {
	s.searchRanges(S2RangesForSearch(union), cells, f)
}

func (s *searcher) SearchCellUnionBackwards(union s2.CellUnion, cells cellIndex, f func(i int)) {
	s.searchRangesBackwards(S2RangesForSearch(union), cells, f)
}

func (s *searcher) SearchLeaves(region s2.Region, cells cellIndex, f func(i int)) {
	covering := s.coverer.Covering(region)
	covering.Normalize()
	s.SearchLeavesCellUnion(covering, cells, f)
}

func (s *searcher) SearchLeavesCellUnion(union s2.CellUnion, cells cellIndex, f func(i int)) {
	ranges := make(S2Ranges, len(union))
	for i, cell := range union {
		ranges[i] = S2Range{Min: cell.RangeMin(), Max: cell.RangeMax()}
	}
	s.searchRanges(ranges, cells, f)
}

func (s *searcher) SearchLeavesCellUnionBackwards(union s2.CellUnion, cells cellIndex, f func(i int)) {
	ranges := make(S2Ranges, len(union))
	for i, cell := range union {
		ranges[i] = S2Range{Min: cell.RangeMin(), Max: cell.RangeMax()}
	}
	s.searchRangesBackwards(ranges, cells, f)
}

// TODO: duplicated with geo, factor out
type CellIDs []s2.CellID

func (c CellIDs) Len() int             { return len(c) }
func (c CellIDs) Cell(i int) s2.CellID { return c[i] }
func (c CellIDs) Swap(i, j int)        { c[i], c[j] = c[j], c[i] }
func (c CellIDs) Less(i, j int) bool   { return c[i] < c[j] }
