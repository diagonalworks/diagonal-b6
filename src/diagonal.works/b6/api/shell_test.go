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
			Begin: 0,
			End:   2,
		}},
		{"LiteralFloat", "42.0", &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_FloatValue{
						FloatValue: 42.0,
					},
				},
			},
			Begin: 0,
			End:   4,
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
			Begin: 0,
			End:   17,
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
			Begin: 0,
			End:   13,
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
			Begin: 0,
			End:   17,
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
			Begin: 0,
			End:   21,
		}},
		{"SimpleCall", `find-feature /n/6082053666`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "find-feature",
						},
						Begin: 0,
						End:   12,
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
							Begin: 13,
							End:   26,
						},
					},
				},
			},
			Begin: 0,
			End:   26,
		}},
		{"Pipeline2Stages", `find "highway=primary" | highlight`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "highlight",
						},
						Begin: 25,
						End:   34,
					},
					Args: []*pb.NodeProto{
						{
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
										{
											Node: &pb.NodeProto_Literal{
												Literal: &pb.LiteralNodeProto{
													Value: &pb.LiteralNodeProto_StringValue{
														StringValue: "highway=primary",
													},
												},
											},
											Begin: 5,
											End:   22,
										},
									},
								},
							},
							Begin: 0,
							End:   22,
						},
					},
					Pipelined: true,
				},
			},
			Begin: 0,
			End:   34,
		}},
		{"Group", `find (intersecting 19.4008, -99.1663)`, &pb.NodeProto{
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
						{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "intersecting",
										},
										Begin: 6,
										End:   18,
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
											Begin: 19,
											End:   36,
										},
									},
								},
							},
							Begin: 6,
							End:   36,
						},
					},
				},
			},
			Begin: 0,
			End:   36,
		}},
		{"FeatureID", "pair 55.614929, -2.8048709 /area/openstreetmap.org/way/115912092", &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "pair",
						},
						Begin: 0,
						End:   4,
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
							Begin: 5,
							End:   26,
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
							Begin: 27,
							End:   64,
						},
					},
				},
			},
			Begin: 0,
			End:   64,
		},
		},
		{"NestedGroups", "find (intersecting (find-area /area/openstreetmap.org/way/115912092))", &pb.NodeProto{
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
						{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "intersecting",
										},
										Begin: 6,
										End:   18,
									},
									Args: []*pb.NodeProto{
										{
											Node: &pb.NodeProto_Call{
												Call: &pb.CallNodeProto{
													Function: &pb.NodeProto{
														Node: &pb.NodeProto_Symbol{
															Symbol: "find-area",
														},
														Begin: 20,
														End:   29,
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
															Begin: 30,
															End:   67,
														},
													},
												},
											},
											Begin: 20,
											End:   67,
										},
									},
								},
							},
							Begin: 6,
							End:   67,
						},
					},
				},
			},
			Begin: 0,
			End:   67,
		}},
		{"ExplicitLambdaWithArg", `map {f -> tag f "name"} (all-areas)`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "map",
						},
						Begin: 0,
						End:   3,
					},
					Args: []*pb.NodeProto{
						{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "tag",
										},
										Begin: 10,
										End:   13,
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
											Begin: 16,
											End:   22,
										},
									},
								},
							},
							Begin: 5,
							End:   22,
						},
						{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "all-areas",
										},
										Begin: 25,
										End:   34,
									},
								},
							},
							Begin: 25,
							End:   34,
						},
					},
				},
			},
			Begin: 0,
			End:   34,
		}},
		{"ImplicitLambda", `map (tag "name") (all-areas)`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "map",
						},
						Begin: 0,
						End:   3,
					},
					Args: []*pb.NodeProto{
						{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "tag",
										},
										Begin: 5,
										End:   8,
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
											Begin: 9,
											End:   15,
										},
									},
								},
							},
							Begin: 5,
							End:   15,
						},
						{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "all-areas",
										},
										Begin: 18,
										End:   27,
									},
								},
							},
							Begin: 18,
							End:   27,
						},
					},
				},
			},
			Begin: 0,
			End:   27,
		}},
		{"ExplicitLambdaWithoutArgs", `with-change {-> building-access}`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "with-change",
						},
						Begin: 0,
						End:   11,
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
													Begin: 16,
													End:   31,
												},
											},
										},
										Begin: 16,
										End:   31,
									},
								},
							},
							Begin: 16,
							End:   31,
						},
					},
				},
			},
			Begin: 0,
			End:   31,
		}},
		{"RootCallWithoutArgs", `all-areas`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "all-areas",
						},
						Begin: 0,
						End:   9,
					},
				},
			},
			Begin: 0,
			End:   9,
		}},
		{"PipeineWithExplicitLambda", `all-areas | {a -> highlight a}`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "highlight",
						},
						Begin: 18,
						End:   27,
					},
					Args: []*pb.NodeProto{
						&pb.NodeProto{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "all-areas",
										},
										Begin: 0,
										End:   9,
									},
								},
							},
							Begin: 0,
							End:   9,
						},
					},
					Pipelined: true,
				},
			},
			Begin: 0,
			End:   29,
		}},
		{"Pipeline3Stages", `all-areas | filter | highlight`, &pb.NodeProto{
			Node: &pb.NodeProto_Call{
				Call: &pb.CallNodeProto{
					Function: &pb.NodeProto{
						Node: &pb.NodeProto_Symbol{
							Symbol: "highlight",
						},
						Begin: 21,
						End:   30,
					},
					Args: []*pb.NodeProto{
						&pb.NodeProto{
							Node: &pb.NodeProto_Call{
								Call: &pb.CallNodeProto{
									Function: &pb.NodeProto{
										Node: &pb.NodeProto_Symbol{
											Symbol: "filter",
										},
										Begin: 12,
										End:   18,
									},
									Args: []*pb.NodeProto{
										&pb.NodeProto{
											Node: &pb.NodeProto_Call{
												Call: &pb.CallNodeProto{
													Function: &pb.NodeProto{
														Node: &pb.NodeProto_Symbol{
															Symbol: "all-areas",
														},
														Begin: 0,
														End:   9,
													},
												},
											},
											Begin: 0,
											End:   9,
										},
									},
									Pipelined: true,
								},
							},
							Begin: 0,
							End:   18,
						},
					},
					Pipelined: true,
				},
			},
			Begin: 0,
			End:   30,
		}},
		{"QueryTagWithoutValue", `find [#building]`, &pb.NodeProto{
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
											Query: &pb.QueryProto_Keyed{
												Keyed: "#building",
											},
										},
									},
								},
							},
							Begin: 6,
							End:   15,
						},
					},
				},
			},
			Begin: 0,
			End:   15,
		}},
		{"QueryNested", `find [#building=yes & [#shop=supermarket | #shop=convenience]]`, &pb.NodeProto{
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
		}},
	}

	functions := FunctionArgCounts{
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
		t.Errorf("Expected no error, found: %s", err)
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
