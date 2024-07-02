package api

import (
	"fmt"
	"sort"
	"strconv"
	"unicode"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
)

const HistogramOriginDestinationKey = "b6:origin_destination"

func NewHistogramFromCollection(c b6.UntypedCollection, id b6.CollectionID) (*ingest.CollectionFeature, error) {
	var h *ingest.CollectionFeature
	var err error
	if isOriginDestinationCollection(c) {
		h, err = newOriginDestinationHistogram(c)
	} else {
		h, err = newBucketedHistogram(c)
	}
	h.CollectionID = id
	h.Tags = append(h.Tags, b6.Tag{Key: "b6", Value: b6.NewStringExpression("histogram")})
	return h, err
}

func newBucketedHistogram(c b6.UntypedCollection) (*ingest.CollectionFeature, error) {
	kvs, err := countValues(c)
	if err != nil {
		return nil, err
	}

	buckets, err := bucketValues(kvs)
	if err != nil {
		return nil, err
	}

	histogram := ingest.CollectionFeature{}
	for i, bucket := range buckets {
		histogram.Tags = append(histogram.Tags, b6.Tag{Key: fmt.Sprintf("bucket:%d", i), Value: b6.NewStringExpression(bucket.label)})
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

func newOriginDestinationHistogram(c b6.UntypedCollection) (*ingest.CollectionFeature, error) {
	histogram := ingest.CollectionFeature{
		Tags: []b6.Tag{{Key: HistogramOriginDestinationKey, Value: b6.NewStringExpression("yes")}},
	}
	i := c.BeginUntyped()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		k, kok := i.Key().(b6.FeatureID)
		v, vok := i.Value().(b6.FeatureID)
		if !kok || !vok {
			return nil, fmt.Errorf("expected all keys and values to be FeatureID")
		}
		histogram.Keys = append(histogram.Keys, k)
		histogram.Values = append(histogram.Values, v)
	}

	// TODO: expose whether the sorted flag on all collections, not just
	// collection features
	histogram.Sort()
	return &histogram, nil
}

func isOriginDestinationCollection(c b6.UntypedCollection) bool {
	// TODO: Add methods to collection to recover the key and
	// value types, so it'll work with empty collections, once
	// we've written code to identify the return type of functions
	// given their args (for, eg, map collections)
	i := c.BeginUntyped()
	ok, err := i.Next()
	if !ok || err != nil {
		return false
	}
	_, kok := i.Key().(b6.FeatureID)
	_, vok := i.Value().(b6.FeatureID)
	return kok && vok
}

func HistogramBucketLabels(c b6.CollectionFeature, n int) []string {
	labels := make([]string, n)
	for i := range labels {
		if label := c.Get(fmt.Sprintf("bucket:%d", i)); label.IsValid() {
			labels[i] = label.Value.String()
		} else {
			labels[i] = strconv.Itoa(i)
		}
	}
	return labels
}

func HistogramBucketCounts(c b6.CollectionFeature) ([]int, int, error) {
	if od := c.Get(HistogramOriginDestinationKey); od.IsValid() {
		return bucketCountsFromOriginDestinationHistogram(c)
	} else {
		return bucketCountsFromBucketedHistogram(c)
	}
}

func bucketCountsFromBucketedHistogram(c b6.CollectionFeature) ([]int, int, error) {
	counts := make([]int, 0)
	total := 0
	i := b6.AdaptCollection[any, int](c).Begin()
	var err error
	for {
		var ok bool
		ok, err = i.Next()
		if err != nil || !ok {
			break
		}
		for len(counts) <= i.Value() {
			counts = append(counts, 0)
		}
		counts[i.Value()]++
		total++
	}
	return counts, total, err
}

func bucketCountsFromOriginDestinationHistogram(c b6.CollectionFeature) ([]int, int, error) {
	i := b6.AdaptCollection[b6.FeatureID, b6.FeatureID](c).Begin()
	id := b6.FeatureIDInvalid
	count := 0
	total := 0
	buckets := make([]int, 0)
	for {
		ok, err := i.Next()
		if err != nil {
			return buckets, 0, err
		} else if !ok {
			break
		}
		if i.Key() != id {
			if id != b6.FeatureIDInvalid {
				for len(buckets) <= count {
					buckets = append(buckets, 0)
				}
				buckets[count]++
			}
			id = i.Key()
			total++
			count = 0
		}
		if i.Value() != b6.FeatureIDInvalid {
			count++
		}
	}
	if id != b6.FeatureIDInvalid {
		for len(buckets) <= count {
			buckets = append(buckets, 0)
		}
		buckets[count]++
	}
	return buckets, total, nil
}

type bound struct {
	lower, upper interface{}
	label        string
	within       func(interface{}) (bool, error)
}
type buckets []bound

func xBound(lower, upper interface{}) bound {
	return bound{
		lower: lower,
		upper: upper,
		label: formatLabel(lower, upper),
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

func formatLabel(lower, upper interface{}) string {
	if upper != nil {
		if l, lok := lower.(int); lok {
			if u, uok := upper.(int); uok {
				if u == l+1 {
					return strconv.Itoa(l)
				}
			}
		}
		return formatLabelValue(lower) + "-" + formatLabelValue(upper)
	}
	return formatLabelValue(lower) + "-"
}

func formatLabelValue(value interface{}) string {
	if f, ok := value.(float64); ok {
		return fmt.Sprintf("%.3g", f)
	}
	return fmt.Sprintf("%v", value)
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

const MaxHistogramBuckets = 6

func categorical(kvs []*kv) (buckets, error) {
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].value > kvs[j].value
	})

	var b buckets
	for i, kv := range kvs {
		if i == MaxHistogramBuckets {
			break
		}

		key := kv.key
		if i == MaxHistogramBuckets-1 && len(kvs) > MaxHistogramBuckets {
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
	if (len(kvs)) <= MaxHistogramBuckets {
		for _, kv := range kvs {
			key := kv.key
			b = append(b, bound{lower: kv.key, label: fmt.Sprint(kv.key), within: func(v interface{}) (bool, error) { return v == key, nil }})
		}
	} else {
		for len(kvs) > 0 {
			bucket_size := len(kvs) / (MaxHistogramBuckets - len(b))
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
	switch any := any.(type) {
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64:
		return true
	case string: // TODO(mari): remove when tag values are literals / numbers are not converted to strings at any point.
		for _, c := range any {
			if !unicode.IsDigit(c) {
				return false
			}
		}
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
