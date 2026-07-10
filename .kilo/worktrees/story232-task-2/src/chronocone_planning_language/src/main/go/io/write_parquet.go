// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

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

// WriteTensorToParquet writes an existing tensor's dense data to a Parquet file.
//
// This function takes raw tensor data and metadata and persists it as a Parquet
// file where each row represents one combination of dimension and grain values
// paired with the corresponding tensor value.
//
// Parquet Schema:
//   - Dimension columns (excluding grain): string type (ByteArray), sorted
//     alphabetically by dimension name.
//   - Grain columns: string type (ByteArray), appearing in the exact order
//     specified by the grain parameter.
//   - Value column: float64 type (Double), containing the tensor value for
//     the row's dimension/grain combination.
//
// The total number of rows written equals len(data), which must equal the
// Cartesian product of all shape elements (the total number of dimension
// value combinations).
//
// Design decisions:
//   - dimensionNames contains all tensor dimension names, which may include
//     grain names. The order of dimensionNames corresponds to the order of
//     the shape slice.
//   - grain is a slice of grain measure names that must also appear in
//     dimensionNames. These names determine which columns are treated as
//     grain for ordering purposes in the output Parquet file.
//   - conceptValues maps each dimension name to its ordered list of string
//     values. These values are used to convert flat tensor indices back into
//     human-readable dimension value strings for each row.
//   - Column order in the output Parquet file is: non-grain dimension columns
//     sorted alphabetically, followed by grain columns in the order specified
//     by the grain parameter, followed by the Value column.
//
// Parameters:
//   - filePath: The path where the Parquet file will be created. If the file
//     already exists, it will be overwritten.
//   - data: A dense []float64 slice containing all tensor values in row-major
//     order. The length must equal the product of all elements in shape.
//   - dimensionNames: The ordered list of all dimension names in the tensor.
//   - shape: The size of each dimension, in the same order as dimensionNames.
//   - grain: The ordered list of grain measure names. Each name must also
//     appear in dimensionNames. When empty, no grain columns are emitted.
//   - conceptValues: A map from dimension name to its ordered list of possible
//     string values. Used to map tensor indices back to dimension values.
//
// Returns:
//   - error: An error if the file cannot be created, the schema cannot be
//     built, writing fails, or if len(data) does not match the product of
//     shape elements. Returns nil on successful completion.
func WriteTensorToParquet(filePath string, data []float64, dimensionNames []string, shape []int, grain []string, conceptValues map[string][]string) (err error) {
	// Validate data length matches product of shape dimensions.
	expectedSize := 1
	for _, s := range shape {
		expectedSize *= s
	}
	if len(data) != expectedSize {
		return fmt.Errorf("data length %d does not match product of shape dimensions %d", len(data), expectedSize)
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

	// Final column order: non-grain (alphabetical), grain (spec order), Value.
	columnNames := make([]string, 0, len(dimensionNames)+1)
	columnNames = append(columnNames, nonGrain...)
	columnNames = append(columnNames, grain...)
	columnNames = append(columnNames, "Value")

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

	// Create writer properties with Snappy compression for balanced
	// performance and file size.
	props := parquet.NewWriterProperties(parquet.WithCompression(compress.Codecs.Snappy))

	// Create the parquet writer.
	writer := file.NewParquetWriter(outputFile, parquetSchema, file.WithWriterProps(props))
	defer func() {
		if closeErr := writer.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close parquet writer: %w", closeErr)
		}
	}()

	// Create a row group writer (write all rows in one row group).
	rowGroupWriter := writer.AppendRowGroup()

	// Map dimension name to its position in dimensionNames/shape for index lookup.
	dimNameToIndex := make(map[string]int, len(dimensionNames))
	for i, dim := range dimensionNames {
		dimNameToIndex[dim] = i
	}

	// Write each column in schema order.
	for _, colName := range columnNames {
		columnWriter, err := rowGroupWriter.NextColumn()
		if err != nil {
			return fmt.Errorf("failed to get column writer: %w", err)
		}

		if colName == "Value" {
			floatWriter, ok := columnWriter.(*file.Float64ColumnChunkWriter)
			if !ok {
				return fmt.Errorf("expected Float64ColumnChunkWriter for Value column, got %T", columnWriter)
			}
			_, err = floatWriter.WriteBatch(data, nil, nil)
			if err != nil {
				return fmt.Errorf("failed to write value column: %w", err)
			}
		} else {
			stringWriter, ok := columnWriter.(*file.ByteArrayColumnChunkWriter)
			if !ok {
				return fmt.Errorf("expected ByteArrayColumnChunkWriter for %q column, got %T", colName, columnWriter)
			}
			values := make([]parquet.ByteArray, len(data))
			dimPos := dimNameToIndex[colName]
			indices := make([]int, len(shape))

			for i := 0; i < len(data); i++ {
				// Compute multi-dimensional indices for flat index i
				// using row-major ordering (last dimension varies fastest).
				remainder := i
				for d := len(shape) - 1; d >= 0; d-- {
					indices[d] = remainder % shape[d]
					remainder /= shape[d]
				}
				dimIdx := indices[dimPos]
				values[i] = parquet.ByteArray(conceptValues[colName][dimIdx])
			}

			_, err = stringWriter.WriteBatch(values, nil, nil)
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
