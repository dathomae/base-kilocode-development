// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package io

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/arrow/go/v16/parquet"
	"github.com/apache/arrow/go/v16/parquet/file"
	"github.com/apache/arrow/go/v16/parquet/schema"
	"github.com/stretchr/testify/require"
)

// writeCustomParquet writes a parquet file with the specified string columns
// and a Value column. Each row in rows provides the string values for the
// corresponding columns, and values provides the float64 Value for each row.
func writeCustomParquet(t *testing.T, filePath string, columns []string, rows []map[string]string, values []float64) {
	t.Helper()

	if len(rows) != len(values) {
		t.Fatalf("rows length (%d) must match values length (%d)", len(rows), len(values))
	}

	// Build schema nodes: one ByteArray per column, then Float64 Value.
	nodeList := make(schema.FieldList, 0, len(columns)+1)
	for _, name := range columns {
		node := schema.NewByteArrayNode(name, parquet.Repetitions.Required, -1)
		nodeList = append(nodeList, node)
	}
	valueNode := schema.NewFloat64Node("Value", parquet.Repetitions.Required, -1)
	nodeList = append(nodeList, valueNode)

	groupNode, err := schema.NewGroupNode("schema", parquet.Repetitions.Required, nodeList, -1)
	require.NoError(t, err, "failed to create schema group node")

	outputFile, err := os.Create(filePath)
	require.NoError(t, err, "failed to create output file")
	defer outputFile.Close()

	props := parquet.NewWriterProperties()
	writer := file.NewParquetWriter(outputFile, groupNode, file.WithWriterProps(props))
	defer writer.Close()

	rowGroupWriter := writer.AppendRowGroup()

	// Write each string column.
	for _, name := range columns {
		columnWriter, err := rowGroupWriter.NextColumn()
		require.NoError(t, err, "failed to get column writer for %q", name)

		stringWriter := columnWriter.(*file.ByteArrayColumnChunkWriter)
		strValues := make([]parquet.ByteArray, len(rows))
		for i, row := range rows {
			strValues[i] = parquet.ByteArray(row[name])
		}
		_, err = stringWriter.WriteBatch(strValues, nil, nil)
		require.NoError(t, err, "failed to write string column %q", name)
	}

	// Write Value column.
	valueColumnWriter, err := rowGroupWriter.NextColumn()
	require.NoError(t, err, "failed to get value column writer")

	floatWriter := valueColumnWriter.(*file.Float64ColumnChunkWriter)
	_, err = floatWriter.WriteBatch(values, nil, nil)
	require.NoError(t, err, "failed to write value column")

	require.NoError(t, rowGroupWriter.Close(), "failed to close row group writer")
}

// TestReadParquetToTensor_NonExistentFile returns an error when the
// specified parquet file does not exist.
func TestReadParquetToTensor_NonExistentFile(t *testing.T) {
	t.Parallel()
	conceptValues := map[string][]string{
		"Year":   {"2023", "2024"},
		"Region": {"North", "South", "East", "West"},
	}
	dimensions := []string{"Year", "Region"}

	_, _, _, err := ReadParquetToTensor("/nonexistent/path/file.parquet", dimensions, nil, conceptValues)

	require.Error(t, err, "ReadParquetToTensor should return an error for non-existent file")
}

// TestReadParquetToTensor_ExactMatchSucceeds verifies that a parquet file
// with exactly the right columns, exactly the right row count, and all
// unique combinations loads without error.
func TestReadParquetToTensor_ExactMatchSucceeds(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "exact.parquet")

	conceptValues := map[string][]string{
		"Year":   {"2023", "2024"},
		"Region": {"North", "South", "East", "West"},
	}
	dimensions := []string{"Year", "Region"}

	// Generate a parquet file that exactly matches the concept values.
	dimMap := map[string][]string{
		"Year":   {"2023", "2024"},
		"Region": {"North", "South", "East", "West"},
	}
	err := GenerateParquetFile(filePath, dimMap, nil, nil)
	require.NoError(t, err, "GenerateParquetFile should succeed")

	data, dimNames, shape, err := ReadParquetToTensor(filePath, dimensions, nil, conceptValues)
	require.NoError(t, err, "ReadParquetToTensor should succeed for exact-match file")
	require.Equal(t, dimensions, dimNames, "dimension names should match")
	require.Equal(t, []int{2, 4}, shape, "shape should be [2, 4]")
	require.Len(t, data, 8, "data should have 8 elements")
}

// TestReadParquetToTensor_MissingDataSucceeds verifies that a parquet file
// with fewer rows than the expected Cartesian product loads successfully,
// with populated cells retaining their values and unpopulated cells set to 0.0.
func TestReadParquetToTensor_MissingDataSucceeds(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "missing.parquet")

	// Write a parquet file with only 2 of the 4 expected rows.
	columns := []string{"Year", "Region"}
	rows := []map[string]string{
		{"Year": "2023", "Region": "North"},
		{"Year": "2024", "Region": "South"},
	}
	values := []float64{10.0, 20.0}
	writeCustomParquet(t, filePath, columns, rows, values)

	conceptValues := map[string][]string{
		"Year":   {"2023", "2024"},
		"Region": {"North", "South"},
	}
	dimensions := []string{"Year", "Region"}

	data, dimNames, shape, err := ReadParquetToTensor(filePath, dimensions, nil, conceptValues)
	require.NoError(t, err, "ReadParquetToTensor should succeed for missing data")
	require.Equal(t, dimensions, dimNames, "dimension names should match")
	require.Equal(t, []int{2, 2}, shape, "shape should be [2, 2]")
	require.Len(t, data, 4, "data should have 4 elements")
	// Year=2023, Region=North => flatIndex 0
	require.Equal(t, 10.0, data[0], "populated cell should have correct value")
	// Year=2023, Region=South => flatIndex 1 (missing)
	require.Equal(t, 0.0, data[1], "missing cell should be zero-initialized")
	// Year=2024, Region=North => flatIndex 2 (missing)
	require.Equal(t, 0.0, data[2], "missing cell should be zero-initialized")
	// Year=2024, Region=South => flatIndex 3
	require.Equal(t, 20.0, data[3], "populated cell should have correct value")
}

// TestReadParquetToTensor_DuplicateRowsReturnsError verifies that a parquet
// file containing a duplicate combination and one missing combination returns
// a duplicate error (detected first) or missing error.
func TestReadParquetToTensor_DuplicateRowsReturnsError(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "duplicate.parquet")

	columns := []string{"Year", "Region"}
	rows := []map[string]string{
		{"Year": "2023", "Region": "North"},
		{"Year": "2023", "Region": "North"}, // duplicate
		{"Year": "2024", "Region": "North"},
		{"Year": "2024", "Region": "South"},
	}
	values := []float64{1.0, 2.0, 3.0, 4.0}
	writeCustomParquet(t, filePath, columns, rows, values)

	conceptValues := map[string][]string{
		"Year":   {"2023", "2024"},
		"Region": {"North", "South"},
	}
	dimensions := []string{"Year", "Region"}

	_, _, _, err := ReadParquetToTensor(filePath, dimensions, nil, conceptValues)
	require.Error(t, err, "ReadParquetToTensor should return an error for duplicate rows")
	errStr := err.Error()
	require.Contains(t, errStr, "duplicate", "error should indicate duplicate combination; got: %s", errStr)
}

// TestReadParquetToTensor_ExtraRowsReturnsError verifies that a parquet file
// with all expected combinations plus one extra duplicate row returns an error
// indicating the actual row count exceeds the expected count.
func TestReadParquetToTensor_ExtraRowsReturnsError(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "extra.parquet")

	columns := []string{"Year", "Region"}
	rows := []map[string]string{
		{"Year": "2023", "Region": "North"},
		{"Year": "2023", "Region": "South"},
		{"Year": "2024", "Region": "North"},
		{"Year": "2024", "Region": "South"},
		{"Year": "2023", "Region": "North"}, // extra duplicate
	}
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	writeCustomParquet(t, filePath, columns, rows, values)

	conceptValues := map[string][]string{
		"Year":   {"2023", "2024"},
		"Region": {"North", "South"},
	}
	dimensions := []string{"Year", "Region"}

	_, _, _, err := ReadParquetToTensor(filePath, dimensions, nil, conceptValues)
	require.Error(t, err, "ReadParquetToTensor should return an error for extra rows")
	errStr := err.Error()
	require.Contains(t, errStr, "expected", "error should mention expected count")
	require.Contains(t, errStr, "found", "error should mention found count")
}

// TestReadParquetToTensor_ExtraColumnsReturnsError verifies that a parquet
// file with an extra unexpected column returns an error listing the extra column.
func TestReadParquetToTensor_ExtraColumnsReturnsError(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "extra_col.parquet")

	// Write a parquet with an extra "Product" column.
	columns := []string{"Year", "Region", "Product"}
	rows := []map[string]string{
		{"Year": "2023", "Region": "North", "Product": "A"},
		{"Year": "2023", "Region": "South", "Product": "A"},
		{"Year": "2024", "Region": "North", "Product": "A"},
		{"Year": "2024", "Region": "South", "Product": "A"},
	}
	values := []float64{1.0, 2.0, 3.0, 4.0}
	writeCustomParquet(t, filePath, columns, rows, values)

	conceptValues := map[string][]string{
		"Year":   {"2023", "2024"},
		"Region": {"North", "South"},
	}
	dimensions := []string{"Year", "Region"}

	_, _, _, err := ReadParquetToTensor(filePath, dimensions, nil, conceptValues)
	require.Error(t, err, "ReadParquetToTensor should return an error for extra columns")
	require.Contains(t, err.Error(), "Product", "error should list the extra column name")
}

// TestReadParquetToTensor_ExplicitZeroValueSucceeds verifies that a parquet
// file where one combination has value 0.0 loads successfully, proving that
// explicit zeros are distinguishable from missing data.
func TestReadParquetToTensor_ExplicitZeroValueSucceeds(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "zero.parquet")

	columns := []string{"Year", "Region"}
	rows := []map[string]string{
		{"Year": "2023", "Region": "North"},
		{"Year": "2023", "Region": "South"},
		{"Year": "2024", "Region": "North"},
		{"Year": "2024", "Region": "South"},
	}
	values := []float64{1.0, 0.0, 3.0, 4.0}
	writeCustomParquet(t, filePath, columns, rows, values)

	conceptValues := map[string][]string{
		"Year":   {"2023", "2024"},
		"Region": {"North", "South"},
	}
	dimensions := []string{"Year", "Region"}

	data, dimNames, shape, err := ReadParquetToTensor(filePath, dimensions, nil, conceptValues)
	require.NoError(t, err, "ReadParquetToTensor should succeed with explicit zero value")
	require.Equal(t, dimensions, dimNames, "dimension names should match")
	require.Equal(t, []int{2, 2}, shape, "shape should be [2, 2]")
	require.Len(t, data, 4, "data should have 4 elements")
	// Year=2023 (idx 0), Region=South (idx 1) => flatIndex = 0*2 + 1 = 1
	require.Equal(t, 0.0, data[1], "explicit zero value should be preserved")
}

// TestReadParquetToTensor_MissingDataMultipleRegionsSucceeds verifies that a
// parquet file with fewer rows than the expected Cartesian product loads
// successfully when the gap spans multiple dimension values.
func TestReadParquetToTensor_MissingDataMultipleRegionsSucceeds(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "missing_detail.parquet")

	// Write a parquet file with 2 regions, but expect 3.
	columns := []string{"Year", "Region"}
	rows := []map[string]string{
		{"Year": "2023", "Region": "North"},
		{"Year": "2023", "Region": "South"},
		{"Year": "2024", "Region": "North"},
		{"Year": "2024", "Region": "South"},
	}
	values := []float64{1.0, 2.0, 3.0, 4.0}
	writeCustomParquet(t, filePath, columns, rows, values)

	conceptValues := map[string][]string{
		"Year":   {"2023", "2024"},
		"Region": {"North", "South", "East"},
	}
	dimensions := []string{"Year", "Region"}

	data, dimNames, shape, err := ReadParquetToTensor(filePath, dimensions, nil, conceptValues)
	require.NoError(t, err, "ReadParquetToTensor should succeed for missing data")
	require.Equal(t, dimensions, dimNames, "dimension names should match")
	require.Equal(t, []int{2, 3}, shape, "shape should be [2, 3]")
	require.Len(t, data, 6, "data should have 6 elements")
	// Verify populated cells and zero-initialized missing cells.
	require.Equal(t, 1.0, data[0], "Year=2023, Region=North should be populated")
	require.Equal(t, 2.0, data[1], "Year=2023, Region=South should be populated")
	require.Equal(t, 0.0, data[2], "Year=2023, Region=East should be zero-initialized")
	require.Equal(t, 3.0, data[3], "Year=2024, Region=North should be populated")
	require.Equal(t, 4.0, data[4], "Year=2024, Region=South should be populated")
	require.Equal(t, 0.0, data[5], "Year=2024, Region=East should be zero-initialized")
}

// TestReadParquetToTensor_ZeroRowFileSucceeds verifies that a parquet file
// with zero rows loads successfully, producing a tensor with all zero values.
func TestReadParquetToTensor_ZeroRowFileSucceeds(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "empty.parquet")

	columns := []string{"Year", "Region"}
	rows := []map[string]string{}
	values := []float64{}
	writeCustomParquet(t, filePath, columns, rows, values)

	conceptValues := map[string][]string{
		"Year":   {"2023", "2024"},
		"Region": {"North", "South"},
	}
	dimensions := []string{"Year", "Region"}

	data, dimNames, shape, err := ReadParquetToTensor(filePath, dimensions, nil, conceptValues)
	require.NoError(t, err, "ReadParquetToTensor should succeed for zero-row file")
	require.Equal(t, dimensions, dimNames, "dimension names should match")
	require.Equal(t, []int{2, 2}, shape, "shape should be [2, 2]")
	require.Len(t, data, 4, "data should have 4 elements")
	for i, v := range data {
		require.Equal(t, 0.0, v, "element %d should be zero-initialized", i)
	}
}
