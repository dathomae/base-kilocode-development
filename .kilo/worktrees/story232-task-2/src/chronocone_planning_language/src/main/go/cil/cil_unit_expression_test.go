// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"fmt"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testUnsupportedNode is a CILNode that ComputeUnitExpression does not handle,
// used to verify error paths for unsupported node types.
type testUnsupportedNode struct{}

func (t *testUnsupportedNode) cilNode() {}

// =============================================================
// UnitExpression structural tests
// =============================================================

func TestNewUnitExpression(t *testing.T) {
	ue := NewUnitExpression()
	assert.NotNil(t, ue.Numerator)
	assert.NotNil(t, ue.Denominator)
	assert.True(t, ue.IsDimensionless())
}

func TestUnitExpressionIsDimensionless(t *testing.T) {
	assert.True(t, NewUnitExpression().IsDimensionless())

	withNum := NewUnitExpression()
	withNum.Numerator["m"] = 1
	assert.False(t, withNum.IsDimensionless())

	withDen := NewUnitExpression()
	withDen.Denominator["s"] = 1
	assert.False(t, withDen.IsDimensionless())
}

func TestUnitExpressionEquals(t *testing.T) {
	a := NewUnitExpression()
	a.Numerator["m"] = 1

	b := NewUnitExpression()
	b.Numerator["m"] = 1

	c := NewUnitExpression()
	c.Numerator["s"] = 1

	d := NewUnitExpression()
	d.Denominator["m"] = 1

	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
	assert.False(t, a.Equals(d))
	assert.True(t, NewUnitExpression().Equals(NewUnitExpression()))
}

func TestUnitExpressionString(t *testing.T) {
	tests := []struct {
		name string
		ue   UnitExpression
		want string
	}{
		{
			name: "dimensionless",
			ue:   NewUnitExpression(),
			want: "1",
		},
		{
			name: "single numerator",
			ue: func() UnitExpression {
				u := NewUnitExpression()
				u.Numerator["m"] = 1
				return u
			}(),
			want: "m",
		},
		{
			name: "single denominator",
			ue: func() UnitExpression {
				u := NewUnitExpression()
				u.Denominator["s"] = 1
				return u
			}(),
			want: "1/s",
		},
		{
			name: "multiple numerator",
			ue: func() UnitExpression {
				u := NewUnitExpression()
				u.Numerator["m"] = 1
				u.Numerator["kg"] = 1
				return u
			}(),
			want: "kg*m",
		},
		{
			name: "mixed",
			ue: func() UnitExpression {
				u := NewUnitExpression()
				u.Numerator["m"] = 1
				u.Denominator["s"] = 1
				return u
			}(),
			want: "m/s",
		},
		{
			name: "with exponents",
			ue: func() UnitExpression {
				u := NewUnitExpression()
				u.Numerator["m"] = 2
				u.Denominator["s"] = 3
				return u
			}(),
			want: "m^2/s^3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ue.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUnitExpressionMultiply(t *testing.T) {
	a := NewUnitExpression()
	a.Numerator["m"] = 1

	b := NewUnitExpression()
	b.Numerator["s"] = 1

	result := a.multiply(b)
	assert.Equal(t, 1.0, result.Numerator["m"])
	assert.Equal(t, 1.0, result.Numerator["s"])
	assert.Empty(t, result.Denominator)
}

func TestUnitExpressionMultiply_Cancellation(t *testing.T) {
	a := NewUnitExpression()
	a.Numerator["m"] = 1

	b := NewUnitExpression()
	b.Denominator["m"] = 1

	result := a.multiply(b)
	assert.Empty(t, result.Numerator)
	assert.Empty(t, result.Denominator)
	assert.True(t, result.IsDimensionless())
}

func TestUnitExpressionDivide(t *testing.T) {
	a := NewUnitExpression()
	a.Numerator["m"] = 1

	b := NewUnitExpression()
	b.Numerator["s"] = 1

	result := a.divide(b)
	assert.Equal(t, 1.0, result.Numerator["m"])
	assert.Equal(t, 1.0, result.Denominator["s"])
}

func TestUnitExpressionPower(t *testing.T) {
	base := NewUnitExpression()
	base.Numerator["m"] = 1
	base.Denominator["s"] = 1

	result := base.power(2)
	assert.Equal(t, 2.0, result.Numerator["m"])
	assert.Equal(t, 2.0, result.Denominator["s"])
}

func TestUnitExpressionPower_Negative(t *testing.T) {
	base := NewUnitExpression()
	base.Numerator["m"] = 1

	result := base.power(-2)
	assert.Empty(t, result.Numerator)
	assert.Equal(t, 2.0, result.Denominator["m"])
}

func TestUnitExpressionPower_Fractional(t *testing.T) {
	base := NewUnitExpression()
	base.Numerator["m"] = 1

	result := base.power(1.5)
	assert.Equal(t, 1.5, result.Numerator["m"])
}

// =============================================================
// ComputeUnitExpression – CILUnitRef
// =============================================================

func TestComputeUnitExpression_CILUnitRef(t *testing.T) {
	node := &CILUnitRef{UnitGroup: "Meter"}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.Equal(t, 1.0, ue.Numerator["Meter"])
	assert.Empty(t, ue.Denominator)
}

// =============================================================
// ComputeUnitExpression – CILNumber
// =============================================================

func TestComputeUnitExpression_CILNumber_Dimensionless(t *testing.T) {
	node := &CILNumber{Value: 42.0, AnnotatedUnit: ""}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.True(t, ue.IsDimensionless())
}

func TestComputeUnitExpression_CILNumber_WithUnit(t *testing.T) {
	node := &CILNumber{Value: 42.0, AnnotatedUnit: "Meter"}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.Equal(t, 1.0, ue.Numerator["Meter"])
	assert.Empty(t, ue.Denominator)
}

// =============================================================
// ComputeUnitExpression – CILLiteral
// =============================================================

func TestComputeUnitExpression_CILLiteral(t *testing.T) {
	node := &CILLiteral{Value: 3.14}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.True(t, ue.IsDimensionless())
}

// =============================================================
// ComputeUnitExpression – CILUnaryOp
// =============================================================

func TestComputeUnitExpression_CILUnaryOp_UnitRefOperand(t *testing.T) {
	node := &CILUnaryOp{
		Op:      operator.Negate,
		Operand: &CILUnitRef{UnitGroup: "Meter"},
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.Equal(t, 1.0, ue.Numerator["Meter"])
	assert.Empty(t, ue.Denominator)
}

func TestComputeUnitExpression_CILUnaryOp_DimensionlessOperand(t *testing.T) {
	node := &CILUnaryOp{
		Op:      operator.Negate,
		Operand: &CILNumber{Value: 42.0},
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.True(t, ue.IsDimensionless())
}

func TestComputeUnitExpression_CILUnaryOp_NilOperand(t *testing.T) {
	node := &CILUnaryOp{
		Op:      operator.Negate,
		Operand: nil,
	}
	_, err := ComputeUnitExpression(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

// =============================================================
// ComputeUnitExpression – CILBinaryOp Multiply
// =============================================================

func TestComputeUnitExpression_BinaryOp_Multiply(t *testing.T) {
	node := &CILBinaryOp{
		Op:    operator.Multiply,
		Left:  &CILUnitRef{UnitGroup: "Meter"},
		Right: &CILUnitRef{UnitGroup: "Second"},
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.Equal(t, 1.0, ue.Numerator["Meter"])
	assert.Equal(t, 1.0, ue.Numerator["Second"])
	assert.Empty(t, ue.Denominator)
}

func TestComputeUnitExpression_BinaryOp_Multiply_Dimensionless(t *testing.T) {
	node := &CILBinaryOp{
		Op:    operator.Multiply,
		Left:  &CILNumber{Value: 5.0},
		Right: &CILUnitRef{UnitGroup: "Meter"},
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.Equal(t, 1.0, ue.Numerator["Meter"])
	assert.Empty(t, ue.Denominator)
}

// =============================================================
// ComputeUnitExpression – CILBinaryOp Divide
// =============================================================

func TestComputeUnitExpression_BinaryOp_Divide(t *testing.T) {
	node := &CILBinaryOp{
		Op:    operator.Divide,
		Left:  &CILUnitRef{UnitGroup: "Meter"},
		Right: &CILUnitRef{UnitGroup: "Second"},
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.Equal(t, 1.0, ue.Numerator["Meter"])
	assert.Equal(t, 1.0, ue.Denominator["Second"])
}

func TestComputeUnitExpression_BinaryOp_Divide_Complex(t *testing.T) {
	// (m/s) / (kg) => m/(s*kg)
	node := &CILBinaryOp{
		Op: operator.Divide,
		Left: &CILBinaryOp{
			Op:    operator.Divide,
			Left:  &CILUnitRef{UnitGroup: "Meter"},
			Right: &CILUnitRef{UnitGroup: "Second"},
		},
		Right: &CILUnitRef{UnitGroup: "Kilogram"},
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.Equal(t, 1.0, ue.Numerator["Meter"])
	assert.Equal(t, 1.0, ue.Denominator["Second"])
	assert.Equal(t, 1.0, ue.Denominator["Kilogram"])
}

// =============================================================
// ComputeUnitExpression – CILBinaryOp Add / Subtract
// =============================================================

func TestComputeUnitExpression_BinaryOp_Add_SameUnit(t *testing.T) {
	node := &CILBinaryOp{
		Op:    operator.Add,
		Left:  &CILUnitRef{UnitGroup: "Meter"},
		Right: &CILUnitRef{UnitGroup: "Meter"},
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.Equal(t, 1.0, ue.Numerator["Meter"])
}

func TestComputeUnitExpression_BinaryOp_Subtract_SameUnit(t *testing.T) {
	node := &CILBinaryOp{
		Op:    operator.Subtract,
		Left:  &CILUnitRef{UnitGroup: "Second"},
		Right: &CILUnitRef{UnitGroup: "Second"},
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.Equal(t, 1.0, ue.Numerator["Second"])
}

func TestComputeUnitExpression_BinaryOp_Add_DifferentUnits(t *testing.T) {
	node := &CILBinaryOp{
		Op:    operator.Add,
		Left:  &CILUnitRef{UnitGroup: "Meter"},
		Right: &CILUnitRef{UnitGroup: "Second"},
	}
	_, err := ComputeUnitExpression(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "incompatible units")
	assert.Contains(t, err.Error(), "Meter")
	assert.Contains(t, err.Error(), "Second")
}

func TestComputeUnitExpression_BinaryOp_Subtract_DifferentUnits(t *testing.T) {
	node := &CILBinaryOp{
		Op:    operator.Subtract,
		Left:  &CILUnitRef{UnitGroup: "Meter"},
		Right: &CILUnitRef{UnitGroup: "Second"},
	}
	_, err := ComputeUnitExpression(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "incompatible units")
}

func TestComputeUnitExpression_BinaryOp_Add_Dimensionless(t *testing.T) {
	node := &CILBinaryOp{
		Op:    operator.Add,
		Left:  &CILNumber{Value: 1.0},
		Right: &CILNumber{Value: 2.0},
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.True(t, ue.IsDimensionless())
}

func TestComputeUnitExpression_BinaryOp_Add_CompositeUnits(t *testing.T) {
	// m/s + m/s should succeed
	node := &CILBinaryOp{
		Op: operator.Add,
		Left: &CILBinaryOp{
			Op:    operator.Divide,
			Left:  &CILUnitRef{UnitGroup: "Meter"},
			Right: &CILUnitRef{UnitGroup: "Second"},
		},
		Right: &CILBinaryOp{
			Op:    operator.Divide,
			Left:  &CILUnitRef{UnitGroup: "Meter"},
			Right: &CILUnitRef{UnitGroup: "Second"},
		},
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.Equal(t, 1.0, ue.Numerator["Meter"])
	assert.Equal(t, 1.0, ue.Denominator["Second"])
}

func TestComputeUnitExpression_BinaryOp_Add_MismatchedComposite(t *testing.T) {
	// m/s + m/kg should fail
	node := &CILBinaryOp{
		Op: operator.Add,
		Left: &CILBinaryOp{
			Op:    operator.Divide,
			Left:  &CILUnitRef{UnitGroup: "Meter"},
			Right: &CILUnitRef{UnitGroup: "Second"},
		},
		Right: &CILBinaryOp{
			Op:    operator.Divide,
			Left:  &CILUnitRef{UnitGroup: "Meter"},
			Right: &CILUnitRef{UnitGroup: "Kilogram"},
		},
	}
	_, err := ComputeUnitExpression(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "incompatible units")
}

// =============================================================
// ComputeUnitExpression – CILPower
// =============================================================

func TestComputeUnitExpression_CILPower_Positive(t *testing.T) {
	node := &CILPower{
		Base:     &CILUnitRef{UnitGroup: "Meter"},
		Exponent: 2,
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.Equal(t, 2.0, ue.Numerator["Meter"])
	assert.Empty(t, ue.Denominator)
}

func TestComputeUnitExpression_CILPower_Negative(t *testing.T) {
	node := &CILPower{
		Base:     &CILUnitRef{UnitGroup: "Meter"},
		Exponent: -2,
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.Empty(t, ue.Numerator)
	assert.Equal(t, 2.0, ue.Denominator["Meter"])
}

func TestComputeUnitExpression_CILPower_Zero(t *testing.T) {
	node := &CILPower{
		Base:     &CILUnitRef{UnitGroup: "Meter"},
		Exponent: 0,
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.True(t, ue.IsDimensionless())
}

func TestComputeUnitExpression_CILPower_DimensionlessBase(t *testing.T) {
	node := &CILPower{
		Base:     &CILNumber{Value: 5.0},
		Exponent: 3,
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.True(t, ue.IsDimensionless())
}

func TestComputeUnitExpression_CILPower_CompositeBase(t *testing.T) {
	// (m/s)^2 => m^2/s^2
	node := &CILPower{
		Base: &CILBinaryOp{
			Op:    operator.Divide,
			Left:  &CILUnitRef{UnitGroup: "Meter"},
			Right: &CILUnitRef{UnitGroup: "Second"},
		},
		Exponent: 2,
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.Equal(t, 2.0, ue.Numerator["Meter"])
	assert.Equal(t, 2.0, ue.Denominator["Second"])
}

func TestComputeUnitExpression_CILPower_CompositeBase_ZeroExponent(t *testing.T) {
	// (m/s)^0 => dimensionless
	node := &CILPower{
		Base: &CILBinaryOp{
			Op:    operator.Divide,
			Left:  &CILUnitRef{UnitGroup: "Meter"},
			Right: &CILUnitRef{UnitGroup: "Second"},
		},
		Exponent: 0,
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.True(t, ue.IsDimensionless())
}

// =============================================================
// ComputeUnitExpression – Unsupported / Edge Cases
// =============================================================

func TestComputeUnitExpression_NilNode(t *testing.T) {
	_, err := ComputeUnitExpression(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestComputeUnitExpression_UnsupportedNodeType(t *testing.T) {
	_, err := ComputeUnitExpression(&testUnsupportedNode{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported CIL node type")
}

func TestComputeUnitExpression_BinaryOp_UnsupportedOperator(t *testing.T) {
	node := &CILBinaryOp{
		Op:    operator.GreaterThan,
		Left:  &CILNumber{Value: 1.0},
		Right: &CILNumber{Value: 2.0},
	}
	_, err := ComputeUnitExpression(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported binary operator")
}

func TestComputeUnitExpression_NilLeftOperand(t *testing.T) {
	node := &CILBinaryOp{
		Op:    operator.Multiply,
		Left:  nil,
		Right: &CILUnitRef{UnitGroup: "Meter"},
	}
	_, err := ComputeUnitExpression(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil node")
}

func TestComputeUnitExpression_NilRightOperand(t *testing.T) {
	node := &CILBinaryOp{
		Op:    operator.Multiply,
		Left:  &CILUnitRef{UnitGroup: "Meter"},
		Right: nil,
	}
	_, err := ComputeUnitExpression(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil node")
}

func TestComputeUnitExpression_Power_NilBase(t *testing.T) {
	node := &CILPower{
		Base:     nil,
		Exponent: 2,
	}
	_, err := ComputeUnitExpression(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil node")
}

// =============================================================
// Integration / end-to-end unit expression scenarios
// =============================================================

func TestComputeUnitExpression_ComplexExpression(t *testing.T) {
	// (m * m) / (s * s)  =>  m^2 / s^2
	node := &CILBinaryOp{
		Op: operator.Divide,
		Left: &CILBinaryOp{
			Op:    operator.Multiply,
			Left:  &CILUnitRef{UnitGroup: "Meter"},
			Right: &CILUnitRef{UnitGroup: "Meter"},
		},
		Right: &CILBinaryOp{
			Op:    operator.Multiply,
			Left:  &CILUnitRef{UnitGroup: "Second"},
			Right: &CILUnitRef{UnitGroup: "Second"},
		},
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.Equal(t, 2.0, ue.Numerator["Meter"])
	assert.Equal(t, 2.0, ue.Denominator["Second"])
}

func TestComputeUnitExpression_Acceleration(t *testing.T) {
	// (m/s) / s  =>  m/s^2
	node := &CILBinaryOp{
		Op: operator.Divide,
		Left: &CILBinaryOp{
			Op:    operator.Divide,
			Left:  &CILUnitRef{UnitGroup: "Meter"},
			Right: &CILUnitRef{UnitGroup: "Second"},
		},
		Right: &CILUnitRef{UnitGroup: "Second"},
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.Equal(t, 1.0, ue.Numerator["Meter"])
	assert.Equal(t, 2.0, ue.Denominator["Second"])
}

func TestComputeUnitExpression_StringDeterminism(t *testing.T) {
	// Build the same expression twice and verify String() is deterministic.
	ue1, err := ComputeUnitExpression(&CILBinaryOp{
		Op:    operator.Multiply,
		Left:  &CILUnitRef{UnitGroup: "b"},
		Right: &CILUnitRef{UnitGroup: "a"},
	})
	require.NoError(t, err)

	ue2, err := ComputeUnitExpression(&CILBinaryOp{
		Op:    operator.Multiply,
		Left:  &CILUnitRef{UnitGroup: "a"},
		Right: &CILUnitRef{UnitGroup: "b"},
	})
	require.NoError(t, err)

	assert.Equal(t, ue1.String(), ue2.String())
	assert.Equal(t, "a*b", ue1.String())
}

func TestUnitExpressionString_FractionalExponents(t *testing.T) {
	ue := NewUnitExpression()
	ue.Numerator["Meter"] = 1.5
	ue.Denominator["Second"] = 0.5
	got := ue.String()
	// Should be deterministic and include fractional exponents
	assert.Equal(t, "Meter^1.5/Second^0.5", got)
}

func TestComputeUnitExpression_DivideCancellation(t *testing.T) {
	// (m/s) * (s/m) => dimensionless
	node := &CILBinaryOp{
		Op: operator.Multiply,
		Left: &CILBinaryOp{
			Op:    operator.Divide,
			Left:  &CILUnitRef{UnitGroup: "Meter"},
			Right: &CILUnitRef{UnitGroup: "Second"},
		},
		Right: &CILBinaryOp{
			Op:    operator.Divide,
			Left:  &CILUnitRef{UnitGroup: "Second"},
			Right: &CILUnitRef{UnitGroup: "Meter"},
		},
	}
	ue, err := ComputeUnitExpression(node)
	require.NoError(t, err)
	assert.True(t, ue.IsDimensionless(), "expected dimensionless but got: %s", ue.String())
}

func TestComputeUnitExpression_BinaryOp_LeftError(t *testing.T) {
	node := &CILBinaryOp{
		Op:    operator.Multiply,
		Left:  &testUnsupportedNode{},
		Right: &CILUnitRef{UnitGroup: "Meter"},
	}
	_, err := ComputeUnitExpression(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "left operand")
}

func TestComputeUnitExpression_BinaryOp_RightError(t *testing.T) {
	node := &CILBinaryOp{
		Op:    operator.Multiply,
		Left:  &CILUnitRef{UnitGroup: "Meter"},
		Right: &testUnsupportedNode{},
	}
	_, err := ComputeUnitExpression(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "right operand")
}

func TestComputeUnitExpression_Power_BaseError(t *testing.T) {
	node := &CILPower{
		Base:     &testUnsupportedNode{},
		Exponent: 2,
	}
	_, err := ComputeUnitExpression(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "power base")
}

// Benchmark for ComputeUnitExpression.
func BenchmarkComputeUnitExpression(b *testing.B) {
	// Build a moderately complex expression tree.
	node := &CILBinaryOp{
		Op: operator.Add,
		Left: &CILBinaryOp{
			Op:    operator.Multiply,
			Left:  &CILUnitRef{UnitGroup: "Meter"},
			Right: &CILUnitRef{UnitGroup: "Meter"},
		},
		Right: &CILBinaryOp{
			Op: operator.Divide,
			Left: &CILBinaryOp{
				Op:    operator.Multiply,
				Left:  &CILUnitRef{UnitGroup: "Meter"},
				Right: &CILUnitRef{UnitGroup: "Meter"},
			},
			Right: &CILNumber{Value: 1.0},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ComputeUnitExpression(node)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// Example output test to document String() behaviour.
func ExampleUnitExpression_String() {
	ue := NewUnitExpression()
	ue.Numerator["Meter"] = 2
	ue.Denominator["Second"] = 2
	fmt.Println(ue.String())
	// Output: Meter^2/Second^2
}

// =============================================================
// UnitExpression.Simplify
// =============================================================

func TestUnitExpressionSimplify_USDTimesEURPerUSD(t *testing.T) {
	// USD * (EUR/USD) => EUR
	ue := NewUnitExpression()
	ue.Numerator["USD"] = 1
	ue.Numerator["EUR"] = 1
	ue.Denominator["USD"] = 1

	simplified := ue.Simplify()
	assert.Equal(t, 1.0, simplified.Numerator["EUR"])
	assert.Empty(t, simplified.Numerator["USD"])
	assert.Empty(t, simplified.Denominator)
	assert.True(t, simplified.Equals(unitRefExpression("EUR")))
}

func TestUnitExpressionSimplify_MeterSquaredPerMeter(t *testing.T) {
	// Meter^2 / Meter => Meter
	ue := NewUnitExpression()
	ue.Numerator["Meter"] = 2
	ue.Denominator["Meter"] = 1

	simplified := ue.Simplify()
	assert.Equal(t, 1.0, simplified.Numerator["Meter"])
	assert.Empty(t, simplified.Denominator)
}

func TestUnitExpressionSimplify_USDPerUSD(t *testing.T) {
	// USD / USD => dimensionless
	ue := NewUnitExpression()
	ue.Numerator["USD"] = 1
	ue.Denominator["USD"] = 1

	simplified := ue.Simplify()
	assert.True(t, simplified.IsDimensionless())
	assert.Empty(t, simplified.Numerator)
	assert.Empty(t, simplified.Denominator)
}

func TestUnitExpressionSimplify_MeterTimesSecondPerSecond(t *testing.T) {
	// (Meter * Second) / Second => Meter
	ue := NewUnitExpression()
	ue.Numerator["Meter"] = 1
	ue.Numerator["Second"] = 1
	ue.Denominator["Second"] = 1

	simplified := ue.Simplify()
	assert.Equal(t, 1.0, simplified.Numerator["Meter"])
	assert.Empty(t, simplified.Numerator["Second"])
	assert.Empty(t, simplified.Denominator)
}

func TestUnitExpressionSimplify_MeterTimesMeterPerMeter(t *testing.T) {
	// Meter * Meter / Meter => Meter
	ue := NewUnitExpression()
	ue.Numerator["Meter"] = 2
	ue.Denominator["Meter"] = 1

	simplified := ue.Simplify()
	assert.Equal(t, 1.0, simplified.Numerator["Meter"])
	assert.Empty(t, simplified.Denominator)
}

func TestUnitExpressionSimplify_AlreadySimplified(t *testing.T) {
	// Already-simplified expression stays the same
	ue := NewUnitExpression()
	ue.Numerator["Meter"] = 1
	ue.Denominator["Second"] = 1

	simplified := ue.Simplify()
	assert.Equal(t, 1.0, simplified.Numerator["Meter"])
	assert.Equal(t, 1.0, simplified.Denominator["Second"])
}

func TestUnitExpressionSimplify_Dimensionless(t *testing.T) {
	// Simplifying a dimensionless expression stays dimensionless
	ue := NewUnitExpression()
	simplified := ue.Simplify()
	assert.True(t, simplified.IsDimensionless())
	assert.Empty(t, simplified.Numerator)
	assert.Empty(t, simplified.Denominator)
}

func TestUnitExpressionSimplify_DoesNotMutateReceiver(t *testing.T) {
	ue := NewUnitExpression()
	ue.Numerator["USD"] = 1
	ue.Numerator["EUR"] = 1
	ue.Denominator["USD"] = 1

	_ = ue.Simplify()

	// Original should remain unchanged
	assert.Equal(t, 1.0, ue.Numerator["USD"])
	assert.Equal(t, 1.0, ue.Numerator["EUR"])
	assert.Equal(t, 1.0, ue.Denominator["USD"])
}

func TestUnitExpressionSimplify_StringMatchesExpected(t *testing.T) {
	// Verify String() output for simplified expressions
	ue := NewUnitExpression()
	ue.Numerator["Meter"] = 2
	ue.Denominator["Meter"] = 1

	assert.Equal(t, "Meter", ue.Simplify().String())
}

func TestUnitExpressionSimplify_NegativeExponentInNumerator(t *testing.T) {
	// Negative exponent in numerator moves to denominator
	ue := NewUnitExpression()
	ue.Numerator["Meter"] = -2

	simplified := ue.Simplify()
	assert.Empty(t, simplified.Numerator)
	assert.Equal(t, 2.0, simplified.Denominator["Meter"])
	assert.Equal(t, "1/Meter^2", simplified.String())
}

func TestUnitExpressionSimplify_NegativeExponentInDenominator(t *testing.T) {
	// Negative exponent in denominator moves to numerator
	ue := NewUnitExpression()
	ue.Denominator["Second"] = -3

	simplified := ue.Simplify()
	assert.Equal(t, 3.0, simplified.Numerator["Second"])
	assert.Empty(t, simplified.Denominator)
	assert.Equal(t, "Second^3", simplified.String())
}
