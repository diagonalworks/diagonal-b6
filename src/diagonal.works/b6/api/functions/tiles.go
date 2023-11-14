package functions

import (
	"strconv"

	"diagonal.works/b6"
	"diagonal.works/b6/api"

	"github.com/golang/geo/s2"
)

func tileIDs(c *api.Context, feature b6.Feature) (b6.Collection[b6.FeatureID, int], error) {
	ids := b6.ArrayCollection[b6.FeatureID, int]{}
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
	return ids.Collection(), nil
}

func tileIDsHex(c *api.Context, feature b6.Feature) (b6.Collection[b6.FeatureID, string], error) {
	ids := b6.ArrayCollection[b6.FeatureID, string]{}
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
	return ids.Collection(), nil
}

func tilePaths(c *api.Context, g b6.Geometry, zoom int) (b6.Collection[int, string], error) {
	coverer := s2.RegionCoverer{MaxLevel: 20, MinLevel: 0}
	paths := make([]string, 0)
	for _, t := range b6.CoverCellUnionWithTiles(g.Covering(coverer), uint(zoom)) {
		paths = append(paths, t.String())
	}
	return b6.ArrayValuesCollection[string](paths).Collection(), nil
}
