// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// TestCILComparisonOpConstruction verifies that each valid comparison operator
// can be set on a CILComparisonOp and read back, with Left/Right child nodes.
func TestCILComparisonOpConstruction(t *testing.T) {
	tests := []struct {
		name string
		op   operator.Operator
	}{
		{"equality", operator.Equal},
		{"inequality", operator.NotEqual},
		{"less than", operator.LessThan},
		{"greater than", operator.GreaterThan},
		{"less than or equal", operator.LessThanOrEqual},
		{"greater than or equal", operator.GreaterThanOrEqual},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &CILComparisonOp{
				Op:    tt.op,
				Left:  &CILLiteral{Value: 1.0},
				Right: &CILLiteral{Value: 2.0},
			}

			if node.Op != tt.op {
				t.Errorf("CILComparisonOp.Op = %v, expected %v", node.Op, tt.op)
			}
			if node.Left == nil {
				t.Errorf("CILComparisonOp.Left is nil, expected non-nil")
			}
			if node.Right == nil {
				t.Errorf("CILComparisonOp.Right is nil, expected non-nil")
			}
		})
	}
}

// TestCILLogicalOpConstruction verifies that logical operators And and Or
// can be set on a CILLogicalOp with Left/Right child nodes.
func TestCILLogicalOpConstruction(t *testing.T) {
	tests := []struct {
		name string
		op   operator.Operator
	}{
		{"logical and", operator.And},
		{"logical or", operator.Or},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &CILLogicalOp{
				Op:    tt.op,
				Left:  &CILLiteral{Value: 1.0},
				Right: &CILLiteral{Value: 2.0},
			}

			if node.Op != tt.op {
				t.Errorf("CILLogicalOp.Op = %v, expected %v", node.Op, tt.op)
			}
			if node.Left == nil {
				t.Errorf("CILLogicalOp.Left is nil, expected non-nil")
			}
			if node.Right == nil {
				t.Errorf("CILLogicalOp.Right is nil, expected non-nil")
			}
		})
	}
}

// TestCILNotOpConstruction verifies that the Operand field of a CILNotOp is
// set correctly.
func TestCILNotOpConstruction(t *testing.T) {
	node := &CILNotOp{
		Operand: &CILLiteral{Value: 42.0},
	}

	if node.Operand == nil {
		t.Errorf("CILNotOp.Operand is nil, expected non-nil")
	}
}

// TestCILNilLiteralConstruction verifies that a CILNilLiteral can be created.
func TestCILNilLiteralConstruction(t *testing.T) {
	node := CILNilLiteral{}
	_ = node
}

// TestCILBoolLiteralConstruction verifies that true and false values can be set
// on a CILBoolLiteral.
func TestCILBoolLiteralConstruction(t *testing.T) {
	tests := []struct {
		name  string
		value bool
	}{
		{"true value", true},
		{"false value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &CILBoolLiteral{Value: tt.value}

			if node.Value != tt.value {
				t.Errorf("CILBoolLiteral.Value = %v, expected %v", node.Value, tt.value)
			}
		})
	}
}

// TestCILConditionalConstruction verifies that a CILConditional can be created
// with one or more branches.
func TestCILConditionalConstruction(t *testing.T) {
	node := &CILConditional{
		Branches: []CILConditionalBranch{
			{
				Condition: &CILLiteral{Value: 1.0},
				Body:      &CILLiteral{Value: 2.0},
			},
			{
				Condition: &CILLiteral{Value: 3.0},
				Body:      &CILLiteral{Value: 4.0},
			},
		},
	}

	if len(node.Branches) != 2 {
		t.Errorf("CILConditional.Branches length = %d, expected 2", len(node.Branches))
	}
	if node.Branches[0].Condition == nil {
		t.Errorf("CILConditional.Branches[0].Condition is nil, expected non-nil")
	}
	if node.Branches[1].Body == nil {
		t.Errorf("CILConditional.Branches[1].Body is nil, expected non-nil")
	}
}

// TestCILConditionalBranchConstruction verifies that the helper struct fields
// are independently settable.
func TestCILConditionalBranchConstruction(t *testing.T) {
	branch := CILConditionalBranch{
		Condition: &CILLiteral{Value: 1.0},
		Body:      &CILLiteral{Value: 2.0},
	}

	if branch.Condition == nil {
		t.Errorf("CILConditionalBranch.Condition is nil, expected non-nil")
	}
	if branch.Body == nil {
		t.Errorf("CILConditionalBranch.Body is nil, expected non-nil")
	}
}

// TestCILWhenClauseConstruction verifies that a CILWhenClause can be created
// with Clauses and a Default body.
func TestCILWhenClauseConstruction(t *testing.T) {
	node := &CILWhenClause{
		Clauses: []CILWhenCase{
			{
				Condition: &CILLiteral{Value: 1.0},
				Body:      &CILLiteral{Value: 2.0},
			},
		},
		Default: &CILLiteral{Value: 0.0},
	}

	if len(node.Clauses) != 1 {
		t.Errorf("CILWhenClause.Clauses length = %d, expected 1", len(node.Clauses))
	}
	if node.Default == nil {
		t.Errorf("CILWhenClause.Default is nil, expected non-nil")
	}
}

// TestCILWhenCaseConstruction verifies that the helper struct fields are
// independently settable.
func TestCILWhenCaseConstruction(t *testing.T) {
	wc := CILWhenCase{
		Condition: &CILLiteral{Value: 1.0},
		Body:      &CILLiteral{Value: 2.0},
	}

	if wc.Condition == nil {
		t.Errorf("CILWhenCase.Condition is nil, expected non-nil")
	}
	if wc.Body == nil {
		t.Errorf("CILWhenCase.Body is nil, expected non-nil")
	}
}

// TestCILConditionalNodeTypesImplementCILNodeInterface is a compile-time
// interface satisfaction check. It will FAIL TO COMPILE until Phase C adds
// the cilNode() marker methods to each conditional node type — this is the
// correct TDD red phase.
func TestCILConditionalNodeTypesImplementCILNodeInterface(t *testing.T) {
	var _ CILNode = (*CILComparisonOp)(nil)
	var _ CILNode = (*CILLogicalOp)(nil)
	var _ CILNode = (*CILNotOp)(nil)
	var _ CILNode = (*CILNilLiteral)(nil)
	var _ CILNode = (*CILBoolLiteral)(nil)
	var _ CILNode = (*CILConditional)(nil)
	var _ CILNode = (*CILWhenClause)(nil)
}

// TestCILConditionalBranchNotCILNode documents that CILConditionalBranch and
// CILWhenCase are helper structs that intentionally do NOT implement the
// CILNode interface.
func TestCILConditionalBranchNotCILNode(t *testing.T) {
	// CILConditionalBranch and CILWhenCase are helper structs used within
	// CILConditional and CILWhenClause respectively. They intentionally do
	// NOT implement the CILNode interface.
	//
	// The following lines would fail to compile because these types do not
	// have a cilNode() marker method:
	//   var _ CILNode = (*CILConditionalBranch)(nil)
	//   var _ CILNode = (*CILWhenCase)(nil)

	branch := CILConditionalBranch{
		Condition: &CILLiteral{Value: 1.0},
		Body:      &CILLiteral{Value: 2.0},
	}
	if branch.Condition == nil || branch.Body == nil {
		t.Errorf("CILConditionalBranch fields should be non-nil")
	}

	wc := CILWhenCase{
		Condition: &CILLiteral{Value: 1.0},
		Body:      &CILLiteral{Value: 2.0},
	}
	if wc.Condition == nil || wc.Body == nil {
		t.Errorf("CILWhenCase fields should be non-nil")
	}
}
