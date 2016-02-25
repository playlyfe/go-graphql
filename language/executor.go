package language

import (
	"log"
	"playlyfe.com/go-graphql/utils"
	"reflect"
	"strings"
)

type ResolveParams struct {
	Context    interface{}
	Source     interface{}
	Args       map[string]interface{}
	Selections []string
}

type Error struct {
	Error error
	Field *Field
}

type RequestContext struct {
	AppContext interface{}
	Document   *Document
	Errors     []*Error
	Variables  map[string]interface{}
}

type FieldParams struct {
	Resolve func(params *ResolveParams) (interface{}, error)
	Before  func(params *ResolveParams) error
	After   func(params *ResolveParams) error
}

type Executor struct {
	ResolveType  func(value interface{}) string
	Schema       *Schema
	Resolvers    map[string]*FieldParams
	ErrorHandler func(err *Error) map[string]interface{}
}

func NewExecutor(schemaDefinition string, resolvers map[string]*FieldParams) (*Executor, error) {
	schema, err := NewSchema(schemaDefinition)
	if err != nil {
		return nil, err
	}
	return &Executor{
		Schema:    schema,
		Resolvers: resolvers,
		ErrorHandler: func(err *Error) map[string]interface{} {
			return map[string]interface{}{
				"message": err.Error.Error(),
				"locations": []map[string]interface{}{
					{
						"line":   err.Field.LOC.Start.Line,
						"column": err.Field.LOC.Start.Column,
					},
				},
			}
		},
	}, nil
}

func (executor *Executor) Execute(context interface{}, request string, variables map[string]interface{}, operationName string) (map[string]interface{}, error) {
	parser := &Parser{}

	document, err := parser.Parse(request)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{}
	reqCtx := &RequestContext{
		AppContext: context,
		Document:   document,
		Errors:     []*Error{},
		Variables:  variables,
	}

	for _, definition := range document.Definitions {
		switch operationDefinition := definition.(type) {
		case *OperationDefinition:
			if (operationDefinition.Name != nil && operationDefinition.Name.Value == operationName) || operationName == "" {
				if operationDefinition.Operation == "query" {
					data, err := executor.selectionSet(reqCtx, executor.Schema.QueryRoot, map[string]interface{}{}, operationDefinition.SelectionSet)
					if err != nil {
						return nil, err
					}
					result["data"] = data
				} else if operationDefinition.Operation == "mutation" {
					data, err := executor.selectionSet(reqCtx, executor.Schema.MutationRoot, map[string]interface{}{}, operationDefinition.SelectionSet)
					if err != nil {
						return nil, err
					}
					result["data"] = data
				}
			}
		}
	}
	if len(reqCtx.Errors) > 0 {
		errs := []map[string]interface{}{}
		for _, err := range reqCtx.Errors {
			errs = append(errs, executor.ErrorHandler(err))
		}
		result["errors"] = errs
	}
	return result, nil
}

func (executor *Executor) selectionSet(reqCtx *RequestContext, objectType *ObjectTypeDefinition, source interface{}, selectionSet *SelectionSet) (map[string]interface{}, error) {
	log.Printf("collecting fields")
	groupedFields, err := executor.collectFields(reqCtx, objectType, selectionSet, &utils.Set{})
	if err != nil {
		return nil, err
	}
	log.Printf("resolving fields")
	return executor.resolveGroupedFields(reqCtx, objectType, source, groupedFields)

}

func (executor *Executor) collectFields(reqCtx *RequestContext, objectType *ObjectTypeDefinition, selectionSet *SelectionSet, visitedFragments *utils.Set) (map[string][]*Field, error) {
	groupedFields := map[string][]*Field{}
	for _, item := range selectionSet.Selections {
		// if skipDirective's if argument is true then continue
		// if includeDirective's if argument is false then continue
		switch selection := item.(type) {
		case *Field:
			var responseKey string
			var groupForResponseKey []*Field
			var ok bool

			if selection.Alias != nil {
				responseKey = selection.Alias.Value
			} else {
				responseKey = selection.Name.Value
			}
			if groupForResponseKey, ok = groupedFields[responseKey]; !ok {
				groupedFields[responseKey] = []*Field{}
				groupForResponseKey = groupedFields[responseKey]
			}
			groupedFields[responseKey] = append(groupForResponseKey, selection)
			log.Printf("adding selection '%s' to grouped fields of '%s'", selection.Name.Value, responseKey)
		case *FragmentSpread:
			var fragmentSpreadName string
			fragmentSpreadName = selection.Name.Value
			if visitedFragments.Has(fragmentSpreadName) {
				continue
			}
			visitedFragments.Add(fragmentSpreadName, true)
			log.Printf("evaluating fragment spread '%s'", fragmentSpreadName)
			fragment, ok := reqCtx.Document.FragmentIndex[fragmentSpreadName]
			if !ok {
				continue
			}
			if !executor.doesFragmentTypeApply(objectType, fragment.TypeCondition) {
				continue
			}
			fragmentGroupedFields, err := executor.collectFields(reqCtx, objectType, fragment.SelectionSet, &utils.Set{})
			if err != nil {
				return nil, err
			}
			for responseKey, fragmentGroup := range fragmentGroupedFields {
				var groupForResponseKey []*Field
				if groupForResponseKey, ok = groupedFields[responseKey]; !ok {
					groupedFields[responseKey] = []*Field{}
					groupForResponseKey = groupedFields[responseKey]
				}
				groupedFields[responseKey] = append(groupForResponseKey, fragmentGroup...)
			}
		case *InlineFragment:
			if selection.TypeCondition != nil && !executor.doesFragmentTypeApply(objectType, selection.TypeCondition) {
				continue
			}
			fragmentGroupedFields, err := executor.collectFields(reqCtx, objectType, selection.SelectionSet, &utils.Set{})
			if err != nil {
				return nil, err
			}
			for responseKey, fragmentGroup := range fragmentGroupedFields {
				var groupForResponseKey []*Field
				var ok bool
				if groupForResponseKey, ok = groupedFields[responseKey]; !ok {
					groupedFields[responseKey] = []*Field{}
					groupForResponseKey = groupedFields[responseKey]
				}
				groupedFields[responseKey] = append(groupForResponseKey, fragmentGroup...)
			}
		}
	}
	return groupedFields, nil
}

func (executor *Executor) doesFragmentTypeApply(objectType *ObjectTypeDefinition, fragmentType *NamedType) bool {
	typeName := fragmentType.Name.Value
	schema := executor.Schema.Document
	if objectType.Name.Value == typeName {
		return true
	}
	if typeDefinition, ok := schema.InterfaceTypeIndex[typeName]; ok {
		for _, implementedInterface := range objectType.Interfaces {
			if implementedInterface.Name.Value == typeDefinition.Name.Value {
				return true
			}
		}
	}
	if typeDefinition, ok := schema.UnionTypeIndex[typeName]; ok {
		for _, possibleType := range typeDefinition.Types {
			if possibleType.Name.Value == typeDefinition.Name.Value {
				return true
			}
		}
	}
	return false
}

func (executor *Executor) resolveGroupedFields(reqCtx *RequestContext, objectType *ObjectTypeDefinition, source interface{}, groupedFields map[string][]*Field) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	// TODO: Use go routines?
	for responseKey, groupForResponseKey := range groupedFields {
		log.Printf("evaluating field entry for '%s'", responseKey)
		key, value, err := executor.getFieldEntry(reqCtx, objectType, source, responseKey, groupForResponseKey)
		if err != nil {
			return nil, err
		}
		log.Printf("Adding '%s' with value '%#v' to response", key, value)
		if key != "" {
			log.Printf("Adding key '%s' with value '%#v' to response", responseKey, value)
			result[responseKey] = value
		}
	}
	return result, nil
}

func (executor *Executor) getFieldEntry(reqCtx *RequestContext, objectType *ObjectTypeDefinition, object interface{}, responseKey string, fields []*Field) (string, interface{}, error) {
	firstField := fields[0]
	log.Printf("Test %#v", firstField.Name.Value)
	fieldType := executor.getFieldTypeFromObjectType(objectType, firstField)
	if fieldType == nil {
		log.Printf("field type of selection '%s' could not be determined", firstField.Name.Value)
		return "", nil, nil
	}
	resolvedObject, err := executor.resolveFieldOnObject(reqCtx, objectType, object, firstField)
	log.Printf("field %s resolved to %#v", firstField.Name.Value, resolvedObject)
	if err != nil {
		return "", nil, err
	}
	if resolvedObject == nil {
		return responseKey, nil, nil
	}
	subSelectionSet := executor.mergeSelectionSets(fields)
	responseValue, err := executor.completeValue(reqCtx, fieldType, resolvedObject, subSelectionSet)
	if err != nil {
		return "", nil, err
	}
	return responseKey, responseValue, nil
}

func (executor *Executor) mergeSelectionSets(fields []*Field) *SelectionSet {
	selectionSet := &SelectionSet{}
	selectionSet.Selections = []ASTNode{}
	for _, field := range fields {
		if field.SelectionSet == nil || len(field.SelectionSet.Selections) == 0 {
			continue
		}
		selectionSet.Selections = append(selectionSet.Selections, field.SelectionSet.Selections...)
	}
	return selectionSet
}

func (executor *Executor) completeValue(reqCtx *RequestContext, fieldType ASTNode, result interface{}, subSelectionSet *SelectionSet) (interface{}, error) {
	//var err error
	log.Printf("completing value on %#v", result)
	if nonNullType, ok := fieldType.(*NonNullType); ok {
		innerType := nonNullType.Type
		completedResult, err := executor.completeValue(reqCtx, innerType, result, nil)
		log.Printf("completed result of %#v is %#v", result, completedResult)
		if err != nil {
			return nil, err
		}
		if completedResult == nil {
			return nil, &GraphQLError{
				Message: "Cannot return null for non-null value",
			}
		}
		return completedResult, nil
	}

	resultVal := reflect.ValueOf(result)
	if !resultVal.IsValid() || result == nil {
		return nil, nil
	}

	if listType, ok := fieldType.(*ListType); ok {
		if resultVal.Type().Kind() != reflect.Slice {
			return nil, &GraphQLError{
				Message: "Expected a list but did not find one",
			}
		}
		innerType := listType.Type
		completedResults := []interface{}{}
		for index := 0; index < resultVal.Len(); index++ {
			val := resultVal.Index(index).Interface()
			completedItem, err := executor.completeValue(reqCtx, innerType, val, nil)
			if err != nil {
				return nil, err
			}
			completedResults = append(completedResults, completedItem)
		}
		return completedResults, nil
	}
	switch typeName := fieldType.(*NamedType).Name.Value; typeName {
	case "Int":
		val, ok := utils.CoerceInt(result)
		if ok {
			return val, nil
		} else {
			return nil, nil
		}
	case "Float":
		val, ok := utils.CoerceFloat(result)
		if ok {
			return val, nil
		} else {
			return nil, nil
		}
	case "String":
		val, ok := utils.CoerceString(result)
		if ok {
			return val, nil
		} else {
			return nil, nil
		}
	case "Boolean":
		val, ok := utils.CoerceBoolean(result)
		if ok {
			return val, nil
		} else {
			return nil, nil
		}
	case "Enum":
		val, ok := utils.CoerceEnum(result)
		if ok {
			return val, nil
		} else {
			return nil, nil
		}
	default:
		log.Printf("completing %s on type %s", result, typeName)
		if objectType, ok := executor.Schema.Document.ObjectTypeIndex[typeName]; ok {
			return executor.selectionSet(reqCtx, objectType, result, subSelectionSet)
		}
		if interfaceType, ok := executor.Schema.Document.InterfaceTypeIndex[typeName]; ok {
			objectType, err := executor.resolveAbstractType(interfaceType, result)
			if err != nil {
				return nil, err
			}
			return executor.selectionSet(reqCtx, objectType, result, subSelectionSet)
		}
		if unionType, ok := executor.Schema.Document.UnionTypeIndex[typeName]; ok {
			objectType, err := executor.resolveAbstractType(unionType, result)
			if err != nil {
				return nil, err
			}
			return executor.selectionSet(reqCtx, objectType, result, subSelectionSet)
		}

		return nil, &GraphQLError{
			Message: "Unknown type",
		}
	}
	return nil, nil
}

func (executor *Executor) resolveFieldOnObject(reqCtx *RequestContext, objectType *ObjectTypeDefinition, object interface{}, firstField *Field) (interface{}, error) {

	sourceVal := reflect.ValueOf(object)
	sourceValType := sourceVal.Type()
	sourceValKind := sourceValType.Kind()
	if sourceVal.IsValid() && sourceValKind == reflect.Ptr {
		sourceVal = sourceVal.Elem()
		sourceValType = sourceVal.Type()
		sourceValKind = sourceValType.Kind()
	}
	println("Resolving", firstField.Name.Value, "on", objectType.Name.Value)
	if !sourceVal.IsValid() {
		return nil, nil
	}

	// try object as a map[string]interface
	if sourceMap, ok := object.(map[string]interface{}); ok {
		if property, ok := sourceMap[firstField.Name.Value]; ok {
			return property, nil
		}
	}

	if sourceValKind == reflect.Struct {
		// find field based on struct's json tag
		// we could potentially create a custom `graphql` tag, but its unnecessary at this point
		// since graphql speaks to client in a json-like way anyway
		// so json tags are a good way to start with
		for i := 0; i < sourceVal.NumField(); i++ {
			valueField := sourceVal.Field(i)
			typeField := sourceValType.Field(i)
			// try matching the field name first
			if typeField.Name == firstField.Name.Value {
				return valueField.Interface(), nil
			}
			tag := typeField.Tag
			jsonTag := tag.Get("json")
			jsonOptions := strings.Split(jsonTag, ",")
			if len(jsonOptions) == 0 {
				continue
			}
			if jsonOptions[0] != firstField.Name.Value {
				continue
			}
			return valueField.Interface(), nil
		}
		return nil, nil
	}

	resolverName := objectType.Name.Value + "/" + firstField.Name.Value
	if resolver, ok := executor.Resolvers[resolverName]; ok {
		result, err := resolver.Resolve(&ResolveParams{
			Context: reqCtx.AppContext,
			Source:  object,
			Args:    executor.buildArguments(firstField.Arguments, reqCtx.Variables),
		})
		if err != nil {
			// TODO: Check how to proceed
			reqCtx.Errors = append(reqCtx.Errors, &Error{
				Error: err,
				Field: firstField,
			})
			return nil, nil
		}
		return result, nil
	}

	// last resort, return nil
	return nil, nil
}

func (executor *Executor) buildArguments(arguments []*Argument, variables map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for _, argument := range arguments {
		result[argument.Name] =
	}
	return result
}

func (executor *Executor) getFieldTypeFromObjectType(objectType *ObjectTypeDefinition, firstField *Field) ASTNode {
	for _, field := range objectType.Fields {
		if field.Name.Value == firstField.Name.Value {
			return field.Type
		}
	}
	return nil
}

func (executor *Executor) resolveAbstractType(abstractType ASTNode, value interface{}) (*ObjectTypeDefinition, error) {
	typeName := executor.ResolveType(value)
	schema := executor.Schema.Document
	if typeName == "" {
		return nil, &GraphQLError{
			Message: "The type of the value could not be determined",
		}
	}
	switch typeValue := abstractType.(type) {
	case *InterfaceTypeDefinition:
		objectType := schema.ObjectTypeIndex[typeName]
		for _, implementedInterface := range objectType.Interfaces {
			if implementedInterface.Name.Value == typeName {
				return objectType, nil
			}
		}
	case *UnionTypeDefinition:
		unionType := schema.UnionTypeIndex[typeValue.Name.Value]
		for _, possibleType := range unionType.Types {
			if possibleType.Name.Value == typeName {
				return schema.ObjectTypeIndex[typeName], nil
			}
		}
	}
	return nil, &GraphQLError{
		Message: "Could not resolve abstract type",
	}
}
