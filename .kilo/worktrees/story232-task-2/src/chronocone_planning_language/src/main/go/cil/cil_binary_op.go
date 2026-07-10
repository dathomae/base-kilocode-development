// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import "github.com/chronocone/chronocone_planning_language/src/main/go/operator"

// CILBinaryOp represents a binary operation in the CIL expression tree.
// Supported operators: Add, Subtract, Multiply, Divide.
type CILBinaryOp struct {
	Op            operator.Operator // The operator
	Left          CILNode           // The left operand
	Right         CILNode           // The right operand
	DerivedUnit   string            // Unit derived from operand units
	AnnotatedUnit string            // Unit explicitly annotated by the user
}

// cilNode marks CILBinaryOp as implementing the CILNode interface.
func (c *CILBinaryOp) cilNode() {}
