package b6

import (
	"reflect"
	"sort"
	"testing"

	"github.com/golang/geo/s2"
)

type testPoints struct {
	points []PointFeature
	i      int
}

func (t *testPoints) Feature() Feature {
	return t.points[t.i]
}

func (t *testPoints) FeatureID() FeatureID {
	return t.points[t.i].FeatureID()
}

func (t *testPoints) Next() bool {
	t.i++
	return t.i < len(t.points)
}

func TestMergedFeatures(t *testing.T) {
	p := s2.PointFromLatLng(s2.LatLngFromDegrees(51.5354872, -0.1253010))
	a := testPoints{
		points: []PointFeature{
			PointFromS2PointAndID(p, MakePointID(NamespaceOSMNode, 1447052072)),
			PointFromS2PointAndID(p, MakePointID(NamespaceOSMNode, 7555211491)),
		},
		i: -1,
	}
	b := testPoints{
		points: []PointFeature{
			PointFromS2PointAndID(p, MakePointID(NamespaceOSMNode, 29740928)),
			PointFromS2PointAndID(p, MakePointID(NamespaceOSMNode, 1237701871)),
			PointFromS2PointAndID(p, MakePointID(NamespaceOSMNode, 1447052072)),
			PointFromS2PointAndID(p, MakePointID(NamespaceOSMNode, 2517853770)),
		},
		i: -1,
	}

	seen := make(map[FeatureID]struct{})
	expected := make(FeatureIDs, 0)
	for _, points := range [][]PointFeature{a.points, b.points} {
		for _, point := range points {
			if _, ok := seen[point.FeatureID()]; !ok {
				expected = append(expected, point.FeatureID())
				seen[point.FeatureID()] = struct{}{}
			}
		}
	}
	sort.Sort(expected)

	actual := make(FeatureIDs, 0)
	merged := MergeFeatures(&a, &b)
	for merged.Next() {
		actual = append(actual, merged.FeatureID())
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %+v, found %+v", expected, actual)
	}
}
