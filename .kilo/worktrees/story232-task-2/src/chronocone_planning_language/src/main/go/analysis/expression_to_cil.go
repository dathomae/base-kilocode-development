// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package analysis

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// ExpressionToCILNode converts an analysis Expression tree into a CIL expression tree.
// It maps each analysis expression type to the corresponding unified CILNode type.
// DerivedUnit and AnnotatedUnit return an error (not yet implemented — Story229/231).
func ExpressionToCILNode(expr Expression) (cil.CILNode, error) {
	if expr == nil {
		return nil, fmt.Errorf("nil expression")
	}

	switch e := expr.(type) {
	case *UnitReference:
		return &cil.CILUnitRef{UnitGroup: e.Name}, nil
	case *NumberNode:
		return &cil.CILNumber{Value: e.Value}, nil
	case *MultiplicationNode:
		left, err := ExpressionToCILNode(e.Left)
		if err != nil {
			return nil, err
		}
		right, err := ExpressionToCILNode(e.Right)
		if err != nil {
			return nil, err
		}
		return &cil.CILBinaryOp{Op: operator.Multiply, Left: left, Right: right}, nil
	case *DivisionNode:
		left, err := ExpressionToCILNode(e.Left)
		if err != nil {
			return nil, err
		}
		right, err := ExpressionToCILNode(e.Right)
		if err != nil {
			return nil, err
		}
		return &cil.CILBinaryOp{Op: operator.Divide, Left: left, Right: right}, nil
	case *AdditionNode:
		left, err := ExpressionToCILNode(e.Left)
		if err != nil {
			return nil, err
		}
		right, err := ExpressionToCILNode(e.Right)
		if err != nil {
			return nil, err
		}
		return &cil.CILBinaryOp{Op: operator.Add, Left: left, Right: right}, nil
	case *SubtractionNode:
		left, err := ExpressionToCILNode(e.Left)
		if err != nil {
			return nil, err
		}
		right, err := ExpressionToCILNode(e.Right)
		if err != nil {
			return nil, err
		}
		return &cil.CILBinaryOp{Op: operator.Subtract, Left: left, Right: right}, nil
	case *PowerNode:
		base, err := ExpressionToCILNode(e.Base)
		if err != nil {
			return nil, err
		}
		return &cil.CILPower{Base: base, Exponent: e.Exponent}, nil
	case *VariableRef:
		return &cil.CILVariableRef{Name: e.Name}, nil
	default:
		// DerivedUnit and AnnotatedUnit are not yet implemented (Story229/231).
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}
