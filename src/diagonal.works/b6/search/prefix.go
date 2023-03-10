package search

import (
	"fmt"
	"strings"
)

type TokenPrefix struct {
	Prefix string
}

func (t TokenPrefix) String() string {
	return fmt.Sprintf("(token-prefix %q)", t.Prefix)
}

func (t TokenPrefix) Compile(index Index) Iterator {
	return newTokenPrefix(t.Prefix, index)
}

type tokenPrefix struct {
	iterator Iterator
	prefix   string
}

func newTokenPrefix(prefix string, index Index) Iterator {
	iterators := make([]Iterator, 0)
	tokens := index.Tokens()
	if !tokens.Advance(prefix) {
		return NewEmptyIterator()
	}

	for strings.HasPrefix(tokens.Token(), prefix) {
		iterators = append(iterators, index.Begin(tokens.Token()))
		if !tokens.Next() {
			break
		}
	}
	return &tokenPrefix{iterator: NewUnion(iterators, index.Values()), prefix: prefix}
}

func (t *tokenPrefix) Next() bool {
	return t.iterator.Next()
}

func (t *tokenPrefix) Advance(key Key) bool {
	return t.iterator.Advance(key)
}

func (t *tokenPrefix) EstimateLength() int {
	return t.iterator.EstimateLength()
}

func (t *tokenPrefix) Value() Value {
	return t.iterator.Value()
}

func (t *tokenPrefix) ToQuery() Query {
	return TokenPrefix{Prefix: t.prefix}
}
