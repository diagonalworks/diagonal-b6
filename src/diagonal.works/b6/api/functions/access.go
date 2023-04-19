package functions

import (
	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/graph"
)

func buildingAccess(origins api.FeatureCollection, limit float64, mode string, c *api.Context) (api.FeatureIDFeatureIDCollection, error) {
	o := make(map[b6.FeatureID]b6.Feature)
	i := origins.Begin()
	for {
		if ok, err := i.Next(); err != nil {
			return nil, err
		} else if !ok {
			break
		}
		if f, ok := i.Value().(b6.Feature); ok {
			o[f.FeatureID()] = f
		}
	}

	from := make([]b6.FeatureID, 0)
	to := make([]b6.FeatureID, 0)

	weights := graph.SimpleHighwayWeights{}
	for _, f := range o {
		s := graph.NewShortestPathSearchFromFeature(f, weights, c.World)
		s.ExpandSearch(limit, weights, graph.PointsAndAreas, c.World)
		for id := range s.AreaDistances() {
			if reached := b6.FindAreaByID(id, c.World); reached != nil {
				if reached.Get("#building").IsValid() {
					from = append(from, f.FeatureID())
					to = append(to, id.FeatureID())
				}
			}
		}
	}
	return &api.ArrayFeatureIDFeatureIDCollection{
		Keys:   from,
		Values: to,
	}, nil
}
