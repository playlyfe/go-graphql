package language

import (
	"sync"
)

type ASTNode interface {
	Free()
}

type Position struct {
	Index  int
	Line   int
	Column int
}

func (node *Position) Free() { PositionPool.Put(node) }

var PositionPool = &sync.Pool{New: func() interface{} { return &Position{} }}

func NewPosition(nodes map[ASTNode]bool) *Position {
	node := PositionPool.Get().(*Position)
	nodes[node] = true
	return node
}

type LOC struct {
	Start  *Position
	End    *Position
	Source string
	ref    int
}

func (node *LOC) Free() { LOCPool.Put(node) }

var LOCPool = &sync.Pool{New: func() interface{} { return &LOC{} }}

func NewLOC(nodes map[ASTNode]bool) *LOC {
	node := LOCPool.Get().(*LOC)
	nodes[node] = true
	return node
}

// TODO: Optimize all indexes
type Document struct {
	Definitions          []ASTNode
	FragmentIndex        map[string]*FragmentDefinition
	ObjectTypeIndex      map[string]*ObjectTypeDefinition
	TypeExtensionIndex   map[string]*TypeExtensionDefinition
	InterfaceTypeIndex   map[string]*InterfaceTypeDefinition
	UnionTypeIndex       map[string]*UnionTypeDefinition
	InputObjectTypeIndex map[string]*InputObjectTypeDefinition
	ScalarTypeIndex      map[string]*ScalarTypeDefinition
	EnumTypeIndex        map[string]*EnumTypeDefinition
	TypeIndex            map[string]ASTNode
	OperationIndex       map[string]*OperationDefinition
	PossibleTypesIndex   map[string][]*ObjectTypeDefinition
	LOC                  *LOC
	ref                  int
}

func (node *Document) Free() { DocumentPool.Put(node) }

var DocumentPool = &sync.Pool{New: func() interface{} { return &Document{} }}

func NewDocument(nodes map[ASTNode]bool) *Document {
	node := DocumentPool.Get().(*Document)
	nodes[node] = true
	return node
}

type OperationDefinition struct {
	Operation               string
	Name                    *Name
	VariableDefinitions     []*VariableDefinition
	VariableDefinitionIndex map[string]*VariableDefinition
	Directives              []*Directive
	DirectiveIndex          map[string]*Directive
	SelectionSet            *SelectionSet
	LOC                     *LOC
	ref                     int
}

func (node *OperationDefinition) Free() { OperationDefinitionPool.Put(node) }

var OperationDefinitionPool = &sync.Pool{New: func() interface{} { return &OperationDefinition{} }}

func NewOperationDefinition(nodes map[ASTNode]bool) *OperationDefinition {
	node := OperationDefinitionPool.Get().(*OperationDefinition)
	nodes[node] = true
	return node
}

type SelectionSet struct {
	Selections []ASTNode
	LOC        *LOC
	ref        int
}

func (node *SelectionSet) Free() { SelectionSetPool.Put(node) }

var SelectionSetPool = &sync.Pool{New: func() interface{} { return &SelectionSet{} }}

func NewSelectionSet(nodes map[ASTNode]bool) *SelectionSet {
	node := SelectionSetPool.Get().(*SelectionSet)
	nodes[node] = true
	return node
}

func (ss *SelectionSet) SelectionNames(doc *Document, whitelist []string, path []string) []string {
	result := []string{}
	whitelistMap := map[string]bool{}
	if path == nil {
		path = []string{}
	}
	if whitelist != nil {
		for _, name := range whitelist {
			whitelistMap[name] = true
		}
	} else {
		whitelistMap = nil
	}
	names := ss.selectionNames(doc, whitelistMap, path, len(path))
	for name, _ := range names {
		result = append(result, name)
	}
	return result
}

func (ss *SelectionSet) selectionNames(doc *Document, whitelistMap map[string]bool, path []string, index int) map[string]bool {
	result := map[string]bool{}
	for _, selection := range ss.Selections {
		switch field := selection.(type) {
		case *InlineFragment:
			if index != 0 {
				names := field.SelectionSet.selectionNames(doc, whitelistMap, path, index)
				for name, _ := range names {
					if _, ok := result[name]; !ok {
						result[name] = true
					}
				}
			} else {
				names := field.SelectionSet.selectionNames(doc, whitelistMap, path, index)
				for name, _ := range names {
					if _, ok := result[name]; !ok {
						result[name] = true
					}
				}
			}
		case *FragmentSpread:
			fragment := doc.FragmentIndex[field.Name.Value]
			if fragment != nil {
				if index != 0 {
					names := fragment.SelectionSet.selectionNames(doc, whitelistMap, path, index)
					for name, _ := range names {
						if _, ok := result[name]; !ok {
							result[name] = true
						}
					}
				} else {
					names := fragment.SelectionSet.selectionNames(doc, whitelistMap, path, index)
					for name, _ := range names {
						if _, ok := result[name]; !ok {
							result[name] = true
						}
					}
				}
			}
		case *Field:
			if index != 0 {
				if field.Name.Value == path[0] {
					names := field.SelectionSet.selectionNames(doc, whitelistMap, path[1:], index-1)
					for name, _ := range names {
						if _, ok := result[name]; !ok {
							result[name] = true
						}
					}
				} else {
					continue
				}
			} else {
				if whitelistMap == nil {
					result[field.Name.Value] = true
				} else if _, ok := whitelistMap[field.Name.Value]; ok {
					result[field.Name.Value] = true
				}
			}
		}
	}
	return result
}

type VariableDefinition struct {
	Variable     *Variable
	Type         ASTNode
	DefaultValue ASTNode
	LOC          *LOC
	ref          int
}

func (node *VariableDefinition) Free() { VariableDefinitionPool.Put(node) }

var VariableDefinitionPool = &sync.Pool{New: func() interface{} { return &VariableDefinition{} }}

func NewVariableDefinition(nodes map[ASTNode]bool) *VariableDefinition {
	node := VariableDefinitionPool.Get().(*VariableDefinition)
	nodes[node] = true
	return node
}

type Variable struct {
	Name *Name
	LOC  *LOC
	ref  int
}

func (node *Variable) Free() { VariablePool.Put(node) }

var VariablePool = &sync.Pool{New: func() interface{} { return &Variable{} }}

func NewVariable(nodes map[ASTNode]bool) *Variable {
	node := VariablePool.Get().(*Variable)
	nodes[node] = true
	return node
}

type Field struct {
	Alias          *Name
	Name           *Name
	Arguments      []*Argument
	ArgumentIndex  map[string]*Argument
	Directives     []*Directive
	DirectiveIndex map[string]*Directive
	SelectionSet   *SelectionSet
	LOC            *LOC
	ref            int
}

func (node *Field) Free() { FieldPool.Put(node) }

var FieldPool = &sync.Pool{New: func() interface{} { return &Field{} }}

func NewField(nodes map[ASTNode]bool) *Field {
	node := FieldPool.Get().(*Field)
	nodes[node] = true
	return node
}

type InlineFragment struct {
	TypeCondition  *NamedType
	Directives     []*Directive
	DirectiveIndex map[string]*Directive
	SelectionSet   *SelectionSet
	LOC            *LOC
	ref            int
}

func (node *InlineFragment) Free() { InlineFragmentPool.Put(node) }

var InlineFragmentPool = &sync.Pool{New: func() interface{} { return &InlineFragment{} }}

func NewInlineFragment(nodes map[ASTNode]bool) *InlineFragment {
	node := InlineFragmentPool.Get().(*InlineFragment)
	nodes[node] = true
	return node
}

type FragmentSpread struct {
	Name           *Name
	Directives     []*Directive
	DirectiveIndex map[string]*Directive
	LOC            *LOC
	ref            int
}

func (node *FragmentSpread) Free() { FragmentSpreadPool.Put(node) }

var FragmentSpreadPool = &sync.Pool{New: func() interface{} { return &FragmentSpread{} }}

func NewFragmentSpread(nodes map[ASTNode]bool) *FragmentSpread {
	node := FragmentSpreadPool.Get().(*FragmentSpread)
	nodes[node] = true
	return node
}

type FragmentDefinition struct {
	Name           *Name
	TypeCondition  *NamedType
	Directives     []*Directive
	DirectiveIndex map[string]*Directive
	SelectionSet   *SelectionSet
	LOC            *LOC
	ref            int
}

func (node *FragmentDefinition) Free() { FragmentDefinitionPool.Put(node) }

var FragmentDefinitionPool = &sync.Pool{New: func() interface{} { return &FragmentDefinition{} }}

func NewFragmentDefinition(nodes map[ASTNode]bool) *FragmentDefinition {
	node := FragmentDefinitionPool.Get().(*FragmentDefinition)
	nodes[node] = true
	return node
}

type Literal struct {
	Type  string
	Value interface{}
	LOC   *LOC
	ref   int
}

func (node *Literal) Free() { LiteralPool.Put(node) }

var LiteralPool = &sync.Pool{New: func() interface{} { return &Literal{} }}

func NewLiteral(nodes map[ASTNode]bool) *Literal {
	node := LiteralPool.Get().(*Literal)
	nodes[node] = true
	return node
}

type List struct {
	Values []ASTNode
	LOC    *LOC
	ref    int
}

func (node *List) Free() { ListPool.Put(node) }

var ListPool = &sync.Pool{New: func() interface{} { return &List{} }}

func NewList(nodes map[ASTNode]bool) *List {
	node := ListPool.Get().(*List)
	nodes[node] = true
	return node
}

type Object struct {
	Fields     []*ObjectField
	FieldIndex map[string]*ObjectField
	LOC        *LOC
	ref        int
}

func (node *Object) Free() { ObjectPool.Put(node) }

var ObjectPool = &sync.Pool{New: func() interface{} { return &Object{} }}

func NewObject(nodes map[ASTNode]bool) *Object {
	node := ObjectPool.Get().(*Object)
	nodes[node] = true
	return node
}

type ObjectField struct {
	Name  *Name
	Value ASTNode
	LOC   *LOC
	ref   int
}

func (node *ObjectField) Free() { ObjectFieldPool.Put(node) }

var ObjectFieldPool = &sync.Pool{New: func() interface{} { return &ObjectField{} }}

func NewObjectField(nodes map[ASTNode]bool) *ObjectField {
	node := ObjectFieldPool.Get().(*ObjectField)
	nodes[node] = true
	return node
}

type ListType struct {
	Type ASTNode
	LOC  *LOC
	ref  int
}

func (node *ListType) Free() { ListTypePool.Put(node) }

var ListTypePool = &sync.Pool{New: func() interface{} { return &ListType{} }}

func NewListType(nodes map[ASTNode]bool) *ListType {
	node := ListTypePool.Get().(*ListType)
	nodes[node] = true
	return node
}

type NonNullType struct {
	Type ASTNode
	LOC  *LOC
	ref  int
}

func (node *NonNullType) Free() { NonNullTypePool.Put(node) }

var NonNullTypePool = &sync.Pool{New: func() interface{} { return &NonNullType{} }}

func NewNonNullType(nodes map[ASTNode]bool) *NonNullType {
	node := NonNullTypePool.Get().(*NonNullType)
	nodes[node] = true
	return node
}

type Name struct {
	Value string
	LOC   *LOC
	ref   int
}

func (node *Name) Free() { NamePool.Put(node) }

var NamePool = &sync.Pool{New: func() interface{} { return &Name{} }}

func NewName(nodes map[ASTNode]bool) *Name {
	node := NamePool.Get().(*Name)
	nodes[node] = true
	return node
}

type NamedType struct {
	Name *Name
	LOC  *LOC
	ref  int
}

func (node *NamedType) Free() { NamedTypePool.Put(node) }

var NamedTypePool = &sync.Pool{New: func() interface{} { return &NamedType{} }}

func NewNamedType(nodes map[ASTNode]bool) *NamedType {
	node := NamedTypePool.Get().(*NamedType)
	nodes[node] = true
	return node
}

type Directive struct {
	Name          *Name
	Arguments     []*Argument
	ArgumentIndex map[string]*Argument
	LOC           *LOC
	ref           int
}

func (node *Directive) Free() { DirectivePool.Put(node) }

var DirectivePool = &sync.Pool{New: func() interface{} { return &Directive{} }}

func NewDirective(nodes map[ASTNode]bool) *Directive {
	node := DirectivePool.Get().(*Directive)
	nodes[node] = true
	return node
}

type Argument struct {
	Name  *Name
	Value ASTNode
	LOC   *LOC
	ref   int
}

func (node *Argument) Free() { ArgumentPool.Put(node) }

var ArgumentPool = &sync.Pool{New: func() interface{} { return &Argument{} }}

func NewArgument(nodes map[ASTNode]bool) *Argument {
	node := ArgumentPool.Get().(*Argument)
	nodes[node] = true
	return node
}

type Int struct {
	Value int32
	LOC   *LOC
	ref   int
}

func (node *Int) Free() { IntPool.Put(node) }

var IntPool = &sync.Pool{New: func() interface{} { return &Int{} }}

func NewInt(nodes map[ASTNode]bool) *Int {
	node := IntPool.Get().(*Int)
	nodes[node] = true
	return node
}

type Float struct {
	Value float32
	LOC   *LOC
	ref   int
}

func (node *Float) Free() { FloatPool.Put(node) }

var FloatPool = &sync.Pool{New: func() interface{} { return &Float{} }}

func NewFloat(nodes map[ASTNode]bool) *Float {
	node := FloatPool.Get().(*Float)
	nodes[node] = true
	return node
}

type String struct {
	Value string
	LOC   *LOC
	ref   int
}

func (node *String) Free() { StringPool.Put(node) }

var StringPool = &sync.Pool{New: func() interface{} { return &String{} }}

func NewString(nodes map[ASTNode]bool) *String {
	node := StringPool.Get().(*String)
	nodes[node] = true
	return node
}

type Boolean struct {
	Value bool
	LOC   *LOC
	ref   int
}

func (node *Boolean) Free() { BooleanPool.Put(node) }

var BooleanPool = &sync.Pool{New: func() interface{} { return &Boolean{} }}

func NewBoolean(nodes map[ASTNode]bool) *Boolean {
	node := BooleanPool.Get().(*Boolean)
	nodes[node] = true
	return node
}

type Enum struct {
	Value string
	LOC   *LOC
	ref   int
}

func (node *Enum) Free() { EnumPool.Put(node) }

var EnumPool = &sync.Pool{New: func() interface{} { return &Enum{} }}

func NewEnum(nodes map[ASTNode]bool) *Enum {
	node := EnumPool.Get().(*Enum)
	nodes[node] = true
	return node
}

type ObjectTypeDefinition struct {
	Name        *Name
	Description string
	Interfaces  []*NamedType
	Fields      []*FieldDefinition
	FieldIndex  map[string]*FieldDefinition
	LOC         *LOC
	ref         int
}

func (node *ObjectTypeDefinition) Free() { ObjectTypeDefinitionPool.Put(node) }

var ObjectTypeDefinitionPool = &sync.Pool{New: func() interface{} { return &ObjectTypeDefinition{} }}

func NewObjectTypeDefinition(nodes map[ASTNode]bool) *ObjectTypeDefinition {
	node := ObjectTypeDefinitionPool.Get().(*ObjectTypeDefinition)
	nodes[node] = true
	return node
}

type FieldDefinition struct {
	Name              *Name
	Description       string
	IsDeprecated      bool
	DeprecationReason string
	Arguments         []*InputValueDefinition
	ArgumentIndex     map[string]*InputValueDefinition
	Type              ASTNode
	LOC               *LOC
	ref               int
}

func (node *FieldDefinition) Free() { FieldDefinitionPool.Put(node) }

var FieldDefinitionPool = &sync.Pool{New: func() interface{} { return &FieldDefinition{} }}

func NewFieldDefinition(nodes map[ASTNode]bool) *FieldDefinition {
	node := FieldDefinitionPool.Get().(*FieldDefinition)
	nodes[node] = true
	return node
}

type InputValueDefinition struct {
	Name         *Name
	Description  string
	Type         ASTNode
	DefaultValue ASTNode
	LOC          *LOC
	ref          int
}

func (node *InputValueDefinition) Free() { InputValueDefinitionPool.Put(node) }

var InputValueDefinitionPool = &sync.Pool{New: func() interface{} { return &InputValueDefinition{} }}

func NewInputValueDefinition(nodes map[ASTNode]bool) *InputValueDefinition {
	node := InputValueDefinitionPool.Get().(*InputValueDefinition)
	nodes[node] = true
	return node
}

type InterfaceTypeDefinition struct {
	Name        *Name
	Description string
	Fields      []*FieldDefinition
	LOC         *LOC
	ref         int
}

func (node *InterfaceTypeDefinition) Free() { InterfaceTypeDefinitionPool.Put(node) }

var InterfaceTypeDefinitionPool = &sync.Pool{New: func() interface{} { return &InterfaceTypeDefinition{} }}

func NewInterfaceTypeDefinition(nodes map[ASTNode]bool) *InterfaceTypeDefinition {
	node := InterfaceTypeDefinitionPool.Get().(*InterfaceTypeDefinition)
	nodes[node] = true
	return node
}

type UnionTypeDefinition struct {
	Name        *Name
	Description string
	Types       []*NamedType
	LOC         *LOC
	ref         int
}

func (node *UnionTypeDefinition) Free() { UnionTypeDefinitionPool.Put(node) }

var UnionTypeDefinitionPool = &sync.Pool{New: func() interface{} { return &UnionTypeDefinition{} }}

func NewUnionTypeDefinition(nodes map[ASTNode]bool) *UnionTypeDefinition {
	node := UnionTypeDefinitionPool.Get().(*UnionTypeDefinition)
	nodes[node] = true
	return node
}

type ScalarTypeDefinition struct {
	Name        *Name
	Description string
	LOC         *LOC
	ref         int
}

func (node *ScalarTypeDefinition) Free() { ScalarTypeDefinitionPool.Put(node) }

var ScalarTypeDefinitionPool = &sync.Pool{New: func() interface{} { return &ScalarTypeDefinition{} }}

func NewScalarTypeDefinition(nodes map[ASTNode]bool) *ScalarTypeDefinition {
	node := ScalarTypeDefinitionPool.Get().(*ScalarTypeDefinition)
	nodes[node] = true
	return node
}

type EnumTypeDefinition struct {
	Name        *Name
	Description string
	Values      []*EnumValueDefinition
	LOC         *LOC
	ref         int
}

func (node *EnumTypeDefinition) Free() { EnumTypeDefinitionPool.Put(node) }

var EnumTypeDefinitionPool = &sync.Pool{New: func() interface{} { return &EnumTypeDefinition{} }}

func NewEnumTypeDefinition(nodes map[ASTNode]bool) *EnumTypeDefinition {
	node := EnumTypeDefinitionPool.Get().(*EnumTypeDefinition)
	nodes[node] = true
	return node
}

type EnumValueDefinition struct {
	Name              *Name
	Description       string
	IsDeprecated      bool
	DeprecationReason string
	LOC               *LOC
	ref               int
}

func (node *EnumValueDefinition) Free() { EnumTypeDefinitionPool.Put(node) }

var EnumValueDefinitionPool = &sync.Pool{New: func() interface{} { return &EnumValueDefinition{} }}

func NewEnumValueDefinition(nodes map[ASTNode]bool) *EnumValueDefinition {
	node := EnumValueDefinitionPool.Get().(*EnumValueDefinition)
	nodes[node] = true
	return node
}

type InputObjectTypeDefinition struct {
	Name        *Name
	Description string
	Fields      []*InputValueDefinition
	FieldIndex  map[string]*InputValueDefinition
	LOC         *LOC
	ref         int
}

func (node *InputObjectTypeDefinition) Free() { InputObjectTypeDefinitionPool.Put(node) }

var InputObjectTypeDefinitionPool = &sync.Pool{New: func() interface{} { return &InputObjectTypeDefinition{} }}

func NewInputObjectTypeDefinition(nodes map[ASTNode]bool) *InputObjectTypeDefinition {
	node := InputObjectTypeDefinitionPool.Get().(*InputObjectTypeDefinition)
	nodes[node] = true
	return node
}

type TypeExtensionDefinition struct {
	Description string
	Definition  *ObjectTypeDefinition
	LOC         *LOC
	ref         int
}

func (node *TypeExtensionDefinition) Free() { TypeExtensionDefinitionPool.Put(node) }

var TypeExtensionDefinitionPool = &sync.Pool{New: func() interface{} { return &TypeExtensionDefinition{} }}

func NewTypeExtensionDefinition(nodes map[ASTNode]bool) *TypeExtensionDefinition {
	node := TypeExtensionDefinitionPool.Get().(*TypeExtensionDefinition)
	nodes[node] = true
	return node
}
