package b6

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

func TestExportCollectionExpressionAsYAML(t *testing.T) {
	c := ArrayCollection[FeatureID, string]{
		Keys: []FeatureID{
			MakePointID(NamespaceOSMNode, 2300722786).FeatureID(),
			MakePointID(NamespaceOSMNode, 3501612811).FeatureID(),
		},
		Values: []string{"good", "best"},
	}
	e := Expression{AnyExpression: &CollectionExpression{UntypedCollection: c.Collection()}}
	m, err := yaml.Marshal(e)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	var ee Expression
	if err := yaml.Unmarshal(m, &ee); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	if collection, ok := ee.AnyExpression.(*CollectionExpression); ok {
		cc := AdaptCollection[FeatureID, string](collection.UntypedCollection)
		if keys, err := cc.AllKeys(nil); err == nil {
			if diff := cmp.Diff(c.Keys, keys); diff != "" {
				t.Errorf("Differing keys (-want, +got):\n%s", diff)
			}
		} else {
			t.Errorf("Failed to fill keys: %s", err)
		}
		if values, err := cc.AllValues(nil); err == nil {
			if diff := cmp.Diff(c.Values, values); diff != "" {
				t.Errorf("Differing values (-want, +got):\n%s", diff)
			}
		} else {
			t.Errorf("Failed to fill keys: %s", err)
		}
	} else {
		t.Errorf("Expected a CollectionExpression, found %T", e.AnyExpression)
	}
}
