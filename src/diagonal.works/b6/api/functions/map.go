package functions

import (
	"fmt"
	"reflect"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"golang.org/x/sync/errgroup"
)

type mapCollection struct {
	c       b6.UntypedCollection
	f       api.Callable
	e       b6.Expression
	v       interface{}
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
	var frames [1]api.StackFrame
	if err = v.context.Context.Err(); err == nil {
		ok, err = v.i.Next()
		if ok && err == nil {
			frames[0].Value = reflect.ValueOf(v.i.Value())
			frames[0].Expression = v.i.ValueExpression()
			v.v, err = v.context.VM.CallWithArgsAndExpressions(v.context, v.f, frames[0:1])
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

func (v *mapCollection) KeyExpression() b6.Expression {
	return v.i.KeyExpression()
}

func (v *mapCollection) ValueExpression() b6.Expression {
	return b6.NewCallExpression(v.e, []b6.Expression{v.i.ValueExpression()})
}

// Return a collection with the result of applying the given function to each value.
// Keys are unmodified.
func map_(context *api.Context, collection b6.UntypedCollection, function api.Callable) (b6.Collection[any, any], error) {
	e := context.VM.ArgExpressions()[1]
	return b6.Collection[any, any]{
		AnyCollection: &mapCollection{c: collection, f: function, e: e, context: context},
	}, nil
}

type mapItemsCollection struct {
	c       b6.UntypedCollection
	f       api.Callable
	e       b6.Expression
	k       interface{}
	v       interface{}
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
	var frames [1]api.StackFrame
	if ok && err == nil {
		frames[0].Value = reflect.ValueOf(api.AnyAnyPair{v.i.Key(), v.i.Value()})
		frames[0].Expression = v.i.ValueExpression()
		var r interface{}
		r, err = v.context.VM.CallWithArgsAndExpressions(v.context, v.f, frames[0:1])
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

func (v *mapItemsCollection) KeyExpression() b6.Expression {
	return v.i.KeyExpression()
}

func (v *mapItemsCollection) ValueExpression() b6.Expression {
	return b6.NewCallExpression(v.e, []b6.Expression{
		b6.NewCallExpression(
			b6.NewSymbolExpression("pair"),
			[]b6.Expression{
				v.i.KeyExpression(),
				v.i.ValueExpression(),
			},
		),
	})
}

// Return a collection of the result of applying the given function to each pair(key, value).
// Keys are unmodified.
func mapItems(context *api.Context, collection b6.UntypedCollection, function api.Callable) (b6.Collection[any, any], error) {
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

type expressionPair struct {
	Key   interface{}
	Value interface{}

	KeyExpression   b6.Expression
	ValueExpression b6.Expression
}

type mapParallelCollection struct {
	c       b6.UntypedCollection
	f       api.Callable
	e       b6.Expression
	v       interface{}
	i       b6.Iterator[any, any]
	context *api.Context

	in      []chan expressionPair
	out     []chan expressionPair
	current expressionPair
	err     error
	read    int
}

func (m *mapParallelCollection) Begin() b6.Iterator[any, any] {
	c := &mapParallelCollection{
		f:       m.f,
		c:       m.c,
		i:       m.c.BeginUntyped(),
		context: m.context,

		in:   make([]chan expressionPair, m.context.Cores),
		out:  make([]chan expressionPair, m.context.Cores),
		read: -1,
	}
	for i := range c.in {
		c.in[i] = make(chan expressionPair, 1)
		c.out[i] = make(chan expressionPair, 1)
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
	return m.current.Key
}

func (m *mapParallelCollection) Value() interface{} {
	return m.current.Value
}

func (m *mapParallelCollection) KeyExpression() b6.Expression {
	return m.current.KeyExpression
}

func (m *mapParallelCollection) ValueExpression() b6.Expression {
	return b6.NewCallExpression(m.e, []b6.Expression{m.current.ValueExpression})
}

func (m *mapParallelCollection) run() {
	g, c := errgroup.WithContext(m.context.Context)
	contexts := m.context.Fork(m.context.Cores)
	for i := range m.in {
		in, out, context := m.in[i], m.out[i], contexts[i]
		g.Go(func() error {
			var frames [1]api.StackFrame
			for pair := range in {
				frames[0].Value = reflect.ValueOf(pair.Value)
				frames[0].Expression = pair.ValueExpression
				v, err := context.VM.CallWithArgsAndExpressions(context, m.f, frames[0:1])
				if err == nil {
					select {
					case out <- expressionPair{
						Key:             pair.Key,
						Value:           v,
						KeyExpression:   pair.KeyExpression,
						ValueExpression: pair.ValueExpression,
					}:
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
				case m.in[write%len(m.in)] <- expressionPair{
					Key:             m.i.Key(),
					Value:           m.i.Value(),
					KeyExpression:   m.i.KeyExpression(),
					ValueExpression: m.i.ValueExpression(),
				}:
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
func mapParallel(context *api.Context, collection b6.UntypedCollection, function api.Callable) (b6.Collection[any, any], error) {
	if context.Cores < 2 {
		return map_(context, collection, function)
	}
	return b6.Collection[any, any]{
		AnyCollection: &mapParallelCollection{c: collection, f: function, context: context},
	}, nil
}
