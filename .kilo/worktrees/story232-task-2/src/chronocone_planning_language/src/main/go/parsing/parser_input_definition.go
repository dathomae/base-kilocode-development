// Package parsing provides data structures for the CPL parsing phase.
//
// This package contains parser-specific data structures that are stored
// in the symbol table during parsing and later retrieved by the analysis
// phase to build the final AST structures.
package parsing

import "github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"

// ParserInputDefinition captures a single input declaration during parsing.
// Input declarations appear within a metric's inputs block and reference
// either metrics or dimensions that the calculation uses.
type ParserInputDefinition struct {
	// Kind indicates the type of input: "metric" or "dimension".
	Kind string

	// Name is the referenced metric or dimension identifier.
	Name string

	// Alias is an optional local alias for this input within the metric's scope.
	// If empty, no alias was specified.
	Alias string

	// WhereClause captures optional aspect filter conditions parsed from
	// a where clause on this input declaration.
	WhereClause *ParserWhereClause

	// TimeShift captures an optional backward() or forward() time shift clause.
	// nil if no time shift was specified.
	TimeShift *ParserTimeShift

	// NameLocation records where the referenced name appears in the source.
	// Used for error reporting during semantic analysis.
	NameLocation symboltable.SourceLocation
}

// IsDefinitionData marks ParserInputDefinition as a valid DefinitionData carrier.
func (ParserInputDefinition) IsDefinitionData() {}
