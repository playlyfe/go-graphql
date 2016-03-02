package language

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type Author struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	IsPublished   string        `json:"isPublished"`
	Author        *Author       `json:"author"`
	Title         string        `json:"title"`
	Body          string        `json:"body"`
	keywords      []interface{} `json:"keywords"`
	RecentArticle *Article      `json:"recentArticle"`
}

type Image struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func (author *Author) Pic(width int, height int) *Image {
	return &Image{
		URL:    fmt.Sprintf("cdn://%d", author.ID),
		Width:  width,
		Height: height,
	}
}

type Article struct {
	ID          string        `json:"id"`
	IsPublished string        `json:"isPublished"`
	Author      *Author       `json:"author"`
	Title       string        `json:"title"`
	Body        string        `json:"body"`
	Hidden      string        `json:"hidden"`
	Keywords    []interface{} `json:"keywords"`
}

func TestExecutor(t *testing.T) {
	Convey("Execute: Handles execution of abstract types", t, func() {

		Convey("ResolveType used to resolve runtime type for Interface", func() {
			schema := `
            interface Pet {
                name: String
            }

            type Dog implements Pet {
                name: String
                woofs: Boolean
            }

            type Cat implements Pet {
                name: String
                meows: Boolean
            }

            type QueryRoot {
                pets: [Pet]
            }
            `

			input := `{
                pets {
                    name
                    ... on Dog {
                        woofs
                    }
                    ... on Cat {
                        meows
                    }
                }
            }`

			resolvers := map[string]interface{}{}
			resolvers["QueryRoot/pets"] = func(params *ResolveParams) (interface{}, error) {
				return []map[string]interface{}{
					{
						"__typename": "Dog",
						"name":       "Odie",
						"woofs":      true,
					},
					{
						"__typename": "Cat",
						"name":       "Garfield",
						"meows":      false,
					},
				}, nil
			}
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "QueryRoot", "", resolvers)
			So(err, ShouldEqual, nil)
			executor.ResolveType = func(value interface{}) string {
				if object, ok := value.(map[string]interface{}); ok {
					return object["__typename"].(string)
				}
				return ""
			}
			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"pets": []interface{}{
						map[string]interface{}{
							"name":  "Odie",
							"woofs": true,
						},
						map[string]interface{}{
							"meows": false,
							"name":  "Garfield",
						},
					},
				},
			})
		})

		Convey("ResolveType used to resolve runtime type for Union", func() {
			schema := `
            union Pet = Dog | Cat

            type Dog {
                name: String
                woofs: Boolean
            }

            type Cat {
                name: String
                meows: Boolean
            }

            type QueryRoot {
                pets: [Pet]
            }
            `

			input := `{
                pets {
                    ... on Dog {
                        name
                        woofs
                    }
                    ... on Cat {
                        name
                        meows
                    }
                }
            }`

			resolvers := map[string]interface{}{}
			resolvers["QueryRoot/pets"] = func(params *ResolveParams) (interface{}, error) {
				return []map[string]interface{}{
					{
						"__typename": "Dog",
						"name":       "Odie",
						"woofs":      true,
					},
					{
						"__typename": "Cat",
						"name":       "Garfield",
						"meows":      false,
					},
				}, nil
			}
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "QueryRoot", "", resolvers)
			So(err, ShouldEqual, nil)
			executor.ResolveType = func(value interface{}) string {
				if object, ok := value.(map[string]interface{}); ok {
					return object["__typename"].(string)
				}
				return ""
			}
			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"pets": []interface{}{
						map[string]interface{}{
							"name":  "Odie",
							"woofs": true,
						},
						map[string]interface{}{
							"meows": false,
							"name":  "Garfield",
						},
					},
				},
			})
		})

		Convey("ResolveType on Interface yields useful error", func() {
			schema := `
            interface Pet {
                name: String
            }

            type Human {
                name: String
            }

            type Dog implements Pet {
                name: String
                woofs: Boolean
            }

            type Cat implements Pet {
                name: String
                meows: Boolean
            }

            type QueryRoot {
                pets: [Pet]
            }
            `

			input := `{
                pets {
                    ... on Dog {
                        name
                        woofs
                    }
                    ... on Cat {
                        name
                        meows
                    }
                }
            }`

			resolvers := map[string]interface{}{}
			resolvers["QueryRoot/pets"] = func(params *ResolveParams) (interface{}, error) {
				return []map[string]interface{}{
					{
						"__typename": "Dog",
						"name":       "Odie",
						"woofs":      true,
					},
					{
						"__typename": "Cat",
						"name":       "Garfield",
						"meows":      false,
					},
					{
						"__typename": "Human",
						"name":       "Jon",
					},
				}, nil
			}
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "QueryRoot", "", resolvers)
			So(err, ShouldEqual, nil)
			executor.ResolveType = func(value interface{}) string {
				if object, ok := value.(map[string]interface{}); ok {
					return object["__typename"].(string)
				}
				return ""
			}
			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"pets": []interface{}{
						map[string]interface{}{
							"name":  "Odie",
							"woofs": true,
						},
						map[string]interface{}{
							"meows": false,
							"name":  "Garfield",
						},
						nil,
					},
				},
				"errors": []map[string]interface{}{
					{
						"message": "GraphQL Runtime Error (2:17) Runtime object type \"Human\" is not a possible type for interface type \"Pet\"\n\n1|{\n2|                pets {\n                  ^^^^\n3|                    ... on Dog {\n4|                        name",
						"locations": []map[string]interface{}{
							{
								"column": 17,
								"line":   2,
							},
						},
					},
				},
			})
		})

	})

	Convey("Execute: Handles directives", t, func() {
		schema := `
        type TestType {
            a: String
            b: String
        }
        `
		resolvers := map[string]interface{}{}
		resolvers["TestType/a"] = func(params *ResolveParams) (interface{}, error) {
			return "a", nil
		}
		resolvers["TestType/b"] = func(params *ResolveParams) (interface{}, error) {
			return "b", nil
		}
		context := map[string]interface{}{}
		variables := map[string]interface{}{}
		executor, err := NewExecutor(schema, "TestType", "", resolvers)
		So(err, ShouldEqual, nil)

		Convey("works without directives", func() {

			Convey("basic query works", func() {
				input := `{ a, b }`
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
						"b": "b",
					},
				})
			})
		})

		Convey("works on scalars", func() {

			Convey("if true includes scalar", func() {
				input := `{ a, b @include(if: true) }`
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
						"b": "b",
					},
				})
			})
			Convey("if false omits on scalar", func() {
				input := `{ a, b @include(if: false) }`
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
					},
				})
			})
			Convey("unless false includes scalar", func() {
				input := `{ a, b @skip(if: false) }`
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
						"b": "b",
					},
				})
			})
			Convey("unless true omits scalar", func() {
				input := `{ a, b @skip(if: true) }`
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
					},
				})
			})

		})

		Convey("works on fragment spreads", func() {
			Convey("if false omits fragment spread", func() {
				input := `
                query Q {
                    a
                    ...Frag @include(if: false)
                }
                fragment Frag on TestType {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
					},
				})
			})
			Convey("if true includes fragment spread", func() {
				input := `
                query Q {
                    a
                    ...Frag @include(if: true)
                }
                fragment Frag on TestType {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
						"b": "b",
					},
				})
			})
			Convey("unless false includes fragment spread", func() {
				input := `
                query Q {
                    a
                    ...Frag @skip(if: false)
                }
                fragment Frag on TestType {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
						"b": "b",
					},
				})
			})
			Convey("unless true omits fragment spread", func() {
				input := `
                query Q {
                    a
                    ...Frag @skip(if: true)
                }
                fragment Frag on TestType {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
					},
				})
			})
		})

		Convey("works on inline fragment", func() {
			Convey("if false omits inline fragment", func() {
				input := `
                query Q {
                    a
                    ... on TestType @include(if: false) {
                        b
                    }
                }
                fragment Frag on TestType {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
					},
				})
			})
			Convey("if true includes inline fragment", func() {
				input := `
                query Q {
                    a
                    ... on TestType @include(if: true) {
                        b
                    }
                }
                fragment Frag on TestType {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
						"b": "b",
					},
				})
			})
			Convey("unless false includes inline fragment", func() {
				input := `
                query Q {
                    a
                    ... on TestType @skip(if: false) {
                        b
                    }
                }
                fragment Frag on TestType {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
						"b": "b",
					},
				})
			})
			Convey("unless true includes inline fragment", func() {
				input := `
                query Q {
                    a
                    ... on TestType @skip(if: true) {
                        b
                    }
                }
                fragment Frag on TestType {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
					},
				})
			})
		})

		Convey("works on anonymous inline fragment", func() {

			Convey("if false omits anonymous inline fragment", func() {
				input := `
                query Q {
                    a
                    ... @include(if: false) {
                        b
                    }
                }
                fragment Frag on TestType {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
					},
				})
			})

			Convey("if true includes anonymous inline fragment", func() {
				input := `
                query Q {
                    a
                    ... @include(if: true) {
                        b
                    }
                }
                fragment Frag on TestType {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
						"b": "b",
					},
				})
			})

			Convey("unless false includes anonymous inline fragment", func() {
				input := `
                query Q {
                    a
                    ... @skip(if: false) {
                        b
                    }
                }
                fragment Frag on TestType {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
						"b": "b",
					},
				})
			})

			Convey("unless true includes anonymous inline fragment", func() {
				input := `
                query Q {
                    a
                    ... @skip(if: true) {
                        b
                    }
                }
                fragment Frag on TestType {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
					},
				})
			})

		})

		Convey("works on fragment", func() {

			Convey("if false omits fragment", func() {
				input := `
                query Q {
                    a
                    ...Frag
                }
                fragment Frag on TestType @include(if: false) {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
					},
				})
			})

			Convey("if true includes fragment", func() {
				input := `
                query Q {
                    a
                    ...Frag
                }
                fragment Frag on TestType @include(if: true) {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
						"b": "b",
					},
				})
			})

			Convey("unless false includes fragment", func() {
				input := `
                query Q {
                    a
                    ...Frag
                }
                fragment Frag on TestType @skip(if: false) {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
						"b": "b",
					},
				})
			})

			Convey("if true omits includes fragment", func() {
				input := `
                query Q {
                    a
                    ...Frag
                }
                fragment Frag on TestType @skip(if: true) {
                    b
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"a": "a",
					},
				})
			})

		})

	})

	Convey("Execute: Handles basic execution tasks", t, func() {

		Convey("executes arbitary code", func() {
			schema := `
            type DataType {
                a: String
                b: String
                c: String
                d: String
                e: String
                f: String
                pic(size: Int): String
                deep: DeepDataType
                promise: DataType
            }

            type DeepDataType {
                a: String
                b: String
                c: [String]
                deeper: [DataType]
            }
            `
			resolvers := map[string]interface{}{}
			resolvers["DataType/a"] = func(params *ResolveParams) (interface{}, error) {
				return "Apple", nil
			}
			resolvers["DataType/b"] = func(params *ResolveParams) (interface{}, error) {
				return "Banana", nil
			}
			resolvers["DataType/c"] = func(params *ResolveParams) (interface{}, error) {
				return "Cookie", nil
			}
			resolvers["DataType/d"] = func(params *ResolveParams) (interface{}, error) {
				return "Donut", nil
			}
			resolvers["DataType/e"] = func(params *ResolveParams) (interface{}, error) {
				return "Egg", nil
			}
			resolvers["DataType/f"] = &FieldParams{
				Resolve: func(params *ResolveParams) (interface{}, error) {
					return "Fish", nil
				},
			}
			resolvers["DataType/pic"] = func(params *ResolveParams) (interface{}, error) {
				var size int32
				var ok bool
				if size, ok = params.Args["size"].(int32); !ok {
					size = 50
				}
				return fmt.Sprintf("Pic of size: %d", size), nil
			}
			resolvers["DataType/deep"] = func(params *ResolveParams) (interface{}, error) {
				return map[string]interface{}{}, nil
			}
			resolvers["DataType/promise"] = func(params *ResolveParams) (interface{}, error) {
				return map[string]interface{}{}, nil
			}
			resolvers["DeepDataType/a"] = func(params *ResolveParams) (interface{}, error) {
				return "Already Been Done", nil
			}
			resolvers["DeepDataType/b"] = func(params *ResolveParams) (interface{}, error) {
				return "Boring", nil
			}
			resolvers["DeepDataType/c"] = func(params *ResolveParams) (interface{}, error) {
				return []interface{}{"Contrived", nil, "Confusing"}, nil
			}
			resolvers["DeepDataType/deeper"] = func(params *ResolveParams) (interface{}, error) {
				return []interface{}{map[string]interface{}{}, nil, map[string]interface{}{}}, nil
			}

			context := map[string]interface{}{}
			variables := map[string]interface{}{
				"size": 100,
			}
			executor, err := NewExecutor(schema, "DataType", "", resolvers)
			So(err, ShouldEqual, nil)
			input := `
            query Example($size: Int) {
                a,
                b,
                x: c
                ...c
                f
                ...on DataType {
                    pic(size: $size)
                    promise {
                        a
                    }
                }
                deep {
                    a
                    b
                    c
                    deeper {
                        a
                        b
                    }
                }
            }
            fragment c on DataType {
                d
                e
            }
            `
			result, err := executor.Execute(context, input, variables, "Example")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"a":   "Apple",
					"b":   "Banana",
					"x":   "Cookie",
					"d":   "Donut",
					"e":   "Egg",
					"f":   "Fish",
					"pic": "Pic of size: 100",
					"promise": map[string]interface{}{
						"a": "Apple",
					},
					"deep": map[string]interface{}{
						"a": "Already Been Done",
						"b": "Boring",
						"c": []interface{}{
							"Contrived",
							nil,
							"Confusing",
						},
						"deeper": []interface{}{
							map[string]interface{}{
								"a": "Apple",
								"b": "Banana",
							},
							nil,
							map[string]interface{}{
								"a": "Apple",
								"b": "Banana",
							},
						},
					},
				},
			})
		})

		Convey("merges parallel fragments", func() {
			schema := `
            type Type {
                a: String
                b: String
                c: String
                deep: Type
            }
            `
			resolvers := map[string]interface{}{}
			resolvers["Type/a"] = func(params *ResolveParams) (interface{}, error) {
				return "Apple", nil
			}
			resolvers["Type/b"] = func(params *ResolveParams) (interface{}, error) {
				return "Banana", nil
			}
			resolvers["Type/c"] = func(params *ResolveParams) (interface{}, error) {
				return "Cherry", nil
			}
			resolvers["Type/deep"] = func(params *ResolveParams) (interface{}, error) {
				return map[string]interface{}{}, nil
			}

			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "Type", "", resolvers)
			So(err, ShouldEqual, nil)
			input := `
            { a, ...FragOne, ...FragTwo }

            fragment FragOne on Type {
                b
                deep { b, deeper: deep { b } }
            }

            fragment FragTwo on Type {
                c
                deep { c, deeper: deep { c } }
            }
            `
			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"a": "Apple",
					"b": "Banana",
					"c": "Cherry",
					"deep": map[string]interface{}{
						"b": "Banana",
						"c": "Cherry",
						"deeper": map[string]interface{}{
							"b": "Banana",
							"c": "Cherry",
						},
					},
				},
			})
		})

		Convey("thread context correctly", func() {
			schema := `
            type Type {
                a: String
            }
            `
			var resolvedContext interface{}
			resolvers := map[string]interface{}{}
			resolvers["Type/a"] = func(params *ResolveParams) (interface{}, error) {
				resolvedContext = params.Context
				return nil, nil
			}

			context := map[string]interface{}{
				"contextThing": "thing",
			}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "Type", "", resolvers)
			So(err, ShouldEqual, nil)
			input := `query Example { a }`
			_, err = executor.Execute(context, input, variables, "Example")
			So(err, ShouldEqual, nil)
			So(resolvedContext, ShouldResemble, context)
		})

		Convey("correctly threads arguments", func() {
			schema := `
            type Type {
                b(numArg: Int, stringArg: String): String
            }
            `
			var resolvedArgs map[string]interface{}
			resolvers := map[string]interface{}{}
			resolvers["Type/b"] = func(params *ResolveParams) (interface{}, error) {
				resolvedArgs = params.Args
				return nil, nil
			}

			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "Type", "", resolvers)
			So(err, ShouldEqual, nil)
			input := `
            query Example {
                b(numArg: 123, stringArg: "foo")
            }
            `
			_, err = executor.Execute(context, input, variables, "Example")
			So(err, ShouldEqual, nil)
			So(resolvedArgs["numArg"], ShouldEqual, int32(123))
			So(resolvedArgs["stringArg"], ShouldEqual, "foo")
		})

		Convey("nulls out error subtrees", func() {
			schema := `
            type Type {
                sync: String
                syncError: String
            }
            `
			resolvers := map[string]interface{}{}
			resolvers["Type/sync"] = func(params *ResolveParams) (interface{}, error) {
				return "sync", nil
			}
			resolvers["Type/syncError"] = func(params *ResolveParams) (interface{}, error) {
				return nil, errors.New("Error getting syncError")
			}

			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "Type", "", resolvers)
			So(err, ShouldEqual, nil)
			input := `
            {
                sync
                syncError
            }
            `
			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"sync":      "sync",
					"syncError": nil,
				},
				"errors": []map[string]interface{}{
					{
						"message": "Error getting syncError",
						"locations": []map[string]interface{}{
							{
								"column": 17,
								"line":   4,
							},
						},
					},
				},
			})

		})

		Convey("uses the inline operation if no operation is provided", func() {
			schema := `
            type Type {
                a: String
            }
            `
			resolvers := map[string]interface{}{}
			resolvers["Type/a"] = func(params *ResolveParams) (interface{}, error) {
				return "b", nil
			}
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "Type", "", resolvers)
			So(err, ShouldEqual, nil)
			input := `{ a }`
			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"a": "b",
				},
			})
		})

		Convey("uses the only operation if no operation is provided", func() {
			schema := `
            type Type {
                a: String
            }
            `
			resolvers := map[string]interface{}{}
			resolvers["Type/a"] = func(params *ResolveParams) (interface{}, error) {
				return "b", nil
			}
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "Type", "", resolvers)
			So(err, ShouldEqual, nil)
			input := `query Example { a }`
			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"a": "b",
				},
			})
		})

		Convey("throws if no operation is provided with multiple operations", func() {
			schema := `
            type Type {
                a: String
            }
            `
			resolvers := map[string]interface{}{}
			resolvers["Type/a"] = func(params *ResolveParams) (interface{}, error) {
				return "b", nil
			}
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "Type", "", resolvers)
			So(err, ShouldEqual, nil)
			input := `query Example { a } query OtherExample { a }`
			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"a": "b",
				},
				"errors": []map[string]interface{}{
					{
						"message": "GraphQL Runtime Error: Must provide operation name if query contains multiple operations",
					},
				},
			})
		})

		Convey("uses the query schema for queries", func() {
			schema := `
            type Q {
                a: String
            }
            type M {
                c: String
            }
            type S {
                a: String
            }
            `
			resolvers := map[string]interface{}{}
			resolvers["Q/a"] = func(params *ResolveParams) (interface{}, error) {
				return "b", nil
			}
			resolvers["M/c"] = func(params *ResolveParams) (interface{}, error) {
				return "d", nil
			}
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "Q", "M", resolvers)
			So(err, ShouldEqual, nil)
			input := `query Q { a } mutation M { c } subscription S { a }`
			result, err := executor.Execute(context, input, variables, "Q")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"a": "b",
				},
			})
		})

		Convey("uses the mutation schema for mutations", func() {
			schema := `
            type Q {
                a: String
            }
            type M {
                c: String
            }
            `
			resolvers := map[string]interface{}{}
			resolvers["Q/a"] = func(params *ResolveParams) (interface{}, error) {
				return "b", nil
			}
			resolvers["M/c"] = func(params *ResolveParams) (interface{}, error) {
				return "d", nil
			}
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "Q", "M", resolvers)
			So(err, ShouldEqual, nil)
			input := `query Q { a } mutation M { c }`
			result, err := executor.Execute(context, input, variables, "M")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"c": "d",
				},
			})
		})

		Convey("avoids recursion", func() {
			schema := `
            type Type {
                a: String
            }
            `
			resolvers := map[string]interface{}{}
			resolvers["Type/a"] = func(params *ResolveParams) (interface{}, error) {
				return "b", nil
			}
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "Type", "", resolvers)
			So(err, ShouldEqual, nil)
			input := `
            query Q {
                a
                ...Frag
                ...Frag
            }

            fragment Frag on Type {
                a,
                ...Frag
            }
            `
			result, err := executor.Execute(context, input, variables, "Q")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"a": "b",
				},
			})
		})

		Convey("does not include illegal fields in output", func() {
			schema := `
            type Q {
                a: String
            }
            type M {
                c: String
            }
            `
			resolvers := map[string]interface{}{}
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "Q", "M", resolvers)
			So(err, ShouldEqual, nil)
			input := `
            mutation M {
                thisIsIllegalDontIncludeMe
            }
            `
			result, err := executor.Execute(context, input, variables, "M")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{},
			})
		})

		Convey("does not include arguments that were not set", func() {
			schema := `
            type Type {
                field(a: Boolean, b: Boolean, c: Boolean, d: Int, e: Int): String
            }
            `
			resolvers := map[string]interface{}{}
			resolvers["Type/field"] = func(params *ResolveParams) (interface{}, error) {
				result, err := json.Marshal(params.Args)
				if err != nil {
					return nil, err
				}
				return string(result), err
			}
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "Type", "", resolvers)
			So(err, ShouldEqual, nil)
			input := `{ field(a: true, c: false, e: 0) }`
			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"field": `{"a":true,"c":false,"e":0}`,
				},
			})
		})

	})

	Convey("Execute: Handles execution with a complex schema", t, func() {
		Convey("executes using a schema", func() {
			schema := `
            type Image {
                url: String
                width: Int
                height: Int
            }

            type Author {
                id: String
                name: String
                pic(width: Int, height: Int): Image
                recentArticle: Article
            }

            type Article {
                id: String!
                isPublished: Boolean
                author: Author
                title: String
                body: String
                keywords: [String]
            }

            type Query {
                article(id: ID): Article
                feed: [Article]
            }
            `

			var johnSmith *Author
			article := func(id string) *Article {
				return &Article{
					ID:          id,
					IsPublished: "true",
					Author:      johnSmith,
					Title:       "My Article " + id,
					Body:        "This is a post",
					Hidden:      "This data is not exposed in the schema",
					Keywords:    []interface{}{"foo", "bar", 1, true, nil},
				}
			}
			johnSmith = &Author{
				ID:            123,
				Name:          "John Smith",
				RecentArticle: article("1"),
			}

			resolvers := map[string]interface{}{}
			resolvers["Author/pic"] = func(params *ResolveParams) (interface{}, error) {
				if author, ok := params.Source.(*Author); ok {
					return author.Pic(int(params.Args["width"].(int32)), int(params.Args["height"].(int32))), nil
				}
				return nil, nil
			}
			resolvers["Query/article"] = func(params *ResolveParams) (interface{}, error) {
				return article(params.Args["id"].(string)), nil
			}
			resolvers["Query/feed"] = func(params *ResolveParams) (interface{}, error) {
				return []*Article{
					article("1"),
					article("2"),
					article("3"),
					article("4"),
					article("5"),
					article("6"),
					article("7"),
					article("8"),
					article("9"),
					article("10"),
				}, nil
			}
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "Query", "", resolvers)
			So(err, ShouldEqual, nil)
			input := `
            {
                feed {
                    id,
                    title
                },
                article(id: "1") {
                    ...articleFields,
                    author {
                        id,
                        name,
                        pic(width: 640, height: 480) {
                            url,
                            width,
                            height
                        },
                        recentArticle {
                            ...articleFields,
                            keywords
                        }
                    }
                }
            }
            fragment articleFields on Article {
                id,
                isPublished,
                title,
                body,
                hidden,
                notdefined
            }
            `
			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"feed": []interface{}{
						map[string]interface{}{
							"id":    "1",
							"title": "My Article 1",
						},
						map[string]interface{}{
							"id":    "2",
							"title": "My Article 2",
						},
						map[string]interface{}{
							"id":    "3",
							"title": "My Article 3",
						},
						map[string]interface{}{
							"id":    "4",
							"title": "My Article 4",
						},
						map[string]interface{}{
							"id":    "5",
							"title": "My Article 5",
						},
						map[string]interface{}{
							"id":    "6",
							"title": "My Article 6",
						},
						map[string]interface{}{
							"id":    "7",
							"title": "My Article 7",
						},
						map[string]interface{}{
							"id":    "8",
							"title": "My Article 8",
						},
						map[string]interface{}{
							"id":    "9",
							"title": "My Article 9",
						},
						map[string]interface{}{
							"id":    "10",
							"title": "My Article 10",
						},
					},
					"article": map[string]interface{}{
						"id":          "1",
						"isPublished": true,
						"title":       "My Article 1",
						"body":        "This is a post",
						"author": map[string]interface{}{
							"id":   "123",
							"name": "John Smith",
							"pic": map[string]interface{}{
								"url":    "cdn://123",
								"width":  int32(640),
								"height": int32(480),
							},
							"recentArticle": map[string]interface{}{
								"id":          "1",
								"isPublished": true,
								"title":       "My Article 1",
								"body":        "This is a post",
								"keywords":    []interface{}{"foo", "bar", "1", "true", nil},
							},
						},
					},
				},
			})
		})
	})

}
