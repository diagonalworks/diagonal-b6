package b6

import (
	"testing"
)

func TestFeatureIDToAndFromString(t *testing.T) {
	id := FeatureID{FeatureTypePath, NamespaceOSMWay, 687471322}
	actual := FeatureIDFromString(id.String())
	if actual != id {
		t.Errorf("Expected %s, found %s", id, actual)
	}
}

func TestFeatureFromStringHandlesLeadingSlash(t *testing.T) {
	token := "/path/openstreetmap.org/way/687471322"
	expected := FeatureID{FeatureTypePath, NamespaceOSMWay, 687471322}
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
		{Tag{Key: "#amenity", Value: NewStringExpression("restaurant")}, "#amenity=restaurant"},
		{Tag{Key: "note", Value: NewStringExpression("Only on match days")}, `note="Only on match days"`},
		{Tag{Key: "note", Value: NewStringExpression("Value with a \" in the middle")}, `note="Value with a \" in the middle"`},
		{Tag{Key: "note", Value: NewStringExpression("Value with a \\ in the middle")}, `note="Value with a \\ in the middle"`},
		{Tag{Key: `Key with = in middle`, Value: NewStringExpression("Value with a \\ in the middle")}, `"Key with = in middle"="Value with a \\ in the middle"`},
	}
	for _, c := range cases {
		if s := c.tag.String(); s != c.s {
			t.Errorf("Expected %s, found %s", c.s, s)
		}
		var tag Tag
		tag.FromString(c.s)

		if !TagExpression(tag).Equal(TagExpression(c.tag)) {
			t.Errorf("Expected %s, found %s", c.tag.String(), tag.String())
		}
	}
}

func TestTagToAndFromStringBrokenStrings(t *testing.T) {
	cases := []struct {
		s   string
		tag Tag
	}{
		{`#amenity="restaurant"nonsense`, Tag{Key: "#amenity", Value: NewStringExpression("restaurant")}},
		{`#amenity    ="restaurant"nonsense`, Tag{Key: "#amenity", Value: NewStringExpression("restaurant")}},
		{`#amenity restaurant`, Tag{Key: "#amenityrestaurant", Value: NewStringExpression("")}},
	}
	for _, c := range cases {
		var tag Tag
		tag.FromString(c.s)
		if !TagExpression(tag).Equal(TagExpression(c.tag)) {
			t.Errorf("Expected %s, found %s", c.tag.String(), tag.String())
		}
	}
}
