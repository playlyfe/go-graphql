package language

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"playlyfe.com/go-graphql/utils"
	"testing"
)

func TestExecutor(t *testing.T) {
	Convey("Executor", t, func() {

		Convey("should be able to execute simple requests", func() {
			schema := `
            type User {
                name: String!
                password: String
            }

            type Viewer {
                user(id: String): User
            }

            type QueryRoot {
                viewer: Viewer
            }

            type MutationRoot {

            }
            `

			input := `
            # Fetch the user
            query name {
                viewer {
                    user(id: "asdasd") {
                        ...F1
                    }
                }
            }
            fragment F1 on User {
                name
                password
            }
            `
			resolvers := map[string]*FieldParams{}
			resolvers["QueryRoot/viewer"] = &FieldParams{
				Resolve: func(params *ResolveParams) (interface{}, error) {
					return map[string]interface{}{
						"user": map[string]interface{}{
							"password": "boo",
						},
					}, nil
				},
			}
			resolvers["User/name"] = &FieldParams{
				Resolve: func(params *ResolveParams) (interface{}, error) {
					println("------------------")
					utils.PrintJSON(params.Args)
					println("------------------")
					return params.Context.(map[string]interface{})["Context"], nil
				},
			}
			context := map[string]interface{}{
				"Context": "ABC",
			}
			variables := map[string]interface{}{}
			executor, err := NewExecutor(schema, resolvers)
			So(err, ShouldEqual, nil)
			result, err := executor.Execute(context, input, variables, "name")
			So(err, ShouldEqual, nil)
			output, err := json.MarshalIndent(result, "\t", "  ")
			println("OUTPUT")
			println(string(output))
		})

	})

}
