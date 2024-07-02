package b6

import (
	"testing"

	pb "diagonal.works/b6/proto"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

func TestExportCollectionExpressionAsYAML(t *testing.T) {
	c := ArrayCollection[FeatureID, string]{
		Keys: []FeatureID{
			FeatureID{FeatureTypePoint, NamespaceOSMNode, 2300722786},
			FeatureID{FeatureTypePoint, NamespaceOSMNode, 3501612811},
		},
		Values: []string{"good", "best"},
	}
	e := Expression{AnyExpression: CollectionExpression{UntypedCollection: c.Collection()}}
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

func TestIntLiteralFromProto(t *testing.T) {
	const speed = 4

	p := &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_IntValue{
					IntValue: speed,
				},
			},
		},
	}

	l, err := LiteralFromProto(p)
	if err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

	if i, ok := l.AnyLiteral.(IntExpression); ok {
		if i != speed {
			t.Fatalf("expected %d, found %d", speed, int(i))
		}
	} else {
		t.Fatalf("expected an int, found %T", l.AnyLiteral)
	}
}

func TestNameAndTokenPositionsFromProto(t *testing.T) {
	p := &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_IntValue{
					IntValue: 4,
				},
			},
		},
		Name:  "speed",
		Begin: 2,
		End:   3,
	}

	e, err := ExpressionFromProto(p)
	if err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

	if e.Name != "speed" {
		t.Errorf("unexpected name: %s", e.Name)
	}
	if e.Begin != 2 {
		t.Errorf("unexpected begin: %d", e.Begin)
	}
	if e.End != 3 {
		t.Errorf("unexpected end: %d", e.End)
	}
}
