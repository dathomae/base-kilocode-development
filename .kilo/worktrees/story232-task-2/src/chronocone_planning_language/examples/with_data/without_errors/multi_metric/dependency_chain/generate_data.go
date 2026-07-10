package main

import (
	"fmt"
	"os"

	cplio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
	"github.com/xuri/excelize/v2"
)

func main() {
	// Generate input parquet: 2 Products × 2 Regions = 4 cells
	dimensions := map[string][]string{
		"Region":  {"North", "South"},
		"Product": {"Widget", "Gadget"},
	}
	grain := []cplio.GrainMeasure{
		{Name: "SingleDay", Values: []string{"Monday"}},
	}
	// Values in Cartesian product order (Product varies slowest? No —
	// dimension columns are alphabetical: Product, then Region)
	// Product=Widget, Region=North  → 100
	// Product=Widget, Region=South  → 200
	// Product=Gadget, Region=North  → 300
	// Product=Gadget, Region=South  → 400
	values := []float64{100, 200, 300, 400}

	err := cplio.GenerateParquetFile("revenue_data.parquet", dimensions, grain, values)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating parquet: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated revenue_data.parquet")

	// Generate expected_output.xlsx
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", "expected_output")

	// Header row
	headers := []string{"Product", "Region", "SingleDay", "Metric", "Value", "Computed Value", "Match"}
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellStr("expected_output", cell, h)
	}

	// Data rows: one per (Product, Region, Daily, Metric) combination
	// We have 4 combos × 2 output metrics = 8 rows
	combos := []struct {
		product string
		region  string
		daily   string
	}{
		{"Widget", "North", "Monday"},
		{"Widget", "South", "Monday"},
		{"Gadget", "North", "Monday"},
		{"Gadget", "South", "Monday"},
	}

	// Expected values:
	// base_revenue = revenue_data (identity)
	// total = base_revenue + tax = base_revenue + base_revenue / 10
	baseValues := []float64{100, 200, 300, 400}
	totalValues := []float64{110, 220, 330, 440}

	row := 2
	for i, combo := range combos {
		// base_revenue row
		writeRow(f, row, combo.product, combo.region, combo.daily, "base_revenue", baseValues[i])
		row++
		// total row
		writeRow(f, row, combo.product, combo.region, combo.daily, "total", totalValues[i])
		row++
	}

	if err := f.SaveAs("expected_output.xlsx"); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving xlsx: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated expected_output.xlsx")
}

func writeRow(f *excelize.File, row int, product, region, daily, metric string, value float64) {
	f.SetCellStr("expected_output", cellRef(0, row), product)
	f.SetCellStr("expected_output", cellRef(1, row), region)
	f.SetCellStr("expected_output", cellRef(2, row), daily)
	f.SetCellStr("expected_output", cellRef(3, row), metric)
	f.SetCellFloat("expected_output", cellRef(4, row), value, -1, 64)
}

func cellRef(col, row int) string {
	ref, _ := excelize.CoordinatesToCellName(col+1, row)
	return ref
}
