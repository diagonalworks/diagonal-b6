package functions

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"diagonal.works/b6/api"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/test/camden"
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

	collection, err := percentiles(&api.ArrayAnyFloatCollection{Keys: keys, Values: values}, nil)
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
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	q := &pb.QueryProto{
		Query: &pb.QueryProto_Key{
			Key: &pb.KeyQueryProto{
				Key: "#building",
			},
		},
	}
	qq, _ := api.NewQueryFromProto(q, granarySquare)

	context := &api.Context{
		World: granarySquare,
	}

	collection, err := Find(qq, context)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}

	count, err := count(collection, context)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
		return
	}
	expected := camden.BuildingsInGranarySquare
	if count != expected {
		t.Errorf("Expected count to return %d, found %d", expected, count)
	}
}
