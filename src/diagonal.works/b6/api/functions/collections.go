package functions

import (
	"container/heap"
	"fmt"
	"math/rand"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
)

// Return a collection of the given key value pairs.
func collection(_ *api.Context, pairs ...interface{}) (b6.Collection[any, any], error) {
	c := &b6.ArrayCollection[interface{}, interface{}]{
		Keys:   make([]interface{}, len(pairs)),
		Values: make([]interface{}, len(pairs)),
	}
	for i, arg := range pairs {
		if pair, ok := arg.(api.Pair); ok {
			c.Keys[i] = pair.First()
			c.Values[i] = pair.Second()
		} else {
			return b6.Collection[any, any]{}, fmt.Errorf("Expected a pair, found %T", arg)
		}
	}
	return c.Collection(), nil
}

type takeCollection struct {
	c b6.UntypedCollection
	i b6.Iterator[any, any]
	n int
	r int
}

func (t *takeCollection) Begin() b6.Iterator[any, any] {
	return &takeCollection{c: t.c, i: t.c.BeginUntyped(), n: t.n, r: t.n}
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

func (t *takeCollection) Count() (int, bool) {
	if n, ok := t.c.Count(); ok {
		if n < t.n {
			return n, true
		} else {
			return t.n, true
		}
	}
	return 0, false
}

var _ b6.AnyCollection[any, any] = &takeCollection{}

// Return a collection with the first n entries of the given collection.
func take(_ *api.Context, collection b6.UntypedCollection, n int) (b6.Collection[any, any], error) {
	return b6.Collection[any, any]{AnyCollection: &takeCollection{c: collection, n: n}}, nil
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

// Return a collection with the n entries from the given collection with the greatest values.
// Requires the values of the given collection to be integers or floats.
func top(_ *api.Context, collection b6.UntypedCollection, n int) (b6.Collection[any, any], error) {
	i := collection.BeginUntyped()
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
				return b6.Collection[any, any]{}, fmt.Errorf("Can't order values of type %T", i.Value())
			}
			first = false
		} else {
			if float {
				if _, ok := i.Value().(float64); !ok {
					return b6.Collection[any, any]{}, fmt.Errorf("Expected float64, found %T", i.Value())
				}
			} else {
				if _, ok := i.Value().(int); !ok {
					return b6.Collection[any, any]{}, fmt.Errorf("Expected int, found %T", i.Value())
				}
			}
		}
		heap.Push(h, api.AnyAnyPair{i.Key(), i.Value()})
		if h.Len() > n {
			heap.Pop(h)
		}
	}
	r := b6.ArrayCollection[interface{}, interface{}]{
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
	return r.Collection(), err
}

type filterCollection struct {
	c       b6.UntypedCollection
	i       b6.Iterator[any, any]
	f       func(*api.Context, interface{}) (bool, error)
	context *api.Context
}

func (f *filterCollection) Begin() b6.Iterator[any, any] {
	return &filterCollection{c: f.c, i: f.c.BeginUntyped(), f: f.f, context: f.context}
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

func (f *filterCollection) Count() (int, bool) {
	return 0, false
}

var _ b6.AnyCollection[any, any] = &filterCollection{}

// Return a collection of the items of the given collection for which the value of the given function applied to each value is true.
func filter(context *api.Context, collection b6.UntypedCollection, function func(*api.Context, interface{}) (bool, error)) (b6.Collection[any, any], error) {
	return b6.Collection[any, any]{AnyCollection: &filterCollection{c: collection, f: function, context: context}}, nil
}

// Return a collection of the result of summing the values of each item with the same key.
// Requires values to be integers.
func sumByKey(_ *api.Context, c b6.Collection[any, int]) (b6.Collection[any, int], error) {
	counts := make(map[interface{}]int)
	i := c.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return b6.Collection[any, int]{}, err
		}
		if !ok {
			break
		}
		counts[i.Key()] += i.Value()
	}
	r := &b6.ArrayCollection[any, int]{
		Keys:   make([]any, 0, len(counts)),
		Values: make([]int, 0, len(counts)),
	}
	for k, v := range counts {
		r.Keys = append(r.Keys, k)
		r.Values = append(r.Values, v)
	}
	return r.Collection(), nil
}

// Return a collection of the number of occurances of each value in the given collection.
func countValues(_ *api.Context, collection b6.Collection[any, any]) (b6.Collection[any, int], error) {
	counts := make(map[interface{}]int)
	i := collection.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return b6.Collection[any, int]{}, err
		}
		if !ok {
			break
		}
		// TODO: return an error if the value can't be used as a map key
		counts[i.Value()]++
	}
	r := &b6.ArrayCollection[interface{}, int]{
		Keys:   make([]interface{}, 0, len(counts)),
		Values: make([]int, 0, len(counts)),
	}
	for k, v := range counts {
		r.Keys = append(r.Keys, k)
		r.Values = append(r.Values, v)
	}
	return r.Collection(), nil
}

type flattenCollection struct {
	c  b6.Collection[any, b6.UntypedCollection]
	i  b6.Iterator[any, b6.UntypedCollection]
	ii b6.Iterator[any, any]
}

func (f *flattenCollection) Begin() b6.Iterator[any, any] {
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
			f.ii = f.i.Value().BeginUntyped()
		}
		ok, err := f.ii.Next()
		if ok || err != nil {
			return ok, err
		} else {
			f.ii = nil
		}
	}
}

func (f *flattenCollection) Count() (int, bool) {
	return 0, false
}

var _ b6.AnyCollection[any, any] = &flattenCollection{}

// Return a collection with keys and values taken from the collections that form the values of the given collection.
func flatten(_ *api.Context, collection b6.Collection[any, b6.UntypedCollection]) (b6.Collection[any, any], error) {
	return b6.Collection[any, any]{
		AnyCollection: &flattenCollection{c: collection},
	}, nil
}

// Return a change that adds a histogram for the given collection.
func histogram(c *api.Context, collection b6.Collection[any, any]) (ingest.Change, error) {
	id := b6.CollectionID{Namespace: b6.NamespaceUI, Value: rand.Uint64()}

	histogram, err := api.NewHistogramFromCollection(collection, id)
	if err != nil {
		return nil, err
	}
	return &ingest.AddFeatures{histogram}, nil
}

// Return a change that adds a histogram with only colour swatches for the given collection.
func histogramSwatch(c *api.Context, collection b6.Collection[any, any]) (ingest.Change, error) {
	id := b6.CollectionID{Namespace: b6.NamespaceUI, Value: rand.Uint64()}

	histogram, err := api.NewHistogramFromCollection(collection, id)
	if err != nil {
		return nil, err
	}
	histogram.AddTag(b6.Tag{Key: "b6:histogram", Value: b6.String("swatch")})
	return &ingest.AddFeatures{histogram}, nil
}
