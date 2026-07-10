// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package synthesis

import (
	"fmt"
	"strconv"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
)

// SynthesizeExpression converts a CIL expression tree into a Go expression string.
// The emitted Go code maintains proper operator precedence through parenthesization.
func SynthesizeExpression(node cil.CILNode) (string, error) {
	if node == nil {
		return "", fmt.Errorf("nil CILNode")
	}

	switch e := node.(type) {
	case *cil.CILUnitRef:
		return e.UnitGroup, nil
	case *cil.CILVariableRef:
		return e.Name, nil
	case *cil.CILNumber:
		// Emit as float literal
		return strconv.FormatFloat(e.Value, 'f', -1, 64), nil
	case *cil.CILBinaryOp:
		left, err := SynthesizeExpression(e.Left)
		if err != nil {
			return "", err
		}
		right, err := SynthesizeExpression(e.Right)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s %s %s)", left, e.Op.Symbol(), right), nil
	case *cil.CILPower:
		base, err := SynthesizeExpression(e.Base)
		if err != nil {
			return "", err
		}
		if e.Exponent == 2 {
			return fmt.Sprintf("(%s * %s)", base, base), nil
		}
		// Use math.Pow for other exponents
		return fmt.Sprintf("math.Pow(%s, %d)", base, e.Exponent), nil
	case *cil.CILConversionApplication:
		operand, err := SynthesizeExpression(e.Operand)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("convertUnit(%q, %q, %s)", e.FromUnit, e.ToUnit, operand), nil
	default:
		return "", fmt.Errorf("unsupported CILNode type: %T", node)
	}
}
