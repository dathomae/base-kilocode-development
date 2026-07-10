// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package synthesis

import (
	"go/parser"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// TestSynthesizeExpression_VariableRef verifies that a variable reference synthesizes
// to its variable name as a bare identifier.
func TestSynthesizeExpression_VariableRef(t *testing.T) {
	expr := &cil.CILVariableRef{Name: "m"}
	got, err := SynthesizeExpression(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "m"
	if got != want {
		t.Errorf("SynthesizeExpression() = %q, want %q", got, want)
	}
}

// TestSynthesizeExpression_UnitRef verifies that a unit reference synthesizes
// to its unit group name.
func TestSynthesizeExpression_UnitRef(t *testing.T) {
	expr := &cil.CILUnitRef{UnitGroup: "Length"}
	got, err := SynthesizeExpression(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "Length"
	if got != want {
		t.Errorf("SynthesizeExpression() = %q, want %q", got, want)
	}
}

// TestSynthesizeExpression_Number verifies that a numeric literal synthesizes
// to a Go float literal string.
func TestSynthesizeExpression_Number(t *testing.T) {
	expr := &cil.CILNumber{Value: 3.14}
	got, err := SynthesizeExpression(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "3.14"
	if got != want {
		t.Errorf("SynthesizeExpression() = %q, want %q", got, want)
	}
}

// TestSynthesizeExpression_BinaryOps verifies that binary operations synthesize
// to parenthesized Go expressions with the correct operator.
func TestSynthesizeExpression_BinaryOps(t *testing.T) {
	tests := []struct {
		name string
		op   operator.Operator
		want string
	}{
		{
			name: "multiplication",
			op:   operator.Multiply,
			want: "(left * right)",
		},
		{
			name: "division",
			op:   operator.Divide,
			want: "(left / right)",
		},
		{
			name: "addition",
			op:   operator.Add,
			want: "(left + right)",
		},
		{
			name: "subtraction",
			op:   operator.Subtract,
			want: "(left - right)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left := &cil.CILUnitRef{UnitGroup: "left"}
			right := &cil.CILUnitRef{UnitGroup: "right"}
			expr := &cil.CILBinaryOp{
				Op:          tt.op,
				Left:        left,
				Right:       right,
				DerivedUnit: "result",
			}
			got, err := SynthesizeExpression(expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("SynthesizeExpression() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSynthesizeExpression_Power verifies that exponentiation synthesizes
// correctly: exponent 2 becomes inline multiplication, others use math.Pow.
func TestSynthesizeExpression_Power(t *testing.T) {
	t.Run("exponent 2", func(t *testing.T) {
		base := &cil.CILUnitRef{UnitGroup: "Length"}
		expr := &cil.CILPower{Base: base, Exponent: 2, DerivedUnit: "Area"}
		got, err := SynthesizeExpression(expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "(Length * Length)"
		if got != want {
			t.Errorf("SynthesizeExpression() = %q, want %q", got, want)
		}
	})

	t.Run("exponent 3", func(t *testing.T) {
		base := &cil.CILUnitRef{UnitGroup: "Length"}
		expr := &cil.CILPower{Base: base, Exponent: 3, DerivedUnit: "Volume"}
		got, err := SynthesizeExpression(expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "math.Pow(Length, 3)"
		if got != want {
			t.Errorf("SynthesizeExpression() = %q, want %q", got, want)
		}
	})

	t.Run("exponent -1", func(t *testing.T) {
		base := &cil.CILUnitRef{UnitGroup: "Time"}
		expr := &cil.CILPower{Base: base, Exponent: -1, DerivedUnit: "Frequency"}
		got, err := SynthesizeExpression(expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "math.Pow(Time, -1)"
		if got != want {
			t.Errorf("SynthesizeExpression() = %q, want %q", got, want)
		}
	})
}

// TestSynthesizeExpression_Nested verifies that complex nested expressions
// are correctly parenthesized.
func TestSynthesizeExpression_Nested(t *testing.T) {
	// (Length / (Time * Time))  =>  (Length / (Time * Time))
	length := &cil.CILUnitRef{UnitGroup: "Length"}
	time := &cil.CILUnitRef{UnitGroup: "Time"}
	timeSquared := &cil.CILPower{Base: time, Exponent: 2, DerivedUnit: "TimeSquared"}
	expr := &cil.CILBinaryOp{
		Op:          operator.Divide,
		Left:        length,
		Right:       timeSquared,
		DerivedUnit: "Acceleration",
	}

	got, err := SynthesizeExpression(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "(Length / (Time * Time))"
	if got != want {
		t.Errorf("SynthesizeExpression() = %q, want %q", got, want)
	}
}

// TestSynthesizeExpression_CompilableGo verifies that emitted Go code parses
// successfully using the go/parser standard library.
func TestSynthesizeExpression_CompilableGo(t *testing.T) {
	tests := []struct {
		name string
		expr cil.CILNode
	}{
		{
			name: "unit ref",
			expr: &cil.CILUnitRef{UnitGroup: "Length"},
		},
		{
			name: "number",
			expr: &cil.CILNumber{Value: 3.14},
		},
		{
			name: "binary multiplication",
			expr: &cil.CILBinaryOp{
				Op:    operator.Multiply,
				Left:  &cil.CILUnitRef{UnitGroup: "Mass"},
				Right: &cil.CILUnitRef{UnitGroup: "Length"},
			},
		},
		{
			name: "power with math.Pow",
			expr: &cil.CILPower{
				Base:     &cil.CILUnitRef{UnitGroup: "Length"},
				Exponent: 3,
			},
		},
		{
			name: "nested complex",
			expr: &cil.CILBinaryOp{
				Op:   operator.Divide,
				Left: &cil.CILUnitRef{UnitGroup: "Length"},
				Right: &cil.CILPower{
					Base:     &cil.CILUnitRef{UnitGroup: "Time"},
					Exponent: 2,
				},
			},
		},
		{
			name: "conversion application",
			expr: &cil.CILConversionApplication{
				Operand:  &cil.CILVariableRef{Name: "distance"},
				FromUnit: "miles",
				ToUnit:   "kilometers",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := SynthesizeExpression(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			_, err = parser.ParseExpr(code)
			if err != nil {
				t.Errorf("emitted Go code does not parse: %v\nCode: %s", err, code)
			}
		})
	}
}

// TestSynthesizeExpression_FieldAccess verifies that synthesized expressions
// involving field-like identifiers produce valid Go selector expressions.
func TestSynthesizeExpression_FieldAccess(t *testing.T) {
	// Simulate accessing an internal field, e.g., result.Value
	expr := &cil.CILUnitRef{UnitGroup: "result.Value"}
	got, err := SynthesizeExpression(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "result.Value"
	if got != want {
		t.Errorf("SynthesizeExpression() = %q, want %q", got, want)
	}

	// Verify it parses as a valid Go expression
	_, err = parser.ParseExpr(got)
	if err != nil {
		t.Errorf("field access expression does not parse: %v\nCode: %s", err, got)
	}
}

// TestSynthesizeExpression_NumberFormatting verifies that numbers are formatted
// correctly without unnecessary trailing zeros.
func TestSynthesizeExpression_NumberFormatting(t *testing.T) {
	tests := []struct {
		value float64
		want  string
	}{
		{value: 0, want: "0"},
		{value: 1, want: "1"},
		{value: -42, want: "-42"},
		{value: 3.14, want: "3.14"},
		{value: 1e9, want: "1000000000"},
	}

	for _, tt := range tests {
		expr := &cil.CILNumber{Value: tt.value}
		got, err := SynthesizeExpression(expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != tt.want {
			t.Errorf("SynthesizeExpression(Number{%v}) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

// TestSynthesizeExpression_ConversionApplication verifies that a conversion
// application node synthesizes to a call to convertUnit.
func TestSynthesizeExpression_ConversionApplication(t *testing.T) {
	t.Run("simple number operand", func(t *testing.T) {
		expr := &cil.CILConversionApplication{
			Operand:  &cil.CILNumber{Value: 5.0},
			FromUnit: "meters",
			ToUnit:   "kilometers",
		}
		got, err := SynthesizeExpression(expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := `convertUnit("meters", "kilometers", 5)`
		if got != want {
			t.Errorf("SynthesizeExpression() = %q, want %q", got, want)
		}
	})

	t.Run("variable operand", func(t *testing.T) {
		expr := &cil.CILConversionApplication{
			Operand:  &cil.CILVariableRef{Name: "distance"},
			FromUnit: "feet",
			ToUnit:   "meters",
		}
		got, err := SynthesizeExpression(expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := `convertUnit("feet", "meters", distance)`
		if got != want {
			t.Errorf("SynthesizeExpression() = %q, want %q", got, want)
		}
	})

	t.Run("nested expression operand", func(t *testing.T) {
		expr := &cil.CILConversionApplication{
			Operand: &cil.CILBinaryOp{
				Op:    operator.Multiply,
				Left:  &cil.CILVariableRef{Name: "length"},
				Right: &cil.CILNumber{Value: 2.0},
			},
			FromUnit: "inches",
			ToUnit:   "centimeters",
		}
		got, err := SynthesizeExpression(expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := `convertUnit("inches", "centimeters", (length * 2))`
		if got != want {
			t.Errorf("SynthesizeExpression() = %q, want %q", got, want)
		}
	})
}

// TestSynthesizeExpression_UnsupportedNode verifies that an unsupported CILNode
// type returns an error.
func TestSynthesizeExpression_UnsupportedNode(t *testing.T) {
	_, err := SynthesizeExpression(&cil.CILInputRef{Kind: "some", Name: "test"})
	if err == nil {
		t.Errorf("expected error for unsupported node, got nil")
	}
}

// TestSynthesizeExpression_NilInput verifies that a nil input returns an error.
func TestSynthesizeExpression_NilInput(t *testing.T) {
	_, err := SynthesizeExpression(nil)
	if err == nil {
		t.Errorf("expected error for nil input, got nil")
	}
}
