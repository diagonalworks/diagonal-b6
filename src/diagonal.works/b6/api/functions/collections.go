package functions

import (
	"container/heap"
	"fmt"
	"sort"

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

type bound struct {
	lower, upper interface{}
	prettyPrint  interface{}
	within       func(interface{}) (bool, error)
}
type buckets []bound

func xBound(lower, upper interface{}) bound {
	return bound{
		lower:       lower,
		upper:       upper,
		prettyPrint: func() interface{} { return fmt.Sprint(lower) + "-" + fmt.Sprint(upper) },
		within: func(v interface{}) (bool, error) {
			below, err := api.Less(lower, v)
			if err != nil {
				return false, err
			}

			above, err := api.Greater(upper, v)
			if err != nil {
				return false, err
			}

			return !below && !above, nil
		},
	}
}

func (b buckets) bucket(v interface{}) (int, error) {
	for i, bound := range b {
		yes, err := bound.within(v)
		if err != nil {
			return -1, err
		}
		if yes {
			return i, nil
		}
	}

	return -1, nil // Allowing data not to be bucketed.
}

const maxBuckets = 5

func categorical(kvs []kv) (buckets, error) {
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].value > kvs[j].value
	})

	var b buckets
	for i, kv := range kvs {
		if i == maxBuckets {
			break
		}

		key := kv.key
		if i == maxBuckets-1 && len(kvs) > maxBuckets {
			key = "other"
		}

		b = append(b, bound{lower: key, prettyPrint: key, within: func(v interface{}) (bool, error) { return v == key || key == "other", nil }})
	}

	return b, nil
}

func uniform(kvs []kv) (buckets, error) {
	sort.Slice(kvs, func(i, j int) bool {
		greater, err := api.Greater(kvs[i].key, kvs[j].key)
		if err != nil {
			panic(err) // Not graceful, but greater should handle all numericals,
		} // and we do the numerical check in histogram call.

		return greater
	})

	var b buckets
	if (len(kvs)) <= maxBuckets {
		for _, kv := range kvs {
			b = append(b, bound{lower: kv.key, prettyPrint: kv.key, within: func(v interface{}) (bool, error) { return v == kv.key, nil }})
		}
	} else {
		for len(kvs) > 0 {
			bucket_size := len(kvs) / (maxBuckets - len(b))
			b = append(b, xBound(kvs[0].key, kvs[bucket_size-1].key))
			kvs = kvs[bucket_size:]
		}
	}

	return b, nil
}

type kv struct {
	key   interface{}
	value int // key-count.
}

func countvalues(c api.Collection) ([]kv, error) {
	m := make(map[interface{}]int)

	i := c.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}

		m[i.Value()]++
	}

	var kvs []kv
	for k, v := range m {
		kvs = append(kvs, kv{k, v})
	}

	return kvs, nil
}

func numerical(any interface{}) bool {
	switch any.(type) {
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64:
		return true
	default: // Ignoring complex numbers.
		return false
	}
}

func histogram(co *api.Context, c api.Collection) (api.Collection, error) {
	values, err := countvalues(c)
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return &api.ArrayAnyIntCollection{}, nil
	}

	var b buckets
	if numerical(values[0].key) { // Checking only first, assuming all elements in collection have the same type. TODO: enforce/check that somewhere else.
		b, err = uniform(values)
		if err != nil {
			return nil, err
		}
	} else {
		b, err = categorical(values)
		if err != nil {
			return nil, err
		}
	}

	h := api.ArrayAnyIntCollection{
		Keys:   make([]interface{}, len(b)),
		Values: make([]int, len(b)),
	}

	for _, kv := range values {
		index, err := b.bucket(kv.key)
		if err != nil {
			return nil, err
		}
		if index == -1 {
			continue
		}

		h.Keys[index] = b[index].prettyPrint
		h.Values[index] += kv.value
	}

	return &h, nil
}
