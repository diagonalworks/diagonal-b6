package renderer

import (
	"testing"
)

func TestColourFromHexString(t *testing.T) {
	expected := "#d3d6fd"
	if found := ColourFromHexString(expected).ToHexString(); found != expected {
		t.Errorf("Expected %q, found %q", expected, found)
	}

	invalid := "invalid"
	if found := ColourFromHexString(invalid).ToHexString(); found != "#000000" {
		t.Errorf("Unexpected invalid colour: %q", found)
	}
}

func TestGradient(t *testing.T) {
	gradient := Gradient{
		{Value: 0.0, Colour: ColourFromHexString("#d3d6fd")},
		{Value: 0.30, Colour: ColourFromHexString("#fca364")},
		{Value: 0.60, Colour: ColourFromHexString("#f88a4f")},
		{Value: 1.00, Colour: ColourFromHexString("#f96c53")},
	}
	expected := "#f99256"
	if found := gradient.Interpolate(0.5).ToHexString(); found != expected {
		t.Errorf("Expected %q, found %q", expected, found)
	}

	expected = "#d3d6fd"
	if found := gradient.Interpolate(-1.0).ToHexString(); found != expected {
		t.Errorf("Expected %q, found %q", expected, found)
	}

	expected = "#f96c53"
	if found := gradient.Interpolate(2.0).ToHexString(); found != expected {
		t.Errorf("Expected %q, found %q", expected, found)
	}
}
