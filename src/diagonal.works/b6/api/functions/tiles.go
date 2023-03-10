package functions

import (
	"strconv"

	"diagonal.works/b6"
	"diagonal.works/b6/api"

	"github.com/golang/geo/s2"
)

func tileIDs(feature b6.Feature, c *api.Context) (api.FeatureIDIntCollection, error) {
	ids := &api.ArrayFeatureIDIntCollection{}
	if a, ok := feature.(b6.AreaFeature); ok {
		ids.Keys = make([]b6.FeatureID, a.Len())
		ids.Values = make([]int, a.Len())
		for i := range ids.Values {
			ids.Values[i] = int(api.TileFeatureIDForPolygon(feature.FeatureID(), i))
		}
	} else {
		ids.Keys = make([]b6.FeatureID, 1)
		ids.Values = []int{int(api.TileFeatureID(feature.FeatureID()))}
	}
	for i := range ids.Keys {
		ids.Keys[i] = feature.FeatureID()
	}
	return ids, nil
}

func tileIDsHex(feature b6.Feature, c *api.Context) (api.FeatureIDStringCollection, error) {
	ids := &api.ArrayFeatureIDStringCollection{}
	if a, ok := feature.(b6.AreaFeature); ok {
		ids.Keys = make([]b6.FeatureID, a.Len())
		ids.Values = make([]string, a.Len())
		for i := range ids.Values {
			ids.Values[i] = strconv.FormatUint(api.TileFeatureIDForPolygon(feature.FeatureID(), i), 16)
		}
	} else {
		ids.Keys = make([]b6.FeatureID, 1)
		ids.Values = []string{strconv.FormatUint(api.TileFeatureID(feature.FeatureID()), 16)}
	}
	for i := range ids.Keys {
		ids.Keys[i] = feature.FeatureID()
	}
	return ids, nil
}

func tilePaths(g b6.Geometry, zoom int, c *api.Context) (api.IntStringCollection, error) {
	coverer := s2.RegionCoverer{MaxLevel: 20, MinLevel: 0}
	paths := make([]string, 0)
	for _, t := range b6.CoverCellUnionWithTiles(g.Covering(coverer), uint(zoom)) {
		paths = append(paths, t.String())
	}
	return &api.ArrayIntStringCollection{Values: paths}, nil
}
