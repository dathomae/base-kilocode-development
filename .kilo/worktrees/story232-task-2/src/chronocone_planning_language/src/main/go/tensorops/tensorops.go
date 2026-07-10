// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

// Package tensorops provides standalone tensor matrix multiplication
// utilities for dimension-reducing and dimension-expanding operations.
// It has no dependencies on CPL parsing or CIL synthesis.
package tensorops

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/calculation"
)

// AggregationMatrix is a 2D weight matrix used for tensor dimension
// reduction (ReduceDimension) and expansion (ExpandDimension).
// Rows correspond to the source dimension's size and Columns to the result.
type AggregationMatrix struct {
	Rows      int
	Columns   int
	Data      [][]float64
	RowLabels []string
	ColLabels []string
}

// ReduceDimension contracts an N-dimensional tensor along collapsedDim using a
// weight matrix.  The collapsed dimension is removed and matrix.Columns is
// inserted at targetDim.  The result always has N dimensions.
//
// Parameters:
//   - src:         Source tensor of shape [d₀,…,d_{N-1}].
//   - matrix:      Weight matrix with Rows == src.Shape()[collapsedDim].
//   - collapsedDim: Which dimension to collapse (0-based).
//   - targetDim:   Where in the remaining N-1 positions to insert matrix.Columns.
//                   Must be in [0, N-1].
//
// The computation for each result cell is:
//
//	result[R] = Σ_{k}  src[S(R,k)] × matrix[k][j]
//
// where S(R,k) is R with the targetDim element removed and k inserted at
// collapsedDim, and j = R[targetDim].
func ReduceDimension(
	src *calculation.ChronoconeTensor[float64],
	matrix *AggregationMatrix,
	collapsedDim int,
	targetDim int,
) (*calculation.ChronoconeTensor[float64], error) {
	// --- nil guards -----------------------------------------------------------
	if src == nil {
		return nil, fmt.Errorf("ReduceDimension: src tensor is nil")
	}
	if matrix == nil {
		return nil, fmt.Errorf("ReduceDimension: matrix is nil")
	}

	srcShape := src.Shape()
	N := len(srcShape)

	// --- validate collapsedDim ------------------------------------------------
	if collapsedDim < 0 || collapsedDim >= N {
		return nil, fmt.Errorf(
			"ReduceDimension: collapsedDim %d out of range [0, %d)",
			collapsedDim, N,
		)
	}

	// --- validate matrix row count --------------------------------------------
	collapsedSize := srcShape[collapsedDim]
	if matrix.Rows != collapsedSize {
		return nil, fmt.Errorf(
			"ReduceDimension: matrix rows %d != src.Shape()[%d] = %d",
			matrix.Rows, collapsedDim, collapsedSize,
		)
	}

	// --- validate matrix data bounds ------------------------------------------
	if len(matrix.Data) != matrix.Rows {
		return nil, fmt.Errorf(
			"ReduceDimension: len(matrix.Data) %d != matrix.Rows %d",
			len(matrix.Data), matrix.Rows,
		)
	}
	for k := 0; k < matrix.Rows; k++ {
		if len(matrix.Data[k]) != matrix.Columns {
			return nil, fmt.Errorf(
				"ReduceDimension: len(matrix.Data[%d]) %d != matrix.Columns %d",
				k, len(matrix.Data[k]), matrix.Columns,
			)
		}
	}

	// --- validate targetDim ---------------------------------------------------
	// After removing collapsedDim, N-1 positions remain.  targetDim chooses
	// among N possible insertion points (0 … N-1).
	if targetDim < 0 || targetDim >= N {
		return nil, fmt.Errorf(
			"ReduceDimension: targetDim %d out of range [0, %d]",
			targetDim, N,
		)
	}

	// --- build result shape and dimension names -------------------------------
	srcNames := src.DimensionNames()
	resultShape := make([]int, 0, N)
	resultNames := make([]string, 0, N)

	for i := 0; i < N; i++ {
		if i == collapsedDim {
			continue
		}
		resultShape = append(resultShape, srcShape[i])
		resultNames = append(resultNames, srcNames[i])
	}

	// Insert the matrix.Columns (and the collapsed dimension's name) at
	// targetDim.
	resultShape = insertInt(resultShape, targetDim, matrix.Columns)
	resultNames = insertString(resultNames, targetDim, srcNames[collapsedDim])

	// --- allocate result data -------------------------------------------------
	resultSize := 1
	for _, s := range resultShape {
		resultSize *= s
	}
	resultData := make([]float64, resultSize)

	// --- compute result -------------------------------------------------------
	for resultIdx := 0; resultIdx < resultSize; resultIdx++ {
		resultCoords := decomposeIndex(resultIdx, resultShape)
		j := resultCoords[targetDim] // matrix column index

		for k := 0; k < matrix.Rows; k++ {
			srcCoords := makeSourceCoords(resultCoords, targetDim, collapsedDim, k)
			srcIdx := composeIndex(srcCoords, srcShape)
			resultData[resultIdx] += src.Data()[srcIdx] * matrix.Data[k][j]
		}
	}

	return calculation.NewChronoconeTensor(resultData, resultNames, resultShape), nil
}

// makeSourceCoords builds N-dimensional source coordinates from N-dimensional
// result coordinates.  The element at resultCoords[targetDim] (the matrix
// column index) is removed, and k is inserted at collapsedDim.
func makeSourceCoords(resultCoords []int, targetDim, collapsedDim, k int) []int {
	N := len(resultCoords)
	srcCoords := make([]int, N)

	for d := 0; d < N; d++ {
		if d == collapsedDim {
			srcCoords[d] = k
			continue
		}
		// Map source dimension d to its position in resultCoords.
		// Source dims ≠ collapsedDim appear in result, shifted by removal of
		// collapsedDim and by insertion of matrix columns at targetDim.
		pos := d
		if d > collapsedDim {
			pos = d - 1 // account for removal of collapsedDim
		}
		if pos >= targetDim {
			pos++ // account for insertion at targetDim
		}
		srcCoords[d] = resultCoords[pos]
	}
	return srcCoords
}

// ExpandDimension expands an N-dimensional tensor to (N+1) dimensions by
// distributing source values through a weight matrix.
//
// Parameters:
//   - src: Source tensor of shape [d₀,…,d_{N-1}]
//   - matrix: Weight matrix with Rows == src.Shape()[srcDim]
//   - srcDim: Index of the source dimension to expand from (0-based)
//   - newDim: Index where the new dimension is inserted (0-based, in [0, N])
//
// Each result cell: result[...j...] = src[...i...] × matrix[i][j]
func ExpandDimension(
	src *calculation.ChronoconeTensor[float64],
	matrix *AggregationMatrix,
	srcDim int,
	newDim int,
) (*calculation.ChronoconeTensor[float64], error) {
	// --- nil guards -----------------------------------------------------------
	if src == nil {
		return nil, fmt.Errorf("ExpandDimension: src tensor is nil")
	}
	if matrix == nil {
		return nil, fmt.Errorf("ExpandDimension: matrix is nil")
	}

	srcShape := src.Shape()
	N := len(srcShape)

	// --- validate srcDim ------------------------------------------------------
	if N == 0 {
		if srcDim != 0 {
			return nil, fmt.Errorf(
				"ExpandDimension: srcDim %d out of range [0, %d)",
				srcDim, N,
			)
		}
	} else {
		if srcDim < 0 || srcDim >= N {
			return nil, fmt.Errorf(
				"ExpandDimension: srcDim %d out of range [0, %d)",
				srcDim, N,
			)
		}
	}

	// --- validate matrix row count --------------------------------------------
	srcDimSize := 1
	if N > 0 {
		srcDimSize = srcShape[srcDim]
	}
	if matrix.Rows != srcDimSize {
		return nil, fmt.Errorf(
			"ExpandDimension: matrix rows %d != src.Shape()[%d] = %d",
			matrix.Rows, srcDim, srcDimSize,
		)
	}

	// --- validate matrix data bounds ------------------------------------------
	if len(matrix.Data) != matrix.Rows {
		return nil, fmt.Errorf(
			"ExpandDimension: len(matrix.Data) %d != matrix.Rows %d",
			len(matrix.Data), matrix.Rows,
		)
	}
	for k := 0; k < matrix.Rows; k++ {
		if len(matrix.Data[k]) != matrix.Columns {
			return nil, fmt.Errorf(
				"ExpandDimension: len(matrix.Data[%d]) %d != matrix.Columns %d",
				k, len(matrix.Data[k]), matrix.Columns,
			)
		}
	}

	// --- validate newDim ------------------------------------------------------
	if newDim < 0 || newDim > N {
		return nil, fmt.Errorf(
			"ExpandDimension: newDim %d out of range [0, %d]",
			newDim, N,
		)
	}

	// --- build result shape and dimension names -------------------------------
	srcNames := src.DimensionNames()
	resultShape := make([]int, 0, N+1)
	resultNames := make([]string, 0, N+1)

	resultShape = append(resultShape, srcShape...)
	resultNames = append(resultNames, srcNames...)

	resultShape = insertInt(resultShape, newDim, matrix.Columns)
	resultNames = insertString(resultNames, newDim, "")

	// --- allocate result data -------------------------------------------------
	resultSize := 1
	for _, s := range resultShape {
		resultSize *= s
	}
	resultData := make([]float64, resultSize)

	// --- compute result -------------------------------------------------------
	for resultIdx := 0; resultIdx < resultSize; resultIdx++ {
		resultCoords := decomposeIndex(resultIdx, resultShape)
		j := resultCoords[newDim] // matrix column index

		// Build source coords by removing the newDim element from result coords.
		srcCoords := make([]int, 0, N)
		for d := 0; d < len(resultCoords); d++ {
			if d == newDim {
				continue
			}
			srcCoords = append(srcCoords, resultCoords[d])
		}

		// Matrix row index from source coordinate at srcDim.
		i := 0
		if N > 0 {
			i = srcCoords[srcDim]
		}

		srcIdx := composeIndex(srcCoords, srcShape)
		resultData[resultIdx] = src.Data()[srcIdx] * matrix.Data[i][j]
	}

	return calculation.NewChronoconeTensor(resultData, resultNames, resultShape), nil
}

// insertInt inserts v at position i into slice s, returning a new slice.
func insertInt(s []int, i int, v int) []int {
	s = append(s, 0) // grow by one
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}

// insertString inserts v at position i into slice s, returning a new slice.
func insertString(s []string, i int, v string) []string {
	s = append(s, "") // grow by one
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}
