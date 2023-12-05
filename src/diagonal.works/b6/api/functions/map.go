package functions

import (
	"fmt"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"golang.org/x/sync/errgroup"
)

type mapCollection struct {
	f       func(*api.Context, interface{}) (interface{}, error)
	v       interface{}
	c       b6.UntypedCollection
	i       b6.Iterator[any, any]
	context *api.Context
}

func (v *mapCollection) Begin() b6.Iterator[any, any] {
	return &mapCollection{f: v.f, c: v.c, i: v.c.BeginUntyped(), context: v.context}
}

func (v *mapCollection) Count() (int, bool) {
	return v.c.Count()
}

func (v *mapCollection) Next() (bool, error) {
	var ok bool
	var err error
	if err = v.context.Context.Err(); err == nil {
		ok, err = v.i.Next()
		if ok && err == nil {
			v.v, err = v.f(v.context, v.i.Value())
		}
	}
	return ok, err
}

func (v *mapCollection) Key() interface{} {
	return v.i.Key()
}

func (v *mapCollection) Value() interface{} {
	return v.v
}

// Return a collection with the result of applying the given function to each value.
// Keys are unmodified.
func map_(context *api.Context, collection b6.UntypedCollection, function func(*api.Context, interface{}) (interface{}, error)) (b6.Collection[any, any], error) {
	return b6.Collection[any, any]{
		AnyCollection: &mapCollection{c: collection, f: function, context: context},
	}, nil
}

type mapItemsCollection struct {
	f       func(*api.Context, api.Pair) (interface{}, error)
	k       interface{}
	v       interface{}
	c       b6.UntypedCollection
	i       b6.Iterator[any, any]
	context *api.Context
}

func (v *mapItemsCollection) Begin() b6.Iterator[any, any] {
	return &mapItemsCollection{f: v.f, i: v.c.BeginUntyped(), c: v.c, context: v.context}
}

func (v *mapItemsCollection) Count() (int, bool) {
	return v.c.Count()
}

func (v *mapItemsCollection) Next() (bool, error) {
	ok, err := v.i.Next()
	if ok && err == nil {
		pair := api.AnyAnyPair{v.i.Key(), v.i.Value()}
		var r interface{}
		r, err = v.f(v.context, pair)
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

// Return a collection of the result of applying the given function to each pair(key, value).
// Keys are unmodified.
func mapItems(context *api.Context, collection b6.UntypedCollection, function func(*api.Context, api.Pair) (interface{}, error)) (b6.Collection[any, any], error) {
	return b6.Collection[any, any]{
		AnyCollection: &mapItemsCollection{c: collection, f: function, context: context},
	}, nil
}

// Return a pair containing the given values.
func pair(c *api.Context, first interface{}, second interface{}) (api.Pair, error) {
	return api.AnyAnyPair{first, second}, nil
}

// Return the first value of the given pair.
func first(c *api.Context, pair api.Pair) (interface{}, error) {
	return pair.First(), nil
}

// Return the second value of the given pair.
func second(c *api.Context, pair api.Pair) (interface{}, error) {
	return pair.Second(), nil
}

type mapParallelCollection struct {
	f       func(*api.Context, interface{}) (interface{}, error)
	v       interface{}
	c       b6.UntypedCollection
	i       b6.Iterator[any, any]
	context *api.Context

	in      []chan api.AnyAnyPair
	out     []chan api.AnyAnyPair
	current api.AnyAnyPair
	err     error
	read    int
}

func (m *mapParallelCollection) Begin() b6.Iterator[any, any] {
	c := &mapParallelCollection{
		f:       m.f,
		c:       m.c,
		i:       m.c.BeginUntyped(),
		context: m.context,

		in:   make([]chan api.AnyAnyPair, m.context.Cores),
		out:  make([]chan api.AnyAnyPair, m.context.Cores),
		read: -1,
	}
	for i := range c.in {
		c.in[i] = make(chan api.AnyAnyPair, 1)
		c.out[i] = make(chan api.AnyAnyPair, 1)
	}
	go c.run()
	return c
}

func (m *mapParallelCollection) Count() (int, bool) {
	return m.c.Count()
}

func (m *mapParallelCollection) Next() (bool, error) {
	m.read++
	var ok bool
	if m.current, ok = <-m.out[m.read%len(m.out)]; ok {
		return true, nil
	}
	return false, m.err
}

func (m *mapParallelCollection) Key() interface{} {
	return m.current.First()
}

func (m *mapParallelCollection) Value() interface{} {
	return m.current.Second()
}

func (m *mapParallelCollection) run() {
	g, c := errgroup.WithContext(m.context.Context)
	vms := m.context.VM.Fork(m.context.Cores)
	contexts := make([]api.Context, m.context.Cores)
	for i := range contexts {
		contexts[i] = *m.context
		contexts[i].Context = c
		contexts[i].VM = &vms[i]
	}
	for i := range m.in {
		in, out, context := m.in[i], m.out[i], &contexts[i]
		g.Go(func() error {
			for pair := range in {
				v, err := m.f(context, pair.Second())
				if err == nil {
					select {
					case out <- api.AnyAnyPair{pair.First(), v}:
					case <-c.Done():
						return nil
					}
				} else {
					return err
				}
			}
			return nil
		})
	}

	g.Go(func() error {
		write := 0
		ok := true
		var err error
		for ok && err == nil {
			ok, err = m.i.Next()
			if ok && err == nil {
				select {
				case m.in[write%len(m.in)] <- api.AnyAnyPair{m.i.Key(), m.i.Value()}:
				case <-c.Done():
					err = c.Err()
				}
				write++
			}
		}
		for i := range m.in {
			close(m.in[i])
		}
		return err
	})

	m.err = g.Wait()
	for i := range m.out {
		close(m.out[i])
	}
}

// Return a collection with the result of applying the given function to each value.
// Keys are unmodified, and function application occurs in parallel, bounded
// by the number of CPU cores allocated to b6.
func mapParallel(context *api.Context, collection b6.UntypedCollection, function func(*api.Context, interface{}) (interface{}, error)) (b6.Collection[any, any], error) {
	if context.Cores < 2 {
		return map_(context, collection, function)
	}
	return b6.Collection[any, any]{
		AnyCollection: &mapParallelCollection{c: collection, f: function, context: context},
	}, nil
}
