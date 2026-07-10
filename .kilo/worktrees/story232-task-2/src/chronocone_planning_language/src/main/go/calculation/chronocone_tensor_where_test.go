// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package calculation

import (
	"testing"
)

// =============================================================================
// TensorWhereFloat64 tests
// =============================================================================

// TestTensorWhereFloat64_SingleBranch verifies that a single branch with all-zero
// mask is selected for every position.
func TestTensorWhereFloat64_SingleBranch(t *testing.T) {
	data := []float64{1.0, 2.0, 3.0, 4.0}
	branch := makeMaskedAnyTensor(data, []uint64{0})

	result, err := TensorWhereFloat64([]*ChronoconeTensor[any]{branch})
	if err != nil {
		t.Fatalf("TensorWhereFloat64 failed: %v", err)
	}
	if result == nil {
		t.Fatal("TensorWhereFloat64 returned nil result")
	}

	// Verify data matches
	for i, want := range data {
		if result.Data()[i] != want {
			t.Errorf("result[%d] = %v, want %v", i, result.Data()[i], want)
		}
	}

	// Verify output mask is all zeros (all true)
	for i, word := range result.Mask() {
		if word != 0 {
			t.Errorf("result mask[%d] = 0x%016X, want 0", i, word)
		}
	}
}

// TestTensorWhereFloat64_TwoBranchesComplementary verifies if/else behavior:
// branch 0 selects first half, branch 1 selects second half.
func TestTensorWhereFloat64_TwoBranchesComplementary(t *testing.T) {
	// 4 elements: branch 0 selects positions 0,1; branch 1 selects positions 2,3
	mask0 := []uint64{0x0C} // bits 2,3 = 1 (false), bits 0,1 = 0 (true)
	mask1 := []uint64{0x03} // bits 0,1 = 1 (false), bits 2,3 = 0 (true)

	branch0 := makeMaskedAnyTensor([]float64{10.0, 20.0, 30.0, 40.0}, mask0)
	branch1 := makeMaskedAnyTensor([]float64{100.0, 200.0, 300.0, 400.0}, mask1)

	result, err := TensorWhereFloat64([]*ChronoconeTensor[any]{branch0, branch1})
	if err != nil {
		t.Fatalf("TensorWhereFloat64 failed: %v", err)
	}

	expected := []float64{10.0, 20.0, 300.0, 400.0}
	for i, want := range expected {
		if result.Data()[i] != want {
			t.Errorf("result[%d] = %v, want %v", i, result.Data()[i], want)
		}
	}
}

// TestTensorWhereFloat64_ThreeBranchesDisjoint verifies if/elseif/else behavior
// with three disjoint masks, each selecting a different third.
func TestTensorWhereFloat64_ThreeBranchesDisjoint(t *testing.T) {
	// 6 elements: branch 0 → positions 0,1; branch 1 → positions 2,3; branch 2 → positions 4,5
	mask0 := []uint64{0x3C} // bits 2,3,4,5 = 1 → positions 0,1 selected
	mask1 := []uint64{0x33} // bits 0,1,4,5 = 1 → positions 2,3 selected
	mask2 := []uint64{0x00} // all selected, but only positions 4,5 fall through

	branch0 := makeMaskedAnyTensor([]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}, mask0)
	branch1 := makeMaskedAnyTensor([]float64{10.0, 20.0, 30.0, 40.0, 50.0, 60.0}, mask1)
	branch2 := makeMaskedAnyTensor([]float64{100.0, 200.0, 300.0, 400.0, 500.0, 600.0}, mask2)

	result, err := TensorWhereFloat64([]*ChronoconeTensor[any]{branch0, branch1, branch2})
	if err != nil {
		t.Fatalf("TensorWhereFloat64 failed: %v", err)
	}

	expected := []float64{1.0, 2.0, 30.0, 40.0, 500.0, 600.0}
	for i, want := range expected {
		if result.Data()[i] != want {
			t.Errorf("result[%d] = %v, want %v", i, result.Data()[i], want)
		}
	}
}

// TestTensorWhereFloat64_OverlappingMasks verifies that the first branch wins
// when multiple branches have mask bit=0 at the same position.
func TestTensorWhereFloat64_OverlappingMasks(t *testing.T) {
	// 4 elements: both branches have all-zero masks (all selected)
	branch0 := makeMaskedAnyTensor([]float64{1.0, 2.0, 3.0, 4.0}, []uint64{0})
	branch1 := makeMaskedAnyTensor([]float64{100.0, 200.0, 300.0, 400.0}, []uint64{0})

	result, err := TensorWhereFloat64([]*ChronoconeTensor[any]{branch0, branch1})
	if err != nil {
		t.Fatalf("TensorWhereFloat64 failed: %v", err)
	}

	// First branch should win for all positions
	expected := []float64{1.0, 2.0, 3.0, 4.0}
	for i, want := range expected {
		if result.Data()[i] != want {
			t.Errorf("result[%d] = %v, want %v", i, result.Data()[i], want)
		}
	}
}

// TestTensorWhereFloat64_BranchDimensionMismatch verifies error when branches
// have different data lengths.
func TestTensorWhereFloat64_BranchDimensionMismatch(t *testing.T) {
	branch0 := makeMaskedAnyTensor([]float64{1.0, 2.0}, []uint64{0})
	branch1 := makeMaskedAnyTensor([]float64{1.0, 2.0, 3.0}, []uint64{0})

	_, err := TensorWhereFloat64([]*ChronoconeTensor[any]{branch0, branch1})
	if err == nil {
		t.Error("TensorWhereFloat64 with mismatched data lengths: expected error, got nil")
	}
}

// TestTensorWhereFloat64_BranchShapeMismatch verifies error when branches
// have different shapes.
func TestTensorWhereFloat64_BranchShapeMismatch(t *testing.T) {
	data0 := []any{1.0, 2.0, 3.0, 4.0}
	data1 := []any{1.0, 2.0, 3.0, 4.0}
	branch0 := NewChronoconeTensor(data0, []string{"A", "B"}, []int{2, 2})
	branch1 := NewChronoconeTensor(data1, []string{"A", "B"}, []int{4, 1})

	_, err := TensorWhereFloat64([]*ChronoconeTensor[any]{branch0, branch1})
	if err == nil {
		t.Error("TensorWhereFloat64 with mismatched shapes: expected error, got nil")
	}
}

// TestTensorWhereFloat64_NilBranch verifies error when a branch in the slice is nil.
func TestTensorWhereFloat64_NilBranch(t *testing.T) {
	branch0 := makeMaskedAnyTensor([]float64{1.0, 2.0}, []uint64{0})

	_, err := TensorWhereFloat64([]*ChronoconeTensor[any]{branch0, nil})
	if err == nil {
		t.Error("TensorWhereFloat64 with nil branch: expected error, got nil")
	}
}

// TestTensorWhereFloat64_EmptyBranches verifies error when branches slice is empty.
func TestTensorWhereFloat64_EmptyBranches(t *testing.T) {
	_, err := TensorWhereFloat64([]*ChronoconeTensor[any]{})
	if err == nil {
		t.Error("TensorWhereFloat64 with empty branches: expected error, got nil")
	}
}

// TestTensorWhereFloat64_NilBranchesSlice verifies error when branches slice is nil.
func TestTensorWhereFloat64_NilBranchesSlice(t *testing.T) {
	_, err := TensorWhereFloat64(nil)
	if err == nil {
		t.Error("TensorWhereFloat64 with nil branches slice: expected error, got nil")
	}
}

// TestTensorWhereFloat64_NilMaskEdgeCase verifies that a nil mask is treated as
// all-zero mask (all true/selected), so the first branch's values are used.
func TestTensorWhereFloat64_NilMaskEdgeCase(t *testing.T) {
	data := []float64{1.0, 2.0, 3.0, 4.0}
	branch := makeMaskedAnyTensor(data, nil)
	// nil mask should be treated as all true → first branch wins

	result, err := TensorWhereFloat64([]*ChronoconeTensor[any]{branch})
	if err != nil {
		t.Fatalf("TensorWhereFloat64 failed: %v", err)
	}

	for i, want := range data {
		if result.Data()[i] != want {
			t.Errorf("result[%d] = %v, want %v", i, result.Data()[i], want)
		}
	}
}

// TestTensorWhereFloat64_NilMaskWithMultipleBranches verifies nil mask behavior
// when there are multiple branches: first branch with nil mask wins all positions.
func TestTensorWhereFloat64_NilMaskWithMultipleBranches(t *testing.T) {
	branch0 := makeMaskedAnyTensor([]float64{1.0, 2.0, 3.0, 4.0}, nil)
	branch1 := makeMaskedAnyTensor([]float64{100.0, 200.0, 300.0, 400.0}, []uint64{0})

	result, err := TensorWhereFloat64([]*ChronoconeTensor[any]{branch0, branch1})
	if err != nil {
		t.Fatalf("TensorWhereFloat64 failed: %v", err)
	}

	expected := []float64{1.0, 2.0, 3.0, 4.0}
	for i, want := range expected {
		if result.Data()[i] != want {
			t.Errorf("result[%d] = %v, want %v", i, result.Data()[i], want)
		}
	}
}

// TestTensorWhereFloat64_AllBranchesMaskedTrue verifies that when all branches
// have mask bit=1 at a position, the last branch's value is used as fallback.
func TestTensorWhereFloat64_AllBranchesMaskedTrue(t *testing.T) {
	// 4 elements: all branches have all-ones mask (all false/unselected)
	branch0 := makeMaskedAnyTensor([]float64{1.0, 2.0, 3.0, 4.0}, []uint64{^uint64(0)})
	branch1 := makeMaskedAnyTensor([]float64{100.0, 200.0, 300.0, 400.0}, []uint64{^uint64(0)})

	result, err := TensorWhereFloat64([]*ChronoconeTensor[any]{branch0, branch1})
	if err != nil {
		t.Fatalf("TensorWhereFloat64 failed: %v", err)
	}

	// Last branch's values should be used as fallback
	expected := []float64{100.0, 200.0, 300.0, 400.0}
	for i, want := range expected {
		if result.Data()[i] != want {
			t.Errorf("result[%d] = %v, want %v", i, result.Data()[i], want)
		}
	}
}

// TestTensorWhereFloat64_DimensionNameMismatch verifies error when branches
// have different dimension names.
func TestTensorWhereFloat64_DimensionNameMismatch(t *testing.T) {
	data0 := []any{1.0, 2.0}
	data1 := []any{1.0, 2.0}
	branch0 := NewChronoconeTensor(data0, []string{"X"}, []int{2})
	branch1 := NewChronoconeTensor(data1, []string{"Y"}, []int{2})

	_, err := TensorWhereFloat64([]*ChronoconeTensor[any]{branch0, branch1})
	if err == nil {
		t.Error("TensorWhereFloat64 with mismatched dimension names: expected error, got nil")
	}
}

// TestTensorWhereFloat64_MultiWordMask verifies Where operation with masks
// spanning multiple uint64 words.
func TestTensorWhereFloat64_MultiWordMask(t *testing.T) {
	// 100 elements across 2 words
	data0 := make([]float64, 100)
	data1 := make([]float64, 100)
	for i := 0; i < 100; i++ {
		data0[i] = float64(i)
		data1[i] = float64(i + 1000)
	}

	// mask0: select first 50 elements (positions 0-49 true, 50-99 false)
	mask0 := make([]uint64, 2)
	for i := 50; i < 100; i++ {
		mask0[i>>6] |= 1 << (uint(i) & 0x3F)
	}
	// mask1: select last 50 elements (positions 0-49 false, 50-99 true)
	mask1 := []uint64{^uint64(0), ^uint64(0)}
	for i := 50; i < 100; i++ {
		mask1[i>>6] &^= 1 << (uint(i) & 0x3F)
	}

	branch0 := makeMaskedAnyTensor(data0, mask0)
	branch1 := makeMaskedAnyTensor(data1, mask1)

	result, err := TensorWhereFloat64([]*ChronoconeTensor[any]{branch0, branch1})
	if err != nil {
		t.Fatalf("TensorWhereFloat64 failed: %v", err)
	}
	if len(result.Mask()) != 2 {
		t.Errorf("mask length: got %d, want 2", len(result.Mask()))
	}

	for i := 0; i < 100; i++ {
		var want float64
		if i < 50 {
			want = data0[i]
		} else {
			want = data1[i]
		}
		if result.Data()[i] != want {
			t.Errorf("result[%d] = %v, want %v", i, result.Data()[i], want)
		}
	}
}
