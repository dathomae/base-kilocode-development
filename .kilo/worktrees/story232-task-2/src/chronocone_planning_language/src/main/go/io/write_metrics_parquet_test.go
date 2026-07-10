// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package io

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/arrow/go/v16/parquet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteMetricsToParquet_SingleMetric verifies that WriteMetricsToParquet
// correctly writes a single metric tensor to a parquet file with the Metric
// column, dimension columns, grain columns, and Value column.
//
// This test checks:
//   - The file is created successfully
//   - Schema has Metric (ByteArray) + dimension columns + grain columns + Value (Float64)
//   - Correct row count (product of shape dimensions)
//   - Metric column populated with metric name for all rows
//   - Values match input data
func TestWriteMetricsToParquet_SingleMetric(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_single.parquet")

	conceptValues := map[string][]string{
		"Region":  {"North", "South"},
		"Product": {"A", "B"},
		"Year":    {"2023", "2024"},
	}

	metrics := []MetricTensorData{
		{
			MetricName:     "total_sales",
			Data:           []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0},
			DimensionNames: []string{"Region", "Product", "Year"},
			Shape:          []int{2, 2, 2},
			Grain:          []string{"Year"},
		},
	}

	err := WriteMetricsToParquet(filePath, metrics, conceptValues)
	require.NoError(t, err, "WriteMetricsToParquet should succeed")

	// Verify file exists
	_, err = os.Stat(filePath)
	require.NoError(t, err, "File should exist")

	// Read the parquet file
	parquetSchema, numRows, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	// Verify schema has expected columns
	columnNames := GetColumnNames(parquetSchema)
	t.Logf("Column names: %v", columnNames)

	// Schema should have: Metric, Region, Product, Year, Value = 5 columns
	assert.Equal(t, 5, parquetSchema.NumColumns(), "Schema should have 5 columns")

	// Verify Metric column is ByteArray
	metricType := GetColumnType(parquetSchema, "Metric")
	assert.Equal(t, parquet.Types.ByteArray, metricType, "Metric column should be ByteArray")

	// Verify dimension columns are ByteArray
	regionType := GetColumnType(parquetSchema, "Region")
	assert.Equal(t, parquet.Types.ByteArray, regionType, "Region column should be ByteArray")

	productType := GetColumnType(parquetSchema, "Product")
	assert.Equal(t, parquet.Types.ByteArray, productType, "Product column should be ByteArray")

	// Verify grain column is ByteArray
	yearType := GetColumnType(parquetSchema, "Year")
	assert.Equal(t, parquet.Types.ByteArray, yearType, "Year column should be ByteArray")

	// Verify Value column is Float64
	valueType := GetColumnType(parquetSchema, "Value")
	assert.Equal(t, parquet.Types.Double, valueType, "Value column should be Double")

	// Expected rows: 2 (Region) × 2 (Product) × 2 (Year) = 8
	expectedRows := 8
	assert.Equal(t, expectedRows, numRows, "Should have %d rows", expectedRows)

	// Read Metric column and verify all rows have the metric name
	metricIdx := indexOf(columnNames, "Metric")
	require.NotEqual(t, -1, metricIdx, "Metric column should be found")
	metricData, err := ReadParquetColumnData(filePath, metricIdx)
	require.NoError(t, err, "Should be able to read Metric column")
	metricNameVals := metricData.([]string)
	for i, m := range metricNameVals {
		assert.Equal(t, "total_sales", m, "Metric column row %d should be 'total_sales'", i)
	}

	// Read Value column
	valueIdx := indexOf(columnNames, "Value")
	require.NotEqual(t, -1, valueIdx, "Value column should be found")
	valueData, err := ReadParquetColumnData(filePath, valueIdx)
	require.NoError(t, err, "Should be able to read Value column")
	values := valueData.([]float64)
	require.Equal(t, expectedRows, len(values), "Value column should have %d rows", expectedRows)

	// The values should be in row-major order: Region (slowest), Product, Year (fastest)
	// shape [2,2,2]: Region=[North,South], Product=[A,B], Year=[2023,2024]
	// row 0: North, A, 2023 -> 1.0; row 1: North, A, 2024 -> 2.0
	// row 2: North, B, 2023 -> 3.0; row 3: North, B, 2024 -> 4.0
	// row 4: South, A, 2023 -> 5.0; row 5: South, A, 2024 -> 6.0
	// row 6: South, B, 2023 -> 7.0; row 7: South, B, 2024 -> 8.0
	expectedValues := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0}
	for i, v := range values {
		assert.Equal(t, expectedValues[i], v, "Value[%d] should be %f", i, expectedValues[i])
	}

	// Spot check row 7: Region=South, Product=B, Year=2024 should be value 8.0
	regionIdx := indexOf(columnNames, "Region")
	require.NotEqual(t, -1, regionIdx)
	regionData, err := ReadParquetColumnData(filePath, regionIdx)
	require.NoError(t, err)
	regions := regionData.([]string)

	productIdx := indexOf(columnNames, "Product")
	require.NotEqual(t, -1, productIdx)
	productData, err := ReadParquetColumnData(filePath, productIdx)
	require.NoError(t, err)
	products := productData.([]string)

	yearIdx := indexOf(columnNames, "Year")
	require.NotEqual(t, -1, yearIdx)
	yearData, err := ReadParquetColumnData(filePath, yearIdx)
	require.NoError(t, err)
	years := yearData.([]string)

	assert.Equal(t, "South", regions[7])
	assert.Equal(t, "B", products[7])
	assert.Equal(t, "2024", years[7])
	assert.Equal(t, 8.0, values[7])

	// Verify all expected Cartesian product combinations exist
	combinations := make(map[string]bool)
	for i := 0; i < len(regions); i++ {
		key := regions[i] + "-" + products[i] + "-" + years[i]
		combinations[key] = true
	}

	expectedCombos := []string{
		"North-A-2023", "North-A-2024",
		"North-B-2023", "North-B-2024",
		"South-A-2023", "South-A-2024",
		"South-B-2023", "South-B-2024",
	}
	for _, combo := range expectedCombos {
		assert.True(t, combinations[combo], "Missing combination: %s", combo)
	}
}

// TestWriteMetricsToParquet_MultipleMetrics verifies that WriteMetricsToParquet
// correctly writes multiple metric tensors to a single parquet file.
//
// This test checks:
//   - All metrics' rows are present in the single file
//   - The Metric column correctly identifies each metric's rows
//   - Rows are interleaved correctly (all rows for metric1, then metric2)
func TestWriteMetricsToParquet_MultipleMetrics(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_multi.parquet")

	conceptValues := map[string][]string{
		"Region": {"North", "South"},
		"Year":   {"2023", "2024"},
	}

	metrics := []MetricTensorData{
		{
			MetricName:     "revenue",
			Data:           []float64{10.0, 20.0, 30.0, 40.0},
			DimensionNames: []string{"Region", "Year"},
			Shape:          []int{2, 2},
			Grain:          []string{"Year"},
		},
		{
			MetricName:     "costs",
			Data:           []float64{100.0, 200.0, 300.0, 400.0},
			DimensionNames: []string{"Region", "Year"},
			Shape:          []int{2, 2},
			Grain:          []string{"Year"},
		},
	}

	err := WriteMetricsToParquet(filePath, metrics, conceptValues)
	require.NoError(t, err, "WriteMetricsToParquet should succeed")

	// Read the parquet file
	parquetSchema, numRows, err := ReadParquetFile(filePath)
	require.NoError(t, err)

	// Expected rows: 4 rows per metric × 2 metrics = 8
	assert.Equal(t, 8, numRows, "Should have 8 rows (4 per metric × 2 metrics)")

	columnNames := GetColumnNames(parquetSchema)
	t.Logf("Column names: %v", columnNames)

	// Read Metric column
	metricIdx := indexOf(columnNames, "Metric")
	require.NotEqual(t, -1, metricIdx)
	metricData, err := ReadParquetColumnData(filePath, metricIdx)
	require.NoError(t, err)
	metricNames := metricData.([]string)

	// First 4 rows should be "revenue", last 4 rows should be "costs"
	for i := 0; i < 4; i++ {
		assert.Equal(t, "revenue", metricNames[i], "Row %d should be revenue", i)
	}
	for i := 4; i < 8; i++ {
		assert.Equal(t, "costs", metricNames[i], "Row %d should be costs", i)
	}

	// Read Value column
	valueIdx := indexOf(columnNames, "Value")
	require.NotEqual(t, -1, valueIdx)
	valueData, err := ReadParquetColumnData(filePath, valueIdx)
	require.NoError(t, err)
	values := valueData.([]float64)

	// First metric (revenue) values: 10.0, 20.0, 30.0, 40.0
	assert.Equal(t, []float64{10.0, 20.0, 30.0, 40.0}, values[0:4])
	// Second metric (costs) values: 100.0, 200.0, 300.0, 400.0
	assert.Equal(t, []float64{100.0, 200.0, 300.0, 400.0}, values[4:8])
}

// TestWriteMetricsToParquet_MismatchedDimensions verifies that
// WriteMetricsToParquet returns an error when metrics have different
// DimensionNames.
func TestWriteMetricsToParquet_MismatchedDimensions(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_mismatch_dims.parquet")

	conceptValues := map[string][]string{
		"Region":  {"North", "South"},
		"Product": {"A", "B"},
	}

	metrics := []MetricTensorData{
		{
			MetricName:     "metric1",
			Data:           []float64{1.0, 2.0},
			DimensionNames: []string{"Region"},
			Shape:          []int{2},
			Grain:          nil,
		},
		{
			MetricName:     "metric2",
			Data:           []float64{1.0, 2.0, 3.0, 4.0},
			DimensionNames: []string{"Region", "Product"},
			Shape:          []int{2, 2},
			Grain:          nil,
		},
	}

	err := WriteMetricsToParquet(filePath, metrics, conceptValues)
	require.Error(t, err, "WriteMetricsToParquet should return error for mismatched dimensions")
	assert.Contains(t, err.Error(), "DimensionNames")
}

// TestWriteMetricsToParquet_MismatchedShapes verifies that
// WriteMetricsToParquet returns an error when metrics have different Shapes.
func TestWriteMetricsToParquet_MismatchedShapes(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_mismatch_shape.parquet")

	conceptValues := map[string][]string{
		"Region": {"North", "South", "East", "West"},
	}

	metrics := []MetricTensorData{
		{
			MetricName:     "metric1",
			Data:           []float64{1.0, 2.0},
			DimensionNames: []string{"Region"},
			Shape:          []int{2},
			Grain:          nil,
		},
		{
			MetricName:     "metric2",
			Data:           []float64{1.0, 2.0, 3.0, 4.0},
			DimensionNames: []string{"Region"},
			Shape:          []int{4},
			Grain:          nil,
		},
	}

	err := WriteMetricsToParquet(filePath, metrics, conceptValues)
	require.Error(t, err, "WriteMetricsToParquet should return error for mismatched shapes")
	assert.Contains(t, err.Error(), "Shape")
}

// TestWriteMetricsToParquet_EmptyMetrics verifies that WriteMetricsToParquet
// returns an error when given an empty metrics slice.
func TestWriteMetricsToParquet_EmptyMetrics(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_empty.parquet")

	err := WriteMetricsToParquet(filePath, nil, nil)
	require.Error(t, err, "WriteMetricsToParquet should return error for nil metrics")
	assert.Contains(t, err.Error(), "no metrics")

	err = WriteMetricsToParquet(filePath, []MetricTensorData{}, nil)
	require.Error(t, err, "WriteMetricsToParquet should return error for empty metrics")
	assert.Contains(t, err.Error(), "no metrics")
}

// TestWriteMetricsToParquet_Overwrite verifies that a second call to
// WriteMetricsToParquet overwrites the existing file.
func TestWriteMetricsToParquet_Overwrite(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_overwrite.parquet")

	conceptValues := map[string][]string{
		"Region": {"North", "South"},
	}

	// First write: metric1 with 2 rows
	metrics1 := []MetricTensorData{
		{
			MetricName:     "metric1",
			Data:           []float64{1.0, 2.0},
			DimensionNames: []string{"Region"},
			Shape:          []int{2},
			Grain:          nil,
		},
	}
	err := WriteMetricsToParquet(filePath, metrics1, conceptValues)
	require.NoError(t, err, "First WriteMetricsToParquet should succeed")

	// Verify first write: 2 rows, metric1
	_, numRows1, err := ReadParquetFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, 2, numRows1, "First write should have 2 rows")

	// Second write: metric2 with 2 rows (overwrites)
	metrics2 := []MetricTensorData{
		{
			MetricName:     "metric2",
			Data:           []float64{10.0, 20.0},
			DimensionNames: []string{"Region"},
			Shape:          []int{2},
			Grain:          nil,
		},
	}
	err = WriteMetricsToParquet(filePath, metrics2, conceptValues)
	require.NoError(t, err, "Second WriteMetricsToParquet should succeed (overwrite)")

	// Verify second write: 2 rows, metric2
	parquetSchema, numRows2, err := ReadParquetFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, 2, numRows2, "Second write should have 2 rows")

	columnNames := GetColumnNames(parquetSchema)
	metricIdx := indexOf(columnNames, "Metric")
	require.NotEqual(t, -1, metricIdx)
	metricData, err := ReadParquetColumnData(filePath, metricIdx)
	require.NoError(t, err)
	metricNames := metricData.([]string)
	assert.Equal(t, "metric2", metricNames[0], "Metric column should be 'metric2' after overwrite")
}

// indexOf returns the index of the target string in the slice, or -1 if not found.
func indexOf(slice []string, target string) int {
	for i, v := range slice {
		if v == target {
			return i
		}
	}
	return -1
}
