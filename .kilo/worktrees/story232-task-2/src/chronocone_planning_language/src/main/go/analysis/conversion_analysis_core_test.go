package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Unit Definition Tests
// ============================================================================

// TestAnalyzeConversions_UnitSuccess verifies that analyzeConversions correctly
// processes a unit conversion scope and appends forward and inverse pairs.
func TestAnalyzeConversions_UnitSuccess(t *testing.T) {
	t.Parallel()

	km := analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}
	m := analysis.AnalysisUnitDefinition{Name: "Meter", ConceptName: "Distance"}
	lookup := map[string]*analysis.AnalysisUnitDefinition{
		"Kilometer": &km,
		"Meter":     &m,
	}
	units := map[string]analysis.AnalysisUnitDefinition{
		"Kilometer": km,
		"Meter":     m,
	}

	unitScope := symboltable.NewSymbolTable(symboltable.UnitScope)
	conversionScope := unitScope.AddChild(symboltable.ConversionScope)

	patternScope := conversionScope.AddChild(symboltable.PatternScope)
	patternScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("Kilometer"),
		makeIdentifierDef("km"),
		makeUnitRefDef("Meter"),
		makeIdentifierDef("m"),
	}

	calcScope := conversionScope.AddChild(symboltable.CalculationScope)
	calcScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("m"),
		makeNumberDef(1000),
		makeOperatorDef(symboltable.MultiplyOperatorDefinition),
	}

	unitDef := analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}

	err := analysis.AnalyzeConversionsForTest(
		unitScope, symboltable.ConversionScope, lookup,
		unitDef.Name, &unitDef, units, "unit", nil,
	)

	require.NoError(t, err)
	require.Len(t, unitDef.Conversions, 2)

	forward := unitDef.Conversions[0]
	assert.Equal(t, "Meter", forward.SourceDef.Name)
	assert.Equal(t, "Kilometer", forward.TargetDef.Name)
	assert.Equal(t, "m", forward.SourceVar)
	assert.Equal(t, "km", forward.TargetVar)
	require.NotNil(t, forward.Calculation)

	inverse := unitDef.Conversions[1]
	assert.Equal(t, "Kilometer", inverse.SourceDef.Name)
	assert.Equal(t, "Meter", inverse.TargetDef.Name)
	assert.Equal(t, "km", inverse.SourceVar)
	assert.Equal(t, "m", inverse.TargetVar)
	require.NotNil(t, inverse.Calculation)
}

// TestAnalyzeConversions_UnitMissingPatternScope rejects a conversion without a PatternScope.
func TestAnalyzeConversions_UnitMissingPatternScope(t *testing.T) {
	t.Parallel()

	km := analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}
	lookup := map[string]*analysis.AnalysisUnitDefinition{
		"Kilometer": &km,
	}
	units := map[string]analysis.AnalysisUnitDefinition{
		"Kilometer": km,
	}

	unitScope := symboltable.NewSymbolTable(symboltable.UnitScope)
	conversionScope := unitScope.AddChild(symboltable.ConversionScope)
	_ = conversionScope.AddChild(symboltable.CalculationScope)

	unitDef := analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}

	err := analysis.AnalyzeConversionsForTest(
		unitScope, symboltable.ConversionScope, lookup,
		unitDef.Name, &unitDef, units, "unit", nil,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing pattern scope")
}

// TestAnalyzeConversions_UnitBareNumbers rejects a calculation with only bare numbers.
func TestAnalyzeConversions_UnitBareNumbers(t *testing.T) {
	t.Parallel()

	km := analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}
	m := analysis.AnalysisUnitDefinition{Name: "Meter", ConceptName: "Distance"}
	lookup := map[string]*analysis.AnalysisUnitDefinition{
		"Kilometer": &km,
		"Meter":     &m,
	}
	units := map[string]analysis.AnalysisUnitDefinition{
		"Kilometer": km,
		"Meter":     m,
	}

	unitScope := symboltable.NewSymbolTable(symboltable.UnitScope)
	conversionScope := unitScope.AddChild(symboltable.ConversionScope)

	patternScope := conversionScope.AddChild(symboltable.PatternScope)
	patternScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("Kilometer"),
		makeIdentifierDef("km"),
		makeUnitRefDef("Meter"),
		makeIdentifierDef("m"),
	}

	calcScope := conversionScope.AddChild(symboltable.CalculationScope)
	calcScope.Definitions = []symboltable.SymbolDefinition{
		makeNumberDef(5),
		makeNumberDef(9),
		makeOperatorDef(symboltable.DivideOperatorDefinition),
	}

	unitDef := analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}

	err := analysis.AnalyzeConversionsForTest(
		unitScope, symboltable.ConversionScope, lookup,
		unitDef.Name, &unitDef, units, "unit", nil,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bare numbers")
}

// TestAnalyzeConversions_UnitMultipleConversions verifies handling of multiple conversion scopes.
func TestAnalyzeConversions_UnitMultipleConversions(t *testing.T) {
	t.Parallel()

	km := analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}
	m := analysis.AnalysisUnitDefinition{Name: "Meter", ConceptName: "Distance"}
	cm := analysis.AnalysisUnitDefinition{Name: "Centimeter", ConceptName: "Distance"}
	lookup := map[string]*analysis.AnalysisUnitDefinition{
		"Kilometer":  &km,
		"Meter":      &m,
		"Centimeter": &cm,
	}
	units := map[string]analysis.AnalysisUnitDefinition{
		"Kilometer":  km,
		"Meter":      m,
		"Centimeter": cm,
	}

	unitScope := symboltable.NewSymbolTable(symboltable.UnitScope)

	// First conversion: Kilometer <-> Meter
	conv1 := unitScope.AddChild(symboltable.ConversionScope)
	pattern1 := conv1.AddChild(symboltable.PatternScope)
	pattern1.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("Kilometer"),
		makeIdentifierDef("km"),
		makeUnitRefDef("Meter"),
		makeIdentifierDef("m"),
	}
	calc1 := conv1.AddChild(symboltable.CalculationScope)
	calc1.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("m"),
		makeNumberDef(1000),
		makeOperatorDef(symboltable.MultiplyOperatorDefinition),
	}

	// Second conversion: Kilometer <-> Centimeter
	conv2 := unitScope.AddChild(symboltable.ConversionScope)
	pattern2 := conv2.AddChild(symboltable.PatternScope)
	pattern2.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("Kilometer"),
		makeIdentifierDef("km"),
		makeUnitRefDef("Centimeter"),
		makeIdentifierDef("cm"),
	}
	calc2 := conv2.AddChild(symboltable.CalculationScope)
	calc2.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("cm"),
		makeNumberDef(100000),
		makeOperatorDef(symboltable.MultiplyOperatorDefinition),
	}

	unitDef := analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}

	err := analysis.AnalyzeConversionsForTest(
		unitScope, symboltable.ConversionScope, lookup,
		unitDef.Name, &unitDef, units, "unit", nil,
	)

	require.NoError(t, err)
	require.Len(t, unitDef.Conversions, 4, "2 conversions x 2 pairs each = 4 pairs")
}

// TestAnalyzeConversions_UnitUndefinedPatternUnit rejects a pattern referencing an undefined unit.
func TestAnalyzeConversions_UnitUndefinedPatternUnit(t *testing.T) {
	t.Parallel()

	km := analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}
	lookup := map[string]*analysis.AnalysisUnitDefinition{
		"Kilometer": &km,
	}
	units := map[string]analysis.AnalysisUnitDefinition{
		"Kilometer": km,
	}

	unitScope := symboltable.NewSymbolTable(symboltable.UnitScope)
	conversionScope := unitScope.AddChild(symboltable.ConversionScope)

	patternScope := conversionScope.AddChild(symboltable.PatternScope)
	patternScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("Kilometer"),
		makeIdentifierDef("km"),
		makeUnitRefDef("Bogus"),
		makeIdentifierDef("b"),
	}

	calcScope := conversionScope.AddChild(symboltable.CalculationScope)
	calcScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("b"),
		makeNumberDef(1000),
		makeOperatorDef(symboltable.MultiplyOperatorDefinition),
	}

	unitDef := analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}

	err := analysis.AnalyzeConversionsForTest(
		unitScope, symboltable.ConversionScope, lookup,
		unitDef.Name, &unitDef, units, "unit", nil,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "undefined")
}

// ============================================================================
// Measure Definition Tests
// ============================================================================

// TestAnalyzeConversions_MeasureSuccess verifies that analyzeConversions correctly
// processes a measure conversion scope and appends forward and inverse pairs.
func TestAnalyzeConversions_MeasureSuccess(t *testing.T) {
	t.Parallel()

	kmMeasure := analysis.AnalysisMeasureDefinition{Name: "KilometerLength"}
	mMeasure := analysis.AnalysisMeasureDefinition{Name: "MeterLength"}
	lookup := map[string]*analysis.AnalysisMeasureDefinition{
		"KilometerLength": &kmMeasure,
		"MeterLength":     &mMeasure,
	}
	units := map[string]analysis.AnalysisUnitDefinition{
		"KilometerUnit": {Name: "KilometerUnit", ConceptName: "Distance"},
		"MeterUnit":     {Name: "MeterUnit", ConceptName: "Distance"},
	}

	measureScope := symboltable.NewSymbolTable(symboltable.MeasureScope)
	convScope := measureScope.AddChild(symboltable.MeasureConversionScope)

	patternScope := convScope.AddChild(symboltable.PatternScope)
	patternScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("KilometerLength"),
		makeIdentifierDef("km"),
		makeUnitRefDef("MeterLength"),
		makeIdentifierDef("m"),
	}

	calcScope := convScope.AddChild(symboltable.CalculationScope)
	calcScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("m"),
		makeNumberDef(1000),
		makeOperatorDef(symboltable.MultiplyOperatorDefinition),
	}

	measureDef := analysis.AnalysisMeasureDefinition{Name: "KilometerLength"}

	err := analysis.AnalyzeConversionsForTest(
		measureScope, symboltable.MeasureConversionScope, lookup,
		measureDef.Name, &measureDef, units, "measure", nil,
	)

	require.NoError(t, err)
	require.Len(t, measureDef.Conversions, 2)

	forward := measureDef.Conversions[0]
	assert.Equal(t, "MeterLength", forward.SourceDef.Name)
	assert.Equal(t, "KilometerLength", forward.TargetDef.Name)
	assert.Equal(t, "m", forward.SourceVar)
	assert.Equal(t, "km", forward.TargetVar)
	require.NotNil(t, forward.Calculation)

	inverse := measureDef.Conversions[1]
	assert.Equal(t, "KilometerLength", inverse.SourceDef.Name)
	assert.Equal(t, "MeterLength", inverse.TargetDef.Name)
	assert.Equal(t, "km", inverse.SourceVar)
	assert.Equal(t, "m", inverse.TargetVar)
	require.NotNil(t, inverse.Calculation)
}

// TestAnalyzeConversions_MeasureMissingCalculationScope rejects a conversion without a CalculationScope.
func TestAnalyzeConversions_MeasureMissingCalculationScope(t *testing.T) {
	t.Parallel()

	kmMeasure := analysis.AnalysisMeasureDefinition{Name: "KilometerLength"}
	lookup := map[string]*analysis.AnalysisMeasureDefinition{
		"KilometerLength": &kmMeasure,
	}
	units := map[string]analysis.AnalysisUnitDefinition{}

	measureScope := symboltable.NewSymbolTable(symboltable.MeasureScope)
	convScope := measureScope.AddChild(symboltable.MeasureConversionScope)
	_ = convScope.AddChild(symboltable.PatternScope)

	measureDef := analysis.AnalysisMeasureDefinition{Name: "KilometerLength"}

	err := analysis.AnalyzeConversionsForTest(
		measureScope, symboltable.MeasureConversionScope, lookup,
		measureDef.Name, &measureDef, units, "measure", nil,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing calculation scope")
}

// TestAnalyzeConversions_MeasureVariableCollision rejects a pattern-bound variable that collides with a unit name.
func TestAnalyzeConversions_MeasureVariableCollision(t *testing.T) {
	t.Parallel()

	kmMeasure := analysis.AnalysisMeasureDefinition{Name: "KilometerLength"}
	mMeasure := analysis.AnalysisMeasureDefinition{Name: "MeterLength"}
	lookup := map[string]*analysis.AnalysisMeasureDefinition{
		"KilometerLength": &kmMeasure,
		"MeterLength":     &mMeasure,
	}
	units := map[string]analysis.AnalysisUnitDefinition{
		"km": {Name: "km", ConceptName: "Distance"}, // collides with variable
	}

	measureScope := symboltable.NewSymbolTable(symboltable.MeasureScope)
	convScope := measureScope.AddChild(symboltable.MeasureConversionScope)

	patternScope := convScope.AddChild(symboltable.PatternScope)
	patternScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("KilometerLength"),
		makeIdentifierDef("km"),
		makeUnitRefDef("MeterLength"),
		makeIdentifierDef("m"),
	}

	calcScope := convScope.AddChild(symboltable.CalculationScope)
	calcScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("m"),
		makeNumberDef(1000),
		makeOperatorDef(symboltable.MultiplyOperatorDefinition),
	}

	measureDef := analysis.AnalysisMeasureDefinition{Name: "KilometerLength"}

	err := analysis.AnalyzeConversionsForTest(
		measureScope, symboltable.MeasureConversionScope, lookup,
		measureDef.Name, &measureDef, units, "measure", nil,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with unit name")
}

// TestAnalyzeConversions_MeasureMultipleConversions verifies handling of multiple measure conversion scopes.
func TestAnalyzeConversions_MeasureMultipleConversions(t *testing.T) {
	t.Parallel()

	kmMeasure := analysis.AnalysisMeasureDefinition{Name: "KmMeasure"}
	mMeasure := analysis.AnalysisMeasureDefinition{Name: "MeterMeasure"}
	cmMeasure := analysis.AnalysisMeasureDefinition{Name: "CmMeasure"}
	lookup := map[string]*analysis.AnalysisMeasureDefinition{
		"KmMeasure":    &kmMeasure,
		"MeterMeasure": &mMeasure,
		"CmMeasure":    &cmMeasure,
	}
	units := map[string]analysis.AnalysisUnitDefinition{}

	measureScope := symboltable.NewSymbolTable(symboltable.MeasureScope)

	// First conversion: KmMeasure <-> MeterMeasure
	conv1 := measureScope.AddChild(symboltable.MeasureConversionScope)
	pattern1 := conv1.AddChild(symboltable.PatternScope)
	pattern1.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("KmMeasure"),
		makeIdentifierDef("km"),
		makeUnitRefDef("MeterMeasure"),
		makeIdentifierDef("m"),
	}
	calc1 := conv1.AddChild(symboltable.CalculationScope)
	calc1.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("m"),
		makeNumberDef(1000),
		makeOperatorDef(symboltable.MultiplyOperatorDefinition),
	}

	// Second conversion: KmMeasure <-> CmMeasure
	conv2 := measureScope.AddChild(symboltable.MeasureConversionScope)
	pattern2 := conv2.AddChild(symboltable.PatternScope)
	pattern2.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("KmMeasure"),
		makeIdentifierDef("km"),
		makeUnitRefDef("CmMeasure"),
		makeIdentifierDef("cm"),
	}
	calc2 := conv2.AddChild(symboltable.CalculationScope)
	calc2.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("cm"),
		makeNumberDef(100000),
		makeOperatorDef(symboltable.MultiplyOperatorDefinition),
	}

	measureDef := analysis.AnalysisMeasureDefinition{Name: "KmMeasure"}

	err := analysis.AnalyzeConversionsForTest(
		measureScope, symboltable.MeasureConversionScope, lookup,
		measureDef.Name, &measureDef, units, "measure", nil,
	)

	require.NoError(t, err)
	require.Len(t, measureDef.Conversions, 4, "2 conversions x 2 pairs each = 4 pairs")
}

// TestAnalyzeConversions_MeasureInsufficientPatternPairs rejects a pattern with fewer than two pairs.
func TestAnalyzeConversions_MeasureInsufficientPatternPairs(t *testing.T) {
	t.Parallel()

	kmMeasure := analysis.AnalysisMeasureDefinition{Name: "KilometerLength"}
	lookup := map[string]*analysis.AnalysisMeasureDefinition{
		"KilometerLength": &kmMeasure,
	}
	units := map[string]analysis.AnalysisUnitDefinition{}

	measureScope := symboltable.NewSymbolTable(symboltable.MeasureScope)
	convScope := measureScope.AddChild(symboltable.MeasureConversionScope)

	patternScope := convScope.AddChild(symboltable.PatternScope)
	patternScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("KilometerLength"),
		makeIdentifierDef("km"),
	}

	calcScope := convScope.AddChild(symboltable.CalculationScope)
	calcScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("km"),
		makeNumberDef(1),
		makeOperatorDef(symboltable.MultiplyOperatorDefinition),
	}

	measureDef := analysis.AnalysisMeasureDefinition{Name: "KilometerLength"}

	err := analysis.AnalyzeConversionsForTest(
		measureScope, symboltable.MeasureConversionScope, lookup,
		measureDef.Name, &measureDef, units, "measure", nil,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least two unit-variable pairs")
}
