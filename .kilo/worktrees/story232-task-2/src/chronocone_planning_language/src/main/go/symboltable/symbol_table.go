// Package symboltable provides symbol table data structures for the CPL parser.
//
// The symbol table is used during parsing to track names and their definitions
// across different scopes. It maintains a hierarchical tree structure that
// supports nested scopes and parent-child relationships between symbol tables.
//
// The symbol table follows the design documented in:
//   - memory-bank/specifications/design/cpl/go-target/symbol-table-design.md
//
// Error handling follows Go conventions with operations returning (result, error)
// pairs. Errors are of type *SymbolTableError and can be inspected using
// errors.As() to check the specific error code.
package symboltable

import "fmt"

// SpacesPerTab is the number of spaces that represent one tab character.
// This value is intentionally fixed at 4 because:
//   - The CPL grammar uses consistent 4-space indentation for all scope levels
//   - Tab characters in source files are expanded to exactly 4 spaces during parsing
//   - No configuration mechanism exists or is needed for this invariant
const SpacesPerTab = 4

// ErrorCode identifies the type of symbol table error.
type ErrorCode int

const (
	// ErrDuplicateDefinition indicates a named symbol is already defined in the current scope.
	ErrDuplicateDefinition ErrorCode = iota
	// ErrSymbolNotFound indicates a symbol lookup failed to find the requested name.
	ErrSymbolNotFound
	// ErrUnbalancedStack indicates the parser ended with an unbalanced symbol table stack.
	ErrUnbalancedStack
	// ErrEmptyStack indicates an operation was attempted on an empty stack.
	ErrEmptyStack
)

// ScopeType enumerates the different kinds of scopes in CPL.
type ScopeType int

const (
	// GlobalScope is the root scope of the CPL document.
	GlobalScope ScopeType = iota

	// ConceptsScope is the scope containing concept definitions.
	ConceptsScope

	// ConceptScope is the scope for a single concept definition.
	ConceptScope

	// ConditionalSpecScope is the scope for a conditional concept specification.
	// It wraps a child concept line that defines values conditional on a parent
	// concept value (e.g., Country[USA] => State=[Values[NY, CA, TX]]).
	ConditionalSpecScope

	// ConstraintScope is the scope for a single constraint definition.
	ConstraintScope

	// MeasuresScope is the scope containing measure declarations.
	MeasuresScope

	// MeasureScope is the scope for a single measure definition.
	MeasureScope

	// TypesScope is the scope containing type declarations.
	TypesScope

	// TypeScope is the scope for a single type definition.
	TypeScope

	// UnitsScope is the scope containing unit declarations.
	UnitsScope

	// UnitScope is the scope for a single unit definition.
	UnitScope

	// The ValueGeneratorScope is the scope for a single value generator definition.
	ValueGeneratorScope

	// IntGeneratorScope is the scope for a single integer range generator definition.
	IntGeneratorScope

	// ConceptGeneratorScope is the scope for a single concept range generator definition.
	ConceptGeneratorScope

	// ConnectorsScope is the scope containing connector definitions.
	ConnectorsScope

	// ConnectorScope is the scope for a single connector definition.
	ConnectorScope

	// AggregatorsScope is the scope containing aggregator definitions.
	AggregatorsScope

	// AggregatorScope is the scope for a single aggregator definition.
	AggregatorScope

	// SpreadersScope is the scope containing spreader definitions.
	SpreadersScope

	// SpreaderScope is the scope for a single spreader definition.
	SpreaderScope

	// LookupsScope is the scope containing lookup definitions.
	LookupsScope

	// MetricsScope is the scope containing metric definitions.
	MetricsScope

	// MetricScope is the scope for a single metric definition.
	MetricScope

	// TempScope is a temporary scope used to capture expressions during parsing.
	TempScope

	// IfBodyScope is the scope for the body of an if-branch.
	IfBodyScope

	// ElseIfBodyScope is the scope for the body of an else-if branch.
	ElseIfBodyScope

	// ElseBodyScope is the scope for the body of an else branch.
	ElseBodyScope

	// WhenBodyScope is the scope for the body of a when clause.
	WhenBodyScope

	// DefaultBodyScope is the scope for the body of a default clause.
	DefaultBodyScope

	// ConversionScope is the scope for a conversion clause within a unit declaration.
	ConversionScope

	// PatternScope is the scope for a pattern clause within a conversion.
	PatternScope

	// CalculationScope is the scope for a calculation clause within a conversion.
	CalculationScope

	// MeasureConversionScope is the scope for a conversion clause within a measure declaration.
	MeasureConversionScope
)

// String returns the string representation of the ScopeType.
// This method implements the fmt.Stringer interface.
func (s ScopeType) String() string {
	switch s {
	case GlobalScope:
		return "GlobalScope"
	case UnitsScope:
		return "UnitsScope"
	case UnitScope:
		return "UnitScope"
	case MeasuresScope:
		return "MeasuresScope"
	case MeasureScope:
		return "MeasureScope"
	case ConceptsScope:
		return "ConceptsScope"
	case ConceptScope:
		return "ConceptScope"
	case ConditionalSpecScope:
		return "ConditionalSpecScope"
	case ValueGeneratorScope:
		return "ValueGeneratorScope"
	case IntGeneratorScope:
		return "IntGeneratorScope"
	case ConceptGeneratorScope:
		return "ConceptGeneratorScope"
	case TypesScope:
		return "TypesScope"
	case TypeScope:
		return "TypeScope"
	case ConnectorsScope:
		return "ConnectorsScope"
	case ConnectorScope:
		return "ConnectorScope"
	case AggregatorsScope:
		return "AggregatorsScope"
	case AggregatorScope:
		return "AggregatorScope"
	case SpreadersScope:
		return "SpreadersScope"
	case SpreaderScope:
		return "SpreaderScope"
	case LookupsScope:
		return "LookupsScope"
	case MetricsScope:
		return "MetricsScope"
	case MetricScope:
		return "MetricScope"
	case TempScope:
		return "TempScope"
	case IfBodyScope:
		return "IfBodyScope"
	case ElseIfBodyScope:
		return "ElseIfBodyScope"
	case ElseBodyScope:
		return "ElseBodyScope"
	case WhenBodyScope:
		return "WhenBodyScope"
	case DefaultBodyScope:
		return "DefaultBodyScope"
	case ConversionScope:
		return "ConversionScope"
	case PatternScope:
		return "PatternScope"
	case CalculationScope:
		return "CalculationScope"
	case MeasureConversionScope:
		return "MeasureConversionScope"
	default:
		return fmt.Sprintf("ScopeType(%d)", s)
	}
}

// DefinitionType enumerates the kinds of definitions that can appear
// in a symbol table.
type DefinitionType int

const (
	// UnitDefinition represents a unit declaration.
	UnitDefinition DefinitionType = iota

	// UnitScopeDefinition represents a unit scope definition.
	UnitScopeDefinition

	// TypeDefinition represents a type declaration.
	TypeDefinition

	// Identifier represents an identifier
	Identifier

	// ConnectorsScopeDefinition represents a connectors section scope definition.
	ConnectorsScopeDefinition

	// ConnectorScopeDefinition represents an individual connector scope definition.
	ConnectorScopeDefinition

	// AggregatorsScopeDefinition represents an aggregators section scope definition.
	AggregatorsScopeDefinition

	// AggregatorScopeDefinition represents an individual aggregator scope definition.
	AggregatorScopeDefinition

	// SpreadersScopeDefinition represents a spreaders section scope definition.
	SpreadersScopeDefinition

	// SpreaderScopeDefinition represents an individual spreader scope definition.
	SpreaderScopeDefinition

	// GrainDefinition represents a grain declaration within a connector.
	GrainDefinition

	// MetricsScopeDefinition represents a metrics section scope definition.
	MetricsScopeDefinition

	// MetricScopeDefinition represents an individual metric scope definition.
	MetricScopeDefinition

	// LookupScopeDefinition represents a lookups section scope definition.
	LookupScopeDefinition

	// InputRefDefinition represents an input reference declaration within a metric.
	InputRefDefinition

	// LiteralDefinition represents a literal value in a calculate expression.
	LiteralDefinition

	// PlusOperatorDefinition represents the + operator in postfix expressions.
	PlusOperatorDefinition

	// MinusOperatorDefinition represents the - operator in postfix expressions.
	MinusOperatorDefinition

	// MultiplyOperatorDefinition represents the * operator in postfix expressions.
	MultiplyOperatorDefinition

	// DivideOperatorDefinition represents the / operator in postfix expressions.
	DivideOperatorDefinition

	// NegateOperatorDefinition represents unary negation (-) in postfix expressions.
	NegateOperatorDefinition

	// UnitReferenceDefinition represents a unit name reference in an arithmetic expression.
	UnitReferenceDefinition

	// PowerOperatorDefinition represents the ^ operator for exponentiation.
	PowerOperatorDefinition

	// NumberDefinition represents a numeric literal in an arithmetic expression.
	NumberDefinition

	// --- Clause definition types (for conditional branches) ---

	// IfClauseDefinition carries a condition as ClauseData.
	IfClauseDefinition
	// ElseIfClauseDefinition carries a condition as ClauseData.
	ElseIfClauseDefinition
	// ElseClauseDefinition has no condition.
	ElseClauseDefinition
	// WhenClauseDefinition carries a condition as ClauseData.
	WhenClauseDefinition
	// DefaultClauseDefinition has no condition.
	DefaultClauseDefinition

	// --- Comparison operators ---

	// EqOperatorDefinition represents the == operator.
	EqOperatorDefinition
	// NotEqOperatorDefinition represents the != operator.
	NotEqOperatorDefinition
	// LessThanOperatorDefinition represents the < operator.
	LessThanOperatorDefinition
	// GreaterThanOperatorDefinition represents the > operator.
	GreaterThanOperatorDefinition
	// LessThanEqOperatorDefinition represents the <= operator.
	LessThanEqOperatorDefinition
	// GreaterThanEqOperatorDefinition represents the >= operator.
	GreaterThanEqOperatorDefinition

	// --- Logical operators ---

	// AndOperatorDefinition represents the and operator.
	AndOperatorDefinition
	// OrOperatorDefinition represents the or operator.
	OrOperatorDefinition
	// NotOperatorDefinition represents the not operator.
	NotOperatorDefinition

	// --- Literal types ---

	// NilLiteralDefinition represents the Nil literal.
	NilLiteralDefinition
	// TrueLiteralDefinition represents the True literal.
	TrueLiteralDefinition
	// FalseLiteralDefinition represents the False literal.
	FalseLiteralDefinition

	// ConversionDefinition represents a conversion clause definition.
	ConversionDefinition

	// PatternDefinition represents a pattern clause definition.
	PatternDefinition

	// CalculationDefinition represents a calculation clause definition.
	CalculationDefinition

	// MeasureConversionDefinition represents a measure conversion clause definition.
	MeasureConversionDefinition
)

// String returns the string representation of the DefinitionType.
// This method implements the fmt.Stringer interface.
func (d DefinitionType) String() string {
	switch d {
	case UnitDefinition:
		return "UnitDefinition"
	case UnitScopeDefinition:
		return "UnitScopeDefinition"
	case TypeDefinition:
		return "TypeDefinition"
	case Identifier:
		return "Identifier"
	case ConnectorsScopeDefinition:
		return "ConnectorsScopeDefinition"
	case ConnectorScopeDefinition:
		return "ConnectorScopeDefinition"
	case AggregatorsScopeDefinition:
		return "AggregatorsScopeDefinition"
	case AggregatorScopeDefinition:
		return "AggregatorScopeDefinition"
	case SpreadersScopeDefinition:
		return "SpreadersScopeDefinition"
	case SpreaderScopeDefinition:
		return "SpreaderScopeDefinition"
	case GrainDefinition:
		return "GrainDefinition"
	case MetricsScopeDefinition:
		return "MetricsScopeDefinition"
	case MetricScopeDefinition:
		return "MetricScopeDefinition"
	case LookupScopeDefinition:
		return "LookupScopeDefinition"
	case InputRefDefinition:
		return "InputRefDefinition"
	case LiteralDefinition:
		return "LiteralDefinition"
	case PlusOperatorDefinition:
		return "PlusOperatorDefinition"
	case MinusOperatorDefinition:
		return "MinusOperatorDefinition"
	case MultiplyOperatorDefinition:
		return "MultiplyOperatorDefinition"
	case DivideOperatorDefinition:
		return "DivideOperatorDefinition"
	case NegateOperatorDefinition:
		return "NegateOperatorDefinition"
	case UnitReferenceDefinition:
		return "UnitReferenceDefinition"
	case PowerOperatorDefinition:
		return "PowerOperatorDefinition"
	case NumberDefinition:
		return "NumberDefinition"

	// Clause definition types
	case IfClauseDefinition:
		return "IfClauseDefinition"
	case ElseIfClauseDefinition:
		return "ElseIfClauseDefinition"
	case ElseClauseDefinition:
		return "ElseClauseDefinition"
	case WhenClauseDefinition:
		return "WhenClauseDefinition"
	case DefaultClauseDefinition:
		return "DefaultClauseDefinition"

	// Comparison operators
	case EqOperatorDefinition:
		return "EqOperatorDefinition"
	case NotEqOperatorDefinition:
		return "NotEqOperatorDefinition"
	case LessThanOperatorDefinition:
		return "LessThanOperatorDefinition"
	case GreaterThanOperatorDefinition:
		return "GreaterThanOperatorDefinition"
	case LessThanEqOperatorDefinition:
		return "LessThanEqOperatorDefinition"
	case GreaterThanEqOperatorDefinition:
		return "GreaterThanEqOperatorDefinition"

	// Logical operators
	case AndOperatorDefinition:
		return "AndOperatorDefinition"
	case OrOperatorDefinition:
		return "OrOperatorDefinition"
	case NotOperatorDefinition:
		return "NotOperatorDefinition"

	// Literal types
	case NilLiteralDefinition:
		return "NilLiteralDefinition"
	case TrueLiteralDefinition:
		return "TrueLiteralDefinition"
	case FalseLiteralDefinition:
		return "FalseLiteralDefinition"

	// Conversion types
	case ConversionDefinition:
		return "ConversionDefinition"
	case PatternDefinition:
		return "PatternDefinition"
	case CalculationDefinition:
		return "CalculationDefinition"
	case MeasureConversionDefinition:
		return "MeasureConversionDefinition"

	default:
		return fmt.Sprintf("DefinitionType(%d)", d)
	}
}

// SourceLocation captures where a symbol is defined in source.
type SourceLocation struct {
	// File is the path to the source file where the symbol is defined.
	File string
	// Line is the 1-based line number in the source file.
	Line int
	// Column is the 0-based column number in the source file.
	Column int
}

// ConnectorDeclaration represents one connector declaration extracted during parsing.
// Parser listeners attach this record to the individual ConnectorScope.Data field so
// later phases can read connector roles from explicit fields instead of inferring
// roles from identifier position in the ConnectorScope.Definitions sequence.
type ConnectorDeclaration struct {
	// Name is the connector identifier.
	Name string
	// NameLocation is the source location where the connector name appears.
	NameLocation SourceLocation
	// DimensionNames is the ordered list of dimension concept names declared by the connector.
	DimensionNames []string
	// DimensionLocations contains the source location for each dimension in DimensionNames.
	DimensionLocations []SourceLocation
	// GrainNames is the ordered list of grain measure names declared by the connector.
	// Each grain references an Interval-derived measure.
	GrainNames []string
	// GrainLocations contains the source location for each grain in GrainNames.
	GrainLocations []SourceLocation
	// CellTypeName is the celltype measure name declared by the connector.
	CellTypeName string
	// CellTypeLocation is the source location where the celltype measure name appears.
	CellTypeLocation SourceLocation
}

// SymbolDefinition captures a name and its definition details.
// Name is a pointer to allow for optional names (nil if no name).
type SymbolDefinition struct {
	// Name is the identifier name, or nil for anonymous definitions.
	Name *string
	// DefinitionType indicates the kind of definition (unit, connector, metric, etc.).
	DefinitionType DefinitionType
	// Location is where the symbol was defined in the source.
	Location SourceLocation
	// Data allows attaching type-specific information during analysis.
	// Use DefinitionData for type-safe extensibility.

	Data DefinitionData
}

// ClauseData carries the condition's postfix definitions for a conditional clause.
// Stored as Data on IfClauseDefinition, ElseIfClauseDefinition, or WhenClauseDefinition.
type ClauseData struct {
	ConditionPostfix []SymbolDefinition // Postfix for the condition expression
	// BodyScope stores a reference to the branch body's child scope so the
	// analysis conditional assembly algorithm can locate the body expression's
	// postfix definitions.
	BodyScope *SymbolTable
}

// IsDefinitionData marks ClauseData as a valid DefinitionData carrier.
func (ClauseData) IsDefinitionData() {}

// SymbolTableError represents an error from symbol table operations.
// It implements the standard Go error interface.
type SymbolTableError struct {
	// Code identifies the type of error.
	Code ErrorCode
	// Message provides a human-readable description of the error.
	Message string
	// Name is the symbol name involved, if applicable.
	Name string
}

// Error returns the error message, implementing the error interface.
func (e *SymbolTableError) Error() string {
	return e.Message
}

// SymbolTable captures a set of names and their corresponding definitions.
// It maintains parent-child relationships to form a hierarchical tree structure.
type SymbolTable struct {
	// ScopeType indicates the kind of scope this table defines.
	ScopeType ScopeType
	// Parent is a pointer to the parent symbol table (nil for root/global).
	Parent *SymbolTable
	// Children is a list of pointers to child symbol tables.
	Children []*SymbolTable
	// Definitions is a list of symbol definitions in this scope.
	Definitions []SymbolDefinition

	// The level of indentation for this scope.
	IndentedSpaceEquivalent int

	// Data allows attaching type-specific information to a scope during parsing.
	// Use interface{} for future extensibility.
	Data interface{}
}

// NewSymbolTable creates a new symbol table with the given scope type.
// The returned table has no parent and no children.
func NewSymbolTable(scopeType ScopeType) *SymbolTable {
	return &SymbolTable{
		ScopeType:   scopeType,
		Parent:      nil,
		Children:    make([]*SymbolTable, 0),
		Definitions: make([]SymbolDefinition, 0),
	}
}

// NewGlobalSymbolTable creates a new global (root) symbol table.
// This is a convenience constructor equivalent to NewSymbolTable(GlobalScope).
func NewGlobalSymbolTable() *SymbolTable {
	return NewSymbolTable(GlobalScope)
}

// AddChild creates a new symbol table as a child of this one.
// The new table is added to this table's Children list and has its
// Parent pointer set to this table.
func (st *SymbolTable) AddChild(scopeType ScopeType) *SymbolTable {
	child := &SymbolTable{
		ScopeType:   scopeType,
		Parent:      st,
		Children:    make([]*SymbolTable, 0),
		Definitions: make([]SymbolDefinition, 0),
	}
	st.Children = append(st.Children, child)
	return child
}

// Define adds a new symbol definition to this table.
// Returns an error if the name is already defined in this scope.
// If name is nil, the definition is added without a name check.
func (st *SymbolTable) Define(name *string, defType DefinitionType, location SourceLocation) error {
	return st.DefineWithData(name, defType, location, nil)
}

// DefineWithData adds a new symbol definition with attached data to this table.
// If name is nil, the definition is added without a name check.
// The data parameter allows attaching type-specific information
// that will be used during the analysis phase.
//
// Note: Duplicate definition detection is handled in the analysis phase, not here.
// Each analysis function (e.g., AnalyzeMeasureScope, AnalyzeUnitScope) checks for
// duplicates across the appropriate definitions map.
func (st *SymbolTable) DefineWithData(name *string, defType DefinitionType, location SourceLocation, data DefinitionData) error {
	// Add the definition
	st.Definitions = append(st.Definitions, SymbolDefinition{
		Name:           name,
		DefinitionType: defType,
		Location:       location,
		Data:           data,
	})

	return nil
}

// Lookup searches for a symbol definition by name starting from this table
// and walking up through parent tables.
// Returns (*SymbolDefinition, nil) if found, or (nil, error) if not found.
// The error will have code ErrSymbolNotFound.
func (st *SymbolTable) Lookup(name string) (*SymbolDefinition, error) {
	// First try local lookup
	def, err := st.LookupLocal(name)
	if err == nil {
		return def, nil
	}

	// If not found locally and we have a parent, try parent
	if st.Parent != nil {
		return st.Parent.Lookup(name)
	}

	// Not found anywhere
	return nil, &SymbolTableError{
		Code:    ErrSymbolNotFound,
		Message: fmt.Sprintf("symbol '%s' not found", name),
		Name:    name,
	}
}

// LookupLocal searches for a symbol definition by name in this table only.
// Returns (*SymbolDefinition, nil) if found, or (nil, error) if not found.
// The error will have code ErrSymbolNotFound.
func (st *SymbolTable) LookupLocal(name string) (*SymbolDefinition, error) {
	for i := range st.Definitions {
		if st.Definitions[i].Name != nil && *st.Definitions[i].Name == name {
			return &st.Definitions[i], nil
		}
	}

	return nil, &SymbolTableError{
		Code:    ErrSymbolNotFound,
		Message: fmt.Sprintf("symbol '%s' not found in current scope", name),
		Name:    name,
	}
}

// WalkSymbolTable recursively walks the symbol table tree,
// calling the provided function for each table.
//
// The walk is performed in pre-order (parent before children).
// The function is called for each table in the tree, starting from
// the provided root table and traversing all descendants.
//
// Example usage:
//
//	WalkSymbolTable(rootTable, func(table *SymbolTable) {
//	    fmt.Printf("Visiting scope: %s\n", table.ScopeType)
//	    for _, def := range table.Definitions {
//	        if def.Name != nil {
//	            fmt.Printf("  - %s\n", *def.Name)
//	        }
//	    }
//	})
func WalkSymbolTable(table *SymbolTable, fn func(*SymbolTable)) {
	if table == nil || fn == nil {
		return
	}

	// Visit the current table first (pre-order)
	fn(table)

	// Recursively visit all children
	for _, child := range table.Children {
		WalkSymbolTable(child, fn)
	}
}
