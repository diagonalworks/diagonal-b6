package search

import (
	"sort"
)

type Intersection []Query

func (i Intersection) String() string {
	s := "(intersection"
	for _, q := range i {
		s = s + " " + q.String()
	}
	return s + ")"
}

func (i Intersection) Compile(index Index) Iterator {
	iterators := make([]Iterator, len(i))
	for j, query := range i {
		iterators[j] = query.Compile(index)
	}
	return newIntersection(iterators, index.Values())
}

type byEstimatedLength []Iterator

func (b byEstimatedLength) Len() int           { return len(b) }
func (b byEstimatedLength) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byEstimatedLength) Less(i, j int) bool { return b[i].EstimateLength() < b[j].EstimateLength() }

type intersection struct {
	iterators []Iterator
	values    Values
}

func (in *intersection) Next() bool {
	if in.iterators[0].Next() {
		return in.advanceToNextIntersecion()
	}
	return false
}

func (in *intersection) Advance(to Key) bool {
	if in.iterators[0].Advance(to) {
		return in.advanceToNextIntersecion()
	}
	return false
}

func (in *intersection) advanceToNextIntersecion() bool {
	for {
		intersected := true
		for i := 1; i < len(in.iterators); i++ {
			if !in.iterators[i].Advance(in.values.Key(in.iterators[0].Value())) {
				return false
			}
			if in.values.Compare(in.iterators[i].Value(), in.iterators[0].Value()) != ComparisonEqual {
				if !in.iterators[0].Advance(in.values.Key(in.iterators[i].Value())) {
					return false
				}
				intersected = false
				break
			}
		}
		if intersected {
			return true
		}
	}
}

func (in *intersection) Value() Value {
	return in.iterators[0].Value()
}

func (in *intersection) EstimateLength() int {
	return in.iterators[0].EstimateLength()
}

func (in *intersection) ToQuery() Query {
	query := make(Intersection, len(in.iterators))
	for i, iterator := range in.iterators {
		query[i] = iterator.ToQuery()
	}
	return query
}

func newIntersection(iterators []Iterator, values Values) *intersection {
	// We make an assumption here that the shortest doesn't change during
	// intersection. We use a stable sort to ensure unit tests are
	// deterministic.
	sort.Stable(byEstimatedLength(iterators))
	return &intersection{iterators: iterators, values: values}
}
