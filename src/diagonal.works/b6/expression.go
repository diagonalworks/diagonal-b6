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

type ValueType int

const (
	// TODO(mari): rename / implement a type for each expression
	ValueTypeString ValueType = iota
	ValueTypePoint
	ValueTypeValues
	ValueTypeFeatureID
	ValueTypeInvalid
)

type AnyExpression interface {
	ToProto() (*pb.NodeProto, error)
	Equal(other AnyExpression) bool
	Clone() Expression
	String() string
	ValueType() ValueType
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
		case IntExpression:
			return int(e), nil
		case FloatExpression:
			return float64(e), nil
		case StringExpression:
			return string(e), nil
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
		e.AnyExpression = IntExpression(v)
		return nil
	case float64:
		e.AnyExpression = FloatExpression(v)
		return nil
	case string:
		e.AnyExpression = StringExpression(v)
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

func ExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	e := Expression{}
	switch n := node.Node.(type) {
	case *pb.NodeProto_Symbol:
		return SymbolExpressionFromProto(node)
	case *pb.NodeProto_Call:
		return CallExpressionFromProto(node)
	case *pb.NodeProto_Lambda_:
		return LambdaExpressionFromProto(node)
	case *pb.NodeProto_Literal:
		switch n.Literal.Value.(type) {
		case *pb.LiteralNodeProto_IntValue:
			return IntExpressionFromProto(node)
		case *pb.LiteralNodeProto_FloatValue:
			return FloatExpressionFromProto(node)
		case *pb.LiteralNodeProto_BoolValue:
			return BoolExpressionFromProto(node)
		case *pb.LiteralNodeProto_StringValue:
			return StringExpressionFromProto(node)
		case *pb.LiteralNodeProto_FeatureIDValue:
			return FeatureIDExpressionFromProto(node)
		case *pb.LiteralNodeProto_TagValue:
			return TagExpressionFromProto(node)
		case *pb.LiteralNodeProto_PointValue:
			return PointExpressionFromProto(node)
		case *pb.LiteralNodeProto_PathValue:
			return PathExpressionFromProto(node)
		case *pb.LiteralNodeProto_AreaValue:
			return AreaExpressionFromProto(node)
		case *pb.LiteralNodeProto_QueryValue:
			return QueryExpressionFromProto(node)
		case *pb.LiteralNodeProto_NilValue:
			return NilExpressionFromProto(node)
		case *pb.LiteralNodeProto_GeoJSONValue:
			return GeoJSONExpressionFromProto(node)
		case *pb.LiteralNodeProto_RouteValue:
			return RouteExpressionFromProto(node)
		case *pb.LiteralNodeProto_CollectionValue:
			return CollectionExpressionFromProto(node)
		default:
			return Expression{}, fmt.Errorf("Can't convert %T from literal proto", n.Literal.Value)
		}
	default:
		return Expression{}, fmt.Errorf("Can't convert expression from proto %T", node.Node)
	}

	e.Name = node.Name
	e.Begin = int(node.Begin)
	e.End = int(node.End)
	return e, nil
}

func (e Expression) Clone() Expression {
	clone := e
	clone.AnyExpression = e.AnyExpression.Clone().AnyExpression
	return clone
}

func (e Expression) Equal(other Expression) bool {
	if e.AnyExpression == nil {
		return other.AnyExpression == nil
	}
	return e.AnyExpression.Equal(other.AnyExpression)
}

type AnyLiteral interface {
	AnyExpression
	Literal() interface{}
}

type Literal struct {
	AnyLiteral
}

func (l Literal) ToProto() (*pb.NodeProto, error) {
	return l.AnyLiteral.ToProto()
}

func LiteralFromProto(node *pb.NodeProto) (Expression, error) {
	var e Expression
	if e, err := ExpressionFromProto(node); err != nil {
		return e, err
	}
	if literal, ok := e.AnyExpression.(AnyLiteral); ok {
		return Expression{AnyExpression: literal}, nil
	} else {
		return Expression{}, fmt.Errorf("Can't convert literal from proto %T", node.Node)
	}
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
		return Literal{AnyLiteral: IntExpression(l)}, nil
	case IntNumber:
		return Literal{AnyLiteral: FloatExpression(int(l))}, nil
	case float64:
		return Literal{AnyLiteral: FloatExpression(l)}, nil
	case FloatNumber: // TODO(mari): rethink number interface + it doesnt make sense to allow it as a literal here
		return Literal{AnyLiteral: FloatExpression(float64(l))}, nil
	case bool:
		return Literal{AnyLiteral: BoolExpression(l)}, nil
	case string:
		return Literal{AnyLiteral: StringExpression(l)}, nil
	case FeatureID:
		return Literal{AnyLiteral: FeatureIDExpression(l)}, nil
	case AreaID:
		return FromLiteral(l.FeatureID())
	case RelationID:
		return FromLiteral(l.FeatureID())
	case CollectionID:
		return FromLiteral(l.FeatureID())
	case ExpressionID:
		return FromLiteral(l.FeatureID())
	case Tag:
		return Literal{AnyLiteral: TagExpression(l)}, nil
	case Feature:
		return Literal{AnyLiteral: FeatureExpression{Feature: l}}, nil
	case Area:
		return Literal{AnyLiteral: AreaExpression{Area: l}}, nil
	case Geometry:
		switch l.GeometryType() {
		case GeometryTypePoint:
			return Literal{AnyLiteral: PointExpression(s2.LatLngFromPoint(l.Point()))}, nil
		case GeometryTypePath:
			return Literal{AnyLiteral: PathExpression{Path: l}}, nil
		}
	case s2.LatLng:
		return Literal{AnyLiteral: PointExpression(l)}, nil
	case geojson.GeoJSON:
		return Literal{AnyLiteral: GeoJSONExpression{GeoJSON: l}}, nil
	case Route:
		return Literal{AnyLiteral: RouteExpression(l)}, nil
	case UntypedCollection:
		return Literal{AnyLiteral: CollectionExpression{UntypedCollection: l}}, nil
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

func (s SymbolExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Symbol{
			Symbol: string(s),
		},
	}, nil
}

func SymbolExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	return Expression{AnyExpression: SymbolExpression(node.GetSymbol())}, nil
}

func (s SymbolExpression) Clone() Expression {
	return Expression{AnyExpression: s}
}

func (s SymbolExpression) String() string {
	return string(s)
}

func (s SymbolExpression) Equal(other AnyExpression) bool {
	if ss, ok := other.(SymbolExpression); ok {
		return s == ss
	}
	return false
}

func (SymbolExpression) ValueType() ValueType {
	return ValueTypeInvalid
}

func NewSymbolExpression(symbol string) Expression {
	return Expression{AnyExpression: SymbolExpression(symbol)}
}

type IntExpression int

func (i IntExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_IntValue{
					IntValue: int64(i),
				},
			},
		},
	}, nil
}

func IntExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	return Expression{AnyExpression: IntExpression(node.GetLiteral().GetIntValue())}, nil
}

func (i IntExpression) Clone() Expression {
	return Expression{AnyExpression: i}
}

func (i IntExpression) Literal() interface{} {
	return int(i)
}

func (i IntExpression) String() string {
	return fmt.Sprintf("%d", i)
}

func (i IntExpression) Equal(other AnyExpression) bool {
	if ii, ok := other.(IntExpression); ok {
		return i == ii
	}
	return false
}

func (IntExpression) ValueType() ValueType {
	return ValueTypeInvalid
}

func NewIntExpression(value int) Expression {
	return Expression{AnyExpression: IntExpression(value)}
}

type FloatExpression float64

func (f FloatExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_FloatValue{
					FloatValue: float64(f),
				},
			},
		},
	}, nil
}

func FloatExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	return Expression{AnyExpression: FloatExpression(node.GetLiteral().GetFloatValue())}, nil
}

func (f FloatExpression) Clone() Expression {
	return Expression{AnyExpression: f}
}

func (f FloatExpression) Literal() interface{} {
	return float64(f)
}

func (f FloatExpression) String() string {
	return fmt.Sprintf("%f", f)
}

func (f FloatExpression) Equal(other AnyExpression) bool {
	if ff, ok := other.(FloatExpression); ok {
		return f == ff
	}
	return false
}

func (FloatExpression) ValueType() ValueType {
	return ValueTypeInvalid
}

func NewFloatExpression(value float64) Expression {
	return Expression{AnyExpression: FloatExpression(value)}
}

type BoolExpression bool

func (b BoolExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_BoolValue{
					BoolValue: bool(b),
				},
			},
		},
	}, nil
}

func BoolExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	return Expression{AnyExpression: BoolExpression(node.GetLiteral().GetBoolValue())}, nil
}

func (b BoolExpression) Clone() Expression {
	return Expression{AnyExpression: b}
}

func (b BoolExpression) Literal() interface{} {
	return bool(b)
}

func (b BoolExpression) Equal(other AnyExpression) bool {
	if bb, ok := other.(BoolExpression); ok {
		return b == bb
	}
	return false
}

func (b BoolExpression) String() string {
	if bool(b) {
		return "true"
	}
	return "false"
}

func (BoolExpression) ValueType() ValueType {
	return ValueTypeInvalid
}

type StringExpression string

func (s StringExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_StringValue{
					StringValue: string(s),
				},
			},
		},
	}, nil
}

func StringExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	return Expression{AnyExpression: StringExpression(node.GetLiteral().GetStringValue())}, nil
}

func (s StringExpression) Clone() Expression {
	return Expression{AnyExpression: s}
}

func (s StringExpression) Literal() interface{} {
	return string(s)
}

func (s StringExpression) Equal(other AnyExpression) bool {
	if ss, ok := other.(StringExpression); ok {
		return s == ss
	}
	return false
}

func (s StringExpression) String() string {
	return string(s)
}

func (StringExpression) ValueType() ValueType {
	return ValueTypeString
}

func NewStringExpression(s string) Expression {
	return Expression{AnyExpression: StringExpression(s)}
}

type FeatureIDExpression FeatureID

func (f FeatureIDExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_FeatureIDValue{
					FeatureIDValue: NewProtoFromFeatureID(FeatureID(f)),
				},
			},
		},
	}, nil
}

func FeatureIDExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	return Expression{AnyExpression: FeatureIDExpression(NewFeatureIDFromProto(node.GetLiteral().GetFeatureIDValue()))}, nil
}

func (f FeatureIDExpression) MarshalYAML() (interface{}, error) {
	return FeatureID(f).MarshalYAML()
}

func (f *FeatureIDExpression) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return (*FeatureID)(f).UnmarshalYAML(unmarshal)
}

func (f FeatureIDExpression) Clone() Expression {
	return Expression{AnyExpression: f}
}

func (f FeatureIDExpression) Literal() interface{} {
	return FeatureID(f)
}

func (f FeatureIDExpression) Equal(other AnyExpression) bool {
	if ff, ok := other.(FeatureIDExpression); ok {
		return f == ff
	}
	return false
}

func (f FeatureIDExpression) String() string {
	return FeatureID(f).String()
}

func (FeatureIDExpression) ValueType() ValueType {
	return ValueTypeFeatureID
}

func (f FeatureIDExpression) Source() FeatureID {
	return FeatureID(f)
}

func (FeatureIDExpression) Index() (int, error) {
	return -1, fmt.Errorf("index not available")
}

func (FeatureIDExpression) SetIndex(i int) {}

func NewFeatureIDExpression(id FeatureID) Expression {
	return Expression{AnyExpression: FeatureIDExpression(id)}
}

type TagExpression Tag

func (t TagExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_TagValue{
					TagValue: &pb.TagProto{
						Key:   Tag(t).Key,
						Value: Tag(t).Value.String(),
					},
				},
			},
		},
	}, nil
}

func TagExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	tt := node.GetLiteral().GetTagValue()
	return Expression{AnyExpression: TagExpression(Tag{Key: tt.Key, Value: StringExpression(tt.Value)})}, nil // TODO(mari): tag expression value should support all expression types
}

func (t TagExpression) MarshalYAML() (interface{}, error) {
	return Tag(t).MarshalYAML()
}

func (t *TagExpression) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return (*Tag)(t).UnmarshalYAML(unmarshal)
}

func (t TagExpression) Clone() Expression {
	return Expression{AnyExpression: t}
}

func (t TagExpression) Literal() interface{} {
	return Tag(t)
}

func (t TagExpression) Equal(other AnyExpression) bool {
	if tt, ok := other.(TagExpression); ok {
		return t == tt
	}
	return false
}

func (TagExpression) ValueType() ValueType {
	return ValueTypeInvalid
}

type QueryExpression struct {
	Query Query
}

func (q QueryExpression) ToProto() (*pb.NodeProto, error) {
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

func QueryExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	q, err := NewQueryFromProto(node.GetLiteral().GetQueryValue())
	return Expression{AnyExpression: QueryExpression{q}}, err
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

func (q QueryExpression) Clone() Expression {
	return Expression{AnyExpression: q}
}

func (q QueryExpression) Literal() interface{} {
	return q.Query
}

func (q QueryExpression) Equal(other AnyExpression) bool {
	if qq, ok := other.(QueryExpression); ok {
		return q.Query.Equal(qq.Query)
	}
	return false
}

func (q QueryExpression) String() string {
	return q.Query.String()
}

func (QueryExpression) ValueType() ValueType {
	return ValueTypeInvalid
}

func NewQueryExpression(query Query) Expression {
	return Expression{AnyExpression: QueryExpression{Query: query}}
}

type GeoJSONExpression struct {
	GeoJSON geojson.GeoJSON
}

func (g GeoJSONExpression) ToProto() (*pb.NodeProto, error) {
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

func GeoJSONExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	panic("Unimplemented")
}

func (g GeoJSONExpression) Clone() Expression {
	return Expression{AnyExpression: g}
}

func (g GeoJSONExpression) Literal() interface{} {
	return g.GeoJSON
}

func (g GeoJSONExpression) Equal(other AnyExpression) bool {
	if gg, ok := other.(GeoJSONExpression); ok {
		if b, err := json.Marshal(g.GeoJSON); err == nil {
			if bb, err := json.Marshal(gg.GeoJSON); err == nil {
				return bytes.Equal(b, bb)
			}
		}
	}
	return false
}

func (g GeoJSONExpression) String() string {
	return "x-geojson"
}

func (GeoJSONExpression) ValueType() ValueType {
	return ValueTypeInvalid
}

type RouteExpression Route

func (r RouteExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_RouteValue{
					RouteValue: NewProtoFromRoute(Route(r)),
				},
			},
		},
	}, nil
}

func RouteExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	return Expression{AnyExpression: RouteExpression(NewRouteFromProto(node.GetLiteral().GetRouteValue()))}, nil
}

func (r RouteExpression) Clone() Expression {
	return Expression{AnyExpression: r}
}

func (r RouteExpression) Literal() interface{} {
	return Route(r)
}

func (r RouteExpression) Equal(other AnyExpression) bool {
	if rr, ok := other.(RouteExpression); ok {
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

func (r RouteExpression) String() string {
	return "x-route"
}

func (RouteExpression) ValueType() ValueType {
	return ValueTypeInvalid
}

type FeatureExpression struct {
	Feature
}

func (f FeatureExpression) ToProto() (*pb.NodeProto, error) {
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

func FeatureExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	// TODO: Remove Feature from the external API, and instead
	// just use FeatureIDs
	return Expression{}, errors.New("Can't import features from protos")
}

func (f FeatureExpression) MarshalYAML() (interface{}, error) {
	// TODO: Remove Feature from the external API, and instead
	// just use FeatureIDs
	return nil, errors.New("Can't export features as YAML")
}

func (f FeatureExpression) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// TODO: Remove Feature from the external API, and instead
	// just use FeatureIDs
	return errors.New("Can't import features from YAML")
}

func (f FeatureExpression) Clone() Expression {
	return Expression{AnyExpression: f}
}

func (f FeatureExpression) Literal() interface{} {
	return f.Feature
}

func (f FeatureExpression) Equal(other AnyExpression) bool {
	if ff, ok := other.(FeatureExpression); ok {
		return f.FeatureID() == ff.FeatureID()
	}
	return false
}

func (f FeatureExpression) String() string {
	return "x-feature"
}

func (FeatureExpression) ValueType() ValueType {
	return ValueTypeInvalid
}

type PointExpression s2.LatLng

func (p PointExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_PointValue{
					PointValue: NewPointProtoFromS2LatLng(s2.LatLng(p)),
				},
			},
		},
	}, nil
}

func PointExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	return Expression{AnyExpression: PointExpression(PointProtoToS2LatLng(node.GetLiteral().GetPointValue()))}, nil
}

func (p PointExpression) MarshalYAML() (interface{}, error) {
	return LatLngToString(s2.LatLng(p)), nil
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

func (p PointExpression) Clone() Expression {
	return Expression{AnyExpression: p}
}

func (p PointExpression) Literal() interface{} {
	return GeometryFromLatLng(s2.LatLng(p))
}

func (p PointExpression) Equal(other AnyExpression) bool {
	if pp, ok := other.(PointExpression); ok {
		return p == pp
	}
	return false
}

func (p PointExpression) String() string {
	return LatLngToString(s2.LatLng(p))
}

func (PointExpression) ValueType() ValueType {
	return ValueTypePoint
}

func NewPointExpressionFromLatLng(ll s2.LatLng) Expression {
	return Expression{AnyExpression: PointExpression(ll)}
}

type PathExpression struct {
	Path Geometry
}

func (p PathExpression) ToProto() (*pb.NodeProto, error) {
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

func PathExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	return Expression{AnyExpression: PathExpression{GeometryFromPoints(*PolylineProtoToS2Polyline(node.GetLiteral().GetPathValue()))}}, nil
}

func (p PathExpression) MarshalYAML() (interface{}, error) {
	points := make([]PointExpression, p.Path.GeometryLen())
	for i := 0; i < len(points); i++ {
		points[i] = PointExpression(s2.LatLngFromPoint(p.Path.PointAt(i)))
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
		p.Path = GeometryFromPoints(points)
	}
	return err
}

func (p PathExpression) Clone() Expression {
	return Expression{AnyExpression: p}
}

func (p PathExpression) Literal() interface{} {
	return p.Path
}

func (p PathExpression) Equal(other AnyExpression) bool {
	if pp, ok := other.(PathExpression); ok {
		return geometry.PolylineEqual(p.Path.Polyline(), pp.Path.Polyline())
	}
	return false
}

func (p PathExpression) String() string {
	return "x-path" // TODO(mari): implement all string representations
}

func (PathExpression) ValueType() ValueType {
	return ValueTypeInvalid
}

type AreaExpression struct {
	Area Area
}

func (a AreaExpression) ToProto() (*pb.NodeProto, error) {
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

func AreaExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	return Expression{AnyExpression: AreaExpression{AreaFromS2Polygons(MultiPolygonProtoToS2MultiPolygon(node.GetLiteral().GetAreaValue()))}}, nil
}

func (a AreaExpression) Clone() Expression {
	return Expression{AnyExpression: a}
}

func (a AreaExpression) Literal() interface{} {
	return a.Area
}

func (a AreaExpression) Equal(other AnyExpression) bool {
	if aa, ok := other.(AreaExpression); ok {
		return geometry.MultiPolygonEqual(a.Area.MultiPolygon(), aa.Area.MultiPolygon())
	}
	return false
}

func (a AreaExpression) String() string {
	return "x-area"
}

func (AreaExpression) ValueType() ValueType {
	return ValueTypeInvalid
}

type NilExpression struct{}

func (NilExpression) ToProto() (*pb.NodeProto, error) {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_NilValue{},
			},
		},
	}, nil
}

func NilExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	return Expression{}, nil
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

func (n NilExpression) String() string {
	return "x-nil"
}

func (NilExpression) ValueType() ValueType {
	return ValueTypeInvalid
}

type CollectionExpression struct {
	UntypedCollection
}

func (c CollectionExpression) ToProto() (*pb.NodeProto, error) {
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

func CollectionExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	p := node.GetLiteral().GetCollectionValue()
	collection := ArrayCollection[interface{}, interface{}]{
		Keys:   make([]interface{}, len(p.Keys)),
		Values: make([]interface{}, len(p.Values)),
	}
	if len(collection.Keys) != len(collection.Values) {
		return Expression{}, fmt.Errorf("Number of keys doesn't match the number of values: %d vs %d", len(collection.Keys), len(collection.Values))
	}
	for i := range p.Keys {
		var k Literal
		n := &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: p.Keys[i],
			},
		}
		expression, err := LiteralFromProto(n)
		if err != nil {
			return Expression{}, err
		}
		k.AnyLiteral = expression.AnyExpression.(AnyLiteral)

		collection.Keys[i] = k.Literal()
		var v Literal
		n = &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: p.Values[i],
			},
		}
		expression, err = LiteralFromProto(n)
		if err != nil {
			return Expression{}, err
		}
		v.AnyLiteral = expression.AnyExpression.(AnyLiteral)
		collection.Values[i] = v.Literal()
	}
	return Expression{AnyExpression: CollectionExpression{UntypedCollection: Collection[any, any]{AnyCollection: collection}}}, nil
}

func (c CollectionExpression) Equal(other AnyExpression) bool {
	if cc, ok := other.(CollectionExpression); ok {
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

func (c CollectionExpression) String() string {
	s := "{"
	i := c.UntypedCollection.BeginUntyped()
	for {
		ok, err := i.Next()
		if err != nil {
			return "x-broken-collection"
		} else if !ok {
			break
		}
		if len(s) > 1 {
			s += ", "
		}
		s += i.KeyExpression().String() + ": " + i.ValueExpression().String()
	}
	return s + "}"
}

func NewCollectionExpression(c UntypedCollection) Expression {
	return Expression{AnyExpression: CollectionExpression{UntypedCollection: c}}
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

func (c CollectionExpression) Clone() Expression {
	return Expression{AnyExpression: CollectionExpression{UntypedCollection: c.UntypedCollection}}
}

func (c CollectionExpression) Literal() interface{} {
	return c.UntypedCollection
}

func (CollectionExpression) ValueType() ValueType {
	return ValueTypeInvalid
}

type CallExpression struct {
	Function  Expression   `yaml:",omitempty"`
	Args      []Expression `yaml:",omitempty"`
	Pipelined bool         `yaml:",omitempty"`
}

func (c CallExpression) ToProto() (*pb.NodeProto, error) {
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

func CallExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	c := CallExpression{}
	call := node.GetCall()
	var err error
	if c.Function, err = ExpressionFromProto(call.Function); err != nil {
		return Expression{}, err
	}
	c.Args = make([]Expression, len(call.Args))
	for i, arg := range call.Args {
		if c.Args[i], err = ExpressionFromProto(arg); err != nil {
			return Expression{}, err
		}
	}
	c.Pipelined = call.Pipelined
	return Expression{AnyExpression: c}, nil
}

func (c CallExpression) Clone() Expression {
	args := make([]Expression, len(c.Args))
	for i, arg := range c.Args {
		args[i] = arg.Clone()
	}
	return Expression{AnyExpression: CallExpression{
		Function: c.Function.Clone(),
		Args:     args,
	}}
}

func (c CallExpression) Equal(other AnyExpression) bool {
	if cc, ok := other.(CallExpression); ok {
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

func (c CallExpression) String() string {
	s := ""
	if _, ok := c.Function.AnyExpression.(CallExpression); ok {
		s += "(" + c.Function.String() + ")"
	} else {
		s += c.Function.String()
	}
	if len(c.Args) > 0 {
		s += " "
		for i, arg := range c.Args {
			if i > 0 {
				s += " "
			}
			if _, ok := arg.AnyExpression.(CallExpression); ok {
				s += "(" + arg.String() + ")"
			} else {
				s += arg.String()
			}
		}
	}
	return s
}

func (CallExpression) ValueType() ValueType {
	return ValueTypeInvalid
}

func NewCallExpression(function Expression, args []Expression) Expression {
	return Expression{AnyExpression: CallExpression{Function: function, Args: args}}
}

type LambdaExpression struct {
	Args       []string   `yaml:",omitempty"`
	Expression Expression `yaml:",omitempty"`
}

func (l LambdaExpression) ToProto() (*pb.NodeProto, error) {
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

func LambdaExpressionFromProto(node *pb.NodeProto) (Expression, error) {
	lambda := node.GetLambda_()
	l := LambdaExpression{}
	l.Args = lambda.Args
	var err error
	l.Expression, err = ExpressionFromProto(lambda.Node)
	if err != nil {
		return Expression{}, err
	}

	return Expression{AnyExpression: l}, nil
}

func (l LambdaExpression) Clone() Expression {
	names := make([]string, len(l.Args))
	for i, name := range l.Args {
		names[i] = name
	}
	return Expression{AnyExpression: LambdaExpression{
		Args:       names,
		Expression: l.Expression.Clone(),
	}}
}

func (l LambdaExpression) Equal(other AnyExpression) bool {
	if ll, ok := other.(LambdaExpression); ok {
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

func (l LambdaExpression) String() string {
	s := "{"
	for i := range l.Args {
		if i > 0 {
			s += ", "
		}
		s += l.Args[i]
	}
	return s + " -> " + l.Expression.String() + "}"

}

func (LambdaExpression) ValueType() ValueType {
	return ValueTypeInvalid
}

func NewLambdaExpression(args []string, e Expression) Expression {
	return Expression{AnyExpression: LambdaExpression{Args: args, Expression: e}}
}
