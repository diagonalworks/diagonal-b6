package functions

import (
	"container/heap"
	"fmt"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"github.com/golang/geo/s2"
)

func emptyPointCollection(context *api.Context) (api.StringPointCollection, error) {
	return &api.ArrayPointCollection{Keys: []string{}, Values: []s2.Point{}}, nil
}

func addPoint(context *api.Context, p b6.Point, c api.StringPointCollection) (api.StringPointCollection, error) {
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

func collection(_ *api.Context, pairs ...interface{}) (api.Collection, error) {
	c := &api.ArrayAnyCollection{
		Keys:   make([]interface{}, len(pairs)),
		Values: make([]interface{}, len(pairs)),
	}
	for i, arg := range pairs {
		if pair, ok := arg.(api.Pair); ok {
			c.Keys[i] = pair.First()
			c.Values[i] = pair.Second()
		} else {
			return nil, fmt.Errorf("Expected a pair, found %T", arg)
		}
	}
	return c, nil
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

func take(_ *api.Context, c api.Collection, n int) (api.Collection, error) {
	return &takeCollection{c: c, n: n}, nil
}

// TODO: Don't just use anyanypair, use an int.
type topFloatHeap []api.AnyAnyPair

func (h topFloatHeap) Len() int           { return len(h) }
func (h topFloatHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h topFloatHeap) Less(i, j int) bool { return h[i][1].(float64) < h[j][1].(float64) }
func (h *topFloatHeap) Push(x any) {
	*h = append(*h, x.(api.AnyAnyPair))
}
func (h *topFloatHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type topIntHeap []api.AnyAnyPair

func (h topIntHeap) Len() int           { return len(h) }
func (h topIntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h topIntHeap) Less(i, j int) bool { return h[i][1].(int) < h[j][1].(int) }
func (h *topIntHeap) Push(x any) {
	*h = append(*h, x.(api.AnyAnyPair))
}
func (h *topIntHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func top(_ *api.Context, c api.Collection, n int) (api.Collection, error) {
	i := c.Begin()
	var err error
	first := true
	float := false
	var h heap.Interface
	for {
		var ok bool
		ok, err = i.Next()
		if !ok || err != nil {
			break
		}
		if first {
			switch i.Value().(type) {
			case int:
				float = false
				ih := make(topIntHeap, 0, 8)
				h = &ih
			case float64:
				float = true
				fh := make(topFloatHeap, 0, 8)
				h = &fh
			default:
				return nil, fmt.Errorf("Can't order values of type %T", i.Value())
			}
			first = false
		} else {
			if float {
				if _, ok := i.Value().(float64); !ok {
					return nil, fmt.Errorf("Expected float64, found %T", i.Value())
				}
			} else {
				if _, ok := i.Value().(int); !ok {
					return nil, fmt.Errorf("Expected int, found %T", i.Value())
				}
			}
		}
		heap.Push(h, api.AnyAnyPair{i.Key(), i.Value()})
		if h.Len() > n {
			heap.Pop(h)
		}
	}
	r := api.ArrayAnyCollection{
		Keys:   make([]interface{}, h.Len()),
		Values: make([]interface{}, h.Len()),
	}
	j := 0
	for h.Len() > 0 {
		p := heap.Pop(h).(api.AnyAnyPair)
		r.Keys[len(r.Keys)-1-j] = p[0]
		r.Values[len(r.Values)-1-j] = p[1]
		j++
	}
	return &r, err
}

type filterCollection struct {
	c       api.Collection
	i       api.CollectionIterator
	f       func(*api.Context, interface{}) (bool, error)
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
		ok, err = f.f(f.context, f.i.Value())
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

func filter(context *api.Context, c api.Collection, f func(*api.Context, interface{}) (bool, error)) (api.Collection, error) {
	return &filterCollection{c: c, f: f, context: context}, nil
}

func sumByKey(_ *api.Context, c api.Collection) (api.Collection, error) {
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

func countValues(_ *api.Context, c api.Collection) (api.Collection, error) {
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

type flattenCollection struct {
	c  api.Collection
	i  api.CollectionIterator
	ii api.CollectionIterator
}

func (f *flattenCollection) Begin() api.CollectionIterator {
	return &flattenCollection{c: f.c, i: f.c.Begin()}
}

func (f *flattenCollection) Key() interface{} {
	return f.ii.Key()
}

func (f *flattenCollection) Value() interface{} {
	return f.ii.Value()
}

func (f *flattenCollection) Next() (bool, error) {
	for {
		if f.ii == nil {
			ok, err := f.i.Next()
			if !ok || err != nil {
				return ok, err
			}
			if c, ok := f.i.Value().(api.Collection); ok {
				f.ii = c.Begin()
			} else {
				return false, fmt.Errorf("flatten: expected a collection, found %T", f.i.Value())
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

var _ api.Collection = &flattenCollection{}

func flatten(_ *api.Context, c api.Collection) (api.Collection, error) {
	return &flattenCollection{c: c}, nil
}
