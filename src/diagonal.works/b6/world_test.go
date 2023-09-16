package b6

import (
	"testing"
)

func TestFeatureIDToAndFromString(t *testing.T) {
	id := MakePathID(NamespaceOSMWay, 687471322).FeatureID()
	actual := FeatureIDFromString(id.String())
	if actual != id {
		t.Errorf("Expected %s, found %s", id, actual)
	}
}

func TestFeatureFromStringHandlesLeadingSlash(t *testing.T) {
	token := "/path/openstreetmap.org/way/687471322"
	expected := MakePathID(NamespaceOSMWay, 687471322).FeatureID()
	actual := FeatureIDFromString(token)
	if actual != expected {
		t.Errorf("Expected %s, found: %s", expected, actual)
	}
}

func TestTagToAndFromStringHappyPath(t *testing.T) {
	cases := []struct {
		tag Tag
		s   string
	}{
		{Tag{Key: "#amenity", Value: "restaurant"}, "#amenity=restaurant"},
		{Tag{Key: "note", Value: "Only on match days"}, `note="Only on match days"`},
		{Tag{Key: "note", Value: "Value with a \" in the middle"}, `note="Value with a \" in the middle"`},
		{Tag{Key: "note", Value: "Value with a \\ in the middle"}, `note="Value with a \\ in the middle"`},
		{Tag{Key: `Key with = in middle`, Value: "Value with a \\ in the middle"}, `"Key with = in middle"="Value with a \\ in the middle"`},
	}
	for _, c := range cases {
		if s := c.tag.String(); s != c.s {
			t.Errorf("Expected %s, found %s", c.s, s)
		}
		tag := TagFromString(c.s)
		if tag.Key != c.tag.Key {
			t.Errorf("Expected key %s, found %s", c.tag.Key, tag.Key)
		}
		if tag.Value != c.tag.Value {
			t.Errorf("Expected value %s, found %s", c.tag.Value, tag.Value)
		}
	}
}

func TestTagToAndFromStringBrokenStrings(t *testing.T) {
	cases := []struct {
		s   string
		tag Tag
	}{
		{`#amenity="restaurant"nonsense`, Tag{Key: "#amenity", Value: "restaurant"}},
		{`#amenity    ="restaurant"nonsense`, Tag{Key: "#amenity", Value: "restaurant"}},
		{`#amenity restaurant`, Tag{Key: "#amenityrestaurant", Value: ""}},
	}
	for _, c := range cases {
		tag := TagFromString(c.s)
		if tag.Key != c.tag.Key {
			t.Errorf("Expected key %s, found %s", c.tag.Key, tag.Key)
		}
		if tag.Value != c.tag.Value {
			t.Errorf("Expected value %s, found %s", c.tag.Value, tag.Value)
		}
	}
}
