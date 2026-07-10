// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/tensorops"
)

func TestCILCodeGenerator_GenerateAggregationMatrix_2x2Sum(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	matrix := &tensorops.AggregationMatrix{
		Rows:      2,
		Columns:   2,
		RowLabels: []string{"USA", "Canada"},
		ColLabels: []string{"widget", "gadget"},
		Data: [][]float64{
			{1.0, 1.0},
			{1.0, 1.0},
		},
	}
	code := gen.GenerateAggregationMatrix("aggregationMatrix_Country", matrix)

	if !strings.Contains(code, "var aggregationMatrix_Country = &tensorops.AggregationMatrix{") {
		t.Errorf("missing AggregationMatrix declaration in:\n%s", code)
	}
	if !strings.Contains(code, "Rows: 2") {
		t.Errorf("missing Rows: 2 in:\n%s", code)
	}
	if !strings.Contains(code, "Columns: 2") {
		t.Errorf("missing Columns: 2 in:\n%s", code)
	}
	if !strings.Contains(code, `RowLabels: []string{"USA", "Canada"}`) {
		t.Errorf("missing RowLabels in:\n%s", code)
	}
	if !strings.Contains(code, `ColLabels: []string{"widget", "gadget"}`) {
		t.Errorf("missing ColLabels in:\n%s", code)
	}
	if !strings.Contains(code, "Data: [][]float64{") {
		t.Errorf("missing Data opening in:\n%s", code)
	}
	if !strings.Contains(code, "{1, 1},") {
		t.Errorf("missing first data row in:\n%s", code)
	}
	if !strings.Contains(code, "{1, 1},") {
		t.Errorf("missing second data row in:\n%s", code)
	}
}

func TestCILCodeGenerator_GenerateAggregationMatrix_3x1Sum(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	matrix := &tensorops.AggregationMatrix{
		Rows:      3,
		Columns:   1,
		RowLabels: []string{"USA", "Canada", "Mexico"},
		ColLabels: []string{"NorthAmerica"},
		Data: [][]float64{
			{1.0},
			{1.0},
			{1.0},
		},
	}
	code := gen.GenerateAggregationMatrix("aggregationMatrix_Region", matrix)

	if !strings.Contains(code, "var aggregationMatrix_Region = &tensorops.AggregationMatrix{") {
		t.Errorf("missing AggregationMatrix declaration in:\n%s", code)
	}
	if !strings.Contains(code, "Rows: 3") {
		t.Errorf("missing Rows: 3 in:\n%s", code)
	}
	if !strings.Contains(code, "Columns: 1") {
		t.Errorf("missing Columns: 1 in:\n%s", code)
	}
	if !strings.Contains(code, `RowLabels: []string{"USA", "Canada", "Mexico"}`) {
		t.Errorf("missing RowLabels in:\n%s", code)
	}
	if !strings.Contains(code, `ColLabels: []string{"NorthAmerica"}`) {
		t.Errorf("missing ColLabels in:\n%s", code)
	}
	if !strings.Contains(code, "{1},") {
		t.Errorf("missing data rows in:\n%s", code)
	}
}

func TestCILCodeGenerator_GenerateAggregationMatrix_Identity(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	matrix := &tensorops.AggregationMatrix{
		Rows:      3,
		Columns:   3,
		RowLabels: []string{"A", "B", "C"},
		ColLabels: []string{"X", "Y", "Z"},
		Data: [][]float64{
			{1.0, 0.0, 0.0},
			{0.0, 1.0, 0.0},
			{0.0, 0.0, 1.0},
		},
	}
	code := gen.GenerateAggregationMatrix("aggregationMatrix_Identity", matrix)

	if !strings.Contains(code, "var aggregationMatrix_Identity = &tensorops.AggregationMatrix{") {
		t.Errorf("missing AggregationMatrix declaration in:\n%s", code)
	}
	if !strings.Contains(code, "Rows: 3") {
		t.Errorf("missing Rows: 3 in:\n%s", code)
	}
	if !strings.Contains(code, "Columns: 3") {
		t.Errorf("missing Columns: 3 in:\n%s", code)
	}
	if !strings.Contains(code, "{1, 0, 0},") {
		t.Errorf("missing first identity row in:\n%s", code)
	}
	if !strings.Contains(code, "{0, 1, 0},") {
		t.Errorf("missing second identity row in:\n%s", code)
	}
	if !strings.Contains(code, "{0, 0, 1},") {
		t.Errorf("missing third identity row in:\n%s", code)
	}
}

func TestCILCodeGenerator_GenerateAggregationMatrix_Nil(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	code := gen.GenerateAggregationMatrix("aggregationMatrix_Nil", nil)

	expected := "var aggregationMatrix_Nil = &tensorops.AggregationMatrix{}"
	if code != expected {
		t.Errorf("expected %q, got %q", expected, code)
	}
}

func TestCILCodeGenerator_GenerateAggregationMatrix_Empty(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	matrix := &tensorops.AggregationMatrix{}
	code := gen.GenerateAggregationMatrix("aggregationMatrix_Empty", matrix)

	expected := "var aggregationMatrix_Empty = &tensorops.AggregationMatrix{}"
	if code != expected {
		t.Errorf("expected %q, got %q", expected, code)
	}
}
