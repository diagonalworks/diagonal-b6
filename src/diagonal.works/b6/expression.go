package b6

import (
	"fmt"

	pb "diagonal.works/b6/proto"
)

type AnyExpression interface {
	ToProto() *pb.NodeProto
	FromProto(node *pb.NodeProto)
	Clone() Expression
}

type Expression struct {
	AnyExpression
}

func (e Expression) MarshalYAML() (interface{}, error) {
	// Fast track types that are handled natively by YAML
	switch e := e.AnyExpression.(type) {
	case *IntExpression:
		return int(*e), nil
	case *StringExpression:
		return string(*e), nil
	}
	var kind string
	switch e.AnyExpression.(type) {
	case *SymbolExpression:
		kind = "symbol"
	case *IntExpression:
		kind = "int"
	case *StringExpression:
		kind = "string"
	case *CallExpression:
		kind = "call"
	case *LambdaExpression:
		kind = "lambda"
	case *FeatureIDExpression:
		kind = "id"
	default:
		return nil, fmt.Errorf("Can't marshal expression yaml: %T", e.AnyExpression)
	}
	return map[string]interface{}{
		kind: e.AnyExpression,
	}, nil
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
	case string:
		s := StringExpression(v)
		e.AnyExpression = &s
		return nil
	}
	var y struct {
		Symbol *SymbolExpression
		Int    *IntExpression
		String *StringExpression
		Call   *CallExpression
		Lambda *LambdaExpression
		ID     *FeatureIDExpression
	}
	if err := unmarshal(&y); err != nil {
		return err
	}
	if y.Symbol != nil {
		e.AnyExpression = y.Symbol
	} else if y.Int != nil {
		e.AnyExpression = y.Int
	} else if y.String != nil {
		e.AnyExpression = y.String
	} else if y.Call != nil {
		e.AnyExpression = y.Call
	} else if y.Lambda != nil {
		e.AnyExpression = y.Lambda
	} else if y.ID != nil {
		e.AnyExpression = y.ID
	} else {
		return fmt.Errorf("Can't unmarshal expression yaml: %+v", y)
	}
	return nil
}

func (e *Expression) ToProto() *pb.NodeProto {
	return e.AnyExpression.ToProto()
}

func (e *Expression) FromProto(node *pb.NodeProto) {
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
		case *pb.LiteralNodeProto_StringValue:
			e.AnyExpression = new(StringExpression)
		default:
			panic(fmt.Sprintf("Can't convert %T from literal proto", n.Literal.Value))
		}
	default:
		panic(fmt.Sprintf("Can't convert expression from proto %T", node.Node))
	}
	e.AnyExpression.FromProto(node)
}

type AnyLiteral interface {
	AnyExpression
	Literal() interface{}
}

type Literal struct {
	AnyLiteral
}

func (l *Literal) ToProto() *pb.NodeProto {
	return l.AnyLiteral.ToProto()
}

func (l *Literal) FromProto(node *pb.NodeProto) {
	var e Expression
	e.FromProto(node)
	if literal, ok := e.AnyExpression.(AnyLiteral); ok {
		l.AnyLiteral = literal
	} else {
		panic(fmt.Sprintf("Can't convert literal from proto %T", node.Node))
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

func FromLiteral(l interface{}) Literal {
	switch l := l.(type) {
	case int:
		i := IntExpression(l)
		return Literal{AnyLiteral: &i}
	case string:
		s := StringExpression(l)
		return Literal{AnyLiteral: &s}
	case FeatureID:
		id := FeatureIDExpression(l)
		return Literal{AnyLiteral: &id}
	}
	panic(fmt.Sprintf("Can't make literal from %T", l))

}

type SymbolExpression string

func (s *SymbolExpression) ToProto() *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Symbol{
			Symbol: string(*s),
		},
	}
}

func (s *SymbolExpression) FromProto(node *pb.NodeProto) {
	*s = SymbolExpression(node.GetSymbol())
}

func (s *SymbolExpression) Clone() Expression {
	clone := *s
	return Expression{AnyExpression: &clone}
}

func NewSymbolExpression(symbol string) Expression {
	s := SymbolExpression(symbol)
	return Expression{AnyExpression: &s}
}

type IntExpression int

func (i *IntExpression) ToProto() *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_IntValue{
					IntValue: int64(*i),
				},
			},
		},
	}
}

func (i *IntExpression) FromProto(node *pb.NodeProto) {
	*i = IntExpression(node.GetLiteral().GetIntValue())
}

func (i *IntExpression) Clone() Expression {
	clone := *i
	return Expression{AnyExpression: &clone}
}

func (i IntExpression) Literal() interface{} {
	return int(i)
}

func NewIntExpression(value int) Expression {
	i := IntExpression(value)
	return Expression{AnyExpression: &i}
}

type StringExpression string

func (s *StringExpression) ToProto() *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_StringValue{
					StringValue: string(*s),
				},
			},
		},
	}
}

func (s *StringExpression) FromProto(node *pb.NodeProto) {
	*s = StringExpression(node.GetLiteral().GetStringValue())
}

func (s *StringExpression) Clone() Expression {
	clone := *s
	return Expression{AnyExpression: &clone}
}

func (s StringExpression) Literal() interface{} {
	return string(s)
}

type FeatureIDExpression FeatureID

func (f *FeatureIDExpression) ToProto() *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_FeatureIDValue{
					FeatureIDValue: NewProtoFromFeatureID(FeatureID(*f)),
				},
			},
		},
	}
}

func (f *FeatureIDExpression) FromProto(node *pb.NodeProto) {
	*f = FeatureIDExpression(NewFeatureIDFromProto(node.GetLiteral().GetFeatureIDValue()))
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

type CallExpression struct {
	Function  Expression   `yaml:",omitempty"`
	Args      []Expression `yaml:",omitempty"`
	Pipelined bool         `yaml:",omitempty"`
}

func (c *CallExpression) ToProto() *pb.NodeProto {
	args := make([]*pb.NodeProto, len(c.Args))
	for i, arg := range c.Args {
		args[i] = arg.ToProto()
	}
	return &pb.NodeProto{
		Node: &pb.NodeProto_Call{
			Call: &pb.CallNodeProto{
				Function:  c.Function.ToProto(),
				Args:      args,
				Pipelined: c.Pipelined,
			},
		},
	}
}

func (c *CallExpression) FromProto(node *pb.NodeProto) {
	call := node.GetCall()
	c.Function.FromProto(call.Function)
	c.Args = make([]Expression, len(call.Args))
	for i, arg := range call.Args {
		c.Args[i].FromProto(arg)
	}
	c.Pipelined = call.Pipelined
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

type LambdaExpression struct {
	Args       []string   `yaml:",omitempty"`
	Expression Expression `yaml:",omitempty"`
}

func (l *LambdaExpression) ToProto() *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Lambda_{
			Lambda_: &pb.LambdaNodeProto{
				Args: l.Args,
				Node: l.Expression.ToProto(),
			},
		},
	}
}

func (l *LambdaExpression) FromProto(node *pb.NodeProto) {
	lambda := node.GetLambda_()
	l.Args = lambda.Args
	l.Expression.FromProto(lambda.Node)
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
