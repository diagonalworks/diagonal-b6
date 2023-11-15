package api

import (
	"fmt"
	"reflect"

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

func ResolvePoint(id b6.IdentifiablePoint, w b6.World) b6.PointFeature {
	if f, ok := id.(b6.PointFeature); ok {
		return f
	}
	return b6.FindPointByID(id.PointID(), w)
}

func IsTrue(v interface{}) bool {
	switch v := v.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case float64:
		return v != 0.0
	case string:
		return v != ""
	case b6.Feature:
		return v != nil
	case b6.Geometry:
		return v != nil
	}
	return false
}

func IsInt(k reflect.Kind) bool {
	return k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64
}

func ToInt(v interface{}) (int, bool) {
	switch v := v.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		return int(v), true
	}
	return 0, false
}

func ToFloat64(v interface{}) (float64, error) {
	switch v := v.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	}

	return 0.0, fmt.Errorf("can't cast %T to float64", v)
}

func Less(a interface{}, b interface{}) (bool, error) {
	if aa, ok := ToInt(a); ok {
		if bb, ok := ToInt(b); ok {
			return aa < bb, nil
		}
	} else if aa, err := ToFloat64(a); err == nil {
		if bb, err := ToFloat64(b); err == nil {
			return aa < bb, nil
		}
	} else if aa, ok := a.(string); ok {
		if bb, ok := b.(string); ok {
			return aa < bb, nil
		}
	} else if aa, ok := a.(b6.FeatureID); ok {
		if bb, ok := b.(b6.FeatureID); ok {
			return aa.Less(bb), nil
		}
	}
	return false, fmt.Errorf("can't compare %T with %T", a, b)
}

func Equal(a interface{}, b interface{}) (bool, error) {
	if aa, ok := ToInt(a); ok {
		if bb, ok := ToInt(b); ok {
			return aa == bb, nil
		}
	} else if aa, err := ToFloat64(a); err == nil {
		if bb, err := ToFloat64(b); err == nil {
			return aa == bb, nil
		}
	} else if aa, ok := a.(string); ok {
		if bb, ok := b.(string); ok {
			return aa == bb, nil
		}
	} else if aa, ok := a.(b6.FeatureID); ok {
		if bb, ok := b.(b6.FeatureID); ok {
			return aa == bb, nil
		}
	}
	return false, fmt.Errorf("can't compare %T with %T", a, b)
}

func Greater(a interface{}, b interface{}) (bool, error) {
	var err error
	var ok bool
	if ok, err = Less(a, b); err == nil && !ok {
		if ok, err = Equal(a, b); err == nil && !ok {
			return true, nil
		}
	}
	return false, err
}

var featureInterface = reflect.TypeOf((*b6.Feature)(nil)).Elem()
var queryInterface = reflect.TypeOf((*b6.Query)(nil)).Elem()
var numberInterface = reflect.TypeOf((*b6.Number)(nil)).Elem()
var queryProtoPtrType = reflect.TypeOf((*pb.QueryProto)(nil))
var featureIDType = reflect.TypeOf(b6.FeatureID{})
var pointIDType = reflect.TypeOf(b6.PointID{})
var pathIDType = reflect.TypeOf(b6.PointID{})
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
	case pointIDType:
		if vv, ok := v.Interface().(b6.Identifiable); ok {
			if id := vv.FeatureID(); id.Type == b6.FeatureTypePoint {
				return reflect.ValueOf(id.ToPointID()), nil
			}
		}
	case pathIDType:
		if vv, ok := v.Interface().(b6.Identifiable); ok {
			if id := vv.FeatureID(); id.Type == b6.FeatureTypePath {
				return reflect.ValueOf(id.ToPathID()), nil
			}
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
			if i, ok := ToInt(vv); ok {
				return reflect.ValueOf(b6.IntNumber(i)), nil
			}
		}
	case reflect.TypeOf(""):
		if tag, ok := v.Interface().(b6.Tag); ok {
			return reflect.ValueOf(tag.Value), nil
		}
	case reflect.TypeOf(int(1)):
		if tag, ok := v.Interface().(b6.Tag); ok {
			i, _ := tag.IntValue()
			return reflect.ValueOf(i), nil
		}
	case reflect.TypeOf(float64(1.0)):
		if tag, ok := v.Interface().(b6.Tag); ok {
			f, _ := tag.FloatValue()
			return reflect.ValueOf(f), nil
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
	} else if t.Implements(untypedCollectionInterface) {
		if v.Type().AssignableTo(t) {
			return v, nil
		} else if adaptor, ok := context.Adaptors.Collections[t]; ok {
			if vc, ok := v.Interface().(b6.UntypedCollection); ok {
				return adaptor(vc), nil
			} else {
				return reflect.Value{}, fmt.Errorf("expected a collection, found %s", v.Type())
			}
		} else {
			return reflect.Value{}, fmt.Errorf("No collection adaptor for %s", t)
		}
	}
	return Convert(v, t, context.World)
}

func convertQueryToCallable(v reflect.Value, t reflect.Type) (Callable, bool) {
	ok := t.Kind() == reflect.Func
	ok = ok && v.Type().Implements(queryInterface)
	ok = ok && t.NumIn() == 2
	ok = ok && t.Out(0).Kind() == reflect.Bool
	if ok {
		q := v.Interface().(b6.Query)
		f := func(c *Context, feature interface{}) (bool, error) {
			if feature, ok := feature.(b6.Feature); ok {
				return q.Matches(feature, c.World), nil
			}
			return false, fmt.Errorf("Expected b6.Feature, found %T", feature)
		}
		return goCall{f: reflect.ValueOf(f), name: fmt.Sprintf("matches %s", q)}, true
	}
	return nil, false
}

func CanUseAsArg(have reflect.Type, want reflect.Type) bool {
	return have.AssignableTo(want) || have.ConvertibleTo(want)
}
