package api

import (
	"fmt"
	"reflect"

	"diagonal.works/b6"
)

type HistogramCollection struct {
	UntypedCollection b6.UntypedCollection
	i                 b6.Iterator[any, any]
}

func (h *HistogramCollection) BeginUntyped() b6.Iterator[any, any] {
	return h.Begin()
}

func (h *HistogramCollection) Begin() b6.Iterator[any, any] {
	return &HistogramCollection{UntypedCollection: h.UntypedCollection, i: h.UntypedCollection.BeginUntyped()}
}

func (h *HistogramCollection) Next() (bool, error) { return h.i.Next() }
func (h *HistogramCollection) Key() interface{}    { return h.i.Key() }
func (h *HistogramCollection) Value() interface{}  { return h.i.Value() }
func (h *HistogramCollection) Count() (int, bool)  { return h.UntypedCollection.Count() }

var _ b6.UntypedCollection = &HistogramCollection{}

func FillMap(c b6.UntypedCollection, toFill interface{}) error {
	f := reflect.ValueOf(toFill)
	if f.Kind() != reflect.Map {
		return fmt.Errorf("expected a map, found %T", toFill)
	}
	kt := f.Type().Key()
	vt := f.Type().Elem()
	i := c.BeginUntyped()
	for {
		ok, err := i.Next()
		if !ok || err != nil {
			return err
		}
		if k, err := Convert(reflect.ValueOf(i.Key()), kt, nil); err == nil {
			if v, err := Convert(reflect.ValueOf(i.Value()), vt, nil); err == nil {
				f.SetMapIndex(k, v)
			} else {
				return err
			}
		} else {
			return fmt.Errorf("Can't assign %T to key %s", i.Key(), f.Type().Key())
		}
	}
}

func FillSliceFromValues(c b6.UntypedCollection, toFill interface{}) error {
	f := reflect.ValueOf(toFill)
	if f.Kind() != reflect.Ptr || f.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("expected a pointer to a slice, found %T", toFill)
	}
	f = f.Elem()
	vt := f.Type().Elem()
	i := c.BeginUntyped()
	for {
		ok, err := i.Next()
		if !ok {
			break
		} else if err != nil {
			return err
		}
		if v, err := Convert(reflect.ValueOf(i.Value()), vt, nil); err == nil {
			f = reflect.Append(f, v)
		} else {
			return fmt.Errorf("Can't append %T to %s", i.Value(), f.Type())
		}
	}
	reflect.ValueOf(toFill).Elem().Set(f)
	return nil
}

func CollectionToTags(c b6.UntypedCollection) (b6.Tags, error) {
	tags := make(b6.Tags, 0)
	i := c.BeginUntyped()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			return tags, nil
		}
		if tag, ok := i.Value().(b6.Tag); ok {
			tags = append(tags, tag)
			continue
		} else if key, ok := i.Key().(string); ok {
			if value, ok := i.Value().(string); ok {
				tags = append(tags, b6.Tag{Key: key, Value: value})
				continue
			}
		}
		return nil, fmt.Errorf("Expected tag values, or string keys and values, found %T and %T", i.Key(), i.Value())
	}
}
