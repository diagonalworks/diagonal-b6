package functions

import (
	"fmt"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/search"
)

func debugTokens(id b6.Identifiable, c *api.Context) (api.IntStringCollection, error) {
	if f := api.Resolve(id, c.World); f != nil {
		if p, ok := f.(b6.PhysicalFeature); ok {
			return &api.ArrayIntStringCollection{
				Values: ingest.TokensForFeature(p),
			}, nil
		} else {
			return &api.ArrayIntStringCollection{
				Values: []string{},
			}, nil
		}
	}
	return nil, fmt.Errorf("No feature with id %s", id.FeatureID())
}

func debugAllQuery(token string, c *api.Context) (search.Query, error) {
	return search.All{Token: token}, nil
}
