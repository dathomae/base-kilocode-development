// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package calculation

import (
	"testing"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

var _ = testing.AllocsPerRun

// registerFloat64Tensor creates a float64 tensor with given values, registers it as a
// MetricComputation on the controller, and returns the tensor for mask manipulation.
func registerFloat64Tensor(t *testing.T, controller *ChronoconeComputationController, name string, values []float64) *ChronoconeTensor[any] {
	t.Helper()
	var data []any
	for _, v := range values {
		data = append(data, v)
	}
	tensor := NewChronoconeTensor(data, []string{}, []int{len(values)})
	mc := NewMetricComputation(1)
	if err := controller.RegisterMetricComputation(name, mc); err != nil {
		t.Fatalf("RegisterMetricComputation(%q): %v", name, err)
	}
	if err := mc.SetTensor(tensor); err != nil {
		t.Fatalf("SetTensor(%q): %v", name, err)
	}
	return tensor
}

// TestEvaluatePostfixFloat64_ComparisonGreaterThan tests the > operator.
func TestEvaluatePostfixFloat64_ComparisonGreaterThan(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{5.0, 2.0})
	registerFloat64Tensor(t, controller, "b", []float64{3.0, 3.0})

	ops := []PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "push", Name: "b"},
		{Kind: "apply", Op: operator.GreaterThan},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	// 5 > 3 should be true (mask bit 0), 2 > 3 should be false (mask bit 1)
	if !result.IsMaskedTrue(0) {
		t.Error("IsMaskedTrue(0) = false, want true for 5 > 3")
	}
	if result.IsMaskedTrue(1) {
		t.Error("IsMaskedTrue(1) = true, want false for 2 > 3")
	}
}

// TestEvaluatePostfixFloat64_ComparisonEqual tests the == operator.
func TestEvaluatePostfixFloat64_ComparisonEqual(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{3.0, 5.0})
	registerFloat64Tensor(t, controller, "b", []float64{3.0, 7.0})

	ops := []PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "push", Name: "b"},
		{Kind: "apply", Op: operator.Equal},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	// 3 == 3 should be true (mask bit 0), 5 == 7 should be false (mask bit 1)
	if !result.IsMaskedTrue(0) {
		t.Error("IsMaskedTrue(0) = false, want true for 3 == 3")
	}
	if result.IsMaskedTrue(1) {
		t.Error("IsMaskedTrue(1) = true, want false for 5 == 7")
	}
}

// TestEvaluatePostfixFloat64_ComparisonNotEqual tests the != operator.
func TestEvaluatePostfixFloat64_ComparisonNotEqual(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{3.0, 5.0})
	registerFloat64Tensor(t, controller, "b", []float64{3.0, 7.0})

	ops := []PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "push", Name: "b"},
		{Kind: "apply", Op: operator.NotEqual},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	// 3 != 3 should be false (mask bit 1), 5 != 7 should be true (mask bit 0)
	if result.IsMaskedTrue(0) {
		t.Error("IsMaskedTrue(0) = true, want false for 3 != 3")
	}
	if !result.IsMaskedTrue(1) {
		t.Error("IsMaskedTrue(1) = false, want true for 5 != 7")
	}
}

// TestEvaluatePostfixFloat64_ComparisonLessThan tests the < operator.
func TestEvaluatePostfixFloat64_ComparisonLessThan(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{2.0, 5.0})
	registerFloat64Tensor(t, controller, "b", []float64{3.0, 7.0})

	ops := []PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "push", Name: "b"},
		{Kind: "apply", Op: operator.LessThan},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	// 2 < 3 should be true (mask bit 0), 5 < 7 should be true (mask bit 0)
	if !result.IsMaskedTrue(0) {
		t.Error("IsMaskedTrue(0) = false, want true for 2 < 3")
	}
	if !result.IsMaskedTrue(1) {
		t.Error("IsMaskedTrue(1) = false, want true for 5 < 7")
	}
}

// TestEvaluatePostfixFloat64_ComparisonLessEqual tests the <= operator.
func TestEvaluatePostfixFloat64_ComparisonLessEqual(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{2.0, 5.0})
	registerFloat64Tensor(t, controller, "b", []float64{3.0, 5.0})

	ops := []PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "push", Name: "b"},
		{Kind: "apply", Op: operator.LessThanOrEqual},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	// 2 <= 3 should be true (mask bit 0), 5 <= 5 should be true (mask bit 0)
	if !result.IsMaskedTrue(0) {
		t.Error("IsMaskedTrue(0) = false, want true for 2 <= 3")
	}
	if !result.IsMaskedTrue(1) {
		t.Error("IsMaskedTrue(1) = false, want true for 5 <= 5")
	}
}

// TestEvaluatePostfixFloat64_ComparisonLessEqualFalse tests the <= operator with a false case.
func TestEvaluatePostfixFloat64_ComparisonGreater(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{5.0, 2.0})
	registerFloat64Tensor(t, controller, "b", []float64{3.0, 7.0})

	ops := []PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "push", Name: "b"},
		{Kind: "apply", Op: operator.GreaterThan},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	// 5 > 3 should be true (mask bit 0), 2 > 7 should be false (mask bit 1)
	if !result.IsMaskedTrue(0) {
		t.Error("IsMaskedTrue(0) = false, want true for 5 > 3")
	}
	if result.IsMaskedTrue(1) {
		t.Error("IsMaskedTrue(1) = true, want false for 2 > 7")
	}
}

// TestEvaluatePostfixFloat64_ScalarComparison tests comparison with scalar literal.
func TestEvaluatePostfixFloat64_ScalarComparison(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{5.0, 2.0})

	ops := []PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "push", Value: 3.0},
		{Kind: "apply", Op: operator.GreaterThan},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	// 5 > 3 should be true (mask bit 0), 2 > 3 should be false (mask bit 1)
	if !result.IsMaskedTrue(0) {
		t.Error("IsMaskedTrue(0) = false, want true for 5 > 3")
	}
	if result.IsMaskedTrue(1) {
		t.Error("IsMaskedTrue(1) = true, want false for 2 > 3")
	}
}

// TestEvaluatePostfixFloat64_LogicalAnd tests the and operator.
func TestEvaluatePostfixFloat64_LogicalAnd(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{1.0, 2.0})
	registerFloat64Tensor(t, controller, "b", []float64{3.0, 4.0})

	// Set mask for a: all false (all bits 1)
	aTensor := controller.GetMetricComputation("a").GetTensor()
	aTensor.SetMask(AllFalseMask(len(aTensor.Mask())))
	// Set mask for b: all true (all bits 0)
	bTensor := controller.GetMetricComputation("b").GetTensor()
	bTensor.SetMask(AllTrueMask(len(bTensor.Mask())))

	ops := []PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "push", Name: "b"},
		{Kind: "apply", Op: operator.And},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	// AND of (T,F) and (T,T) should be (F,F) - both false
	if result.IsMaskedTrue(0) {
		t.Error("IsMaskedTrue(0) = true, want false for AND(T,F)")
	}
	if result.IsMaskedTrue(1) {
		t.Error("IsMaskedTrue(1) = true, want false for AND(F,T)")
	}
}

// TestEvaluatePostfixFloat64_LogicalOr tests the or operator.
func TestEvaluatePostfixFloat64_LogicalOr(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{1.0, 2.0})
	registerFloat64Tensor(t, controller, "b", []float64{3.0, 4.0})

	// Set mask for a: all false (all bits 1)
	aTensor := controller.GetMetricComputation("a").GetTensor()
	aTensor.SetMask(AllFalseMask(len(aTensor.Mask())))
	// Set mask for b: all false (all bits 1)
	bTensor := controller.GetMetricComputation("b").GetTensor()
	bTensor.SetMask(AllFalseMask(len(bTensor.Mask())))

	ops := []PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "push", Name: "b"},
		{Kind: "apply", Op: operator.Or},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	// OR of (F,F) and (F,F) should be (F,F) - both false
	if result.IsMaskedTrue(0) {
		t.Error("IsMaskedTrue(0) = true, want false for OR(F,F)")
	}
	if result.IsMaskedTrue(1) {
		t.Error("IsMaskedTrue(1) = true, want false for OR(F,F)")
	}
}

// TestEvaluatePostfixFloat64_LogicalNot tests the not operator.
func TestEvaluatePostfixFloat64_LogicalNot(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{1.0, 2.0})

	// Set mask for a: all false (all bits 1)
	aTensor := controller.GetMetricComputation("a").GetTensor()
	aTensor.SetMask(AllFalseMask(len(aTensor.Mask())))

	ops := []PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "apply", Op: operator.Not},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	// NOT of (F,F) should be (T,T) - both true after inversion
	if !result.IsMaskedTrue(0) {
		t.Error("IsMaskedTrue(0) = false, want true for NOT(F)")
	}
	if !result.IsMaskedTrue(1) {
		t.Error("IsMaskedTrue(1) = false, want true for NOT(F)")
	}
}

// TestEvaluatePostfixFloat64_BoolLiteralTrue tests bool literal with value 1.0.
func TestEvaluatePostfixFloat64_BoolLiteralTrue(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	ops := []PostfixOp{
		{Kind: "push", Op: operator.Bool, Value: 1.0},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if !result.IsMaskedTrue(0) {
		t.Error("IsMaskedTrue(0) = false, want true for bool literal 1.0")
	}
}

// TestEvaluatePostfixFloat64_BoolLiteralFalse tests bool literal with value 0.0.
func TestEvaluatePostfixFloat64_BoolLiteralFalse(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	ops := []PostfixOp{
		{Kind: "push", Op: operator.Bool, Value: 0.0},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if result.IsMaskedTrue(0) {
		t.Error("IsMaskedTrue(0) = true, want false for bool literal 0.0")
	}
}

// TestEvaluatePostfixFloat64_NilLiteral tests nil literal.
func TestEvaluatePostfixFloat64_NilLiteral(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	ops := []PostfixOp{
		{Kind: "push", Op: operator.Nil},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if result.IsMaskedTrue(0) {
		t.Error("IsMaskedTrue(0) = true, want false for nil literal")
	}
}

// TestEvaluatePostfixFloat64_ComparisonStackUnderflow tests error when stack underflows.
func TestEvaluatePostfixFloat64_ComparisonStackUnderflow(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{5.0})

	ops := []PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "apply", Op: operator.GreaterThan},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err == nil {
		t.Error("expected error for stack underflow, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result for stack underflow, got non-nil")
	}
}

// TestEvaluatePostfixFloat64_SimpleConditionalIfElse verifies two-branch conditional
// with complementary masks: if a > 0 then b else c.
func TestEvaluatePostfixFloat64_SimpleConditionalIfElse(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{5.0, -3.0})
	registerFloat64Tensor(t, controller, "b", []float64{100.0, 200.0})
	registerFloat64Tensor(t, controller, "c", []float64{300.0, 400.0})

	ops := []PostfixOp{
		{
			Kind: "conditional",
			Branches: []PostfixBranch{
				{
					Condition: []PostfixOp{
						{Kind: "push", Name: "a"},
						{Kind: "push", Value: 0.0},
						{Kind: "apply", Op: operator.GreaterThan},
					},
					Body: []PostfixOp{
						{Kind: "push", Name: "b"},
					},
				},
				{
					Condition: nil,
					Body: []PostfixOp{
						{Kind: "push", Name: "c"},
					},
				},
			},
		},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}

	expected := []float64{100.0, 400.0}
	if len(result.Data()) != len(expected) {
		t.Fatalf("result data length %d, expected %d", len(result.Data()), len(expected))
	}
	for i, want := range expected {
		if result.Data()[i] != want {
			t.Errorf("result[%d] = %v, want %v", i, result.Data()[i], want)
		}
	}
}

// TestEvaluatePostfixFloat64_ConditionalIfElseifElse verifies three-branch conditional
// with proper fallthrough: if a > 0 then b; elseif a < 0 then c; else d.
func TestEvaluatePostfixFloat64_ConditionalIfElseifElse(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{5.0, -3.0, 0.0})
	registerFloat64Tensor(t, controller, "b", []float64{100.0, 200.0, 300.0})
	registerFloat64Tensor(t, controller, "c", []float64{400.0, 500.0, 600.0})
	registerFloat64Tensor(t, controller, "d", []float64{700.0, 800.0, 900.0})

	ops := []PostfixOp{
		{
			Kind: "conditional",
			Branches: []PostfixBranch{
				{
					Condition: []PostfixOp{
						{Kind: "push", Name: "a"},
						{Kind: "push", Value: 0.0},
						{Kind: "apply", Op: operator.GreaterThan},
					},
					Body: []PostfixOp{
						{Kind: "push", Name: "b"},
					},
				},
				{
					Condition: []PostfixOp{
						{Kind: "push", Name: "a"},
						{Kind: "push", Value: 0.0},
						{Kind: "apply", Op: operator.LessThan},
					},
					Body: []PostfixOp{
						{Kind: "push", Name: "c"},
					},
				},
				{
					Condition: nil,
					Body: []PostfixOp{
						{Kind: "push", Name: "d"},
					},
				},
			},
		},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}

	expected := []float64{100.0, 500.0, 900.0}
	if len(result.Data()) != len(expected) {
		t.Fatalf("result data length %d, expected %d", len(result.Data()), len(expected))
	}
	for i, want := range expected {
		if result.Data()[i] != want {
			t.Errorf("result[%d] = %v, want %v", i, result.Data()[i], want)
		}
	}
}

// TestEvaluatePostfixFloat64_WhenClauseWithDefault verifies multiple when clauses
// plus a default branch.
func TestEvaluatePostfixFloat64_WhenClauseWithDefault(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{5.0, -3.0, 0.0})
	registerFloat64Tensor(t, controller, "b", []float64{100.0, 200.0, 300.0})
	registerFloat64Tensor(t, controller, "c", []float64{400.0, 500.0, 600.0})
	registerFloat64Tensor(t, controller, "d", []float64{700.0, 800.0, 900.0})

	ops := []PostfixOp{
		{
			Kind: "conditional",
			Branches: []PostfixBranch{
				{
					Condition: []PostfixOp{
						{Kind: "push", Name: "a"},
						{Kind: "push", Value: 0.0},
						{Kind: "apply", Op: operator.GreaterThan},
					},
					Body: []PostfixOp{
						{Kind: "push", Name: "b"},
					},
				},
				{
					Condition: []PostfixOp{
						{Kind: "push", Name: "a"},
						{Kind: "push", Value: 0.0},
						{Kind: "apply", Op: operator.LessThan},
					},
					Body: []PostfixOp{
						{Kind: "push", Name: "c"},
					},
				},
				{
					Condition: nil,
					Body: []PostfixOp{
						{Kind: "push", Name: "d"},
					},
				},
			},
		},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}

	expected := []float64{100.0, 500.0, 900.0}
	if len(result.Data()) != len(expected) {
		t.Fatalf("result data length %d, expected %d", len(result.Data()), len(expected))
	}
	for i, want := range expected {
		if result.Data()[i] != want {
			t.Errorf("result[%d] = %v, want %v", i, result.Data()[i], want)
		}
	}
}

// TestEvaluatePostfixFloat64_NestedConditional verifies a conditional inside a
// branch body.
func TestEvaluatePostfixFloat64_NestedConditional(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{5.0, -5.0})
	registerFloat64Tensor(t, controller, "b", []float64{10.0, -10.0})
	registerFloat64Tensor(t, controller, "x", []float64{1.0, 2.0})
	registerFloat64Tensor(t, controller, "y", []float64{3.0, 4.0})
	registerFloat64Tensor(t, controller, "z", []float64{5.0, 6.0})

	ops := []PostfixOp{
		{
			Kind: "conditional",
			Branches: []PostfixBranch{
				{
					Condition: []PostfixOp{
						{Kind: "push", Name: "a"},
						{Kind: "push", Value: 0.0},
						{Kind: "apply", Op: operator.GreaterThan},
					},
					Body: []PostfixOp{
						{
							Kind: "conditional",
							Branches: []PostfixBranch{
								{
									Condition: []PostfixOp{
										{Kind: "push", Name: "b"},
										{Kind: "push", Value: 0.0},
										{Kind: "apply", Op: operator.GreaterThan},
									},
									Body: []PostfixOp{
										{Kind: "push", Name: "x"},
									},
								},
								{
									Condition: nil,
									Body: []PostfixOp{
										{Kind: "push", Name: "y"},
									},
								},
							},
						},
					},
				},
				{
					Condition: nil,
					Body: []PostfixOp{
						{Kind: "push", Name: "z"},
					},
				},
			},
		},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("EvaluatePostfixFloat64 failed: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}

	expected := []float64{1.0, 6.0}
	if len(result.Data()) != len(expected) {
		t.Fatalf("result data length %d, expected %d", len(result.Data()), len(expected))
	}
	for i, want := range expected {
		if result.Data()[i] != want {
			t.Errorf("result[%d] = %v, want %v", i, result.Data()[i], want)
		}
	}
}

// TestEvaluatePostfixFloat64_ConditionalMissingBranches verifies error when
// a conditional op has no branches.
func TestEvaluatePostfixFloat64_ConditionalMissingBranches(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	ops := []PostfixOp{
		{Kind: "conditional", Branches: []PostfixBranch{}},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err == nil {
		t.Error("expected error for missing branches, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result for missing branches, got non-nil")
	}
}

// TestEvaluatePostfixFloat64_ConditionalEmptyBody verifies error when a branch
// body is empty.
func TestEvaluatePostfixFloat64_ConditionalEmptyBody(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	ops := []PostfixOp{
		{
			Kind: "conditional",
			Branches: []PostfixBranch{
				{
					Condition: nil,
					Body:      []PostfixOp{},
				},
			},
		},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err == nil {
		t.Error("expected error for empty branch body, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result for empty branch body, got non-nil")
	}
}

// TestEvaluatePostfixFloat64_ConditionalBodyStackError verifies error when a
// branch body causes a stack underflow.
func TestEvaluatePostfixFloat64_ConditionalBodyStackError(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{})

	registerFloat64Tensor(t, controller, "a", []float64{5.0})

	ops := []PostfixOp{
		{
			Kind: "conditional",
			Branches: []PostfixBranch{
				{
					Condition: nil,
					Body: []PostfixOp{
						{Kind: "push", Name: "a"},
						{Kind: "apply", Op: operator.Add}, // stack underflow
					},
				},
			},
		},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err == nil {
		t.Error("expected error for body stack underflow, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result for body stack underflow, got non-nil")
	}
}


