package api

import (
	"fmt"
	"strconv"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
)

func Materialise(id b6.RelationID, value interface{}) (*ingest.RelationFeature, error) {
	r := &ingest.RelationFeature{
		RelationID: id,
	}
	switch v := value.(type) {
	case Collection:
		return r, materialiseCollection(v, r)
	}
	return nil, fmt.Errorf("can't materialise values of type %T", value)
}

func materialiseCollection(c Collection, r *ingest.RelationFeature) error {
	r.AddTag(b6.Tag{Key: "b6", Value: "collection"})
	i := c.Begin()

	ok, err := i.Next()
	if !ok || err != nil {
		return err
	}

	if _, ok := i.Key().(b6.Identifiable); ok {
		if _, ok := i.Value().(b6.Identifiable); ok {
			return materialiseFeatureFeatureCollection(r, i)
		}
		r.Tags.AddTag(b6.Tag{Key: "keys", Value: "features"})
		if t, err := typeForMaterialisedRole(i.Value()); err == nil {
			r.Tags.AddTag(b6.Tag{Key: "role", Value: t})
		} else {
			return err
		}
		for {
			if key, ok := i.Key().(b6.Identifiable); ok {
				r.Members = append(r.Members, b6.RelationMember{
					ID:   key.FeatureID(),
					Role: materialiseToRole(i.Value()),
				})
			} else {
				return fmt.Errorf("expected a FeatureID, found %T", i.Key())
			}
			ok, err := i.Next()
			if !ok || err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("can't materialise collections with keys of type %T", i.Key())
	}
}

func materialiseFeatureFeatureCollection(r *ingest.RelationFeature, i CollectionIterator) error {
	r.Tags.AddTag(b6.Tag{Key: "keys", Value: "features"})
	r.Tags.AddTag(b6.Tag{Key: "values", Value: "features"})
	for {
		if key, ok := i.Key().(b6.Identifiable); ok {
			if value, ok := i.Value().(b6.Identifiable); ok {
				r.Members = append(r.Members,
					b6.RelationMember{
						ID: key.FeatureID(),
					},
					b6.RelationMember{
						ID: value.FeatureID(),
					},
				)
			}
		} else {
			return fmt.Errorf("expected a FeatureID key, found %T", i.Key())
		}
		ok, err := i.Next()
		if !ok || err != nil {
			return err
		}
	}
}

func typeForMaterialisedRole(value interface{}) (string, error) {
	switch value.(type) {
	case int:
		return "int", nil
	case float64:
		return "float", nil
	case string:
		return "string", nil
	case b6.Identifiable:
		return "id", nil
	}
	return "", fmt.Errorf("can't materialise %T as role", value)
}

func materialiseToRole(value interface{}) string {
	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'e', -1, 64)
	case string:
		return v
	case b6.Identifiable:
		return "/" + v.FeatureID().String()
	}
	panic(fmt.Sprintf("can't materialise %T to role", value)) // Checked in materialiseAsRoleType
}

func dematerialiseFromRole(role string, t b6.Tag) (interface{}, error) {
	if !t.IsValid() {
		return nil, fmt.Errorf("No role tag")
	}
	switch t.Value {
	case "int":
		return strconv.Atoi(role)
	case "float":
		return strconv.ParseFloat(role, 64)
	case "string":
		return role, nil
	case "id":
		if len(role) > 0 {
			return b6.FeatureIDFromString(role[1:]), nil
		}
		return b6.FeatureIDInvalid, nil
	}
	return nil, fmt.Errorf("can't dematerialise %s from role", t.Value)
}

func Dematerialise(f b6.RelationFeature) (interface{}, error) {
	t := f.Get("b6")
	if !t.IsValid() {
		return nil, fmt.Errorf("%s isn't a materialised value", f.FeatureID())
	}
	switch t.Value {
	case "collection":
		return dematerialiseCollection(f)
	}
	return nil, fmt.Errorf("can't dematerialise features with tag %s", t)
}

func dematerialiseCollection(f b6.RelationFeature) (Collection, error) {
	// TODO: To support large collections, we could avoid the copy and
	// use a collection that dematerialised from the feature on the fly.
	c := &ArrayAnyCollection{
		Keys:   make([]interface{}, 0, f.Len()),
		Values: make([]interface{}, 0, f.Len()),
	}
	if keys := f.Get("keys"); keys.Value == "features" {
		if values := f.Get("values"); values.Value == "features" {
			if err := dematerialiseFeatureFeatureCollection(c, f); err == nil {
				return c, nil
			} else {
				return nil, err
			}
		}
	}
	role := f.Get("role")
	if !role.IsValid() {
		return nil, fmt.Errorf("no role tag")
	}
	for i := 0; i < f.Len(); i++ {
		member := f.Member(i)
		c.Keys = append(c.Keys, member.ID)
		if v, err := dematerialiseFromRole(member.Role, role); err == nil {
			c.Values = append(c.Values, v)
		} else {
			return nil, err
		}
	}
	return c, nil
}

func dematerialiseFeatureFeatureCollection(c *ArrayAnyCollection, f b6.RelationFeature) error {
	if f.Len()%2 != 0 {
		return fmt.Errorf("incorrectly materialised collection")
	}
	for i := 0; i < f.Len(); i += 2 {
		c.Keys = append(c.Keys, f.Member(i).ID)
		c.Values = append(c.Values, f.Member(i+1).ID)
	}
	return nil
}
