// Package parser_test contains tests for the CPL parser units section.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findUnitScopeByUnitName finds a UnitScope child by its unit name (first identifier).
func findUnitScopeByUnitName(unitsScope *symboltable.SymbolTable, unitName string) *symboltable.SymbolTable {
	for _, child := range unitsScope.Children {
		if child.ScopeType == symboltable.UnitScope && len(child.Definitions) >= 1 {
			if child.Definitions[0].Name != nil && *child.Definitions[0].Name == unitName {
				return child
			}
		}
	}
	return nil
}

// TestParseSingleUnitDefinition tests parsing a single unit definition.
func TestParseSingleUnitDefinition(t *testing.T) {
	t.Parallel()

	input := "units\n\tCurrency = Country"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have a UnitsScope child
	require.Len(t, doc.Children, 1, "Expected one child scope (UnitsScope)")
	unitsScope := doc.Children[0]
	assert.Equal(t, symboltable.UnitsScope, unitsScope.ScopeType, "Child scope should be UnitsScope")

	require.Len(t, unitsScope.Children, 1, "Expected one child scope (UnitScope)")
	unitScope := unitsScope.Children[0]
	assert.Equal(t, symboltable.UnitScope, unitScope.ScopeType, "Child scope should be UnitScope")
	// Verify the unit definition was added to the UnitsScope
	require.Equal(t, len(unitScope.Definitions), 2, "Expected at one unit definition that names a unit and a concept that it's derived from")

	require.NotNil(t, unitScope.Definitions[0].Name, "Currency unit should be defined")
	assert.Equal(t, "Currency", *unitScope.Definitions[0].Name, "Unit name should be Currency")

	// Verify the ParserUnitDefinition data is attached
	require.NotNil(t, unitScope.Definitions[1].Name, "Concept name should be defined")
	assert.Equal(t, "Country", *unitScope.Definitions[1].Name, "Parser unit definition concept should be Country")
}

// TestParseMultipleUnitDefinitions tests parsing multiple unit definitions in one section.
func TestParseMultipleUnitDefinitions(t *testing.T) {
	t.Parallel()

	input := `units
	Currency = Country
	Weight = Mass
	Length = Distance`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have a UnitsScope child
	require.Len(t, doc.Children, 1, "Expected one child scope (UnitsScope)")
	unitsScope := doc.Children[0]
	assert.Equal(t, symboltable.UnitsScope, unitsScope.ScopeType, "Child scope should be UnitsScope")

	// Verify we have three UnitScope children
	require.Len(t, unitsScope.Children, 3, "Expected three UnitScope children")

	// Verify each UnitScope contains the correct identifiers
	for _, child := range unitsScope.Children {
		require.Len(t, child.Definitions, 2, "Each UnitScope should have two identifiers (unit name and concept name)")

		unitName := *child.Definitions[0].Name
		conceptName := *child.Definitions[1].Name

		switch unitName {
		case "Currency":
			assert.Equal(t, "Country", conceptName, "Currency should map to Country")
		case "Weight":
			assert.Equal(t, "Mass", conceptName, "Weight should map to Mass")
		case "Length":
			assert.Equal(t, "Distance", conceptName, "Length should map to Distance")
		default:
			t.Errorf("Unexpected unit name: %s", unitName)
		}
	}
}

// TestParseUnitsWithWhitespaceVariations tests parsing with various whitespace configurations.
func TestParseUnitsWithWhitespaceVariations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "tabs for indentation",
			input: "units\n\tCurrency = Country",
		},
		{
			name:  "spaces for indentation",
			input: "units\n    Currency = Country",
		},
		{
			name:  "multiple spaces around equals",
			input: "units\n\tCurrency  =  Country",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := parser.Parse(tc.input)

			require.NoError(t, err, "Expected no error for input: %s", tc.input)
			require.NotNil(t, doc, "Document should not be nil")

			// Verify UnitsScope was created
			require.Len(t, doc.Children, 1, "Expected one child scope")
		})
	}
}

// TestParseUnitsSourceLocation tests that source locations are captured correctly.
func TestParseUnitsSourceLocation(t *testing.T) {
	t.Parallel()

	input := `units
	Currency = Country
	Weight = Mass`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error")
	require.NotNil(t, doc, "Document should not be nil")

	// Get the UnitsScope
	unitsScope := doc.Children[0]

	// Verify source locations are present
	for _, def := range unitsScope.Definitions {
		assert.Greater(t, def.Location.Line, 0, "Source location line should be greater than 0")
		assert.GreaterOrEqual(t, def.Location.Column, 0, "Source location column should be >= 0")
	}
}

// TestParseUnitsScopeIntegration tests that UnitsScope is properly pushed/popped.
func TestParseUnitsScopeIntegration(t *testing.T) {
	t.Parallel()

	// Parse a simple units section with proper indentation
	input := "units\n\tCurrency = Country"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error")
	require.NotNil(t, doc, "Document should not be nil")

	// Verify UnitsScope was created
	require.Len(t, doc.Children, 1, "Expected one child scope")
	unitsScope := doc.Children[0]
	assert.Equal(t, symboltable.UnitsScope, unitsScope.ScopeType, "Child scope should be UnitsScope")

	// Verify the unitsScope parent is the GlobalScope
	assert.Equal(t, doc, unitsScope.Parent, "UnitsScope parent should be GlobalScope")
}

// TestParseUnitsSymbolTableIntegration tests that unit definitions are stored properly in symbol table.
func TestParseUnitsSymbolTableIntegration(t *testing.T) {
	t.Parallel()

	// Parse multiple unit definitions
	input := `units
	Currency = Country
	Weight = Mass`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error")
	require.NotNil(t, doc, "Document should not be nil")

	// Get the UnitsScope
	unitsScope := doc.Children[0]
	require.Equal(t, symboltable.UnitsScope, unitsScope.ScopeType)

	// Verify both UnitScope children exist
	require.Len(t, unitsScope.Children, 2, "Should have two UnitScope children")

	// Find Currency UnitScope
	currencyUnitScope := findUnitScopeByUnitName(unitsScope, "Currency")
	require.NotNil(t, currencyUnitScope, "Currency unit should be found in UnitsScope")
	require.Len(t, currencyUnitScope.Definitions, 2, "UnitScope should have two identifiers")
	assert.Equal(t, "Currency", *currencyUnitScope.Definitions[0].Name)

	// Find Weight UnitScope
	weightUnitScope := findUnitScopeByUnitName(unitsScope, "Weight")
	require.NotNil(t, weightUnitScope, "Weight unit should be found in UnitsScope")
	require.Len(t, weightUnitScope.Definitions, 2, "UnitScope should have two identifiers")
	assert.Equal(t, "Weight", *weightUnitScope.Definitions[0].Name)
}

// TestParseUnitsInvalidSyntax tests error handling for invalid syntax patterns in units section.
func TestParseUnitsInvalidSyntax(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name:          "missing equals sign",
			input:         "units\n\tCurrency Country",
			expectError:   true,
			errorContains: "missing '='",
		},
		{
			name:          "missing concept name",
			input:         "units\n\tCurrency =",
			expectError:   true,
			errorContains: "missing",
		},
		{
			name:          "invalid unit name identifier",
			input:         "units\n\t123invalid = Country",
			expectError:   true,
			errorContains: "extraneous input '123invalid'",
		},
		{
			name:          "invalid concept name identifier",
			input:         "units\n\tCurrency = 123invalid",
			expectError:   true,
			errorContains: "mismatched input '123invalid' expecting",
		},
		{
			name:          "extra equals sign",
			input:         "units\n\tCurrency == Country",
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

// TestParseUnitsRequiresAtLeastOneDefinition tests that units section requires at least one definition.
func TestParseUnitsRequiresAtLeastOneDefinition(t *testing.T) {
	t.Parallel()

	// The grammar requires unitDeclaration+, so at least one definition is required
	input := "units"

	_, err := parser.Parse(input)

	// Should fail because unitDeclaration+ requires at least one unit declaration
	require.Error(t, err, "Expected an error for units section without definitions")
}

// TestParseUnitsAndConceptsTogether tests parsing both concepts and units in the same document.
func TestParseUnitsAndConceptsTogether(t *testing.T) {
	t.Parallel()

	input := `concepts
	Country = [Values[USA, Canada, Mexico]]
	Mass = [Values[kg, lb]]

units
	Currency = Country
	Weight = Mass`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have both a ConceptsScope and UnitsScope
	require.Len(t, doc.Children, 2, "Expected two child scopes (ConceptsScope and UnitsScope)")

	// Find the ConceptsScope and UnitsScope
	var conceptsScope, unitsScope *symboltable.SymbolTable
	for _, child := range doc.Children {
		if child.ScopeType == symboltable.ConceptsScope {
			conceptsScope = child
		} else if child.ScopeType == symboltable.UnitsScope {
			unitsScope = child
		}
	}

	require.NotNil(t, conceptsScope, "ConceptsScope should exist")
	require.NotNil(t, unitsScope, "UnitsScope should exist")

	// Verify concepts are present
	countryFound := false
	for _, scope := range conceptsScope.Children {
		_, err := scope.Lookup("Country")
		if err == nil {
			countryFound = true
			break
		}
	}
	assert.True(t, countryFound, "Country concept should be found")

	// Verify units are present - find the UnitScope that contains Currency
	currencyUnitScope := findUnitScopeByUnitName(unitsScope, "Currency")
	require.NotNil(t, currencyUnitScope, "Currency unit should be found in UnitsScope")
	require.Len(t, currencyUnitScope.Definitions, 2, "UnitScope should have two identifiers")
	assert.Equal(t, "Country", *currencyUnitScope.Definitions[1].Name, "Currency should reference Country concept")
}

// TestParseUnitsAndTypesTogether tests parsing types and units in the same document.
func TestParseUnitsAndTypesTogether(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)
	Mass = Number()

units
	Weight = Mass`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have both a TypesScope and UnitsScope
	require.Len(t, doc.Children, 2, "Expected two child scopes (TypesScope and UnitsScope)")

	// Find the TypesScope and UnitsScope
	var typesScope, unitsScope *symboltable.SymbolTable
	for _, child := range doc.Children {
		if child.ScopeType == symboltable.TypesScope {
			typesScope = child
		} else if child.ScopeType == symboltable.UnitsScope {
			unitsScope = child
		}
	}

	require.NotNil(t, typesScope, "TypesScope should exist")
	require.NotNil(t, unitsScope, "UnitsScope should exist")

	// Verify types are present
	massFound := false
	for _, scope := range typesScope.Children {
		_, err := scope.Lookup("Mass")
		if err == nil {
			massFound = true
			break
		}
	}
	assert.True(t, massFound, "Mass type should be found")

	// Verify units are present - find the UnitScope that contains Weight
	weightUnitScope := findUnitScopeByUnitName(unitsScope, "Weight")
	require.NotNil(t, weightUnitScope, "Weight unit should be found in UnitsScope")
	require.Len(t, weightUnitScope.Definitions, 2, "UnitScope should have two identifiers")
	assert.Equal(t, "Mass", *weightUnitScope.Definitions[1].Name, "Weight should reference Mass")
}
