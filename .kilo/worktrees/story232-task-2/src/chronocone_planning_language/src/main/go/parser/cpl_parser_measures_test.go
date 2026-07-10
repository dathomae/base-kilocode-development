// Package parser_test contains tests for the CPL parser measures section.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findMeasureScopeByMeasureName finds a MeasureScope child by its measure name (first identifier).
func findMeasureScopeByMeasureName(measuresScope *symboltable.SymbolTable, measureName string) *symboltable.SymbolTable {
	for _, child := range measuresScope.Children {
		if child.ScopeType == symboltable.MeasureScope && len(child.Definitions) >= 1 {
			if child.Definitions[0].Name != nil && *child.Definitions[0].Name == measureName {
				return child
			}
		}
	}
	return nil
}

// TestParseSingleMeasureDefinition tests parsing a single measure definition.
func TestParseSingleMeasureDefinition(t *testing.T) {
	t.Parallel()

	input := "measures\n\tDistance = Length(unit=Meter)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have a MeasuresScope child
	require.Len(t, doc.Children, 1, "Expected one child scope (MeasuresScope)")
	measuresScope := doc.Children[0]
	assert.Equal(t, symboltable.MeasuresScope, measuresScope.ScopeType, "Child scope should be MeasuresScope")

	require.Len(t, measuresScope.Children, 1, "Expected one child scope (MeasureScope)")
	measureScope := measuresScope.Children[0]
	assert.Equal(t, symboltable.MeasureScope, measureScope.ScopeType, "Child scope should be MeasureScope")

	// Verify the measure definition was added to the MeasureScope
	// Measure definition has 2 identifiers: measure_name, type_name
	// Plus a UnitReferenceDefinition for the unit expression
	require.Equal(t, 3, len(measureScope.Definitions), "Expected three definitions: measure_name, type_name, unit_reference")

	require.NotNil(t, measureScope.Definitions[0].Name, "Measure name should be defined")
	assert.Equal(t, "Distance", *measureScope.Definitions[0].Name, "Measure name should be Distance")

	require.NotNil(t, measureScope.Definitions[1].Name, "Type name should be defined")
	assert.Equal(t, "Length", *measureScope.Definitions[1].Name, "Type name should be Length")

	require.Equal(t, symboltable.UnitReferenceDefinition, measureScope.Definitions[2].DefinitionType, "Third definition should be UnitReferenceDefinition")
	unitRef, ok := measureScope.Definitions[2].Data.(parser.ParserUnitReference)
	require.True(t, ok, "Third definition data should be ParserUnitReference")
	assert.Equal(t, "Meter", unitRef.Name, "Unit reference should be Meter")

	// unit= now routes through unitParameter, not constraint, so no ConstraintScope child
	require.Len(t, measureScope.Children, 0, "Expected no ConstraintScope children for unit-only measure")
}

// TestParseMultipleMeasureDefinitions tests parsing multiple measure definitions in one section.
func TestParseMultipleMeasureDefinitions(t *testing.T) {
	t.Parallel()

	input := `measures
	Distance = Length(unit=Meter)
	Weight = Mass(unit=Kilogram)
	Price = Currency(unit=Dollar)`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have a MeasuresScope child
	require.Len(t, doc.Children, 1, "Expected one child scope (MeasuresScope)")
	measuresScope := doc.Children[0]
	assert.Equal(t, symboltable.MeasuresScope, measuresScope.ScopeType, "Child scope should be MeasuresScope")

	// Verify we have three MeasureScope children
	require.Len(t, measuresScope.Children, 3, "Expected three MeasureScope children")

	// Verify each MeasureScope contains the correct identifiers and unit reference
	for _, child := range measuresScope.Children {
		require.Len(t, child.Definitions, 3, "Each MeasureScope should have three definitions (measure name, type name, unit reference)")

		measureName := *child.Definitions[0].Name
		typeName := *child.Definitions[1].Name

		switch measureName {
		case "Distance":
			assert.Equal(t, "Length", typeName, "Distance should have type Length")
			require.Equal(t, symboltable.UnitReferenceDefinition, child.Definitions[2].DefinitionType)
			unitRef, ok := child.Definitions[2].Data.(parser.ParserUnitReference)
			require.True(t, ok)
			assert.Equal(t, "Meter", unitRef.Name)
			require.Len(t, child.Children, 0, "Distance should have no ConstraintScope children")
		case "Weight":
			assert.Equal(t, "Mass", typeName, "Weight should have type Mass")
			require.Equal(t, symboltable.UnitReferenceDefinition, child.Definitions[2].DefinitionType)
			unitRef, ok := child.Definitions[2].Data.(parser.ParserUnitReference)
			require.True(t, ok)
			assert.Equal(t, "Kilogram", unitRef.Name)
			require.Len(t, child.Children, 0, "Weight should have no ConstraintScope children")
		case "Price":
			assert.Equal(t, "Currency", typeName, "Price should have type Currency")
			require.Equal(t, symboltable.UnitReferenceDefinition, child.Definitions[2].DefinitionType)
			unitRef, ok := child.Definitions[2].Data.(parser.ParserUnitReference)
			require.True(t, ok)
			assert.Equal(t, "Dollar", unitRef.Name)
			require.Len(t, child.Children, 0, "Price should have no ConstraintScope children")
		default:
			t.Errorf("Unexpected measure name: %s", measureName)
		}
	}
}

// TestParseMeasuresWithWhitespaceVariations tests parsing with various whitespace configurations.
func TestParseMeasuresWithWhitespaceVariations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "tabs for indentation",
			input: "measures\n\tDistance = Length(unit=Meter)",
		},
		{
			name:  "spaces for indentation",
			input: "measures\n    Distance = Length(unit=Meter)",
		},
		{
			name:  "multiple spaces around equals",
			input: "measures\n\tDistance  =  Length(unit=Meter)",
		},
		{
			name:  "no spaces around equals",
			input: "measures\n\tDistance=Length(unit=Meter)",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := parser.Parse(tc.input)

			require.NoError(t, err, "Expected no error for input: %s", tc.input)
			require.NotNil(t, doc, "Document should not be nil")

			// Verify MeasuresScope was created
			require.Len(t, doc.Children, 1, "Expected one child scope")
		})
	}
}

// TestParseMeasuresSourceLocation tests that source locations are captured correctly.
func TestParseMeasuresSourceLocation(t *testing.T) {
	t.Parallel()

	input := `measures
	Distance = Length(unit=Meter)
	Weight = Mass(unit=Kilogram)`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error")
	require.NotNil(t, doc, "Document should not be nil")

	// Get the MeasuresScope
	measuresScope := doc.Children[0]

	// Verify source locations are present on MeasureScope children
	for _, measureScope := range measuresScope.Children {
		for _, def := range measureScope.Definitions {
			assert.Greater(t, def.Location.Line, 0, "Source location line should be greater than 0")
			assert.GreaterOrEqual(t, def.Location.Column, 0, "Source location column should be >= 0")
		}
	}
}

// TestParseMeasuresScopeIntegration tests that MeasuresScope is properly pushed/popped.
func TestParseMeasuresScopeIntegration(t *testing.T) {
	t.Parallel()

	// Parse a simple measures section with proper indentation
	input := "measures\n\tDistance = Length(unit=Meter)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error")
	require.NotNil(t, doc, "Document should not be nil")

	// Verify MeasuresScope was created
	require.Len(t, doc.Children, 1, "Expected one child scope")
	measuresScope := doc.Children[0]
	assert.Equal(t, symboltable.MeasuresScope, measuresScope.ScopeType, "Child scope should be MeasuresScope")

	// Verify the measuresScope parent is the GlobalScope
	assert.Equal(t, doc, measuresScope.Parent, "MeasuresScope parent should be GlobalScope")
}

// TestParseMeasuresSymbolTableIntegration tests that measure definitions are stored properly in symbol table.
func TestParseMeasuresSymbolTableIntegration(t *testing.T) {
	t.Parallel()

	// Parse multiple measure definitions
	input := `measures
	Distance = Length(unit=Meter)
	Weight = Mass(unit=Kilogram)`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error")
	require.NotNil(t, doc, "Document should not be nil")

	// Get the MeasuresScope
	measuresScope := doc.Children[0]
	require.Equal(t, symboltable.MeasuresScope, measuresScope.ScopeType)

	// Verify both MeasureScope children exist
	require.Len(t, measuresScope.Children, 2, "Should have two MeasureScope children")

	// Find Distance MeasureScope
	distanceMeasureScope := findMeasureScopeByMeasureName(measuresScope, "Distance")
	require.NotNil(t, distanceMeasureScope, "Distance measure should be found in MeasuresScope")
	require.Len(t, distanceMeasureScope.Definitions, 3, "MeasureScope should have three definitions")
	assert.Equal(t, "Distance", *distanceMeasureScope.Definitions[0].Name)
	assert.Equal(t, "Length", *distanceMeasureScope.Definitions[1].Name)
	require.Equal(t, symboltable.UnitReferenceDefinition, distanceMeasureScope.Definitions[2].DefinitionType)
	unitRef, ok := distanceMeasureScope.Definitions[2].Data.(parser.ParserUnitReference)
	require.True(t, ok)
	assert.Equal(t, "Meter", unitRef.Name)
	require.Len(t, distanceMeasureScope.Children, 0, "MeasureScope should have no ConstraintScope children")

	// Find Weight MeasureScope
	weightMeasureScope := findMeasureScopeByMeasureName(measuresScope, "Weight")
	require.NotNil(t, weightMeasureScope, "Weight measure should be found in MeasuresScope")
	require.Len(t, weightMeasureScope.Definitions, 3, "MeasureScope should have three definitions")
	assert.Equal(t, "Weight", *weightMeasureScope.Definitions[0].Name)
	assert.Equal(t, "Mass", *weightMeasureScope.Definitions[1].Name)
	require.Equal(t, symboltable.UnitReferenceDefinition, weightMeasureScope.Definitions[2].DefinitionType)
	unitRef, ok = weightMeasureScope.Definitions[2].Data.(parser.ParserUnitReference)
	require.True(t, ok)
	assert.Equal(t, "Kilogram", unitRef.Name)
	require.Len(t, weightMeasureScope.Children, 0, "MeasureScope should have no ConstraintScope children")
}

// TestParseMeasuresInvalidSyntax tests error handling for invalid syntax patterns in measures section.
func TestParseMeasuresInvalidSyntax(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name:          "missing equals sign",
			input:         "measures\n\tDistance Length(unit=Meter)",
			expectError:   true,
			errorContains: "missing '='",
		},
		{
			name:          "missing unit keyword",
			input:         "measures\n\tDistance = Length(Meter)",
			expectError:   true,
			errorContains: "expecting '='",
		},
		{
			name:          "missing opening paren",
			input:         "measures\n\tDistance = Length unit=Meter)",
			expectError:   true,
			errorContains: "missing '('",
		},
		{
			name:          "missing closing paren",
			input:         "measures\n\tDistance = Length(unit=Meter",
			expectError:   true,
			errorContains: "missing ')'",
		},
		{
			name:          "invalid measure name identifier",
			input:         "measures\n\t123invalid = Length(unit=Meter)",
			expectError:   true,
			errorContains: "extraneous input '123invalid'",
		},
		{
			name:          "extra equals sign",
			input:         "measures\n\tDistance == Length(unit=Meter)",
			expectError:   true,
			errorContains: "mismatched input '=='",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := parser.Parse(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected an error for input: %s", tc.input)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, "Error message should contain expected text")
				}
				return
			}

			require.NoError(t, err, "Expected no error for input: %s", tc.input)
		})
	}
}

// TestParseMeasuresRequiresAtLeastOneDefinition tests that measures section requires at least one definition.
func TestParseMeasuresRequiresAtLeastOneDefinition(t *testing.T) {
	t.Parallel()

	// The grammar requires measureDeclaration+, so at least one definition is required
	input := "measures"

	_, err := parser.Parse(input)

	// Should fail because measureDeclaration+ requires at least one measure declaration
	require.Error(t, err, "Expected an error for measures section without definitions")
}

// TestParseMeasuresWithTypesAndUnitsTogether tests parsing types, units, and measures in the same document.
func TestParseMeasuresWithTypesAndUnitsTogether(t *testing.T) {
	t.Parallel()

	input := `types
	Length = Number(precision=2)
	Mass = Number()

units
	Meter = Length
	Kilogram = Mass

measures
	Distance = Length(unit=Meter)
	Weight = Mass(unit=Kilogram)`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have TypesScope, UnitsScope, and MeasuresScope
	require.Len(t, doc.Children, 3, "Expected three child scopes (TypesScope, UnitsScope, MeasuresScope)")

	// Find the MeasuresScope
	var measuresScope *symboltable.SymbolTable
	for _, child := range doc.Children {
		if child.ScopeType == symboltable.MeasuresScope {
			measuresScope = child
			break
		}
	}

	require.NotNil(t, measuresScope, "MeasuresScope should exist")

	// Verify measures are present
	distanceMeasureScope := findMeasureScopeByMeasureName(measuresScope, "Distance")
	require.NotNil(t, distanceMeasureScope, "Distance measure should be found in MeasuresScope")
	require.Len(t, distanceMeasureScope.Definitions, 3, "MeasureScope should have three definitions")
	assert.Equal(t, "Distance", *distanceMeasureScope.Definitions[0].Name, "Measure name should be Distance")
	assert.Equal(t, "Length", *distanceMeasureScope.Definitions[1].Name, "Type name should be Length")
	require.Equal(t, symboltable.UnitReferenceDefinition, distanceMeasureScope.Definitions[2].DefinitionType)
	unitRef, ok := distanceMeasureScope.Definitions[2].Data.(parser.ParserUnitReference)
	require.True(t, ok)
	assert.Equal(t, "Meter", unitRef.Name)
	require.Len(t, distanceMeasureScope.Children, 0, "MeasureScope should have no ConstraintScope children")
}

// TestParseMeasureWithDivisionUnit tests parsing a measure with a division compound unit.
func TestParseMeasureWithDivisionUnit(t *testing.T) {
	t.Parallel()

	input := "measures\n\tSpeed = Number(unit=Length/Time)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	measuresScope := doc.Children[0]
	require.Equal(t, symboltable.MeasuresScope, measuresScope.ScopeType)

	speedScope := findMeasureScopeByMeasureName(measuresScope, "Speed")
	require.NotNil(t, speedScope, "Speed measure should be found")

	// Definitions: Speed, Number, Length(UnitRef), Time(UnitRef), DivideOperator
	require.Len(t, speedScope.Definitions, 5, "Expected 5 definitions: measure_name, type_name, Length, Time, /")
	assert.Equal(t, "Speed", *speedScope.Definitions[0].Name)
	assert.Equal(t, "Number", *speedScope.Definitions[1].Name)

	require.Equal(t, symboltable.UnitReferenceDefinition, speedScope.Definitions[2].DefinitionType)
	ref, ok := speedScope.Definitions[2].Data.(parser.ParserUnitReference)
	require.True(t, ok)
	assert.Equal(t, "Length", ref.Name)

	require.Equal(t, symboltable.UnitReferenceDefinition, speedScope.Definitions[3].DefinitionType)
	ref, ok = speedScope.Definitions[3].Data.(parser.ParserUnitReference)
	require.True(t, ok)
	assert.Equal(t, "Time", ref.Name)

	require.Equal(t, symboltable.DivideOperatorDefinition, speedScope.Definitions[4].DefinitionType)
	require.Len(t, speedScope.Children, 0, "Expected no ConstraintScope children")
}

// TestParseMeasureWithPowerUnit tests parsing a measure with a power compound unit.
func TestParseMeasureWithPowerUnit(t *testing.T) {
	t.Parallel()

	input := "measures\n\tAcceleration = Number(unit=Length/Time^2)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	measuresScope := doc.Children[0]
	require.Equal(t, symboltable.MeasuresScope, measuresScope.ScopeType)

	accelScope := findMeasureScopeByMeasureName(measuresScope, "Acceleration")
	require.NotNil(t, accelScope, "Acceleration measure should be found")

	// Definitions: Acceleration, Number, Length(UnitRef), Time(UnitRef), 2(Number), ^(Power), /(Divide)
	require.Len(t, accelScope.Definitions, 7, "Expected 7 definitions for Length/Time^2 postfix")
	assert.Equal(t, "Acceleration", *accelScope.Definitions[0].Name)
	assert.Equal(t, "Number", *accelScope.Definitions[1].Name)

	require.Equal(t, symboltable.UnitReferenceDefinition, accelScope.Definitions[2].DefinitionType)
	ref, ok := accelScope.Definitions[2].Data.(parser.ParserUnitReference)
	require.True(t, ok)
	assert.Equal(t, "Length", ref.Name)

	require.Equal(t, symboltable.UnitReferenceDefinition, accelScope.Definitions[3].DefinitionType)
	ref, ok = accelScope.Definitions[3].Data.(parser.ParserUnitReference)
	require.True(t, ok)
	assert.Equal(t, "Time", ref.Name)

	require.Equal(t, symboltable.NumberDefinition, accelScope.Definitions[4].DefinitionType)
	numData, ok := accelScope.Definitions[4].Data.(symboltable.FloatData)
	require.True(t, ok)
	assert.Equal(t, float64(2), numData.Value)

	require.Equal(t, symboltable.PowerOperatorDefinition, accelScope.Definitions[5].DefinitionType)
	powOp, ok := accelScope.Definitions[5].Data.(parser.ParserOperator)
	require.True(t, ok)
	assert.Equal(t, parser.OpPower, powOp.Operator)

	require.Equal(t, symboltable.DivideOperatorDefinition, accelScope.Definitions[6].DefinitionType)
	divOp, ok := accelScope.Definitions[6].Data.(parser.ParserOperator)
	require.True(t, ok)
	assert.Equal(t, parser.OpDivide, divOp.Operator)

	require.Len(t, accelScope.Children, 0, "Expected no ConstraintScope children")
}

// TestParseMeasureWithMultiplyDivideUnit tests parsing a measure with multiplication and division.
func TestParseMeasureWithMultiplyDivideUnit(t *testing.T) {
	t.Parallel()

	input := "measures\n\tMomentum = Number(unit=Mass*Length/Time)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	measuresScope := doc.Children[0]
	require.Equal(t, symboltable.MeasuresScope, measuresScope.ScopeType)

	momentumScope := findMeasureScopeByMeasureName(measuresScope, "Momentum")
	require.NotNil(t, momentumScope, "Momentum measure should be found")

	// Postfix for (Mass*Length)/Time: Mass, Length, *, Time, /
	// Definitions: Momentum, Number, Mass(UnitRef), Length(UnitRef), *(Multiply), Time(UnitRef), /(Divide)
	require.Len(t, momentumScope.Definitions, 7, "Expected 7 definitions for Mass*Length/Time postfix")
	assert.Equal(t, "Momentum", *momentumScope.Definitions[0].Name)
	assert.Equal(t, "Number", *momentumScope.Definitions[1].Name)

	require.Equal(t, symboltable.UnitReferenceDefinition, momentumScope.Definitions[2].DefinitionType)
	ref, ok := momentumScope.Definitions[2].Data.(parser.ParserUnitReference)
	require.True(t, ok)
	assert.Equal(t, "Mass", ref.Name)

	require.Equal(t, symboltable.UnitReferenceDefinition, momentumScope.Definitions[3].DefinitionType)
	ref, ok = momentumScope.Definitions[3].Data.(parser.ParserUnitReference)
	require.True(t, ok)
	assert.Equal(t, "Length", ref.Name)

	require.Equal(t, symboltable.MultiplyOperatorDefinition, momentumScope.Definitions[4].DefinitionType)
	mulOp, ok := momentumScope.Definitions[4].Data.(parser.ParserOperator)
	require.True(t, ok)
	assert.Equal(t, parser.OpMultiply, mulOp.Operator)

	require.Equal(t, symboltable.UnitReferenceDefinition, momentumScope.Definitions[5].DefinitionType)
	ref, ok = momentumScope.Definitions[5].Data.(parser.ParserUnitReference)
	require.True(t, ok)
	assert.Equal(t, "Time", ref.Name)

	require.Equal(t, symboltable.DivideOperatorDefinition, momentumScope.Definitions[6].DefinitionType)
	divOp, ok := momentumScope.Definitions[6].Data.(parser.ParserOperator)
	require.True(t, ok)
	assert.Equal(t, parser.OpDivide, divOp.Operator)

	require.Len(t, momentumScope.Children, 0, "Expected no ConstraintScope children")
}

// TestParseMeasureWithParenthesizedUnit tests parsing a measure with parenthesized compound unit.
func TestParseMeasureWithParenthesizedUnit(t *testing.T) {
	t.Parallel()

	input := "measures\n\tAreaDensity = Number(unit=Mass/(Length^2))"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	measuresScope := doc.Children[0]
	require.Equal(t, symboltable.MeasuresScope, measuresScope.ScopeType)

	areaDensityScope := findMeasureScopeByMeasureName(measuresScope, "AreaDensity")
	require.NotNil(t, areaDensityScope, "AreaDensity measure should be found")

	// Postfix for Mass/(Length^2): Mass, Length, 2, ^, /
	// Definitions: AreaDensity, Number, Mass(UnitRef), Length(UnitRef), 2(Number), ^(Power), /(Divide)
	require.Len(t, areaDensityScope.Definitions, 7, "Expected 7 definitions for Mass/(Length^2) postfix")
	assert.Equal(t, "AreaDensity", *areaDensityScope.Definitions[0].Name)
	assert.Equal(t, "Number", *areaDensityScope.Definitions[1].Name)

	require.Equal(t, symboltable.UnitReferenceDefinition, areaDensityScope.Definitions[2].DefinitionType)
	ref, ok := areaDensityScope.Definitions[2].Data.(parser.ParserUnitReference)
	require.True(t, ok)
	assert.Equal(t, "Mass", ref.Name)

	require.Equal(t, symboltable.UnitReferenceDefinition, areaDensityScope.Definitions[3].DefinitionType)
	ref, ok = areaDensityScope.Definitions[3].Data.(parser.ParserUnitReference)
	require.True(t, ok)
	assert.Equal(t, "Length", ref.Name)

	require.Equal(t, symboltable.NumberDefinition, areaDensityScope.Definitions[4].DefinitionType)
	numData, ok := areaDensityScope.Definitions[4].Data.(symboltable.FloatData)
	require.True(t, ok)
	assert.Equal(t, float64(2), numData.Value)

	require.Equal(t, symboltable.PowerOperatorDefinition, areaDensityScope.Definitions[5].DefinitionType)
	powOp, ok := areaDensityScope.Definitions[5].Data.(parser.ParserOperator)
	require.True(t, ok)
	assert.Equal(t, parser.OpPower, powOp.Operator)

	require.Equal(t, symboltable.DivideOperatorDefinition, areaDensityScope.Definitions[6].DefinitionType)
	divOp, ok := areaDensityScope.Definitions[6].Data.(parser.ParserOperator)
	require.True(t, ok)
	assert.Equal(t, parser.OpDivide, divOp.Operator)

	require.Len(t, areaDensityScope.Children, 0, "Expected no ConstraintScope children")
}

// TestParseMeasureWithConstraintAndUnit tests parsing a measure with both constraint and unit parameter.
func TestParseMeasureWithConstraintAndUnit(t *testing.T) {
	t.Parallel()

	input := "measures\n\tIntervalMeasure = Interval(starts_at=Midnight, unit=Duration)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	measuresScope := doc.Children[0]
	require.Equal(t, symboltable.MeasuresScope, measuresScope.ScopeType)

	intervalScope := findMeasureScopeByMeasureName(measuresScope, "IntervalMeasure")
	require.NotNil(t, intervalScope, "IntervalMeasure should be found")

	// Definitions: IntervalMeasure, Interval, Duration(UnitRef)
	require.Len(t, intervalScope.Definitions, 3, "Expected 3 definitions")
	assert.Equal(t, "IntervalMeasure", *intervalScope.Definitions[0].Name)
	assert.Equal(t, "Interval", *intervalScope.Definitions[1].Name)
	require.Equal(t, symboltable.UnitReferenceDefinition, intervalScope.Definitions[2].DefinitionType)
	ref, ok := intervalScope.Definitions[2].Data.(parser.ParserUnitReference)
	require.True(t, ok)
	assert.Equal(t, "Duration", ref.Name)

	// starts_at constraint should produce a ConstraintScope child
	require.Len(t, intervalScope.Children, 1, "Expected one ConstraintScope child for starts_at")
	constraintScope := intervalScope.Children[0]
	assert.Equal(t, symboltable.ConstraintScope, constraintScope.ScopeType)

	// ConstraintScope should have the constraint name and value
	require.GreaterOrEqual(t, len(constraintScope.Definitions), 1, "ConstraintScope should have definitions")
}
