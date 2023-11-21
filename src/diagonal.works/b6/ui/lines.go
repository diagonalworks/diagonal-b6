package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

type DefaultUIRenderer struct {
	BasemapRules    renderer.RenderRules
	FunctionSymbols api.FunctionSymbols
	World           b6.World
}

func (d *DefaultUIRenderer) Render(response *UIResponseJSON, value interface{}, context b6.CollectionFeature, locked bool) error {
	if err := d.fillResponseFromResult(response, value); err == nil {
		shell := &pb.ShellLineProto{
			Functions: make([]string, 0),
		}
		shell.Functions = fillMatchingFunctionSymbols(shell.Functions, value, d.FunctionSymbols)
		response.Proto.Stack.Substacks = append(response.Proto.Stack.Substacks, &pb.SubstackProto{
			Lines: []*pb.LineProto{{Line: &pb.LineProto_Shell{Shell: shell}}},
		})
		return nil
	} else {
		return d.fillResponseFromResult(response, err)
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

const CollectionLineLimit = 200
const CollectionHighlightLimit = 10000

func fillKeyValues(c b6.UntypedCollection, keys []interface{}, values []interface{}) ([]interface{}, []interface{}, error) {
	i := c.BeginUntyped()
	var err error
	for {
		var ok bool
		ok, err = i.Next()
		if !ok || err != nil {
			break
		}
		keys = append(keys, i.Key())
		values = append(values, i.Value())
		if len(keys) >= CollectionHighlightLimit {
			break
		}
	}
	return keys, values, err
}

func isFeatureCollection(keys []interface{}, values []interface{}) bool {
	if len(keys) > 0 {
		if k, ok := keys[0].(b6.Identifiable); ok {
			if v, ok := values[0].(b6.Identifiable); ok {
				return k.FeatureID() == v.FeatureID()
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

func fillMatchingFunctionSymbols(symbols []string, result interface{}, functions api.FunctionSymbols) []string {
	if result != nil {
		t := reflect.TypeOf(result)
		for symbol, f := range functions {
			tt := reflect.TypeOf(f)
			if tt.Kind() == reflect.Func && tt.NumIn() > 1 {
				if api.CanUseAsArg(t, tt.In(1)) {
					symbols = append(symbols, symbol)
				}
			}
		}
	}
	return symbols
}

func NewUIHandler(renderer UIRenderer, w ingest.MutableWorld, options api.Options) *UIHandler {
	return &UIHandler{
		Renderer:        renderer,
		World:           w,
		Options:         options,
		FunctionSymbols: functions.Functions(),
		Adaptors:        functions.Adaptors(),
	}
}

type UIResponseProtoJSON pb.UIResponseProto

func (b *UIResponseProtoJSON) MarshalJSON() ([]byte, error) {
	return protojson.Marshal((*pb.UIResponseProto)(b))
}

func (b *UIResponseProtoJSON) UnmarshalJSON(buffer []byte) error {
	return protojson.Unmarshal(buffer, (*pb.UIResponseProto)(b))
}

type UIResponseJSON struct {
	Proto   *UIResponseProtoJSON `json:"proto,omitempty"`
	GeoJSON []geojson.GeoJSON    `json:"geoJSON,omitempty"`
}

func NewUIResponseJSON() *UIResponseJSON {
	return &UIResponseJSON{
		Proto: &UIResponseProtoJSON{
			Stack: &pb.StackProto{},
		},
	}
}

func (u *UIResponseJSON) AddGeoJSON(g geojson.GeoJSON) {
	u.GeoJSON = append(u.GeoJSON, g)
	u.Proto.GeoJSON = append(u.Proto.GeoJSON, &pb.GeoJSONProto{
		Index: int32(len(u.GeoJSON) - 1),
	})
}

type UIHandler struct {
	World           ingest.MutableWorld
	Renderer        UIRenderer
	Options         api.Options
	FunctionSymbols api.FunctionSymbols
	Adaptors        api.Adaptors
}

func (u *UIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := &pb.UIRequestProto{}
	response := NewUIResponseJSON()

	if r.Method == "GET" {
		request.Expression = r.URL.Query().Get("e")
	} else if r.Method == "POST" {
		var err error
		var body []byte
		if body, err = io.ReadAll(r.Body); err == nil {
			r.Body.Close()
			err = protojson.Unmarshal(body, request)
		}
		if err != nil {
			log.Printf("Bad request body")
			http.Error(w, "Bad request body", http.StatusBadRequest)
			return
		}
	} else {
		log.Printf("Bad method")
		http.Error(w, "Bad method", http.StatusMethodNotAllowed)
		return
	}

	if request.Expression == "" && request.Node == nil {
		log.Printf("No expression")
		http.Error(w, "No expression", http.StatusBadRequest)
		return
	}

	var root b6.CollectionFeature
	if request.Context != nil && request.Context.Type == pb.FeatureType_FeatureTypeCollection {
		root = b6.FindCollectionByID(b6.NewFeatureIDFromProto(request.Context).ToCollectionID(), u.World)
	}

	var expression b6.Expression
	if request.Expression != "" {
		var err error
		if request.Node == nil {
			expression, err = api.ParseExpression(request.Expression)
		} else {
			var lhs b6.Expression
			if err = lhs.FromProto(request.Node); err == nil {
				expression, err = api.ParseExpressionWithLHS(request.Expression, lhs)
			}
		}
		if err != nil {
			u.Renderer.Render(response, err, root, request.Locked)
			var substack pb.SubstackProto
			fillSubstackFromError(&substack, err)
			response.Proto.Stack.Substacks = append(response.Proto.Stack.Substacks, &substack)
			sendUIResponse(response, w)
			return
		}
	} else {
		expression.FromProto(request.Node)
	}
	expression = api.Simplify(expression, u.FunctionSymbols)

	if !request.Locked {
		substack := &pb.SubstackProto{}
		fillSubstackFromExpression(substack, expression, true)
		if len(substack.Lines) > 0 {
			response.Proto.Stack.Substacks = append(response.Proto.Stack.Substacks, substack)
		}
	}

	if unparsed, ok := api.UnparseExpression(expression); ok {
		response.Proto.Expression = unparsed
	}

	var err error
	if response.Proto.Node, err = expression.ToProto(); err != nil {
		u.Renderer.Render(response, err, root, request.Locked)
		sendUIResponse(response, w)
		return
	}

	vmContext := api.Context{
		World:           u.World,
		FunctionSymbols: u.FunctionSymbols,
		Adaptors:        u.Adaptors,
		Context:         context.Background(),
	}
	vmContext.FillFromOptions(&u.Options)

	result, err := api.Evaluate(expression, &vmContext)
	if err == nil {
		err = u.Renderer.Render(response, result, root, request.Locked)
	}
	if err != nil {
		u.Renderer.Render(response, err, root, request.Locked)
	}
	sendUIResponse(response, w)
}

func sendUIResponse(response *UIResponseJSON, w http.ResponseWriter) {
	output, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(output))
}

func highlightInResponse(response *pb.UIResponseProto, id b6.FeatureID) {
	if response.Highlighted == nil {
		response.Highlighted = &pb.FeatureIDsProto{}
	}
	n := fmt.Sprintf("/%s/%s", id.Type.String(), id.Namespace.String())
	ids := -1
	for i, nn := range response.Highlighted.Namespaces {
		if n == nn {
			ids = i
			break
		}
	}
	if ids < 0 {
		ids = len(response.Highlighted.Ids)
		response.Highlighted.Namespaces = append(response.Highlighted.Namespaces, n)
		response.Highlighted.Ids = append(response.Highlighted.Ids, &pb.IDsProto{})
	}
	response.Highlighted.Ids[ids].Ids = append(response.Highlighted.Ids[ids].Ids, id.Value)
}

func fillSubstackFromExpression(lines *pb.SubstackProto, expression b6.Expression, root bool) {
	if call, ok := expression.AnyExpression.(*b6.CallExpression); ok {
		if call.Pipelined {
			left := call.Args[0]
			right := b6.Expression{
				AnyExpression: &b6.CallExpression{
					Function: call.Function,
					Args:     call.Args[1:],
				},
			}
			fillSubstackFromExpression(lines, left, false)
			fillSubstackFromExpression(lines, right, false)
			return
		}
	}
	_, isLiteral := expression.AnyExpression.(b6.AnyLiteral)
	if expression, ok := api.UnparseExpression(expression); ok {
		if !isLiteral || !root {
			lines.Lines = append(lines.Lines, &pb.LineProto{
				Line: &pb.LineProto_Expression{
					Expression: &pb.ExpressionLineProto{
						Expression: expression,
					},
				},
			})
		}
	} else {
		lines.Lines = append(lines.Lines, &pb.LineProto{
			Line: &pb.LineProto_Error{
				Error: &pb.ErrorLineProto{
					Error: "can't convert expression",
				},
			},
		})
	}
}

func fillSubstackFromCollection(substack *pb.SubstackProto, c b6.UntypedCollection, response *pb.UIResponseProto, w b6.World) error {
	// TODO: Set collection title based on collection contents
	if count, ok := c.Count(); ok {
		line := leftRightValueLineFromValues("Collection", count, w)
		substack.Lines = append(substack.Lines, line)
	} else {
		substack.Lines = append(substack.Lines, &pb.LineProto{
			Line: &pb.LineProto_Value{
				Value: &pb.ValueLineProto{
					Atom: AtomFromString("Collection"),
				},
			},
		})
	}

	keys := make([]interface{}, 0, 8)
	values := make([]interface{}, 0, 8)
	var err error
	keys, values, err = fillKeyValues(c, keys, values)
	if err != nil {
		return err
	}
	if isFeatureCollection(keys, values) || isArrayCollection(keys, values) {
		for i := range values {
			if i < CollectionLineLimit {
				line := ValueLineFromValue(values[i], w)
				substack.Lines = append(substack.Lines, line)
			} else {
				break
			}
		}
	} else {
		for i := range keys {
			if i < CollectionLineLimit {
				line := leftRightValueLineFromValues(keys[i], values[i], w)
				substack.Lines = append(substack.Lines, line)
			} else {
				break
			}
		}
	}

	for i := range keys {
		if id, ok := keys[i].(b6.Identifiable); ok {
			highlightInResponse(response, id.FeatureID())
		}
		if id, ok := values[i].(b6.Identifiable); ok {
			highlightInResponse(response, id.FeatureID())
		}
	}
	return nil
}

func lineFromTags(f b6.Feature) *pb.LineProto {
	tags := f.AllTags()
	tl := &pb.TagsLineProto{
		Tags: make([]*pb.TagAtomProto, len(tags)),
	}
	for i, tag := range tags {
		if strings.HasPrefix(tag.Key, "#") || strings.HasPrefix(tag.Key, "#") {
			tl.Tags[i] = &pb.TagAtomProto{Prefix: tag.Key[0:1], Key: tag.Key[1:], Value: tag.Value}
		} else {
			tl.Tags[i] = &pb.TagAtomProto{Prefix: "", Key: tag.Key, Value: tag.Value}
		}
		tl.Tags[i].ClickExpression = getStringExpression(f, tag.Key)
	}
	return &pb.LineProto{
		Line: &pb.LineProto_Tags{
			Tags: tl,
		},
	}
}

func AtomFromString(value string) *pb.AtomProto {
	return &pb.AtomProto{
		Atom: &pb.AtomProto_Value{
			Value: value,
		},
	}
}

func AtomFromValue(value interface{}, w b6.World) *pb.AtomProto {
	if i, ok := api.ToInt(value); ok {
		return AtomFromString(strconv.Itoa(i))
	} else if f, err := api.ToFloat64(value); err == nil {
		return AtomFromString(fmt.Sprintf("%f", f))
	} else {
		switch v := value.(type) {
		case string:
			return AtomFromString(v)
		case b6.Feature:
			return &pb.AtomProto{
				Atom: &pb.AtomProto_LabelledIcon{
					LabelledIcon: &pb.LabelledIconProto{
						Icon:  v.FeatureID().Type.String(),
						Label: featureLabel(v),
					},
				},
			}
		case b6.FeatureID:
			if f := w.FindFeatureByID(v); f != nil {
				return AtomFromValue(f, w)
			} else {
				return &pb.AtomProto{
					Atom: &pb.AtomProto_LabelledIcon{
						LabelledIcon: &pb.LabelledIconProto{
							Icon:  v.Type.String(),
							Label: strings.Title(v.Type.String()),
						},
					},
				}
			}
		case b6.Tag:
			return AtomFromString(api.UnparseTag(v))
		case b6.Point:
			ll := s2.LatLngFromPoint(v.Point())
			return AtomFromString(fmt.Sprintf("%f, %f", ll.Lat.Degrees(), ll.Lng.Degrees()))
		default:
			return AtomFromString(fmt.Sprintf("%v", v))
		}
	}
}

func clickExpressionFromIdentifiable(f b6.Identifiable) *pb.NodeProto {
	if !f.FeatureID().IsValid() {
		return nil
	}
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

func clickExpressionFromValue(value interface{}) *pb.NodeProto {
	switch v := value.(type) {
	case b6.Identifiable:
		return clickExpressionFromIdentifiable(v)
	}
	return nil
}

func ValueLineFromValue(value interface{}, w b6.World) *pb.LineProto {
	return &pb.LineProto{
		Line: &pb.LineProto_Value{
			Value: &pb.ValueLineProto{
				Atom:            AtomFromValue(value, w),
				ClickExpression: clickExpressionFromValue(value),
			},
		},
	}
}

func leftRightValueLineFromValues(first interface{}, second interface{}, w b6.World) *pb.LineProto {
	return &pb.LineProto{
		Line: &pb.LineProto_LeftRightValue{
			LeftRightValue: &pb.LeftRightValueLineProto{
				Left: []*pb.ClickableAtomProto{
					&pb.ClickableAtomProto{
						Atom:            AtomFromValue(first, w),
						ClickExpression: clickExpressionFromValue(first),
					},
				},
				Right: &pb.ClickableAtomProto{
					Atom:            AtomFromValue(second, w),
					ClickExpression: clickExpressionFromValue(second),
				},
			},
		},
	}
}

func FillSubstackFromValue(substack *pb.SubstackProto, value interface{}, response *pb.UIResponseProto, w b6.World) {
	if c, ok := value.(b6.UntypedCollection); ok {
		fillSubstackFromCollection(substack, c, response, w)
	} else {
		substack.Lines = append(substack.Lines, ValueLineFromValue(value, w))
	}
}

func fillSubstackFromAtom(substack *pb.SubstackProto, atom *pb.AtomProto) {
	substack.Lines = append(substack.Lines, &pb.LineProto{
		Line: &pb.LineProto_Value{
			Value: &pb.ValueLineProto{
				Atom: atom,
			},
		},
	})
}

func fillSubstackFromError(substack *pb.SubstackProto, err error) {
	substack.Lines = append(substack.Lines, &pb.LineProto{
		Line: &pb.LineProto_Error{
			Error: &pb.ErrorLineProto{
				Error: err.Error(),
			},
		},
	})
}

func fillSubstacksFromFeature(substacks []*pb.SubstackProto, f b6.Feature, w b6.World) []*pb.SubstackProto {
	substack := &pb.SubstackProto{}
	substack.Lines = append(substack.Lines, ValueLineFromValue(f, w))
	if len(f.AllTags()) > 0 {
		substack.Lines = append(substack.Lines, lineFromTags(f))
	}
	substacks = append(substacks, substack)

	if path, ok := f.(b6.PathFeature); ok {
		substack := &pb.SubstackProto{Collapsable: true}
		line := leftRightValueLineFromValues("Points", path.Len(), w)
		substack.Lines = append(substack.Lines, line)
		for i := 0; i < path.Len(); i++ {
			if point := path.Feature(i); point != nil {
				substack.Lines = append(substack.Lines, ValueLineFromValue(point, w))
			} else {
				substack.Lines = append(substack.Lines, ValueLineFromValue(path.Point(i), w))
			}
		}
		substacks = append(substacks, substack)
	}

	if relation, ok := f.(b6.RelationFeature); ok {
		substack := &pb.SubstackProto{Collapsable: true}
		line := leftRightValueLineFromValues("Members", relation.Len(), w)
		substack.Lines = append(substack.Lines, line)
		var i int
		for i = 0; i < relation.Len() && i < CollectionLineLimit; i++ {
			member := relation.Member(i)
			if member.Role != "" {
				substack.Lines = append(substack.Lines, leftRightValueLineFromValues(member.ID, member.Role, w))
			} else {
				substack.Lines = append(substack.Lines, ValueLineFromValue(member.ID, w))
			}
		}
		if n := relation.Len() - CollectionLineLimit; n > 0 {
			substack.Lines = append(substack.Lines, ValueLineFromValue(fmt.Sprintf("%d more", n), w))
		}
		substacks = append(substacks, substack)
	}

	relations := b6.AllRelations(w.FindRelationsByFeature(f.FeatureID()))
	if len(relations) > 0 {
		substack := &pb.SubstackProto{Collapsable: true}
		line := leftRightValueLineFromValues("Relations", len(relations), w)
		substack.Lines = append(substack.Lines, line)
		for _, r := range relations {
			substack.Lines = append(substack.Lines, ValueLineFromValue(r, w))
		}
		substacks = append(substacks, substack)
	}
	return substacks
}

func (d *DefaultUIRenderer) fillResponseFromResult(response *UIResponseJSON, result interface{}) error {
	p := (*pb.UIResponseProto)(response.Proto)
	switch r := result.(type) {
	case error:
		var substack pb.SubstackProto
		fillSubstackFromError(&substack, r)
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
	case string:
		var substack pb.SubstackProto
		fillSubstackFromAtom(&substack, AtomFromString(r))
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
	case b6.ExpressionFeature:
		// This is not perfect, as it makes original expression that
		// returned the ExpressionFeature, and the expression from the
		// feature itself, look like part of the same stack.
		// TODO: improve the UX for expression features
		substack := &pb.SubstackProto{}
		expression := api.AddPipelines(api.Simplify(b6.NewCallExpression(r.Expression(), []b6.Expression{}), d.FunctionSymbols))
		fillSubstackFromExpression(substack, expression, true)
		if len(substack.Lines) > 0 {
			response.Proto.Stack.Substacks = append(response.Proto.Stack.Substacks, substack)
		}
		id := b6.MakeCollectionID(r.ExpressionID().Namespace, r.ExpressionID().Value)
		if c := b6.FindCollectionByID(id, d.World); c != nil {
			substack := &pb.SubstackProto{}
			if err := fillSubstackFromCollection(substack, c, p, d.World); err == nil {
				p.Stack.Substacks = append(p.Stack.Substacks, substack)
			} else {
				return err
			}
		}
	case b6.Feature:
		p.Stack.Substacks = fillSubstacksFromFeature(p.Stack.Substacks, r, d.World)
		highlightInResponse(p, r.FeatureID())
	case b6.Query:
		if q, ok := api.UnparseQuery(r); ok {
			var substack pb.SubstackProto
			fillSubstackFromAtom(&substack, AtomFromString(q))
			p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		} else {
			// TODO: Improve the rendering of queries
			var substack pb.SubstackProto
			fillSubstackFromAtom(&substack, AtomFromString("Query"))
			p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		}
	case b6.Tag:
		var substack pb.SubstackProto
		fillSubstackFromAtom(&substack, AtomFromValue(r, d.World))
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		if !d.BasemapRules.IsRendered(r) {
			if q, ok := api.UnparseQuery(b6.Tagged(r)); ok {
				before := pb.MapLayerPosition_MapLayerPositionEnd
				if r.Key == "#boundary" {
					before = pb.MapLayerPosition_MapLayerPositionBuildings
				}
				p.Layers = append(p.Layers, &pb.MapLayerProto{Query: q, Before: before})
			}
		}
	case *api.HistogramCollection:
		if err := fillResponseFromHistogram(p, r, d.World); err != nil {
			return err
		}
	case b6.UntypedCollection:
		substack := &pb.SubstackProto{}
		if err := fillSubstackFromCollection(substack, r, p, d.World); err == nil {
			p.Stack.Substacks = append(p.Stack.Substacks, substack)
		} else {
			return err
		}
	case b6.Area:
		dimension := 0.0
		for i := 0; i < r.Len(); i++ {
			dimension += b6.AreaToMeters2(r.Polygon(i).Area())
		}
		atom := &pb.AtomProto{
			Atom: &pb.AtomProto_Download{
				Download: fmt.Sprintf("%.2fm² area", dimension),
			},
		}
		var substack pb.SubstackProto
		fillSubstackFromAtom(&substack, atom)
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		response.AddGeoJSON(r.ToGeoJSON())
	case b6.Path:
		dimension := b6.AngleToMeters(r.Polyline().Length())
		atom := &pb.AtomProto{
			Atom: &pb.AtomProto_Download{
				Download: fmt.Sprintf("%.2fm path", dimension),
			},
		}
		var substack pb.SubstackProto
		fillSubstackFromAtom(&substack, atom)
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		response.AddGeoJSON(r.ToGeoJSON())
	case *geojson.FeatureCollection:
		var label string
		if n := len(r.Features); n == 1 {
			label = "1 GeoJSON feature"
		} else {
			label = fmt.Sprintf("%d GeoJSON features", n)
		}
		atom := &pb.AtomProto{
			Atom: &pb.AtomProto_Download{
				Download: fmt.Sprintf(label),
			},
		}
		var substack pb.SubstackProto
		fillSubstackFromAtom(&substack, atom)
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		response.AddGeoJSON(r)
	case *geojson.Feature:
		atom := &pb.AtomProto{
			Atom: &pb.AtomProto_Download{
				Download: "GeoJSON feature",
			},
		}
		var substack pb.SubstackProto
		fillSubstackFromAtom(&substack, atom)
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		response.AddGeoJSON(r)
	case *geojson.Geometry:
		atom := &pb.AtomProto{
			Atom: &pb.AtomProto_Download{
				Download: "GeoJSON geometry",
			},
		}
		var substack pb.SubstackProto
		fillSubstackFromAtom(&substack, atom)
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		response.AddGeoJSON(geojson.NewFeatureWithGeometry(*r))
	default:
		substack := &pb.SubstackProto{
			Lines: []*pb.LineProto{{
				Line: &pb.LineProto_Value{
					Value: &pb.ValueLineProto{
						Atom: AtomFromValue(r, d.World),
					},
				},
			}},
		}
		p.Stack.Substacks = append(p.Stack.Substacks, substack)
	}
	switch r := result.(type) {
	case b6.Geometry:
		response.Proto.MapCenter = b6.NewPointProtoFromS2Point(b6.Centroid(r))
	case geojson.GeoJSON:
		response.Proto.MapCenter = b6.NewPointProtoFromS2Point(r.Centroid().ToS2Point())
	}
	return nil
}
