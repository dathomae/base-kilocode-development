// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package io

import (
	"path/filepath"
	"testing"

	"github.com/apache/arrow/go/v16/parquet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// flatIndexToDimensionIndices converts a flat row-major index into per-dimension
// indices, assuming the last dimension varies fastest.
func flatIndexToDimensionIndices(flatIndex int, shape []int) []int {
	indices := make([]int, len(shape))
	remainder := flatIndex
	for d := len(shape) - 1; d >= 0; d-- {
		indices[d] = remainder % shape[d]
		remainder /= shape[d]
	}
	return indices
}

// TestWriteTensorToParquet_CreatesValidFile verifies that WriteTensorToParquet
// creates a valid parquet file with the correct number of rows.
//
// This test checks:
//   - The function returns without error for valid inputs
//   - The file is created and is a valid parquet file
//   - The row count equals the product of all shape dimensions
func TestWriteTensorToParquet_CreatesValidFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.parquet")

	dimensionNames := []string{"Region", "Product"}
	shape := []int{2, 3}
	data := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}
	conceptValues := map[string][]string{
		"Region":  {"North", "South"},
		"Product": {"A", "B", "C"},
	}

	err := WriteTensorToParquet(filePath, data, dimensionNames, shape, nil, conceptValues)
	require.NoError(t, err, "WriteTensorToParquet should not return an error")

	schema, numRows, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")
	require.NotNil(t, schema, "Schema should not be nil")
	assert.Equal(t, 6, numRows, "File should have 6 rows (2x3)")
}

// TestWriteTensorToParquet_CorrectSchema verifies the parquet schema has the
// expected columns with correct types.
//
// This test checks:
//   - Non-grain dimension columns are present and sorted alphabetically
//   - The "Value" column is present
//   - Total column count equals number of dimensions plus one
//   - Dimension columns are ByteArray (string) type
//   - Value column is Double (float64) type
func TestWriteTensorToParquet_CorrectSchema(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.parquet")

	dimensionNames := []string{"Region", "Product"}
	shape := []int{2, 3}
	data := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}
	conceptValues := map[string][]string{
		"Region":  {"North", "South"},
		"Product": {"A", "B", "C"},
	}

	err := WriteTensorToParquet(filePath, data, dimensionNames, shape, nil, conceptValues)
	require.NoError(t, err, "WriteTensorToParquet should not return an error")

	parquetSchema, _, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	// 2 dimensions + 1 Value = 3 columns
	assert.Equal(t, 3, parquetSchema.NumColumns(), "Schema should have 3 columns")

	columnNames := GetColumnNames(parquetSchema)
	assert.Contains(t, columnNames, "Region", "Schema should have Region column")
	assert.Contains(t, columnNames, "Product", "Schema should have Product column")
	assert.Contains(t, columnNames, "Value", "Schema should have Value column")

	// Dimension columns are ByteArray (string)
	regionType := GetColumnType(parquetSchema, "Region")
	assert.Equal(t, parquet.Types.ByteArray, regionType, "Region column should be ByteArray type")

	productType := GetColumnType(parquetSchema, "Product")
	assert.Equal(t, parquet.Types.ByteArray, productType, "Product column should be ByteArray type")

	// Value column is Double (float64)
	valueType := GetColumnType(parquetSchema, "Value")
	assert.Equal(t, parquet.Types.Double, valueType, "Value column should be Double type")
}

// TestWriteTensorToParquet_CorrectDataValues verifies that the Value column
// contains exactly the input data values in the correct order.
//
// This test checks:
//   - The Value column exists and can be read
//   - The number of values matches len(data)
//   - Each value in the parquet file matches the corresponding input value
func TestWriteTensorToParquet_CorrectDataValues(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.parquet")

	dimensionNames := []string{"Region", "Product"}
	shape := []int{2, 3}
	data := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}
	conceptValues := map[string][]string{
		"Region":  {"North", "South"},
		"Product": {"A", "B", "C"},
	}

	err := WriteTensorToParquet(filePath, data, dimensionNames, shape, nil, conceptValues)
	require.NoError(t, err, "WriteTensorToParquet should not return an error")

	parquetSchema, _, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	columnNames := GetColumnNames(parquetSchema)
	valueIndex := -1
	for i, name := range columnNames {
		if name == "Value" {
			valueIndex = i
			break
		}
	}
	require.NotEqual(t, -1, valueIndex, "Value column should be found")

	valueData, err := ReadParquetColumnData(filePath, valueIndex)
	require.NoError(t, err, "Should be able to read Value column")

	values := valueData.([]float64)
	require.Equal(t, len(data), len(values), "Value column should have same length as input data")
	for i, expected := range data {
		assert.InDelta(t, expected, values[i], 1e-9, "Value at row %d should match input", i)
	}
}

// TestWriteTensorToParquet_CorrectDimensionValues verifies that dimension
// columns contain the correct concept values for each flat tensor index.
//
// This test creates a 2x3 tensor and verifies that for every row, the
// dimension values correspond to the dimension indices derived from the
// flat index using row-major ordering (last dimension varies fastest).
func TestWriteTensorToParquet_CorrectDimensionValues(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.parquet")

	dimensionNames := []string{"Region", "Product"}
	shape := []int{2, 3}
	data := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}
	conceptValues := map[string][]string{
		"Region":  {"North", "South"},
		"Product": {"A", "B", "C"},
	}

	err := WriteTensorToParquet(filePath, data, dimensionNames, shape, nil, conceptValues)
	require.NoError(t, err, "WriteTensorToParquet should not return an error")

	parquetSchema, numRows, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")
	require.Equal(t, 6, numRows, "Should have 6 rows")

	columnNames := GetColumnNames(parquetSchema)

	// Find column indices
	regionIndex := -1
	productIndex := -1
	for i, name := range columnNames {
		switch name {
		case "Region":
			regionIndex = i
		case "Product":
			productIndex = i
		}
	}
	require.NotEqual(t, -1, regionIndex, "Region column should be found")
	require.NotEqual(t, -1, productIndex, "Product column should be found")

	regionData, err := ReadParquetColumnData(filePath, regionIndex)
	require.NoError(t, err, "Should be able to read Region column")
	regions := regionData.([]string)

	productData, err := ReadParquetColumnData(filePath, productIndex)
	require.NoError(t, err, "Should be able to read Product column")
	products := productData.([]string)

	require.Equal(t, 6, len(regions), "Region column should have 6 values")
	require.Equal(t, 6, len(products), "Product column should have 6 values")

	// Verify each row has the correct dimension values for its flat index
	for i := 0; i < 6; i++ {
		indices := flatIndexToDimensionIndices(i, shape)
		expectedRegion := conceptValues["Region"][indices[0]]
		expectedProduct := conceptValues["Product"][indices[1]]

		assert.Equal(t, expectedRegion, regions[i], "Region at row %d should match flat index %d", i, i)
		assert.Equal(t, expectedProduct, products[i], "Product at row %d should match flat index %d", i, i)
	}
}

// TestWriteTensorToParquet_WithGrain verifies that grain columns are included
// in the schema after non-grain dimension columns and before Value, and that
// grain values are written correctly.
//
// This test checks:
//   - Grain columns appear in the schema
//   - Column order is: non-grain dimensions (alphabetical), grain, Value
//   - Grain column type is ByteArray (string)
//   - Grain values match the concept values for each flat index
func TestWriteTensorToParquet_WithGrain(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_grain.parquet")

	dimensionNames := []string{"Region", "Product", "Year"}
	shape := []int{2, 3, 2}
	data := make([]float64, 12)
	for i := range data {
		data[i] = float64(i + 1)
	}
	conceptValues := map[string][]string{
		"Region":  {"North", "South"},
		"Product": {"A", "B", "C"},
		"Year":    {"2023", "2024"},
	}
	grain := []string{"Year"}

	err := WriteTensorToParquet(filePath, data, dimensionNames, shape, grain, conceptValues)
	require.NoError(t, err, "WriteTensorToParquet should not return an error")

	parquetSchema, numRows, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	// 3 dimensions + 1 Value = 4 columns
	assert.Equal(t, 4, parquetSchema.NumColumns(), "Schema should have 4 columns")
	assert.Equal(t, 12, numRows, "Should have 12 rows (2x3x2)")

	columnNames := GetColumnNames(parquetSchema)

	// Expected order: Product, Region (non-grain, alphabetical), Year (grain), Value
	require.Equal(t, 4, len(columnNames), "Schema should have 4 columns")
	assert.Equal(t, "Product", columnNames[0], "First column should be Product (non-grain, alphabetical)")
	assert.Equal(t, "Region", columnNames[1], "Second column should be Region (non-grain, alphabetical)")
	assert.Equal(t, "Year", columnNames[2], "Third column should be Year (grain)")
	assert.Equal(t, "Value", columnNames[3], "Fourth column should be Value")

	// Verify grain column type
	yearType := GetColumnType(parquetSchema, "Year")
	assert.Equal(t, parquet.Types.ByteArray, yearType, "Year column should be ByteArray type")

	// Read Year column and verify values
	yearIndex := 2 // verified above
	yearData, err := ReadParquetColumnData(filePath, yearIndex)
	require.NoError(t, err, "Should be able to read Year column")
	years := yearData.([]string)
	require.Equal(t, 12, len(years), "Year column should have 12 values")

	// In flat-index order for shape [2,3,2] with dims [Region, Product, Year]:
	// Year is the last dimension, so it varies fastest: 0,1,0,1,0,1,...
	expectedYears := []string{
		"2023", "2024", "2023", "2024", "2023", "2024",
		"2023", "2024", "2023", "2024", "2023", "2024",
	}
	for i, expected := range expectedYears {
		assert.Equal(t, expected, years[i], "Year at row %d should match", i)
	}
}

// TestWriteTensorToParquet_ColumnOrder verifies that non-grain dimension
// columns are sorted alphabetically, followed by grain columns in spec order,
// followed by Value.
//
// This test uses dimension names that are NOT alphabetical to confirm that
// sorting is applied correctly.
func TestWriteTensorToParquet_ColumnOrder(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_order.parquet")

	dimensionNames := []string{"Zebra", "Alpha", "Beta"}
	shape := []int{1, 1, 1}
	data := []float64{42.0}
	conceptValues := map[string][]string{
		"Zebra": {"z"},
		"Alpha": {"a"},
		"Beta":  {"b"},
	}

	err := WriteTensorToParquet(filePath, data, dimensionNames, shape, nil, conceptValues)
	require.NoError(t, err, "WriteTensorToParquet should not return an error")

	parquetSchema, _, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	columnNames := GetColumnNames(parquetSchema)
	require.Equal(t, 4, len(columnNames), "Schema should have 4 columns")

	// Non-grain dimensions sorted alphabetically: Alpha, Beta, Zebra, then Value
	assert.Equal(t, "Alpha", columnNames[0], "First column should be Alpha (alphabetical)")
	assert.Equal(t, "Beta", columnNames[1], "Second column should be Beta (alphabetical)")
	assert.Equal(t, "Zebra", columnNames[2], "Third column should be Zebra (alphabetical)")
	assert.Equal(t, "Value", columnNames[3], "Fourth column should be Value")
}

// TestWriteTensorToParquet_DataLengthMismatch verifies that an error is
// returned when the data slice length does not match the product of shape
// dimensions.
func TestWriteTensorToParquet_DataLengthMismatch(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.parquet")

	dimensionNames := []string{"Region", "Product"}
	shape := []int{2, 3}
	// Only 5 values, but shape implies 6
	data := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	conceptValues := map[string][]string{
		"Region":  {"North", "South"},
		"Product": {"A", "B", "C"},
	}

	err := WriteTensorToParquet(filePath, data, dimensionNames, shape, nil, conceptValues)
	require.Error(t, err, "WriteTensorToParquet should return an error when data length mismatches shape")
}

// TestWriteTensorToParquet_MissingConceptValues verifies that an error is
// returned when a dimension name is not found in conceptValues.
func TestWriteTensorToParquet_MissingConceptValues(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.parquet")

	dimensionNames := []string{"Region", "Product"}
	shape := []int{2, 3}
	data := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}
	// Missing "Product" from conceptValues
	conceptValues := map[string][]string{
		"Region": {"North", "South"},
	}

	err := WriteTensorToParquet(filePath, data, dimensionNames, shape, nil, conceptValues)
	require.Error(t, err, "WriteTensorToParquet should return an error when a dimension is missing from conceptValues")
}

// TestWriteTensorToParquet_MissingGrainInDimensions verifies that an error is
// returned when a grain name is not present in dimensionNames.
func TestWriteTensorToParquet_MissingGrainInDimensions(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.parquet")

	dimensionNames := []string{"Region", "Product"}
	shape := []int{2, 3}
	data := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}
	conceptValues := map[string][]string{
		"Region":  {"North", "South"},
		"Product": {"A", "B", "C"},
		"Year":    {"2023", "2024"},
	}
	// Year is in conceptValues but not in dimensionNames
	grain := []string{"Year"}

	err := WriteTensorToParquet(filePath, data, dimensionNames, shape, grain, conceptValues)
	require.Error(t, err, "WriteTensorToParquet should return an error when grain is not in dimensionNames")
}

// TestWriteTensorToParquet_RoundTrip verifies that data written with
// WriteTensorToParquet can be read back correctly with ReadParquetToTensor.
//
// This test checks:
//   - The written parquet file can be read by ReadParquetToTensor
//   - The returned data slice matches the original input
//   - The returned dimension names match the original
//   - The returned shape matches the original
func TestWriteTensorToParquet_RoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_roundtrip.parquet")

	dimensionNames := []string{"Region", "Product"}
	shape := []int{2, 3}
	data := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}
	conceptValues := map[string][]string{
		"Region":  {"North", "South"},
		"Product": {"A", "B", "C"},
	}

	err := WriteTensorToParquet(filePath, data, dimensionNames, shape, nil, conceptValues)
	require.NoError(t, err, "WriteTensorToParquet should not return an error")

	// Read back with ReadParquetToTensor
	readData, readDimNames, readShape, err := ReadParquetToTensor(filePath, dimensionNames, nil, conceptValues)
	require.NoError(t, err, "ReadParquetToTensor should not return an error")

	// Verify dimension names match
	assert.Equal(t, dimensionNames, readDimNames, "Dimension names should match")

	// Verify shape matches
	assert.Equal(t, shape, readShape, "Shape should match")

	// Verify data matches
	require.Equal(t, len(data), len(readData), "Data length should match")
	for i, expected := range data {
		assert.InDelta(t, expected, readData[i], 1e-9, "Data at index %d should match", i)
	}
}
