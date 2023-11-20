package ui

import (
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/test/camden"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestHistogramWithStrings(t *testing.T) {
	collection := b6.ArrayCollection[string, string]{
		Keys:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16"},
		Values: []string{"heath", "fold", "heath", "fold", "epping", "fold", "epping", "briki", "epping", "briki", "fold", "unfold", "heath", "fold", "epping", "home"},
	}

	histogram := api.HistogramCollection{UntypedCollection: collection.Collection()}
	response := &pb.UIResponseProto{}
	if err := fillResponseFromHistogram(response, &histogram, b6.EmptyWorld{}); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	total := int32(len(collection.Keys))
	expected := &pb.StackProto{
		Substacks: []*pb.SubstackProto{
			{
				Lines: []*pb.LineProto{
					{
						Line: &pb.LineProto_HistogramBar{
							HistogramBar: &pb.HistogramBarLineProto{
								Range: AtomFromString("fold"),
								Value: 5,
								Total: total,
								Index: 0,
							},
						},
					},
					{
						Line: &pb.LineProto_HistogramBar{
							HistogramBar: &pb.HistogramBarLineProto{
								Range: AtomFromString("epping"),
								Value: 4,
								Total: total,
								Index: 1,
							},
						},
					},
					{
						Line: &pb.LineProto_HistogramBar{
							HistogramBar: &pb.HistogramBarLineProto{
								Range: AtomFromString("heath"),
								Value: 3,
								Total: total,
								Index: 2,
							},
						},
					},
					{
						Line: &pb.LineProto_HistogramBar{
							HistogramBar: &pb.HistogramBarLineProto{
								Range: AtomFromString("briki"),
								Value: 2,
								Total: total,
								Index: 3,
							},
						},
					},
					{
						Line: &pb.LineProto_HistogramBar{
							HistogramBar: &pb.HistogramBarLineProto{
								Range: AtomFromString("other"),
								Value: 2,
								Total: total,
								Index: 4,
							},
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(expected, response.Stack, protocmp.Transform()); diff != "" {
		t.Errorf("Unexpected diff: %s", diff)
	}
}

func TestHistogramWithIntegers(t *testing.T) {
	collection := b6.ArrayCollection[string, int]{
		Keys:   []string{"1", "2", "3", "4", "5", "6", "7"},
		Values: []int{1, 1, 1, 1, 1, 1, 2},
	}

	histogram := api.HistogramCollection{UntypedCollection: collection.Collection()}
	response := &pb.UIResponseProto{}
	if err := fillResponseFromHistogram(response, &histogram, b6.EmptyWorld{}); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	total := int32(len(collection.Keys))
	expected := &pb.StackProto{
		Substacks: []*pb.SubstackProto{
			{
				Lines: []*pb.LineProto{
					{
						Line: &pb.LineProto_HistogramBar{
							HistogramBar: &pb.HistogramBarLineProto{
								Range: AtomFromString("1"),
								Value: 6,
								Total: total,
								Index: 0,
							},
						},
					},
					{
						Line: &pb.LineProto_HistogramBar{
							HistogramBar: &pb.HistogramBarLineProto{
								Range: AtomFromString("2"),
								Value: 1,
								Total: total,
								Index: 1,
							},
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(expected, response.Stack, protocmp.Transform()); diff != "" {
		t.Errorf("Unexpected diff: %s", diff)
	}
}

func TestHistogramWithIntegersAndMoreThan5Buckets(t *testing.T) {
	collection := b6.ArrayValuesCollection[int]{
		1, 1, 1, 1, 2, 2, 2, 3, 3, 4, 5, 6, 7,
	}

	histogram := api.HistogramCollection{UntypedCollection: collection.Collection()}
	response := &pb.UIResponseProto{}
	if err := fillResponseFromHistogram(response, &histogram, b6.EmptyWorld{}); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	total := int32(len(collection))
	expected := &pb.StackProto{
		Substacks: []*pb.SubstackProto{
			{
				Lines: []*pb.LineProto{
					{
						Line: &pb.LineProto_HistogramBar{
							HistogramBar: &pb.HistogramBarLineProto{
								Range: AtomFromString("1-2"),
								Value: 4,
								Total: total,
								Index: 0,
							},
						},
					},
					{
						Line: &pb.LineProto_HistogramBar{
							HistogramBar: &pb.HistogramBarLineProto{
								Range: AtomFromString("2-3"),
								Value: 3,
								Total: total,
								Index: 1,
							},
						},
					},
					{
						Line: &pb.LineProto_HistogramBar{
							HistogramBar: &pb.HistogramBarLineProto{
								Range: AtomFromString("3-4"),
								Value: 2,
								Total: total,
								Index: 2,
							},
						},
					},
					{
						Line: &pb.LineProto_HistogramBar{
							HistogramBar: &pb.HistogramBarLineProto{
								Range: AtomFromString("4-6"),
								Value: 2,
								Total: total,
								Index: 3,
							},
						},
					},
					{
						Line: &pb.LineProto_HistogramBar{
							HistogramBar: &pb.HistogramBarLineProto{
								Range: AtomFromString("6-"),
								Value: 2,
								Total: total,
								Index: 4,
							},
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(expected, response.Stack, protocmp.Transform()); diff != "" {
		t.Errorf("Unexpected diff: %s", diff)
	}
}

func TestHistogramWithFeatures(t *testing.T) {
	collection := b6.ArrayCollection[b6.FeatureID, string]{
		Keys:   []b6.FeatureID{camden.VermuteriaID.FeatureID(), camden.LightermanID.FeatureID(), camden.GranarySquareID.FeatureID()},
		Values: []string{"amenity", "amenity", "highway"},
	}

	histogram := api.HistogramCollection{UntypedCollection: collection.Collection()}
	response := &pb.UIResponseProto{}
	if err := fillResponseFromHistogram(response, &histogram, b6.EmptyWorld{}); err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	expected := []*pb.BucketedProto{{
		Buckets: []*pb.FeatureIDsProto{
			{
				Namespaces: []string{
					"/point/openstreetmap.org/node",
					"/area/openstreetmap.org/way",
				},
				Ids: []*pb.IDsProto{
					{Ids: []uint64{camden.VermuteriaID.Value}},
					{Ids: []uint64{camden.LightermanID.Value}},
				},
			},
			{
				Namespaces: []string{
					"/area/openstreetmap.org/way",
				},
				Ids: []*pb.IDsProto{
					{Ids: []uint64{camden.GranarySquareID.Value}},
				},
			},
		},
	}}

	if diff := cmp.Diff(expected, response.Bucketed, protocmp.Transform()); diff != "" {
		t.Errorf("Unexpected diff: %s", diff)
	}
}
