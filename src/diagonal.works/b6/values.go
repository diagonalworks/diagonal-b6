package b6

// TODO: Move these functions to methods on the b6.Value
// interface

import (
	"fmt"
	"reflect"
)

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
	case Feature:
		return v != nil
	case Geometry:
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
	case IntNumber:
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
	} else if aa, ok := a.(FeatureID); ok {
		if bb, ok := b.(FeatureID); ok {
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
	} else if aa, ok := a.(FeatureID); ok {
		if bb, ok := b.(FeatureID); ok {
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
