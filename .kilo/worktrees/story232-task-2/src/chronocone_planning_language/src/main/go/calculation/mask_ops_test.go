// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package calculation

import (
	"testing"
)

// bit helpers operating on []uint64 slices, treating them as flat bit arrays.

// getMaskBit returns the value (0 or 1) of the bit at position pos within the []uint64 slice.
// pos is a flat bit index across all words.
func getMaskBit(mask []uint64, pos int) uint64 {
	if mask == nil || pos < 0 || pos >= len(mask)*64 {
		return 0
	}
	return (mask[pos>>6] >> (uint(pos) & 0x3F)) & 1
}

// setMaskBit sets the bit at position pos within the []uint64 slice to val (0 or 1).
func setMaskBit(mask []uint64, pos int, val uint64) {
	if mask == nil || pos < 0 || pos >= len(mask)*64 {
		return
	}
	if val == 0 {
		mask[pos>>6] &^= (1 << (uint(pos) & 0x3F))
	} else {
		mask[pos>>6] |= (1 << (uint(pos) & 0x3F))
	}
}

// makeMask creates a new []uint64 mask of nWords words, initialized to all zeros.
func makeMask(nWords int) []uint64 {
	return make([]uint64, nWords)
}

// =============================================================================
// TensorAndMask tests
// =============================================================================

// TestTensorAndMask_AllBitCombinations tests all four bit combinations
// (0|0, 0|1, 1|0, 1|1) at each of the 64 bit positions in a single word.
func TestTensorAndMask_AllBitCombinations(t *testing.T) {
	for bitPos := 0; bitPos < 64; bitPos++ {
		// AND truth table (convention: 0=true, 1=false):
		//   0,0 → AND → 0 (T AND T → T)
		//   0,1 → AND → 1 (T AND F → F)
		//   1,0 → AND → 1 (F AND T → F)
		//   1,1 → AND → 1 (F AND F → F)

		// 0,0
		a := makeMask(1)
		b := makeMask(1)
		result, err := TensorAndMask(a, b)
		if err != nil {
			t.Fatalf("bit %d: 0,0: unexpected error: %v", bitPos, err)
		}
		if got := getMaskBit(result, bitPos); got != 0 {
			t.Errorf("bit %d: 0,0: expected 0, got %d", bitPos, got)
		}

		// 0,1
		a = makeMask(1)
		b = makeMask(1)
		setMaskBit(b, bitPos, 1)
		result, err = TensorAndMask(a, b)
		if err != nil {
			t.Fatalf("bit %d: 0,1: unexpected error: %v", bitPos, err)
		}
		if got := getMaskBit(result, bitPos); got != 1 {
			t.Errorf("bit %d: 0,1: expected 1, got %d", bitPos, got)
		}

		// 1,0
		a = makeMask(1)
		b = makeMask(1)
		setMaskBit(a, bitPos, 1)
		result, err = TensorAndMask(a, b)
		if err != nil {
			t.Fatalf("bit %d: 1,0: unexpected error: %v", bitPos, err)
		}
		if got := getMaskBit(result, bitPos); got != 1 {
			t.Errorf("bit %d: 1,0: expected 1, got %d", bitPos, got)
		}

		// 1,1
		a = makeMask(1)
		b = makeMask(1)
		setMaskBit(a, bitPos, 1)
		setMaskBit(b, bitPos, 1)
		result, err = TensorAndMask(a, b)
		if err != nil {
			t.Fatalf("bit %d: 1,1: unexpected error: %v", bitPos, err)
		}
		if got := getMaskBit(result, bitPos); got != 1 {
			t.Errorf("bit %d: 1,1: expected 1, got %d", bitPos, got)
		}
	}
}

// TestTensorOrMask_AllBitCombinations tests all four bit combinations
// (0|0, 0|1, 1|0, 1|1) at each of the 64 bit positions in a single word.
func TestTensorOrMask_AllBitCombinations(t *testing.T) {
	for bitPos := 0; bitPos < 64; bitPos++ {
		// OR truth table (convention: 0=true, 1=false):
		//   0,0 → OR → 0 (T OR T → T)
		//   0,1 → OR → 0 (T OR F → T)
		//   1,0 → OR → 0 (F OR T → T)
		//   1,1 → OR → 1 (F OR F → F)

		// 0,0
		a := makeMask(1)
		b := makeMask(1)
		result, err := TensorOrMask(a, b)
		if err != nil {
			t.Fatalf("bit %d: 0,0: unexpected error: %v", bitPos, err)
		}
		if got := getMaskBit(result, bitPos); got != 0 {
			t.Errorf("bit %d: 0,0: expected 0, got %d", bitPos, got)
		}

		// 0,1
		a = makeMask(1)
		b = makeMask(1)
		setMaskBit(b, bitPos, 1)
		result, err = TensorOrMask(a, b)
		if err != nil {
			t.Fatalf("bit %d: 0,1: unexpected error: %v", bitPos, err)
		}
		if got := getMaskBit(result, bitPos); got != 0 {
			t.Errorf("bit %d: 0,1: expected 0, got %d", bitPos, got)
		}

		// 1,0
		a = makeMask(1)
		b = makeMask(1)
		setMaskBit(a, bitPos, 1)
		result, err = TensorOrMask(a, b)
		if err != nil {
			t.Fatalf("bit %d: 1,0: unexpected error: %v", bitPos, err)
		}
		if got := getMaskBit(result, bitPos); got != 0 {
			t.Errorf("bit %d: 1,0: expected 0, got %d", bitPos, got)
		}

		// 1,1
		a = makeMask(1)
		b = makeMask(1)
		setMaskBit(a, bitPos, 1)
		setMaskBit(b, bitPos, 1)
		result, err = TensorOrMask(a, b)
		if err != nil {
			t.Fatalf("bit %d: 1,1: unexpected error: %v", bitPos, err)
		}
		if got := getMaskBit(result, bitPos); got != 1 {
			t.Errorf("bit %d: 1,1: expected 1, got %d", bitPos, got)
		}
	}
}

// =============================================================================
// TensorNotMask tests
// =============================================================================

// TestTensorNotMask_BasicInversion tests that NOT flips all bits: 0→1, 1→0.
func TestTensorNotMask_BasicInversion(t *testing.T) {
	// All zeros: NOT should produce all-ones (^0 = 0xFFFFFFFFFFFFFFFF)
	mask := []uint64{0}
	result := TensorNotMask(mask)
	if result == nil {
		t.Fatal("TensorNotMask([0]) returned nil")
	}
	if result[0] != ^uint64(0) {
		t.Errorf("TensorNotMask([0])[0] = %#x, expected %#x", result[0], ^uint64(0))
	}

	// All ones: NOT should produce all-zeros
	mask = []uint64{^uint64(0)}
	result = TensorNotMask(mask)
	if result == nil {
		t.Fatal("TensorNotMask([^0]) returned nil")
	}
	if result[0] != 0 {
		t.Errorf("TensorNotMask([^0])[0] = %#x, expected 0", result[0])
	}

	// Mixed pattern: 0x0F0F0F0F0F0F0F0F, NOT should produce 0xF0F0F0F0F0F0F0F0
	input := uint64(0x0F0F0F0F0F0F0F0F)
	expected := ^input
	mask = []uint64{input}
	result = TensorNotMask(mask)
	if result == nil {
		t.Fatal("TensorNotMask([0x0F...]) returned nil")
	}
	if result[0] != expected {
		t.Errorf("TensorNotMask([%#x])[0] = %#x, expected %#x", input, result[0], expected)
	}

	// Single bit: NOT of 0x01 should flip bit 0 and everything above
	inputBytes := uint64(0x01)
	expectedBytes := ^inputBytes // 0xFFFFFFFFFFFFFFFE
	mask = []uint64{inputBytes}
	result = TensorNotMask(mask)
	if result == nil {
		t.Fatal("TensorNotMask([1]) returned nil")
	}
	if result[0] != expectedBytes {
		t.Errorf("TensorNotMask([1])[0] = %#x, expected %#x", result[0], expectedBytes)
	}

	// Verify each bit position individually
	for bitPos := 0; bitPos < 64; bitPos++ {
		word := uint64(1) << uint(bitPos)
		mask := []uint64{word}
		result := TensorNotMask(mask)
		if result == nil {
			t.Fatalf("bit %d: TensorNotMask returned nil", bitPos)
		}
		// The bit at bitPos should be 0 (was 1, now 0)
		if got := getMaskBit(result, bitPos); got != 0 {
			t.Errorf("bit %d: NOT(%#x): expected 0, got %d", bitPos, word, got)
		}
		// Some other bit should be 1 (was 0, now 1)
		otherPos := (bitPos + 3) % 64
		if got := getMaskBit(result, otherPos); got != 1 {
			t.Errorf("bit %d: NOT(%#x) at other bit %d: expected 1, got %d", bitPos, word, otherPos, got)
		}
	}
}

// =============================================================================
// Multi-word tests
// =============================================================================

// TestMaskOps_DifferentLengths tests operations on masks of 1, 2, and 5 words.
func TestMaskOps_DifferentLengths(t *testing.T) {
	lengths := []int{1, 2, 5}
	for _, nWords := range lengths {
		a := makeMask(nWords)
		b := makeMask(nWords)
		// Set all bits in a to 1, b to 0
		for i := range a {
			a[i] = ^uint64(0)
			b[i] = 0
		}

		// AND with a=all-1s (all false), b=all-0s (all true):
		// Logical AND: false AND true = false → result bit=1
		result, err := TensorAndMask(a, b)
		if err != nil {
			t.Fatalf("%d words: TensorAndMask error: %v", nWords, err)
		}
		for i, w := range result {
			if w != ^uint64(0) {
				t.Errorf("%d words: TensorAndMask word[%d] = %#x, expected all-ones", nWords, i, w)
			}
		}

		// OR with a=all-1s (all false), b=all-0s (all true):
		// Logical OR: false OR true = true → result bit=0
		result, err = TensorOrMask(a, b)
		if err != nil {
			t.Fatalf("%d words: TensorOrMask error: %v", nWords, err)
		}
		for i, w := range result {
			if w != 0 {
				t.Errorf("%d words: TensorOrMask word[%d] = %#x, expected 0", nWords, i, w)
			}
		}

		// NOT on all-zeros mask:
		notMask := makeMask(nWords)
		notResult := TensorNotMask(notMask)
		if notResult == nil {
			t.Fatalf("%d words: TensorNotMask returned nil", nWords)
		}
		for i, w := range notResult {
			if w != ^uint64(0) {
				t.Errorf("%d words: TensorNotMask word[%d] = %#x, expected all-ones", nWords, i, w)
			}
		}

		// NOT on all-ones mask:
		allOnes := makeMask(nWords)
		for i := range allOnes {
			allOnes[i] = ^uint64(0)
		}
		notResult = TensorNotMask(allOnes)
		if notResult == nil {
			t.Fatalf("%d words: TensorNotMask(all-ones) returned nil", nWords)
		}
		for i, w := range notResult {
			if w != 0 {
				t.Errorf("%d words: TensorNotMask(all-ones) word[%d] = %#x, expected 0", nWords, i, w)
			}
		}
	}
}

// =============================================================================
// Nil input handling
// =============================================================================

// TestTensorAndMask_NilInputs tests that nil inputs return errors.
func TestTensorAndMask_NilInputs(t *testing.T) {
	valid := []uint64{0}

	// Both nil
	_, err := TensorAndMask(nil, nil)
	if err == nil {
		t.Error("TensorAndMask(nil, nil): expected error, got nil")
	}

	// First nil
	_, err = TensorAndMask(nil, valid)
	if err == nil {
		t.Error("TensorAndMask(nil, valid): expected error, got nil")
	}

	// Second nil
	_, err = TensorAndMask(valid, nil)
	if err == nil {
		t.Error("TensorAndMask(valid, nil): expected error, got nil")
	}
}

// TestTensorOrMask_NilInputs tests that nil inputs return errors.
func TestTensorOrMask_NilInputs(t *testing.T) {
	valid := []uint64{0}

	// Both nil
	_, err := TensorOrMask(nil, nil)
	if err == nil {
		t.Error("TensorOrMask(nil, nil): expected error, got nil")
	}

	// First nil
	_, err = TensorOrMask(nil, valid)
	if err == nil {
		t.Error("TensorOrMask(nil, valid): expected error, got nil")
	}

	// Second nil
	_, err = TensorOrMask(valid, nil)
	if err == nil {
		t.Error("TensorOrMask(valid, nil): expected error, got nil")
	}
}

// TestTensorNotMask_NilInput tests that nil input returns nil.
func TestTensorNotMask_NilInput(t *testing.T) {
	result := TensorNotMask(nil)
	if result != nil {
		t.Errorf("TensorNotMask(nil): expected nil, got %v", result)
	}
}

// =============================================================================
// Length mismatch tests
// =============================================================================

// TestTensorAndMask_LengthMismatch tests that length mismatches produce an error.
func TestTensorAndMask_LengthMismatch(t *testing.T) {
	a := []uint64{0, 1}
	b := []uint64{0, 1, 2}
	_, err := TensorAndMask(a, b)
	if err == nil {
		t.Error("TensorAndMask([2 words], [3 words]): expected error, got nil")
	}
}

// TestTensorOrMask_LengthMismatch tests that length mismatches produce an error.
func TestTensorOrMask_LengthMismatch(t *testing.T) {
	a := []uint64{0, 1}
	b := []uint64{0, 1, 2}
	_, err := TensorOrMask(a, b)
	if err == nil {
		t.Error("TensorOrMask([2 words], [3 words]): expected error, got nil")
	}
}

// =============================================================================
// Word-boundary tests
// =============================================================================

// TestMaskOps_WordBoundary tests bits just below and above the 64-bit boundary
// (bits 63 and 64 spanning words 0 and 1).
func TestMaskOps_WordBoundary(t *testing.T) {
	// 2-word mask: bits 0-63 in word[0], bits 64-127 in word[1]
	// Boundary is between bit 63 (last bit of word[0]) and bit 64 (first bit of word[1])

	// Test AND at boundary
	a := makeMask(2)
	b := makeMask(2)
	// Set bit 63 in a: first word's MSB
	setMaskBit(a, 63, 0)
	// Set bit 64 in b: second word's LSB
	setMaskBit(b, 64, 0)
	// All other bits are 0 (true)

	result, err := TensorAndMask(a, b)
	if err != nil {
		t.Fatal(err)
	}
	// Since all bits are 0 in both (true AND true = true), result should be all zeros
	for i, w := range result {
		if w != 0 {
			t.Errorf("TensorAndMask at boundary: word[%d] = %#x, expected 0", i, w)
		}
	}

	// Test bit 63 with different combinations (AND)
	// A[63]=0 (true), B[63]=1 (false) → AND: 0 AND 1 = false → result bit 63 should be 1
	a = makeMask(2)
	b = makeMask(2)
	// A all zeros
	setMaskBit(b, 63, 1)
	result, err = TensorAndMask(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if got := getMaskBit(result, 63); got != 1 {
		t.Errorf("bit 63 (AND 0,1): expected 1, got %d", got)
	}

	// Test bit 64 (second word first bit) with OR
	a = makeMask(2)
	b = makeMask(2)
	// A[64]=1 (false), B[64]=0 (true) → OR: 1 OR 0 = true → result bit 64 should be 0
	setMaskBit(a, 64, 1)
	result, err = TensorOrMask(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if got := getMaskBit(result, 64); got != 0 {
		t.Errorf("bit 64 (OR 1,0): expected 0, got %d", got)
	}

	// Test NOT at boundary: verify bits 63 and 64 flip correctly
	mask := makeMask(2)
	setMaskBit(mask, 63, 1) // set bit 63 to 1 (false)
	setMaskBit(mask, 64, 1) // set bit 64 to 1 (false)
	// all other bits are 0 (true)

	notResult := TensorNotMask(mask)
	if notResult == nil {
		t.Fatal("TensorNotMask at boundary returned nil")
	}
	// Bit 63 should be 0 (was 1)
	if got := getMaskBit(notResult, 63); got != 0 {
		t.Errorf("NOT bit 63: expected 0, got %d", got)
	}
	// Bit 64 should be 0 (was 1)
	if got := getMaskBit(notResult, 64); got != 0 {
		t.Errorf("NOT bit 64: expected 0, got %d", got)
	}
	// Other bits should be 1 (were 0)
	for bitPos := 0; bitPos < 128; bitPos++ {
		if bitPos == 63 || bitPos == 64 {
			continue
		}
		if got := getMaskBit(notResult, bitPos); got != 1 {
			t.Errorf("NOT bit %d: expected 1, got %d", bitPos, got)
		}
	}
}

// =============================================================================
// Composition test: AND then OR, OR then NOT
// =============================================================================

// TestMaskOps_Composition tests chaining of mask operations.
func TestMaskOps_Composition(t *testing.T) {
	// AND of two masks followed by OR with a third
	a := []uint64{0x0F, 0xF0}
	b := []uint64{0x33, 0xCC}
	c := []uint64{0x55, 0xAA}

	// AND(a, b): using | for logical AND under inverted convention
	// a|b: 0x0F|0x33 = 0x3F, 0xF0|0xCC = 0xFC
	andAB, err := TensorAndMask(a, b)
	if err != nil {
		t.Fatal(err)
	}

	// OR(andAB, c): using & for logical OR under inverted convention
	// andAB & c: 0x3F&0x55 = 0x15, 0xFC&0xAA = 0xA8
	final, err := TensorOrMask(andAB, c)
	if err != nil {
		t.Fatal(err)
	}

	if final[0] != 0x15 {
		t.Errorf("composition AND+OR word[0] = %#x, expected 0x15", final[0])
	}
	if final[1] != 0xA8 {
		t.Errorf("composition AND+OR word[1] = %#x, expected 0xA8", final[1])
	}

	// NOT of the final result
	notFinal := TensorNotMask(final)
	if notFinal == nil {
		t.Fatal("TensorNotMask in composition test returned nil")
	}
	if notFinal[0] != ^uint64(0x15) {
		t.Errorf("NOT of composition word[0] = %#x, expected %#x", notFinal[0], ^uint64(0x15))
	}
	if notFinal[1] != ^uint64(0xA8) {
		t.Errorf("NOT of composition word[1] = %#x, expected %#x", notFinal[1], ^uint64(0xA8))
	}
}

// =============================================================================
// AllTrueMask tests
// =============================================================================

// TestAllTrueMask_ZeroAndNegative returns an empty slice for n <= 0.
func TestAllTrueMask_ZeroAndNegative(t *testing.T) {
	m := AllTrueMask(0)
	if m == nil {
		t.Fatal("AllTrueMask(0) returned nil, expected empty slice")
	}
	if len(m) != 0 {
		t.Errorf("AllTrueMask(0) len = %d, expected 0", len(m))
	}

	mNeg := AllTrueMask(-5)
	if mNeg == nil {
		t.Fatal("AllTrueMask(-5) returned nil, expected empty slice")
	}
	if len(mNeg) != 0 {
		t.Errorf("AllTrueMask(-5) len = %d, expected 0", len(mNeg))
	}
}

// TestAllTrueMask_SingleElement verifies one element produces one word of all zeros.
func TestAllTrueMask_SingleElement(t *testing.T) {
	m := AllTrueMask(1)
	if len(m) != 1 {
		t.Fatalf("AllTrueMask(1) len = %d, expected 1", len(m))
	}
	if m[0] != 0 {
		t.Errorf("AllTrueMask(1)[0] = 0x%016X, expected 0", m[0])
	}
}

// TestAllTrueMask_MultiWord verifies n=65 produces 2 words of all zeros.
func TestAllTrueMask_MultiWord(t *testing.T) {
	m := AllTrueMask(65)
	if len(m) != 2 {
		t.Fatalf("AllTrueMask(65) len = %d, expected 2", len(m))
	}
	for i, word := range m {
		if word != 0 {
			t.Errorf("AllTrueMask(65)[%d] = 0x%016X, expected 0", i, word)
		}
	}
}

// TestAllTrueMask_BoundarySizes verifies correct word counts at 64-element boundaries.
func TestAllTrueMask_BoundarySizes(t *testing.T) {
	tests := []struct {
		n       int
		wantLen int
	}{
		{1, 1},
		{63, 1},
		{64, 1},
		{65, 2},
		{128, 2},
		{129, 3},
	}
	for _, tt := range tests {
		m := AllTrueMask(tt.n)
		if len(m) != tt.wantLen {
			t.Errorf("AllTrueMask(%d) len = %d, expected %d", tt.n, len(m), tt.wantLen)
		}
		for i, word := range m {
			if word != 0 {
				t.Errorf("AllTrueMask(%d)[%d] = 0x%016X, expected 0", tt.n, i, word)
			}
		}
	}
}

// =============================================================================
// AllFalseMask tests
// =============================================================================

// TestAllFalseMask_ZeroAndNegative returns an empty slice for n <= 0.
func TestAllFalseMask_ZeroAndNegative(t *testing.T) {
	m := AllFalseMask(0)
	if m == nil {
		t.Fatal("AllFalseMask(0) returned nil, expected empty slice")
	}
	if len(m) != 0 {
		t.Errorf("AllFalseMask(0) len = %d, expected 0", len(m))
	}

	mNeg := AllFalseMask(-3)
	if mNeg == nil {
		t.Fatal("AllFalseMask(-3) returned nil, expected empty slice")
	}
	if len(mNeg) != 0 {
		t.Errorf("AllFalseMask(-3) len = %d, expected 0", len(mNeg))
	}
}

// TestAllFalseMask_SingleElement verifies one element produces one word of all ones.
func TestAllFalseMask_SingleElement(t *testing.T) {
	m := AllFalseMask(1)
	if len(m) != 1 {
		t.Fatalf("AllFalseMask(1) len = %d, expected 1", len(m))
	}
	if m[0] != ^uint64(0) {
		t.Errorf("AllFalseMask(1)[0] = 0x%016X, expected all-ones", m[0])
	}
}

// TestAllFalseMask_MultiWord verifies n=65 produces 2 words of all ones.
func TestAllFalseMask_MultiWord(t *testing.T) {
	m := AllFalseMask(65)
	if len(m) != 2 {
		t.Fatalf("AllFalseMask(65) len = %d, expected 2", len(m))
	}
	for i, word := range m {
		if word != ^uint64(0) {
			t.Errorf("AllFalseMask(65)[%d] = 0x%016X, expected all-ones", i, word)
		}
	}
}

// TestAllFalseMask_BoundarySizes verifies correct word counts at 64-element boundaries.
func TestAllFalseMask_BoundarySizes(t *testing.T) {
	tests := []struct {
		n       int
		wantLen int
	}{
		{1, 1},
		{63, 1},
		{64, 1},
		{65, 2},
		{128, 2},
		{129, 3},
	}
	for _, tt := range tests {
		m := AllFalseMask(tt.n)
		if len(m) != tt.wantLen {
			t.Errorf("AllFalseMask(%d) len = %d, expected %d", tt.n, len(m), tt.wantLen)
		}
		for i, word := range m {
			if word != ^uint64(0) {
				t.Errorf("AllFalseMask(%d)[%d] = 0x%016X, expected all-ones", tt.n, i, word)
			}
		}
	}
}

// =============================================================================
// Edge case: empty mask
// =============================================================================

// TestMaskOps_EmptyMask tests zero-length masks.
func TestMaskOps_EmptyMask(t *testing.T) {
	a := []uint64{}
	b := []uint64{}

	// AND on empty slices
	result, err := TensorAndMask(a, b)
	if err != nil {
		t.Fatalf("TensorAndMask(empty, empty) unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("TensorAndMask(empty, empty) len = %d, expected 0", len(result))
	}

	// OR on empty slices
	result, err = TensorOrMask(a, b)
	if err != nil {
		t.Fatalf("TensorOrMask(empty, empty) unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("TensorOrMask(empty, empty) len = %d, expected 0", len(result))
	}

	// NOT on empty slice
	notResult := TensorNotMask(a)
	if notResult == nil {
		t.Fatal("TensorNotMask(empty) returned nil")
	}
	if len(notResult) != 0 {
		t.Errorf("TensorNotMask(empty) len = %d, expected 0", len(notResult))
	}

	// Length mismatch with empty and non-empty
	nonEmpty := []uint64{0}
	_, err = TensorAndMask(a, nonEmpty)
	if err == nil {
		t.Error("TensorAndMask(empty, non-empty): expected error, got nil")
	}
	_, err = TensorOrMask(nonEmpty, a)
	if err == nil {
		t.Error("TensorOrMask(non-empty, empty): expected error, got nil")
	}
}
