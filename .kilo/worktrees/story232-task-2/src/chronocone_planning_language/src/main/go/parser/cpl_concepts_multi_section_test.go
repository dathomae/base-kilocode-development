// Package parser_test contains tests for the CPL concepts multi-section parser.
//
// This file tests Increment 3 functionality: multiple concepts sections in the same file.
// Each section creates its own ConceptsScope symbol table as siblings under GlobalScope.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseMultipleConceptsSections_Basic verifies that two concepts sections
// in the same file are parsed successfully and create separate ConceptsScope tables.
func TestParseMultipleConceptsSections_Basic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name: "two concepts sections in same file",
			input: `concepts
	Country=[Values[USA, Canada, Mexico]]

concepts
	Priority=[Values[Low, Medium, High]]`,
			expectError: false,
		},
		{
			name: "three concepts sections",
			input: `concepts
	Status=[Values[Active, Inactive]]

concepts
	Type=[Values[A, B, C]]

concepts
	Category=[Values[X, Y]]`,
			expectError: false,
		},
		{
			name: "concepts sections with other sections between",
			input: `concepts
	Country=[Values[USA]]

types
	Metric = Number()

concepts
	Priority=[Values[Low, High]]`,
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

// TestParseMultipleConceptsSections_SymbolTableStructure verifies that multiple
// concepts sections create separate ConceptsScope tables as siblings under GlobalScope.
func TestParseMultipleConceptsSections_SymbolTableStructure(t *testing.T) {
	t.Parallel()

	input := `concepts
	Country=[Values[USA, Canada]]
	Region=[Values[North, South]]

concepts
	Priority=[Values[Low, Medium, High]]
	Status=[Values[Active, Inactive]]`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should succeed")

	// Get the root symbol table
	root, err := listener.GetSymbolStack().Root()
	require.NoError(t, err, "Should have root symbol table")

	// Root should be GlobalScope
	assert.Equal(t, symboltable.GlobalScope, root.ScopeType, "Root should be GlobalScope")

	// Should have exactly 2 ConceptsScope children
	conceptsScopes := filterScopesByType(root.Children, symboltable.ConceptsScope)
	assert.Len(t, conceptsScopes, 2, "Should have exactly 2 ConceptsScope tables")

	// Verify first section has at least 2 child ConceptScope tables (Country and Region)
	// Note: there may be more children due to ValueGeneratorScope tables
	firstScope := conceptsScopes[0]
	firstChildScopes := filterScopesByType(firstScope.Children, symboltable.ConceptScope)
	assert.GreaterOrEqual(t, len(firstChildScopes), 2, "First ConceptsScope should have at least 2 child ConceptScope tables")

	// Verify second section has at least 2 child ConceptScope tables (Priority and Status)
	secondScope := conceptsScopes[1]
	secondChildScopes := filterScopesByType(secondScope.Children, symboltable.ConceptScope)
	assert.GreaterOrEqual(t, len(secondChildScopes), 2, "Second ConceptsScope should have at least 2 child ConceptScope tables")
}

// TestParseMultipleConceptsSections_Hierarchy verifies that hierarchy within
// each concepts section is preserved when there are multiple sections.
func TestParseMultipleConceptsSections_Hierarchy(t *testing.T) {
	t.Parallel()

	input := `concepts
	Country=[Values[USA, Canada]]
		Region=[Values[North, South]]

concepts
	Product=[Values[Widget, Gadget]]
		Variant=[Values[A, B]]`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should succeed")

	// Get the root symbol table
	root, err := listener.GetSymbolStack().Root()
	require.NoError(t, err, "Should have root symbol table")

	// Should have 2 ConceptsScope children
	conceptsScopes := filterScopesByType(root.Children, symboltable.ConceptsScope)
	require.Len(t, conceptsScopes, 2, "Should have exactly 2 ConceptsScope tables")

	// Verify first section has parent-child hierarchy
	firstScope := conceptsScopes[0]
	assert.Equal(t, symboltable.ConceptsScope, firstScope.ScopeType)
	// First scope should have children (for the nested Region concept)
	assert.NotEmpty(t, firstScope.Children, "First scope should have child scopes for hierarchy")

	// Verify second section has parent-child hierarchy
	secondScope := conceptsScopes[1]
	assert.Equal(t, symboltable.ConceptsScope, secondScope.ScopeType)
	// Second scope should have children (for the nested Variant concept)
	assert.NotEmpty(t, secondScope.Children, "Second scope should have child scopes for hierarchy")
}

// TestParseMultipleConceptsSections_EmptySection verifies behavior with empty concepts sections.
// Note: Empty section validation is deferred to the analysis phase.
func TestParseMultipleConceptsSections_EmptySection(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name: "empty concepts section followed by valid section",
			input: `concepts

concepts
	Country=[Values[USA]]`,
			expectError: false, // Parser accepts this; validation in analysis phase
		},
		{
			name: "valid section followed by empty section",
			input: `concepts
	Country=[Values[USA]]

concepts`,
			expectError: true, // Empty concepts section at EOF fails
		},
		{
			name: "two empty concepts sections",
			input: `concepts

concepts`,
			expectError: true, // Empty concepts sections fail
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := parser.Parse(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected parsing to fail for empty concepts section")
			} else {
				require.NoError(t, err, "Expected parsing to succeed")
			}
		})
	}
}

// TestParseMultipleConceptsSections_StackBehavior verifies that the symbol stack
// is properly managed when entering and exiting multiple concepts sections.
func TestParseMultipleConceptsSections_StackBehavior(t *testing.T) {
	t.Parallel()

	input := `concepts
	Country=[Values[USA]]

concepts
	Priority=[Values[Low]]`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should succeed")

	// After parsing, stack should be back to just global scope
	assert.Equal(t, 1, listener.GetSymbolStack().Size(),
		"Stack should be back to global scope after parsing multiple sections")

	// Verify no errors
	assert.False(t, listener.HasErrors(), "Should have no semantic errors")
}

// TestParseMultipleConceptsSections_DuplicateConceptsAcrossSections verifies that
// duplicate concept definitions across sections are allowed during parsing
// (detection happens in analysis phase).
func TestParseMultipleConceptsSections_DuplicateConceptsAcrossSections(t *testing.T) {
	t.Parallel()

	// Same concept name in different sections - should parse fine
	// (duplicate detection happens in analysis phase)
	input := `concepts
	Status=[Values[Active, Inactive]]

concepts
	Status=[Values[Pending, Completed]]`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should succeed even with duplicate concept names")

	// Get the root symbol table
	root, err := listener.GetSymbolStack().Root()
	require.NoError(t, err, "Should have root symbol table")

	// Should have 2 ConceptsScope children
	conceptsScopes := filterScopesByType(root.Children, symboltable.ConceptsScope)
	require.Len(t, conceptsScopes, 2, "Should have exactly 2 ConceptsScope tables")

	// Each scope should have at least 1 child ConceptScope (Status concept)
	// Note: there may be more children due to ValueGeneratorScope tables
	for i, scope := range conceptsScopes {
		childScopes := filterScopesByType(scope.Children, symboltable.ConceptScope)
		assert.GreaterOrEqual(t, len(childScopes), 1, "ConceptsScope %d should have at least 1 child ConceptScope", i)
	}
}

// Helper functions

// filterScopesByType returns a slice of symbol tables matching the given scope type.
func filterScopesByType(scopes []*symboltable.SymbolTable, scopeType symboltable.ScopeType) []*symboltable.SymbolTable {
	var result []*symboltable.SymbolTable
	for _, scope := range scopes {
		if scope.ScopeType == scopeType {
			result = append(result, scope)
		}
	}
	return result
}

// filterDefinitionsByType returns a slice of symbol definitions matching the given definition type.
func filterDefinitionsByType(defs []symboltable.SymbolDefinition, defType symboltable.DefinitionType) []symboltable.SymbolDefinition {
	var result []symboltable.SymbolDefinition
	for _, def := range defs {
		if def.DefinitionType == defType {
			result = append(result, def)
		}
	}
	return result
}

// getConceptNames returns the names of all concept definitions (identifiers) in the scope.
func getConceptNames(scope *symboltable.SymbolTable) []string {
	var names []string
	for _, def := range scope.Definitions {
		if def.DefinitionType == symboltable.Identifier && def.Name != nil {
			names = append(names, *def.Name)
		}
	}
	return names
}
