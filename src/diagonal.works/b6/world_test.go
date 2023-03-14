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
