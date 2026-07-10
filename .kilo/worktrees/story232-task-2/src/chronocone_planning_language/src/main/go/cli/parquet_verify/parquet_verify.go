// parquet_verify is a command-line tool for verifying the contents of a
// synthesized output parquet file against an expected JSON specification.
//
// Usage:
//
//	parquet_verify -parquet=<path_to_parquet> -expected=<path_to_expected_json>
//
// Command-line Flags:
//
//	-parquet string
//	    Path to the parquet file to verify (required).
//	    This is typically the output.parquet file produced by the CPL synthesis pipeline.
//	-expected string
//	    Path to the expected output JSON file (required).
//	    This file defines the expected schema, row count, and cell values
//	    for comparison against the parquet file.
//
// Expected JSON File Format:
//
// The expected output JSON file defines the expected content of a synthesized
// output parquet file. It has the following structure:
//
//	{
//	    "output_file": "output.parquet",
//	    "metrics": ["sum_metric"],
//	    "dimensions": ["Region"],
//	    "grain": ["Day"],
//	    "rows": [
//	        {
//	            "Metric": "sum_metric",
//	            "dimension_values": { "Region": "North" },
//	            "grain_values": { "Day": "Monday" },
//	            "value": 10.5
//	        }
//	    ]
//	}
//
// Exit Codes:
//
//	0 — All rows match expected values. Verification passed.
//	1 — Differences found between parquet and expected output. Details are
//	    printed to stdout before exiting.
//	2 — Error condition: file not found, JSON parse error, schema mismatch,
//	    or any other unrecoverable error. Details are printed to stderr.
//
// Verification Workflow:
//
// The tool performs the following steps in order:
//  1. Read the parquet file using [cplio.ReadParquetFile] to obtain schema
//     metadata and total row count.
//  2. Parse the expected JSON file into an [ExpectedOutput] struct.
//  3. Call [verifySchema] to confirm the parquet schema contains exactly the
//     expected columns with correct types.
//  4. Call [verifyRowCount] to confirm the parquet row count matches the
//     expected row count.
//  5. Call [matchRows] to pair each expected row with a parquet row and
//     compare values, collecting any [RowMismatch] results.
//  6. If mismatches were found, call [printDifferences] to report them to
//     stdout and exit with code 1.
//  7. If no mismatches, exit with code 0.
//
// Dependencies:
//   - [cplio.ReadParquetFile] from the chronocone_planning_language io package
//   - encoding/json for expected JSON parsing
//   - flag for CLI argument parsing
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"

	cplio "github.com/chronocone/chronocone_planning_language/src/main/go/io"

	"github.com/apache/arrow/go/v16/parquet"
	"github.com/apache/arrow/go/v16/parquet/schema"
)

// ExpectedOutput represents the expected_output.json file format.
//
// It defines the expected content of a synthesized output parquet file:
// the file name (for reference), the list of metrics present in the output,
// the ordered list of dimension column names, the ordered list of grain
// column names, and the complete set of expected rows with their values.
type ExpectedOutput struct {
	// OutputFile is the name of the output parquet file (e.g., "output.parquet").
	// Included for cross-reference; the verification tool does not use this
	// field to locate the parquet file (the -parquet flag serves that purpose).
	OutputFile string `json:"output_file"`
	// Metrics lists the metric names present in the output. The output parquet
	// file's "Metric" column must contain only values from this list.
	Metrics []string `json:"metrics"`
	// Dimensions lists the dimension column names in the expected order.
	// The parquet schema must contain a string column for each dimension.
	Dimensions []string `json:"dimensions"`
	// Grain lists the grain column names in the expected order.
	// The parquet schema must contain a string column for each grain.
	Grain []string `json:"grain"`
	// Rows contains the complete set of expected output rows.
	// The parquet file must contain exactly these rows (same count, same
	// metric/dimension/grain values, same cell values within tolerance).
	Rows []ExpectedRow `json:"rows"`
}

// ExpectedRow represents one expected row in the output.
//
// Each row identifies the metric it belongs to, the dimension values that
// locate the cell in the output tensor, the grain values that identify the
// temporal or categorical level, and the expected float64 cell value.
type ExpectedRow struct {
	// Metric identifies which output metric this row belongs to (e.g., "sum_metric").
	// Must match one of the values in ExpectedOutput.Metrics.
	Metric string `json:"Metric"`
	// DimensionValues maps each dimension name to its string value for this row.
	// Keys must match the names in ExpectedOutput.Dimensions.
	DimensionValues map[string]string `json:"dimension_values"`
	// GrainValues maps each grain name to its string value for this row.
	// Keys must match the names in ExpectedOutput.Grain.
	GrainValues map[string]string `json:"grain_values"`
	// Value is the expected float64 cell value for this row.
	// Compared against the parquet "Value" column using [compareValues]
	// with a default tolerance of 1e-9.
	Value float64 `json:"value"`
}

// RowMismatch represents a difference between expected and actual output.
//
// It captures the type of mismatch (missing row, extra row, or value
// difference), the identifying information for the row in question,
// and the expected and actual values.
type RowMismatch struct {
	// Type describes the kind of mismatch:
	//   - "missing": The row is in expected_output.json but not in the parquet file.
	//   - "extra": The row is in the parquet file but not in expected_output.json.
	//   - "value_mismatch": The row exists in both but the Value differs beyond tolerance.
	Type string
	// ExpectedRowIndex is the 0-based index of the failing row in ExpectedOutput.Rows.
	// Set to -1 for "extra" rows (which have no corresponding expected entry).
	ExpectedRowIndex int
	// Metric identifies the metric for this row.
	Metric string
	// DimensionValues holds the dimension key-value pairs that locate this row.
	DimensionValues map[string]string
	// GrainValues holds the grain key-value pairs that locate this row.
	GrainValues map[string]string
	// ExpectedValue is the value from the expected JSON.
	// Set to 0 for "missing" and "extra" types (not applicable).
	ExpectedValue float64
	// ActualValue is the value read from the parquet file.
	// Set to 0 for "missing" type (not applicable).
	ActualValue float64
}

// parquetRowData represents a row read from a parquet file.
// It mirrors the structure consumed by matchRows:
// a Metric column, dimension value columns, grain value columns, and a Value column.
type parquetRowData struct {
	Metric          string
	DimensionValues map[string]string
	GrainValues     map[string]string
	Value           float64
}

// verificationError is an error type that signals verification differences
// were found (as opposed to an I/O or other unrecoverable error).
// main() uses this to distinguish exit code 1 from exit code 2.
type verificationError struct {
	mismatchCount int
}

func (e *verificationError) Error() string {
	return fmt.Sprintf("verification failed: %d differences found", e.mismatchCount)
}

// defaultTolerance is the default absolute tolerance for float64 value comparison.
const defaultTolerance = 1e-9

// main is the entry point for the parquet_verify CLI tool.
//
// It parses command-line flags (-parquet and -expected), validates that both
// required flags are provided, and invokes the [verify] function. On error,
// it prints the error message to stderr and exits with code 2. On success,
// it exits with code 0. If verify returns an error indicating differences
// were found, the differences have already been printed to stdout by
// [printDifferences] and the process exits with code 1.
//
// Exit codes:
//   - 0: All rows match expected values
//   - 1: Differences found (details printed to stdout)
//   - 2: Error (file not found, parse error, schema mismatch, etc.)
func main() {
	var parquetPath string
	var expectedJSONPath string

	flag.StringVar(&parquetPath, "parquet", "", "Path to the parquet file to verify (required)")
	flag.StringVar(&expectedJSONPath, "expected", "", "Path to the expected output JSON file (required)")

	flag.Parse()

	// Validate required flags
	if parquetPath == "" || expectedJSONPath == "" {
		flag.Usage()
		os.Exit(2)
	}

	// Execute the verification workflow
	err := verify(parquetPath, expectedJSONPath)
	if err == nil {
		os.Exit(0)
	}

	// Check if this is a verification failure (exit 1) or a real error (exit 2)
	if _, ok := err.(*verificationError); ok {
		// Differences have already been printed to stdout by printDifferences
		// Print error to stderr for logging/debugging
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(2)
}

// verify orchestrates the parquet verification workflow.
//
// It reads the parquet file and expected JSON, then performs a series of
// verification steps: schema validation, row count validation, row matching,
// and value comparison. If any mismatches are found, they are printed to
// stdout via [printDifferences] and verify returns an error to signal that
// differences exist (causing the caller to exit with code 1).
//
// Parameters:
//   - parquetPath: Path to the parquet file to verify.
//   - expectedJSONPath: Path to the expected output JSON specification.
//
// Returns:
//   - error: If file reading fails, JSON parsing fails, schema verification
//     fails, or differences are found between expected and actual data.
//     Returns nil when all rows match exactly.
func verify(parquetPath, expectedJSONPath string) error {
	// 1. Read the parquet file
	parquetSchema, numRows, err := cplio.ReadParquetFile(parquetPath)
	if err != nil {
		return fmt.Errorf("failed to read parquet file: %w", err)
	}

	// 2. Read and parse the expected JSON file
	jsonData, err := os.ReadFile(expectedJSONPath)
	if err != nil {
		return fmt.Errorf("failed to read expected JSON file: %w", err)
	}

	var expected ExpectedOutput
	if err := json.Unmarshal(jsonData, &expected); err != nil {
		return fmt.Errorf("failed to parse expected JSON: %w", err)
	}

	// 3. Verify schema
	if err := verifySchema(parquetSchema, &expected); err != nil {
		return err
	}

	// 4. Verify row count
	if err := verifyRowCount(numRows, &expected); err != nil {
		return err
	}

	// 5. Read parquet row data and match against expected rows
	parquetRows, err := readParquetRows(parquetPath, parquetSchema, expected.Dimensions, expected.Grain)
	if err != nil {
		return fmt.Errorf("failed to read parquet row data: %w", err)
	}

	mismatches, err := matchRows(parquetRows, &expected)
	if err != nil {
		return fmt.Errorf("failed to match rows: %w", err)
	}

	// 6. If mismatches found, print differences and return verification error
	if len(mismatches) > 0 {
		printDifferences(mismatches)
		return &verificationError{mismatchCount: len(mismatches)}
	}

	// 7. All matched
	return nil
}

// verifySchema compares the parquet file schema against the expected output
// specification.
//
// It verifies that the parquet file contains exactly the expected columns:
//   - A "Metric" column of string type (ByteArray)
//   - All expected dimension columns (string type, ByteArray)
//   - All expected grain columns (string type, ByteArray)
//   - A "Value" column of float64 type (Double)
//
// No extra columns beyond these should exist. The column order is not
// validated (parquet column order is not significant for verification).
//
// Parameters:
//   - parquetSchema: The schema read from the parquet file by
//     [cplio.ReadParquetFile]. The concrete type is
//     *github.com/apache/arrow/go/v16/parquet/schema.Schema.
//   - expected: The parsed ExpectedOutput containing the expected column
//     definitions.
//
// Returns:
//   - error: If any required column is missing, has the wrong physical type,
//     or extra columns are present. Returns nil when the schema matches all
//     expectations.
func verifySchema(parquetSchema *schema.Schema, expected *ExpectedOutput) error {
	// Build a set of expected column names
	expectedColumns := make(map[string]bool)
	expectedColumns["Metric"] = true
	expectedColumns["Value"] = true
	for _, dim := range expected.Dimensions {
		expectedColumns[dim] = true
	}
	for _, grain := range expected.Grain {
		expectedColumns[grain] = true
	}

	// Build a set of actual column names and check types
	actualColumns := make(map[string]parquet.Type)
	for i := 0; i < parquetSchema.NumColumns(); i++ {
		col := parquetSchema.Column(i)
		name := col.Name()
		actualColumns[name] = col.PhysicalType()

		// Check for extra columns not in expected set
		if !expectedColumns[name] {
			return fmt.Errorf("unexpected column %q in parquet schema (not in expected output)", name)
		}
	}

	// Verify "Metric" column exists and is ByteArray
	metricType, hasMetric := actualColumns["Metric"]
	if !hasMetric {
		return fmt.Errorf("parquet schema missing required column %q", "Metric")
	}
	if metricType != parquet.Types.ByteArray {
		return fmt.Errorf("column %q has type %v but expected ByteArray (string)", "Metric", metricType)
	}

	// Verify each dimension column exists and is ByteArray
	for _, dim := range expected.Dimensions {
		dimType, ok := actualColumns[dim]
		if !ok {
			return fmt.Errorf("parquet schema missing expected dimension column %q", dim)
		}
		if dimType != parquet.Types.ByteArray {
			return fmt.Errorf("dimension column %q has type %v but expected ByteArray (string)", dim, dimType)
		}
	}

	// Verify each grain column exists and is ByteArray
	for _, grain := range expected.Grain {
		grainType, ok := actualColumns[grain]
		if !ok {
			return fmt.Errorf("parquet schema missing expected grain column %q", grain)
		}
		if grainType != parquet.Types.ByteArray {
			return fmt.Errorf("grain column %q has type %v but expected ByteArray (string)", grain, grainType)
		}
	}

	// Verify "Value" column exists and is Double (float64)
	valueType, hasValue := actualColumns["Value"]
	if !hasValue {
		return fmt.Errorf("parquet schema missing required column %q", "Value")
	}
	if valueType != parquet.Types.Double {
		return fmt.Errorf("column %q has type %v but expected Double (float64)", "Value", valueType)
	}

	return nil
}

// verifyRowCount compares the number of rows in the parquet file against
// the number of rows in the expected output specification.
//
// Parameters:
//   - actualRowCount: The number of rows read from the parquet file
//     (returned by [cplio.ReadParquetFile]).
//   - expected: The parsed ExpectedOutput containing the expected rows.
//
// Returns:
//   - error: If actualRowCount != len(expected.Rows), with a message
//     describing the difference (e.g., "expected 4 rows but found 3").
//     Returns nil when the counts match.
func verifyRowCount(actualRowCount int, expected *ExpectedOutput) error {
	expectedCount := len(expected.Rows)
	if actualRowCount != expectedCount {
		return fmt.Errorf("expected %d rows but found %d", expectedCount, actualRowCount)
	}
	return nil
}

// matchRows pairs expected rows with actual parquet rows and compares values.
//
// For each expected row, it finds the matching parquet row by matching all of:
//   - Metric column value
//   - All dimension column values
//   - All grain column values
//
// Each expected row must match exactly one parquet row (no missing matches,
// no duplicate matches). After matching, the Value column is compared using
// [compareValues] with a default tolerance of 1e-9.
//
// Any rows in the parquet file that do not match an expected row are reported
// as "extra" mismatches. Any expected rows that do not match a parquet row
// are reported as "missing" mismatches.
//
// Parameters:
//   - parquetRows: The row data read from the parquet file, as []parquetRowData.
//   - expected: The parsed ExpectedOutput containing expected row data.
//
// Returns:
//   - []RowMismatch: A slice describing all mismatches found. Empty when all
//     rows match perfectly (same count, same values within tolerance).
//   - error: If row matching encounters an unrecoverable error (e.g., data
//     structure mismatch). Mismatches themselves are returned in the slice
//     rather than as errors. Returns nil on success.
func matchRows(parquetRows interface{}, expected *ExpectedOutput) ([]RowMismatch, error) {
	rows, ok := parquetRows.([]parquetRowData)
	if !ok {
		return nil, fmt.Errorf("invalid parquetRows type: expected []parquetRowData")
	}

	var mismatches []RowMismatch

	// Build a lookup index from parquet rows keyed by (Metric, dimension_values, grain_values)
	// Key format: "Metric|dim1=val1,dim2=val2,...|grain1=val1,grain2=val2,..."
	type rowKey struct {
		metric   string
		dimKey   string
		grainKey string
	}
	parquetByKey := make(map[rowKey]*parquetRowData)
	matchedKeys := make(map[rowKey]bool)

	for i := range rows {
		k := rowKey{
			metric:   rows[i].Metric,
			dimKey:   mapToKeyString(rows[i].DimensionValues),
			grainKey: mapToKeyString(rows[i].GrainValues),
		}
		parquetByKey[k] = &rows[i]
	}

	// For each expected row, look up the matching parquet row
	for i, expRow := range expected.Rows {
		k := rowKey{
			metric:   expRow.Metric,
			dimKey:   mapToKeyString(expRow.DimensionValues),
			grainKey: mapToKeyString(expRow.GrainValues),
		}

		pRow, found := parquetByKey[k]
		if !found {
			mismatches = append(mismatches, RowMismatch{
				Type:             "missing",
				ExpectedRowIndex: i,
				Metric:           expRow.Metric,
				DimensionValues:  expRow.DimensionValues,
				GrainValues:      expRow.GrainValues,
				ExpectedValue:    expRow.Value,
			})
			continue
		}

		matchedKeys[k] = true

		if !compareValues(expRow.Value, pRow.Value, defaultTolerance) {
			mismatches = append(mismatches, RowMismatch{
				Type:             "value_mismatch",
				ExpectedRowIndex: i,
				Metric:           expRow.Metric,
				DimensionValues:  expRow.DimensionValues,
				GrainValues:      expRow.GrainValues,
				ExpectedValue:    expRow.Value,
				ActualValue:      pRow.Value,
			})
		}
	}

	// After processing all expected rows, any unmatched parquet rows are "extra"
	for k, pRow := range parquetByKey {
		if !matchedKeys[k] {
			mismatches = append(mismatches, RowMismatch{
				Type:             "extra",
				ExpectedRowIndex: -1,
				Metric:           pRow.Metric,
				DimensionValues:  pRow.DimensionValues,
				GrainValues:      pRow.GrainValues,
				ActualValue:      pRow.Value,
			})
		}
	}

	return mismatches, nil
}

// mapToKeyString converts a map[string]string to a stable string key.
// Keys are sorted for deterministic output. Empty maps produce an empty string.
func mapToKeyString(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	// Collect and sort keys for deterministic ordering
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Simple bubble sort for small maps (dimension/grain maps are typically small)
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	result := ""
	for i, k := range keys {
		if i > 0 {
			result += ","
		}
		result += k + "=" + m[k]
	}
	return result
}

// compareValues compares two float64 values within a given absolute tolerance.
//
// The comparison uses absolute difference: |expected - actual| <= tolerance.
// This handles floating-point rounding errors that naturally occur in
// numeric computations.
//
// Parameters:
//   - expected: The expected float64 value from expected_output.json.
//   - actual: The actual float64 value read from the parquet file.
//   - tolerance: The maximum allowed absolute difference. Values differing
//     by more than this amount are considered unequal.
//
// Returns:
//   - bool: true if the values are equal within tolerance (|expected - actual| <= tolerance),
//     false otherwise.
func compareValues(expected, actual float64, tolerance float64) bool {
	diff := math.Abs(expected - actual)
	// Scale tolerance by a small relative factor (1e-7) to account for
	// IEEE 754 float64 precision loss. When operands span ~30 orders of
	// magnitude (e.g., 10.5 + 1e-9), the addition loses ~30 mantissa bits,
	// making the computed difference slightly larger than the mathematical
	// tolerance. A relative factor of 1e-7 provides ~1e-16 absolute margin
	// at tolerance=1e-9, which is sufficient for the ~8e-17 observed error.
	return diff <= tolerance*(1+1e-7)
}

// printDifferences outputs mismatch details to stdout in a human-readable format.
//
// Each mismatch is printed on one or more lines with clear labeling:
//   - "missing": Row is in expected but not found in parquet
//   - "extra": Row is in parquet but not in expected
//   - "value_mismatch": Row exists in both but values differ
//
// Output format example:
//
//	Row mismatch at expected row 3 (Metric=profit, Region=North, Product=Widget):
//	  Value: expected 100.0, got 99.999999
//	Extra row in parquet (not in expected):
//	  Metric=profit, Region=South, Product=Widget, Value=200.0
//	Missing row (in expected but not in parquet):
//	  Metric=margin_pct, Region=North, Product=Widget, Value=20.0
//
// Parameters:
//   - mismatches: A slice of RowMismatch structs describing the differences
//     between expected and actual data. If empty, no output is produced.
func printDifferences(mismatches []RowMismatch) {
	if len(mismatches) == 0 {
		return
	}

	for _, m := range mismatches {
		switch m.Type {
		case "missing":
			fmt.Printf("Missing row (in expected but not in parquet):\n")
			fmt.Printf("  Metric=%s", m.Metric)
			for k, v := range m.DimensionValues {
				fmt.Printf(", %s=%s", k, v)
			}
			for k, v := range m.GrainValues {
				fmt.Printf(", %s=%s", k, v)
			}
			fmt.Printf(", Value=%v\n", m.ExpectedValue)

		case "extra":
			fmt.Printf("Extra row in parquet (not in expected):\n")
			fmt.Printf("  Metric=%s", m.Metric)
			for k, v := range m.DimensionValues {
				fmt.Printf(", %s=%s", k, v)
			}
			for k, v := range m.GrainValues {
				fmt.Printf(", %s=%s", k, v)
			}
			fmt.Printf(", Value=%v\n", m.ActualValue)

		case "value_mismatch":
			fmt.Printf("Row mismatch at expected row %d (Metric=%s", m.ExpectedRowIndex, m.Metric)
			for k, v := range m.DimensionValues {
				fmt.Printf(", %s=%s", k, v)
			}
			for k, v := range m.GrainValues {
				fmt.Printf(", %s=%s", k, v)
			}
			fmt.Printf("):\n")
			fmt.Printf("  Value: expected %v, got %v\n", m.ExpectedValue, m.ActualValue)
		}
	}
}

// readParquetRows reads all row data from a parquet file and returns it as
// a slice of parquetRowData structs.
//
// It reads each column from the parquet file using [cplio.ReadParquetColumnData],
// then assembles the column data into rows keyed by column name. The dimensions
// and grain slices specify which columns map to DimensionValues and GrainValues
// in each row. The "Metric" column (if present) maps to the Metric field, and
// the "Value" column maps to the Value field.
//
// Parameters:
//   - filePath: Path to the parquet file.
//   - sch: The schema of the parquet file (for column count).
//   - dimensions: The ordered list of dimension column names.
//   - grain: The ordered list of grain column names.
//
// Returns:
//   - []parquetRowData: The complete row data from the parquet file.
//   - error: If a column read fails.
func readParquetRows(filePath string, sch *schema.Schema, dimensions []string, grain []string) ([]parquetRowData, error) {
	numCols := sch.NumColumns()
	if numCols == 0 {
		return nil, nil
	}

	// Read each column's data
	type columnData struct {
		name    string
		strVals []string
		f64Vals []float64
	}
	columns := make([]columnData, numCols)

	var numRows int
	for i := 0; i < numCols; i++ {
		col := sch.Column(i)
		columns[i].name = col.Name()

		data, err := cplio.ReadParquetColumnData(filePath, i)
		if err != nil {
			return nil, fmt.Errorf("failed to read column %q: %w", col.Name(), err)
		}

		switch v := data.(type) {
		case []string:
			columns[i].strVals = v
			if numRows == 0 {
				numRows = len(v)
			}
		case []float64:
			columns[i].f64Vals = v
			if numRows == 0 {
				numRows = len(v)
			}
		}
	}

	if numRows == 0 {
		return nil, nil
	}

	// Build a set of dimension and grain names for fast lookup
	dimSet := make(map[string]bool, len(dimensions))
	for _, d := range dimensions {
		dimSet[d] = true
	}
	grainSet := make(map[string]bool, len(grain))
	for _, g := range grain {
		grainSet[g] = true
	}

	// Assemble rows
	rows := make([]parquetRowData, numRows)
	for i := 0; i < numRows; i++ {
		rows[i].DimensionValues = make(map[string]string)
		rows[i].GrainValues = make(map[string]string)
	}

	for _, col := range columns {
		switch col.name {
		case "Metric":
			for rowIdx, v := range col.strVals {
				rows[rowIdx].Metric = v
			}
		case "Value":
			for rowIdx, v := range col.f64Vals {
				rows[rowIdx].Value = v
			}
		default:
			if dimSet[col.name] {
				for rowIdx, v := range col.strVals {
					rows[rowIdx].DimensionValues[col.name] = v
				}
			} else if grainSet[col.name] {
				for rowIdx, v := range col.strVals {
					rows[rowIdx].GrainValues[col.name] = v
				}
			}
		}
	}

	return rows, nil
}
