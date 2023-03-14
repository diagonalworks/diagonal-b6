package functions

import (
	"fmt"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"github.com/golang/geo/s2"
)

func emptyPointCollection(context *api.Context) (api.StringPointCollection, error) {
	return &api.ArrayPointCollection{Keys: []string{}, Values: []s2.Point{}}, nil
}

func addPoint(p b6.Point, c api.StringPointCollection, context *api.Context) (api.StringPointCollection, error) {
	// TODO: We can potentially fast-path this if c is an arrayPointCollection by
	// modifying the underlying slices, and if not, then add a wrapper that
	// returns the new point when Next() returns false.
	n := 1
	if c, ok := c.(api.Countable); ok {
		n = c.Count()
	}
	keys := make([]string, 0, n+1)
	values := make([]s2.Point, 0, n+1)
	i := c.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		values = append(values, i.Value().(b6.Point).Point())
	}
	ll := s2.LatLngFromPoint(p.Point())
	keys = append(keys, fmt.Sprintf("%f,%f", ll.Lat.Degrees(), ll.Lng.Degrees()))
	values = append(values, p.Point())
	return &api.ArrayPointCollection{Keys: keys, Values: values}, nil
}

type singletonCollection struct {
	k    interface{}
	v    interface{}
	done bool
}

func (s *singletonCollection) Count() int { return 1 }

func (s *singletonCollection) Begin() api.CollectionIterator {
	return &singletonCollection{k: s.k, v: s.v}
}

func (s *singletonCollection) Key() interface{} {
	return s.k
}

func (s *singletonCollection) Value() interface{} {
	return s.v
}

func (s *singletonCollection) Next() (bool, error) {
	ok := !s.done
	s.done = true
	return ok, nil
}

var _ api.Collection = &singletonCollection{}
var _ api.Countable = &singletonCollection{}

func collection(k interface{}, v interface{}, _ *api.Context) (api.Collection, error) {
	return &singletonCollection{k: k, v: v}, nil
}

type takeCollection struct {
	c api.Collection
	i api.CollectionIterator
	n int
	r int
}

func (t *takeCollection) Begin() api.CollectionIterator {
	return &takeCollection{c: t.c, i: t.c.Begin(), n: t.n, r: t.n}
}

func (t *takeCollection) Next() (bool, error) {
	if t.r > 0 {
		t.r--
		return t.i.Next()
	}
	return false, nil
}

func (t *takeCollection) Key() interface{} {
	return t.i.Key()
}

func (t *takeCollection) Value() interface{} {
	return t.i.Value()
}

var _ api.Collection = &takeCollection{}

func take(c api.Collection, n int, _ *api.Context) (api.Collection, error) {
	return &takeCollection{c: c, n: n}, nil
}

type filterCollection struct {
	c       api.Collection
	i       api.CollectionIterator
	f       func(interface{}, *api.Context) (bool, error)
	context *api.Context
}

func (f *filterCollection) Begin() api.CollectionIterator {
	return &filterCollection{c: f.c, i: f.c.Begin(), f: f.f, context: f.context}
}

func (f *filterCollection) Next() (bool, error) {
	for {
		ok, err := f.i.Next()
		if !ok || err != nil {
			return ok, err
		}
		ok, err = f.f(f.i.Value(), f.context)
		if ok || err != nil {
			return ok, err
		}
	}
}

func (f *filterCollection) Key() interface{} {
	return f.i.Key()
}

func (f *filterCollection) Value() interface{} {
	return f.i.Value()
}

var _ api.Collection = &filterCollection{}

func filter(c api.Collection, f func(interface{}, *api.Context) (bool, error), context *api.Context) (api.Collection, error) {
	return &filterCollection{c: c, f: f, context: context}, nil
}

func sumByKey(c api.Collection, _ *api.Context) (api.Collection, error) {
	counts := make(map[interface{}]int)
	i := c.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		if n, ok := i.Value().(int); ok {
			counts[i.Key()] += n
		}
	}
	r := &api.ArrayAnyIntCollection{
		Keys:   make([]interface{}, 0, len(counts)),
		Values: make([]int, 0, len(counts)),
	}
	for k, v := range counts {
		r.Keys = append(r.Keys, k)
		r.Values = append(r.Values, v)
	}
	return r, nil
}

func countValues(c api.Collection, _ *api.Context) (api.Collection, error) {
	counts := make(map[interface{}]int)
	i := c.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		// TODO: return an error if the value can't be used as a map key
		counts[i.Value()]++
	}
	r := &api.ArrayAnyIntCollection{
		Keys:   make([]interface{}, 0, len(counts)),
		Values: make([]int, 0, len(counts)),
	}
	for k, v := range counts {
		r.Keys = append(r.Keys, k)
		r.Values = append(r.Values, v)
	}
	return r, nil
}

type flatternCollection struct {
	c  api.Collection
	i  api.CollectionIterator
	ii api.CollectionIterator
}

func (f *flatternCollection) Begin() api.CollectionIterator {
	return &flatternCollection{c: f.c, i: f.c.Begin()}
}

func (f *flatternCollection) Key() interface{} {
	return f.ii.Key()
}

func (f *flatternCollection) Value() interface{} {
	return f.ii.Value()
}

func (f *flatternCollection) Next() (bool, error) {
	for {
		if f.ii == nil {
			ok, err := f.i.Next()
			if !ok || err != nil {
				return ok, err
			}
			if c, ok := f.i.Value().(api.Collection); ok {
				f.ii = c.Begin()
			} else {
				return false, fmt.Errorf("flattern: expected a collection, found %T", f.i.Value())
			}
		}
		ok, err := f.ii.Next()
		if ok || err != nil {
			return ok, err
		} else {
			f.ii = nil
		}
	}
}

var _ api.Collection = &flatternCollection{}

func flattern(c api.Collection, _ *api.Context) (api.Collection, error) {
	return &flatternCollection{c: c}, nil
}
