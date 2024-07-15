package functions

import (
	"container/heap"
	"fmt"
	"math/rand"
	"reflect"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
)

type pairCollection struct {
	ps []api.Pair
	es []b6.Expression
	i  int
}

func (p *pairCollection) Begin() b6.Iterator[any, any] {
	return &pairCollection{ps: p.ps, es: p.es}
}

func (p *pairCollection) Next() (bool, error) {
	p.i++
	return p.i <= len(p.ps), nil
}

func (p *pairCollection) Key() interface{} {
	return p.ps[p.i-1].First()
}

func (p *pairCollection) Value() interface{} {
	return p.ps[p.i-1].Second()
}

func (p *pairCollection) KeyExpression() b6.Expression {
	return b6.NewCallExpression(
		b6.NewSymbolExpression("first"),
		[]b6.Expression{p.es[p.i-1]},
	)
}

func (p *pairCollection) ValueExpression() b6.Expression {
	return b6.NewCallExpression(
		b6.NewSymbolExpression("second"),
		[]b6.Expression{p.es[p.i-1]},
	)
}

func (p *pairCollection) Count() (int, bool) {
	return len(p.ps) - p.i, true
}

// Return a collection of the given key value pairs.
func collection(context *api.Context, pairs ...interface{}) (b6.Collection[any, any], error) {
	c := &pairCollection{
		ps: make([]api.Pair, len(pairs)),
		es: context.VM.ArgExpressions(),
	}
	for i, arg := range pairs {
		if pair, ok := arg.(api.Pair); ok {
			c.ps[i] = pair
		} else {
			return b6.Collection[any, any]{}, fmt.Errorf("Expected a pair, found %T", arg)
		}
	}
	return b6.Collection[any, any]{AnyCollection: c}, nil
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

func (t *takeCollection) KeyExpression() b6.Expression {
	return t.i.KeyExpression()
}

func (t *takeCollection) ValueExpression() b6.Expression {
	return t.i.ValueExpression()
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
	f       api.Callable
	context *api.Context
}

func (f *filterCollection) Begin() b6.Iterator[any, any] {
	return &filterCollection{c: f.c, i: f.c.BeginUntyped(), f: f.f, context: f.context}
}

func (f *filterCollection) Next() (bool, error) {
	var frames [1]api.StackFrame
	for {
		ok, err := f.i.Next()
		if !ok || err != nil {
			return ok, err
		}
		frames[0].Value = reflect.ValueOf(f.i.Value())
		frames[0].Expression = f.i.ValueExpression()
		r, err := f.context.VM.CallWithArgsAndExpressions(f.context, f.f, frames[0:1])
		if err != nil {
			return false, err
		}
		if b, ok := r.(bool); ok {
			if b {
				return true, nil
			}
		} else {
			return false, fmt.Errorf("expected bool, found %T", r)
		}
	}
}

func (f *filterCollection) Key() interface{} {
	return f.i.Key()
}

func (f *filterCollection) Value() interface{} {
	return f.i.Value()
}

func (f *filterCollection) KeyExpression() b6.Expression {
	return f.i.KeyExpression()
}

func (f *filterCollection) ValueExpression() b6.Expression {
	return f.i.ValueExpression()
}

func (f *filterCollection) Count() (int, bool) {
	return 0, false
}

var _ b6.AnyCollection[any, any] = &filterCollection{}

// Return a collection of the items of the given collection for which the value of the given function applied to each value is true.
func filter(context *api.Context, collection b6.UntypedCollection, function api.Callable) (b6.Collection[any, any], error) {
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

// Return a collection of the number of occurances of each value in the given collection.
func countKeys(_ *api.Context, collection b6.Collection[any, any]) (b6.Collection[any, int], error) {
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
		counts[i.Key()]++
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

// Return a collection of the number of occurances of each valid value in the given collection.
// Invalid values are not counted, but case the key to appear in the output.
func countValidKeys(_ *api.Context, collection b6.Collection[any, any]) (b6.Collection[any, int], error) {
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
		if id, ok := i.Value().(b6.FeatureID); ok {
			if id.IsValid() {
				counts[i.Key()]++
			} else {
				counts[i.Key()] += 0
			}
		} else {
			counts[i.Key()]++
		}
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

func (f *flattenCollection) KeyExpression() b6.Expression {
	return f.ii.KeyExpression()
}

func (f *flattenCollection) ValueExpression() b6.Expression {
	return f.ii.ValueExpression()
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
	return histogramWithID(c, collection, id)
}

// Return a change that adds a histogram for the given collection with the given ID.
func histogramWithID(c *api.Context, collection b6.Collection[any, any], id b6.CollectionID) (ingest.Change, error) {
	histogram, err := api.NewHistogramFromCollection(collection, id)
	if err != nil {
		return nil, err
	}
	expression := &ingest.GenericFeature{
		ID:   b6.FeatureID{b6.FeatureTypeExpression, id.Namespace, id.Value},
		Tags: []b6.Tag{{Key: b6.ExpressionTag, Value: c.VM.Expression()}},
	}
	return &ingest.AddFeatures{histogram, expression}, nil
}

// Return a change that adds a histogram with only colour swatches for the given collection.
func histogramSwatch(c *api.Context, collection b6.Collection[any, any]) (ingest.Change, error) {
	id := b6.CollectionID{Namespace: b6.NamespaceUI, Value: rand.Uint64()}
	return histogramSwatchWithID(c, collection, id)
}

// Return a change that adds a histogram with only colour swatches for the given collection.
func histogramSwatchWithID(c *api.Context, collection b6.Collection[any, any], id b6.CollectionID) (ingest.Change, error) {
	histogram, err := api.NewHistogramFromCollection(collection, id)
	if err != nil {
		return nil, err
	}
	histogram.AddTag(b6.Tag{Key: "b6:histogram", Value: b6.NewStringExpression("swatch")})
	expression := &ingest.GenericFeature{
		ID:   b6.FeatureID{b6.FeatureTypeExpression, id.Namespace, id.Value},
		Tags: []b6.Tag{{Key: b6.ExpressionTag, Value: c.VM.Expression()}},
	}
	return &ingest.AddFeatures{histogram, expression}, nil
}

type joinMissingCollection struct {
	base   b6.UntypedCollection
	joined b6.UntypedCollection
	bi     b6.Iterator[any, any]
	ji     b6.Iterator[any, any]
	bok    bool
	jok    bool
}

func (j *joinMissingCollection) Begin() b6.Iterator[any, any] {
	return &joinMissingCollection{
		base:   j.base,
		joined: j.joined,
	}
}

func (j *joinMissingCollection) Next() (bool, error) {
	// If we've started, advance the iterator with the
	// lesser of the two keys, and then advance the joined iterator
	// if the key's equal
	var err error
	if j.bi == nil {
		j.bi = j.base.BeginUntyped()
		j.ji = j.joined.BeginUntyped()
		if j.bok, err = j.bi.Next(); err == nil {
			j.jok, err = j.ji.Next()
		}
	} else {
		if j.bok && j.jok {
			var less bool
			less, err = b6.Less(j.ji.Key(), j.bi.Key())
			if err == nil {
				if less {
					j.jok, err = j.ji.Next()
				} else {
					j.bok, err = j.bi.Next()
				}
			}
		} else if j.bok {
			j.bok, err = j.bi.Next()
		} else if j.jok {
			j.jok, err = j.ji.Next()
		}
	}

	for j.bok && j.jok && err == nil {
		var equal bool
		equal, err = b6.Equal(j.ji.Key(), j.bi.Key())
		if equal && err == nil {
			j.jok, err = j.ji.Next()
		} else {
			break
		}
	}

	return j.bok || j.jok, err
}

func (j *joinMissingCollection) Key() interface{} {
	return j.iterator().Key()
}

func (j *joinMissingCollection) Value() interface{} {
	return j.iterator().Value()
}

func (j *joinMissingCollection) KeyExpression() b6.Expression {
	return j.iterator().KeyExpression()
}

func (j *joinMissingCollection) ValueExpression() b6.Expression {
	return j.iterator().ValueExpression()
}

func (j *joinMissingCollection) Count() (int, bool) {
	return 0, false
}

func (j *joinMissingCollection) iterator() b6.Iterator[any, any] {
	if j.bok && j.jok {
		less, _ := b6.Less(j.ji.Key(), j.bi.Key())
		if less {
			return j.ji
		} else {
			return j.bi
		}
	} else if j.jok {
		return j.ji
	} else if j.bok {
		return j.bi
	}
	return nil
}

func joinMissing(_ *api.Context, base b6.Collection[any, any], joined b6.Collection[any, any]) (b6.Collection[any, any], error) {
	return b6.Collection[any, any]{
		AnyCollection: &joinMissingCollection{base: base, joined: joined},
	}, nil
}
