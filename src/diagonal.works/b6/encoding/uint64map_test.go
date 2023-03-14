package encoding

import (
	"fmt"
	"sync"
	"testing"

	"diagonal.works/b6/test/camden"
)

const (
	TestNameTag    Tag = 0
	TestAmenityTag     = 1
)

func TestUint64Map(t *testing.T) {
	names, amenities := loadGranarySquareForTests(t)
	if names == nil || amenities == nil {
		return
	}

	const NameTag = 0
	const AmenityTag = 1

	offset := Offset(42)
	builder := NewUint64MapBuilder(10, 1)
	ids := make(map[uint64]struct{})
	for id, name := range names {
		ids[uint64(id)] = struct{}{}
		builder.Reserve(uint64(id), TestNameTag, len(name))
	}
	for id, amenity := range amenities {
		ids[uint64(id)] = struct{}{}
		builder.Reserve(uint64(id), TestAmenityTag, len(amenity))
	}

	var output Buffer
	builder.WriteHeader(&output, offset)
	for id, name := range names {
		if err := builder.WriteItem(uint64(id), NameTag, []byte(name), &output); err != nil {
			t.Errorf("Expected no error from WriteItem(), found: %s", err)
			return
		}
	}
	for id, amenity := range amenities {
		if err := builder.WriteItem(uint64(id), AmenityTag, []byte(amenity), &output); err != nil {
			t.Errorf("Expected no error from WriteItem(), found: %s", err)
			return
		}
	}

	// Ensure the end offset really is beyond the map data by writing over it
	var zeros [1024]byte
	if _, err := output.WriteAt(zeros[0:], int64(offset+Offset(builder.Length()))); err != nil {
		t.Errorf("Failed to pad output")
		return
	}

	tests := []struct {
		name string
		f    func(m *Uint64Map, ids map[uint64]struct{}, t *testing.T)
	}{
		{"FillTagged", ValidateFillTagged},
		{"EachItem", ValidateEachItem},
		{"EachItemEachItemWithError", ValidateEachItemEachItemWithError},
		{"Iterator", ValidateIterator},
		{"FindFirst", ValidateFindFirst},
	}

	m := NewUint64Map(output.Bytes()[offset:])
	if m.Length() != builder.Length() {
		t.Errorf("Expected map and builder lengths to be identical, found %d vs %d", m.Length(), builder.Length())
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) { test.f(m, ids, t) })
	}
}

func ValidateFillTagged(m *Uint64Map, ids map[uint64]struct{}, t *testing.T) {
	tagged := make([]Tagged, 0, 1)
	tagged = m.FillTagged(uint64(camden.VermuteriaNode), tagged)
	if len(tagged) != 2 {
		t.Errorf("Expected 2 entries, found %d", len(tagged))
	} else {
		if tagged[0].Tag > tagged[1].Tag {
			tagged[0], tagged[1] = tagged[1], tagged[0]
		}
		expected := []struct {
			tag   Tag
			value string
		}{
			{TestNameTag, "Vermuteria"},
			{TestAmenityTag, "cafe"},
		}
		for i, e := range expected {
			if tagged[i].Tag != e.tag || string(tagged[i].Data) != e.value {
				t.Errorf("Unexpected tag value %q with tag %d", string(tagged[i].Data), tagged[i].Tag)
			}
		}
	}
}

func ValidateEachItem(m *Uint64Map, ids map[uint64]struct{}, t *testing.T) {
	calls := 0
	var tags [2]int
	seen := make(map[uint64]struct{})
	found := make(map[uint64]string)
	var lock sync.Mutex
	f := func(id uint64, tagged []Tagged, _ int) error {
		lock.Lock()
		seen[id] = struct{}{}
		calls++
		for _, t := range tagged {
			tags[t.Tag]++
			if t.Tag == TestNameTag {
				found[id] = string(t.Data)
			}
		}
		lock.Unlock()

		return nil
	}
	m.EachItem(f, 2)

	if calls != len(ids) {
		t.Errorf("Expected one call per item, found %d calls for %d items", calls, len(ids))
	}
	if tags[0] == 0 || tags[1] == 0 {
		t.Errorf("Expected values for each tag, found: %+v", tags)
	}
	for id := range ids {
		if _, ok := seen[id]; !ok {
			t.Errorf("Expected to see id %d", id)
		}
	}
	if name, ok := found[uint64(camden.VermuteriaNode)]; !ok || name != "Vermuteria" {
		t.Errorf("Expected to iterate past Vermuteria")
	}
}

func ValidateEachItemEachItemWithError(m *Uint64Map, ids map[uint64]struct{}, t *testing.T) {
	cause := fmt.Errorf("Error from callback")
	f := func(id uint64, tagged []Tagged, _ int) error {
		for _, t := range tagged {
			if string(t.Data) == "Vermuteria" {
				return cause
			}
		}
		return nil
	}
	if err := m.EachItem(f, 2); err != cause {
		t.Errorf("Expected error %q, found: %s", cause, err)
	}
}

func ValidateIterator(m *Uint64Map, ids map[uint64]struct{}, t *testing.T) {
	found := make(map[uint64]string)
	i := m.Begin()
	for i.Next() {
		found[i.ID()] = string(i.Data(0))
	}

	if name, ok := found[uint64(camden.VermuteriaNode)]; !ok || name != "Vermuteria" {
		t.Errorf("Expected to iterate past Vermuteria")
	}
}

func ValidateFindFirst(m *Uint64Map, ids map[uint64]struct{}, t *testing.T) {
	v := m.FindFirstWithTag(uint64(camden.VermuteriaNode), TestNameTag)
	if v == nil || string(v) != "Vermuteria" {
		t.Errorf("Expected to lookup name for Vermuteria, found %s", v)
	}

	tag, ok := m.FindFirst(uint64(camden.VermuteriaNode))
	if !ok {
		t.Errorf("Expected to find name or amenity for Vermuteria")
	} else if (tag.Tag == TestNameTag && string(tag.Data) != "Vermuteria") || (tag.Tag == TestAmenityTag && string(tag.Data) != "cafe") {
		t.Errorf("Expected to find name or amenity for Vermuteria, found %s", v)
	}
}
