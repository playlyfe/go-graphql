package graphql

import (
	"fmt"

	. "github.com/playlyfe/go-graphql/language"
)

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
scalar ID

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

func (executor *Executor) introspectType(params *ResolveParams, typeValue interface{}) map[string]interface{} {
	schema := executor.Schema.Document
	typeIndex := schema.TypeIndex
	switch ttype := typeValue.(type) {
	case *NonNullType:
		return map[string]interface{}{
			"kind":   "NON_NULL",
			"ofType": executor.introspectType(params, ttype.Type),
		}
	case *ListType:
		return map[string]interface{}{
			"kind":   "LIST",
			"ofType": executor.introspectType(params, ttype.Type),
		}
	case *NamedType:
		return executor.introspectType(params, ttype.Name.Value)
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

		case *InputObjectTypeDefinition:
			typeInfo["kind"] = "INPUT_OBJECT"
			typeInfo["description"] = __type.Description
			inputFields := []map[string]interface{}{}
			for _, inputValueDefinition := range __type.Fields {
				defaultValue, err := executor.valueFromAST(params.Context, inputValueDefinition.DefaultValue, executor.resolveNamedType(inputValueDefinition.Type), nil, nil)
				if err != nil {
					panic(err)
				}
				inputFields = append(inputFields, map[string]interface{}{
					"name":         inputValueDefinition.Name.Value,
					"description":  inputValueDefinition.Description,
					"type":         executor.introspectType(params, inputValueDefinition.Type),
					"defaultValue": defaultValue,
				})
			}
			typeInfo["inputFields"] = inputFields
		case *InterfaceTypeDefinition:
			typeInfo["kind"] = "INTERFACE"
			typeInfo["description"] = __type.Description
			possibleTypes := []map[string]interface{}{}
			for _, objectType := range schema.PossibleTypesIndex[__type.Name.Value] {
				possibleTypes = append(possibleTypes, executor.introspectType(params, objectType.Name.Value))
			}
			typeInfo["possibleTypes"] = possibleTypes
		case *UnionTypeDefinition:
			typeInfo["kind"] = "UNION"
			typeInfo["description"] = __type.Description
			possibleTypes := []map[string]interface{}{}
			for _, objectType := range schema.PossibleTypesIndex[__type.Name.Value] {
				possibleTypes = append(possibleTypes, executor.introspectType(params, objectType.Name.Value))
			}
			typeInfo["possibleTypes"] = possibleTypes
		case *EnumTypeDefinition:
			typeInfo["kind"] = "ENUM"
			typeInfo["description"] = __type.Description
		default:
			panic(fmt.Sprintf("Unknown Type %s", ttype))
		}
		return typeInfo
	default:
		return nil
	}
}

func typenameResolver(typename string) func(params *ResolveParams) (interface{}, error) {
	return func(params *ResolveParams) (interface{}, error) {
		return typename, nil
	}
}

func NewSchema(schemaDefinition string, queryRoot string, mutationRoot string) (*Schema, map[string]interface{}, error) {
	parser := &Parser{}
	schema := &Schema{}
	resolvers := map[string]interface{}{}
	ast, err := parser.Parse(&ParseParams{
		Source: schemaDefinition + INTROSPECTION_SCHEMA,
	})
	if err != nil {
		return nil, nil, err
	}
	for _, definition := range ast.Definitions {
		switch operationDefinition := definition.(type) {
		case *ObjectTypeDefinition:
			if operationDefinition.Name.Value == queryRoot {
				schema.QueryRoot = operationDefinition
			} else if operationDefinition.Name.Value == mutationRoot {
				schema.MutationRoot = operationDefinition
			}
			resolvers[operationDefinition.Name.Value+"/__typename"] = typenameResolver(operationDefinition.Name.Value)
		}
	}
	if schema.QueryRoot == nil {
		return nil, nil, &GraphQLError{
			Message: "The QueryRoot could not be found",
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
	resolvers[queryRoot+"/__schema"] = func(params *ResolveParams) (interface{}, error) {
		executor := params.Executor

		result := map[string]interface{}{
			"queryType": executor.introspectType(params, queryRoot),
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
		}

		//TODO: better handling for empty mutationRoot
		if schema.MutationRoot != nil {
			result["mutationType"] = executor.introspectType(params, mutationRoot)
		}

		return result, nil
	}
	resolvers["__Schema/types"] = func(params *ResolveParams) (interface{}, error) {
		types := []map[string]interface{}{}
		for typeName, _ := range params.Schema.TypeIndex {
			if typeName == "__Schema" || typeName == "__Type" || typeName == "__Field" || typeName == "__InputValue" || typeName == "__EnumValue" || typeName == "__TypeKind" || typeName == "__Directive" {
				continue
			}
			types = append(types, params.Executor.introspectType(params, typeName))
		}
		return types, nil
	}
	resolvers[queryRoot+"/__type"] = func(params *ResolveParams) (interface{}, error) {
		return params.Executor.introspectType(params, params.Args["name"]), nil
	}
	resolvers["__Type/fields"] = func(params *ResolveParams) (interface{}, error) {
		executor := params.Executor
		if typeInfo, ok := params.Source.(map[string]interface{}); ok {
			typeName := typeInfo["name"].(string)
			if typeInfo["kind"] == "OBJECT" {
				__type := params.Schema.TypeIndex[typeName].(*ObjectTypeDefinition)
				fields := []map[string]interface{}{}
				includeDeprecated := params.Args["includeDeprecated"].(bool)
				for _, fieldDefinition := range __type.Fields {
					if !includeDeprecated && fieldDefinition.IsDeprecated {
						continue
					}
					if typeName == queryRoot {
						fieldName := fieldDefinition.Name.Value
						if fieldName == "__schema" || fieldName == "__type" {
							continue
						}
					}
					args := []map[string]interface{}{}
					for _, inputValueDefinition := range fieldDefinition.Arguments {
						defaultValue, err := executor.valueFromAST(params.Context, inputValueDefinition.DefaultValue, executor.resolveNamedType(inputValueDefinition.Type), nil, nil)
						if err != nil {
							return nil, err
						}
						args = append(args, map[string]interface{}{
							"name":         inputValueDefinition.Name.Value,
							"description":  inputValueDefinition.Description,
							"type":         executor.introspectType(params, inputValueDefinition.Type),
							"defaultValue": defaultValue,
						})
					}
					fields = append(fields, map[string]interface{}{
						"name":              fieldDefinition.Name.Value,
						"description":       fieldDefinition.Description,
						"args":              args,
						"type":              executor.introspectType(params, fieldDefinition.Type),
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
						defaultValue, err := executor.valueFromAST(params.Context, inputValueDefinition.DefaultValue, executor.resolveNamedType(inputValueDefinition.Type), nil, nil)
						if err != nil {
							return nil, err
						}
						args = append(args, map[string]interface{}{
							"name":         inputValueDefinition.Name.Value,
							"description":  inputValueDefinition.Description,
							"type":         executor.introspectType(params, inputValueDefinition.Type),
							"defaultValue": defaultValue,
						})
					}
					fields = append(fields, map[string]interface{}{
						"name":              fieldDefinition.Name.Value,
						"description":       fieldDefinition.Description,
						"args":              args,
						"type":              executor.introspectType(params, fieldDefinition.Type),
						"isDeprecated":      fieldDefinition.IsDeprecated,
						"deprecationReason": fieldDefinition.DeprecationReason,
					})
				}
				return fields, nil
			}
		}
		return nil, nil
	}
	resolvers["__Type/interfaces"] = func(params *ResolveParams) (interface{}, error) {
		if typeInfo, ok := params.Source.(map[string]interface{}); ok {
			if typeName, ok := typeInfo["name"].(string); ok {
				__type := params.Schema.ObjectTypeIndex[typeName]
				if __type != nil {
					interfaceTypes := []map[string]interface{}{}
					if __type.Interfaces != nil {
						for _, namedType := range __type.Interfaces {
							interfaceTypes = append(interfaceTypes, params.Executor.introspectType(params, namedType.Name.Value))
						}
					}
					return interfaceTypes, nil
				}
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
