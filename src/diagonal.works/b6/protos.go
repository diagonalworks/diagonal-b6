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
	return &pb.FeatureIDProto{
		Namespace: string(id.Namespace),
		Type:      NewProtoFromFeatureType(id.Type),
		Value:     id.Value,
	}
}

func NewFeatureIDFromProto(p *pb.FeatureIDProto) FeatureID {
	return FeatureID{
		Type:      NewFeatureTypeFromProto(p.Type),
		Namespace: Namespace(p.Namespace),
		Value:     p.Value,
	}
}

func NewFeatureTypeFromProto(t pb.FeatureType) FeatureType {
	switch t {
	case pb.FeatureType_FeatureTypePoint:
		return FeatureTypePoint
	case pb.FeatureType_FeatureTypePath:
		return FeatureTypePath
	case pb.FeatureType_FeatureTypeArea:
		return FeatureTypeArea
	case pb.FeatureType_FeatureTypeRelation:
		return FeatureTypeRelation
	case pb.FeatureType_FeatureTypeCollection:
		return FeatureTypeCollection
	case pb.FeatureType_FeatureTypeExpression:
		return FeatureTypeExpression
	}
	panic(fmt.Sprintf("Invalid pb.FeatureType: %s", t))
}

func NewProtoFromFeatureType(t FeatureType) pb.FeatureType {
	switch t {
	case FeatureTypePoint:
		return pb.FeatureType_FeatureTypePoint
	case FeatureTypePath:
		return pb.FeatureType_FeatureTypePath
	case FeatureTypeArea:
		return pb.FeatureType_FeatureTypeArea
	case FeatureTypeRelation:
		return pb.FeatureType_FeatureTypeRelation
	case FeatureTypeCollection:
		return pb.FeatureType_FeatureTypeCollection
	case FeatureTypeExpression:
		return pb.FeatureType_FeatureTypeExpression
	}
	panic(fmt.Sprintf("Invalid FeatureType: %s", t))
}

func newProtoFromTagged(t Taggable) []*pb.TagProto {
	tags := t.AllTags()
	p := make([]*pb.TagProto, len(tags))
	for i, tag := range tags {
		p[i] = &pb.TagProto{Key: tag.Key, Value: tag.Value}
	}
	return p
}

func NewProtoFromFeature(feature Feature) (*pb.FeatureProto, error) {
	switch f := feature.(type) {
	case PointFeature:
		return newProtoFromPointFeature(f)
	case PathFeature:
		return newProtoFromPathFeature(f)
	case AreaFeature:
		return newProtoFromAreaFeature(f)
	case RelationFeature:
		return newProtoFromRelationFeature(f)
	case CollectionFeature:
		return newProtoFromCollectionFeature(f)
	case ExpressionFeature:
		return newProtoFromExpressionFeature(f)
	}
	panic(fmt.Sprintf("Can't handle feature %T", feature))
}

func newProtoFromPointFeature(f PointFeature) (*pb.FeatureProto, error) {
	return &pb.FeatureProto{
		Feature: &pb.FeatureProto_Point{
			Point: &pb.PointFeatureProto{
				Id:   NewProtoFromFeatureID(f.FeatureID()),
				Tags: newProtoFromTagged(f),
			},
		},
	}, nil
}

func newProtoFromPathFeature(f PathFeature) (*pb.FeatureProto, error) {
	return &pb.FeatureProto{
		Feature: &pb.FeatureProto_Path{
			Path: &pb.PathFeatureProto{
				Id:           NewProtoFromFeatureID(f.FeatureID()),
				Tags:         newProtoFromTagged(f),
				LengthMeters: units.AngleToMeters(f.Polyline().Length()),
			},
		},
	}, nil
}

func newProtoFromAreaFeature(f AreaFeature) (*pb.FeatureProto, error) {
	return &pb.FeatureProto{
		Feature: &pb.FeatureProto_Area{
			Area: &pb.AreaFeatureProto{
				Id:   NewProtoFromFeatureID(f.FeatureID()),
				Tags: newProtoFromTagged(f),
			},
		},
	}, nil
}

func newProtoFromRelationFeature(f RelationFeature) (*pb.FeatureProto, error) {
	return &pb.FeatureProto{
		Feature: &pb.FeatureProto_Relation{
			Relation: &pb.RelationFeatureProto{
				Id:      NewProtoFromFeatureID(f.FeatureID()),
				Tags:    newProtoFromTagged(f),
				Members: newProtoFromRelationMembers(f),
			},
		},
	}, nil
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

func newProtoFromCollectionFeature(f CollectionFeature) (*pb.FeatureProto, error) {
	c := CollectionExpression{UntypedCollection: f}
	if node, err := c.ToProto(); err == nil {
		return &pb.FeatureProto{
			Feature: &pb.FeatureProto_Collection{
				Collection: &pb.CollectionFeatureProto{
					Id:         NewProtoFromFeatureID(f.FeatureID()),
					Tags:       newProtoFromTagged(f),
					Collection: node.GetLiteral().GetCollectionValue(),
				},
			},
		}, nil
	} else {
		return nil, err
	}
}

func newProtoFromExpressionFeature(f ExpressionFeature) (*pb.FeatureProto, error) {
	e := f.Expression()
	if node, err := e.ToProto(); err == nil {
		return &pb.FeatureProto{
			Feature: &pb.FeatureProto_Expression{
				Expression: &pb.ExpressionFeatureProto{
					Id:         NewProtoFromFeatureID(f.FeatureID()),
					Tags:       newProtoFromTagged(f),
					Expression: node,
				},
			},
		}, nil
	} else {
		return nil, err
	}
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

func NewProtoFromRoute(route Route) *pb.RouteProto {
	p := &pb.RouteProto{
		Origin: NewProtoFromFeatureID(route.Origin.FeatureID()),
		Steps:  make([]*pb.StepProto, len(route.Steps)),
	}
	for i, step := range route.Steps {
		p.Steps[i] = &pb.StepProto{
			Destination: NewProtoFromFeatureID(step.Destination.FeatureID()),
			Via:         NewProtoFromFeatureID(step.Via.FeatureID()),
			Cost:        step.Cost,
		}
	}
	return p
}

func NewRouteFromProto(p *pb.RouteProto) Route {
	route := Route{
		Origin: NewFeatureIDFromProto(p.Origin).ToPointID(),
		Steps:  make([]Step, len(p.Steps)),
	}
	for i, step := range p.Steps {
		route.Steps[i] = Step{
			Destination: NewFeatureIDFromProto(step.Destination).ToPointID(),
			Via:         NewFeatureIDFromProto(step.Via).ToPathID(),
			Cost:        step.Cost,
		}
	}
	return route
}
