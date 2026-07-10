// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"fmt"
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// TestTransformMetric verifies the structural behavior of TransformMetric.
// It tests happy path, error cases, and table-driven literal value generation.
func TestTransformMetric(t *testing.T) {
	tests := []struct {
		name                string
		metricName          string
		root                interface{}
		expectError         bool
		expectContains      []string
		expectNotContains   []string
		expectErrorContains string
		skipTransformerInit bool
	}{
		{
			name:       "happy path with CILLiteral child",
			metricName: "TotalSales",
			root: &CILMetricNode{
				CILNode: &CILLiteral{Value: 42.0},
			},
			expectError: false,
			expectContains: []string{
				"func computeTotalSales(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"_ = inputs",
				"[]calculation.PostfixOp",
				"EvaluatePostfixFloat64",
				"Value: 42",
				"controller.GetMetricComputation(\"TotalSales\").SetTensor(result)",
				"Failed to set tensor for TotalSales",
				"check(",
			},
			expectNotContains: []string{
				"// TODO:",
				"calculation.AllocateTensor",
				"data := result.Data()",
				"for i := range data",
				"var err error",
				"var result ",
			},
		},
		{
			name:       "happy path with zero literal",
			metricName: "ZeroMetric",
			root: &CILMetricNode{
				CILNode: &CILLiteral{Value: 0.0},
			},
			expectError: false,
			expectContains: []string{
				"func computeZeroMetric(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"_ = inputs",
				"EvaluatePostfixFloat64",
				"Value: 0",
				"controller.GetMetricComputation(\"ZeroMetric\").SetTensor(result)",
			},
			expectNotContains: []string{
				"calculation.AllocateTensor",
				"data := result.Data()",
				"for i := range data",
			},
		},
		{
			name:       "happy path with negative literal",
			metricName: "NegativeMetric",
			root: &CILMetricNode{
				CILNode: &CILLiteral{Value: -3.14},
			},
			expectError: false,
			expectContains: []string{
				"func computeNegativeMetric(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"_ = inputs",
				"EvaluatePostfixFloat64",
				"Value: -3.14",
				"controller.GetMetricComputation(\"NegativeMetric\").SetTensor(result)",
			},
			expectNotContains: []string{
				"calculation.AllocateTensor",
				"data := result.Data()",
				"for i := range data",
			},
		},
		{
			name:       "happy path with dimension input refs",
			metricName: "DimensionalMetric",
			root: &CILMetricNode{
				CILNode: &CILLiteral{Value: 99.0},
				InputRefs: []CILInputRef{
					{Kind: "dimension", Name: "Product"},
					{Kind: "dimension", Name: "Region"},
				},
			},
			expectError: false,
			expectContains: []string{
				"func computeDimensionalMetric(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"_ = inputs",
				"calculation.AllocateTensor(controller, \"Product\", \"Region\")",
				"data := result.Data()",
				"for i := range data",
				"data[i] = float64(99)",
				"controller.GetMetricComputation(\"DimensionalMetric\").SetTensor(result)",
			},
			expectNotContains: []string{
				"EvaluatePostfixFloat64",
			},
		},
		{
			name:       "happy path with mixed input refs filters non-dimension",
			metricName: "MixedRefsMetric",
			root: &CILMetricNode{
				CILNode: &CILLiteral{Value: 77.0},
				InputRefs: []CILInputRef{
					{Kind: "metric", Name: "OtherMetric"},
					{Kind: "dimension", Name: "Time"},
				},
			},
			expectError: false,
			expectContains: []string{
				"calculation.AllocateTensor(controller, \"Time\")",
				"data := result.Data()",
				"for i := range data",
				"data[i] = float64(77)",
			},
			expectNotContains: []string{
				"_ = inputs",
				"EvaluatePostfixFloat64",
			},
		},
		{
			name:                "nil transformer",
			metricName:          "AnyMetric",
			root:                &CILMetricNode{CILNode: &CILLiteral{Value: 1.0}},
			expectError:         true,
			expectErrorContains: "transformer is nil",
			skipTransformerInit: true,
		},
		{
			name:                "wrong root type",
			metricName:          "AnyMetric",
			root:                "not a CILMetricNode",
			expectError:         true,
			expectErrorContains: "root is not a *CILMetricNode",
		},
		{
			name:                "nil metric node",
			metricName:          "AnyMetric",
			root:                (*CILMetricNode)(nil),
			expectError:         true,
			expectErrorContains: "metric node is nil",
		},
		{
			name:       "wrong child type",
			metricName: "AnyMetric",
			root: &CILMetricNode{
				CILNode: &mockCILNode{},
			},
			expectError:         true,
			expectErrorContains: "unsupported CIL node type",
		},
		{
			name:                "nil literal node",
			metricName:          "AnyMetric",
			root:                &CILMetricNode{CILNode: (*CILLiteral)(nil)},
			expectError:         true,
			expectErrorContains: "literal node is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var transformer *CILAbstractTransformer
			if !tt.skipTransformerInit {
				transformer = NewCILAbstractTransformer()
				if transformer == nil {
					t.Fatalf("NewCILAbstractTransformer() returned nil")
				}
			}

			code, err := transformer.TransformMetric(tt.metricName, tt.root)

			if tt.expectError {
				if err == nil {
					t.Errorf("TransformMetric() returned nil error, expected error")
					return
				}
				if tt.expectErrorContains != "" && !strings.Contains(err.Error(), tt.expectErrorContains) {
					t.Errorf("TransformMetric() error = %q, expected to contain %q", err.Error(), tt.expectErrorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("TransformMetric() returned unexpected error: %v", err)
				return
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(code, expected) {
					t.Errorf("TransformMetric() output missing expected content: %q", expected)
				}
			}

			for _, notExpected := range tt.expectNotContains {
				if strings.Contains(code, notExpected) {
					t.Errorf("TransformMetric() output contains unexpected content: %q", notExpected)
				}
			}
		})
	}
}

// TestTransformMetric_ValueTable verifies that TransformMetric generates the correct metric
// compute function for a variety of literal values.
func TestTransformMetric_ValueTable(t *testing.T) {
	tests := []struct {
		name           string
		metricName     string
		literalValue   float64
		expectContains string
	}{
		{
			name:           "positive integer",
			metricName:     "PosInt",
			literalValue:   100,
			expectContains: "100",
		},
		{
			name:           "positive float",
			metricName:     "PosFloat",
			literalValue:   2.718,
			expectContains: "2.718",
		},
		{
			name:           "zero",
			metricName:     "ZeroVal",
			literalValue:   0,
			expectContains: "0",
		},
		{
			name:           "negative integer",
			metricName:     "NegInt",
			literalValue:   -7,
			expectContains: "-7",
		},
		{
			name:           "negative float",
			metricName:     "NegFloat",
			literalValue:   -0.5,
			expectContains: "-0.5",
		},
		{
			name:           "large value",
			metricName:     "LargeVal",
			literalValue:   999999.999,
			expectContains: "999999.999",
		},
		{
			name:           "small decimal",
			metricName:     "SmallVal",
			literalValue:   0.001,
			expectContains: "0.001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := NewCILAbstractTransformer()
			if transformer == nil {
				t.Fatalf("NewCILAbstractTransformer() returned nil")
			}

			root := &CILMetricNode{
				CILNode: &CILLiteral{Value: tt.literalValue},
			}

			code, err := transformer.TransformMetric(tt.metricName, root)
			if err != nil {
				t.Errorf("TransformMetric() returned unexpected error: %v", err)
				return
			}

			funcName := fmt.Sprintf("func compute%s(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])", tt.metricName)
			if !strings.Contains(code, funcName) {
				t.Errorf("TransformMetric() output missing function declaration: %q", funcName)
			}

			if !strings.Contains(code, tt.expectContains) {
				t.Errorf("TransformMetric() output missing expected literal value %q", tt.expectContains)
			}
		})
	}
}

// TestTransformMetric_ArithmeticBinaryOp verifies that TransformMetric
// generates correct compute function code for binary and unary operations
// using the postfix table approach.
func TestTransformMetric_ArithmeticBinaryOp(t *testing.T) {
	tests := []struct {
		name              string
		metricName        string
		root              interface{}
		expectContains    []string
		expectNotContains []string
	}{
		{
			name:       "subtract two connector inputs",
			metricName: "profit",
			root: &CILMetricNode{
				CILNode: &CILBinaryOp{
					Op:    operator.Subtract,
					Left:  &CILInputRef{Kind: "connector", Name: "revenue_amt"},
					Right: &CILInputRef{Kind: "connector", Name: "cost_amt"},
				},
				InputRefs: []CILInputRef{
					{Kind: "connector", Name: "revenue_amt"},
					{Kind: "connector", Name: "cost_amt"},
				},
			},
			expectContains: []string{
				"func computeprofit(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"[]calculation.PostfixOp",
				`Kind: "push"`,
				`Kind: "apply"`,
				`Op: operator.Subtract`,
				`Name: "revenue_amt"`,
				`Name: "cost_amt"`,
				"EvaluatePostfixFloat64",
				"controller.GetMetricComputation(\"profit\").SetTensor(result)",
			},
			expectNotContains: []string{
				"data := result.Data()",
				"for i := range data",
				"calculation.AllocateTensor",
				"TensorSubtractFloat64",
				"TensorDimensionNamesMatch",
				"var err error",
				"var result ",
			},
		},
		{
			name:       "add two metric inputs",
			metricName: "total",
			root: &CILMetricNode{
				CILNode: &CILBinaryOp{
					Op:    operator.Add,
					Left:  &CILInputRef{Kind: "metric", Name: "sales"},
					Right: &CILInputRef{Kind: "metric", Name: "returns"},
				},
				InputRefs: []CILInputRef{
					{Kind: "metric", Name: "sales"},
					{Kind: "metric", Name: "returns"},
				},
			},
			expectContains: []string{
				"func computetotal(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"[]calculation.PostfixOp",
				`Kind: "push"`,
				`Kind: "apply"`,
				`Op: operator.Add`,
				`Name: "sales"`,
				`Name: "returns"`,
				"EvaluatePostfixFloat64",
				"controller.GetMetricComputation(\"total\").SetTensor(result)",
			},
			expectNotContains: []string{
				"data := result.Data()",
				"for i := range data",
				"TensorAddFloat64",
				"var err error",
				"var result ",
			},
		},
		{
			name:       "multiply with literal scalar",
			metricName: "doubled",
			root: &CILMetricNode{
				CILNode: &CILBinaryOp{
					Op:    operator.Multiply,
					Left:  &CILInputRef{Kind: "connector", Name: "base_value"},
					Right: &CILLiteral{Value: 2.0},
				},
				InputRefs: []CILInputRef{
					{Kind: "connector", Name: "base_value"},
				},
			},
			expectContains: []string{
				"func computedoubled(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"[]calculation.PostfixOp",
				`Kind: "push"`,
				`Kind: "apply"`,
				`Op: operator.Multiply`,
				`Name: "base_value"`,
				"Value: 2",
				"EvaluatePostfixFloat64",
				"controller.GetMetricComputation(\"doubled\").SetTensor(result)",
			},
			expectNotContains: []string{
				"data := result.Data()",
				"for i := range data",
				"TensorMultiplyFloat64ScalarFloat64",
				"DimensionNamesMatch",
				"var err error",
				"var result ",
			},
		},
		{
			name:       "unary negation of input",
			metricName: "negated",
			root: &CILMetricNode{
				CILNode: &CILUnaryOp{
					Op:      operator.Negate,
					Operand: &CILInputRef{Kind: "connector", Name: "value"},
				},
				InputRefs: []CILInputRef{
					{Kind: "connector", Name: "value"},
				},
			},
			expectContains: []string{
				"func computenegated(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"[]calculation.PostfixOp",
				`Kind: "push"`,
				`Kind: "apply"`,
				`Op: operator.Negate`,
				`Name: "value"`,
				"EvaluatePostfixFloat64",
				"controller.GetMetricComputation(\"negated\").SetTensor(result)",
			},
			expectNotContains: []string{
				"data := result.Data()",
				"for i := range data",
				"TensorNegateFloat64",
				"var err error",
				"var result ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := NewCILAbstractTransformer()
			if transformer == nil {
				t.Fatalf("NewCILAbstractTransformer() returned nil")
			}

			code, err := transformer.TransformMetric(tt.metricName, tt.root)
			if err != nil {
				t.Fatalf("TransformMetric() returned unexpected error: %v", err)
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(code, expected) {
					t.Errorf("TransformMetric() output missing expected content: %q\nGot:\n%s", expected, code)
				}
			}

			for _, notExpected := range tt.expectNotContains {
				if strings.Contains(code, notExpected) {
					t.Errorf("TransformMetric() output contains unexpected content: %q\nGot:\n%s", notExpected, code)
				}
			}
		})
	}
}

// TestTransformMetric_PostfixTableOutput verifies that CILBinaryOp and CILUnaryOp
// produce table-driven postfix output via generatePostfixComputeBody rather than
// the old per-operation visitor snippets. This is the RED-phase test: it asserts
// the new postfix patterns and therefore fails while TransformMetric still uses
// the old visitor approach.
func TestTransformMetric_PostfixTableOutput(t *testing.T) {
	tests := []struct {
		name              string
		metricName        string
		root              interface{}
		expectContains    []string
		expectNotContains []string
	}{
		{
			name:       "binary op subtract two connectors",
			metricName: "profit",
			root: &CILMetricNode{
				CILNode: &CILBinaryOp{
					Op:    operator.Subtract,
					Left:  &CILInputRef{Kind: "connector", Name: "revenue"},
					Right: &CILInputRef{Kind: "connector", Name: "cost"},
				},
				InputRefs: []CILInputRef{
					{Kind: "connector", Name: "revenue"},
					{Kind: "connector", Name: "cost"},
				},
			},
			expectContains: []string{
				"func computeprofit(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"[]calculation.PostfixOp",
				`Kind: "push"`,
				`Kind: "apply"`,
				`Op: operator.Subtract`,
				"EvaluatePostfixFloat64",
				"controller.GetMetricComputation(\"profit\").SetTensor(result)",
			},
			expectNotContains: []string{
				"TensorSubtractFloat64",
				"TensorDimensionNamesMatch",
				"var err error",
				"var result ",
			},
		},
		{
			name:       "unary op negate connector input",
			metricName: "negated",
			root: &CILMetricNode{
				CILNode: &CILUnaryOp{
					Op:      operator.Negate,
					Operand: &CILInputRef{Kind: "connector", Name: "value"},
				},
				InputRefs: []CILInputRef{
					{Kind: "connector", Name: "value"},
				},
			},
			expectContains: []string{
				"func computenegated(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"EvaluatePostfixFloat64",
				"operator.Negate",
				"controller.GetMetricComputation(\"negated\").SetTensor(result)",
			},
			expectNotContains: []string{
				"TensorNegateFloat64",
			},
		},
		{
			name:       "nested expression a - b / c",
			metricName: "complex",
			root: &CILMetricNode{
				CILNode: &CILBinaryOp{
					Op:   operator.Subtract,
					Left: &CILInputRef{Kind: "metric", Name: "a"},
					Right: &CILBinaryOp{
						Op:    operator.Divide,
						Left:  &CILInputRef{Kind: "metric", Name: "b"},
						Right: &CILInputRef{Kind: "metric", Name: "c"},
					},
				},
				InputRefs: []CILInputRef{
					{Kind: "metric", Name: "a"},
					{Kind: "metric", Name: "b"},
					{Kind: "metric", Name: "c"},
				},
			},
			expectContains: []string{
				"func computecomplex(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"[]calculation.PostfixOp",
				"EvaluatePostfixFloat64",
				"controller.GetMetricComputation(\"complex\").SetTensor(result)",
			},
			expectNotContains: []string{
				"var err error",
				"var result ",
				"TensorSubtractFloat64",
				"TensorDivideFloat64",
				"TensorDimensionNamesMatch",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := NewCILAbstractTransformer()
			if transformer == nil {
				t.Fatalf("NewCILAbstractTransformer() returned nil")
			}

			code, err := transformer.TransformMetric(tt.metricName, tt.root)
			if err != nil {
				t.Fatalf("TransformMetric() returned unexpected error: %v", err)
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(code, expected) {
					t.Errorf("TransformMetric() output missing expected content: %q\nGot:\n%s", expected, code)
				}
			}

			for _, notExpected := range tt.expectNotContains {
				if strings.Contains(code, notExpected) {
					t.Errorf("TransformMetric() output contains unexpected content: %q\nGot:\n%s", notExpected, code)
				}
			}
		})
	}
}

// TestTransformMetric_LiteralPostfixPath verifies that TransformMetric routes
// CILLiteral-only metrics through the appropriate path:
//   - Literal without dimension refs: postfix walker path (EvaluatePostfixFloat64).
//   - Literal with dimension refs: AllocateTensor + fill-loop path to preserve
//     dimension shape for downstream output writing.
func TestTransformMetric_LiteralPostfixPath(t *testing.T) {
	tests := []struct {
		name              string
		metricName        string
		root              interface{}
		expectContains    []string
		expectNotContains []string
	}{
		{
			name:       "literal metric without dimensions",
			metricName: "StaticValue",
			root: &CILMetricNode{
				CILNode: &CILLiteral{Value: 42.0},
			},
			expectContains: []string{
				"func computeStaticValue(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"_ = inputs",
				"EvaluatePostfixFloat64",
				"[]calculation.PostfixOp",
				"Value: 42",
				"controller.GetMetricComputation(\"StaticValue\").SetTensor(result)",
				"Failed to set tensor for StaticValue",
			},
			expectNotContains: []string{
				"calculation.AllocateTensor",
				"data := result.Data()",
				"for i := range data",
				"var err error",
				"var result ",
			},
		},
		{
			name:       "literal metric with dimensions",
			metricName: "DimensionalLiteral",
			root: &CILMetricNode{
				CILNode: &CILLiteral{Value: 99.0},
				InputRefs: []CILInputRef{
					{Kind: "dimension", Name: "Product"},
					{Kind: "dimension", Name: "Region"},
				},
			},
			expectContains: []string{
				"func computeDimensionalLiteral(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"_ = inputs",
				"calculation.AllocateTensor(controller, \"Product\", \"Region\")",
				"data := result.Data()",
				"for i := range data",
				"data[i] = float64(99)",
				"controller.GetMetricComputation(\"DimensionalLiteral\").SetTensor(result)",
			},
			expectNotContains: []string{
				"EvaluatePostfixFloat64",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := NewCILAbstractTransformer()
			if transformer == nil {
				t.Fatalf("NewCILAbstractTransformer() returned nil")
			}

			code, err := transformer.TransformMetric(tt.metricName, tt.root)
			if err != nil {
				t.Fatalf("TransformMetric() returned unexpected error: %v", err)
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(code, expected) {
					t.Errorf("TransformMetric() output missing expected content: %q\nGot:\n%s", expected, code)
				}
			}

			for _, notExpected := range tt.expectNotContains {
				if strings.Contains(code, notExpected) {
					t.Errorf("TransformMetric() output contains unexpected content: %q\nGot:\n%s", notExpected, code)
				}
			}
		})
	}
}
