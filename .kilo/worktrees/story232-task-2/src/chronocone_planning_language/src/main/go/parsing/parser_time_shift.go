// Package parsing provides data structures for the CPL parsing phase.
//
// This package contains parser-specific data structures that are stored
// in the symbol table during parsing and later retrieved by the analysis
// phase to build the final AST structures.
package parsing

import "github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"

// ParserTimeShift captures a backward() or forward() time shift clause
// on an input declaration.
type ParserTimeShift struct {
	// Direction indicates the shift direction: "backward" or "forward".
	Direction string

	// Field is the Instant constraint field keyword (e.g., "year", "month", "day",
	// "hour", "minute", "second", "day_of_week", "week_number").
	Field string

	// Expression is the postfix sequence for the shift amount expression.
	// When non-nil, this takes precedence over Amount during analysis.
	// Populated when the grammar produces an expression inside backward()/forward().
	Expression []PostfixElement

	// Amount is the simple numeric shift amount. Used when Expression is nil.
	// The listener always initializes this to 1.0 when constructing a ParserTimeShift.
	Amount float64

	// SourceLocation records where the time-shift clause appears in the source.
	// Used for error reporting during semantic analysis.
	SourceLocation symboltable.SourceLocation
}
