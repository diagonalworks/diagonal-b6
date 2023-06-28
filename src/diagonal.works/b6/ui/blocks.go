package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/ingest"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/renderer"
	"github.com/golang/geo/s2"
	"google.golang.org/protobuf/encoding/protojson"
)

// Design scratch:
// Places where a given value can be rendered:
// - As the result block of an evaluation
// - As the key of a collection
// - As the value of a collection
// - Of a value of a collection where we omit the key
// - As the value of a tag?? (though technically they're all strings)
// - In the points section of a feature (same as collection?)
// Call Type Class sine that what it impacts in the CSS?

type Block interface {
	BlockType() string
}

type PipelineStageBlockJSON struct {
	Type       string
	Expression string
}

func (f PipelineStageBlockJSON) BlockType() string {
	return f.Type
}

type TagJSON struct {
	Prefix          string
	Key             string
	Value           string
	KeyExpression   *NodeJSON
	ValueExpression *NodeJSON
}

type PointJSON struct {
	Label      string
	Expression string
}

type FeatureBlockJSON struct {
	Type      string
	Tags      []TagJSON
	Points    []Block
	MapCenter *geojson.Point
}

func (f FeatureBlockJSON) BlockType() string {
	return f.Type
}

func (f *FeatureBlockJSON) Fill(feature b6.Feature, w b6.World) {
	tags := feature.AllTags()
	f.Tags = make([]TagJSON, len(tags))
	for i, tag := range tags {
		if strings.HasPrefix(tag.Key, "#") || strings.HasPrefix(tag.Key, "#") {
			f.Tags[i] = TagJSON{Prefix: tag.Key[0:1], Key: tag.Key[1:], Value: tag.Value}
			f.Tags[i].KeyExpression = (*NodeJSON)(taggedQueryLiteral(tag.Key, tag.Value))
		} else {
			f.Tags[i] = TagJSON{Key: tag.Key, Value: tag.Value}
		}
		f.Tags[i].ValueExpression = (*NodeJSON)(getStringExpression(feature, tag.Key))
	}
	if path, ok := feature.(b6.PathFeature); ok {
		for i := 0; i < path.Len(); i++ {
			if point := path.Feature(i); point != nil {
				var b CollectionFeatureBlockJSON
				b.Fill(point)
				f.Points = append(f.Points, b)
			} else {
				f.Points = append(f.Points)
				f.Points = append(f.Points, StringBlockJSON{Type: "string", Value: pointToExpression(path.Point(i))})
			}
		}
	}
	if p, ok := feature.(b6.PhysicalFeature); ok {
		center := geojson.FromS2Point(b6.Center(p))
		f.MapCenter = &center
	}
}

func intLiteral(value int) *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_IntValue{
					IntValue: int64(value),
				},
			},
		},
	}
}

func floatLiteral(value float64) *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_FloatValue{
					FloatValue: value,
				},
			},
		},
	}
}

func stringLiteral(value string) *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_StringValue{
					StringValue: value,
				},
			},
		},
	}
}

func taggedQueryLiteral(key string, value string) *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Literal{
			Literal: &pb.LiteralNodeProto{
				Value: &pb.LiteralNodeProto_QueryValue{
					QueryValue: &pb.QueryProto{
						Query: &pb.QueryProto_Tagged{
							Tagged: &pb.TagProto{
								Key:   key,
								Value: value,
							},
						},
					},
				},
			},
		},
	}
}

func getStringExpression(f b6.Feature, key string) *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Call{
			Call: &pb.CallNodeProto{
				Function: &pb.NodeProto{
					Node: &pb.NodeProto_Symbol{
						Symbol: "get-string",
					},
				},
				Args: []*pb.NodeProto{
					&pb.NodeProto{
						Node: &pb.NodeProto_Literal{
							Literal: &pb.LiteralNodeProto{
								Value: &pb.LiteralNodeProto_FeatureIDValue{
									FeatureIDValue: b6.NewProtoFromFeatureID(f.FeatureID()),
								},
							},
						},
					},
					&pb.NodeProto{
						Node: &pb.NodeProto_Literal{
							Literal: &pb.LiteralNodeProto{
								Value: &pb.LiteralNodeProto_StringValue{
									StringValue: key,
								},
							},
						},
					},
				},
			},
		},
	}
}

func findFeatureExpression(f b6.Identifiable) *pb.NodeProto {
	return &pb.NodeProto{
		Node: &pb.NodeProto_Call{
			Call: &pb.CallNodeProto{
				Function: &pb.NodeProto{
					Node: &pb.NodeProto_Symbol{
						Symbol: "find-feature",
					},
				},
				Args: []*pb.NodeProto{
					&pb.NodeProto{
						Node: &pb.NodeProto_Literal{
							Literal: &pb.LiteralNodeProto{
								Value: &pb.LiteralNodeProto_FeatureIDValue{
									FeatureIDValue: b6.NewProtoFromFeatureID(f.FeatureID()),
								},
							},
						},
					},
				},
			},
		},
	}
}

func pointToExpression(p s2.Point) string {
	ll := s2.LatLngFromPoint(p)
	return fmt.Sprintf("%f, %f", ll.Lat.Degrees(), ll.Lng.Degrees())
}

type TitleCountBlockJSON struct {
	Type  string
	Title string
	Count int
}

func (t *TitleCountBlockJSON) Fill(title string, count int) {
	t.Type = "title-count"
	t.Title = title
	t.Count = count
}

type CollectionBlockJSON struct {
	Type  string
	Title TitleCountBlockJSON
	Items []Block
}

const CollectionBlockItemLimit = 200
const CollectionBlockHighlightLimit = 10000

func (f CollectionBlockJSON) BlockType() string {
	return f.Type
}

func (b *CollectionBlockJSON) Fill(c api.Collection, r *BlockResponseJSON, w b6.World) error {
	count := -1
	if countable, ok := c.(api.Countable); ok {
		count = countable.Count()
	}
	b.Title.Fill("Collection", count)
	keys := make([]interface{}, 0, 8)
	values := make([]interface{}, 0, 8)
	var err error
	keys, values, err = fillKeyValues(c, keys, values)
	if err != nil {
		return err
	}
	if isFeatureCollection(keys, values) || isArrayCollection(keys, values) {
		for i := range values {
			if f, ok := values[i].(b6.Feature); ok {
				if len(b.Items) < CollectionBlockItemLimit {
					var item CollectionFeatureBlockJSON
					item.Fill(f)
					b.Items = append(b.Items, item)
				}
				r.Highlighted.Add(f.FeatureID())
			} else {
				if len(b.Items) < CollectionBlockItemLimit {
					b.Items = append(b.Items, StringBlockJSON{Type: "string", Value: fmt.Sprintf("%+v", values[i])})
				}
			}
		}
	} else {
		for i := range keys {
			if len(b.Items) < CollectionBlockItemLimit {
				var item CollectionKeyValueBlockJSON
				item.Fill(keys[i], values[i])
				b.Items = append(b.Items, &item)
			}
			if f, ok := keys[i].(b6.Identifiable); ok {
				r.Highlighted.Add(f.FeatureID())
			}
			if f, ok := values[i].(b6.Identifiable); ok {
				r.Highlighted.Add(f.FeatureID())
			}
		}
	}
	return nil
}

func fillKeyValues(c api.Collection, keys []interface{}, values []interface{}) ([]interface{}, []interface{}, error) {
	i := c.Begin()
	var err error
	for {
		var ok bool
		ok, err = i.Next()
		if !ok || err != nil {
			break
		}
		keys = append(keys, i.Key())
		values = append(values, i.Value())
		if len(keys) >= CollectionBlockHighlightLimit {
			break
		}
	}
	return keys, values, err
}

func isFeatureCollection(keys []interface{}, values []interface{}) bool {
	if len(keys) > 0 {
		if id, ok := keys[0].(b6.FeatureID); ok {
			if f, ok := values[0].(b6.Feature); ok {
				return id == f.FeatureID()
			}
		}
	}
	return false
}

func isArrayCollection(keys []interface{}, values []interface{}) bool {
	if len(keys) > 0 {
		for i, k := range keys {
			if ii, ok := k.(int); !ok || i != ii {
				return false
			}
		}
	} else {
		return false
	}
	return true
}

type CollectionKeyValueBlockJSON struct {
	Type  string
	Key   Block
	Value Block
}

func (b CollectionKeyValueBlockJSON) BlockType() string {
	return b.Type
}

func (b *CollectionKeyValueBlockJSON) Fill(key interface{}, value interface{}) {
	b.Type = "collection-key-value"
	if f, ok := key.(b6.Identifiable); ok {
		var block CollectionFeatureKeyBlockJSON
		block.Fill(f)
		b.Key = &block
	} else {
		var k CollectionKeyOrValueBlockJSON
		k.Fill(key)
		b.Key = &k
	}
	var v CollectionKeyOrValueBlockJSON
	v.Fill(value)
	b.Value = &v
}

type CollectionFeatureBlockJSON struct {
	Type       string
	Icon       string
	Label      string
	Namespace  string
	ID         string
	Expression *NodeJSON
}

func (b CollectionFeatureBlockJSON) BlockType() string {
	return b.Type
}

func (b *CollectionFeatureBlockJSON) Fill(f b6.Feature) {
	b.Type = "collection-feature"
	b.Icon = featureIcon(f)
	b.Label = LabelForFeature(f).Singular
	b.Namespace = LabelForNamespace(f.FeatureID().Namespace)
	b.ID = featureID(f)
	b.Expression = (*NodeJSON)(findFeatureExpression(f))
}

func featureIcon(f b6.Identifiable) string {
	switch f.FeatureID().Type {
	case b6.FeatureTypePoint:
		return "point"
	case b6.FeatureTypePath:
		return "path"
	case b6.FeatureTypeArea:
		return "area"
	case b6.FeatureTypeRelation:
		return "area"
	}
	return ""
}

func featureID(f b6.Feature) string {
	if name := f.Get("name"); name.IsValid() {
		return name.Value
	} else if code := f.Get("code"); code.IsValid() {
		return code.Value
	} else if ref := f.Get("ref"); ref.IsValid() {
		return ref.Value
	} else {
		switch f.FeatureID().Namespace {
		case b6.NamespaceGBCodePoint:
			if postcode, ok := b6.PostcodeFromPointID(f.FeatureID().ToPointID()); ok {
				return postcode
			}
		case b6.NamespaceUKONSBoundaries:
			if code, _, ok := b6.UKONSCodeFromFeatureID(f.FeatureID()); ok {
				return code
			}
		}
	}
	return fmt.Sprintf("%d", f.FeatureID().Value)
}

type CollectionFeatureKeyBlockJSON struct {
	Type       string
	Icon       string
	Namespace  string
	ID         string
	Expression *NodeJSON
}

func (b *CollectionFeatureKeyBlockJSON) BlockType() string {
	return b.Type
}

func (b *CollectionFeatureKeyBlockJSON) Fill(f b6.Identifiable) {
	b.Type = "collection-feature-key"
	b.Icon = featureIcon(f)
	b.Namespace = LabelForNamespace(f.FeatureID().Namespace)
	b.ID = fmt.Sprintf("%v", f.FeatureID().Value)
	b.Expression = (*NodeJSON)(findFeatureExpression(f))
}

type CollectionKeyOrValueBlockJSON struct {
	Type       string
	Value      string
	Expression *NodeJSON
}

func (b *CollectionKeyOrValueBlockJSON) Fill(v interface{}) {
	b.Type = "collection-key-or-value"
	// TODO: Factor out when we rework blocks
	switch v := v.(type) {
	case int:
		b.Value = strconv.Itoa(v)
		b.Expression = (*NodeJSON)(intLiteral(v))
	case float64:
		b.Value = fmt.Sprintf("%f", v)
		b.Expression = (*NodeJSON)(floatLiteral(v))
	case string:
		b.Value = v
		b.Expression = (*NodeJSON)(stringLiteral(v))
	case b6.Tag:
		b.Value = api.UnparseTag(v)
		b.Expression = (*NodeJSON)(taggedQueryLiteral(v.Key, v.Value))
	default:
		b.Value = fmt.Sprintf("%+v", v)
	}
}

func (b *CollectionKeyOrValueBlockJSON) BlockType() string {
	return b.Type
}

type FloatBlockJSON struct {
	Type  string
	Value float64
}

func (f FloatBlockJSON) BlockType() string {
	return f.Type
}

type IntBlockJSON struct {
	Type  string
	Value int
}

func (i IntBlockJSON) BlockType() string {
	return i.Type
}

type StringBlockJSON struct {
	Type  string
	Value string
}

func (i StringBlockJSON) BlockType() string {
	return i.Type
}

type PointBlockJSON struct {
	Type      string
	Value     string
	MapCenter *geojson.Point
}

func (p PointBlockJSON) BlockType() string {
	return p.Type
}

type PlaceholderBlockJSON struct {
	Type     string
	RawValue string
}

func (p PlaceholderBlockJSON) BlockType() string {
	return p.Type
}

type GeometryBlockJSON struct {
	Type      string
	GeoJSON   geojson.GeoJSON
	Dimension float64
}

func (p GeometryBlockJSON) BlockType() string {
	return p.Type
}

type NodeJSON pb.NodeProto

func (n *NodeJSON) MarshalJSON() ([]byte, error) {
	return protojson.Marshal((*pb.NodeProto)(n))
}

func (n *NodeJSON) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, (*pb.NodeProto)(n))
}

type FeatureIDSetJSON map[string][]string

func (f FeatureIDSetJSON) Add(id b6.FeatureID) {
	key := fmt.Sprintf("/%s/%s", id.Type.String(), id.Namespace.String())
	f[key] = append(f[key], strconv.FormatUint(id.Value, 16))
}

type BlocksJSON []Block

type BlockRequestJSON struct {
	Expression string
	Node       *NodeJSON
}

type BlockResponseJSON struct {
	Blocks      BlocksJSON
	Node        *NodeJSON
	Functions   []string
	Highlighted FeatureIDSetJSON
	QueryLayers []string
}

type BlockHandler struct {
	World            ingest.MutableWorld
	RenderRules      renderer.RenderRules
	Cores            int
	FunctionSymbols  api.FunctionSymbols
	FunctionWrappers api.FunctionWrappers
}

func (b *BlockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := BlockRequestJSON{Node: &NodeJSON{}}
	response := BlockResponseJSON{Node: &NodeJSON{}, Highlighted: make(FeatureIDSetJSON)}

	if r.Method == "GET" {
		request.Expression = r.URL.Query().Get("e")
	} else if r.Method == "POST" {
		d := json.NewDecoder(r.Body)
		if err := d.Decode(&request); err != nil {
			log.Printf("err: %s", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		r.Body.Close()
	} else {
		log.Printf("Bad method")
		http.Error(w, "Bad method", http.StatusMethodNotAllowed)
		return
	}
	if request.Expression == "" && request.Node == nil {
		log.Printf("No node or expression")
		http.Error(w, "No expression or node", http.StatusBadRequest)
		return
	}

	var node *pb.NodeProto
	if request.Expression != "" {
		var err error
		if request.Node == nil || request.Node.Node == nil {
			node, err = api.ParseExpression(request.Expression)
		} else {
			node, err = api.ParseExpressionWithLHS(request.Expression, (*pb.NodeProto)(request.Node))
		}
		if err != nil {
			response.Blocks = fillBlocksFromError(response.Blocks, err)
			sendBlockResponse(&response, w)
			return
		}

	} else {
		node = (*pb.NodeProto)(request.Node)
	}

	node = api.Simplify(node, b.FunctionSymbols)
	context := api.Context{
		World:            b.World,
		FunctionSymbols:  b.FunctionSymbols,
		FunctionWrappers: b.FunctionWrappers,
		Cores:            b.Cores,
		Context:          context.Background(),
	}
	response.Blocks = fillBlocksFromExpression(response.Blocks, node, true)
	response.Node = (*NodeJSON)(node)
	result, err := api.Evaluate(node, &context)
	if err == nil {
		fillResponseFromResult(&response, result, b.RenderRules, b.World)
		response.Functions = fillMatchingFunctionSymbols(response.Functions, result, b.FunctionSymbols)
	} else {
		response.Blocks = fillBlocksFromError(response.Blocks, err)
	}
	sendBlockResponse(&response, w)
}

func sendBlockResponse(response *BlockResponseJSON, w http.ResponseWriter) {
	output, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(output))
}

func fillBlocksFromExpression(blocks BlocksJSON, node *pb.NodeProto, root bool) BlocksJSON {
	if call, ok := node.Node.(*pb.NodeProto_Call); ok {
		if call.Call.Pipelined {
			left := call.Call.Args[0]
			right := &pb.NodeProto{
				Node: &pb.NodeProto_Call{
					Call: &pb.CallNodeProto{
						Function: call.Call.Function,
						Args:     call.Call.Args[1:],
					},
				},
			}
			blocks = fillBlocksFromExpression(blocks, left, false)
			blocks = fillBlocksFromExpression(blocks, right, false)
			return blocks
		}
	}
	_, isLiteral := node.Node.(*pb.NodeProto_Literal)
	if expression, ok := api.UnparseNode(node); ok {
		if !isLiteral || !root {
			blocks = append(blocks, PipelineStageBlockJSON{
				Type:       "pipeline-stage",
				Expression: expression,
			})
		}
	} else {
		// TODO: Not actually an error, as we can click on it
		blocks = append(blocks, PipelineStageBlockJSON{
			Type:       "error",
			Expression: "can't convert",
		})
	}
	return blocks
}

func fillResponseFromResult(response *BlockResponseJSON, result interface{}, rules renderer.RenderRules, w b6.World) {
	if i, ok := api.ToInt(result); ok {
		response.Blocks = append(response.Blocks, IntBlockJSON{Type: "int-result", Value: i})
	} else if f, ok := api.ToFloat(result); ok {
		response.Blocks = append(response.Blocks, FloatBlockJSON{Type: "float-result", Value: f})
	} else {
		switch r := result.(type) {
		case b6.Feature:
			block := FeatureBlockJSON{Type: "feature"}
			block.Fill(r, w)
			response.Blocks = append(response.Blocks, block)
			response.Highlighted.Add(r.FeatureID())
		case b6.Tag:
			response.Blocks = append(response.Blocks, StringBlockJSON{Type: "string-result", Value: api.UnparseTag(r)})
			if !rules.IsRendered(r) {
				if q, ok := api.UnparseQuery(b6.Tagged(r)); ok {
					response.QueryLayers = append(response.QueryLayers, q)
				}
			}
		case b6.Query:
			if q, ok := api.UnparseQuery(r); ok {
				response.Blocks = append(response.Blocks, StringBlockJSON{Type: "string-result", Value: q})
			} else {
				response.Blocks = append(response.Blocks, StringBlockJSON{Type: "string-result", Value: "query"})
			}
		case api.Collection:
			block := CollectionBlockJSON{Type: "collection"}
			if err := block.Fill(r, response, w); err == nil {
				response.Blocks = append(response.Blocks, block)
			} else {
				response.Blocks = fillBlocksFromError(response.Blocks, err)
			}
		case b6.Area:
			block := GeometryBlockJSON{
				Type:    "area",
				GeoJSON: r.ToGeoJSON(),
			}
			for i := 0; i < r.Len(); i++ {
				block.Dimension += b6.AreaToMeters2(r.Polygon(i).Area())
			}
			response.Blocks = append(response.Blocks, block)
		case b6.Path:
			block := GeometryBlockJSON{
				Type:      "path",
				GeoJSON:   r.ToGeoJSON(),
				Dimension: b6.AngleToMeters(r.Polyline().Length()),
			}
			response.Blocks = append(response.Blocks, block)
		case *geojson.FeatureCollection:
			block := GeometryBlockJSON{
				Type:      "geojson-feature-collection",
				GeoJSON:   r,
				Dimension: float64(len(r.Features)),
			}
			response.Blocks = append(response.Blocks, block)
		case *geojson.Feature:
			block := GeometryBlockJSON{
				Type:    "geojson-feature",
				GeoJSON: r,
			}
			response.Blocks = append(response.Blocks, block)
		case string:
			response.Blocks = append(response.Blocks, StringBlockJSON{Type: "string-result", Value: r})
		case b6.Point:
			block := PointBlockJSON{Type: "string-result", Value: pointToExpression(r.Point())}
			center := geojson.FromS2Point(r.Point())
			block.MapCenter = &center
			response.Blocks = append(response.Blocks, block)
		default:
			response.Blocks = append(response.Blocks, PlaceholderBlockJSON{Type: "placeholder", RawValue: fmt.Sprintf("%+v", r)})
		}
	}
}

func fillBlocksFromError(blocks BlocksJSON, err error) BlocksJSON {
	return append(blocks, StringBlockJSON{Type: "error", Value: err.Error()})
}

func fillMatchingFunctionSymbols(symbols []string, result interface{}, functions api.FunctionSymbols) []string {
	t := reflect.TypeOf(result)
	for symbol, f := range functions {
		tt := reflect.TypeOf(f)
		if tt.Kind() == reflect.Func && tt.NumIn() > 0 {
			if api.CanUseAsArg(t, tt.In(0)) {
				symbols = append(symbols, symbol)
			}
		}
	}
	return symbols
}

func NewBlockHandler(w ingest.MutableWorld, cores int) *BlockHandler {
	local := make(api.FunctionSymbols)
	for name, f := range functions.Functions() {
		local[name] = f
	}
	return &BlockHandler{
		World:            w,
		RenderRules:      renderer.BasemapRenderRules,
		Cores:            cores,
		FunctionSymbols:  local,
		FunctionWrappers: functions.Wrappers(),
	}
}
