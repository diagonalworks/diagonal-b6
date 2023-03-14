package search

import (
	"fmt"
	"sort"
	"strings"
)

type SortableIndex interface {
	Swap(i int, j int)
}

type tokenIterator struct {
	i      int
	tokens []string
}

func newTokenIterator(tokens []string) *tokenIterator {
	return &tokenIterator{i: -1, tokens: tokens}
}

func (t *tokenIterator) Token() string {
	return t.tokens[t.i]
}

func (t *tokenIterator) Next() bool {
	t.i++
	return t.i < len(t.tokens)
}

func (t *tokenIterator) Advance(token string) bool {
	if t.i < 0 {
		t.i = 0
	}
	t.i += sort.SearchStrings(t.tokens[t.i:], token)
	return t.i < len(t.tokens)
}

type Tokens struct {
	tokens  []string
	indices map[string]int
}

func NewTokens() *Tokens {
	return &Tokens{tokens: make([]string, 0), indices: make(map[string]int)}
}

func NewFilledTokens(tokens []string, indices map[string]int) *Tokens {
	return &Tokens{tokens: tokens, indices: indices}
}

func (t *Tokens) Len() int {
	return len(t.tokens)
}

func (t *Tokens) Token(i int) string {
	return t.tokens[i]
}

func (t *Tokens) Tokens() TokenIterator {
	return newTokenIterator(t.tokens)
}

func (t *Tokens) Lookup(token string) (int, bool) {
	i, ok := t.indices[token]
	return i, ok
}

func (t *Tokens) LookupOrAdd(token string) int {
	if i, ok := t.indices[token]; ok {
		return i
	}
	t.tokens = append(t.tokens, token)
	i := len(t.tokens) - 1
	t.indices[token] = i
	return i
}

type byToken struct {
	index  SortableIndex
	tokens []string
}

func (t byToken) Len() int { return len(t.tokens) }

func (t byToken) Swap(i, j int) {
	t.tokens[i], t.tokens[j] = t.tokens[j], t.tokens[i]
	t.index.Swap(i, j)
}

func (t byToken) Less(i, j int) bool {
	return t.tokens[i] < t.tokens[j]
}

func (t *Tokens) Sort(index SortableIndex) {
	sort.Sort(byToken{index: index, tokens: t.tokens})
	for i, token := range t.tokens {
		t.indices[token] = i
	}
}

type tokenCount struct {
	prefix   string
	tokens   int
	examples []string
}

type tokenCounts []tokenCount

const maxTokenExamples = 4

func (b tokenCounts) Len() int      { return len(b) }
func (b tokenCounts) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b tokenCounts) Less(i, j int) bool {
	return b[i].prefix < b[j].prefix
}

func TokensToHTML(tokens TokenIterator) string {
	total := 0
	output := ""
	counts := make(map[string]tokenCount)
	for tokens.Next() {
		total++
		token := tokens.Token()
		if pos := strings.Index(token, ":"); pos > 0 {
			token = token[0 : pos+1]
		}
		var count tokenCount
		var ok bool
		if count, ok = counts[token]; ok {
			count = tokenCount{token, count.tokens + 1, count.examples}
		} else {
			count = tokenCount{token, 1, make([]string, 0, maxTokenExamples)}
		}
		if len(count.examples) < maxTokenExamples {
			// TODO: Take a random sample, instead of the first items?
			count.examples = append(count.examples, tokens.Token())
		}
		counts[token] = count
	}
	sorted := make(tokenCounts, 0, len(counts))
	for _, count := range counts {
		sorted = append(sorted, count)
	}
	sort.Sort(sorted)
	for _, count := range sorted {
		if count.tokens > 1 {
			output += fmt.Sprintf("  %s[%d tokens]\n", count.prefix, count.tokens)
			for _, example := range count.examples {
				output += "    " + example + "\n"
			}
		} else {
			output += fmt.Sprintf("  %s\n", count.prefix)
		}
	}
	return fmt.Sprintf("%d tokens:\n", total) + output
}
