package functions

import (
	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/graph"
)

func buildingAccess(c *api.Context, origins b6.Collection[interface{}, b6.Feature], limit float64, mode string) (b6.Collection[b6.FeatureID, b6.FeatureID], error) {
	o := make(map[b6.FeatureID]b6.Feature)
	i := origins.Begin()
	for {
		if ok, err := i.Next(); err != nil {
			return b6.Collection[b6.FeatureID, b6.FeatureID]{}, err
		} else if !ok {
			break
		}
		o[i.Value().FeatureID()] = i.Value()
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
	return b6.ArrayCollection[b6.FeatureID, b6.FeatureID]{
		Keys:   from,
		Values: to,
	}.Collection(), nil
}
