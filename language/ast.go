package language

type ASTNode interface{}

type Position struct {
	Index  int
	Line   int
	Column int
}

type LOC struct {
	Start  *Position
	End    *Position
	Source string
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
}

type SelectionSet struct {
	Selections []ASTNode
	LOC        *LOC
}

func (ss *SelectionSet) SelectionNames(doc *Document, whitelist []string) []string {
	result := []string{}
	whitelistMap := map[string]bool{}
	if whitelist != nil {
		for _, name := range whitelist {
			whitelistMap[name] = true
		}
	} else {
		whitelistMap = nil
	}
	names := ss.selectionNames(doc, whitelistMap)
	for name, _ := range names {
		result = append(result, name)
	}
	return result
}

func (ss *SelectionSet) selectionNames(doc *Document, whitelistMap map[string]bool) map[string]bool {
	result := map[string]bool{}
	for _, selection := range ss.Selections {
		switch field := selection.(type) {
		case *InlineFragment:
			names := field.SelectionSet.selectionNames(doc, whitelistMap)
			for name, _ := range names {
				if _, ok := result[name]; !ok {
					result[name] = true
				}
			}
		case *FragmentSpread:
			fragment := doc.FragmentIndex[field.Name.Value]
			if fragment != nil {
				names := fragment.SelectionSet.selectionNames(doc, whitelistMap)
				for name, _ := range names {
					if _, ok := result[name]; !ok {
						result[name] = true
					}
				}
			}
		case *Field:
			if whitelistMap == nil {
				result[field.Name.Value] = true
			} else if _, ok := whitelistMap[field.Name.Value]; ok {
				result[field.Name.Value] = true
			}
		}
	}
	return result
}

type VariableDefinition struct {
	Variable     *Variable
	Type         ASTNode
	DefaultValue interface{}
	LOC          *LOC
}

type Variable struct {
	Name *Name
	LOC  *LOC
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
}

type InlineFragment struct {
	TypeCondition  *NamedType
	Directives     []*Directive
	DirectiveIndex map[string]*Directive
	SelectionSet   *SelectionSet
	LOC            *LOC
}

type FragmentSpread struct {
	Name           *Name
	Directives     []*Directive
	DirectiveIndex map[string]*Directive
	LOC            *LOC
}

type FragmentDefinition struct {
	Name           *Name
	TypeCondition  *NamedType
	Directives     []*Directive
	DirectiveIndex map[string]*Directive
	SelectionSet   *SelectionSet
	LOC            *LOC
}

type Literal struct {
	Type  string
	Value interface{}
	LOC   *LOC
}

type List struct {
	Values []ASTNode
	LOC    *LOC
}

type Object struct {
	Fields []*ObjectField
	LOC    *LOC
}

type ObjectField struct {
	Name  *Name
	Value ASTNode
	LOC   *LOC
}

type ListType struct {
	Type ASTNode
	LOC  *LOC
}

type NonNullType struct {
	Type ASTNode
	LOC  *LOC
}

type Name struct {
	Value string
	LOC   *LOC
}

type NamedType struct {
	Name *Name
	LOC  *LOC
}

type Directive struct {
	Name          *Name
	Arguments     []*Argument
	ArgumentIndex map[string]*Argument
	LOC           *LOC
}

type Argument struct {
	Name  *Name
	Value ASTNode
	LOC   *LOC
}

type Int struct {
	Value int32
	LOC   *LOC
}

type Float struct {
	Value float32
	LOC   *LOC
}

type String struct {
	Value string
	LOC   *LOC
}

type Boolean struct {
	Value bool
	LOC   *LOC
}

type Enum struct {
	Value string
	LOC   *LOC
}

type ObjectTypeDefinition struct {
	Name        *Name
	Description string
	Interfaces  []*NamedType
	Fields      []*FieldDefinition
	FieldIndex  map[string]*FieldDefinition
	LOC         *LOC
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
}

type InputValueDefinition struct {
	Name         *Name
	Description  string
	Type         ASTNode
	DefaultValue ASTNode
	LOC          *LOC
}

type InterfaceTypeDefinition struct {
	Name        *Name
	Description string
	Fields      []*FieldDefinition
	LOC         *LOC
}

type UnionTypeDefinition struct {
	Name        *Name
	Description string
	Types       []*NamedType
	LOC         *LOC
}

type ScalarTypeDefinition struct {
	Name        *Name
	Description string
	LOC         *LOC
}

type EnumTypeDefinition struct {
	Name        *Name
	Description string
	Values      []*EnumValueDefinition
	LOC         *LOC
}

type EnumValueDefinition struct {
	Name              *Name
	Description       string
	IsDeprecated      bool
	DeprecationReason string
	LOC               *LOC
}

type InputObjectTypeDefinition struct {
	Name        *Name
	Description string
	Fields      []*InputValueDefinition
	LOC         *LOC
}

type TypeExtensionDefinition struct {
	Description string
	Definition  *ObjectTypeDefinition
	LOC         *LOC
}
