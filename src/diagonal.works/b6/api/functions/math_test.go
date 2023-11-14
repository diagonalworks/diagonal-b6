package functions

import (
	"math"
	"math/rand"
	"reflect"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/test/camden"
)

func TestPercentiles(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	input := b6.ArrayValuesCollection[float64](make([]float64, 1000))
	for i := range input {
		input[i] = r.Float64() * 5.0
	}

	collection, err := percentiles(&api.Context{}, input.Collection().Values())
	if err != nil {
		t.Fatal(err)
	}

	i := 0
	j := collection.Begin()
	for {
		ok, err := j.Next()
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			break
		}
		expected := input[i] / 5.0
		if math.Abs(j.Value()-expected) > 0.05 {
			t.Fatalf("Expected a value close to %f, found %f", expected, j.Value())
		}
		i++
	}
	if i != len(input) {
		t.Errorf("Expected %d values, found %d", len(input), i)
	}
}

func TestCount(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	context := &api.Context{
		World: granarySquare,
	}
	collection, err := find(context, b6.Keyed{Key: "#building"})
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	c := b6.AdaptCollection[any, any](collection)
	count, err := count(context, c)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	expected := camden.BuildingsInGranarySquare
	if count != expected {
		t.Errorf("Expected count to return %d, found %d", expected, count)
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		a b6.Number
		b b6.Number
		r b6.Number
	}{
		{b6.IntNumber(2), b6.IntNumber(3), b6.IntNumber(5)},
		{b6.IntNumber(2), b6.FloatNumber(3.0), b6.FloatNumber(5.0)},
		{b6.FloatNumber(2.0), b6.IntNumber(3.0), b6.FloatNumber(5.0)},
		{b6.FloatNumber(2.0), b6.FloatNumber(3.0), b6.FloatNumber(5.0)},
	}
	for _, test := range tests {
		if r, err := add(&api.Context{}, test.a, test.b); err != nil || !reflect.DeepEqual(r, test.r) {
			t.Errorf("Expected %T(%v) + %T(%v) = %T(%v), found %T(%v)", test.a, test.a, test.b, test.b, test.r, test.r, r, r)
		}
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		a b6.Number
		b b6.Number
		r b6.Number
	}{
		{b6.IntNumber(6), b6.IntNumber(2), b6.IntNumber(3)},
		{b6.IntNumber(6), b6.FloatNumber(2.0), b6.FloatNumber(3.0)},
		{b6.FloatNumber(6.0), b6.IntNumber(2.0), b6.FloatNumber(3.0)},
		{b6.FloatNumber(6.0), b6.FloatNumber(2.0), b6.FloatNumber(3.0)},
	}
	for _, test := range tests {
		if r, err := divide(&api.Context{}, test.a, test.b); err != nil || !reflect.DeepEqual(r, test.r) {
			t.Errorf("Expected %T(%v) / %T(%v) = %T(%v), found %T(%v)", test.a, test.a, test.b, test.b, test.r, test.r, r, r)
		}
	}
}
