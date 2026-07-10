// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/calculation"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// WalkCILTreeToPostfix walks a CIL expression tree in post-order and returns a
// postfix expression table suitable for stack-based evaluation. This is the main
// entry point for converting a CIL abstract syntax tree into a flat sequence of
// PostfixOp entries.
//
// The function creates a []calculation.PostfixOp slice and delegates to the private
// walkToPostfix helper to perform the recursive traversal.
//
// Parameters:
//   - root: The root CILNode of the expression tree to walk.
//
// Returns:
//   - []calculation.PostfixOp: The postfix expression table representing the tree.
//   - error: An error if the tree cannot be walked, or nil on success.
func WalkCILTreeToPostfix(root CILNode) ([]calculation.PostfixOp, error) {
	if root == nil {
		return nil, fmt.Errorf("root node is nil")
	}
	ops := make([]calculation.PostfixOp, 0)
	if err := walkToPostfix(root, &ops); err != nil {
		return nil, err
	}
	return ops, nil
}

// walkToPostfix is a private recursive helper that traverses a CIL expression tree
// in post-order and appends PostfixOp entries to the provided ops slice.
//
// Node type handling:
//   - *CILLiteral: appends a "push" op with the float64 string representation as Name.
//   - *CILInputRef: appends a "push" op with Name from the input reference.
//   - *CILLookupNode: appends a "push" op with Name from the lookup reference.
//   - *CILBinaryOp: recurses left, recurses right, then appends an "apply" op
//     with the binary operator as Op.
//   - *CILUnaryOp: recurses the operand, then appends an "apply" op with the
//     unary operator as Op.
//   - *CILComparisonOp: recurses left, recurses right, then appends an "apply" op
//     with the comparison operator as Op.
//   - *CILLogicalOp: recurses left, recurses right, then appends an "apply" op
//     with the logical operator as Op.
//   - *CILNotOp: recurses the operand, then appends an "apply" op with Op Not.
//   - *CILBoolLiteral: appends a "push" op with Op Bool and Value 1.0 for true
//     or 0.0 for false.
//   - *CILNilLiteral: appends a "push" op with Op Nil and Value 0.0.
//
// Parameters:
//   - node: The current CILNode being visited.
//   - ops: Pointer to the slice of PostfixOp entries to append to.
//
// Returns:
//   - error: An error if an unsupported node type is encountered or recursion fails,
//     or nil on success.
func walkToPostfix(node CILNode, ops *[]calculation.PostfixOp) error {
	if node == nil {
		return fmt.Errorf("node is nil")
	}
	switch n := node.(type) {
	case *CILLiteral:
		*ops = append(*ops, calculation.PostfixOp{Kind: "push", Value: n.Value})
	case *CILInputRef:
		*ops = append(*ops, calculation.PostfixOp{Kind: "push", Name: n.Name})
	case *CILLookupNode:
		*ops = append(*ops, calculation.PostfixOp{Kind: "push", Name: n.LookupName})
	case *CILBinaryOp:
		if err := walkToPostfix(n.Left, ops); err != nil {
			return err
		}
		if err := walkToPostfix(n.Right, ops); err != nil {
			return err
		}
		*ops = append(*ops, calculation.PostfixOp{Kind: "apply", Op: n.Op})
	case *CILUnaryOp:
		if err := walkToPostfix(n.Operand, ops); err != nil {
			return err
		}
		*ops = append(*ops, calculation.PostfixOp{Kind: "apply", Op: n.Op})
	case *CILComparisonOp:
		if err := walkToPostfix(n.Left, ops); err != nil {
			return err
		}
		if err := walkToPostfix(n.Right, ops); err != nil {
			return err
		}
		*ops = append(*ops, calculation.PostfixOp{Kind: "apply", Op: n.Op})
	case *CILLogicalOp:
		if err := walkToPostfix(n.Left, ops); err != nil {
			return err
		}
		if err := walkToPostfix(n.Right, ops); err != nil {
			return err
		}
		*ops = append(*ops, calculation.PostfixOp{Kind: "apply", Op: n.Op})
	case *CILNotOp:
		if err := walkToPostfix(n.Operand, ops); err != nil {
			return err
		}
		*ops = append(*ops, calculation.PostfixOp{Kind: "apply", Op: operator.Not})
	case *CILBoolLiteral:
		val := 0.0
		if n.Value {
			val = 1.0
		}
		*ops = append(*ops, calculation.PostfixOp{Kind: "push", Op: operator.Bool, Value: val})
	case *CILNilLiteral:
		*ops = append(*ops, calculation.PostfixOp{Kind: "push", Op: operator.Nil, Value: 0.0})
	case *CILConditional:
		branches := make([]calculation.PostfixBranch, 0, len(n.Branches))
		for _, b := range n.Branches {
			var condOps []calculation.PostfixOp
			if b.Condition != nil {
				condOps = make([]calculation.PostfixOp, 0)
				if err := walkToPostfix(b.Condition, &condOps); err != nil {
					return err
				}
			}
			bodyOps := make([]calculation.PostfixOp, 0)
			if err := walkToPostfix(b.Body, &bodyOps); err != nil {
				return err
			}
			branches = append(branches, calculation.PostfixBranch{
				Condition: condOps,
				Body:      bodyOps,
			})
		}
		*ops = append(*ops, calculation.PostfixOp{Kind: "conditional", Branches: branches})
	case *CILWhenClause:
		branches := make([]calculation.PostfixBranch, 0, len(n.Clauses)+1)
		for _, c := range n.Clauses {
			condOps := make([]calculation.PostfixOp, 0)
			if err := walkToPostfix(c.Condition, &condOps); err != nil {
				return err
			}
			bodyOps := make([]calculation.PostfixOp, 0)
			if err := walkToPostfix(c.Body, &bodyOps); err != nil {
				return err
			}
			branches = append(branches, calculation.PostfixBranch{
				Condition: condOps,
				Body:      bodyOps,
			})
		}
		if n.Default != nil {
			bodyOps := make([]calculation.PostfixOp, 0)
			if err := walkToPostfix(n.Default, &bodyOps); err != nil {
				return err
			}
			branches = append(branches, calculation.PostfixBranch{
				Condition: nil,
				Body:      bodyOps,
			})
		}
		*ops = append(*ops, calculation.PostfixOp{Kind: "conditional", Branches: branches})
	case *CILAggregateNode:
		return fmt.Errorf("CILAggregateNode must be handled by TransformMetric, not walkToPostfix")
	default:
		return fmt.Errorf("unsupported CIL node type: %T", node)
	}
	return nil
}
