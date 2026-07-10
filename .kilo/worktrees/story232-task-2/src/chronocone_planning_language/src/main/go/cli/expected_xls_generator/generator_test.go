// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package expected_xls_generator

import (
	"strconv"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/io"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
	"github.com/xuri/excelize/v2"
)

func TestGenerateXLSX_Header(t *testing.T) {
	defs := &analysis.AnalysisDefinitions{
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"m1": {
				Name: "m1",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "data_a"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode: &cil.CILInputRef{Name: "data_a"},
				},
			},
		},
	}
	iospec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{ConnectorName: "data_a"},
		},
		Outputs: []io.OutputSpecification{
			{Dimensions: []string{"dim1"}, Grain: []string{"grain1"}},
		},
	}
	inputData := map[string][]map[string]interface{}{
		"data_a": {
			{"dim1": "A", "grain1": "X", "Value": 10.0},
		},
	}

	f, _, err := GenerateXLSX(defs, iospec, inputData, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	// Verify the header row
	rows, err := f.GetRows("expected_output")
	if err != nil {
		t.Fatalf("failed to get rows: %v", err)
	}
	if len(rows) < 1 {
		t.Fatalf("expected at least header row, got %d rows", len(rows))
	}

	header := rows[0]
	expectedHeaders := []string{"dim1", "grain1", "data_a", "Metric", "Value", "Computed Value", "Match"}
	if len(header) != len(expectedHeaders) {
		t.Fatalf("expected %d header cells, got %d", len(expectedHeaders), len(header))
	}
	for i, expected := range expectedHeaders {
		if header[i] != expected {
			t.Errorf("header[%d]: expected %q, got %q", i, expected, header[i])
		}
	}
}

func TestGenerateXLSX_SingleRowPerMetricCombo(t *testing.T) {
	defs := &analysis.AnalysisDefinitions{
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"m1": {
				Name: "m1",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "data_a"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode: &cil.CILInputRef{Name: "data_a"},
				},
			},
		},
	}
	iospec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{ConnectorName: "data_a"},
		},
		Outputs: []io.OutputSpecification{
			{Dimensions: []string{"dim1"}, Grain: []string{"grain1"}},
		},
	}
	inputData := map[string][]map[string]interface{}{
		"data_a": {
			{"dim1": "A", "grain1": "X", "Value": 10.0},
			{"dim1": "B", "grain1": "Y", "Value": 20.0},
		},
	}

	f, _, err := GenerateXLSX(defs, iospec, inputData, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	rows, _ := f.GetRows("expected_output")
	// Header + 2 data rows
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows (1 header + 2 data), got %d", len(rows))
	}
}

func TestGenerateXLSX_ConditionalSingleRow(t *testing.T) {
	defs := &analysis.AnalysisDefinitions{
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"m1": {
				Name: "m1",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "data_a"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode: &cil.CILConditional{
						Branches: []cil.CILConditionalBranch{
							{Condition: &cil.CILComparisonOp{Op: operator.Equal, Left: &cil.CILInputRef{Name: "data_a"}, Right: &cil.CILLiteral{Value: 10}}, Body: &cil.CILLiteral{Value: 100}},
							{Condition: nil, Body: &cil.CILLiteral{Value: 0}},
						},
					},
				},
			},
		},
	}
	iospec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{ConnectorName: "data_a"},
		},
		Outputs: []io.OutputSpecification{
			{Dimensions: []string{"dim1"}, Grain: []string{"grain1"}},
		},
	}
	inputData := map[string][]map[string]interface{}{
		"data_a": {
			{"dim1": "A", "grain1": "X", "Value": 10.0},
		},
	}

	f, _, err := GenerateXLSX(defs, iospec, inputData, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	rows, _ := f.GetRows("expected_output")
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (1 header + 1 data), got %d", len(rows))
	}
}

func TestGenerateXLSX_InputColumnsAsFormulas(t *testing.T) {
	defs := &analysis.AnalysisDefinitions{
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"m1": {
				Name: "m1",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "data_a"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode: &cil.CILInputRef{Name: "data_a"},
				},
			},
		},
	}
	iospec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{ConnectorName: "data_a"},
		},
		Outputs: []io.OutputSpecification{
			{Dimensions: []string{"dim1"}, Grain: []string{"grain1"}},
		},
	}
	inputData := map[string][]map[string]interface{}{
		"data_a": {
			{"dim1": "A", "grain1": "X", "Value": 10.0},
		},
	}

	f, _, err := GenerateXLSX(defs, iospec, inputData, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	// input column "data_a" is at column C (col index 2, 0-based) in the output sheet
	formula, err := f.GetCellFormula("expected_output", "C2")
	if err != nil {
		t.Fatalf("failed to get formula: %v", err)
	}
	if formula == "" {
		t.Errorf("expected input column to have formula, got empty")
	}
	if formula != "data_a!C2" {
		t.Errorf("expected input formula %q, got %q", "data_a!C2", formula)
	}
}

func TestGenerateXLSX_ValueStatic(t *testing.T) {
	defs := &analysis.AnalysisDefinitions{
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"m1": {
				Name: "m1",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "data_a"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode: &cil.CILBinaryOp{Op: operator.Add, Left: &cil.CILInputRef{Name: "data_a"}, Right: &cil.CILLiteral{Value: 5}},
				},
			},
		},
	}
	iospec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{ConnectorName: "data_a"},
		},
		Outputs: []io.OutputSpecification{
			{Dimensions: []string{"dim1"}, Grain: []string{"grain1"}},
		},
	}
	inputData := map[string][]map[string]interface{}{
		"data_a": {
			{"dim1": "A", "grain1": "X", "Value": 10.0},
		},
	}

	f, _, err := GenerateXLSX(defs, iospec, inputData, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	// Value is at column E (index 4) — it should have no formula
	formula, err := f.GetCellFormula("expected_output", "E2")
	if err != nil {
		t.Fatalf("failed to get formula: %v", err)
	}
	if formula != "" {
		t.Errorf("expected Value column to be static, got formula %q", formula)
	}

	val, _ := f.GetCellValue("expected_output", "E2")
	expected, _ := strconv.ParseFloat(val, 64)
	if expected != 15 {
		t.Errorf("expected Value 15, got %v", expected)
	}
}

func TestGenerateXLSX_ComputedValueFormula(t *testing.T) {
	defs := &analysis.AnalysisDefinitions{
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"m1": {
				Name: "m1",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "data_a"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode: &cil.CILBinaryOp{Op: operator.Add, Left: &cil.CILInputRef{Name: "data_a"}, Right: &cil.CILLiteral{Value: 5}},
				},
			},
		},
	}
	iospec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{ConnectorName: "data_a"},
		},
		Outputs: []io.OutputSpecification{
			{Dimensions: []string{"dim1"}, Grain: []string{"grain1"}},
		},
	}
	inputData := map[string][]map[string]interface{}{
		"data_a": {
			{"dim1": "A", "grain1": "X", "Value": 10.0},
		},
	}

	f, _, err := GenerateXLSX(defs, iospec, inputData, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	// Computed Value is at column F (index 5)
	formula, err := f.GetCellFormula("expected_output", "F2")
	if err != nil {
		t.Fatalf("failed to get formula: %v", err)
	}
	if formula == "" {
		t.Errorf("expected Computed Value column to have formula, got empty")
	}
	expectedFormula := "(C2+5)"
	if formula != expectedFormula {
		t.Errorf("expected computed formula %q, got %q", expectedFormula, formula)
	}
}

func TestGenerateXLSX_MatchFormula(t *testing.T) {
	defs := &analysis.AnalysisDefinitions{
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"m1": {
				Name: "m1",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "data_a"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode: &cil.CILInputRef{Name: "data_a"},
				},
			},
		},
	}
	iospec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{ConnectorName: "data_a"},
		},
		Outputs: []io.OutputSpecification{
			{Dimensions: []string{"dim1"}, Grain: []string{"grain1"}},
		},
	}
	inputData := map[string][]map[string]interface{}{
		"data_a": {
			{"dim1": "A", "grain1": "X", "Value": 10.0},
		},
	}

	f, _, err := GenerateXLSX(defs, iospec, inputData, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	// Match is at column G (index 6)
	formula, err := f.GetCellFormula("expected_output", "G2")
	if err != nil {
		t.Fatalf("failed to get formula: %v", err)
	}
	if formula == "" {
		t.Errorf("expected Match column to have formula, got empty")
	}
	expectedMatch := `IF(ABS(E2-F2)<0.000000001,"OK","Not A Match")`
	if formula != expectedMatch {
		t.Errorf("expected match formula %q, got %q", expectedMatch, formula)
	}
}

func TestGenerateXLSX_UsesExcelCellReferences(t *testing.T) {
	// Verify excelize utility functions
	if ColName(0) != "A" {
		t.Errorf("expected ColName(0)=A, got %q", ColName(0))
	}
	if ColName(25) != "Z" {
		t.Errorf("expected ColName(25)=Z, got %q", ColName(25))
	}
	if ColName(26) != "AA" {
		t.Errorf("expected ColName(26)=AA, got %q", ColName(26))
	}

	if CellRef(0, 1) != "A1" {
		t.Errorf("expected CellRef(0,1)=A1, got %q", CellRef(0, 1))
	}
	if CellRef(3, 5) != "D5" {
		t.Errorf("expected CellRef(3,5)=D5, got %q", CellRef(3, 5))
	}
}

func TestGenerateXLSX_MultipleInputs(t *testing.T) {
	defs := &analysis.AnalysisDefinitions{
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"m1": {
				Name: "m1",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "data_a"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode: &cil.CILInputRef{Name: "data_a"},
				},
			},
		},
	}
	iospec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{ConnectorName: "data_a"},
			{ConnectorName: "data_b"},
		},
		Outputs: []io.OutputSpecification{
			{Dimensions: []string{"dim1"}, Grain: []string{"grain1"}},
		},
	}
	inputData := map[string][]map[string]interface{}{
		"data_a": {
			{"dim1": "A", "grain1": "X", "Value": 10.0},
		},
		"data_b": {
			{"dim1": "B", "grain1": "Y", "Value": 20.0},
		},
	}

	f, _, err := GenerateXLSX(defs, iospec, inputData, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	// Verify both input sheets exist
	sheets := f.GetSheetList()
	foundA, foundB := false, false
	for _, s := range sheets {
		if s == "data_a" {
			foundA = true
		}
		if s == "data_b" {
			foundB = true
		}
	}
	if !foundA {
		t.Errorf("expected sheet data_a to exist")
	}
	if !foundB {
		t.Errorf("expected sheet data_b to exist")
	}

	// Verify input data in data_a sheet. The keys are sorted alphabetically:
	// "Value", "dim1", "grain1" — so Value is in column A, dim1 in B, grain1 in C
	// Cell A2 = "10" (Value), B2 = "A" (dim1), C2 = "X" (grain1)
	rowsA, _ := f.GetRows("data_a")
	if len(rowsA) < 2 {
		t.Fatalf("expected at least 2 rows in data_a, got %d", len(rowsA))
	}
	// Header row: Value, dim1, grain1
	if rowsA[0][0] != "Value" {
		t.Errorf("expected data_a header col A to be 'Value', got %q", rowsA[0][0])
	}
	// Data row: 10, A, X
	parsed, _ := strconv.ParseFloat(rowsA[1][0], 64)
	if parsed != 10.0 {
		t.Errorf("expected data_a row 1 col A=10, got %v", parsed)
	}

	// Verify data_b has input data
	rowsB, _ := f.GetRows("data_b")
	if len(rowsB) < 2 {
		t.Fatalf("expected at least 2 rows in data_b, got %d", len(rowsB))
	}
	parsed, _ = strconv.ParseFloat(rowsB[1][0], 64)
	if parsed != 20.0 {
		t.Errorf("expected data_b row 1 col A=20, got %v", parsed)
	}
}

func TestGenerateXLSX_FormulaComputedExpr(t *testing.T) {
	defs := &analysis.AnalysisDefinitions{
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"m1": {
				Name: "m1",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "data_a"},
					{Kind: "metric", Name: "data_b"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode: &cil.CILConditional{
						Branches: []cil.CILConditionalBranch{
							{
								Condition: &cil.CILComparisonOp{
									Op: operator.GreaterThan,
									Left:  &cil.CILInputRef{Name: "data_a"},
									Right: &cil.CILLiteral{Value: 5},
								},
								Body: &cil.CILBinaryOp{
									Op: operator.Add,
									Left:  &cil.CILInputRef{Name: "data_a"},
									Right: &cil.CILInputRef{Name: "data_b"},
								},
							},
							{
								Condition: nil,
								Body:      &cil.CILLiteral{Value: 0},
							},
						},
					},
				},
			},
		},
	}
	iospec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{ConnectorName: "data_a"},
			{ConnectorName: "data_b"},
		},
		Outputs: []io.OutputSpecification{
			{Dimensions: []string{"dim1"}, Grain: []string{"grain1"}},
		},
	}
	inputData := map[string][]map[string]interface{}{
		"data_a": {
			{"dim1": "A", "grain1": "X", "Value": 10.0},
		},
		"data_b": {
			{"dim1": "A", "grain1": "X", "Value": 20.0},
		},
	}

	f, _, err := GenerateXLSX(defs, iospec, inputData, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	// With 2 input columns (data_a references data_a!C2, data_b references data_b!C2), the layout is:
	// A: dim1, B: grain1, C: data_a, D: data_b, E: Metric, F: Value, G: Computed Value, H: Match
	// Value column (F2) should be static 30.0
	val, _ := f.GetCellValue("expected_output", "F2")
	parsed, _ := strconv.ParseFloat(val, 64)
	if parsed != 30.0 {
		t.Errorf("expected Value 30, got %v", parsed)
	}

	// Computed Value (G2) should be formula IF(C2>5,(C2+D2),0)
	formula, err := f.GetCellFormula("expected_output", "G2")
	if err != nil {
		t.Fatalf("failed to get formula: %v", err)
	}
	expected := "IF(C2>5,(C2+D2),0)"
	if formula != expected {
		t.Errorf("expected formula %q, got %q", expected, formula)
	}
}

// TestGenerateXLSX_InputColumnsReferenceCorrectSheets verifies that each
// input column formula references its own data sheet, not always the first.
func TestGenerateXLSX_InputColumnsReferenceCorrectSheets(t *testing.T) {
	defs := &analysis.AnalysisDefinitions{
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"m1": {
				Name: "m1",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "data_a"},
					{Kind: "metric", Name: "data_b"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode: &cil.CILBinaryOp{
						Op:    operator.Add,
						Left:  &cil.CILInputRef{Name: "data_a"},
						Right: &cil.CILInputRef{Name: "data_b"},
					},
				},
			},
		},
	}
	iospec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{ConnectorName: "data_a"},
			{ConnectorName: "data_b"},
		},
		Outputs: []io.OutputSpecification{
			{Dimensions: []string{"dim1"}, Grain: []string{"grain1"}},
		},
	}
	inputData := map[string][]map[string]interface{}{
		"data_a": {
			{"dim1": "A", "grain1": "X", "Value": 10.0},
		},
		"data_b": {
			{"dim1": "A", "grain1": "X", "Value": 20.0},
		},
	}

	f, _, err := GenerateXLSX(defs, iospec, inputData, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	// data_a input column (C2) should reference data_a!C2
	formulaA, err := f.GetCellFormula("expected_output", "C2")
	if err != nil {
		t.Fatalf("failed to get formula for C2: %v", err)
	}
	if formulaA != "data_a!C2" {
		t.Errorf("expected data_a column formula 'data_a!C2', got %q", formulaA)
	}

	// data_b input column (D2) should reference data_b!C2, NOT data_a!C2
	formulaB, err := f.GetCellFormula("expected_output", "D2")
	if err != nil {
		t.Fatalf("failed to get formula for D2: %v", err)
	}
	if formulaB != "data_b!C2" {
		t.Errorf("expected data_b column formula 'data_b!C2', got %q", formulaB)
	}
}

// TestWriteToBuffer verifies in-memory generation via WriteToBuffer
func TestWriteToBuffer(t *testing.T) {
	f := excelize.NewFile()
	f.SetCellValue("Sheet1", "A1", "test")
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("unexpected error from WriteToBuffer: %v", err)
	}
	if buf.Len() == 0 {
		t.Errorf("expected non-empty buffer from WriteToBuffer")
	}
	f.Close()
}

func TestGenerateXLSX_LookupSheets(t *testing.T) {
	defs := &analysis.AnalysisDefinitions{
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"m1": {
				Name: "m1",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "data_a"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode: &cil.CILInputRef{Name: "data_a"},
				},
			},
		},
	}
	iospec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{ConnectorName: "data_a"},
		},
		Outputs: []io.OutputSpecification{
			{Dimensions: []string{"dim1"}, Grain: []string{"grain1"}},
		},
		Lookups: []io.LookupSpecification{
			{LookupName: "strength_adjusted_for_material"},
		},
	}
	inputData := map[string][]map[string]interface{}{
		"data_a": {
			{"dim1": "A", "grain1": "X", "Value": 10.0},
		},
	}
	lookupData := map[string][]map[string]interface{}{
		"strength_adjusted_for_material": {
			{"Material": "Steel", "Value": 1.5},
			{"Material": "Aluminum", "Value": 0.8},
		},
	}

	f, _, err := GenerateXLSX(defs, iospec, inputData, lookupData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	// Verify the lookup sheet exists
	sheets := f.GetSheetList()
	found := false
	for _, s := range sheets {
		if s == "strength_adjusted_for_material" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected sheet 'strength_adjusted_for_material' to exist")
	}

	// Verify header row includes dimension column and Value column
	rows, err := f.GetRows("strength_adjusted_for_material")
	if err != nil {
		t.Fatalf("failed to get rows for lookup sheet: %v", err)
	}
	if len(rows) < 3 {
		t.Fatalf("expected at least 3 rows (header + 2 data), got %d", len(rows))
	}

	// Verify header row: key, Material, Value
	header := rows[0]
	foundKey := false
	foundMaterial := false
	foundValue := false
	for _, h := range header {
		if h == "key" {
			foundKey = true
		}
		if h == "Material" {
			foundMaterial = true
		}
		if h == "Value" {
			foundValue = true
		}
	}
	if !foundKey {
		t.Errorf("expected 'key' in header, got %v", header)
	}
	if !foundMaterial {
		t.Errorf("expected 'Material' in header, got %v", header)
	}
	if !foundValue {
		t.Errorf("expected 'Value' in header, got %v", header)
	}

	// Verify data rows:
	// Column order: key, Material, Value
	// Input order: row 1 = Steel/1.5, row 2 = Aluminum/0.8
	// Single-dim key = dimension value.
	if rows[1][0] != "Steel" {
		t.Errorf("expected row 1 key='Steel', got %q", rows[1][0])
	}
	if rows[2][0] != "Aluminum" {
		t.Errorf("expected row 2 key='Aluminum', got %q", rows[2][0])
	}
	if rows[1][1] != "Steel" {
		t.Errorf("expected row 1 Material='Steel', got %q", rows[1][1])
	}
	if rows[2][1] != "Aluminum" {
		t.Errorf("expected row 2 Material='Aluminum', got %q", rows[2][1])
	}
	parsed, _ := strconv.ParseFloat(rows[1][2], 64)
	if parsed != 1.5 {
		t.Errorf("expected row 1 Value=1.5, got %v", parsed)
	}
	parsed, _ = strconv.ParseFloat(rows[2][2], 64)
	if parsed != 0.8 {
		t.Errorf("expected row 2 Value=0.8, got %v", parsed)
	}
}

func TestGenerateXLSX_LookupFormulaInOutput(t *testing.T) {
	defs := &analysis.AnalysisDefinitions{
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"product_strength": {
				Name: "product_strength",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "raw_strength"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode: &cil.CILBinaryOp{
						Op:   operator.Multiply,
						Left: &cil.CILInputRef{Name: "raw_strength"},
						Right: &cil.CILLookupNode{
							LookupName: "strength_adjusted_for_material",
							Dimensions: []string{"Material"},
						},
					},
				},
			},
		},
	}
	iospec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{ConnectorName: "raw_strength"},
		},
		Outputs: []io.OutputSpecification{
			{Dimensions: []string{"Material"}, Grain: []string{}},
		},
		Lookups: []io.LookupSpecification{
			{LookupName: "strength_adjusted_for_material"},
		},
	}
	inputData := map[string][]map[string]interface{}{
		"raw_strength": {
			{"Material": "Steel", "Value": 100.0},
		},
	}
	lookupData := map[string][]map[string]interface{}{
		"strength_adjusted_for_material": {
			{"Material": "Steel", "Value": 1.5},
		},
	}

	f, _, err := GenerateXLSX(defs, iospec, inputData, lookupData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	// Verify the lookup sheet exists
	sheets := f.GetSheetList()
	foundLookup := false
	for _, s := range sheets {
		if s == "strength_adjusted_for_material" {
			foundLookup = true
			break
		}
	}
	if !foundLookup {
		t.Fatalf("expected lookup sheet 'strength_adjusted_for_material' to exist")
	}

	// The expected_output sheet layout:
	// Column A (0): Material (key column)
	// Column B (1): raw_strength (input column)
	// Column C (2): Metric
	// Column D (3): Value
	// Column E (4): Computed Value
	// Column F (5): Match
	//
	// For the formula: raw_strength * VLOOKUP(Material, strength_adjusted_for_material!A:C,3,FALSE)
	// Material is in column A, raw_strength is in column B
	// Formula: (B2*VLOOKUP(A2,strength_adjusted_for_material!A:C,3,FALSE))

	// Get Computed Value column formula
	computedCol := ColName(4) // column E
	cellRef := computedCol + "2"
	formula, err := f.GetCellFormula("expected_output", cellRef)
	if err != nil {
		t.Fatalf("failed to get formula for %s: %v", cellRef, err)
	}
	if formula == "" {
		t.Errorf("expected Computed Value formula to be non-empty")
	}

	expectedFormula := "(B2*VLOOKUP(A2,strength_adjusted_for_material!A:C,3,FALSE))"
	if formula != expectedFormula {
		t.Errorf("expected formula %q, got %q", expectedFormula, formula)
	}
}
