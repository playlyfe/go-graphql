package language

import (
	"log"
	"sync"
)

type ASTNode interface {
	Free()
}

var GlobalRefList = map[ASTNode]bool{}

type Position struct {
	Index  int
	Line   int
	Column int
	ref    int
}

func (node *Position) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		//log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	PositionPool.Put(node)
}

var PositionPool = &sync.Pool{New: func() interface{} { return &Position{} }}

func NewPosition() *Position {
	node := PositionPool.Get().(*Position)
	node.ref = 1
	return node
}

type LOC struct {
	Start  *Position
	End    *Position
	Source string
	ref    int
}

func (node *LOC) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		//log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Start.Free()
	node.End.Free()
	LOCPool.Put(node)
}

var LOCPool = &sync.Pool{New: func() interface{} { return &LOC{} }}

func NewLOC() *LOC {
	node := LOCPool.Get().(*LOC)
	node.ref = 1
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

func (node *Document) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	for _, definition := range node.Definitions {
		definition.Free()
	}
	DocumentPool.Put(node)
}

var DocumentPool = &sync.Pool{New: func() interface{} { return &Document{} }}

func NewDocument() *Document {
	node := DocumentPool.Get().(*Document)
	node.ref = 1
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

func (node *OperationDefinition) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	if node.Name != nil {
		node.Name.Free()
	}
	for _, variableDefinition := range node.VariableDefinitions {
		variableDefinition.Free()
	}
	for _, directive := range node.Directives {
		directive.Free()
	}
	node.SelectionSet.Free()
	node.LOC.Free()
	OperationDefinitionPool.Put(node)
}

var OperationDefinitionPool = &sync.Pool{New: func() interface{} { return &OperationDefinition{} }}

func NewOperationDefinition() *OperationDefinition {
	node := OperationDefinitionPool.Get().(*OperationDefinition)
	node.ref = 1
	return node
}

type SelectionSet struct {
	Selections []ASTNode
	LOC        *LOC
	ref        int
}

func (node *SelectionSet) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	for _, selection := range node.Selections {
		selection.Free()
	}
	node.LOC.Free()
	SelectionSetPool.Put(node)
}

var SelectionSetPool = &sync.Pool{New: func() interface{} { return &SelectionSet{} }}

func NewSelectionSet() *SelectionSet {
	node := SelectionSetPool.Get().(*SelectionSet)
	node.ref = 1
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

func (node *VariableDefinition) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Variable.Free()
	node.Type.Free()
	if node.DefaultValue != nil {
		node.DefaultValue.Free()
	}
	node.LOC.Free()
	VariableDefinitionPool.Put(node)
}

var VariableDefinitionPool = &sync.Pool{New: func() interface{} { return &VariableDefinition{} }}

func NewVariableDefinition() *VariableDefinition {
	node := VariableDefinitionPool.Get().(*VariableDefinition)
	node.ref = 1
	return node
}

type Variable struct {
	Name *Name
	LOC  *LOC
	ref  int
}

func (node *Variable) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	node.LOC.Free()
	VariablePool.Put(node)
}

var VariablePool = &sync.Pool{New: func() interface{} { return &Variable{} }}

func NewVariable() *Variable {
	node := VariablePool.Get().(*Variable)
	node.ref = 1
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

func (node *Field) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	if node.Alias != nil {
		node.Alias.Free()
	}
	node.Name.Free()
	for _, argument := range node.Arguments {
		argument.Free()
	}
	for _, directive := range node.Directives {
		directive.Free()
	}
	if node.SelectionSet != nil {
		node.SelectionSet.Free()
	}
	node.LOC.Free()
	FieldPool.Put(node)
}

var FieldPool = &sync.Pool{New: func() interface{} { return &Field{} }}

func NewField() *Field {
	node := FieldPool.Get().(*Field)
	node.ref = 1
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

func (node *InlineFragment) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	if node.TypeCondition != nil {
		node.TypeCondition.Free()
	}
	for _, directive := range node.Directives {
		directive.Free()
	}
	if node.SelectionSet != nil {
		node.SelectionSet.Free()
	}
	node.LOC.Free()
	InlineFragmentPool.Put(node)
}

var InlineFragmentPool = &sync.Pool{New: func() interface{} { return &InlineFragment{} }}

func NewInlineFragment() *InlineFragment {
	node := InlineFragmentPool.Get().(*InlineFragment)
	node.ref = 1
	return node
}

type FragmentSpread struct {
	Name           *Name
	Directives     []*Directive
	DirectiveIndex map[string]*Directive
	LOC            *LOC
	ref            int
}

func (node *FragmentSpread) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	for _, directive := range node.Directives {
		directive.Free()
	}
	node.LOC.Free()
	FragmentSpreadPool.Put(node)
}

var FragmentSpreadPool = &sync.Pool{New: func() interface{} { return &FragmentSpread{} }}

func NewFragmentSpread() *FragmentSpread {
	node := FragmentSpreadPool.Get().(*FragmentSpread)
	node.ref = 1
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

func (node *FragmentDefinition) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	node.TypeCondition.Free()
	for _, directive := range node.Directives {
		directive.Free()
	}
	node.SelectionSet.Free()
	node.LOC.Free()
	FragmentDefinitionPool.Put(node)
}

var FragmentDefinitionPool = &sync.Pool{New: func() interface{} { return &FragmentDefinition{} }}

func NewFragmentDefinition() *FragmentDefinition {
	node := FragmentDefinitionPool.Get().(*FragmentDefinition)
	node.ref = 1
	return node
}

type Literal struct {
	Type  string
	Value interface{}
	LOC   *LOC
	ref   int
}

func (node *Literal) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.LOC.Free()
	LiteralPool.Put(node)
}

var LiteralPool = &sync.Pool{New: func() interface{} { return &Literal{} }}

func NewLiteral() *Literal {
	node := LiteralPool.Get().(*Literal)
	node.ref = 1
	return node
}

type List struct {
	Values []ASTNode
	LOC    *LOC
	ref    int
}

func (node *List) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	for _, value := range node.Values {
		value.Free()
	}
	node.LOC.Free()
	ListPool.Put(node)
}

var ListPool = &sync.Pool{New: func() interface{} { return &List{} }}

func NewList() *List {
	node := ListPool.Get().(*List)
	node.ref = 1
	return node
}

type Object struct {
	Fields     []*ObjectField
	FieldIndex map[string]*ObjectField
	LOC        *LOC
	ref        int
}

func (node *Object) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	for _, field := range node.Fields {
		field.Free()
	}
	node.LOC.Free()
	ObjectPool.Put(node)
}

var ObjectPool = &sync.Pool{New: func() interface{} { return &Object{} }}

func NewObject() *Object {
	node := ObjectPool.Get().(*Object)
	node.ref = 1
	return node
}

type ObjectField struct {
	Name  *Name
	Value ASTNode
	LOC   *LOC
	ref   int
}

func (node *ObjectField) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	node.Value.Free()
	node.LOC.Free()
	ObjectFieldPool.Put(node)
}

var ObjectFieldPool = &sync.Pool{New: func() interface{} { return &ObjectField{} }}

func NewObjectField() *ObjectField {
	node := ObjectFieldPool.Get().(*ObjectField)
	node.ref = 1
	return node
}

type ListType struct {
	Type ASTNode
	LOC  *LOC
	ref  int
}

func (node *ListType) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Type.Free()
	node.LOC.Free()
	ListTypePool.Put(node)
}

var ListTypePool = &sync.Pool{New: func() interface{} { return &ListType{} }}

func NewListType() *ListType {
	node := ListTypePool.Get().(*ListType)
	node.ref = 1
	return node
}

type NonNullType struct {
	Type ASTNode
	LOC  *LOC
	ref  int
}

func (node *NonNullType) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Type.Free()
	node.LOC.Free()
	NonNullTypePool.Put(node)
}

var NonNullTypePool = &sync.Pool{New: func() interface{} { return &NonNullType{} }}

func NewNonNullType() *NonNullType {
	node := NonNullTypePool.Get().(*NonNullType)
	node.ref = 1
	return node
}

type Name struct {
	Value string
	LOC   *LOC
	ref   int
}

func (node *Name) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	if node.LOC != nil {
		node.LOC.Free()
	}
	NamePool.Put(node)
}

var NamePool = &sync.Pool{New: func() interface{} { return &Name{} }}

func NewName() *Name {
	node := NamePool.Get().(*Name)
	node.ref = 1
	return node
}

type NamedType struct {
	Name *Name
	LOC  *LOC
	ref  int
}

func (node *NamedType) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	node.LOC.Free()
	NamedTypePool.Put(node)
}

var NamedTypePool = &sync.Pool{New: func() interface{} { return &NamedType{} }}

func NewNamedType() *NamedType {
	node := NamedTypePool.Get().(*NamedType)
	node.ref = 1
	return node
}

type Directive struct {
	Name          *Name
	Arguments     []*Argument
	ArgumentIndex map[string]*Argument
	LOC           *LOC
	ref           int
}

func (node *Directive) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	for _, argument := range node.Arguments {
		argument.Free()
	}
	node.LOC.Free()
	DirectivePool.Put(node)
}

var DirectivePool = &sync.Pool{New: func() interface{} { return &Directive{} }}

func NewDirective() *Directive {
	node := DirectivePool.Get().(*Directive)
	node.ref = 1
	return node
}

type Argument struct {
	Name  *Name
	Value ASTNode
	LOC   *LOC
	ref   int
}

func (node *Argument) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	node.Value.Free()
	node.LOC.Free()
	ArgumentPool.Put(node)
}

var ArgumentPool = &sync.Pool{New: func() interface{} { return &Argument{} }}

func NewArgument() *Argument {
	node := ArgumentPool.Get().(*Argument)
	node.ref = 1
	return node
}

type Int struct {
	Value int32
	LOC   *LOC
	ref   int
}

func (node *Int) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.LOC.Free()
	IntPool.Put(node)
}

var IntPool = &sync.Pool{New: func() interface{} { return &Int{} }}

func NewInt() *Int {
	node := IntPool.Get().(*Int)
	node.ref = 1
	return node
}

type Float struct {
	Value float32
	LOC   *LOC
	ref   int
}

func (node *Float) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.LOC.Free()
	FloatPool.Put(node)
}

var FloatPool = &sync.Pool{New: func() interface{} { return &Float{} }}

func NewFloat() *Float {
	node := FloatPool.Get().(*Float)
	node.ref = 1
	return node
}

type String struct {
	Value string
	LOC   *LOC
	ref   int
}

func (node *String) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.LOC.Free()
	StringPool.Put(node)
}

var StringPool = &sync.Pool{New: func() interface{} { return &String{} }}

func NewString() *String {
	node := StringPool.Get().(*String)
	node.ref = 1
	return node
}

type Boolean struct {
	Value bool
	LOC   *LOC
	ref   int
}

func (node *Boolean) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.LOC.Free()
	BooleanPool.Put(node)
}

var BooleanPool = &sync.Pool{New: func() interface{} { return &Boolean{} }}

func NewBoolean() *Boolean {
	node := BooleanPool.Get().(*Boolean)
	node.ref = 1
	return node
}

type Enum struct {
	Value string
	LOC   *LOC
	ref   int
}

func (node *Enum) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.LOC.Free()
	EnumPool.Put(node)
}

var EnumPool = &sync.Pool{New: func() interface{} { return &Enum{} }}

func NewEnum() *Enum {
	node := EnumPool.Get().(*Enum)
	node.ref = 1
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

func (node *ObjectTypeDefinition) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	for _, interfac := range node.Interfaces {
		interfac.Free()
	}
	for _, field := range node.Fields {
		field.Free()
	}
	node.LOC.Free()
	ObjectTypeDefinitionPool.Put(node)
}

var ObjectTypeDefinitionPool = &sync.Pool{New: func() interface{} { return &ObjectTypeDefinition{} }}

func NewObjectTypeDefinition() *ObjectTypeDefinition {
	node := ObjectTypeDefinitionPool.Get().(*ObjectTypeDefinition)
	node.ref = 1
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

func (node *FieldDefinition) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	for _, argument := range node.Arguments {
		argument.Free()
	}
	node.Type.Free()
	node.LOC.Free()
	FieldDefinitionPool.Put(node)
}

var FieldDefinitionPool = &sync.Pool{New: func() interface{} { return &FieldDefinition{} }}

func NewFieldDefinition() *FieldDefinition {
	node := FieldDefinitionPool.Get().(*FieldDefinition)
	node.ref = 1
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

func (node *InputValueDefinition) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	node.Type.Free()
	if node.DefaultValue != nil {
		node.DefaultValue.Free()
	}
	node.LOC.Free()
	InputValueDefinitionPool.Put(node)
}

var InputValueDefinitionPool = &sync.Pool{New: func() interface{} { return &InputValueDefinition{} }}

func NewInputValueDefinition() *InputValueDefinition {
	node := InputValueDefinitionPool.Get().(*InputValueDefinition)
	node.ref = 1
	return node
}

type InterfaceTypeDefinition struct {
	Name        *Name
	Description string
	Fields      []*FieldDefinition
	LOC         *LOC
	ref         int
}

func (node *InterfaceTypeDefinition) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	for _, field := range node.Fields {
		field.Free()
	}
	node.LOC.Free()
	InterfaceTypeDefinitionPool.Put(node)
}

var InterfaceTypeDefinitionPool = &sync.Pool{New: func() interface{} { return &InterfaceTypeDefinition{} }}

func NewInterfaceTypeDefinition() *InterfaceTypeDefinition {
	node := InterfaceTypeDefinitionPool.Get().(*InterfaceTypeDefinition)
	node.ref = 1
	return node
}

type UnionTypeDefinition struct {
	Name        *Name
	Description string
	Types       []*NamedType
	LOC         *LOC
	ref         int
}

func (node *UnionTypeDefinition) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	for _, typ := range node.Types {
		typ.Free()
	}
	node.LOC.Free()
	UnionTypeDefinitionPool.Put(node)
}

var UnionTypeDefinitionPool = &sync.Pool{New: func() interface{} { return &UnionTypeDefinition{} }}

func NewUnionTypeDefinition() *UnionTypeDefinition {
	node := UnionTypeDefinitionPool.Get().(*UnionTypeDefinition)
	node.ref = 1
	return node
}

type ScalarTypeDefinition struct {
	Name        *Name
	Description string
	LOC         *LOC
	ref         int
}

func (node *ScalarTypeDefinition) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	node.LOC.Free()
	ScalarTypeDefinitionPool.Put(node)
}

var ScalarTypeDefinitionPool = &sync.Pool{New: func() interface{} { return &ScalarTypeDefinition{} }}

func NewScalarTypeDefinition() *ScalarTypeDefinition {
	node := ScalarTypeDefinitionPool.Get().(*ScalarTypeDefinition)
	node.ref = 1
	return node
}

type EnumTypeDefinition struct {
	Name        *Name
	Description string
	Values      []*EnumValueDefinition
	LOC         *LOC
	ref         int
}

func (node *EnumTypeDefinition) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	for _, value := range node.Values {
		value.Free()
	}
	node.LOC.Free()
	EnumTypeDefinitionPool.Put(node)
}

var EnumTypeDefinitionPool = &sync.Pool{New: func() interface{} { return &EnumTypeDefinition{} }}

func NewEnumTypeDefinition() *EnumTypeDefinition {
	node := EnumTypeDefinitionPool.Get().(*EnumTypeDefinition)
	node.ref = 1
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

func (node *EnumValueDefinition) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	node.LOC.Free()
	EnumTypeDefinitionPool.Put(node)
}

var EnumValueDefinitionPool = &sync.Pool{New: func() interface{} { return &EnumValueDefinition{} }}

func NewEnumValueDefinition() *EnumValueDefinition {
	node := EnumValueDefinitionPool.Get().(*EnumValueDefinition)
	node.ref = 1
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

func (node *InputObjectTypeDefinition) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Name.Free()
	for _, field := range node.Fields {
		field.Free()
	}
	node.LOC.Free()
	InputObjectTypeDefinitionPool.Put(node)
}

var InputObjectTypeDefinitionPool = &sync.Pool{New: func() interface{} { return &InputObjectTypeDefinition{} }}

func NewInputObjectTypeDefinition() *InputObjectTypeDefinition {
	node := InputObjectTypeDefinitionPool.Get().(*InputObjectTypeDefinition)
	node.ref = 1
	return node
}

type TypeExtensionDefinition struct {
	Description string
	Definition  *ObjectTypeDefinition
	LOC         *LOC
	ref         int
}

func (node *TypeExtensionDefinition) Free() {
	if node.ref > 0 {
		node.ref--
	} else {
		node.ref--
		log.Printf("%#v\n freed multiple times (%d extra times)", node, -node.ref)
		return
	}
	node.Definition.Free()
	node.LOC.Free()
	TypeExtensionDefinitionPool.Put(node)
}

var TypeExtensionDefinitionPool = &sync.Pool{New: func() interface{} { return &TypeExtensionDefinition{} }}

func NewTypeExtensionDefinition() *TypeExtensionDefinition {
	node := TypeExtensionDefinitionPool.Get().(*TypeExtensionDefinition)
	node.ref = 1
	return node
}
