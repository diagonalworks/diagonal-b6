package search

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type intValues struct {
	Comparisons int
}

func (i *intValues) Compare(a Value, b Value) Comparison {
	i.Comparisons++
	ai, bi := a.(int), b.(int)
	if ai < bi {
		return ComparisonLess
	} else if ai > bi {
		return ComparisonGreater
	}
	return ComparisonEqual
}

func (i *intValues) CompareKey(v Value, k Key) Comparison {
	return i.Compare(v, k)
}

func (i *intValues) Key(v Value) Key {
	return v
}

type iteratorBuilder func(ints []int) Iterator

func buildArrayIterator(ints []int) Iterator {
	sorted := make([]int, len(ints))
	for i, v := range ints {
		sorted[i] = v
	}
	sort.Ints(sorted)
	vs := make([]Value, len(sorted))
	for i, v := range sorted {
		vs[i] = v
	}
	return &arrayIndexIterator{i: -1, list: vs, token: "", values: &intValues{}}
}

func buildTreeIterator(ints []int) Iterator {
	list := newTreeList(&intValues{})
	for _, v := range ints {
		list.Insert(v)
	}
	return &treeIndexEntryIterator{treeListIterator: *list.Begin(), token: ""}
}

func TestIterators(t *testing.T) {
	builders := []struct {
		name  string
		build iteratorBuilder
	}{
		{"Array", buildArrayIterator},
		{"Tree", buildTreeIterator},
	}

	cases := []struct {
		name string
		f    func(iteratorBuilder, *testing.T)
	}{
		{"Next", ValidateNext},
		{"Advance", ValidateAdvance},
		{"AdvanceToPreviousItem", ValidateAdvanceToPreviousItem},
		{"NextAtEndOfValues", ValidateNextAtEndOfValues},
	}

	for _, builder := range builders {
		for _, c := range cases {
			t.Run(fmt.Sprintf("%s/%s", builder.name, c.name), func(t *testing.T) {
				c.f(builder.build, t)
			})
		}
	}
}

func ValidateNext(build iteratorBuilder, t *testing.T) {
	iterator := build([]int{1, 2, 3, 4, 5})

	for i := 1; i <= 5; i++ {
		if !iterator.Next() {
			t.Errorf("Expected Next() to return true at position %d", i)
		}
		if iterator.Value().(int) != i {
			t.Errorf("Expected Value() to return %d, found %d", i, iterator.Value().(int))
		}
	}

	for i := 0; i < 2; i++ {
		if iterator.Next() {
			t.Errorf("Expected Next() to return false on call %d", i)
		}
	}
}

func ValidateAdvance(build iteratorBuilder, t *testing.T) {
	cases := []struct {
		name     string
		add      []int
		advance  int
		ok       bool
		expected []int
	}{
		{"HappyPath", []int{7, 1, 3, 5, 2, 4, 6}, 4, true, []int{4, 5, 6, 7}},
		{"AdvanceBeforeStart", []int{7, 3, 5, 4, 6}, 2, true, []int{3, 4, 5, 6, 7}},
		{"AdvanceBeyondEnd", []int{7, 1, 3, 5, 2, 4, 6}, 8, false, []int{}},
		{"SequenceMemberMissing", []int{1, 2, 3, 4, 5, 6, 7, 8, 10}, 2, true, []int{2, 3, 4, 5, 6, 7, 8, 10}},
		{"Empty", []int{}, 8, false, []int{}},
		{"NonTrivalBinarySearchTreeOrdering", []int{5, 2, 6, 1, 4}, 3, true, []int{4, 5, 6}},
	}

	for _, next := range []bool{false, true} {
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				i := build(c.add)
				result := make([]int, 0)
				if next {
					i.Next() // Calling Next() before Advance() should make no difference
				}
				ok := i.Advance(c.advance)
				if ok != c.ok {
					t.Fatalf("[ok=%v,add=%v,advance=%v,next=%v] Expected Advance() to return %v, found %v ", c.ok, c.add, c.advance, next, c.ok, ok)
				}
				if ok {
					result = append(result, i.Value().(int))
					for i.Next() {
						result = append(result, i.Value().(int))
					}
				} else if i.Next() {
					t.Errorf("[ok=%v,add=%v,advance=%v,next=%v] Expected Next() to return false if Advance() returned false", c.ok, c.add, c.advance, next)
				}
				if diff := cmp.Diff(c.expected, result); diff != "" {
					t.Errorf("[ok=%v,add=%v,advance=%v,next=%v] comparison got diff (-want, +got):\n%s", c.ok, c.add, c.advance, next, diff)
				}
			})
		}
	}
}

func ValidateAdvanceToPreviousItem(build iteratorBuilder, t *testing.T) {
	input := []int{7, 1, 3, 5, 2, 4, 6}
	i := build(input)

	expected := 3
	if !i.Next() || !i.Next() || !i.Next() || i.Value() != expected {
		t.Errorf("Expected to use Next() to reach %d, found %d", expected, i.Value())
	}

	if !i.Advance(1) {
		t.Error("Expected Advance() to return true")
	}
	if i.Value() != expected {
		t.Errorf("Expected unchanged value of %d, found %d", expected, i.Value())
	}
}

func ValidateNextAtEndOfValues(build iteratorBuilder, t *testing.T) {
	input := []int{7, 1, 3, 5, 2, 4, 6}
	i := build(input)

	for j := 0; j < len(input); j++ {
		if !i.Next() {
			t.Fatal("Expected Next() to return true")
		}
	}

	if i.Next() {
		t.Fatal("Expected Next() to return false")
	}

	lastValue := 7
	if i.Value() != lastValue {
		t.Error("Expected Value() to be unchanged after Next() returns false")
	}

	if i.Next() {
		t.Error("Expected repeated calls to Next() to return false")
	}

	if i.Value() != lastValue {
		t.Error("Expected Value() to unchanged after multiple Next() calling returning false")
	}
}

type indexedInt struct {
	value  int
	tokens []string
}

type indexBuilder func(indexed []indexedInt) Index

func buildArrayIndex(indexed []indexedInt) Index {
	index := NewArrayIndex(&intValues{})
	for _, i := range indexed {
		index.Add(i.value, i.tokens)
	}
	index.Finish(2)
	return index
}

func buildTreeIndex(indexed []indexedInt) Index {
	index := NewTreeIndex(&intValues{})
	for _, i := range indexed {
		index.Add(i.value, i.tokens)
	}
	return index
}

func TestIndices(t *testing.T) {
	builders := []struct {
		name  string
		build indexBuilder
	}{
		{"Array", buildArrayIndex},
		{"Tree", buildTreeIndex},
	}

	cases := []struct {
		name string
		f    func(indexBuilder, *testing.T)
	}{
		{"SimpleAll", ValidateSimpleAll},
		{"Union", ValidateUnion},
		{"UnionsAreDeduplicated", ValidateUnionsAreDeduplicated},
		{"UnionsAreDeduplicatedWithSingleElementLists", ValidateUnionsAreDeduplicatedWithSingleElementLists},
		{"Prefix", ValidatePrefix},
		{"Intersection", ValidateIntersection},
		{"IntersectionNumberOfComparisions", ValidateIntersectionNumberOfComparisons},
		{"AdvanceOnIntersectionToPositionThatIsntAnIntersection", ValidateAdvanceOnIntersectionToPositionThatIsntAnIntersection},
		{"IntersectionOnEmptyUnion", ValidateIntersectionOnEmptyUnion},
		{"KeyRange", ValidateKeyRange},
		{"KeyRangeAdvanceBeyondEndOfRange", ValidateKeyRangeAdvanceBeyondEndOfRange},
		{"EntriesAreDeduplicated", ValidateEntriesAreDeduplicated},
	}

	for _, builder := range builders {
		for _, c := range cases {
			t.Run(fmt.Sprintf("%s/%s", builder.name, c.name), func(t *testing.T) {
				c.f(builder.build, t)
			})
		}
	}
}

func ValidateSimpleAll(build indexBuilder, t *testing.T) {
	indexed := []indexedInt{
		{value: 2, tokens: []string{"0"}},
		{value: 1, tokens: []string{"0"}},
		{value: 4, tokens: []string{"0"}},
		{value: 3, tokens: []string{"0"}},
		{value: 5, tokens: []string{"0"}},
	}

	i := Union{All{"0"}}.Compile(build(indexed))
	result := make([]int, 0)
	for i.Next() {
		result = append(result, i.Value().(int))
	}

	expected := []int{1, 2, 3, 4, 5}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Union got diff (-want, +got):\n%s", diff)
	}
}

func ValidateUnion(build indexBuilder, t *testing.T) {
	indexed := []indexedInt{
		{value: 1, tokens: []string{"0"}},
		{value: 2, tokens: []string{"0"}},
		{value: 3, tokens: []string{"0"}},
		{value: 4, tokens: []string{"0"}},
		{value: 5, tokens: []string{"0"}},
		{value: 6, tokens: []string{"1"}},
		{value: 7, tokens: []string{"1"}},
		{value: 8, tokens: []string{"1"}},
		{value: 9, tokens: []string{"1"}},
		{value: 10, tokens: []string{"1"}},
	}

	i := Union{All{"0"}, All{"1"}}.Compile(build(indexed))
	result := make([]int, 0)
	for i.Next() {
		result = append(result, i.Value().(int))
	}

	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Union got diff (-want, +got):\n%s", diff)
	}
}

func ValidateUnionsAreDeduplicated(build indexBuilder, t *testing.T) {
	indexed := []indexedInt{
		{value: 1, tokens: []string{"0"}},
		{value: 2, tokens: []string{"0"}},
		{value: 3, tokens: []string{"0"}},
		{value: 4, tokens: []string{"0"}},
		{value: 5, tokens: []string{"0", "1"}},
		{value: 6, tokens: []string{"0", "1"}},
		{value: 7, tokens: []string{"2", "1"}},
		{value: 8, tokens: []string{"2", "1"}},
		{value: 9, tokens: []string{"2"}},
		{value: 10, tokens: []string{"2"}},
	}

	i := Union{All{"0"}, All{"1"}, All{"2"}}.Compile(build(indexed))
	result := make([]int, 0)
	for i.Next() {
		result = append(result, i.Value().(int))
	}

	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Union got diff (-want, +got):\n%s", diff)
	}
}

func ValidateUnionsAreDeduplicatedWithSingleElementLists(build indexBuilder, t *testing.T) {
	indexed := []indexedInt{
		{value: 1, tokens: []string{"1"}},
		{value: 2, tokens: []string{"2"}},
		{value: 3, tokens: []string{"3"}},
		{value: 1, tokens: []string{"4"}},
		{value: 3, tokens: []string{"5"}},
	}

	i := Union{All{"1"}, All{"2"}, All{"3"}, All{"4"}, All{"5"}}.Compile(build(indexed))
	result := make([]int, 0)
	for i.Next() {
		result = append(result, i.Value().(int))
	}

	expected := []int{1, 2, 3}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Union got diff (-want, +got):\n%s", diff)
	}
}

func ValidatePrefix(build indexBuilder, t *testing.T) {
	indexed := []indexedInt{
		{value: 1, tokens: []string{"0", "1"}},
		{value: 3, tokens: []string{"1"}},
		{value: 4, tokens: []string{"2"}},
		{value: 5, tokens: []string{"10"}},
		{value: 6, tokens: []string{"10"}},
		{value: 12, tokens: []string{"2"}},
		{value: 13, tokens: []string{"3"}},
		{value: 14, tokens: []string{"4"}},
		{value: 15, tokens: []string{"5"}},
	}

	i := TokenPrefix{Prefix: "1"}.Compile(build(indexed))
	result := make([]int, 0)
	for i.Next() {
		result = append(result, i.Value().(int))
	}

	expected := []int{1, 3, 5, 6}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Prefix got diff (-want, +got):\n%s", diff)
	}
}

func ValidateIntersection(build indexBuilder, t *testing.T) {
	indexed := []indexedInt{
		{value: 1, tokens: []string{"0"}},
		{value: 2, tokens: []string{"0", "2"}},
		{value: 3, tokens: []string{"0"}},
		{value: 4, tokens: []string{"0"}},
		{value: 5, tokens: []string{"0", "2"}},
		{value: 6, tokens: []string{"0", "1"}},
		{value: 7, tokens: []string{"0", "2"}},
		{value: 8, tokens: []string{"0", "1"}},
		{value: 9, tokens: []string{"1", "2"}},
		{value: 10, tokens: []string{"0"}},
	}

	result := make([]int, 0)
	i := Intersection{All{Token: "0"}, All{Token: "2"}}.Compile(build(indexed))
	for i.Next() {
		result = append(result, i.Value().(int))
	}

	expected := []int{2, 5, 7}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Intersection got diff (-want, +got):\n%s", diff)
	}
}

func ValidateIntersectionNumberOfComparisons(build indexBuilder, t *testing.T) {
	indexed := make([]indexedInt, 1024)
	for i := 0; i < 1023; i++ {
		indexed[i] = indexedInt{value: i, tokens: []string{"0"}}
	}
	indexed[1023] = indexedInt{value: 1023, tokens: []string{"0", "1"}}
	index := build(indexed)
	values := index.Values().(*intValues)
	values.Comparisons = 0 // Reset after index creation

	result := make([]int, 0)
	i := Intersection{All{Token: "0"}, All{Token: "1"}}.Compile(index)
	for i.Next() {
		result = append(result, i.Value().(int))
	}

	expected := []int{1023}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Intersection got diff (-want, +got):\n%s", diff)
	}

	if values.Comparisons > 100 {
		t.Errorf("Unexpectedly high number of comparisons: %d", values.Comparisons)
	}
}

func ValidateIntersectionOnEmptyUnion(build indexBuilder, t *testing.T) {
	indexed := []indexedInt{
		{value: 1, tokens: []string{"0"}},
		{value: 2, tokens: []string{"0"}},
		{value: 3, tokens: []string{"0"}},
	}

	result := make([]int, 0)
	i := Intersection{All{Token: "0"}, Union{}}.Compile(build(indexed))
	for i.Next() {
		result = append(result, i.Value().(int))
	}

	expected := []int{}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Intersection got diff (-want, +got):\n%s", diff)
	}
}

func ValidateAdvanceOnIntersectionToPositionThatIsntAnIntersection(build indexBuilder, t *testing.T) {
	indexed := []indexedInt{
		{value: 1, tokens: []string{"0"}},
		{value: 2, tokens: []string{"0"}},
		{value: 3, tokens: []string{"0"}},
		{value: 4, tokens: []string{"0"}},
		{value: 5, tokens: []string{"0"}},
		{value: 6, tokens: []string{"0", "1"}},
		{value: 8, tokens: []string{"0", "1"}},
		{value: 9, tokens: []string{"0", "1"}},
		{value: 10, tokens: []string{"0"}},
		{value: 11, tokens: []string{"1"}},
	}

	i := Intersection{All{"0"}, All{"1"}}.Compile(build(indexed))
	result := make([]int, 0)
	// Advance to a position that isn't an intersection between the lists
	if !i.Advance(7) {
		t.Error("Expected to be able to advance")
	} else {
		for {
			result = append(result, i.Value().(int))
			if !i.Next() {
				break
			}
		}
	}

	expected := []int{8, 9}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Intersection got diff (-want, +got):\n%s", diff)
	}
}

func ValidateKeyRange(build indexBuilder, t *testing.T) {
	indexed := []indexedInt{
		{value: 1, tokens: []string{"0"}},
		{value: 2, tokens: []string{"0"}},
		{value: 3, tokens: []string{"0"}},
		{value: 4, tokens: []string{"0"}},
		{value: 5, tokens: []string{"0"}},
		{value: 6, tokens: []string{"0"}},
		{value: 7, tokens: []string{"0"}},
		{value: 8, tokens: []string{"0"}},
		{value: 10, tokens: []string{"0"}},
	}

	cases := []struct {
		begin    int
		end      int
		expected []int
	}{
		{begin: 2, end: 6, expected: []int{2, 3, 4, 5}}, // Happy path
		{begin: 7, end: 7, expected: []int{}},           // Range is empty
		{begin: 0, end: 1, expected: []int{}},           // Range not empty, but before list
		{begin: 12, end: 13, expected: []int{}},         // Range not empty, but after list
		{begin: 8, end: 13, expected: []int{8, 10}},     // Range runs off the end of the list
	}

	index := build(indexed)
	for _, c := range cases {
		q := KeyRange{Query: All{Token: "0"}, Begin: c.begin, End: c.end}
		result := make([]int, 0)
		i := q.Compile(index)
		for i.Next() {
			result = append(result, i.Value().(int))
		}
		if diff := cmp.Diff(c.expected, result); diff != "" {
			t.Errorf("[begin=%b,end=%v] Got diff (-want, +got):\n%s", c.begin, c.end, diff)
		}
	}
}

func ValidateKeyRangeAdvanceBeyondEndOfRange(build indexBuilder, t *testing.T) {
	indexed := []indexedInt{
		{value: 1, tokens: []string{"0"}},
		{value: 2, tokens: []string{"0"}},
		{value: 3, tokens: []string{"0"}},
		{value: 4, tokens: []string{"0"}},
		{value: 5, tokens: []string{"0"}},
		{value: 6, tokens: []string{"0"}},
		{value: 7, tokens: []string{"0"}},
		{value: 8, tokens: []string{"0"}},
		{value: 10, tokens: []string{"0"}},
	}

	result := make([]int, 0)
	i := KeyRange{Query: All{Token: "0"}, Begin: 2, End: 6}.Compile(build(indexed))
	i.Advance(7)
	for i.Next() {
		result = append(result, i.Value().(int))
	}

	if len(result) != 0 {
		t.Errorf("Expect nothing, found %v", result)
	}
}

func ValidateEntriesAreDeduplicated(build indexBuilder, t *testing.T) {
	indexed := []indexedInt{
		{value: 1, tokens: []string{"0", "0"}},
		{value: 2, tokens: []string{"0"}},
	}

	result := make([]int, 0)
	i := All{"0"}.Compile(build(indexed))
	for i.Next() {
		result = append(result, i.Value().(int))
	}

	expected := []int{1, 2}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Got diff (-want, +got):\n%s", diff)
	}
}

func TestEmptyUnion(t *testing.T) {
	union := NewUnion([]Iterator{}, &intValues{})
	result := make([]int, 0)
	for union.Next() {
		result = append(result, union.Value().(int))
	}

	if len(result) != 0 {
		t.Error("Expected a union of no lists to be empty")
	}
}

func TestQueryString(t *testing.T) {
	q := Intersection{Union{All{"0"}, All{"1"}}, All{"2"}}

	expectedString := "(intersection (union (all \"0\") (all \"1\")) (all \"2\"))"
	if q.String() != expectedString {
		t.Errorf("Expected query string %s, found %s", expectedString, q.String())
	}
}
