package functions

import (
	"math"
	"sort"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
)

func divide(context *api.Context, a b6.Number, b b6.Number) (b6.Number, error) {
	if a, ok := a.(b6.IntNumber); ok {
		if b, ok := b.(b6.IntNumber); ok {
			return b6.IntNumber(int(a) / int(b)), nil
		}
		return b6.FloatNumber(float64(a) / float64(b.(b6.FloatNumber))), nil
	}
	if b, ok := b.(b6.IntNumber); ok {
		return b6.FloatNumber(float64(a.(b6.FloatNumber)) / float64(b)), nil
	}
	return b6.FloatNumber(float64(a.(b6.FloatNumber)) / float64(b.(b6.FloatNumber))), nil
}

func divideInt(context *api.Context, a int, b float64) (float64, error) {
	return float64(a) / b, nil
}

func add(context *api.Context, a b6.Number, b b6.Number) (b6.Number, error) {
	if a, ok := a.(b6.IntNumber); ok {
		if b, ok := b.(b6.IntNumber); ok {
			return b6.IntNumber(int(a) + int(b)), nil
		}
		return b6.FloatNumber(float64(a) + float64(b.(b6.FloatNumber))), nil
	}
	if b, ok := b.(b6.IntNumber); ok {
		return b6.FloatNumber(float64(a.(b6.FloatNumber)) + float64(b)), nil
	}
	return b6.FloatNumber(float64(a.(b6.FloatNumber)) + float64(b.(b6.FloatNumber))), nil
}

func addInts(context *api.Context, a int, b int) (int, error) {
	return a + b, nil
}

func clamp(context *api.Context, v int, low int, high int) (int, error) {
	if v < low {
		return low, nil
	} else if v > high {
		return high, nil
	}
	return v, nil
}

func gt(context *api.Context, a interface{}, b interface{}) (bool, error) {
	return api.Greater(a, b)
}

type byIndex struct {
	values  []float64
	indices []uint16
}

func (b byIndex) Len() int { return len(b.values) }
func (b byIndex) Swap(i, j int) {
	b.indices[i], b.indices[j] = b.indices[j], b.indices[i]
}
func (b byIndex) Less(i, j int) bool {
	return b.values[b.indices[i]] < b.values[b.indices[j]]
}

// TODO: percentiles inefficiently calculates the exact percentile by sorting the entire
// collection. We could use a histogram sketch instead, maybe constructed in the
// background with Collection
func percentiles(context *api.Context, collection b6.Collection[interface{}, float64]) (b6.Collection[interface{}, float64], error) {
	keys := make([]interface{}, 0)
	values := make([]float64, 0)
	i := collection.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return b6.Collection[interface{}, float64]{}, err
		}
		if !ok {
			break
		}
		keys = append(keys, i.Key())
		values = append(values, i.Value())
	}
	indices := make([]uint16, len(values))
	for i := range indices {
		indices[i] = uint16(i)
	}
	sort.Sort(byIndex{values: values, indices: indices})
	firstIndex := -1
	firstValue := math.NaN()
	for i := range indices {
		if values[indices[i]] != firstValue {
			firstValue = values[indices[i]]
			firstIndex = i
		}
		values[indices[i]] = float64(firstIndex) / float64(len(indices))
	}
	return b6.ArrayCollection[interface{}, float64]{Keys: keys, Values: values}.Collection(), nil
}

func count(context *api.Context, collection b6.Collection[any, any]) (int, error) {
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
