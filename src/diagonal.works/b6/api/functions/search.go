package functions

import (
	"diagonal.works/b6"
	"diagonal.works/b6/api"
	pb "diagonal.works/b6/proto"

	"github.com/golang/geo/s2"
)

type SearchFeatureCollection struct {
	query b6.Query
	w     b6.World
	i     b6.Features
}

func (s *SearchFeatureCollection) Begin() api.CollectionIterator {
	return &SearchFeatureCollection{
		query: s.query,
		w:     s.w,
	}
}

func (s *SearchFeatureCollection) Key() interface{} {
	return s.FeatureIDKey()
}

func (s *SearchFeatureCollection) Value() interface{} {
	return s.FeatureValue()
}

func (s *SearchFeatureCollection) FeatureIDKey() b6.FeatureID {
	return s.i.FeatureID()
}

func (s *SearchFeatureCollection) FeatureValue() b6.Feature {
	return s.i.Feature()
}

func (s *SearchFeatureCollection) Next() (bool, error) {
	if s.i == nil {
		s.i = s.w.FindFeatures(s.query)
	}
	return s.i.Next(), nil
}

func (s *SearchFeatureCollection) Count() int {
	n := 0
	i := s.w.FindFeatures(s.query)
	for i.Next() {
		n++
	}
	return n
}

var _ api.Collection = &SearchFeatureCollection{}

func Find(query b6.Query, context *api.Context) (api.FeatureCollection, error) {
	return &SearchFeatureCollection{query: query, w: context.World}, nil
}

func FindPointFeatures(query b6.Query, context *api.Context) (api.PointFeatureCollection, error) {
	tq := b6.Typed{Type: b6.FeatureTypePoint, Query: query}
	return &SearchFeatureCollection{query: tq, w: context.World}, nil
}

func FindPathFeatures(query b6.Query, context *api.Context) (api.PathFeatureCollection, error) {
	tq := b6.Typed{Type: b6.FeatureTypePath, Query: query}
	return &SearchFeatureCollection{query: tq, w: context.World}, nil
}

func FindAreaFeatures(query b6.Query, context *api.Context) (api.AreaFeatureCollection, error) {
	tq := b6.Typed{Type: b6.FeatureTypeArea, Query: query}
	return &SearchFeatureCollection{query: tq, w: context.World}, nil
}

func FindRelationFeatures(query b6.Query, context *api.Context) (api.RelationFeatureCollection, error) {
	tq := b6.Typed{Type: b6.FeatureTypeRelation, Query: query}
	return &SearchFeatureCollection{query: tq, w: context.World}, nil
}

func intersecting(g b6.Geometry, context *api.Context) (b6.Query, error) {
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

func intersectingCap(center b6.Point, radius float64, context *api.Context) (b6.Query, error) {
	return &b6.IntersectsCap{Cap: s2.CapFromCenterAngle(center.Point(), b6.MetersToAngle(radius))}, nil
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

func within(a b6.Area, context *api.Context) (b6.Query, error) {
	return b6.IntersectsMultiPolygon{MultiPolygon: a.MultiPolygon()}, nil
}

func withinCap(p b6.Point, radius float64, context *api.Context) (b6.Query, error) {
	return &b6.IntersectsCap{Cap: s2.CapFromCenterAngle(p.Point(), b6.MetersToAngle(radius))}, nil
}

func tagged(key string, value string, context *api.Context) (b6.Query, error) {
	return b6.Tagged{Key: key, Value: value}, nil
}

func keyed(key string, context *api.Context) (b6.Query, error) {
	return b6.Keyed{key}, nil
}

func and(a b6.Query, b b6.Query, context *api.Context) (b6.Query, error) {
	return b6.Intersection{a, b}, nil
}

func or(a b6.Query, b b6.Query, context *api.Context) (b6.Query, error) {
	return b6.Union{a, b}, nil
}

func all(context *api.Context) (b6.Query, error) {
	return b6.All{}, nil
}

func queryFromCap(cap s2.Cap) *pb.QueryProto {
	return &pb.QueryProto{
		Query: &pb.QueryProto_IntersectsCap{
			IntersectsCap: &pb.CapProto{
				Center:       b6.NewPointProtoFromS2Point(cap.Center()),
				RadiusMeters: b6.AngleToMeters(cap.Radius()),
			},
		},
	}
}
