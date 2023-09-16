package encoding

import (
	"fmt"
	"testing"
)

func TestHashString(t *testing.T) {
	hashes := make(map[uint64]struct{})
	n := 1024
	for i := 0; i < n; i++ {
		token := fmt.Sprintf("building:levels=%d", i)
		hashes[HashString(token)] = struct{}{}
	}
	if len(hashes) != n {
		t.Errorf("Expected no collisons")
	}
}

func TestBuildStringTable(t *testing.T) {
	names, amenities := loadGranarySquareForTests(t)

	start := Offset(42)
	builder := NewStringTableBuilder()
	for _, name := range names {
		builder.Add(name)
	}
	for _, amenity := range amenities {
		builder.Add(amenity)
	}

	var output Buffer
	offset, err := builder.Write(&output, start)
	if err != nil {
		t.Errorf("Expected no error when writing string table, found: %s", err)
	}

	if builder.Lookup("bench") > builder.Lookup("cafe") || builder.Lookup("cafe") > builder.Lookup("Vermuteria") {
		t.Error("Expected the most frequent strings to have the smallest IDs")
	}

	// Ensure the end offset really is beyond the string data by writing over it
	var zeros [1024]byte
	if _, err := output.WriteAt(zeros[0:], int64(offset+Offset(builder.Length()))); err != nil {
		t.Error("Failed to pad output")
	}

	strings := NewStringTable(output.Bytes()[start:])
	for _, str := range []string{"bench", "cafe", "Vermuteria"} {
		if found := strings.Lookup(builder.Lookup(str)); found != str {
			t.Errorf("Expected to find %q, found %q", str, found)
		}
	}

	if int(offset-start) != builder.Length() {
		t.Errorf("Expected builder length to equal offset difference (%d vs %d)", offset-start, builder.Length())
	}
}
