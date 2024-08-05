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

func (s *searchFeatureCollection) KeyExpression() b6.Expression {
	return b6.NewFeatureIDExpression(s.i.FeatureID())
}

func (s *searchFeatureCollection) ValueExpression() b6.Expression {
	return b6.NewCallExpression(
		b6.NewSymbolExpression("find-feature"),
		[]b6.Expression{
			s.KeyExpression(),
		},
	)
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

// Return a collection of the features present in the world that match the given query.
// Keys are IDs, and values are features.
func find(context *api.Context, query b6.Query) (b6.Collection[b6.FeatureID, b6.Feature], error) {
	return b6.Collection[b6.FeatureID, b6.Feature]{
		AnyCollection: &searchFeatureCollection{query: query, w: context.World},
	}, nil
}

// Return a collection of the area features present in the world that match the given query.
// Keys are IDs, and values are features.
func findAreaFeatures(context *api.Context, query b6.Query) (b6.Collection[b6.FeatureID, b6.AreaFeature], error) {
	tq := b6.Typed{Type: b6.FeatureTypeArea, Query: query}
	c := b6.Collection[b6.FeatureID, b6.Feature]{
		AnyCollection: &searchFeatureCollection{query: tq, w: context.World},
	}
	return b6.AdaptCollection[b6.FeatureID, b6.AreaFeature](c), nil
}

// Return a collection of the relation features present in the world that match the given query.
// Keys are IDs, and values are features.
func findRelationFeatures(context *api.Context, query b6.Query) (b6.Collection[b6.FeatureID, b6.RelationFeature], error) {
	tq := b6.Typed{Type: b6.FeatureTypeRelation, Query: query}
	c := b6.Collection[b6.FeatureID, b6.Feature]{
		AnyCollection: &searchFeatureCollection{query: tq, w: context.World},
	}
	return b6.AdaptCollection[b6.FeatureID, b6.RelationFeature](c), nil
}

// Return a query that will match features that intersect the given geometry.
func intersecting(context *api.Context, geometry b6.Geometry) (b6.Query, error) {
	if geometry != nil {
		switch geometry.GeometryType() {
		case b6.GeometryTypePoint:
			return b6.IntersectsPoint{Point: geometry.Point()}, nil
		case b6.GeometryTypePath:
			return b6.IntersectsPolyline{Polyline: geometry.Polyline()}, nil
		case b6.GeometryTypeArea:
			return b6.IntersectsMultiPolygon{MultiPolygon: geometry.(b6.Area).MultiPolygon()}, nil
		}
	}
	return b6.Empty{}, nil
}

// Return a query that will match features that intersect a spherical cap centred on the given point, with the given radius in meters.
func intersectingCap(context *api.Context, center b6.Geometry, radius float64) (b6.Query, error) {
	return b6.NewIntersectsCap(s2.CapFromCenterAngle(center.Point(), b6.MetersToAngle(radius))), nil
}

// Return a query that will match point features.
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

// Return a query that will match path features.
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

// Return a query that will match area features.
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

// Return a query that will match features that intersect the given area.
// Deprecated. Use intersecting.
func within(context *api.Context, a b6.Area) (b6.Query, error) {
	return b6.IntersectsMultiPolygon{MultiPolygon: a.MultiPolygon()}, nil
}

// Return a query that will match features that intersect a spherical cap centred on the given point, with the given radius in meters.
// Deprecated. Use intersecting-cap.
func withinCap(context *api.Context, point b6.Geometry, radius float64) (b6.Query, error) {
	return b6.NewIntersectsCap(s2.CapFromCenterAngle(point.Point(), b6.MetersToAngle(radius))), nil
}

// Return a query that will match features tagged with the given key and value.
func tagged(context *api.Context, key string, value string) (b6.Query, error) {
	return b6.Tagged{Key: key, Value: b6.NewStringExpression(value)}, nil
}

// Return a query that will match features tagged with the given key independent of value.
func keyed(context *api.Context, key string) (b6.Query, error) {
	return b6.Keyed{Key: key}, nil
}

// Wrap a query to only match features with the given feature type.
func typed(context *api.Context, typ string, q b6.Query) (b6.Query, error) {
	return b6.Typed{Type: b6.FeatureTypeFromString(typ), Query: q}, nil
}

// Return a query that will match features that match both given queries.
func and(context *api.Context, a b6.Query, b b6.Query) (b6.Query, error) {
	return b6.Intersection{a, b}, nil
}

// Return a query that will match features that match either of the given queries.
func or(context *api.Context, a b6.Query, b b6.Query) (b6.Query, error) {
	return b6.Union{a, b}, nil
}

// Return a query that will match any feature.
func all(context *api.Context) (b6.Query, error) {
	return b6.All{}, nil
}
