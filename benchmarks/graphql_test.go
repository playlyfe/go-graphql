package graphql_test

import (
	"fmt"
	"github.com/graphql-go/graphql"
	pgql "github.com/playlyfe/go-graphql"
	"testing"
)

var query = `
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

func BenchmarkGoGraphQLMaster(b *testing.B) {
	b.StopTimer()
	var DeepDataType *graphql.Object
	var DataType *graphql.Object
	var schema graphql.Schema
	DeepDataType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "DeepDataType",
			Fields: graphql.FieldsThunk(func() graphql.Fields {
				return graphql.Fields{
					"a": &graphql.Field{
						Type: graphql.String,
						Resolve: func(params graphql.ResolveParams) (interface{}, error) {
							return "Already Been Done", nil
						},
					},
					"b": &graphql.Field{
						Type: graphql.String,
						Resolve: func(params graphql.ResolveParams) (interface{}, error) {
							return "Boring", nil
						},
					},
					"c": &graphql.Field{
						Type: graphql.NewList(graphql.String),
						Resolve: func(params graphql.ResolveParams) (interface{}, error) {
							return []interface{}{"Contrived", nil, "Confusing"}, nil
						},
					},
					"deeper": &graphql.Field{
						Type: graphql.NewList(DataType),
						Resolve: func(params graphql.ResolveParams) (interface{}, error) {
							return []interface{}{map[string]interface{}{}, nil, map[string]interface{}{}}, nil
						},
					},
				}
			}),
		})

	DataType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "DataType",
			Fields: graphql.FieldsThunk(func() graphql.Fields {
				return graphql.Fields{
					"a": &graphql.Field{
						Type: graphql.String,
						Resolve: func(params graphql.ResolveParams) (interface{}, error) {
							return "Apple", nil
						},
					},
					"b": &graphql.Field{
						Type: graphql.String,
						Resolve: func(params graphql.ResolveParams) (interface{}, error) {
							return "Banana", nil
						},
					},
					"c": &graphql.Field{
						Type: graphql.String,
						Resolve: func(params graphql.ResolveParams) (interface{}, error) {
							return "Cookie", nil
						},
					},
					"d": &graphql.Field{
						Type: graphql.String,
						Resolve: func(params graphql.ResolveParams) (interface{}, error) {
							return "Donut", nil
						},
					},
					"e": &graphql.Field{
						Type: graphql.String,
						Resolve: func(params graphql.ResolveParams) (interface{}, error) {
							return "Egg", nil
						},
					},
					"f": &graphql.Field{
						Type: graphql.String,
						Resolve: func(params graphql.ResolveParams) (interface{}, error) {
							return "Fish", nil
						},
					},
					"pic": &graphql.Field{
						Type: graphql.String,
						Args: graphql.FieldConfigArgument{
							"size": &graphql.ArgumentConfig{
								Type: graphql.Int,
							},
						},
						Resolve: func(params graphql.ResolveParams) (interface{}, error) {
							var size int32
							var ok bool
							if size, ok = params.Args["size"].(int32); !ok {
								size = 50
							}
							return fmt.Sprintf("Pic of size: %d", size), nil
						},
					},
					"deep": &graphql.Field{
						Type: DeepDataType,
						Resolve: func(params graphql.ResolveParams) (interface{}, error) {
							return map[string]interface{}{}, nil
						},
					},
					"promise": &graphql.Field{
						Type: DataType,
						Resolve: func(params graphql.ResolveParams) (interface{}, error) {
							return map[string]interface{}{}, nil
						},
					},
				}
			}),
		})
	schema, _ = graphql.NewSchema(
		graphql.SchemaConfig{
			Query: DataType,
		},
	)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		variables := map[string]interface{}{}
		graphql.Do(graphql.Params{
			Schema:         schema,
			RequestString:  query,
			VariableValues: variables,
		})
	}
}

var resolvers = map[string]interface{}{
	"DataType/a": func(params *pgql.ResolveParams) (interface{}, error) {
		return "Apple", nil
	},
	"DataType/b": func(params *pgql.ResolveParams) (interface{}, error) {
		return "Banana", nil
	},
	"DataType/c": func(params *pgql.ResolveParams) (interface{}, error) {
		return "Cookie", nil
	},
	"DataType/d": func(params *pgql.ResolveParams) (interface{}, error) {
		return "Donut", nil
	},
	"DataType/e": func(params *pgql.ResolveParams) (interface{}, error) {
		return "Egg", nil
	},
	"DataType/f": func(params *pgql.ResolveParams) (interface{}, error) {
		return "Fish", nil
	},
	"DataType/pic": func(params *pgql.ResolveParams) (interface{}, error) {
		var size int32
		var ok bool
		if size, ok = params.Args["size"].(int32); !ok {
			size = 50
		}
		return fmt.Sprintf("Pic of size: %d", size), nil
	},
	"DataType/deep": func(params *pgql.ResolveParams) (interface{}, error) {
		return map[string]interface{}{}, nil
	},
	"DataType/promise": func(params *pgql.ResolveParams) (interface{}, error) {
		return map[string]interface{}{}, nil
	},
	"DeepDataType/a": func(params *pgql.ResolveParams) (interface{}, error) {
		return "Already Been Done", nil
	},
	"DeepDataType/b": func(params *pgql.ResolveParams) (interface{}, error) {
		return "Boring", nil
	},
	"DeepDataType/c": func(params *pgql.ResolveParams) (interface{}, error) {
		return []interface{}{"Contrived", nil, "Confusing"}, nil
	},
	"DeepDataType/deeper": func(params *pgql.ResolveParams) (interface{}, error) {
		return []interface{}{map[string]interface{}{}, nil, map[string]interface{}{}}, nil
	},
}
var schema2 = `
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

var executor, _ = pgql.NewExecutor(schema2, "DataType", "", resolvers)

func BenchmarkPlaylyfeGraphQLMaster(b *testing.B) {
	for i := 0; i < b.N; i++ {
		context := map[string]interface{}{}
		variables := map[string]interface{}{}
		executor.Execute(context, query, variables, "")
	}
}
