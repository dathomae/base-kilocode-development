// Package main contains tests for the parquet_verify CLI tool.
//
// These tests verify the command-line tool's ability to compare a synthesized
// output parquet file against an expected JSON specification. The tests follow
// the Logical TDD Lifecycle (BTDD) approach.
//
// Phase B tests are written against the existing function signatures from
// Phase A. All tests that require logic beyond the stub implementations
// are expected to FAIL. Tests that assert on zero-value return behavior
// (e.g., compareValues returning false for equal values) will PASS
// incorrectly but will catch regressions once Phase C is implemented.
//
// Test categories:
//   - JSON parsing of expected output files
//   - Schema verification (column names and types)
//   - Row count verification
//   - Row matching (metric, dimension values, grain values)
//   - Value comparison with floating-point tolerance
//   - Mismatch reporting (missing, extra, and value differences)
//   - End-to-end CLI integration tests with exit codes
package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	cplio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. JSON Parsing Tests — parse expected_output.json into ExpectedOutput struct
// =============================================================================

// TestParseExpectedOutput_ValidJSON verifies that a well-formed expected_output.json
// with all fields populated can be unmarshaled into an ExpectedOutput struct.
//
// All fields (output_file, metrics, dimensions, grain, rows) are checked for
// correct population including nested ExpectedRow fields.
func TestParseExpectedOutput_ValidJSON(t *testing.T) {
	jsonContent := `{
		"output_file": "output.parquet",
		"metrics": ["sum_metric"],
		"dimensions": ["Region"],
		"grain": ["Day"],
		"rows": [
			{
				"Metric": "sum_metric",
				"dimension_values": { "Region": "North" },
				"grain_values": { "Day": "Monday" },
				"value": 10.5
			},
			{
				"Metric": "sum_metric",
				"dimension_values": { "Region": "South" },
				"grain_values": { "Day": "Monday" },
				"value": 42.25
			}
		]
	}`

	var eo ExpectedOutput
	err := json.Unmarshal([]byte(jsonContent), &eo)
	require.NoError(t, err)

	assert.Equal(t, "output.parquet", eo.OutputFile)
	assert.Equal(t, []string{"sum_metric"}, eo.Metrics)
	assert.Equal(t, []string{"Region"}, eo.Dimensions)
	assert.Equal(t, []string{"Day"}, eo.Grain)

	require.Len(t, eo.Rows, 2)

	// First row
	assert.Equal(t, "sum_metric", eo.Rows[0].Metric)
	assert.Equal(t, map[string]string{"Region": "North"}, eo.Rows[0].DimensionValues)
	assert.Equal(t, map[string]string{"Day": "Monday"}, eo.Rows[0].GrainValues)
	assert.InDelta(t, 10.5, eo.Rows[0].Value, 1e-15)

	// Second row
	assert.Equal(t, "sum_metric", eo.Rows[1].Metric)
	assert.Equal(t, map[string]string{"Region": "South"}, eo.Rows[1].DimensionValues)
	assert.Equal(t, map[string]string{"Day": "Monday"}, eo.Rows[1].GrainValues)
	assert.InDelta(t, 42.25, eo.Rows[1].Value, 1e-15)
}

// TestParseExpectedOutput_WithMultipleMetrics verifies that an expected_output.json
// with multiple metrics in the "metrics" array is parsed correctly.
func TestParseExpectedOutput_WithMultipleMetrics(t *testing.T) {
	jsonContent := `{
		"output_file": "output.parquet",
		"metrics": ["profit", "margin_pct", "markup"],
		"dimensions": ["Region", "Product"],
		"grain": ["Year"],
		"rows": [
			{
				"Metric": "profit",
				"dimension_values": { "Region": "North", "Product": "Widget" },
				"grain_values": { "Year": "2024" },
				"value": 100.0
			}
		]
	}`

	var eo ExpectedOutput
	err := json.Unmarshal([]byte(jsonContent), &eo)
	require.NoError(t, err)

	assert.Equal(t, []string{"profit", "margin_pct", "markup"}, eo.Metrics)
	assert.Equal(t, []string{"Region", "Product"}, eo.Dimensions)
	assert.Equal(t, []string{"Year"}, eo.Grain)

	require.Len(t, eo.Rows, 1)
	assert.Equal(t, "profit", eo.Rows[0].Metric)
	assert.Equal(t, map[string]string{"Region": "North", "Product": "Widget"}, eo.Rows[0].DimensionValues)
	assert.Equal(t, map[string]string{"Year": "2024"}, eo.Rows[0].GrainValues)
}

// TestParseExpectedOutput_MissingOptionalFields verifies that JSON with missing
// optional fields is parsed gracefully. The output_file field is optional
// (for reference only); missing it should result in an empty string.
// The metrics, dimensions, grain, and rows fields are required by schema
// but Go JSON decoding handles missing fields by leaving zero values.
func TestParseExpectedOutput_MissingOptionalFields(t *testing.T) {
	// Missing output_file field
	jsonContent := `{
		"metrics": ["sum_metric"],
		"dimensions": ["Region"],
		"grain": [],
		"rows": []
	}`

	var eo ExpectedOutput
	err := json.Unmarshal([]byte(jsonContent), &eo)
	require.NoError(t, err)

	assert.Empty(t, eo.OutputFile, "output_file should default to empty string when missing")
	assert.Equal(t, []string{"sum_metric"}, eo.Metrics)
	assert.Empty(t, eo.Grain)
	assert.Empty(t, eo.Rows)
}

// TestParseExpectedOutput_InvalidJSON verifies that malformed JSON produces
// an appropriate unmarshal error and the struct is not populated.
func TestParseExpectedOutput_InvalidJSON(t *testing.T) {
	jsonContent := `{this is not valid json`

	var eo ExpectedOutput
	err := json.Unmarshal([]byte(jsonContent), &eo)
	assert.Error(t, err, "Invalid JSON should produce an error")
}

// TestParseExpectedOutput_InvalidFieldTypes verifies that JSON with incorrect
// field types (e.g., string where array is expected) produces an unmarshal error.
func TestParseExpectedOutput_InvalidFieldTypes(t *testing.T) {
	// metrics should be an array, but we provide a string
	jsonContent := `{
		"output_file": "output.parquet",
		"metrics": "not_an_array",
		"dimensions": ["Region"],
		"grain": [],
		"rows": []
	}`

	var eo ExpectedOutput
	err := json.Unmarshal([]byte(jsonContent), &eo)
	assert.Error(t, err, "Wrong JSON type for metrics field should produce an error")
}

// TestParseExpectedOutput_EmptyRows verifies that an expected_output.json with
// an empty rows array is parsed correctly (no rows expected).
func TestParseExpectedOutput_EmptyRows(t *testing.T) {
	jsonContent := `{
		"output_file": "output.parquet",
		"metrics": ["sum_metric"],
		"dimensions": [],
		"grain": [],
		"rows": []
	}`

	var eo ExpectedOutput
	err := json.Unmarshal([]byte(jsonContent), &eo)
	require.NoError(t, err)

	assert.Empty(t, eo.Rows)
	assert.Empty(t, eo.Dimensions)
	assert.Empty(t, eo.Grain)
}

// TestParseExpectedOutput_NegativeAndZeroValues verifies that ExpectedRow values
// can be negative, zero, or very small, all of which are valid float64 values.
func TestParseExpectedOutput_NegativeAndZeroValues(t *testing.T) {
	jsonContent := `{
		"output_file": "output.parquet",
		"metrics": ["net_profit"],
		"dimensions": ["Region"],
		"grain": [],
		"rows": [
			{
				"Metric": "net_profit",
				"dimension_values": { "Region": "North" },
				"grain_values": {},
				"value": -15.75
			},
			{
				"Metric": "net_profit",
				"dimension_values": { "Region": "South" },
				"grain_values": {},
				"value": 0.0
			}
		]
	}`

	var eo ExpectedOutput
	err := json.Unmarshal([]byte(jsonContent), &eo)
	require.NoError(t, err)

	require.Len(t, eo.Rows, 2)
	assert.InDelta(t, -15.75, eo.Rows[0].Value, 1e-15)
	assert.InDelta(t, 0.0, eo.Rows[1].Value, 1e-15)
}

// =============================================================================
// 2. Schema Verification Tests — verifySchema()
// =============================================================================

// TestVerifySchema_Matching verifies that verifySchema returns nil when the
// parquet schema exactly matches the expected columns.
//
// Phase B note: This test PASSES because the stub returns nil. If the stub
// were later changed to return a not-implemented error, this would catch it.
func TestVerifySchema_Matching(t *testing.T) {
	tempDir := t.TempDir()
	parquetPath := filepath.Join(tempDir, "test.parquet")

	dimensions := map[string][]string{
		"Metric": {"sum_metric"},
		"Region": {"North"},
	}
	grain := []cplio.GrainMeasure{
		{Name: "Day", Values: []string{"Monday"}},
	}

	err := cplio.GenerateParquetFile(parquetPath, dimensions, grain, []float64{10.5})
	require.NoError(t, err)

	schema, _, err := cplio.ReadParquetFile(parquetPath)
	require.NoError(t, err)

	expected := &ExpectedOutput{
		Metrics:    []string{"sum_metric"},
		Dimensions: []string{"Region"},
		Grain:      []string{"Day"},
	}

	// Phase B: stub returns nil, so this passes (but verifies the signature compiles)
	err = verifySchema(schema, expected)
	assert.NoError(t, err, "Matching schema should not produce an error")
}

// TestVerifySchema_MissingMetricColumn verifies that verifySchema returns an
// error when the parquet schema is missing the required "Metric" column.
//
// Phase B: This test FAILS because the stub returns nil. It documents the
// expected behavior: a schema without a Metric column is invalid.
func TestVerifySchema_MissingMetricColumn(t *testing.T) {
	tempDir := t.TempDir()
	parquetPath := filepath.Join(tempDir, "test.parquet")

	// Generate a parquet file WITHOUT a Metric column (using dimensions only,
	// which GenerateParquetFile does not include a Metric column)
	dimensions := map[string][]string{
		"Region": {"North"},
	}
	grain := []cplio.GrainMeasure{
		{Name: "Day", Values: []string{"Monday"}},
	}

	err := cplio.GenerateParquetFile(parquetPath, dimensions, grain, []float64{10.5})
	require.NoError(t, err)

	schema, _, err := cplio.ReadParquetFile(parquetPath)
	require.NoError(t, err)

	expected := &ExpectedOutput{
		Metrics:    []string{"sum_metric"},
		Dimensions: []string{"Region"},
		Grain:      []string{"Day"},
	}

	// Phase B: stub returns nil; this assertion will FAIL until Phase C
	err = verifySchema(schema, expected)
	assert.Error(t, err, "Schema missing Metric column should produce an error")
}

// TestVerifySchema_MissingDimensionColumn verifies that verifySchema returns an
// error when the parquet schema is missing a required dimension column.
//
// Phase B: This test FAILS because the stub returns nil.
func TestVerifySchema_MissingDimensionColumn(t *testing.T) {
	tempDir := t.TempDir()
	parquetPath := filepath.Join(tempDir, "test.parquet")

	// Generate a parquet with only "Region" dimension
	dimensions := map[string][]string{
		"Region": {"North"},
	}
	grain := []cplio.GrainMeasure{
		{Name: "Day", Values: []string{"Monday"}},
	}

	err := cplio.GenerateParquetFile(parquetPath, dimensions, grain, []float64{10.5})
	require.NoError(t, err)

	schema, _, err := cplio.ReadParquetFile(parquetPath)
	require.NoError(t, err)

	// Expected declares "Product" dimension which is not in the parquet schema
	expected := &ExpectedOutput{
		Metrics:    []string{"sum_metric"},
		Dimensions: []string{"Region", "Product"}, // Product not in schema
		Grain:      []string{"Day"},
	}

	err = verifySchema(schema, expected)
	assert.Error(t, err, "Schema missing expected dimension column should produce an error")
}

// TestVerifySchema_ExtraColumn verifies that verifySchema returns an error
// when the parquet schema contains columns not expected in the output spec.
//
// Phase B: This test FAILS because the stub returns nil.
func TestVerifySchema_ExtraColumn(t *testing.T) {
	tempDir := t.TempDir()
	parquetPath := filepath.Join(tempDir, "test.parquet")

	// Generate a parquet with both Region and Product dimensions,
	// but the expected only lists Region
	dimensions := map[string][]string{
		"Region":  {"North"},
		"Product": {"Widget"},
	}
	grain := []cplio.GrainMeasure{
		{Name: "Day", Values: []string{"Monday"}},
	}

	err := cplio.GenerateParquetFile(parquetPath, dimensions, grain, []float64{10.5})
	require.NoError(t, err)

	schema, _, err := cplio.ReadParquetFile(parquetPath)
	require.NoError(t, err)

	expected := &ExpectedOutput{
		Metrics:    []string{"sum_metric"},
		Dimensions: []string{"Region"}, // Product dimension not expected
		Grain:      []string{"Day"},
	}

	err = verifySchema(schema, expected)
	assert.Error(t, err, "Schema with extra columns should produce an error")
}

// =============================================================================
// 3. Row Count Verification Tests — verifyRowCount()
// =============================================================================

// TestVerifyRowCount_Matching verifies that verifyRowCount returns nil when
// actual and expected row counts are equal.
//
// Phase B: This test PASSES because the stub returns nil.
func TestVerifyRowCount_Matching(t *testing.T) {
	expected := &ExpectedOutput{
		Rows: []ExpectedRow{
			{Metric: "m1", Value: 1.0},
			{Metric: "m2", Value: 2.0},
		},
	}

	err := verifyRowCount(2, expected)
	assert.NoError(t, err, "Matching row counts should not produce an error")
}

// TestVerifyRowCount_Mismatch verifies that verifyRowCount returns an error
// when the actual row count differs from the expected row count.
//
// Phase B: This test FAILS because the stub returns nil.
func TestVerifyRowCount_Mismatch(t *testing.T) {
	expected := &ExpectedOutput{
		Rows: []ExpectedRow{
			{Metric: "m1", Value: 1.0},
			{Metric: "m2", Value: 2.0},
			{Metric: "m3", Value: 3.0},
		},
	}

	err := verifyRowCount(2, expected)
	require.Error(t, err, "Mismatched row counts should produce an error")
	assert.Contains(t, err.Error(), "expected 3 rows but found 2")
}

// TestVerifyRowCount_TooMany verifies that verifyRowCount returns an error
// when the parquet has more rows than expected.
//
// Phase B: This test FAILS because the stub returns nil.
func TestVerifyRowCount_TooMany(t *testing.T) {
	expected := &ExpectedOutput{
		Rows: []ExpectedRow{
			{Metric: "m1", Value: 1.0},
		},
	}

	err := verifyRowCount(5, expected)
	assert.Error(t, err, "Parquet with more rows than expected should produce an error")
}

// =============================================================================
// 4. Row Matching Tests — matchRows()
// =============================================================================

// TestMatchRows_AllMatch verifies that matchRows returns an empty mismatch
// slice when all expected rows have exact matches in the parquet data.
//
// Phase B: This test FAILS because the stub returns (nil, nil).
func TestMatchRows_AllMatch(t *testing.T) {
	parquetRows := []parquetRowData{
		{
			Metric:          "sum_metric",
			DimensionValues: map[string]string{"Region": "North"},
			GrainValues:     map[string]string{"Day": "Monday"},
			Value:           10.5,
		},
		{
			Metric:          "sum_metric",
			DimensionValues: map[string]string{"Region": "South"},
			GrainValues:     map[string]string{"Day": "Monday"},
			Value:           42.25,
		},
	}

	expected := &ExpectedOutput{
		Rows: []ExpectedRow{
			{
				Metric:          "sum_metric",
				DimensionValues: map[string]string{"Region": "North"},
				GrainValues:     map[string]string{"Day": "Monday"},
				Value:           10.5,
			},
			{
				Metric:          "sum_metric",
				DimensionValues: map[string]string{"Region": "South"},
				GrainValues:     map[string]string{"Day": "Monday"},
				Value:           42.25,
			},
		},
	}

	mismatches, err := matchRows(parquetRows, expected)
	require.NoError(t, err)
	assert.Empty(t, mismatches, "All rows match, expected no mismatches")
}

// TestMatchRows_MissingRow verifies that matchRows reports a "missing" mismatch
// when an expected row is not found in the parquet data.
//
// Phase B: This test FAILS because the stub returns (nil, nil).
func TestMatchRows_MissingRow(t *testing.T) {
	parquetRows := []parquetRowData{
		{
			Metric:          "sum_metric",
			DimensionValues: map[string]string{"Region": "North"},
			GrainValues:     map[string]string{"Day": "Monday"},
			Value:           10.5,
		},
	}

	expected := &ExpectedOutput{
		Rows: []ExpectedRow{
			{
				Metric:          "sum_metric",
				DimensionValues: map[string]string{"Region": "North"},
				GrainValues:     map[string]string{"Day": "Monday"},
				Value:           10.5,
			},
			{
				Metric:          "sum_metric",
				DimensionValues: map[string]string{"Region": "South"},
				GrainValues:     map[string]string{"Day": "Monday"},
				Value:           42.25,
			},
		},
	}

	mismatches, err := matchRows(parquetRows, expected)
	require.NoError(t, err)
	require.Len(t, mismatches, 1, "One row is missing from parquet")
	assert.Equal(t, "missing", mismatches[0].Type)
	assert.Equal(t, 1, mismatches[0].ExpectedRowIndex)
}

// TestMatchRows_ExtraRow verifies that matchRows reports an "extra" mismatch
// when a parquet row has no corresponding expected row.
//
// Phase B: This test FAILS because the stub returns (nil, nil).
func TestMatchRows_ExtraRow(t *testing.T) {
	parquetRows := []parquetRowData{
		{
			Metric:          "sum_metric",
			DimensionValues: map[string]string{"Region": "North"},
			GrainValues:     map[string]string{"Day": "Monday"},
			Value:           10.5,
		},
		{
			Metric:          "sum_metric",
			DimensionValues: map[string]string{"Region": "East"},
			GrainValues:     map[string]string{"Day": "Monday"},
			Value:           99.9,
		},
	}

	expected := &ExpectedOutput{
		Rows: []ExpectedRow{
			{
				Metric:          "sum_metric",
				DimensionValues: map[string]string{"Region": "North"},
				GrainValues:     map[string]string{"Day": "Monday"},
				Value:           10.5,
			},
		},
	}

	mismatches, err := matchRows(parquetRows, expected)
	require.NoError(t, err)
	require.Len(t, mismatches, 1, "One extra row in parquet not in expected")
	assert.Equal(t, "extra", mismatches[0].Type)
	assert.Equal(t, "East", mismatches[0].DimensionValues["Region"])
}

// TestMatchRows_ValueMismatch verifies that matchRows reports a "value_mismatch"
// when a row exists in both but the Value differs beyond the tolerance.
//
// Phase B: This test FAILS because the stub returns (nil, nil).
func TestMatchRows_ValueMismatch(t *testing.T) {
	parquetRows := []parquetRowData{
		{
			Metric:          "sum_metric",
			DimensionValues: map[string]string{"Region": "North"},
			GrainValues:     map[string]string{"Day": "Monday"},
			Value:           10.5,
		},
	}

	expected := &ExpectedOutput{
		Rows: []ExpectedRow{
			{
				Metric:          "sum_metric",
				DimensionValues: map[string]string{"Region": "North"},
				GrainValues:     map[string]string{"Day": "Monday"},
				Value:           999.0, // Very different from 10.5
			},
		},
	}

	mismatches, err := matchRows(parquetRows, expected)
	require.NoError(t, err)
	require.Len(t, mismatches, 1, "Value mismatch should be reported")
	assert.Equal(t, "value_mismatch", mismatches[0].Type)
	assert.InDelta(t, 999.0, mismatches[0].ExpectedValue, 1e-9)
	assert.InDelta(t, 10.5, mismatches[0].ActualValue, 1e-9)
}

// TestMatchRows_MultipleMetrics verifies that matchRows correctly distinguishes
// rows by their Metric column value, so rows with different metrics are not
// incorrectly matched.
//
// Phase B: This test FAILS because the stub returns (nil, nil).
func TestMatchRows_MultipleMetrics(t *testing.T) {
	parquetRows := []parquetRowData{
		{
			Metric:          "profit",
			DimensionValues: map[string]string{"Region": "North"},
			GrainValues:     map[string]string{"Day": "Monday"},
			Value:           100.0,
		},
		{
			Metric:          "margin_pct",
			DimensionValues: map[string]string{"Region": "North"},
			GrainValues:     map[string]string{"Day": "Monday"},
			Value:           20.0,
		},
	}

	expected := &ExpectedOutput{
		Rows: []ExpectedRow{
			{
				Metric:          "profit",
				DimensionValues: map[string]string{"Region": "North"},
				GrainValues:     map[string]string{"Day": "Monday"},
				Value:           100.0,
			},
			{
				Metric:          "margin_pct",
				DimensionValues: map[string]string{"Region": "North"},
				GrainValues:     map[string]string{"Day": "Monday"},
				Value:           20.0,
			},
		},
	}

	mismatches, err := matchRows(parquetRows, expected)
	require.NoError(t, err)
	assert.Empty(t, mismatches, "Rows with different metrics should be matched correctly")
}

// =============================================================================
// 5. Value Comparison Tests — compareValues()
// =============================================================================

// TestCompareValues_ExactlyEqual verifies that compareValues returns true when
// two float64 values are exactly equal within the default tolerance of 1e-9.
//
// Phase B: This test FAILS because the stub returns false.
func TestCompareValues_ExactlyEqual(t *testing.T) {
	assert.True(t, compareValues(10.5, 10.5, 1e-9), "Identical values should be equal")
}

// TestCompareValues_WithinTolerance verifies that compareValues returns true
// when values differ by less than the tolerance.
//
// Phase B: This test FAILS because the stub returns false.
func TestCompareValues_WithinTolerance(t *testing.T) {
	// Difference of 5e-10 is within 1e-9 tolerance
	assert.True(t, compareValues(10.5, 10.5+5e-10, 1e-9), "Values within tolerance should be equal")
	assert.True(t, compareValues(42.25, 42.25-3e-10, 1e-9), "Values within tolerance should be equal")
}

// TestCompareValues_BeyondTolerance verifies that compareValues returns false
// when values differ by more than the tolerance.
func TestCompareValues_BeyondTolerance(t *testing.T) {
	// Difference of 1e-8 is beyond 1e-9 tolerance
	assert.False(t, compareValues(10.5, 10.5+1e-8, 1e-9), "Values beyond tolerance should be unequal")
	assert.False(t, compareValues(0.0, 1e-7, 1e-9), "Values beyond tolerance should be unequal")
}

// TestCompareValues_AtToleranceBoundary verifies that compareValues returns true
// when the absolute difference exactly equals the tolerance.
//
// Phase B: This test FAILS because the stub returns false.
func TestCompareValues_AtToleranceBoundary(t *testing.T) {
	// Difference exactly equals tolerance -> should be true (|diff| <= tolerance)
	assert.True(t, compareValues(10.5, 10.5+1e-9, 1e-9), "Value at tolerance boundary should be equal")
	assert.True(t, compareValues(10.5, 10.5-1e-9, 1e-9), "Value at tolerance boundary should be equal")
}

// TestCompareValues_ZeroValues verifies that compareValues handles zero values
// correctly.
//
// Phase B: This test FAILS for the true assertions because the stub returns false.
func TestCompareValues_ZeroValues(t *testing.T) {
	assert.True(t, compareValues(0.0, 0.0, 1e-9), "Two zeros should be equal")
	assert.True(t, compareValues(0.0, 1e-15, 1e-9), "Near-zero within tolerance should be equal")
	assert.False(t, compareValues(0.0, 1e-6, 1e-9), "Zero vs non-zero beyond tolerance should be unequal")
}

// TestCompareValues_NegativeValues verifies that compareValues handles negative
// values correctly (absolute difference works regardless of sign).
//
// Phase B: This test FAILS for the true assertions because the stub returns false.
func TestCompareValues_NegativeValues(t *testing.T) {
	assert.True(t, compareValues(-10.5, -10.5, 1e-9), "Identical negative values should be equal")
	assert.True(t, compareValues(-10.5, -10.5+1e-10, 1e-9), "Negative values within tolerance should be equal")
	assert.False(t, compareValues(-10.5, -10.5-1e-8, 1e-9), "Negative values beyond tolerance should be unequal")
}

// TestCompareValues_CrossingZero verifies that compareValues handles comparisons
// that cross zero (one positive, one negative).
//
// Phase B: This test FAILS for the true assertions because the stub returns false.
func TestCompareValues_CrossingZero(t *testing.T) {
	assert.True(t, compareValues(1e-10, -1e-10, 1e-9), "Values crossing zero within tolerance should be equal")
	assert.False(t, compareValues(5.0, -5.0, 1e-9), "Values crossing zero beyond tolerance should be unequal")
}

// TestCompareValues_VerySmallValues verifies that compareValues handles very
// small float64 values (near subnormal range).
//
// Phase B: This test FAILS for the true assertions because the stub returns false.
func TestCompareValues_VerySmallValues(t *testing.T) {
	small := 1e-300
	assert.True(t, compareValues(small, small, 1e-9), "Identical very small values should be equal")
	assert.True(t, compareValues(small, small*1.000001, 1e-9), "Very small values within tolerance should be equal")
}

// TestCompareValues_CustomTolerance verifies that compareValues respects a
// custom tolerance value.
//
// Phase B: This test FAILS for the true assertions because the stub returns false.
func TestCompareValues_CustomTolerance(t *testing.T) {
	// Large tolerance: 0.1
	assert.True(t, compareValues(10.5, 10.55, 0.1), "Values within large tolerance should be equal")
	assert.False(t, compareValues(10.5, 10.7, 0.1), "Values beyond large tolerance should be unequal")

	// Tiny tolerance: 1e-15
	assert.True(t, compareValues(10.5, 10.5, 1e-15), "Exact match with tiny tolerance should be equal")
	assert.False(t, compareValues(10.5, 10.5+1e-12, 1e-15), "Tiny difference with tiny tolerance should be unequal")
}

// =============================================================================
// 6. Mismatch Reporting Tests — printDifferences()
// =============================================================================

// TestPrintDifferences_EmptySlice verifies that printDifferences does not panic
// when called with an empty mismatches slice.
func TestPrintDifferences_EmptySlice(t *testing.T) {
	// Should not panic
	assert.NotPanics(t, func() {
		printDifferences([]RowMismatch{})
	})
}

// TestPrintDifferences_NilSlice verifies that printDifferences does not panic
// when called with a nil slice.
func TestPrintDifferences_NilSlice(t *testing.T) {
	// Should not panic with nil slice
	assert.NotPanics(t, func() {
		printDifferences(nil)
	})
}

// TestPrintDifferences_WithMismatches verifies that printDifferences does not
// panic when called with mismatch records of all three types.
func TestPrintDifferences_WithMismatches(t *testing.T) {
	mismatches := []RowMismatch{
		{
			Type:             "missing",
			ExpectedRowIndex: 2,
			Metric:           "profit",
			DimensionValues:  map[string]string{"Region": "North"},
			GrainValues:      map[string]string{"Day": "Monday"},
			ExpectedValue:    100.0,
		},
		{
			Type:             "extra",
			ExpectedRowIndex: -1,
			Metric:           "profit",
			DimensionValues:  map[string]string{"Region": "South"},
			GrainValues:      map[string]string{"Day": "Monday"},
			ActualValue:      200.0,
		},
		{
			Type:             "value_mismatch",
			ExpectedRowIndex: 3,
			Metric:           "margin_pct",
			DimensionValues:  map[string]string{"Region": "North", "Product": "Widget"},
			GrainValues:      map[string]string{"Year": "2024"},
			ExpectedValue:    100.0,
			ActualValue:      99.999999,
		},
	}

	// Should not panic with populated mismatches
	assert.NotPanics(t, func() {
		printDifferences(mismatches)
	})
}

// =============================================================================
// 7. End-to-End Integration Tests — verify()
// =============================================================================

// writeExpectedJSON writes an ExpectedOutput struct to a temporary JSON file
// and returns the file path. This is a test helper.
func writeExpectedJSON(t *testing.T, dir string, eo *ExpectedOutput) string {
	t.Helper()
	data, err := json.MarshalIndent(eo, "", "  ")
	require.NoError(t, err)
	jsonPath := filepath.Join(dir, "expected_output.json")
	err = os.WriteFile(jsonPath, data, 0644)
	require.NoError(t, err)
	return jsonPath
}

// TestVerify_Success verifies that verify returns nil when the parquet file
// exactly matches the expected output specification.
//
// This test generates a parquet file using cplio.GenerateParquetFile with
// fixed values, then writes a matching expected_output.json, and finally
// calls verify() to confirm no error is returned.
//
// Phase B: This test PASSES because the stub returns nil.
func TestVerify_Success(t *testing.T) {
	tempDir := t.TempDir()
	parquetPath := filepath.Join(tempDir, "test.parquet")

	// Generate a parquet file with known values.
	// Include "Metric" column to match the expected output schema.
	dimensions := map[string][]string{
		"Metric": {"sum_metric"},
		"Region": {"North", "South"},
	}
	grain := []cplio.GrainMeasure{
		{Name: "Day", Values: []string{"Monday"}},
	}
	values := []float64{10.5, 42.25}

	err := cplio.GenerateParquetFile(parquetPath, dimensions, grain, values)
	require.NoError(t, err)

	// Write expected JSON matching the generated parquet
	expected := &ExpectedOutput{
		OutputFile: "test.parquet",
		Metrics:    []string{"sum_metric"},
		Dimensions: []string{"Region"},
		Grain:      []string{"Day"},
		Rows: []ExpectedRow{
			{
				Metric:          "sum_metric",
				DimensionValues: map[string]string{"Region": "North"},
				GrainValues:     map[string]string{"Day": "Monday"},
				Value:           10.5,
			},
			{
				Metric:          "sum_metric",
				DimensionValues: map[string]string{"Region": "South"},
				GrainValues:     map[string]string{"Day": "Monday"},
				Value:           42.25,
			},
		},
	}

	expectedJSONPath := writeExpectedJSON(t, tempDir, expected)

	// Call verify - should succeed for matching data
	err = verify(parquetPath, expectedJSONPath)
	assert.NoError(t, err, "verify() should return nil when all data matches")
}

// TestVerify_Mismatch verifies that verify returns an error when the parquet
// data differs from the expected output specification.
//
// Phase B: This test FAILS because the stub returns nil.
func TestVerify_Mismatch(t *testing.T) {
	tempDir := t.TempDir()
	parquetPath := filepath.Join(tempDir, "test.parquet")

	dimensions := map[string][]string{
		"Metric": {"sum_metric"},
		"Region": {"North", "South"},
	}
	grain := []cplio.GrainMeasure{
		{Name: "Day", Values: []string{"Monday"}},
	}
	values := []float64{10.5, 42.25}

	err := cplio.GenerateParquetFile(parquetPath, dimensions, grain, values)
	require.NoError(t, err)

	// Write expected JSON with values that DON'T match the generated parquet
	expected := &ExpectedOutput{
		OutputFile: "test.parquet",
		Metrics:    []string{"sum_metric"},
		Dimensions: []string{"Region"},
		Grain:      []string{"Day"},
		Rows: []ExpectedRow{
			{
				Metric:          "sum_metric",
				DimensionValues: map[string]string{"Region": "North"},
				GrainValues:     map[string]string{"Day": "Monday"},
				Value:           999.0, // Mismatch: generated as 10.5
			},
			{
				Metric:          "sum_metric",
				DimensionValues: map[string]string{"Region": "South"},
				GrainValues:     map[string]string{"Day": "Monday"},
				Value:           888.0, // Mismatch: generated as 42.25
			},
		},
	}

	expectedJSONPath := writeExpectedJSON(t, tempDir, expected)

	err = verify(parquetPath, expectedJSONPath)
	assert.Error(t, err, "verify() should return an error when data mismatches")
}

// TestVerify_MissingFile verifies that verify returns an error when a required
// file does not exist.
//
// Phase B: This test FAILS because the stub returns nil.
func TestVerify_MissingFile(t *testing.T) {
	tempDir := t.TempDir()

	err := verify(
		filepath.Join(tempDir, "nonexistent.parquet"),
		filepath.Join(tempDir, "nonexistent.json"),
	)
	assert.Error(t, err, "verify() should return an error when files are missing")
}

// TestVerify_InvalidJSON verifies that verify returns an error when the
// expected JSON file contains invalid JSON.
//
// Phase B: This test FAILS because the stub returns nil.
func TestVerify_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	parquetPath := filepath.Join(tempDir, "test.parquet")

	dimensions := map[string][]string{
		"Region": {"North"},
	}
	err := cplio.GenerateParquetFile(parquetPath, dimensions, nil, []float64{10.5})
	require.NoError(t, err)

	// Write invalid JSON
	jsonPath := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(jsonPath, []byte("not valid json"), 0644)
	require.NoError(t, err)

	err = verify(parquetPath, jsonPath)
	assert.Error(t, err, "verify() should return an error when JSON is invalid")
}

// =============================================================================
// 8. Exit Code / CLI Integration Tests — main() via exec.Command
// =============================================================================

// buildParquetVerifyBinary compiles the parquet_verify binary and returns the
// path to the compiled executable. This is a test helper.
func buildParquetVerifyBinary(t *testing.T, tempDir string) string {
	t.Helper()
	binaryPath := filepath.Join(tempDir, "parquet_verify")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildOutput, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build parquet_verify binary: %s", string(buildOutput))
	return binaryPath
}

// TestMain_Success verifies that the parquet_verify CLI exits with code 0
// when the parquet file matches the expected output specification.
//
// Phase B: This test FAILS because the implementation is a stub that exits 0
// regardless, so the verification logic isn't actually checking anything.
// The exit code 0 assertion will PASS trivially but the test serves as
// documentation of the expected behavior.
func TestMain_Success(t *testing.T) {
	tempDir := t.TempDir()
	parquetPath := filepath.Join(tempDir, "test.parquet")

	dimensions := map[string][]string{
		"Metric": {"sum_metric"},
		"Region": {"North", "South"},
	}
	grain := []cplio.GrainMeasure{
		{Name: "Day", Values: []string{"Monday"}},
	}
	values := []float64{10.5, 42.25}

	err := cplio.GenerateParquetFile(parquetPath, dimensions, grain, values)
	require.NoError(t, err)

	expected := &ExpectedOutput{
		OutputFile: "test.parquet",
		Metrics:    []string{"sum_metric"},
		Dimensions: []string{"Region"},
		Grain:      []string{"Day"},
		Rows: []ExpectedRow{
			{Metric: "sum_metric", DimensionValues: map[string]string{"Region": "North"}, GrainValues: map[string]string{"Day": "Monday"}, Value: 10.5},
			{Metric: "sum_metric", DimensionValues: map[string]string{"Region": "South"}, GrainValues: map[string]string{"Day": "Monday"}, Value: 42.25},
		},
	}
	expectedJSONPath := writeExpectedJSON(t, tempDir, expected)

	binaryPath := buildParquetVerifyBinary(t, tempDir)

	cmd := exec.Command(binaryPath, "-parquet="+parquetPath, "-expected="+expectedJSONPath)
	output, err := cmd.CombinedOutput()
	assert.NoError(t, err, "Matching data should exit 0; output: %s", string(output))
}

// TestMain_Mismatch verifies that the parquet_verify CLI exits with code 1
// when differences are found between the parquet and expected output.
//
// Phase B: This test FAILS because the stub returns nil from verify(), so
// main() exits 0 regardless of content. The assertion for exit code 1
// will fail until Phase C is implemented.
func TestMain_Mismatch(t *testing.T) {
	tempDir := t.TempDir()
	parquetPath := filepath.Join(tempDir, "test.parquet")

	dimensions := map[string][]string{
		"Metric": {"sum_metric"},
		"Region": {"North"},
	}
	values := []float64{10.5}
	err := cplio.GenerateParquetFile(parquetPath, dimensions, nil, values)
	require.NoError(t, err)

	// Write mismatched expected JSON
	expected := &ExpectedOutput{
		OutputFile: "test.parquet",
		Metrics:    []string{"sum_metric"},
		Dimensions: []string{"Region"},
		Grain:      []string{},
		Rows: []ExpectedRow{
			{Metric: "sum_metric", DimensionValues: map[string]string{"Region": "North"}, GrainValues: map[string]string{}, Value: 999.0},
		},
	}
	expectedJSONPath := writeExpectedJSON(t, tempDir, expected)

	binaryPath := buildParquetVerifyBinary(t, tempDir)

	cmd := exec.Command(binaryPath, "-parquet="+parquetPath, "-expected="+expectedJSONPath)
	output, err := cmd.CombinedOutput()

	// Phase B: err will be nil because stub returns nil; this assertion FAILS
	assert.Error(t, err, "Mismatched data should exit with non-zero code")
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			assert.Equal(t, 1, exitErr.ExitCode(), "Mismatch should exit with code 1")
		}
	}
	_ = output
}

// TestMain_MissingFlags verifies that the parquet_verify CLI exits with code 2
// when required command-line flags are missing.
//
// Phase B: This test PASSES because main() validates flags and exits 2.
func TestMain_MissingFlags(t *testing.T) {
	tempDir := t.TempDir()
	binaryPath := buildParquetVerifyBinary(t, tempDir)

	// Run with no flags
	cmd := exec.Command(binaryPath)
	output, err := cmd.CombinedOutput()

	assert.Error(t, err, "Missing required flags should produce an error")
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok, "Error should be ExitError")
	assert.Equal(t, 2, exitErr.ExitCode(), "Missing flags should exit with code 2")
	assert.Contains(t, string(output), "Usage", "Should print usage information")
}

// TestMain_MissingParquetFile verifies that the parquet_verify CLI exits with
// code 2 when the specified parquet file does not exist.
//
// Phase B: This test FAILS because the stub verify() returns nil, so main()
// exits 0 regardless. It will pass once Phase C file reading is implemented.
func TestMain_MissingParquetFile(t *testing.T) {
	tempDir := t.TempDir()

	// Write a valid expected JSON
	expected := &ExpectedOutput{
		OutputFile: "nonexistent.parquet",
		Metrics:    []string{"sum_metric"},
		Dimensions: []string{"Region"},
		Grain:      []string{},
		Rows:       []ExpectedRow{},
	}
	expectedJSONPath := writeExpectedJSON(t, tempDir, expected)

	binaryPath := buildParquetVerifyBinary(t, tempDir)

	cmd := exec.Command(binaryPath,
		"-parquet="+filepath.Join(tempDir, "nonexistent.parquet"),
		"-expected="+expectedJSONPath,
	)
	output, err := cmd.CombinedOutput()

	assert.Error(t, err, "Missing parquet file should produce an error")
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			assert.Equal(t, 2, exitErr.ExitCode(), "Missing file should exit with code 2")
		}
	}
	_ = output
}

// TestMain_InvalidExpectedJSON verifies that the parquet_verify CLI exits with
// code 2 when the expected JSON file contains malformed content.
//
// Phase B: This test FAILS because the stub verify() returns nil.
func TestMain_InvalidExpectedJSON(t *testing.T) {
	tempDir := t.TempDir()
	parquetPath := filepath.Join(tempDir, "test.parquet")

	dimensions := map[string][]string{
		"Metric": {"sum_metric"},
		"Region": {"North"},
	}
	err := cplio.GenerateParquetFile(parquetPath, dimensions, nil, []float64{10.5})
	require.NoError(t, err)

	// Write invalid JSON
	invalidJSONPath := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidJSONPath, []byte("{not valid"), 0644)
	require.NoError(t, err)

	binaryPath := buildParquetVerifyBinary(t, tempDir)

	cmd := exec.Command(binaryPath, "-parquet="+parquetPath, "-expected="+invalidJSONPath)
	output, err := cmd.CombinedOutput()

	assert.Error(t, err, "Invalid JSON should produce an error")
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			assert.Equal(t, 2, exitErr.ExitCode(), "Invalid JSON should exit with code 2")
		}
	}
	_ = output
}
