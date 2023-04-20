package api

import (
	"fmt"
	"reflect"

	"diagonal.works/b6"

	"github.com/golang/geo/s2"
)

type CollectionIterator interface {
	Next() (bool, error)
	Key() interface{}
	Value() interface{}
}

type Collection interface {
	Begin() CollectionIterator
}

type Countable interface {
	Count() int
}

func Count(c Collection) int {
	if c, ok := c.(Countable); ok {
		return c.Count()
	}
	i := c.Begin()
	n := 0
	for {
		ok, err := i.Next()
		if !ok || err != nil {
			break
		}
		n++
	}
	return n
}

type EmptyCollection struct{}

func (e EmptyCollection) Begin() CollectionIterator { return e }
func (e EmptyCollection) Count() int                { return 0 }
func (e EmptyCollection) Next() (bool, error)       { return false, nil }
func (e EmptyCollection) Key() interface{}          { return nil }
func (e EmptyCollection) Value() interface{}        { return nil }

var _ Collection = EmptyCollection{}
var _ Countable = EmptyCollection{}

type PointCollection Collection
type StringPointCollection Collection
type PathCollection Collection
type AreaCollection Collection
type StringAreaCollection Collection
type FeatureCollection Collection
type PointFeatureCollection Collection
type PathFeatureCollection Collection
type AreaFeatureCollection Collection
type RelationFeatureCollection Collection
type IntStringCollection Collection
type IntTagCollection Collection
type StringStringCollection Collection
type FeatureIDAnyCollection Collection
type FeatureIDIntCollection Collection
type FeatureIDStringCollection Collection
type FeatureIDTagCollection Collection
type FeatureIDFeatureIDCollection Collection
type FeatureIDStringStringPairCollection Collection
type AnyFloatCollection Collection
type AnyRenderableCollection Collection

type ArrayAnyCollection struct {
	Keys   []interface{}
	Values []interface{}
	i      int
}

func (a *ArrayAnyCollection) Count() int { return len(a.Keys) }

func (a *ArrayAnyCollection) Begin() CollectionIterator {
	return &ArrayAnyCollection{
		Keys:   a.Keys,
		Values: a.Values,
	}
}

func (a *ArrayAnyCollection) Key() interface{} {
	return a.Keys[a.i-1]
}

func (a *ArrayAnyCollection) Value() interface{} {
	return a.Values[a.i-1]
}

func (a *ArrayAnyCollection) Next() (bool, error) {
	a.i++
	return a.i <= len(a.Keys), nil
}

var _ Collection = &ArrayAnyCollection{}
var _ Countable = &ArrayAnyCollection{}

type ArrayIntStringCollection struct {
	Values []string
	i      int
}

func (a *ArrayIntStringCollection) Count() int { return len(a.Values) }

func (a *ArrayIntStringCollection) Begin() CollectionIterator {
	return &ArrayIntStringCollection{
		Values: a.Values,
	}
}

func (a *ArrayIntStringCollection) Key() interface{} {
	return a.IntKey()
}

func (a *ArrayIntStringCollection) Value() interface{} {
	return a.StringValue()
}

func (a *ArrayIntStringCollection) IntKey() int {
	return a.i - 1
}

func (a *ArrayIntStringCollection) StringValue() string {
	return a.Values[a.i-1]
}

func (a *ArrayIntStringCollection) Next() (bool, error) {
	a.i++
	return a.i <= len(a.Values), nil
}

var _ Collection = &ArrayIntStringCollection{}
var _ Countable = &ArrayIntStringCollection{}

type ArrayStringStringCollection struct {
	Keys   []string
	Values []string
	i      int
}

func (a *ArrayStringStringCollection) Count() int { return len(a.Keys) }

func (a *ArrayStringStringCollection) Begin() CollectionIterator {
	return &ArrayStringStringCollection{
		Keys:   a.Keys,
		Values: a.Values,
	}
}

func (a *ArrayStringStringCollection) Key() interface{} {
	return a.StringKey()
}

func (a *ArrayStringStringCollection) Value() interface{} {
	return a.StringValue()
}

func (a *ArrayStringStringCollection) StringKey() string {
	return a.Keys[a.i-1]
}

func (a *ArrayStringStringCollection) StringValue() string {
	return a.Values[a.i-1]
}

func (a *ArrayStringStringCollection) Next() (bool, error) {
	a.i++
	return a.i <= len(a.Keys), nil
}

var _ Collection = &ArrayStringStringCollection{}
var _ Countable = &ArrayStringStringCollection{}

type ArrayFeatureCollection struct {
	Features []b6.Feature
	i        int
}

func (a *ArrayFeatureCollection) Count() int { return len(a.Features) }

func (a *ArrayFeatureCollection) Begin() CollectionIterator {
	return &ArrayFeatureCollection{
		Features: a.Features,
	}
}

func (a *ArrayFeatureCollection) Key() interface{} {
	return a.FeatureIDKey()
}

func (a *ArrayFeatureCollection) Value() interface{} {
	return a.FeatureValue()
}

func (a *ArrayFeatureCollection) FeatureIDKey() b6.FeatureID {
	return a.Features[a.i-1].FeatureID()
}

func (a *ArrayFeatureCollection) FeatureValue() b6.Feature {
	return a.Features[a.i-1]
}

func (a *ArrayFeatureCollection) Next() (bool, error) {
	a.i++
	return a.i <= len(a.Features), nil
}

var _ Collection = &ArrayFeatureCollection{}
var _ Countable = &ArrayFeatureCollection{}

type ArrayPointCollection struct {
	Keys   []string
	Values []s2.Point
	i      int
}

func (a *ArrayPointCollection) Count() int { return len(a.Keys) }

func (a *ArrayPointCollection) Begin() CollectionIterator {
	return &ArrayPointCollection{
		Keys:   a.Keys,
		Values: a.Values,
	}
}

func (a *ArrayPointCollection) Key() interface{} {
	return a.StringKey()
}

func (a *ArrayPointCollection) Value() interface{} {
	return a.PointValue()
}

func (a *ArrayPointCollection) StringKey() string {
	return a.Keys[a.i-1]
}

func (a *ArrayPointCollection) PointValue() b6.Point {
	return b6.PointFromS2Point(a.Values[a.i-1])
}

func (a *ArrayPointCollection) Next() (bool, error) {
	a.i++
	return a.i <= len(a.Keys), nil
}

var _ Collection = &ArrayPointCollection{}
var _ Countable = &ArrayPointCollection{}

type ArrayPointFeatureCollection struct {
	Features []b6.PointFeature
	i        int
}

func (a *ArrayPointFeatureCollection) Count() int { return len(a.Features) }

func (a *ArrayPointFeatureCollection) Begin() CollectionIterator {
	return &ArrayPointFeatureCollection{
		Features: a.Features,
	}
}

func (a *ArrayPointFeatureCollection) Key() interface{} {
	return a.FeatureIDKey()
}

func (a *ArrayPointFeatureCollection) Value() interface{} {
	return a.PointFeatureValue()
}

func (a *ArrayPointFeatureCollection) FeatureIDKey() b6.FeatureID {
	return a.Features[a.i-1].FeatureID()
}

func (a *ArrayPointFeatureCollection) PointFeatureValue() b6.PointFeature {
	return a.Features[a.i-1]
}

func (a *ArrayPointFeatureCollection) Next() (bool, error) {
	a.i++
	return a.i <= len(a.Features), nil
}

var _ Collection = &ArrayPointFeatureCollection{}
var _ Countable = &ArrayPointFeatureCollection{}

type ArrayPathFeatureCollection struct {
	Features []b6.PathFeature
	i        int
}

func (a *ArrayPathFeatureCollection) Count() int { return len(a.Features) }

func (a *ArrayPathFeatureCollection) Begin() CollectionIterator {
	return &ArrayPathFeatureCollection{
		Features: a.Features,
	}
}

func (a *ArrayPathFeatureCollection) Key() interface{} {
	return a.FeatureIDKey()
}

func (a *ArrayPathFeatureCollection) Value() interface{} {
	return a.PathFeatureValue()
}

func (a *ArrayPathFeatureCollection) FeatureIDKey() b6.FeatureID {
	return a.Features[a.i-1].FeatureID()
}

func (a *ArrayPathFeatureCollection) PathFeatureValue() b6.PathFeature {
	return a.Features[a.i-1]
}

func (a *ArrayPathFeatureCollection) Next() (bool, error) {
	a.i++
	return a.i <= len(a.Features), nil
}

var _ Collection = &ArrayPathFeatureCollection{}
var _ Countable = &ArrayPathFeatureCollection{}

type ArrayAreaCollection struct {
	Keys   []string
	Values []b6.Area
	i      int
}

func (a *ArrayAreaCollection) Count() int { return len(a.Keys) }

func (a *ArrayAreaCollection) Begin() CollectionIterator {
	return &ArrayAreaCollection{
		Keys:   a.Keys,
		Values: a.Values,
	}
}

func (a *ArrayAreaCollection) Key() interface{} {
	return a.StringKey()
}

func (a *ArrayAreaCollection) Value() interface{} {
	return a.AreaValue()
}

func (a *ArrayAreaCollection) StringKey() string {
	return a.Keys[a.i-1]
}

func (a *ArrayAreaCollection) AreaValue() b6.Area {
	return a.Values[a.i-1]
}

func (a *ArrayAreaCollection) Next() (bool, error) {
	a.i++
	return a.i <= len(a.Keys), nil
}

var _ Collection = &ArrayAreaCollection{}
var _ Countable = &ArrayAreaCollection{}

type ArrayFeatureIDIntCollection struct {
	Keys   []b6.FeatureID
	Values []int
	i      int
}

func (a *ArrayFeatureIDIntCollection) Count() int { return len(a.Keys) }

func (a *ArrayFeatureIDIntCollection) Begin() CollectionIterator {
	return &ArrayFeatureIDIntCollection{
		Keys:   a.Keys,
		Values: a.Values,
	}
}

func (a *ArrayFeatureIDIntCollection) Key() interface{} {
	return a.FeatureIDKey()
}

func (a *ArrayFeatureIDIntCollection) Value() interface{} {
	return a.IntValue()
}

func (a *ArrayFeatureIDIntCollection) FeatureIDKey() b6.FeatureID {
	return a.Keys[a.i-1]
}

func (a *ArrayFeatureIDIntCollection) IntValue() int {
	return a.Values[a.i-1]
}

func (a *ArrayFeatureIDIntCollection) Next() (bool, error) {
	a.i++
	return a.i <= len(a.Keys), nil
}

var _ Collection = &ArrayFeatureIDIntCollection{}
var _ Countable = &ArrayFeatureIDIntCollection{}

type ArrayFeatureIDStringCollection struct {
	Keys   []b6.FeatureID
	Values []string
	i      int
}

func (a *ArrayFeatureIDStringCollection) Count() int { return len(a.Keys) }

func (a *ArrayFeatureIDStringCollection) Begin() CollectionIterator {
	return &ArrayFeatureIDStringCollection{
		Keys:   a.Keys,
		Values: a.Values,
		i:      0,
	}
}

func (a *ArrayFeatureIDStringCollection) Key() interface{} {
	return a.FeatureIDKey()
}

func (a *ArrayFeatureIDStringCollection) Value() interface{} {
	return a.StringValue()
}

func (a *ArrayFeatureIDStringCollection) FeatureIDKey() b6.FeatureID {
	return a.Keys[a.i-1]
}

func (a *ArrayFeatureIDStringCollection) StringValue() string {
	return a.Values[a.i-1]
}

func (a *ArrayFeatureIDStringCollection) Next() (bool, error) {
	a.i++
	return a.i <= len(a.Keys), nil
}

var _ Collection = &ArrayFeatureIDStringCollection{}
var _ Countable = &ArrayFeatureIDStringCollection{}

type ArrayFeatureIDFeatureIDCollection struct {
	Keys   []b6.FeatureID
	Values []b6.FeatureID
	i      int
}

func (a *ArrayFeatureIDFeatureIDCollection) Count() int { return len(a.Keys) }

func (a *ArrayFeatureIDFeatureIDCollection) Begin() CollectionIterator {
	return &ArrayFeatureIDFeatureIDCollection{
		Keys:   a.Keys,
		Values: a.Values,
	}
}

func (a *ArrayFeatureIDFeatureIDCollection) Key() interface{} {
	return a.FeatureIDKey()
}

func (a *ArrayFeatureIDFeatureIDCollection) Value() interface{} {
	return a.FeatureIDValue()
}

func (a *ArrayFeatureIDFeatureIDCollection) FeatureIDKey() b6.FeatureID {
	return a.Keys[a.i-1]
}

func (a *ArrayFeatureIDFeatureIDCollection) FeatureIDValue() b6.FeatureID {
	return a.Values[a.i-1]
}

func (a *ArrayFeatureIDFeatureIDCollection) Next() (bool, error) {
	a.i++
	return a.i <= len(a.Keys), nil
}

var _ Collection = &ArrayFeatureIDFeatureIDCollection{}
var _ Countable = &ArrayFeatureIDFeatureIDCollection{}

type ArrayAnyFloatCollection struct {
	Keys   []interface{}
	Values []float64
	i      int
}

func (a *ArrayAnyFloatCollection) Count() int { return len(a.Keys) }

func (a *ArrayAnyFloatCollection) Begin() CollectionIterator {
	return &ArrayAnyFloatCollection{
		Keys:   a.Keys,
		Values: a.Values,
		i:      0,
	}
}

func (a *ArrayAnyFloatCollection) Key() interface{} {
	return a.Keys[a.i-1]
}

func (a *ArrayAnyFloatCollection) Value() interface{} {
	return a.FloatValue()
}

func (a *ArrayAnyFloatCollection) FloatValue() float64 {
	return a.Values[a.i-1]
}

func (a *ArrayAnyFloatCollection) Next() (bool, error) {
	a.i++
	return a.i <= len(a.Keys), nil
}

var _ Collection = &ArrayAnyFloatCollection{}
var _ Countable = &ArrayAnyFloatCollection{}

type ArrayAnyIntCollection struct {
	Keys   []interface{}
	Values []int
	i      int
}

func (a *ArrayAnyIntCollection) IsCountable() bool { return true }
func (a *ArrayAnyIntCollection) Count() int        { return len(a.Keys) }

func (a *ArrayAnyIntCollection) Begin() CollectionIterator {
	return &ArrayAnyIntCollection{
		Keys:   a.Keys,
		Values: a.Values,
	}
}

func (a *ArrayAnyIntCollection) Key() interface{} {
	return a.Keys[a.i-1]
}

func (a *ArrayAnyIntCollection) Value() interface{} {
	return a.IntValue()
}

func (a *ArrayAnyIntCollection) IntValue() int {
	return a.Values[a.i-1]
}

func (a *ArrayAnyIntCollection) Next() (bool, error) {
	a.i++
	return a.i <= len(a.Keys), nil
}

var _ Collection = &ArrayAnyIntCollection{}
var _ Countable = &ArrayAnyIntCollection{}

type ArrayTagCollection struct {
	Tags []b6.Tag
	i    int
}

func (a *ArrayTagCollection) IsCountable() bool { return true }
func (a *ArrayTagCollection) Count() int        { return len(a.Tags) }

func (a *ArrayTagCollection) Begin() CollectionIterator {
	return &ArrayTagCollection{
		Tags: a.Tags,
	}
}

func (a *ArrayTagCollection) Key() interface{} {
	return a.i
}

func (a *ArrayTagCollection) Value() interface{} {
	return a.TagValue()
}

func (a *ArrayTagCollection) TagValue() b6.Tag {
	return a.Tags[a.i-1]
}

func (a *ArrayTagCollection) Next() (bool, error) {
	a.i++
	return a.i <= len(a.Tags), nil
}

var _ Collection = &ArrayTagCollection{}
var _ Countable = &ArrayTagCollection{}

func FillMap(c Collection, toFill interface{}) error {
	f := reflect.ValueOf(toFill)
	if f.Kind() != reflect.Map {
		return fmt.Errorf("expected a map, found %T", f)
	}
	kt := f.Type().Key()
	vt := f.Type().Elem()
	i := c.Begin()
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

func FillSliceFromValues(c Collection, toFill interface{}) error {
	f := reflect.ValueOf(toFill)
	if f.Kind() != reflect.Ptr || f.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("expected a pointer to a slice, found %T", toFill)
	}
	f = f.Elem()
	vt := f.Type().Elem()
	i := c.Begin()
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
