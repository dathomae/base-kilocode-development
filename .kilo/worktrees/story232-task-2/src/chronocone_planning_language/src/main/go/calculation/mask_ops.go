// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package calculation

import (
	"fmt"
	"math"
)

// Float64TruthinessMask computes a mask from float64 data using truthiness rules:
// zero (0) → false (mask bit = 1 = unselected), non-zero → true (mask bit = 0 = selected).
//
// NaN values are treated as false (unselected).
//
// Mask convention: bit=0 means true/selected, bit=1 means false/unselected.
//
// Parameters:
//   - data: The float64 data slice to convert to a truthiness mask.
//
// Returns:
//   - []uint64: A new mask with bits set according to truthiness rules.
func Float64TruthinessMask(data []float64) []uint64 {
	n := len(data)
	if n == 0 {
		return []uint64{}
	}
	mask := make([]uint64, (n+63)/64)
	for i, v := range data {
		if v == 0 || math.IsNaN(v) {
			// false: mark as unselected (bit = 1)
			mask[i>>6] |= (1 << (uint(i) & 0x3F))
		}
		// true: bit stays 0 (already initialized to zero)
	}
	return mask
}

// TensorAndMask combines two boolean masks using logical AND and returns a new mask.
//
// Mask convention: bit=0 means true/selected, bit=1 means false/unselected.
// Under this inverted convention, logical AND (both true → result true) maps to
// bitwise OR (a | b):
//
//	0|0=0 (T AND T → T), 0|1=1 (T AND F → F), 1|0=1 (F AND T → F), 1|1=1 (F AND F → F)
//
// Parameters:
//   - a: The first mask slice.
//   - b: The second mask slice.
//
// Returns:
//   - []uint64: A new mask containing the element-wise logical AND.
//   - error: An error if either mask is nil or lengths differ.
func TensorAndMask(a, b []uint64) ([]uint64, error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("mask is nil")
	}
	if len(a) != len(b) {
		return nil, fmt.Errorf("mask length mismatch: %d != %d", len(a), len(b))
	}
	result := make([]uint64, len(a))
	for i := range a {
		result[i] = a[i] | b[i]
	}
	return result, nil
}

// TensorOrMask combines two boolean masks using logical OR and returns a new mask.
//
// Mask convention: bit=0 means true/selected, bit=1 means false/unselected.
// Under this inverted convention, logical OR (at least one true → result true) maps to
// bitwise AND (a & b):
//
//	0&0=0 (T OR T → T), 0&1=0 (T OR F → T), 1&0=0 (F OR T → T), 1&1=1 (F OR F → F)
//
// Parameters:
//   - a: The first mask slice.
//   - b: The second mask slice.
//
// Returns:
//   - []uint64: A new mask containing the element-wise logical OR.
//   - error: An error if either mask is nil or lengths differ.
func TensorOrMask(a, b []uint64) ([]uint64, error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("mask is nil")
	}
	if len(a) != len(b) {
		return nil, fmt.Errorf("mask length mismatch: %d != %d", len(a), len(b))
	}
	result := make([]uint64, len(a))
	for i := range a {
		result[i] = a[i] & b[i]
	}
	return result, nil
}

// AllTrueMask returns a mask of n elements where every bit is 0 (true/selected).
//
// Mask convention: bit=0 means true/selected, bit=1 means false/unselected.
// The returned slice has length (n+63)/64 uint64 words. If n <= 0, an empty
// slice is returned.
//
// Note: (n+63)/64 will overflow if n > math.MaxInt-63, which is an extreme
// edge case not expected in practice.
//
// Parameters:
//   - n: The number of elements to represent in the mask.
//
// Returns:
//   - []uint64: A zero-initialized mask (all bits 0 = all true).
func AllTrueMask(n int) []uint64 {
	if n <= 0 {
		return []uint64{}
	}
	return make([]uint64, (n+63)/64)
}

// AllFalseMask returns a mask of n elements where every bit is 1 (false/unselected).
//
// Mask convention: bit=0 means true/selected, bit=1 means false/unselected.
// The returned slice has length (n+63)/64 uint64 words. If n <= 0, an empty
// slice is returned.
//
// Note: (n+63)/64 will overflow if n > math.MaxInt-63, which is an extreme
// edge case not expected in practice.
//
// Parameters:
//   - n: The number of elements to represent in the mask.
//
// Returns:
//   - []uint64: A mask with all bits set to 1 = all false.
func AllFalseMask(n int) []uint64 {
	if n <= 0 {
		return []uint64{}
	}
	mask := make([]uint64, (n+63)/64)
	for i := range mask {
		mask[i] = ^uint64(0)
	}
	return mask
}

// TensorNotMask inverts a boolean mask using logical NOT and returns a new mask.
//
// Mask convention: bit=0 means true/selected, bit=1 means false/unselected.
// NOT inverts every bit: bitwise NOT (^a) maps 0→1 (T→F) and 1→0 (F→T).
//
// Parameters:
//   - mask: The mask slice to invert.
//
// Returns:
//   - []uint64: A new mask with all bits inverted. Returns nil if mask is nil.
func TensorNotMask(mask []uint64) []uint64 {
	if mask == nil {
		return nil
	}
	result := make([]uint64, len(mask))
	for i := range mask {
		result[i] = ^mask[i]
	}
	return result
}
