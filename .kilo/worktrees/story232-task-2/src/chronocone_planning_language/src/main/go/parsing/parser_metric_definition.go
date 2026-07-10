// Package parsing provides data structures for the CPL parsing phase.
//
// This package contains parser-specific data structures that are stored
// in the symbol table during parsing and later retrieved by the analysis
// phase to build the final AST structures.
package parsing

import "github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"

// ParserMetricDefinition is stored in SymbolDefinition.Data during parsing.
// The analysis phase retrieves this data to build AnalysisMetricDefinition.
type ParserMetricDefinition struct {
	// Name is the metric identifier.
	Name string

	// Inputs contains the input declarations for this metric.
	Inputs []ParserInputDefinition

	// CalculatePostfix holds the postfix expression sequence for the calculate clause.
	// Each element is a typed PostfixElement: PostfixIdentifier, PostfixLiteral, or PostfixOperator.
	// For Increment 1, this contains a single literal 0.
	CalculatePostfix []PostfixElement

	// ShortForm is true if this metric was declared using short form syntax.
	// Short form: "metric_name = expression" (deferred to later increment).
	// Defaults to false (Go zero value).
	ShortForm bool

	// SourceLocation records where this metric was declared in the source.
	// Used for error reporting in subsequent phases.
	SourceLocation symboltable.SourceLocation
}

// IsDefinitionData marks ParserMetricDefinition as implementing symboltable.DefinitionData.
func (ParserMetricDefinition) IsDefinitionData() {}
