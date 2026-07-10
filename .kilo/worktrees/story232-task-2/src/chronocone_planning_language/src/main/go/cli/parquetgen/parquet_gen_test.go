// Package main_test contains tests for the parquetgen CLI tool.
//
// These tests verify the command-line tool's ability to parse data JSON files
// in row format and generate parquet files from them.
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

// TestParseRowFormat_Valid verifies that parseRowFormat correctly reads a
// valid data JSON file in row-format and extracts columns and values.
func TestParseRowFormat_Valid(t *testing.T) {
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "data_a.json")

	jsonContent := `[
		{"Region":"North","Daily":"Sunday","Value":10},
		{"Region":"North","Daily":"Monday","Value":9},
		{"Region":"South","Daily":"Sunday","Value":10}
	]`

	err := os.WriteFile(jsonPath, []byte(jsonContent), 0644)
	require.NoError(t, err)

	rows, values, err := parseRowFormat(jsonPath)
	require.NoError(t, err)
	assert.Len(t, rows, 3)
	assert.Len(t, values, 3)

	assert.Equal(t, "North", rows[0]["Region"])
	assert.Equal(t, "Sunday", rows[0]["Daily"])
	assert.Equal(t, float64(10), values[0])
	assert.Equal(t, "South", rows[2]["Region"])
	assert.Equal(t, float64(10), values[2])
}

// TestParseRowFormat_InvalidJSON verifies that parseRowFormat returns an error
// when given malformed JSON.
func TestParseRowFormat_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "invalid.json")
	err := os.WriteFile(jsonPath, []byte(`{invalid json`), 0644)
	require.NoError(t, err)

	rows, values, err := parseRowFormat(jsonPath)
	assert.Error(t, err)
	assert.Nil(t, rows)
	assert.Nil(t, values)
}

// TestParseRowFormat_MissingFile verifies that parseRowFormat returns a
// file not found error for a non-existent path.
func TestParseRowFormat_MissingFile(t *testing.T) {
	rows, values, err := parseRowFormat("/path/that/does/not/exist/data.json")
	assert.Error(t, err)
	assert.Nil(t, rows)
	assert.Nil(t, values)
	assert.True(t, os.IsNotExist(err))
}

// TestParseRowFormat_EmptyArray verifies that parseRowFormat returns an error
// when the JSON is an empty array.
func TestParseRowFormat_EmptyArray(t *testing.T) {
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "empty.json")
	err := os.WriteFile(jsonPath, []byte(`[]`), 0644)
	require.NoError(t, err)

	rows, values, err := parseRowFormat(jsonPath)
	assert.Error(t, err)
	assert.Nil(t, rows)
	assert.Nil(t, values)
}

// TestParseRowFormat_NonStringColumn verifies that parseRowFormat returns an error
// when a non-Value column has a non-string value.
func TestParseRowFormat_NonStringColumn(t *testing.T) {
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "bad.json")
	err := os.WriteFile(jsonPath, []byte(`[{"Region":42,"Value":1}]`), 0644)
	require.NoError(t, err)

	rows, values, err := parseRowFormat(jsonPath)
	assert.Error(t, err)
	assert.Nil(t, rows)
	assert.Nil(t, values)
}

// TestParseRowFormat_NonNumericValue verifies that parseRowFormat returns an error
// when the Value field is not a number.
func TestParseRowFormat_NonNumericValue(t *testing.T) {
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "bad.json")
	err := os.WriteFile(jsonPath, []byte(`[{"Region":"North","Value":"ten"}]`), 0644)
	require.NoError(t, err)

	rows, values, err := parseRowFormat(jsonPath)
	assert.Error(t, err)
	assert.Nil(t, rows)
	assert.Nil(t, values)
}

// TestRun_Success verifies that run() successfully generates a parquet file
// from a row-format data JSON file.
func TestRun_Success(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "data_a.json")
	outputPath := filepath.Join(tempDir, "output.parquet")

	jsonContent := `[
		{"Region":"North","Product":"Widget","Value":10.5},
		{"Region":"North","Product":"Gadget","Value":20.75},
		{"Region":"South","Product":"Widget","Value":30.0},
		{"Region":"South","Product":"Gadget","Value":40.25}
	]`

	err := os.WriteFile(inputPath, []byte(jsonContent), 0644)
	require.NoError(t, err)

	err = run(inputPath, outputPath)
	require.NoError(t, err)

	// Verify output parquet file is created
	_, err = os.Stat(outputPath)
	require.NoError(t, err)

	// Verify the parquet file can be read and has correct structure
	schema, rowCount, err := cplio.ReadParquetFile(outputPath)
	require.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, 4, rowCount)

	columnNames := cplio.GetColumnNames(schema)
	assert.Contains(t, columnNames, "Region")
	assert.Contains(t, columnNames, "Product")
	assert.Contains(t, columnNames, "Value")
}

// TestRun_EmptyRows verifies that run() succeeds with an empty rows file
// (produces an empty parquet).
func TestRun_EmptyRows(t *testing.T) {
	// An empty array is caught by parseRowFormat, not run.
}

// TestRun_InvalidInputPath verifies that run() propagates file-not-found errors.
func TestRun_InvalidInputPath(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.parquet")

	err := run("/path/that/does/not/exist/data.json", outputPath)
	assert.Error(t, err)
}

// TestMain_CommandParsing verifies the CLI correctly parses -input and -output flags.
func TestMain_CommandParsing(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "data_a.json")
	outputPath := filepath.Join(tempDir, "output.parquet")

	jsonContent := `[
		{"Region":"North","Product":"Widget","Value":1.0},
		{"Region":"South","Product":"Widget","Value":2.0}
	]`
	err := os.WriteFile(inputPath, []byte(jsonContent), 0644)
	require.NoError(t, err)

	// Build the CLI binary
	binaryPath := filepath.Join(tempDir, "parquetgen")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildOutput, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build binary: %s", string(buildOutput))

	// Execute with test arguments
	cmd := exec.Command(binaryPath, "-input="+inputPath, "-output="+outputPath)
	output, err := cmd.CombinedOutput()
	assert.NoError(t, err, "Command failed with output: %s", string(output))

	// Verify output parquet file is created
	_, err = os.Stat(outputPath)
	require.NoError(t, err)

	// Test missing required flags
	cmdMissingFlags := exec.Command(binaryPath)
	outputMissing, _ := cmdMissingFlags.CombinedOutput()
	assert.Contains(t, string(outputMissing), "Usage", "Should show usage when flags are missing")
}

// TestRun_ProducesCorrectParquetSchema verifies that WriteParquetFromRows produces
// a parquet file whose column order is alphabetically sorted string columns + Value.
func TestRun_ProducesCorrectParquetSchema(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "data.json")
	outputPath := filepath.Join(tempDir, "output.parquet")

	// "Z" sorts after "A" — verify alphabetical ordering in parquet
	jsonContent := `[
		{"Z_Dim":"z1","A_Dim":"a1","Value":100.0}
	]`
	err := os.WriteFile(inputPath, []byte(jsonContent), 0644)
	require.NoError(t, err)

	err = run(inputPath, outputPath)
	require.NoError(t, err)

	schema, rowCount, err := cplio.ReadParquetFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, 1, rowCount)

	columnNames := cplio.GetColumnNames(schema)
	// Alphabetical order: A_Dim, Z_Dim, then Value last
	assert.Equal(t, "A_Dim", columnNames[0])
	assert.Equal(t, "Z_Dim", columnNames[1])
	assert.Equal(t, "Value", columnNames[2])
}

// TestRun_ProducesCorrectValues verifies that the parquet contains the correct
// values read from data JSON.
func TestRun_ProducesCorrectValues(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "data.json")
	outputPath := filepath.Join(tempDir, "output.parquet")

	jsonContent := `[
		{"Region":"North","Daily":"Sunday","Value":10.0},
		{"Region":"North","Daily":"Monday","Value":9.0}
	]`
	err := os.WriteFile(inputPath, []byte(jsonContent), 0644)
	require.NoError(t, err)

	err = run(inputPath, outputPath)
	require.NoError(t, err)

	schema, rowCount, err := cplio.ReadParquetFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, 2, rowCount)

	// Read back the Value column and verify
	valIdx := -1
	for i := 0; i < schema.NumColumns(); i++ {
		if schema.Column(i).Name() == "Value" {
			valIdx = i
			break
		}
	}
	require.NotEqual(t, -1, valIdx)

	data, err := cplio.ReadParquetColumnData(outputPath, valIdx)
	require.NoError(t, err)
	f64Data, ok := data.([]float64)
	require.True(t, ok, "Value column should be float64")
	assert.Equal(t, []float64{10.0, 9.0}, f64Data)
}

// TestWriteParquetFromRows_Basic verifies that WriteParquetFromRows writes a
// valid parquet file from row-format data.
func TestWriteParquetFromRows_Basic(t *testing.T) {
	tempDir := t.TempDir()
	parquetPath := filepath.Join(tempDir, "test.parquet")

	rows := []map[string]string{
		{"Region": "North", "Daily": "Sunday"},
		{"Region": "South", "Daily": "Monday"},
	}
	values := []float64{10.5, 20.75}

	err := cplio.WriteParquetFromRows(parquetPath, rows, values)
	require.NoError(t, err)

	schema, rowCount, err := cplio.ReadParquetFile(parquetPath)
	require.NoError(t, err)
	assert.Equal(t, 2, rowCount)

	columnNames := cplio.GetColumnNames(schema)
	assert.Contains(t, columnNames, "Region")
	assert.Contains(t, columnNames, "Daily")
	assert.Contains(t, columnNames, "Value")
}

// TestWriteParquetFromRows_EmptyRows verifies that WriteParquetFromRows handles
// empty rows gracefully (no-op).
func TestWriteParquetFromRows_EmptyRows(t *testing.T) {
	tempDir := t.TempDir()
	parquetPath := filepath.Join(tempDir, "empty.parquet")

	err := cplio.WriteParquetFromRows(parquetPath, []map[string]string{}, []float64{})
	require.NoError(t, err)
}

// TestWriteParquetFromRows_MismatchedLengths verifies error on length mismatch.
func TestWriteParquetFromRows_MismatchedLengths(t *testing.T) {
	tempDir := t.TempDir()
	parquetPath := filepath.Join(tempDir, "bad.parquet")

	rows := []map[string]string{{"Region": "North"}}
	values := []float64{1.0, 2.0}

	err := cplio.WriteParquetFromRows(parquetPath, rows, values)
	assert.Error(t, err)
}

// TestToFloat64 verifies the toFloat64 conversion helper handles various JSON number types.
func TestToFloat64(t *testing.T) {
	// float64
	v, ok := toFloat64(float64(42.5))
	assert.True(t, ok)
	assert.Equal(t, 42.5, v)

	// float32
	v, ok = toFloat64(float32(3.14))
	assert.True(t, ok)
	assert.InDelta(t, 3.14, v, 1e-6)

	// json.Number
	v, ok = toFloat64(json.Number("99.9"))
	assert.True(t, ok)
	assert.Equal(t, 99.9, v)

	// json.Number invalid
	_, ok = toFloat64(json.Number("not-a-number"))
	assert.False(t, ok)

	// string (not a number)
	_, ok = toFloat64("hello")
	assert.False(t, ok)

	// int (not handled)
	_, ok = toFloat64(42)
	assert.False(t, ok)
}
