package functions

import (
	"math"
	"sort"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
)

// Return a divided by b.
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

// Deprecated.
func divideInt(context *api.Context, a int, b float64) (float64, error) {
	return float64(a) / b, nil
}

// Return a added to b.
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

// Deprecated.
func addInts(context *api.Context, a int, b int) (int, error) {
	return a + b, nil
}

// Return the given value, unless it falls outside the given inclusive bounds, in which case return the boundary.
func clamp(context *api.Context, v int, low int, high int) (int, error) {
	if v < low {
		return low, nil
	} else if v > high {
		return high, nil
	}
	return v, nil
}

// Return true if a is greater than b.
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

// Return a collection where values represent the perentile of the corresponding value in the given collection.
// The returned collection is ordered by percentile, with keys drawn from the
// given collection.
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

// Return the number of items in the given collection.
// The function will not evaluate and traverse the entire collection if it's possible to count
// the collection efficiently.
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
