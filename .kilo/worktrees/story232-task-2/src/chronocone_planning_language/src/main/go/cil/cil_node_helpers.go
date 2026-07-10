// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

// HasNonLiteralCILNode returns true if the CIL expression tree rooted at node
// contains any node other than CILLiteral, CILInputRef, CILBoolLiteral, or CILNilLiteral.
// This is used during synthesis to determine whether the generated Go code will
// reference operator constants (e.g., operator.Add) and therefore needs the
// operator package imported.
func HasNonLiteralCILNode(node CILNode) bool {
	if node == nil {
		return false
	}
	switch n := node.(type) {
	case *CILLiteral, *CILInputRef, *CILBoolLiteral, *CILNilLiteral:
		return false
	case *CILBinaryOp:
		return true
	case *CILUnaryOp:
		return true
	case *CILComparisonOp:
		return true
	case *CILLogicalOp:
		return true
	case *CILNotOp:
		// CILNotOp always emits operator.Not in the postfix table,
		// so it always requires the operator package import.
		return true
	case *CILConditional:
		for _, b := range n.Branches {
			if b.Condition != nil && HasNonLiteralCILNode(b.Condition) {
				return true
			}
			if HasNonLiteralCILNode(b.Body) {
				return true
			}
		}
		return false
	case *CILWhenClause:
		for _, c := range n.Clauses {
			if HasNonLiteralCILNode(c.Condition) {
				return true
			}
			if HasNonLiteralCILNode(c.Body) {
				return true
			}
		}
		if n.Default != nil && HasNonLiteralCILNode(n.Default) {
			return true
		}
		return false
	default:
		return false
	}
}
