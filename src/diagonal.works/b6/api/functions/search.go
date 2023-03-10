package functions

import (
	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/geometry"
	"diagonal.works/b6/ingest"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/search"
	ws "diagonal.works/b6/search/world"

	"github.com/golang/geo/s2"
)

type SearchFeatureCollection struct {
	query search.Query
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

func Find(query search.Query, context *api.Context) (api.FeatureCollection, error) {
	return &SearchFeatureCollection{query: query, w: context.World}, nil
}

func FindPointFeatures(query search.Query, context *api.Context) (api.PointFeatureCollection, error) {
	tq := b6.FeatureTypeQuery{Type: b6.FeatureTypePoint, Query: query}
	return &SearchFeatureCollection{query: tq, w: context.World}, nil
}

func FindPathFeatures(query search.Query, context *api.Context) (api.PathFeatureCollection, error) {
	tq := b6.FeatureTypeQuery{Type: b6.FeatureTypePath, Query: query}
	return &SearchFeatureCollection{query: tq, w: context.World}, nil
}

func FindAreaFeatures(query search.Query, context *api.Context) (api.AreaFeatureCollection, error) {
	tq := b6.FeatureTypeQuery{Type: b6.FeatureTypeArea, Query: query}
	return &SearchFeatureCollection{query: tq, w: context.World}, nil
}

func FindRelationFeatures(query search.Query, context *api.Context) (api.RelationFeatureCollection, error) {
	tq := b6.FeatureTypeQuery{Type: b6.FeatureTypeRelation, Query: query}
	return &SearchFeatureCollection{query: tq, w: context.World}, nil
}

func intersecting(g b6.Geometry, context *api.Context) (*pb.QueryProto, error) {
	// TODO: Clean up the definition of SpatialQueryProto, which dates from a time
	// before we had the full power of a5, allowing queries to be constructed server-side.
	// This also causes issues in the exact matching sematics, since we have both exact and
	// rough matching implemented.
	switch g := g.(type) {
	case b6.Point:
		return queryFromCap(s2.CapFromCenterAngle(g.Point(), b6.MetersToAngle(1.0))), nil
	case b6.Path:
		return queryFromCap(g.Polyline().CapBound()), nil
	case b6.Area:
		p := make(geometry.MultiPolygon, g.Len())
		for i := range p {
			p[i] = g.Polygon(i)
		}
		return &pb.QueryProto{
			Query: &pb.QueryProto_Spatial{
				Spatial: &pb.SpatialQueryProto{
					Area: &pb.AreaProto{
						Area: &pb.AreaProto_MultiPolygon{
							MultiPolygon: b6.NewMultiPolygonProto(p),
						},
					},
				},
			},
		}, nil
	}
	return &pb.QueryProto{
		Query: &pb.QueryProto_Intersection{
			Intersection: &pb.IntersectionQueryProto{
				Queries: []*pb.QueryProto{},
			},
		},
	}, nil
}

func typePoint(context *api.Context) (*pb.QueryProto, error) {
	return &pb.QueryProto{
		Query: &pb.QueryProto_Type{
			Type: &pb.TypeQueryProto{
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
	}, nil
}

func typeArea(context *api.Context) (*pb.QueryProto, error) {
	return &pb.QueryProto{
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
	}, nil
}

func within(a b6.Area, context *api.Context) (search.Query, error) {
	return ws.NewIntersectsArea(a), nil
}

func withinCap(p b6.Point, radius float64, context *api.Context) (search.Query, error) {
	return ws.NewIntersectsCap(s2.CapFromCenterAngle(p.Point(), b6.MetersToAngle(radius))), nil
}

func tagged(key string, value string, context *api.Context) (search.Query, error) {
	return ingest.QueryForKeyValue(key, value)
}

func keyed(key string, context *api.Context) (search.Query, error) {
	return ingest.QueryForAllValues(key)
}

func and(a search.Query, b search.Query, context *api.Context) (search.Query, error) {
	return search.Intersection{a, b}, nil
}

func or(a search.Query, b search.Query, context *api.Context) (search.Query, error) {
	return search.Union{a, b}, nil
}

func all(context *api.Context) (search.Query, error) {
	return search.All{Token: search.AllToken}, nil
}

func queryFromCap(cap s2.Cap) *pb.QueryProto {
	return &pb.QueryProto{
		Query: &pb.QueryProto_Spatial{
			Spatial: &pb.SpatialQueryProto{
				Area: &pb.AreaProto{
					Area: &pb.AreaProto_Cap{
						Cap: &pb.CapProto{
							Center:       b6.NewPointProtoFromS2Point(cap.Center()),
							RadiusMeters: b6.AngleToMeters(cap.Radius()),
						},
					},
				},
			},
		},
	}
}
