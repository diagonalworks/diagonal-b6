package search

import (
	"log"
)

// An implementation of Index based on AVL trees

type treeNode struct {
	parent  *treeNode
	left    *treeNode
	right   *treeNode
	v       Value
	balance int8
}

func (t *treeNode) markDeleted() {
	t.parent = t
}

func (t *treeNode) isDeleted() bool {
	return t.parent == t
}

// treeList represents a sorted list of values, backed by an AVL tree to allow
// efficient modification.
// The implementation largely follows the outline shown on Wikipedia:
// https://en.wikipedia.org/wiki/AVL_tree
type treeList struct {
	root   *treeNode
	length int
	values Values
}

func newTreeList(values Values) *treeList {
	return &treeList{root: nil, length: 0, values: values}
}

func (t *treeList) Len() int {
	return t.length
}

func (t *treeList) Insert(v Value) {
	if t.root == nil {
		t.root = &treeNode{parent: nil, left: nil, right: nil, v: v}
	} else {
		node := t.root
	loop:
		for {
			switch t.values.Compare(node.v, v) {
			case ComparisonEqual:
				node.v = v
				return
			case ComparisonLess:
				if node.right == nil {
					node.right = &treeNode{parent: node, left: nil, right: nil, v: v}
					node = node.right
					break loop
				}
				node = node.right
			case ComparisonGreater:
				if node.left == nil {
					node.left = &treeNode{parent: node, left: nil, right: nil, v: v}
					node = node.left
					break loop
				}
				node = node.left
			}
		}
		t.length++
		t.rebalanceAfterInsert(node)
	}
}

func (t *treeList) rebalanceAfterInsert(child *treeNode) {
	for parent := child.parent; parent != nil; parent = child.parent {
		grandparent := parent.parent
		var newParent *treeNode
		if child == parent.right {
			if parent.balance > 0 {
				if child.balance < 0 {
					newParent = rotateRightLeft(parent, child)
				} else {
					newParent = rotateLeft(parent, child)
				}
			} else {
				parent.balance++
				if parent.balance == 0 {
					break
				}
				child = parent
				continue
			}
		} else {
			if parent.balance < 0 {
				if child.balance > 0 {
					newParent = rotateLeftRight(parent, child)
				} else {
					newParent = rotateRight(parent, child)
				}
			} else {
				parent.balance--
				if parent.balance == 0 {
					break
				}
				child = parent
				continue
			}
		}
		newParent.parent = grandparent
		if grandparent != nil {
			if parent == grandparent.left {
				grandparent.left = newParent
			} else {
				grandparent.right = newParent
			}
		} else {
			t.root = newParent
		}
		break
	}
}

func rotateLeft(parent *treeNode, child *treeNode) *treeNode {
	parent.right = child.left
	if parent.right != nil {
		parent.right.parent = parent
	}
	child.left = parent
	child.left.parent = child

	if child.balance == 0 {
		parent.balance = 1
		child.balance = -1
	} else {
		child.balance = 0
		parent.balance = 0
	}
	return child
}

func rotateRight(parent *treeNode, child *treeNode) *treeNode {
	parent.left = child.right
	if parent.left != nil {
		parent.left.parent = parent
	}
	child.right = parent
	child.right.parent = child

	if child.balance == 0 {
		parent.balance = -1
		child.balance = 1
	} else {
		child.balance = 0
		parent.balance = 0
	}
	return child
}

func rotateRightLeft(parent *treeNode, child *treeNode) *treeNode {
	// Rotate right
	newParent := child.left
	child.left = newParent.right
	if child.left != nil {
		child.left.parent = child
	}
	newParent.right = child
	newParent.right.parent = newParent

	// Rotate left
	parent.right = newParent.left
	if parent.right != nil {
		parent.right.parent = parent
	}
	newParent.left = parent
	newParent.left.parent = newParent

	if newParent.balance > 0 {
		parent.balance = -1
		child.balance = 0
	} else if newParent.balance == 0 {
		parent.balance = 0
		child.balance = 0
	} else {
		parent.balance = 0
		child.balance = 1
	}
	newParent.balance = 0

	return newParent
}

func rotateLeftRight(parent *treeNode, child *treeNode) *treeNode {
	// Rotate left
	newParent := child.right
	child.right = newParent.left
	if child.right != nil {
		child.right.parent = child
	}
	newParent.left = child
	newParent.left.parent = newParent

	// Rotate right
	parent.left = newParent.right
	if parent.left != nil {
		parent.left.parent = parent
	}
	newParent.right = parent
	newParent.right.parent = newParent

	if newParent.balance < 0 {
		parent.balance = 1
		child.balance = 0
	} else if newParent.balance == 0 {
		parent.balance = 0
		child.balance = 0
	} else {
		parent.balance = 0
		child.balance = -1
	}
	newParent.balance = 0

	return newParent
}

func (t *treeList) Delete(v Value) {
	t.DeleteKey(t.values.Key(v))
}

func (t *treeList) DeleteKey(k Key) {
	node := t.root
loop:
	for node != nil {
		switch t.values.CompareKey(node.v, k) {
		case ComparisonEqual:
			if node.left != nil && node.right != nil {
				next := t.findMinimum(node.right)
				t.replaceInGrandparent(next, next.right)
				// While we could simply set node.v = next.v, it would invalidate iterators
				// currently at next. Instead, we graft next into the correct position in
				// the tree.
				next.parent = node.parent
				next.left = node.left
				if next.left != nil {
					next.left.parent = next
				}
				next.right = node.right
				if next.right != nil {
					next.right.parent = next
				}
				next.balance = node.balance
				if node.parent != nil {
					if node.parent.left == node {
						node.parent.left = next
					} else if node.parent.right == node {
						node.parent.right = next
					} else {
						panic("Node is not a child of its parent")
					}
				} else {
					t.root = next
				}
			} else if node.left != nil {
				t.replaceInGrandparent(node, node.left)
			} else if node.right != nil {
				t.replaceInGrandparent(node, node.right)
			} else {
				t.replaceInGrandparent(node, nil)
			}
			node.markDeleted()
			break loop
		case ComparisonLess:
			node = node.right
		case ComparisonGreater:
			node = node.left
		}
	}
	t.length--
}

func (t *treeList) findMinimum(node *treeNode) *treeNode {
	if node == nil {
		return nil
	}
	for node.left != nil {
		node = node.left
	}
	return node
}

func (t *treeList) replaceInGrandparent(parent *treeNode, child *treeNode) {
	if parent.parent == nil {
		t.root = child
		if child != nil {
			child.parent = nil
		}
	} else {
		t.rebalanceBeforeDelete(parent)
		grandparent := parent.parent
		if child != nil {
			child.parent = grandparent
		}
		if grandparent.left == parent {
			grandparent.left = child
		} else {
			grandparent.right = child
		}
	}
}

func (t *treeList) rebalanceBeforeDelete(child *treeNode) {
	for parent := child.parent; parent != nil; parent = child.parent {
		grandparent := parent.parent
		var newParent *treeNode
		var balance int8
		if child == parent.left {
			if parent.balance > 0 {
				sibling := parent.right
				balance = sibling.balance
				if balance < 0 {
					newParent = rotateRightLeft(parent, sibling)
				} else {
					newParent = rotateLeft(parent, sibling)
				}
			} else {
				parent.balance++
				if parent.balance == 1 {
					break
				}
				child = parent
				continue
			}
		} else {
			if parent.balance < 0 {
				sibling := parent.left
				balance = sibling.balance
				if balance > 0 {
					newParent = rotateLeftRight(parent, sibling)
				} else {
					newParent = rotateRight(parent, sibling)
				}
			} else {
				parent.balance--
				if parent.balance == -1 {
					break
				}
				child = parent
				continue
			}
		}
		newParent.parent = grandparent
		if grandparent != nil {
			if parent == grandparent.left {
				grandparent.left = newParent
			} else {
				grandparent.right = newParent
			}
		} else {
			t.root = newParent
		}
		child = newParent
		if balance == 0 {
			break
		}
	}
}

func (t *treeList) Validate() bool {
	ok, _ := t.validate(t.root, nil)
	return ok
}

func (t *treeList) validate(node *treeNode, parent *treeNode) (bool, int) {
	ok := true
	count := 0
	if node != nil {
		leftOK, leftCount := t.validate(node.left, node)
		dl, dr := t.depth(node.left), t.depth(node.right)
		if node.balance != int8(dr-dl) {
			log.Printf("Unexpected balance at node %v: %d vs %d (%d-%d)", t.values.Key(node.v), node.balance, dr-dl, dr, dl)
			ok = false
		}
		if node.left != nil && t.values.Compare(node.left.v, node.v) != ComparisonLess {
			log.Printf("Left not is not less at %s", t.values.Key(node.v))
			ok = false
		}
		if node.parent != parent {
			log.Printf("Broken parent: %v (%p vs %p, root %p)", t.values.Key(node.parent.v), node.parent, parent, t.root)
			ok = false
		}
		if node.right != nil && t.values.Compare(node.right.v, node.v) != ComparisonGreater {
			log.Printf("Right not is not greater at %s", t.values.Key(node.v))
			ok = false
		}
		rightOK, rightCount := t.validate(node.right, node)
		ok = ok && leftOK && rightOK
		count += 1 + leftCount + rightCount
	}
	return ok, count
}

func (t *treeList) depth(node *treeNode) int {
	if node != nil {
		l, r := t.depth(node.left), t.depth(node.right)
		if l > r {
			return 1 + l
		} else {
			return 1 + r
		}
	}
	return 0
}

func (t *treeList) Log() {
	t.log(t.root, 0)
}

func (t *treeList) log(node *treeNode, depth int) {
	if node == nil {
		return
	}
	t.log(node.left, depth+1)
	prefix := ""
	for i := 0; i < depth; i++ {
		prefix += "  "
	}
	log.Printf("%s%v b=%d", prefix, node.v, node.balance)
	t.log(node.right, depth+1)
}

func (t *treeList) Begin() *treeListIterator {
	return newtreeListIterator(t, t.values)
}

func (t *treeList) Lookup(key Key) (Value, bool) {
	node := t.root
	for node != nil {
		switch t.values.CompareKey(node.v, key) {
		case ComparisonEqual:
			return node.v, true
		case ComparisonLess:
			node = node.right
		case ComparisonGreater:
			node = node.left
		}
	}
	return nil, false
}

type treeListIterator struct {
	list    *treeList
	node    *treeNode
	values  Values
	started bool
	done    bool
}

func newtreeListIterator(list *treeList, values Values) *treeListIterator {
	return &treeListIterator{list: list, values: values, started: false}
}

func (t *treeListIterator) Next() bool {
	if !t.started {
		return t.start()
	}

	if t.node == nil || t.done {
		return false
	}

	if t.node.isDeleted() {
		key := t.values.Key(t.node.v)
		t.started = false
		ok := t.Advance(key)
		if ok {
			if t.values.CompareKey(t.node.v, key) == ComparisonEqual {
				return t.Next()
			}
		}
		return ok
	}

	if t.node.right != nil {
		t.node = t.node.right
		for t.node.left != nil {
			t.node = t.node.left
		}
		return true
	}

	// Ascend to the first parent node larger than the current node, stopping at the root.
	node := t.node
	for {
		if node.parent == nil {
			t.done = true
			return false
		} else if node == node.parent.left {
			t.node = node.parent
			return true
		}
		node = node.parent
	}
}

func (t *treeListIterator) start() bool {
	next := t.list.root
	for next != nil {
		t.node = next
		next = next.left
	}
	t.started = true
	return t.node != nil
}

func (t *treeListIterator) Advance(key Key) bool {
	if !t.started {
		if !t.start() {
			return false
		}
	}

	if t.node == nil {
		return false
	}

	if t.node.isDeleted() {
		start := t.values.Key(t.node.v)
		t.started = false
		return t.Advance(start) && t.Advance(key)
	}

	if t.values.CompareKey(t.node.v, key) != ComparisonLess {
		return true
	}

	// Ascend to the first parent node larger than the key, then redescend to the first value
	// after the key
	for t.node != nil {
		if c := t.values.CompareKey(t.node.v, key); c != ComparisonLess || t.node.parent == nil {
			break
		}
		t.node = t.node.parent
	}

	node := t.node
	t.node = nil
	for node != nil {
		c := t.values.CompareKey(node.v, key)
		if c == ComparisonLess {
			node = node.right
		} else if c == ComparisonGreater {
			t.node = node
			node = node.left
		} else { // ComparisonEqual
			t.node = node
			break
		}
	}
	if t.node == nil {
		t.done = true
	}
	return t.node != nil
}

func (t *treeListIterator) EstimateLength() int {
	return t.list.length // Doesn't take into account the iterator's position
}

func (t *treeListIterator) Key() Key {
	if t.node != nil {
		return t.values.Key(t.node.v)
	}
	return nil
}

func (t *treeListIterator) Value() Value {
	if t.node != nil {
		return t.node.v
	}
	return nil
}

type treeIndexEntry struct {
	token string
	list  *treeList
}

func newTreeIndexEntry(token string, values Values) treeIndexEntry {
	return treeIndexEntry{token: token, list: newTreeList(values)}
}

func (t *treeIndexEntry) Value() interface{} {
	return t.list
}

// TreeIndex is an implementation of Index backed by AVL trees.
// It's more expensive in terms of storage than ArrayIndex, though the use
// of AVL trees makes the Index modifiable.
type TreeIndex struct {
	lists  *treeList
	values Values
}

type treeIndexTokenValues struct{}

func (t treeIndexTokenValues) Compare(a Value, b Value) Comparison {
	tokenA, tokenB := a.(treeIndexEntry).token, b.(treeIndexEntry).token
	if tokenA < tokenB {
		return ComparisonLess
	} else if tokenA > tokenB {
		return ComparisonGreater
	}
	return ComparisonEqual
}

func (t treeIndexTokenValues) CompareKey(v Value, k Key) Comparison {
	tokenA, tokenB := v.(treeIndexEntry).token, k.(string)
	if tokenA < tokenB {
		return ComparisonLess
	} else if tokenA > tokenB {
		return ComparisonGreater
	}
	return ComparisonEqual
}

func (t treeIndexTokenValues) Key(v Value) Key {
	return v.(treeIndexEntry).token
}

func NewTreeIndex(values Values) *TreeIndex {
	return &TreeIndex{lists: newTreeList(treeIndexTokenValues{}), values: values}
}

type treeIndexEntryIterator struct {
	treeListIterator
	token string
}

func (t *TreeIndex) Begin(token string) Iterator {
	if entry, ok := t.lists.Lookup(token); ok {
		list := entry.(treeIndexEntry).list
		return &treeIndexEntryIterator{treeListIterator: *list.Begin(), token: token}
	}
	return NewEmptyIterator()
}

type treeTokenIterator struct {
	iterator *treeListIterator
}

func (t treeTokenIterator) Token() string {
	return t.iterator.Value().(treeIndexEntry).token
}

func (t treeTokenIterator) Next() bool {
	return t.iterator.Next()
}

func (t treeTokenIterator) Advance(token string) bool {
	return t.iterator.Advance(token)
}

func (t *TreeIndex) Tokens() TokenIterator {
	return treeTokenIterator{iterator: t.lists.Begin()}
}

func (t *TreeIndex) NumTokens() int {
	return t.lists.Len()
}

func (t *TreeIndex) Add(v Value, tokens []string) {
	for _, token := range tokens {
		var list *treeList
		if e, ok := t.lists.Lookup(token); ok {
			list = e.(treeIndexEntry).list
		} else {
			entry := newTreeIndexEntry(token, t.values)
			t.lists.Insert(entry)
			list = entry.list
		}
		list.Insert(v)
	}
}

func (t *TreeIndex) Remove(v Value, tokens []string) {
	for _, token := range tokens {
		if e, ok := t.lists.Lookup(token); ok {
			e.(treeIndexEntry).list.Delete(v)
		}
	}
}

func (t *TreeIndex) Values() Values {
	return t.values
}
