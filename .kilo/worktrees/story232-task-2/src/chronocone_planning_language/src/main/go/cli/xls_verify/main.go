// Package main contains the xls_verify CLI tool for comparing XLSX expected_output
// against parquet results using composite keys and tolerance-based comparison.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	cplio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
	gen "github.com/chronocone/chronocone_planning_language/src/main/go/cli/expected_xls_generator"
	"github.com/xuri/excelize/v2"
)

// MismatchEntry represents a single mismatch between expected and actual values.
type MismatchEntry struct {
	Type     string
	Key      string
	Expected float64
	Actual   float64
}

// compareValues checks if two floating-point values are within tolerance.
func compareValues(expected, actual, tolerance float64) bool {
	if expected == 0 && actual == 0 {
		return true
	}
	if math.IsNaN(expected) || math.IsNaN(actual) {
		return false
	}
	diff := math.Abs(expected - actual)
	maxVal := math.Max(math.Abs(expected), math.Abs(actual))
	return diff <= tolerance*(1+maxVal*0.5)
}

// buildCompositeKey joins multiple string parts with "|" separator.
func buildCompositeKey(parts []string) string {
	return strings.Join(parts, "|")
}

// findMismatches compares expected and actual maps and returns list of mismatches.
func findMismatches(expected, actual map[string]float64, tolerance float64) []MismatchEntry {
	var mismatches []MismatchEntry
	for key, expVal := range expected {
		actVal, ok := actual[key]
		if !ok {
			mismatches = append(mismatches, MismatchEntry{Type: "missing", Key: key, Expected: expVal})
		} else if !compareValues(expVal, actVal, tolerance) {
			mismatches = append(mismatches, MismatchEntry{Type: "value_mismatch", Key: key, Expected: expVal, Actual: actVal})
		}
	}
	for key, actVal := range actual {
		if _, ok := expected[key]; !ok {
			mismatches = append(mismatches, MismatchEntry{Type: "extra", Key: key, Actual: actVal})
		}
	}
	return mismatches
}

func main() {
	var parquetPath string
	var xlsxPath string
	var metricFilter string
	flag.StringVar(&parquetPath, "parquet", "", "Path to parquet file (required)")
	flag.StringVar(&xlsxPath, "xlsx", "", "Path to XLSX file (required)")
	flag.StringVar(&metricFilter, "metric", "", "Only verify rows for the specified metric name (optional)")
	flag.Parse()

	if parquetPath == "" || xlsxPath == "" {
		flag.Usage()
		os.Exit(2)
	}

	// 1. Auto-discover schema.json from XLSX path
	schemaPath := filepath.Join(filepath.Dir(xlsxPath), "expected_output_schema.json")
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading schema: %v\n", err)
		os.Exit(2)
	}
	var meta gen.SchemaMeta
	if err := json.Unmarshal(schemaData, &meta); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing schema: %v\n", err)
		os.Exit(2)
	}

	// 2. Open XLSX and read expected_output sheet
	f, err := excelize.OpenFile(xlsxPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening XLSX: %v\n", err)
		os.Exit(2)
	}
	defer f.Close()

	rows, err := f.GetRows("expected_output")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: sheet 'expected_output' not found: %v\n", err)
		os.Exit(2)
	}

	// 3. Build expected map from XLSX static values
	expected := make(map[string]float64)
	metricCol := meta.MetricColumn
	valueCol := meta.ValueColumn

	for r := 1; r < len(rows); r++ { // skip header (row 0)
		row := rows[r]
		if len(row) <= valueCol {
			continue
		}

		// Apply metric filter if specified
		if metricFilter != "" {
			if metricCol < len(row) && row[metricCol] != metricFilter {
				continue
			}
		}

		// Read key values from columns 0..len(keyCols)-1
		parts := make([]string, 0, len(meta.KeyColumns)+1)
		for i := range meta.KeyColumns {
			if i < len(row) {
				parts = append(parts, row[i])
			} else {
				parts = append(parts, "")
			}
		}
		// Append metric name
		if metricCol < len(row) {
			parts = append(parts, row[metricCol])
		} else {
			parts = append(parts, "")
		}
		key := buildCompositeKey(parts)

		// Read value (static float cell)
		if valueCol < len(row) {
			val, err := strconv.ParseFloat(row[valueCol], 64)
			if err == nil {
				expected[key] = val
			}
		}
	}

	// 4. Read parquet and build actual map
	parquetSchema, numRows, err := cplio.ReadParquetFile(parquetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading parquet: %v\n", err)
		os.Exit(2)
	}

	numCols := parquetSchema.NumColumns()
	type colData struct {
		name    string
		strVals []string
		f64Vals []float64
	}
	columns := make([]colData, numCols)
	for i := 0; i < numCols; i++ {
		col := parquetSchema.Column(i)
		columns[i].name = col.Name()
		data, err := cplio.ReadParquetColumnData(parquetPath, i)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading parquet column %q: %v\n", col.Name(), err)
			os.Exit(2)
		}
		switch v := data.(type) {
		case []string:
			columns[i].strVals = v
		case []float64:
			columns[i].f64Vals = v
		}
	}

	if numRows == 0 {
		fmt.Fprintf(os.Stderr, "Error: parquet file has zero rows\n")
		os.Exit(2)
	}

	actual := make(map[string]float64)
	for rowIdx := 0; rowIdx < numRows; rowIdx++ {
		keyValuesForRow := make(map[string]string)
		var metricName string
		var value float64

		for _, col := range columns {
			if col.name == "Metric" {
				if rowIdx < len(col.strVals) {
					metricName = col.strVals[rowIdx]
				}
			} else if col.name == "Value" {
				if rowIdx < len(col.f64Vals) {
					value = col.f64Vals[rowIdx]
				}
			} else {
				if rowIdx < len(col.strVals) {
					keyValuesForRow[col.name] = col.strVals[rowIdx]
				}
			}
		}

		parts := make([]string, 0, len(meta.KeyColumns)+1)
		for _, kc := range meta.KeyColumns {
			parts = append(parts, keyValuesForRow[kc])
		}
		parts = append(parts, metricName)
		key := buildCompositeKey(parts)
		actual[key] = value
	}

	// 5. Compare, report, exit
	mismatches := findMismatches(expected, actual, 1e-9)
	if len(mismatches) > 0 {
		for _, m := range mismatches {
			switch m.Type {
			case "missing":
				fmt.Printf("Missing: %s (expected %v)\n", m.Key, m.Expected)
			case "extra":
				fmt.Printf("Extra: %s (actual %v)\n", m.Key, m.Actual)
			case "value_mismatch":
				fmt.Printf("Value mismatch: %s (expected %v, got %v)\n", m.Key, m.Expected, m.Actual)
			}
		}
		os.Exit(1)
	}

	os.Exit(0)
}
