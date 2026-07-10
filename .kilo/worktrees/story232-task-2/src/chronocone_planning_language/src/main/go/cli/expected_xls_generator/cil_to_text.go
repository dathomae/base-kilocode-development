package expected_xls_generator

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// ExpressionToText converts a CIL node to its human-readable text representation.
func ExpressionToText(node cil.CILNode) string {
	switch n := node.(type) {
	case *cil.CILLiteral:
		return fmt.Sprintf("%v", n.Value)

	case *cil.CILInputRef:
		return n.Name

	case *cil.CILBinaryOp:
		left := ExpressionToText(n.Left)
		right := ExpressionToText(n.Right)
		switch n.Op {
		case operator.Add:
			return fmt.Sprintf("%s + %s", left, right)
		case operator.Subtract:
			return fmt.Sprintf("%s - %s", left, right)
		case operator.Multiply:
			return fmt.Sprintf("%s * %s", left, right)
		case operator.Divide:
			return fmt.Sprintf("%s / %s", left, right)
		default:
			return fmt.Sprintf("%s %s %s", left, n.Op.Symbol(), right)
		}

	case *cil.CILUnaryOp:
		if n.Op == operator.Negate {
			return "-" + ExpressionToText(n.Operand)
		}
		return fmt.Sprintf("(%s)", ExpressionToText(n.Operand))

	case *cil.CILComparisonOp:
		left := ExpressionToText(n.Left)
		right := ExpressionToText(n.Right)
		switch n.Op {
		case operator.Equal:
			return fmt.Sprintf("%s == %s", left, right)
		case operator.NotEqual:
			return fmt.Sprintf("%s != %s", left, right)
		case operator.LessThan:
			return fmt.Sprintf("%s < %s", left, right)
		case operator.GreaterThan:
			return fmt.Sprintf("%s > %s", left, right)
		case operator.LessThanOrEqual:
			return fmt.Sprintf("%s <= %s", left, right)
		case operator.GreaterThanOrEqual:
			return fmt.Sprintf("%s >= %s", left, right)
		default:
			return fmt.Sprintf("%s %s %s", left, n.Op.Symbol(), right)
		}

	case *cil.CILLogicalOp:
		left := ExpressionToText(n.Left)
		right := ExpressionToText(n.Right)
		switch n.Op {
		case operator.And:
			return fmt.Sprintf("%s and %s", left, right)
		case operator.Or:
			return fmt.Sprintf("%s or %s", left, right)
		default:
			return fmt.Sprintf("(%s) %s (%s)", left, n.Op.Symbol(), right)
		}

	case *cil.CILNotOp:
		return "not " + ExpressionToText(n.Operand)

	case *cil.CILBoolLiteral:
		if n.Value {
			return "True"
		}
		return "False"

	case *cil.CILNilLiteral:
		return "nil"

	case *cil.CILConditional:
		if len(n.Branches) > 0 && n.Branches[0].Condition != nil {
			return "if " + ExpressionToText(n.Branches[0].Condition)
		}
		return "else"

	case *cil.CILWhenClause:
		if len(n.Clauses) > 0 {
			return "when " + ExpressionToText(n.Clauses[0].Condition)
		}
		return "default"

	default:
		return fmt.Sprintf("%T(%v)", node, node)
	}
}
