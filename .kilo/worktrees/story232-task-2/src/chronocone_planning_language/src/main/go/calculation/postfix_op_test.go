// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package calculation

import (
	"math"
	"testing"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// assertPostfixResult validates the result of EvaluatePostfixFloat64.
// In Phase B the implementation is a stub; success-case subtests will fail
// (red phase) and error-case subtests expect non-nil errors.
func assertPostfixResult(t *testing.T, result *ChronoconeTensor[any], err error, expectedData []float64, expectError bool, expectedDims []string, expectedShape []int) {
	t.Helper()

	if expectError {
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		return
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatalf("expected non-nil result, got nil")
	}

	dims := result.DimensionNames()
	if len(dims) != len(expectedDims) {
		t.Errorf("DimensionNames() length = %d, expected %d", len(dims), len(expectedDims))
	} else {
		for i, name := range expectedDims {
			if dims[i] != name {
				t.Errorf("DimensionNames()[%d] = %q, expected %q", i, dims[i], name)
			}
		}
	}

	shape := result.Shape()
	if len(shape) != len(expectedShape) {
		t.Errorf("Shape() length = %d, expected %d", len(shape), len(expectedShape))
	} else {
		for i, s := range expectedShape {
			if shape[i] != s {
				t.Errorf("Shape()[%d] = %d, expected %d", i, shape[i], s)
			}
		}
	}

	data := result.Data()
	if len(data) != len(expectedData) {
		t.Fatalf("Data() length = %d, expected %d", len(data), len(expectedData))
	}
	resultFloats := make([]float64, len(data))
	for i, v := range data {
		f, ok := v.(float64)
		if !ok {
			t.Fatalf("Data()[%d] is not float64: %T", i, v)
		}
		resultFloats[i] = f
	}
	if !float64SlicesEqual(resultFloats, expectedData, 1e-9) {
		t.Errorf("Data() = %v, expected %v", resultFloats, expectedData)
	}
}

// setupPostfixController creates a controller with concept values and registers
// named tensors as metric computations so EvaluatePostfixFloat64 can resolve
// push names.
func setupPostfixController(t *testing.T, namedTensors map[string]*ChronoconeTensor[any]) *ChronoconeComputationController {
	t.Helper()
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{
		"Time":    {"2024-01", "2024-02"},
		"Product": {"A", "B"},
	})

	for name, tensor := range namedTensors {
		mc := NewMetricComputation(1)
		if err := controller.RegisterMetricComputation(name, mc); err != nil {
			t.Fatalf("RegisterMetricComputation(%q): %v", name, err)
		}
		if err := mc.SetTensor(tensor); err != nil {
			t.Fatalf("SetTensor(%q): %v", name, err)
		}
	}

	return controller
}

// allocateFloat64Tensor allocates a tensor via AllocateTensor, populates its
// data with the supplied float64 values, registers a MetricComputation under
// name, and sets the tensor on it.
func allocateFloat64Tensor(t *testing.T, controller *ChronoconeComputationController, name string, dimNames []string, values []float64) {
	t.Helper()
	tensor, err := AllocateTensor(controller, dimNames...)
	if err != nil {
		t.Fatalf("AllocateTensor(%q): %v", name, err)
	}
	data := tensor.Data()
	if len(data) != len(values) {
		t.Fatalf("allocated tensor size %d does not match values length %d", len(data), len(values))
	}
	for i, v := range values {
		data[i] = v
	}
	mc := NewMetricComputation(1)
	if err := controller.RegisterMetricComputation(name, mc); err != nil {
		t.Fatalf("RegisterMetricComputation(%q): %v", name, err)
	}
	if err := mc.SetTensor(tensor); err != nil {
		t.Fatalf("SetTensor(%q): %v", name, err)
	}
}

// TestEvaluatePostfixFloat64_SingleBinaryOps verifies each of the four
// supported binary operators with a single push-push-apply sequence.
func TestEvaluatePostfixFloat64_SingleBinaryOps(t *testing.T) {
	tests := []struct {
		name          string
		op            operator.Operator
		a             []float64
		b             []float64
		expected      []float64
		expectedDims  []string
		expectedShape []int
	}{
		{
			name:          "addition",
			op:            operator.Add,
			a:             []float64{1.0, 2.0},
			b:             []float64{3.0, 4.0},
			expected:      []float64{4.0, 6.0},
			expectedDims:  []string{"Time"},
			expectedShape: []int{2},
		},
		{
			name:          "subtraction",
			op:            operator.Subtract,
			a:             []float64{5.0, 10.0},
			b:             []float64{3.0, 4.0},
			expected:      []float64{2.0, 6.0},
			expectedDims:  []string{"Time"},
			expectedShape: []int{2},
		},
		{
			name:          "multiplication",
			op:            operator.Multiply,
			a:             []float64{2.0, 3.0},
			b:             []float64{4.0, 5.0},
			expected:      []float64{8.0, 15.0},
			expectedDims:  []string{"Time"},
			expectedShape: []int{2},
		},
		{
			name:          "division",
			op:            operator.Divide,
			a:             []float64{10.0, 20.0},
			b:             []float64{2.0, 5.0},
			expected:      []float64{5.0, 4.0},
			expectedDims:  []string{"Time"},
			expectedShape: []int{2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewChronoconeComputationController()
			controller.SetConceptValues(map[string][]string{
				"Time": {"2024-01", "2024-02"},
			})
			allocateFloat64Tensor(t, controller, "a", []string{"Time"}, tt.a)
			allocateFloat64Tensor(t, controller, "b", []string{"Time"}, tt.b)

			ops := []PostfixOp{
				{Kind: "push", Name: "a"},
				{Kind: "push", Name: "b"},
				{Kind: "apply", Op: tt.op},
			}

			result, err := EvaluatePostfixFloat64(controller, ops)
			assertPostfixResult(t, result, err, tt.expected, false, tt.expectedDims, tt.expectedShape)
		})
	}
}

// TestEvaluatePostfixFloat64_NestedOps verifies a multi-op expression
// equivalent to (a + b) * c.
func TestEvaluatePostfixFloat64_NestedOps(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{
		"Time": {"2024-01", "2024-02"},
	})
	allocateFloat64Tensor(t, controller, "a", []string{"Time"}, []float64{1.0, 2.0})
	allocateFloat64Tensor(t, controller, "b", []string{"Time"}, []float64{3.0, 4.0})
	allocateFloat64Tensor(t, controller, "c", []string{"Time"}, []float64{2.0, 2.0})

	ops := []PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "push", Name: "b"},
		{Kind: "apply", Op: operator.Add},
		{Kind: "push", Name: "c"},
		{Kind: "apply", Op: operator.Multiply},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	// (1+3)*2 = 8, (2+4)*2 = 12
	assertPostfixResult(t, result, err, []float64{8.0, 12.0}, false, []string{"Time"}, []int{2})
}

// TestEvaluatePostfixFloat64_DimensionMismatch verifies that applying a
// binary operator to tensors with incompatible dimensions produces an error.
func TestEvaluatePostfixFloat64_DimensionMismatch(t *testing.T) {
	controller := setupPostfixController(t, map[string]*ChronoconeTensor[any]{
		"a": NewChronoconeTensor([]any{1.0, 2.0}, []string{"Time"}, []int{2}),
		"b": NewChronoconeTensor([]any{10.0, 20.0}, []string{"Product"}, []int{2}),
	})

	ops := []PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "push", Name: "b"},
		{Kind: "apply", Op: operator.Add},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err == nil {
		t.Errorf("expected error for dimension mismatch, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result for dimension mismatch, got non-nil")
	}
}

// TestEvaluatePostfixFloat64_EmptyTable verifies that a zero-length ops
// slice returns an error.
func TestEvaluatePostfixFloat64_EmptyTable(t *testing.T) {
	controller := NewChronoconeComputationController()
	ops := []PostfixOp{}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err == nil {
		t.Errorf("expected error for empty ops table, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result for empty ops table, got non-nil")
	}
}

// TestEvaluatePostfixFloat64_StackUnderflow verifies that an apply
// operation without enough operands on the stack returns an error.
func TestEvaluatePostfixFloat64_StackUnderflow(t *testing.T) {
	controller := setupPostfixController(t, map[string]*ChronoconeTensor[any]{
		"a": NewChronoconeTensor([]any{1.0}, []string{"Time"}, []int{1}),
	})

	ops := []PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "apply", Op: operator.Add},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err == nil {
		t.Errorf("expected error for stack underflow, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result for stack underflow, got non-nil")
	}
}

// TestEvaluatePostfixFloat64_UnknownOperator verifies that an unsupported
// operator in an apply op returns an error.
func TestEvaluatePostfixFloat64_UnknownOperator(t *testing.T) {
	controller := setupPostfixController(t, map[string]*ChronoconeTensor[any]{
		"a": NewChronoconeTensor([]any{1.0}, []string{"Time"}, []int{1}),
		"b": NewChronoconeTensor([]any{2.0}, []string{"Time"}, []int{1}),
	})

	ops := []PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "push", Name: "b"},
		{Kind: "apply", Op: operator.Unknown},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err == nil {
		t.Errorf("expected error for unknown operator, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result for unknown operator, got non-nil")
	}
}

// TestEvaluatePostfixFloat64_MultiplyWithLookup verifies arithmetic broadcasting:
// a 3D connector tensor multiplied by a 1D lookup tensor broadcasts the lookup
// across missing dimensions.
func TestEvaluatePostfixFloat64_MultiplyWithLookup(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{
		"Product":  {"A", "B", "C"},
		"Material": {"Steel", "Aluminum", "Copper"},
		"Daily":    {"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
	})

	// Connector tensor: [Product:3, Material:3, Daily:7] — all 1.0
	connectorData := make([]float64, 3*3*7)
	for i := range connectorData {
		connectorData[i] = 1.0
	}
	allocateFloat64Tensor(t, controller, "connector", []string{"Product", "Material", "Daily"}, connectorData)

	// Lookup tensor: [Material:3] — Steel=1.0, Aluminum=0.8, Copper=0.9
	allocateFloat64Tensor(t, controller, "lookup", []string{"Material"}, []float64{1.0, 0.8, 0.9})

	ops := []PostfixOp{
		{Kind: "push", Name: "connector"},
		{Kind: "push", Name: "lookup"},
		{Kind: "apply", Op: operator.Multiply},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatalf("expected non-nil result, got nil")
	}

	// Verify dimensions and shape.
	dims := result.DimensionNames()
	if len(dims) != 3 || dims[0] != "Product" || dims[1] != "Material" || dims[2] != "Daily" {
		t.Errorf("DimensionNames() = %v, want [Product Material Daily]", dims)
	}
	shape := result.Shape()
	if len(shape) != 3 || shape[0] != 3 || shape[1] != 3 || shape[2] != 7 {
		t.Errorf("Shape() = %v, want [3 3 7]", shape)
	}

	// Verify that for each Product and Daily, Material values are multiplied correctly.
	data := result.Data()
	for product := 0; product < 3; product++ {
		for material := 0; material < 3; material++ {
			for day := 0; day < 7; day++ {
				flatIdx := product*21 + material*7 + day
				var expected float64
				switch material {
				case 0:
					expected = 1.0 // Steel
				case 1:
					expected = 0.8 // Aluminum
				case 2:
					expected = 0.9 // Copper
				}
				got, ok := data[flatIdx].(float64)
				if !ok {
					t.Fatalf("result.Data()[%d] is not float64: %T", flatIdx, data[flatIdx])
				}
				if math.Abs(got-expected) > 1e-9 {
					t.Errorf("data[%d] (Product=%d, Material=%d, Day=%d) = %v, want %v",
						flatIdx, product, material, day, got, expected)
				}
			}
		}
	}

	// Verify mask: all elements should be valid since both operands have all-true masks.
	for i := 0; i < len(data); i++ {
		if !result.IsMaskedTrue(i) {
			t.Errorf("IsMaskedTrue(%d) = false, want true", i)
		}
	}
}

// TestEvaluatePostfixFloat64_ComparisonWithLookup verifies comparison broadcasting:
// a 3D connector tensor compared against a 1D lookup tensor broadcasts the lookup
// and produces the correct mask.
func TestEvaluatePostfixFloat64_ComparisonWithLookup(t *testing.T) {
	controller := NewChronoconeComputationController()
	controller.SetConceptValues(map[string][]string{
		"Product":  {"A", "B", "C"},
		"Material": {"Steel", "Aluminum", "Copper"},
		"Daily":    {"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
	})

	// Connector tensor: [Product:3, Material:3, Daily:7]
	// Steel=5.0, Aluminum=3.0, Copper=1.0
	connectorData := make([]float64, 3*3*7)
	for product := 0; product < 3; product++ {
		for material := 0; material < 3; material++ {
			for day := 0; day < 7; day++ {
				flatIdx := product*21 + material*7 + day
				switch material {
				case 0:
					connectorData[flatIdx] = 5.0 // Steel
				case 1:
					connectorData[flatIdx] = 3.0 // Aluminum
				case 2:
					connectorData[flatIdx] = 1.0 // Copper
				}
			}
		}
	}
	allocateFloat64Tensor(t, controller, "connector", []string{"Product", "Material", "Daily"}, connectorData)

	// Lookup tensor: [Material:3] — threshold 2.0 for all materials
	allocateFloat64Tensor(t, controller, "lookup", []string{"Material"}, []float64{2.0, 2.0, 2.0})

	ops := []PostfixOp{
		{Kind: "push", Name: "connector"},
		{Kind: "push", Name: "lookup"},
		{Kind: "apply", Op: operator.GreaterThan},
	}

	result, err := EvaluatePostfixFloat64(controller, ops)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatalf("expected non-nil result, got nil")
	}

	// Verify dimensions and shape (should match the larger/left operand).
	dims := result.DimensionNames()
	if len(dims) != 3 || dims[0] != "Product" || dims[1] != "Material" || dims[2] != "Daily" {
		t.Errorf("DimensionNames() = %v, want [Product Material Daily]", dims)
	}
	shape := result.Shape()
	if len(shape) != 3 || shape[0] != 3 || shape[1] != 3 || shape[2] != 7 {
		t.Errorf("Shape() = %v, want [3 3 7]", shape)
	}

	// Verify mask: Steel (5>2)=valid, Aluminum (3>2)=valid, Copper (1>2)=invalid.
	for product := 0; product < 3; product++ {
		for material := 0; material < 3; material++ {
			for day := 0; day < 7; day++ {
				flatIdx := product*21 + material*7 + day
				var expectedValid bool
				switch material {
				case 0: // Steel
					expectedValid = true
				case 1: // Aluminum
					expectedValid = true
				case 2: // Copper
					expectedValid = false
				}
				gotValid := result.IsMaskedTrue(flatIdx)
				if gotValid != expectedValid {
					t.Errorf("IsMaskedTrue(%d) (Product=%d, Material=%d, Day=%d) = %v, want %v",
						flatIdx, product, material, day, gotValid, expectedValid)
				}
			}
		}
	}
}

// TestEvaluatePostfixFloat64_NilInputs verifies error handling when either
// the controller or the ops slice is nil.
func TestEvaluatePostfixFloat64_NilInputs(t *testing.T) {
	controller := setupPostfixController(t, map[string]*ChronoconeTensor[any]{
		"a": NewChronoconeTensor([]any{1.0}, []string{"Time"}, []int{1}),
	})

	t.Run("nil controller", func(t *testing.T) {
		ops := []PostfixOp{
			{Kind: "push", Name: "a"},
		}
		result, err := EvaluatePostfixFloat64(nil, ops)
		if err == nil {
			t.Errorf("expected error for nil controller, got nil")
		}
		if result != nil {
			t.Errorf("expected nil result for nil controller, got non-nil")
		}
	})

	t.Run("nil ops", func(t *testing.T) {
		result, err := EvaluatePostfixFloat64(controller, nil)
		if err == nil {
			t.Errorf("expected error for nil ops, got nil")
		}
		if result != nil {
			t.Errorf("expected nil result for nil ops, got non-nil")
		}
	})
}
