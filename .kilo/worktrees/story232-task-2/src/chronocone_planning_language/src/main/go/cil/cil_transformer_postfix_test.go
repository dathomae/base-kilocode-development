// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"reflect"
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/calculation"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// TestWalkCILTreeToPostfix_SimpleBinaryOps verifies that simple binary operations
// with input reference operands produce the correct postfix sequence.
func TestWalkCILTreeToPostfix_SimpleBinaryOps(t *testing.T) {
	tests := []struct {
		name string
		op   operator.Operator
	}{
		{"addition", operator.Add},
		{"subtraction", operator.Subtract},
		{"multiplication", operator.Multiply},
		{"division", operator.Divide},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &CILBinaryOp{
				Op:    tt.op,
				Left:  &CILInputRef{Kind: "metric", Name: "input_a"},
				Right: &CILInputRef{Kind: "metric", Name: "input_b"},
			}
			ops, err := WalkCILTreeToPostfix(root)
			if err != nil {
				t.Fatalf("WalkCILTreeToPostfix() returned error: %v", err)
			}
			expected := []calculation.PostfixOp{
				{Kind: "push", Name: "input_a"},
				{Kind: "push", Name: "input_b"},
				{Kind: "apply", Op: tt.op},
			}
			if !reflect.DeepEqual(ops, expected) {
				t.Errorf("WalkCILTreeToPostfix() = %v, expected %v", ops, expected)
			}
		})
	}
}

// TestWalkCILTreeToPostfix_Nested verifies that nested and unary expressions
// produce correct postfix output with proper post-order traversal.
func TestWalkCILTreeToPostfix_Nested(t *testing.T) {
	tests := []struct {
		name     string
		root     CILNode
		expected []calculation.PostfixOp
	}{
		{
			name: "(a + b) * c",
			root: &CILBinaryOp{
				Op: operator.Multiply,
				Left: &CILBinaryOp{
					Op:    operator.Add,
					Left:  &CILInputRef{Kind: "metric", Name: "a"},
					Right: &CILInputRef{Kind: "metric", Name: "b"},
				},
				Right: &CILInputRef{Kind: "metric", Name: "c"},
			},
			expected: []calculation.PostfixOp{
				{Kind: "push", Name: "a"},
				{Kind: "push", Name: "b"},
				{Kind: "apply", Op: operator.Add},
				{Kind: "push", Name: "c"},
				{Kind: "apply", Op: operator.Multiply},
			},
		},
		{
			name: "a - b / c",
			root: &CILBinaryOp{
				Op:   operator.Subtract,
				Left: &CILInputRef{Kind: "metric", Name: "a"},
				Right: &CILBinaryOp{
					Op:    operator.Divide,
					Left:  &CILInputRef{Kind: "metric", Name: "b"},
					Right: &CILInputRef{Kind: "metric", Name: "c"},
				},
			},
			expected: []calculation.PostfixOp{
				{Kind: "push", Name: "a"},
				{Kind: "push", Name: "b"},
				{Kind: "push", Name: "c"},
				{Kind: "apply", Op: operator.Divide},
				{Kind: "apply", Op: operator.Subtract},
			},
		},
		{
			name: "-(a + b)",
			root: &CILUnaryOp{
				Op: operator.Negate,
				Operand: &CILBinaryOp{
					Op:    operator.Add,
					Left:  &CILInputRef{Kind: "metric", Name: "a"},
					Right: &CILInputRef{Kind: "metric", Name: "b"},
				},
			},
			expected: []calculation.PostfixOp{
				{Kind: "push", Name: "a"},
				{Kind: "push", Name: "b"},
				{Kind: "apply", Op: operator.Add},
				{Kind: "apply", Op: operator.Negate},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, err := WalkCILTreeToPostfix(tt.root)
			if err != nil {
				t.Fatalf("WalkCILTreeToPostfix() returned error: %v", err)
			}
			if !reflect.DeepEqual(ops, tt.expected) {
				t.Errorf("WalkCILTreeToPostfix() = %v, expected %v", ops, tt.expected)
			}
		})
	}
}

// TestWalkCILTreeToPostfix_ScalarLiteral verifies that scalar literal operands
// produce push ops with the Value field set (not Name, which is used for
// tensor and input references).
func TestWalkCILTreeToPostfix_ScalarLiteral(t *testing.T) {
	tests := []struct {
		name     string
		root     CILNode
		expected []calculation.PostfixOp
	}{
		{
			name: "simple literal 42.0",
			root: &CILLiteral{Value: 42.0},
			expected: []calculation.PostfixOp{
				{Kind: "push", Value: 42.0},
			},
		},
		{
			name: "input_a * 3.0",
			root: &CILBinaryOp{
				Op:   operator.Multiply,
				Left: &CILInputRef{Kind: "metric", Name: "input_a"},
				Right: &CILLiteral{
					Value: 3.0,
				},
			},
			expected: []calculation.PostfixOp{
				{Kind: "push", Name: "input_a"},
				{Kind: "push", Value: 3.0},
				{Kind: "apply", Op: operator.Multiply},
			},
		},
		{
			name: "2.0 + 3.0",
			root: &CILBinaryOp{
				Op:    operator.Add,
				Left:  &CILLiteral{Value: 2.0},
				Right: &CILLiteral{Value: 3.0},
			},
			expected: []calculation.PostfixOp{
				{Kind: "push", Value: 2.0},
				{Kind: "push", Value: 3.0},
				{Kind: "apply", Op: operator.Add},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, err := WalkCILTreeToPostfix(tt.root)
			if err != nil {
				t.Fatalf("WalkCILTreeToPostfix() returned error: %v", err)
			}
			if !reflect.DeepEqual(ops, tt.expected) {
				t.Errorf("WalkCILTreeToPostfix() = %v, expected %v", ops, tt.expected)
			}
		})
	}
}

// TestWalkCILTreeToPostfix_InputRefs verifies that input references of different
// kinds are preserved correctly in the postfix output.
func TestWalkCILTreeToPostfix_InputRefs(t *testing.T) {
	tests := []struct {
		name     string
		root     CILNode
		expected []calculation.PostfixOp
	}{
		{
			name: "metric kind sales",
			root: &CILInputRef{Kind: "metric", Name: "sales"},
			expected: []calculation.PostfixOp{
				{Kind: "push", Name: "sales"},
			},
		},
		{
			name: "dimension kind Product",
			root: &CILInputRef{Kind: "dimension", Name: "Product"},
			expected: []calculation.PostfixOp{
				{Kind: "push", Name: "Product"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, err := WalkCILTreeToPostfix(tt.root)
			if err != nil {
				t.Fatalf("WalkCILTreeToPostfix() returned error: %v", err)
			}
			if !reflect.DeepEqual(ops, tt.expected) {
				t.Errorf("WalkCILTreeToPostfix() = %v, expected %v", ops, tt.expected)
			}
		})
	}
}

// TestWalkCILTreeToPostfix_ErrorCases verifies error handling for invalid inputs.
func TestWalkCILTreeToPostfix_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		root        CILNode
		expectError bool
	}{
		{
			name:        "nil node",
			root:        nil,
			expectError: true,
		},
		{
			name:        "unsupported node type",
			root:        &mockCILNode{},
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, err := WalkCILTreeToPostfix(tt.root)
			if tt.expectError {
				if err == nil {
					t.Errorf("WalkCILTreeToPostfix() returned nil error, expected non-nil error")
				}
				if ops != nil {
					t.Errorf("WalkCILTreeToPostfix() returned non-nil ops %v on error, expected nil", ops)
				}
				return
			}
			if err != nil {
				t.Fatalf("WalkCILTreeToPostfix() returned unexpected error: %v", err)
			}
		})
	}
}

// TestGeneratePostfixTableLiteral verifies that GeneratePostfixTableLiteral
// produces valid Go struct literal output for various PostfixOp slice inputs.
func TestGeneratePostfixTableLiteral(t *testing.T) {
	tests := []struct {
		name              string
		ops               []calculation.PostfixOp
		expectContains    []string
		expectNotContains []string
	}{
		{
			name: "empty ops slice",
			ops:  []calculation.PostfixOp{},
			expectContains: []string{
				"[]calculation.PostfixOp{}",
			},
			expectNotContains: []string{
				"Kind:",
				"Op:",
				"Name:",
			},
		},
		{
			name: "single push op",
			ops: []calculation.PostfixOp{
				{Kind: "push", Name: "sales"},
			},
			expectContains: []string{
				"[]calculation.PostfixOp{",
				`{Kind: "push", Name: "sales"}`,
			},
			expectNotContains: []string{
				"[]PostfixOp{",
			},
		},
		{
			name: "single apply op",
			ops: []calculation.PostfixOp{
				{Kind: "apply", Op: operator.Add},
			},
			expectContains: []string{
				"[]calculation.PostfixOp{",
				`{Kind: "apply", Op: operator.Add}`,
			},
			expectNotContains: []string{
				"[]PostfixOp{",
			},
		},
		{
			name: "push-push-apply combo",
			ops: []calculation.PostfixOp{
				{Kind: "push", Name: "input_a"},
				{Kind: "push", Name: "input_b"},
				{Kind: "apply", Op: operator.Subtract},
			},
			expectContains: []string{
				"[]calculation.PostfixOp{",
				`{Kind: "push", Name: "input_a"}`,
				`{Kind: "push", Name: "input_b"}`,
				`{Kind: "apply", Op: operator.Subtract}`,
			},
			expectNotContains: []string{
				"[]PostfixOp{",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := NewCILAbstractTransformer()
			if transformer == nil {
				t.Fatalf("NewCILAbstractTransformer() returned nil")
			}

			code := transformer.GeneratePostfixTableLiteral(tt.ops)

			for _, expected := range tt.expectContains {
				if !strings.Contains(code, expected) {
					t.Errorf("GeneratePostfixTableLiteral() output missing expected content: %q\nGot: %q", expected, code)
				}
			}

			for _, notExpected := range tt.expectNotContains {
				if strings.Contains(code, notExpected) {
					t.Errorf("GeneratePostfixTableLiteral() output contains unexpected content: %q\nGot: %q", notExpected, code)
				}
			}
		})
	}
}

// TestGeneratePostfixTableLiteral_Formatting verifies formatting correctness
// of the generated Go struct literal output, including package-qualified type
// names, proper brace structure, quoted strings, and absence of Go syntax errors.
func TestGeneratePostfixTableLiteral_Formatting(t *testing.T) {
	transformer := NewCILAbstractTransformer()
	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}

	ops := []calculation.PostfixOp{
		{Kind: "push", Name: "sales"},
		{Kind: "push", Name: "costs"},
		{Kind: "apply", Op: operator.Subtract},
	}

	code := transformer.GeneratePostfixTableLiteral(ops)

	// Verify package-qualified type name is used
	if !strings.Contains(code, "calculation.PostfixOp") {
		t.Errorf("GeneratePostfixTableLiteral() output missing package-qualified type 'calculation.PostfixOp': %q", code)
	}

	// Verify proper brace structure
	if !strings.Contains(code, "{") {
		t.Errorf("GeneratePostfixTableLiteral() output missing opening brace: %q", code)
	}

	// Verify no unqualified PostfixOp (without calculation. prefix)
	if strings.Contains(code, "[]PostfixOp{") {
		t.Errorf("GeneratePostfixTableLiteral() output contains unqualified '[]PostfixOp{': %q", code)
	}

	// Verify no bare PostfixOp struct literal without package qualification.
	// The pattern "}{PostfixOp{" would appear if a struct literal were emitted
	// without the "calculation." prefix, which would be invalid.
	if strings.Contains(code, "}{PostfixOp{") {
		t.Errorf("GeneratePostfixTableLiteral() output contains unqualified struct literal 'PostfixOp{': %q", code)
	}

	// Verify string values are properly quoted
	if !strings.Contains(code, `"push"`) {
		t.Errorf("GeneratePostfixTableLiteral() output missing quoted string %q: %q", `"push"`, code)
	}
	if !strings.Contains(code, `"apply"`) {
		t.Errorf("GeneratePostfixTableLiteral() output missing quoted string %q: %q", `"apply"`, code)
	}
	if !strings.Contains(code, `"sales"`) {
		t.Errorf("GeneratePostfixTableLiteral() output missing quoted string %q: %q", `"sales"`, code)
	}

	// Verify Go field notation uses proper syntax (Kind:, Op:, Name:)
	if !strings.Contains(code, "Kind:") {
		t.Errorf("GeneratePostfixTableLiteral() output missing 'Kind:' field: %q", code)
	}
	if !strings.Contains(code, "Name:") {
		t.Errorf("GeneratePostfixTableLiteral() output missing 'Name:' field: %q", code)
	}
	if !strings.Contains(code, "Op:") {
		t.Errorf("GeneratePostfixTableLiteral() output missing 'Op:' field: %q", code)
	}
}

// TestGenerateConnectorRegistrationCode verifies that
// generateConnectorRegistrationCode returns a non-empty code string that
// registers connector inputs as MetricComputation objects on the controller.
func TestGenerateConnectorRegistrationCode(t *testing.T) {
	transformer := NewCILAbstractTransformer()
	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}

	code := transformer.generateConnectorRegistrationCode()

	if code == "" {
		t.Errorf("generateConnectorRegistrationCode() returned empty string, expected non-empty")
	}
	if !strings.Contains(code, "for name, tensor := range inputs") {
		t.Errorf("generateConnectorRegistrationCode() missing expected loop: %q", "for name, tensor := range inputs")
	}
	if !strings.Contains(code, "controller.RegisterMetricComputation") {
		t.Errorf("generateConnectorRegistrationCode() missing expected registration call: %q", "controller.RegisterMetricComputation")
	}
	if !strings.Contains(code, "calculation.NewMetricComputation(0)") {
		t.Errorf("generateConnectorRegistrationCode() missing expected MetricComputation construction: %q", "calculation.NewMetricComputation(0)")
	}
	if !strings.Contains(code, "SetTensor") {
		t.Errorf("generateConnectorRegistrationCode() missing expected SetTensor call: %q", "SetTensor")
	}
}

// TestGeneratePostfixComputeBody_DoesNotContainConnectorRegistration verifies that
// the compute body generated by generatePostfixComputeBody does NOT include the
// connector registration block (it was moved to GenerateComputeFunction to avoid
// data races when metrics run in parallel goroutines).
func TestGeneratePostfixComputeBody_DoesNotContainConnectorRegistration(t *testing.T) {
	transformer := NewCILAbstractTransformer()
	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}

	root := &CILMetricNode{
		CILNode: &CILBinaryOp{
			Op:    operator.Subtract,
			Left:  &CILInputRef{Kind: "connector", Name: "revenue_amt"},
			Right: &CILInputRef{Kind: "connector", Name: "cost_amt"},
		},
		InputRefs: []CILInputRef{
			{Kind: "connector", Name: "revenue_amt"},
			{Kind: "connector", Name: "cost_amt"},
		},
	}

	code, err := transformer.generatePostfixComputeBody("profit", root)
	if err != nil {
		t.Fatalf("generatePostfixComputeBody() returned unexpected error: %v", err)
	}

	if strings.Contains(code, "for name, tensor := range inputs") {
		t.Errorf("generatePostfixComputeBody() should NOT contain connector registration loop")
	}

	// Verify existing postfix patterns are still present
	if !strings.Contains(code, "[]calculation.PostfixOp") {
		t.Errorf("generatePostfixComputeBody() missing postfix ops literal")
	}
	if !strings.Contains(code, "controller.GetMetricComputation(\"profit\").SetTensor(result)") {
		t.Errorf("generatePostfixComputeBody() missing SetTensor for profit metric")
	}
}

// TestWalkCILTreeToPostfix_LookupRef verifies that lookup references produce
// correct postfix push operations with the lookup name.
func TestWalkCILTreeToPostfix_LookupRef(t *testing.T) {
	tests := []struct {
		name     string
		root     CILNode
		expected []calculation.PostfixOp
	}{
		{
			name: "simple lookup strength_factor",
			root: &CILLookupNode{LookupName: "strength_factor"},
			expected: []calculation.PostfixOp{
				{Kind: "push", Name: "strength_factor"},
			},
		},
		{
			name: "connector_ref * lookup_ref",
			root: &CILBinaryOp{
				Op:   operator.Multiply,
				Left: &CILInputRef{Kind: "connector", Name: "connector_ref"},
				Right: &CILLookupNode{
					LookupName: "lookup_ref",
				},
			},
			expected: []calculation.PostfixOp{
				{Kind: "push", Name: "connector_ref"},
				{Kind: "push", Name: "lookup_ref"},
				{Kind: "apply", Op: operator.Multiply},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, err := WalkCILTreeToPostfix(tt.root)
			if err != nil {
				t.Fatalf("WalkCILTreeToPostfix() returned error: %v", err)
			}
			if !reflect.DeepEqual(ops, tt.expected) {
				t.Errorf("WalkCILTreeToPostfix() = %v, expected %v", ops, tt.expected)
			}
		})
	}
}

// TestWalkCILTreeToPostfix_InputTimesLookup verifies that a binary expression
// with an input reference and a lookup reference produces the correct postfix
// sequence.
func TestWalkCILTreeToPostfix_InputTimesLookup(t *testing.T) {
	root := &CILBinaryOp{
		Op:    operator.Multiply,
		Left:  &CILInputRef{Kind: "metric", Name: "connector_ref"},
		Right: &CILLookupNode{LookupName: "strength_factor"},
	}
	ops, err := WalkCILTreeToPostfix(root)
	if err != nil {
		t.Fatalf("WalkCILTreeToPostfix() returned error: %v", err)
	}
	expected := []calculation.PostfixOp{
		{Kind: "push", Name: "connector_ref"},
		{Kind: "push", Name: "strength_factor"},
		{Kind: "apply", Op: operator.Multiply},
	}
	if !reflect.DeepEqual(ops, expected) {
		t.Errorf("WalkCILTreeToPostfix() = %v, expected %v", ops, expected)
	}
}

// mockCILNode is a test-only type that implements CILNode but is not *CILLiteral.
type mockCILNode struct{}

func (m *mockCILNode) cilNode() {}

// TestTransformMetric_LiteralPostfix verifies that TransformMetric routes
// CILLiteral-only metrics through the appropriate path:
//   - Literal without dimension refs: postfix walker path (EvaluatePostfixFloat64).
//   - Literal with dimension refs: AllocateTensor + fill-loop path to preserve
//     dimension shape for downstream output writing.
func TestTransformMetric_LiteralPostfix(t *testing.T) {
	tests := []struct {
		name              string
		metricName        string
		root              interface{}
		expectContains    []string
		expectNotContains []string
	}{
		{
			name:       "literal 42 without dimensions",
			metricName: "StaticValue",
			root: &CILMetricNode{
				CILNode: &CILLiteral{Value: 42.0},
			},
			expectContains: []string{
				"func computeStaticValue(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"_ = inputs",
				"[]calculation.PostfixOp",
				"EvaluatePostfixFloat64",
				"controller.GetMetricComputation(\"StaticValue\").SetTensor(result)",
				"Value: 42",
			},
			expectNotContains: []string{
				"calculation.AllocateTensor",
				"data := result.Data()",
				"for i := range data",
				"data[i] = float64",
				"var err error",
				"var result ",
			},
		},
		{
			name:       "literal -3.14 without dimensions",
			metricName: "NegValue",
			root: &CILMetricNode{
				CILNode: &CILLiteral{Value: -3.14},
			},
			expectContains: []string{
				"EvaluatePostfixFloat64",
				"Value: -3.14",
				"_ = inputs",
			},
			expectNotContains: []string{
				"calculation.AllocateTensor",
				"data := result.Data()",
				"for i := range data",
			},
		},
		{
			name:       "literal with dimension input refs",
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
