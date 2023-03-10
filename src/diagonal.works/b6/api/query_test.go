package api

import (
	"testing"

	"diagonal.works/b6"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/test"
	"diagonal.works/b6/test/camden"
)

func TestMatchKeyQueryProto(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	lighterman := test.FindFeatureByID(camden.LightermanID.FeatureID(), granarySquare, t)
	if lighterman == nil {
		return
	}

	query := &pb.QueryProto{
		Query: &pb.QueryProto_Key{
			Key: &pb.KeyQueryProto{
				Key: "#building",
			},
		},
	}
	if !Matches(lighterman, query, granarySquare) {
		t.Errorf("Expected %v to match %s", lighterman.AllTags(), query)
	}

	query = &pb.QueryProto{
		Query: &pb.QueryProto_Key{
			Key: &pb.KeyQueryProto{
				Key: "#landuse",
			},
		},
	}
	if Matches(lighterman, query, granarySquare) {
		t.Errorf("Didn't expect %v to match %s", lighterman.AllTags(), query)
	}
}

func TestMatchKeyValueQueryProto(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	lighterman := test.FindFeatureByID(camden.LightermanID.FeatureID(), granarySquare, t)
	if lighterman == nil {
		return
	}

	query := &pb.QueryProto{
		Query: &pb.QueryProto_KeyValue{
			KeyValue: &pb.KeyValueQueryProto{
				Key:   "#building",
				Value: "yes",
			},
		},
	}
	if !Matches(lighterman, query, granarySquare) {
		t.Errorf("Expected %v to match %s", lighterman.AllTags(), query)
	}

	query = &pb.QueryProto{
		Query: &pb.QueryProto_KeyValue{
			KeyValue: &pb.KeyValueQueryProto{
				Key:   "#building",
				Value: "university",
			},
		},
	}
	if Matches(lighterman, query, granarySquare) {
		t.Errorf("Didn't expect %v to match %s", lighterman.AllTags(), query)
	}
}

func TestMatchIntersectionQueryProto(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	lighterman := test.FindFeatureByID(camden.LightermanID.FeatureID(), granarySquare, t)
	if lighterman == nil {
		return
	}

	query := &pb.QueryProto{
		Query: &pb.QueryProto_Intersection{
			Intersection: &pb.IntersectionQueryProto{
				Queries: []*pb.QueryProto{
					{
						Query: &pb.QueryProto_KeyValue{
							KeyValue: &pb.KeyValueQueryProto{
								Key:   "#building",
								Value: "yes",
							},
						},
					},
					{
						Query: &pb.QueryProto_KeyValue{
							KeyValue: &pb.KeyValueQueryProto{
								Key:   "name",
								Value: "The Lighterman",
							},
						},
					},
				},
			},
		},
	}
	if !Matches(lighterman, query, granarySquare) {
		t.Errorf("Expected %v to match %s", lighterman.AllTags(), query)
	}

	query = &pb.QueryProto{
		Query: &pb.QueryProto_Intersection{
			Intersection: &pb.IntersectionQueryProto{
				Queries: []*pb.QueryProto{
					{
						Query: &pb.QueryProto_KeyValue{
							KeyValue: &pb.KeyValueQueryProto{
								Key:   "#building",
								Value: "yes",
							},
						},
					},
					{
						Query: &pb.QueryProto_KeyValue{
							KeyValue: &pb.KeyValueQueryProto{
								Key:   "name",
								Value: "Caravan",
							},
						},
					},
				},
			},
		},
	}
	if Matches(lighterman, query, granarySquare) {
		t.Errorf("Didn't expect %v to match %s", lighterman.AllTags(), query)
	}
}

func TestMatchUnionQueryProto(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	lighterman := test.FindFeatureByID(camden.LightermanID.FeatureID(), granarySquare, t)
	if lighterman == nil {
		return
	}

	query := &pb.QueryProto{
		Query: &pb.QueryProto_Union{
			Union: &pb.UnionQueryProto{
				Queries: []*pb.QueryProto{
					{
						Query: &pb.QueryProto_KeyValue{
							KeyValue: &pb.KeyValueQueryProto{
								Key:   "#building",
								Value: "yes",
							},
						},
					},
					{
						Query: &pb.QueryProto_KeyValue{
							KeyValue: &pb.KeyValueQueryProto{
								Key:   "#building",
								Value: "university",
							},
						},
					},
				},
			},
		},
	}
	if !Matches(lighterman, query, granarySquare) {
		t.Errorf("Expected %v to match %s", lighterman.AllTags(), query)
	}

	query = &pb.QueryProto{
		Query: &pb.QueryProto_Intersection{
			Intersection: &pb.IntersectionQueryProto{
				Queries: []*pb.QueryProto{
					{
						Query: &pb.QueryProto_KeyValue{
							KeyValue: &pb.KeyValueQueryProto{
								Key:   "#landuse",
								Value: "grass",
							},
						},
					},
					{
						Query: &pb.QueryProto_KeyValue{
							KeyValue: &pb.KeyValueQueryProto{
								Key:   "#building",
								Value: "university",
							},
						},
					},
				},
			},
		},
	}
	if Matches(lighterman, query, granarySquare) {
		t.Errorf("Didn't expect %v to match %s", lighterman.AllTags(), query)
	}
}

func TestMatchAllQueryProto(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	lighterman := test.FindFeatureByID(camden.LightermanID.FeatureID(), granarySquare, t)
	if lighterman == nil {
		return
	}

	query := &pb.QueryProto{
		Query: &pb.QueryProto_All{
			All: &pb.AllQueryProto{},
		},
	}
	if !Matches(lighterman, query, granarySquare) {
		t.Errorf("Expected %v to match %s", lighterman.AllTags(), query)
	}
}

func TestMatchTypeQueryProto(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	lighterman := test.FindFeatureByID(camden.LightermanID.FeatureID(), granarySquare, t)
	if lighterman == nil {
		return
	}

	query := &pb.QueryProto{
		Query: &pb.QueryProto_Type{
			Type: &pb.TypeQueryProto{
				Type: pb.FeatureType_FeatureTypeArea,
				Query: &pb.QueryProto{
					Query: &pb.QueryProto_All{
						All: &pb.AllQueryProto{},
					},
				},
			},
		},
	}
	if !Matches(lighterman, query, granarySquare) {
		t.Errorf("Expected %v to match %s", lighterman.AllTags(), query)
	}

	query = &pb.QueryProto{
		Query: &pb.QueryProto_Type{
			Type: &pb.TypeQueryProto{
				Type: pb.FeatureType_FeatureTypePath,
				Query: &pb.QueryProto{
					Query: &pb.QueryProto_All{
						All: &pb.AllQueryProto{},
					},
				},
			},
		},
	}
	if Matches(lighterman, query, granarySquare) {
		t.Errorf("Didn't expect %v to match %s", lighterman.AllTags(), query)
	}
}

func TestMatchSpatialQueryProto(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	lighterman := test.FindFeatureByID(camden.LightermanID.FeatureID(), granarySquare, t)
	if lighterman == nil {
		return
	}

	query := &pb.QueryProto{
		Query: &pb.QueryProto_Spatial{
			Spatial: &pb.SpatialQueryProto{
				Area: &pb.AreaProto{
					Area: &pb.AreaProto_Cap{
						Cap: &pb.CapProto{
							Center: &pb.PointProto{
								LatE7: 515352700,
								LngE7: -1243500,
							},
							RadiusMeters: 100.0,
						},
					},
				},
			},
		},
	}
	if !Matches(lighterman, query, granarySquare) {
		t.Errorf("Expected %v to match %s", lighterman.AllTags(), query)
	}

	query = &pb.QueryProto{
		Query: &pb.QueryProto_Spatial{
			Spatial: &pb.SpatialQueryProto{
				Area: &pb.AreaProto{
					Area: &pb.AreaProto_Cap{
						Cap: &pb.CapProto{
							Center: &pb.PointProto{
								LatE7: 515365100,
								LngE7: -1273400,
							},
							RadiusMeters: 100.0,
						},
					},
				},
			},
		},
	}
	if Matches(lighterman, query, granarySquare) {
		t.Errorf("Didn't expect %v to match %s", lighterman.AllTags(), query)
	}
}

func TestMatchSpatialQueryProtoWithAreaFeature(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	bikeParking := test.FindFeatureByID(camden.GranarySquareBikeParkingID.FeatureID(), granarySquare, t)
	if bikeParking == nil {
		return
	}

	query := &pb.QueryProto{
		Query: &pb.QueryProto_Spatial{
			Spatial: &pb.SpatialQueryProto{
				Area: &pb.AreaProto{
					Area: &pb.AreaProto_Id{
						Id: &pb.FeatureIDProto{
							Type:      pb.FeatureType_FeatureTypeArea,
							Namespace: b6.NamespaceOSMWay.String(),
							Value:     uint64(camden.GranarySquareWay),
						},
					},
				},
			},
		},
	}
	if !Matches(bikeParking, query, granarySquare) {
		t.Errorf("Expected %v to match %s", bikeParking.AllTags(), query)
	}
}
