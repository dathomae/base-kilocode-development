// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/calculation"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// postfixOpEqual compares two PostfixOp values deeply, including Branches slices.
func postfixOpEqual(a, b calculation.PostfixOp) bool {
	if a.Kind != b.Kind || a.Op != b.Op || a.Name != b.Name || a.Value != b.Value {
		return false
	}
	if len(a.Branches) != len(b.Branches) {
		return false
	}
	for i := range a.Branches {
		if !postfixOpsEqual(a.Branches[i].Condition, b.Branches[i].Condition) {
			return false
		}
		if !postfixOpsEqual(a.Branches[i].Body, b.Branches[i].Body) {
			return false
		}
	}
	return true
}

// postfixOpsEqual compares two PostfixOp slices element-by-element.
func postfixOpsEqual(a, b []calculation.PostfixOp) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !postfixOpEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func TestWalkToPostfix_ComparisonOp(t *testing.T) {
	node := &CILComparisonOp{
		Op: operator.GreaterThan,
		Left:  &CILInputRef{Name: "a"},
		Right: &CILLiteral{Value: 5.0},
	}
	ops, err := WalkCILTreeToPostfix(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ops) != 3 {
		t.Fatalf("expected 3 ops, got %d", len(ops))
	}
	if ops[0].Kind != "push" || ops[0].Name != "a" {
		t.Errorf("expected first op push Name=a, got %+v", ops[0])
	}
	if ops[1].Kind != "push" || ops[1].Value != 5.0 {
		t.Errorf("expected second op push Value=5.0, got %+v", ops[1])
	}
	if ops[2].Kind != "apply" || ops[2].Op != operator.GreaterThan {
		t.Errorf("expected third op apply Op=>, got %+v", ops[2])
	}
}

func TestWalkToPostfix_LogicalOp(t *testing.T) {
	node := &CILLogicalOp{
		Op: operator.And,
		Left: &CILComparisonOp{
			Op: operator.GreaterThan,
			Left:  &CILInputRef{Name: "a"},
			Right: &CILLiteral{Value: 0.0},
		},
		Right: &CILComparisonOp{
			Op: operator.GreaterThan,
			Left:  &CILInputRef{Name: "b"},
			Right: &CILLiteral{Value: 0.0},
		},
	}
	ops, err := WalkCILTreeToPostfix(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ops) != 7 {
		t.Fatalf("expected 7 ops, got %d", len(ops))
	}
	expected := []calculation.PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "push", Value: 0.0},
		{Kind: "apply", Op: operator.GreaterThan},
		{Kind: "push", Name: "b"},
		{Kind: "push", Value: 0.0},
		{Kind: "apply", Op: operator.GreaterThan},
		{Kind: "apply", Op: operator.And},
	}
	for i, exp := range expected {
		if !postfixOpEqual(ops[i], exp) {
			t.Errorf("op[%d] expected %+v, got %+v", i, exp, ops[i])
		}
	}
}

func TestWalkToPostfix_NotOp(t *testing.T) {
	node := &CILNotOp{
		Operand: &CILComparisonOp{
			Op: operator.GreaterThan,
			Left:  &CILInputRef{Name: "a"},
			Right: &CILLiteral{Value: 0.0},
		},
	}
	ops, err := WalkCILTreeToPostfix(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ops) != 4 {
		t.Fatalf("expected 4 ops, got %d", len(ops))
	}
	if ops[3].Kind != "apply" || ops[3].Op != operator.Not {
		t.Errorf("expected last op apply Op=not, got %+v", ops[3])
	}
}

func TestWalkToPostfix_BoolLiteralTrue(t *testing.T) {
	node := &CILBoolLiteral{Value: true}
	ops, err := WalkCILTreeToPostfix(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("expected 1 op, got %d", len(ops))
	}
	op := ops[0]
	if op.Kind != "push" || op.Op != operator.Bool || op.Value != 1.0 {
		t.Errorf("expected push{Op:bool, Value:1.0}, got %+v", op)
	}
}

func TestWalkToPostfix_BoolLiteralFalse(t *testing.T) {
	node := &CILBoolLiteral{Value: false}
	ops, err := WalkCILTreeToPostfix(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("expected 1 op, got %d", len(ops))
	}
	op := ops[0]
	if op.Kind != "push" || op.Op != operator.Bool || op.Value != 0.0 {
		t.Errorf("expected push{Op:bool, Value:0.0}, got %+v", op)
	}
}

func TestWalkToPostfix_NilLiteral(t *testing.T) {
	node := &CILNilLiteral{}
	ops, err := WalkCILTreeToPostfix(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("expected 1 op, got %d", len(ops))
	}
	op := ops[0]
	if op.Kind != "push" || op.Op != operator.Nil || op.Value != 0.0 {
		t.Errorf("expected push{Op:nil, Value:0.0}, got %+v", op)
	}
}

func TestWalkToPostfix_AllComparisonOperators(t *testing.T) {
	operators := []operator.Operator{operator.Equal, operator.NotEqual, operator.LessThan, operator.GreaterThan, operator.LessThanOrEqual, operator.GreaterThanOrEqual}
	for _, op := range operators {
		node := &CILComparisonOp{
			Op:    op,
			Left:  &CILInputRef{Name: "a"},
			Right: &CILLiteral{Value: 1.0},
		}
		ops, err := WalkCILTreeToPostfix(node)
		if err != nil {
			t.Fatalf("operator %s: unexpected error: %v", op, err)
		}
		if len(ops) != 3 {
			t.Fatalf("operator %s: expected 3 ops, got %d", op, len(ops))
		}
		if ops[2].Kind != "apply" || ops[2].Op != op {
			t.Errorf("operator %s: expected apply Op=%s, got %+v", op, op, ops[2])
		}
	}
}

func TestWalkToPostfix_AllLogicalOperators(t *testing.T) {
	operators := []operator.Operator{operator.And, operator.Or}
	for _, op := range operators {
		node := &CILLogicalOp{
			Op:    op,
			Left:  &CILBoolLiteral{Value: true},
			Right: &CILBoolLiteral{Value: false},
		}
		ops, err := WalkCILTreeToPostfix(node)
		if err != nil {
			t.Fatalf("operator %s: unexpected error: %v", op, err)
		}
		if len(ops) != 3 {
			t.Fatalf("operator %s: expected 3 ops, got %d", op, len(ops))
		}
		if ops[2].Kind != "apply" || ops[2].Op != op {
			t.Errorf("operator %s: expected apply Op=%s, got %+v", op, op, ops[2])
		}
	}
}

func TestWalkToPostfix_UnsupportedNodeType(t *testing.T) {
	// mockCILNode is not supported by walkToPostfix
	node := &mockCILNode{}
	_, err := WalkCILTreeToPostfix(node)
	if err == nil {
		t.Fatal("expected error for unsupported node type, got nil")
	}
}

// TestWalkToPostfix_SimpleConditional verifies postfix table for if/else.
func TestWalkToPostfix_SimpleConditional(t *testing.T) {
	node := &CILConditional{
		Branches: []CILConditionalBranch{
			{
				Condition: &CILComparisonOp{Op: operator.GreaterThan, Left: &CILInputRef{Name: "a"}, Right: &CILLiteral{Value: 0.0}},
				Body:      &CILLiteral{Value: 1.0},
			},
			{
				Condition: nil,
				Body:      &CILLiteral{Value: 2.0},
			},
		},
	}
	ops, err := WalkCILTreeToPostfix(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("expected 1 op, got %d", len(ops))
	}
	op := ops[0]
	if op.Kind != "conditional" {
		t.Fatalf("expected conditional op, got %q", op.Kind)
	}
	if len(op.Branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(op.Branches))
	}
	// Branch 0 condition: push a, push 0, apply >
	if len(op.Branches[0].Condition) != 3 {
		t.Errorf("branch 0 condition expected 3 ops, got %d", len(op.Branches[0].Condition))
	}
	if len(op.Branches[0].Body) != 1 || op.Branches[0].Body[0].Kind != "push" || op.Branches[0].Body[0].Value != 1.0 {
		t.Errorf("branch 0 body expected push 1.0, got %+v", op.Branches[0].Body)
	}
	// Branch 1 condition: nil (else)
	if op.Branches[1].Condition != nil {
		t.Errorf("branch 1 condition expected nil, got %+v", op.Branches[1].Condition)
	}
	if len(op.Branches[1].Body) != 1 || op.Branches[1].Body[0].Kind != "push" || op.Branches[1].Body[0].Value != 2.0 {
		t.Errorf("branch 1 body expected push 2.0, got %+v", op.Branches[1].Body)
	}
}

// TestWalkToPostfix_ConditionalWithElseif verifies three-branch postfix.
func TestWalkToPostfix_ConditionalWithElseif(t *testing.T) {
	node := &CILConditional{
		Branches: []CILConditionalBranch{
			{
				Condition: &CILComparisonOp{Op: operator.GreaterThan, Left: &CILInputRef{Name: "a"}, Right: &CILLiteral{Value: 0.0}},
				Body:      &CILLiteral{Value: 1.0},
			},
			{
				Condition: &CILComparisonOp{Op: operator.LessThan, Left: &CILInputRef{Name: "a"}, Right: &CILLiteral{Value: 0.0}},
				Body:      &CILLiteral{Value: -1.0},
			},
			{
				Condition: nil,
				Body:      &CILLiteral{Value: 0.0},
			},
		},
	}
	ops, err := WalkCILTreeToPostfix(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ops) != 1 || ops[0].Kind != "conditional" {
		t.Fatalf("expected single conditional op, got %+v", ops)
	}
	if len(ops[0].Branches) != 3 {
		t.Fatalf("expected 3 branches, got %d", len(ops[0].Branches))
	}
	// Verify each branch has a body and the first two have conditions.
	for i := 0; i < 2; i++ {
		if len(ops[0].Branches[i].Condition) == 0 {
			t.Errorf("branch %d condition expected non-empty", i)
		}
	}
	if ops[0].Branches[2].Condition != nil {
		t.Errorf("branch 2 (else) condition expected nil")
	}
}

// TestWalkToPostfix_WhenClause verifies when/default postfix.
func TestWalkToPostfix_WhenClause(t *testing.T) {
	node := &CILWhenClause{
		Clauses: []CILWhenCase{
			{
				Condition: &CILComparisonOp{Op: operator.GreaterThan, Left: &CILInputRef{Name: "a"}, Right: &CILLiteral{Value: 0.0}},
				Body:      &CILLiteral{Value: 1.0},
			},
			{
				Condition: &CILComparisonOp{Op: operator.LessThan, Left: &CILInputRef{Name: "a"}, Right: &CILLiteral{Value: 0.0}},
				Body:      &CILLiteral{Value: -1.0},
			},
		},
		Default: &CILLiteral{Value: 0.0},
	}
	ops, err := WalkCILTreeToPostfix(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ops) != 1 || ops[0].Kind != "conditional" {
		t.Fatalf("expected single conditional op, got %+v", ops)
	}
	if len(ops[0].Branches) != 3 {
		t.Fatalf("expected 3 branches (2 cases + default), got %d", len(ops[0].Branches))
	}
	// First two have conditions, last (default) has nil condition.
	for i := 0; i < 2; i++ {
		if len(ops[0].Branches[i].Condition) == 0 {
			t.Errorf("branch %d condition expected non-empty", i)
		}
	}
	if ops[0].Branches[2].Condition != nil {
		t.Errorf("default branch condition expected nil")
	}
}

// TestWalkToPostfix_WhenClauseNoDefault verifies when without default.
func TestWalkToPostfix_WhenClauseNoDefault(t *testing.T) {
	node := &CILWhenClause{
		Clauses: []CILWhenCase{
			{
				Condition: &CILComparisonOp{Op: operator.GreaterThan, Left: &CILInputRef{Name: "a"}, Right: &CILLiteral{Value: 0.0}},
				Body:      &CILLiteral{Value: 1.0},
			},
		},
		Default: nil,
	}
	ops, err := WalkCILTreeToPostfix(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ops) != 1 || ops[0].Kind != "conditional" {
		t.Fatalf("expected single conditional op, got %+v", ops)
	}
	if len(ops[0].Branches) != 1 {
		t.Fatalf("expected 1 branch, got %d", len(ops[0].Branches))
	}
	if len(ops[0].Branches[0].Condition) == 0 {
		t.Errorf("branch 0 condition expected non-empty")
	}
}

// TestWalkToPostfix_NestedConditional verifies nested conditional produces correct postfix.
func TestWalkToPostfix_NestedConditional(t *testing.T) {
	node := &CILConditional{
		Branches: []CILConditionalBranch{
			{
				Condition: &CILComparisonOp{Op: operator.GreaterThan, Left: &CILInputRef{Name: "a"}, Right: &CILLiteral{Value: 0.0}},
				Body: &CILConditional{
					Branches: []CILConditionalBranch{
						{
							Condition: &CILComparisonOp{Op: operator.LessThan, Left: &CILInputRef{Name: "b"}, Right: &CILLiteral{Value: 0.0}},
							Body:      &CILLiteral{Value: 1.0},
						},
						{
							Condition: nil,
							Body:      &CILLiteral{Value: 2.0},
						},
					},
				},
			},
			{
				Condition: nil,
				Body:      &CILLiteral{Value: 3.0},
			},
		},
	}
	ops, err := WalkCILTreeToPostfix(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ops) != 1 || ops[0].Kind != "conditional" {
		t.Fatalf("expected single conditional op, got %+v", ops)
	}
	outer := ops[0]
	if len(outer.Branches) != 2 {
		t.Fatalf("expected 2 outer branches, got %d", len(outer.Branches))
	}
	// Outer branch 0 body should be a nested conditional.
	inner := outer.Branches[0].Body
	if len(inner) != 1 || inner[0].Kind != "conditional" {
		t.Fatalf("expected nested conditional in outer branch 0 body, got %+v", inner)
	}
	if len(inner[0].Branches) != 2 {
		t.Fatalf("expected 2 inner branches, got %d", len(inner[0].Branches))
	}
	if inner[0].Branches[1].Condition != nil {
		t.Errorf("inner else branch condition expected nil")
	}
	// Outer else branch body should be push 3.0.
	if len(outer.Branches[1].Body) != 1 || outer.Branches[1].Body[0].Value != 3.0 {
		t.Errorf("outer else body expected push 3.0, got %+v", outer.Branches[1].Body)
	}
}

// TestTransformMetric_CILConditionalRoot verifies TransformMetric with CILConditional root produces valid Go code containing EvaluatePostfixFloat64.
func TestTransformMetric_CILConditionalRoot(t *testing.T) {
	transformer := NewCILAbstractTransformer()
	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}

	root := &CILMetricNode{
		CILNode: &CILConditional{
			Branches: []CILConditionalBranch{
				{
					Condition: &CILComparisonOp{Op: operator.GreaterThan, Left: &CILInputRef{Name: "a"}, Right: &CILLiteral{Value: 0.0}},
					Body:      &CILLiteral{Value: 1.0},
				},
				{
					Condition: nil,
					Body:      &CILLiteral{Value: 2.0},
				},
			},
		},
		InputRefs: []CILInputRef{
			{Kind: "connector", Name: "a"},
		},
	}

	code, err := transformer.TransformMetric("ConditionalTest", root)
	if err != nil {
		t.Fatalf("TransformMetric() returned unexpected error: %v", err)
	}

	expectedContains := []string{
		"func computeConditionalTest(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
		// Note: "_ = inputs" is NOT present because usesInputs=true (connector ref)
		"[]calculation.PostfixOp",
		`Kind: "conditional"`,
		"EvaluatePostfixFloat64",
		"controller.GetMetricComputation(\"ConditionalTest\").SetTensor(result)",
		// Verify branches are serialized in the generated code
		"Branches:",
		"Condition:",
		"Body:",
	}

	for _, expected := range expectedContains {
		if !strings.Contains(code, expected) {
			t.Errorf("TransformMetric() output missing expected content: %q\nGot:\n%s", expected, code)
		}
	}

	expectedNotContains := []string{
		"data := result.Data()",
		"for i := range data",
		"calculation.AllocateTensor",
	}

	for _, notExpected := range expectedNotContains {
		if strings.Contains(code, notExpected) {
			t.Errorf("TransformMetric() output contains unexpected content: %q\nGot:\n%s", notExpected, code)
		}
	}
}

// TestTransformMetric_CILWhenClauseRoot verifies TransformMetric with CILWhenClause root produces valid Go code.
func TestTransformMetric_CILWhenClauseRoot(t *testing.T) {
	transformer := NewCILAbstractTransformer()
	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}

	root := &CILMetricNode{
		CILNode: &CILWhenClause{
			Clauses: []CILWhenCase{
				{
					Condition: &CILComparisonOp{Op: operator.GreaterThan, Left: &CILInputRef{Name: "a"}, Right: &CILLiteral{Value: 0.0}},
					Body:      &CILLiteral{Value: 1.0},
				},
				{
					Condition: &CILComparisonOp{Op: operator.LessThan, Left: &CILInputRef{Name: "a"}, Right: &CILLiteral{Value: 0.0}},
					Body:      &CILLiteral{Value: -1.0},
				},
			},
			Default: &CILLiteral{Value: 0.0},
		},
		InputRefs: []CILInputRef{
			{Kind: "connector", Name: "a"},
		},
	}

	code, err := transformer.TransformMetric("WhenClauseTest", root)
	if err != nil {
		t.Fatalf("TransformMetric() returned unexpected error: %v", err)
	}

	expectedContains := []string{
		"func computeWhenClauseTest(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
		// Note: "_ = inputs" is NOT present because usesInputs=true (connector ref)
		"[]calculation.PostfixOp",
		`Kind: "conditional"`,
		"EvaluatePostfixFloat64",
		"controller.GetMetricComputation(\"WhenClauseTest\").SetTensor(result)",
		// Verify branches are serialized in the generated code
		"Branches:",
		"Condition:",
		"Body:",
	}

	for _, expected := range expectedContains {
		if !strings.Contains(code, expected) {
			t.Errorf("TransformMetric() output missing expected content: %q\nGot:\n%s", expected, code)
		}
	}

	expectedNotContains := []string{
		"data := result.Data()",
		"for i := range data",
		"calculation.AllocateTensor",
	}

	for _, notExpected := range expectedNotContains {
		if strings.Contains(code, notExpected) {
			t.Errorf("TransformMetric() output contains unexpected content: %q\nGot:\n%s", notExpected, code)
		}
	}
}

// TestGeneratePostfixTableLiteral_WithBranches verifies that GeneratePostfixTableLiteral
// correctly serializes the Branches field for conditional postfix ops.
func TestGeneratePostfixTableLiteral_WithBranches(t *testing.T) {
	transformer := NewCILAbstractTransformer()
	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}

	ops := []calculation.PostfixOp{
		{
			Kind: "conditional",
			Branches: []calculation.PostfixBranch{
				{
					Condition: []calculation.PostfixOp{
						{Kind: "push", Name: "a"},
						{Kind: "push", Value: 0.0},
						{Kind: "apply", Op: operator.GreaterThan},
					},
					Body: []calculation.PostfixOp{
						{Kind: "push", Value: 1.0},
					},
				},
				{
					Condition: nil,
					Body: []calculation.PostfixOp{
						{Kind: "push", Value: 2.0},
					},
				},
			},
		},
	}

	literal := transformer.GeneratePostfixTableLiteral(ops)

	expectedContains := []string{
		`Kind: "conditional"`,
		"Branches:",
		"Condition:",
		"Body:",
		`Kind: "push", Name: "a"`,
		`Kind: "apply", Op: operator.GreaterThan`,
		`Kind: "push", Value: 1`,
		`Kind: "push", Value: 2`,
	}

	for _, expected := range expectedContains {
		if !strings.Contains(literal, expected) {
			t.Errorf("GeneratePostfixTableLiteral() output missing expected content: %q\nGot:\n%s", expected, literal)
		}
	}
}

// TestGeneratePostfixTableLiteral_NestedBranches verifies that GeneratePostfixTableLiteral
// correctly serializes nested conditional branches.
func TestGeneratePostfixTableLiteral_NestedBranches(t *testing.T) {
	transformer := NewCILAbstractTransformer()
	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}

	ops := []calculation.PostfixOp{
		{
			Kind: "conditional",
			Branches: []calculation.PostfixBranch{
				{
					Condition: []calculation.PostfixOp{
						{Kind: "push", Name: "a"},
						{Kind: "push", Value: 0.0},
						{Kind: "apply", Op: operator.GreaterThan},
					},
					Body: []calculation.PostfixOp{
						{
							Kind: "conditional",
							Branches: []calculation.PostfixBranch{
								{
									Condition: []calculation.PostfixOp{
										{Kind: "push", Name: "b"},
										{Kind: "push", Value: 0.0},
										{Kind: "apply", Op: operator.LessThan},
									},
									Body: []calculation.PostfixOp{
										{Kind: "push", Value: 1.0},
									},
								},
								{
									Condition: nil,
									Body: []calculation.PostfixOp{
										{Kind: "push", Value: 2.0},
									},
								},
							},
						},
					},
				},
				{
					Condition: nil,
					Body: []calculation.PostfixOp{
						{Kind: "push", Value: 3.0},
					},
				},
			},
		},
	}

	literal := transformer.GeneratePostfixTableLiteral(ops)

	// Count occurrences of "conditional" to verify nested serialization
	count := strings.Count(literal, `Kind: "conditional"`)
	if count != 2 {
		t.Errorf("expected 2 conditional ops in literal, got %d\nLiteral:\n%s", count, literal)
	}

	// Verify all branch bodies are present
	for _, expected := range []string{"Value: 1", "Value: 2", "Value: 3"} {
		if !strings.Contains(literal, expected) {
			t.Errorf("GeneratePostfixTableLiteral() output missing expected content: %q\nGot:\n%s", expected, literal)
		}
	}
}
