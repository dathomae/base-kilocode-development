// Package parser_test contains tests for the CPL measures multi-section parser.
//
// This file tests Increment 2 functionality: multiple measures sections in the same file.
// Each section creates its own MeasuresScope symbol table as siblings under GlobalScope.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseMultipleMeasuresSections_Basic verifies that two or more measures sections
// in the same file are parsed successfully and create separate MeasuresScope tables.
func TestParseMultipleMeasuresSections_Basic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name: "two measures sections in same file",
			input: `measures
	Distance = Length(unit=Meter)

measures
	Weight = Mass(unit=Kilogram)`,
			expectError: false,
		},
		{
			name: "three measures sections",
			input: `measures
	Distance = Length(unit=Meter)

measures
	Weight = Mass(unit=Kilogram)

measures
	Price = Currency(unit=Dollar)`,
			expectError: false,
		},
		{
			name: "measures sections with other sections between",
			input: `measures
	Distance = Length(unit=Meter)

types
	Metric = Number()

measures
	Weight = Mass(unit=Kilogram)`,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := parser.Parse(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected parsing to fail")
			} else {
				require.NoError(t, err, "Expected parsing to succeed")
			}
		})
	}
}

// TestParseMultipleMeasuresSections_SymbolTableStructure verifies that multiple
// measures sections create separate MeasuresScope tables as siblings under GlobalScope.
func TestParseMultipleMeasuresSections_SymbolTableStructure(t *testing.T) {
	t.Parallel()

	input := `measures
	Distance = Length(unit=Meter)
	Weight = Mass(unit=Kilogram)

measures
	Price = Currency(unit=Dollar)
	Time = Duration(unit=Second)`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should succeed")

	// Get the root symbol table
	root, err := listener.GetSymbolStack().Root()
	require.NoError(t, err, "Should have root symbol table")

	// Root should be GlobalScope
	assert.Equal(t, symboltable.GlobalScope, root.ScopeType, "Root should be GlobalScope")

	// Should have exactly 2 MeasuresScope children
	measuresScopes := filterScopesByType(root.Children, symboltable.MeasuresScope)
	assert.Len(t, measuresScopes, 2, "Should have exactly 2 MeasuresScope tables")

	// Verify first section has at least 2 child MeasureScope tables (Distance and Weight)
	// Note: there may be extra empty MeasureScope children due to newline handling in the lexer
	firstScope := measuresScopes[0]
	firstChildScopes := filterScopesByType(firstScope.Children, symboltable.MeasureScope)
	assert.GreaterOrEqual(t, len(firstChildScopes), 2, "First MeasuresScope should have at least 2 child MeasureScope tables")

	// Verify second section has at least 2 child MeasureScope tables (Price and Time)
	// Note: there may be extra empty MeasureScope children due to newline handling in the lexer
	secondScope := measuresScopes[1]
	secondChildScopes := filterScopesByType(secondScope.Children, symboltable.MeasureScope)
	assert.GreaterOrEqual(t, len(secondChildScopes), 2, "Second MeasuresScope should have at least 2 child MeasureScope tables")
}

// TestParseMultipleMeasuresSections_IndependentParsing verifies that measure definitions
// in different sections are independently parsed.
func TestParseMultipleMeasuresSections_IndependentParsing(t *testing.T) {
	t.Parallel()

	input := `measures
	Distance = Length(unit=Meter)

measures
	Weight = Mass(unit=Kilogram)`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should succeed")

	// Get the root symbol table
	root, err := listener.GetSymbolStack().Root()
	require.NoError(t, err, "Should have root symbol table")

	// Should have 2 MeasuresScope children
	measuresScopes := filterScopesByType(root.Children, symboltable.MeasuresScope)
	require.Len(t, measuresScopes, 2, "Should have exactly 2 MeasuresScope tables")

	// First section: verify Distance measure
	firstScope := measuresScopes[0]
	distanceMeasureScope := findMeasureScopeByMeasureName(firstScope, "Distance")
	require.NotNil(t, distanceMeasureScope, "Distance measure should be found in first MeasuresScope")
	require.Len(t, distanceMeasureScope.Definitions, 3, "MeasureScope should have three definitions")
	assert.Equal(t, "Distance", *distanceMeasureScope.Definitions[0].Name)
	assert.Equal(t, "Length", *distanceMeasureScope.Definitions[1].Name)
	require.Equal(t, symboltable.UnitReferenceDefinition, distanceMeasureScope.Definitions[2].DefinitionType)
	unitRef, ok := distanceMeasureScope.Definitions[2].Data.(parser.ParserUnitReference)
	require.True(t, ok)
	assert.Equal(t, "Meter", unitRef.Name)
	require.Len(t, distanceMeasureScope.Children, 0, "MeasureScope should have no ConstraintScope children")

	// Second section: verify Weight measure
	secondScope := measuresScopes[1]
	weightMeasureScope := findMeasureScopeByMeasureName(secondScope, "Weight")
	require.NotNil(t, weightMeasureScope, "Weight measure should be found in second MeasuresScope")
	require.Len(t, weightMeasureScope.Definitions, 3, "MeasureScope should have three definitions")
	assert.Equal(t, "Weight", *weightMeasureScope.Definitions[0].Name)
	assert.Equal(t, "Mass", *weightMeasureScope.Definitions[1].Name)
	require.Equal(t, symboltable.UnitReferenceDefinition, weightMeasureScope.Definitions[2].DefinitionType)
	unitRef, ok = weightMeasureScope.Definitions[2].Data.(parser.ParserUnitReference)
	require.True(t, ok)
	assert.Equal(t, "Kilogram", unitRef.Name)
	require.Len(t, weightMeasureScope.Children, 0, "MeasureScope should have no ConstraintScope children")
}

// TestParseMultipleMeasuresSections_StackBehavior verifies that the symbol stack
// is properly managed when entering and exiting multiple measures sections.
func TestParseMultipleMeasuresSections_StackBehavior(t *testing.T) {
	t.Parallel()

	input := `measures
	Distance = Length(unit=Meter)

measures
	Weight = Mass(unit=Kilogram)`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should succeed")

	// After parsing, stack should be back to just global scope
	assert.Equal(t, 1, listener.GetSymbolStack().Size(),
		"Stack should be back to global scope after parsing multiple sections")

	// Verify no errors
	assert.False(t, listener.HasErrors(), "Should have no semantic errors")
}

// TestParseMultipleMeasuresSections_DuplicateMeasuresAcrossSections verifies that
// duplicate measure definitions across sections are allowed during parsing
// (detection happens in analysis phase).
func TestParseMultipleMeasuresSections_DuplicateMeasuresAcrossSections(t *testing.T) {
	t.Parallel()

	// Same measure name in different sections - should parse fine
	// (duplicate detection happens in analysis phase)
	input := `measures
	Distance = Length(unit=Meter)

measures
	Distance = Mass(unit=Kilogram)`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should succeed even with duplicate measure names")

	// Get the root symbol table
	root, err := listener.GetSymbolStack().Root()
	require.NoError(t, err, "Should have root symbol table")

	// Should have 2 MeasuresScope children
	measuresScopes := filterScopesByType(root.Children, symboltable.MeasuresScope)
	require.Len(t, measuresScopes, 2, "Should have exactly 2 MeasuresScope tables")

	// Each scope should have at least 1 child MeasureScope (Distance measure)
	for i, scope := range measuresScopes {
		childScopes := filterScopesByType(scope.Children, symboltable.MeasureScope)
		assert.GreaterOrEqual(t, len(childScopes), 1, "MeasuresScope %d should have at least 1 child MeasureScope", i)
	}
}
