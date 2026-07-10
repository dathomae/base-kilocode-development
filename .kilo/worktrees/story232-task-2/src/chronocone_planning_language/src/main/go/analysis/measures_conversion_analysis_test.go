// Package analysis_test contains unit tests for measure conversion analysis.
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
// Success Cases — Full Pipeline Integration Tests
// ============================================================================

// TestAnalyzeMeasureConversions_KilometerMeter verifies a simple conversion
// between KilometerMeasure and MeterMeasure with algebraic reversal generating
// both forward and inverse pairs.
func TestAnalyzeMeasureConversions_KilometerMeter(t *testing.T) {
	t.Parallel()

	// Pattern: KilometerLength(km) = MeterLength(m)
	// Calculation: m * 1000
	// This means target=KilometerLength, source=MeterLength
	// Forward: from MeterLength to KilometerLength, km = m * 1000
	// Inverse: from KilometerLength to MeterLength, m = km / 1000
	input := `concepts
	Distance = [Values[km, m]]

units
	KilometerUnit = Distance
	MeterUnit = Distance

measures
	KilometerLength = Number(unit=KilometerUnit)
		conversion
			pattern KilometerLength(km) = MeterLength(m)
			calculation m * 1000
	MeterLength = Number(unit=MeterUnit)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	kmMeasure, exists := defs.Measures["KilometerLength"]
	require.True(t, exists, "KilometerLength should exist")
	require.Len(t, kmMeasure.Conversions, 2, "Should have forward + inverse = 2 conversion pairs")

	// Forward pair: MeterLength → KilometerLength, calculation: m * 1000
	forward := kmMeasure.Conversions[0]
	assert.Equal(t, "MeterLength", forward.SourceDef.Name, "forward source should be MeterLength")
	assert.Equal(t, "KilometerLength", forward.TargetDef.Name, "forward target should be KilometerLength")
	assert.Equal(t, "m", forward.SourceVar)
	assert.Equal(t, "km", forward.TargetVar)
	require.NotNil(t, forward.Calculation, "forward calculation should not be nil")
	_, forwardIsMult := forward.Calculation.(*analysis.MultiplicationNode)
	assert.True(t, forwardIsMult, "forward calculation should be multiplication")

	// Inverse pair: KilometerLength → MeterLength, calculation: km / 1000
	inverse := kmMeasure.Conversions[1]
	assert.Equal(t, "KilometerLength", inverse.SourceDef.Name, "inverse source should be KilometerLength")
	assert.Equal(t, "MeterLength", inverse.TargetDef.Name, "inverse target should be MeterLength")
	assert.Equal(t, "km", inverse.SourceVar)
	assert.Equal(t, "m", inverse.TargetVar)
	require.NotNil(t, inverse.Calculation, "inverse calculation should not be nil")
	_, inverseIsDiv := inverse.Calculation.(*analysis.DivisionNode)
	assert.True(t, inverseIsDiv, "inverse calculation should be division")
}

// TestAnalyzeMeasureConversions_AlgebraicReversal_Division verifies
// that division is inverted to multiplication.
func TestAnalyzeMeasureConversions_AlgebraicReversal_Division(t *testing.T) {
	t.Parallel()

	// Pattern: KmMeasure(km) = MeterMeasure(m)
	// Calculation: m / 1000  (using source var m)
	// Forward: MeterMeasure → KmMeasure, km = m / 1000
	// Inverse: KmMeasure → MeterMeasure, m = km * 1000
	input := `concepts
	Distance = [Values[km, m]]

units
	KilometerUnit = Distance
	MeterUnit = Distance

measures
	KmMeasure = Number(unit=KilometerUnit)
		conversion
			pattern KmMeasure(km) = MeterMeasure(m)
			calculation m / 1000
	MeterMeasure = Number(unit=MeterUnit)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)

	kmDef := defs.Measures["KmMeasure"]
	require.Len(t, kmDef.Conversions, 2)

	// Forward: source=MeterMeasure, target=KmMeasure, calc=m/1000
	forward := kmDef.Conversions[0]
	assert.Equal(t, "MeterMeasure", forward.SourceDef.Name)
	assert.Equal(t, "KmMeasure", forward.TargetDef.Name)
	_, forwardIsDiv := forward.Calculation.(*analysis.DivisionNode)
	assert.True(t, forwardIsDiv, "forward should be division")

	// Inverse: source=KmMeasure, target=MeterMeasure, calc=km*1000
	inverse := kmDef.Conversions[1]
	assert.Equal(t, "KmMeasure", inverse.SourceDef.Name)
	assert.Equal(t, "MeterMeasure", inverse.TargetDef.Name)
	_, inverseIsMult := inverse.Calculation.(*analysis.MultiplicationNode)
	assert.True(t, inverseIsMult, "inverse should be multiplication")
}

// TestAnalyzeMeasureConversions_AlgebraicReversal_VariableSwap verifies
// that variables are correctly swapped during inversion.
func TestAnalyzeMeasureConversions_AlgebraicReversal_VariableSwap(t *testing.T) {
	t.Parallel()

	// This also verifies the inversion produces a correct DivisionNode
	// with the expected variable references and number values.
	input := `concepts
	Distance = [Values[km, m]]

units
	KilometerUnit = Distance
	MeterUnit = Distance

measures
	KmMeasure = Number(unit=KilometerUnit)
		conversion
			pattern KmMeasure(km) = MeterMeasure(m)
			calculation m * 1000
	MeterMeasure = Number(unit=MeterUnit)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)

	kmDef := defs.Measures["KmMeasure"]
	require.Len(t, kmDef.Conversions, 2)

	// Inverse pair: km / 1000
	inverse := kmDef.Conversions[1]
	assert.Equal(t, "KmMeasure", inverse.SourceDef.Name)
	assert.Equal(t, "MeterMeasure", inverse.TargetDef.Name)
	assert.Equal(t, "km", inverse.SourceVar)
	assert.Equal(t, "m", inverse.TargetVar)

	div, ok := inverse.Calculation.(*analysis.DivisionNode)
	require.True(t, ok, "inverse should be division")

	// Left should be VariableRef with new variable name
	leftRef, ok := div.Left.(*analysis.VariableRef)
	require.True(t, ok, "left should be VariableRef")
	assert.Equal(t, "km", leftRef.Name)

	// Right should be NumberNode with the original value
	rightNum, ok := div.Right.(*analysis.NumberNode)
	require.True(t, ok, "right should be NumberNode")
	assert.Equal(t, float64(1000), rightNum.Value)
}

// TestAnalyzeMeasureConversions_MultipleConversions verifies a measure
// that declares multiple conversion clauses.
func TestAnalyzeMeasureConversions_MultipleConversions(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m, cm]]

units
	KilometerUnit = Distance
	MeterUnit = Distance
	CentimeterUnit = Distance

measures
	KmMeasure = Number(unit=KilometerUnit)
		conversion
			pattern KmMeasure(km) = MeterMeasure(m)
			calculation m * 1000
		conversion
			pattern KmMeasure(km) = CmMeasure(cm)
			calculation cm * 100000
	MeterMeasure = Number(unit=MeterUnit)
	CmMeasure = Number(unit=CentimeterUnit)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)

	kmMeasure := defs.Measures["KmMeasure"]
	require.Len(t, kmMeasure.Conversions, 4, "2 conversions × 2 pairs each = 4 pairs")
}

// TestAnalyzeMeasureConversions_NoConversions verifies that a measure
// without conversion clauses has an empty Conversions slice.
func TestAnalyzeMeasureConversions_NoConversions(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[m]]

units
	Meter = Distance

measures
	SimpleMeasure = Number(unit=Meter)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)

	measure := defs.Measures["SimpleMeasure"]
	assert.Empty(t, measure.Conversions, "Should have no conversions")
}

// ============================================================================
// Validation Error Cases — Variable Name Collisions
// ============================================================================

// TestAnalyzeMeasureConversions_VariableNameCollision rejects a pattern-bound
// variable whose name collides with a declared unit.
func TestAnalyzeMeasureConversions_VariableNameCollision(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]

units
	Kilometer = Distance
	m = Distance

measures
	KmMeasure = Number(unit=Kilometer)
		conversion
			pattern KmMeasure(km) = MeterMeasure(m)
			calculation m * 1000
	MeterMeasure = Number(unit=Kilometer)`

	_, err := parseAndAnalyze(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with unit name")
}

// ============================================================================
// Validation Error Cases — Bare Numbers
// ============================================================================

// TestAnalyzeMeasureConversions_BareNumbers rejects a calculation with only
// bare numbers in an additive context. Numbers in multiplicative (*, /) contexts
// are allowed as scaling factors, but addition/subtraction requires unit annotations.
func TestAnalyzeMeasureConversions_BareNumbers(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]

units
	KilometerUnit = Distance
	MeterUnit = Distance

measures
	KmMeasure = Number(unit=KilometerUnit)
		conversion
			pattern KmMeasure(km) = MeterMeasure(m)
			calculation 5 + 9
	MeterMeasure = Number(unit=MeterUnit)`

	_, err := parseAndAnalyze(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bare numbers without unit annotations")
}

// ============================================================================
// Validation Error Cases — Self-Conversion
// ============================================================================

// TestAnalyzeMeasureConversions_SelfConversion rejects a conversion where
// source and target are the same measure.
func TestAnalyzeMeasureConversions_SelfConversion(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[m]]

units
	Meter = Distance

measures
	LenMeasure = Number(unit=Meter)
		conversion
			pattern LenMeasure(a) = LenMeasure(b)
			calculation a * 1`

	_, err := parseAndAnalyze(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "self-conversion")
}

// ============================================================================
// Validation Error Cases — Undefined Measure Reference
// ============================================================================

// TestAnalyzeMeasureConversions_UndefinedMeasureInPattern rejects a pattern
// that references a measure not present in the measures map.
func TestAnalyzeMeasureConversions_UndefinedMeasureInPattern(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[m]]

units
	MeterUnit = Distance

measures
	LenMeasure = Number(unit=MeterUnit)
		conversion
			pattern LenMeasure(x) = BogusMeasure(y)
			calculation x * 1000`

	_, err := parseAndAnalyze(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "undefined")
}

// ============================================================================
// Validation Error Cases — Missing Pattern or Calculation
// ============================================================================

// TestAnalyzeMeasureConversions_MissingPatternScope rejects a conversion
// without a pattern clause. Because the parser requires both pattern and
// calculation, we construct the symbol table manually.
func TestAnalyzeMeasureConversions_MissingPatternScope(t *testing.T) {
	t.Parallel()

	// Build symbol table manually
	root := symboltable.NewGlobalSymbolTable()
	measuresScope := root.AddChild(symboltable.MeasuresScope)

	// KmMeasure with a broken conversion (no PatternScope, only CalculationScope)
	kmScope := measuresScope.AddChild(symboltable.MeasureScope)
	n1 := "KmMeasure"
	n2 := "Number"
	kmScope.Define(&n1, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
	kmScope.Define(&n2, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 2})
	kmScope.DefineWithData(nil, symboltable.UnitReferenceDefinition, symboltable.SourceLocation{Line: 1, Column: 3}, parser.ParserUnitReference{Name: "MeterUnit"})
	convScope := kmScope.AddChild(symboltable.MeasureConversionScope)
	_ = convScope.AddChild(symboltable.CalculationScope) // only calculation, no pattern

	// Add a second valid measure
	mScope := measuresScope.AddChild(symboltable.MeasureScope)
	n3 := "MeterMeasure"
	mScope.Define(&n3, symboltable.Identifier, symboltable.SourceLocation{Line: 2, Column: 1})
	mScope.Define(&n2, symboltable.Identifier, symboltable.SourceLocation{Line: 2, Column: 2})
	mScope.DefineWithData(nil, symboltable.UnitReferenceDefinition, symboltable.SourceLocation{Line: 2, Column: 3}, parser.ParserUnitReference{Name: "MeterUnit"})

	// Walk with handlers to populate the analyzer
	analyzer := analysis.NewCPLAnalyzer()
	handlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MeasureScope: analysis.AnalyzeMeasureScope,
	}
	err := analysis.WalkSymbolTableWithHandlers(root, handlers, nil, analyzer)
	require.NoError(t, err, "initial walk should succeed")

	// Now simulate the post-processing step:
	// Walk MeasureScopes and call AnalyzeMeasureConversions
	var walkErr error
	symboltable.WalkSymbolTable(root, func(table *symboltable.SymbolTable) {
		if walkErr != nil {
			return
		}
		if table.ScopeType != symboltable.MeasureScope {
			return
		}
		if len(table.Definitions) < 2 {
			return
		}
		measureName := *table.Definitions[0].Name
		// Get the measure from the analyzer's map (we need a mutable copy)
		defs := analyzer.GetAnalysisDefinitions()
		measureDef, exists := defs.Measures[measureName]
		if !exists {
			return
		}
		if err := analyzer.AnalyzeMeasureConversions(table, &measureDef); err != nil {
			walkErr = err
			return
		}
	})
	require.Error(t, walkErr)
	assert.Contains(t, walkErr.Error(), "missing pattern scope")
}

// TestAnalyzeMeasureConversions_MissingCalculationScope rejects a conversion
// without a calculation clause.
func TestAnalyzeMeasureConversions_MissingCalculationScope(t *testing.T) {
	t.Parallel()

	// Build symbol table manually with missing CalculationScope
	root := symboltable.NewGlobalSymbolTable()
	measuresScope := root.AddChild(symboltable.MeasuresScope)

	kmScope := measuresScope.AddChild(symboltable.MeasureScope)
	n1 := "KmMeasure"
	n2 := "Number"
	kmScope.Define(&n1, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
	kmScope.Define(&n2, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 2})
	kmScope.DefineWithData(nil, symboltable.UnitReferenceDefinition, symboltable.SourceLocation{Line: 1, Column: 3}, parser.ParserUnitReference{Name: "MeterUnit"})
	convScope := kmScope.AddChild(symboltable.MeasureConversionScope)
	_ = convScope.AddChild(symboltable.PatternScope) // only pattern, no calculation

	mScope := measuresScope.AddChild(symboltable.MeasureScope)
	n3 := "MeterMeasure"
	mScope.Define(&n3, symboltable.Identifier, symboltable.SourceLocation{Line: 2, Column: 1})
	mScope.Define(&n2, symboltable.Identifier, symboltable.SourceLocation{Line: 2, Column: 2})
	mScope.DefineWithData(nil, symboltable.UnitReferenceDefinition, symboltable.SourceLocation{Line: 2, Column: 3}, parser.ParserUnitReference{Name: "MeterUnit"})

	analyzer := analysis.NewCPLAnalyzer()
	handlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MeasureScope: analysis.AnalyzeMeasureScope,
	}
	err := analysis.WalkSymbolTableWithHandlers(root, handlers, nil, analyzer)
	require.NoError(t, err)

	var walkErr error
	symboltable.WalkSymbolTable(root, func(table *symboltable.SymbolTable) {
		if walkErr != nil {
			return
		}
		if table.ScopeType != symboltable.MeasureScope {
			return
		}
		if len(table.Definitions) < 2 {
			return
		}
		measureName := *table.Definitions[0].Name
		defs := analyzer.GetAnalysisDefinitions()
		measureDef, exists := defs.Measures[measureName]
		if !exists {
			return
		}
		if err := analyzer.AnalyzeMeasureConversions(table, &measureDef); err != nil {
			walkErr = err
			return
		}
	})
	require.Error(t, walkErr)
	assert.Contains(t, walkErr.Error(), "missing calculation scope")
}

// TestAnalyze_MeasureConversions_Integration verifies that the full Analyze
// pipeline correctly processes measure scopes with conversions.
func TestAnalyze_MeasureConversions_Integration(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]

units
	Kilometer = Distance
	Meter = Distance

measures
	DistKm = Number(unit=Kilometer)
		conversion
			pattern DistKm(km) = DistM(m)
			calculation m * 1000
	DistM = Number(unit=Meter)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	// Verify DistKm has conversions (attached to the declaring measure)
	distKm, exists := defs.Measures["DistKm"]
	require.True(t, exists, "DistKm should be in measures map")
	require.Len(t, distKm.Conversions, 2, "Should have forward + inverse pairs")

	// Verify DistM has no conversions (conversions are on the declaring measure)
	distM, exists := defs.Measures["DistM"]
	require.True(t, exists, "DistM should be in measures map")
	assert.Empty(t, distM.Conversions, "DistM should have no conversions")
}

// TestAnalyze_MeasureConversions_CrossMeasure_Integration verifies that
// a conversion between measures with different underlying units works correctly.
func TestAnalyze_MeasureConversions_CrossMeasure_Integration(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]

units
	KilometerUnit = Distance
	MeterUnit = Distance

measures
	KmMeasure = Number(unit=KilometerUnit)
		conversion
			pattern KmMeasure(km) = MeterMeasure(m)
			calculation m * 1000
	MeterMeasure = Number(unit=MeterUnit)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)

	kmDef := defs.Measures["KmMeasure"]
	assert.Len(t, kmDef.Conversions, 2)

	// Verify source and target measures are correct
	forward := kmDef.Conversions[0]
	assert.Equal(t, "MeterMeasure", forward.SourceDef.Name)
	assert.Equal(t, "KmMeasure", forward.TargetDef.Name)
}
