// Package analysis_test contains tests for the CPL analysis expression types.
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
)

// Compile-time checks that all concrete types implement the Expression interface.

func TestUnitReference_ImplementsExpression(t *testing.T) {
	var _ analysis.Expression = &analysis.UnitReference{Name: "Length", UnitGroup: nil}
}

func TestMultiplicationNode_ImplementsExpression(t *testing.T) {
	var _ analysis.Expression = &analysis.MultiplicationNode{}
}

func TestDivisionNode_ImplementsExpression(t *testing.T) {
	var _ analysis.Expression = &analysis.DivisionNode{}
}

func TestPowerNode_ImplementsExpression(t *testing.T) {
	var _ analysis.Expression = &analysis.PowerNode{}
}

func TestAdditionNode_ImplementsExpression(t *testing.T) {
	var _ analysis.Expression = &analysis.AdditionNode{}
}

func TestSubtractionNode_ImplementsExpression(t *testing.T) {
	var _ analysis.Expression = &analysis.SubtractionNode{}
}

func TestNumberNode_ImplementsExpression(t *testing.T) {
	var _ analysis.Expression = &analysis.NumberNode{}
}

// Field access tests for each concrete type.

func TestUnitReference_Fields(t *testing.T) {
	ur := &analysis.UnitReference{Name: "Length", UnitGroup: nil}
	if ur.Name != "Length" {
		t.Errorf("expected Name='Length', got %q", ur.Name)
	}
	if ur.UnitGroup != nil {
		t.Error("expected UnitGroup to be nil")
	}
}

func TestUnitReference_WithNonNilUnitGroup(t *testing.T) {
	ug := &analysis.AnalysisUnitDefinition{Name: "Currency", ConceptName: "Country"}
	ur := &analysis.UnitReference{Name: "Currency", UnitGroup: ug}
	if ur.UnitGroup != ug {
		t.Error("expected UnitGroup pointer to match")
	}
	if ur.UnitGroup.Name != "Currency" {
		t.Errorf("expected UnitGroup.Name='Currency', got %q", ur.UnitGroup.Name)
	}
}

func TestMultiplicationNode_Fields(t *testing.T) {
	left := &analysis.NumberNode{Value: 2.0}
	right := &analysis.NumberNode{Value: 3.0}
	mn := &analysis.MultiplicationNode{Left: left, Right: right}
	if mn.Left != left {
		t.Error("Left pointer mismatch")
	}
	if mn.Right != right {
		t.Error("Right pointer mismatch")
	}
}

func TestDivisionNode_Fields(t *testing.T) {
	left := &analysis.NumberNode{Value: 10.0}
	right := &analysis.NumberNode{Value: 2.0}
	dn := &analysis.DivisionNode{Left: left, Right: right}
	if dn.Left != left {
		t.Error("Left pointer mismatch")
	}
	if dn.Right != right {
		t.Error("Right pointer mismatch")
	}
}

func TestPowerNode_Fields(t *testing.T) {
	base := &analysis.NumberNode{Value: 5.0}
	pn := &analysis.PowerNode{Base: base, Exponent: 2}
	if pn.Base != base {
		t.Error("Base pointer mismatch")
	}
	if pn.Exponent != 2 {
		t.Errorf("expected Exponent=2, got %d", pn.Exponent)
	}
}

func TestPowerNode_ExponentValues(t *testing.T) {
	cases := []struct {
		name     string
		exponent int
	}{
		{"zero", 0},
		{"one", 1},
		{"two", 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pn := &analysis.PowerNode{Base: &analysis.NumberNode{Value: 3.0}, Exponent: tc.exponent}
			if pn.Exponent != tc.exponent {
				t.Errorf("expected Exponent=%d, got %d", tc.exponent, pn.Exponent)
			}
		})
	}
}

func TestAdditionNode_Fields(t *testing.T) {
	left := &analysis.NumberNode{Value: 1.0}
	right := &analysis.NumberNode{Value: 2.0}
	an := &analysis.AdditionNode{Left: left, Right: right}
	if an.Left != left {
		t.Error("Left pointer mismatch")
	}
	if an.Right != right {
		t.Error("Right pointer mismatch")
	}
}

func TestSubtractionNode_Fields(t *testing.T) {
	left := &analysis.NumberNode{Value: 5.0}
	right := &analysis.NumberNode{Value: 3.0}
	sn := &analysis.SubtractionNode{Left: left, Right: right}
	if sn.Left != left {
		t.Error("Left pointer mismatch")
	}
	if sn.Right != right {
		t.Error("Right pointer mismatch")
	}
}

func TestNumberNode_Fields(t *testing.T) {
	nn := &analysis.NumberNode{Value: 42.0}
	if nn.Value != 42.0 {
		t.Errorf("expected Value=42.0, got %f", nn.Value)
	}
}

func TestNumberNode_IntAndFloatValues(t *testing.T) {
	intNode := &analysis.NumberNode{Value: 7}
	if intNode.Value != 7.0 {
		t.Errorf("expected int value 7 to be 7.0, got %f", intNode.Value)
	}

	floatNode := &analysis.NumberNode{Value: 3.14}
	if floatNode.Value != 3.14 {
		t.Errorf("expected Value=3.14, got %f", floatNode.Value)
	}
}

func TestBinaryNode_Nesting(t *testing.T) {
	num := &analysis.NumberNode{Value: 10.0}
	ur := &analysis.UnitReference{Name: "Meter", UnitGroup: nil}
	mn := &analysis.MultiplicationNode{Left: num, Right: ur}

	if mn.Left != num {
		t.Error("expected Left to be NumberNode")
	}
	if mn.Right != ur {
		t.Error("expected Right to be UnitReference")
	}

	// Verify nested types via interface assignment
	if _, ok := mn.Left.(*analysis.NumberNode); !ok {
		t.Error("expected Left to be *NumberNode")
	}
	if _, ok := mn.Right.(*analysis.UnitReference); !ok {
		t.Error("expected Right to be *UnitReference")
	}
}
