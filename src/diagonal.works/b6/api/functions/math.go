package functions

import (
	"math"
	"sort"

	"diagonal.works/b6/api"
)

func divide(context *api.Context, a api.Number, b api.Number) (api.Number, error) {
	if a, ok := a.(api.IntNumber); ok {
		if b, ok := b.(api.IntNumber); ok {
			return api.IntNumber(int(a) / int(b)), nil
		}
		return api.FloatNumber(float64(a) / float64(b.(api.FloatNumber))), nil
	}
	if b, ok := b.(api.IntNumber); ok {
		return api.FloatNumber(float64(a.(api.FloatNumber)) / float64(b)), nil
	}
	return api.FloatNumber(float64(a.(api.FloatNumber)) / float64(b.(api.FloatNumber))), nil
}

func divideInt(context *api.Context, a int, b float64) (float64, error) {
	return float64(a) / b, nil
}

func add(context *api.Context, a api.Number, b api.Number) (api.Number, error) {
	if a, ok := a.(api.IntNumber); ok {
		if b, ok := b.(api.IntNumber); ok {
			return api.IntNumber(int(a) + int(b)), nil
		}
		return api.FloatNumber(float64(a) + float64(b.(api.FloatNumber))), nil
	}
	if b, ok := b.(api.IntNumber); ok {
		return api.FloatNumber(float64(a.(api.FloatNumber)) + float64(b)), nil
	}
	return api.FloatNumber(float64(a.(api.FloatNumber)) + float64(b.(api.FloatNumber))), nil
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
func percentiles(context *api.Context, collection api.AnyFloatCollection) (api.AnyFloatCollection, error) {
	keys := make([]interface{}, 0)
	values := make([]float64, 0)
	i := collection.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		keys = append(keys, i.Key())
		values = append(values, i.Value().(float64))
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
	return &api.ArrayAnyFloatCollection{Keys: keys, Values: values}, nil
}

func count(context *api.Context, collection api.Collection) (int, error) {
	if c, ok := collection.(api.Countable); ok {
		return c.Count(), nil
	}
	n := 0
	i := collection.Begin()
	for {
		ok, err := i.Next()
		if !ok || err != nil {
			return n, err
		}
		n++
	}
}
