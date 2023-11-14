package b6

import (
	"fmt"
	"reflect"
	"strings"
)

func marshalChoiceYAML(choices interface{}, value interface{}) (interface{}, error) {
	ct := reflect.TypeOf(choices).Elem()
	vt := reflect.TypeOf(value)
	for i := 0; i < ct.NumField(); i++ {
		f := ct.Field(i)
		if f.Type == vt || f.Type.Elem() == vt {
			return map[string]interface{}{strings.ToLower(f.Name): value}, nil
		}
	}
	return nil, fmt.Errorf("can't marshal %T for %T", value, choices)
}

func unmarshalChoiceYAML(choices interface{}, unmarshal func(interface{}) error) (interface{}, error) {
	if err := unmarshal(choices); err != nil {
		return nil, err
	}
	t := reflect.TypeOf(choices).Elem()
	v := reflect.ValueOf(choices).Elem()
	for i := 0; i < t.NumField(); i++ {
		if f := v.Field(i); !f.IsNil() {
			return f.Interface(), nil
		}
	}
	var broken interface{}
	unmarshal(&broken)
	return nil, fmt.Errorf("can't unmarshal %+v for %T", broken, choices)
}
