package api

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/geometry"
	"diagonal.works/b6/ingest"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/search"
	ws "diagonal.works/b6/search/world"

	"github.com/golang/geo/s2"
	"google.golang.org/protobuf/proto"
)

func ToProto(v interface{}) (*pb.NodeProto, error) {
	if v == nil {
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_NilValue{},
				},
			},
		}, nil
	}
	switch v := v.(type) {
	case bool:
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_BoolValue{
						BoolValue: v,
					},
				},
			},
		}, nil
	case int:
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_IntValue{
						IntValue: int64(v),
					},
				},
			},
		}, nil
	case int8:
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_IntValue{
						IntValue: int64(v),
					},
				},
			},
		}, nil
	case int16:
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_IntValue{
						IntValue: int64(v),
					},
				},
			},
		}, nil
	case int32:
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_IntValue{
						IntValue: int64(v),
					},
				},
			},
		}, nil
	case int64:
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_IntValue{
						IntValue: int64(v),
					},
				},
			},
		}, nil
	case float64:
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_FloatValue{
						FloatValue: v,
					},
				},
			},
		}, nil
	case IntNumber:
		return ToProto(int(v))
	case FloatNumber:
		return ToProto(float64(v))
	case string:
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_StringValue{
						StringValue: v,
					},
				},
			},
		}, nil
	case b6.FeatureID:
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_FeatureIDValue{
						FeatureIDValue: b6.NewProtoFromFeatureID(v),
					},
				},
			},
		}, nil
	// Order is significant here, as Features will also implement
	// one of the geometry types below.
	case b6.Feature:
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_FeatureValue{
						FeatureValue: b6.NewProtoFromFeature(v),
					},
				},
			},
		}, nil
	case b6.Point:
		ll := s2.LatLngFromPoint(v.Point())
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_PointValue{
						PointValue: b6.S2LatLngToPointProto(ll),
					},
				},
			},
		}, nil
	case b6.Path:
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_PathValue{
						PathValue: b6.NewPolylineProto(v.Polyline()),
					},
				},
			},
		}, nil
	case b6.Area:
		polygons := make(geometry.MultiPolygon, v.Len())
		for i := 0; i < v.Len(); i++ {
			polygons[i] = v.Polygon(i)
		}
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_AreaValue{
						AreaValue: b6.NewMultiPolygonProto(polygons),
					},
				},
			},
		}, nil
	case Collection:
		return collectionToProto(v)
	case b6.Tag:
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_TagValue{
						TagValue: &pb.TagProto{
							Key:   v.Key,
							Value: v.Value,
						},
					},
				},
			},
		}, nil
	case Pair:
		fp, err := ToProto(v.First())
		if err != nil {
			return nil, err
		}
		sp, err := ToProto(v.Second())
		if err != nil {
			return nil, err
		}
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_PairValue{
						PairValue: &pb.PairProto{
							First:  fp.GetLiteral(),
							Second: sp.GetLiteral(),
						},
					},
				},
			},
		}, nil
	case geojson.GeoJSON:
		marshalled, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		var buffer bytes.Buffer
		writer := gzip.NewWriter(&buffer)
		_, err = writer.Write(marshalled)
		if err != nil {
			return nil, err
		}
		if err = writer.Close(); err != nil {
			return nil, err
		}
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_GeoJSONValue{
						GeoJSONValue: buffer.Bytes(),
					},
				},
			},
		}, nil
	case *pb.QueryProto:
		n := &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_QueryValue{
						QueryValue: &pb.QueryProto{},
					},
				},
			},
		}
		proto.Merge(n.GetLiteral().GetQueryValue(), v)
		return n, nil
	case search.Query:
		if q, err := QueryToProto(v); err == nil {
			return ToProto(q)
		} else {
			return nil, err
		}
	case ingest.AppliedChange:
		c := &pb.AppliedChangeProto{
			Original: make([]*pb.FeatureIDProto, 0, len(v)),
			Modified: make([]*pb.FeatureIDProto, 0, len(v)),
		}
		for original, modified := range v {
			c.Original = append(c.Original, b6.NewProtoFromFeatureID(original))
			c.Modified = append(c.Modified, b6.NewProtoFromFeatureID(modified))
		}
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_AppliedChangeValue{
						AppliedChangeValue: c,
					},
				},
			},
		}, nil
	case Callable:
		// TODO: this could be better - in particular, we'd want to return an expression, not
		// the debug representation we use here.
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_StringValue{
						StringValue: v.String(),
					},
				},
			},
		}, nil
	default:
		panic(fmt.Sprintf("can't convert %T to proto", v))
	}
}

func collectionToProto(collection Collection) (*pb.NodeProto, error) {
	i := collection.Begin()
	keys := make([]*pb.LiteralNodeProto, 0)
	values := make([]*pb.LiteralNodeProto, 0)
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		if p, err := ToProto(i.Key()); err == nil {
			keys = append(keys, p.GetLiteral())
		}
		if p, err := ToProto(i.Value()); err == nil {
			values = append(values, p.GetLiteral())
		} else {
			return p, err
		}
	}
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_CollectionValue{
					CollectionValue: &pb.CollectionProto{
						Keys:   keys,
						Values: values,
					},
				},
			},
		},
	}, nil
}

func NewQueryFromProto(p *pb.QueryProto, w b6.World) (search.Query, error) {
	switch q := p.Query.(type) {
	case *pb.QueryProto_Key:
		return ingest.QueryForAllValues(q.Key.Key)
	case *pb.QueryProto_KeyValue:
		return ingest.QueryForKeyValue(q.KeyValue.Key, q.KeyValue.Value)
	case *pb.QueryProto_Spatial:
		// TODO: rename spatial to intersects?
		return NewIntersectQueryFromArea(q.Spatial.Area, w), nil
	case *pb.QueryProto_Type:
		if q.Type.Query != nil {
			child, err := NewQueryFromProto(q.Type.Query, w)
			if err != nil {
				return search.Empty{}, err
			}
			return b6.FeatureTypeQuery{Type: b6.NewFeatureTypeFromProto(q.Type.Type), Query: child}, nil
		}
	case *pb.QueryProto_Intersection:
		intersection := make(search.Intersection, len(q.Intersection.Queries))
		for i, next := range q.Intersection.Queries {
			if child, err := NewQueryFromProto(next, w); err == nil {
				intersection[i] = child
			} else {
				return search.Empty{}, err
			}
		}
		return intersection, nil
	case *pb.QueryProto_Union:
		union := make(search.Union, len(q.Union.Queries))
		for i, next := range q.Union.Queries {
			if child, err := NewQueryFromProto(next, w); err == nil {
				union[i] = child
			} else {
				return search.Empty{}, err
			}
		}
		return union, nil
	case *pb.QueryProto_All:
		return search.All{Token: search.AllToken}, nil
	}
	return search.Empty{}, fmt.Errorf("Can't handle query %v", p)
}

func QueryToProto(q search.Query) (*pb.QueryProto, error) {
	switch q := q.(type) {
	case search.All:
		if q.Token == search.AllToken {
			return &pb.QueryProto{
				Query: &pb.QueryProto_All{},
			}, nil
		} else if i := strings.Index(q.Token, "="); i >= 0 {
			return &pb.QueryProto{
				Query: &pb.QueryProto_KeyValue{
					KeyValue: &pb.KeyValueQueryProto{
						Key:   "#" + q.Token[0:i],
						Value: q.Token[i+1:],
					},
				},
			}, nil
		}
	case search.TokenPrefix:
		if strings.Index(q.Prefix, "=") == len(q.Prefix)-1 {
			return &pb.QueryProto{
				Query: &pb.QueryProto_Key{
					Key: &pb.KeyQueryProto{
						Key: "#" + q.Prefix[0:len(q.Prefix)-1],
					},
				},
			}, nil
		}
	case search.Intersection:
		p := &pb.IntersectionQueryProto{
			Queries: make([]*pb.QueryProto, len(q)),
		}
		for i, qq := range q {
			if pp, err := QueryToProto(qq); err == nil {
				p.Queries[i] = pp
			} else {
				return nil, err
			}
		}
		return &pb.QueryProto{
			Query: &pb.QueryProto_Intersection{
				Intersection: p,
			},
		}, nil
	case search.Union:
		p := &pb.UnionQueryProto{
			Queries: make([]*pb.QueryProto, len(q)),
		}
		for i, qq := range q {
			if pp, err := QueryToProto(qq); err == nil {
				p.Queries[i] = pp
			} else {
				return nil, err
			}
		}
		return &pb.QueryProto{
			Query: &pb.QueryProto_Union{
				Union: p,
			},
		}, nil
	}
	return nil, fmt.Errorf("Can't convert query %s", q)
}

func NewIntersectQueryFromArea(p *pb.AreaProto, w b6.World) b6.FeatureQuery {
	switch a := p.Area.(type) {
	case *pb.AreaProto_Cap:
		ll := b6.PointProtoToS2LatLng(a.Cap.Center)
		cap := s2.CapFromCenterAngle(s2.PointFromLatLng(ll), b6.MetersToAngle(a.Cap.RadiusMeters))
		return ws.NewIntersectsCap(cap)
	case *pb.AreaProto_Id:
		if f := w.FindFeatureByID(b6.NewFeatureIDFromProto(a.Id)); f != nil {
			if p, ok := f.(b6.PhysicalFeature); ok {
				return ws.NewIntersectsFeature(p)
			}
		}
		return ws.Empty{}
	case *pb.AreaProto_MultiPolygon:
		return ws.NewIntersectsMultiPolygon(b6.MultiPolygonProtoToS2MultiPolygon(a.MultiPolygon))
	default:
		panic(fmt.Sprintf("Unhandled area: %s", p))
	}
}
