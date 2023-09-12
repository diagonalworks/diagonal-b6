package api

import (
	"testing"
)

func TestGreaterHappyPath(t *testing.T) {
	cases := []struct {
		a        interface{}
		b        interface{}
		expected bool
	}{
		{3, 0, true},
		{0, 3, false},
		{3, int32(0), true},
		{int32(0), 3, false},
		{uint64((1 << 33) + 1), 0, true},
		{0, uint64((1 << 33) + 1), false},
		{uint64((1 << 33) + 2), (1 << 33) + 1, true},
		{(1 << 33) + 1, uint64((1 << 33) + 2), false},
	}
	for _, c := range cases {
		greater, err := Greater(c.a, c.b)
		if err != nil {
			t.Errorf("Expected no error, found: %s", err)
		} else if greater != c.expected {
			t.Errorf("Expected %v > %v == %v, found %v", c.a, c.b, c.expected, greater)
		}
	}
}
