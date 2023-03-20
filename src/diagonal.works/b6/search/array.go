package search

import (
	"log"
	"sort"
	"sync"
)

// ArrayIndex is an implementation of Index backed by simple arrays.
// It provides efficient storage, but needs the arrays to be explicitly sorted
// (via the Finish method) before use, and therefore isn't sutable for
// modification.
type ArrayIndex struct {
	lists  [][]Value
	tokens *Tokens
	values Values
}

func NewArrayIndex(values Values) *ArrayIndex {
	return &ArrayIndex{lists: make([][]Value, 0), tokens: NewTokens(), values: values}
}

func (a *ArrayIndex) Add(v Value, tokens []string) {
	for _, token := range tokens {
		i := a.tokens.LookupOrAdd(token)
		if i == len(a.lists) {
			a.lists = append(a.lists, make([]Value, 0, 1))
		}
		a.lists[i] = append(a.lists[i], v)
	}
}

func (a *ArrayIndex) Swap(i int, j int) {
	a.lists[i], a.lists[j] = a.lists[j], a.lists[i]
}

func (a *ArrayIndex) Values() Values {
	return a.values
}

func (a *ArrayIndex) Tokens() TokenIterator {
	return a.tokens.Tokens()
}

func (a *ArrayIndex) NumTokens() int {
	return a.tokens.Len()
}

type byKey struct {
	sort []Value
	v    Values
}

func (k byKey) Len() int           { return len(k.sort) }
func (k byKey) Swap(i, j int)      { k.sort[i], k.sort[j] = k.sort[j], k.sort[i] }
func (k byKey) Less(i, j int) bool { return k.v.Compare(k.sort[i], k.sort[j]) == ComparisonLess }

func (a *ArrayIndex) Finish(cores int) {
	a.tokens.Sort(a)
	sortList := func(i int) error {
		sort.Sort(byKey{sort: a.lists[i], v: a.values})
		// Deduplicate the KeyValues associated with each token
		// TODO: Make this clear in the interface documentation, since it's an
		// expensive operation
		read := 0
		write := 0
		for read < len(a.lists[i]) {
			if read != write {
				a.lists[i][write] = a.lists[i][read]
			}
			read++
			for read < len(a.lists[i]) && a.values.Compare(a.lists[i][read], a.lists[i][write]) == ComparisonEqual {
				read++
			}
			write++
		}
		a.lists[i] = a.lists[i][0:write]
		return nil
	}
	mapNWithLimit(sortList, len(a.lists), cores)
}

// TODO: Refactor
func mapNWithLimit(f func(i int) error, n int, goroutines int) error {
	wg := sync.WaitGroup{}
	wg.Add(goroutines)
	errors := make([]error, n)
	i := 0
	iLock := sync.Mutex{}
	for j := 0; j < goroutines; j++ {
		go func() {
			defer wg.Done()
			for {
				iLock.Lock()
				if i >= n {
					iLock.Unlock()
					return
				}
				run := i
				i++
				iLock.Unlock()
				errors[run] = f(run)
			}
		}()
	}
	wg.Wait()

	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

type arrayIndexIterator struct {
	i      int
	list   []Value
	token  string
	values Values
}

func (a *arrayIndexIterator) Next() bool {
	if a.i+1 >= len(a.list) {
		return false
	}
	a.i++
	return true
}

func (a *arrayIndexIterator) Advance(key Key) bool {
	if a.i < 0 {
		a.i = 0
	}
	a.i += sort.Search(len(a.list)-a.i, func(j int) bool {
		return a.values.CompareKey(a.list[a.i+j], key) != ComparisonLess
	})
	return a.i < len(a.list)
}

func (a *arrayIndexIterator) EstimateLength() int {
	return len(a.list) - a.i
}

func (a *arrayIndexIterator) Value() Value {
	return a.list[a.i]
}

func (a *ArrayIndex) Begin(token string) Iterator {
	if i, ok := a.tokens.Lookup(token); ok {
		return &arrayIndexIterator{i: -1, list: a.lists[i], token: token, values: a.values}
	}
	return NewEmptyIterator()
}

func (a *ArrayIndex) token(i int) string {
	return a.tokens.Token(i)
}

func (a *ArrayIndex) LogContents() {
	for i := 0; i < a.tokens.Len(); i++ {
		log.Printf("%s:", a.tokens.Token(i))
		for j := 0; j < len(a.lists[i]); j++ {
			log.Printf("  %s", a.lists[i][j])
		}
	}
}
