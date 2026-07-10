// Package parser_test contains tests for the CPL concepts conditional specification parser.
//
// This file tests Increment 4 functionality: conditional concept value specifications.
// Conditional specs define child concept values that depend on specific parent values
// using the syntax: Parent[Value] => Child=[Values[...]]
//
// Phase B (Logical Tests) - Test structure with skeleton assertions.
// Phase C will add deep symbol table structure verification.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/ast"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/require"
)

// TestConditionalSpec_SimpleParsing verifies that a basic conditional specification
// parses without syntax errors. A conditional spec defines child concept values
// that are available only when a specific parent value is selected.
//
// Example: Country[USA] => State=[Values[NY, CA, TX]]
// This means "State" has values NY, CA, TX when "Country" is USA.
func TestConditionalSpec_SimpleParsing(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name: "simple conditional spec with Values generator",
			input: `concepts
	Country[USA] => State=[Values[NY, CA, TX]]`,
			expectError: false,
		},
		{
			name: "conditional spec with single value",
			input: `concepts
	Country[USA] => State=[Values[NY]]`,
			expectError: false,
		},
		{
			name: "conditional spec with integer values",
			input: `concepts
	Year[2024] => Quarter=[Values[Q1, Q2, Q3, Q4]]`,
			expectError: false,
		},
		{
			name: "concept definition and conditional spec for same parent name",
			input: `concepts
	Country=[Values[USA, Canada]]
	Country[USA] => Region=[Values[North, South]]`,
			expectError: false,
		},
		{
			name: "regional concept selection via parent value",
			input: `concepts
	Country=[Values[USA, Canada]]
	Country[USA] => Region=[Values[North, South, East, West]]
	Country[Canada] => Region=[Values[Atlantic, Pacific, Central]]`,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootTable, err := parser.Parse(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected parsing to fail for input: %s", tc.input)
				return
			}

			require.NoError(t, err, "Expected parsing to succeed for input: %s", tc.input)
			require.NotNil(t, rootTable, "Root symbol table should not be nil")

			// Phase C: Verify ConditionalSpec data exists in the symbol table
			// Phase C: Navigate to ConceptsScope -> ConceptScope and verify
			// ConceptDefinitionData.ConditionalSpecs contains the expected parent value and child values
		})
	}
}

// TestConditionalSpec_DataLocation verifies that ConditionalSpec data is stored
// in the child ConceptScope's ConceptDefinitionData.ConditionalSpecs slice,
// not in the ConditionalSpecScope's Data field.
//
// Symbol table traversal path:
// RootTable → ConceptsScope → ConceptScope ("Country") → ConditionalSpecScope → ConceptScope ("State")
func TestConditionalSpec_DataLocation(t *testing.T) {
	t.Parallel()

	input := `concepts
	Country=[Values[USA, Canada]]
	Country[USA] => State=[Values[NY, CA, TX]]`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Expected parsing to succeed")
	require.NotNil(t, rootTable, "Root symbol table should not be nil")

	// Navigate: RootTable → ConceptsScope
	require.NotEmpty(t, rootTable.Children, "Root should have children")
	var conceptsScope *symboltable.SymbolTable
	for _, child := range rootTable.Children {
		if child.ScopeType == symboltable.ConceptsScope {
			conceptsScope = child
			break
		}
	}
	require.NotNil(t, conceptsScope, "Should find ConceptsScope")

	// Navigate: ConceptsScope → ConceptScope ("Country")
	require.NotEmpty(t, conceptsScope.Children, "ConceptsScope should have children")
	var countryScope *symboltable.SymbolTable
	for _, child := range conceptsScope.Children {
		if child.ScopeType == symboltable.ConceptScope {
			for _, cchild := range child.Children {
				if cchild.ScopeType == symboltable.ConditionalSpecScope {
					countryScope = child
					break
				}
			}
			if countryScope != nil {
				break
			}
		}
	}
	require.NotNil(t, countryScope, "Should find ConceptScope for Country containing ConditionalSpecScope")

	// Navigate: ConceptScope ("Country") → ConditionalSpecScope
	require.NotEmpty(t, countryScope.Children, "Country ConceptScope should have children")
	var conditionalSpecScope *symboltable.SymbolTable
	for _, child := range countryScope.Children {
		if child.ScopeType == symboltable.ConditionalSpecScope {
			conditionalSpecScope = child
			break
		}
	}
	require.NotNil(t, conditionalSpecScope, "Should find ConditionalSpecScope under Country ConceptScope")

	// Navigate: ConditionalSpecScope → ConceptScope ("State")
	require.NotEmpty(t, conditionalSpecScope.Children, "ConditionalSpecScope should have children")
	var stateScope *symboltable.SymbolTable
	for _, child := range conditionalSpecScope.Children {
		if child.ScopeType == symboltable.ConceptScope {
			stateScope = child
			break
		}
	}
	require.NotNil(t, stateScope, "Should find child ConceptScope (State) under ConditionalSpecScope")

	// Assert: child ConceptScope Data is *ast.ConceptDefinitionData with correct ConditionalSpecs
	require.NotNil(t, stateScope.Data, "Child ConceptScope Data should not be nil")
	conceptData, ok := stateScope.Data.(*ast.ConceptDefinitionData)
	require.True(t, ok, "Child ConceptScope Data should be *ast.ConceptDefinitionData")
	require.Len(t, conceptData.ConditionalSpecs, 1, "Should have exactly one ConditionalSpec")
	conditionalSpec := conceptData.ConditionalSpecs[0]
	require.Equal(t, "USA", conditionalSpec.ParentValue, "ParentValue should be USA")
	require.Equal(t, []string{"NY", "CA", "TX"}, conditionalSpec.Values, "Values should be NY, CA, TX")

	// Assert: ConditionalSpecScope.Data does NOT contain *ast.ConditionalSpec
	if conditionalSpecScope.Data != nil {
		_, isConditionalSpec := conditionalSpecScope.Data.(*ast.ConditionalSpec)
		require.False(t, isConditionalSpec, "ConditionalSpecScope.Data should not be *ast.ConditionalSpec")
	}
}

// TestConditionalSpec_MultipleConditions verifies that multiple conditional
// specifications can be parsed for the same or different child concepts.
// Each condition maps a different parent value to a different set of child values.
//
// Example:
//
//	Country[USA] => State=[Values[NY, CA, TX]]
//	Country[Canada] => Province=[Values[Ontario, Quebec]]
func TestConditionalSpec_MultipleConditions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name: "multiple conditions for different parent values",
			input: `concepts
	Country[USA] => State=[Values[NY, CA, TX]]
	Country[Canada] => Province=[Values[Ontario, Quebec]]`,
			expectError: false,
		},
		{
			name: "multiple conditions for same child concept",
			input: `concepts
	Country[USA] => Region=[Values[North, South, East, West]]
	Country[Canada] => Region=[Values[Atlantic, Pacific, Central]]`,
			expectError: false,
		},
		{
			name: "three conditions for same parent concept",
			input: `concepts
	Country[USA] => State=[Values[NY, CA]]
	Country[Canada] => Province=[Values[Ontario, Quebec]]
	Country[Mexico] => State=[Values[Jalisco, Oaxaca]]`,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootTable, err := parser.Parse(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected parsing to fail for input: %s", tc.input)
				return
			}

			require.NoError(t, err, "Expected parsing to succeed for input: %s", tc.input)
			require.NotNil(t, rootTable, "Root symbol table should not be nil")

			// Phase C: Verify all ConditionalSpecs are stored in the appropriate ConceptScope
			// Phase C: Verify each ConditionalSpec has correct ParentValue and Values slice
		})
	}
}

// TestConditionalSpec_WithHierarchy verifies that conditional specifications
// can be parsed alongside parent-child concept hierarchies. The parent concept
// is defined with Values, and conditional specs define child values based on
// specific parent values.
//
// Example:
//
//	Country=[Values[USA, Canada]]
//	  Country[USA] => State=[Values[NY, CA]]
func TestConditionalSpec_WithHierarchy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name: "conditional spec with parent concept defined above",
			input: `concepts
	Country=[Values[USA, Canada]]
		Country[USA] => State=[Values[NY, CA]]`,
			expectError: false,
		},
		{
			name: "conditional spec with deeply nested hierarchy",
			input: `concepts
	Country=[Values[USA, Canada]]
		Region=[Values[North, South]]
			Country[USA] => State=[Values[NY, CA]]`,
			expectError: false,
		},
		{
			name: "multiple conditional specs under same parent hierarchy",
			input: `concepts
	Country=[Values[USA, Canada]]
		Country[USA] => State=[Values[NY, CA, TX]]
		Country[Canada] => Province=[Values[Ontario, Quebec]]`,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootTable, err := parser.Parse(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected parsing to fail for input: %s", tc.input)
				return
			}

			require.NoError(t, err, "Expected parsing to succeed for input: %s", tc.input)
			require.NotNil(t, rootTable, "Root symbol table should not be nil")

			// Phase C: Verify hierarchy is preserved alongside conditional specs
			// Phase C: Verify parent ConceptScope contains child ConceptScope with ConditionalSpecs
		})
	}
}

// TestConditionalSpec_ErrorCases verifies that invalid conditional specification
// syntax produces appropriate parser errors. These cases test malformed conditional
// specs that violate the grammar rules.
func TestConditionalSpec_ErrorCases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name: "ARROW with no conceptLine after",
			input: `concepts
	Country[USA] =>`,
			expectError:   true,
			errorContains: "missing",
		},
		{
			name: "missing parent value in brackets",
			input: `concepts
	Country[] => State=[Values[NY]]`,
			expectError:   true,
			errorContains: "missing",
		},
		{
			name: "invalid syntax no ARROW",
			input: `concepts
	Country[USA] State=[Values[NY]]`,
			expectError:   true,
			errorContains: "missing '=>'",
		},
		{
			name: "missing left bracket around parent value",
			input: `concepts
	CountryUSA] => State=[Values[NY]]`,
			expectError:   true,
			errorContains: "no viable alternative",
		},
		{
			name: "missing right bracket around parent value",
			input: `concepts
	Country[USA => State=[Values[NY]]`,
			expectError:   true,
			errorContains: "missing",
		},
		{
			name: "missing equals in conceptLine",
			input: `concepts
	Country[USA] => State[Values[NY]]`,
			expectError:   true,
			errorContains: "missing",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := parser.Parse(tc.input)

			if !tc.expectError {
				require.NoError(t, err, "Expected parsing to succeed for input: %s", tc.input)
				return
			}

			require.Error(t, err, "Expected parsing to fail for input: %s", tc.input)
			if tc.errorContains != "" {
				require.Contains(t, err.Error(), tc.errorContains,
					"Error message should contain '%s'", tc.errorContains)
			}
		})
	}
}
