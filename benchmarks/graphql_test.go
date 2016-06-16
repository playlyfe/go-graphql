package graphql_test

import (
	"testing"

	"github.com/graphql-go/graphql"
	pgql "github.com/playlyfe/go-graphql"
)

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: graphql.NewObject(
			graphql.ObjectConfig{
				Name: "RootQueryType",
				Fields: graphql.Fields{
					"hello": &graphql.Field{
						Type: graphql.String,
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							return "world", nil
						},
					},
				},
			}),
	},
)

func BenchmarkGoGraphQLMaster(b *testing.B) {
	for i := 0; i < b.N; i++ {
		variables := map[string]interface{}{}
		graphql.Do(graphql.Params{
			Schema:         schema,
			RequestString:  "{hello}",
			VariableValues: variables,
		})
	}
}

var schema2 = `
    type RootQueryType {
        hello: String
    }
  `
var resolvers = map[string]interface{}{
	"RootQueryType/hello": func(params *pgql.ResolveParams) (interface{}, error) {
		return "world", nil
	},
}
var executor, _ = pgql.NewExecutor(schema2, "RootQueryType", "", resolvers)

func BenchmarkPlaylyfeGraphQLMaster(b *testing.B) {
	for i := 0; i < b.N; i++ {
		context := map[string]interface{}{}
		variables := map[string]interface{}{}
		executor.Execute(context, "{hello}", variables, "")
	}
}
