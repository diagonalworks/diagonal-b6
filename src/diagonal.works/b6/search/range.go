package search

import (
	"fmt"
)

type KeyRange struct {
	Begin Key
	End   Key
	Query Query
}

func (k KeyRange) String() string {
	return fmt.Sprintf("(key-range %q %q %s)", k.Begin, k.End, k.Query.String())
}

func (k KeyRange) Compile(index Index) Iterator {
	return newKeyRange(k.Query.Compile(index), k.Begin, k.End, index.Values())
}

type keyRange struct {
	iterator Iterator
	begin    Key
	end      Key
	started  bool
	values   Values
}

func (k *keyRange) Next() bool {
	var ok bool
	if !k.started {
		k.started = true
		ok = k.iterator.Advance(k.begin)
	} else {
		ok = k.iterator.Next()
	}
	return ok && k.values.CompareKey(k.iterator.Value(), k.end) == ComparisonLess
}

func (k *keyRange) Advance(key Key) bool {
	if !k.started {
		k.started = true
		if !k.iterator.Advance(k.begin) {
			return false
		}
	}
	return k.iterator.Advance(key) && k.values.CompareKey(k.iterator.Value(), k.end) == ComparisonLess
}

func (k *keyRange) EstimateLength() int {
	return k.iterator.EstimateLength()
}

func (k *keyRange) Value() Value {
	return k.iterator.Value()
}

func (k *keyRange) ToQuery() Query {
	return KeyRange{Begin: k.begin, End: k.end, Query: k.iterator.ToQuery()}
}

func newKeyRange(iterator Iterator, begin Key, end Key, values Values) *keyRange {
	return &keyRange{iterator: iterator, begin: begin, end: end, started: false, values: values}
}
