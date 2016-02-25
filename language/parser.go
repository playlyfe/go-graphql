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
}

type ParseException struct {
	Name           string
	Message        string
	Token          *Token
	ExpectedTokens []TokenType
}

func (e *ParseException) Error() string {
	if len(e.ExpectedTokens) > 0 {
		tokens := []string{}
		for _, token := range e.ExpectedTokens {
			tokens = append(tokens, token.String())
		}
		return fmt.Sprintf("[Line %d, Column %d] Invalid syntax detected, expected one of %s but found: '%s'", e.Token.Start.Line, e.Token.Start.Column, strings.Join(tokens, ", "), e.Token.Val)
	} else {
		return e.Message
	}
}

func (parser *Parser) Parse(input string) (*Document, error) {
	parser.source = input
	parser.tokens = Lex(LexText, input)
	token := <-parser.tokens
	parser.lookahead = &token
	return parser.document()
}

func (parser *Parser) match(symbol TokenType) error {
	if parser.lookahead.Type == symbol {
		parser.prevEnd = parser.lookahead.End
		token := <-parser.tokens
		parser.lookahead = &token
		return nil
	} else {
		return &ParseException{
			Name:           "syntax_error",
			Message:        "Invalid syntax",
			Token:          parser.lookahead,
			ExpectedTokens: []TokenType{parser.lookahead.Type},
		}
	}
}

func (parser *Parser) value() (ASTNode, error) {
	return parser.valueLiteral(false)
}

func (parser *Parser) loc(start *Position) *LOC {
	return &LOC{
		Start:  start,
		End:    parser.prevEnd,
		Source: parser.source[start.Index:parser.prevEnd.Index],
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
	for {
		definition, err := parser.definition()
		if err != nil {
			return nil, err
		}
		switch item := definition.(type) {
		case *FragmentDefinition:
			fragmentIndex[item.Name.Value] = item
		case *ObjectTypeDefinition:
			objectTypeIndex[item.Name.Value] = item
		case *InterfaceTypeDefinition:
			interfaceTypeIndex[item.Name.Value] = item
		case *UnionTypeDefinition:
			unionTypeIndex[item.Name.Value] = item
		}
		definitions = append(definitions, definition)
		if parser.lookahead.Type == EOF {
			break
		}
	}

	return &Document{
		Definitions:        definitions,
		FragmentIndex:      fragmentIndex,
		ObjectTypeIndex:    objectTypeIndex,
		InterfaceTypeIndex: interfaceTypeIndex,
		UnionTypeIndex:     unionTypeIndex,
		LOC:                parser.loc(start),
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
	switch parser.lookahead.Type {
	case FRAGMENT:
		return parser.fragmentDefinition()
	case LBRACE, MUTATION, QUERY:
		return parser.operationDefinition()
	case TYPE, INTERFACE, UNION, SCALAR, ENUM, INPUT:
		return parser.typeDefinition()
	case EXTEND:
		return parser.typeExtensionDefinition()
	default:
		return nil, &ParseException{
			Name:           "syntax_error",
			Message:        "Invalid syntax",
			Token:          parser.lookahead,
			ExpectedTokens: []TokenType{FRAGMENT, LBRACE, MUTATION, QUERY},
		}
	}
}

/**
 * OperationDefinition :
 *  - SelectionSet
 *  - OperationType Name? VariableDefinitions? Directives? SelectionSet
 *
 * OperationType : one of query mutation
 */
func (parser *Parser) operationDefinition() (ASTNode, error) {
	start := parser.lookahead.Start
	switch parser.lookahead.Type {
	case LBRACE:
		selectionSet, err := parser.selectionSet()
		if err != nil {
			return nil, err
		}
		return OperationDefinition{
			Operation:    "query",
			SelectionSet: selectionSet,
			LOC:          parser.loc(start),
		}, nil
	case MUTATION, QUERY:
		node := &OperationDefinition{}
		node.Operation = parser.lookahead.Val
		err := parser.match(parser.lookahead.Type)
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
			node.VariableDefinitions, err = parser.variableDefinitions()
			if err != nil {
				return nil, err
			}
		}
		if parser.lookahead.Type == AT {
			node.Directives, err = parser.directives()
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
		return nil, &ParseException{
			Name:           "syntax_error",
			Message:        "Invalid syntax",
			Token:          parser.lookahead,
			ExpectedTokens: []TokenType{LBRACE, MUTATION, QUERY},
		}
	}
}

/**
 * VariableDefinitions : ( VariableDefinition+ )
 */
func (parser *Parser) variableDefinitions() ([]*VariableDefinition, error) {
	variableDefinitions := []*VariableDefinition{}
	err := parser.match(LPAREN)
	if err != nil {
		return nil, err
	}
	for {
		variableDefinition, err := parser.variableDefinition()
		if err != nil {
			return nil, err
		}
		variableDefinitions = append(variableDefinitions, variableDefinition)
		if parser.lookahead.Type == RPAREN {
			break
		}
	}
	err = parser.match(RPAREN)
	if err != nil {
		return nil, err
	}
	return variableDefinitions, nil
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
		return nil, &ParseException{
			Name:           "syntax_error",
			Message:        "Invalid syntax",
			Token:          parser.lookahead,
			ExpectedTokens: []TokenType{SPREAD, NAME},
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
		node.Alias = nameOrAlias
		node.Name, err = parser.name()
		if err != nil {
			return nil, err
		}
	} else {
		node.Name = nameOrAlias
	}
	if parser.lookahead.Type == LPAREN {
		node.Arguments, err = parser.arguments()
		if err != nil {
			return nil, err
		}
	}
	if parser.lookahead.Type == AT {
		node.Directives, err = parser.directives()
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
func (parser *Parser) arguments() ([]*Argument, error) {
	err := parser.match(LPAREN)
	if err != nil {
		return nil, err
	}
	arguments := []*Argument{}
	for {
		argument, err := parser.argument()
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, argument)
		if parser.lookahead.Type == RPAREN {
			break
		}
	}
	err = parser.match(RPAREN)
	if err != nil {
		return nil, err
	}
	return arguments, nil
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
			node.Directives, err = parser.directives()
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
		node.Directives, err = parser.directives()
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
	err := parser.match(FRAGMENT)
	if err != nil {
		return nil, err
	}
	node := &FragmentDefinition{}
	node.Name, err = parser.fragmentName()
	if err != nil {
		return nil, err
	}
	if parser.lookahead.Val == "on" {
		err = parser.match(NAME)
		if err != nil {
			return nil, err
		}
		node.TypeCondition, err = parser.namedType()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, &ParseException{
			Name:           "syntax_error",
			Message:        "Invalid syntax",
			Token:          parser.lookahead,
			ExpectedTokens: []TokenType{ON},
		}
	}
	if parser.lookahead.Type == AT {
		node.Directives, err = parser.directives()
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
		// @TODO: Need better error here explaining that 'on' is not allowed in fragment names
		return nil, &ParseException{
			Name:           "syntax_error",
			Message:        "Invalid syntax",
			Token:          parser.lookahead,
			ExpectedTokens: []TokenType{NAME},
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
		val, err := strconv.ParseInt(token.Val, 10, 32)
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
		val, err := strconv.ParseFloat(token.Val, 32)
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
	case BOOL:
		token := parser.lookahead
		err := parser.match(BOOL)
		if err != nil {
			return nil, err
		}
		val := false
		if token.Val == "true" {
			val = true
		}
		return &Boolean{
			Value: val,
			LOC:   parser.loc(start),
		}, nil
	case NAME:
		token := parser.lookahead
		err := parser.match(NAME)
		if err != nil {
			return nil, err
		}
		return &Enum{
			Value: token.Val,
			LOC:   parser.loc(start),
		}, nil
	case DOLLAR:
		if !isConstant {
			return parser.variable()
		}
	}
	return nil, &ParseException{
		Name:           "syntax_error",
		Message:        "Invalid syntax",
		Token:          parser.lookahead,
		ExpectedTokens: []TokenType{LBRACE, LBRACK, INT, FLOAT, STRING, BOOL, NAME},
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
	for parser.lookahead.Type != RBRACE {
		field, err := parser.objectField(isConstant)
		if err != nil {
			return nil, err
		}
		node.Fields = append(node.Fields, field)
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
func (parser *Parser) directives() ([]*Directive, error) {
	directives := []*Directive{}
	for {
		directive, err := parser.directive()
		if err != nil {
			return nil, err
		}
		directives = append(directives, directive)
		if parser.lookahead.Type != AT {
			break
		}
	}
	return directives, nil
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
	node.Arguments, err = parser.arguments()
	if err != nil {
		return nil, err
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
func (parser *Parser) typeDefinition() (ASTNode, error) {
	switch parser.lookahead.Type {
	case TYPE:
		return parser.objectTypeDefinition()
	case INTERFACE:
		return parser.interfaceTypeDefinition()
	case UNION:
		return parser.unionTypeDefinition()
	case SCALAR:
		return parser.scalarTypeDefinition()
	case ENUM:
		return parser.enumTypeDefinition()
	case INPUT:
		return parser.inputObjectTypeDefinition()
	default:
		return nil, &ParseException{
			Name:           "syntax_error",
			Message:        "Invalid syntax",
			Token:          parser.lookahead,
			ExpectedTokens: []TokenType{TYPE, INTERFACE, UNION, SCALAR, ENUM, INPUT},
		}
	}
}

/**
 * ObjectTypeDefinition : type Name ImplementsInterfaces? { FieldDefinition+ }
 */
func (parser *Parser) objectTypeDefinition() (*ObjectTypeDefinition, error) {
	var err error
	start := parser.lookahead.Start
	node := &ObjectTypeDefinition{}
	err = parser.match(TYPE)
	if err != nil {
		return nil, err
	}
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}
	if parser.lookahead.Type == IMPLEMENTS {
		node.Interfaces, err = parser.implementsInterfaces()
		if err != nil {
			return nil, err
		}
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
 * ImplementsInterfaces : implements NamedType+
 */
func (parser *Parser) implementsInterfaces() ([]*NamedType, error) {
	err := parser.match(IMPLEMENTS)
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
 * FieldDefinition : Name ArgumentsDefinition? : Type
 */
func (parser *Parser) fieldDefinition() (*FieldDefinition, error) {
	var err error
	node := &FieldDefinition{}
	start := parser.lookahead.Start
	node.Name, err = parser.name()
	if err != nil {
		return nil, err
	}
	if parser.lookahead.Type == LPAREN {
		node.Arguments, err = parser.argumentDefs()
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
func (parser *Parser) argumentDefs() ([]*InputValueDefinition, error) {
	var err error
	err = parser.match(LPAREN)
	if err != nil {
		return nil, err
	}
	arguments := []*InputValueDefinition{}
	for {
		argument, err := parser.inputValueDef()
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, argument)
		if parser.lookahead.Type == RPAREN {
			break
		}
	}
	err = parser.match(RPAREN)
	if err != nil {
		return nil, err
	}
	return arguments, nil
}

/**
 * InputValueDefinition : Name : Type DefaultValue?
 */
func (parser *Parser) inputValueDef() (*InputValueDefinition, error) {
	var err error
	node := &InputValueDefinition{}
	start := parser.lookahead.Start
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
func (parser *Parser) interfaceTypeDefinition() (*InterfaceTypeDefinition, error) {
	var err error
	node := &InterfaceTypeDefinition{}
	start := parser.lookahead.Start
	err = parser.match(INTERFACE)
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
func (parser *Parser) unionTypeDefinition() (*UnionTypeDefinition, error) {
	var err error
	node := &UnionTypeDefinition{}
	start := parser.lookahead.Start
	err = parser.match(UNION)
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
func (parser *Parser) scalarTypeDefinition() (*ScalarTypeDefinition, error) {
	node := &ScalarTypeDefinition{}
	start := parser.lookahead.Start
	err := parser.match(SCALAR)
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
func (parser *Parser) enumTypeDefinition() (*EnumTypeDefinition, error) {
	node := &EnumTypeDefinition{}
	start := parser.lookahead.Start
	err := parser.match(ENUM)
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
func (parser *Parser) inputObjectTypeDefinition() (*InputObjectTypeDefinition, error) {
	node := &InputObjectTypeDefinition{}
	start := parser.lookahead.Start
	err := parser.match(INPUT)
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
	}
	err = parser.match(RBRACE)
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}

/**
 * TypeExtensionDefinition : extend ObjectTypeDefinition
 */
func (parser *Parser) typeExtensionDefinition() (*TypeExtensionDefinition, error) {
	node := &TypeExtensionDefinition{}
	start := parser.lookahead.Start
	err := parser.match(EXTEND)
	if err != nil {
		return nil, err
	}
	node.Definition, err = parser.objectTypeDefinition()
	if err != nil {
		return nil, err
	}
	node.LOC = parser.loc(start)
	return node, nil
}
