package ui

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	pb "diagonal.works/b6/proto"
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

func lineFromTags(f b6.Feature) *pb.LineProto {
	tags := f.AllTags()
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
		case b6.Geometry:
			switch v.GeometryType() {
			case b6.GeometryTypePoint:
				ll := v.Location()
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

func fillSubstacksFromFeature(substacks []*pb.SubstackProto, f b6.Feature, w b6.World) []*pb.SubstackProto {
	substack := &pb.SubstackProto{}
	substack.Lines = append(substack.Lines, ValueLineFromValue(f, w))
	if len(f.AllTags()) > 0 {
		substack.Lines = append(substack.Lines, lineFromTags(f))
	}
	substacks = append(substacks, substack)

	if point, ok := f.(b6.PhysicalFeature); ok && point.GeometryType() == b6.GeometryTypePoint {
		paths := b6.AllPaths(w.FindPathsByPoint(point.FeatureID()))
		substack := &pb.SubstackProto{Collapsable: true}
		line := leftRightValueLineFromValues("Paths", len(paths), w)
		substack.Lines = append(substack.Lines, line)
		for _, path := range paths {
			substack.Lines = append(substack.Lines, ValueLineFromValue(path, w))
		}
		substacks = append(substacks, substack)
	}

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
