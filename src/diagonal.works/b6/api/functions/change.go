package functions

import (
	"fmt"
	"os"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/encoding"
	"diagonal.works/b6/ingest"
	"github.com/golang/geo/s2"
)

// Deprecated.
func idToRelationID(c *api.Context, namespace string, id b6.Identifiable) b6.FeatureID {
	return b6.MakeRelationID(b6.Namespace(namespace), encoding.HashString(id.FeatureID().String())).FeatureID()
}

// Add the given tag to the given feature.
func addTag(c *api.Context, id b6.Identifiable, tag b6.Tag) (ingest.Change, error) {
	tags := make(ingest.AddTags, 1)
	tags[0] = ingest.AddTag{ID: id.FeatureID(), Tag: tag}
	return tags, nil
}

// Add the given tags to the given features.
// The keys of the given collection specify the features to change, the
// values provide the tag to be added.
func addTags(c *api.Context, collection b6.Collection[b6.FeatureID, b6.Tag]) (ingest.Change, error) {
	i := collection.Begin()
	tags := make(ingest.AddTags, 0)
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		tags = append(tags, ingest.AddTag{ID: i.Key(), Tag: i.Value()})
	}
	return tags, nil
}

// Remove the tag with the given key from the given feature.
func removeTag(c *api.Context, id b6.Identifiable, key string) (ingest.Change, error) {
	tags := make(ingest.RemoveTags, 1)
	tags[0] = ingest.RemoveTag{ID: id.FeatureID(), Key: key}
	return tags, nil
}

// Remove the given tags from the given features.
// The keys of the given collection specify the features to change, the
// values provide the key of the tag to be removed.
func removeTags(c *api.Context, collection b6.Collection[b6.FeatureID, string]) (ingest.Change, error) {
	i := collection.Begin()
	tags := make(ingest.RemoveTags, 0)
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		tags = append(tags, ingest.RemoveTag{ID: i.Key(), Key: i.Value()})
	}
	return tags, nil
}

// Adds a point feature with the given id, tags and members.
func addPoint(_ *api.Context, point b6.Geometry, id b6.FeatureID, tags b6.Collection[interface{}, b6.Tag]) (ingest.Change, error) {
	p := &ingest.GenericFeature{
		ID: id,
		Tags: []b6.Tag{
			{Key: b6.PointTag, Value: b6.LatLng(s2.LatLngFromPoint(point.Point()))},
		}}

	t := tags.Begin()
	for {
		ok, err := t.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		p.Tags = append(p.Tags, t.Value())
	}

	add := ingest.AddFeatures([]ingest.Feature{p})
	return &add, nil
}

// Add a relation feature with the given id, tags and members.
func addRelation(c *api.Context, id b6.RelationID, tags b6.Collection[interface{}, b6.Tag], members b6.Collection[b6.Identifiable, string]) (ingest.Change, error) {
	r := &ingest.RelationFeature{
		RelationID: id,
	}

	t := tags.Begin()
	for {
		ok, err := t.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		r.Tags = append(r.Tags, t.Value())
	}

	m := members.Begin()
	for {
		ok, err := m.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		r.Members = append(r.Members, b6.RelationMember{ID: m.Key().FeatureID(), Role: m.Value()})
	}

	add := ingest.AddFeatures([]ingest.Feature{r})
	return &add, nil
}

// Add a collection feature with the given id, tags and items.
func addCollection(c *api.Context, id b6.CollectionID, tags b6.Collection[any, b6.Tag], collection b6.UntypedCollection) (ingest.Change, error) {
	feature := &ingest.CollectionFeature{
		CollectionID: id,
	}

	t := tags.Begin()
	for {
		ok, err := t.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		feature.Tags = append(feature.Tags, t.Value())
	}

	i := collection.BeginUntyped()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}

		feature.Keys = append(feature.Keys, i.Key())
		feature.Values = append(feature.Values, i.Value())
	}

	add := ingest.AddFeatures([]ingest.Feature{feature})
	return &add, nil
}

// Return a change that will apply all the changes in the given collection.
// Changes are applied transactionally. If the application of one change
// fails (for example, because it includes a path that references a missing
// point), then no changes will be applied.
func mergeChanges(c *api.Context, collection b6.Collection[any, ingest.Change]) (ingest.Change, error) {
	i := collection.Begin()
	merged := make(ingest.MergedChange, 0)
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		merged = append(merged, i.Value())
	}
	return merged, nil
}

// Return the result of calling the given function in a world in which the given change has been applied.
// The underlying world used by the server is not modified.
func withChange(c *api.Context, change ingest.Change, function func(c *api.Context) (interface{}, error)) (interface{}, error) {
	modified := *c
	m := ingest.NewMutableOverlayWorld(c.World)
	modified.World = m
	if _, err := change.Apply(m); err != nil {
		return nil, err
	}
	return function(&modified)
}

// Export the changes that have been applied to the world to the given filename as yaml.
// As the file is written by the b6 server process, the filename it relative
// to the filesystems it sees. Writing files to cloud storage is
// supported.
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

// Return the changes contained in the given file.
// As the file is read by the b6 server process, the filename it relative
// to the filesystems it sees. Reading from files on cloud storage is
// supported.
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
