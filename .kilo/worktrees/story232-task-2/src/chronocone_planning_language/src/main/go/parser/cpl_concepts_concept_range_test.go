// Package parser_test contains tests for the CPL concepts Concept Range parser.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/ast"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseConceptRangeGenerator_BasicRanges verifies basic Concept Range generator parsing.
func TestParseConceptRangeGenerator_BasicRanges(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "basic concept range generator",
			input:       "concepts\n\tQuarter=[Values[Q1, Q2, Q3, Q4]]\n\tFirstHalf=[Quarter[Q1..Q2]]",
			expectError: false,
		},
		{
			name:        "concept range with different start and end",
			input:       "concepts\n\tMonth=[Values[Jan, Feb, Mar, Apr, May, Jun]]\n\tFirstQuarter=[Month[Jan..Mar]]",
			expectError: false,
		},
		{
			name:        "single value concept range",
			input:       "concepts\n\tStatus=[Values[Active, Inactive]]\n\tOnlyActive=[Status[Active..Active]]",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := parser.Parse(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected an error for input: %s", tc.input)
			} else {
				require.NoError(t, err, "Expected no error for input: %s", tc.input)
			}
		})
	}
}

// TestParseConceptRangeGenerator_SymbolTableStructure verifies that parsing a Concept Range
// creates the correct symbol table scope hierarchy with ConceptRangeGeneratorDef data.
func TestParseConceptRangeGenerator_SymbolTableStructure(t *testing.T) {
	t.Parallel()

	input := "concepts\n\tQuarter=[Values[Q1, Q2, Q3, Q4]]\n\tFirstHalf=[Quarter[Q1..Q2]]"

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Navigate to ConceptsScope
	require.NotEmpty(t, rootTable.Children, "Root should have children")
	var conceptsScope *symboltable.SymbolTable
	for _, child := range rootTable.Children {
		if child.ScopeType == symboltable.ConceptsScope {
			conceptsScope = child
			break
		}
	}
	require.NotNil(t, conceptsScope, "Should find ConceptsScope")

	// Find the ConceptScope for FirstHalf (should be the second concept)
	require.NotEmpty(t, conceptsScope.Children, "ConceptsScope should have children")
	var firstHalfScope *symboltable.SymbolTable
	for _, child := range conceptsScope.Children {
		if child.ScopeType == symboltable.ConceptScope {
			// Check if this concept scope has a ConceptGeneratorScope child
			for _, cchild := range child.Children {
				if cchild.ScopeType == symboltable.ConceptGeneratorScope {
					firstHalfScope = child
					break
				}
			}
		}
		if firstHalfScope != nil {
			break
		}
	}
	require.NotNil(t, firstHalfScope, "Should find ConceptScope with ConceptGeneratorScope")

	// Navigate to ConceptGeneratorScope
	var conceptGenScope *symboltable.SymbolTable
	for _, child := range firstHalfScope.Children {
		if child.ScopeType == symboltable.ConceptGeneratorScope {
			conceptGenScope = child
			break
		}
	}
	require.NotNil(t, conceptGenScope, "Should find ConceptGeneratorScope")

	// Verify the ConceptRangeGeneratorDef is stored as Data
	require.NotNil(t, conceptGenScope.Data, "ConceptGeneratorScope should have Data")
	conceptRangeDef, ok := conceptGenScope.Data.(*ast.ConceptRangeGeneratorDef)
	require.True(t, ok, "Data should be *ast.ConceptRangeGeneratorDef")
	assert.Equal(t, "Quarter", conceptRangeDef.ConceptName, "ConceptName should be 'Quarter'")
	assert.Equal(t, "Q1", conceptRangeDef.StartValue, "StartValue should be 'Q1'")
	assert.Equal(t, "Q2", conceptRangeDef.EndValue, "EndValue should be 'Q2'")
}

// TestParseConceptRangeGenerator_CombinedGenerators verifies Concept Range combined with
// Values and Int Range generators.
func TestParseConceptRangeGenerator_CombinedGenerators(t *testing.T) {
	t.Parallel()

	input := "concepts\n\tQuarter=[Values[Q1, Q2, Q3, Q4]]\n\tMixedConcept=[Values[A, B], Int[1, 3, 1], Quarter[Q1..Q2]]"

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Navigate to find the MixedConcept scope
	var mixedConceptScope *symboltable.SymbolTable
	for _, child := range rootTable.Children {
		if child.ScopeType == symboltable.ConceptsScope {
			for _, cchild := range child.Children {
				if cchild.ScopeType == symboltable.ConceptScope {
					for _, ccchild := range cchild.Children {
						if ccchild.ScopeType == symboltable.ConceptGeneratorScope {
							mixedConceptScope = cchild
							break
						}
					}
				}
				if mixedConceptScope != nil {
					break
				}
			}
		}
		if mixedConceptScope != nil {
			break
		}
	}
	require.NotNil(t, mixedConceptScope, "Should find MixedConcept ConceptScope")

	// Verify it has all three generator types
	var hasValueGen, hasIntGen, hasConceptGen bool
	for _, child := range mixedConceptScope.Children {
		switch child.ScopeType {
		case symboltable.ValueGeneratorScope:
			hasValueGen = true
		case symboltable.IntGeneratorScope:
			hasIntGen = true
		case symboltable.ConceptGeneratorScope:
			hasConceptGen = true
			require.NotNil(t, child.Data, "ConceptGeneratorScope should have Data")
			conceptRangeDef, ok := child.Data.(*ast.ConceptRangeGeneratorDef)
			require.True(t, ok, "Data should be *ast.ConceptRangeGeneratorDef")
			assert.Equal(t, "Quarter", conceptRangeDef.ConceptName)
			assert.Equal(t, "Q1", conceptRangeDef.StartValue)
			assert.Equal(t, "Q2", conceptRangeDef.EndValue)
		}
	}

	assert.True(t, hasValueGen, "Should have ValueGeneratorScope")
	assert.True(t, hasIntGen, "Should have IntGeneratorScope")
	assert.True(t, hasConceptGen, "Should have ConceptGeneratorScope")
}

// TestParseConceptRangeGenerator_ConditionalSpec verifies Concept Range in conditional specs.
func TestParseConceptRangeGenerator_ConditionalSpec(t *testing.T) {
	t.Parallel()

	input := "concepts\n\tQuarter=[Values[Q1, Q2, Q3, Q4]]\n\tQuarter[Q1] => FirstHalf=[Quarter[Q2..Q3]]"

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Navigate to the concept containing the conditional spec (Quarter)
	var quarterScope *symboltable.SymbolTable
	for _, child := range rootTable.Children {
		if child.ScopeType == symboltable.ConceptsScope {
			for _, cchild := range child.Children {
				if cchild.ScopeType == symboltable.ConceptScope {
					for _, ccchild := range cchild.Children {
						if ccchild.ScopeType == symboltable.ConditionalSpecScope {
							quarterScope = cchild
							break
						}
					}
				}
				if quarterScope != nil {
					break
				}
			}
		}
		if quarterScope != nil {
			break
		}
	}
	require.NotNil(t, quarterScope, "Should find ConceptScope with ConditionalSpecScope")

	// Navigate to ConditionalSpecScope -> ConceptScope for "FirstHalf"
	var condSpecScope *symboltable.SymbolTable
	for _, child := range quarterScope.Children {
		if child.ScopeType == symboltable.ConditionalSpecScope {
			condSpecScope = child
			break
		}
	}
	require.NotNil(t, condSpecScope, "Should find ConditionalSpecScope")

	var firstHalfScope *symboltable.SymbolTable
	for _, child := range condSpecScope.Children {
		if child.ScopeType == symboltable.ConceptScope {
			firstHalfScope = child
			break
		}
	}
	require.NotNil(t, firstHalfScope, "Should find child ConceptScope (FirstHalf)")

	// Verify ConditionalSpec data is stored in the child ConceptScope's ConceptDefinitionData
	require.NotNil(t, firstHalfScope.Data, "FirstHalf ConceptScope Data should not be nil")
	conceptData, ok := firstHalfScope.Data.(*ast.ConceptDefinitionData)
	require.True(t, ok, "FirstHalf Data should be *ast.ConceptDefinitionData")
	require.Len(t, conceptData.ConditionalSpecs, 1, "Should have exactly one ConditionalSpec")

	condSpec := conceptData.ConditionalSpecs[0]
	assert.Equal(t, "Q1", condSpec.ParentValue, "ParentValue should be Q1")
	assert.Equal(t, ast.ConceptRangeGenerator, condSpec.GeneratorType, "GeneratorType should be ConceptRangeGenerator")
	assert.Equal(t, "Quarter", condSpec.SourceConceptName, "SourceConceptName should be Quarter")

	require.NotNil(t, condSpec.RangeBounds, "RangeBounds should not be nil")
	assert.Equal(t, "Q2", condSpec.RangeBounds.Start, "RangeBounds.Start should be Q2")
	assert.Equal(t, "Q3", condSpec.RangeBounds.End, "RangeBounds.End should be Q3")
}

// TestParseConceptRangeGenerator_InvalidSyntax verifies error detection for invalid syntax.
func TestParseConceptRangeGenerator_InvalidSyntax(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          string
		expectedErrMsg string
	}{
		{
			name:           "missing dotdot",
			input:          "concepts\n\tQuarter=[Values[Q1, Q2]]\n\tBad=[Quarter[Q1Q2]]",
			expectedErrMsg: "no viable alternative at input",
		},
		{
			name:           "missing closing bracket",
			input:          "concepts\n\tQuarter=[Values[Q1, Q2]]\n\tBad=[Quarter[Q1..Q2",
			expectedErrMsg: "no viable alternative at input",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := parser.Parse(tc.input)

			require.Error(t, err, "Expected an error for input: %s", tc.input)
			assert.Contains(t, err.Error(), tc.expectedErrMsg,
				"Error message should contain '%s'", tc.expectedErrMsg)
		})
	}
}

// TestParseConceptRangeGenerator_SingleValueRange verifies a single-value
// Concept range (start == end) is correctly stored in the symbol table.
func TestParseConceptRangeGenerator_SingleValueRange(t *testing.T) {
	t.Parallel()

	input := "concepts\n\tStatus=[Values[Active, Inactive]]\n\tOnlyActive=[Status[Active..Active]]"

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	var conceptGenScope *symboltable.SymbolTable
	for _, child := range rootTable.Children {
		if child.ScopeType == symboltable.ConceptsScope {
			for _, cchild := range child.Children {
				if cchild.ScopeType == symboltable.ConceptScope {
					for _, ccchild := range cchild.Children {
						if ccchild.ScopeType == symboltable.ConceptGeneratorScope {
							conceptGenScope = ccchild
							break
						}
					}
				}
			}
		}
		if conceptGenScope != nil {
			break
		}
	}
	require.NotNil(t, conceptGenScope, "Should find ConceptGeneratorScope")

	require.NotNil(t, conceptGenScope.Data, "ConceptGeneratorScope should have Data")
	conceptRangeDef, ok := conceptGenScope.Data.(*ast.ConceptRangeGeneratorDef)
	require.True(t, ok, "Data should be *ast.ConceptRangeGeneratorDef")
	assert.Equal(t, "Status", conceptRangeDef.ConceptName, "ConceptName should be 'Status'")
	assert.Equal(t, "Active", conceptRangeDef.StartValue, "StartValue should be 'Active'")
	assert.Equal(t, "Active", conceptRangeDef.EndValue, "EndValue should be 'Active'")
}
