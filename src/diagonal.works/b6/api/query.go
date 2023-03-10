package api

import (
	"fmt"

	"diagonal.works/b6"
	pb "diagonal.works/b6/proto"
)

func Matches(feature b6.Feature, query *pb.QueryProto, w b6.World) bool {
	switch q := query.Query.(type) {
	case *pb.QueryProto_Key:
		return feature.Get(q.Key.Key).IsValid()
	case *pb.QueryProto_KeyValue:
		tag := feature.Get(q.KeyValue.Key)
		return tag.IsValid() && tag.Value == q.KeyValue.Value
	case *pb.QueryProto_Intersection:
		for _, subQuery := range q.Intersection.Queries {
			if !Matches(feature, subQuery, w) {
				return false
			}
		}
		return true
	case *pb.QueryProto_Union:
		for _, subQuery := range q.Union.Queries {
			if Matches(feature, subQuery, w) {
				return true
			}
		}
		return false
	case *pb.QueryProto_All:
		return true
	case *pb.QueryProto_Type:
		return feature.FeatureID().Type == b6.FeatureType(q.Type.Type) && Matches(feature, q.Type.Query, w)
	case *pb.QueryProto_Spatial:
		return NewIntersectQueryFromArea(q.Spatial.Area, w).Matches(feature)
	default:
		panic(fmt.Sprintf("Unhandled query %s", query))
	}
}
