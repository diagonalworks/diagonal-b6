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
	"github.com/golang/geo/s2"
)

type SymbolArgCounts interface {
	ArgCount(symbol b6.SymbolExpression) (int, bool)
	IsVariadic(symbol b6.SymbolExpression) (bool, bool)
}

type Symbols interface {
	SymbolArgCounts
	Function(symbol b6.SymbolExpression) (reflect.Value, bool)
}

type FunctionSymbols map[string]interface{}

func (fs FunctionSymbols) Function(symbol b6.SymbolExpression) (reflect.Value, bool) {
	if f, ok := fs[symbol.String()]; ok {
		return reflect.ValueOf(f), ok
	}
	return reflect.Value{}, false
}

func (fs FunctionSymbols) ArgCount(symbol b6.SymbolExpression) (int, bool) {
	if f, ok := fs[symbol.String()]; ok {
		t := reflect.TypeOf(f)
		return t.NumIn() - 1, true
	}
	return 0, false
}

func (fs FunctionSymbols) IsVariadic(symbol b6.SymbolExpression) (bool, bool) {
	if f, ok := fs[symbol.String()]; ok {
		t := reflect.TypeOf(f)
		return t.IsVariadic(), true
	}
	return false, false
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
	LHS        b6.Expression
	Index      int
	Top        b6.Expression
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
	case ',', '(', ')', '|', '>', '{', '}', '[', ']', '=', '&', ':':
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
	l.Err = fmt.Errorf("bad token %q", l.Expression[l.Index:])
	return eof
}

func (l *lexer) consume(end int) (b6.Expression, string) {
	begin := l.Index
	l.Index = end
	return b6.Expression{Begin: begin, End: end}, l.Expression[begin:end]
}

func (l *lexer) lexStringLiteral(yylval *yySymType) int {
	i := l.Index + 1
	for i < len(l.Expression) {
		r, w := utf8.DecodeRuneInString(l.Expression[i:])
		i += w
		if r == '"' {
			e, token := l.consume(i)
			e.AnyExpression = b6.NewStringExpression(token[1 : len(token)-1]).AnyExpression
			yylval.e = e
			return STRING
		}
	}
	l.Err = fmt.Errorf("unterminated string constant")
	return eof
}

func (l *lexer) lexFeatureIDLiteral(yylval *yySymType) int {
	i := l.Index
	for i < len(l.Expression) {
		r, w := utf8.DecodeRuneInString(l.Expression[i:])
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '-' || r == '/' || r == '_') {
			break
		}
		i += w
	}
	e, token := l.consume(i)
	id, err := ParseFeatureIDToken(token)
	if err == nil {
		eid := b6.FeatureIDExpression(id)
		e.AnyExpression = &eid
		yylval.e = e
	} else {
		l.Err = err
	}
	return FEATURE_ID
}

func (l *lexer) lexTagKeyLiteral(yylval *yySymType) int {
	i := l.Index + 1
	for i < len(l.Expression) {
		r, w := utf8.DecodeRuneInString(l.Expression[i:])
		if !isValidSymbolRune(r) {
			break
		}
		i += w
	}
	e, token := l.consume(i)
	e.AnyExpression = b6.NewStringExpression(token).AnyExpression
	yylval.e = e
	return TAG_KEY
}

func (l *lexer) lexNumericLiteral(yylval *yySymType) int {
	i := l.Index
	decimal := false
	for i < len(l.Expression) {
		r, w := utf8.DecodeRuneInString(l.Expression[i:])
		if r == '-' {
			if i != l.Index {
				l.Err = fmt.Errorf("unexpected -")
				return eof
			}
		} else if r == '.' {
			if decimal {
				l.Err = fmt.Errorf("unexpected .")
				return eof
			}
			decimal = true

		} else if !unicode.IsDigit(r) {
			break
		}
		i += w
	}
	e, token := l.consume(i)
	if !decimal {
		v, err := strconv.Atoi(token)
		if err != nil {
			l.Err = fmt.Errorf("%q: %s", token, err)
			return eof
		}
		e.AnyExpression = b6.NewIntExpression(v).AnyExpression
		yylval.e = e
		return INT
	} else {
		v, err := strconv.ParseFloat(token, 64)
		if err != nil {
			l.Err = fmt.Errorf("%q: %s", token, err)
			return eof
		}
		e.AnyExpression = b6.NewFloatExpression(v).AnyExpression
		yylval.e = e
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
	e, token := l.consume(i)
	e.AnyExpression = b6.NewSymbolExpression(token).AnyExpression
	yylval.e = e
	return SYMBOL
}

func (l *lexer) Error(s string) {
	if l.Err == nil {
		l.Err = fmt.Errorf("%s at %q", s, l.Expression[l.Index:])
	}
}

func reduceLatLng(lat b6.Expression, lng b6.Expression, l *lexer) b6.Expression {
	latf := lat.AnyExpression.(*b6.FloatExpression)
	lngf := lng.AnyExpression.(*b6.FloatExpression)
	ll := s2.LatLngFromDegrees(float64(*latf), float64(*lngf))
	return b6.NewPointExpressionFromLatLng(ll)
}

func expressionToString(expression b6.Expression) string {
	switch e := expression.AnyExpression.(type) {
	case *b6.StringExpression:
		return string(*e)
	case *b6.SymbolExpression:
		return string(*e)
	default:
		panic("Not a symbol or string")
	}
}

func reduceTag(key b6.Expression, value b6.Expression, l *lexer) b6.Expression {
	return b6.Expression{
		AnyExpression: &b6.TagExpression{
			Key:   expressionToString(key),
			Value: expressionToString(value),
		},
		Begin: key.Begin,
		End:   value.End,
	}
}

func reduceCall(symbol b6.Expression, l *lexer) b6.Expression {
	return reduceCallWithArgs(symbol, []b6.Expression{}, l)
}

func reduceCallWithArgs(symbol b6.Expression, args []b6.Expression, l *lexer) b6.Expression {
	for _, arg := range args {
		if arg.AnyExpression == nil {
			return b6.Expression{}
		}
	}

	begin, end := symbol.Begin, symbol.End
	if len(args) > 0 {
		end = args[len(args)-1].End
	}

	return b6.Expression{
		AnyExpression: &b6.CallExpression{
			Function: symbol,
			Args:     args,
		},
		Begin: begin,
		End:   end,
	}
}

func reduceArgs(args []b6.Expression, arg b6.Expression) []b6.Expression {
	return append(args, arg)
}

func reduceArg(arg b6.Expression) []b6.Expression {
	return []b6.Expression{arg}
}

func reduceRootCall(root b6.Expression, l *lexer) b6.Expression {
	if l.LHS.AnyExpression != nil {
		root = Pipeline(l.LHS, root)
		l.LHS.AnyExpression = nil
	}
	return root
}

// Return a node that calls right with left as an argument
func Pipeline(left b6.Expression, right b6.Expression) b6.Expression {
	return b6.Expression{
		AnyExpression: &b6.CallExpression{
			Function:  right,
			Args:      []b6.Expression{left},
			Pipelined: true,
		},
		Begin: left.Begin,
		End:   right.End,
	}
}

func reducePipeline(left b6.Expression, right b6.Expression, l *lexer) b6.Expression {
	if left.AnyExpression == nil || right.AnyExpression == nil {
		return b6.Expression{}
	}
	if l.LHS.AnyExpression != nil {
		left = Pipeline(l.LHS, left)
		l.LHS.AnyExpression = nil
	}
	return Pipeline(left, right)
}

func reduceLambda(symbols []b6.Expression, e b6.Expression) b6.Expression {
	args := make([]string, len(symbols))
	for i, s := range symbols {
		args[i] = s.AnyExpression.(*b6.SymbolExpression).String()
	}
	begin := e.Begin
	if len(symbols) > 0 {
		begin = symbols[0].Begin
	}
	return b6.Expression{
		AnyExpression: &b6.LambdaExpression{
			Args:       args,
			Expression: e,
		},
		Begin: begin,
		End:   e.End,
	}
}

func reduceLambdaWithoutArgs(e b6.Expression) b6.Expression {
	return reduceLambda([]b6.Expression{}, e)
}

func reduceSymbolsSymbol(s b6.Expression) []b6.Expression {
	return []b6.Expression{s}
}

func reduceSymbolsSymbols(ss []b6.Expression, s b6.Expression) []b6.Expression {
	reduced := make([]b6.Expression, len(ss)+1)
	for i, s := range ss {
		reduced[i] = s
	}
	reduced[len(reduced)-1] = s
	return reduced
}

func reduceCollectionItems(collection b6.Expression) b6.Expression {
	// Fill in integer indicies for collection keys if they weren't
	// explicitly specified.
	if call, ok := collection.AnyExpression.(*b6.CallExpression); ok {
		for i, a := range call.Args {
			if pair, ok := a.AnyExpression.(*b6.CallExpression); ok {
				if len(pair.Args) > 0 && pair.Args[0].AnyExpression == nil {
					pair.Args[0].AnyExpression = b6.NewIntExpression(i).AnyExpression
				}
			}
		}
	}
	return collection
}

func reduceCollectionItemsKeyValue(kv b6.Expression) b6.Expression {
	return b6.Expression{
		AnyExpression: &b6.CallExpression{
			Function: b6.NewSymbolExpression("collection"),
			Args:     []b6.Expression{kv},
		},
	}
}

func reduceCollectionItemsItemsKeyValue(items b6.Expression, kv b6.Expression) b6.Expression {
	if call, ok := items.AnyExpression.(*b6.CallExpression); ok {
		call.Args = append(call.Args, kv)
	}
	return items
}

func reduceCollectionKeyValue(key b6.Expression, value b6.Expression) b6.Expression {
	return b6.Expression{
		AnyExpression: &b6.CallExpression{
			Function: b6.NewSymbolExpression("pair"),
			Args:     []b6.Expression{key, value},
		},
	}
}

func reduceCollectionValueWithImplictKey(value b6.Expression) b6.Expression {
	return b6.Expression{
		AnyExpression: &b6.CallExpression{
			Function: b6.NewSymbolExpression("pair"),
			Args:     []b6.Expression{b6.Expression{}, value},
		},
	}
}

func reduceTagKey(key b6.Expression) b6.Expression {
	return b6.Expression{
		AnyExpression: &b6.QueryExpression{
			Query: b6.Keyed{Key: expressionToString(key)},
		},
		Begin: key.Begin,
		End:   key.End,
	}
}

func reduceTagKeyValue(key b6.Expression, value b6.Expression) b6.Expression {
	return b6.Expression{
		AnyExpression: &b6.QueryExpression{
			Query: b6.Tagged{
				Key:   expressionToString(key),
				Value: expressionToString(value),
			},
		},
		Begin: key.Begin,
		End:   value.End,
	}
}

func reduceAnd(a b6.Expression, b b6.Expression) b6.Expression {
	aq := a.AnyExpression.(*b6.QueryExpression)
	bq := b.AnyExpression.(*b6.QueryExpression)
	return b6.Expression{
		AnyExpression: &b6.QueryExpression{
			Query: b6.Intersection{aq.Query, bq.Query},
		},
		Begin: a.Begin,
		End:   b.End,
	}
}

func reduceOr(a b6.Expression, b b6.Expression) b6.Expression {
	aq := a.AnyExpression.(*b6.QueryExpression)
	bq := b.AnyExpression.(*b6.QueryExpression)
	return b6.Expression{
		AnyExpression: &b6.QueryExpression{
			Query: b6.Union{aq.Query, bq.Query},
		},
		Begin: a.Begin,
		End:   b.End,
	}
}

func ParseExpression(expression string) (b6.Expression, error) {
	yyErrorVerbose = true
	l := lexer{Expression: expression}
	yyParse(&l)
	if l.Top.AnyExpression == nil {
		return b6.Expression{}, l.Err
	}
	return l.Top, l.Err
}

func ParseExpressionWithLHS(expression string, lhs b6.Expression) (b6.Expression, error) {
	yyErrorVerbose = true
	l := lexer{Expression: expression, LHS: lhs}
	yyParse(&l)
	if l.Top.AnyExpression == nil {
		return b6.Expression{}, l.Err
	}
	return l.Top, l.Err
}

type byBeginThenLength []b6.Expression

func (b byBeginThenLength) Len() int      { return len(b) }
func (b byBeginThenLength) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byBeginThenLength) Less(i, j int) bool {
	if b[i].Begin == b[j].Begin {
		return (b[i].End - b[i].Begin) > (b[j].End - b[j].Begin)
	}
	return b[i].Begin < b[j].Begin
}

func OrderTokens(e b6.Expression) []b6.Expression {
	tokens := []b6.Expression{}
	queue := []b6.Expression{e}
	for len(queue) > 0 {
		e := queue[len(queue)-1]
		queue = queue[0 : len(queue)-1]
		if c, ok := e.AnyExpression.(*b6.CallExpression); ok {
			queue = append(queue, c.Function)
			queue = append(queue, c.Args...)
		} else if l, ok := e.AnyExpression.(*b6.LambdaExpression); ok {
			queue = append(queue, l.Expression)
		}
		if e.End > e.Begin {
			tokens = append(tokens, e)
		}
	}
	sort.Sort(byBeginThenLength(tokens))
	filtered := make([]b6.Expression, 0, len(tokens)/2)
	for _, t := range tokens {
		if len(filtered) > 0 && filtered[len(filtered)-1].End > t.Begin {
			filtered[len(filtered)-1] = t
		} else {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func Simplify(expression b6.Expression, functions SymbolArgCounts) b6.Expression {
	if expression.AnyExpression == nil {
		return expression
	}
	switch expression.AnyExpression.(type) {
	case *b6.CallExpression:
		return simplifyCall(expression, functions)
	case *b6.LambdaExpression:
		return simplifyLambda(expression, functions)
	}
	return expression
}

func simplifyCall(expression b6.Expression, functions SymbolArgCounts) b6.Expression {
	call := expression.AnyExpression.(*b6.CallExpression)
	call.Function = Simplify(call.Function, functions)
	for i, arg := range call.Args {
		call.Args[i] = Simplify(arg, functions)
	}
	if e, ok := simplifyCallWithNoArguments(expression, functions); ok {
		return e
	}
	return expression
}

func simplifyCallWithNoArguments(expression b6.Expression, functions SymbolArgCounts) (b6.Expression, bool) {
	// Calling a function that expects arguments with no arguments is
	// semantically equivilent to just using that function.
	call := expression.AnyExpression.(*b6.CallExpression)
	if len(call.Args) == 0 {
		if symbol, ok := call.Function.AnyExpression.(*b6.SymbolExpression); ok {
			if n, ok := functions.ArgCount(*symbol); ok && n > 0 {
				return call.Function, true
			}
		}
	}
	return expression, false
}

func simplifyLambda(expression b6.Expression, functions SymbolArgCounts) b6.Expression {
	// '{a -> area a}' is semantically equivalent to 'area'
	lambda := expression.AnyExpression.(*b6.LambdaExpression)
	if call, ok := lambda.Expression.AnyExpression.(*b6.CallExpression); ok && len(lambda.Args) > 0 {
		i := 0
		for i < len(lambda.Args) && i < len(call.Args) {
			if s, ok := call.Args[i].AnyExpression.(*b6.SymbolExpression); ok {
				if s.String() != lambda.Args[i] {
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
			return simplifyCall(b6.Expression{
				AnyExpression: &b6.CallExpression{
					Function: call.Function,
					Args:     call.Args[i:len(call.Args)],
				},
				Begin: expression.Begin,
				End:   expression.End,
			}, functions)
		}
	}
	return expression
}

func EscapeTagKey(v string) string {
	if v == "" {
		return ""
	}
	escape := (v[0] < 'a' && v[0] > 'z') && (v[0] < 'A' && v[0] > 'Z') && v[0] != '_' && v[0] != '#' && v[0] != '@'
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
	escape := (v[0] < 'a' && v[0] > 'z') && (v[0] < 'A' && v[0] > 'Z') && v[0] != '_'
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

func UnparseTag(t b6.Tag) string {
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
		return UnparseTag(b6.Tag(q)), true
	case b6.Keyed:
		return q.Key, true
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

func UnparseExpression(e b6.Expression) (string, bool) {
	return unparseExpression(e, true)
}

func unparseExpression(e b6.Expression, top bool) (string, bool) {
	switch e := e.AnyExpression.(type) {
	case *b6.SymbolExpression:
		return e.String(), true
	case *b6.CallExpression:
		if e.Pipelined {
			return unparsePipelinedCall(e, top)
		} else {
			return unparseCall(e, top)
		}
	case *b6.LambdaExpression:
		return unparseLambda(e)
	case b6.AnyLiteral:
		return unparseLiteral(e)
	}
	return "", false
}

func unparsePipelinedCall(call *b6.CallExpression, top bool) (string, bool) {
	lhs, ok := unparseExpression(call.Args[0], true)
	if !ok {
		return "", false
	}
	rhs, ok := unparseCall(&b6.CallExpression{Function: call.Function, Args: call.Args[1:]}, true)
	if !ok {
		return "", false
	}
	if top {
		return lhs + " | " + rhs, true
	} else {
		return "(" + lhs + " | " + rhs + ")", true
	}
}

func unparseCall(call *b6.CallExpression, top bool) (string, bool) {
	if len(call.Args) == 0 {
		return unparseExpression(call.Function, top)
	}
	parts := []string{}
	if part, ok := unparseExpression(call.Function, false); ok {
		parts = append(parts, part)
	} else {
		return "", false
	}
	for _, arg := range call.Args {
		if part, ok := unparseExpression(arg, false); ok {
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

func unparseLiteral(l b6.AnyLiteral) (string, bool) {
	switch l := l.(type) {
	case *b6.StringExpression:
		return UnparseString(string(*l)), true
	case *b6.IntExpression:
		return fmt.Sprintf("%d", int(*l)), true
	case *b6.FloatExpression:
		return fmt.Sprintf("%.2f", float64(*l)), true
	case *b6.TagExpression:
		return UnparseTag(b6.Tag{Key: l.Key, Value: l.Value}), true
	case *b6.FeatureIDExpression:
		return UnparseFeatureID(b6.FeatureID(*l), true), true
	case *b6.PointExpression:
		return fmt.Sprintf("%f, %f", l.Lat.Degrees(), l.Lng.Degrees()), true
	case *b6.QueryExpression:
		return UnparseQuery(l.Query)
	default:
		return fmt.Sprintf("(broken-value \"%v\")", l), true
	}
}

func unparseLambda(l *b6.LambdaExpression) (string, bool) {
	body, ok := unparseExpression(l.Expression, true)
	if !ok {
		return "", false
	}
	if len(l.Args) > 0 {
		args := strings.Join(l.Args, ", ")
		return fmt.Sprintf("{%s -> %s}", args, body), true
	} else {
		return fmt.Sprintf("{-> %s}", body), true
	}
}
