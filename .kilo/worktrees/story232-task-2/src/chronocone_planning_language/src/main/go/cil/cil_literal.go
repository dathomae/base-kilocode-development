// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

// CILLiteral represents a literal value in the CIL expression tree.
type CILLiteral struct {
	Value float64 // The numeric value of the literal
}

// cilNode marks CILLiteral as implementing the CILNode interface.
func (c *CILLiteral) cilNode() {}
