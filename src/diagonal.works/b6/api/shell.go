package api

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"diagonal.works/b6"
	pb "diagonal.works/b6/proto"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"google.golang.org/protobuf/proto"
)

type SymbolArgCounts interface {
	ArgCount(symbol string) (int, bool)
}

type Symbols interface {
	SymbolArgCounts
	Function(symbol string) (reflect.Value, bool)
}

type FunctionSymbols map[string]interface{}

func (fs FunctionSymbols) Function(symbol string) (reflect.Value, bool) {
	if f, ok := fs[symbol]; ok {
		return reflect.ValueOf(f), ok
	}
	return reflect.Value{}, false
}

func (fs FunctionSymbols) ArgCount(symbol string) (int, bool) {
	if f, ok := fs[symbol]; ok {
		t := reflect.TypeOf(f)
		return t.NumIn() - 1, true
	}
	return 0, false
}

type FunctionArgCounts map[string]int

func (fs FunctionArgCounts) ArgCount(symbol string) (int, bool) {
	n, ok := fs[symbol]
	return n, ok
}

type NamespaceAlias struct {
	Prefix     string
	Namespace  b6.Namespace
	Type       b6.FeatureType
	FromString func(a *NamespaceAlias, token string) (b6.FeatureID, error)
	ToString   func(a *NamespaceAlias, id b6.FeatureID) string
}

var aliases = []NamespaceAlias{
	{
		Prefix:     "/n/",
		Namespace:  b6.NamespaceOSMNode,
		Type:       b6.FeatureTypePoint,
		FromString: idFromUint64,
		ToString:   idToUint64,
	},
	{
		Prefix:     "/w/",
		Namespace:  b6.NamespaceOSMWay,
		Type:       b6.FeatureTypePath,
		FromString: idFromUint64,
		ToString:   idToUint64,
	},
	{
		Prefix:     "/a/",
		Namespace:  b6.NamespaceOSMWay,
		Type:       b6.FeatureTypeArea,
		FromString: idFromUint64,
		ToString:   idToUint64,
	},
	{
		Prefix:     "/r/",
		Namespace:  b6.NamespaceOSMRelation,
		Type:       b6.FeatureTypeRelation,
		FromString: idFromUint64,
		ToString:   idToUint64,
	},
	{
		Prefix:     "/uk/ons/",
		Namespace:  b6.NamespaceUKONSBoundaries,
		Type:       b6.FeatureTypeArea,
		FromString: idFromUKONS,
		ToString:   idToUKONS,
	},
	{
		Prefix:     "/gb/codepoint/",
		Namespace:  b6.NamespaceGBCodePoint,
		Type:       b6.FeatureTypePoint,
		FromString: idFromGBCodePoint,
		ToString:   idToGBCodePoint,
	},
	{
		Prefix:     "/gb/uprn/",
		Namespace:  b6.NamespaceGBUPRN,
		Type:       b6.FeatureTypePoint,
		FromString: idFromUint64,
		ToString:   idToUint64,
	},
}

func idFromUint64(a *NamespaceAlias, token string) (b6.FeatureID, error) {
	id := b6.FeatureIDInvalid
	v, err := strconv.ParseUint(token[len(a.Prefix):], 10, 64)
	if err == nil {
		id = b6.FeatureID{Type: a.Type, Namespace: a.Namespace, Value: v}
	}
	return id, err
}

func idToUint64(a *NamespaceAlias, id b6.FeatureID) string {
	return a.Prefix + strconv.FormatUint(id.Value, 10)
}

func idFromUKONS(a *NamespaceAlias, token string) (b6.FeatureID, error) {
	id := b6.FeatureIDInvalid
	parts := strings.Split(token[len(a.Prefix):], "/")
	if len(parts) == 2 {
		year, err := strconv.Atoi(parts[0])
		if err == nil {
			id = b6.FeatureIDFromUKONSCode(parts[1], year, a.Type)
		}
	}
	var err error
	if !id.IsValid() {
		err = fmt.Errorf("expected, for example, %s2011/E01000953", a.Prefix)
	}
	return id, err
}

func idToUKONS(a *NamespaceAlias, id b6.FeatureID) string {
	code, year, ok := b6.UKONSCodeFromFeatureID(id)
	if ok {
		return fmt.Sprintf("%s%d/%s", a.Prefix, year, code)
	}
	return b6.FeatureIDInvalid.String()
}

func idFromGBCodePoint(a *NamespaceAlias, token string) (b6.FeatureID, error) {
	return b6.PointIDFromGBPostcode(token[len(a.Prefix):]).FeatureID(), nil
}

func idToGBCodePoint(a *NamespaceAlias, id b6.FeatureID) string {
	postcode, _ := b6.PostcodeFromPointID(id.ToPointID())
	return a.Prefix + strings.ToLower(postcode)
}

func ParseFeatureIDToken(token string) (b6.FeatureID, error) {
	for _, alias := range aliases {
		if strings.HasPrefix(token, alias.Prefix) {
			return alias.FromString(&alias, token)
		}
	}
	id := b6.FeatureIDFromString(token[1:])
	var err error
	if !id.IsValid() {
		err = errors.New("expected, for example,  /point/openstreetmap.org/node/3501612811")
	}
	return id, err
}

func UnparseString(s string) string {
	// TODO: Formalise our own string token semantics, rather than kind-of
	// relying on go.
	return fmt.Sprintf("%q", s)
}

func UnparseFeatureID(id b6.FeatureID, abbreviate bool) string {
	if abbreviate {
		for _, alias := range aliases {
			if alias.Namespace == id.Namespace && (alias.Type == b6.FeatureTypeInvalid || alias.Type == id.Type) {
				return alias.ToString(&alias, id)
			}
		}
	}
	return "/" + id.String()
}

type TokenType int

const (
	TokenTypePunctuation TokenType = iota
	TokenTypeLambdaArg
	TokenTypeString
	TokenTypeInt
	TokenTypeFloat
	TokenTypeLatLng
	TokenTypeFeatureID
	TokenTypeSymbol
	TokenTypeQuery
	TokenTypeTag
)

type Token struct {
	Type  TokenType
	Begin int
	End   int
}

type lexer struct {
	Expression string
	LHS        *pb.NodeProto
	Index      int
	Top        *pb.NodeProto
	Err        error
}

const eof = 0

func (l *lexer) Lex(yylval *yySymType) int {
	for l.Index < len(l.Expression) {
		r, w := utf8.DecodeRuneInString(l.Expression[l.Index:])
		if !unicode.IsSpace(r) {
			break
		}
		l.Index += w
	}

	if l.Index >= len(l.Expression) {
		return eof
	}
	switch c := l.Expression[l.Index]; c {
	case ',', '(', ')', '|', '>', '{', '}', '[', ']', '=', '&':
		l.Index++
		return int(c)
	case '"':
		return l.lexStringLiteral(yylval)
	case '/':
		return l.lexFeatureIDLiteral(yylval)
	case '#', '@':
		return l.lexTagKeyLiteral(yylval)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', '.':
		if c == '-' && l.Index+1 < len(l.Expression) && l.Expression[l.Index+1] == '>' {
			l.Index += 2
			return ARROW
		}
		return l.lexNumericLiteral(yylval)
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		return l.lexSymbolLiteral(yylval)
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		return l.lexSymbolLiteral(yylval)
	}
	l.Err = fmt.Errorf("Bad token %q", l.Expression[l.Index:])
	return eof
}

func (l *lexer) consume(end int) (*pb.NodeProto, string) {
	begin := l.Index
	l.Index = end
	return &pb.NodeProto{Begin: int32(begin), End: int32(end)}, l.Expression[begin:end]
}

func (l *lexer) lexStringLiteral(yylval *yySymType) int {
	i := l.Index + 1
	for i < len(l.Expression) {
		r, w := utf8.DecodeRuneInString(l.Expression[i:])
		i += w
		if r == '"' {
			node, token := l.consume(i)
			node.Node = &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_StringValue{
						StringValue: token[1 : len(token)-1],
					},
				},
			}
			yylval.node = node
			return STRING
		}
	}
	l.Err = fmt.Errorf("Unterminated string constant")
	return eof
}

func (l *lexer) lexFeatureIDLiteral(yylval *yySymType) int {
	i := l.Index
	for i < len(l.Expression) {
		r, w := utf8.DecodeRuneInString(l.Expression[i:])
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '-' || r == '/') {
			break
		}
		i += w
	}
	node, token := l.consume(i)
	id, err := ParseFeatureIDToken(token)
	if err == nil {
		node.Node = &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_FeatureIDValue{
					FeatureIDValue: b6.NewProtoFromFeatureID(id),
				},
			},
		}
		yylval.node = node
	} else {
		l.Err = err
	}
	return FEATURE_ID
}

func (l *lexer) lexTagKeyLiteral(yylval *yySymType) int {
	i := l.Index + 1
	for i < len(l.Expression) {
		r, w := utf8.DecodeRuneInString(l.Expression[i:])
		if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			break
		}
		i += w
	}
	node, token := l.consume(i)
	node.Node = &pb.NodeProto_Literal{
		Literal: &pb.LiteralNodeProto{
			Value: &pb.LiteralNodeProto_StringValue{
				StringValue: token,
			},
		},
	}
	yylval.node = node
	return TAG_KEY
}

func (l *lexer) lexNumericLiteral(yylval *yySymType) int {
	i := l.Index
	decimal := false
	for i < len(l.Expression) {
		r, w := utf8.DecodeRuneInString(l.Expression[i:])
		if r == '-' {
			if i != l.Index {
				l.Err = fmt.Errorf("Unexpected -")
				return eof
			}
		} else if r == '.' {
			if decimal {
				l.Err = fmt.Errorf("Unexpected .")
				return eof
			}
			decimal = true

		} else if !unicode.IsDigit(r) {
			break
		}
		i += w
	}
	node, token := l.consume(i)
	if !decimal {
		v, err := strconv.Atoi(token)
		if err != nil {
			l.Err = fmt.Errorf("%q: %s", token, err)
			return eof
		}
		node.Node = &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_IntValue{
					IntValue: int64(v),
				},
			},
		}
		yylval.node = node
		return INT
	} else {
		v, err := strconv.ParseFloat(token, 64)
		if err != nil {
			l.Err = fmt.Errorf("%q: %s", token, err)
			return eof
		}
		node.Node = &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_FloatValue{
					FloatValue: v,
				},
			},
		}
		yylval.node = node
		return FLOAT
	}
}

func isValidSymbolRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == ':' || r == '_'
}

func (l *lexer) lexSymbolLiteral(yylval *yySymType) int {
	i := l.Index
	for i < len(l.Expression) {
		r, w := utf8.DecodeRuneInString(l.Expression[i:])
		if !isValidSymbolRune(r) {
			break
		}
		i += w
	}
	node, token := l.consume(i)
	node.Node = &pb.NodeProto_Symbol{
		Symbol: token,
	}
	yylval.node = node
	return SYMBOL
}

func (l *lexer) Error(s string) {
	if l.Err == nil {
		l.Err = fmt.Errorf("%s at %q", s, l.Expression[l.Index:])
	}
}

func reduceLatLng(lat *pb.NodeProto, lng *pb.NodeProto, l *lexer) *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_PointValue{
					PointValue: b6.NewPointProto(
						lat.GetLiteral().GetFloatValue(),
						lng.GetLiteral().GetFloatValue(),
					),
				},
			},
		},
		Begin: lat.Begin,
		End:   lng.End,
	}
}

func nodeToString(node *pb.NodeProto) string {
	if s, ok := node.Node.(*pb.NodeProto_Symbol); ok {
		return s.Symbol
	} else if l, ok := node.Node.(*pb.NodeProto_Literal); ok {
		if s, ok := l.Literal.Value.(*pb.LiteralNodeProto_StringValue); ok {
			return s.StringValue
		}
	}
	panic("Not a symbol or string")
}

func reduceTag(key *pb.NodeProto, value *pb.NodeProto, l *lexer) *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_TagValue{
					TagValue: &pb.TagProto{
						Key:   nodeToString(key),
						Value: nodeToString(value),
					},
				},
			},
		},
		Begin: key.Begin,
		End:   value.End,
	}
}

func reduceCall(symbol *pb.NodeProto, l *lexer) *pb.NodeProto {
	return reduceCallWithArgs(symbol, []*pb.NodeProto{}, l)
}

func reduceCallWithArgs(symbol *pb.NodeProto, args []*pb.NodeProto, l *lexer) *pb.NodeProto {
	for _, arg := range args {
		if arg == nil {
			return nil
		}
	}

	begin, end := symbol.Begin, symbol.End
	if len(args) > 0 {
		end = args[len(args)-1].End
	}

	return &pb.NodeProto{
		Node: &pb.NodeProto_Call{
			Call: &pb.CallNodeProto{
				Function: symbol,
				Args:     args,
			},
		},
		Begin: begin,
		End:   end,
	}
}

func reduceArgs(args []*pb.NodeProto, arg *pb.NodeProto) []*pb.NodeProto {
	return append(args, arg)
}

func reduceArg(arg *pb.NodeProto) []*pb.NodeProto {
	return []*pb.NodeProto{arg}
}

func reduceRootCall(root *pb.NodeProto, l *lexer) *pb.NodeProto {
	if l.LHS != nil {
		root = Pipeline(l.LHS, root)
		l.LHS = nil
	}
	return root
}

// Return a node that calls right with left as an argument
func Pipeline(left *pb.NodeProto, right *pb.NodeProto) *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Call{
			Call: &pb.CallNodeProto{
				Function:  right,
				Args:      []*pb.NodeProto{left},
				Pipelined: true,
			},
		},
		Begin: left.Begin,
		End:   right.End,
	}
}

func reducePipeline(left *pb.NodeProto, right *pb.NodeProto, l *lexer) *pb.NodeProto {
	if left == nil || right == nil {
		return nil
	}
	if l.LHS != nil {
		left = Pipeline(l.LHS, left)
		l.LHS = nil
	}
	return Pipeline(left, right)
}

func reduceLambda(symbols []*pb.NodeProto, e *pb.NodeProto) *pb.NodeProto {
	args := make([]string, len(symbols))
	for i, s := range symbols {
		args[i] = s.GetSymbol()
	}
	begin := e.Begin
	if len(symbols) > 0 {
		begin = symbols[0].Begin
	}
	return &pb.NodeProto{
		Node: &pb.NodeProto_Lambda_{
			Lambda_: &pb.LambdaNodeProto{
				Args: args,
				Node: e,
			},
		},
		Begin: begin,
		End:   e.End,
	}
}

func reduceLambdaWithoutArgs(e *pb.NodeProto) *pb.NodeProto {
	return reduceLambda([]*pb.NodeProto{}, e)
}

func reduceSymbolsSymbol(s *pb.NodeProto) []*pb.NodeProto {
	return []*pb.NodeProto{s}
}

func reduceSymbolsSymbols(ss []*pb.NodeProto, s *pb.NodeProto) []*pb.NodeProto {
	reduced := make([]*pb.NodeProto, len(ss)+1)
	for i, s := range ss {
		reduced[i] = s
	}
	reduced[len(reduced)-1] = s
	return reduced
}

func reduceTagKey(key *pb.NodeProto) *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_QueryValue{
					QueryValue: &pb.QueryProto{
						Query: &pb.QueryProto_Keyed{
							Keyed: nodeToString(key),
						},
					},
				},
			},
		},
		Begin: key.Begin,
		End:   key.End,
	}
}

func reduceTagKeyValue(key *pb.NodeProto, value *pb.NodeProto) *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_QueryValue{
					QueryValue: &pb.QueryProto{
						Query: &pb.QueryProto_Tagged{
							Tagged: &pb.TagProto{
								Key:   nodeToString(key),
								Value: nodeToString(value),
							},
						},
					},
				},
			},
		},
		Begin: key.Begin,
		End:   value.End,
	}
}

func reduceAnd(a *pb.NodeProto, b *pb.NodeProto) *pb.NodeProto {
	aq := a.GetLiteral().GetQueryValue()
	bq := b.GetLiteral().GetQueryValue()
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_QueryValue{
					QueryValue: &pb.QueryProto{
						Query: &pb.QueryProto_Intersection{
							Intersection: &pb.QueriesProto{
								Queries: []*pb.QueryProto{aq, bq},
							},
						},
					},
				},
			},
		},
		Begin: a.Begin,
		End:   b.End,
	}
}

func reduceOr(a *pb.NodeProto, b *pb.NodeProto) *pb.NodeProto {
	aq := a.GetLiteral().GetQueryValue()
	bq := b.GetLiteral().GetQueryValue()
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_QueryValue{
					QueryValue: &pb.QueryProto{
						Query: &pb.QueryProto_Union{
							Union: &pb.QueriesProto{
								Queries: []*pb.QueryProto{aq, bq},
							},
						},
					},
				},
			},
		},
		Begin: a.Begin,
		End:   b.End,
	}
}

func ParseExpression(expression string) (*pb.NodeProto, error) {
	yyErrorVerbose = true
	l := lexer{Expression: expression}
	yyParse(&l)
	if l.Top == nil {
		return nil, l.Err
	}
	return l.Top, l.Err
}

func ParseExpressionWithLHS(expression string, lhs *pb.NodeProto) (*pb.NodeProto, error) {
	yyErrorVerbose = true
	l := lexer{Expression: expression, LHS: lhs}
	yyParse(&l)
	if l.Top == nil {
		return nil, l.Err
	}
	return l.Top, l.Err
}

type byBeginThenLength []*pb.NodeProto

func (b byBeginThenLength) Len() int      { return len(b) }
func (b byBeginThenLength) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byBeginThenLength) Less(i, j int) bool {
	if b[i].Begin == b[j].Begin {
		return (b[i].End - b[i].Begin) > (b[j].End - b[j].Begin)
	}
	return b[i].Begin < b[j].Begin
}

func OrderTokens(n *pb.NodeProto) []*pb.NodeProto {
	tokens := []*pb.NodeProto{}
	queue := []*pb.NodeProto{n}
	for len(queue) > 0 {
		n := queue[len(queue)-1]
		queue = queue[0 : len(queue)-1]
		if c, ok := n.Node.(*pb.NodeProto_Call); ok {
			queue = append(queue, c.Call.Function)
			queue = append(queue, c.Call.Args...)
		} else if l, ok := n.Node.(*pb.NodeProto_Lambda_); ok {
			queue = append(queue, l.Lambda_.Node)
		}
		if n.End > n.Begin {
			tokens = append(tokens, n)
		}
	}
	sort.Sort(byBeginThenLength(tokens))
	filtered := make([]*pb.NodeProto, 0, len(tokens)/2)
	for _, t := range tokens {
		if len(filtered) > 0 && filtered[len(filtered)-1].End > t.Begin {
			filtered[len(filtered)-1] = t
		} else {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func Simplify(n *pb.NodeProto, functions SymbolArgCounts) *pb.NodeProto {
	if n == nil {
		return nil
	}
	if _, ok := n.Node.(*pb.NodeProto_Call); ok {
		return simplifyCall(n, functions)
	} else if _, ok := n.Node.(*pb.NodeProto_Lambda_); ok {
		return simplifyLambda(n, functions)
	}
	return n
}

func simplifyCall(n *pb.NodeProto, functions SymbolArgCounts) *pb.NodeProto {
	call := n.Node.(*pb.NodeProto_Call).Call
	call.Function = Simplify(call.Function, functions)
	for i, arg := range call.Args {
		call.Args[i] = Simplify(arg, functions)
	}
	if n, ok := simplifyCallWithNoArguments(n, functions); ok {
		return n
	}
	return n
}

func simplifyCallWithNoArguments(n *pb.NodeProto, functions SymbolArgCounts) (*pb.NodeProto, bool) {
	// Calling a function that expects arguments with no arguments is
	// semantically equivilent to just using that function.
	call := n.Node.(*pb.NodeProto_Call).Call
	if len(call.Args) == 0 {
		if symbol, ok := call.Function.Node.(*pb.NodeProto_Symbol); ok {
			if n, ok := functions.ArgCount(symbol.Symbol); ok && n > 0 {
				return call.Function, true
			}
		}
	}
	return n, false
}

func simplifyLambda(n *pb.NodeProto, functions SymbolArgCounts) *pb.NodeProto {
	// '{a -> area a}' is semantically equivalent to 'area'
	lambda := n.Node.(*pb.NodeProto_Lambda_).Lambda_
	if c, ok := lambda.Node.Node.(*pb.NodeProto_Call); ok && len(lambda.Args) > 0 {
		call := c.Call
		i := 0
		for i < len(lambda.Args) && i < len(call.Args) {
			if s, ok := call.Args[i].Node.(*pb.NodeProto_Symbol); ok {
				if s.Symbol != lambda.Args[i] {
					break
				}
			} else {
				break
			}
			i++
		}
		if i > 0 {
			if i == len(call.Args) {
				return call.Function
			}
			return simplifyCall(&pb.NodeProto{
				Node: &pb.NodeProto_Call{
					Call: &pb.CallNodeProto{
						Function: call.Function,
						Args:     call.Args[i:len(call.Args)],
					},
				},
				Begin: n.Begin,
				End:   n.End,
			}, functions)
		}
	}
	return n
}

func EscapeTagKey(v string) string {
	if v == "" {
		return ""
	}
	escape := (v[0] < 'a' && v[0] > 'z') && (v[0] < 'A' && v[0] > 'Z') && v[0] != '#' && v[0] != '@'
	if !escape {
		for _, r := range v[1:] {
			if escape = !isValidSymbolRune(r); escape {
				break
			}
		}
	}
	if escape {
		// TODO: Actually esape string literals properly
		v = fmt.Sprintf("%q", v)
	}
	return v
}

func EscapeTagValue(v string) string {
	if v == "" {
		return ""
	}
	escape := (v[0] < 'a' && v[0] > 'z') && (v[0] < 'A' && v[0] > 'Z')
	if !escape {
		for _, r := range v[1:] {
			if escape = !isValidSymbolRune(r); escape {
				break
			}
		}
	}
	if escape {
		// TODO: Actually esape string literals properly
		v = fmt.Sprintf("%q", v)
	}
	return v
}

func TagToExpression(t b6.Tag) string {
	return EscapeTagKey(t.Key) + "=" + EscapeTagValue(t.Value)
}

func UnparseQuery(q b6.Query) (string, bool) {
	if expression, ok := unparseQuery(q); ok {
		return "[" + expression + "]", true
	}
	return "", false
}

func unparseQuery(q b6.Query) (string, bool) {
	// TODO: Escape query literals properly
	switch q := q.(type) {
	case b6.Tagged:
		return TagToExpression(b6.Tag(q)), true
	case b6.Keyed:
		return "[" + q.Key + "]", true
	case b6.Intersection:
		qs := make([]string, len(q))
		for i := range q {
			var ok bool
			if qs[i], ok = unparseQuery(q[i]); !ok {
				return "", false
			}
		}
		return strings.Join(qs, " & "), true
	case b6.Union:
		qs := make([]string, len(q))
		for i := range q {
			var ok bool
			if qs[i], ok = unparseQuery(q[i]); !ok {
				return "", false
			}
		}
		return strings.Join(qs, " | "), true
	}
	return "", false
}

func UnparseNode(n *pb.NodeProto) (string, bool) {
	return unparseNode(n, true)
}

func unparseNode(n *pb.NodeProto, top bool) (string, bool) {
	switch n := n.Node.(type) {
	case *pb.NodeProto_Symbol:
		return n.Symbol, true
	case *pb.NodeProto_Call:
		if n.Call.Pipelined {
			return unparsePipelinedCall(n.Call, top)
		} else {
			return unparseCall(n.Call, top)
		}
	case *pb.NodeProto_Literal:
		return unparseLiteral(n.Literal)
	}
	return "", false
}

func unparsePipelinedCall(c *pb.CallNodeProto, top bool) (string, bool) {
	lhs, ok := unparseNode(c.Args[0], true)
	if !ok {
		return "", false
	}
	var cc pb.CallNodeProto
	proto.Merge(&cc, c)
	cc.Args = cc.Args[1:]
	rhs, ok := unparseCall(&cc, true)
	if !ok {
		return "", false
	}
	if top {
		return lhs + " | " + rhs, true
	} else {
		return "(" + lhs + " | " + rhs + ")", true
	}
}

func unparseCall(c *pb.CallNodeProto, top bool) (string, bool) {
	if len(c.Args) == 0 {
		return unparseNode(c.Function, top)
	}
	parts := []string{}
	if part, ok := unparseNode(c.Function, false); ok {
		parts = append(parts, part)
	} else {
		return "", false
	}
	for _, arg := range c.Args {
		if part, ok := unparseNode(arg, false); ok {
			parts = append(parts, part)
		} else {
			return "", false
		}
	}
	joined := strings.Join(parts, " ")
	if top {
		return joined, true
	} else {
		return "(" + joined + ")", true
	}
}

func unparseLiteral(l *pb.LiteralNodeProto) (string, bool) {
	switch l := l.Value.(type) {
	case *pb.LiteralNodeProto_StringValue:
		return UnparseString(l.StringValue), true
	case *pb.LiteralNodeProto_IntValue:
		return fmt.Sprintf("%d", l.IntValue), true
	case *pb.LiteralNodeProto_FloatValue:
		return fmt.Sprintf("%.2f", l.FloatValue), true
	case *pb.LiteralNodeProto_TagValue:
		return TagToExpression(b6.Tag{Key: l.TagValue.Key, Value: l.TagValue.Value}), true
	case *pb.LiteralNodeProto_FeatureIDValue:
		id := b6.NewFeatureIDFromProto(l.FeatureIDValue)
		return UnparseFeatureID(id, true), true
	case *pb.LiteralNodeProto_PointValue:
		ll := s2.LatLng{Lat: s1.Angle(l.PointValue.LatE7) * s1.E7, Lng: s1.Angle(l.PointValue.LngE7) * s1.E7}
		return fmt.Sprintf("%f, %f", ll.Lat.Degrees(), ll.Lng.Degrees()), true
	case *pb.LiteralNodeProto_QueryValue:
		if q, err := b6.NewQueryFromProto(l.QueryValue); err == nil {
			return UnparseQuery(q)
		} else {
			return "", false
		}
	default:
		return fmt.Sprintf("(broken-value \"%v\")", l), true
	}
}
