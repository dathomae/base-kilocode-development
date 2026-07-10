// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

var _ CILNode = (*CILBinaryOp)(nil)

// TestCILBinaryOpConstruction verifies all fields are stored correctly.
func TestCILBinaryOpConstruction(t *testing.T) {
	left := &CILNumber{Value: 10.0, AnnotatedUnit: "m"}
	right := &CILNumber{Value: 2.0, AnnotatedUnit: "s"}
	node := &CILBinaryOp{
		Op:            operator.Divide,
		Left:          left,
		Right:         right,
		DerivedUnit:   "m/s",
		AnnotatedUnit: "",
	}

	if node.Op != operator.Divide {
		t.Errorf("CILBinaryOp.Op = %v, expected operator.Divide", node.Op)
	}
	if node.Left == nil {
		t.Errorf("CILBinaryOp.Left is nil, expected non-nil")
	}
	if node.Right == nil {
		t.Errorf("CILBinaryOp.Right is nil, expected non-nil")
	}
	if node.DerivedUnit != "m/s" {
		t.Errorf("CILBinaryOp.DerivedUnit = %q, expected %q", node.DerivedUnit, "m/s")
	}
	if node.AnnotatedUnit != "" {
		t.Errorf("CILBinaryOp.AnnotatedUnit = %q, expected %q", node.AnnotatedUnit, "")
	}
}

// TestCILBinaryOpNilOperands verifies that a CILBinaryOp with nil
// Left and Right does not panic on construction or field access.
func TestCILBinaryOpNilOperands(t *testing.T) {
	node := &CILBinaryOp{
		Op:            operator.Add,
		Left:          nil,
		Right:         nil,
		DerivedUnit:   "",
		AnnotatedUnit: "",
	}

	if node.Left != nil {
		t.Errorf("CILBinaryOp.Left expected nil, got non-nil")
	}
	if node.Right != nil {
		t.Errorf("CILBinaryOp.Right expected nil, got non-nil")
	}
	if node.Op != operator.Add {
		t.Errorf("CILBinaryOp.Op = %v, expected operator.Add", node.Op)
	}
}

// TestOperatorValues verifies that each operator.Operator constant has the
// expected value.
func TestOperatorValues(t *testing.T) {
	tests := []struct {
		op   operator.Operator
		want int
		name string
	}{
		{operator.Unknown, 0, "Unknown"},
		{operator.Add, 1, "Add"},
		{operator.Subtract, 2, "Subtract"},
		{operator.Multiply, 3, "Multiply"},
		{operator.Divide, 4, "Divide"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.op) != tt.want {
				t.Errorf("%s = %d, expected %d", tt.name, int(tt.op), tt.want)
			}
		})
	}
}

// TestNestedCILNodeConstruction verifies that CIL node nodes can
// be nested arbitrarily.
func TestNestedCILNodeConstruction(t *testing.T) {
	inner := &CILBinaryOp{
		Op:            operator.Multiply,
		Left:          &CILUnitRef{UnitGroup: "Length"},
		Right:         &CILUnitRef{UnitGroup: "Time"},
		DerivedUnit:   "Length*Time",
		AnnotatedUnit: "",
	}
	outer := &CILBinaryOp{
		Op:            operator.Divide,
		Left:          inner,
		Right:         &CILNumber{Value: 2.0, AnnotatedUnit: "s"},
		DerivedUnit:   "Length*Time/s",
		AnnotatedUnit: "",
	}

	if outer.Left == nil {
		t.Fatalf("outer.Left is nil, expected non-nil")
	}
	leftExpr, ok := outer.Left.(*CILBinaryOp)
	if !ok {
		t.Fatalf("outer.Left type = %T, expected *CILBinaryOp", outer.Left)
	}
	if leftExpr.Op != operator.Multiply {
		t.Errorf("outer.Left.Op = %v, expected operator.Multiply", leftExpr.Op)
	}
	if outer.Right == nil {
		t.Errorf("outer.Right is nil, expected non-nil")
	}
}
