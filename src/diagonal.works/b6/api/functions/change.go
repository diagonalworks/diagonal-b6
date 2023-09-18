package functions

import (
	"fmt"
	"os"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/encoding"
	"diagonal.works/b6/ingest"
)

func idToRelationID(c *api.Context, namespace string, id b6.Identifiable) b6.FeatureID {
	return b6.MakeRelationID(b6.Namespace(namespace), encoding.HashString(id.FeatureID().String())).FeatureID()
}

func addTag(c *api.Context, id b6.Identifiable, tag b6.Tag) (ingest.Change, error) {
	tags := make(ingest.AddTags, 1)
	tags[0] = ingest.AddTag{ID: id.FeatureID(), Tag: tag}
	return tags, nil
}

func addTags(c *api.Context, collection api.FeatureIDTagCollection) (ingest.Change, error) {
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

func removeTag(c *api.Context, id b6.Identifiable, key string) (ingest.Change, error) {
	tags := make(ingest.RemoveTags, 1)
	tags[0] = ingest.RemoveTag{ID: id.FeatureID(), Key: key}
	return tags, nil
}

func removeTags(c *api.Context, collection api.FeatureIDStringCollection) (ingest.Change, error) {
	i := collection.Begin()
	tags := make(ingest.RemoveTags, 0)
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		if key, ok := i.Value().(string); ok {
			tags = append(tags, ingest.RemoveTag{ID: i.Key().(b6.FeatureID), Key: key})
		} else {
			return nil, fmt.Errorf("Expected string, found %T", i.Value())
		}
	}
	return tags, nil
}

func addRelation(c *api.Context, id b6.RelationID, tags api.TagCollection, members api.FeatureIDStringCollection) (ingest.Change, error) {
	r := &ingest.RelationFeature{
		RelationID: id,
	}

	i := tags.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		tag, ok := i.Value().(b6.Tag)
		if !ok {
			return nil, fmt.Errorf("Expected b6.Tag, found %T", i.Value())
		}
		r.Tags = append(r.Tags, tag)
	}

	i = members.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		member, ok := i.Key().(b6.Identifiable)
		if !ok {
			return nil, fmt.Errorf("Expected b6.FeatureID for member, found %T", i.Key())
		}
		role, ok := i.Value().(string)
		if !ok {
			return nil, fmt.Errorf("Expected string for role, found %T", i.Value())
		}
		r.Members = append(r.Members, b6.RelationMember{ID: member.FeatureID(), Role: role})
	}
	return &ingest.AddFeatures{
		Relations: []*ingest.RelationFeature{r},
	}, nil
}

func mergeChanges(c *api.Context, collection api.AnyChangeCollection) (ingest.Change, error) {
	i := collection.Begin()
	merged := make(ingest.MergedChange, 0)
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		if c, ok := i.Value().(ingest.Change); ok {
			merged = append(merged, c)
		} else {
			return nil, fmt.Errorf("Expected Change, found %T", i.Value())
		}
	}
	return merged, nil
}

func withChange(c *api.Context, change ingest.Change, f func(c *api.Context) (interface{}, error)) (interface{}, error) {
	modified := *c
	m := ingest.NewMutableOverlayWorld(c.World)
	modified.World = m
	if _, err := change.Apply(m); err != nil {
		return nil, err
	}
	return f(&modified)
}

func changesToFile(c *api.Context, filename string) (string, error) {
	if !c.FileIOAllowed {
		return "", fmt.Errorf("File IO is not allowed")
	}

	w, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open %s for write: %w", filename, err)
	}
	if m, ok := c.World.(ingest.MutableWorld); ok {
		err = ingest.ExportChangesAsYAML(m, w)
	}
	return filename, w.Close()
}

func changesFromFile(c *api.Context, filename string) (ingest.Change, error) {
	if !c.FileIOAllowed {
		return nil, fmt.Errorf("File IO is not allowed")
	}

	r, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s for read: %w", filename, err)
	}
	return ingest.IngestChangesFromYAML(r), nil
}
