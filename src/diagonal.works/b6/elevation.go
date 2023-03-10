package b6 

import (
	"sort"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

type Elevations interface {
	Elevation(p s2.Point) (float64, bool)
}

// ElevationField stores a set of spot elevations, and returns the
// elevation of a point based on the weighted sum of spot heights
// within Radius. Finish() must be called after all Add() calls,
// and before Elevation()
type ElevationField struct {
	Radius     s1.Angle
	cells      []s2.CellID
	elevations []float32
}

type elevationField ElevationField // Hide internal Sort methods

func (e *elevationField) Len() int { return len(e.cells) }
func (e *elevationField) Swap(i, j int) {
	e.cells[i], e.cells[j] = e.cells[j], e.cells[i]
	e.elevations[i], e.elevations[j] = e.elevations[j], e.elevations[i]
}
func (e *elevationField) Less(i, j int) bool { return e.cells[i] < e.cells[j] }

func (e *ElevationField) Add(ll s2.LatLng, height float64) {
	e.cells = append(e.cells, s2.CellIDFromLatLng(ll))
	e.elevations = append(e.elevations, float32(height))
}

func (e *ElevationField) Finish() {
	sort.Sort((*elevationField)(e))
}

func (e *ElevationField) Elevation(p s2.Point) (float64, bool) {
	coverer := s2.RegionCoverer{MaxCells: 5, MaxLevel: 20}
	cap := s2.CapFromCenterAngle(p, e.Radius)
	et := 0.0
	wt := 0.0
	for _, cell := range coverer.Covering(cap) {
		j := sort.Search(len(e.cells), func(i int) bool {
			return e.cells[i] >= cell.RangeMin()
		})
		for j < len(e.cells) && e.cells[j] <= cell.RangeMax() {
			if d := p.Distance(e.cells[j].Point()); d < e.Radius {
				w := float64(1.0 / d)
				wt += w
				et += w * float64(e.elevations[j])
			}
			j++
		}
	}
	if wt > 0.0 {
		return et / wt, true
	}
	return 0.0, false
}
