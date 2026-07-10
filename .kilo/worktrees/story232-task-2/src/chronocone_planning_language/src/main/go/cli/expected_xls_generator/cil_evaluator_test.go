// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package expected_xls_generator

import (
	"math"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

func TestEvaluateExpression_Literal(t *testing.T) {
	tests := []struct {
		name     string
		node     cil.CILNode
		expected float64
	}{
		{"integer literal", &cil.CILLiteral{Value: 42}, 42},
		{"float literal", &cil.CILLiteral{Value: 3.14}, 3.14},
		{"zero literal", &cil.CILLiteral{Value: 0}, 0},
		{"negative literal", &cil.CILLiteral{Value: -7.5}, -7.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EvaluateExpression(tt.node, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluateExpression_BinaryArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		node     cil.CILNode
		expected float64
	}{
		{"addition", &cil.CILBinaryOp{Op: operator.Add, Left: &cil.CILLiteral{Value: 3}, Right: &cil.CILLiteral{Value: 5}}, 8},
		{"subtraction", &cil.CILBinaryOp{Op: operator.Subtract, Left: &cil.CILLiteral{Value: 10}, Right: &cil.CILLiteral{Value: 4}}, 6},
		{"multiplication", &cil.CILBinaryOp{Op: operator.Multiply, Left: &cil.CILLiteral{Value: 3}, Right: &cil.CILLiteral{Value: 7}}, 21},
		{"division", &cil.CILBinaryOp{Op: operator.Divide, Left: &cil.CILLiteral{Value: 20}, Right: &cil.CILLiteral{Value: 4}}, 5},
		{"division with float", &cil.CILBinaryOp{Op: operator.Divide, Left: &cil.CILLiteral{Value: 7}, Right: &cil.CILLiteral{Value: 2}}, 3.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EvaluateExpression(tt.node, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluateExpression_UnaryNegation(t *testing.T) {
	tests := []struct {
		name     string
		node     cil.CILNode
		expected float64
	}{
		{"negate positive", &cil.CILUnaryOp{Op: operator.Negate, Operand: &cil.CILLiteral{Value: 5}}, -5},
		{"negate negative", &cil.CILUnaryOp{Op: operator.Negate, Operand: &cil.CILLiteral{Value: -3}}, 3},
		{"negate zero", &cil.CILUnaryOp{Op: operator.Negate, Operand: &cil.CILLiteral{Value: 0}}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EvaluateExpression(tt.node, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluateExpression_ComparisonOperators(t *testing.T) {
	tests := []struct {
		name     string
		node     cil.CILNode
		expected float64
	}{
		{"equals true", &cil.CILComparisonOp{Op: operator.Equal, Left: &cil.CILLiteral{Value: 5}, Right: &cil.CILLiteral{Value: 5}}, 1.0},
		{"equals false", &cil.CILComparisonOp{Op: operator.Equal, Left: &cil.CILLiteral{Value: 5}, Right: &cil.CILLiteral{Value: 3}}, 0.0},
		{"not equals true", &cil.CILComparisonOp{Op: operator.NotEqual, Left: &cil.CILLiteral{Value: 5}, Right: &cil.CILLiteral{Value: 3}}, 1.0},
		{"not equals false", &cil.CILComparisonOp{Op: operator.NotEqual, Left: &cil.CILLiteral{Value: 5}, Right: &cil.CILLiteral{Value: 5}}, 0.0},
		{"less than true", &cil.CILComparisonOp{Op: operator.LessThan, Left: &cil.CILLiteral{Value: 3}, Right: &cil.CILLiteral{Value: 5}}, 1.0},
		{"less than false", &cil.CILComparisonOp{Op: operator.LessThan, Left: &cil.CILLiteral{Value: 5}, Right: &cil.CILLiteral{Value: 3}}, 0.0},
		{"greater than true", &cil.CILComparisonOp{Op: operator.GreaterThan, Left: &cil.CILLiteral{Value: 5}, Right: &cil.CILLiteral{Value: 3}}, 1.0},
		{"greater than false", &cil.CILComparisonOp{Op: operator.GreaterThan, Left: &cil.CILLiteral{Value: 3}, Right: &cil.CILLiteral{Value: 5}}, 0.0},
		{"less than or equal true", &cil.CILComparisonOp{Op: operator.LessThanOrEqual, Left: &cil.CILLiteral{Value: 3}, Right: &cil.CILLiteral{Value: 5}}, 1.0},
		{"less than or equal equal", &cil.CILComparisonOp{Op: operator.LessThanOrEqual, Left: &cil.CILLiteral{Value: 5}, Right: &cil.CILLiteral{Value: 5}}, 1.0},
		{"less than or equal false", &cil.CILComparisonOp{Op: operator.LessThanOrEqual, Left: &cil.CILLiteral{Value: 5}, Right: &cil.CILLiteral{Value: 3}}, 0.0},
		{"greater than or equal true", &cil.CILComparisonOp{Op: operator.GreaterThanOrEqual, Left: &cil.CILLiteral{Value: 5}, Right: &cil.CILLiteral{Value: 3}}, 1.0},
		{"greater than or equal equal", &cil.CILComparisonOp{Op: operator.GreaterThanOrEqual, Left: &cil.CILLiteral{Value: 5}, Right: &cil.CILLiteral{Value: 5}}, 1.0},
		{"greater than or equal false", &cil.CILComparisonOp{Op: operator.GreaterThanOrEqual, Left: &cil.CILLiteral{Value: 3}, Right: &cil.CILLiteral{Value: 5}}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EvaluateExpression(tt.node, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluateExpression_LogicalOperators(t *testing.T) {
	tests := []struct {
		name     string
		node     cil.CILNode
		expected float64
	}{
		{"and both true", &cil.CILLogicalOp{Op: operator.And, Left: &cil.CILLiteral{Value: 1}, Right: &cil.CILLiteral{Value: 1}}, 1.0},
		{"and left false", &cil.CILLogicalOp{Op: operator.And, Left: &cil.CILLiteral{Value: 0}, Right: &cil.CILLiteral{Value: 1}}, 0.0},
		{"and right false", &cil.CILLogicalOp{Op: operator.And, Left: &cil.CILLiteral{Value: 1}, Right: &cil.CILLiteral{Value: 0}}, 0.0},
		{"and both false", &cil.CILLogicalOp{Op: operator.And, Left: &cil.CILLiteral{Value: 0}, Right: &cil.CILLiteral{Value: 0}}, 0.0},
		{"or both true", &cil.CILLogicalOp{Op: operator.Or, Left: &cil.CILLiteral{Value: 1}, Right: &cil.CILLiteral{Value: 1}}, 1.0},
		{"or left true", &cil.CILLogicalOp{Op: operator.Or, Left: &cil.CILLiteral{Value: 1}, Right: &cil.CILLiteral{Value: 0}}, 1.0},
		{"or right true", &cil.CILLogicalOp{Op: operator.Or, Left: &cil.CILLiteral{Value: 0}, Right: &cil.CILLiteral{Value: 1}}, 1.0},
		{"or both false", &cil.CILLogicalOp{Op: operator.Or, Left: &cil.CILLiteral{Value: 0}, Right: &cil.CILLiteral{Value: 0}}, 0.0},
		{"not true", &cil.CILNotOp{Operand: &cil.CILLiteral{Value: 0}}, 1.0},
		{"not false", &cil.CILNotOp{Operand: &cil.CILLiteral{Value: 1}}, 0.0},
		{"not non-zero", &cil.CILNotOp{Operand: &cil.CILLiteral{Value: 5}}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EvaluateExpression(tt.node, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluateExpression_BoolAndNilLiterals(t *testing.T) {
	tests := []struct {
		name     string
		node     cil.CILNode
		expected float64
	}{
		{"true literal", &cil.CILBoolLiteral{Value: true}, 1.0},
		{"false literal", &cil.CILBoolLiteral{Value: false}, 0.0},
		{"nil literal", &cil.CILNilLiteral{}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EvaluateExpression(tt.node, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluateExpression_SimpleConditional(t *testing.T) {
	cond := &cil.CILConditional{
		Branches: []cil.CILConditionalBranch{
			{
				Condition: &cil.CILComparisonOp{Op: operator.Equal, Left: &cil.CILInputRef{Name: "data_a"}, Right: &cil.CILLiteral{Value: 10}},
				Body:      &cil.CILLiteral{Value: 100},
			},
			{
				Condition: nil,
				Body:      &cil.CILLiteral{Value: 0},
			},
		},
	}

	lookupTrue := map[string]float64{"data_a": 10}
	result, err := EvaluateExpression(cond, lookupTrue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 100 {
		t.Errorf("expected 100, got %v", result)
	}

	lookupFalse := map[string]float64{"data_a": 5}
	result, err = EvaluateExpression(cond, lookupFalse)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 0 {
		t.Errorf("expected 0, got %v", result)
	}
}

func TestEvaluateExpression_NestedExpressions(t *testing.T) {
	node := &cil.CILBinaryOp{
		Op: operator.Multiply,
		Left: &cil.CILBinaryOp{
			Op:    operator.Add,
			Left:  &cil.CILInputRef{Name: "data_a"},
			Right: &cil.CILInputRef{Name: "data_b"},
		},
		Right: &cil.CILLiteral{Value: 2},
	}

	lookup := map[string]float64{"data_a": 3, "data_b": 5}
	result, err := EvaluateExpression(node, lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := 16.0
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestEvaluateExpression_DivisionByZero(t *testing.T) {
	node := &cil.CILBinaryOp{Op: operator.Divide, Left: &cil.CILLiteral{Value: 10}, Right: &cil.CILLiteral{Value: 0}}
	result, err := EvaluateExpression(node, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !math.IsInf(result, 1) {
		t.Errorf("expected +Inf for division by zero, got %v", result)
	}
}

func TestEvaluateExpression_LookupNode(t *testing.T) {
	lookupData := map[string][]map[string]interface{}{
		"strength_adjusted_for_material": {
			{"Material": "Steel", "Value": 1.0},
			{"Material": "Copper", "Value": 0.9},
			{"Material": "Aluminum", "Value": 0.8},
		},
	}

	tests := []struct {
		name              string
		dimensionValues   map[string]string
		expected          float64
		expectErr         bool
		errContains       string
	}{
		{
			name:            "lookup Steel",
			dimensionValues: map[string]string{"Material": "Steel"},
			expected:        1.0,
		},
		{
			name:            "lookup Copper",
			dimensionValues: map[string]string{"Material": "Copper"},
			expected:        0.9,
		},
		{
			name:            "lookup Aluminum",
			dimensionValues: map[string]string{"Material": "Aluminum"},
			expected:        0.8,
		},
		{
			name:            "lookup not found dimension",
			dimensionValues: map[string]string{"Material": "Titanium"},
			expectErr:       true,
			errContains:     "no matching row found",
		},
		{
			name:            "dimension missing from context",
			dimensionValues: map[string]string{},
			expectErr:       true,
			errContains:     "dimension \"Material\" not found in current context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &cil.CILLookupNode{
				LookupName: "strength_adjusted_for_material",
				Dimensions: []string{"Material"},
				Celltype:   "Strength",
			}
			ctx := &EvalContext{
				LookupData:      lookupData,
				DimensionValues: tt.dimensionValues,
			}
			result, err := EvaluateExpressionWithContext(node, ctx)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errContains)
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluateExpression_LookupNode_MissingData(t *testing.T) {
	node := &cil.CILLookupNode{
		LookupName: "missing_lookup",
		Dimensions: []string{"Material"},
		Celltype:   "Strength",
	}
	ctx := &EvalContext{
		LookupData:      map[string][]map[string]interface{}{},
		DimensionValues: map[string]string{"Material": "Steel"},
	}
	_, err := EvaluateExpressionWithContext(node, ctx)
	if err == nil {
		t.Fatal("expected error for missing lookup data, got nil")
	}
	if !contains(err.Error(), "not found in lookup data") {
		t.Errorf("expected error about missing lookup data, got %q", err.Error())
	}
}

func TestEvaluateExpression_LookupNode_BinaryOp(t *testing.T) {
	// Test lookup node inside a binary operation: raw_strength * strength_adjusted_for_material
	lookupData := map[string][]map[string]interface{}{
		"strength_adjusted_for_material": {
			{"Material": "Steel", "Value": 1.0},
			{"Material": "Copper", "Value": 0.9},
		},
	}
	node := &cil.CILBinaryOp{
		Op: operator.Multiply,
		Left: &cil.CILInputRef{Name: "raw_strength"},
		Right: &cil.CILLookupNode{
			LookupName: "strength_adjusted_for_material",
			Dimensions: []string{"Material"},
			Celltype:   "Strength",
		},
	}
	ctx := &EvalContext{
		ValueLookup:     map[string]float64{"raw_strength": 10.0},
		LookupData:      lookupData,
		DimensionValues: map[string]string{"Material": "Copper"},
	}
	result, err := EvaluateExpressionWithContext(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := 9.0 // 10.0 * 0.9
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
