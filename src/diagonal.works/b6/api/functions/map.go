package functions

import (
	"fmt"

	"diagonal.works/b6/api"
)

type mapCollection struct {
	f       func(interface{}, *api.Context) (interface{}, error)
	v       interface{}
	i       api.CollectionIterator
	c       api.Collection
	context *api.Context
}

func (v *mapCollection) Begin() api.CollectionIterator {
	return &mapCollection{f: v.f, i: v.c.Begin(), c: v.c, context: v.context}
}

func (v *mapCollection) Count() int {
	return api.Count(v.c)
}

func (v *mapCollection) Next() (bool, error) {
	ok, err := v.i.Next()
	if ok && err == nil {
		v.v, err = v.f(v.i.Value(), v.context)
	}
	return ok, err
}

func (v *mapCollection) Key() interface{} {
	return v.i.Key()
}

func (v *mapCollection) Value() interface{} {
	return v.v
}

func map_(collection api.Collection, f func(interface{}, *api.Context) (interface{}, error), context *api.Context) (api.Collection, error) {
	return &mapCollection{c: collection, f: f, context: context}, nil
}

type mapItemsCollection struct {
	f       func(api.Pair, *api.Context) (interface{}, error)
	k       interface{}
	v       interface{}
	i       api.CollectionIterator
	c       api.Collection
	context *api.Context
}

func (v *mapItemsCollection) Begin() api.CollectionIterator {
	return &mapItemsCollection{f: v.f, i: v.c.Begin(), c: v.c, context: v.context}
}

func (v *mapItemsCollection) Count() int {
	return api.Count(v.c)
}

func (v *mapItemsCollection) Next() (bool, error) {
	ok, err := v.i.Next()
	if ok && err == nil {
		pair := api.AnyAnyPair{v.i.Key(), v.i.Value()}
		var r interface{}
		r, err = v.f(pair, v.context)
		if err == nil {
			if pair, ok := r.(api.Pair); ok {
				v.k = pair.First()
				v.v = pair.Second()
			} else {
				err = fmt.Errorf("expected a pair, found %T", r)
			}
		}
	}
	return ok, err
}

func (v *mapItemsCollection) Key() interface{} {
	return v.k
}

func (v *mapItemsCollection) Value() interface{} {
	return v.v
}

func mapItems(collection api.Collection, f func(api.Pair, *api.Context) (interface{}, error), context *api.Context) (api.Collection, error) {
	return &mapItemsCollection{c: collection, f: f, context: context}, nil
}

func pair(first interface{}, second interface{}, c *api.Context) (api.Pair, error) {
	return api.AnyAnyPair{first, second}, nil
}

func first(pair api.Pair, c *api.Context) (interface{}, error) {
	return pair.First(), nil
}

func second(pair api.Pair, c *api.Context) (interface{}, error) {
	return pair.Second(), nil
}
