package functions

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/test/testcamden"
)

func TestPercentiles(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	values := make([]float64, 1000)
	for i := range values {
		values[i] = r.Float64() * 5.0
	}
	keys := make([]interface{}, len(values))
	for i := range keys {
		keys[i] = fmt.Sprintf("%d", i)
	}

	collection, err := percentiles(&api.Context{}, &api.ArrayAnyFloatCollection{Keys: keys, Values: values})
	if err != nil {
		t.Error(err)
		return
	}

	i := 0
	j := collection.Begin()
	for {
		ok, err := j.Next()
		if err != nil {
			t.Error(err)
			return
		}
		if !ok {
			break
		}
		expected := values[i] / 5.0
		if math.Abs(j.Value().(float64)-expected) > 0.05 {
			t.Errorf("Expected a value close to %f, found %f", expected, j.Value().(float64))
			return
		}
		i++
	}
	if i != len(values) {
		t.Errorf("Expected %d values, found %d", len(values), i)
	}
}

func TestCount(t *testing.T) {
	granarySquare := testcamden.BuildGranarySquare(t)
	if granarySquare == nil {
		return
	}

	context := &api.Context{
		World: granarySquare,
	}
	collection, err := find(context, b6.Keyed{"#building"})
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	count, err := count(context, collection)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}
	expected := testcamden.BuildingsInGranarySquare
	if count != expected {
		t.Errorf("Expected count to return %d, found %d", expected, count)
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		a api.Number
		b api.Number
		r api.Number
	}{
		{api.IntNumber(2), api.IntNumber(3), api.IntNumber(5)},
		{api.IntNumber(2), api.FloatNumber(3.0), api.FloatNumber(5.0)},
		{api.FloatNumber(2.0), api.IntNumber(3.0), api.FloatNumber(5.0)},
		{api.FloatNumber(2.0), api.FloatNumber(3.0), api.FloatNumber(5.0)},
	}
	for _, test := range tests {
		if r, err := add(&api.Context{}, test.a, test.b); err != nil || !reflect.DeepEqual(r, test.r) {
			t.Errorf("Expected %T(%v) + %T(%v) = %T(%v), found %T(%v)", test.a, test.a, test.b, test.b, test.r, test.r, r, r)
		}
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		a api.Number
		b api.Number
		r api.Number
	}{
		{api.IntNumber(6), api.IntNumber(2), api.IntNumber(3)},
		{api.IntNumber(6), api.FloatNumber(2.0), api.FloatNumber(3.0)},
		{api.FloatNumber(6.0), api.IntNumber(2.0), api.FloatNumber(3.0)},
		{api.FloatNumber(6.0), api.FloatNumber(2.0), api.FloatNumber(3.0)},
	}
	for _, test := range tests {
		if r, err := divide(&api.Context{}, test.a, test.b); err != nil || !reflect.DeepEqual(r, test.r) {
			t.Errorf("Expected %T(%v) / %T(%v) = %T(%v), found %T(%v)", test.a, test.a, test.b, test.b, test.r, test.r, r, r)
		}
	}
}
