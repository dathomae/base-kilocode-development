// parquetgen is a command-line tool for generating parquet files from data JSON files.
//
// Usage:
//
//	parquetgen -input=<path_to_json> -output=<path_to_parquet>
//
// Command-line Flags:
//
//	-input string
//	    Path to a data JSON file in row format: a JSON array of objects,
//	    each containing string dimension/grain columns and a "Value" float64 field.
//	    Example:
//	        [{"Region":"North","Daily":"Sunday","Value":10}, ...]
//	-output string
//	    Path for the output parquet file.
//
// The input JSON must be an array of row objects. All keys except "Value" are treated
// as string columns written in alphabetical order, followed by a float64 "Value" column.
//
// Error Handling:
//   - Missing required flags: prints usage message and exits with code 1
//   - Invalid JSON file: returns error with details about parsing failure
//   - File read/write errors: returns error with file path and underlying cause
//
// Example:
//
//	parquetgen -input=./data_a.json -output=./data_a.parquet
//
// Dependencies:
//
// This tool uses the github.com/apache/arrow/go/v16/parquet library for parquet file generation.
// The core generation logic is provided by cplio.WriteParquetFromRows.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	cplio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
)

// main is the entry point for the parquetgen CLI tool.
// It parses command-line flags and executes the parquet generation workflow.
// On error, it prints the error message to stderr and exits with code 1.
func main() {
	var inputPath string
	var outputPath string

	flag.StringVar(&inputPath, "input", "", "Path to data JSON file in row format (required)")
	flag.StringVar(&outputPath, "output", "", "Path for output parquet file (required)")

	flag.Parse()

	if inputPath == "" || outputPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(inputPath, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run reads a data JSON file and writes the corresponding parquet file.
func run(inputPath, outputPath string) error {
	rows, values, err := parseRowFormat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to parse input: %w", err)
	}

	if err := cplio.WriteParquetFromRows(outputPath, rows, values); err != nil {
		return fmt.Errorf("failed to write parquet: %w", err)
	}

	return nil
}

// parseRowFormat reads a data JSON file in row-array format and returns the
// column-value maps and the extracted Value column.
//
// The expected JSON format is an array of row objects, each containing string
// dimension/grain columns and a float64 "Value" field:
//
//	[
//	  {"Region": "North", "Daily": "Sunday", "Value": 10},
//	  ...
//	]
//
// Returns the rows (maps without "Value" key) and the values slice in order.
func parseRowFormat(jsonPath string) ([]map[string]string, []float64, error) {
	file, err := os.Open(jsonPath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var raw []map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&raw); err != nil {
		return nil, nil, fmt.Errorf("invalid row-format JSON: %w", err)
	}

	if len(raw) == 0 {
		return nil, nil, fmt.Errorf("input JSON has no rows")
	}

	rows := make([]map[string]string, len(raw))
	values := make([]float64, len(raw))

	for i, obj := range raw {
		rowMap := make(map[string]string, len(obj))
		for key, val := range obj {
			if key == "Value" {
				fv, ok := toFloat64(val)
				if !ok {
					return nil, nil, fmt.Errorf("row %d: Value is not a number: %v", i, val)
				}
				values[i] = fv
			} else {
				s, ok := val.(string)
				if !ok {
					return nil, nil, fmt.Errorf("row %d: column %q is not a string: %v", i, key, val)
				}
				rowMap[key] = s
			}
		}
		rows[i] = rowMap
	}

	return rows, values, nil
}

// toFloat64 converts a JSON-decoded value to float64.
func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}
