package b6

import (
	"fmt"

	"diagonal.works/b6/geometry"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/units"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

func NewProtoFromFeatureID(id FeatureID) *pb.FeatureIDProto {
	p := &pb.FeatureIDProto{
		Namespace: string(id.Namespace),
		Value:     id.Value,
	}
	switch id.Type {
	case FeatureTypePoint:
		p.Type = pb.FeatureType_FeatureTypePoint
	case FeatureTypePath:
		p.Type = pb.FeatureType_FeatureTypePath
	case FeatureTypeArea:
		p.Type = pb.FeatureType_FeatureTypeArea
	case FeatureTypeRelation:
		p.Type = pb.FeatureType_FeatureTypeRelation
	case FeatureTypeInvalid:
		p.Type = pb.FeatureType_FeatureTypeInvalid
	default:
		panic(fmt.Sprintf("Invalid FeatureID.Type(): %s", id.Type))
	}
	return p
}

func NewFeatureIDFromProto(p *pb.FeatureIDProto) FeatureID {
	switch p.Type {
	case pb.FeatureType_FeatureTypePoint:
		return MakePointID(Namespace(p.Namespace), p.Value).FeatureID()
	case pb.FeatureType_FeatureTypePath:
		return MakePathID(Namespace(p.Namespace), p.Value).FeatureID()
	case pb.FeatureType_FeatureTypeArea:
		return MakeAreaID(Namespace(p.Namespace), p.Value).FeatureID()
	case pb.FeatureType_FeatureTypeRelation:
		return MakeRelationID(Namespace(p.Namespace), p.Value).FeatureID()
	case pb.FeatureType_FeatureTypeInvalid:
		return FeatureIDInvalid
	}
	panic(fmt.Sprintf("Invalid FeatureType: %s", p.Type))
}

func NewFeatureTypeFromProto(p pb.FeatureType) FeatureType {
	switch p {
	case pb.FeatureType_FeatureTypePoint:
		return FeatureTypePoint
	case pb.FeatureType_FeatureTypePath:
		return FeatureTypePath
	case pb.FeatureType_FeatureTypeArea:
		return FeatureTypeArea
	case pb.FeatureType_FeatureTypeRelation:
		return FeatureTypeRelation
	}
	panic(fmt.Sprintf("Invalid FeatureType: %s", p))
}

func newProtoFromTagged(tagged Tagged) []*pb.TagProto {
	tags := tagged.AllTags()
	p := make([]*pb.TagProto, len(tags))
	for i, tag := range tags {
		p[i] = &pb.TagProto{Key: tag.Key, Value: tag.Value}
	}
	return p
}

func NewProtoFromFeature(feature Feature) *pb.FeatureProto {
	switch f := feature.(type) {
	case PointFeature:
		return newProtoFromPointFeature(f)
	case PathFeature:
		return newProtoFromPathFeature(f)
	case AreaFeature:
		return newProtoFromAreaFeature(f)
	case RelationFeature:
		return newProtoFromRelationFeature(f)
	}
	panic(fmt.Sprintf("Can't handle feature %T", feature))
}

func newProtoFromPointFeature(f PointFeature) *pb.FeatureProto {
	return &pb.FeatureProto{
		Feature: &pb.FeatureProto_Point{
			Point: &pb.PointFeatureProto{
				Id:   NewProtoFromFeatureID(f.FeatureID()),
				Tags: newProtoFromTagged(f),
			},
		},
	}
}

func newProtoFromPathFeature(f PathFeature) *pb.FeatureProto {
	return &pb.FeatureProto{
		Feature: &pb.FeatureProto_Path{
			Path: &pb.PathFeatureProto{
				Id:           NewProtoFromFeatureID(f.FeatureID()),
				Tags:         newProtoFromTagged(f),
				LengthMeters: units.AngleToMeters(f.Polyline().Length()),
			},
		},
	}
}

func newProtoFromAreaFeature(f AreaFeature) *pb.FeatureProto {
	return &pb.FeatureProto{
		Feature: &pb.FeatureProto_Area{
			Area: &pb.AreaFeatureProto{
				Id:   NewProtoFromFeatureID(f.FeatureID()),
				Tags: newProtoFromTagged(f),
			},
		},
	}
}

func newProtoFromRelationFeature(f RelationFeature) *pb.FeatureProto {
	return &pb.FeatureProto{
		Feature: &pb.FeatureProto_Relation{
			Relation: &pb.RelationFeatureProto{
				Id:      NewProtoFromFeatureID(f.FeatureID()),
				Tags:    newProtoFromTagged(f),
				Members: newProtoFromRelationMembers(f),
			},
		},
	}
}

func newProtoFromRelationMembers(f RelationFeature) []*pb.RelationMemberProto {
	members := make([]*pb.RelationMemberProto, f.Len())
	for i := range members {
		member := f.Member(i)
		members[i] = &pb.RelationMemberProto{
			Id:   NewProtoFromFeatureID(member.ID),
			Role: member.Role,
		}
	}
	return members
}

func NewPointProto(lat float64, lng float64) *pb.PointProto {
	ll := s2.LatLngFromDegrees(lat, lng)
	point := &pb.PointProto{}
	point.LatE7 = ll.Lat.E7()
	point.LngE7 = ll.Lng.E7()
	return point
}

func NewPointProtoFromS2Point(point s2.Point) *pb.PointProto {
	return NewPointProtoFromS2LatLng(s2.LatLngFromPoint(point))
}

func NewPointProtoFromS2LatLng(ll s2.LatLng) *pb.PointProto {
	return NewPointProto(ll.Lat.Degrees(), ll.Lng.Degrees())
}

func PointProtoToS2LatLng(point *pb.PointProto) s2.LatLng {
	return s2.LatLng{
		Lat: s1.Angle(point.LatE7) * s1.E7,
		Lng: s1.Angle(point.LngE7) * s1.E7,
	}
}

func PointProtoToS2Point(point *pb.PointProto) s2.Point {
	return s2.PointFromLatLng(PointProtoToS2LatLng(point))
}

func S2LatLngToPointProto(ll s2.LatLng) *pb.PointProto {
	return NewPointProto(ll.Lat.Degrees(), ll.Lng.Degrees())
}

func NewPolylineProto(polyline *s2.Polyline) *pb.PolylineProto {
	p := &pb.PolylineProto{
		Points:       make([]*pb.PointProto, len(*polyline)),
		LengthMeters: AngleToMeters(polyline.Length()),
	}
	for i := 0; i < len(*polyline); i++ {
		p.Points[i] = NewPointProtoFromS2Point((*polyline)[i])
	}
	return p
}

func PolylineProtoToS2Polyline(p *pb.PolylineProto) *s2.Polyline {
	if p == nil {
		return nil
	}
	lls := make([]s2.LatLng, len(p.Points), len(p.Points))
	for i, point := range p.Points {
		lls[i] = PointProtoToS2LatLng(point)
	}
	return s2.PolylineFromLatLngs(lls)
}

func NewPolygonProto(polygon *s2.Polygon) *pb.PolygonProto {
	if polygon.NumLoops() > 0 {
		loopProtos := make([]*pb.LoopProto, polygon.NumLoops(), polygon.NumLoops())
		for i, loop := range polygon.Loops() {
			n := loop.NumVertices()
			loopProtos[i] = &pb.LoopProto{Points: make([]*pb.PointProto, n, n)}
			for j, point := range loop.Vertices() {
				if loop.IsHole() {
					loopProtos[i].Points[n-j-1] = S2LatLngToPointProto(s2.LatLngFromPoint(point))
				} else {
					loopProtos[i].Points[j] = S2LatLngToPointProto(s2.LatLngFromPoint(point))
				}
			}
		}
		return &pb.PolygonProto{Loops: loopProtos}
	}
	return nil
}

func NewPolygonProtoFromLoop(loop *s2.Loop) *pb.PolygonProto {
	return NewPolygonProto(s2.PolygonFromLoops([]*s2.Loop{loop}))
}

func NewMultiPolygonProto(polygons geometry.MultiPolygon) *pb.MultiPolygonProto {
	polygonProtos := make([]*pb.PolygonProto, len(polygons))
	for i, p := range polygons {
		polygonProtos[i] = NewPolygonProto(p)
	}
	return &pb.MultiPolygonProto{
		Polygons: polygonProtos,
	}
}

func PolygonProtoToS2Polygon(polygon *pb.PolygonProto) *s2.Polygon {
	if polygon == nil {
		return nil
	}
	s2loops := make([]*s2.Loop, 0, len(polygon.Loops))
	for _, loop := range polygon.Loops {
		s2loop := LoopProtoToS2Loop(loop)
		if s2loop != nil {
			s2loops = append(s2loops, s2loop)
		}
	}
	return s2.PolygonFromLoops(s2loops)
}

func MultiPolygonProtoToS2MultiPolygon(polygons *pb.MultiPolygonProto) geometry.MultiPolygon {
	s2MultiPolygon := make(geometry.MultiPolygon, len(polygons.Polygons))
	for i, polygon := range polygons.Polygons {
		s2MultiPolygon[i] = PolygonProtoToS2Polygon(polygon)
	}
	return s2MultiPolygon
}

func LoopProtoToS2Loop(loop *pb.LoopProto) *s2.Loop {
	if loop == nil {
		return nil
	}
	return s2.LoopFromPoints(LoopProtoToS2Points(loop))
}

func LoopProtoToS2Points(loop *pb.LoopProto) []s2.Point {
	if loop == nil {
		return nil
	}
	points := make([]s2.Point, len(loop.Points), len(loop.Points))
	for i, point := range loop.Points {
		points[i] = PointProtoToS2Point(point)
	}
	return points
}
