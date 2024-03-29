package b6

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"

	"diagonal.works/b6/geojson"
	"diagonal.works/b6/geometry"
	pb "diagonal.works/b6/proto"
	"github.com/golang/geo/s2"
	"gopkg.in/yaml.v2"
)

// TODO: Use constraints.Integer etc?
type Number interface {
	isNumber()
}

type IntNumber int

func (IntNumber) isNumber() {}

type FloatNumber float64

func (FloatNumber) isNumber() {}

type AnyExpression interface {
	ToProto() (*pb.NodeProto, error)
	FromProto(node *pb.NodeProto) error
	Equal(other AnyExpression) bool
	Clone() Expression
}

type Expression struct {
	AnyExpression
	Name  string
	Begin int
	End   int
}

func (e *Expression) IsValid() bool {
	return e.AnyExpression != nil
}

type expressionChoices struct {
	Symbol     *SymbolExpression
	Int        *IntExpression
	Float      *FloatExpression
	Bool       *BoolExpression
	String     *StringExpression
	ID         *FeatureIDExpression
	Tag        *TagExpression
	Query      *QueryExpression
	GeoJSON    *GeoJSONExpression
	Feature    *FeatureExpression
	Point      *PointExpression
	Path       *PathExpression
	Area       *AreaExpression
	Nil        *NilExpression
	Collection *CollectionExpression
	Call       *CallExpression
	Lambda     *LambdaExpression
}

func (e Expression) MarshalYAML() (interface{}, error) {
	// Fast track types that are handled natively by YAML
	if e.Name == "" {
		switch e := e.AnyExpression.(type) {
		case *IntExpression:
			return int(*e), nil
		case *FloatExpression:
			return float64(*e), nil
		case *StringExpression:
			return string(*e), nil
		}
	}

	y := expressionYAML{
		Name:  e.Name,
		Begin: e.Begin,
		End:   e.End,
	}
	return marshalChoiceYAML(&expressionChoices{}, e.AnyExpression, &y)
}

func (e Expression) Format() string {
	if j, err := yaml.Marshal(e); err == nil {
		return string(j)
	} else {
		return err.Error()
	}
}

type expressionYAML struct {
	Name  string `yaml:"name,omitempty"`
	Begin int    `yaml:"begin,omitempty"`
	End   int    `yaml:"end,omitempty"`
}

func (e *Expression) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Fast track types that are handled natively by YAML
	var v interface{}
	if err := unmarshal(&v); err != nil {
		return err
	}
	switch v := v.(type) {
	case int:
		i := IntExpression(v)
		e.AnyExpression = &i
		return nil
	case float64:
		f := FloatExpression(v)
		e.AnyExpression = &f
		return nil
	case string:
		s := StringExpression(v)
		e.AnyExpression = &s
		return nil
	}
	choice, err := unmarshalChoiceYAML(&expressionChoices{}, unmarshal)
	if err == nil {
		e.AnyExpression = choice.(AnyExpression)
	}
	y := expressionYAML{}
	if err := unmarshal(&y); err == nil {
		e.Name = y.Name
		e.Begin = y.Begin
		e.End = y.End
	}
	return err
}

func (e Expression) ToProto() (*pb.NodeProto, error) {
	p, err := e.AnyExpression.ToProto()
	p.Name = e.Name
	p.Begin = int32(e.Begin)
	p.End = int32(e.End)
	return p, err
}

func (e *Expression) FromProto(node *pb.NodeProto) error {
	switch n := node.Node.(type) {
	case *pb.NodeProto_Symbol:
		e.AnyExpression = new(SymbolExpression)
	case *pb.NodeProto_Call:
		e.AnyExpression = &CallExpression{}
	case *pb.NodeProto_Lambda_:
		e.AnyExpression = &LambdaExpression{}
	case *pb.NodeProto_Literal:
		switch n.Literal.Value.(type) {
		case *pb.LiteralNodeProto_IntValue:
			e.AnyExpression = new(IntExpression)
		case *pb.LiteralNodeProto_FloatValue:
			e.AnyExpression = new(FloatExpression)
		case *pb.LiteralNodeProto_BoolValue:
			e.AnyExpression = new(BoolExpression)
		case *pb.LiteralNodeProto_StringValue:
			e.AnyExpression = new(StringExpression)
		case *pb.LiteralNodeProto_FeatureIDValue:
			e.AnyExpression = new(FeatureIDExpression)
		case *pb.LiteralNodeProto_TagValue:
			e.AnyExpression = new(TagExpression)
		case *pb.LiteralNodeProto_PointValue:
			e.AnyExpression = new(PointExpression)
		case *pb.LiteralNodeProto_PathValue:
			e.AnyExpression = new(PathExpression)
		case *pb.LiteralNodeProto_AreaValue:
			e.AnyExpression = new(AreaExpression)
		case *pb.LiteralNodeProto_QueryValue:
			e.AnyExpression = new(QueryExpression)
		case *pb.LiteralNodeProto_NilValue:
			e.AnyExpression = new(NilExpression)
		case *pb.LiteralNodeProto_GeoJSONValue:
			e.AnyExpression = new(GeoJSONExpression)
		case *pb.LiteralNodeProto_RouteValue:
			e.AnyExpression = new(RouteExpression)
		case *pb.LiteralNodeProto_CollectionValue:
			e.AnyExpression = new(CollectionExpression)
		default:
			return fmt.Errorf("Can't convert %T from literal proto", n.Literal.Value)
		}
	default:
		return fmt.Errorf("Can't convert expression from proto %T", node.Node)
	}
	if err := e.AnyExpression.FromProto(node); err != nil {
		return err
	}
	e.Name = node.Name
	e.Begin = int(node.Begin)
	e.End = int(node.End)
	return nil
}

func (e Expression) Clone() Expression {
	clone := e
	clone.AnyExpression = e.AnyExpression.Clone().AnyExpression
	return clone
}

func (e Expression) Equal(other Expression) bool {
	return e.AnyExpression.Equal(other.AnyExpression)
}

type AnyLiteral interface {
	AnyExpression
	Literal() interface{}
}

type Literal struct {
	AnyLiteral
}

func (l *Literal) ToProto() (*pb.NodeProto, error) {
	return l.AnyLiteral.ToProto()
}

func (l *Literal) FromProto(node *pb.NodeProto) error {
	var e Expression
	if err := e.FromProto(node); err != nil {
		return err
	}
	if literal, ok := e.AnyExpression.(AnyLiteral); ok {
		l.AnyLiteral = literal
	} else {
		return fmt.Errorf("Can't convert literal from proto %T", node.Node)
	}
	return nil
}

func (l Literal) MarshalYAML() (interface{}, error) {
	e := Expression{AnyExpression: l.AnyLiteral}
	return e.MarshalYAML()
}

func (l *Literal) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var e Expression
	e.UnmarshalYAML(unmarshal)
	if literal, ok := e.AnyExpression.(AnyLiteral); ok {
		l.AnyLiteral = literal
	} else {
		return fmt.Errorf("Can't convert literal from yaml %T", e.AnyExpression)
	}
	return nil
}

func (l Literal) Equal(other Literal) bool {
	return l.AnyLiteral.Equal(other.AnyLiteral)
}

func FromLiteral(l interface{}) (Literal, error) {
	if l == nil {
		return Literal{AnyLiteral: NilExpression{}}, nil
	}
	switch l := l.(type) {
	case int:
		i := IntExpression(l)
		return Literal{AnyLiteral: &i}, nil
	case IntNumber:
		i := FloatExpression(int(l))
		return Literal{AnyLiteral: &i}, nil
	case float64:
		f := FloatExpression(l)
		return Literal{AnyLiteral: &f}, nil
	case FloatNumber:
		f := FloatExpression(float64(l))
		return Literal{AnyLiteral: &f}, nil
	case bool:
		b := BoolExpression(l)
		return Literal{AnyLiteral: &b}, nil
	case string:
		s := StringExpression(l)
		return Literal{AnyLiteral: &s}, nil
	case FeatureID:
		id := FeatureIDExpression(l)
		return Literal{AnyLiteral: &id}, nil
	case PathID:
		return FromLiteral(l.FeatureID())
	case AreaID:
		return FromLiteral(l.FeatureID())
	case RelationID:
		return FromLiteral(l.FeatureID())
	case CollectionID:
		return FromLiteral(l.FeatureID())
	case ExpressionID:
		return FromLiteral(l.FeatureID())
	case Tag:
		tag := TagExpression(l)
		return Literal{AnyLiteral: &tag}, nil
	case Feature:
		f := FeatureExpression{Feature: l}
		return Literal{AnyLiteral: &f}, nil
	case LatLng:
		// TODO: Remove and only use Geo(metry). Needed because
		// tag values can current be LatLng.
		ll := PointExpression(l)
		return Literal{AnyLiteral: &ll}, nil
	case Geo:
		ll := PointExpression(l.Location())
		return Literal{AnyLiteral: &ll}, nil
	case s2.LatLng:
		ll := PointExpression(l)
		return Literal{AnyLiteral: &ll}, nil
	case Path:
		return Literal{AnyLiteral: &PathExpression{Path: l}}, nil
	case Area:
		return Literal{AnyLiteral: &AreaExpression{Area: l}}, nil
	case geojson.GeoJSON:
		return Literal{AnyLiteral: &GeoJSONExpression{GeoJSON: l}}, nil
	case Route:
		route := RouteExpression(l)
		return Literal{AnyLiteral: &route}, nil
	case UntypedCollection:
		return Literal{AnyLiteral: &CollectionExpression{UntypedCollection: l}}, nil
	}
	return Literal{}, fmt.Errorf("Can't make literal from %T", l)
}

func LiteralEqual(v interface{}, vv interface{}) bool {
	if l, err := FromLiteral(v); err == nil {
		if ll, err := FromLiteral(vv); err == nil {
			return l.Equal(ll)
		}
	}
	return false
}

type SymbolExpression string

func (s *SymbolExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Symbol{
			Symbol: string(*s),
		},
	}, nil
}

func (s *SymbolExpression) FromProto(node *pb.NodeProto) error {
	*s = SymbolExpression(node.GetSymbol())
	return nil
}

func (s *SymbolExpression) Clone() Expression {
	clone := *s
	return Expression{AnyExpression: &clone}
}

func (s SymbolExpression) String() string {
	return string(s)
}

func (s SymbolExpression) Equal(other AnyExpression) bool {
	if ss, ok := other.(*SymbolExpression); ok {
		return s == *ss
	}
	return false
}

func NewSymbolExpression(symbol string) Expression {
	s := SymbolExpression(symbol)
	return Expression{AnyExpression: &s}
}

type IntExpression int

func (i *IntExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_IntValue{
					IntValue: int64(*i),
				},
			},
		},
	}, nil
}

func (i *IntExpression) FromProto(node *pb.NodeProto) error {
	*i = IntExpression(node.GetLiteral().GetIntValue())
	return nil
}

func (i *IntExpression) Clone() Expression {
	clone := *i
	return Expression{AnyExpression: &clone}
}

func (i IntExpression) Literal() interface{} {
	return int(i)
}

func (i IntExpression) Equal(other AnyExpression) bool {
	if ii, ok := other.(*IntExpression); ok {
		return i == *ii
	}
	return false
}

func NewIntExpression(value int) Expression {
	i := IntExpression(value)
	return Expression{AnyExpression: &i}
}

type FloatExpression float64

func (f *FloatExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_FloatValue{
					FloatValue: float64(*f),
				},
			},
		},
	}, nil
}

func (f *FloatExpression) FromProto(node *pb.NodeProto) error {
	*f = FloatExpression(node.GetLiteral().GetFloatValue())
	return nil
}

func (f *FloatExpression) Clone() Expression {
	clone := *f
	return Expression{AnyExpression: &clone}
}

func (f FloatExpression) Literal() interface{} {
	return float64(f)
}

func (f FloatExpression) Equal(other AnyExpression) bool {
	if ff, ok := other.(*FloatExpression); ok {
		return f == *ff
	}
	return false
}

func NewFloatExpression(value float64) Expression {
	i := FloatExpression(value)
	return Expression{AnyExpression: &i}
}

type BoolExpression bool

func (b *BoolExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_BoolValue{
					BoolValue: bool(*b),
				},
			},
		},
	}, nil
}

func (b *BoolExpression) FromProto(node *pb.NodeProto) error {
	*b = BoolExpression(node.GetLiteral().GetBoolValue())
	return nil
}

func (b *BoolExpression) Clone() Expression {
	clone := *b
	return Expression{AnyExpression: &clone}
}

func (b BoolExpression) Literal() interface{} {
	return bool(b)
}

func (b BoolExpression) Equal(other AnyExpression) bool {
	if bb, ok := other.(*BoolExpression); ok {
		return b == *bb
	}
	return false
}

type StringExpression string

func (s *StringExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_StringValue{
					StringValue: string(*s),
				},
			},
		},
	}, nil
}

func (s *StringExpression) FromProto(node *pb.NodeProto) error {
	*s = StringExpression(node.GetLiteral().GetStringValue())
	return nil
}

func (s *StringExpression) Clone() Expression {
	clone := *s
	return Expression{AnyExpression: &clone}
}

func (s StringExpression) Literal() interface{} {
	return string(s)
}

func (s StringExpression) Equal(other AnyExpression) bool {
	if ss, ok := other.(*StringExpression); ok {
		return s == *ss
	}
	return false
}

func NewStringExpression(s string) Expression {
	ss := StringExpression(s)
	return Expression{AnyExpression: &ss}
}

type FeatureIDExpression FeatureID

func (f *FeatureIDExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_FeatureIDValue{
					FeatureIDValue: NewProtoFromFeatureID(FeatureID(*f)),
				},
			},
		},
	}, nil
}

func (f *FeatureIDExpression) FromProto(node *pb.NodeProto) error {
	*f = FeatureIDExpression(NewFeatureIDFromProto(node.GetLiteral().GetFeatureIDValue()))
	return nil
}

func (f FeatureIDExpression) MarshalYAML() (interface{}, error) {
	return FeatureID(f).MarshalYAML()
}

func (f *FeatureIDExpression) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return (*FeatureID)(f).UnmarshalYAML(unmarshal)
}

func (f *FeatureIDExpression) Clone() Expression {
	clone := *f
	return Expression{AnyExpression: &clone}
}

func (f FeatureIDExpression) Literal() interface{} {
	return FeatureID(f)
}

func (f FeatureIDExpression) Equal(other AnyExpression) bool {
	if ff, ok := other.Clone().AnyExpression.(*FeatureIDExpression); ok {
		return f == *ff
	}
	return false
}

func NewFeatureIDExpression(id FeatureID) Expression {
	l := FeatureIDExpression(id)
	return Expression{AnyExpression: &l}
}

type TagExpression Tag

func (t *TagExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_TagValue{
					TagValue: &pb.TagProto{
						Key:   Tag(*t).Key,
						Value: Tag(*t).Value.String(),
					},
				},
			},
		},
	}, nil
}

func (t *TagExpression) FromProto(node *pb.NodeProto) error {
	tt := node.GetLiteral().GetTagValue()
	*t = TagExpression(Tag{Key: tt.Key, Value: String(tt.Value)})
	return nil
}

func (t TagExpression) MarshalYAML() (interface{}, error) {
	return Tag(t).MarshalYAML()
}

func (t *TagExpression) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return (*Tag)(t).UnmarshalYAML(unmarshal)
}

func (t *TagExpression) Clone() Expression {
	clone := *t
	return Expression{AnyExpression: &clone}
}

func (t TagExpression) Literal() interface{} {
	return Tag(t)
}

func (t TagExpression) Equal(other AnyExpression) bool {
	if tt, ok := other.(*TagExpression); ok {
		return t == *tt
	}
	return false
}

type QueryExpression struct {
	Query Query
}

func (q *QueryExpression) ToProto() (*pb.NodeProto, error) {
	if p, err := q.Query.ToProto(); err == nil {
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_QueryValue{
						QueryValue: p,
					},
				},
			},
		}, nil
	} else {
		return nil, err
	}
}

func (q *QueryExpression) FromProto(node *pb.NodeProto) error {
	var err error
	q.Query, err = NewQueryFromProto(node.GetLiteral().GetQueryValue())
	return err
}

func (q QueryExpression) MarshalYAML() (interface{}, error) {
	return queryYAML{Query: q.Query}.MarshalYAML()
}

func (q *QueryExpression) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var y queryYAML
	err := y.UnmarshalYAML(unmarshal)
	if err == nil {
		q.Query = y.Query
	}
	return err
}

func (q *QueryExpression) Clone() Expression {
	return Expression{AnyExpression: &QueryExpression{Query: q.Query}}
}

func (q QueryExpression) Literal() interface{} {
	return q.Query
}

func (q QueryExpression) Equal(other AnyExpression) bool {
	if qq, ok := other.(*QueryExpression); ok {
		return q.Query.Equal(qq.Query)
	}
	return false
}

func NewQueryExpression(query Query) Expression {
	return Expression{AnyExpression: &QueryExpression{
		Query: query,
	}}
}

type GeoJSONExpression struct {
	GeoJSON geojson.GeoJSON
}

func (g *GeoJSONExpression) ToProto() (*pb.NodeProto, error) {
	marshalled, err := json.Marshal(g.GeoJSON)
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
}

func (g *GeoJSONExpression) FromProto(node *pb.NodeProto) error {
	panic("Unimplemented")
}

func (g *GeoJSONExpression) Clone() Expression {
	clone := *g
	return Expression{AnyExpression: &clone}
}

func (g GeoJSONExpression) Literal() interface{} {
	return g.GeoJSON
}

func (g GeoJSONExpression) Equal(other AnyExpression) bool {
	if gg, ok := other.(*GeoJSONExpression); ok {
		if b, err := json.Marshal(g.GeoJSON); err == nil {
			if bb, err := json.Marshal(gg.GeoJSON); err == nil {
				return bytes.Equal(b, bb)
			}
		}
	}
	return false
}

type RouteExpression Route

func (r *RouteExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_RouteValue{
					RouteValue: NewProtoFromRoute(Route(*r)),
				},
			},
		},
	}, nil
}

func (r *RouteExpression) FromProto(node *pb.NodeProto) error {
	*r = RouteExpression(NewRouteFromProto(node.GetLiteral().GetRouteValue()))
	return nil
}

func (r *RouteExpression) Clone() Expression {
	clone := *r
	return Expression{AnyExpression: &clone}
}

func (r *RouteExpression) Literal() interface{} {
	return Route(*r)
}

func (r *RouteExpression) Equal(other AnyExpression) bool {
	if rr, ok := other.(*RouteExpression); ok {
		if r.Origin != rr.Origin {
			return false
		}
		if len(r.Steps) != len(rr.Steps) {
			return false
		}
		for i := range r.Steps {
			if r.Steps[i] != rr.Steps[i] {
				return false
			}
		}
	}
	return true
}

type FeatureExpression struct {
	Feature
}

func (f *FeatureExpression) ToProto() (*pb.NodeProto, error) {
	if p, err := NewProtoFromFeature(f.Feature); err == nil {
		return &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_FeatureValue{
						FeatureValue: p,
					},
				},
			},
		}, nil
	} else {
		return nil, err
	}
}

func (f *FeatureExpression) FromProto(node *pb.NodeProto) error {
	// TODO: Remove Feature from the external API, and instead
	// just use FeatureIDs
	return errors.New("Can't import features from protos")
}

func (f *FeatureExpression) MarshalYAML() (interface{}, error) {
	// TODO: Remove Feature from the external API, and instead
	// just use FeatureIDs
	return nil, errors.New("Can't export features as YAML")
}

func (f *FeatureExpression) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// TODO: Remove Feature from the external API, and instead
	// just use FeatureIDs
	return errors.New("Can't import features from YAML")
}

func (f *FeatureExpression) Clone() Expression {
	clone := *f
	return Expression{AnyExpression: &clone}
}

func (f *FeatureExpression) Literal() interface{} {
	return f.Feature
}

func (f FeatureExpression) Equal(other AnyExpression) bool {
	if ff, ok := other.(*FeatureExpression); ok {
		return f.FeatureID() == ff.FeatureID()
	}
	return false
}

type PointExpression s2.LatLng

func (p *PointExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_PointValue{
					PointValue: NewPointProtoFromS2LatLng(s2.LatLng(*p)),
				},
			},
		},
	}, nil
}

func (p *PointExpression) FromProto(node *pb.NodeProto) error {
	point := node.GetLiteral().GetPointValue()
	*p = PointExpression(PointProtoToS2LatLng(point))
	return nil
}

func (p PointExpression) MarshalYAML() (interface{}, error) {
	return fmt.Sprintf("%f, %f", p.Lat.Degrees(), p.Lng.Degrees()), nil
}

func (p *PointExpression) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	ll, err := LatLngFromString(s)
	if err != nil {
		return err
	}
	*p = PointExpression(ll)
	return nil
}

func (p *PointExpression) Clone() Expression {
	clone := *p
	return Expression{AnyExpression: &clone}
}

func (p PointExpression) Literal() interface{} {
	return GeometryFromLatLng(s2.LatLng(p))
}

func (p PointExpression) Equal(other AnyExpression) bool {
	if pp, ok := other.(*PointExpression); ok {
		return p == *pp
	}
	return false
}

func NewPointExpressionFromLatLng(ll s2.LatLng) Expression {
	p := PointExpression(ll)
	return Expression{AnyExpression: &p}
}

type PathExpression struct {
	Path Path
}

func (p *PathExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_PathValue{
					PathValue: NewPolylineProto(p.Path.Polyline()),
				},
			},
		},
	}, nil
}

func (p *PathExpression) FromProto(node *pb.NodeProto) error {
	path := node.GetLiteral().GetPathValue()
	p.Path = PathFromS2Points(*PolylineProtoToS2Polyline(path))
	return nil
}

func (p PathExpression) MarshalYAML() (interface{}, error) {
	points := make([]PointExpression, p.Path.Len())
	for i := 0; i < len(points); i++ {
		points[i] = PointExpression(s2.LatLngFromPoint(p.Path.Point(i)))
	}
	return points, nil
}

func (p *PathExpression) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var expressions []PointExpression
	err := unmarshal(&expressions)
	if err == nil {
		points := make([]s2.Point, len(expressions))
		for i, p := range expressions {
			points[i] = s2.PointFromLatLng(s2.LatLng(p))
		}
		p.Path = PathFromS2Points(points)
	}
	return err
}

func (p *PathExpression) Clone() Expression {
	clone := *p
	return Expression{AnyExpression: &clone}
}

func (p PathExpression) Literal() interface{} {
	return p.Path
}

func (p PathExpression) Equal(other AnyExpression) bool {
	if pp, ok := other.(*PathExpression); ok {
		return geometry.PolylineEqual(p.Path.Polyline(), pp.Path.Polyline())
	}
	return false
}

type AreaExpression struct {
	Area Area
}

func (a *AreaExpression) ToProto() (*pb.NodeProto, error) {
	polygons := make(geometry.MultiPolygon, a.Area.Len())
	for i := 0; i < a.Area.Len(); i++ {
		polygons[i] = a.Area.Polygon(i)
	}
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_AreaValue{
					AreaValue: NewMultiPolygonProto(polygons),
				},
			},
		},
	}, nil
}

func (a *AreaExpression) FromProto(node *pb.NodeProto) error {
	m := MultiPolygonProtoToS2MultiPolygon(node.GetLiteral().GetAreaValue())
	a.Area = AreaFromS2Polygons(m)
	return nil
}

func (a *AreaExpression) Clone() Expression {
	clone := *a
	return Expression{AnyExpression: &clone}
}

func (a AreaExpression) Literal() interface{} {
	return a.Area
}

func (a AreaExpression) Equal(other AnyExpression) bool {
	if aa, ok := other.(*AreaExpression); ok {
		return geometry.MultiPolygonEqual(a.Area.MultiPolygon(), aa.Area.MultiPolygon())
	}
	return false
}

type NilExpression struct{}

func (_ NilExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_NilValue{},
			},
		},
	}, nil
}

func (_ NilExpression) FromProto(node *pb.NodeProto) error {
	return nil
}

func (n NilExpression) Clone() Expression {
	return Expression{AnyExpression: n}
}

func (n NilExpression) Literal() interface{} {
	return nil
}

func (n NilExpression) Equal(other AnyExpression) bool {
	_, ok := other.(NilExpression)
	return ok
}

type CollectionExpression struct {
	UntypedCollection
}

func (c *CollectionExpression) ToProto() (*pb.NodeProto, error) {
	i := c.UntypedCollection.BeginUntyped()
	keys := make([]*pb.LiteralNodeProto, 0)
	values := make([]*pb.LiteralNodeProto, 0)
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		if k, err := FromLiteral(i.Key()); err == nil {
			if p, err := k.ToProto(); err == nil {
				keys = append(keys, p.GetLiteral())
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
		if v, err := FromLiteral(i.Value()); err == nil {
			if p, err := v.ToProto(); err == nil {
				values = append(values, p.GetLiteral())
			} else {
				return nil, err
			}
		} else {
			return nil, err
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

func (c *CollectionExpression) FromProto(node *pb.NodeProto) error {
	p := node.GetLiteral().GetCollectionValue()
	collection := ArrayCollection[interface{}, interface{}]{
		Keys:   make([]interface{}, len(p.Keys)),
		Values: make([]interface{}, len(p.Values)),
	}
	if len(collection.Keys) != len(collection.Values) {
		return fmt.Errorf("Number of keys doesn't match the number of values: %d vs %d", len(collection.Keys), len(collection.Values))
	}
	for i := range p.Keys {
		var k Literal
		n := &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: p.Keys[i],
			},
		}
		if err := k.FromProto(n); err != nil {
			return err
		}
		collection.Keys[i] = k.Literal()
		var v Literal
		n = &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: p.Values[i],
			},
		}
		if err := v.FromProto(n); err != nil {
			return err
		}
		collection.Values[i] = k.Literal()
	}
	c.UntypedCollection = Collection[any, any]{
		AnyCollection: collection,
	}
	return nil
}

func (c CollectionExpression) Equal(other AnyExpression) bool {
	if cc, ok := other.(*CollectionExpression); ok {
		i := [2]Iterator[any, any]{c.BeginUntyped(), cc.BeginUntyped()}
		for {
			var ok [2]bool
			var err [2]error
			for j := range ok {
				ok[j], err[j] = i[j].Next()
			}
			if err[0] != nil || err[1] != nil {
				return false
			}
			if ok[0] != ok[1] {
				return false
			}
			if !ok[0] {
				break
			}
			if !LiteralEqual(i[0].Key(), i[1].Key()) || !LiteralEqual(i[0].Value(), i[1].Value()) {
				return false
			}
		}
		return true
	}
	return false
}

func NewCollectionExpression(c UntypedCollection) Expression {
	return Expression{AnyExpression: &CollectionExpression{UntypedCollection: c}}
}

type collectionYAML struct {
	Keys   []Literal
	Values []Literal
}

func (c CollectionExpression) MarshalYAML() (interface{}, error) {
	i := c.UntypedCollection.BeginUntyped()
	y := make([][2]Literal, 0)
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		var pair [2]Literal
		if pair[0], err = FromLiteral(i.Key()); err == nil {
			pair[1], err = FromLiteral(i.Value())
		}
		y = append(y, pair)
	}
	return y, nil
}

func (c *CollectionExpression) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var y [][2]Literal
	err := unmarshal(&y)
	if err == nil {
		collection := ArrayCollection[any, any]{
			Keys:   make([]any, len(y)),
			Values: make([]any, len(y)),
		}
		for i := range y {
			collection.Keys[i] = y[i][0].Literal()
			collection.Values[i] = y[i][1].Literal()
		}
		c.UntypedCollection = collection.Collection()
	}
	return err
}

func (c *CollectionExpression) Clone() Expression {
	return Expression{
		AnyExpression: &CollectionExpression{
			UntypedCollection: c.UntypedCollection,
		},
	}
}

func (c CollectionExpression) Literal() interface{} {
	return c.UntypedCollection
}

type CallExpression struct {
	Function  Expression   `yaml:",omitempty"`
	Args      []Expression `yaml:",omitempty"`
	Pipelined bool         `yaml:",omitempty"`
}

func (c *CallExpression) ToProto() (*pb.NodeProto, error) {
	var err error
	args := make([]*pb.NodeProto, len(c.Args))
	for i, arg := range c.Args {
		if args[i], err = arg.ToProto(); err != nil {
			return nil, err
		}
	}
	if f, err := c.Function.ToProto(); err == nil {
		return &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function:  f,
					Args:      args,
					Pipelined: c.Pipelined,
				},
			},
		}, nil
	} else {
		return nil, err
	}
}

func (c *CallExpression) FromProto(node *pb.NodeProto) error {
	call := node.GetCall()
	if err := c.Function.FromProto(call.Function); err != nil {
		return err
	}
	c.Args = make([]Expression, len(call.Args))
	for i, arg := range call.Args {
		if err := c.Args[i].FromProto(arg); err != nil {
			return err
		}
	}
	c.Pipelined = call.Pipelined
	return nil
}

func (c *CallExpression) Clone() Expression {
	args := make([]Expression, len(c.Args))
	for i, arg := range c.Args {
		args[i] = arg.Clone()
	}
	return Expression{AnyExpression: &CallExpression{
		Function: c.Function.Clone(),
		Args:     args,
	}}
}

func (c *CallExpression) Equal(other AnyExpression) bool {
	if cc, ok := other.(*CallExpression); ok {
		if !c.Function.Equal(cc.Function) {
			return false
		}
		if len(c.Args) != len(cc.Args) {
			return false
		}
		for i := range c.Args {
			if !c.Args[i].Equal(cc.Args[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func NewCallExpression(function Expression, args []Expression) Expression {
	return Expression{AnyExpression: &CallExpression{Function: function, Args: args}}
}

type LambdaExpression struct {
	Args       []string   `yaml:",omitempty"`
	Expression Expression `yaml:",omitempty"`
}

func (l *LambdaExpression) ToProto() (*pb.NodeProto, error) {
	if e, err := l.Expression.ToProto(); err == nil {
		return &pb.NodeProto{
			Node: &pb.NodeProto_Lambda_{
				Lambda_: &pb.LambdaNodeProto{
					Args: l.Args,
					Node: e,
				},
			},
		}, nil
	} else {
		return nil, err
	}
}

func (l *LambdaExpression) FromProto(node *pb.NodeProto) error {
	lambda := node.GetLambda_()
	l.Args = lambda.Args
	return l.Expression.FromProto(lambda.Node)
}

func (l *LambdaExpression) Clone() Expression {
	names := make([]string, len(l.Args))
	for i, name := range l.Args {
		names[i] = name
	}
	return Expression{AnyExpression: &LambdaExpression{
		Args:       names,
		Expression: l.Expression.Clone(),
	}}
}

func (l *LambdaExpression) Equal(other AnyExpression) bool {
	if ll, ok := other.(*LambdaExpression); ok {
		if !l.Expression.Equal(ll.Expression) {
			return false
		}
		if len(l.Args) != len(ll.Args) {
			return false
		}
		for i := range l.Args {
			if l.Args[i] != ll.Args[i] {
				return false
			}
		}
		return true
	}
	return false
}

func NewLambdaExpression(args []string, e Expression) Expression {
	return Expression{AnyExpression: &LambdaExpression{Args: args, Expression: e}}
}
