package search

import (
	"fmt"
)

type Comparison int

const (
	ComparisonLess    Comparison = -1
	ComparisonEqual   Comparison = 0
	ComparisonGreater Comparison = 1
)

type Key interface{}
type Value interface{}

type Values interface {
	Compare(a Value, b Value) Comparison
	CompareKey(v Value, k Key) Comparison
	Key(v Value) Key
}

type Iterator interface {
	Next() bool
	Advance(key Key) bool
	Value() Value
	EstimateLength() int
	ToQuery() Query
}

type TokenIterator interface {
	Next() bool
	Advance(token string) bool
	Token() string
}

func AllTokens(t TokenIterator) []string {
	tokens := make([]string, 0)
	for t.Next() {
		tokens = append(tokens, t.Token())
	}
	return tokens
}

type Index interface {
	Begin(token string) Iterator
	Tokens() TokenIterator
	Values() Values
	NumTokens() int
}

type Query interface {
	Compile(index Index) Iterator
	String() string
}

type Empty struct{}

func (_ Empty) String() string {
	return "(empty)"
}

func (_ Empty) Compile(index Index) Iterator {
	return NewEmptyIterator()
}

type emptyIterator struct{}

func (e emptyIterator) Next() bool {
	return false
}

func (e emptyIterator) Advance(key Key) bool {
	return false
}

func (e emptyIterator) EstimateLength() int {
	return 0
}

func (e emptyIterator) Value() Value {
	return nil
}

func (e emptyIterator) ToQuery() Query {
	return Empty{}
}

func NewEmptyIterator() Iterator {
	return emptyIterator{}
}

const AllToken = "*"

type All struct {
	Token string
}

func (a All) String() string {
	return fmt.Sprintf("(all %q)", a.Token)
}

func (a All) Compile(index Index) Iterator {
	return index.Begin(a.Token)
}
