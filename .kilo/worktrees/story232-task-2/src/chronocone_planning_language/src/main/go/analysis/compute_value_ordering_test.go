// Package analysis_test contains tests for the CPL analysis phase.
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComputeValueOrdering_BeforeClause tests that before clause inserts values
// before the specified existing value. Table-driven with cases for: single before
// insertion, before first value, before middle value.
func TestComputeValueOrdering_BeforeClause(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		baseValues     []string
		insertions     []analysis.InsertionRequest
		expectedValues []string
	}{
		{
			name:       "single before insertion",
			baseValues: []string{"Low", "Medium", "High"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{BeforeValue: "Low"}, Values: []string{"Critical"}},
			},
			expectedValues: []string{"Critical", "Low", "Medium", "High"},
		},
		{
			name:       "before first value",
			baseValues: []string{"Low", "Medium", "High"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{BeforeValue: "Low"}, Values: []string{"Urgent"}},
			},
			expectedValues: []string{"Urgent", "Low", "Medium", "High"},
		},
		{
			name:       "before middle value",
			baseValues: []string{"Low", "Medium", "High"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{BeforeValue: "Medium"}, Values: []string{"Normal"}},
			},
			expectedValues: []string{"Low", "Normal", "Medium", "High"},
		},
		{
			name:       "before only value in single-element base",
			baseValues: []string{"Only"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{BeforeValue: "Only"}, Values: []string{"First"}},
			},
			expectedValues: []string{"First", "Only"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result, err := analysis.ComputeValueOrdering(tc.baseValues, tc.insertions)

			// Assert
			require.NoError(t, err, "should not return an error for valid before insertion")
			require.Len(t, result, len(tc.expectedValues), "result length should match expected")

			for i, expected := range tc.expectedValues {
				assert.Equal(t, expected, result[i].Name, "value at position %d should match", i)
				assert.Equal(t, i, result[i].Position, "Position field at index %d should be %d", i, i)
			}
		})
	}
}

// TestComputeValueOrdering_AfterClause tests that after clause inserts values
// after the specified existing value. Table-driven with cases for: single after
// insertion, after last value, after middle value.
func TestComputeValueOrdering_AfterClause(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		baseValues     []string
		insertions     []analysis.InsertionRequest
		expectedValues []string
	}{
		{
			name:       "single after insertion",
			baseValues: []string{"Low", "Medium", "High"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{AfterValue: "Medium"}, Values: []string{"Normal"}},
			},
			expectedValues: []string{"Low", "Medium", "Normal", "High"},
		},
		{
			name:       "after last value",
			baseValues: []string{"Low", "Medium", "High"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{AfterValue: "High"}, Values: []string{"Critical"}},
			},
			expectedValues: []string{"Low", "Medium", "High", "Critical"},
		},
		{
			name:       "after middle value",
			baseValues: []string{"Low", "Medium", "High"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{AfterValue: "Low"}, Values: []string{"Urgent"}},
			},
			expectedValues: []string{"Low", "Urgent", "Medium", "High"},
		},
		{
			name:       "after only value in single-element base",
			baseValues: []string{"Only"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{AfterValue: "Only"}, Values: []string{"Last"}},
			},
			expectedValues: []string{"Only", "Last"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result, err := analysis.ComputeValueOrdering(tc.baseValues, tc.insertions)

			// Assert
			require.NoError(t, err, "should not return an error for valid after insertion")
			require.Len(t, result, len(tc.expectedValues), "result length should match expected")

			for i, expected := range tc.expectedValues {
				assert.Equal(t, expected, result[i].Name, "value at position %d should match", i)
				assert.Equal(t, i, result[i].Position, "Position field at index %d should be %d", i, i)
			}
		})
	}
}

// TestComputeValueOrdering_MultipleExtensions tests that multiple insertions
// are processed in order. Cases: apply before then after, apply multiple befores,
// multiple values in single insertion.
func TestComputeValueOrdering_MultipleExtensions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		baseValues     []string
		insertions     []analysis.InsertionRequest
		expectedValues []string
	}{
		{
			name:       "apply before then after",
			baseValues: []string{"Low", "Medium", "High"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{BeforeValue: "Low"}, Values: []string{"Critical"}},
				{Position: ast.InsertionPoint{AfterValue: "Medium"}, Values: []string{"Urgent"}},
			},
			expectedValues: []string{"Critical", "Low", "Medium", "Urgent", "High"},
		},
		{
			name:       "apply multiple befores",
			baseValues: []string{"Low", "Medium", "High"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{BeforeValue: "Low"}, Values: []string{"Critical"}},
				{Position: ast.InsertionPoint{BeforeValue: "Medium"}, Values: []string{"Urgent"}},
			},
			expectedValues: []string{"Critical", "Low", "Urgent", "Medium", "High"},
		},
		{
			name:       "multiple values in single insertion",
			baseValues: []string{"A", "D"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{AfterValue: "A"}, Values: []string{"B", "C"}},
			},
			expectedValues: []string{"A", "B", "C", "D"},
		},
		{
			name:       "insertion references value from prior insertion (after)",
			baseValues: []string{"A", "C"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{AfterValue: "A"}, Values: []string{"B"}},
				{Position: ast.InsertionPoint{AfterValue: "B"}, Values: []string{"D"}},
			},
			expectedValues: []string{"A", "B", "D", "C"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result, err := analysis.ComputeValueOrdering(tc.baseValues, tc.insertions)

			// Assert
			require.NoError(t, err, "should not return an error for valid multiple insertions")
			require.Len(t, result, len(tc.expectedValues), "result length should match expected")

			for i, expected := range tc.expectedValues {
				assert.Equal(t, expected, result[i].Name, "value at position %d should match", i)
				assert.Equal(t, i, result[i].Position, "Position field at index %d should be %d", i, i)
			}
		})
	}
}

// TestComputeValueOrdering_InvalidBeforeValue tests that referencing a
// non-existent before value returns an error.
func TestComputeValueOrdering_InvalidBeforeValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		baseValues []string
		insertions []analysis.InsertionRequest
	}{
		{
			name:       "before non-existent value in non-empty list",
			baseValues: []string{"Low", "Medium", "High"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{BeforeValue: "NonExistent"}, Values: []string{"New"}},
			},
		},
		{
			name:       "before value in empty base list",
			baseValues: []string{},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{BeforeValue: "Low"}, Values: []string{"New"}},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result, err := analysis.ComputeValueOrdering(tc.baseValues, tc.insertions)

			// Assert
			assert.Error(t, err, "should return an error for invalid before value")
			assert.Nil(t, result, "result should be nil on error")
		})
	}
}

// TestComputeValueOrdering_InvalidAfterValue tests that referencing a
// non-existent after value returns an error.
func TestComputeValueOrdering_InvalidAfterValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		baseValues []string
		insertions []analysis.InsertionRequest
	}{
		{
			name:       "after non-existent value in non-empty list",
			baseValues: []string{"Low", "Medium", "High"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{AfterValue: "NonExistent"}, Values: []string{"New"}},
			},
		},
		{
			name:       "after value in empty base list",
			baseValues: []string{},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{AfterValue: "High"}, Values: []string{"New"}},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result, err := analysis.ComputeValueOrdering(tc.baseValues, tc.insertions)

			// Assert
			assert.Error(t, err, "should return an error for invalid after value")
			assert.Nil(t, result, "result should be nil on error")
		})
	}
}

// TestComputeValueOrdering_OriginalOrderPreserved tests that the relative order
// of original values is preserved after insertions.
func TestComputeValueOrdering_OriginalOrderPreserved(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		baseValues     []string
		insertions     []analysis.InsertionRequest
		expectedValues []string
	}{
		{
			name:       "preserve order with before insertion",
			baseValues: []string{"A", "B", "C"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{BeforeValue: "B"}, Values: []string{"X"}},
			},
			expectedValues: []string{"A", "X", "B", "C"},
		},
		{
			name:       "preserve order with after insertion",
			baseValues: []string{"A", "B", "C"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{AfterValue: "B"}, Values: []string{"X"}},
			},
			expectedValues: []string{"A", "B", "X", "C"},
		},
		{
			name:       "preserve order with multiple insertions",
			baseValues: []string{"A", "B", "C", "D"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{BeforeValue: "A"}, Values: []string{"X"}},
				{Position: ast.InsertionPoint{AfterValue: "D"}, Values: []string{"Y"}},
			},
			expectedValues: []string{"X", "A", "B", "C", "D", "Y"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result, err := analysis.ComputeValueOrdering(tc.baseValues, tc.insertions)

			// Assert
			require.NoError(t, err, "should not return an error for valid insertions")
			require.Len(t, result, len(tc.expectedValues), "result length should match expected")

			for i, expected := range tc.expectedValues {
				assert.Equal(t, expected, result[i].Name, "value at position %d should be '%s'", i, expected)
				assert.Equal(t, i, result[i].Position, "Position field at index %d should be %d", i, i)
			}
		})
	}
}

// TestComputeValueOrdering_EmptyInsertions tests that passing an empty
// insertions slice returns the base values unchanged.
func TestComputeValueOrdering_EmptyInsertions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		baseValues     []string
		expectedValues []string
	}{
		{
			name:           "empty insertions with populated base values",
			baseValues:     []string{"Low", "Medium", "High"},
			expectedValues: []string{"Low", "Medium", "High"},
		},
		{
			name:           "empty insertions with single base value",
			baseValues:     []string{"Single"},
			expectedValues: []string{"Single"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result, err := analysis.ComputeValueOrdering(tc.baseValues, []analysis.InsertionRequest{})

			// Assert
			require.NoError(t, err, "should not return an error for empty insertions")
			require.Len(t, result, len(tc.expectedValues), "result length should match expected")

			for i, expected := range tc.expectedValues {
				assert.Equal(t, expected, result[i].Name, "value at position %d should match", i)
				assert.Equal(t, i, result[i].Position, "Position field at index %d should be %d", i, i)
			}
		})
	}
}

// TestComputeValueOrdering_EmptyBaseValues tests that passing empty base values
// with insertions returns an error (since before/after values cannot exist in an
// empty list).
func TestComputeValueOrdering_EmptyBaseValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		baseValues []string
		insertions []analysis.InsertionRequest
	}{
		{
			name:       "empty base values with before insertion",
			baseValues: []string{},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{BeforeValue: "X"}, Values: []string{"New"}},
			},
		},
		{
			name:       "empty base values with after insertion",
			baseValues: []string{},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{AfterValue: "X"}, Values: []string{"New"}},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result, err := analysis.ComputeValueOrdering(tc.baseValues, tc.insertions)

			// Assert
			assert.Error(t, err, "should return an error when base values are empty and insertions reference values")
			assert.Nil(t, result, "result should be nil on error")
		})
	}
}

// TestComputeValueOrdering_InvalidInsertionPoint tests that an insertion point
// with both BeforeValue and AfterValue set (or neither) returns an error.
func TestComputeValueOrdering_InvalidInsertionPoint(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		insertions []analysis.InsertionRequest
	}{
		{
			name: "both before and after set",
			insertions: []analysis.InsertionRequest{
				{
					Position: ast.InsertionPoint{BeforeValue: "A", AfterValue: "B"},
					Values:   []string{"New"},
				},
			},
		},
		{
			name: "neither before nor after set",
			insertions: []analysis.InsertionRequest{
				{
					Position: ast.InsertionPoint{},
					Values:   []string{"New"},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result, err := analysis.ComputeValueOrdering([]string{"A", "B"}, tc.insertions)

			// Assert
			assert.Error(t, err, "should return an error for invalid insertion point")
			assert.Nil(t, result, "result should be nil on error")
		})
	}
}

// TestComputeValueOrdering_EmptyValuesSlice tests that an insertion request
// with a valid anchor but an empty Values slice succeeds as a no-op: the
// anchor is validated but no values are inserted.
func TestComputeValueOrdering_EmptyValuesSlice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		baseValues     []string
		insertions     []analysis.InsertionRequest
		expectedValues []string
	}{
		{
			name:       "empty values with before anchor",
			baseValues: []string{"A", "B", "C"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{BeforeValue: "B"}, Values: []string{}},
			},
			expectedValues: []string{"A", "B", "C"},
		},
		{
			name:       "empty values with after anchor",
			baseValues: []string{"A", "B", "C"},
			insertions: []analysis.InsertionRequest{
				{Position: ast.InsertionPoint{AfterValue: "B"}, Values: []string{}},
			},
			expectedValues: []string{"A", "B", "C"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result, err := analysis.ComputeValueOrdering(tc.baseValues, tc.insertions)

			// Assert
			require.NoError(t, err, "should not return an error for valid anchor with empty values")
			require.Len(t, result, len(tc.expectedValues), "result length should match expected")

			for i, expected := range tc.expectedValues {
				assert.Equal(t, expected, result[i].Name, "value at position %d should match", i)
				assert.Equal(t, i, result[i].Position, "Position field at index %d should be %d", i, i)
			}
		})
	}
}
