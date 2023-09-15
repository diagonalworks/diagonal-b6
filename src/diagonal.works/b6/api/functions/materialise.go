package functions

import (
	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
)

func materialise(context *api.Context, id b6.RelationID, value interface{}) (ingest.Change, error) {
	if r, err := api.Materialise(id, value); err == nil {
		return &ingest.AddFeatures{
			Relations: []*ingest.RelationFeature{r},
		}, nil
	} else {
		return nil, err
	}
}

func dematerialise(context *api.Context, r b6.RelationFeature) (interface{}, error) {
	return api.Dematerialise(r)
}
