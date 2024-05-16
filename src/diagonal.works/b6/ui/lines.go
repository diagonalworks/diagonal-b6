package ui

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/geojson"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/renderer"
	"github.com/golang/geo/s2"
)

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

func lineFromTags(tags []b6.Tag, f b6.Feature) *pb.LineProto {
	tl := &pb.TagsLineProto{
		Tags: make([]*pb.TagAtomProto, len(tags)),
	}
	for i, tag := range tags {
		if strings.HasPrefix(tag.Key, "#") || strings.HasPrefix(tag.Key, "#") {
			tl.Tags[i] = &pb.TagAtomProto{Prefix: tag.Key[0:1], Key: tag.Key[1:], Value: tag.Value.String()}
		} else {
			tl.Tags[i] = &pb.TagAtomProto{Prefix: "", Key: tag.Key, Value: tag.Value.String()}
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
	if i, ok := b6.ToInt(value); ok {
		return AtomFromString(strconv.Itoa(i))
	} else if f, err := b6.ToFloat64(value); err == nil {
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
		case b6.Geometry:
			switch v.GeometryType() {
			case b6.GeometryTypePoint:
				ll := s2.LatLngFromPoint(v.Point())
				return AtomFromString(fmt.Sprintf("%f,%f", ll.Lat.Degrees(), ll.Lng.Degrees()))
			case b6.GeometryTypePath:
				return AtomFromString("Path")
			case b6.GeometryTypeArea:
				return AtomFromString("Area")
			}
			return AtomFromString("Geometry")
		case b6.Tag:
			return AtomFromString(api.UnparseTag(v))
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

func fillSubstacksFromFeature(response *UIResponseJSON, substacks []*pb.SubstackProto, f b6.Feature, w b6.World) []*pb.SubstackProto {
	substack := &pb.SubstackProto{}
	substack.Lines = append(substack.Lines, ValueLineFromValue(f, w))
	substacks = append(substacks, substack)
	substacks = fillSubstacksFromHistogramReferences(response, substacks, f, w)

	if area, ok := f.(b6.AreaFeature); ok {
		if b := f.Get("#building"); b.IsValid() {
			substacks = fillSubstacksFromContainedAmenities(substacks, area, w)
		}
	}

	if tags := f.AllTags(); len(tags) > 0 {
		substack := &pb.SubstackProto{Collapsable: true}
		line := leftRightValueLineFromValues("Tags", len(tags), w)
		substack.Lines = append(substack.Lines, line, lineFromTags(tags, f))
		substacks = append(substacks, substack)
	}

	if point, ok := f.(b6.PhysicalFeature); ok && point.GeometryType() == b6.GeometryTypePoint {
		paths := b6.AllFeatures(w.FindReferences(point.FeatureID(), b6.FeatureTypePath))
		substack := &pb.SubstackProto{Collapsable: true}
		line := leftRightValueLineFromValues("Paths", len(paths), w)
		substack.Lines = append(substack.Lines, line)
		for _, path := range paths {
			substack.Lines = append(substack.Lines, ValueLineFromValue(path, w))
		}
		substacks = append(substacks, substack)
	}

	if path, ok := f.(b6.NestedPhysicalFeature); ok && path.GeometryType() == b6.GeometryTypePath {
		substack := &pb.SubstackProto{Collapsable: true}
		line := leftRightValueLineFromValues("Points", path.GeometryLen(), w)
		substack.Lines = append(substack.Lines, line)
		for i := 0; i < path.GeometryLen(); i++ {
			if point := path.Feature(i); point != nil {
				substack.Lines = append(substack.Lines, ValueLineFromValue(point, w))
			} else {
				substack.Lines = append(substack.Lines, ValueLineFromValue(path.PointAt(i), w))
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

func fillSubstacksFromContainedAmenities(substacks []*pb.SubstackProto, f b6.AreaFeature, w b6.World) []*pb.SubstackProto {
	q := b6.Intersection{
		b6.Union{
			b6.Keyed{Key: "#amenity"},
			b6.Keyed{Key: "#leisure"},
			b6.Keyed{Key: "#shop"},
		},
		b6.IntersectsFeature{
			ID: f.FeatureID(),
		},
	}

	substack := &pb.SubstackProto{
		Collapsable: true,
		Lines:       []*pb.LineProto{nil}, // Reserved for header
	}
	i := w.FindFeatures(q)
	for i.Next() {
		if i.FeatureID() != f.FeatureID() {
			substack.Lines = append(substack.Lines, ValueLineFromValue(i.Feature(), w))
		}
	}
	if len(substack.Lines) > 1 {
		substack.Lines[0] = leftRightValueLineFromValues("Within building", len(substack.Lines)-1, w)
		substacks = append(substacks, substack)
	}
	return substacks
}

func fillSubstacksFromHistogramReferences(response *UIResponseJSON, substacks []*pb.SubstackProto, f b6.Feature, w b6.World) []*pb.SubstackProto {
	i := w.FindReferences(f.FeatureID(), b6.FeatureTypeCollection)
	destinations := make(map[b6.FeatureID]struct{})
	for i.Next() {
		c := i.Feature().(b6.CollectionFeature)
		if t := c.Get("b6"); t.IsValid() && t.Value.String() == "histogram" {
			if label := c.Get("b6:label"); label.IsValid() {
				substack := &pb.SubstackProto{Collapsable: true}
				if b6.CanAdaptCollection[b6.FeatureID, int](c) {
					if value, ok := c.FindValue(f.FeatureID()); ok {
						line := leftRightValueLineFromValues(label.Value.String(), value, w)
						substack.Lines = append(substack.Lines, line)
					}
				} else if b6.CanAdaptCollection[b6.FeatureID, b6.FeatureID](c) {
					values := c.FindValues(f.FeatureID(), []any{})
					count := 0
					substack.Lines = []*pb.LineProto{nil}
					for _, v := range values {
						if v.(b6.FeatureID).IsValid() {
							destinations[v.(b6.FeatureID)] = struct{}{}
							count++
							substack.Lines = append(substack.Lines, ValueLineFromValue(v, w))
						}
					}
					substack.Lines[0] = leftRightValueLineFromValues(label.Value.String(), count, w)
				}
				substacks = append(substacks, substack)
			}
		}
	}
	fillResponseFromDestinations(response, destinations, w)
	return substacks
}

func fillResponseFromDestinations(response *UIResponseJSON, destinations map[b6.FeatureID]struct{}, w b6.World) {
	g := geojson.NewFeatureCollection()
	for d := range destinations {
		if f := w.FindFeatureByID(d); f != nil {
			if p, ok := f.(b6.PhysicalFeature); ok {
				if centroid, ok := b6.Centroid(p); ok {
					gf := geojson.NewFeatureFromS2Point(centroid)
					icon, _ := renderer.IconForFeature(f)
					gf.Properties["-b6-icon"] = icon
					g.AddFeature(gf)
				}
			}
		}
	}
	if len(g.Features) > 0 {
		response.AddGeoJSON(g)
	}
}

func SortableKeyForAtom(a *pb.AtomProto) string {
	switch a := a.Atom.(type) {
	case *pb.AtomProto_Value:
		return "0 " + sortableKeyForString(a.Value)
	case *pb.AtomProto_LabelledIcon:
		return "0 " + sortableKeyForString(a.LabelledIcon.Label) + sortableKeyForString(a.LabelledIcon.Icon)
	case *pb.AtomProto_Download:
		return "0 " + sortableKeyForString(a.Download)
	case *pb.AtomProto_Chip:
		if len(a.Chip.Labels) > 0 {
			return "1 " + sortableKeyForString(a.Chip.Labels[0])
		}
	case *pb.AtomProto_Conditional:
		if len(a.Conditional.Atoms) > 0 {
			return SortableKeyForAtom(a.Conditional.Atoms[0])
		}
	}
	return "2"
}

func sortableKeyForString(s string) string {
	if i, err := strconv.Atoi(s); err == nil {
		return fmt.Sprintf("%010d", i)
	}
	return s
}
