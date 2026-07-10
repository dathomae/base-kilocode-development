// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package calculation

import (
	"math"
	"testing"
	"time"
)

// TestNewChronoconeComputationController verifies that NewChronoconeComputationController initializes correctly.
func TestNewChronoconeComputationController(t *testing.T) {
	controller := NewChronoconeComputationController()

	if controller == nil {
		t.Errorf("NewChronoconeComputationController() returned nil, expected non-nil")
		return
	}

	// Verify that all maps are initialized (empty but not nil)
	if controller.conceptValues == nil {
		t.Errorf("NewChronoconeComputationController().conceptValues is nil, expected initialized map")
	}

	if controller.metricComputations == nil {
		t.Errorf("NewChronoconeComputationController().metricComputations is nil, expected initialized map")
	}

	// Verify that work channel is nil before SetArchitectureConfig is called
	if controller.workChannel != nil {
		t.Errorf("NewChronoconeComputationController().workChannel is non-nil before SetArchitectureConfig")
	}

	// Verify default values for configuration fields
	if controller.numWorkers != 0 {
		t.Errorf("NewChronoconeComputationController().numWorkers = %d, expected 0", controller.numWorkers)
	}

	if controller.cpuCoreCount != 0 {
		t.Errorf("NewChronoconeComputationController().cpuCoreCount = %d, expected 0", controller.cpuCoreCount)
	}
}

// TestNewChronoconeTensor verifies that NewChronoconeTensor creates tensor with correct dimensions.
func TestNewChronoconeTensor(t *testing.T) {
	tests := []struct {
		name           string
		data           []float64
		dimensionNames []string
		shape          []int
	}{
		{
			name:           "1D tensor",
			data:           []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			dimensionNames: []string{"Time"},
			shape:          []int{5},
		},
		{
			name:           "2D tensor",
			data:           make([]float64, 12),
			dimensionNames: []string{"Product", "Region"},
			shape:          []int{3, 4},
		},
		{
			name:           "3D tensor",
			data:           make([]float64, 24),
			dimensionNames: []string{"Time", "Product", "Store"},
			shape:          []int{2, 3, 4},
		},
		{
			name:           "empty tensor",
			data:           []float64{},
			dimensionNames: []string{},
			shape:          []int{},
		},
		{
			name:           "scalar tensor",
			data:           []float64{42.0},
			dimensionNames: []string{},
			shape:          []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tensor := NewChronoconeTensor(tt.data, tt.dimensionNames, tt.shape)

			if tensor == nil {
				t.Errorf("NewChronoconeTensor() returned nil, expected non-nil")
				return
			}

			// Verify dimension names
			dims := tensor.DimensionNames()
			if len(dims) != len(tt.dimensionNames) {
				t.Errorf("tensor.DimensionNames() length = %d, expected %d", len(dims), len(tt.dimensionNames))
			} else {
				for i, name := range tt.dimensionNames {
					if dims[i] != name {
						t.Errorf("tensor.DimensionNames()[%d] = %q, expected %q", i, dims[i], name)
					}
				}
			}

			// Verify shape
			shape := tensor.Shape()
			if len(shape) != len(tt.shape) {
				t.Errorf("tensor.Shape() length = %d, expected %d", len(shape), len(tt.shape))
			} else {
				for i, s := range tt.shape {
					if shape[i] != s {
						t.Errorf("tensor.Shape()[%d] = %d, expected %d", i, shape[i], s)
					}
				}
			}

			// Verify data
			data := tensor.Data()
			if len(data) != len(tt.data) {
				t.Errorf("tensor.Data() length = %d, expected %d", len(data), len(tt.data))
			}
		})
	}
}

// TestMetricComputationGetTensorBlocks verifies that GetTensor blocks until SetTensor is called.
func TestMetricComputationGetTensorBlocks(t *testing.T) {
	mc := NewMetricComputation(1)

	if mc == nil {
		t.Fatalf("NewMetricComputation() returned nil")
	}

	// Create a channel to signal when GetTensor returns
	done := make(chan *ChronoconeTensor[any])

	// Start GetTensor in a goroutine - it should block
	go func() {
		tensor := mc.GetTensor()
		done <- tensor
	}()

	// Give the goroutine time to start and block
	time.Sleep(50 * time.Millisecond)

	// At this point, the goroutine should still be blocked
	select {
	case <-done:
		t.Errorf("GetTensor() returned before SetTensor was called")
	default:
		// Expected - GetTensor is still blocking
	}

	// Now set the tensor
	testTensor := NewChronoconeTensor([]any{1.0, 2.0, 3.0}, []string{"Time"}, []int{3})
	err := mc.SetTensor(testTensor)

	if err != nil {
		t.Errorf("SetTensor() returned error: %v", err)
	}

	// Now GetTensor should return
	select {
	case tensor := <-done:
		if tensor != testTensor {
			t.Errorf("GetTensor() returned wrong tensor: %v, expected %v", tensor, testTensor)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("GetTensor() did not return after SetTensor was called")
	}
}

// TestMetricComputationSetTensorClosesChannel verifies that SetTensor closes the done channel.
func TestMetricComputationSetTensorClosesChannel(t *testing.T) {
	mc := NewMetricComputation(1)

	if mc == nil {
		t.Fatalf("NewMetricComputation() returned nil")
	}

	// Verify IsReady is false before setting tensor
	if mc.IsReady() {
		t.Errorf("IsReady() = true before SetTensor was called, expected false")
	}

	testTensor := NewChronoconeTensor([]any{1.0, 2.0}, []string{"X"}, []int{2})
	err := mc.SetTensor(testTensor)

	if err != nil {
		t.Errorf("SetTensor() returned error: %v", err)
	}

	// Verify IsReady is true after setting tensor
	if !mc.IsReady() {
		t.Errorf("IsReady() = false after SetTensor was called, expected true")
	}
}

// TestMetricComputationSetTensorTwiceError verifies that calling SetTensor twice returns an error.
func TestMetricComputationSetTensorTwiceError(t *testing.T) {
	mc := NewMetricComputation(1)

	if mc == nil {
		t.Fatalf("NewMetricComputation() returned nil")
	}

	testTensor := NewChronoconeTensor([]any{1.0}, []string{"X"}, []int{1})

	// First call should succeed
	err1 := mc.SetTensor(testTensor)
	if err1 != nil {
		t.Errorf("First SetTensor() returned error: %v", err1)
	}

	// Second call should fail
	err2 := mc.SetTensor(testTensor)
	if err2 == nil {
		t.Errorf("Second SetTensor() returned nil error, expected error")
	}
}

// TestTensorAllocationWithVariousDimensions verifies tensor allocation with various dimension sizes.
func TestTensorAllocationWithVariousDimensions(t *testing.T) {
	tests := []struct {
		name           string
		dimensionNames []string
		shape          []int
		expectedSize   int
	}{
		{
			name:           "small 1D tensor",
			dimensionNames: []string{"Time"},
			shape:          []int{10},
			expectedSize:   10,
		},
		{
			name:           "medium 2D tensor",
			dimensionNames: []string{"Product", "Region"},
			shape:          []int{100, 50},
			expectedSize:   5000,
		},
		{
			name:           "large 3D tensor",
			dimensionNames: []string{"Time", "Product", "Store"},
			shape:          []int{365, 1000, 100},
			expectedSize:   36500000,
		},
		{
			name:           "tensor with single element dimensions",
			dimensionNames: []string{"A", "B", "C"},
			shape:          []int{1, 1, 1},
			expectedSize:   1,
		},
		{
			name:           "empty tensor (scalar)",
			dimensionNames: []string{},
			shape:          []int{},
			expectedSize:   1, // Scalar has 1 element
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create data array of the expected size
			data := make([]float64, tt.expectedSize)
			for i := range data {
				data[i] = float64(i)
			}

			tensor := NewChronoconeTensor(data, tt.dimensionNames, tt.shape)

			if tensor == nil {
				t.Errorf("NewChronoconeTensor() returned nil")
				return
			}

			// Verify shape matches
			shape := tensor.Shape()
			if len(shape) != len(tt.shape) {
				t.Errorf("tensor.Shape() length = %d, expected %d", len(shape), len(tt.shape))
				return
			}

			// Calculate actual size from shape
			actualSize := 1
			for _, s := range shape {
				actualSize *= s
			}
			if len(shape) == 0 {
				actualSize = 1 // Scalar case
			}

			if actualSize != tt.expectedSize {
				t.Errorf("tensor total size = %d, expected %d", actualSize, tt.expectedSize)
			}

			// Verify data length
			tensorData := tensor.Data()
			if len(tensorData) != tt.expectedSize {
				t.Errorf("tensor.Data() length = %d, expected %d", len(tensorData), tt.expectedSize)
			}
		})
	}
}

// TestAllocateTensorMask verifies that AllocateTensor allocates a mask
// alongside the data array.
func TestAllocateTensorMask(t *testing.T) {
	controller := NewChronoconeComputationController()

	if controller == nil {
		t.Fatalf("NewChronoconeComputationController() returned nil")
	}

	controller.conceptValues = map[string][]interface{}{
		"Product": {"A", "B", "C"},
	}

	tensor, err := AllocateTensor(controller, "Product")
	if err != nil {
		t.Fatalf("AllocateTensor() returned error: %v", err)
	}
	if tensor == nil {
		t.Fatal("AllocateTensor() returned nil tensor")
	}

	mask := tensor.Mask()
	expectedMaskLen := (3 + 63) / 64 // 1 uint64 word
	if len(mask) != expectedMaskLen {
		t.Errorf("AllocateTensor mask length = %d, want %d", len(mask), expectedMaskLen)
	}
	// Verify mask is zero-initialized (all positions true/selected by default).
	for i, word := range mask {
		if word != 0 {
			t.Errorf("mask[%d] = 0x%016X, want 0 (zero-initialized)", i, word)
		}
	}
}

// TestAllocateTensor verifies the AllocateTensor function.
func TestAllocateTensor(t *testing.T) {
	controller := NewChronoconeComputationController()

	if controller == nil {
		t.Fatalf("NewChronoconeComputationController() returned nil")
	}

	// Set up some concept values in the controller
	controller.conceptValues = map[string][]interface{}{
		"Product": {"A", "B", "C"},
		"Region":  {"North", "South", "East", "West"},
		"Time":    {"2024-01", "2024-02"},
	}

	tests := []struct {
		name           string
		dimensionNames []string
		expectError    bool
	}{
		{
			name:           "valid single dimension",
			dimensionNames: []string{"Product"},
			expectError:    false,
		},
		{
			name:           "valid multiple dimensions",
			dimensionNames: []string{"Product", "Region"},
			expectError:    false,
		},
		{
			name:           "unknown dimension",
			dimensionNames: []string{"UnknownConcept"},
			expectError:    true,
		},
		{
			name:           "mixed valid and invalid dimensions",
			dimensionNames: []string{"Product", "Unknown"},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tensor, err := AllocateTensor(controller, tt.dimensionNames...)

			if tt.expectError {
				if err == nil {
					t.Errorf("AllocateTensor() returned nil error, expected error")
				}
				if tensor != nil {
					t.Errorf("AllocateTensor() returned non-nil tensor with error, expected nil")
				}
			} else {
				if err != nil {
					t.Errorf("AllocateTensor() returned error: %v", err)
				}
				if tensor == nil {
					t.Errorf("AllocateTensor() returned nil tensor, expected non-nil")
					return
				}

				// Verify dimensions match
				dims := tensor.DimensionNames()
				if len(dims) != len(tt.dimensionNames) {
					t.Errorf("tensor.DimensionNames() length = %d, expected %d", len(dims), len(tt.dimensionNames))
				}
			}
		})
	}
}

// TestNewMetricComputation verifies NewMetricComputation creates correct initial state.
func TestNewMetricComputation(t *testing.T) {
	tests := []struct {
		name          string
		consumerCount int32
	}{
		{
			name:          "single consumer",
			consumerCount: 1,
		},
		{
			name:          "multiple consumers",
			consumerCount: 5,
		},
		{
			name:          "zero consumers",
			consumerCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricComputation(tt.consumerCount)

			if mc == nil {
				t.Errorf("NewMetricComputation() returned nil")
				return
			}

			// Verify initial state
			if mc.tensor.Load() != nil {
				t.Errorf("NewMetricComputation().tensor is non-nil, expected nil")
			}

			if mc.IsReady() {
				t.Errorf("NewMetricComputation().IsReady() = true, expected false")
			}

			if mc.consumerCount != tt.consumerCount {
				t.Errorf("NewMetricComputation().consumerCount = %d, expected %d", mc.consumerCount, tt.consumerCount)
			}
		})
	}
}

// TestMetricComputationReleaseTensor verifies the ReleaseTensor functionality.
func TestMetricComputationReleaseTensor(t *testing.T) {
	mc := NewMetricComputation(2)

	if mc == nil {
		t.Fatalf("NewMetricComputation() returned nil")
	}

	testTensor := NewChronoconeTensor([]any{1.0, 2.0}, []string{"X"}, []int{2})
	err := mc.SetTensor(testTensor)
	if err != nil {
		t.Fatalf("SetTensor() failed: %v", err)
	}

	// First release - count goes from 2 to 1
	mc.ReleaseTensor()
	if mc.consumerCount != 1 {
		t.Errorf("After first ReleaseTensor(), consumerCount = %d, expected 1", mc.consumerCount)
	}

	// Second release - count goes from 1 to 0, tensor should be nil
	mc.ReleaseTensor()
	if mc.consumerCount != 0 {
		t.Errorf("After second ReleaseTensor(), consumerCount = %d, expected 0", mc.consumerCount)
	}

	// Tensor should still be accessible via GetTensor (it's already set)
	// but the internal reference might be cleared for GC
}

// maskEqual compares two uint64 slices for equality.
func maskEqual(a, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// makeMaskedAnyTensor creates a *ChronoconeTensor[any] with float64 data and a custom mask.
// Used by mask propagation tests to set up operands with known mask patterns.
func makeMaskedAnyTensor(data []float64, mask []uint64) *ChronoconeTensor[any] {
	anyData := make([]any, len(data))
	for i, v := range data {
		anyData[i] = v
	}
	tensor := NewChronoconeTensor(anyData, []string{"X"}, []int{len(data)})
	tensor.mask = mask
	return tensor
}

// TestMaskAllocationSize verifies that NewChronoconeTensor allocates a mask
// with correct size (len(data)+63)/64 for various data sizes.
func TestMaskAllocationSize(t *testing.T) {
	tests := []struct {
		name     string
		dataSize int
		wantLen  int
	}{
		{"zero elements", 0, 0},
		{"one element", 1, 1},
		{"63 elements", 63, 1},
		{"64 elements", 64, 1},
		{"65 elements", 65, 2},
		{"100 elements", 100, 2},
		{"127 elements", 127, 2},
		{"128 elements", 128, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]float64, tt.dataSize)
			tensor := NewChronoconeTensor(data, []string{"X"}, []int{tt.dataSize})
			mask := tensor.Mask()
			if len(mask) != tt.wantLen {
				t.Errorf("Mask() length = %d, want %d", len(mask), tt.wantLen)
			}
		})
	}
}

// TestMaskAllocationAllBitsZero verifies that a freshly allocated tensor
// has all mask bits set to zero, following the all-true convention.
func TestMaskAllocationAllBitsZero(t *testing.T) {
	data := make([]float64, 100)
	tensor := NewChronoconeTensor(data, []string{"X"}, []int{100})
	mask := tensor.Mask()
	for i, word := range mask {
		if word != 0 {
			t.Errorf("mask[%d] = 0x%016X, want 0", i, word)
		}
	}
}

// TestMaskAccessor verifies that Mask() returns the underlying mask slice
// with correct length and that the returned slice references the tensor's mask.
func TestMaskAccessor(t *testing.T) {
	data := []float64{1.0, 2.0, 3.0}
	tensor := NewChronoconeTensor(data, []string{"X"}, []int{3})
	mask := tensor.Mask()
	if mask == nil {
		t.Fatal("Mask() returned nil, expected non-nil slice")
	}
	if len(mask) != 1 {
		t.Errorf("Mask() length = %d, want 1", len(mask))
	}
	// Verify the returned slice is a reference to the underlying mask.
	mask[0] = 0xDEADBEEFCAFEBABE
	if tensor.mask[0] != 0xDEADBEEFCAFEBABE {
		t.Error("Mask() returned a copy, expected reference to underlying slice")
	}
}

// TestMaskLen verifies that MaskLen() returns the correct mask length
// in uint64 words: (len(data) + 63) / 64.
func TestMaskLen(t *testing.T) {
	tests := []struct {
		name     string
		dataSize int
		wantLen  int
	}{
		{"zero elements", 0, 0},
		{"one element", 1, 1},
		{"63 elements", 63, 1},
		{"64 elements", 64, 1},
		{"65 elements", 65, 2},
		{"100 elements", 100, 2},
		{"127 elements", 127, 2},
		{"128 elements", 128, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]float64, tt.dataSize)
			tensor := NewChronoconeTensor(data, []string{"X"}, []int{tt.dataSize})
			got := tensor.MaskLen()
			if got != tt.wantLen {
				t.Errorf("MaskLen() = %d, want %d", got, tt.wantLen)
			}
		})
	}
}

// TestMaskBitGetSetWordBoundaries verifies SetMaskBit and IsMaskedTrue
// at various indices including word boundaries: 0, 63, 64, 127, 128.
func TestMaskBitGetSetWordBoundaries(t *testing.T) {
	data := make([]float64, 129)
	tensor := NewChronoconeTensor(data, []string{"X"}, []int{129})
	indices := []int{0, 63, 64, 127, 128}
	for _, idx := range indices {
		// Fresh tensor: bit is 0, so IsMaskedTrue should return true.
		if !tensor.IsMaskedTrue(idx) {
			t.Errorf("IsMaskedTrue(%d) on fresh tensor = false, want true", idx)
		}
		// Set bit to 1 (false/unselected): value=false sets bit to 1.
		tensor.SetMaskBit(idx, false)
		if tensor.IsMaskedTrue(idx) {
			t.Errorf("IsMaskedTrue(%d) after unmasking = true, want false", idx)
		}
		// Set bit back to 0 (true/selected): value=true sets bit to 0.
		tensor.SetMaskBit(idx, true)
		if !tensor.IsMaskedTrue(idx) {
			t.Errorf("IsMaskedTrue(%d) after masking = false, want true", idx)
		}
	}
}

// TestMaskZeroMeansTrue verifies the 0=true convention:
// IsMaskedTrue returns true when the corresponding bit is 0,
// and false when the bit is 1.
func TestMaskZeroMeansTrue(t *testing.T) {
	data := make([]float64, 64)
	tensor := NewChronoconeTensor(data, []string{"X"}, []int{64})

	// All bits 0: IsMaskedTrue should return true for every index.
	tensor.mask = []uint64{0}
	for i := 0; i < 64; i++ {
		if !tensor.IsMaskedTrue(i) {
			t.Errorf("IsMaskedTrue(%d) with bit=0 = false, want true", i)
		}
	}

	// All bits 1: IsMaskedTrue should return false for every index.
	tensor.mask = []uint64{^uint64(0)}
	for i := 0; i < 64; i++ {
		if tensor.IsMaskedTrue(i) {
			t.Errorf("IsMaskedTrue(%d) with bit=1 = true, want false", i)
		}
	}
}

// TestMaskBitRoundTrip verifies that setting a mask bit and reading it
// back produces the expected value.
func TestMaskBitRoundTrip(t *testing.T) {
	data := make([]float64, 10)
	tensor := NewChronoconeTensor(data, []string{"X"}, []int{10})
	for i := 0; i < 10; i++ {
		// Set to false (bit=1, unmasked) -> IsMaskedTrue returns false.
		tensor.SetMaskBit(i, false)
		if tensor.IsMaskedTrue(i) {
			t.Errorf("round-trip %d: IsMaskedTrue after SetMaskBit(false) = true, want false", i)
		}
		// Set to true (bit=0, masked) -> IsMaskedTrue returns true.
		tensor.SetMaskBit(i, true)
		if !tensor.IsMaskedTrue(i) {
			t.Errorf("round-trip %d: IsMaskedTrue after SetMaskBit(true) = false, want true", i)
		}
	}
}

// TestMaskBitToggle verifies setting a bit from false->true->false.
func TestMaskBitToggle(t *testing.T) {
	data := make([]float64, 5)
	tensor := NewChronoconeTensor(data, []string{"X"}, []int{5})

	// Start unmasked (bit=1): SetMaskBit(false) sets bit to 1.
	tensor.SetMaskBit(0, false)
	if tensor.IsMaskedTrue(0) {
		t.Error("after SetMaskBit(false): IsMaskedTrue = true, want false")
	}

	// Toggle to masked (bit=0): SetMaskBit(true) sets bit to 0.
	tensor.SetMaskBit(0, true)
	if !tensor.IsMaskedTrue(0) {
		t.Error("after SetMaskBit(true): IsMaskedTrue = false, want true")
	}

	// Toggle back to unmasked (bit=1).
	tensor.SetMaskBit(0, false)
	if tensor.IsMaskedTrue(0) {
		t.Error("after second SetMaskBit(false): IsMaskedTrue = true, want false")
	}
}

// TestMaskPropagationBinaryOps verifies that binary arithmetic operations
// copy the left operand's mask to the output tensor.
func TestMaskPropagationBinaryOps(t *testing.T) {
	a := makeMaskedAnyTensor([]float64{1.0, 2.0, 3.0, 4.0}, []uint64{0x0A})
	b := makeMaskedAnyTensor([]float64{5.0, 6.0, 7.0, 8.0}, []uint64{0x05})

	tests := []struct {
		name string
		fn   func() (*ChronoconeTensor[any], error)
	}{
		{"TensorAddFloat64", func() (*ChronoconeTensor[any], error) { return TensorAddFloat64(a, b) }},
		{"TensorSubtractFloat64", func() (*ChronoconeTensor[any], error) { return TensorSubtractFloat64(a, b) }},
		{"TensorMultiplyFloat64", func() (*ChronoconeTensor[any], error) { return TensorMultiplyFloat64(a, b) }},
		{"TensorDivideFloat64", func() (*ChronoconeTensor[any], error) { return TensorDivideFloat64(a, b) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn()
			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}
			if !maskEqual(result.Mask(), a.Mask()) {
				t.Errorf("%s result mask = %v, want %v (left operand mask)", tt.name, result.Mask(), a.Mask())
			}
		})
	}
}

// TestMaskPropagationUnaryOps verifies that unary arithmetic operations
// copy the input tensor's mask to the output tensor.
func TestMaskPropagationUnaryOps(t *testing.T) {
	input := makeMaskedAnyTensor([]float64{1.0, 2.0, 3.0}, []uint64{0x03})
	result, err := TensorNegateFloat64(input)
	if err != nil {
		t.Fatalf("TensorNegateFloat64 failed: %v", err)
	}
	if !maskEqual(result.Mask(), input.Mask()) {
		t.Errorf("TensorNegateFloat64 result mask = %v, want %v (input mask)", result.Mask(), input.Mask())
	}
}

// TestMaskPropagationScalarOps verifies that scalar arithmetic operations
// copy the input tensor's mask to the output tensor.
func TestMaskPropagationScalarOps(t *testing.T) {
	input := makeMaskedAnyTensor([]float64{1.0, 2.0, 3.0, 4.0}, []uint64{0x05})

	tests := []struct {
		name string
		fn   func() (*ChronoconeTensor[any], error)
	}{
		{"TensorAddScalarFloat64", func() (*ChronoconeTensor[any], error) { return TensorAddScalarFloat64(input, 10.0) }},
		{"TensorSubtractScalarFloat64", func() (*ChronoconeTensor[any], error) { return TensorSubtractScalarFloat64(input, 10.0) }},
		{"TensorMultiplyScalarFloat64", func() (*ChronoconeTensor[any], error) { return TensorMultiplyScalarFloat64(input, 10.0) }},
		{"TensorDivideScalarFloat64", func() (*ChronoconeTensor[any], error) { return TensorDivideScalarFloat64(input, 2.0) }},
		{"ScalarSubtractTensorFloat64", func() (*ChronoconeTensor[any], error) { return ScalarSubtractTensorFloat64(10.0, input) }},
		{"ScalarDivideTensorFloat64", func() (*ChronoconeTensor[any], error) { return ScalarDivideTensorFloat64(10.0, input) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn()
			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}
			if !maskEqual(result.Mask(), input.Mask()) {
				t.Errorf("%s result mask = %v, want %v (input mask)", tt.name, result.Mask(), input.Mask())
			}
		})
	}
}

// TestMaskNilTensorEdgeCases verifies edge case behavior for nil tensors:
// Mask() returns nil, MaskLen() returns 0, IsMaskedTrue returns false,
// and SetMaskBit is a no-op.
func TestMaskNilTensorEdgeCases(t *testing.T) {
	var tensor *ChronoconeTensor[float64]
	if tensor.Mask() != nil {
		t.Error("nil tensor Mask() returned non-nil, expected nil")
	}
	if tensor.MaskLen() != 0 {
		t.Errorf("nil tensor MaskLen() = %d, expected 0", tensor.MaskLen())
	}
	if tensor.IsMaskedTrue(0) {
		t.Error("nil tensor IsMaskedTrue(0) = true, expected false")
	}
	// SetMaskBit on nil tensor should be a no-op (no panic).
	tensor.SetMaskBit(0, true)
}

// TestMaskSetMaskNil verifies that SetMask accepts a nil mask without panic.
func TestMaskSetMaskNil(t *testing.T) {
	data := []float64{1.0, 2.0}
	tensor := NewChronoconeTensor(data, []string{"X"}, []int{2})
	// Should not panic.
	tensor.SetMask(nil)
	// Verify that IsMaskedTrue and SetMaskBit are safe after SetMask(nil).
	if tensor.IsMaskedTrue(0) {
		t.Error("IsMaskedTrue(0) after SetMask(nil) = true, expected false")
	}
	tensor.SetMaskBit(0, true) // should not panic
}

// TestMaskSetMaskNonNil verifies that SetMask stores a non-nil mask
// and that Mask() reflects the updated value.
func TestMaskSetMaskNonNil(t *testing.T) {
	data := []float64{1.0, 2.0, 3.0, 4.0}
	tensor := NewChronoconeTensor(data, []string{"X"}, []int{4})
	newMask := []uint64{0xDEADBEEFCAFEBABE}
	tensor.SetMask(newMask)
	got := tensor.Mask()
	if !maskEqual(got, newMask) {
		t.Errorf("Mask() = %v, want %v after SetMask", got, newMask)
	}
}

// TestMaskEmptyTensor verifies mask allocation for a tensor with zero elements.
func TestMaskEmptyTensor(t *testing.T) {
	tensor := NewChronoconeTensor([]float64{}, []string{}, []int{})
	mask := tensor.Mask()
	if len(mask) != 0 {
		t.Errorf("empty tensor Mask() length = %d, expected 0", len(mask))
	}
	if tensor.MaskLen() != 0 {
		t.Errorf("empty tensor MaskLen() = %d, expected 0", tensor.MaskLen())
	}
}

// TestMaskOutOfBounds verifies graceful handling of out-of-bounds indices:
// IsMaskedTrue returns false, SetMaskBit is a no-op.
func TestMaskOutOfBounds(t *testing.T) {
	data := make([]float64, 10)
	tensor := NewChronoconeTensor(data, []string{"X"}, []int{10})

	// Negative index.
	if tensor.IsMaskedTrue(-1) {
		t.Error("IsMaskedTrue(-1) = true, expected false")
	}

	// Index equal to data length.
	if tensor.IsMaskedTrue(10) {
		t.Error("IsMaskedTrue(10) = true, expected false")
	}

	// Index well beyond data length.
	if tensor.IsMaskedTrue(1000) {
		t.Error("IsMaskedTrue(1000) = true, expected false")
	}
}

// TestComparisonTensorTensor verifies all 6 tensor-tensor comparison operators
// produce correct mask tensors with left-operand data and matching dimension names.
func TestComparisonTensorTensor(t *testing.T) {
	a := makeMaskedAnyTensor([]float64{1.0, 3.0, 2.0, 5.0}, []uint64{0})
	b := makeMaskedAnyTensor([]float64{2.0, 3.0, 1.0, 4.0}, []uint64{0})

	tests := []struct {
		name     string
		fn       func() (*ChronoconeTensor[any], error)
		expected []uint64
	}{
		{"GT", func() (*ChronoconeTensor[any], error) { return TensorGreaterThanFloat64Mask(a, b) }, []uint64{0x03}},
		{"GE", func() (*ChronoconeTensor[any], error) { return TensorGreaterEqualFloat64Mask(a, b) }, []uint64{0x01}},
		{"LT", func() (*ChronoconeTensor[any], error) { return TensorLessThanFloat64Mask(a, b) }, []uint64{0x0E}},
		{"LE", func() (*ChronoconeTensor[any], error) { return TensorLessEqualFloat64Mask(a, b) }, []uint64{0x0C}},
		{"EQ", func() (*ChronoconeTensor[any], error) { return TensorEqualFloat64Mask(a, b) }, []uint64{0x0D}},
		{"NE", func() (*ChronoconeTensor[any], error) { return TensorNotEqualFloat64Mask(a, b) }, []uint64{0x02}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn()
			if err != nil {
				t.Fatalf("%s returned error: %v", tt.name, err)
			}
			if result == nil {
				t.Fatalf("%s returned nil result", tt.name)
			}
			if !maskEqual(result.Mask(), tt.expected) {
				t.Errorf("%s mask = 0x%02X, want 0x%02X", tt.name, result.Mask()[0], tt.expected[0])
			}
			if len(result.Data()) != len(a.Data()) {
				t.Errorf("%s data length = %d, want %d", tt.name, len(result.Data()), len(a.Data()))
			} else {
				for i := range a.Data() {
					if result.Data()[i] != a.Data()[i] {
						t.Errorf("%s data[%d] = %v, want %v", tt.name, i, result.Data()[i], a.Data()[i])
					}
				}
			}
			dimsA := a.DimensionNames()
			dimsR := result.DimensionNames()
			if len(dimsR) != len(dimsA) {
				t.Errorf("%s dimension names length = %d, want %d", tt.name, len(dimsR), len(dimsA))
			} else {
				for i := range dimsA {
					if dimsR[i] != dimsA[i] {
						t.Errorf("%s dimensionNames[%d] = %q, want %q", tt.name, i, dimsR[i], dimsA[i])
					}
				}
			}
		})
	}
}

// TestComparisonScalarTensor verifies all 6 scalar-tensor comparison operators
// produce correct mask tensors with input data and matching dimension names.
func TestComparisonScalarTensor(t *testing.T) {
	a := makeMaskedAnyTensor([]float64{1.0, 3.0, 2.0, 5.0}, []uint64{0})
	scalar := 3.0

	tests := []struct {
		name     string
		fn       func() (*ChronoconeTensor[any], error)
		expected []uint64
	}{
		{"GT", func() (*ChronoconeTensor[any], error) { return TensorGreaterThanScalarFloat64Mask(a, scalar) }, []uint64{0x07}},
		{"GE", func() (*ChronoconeTensor[any], error) { return TensorGreaterEqualScalarFloat64Mask(a, scalar) }, []uint64{0x05}},
		{"LT", func() (*ChronoconeTensor[any], error) { return TensorLessThanScalarFloat64Mask(a, scalar) }, []uint64{0x0A}},
		{"LE", func() (*ChronoconeTensor[any], error) { return TensorLessEqualScalarFloat64Mask(a, scalar) }, []uint64{0x08}},
		{"EQ", func() (*ChronoconeTensor[any], error) { return TensorEqualScalarFloat64Mask(a, scalar) }, []uint64{0x0D}},
		{"NE", func() (*ChronoconeTensor[any], error) { return TensorNotEqualScalarFloat64Mask(a, scalar) }, []uint64{0x02}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn()
			if err != nil {
				t.Fatalf("%s returned error: %v", tt.name, err)
			}
			if result == nil {
				t.Fatalf("%s returned nil result", tt.name)
			}
			if !maskEqual(result.Mask(), tt.expected) {
				t.Errorf("%s mask = 0x%02X, want 0x%02X", tt.name, result.Mask()[0], tt.expected[0])
			}
			if len(result.Data()) != len(a.Data()) {
				t.Errorf("%s data length = %d, want %d", tt.name, len(result.Data()), len(a.Data()))
			} else {
				for i := range a.Data() {
					if result.Data()[i] != a.Data()[i] {
						t.Errorf("%s data[%d] = %v, want %v", tt.name, i, result.Data()[i], a.Data()[i])
					}
				}
			}
			dimsA := a.DimensionNames()
			dimsR := result.DimensionNames()
			if len(dimsR) != len(dimsA) {
				t.Errorf("%s dimension names length = %d, want %d", tt.name, len(dimsR), len(dimsA))
			} else {
				for i := range dimsA {
					if dimsR[i] != dimsA[i] {
						t.Errorf("%s dimensionNames[%d] = %q, want %q", tt.name, i, dimsR[i], dimsA[i])
					}
				}
			}
		})
	}
}

// TestComparisonDimensionMismatch verifies that tensor-tensor comparison
// functions return an error when dimension names do not match.
func TestComparisonDimensionMismatch(t *testing.T) {
	a := makeMaskedAnyTensor([]float64{1.0, 2.0}, []uint64{0})
	b := NewChronoconeTensor([]any{1.0, 2.0}, []string{"Y"}, []int{2})

	tests := []struct {
		name string
		fn   func() (*ChronoconeTensor[any], error)
	}{
		{"GT", func() (*ChronoconeTensor[any], error) { return TensorGreaterThanFloat64Mask(a, b) }},
		{"GE", func() (*ChronoconeTensor[any], error) { return TensorGreaterEqualFloat64Mask(a, b) }},
		{"LT", func() (*ChronoconeTensor[any], error) { return TensorLessThanFloat64Mask(a, b) }},
		{"LE", func() (*ChronoconeTensor[any], error) { return TensorLessEqualFloat64Mask(a, b) }},
		{"EQ", func() (*ChronoconeTensor[any], error) { return TensorEqualFloat64Mask(a, b) }},
		{"NE", func() (*ChronoconeTensor[any], error) { return TensorNotEqualFloat64Mask(a, b) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn()
			if err == nil {
				t.Errorf("%s returned nil error, expected non-nil error for dimension mismatch", tt.name)
			}
			if result != nil {
				t.Errorf("%s returned non-nil result with error, expected nil", tt.name)
			}
		})
	}
}

// TestComparisonNilInput verifies that all comparison functions return a
// non-nil error when passed nil tensor operands.
func TestComparisonNilInput(t *testing.T) {
	valid := makeMaskedAnyTensor([]float64{1.0, 2.0}, []uint64{0})

	// Tensor-tensor functions: nil a, nil b, both nil.
	ttOps := []struct {
		name string
		fn   func() (*ChronoconeTensor[any], error)
	}{
		{"GT_nilA", func() (*ChronoconeTensor[any], error) { return TensorGreaterThanFloat64Mask(nil, valid) }},
		{"GT_nilB", func() (*ChronoconeTensor[any], error) { return TensorGreaterThanFloat64Mask(valid, nil) }},
		{"GT_nilBoth", func() (*ChronoconeTensor[any], error) { return TensorGreaterThanFloat64Mask(nil, nil) }},
		{"GE_nilA", func() (*ChronoconeTensor[any], error) { return TensorGreaterEqualFloat64Mask(nil, valid) }},
		{"GE_nilB", func() (*ChronoconeTensor[any], error) { return TensorGreaterEqualFloat64Mask(valid, nil) }},
		{"GE_nilBoth", func() (*ChronoconeTensor[any], error) { return TensorGreaterEqualFloat64Mask(nil, nil) }},
		{"LT_nilA", func() (*ChronoconeTensor[any], error) { return TensorLessThanFloat64Mask(nil, valid) }},
		{"LT_nilB", func() (*ChronoconeTensor[any], error) { return TensorLessThanFloat64Mask(valid, nil) }},
		{"LT_nilBoth", func() (*ChronoconeTensor[any], error) { return TensorLessThanFloat64Mask(nil, nil) }},
		{"LE_nilA", func() (*ChronoconeTensor[any], error) { return TensorLessEqualFloat64Mask(nil, valid) }},
		{"LE_nilB", func() (*ChronoconeTensor[any], error) { return TensorLessEqualFloat64Mask(valid, nil) }},
		{"LE_nilBoth", func() (*ChronoconeTensor[any], error) { return TensorLessEqualFloat64Mask(nil, nil) }},
		{"EQ_nilA", func() (*ChronoconeTensor[any], error) { return TensorEqualFloat64Mask(nil, valid) }},
		{"EQ_nilB", func() (*ChronoconeTensor[any], error) { return TensorEqualFloat64Mask(valid, nil) }},
		{"EQ_nilBoth", func() (*ChronoconeTensor[any], error) { return TensorEqualFloat64Mask(nil, nil) }},
		{"NE_nilA", func() (*ChronoconeTensor[any], error) { return TensorNotEqualFloat64Mask(nil, valid) }},
		{"NE_nilB", func() (*ChronoconeTensor[any], error) { return TensorNotEqualFloat64Mask(valid, nil) }},
		{"NE_nilBoth", func() (*ChronoconeTensor[any], error) { return TensorNotEqualFloat64Mask(nil, nil) }},
	}

	for _, tt := range ttOps {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn()
			if err == nil {
				t.Errorf("%s returned nil error, expected non-nil error for nil input", tt.name)
			}
			if result != nil {
				t.Errorf("%s returned non-nil result with error, expected nil", tt.name)
			}
		})
	}

	// Scalar-tensor functions: nil tensor.
	stOps := []struct {
		name string
		fn   func() (*ChronoconeTensor[any], error)
	}{
		{"GT_scalar", func() (*ChronoconeTensor[any], error) { return TensorGreaterThanScalarFloat64Mask(nil, 1.0) }},
		{"GE_scalar", func() (*ChronoconeTensor[any], error) { return TensorGreaterEqualScalarFloat64Mask(nil, 1.0) }},
		{"LT_scalar", func() (*ChronoconeTensor[any], error) { return TensorLessThanScalarFloat64Mask(nil, 1.0) }},
		{"LE_scalar", func() (*ChronoconeTensor[any], error) { return TensorLessEqualScalarFloat64Mask(nil, 1.0) }},
		{"EQ_scalar", func() (*ChronoconeTensor[any], error) { return TensorEqualScalarFloat64Mask(nil, 1.0) }},
		{"NE_scalar", func() (*ChronoconeTensor[any], error) { return TensorNotEqualScalarFloat64Mask(nil, 1.0) }},
	}

	for _, tt := range stOps {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn()
			if err == nil {
				t.Errorf("%s returned nil error, expected non-nil error for nil input", tt.name)
			}
			if result != nil {
				t.Errorf("%s returned non-nil result with error, expected nil", tt.name)
			}
		})
	}
}

// TestComparisonNaNHandling verifies that comparisons involving NaN always
// produce false (mask bit = 1) because NaN is not ordered.
func TestComparisonNaNHandling(t *testing.T) {
	na := makeMaskedAnyTensor([]float64{math.NaN(), 3.0, math.NaN(), 5.0}, []uint64{0})
	nb := makeMaskedAnyTensor([]float64{2.0, math.NaN(), 1.0, math.NaN()}, []uint64{0})

	// For GT: NaN > anything = F, 3.0 > NaN = F, NaN > 1.0 = F, 5.0 > NaN = F
	// All false → all bits 1 → mask word = 0x0F
	result, err := TensorGreaterThanFloat64Mask(na, nb)
	if err != nil {
		t.Fatalf("GT NaN test returned error: %v", err)
	}
	if result == nil {
		t.Fatalf("GT NaN test returned nil result")
	}
	expected := []uint64{0x0F}
	if !maskEqual(result.Mask(), expected) {
		t.Errorf("GT NaN mask = 0x%02X, want 0x%02X (all false)", result.Mask()[0], expected[0])
	}

	// Scalar NaN test: all elements compared against NaN should be false.
	scalarNaN := math.NaN()
	result2, err2 := TensorEqualScalarFloat64Mask(na, scalarNaN)
	if err2 != nil {
		t.Fatalf("EQ scalar NaN test returned error: %v", err2)
	}
	if result2 == nil {
		t.Fatalf("EQ scalar NaN test returned nil result")
	}
	if !maskEqual(result2.Mask(), expected) {
		t.Errorf("EQ scalar NaN mask = 0x%02X, want 0x%02X (all false)", result2.Mask()[0], expected[0])
	}
}

// TestTensorBroadcastTo_MaterialToProductMaterialDaily verifies broadcasting
// a [Material:3] tensor to [Product:3, Material:3, Daily:7].
func TestTensorBroadcastTo_MaterialToProductMaterialDaily(t *testing.T) {
	srcData := []any{1.0, 0.8, 0.9} // Steel, Aluminum, Copper
	src := NewChronoconeTensor(srcData, []string{"Material"}, []int{3})

	result, err := TensorBroadcastTo(src, []string{"Product", "Material", "Daily"}, []int{3, 3, 7})
	if err != nil {
		t.Fatalf("TensorBroadcastTo returned error: %v", err)
	}
	if result == nil {
		t.Fatal("TensorBroadcastTo returned nil")
	}

	// Verify dimensions and shape
	dims := result.DimensionNames()
	if len(dims) != 3 || dims[0] != "Product" || dims[1] != "Material" || dims[2] != "Daily" {
		t.Errorf("result.DimensionNames() = %v, want [Product Material Daily]", dims)
	}
	shape := result.Shape()
	if len(shape) != 3 || shape[0] != 3 || shape[1] != 3 || shape[2] != 7 {
		t.Errorf("result.Shape() = %v, want [3 3 7]", shape)
	}

	// Verify total size
	data := result.Data()
	if len(data) != 63 {
		t.Errorf("result.Data() length = %d, want 63", len(data))
	}

	// Verify that for each Product and Daily, Material values are repeated correctly.
	for product := 0; product < 3; product++ {
		for material := 0; material < 3; material++ {
			for day := 0; day < 7; day++ {
				flatIdx := product*21 + material*7 + day
				expected := srcData[material]
				if data[flatIdx] != expected {
					t.Errorf("data[%d] (Product=%d, Material=%d, Day=%d) = %v, want %v",
						flatIdx, product, material, day, data[flatIdx], expected)
				}
			}
		}
	}
}

// TestTensorBroadcastTo_NotSubset verifies that a source whose dimensions
// are not a subset of the target returns an error.
func TestTensorBroadcastTo_NotSubset(t *testing.T) {
	src := NewChronoconeTensor([]any{1.0, 2.0}, []string{"Extra"}, []int{2})
	_, err := TensorBroadcastTo(src, []string{"Product", "Material"}, []int{3, 3})
	if err == nil {
		t.Error("TensorBroadcastTo with non-subset dimensions returned nil error, expected error")
	}
}

// TestTensorBroadcastTo_Identity verifies that when source and target dimensions
// match exactly, broadcasting produces an identical copy.
func TestTensorBroadcastTo_Identity(t *testing.T) {
	srcData := []any{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}
	src := NewChronoconeTensor(srcData, []string{"Product", "Material"}, []int{2, 3})
	// Set a mask bit to verify mask preservation in identity broadcast.
	src.SetMaskBit(2, false)

	result, err := TensorBroadcastTo(src, []string{"Product", "Material"}, []int{2, 3})
	if err != nil {
		t.Fatalf("TensorBroadcastTo returned error: %v", err)
	}
	if result == nil {
		t.Fatal("TensorBroadcastTo returned nil")
	}

	if len(result.Data()) != len(srcData) {
		t.Errorf("result.Data() length = %d, want %d", len(result.Data()), len(srcData))
	}
	for i := range srcData {
		if result.Data()[i] != srcData[i] {
			t.Errorf("result.Data()[%d] = %v, want %v", i, result.Data()[i], srcData[i])
		}
	}
	// Verify mask is preserved.
	if !result.IsMaskedTrue(0) {
		t.Errorf("IsMaskedTrue(0) = false, want true")
	}
	if result.IsMaskedTrue(2) {
		t.Errorf("IsMaskedTrue(2) = true, want false (mask bit should be preserved)")
	}
	if !result.IsMaskedTrue(3) {
		t.Errorf("IsMaskedTrue(3) = false, want true")
	}
}

// TestTensorBroadcastTo_WrongOrder verifies that source dimensions in wrong
// order relative to target returns an error.
func TestTensorBroadcastTo_WrongOrder(t *testing.T) {
	src := NewChronoconeTensor([]any{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0},
		[]string{"Material", "Product"}, []int{3, 3})
	_, err := TensorBroadcastTo(src, []string{"Product", "Material", "Daily"}, []int{3, 3, 7})
	if err == nil {
		t.Error("TensorBroadcastTo with wrong dimension order returned nil error, expected error")
	}
}

// TestTensorBroadcastTo_MaterialToProductMaterial verifies broadcasting
// a [Material:3] tensor to [Product:3, Material:3].
func TestTensorBroadcastTo_MaterialToProductMaterial(t *testing.T) {
	srcData := []any{1.0, 0.8, 0.9}
	src := NewChronoconeTensor(srcData, []string{"Material"}, []int{3})

	result, err := TensorBroadcastTo(src, []string{"Product", "Material"}, []int{3, 3})
	if err != nil {
		t.Fatalf("TensorBroadcastTo returned error: %v", err)
	}
	if result == nil {
		t.Fatal("TensorBroadcastTo returned nil")
	}

	shape := result.Shape()
	if len(shape) != 2 || shape[0] != 3 || shape[1] != 3 {
		t.Errorf("result.Shape() = %v, want [3 3]", shape)
	}

	data := result.Data()
	if len(data) != 9 {
		t.Errorf("result.Data() length = %d, want 9", len(data))
	}

	for product := 0; product < 3; product++ {
		for material := 0; material < 3; material++ {
			flatIdx := product*3 + material
			expected := srcData[material]
			if data[flatIdx] != expected {
				t.Errorf("data[%d] (Product=%d, Material=%d) = %v, want %v",
					flatIdx, product, material, data[flatIdx], expected)
			}
		}
	}
}

// TestTensorBroadcastTo_NilSource verifies that passing a nil source tensor
// returns a non-nil error.
func TestTensorBroadcastTo_NilSource(t *testing.T) {
	_, err := TensorBroadcastTo(nil, []string{"Product", "Material"}, []int{3, 3})
	if err == nil {
		t.Error("TensorBroadcastTo(nil, ...) returned nil error, expected error")
	}
}

// TestTensorBroadcastTo_ShapeMismatchSharedDim verifies that when source and
// target share a dimension name but with different sizes, an error is returned.
func TestTensorBroadcastTo_ShapeMismatchSharedDim(t *testing.T) {
	src := NewChronoconeTensor([]any{1.0, 2.0}, []string{"Material"}, []int{2})
	_, err := TensorBroadcastTo(src, []string{"Product", "Material", "Daily"}, []int{3, 3, 7})
	if err == nil {
		t.Error("TensorBroadcastTo with shape mismatch returned nil error, expected error")
	}
}

// TestTensorBroadcastTo_MismatchedTargetLengths verifies that mismatched
// targetDimNames and targetShape lengths return an error.
func TestTensorBroadcastTo_MismatchedTargetLengths(t *testing.T) {
	src := NewChronoconeTensor([]any{1.0, 2.0, 3.0}, []string{"Material"}, []int{3})
	_, err := TensorBroadcastTo(src, []string{"Product", "Material"}, []int{3, 3, 7})
	if err == nil {
		t.Error("TensorBroadcastTo with mismatched targetDimNames/targetShape lengths returned nil error, expected error")
	}
}

// TestTensorBroadcastTo_MaskPropagation verifies that invalid mask bits
// from the source are broadcast to all corresponding target positions.
func TestTensorBroadcastTo_MaskPropagation(t *testing.T) {
	srcData := []any{1.0, 2.0, 3.0}
	src := NewChronoconeTensor(srcData, []string{"Material"}, []int{3})
	// Mask index 1 (Aluminum) as invalid (bit = 1)
	src.SetMaskBit(1, false)

	result, err := TensorBroadcastTo(src, []string{"Product", "Material"}, []int{2, 3})
	if err != nil {
		t.Fatalf("TensorBroadcastTo returned error: %v", err)
	}

	// For each product, material 0 and 2 should be valid, material 1 should be invalid.
	for product := 0; product < 2; product++ {
		for material := 0; material < 3; material++ {
			flatIdx := product*3 + material
			isValid := result.IsMaskedTrue(flatIdx)
			if material == 1 {
				if isValid {
					t.Errorf("IsMaskedTrue(%d) (Product=%d, Material=1) = true, want false", flatIdx, product)
				}
			} else {
				if !isValid {
					t.Errorf("IsMaskedTrue(%d) (Product=%d, Material=%d) = false, want true", flatIdx, product, material)
				}
			}
		}
	}
}

// TestTensorDimensionNamesAreBroadcastable verifies that the broadcastable check
// correctly identifies compatible dimension subsets.
func TestTensorDimensionNamesAreBroadcastable(t *testing.T) {
	tests := []struct {
		name     string
		smaller  []string
		larger   []string
		expected bool
	}{
		{
			name:     "non-contiguous subset",
			smaller:  []string{"Product", "Daily"},
			larger:   []string{"Product", "Material", "Daily"},
			expected: true,
		},
		{
			name:     "single dim subset",
			smaller:  []string{"Material"},
			larger:   []string{"Product", "Material", "Daily"},
			expected: true,
		},
		{
			name:     "two dim subset",
			smaller:  []string{"Product", "Material"},
			larger:   []string{"Product", "Material", "Daily"},
			expected: true,
		},
		{
			name:     "wrong order",
			smaller:  []string{"Material", "Product"},
			larger:   []string{"Product", "Material", "Daily"},
			expected: false,
		},
		{
			name:     "not a subset",
			smaller:  []string{"Category"},
			larger:   []string{"Product", "Material", "Daily"},
			expected: false,
		},
		{
			name:     "equal dims",
			smaller:  []string{"Product", "Material", "Daily"},
			larger:   []string{"Product", "Material", "Daily"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TensorDimensionNamesAreBroadcastable(tt.smaller, tt.larger)
			if got != tt.expected {
				t.Errorf("TensorDimensionNamesAreBroadcastable(%v, %v) = %v, want %v",
					tt.smaller, tt.larger, got, tt.expected)
			}
		})
	}
}

// TestChronoconeTensorIsScalar verifies the IsScalar method for various tensor configurations.
func TestChronoconeTensorIsScalar(t *testing.T) {
	tests := []struct {
		name     string
		data     []float64
		dims     []string
		shape    []int
		expected bool
		isNil    bool
	}{
		{
			name:     "scalar tensor",
			data:     []float64{42.0},
			dims:     []string{},
			shape:    []int{},
			expected: true,
		},
		{
			name:     "multi-dimensional tensor",
			data:     []float64{1.0, 2.0, 3.0, 4.0},
			dims:     []string{"Product", "Region"},
			shape:    []int{2, 2},
			expected: false,
		},
		{
			name:     "empty tensor",
			data:     []float64{},
			dims:     []string{},
			shape:    []int{},
			expected: false,
		},
		{
			name:     "has dimensions but only one data point",
			data:     []float64{99.0},
			dims:     []string{"X"},
			shape:    []int{1},
			expected: false,
		},
		{
			name:     "has shape but no dims",
			data:     []float64{1.0, 2.0},
			dims:     []string{},
			shape:    []int{2},
			expected: false,
		},
		{
			name:     "nil tensor",
			data:     nil,
			dims:     nil,
			shape:    nil,
			expected: false,
			isNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tensor *ChronoconeTensor[float64]
			if !tt.isNil {
				tensor = NewChronoconeTensor(tt.data, tt.dims, tt.shape)
			}
			got := tensor.IsScalar()
			if got != tt.expected {
				t.Errorf("IsScalar() = %v, want %v", got, tt.expected)
			}
		})
	}
}
