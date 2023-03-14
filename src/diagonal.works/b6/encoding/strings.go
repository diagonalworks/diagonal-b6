package encoding

import (
	"fmt"
	"io"
	"sort"
	"sync"
)

type StringTableBuilder struct {
	strings map[string]uint32
	length  int
	lock    sync.Mutex
}

// TODO: Maybe it's more practical to give the offset on writing later?
func NewStringTableBuilder() *StringTableBuilder {
	return &StringTableBuilder{strings: make(map[string]uint32), length: -1}
}

func (b *StringTableBuilder) Add(str string) {
	b.lock.Lock()
	b.strings[str]++
	b.lock.Unlock()
}

func (b *StringTableBuilder) NumStrings() int {
	return len(b.strings)
}

type countedStrings struct {
	strings []string
	counts  []uint32
}

func (c countedStrings) Len() int { return len(c.strings) }
func (c countedStrings) Swap(i, j int) {
	c.strings[i], c.strings[j] = c.strings[j], c.strings[i]
	c.counts[i], c.counts[j] = c.counts[j], c.counts[i]
}
func (c countedStrings) Less(i, j int) bool { return c.counts[i] > c.counts[j] }

func (b *StringTableBuilder) Write(w io.WriterAt, offset Offset) (Offset, error) {
	start := offset
	b.lock.Lock()
	defer b.lock.Unlock()

	counted := countedStrings{strings: make([]string, 0, len(b.strings)), counts: make([]uint32, 0, len(b.strings))}
	for s, count := range b.strings {
		counted.strings = append(counted.strings, s)
		counted.counts = append(counted.counts, count)
	}
	sort.Sort(counted)
	for i, s := range counted.strings {
		b.strings[s] = uint32(i)
	}

	arrays := NewByteArraysBuilder(len(counted.strings))
	for i, s := range counted.strings {
		arrays.Reserve(i, len(s))
	}
	var err error
	if offset, err = arrays.WriteHeader(w, offset); err == nil {
		for i, s := range counted.strings {
			if err = arrays.WriteItem(w, i, []byte(s)); err != nil {
				break
			}
		}
	}
	b.length = offset.Difference(start)
	return offset, err
}

func (b *StringTableBuilder) Lookup(s string) int {
	if b.length < 0 {
		panic("Must call Write() before Lookup()")
	}
	if i, ok := b.strings[s]; ok {
		return int(i)
	}
	panic(fmt.Sprintf("Lookup(): string %q not added", s))
}

func (b *StringTableBuilder) Length() int {
	if b.length < 0 {
		panic("Must call Write() before Length()")
	}
	return b.length
}

type Strings interface {
	Lookup(i int) string
	Equal(i int, other string) bool
}

type StringMap map[int]string // For testing

func (m StringMap) Lookup(i int) string {
	return m[i]
}

func (m StringMap) Equal(i int, other string) bool {
	return m[i] == other
}

type StringTable struct {
	arrays *ByteArrays
}

func NewStringTable(data []byte) *StringTable {
	return &StringTable{
		arrays: NewByteArrays(data),
	}
}

func (t *StringTable) Lookup(i int) string {
	return string(t.arrays.Item(i))
}

func (t *StringTable) Equal(i int, other string) bool {
	item := t.arrays.Item(i)
	if len(item) != len(other) {
		return false
	}
	for i := 0; i < len(item); i++ {
		if item[i] != other[i] {
			return false
		}
	}
	return true
}
