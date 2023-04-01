package b6

import (
	"testing"
)

func TestPointIDFromValidGBPostcode(t *testing.T) {
	postcodes := []string{
		"N1C4AB",
		"N1C 4AB",
		"N 1C4AB",
		"n1c 4ab",
		"n1c4ab",
	}

	expected := PointIDFromGBPostcode(postcodes[0])
	if postcode, ok := PostcodeFromPointID(expected); !ok || postcode != postcodes[0] {
		t.Errorf("Expected %q, found %q", postcodes[0], postcode)
	}

	for i := 1; i < len(postcodes); i++ {
		id := PointIDFromGBPostcode(postcodes[i])
		if id != expected {
			t.Errorf("Expected %s, found %s for postcode %q", expected, id, postcodes[i])
		}
		if postcode, ok := PostcodeFromPointID(expected); !ok || postcode != postcodes[0] {
			t.Errorf("Expected %q, found %q", postcodes[0], postcode)
		}
	}
}

func TestPointIDFromInvalidGBPostcode(t *testing.T) {
	postcodes := []string{
		"N1CZ 4ABZ",
		"N1C 4!B",
	}

	for _, postcode := range postcodes {
		id := PointIDFromGBPostcode(postcode)
		if id != PointIDInvalid {
			t.Errorf("Expected invalid id for %q, found %s", postcode, id)
		}
	}

	ids := []PointID{
		PointIDInvalid,
		MakePointID(NamespaceOSMNode, 3501612811),
	}
	for _, id := range ids {
		if _, ok := PostcodeFromPointID(id); ok {
			t.Errorf("Expected false when converting %s", id)
		}
	}
}

func TestFeatureIDFromGBONSCode(t *testing.T) {
	codes := []string{
		"E01000001", // City of London 001A
		"E01000869", // Camden 015E
	}

	year := 2011
	for _, code := range codes {
		id := FeatureIDFromGBONSCode(code, year, FeatureTypeArea)
		if c, y, ok := GBONSCodeFromFeatureID(id); !ok || c != code || y != year {
			t.Errorf("Expected %s, year %d, found %s, year %d", code, year, c, y)
		}
	}
}
