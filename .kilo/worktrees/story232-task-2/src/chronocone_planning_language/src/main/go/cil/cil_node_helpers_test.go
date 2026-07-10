// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

func TestHasNonLiteralCILNode_Nil(t *testing.T) {
	if HasNonLiteralCILNode(nil) {
		t.Error("HasNonLiteralCILNode(nil) = true, want false")
	}
}

func TestHasNonLiteralCILNode_Literals(t *testing.T) {
	tests := []struct {
		name string
		node CILNode
	}{
		{"CILLiteral", &CILLiteral{Value: 42.0}},
		{"CILInputRef", &CILInputRef{Kind: "dimension", Name: "Region"}},
		{"CILBoolLiteral true", &CILBoolLiteral{Value: true}},
		{"CILBoolLiteral false", &CILBoolLiteral{Value: false}},
		{"CILNilLiteral", &CILNilLiteral{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if HasNonLiteralCILNode(tt.node) {
				t.Errorf("HasNonLiteralCILNode(%s) = true, want false", tt.name)
			}
		})
	}
}

func TestHasNonLiteralCILNode_AlwaysTrue(t *testing.T) {
	// These node types always emit operator references in generated code
	// regardless of their operands. HasNonLiteralCILNode must return true
	// unconditionally so the synthesizer imports the operator package.
	tests := []struct {
		name string
		node CILNode
	}{
		{"CILBinaryOp", &CILBinaryOp{Op: operator.Add, Left: &CILLiteral{Value: 1.0}, Right: &CILLiteral{Value: 2.0}}},
		{"CILUnaryOp", &CILUnaryOp{Op: operator.Negate, Operand: &CILLiteral{Value: 5.0}}},
		{"CILComparisonOp", &CILComparisonOp{Op: operator.Equal, Left: &CILLiteral{Value: 1.0}, Right: &CILLiteral{Value: 2.0}}},
		{"CILLogicalOp", &CILLogicalOp{Op: operator.And, Left: &CILBoolLiteral{Value: true}, Right: &CILBoolLiteral{Value: false}}},
		// CILNotOp wrapping a literal: even though the operand is a literal,
		// CILNotOp emits operator.Not in the postfix table, so it must
		// return true to ensure the operator package is imported.
		{"CILNotOp wrapping literal", &CILNotOp{Operand: &CILBoolLiteral{Value: true}}},
		{"CILNotOp wrapping non-literal", &CILNotOp{Operand: &CILComparisonOp{Op: operator.Equal, Left: &CILLiteral{Value: 1.0}, Right: &CILLiteral{Value: 2.0}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !HasNonLiteralCILNode(tt.node) {
				t.Errorf("HasNonLiteralCILNode(%s) = false, want true (would cause missing operator import)", tt.name)
			}
		})
	}
}

func TestHasNonLiteralCILNode_Conditional(t *testing.T) {
	t.Run("all literal branches returns false", func(t *testing.T) {
		node := &CILConditional{
			Branches: []CILConditionalBranch{
				{Condition: &CILBoolLiteral{Value: true}, Body: &CILLiteral{Value: 100.0}},
				{Condition: &CILBoolLiteral{Value: false}, Body: &CILLiteral{Value: 200.0}},
			},
		}
		if HasNonLiteralCILNode(node) {
			t.Error("HasNonLiteralCILNode with all-literal conditional = true, want false")
		}
	})

	t.Run("non-literal condition returns true", func(t *testing.T) {
		node := &CILConditional{
			Branches: []CILConditionalBranch{
				{Condition: &CILComparisonOp{Op: operator.GreaterThan, Left: &CILLiteral{Value: 5.0}, Right: &CILLiteral{Value: 0.0}}, Body: &CILLiteral{Value: 100.0}},
			},
		}
		if !HasNonLiteralCILNode(node) {
			t.Error("HasNonLiteralCILNode with non-literal condition = false, want true")
		}
	})

	t.Run("non-literal body returns true", func(t *testing.T) {
		node := &CILConditional{
			Branches: []CILConditionalBranch{
				{Condition: &CILBoolLiteral{Value: true}, Body: &CILBinaryOp{Op: operator.Add, Left: &CILLiteral{Value: 1.0}, Right: &CILLiteral{Value: 1.0}}},
			},
		}
		if !HasNonLiteralCILNode(node) {
			t.Error("HasNonLiteralCILNode with non-literal body = false, want true")
		}
	})

	t.Run("nil condition is skipped", func(t *testing.T) {
		node := &CILConditional{
			Branches: []CILConditionalBranch{
				{Condition: nil, Body: &CILLiteral{Value: 100.0}},
			},
		}
		if HasNonLiteralCILNode(node) {
			t.Error("HasNonLiteralCILNode with nil condition and literal body = true, want false")
		}
	})
}

func TestHasNonLiteralCILNode_WhenClause(t *testing.T) {
	t.Run("all literal clauses returns false", func(t *testing.T) {
		node := &CILWhenClause{
			Clauses: []CILWhenCase{
				{Condition: &CILBoolLiteral{Value: true}, Body: &CILLiteral{Value: 100.0}},
			},
		}
		if HasNonLiteralCILNode(node) {
			t.Error("HasNonLiteralCILNode with all-literal when clause = true, want false")
		}
	})

	t.Run("non-literal condition returns true", func(t *testing.T) {
		node := &CILWhenClause{
			Clauses: []CILWhenCase{
				{Condition: &CILComparisonOp{Op: operator.Equal, Left: &CILLiteral{Value: 1.0}, Right: &CILLiteral{Value: 2.0}}, Body: &CILLiteral{Value: 100.0}},
			},
		}
		if !HasNonLiteralCILNode(node) {
			t.Error("HasNonLiteralCILNode with non-literal when condition = false, want true")
		}
	})

	t.Run("non-literal body returns true", func(t *testing.T) {
		node := &CILWhenClause{
			Clauses: []CILWhenCase{
				{Condition: &CILBoolLiteral{Value: true}, Body: &CILBinaryOp{Op: operator.Add, Left: &CILLiteral{Value: 1.0}, Right: &CILLiteral{Value: 1.0}}},
			},
		}
		if !HasNonLiteralCILNode(node) {
			t.Error("HasNonLiteralCILNode with non-literal when body = false, want true")
		}
	})

	t.Run("non-literal default returns true", func(t *testing.T) {
		node := &CILWhenClause{
			Clauses: []CILWhenCase{},
			Default: &CILNotOp{Operand: &CILBoolLiteral{Value: true}},
		}
		if !HasNonLiteralCILNode(node) {
			t.Error("HasNonLiteralCILNode with non-literal default = false, want true")
		}
	})

	t.Run("nil default is skipped", func(t *testing.T) {
		node := &CILWhenClause{
			Clauses: []CILWhenCase{},
			Default: nil,
		}
		if HasNonLiteralCILNode(node) {
			t.Error("HasNonLiteralCILNode with empty when clause and nil default = true, want false")
		}
	})
}
