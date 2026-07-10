// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import "github.com/chronocone/chronocone_planning_language/src/main/go/operator"

// CILComparisonOp represents a comparison operation in the CIL expression tree.
// Supported operators: Equal, NotEqual, LessThan, GreaterThan, LessThanOrEqual, GreaterThanOrEqual.
type CILComparisonOp struct {
	Op    operator.Operator // The operator
	Left  CILNode           // The left operand
	Right CILNode           // The right operand
}

// CILLogicalOp represents a logical operation in the CIL expression tree.
// Supported operators: And, Or.
type CILLogicalOp struct {
	Op    operator.Operator // The operator
	Left  CILNode           // The left operand
	Right CILNode           // The right operand
}

// CILNotOp represents a logical negation operation in the CIL expression tree.
type CILNotOp struct {
	Operand CILNode // The operand to negate
}

// CILNilLiteral represents a nil (null) literal value in the CIL expression tree.
// The value is always nil.
type CILNilLiteral struct{}

// CILBoolLiteral represents a boolean literal value in the CIL expression tree.
type CILBoolLiteral struct {
	Value bool // The boolean value
}

// CILConditional represents a conditional (if/elseif/else) expression in the CIL expression tree.
type CILConditional struct {
	Branches []CILConditionalBranch // Ordered list of conditional branches
}

// CILConditionalBranch represents a single branch of a CILConditional.
// It is a helper struct and does not itself implement CILNode.
type CILConditionalBranch struct {
	Condition CILNode // The condition to evaluate
	Body      CILNode // The body to execute if the condition is true
}

// CILWhenClause represents a when/case expression in the CIL expression tree.
type CILWhenClause struct {
	Clauses []CILWhenCase // Ordered list of when cases
	Default CILNode       // The default body if no case matches
}

// CILWhenCase represents a single case within a CILWhenClause.
// It is a helper struct and does not itself implement CILNode.
type CILWhenCase struct {
	Condition CILNode // The condition to evaluate
	Body      CILNode // The body to execute if the condition is true
}

// cilNode marks CILComparisonOp as implementing the CILNode interface.
func (c *CILComparisonOp) cilNode() {}

// cilNode marks CILLogicalOp as implementing the CILNode interface.
func (c *CILLogicalOp) cilNode() {}

// cilNode marks CILNotOp as implementing the CILNode interface.
func (c *CILNotOp) cilNode() {}

// cilNode marks CILNilLiteral as implementing the CILNode interface.
func (c *CILNilLiteral) cilNode() {}

// cilNode marks CILBoolLiteral as implementing the CILNode interface.
func (c *CILBoolLiteral) cilNode() {}

// cilNode marks CILConditional as implementing the CILNode interface.
func (c *CILConditional) cilNode() {}

// cilNode marks CILWhenClause as implementing the CILNode interface.
func (c *CILWhenClause) cilNode() {}
