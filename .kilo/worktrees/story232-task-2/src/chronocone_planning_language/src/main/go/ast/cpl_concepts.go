// Package ast provides the Abstract Syntax Tree types for the Chronocone Planning Language.
//
// This file contains AST types related to the CPLConcepts.g4 grammar file,
// which defines the concepts section and concept definitions.
package ast

// GeneratorType identifies the type of concept value generator.
// Different generator types produce values in different ways (e.g., explicit list, range, etc.).
type GeneratorType int

const (
	// ValuesGenerator represents Values[Value1, Value2, ...] - an explicit list of values.
	ValuesGenerator GeneratorType = iota
	// IntRangeGenerator represents Int[Start, End, Step] - a range of integer values.
	IntRangeGenerator
	// ConceptRangeGenerator represents Concept[StartValue..EndValue] - a range of values from another concept.
	ConceptRangeGenerator
)

// ValuesGeneratorDef represents Values[Value1, Value2, ...] in the AST.
// This corresponds to the 'valuesPatternGenerator' rule in CPLConcepts.g4.
type ValuesGeneratorDef struct {
	Values []string
}

// IntRangeGeneratorDef represents Int[Start, End, Step] in the AST.
// This corresponds to the 'intRangeGenerator' rule in CPLConcepts.g4.
// Step defaults to 1 if not specified in source.
type IntRangeGeneratorDef struct {
	Start int
	End   int
	Step  int
}

// ConceptRangeGeneratorDef represents Concept[StartValue..EndValue] in the AST.
// This is a data structure that captures the parsed syntax only.
// Value resolution happens during analysis phase after all concepts are defined.
type ConceptRangeGeneratorDef struct {
	ConceptName string // Name of the source concept to reference
	StartValue  string // Start value (inclusive)
	EndValue    string // End value (inclusive)
}

// ConditionalSpec represents a conditional concept value specification.
// It defines child concept values that apply only when a parent concept
// has a specific value. Example: Country[USA] => State=[Values[NY, CA, TX]]
type ConditionalSpec struct {
	ParentValue       string        // Value of parent concept (e.g., "USA")
	GeneratorType     GeneratorType // Type of generator (ValuesGenerator, IntRangeGenerator, etc.)
	Values            []string      // Generated values for this condition
	RangeBounds       *RangeBounds  // For Int/Concept range generators: bounds for value expansion
	SourceConceptName string        // For ConceptRangeGenerator: name of the concept to reference
}

// RangeBounds captures the bounds of a range generator for conditional specs.
type RangeBounds struct {
	Start string // Start value
	End   string // End value
	Step  int    // Step (for Int ranges, defaults to 1)
}

// ConceptDefinitionData holds concept-related data stored in SymbolDefinition.Data
// during parsing. The analysis phase retrieves this data to build AnalysisConceptDefinition.
type ConceptDefinitionData struct {
	// Name is the concept identifier (e.g., "Country", "Priority").
	Name string
	// InsertionPoint specifies where new values should be inserted relative to
	// existing values for concept extensions. Zero value means no insertion point
	// (this is an initial definition, not an extension).
	InsertionPoint InsertionPoint
	// ConditionalSpecs holds conditional value specifications for this concept.
	ConditionalSpecs []ConditionalSpec
	// DefaultValues holds unconditional (default) values for this concept.
	DefaultValues []string
	// SourceLocation is where the concept was defined in source.
	SourceLocation SourceLocation
}

// InsertionPoint stores the before/after positioning information for concept extensions.
// It specifies where new values should be inserted relative to existing values.
// At least one of BeforeValue or AfterValue must be set.
type InsertionPoint struct {
	BeforeValue string // If set, insert new values before this existing value
	AfterValue  string // If set, insert new values after this existing value
}

// SourceLocation captures where a symbol is defined in source.
// This is used for error reporting and debugging.
type SourceLocation struct {
	Line   int
	Column int
}
