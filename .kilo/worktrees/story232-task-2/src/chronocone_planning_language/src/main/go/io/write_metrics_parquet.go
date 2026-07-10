package io

import (
	"fmt"
	"os"
	"sort"

	"github.com/apache/arrow/go/v16/parquet"
	"github.com/apache/arrow/go/v16/parquet/compress"
	"github.com/apache/arrow/go/v16/parquet/file"
	"github.com/apache/arrow/go/v16/parquet/schema"
)

// MetricTensorData bundles a single output metric's tensor data for writing to parquet.
// All metrics written together via WriteMetricsToParquet must share the same dimension
// and grain structure (same DimensionNames, same Shape, same Grain values).
//
// Fields:
//   - MetricName: the name of the metric (written to the "Metric" column)
//   - Data: flat slice of float64 cell values in row-major order
//   - DimensionNames: ordered list of dimension column names
//   - Shape: size of each dimension (product equals len(Data))
//   - Grain: list of grain measure value strings (same for all rows of this metric)
type MetricTensorData struct {
	MetricName     string
	Data           []float64
	DimensionNames []string
	Shape          []int
	Grain          []string
}

// WriteMetricsToParquet writes multiple metric tensors to a single parquet file.
//
// The output parquet schema is:
//   - "Metric": ByteArray (UTF-8 string) — the metric name for each row
//   - One ByteArray column per non-grain dimension (alphabetically sorted)
//   - One ByteArray column per grain measure (in Grain order)
//   - "Value": Float64 — the cell value
//
// All metrics in the slice must share identical dimension and grain structure
// (same DimensionNames, same Shape, same Grain). If the metrics have different
// structures, an error is returned.
//
// The conceptValues parameter maps each dimension name to its ordered list of
// concept values; these are used to populate the dimension columns in Cartesian
// product order.
//
// Parameters:
//   - filePath: path to the output parquet file (will be overwritten if exists)
//   - metrics: slice of metric tensor data to write
//   - conceptValues: map from dimension name to ordered list of concept values
//
// Returns nil on success, or an error if validation fails, file creation fails,
// or parquet writing fails.
//
// WriteTensorToParquet is retained for backward compatibility with callers
// that write a single tensor without a Metric column.
func WriteMetricsToParquet(filePath string, metrics []MetricTensorData, conceptValues map[string][]string) (err error) {
	if len(metrics) == 0 {
		return fmt.Errorf("no metrics provided")
	}

	// Validate all metrics share identical dimension and grain structure.
	ref := metrics[0]
	for i := 1; i < len(metrics); i++ {
		m := metrics[i]
		if !stringSlicesEqual(ref.DimensionNames, m.DimensionNames) {
			return fmt.Errorf("metrics[%d] has different DimensionNames: %v != %v", i, m.DimensionNames, ref.DimensionNames)
		}
		if !intSlicesEqual(ref.Shape, m.Shape) {
			return fmt.Errorf("metrics[%d] has different Shape: %v != %v", i, m.Shape, ref.Shape)
		}
		if !stringSlicesEqual(ref.Grain, m.Grain) {
			return fmt.Errorf("metrics[%d] has different Grain: %v != %v", i, m.Grain, ref.Grain)
		}
	}

	dimensionNames := ref.DimensionNames
	shape := ref.Shape
	grain := ref.Grain

	// Validate each metric's data length matches the product of shape dimensions.
	expectedSize := 1
	for _, s := range shape {
		expectedSize *= s
	}
	for i, m := range metrics {
		if len(m.Data) != expectedSize {
			return fmt.Errorf("metrics[%d] data length %d does not match product of shape dimensions %d", i, len(m.Data), expectedSize)
		}
	}

	// Validate every name in dimensionNames exists as a key in conceptValues
	// and that the number of concept values matches the shape.
	for i, dim := range dimensionNames {
		values, ok := conceptValues[dim]
		if !ok {
			return fmt.Errorf("dimension %q not found in conceptValues", dim)
		}
		if len(values) != shape[i] {
			return fmt.Errorf("dimension %q has %d concept values but shape expects %d", dim, len(values), shape[i])
		}
	}

	// Validate every name in grain exists in dimensionNames.
	for _, g := range grain {
		found := false
		for _, dim := range dimensionNames {
			if dim == g {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("grain %q not found in dimensionNames", g)
		}
	}

	// Build a set of grain names for quick lookup.
	grainSet := make(map[string]struct{}, len(grain))
	for _, g := range grain {
		grainSet[g] = struct{}{}
	}

	// Separate non-grain dimensions and sort alphabetically.
	nonGrain := make([]string, 0, len(dimensionNames))
	for _, dim := range dimensionNames {
		if _, ok := grainSet[dim]; !ok {
			nonGrain = append(nonGrain, dim)
		}
	}
	sort.Strings(nonGrain)

	// Final column order: Metric, non-grain (alphabetical), grain (spec order), Value.
	columnNames := make([]string, 0, 1+len(dimensionNames)+1)
	columnNames = append(columnNames, "Metric")
	columnNames = append(columnNames, nonGrain...)
	columnNames = append(columnNames, grain...)
	columnNames = append(columnNames, "Value")

	// Map dimension name to its position in dimensionNames/shape for index lookup.
	dimNameToIndex := make(map[string]int, len(dimensionNames))
	for i, dim := range dimensionNames {
		dimNameToIndex[dim] = i
	}

	// Build the Parquet schema nodes.
	nodeList := make(schema.FieldList, 0, len(columnNames))
	for _, name := range columnNames {
		if name == "Value" {
			nodeList = append(nodeList, schema.NewFloat64Node("Value", parquet.Repetitions.Required, -1))
		} else {
			nodeList = append(nodeList, schema.NewByteArrayNode(name, parquet.Repetitions.Required, -1))
		}
	}

	parquetSchema, err := schema.NewGroupNode("schema", parquet.Repetitions.Required, nodeList, -1)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Create the output file.
	outputFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outputFile.Close()

	// Create writer properties with Snappy compression.
	props := parquet.NewWriterProperties(parquet.WithCompression(compress.Codecs.Snappy))

	// Create the parquet writer.
	writer := file.NewParquetWriter(outputFile, parquetSchema, file.WithWriterProps(props))
	defer func() {
		if closeErr := writer.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close parquet writer: %w", closeErr)
		}
	}()

	// Compute total number of rows: sum of data lengths across all metrics.
	totalRows := 0
	for _, m := range metrics {
		totalRows += len(m.Data)
	}

	// Create a row group writer.
	rowGroupWriter := writer.AppendRowGroup()

	// Write each column in schema order.
	for _, colName := range columnNames {
		columnWriter, colErr := rowGroupWriter.NextColumn()
		if colErr != nil {
			return fmt.Errorf("failed to get column writer: %w", colErr)
		}

		if colName == "Value" {
			floatWriter, ok := columnWriter.(*file.Float64ColumnChunkWriter)
			if !ok {
				return fmt.Errorf("expected Float64ColumnChunkWriter for Value column, got %T", columnWriter)
			}

			// Collect all values across all metrics.
			allValues := make([]float64, 0, totalRows)
			for _, m := range metrics {
				allValues = append(allValues, m.Data...)
			}
			_, err = floatWriter.WriteBatch(allValues, nil, nil)
			if err != nil {
				return fmt.Errorf("failed to write value column: %w", err)
			}
		} else if colName == "Metric" {
			stringWriter, ok := columnWriter.(*file.ByteArrayColumnChunkWriter)
			if !ok {
				return fmt.Errorf("expected ByteArrayColumnChunkWriter for Metric column, got %T", columnWriter)
			}

			metricColVals := make([]parquet.ByteArray, 0, totalRows)
			for _, m := range metrics {
				for range m.Data {
					metricColVals = append(metricColVals, parquet.ByteArray(m.MetricName))
				}
			}
			_, err = stringWriter.WriteBatch(metricColVals, nil, nil)
			if err != nil {
				return fmt.Errorf("failed to write Metric column: %w", err)
			}
		} else {
			stringWriter, ok := columnWriter.(*file.ByteArrayColumnChunkWriter)
			if !ok {
				return fmt.Errorf("expected ByteArrayColumnChunkWriter for %q column, got %T", colName, columnWriter)
			}

			colVals := make([]parquet.ByteArray, 0, totalRows)
			dimPos := dimNameToIndex[colName]

			// For each metric, iterate over its data in row-major order.
			for _, m := range metrics {
				indices := make([]int, len(m.Shape))
				for i := 0; i < len(m.Data); i++ {
					// Compute multi-dimensional indices for flat index i
					// using row-major ordering (last dimension varies fastest).
					remainder := i
					for d := len(m.Shape) - 1; d >= 0; d-- {
						indices[d] = remainder % m.Shape[d]
						remainder /= m.Shape[d]
					}
					dimIdx := indices[dimPos]
					colVals = append(colVals, parquet.ByteArray(conceptValues[colName][dimIdx]))
				}
			}
			_, err = stringWriter.WriteBatch(colVals, nil, nil)
			if err != nil {
				return fmt.Errorf("failed to write string column %s: %w", colName, err)
			}
		}
	}

	// Close the row group writer.
	err = rowGroupWriter.Close()
	if err != nil {
		return fmt.Errorf("failed to close row group writer: %w", err)
	}

	return nil
}

// stringSlicesEqual returns true if a and b contain the same elements in the same order.
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// intSlicesEqual returns true if a and b contain the same elements in the same order.
func intSlicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
