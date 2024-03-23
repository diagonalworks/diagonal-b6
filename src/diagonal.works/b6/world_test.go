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
		{Tag{Key: "#amenity", Value: String("restaurant")}, "#amenity=restaurant"},
		{Tag{Key: "note", Value: String("Only on match days")}, `note="Only on match days"`},
		{Tag{Key: "note", Value: String("Value with a \" in the middle")}, `note="Value with a \" in the middle"`},
		{Tag{Key: "note", Value: String("Value with a \\ in the middle")}, `note="Value with a \\ in the middle"`},
		{Tag{Key: `Key with = in middle`, Value: String("Value with a \\ in the middle")}, `"Key with = in middle"="Value with a \\ in the middle"`},
	}
	for _, c := range cases {
		if s := c.tag.String(); s != c.s {
			t.Errorf("Expected %s, found %s", c.s, s)
		}
		var tag Tag
		tag.FromString(c.s)
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
		{`#amenity="restaurant"nonsense`, Tag{Key: "#amenity", Value: String("restaurant")}},
		{`#amenity    ="restaurant"nonsense`, Tag{Key: "#amenity", Value: String("restaurant")}},
		{`#amenity restaurant`, Tag{Key: "#amenityrestaurant", Value: String("")}},
	}
	for _, c := range cases {
		var tag Tag
		tag.FromString(c.s)
		if tag.Key != c.tag.Key {
			t.Errorf("Expected key %s, found %s", c.tag.Key, tag.Key)
		}
		if tag.Value != c.tag.Value {
			t.Errorf("Expected value %s, found %s", c.tag.Value, tag.Value)
		}
	}
}
