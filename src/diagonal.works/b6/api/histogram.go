package api

import (
	"fmt"
	"sort"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
)

func NewHistogramFromCollection(c b6.UntypedCollection, id b6.CollectionID) (*ingest.CollectionFeature, error) {
	kvs, err := countValues(c)
	if err != nil {
		return nil, err
	}

	buckets, err := bucketValues(kvs)
	if err != nil {
		return nil, err
	}

	tags := []b6.Tag{{Key: "b6", Value: b6.String("histogram")}}
	for i, bucket := range buckets {
		tags = append(tags, b6.Tag{Key: fmt.Sprintf("bucket:%d", i), Value: b6.String(bucket.label)})
	}

	histogram := ingest.CollectionFeature{
		Tags:         tags,
		CollectionID: id,
	}

	i := c.BeginUntyped()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		if index, err := buckets.bucket(i.Value()); err == nil && index >= 0 {
			histogram.Keys = append(histogram.Keys, i.Key())
			histogram.Values = append(histogram.Values, index)
		}
	}
	histogram.Sort()
	return &histogram, nil
}

func HistogramBucketLabels(c b6.CollectionFeature) []string {
	i := 0
	labels := make([]string, 0)
	for {
		if label := c.Get(fmt.Sprintf("bucket:%d", i)); label.IsValid() {
			labels = append(labels, label.Value.String())
		} else {
			break
		}
		i++
	}
	return labels
}

// TODO: Turn this into a HistogramFeature struct?

func HistogramBucketCounts(c b6.CollectionFeature) []int {
	counts := make([]int, 0)
	i := b6.AdaptCollection[any, int](c).Begin()
	for {
		ok, err := i.Next()
		if err != nil || !ok {
			break
		}
		for len(counts) <= i.Value() {
			counts = append(counts, 0)
		}
		counts[i.Value()]++
	}
	return counts
}

type bound struct {
	lower, upper interface{}
	label        string
	within       func(interface{}) (bool, error)
}
type buckets []bound

func xBound(lower, upper interface{}) bound {
	var label string
	if upper != nil {
		label = fmt.Sprint(lower) + "-" + fmt.Sprint(upper)
	} else {
		label = fmt.Sprint(lower) + "-"
	}
	return bound{
		lower: lower,
		upper: upper,
		label: label,
		within: func(v interface{}) (bool, error) {
			below, err := b6.Less(v, lower)
			if err != nil {
				return false, err
			}

			notAbove := true
			if upper != nil {
				notAbove, err = b6.Less(v, upper)
				if err != nil {
					return false, err
				}
			}
			return !below && notAbove, nil
		},
	}
}

func (b buckets) bucket(v interface{}) (int, error) {
	for i, bound := range b {
		yes, err := bound.within(v)
		if err != nil {
			return -1, err
		}
		if yes {
			return i, nil
		}
	}

	return -1, nil // Allowing data not to be bucketed.
}

const maxBuckets = 6

func categorical(kvs []*kv) (buckets, error) {
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].value > kvs[j].value
	})

	var b buckets
	for i, kv := range kvs {
		if i == maxBuckets {
			break
		}

		key := kv.key
		if i == maxBuckets-1 && len(kvs) > maxBuckets {
			key = "other"
		}

		b = append(b, bound{lower: key, label: fmt.Sprint(key), within: func(v interface{}) (bool, error) { return v == key || key == "other", nil }})
	}

	return b, nil
}

func uniform(kvs []*kv) (buckets, error) {
	sort.Slice(kvs, func(i, j int) bool {
		less, err := b6.Less(kvs[i].key, kvs[j].key)
		if err != nil {
			panic(err) // Not graceful, but greater should handle all numericals,
		} // and we do the numerical check in histogram call.

		return less
	})

	var b buckets
	if (len(kvs)) <= maxBuckets {
		for _, kv := range kvs {
			key := kv.key
			b = append(b, bound{lower: kv.key, label: fmt.Sprint(kv.key), within: func(v interface{}) (bool, error) { return v == key, nil }})
		}
	} else {
		for len(kvs) > 0 {
			bucket_size := len(kvs) / (maxBuckets - len(b))
			if len(kvs) > bucket_size {
				b = append(b, xBound(kvs[0].key, kvs[bucket_size].key))
				kvs = kvs[bucket_size:]

			} else {
				b = append(b, xBound(kvs[0].key, nil))
				kvs = kvs[0:0]
			}
		}
	}

	return b, nil
}

type kv struct {
	key      interface{}
	value    int            // key-count.
	features []b6.FeatureID // features in this bucket, if the values are Identifiable
}

func countValues(c b6.UntypedCollection) ([]*kv, error) {
	m := make(map[interface{}]*kv)
	kvs := make([]*kv, 0)

	i := c.BeginUntyped()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}

		var e *kv
		if e, ok = m[i.Value()]; ok {
			e.value++
		} else {
			e = &kv{key: i.Value(), value: 1}
			m[i.Value()] = e
			kvs = append(kvs, e)
		}
		if id, ok := i.Key().(b6.Identifiable); ok {
			e.features = append(e.features, id.FeatureID())
		}
	}

	return kvs, nil
}

func numerical(any interface{}) bool {
	switch any.(type) {
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64:
		return true
	default: // Ignoring complex numbers.
		return false
	}
}

func bucketValues(kvs []*kv) (buckets, error) {
	if len(kvs) == 0 {
		return buckets{}, nil
	}

	var b buckets
	var err error
	if numerical(kvs[0].key) { // Checking only first, assuming all elements in collection have the same type. TODO: enforce/check that somewhere else.
		b, err = uniform(kvs)
		if err != nil {
			return nil, err
		}
	} else {
		b, err = categorical(kvs)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}
