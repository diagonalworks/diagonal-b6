package functions

import (
	"fmt"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
)

func addTag(id b6.Identifiable, tag b6.Tag, c *api.Context) (ingest.Change, error) {
	tags := make(ingest.AddTags, 1)
	tags[0] = ingest.AddTag{ID: id.FeatureID(), Tag: tag}
	return tags, nil
}

func addTags(collection api.FeatureIDTagCollection, c *api.Context) (ingest.Change, error) {
	i := collection.Begin()
	tags := make(ingest.AddTags, 0)
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		if tag, ok := i.Value().(b6.Tag); ok {
			tags = append(tags, ingest.AddTag{ID: i.Key().(b6.FeatureID), Tag: tag})
		} else {
			return nil, fmt.Errorf("Expected %T, found %T", b6.Tag{}, i.Value())
		}
	}
	return tags, nil
}
