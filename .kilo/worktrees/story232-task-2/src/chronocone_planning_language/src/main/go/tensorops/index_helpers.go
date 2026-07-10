// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package tensorops

// composeIndex converts N-dimensional coordinates to a flat row-major index.
// The last dimension varies fastest. For example, with shape [2, 3]:
//
//	[0,0] → 0, [0,1] → 1, [0,2] → 2
//	[1,0] → 3, [1,1] → 4, [1,2] → 5
//
// When len(coords) != len(shape), extra coords are ignored and missing coords
// are treated as 0. Callers are responsible for ensuring lengths match if they
// require strict behavior.
//
// Preconditions: all coordinates must be non-negative and within the bounds of
// their respective shape dimensions. Callers must perform validation; these
// functions are designed for performance in inner loops and do not validate
// inputs. All shape dimensions must be > 0.
func composeIndex(coords []int, shape []int) int {
	if len(shape) == 0 {
		return 0
	}

	strides := computeStrides(shape)
	index := 0
	for i := 0; i < len(coords) && i < len(shape); i++ {
		index += coords[i] * strides[i]
	}
	return index
}

// decomposeIndex converts a flat row-major index to N-dimensional coordinates.
// The last dimension varies fastest. For example, with shape [2, 3]:
//
//	0 → [0,0], 1 → [0,1], 2 → [0,2]
//	3 → [1,0], 4 → [1,1], 5 → [1,2]
//
// The returned slice always has len == len(shape).
//
// Preconditions: index must be non-negative and within bounds. Callers must
// perform validation; these functions are designed for performance in inner
// loops and do not validate inputs. All shape dimensions must be > 0; callers
// must pre-validate shapes to avoid division by zero.
func decomposeIndex(index int, shape []int) []int {
	if len(shape) == 0 {
		return []int{}
	}

	strides := computeStrides(shape)
	coords := make([]int, len(shape))
	remaining := index
	for i := 0; i < len(shape); i++ {
		coords[i] = remaining / strides[i]
		remaining = remaining % strides[i]
	}
	return coords
}

// computeStrides calculates the stride for each dimension in row-major order.
// stride[i] is the number of elements to skip when moving by one in dimension i.
//
// All shape dimensions must be > 0.
func computeStrides(shape []int) []int {
	strides := make([]int, len(shape))
	stride := 1
	for i := len(shape) - 1; i >= 0; i-- {
		strides[i] = stride
		stride *= shape[i]
	}
	return strides
}
