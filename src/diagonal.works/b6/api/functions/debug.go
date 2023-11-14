package functions

import (
	"fmt"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/search"
)

func debugTokens(c *api.Context, id b6.Identifiable) (b6.Collection[int, string], error) {
	if f := api.Resolve(id, c.World); f != nil {
		return b6.ArrayValuesCollection[string](ingest.TokensForFeature(f)).Collection(), nil
	}
	return b6.Collection[int, string]{}, fmt.Errorf("No feature with id %s", id.FeatureID())
}

func debugAllQuery(c *api.Context, token string) (search.Query, error) {
	return search.All{Token: token}, nil
}
