package language

import (
	//"encoding/json"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
	"unicode/utf8"
)

var KITCHEN_SINK = `
# Copyright (c) 2015, Facebook, Inc.
# All rights reserved.
#
# This source code is licensed under the BSD-style license found in the
# LICENSE file in the root directory of this source tree. An additional grant
# of patent rights can be found in the PATENTS file in the same directory.

query queryName($foo: ComplexType, $site: Site = MOBILE) {
  whoever123is: node(id: [123, 456]) {
    id ,
    ... on User @defer {
      field2 {
        id ,
        alias: field1(first:10, after:$foo,) @include(if: $foo) {
          id,
          ...frag
        }
      }
    }
    ... @skip(unless: $foo) {
      id
    }
    ... {
      id
    }
  }
}

mutation likeStory {
  like(story: 123) @defer {
    story {
      id
    }
  }
}

subscription StoryLikeSubscription($input: StoryLikeSubscribeInput) {
  storyLikeSubscribe(input: $input) {
    story {
      likers {
        count
      }
      likeSentence {
        text
      }
    }
  }
}

fragment frag on Friend {
  foo(size: $size, bar: $b, obj: {key: "value"})
}

{
  unnamed(truthy: true, falsey: false),
  query
}
`

func createLOCFn(body string) func(start int, end int) *LOC {
	lines := strings.Split(body, "\n")
	return func(start int, end int) *LOC {
		startIndex := 0
		endIndex := 0
		startCol := 1
		endCol := 1
		startLine := 1
		endLine := 1
		index := start
		for {
			lineLen := utf8.RuneCountInString(lines[startLine-1])
			if lineLen > index {
				startCol = index + 1
				line := lines[startLine-1]
				for index > 0 {
					index--
					_, width := utf8.DecodeRuneInString(line)
					startIndex += width
					line = line[width:]
				}
				break
			} else {
				index -= lineLen + 1
				startIndex += len(lines[startLine-1]) + 1
				startLine++
			}
		}
		index = end
		for {
			lineLen := utf8.RuneCountInString(lines[endLine-1])
			if lineLen >= index {
				endCol = index + 1
				line := lines[endLine-1]
				for index > 0 {
					index--
					_, width := utf8.DecodeRuneInString(line)
					endIndex += width
					line = line[width:]
				}
				break
			} else {
				index -= lineLen + 1
				endIndex += len(lines[endLine-1]) + 1
				endLine++
			}
		}
		return &LOC{
			Source: body,
			Start: &Position{
				Index:  startIndex,
				Line:   startLine,
				Column: startCol,
			},
			End: &Position{
				Index:  endIndex,
				Line:   endLine,
				Column: endCol,
			},
		}
	}
}

func typeNode(name string, loc *LOC) *NamedType {
	return &NamedType{
		Name: nameNode(name, loc),
		LOC:  loc,
	}
}

func nameNode(name string, loc *LOC) *Name {
	return &Name{
		Value: name,
		LOC:   loc,
	}
}

func fieldNode(name *Name, ntype ASTNode, loc *LOC) *FieldDefinition {
	return fieldNodeWithArgs(name, ntype, nil, loc)
}

func fieldNodeWithArgs(name *Name, ntype ASTNode, args []*InputValueDefinition, loc *LOC) *FieldDefinition {
	argsIndex := map[string]*InputValueDefinition{}
	if args != nil {
		for _, arg := range args {
			argsIndex[arg.Name.Value] = arg
		}
	} else {
		argsIndex = nil
	}
	return &FieldDefinition{
		Name:          name,
		Arguments:     args,
		ArgumentIndex: argsIndex,
		Type:          ntype,
		LOC:           loc,
	}

}

func enumValueNode(name string, loc *LOC) *EnumValueDefinition {
	return &EnumValueDefinition{
		Name: nameNode(name, loc),
		LOC:  loc,
	}
}

func inputValueNode(name *Name, ntype ASTNode, defaultValue ASTNode, loc *LOC) *InputValueDefinition {
	return &InputValueDefinition{
		Name:         name,
		Type:         ntype,
		DefaultValue: defaultValue,
		LOC:          loc,
	}
}

func TestParser(t *testing.T) {

	Convey("Parser", t, func() {

		var result *Document
		var err error
		parser := &Parser{}

		Convey("accepts option to no include source", func() {
			result, err = parser.Parse(&ParseParams{
				Source:   `{ field }`,
				NoSource: true,
			})
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, &Document{
				Definitions: []ASTNode{
					&OperationDefinition{
						Operation: "query",
						SelectionSet: &SelectionSet{
							Selections: []ASTNode{
								&Field{
									Name: &Name{
										Value: "field",
										LOC: &LOC{
											Start: &Position{
												Index:  2,
												Line:   1,
												Column: 3,
											},
											End: &Position{
												Index:  7,
												Line:   1,
												Column: 8,
											},
										},
									},
									LOC: &LOC{
										Start: &Position{
											Index:  2,
											Line:   1,
											Column: 3,
										},
										End: &Position{
											Index:  7,
											Line:   1,
											Column: 8,
										},
									},
								},
							},
							LOC: &LOC{
								Start: &Position{
									Index:  0,
									Line:   1,
									Column: 1,
								},
								End: &Position{
									Index:  9,
									Line:   1,
									Column: 10,
								},
							},
						},
						LOC: &LOC{
							Start: &Position{
								Index:  0,
								Line:   1,
								Column: 1,
							},
							End: &Position{
								Index:  9,
								Line:   1,
								Column: 10,
							},
						},
					},
				},
				LOC: &LOC{
					Start: &Position{
						Index:  0,
						Line:   1,
						Column: 1,
					},
					End: &Position{
						Index:  9,
						Line:   1,
						Column: 10,
					},
				},
			})
		})

		Convey("parse provides useful errors", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `{`,
			})
			So(result, ShouldEqual, nil)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:2) Expected a selection or fragment spread, found EOF\n\n1|{\n   ^")

			result, err = parser.Parse(&ParseParams{
				Source: `{ ...MissingOn }
fragment MissingOn Type`,
			})
			So(result, ShouldEqual, nil)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (2:20) Expected \"on\", found Name \"Type\"\n\n1|{ ...MissingOn }\n2|fragment MissingOn Type\n                     ^^^^")

			result, err = parser.Parse(&ParseParams{
				Source: `{ field: {} }`,
			})
			So(result, ShouldEqual, nil)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:10) Expected Name, found {\n\n1|{ field: {} }\n           ^")

			result, err = parser.Parse(&ParseParams{
				Source: `notanoperation Foo { field }`,
			})
			So(result, ShouldEqual, nil)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:1) Unexpected Name \"notanoperation\"\n\n1|notanoperation Foo { field }\n  ^^^^^^^^^^^^^^")

			result, err = parser.Parse(&ParseParams{
				Source: `...`,
			})
			So(result, ShouldEqual, nil)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:1) Unexpected ...\n\n1|...\n  ^^^")

			result, err = parser.Parse(&ParseParams{
				Source: `query`,
			})
			So(result, ShouldEqual, nil)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:6) Expected {, found EOF\n\n1|query\n       ^")

		})

		Convey("parses variable inline values", func() {
			_, err := parser.Parse(&ParseParams{
				Source: `{ field(complex: { a: { b: [ $var ] } }) }`,
			})
			So(err, ShouldEqual, nil)
		})

		Convey("parses constant default values", func() {
			_, err := parser.Parse(&ParseParams{
				Source: `query Foo($x: Complex = { a: { b: [ $var ] } }) { field }`,
			})
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:37) Unexpected $\n\n1|query Foo($x: Complex = { a: { b: [ $var ] } }) { field }\n                                      ^")
		})

		Convey("does not accept fragments named \"on\"", func() {
			_, err := parser.Parse(&ParseParams{
				Source: `fragment on on on { on }`,
			})
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:10) Fragment cannot be named \"on\"\n\n1|fragment on on on { on }\n           ^^")

		})

		Convey("does not accept fragments spread of \"on\"", func() {
			_, err := parser.Parse(&ParseParams{
				Source: `{ ...on }`,
			})
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:9) Expected Name, found }\n\n1|{ ...on }\n          ^")
		})

		Convey("does not allow null as value", func() {
			_, err := parser.Parse(&ParseParams{
				Source: `{ fieldWithNullableStringInput(input: null) }`,
			})
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:39) Unexpected Name \"null\"\n\n1|{ fieldWithNullableStringInput(input: null) }\n                                        ^^^^")
		})

		// @TODO
		Convey("parses multi-byte characters", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `
                    # This comment has a \u0A0A multi-byte character.
                    { field(arg: "Has a \u0A0A multi-byte character.") }
                `,
				NoSource: true,
			})
			So(err, ShouldEqual, nil)
			So(result.Definitions[0].(*OperationDefinition).SelectionSet.Selections[0].(*Field).ArgumentIndex["arg"].Value.(*String).Value, ShouldEqual, "Has a \u0A0A multi-byte character.")
		})

		Convey("parses kitchen sick", func() {
			_, err := parser.Parse(&ParseParams{
				Source: KITCHEN_SINK,
			})
			So(err, ShouldEqual, nil)
		})

		Convey("allows non-keywords anywhere a Name is allowed", func() {
			nonKeywords := []string{
				"on",
				"fragment",
				"query",
				"mutation",
				"subscription",
				"true",
				"false",
			}

			for _, keyword := range nonKeywords {
				fragmentName := keyword
				if keyword == "on" {
					fragmentName = "a"
				}
				_, err := parser.Parse(&ParseParams{
					Source: fmt.Sprintf(`query %s {
                        ... %s
                        ... on %s { field }
                    }
                    fragment %s on Type {
                        %s(%s:$%s) @%s(%s: %s)
                    }`, keyword, fragmentName, keyword, fragmentName, keyword, keyword, keyword, keyword, keyword, keyword),
				})
				So(err, ShouldEqual, nil)
			}
		})

		Convey("parses anonymous mutation operations", func() {
			_, err := parser.Parse(&ParseParams{
				Source: `
                mutation {
                    mutationField
                }
                `,
			})
			So(err, ShouldEqual, nil)
		})

		Convey("parses anonymous subscription operations", func() {
			_, err := parser.Parse(&ParseParams{
				Source: `
                    subscription {
                        subscriptionField
                    }
                `,
			})
			So(err, ShouldEqual, nil)
		})

		Convey("parses named mutation operations", func() {
			_, err := parser.Parse(&ParseParams{
				Source: `
                mutation Foo{
                    mutationField
                }
                `,
			})
			So(err, ShouldEqual, nil)
		})

		Convey("parses named subscription operations", func() {
			_, err := parser.Parse(&ParseParams{
				Source: `
                subscription Foo{
                    subscriptionField
                }
                `,
			})
			So(err, ShouldEqual, nil)
		})

		Convey("parse creates ast", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `{
  node(id: 4) {
    id,
    name
  }
}
`,
			})
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, &Document{
				LOC: &LOC{
					Start: &Position{
						Index:  0,
						Line:   1,
						Column: 1,
					},
					End: &Position{
						Index:  40,
						Line:   6,
						Column: 2,
					},
					Source: parser.source,
				},
				Definitions: []ASTNode{
					&OperationDefinition{
						LOC: &LOC{
							Start: &Position{
								Index:  0,
								Line:   1,
								Column: 1,
							},
							End: &Position{
								Index:  40,
								Line:   6,
								Column: 2,
							},
							Source: parser.source,
						},
						Operation: "query",
						SelectionSet: &SelectionSet{
							LOC: &LOC{
								Start: &Position{
									Index:  0,
									Line:   1,
									Column: 1,
								},
								End: &Position{
									Index:  40,
									Line:   6,
									Column: 2,
								},
								Source: parser.source,
							},
							Selections: []ASTNode{
								&Field{
									LOC: &LOC{
										Start: &Position{
											Index:  4,
											Line:   2,
											Column: 3,
										},
										End: &Position{
											Index:  38,
											Line:   5,
											Column: 4,
										},
										Source: parser.source,
									},
									Name: &Name{
										LOC: &LOC{
											Start: &Position{
												Index:  4,
												Line:   2,
												Column: 3,
											},
											End: &Position{
												Index:  8,
												Line:   2,
												Column: 7,
											},
											Source: parser.source,
										},
										Value: "node",
									},
									Arguments: []*Argument{
										&Argument{
											LOC: &LOC{
												Start: &Position{
													Index:  9,
													Line:   2,
													Column: 8,
												},
												End: &Position{
													Index:  14,
													Line:   2,
													Column: 13,
												},
												Source: parser.source,
											},
											Name: &Name{
												LOC: &LOC{
													Start: &Position{
														Index:  9,
														Line:   2,
														Column: 8,
													},
													End: &Position{
														Index:  11,
														Line:   2,
														Column: 10,
													},
													Source: parser.source,
												},
												Value: "id",
											},
											Value: &Int{
												LOC: &LOC{
													Start: &Position{
														Index:  13,
														Line:   2,
														Column: 12,
													},
													End: &Position{
														Index:  14,
														Line:   2,
														Column: 13,
													},
													Source: parser.source,
												},
												Value: 4,
											},
										},
									},
									ArgumentIndex: map[string]*Argument{
										"id": &Argument{
											LOC: &LOC{
												Start: &Position{
													Index:  9,
													Line:   2,
													Column: 8,
												},
												End: &Position{
													Index:  14,
													Line:   2,
													Column: 13,
												},
												Source: parser.source,
											},
											Name: &Name{
												LOC: &LOC{
													Start: &Position{
														Index:  9,
														Line:   2,
														Column: 8,
													},
													End: &Position{
														Index:  11,
														Line:   2,
														Column: 10,
													},
													Source: parser.source,
												},
												Value: "id",
											},
											Value: &Int{
												LOC: &LOC{
													Start: &Position{
														Index:  13,
														Line:   2,
														Column: 12,
													},
													End: &Position{
														Index:  14,
														Line:   2,
														Column: 13,
													},
													Source: parser.source,
												},
												Value: 4,
											},
										},
									},
									SelectionSet: &SelectionSet{
										LOC: &LOC{
											Start: &Position{
												Index:  16,
												Line:   2,
												Column: 15,
											},
											End: &Position{
												Index:  38,
												Line:   5,
												Column: 4,
											},
											Source: parser.source,
										},
										Selections: []ASTNode{
											&Field{
												LOC: &LOC{
													Start: &Position{
														Index:  22,
														Line:   3,
														Column: 5,
													},
													End: &Position{
														Index:  24,
														Line:   3,
														Column: 7,
													},
													Source: parser.source,
												},
												Name: &Name{
													LOC: &LOC{
														Start: &Position{
															Index:  22,
															Line:   3,
															Column: 5,
														},
														End: &Position{
															Index:  24,
															Line:   3,
															Column: 7,
														},
														Source: parser.source,
													},
													Value: "id",
												},
											},
											&Field{
												LOC: &LOC{
													Start: &Position{
														Index:  30,
														Line:   4,
														Column: 5,
													},
													End: &Position{
														Index:  34,
														Line:   4,
														Column: 9,
													},
													Source: parser.source,
												},
												Name: &Name{
													LOC: &LOC{
														Start: &Position{
															Index:  30,
															Line:   4,
															Column: 5,
														},
														End: &Position{
															Index:  34,
															Line:   4,
															Column: 9,
														},
														Source: parser.source,
													},
													Value: "name",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			})
		})
	})

	Convey("Schema Parser", t, func() {

		var result *Document
		var err error
		parser := &Parser{}

		Convey("simple type", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `
type Hello {
  world: String
}`,
			})
			So(err, ShouldEqual, nil)
			loc := createLOCFn(parser.source)

			helloTypeDef := &ObjectTypeDefinition{
				Name: nameNode("Hello", loc(6, 11)),
				Fields: []*FieldDefinition{
					fieldNode(
						nameNode("world", loc(16, 21)),
						typeNode("String", loc(23, 29)),
						loc(16, 29),
					),
				},
				FieldIndex: map[string]*FieldDefinition{
					"world": fieldNode(
						nameNode("world", loc(16, 21)),
						typeNode("String", loc(23, 29)),
						loc(16, 29),
					),
				},
				LOC: loc(1, 31),
			}

			So(result, ShouldResemble, &Document{
				Definitions: []ASTNode{
					helloTypeDef,
				},
				ObjectTypeIndex: map[string]*ObjectTypeDefinition{
					"Hello": helloTypeDef,
				},
				TypeIndex: map[string]ASTNode{
					"Hello": helloTypeDef,
				},
				LOC: loc(1, 31),
			})
		})

		Convey("simple extension", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `
extend type Hello {
  world: String
}`,
			})
			So(err, ShouldEqual, nil)
			loc := createLOCFn(parser.source)

			helloTypeExt := &TypeExtensionDefinition{
				Definition: &ObjectTypeDefinition{
					Name: nameNode("Hello", loc(13, 18)),
					Fields: []*FieldDefinition{
						fieldNode(
							nameNode("world", loc(23, 28)),
							typeNode("String", loc(30, 36)),
							loc(23, 36),
						),
					},
					FieldIndex: map[string]*FieldDefinition{
						"world": fieldNode(
							nameNode("world", loc(23, 28)),
							typeNode("String", loc(30, 36)),
							loc(23, 36),
						),
					},
					LOC: loc(8, 38),
				},
				LOC: loc(1, 38),
			}

			So(result, ShouldResemble, &Document{
				Definitions: []ASTNode{
					helloTypeExt,
				},
				TypeExtensionIndex: map[string]*TypeExtensionDefinition{
					"Hello": helloTypeExt,
				},
				LOC: loc(1, 38),
			})
		})

		Convey("simple non-null type", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `
type Hello {
  world: String!
}`,
			})
			So(err, ShouldEqual, nil)
			loc := createLOCFn(parser.source)

			helloTypeDef := &ObjectTypeDefinition{
				Name: nameNode("Hello", loc(6, 11)),
				Fields: []*FieldDefinition{
					fieldNode(
						nameNode("world", loc(16, 21)),
						&NonNullType{
							Type: typeNode("String", loc(23, 29)),
							LOC:  loc(23, 30),
						},
						loc(16, 30),
					),
				},
				FieldIndex: map[string]*FieldDefinition{
					"world": fieldNode(
						nameNode("world", loc(16, 21)),
						&NonNullType{
							Type: typeNode("String", loc(23, 29)),
							LOC:  loc(23, 30),
						},
						loc(16, 30),
					),
				},
				LOC: loc(1, 32),
			}

			So(result, ShouldResemble, &Document{
				Definitions: []ASTNode{
					helloTypeDef,
				},
				ObjectTypeIndex: map[string]*ObjectTypeDefinition{
					"Hello": helloTypeDef,
				},
				TypeIndex: map[string]ASTNode{
					"Hello": helloTypeDef,
				},
				LOC: loc(1, 32),
			})
		})

		Convey("simple type inheriting multiple interface", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `type Hello implements Wo, rld { }`,
			})
			So(err, ShouldEqual, nil)
			loc := createLOCFn(parser.source)

			helloTypeDef := &ObjectTypeDefinition{
				Name: nameNode("Hello", loc(5, 10)),
				Interfaces: []*NamedType{
					typeNode("Wo", loc(22, 24)),
					typeNode("rld", loc(26, 29)),
				},
				LOC: loc(0, 33),
			}
			So(result, ShouldResemble, &Document{
				Definitions: []ASTNode{
					helloTypeDef,
				},
				ObjectTypeIndex: map[string]*ObjectTypeDefinition{
					"Hello": helloTypeDef,
				},
				TypeIndex: map[string]ASTNode{
					"Hello": helloTypeDef,
				},
				PossibleTypesIndex: map[string][]*ObjectTypeDefinition{
					"Wo":  []*ObjectTypeDefinition{helloTypeDef},
					"rld": []*ObjectTypeDefinition{helloTypeDef},
				},
				LOC: loc(0, 33),
			})
		})

		Convey("single value enum", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `enum Hello { WORLD }`,
			})
			So(err, ShouldEqual, nil)
			loc := createLOCFn(parser.source)

			helloTypeDef := &EnumTypeDefinition{
				Name: nameNode("Hello", loc(5, 10)),
				Values: []*EnumValueDefinition{
					enumValueNode("WORLD", loc(13, 18)),
				},
				LOC: loc(0, 20),
			}
			So(result, ShouldResemble, &Document{
				Definitions: []ASTNode{
					helloTypeDef,
				},
				EnumTypeIndex: map[string]*EnumTypeDefinition{
					"Hello": helloTypeDef,
				},
				TypeIndex: map[string]ASTNode{
					"Hello": helloTypeDef,
				},
				LOC: loc(0, 20),
			})
		})

		Convey("invalid enum value", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `enum Hello { null }`,
			})
			So(result, ShouldEqual, nil)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:14) Enum value cannot be \"null\"\n\n1|enum Hello { null }\n               ^^^^")

			result, err = parser.Parse(&ParseParams{
				Source: `enum Hello { WORLD true }`,
			})
			So(result, ShouldEqual, nil)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:20) Enum value cannot be \"true\"\n\n1|enum Hello { WORLD true }\n                     ^^^^")

			result, err = parser.Parse(&ParseParams{
				Source: `enum Hello { false WORLD }`,
			})
			So(result, ShouldEqual, nil)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:14) Enum value cannot be \"false\"\n\n1|enum Hello { false WORLD }\n               ^^^^^")
		})

		Convey("double value enum", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `enum Hello { WO, RLD }`,
			})
			So(err, ShouldEqual, nil)
			loc := createLOCFn(parser.source)

			helloTypeDef := &EnumTypeDefinition{
				Name: nameNode("Hello", loc(5, 10)),
				Values: []*EnumValueDefinition{
					enumValueNode("WO", loc(13, 15)),
					enumValueNode("RLD", loc(17, 20)),
				},
				LOC: loc(0, 22),
			}
			So(result, ShouldResemble, &Document{
				Definitions: []ASTNode{
					helloTypeDef,
				},
				EnumTypeIndex: map[string]*EnumTypeDefinition{
					"Hello": helloTypeDef,
				},
				TypeIndex: map[string]ASTNode{
					"Hello": helloTypeDef,
				},
				LOC: loc(0, 22),
			})
		})

		Convey("simple interface", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `
interface Hello {
  world: String
}`,
			})
			So(err, ShouldEqual, nil)
			loc := createLOCFn(parser.source)

			helloTypeDef := &InterfaceTypeDefinition{
				Name: nameNode("Hello", loc(11, 16)),
				Fields: []*FieldDefinition{
					&FieldDefinition{
						Name: nameNode("world", loc(21, 26)),
						Type: typeNode("String", loc(28, 34)),
						LOC:  loc(21, 34),
					},
				},
				LOC: loc(1, 36),
			}
			So(result, ShouldResemble, &Document{
				Definitions: []ASTNode{
					helloTypeDef,
				},
				InterfaceTypeIndex: map[string]*InterfaceTypeDefinition{
					"Hello": helloTypeDef,
				},
				TypeIndex: map[string]ASTNode{
					"Hello": helloTypeDef,
				},
				LOC: loc(1, 36),
			})
		})

		Convey("simple field with arg", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `
type Hello {
  world(flag: Boolean): String
}`,
			})
			So(err, ShouldEqual, nil)
			loc := createLOCFn(parser.source)

			helloTypeDef := &ObjectTypeDefinition{
				Name: nameNode("Hello", loc(6, 11)),
				Fields: []*FieldDefinition{
					fieldNodeWithArgs(
						nameNode("world", loc(16, 21)),
						typeNode("String", loc(38, 44)),
						[]*InputValueDefinition{
							inputValueNode(
								nameNode("flag", loc(22, 26)),
								typeNode("Boolean", loc(28, 35)),
								nil,
								loc(22, 35),
							),
						},
						loc(16, 44),
					),
				},
				FieldIndex: map[string]*FieldDefinition{
					"world": fieldNodeWithArgs(
						nameNode("world", loc(16, 21)),
						typeNode("String", loc(38, 44)),
						[]*InputValueDefinition{
							inputValueNode(
								nameNode("flag", loc(22, 26)),
								typeNode("Boolean", loc(28, 35)),
								nil,
								loc(22, 35),
							),
						},
						loc(16, 44),
					),
				},
				LOC: loc(1, 46),
			}
			So(result, ShouldResemble, &Document{
				Definitions: []ASTNode{
					helloTypeDef,
				},
				ObjectTypeIndex: map[string]*ObjectTypeDefinition{
					"Hello": helloTypeDef,
				},
				TypeIndex: map[string]ASTNode{
					"Hello": helloTypeDef,
				},
				LOC: loc(1, 46),
			})
		})

		Convey("simple field with arg with default value", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `
type Hello {
  world(flag: Boolean = true): String
}`,
			})
			So(err, ShouldEqual, nil)
			loc := createLOCFn(parser.source)

			helloTypeDef := &ObjectTypeDefinition{
				Name: nameNode("Hello", loc(6, 11)),
				Fields: []*FieldDefinition{
					fieldNodeWithArgs(
						nameNode("world", loc(16, 21)),
						typeNode("String", loc(45, 51)),
						[]*InputValueDefinition{
							inputValueNode(
								nameNode("flag", loc(22, 26)),
								typeNode("Boolean", loc(28, 35)),
								&Boolean{
									Value: true,
									LOC:   loc(38, 42),
								},
								loc(22, 42),
							),
						},
						loc(16, 51),
					),
				},
				FieldIndex: map[string]*FieldDefinition{
					"world": fieldNodeWithArgs(
						nameNode("world", loc(16, 21)),
						typeNode("String", loc(45, 51)),
						[]*InputValueDefinition{
							inputValueNode(
								nameNode("flag", loc(22, 26)),
								typeNode("Boolean", loc(28, 35)),
								&Boolean{
									Value: true,
									LOC:   loc(38, 42),
								},
								loc(22, 42),
							),
						},
						loc(16, 51),
					),
				},
				LOC: loc(1, 53),
			}
			So(result, ShouldResemble, &Document{
				Definitions: []ASTNode{
					helloTypeDef,
				},
				ObjectTypeIndex: map[string]*ObjectTypeDefinition{
					"Hello": helloTypeDef,
				},
				TypeIndex: map[string]ASTNode{
					"Hello": helloTypeDef,
				},
				LOC: loc(1, 53),
			})
		})

		Convey("simple field with list arg", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `
type Hello {
  world(things: [String]): String
}`,
			})
			So(err, ShouldEqual, nil)
			loc := createLOCFn(parser.source)

			helloTypeDef := &ObjectTypeDefinition{
				Name: nameNode("Hello", loc(6, 11)),
				Fields: []*FieldDefinition{
					fieldNodeWithArgs(
						nameNode("world", loc(16, 21)),
						typeNode("String", loc(41, 47)),
						[]*InputValueDefinition{
							inputValueNode(
								nameNode("things", loc(22, 28)),
								&ListType{
									Type: typeNode("String", loc(31, 37)),
									LOC:  loc(30, 38),
								},
								nil,
								loc(22, 38),
							),
						},
						loc(16, 47),
					),
				},
				FieldIndex: map[string]*FieldDefinition{
					"world": fieldNodeWithArgs(
						nameNode("world", loc(16, 21)),
						typeNode("String", loc(41, 47)),
						[]*InputValueDefinition{
							inputValueNode(
								nameNode("things", loc(22, 28)),
								&ListType{
									Type: typeNode("String", loc(31, 37)),
									LOC:  loc(30, 38),
								},
								nil,
								loc(22, 38),
							),
						},
						loc(16, 47),
					),
				},
				LOC: loc(1, 49),
			}
			So(result, ShouldResemble, &Document{
				Definitions: []ASTNode{
					helloTypeDef,
				},
				ObjectTypeIndex: map[string]*ObjectTypeDefinition{
					"Hello": helloTypeDef,
				},
				TypeIndex: map[string]ASTNode{
					"Hello": helloTypeDef,
				},
				LOC: loc(1, 49),
			})
		})

		Convey("simple field with two arg", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `
type Hello {
  world(argOne: Boolean, argTwo: Int): String
}`,
			})
			So(err, ShouldEqual, nil)
			loc := createLOCFn(parser.source)

			helloTypeDef := &ObjectTypeDefinition{
				Name: nameNode("Hello", loc(6, 11)),
				Fields: []*FieldDefinition{
					fieldNodeWithArgs(
						nameNode("world", loc(16, 21)),
						typeNode("String", loc(53, 59)),
						[]*InputValueDefinition{
							inputValueNode(
								nameNode("argOne", loc(22, 28)),
								typeNode("Boolean", loc(30, 37)),
								nil,
								loc(22, 37),
							),
							inputValueNode(
								nameNode("argTwo", loc(39, 45)),
								typeNode("Int", loc(47, 50)),
								nil,
								loc(39, 50),
							),
						},
						loc(16, 59),
					),
				},
				FieldIndex: map[string]*FieldDefinition{
					"world": fieldNodeWithArgs(
						nameNode("world", loc(16, 21)),
						typeNode("String", loc(53, 59)),
						[]*InputValueDefinition{
							inputValueNode(
								nameNode("argOne", loc(22, 28)),
								typeNode("Boolean", loc(30, 37)),
								nil,
								loc(22, 37),
							),
							inputValueNode(
								nameNode("argTwo", loc(39, 45)),
								typeNode("Int", loc(47, 50)),
								nil,
								loc(39, 50),
							),
						},
						loc(16, 59),
					),
				},
				LOC: loc(1, 61),
			}
			So(result, ShouldResemble, &Document{
				Definitions: []ASTNode{
					helloTypeDef,
				},
				ObjectTypeIndex: map[string]*ObjectTypeDefinition{
					"Hello": helloTypeDef,
				},
				TypeIndex: map[string]ASTNode{
					"Hello": helloTypeDef,
				},
				LOC: loc(1, 61),
			})
		})

		Convey("Union with two types", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `union Hello = Wo | Rld`,
			})
			So(err, ShouldEqual, nil)
			loc := createLOCFn(parser.source)

			helloTypeDef := &UnionTypeDefinition{
				Name: nameNode("Hello", loc(6, 11)),
				Types: []*NamedType{
					typeNode("Wo", loc(14, 16)),
					typeNode("Rld", loc(19, 22)),
				},
				LOC: loc(0, 22),
			}
			So(result, ShouldResemble, &Document{
				Definitions: []ASTNode{
					helloTypeDef,
				},
				UnionTypeIndex: map[string]*UnionTypeDefinition{
					"Hello": helloTypeDef,
				},
				PossibleTypesIndex: map[string][]*ObjectTypeDefinition{
					"Hello": []*ObjectTypeDefinition{nil, nil}, // These are nil because the object types are not defined in the document
				},
				TypeIndex: map[string]ASTNode{
					"Hello": helloTypeDef,
				},
				LOC: loc(0, 22),
			})
		})

		Convey("Scalar", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `scalar Hello`,
			})
			So(err, ShouldEqual, nil)
			loc := createLOCFn(parser.source)

			helloTypeDef := &ScalarTypeDefinition{
				Name: nameNode("Hello", loc(7, 12)),
				LOC:  loc(0, 12),
			}
			So(result, ShouldResemble, &Document{
				Definitions: []ASTNode{
					helloTypeDef,
				},
				ScalarTypeIndex: map[string]*ScalarTypeDefinition{
					"Hello": helloTypeDef,
				},
				TypeIndex: map[string]ASTNode{
					"Hello": helloTypeDef,
				},
				LOC: loc(0, 12),
			})
		})

		Convey("simple input object", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `
input Hello {
  world: String
}`,
			})
			So(err, ShouldEqual, nil)
			loc := createLOCFn(parser.source)

			worldField := inputValueNode(
				nameNode("world", loc(17, 22)),
				typeNode("String", loc(24, 30)),
				nil,
				loc(17, 30),
			)

			helloTypeDef := &InputObjectTypeDefinition{
				Name: nameNode("Hello", loc(7, 12)),
				Fields: []*InputValueDefinition{
					worldField,
				},
				FieldIndex: map[string]*InputValueDefinition{
					"world": worldField,
				},
				LOC: loc(1, 32),
			}
			So(result, ShouldResemble, &Document{
				Definitions: []ASTNode{
					helloTypeDef,
				},
				InputObjectTypeIndex: map[string]*InputObjectTypeDefinition{
					"Hello": helloTypeDef,
				},
				TypeIndex: map[string]ASTNode{
					"Hello": helloTypeDef,
				},
				LOC: loc(1, 32),
			})
		})

		Convey("simple input object with args should fail", func() {
			result, err = parser.Parse(&ParseParams{
				Source: `
input Hello {
  world(foo: Int): String
}`,
			})
			So(err, ShouldNotEqual, nil)
		})
	})

}
