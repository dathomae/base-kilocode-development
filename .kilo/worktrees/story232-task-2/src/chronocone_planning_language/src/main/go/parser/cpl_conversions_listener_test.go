// Package parser_test contains tests for the CPL conversion listener methods.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Helper functions
// ============================================================================

// findChildScopeByType returns the first child of the given scope with the
// specified ScopeType, or nil if not found.
func findChildScopeByType(scope *symboltable.SymbolTable, scopeType symboltable.ScopeType) *symboltable.SymbolTable {
	for _, child := range scope.Children {
		if child.ScopeType == scopeType {
			return child
		}
	}
	return nil
}

// requireUnitReferenceName asserts that the definition at the given index is a
// UnitReferenceDefinition with the expected name in its Data.
func requireUnitReferenceName(t *testing.T, defs []symboltable.SymbolDefinition, index int, expectedName string) {
	t.Helper()
	require.Greater(t, len(defs), index, "Expected at least %d definitions", index+1)
	require.Equal(t, symboltable.UnitReferenceDefinition, defs[index].DefinitionType,
		"Definition at index %d should be UnitReferenceDefinition", index)
	unitRef, ok := defs[index].Data.(parser.ParserUnitReference)
	require.True(t, ok, "Definition data at index %d should be ParserUnitReference", index)
	assert.Equal(t, expectedName, unitRef.Name, "Unit reference name at index %d should match", index)
}

// requireIdentifierName asserts that the definition at the given index is an
// Identifier with the expected name.
func requireIdentifierName(t *testing.T, defs []symboltable.SymbolDefinition, index int, expectedName string) {
	t.Helper()
	require.Greater(t, len(defs), index, "Expected at least %d definitions", index+1)
	require.Equal(t, symboltable.Identifier, defs[index].DefinitionType,
		"Definition at index %d should be Identifier", index)
	require.NotNil(t, defs[index].Name, "Identifier name at index %d should not be nil", index)
	assert.Equal(t, expectedName, *defs[index].Name, "Identifier name at index %d should match", index)
}

// requireNumberLiteral asserts that the definition at the given index is a
// NumberDefinition with the expected value in its Data.
func requireNumberLiteral(t *testing.T, defs []symboltable.SymbolDefinition, index int, expectedValue float64) {
	t.Helper()
	require.Greater(t, len(defs), index, "Expected at least %d definitions", index+1)
	require.Equal(t, symboltable.NumberDefinition, defs[index].DefinitionType,
		"Definition at index %d should be NumberDefinition", index)
	numData, ok := defs[index].Data.(symboltable.FloatData)
	require.True(t, ok, "Definition data at index %d should be FloatData", index)
	assert.InDelta(t, expectedValue, numData.Value, 0.001, "Number value at index %d should match", index)
}

// requireOperatorDefinition asserts that the definition at the given index is
// the expected operator definition type.
func requireOperatorDefinition(t *testing.T, defs []symboltable.SymbolDefinition, index int, expectedType symboltable.DefinitionType) {
	t.Helper()
	require.Greater(t, len(defs), index, "Expected at least %d definitions", index+1)
	assert.Equal(t, expectedType, defs[index].DefinitionType,
		"Definition at index %d should be %s", index, expectedType)
}

// ============================================================================
// Unit Conversion Listener Tests
// ============================================================================

// TestUnitConversionScopeTree verifies that parsing a unit with a conversion
// clause produces the correct symbol table tree structure:
//   UnitsScope → UnitScope → ConversionScope → PatternScope + CalculationScope
func TestUnitConversionScopeTree(t *testing.T) {
	t.Parallel()

	input := "units\n\tLength = Distance\n\t\tconversion\n\t\t\tpattern Kilometer(km) = Meter(m)\n\t\t\tcalculation km * 1000"

	doc, err := parser.Parse(input)
	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify global → UnitsScope
	require.Len(t, doc.Children, 1, "Expected one child scope (UnitsScope)")
	unitsScope := doc.Children[0]
	assert.Equal(t, symboltable.UnitsScope, unitsScope.ScopeType)

	// Verify UnitsScope → UnitScope
	require.Len(t, unitsScope.Children, 1, "Expected one UnitScope child")
	unitScope := unitsScope.Children[0]
	assert.Equal(t, symboltable.UnitScope, unitScope.ScopeType)

	// Verify UnitScope has the unit and concept identifiers
	require.Len(t, unitScope.Definitions, 2, "UnitScope should have 2 identifier definitions")
	requireIdentifierName(t, unitScope.Definitions, 0, "Length")
	requireIdentifierName(t, unitScope.Definitions, 1, "Distance")

	// Verify UnitScope → ConversionScope
	conversionScope := findChildScopeByType(unitScope, symboltable.ConversionScope)
	require.NotNil(t, conversionScope, "UnitScope should have a ConversionScope child")
	assert.Empty(t, conversionScope.Definitions, "ConversionScope should have no direct definitions")

	// Verify ConversionScope → PatternScope + CalculationScope
	require.Len(t, conversionScope.Children, 2, "ConversionScope should have 2 children")
	patternScope := findChildScopeByType(conversionScope, symboltable.PatternScope)
	calculationScope := findChildScopeByType(conversionScope, symboltable.CalculationScope)
	require.NotNil(t, patternScope, "ConversionScope should have a PatternScope child")
	require.NotNil(t, calculationScope, "ConversionScope should have a CalculationScope child")
}

// TestUnitConversionPatternDefinitions verifies that pattern expressions and
// bound variables are captured in postfix order within PatternScope.
func TestUnitConversionPatternDefinitions(t *testing.T) {
	t.Parallel()

	input := "units\n\tLength = Distance\n\t\tconversion\n\t\t\tpattern Kilometer(km) = Meter(m)\n\t\t\tcalculation km * 1000"

	doc, err := parser.Parse(input)
	require.NoError(t, err)
	require.NotNil(t, doc)

	unitsScope := doc.Children[0]
	unitScope := unitsScope.Children[0]
	conversionScope := findChildScopeByType(unitScope, symboltable.ConversionScope)
	patternScope := findChildScopeByType(conversionScope, symboltable.PatternScope)

	// Pattern: Kilometer(km) = Meter(m)
	// Postfix order: Kilometer, km, Meter, m
	require.Len(t, patternScope.Definitions, 4, "PatternScope should have 4 postfix definitions")
	requireUnitReferenceName(t, patternScope.Definitions, 0, "Kilometer")
	requireIdentifierName(t, patternScope.Definitions, 1, "km")
	requireUnitReferenceName(t, patternScope.Definitions, 2, "Meter")
	requireIdentifierName(t, patternScope.Definitions, 3, "m")
}

// TestUnitConversionCalculationDefinitions verifies that calculation
// expressions are captured in postfix order within CalculationScope.
func TestUnitConversionCalculationDefinitions(t *testing.T) {
	t.Parallel()

	input := "units\n\tLength = Distance\n\t\tconversion\n\t\t\tpattern Kilometer(km) = Meter(m)\n\t\t\tcalculation km * 1000"

	doc, err := parser.Parse(input)
	require.NoError(t, err)
	require.NotNil(t, doc)

	unitsScope := doc.Children[0]
	unitScope := unitsScope.Children[0]
	conversionScope := findChildScopeByType(unitScope, symboltable.ConversionScope)
	calculationScope := findChildScopeByType(conversionScope, symboltable.CalculationScope)

	// Calculation: km * 1000
	// Postfix order: km, 1000, *
	require.Len(t, calculationScope.Definitions, 3, "CalculationScope should have 3 postfix definitions")
	requireUnitReferenceName(t, calculationScope.Definitions, 0, "km")
	requireNumberLiteral(t, calculationScope.Definitions, 1, 1000)
	requireOperatorDefinition(t, calculationScope.Definitions, 2, symboltable.MultiplyOperatorDefinition)
}

// TestUnitConversionMultiplePairs verifies that a conversion clause with
// multiple conversion pairs produces the correct scope structure.
func TestUnitConversionMultiplePairs(t *testing.T) {
	t.Parallel()

	input := `units
	Length = Distance
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation km * 1000
			pattern Meter(m) = Kilometer(km)
			calculation m / 1000`

	doc, err := parser.Parse(input)
	require.NoError(t, err)
	require.NotNil(t, doc)

	unitsScope := doc.Children[0]
	unitScope := unitsScope.Children[0]
	conversionScope := findChildScopeByType(unitScope, symboltable.ConversionScope)

	// Two pairs → two PatternScopes + two CalculationScopes
	require.Len(t, conversionScope.Children, 4, "ConversionScope should have 4 children (2 patterns + 2 calculations)")

	patternScopes := 0
	calculationScopes := 0
	for _, child := range conversionScope.Children {
		switch child.ScopeType {
		case symboltable.PatternScope:
			patternScopes++
		case symboltable.CalculationScope:
			calculationScopes++
		}
	}
	assert.Equal(t, 2, patternScopes, "Expected 2 PatternScope children")
	assert.Equal(t, 2, calculationScopes, "Expected 2 CalculationScope children")
}

// ============================================================================
// Measure Conversion Listener Tests
// ============================================================================

// TestMeasureConversionScopeTree verifies that parsing a measure with a
// conversion clause produces the correct symbol table tree structure:
//   MeasuresScope → MeasureScope → MeasureConversionScope → PatternScope + CalculationScope
func TestMeasureConversionScopeTree(t *testing.T) {
	t.Parallel()

	input := "measures\n\tDistance = Length(unit=Meter)\n\t\tconversion\n\t\t\tpattern Kilometer(km) = Meter(m)\n\t\t\tcalculation km * 1000"

	doc, err := parser.Parse(input)
	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify global → MeasuresScope
	require.Len(t, doc.Children, 1, "Expected one child scope (MeasuresScope)")
	measuresScope := doc.Children[0]
	assert.Equal(t, symboltable.MeasuresScope, measuresScope.ScopeType)

	// Verify MeasuresScope → MeasureScope
	require.Len(t, measuresScope.Children, 1, "Expected one MeasureScope child")
	measureScope := measuresScope.Children[0]
	assert.Equal(t, symboltable.MeasureScope, measureScope.ScopeType)

	// Verify MeasureScope has measure name, type name, and unit reference
	require.Len(t, measureScope.Definitions, 3, "MeasureScope should have 3 definitions")
	requireIdentifierName(t, measureScope.Definitions, 0, "Distance")
	requireIdentifierName(t, measureScope.Definitions, 1, "Length")
	require.Equal(t, symboltable.UnitReferenceDefinition, measureScope.Definitions[2].DefinitionType)

	// Verify MeasureScope → MeasureConversionScope
	measureConversionScope := findChildScopeByType(measureScope, symboltable.MeasureConversionScope)
	require.NotNil(t, measureConversionScope, "MeasureScope should have a MeasureConversionScope child")
	assert.Empty(t, measureConversionScope.Definitions, "MeasureConversionScope should have no direct definitions")

	// Verify MeasureConversionScope → PatternScope + CalculationScope
	require.Len(t, measureConversionScope.Children, 2, "MeasureConversionScope should have 2 children")
	patternScope := findChildScopeByType(measureConversionScope, symboltable.PatternScope)
	calculationScope := findChildScopeByType(measureConversionScope, symboltable.CalculationScope)
	require.NotNil(t, patternScope, "MeasureConversionScope should have a PatternScope child")
	require.NotNil(t, calculationScope, "MeasureConversionScope should have a CalculationScope child")
}

// TestMeasureConversionPatternDefinitions verifies that pattern expressions
// and bound variables are captured in postfix order within PatternScope for
// measure conversions.
func TestMeasureConversionPatternDefinitions(t *testing.T) {
	t.Parallel()

	input := "measures\n\tDistance = Length(unit=Meter)\n\t\tconversion\n\t\t\tpattern Kilometer(km) = Meter(m)\n\t\t\tcalculation km * 1000"

	doc, err := parser.Parse(input)
	require.NoError(t, err)
	require.NotNil(t, doc)

	measuresScope := doc.Children[0]
	measureScope := measuresScope.Children[0]
	measureConversionScope := findChildScopeByType(measureScope, symboltable.MeasureConversionScope)
	patternScope := findChildScopeByType(measureConversionScope, symboltable.PatternScope)

	// Pattern: Kilometer(km) = Meter(m)
	require.Len(t, patternScope.Definitions, 4, "PatternScope should have 4 postfix definitions")
	requireUnitReferenceName(t, patternScope.Definitions, 0, "Kilometer")
	requireIdentifierName(t, patternScope.Definitions, 1, "km")
	requireUnitReferenceName(t, patternScope.Definitions, 2, "Meter")
	requireIdentifierName(t, patternScope.Definitions, 3, "m")
}

// TestMeasureConversionCalculationDefinitions verifies that calculation
// expressions are captured in postfix order within CalculationScope for
// measure conversions.
func TestMeasureConversionCalculationDefinitions(t *testing.T) {
	t.Parallel()

	input := "measures\n\tDistance = Length(unit=Meter)\n\t\tconversion\n\t\t\tpattern Kilometer(km) = Meter(m)\n\t\t\tcalculation km * 1000"

	doc, err := parser.Parse(input)
	require.NoError(t, err)
	require.NotNil(t, doc)

	measuresScope := doc.Children[0]
	measureScope := measuresScope.Children[0]
	measureConversionScope := findChildScopeByType(measureScope, symboltable.MeasureConversionScope)
	calculationScope := findChildScopeByType(measureConversionScope, symboltable.CalculationScope)

	// Calculation: km * 1000
	require.Len(t, calculationScope.Definitions, 3, "CalculationScope should have 3 postfix definitions")
	requireUnitReferenceName(t, calculationScope.Definitions, 0, "km")
	requireNumberLiteral(t, calculationScope.Definitions, 1, 1000)
	requireOperatorDefinition(t, calculationScope.Definitions, 2, symboltable.MultiplyOperatorDefinition)
}

// ============================================================================
// Complex Expression Tests
// ============================================================================

// TestUnitConversionComplexCalculation verifies that more complex arithmetic
// expressions in calculation clauses produce the correct postfix sequence.
func TestUnitConversionComplexCalculation(t *testing.T) {
	t.Parallel()

	input := "units\n\tTemperature = Heat\n\t\tconversion\n\t\t\tpattern Celsius(c) = Fahrenheit(f)\n\t\t\tcalculation c * 9 / 5 + 32"

	doc, err := parser.Parse(input)
	require.NoError(t, err)
	require.NotNil(t, doc)

	unitsScope := doc.Children[0]
	unitScope := unitsScope.Children[0]
	conversionScope := findChildScopeByType(unitScope, symboltable.ConversionScope)
	calculationScope := findChildScopeByType(conversionScope, symboltable.CalculationScope)

	// Calculation: c * 9 / 5 + 32
	// Postfix order: c, 9, *, 5, /, 32, +
	require.Len(t, calculationScope.Definitions, 7, "CalculationScope should have 7 postfix definitions")
	requireUnitReferenceName(t, calculationScope.Definitions, 0, "c")
	requireNumberLiteral(t, calculationScope.Definitions, 1, 9)
	requireOperatorDefinition(t, calculationScope.Definitions, 2, symboltable.MultiplyOperatorDefinition)
	requireNumberLiteral(t, calculationScope.Definitions, 3, 5)
	requireOperatorDefinition(t, calculationScope.Definitions, 4, symboltable.DivideOperatorDefinition)
	requireNumberLiteral(t, calculationScope.Definitions, 5, 32)
	requireOperatorDefinition(t, calculationScope.Definitions, 6, symboltable.PlusOperatorDefinition)
}

// TestUnitConversionWithoutConversionClause verifies that a unit without a
// conversion clause does not produce a ConversionScope child.
func TestUnitConversionWithoutConversionClause(t *testing.T) {
	t.Parallel()

	input := "units\n\tCurrency = Country"

	doc, err := parser.Parse(input)
	require.NoError(t, err)
	require.NotNil(t, doc)

	unitsScope := doc.Children[0]
	unitScope := unitsScope.Children[0]
	conversionScope := findChildScopeByType(unitScope, symboltable.ConversionScope)
	assert.Nil(t, conversionScope, "UnitScope without conversion should not have a ConversionScope child")
}
