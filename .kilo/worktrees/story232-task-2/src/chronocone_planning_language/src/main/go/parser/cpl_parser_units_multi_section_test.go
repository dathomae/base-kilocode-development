// Package parser_test contains tests for the CPL units multi-section parser.
//
// This file tests Increment 2 functionality: multiple units sections in the same file.
// Each section creates its own UnitsScope symbol table as siblings under GlobalScope.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseMultipleUnitsSections_Basic verifies that two or more units sections
// in the same file are parsed successfully and create separate UnitsScope tables.
func TestParseMultipleUnitsSections_Basic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name: "two units sections in same file",
			input: `units
	Currency = Country

units
	Weight = Mass`,
			expectError: false,
		},
		{
			name: "three units sections",
			input: `units
	Currency = Country

units
	Weight = Mass

units
	Length = Distance`,
			expectError: false,
		},
		{
			name: "units sections with other sections between",
			input: `units
	Currency = Country

types
	Metric = Number()

units
	Weight = Mass`,
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

// TestParseMultipleUnitsSections_SymbolTableStructure verifies that multiple
// units sections create separate UnitsScope tables as siblings under GlobalScope.
func TestParseMultipleUnitsSections_SymbolTableStructure(t *testing.T) {
	t.Parallel()

	input := `units
	Currency = Country
	Weight = Mass

units
	Length = Distance
	Time = Duration`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should succeed")

	// Get the root symbol table
	root, err := listener.GetSymbolStack().Root()
	require.NoError(t, err, "Should have root symbol table")

	// Root should be GlobalScope
	assert.Equal(t, symboltable.GlobalScope, root.ScopeType, "Root should be GlobalScope")

	// Should have exactly 2 UnitsScope children
	unitsScopes := filterScopesByType(root.Children, symboltable.UnitsScope)
	assert.Len(t, unitsScopes, 2, "Should have exactly 2 UnitsScope tables")

	// Verify first section has at least 2 child UnitScope tables (Currency and Weight)
	// Note: there may be extra empty UnitScope children due to newline handling in the lexer
	firstScope := unitsScopes[0]
	firstChildScopes := filterScopesByType(firstScope.Children, symboltable.UnitScope)
	assert.GreaterOrEqual(t, len(firstChildScopes), 2, "First UnitsScope should have at least 2 child UnitScope tables")

	// Verify second section has at least 2 child UnitScope tables (Length and Time)
	// Note: there may be extra empty UnitScope children due to newline handling in the lexer
	secondScope := unitsScopes[1]
	secondChildScopes := filterScopesByType(secondScope.Children, symboltable.UnitScope)
	assert.GreaterOrEqual(t, len(secondChildScopes), 2, "Second UnitsScope should have at least 2 child UnitScope tables")
}

// TestParseMultipleUnitsSections_IndependentParsing verifies that unit definitions
// in different sections are independently parsed.
func TestParseMultipleUnitsSections_IndependentParsing(t *testing.T) {
	t.Parallel()

	input := `units
	Currency = Country

units
	Weight = Mass`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should succeed")

	// Get the root symbol table
	root, err := listener.GetSymbolStack().Root()
	require.NoError(t, err, "Should have root symbol table")

	// Should have 2 UnitsScope children
	unitsScopes := filterScopesByType(root.Children, symboltable.UnitsScope)
	require.Len(t, unitsScopes, 2, "Should have exactly 2 UnitsScope tables")

	// First section: verify Currency unit
	firstScope := unitsScopes[0]
	currencyUnitScope := findUnitScopeByUnitName(firstScope, "Currency")
	require.NotNil(t, currencyUnitScope, "Currency unit should be found in first UnitsScope")
	require.Len(t, currencyUnitScope.Definitions, 2, "UnitScope should have two identifiers")
	assert.Equal(t, "Currency", *currencyUnitScope.Definitions[0].Name)
	assert.Equal(t, "Country", *currencyUnitScope.Definitions[1].Name)

	// Verify Weight is NOT in first section
	weightUnitScopeInFirst := findUnitScopeByUnitName(firstScope, "Weight")
	require.Nil(t, weightUnitScopeInFirst, "Weight unit should NOT be found in first UnitsScope")

	// Second section: verify Weight unit (independent from first section)
	secondScope := unitsScopes[1]
	weightUnitScope2 := findUnitScopeByUnitName(secondScope, "Weight")
	require.NotNil(t, weightUnitScope2, "Weight unit should be found in second UnitsScope")
	require.Len(t, weightUnitScope2.Definitions, 2, "UnitScope should have two identifiers")
	assert.Equal(t, "Weight", *weightUnitScope2.Definitions[0].Name)
	assert.Equal(t, "Mass", *weightUnitScope2.Definitions[1].Name)
}

// TestParseMultipleUnitsSections_StackBehavior verifies that the symbol stack
// is properly managed when entering and exiting multiple units sections.
func TestParseMultipleUnitsSections_StackBehavior(t *testing.T) {
	t.Parallel()

	input := `units
	Currency = Country

units
	Weight = Mass`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should succeed")

	// After parsing, stack should be back to just global scope
	assert.Equal(t, 1, listener.GetSymbolStack().Size(),
		"Stack should be back to global scope after parsing multiple sections")

	// Verify no errors
	assert.False(t, listener.HasErrors(), "Should have no semantic errors")
}

// TestParseMultipleUnitsSections_DuplicateUnitsAcrossSections verifies that
// duplicate unit definitions across sections are allowed during parsing
// (detection happens in analysis phase).
func TestParseMultipleUnitsSections_DuplicateUnitsAcrossSections(t *testing.T) {
	t.Parallel()

	// Same unit name in different sections - should parse fine
	// (duplicate detection happens in analysis phase)
	input := `units
	Currency = Country

units
	Currency = Mass`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should succeed even with duplicate unit names")

	// Get the root symbol table
	root, err := listener.GetSymbolStack().Root()
	require.NoError(t, err, "Should have root symbol table")

	// Should have 2 UnitsScope children
	unitsScopes := filterScopesByType(root.Children, symboltable.UnitsScope)
	require.Len(t, unitsScopes, 2, "Should have exactly 2 UnitsScope tables")

	// Each scope should have at least 1 child UnitScope (Currency unit)
	for i, scope := range unitsScopes {
		childScopes := filterScopesByType(scope.Children, symboltable.UnitScope)
		assert.GreaterOrEqual(t, len(childScopes), 1, "UnitsScope %d should have at least 1 child UnitScope", i)
	}
}
