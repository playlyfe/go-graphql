package language

type Schema struct {
	Document     *Document
	QueryRoot    *ObjectTypeDefinition
	MutationRoot *ObjectTypeDefinition
}

const INTROSPECTION_SCHEMA = `
scalar String
scalar Boolean
scalar Int
scalar Float

type __Schema {
  types: [__Type!]!
  queryType: __Type!
  mutationType: __Type
  directives: [__Directive!]!
}

type __Type {
  kind: __TypeKind!
  name: String
  description: String

  # OBJECT and INTERFACE only
  fields(includeDeprecated: Boolean = false): [__Field!]

  # OBJECT only
  interfaces: [__Type!]

  # INTERFACE and UNION only
  possibleTypes: [__Type!]

  # ENUM only
  enumValues(includeDeprecated: Boolean = false): [__EnumValue!]

  # INPUT_OBJECT only
  inputFields: [__InputValue!]

  # NON_NULL and LIST only
  ofType: __Type
}

type __Field {
  name: String!
  description: String
  args: [__InputValue!]!
  type: __Type!
  isDeprecated: Boolean!
  deprecationReason: String
}

type __InputValue {
  name: String!
  description: String
  type: __Type!
  defaultValue: String
}

type __EnumValue {
  name: String!
  description: String
  isDeprecated: Boolean!
  deprecationReason: String
}

enum __TypeKind {
  SCALAR
  OBJECT
  INTERFACE
  UNION
  ENUM
  INPUT_OBJECT
  LIST
  NON_NULL
}

type __Directive {
  name: String!
  description: String
  args: [__InputValue!]!
  onOperation: Boolean!
  onFragment: Boolean!
  onField: Boolean!
}
`

func introspectType(schema *Document, typeValue interface{}) map[string]interface{} {
	typeIndex := schema.TypeIndex
	switch ttype := typeValue.(type) {
	case *NonNullType:
		return map[string]interface{}{
			"kind":   "NON_NULL",
			"ofType": introspectType(schema, ttype.Type),
		}
	case *ListType:
		return map[string]interface{}{
			"kind":   "LIST",
			"ofType": introspectType(schema, ttype.Type),
		}
	case *NamedType:
		return introspectType(schema, ttype.Name.Value)
	case string:
		typeInfo := map[string]interface{}{
			"name": ttype,
		}
		typeValue := typeIndex[ttype]
		switch __type := typeValue.(type) {
		case *ScalarTypeDefinition:
			typeInfo["kind"] = "SCALAR"
			typeInfo["description"] = __type.Description
		case *ObjectTypeDefinition:
			typeInfo["kind"] = "OBJECT"
			typeInfo["description"] = __type.Description
			interfaceTypes := []map[string]interface{}{}
			for _, namedType := range __type.Interfaces {
				interfaceTypes = append(interfaceTypes, introspectType(schema, namedType.Name.Value))
			}
			typeInfo["interfaces"] = interfaceTypes
		case *InputObjectTypeDefinition:
			typeInfo["kind"] = "INPUT_OBJECT"
			typeInfo["description"] = __type.Description
			inputFields := []map[string]interface{}{}
			for _, inputValueDefinition := range __type.Fields {
				inputFields = append(inputFields, map[string]interface{}{
					"name":         inputValueDefinition.Name.Value,
					"description":  inputValueDefinition.Description,
					"type":         introspectType(schema, inputValueDefinition.Type),
					"defaultValue": BuildValue(inputValueDefinition.DefaultValue),
				})
			}
			typeInfo["inputFields"] = inputFields
		case *InterfaceTypeDefinition:
			typeInfo["kind"] = "INTERFACE"
			typeInfo["description"] = __type.Description
			possibleTypes := []map[string]interface{}{}
			for _, objectType := range schema.PossibleTypesIndex[__type.Name.Value] {
				possibleTypes = append(possibleTypes, introspectType(schema, objectType.Name.Value))
			}
			typeInfo["possibleTypes"] = possibleTypes
		case *UnionTypeDefinition:
			typeInfo["kind"] = "UNION"
			typeInfo["description"] = __type.Description
			possibleTypes := []map[string]interface{}{}
			for _, objectType := range schema.PossibleTypesIndex[__type.Name.Value] {
				possibleTypes = append(possibleTypes, introspectType(schema, objectType.Name.Value))
			}
			typeInfo["possibleTypes"] = possibleTypes
		case *EnumTypeDefinition:
			typeInfo["kind"] = "ENUM"
			typeInfo["description"] = __type.Description
		}
		return typeInfo
	default:
		return nil
	}
}

func NewSchema(schemaDefinition string) (*Schema, map[string]interface{}, error) {
	parser := &Parser{}
	schema := &Schema{}
	ast, err := parser.Parse(schemaDefinition + INTROSPECTION_SCHEMA)
	if err != nil {
		return nil, nil, err
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
		return nil, nil, &GraphQLError{
			Message: "The QueryRoot could not be found",
		}
	}
	if schema.MutationRoot == nil {
		return nil, nil, &GraphQLError{
			Message: "The MutationRoot could not be found",
		}
	}
	// Add implict fields to query root
	schemaField := &FieldDefinition{
		Name: &Name{
			Value: "__schema",
		},
		Description: "The GraphQL schema",
		Type: &NonNullType{
			Type: &NamedType{
				Name: &Name{
					Value: "__Schema",
				},
			},
		},
		Arguments:     []*InputValueDefinition{},
		ArgumentIndex: map[string]*InputValueDefinition{},
	}
	nameArgument := &InputValueDefinition{
		Name: &Name{
			Value: "name",
		},
		Description: "The name of the Type being inspected",
		Type: &NonNullType{
			Type: &NamedType{
				Name: &Name{
					Value: "String",
				},
			},
		},
	}
	typeField := &FieldDefinition{
		Name: &Name{
			Value: "__type",
		},
		Description: "GraphQL Type introspection information",
		Arguments: []*InputValueDefinition{
			nameArgument,
		},
		ArgumentIndex: map[string]*InputValueDefinition{
			"name": nameArgument,
		},
		Type: &NamedType{
			Name: &Name{
				Value: "__Type",
			},
		},
	}
	schema.QueryRoot.FieldIndex["__schema"] = schemaField
	schema.QueryRoot.Fields = append(schema.QueryRoot.Fields, schemaField)
	schema.QueryRoot.FieldIndex["__type"] = typeField
	schema.QueryRoot.Fields = append(schema.QueryRoot.Fields, typeField)
	resolvers := map[string]interface{}{}
	resolvers["QueryRoot/__schema"] = func(params *ResolveParams) (interface{}, error) {
		return map[string]interface{}{
			"queryType":    introspectType(params.Schema, "QueryRoot"),
			"mutationType": introspectType(params.Schema, "MutationRoot"),
			"directives": []map[string]interface{}{
				{
					"name":        "skip",
					"description": "Conditionally exclude a field or fragment during execution",
					"args": []map[string]interface{}{
						{
							"name":        "if",
							"description": "The condition value",
							"type": map[string]interface{}{
								"kind": "NON_NULL",
								"ofType": map[string]interface{}{
									"kind":        "SCALAR",
									"name":        "Boolean",
									"description": "A boolean value representing `true` or `false`",
								},
							},
						},
					},
					"onOperation": false,
					"onField":     true,
					"onFragment":  true,
				},
				{
					"name":        "include",
					"description": "Conditionally include a field or fragment during execution",
					"args": []map[string]interface{}{
						{
							"name":        "if",
							"description": "The condition value",
							"type": map[string]interface{}{
								"kind": "NON_NULL",
								"ofType": map[string]interface{}{
									"kind":        "SCALAR",
									"name":        "Boolean",
									"description": "A boolean value representing `true` or `false`",
								},
							},
						},
					},
					"onOperation": false,
					"onField":     true,
					"onFragment":  true,
				},
			},
		}, nil
	}
	resolvers["__Schema/types"] = func(params *ResolveParams) (interface{}, error) {
		types := []map[string]interface{}{}
		for typeName, _ := range params.Schema.TypeIndex {
			types = append(types, introspectType(params.Schema, typeName))
		}
		return types, nil
	}
	resolvers["QueryRoot/__type"] = func(params *ResolveParams) (interface{}, error) {
		return introspectType(params.Schema, params.Args["name"]), nil
	}
	resolvers["__Type/fields"] = func(params *ResolveParams) (interface{}, error) {
		if typeInfo, ok := params.Source.(map[string]interface{}); ok {
			typeName := typeInfo["name"].(string)
			if typeInfo["kind"] == "OBJECT" {
				__type := params.Schema.TypeIndex[typeName].(*ObjectTypeDefinition)
				fields := []map[string]interface{}{}
				print(params.Args)
				includeDeprecated := params.Args["includeDeprecated"].(bool)
				for _, fieldDefinition := range __type.Fields {
					if !includeDeprecated && fieldDefinition.IsDeprecated {
						continue
					}
					args := []map[string]interface{}{}
					for _, inputValueDefinition := range fieldDefinition.Arguments {
						args = append(args, map[string]interface{}{
							"name":         inputValueDefinition.Name.Value,
							"description":  inputValueDefinition.Description,
							"type":         inputValueDefinition.Type,
							"defaultValue": BuildValue(inputValueDefinition.DefaultValue),
						})
					}
					fields = append(fields, map[string]interface{}{
						"name":              fieldDefinition.Name.Value,
						"description":       fieldDefinition.Description,
						"args":              args,
						"type":              introspectType(params.Schema, fieldDefinition.Type),
						"isDeprecated":      fieldDefinition.IsDeprecated,
						"deprecationReason": fieldDefinition.DeprecationReason,
					})
				}
				return fields, nil
			} else if typeInfo["kind"] == "INTERFACE" {
				__type := params.Schema.TypeIndex[typeName].(*InterfaceTypeDefinition)
				fields := []map[string]interface{}{}
				includeDeprecated := params.Args["includeDeprecated"].(bool)
				for _, fieldDefinition := range __type.Fields {
					if !includeDeprecated && fieldDefinition.IsDeprecated {
						continue
					}
					args := []map[string]interface{}{}
					for _, inputValueDefinition := range fieldDefinition.Arguments {
						args = append(args, map[string]interface{}{
							"name":         inputValueDefinition.Name.Value,
							"description":  inputValueDefinition.Description,
							"type":         inputValueDefinition.Type,
							"defaultValue": BuildValue(inputValueDefinition.DefaultValue),
						})
					}
					fields = append(fields, map[string]interface{}{
						"name":              fieldDefinition.Name.Value,
						"description":       fieldDefinition.Description,
						"args":              args,
						"type":              introspectType(params.Schema, fieldDefinition.Type),
						"isDeprecated":      fieldDefinition.IsDeprecated,
						"deprecationReason": fieldDefinition.DeprecationReason,
					})
				}
				return fields, nil
			}
		}
		return nil, nil
	}

	resolvers["__Type/enumValues"] = func(params *ResolveParams) (interface{}, error) {
		if typeInfo, ok := params.Source.(map[string]interface{}); ok {
			typeName := typeInfo["name"].(string)
			if typeInfo["kind"] == "ENUM" {
				__type := params.Schema.TypeIndex[typeName].(*EnumTypeDefinition)
				includeDeprecated := params.Args["includeDeprecated"].(bool)
				enumValues := []map[string]interface{}{}
				for _, enumValue := range __type.Values {
					if !includeDeprecated && enumValue.IsDeprecated {
						continue
					}
					enumValues = append(enumValues, map[string]interface{}{
						"name":              enumValue.Name.Value,
						"description":       enumValue.Description,
						"isDeprecated":      enumValue.IsDeprecated,
						"deprecationReason": enumValue.DeprecationReason,
					})
				}
				return enumValues, nil
			}
		}
		return nil, nil
	}
	schema.Document = ast
	return schema, resolvers, nil
}
