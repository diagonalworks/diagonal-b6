package b6

import (
	"fmt"
)

type Iterator[Key any, Value any] interface {
	Key() Key
	Value() Value
	Next() (bool, error)

	KeyExpression() Expression
	ValueExpression() Expression
}

type emptyIterator[Key any, Value any] struct{}

func (e emptyIterator[Key, _]) Key() Key {
	panic("Key() on emptyIterator")
}

func (e emptyIterator[_, Value]) Value() Value {
	panic("Value() on emptyIterator")
}

func (e emptyIterator[_, _]) Next() (bool, error) {
	return false, nil
}

func (e emptyIterator[_, _]) KeyExpression() Expression {
	panic("KeyExpression() on emptyIterator")
}

func (e emptyIterator[_, _]) ValueExpression() Expression {
	panic("ValueExpression() on emptyIterator")
}

type AnyCollection[Key any, Value any] interface {
	Begin() Iterator[Key, Value]
	Count() (int, bool)
}

type UntypedCollection interface {
	BeginUntyped() Iterator[any, any]
	// Count() returns (length of the collection, true) is the length
	// is trivially computed, otherwise it returns (_, false)
	Count() (int, bool)
}

type Collection[Key any, Value any] struct {
	AnyCollection[Key, Value]
}

func (c Collection[Key, Value]) Begin() Iterator[Key, Value] {
	if c.AnyCollection == nil {
		return emptyIterator[Key, Value]{}
	}
	return c.AnyCollection.Begin()
}

type adaptBeginUntyped[Key any, Value any] struct {
	i Iterator[Key, Value]
}

func (a adaptBeginUntyped[_, _]) Key() any {
	return a.i.Key()
}

func (a adaptBeginUntyped[_, _]) Value() any {
	return a.i.Value()
}

func (a adaptBeginUntyped[_, _]) Next() (bool, error) {
	return a.i.Next()
}

func (a adaptBeginUntyped[_, _]) KeyExpression() Expression {
	return a.i.KeyExpression()
}

func (a adaptBeginUntyped[_, _]) ValueExpression() Expression {
	return a.i.ValueExpression()
}

func (c Collection[Key, Value]) BeginUntyped() Iterator[any, any] {
	return adaptBeginUntyped[Key, Value]{
		i: c.AnyCollection.Begin(),
	}
}

type adaptBeginValues[Key any, Value any] struct {
	i Iterator[Key, Value]
}

func (a adaptBeginValues[_, _]) Key() any {
	return a.i.Key()
}

func (a adaptBeginValues[_, Value]) Value() Value {
	return a.i.Value()
}

func (a adaptBeginValues[_, _]) Next() (bool, error) {
	return a.i.Next()
}

func (a adaptBeginValues[_, _]) KeyExpression() Expression {
	return a.i.KeyExpression()
}

func (a adaptBeginValues[_, Value]) ValueExpression() Expression {
	return a.i.ValueExpression()
}

func (c Collection[Key, Value]) BeginValues() Iterator[any, Value] {
	return adaptBeginValues[Key, Value]{
		i: c.AnyCollection.Begin(),
	}
}

func (c Collection[_, _]) Count() (int, bool) {
	return c.AnyCollection.Count()
}

type adaptValues[Key any, Value any] struct {
	c Collection[Key, Value]
}

func (a adaptValues[_, Value]) Begin() Iterator[any, Value] {
	return a.c.BeginValues()
}

func (a adaptValues[_, Value]) Count() (int, bool) {
	return a.c.Count()
}

func (c Collection[Key, Value]) Values() Collection[any, Value] {
	return Collection[any, Value]{
		AnyCollection: adaptValues[Key, Value]{c: c},
	}
}

func (c Collection[Key, _]) AllKeys(keys []Key) ([]Key, error) {
	i := c.Begin()
	var err error
	for {
		var ok bool
		if ok, err = i.Next(); !ok || err != nil {
			break
		}
		keys = append(keys, i.Key())
	}
	return keys, err
}

func (c Collection[_, Value]) AllValues(values []Value) ([]Value, error) {
	i := c.Begin()
	var err error
	for {
		var ok bool
		if ok, err = i.Next(); !ok || err != nil {
			break
		}
		values = append(values, i.Value())
	}
	return values, err
}

var _ UntypedCollection = Collection[int, string]{}

type adaptCollection[Key any, Value any] struct {
	UntypedCollection
}

func (a adaptCollection[Key, Value]) Begin() Iterator[Key, Value] {
	return &adaptIterator[Key, Value]{i: a.UntypedCollection.BeginUntyped()}
}

func (a adaptCollection[_, _]) Count() (int, bool) {
	return a.UntypedCollection.Count()
}

type adaptIterator[Key any, Value any] struct {
	i Iterator[any, any]
	k Key
	v Value
}

func (a *adaptIterator[Key, Value]) Next() (bool, error) {
	ok, err := a.i.Next()
	if !ok || err != nil {
		return ok, err
	}
	if a.k, ok = a.i.Key().(Key); !ok {
		return false, fmt.Errorf("Expected key %T, found %T", a.k, a.i.Key())
	}
	if a.v, ok = a.i.Value().(Value); !ok {
		return false, fmt.Errorf("Expected value %T, found %T", a.v, a.i.Value())
	}
	return true, nil
}

func (a *adaptIterator[Key, _]) Key() Key {
	return a.k
}

func (a *adaptIterator[_, Value]) Value() Value {
	return a.v
}

func (a *adaptIterator[Key, _]) AnyKey() interface{} {
	return a.Key()
}

func (a *adaptIterator[_, Value]) AnyValue() interface{} {
	return a.Value()
}

func (a *adaptIterator[_, _]) KeyExpression() Expression {
	return a.i.KeyExpression()
}

func (a *adaptIterator[_, _]) ValueExpression() Expression {
	return a.i.ValueExpression()
}

func FillMap[Key comparable, Value any](c Collection[Key, Value], m map[Key]Value) error {
	i := c.Begin()
	for {
		ok, err := i.Next()
		if !ok || err != nil {
			return err
		}
		m[i.Key()] = i.Value()
	}
}

func AdaptCollection[Key any, Value any](c UntypedCollection) Collection[Key, Value] {
	return Collection[Key, Value]{
		AnyCollection: adaptCollection[Key, Value]{
			UntypedCollection: c,
		},
	}
}

func CanAdaptCollection[Key any, Value any](c UntypedCollection) bool {
	i := c.BeginUntyped()
	if ok, err := i.Next(); !ok || err != nil {
		return true
	}
	if _, ok := i.Key().(Key); !ok {
		return false
	}
	if _, ok := i.Value().(Value); !ok {
		return false
	}
	return true
}

type ArrayCollection[Key any, Value any] struct {
	Keys   []Key
	Values []Value
}

type arrayIterator[Key any, Value any] struct {
	keys   []Key
	values []Value
	i      int
}

func (a ArrayCollection[Key, Value]) Begin() Iterator[Key, Value] {
	return &arrayIterator[Key, Value]{keys: a.Keys, values: a.Values}
}

func (a ArrayCollection[_, _]) Count() (int, bool) {
	return len(a.Keys), true
}

func (a arrayIterator[Key, _]) Key() Key {
	return a.keys[a.i-1]
}

func (a arrayIterator[_, Value]) Value() Value {
	return a.values[a.i-1]
}

func (a *arrayIterator[_, _]) Next() (bool, error) {
	a.i++
	return a.i <= len(a.keys), nil
}

func (a arrayIterator[_, _]) KeyExpression() Expression {
	if l, err := FromLiteral(a.Key()); err == nil {
		return Expression{AnyExpression: l.AnyLiteral}
	} else {
		panic(err.Error())
	}
}

func (a arrayIterator[_, _]) ValueExpression() Expression {
	if l, err := FromLiteral(a.Value()); err == nil {
		return Expression{AnyExpression: l.AnyLiteral}
	} else {
		panic(err.Error())
	}
}

func (a ArrayCollection[Key, Value]) Collection() Collection[Key, Value] {
	return Collection[Key, Value]{AnyCollection: a}
}

func (a ArrayCollection[Key, Value]) Clone() ArrayCollection[Key, Value] {
	clone := ArrayCollection[Key, Value]{
		Keys:   make([]Key, len(a.Keys)),
		Values: make([]Value, len(a.Values)),
	}
	copy(clone.Keys, a.Keys)
	copy(clone.Values, a.Values)
	return clone
}

var _ AnyCollection[int, string] = &ArrayCollection[int, string]{}

type ArrayValuesCollection[Value any] []Value

type arrayValuesIterator[Value any] struct {
	values []Value
	i      int
}

func (a ArrayValuesCollection[Value]) Begin() Iterator[int, Value] {
	return &arrayValuesIterator[Value]{values: a}
}

func (a ArrayValuesCollection[_]) Count() (int, bool) {
	return len(a), true
}

func (a arrayValuesIterator[_]) Key() int {
	return a.i - 1
}

func (a arrayValuesIterator[Value]) Value() Value {
	return a.values[a.i-1]
}

func (a *arrayValuesIterator[_]) Next() (bool, error) {
	a.i++
	return a.i <= len(a.values), nil
}

func (a arrayValuesIterator[_]) KeyExpression() Expression {
	return NewIntExpression(a.Key())
}

func (a arrayValuesIterator[_]) ValueExpression() Expression {
	if l, err := FromLiteral(a.Value()); err == nil {
		return Expression{AnyExpression: l.AnyLiteral}
	} else {
		panic(err.Error())
	}
}

func (a ArrayValuesCollection[Value]) Collection() Collection[int, Value] {
	return Collection[int, Value]{AnyCollection: a}
}

var _ AnyCollection[int, string] = &ArrayValuesCollection[string]{}

type ArrayFeatureCollection[Value Feature] []Value

type arrayFeatureIterator[Value Feature] struct {
	values []Value
	i      int
}

func (a ArrayFeatureCollection[Value]) Begin() Iterator[FeatureID, Value] {
	return &arrayFeatureIterator[Value]{values: []Value(a)}
}

func (a ArrayFeatureCollection[_]) Count() (int, bool) {
	return len(a), true
}

func (a arrayFeatureIterator[_]) AnyKey() interface{} {
	return a.Key()
}

func (a arrayFeatureIterator[_]) AnyValue() interface{} {
	return a.Value()
}

func (a arrayFeatureIterator[_]) Key() FeatureID {
	return a.values[a.i-1].FeatureID()
}

func (a arrayFeatureIterator[Value]) Value() Value {
	return a.values[a.i-1]
}

func (a *arrayFeatureIterator[_]) Next() (bool, error) {
	a.i++
	return a.i <= len(a.values), nil
}

func (a *arrayFeatureIterator[_]) KeyExpression() Expression {
	return NewFeatureIDExpression(a.Key())
}

func (a *arrayFeatureIterator[_]) ValueExpression() Expression {
	if l, err := FromLiteral(a.Value()); err == nil {
		return Expression{AnyExpression: l.AnyLiteral}
	} else {
		panic(err.Error())
	}
}

func (a ArrayFeatureCollection[Value]) Collection() Collection[FeatureID, Value] {
	return Collection[FeatureID, Value]{AnyCollection: a}
}

var _ AnyCollection[FeatureID, PhysicalFeature] = &ArrayFeatureCollection[PhysicalFeature]{}

func Count(collection UntypedCollection) (int, error) {
	if n, ok := collection.Count(); ok {
		return n, nil
	}
	n := 0
	i := collection.BeginUntyped()
	for {
		ok, err := i.Next()
		if !ok || err != nil {
			return n, err
		}
		n++
	}
}
