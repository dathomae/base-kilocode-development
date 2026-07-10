// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package expected_xls_generator

import "github.com/xuri/excelize/v2"

// ColName converts 0-indexed column index to Excel column letter.
// col=0 → "A", col=25 → "Z", col=26 → "AA"
func ColName(col int) string {
	name, _ := excelize.ColumnNumberToName(col + 1)
	return name
}

// CellRef returns an Excel cell reference from 0-based column and 1-based row.
// CellRef(0, 1) → "A1", CellRef(3, 5) → "D5"
func CellRef(col int, row int) string {
	ref, _ := excelize.CoordinatesToCellName(col+1, row)
	return ref
}

// SchemaMeta represents metadata about an XLSX file schema.
type SchemaMeta struct {
	XLSXFile      string   `json:"xlsx_file"`
	KeyColumns    []string `json:"key_columns"`
	InputColumns  []string `json:"input_columns"`
	MetricColumn  int      `json:"metric_column"`
	ValueColumn   int      `json:"value_column"`
}
