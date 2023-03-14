package api

import (
	"fmt"
	"reflect"
	"time"

	"diagonal.works/b6"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/search"
)

type Context struct {
	World  b6.World
	Cores  int
	Clock  func() time.Time
	Values map[interface{}]interface{}

	VM *VM
}

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
	if f, ok := id.(b6.Feature); ok {
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

type Number interface {
	isNumber()
}

type IntNumber int

func (_ IntNumber) isNumber() {}

type FloatNumber float64

func (_ FloatNumber) isNumber() {}

var featureInterface = reflect.TypeOf((*b6.Feature)(nil)).Elem()
var queryInterface = reflect.TypeOf((*search.Query)(nil)).Elem()
var numberInterface = reflect.TypeOf((*Number)(nil)).Elem()
var queryProtoPtrType = reflect.TypeOf((*pb.QueryProto)(nil))
var featureIDType = reflect.TypeOf(b6.FeatureID{})
var pointIDType = reflect.TypeOf(b6.PointID{})
var pathIDType = reflect.TypeOf(b6.PointID{})
var areaIDType = reflect.TypeOf(b6.AreaID{})
var relationIDType = reflect.TypeOf(b6.RelationID{})

func Convert(v reflect.Value, t reflect.Type, w b6.World) (reflect.Value, error) {
	if v.Type().AssignableTo(t) {
		return v, nil
	} else if v.CanConvert(t) {
		return v.Convert(t), nil
	} else if vv, ok := convertInterface(v, t); ok {
		return vv, nil
	}
	switch t {
	case queryInterface:
		// TODO: Harmonise query interfaces
		if vv, ok := v.Interface().(*pb.QueryProto); ok {
			if q, err := NewQueryFromProto(vv, w); err == nil {
				return reflect.ValueOf(q), nil
			} else {
				return reflect.Value{}, err
			}
		}
	case queryProtoPtrType:
		if vv, ok := v.Interface().(search.Query); ok {
			if q, err := QueryToProto(vv); err == nil {
				return reflect.ValueOf(q), nil
			} else {
				return reflect.Value{}, err
			}
		}
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
	case numberInterface:
		switch vv := v.Interface().(type) {
		case Number:
			return v, nil
		case float64:
			return reflect.ValueOf(FloatNumber(vv)), nil
		default:
			if i, ok := ToInt(vv); ok {
				return reflect.ValueOf(IntNumber(i)), nil
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
