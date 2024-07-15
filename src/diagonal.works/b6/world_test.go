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
