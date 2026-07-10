// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file defines AggregatorInfo, which holds analyzed aggregator information
// including resolved concept references and validated dimensional unit effects.
package analysis

import "github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"

// AggregatorInfo contains the analyzed aggregator information.
// Concept references are resolved, and the dimensional unit effect is validated.
//
// NOTE: AggregatorFunctionInfo, FunctionRegistry, and NewAggregatorFunctionRegistry
// are defined in function_registry.go.
type AggregatorInfo struct {
	FromConcepts         []*AnalysisConceptDefinition
	ToConcept            *AnalysisConceptDefinition
	ToEmpty              bool
	FromDimensionalUnits *DimensionalUnitExpression
	ToDimensionalUnits   *DimensionalUnitExpression
	ExpressionDimUnits   *DimensionalUnitExpression
	CellOperation        string
	SourceLocation       symboltable.SourceLocation
}
