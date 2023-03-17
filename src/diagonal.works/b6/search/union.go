package search

import (
	"container/heap"
)

type Union []Query

func (u Union) String() string {
	joined := "(union"
	for _, q := range u {
		joined = joined + " " + q.String()
	}
	return joined + ")"
}

func (u Union) Compile(index Index) Iterator {
	iterators := make([]Iterator, len(u))
	for i, query := range u {
		iterators[i] = query.Compile(index)
	}
	return NewUnion(iterators, index.Values())
}

type iteratorLength struct {
	iterator Iterator
	length   int
}

type union struct {
	iterators []Iterator
	started   bool
	values    Values
}

func (u *union) Next() bool {
	if u.started {
		if len(u.iterators) > 0 {
			current := u.values.Key(u.iterators[0].Value())
			h := unionHeap{iterators: u.iterators, values: u.values}
			for len(h.iterators) > 0 {
				if u.values.CompareKey(h.iterators[0].Value(), current) != ComparisonEqual {
					break
				}
				if h.iterators[0].Next() {
					heap.Fix(&h, 0)
				} else {
					heap.Pop(&h)
				}
			}
			u.iterators = h.iterators
		}
	} else {
		u.start()
	}
	return len(u.iterators) > 0
}

func (u *union) Advance(to Key) bool {
	if !u.started {
		u.start()
	}
	h := unionHeap{iterators: u.iterators, values: u.values}
	for h.Len() > 0 {
		if u.values.CompareKey(h.iterators[0].Value(), to) == ComparisonLess {
			if u.iterators[0].Advance(to) {
				heap.Fix(&h, 0)
			} else {
				heap.Pop(&h)
			}
		} else {
			break
		}
	}
	u.iterators = h.iterators
	return len(u.iterators) > 0
}

func (u *union) start() {
	h := unionHeap{iterators: make([]Iterator, 0, len(u.iterators)), values: u.values}
	for _, iterator := range u.iterators {
		if iterator.Next() {
			h.iterators = append(h.iterators, iterator)
		}
	}
	heap.Init(&h)
	u.iterators = h.iterators
	u.started = true
}

func (u *union) Value() Value {
	if len(u.iterators) > 0 {
		return u.iterators[0].Value()
	}
	return nil
}

func (u *union) EstimateLength() int {
	if len(u.iterators) == 0 {
		return 0
	}
	max := u.iterators[0].EstimateLength()
	for i := 1; i < len(u.iterators); i++ {
		l := u.iterators[i].EstimateLength()
		if l > max {
			max = l
		}
	}
	return max
}

func NewUnion(iterators []Iterator, values Values) Iterator {
	return &union{iterators: iterators, started: false, values: values}
}

type unionHeap struct {
	iterators []Iterator
	values    Values
}

func (h *unionHeap) Len() int { return len(h.iterators) }

func (h *unionHeap) Less(i, j int) bool {
	return h.values.Compare(h.iterators[i].Value(), h.iterators[j].Value()) == ComparisonLess
}

func (h *unionHeap) Swap(i, j int) {
	h.iterators[i], h.iterators[j] = h.iterators[j], h.iterators[i]
}

func (h *unionHeap) Push(x interface{}) {
	h.iterators = append(h.iterators, x.(Iterator))
}

func (h *unionHeap) Pop() interface{} {
	old := h.iterators
	n := len(old)
	x := old[n-1]
	h.iterators = old[0 : n-1]
	return x
}
