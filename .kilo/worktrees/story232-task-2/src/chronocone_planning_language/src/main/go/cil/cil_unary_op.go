// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import "github.com/chronocone/chronocone_planning_language/src/main/go/operator"

// CILUnaryOp represents a unary operation in the CIL expression tree.
// Currently only supports negation (Negate).
type CILUnaryOp struct {
	Op      operator.Operator // The operator
	Operand CILNode           // The operand to negate
}

// cilNode marks CILUnaryOp as implementing the CILNode interface.
func (c *CILUnaryOp) cilNode() {}
