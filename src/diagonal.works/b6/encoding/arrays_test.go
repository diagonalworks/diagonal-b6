package encoding

import (
	"fmt"
	"testing"
)

func TestByteArrays(t *testing.T) {
	namesByID, _ := loadGranarySquareForTests(t)

	names := make([]string, 0, len(namesByID))
	for _, name := range namesByID {
		names = append(names, name)
	}

	offset := Offset(42)
	builder := NewByteArraysBuilder(len(names))
	for i, name := range names {
		builder.Reserve(i, len(name))
	}

	var output Buffer
	builder.WriteHeader(&output, offset)
	for i, name := range names {
		if err := builder.WriteItem(&output, i, []byte(name)); err != nil {
			t.Fatalf("Expected no error from WriteItem(), found: %s", err)
		}
	}

	// Ensure the end offset really is beyond the map data by writing over it
	var zeros [1024]byte
	if _, err := output.WriteAt(zeros[0:], int64(offset+Offset(builder.Length()))); err != nil {
		t.Fatal("Failed to pad output")
	}

	b := NewByteArrays(output.Bytes()[offset:])
	for i, expected := range names {
		if item := b.Item(i); string(item) != string(expected) {
			t.Errorf("Expected %q, found %q at item %d", expected, string(item), i)
		}
	}
}

func TestByteArrayReservesAreCumulative(t *testing.T) {
	namesByID, _ := loadGranarySquareForTests(t)

	names := make([]string, 0, len(namesByID))
	ids := make([]string, 0, len(namesByID))
	for id, name := range namesByID {
		names = append(names, name)
		ids = append(ids, fmt.Sprintf("%d", id))
	}

	offset := Offset(42)
	builder := NewByteArraysBuilder(len(names))
	for i := 0; i < len(namesByID); i++ {
		builder.Reserve(i, len(names[i]))
		builder.Reserve(i, len(ids[i]))
	}

	var output Buffer
	builder.WriteHeader(&output, offset)
	for i := 0; i < len(namesByID); i++ {
		if err := builder.WriteItem(&output, i, []byte(names[i]), []byte(ids[i])); err != nil {
			t.Fatalf("Expected no error from WriteItem(), found: %s", err)
		}
	}

	b := NewByteArrays(output.Bytes()[offset:])
	for i := 0; i < len(namesByID); i++ {
		expected := names[i] + ids[i]
		if name := b.Item(i); string(name) != string(expected) {
			t.Errorf("Expected %q, found %q at item %d", expected, string(name), i)
		}
	}
}
