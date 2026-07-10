// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package expected_xls_generator

import (
	"fmt"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
	"github.com/xuri/excelize/v2"
)

func testColName(col int) string {
	name, _ := excelize.ColumnNumberToName(col + 1)
	return name
}

func TestExpressionToExcel_Literal(t *testing.T) {
	result := ExpressionToExcel(&cil.CILLiteral{Value: 42}, "", 0, 0, testColName, nil)
	if result != "42" {
		t.Errorf("expected '42', got %q", result)
	}
}

func TestExpressionToExcel_InputRef(t *testing.T) {
	result := ExpressionToExcel(&cil.CILInputRef{Name: "data_a"}, "data_a", 0, 2, testColName, nil)
	expected := "data_a!C2"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpressionToExcel_BinaryOp_Add(t *testing.T) {
	result := ExpressionToExcel(&cil.CILBinaryOp{
		Op: operator.Add, Left: &cil.CILLiteral{Value: 3}, Right: &cil.CILLiteral{Value: 5},
	}, "", 0, 0, testColName, nil)
	if result != "(3+5)" {
		t.Errorf("expected '(3+5)', got %q", result)
	}
}

func TestExpressionToExcel_UnaryNegate(t *testing.T) {
	result := ExpressionToExcel(&cil.CILUnaryOp{
		Op: operator.Negate, Operand: &cil.CILLiteral{Value: 5},
	}, "", 0, 0, testColName, nil)
	if result != "-5" {
		t.Errorf("expected '-5', got %q", result)
	}
}

func TestExpressionToExcel_Comparison_Equal(t *testing.T) {
	result := ExpressionToExcel(&cil.CILComparisonOp{
		Op: operator.Equal, Left: &cil.CILInputRef{Name: "data_a"}, Right: &cil.CILLiteral{Value: 10},
	}, "data_a", 0, 2, testColName, nil)
	if result != "data_a!C2=10" {
		t.Errorf("expected 'data_a!C2=10', got %q", result)
	}
}

func TestExpressionToExcel_Logical_And(t *testing.T) {
	result := ExpressionToExcel(&cil.CILLogicalOp{
		Op: operator.And, Left: &cil.CILInputRef{Name: "data_a"}, Right: &cil.CILInputRef{Name: "data_b"},
	}, "data_a", 0, 2, testColName, nil)
	// Both input refs use the passed connectorSheet "data_a"
	expected := "AND(data_a!C2,data_a!C2)"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpressionToExcel_Not(t *testing.T) {
	result := ExpressionToExcel(&cil.CILNotOp{
		Operand: &cil.CILInputRef{Name: "data_a"},
	}, "data_a", 0, 2, testColName, nil)
	if result != "NOT(data_a!C2)" {
		t.Errorf("expected 'NOT(data_a!C2)', got %q", result)
	}
}

func TestExpressionToExcel_BoolTrue(t *testing.T) {
	result := ExpressionToExcel(&cil.CILBoolLiteral{Value: true}, "", 0, 0, testColName, nil)
	if result != "TRUE()" {
		t.Errorf("expected 'TRUE()', got %q", result)
	}
}

func TestExpressionToExcel_BoolFalse(t *testing.T) {
	result := ExpressionToExcel(&cil.CILBoolLiteral{Value: false}, "", 0, 0, testColName, nil)
	if result != "FALSE()" {
		t.Errorf("expected 'FALSE()', got %q", result)
	}
}

func TestExpressionToExcel_Nil(t *testing.T) {
	result := ExpressionToExcel(&cil.CILNilLiteral{}, "", 0, 0, testColName, nil)
	if result != "0" {
		t.Errorf("expected '0', got %q", result)
	}
}

func TestExpressionToExcel_WhenClause(t *testing.T) {
	result := ExpressionToExcel(&cil.CILWhenClause{
		Clauses: []cil.CILWhenCase{
			{Condition: &cil.CILComparisonOp{Op: operator.Equal, Left: &cil.CILInputRef{Name: "data_a"}, Right: &cil.CILLiteral{Value: 1}}, Body: &cil.CILLiteral{Value: 100}},
			{Condition: &cil.CILComparisonOp{Op: operator.Equal, Left: &cil.CILInputRef{Name: "data_a"}, Right: &cil.CILLiteral{Value: 2}}, Body: &cil.CILLiteral{Value: 200}},
		},
		Default: &cil.CILLiteral{Value: 0},
	}, "data_a", 0, 2, testColName, nil)
	expected := "IF(data_a!C2=1,100,IF(data_a!C2=2,200,0))"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpressionToExcel_LookupNode_SingleDim(t *testing.T) {
	dimensionColumnIndexes := map[string]int{"Material": 1}
	node := &cil.CILLookupNode{
		LookupName: "strength_adjusted_for_material",
		Dimensions: []string{"Material"},
	}
	result := ExpressionToExcel(node, "data_a", 0, 2, testColName, dimensionColumnIndexes)
	expected := "VLOOKUP(expected_output!B2,strength_adjusted_for_material!A:C,3,FALSE)"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpressionToCellExcel_Literal(t *testing.T) {
	result := ExpressionToCellExcel(&cil.CILLiteral{Value: 42}, map[string]int{}, testColName, 5, 0, nil)
	if result != "42" {
		t.Errorf("expected '42', got %q", result)
	}
}

func TestExpressionToCellExcel_InputRef(t *testing.T) {
	inputColumnIndexes := map[string]int{"data_a": 0}
	result := ExpressionToCellExcel(&cil.CILInputRef{Name: "data_a"}, inputColumnIndexes, testColName, 5, 2, nil)
	expected := "C5"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpressionToCellExcel_BinaryOp(t *testing.T) {
	inputColumnIndexes := map[string]int{"data_a": 0, "data_b": 1}
	node := &cil.CILBinaryOp{
		Op: operator.Add,
		Left:  &cil.CILInputRef{Name: "data_a"},
		Right: &cil.CILInputRef{Name: "data_b"},
	}
	result := ExpressionToCellExcel(node, inputColumnIndexes, testColName, 5, 2, nil)
	expected := "(C5+D5)"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpressionToCellExcel_Conditional(t *testing.T) {
	inputColumnIndexes := map[string]int{"data_a": 0, "data_b": 1}
	node := &cil.CILConditional{
		Branches: []cil.CILConditionalBranch{
			{
				Condition: &cil.CILComparisonOp{
					Op: operator.Equal,
					Left:  &cil.CILInputRef{Name: "data_a"},
					Right: &cil.CILLiteral{Value: 10},
				},
				Body: &cil.CILBinaryOp{
					Op: operator.Add,
					Left:  &cil.CILInputRef{Name: "data_a"},
					Right: &cil.CILInputRef{Name: "data_b"},
				},
			},
			{
				Condition: &cil.CILComparisonOp{
					Op: operator.NotEqual,
					Left:  &cil.CILInputRef{Name: "data_a"},
					Right: &cil.CILLiteral{Value: 5},
				},
				Body: &cil.CILBinaryOp{
					Op: operator.Multiply,
					Left:  &cil.CILInputRef{Name: "data_a"},
					Right: &cil.CILLiteral{Value: 2},
				},
			},
		},
	}
	result := ExpressionToCellExcel(node, inputColumnIndexes, testColName, 5, 2, nil)
	expected := "IF(C5=10,(C5+D5),(C5*2))"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpressionToCellExcel_WhenClause(t *testing.T) {
	inputColumnIndexes := map[string]int{"data_a": 0, "data_b": 1}
	node := &cil.CILWhenClause{
		Clauses: []cil.CILWhenCase{
			{Condition: &cil.CILComparisonOp{Op: operator.Equal, Left: &cil.CILInputRef{Name: "data_a"}, Right: &cil.CILLiteral{Value: 1}}, Body: &cil.CILLiteral{Value: 100}},
			{Condition: &cil.CILComparisonOp{Op: operator.Equal, Left: &cil.CILInputRef{Name: "data_a"}, Right: &cil.CILLiteral{Value: 2}}, Body: &cil.CILLiteral{Value: 200}},
		},
		Default: &cil.CILLiteral{Value: 0},
	}
	result := ExpressionToCellExcel(node, inputColumnIndexes, testColName, 5, 2, nil)
	expected := "IF(C5=1,100,IF(C5=2,200,0))"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpressionToCellExcel_ConditionalWithElse(t *testing.T) {
	inputColumnIndexes := map[string]int{"data_a": 0}
	node := &cil.CILConditional{
		Branches: []cil.CILConditionalBranch{
			{Condition: &cil.CILComparisonOp{Op: operator.Equal, Left: &cil.CILInputRef{Name: "data_a"}, Right: &cil.CILLiteral{Value: 10}}, Body: &cil.CILLiteral{Value: 100}},
			{Condition: nil, Body: &cil.CILLiteral{Value: 0}},
		},
	}
	result := ExpressionToCellExcel(node, inputColumnIndexes, testColName, 5, 2, nil)
	expected := "IF(C5=10,100,0)"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpressionToCellExcel_LogicalOp(t *testing.T) {
	inputColumnIndexes := map[string]int{"data_a": 0, "data_b": 1}
	node := &cil.CILLogicalOp{
		Op: operator.And,
		Left:  &cil.CILInputRef{Name: "data_a"},
		Right: &cil.CILInputRef{Name: "data_b"},
	}
	result := ExpressionToCellExcel(node, inputColumnIndexes, testColName, 5, 2, nil)
	expected := "AND(C5,D5)"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpressionToCellExcel_Comparison_OtherOps(t *testing.T) {
	inputColumnIndexes := map[string]int{"data_a": 0}
	tests := []struct {
		op         operator.Operator
		expectedOp string
	}{
		{operator.LessThanOrEqual, "<="},
		{operator.GreaterThanOrEqual, ">="},
		{operator.GreaterThan, ">"},
		{operator.LessThan, "<"},
	}

	for _, tt := range tests {
		node := &cil.CILComparisonOp{
			Op:    tt.op,
			Left:  &cil.CILInputRef{Name: "data_a"},
			Right: &cil.CILLiteral{Value: 10},
		}
		result := ExpressionToCellExcel(node, inputColumnIndexes, testColName, 5, 2, nil)
		expected := fmt.Sprintf("C5%s10", tt.expectedOp)
		if result != expected {
			t.Errorf("op %q: expected %q, got %q", tt.op, expected, result)
		}
	}
}

func TestExpressionToCellExcel_UnaryNegate(t *testing.T) {
	inputColumnIndexes := map[string]int{"data_a": 0}
	node := &cil.CILUnaryOp{
		Op:      operator.Negate,
		Operand: &cil.CILInputRef{Name: "data_a"},
	}
	result := ExpressionToCellExcel(node, inputColumnIndexes, testColName, 5, 2, nil)
	expected := "-C5"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpressionToCellExcel_LookupNode_SingleDim(t *testing.T) {
	dimensionColumnIndexes := map[string]int{"Material": 1}
	inputColumnIndexes := map[string]int{"raw_strength": 0}
	node := &cil.CILLookupNode{
		LookupName: "strength_adjusted_for_material",
		Dimensions: []string{"Material"},
	}
	result := ExpressionToCellExcel(node, inputColumnIndexes, testColName, 2, 2, dimensionColumnIndexes)
	expected := "VLOOKUP(B2,strength_adjusted_for_material!A:C,3,FALSE)"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpressionToCellExcel_LookupNode_MultiDim(t *testing.T) {
	dimensionColumnIndexes := map[string]int{"Country": 0, "Product": 1}
	inputColumnIndexes := map[string]int{}
	node := &cil.CILLookupNode{
		LookupName: "country_product_index",
		Dimensions: []string{"Country", "Product"},
	}
	result := ExpressionToCellExcel(node, inputColumnIndexes, testColName, 2, 2, dimensionColumnIndexes)
	expected := `VLOOKUP(A2&"|"&B2,country_product_index!A:D,4,FALSE)`
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpressionToCellExcel_LookupNode_BinaryOp(t *testing.T) {
	dimensionColumnIndexes := map[string]int{"Material": 1}
	inputColumnIndexes := map[string]int{"raw_strength": 0}
	node := &cil.CILBinaryOp{
		Op:    operator.Multiply,
		Left:  &cil.CILInputRef{Name: "raw_strength"},
		Right: &cil.CILLookupNode{
			LookupName: "strength_adjusted_for_material",
			Dimensions: []string{"Material"},
		},
	}
	result := ExpressionToCellExcel(node, inputColumnIndexes, testColName, 2, 2, dimensionColumnIndexes)
	expected := "(C2*VLOOKUP(B2,strength_adjusted_for_material!A:C,3,FALSE))"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
