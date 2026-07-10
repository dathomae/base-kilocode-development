// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package expected_xls_generator

import (
	"fmt"
	"strconv"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// EvaluateExpression recursively evaluates a CIL expression tree.
// valueLookup maps input variable names to their float64 values.
func EvaluateExpression(node cil.CILNode, valueLookup map[string]float64) (float64, error) {
	return EvaluateExpressionWithContext(node, &EvalContext{ValueLookup: valueLookup})
}

// EvalContext holds the runtime context for CIL expression evaluation.
type EvalContext struct {
	// ValueLookup maps input variable names to their float64 values.
	ValueLookup map[string]float64
	// LookupData maps lookup names to their rows (each row is a map of column names to values).
	LookupData map[string][]map[string]interface{}
	// DimensionValues maps dimension concept names to their current row values.
	DimensionValues map[string]string
}

// EvaluateExpressionWithContext recursively evaluates a CIL expression tree using the provided context.
func EvaluateExpressionWithContext(node cil.CILNode, ctx *EvalContext) (float64, error) {
	if ctx == nil {
		ctx = &EvalContext{}
	}
	switch n := node.(type) {
	case *cil.CILLiteral:
		return n.Value, nil

	case *cil.CILInputRef:
		val, ok := ctx.ValueLookup[n.Name]
		if !ok {
			return 0, fmt.Errorf("input reference %q not found in value lookup", n.Name)
		}
		return val, nil

	case *cil.CILLookupNode:
		if ctx.LookupData == nil {
			return 0, fmt.Errorf("lookup %q requires lookup data, but none was provided", n.LookupName)
		}
		rows, ok := ctx.LookupData[n.LookupName]
		if !ok {
			return 0, fmt.Errorf("lookup %q not found in lookup data", n.LookupName)
		}
		// Match the current row's dimension values against the lookup data rows.
		// All dimensions declared on the lookup must match for a row to be selected.
		for _, row := range rows {
			match := true
			for _, dim := range n.Dimensions {
				dimVal, ok := ctx.DimensionValues[dim]
				if !ok {
					return 0, fmt.Errorf("dimension %q not found in current context for lookup %q", dim, n.LookupName)
				}
				rowVal, ok := row[dim].(string)
				if !ok {
					// Try float64 for numeric dimension values
					if rowValFloat, ok := row[dim].(float64); ok {
						rowVal = strconv.FormatFloat(rowValFloat, 'f', -1, 64)
					} else {
						match = false
						break
					}
				}
				if rowVal != dimVal {
					match = false
					break
				}
			}
			if match {
				if val, ok := row["Value"].(float64); ok {
					return val, nil
				}
				return 0, fmt.Errorf("lookup %q matched row but Value is not a float64", n.LookupName)
			}
		}
		return 0, fmt.Errorf("no matching row found for lookup %q with dimensions %v", n.LookupName, n.Dimensions)

	case *cil.CILBinaryOp:
		left, err := EvaluateExpressionWithContext(n.Left, ctx)
		if err != nil {
			return 0, err
		}
		right, err := EvaluateExpressionWithContext(n.Right, ctx)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case operator.Add:
			return left + right, nil
		case operator.Subtract:
			return left - right, nil
		case operator.Multiply:
			return left * right, nil
		case operator.Divide:
			return left / right, nil
		default:
			return 0, fmt.Errorf("unsupported binary operator: %s", n.Op.Symbol())
		}

	case *cil.CILUnaryOp:
		operand, err := EvaluateExpressionWithContext(n.Operand, ctx)
		if err != nil {
			return 0, err
		}
		if n.Op == operator.Negate {
			return -operand, nil
		}
		return 0, fmt.Errorf("unsupported unary operator: %s", n.Op.Symbol())

	case *cil.CILComparisonOp:
		left, err := EvaluateExpressionWithContext(n.Left, ctx)
		if err != nil {
			return 0, err
		}
		right, err := EvaluateExpressionWithContext(n.Right, ctx)
		if err != nil {
			return 0, err
		}
		var result bool
		switch n.Op {
		case operator.Equal:
			result = left == right
		case operator.NotEqual:
			result = left != right
		case operator.LessThan:
			result = left < right
		case operator.GreaterThan:
			result = left > right
		case operator.LessThanOrEqual:
			result = left <= right
		case operator.GreaterThanOrEqual:
			result = left >= right
		default:
			return 0, fmt.Errorf("unsupported comparison operator: %s", n.Op.Symbol())
		}
		if result {
			return 1.0, nil
		}
		return 0.0, nil

	case *cil.CILLogicalOp:
		left, err := EvaluateExpressionWithContext(n.Left, ctx)
		if err != nil {
			return 0, err
		}
		right, err := EvaluateExpressionWithContext(n.Right, ctx)
		if err != nil {
			return 0, err
		}
		leftTrue := left != 0
		rightTrue := right != 0
		switch n.Op {
		case operator.And:
			if leftTrue && rightTrue {
				return 1.0, nil
			}
			return 0.0, nil
		case operator.Or:
			if leftTrue || rightTrue {
				return 1.0, nil
			}
			return 0.0, nil
		default:
			return 0, fmt.Errorf("unsupported logical operator: %s", n.Op.Symbol())
		}

	case *cil.CILNotOp:
		operand, err := EvaluateExpressionWithContext(n.Operand, ctx)
		if err != nil {
			return 0, err
		}
		if operand == 0 {
			return 1.0, nil
		}
		return 0.0, nil

	case *cil.CILBoolLiteral:
		if n.Value {
			return 1.0, nil
		}
		return 0.0, nil

	case *cil.CILNilLiteral:
		return 0.0, nil

	case *cil.CILConditional:
		for _, branch := range n.Branches {
			if branch.Condition == nil {
				return EvaluateExpressionWithContext(branch.Body, ctx)
			}
			condVal, err := EvaluateExpressionWithContext(branch.Condition, ctx)
			if err != nil {
				return 0, err
			}
			if condVal != 0 {
				return EvaluateExpressionWithContext(branch.Body, ctx)
			}
		}
		return 0.0, nil

	case *cil.CILWhenClause:
		for _, clause := range n.Clauses {
			condVal, err := EvaluateExpressionWithContext(clause.Condition, ctx)
			if err != nil {
				return 0, err
			}
			if condVal != 0 {
				return EvaluateExpressionWithContext(clause.Body, ctx)
			}
		}
		if n.Default != nil {
			return EvaluateExpressionWithContext(n.Default, ctx)
		}
		return 0.0, nil

	default:
		return 0, fmt.Errorf("unsupported CIL node type: %T", node)
	}
}
