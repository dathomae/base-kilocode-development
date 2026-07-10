// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package expected_xls_generator

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/io"
	"github.com/xuri/excelize/v2"
)

// GenerateXLSX generates the XLSX spreadsheet with input data sheets and an expected_output sheet.
// Returns the *excelize.File, sorted list of input column names, and any error.
func GenerateXLSX(defs *analysis.AnalysisDefinitions, iospec *io.IOSpecification, inputData map[string][]map[string]interface{}, lookupData map[string][]map[string]interface{}) (*excelize.File, []string, error) {
	f := excelize.NewFile()

	// Step 1: Input sheets
	for i, inputSpec := range iospec.Connectors {
		sheetName := inputSpec.ConnectorName
		if i == 0 {
			// Rename default "Sheet1" to the first connector name
			if err := f.SetSheetName("Sheet1", sheetName); err != nil {
				return nil, nil, fmt.Errorf("failed to rename Sheet1 to %q: %w", sheetName, err)
			}
		} else {
			_, err := f.NewSheet(sheetName)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create sheet %q: %w", sheetName, err)
			}
		}
		populateInputSheetXLSX(f, sheetName, inputData[inputSpec.ConnectorName])
	}

	// Step 1b: Lookup data sheets
	// Each lookup sheet uses a concatenated key column (column A) built from
	// the dimension values of each row, pipe-separated for multi-dim lookups.
	// Column order is: [key] + [dimensions sorted] + [Value].
	for _, lookupSpec := range iospec.Lookups {
		sheetName := lookupSpec.LookupName
		rows, ok := lookupData[lookupSpec.LookupName]
		if !ok || len(rows) == 0 {
			continue
		}
		_, err := f.NewSheet(sheetName)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create sheet %q: %w", sheetName, err)
		}
		populateLookupSheetXLSX(f, sheetName, rows)
	}

	// Step 2: Collect dimension/grain combinations
	var keyColumns []string
	if len(iospec.Outputs) > 0 {
		keyColumns = append(keyColumns, iospec.Outputs[0].Dimensions...)
		keyColumns = append(keyColumns, iospec.Outputs[0].Grain...)
	}

	dimensionColumnIndexes := make(map[string]int)
	for i, kc := range keyColumns {
		dimensionColumnIndexes[kc] = i
	}

	combos := make([]map[string]string, 0)
	seen := make(map[string]bool)

	for _, inputSpec := range iospec.Connectors {
		rows := inputData[inputSpec.ConnectorName]
		for _, row := range rows {
			combo := make(map[string]string)
			for _, kc := range keyColumns {
				if v, ok := row[kc].(string); ok {
					combo[kc] = v
				}
			}
			key := comboKey(combo, keyColumns)
			if !seen[key] {
				seen[key] = true
				combos = append(combos, combo)
			}
		}
	}

	// Step 3: expected_output sheet
	_, err := f.NewSheet("expected_output")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create expected_output sheet: %w", err)
	}

	if len(iospec.Connectors) == 0 {
		return f, nil, nil
	}
	primaryConnectorSheet := iospec.Connectors[0].ConnectorName
	keyColCount := len(keyColumns)

	// Collect input column names from all metric CIL trees
	inputColumns := collectInputNames(defs)

	// Build map from input column name to its 0-based index within the input columns block
	inputColumnIndexes := make(map[string]int)
	for i, name := range inputColumns {
		inputColumnIndexes[name] = i
	}

	// Build map from input column name to its connector sheet
	// Each iospec input maps to a metric input by position:
	// iospec.Connectors[i].ConnectorName → metricDef.Inputs[i].Name
	inputColumnToSheet := make(map[string]string)
	for _, metricDef := range defs.Metrics {
		for i, input := range metricDef.Inputs {
			if i < len(iospec.Connectors) {
				inputColumnToSheet[input.Name] = iospec.Connectors[i].ConnectorName
			}
		}
	}

	// Write header row: key columns → input columns → Metric, Value, Computed Value, Match
	headerRow := 1
	col := 0
	for _, kc := range keyColumns {
		cell := CellRef(col, headerRow)
		f.SetCellStr("expected_output", cell, kc)
		col++
	}
	for _, ic := range inputColumns {
		cell := CellRef(col, headerRow)
		f.SetCellStr("expected_output", cell, ic)
		col++
	}
	f.SetCellStr("expected_output", CellRef(col, headerRow), "Metric"); col++
	f.SetCellStr("expected_output", CellRef(col, headerRow), "Value"); col++
	f.SetCellStr("expected_output", CellRef(col, headerRow), "Computed Value"); col++
	f.SetCellStr("expected_output", CellRef(col, headerRow), "Match")

	// Sort metric names for deterministic output
	var metricNames []string
	for name := range defs.Metrics {
		metricNames = append(metricNames, name)
	}
	sort.Strings(metricNames)

	currentRow := 2 // data starts at row 2 (1-indexed)

	for _, metricName := range metricNames {
		metricDef := defs.Metrics[metricName]
		if metricDef.CILRoot == nil || metricDef.CILRoot.CILNode == nil {
			continue
		}
		for _, combo := range combos {
			rowIdx := findRowIndex(inputData[primaryConnectorSheet], combo, keyColumns)
			if rowIdx < 0 {
				continue
			}

			valueLookup := buildValueLookup(&metricDef, iospec, inputData, combo)
			ctx := &EvalContext{
				ValueLookup:     valueLookup,
				LookupData:      lookupData,
				DimensionValues: combo,
			}
			result, err := EvaluateExpressionWithContext(metricDef.CILRoot.CILNode, ctx)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to evaluate metric %s: %w", metricDef.Name, err)
			}

			// Build input column formulas, each referencing its own connector sheet
			inputFormulas := make([]string, len(inputColumns))
			for i, ic := range inputColumns {
				connectorSheet := inputColumnToSheet[ic]
				if connectorSheet == "" {
					connectorSheet = primaryConnectorSheet // fallback
				}
				// Find the matching row in this input column's own sheet
				colRowIdx := findRowIndex(inputData[connectorSheet], combo, keyColumns)
				if colRowIdx < 0 {
					colRowIdx = rowIdx // fallback to primary row
				}
				inputFormulas[i] = ExpressionToExcel(&cil.CILInputRef{Name: ic}, connectorSheet, colRowIdx, keyColCount, ColName, dimensionColumnIndexes)
			}

			// Computed value formula using same-sheet cell references
			computedFormula := ExpressionToCellExcel(metricDef.CILRoot.CILNode, inputColumnIndexes, ColName, currentRow, keyColCount, dimensionColumnIndexes)

			// Match formula
			valueCol := keyColCount + len(inputColumns) + 1
			computedCol := valueCol + 1
			valueCell := ColName(valueCol)
			computedCell := ColName(computedCol)
			matchFormula := "IF(ABS(" + valueCell + strconv.Itoa(currentRow) + "-" + computedCell + strconv.Itoa(currentRow) + ")<0.000000001,\"OK\",\"Not A Match\")"

			// Write the row
			writeOutputRowXLSX(f, "expected_output", currentRow, keyColumns, inputColumns, combo, inputFormulas, metricDef.Name, strconv.FormatFloat(result, 'f', -1, 64), computedFormula, matchFormula)
			currentRow++
		}
	}

	return f, inputColumns, nil
}

func collectInputNames(defs *analysis.AnalysisDefinitions) []string {
	seen := make(map[string]bool)
	for _, metricDef := range defs.Metrics {
		for _, input := range metricDef.Inputs {
			if input.Kind == "metric" {
				seen[input.Name] = true
			}
		}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func populateInputSheetXLSX(f *excelize.File, sheetName string, rows []map[string]interface{}) {
	if len(rows) == 0 {
		return
	}
	keys := getSortedKeys(rows[0])

	// Write header row (row 1)
	for i, k := range keys {
		cell := CellRef(i, 1)
		f.SetCellStr(sheetName, cell, k)
	}

	// Write data rows
	for rowIdx, row := range rows {
		r := rowIdx + 2 // data rows start at row 2
		for i, k := range keys {
			cell := CellRef(i, r)
			v := row[k]
			switch val := v.(type) {
			case float64:
				f.SetCellFloat(sheetName, cell, val, -1, 64)
			case string:
				f.SetCellStr(sheetName, cell, val)
			default:
				f.SetCellStr(sheetName, cell, fmt.Sprintf("%v", val))
			}
		}
	}
}

func getSortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// populateLookupSheetXLSX writes lookup data to a new Excel sheet with a
// concatenated key column (column A) followed by dimension columns and Value.
// For single-dimension lookups the key is the dimension value.
// For multi-dimension lookups the key is pipe-separated dimension values.
// Column order: key, <dimension columns sorted>, Value
func populateLookupSheetXLSX(f *excelize.File, sheetName string, rows []map[string]interface{}) {
	if len(rows) == 0 {
		return
	}

	// Derive dimension names from the first row (exclude "Value")
	dimensionNames := make([]string, 0)
	for k := range rows[0] {
		if k != "Value" {
			dimensionNames = append(dimensionNames, k)
		}
	}
	sort.Strings(dimensionNames)

	// Header order: key, then dimensions (alphabetical), then Value
	header := make([]string, 0, len(dimensionNames)+2)
	header = append(header, "key")
	header = append(header, dimensionNames...)
	header = append(header, "Value")

	// Write header row (row 1)
	for i, h := range header {
		cell := CellRef(i, 1)
		f.SetCellStr(sheetName, cell, h)
	}

	// Write data rows
	for rowIdx, row := range rows {
		r := rowIdx + 2 // data rows start at row 2

		// Build the concatenated key from dimension values
		keyParts := make([]string, 0, len(dimensionNames))
		for _, dim := range dimensionNames {
			keyParts = append(keyParts, fmt.Sprintf("%v", row[dim]))
		}
		key := strings.Join(keyParts, "|")

		// Write key column (A)
		f.SetCellStr(sheetName, CellRef(0, r), key)

		// Write dimension columns
		for i, dim := range dimensionNames {
			cell := CellRef(i+1, r)
			v := row[dim]
			switch val := v.(type) {
			case float64:
				f.SetCellFloat(sheetName, cell, val, -1, 64)
			case string:
				f.SetCellStr(sheetName, cell, val)
			default:
				f.SetCellStr(sheetName, cell, fmt.Sprintf("%v", val))
			}
		}

		// Write Value column (last)
		valueCol := len(dimensionNames) + 1
		v := row["Value"]
		switch val := v.(type) {
		case float64:
			f.SetCellFloat(sheetName, CellRef(valueCol, r), val, -1, 64)
		case string:
			f.SetCellStr(sheetName, CellRef(valueCol, r), val)
		default:
			f.SetCellStr(sheetName, CellRef(valueCol, r), fmt.Sprintf("%v", val))
		}
	}
}

func comboKey(combo map[string]string, keyColumns []string) string {
	key := ""
	for _, kc := range keyColumns {
		key += kc + "=" + combo[kc] + ";"
	}
	return key
}

func findRowIndex(rows []map[string]interface{}, combo map[string]string, keyColumns []string) int {
	for i, row := range rows {
		match := true
		for _, kc := range keyColumns {
			if rv, ok := row[kc].(string); !ok || rv != combo[kc] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func buildValueLookup(metricDef *analysis.AnalysisMetricDefinition, iospec *io.IOSpecification, inputData map[string][]map[string]interface{}, combo map[string]string) map[string]float64 {
	valueLookup := make(map[string]float64)
	for i, input := range metricDef.Inputs {
		if i >= len(iospec.Connectors) {
			break
		}
		connectorName := iospec.Connectors[i].ConnectorName
		rows := inputData[connectorName]
		for _, row := range rows {
			match := true
			for kc, comboVal := range combo {
				if rv, ok := row[kc].(string); !ok || rv != comboVal {
					match = false
					break
				}
			}
			if match {
				if val, ok := row["Value"].(float64); ok {
					valueLookup[input.Name] = val
				}
				break
			}
		}
	}
	return valueLookup
}

func writeOutputRowXLSX(f *excelize.File, sheetName string, row int, keyColumns []string, inputColumns []string, combo map[string]string, inputFormulas []string, metricName, value, computedFormula, matchFormula string) {
	col := 0

	// Key column values
	for _, kc := range keyColumns {
		f.SetCellStr(sheetName, CellRef(col, row), combo[kc])
		col++
	}

	// Input column formulas (cross-sheet references)
	for _, formula := range inputFormulas {
		f.SetCellFormula(sheetName, CellRef(col, row), formula)
		col++
	}

	// Metric name
	f.SetCellStr(sheetName, CellRef(col, row), metricName)
	col++

	// Value (static)
	fValue, _ := strconv.ParseFloat(value, 64)
	f.SetCellFloat(sheetName, CellRef(col, row), fValue, -1, 64)
	col++

	// Computed Value (formula)
	f.SetCellFormula(sheetName, CellRef(col, row), computedFormula)
	col++

	// Match (formula)
	f.SetCellFormula(sheetName, CellRef(col, row), matchFormula)
}
