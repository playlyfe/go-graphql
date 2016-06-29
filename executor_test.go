package graphql

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/playlyfe/go-graphql/language"
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

type NumberHolder struct {
	TheNumber int32 `json:"theNumber"`
}

type Cat struct {
	Name  string `json:"name"`
	Meows bool   `json:"meows"`
}

type Dog struct {
	Name  string `json:"name"`
	Barks bool   `json:"barks"`
}

type Person struct {
	Name    string        `json:"name"`
	Pets    []interface{} `json:"pets"`
	Friends []interface{} `json:"friends"`
}

type Root struct {
	NumberHolder *NumberHolder
}

func NewRoot(number int32) *Root {
	return &Root{
		NumberHolder: &NumberHolder{
			TheNumber: number,
		},
	}
}

func (r *Root) changeTheNumber(newNumber int32) *NumberHolder {
	r.NumberHolder.TheNumber = newNumber
	return r.NumberHolder
}
func (r *Root) failToChangeTheNumber() error {
	return errors.New("Cannot change the number")
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

	Convey("Execute: Handles list nullability", t, func() {

		check := func(testType string, testData interface{}, expected interface{}) {
			data := map[string]interface{}{
				"test": testData,
			}

			schema := fmt.Sprintf(`
            type DataType {
                test: %s
                nest: DataType
            }
            `, testType)

			resolvers := map[string]interface{}{}
			resolvers["DataType/nest"] = func(params *ResolveParams) (interface{}, error) {
				return data, nil
			}
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, "DataType", "", resolvers)
			So(err, ShouldEqual, nil)
			result, err := executor.Execute(context, `{ nest { test } }`, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, expected)
		}

		Convey("[T]", func() {
			testType := "[Int]"
			Convey("Array<T>", func() {
				Convey("contains values", func() {
					check(testType, []interface{}{1, 2}, map[string]interface{}{
						"data": map[string]interface{}{
							"nest": map[string]interface{}{
								"test": []interface{}{int32(1), int32(2)},
							},
						},
					})
				})
				Convey("contains null", func() {
					check(testType, []interface{}{1, nil, 2}, map[string]interface{}{
						"data": map[string]interface{}{
							"nest": map[string]interface{}{
								"test": []interface{}{int32(1), nil, int32(2)},
							},
						},
					})
				})
				Convey("returns null", func() {
					check(testType, nil, map[string]interface{}{
						"data": map[string]interface{}{
							"nest": map[string]interface{}{
								"test": nil,
							},
						},
					})
				})
			})
		})

		Convey("[T]!", func() {
			testType := "[Int]!"
			Convey("Array<T>", func() {
				Convey("contains values", func() {
					check(testType, []interface{}{1, 2}, map[string]interface{}{
						"data": map[string]interface{}{
							"nest": map[string]interface{}{
								"test": []interface{}{int32(1), int32(2)},
							},
						},
					})
				})
				Convey("contains null", func() {
					check(testType, []interface{}{1, nil, 2}, map[string]interface{}{
						"data": map[string]interface{}{
							"nest": map[string]interface{}{
								"test": []interface{}{int32(1), nil, int32(2)},
							},
						},
					})
				})
				Convey("returns null", func() {
					check(testType, nil, map[string]interface{}{
						"data": map[string]interface{}{
							"nest": nil,
						},
						"errors": []map[string]interface{}{
							{
								"message": "Cannot return null for non-nullable field DataType.test",
								"locations": []map[string]interface{}{
									{
										"column": 10,
										"line":   1,
									},
								},
							},
						},
					})
				})
			})
		})

		Convey("[T!]", func() {
			testType := "[Int!]"
			Convey("Array<T>", func() {
				Convey("contains values", func() {
					check(testType, []interface{}{1, 2}, map[string]interface{}{
						"data": map[string]interface{}{
							"nest": map[string]interface{}{
								"test": []interface{}{int32(1), int32(2)},
							},
						},
					})
				})
				Convey("contains null", func() {
					check(testType, []interface{}{1, nil, 2}, map[string]interface{}{
						"data": map[string]interface{}{
							"nest": map[string]interface{}{
								"test": nil,
							},
						},
						"errors": []map[string]interface{}{
							{
								"message": "Cannot return null for non-nullable field DataType.test",
								"locations": []map[string]interface{}{
									{
										"column": 10,
										"line":   1,
									},
								},
							},
						},
					})
				})
				Convey("returns null", func() {
					check(testType, nil, map[string]interface{}{
						"data": map[string]interface{}{
							"nest": map[string]interface{}{
								"test": nil,
							},
						},
					})
				})
			})
		})

		Convey("[T!]!", func() {
			testType := "[Int!]!"
			Convey("Array<T>", func() {
				Convey("contains values", func() {
					check(testType, []interface{}{1, 2}, map[string]interface{}{
						"data": map[string]interface{}{
							"nest": map[string]interface{}{
								"test": []interface{}{int32(1), int32(2)},
							},
						},
					})
				})
				Convey("contains null", func() {
					check(testType, []interface{}{1, nil, 2}, map[string]interface{}{
						"data": map[string]interface{}{
							"nest": nil,
						},
						"errors": []map[string]interface{}{
							{
								"message": "Cannot return null for non-nullable field DataType.test",
								"locations": []map[string]interface{}{
									{
										"column": 10,
										"line":   1,
									},
								},
							},
						},
					})
				})
				Convey("returns null", func() {
					check(testType, nil, map[string]interface{}{
						"data": map[string]interface{}{
							"nest": nil,
						},
						"errors": []map[string]interface{}{
							{
								"message": "Cannot return null for non-nullable field DataType.test",
								"locations": []map[string]interface{}{
									{
										"column": 10,
										"line":   1,
									},
								},
							},
						},
					})
				})
			})
		})

	})

	Convey("Execute: Handles mutation execution ordering", t, func() {
		schema := `
        type NumberHolder {
            theNumber: Int
        }

        type Query {
            numberHolder: NumberHolder
        }

        type Mutation {
            changeTheNumber(newNumber: Int): NumberHolder
            failToChangeTheNumber(newNumber: Int): NumberHolder
        }
        `
		resolvers := map[string]interface{}{}
		resolvers["Mutation/changeTheNumber"] = func(params *ResolveParams) (interface{}, error) {
			obj := params.Context.(*Root)
			obj.changeTheNumber(params.Args["newNumber"].(int32))
			return obj.NumberHolder, nil
		}
		resolvers["Mutation/failToChangeTheNumber"] = func(params *ResolveParams) (interface{}, error) {
			obj := params.Context.(*Root)
			return nil, obj.failToChangeTheNumber()
		}
		executor, err := NewExecutor(schema, "Query", "Mutation", resolvers)
		So(err, ShouldEqual, nil)

		Convey("evaluates mutations serially", func() {
			input := `
            mutation M {
                first: changeTheNumber(newNumber: 1) {
                    theNumber
                },
                second: changeTheNumber(newNumber: 2) {
                    theNumber
                }
                third: changeTheNumber(newNumber: 3) {
                    theNumber
                }
            }
            `
			context := NewRoot(6)
			variables := map[string]interface{}{}
			result, err := executor.Execute(context, input, variables, "M")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"first": map[string]interface{}{
						"theNumber": int32(1),
					},
					"second": map[string]interface{}{
						"theNumber": int32(2),
					},
					"third": map[string]interface{}{
						"theNumber": int32(3),
					},
				},
			})

		})

		Convey("evaluates mutations correctly in the presence of a failed mutation", func() {
			input := `
            mutation M {
                first: changeTheNumber(newNumber: 1) {
                    theNumber
                },
                second: failToChangeTheNumber(newNumber: 2) {
                    theNumber
                }
                third: changeTheNumber(newNumber: 3) {
                    theNumber
                }
                fourth: failToChangeTheNumber(newNumber: 4) {
                    theNumber
                }
                fifth: changeTheNumber(newNumber: 5) {
                    theNumber
                }
            }
            `
			context := NewRoot(6)
			variables := map[string]interface{}{}
			result, err := executor.Execute(context, input, variables, "M")
			So(err, ShouldEqual, nil)
			So(result["data"], ShouldResemble, map[string]interface{}{
				"first": map[string]interface{}{
					"theNumber": int32(1),
				},
				"second": nil,
				"third": map[string]interface{}{
					"theNumber": int32(3),
				},
				"fourth": nil,
				"fifth": map[string]interface{}{
					"theNumber": int32(5),
				},
			})
			errors := result["errors"].([]map[string]interface{})
			So(len(errors), ShouldEqual, 2)
			So(errors[0]["message"], ShouldEqual, "Cannot change the number")
			So(errors[1]["message"], ShouldEqual, "Cannot change the number")
		})
	})

	// Note: Error handling behaviour differs significantly from the graphql js reference implementation
	Convey("Execute: Handles non-nullable types", t, func() {

		schema := `
        type DataType {
            sync: String
            nonNullSync: String!
            nest: DataType
            nonNullNest: DataType!
        }
        `
		syncError := errors.New("sync")
		nonNullSyncError := errors.New("nonNullSync")
		throwingResolvers := map[string]interface{}{}
		throwingResolvers["DataType/sync"] = func(params *ResolveParams) (interface{}, error) {
			return nil, syncError
		}
		throwingResolvers["DataType/nonNullSync"] = func(params *ResolveParams) (interface{}, error) {
			return nil, nonNullSyncError
		}
		throwingResolvers["DataType/nest"] = func(params *ResolveParams) (interface{}, error) {
			return map[string]interface{}{}, nil
		}
		throwingResolvers["DataType/nonNullNest"] = func(params *ResolveParams) (interface{}, error) {
			return map[string]interface{}{}, nil
		}

		nullingResolvers := map[string]interface{}{}
		nullingResolvers["DataType/sync"] = func(params *ResolveParams) (interface{}, error) {
			return nil, nil
		}
		nullingResolvers["DataType/nonNullSync"] = func(params *ResolveParams) (interface{}, error) {
			return nil, nil
		}
		nullingResolvers["DataType/nest"] = func(params *ResolveParams) (interface{}, error) {
			return map[string]interface{}{}, nil
		}
		nullingResolvers["DataType/nonNullNest"] = func(params *ResolveParams) (interface{}, error) {
			return map[string]interface{}{}, nil
		}

		Convey("nulls a nullable field that throws synchronously", func() {
			input := `
            query Q {
                sync
            }
            `
			executor, err := NewExecutor(schema, "DataType", "", throwingResolvers)
			So(err, ShouldEqual, nil)
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			result, err := executor.Execute(context, input, variables, "Q")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"sync": nil,
				},
				"errors": []map[string]interface{}{
					{
						"message": syncError.Error(),
						"locations": []map[string]interface{}{
							{
								"line":   3,
								"column": 17,
							},
						},
					},
				},
			})
		})

		Convey("nulls a synchronously returned object that contains a non-nullable field that throws synchronously", func() {
			input := `
            query Q {
                nest {
                    nonNullSync
                }
            }
            `
			executor, err := NewExecutor(schema, "DataType", "", throwingResolvers)
			So(err, ShouldEqual, nil)
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			result, err := executor.Execute(context, input, variables, "Q")
			So(err, ShouldEqual, nil)
			So(result["data"], ShouldResemble, map[string]interface{}{
				"nest": nil,
			})
			errors := result["errors"].([]map[string]interface{})
			So(len(errors), ShouldEqual, 2)
			So(errors[0]["message"], ShouldEqual, nonNullSyncError.Error())
			So(errors[1]["message"], ShouldEqual, "Cannot return null for non-nullable field DataType.nonNullSync")
		})

		Convey("nulls a complex tree of nullable fields that throw", func() {
			input := `
            query Q {
                nest {
                    sync
                    nest {
                        sync
                    }
                }
            }
            `
			executor, err := NewExecutor(schema, "DataType", "", throwingResolvers)
			So(err, ShouldEqual, nil)
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			result, err := executor.Execute(context, input, variables, "Q")
			So(err, ShouldEqual, nil)
			So(result["data"], ShouldResemble, map[string]interface{}{
				"nest": map[string]interface{}{
					"sync": nil,
					"nest": map[string]interface{}{
						"sync": nil,
					},
				},
			})
			errors := result["errors"].([]map[string]interface{})
			So(len(errors), ShouldEqual, 2)
			So(errors[0]["message"], ShouldEqual, syncError.Error())
			So(errors[1]["message"], ShouldEqual, syncError.Error())
		})

		Convey("nulls the first nullable object after a field throws in a long chaing of fields that are non-null", func() {
			input := `
            query Q {
                nest {
                    nonNullNest {
                        nonNullNest {
                            nonNullNest {
                                nonNullNest {
                                    nonNullSync
                                }
                            }
                        }
                    }
                }
            }
            `
			executor, err := NewExecutor(schema, "DataType", "", throwingResolvers)
			So(err, ShouldEqual, nil)
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			result, err := executor.Execute(context, input, variables, "Q")
			So(err, ShouldEqual, nil)
			So(result["data"], ShouldResemble, map[string]interface{}{
				"nest": nil,
			})
			errors := result["errors"].([]map[string]interface{})
			So(len(errors), ShouldEqual, 2)
			So(errors[0]["message"], ShouldEqual, nonNullSyncError.Error())
			So(errors[1]["message"], ShouldEqual, "Cannot return null for non-nullable field DataType.nonNullSync")
		})

		Convey("nulls a nullable field that synchronously returns null", func() {
			input := `
            query Q {
                sync
            }
            `
			executor, err := NewExecutor(schema, "DataType", "", nullingResolvers)
			So(err, ShouldEqual, nil)
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			result, err := executor.Execute(context, input, variables, "Q")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"sync": nil,
				},
			})
		})

		Convey("nulls a synchronously returned object that contains a non-nullable field that returns null synchronously", func() {
			input := `
            query Q {
                nest {
                    nonNullSync,
                }
            }
            `
			executor, err := NewExecutor(schema, "DataType", "", nullingResolvers)
			So(err, ShouldEqual, nil)
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			result, err := executor.Execute(context, input, variables, "Q")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"nest": nil,
				},
				"errors": []map[string]interface{}{
					{
						"message": "Cannot return null for non-nullable field DataType.nonNullSync",
						"locations": []map[string]interface{}{
							{
								"column": 21,
								"line":   4,
							},
						},
					},
				},
			})
		})

		Convey("nulls a complex tree of nullable fields that return null", func() {
			input := `
            query Q {
                nest {
                    sync
                    nest {
                        sync
                    }
                }
            }
            `
			executor, err := NewExecutor(schema, "DataType", "", nullingResolvers)
			So(err, ShouldEqual, nil)
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			result, err := executor.Execute(context, input, variables, "Q")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"nest": map[string]interface{}{
						"sync": nil,
						"nest": map[string]interface{}{
							"sync": nil,
						},
					},
				},
			})
		})

		Convey("nulls the first nullable object after a field returns null in a long chaing of fields that are non-null", func() {
			input := `
            query Q {
                nest {
                    nonNullNest {
                        nonNullNest {
                            nonNullNest {
                                nonNullNest {
                                    nonNullSync
                                }
                            }
                        }
                    }
                }
            }
            `
			executor, err := NewExecutor(schema, "DataType", "", throwingResolvers)
			So(err, ShouldEqual, nil)
			context := map[string]interface{}{}
			variables := map[string]interface{}{}
			result, err := executor.Execute(context, input, variables, "Q")
			So(err, ShouldEqual, nil)
			So(result["data"], ShouldResemble, map[string]interface{}{
				"nest": nil,
			})
			errors := result["errors"].([]map[string]interface{})
			So(len(errors), ShouldEqual, 2)
			So(errors[0]["message"], ShouldEqual, nonNullSyncError.Error())
			So(errors[1]["message"], ShouldEqual, "Cannot return null for non-nullable field DataType.nonNullSync")
		})

	})

	Convey("Execute: Union and intersection types", t, func() {
		schema := `
        interface Named {
            name: String
        }

        type Dog implements Named {
            name: String
            barks: Boolean
        }

        type Cat implements Named {
            name: String
            meows: Boolean
        }

        union Pet = Dog | Cat

        type Person implements Named {
            name: String
            pets: [Pet]
            friends: [Named]
        }
        `

		garfield := &Cat{
			Name:  "Garfield",
			Meows: false,
		}
		odie := &Dog{
			Name:  "Odie",
			Barks: true,
		}
		liz := &Person{
			Name: "Liz",
		}
		john := &Person{
			Name:    "John",
			Pets:    []interface{}{garfield, odie},
			Friends: []interface{}{liz, odie},
		}
		resolvers := map[string]interface{}{}
		resolvers["Person/name"] = func(params *ResolveParams) (interface{}, error) {
			if person, ok := params.Source.(*Person); ok {
				return person.Name, nil
			} else {
				return john.Name, nil
			}
		}
		resolvers["Person/pets"] = func(params *ResolveParams) (interface{}, error) {
			if person, ok := params.Source.(*Person); ok {
				return person.Pets, nil
			} else {
				return john.Pets, nil
			}
		}
		resolvers["Person/friends"] = func(params *ResolveParams) (interface{}, error) {
			if person, ok := params.Source.(*Person); ok {
				return person.Friends, nil
			} else {
				return john.Friends, nil
			}
		}
		context := map[string]interface{}{}
		variables := map[string]interface{}{}
		executor, err := NewExecutor(schema, "Person", "", resolvers)
		So(err, ShouldEqual, nil)
		executor.ResolveType = func(value interface{}) string {
			switch value.(type) {
			case *Cat:
				return "Cat"
			case *Dog:
				return "Dog"
			case *Person:
				return "Person"
			}
			return ""
		}

		Convey("can introspect on union and intersection types", func() {
			input := `
            {
                Named: __type(name: "Named") {
                    kind
                    name
                    fields { name }
                    interfaces { name }
                    possibleTypes { name }
                    enumValues { name }
                    inputFields { name }
                }
                Pet: __type(name: "Pet") {
                    kind
                    name
                    fields { name }
                    interfaces { name }
                    possibleTypes { name }
                    enumValues { name }
                    inputFields { name }
                }
            }
            `

			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"Named": map[string]interface{}{
						"kind": "INTERFACE",
						"name": "Named",
						"fields": []interface{}{
							map[string]interface{}{
								"name": "name",
							},
						},
						"interfaces": nil,
						"possibleTypes": []interface{}{
							map[string]interface{}{"name": "Dog"},
							map[string]interface{}{"name": "Cat"},
							map[string]interface{}{"name": "Person"},
						},
						"enumValues":  nil,
						"inputFields": nil,
					},
					"Pet": map[string]interface{}{
						"kind":       "UNION",
						"name":       "Pet",
						"fields":     nil,
						"interfaces": nil,
						"possibleTypes": []interface{}{
							map[string]interface{}{"name": "Dog"},
							map[string]interface{}{"name": "Cat"},
						},
						"enumValues":  nil,
						"inputFields": nil,
					},
				},
			})
		})

		Convey("executes using union types", func() {
			// NOTE: This is an *invalid* query, but it should be an *executable* query.

			input := `
            {
                __typename
                name
                pets {
                    __typename
                    name
                    barks
                    meows
                }
            }
            `

			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"__typename": "Person",
					"name":       "John",
					"pets": []interface{}{
						map[string]interface{}{
							"__typename": "Cat",
							"name":       "Garfield",
							"meows":      false,
						},
						map[string]interface{}{
							"__typename": "Dog",
							"name":       "Odie",
							"barks":      true,
						},
					},
				},
			})
		})

		Convey("executes union types with inline fragments", func() {
			input := `
            {
                __typename
                name
                pets {
                    __typename
                    ... on Dog {
                        name
                        barks
                    }
                    ... on Cat {
                        name
                        meows
                    }
                }
            }
            `

			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"__typename": "Person",
					"name":       "John",
					"pets": []interface{}{
						map[string]interface{}{
							"__typename": "Cat",
							"name":       "Garfield",
							"meows":      false,
						},
						map[string]interface{}{
							"__typename": "Dog",
							"name":       "Odie",
							"barks":      true,
						},
					},
				},
			})

			input = `
            {
                __typename
                name
                friends {
                    __typename
                    name
                    ... on Dog {
                        barks
                    }
                    ... on Cat {
                        meows
                    }
                }
            }
            `

			result, err = executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"__typename": "Person",
					"name":       "John",
					"friends": []interface{}{
						map[string]interface{}{
							"__typename": "Person",
							"name":       "Liz",
						},
						map[string]interface{}{
							"__typename": "Dog",
							"name":       "Odie",
							"barks":      true,
						},
					},
				},
			})
		})

		Convey("executes using interface types", func() {

			// NOTE: This is an *invalid* query, but it should be an *executable* query
			input := `
            {
                __typename
                name
                friends {
                    __typename
                    name
                    barks
                    meows
                }
            }
            `

			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"__typename": "Person",
					"name":       "John",
					"friends": []interface{}{
						map[string]interface{}{
							"__typename": "Person",
							"name":       "Liz",
						},
						map[string]interface{}{
							"__typename": "Dog",
							"name":       "Odie",
							"barks":      true,
						},
					},
				},
			})
		})

		Convey("allows fragment conditions to be abstract types", func() {
			input := `
            {
                __typename
                name
                pets { ...PetFields }
                friends { ...FriendFields }
            }

            fragment PetFields on Pet {
                __typename
                ... on Dog {
                    name
                    barks
                }
                ... on Cat {
                    name
                    meows
                }
            }
            fragment FriendFields on Named {
                __typename
                name
                ... on Dog {
                    barks
                }
                ... on Cat {
                    meows
                }
            }
            `

			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"__typename": "Person",
					"name":       "John",
					"pets": []interface{}{
						map[string]interface{}{
							"__typename": "Cat",
							"name":       "Garfield",
							"meows":      false,
						},
						map[string]interface{}{
							"__typename": "Dog",
							"name":       "Odie",
							"barks":      true,
						},
					},
					"friends": []interface{}{
						map[string]interface{}{
							"__typename": "Person",
							"name":       "Liz",
						},
						map[string]interface{}{
							"__typename": "Dog",
							"name":       "Odie",
							"barks":      true,
						},
					},
				},
			})
		})
	})

	Convey("Execute: Handles inputs", t, func() {

		schema := `
        scalar ComplexScalar

        input TestInputObject {
            a: String
            b: [String]
            c: String!
            d: ComplexScalar
        }

        input TestNestedInputObject {
            na: TestInputObject!
            nb: String!
        }

        type TestType {
            fieldWithObjectInput(input: TestInputObject): String
            fieldWithNullableStringInput(input: String): String
            fieldWithNonNullableStringInput(input: String!): String
            fieldWithDefaultArgumentValue(input: String = "Hello World"): String
            fieldWithNestedObjectInput(input: TestNestedInputObject = "Hello World"): String
            list(input: [String]): String
            nnList(input: [String]!): String
            listNN(input: [String!]): String
            nnlistNN(input: [String!]!): String
            deserializedValue(input: ComplexScalar!): ComplexScalar
        }
        `
		resolvers := map[string]interface{}{}
		fieldResolver := func(params *ResolveParams) (interface{}, error) {
			if input, ok := params.Args["input"]; ok {
				result, err := json.Marshal(input)
				if err != nil {
					return nil, err
				}
				return string(result), nil
			}
			return nil, nil
		}
		resolvers["TestType/fieldWithObjectInput"] = fieldResolver
		resolvers["TestType/fieldWithNullableStringInput"] = fieldResolver
		resolvers["TestType/fieldWithNonNullableStringInput"] = fieldResolver
		resolvers["TestType/fieldWithDefaultArgumentValue"] = fieldResolver
		resolvers["TestType/fieldWithNestedObjectInput"] = fieldResolver
		resolvers["TestType/list"] = fieldResolver
		resolvers["TestType/nnList"] = fieldResolver
		resolvers["TestType/listNN"] = fieldResolver
		resolvers["TestType/nnlistNN"] = fieldResolver
		resolvers["TestType/deserializedValue"] = func(params *ResolveParams) (interface{}, error) {
			return params.Args["input"], nil
		}
		context := map[string]interface{}{}
		variables := map[string]interface{}{}
		executor, err := NewExecutor(schema, "TestType", "", resolvers)
		So(err, ShouldEqual, nil)
		executor.Scalars["ComplexScalar"] = &Scalar{
			ParseValue: func(context interface{}, value interface{}) (interface{}, error) {
				if val, ok := value.(string); ok {
					if val == "SerializedValue" {
						return "DeserializedValue", nil
					}
				}
				return nil, &GraphQLError{
					Message: "Failed to parse ComplexScalar value",
				}
			},
			ParseLiteral: func(context interface{}, value interface{}) (interface{}, error) {
				if ast, ok := value.(*String); ok {
					if ast.Value == "SerializedValue" {
						return "DeserializedValue", nil
					}
				}
				return nil, &GraphQLError{
					Message: "Failed to parse ComplexScalar value",
				}
			},
			Serialize: func(context interface{}, value interface{}) (interface{}, error) {
				if val, ok := value.(string); ok {
					if val == "DeserializedValue" {
						return "SerializedValue", nil
					}
				}
				return nil, &GraphQLError{
					Message: "Failed to serialize ComplexScalar value",
				}
			},
		}

		Convey("Handles objects and nullability", func() {

			Convey("using inline structs", func() {

				Convey("executes with complex input", func() {
					input := `
                    {
                        fieldWithObjectInput(input: {a: "foo", b: ["bar"], c:"baz"})
                    }
                    `
					result, err := executor.Execute(context, input, variables, "")
					So(err, ShouldEqual, nil)
					So(result, ShouldResemble, map[string]interface{}{
						"data": map[string]interface{}{
							"fieldWithObjectInput": `{"a":"foo","b":["bar"],"c":"baz"}`,
						},
					})
				})

				Convey("properly parses single value to list", func() {
					input := `
                    {
                        fieldWithObjectInput(input: {a: "foo", b: "bar", c:"baz"})
                    }
                    `
					result, err := executor.Execute(context, input, variables, "")
					So(err, ShouldEqual, nil)
					So(result, ShouldResemble, map[string]interface{}{
						"data": map[string]interface{}{
							"fieldWithObjectInput": `{"a":"foo","b":["bar"],"c":"baz"}`,
						},
					})
				})

				Convey("does not use incorrect value", func() {
					input := `
                    {
                        fieldWithObjectInput(input: ["foo", "bar", "baz"])
                    }
                    `
					result, err := executor.Execute(context, input, variables, "")
					So(err, ShouldEqual, nil)
					So(result, ShouldResemble, map[string]interface{}{
						"data": map[string]interface{}{
							"fieldWithObjectInput": nil,
						},
					})
				})

				Convey("properly runs parseLiteral on complex scalar types", func() {
					input := `
                    {
                        fieldWithObjectInput(input: {a: "foo", d: "SerializedValue", c: "bar"})
                    }
                    `
					result, err := executor.Execute(context, input, variables, "")
					So(err, ShouldEqual, nil)
					So(result, ShouldResemble, map[string]interface{}{
						"data": map[string]interface{}{
							"fieldWithObjectInput": `{"a":"foo","c":"bar","d":"DeserializedValue"}`,
						},
					})
				})

				Convey("properly serializes scalar values", func() {
					input := `
                    {
                        deserializedValue(input: "SerializedValue")
                    }
                    `
					result, err := executor.Execute(context, input, variables, "")
					So(err, ShouldEqual, nil)
					So(result, ShouldResemble, map[string]interface{}{
						"data": map[string]interface{}{
							"deserializedValue": "SerializedValue",
						},
					})
				})

				Convey("errors on bad serialization", func() {
					input := `
                    {
                        deserializedValue(input: "BAD")
                    }
                    `
					result, err := executor.Execute(context, input, variables, "")
					So(err, ShouldEqual, nil)
					So(result, ShouldResemble, map[string]interface{}{
						"data": map[string]interface{}{
							"deserializedValue": nil,
						},
						"errors": []map[string]interface{}{
							{
								"message": "Failed to parse ComplexScalar value\n\n1|\n2|                    {\n3|                        deserializedValue(input: \"BAD\")\n                          ^^^^^^^^^^^^^^^^^\n4|                    }\n5|                    ",
								"locations": []map[string]interface{}{
									{
										"column": 25,
										"line":   3,
									},
								},
							},
						},
					})
				})

			})

			Convey("using variables", func() {
				input := `
                query q($input: TestInputObject) {
                    fieldWithObjectInput(input: $input)
                }
                `

				Convey("executes with complex input", func() {
					variables = map[string]interface{}{
						"input": map[string]interface{}{
							"a": "foo",
							"b": []interface{}{"bar"},
							"c": "baz",
						},
					}
					result, err := executor.Execute(context, input, variables, "q")
					So(err, ShouldEqual, nil)
					So(result, ShouldResemble, map[string]interface{}{
						"data": map[string]interface{}{
							"fieldWithObjectInput": `{"a":"foo","b":["bar"],"c":"baz"}`,
						},
					})
				})

				Convey("uses default value when not provided", func() {
					input = `
                    query q($input: TestInputObject = {a: "foo", b: ["bar"], c: "baz"}) {
                        fieldWithObjectInput(input: $input)
                    }
                    `
					result, err := executor.Execute(context, input, variables, "q")
					So(err, ShouldEqual, nil)
					So(result, ShouldResemble, map[string]interface{}{
						"data": map[string]interface{}{
							"fieldWithObjectInput": `{"a":"foo","b":["bar"],"c":"baz"}`,
						},
					})

				})

				Convey("properly parses single value to list", func() {
					variables = map[string]interface{}{
						"input": map[string]interface{}{
							"a": "foo",
							"b": "bar",
							"c": "baz",
						},
					}
					result, err := executor.Execute(context, input, variables, "")
					So(err, ShouldEqual, nil)
					So(result, ShouldResemble, map[string]interface{}{
						"data": map[string]interface{}{
							"fieldWithObjectInput": `{"a":"foo","b":["bar"],"c":"baz"}`,
						},
					})
				})

				Convey("executes with complex scalar input", func() {
					variables = map[string]interface{}{
						"input": map[string]interface{}{
							"c": "foo",
							"d": "SerializedValue",
						},
					}
					result, err := executor.Execute(context, input, variables, "")
					So(err, ShouldEqual, nil)
					So(result, ShouldResemble, map[string]interface{}{
						"data": map[string]interface{}{
							"fieldWithObjectInput": `{"c":"foo","d":"DeserializedValue"}`,
						},
					})
				})

				Convey("errors on null for nested non-null", func() {
					variables = map[string]interface{}{
						"input": map[string]interface{}{
							"a": "foo",
							"b": "bar",
							"c": nil,
						},
					}
					result, err := executor.Execute(context, input, variables, "")
					So(err, ShouldEqual, nil)
					So(result, ShouldResemble, map[string]interface{}{
						"data": map[string]interface{}{
							"fieldWithObjectInput": nil,
						},
						"errors": []map[string]interface{}{
							{
								"locations": []map[string]interface{}{
									{
										"column": 21,
										"line":   3,
									},
								},
								"message": "Variable \"$input\" got invalid value \nIn field \"c\": Expected \"String!\", found null\n\n1|\n2|                query q($input: TestInputObject) {\n3|                    fieldWithObjectInput(input: $input)\n                      ^^^^^^^^^^^^^^^^^^^^\n4|                }\n5|                ",
							},
						},
					})
				})

				Convey("errors on incorrect type", func() {
					variables = map[string]interface{}{
						"input": "foo bar",
					}
					result, err := executor.Execute(context, input, variables, "")
					So(err, ShouldEqual, nil)
					So(result, ShouldResemble, map[string]interface{}{
						"data": map[string]interface{}{
							"fieldWithObjectInput": nil,
						},
						"errors": []map[string]interface{}{
							{
								"locations": []map[string]interface{}{
									{
										"column": 21,
										"line":   3,
									},
								},
								"message": "Variable \"$input\" got invalid value \nExpected \"TestInputObject\", found not an object\n\n1|\n2|                query q($input: TestInputObject) {\n3|                    fieldWithObjectInput(input: $input)\n                      ^^^^^^^^^^^^^^^^^^^^\n4|                }\n5|                ",
							},
						},
					})
				})

				Convey("errors on omission of nested non-null", func() {
					variables = map[string]interface{}{
						"input": map[string]interface{}{
							"a": "foo",
							"b": "bar",
						},
					}
					result, err := executor.Execute(context, input, variables, "")
					So(err, ShouldEqual, nil)
					So(result, ShouldResemble, map[string]interface{}{
						"data": map[string]interface{}{
							"fieldWithObjectInput": nil,
						},
						"errors": []map[string]interface{}{
							{
								"locations": []map[string]interface{}{
									{
										"column": 21,
										"line":   3,
									},
								},
								"message": "Variable \"$input\" got invalid value \nIn field \"c\": Expected \"String!\", found null\n\n1|\n2|                query q($input: TestInputObject) {\n3|                    fieldWithObjectInput(input: $input)\n                      ^^^^^^^^^^^^^^^^^^^^\n4|                }\n5|                ",
							},
						},
					})
				})

				Convey("errors on deep nested errors with first error", func() {
					input = `
                    query q($input: TestNestedInputObject) {
                        fieldWithNestedObjectInput(input: $input)
                    }
                    `
					variables = map[string]interface{}{
						"input": map[string]interface{}{
							"na": map[string]interface{}{
								"a": "foo",
							},
						},
					}
					result, err := executor.Execute(context, input, variables, "q")
					So(err, ShouldEqual, nil)
					So(result, ShouldResemble, map[string]interface{}{
						"data": map[string]interface{}{
							"fieldWithNestedObjectInput": nil,
						},
						"errors": []map[string]interface{}{
							{
								"locations": []map[string]interface{}{
									{
										"column": 25,
										"line":   3,
									},
								},
								"message": "Variable \"$input\" got invalid value \nIn field \"na\": In field \"c\": Expected \"String!\", found null\n\n1|\n2|                    query q($input: TestNestedInputObject) {\n3|                        fieldWithNestedObjectInput(input: $input)\n                          ^^^^^^^^^^^^^^^^^^^^^^^^^^\n4|                    }\n5|                    ",
							},
						},
					})
				})

				Convey("errors on addition of unknown scalars", func() {
					variables = map[string]interface{}{
						"input": map[string]interface{}{
							"a":     "foo",
							"b":     "bar",
							"c":     "baz",
							"extra": "dog",
						},
					}
					result, err := executor.Execute(context, input, variables, "")
					So(err, ShouldEqual, nil)
					So(result, ShouldResemble, map[string]interface{}{
						"data": map[string]interface{}{
							"fieldWithObjectInput": nil,
						},
						"errors": []map[string]interface{}{
							{
								"locations": []map[string]interface{}{
									{
										"column": 21,
										"line":   3,
									},
								},
								"message": "Variable \"$input\" got invalid value \nIn field \"extra\": Unknown field\n\n1|\n2|                query q($input: TestInputObject) {\n3|                    fieldWithObjectInput(input: $input)\n                      ^^^^^^^^^^^^^^^^^^^^\n4|                }\n5|                ",
							},
						},
					})
				})

			})

		})
		Convey("Handles nullable scalars", func() {

			Convey("allows nullable inputs to be omitted", func() {
				input := `
                {
                    fieldWithNullableStringInput
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithNullableStringInput": nil,
					},
				})
			})

			Convey("allows nullable inputs to be omitted in a variable", func() {
				input := `
                query SetsNullable($value: String) {
                    fieldWithNullableStringInput(input: $value)
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithNullableStringInput": nil,
					},
				})
			})

			Convey("allows nullable inputs to be omitted in an unlisted variable", func() {
				input := `
                query SetsNullable {
                    fieldWithNullableStringInput(input: $value)
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithNullableStringInput": nil,
					},
				})
			})

			Convey("allows nullable inputs to be set to null in a variable", func() {
				input := `
                query SetsNullable($value: String) {
                    fieldWithNullableStringInput(input: $value)
                }
                `
				variables = map[string]interface{}{
					"value": nil,
				}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithNullableStringInput": nil,
					},
				})
			})

			Convey("allows nullable inputs to be set to a value in a variable", func() {
				input := `
                query SetsNullable($value: String) {
                    fieldWithNullableStringInput(input: $value)
                }
                `
				variables = map[string]interface{}{
					"value": "a",
				}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithNullableStringInput": `"a"`,
					},
				})
			})

			Convey("allows nullable inputs to be set to a value directly", func() {
				input := `
                {
                    fieldWithNullableStringInput(input: "a")
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithNullableStringInput": `"a"`,
					},
				})
			})
		})

		Convey("Handles non-nullable scalars", func() {

			Convey("does not allow non-nullable inputs to be omitted in a variable", func() {
				input := `
                query SetsNonNullable($value: String!) {
                    fieldWithNonNullableStringInput(input: $value)
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithNonNullableStringInput": nil,
					},
					"errors": []map[string]interface{}{
						{
							"message": "Variable \"$value\" of required type \"String!\" was not provided\n\n1|\n2|                query SetsNonNullable($value: String!) {\n3|                    fieldWithNonNullableStringInput(input: $value)\n                      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^\n4|                }\n5|                ",
							"locations": []map[string]interface{}{
								{
									"column": 21,
									"line":   3,
								},
							},
						},
					},
				})
			})

			Convey("does not allow non-nullable inputs to be set to null in a variable", func() {
				input := `
                query SetsNonNullable($value: String!) {
                    fieldWithNonNullableStringInput(input: $value)
                }
                `
				variables = map[string]interface{}{
					"value": nil,
				}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithNonNullableStringInput": nil,
					},
					"errors": []map[string]interface{}{
						{
							"message": "Variable \"$value\" of required type \"String!\" was not provided\n\n1|\n2|                query SetsNonNullable($value: String!) {\n3|                    fieldWithNonNullableStringInput(input: $value)\n                      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^\n4|                }\n5|                ",
							"locations": []map[string]interface{}{
								{
									"column": 21,
									"line":   3,
								},
							},
						},
					},
				})
			})

			Convey("allows non-nullable inputs to be set to a value in a variable", func() {
				input := `
                query SetsNonNullable($value: String!) {
                    fieldWithNonNullableStringInput(input: $value)
                }
                `

				variables = map[string]interface{}{
					"value": "a",
				}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithNonNullableStringInput": `"a"`,
					},
				})
			})

			Convey("allows non-nullable inputs to be set to a value directly", func() {
				input := `
                {
                    fieldWithNonNullableStringInput(input: "a")
                }
                `

				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithNonNullableStringInput": `"a"`,
					},
				})
			})

			Convey("errors for non-nullable inputs if not explcitly set in the query", func() {
				input := `
                {
                    fieldWithNonNullableStringInput
                }
                `

				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithNonNullableStringInput": nil,
					},
					"errors": []map[string]interface{}{
						{
							"message": "Value required of type \"String!\" was not provided\n\n1|\n2|                {\n3|                    fieldWithNonNullableStringInput\n                      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^\n4|                }\n5|                ",
							"locations": []map[string]interface{}{
								{
									"line":   3,
									"column": 21,
								},
							},
						},
					},
				})
			})
		})

		Convey("Handles lists and nullability", func() {

			Convey("allows lists to be null", func() {
				input := `
                query q($input: [String]) {
                    list(input: $input)
                }
                `
				variables = map[string]interface{}{
					"input": nil,
				}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"list": nil,
					},
				})
			})

			Convey("allows lists to contain values", func() {
				input := `
                query q($input: [String]) {
                    list(input: $input)
                }
                `
				variables = map[string]interface{}{
					"input": []interface{}{"A"},
				}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"list": `["A"]`,
					},
				})
			})

			Convey("allows lists to contain null", func() {
				input := `
                query q($input: [String]) {
                    list(input: $input)
                }
                `
				variables = map[string]interface{}{
					"input": []interface{}{"A", nil, "B"},
				}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"list": `["A",null,"B"]`,
					},
				})
			})

			Convey("does not allow non-null lists to be null", func() {
				input := `
                query q($input: [String]!) {
                    nnList(input: $input)
                }
                `
				variables = map[string]interface{}{
					"input": nil,
				}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"nnList": nil,
					},
					"errors": []map[string]interface{}{
						{
							"message": "Variable \"$input\" of required type \"String!\" was not provided\n\n1|\n2|                query q($input: [String]!) {\n3|                    nnList(input: $input)\n                      ^^^^^^\n4|                }\n5|                ",
							"locations": []map[string]interface{}{
								{
									"column": 21,
									"line":   3,
								},
							},
						},
					},
				})
			})

			Convey("allows non-null lists to contain values", func() {
				input := `
                query q($input: [String]!) {
                    nnList(input: $input)
                }
                `
				variables = map[string]interface{}{
					"input": []interface{}{"A"},
				}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"nnList": `["A"]`,
					},
				})
			})

			Convey("allows lists of non-nulls to be null", func() {
				input := `
                query q($input: [String!]) {
                    listNN(input: $input)
                }
                `
				variables = map[string]interface{}{
					"input": nil,
				}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"listNN": nil,
					},
				})
			})

			Convey("allows lists of non-nulls to contain values", func() {
				input := `
                query q($input: [String!]) {
                    listNN(input: $input)
                }
                `
				variables = map[string]interface{}{
					"input": []interface{}{"A"},
				}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"listNN": `["A"]`,
					},
				})
			})

			Convey("does not allow lists of non-nulls to contain null", func() {
				input := `
                query q($input: [String!]) {
                    listNN(input: $input)
                }
                `
				variables = map[string]interface{}{
					"input": []interface{}{"A", nil, "B"},
				}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"listNN": nil,
					},
					"errors": []map[string]interface{}{
						{
							"message": "Variable \"$input\" got invalid value \nIn element #1: Expected \"String!\", found null\n\n1|\n2|                query q($input: [String!]) {\n3|                    listNN(input: $input)\n                      ^^^^^^\n4|                }\n5|                ",
							"locations": []map[string]interface{}{
								{
									"line":   3,
									"column": 21,
								},
							},
						},
					},
				})
			})

			Convey("does not allow invalid types to be used as values", func() {
				input := `
                query q($input: TestType!) {
                    fieldWithObjectInput(input: $input)
                }
                `
				variables = map[string]interface{}{
					"input": map[string]interface{}{
						"list": []interface{}{"A", "B"},
					},
				}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithObjectInput": nil,
					},
					"errors": []map[string]interface{}{
						{
							"message": "Variable \"$input\" expected value of type \"TestType\" which cannot be used as an input type\n\n1|\n2|                query q($input: TestType!) {\n3|                    fieldWithObjectInput(input: $input)\n                      ^^^^^^^^^^^^^^^^^^^^\n4|                }\n5|                ",
							"locations": []map[string]interface{}{
								{
									"line":   3,
									"column": 21,
								},
							},
						},
					},
				})
			})

			Convey("does not allow unknown types to be used as values", func() {
				input := `
                query q($input: UnknownType!) {
                    fieldWithObjectInput(input: $input)
                }
                `
				variables = map[string]interface{}{
					"input": map[string]interface{}{
						"list": []interface{}{"A", "B"},
					},
				}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithObjectInput": nil,
					},
					"errors": []map[string]interface{}{
						{
							"message": "Variable \"$input\" expected value of type \"UnknownType\" which cannot be used as an input type\n\n1|\n2|                query q($input: UnknownType!) {\n3|                    fieldWithObjectInput(input: $input)\n                      ^^^^^^^^^^^^^^^^^^^^\n4|                }\n5|                ",
							"locations": []map[string]interface{}{
								{
									"line":   3,
									"column": 21,
								},
							},
						},
					},
				})
			})

		})

		Convey("Execute: Uses argument default values", func() {

			Convey("when no argument provided", func() {
				input := `
                {
                    fieldWithDefaultArgumentValue
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithDefaultArgumentValue": `"Hello World"`,
					},
				})
			})

			Convey("when nullable variable provided", func() {
				input := `
                query optionalVariable($optional: String) {
                    fieldWithDefaultArgumentValue(input: $optional)
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithDefaultArgumentValue": `"Hello World"`,
					},
				})
			})

			Convey("when argument provided cannot be parsed", func() {
				input := `
                {
                    fieldWithDefaultArgumentValue(input: WRONG_TYPE)
                }
                `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"fieldWithDefaultArgumentValue": `"Hello World"`,
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

	Convey("Execute: Handle inputs with context", t, func() {
		Convey("Parse Literal", func() {
			schema := `
      scalar FileScalar

      input TestInputObject {
        file: FileScalar
      }

      type TestType {
          deserializedValue(input: FileScalar!): FileScalar
      }
      `
			resolvers := map[string]interface{}{}
			resolvers["TestType/deserializedValue"] = func(params *ResolveParams) (interface{}, error) {
				return params.Args["input"], nil
			}
			executor, err := NewExecutor(schema, "TestType", "", resolvers)
			executor.Debug = true
			So(err, ShouldEqual, nil)
			executor.Scalars["FileScalar"] = &Scalar{
				ParseValue: func(context interface{}, value interface{}) (interface{}, error) {
					if _, ok := value.(string); ok {
						if val, ok := context.(map[string]interface{}); ok {
							if id, ok := val["id"].(string); ok {
								if id == "admin" {
									return "GotAdmin", nil
								}
							}
						}
					}
					return nil, &GraphQLError{
						Message: "Failed to parse value FileScalar value",
					}
				},
				ParseLiteral: func(context interface{}, value interface{}) (interface{}, error) {
					if _, ok := value.(*String); ok {
						if val, ok := context.(map[string]interface{}); ok {
							if id, ok := val["id"].(string); ok {
								if id == "admin" {
									return "GotAdmin", nil
								}
							}
						}
					}
					return nil, &GraphQLError{
						Message: "Failed to parse literal FileScalar value",
					}
				},
				Serialize: func(context interface{}, value interface{}) (interface{}, error) {
					return value, nil
				},
			}
			executor.ResolveType = func(value interface{}) string {
				if object, ok := value.(map[string]interface{}); ok {
					return object["__typename"].(string)
				}
				return ""
			}

			Convey("Cannot Parse Literal on FileScalar when no context", func() {
				context := map[string]interface{}{}
				variables := map[string]interface{}{}
				input := `
          {
            deserializedValue(input: "hello")
          }
        `
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"deserializedValue": nil,
					},
					"errors": []map[string]interface{}{
						map[string]interface{}{
							"locations": []map[string]interface{}{
								map[string]interface{}{
									"line":   3,
									"column": 13,
								},
							},
							"message": "Failed to parse literal FileScalar value\n\n1|\n2|          {\n3|            deserializedValue(input: \"hello\")\n              ^^^^^^^^^^^^^^^^^\n4|          }\n5|        ",
						},
					},
				})
			})

			Convey("Can Parse Literal on FileScalar when context is there", func() {
				input := `
          {
            deserializedValue(input: "hello")
          }
        `
				context := map[string]interface{}{"id": "admin"}
				variables := map[string]interface{}{}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"deserializedValue": "GotAdmin",
					},
				})
			})
		})

		Convey("Parse Value", func() {
			schema := `
      scalar FileScalar

      input TestInputObject {
        file: FileScalar
      }

      type TestType {
          deserializedValue(input: FileScalar!): FileScalar
      }
      `
			resolvers := map[string]interface{}{}
			resolvers["TestType/deserializedValue"] = func(params *ResolveParams) (interface{}, error) {
				return params.Args["input"].(map[string]interface{})["file"], nil
			}
			executor, err := NewExecutor(schema, "TestType", "", resolvers)
			executor.Debug = true
			So(err, ShouldEqual, nil)
			executor.Scalars["FileScalar"] = &Scalar{
				ParseValue: func(context interface{}, value interface{}) (interface{}, error) {
					if _, ok := value.(string); ok {
						if val, ok := context.(map[string]interface{}); ok {
							if id, ok := val["id"].(string); ok {
								if id == "admin" {
									return "GotAdmin", nil
								}
							}
						}
					}
					return nil, &GraphQLError{
						Message: "Failed to parse value FileScalar value",
					}
				},
				ParseLiteral: func(context interface{}, value interface{}) (interface{}, error) {
					if _, ok := value.(*String); ok {
						if val, ok := context.(map[string]interface{}); ok {
							if id, ok := val["id"].(string); ok {
								if id == "admin" {
									return "GotAdmin", nil
								}
							}
						}
					}
					return nil, &GraphQLError{
						Message: "Failed to parse literal FileScalar value",
					}
				},
				Serialize: func(context interface{}, value interface{}) (interface{}, error) {
					return value, nil
				},
			}
			executor.ResolveType = func(value interface{}) string {
				if object, ok := value.(map[string]interface{}); ok {
					return object["__typename"].(string)
				}
				return ""
			}
			Convey("Cannot Parse Value on FileScalar when no context", func() {
				input := `
        query q($input: TestInputObject) {
          deserializedValue(input: $input)
        }
      `
				result, err := executor.Execute(
					map[string]interface{}{},
					input,
					map[string]interface{}{
						"input": map[string]interface{}{
							"file": "hello",
						},
					}, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"deserializedValue": nil,
					},
					"errors": []map[string]interface{}{
						map[string]interface{}{
							"locations": []map[string]interface{}{
								map[string]interface{}{
									"line":   3,
									"column": 11,
								},
							},
							"message": "Variable \"$input\" got invalid value \nIn field \"file\": Failed to parse value FileScalar value\n\n1|\n2|        query q($input: TestInputObject) {\n3|          deserializedValue(input: $input)\n            ^^^^^^^^^^^^^^^^^\n4|        }\n5|      ",
						},
					},
				})
			})

			Convey("Can Parse Value on FileScalar when context is there", func() {
				input := `
        query q($input: TestInputObject) {
          deserializedValue(input: $input)
        }
      `
				context := map[string]interface{}{"id": "admin"}
				variables := map[string]interface{}{
					"input": map[string]interface{}{
						"file": "hello",
					},
				}
				result, err := executor.Execute(context, input, variables, "")
				So(err, ShouldEqual, nil)
				So(result, ShouldResemble, map[string]interface{}{
					"data": map[string]interface{}{
						"deserializedValue": "GotAdmin",
					},
				})
			})
		})
	})

	Convey("Execute: Applies before and after middleware", t, func() {
		schema := `
        type TestType {
            a(replace: String, test: Boolean): String
            b(suffix: String, test: Boolean): String
        }
        `
		resolvers := map[string]interface{}{}
		resolvers["TestType/a"] = &FieldParams{
			Before: func(params *ResolveParams) (interface{}, error) {
				if test, ok := params.Args["test"].(bool); ok && test {
					return nil, errors.New("Test failed")
				}
				if replace, ok := params.Args["replace"].(string); ok {
					return replace, nil
				}
				return nil, nil

			},
			Resolve: func(params *ResolveParams) (interface{}, error) {
				return "a", nil
			},
		}
		resolvers["TestType/b"] = &FieldParams{
			Resolve: func(params *ResolveParams) (interface{}, error) {
				return "b", nil
			},
			After: func(params *ResolveParams, result interface{}) (interface{}, error) {
				if test, ok := params.Args["test"].(bool); ok && test {
					return nil, errors.New("Test failed")
				}
				if suffix, ok := params.Args["suffix"].(string); ok {
					return result.(string) + suffix, nil
				}
				return result, nil
			},
		}
		context := map[string]interface{}{}
		variables := map[string]interface{}{}
		executor, err := NewExecutor(schema, "TestType", "", resolvers)
		So(err, ShouldEqual, nil)

		Convey("before middleware", func() {
			input := `{ a(test: true) }`
			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"a": nil,
				},
				"errors": []map[string]interface{}{
					{
						"locations": []map[string]interface{}{
							{
								"column": 3,
								"line":   1,
							},
						},
						"message": "Test failed",
					},
				},
			})

			input = `{ a(test: false) }`
			result, err = executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"a": "a",
				},
			})

			input = `{ a(replace: "X", test: false) }`
			result, err = executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"a": "X",
				},
			})
		})
		Convey("after middleware", func() {
			input := `{ a, b(suffix: "XX") }`
			result, err := executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"a": "a",
					"b": "bXX",
				},
			})

			input = `{ a, b(suffix: "XX", test: false) }`
			result, err = executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"a": "a",
					"b": "bXX",
				},
			})

			input = `{ a, b(suffix: "XX", test: true) }`
			result, err = executor.Execute(context, input, variables, "")
			So(err, ShouldEqual, nil)
			So(result, ShouldResemble, map[string]interface{}{
				"data": map[string]interface{}{
					"a": "a",
					"b": nil,
				},
				"errors": []map[string]interface{}{
					{
						"locations": []map[string]interface{}{
							{
								"column": 6,
								"line":   1,
							},
						},
						"message": "Test failed",
					},
				},
			})
		})
	})

}

func SetupBenchmark(name string) (*Executor, interface{}, map[string]interface{}) {
	switch name {
	case "simple":
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
		if err != nil {
			panic(err)
		}
		return executor, context, variables
	case "complex":
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
		if err != nil {
			panic(err)
		}

		return executor, context, variables
	default:
		panic("Benchmark not found")
	}
}

func BenchmarkSimpleEmptyRequest(b *testing.B) {
	b.StopTimer()
	executor, context, variables := SetupBenchmark("simple")
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		input := `
            {}
        `
		_, err := executor.Execute(context, input, variables, "")
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkSimpleRequest(b *testing.B) {
	b.StopTimer()
	executor, context, variables := SetupBenchmark("simple")
	b.StartTimer()
	for n := 0; n < b.N; n++ {
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
		_, err := executor.Execute(context, input, variables, "")
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkComplexRequest(b *testing.B) {
	b.StopTimer()
	executor, context, variables := SetupBenchmark("complex")
	b.StartTimer()
	for n := 0; n < b.N; n++ {
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
		_, err := executor.Execute(context, input, variables, "Example")
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkComplexEmptyRequest(b *testing.B) {
	b.StopTimer()
	executor, context, variables := SetupBenchmark("complex")
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		input := `
            {}
        `
		_, err := executor.Execute(context, input, variables, "")
		if err != nil {
			panic(err)
		}
	}
}
