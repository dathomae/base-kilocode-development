// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file defines AggregationMatrix and its builder functions, which produce
// pre-computed weight matrices used for tensor dimension reduction during aggregation.
package analysis

import (
	cil "github.com/chronocone/chronocone_planning_language/src/main/go/cil"
)

// AggregationMatrix is a 2D weight matrix used for tensor dimension reduction.
// Rows correspond to values of the concept being collapsed.
// Columns correspond to values of the concept being preserved (or 1 column for scalar).
// For sum(Country): matrix[m][n] = 1.0 for each country m → product n mapping.
type AggregationMatrix struct {
	Rows      int
	Columns   int
	Data      [][]float64
	RowLabels []string
	ColLabels []string
}

// AnalysisMatrixToCIL converts an analysis AggregationMatrix to its CIL equivalent.
func AnalysisMatrixToCIL(m *AggregationMatrix) *cil.AggregationMatrix {
	if m == nil {
		return nil
	}
	return &cil.AggregationMatrix{
		Rows:      m.Rows,
		Columns:   m.Columns,
		Data:      m.Data,
		RowLabels: m.RowLabels,
		ColLabels: m.ColLabels,
	}
}

// BuildSumAggregationMatrix creates an aggregation matrix for sum(Concept).
// For each source value, the weight to the target value is 1.0
// (representing the value being included in the sum).
// When targetValues has length 0, columns is 1 (scalar reduction case).
func BuildSumAggregationMatrix(collapsedValues []string, targetValues []string) *AggregationMatrix {
	rows := len(collapsedValues)
	cols := 1
	if len(targetValues) > 0 {
		cols = len(targetValues)
	}

	data := make([][]float64, rows)
	for i := range data {
		data[i] = make([]float64, cols)
		for j := range data[i] {
			data[i][j] = 1.0 // Each source value contributes fully to each target
		}
	}

	return &AggregationMatrix{
		Rows:      rows,
		Columns:   cols,
		Data:      data,
		RowLabels: collapsedValues,
		ColLabels: targetValues,
	}
}

// BuildSizeAggregationMatrix creates an aggregation matrix for size(Concept).
// size() is unitless — each entry is 1.0 (counting each value once).
// Produces a single-column matrix with one row per collapsed value.
func BuildSizeAggregationMatrix(collapsedValues []string) *AggregationMatrix {
	rows := len(collapsedValues)
	data := make([][]float64, rows)
	for i := range data {
		data[i] = []float64{1.0}
	}
	return &AggregationMatrix{
		Rows:      rows,
		Columns:   1,
		Data:      data,
		RowLabels: collapsedValues,
		ColLabels: []string{"count"},
	}
}

// BuildInverseSumAggregationMatrix creates a matrix for 1/sum(Concept).
// Each entry is 1.0 / len(collapsedValues), representing the inverse contribution.
// When targetValues has length 0, columns is 1 (scalar reduction case).
func BuildInverseSumAggregationMatrix(collapsedValues []string, targetValues []string) *AggregationMatrix {
	rows := len(collapsedValues)
	cols := 1
	if len(targetValues) > 0 {
		cols = len(targetValues)
	}

	weight := 1.0 / float64(rows)

	data := make([][]float64, rows)
	for i := range data {
		data[i] = make([]float64, cols)
		for j := range data[i] {
			data[i][j] = weight
		}
	}

	return &AggregationMatrix{
		Rows:      rows,
		Columns:   cols,
		Data:      data,
		RowLabels: collapsedValues,
		ColLabels: targetValues,
	}
}
