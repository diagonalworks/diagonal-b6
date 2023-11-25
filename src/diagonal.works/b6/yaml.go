package b6

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

func marshalChoiceYAML(choices interface{}, value interface{}, extra interface{}) (interface{}, error) {
	ct := reflect.TypeOf(choices).Elem()
	vt := reflect.TypeOf(value)
	for i := 0; i < ct.NumField(); i++ {
		f := ct.Field(i)
		if f.Type == vt || f.Type.Elem() == vt {
			y := map[string]interface{}{strings.ToLower(f.Name): value}
			var err error
			if extra != nil {
				// The round trip to bytes here is unfortunate, but it's
				// the only option with the API available
				var b []byte
				if b, err = yaml.Marshal(extra); err == nil {
					yy := map[string]interface{}{}
					if err = yaml.Unmarshal(b, &yy); err == nil {
						for k, v := range yy {
							y[k] = v
						}
					}
				}
			}
			return y, err
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
