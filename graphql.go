package graphql

type GraphQLParams struct {
	SchemaDefinition string
	QueryRoot        string
	MutationRoot     string
	Resolvers        map[string]interface{}
	ResolveType      func(value interface{}) string
}

func NewGraphQL(params *GraphQLParams) (*Executor, error) {
	executor, err := NewExecutor(params.SchemaDefinition, params.QueryRoot, params.MutationRoot, params.Resolvers)
	if err != nil {
		return nil, err
	}
	if params.ResolveType != nil {
		executor.ResolveType = params.ResolveType
	}
	return executor, nil
}
