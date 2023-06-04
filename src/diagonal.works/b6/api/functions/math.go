package functions

import (
	"math"
	"sort"

	"diagonal.works/b6/api"
)

func divide(a api.Number, b api.Number, context *api.Context) (api.Number, error) {
	switch b := b.(type) {
	case api.IntNumber:
		switch a := a.(type) {
		case api.IntNumber:
			return api.IntNumber(int(a) / int(b)), nil
		case api.FloatNumber:
			return api.FloatNumber(float64(a) / float64(b)), nil
		}
	case api.FloatNumber:
		switch a := a.(type) {
		case api.IntNumber:
			return api.FloatNumber(float64(a) / float64(b)), nil
		case api.FloatNumber:
			return api.FloatNumber(float64(a) / float64(b)), nil
		}
	}
	panic("bad number")
}

func divideInt(a int, b float64, context *api.Context) (float64, error) {
	return float64(a) / b, nil
}

func addInts(a int, b int, context *api.Context) (int, error) {
	return a + b, nil
}

func clamp(v int, low int, high int, context *api.Context) (int, error) {
	if v < low {
		return low, nil
	} else if v > high {
		return high, nil
	}
	return v, nil
}

func gt(a interface{}, b interface{}, context *api.Context) (bool, error) {
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
func percentiles(collection api.AnyFloatCollection, context *api.Context) (api.AnyFloatCollection, error) {
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

func count(collection api.Collection, context *api.Context) (int, error) {
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
