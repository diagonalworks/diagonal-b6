package b6

import (
	"reflect"
	"sort"
	"testing"
)

type testFeature FeatureID

func (t testFeature) FeatureID() FeatureID {
	return FeatureID(t)
}

func (t testFeature) AllTags() []Tag {
	return []Tag{}
}

func (t testFeature) Get(string) Tag {
	return Tag{}
}

func TestMergedFeatures(t *testing.T) {
	a := []Feature{
		testFeature(FeatureID{FeatureTypePoint, NamespaceOSMNode, 1447052072}),
		testFeature(FeatureID{FeatureTypePoint, NamespaceOSMNode, 7555211491}),
	}

	b := []Feature{
		testFeature(FeatureID{FeatureTypePoint, NamespaceOSMNode, 29740928}),
		testFeature(FeatureID{FeatureTypePoint, NamespaceOSMNode, 1237701871}),
		testFeature(FeatureID{FeatureTypePoint, NamespaceOSMNode, 1447052072}),
		testFeature(FeatureID{FeatureTypePoint, NamespaceOSMNode, 2517853770}),
	}

	seen := make(map[FeatureID]struct{})
	expected := make(FeatureIDs, 0)
	for _, features := range [][]Feature{a, b} {
		for _, feature := range features {
			if _, ok := seen[feature.FeatureID()]; !ok {
				expected = append(expected, feature.FeatureID())
				seen[feature.FeatureID()] = struct{}{}
			}
		}
	}
	sort.Sort(expected)

	actual := make(FeatureIDs, 0)
	merged := MergeFeatures(NewFeatureIterator(a), NewFeatureIterator(b))
	for merged.Next() {
		actual = append(actual, merged.FeatureID())
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %+v, found %+v", expected, actual)
	}
}
