// Package analysis_test contains unit tests for unit conversion analysis.
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Helper Factories
// ============================================================================

func makeUnitRefDef(name string) symboltable.SymbolDefinition {
	return symboltable.SymbolDefinition{
		DefinitionType: symboltable.UnitReferenceDefinition,
		Data:           parser.ParserUnitReference{Name: name},
	}
}

func makeIdentifierDef(name string) symboltable.SymbolDefinition {
	return symboltable.SymbolDefinition{
		DefinitionType: symboltable.Identifier,
		Name:           &name,
	}
}

func makeNumberDef(value float64) symboltable.SymbolDefinition {
	return symboltable.SymbolDefinition{
		DefinitionType: symboltable.NumberDefinition,
		Data:           symboltable.FloatData{Value: value},
	}
}

func makeOperatorDef(defType symboltable.DefinitionType) symboltable.SymbolDefinition {
	return symboltable.SymbolDefinition{
		DefinitionType: defType,
		Data:           parser.ParserOperator{},
	}
}

// ============================================================================
// Success Cases
// ============================================================================

// TestAnalyzeUnitConversions_KilometerMeter verifies a simple valid conversion
// between Kilometer and Meter.
func TestAnalyzeUnitConversions_KilometerMeter(t *testing.T) {
	t.Parallel()

	// Arrange
	units := map[string]analysis.AnalysisUnitDefinition{
		"Kilometer": {Name: "Kilometer", ConceptName: "Distance"},
		"Meter":     {Name: "Meter", ConceptName: "Distance"},
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

	// Act
	err := analysis.AnalyzeUnitConversions(unitScope, units, &unitDef)

	// Assert
	require.NoError(t, err)
	require.Len(t, unitDef.Conversions, 2)

	pair := unitDef.Conversions[0]
	assert.Equal(t, "Meter", pair.SourceDef.Name)
	assert.Equal(t, "Kilometer", pair.TargetDef.Name)
	assert.Equal(t, "m", pair.SourceVar)
	assert.Equal(t, "km", pair.TargetVar)
	require.NotNil(t, pair.Calculation)

	// Inverse pair
	inverse := unitDef.Conversions[1]
	assert.Equal(t, "Kilometer", inverse.SourceDef.Name)
	assert.Equal(t, "Meter", inverse.TargetDef.Name)
	assert.Equal(t, "km", inverse.SourceVar)
	assert.Equal(t, "m", inverse.TargetVar)
	require.NotNil(t, inverse.Calculation)
}

// TestAnalyzeUnitConversions_CelsiusFahrenheit verifies a conversion with
// numbers and multiple operators.
func TestAnalyzeUnitConversions_CelsiusFahrenheit(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"Celsius":    {Name: "Celsius", ConceptName: "Temperature"},
		"Fahrenheit": {Name: "Fahrenheit", ConceptName: "Temperature"},
	}

	unitScope := symboltable.NewSymbolTable(symboltable.UnitScope)
	conversionScope := unitScope.AddChild(symboltable.ConversionScope)

	patternScope := conversionScope.AddChild(symboltable.PatternScope)
	patternScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("Celsius"),
		makeIdentifierDef("C"),
		makeUnitRefDef("Fahrenheit"),
		makeIdentifierDef("F"),
	}

	calcScope := conversionScope.AddChild(symboltable.CalculationScope)
	// Expression: C * 9 / 5 + 32
	// Postfix: C, 9, *, 5, /, 32, +
	calcScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("C"),
		makeNumberDef(9),
		makeOperatorDef(symboltable.MultiplyOperatorDefinition),
		makeNumberDef(5),
		makeOperatorDef(symboltable.DivideOperatorDefinition),
		makeNumberDef(32),
		makeOperatorDef(symboltable.PlusOperatorDefinition),
	}

	unitDef := analysis.AnalysisUnitDefinition{Name: "Celsius", ConceptName: "Temperature"}

	err := analysis.AnalyzeUnitConversions(unitScope, units, &unitDef)
	require.NoError(t, err)
	require.Len(t, unitDef.Conversions, 2)

	// Forward: Fahrenheit → Celsius
	assert.Equal(t, "Fahrenheit", unitDef.Conversions[0].SourceDef.Name)
	assert.Equal(t, "Celsius", unitDef.Conversions[0].TargetDef.Name)

	// Inverse: Celsius → Fahrenheit
	assert.Equal(t, "Celsius", unitDef.Conversions[1].SourceDef.Name)
	assert.Equal(t, "Fahrenheit", unitDef.Conversions[1].TargetDef.Name)
}

// TestAnalyzeUnitConversions_MultipleConversions verifies a unit with more than
// one conversion scope.
func TestAnalyzeUnitConversions_MultipleConversions(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"Kilometer": {Name: "Kilometer", ConceptName: "Distance"},
		"Meter":     {Name: "Meter", ConceptName: "Distance"},
		"Centimeter": {Name: "Centimeter", ConceptName: "Distance"},
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

	err := analysis.AnalyzeUnitConversions(unitScope, units, &unitDef)
	require.NoError(t, err)
	require.Len(t, unitDef.Conversions, 4, "2 conversions x 2 pairs each = 4 pairs")
}

// ============================================================================
// Validation Error Cases
// ============================================================================

// TestAnalyzeUnitConversions_DifferentConceptGroups rejects a conversion where
// the two pattern units belong to different concepts.
func TestAnalyzeUnitConversions_DifferentConceptGroups(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"Kilometer": {Name: "Kilometer", ConceptName: "Distance"},
		"Second":    {Name: "Second", ConceptName: "Time"},
	}

	unitScope := symboltable.NewSymbolTable(symboltable.UnitScope)
	conversionScope := unitScope.AddChild(symboltable.ConversionScope)

	patternScope := conversionScope.AddChild(symboltable.PatternScope)
	patternScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("Kilometer"),
		makeIdentifierDef("km"),
		makeUnitRefDef("Second"),
		makeIdentifierDef("s"),
	}

	calcScope := conversionScope.AddChild(symboltable.CalculationScope)
	calcScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("s"),
		makeNumberDef(1000),
		makeOperatorDef(symboltable.MultiplyOperatorDefinition),
	}

	unitDef := analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}

	err := analysis.AnalyzeUnitConversions(unitScope, units, &unitDef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "different groups")
}

// TestAnalyzeUnitConversions_VariableNameCollision rejects a pattern-bound
// variable that shares its name with a unit.
func TestAnalyzeUnitConversions_VariableNameCollision(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"Kilometer": {Name: "Kilometer", ConceptName: "Distance"},
		"Meter":     {Name: "Meter", ConceptName: "Distance"},
		"m":         {Name: "m", ConceptName: "Distance"}, // conflicts with variable
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

	err := analysis.AnalyzeUnitConversions(unitScope, units, &unitDef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with unit name")
}

// TestAnalyzeUnitConversions_BareNumbers rejects a calculation that contains
// only bare numbers and operators (no unit references).
func TestAnalyzeUnitConversions_BareNumbers(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"Kilometer": {Name: "Kilometer", ConceptName: "Distance"},
		"Meter":     {Name: "Meter", ConceptName: "Distance"},
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

	err := analysis.AnalyzeUnitConversions(unitScope, units, &unitDef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bare numbers")
}

// TestAnalyzeUnitConversions_UndefinedUnitInPattern rejects a pattern that
// references a unit not present in the units map.
func TestAnalyzeUnitConversions_UndefinedUnitInPattern(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"Kilometer": {Name: "Kilometer", ConceptName: "Distance"},
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

	err := analysis.AnalyzeUnitConversions(unitScope, units, &unitDef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "undefined unit")
}

// TestAnalyzeUnitConversions_UndefinedCalcReference rejects a calculation that
// references an identifier that is neither a pattern-bound variable nor a known unit.
func TestAnalyzeUnitConversions_UndefinedCalcReference(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"Kilometer": {Name: "Kilometer", ConceptName: "Distance"},
		"Meter":     {Name: "Meter", ConceptName: "Distance"},
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
		makeUnitRefDef("bogus"),
		makeNumberDef(1000),
		makeOperatorDef(symboltable.MultiplyOperatorDefinition),
	}

	unitDef := analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}

	err := analysis.AnalyzeUnitConversions(unitScope, units, &unitDef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "undefined identifier")
}

// TestAnalyzeUnitConversions_MissingPatternScope rejects a conversion without
// a PatternScope child.
func TestAnalyzeUnitConversions_MissingPatternScope(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"Kilometer": {Name: "Kilometer", ConceptName: "Distance"},
	}

	unitScope := symboltable.NewSymbolTable(symboltable.UnitScope)
	conversionScope := unitScope.AddChild(symboltable.ConversionScope)
	_ = conversionScope.AddChild(symboltable.CalculationScope)

	unitDef := analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}

	err := analysis.AnalyzeUnitConversions(unitScope, units, &unitDef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing pattern scope")
}

// TestAnalyzeUnitConversions_MissingCalculationScope rejects a conversion without
// a CalculationScope child.
func TestAnalyzeUnitConversions_MissingCalculationScope(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"Kilometer": {Name: "Kilometer", ConceptName: "Distance"},
	}

	unitScope := symboltable.NewSymbolTable(symboltable.UnitScope)
	conversionScope := unitScope.AddChild(symboltable.ConversionScope)
	_ = conversionScope.AddChild(symboltable.PatternScope)

	unitDef := analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}

	err := analysis.AnalyzeUnitConversions(unitScope, units, &unitDef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing calculation scope")
}

// TestAnalyzeUnitConversions_InsufficientPatternPairs rejects a pattern with
// fewer than two unit-variable pairs.
func TestAnalyzeUnitConversions_InsufficientPatternPairs(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"Kilometer": {Name: "Kilometer", ConceptName: "Distance"},
	}

	unitScope := symboltable.NewSymbolTable(symboltable.UnitScope)
	conversionScope := unitScope.AddChild(symboltable.ConversionScope)

	patternScope := conversionScope.AddChild(symboltable.PatternScope)
	patternScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("Kilometer"),
		makeIdentifierDef("km"),
	}

	calcScope := conversionScope.AddChild(symboltable.CalculationScope)
	calcScope.Definitions = []symboltable.SymbolDefinition{
		makeUnitRefDef("km"),
		makeNumberDef(1),
		makeOperatorDef(symboltable.MultiplyOperatorDefinition),
	}

	unitDef := analysis.AnalysisUnitDefinition{Name: "Kilometer", ConceptName: "Distance"}

	err := analysis.AnalyzeUnitConversions(unitScope, units, &unitDef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least two unit-variable pairs")
}

// ============================================================================
// Integration Tests via Analyze()
// ============================================================================

// TestAnalyze_UnitConversions_Integration verifies that the full Analyze pipeline
// correctly processes unit scopes that contain conversions.
func TestAnalyze_UnitConversions_Integration(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]

units
	Kilometer = Distance
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation m * 1000
	Meter = Distance`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	kmUnit, exists := defs.Units["Kilometer"]
	require.True(t, exists, "Kilometer should be in units map")
	require.Len(t, kmUnit.Conversions, 2, "Kilometer should have forward + inverse = 2 conversions")
	assert.Equal(t, "Meter", kmUnit.Conversions[0].SourceDef.Name)
	assert.Equal(t, "Kilometer", kmUnit.Conversions[0].TargetDef.Name)
	assert.Equal(t, "m", kmUnit.Conversions[0].SourceVar)

	// Verify inverse exists
	assert.Equal(t, "Kilometer", kmUnit.Conversions[1].SourceDef.Name)
	assert.Equal(t, "Meter", kmUnit.Conversions[1].TargetDef.Name)
}

// TestAnalyze_UnitConversions_InvalidConceptGroup_Integration verifies that an
// invalid conversion across different concept groups is rejected by the full
// Analyze pipeline.
func TestAnalyze_UnitConversions_InvalidConceptGroup_Integration(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]
	Time = [Values[s, ms]]

units
	Kilometer = Distance
		conversion
			pattern Kilometer(km) = Second(s)
			calculation s * 1000
	Second = Time`

	_, err := parseAndAnalyze(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "different groups")
}
