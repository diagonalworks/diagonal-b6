package b6 

import (
	"container/heap"
)

type featuresHeap []Features

func (f *featuresHeap) Len() int { return len(*f) }

func (f *featuresHeap) Less(i, j int) bool {
	return (*f)[i].FeatureID().Less((*f)[j].FeatureID())
}

func (f *featuresHeap) Swap(i, j int) {
	(*f)[i], (*f)[j] = (*f)[j], (*f)[i]
}

func (f *featuresHeap) Push(x interface{}) {
	*f = append(*f, x.(Features))
}

func (f *featuresHeap) Pop() interface{} {
	old := *f
	n := len(old)
	x := old[n-1]
	*f = old[0 : n-1]
	return x
}

type mergedFeatures struct {
	features featuresHeap
	started  bool
}

func (m *mergedFeatures) Next() bool {
	if !m.started {
		read, write := 0, 0
		for read < len(m.features) {
			if ok := m.features[read].Next(); ok {
				if write != read {
					m.features[write] = m.features[read]
				}
				write++
				read++
			} else {
				read++
			}
		}
		m.features = m.features[0:write]
		heap.Init(&m.features)
		m.started = true
		return len(m.features) > 0
	} else if len(m.features) > 0 {
		current := m.features[0].FeatureID()
		for {
			if m.features[0].Next() {
				heap.Fix(&m.features, 0)
			} else {
				heap.Pop(&m.features)
			}
			if len(m.features) == 0 || m.features[0].FeatureID() != current {
				break
			}
		}
		return len(m.features) > 0
	}
	return false
}

func (m *mergedFeatures) Feature() Feature {
	return m.features[0].Feature()
}

func (m *mergedFeatures) FeatureID() FeatureID {
	return m.features[0].FeatureID()
}

func MergeFeatures(features ...Features) Features {
	return &mergedFeatures{features: features, started: false}
}
