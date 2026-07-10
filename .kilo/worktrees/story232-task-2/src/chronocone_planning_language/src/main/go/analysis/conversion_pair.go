// Package analysis provides the analysis phase data structures for the Chronocone Planning Language.
//
// This file defines the generic ConversionPair type that replaces the two concrete
// conversion pair types with a single parameterized struct.
package analysis

// ConversionPair represents a single analyzed conversion between two definitions.
//
// The type parameter Def is the concrete definition type (e.g. AnalysisUnitDefinition
// or AnalysisMeasureDefinition).  The struct stores pointers to the source and target
// definitions, the pattern-bound variable names, and the resolved calculation
// expression tree.
//
// SourceDef and TargetDef are pointers so that the zero value of ConversionPair
// is usable (nil pointers) and so that the pair can reference definitions stored
// in maps without copying them.
type ConversionPair[Def any] struct {
	// SourceDef is the source definition in the conversion (the "from" side).
	SourceDef *Def

	// TargetDef is the target definition in the conversion (the "to" side).
	TargetDef *Def

	// TargetVar is the pattern-bound variable for the target definition.
	TargetVar string

	// SourceVar is the pattern-bound variable for the source definition.
	SourceVar string

	// Calculation is the resolved expression tree for the conversion formula.
	Calculation Expression
}
