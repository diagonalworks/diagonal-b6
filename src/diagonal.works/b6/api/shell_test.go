package api

import (
	"reflect"
	"strings"
	"testing"

	"diagonal.works/b6"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/test/camden"

	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

func zeroBeginAndEndLocations(n *pb.NodeProto) {
	n.Begin = 0
	n.End = 0
	switch n := n.Node.(type) {
	case *pb.NodeProto_Call:
		zeroBeginAndEndLocations(n.Call.Function)
		for _, arg := range n.Call.Args {
			zeroBeginAndEndLocations(arg)
		}
	case *pb.NodeProto_Lambda_:
		zeroBeginAndEndLocations(n.Lambda_.Node)
	}
}

type testFunctionArgCounts map[string]int

func (fs testFunctionArgCounts) ArgCount(symbol string) (int, bool) {
	n, ok := fs[symbol]
	return n, ok
}

func (fs testFunctionArgCounts) IsVariadic(symbol string) (bool, bool) {
	_, ok := fs[symbol]
	return false, ok
}

func TestParseExpression(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		top        *pb.NodeProto
	}{
		{"LiteralInt", "42", &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_IntValue{
						IntValue: 42,
					},
				},
			},
		}},
		{"LiteralFloat", "42.0", &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_FloatValue{
						FloatValue: 42.0,
					},
				},
			},
		}},
		{"LiteralLatLng", `19.4008, -99.1663`, &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_PointValue{
						PointValue: &pb.PointProto{
							LatE7: 194008000,
							LngE7: -991663000,
						},
					},
				},
			},
		}},
		{"LiteralTag", `#highway=path`, &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_TagValue{
						TagValue: &pb.TagProto{
							Key:   "#highway",
							Value: "path",
						},
					},
				},
			},
		}},
		{"LiteralSearchableTagWithToken", `#nhs:hospital=yes`, &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_TagValue{
						TagValue: &pb.TagProto{
							Key:   "#nhs:hospital",
							Value: "yes",
						},
					},
				},
			},
		}},
		{"LiteralTagWithQuotes", `name="The Lighterman"`, &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_TagValue{
						TagValue: &pb.TagProto{
							Key:   "name",
							Value: "The Lighterman",
						},
					},
				},
			},
		}},
		{"SimpleCall", `find-feature /n/6082053666`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "find-feature",
						},
					},
					Args: []*pb.NodeProto{
						{
							Node: &pb.NodeProto_Literal{
								Literal: &pb.LiteralNodeProto{
									Value: &pb.LiteralNodeProto_FeatureIDValue{
										FeatureIDValue: &pb.FeatureIDProto{
											Type:      pb.FeatureType_FeatureTypePoint,
											Namespace: b6.NamespaceOSMNode.String(),
											Value:     6082053666,
										},
									},
								},
							},
						},
					},
				},
			},
		}},
		{"Pipeline2Stages", `find "highway=primary" | highlight`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "highlight",
						},
					},
					Args: []*pb.NodeProto{
						{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "find",
										},
									},
									Args: []*pb.NodeProto{
										{
											Node: &pb.NodeProto_Literal{
												Literal: &pb.LiteralNodeProto{
													Value: &pb.LiteralNodeProto_StringValue{
														StringValue: "highway=primary",
													},
												},
											},
										},
									},
								},
							},
						},
					},
					Pipelined: true,
				},
			},
		}},
		{"Group", `find (intersecting 19.4008, -99.1663)`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "find",
						},
					},
					Args: []*pb.NodeProto{
						{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "intersecting",
										},
									},
									Args: []*pb.NodeProto{
										{
											Node: &pb.NodeProto_Literal{
												Literal: &pb.LiteralNodeProto{
													Value: &pb.LiteralNodeProto_PointValue{
														PointValue: &pb.PointProto{
															LatE7: 194008000,
															LngE7: -991663000,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}},
		{"FeatureID", "pair 55.614929, -2.8048709 /area/openstreetmap.org/way/115912092", &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "pair",
						},
					},
					Args: []*pb.NodeProto{
						{
							Node: &pb.NodeProto_Literal{
								Literal: &pb.LiteralNodeProto{
									Value: &pb.LiteralNodeProto_PointValue{
										PointValue: &pb.PointProto{
											LatE7: 556149290,
											LngE7: -28048709,
										},
									},
								},
							},
						},
						{
							Node: &pb.NodeProto_Literal{
								Literal: &pb.LiteralNodeProto{
									Value: &pb.LiteralNodeProto_FeatureIDValue{
										FeatureIDValue: &pb.FeatureIDProto{
											Type:      pb.FeatureType_FeatureTypeArea,
											Namespace: "openstreetmap.org/way",
											Value:     115912092,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		},
		{"NestedGroups", "find (intersecting (find-area /area/openstreetmap.org/way/115912092))", &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "find",
						},
					},
					Args: []*pb.NodeProto{
						{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "intersecting",
										},
									},
									Args: []*pb.NodeProto{
										{
											Node: &pb.NodeProto_Call{
												Call: &pb.CallNodeProto{
													Function: &pb.NodeProto{
														Node: &pb.NodeProto_Symbol{
															Symbol: "find-area",
														},
													},
													Args: []*pb.NodeProto{
														{
															Node: &pb.NodeProto_Literal{
																Literal: &pb.LiteralNodeProto{
																	Value: &pb.LiteralNodeProto_FeatureIDValue{
																		FeatureIDValue: &pb.FeatureIDProto{
																			Type:      pb.FeatureType_FeatureTypeArea,
																			Namespace: "openstreetmap.org/way",
																			Value:     115912092,
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}},
		{"ExplicitLambdaWithArg", `map {f -> tag f "name"} (all-areas)`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "map",
						},
					},
					Args: []*pb.NodeProto{
						{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "tag",
										},
									},
									Args: []*pb.NodeProto{
										{
											Node: &pb.NodeProto_Literal{
												Literal: &pb.LiteralNodeProto{
													Value: &pb.LiteralNodeProto_StringValue{
														StringValue: "name",
													},
												},
											},
										},
									},
								},
							},
						},
						{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "all-areas",
										},
									},
								},
							},
						},
					},
				},
			},
		}},
		{"ImplicitLambda", `map (tag "name") (all-areas)`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "map",
						},
					},
					Args: []*pb.NodeProto{
						{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "tag",
										},
									},
									Args: []*pb.NodeProto{
										{
											Node: &pb.NodeProto_Literal{
												Literal: &pb.LiteralNodeProto{
													Value: &pb.LiteralNodeProto_StringValue{
														StringValue: "name",
													},
												},
											},
										},
									},
								},
							},
						},
						{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "all-areas",
										},
									},
								},
							},
						},
					},
				},
			},
		}},
		{"ExplicitLambdaWithoutArgs", `with-change {-> building-access}`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "with-change",
						},
					},
					Args: []*pb.NodeProto{
						{
							Node: &pb.NodeProto_Lambda_{
								Lambda_: &pb.LambdaNodeProto{
									Args: []string{},
									Node: &pb.NodeProto{
										Node: &pb.NodeProto_Call{
											Call: &pb.CallNodeProto{
												Function: &pb.NodeProto{
													Node: &pb.NodeProto_Symbol{
														Symbol: "building-access",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}},
		{"RootCallWithoutArgs", `all-areas`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "all-areas",
						},
					},
				},
			},
		}},
		{"PipeineWithExplicitLambda", `all-areas | {a -> highlight a}`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "highlight",
						},
					},
					Args: []*pb.NodeProto{
						&pb.NodeProto{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "all-areas",
										},
									},
								},
							},
						},
					},
					Pipelined: true,
				},
			},
		}},
		{"Pipeline3Stages", `all-areas | filter | highlight`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "highlight",
						},
					},
					Args: []*pb.NodeProto{
						&pb.NodeProto{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "filter",
										},
									},
									Args: []*pb.NodeProto{
										&pb.NodeProto{
											Node: &pb.NodeProto_Call{
												Call: &pb.CallNodeProto{
													Function: &pb.NodeProto{
														Node: &pb.NodeProto_Symbol{
															Symbol: "all-areas",
														},
													},
												},
											},
										},
									},
									Pipelined: true,
								},
							},
						},
					},
					Pipelined: true,
				},
			},
		}},
		{"QueryTagWithoutValue", `find [#building]`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "find",
						},
					},
					Args: []*pb.NodeProto{
						&pb.NodeProto{
							Node: &pb.NodeProto_Literal{
								Literal: &pb.LiteralNodeProto{
									Value: &pb.LiteralNodeProto_QueryValue{
										QueryValue: &pb.QueryProto{
											Query: &pb.QueryProto_Keyed{
												Keyed: "#building",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}},
		{"QueryNested", `find [#building=yes & [#shop=supermarket | #shop=convenience]]`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "find",
						},
					},
					Args: []*pb.NodeProto{
						&pb.NodeProto{
							Node: &pb.NodeProto_Literal{
								Literal: &pb.LiteralNodeProto{
									Value: &pb.LiteralNodeProto_QueryValue{
										QueryValue: &pb.QueryProto{
											Query: &pb.QueryProto_Intersection{
												Intersection: &pb.QueriesProto{
													Queries: []*pb.QueryProto{
														&pb.QueryProto{
															Query: &pb.QueryProto_Tagged{
																Tagged: &pb.TagProto{
																	Key:   "#building",
																	Value: "yes",
																},
															},
														},
														&pb.QueryProto{
															Query: &pb.QueryProto_Union{
																Union: &pb.QueriesProto{
																	Queries: []*pb.QueryProto{
																		&pb.QueryProto{
																			Query: &pb.QueryProto_Tagged{
																				Tagged: &pb.TagProto{
																					Key:   "#shop",
																					Value: "supermarket",
																				},
																			},
																		},
																		&pb.QueryProto{
																			Query: &pb.QueryProto_Tagged{
																				Tagged: &pb.TagProto{
																					Key:   "#shop",
																					Value: "convenience",
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}},
		{"CollectionLiteral", `{"motorway": 36.0, "primary": 32.0}`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "collection",
						},
					},
					Args: []*pb.NodeProto{
						&pb.NodeProto{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "pair",
										},
									},
									Args: []*pb.NodeProto{
										&pb.NodeProto{
											Node: &pb.NodeProto_Literal{
												Literal: &pb.LiteralNodeProto{
													Value: &pb.LiteralNodeProto_StringValue{
														StringValue: "motorway",
													},
												},
											},
										},
										&pb.NodeProto{
											Node: &pb.NodeProto_Literal{
												Literal: &pb.LiteralNodeProto{
													Value: &pb.LiteralNodeProto_FloatValue{
														FloatValue: 36.0,
													},
												},
											},
										},
									},
								},
							},
						},
						&pb.NodeProto{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "pair",
										},
									},
									Args: []*pb.NodeProto{
										&pb.NodeProto{
											Node: &pb.NodeProto_Literal{
												Literal: &pb.LiteralNodeProto{
													Value: &pb.LiteralNodeProto_StringValue{
														StringValue: "primary",
													},
												},
											},
										},
										&pb.NodeProto{
											Node: &pb.NodeProto_Literal{
												Literal: &pb.LiteralNodeProto{
													Value: &pb.LiteralNodeProto_FloatValue{
														FloatValue: 32.0,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}},
		{"CollectionLiteralWithImplicitKeys", `{"motorway", "primary"}`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "collection",
						},
					},
					Args: []*pb.NodeProto{
						&pb.NodeProto{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "pair",
										},
									},
									Args: []*pb.NodeProto{
										&pb.NodeProto{
											Node: &pb.NodeProto_Literal{
												Literal: &pb.LiteralNodeProto{
													Value: &pb.LiteralNodeProto_IntValue{
														IntValue: 0,
													},
												},
											},
										},
										&pb.NodeProto{
											Node: &pb.NodeProto_Literal{
												Literal: &pb.LiteralNodeProto{
													Value: &pb.LiteralNodeProto_StringValue{
														StringValue: "motorway",
													},
												},
											},
										},
									},
								},
							},
						},
						&pb.NodeProto{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "pair",
										},
									},
									Args: []*pb.NodeProto{
										&pb.NodeProto{
											Node: &pb.NodeProto_Literal{
												Literal: &pb.LiteralNodeProto{
													Value: &pb.LiteralNodeProto_IntValue{
														IntValue: 1,
													},
												},
											},
										},
										&pb.NodeProto{
											Node: &pb.NodeProto_Literal{
												Literal: &pb.LiteralNodeProto{
													Value: &pb.LiteralNodeProto_StringValue{
														StringValue: "primary",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}},
	}

	functions := testFunctionArgCounts{
		"all-areas":    0,
		"area":         1,
		"count-values": 1,
		"find-area":    1,
		"find":         1,
		"find-feature": 1,
		"filter":       1,
		"gt":           1,
		"highlight":    1,
		"intersecting": 1,
		"map":          2,
		"pair":         2,
		"tag":          2,
	}
	for _, test := range tests {
		f := func(t *testing.T) {
			top, err := ParseExpression(test.expression)
			top = Simplify(top, functions)
			zeroBeginAndEndLocations(top) // Locations are tested separately below
			if err == nil {
				if !proto.Equal(top, test.top) {
					t.Errorf("%s: expected %s, found %s", test.expression, prototext.Format(test.top), prototext.Format(top))
				}
			} else {
				t.Errorf("%s: expected no error, found %s", test.expression, err)
			}
		}
		t.Run(test.name, f)
	}
}
func TestParseExpressionFillsBeginAndEndLocations(t *testing.T) {
	expression := `find [#building=yes & [#shop=supermarket | #shop=convenience]]`
	expected := &pb.NodeProto{
		Node: &pb.NodeProto_Call{
			Call: &pb.CallNodeProto{
				Function: &pb.NodeProto{
					Node: &pb.NodeProto_Symbol{
						Symbol: "find",
					},
					Begin: 0,
					End:   4,
				},
				Args: []*pb.NodeProto{
					&pb.NodeProto{
						Node: &pb.NodeProto_Literal{
							Literal: &pb.LiteralNodeProto{
								Value: &pb.LiteralNodeProto_QueryValue{
									QueryValue: &pb.QueryProto{
										Query: &pb.QueryProto_Intersection{
											Intersection: &pb.QueriesProto{
												Queries: []*pb.QueryProto{
													&pb.QueryProto{
														Query: &pb.QueryProto_Tagged{
															Tagged: &pb.TagProto{
																Key:   "#building",
																Value: "yes",
															},
														},
													},
													&pb.QueryProto{
														Query: &pb.QueryProto_Union{
															Union: &pb.QueriesProto{
																Queries: []*pb.QueryProto{
																	&pb.QueryProto{
																		Query: &pb.QueryProto_Tagged{
																			Tagged: &pb.TagProto{
																				Key:   "#shop",
																				Value: "supermarket",
																			},
																		},
																	},
																	&pb.QueryProto{
																		Query: &pb.QueryProto_Tagged{
																			Tagged: &pb.TagProto{
																				Key:   "#shop",
																				Value: "convenience",
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						Begin: 6,
						End:   60,
					},
				},
			},
		},
		Begin: 0,
		End:   60,
	}
	top, err := ParseExpression(expression)
	if err == nil {
		if !proto.Equal(top, expected) {
			t.Errorf("%s: expected %s, found %s", expression, prototext.Format(expected), prototext.Format(top))
		}
	} else {
		t.Errorf("%s: expected no error, found %s", expression, err)
	}
}

func TestParseExpressionFailsWithArgOutsideLambda(t *testing.T) {
	if top, err := ParseExpression("all-areas | highlight $"); err == nil {
		t.Errorf("Expected parsing to fail, found %+v", top)
	} else if strings.Index(err.Error(), "$") < 0 {
		t.Errorf("Expected error to mention $, found %q", err)
	}
}

func TestOrderTokens(t *testing.T) {
	e := `show-accessibility [#building=school] 900 "walking" {b -> pair "#building" (building-category b)}`
	top, err := ParseExpression(e)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	tokens := OrderTokens(top)
	expected := []*pb.NodeProto{
		{
			Node: &pb.NodeProto_Symbol{
				Symbol: "show-accessibility",
			},
			Begin: 0,
			End:   18,
		},
		{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_QueryValue{
						QueryValue: &pb.QueryProto{
							Query: &pb.QueryProto_Tagged{
								Tagged: &pb.TagProto{
									Key:   "#building",
									Value: "school",
								},
							},
						},
					},
				},
			},
			Begin: 20,
			End:   36,
		},
		{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_IntValue{
						IntValue: 900,
					},
				},
			},
			Begin: 38,
			End:   41,
		},
		{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_StringValue{
						StringValue: "walking",
					},
				},
			},
			Begin: 42,
			End:   51,
		},
		{
			Node: &pb.NodeProto_Symbol{
				Symbol: "pair",
			},
			Begin: 58,
			End:   62,
		},
		{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_StringValue{
						StringValue: "#building",
					},
				},
			},
			Begin: 63,
			End:   74,
		},
		{
			Node: &pb.NodeProto_Symbol{
				Symbol: "building-category",
			},
			Begin: 76,
			End:   93,
		},
		{
			Node: &pb.NodeProto_Symbol{
				Symbol: "b",
			},
			Begin: 94,
			End:   95,
		},
	}
	if !reflect.DeepEqual(tokens, expected) {
		t.Errorf("Expected %s, found %s", expected, tokens)
	}
}

func TestToFeatureIDExpression(t *testing.T) {
	tests := []struct {
		ID    b6.FeatureID
		Token string
	}{
		{camden.StableStreetBridgeID.FeatureID(), "/w/140633010"},
		{camden.LightermanID.FeatureID(), "/a/427900370"},
		{b6.MakePointID(b6.NamespaceGBUPRN, 116000008).FeatureID(), "/gb/uprn/116000008"},
		{b6.FeatureIDFromUKONSCode("E01000953", 2011, b6.FeatureTypeArea).FeatureID(), "/uk/ons/2011/E01000953"},
	}
	for _, test := range tests {
		if token := UnparseFeatureID(test.ID, true); token != test.Token {
			t.Errorf("Expected token %q for %s, found %q", test.Token, test.ID, token)
		}
		if id, err := ParseFeatureIDToken(test.Token); err != nil || id != test.ID {
			t.Errorf("Expected id %s for %q, found %s", test.ID, test.Token, id)
		}
	}
}

func TestUnparseExpression(t *testing.T) {
	tests := []string{
		"42",
		"/w/140633010",
		"[#amenity=cafe]",
		"[#amenity=cafe | #amenity=restaurant]",
		"area (find-feature /a/427900370)",
		"find-feature /a/427900370 | area",
		"find [#place=uprn] | filter {u -> gt (all-tags u | count) 1}",
	}
	for _, test := range tests {
		if node, err := ParseExpression(test); err == nil {
			if roundtrip, ok := UnparseNode(node); !ok || roundtrip != test {
				t.Errorf("Expected %q, found %q", test, roundtrip)
			}
		} else {
			t.Errorf("Failed to parse %q: %s", test, err)
		}
	}
}
