package language

type Schema struct {
	Document     *Document
	QueryRoot    *ObjectTypeDefinition
	MutationRoot *ObjectTypeDefinition
}

func NewSchema(schemaDefinition string) (*Schema, error) {
	parser := &Parser{}
	schema := &Schema{}
	ast, err := parser.Parse(schemaDefinition)
	if err != nil {
		return nil, err
	}
	for _, definition := range ast.Definitions {
		switch operationDefinition := definition.(type) {
		case *ObjectTypeDefinition:
			if operationDefinition.Name.Value == "QueryRoot" {
				schema.QueryRoot = operationDefinition
			} else if operationDefinition.Name.Value == "MutationRoot" {
				schema.MutationRoot = operationDefinition
			}
		}
	}
	if schema.QueryRoot == nil {
		return nil, &GraphQLError{
			Message: "The QueryRoot could not be found",
		}
	}
	if schema.MutationRoot == nil {
		return nil, &GraphQLError{
			Message: "The MutationRoot could not be found",
		}
	}
	schema.Document = ast
	return schema, nil
}
