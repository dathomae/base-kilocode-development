// Package parsing provides data structures for the CPL parsing phase.
//
// This package contains parser-specific data structures that are stored
// in the symbol table during parsing and later retrieved by the analysis
// phase to build the final AST structures.
package parsing

// ParserWhereClause represents an OR-of-AND tree for where clause conditions.
// Each AndGroup is combined with OR; conditions within a group are combined with AND.
type ParserWhereClause struct {
	AndGroups []ParserWhereAndGroup
}

// ParserWhereAndGroup represents a group of conditions combined with AND.
type ParserWhereAndGroup struct {
	Conditions []ParserWhereCondition
}

// ParserWhereCondition captures a single aspect filter condition from a where clause.
// Each condition represents one aspectExpression: e.g., aspect=ACTUALS, aspect=FORECAST.
// Negation is handled per-condition via the Negated field.
type ParserWhereCondition struct {
	// Aspect is the aspect value: "ACTUALS" or "FORECAST"
	Aspect string

	// Negated is true if the condition is preceded by the "not" keyword.
	Negated bool
}
