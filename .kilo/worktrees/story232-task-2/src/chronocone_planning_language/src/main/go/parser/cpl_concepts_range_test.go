// Package parser_test contains tests for the CPL concepts Int range parser.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/ast"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseIntRangeGenerator_BasicRanges verifies basic Int range generator parsing.
func TestParseIntRangeGenerator_BasicRanges(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "month range with three params",
			input:       "concepts\n\tMonth=[Int[1, 12, 1]]",
			expectError: false,
		},
		{
			name:        "fiscal period range with three params",
			input:       "concepts\n\tFiscalPeriod=[Int[1, 13, 1]]",
			expectError: false,
		},
		{
			name:        "range with two params defaults step to 1",
			input:       "concepts\n\tNumbers=[Int[1, 5]]",
			expectError: false,
		},
		{
			name:        "even numbers with custom step",
			input:       "concepts\n\tEvenNumbers=[Int[2, 10, 2]]",
			expectError: false,
		},
		{
			name:        "multiple generators combined",
			input:       "concepts\n\tMixed=[Values[A, B], Int[1, 3, 1]]",
			expectError: false,
		},
		{
			name:        "single value range with positive step",
			input:       "concepts\n\tSingleton=[Int[7, 7, 1]]",
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

// TestParseIntRangeGenerator_SymbolTableStructure verifies that parsing an Int range
// creates the correct symbol table scope hierarchy with IntRangeGeneratorDef data.
func TestParseIntRangeGenerator_SymbolTableStructure(t *testing.T) {
	t.Parallel()

	input := "concepts\n\tMonth=[Int[1, 12, 1]]"

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

	// Navigate to ConceptScope
	require.NotEmpty(t, conceptsScope.Children, "ConceptsScope should have children")
	var conceptScope *symboltable.SymbolTable
	for _, child := range conceptsScope.Children {
		if child.ScopeType == symboltable.ConceptScope {
			conceptScope = child
			break
		}
	}
	require.NotNil(t, conceptScope, "Should find ConceptScope")

	// Navigate to IntGeneratorScope
	require.NotEmpty(t, conceptScope.Children, "ConceptScope should have children")
	var intGenScope *symboltable.SymbolTable
	for _, child := range conceptScope.Children {
		if child.ScopeType == symboltable.IntGeneratorScope {
			intGenScope = child
			break
		}
	}
	require.NotNil(t, intGenScope, "Should find IntGeneratorScope under ConceptScope")

	// Verify the IntRangeGeneratorDef is stored as Data
	require.NotNil(t, intGenScope.Data, "IntGeneratorScope should have Data")
	intRangeDef, ok := intGenScope.Data.(*ast.IntRangeGeneratorDef)
	require.True(t, ok, "Data should be *ast.IntRangeGeneratorDef")
	assert.Equal(t, 1, intRangeDef.Start, "Start should be 1")
	assert.Equal(t, 12, intRangeDef.End, "End should be 12")
	assert.Equal(t, 1, intRangeDef.Step, "Step should be 1")
}

// TestParseIntRangeGenerator_TwoParamsDefaultsStep verifies that an Int range with only
// two parameters defaults Step to 1 in the symbol table data.
func TestParseIntRangeGenerator_TwoParamsDefaultsStep(t *testing.T) {
	t.Parallel()

	input := "concepts\n\tNumbers=[Int[1, 5]]"

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	var intGenScope *symboltable.SymbolTable
	for _, child := range rootTable.Children {
		if child.ScopeType == symboltable.ConceptsScope {
			for _, cchild := range child.Children {
				if cchild.ScopeType == symboltable.ConceptScope {
					for _, ccchild := range cchild.Children {
						if ccchild.ScopeType == symboltable.IntGeneratorScope {
							intGenScope = ccchild
							break
						}
					}
				}
			}
		}
	}
	require.NotNil(t, intGenScope, "Should find IntGeneratorScope")

	require.NotNil(t, intGenScope.Data, "IntGeneratorScope should have Data")
	intRangeDef, ok := intGenScope.Data.(*ast.IntRangeGeneratorDef)
	require.True(t, ok, "Data should be *ast.IntRangeGeneratorDef")
	assert.Equal(t, 1, intRangeDef.Start, "Start should be 1")
	assert.Equal(t, 5, intRangeDef.End, "End should be 5")
	assert.Equal(t, 1, intRangeDef.Step, "Step should default to 1")
}

// TestParseIntRangeGenerator_ValidationErrors verifies validation of Int range parameters.
func TestParseIntRangeGenerator_ValidationErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          string
		expectedErrMsg string
	}{
		{
			name:           "step cannot be zero",
			input:          "concepts\n\tBad=[Int[1, 5, 0]]",
			expectedErrMsg: "step",
		},
		{
			name:           "start must be less than end when step positive",
			input:          "concepts\n\tBad=[Int[5, 1, 1]]",
			expectedErrMsg: "start",
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

// TestParseIntRangeGenerator_InvalidDataNotStored verifies that when validation fails,
// no IntRangeGeneratorDef is stored in the symbol table scope.
func TestParseIntRangeGenerator_InvalidDataNotStored(t *testing.T) {
	t.Parallel()

	input := "concepts\n\tBad=[Int[5, 1, 1]]"

	rootTable, err := parser.Parse(input)
	require.Error(t, err, "Parsing should fail due to validation error")
	require.NotNil(t, rootTable, "Root table should still be returned")

	// Navigate to IntGeneratorScope
	var intGenScope *symboltable.SymbolTable
	for _, child := range rootTable.Children {
		if child.ScopeType == symboltable.ConceptsScope {
			for _, cchild := range child.Children {
				if cchild.ScopeType == symboltable.ConceptScope {
					for _, ccchild := range cchild.Children {
						if ccchild.ScopeType == symboltable.IntGeneratorScope {
							intGenScope = ccchild
							break
						}
					}
				}
			}
		}
	}
	require.NotNil(t, intGenScope, "Should find IntGeneratorScope")
	assert.Nil(t, intGenScope.Data, "IntGeneratorScope.Data should be nil when validation fails")
}

// TestParseIntRangeGenerator_CustomStepData verifies that a custom step value
// is correctly stored in the symbol table.
func TestParseIntRangeGenerator_CustomStepData(t *testing.T) {
	t.Parallel()

	input := "concepts\n\tEvenNumbers=[Int[2, 10, 2]]"

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	var intGenScope *symboltable.SymbolTable
	for _, child := range rootTable.Children {
		if child.ScopeType == symboltable.ConceptsScope {
			for _, cchild := range child.Children {
				if cchild.ScopeType == symboltable.ConceptScope {
					for _, ccchild := range cchild.Children {
						if ccchild.ScopeType == symboltable.IntGeneratorScope {
							intGenScope = ccchild
							break
						}
					}
				}
			}
		}
	}
	require.NotNil(t, intGenScope, "Should find IntGeneratorScope")

	require.NotNil(t, intGenScope.Data, "IntGeneratorScope should have Data")
	intRangeDef, ok := intGenScope.Data.(*ast.IntRangeGeneratorDef)
	require.True(t, ok, "Data should be *ast.IntRangeGeneratorDef")
	assert.Equal(t, 2, intRangeDef.Start, "Start should be 2")
	assert.Equal(t, 10, intRangeDef.End, "End should be 10")
	assert.Equal(t, 2, intRangeDef.Step, "Step should be 2")
}

// TestParseIntRangeGenerator_SingleValueRangeData verifies that a single-value
// Int range (start == end) is correctly stored in the symbol table.
func TestParseIntRangeGenerator_SingleValueRangeData(t *testing.T) {
	t.Parallel()

	input := "concepts\n\tSingleton=[Int[7, 7, 1]]"

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	var intGenScope *symboltable.SymbolTable
	for _, child := range rootTable.Children {
		if child.ScopeType == symboltable.ConceptsScope {
			for _, cchild := range child.Children {
				if cchild.ScopeType == symboltable.ConceptScope {
					for _, ccchild := range cchild.Children {
						if ccchild.ScopeType == symboltable.IntGeneratorScope {
							intGenScope = ccchild
							break
						}
					}
				}
			}
		}
	}
	require.NotNil(t, intGenScope, "Should find IntGeneratorScope")

	require.NotNil(t, intGenScope.Data, "IntGeneratorScope should have Data")
	intRangeDef, ok := intGenScope.Data.(*ast.IntRangeGeneratorDef)
	require.True(t, ok, "Data should be *ast.IntRangeGeneratorDef")
	assert.Equal(t, 7, intRangeDef.Start, "Start should be 7")
	assert.Equal(t, 7, intRangeDef.End, "End should be 7")
	assert.Equal(t, 1, intRangeDef.Step, "Step should be 1")
}
