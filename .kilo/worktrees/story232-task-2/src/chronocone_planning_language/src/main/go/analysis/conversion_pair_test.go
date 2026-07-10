// Package analysis_test contains unit tests for the generic ConversionPair type.
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// ConversionPair[AnalysisUnitDefinition] tests
// ============================================================================

// TestConversionPair_UnitDefinition_Construction verifies that a
// ConversionPair[AnalysisUnitDefinition] can be constructed and its fields
// accessed correctly.
func TestConversionPair_UnitDefinition_Construction(t *testing.T) {
	t.Parallel()

	source := &analysis.AnalysisUnitDefinition{
		Name:        "Meter",
		ConceptName: "Distance",
		Location:    symboltable.SourceLocation{Line: 10},
	}
	target := &analysis.AnalysisUnitDefinition{
		Name:        "Kilometer",
		ConceptName: "Distance",
		Location:    symboltable.SourceLocation{Line: 20},
	}
	calc := &analysis.MultiplicationNode{
		Left:  &analysis.VariableRef{Name: "m"},
		Right: &analysis.NumberNode{Value: 1000},
	}

	pair := analysis.ConversionPair[analysis.AnalysisUnitDefinition]{
		SourceDef:   source,
		TargetDef:   target,
		SourceVar:   "m",
		TargetVar:   "km",
		Calculation: calc,
	}

	assert.Same(t, source, pair.SourceDef)
	assert.Same(t, target, pair.TargetDef)
	assert.Equal(t, "m", pair.SourceVar)
	assert.Equal(t, "km", pair.TargetVar)
	assert.Same(t, calc, pair.Calculation)
}

// TestConversionPair_UnitDefinition_ZeroValue verifies that the zero value of
// ConversionPair[AnalysisUnitDefinition] has nil definition pointers and empty
// strings.
func TestConversionPair_UnitDefinition_ZeroValue(t *testing.T) {
	t.Parallel()

	var pair analysis.ConversionPair[analysis.AnalysisUnitDefinition]

	assert.Nil(t, pair.SourceDef)
	assert.Nil(t, pair.TargetDef)
	assert.Empty(t, pair.SourceVar)
	assert.Empty(t, pair.TargetVar)
	assert.Nil(t, pair.Calculation)
}

// ============================================================================
// ConversionPair[AnalysisMeasureDefinition] tests
// ============================================================================

// TestConversionPair_MeasureDefinition_Construction verifies that a
// ConversionPair[AnalysisMeasureDefinition] can be constructed and its fields
// accessed correctly.
func TestConversionPair_MeasureDefinition_Construction(t *testing.T) {
	t.Parallel()

	source := &analysis.AnalysisMeasureDefinition{
		Name:     "MeterLength",
		TypeName: "Number",
		Location: symboltable.SourceLocation{Line: 10},
	}
	target := &analysis.AnalysisMeasureDefinition{
		Name:     "KilometerLength",
		TypeName: "Number",
		Location: symboltable.SourceLocation{Line: 20},
	}
	calc := &analysis.DivisionNode{
		Left:  &analysis.VariableRef{Name: "km"},
		Right: &analysis.NumberNode{Value: 1000},
	}

	pair := analysis.ConversionPair[analysis.AnalysisMeasureDefinition]{
		SourceDef:   source,
		TargetDef:   target,
		SourceVar:   "m",
		TargetVar:   "km",
		Calculation: calc,
	}

	assert.Same(t, source, pair.SourceDef)
	assert.Same(t, target, pair.TargetDef)
	assert.Equal(t, "m", pair.SourceVar)
	assert.Equal(t, "km", pair.TargetVar)
	assert.Same(t, calc, pair.Calculation)
}

// TestConversionPair_MeasureDefinition_ZeroValue verifies that the zero value of
// ConversionPair[AnalysisMeasureDefinition] has nil definition pointers and empty
// strings.
func TestConversionPair_MeasureDefinition_ZeroValue(t *testing.T) {
	t.Parallel()

	var pair analysis.ConversionPair[analysis.AnalysisMeasureDefinition]

	assert.Nil(t, pair.SourceDef)
	assert.Nil(t, pair.TargetDef)
	assert.Empty(t, pair.SourceVar)
	assert.Empty(t, pair.TargetVar)
	assert.Nil(t, pair.Calculation)
}

// ============================================================================
// ConversionPair field mutation tests
// ============================================================================

// TestConversionPair_UnitDefinition_FieldMutation verifies that fields can be
// mutated after construction.
func TestConversionPair_UnitDefinition_FieldMutation(t *testing.T) {
	t.Parallel()

	source := &analysis.AnalysisUnitDefinition{Name: "Meter", ConceptName: "Distance"}
	target := &analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}

	pair := analysis.ConversionPair[analysis.AnalysisUnitDefinition]{
		SourceDef: source,
		TargetDef: target,
		SourceVar: "m",
		TargetVar: "km",
	}

	pair.SourceVar = "meter"
	pair.TargetVar = "kilometer"
	pair.Calculation = &analysis.NumberNode{Value: 42}

	assert.Equal(t, "meter", pair.SourceVar)
	assert.Equal(t, "kilometer", pair.TargetVar)
	assert.NotNil(t, pair.Calculation)
}

// TestConversionPair_MeasureDefinition_FieldMutation verifies that fields can be
// mutated after construction.
func TestConversionPair_MeasureDefinition_FieldMutation(t *testing.T) {
	t.Parallel()

	source := &analysis.AnalysisMeasureDefinition{Name: "MeterLength", TypeName: "Number"}
	target := &analysis.AnalysisMeasureDefinition{Name: "KilometerLength", TypeName: "Number"}

	pair := analysis.ConversionPair[analysis.AnalysisMeasureDefinition]{
		SourceDef: source,
		TargetDef: target,
		SourceVar: "m",
		TargetVar: "km",
	}

	pair.SourceVar = "meter"
	pair.TargetVar = "kilometer"
	pair.Calculation = &analysis.NumberNode{Value: 42}

	assert.Equal(t, "meter", pair.SourceVar)
	assert.Equal(t, "kilometer", pair.TargetVar)
	assert.NotNil(t, pair.Calculation)
}
