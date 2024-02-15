package ingest

import (
	"bytes"
	"fmt"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/osm"
	"diagonal.works/b6/test"
	"github.com/golang/geo/s2"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestExportModificationsAsYAML(t *testing.T) {
	nodes, ways, relations, err := osm.ReadWholePBF(test.Data(test.GranarySquarePBF))
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	caravan := osm.Node{
		ID:       osm.NodeID(2300722786),
		Location: osm.LatLng{Lat: 51.5357237, Lng: -0.1253052},
		Tags:     []osm.Tag{{Key: "name", Value: "Caravan"}, {Key: "cuisine", Value: "coffee_shop"}},
	}
	nodes = append(nodes, caravan)

	dishoom := osm.Node{
		ID:       osm.NodeID(3501612811),
		Location: osm.LatLng{Lat: 51.536454, Lng: -0.126826},
		Tags:     []osm.Tag{{Key: "name", Value: "Dishoom"}},
	}
	nodes = append(nodes, caravan)

	base, err := BuildWorldFromOSM(nodes, ways, relations, &BuildOptions{Cores: 2})
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	m := NewMutableOverlayWorld(base)
	m.AddTag(FromOSMNodeID(caravan.ID).FeatureID(), b6.Tag{Key: "wheelchair", Value: "yes"})
	m.RemoveTag(FromOSMNodeID(caravan.ID).FeatureID(), "cuisine")
	m.AddTag(FromOSMNodeID(dishoom.ID).FeatureID(), b6.Tag{Key: "wheelchair", Value: "no"})

	ifo := NewPointFeature(FromOSMNodeID(osm.NodeID(3868276529)), s2.LatLngFromDegrees(51.5321749, -0.1250181))
	ifo.AddTag(b6.Tag{Key: "name", Value: "Identified Flying Object"})
	ifo.AddTag(b6.Tag{Key: "tourism", Value: "attraction"})
	if err := m.AddFeature(ifo); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	footway := NewPathFeature(3)
	footway.PathID = b6.MakePathID(b6.Namespace("diagonal.works/test"), 1)
	footway.SetPointID(0, FromOSMNodeID(caravan.ID))
	footway.SetLatLng(1, s2.LatLngFromDegrees(51.535632, -0.126046))
	footway.SetPointID(2, FromOSMNodeID(dishoom.ID))
	footway.AddTag(b6.Tag{Key: "highway", Value: "footway"})
	if err := m.AddFeature(footway); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	boundary := NewPathFeature(4)
	boundary.PathID = b6.MakePathID(b6.Namespace("diagonal.works/test"), 2)
	boundary.SetPointID(0, FromOSMNodeID(caravan.ID))
	boundary.SetPointID(1, FromOSMNodeID(dishoom.ID))
	boundary.SetLatLng(2, s2.LatLngFromDegrees(51.535632, -0.126046))
	boundary.SetPointID(3, FromOSMNodeID(caravan.ID))
	boundary.AddTag(b6.Tag{Key: "highway", Value: "footway"})
	if err := m.AddFeature(boundary); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	square := NewAreaFeature(1)
	square.AreaID = b6.MakeAreaID(b6.Namespace("diagonal.works/test"), 3)
	square.SetPathIDs(0, []b6.PathID{boundary.PathID})
	if err := m.AddFeature(square); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	ranking := NewRelationFeature(2)
	ranking.RelationID = b6.MakeRelationID(b6.Namespace("diagonal.works/test"), 4)
	ranking.Members = []b6.RelationMember{
		{ID: FromOSMNodeID(caravan.ID).FeatureID(), Role: "good"},
		{ID: FromOSMNodeID(dishoom.ID).FeatureID(), Role: "best"},
	}
	ranking.AddTag(b6.Tag{Key: "source", Value: "diagonal"})
	if err := m.AddFeature(ranking); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	var analysis CollectionFeature
	analysis.CollectionID = b6.MakeCollectionID(b6.Namespace("diagonal.works/test"), 5)
	analysis.Keys = []interface{}{FromOSMNodeID(caravan.ID).FeatureID(), FromOSMNodeID(dishoom.ID).FeatureID()}
	analysis.Values = []interface{}{"good", "best"}
	analysis.AddTag(b6.Tag{Key: "source", Value: "diagonal"})
	if err := m.AddFeature(&analysis); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	var expression ExpressionFeature
	expression.ExpressionID = b6.MakeExpressionID(b6.Namespace("diagonal.works/test"), 6)
	expression.Expression = b6.Expression{
		AnyExpression: &b6.CallExpression{
			Function: b6.NewSymbolExpression("find"),
			Args: []b6.Expression{
				{
					AnyExpression: &b6.QueryExpression{
						Query: b6.Intersection{
							b6.Tagged{Key: "#highway", Value: "cycleway"},
							b6.IntersectsFeature{
								ID: AreaIDFromOSMWayID(222021571).FeatureID(),
							},
						},
					},
					Name:  "bike paths",
					Begin: 36,
					End:   42,
				},
			},
		},
	}

	expression.AddTag(b6.Tag{Key: "source", Value: "diagonal"})
	if err := m.AddFeature(&expression); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	var buffer bytes.Buffer
	if err := ExportChangesAsYAML(m, &buffer); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	ingested := NewMutableOverlayWorld(base)
	change := IngestChangesFromYAML(&buffer)
	if _, err := change.Apply(ingested); err != nil {
		t.Fatalf("Expected no error from ingest, found: %s", err)
	}

	compared := map[b6.FeatureID]bool{
		ifo.FeatureID():        false,
		footway.FeatureID():    false,
		boundary.FeatureID():   false,
		square.FeatureID():     false,
		ranking.FeatureID():    false,
		analysis.FeatureID():   false,
		expression.FeatureID(): false,
	}

	compare := func(f b6.Feature, goroutine int) error {
		if diff := DiffFeatures(f, ingested.FindFeatureByID(f.FeatureID())); diff != "" {
			t.Error(diff)
		}
		compared[f.FeatureID()] = true
		return nil
	}
	if err := m.EachFeature(compare, &b6.EachFeatureOptions{}); err != nil {
		t.Errorf("Expected no error in comparison, found: %s", err)
	}

	for id, seen := range compared {
		if !seen {
			t.Errorf("Expected to compared %s", id)
		}
	}
}

func DiffFeatures(expected b6.Feature, actual b6.Feature) string {
	if actual == nil {
		return fmt.Sprintf("- %s\n+ nil", expected.FeatureID())
	}
	if expected.FeatureID().Type != actual.FeatureID().Type {
		return fmt.Sprintf("-  %s\n+ %s", expected.FeatureID().Type, actual.FeatureID().Type)
	}
	diffs := ""
	if e, ok := expected.(b6.RelationFeature); ok {
		a := actual.(b6.RelationFeature)
		if e.Len() != a.Len() {
			return fmt.Sprintf("- %d members\n+ %d members", e.Len(), a.Len())
		}
		for i := 0; i < a.Len(); i++ {
			if diff := cmp.Diff(e.Member(i), a.Member(i)); diff != "" {
				diffs += diff
			}
		}
	} else if e, ok := expected.(b6.CollectionFeature); ok {
		a := actual.(b6.CollectionFeature)

		ei := e.BeginUntyped()
		ai := a.BeginUntyped()
		for {
			eOk, err := ei.Next()
			if err != nil {
				return fmt.Sprintf("expected no error found %s", err.Error())
			}
			aOk, err := ai.Next()
			if err != nil {
				return fmt.Sprintf("expected no error found %s", err.Error())
			}

			if !eOk && !aOk {
				break
			}

			if !(eOk && aOk) {
				return "collection items differ in size"
			}

			if diff := cmp.Diff(ei.Key(), ai.Key()); diff != "" {
				diffs += diff
			}

			if diff := cmp.Diff(ei.Value(), ai.Value()); diff != "" {
				diffs += diff
			}
		}
	} else if e, ok := expected.(b6.PhysicalFeature); ok {
		a := actual.(b6.PhysicalFeature)
		coverer := s2.RegionCoverer{MaxLevel: 18, MaxCells: 10} // 18 implies 3cm accuracy
		diffs += cmp.Diff(e.Covering(coverer), a.Covering(coverer))
	} else if e, ok := expected.(b6.ExpressionFeature); ok {
		a := actual.(b6.ExpressionFeature)
		ae, _ := a.Expression().ToProto()
		ee, _ := e.Expression().ToProto()
		// TODO: implement a cmp.Diff transformer for expressions
		diffs += cmp.Diff(ee, ae, protocmp.Transform())
	}
	diffs += cmp.Diff(expected.AllTags(), actual.AllTags())
	return diffs
}
