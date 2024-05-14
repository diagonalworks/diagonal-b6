package api

import (
	"fmt"
	"reflect"
	"strconv"

	"diagonal.works/b6"
	pb "diagonal.works/b6/proto"
)

type Pair interface {
	First() interface{}
	Second() interface{}
}

type AnyAnyPair [2]interface{}

func (a AnyAnyPair) First() interface{}  { return a[0] }
func (a AnyAnyPair) Second() interface{} { return a[1] }

var _ Pair = AnyAnyPair{}

type StringStringPair [2]string

func (s StringStringPair) First() interface{}  { return s[0] }
func (s StringStringPair) Second() interface{} { return s[1] }

func (s *StringStringPair) FromPair(p Pair) error {
	if first, ok := p.First().(string); ok {
		if second, ok := p.Second().(string); ok {
			(*s)[0] = first
			(*s)[1] = second
			return nil
		}
	}
	return fmt.Errorf("Expected a pair of strings, found {%T,%T}", p.First(), p.Second())
}

var _ Pair = StringStringPair{}

func Resolve(id b6.Identifiable, w b6.World) b6.Feature {
	if id == nil {
		return nil
	} else if f, ok := id.(b6.Feature); ok {
		return f
	}
	return w.FindFeatureByID(id.FeatureID())
}

var featureInterface = reflect.TypeOf((*b6.Feature)(nil)).Elem()
var queryInterface = reflect.TypeOf((*b6.Query)(nil)).Elem()
var numberInterface = reflect.TypeOf((*b6.Number)(nil)).Elem()
var expressionInterface = reflect.TypeOf((*b6.Expression)(nil)).Elem()
var queryProtoPtrType = reflect.TypeOf((*pb.QueryProto)(nil))
var featureIDType = reflect.TypeOf(b6.FeatureID{})

var areaIDType = reflect.TypeOf(b6.AreaID{})
var relationIDType = reflect.TypeOf(b6.RelationID{})
var collectionIDType = reflect.TypeOf(b6.CollectionID{})
var expressionIDType = reflect.TypeOf(b6.ExpressionID{})
var callableInterface = reflect.TypeOf((*Callable)(nil)).Elem()
var untypedCollectionInterface = reflect.TypeOf((*b6.UntypedCollection)(nil)).Elem()

// Convert v to type t, if possible. Doesn't convert functions.
func Convert(v reflect.Value, t reflect.Type, w b6.World) (reflect.Value, error) {
	if v.Type().AssignableTo(t) {
		return v, nil
	} else if v.CanConvert(t) {
		return v.Convert(t), nil
	} else if vv, ok := convertInterface(v, t); ok {
		return vv, nil
	}
	switch t {
	case featureIDType:
		if vv, ok := v.Interface().(b6.Identifiable); ok {
			return reflect.ValueOf(vv.FeatureID()), nil
		}
	case areaIDType:
		if vv, ok := v.Interface().(b6.Identifiable); ok {
			if id := vv.FeatureID(); id.Type == b6.FeatureTypeArea {
				return reflect.ValueOf(id.ToAreaID()), nil
			}
		}
	case relationIDType:
		if vv, ok := v.Interface().(b6.Identifiable); ok {
			if id := vv.FeatureID(); id.Type == b6.FeatureTypeRelation {
				return reflect.ValueOf(id.ToRelationID()), nil
			}
		}
	case collectionIDType:
		if vv, ok := v.Interface().(b6.Identifiable); ok {
			if id := vv.FeatureID(); id.Type == b6.FeatureTypeCollection {
				return reflect.ValueOf(id.ToCollectionID()), nil
			}
		}
	case expressionIDType:
		if vv, ok := v.Interface().(b6.Identifiable); ok {
			if id := vv.FeatureID(); id.Type == b6.FeatureTypeExpression {
				return reflect.ValueOf(id.ToExpressionID()), nil
			}
		}
	case numberInterface:
		switch vv := v.Interface().(type) {
		case b6.Number:
			return v, nil
		case float64:
			return reflect.ValueOf(b6.FloatNumber(vv)), nil
		default:
			if i, ok := b6.ToInt(vv); ok {
				return reflect.ValueOf(b6.IntNumber(i)), nil
			}
		}
	case expressionInterface:
		switch vv := v.Interface().(type) {
		case Callable:
			return reflect.ValueOf(vv.Expression()), nil
		}
	case reflect.TypeOf(""):
		if tag, ok := v.Interface().(b6.Tag); ok {
			return reflect.ValueOf(tag.Value.String()), nil
		}
	case reflect.TypeOf(int(1)):
		switch v := v.Interface().(type) {
		case b6.Tag:
			i, _ := strconv.Atoi(v.Value.String())
			return reflect.ValueOf(i), nil
		case b6.IntNumber:
			return reflect.ValueOf(int(v)), nil
		case b6.FloatNumber:
			return reflect.ValueOf(int(v)), nil
		}
	case reflect.TypeOf(float64(1.0)):
		switch v := v.Interface().(type) {
		case b6.Tag:
			f, _ := strconv.ParseFloat(v.Value.String(), 64)
			return reflect.ValueOf(f), nil
		case b6.IntNumber:
			return reflect.ValueOf(float64(v)), nil
		case b6.FloatNumber:
			return reflect.ValueOf(float64(v)), nil
		}
	}
	return reflect.Value{}, fmt.Errorf("expected %s, found %s", t, v.Type())
}

func convertInterface(v reflect.Value, t reflect.Type) (reflect.Value, bool) {
	if t.Kind() != reflect.Interface {
		return v, false
	}
	if v.CanInterface() {
		i := v.Interface()
		if tt := reflect.TypeOf(i); tt.Implements(t) {
			return reflect.ValueOf(i).Convert(t), true
		}
	}
	return v, false
}

// Convert v to type t, if possible. If v represents a b6 function, it'll be
// turned into a go function that executes it in a vm.
func ConvertWithContext(v reflect.Value, t reflect.Type, context *Context) (reflect.Value, error) {
	if t.Kind() == reflect.Func {
		var c Callable
		if vc, ok := v.Interface().(Callable); ok {
			c = vc
		} else if matches, ok := convertQueryToCallable(v, t); ok {
			c = matches
		}
		if c != nil {
			if c.NumArgs()+1 == t.NumIn() {
				return c.ToFunctionValue(t, context), nil
			} else {
				return reflect.Value{}, fmt.Errorf("expected a function with %d args, found %d", t.NumIn()-1, c.NumArgs())
			}
		}
	} else if t.Implements(callableInterface) {
		if callable, ok := v.Interface().(Callable); ok {
			return reflect.ValueOf(callable), nil
		} else if matches, ok := convertQueryToCallable(v, t); ok {
			return reflect.ValueOf(matches), nil
		} else {
			return reflect.Value{}, fmt.Errorf("expected a function, found %s", v.Type())
		}
	} else if t.Implements(untypedCollectionInterface) {
		if v.Type().AssignableTo(t) {
			return v, nil
		} else if adaptor, ok := context.Adaptors.Collections[t]; ok {
			if vc, ok := v.Interface().(b6.UntypedCollection); ok {
				return adaptor(vc), nil
			} else {
				return reflect.Value{}, fmt.Errorf("expected a collection, found %s", v.Type())
			}
		} else if c, ok := v.Interface().(b6.CollectionFeature); ok && t == untypedCollectionInterface {
			return reflect.ValueOf(c.(b6.UntypedCollection)), nil
		} else {
			return reflect.Value{}, fmt.Errorf("no collection adaptor for %s", t)
		}
	}
	return Convert(v, t, context.World)
}

func convertQueryToCallable(v reflect.Value, t reflect.Type) (Callable, bool) {
	if !v.Type().Implements(queryInterface) {
		return nil, false
	}
	if t.Kind() == reflect.Func {
		if t.NumIn() != 2 || t.Out(0).Kind() != reflect.Bool {
			return nil, false
		}
	} else if !t.Implements(callableInterface) {
		return nil, false
	}
	q := v.Interface().(b6.Query)
	f := func(c *Context, feature interface{}) (bool, error) {
		if feature, ok := feature.(b6.Feature); ok {
			return q.Matches(feature, c.World), nil
		}
		return false, fmt.Errorf("Expected b6.Feature, found %T", feature)
	}
	e := b6.NewCallExpression(
		b6.NewSymbolExpression("matches"),
		[]b6.Expression{b6.NewQueryExpression(q)},
	)
	return NewNativeFunction1(f, e), true
}

func CanUseAsArg(have reflect.Type, want reflect.Type) bool {
	return have.AssignableTo(want) || have.ConvertibleTo(want)
}
