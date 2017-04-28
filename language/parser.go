package language

import (
	"fmt"
	"strconv"
	"strings"
)

type Parser struct {
	tokens    chan Token
	lookahead *Token
	prevEnd   *Position
	source    string
	ast       interface{}
	noSource  bool
}

type ParseParams struct {
	Source   string
	NoSource bool
}

func (parser *Parser) Parse(params *ParseParams) (*Document, error) {
	parser.source = params.Source
	parser.noSource = params.NoSource
	parser.tokens = Lex(LexText, parser.source)
	token := <-parser.tokens
	if token.Type == ILLEGAL {
		return nil, &GraphQLError{
			Message: token.Val,
			Source:  parser.source,
			Start:   token.Start,
			End:     token.End,
		}
	}
	parser.lookahead = &token
	return parser.document()
}

func (parser *Parser) match(symbol TokenType) error {
	if parser.lookahead.Type == symbol {
		parser.prevEnd = parser.lookahead.End
		token := <-parser.tokens
		if token.Type == ILLEGAL {
			return &GraphQLError{
				Message: token.Val,
				Source:  parser.source,
				Start:   token.Start,
				End:     token.End,
			}
		}
		parser.lookahead = &token
		return nil
	} else {
		return &GraphQLError{
			Message: fmt.Sprintf("GraphQL Syntax Error (%d:%d) Expected %s, found %s", parser.lookahead.Start.Line, parser.lookahead.Start.Column, symbol, parser.lookahead.String()),
			Source:  parser.source,
			Start:   parser.lookahead.Start,
			End:     parser.lookahead.End,
		}
	}
}

func (parser *Parser) matchName(value string) error {
	if parser.lookahead.Type == NAME && parser.lookahead.Val == value {
		parser.prevEnd = parser.lookahead.End
		token := <-parser.tokens
		parser.lookahead = &token
		return nil
	} else {
		return &GraphQLError{
			Message: fmt.Sprintf("GraphQL Syntax Error (%d:%d) Expected \"%s\", found %s", parser.lookahead.Start.Line, parser.lookahead.Start.Column, value, parser.lookahead.String()),
			Source:  parser.source,
			Start:   parser.lookahead.Start,
			End:     parser.lookahead.End,
		}
	}
}

func (parser *Parser) value() (ASTNode, error) {
	return parser.valueLiteral(false)
}

func (parser *Parser) loc(start *Position) *LOC {
	if parser.noSource {
		return &LOC{
			Start: start,
			End:   parser.prevEnd,
		}
	} else {
		return &LOC{
			Start:  start,
			End:    parser.prevEnd,
			Source: parser.source,
		}
	}
}

func (parser *Parser) name() (*Name, error) {
	token := parser.lookahead
	err := parser.match(NAME)
	if err != nil {
		return nil, err
	}
	return &Name{
		Value: token.Val,
		LOC:   parser.loc(token.Start),
	}, nil
}

func (parser *Parser) description() (string, error) {
	text := ""
	isBody := false
	token := parser.lookahead
	for token.Type == DESCRIPTION {
		text += strings.Trim(token.Val[2:], " ")
		if isBody {
			text += " "
		} else {
			isBody = true
		}
		err := parser.match(DESCRIPTION)
		if err != nil {
			return text, err
		}
		token = parser.lookahead
	}
	return text, nil
}

/**
 * Document : Definition+
 */
func (parser *Parser) document() (*Document, error) {
	start := parser.lookahead.Start
	definitions := []ASTNode{}
	fragmentIndex := map[string]*FragmentDefinition{}
	objectTypeIndex := map[string]*ObjectTypeDefinition{}
	interfaceTypeIndex := map[string]*InterfaceTypeDefinition{}
	unionTypeIndex := map[string]*UnionTypeDefinition{}
	inputObjectTypeIndex := map[string]*InputObjectTypeDefinition{}
	scalarTypeIndex := map[string]*ScalarTypeDefinition{}
	enumTypeIndex := map[string]*EnumTypeDefinition{}
	typeIndex := map[string]ASTNode{}
	typeExtensionIndex := map[string]*TypeExtensionDefinition{}
	possibleTypesIndex := map[string][]*ObjectTypeDefinition{}
	for {
		definition, err := parser.definition()
		if err != nil {
			return nil, err
		}
		switch item := definition.(type) {
		case *FragmentDefinition:
			fragmentIndex[item.Name.Value] = item
			typeIndex[item.Name.Value] = item
		case *ObjectTypeDefinition:
			objectTypeIndex[item.Name.Value] = item
			typeIndex[item.Name.Value] = item
			if len(item.Interfaces) > 0 {
				for _, implementedInterface := range item.Interfaces {
					interfaceName := implementedInterface.Name.Value
					if possibleTypesIndex[interfaceName] == nil {
						possibleTypesIndex[interfaceName] = []*ObjectTypeDefinition{}
					}
					possibleTypesIndex[interfaceName] = append(possibleTypesIndex[interfaceName], item)
				}
			}
		case *TypeExtensionDefinition:
			typeExtensionIndex[item.Definition.Name.Value] = item

		case *InterfaceTypeDefinition:
			interfaceTypeIndex[item.Name.Value] = item
			typeIndex[item.Name.Value] = item
		case *UnionTypeDefinition:
			unionTypeIndex[item.Name.Value] = item
			typeIndex[item.Name.Value] = item
		case *InputObjectTypeDefinition:
			inputObjectTypeIndex[item.Name.Value] = item
			typeIndex[item.Name.Value] = item
		case *EnumTypeDefinition:
			enumTypeIndex[item.Name.Value] = item
			typeIndex[item.Name.Value] = item
		case *ScalarTypeDefinition:
			scalarTypeIndex[item.Name.Value] = item
			typeIndex[item.Name.Value] = item
		}
		definitions = append(definitions, definition)
		if parser.lookahead.Type == EOF {
			break
		}
	}

	// Find possible types for unions
	for _, unionType := range unionTypeIndex {
		unionName := unionType.Name.Value
		if possibleTypesIndex[unionName] == nil {
			possibleTypesIndex[unionName] = []*ObjectTypeDefinition{}
		}
		for _, possibleType := range unionType.Types {
			possibleTypesIndex[unionName] = append(possibleTypesIndex[unionName], objectTypeIndex[possibleType.Name.Value])
		}
	}

	if len(fragmentIndex) == 0 {
		fragmentIndex = nil
	}
	if len(objectTypeIndex) == 0 {
		objectTypeIndex = nil
	}
	if len(typeExtensionIndex) == 0 {
		typeExtensionIndex = nil
	}
	if len(interfaceTypeIndex) == 0 {
		interfaceTypeIndex = nil
	}
	if len(unionTypeIndex) == 0 {
		unionTypeIndex = nil
	}
	if len(inputObjectTypeIndex) == 0 {
		inputObjectTypeIndex = nil
	}
	if len(scalarTypeIndex) == 0 {
		scalarTypeIndex = nil
	}
	if len(enumTypeIndex) == 0 {
		enumTypeIndex = nil
	}
	if len(possibleTypesIndex) == 0 {
		possibleTypesIndex = nil
	}
	if len(typeIndex) == 0 {
		typeIndex = nil
	}

	return &Document{
		Definitions:          definitions,
		FragmentIndex:        fragmentIndex,
		ObjectTypeIndex:      objectTypeIndex,
		TypeExtensionIndex:   typeExtensionIndex,
		InterfaceTypeIndex:   interfaceTypeIndex,
		UnionTypeIndex:       unionTypeIndex,
		InputObjectTypeIndex: inputObjectTypeIndex,
		ScalarTypeIndex:      scalarTypeIndex,
		EnumTypeIndex:        enumTypeIndex,
		PossibleTypesIndex:   possibleTypesIndex,
		TypeIndex:            typeIndex,
		LOC:                  parser.loc(start),
	}, nil
}

/**
 * Definition :
 *   - OperationDefinition
 *   - FragmentDefinition
 *   - TypeDefinition
 *   - TypeExtensionDefinition
 */
func (parser *Parser) definition() (ASTNode, error) {
	description, err := parser.description()
	if err != nil {
		return nil, err
	}
	switch parser.lookahead.Type {
	case NAME:
		switch parser.lookahead.Val {
		case "fragment":
			return parser.fragmentDefinition()
		case "mutation", "query", "subscription":
			return parser.operationDefinition()
		case "type", "interface", "union", "scalar", "enum", "input":
			return parser.typeDefinition(description)
		case "extend":
			return parser.typeExtensionDefinition(description)
		default:
			return nil, &GraphQLError{
				Message: fmt.Sprintf("GraphQL Syntax Error (%d:%d) Unexpected %s", parser.lookahead.Start.Line, parser.lookahead.Start.Column, parser.lookahead.String()),
				Source:  parser.source,
				Start:   parser.lookahead.Start,
				End:     parser.lookahead.End,
			}
		}
	case LBRACE:
		return parser.operationDefinition()
	default:
		return nil, &GraphQLError{
			Message: fmt.Sprintf("GraphQL Syntax Error (%d:%d) Unexpected %s", parser.lookahead.Start.Line, parser.lookahead.Start.Column, parser.lookahead.String()),
			Source:  parser.source,
			Start:   parser.lookahead.Start,
			End:     parser.lookahead.End,
		}
	}
}

/**
 * OperationDefinition :
 *  - SelectionSet
 *  - OperationType Name? VariableDefinitions? Directives? SelectionSet
 *
 * OperationType : one of query mutation subscription
 */
func (parser *Parser) operationDefinition() (ASTNode, error) {
	start := parser.lookahead.Start
	switch parser.lookahead.Type {
	case LBRACE:
		selectionSet, err := parser.selectionSet()
		if err != nil {
			return nil, err
		}
		return &OperationDefinition{
			Operation:    "query",
			SelectionSet: selectionSet,
			LOC:          parser.loc(start),
		}, nil
	case NAME:
		node := &OperationDefinition{}
		node.Operation = parser.lookahead.Val
		err := parser.matchName(parser.lookahead.Val)
		if err != nil {
			return nil, err
		}
		if parser.lookahead.Type == NAME {
			node.Name, err = parser.name()
			if err != nil {
				return nil, err
			}
		}
		if parser.lookahead.Type == LPAREN {
			node.VariableDefinitions, node.VariableDefinitionIndex, err = parser.variableDefinitions()
			if err != nil {
				return nil, err
			}
		}
		if parser.lookahead.Type == AT {
			node.Directives, node.DirectiveIndex, err = parser.directives()
			if err != nil {
				return nil, err
			}
		}
		node.SelectionSet, err = parser.selectionSet()
		if err != nil {
			return nil, err
		}
		node.LOC = parser.loc(start)
		return node, nil
	default:
		return nil, &GraphQLError{
			Message: fmt.Sprintf("GraphQL Syntax Error (%d:%d) Unexpected %s", parser.lookahead.Start.Line, parser.lookahead.Start.Column, parser.lookahead.String()),
			Source:  parser.source,
			Start:   parser.lookahead.Start,
			End:     parser.lookahead.End,
		}
	}
}

/**
 * VariableDefinitions : ( VariableDefinition+ )
 */
func (parser *Parser) variableDefinitions() ([]*VariableDefinition, map[string]*VariableDefinition, error) {
	variableDefinitions := []*VariableDefinition{}
	variableDefinitionIndex := map[string]*VariableDefinition{}
	err := parser.match(LPAREN)
	if err != nil {
		return nil, nil, err
	}
	for {
		variableDefinition, err := parser.variableDefinition()
		if err != nil {
			return nil, nil, err
		}
		variableDefinitions = append(variableDefinitions, variableDefinition)
		variableDefinitionIndex[variableDefinition.Variable.Name.Value] = variableDefinition
		if parser.lookahead.Type == RPAREN {
			break
		}
	}
	if len(variableDefinitions) == 0 {
		variableDefinitionIndex = nil
	}
	err = parser.match(RPAREN)
	if err != nil {
		return nil, nil, err
	}
	return variableDefinitions, variableDefinitionIndex, nil
}

/**
 * VariableDefinition : Variable : Type DefaultValue?
 */
func (parser *Parser) variableDefinition() (*VariableDefinition, error) {
	var err error
	start := parser.lookahead.Start
	node := &VariableDefinition{}
	node.Variable, err = parser.variable()
	if err != nil {
		return nil, err
	}
	err = parser.match(COLON)
	if err != nil {
		return nil, err
	}
	node.Type, err = parser.type_()
	if err != nil {
		return nil, err
	}
	if parser.lookahead.Type == EQ {
		err = parser.match(EQ)
		if err != nil {
			return nil, err
		}
		node.DefaultValue, err = parser.valueLiteral(true)
		if err != nil {
			return nil, err
		}
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * Variable : $ Name
 */
func (parser *Parser) variable() (*Variable, error) {
	var err error
	start := parser.lookahead.Start
	node := &Variable{}
	err = parser.match(DOLLAR)
	if err != nil {
		return nil, err
	}
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * SelectionSet : { Selection+ }
 */
func (parser *Parser) selectionSet() (*SelectionSet, error) {
	start := parser.lookahead.Start
	node := &SelectionSet{}
	err := parser.match(LBRACE)
	if err != nil {
		return nil, err
	}
	for {
		selection, err := parser.selection()
		if err != nil {
			return nil, err
		}
		node.Selections = append(node.Selections, selection)
		if parser.lookahead.Type == RBRACE {
			break
		}
	}
	err = parser.match(RBRACE)
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * Selection :
 *   - Field
 *   - FragmentSpread
 *   - InlineFragment
 */
func (parser *Parser) selection() (ASTNode, error) {
	if parser.lookahead.Type == SPREAD {
		return parser.fragment()
	} else if parser.lookahead.Type == NAME {
		return parser.field()
	} else {
		return nil, &GraphQLError{
			Message: fmt.Sprintf(`GraphQL Syntax Error (%d:%d) Expected a selection or fragment spread, found %s`, parser.lookahead.Start.Line, parser.lookahead.Start.Column, parser.lookahead.String()),
			Source:  parser.source,
			Start:   parser.lookahead.Start,
			End:     parser.lookahead.End,
		}
	}
}

/**
 * Field : Alias? Name Arguments? Directives? SelectionSet?
 *
 * Alias : Name :
 */
func (parser *Parser) field() (*Field, error) {
	var err error
	start := parser.lookahead.Start
	node := &Field{}
	nameOrAlias, err := parser.name()
	if err != nil {
		return nil, err
	}
	if parser.lookahead.Type == COLON {
		err = parser.match(COLON)
		if err != nil {
			return nil, err
		}
		node.Alias = nameOrAlias
		node.Name, err = parser.name()
		if err != nil {
			return nil, err
		}
	} else {
		node.Name = nameOrAlias
	}
	if parser.lookahead.Type == LPAREN {
		node.Arguments, node.ArgumentIndex, err = parser.arguments()
		if err != nil {
			return nil, err
		}
	}
	if parser.lookahead.Type == AT {
		node.Directives, node.DirectiveIndex, err = parser.directives()
		if err != nil {
			return nil, err
		}
	}
	if parser.lookahead.Type == LBRACE {
		node.SelectionSet, err = parser.selectionSet()
		if err != nil {
			return nil, err
		}
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * Arguments : ( Argument+ )
 */
func (parser *Parser) arguments() ([]*Argument, map[string]*Argument, error) {
	err := parser.match(LPAREN)
	if err != nil {
		return nil, nil, err
	}
	arguments := []*Argument{}
	argumentIndex := map[string]*Argument{}
	for {
		argument, err := parser.argument()
		if err != nil {
			return nil, nil, err
		}
		arguments = append(arguments, argument)
		argumentIndex[argument.Name.Value] = argument
		if parser.lookahead.Type == RPAREN {
			break
		}
	}
	err = parser.match(RPAREN)
	if err != nil {
		return nil, nil, err
	}
	return arguments, argumentIndex, nil
}

/**
 * Argument : Name : Value
 */
func (parser *Parser) argument() (*Argument, error) {
	var err error
	start := parser.lookahead.Start
	node := &Argument{}
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}
	err = parser.match(COLON)
	if err != nil {
		return nil, err
	}
	node.Value, err = parser.valueLiteral(false)
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * Corresponds to both FragmentSpread and InlineFragment in the spec.
 *
 * FragmentSpread : ... FragmentName Directives?
 *
 * InlineFragment : ... TypeCondition? Directives? SelectionSet
 */
func (parser *Parser) fragment() (ASTNode, error) {
	start := parser.lookahead.Start
	err := parser.match(SPREAD)
	if err != nil {
		return nil, err
	}
	if parser.lookahead.Type == NAME && parser.lookahead.Val != "on" {
		node := &FragmentSpread{}
		node.Name, err = parser.fragmentName()
		if err != nil {
			return nil, err
		}
		if parser.lookahead.Type == AT {
			node.Directives, node.DirectiveIndex, err = parser.directives()
			if err != nil {
				return nil, err
			}
		}
		node.LOC = parser.loc(start)
		return node, nil
	}
	node := &InlineFragment{}
	if parser.lookahead.Val == "on" {
		err = parser.match(NAME)
		if err != nil {
			return nil, err
		}
		node.TypeCondition, err = parser.namedType()
		if err != nil {
			return nil, err
		}
	}
	if parser.lookahead.Type == AT {
		node.Directives, node.DirectiveIndex, err = parser.directives()
		if err != nil {
			return nil, err
		}
	}
	if parser.lookahead.Type == LBRACE {
		node.SelectionSet, err = parser.selectionSet()
		if err != nil {
			return nil, err
		}
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * FragmentDefinition :
 *   - fragment FragmentName on TypeCondition Directives? SelectionSet
 *
 * TypeCondition : NamedType
 */
func (parser *Parser) fragmentDefinition() (*FragmentDefinition, error) {
	start := parser.lookahead.Start
	err := parser.matchName("fragment")
	if err != nil {
		return nil, err
	}
	node := &FragmentDefinition{}
	node.Name, err = parser.fragmentName()
	if err != nil {
		return nil, err
	}
	err = parser.matchName("on")
	if err != nil {
		return nil, err
	}
	node.TypeCondition, err = parser.namedType()
	if err != nil {
		return nil, err
	}
	if parser.lookahead.Type == AT {
		node.Directives, node.DirectiveIndex, err = parser.directives()
		if err != nil {
			return nil, err
		}
	}
	node.SelectionSet, err = parser.selectionSet()
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * FragmentName : Name but not `on`
 */
func (parser *Parser) fragmentName() (*Name, error) {
	if parser.lookahead.Val == "on" {
		return nil, &GraphQLError{
			Message: fmt.Sprintf("GraphQL Syntax Error (%d:%d) Fragment cannot be named \"on\"", parser.lookahead.Start.Line, parser.lookahead.Start.Column),
			Source:  parser.source,
			Start:   parser.lookahead.Start,
			End:     parser.lookahead.End,
		}
	}
	return parser.name()
}

/**
 * Value[Const] :
 *   - [~Const] Variable
 *   - IntValue
 *   - FloatValue
 *   - StringValue
 *   - BooleanValue
 *   - EnumValue
 *   - ListValue[?Const]
 *   - ObjectValue[?Const]
 *
 * BooleanValue : one of `true` `false`
 *
 * EnumValue : Name but not `true`, `false` or `null`
 */
func (parser *Parser) valueLiteral(isConstant bool) (ASTNode, error) {
	start := parser.lookahead.Start
	switch parser.lookahead.Type {
	case LBRACK:
		return parser.list(isConstant)
	case LBRACE:
		return parser.object(isConstant)
	case INT:
		token := parser.lookahead
		err := parser.match(INT)
		if err != nil {
			return nil, err
		}
		val, err := strconv.ParseInt(token.Val, 10, 64)
		if err != nil {
			return nil, err
		}
		return &Int{
			Value: int32(val),
			LOC:   parser.loc(start),
		}, nil
	case FLOAT:
		token := parser.lookahead
		err := parser.match(FLOAT)
		if err != nil {
			return nil, err
		}
		val, err := strconv.ParseFloat(token.Val, 64)
		if err != nil {
			return nil, err
		}
		return &Float{
			Value: float32(val),
			LOC:   parser.loc(start),
		}, nil
	case STRING:
		token := parser.lookahead
		err := parser.match(STRING)
		if err != nil {
			return nil, err
		}
		return &String{
			Value: token.Val,
			LOC:   parser.loc(start),
		}, nil
	case NAME:
		token := parser.lookahead
		err := parser.match(NAME)
		if err != nil {
			return nil, err
		}
		if token.Val == "true" {
			return &Boolean{
				Value: true,
				LOC:   parser.loc(start),
			}, nil
		} else if token.Val == "false" {
			return &Boolean{
				Value: false,
				LOC:   parser.loc(start),
			}, nil
		}
		if token.Val == "true" || token.Val == "false" || token.Val == "null" {
			return nil, &GraphQLError{
				Message: fmt.Sprintf("GraphQL Syntax Error (%d:%d) Unexpected %s", token.Start.Line, token.Start.Column, token.String()),
				Source:  parser.source,
				Start:   token.Start,
				End:     token.End,
			}
		} else {
			return &Enum{
				Value: token.Val,
				LOC:   parser.loc(start),
			}, nil
		}
	case DOLLAR:
		if !isConstant {
			return parser.variable()
		}
	}

	return nil, &GraphQLError{
		Message: fmt.Sprintf("GraphQL Syntax Error (%d:%d) Unexpected %s", parser.lookahead.Start.Line, parser.lookahead.Start.Column, parser.lookahead.String()),
		Source:  parser.source,
		Start:   parser.lookahead.Start,
		End:     parser.lookahead.End,
	}

}

func (parser *Parser) constValue() (ASTNode, error) {
	return parser.valueLiteral(true)
}

func (parser *Parser) valueValue() (ASTNode, error) {
	return parser.valueLiteral(false)
}

/**
 * ListValue[Const] :
 *   - [ ]
 *   - [ Value[?Const]+ ]
 */
func (parser *Parser) list(isConstant bool) (*List, error) {
	start := parser.lookahead.Start
	node := &List{}
	err := parser.match(LBRACK)
	if err != nil {
		return nil, err
	}
	for parser.lookahead.Type != RBRACK {
		if isConstant {
			item, err := parser.constValue()
			if err != nil {
				return nil, err
			}
			node.Values = append(node.Values, item)
		} else {
			item, err := parser.valueValue()
			if err != nil {
				return nil, err
			}
			node.Values = append(node.Values, item)
		}
	}
	err = parser.match(RBRACK)
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * ObjectValue[Const] :
 *   - { }
 *   - { ObjectField[?Const]+ }
 */
func (parser *Parser) object(isConstant bool) (*Object, error) {
	start := parser.lookahead.Start
	err := parser.match(LBRACE)
	if err != nil {
		return nil, err
	}
	node := &Object{}
	fieldIndex := map[string]*ObjectField{}
	for parser.lookahead.Type != RBRACE {
		field, err := parser.objectField(isConstant)
		if err != nil {
			return nil, err
		}
		node.Fields = append(node.Fields, field)
		fieldIndex[field.Name.Value] = field
	}
	err = parser.match(RBRACE)
	if err != nil {
		return nil, err
	}
	if len(node.Fields) > 0 {
		node.FieldIndex = fieldIndex
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * ObjectField[Const] : Name : Value[?Const]
 */
func (parser *Parser) objectField(isConstant bool) (*ObjectField, error) {
	var err error
	start := parser.lookahead.Start
	node := &ObjectField{}
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}
	err = parser.match(COLON)
	if err != nil {
		return nil, err
	}
	node.Value, err = parser.valueLiteral(isConstant)
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * Directives : Directive+
 */
func (parser *Parser) directives() ([]*Directive, map[string]*Directive, error) {
	directives := []*Directive{}
	directiveIndex := map[string]*Directive{}

	for {
		directive, err := parser.directive()
		if err != nil {
			return nil, nil, err
		}
		directives = append(directives, directive)
		directiveIndex[directive.Name.Value] = directive
		if parser.lookahead.Type != AT {
			break
		}
	}
	return directives, directiveIndex, nil
}

/**
 * Directive : @ Name Arguments?
 */
func (parser *Parser) directive() (*Directive, error) {
	start := parser.lookahead.Start
	node := &Directive{}
	err := parser.match(AT)
	if err != nil {
		return nil, err
	}
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}
	if parser.lookahead.Type == LPAREN {
		node.Arguments, node.ArgumentIndex, err = parser.arguments()
		if err != nil {
			return nil, err
		}
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * Type :
 *   - NamedType
 *   - ListType
 *   - NonNullType
 */
func (parser *Parser) type_() (ASTNode, error) {
	var node ASTNode
	var err error
	start := parser.lookahead.Start
	if parser.lookahead.Type == LBRACK {
		err = parser.match(LBRACK)
		if err != nil {
			return nil, err
		}
		list := &ListType{}
		list.Type, err = parser.type_()
		if err != nil {
			return nil, err
		}
		err = parser.match(RBRACK)
		if err != nil {
			return nil, err
		}
		list.LOC = parser.loc(start)
		node = list
	} else {
		node, err = parser.namedType()
		if err != nil {
			return nil, err
		}
	}

	if parser.lookahead.Type == BANG {
		err = parser.match(BANG)
		if err != nil {
			return nil, err
		}
		return &NonNullType{
			Type: node,
			LOC:  parser.loc(start),
		}, nil
	}
	return node, nil
}

/**
 * NamedType : Name
 */
func (parser *Parser) namedType() (*NamedType, error) {
	var err error
	start := parser.lookahead.Start
	node := &NamedType{}
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * TypeDefinition :
 *   - ObjectTypeDefinition
 *   - InterfaceTypeDefinition
 *   - UnionTypeDefinition
 *   - ScalarTypeDefinition
 *   - EnumTypeDefinition
 *   - InputObjectTypeDefinition
 */
func (parser *Parser) typeDefinition(description string) (ASTNode, error) {
	switch parser.lookahead.Val {
	case "type":
		return parser.objectTypeDefinition(description)
	case "interface":
		return parser.interfaceTypeDefinition(description)
	case "union":
		return parser.unionTypeDefinition(description)
	case "scalar":
		return parser.scalarTypeDefinition(description)
	case "enum":
		return parser.enumTypeDefinition(description)
	case "input":
		return parser.inputObjectTypeDefinition(description)
	default:
		return nil, &GraphQLError{
			Message: fmt.Sprintf("GraphQL Syntax Error (%d:%d) Unexpected %s", parser.lookahead.Start.Line, parser.lookahead.Start.Column, parser.lookahead.String()),
			Source:  parser.source,
			Start:   parser.lookahead.Start,
			End:     parser.lookahead.End,
		}
	}
}

/**
 * ObjectTypeDefinition :  type Name ImplementsInterfaces? { FieldDefinition+ }
 */
func (parser *Parser) objectTypeDefinition(description string) (*ObjectTypeDefinition, error) {
	var err error
	start := parser.lookahead.Start
	node := &ObjectTypeDefinition{
		Description: description,
	}

	err = parser.matchName("type")
	if err != nil {
		return nil, err
	}
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}
	if parser.lookahead.Type == NAME && parser.lookahead.Val == "implements" {
		node.Interfaces, err = parser.implementsInterfaces()
		if err != nil {
			return nil, err
		}
	}

	err = parser.match(LBRACE)
	if err != nil {
		return nil, err
	}
	fields := []*FieldDefinition{}
	fieldIndex := map[string]*FieldDefinition{}
	for parser.lookahead.Type != RBRACE {
		field, err := parser.fieldDefinition()
		if err != nil {
			return nil, err
		}
		fieldIndex[field.Name.Value] = field
		fields = append(fields, field)
	}
	err = parser.match(RBRACE)
	if err != nil {
		return nil, err
	}
	if len(fieldIndex) > 0 {
		node.FieldIndex = fieldIndex
		node.Fields = fields
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * ImplementsInterfaces : implements NamedType+
 */
func (parser *Parser) implementsInterfaces() ([]*NamedType, error) {
	err := parser.matchName("implements")
	if err != nil {
		return nil, err
	}
	types := []*NamedType{}
	for {
		namedType, err := parser.namedType()
		if err != nil {
			return nil, err
		}
		types = append(types, namedType)
		if parser.lookahead.Type == LBRACE {
			break
		}
	}
	return types, nil
}

/**
 * FieldDefinition : Description* Name ArgumentsDefinition? : Type
 */
func (parser *Parser) fieldDefinition() (*FieldDefinition, error) {
	var err error
	node := &FieldDefinition{}
	start := parser.lookahead.Start
	node.Description, err = parser.description()
	if err != nil {
		return nil, err
	}
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}
	if parser.lookahead.Type == LPAREN {
		node.Arguments, node.ArgumentIndex, err = parser.argumentDefs(node)
		if err != nil {
			return nil, err
		}
	}
	err = parser.match(COLON)
	if err != nil {
		return nil, err
	}
	node.Type, err = parser.type_()
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * ArgumentsDefinition : ( InputValueDefinition+ )
 */
func (parser *Parser) argumentDefs(node *FieldDefinition) ([]*InputValueDefinition, map[string]*InputValueDefinition, error) {
	var err error
	err = parser.match(LPAREN)
	if err != nil {
		return nil, nil, err
	}
	arguments := []*InputValueDefinition{}
	argumentIndex := map[string]*InputValueDefinition{}
	for {
		argument, err := parser.inputValueDef()
		if err != nil {
			return nil, nil, err
		}
		arguments = append(arguments, argument)
		argumentIndex[argument.Name.Value] = argument
		if parser.lookahead.Type == RPAREN {
			break
		}
	}
	err = parser.match(RPAREN)
	if err != nil {
		return nil, nil, err
	}
	return arguments, argumentIndex, nil
}

/**
 * InputValueDefinition : Description* Name : Type DefaultValue?
 */
func (parser *Parser) inputValueDef() (*InputValueDefinition, error) {
	var err error
	node := &InputValueDefinition{}
	start := parser.lookahead.Start

	node.Description, err = parser.description()
	if err != nil {
		return nil, err
	}
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}
	err = parser.match(COLON)
	if err != nil {
		return nil, err
	}
	node.Type, err = parser.type_()
	if err != nil {
		return nil, err
	}
	if parser.lookahead.Type == EQ {
		err = parser.match(EQ)
		if err != nil {
			return nil, err
		}
		node.DefaultValue, err = parser.constValue()
		if err != nil {
			return nil, err
		}
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * InterfaceTypeDefinition : interface Name { FieldDefinition+ }
 */
func (parser *Parser) interfaceTypeDefinition(description string) (*InterfaceTypeDefinition, error) {
	var err error
	node := &InterfaceTypeDefinition{
		Description: description,
	}
	start := parser.lookahead.Start
	err = parser.matchName("interface")
	if err != nil {
		return nil, err
	}
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}

	err = parser.match(LBRACE)
	if err != nil {
		return nil, err
	}
	node.Fields = []*FieldDefinition{}
	for parser.lookahead.Type != RBRACE {
		field, err := parser.fieldDefinition()
		if err != nil {
			return nil, err
		}
		node.Fields = append(node.Fields, field)
	}
	err = parser.match(RBRACE)
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * UnionTypeDefinition : union Name = UnionMembers
 */
func (parser *Parser) unionTypeDefinition(description string) (*UnionTypeDefinition, error) {
	var err error
	node := &UnionTypeDefinition{
		Description: description,
	}
	start := parser.lookahead.Start
	err = parser.matchName("union")
	if err != nil {
		return nil, err
	}
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}
	err = parser.match(EQ)
	if err != nil {
		return nil, err
	}
	node.Types, err = parser.unionMembers()
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * UnionMembers :
 *   - NamedType
 *   - UnionMembers | NamedType
 */
func (parser *Parser) unionMembers() ([]*NamedType, error) {
	members := []*NamedType{}
	for {
		namedType, err := parser.namedType()
		if err != nil {
			return nil, err
		}
		members = append(members, namedType)
		if parser.lookahead.Type == PIPE {
			err = parser.match(PIPE)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}
	return members, nil
}

/**
 * ScalarTypeDefinition : scalar Name
 */
func (parser *Parser) scalarTypeDefinition(description string) (*ScalarTypeDefinition, error) {
	node := &ScalarTypeDefinition{
		Description: description,
	}
	start := parser.lookahead.Start
	err := parser.matchName("scalar")
	if err != nil {
		return nil, err
	}
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * EnumTypeDefinition : enum Name { EnumValueDefinition+ }
 */
func (parser *Parser) enumTypeDefinition(description string) (*EnumTypeDefinition, error) {
	node := &EnumTypeDefinition{
		Description: description,
	}
	start := parser.lookahead.Start
	err := parser.matchName("enum")
	if err != nil {
		return nil, err
	}
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}
	err = parser.match(LBRACE)
	if err != nil {
		return nil, err
	}
	for {
		enumType, err := parser.enumValueDefinition()
		if err != nil {
			return nil, err
		}
		node.Values = append(node.Values, enumType)
		if parser.lookahead.Type == RBRACE {
			break
		}
	}
	err = parser.match(RBRACE)
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * EnumValueDefinition : EnumValue
 *
 * EnumValue : Name
 */
func (parser *Parser) enumValueDefinition() (*EnumValueDefinition, error) {
	var err error
	node := &EnumValueDefinition{}
	start := parser.lookahead.Start
	node.Description, err = parser.description()
	if err != nil {
		return nil, err
	}
	if parser.lookahead.Val == "true" || parser.lookahead.Val == "false" || parser.lookahead.Val == "null" {
		return nil, &GraphQLError{
			Message: fmt.Sprintf("GraphQL Syntax Error (%d:%d) Enum value cannot be %q", parser.lookahead.Start.Line, parser.lookahead.Start.Column, parser.lookahead.Val),
			Source:  parser.source,
			Start:   parser.lookahead.Start,
			End:     parser.lookahead.End,
		}
	}
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * InputObjectTypeDefinition : input Name { InputValueDefinition+ }
 */
func (parser *Parser) inputObjectTypeDefinition(description string) (*InputObjectTypeDefinition, error) {
	node := &InputObjectTypeDefinition{
		Description: description,
	}
	fieldIndex := map[string]*InputValueDefinition{}
	start := parser.lookahead.Start
	err := parser.matchName("input")
	if err != nil {
		return nil, err
	}
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}
	err = parser.match(LBRACE)
	if err != nil {
		return nil, err
	}
	for parser.lookahead.Type != RBRACE {
		field, err := parser.inputValueDef()
		if err != nil {
			return nil, err
		}
		node.Fields = append(node.Fields, field)
		fieldIndex[field.Name.Value] = field
	}
	err = parser.match(RBRACE)
	if err != nil {
		return nil, err
	}
	if len(node.Fields) > 0 {
		node.FieldIndex = fieldIndex
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * TypeExtensionDefinition : extend ObjectTypeDefinition
 */
func (parser *Parser) typeExtensionDefinition(description string) (*TypeExtensionDefinition, error) {
	node := &TypeExtensionDefinition{
		Description: description,
	}
	start := parser.lookahead.Start
	err := parser.matchName("extend")
	if err != nil {
		return nil, err
	}
	node.Definition, err = parser.objectTypeDefinition(description)
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}
