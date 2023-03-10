package search

import (
	"log"

	"github.com/golang/geo/s2"
)

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
	if spatialIndex, ok := index.(SpatialIndex); ok {
		return spatialIndex.RewriteSpatialQuery(s).Compile(index)
	}
	log.Printf("Using expensive All query for %s", s.String())
	return All{Token: AllToken}.Compile(index)
}
