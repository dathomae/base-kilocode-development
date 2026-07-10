// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package analysis

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unsupportedExpression is a custom Expression type used to test the default case.
type unsupportedExpression struct{}

func (unsupportedExpression) isExpression() {}

func TestExpressionToCILNode_NilExpression(t *testing.T) {
	t.Parallel()

	_, err := ExpressionToCILNode(nil)
	require.Error(t, err, "nil expression should return an error")
	assert.Contains(t, err.Error(), "nil expression", "error message should mention nil expression")
}

func TestExpressionToCILNode_UnitReference(t *testing.T) {
	t.Parallel()

	expr := &UnitReference{Name: "Length"}
	node, err := ExpressionToCILNode(expr)
	require.NoError(t, err)

	ref, ok := node.(*cil.CILUnitRef)
	require.True(t, ok, "expected *cil.CILUnitRef, got %T", node)
	assert.Equal(t, "Length", ref.UnitGroup, "UnitGroup should match")
}

func TestExpressionToCILNode_NumberNode(t *testing.T) {
	t.Parallel()

	expr := &NumberNode{Value: 3.14}
	node, err := ExpressionToCILNode(expr)
	require.NoError(t, err)

	num, ok := node.(*cil.CILNumber)
	require.True(t, ok, "expected *cil.CILNumber, got %T", node)
	assert.InDelta(t, 3.14, num.Value, 0.0001, "Value should match")
}

func TestExpressionToCILNode_MultiplicationNode(t *testing.T) {
	t.Parallel()

	expr := &MultiplicationNode{
		Left:  &UnitReference{Name: "Length"},
		Right: &UnitReference{Name: "Time"},
	}
	node, err := ExpressionToCILNode(expr)
	require.NoError(t, err)

	binOp, ok := node.(*cil.CILBinaryOp)
	require.True(t, ok, "expected *cil.CILBinaryOp, got %T", node)
	assert.Equal(t, operator.Multiply, binOp.Op, "operator should be Multiply")
	assertIsUnitRef(t, binOp.Left, "Length")
	assertIsUnitRef(t, binOp.Right, "Time")
}

func TestExpressionToCILNode_DivisionNode(t *testing.T) {
	t.Parallel()

	expr := &DivisionNode{
		Left:  &UnitReference{Name: "Length"},
		Right: &UnitReference{Name: "Time"},
	}
	node, err := ExpressionToCILNode(expr)
	require.NoError(t, err)

	binOp, ok := node.(*cil.CILBinaryOp)
	require.True(t, ok, "expected *cil.CILBinaryOp, got %T", node)
	assert.Equal(t, operator.Divide, binOp.Op, "operator should be Divide")
	assertIsUnitRef(t, binOp.Left, "Length")
	assertIsUnitRef(t, binOp.Right, "Time")
}

func TestExpressionToCILNode_AdditionNode(t *testing.T) {
	t.Parallel()

	expr := &AdditionNode{
		Left:  &UnitReference{Name: "Length"},
		Right: &UnitReference{Name: "Time"},
	}
	node, err := ExpressionToCILNode(expr)
	require.NoError(t, err)

	binOp, ok := node.(*cil.CILBinaryOp)
	require.True(t, ok, "expected *cil.CILBinaryOp, got %T", node)
	assert.Equal(t, operator.Add, binOp.Op, "operator should be Add")
	assertIsUnitRef(t, binOp.Left, "Length")
	assertIsUnitRef(t, binOp.Right, "Time")
}

func TestExpressionToCILNode_SubtractionNode(t *testing.T) {
	t.Parallel()

	expr := &SubtractionNode{
		Left:  &UnitReference{Name: "Length"},
		Right: &UnitReference{Name: "Time"},
	}
	node, err := ExpressionToCILNode(expr)
	require.NoError(t, err)

	binOp, ok := node.(*cil.CILBinaryOp)
	require.True(t, ok, "expected *cil.CILBinaryOp, got %T", node)
	assert.Equal(t, operator.Subtract, binOp.Op, "operator should be Subtract")
	assertIsUnitRef(t, binOp.Left, "Length")
	assertIsUnitRef(t, binOp.Right, "Time")
}

func TestExpressionToCILNode_PowerNode(t *testing.T) {
	t.Parallel()

	expr := &PowerNode{
		Base:     &UnitReference{Name: "Length"},
		Exponent: 3,
	}
	node, err := ExpressionToCILNode(expr)
	require.NoError(t, err)

	power, ok := node.(*cil.CILPower)
	require.True(t, ok, "expected *cil.CILPower, got %T", node)
	assert.Equal(t, 3, power.Exponent, "Exponent should match")
	assertIsUnitRef(t, power.Base, "Length")
}

func TestExpressionToCILNode_VariableRef(t *testing.T) {
	t.Parallel()

	expr := &VariableRef{Name: "m"}
	node, err := ExpressionToCILNode(expr)
	require.NoError(t, err)

	ref, ok := node.(*cil.CILVariableRef)
	require.True(t, ok, "expected *cil.CILVariableRef, got %T", node)
	assert.Equal(t, "m", ref.Name, "Name should match variable name")
}

func TestExpressionToCILNode_UnsupportedType(t *testing.T) {
	t.Parallel()

	expr := unsupportedExpression{}
	_, err := ExpressionToCILNode(expr)
	require.Error(t, err, "unsupported expression type should return an error")
	assert.Contains(t, err.Error(), "unsupported expression type", "error message should mention unsupported type")
	assert.Contains(t, err.Error(), "analysis.unsupportedExpression", "error message should include the concrete type")
}

func TestExpressionToCILNode_NestedBinaryOp(t *testing.T) {
	t.Parallel()

	expr := &MultiplicationNode{
		Left: &DivisionNode{
			Left:  &UnitReference{Name: "Length"},
			Right: &UnitReference{Name: "Time"},
		},
		Right: &UnitReference{Name: "Mass"},
	}
	node, err := ExpressionToCILNode(expr)
	require.NoError(t, err)

	outer, ok := node.(*cil.CILBinaryOp)
	require.True(t, ok, "expected *cil.CILBinaryOp, got %T", node)
	assert.Equal(t, operator.Multiply, outer.Op, "outer operator should be Multiply")
	assertIsUnitRef(t, outer.Right, "Mass")

	inner, ok := outer.Left.(*cil.CILBinaryOp)
	require.True(t, ok, "expected inner *cil.CILBinaryOp, got %T", outer.Left)
	assert.Equal(t, operator.Divide, inner.Op, "inner operator should be Divide")
	assertIsUnitRef(t, inner.Left, "Length")
	assertIsUnitRef(t, inner.Right, "Time")
}

// assertIsUnitRef asserts that node is a *cil.CILUnitRef with the expected UnitGroup.
func assertIsUnitRef(t *testing.T, node cil.CILNode, expected string) {
	t.Helper()
	ref, ok := node.(*cil.CILUnitRef)
	require.True(t, ok, "expected *cil.CILUnitRef, got %T", node)
	assert.Equal(t, expected, ref.UnitGroup, "UnitGroup should match")
}
