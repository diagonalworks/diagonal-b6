package ui

import (
	"fmt"
	"sort"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	pb "diagonal.works/b6/proto"
)

func fillResponseFromHistogram(response *pb.UIResponseProto, c *api.HistogramCollection, w b6.World) error {
	kvs, err := countValues(c)
	if err != nil {
		return err
	}

	buckets, err := bucketValues(kvs)
	if err != nil {
		return err
	}

	values := make([]int, len(buckets))
	for _, kv := range kvs {
		if index, err := buckets.bucket(kv.key); err == nil {
			if index >= 0 {
				values[index] += kv.value
				if kv.features != nil {
					for len(response.Bucketed) <= index {
						response.Bucketed = append(response.Bucketed, &pb.FeatureIDsProto{})
					}
					for _, id := range kv.features {
						addToFeatureIDs(response.Bucketed[index], id)
					}
				}
			}
		} else {
			return err
		}
	}

	substack := &pb.SubstackProto{}
	total := 0
	begin := len(substack.Lines)
	for i, v := range values {
		substack.Lines = append(substack.Lines, &pb.LineProto{
			Line: &pb.LineProto_HistogramBar{
				HistogramBar: &pb.HistogramBarLineProto{
					Range: AtomFromValue(buckets[i].label, w),
					Value: int32(v),
					Index: int32(i),
				},
			},
		})
		total += v
	}
	for i := begin; i < len(substack.Lines); i++ {
		substack.Lines[i].GetHistogramBar().Total = int32(total)
	}
	if response.Stack == nil {
		response.Stack = &pb.StackProto{}
	}
	response.Stack.Substacks = append(response.Stack.Substacks, substack)
	return nil
}

func addToFeatureIDs(p *pb.FeatureIDsProto, id b6.FeatureID) {
	n := fmt.Sprintf("/%s/%s", id.Type.String(), id.Namespace.String())
	ids := -1
	for i, nn := range p.Namespaces {
		if n == nn {
			ids = i
			break
		}
	}
	if ids < 0 {
		ids = len(p.Ids)
		p.Namespaces = append(p.Namespaces, n)
		p.Ids = append(p.Ids, &pb.IDsProto{})
	}
	p.Ids[ids].Ids = append(p.Ids[ids].Ids, id.Value)
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
			below, err := api.Less(v, lower)
			if err != nil {
				return false, err
			}

			notAbove := true
			if upper != nil {
				notAbove, err = api.Less(v, upper)
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

const maxBuckets = 5

func categorical(kvs []kv) (buckets, error) {
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

func uniform(kvs []kv) (buckets, error) {
	sort.Slice(kvs, func(i, j int) bool {
		less, err := api.Less(kvs[i].key, kvs[j].key)
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

func countValues(c api.Collection) ([]kv, error) {
	m := make(map[interface{}]*kv)

	i := c.Begin()
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
		}
		if id, ok := i.Key().(b6.Identifiable); ok {
			e.features = append(e.features, id.FeatureID())
		}
	}

	var kvs []kv
	for _, kv := range m {
		kvs = append(kvs, *kv)
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

func bucketValues(kvs []kv) (buckets, error) {
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
