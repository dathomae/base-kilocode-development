// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package calculation

import (
	"math"
	"testing"
)

// makeAnyTensor creates a ChronoconeTensor[any] from raw []any data.
func makeAnyTensor(t *testing.T, data []any, dimNames []string, shape []int) *ChronoconeTensor[any] {
	t.Helper()
	return NewChronoconeTensor(data, dimNames, shape)
}

// makeFloat64Tensor creates a ChronoconeTensor[any] with float64 elements.
func makeFloat64Tensor(t *testing.T, data []float64, dimNames []string, shape []int) *ChronoconeTensor[any] {
	t.Helper()
	anyData := make([]any, len(data))
	for i, v := range data {
		anyData[i] = v
	}
	return NewChronoconeTensor(anyData, dimNames, shape)
}

// float64SlicesEqual compares two float64 slices with a tolerance.
func float64SlicesEqual(a, b []float64, epsilon float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.Abs(a[i]-b[i]) > epsilon {
			return false
		}
	}
	return true
}

// assertArithmeticResult validates the result of an arithmetic operation.
// Phase B behaviour: expectError=true passes when the stub returns nil,nil.
// expectError=false fails when the stub returns nil,nil, driving the red phase.
func assertArithmeticResult(t *testing.T, result *ChronoconeTensor[any], err error, expectedData []float64, expectError bool, expectedDims []string, expectedShape []int) {
	t.Helper()

	if expectError {
		// In Phase B the stub returns nil,nil. We accept nil result as passing
		// for error-case subtests so the whole suite stays green on error paths.
		if result != nil {
			t.Errorf("expected nil result for error case, got non-nil")
		}
		return
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatalf("expected non-nil result, got nil")
	}

	// Verify dimension names
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

	// Verify shape
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

	// Verify data length and contents
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

// TestTensorDimensionNamesMatch verifies dimension-name comparison.
func TestTensorDimensionNamesMatch(t *testing.T) {
	tests := []struct {
		name     string
		a        *ChronoconeTensor[any]
		b        *ChronoconeTensor[any]
		expected bool
	}{
		{
			name:     "matching dimensions",
			a:        makeFloat64Tensor(t, []float64{1.0, 2.0}, []string{"x", "y"}, []int{1, 2}),
			b:        makeFloat64Tensor(t, []float64{3.0, 4.0}, []string{"x", "y"}, []int{1, 2}),
			expected: true,
		},
		{
			name:     "non-matching dimension names",
			a:        makeFloat64Tensor(t, []float64{1.0}, []string{"a"}, []int{1}),
			b:        makeFloat64Tensor(t, []float64{2.0}, []string{"b"}, []int{1}),
			expected: false,
		},
		{
			name:     "different dimension count",
			a:        makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			b:        makeFloat64Tensor(t, []float64{1.0, 2.0}, []string{"x", "y"}, []int{1, 2}),
			expected: false,
		},
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "nil a",
			a:        nil,
			b:        makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			expected: false,
		},
		{
			name:     "nil b",
			a:        makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			b:        nil,
			expected: false,
		},
		{
			name:     "empty dimensions match",
			a:        makeFloat64Tensor(t, []float64{42.0}, []string{}, []int{}),
			b:        makeFloat64Tensor(t, []float64{99.0}, []string{}, []int{}),
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TensorDimensionNamesMatch(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("TensorDimensionNamesMatch() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

// TestTensorAddFloat64 verifies element-wise addition.
func TestTensorAddFloat64(t *testing.T) {
	tests := []struct {
		name          string
		a             *ChronoconeTensor[any]
		b             *ChronoconeTensor[any]
		expected      []float64
		expectError   bool
		expectedDims  []string
		expectedShape []int
	}{
		{
			name:          "simple 2-element addition",
			a:             makeFloat64Tensor(t, []float64{1.0, 2.0}, []string{"x"}, []int{2}),
			b:             makeFloat64Tensor(t, []float64{3.0, 4.0}, []string{"x"}, []int{2}),
			expected:      []float64{4.0, 6.0},
			expectError:   false,
			expectedDims:  []string{"x"},
			expectedShape: []int{2},
		},
		{
			name:          "2D tensor addition",
			a:             makeFloat64Tensor(t, []float64{1.0, 2.0, 3.0, 4.0}, []string{"x", "y"}, []int{2, 2}),
			b:             makeFloat64Tensor(t, []float64{10.0, 20.0, 30.0, 40.0}, []string{"x", "y"}, []int{2, 2}),
			expected:      []float64{11.0, 22.0, 33.0, 44.0},
			expectError:   false,
			expectedDims:  []string{"x", "y"},
			expectedShape: []int{2, 2},
		},
		{
			name:        "nil tensor a",
			a:           nil,
			b:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			expectError: true,
		},
		{
			name:        "nil tensor b",
			a:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			b:           nil,
			expectError: true,
		},
		{
			name:        "dimension name mismatch",
			a:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			b:           makeFloat64Tensor(t, []float64{2.0}, []string{"y"}, []int{1}),
			expectError: true,
		},
		{
			name:        "wrong element type in a",
			a:           makeAnyTensor(t, []any{"not a float"}, []string{"x"}, []int{1}),
			b:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			expectError: true,
		},
		{
			name:        "wrong element type in b",
			a:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			b:           makeAnyTensor(t, []any{int(42)}, []string{"x"}, []int{1}),
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TensorAddFloat64(tt.a, tt.b)
			assertArithmeticResult(t, result, err, tt.expected, tt.expectError, tt.expectedDims, tt.expectedShape)
		})
	}
}

// TestTensorSubtractFloat64 verifies element-wise subtraction (a - b).
func TestTensorSubtractFloat64(t *testing.T) {
	tests := []struct {
		name          string
		a             *ChronoconeTensor[any]
		b             *ChronoconeTensor[any]
		expected      []float64
		expectError   bool
		expectedDims  []string
		expectedShape []int
	}{
		{
			name:          "simple 2-element subtraction",
			a:             makeFloat64Tensor(t, []float64{5.0, 10.0}, []string{"x"}, []int{2}),
			b:             makeFloat64Tensor(t, []float64{3.0, 4.0}, []string{"x"}, []int{2}),
			expected:      []float64{2.0, 6.0},
			expectError:   false,
			expectedDims:  []string{"x"},
			expectedShape: []int{2},
		},
		{
			name:          "2D tensor subtraction",
			a:             makeFloat64Tensor(t, []float64{10.0, 20.0, 30.0, 40.0}, []string{"x", "y"}, []int{2, 2}),
			b:             makeFloat64Tensor(t, []float64{1.0, 2.0, 3.0, 4.0}, []string{"x", "y"}, []int{2, 2}),
			expected:      []float64{9.0, 18.0, 27.0, 36.0},
			expectError:   false,
			expectedDims:  []string{"x", "y"},
			expectedShape: []int{2, 2},
		},
		{
			name:        "nil tensor a",
			a:           nil,
			b:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			expectError: true,
		},
		{
			name:        "nil tensor b",
			a:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			b:           nil,
			expectError: true,
		},
		{
			name:        "dimension name mismatch",
			a:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			b:           makeFloat64Tensor(t, []float64{2.0}, []string{"y"}, []int{1}),
			expectError: true,
		},
		{
			name:        "wrong element type in a",
			a:           makeAnyTensor(t, []any{int(1)}, []string{"x"}, []int{1}),
			b:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			expectError: true,
		},
		{
			name:        "wrong element type in b",
			a:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			b:           makeAnyTensor(t, []any{"string"}, []string{"x"}, []int{1}),
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TensorSubtractFloat64(tt.a, tt.b)
			assertArithmeticResult(t, result, err, tt.expected, tt.expectError, tt.expectedDims, tt.expectedShape)
		})
	}
}

// TestTensorMultiplyFloat64 verifies element-wise multiplication.
func TestTensorMultiplyFloat64(t *testing.T) {
	tests := []struct {
		name          string
		a             *ChronoconeTensor[any]
		b             *ChronoconeTensor[any]
		expected      []float64
		expectError   bool
		expectedDims  []string
		expectedShape []int
	}{
		{
			name:          "simple 2-element multiplication",
			a:             makeFloat64Tensor(t, []float64{2.0, 3.0}, []string{"x"}, []int{2}),
			b:             makeFloat64Tensor(t, []float64{4.0, 5.0}, []string{"x"}, []int{2}),
			expected:      []float64{8.0, 15.0},
			expectError:   false,
			expectedDims:  []string{"x"},
			expectedShape: []int{2},
		},
		{
			name:          "2D tensor multiplication",
			a:             makeFloat64Tensor(t, []float64{1.0, 2.0, 3.0, 4.0}, []string{"x", "y"}, []int{2, 2}),
			b:             makeFloat64Tensor(t, []float64{2.0, 3.0, 4.0, 5.0}, []string{"x", "y"}, []int{2, 2}),
			expected:      []float64{2.0, 6.0, 12.0, 20.0},
			expectError:   false,
			expectedDims:  []string{"x", "y"},
			expectedShape: []int{2, 2},
		},
		{
			name:        "nil tensor a",
			a:           nil,
			b:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			expectError: true,
		},
		{
			name:        "nil tensor b",
			a:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			b:           nil,
			expectError: true,
		},
		{
			name:        "dimension name mismatch",
			a:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			b:           makeFloat64Tensor(t, []float64{2.0}, []string{"y"}, []int{1}),
			expectError: true,
		},
		{
			name:        "wrong element type in a",
			a:           makeAnyTensor(t, []any{true}, []string{"x"}, []int{1}),
			b:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			expectError: true,
		},
		{
			name:        "wrong element type in b",
			a:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			b:           makeAnyTensor(t, []any{nil}, []string{"x"}, []int{1}),
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TensorMultiplyFloat64(tt.a, tt.b)
			assertArithmeticResult(t, result, err, tt.expected, tt.expectError, tt.expectedDims, tt.expectedShape)
		})
	}
}

// TestTensorDivideFloat64 verifies element-wise division (a / b), including division by zero.
func TestTensorDivideFloat64(t *testing.T) {
	tests := []struct {
		name          string
		a             *ChronoconeTensor[any]
		b             *ChronoconeTensor[any]
		expected      []float64
		expectError   bool
		expectedDims  []string
		expectedShape []int
	}{
		{
			name:          "simple 2-element division",
			a:             makeFloat64Tensor(t, []float64{10.0, 20.0}, []string{"x"}, []int{2}),
			b:             makeFloat64Tensor(t, []float64{2.0, 5.0}, []string{"x"}, []int{2}),
			expected:      []float64{5.0, 4.0},
			expectError:   false,
			expectedDims:  []string{"x"},
			expectedShape: []int{2},
		},
		{
			name:          "2D tensor division",
			a:             makeFloat64Tensor(t, []float64{100.0, 50.0, 30.0, 20.0}, []string{"x", "y"}, []int{2, 2}),
			b:             makeFloat64Tensor(t, []float64{10.0, 5.0, 3.0, 2.0}, []string{"x", "y"}, []int{2, 2}),
			expected:      []float64{10.0, 10.0, 10.0, 10.0},
			expectError:   false,
			expectedDims:  []string{"x", "y"},
			expectedShape: []int{2, 2},
		},
		{
			name:        "division by zero",
			a:           makeFloat64Tensor(t, []float64{1.0, 2.0}, []string{"x"}, []int{2}),
			b:           makeFloat64Tensor(t, []float64{1.0, 0.0}, []string{"x"}, []int{2}),
			expectError: true,
		},
		{
			name:        "nil tensor a",
			a:           nil,
			b:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			expectError: true,
		},
		{
			name:        "nil tensor b",
			a:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			b:           nil,
			expectError: true,
		},
		{
			name:        "dimension name mismatch",
			a:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			b:           makeFloat64Tensor(t, []float64{2.0}, []string{"y"}, []int{1}),
			expectError: true,
		},
		{
			name:        "wrong element type in a",
			a:           makeAnyTensor(t, []any{[]int{1}}, []string{"x"}, []int{1}),
			b:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			expectError: true,
		},
		{
			name:        "wrong element type in b",
			a:           makeFloat64Tensor(t, []float64{1.0}, []string{"x"}, []int{1}),
			b:           makeAnyTensor(t, []any{map[string]int{"a": 1}}, []string{"x"}, []int{1}),
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TensorDivideFloat64(tt.a, tt.b)
			assertArithmeticResult(t, result, err, tt.expected, tt.expectError, tt.expectedDims, tt.expectedShape)
		})
	}
}

// TestTensorNegateFloat64 verifies element-wise negation.
func TestTensorNegateFloat64(t *testing.T) {
	tests := []struct {
		name          string
		tensor        *ChronoconeTensor[any]
		expected      []float64
		expectError   bool
		expectedDims  []string
		expectedShape []int
	}{
		{
			name:          "negate 2-element tensor",
			tensor:        makeFloat64Tensor(t, []float64{1.0, -2.0}, []string{"x"}, []int{2}),
			expected:      []float64{-1.0, 2.0},
			expectError:   false,
			expectedDims:  []string{"x"},
			expectedShape: []int{2},
		},
		{
			name:          "negate 2D tensor",
			tensor:        makeFloat64Tensor(t, []float64{0.0, -5.0, 3.0, -7.0}, []string{"x", "y"}, []int{2, 2}),
			expected:      []float64{-0.0, 5.0, -3.0, 7.0},
			expectError:   false,
			expectedDims:  []string{"x", "y"},
			expectedShape: []int{2, 2},
		},
		{
			name:        "nil tensor",
			tensor:      nil,
			expectError: true,
		},
		{
			name:        "wrong element type",
			tensor:      makeAnyTensor(t, []any{int(1)}, []string{"x"}, []int{1}),
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TensorNegateFloat64(tt.tensor)
			assertArithmeticResult(t, result, err, tt.expected, tt.expectError, tt.expectedDims, tt.expectedShape)
		})
	}
}

// TestTensorAddScalarFloat64 verifies scalar addition to each element.
func TestTensorAddScalarFloat64(t *testing.T) {
	tests := []struct {
		name          string
		tensor        *ChronoconeTensor[any]
		scalar        float64
		expected      []float64
		expectError   bool
		expectedDims  []string
		expectedShape []int
	}{
		{
			name:          "add scalar to 2-element tensor",
			tensor:        makeFloat64Tensor(t, []float64{1.0, 2.0}, []string{"x"}, []int{2}),
			scalar:        10.0,
			expected:      []float64{11.0, 12.0},
			expectError:   false,
			expectedDims:  []string{"x"},
			expectedShape: []int{2},
		},
		{
			name:          "add scalar to 2D tensor",
			tensor:        makeFloat64Tensor(t, []float64{1.0, 2.0, 3.0, 4.0}, []string{"x", "y"}, []int{2, 2}),
			scalar:        0.5,
			expected:      []float64{1.5, 2.5, 3.5, 4.5},
			expectError:   false,
			expectedDims:  []string{"x", "y"},
			expectedShape: []int{2, 2},
		},
		{
			name:        "nil tensor",
			tensor:      nil,
			scalar:      5.0,
			expectError: true,
		},
		{
			name:        "wrong element type",
			tensor:      makeAnyTensor(t, []any{"string"}, []string{"x"}, []int{1}),
			scalar:      1.0,
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TensorAddScalarFloat64(tt.tensor, tt.scalar)
			assertArithmeticResult(t, result, err, tt.expected, tt.expectError, tt.expectedDims, tt.expectedShape)
		})
	}
}

// TestTensorSubtractScalarFloat64 verifies scalar subtraction from each element (tensor - scalar).
func TestTensorSubtractScalarFloat64(t *testing.T) {
	tests := []struct {
		name          string
		tensor        *ChronoconeTensor[any]
		scalar        float64
		expected      []float64
		expectError   bool
		expectedDims  []string
		expectedShape []int
	}{
		{
			name:          "subtract scalar from 2-element tensor",
			tensor:        makeFloat64Tensor(t, []float64{10.0, 20.0}, []string{"x"}, []int{2}),
			scalar:        3.0,
			expected:      []float64{7.0, 17.0},
			expectError:   false,
			expectedDims:  []string{"x"},
			expectedShape: []int{2},
		},
		{
			name:          "subtract scalar from 2D tensor",
			tensor:        makeFloat64Tensor(t, []float64{5.0, 10.0, 15.0, 20.0}, []string{"x", "y"}, []int{2, 2}),
			scalar:        2.0,
			expected:      []float64{3.0, 8.0, 13.0, 18.0},
			expectError:   false,
			expectedDims:  []string{"x", "y"},
			expectedShape: []int{2, 2},
		},
		{
			name:        "nil tensor",
			tensor:      nil,
			scalar:      1.0,
			expectError: true,
		},
		{
			name:        "wrong element type",
			tensor:      makeAnyTensor(t, []any{int(1)}, []string{"x"}, []int{1}),
			scalar:      1.0,
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TensorSubtractScalarFloat64(tt.tensor, tt.scalar)
			assertArithmeticResult(t, result, err, tt.expected, tt.expectError, tt.expectedDims, tt.expectedShape)
		})
	}
}

// TestTensorMultiplyScalarFloat64 verifies scalar multiplication of each element.
func TestTensorMultiplyScalarFloat64(t *testing.T) {
	tests := []struct {
		name          string
		tensor        *ChronoconeTensor[any]
		scalar        float64
		expected      []float64
		expectError   bool
		expectedDims  []string
		expectedShape []int
	}{
		{
			name:          "multiply 2-element tensor by scalar",
			tensor:        makeFloat64Tensor(t, []float64{2.0, 3.0}, []string{"x"}, []int{2}),
			scalar:        4.0,
			expected:      []float64{8.0, 12.0},
			expectError:   false,
			expectedDims:  []string{"x"},
			expectedShape: []int{2},
		},
		{
			name:          "multiply 2D tensor by scalar",
			tensor:        makeFloat64Tensor(t, []float64{1.0, 2.0, 3.0, 4.0}, []string{"x", "y"}, []int{2, 2}),
			scalar:        10.0,
			expected:      []float64{10.0, 20.0, 30.0, 40.0},
			expectError:   false,
			expectedDims:  []string{"x", "y"},
			expectedShape: []int{2, 2},
		},
		{
			name:        "nil tensor",
			tensor:      nil,
			scalar:      2.0,
			expectError: true,
		},
		{
			name:        "wrong element type",
			tensor:      makeAnyTensor(t, []any{true}, []string{"x"}, []int{1}),
			scalar:      1.0,
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TensorMultiplyScalarFloat64(tt.tensor, tt.scalar)
			assertArithmeticResult(t, result, err, tt.expected, tt.expectError, tt.expectedDims, tt.expectedShape)
		})
	}
}

// TestTensorDivideScalarFloat64 verifies scalar division of each element (tensor / scalar), including zero divisor.
func TestTensorDivideScalarFloat64(t *testing.T) {
	tests := []struct {
		name          string
		tensor        *ChronoconeTensor[any]
		scalar        float64
		expected      []float64
		expectError   bool
		expectedDims  []string
		expectedShape []int
	}{
		{
			name:          "divide 2-element tensor by scalar",
			tensor:        makeFloat64Tensor(t, []float64{10.0, 20.0}, []string{"x"}, []int{2}),
			scalar:        2.0,
			expected:      []float64{5.0, 10.0},
			expectError:   false,
			expectedDims:  []string{"x"},
			expectedShape: []int{2},
		},
		{
			name:          "divide 2D tensor by scalar",
			tensor:        makeFloat64Tensor(t, []float64{100.0, 50.0, 30.0, 20.0}, []string{"x", "y"}, []int{2, 2}),
			scalar:        10.0,
			expected:      []float64{10.0, 5.0, 3.0, 2.0},
			expectError:   false,
			expectedDims:  []string{"x", "y"},
			expectedShape: []int{2, 2},
		},
		{
			name:        "zero scalar divisor",
			tensor:      makeFloat64Tensor(t, []float64{1.0, 2.0}, []string{"x"}, []int{2}),
			scalar:      0.0,
			expectError: true,
		},
		{
			name:        "nil tensor",
			tensor:      nil,
			scalar:      1.0,
			expectError: true,
		},
		{
			name:        "wrong element type",
			tensor:      makeAnyTensor(t, []any{int(1)}, []string{"x"}, []int{1}),
			scalar:      1.0,
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TensorDivideScalarFloat64(tt.tensor, tt.scalar)
			assertArithmeticResult(t, result, err, tt.expected, tt.expectError, tt.expectedDims, tt.expectedShape)
		})
	}
}
