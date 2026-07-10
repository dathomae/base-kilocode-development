// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package io

import (
	"fmt"
	"os"
	"strings"

	"github.com/apache/arrow/go/v16/parquet"
	"github.com/apache/arrow/go/v16/parquet/file"
)

// ReadParquetToTensor reads a parquet file and returns raw tensor data suitable
// for constructing a ChronoconeTensor.
//
// This function reads a parquet file where each row represents a single data cell
// with dimension columns (string/ByteArray type) and a "Value" column (float64/Double
// type). It maps the parquet rows into a dense row-major tensor by computing the
// flat index for each row based on the dimension values and their positions within
// the provided conceptValues.
//
// The parquet file columns are sorted alphabetically (e.g., "Region", "Year", "Value"),
// but the connector may declare dimensions in a different order (e.g., ["Year", "Region"]).
// This function correctly maps columns regardless of ordering by looking up each
// dimension column by name.
//
// The returned data slice is a dense row-major array where the last dimension varies
// fastest. The shape and dimensionNames follow the order specified by the dimensions
// argument, not the parquet column order.
//
// Parameters:
//   - filePath: The path to the parquet file to read.
//   - dimensions: The ordered list of dimension names defining the tensor shape.
//     This order determines the layout of the returned data and shape.
//   - columnToConcept: A map from parquet column name (dimension name) to the
//     concept name used for conceptValues lookup. When nil or empty, column
//     names map to themselves (identity mapping), preserving backward
//     compatibility. For grain measures, the map must contain an entry mapping
//     the measure column name to its underlying concept name.
//   - conceptValues: A map from concept name to the ordered list of possible
//     values for that dimension. The position of each value within its slice
//     determines its index in the tensor. This map must contain an entry for
//     every resolved concept name corresponding to the dimensions argument.
//
// Returns:
//   - data: A dense []float64 slice containing the tensor values in row-major order.
//     The length equals the product of all shape dimensions. Missing combinations
//     are permitted and zero-initialized; duplicate or extraneous rows produce
//     an error.
//   - dimensionNames: The ordered list of dimension names (a copy of the dimensions
//     argument), matching the shape entries.
//   - shape: The size of each dimension, in the same order as dimensionNames.
//     Each entry is len(conceptValues[conceptName]) for the corresponding
//     dimension, where conceptName is resolved via columnToConcept.
//   - err: An error if the file cannot be opened, the schema is invalid, required
//     columns are missing, or any dimension value in the parquet file is not found
//     in conceptValues.
//
// Algorithm:
//  1. Open the parquet file and read its schema.
//  2. Verify that each dimension in the dimensions argument has a corresponding
//     column in the parquet file, and that a "Value" column exists.
//  3. Compute the tensor shape from len(conceptValues[conceptName]) for each
//     dimension, resolving conceptName via columnToConcept.
//  4. Allocate a dense data slice and a `seen` tracker of size product(shape).
//  5. For each row in the parquet file:
//     a. Read each dimension value.
//     b. Look up the dimension value's position within conceptValues[conceptName].
//     c. Compute the flat row-major index using the dimension indices.
//     d. Read the Value and store it at the computed index.
//  6. Return the data, dimensionNames, and shape.
//
// Important: This function does NOT import the calculation package. The caller
// (typically generated package main code) is responsible for constructing the
// ChronoconeTensor from the returned raw data, dimensionNames, and shape.
// This avoids an import cycle since calculation already imports io.
//
// Example:
//
//	conceptValues := map[string][]string{
//	    "Year":   {"2023", "2024"},
//	    "Region": {"North", "South", "East", "West"},
//	}
//	dims := []string{"Year", "Region"}
//	data, dimNames, shape, err := ReadParquetToTensor("data.parquet", dims, nil, conceptValues)
//	// shape == []int{2, 4}
//	// data has 8 elements in row-major order (Year varies slowest, Region fastest)
func ReadParquetToTensor(filePath string, dimensions []string, columnToConcept map[string]string, conceptValues map[string][]string) (data []float64, dimensionNames []string, shape []int, err error) {
	// Validate that every requested dimension has an entry in conceptValues.
	for _, dim := range dimensions {
		conceptName := dim
		if columnToConcept != nil && columnToConcept[dim] != "" {
			conceptName = columnToConcept[dim]
		}
		if _, ok := conceptValues[conceptName]; !ok {
			return nil, nil, nil, fmt.Errorf("dimension %q not found in conceptValues", dim)
		}
	}

	// Open the parquet file.
	inputFile, err := os.Open(filePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open parquet file: %w", err)
	}
	defer inputFile.Close()

	// Create the parquet reader.
	reader, err := file.NewParquetReader(inputFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create parquet reader: %w", err)
	}
	defer reader.Close()

	// Map column names to their indices in the parquet schema.
	schema := reader.MetaData().Schema
	columnIndexByName := make(map[string]int, schema.NumColumns())
	for i := 0; i < schema.NumColumns(); i++ {
		columnIndexByName[schema.Column(i).Name()] = i
	}

	// Validate that the parquet file has exactly the expected columns.
	if schema.NumColumns() != len(dimensions)+1 {
		var unexpected []string
		for name := range columnIndexByName {
			if name == "Value" {
				continue
			}
			found := false
			for _, dim := range dimensions {
				if dim == name {
					found = true
					break
				}
			}
			if !found {
				unexpected = append(unexpected, name)
			}
		}
		return nil, nil, nil, fmt.Errorf("parquet file has %d columns but expected %d (dimensions + Value); unexpected columns: %v", schema.NumColumns(), len(dimensions)+1, unexpected)
	}

	// Verify that a "Value" column exists.
	valueColIdx, ok := columnIndexByName["Value"]
	if !ok {
		return nil, nil, nil, fmt.Errorf("parquet file missing required 'Value' column")
	}

	// Verify that each dimension has a corresponding column and build shape.
	shape = make([]int, len(dimensions))
	dimColumnIndices := make([]int, len(dimensions))
	for i, dim := range dimensions {
		idx, ok := columnIndexByName[dim]
		if !ok {
			return nil, nil, nil, fmt.Errorf("parquet file missing dimension column %q", dim)
		}
		dimColumnIndices[i] = idx
		conceptName := dim
		if columnToConcept != nil && columnToConcept[dim] != "" {
			conceptName = columnToConcept[dim]
		}
		shape[i] = len(conceptValues[conceptName])
	}

	// Compute total number of elements.
	totalSize := 1
	for _, s := range shape {
		totalSize *= s
	}

	// Allocate the dense data slice and a seen tracker.
	data = make([]float64, totalSize)
	seen := make([]bool, totalSize)

	// Build lookup maps: dimension -> value -> index.
	// This avoids repeated linear scans of conceptValues during row processing.
	valueIndices := make([]map[string]int, len(dimensions))
	for i, dim := range dimensions {
		conceptName := dim
		if columnToConcept != nil && columnToConcept[dim] != "" {
			conceptName = columnToConcept[dim]
		}
		values := conceptValues[conceptName]
		m := make(map[string]int, len(values))
		for j, v := range values {
			m[v] = j
		}
		valueIndices[i] = m
	}

	// Process each row group in the parquet file.
	totalRowsRead := 0
	var duplicateFound bool
	var duplicateIndex int
	for rgIdx := 0; rgIdx < reader.NumRowGroups(); rgIdx++ {
		rowGroupReader := reader.RowGroup(rgIdx)
		numRows := int(rowGroupReader.NumRows())

		// Read each dimension column into memory as []string.
		dimColumnData := make([][]string, len(dimensions))
		for i, colIdx := range dimColumnIndices {
			columnReader, err := rowGroupReader.Column(colIdx)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to get column reader for %q: %w", dimensions[i], err)
			}

			stringReader := columnReader.(*file.ByteArrayColumnChunkReader)
			values := make([]parquet.ByteArray, numRows)
			totalRead := 0
			for totalRead < numRows {
				batchRead, _, err := stringReader.ReadBatch(int64(numRows-totalRead), values[totalRead:], nil, nil)
				if err != nil {
					return nil, nil, nil, fmt.Errorf("failed to read string column %q: %w", dimensions[i], err)
				}
				totalRead += int(batchRead)
			}

			// Convert parquet.ByteArray to string.
			strValues := make([]string, numRows)
			for j, v := range values {
				strValues[j] = string(v)
			}
			dimColumnData[i] = strValues
		}

		// Read the Value column into memory as []float64.
		valueColumnReader, err := rowGroupReader.Column(valueColIdx)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to get value column reader: %w", err)
		}

		floatReader := valueColumnReader.(*file.Float64ColumnChunkReader)
		floatValues := make([]float64, numRows)
		totalRead := 0
		for totalRead < numRows {
			batchRead, _, err := floatReader.ReadBatch(int64(numRows-totalRead), floatValues[totalRead:], nil, nil)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to read value column: %w", err)
			}
			totalRead += int(batchRead)
		}

		// Process each row: look up dimension indices and compute flat row-major index.
		for row := 0; row < numRows; row++ {
			flatIndex := 0
			for d := 0; d < len(dimensions); d++ {
				dimValue := dimColumnData[d][row]
				idx, ok := valueIndices[d][dimValue]
				if !ok {
					return nil, nil, nil, fmt.Errorf(
						"row group %d row %d: dimension value %q for dimension %q not found in conceptValues",
						rgIdx, row, dimValue, dimensions[d],
					)
				}
				flatIndex = flatIndex*shape[d] + idx
			}
			if seen[flatIndex] {
				if !duplicateFound {
					duplicateIndex = flatIndex
				}
				duplicateFound = true
			} else {
				seen[flatIndex] = true
				data[flatIndex] = floatValues[row]
			}
		}
		totalRowsRead += numRows
	}

	// Check for excess rows. Fewer rows than the product is allowed (sparse data);
	// only rows exceeding the Cartesian product are rejected.
	if totalRowsRead > totalSize {
		return nil, nil, nil, fmt.Errorf("expected at most %d rows but found %d (parquet file exceeds cartesian product of dimension sizes)", totalSize, totalRowsRead)
	}

	// Check for duplicate combinations.
	if duplicateFound {
		dimValues := dimIndexToValues(duplicateIndex, shape, conceptValues, dimensions, columnToConcept)
		return nil, nil, nil, fmt.Errorf("duplicate combination at flat index %d (dimension values: %s)", duplicateIndex, strings.Join(dimValues, ", "))
	}

	// Copy dimension names to ensure the returned slice is independent.
	dimensionNames = make([]string, len(dimensions))
	copy(dimensionNames, dimensions)

	return data, dimensionNames, shape, nil
}

// dimIndexToValues reverse-computes dimension values from a flat index.
func dimIndexToValues(flatIndex int, shape []int, conceptValues map[string][]string, dimensions []string, columnToConcept map[string]string) []string {
	values := make([]string, len(dimensions))
	remainder := flatIndex
	for i := len(shape) - 1; i >= 0; i-- {
		conceptName := dimensions[i]
		if columnToConcept != nil && columnToConcept[dimensions[i]] != "" {
			conceptName = columnToConcept[dimensions[i]]
		}
		values[i] = conceptValues[conceptName][remainder%shape[i]]
		remainder /= shape[i]
	}
	return values
}
