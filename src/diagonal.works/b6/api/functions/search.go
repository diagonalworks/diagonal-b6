package functions

import (
	"diagonal.works/b6"
	"diagonal.works/b6/api"
	pb "diagonal.works/b6/proto"

	"github.com/golang/geo/s2"
)

type searchFeatureCollection struct {
	query b6.Query
	w     b6.World
	i     b6.Features
}

func (s *searchFeatureCollection) Begin() b6.Iterator[b6.FeatureID, b6.Feature] {
	return &searchFeatureCollection{
		query: s.query,
		w:     s.w,
	}
}

func (s *searchFeatureCollection) Key() b6.FeatureID {
	return s.i.FeatureID()
}

func (s *searchFeatureCollection) Value() b6.Feature {
	return s.i.Feature()
}

func (s *searchFeatureCollection) Next() (bool, error) {
	if s.i == nil {
		s.i = s.w.FindFeatures(s.query)
	}
	return s.i.Next(), nil
}

func (s *searchFeatureCollection) Count() (int, bool) {
	n := 0
	i := s.w.FindFeatures(s.query)
	for i.Next() {
		n++
	}
	return n, true
}

var _ b6.AnyCollection[b6.FeatureID, b6.Feature] = &searchFeatureCollection{}

func find(context *api.Context, query b6.Query) (b6.Collection[b6.FeatureID, b6.Feature], error) {
	return b6.Collection[b6.FeatureID, b6.Feature]{
		AnyCollection: &searchFeatureCollection{query: query, w: context.World},
	}, nil
}

func findPointFeatures(context *api.Context, query b6.Query) (b6.Collection[b6.FeatureID, b6.PointFeature], error) {
	tq := b6.Typed{Type: b6.FeatureTypePoint, Query: query}
	c := b6.Collection[b6.FeatureID, b6.Feature]{
		AnyCollection: &searchFeatureCollection{query: tq, w: context.World},
	}
	return b6.AdaptCollection[b6.FeatureID, b6.PointFeature](c), nil
}

func findPathFeatures(context *api.Context, query b6.Query) (b6.Collection[b6.FeatureID, b6.PathFeature], error) {
	tq := b6.Typed{Type: b6.FeatureTypePath, Query: query}
	c := b6.Collection[b6.FeatureID, b6.Feature]{
		AnyCollection: &searchFeatureCollection{query: tq, w: context.World},
	}
	return b6.AdaptCollection[b6.FeatureID, b6.PathFeature](c), nil
}

func findAreaFeatures(context *api.Context, query b6.Query) (b6.Collection[b6.FeatureID, b6.AreaFeature], error) {
	tq := b6.Typed{Type: b6.FeatureTypeArea, Query: query}
	c := b6.Collection[b6.FeatureID, b6.Feature]{
		AnyCollection: &searchFeatureCollection{query: tq, w: context.World},
	}
	return b6.AdaptCollection[b6.FeatureID, b6.AreaFeature](c), nil
}

func findRelationFeatures(context *api.Context, query b6.Query) (b6.Collection[b6.FeatureID, b6.RelationFeature], error) {
	tq := b6.Typed{Type: b6.FeatureTypeRelation, Query: query}
	c := b6.Collection[b6.FeatureID, b6.Feature]{
		AnyCollection: &searchFeatureCollection{query: tq, w: context.World},
	}
	return b6.AdaptCollection[b6.FeatureID, b6.RelationFeature](c), nil
}

func intersecting(context *api.Context, g b6.Geometry) (b6.Query, error) {
	switch g := g.(type) {
	case b6.Point:
		return b6.IntersectsPoint{Point: g.Point()}, nil
	case b6.Path:
		return b6.IntersectsPolyline{Polyline: g.Polyline()}, nil
	case b6.Area:
		return b6.IntersectsMultiPolygon{MultiPolygon: g.MultiPolygon()}, nil
	}
	return b6.Empty{}, nil
}

func intersectingCap(context *api.Context, center b6.Point, radius float64) (b6.Query, error) {
	return b6.NewIntersectsCap(s2.CapFromCenterAngle(center.Point(), b6.MetersToAngle(radius))), nil
}

func typePoint(context *api.Context) (*pb.QueryProto, error) {
	return &pb.QueryProto{
		Query: &pb.QueryProto_Typed{
			Typed: &pb.TypedQueryProto{
				Type: pb.FeatureType_FeatureTypePoint,
				Query: &pb.QueryProto{
					Query: &pb.QueryProto_All{
						All: &pb.AllQueryProto{},
					},
				},
			},
		},
	}, nil
}

func typePath(context *api.Context) (*pb.QueryProto, error) {
	return &pb.QueryProto{
		Query: &pb.QueryProto_Typed{
			Typed: &pb.TypedQueryProto{
				Type: pb.FeatureType_FeatureTypePath,
				Query: &pb.QueryProto{
					Query: &pb.QueryProto_All{
						All: &pb.AllQueryProto{},
					},
				},
			},
		},
	}, nil
}

func typeArea(context *api.Context) (*pb.QueryProto, error) {
	return &pb.QueryProto{
		Query: &pb.QueryProto_Typed{
			Typed: &pb.TypedQueryProto{
				Type: pb.FeatureType_FeatureTypeArea,
				Query: &pb.QueryProto{
					Query: &pb.QueryProto_All{
						All: &pb.AllQueryProto{},
					},
				},
			},
		},
	}, nil
}

func within(context *api.Context, a b6.Area) (b6.Query, error) {
	return b6.IntersectsMultiPolygon{MultiPolygon: a.MultiPolygon()}, nil
}

func withinCap(context *api.Context, p b6.Point, radius float64) (b6.Query, error) {
	return b6.NewIntersectsCap(s2.CapFromCenterAngle(p.Point(), b6.MetersToAngle(radius))), nil
}

func tagged(context *api.Context, key string, value string) (b6.Query, error) {
	return b6.Tagged{Key: key, Value: value}, nil
}

func keyed(context *api.Context, key string) (b6.Query, error) {
	return b6.Keyed{Key: key}, nil
}

func and(context *api.Context, a b6.Query, b b6.Query) (b6.Query, error) {
	return b6.Intersection{a, b}, nil
}

func or(context *api.Context, a b6.Query, b b6.Query) (b6.Query, error) {
	return b6.Union{a, b}, nil
}

func all(context *api.Context) (b6.Query, error) {
	return b6.All{}, nil
}
