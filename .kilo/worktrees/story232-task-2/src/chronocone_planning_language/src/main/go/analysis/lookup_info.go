// Package analysis provides the analysis phase data structures for the Chronocone Planning Language.
//
// This file defines LookupInfo, the analysis-level data structure that carries
// results for a CPL lookup declaration through the analysis pipeline.
// See Story242: Lookups Increment 1 - Analysis Phase for the implementation plan.
package analysis

import (
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// LookupInfo contains the analyzed lookup information.
//
// This structure represents a fully-processed lookup declaration after the analysis phase.
// All references (dimensions and celltype measure) have been validated and resolved to their
// corresponding analysis-level definitions. The resolved references enable efficient
// validation and code generation during the synthesis phase.
//
// Lookups are analogous to connectors but simpler: a lookup declares a set of dimension
// concepts and a celltype measure, allowing metric calculations to reference pre-computed
// value maps. Unlike connectors, lookups in Increment 1 have no grain.
//
// Example lookup declaration in CPL:
//
//	lookup CountryProductIndex
//	dimensions [Country, Product]
//	celltype UnitsSold
//
// This would result in a LookupInfo with:
//   - Name: "CountryProductIndex"
//   - Dimensions: []*AnalysisConceptDefinition{CountryConcept, ProductConcept}
//   - CelltypeMeasure: *AnalysisMeasureDefinition(UnitsSoldMeasure)
//   - SourceLocation: Location of the lookup declaration in source
type LookupInfo struct {
	// Name is the identifier for this lookup (e.g., "CountryProductIndex", "RegionMetricMap").
	// This name is used to reference the lookup in metrics and other CPL constructs.
	// It must be unique within the CPL document.
	Name string

	// Dimensions is a slice of pointers to resolved concept definitions.
	//
	// Each dimension represents a categorical axis of the lookup's data.
	// The order of dimensions is significant as it defines the dimensional
	// structure of the lookup key. All dimensions must reference valid concepts
	// that have been previously defined in the CPL document.
	//
	// During analysis, each dimension name is looked up in the Concepts map
	// and replaced with a pointer to the corresponding AnalysisConceptDefinition.
	// This ensures that all dimension references are valid and provides
	// direct access to concept information during code generation.
	Dimensions []*AnalysisConceptDefinition

	// CelltypeMeasure is a pointer to the resolved measure definition for the cell values.
	//
	// The celltype defines the data type and unit of the values stored in the
	// lookup. It must reference a valid measure that has been previously
	// defined in the CPL document.
	//
	// During analysis, the celltype measure name is looked up in the Measures map
	// and replaced with a pointer to the corresponding AnalysisMeasureDefinition.
	// This ensures that the celltype reference is valid and provides direct
	// access to measure information during code generation.
	CelltypeMeasure *AnalysisMeasureDefinition

	// SourceLocation records where this lookup was defined in the source CPL.
	//
	// This location information is essential for error reporting during analysis
	// and synthesis phases. When validation errors occur (e.g., undefined
	// dimension or celltype), this location allows the compiler to report
	// precise error messages pointing to the lookup declaration.
	SourceLocation symboltable.SourceLocation
}
