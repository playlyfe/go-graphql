package language

import (
	//"encoding/json"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
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

}
