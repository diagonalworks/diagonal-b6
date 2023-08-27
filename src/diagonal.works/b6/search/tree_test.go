package search

import (
	"sort"
	"testing"
)

func TestTreeListAdvanceFromIteratorAtRoot(t *testing.T) {
	// See ValidateAdvance for most tests.
	// Tests here are specific to the implementation of treeList.
	input := []int{10, 5, 15, 4, 6, 13, 17}
	tree := newTreeList(&intValues{})
	for _, x := range input {
		tree.Insert(x)
	}

	i := tree.Begin()
	for i.Next() && i.Value() != 10 {
	}
	if i.Value() != 10 {
		t.Fatal("Expected to find 10")
	}

	if !i.Advance(17) || i.Value() != 17 {
		t.Error("Expected to be able to advance to 17")
	}
}

func TestTreeListRebalancing(t *testing.T) {
	cases := [][]int{
		{10, 5, 15, 4, 3},   // Triggers rotateRight
		{10, 5, 15, 16, 17}, // Triggers rotateLeft
		{10, 5, 15, 1, 2},   // Triggers rotateLeftRight
		{10, 5, 15, 7, 6},   // Triggers rotateRightLeft
	}

	for _, c := range cases {
		tree := newTreeList(&intValues{})
		expected := make([]int, len(c))
		for i, x := range c {
			tree.Insert(x)
			if !tree.Validate() {
				t.Errorf("Tree failed to validate after adding %d in %v", x, c)
			}
			expected[i] = x
		}
		sort.Ints(expected)

		result := make([]int, 0)
		i := tree.Begin()
		for i.Next() {
			result = append(result, i.Value().(int))
		}

		if !equals(result, expected) {
			t.Errorf("Expected %v, found %v for case %v", expected, result, c)
		}
	}
}

func TestTreeListReplaceItem(t *testing.T) {
	input := []int{10, 5, 15, 5, 20}

	tree := newTreeList(&intValues{})
	for _, x := range input {
		tree.Insert(x)
		if !tree.Validate() {
			t.Errorf("Tree validation failed after adding %d", x)
		}
	}

	result := make([]int, 0)
	i := tree.Begin()
	for i.Next() {
		result = append(result, i.Value().(int))
	}

	expected := []int{5, 10, 15, 20}
	if !equals(result, expected) {
		t.Errorf("Expected %v, found %v", expected, result)
	}
}

func TestTreeListDelete(t *testing.T) {
	cases := []struct {
		input  []int
		delete int
	}{
		{[]int{10, 5, 15, 4, 6, 13, 16}, 4},        // Happy path, no rebalancing
		{[]int{10, 5, 15, 4, 6, 13, 16, 17}, 13},   // Triggers rotateLeft
		{[]int{10, 5, 15, 4, 6, 13, 16, 3}, 6},     // Triggers rotateRight
		{[]int{10, 5, 15, 3, 6, 13, 16, 4}, 6},     // Triggers rotateLeftRight
		{[]int{10, 5, 15, 4, 6, 13, 17, 16}, 13},   // Triggers rotateRightLeft
		{[]int{10, 5, 15, 4, 6, 13, 16, 17, 3}, 5}, // Delete a non-leaf node, triggering rotateRight
		{[]int{10, 5, 15, 4, 6, 13, 16}, 10},       // Delete the root
		{[]int{10, 5, 15, 4, 6, 13, 16}, 42},       // Delete a key that isn't present
		{[]int{10}, 10},                            // Delete the root, when it's the only item
		{[]int{}, 10},                              // Delete from an empty tree
	}

	for _, c := range cases {
		tree := newTreeList(&intValues{})
		for _, x := range c.input {
			tree.Insert(x)
		}

		tree.Delete(c.delete)
		if !tree.Validate() {
			t.Errorf("Tree validation failed after deleting %d from %v", c.delete, c.input)
		}

		result := make([]int, 0)
		i := tree.Begin()
		for i.Next() {
			result = append(result, i.Value().(int))
		}

		expected := make([]int, 0, len(c.input))
		for _, x := range c.input {
			if x != c.delete {
				expected = append(expected, x)
			}
		}
		sort.Ints(expected)
		if !equals(result, expected) {
			t.Errorf("Expected %v, found %v input: %v delete: %v", expected, result, c.input, c.delete)
		}
	}
}

func TestTreeListNextOnDeletedIterator(t *testing.T) {
	input := []int{10, 5, 15, 4, 6, 13, 16, 12}

	tree := newTreeList(&intValues{})
	for _, v := range input {
		tree.Insert(v)
	}

	delete := 6
	i := tree.Begin()
	if !i.Next() || !i.Next() || !i.Next() || i.Value() != delete {
		t.Fatalf("Expected to use Next() to reach %d, found %d", delete, i.Value())
	}
	tree.Delete(delete)
	if i.Value() != delete {
		t.Fatalf("Expected value on deleted iterator to be %d, found %d", delete, i.Value())
	}

	if !i.Next() {
		t.Fatalf("Expected to be able to call Next() on a deleted iterator")
	}
	expected := 10
	if i.Value() != expected {
		t.Errorf("Expected %d, found %d", expected, i.Value())
	}
}

func TestTreeListAdvanceOnDeletedIterator(t *testing.T) {
	//4, 5, 6, 10, 12, 13, 15, 16
	input := []int{10, 5, 15, 4, 6, 13, 16, 12}
	cases := []struct {
		name     string
		delete   int
		advance  int
		ok       bool
		expected []int
	}{
		{"HappyPath", 6, 10, true, []int{10, 12, 13, 15, 16}},
		{"AdvanceToDeleted", 6, 6, true, []int{10, 12, 13, 15, 16}},
		{"AdvanceToPrevious", 6, 5, true, []int{10, 12, 13, 15, 16}},
		// TODO: Make this test case pass.
		// {"AdvanceBeyondEnd", 6, 17, true, []int{}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tree := newTreeList(&intValues{})
			for _, v := range input {
				tree.Insert(v)
			}

			i := tree.Begin()
			found := false
			for i.Next() {
				if i.Value() == c.delete {
					found = true
					break
				}
			}

			if !found {
				t.Fatalf("Failed to find %d", c.delete)
			}
			tree.Delete(c.delete)
			var ok bool
			if ok = i.Advance(c.advance); ok != c.ok {
				t.Fatalf("Expected Advance() to return %v, found %v", c.ok, ok)
			}
			if ok {
				result := make([]int, 0)
				result = append(result, i.Value().(int))
				for i.Next() {
					result = append(result, i.Value().(int))
				}
				if !equals(result, c.expected) {
					t.Errorf("Expected %v, found %v with delete: %d advance: %d", c.expected, result, c.delete, c.advance)
				}

			} else if i.Next() {
				t.Error("Expected Next() to return false if Advance() returned false")
			}
		})
	}
}

func TestTreeListInsertAndRemove(t *testing.T) {
	// This case previously failed, as we forgot to clear the parent
	// pointer when giving the tree a new root.
	tree := newTreeList(&intValues{})
	tree.Insert(5)
	tree.Insert(10)
	tree.Delete(5)
	tree.Delete(10)

	result := make([]int, 0)
	i := tree.Begin()
	for i.Next() {
		result = append(result, i.Value().(int))
	}

	if len(result) != 0 {
		t.Errorf("Expected %v, found %v", []int{}, result)
	}
}

func TestTreeListLookup(t *testing.T) {
	tree := newTreeList(&intValues{})
	tree.Insert(10)
	tree.Insert(5)
	tree.Insert(15)

	expected := []int{5, 15}
	for _, e := range expected {
		found, ok := tree.Lookup(e)
		if !ok || found.(int) != e {
			t.Errorf("Expected to find %d", e)
		}
	}

	found, ok := tree.Lookup(42)
	if ok || found != nil {
		t.Errorf("Expected not to find value")
	}
}

func TestTreeIndexAdd(t *testing.T) {
	tree := NewTreeIndex(&intValues{})
	tree.Add(1, []string{"0"})
	tree.Add(3, []string{"0", "2"})

	q := All{"0"}
	result := make([]int, 0)
	i := q.Compile(tree)
	for i.Next() {
		result = append(result, i.Value().(int))
	}

	expected := []int{1, 3}
	if !equals(result, expected) {
		t.Errorf("Expected %v, found %v", expected, result)
	}

	tree.Add(2, []string{"0", "2"})

	result = make([]int, 0)
	i = q.Compile(tree)
	for i.Next() {
		result = append(result, i.Value().(int))
	}

	expected = []int{1, 2, 3}
	if !equals(result, expected) {
		t.Errorf("Expected %v, found %v", expected, result)
	}
}

func TestTreeIndexRemove(t *testing.T) {
	tree := NewTreeIndex(&intValues{})
	tree.Add(1, []string{"0"})
	tree.Add(2, []string{"0", "2"})
	tree.Add(3, []string{"0", "2"})

	tree.Remove(2, []string{"0"})

	cases := []struct {
		query    Query
		expected []int
	}{
		{All{"0"}, []int{1, 3}},
		{All{"2"}, []int{2, 3}},
	}

	for _, c := range cases {
		result := make([]int, 0)
		i := c.query.Compile(tree)
		for i.Next() {
			result = append(result, i.Value().(int))
		}

		if !equals(result, c.expected) {
			t.Errorf("Expected %v, found %v", c.expected, result)
		}
	}
}

func TestTreeIndexDeleteWhileIterating(t *testing.T) {
	tree := newTreeList(&intValues{})
	for _, v := range []int{10, 5, 15, 4, 6, 13, 16, 12} {
		tree.Insert(v)
	}

	result := make([]int, 0)
	i := tree.Begin()
	for i.Next() {
		if i.Value() == 12 {
			tree.Delete(13)
		}
		result = append(result, i.Value().(int))
	}

	expected := []int{4, 5, 6, 10, 12, 15, 16}
	if !equals(result, expected) {
		t.Errorf("Expected %v, found %v", expected, result)
	}
}

func TestTreeIndexDeleteAndInsertWhileIterating(t *testing.T) {
	input := []int{10, 5, 15, 4, 6, 13, 16, 12}

	for _, delete := range input {
		for _, at := range input {
			tree := newTreeList(&intValues{})
			for _, v := range input {
				tree.Insert(v)
			}

			result := make([]int, 0)
			i := tree.Begin()
			for i.Next() {
				if i.Value() == at {
					tree.Delete(delete)
					tree.Insert(delete)
				}
				result = append(result, i.Value().(int))
			}

			expected := []int{4, 5, 6, 10, 12, 13, 15, 16}
			if !equals(result, expected) {
				t.Errorf("Expected %v, found %v with delete %d at %d", expected, result, delete, at)
			}
		}
	}
}
