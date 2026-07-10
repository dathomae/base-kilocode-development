package main

import (
	"encoding/json"
	"math"
	"testing"

	gen "github.com/chronocone/chronocone_planning_language/src/main/go/cli/expected_xls_generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. compareValues Tests
// =============================================================================

func TestCompareValues_Zero(t *testing.T) {
	assert.True(t, compareValues(0.0, 0.0, 1e-9), "Two zeros should be equal")
}

func TestCompareValues_WithinTolerance(t *testing.T) {
	assert.True(t, compareValues(1.0, 1.0000000001, 1e-9), "Values within tolerance should match")
}

func TestCompareValues_BeyondTolerance(t *testing.T) {
	assert.False(t, compareValues(1.0, 1.1, 1e-9), "Values beyond tolerance should not match")
}

func TestCompareValues_Negative(t *testing.T) {
	assert.True(t, compareValues(-1.0, -1.0, 1e-9), "Identical negative values should match")
}

func TestCompareValues_NaN(t *testing.T) {
	assert.False(t, compareValues(math.NaN(), math.NaN(), 1e-9), "NaN should not equal NaN")
}

func TestCompareValues_LargeMagnitude(t *testing.T) {
	assert.True(t, compareValues(1e15, 1e15+1e5, 1e-9), "Large identical values should match")
	assert.False(t, compareValues(1e15, 1e15+1e6, 1e-9), "Large values beyond tolerance should not match")
}

func TestCompareValues_AtToleranceBoundary(t *testing.T) {
	assert.True(t, compareValues(10.5, 10.5+1e-9, 1e-9), "Value at tolerance boundary should match")
	assert.True(t, compareValues(10.5, 10.5-1e-9, 1e-9), "Value at tolerance boundary should match")
}

// =============================================================================
// 2. Composite Key Construction Tests
// =============================================================================

func TestBuildCompositeKey_Basic(t *testing.T) {
	key := buildCompositeKey([]string{"North", "Monday", "simple_addition"})
	assert.Equal(t, "North|Monday|simple_addition", key)
}

func TestBuildCompositeKey_SinglePart(t *testing.T) {
	key := buildCompositeKey([]string{"only"})
	assert.Equal(t, "only", key)
}

func TestBuildCompositeKey_EmptyParts(t *testing.T) {
	key := buildCompositeKey([]string{"", "b", ""})
	assert.Equal(t, "|b|", key)
}

func TestBuildCompositeKey_WithSpecialCharacters(t *testing.T) {
	key := buildCompositeKey([]string{"a|b", "c"})
	assert.Equal(t, "a|b|c", key)
}

// =============================================================================
// 3. SchemaMeta JSON Parsing Tests
// =============================================================================

func TestParseSchemaMeta_ValidJSON(t *testing.T) {
	jsonContent := `{
		"xlsx_file": "test.xlsx",
		"key_columns": ["Region", "Day"],
		"input_columns": ["data_a", "data_b"],
		"metric_column": 4,
		"value_column": 5
	}`

	var meta gen.SchemaMeta
	err := json.Unmarshal([]byte(jsonContent), &meta)
	require.NoError(t, err)

	assert.Equal(t, "test.xlsx", meta.XLSXFile)
	assert.Equal(t, []string{"Region", "Day"}, meta.KeyColumns)
	assert.Equal(t, 4, meta.MetricColumn)
	assert.Equal(t, 5, meta.ValueColumn)
}

func TestParseSchemaMeta_EmptyKeyColumns(t *testing.T) {
	jsonContent := `{
		"xlsx_file": "test.xlsx",
		"key_columns": [],
		"input_columns": [],
		"metric_column": 0,
		"value_column": 1
	}`

	var meta gen.SchemaMeta
	err := json.Unmarshal([]byte(jsonContent), &meta)
	require.NoError(t, err)

	assert.Empty(t, meta.KeyColumns)
	assert.Equal(t, 0, meta.MetricColumn)
}

func TestParseSchemaMeta_InvalidJSON(t *testing.T) {
	jsonContent := `{this is not valid json`

	var meta gen.SchemaMeta
	err := json.Unmarshal([]byte(jsonContent), &meta)
	assert.Error(t, err, "Invalid JSON should produce an error")
}

// =============================================================================
// 4. Value Comparison and Matching Logic Tests
// =============================================================================

func TestFindMismatches_AllMatch(t *testing.T) {
	expected := map[string]float64{
		"North|Monday|sum": 10.5,
		"South|Monday|sum": 42.25,
	}
	actual := map[string]float64{
		"North|Monday|sum": 10.5,
		"South|Monday|sum": 42.25,
	}

	mismatches := findMismatches(expected, actual, 1e-9)
	assert.Empty(t, mismatches)
}

func TestFindMismatches_Missing(t *testing.T) {
	expected := map[string]float64{
		"North|Monday|sum": 10.5,
		"South|Monday|sum": 42.25,
	}
	actual := map[string]float64{
		"North|Monday|sum": 10.5,
	}

	mismatches := findMismatches(expected, actual, 1e-9)
	require.Len(t, mismatches, 1)
	assert.Equal(t, "missing", mismatches[0].Type)
	assert.Equal(t, "South|Monday|sum", mismatches[0].Key)
	assert.InDelta(t, 42.25, mismatches[0].Expected, 1e-15)
}

func TestFindMismatches_Extra(t *testing.T) {
	expected := map[string]float64{
		"North|Monday|sum": 10.5,
	}
	actual := map[string]float64{
		"North|Monday|sum": 10.5,
		"South|Monday|sum": 42.25,
	}

	mismatches := findMismatches(expected, actual, 1e-9)
	require.Len(t, mismatches, 1)
	assert.Equal(t, "extra", mismatches[0].Type)
	assert.Equal(t, "South|Monday|sum", mismatches[0].Key)
	assert.InDelta(t, 42.25, mismatches[0].Actual, 1e-15)
}

func TestFindMismatches_ValueMismatch(t *testing.T) {
	expected := map[string]float64{
		"North|Monday|sum": 10.5,
	}
	actual := map[string]float64{
		"North|Monday|sum": 99.9,
	}

	mismatches := findMismatches(expected, actual, 1e-9)
	require.Len(t, mismatches, 1)
	assert.Equal(t, "value_mismatch", mismatches[0].Type)
	assert.Equal(t, "North|Monday|sum", mismatches[0].Key)
	assert.InDelta(t, 10.5, mismatches[0].Expected, 1e-15)
	assert.InDelta(t, 99.9, mismatches[0].Actual, 1e-15)
}

func TestFindMismatches_WithinTolerance(t *testing.T) {
	expected := map[string]float64{
		"North|Monday|sum": 10.5,
	}
	actual := map[string]float64{
		"North|Monday|sum": 10.5000000001,
	}

	mismatches := findMismatches(expected, actual, 1e-9)
	assert.Empty(t, mismatches)
}

func TestFindMismatches_MultipleIssues(t *testing.T) {
	expected := map[string]float64{
		"North|Monday|sum": 10.5,
		"East|Monday|sum":  20.0,
	}
	actual := map[string]float64{
		"North|Monday|sum": 99.9,  // value mismatch
		"South|Monday|sum": 42.25, // extra
	}

	mismatches := findMismatches(expected, actual, 1e-9)
	require.Len(t, mismatches, 3)

	types := make(map[string]int)
	for _, m := range mismatches {
		types[m.Type]++
	}
	assert.Equal(t, 1, types["missing"])
	assert.Equal(t, 1, types["extra"])
	assert.Equal(t, 1, types["value_mismatch"])
}

func TestFindMismatches_EmptyMaps(t *testing.T) {
	mismatches := findMismatches(map[string]float64{}, map[string]float64{}, 1e-9)
	assert.Empty(t, mismatches)
}
