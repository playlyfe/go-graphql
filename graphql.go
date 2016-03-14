package graphql

type GraphQLParams struct {
	SchemaDefinition string
	QueryRoot        string
	MutationRoot     string
	Resolvers        map[string]interface{}
	Scalars          map[string]*Scalar
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
	if params.Scalars != nil {
		executor.Scalars = params.Scalars
	}
	return executor, nil
}
