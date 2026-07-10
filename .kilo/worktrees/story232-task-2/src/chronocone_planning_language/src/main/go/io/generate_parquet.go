// Package io provides I/O utilities for the Chronocone Planning Language.
// This package contains helper functions for generating and reading parquet files
// with various dimensional structures for use in both normal operation and testing.
//
// Key functions:
//   - GenerateParquetFile: Creates a parquet file from dimension definitions for testing.
//   - ReadParquetToTensor: Reads a parquet file into raw tensor data (dense []float64,
//     dimension names, and shape). This is used by generated CPL code to load connector
//     data without creating an import cycle with the calculation package.
//   - ReadParquetFile / ReadParquetColumnData: Lower-level helpers for inspecting
//     parquet files in tests.
package io

import (
	"fmt"
	"math/rand/v2"
	"os"
	"sort"

	"github.com/apache/arrow/go/v16/parquet"
	"github.com/apache/arrow/go/v16/parquet/file"
	"github.com/apache/arrow/go/v16/parquet/schema"
)

// GrainMeasure defines a single grain measure column for parquet generation.
//
// Grain measures represent hierarchical temporal or categorical levels in a
// connector's grain clause. The order of GrainMeasure values in the slice
// defines the hierarchy: the first element is the highest level (varies
// slowest), and the last element is the lowest level (varies fastest).
type GrainMeasure struct {
	Name   string
	Values []string
}

// ConditionalDimMapping maps a child dimension to its parent dimension and
// the allowed child values for each parent value.
type ConditionalDimMapping struct {
	Parent string              `json:"parent"`
	Values map[string][]string `json:"values"`
}

// GenerateParquetFile creates a parquet file with cell values for testing purposes.
//
// This function generates a parquet file at the specified file path with a schema
// derived from the provided dimensions and optional grain measures. Each dimension
// becomes a string column in the parquet schema, each grain measure becomes a string
// column, and an additional "Value" column of type float64 is added to store the
// cell values.
//
// The function creates all combinations (Cartesian product) of dimension values and
// grain measure values, with each combination forming a row in the parquet file.
// When values is nil, random float64 values are assigned to each row's "Value"
// column. When values is non-nil, the provided values are used in order (one per
// row, matching the Cartesian product order).
//
// Dimension columns are sorted alphabetically by name. Grain measure columns appear
// in the order specified by the grain slice, after all dimension columns.
//
// For example, given dimensions:
//
//	map[string][]string{
//	  "Region":  {"North", "South"},
//	  "Product": {"A", "B"},
//	}
//
// And grain:
//
//	[]GrainMeasure{
//	  {Name: "Year", Values: []string{"2023", "2024"}},
//	}
//
// The resulting parquet file will have columns: "Product", "Region", "Year", "Value"
// with 8 rows representing all combinations.
//
// Parameters:
//   - filePath: The path where the parquet file will be created. If the file
//     already exists, it will be overwritten.
//   - dimensions: A map where keys are dimension names and values are slices
//     of possible dimension values. Each dimension becomes a string column
//     in the parquet schema.
//   - grain: A slice of GrainMeasure defining hierarchical grain measure columns.
//     When nil or empty, no grain columns are generated and behavior is equivalent
//     to dimensions-only generation.
//   - values: Optional fixed float64 values for the "Value" column. When non-nil,
//     the length must equal the Cartesian product of all dimension and grain
//     value counts. The values are used in the same order as the generated rows.
//     When nil, random values in [0.0, 1.0) are generated for each row.
//
// Returns:
//   - error: An error if the file cannot be created, if the schema cannot be
//     built, if writing the parquet data fails, or if any other I/O error
//     occurs. Returns nil on successful completion.
//
// Dependencies:
//   - github.com/apache/arrow/go/v16/parquet
//
// Example usage:
//
//	dimensions := map[string][]string{
//	    "Region":  {"North", "South", "East", "West"},
//	    "Product": {"Widget", "Gadget"},
//	}
//	err := GenerateParquetFile("test_data.parquet", dimensions, nil, nil)
//	if err != nil {
//	    log.Fatalf("Failed to generate parquet file: %v", err)
//	}
func GenerateParquetFile(filePath string, dimensions map[string][]string, grain []GrainMeasure, values []float64) error {
	// Get sorted dimension names for consistent column ordering
	dimensionNames := make([]string, 0, len(dimensions))
	for name := range dimensions {
		dimensionNames = append(dimensionNames, name)
	}
	sort.Strings(dimensionNames)

	// Extract grain measure names in specification order
	grainNames := make([]string, 0, len(grain))
	for _, g := range grain {
		grainNames = append(grainNames, g.Name)
	}

	// Build combined value map for cartesian product (dimensions + grain)
	combined := make(map[string][]string, len(dimensions)+len(grain))
	for k, v := range dimensions {
		combined[k] = v
	}
	for _, g := range grain {
		combined[g.Name] = g.Values
	}

	// Build combined ordered names: dimensions (alphabetical) then grain (spec order)
	allNames := make([]string, 0, len(dimensionNames)+len(grainNames))
	allNames = append(allNames, dimensionNames...)
	allNames = append(allNames, grainNames...)

	// Generate all combinations using Cartesian product
	combinations, err := generateCartesianProduct(combined, allNames)
	if err != nil {
		return fmt.Errorf("failed to generate cartesian product: %w", err)
	}

	return writeParquetFromCombinations(filePath, dimensionNames, grainNames, combinations, values)
}

// GenerateParquetFileWithConditionals creates a parquet file respecting conditional
// dimension mappings, producing rows only for valid (parent, child) combinations.
//
// This function behaves like GenerateParquetFile but accepts an additional
// conditionalDims parameter that maps a child dimension name to a ConditionalDimMapping
// specifying the parent dimension and the allowed child values per parent value.
// When generating combinations, if a dimension has conditional entries, only the
// child values allowed for the specific parent value already present in the partial
// combination are used.
//
// If conditionalDims is empty or nil, the function falls back to the full
// Cartesian product (same behavior as GenerateParquetFile).
//
// Parameters:
//   - filePath: The path where the parquet file will be created.
//   - dimensions: A map of dimension names to their possible values.
//   - grain: A slice of GrainMeasure defining hierarchical grain columns.
//   - values: Optional fixed float64 values for the "Value" column.
//   - conditionalDims: Maps child dimension name → ConditionalDimMapping with Parent
//     and Values. When a child dimension is processed and its parent value is
//     already in the partial combination, only the listed child values are used.
//
// Returns an error if file creation, schema building, or writing fails.
func GenerateParquetFileWithConditionals(
	filePath string,
	dimensions map[string][]string,
	grain []GrainMeasure,
	values []float64,
	conditionalDims map[string]ConditionalDimMapping,
) error {
	// Get sorted dimension names for consistent column ordering
	dimensionNames := make([]string, 0, len(dimensions))
	for name := range dimensions {
		dimensionNames = append(dimensionNames, name)
	}
	sort.Strings(dimensionNames)

	// Extract grain measure names in specification order
	grainNames := make([]string, 0, len(grain))
	for _, g := range grain {
		grainNames = append(grainNames, g.Name)
	}

	// Build combined value map for combination generation (dimensions + grain)
	combined := make(map[string][]string, len(dimensions)+len(grain))
	for k, v := range dimensions {
		combined[k] = v
	}
	for _, g := range grain {
		combined[g.Name] = g.Values
	}

	// Build combined ordered names: dimensions (alphabetical) then grain (spec order)
	allNames := make([]string, 0, len(dimensionNames)+len(grainNames))
	allNames = append(allNames, dimensionNames...)
	allNames = append(allNames, grainNames...)

	// Generate combinations respecting conditional mappings
	combinations, err := GenerateConditionalProduct(combined, allNames, conditionalDims)
	if err != nil {
		return fmt.Errorf("failed to generate conditional product: %w", err)
	}

	return writeParquetFromCombinations(filePath, dimensionNames, grainNames, combinations, values)
}

// writeParquetFromCombinations writes a parquet file from pre-generated combinations.
// This helper is shared between GenerateParquetFile and GenerateParquetFileWithConditionals.
func writeParquetFromCombinations(
	filePath string,
	dimensionNames []string,
	grainNames []string,
	combinations []map[string]string,
	values []float64,
) error {
	// Create the Parquet schema nodes
	nodeList := make(schema.FieldList, 0, len(dimensionNames)+len(grainNames)+1)

	// Add dimension fields (all string type using ByteArray with String logical type)
	for _, name := range dimensionNames {
		node := schema.NewByteArrayNode(name, parquet.Repetitions.Required, -1)
		nodeList = append(nodeList, node)
	}

	// Add grain measure fields (all string type using ByteArray with String logical type)
	for _, name := range grainNames {
		node := schema.NewByteArrayNode(name, parquet.Repetitions.Required, -1)
		nodeList = append(nodeList, node)
	}

	// Add Value field (float64)
	valueNode := schema.NewFloat64Node("Value", parquet.Repetitions.Required, -1)
	nodeList = append(nodeList, valueNode)

	// Create the schema group
	parquetSchema, err := schema.NewGroupNode("schema", parquet.Repetitions.Required, nodeList, -1)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Create the output file
	outputFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outputFile.Close()

	// Create writer properties
	props := parquet.NewWriterProperties()

	// Create the parquet writer
	writer := file.NewParquetWriter(outputFile, parquetSchema, file.WithWriterProps(props))
	defer writer.Close()

	// Create a column group writer (we'll write all columns in one row group)
	rowGroupWriter := writer.AppendRowGroup()

	// Write each dimension column
	for _, name := range dimensionNames {
		columnWriter, err := rowGroupWriter.NextColumn()
		if err != nil {
			return fmt.Errorf("failed to get column writer: %w", err)
		}

		stringWriter := columnWriter.(*file.ByteArrayColumnChunkWriter)
		colValues := make([]parquet.ByteArray, len(combinations))
		for i, combo := range combinations {
			colValues[i] = parquet.ByteArray(combo[name])
		}
		_, err = stringWriter.WriteBatch(colValues, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to write string column %s: %w", name, err)
		}
	}

	// Write each grain measure column
	for _, name := range grainNames {
		columnWriter, err := rowGroupWriter.NextColumn()
		if err != nil {
			return fmt.Errorf("failed to get column writer: %w", err)
		}

		stringWriter := columnWriter.(*file.ByteArrayColumnChunkWriter)
		colValues := make([]parquet.ByteArray, len(combinations))
		for i, combo := range combinations {
			colValues[i] = parquet.ByteArray(combo[name])
		}
		_, err = stringWriter.WriteBatch(colValues, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to write string column %s: %w", name, err)
		}
	}

	// Write the Value column
	valueColumnWriter, err := rowGroupWriter.NextColumn()
	if err != nil {
		return fmt.Errorf("failed to get value column writer: %w", err)
	}

	floatWriter := valueColumnWriter.(*file.Float64ColumnChunkWriter)
	floatValues := make([]float64, len(combinations))
	if values != nil {
		copy(floatValues, values)
	} else {
		for i := range combinations {
			floatValues[i] = rand.Float64()
		}
	}
	_, err = floatWriter.WriteBatch(floatValues, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to write value column: %w", err)
	}

	// Close the row group writer
	err = rowGroupWriter.Close()
	if err != nil {
		return fmt.Errorf("failed to close row group writer: %w", err)
	}

	// Close the writer
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}

// WriteParquetFromRows writes a parquet file directly from row data.
// Each row is a map from column name to string value, plus a float64 Value.
// Column names are sorted alphabetically in the output schema.
// No dimension/grain distinction is made — all non-Value columns are treated uniformly.
//
// Parameters:
//   - filePath: Path where the parquet file will be written.
//   - rows: Slice of row maps, each keyed by column name, plus a "Value" float64 field.
//   - values: Float64 values in the same order as rows. Must have len(rows) elements.
//
// Returns nil on success, or an error if file creation, schema building, or writing fails.
func WriteParquetFromRows(filePath string, rows []map[string]string, values []float64) error {
	if len(rows) == 0 {
		return nil
	}
	if len(values) != len(rows) {
		return fmt.Errorf("values length %d does not match rows length %d", len(values), len(rows))
	}

	// Collect column names from the first row, sorted alphabetically.
	colNames := make([]string, 0, len(rows[0]))
	for name := range rows[0] {
		colNames = append(colNames, name)
	}
	sort.Strings(colNames)

	// Create the Parquet schema: all string columns + "Value" float64 column.
	nodeList := make(schema.FieldList, 0, len(colNames)+1)
	for _, name := range colNames {
		nodeList = append(nodeList, schema.NewByteArrayNode(name, parquet.Repetitions.Required, -1))
	}
	nodeList = append(nodeList, schema.NewFloat64Node("Value", parquet.Repetitions.Required, -1))

	parquetSchema, err := schema.NewGroupNode("schema", parquet.Repetitions.Required, nodeList, -1)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	outputFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outputFile.Close()

	props := parquet.NewWriterProperties()
	writer := file.NewParquetWriter(outputFile, parquetSchema, file.WithWriterProps(props))
	defer writer.Close()

	rowGroupWriter := writer.AppendRowGroup()

	// Write each string column.
	for _, name := range colNames {
		columnWriter, err := rowGroupWriter.NextColumn()
		if err != nil {
			return fmt.Errorf("failed to get column writer for %q: %w", name, err)
		}
		stringWriter := columnWriter.(*file.ByteArrayColumnChunkWriter)
		colVals := make([]parquet.ByteArray, len(rows))
		for i, row := range rows {
			colVals[i] = parquet.ByteArray(row[name])
		}
		if _, err := stringWriter.WriteBatch(colVals, nil, nil); err != nil {
			return fmt.Errorf("failed to write string column %q: %w", name, err)
		}
	}

	// Write the Value column.
	valueWriter, err := rowGroupWriter.NextColumn()
	if err != nil {
		return fmt.Errorf("failed to get value column writer: %w", err)
	}
	floatWriter := valueWriter.(*file.Float64ColumnChunkWriter)
	floatVals := make([]float64, len(values))
	copy(floatVals, values)
	if _, err := floatWriter.WriteBatch(floatVals, nil, nil); err != nil {
		return fmt.Errorf("failed to write value column: %w", err)
	}

	if err := rowGroupWriter.Close(); err != nil {
		return fmt.Errorf("failed to close row group writer: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}

// GenerateConditionalProduct generates combinations respecting conditional mappings.
// It uses the explicit parent mapping from conditionalDims and produces rows only
// for valid (parent, child) combinations. When conditionalDims is empty it falls
// back to the full Cartesian product.
func GenerateConditionalProduct(
	combined map[string][]string,
	allNames []string,
	conditionalDims map[string]ConditionalDimMapping,
) ([]map[string]string, error) {
	if len(conditionalDims) == 0 {
		return generateCartesianProduct(combined, allNames)
	}

	// Build child-to-parent map from explicit mappings.
	childToParent := make(map[string]string)
	for childDim, mapping := range conditionalDims {
		if len(mapping.Values) > 0 && mapping.Parent == "" {
			return nil, fmt.Errorf("conditional mapping for %q has values but no parent", childDim)
		}
		if mapping.Parent != "" {
			childToParent[childDim] = mapping.Parent
		}
	}

	// Validate that every referenced parent exists in allNames.
	nameSet := make(map[string]struct{}, len(allNames))
	for _, name := range allNames {
		nameSet[name] = struct{}{}
	}
	for child, parent := range childToParent {
		if _, ok := nameSet[parent]; !ok {
			return nil, fmt.Errorf("conditional mapping for %q references parent %q which does not exist in dimensions or grain", child, parent)
		}
	}

	// Topological sort preserving original relative order among independent nodes.
	inDegree := make(map[string]int)
	for _, name := range allNames {
		inDegree[name] = 0
	}
	adj := make(map[string][]string)
	for _, name := range allNames {
		adj[name] = nil
	}
	for child, parent := range childToParent {
		adj[parent] = append(adj[parent], child)
		inDegree[child]++
	}

	generationOrder := make([]string, 0, len(allNames))
	processed := make(map[string]bool)
	for len(generationOrder) < len(allNames) {
		found := false
		for _, name := range allNames {
			if processed[name] {
				continue
			}
			if inDegree[name] == 0 {
				generationOrder = append(generationOrder, name)
				processed[name] = true
				for _, child := range adj[name] {
					inDegree[child]--
				}
				found = true
				break
			}
		}
		// Cycle detected: all node ordering constraints cannot be satisfied.
		// Return an error rather than silently producing empty or incorrect results.
		if !found {
			return nil, fmt.Errorf("cycle detected in conditional dimension mappings: %d node(s) cannot be scheduled", len(allNames)-len(generationOrder))
		}
	}

	result := []map[string]string{{}}
	for _, name := range generationOrder {
		values := combined[name]
		if len(values) == 0 {
			return nil, fmt.Errorf("dimension %q has no values", name)
		}

		mapping, hasConditionals := conditionalDims[name]
		parentName := childToParent[name]

		newResult := make([]map[string]string, 0, len(result)*len(values))
		for _, combo := range result {
			if hasConditionals && len(mapping.Values) > 0 && parentName != "" {
				parentValue := combo[parentName]
				if childValues, ok := mapping.Values[parentValue]; ok && len(childValues) > 0 {
					for _, value := range childValues {
						newCombo := make(map[string]string, len(combo)+1)
						for k, v := range combo {
							newCombo[k] = v
						}
						newCombo[name] = value
						newResult = append(newResult, newCombo)
					}
				}
				// If parentValue not in mapping.Values or childValues empty, skip this branch
			} else {
				for _, value := range values {
					newCombo := make(map[string]string, len(combo)+1)
					for k, v := range combo {
						newCombo[k] = v
					}
					newCombo[name] = value
					newResult = append(newResult, newCombo)
				}
			}
		}
		result = newResult
	}

	return result, nil
}

// generateCartesianProduct generates all combinations of dimension values.
// It returns a slice of maps, where each map represents one combination
// of dimension values.
func generateCartesianProduct(dimensions map[string][]string, dimensionNames []string) ([]map[string]string, error) {
	// Start with a single empty combination
	result := []map[string]string{{}}

	// For each dimension, expand the combinations
	for _, name := range dimensionNames {
		values := dimensions[name]
		if len(values) == 0 {
			return nil, fmt.Errorf("dimension %q has no values", name)
		}

		newResult := make([]map[string]string, 0, len(result)*len(values))

		for _, combo := range result {
			for _, value := range values {
				newCombo := make(map[string]string, len(combo)+1)
				for k, v := range combo {
					newCombo[k] = v
				}
				newCombo[name] = value
				newResult = append(newResult, newCombo)
			}
		}

		result = newResult
	}

	return result, nil
}

// ReadParquetFile reads a parquet file and returns the schema and number of rows.
// This is a helper function for testing purposes.
func ReadParquetFile(filePath string) (*schema.Schema, int, error) {
	inputFile, err := os.Open(filePath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer inputFile.Close()

	reader, err := file.NewParquetReader(inputFile)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create parquet reader: %w", err)
	}
	defer reader.Close()

	fileMeta := reader.MetaData()
	numRows := int(reader.NumRows())

	return fileMeta.Schema, numRows, nil
}

// GetColumnNames returns the column names from a parquet schema.
// This is a helper function for testing purposes.
func GetColumnNames(parquetSchema *schema.Schema) []string {
	if parquetSchema == nil {
		return nil
	}

	names := make([]string, 0, parquetSchema.NumColumns())
	for i := 0; i < parquetSchema.NumColumns(); i++ {
		col := parquetSchema.Column(i)
		names = append(names, col.Name())
	}
	return names
}

// GetColumnType returns the type of a column by name.
// This is a helper function for testing purposes.
func GetColumnType(parquetSchema *schema.Schema, columnName string) parquet.Type {
	if parquetSchema == nil {
		return parquet.Types.Undefined
	}

	for i := 0; i < parquetSchema.NumColumns(); i++ {
		col := parquetSchema.Column(i)
		if col.Name() == columnName {
			return col.PhysicalType()
		}
	}
	return parquet.Types.Undefined
}

// ReadParquetColumnData reads a column's data from a parquet file.
// This is a helper function for testing purposes.
func ReadParquetColumnData(filePath string, columnIndex int) (interface{}, error) {
	inputFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer inputFile.Close()

	reader, err := file.NewParquetReader(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet reader: %w", err)
	}
	defer reader.Close()

	// Get the column chunk reader
	rowGroupReader := reader.RowGroup(0)
	columnReader, err := rowGroupReader.Column(columnIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get column reader: %w", err)
	}

	// Read based on column type
	colType := reader.MetaData().Schema.Column(columnIndex).PhysicalType()
	numRows := int(reader.NumRows())

	switch colType {
	case parquet.Types.ByteArray:
		stringReader := columnReader.(*file.ByteArrayColumnChunkReader)
		values := make([]parquet.ByteArray, numRows)
		totalRead := 0
		for totalRead < numRows {
			batchRead, _, err := stringReader.ReadBatch(int64(numRows-totalRead), values[totalRead:], nil, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to read string batch: %w", err)
			}
			totalRead += int(batchRead)
		}
		// Convert to []string
		result := make([]string, len(values))
		for i, v := range values {
			result[i] = string(v)
		}
		return result, nil

	case parquet.Types.Double:
		floatReader := columnReader.(*file.Float64ColumnChunkReader)
		values := make([]float64, numRows)
		totalRead := 0
		for totalRead < numRows {
			batchRead, _, err := floatReader.ReadBatch(int64(numRows-totalRead), values[totalRead:], nil, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to read float batch: %w", err)
			}
			totalRead += int(batchRead)
		}
		return values, nil

	default:
		return nil, fmt.Errorf("unsupported column type: %v", colType)
	}
}
