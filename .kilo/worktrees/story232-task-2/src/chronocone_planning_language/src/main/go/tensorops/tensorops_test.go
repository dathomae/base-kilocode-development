// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package tensorops

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/calculation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComposeIndex_1D verifies composeIndex for 1D shapes.
func TestComposeIndex_1D(t *testing.T) {
	shape := []int{5}
	assert.Equal(t, 0, composeIndex([]int{0}, shape))
	assert.Equal(t, 4, composeIndex([]int{4}, shape))
}

// TestComposeIndex_2D verifies composeIndex for 2D shapes in row-major order.
func TestComposeIndex_2D(t *testing.T) {
	shape := []int{3, 4}
	assert.Equal(t, 0, composeIndex([]int{0, 0}, shape))
	assert.Equal(t, 1, composeIndex([]int{0, 1}, shape))
	assert.Equal(t, 3, composeIndex([]int{0, 3}, shape))
	assert.Equal(t, 4, composeIndex([]int{1, 0}, shape))
	assert.Equal(t, 7, composeIndex([]int{1, 3}, shape))
	assert.Equal(t, 11, composeIndex([]int{2, 3}, shape))
}

// TestComposeIndex_3D verifies composeIndex for 3D shapes in row-major order.
func TestComposeIndex_3D(t *testing.T) {
	shape := []int{2, 3, 4}
	assert.Equal(t, 0, composeIndex([]int{0, 0, 0}, shape))
	assert.Equal(t, 1, composeIndex([]int{0, 0, 1}, shape))
	assert.Equal(t, 3, composeIndex([]int{0, 0, 3}, shape))
	assert.Equal(t, 4, composeIndex([]int{0, 1, 0}, shape))
	assert.Equal(t, 12, composeIndex([]int{1, 0, 0}, shape))
	assert.Equal(t, 23, composeIndex([]int{1, 2, 3}, shape))
}

// TestDecomposeIndex_1D verifies decomposeIndex for 1D shapes.
func TestDecomposeIndex_1D(t *testing.T) {
	shape := []int{5}
	assert.Equal(t, []int{0}, decomposeIndex(0, shape))
	assert.Equal(t, []int{4}, decomposeIndex(4, shape))
}

// TestDecomposeIndex_2D verifies decomposeIndex for 2D shapes in row-major order.
func TestDecomposeIndex_2D(t *testing.T) {
	shape := []int{3, 4}
	assert.Equal(t, []int{0, 0}, decomposeIndex(0, shape))
	assert.Equal(t, []int{0, 1}, decomposeIndex(1, shape))
	assert.Equal(t, []int{0, 3}, decomposeIndex(3, shape))
	assert.Equal(t, []int{1, 0}, decomposeIndex(4, shape))
	assert.Equal(t, []int{1, 3}, decomposeIndex(7, shape))
	assert.Equal(t, []int{2, 3}, decomposeIndex(11, shape))
}

// TestDecomposeIndex_3D verifies decomposeIndex for 3D shapes in row-major order.
func TestDecomposeIndex_3D(t *testing.T) {
	shape := []int{2, 3, 4}
	assert.Equal(t, []int{0, 0, 0}, decomposeIndex(0, shape))
	assert.Equal(t, []int{0, 0, 1}, decomposeIndex(1, shape))
	assert.Equal(t, []int{0, 0, 3}, decomposeIndex(3, shape))
	assert.Equal(t, []int{0, 1, 0}, decomposeIndex(4, shape))
	assert.Equal(t, []int{1, 0, 0}, decomposeIndex(12, shape))
	assert.Equal(t, []int{1, 2, 3}, decomposeIndex(23, shape))
}

// TestRoundTrip_1D verifies that decomposeIndex followed by composeIndex is identity for 1D.
func TestRoundTrip_1D(t *testing.T) {
	shape := []int{5}
	for i := 0; i < 5; i++ {
		coords := decomposeIndex(i, shape)
		flat := composeIndex(coords, shape)
		assert.Equal(t, i, flat, "round-trip failed for index %d", i)
	}
}

// TestRoundTrip_2D verifies that decomposeIndex followed by composeIndex is identity for 2D.
func TestRoundTrip_2D(t *testing.T) {
	shape := []int{3, 4}
	total := 12
	for i := 0; i < total; i++ {
		coords := decomposeIndex(i, shape)
		flat := composeIndex(coords, shape)
		assert.Equal(t, i, flat, "round-trip failed for index %d", i)
	}
}

// TestRoundTrip_3D verifies that decomposeIndex followed by composeIndex is identity for 3D.
func TestRoundTrip_3D(t *testing.T) {
	shape := []int{2, 3, 4}
	total := 24
	for i := 0; i < total; i++ {
		coords := decomposeIndex(i, shape)
		flat := composeIndex(coords, shape)
		assert.Equal(t, i, flat, "round-trip failed for index %d", i)
	}
}

// TestRoundTrip_ComposeThenDecompose verifies that composeIndex followed by decomposeIndex is identity.
func TestRoundTrip_ComposeThenDecompose(t *testing.T) {
	shape := []int{2, 3, 4}
	total := 24
	for i := 0; i < total; i++ {
		coords := decomposeIndex(i, shape)
		flat := composeIndex(coords, shape)
		coordsBack := decomposeIndex(flat, shape)
		assert.Equal(t, coords, coordsBack, "compose→decompose round-trip failed for index %d", i)
	}
}

// TestBoundaryValues_LastIndex verifies last valid index for various shapes.
func TestBoundaryValues_LastIndex(t *testing.T) {
	cases := []struct {
		shape []int
		last  int
	}{
		{[]int{1}, 0},
		{[]int{5}, 4},
		{[]int{3, 4}, 11},
		{[]int{2, 3, 4}, 23},
		{[]int{1, 1, 1, 1}, 0},
	}
	for _, tc := range cases {
		coords := decomposeIndex(tc.last, tc.shape)
		flat := composeIndex(coords, tc.shape)
		require.Equal(t, tc.last, flat, "last index round-trip failed for shape %v", tc.shape)
	}
}

// TestComposeIndex_EmptyShape verifies composeIndex with an empty shape returns 0.
func TestComposeIndex_EmptyShape(t *testing.T) {
	assert.Equal(t, 0, composeIndex([]int{}, []int{}))
	assert.Equal(t, 0, composeIndex([]int{1, 2, 3}, []int{}))
}

// TestDecomposeIndex_EmptyShape verifies decomposeIndex with an empty shape returns an empty slice.
func TestDecomposeIndex_EmptyShape(t *testing.T) {
	assert.Equal(t, []int{}, decomposeIndex(0, []int{}))
	assert.Equal(t, []int{}, decomposeIndex(5, []int{}))
}

// TestComposeIndex_MismatchedLengths verifies that extra coords are ignored and
// missing coords are treated as 0.
func TestComposeIndex_MismatchedLengths(t *testing.T) {
	shape := []int{2, 3}
	// Extra coords beyond len(shape) are ignored.
	assert.Equal(t, 5, composeIndex([]int{1, 2, 99}, shape))

	shape2 := []int{2, 3, 4}
	// Missing coords are treated as 0.
	assert.Equal(t, 20, composeIndex([]int{1, 2}, shape2))
}

// TestDecomposeIndex_ReturnedLength verifies the returned slice always has
// len == len(shape).
func TestDecomposeIndex_ReturnedLength(t *testing.T) {
	for _, shape := range [][]int{
		{},
		{5},
		{3, 4},
		{2, 3, 4},
	} {
		coords := decomposeIndex(0, shape)
		assert.Len(t, coords, len(shape), "shape %v", shape)
	}
}

// TestDecomposeIndex_ZeroSizeDimension verifies that a zero-size dimension
// causes a panic (division by zero) because callers must pre-validate shapes.
func TestDecomposeIndex_ZeroSizeDimension(t *testing.T) {
	assert.Panics(t, func() {
		decomposeIndex(0, []int{2, 0})
	})
}

// TestComposeIndex_NegativeCoords verifies that negative coordinates are not
// validated and produce negative or incorrect indices.
func TestComposeIndex_NegativeCoords(t *testing.T) {
	shape := []int{3, 4}
	assert.Equal(t, -4, composeIndex([]int{-1, 0}, shape))
	assert.Equal(t, -1, composeIndex([]int{0, -1}, shape))
}

// TestDecomposeIndex_NegativeIndex verifies that a negative index is not
// validated and produces coordinates based on Go's truncating integer division.
func TestDecomposeIndex_NegativeIndex(t *testing.T) {
	shape := []int{3, 4}
	assert.Equal(t, []int{0, -1}, decomposeIndex(-1, shape))
}

// TestComposeIndex_OutOfBounds verifies that out-of-bounds coordinates are not
// validated and produce indices outside the tensor size.
func TestComposeIndex_OutOfBounds(t *testing.T) {
	shape := []int{2, 3}
	assert.Equal(t, 20, composeIndex([]int{5, 5}, shape))
}

// TestDecomposeIndex_OutOfBounds verifies that out-of-bounds indices are not
// validated and produce coordinates outside the shape bounds.
func TestDecomposeIndex_OutOfBounds(t *testing.T) {
	shape := []int{2, 3}
	assert.Equal(t, []int{3, 1}, decomposeIndex(10, shape))
}

// ---------------------------------------------------------------------------
// ReduceDimension tests
// ---------------------------------------------------------------------------

// TestReduceDimension_1DTo1D verifies reduction of a 1D tensor along its
// sole dimension using a 2×1 weight matrix.  The result remains 1D (shape [1]).
func TestReduceDimension_1DTo1D(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{10.0, 20.0},
		[]string{"x"},
		[]int{2},
	)
	matrix := &AggregationMatrix{
		Rows:    2,
		Columns: 1,
		Data:    [][]float64{{1.0}, {1.0}},
	}

	result, err := ReduceDimension(src, matrix, 0, 0)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, []float64{30.0}, result.Data())
	assert.Equal(t, []int{1}, result.Shape())
	assert.Equal(t, []string{"x"}, result.DimensionNames())
}

// TestReduceDimension_2DTo2D verifies reduction of a 2D tensor along
// dimension 1 using a 3×2 weight matrix.  The collapsed dimension (size 3)
// is replaced by matrix columns (size 2), producing a [2,2] result.
func TestReduceDimension_2DTo2D(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0},
		[]string{"row", "col"},
		[]int{2, 3},
	)
	matrix := &AggregationMatrix{
		Rows:    3,
		Columns: 2,
		Data: [][]float64{
			{1.0, 0.0},
			{0.0, 1.0},
			{1.0, 0.0},
		},
	}

	result, err := ReduceDimension(src, matrix, 1, 1)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, []int{2, 2}, result.Shape())
	assert.Equal(t, []string{"row", "col"}, result.DimensionNames())
	assert.InDeltaSlice(t, []float64{4.0, 2.0, 10.0, 5.0}, result.Data(), 1e-9)
}

// TestReduceDimension_3DTo3D verifies reduction of a 3D tensor along
// dimension 1 using a 3×1 weight matrix.  The collapsed dimension is removed
// and the single matrix column is placed at targetDim=2, producing [2,2,1].
func TestReduceDimension_3DTo3D(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		[]string{"a", "b", "c"},
		[]int{2, 3, 2},
	)
	matrix := &AggregationMatrix{
		Rows:    3,
		Columns: 1,
		Data: [][]float64{
			{1.0},
			{2.0},
			{3.0},
		},
	}

	result, err := ReduceDimension(src, matrix, 1, 2)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, []int{2, 2, 1}, result.Shape())
	assert.Equal(t, []string{"a", "c", "b"}, result.DimensionNames())
	assert.InDeltaSlice(t, []float64{22.0, 28.0, 58.0, 64.0}, result.Data(), 1e-9)
}

// TestReduceDimension_TargetDimLessThanCollapsedDim verifies reduction when
// targetDim < collapsedDim (3D tensor shape=[2,3,2], collapsedDim=2,
// targetDim=0) using a 2×1 weight matrix.
func TestReduceDimension_TargetDimLessThanCollapsedDim(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		[]string{"a", "b", "c"},
		[]int{2, 3, 2},
	)
	matrix := &AggregationMatrix{
		Rows:    2,
		Columns: 1,
		Data: [][]float64{
			{1.0},
			{1.0},
		},
	}

	result, err := ReduceDimension(src, matrix, 2, 0)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, []int{1, 2, 3}, result.Shape())
	assert.Equal(t, []string{"c", "a", "b"}, result.DimensionNames())
	assert.InDeltaSlice(t, []float64{3, 7, 11, 15, 19, 23}, result.Data(), 1e-9)
}

// TestReduceDimension_TargetDimEqualsCollapsedDim verifies reduction when
// targetDim == collapsedDim (2D tensor shape=[2,3], collapsedDim=1,
// targetDim=1) using a 3×2 weight matrix.
func TestReduceDimension_TargetDimEqualsCollapsedDim(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0},
		[]string{"row", "col"},
		[]int{2, 3},
	)
	matrix := &AggregationMatrix{
		Rows:    3,
		Columns: 2,
		Data: [][]float64{
			{1.0, 0.0},
			{0.0, 1.0},
			{1.0, 0.0},
		},
	}

	result, err := ReduceDimension(src, matrix, 1, 1)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, []int{2, 2}, result.Shape())
	assert.Equal(t, []string{"row", "col"}, result.DimensionNames())
	assert.InDeltaSlice(t, []float64{4.0, 2.0, 10.0, 5.0}, result.Data(), 1e-9)
}

// TestReduceDimension_ErrorDimOutOfRange verifies that collapsedDim >= N
// returns an error.
func TestReduceDimension_ErrorDimOutOfRange(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0},
		[]string{"x"},
		[]int{2},
	)
	matrix := &AggregationMatrix{
		Rows:    2,
		Columns: 1,
		Data:    [][]float64{{1.0}, {1.0}},
	}

	// collapsedDim == N
	_, err := ReduceDimension(src, matrix, 1, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collapsedDim")

	// collapsedDim > N
	_, err = ReduceDimension(src, matrix, 2, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collapsedDim")

	// collapsedDim < 0
	_, err = ReduceDimension(src, matrix, -1, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collapsedDim")
}

// TestReduceDimension_ErrorMatrixRowMismatch verifies that
// matrix.Rows != src.Shape()[collapsedDim] returns an error.
func TestReduceDimension_ErrorMatrixRowMismatch(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0},
		[]string{"row", "col"},
		[]int{2, 3},
	)
	matrix := &AggregationMatrix{
		Rows:    99,
		Columns: 2,
		Data:    [][]float64{{1.0, 0.0}},
	}

	_, err := ReduceDimension(src, matrix, 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "matrix rows")
}

// TestReduceDimension_ErrorTargetDimOutOfRange verifies that targetDim out of
// range returns an error.
func TestReduceDimension_ErrorTargetDimOutOfRange(t *testing.T) {
	// Use a tensor where matrix.Rows matches so we actually hit the targetDim check.
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0},
		[]string{"row", "col"},
		[]int{2, 3},
	)
	matrix := &AggregationMatrix{
		Rows:    2,
		Columns: 1,
		Data:    [][]float64{{1.0}, {1.0}},
	}

	// Collapse dim 0, N=2. Valid targetDim=0,1. targetDim=2 is out of range.
	_, err := ReduceDimension(src, matrix, 0, 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "targetDim")

	// targetDim < 0
	_, err = ReduceDimension(src, matrix, 0, -1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "targetDim")
}

// TestReduceDimension_NilTensorAndMatrix verifies nil handling.
func TestReduceDimension_NilTensorAndMatrix(t *testing.T) {
	_, err := ReduceDimension(nil, nil, 0, 0)
	require.Error(t, err)

	src := calculation.NewChronoconeTensor(
		[]float64{1.0},
		[]string{"x"},
		[]int{1},
	)
	_, err = ReduceDimension(src, nil, 0, 0)
	require.Error(t, err)

	matrix := &AggregationMatrix{
		Rows:    1,
		Columns: 1,
		Data:    [][]float64{{1.0}},
	}
	_, err = ReduceDimension(nil, matrix, 0, 0)
	require.Error(t, err)
}

// TestReduceDimension_ErrorMatrixDataLengthMismatch verifies that
// len(matrix.Data) != matrix.Rows returns an error.
func TestReduceDimension_ErrorMatrixDataLengthMismatch(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0, 3.0},
		[]string{"x"},
		[]int{3},
	)
	matrix := &AggregationMatrix{
		Rows:    3,
		Columns: 1,
		Data:    [][]float64{{1.0}}, // only 1 row, not 3
	}

	_, err := ReduceDimension(src, matrix, 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "len(matrix.Data)")
}

// TestReduceDimension_ErrorMatrixRowColumnMismatch verifies that a row
// in matrix.Data with length != matrix.Columns returns an error.
func TestReduceDimension_ErrorMatrixRowColumnMismatch(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0},
		[]string{"x"},
		[]int{2},
	)
	matrix := &AggregationMatrix{
		Rows:    2,
		Columns: 2,
		Data: [][]float64{
			{1.0, 2.0},
			{3.0}, // second row has 1 column, not 2
		},
	}

	_, err := ReduceDimension(src, matrix, 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "matrix.Data[1]")
}

// ---------------------------------------------------------------------------
// ExpandDimension tests
// ---------------------------------------------------------------------------

// TestExpandDimension_ScalarTo1D verifies expansion of a scalar tensor
// (0 dims, 1 element) via a 1×3 matrix → 3-element 1D tensor.
func TestExpandDimension_ScalarTo1D(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{5.0},
		[]string{},
		[]int{},
	)
	matrix := &AggregationMatrix{
		Rows:    1,
		Columns: 3,
		Data:    [][]float64{{1.0, 2.0, 3.0}},
	}

	result, err := ExpandDimension(src, matrix, 0, 0)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, []int{3}, result.Shape())
	assert.InDeltaSlice(t, []float64{5.0, 10.0, 15.0}, result.Data(), 1e-9)
}

// TestExpandDimension_1DTo2D verifies expansion of a 1D tensor (shape [2])
// via a 2×3 matrix → 2D result shape [2,3].
func TestExpandDimension_1DTo2D(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0},
		[]string{"x"},
		[]int{2},
	)
	matrix := &AggregationMatrix{
		Rows:    2,
		Columns: 3,
		Data: [][]float64{
			{1.0, 2.0, 3.0},
			{4.0, 5.0, 6.0},
		},
	}

	result, err := ExpandDimension(src, matrix, 0, 1)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, []int{2, 3}, result.Shape())
	assert.Equal(t, []string{"x", ""}, result.DimensionNames())
	assert.InDeltaSlice(t, []float64{1.0, 2.0, 3.0, 8.0, 10.0, 12.0}, result.Data(), 1e-9)
}

// TestExpandDimension_2DTo3D verifies expansion of a 2D tensor via a matrix
// → 3D result.
func TestExpandDimension_2DTo3D(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0},
		[]string{"a", "b"},
		[]int{2, 3},
	)
	matrix := &AggregationMatrix{
		Rows:    2,
		Columns: 2,
		Data: [][]float64{
			{1.0, 2.0},
			{3.0, 4.0},
		},
	}

	result, err := ExpandDimension(src, matrix, 0, 1)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, []int{2, 2, 3}, result.Shape())
	assert.Equal(t, []string{"a", "", "b"}, result.DimensionNames())
	assert.InDeltaSlice(t, []float64{1.0, 2.0, 3.0, 2.0, 4.0, 6.0, 12.0, 15.0, 18.0, 16.0, 20.0, 24.0}, result.Data(), 1e-9)
}

// TestExpandDimension_NewDimEqualsSrcDim verifies expansion when
// newDim == srcDim (2D tensor shape=[2,3], srcDim=1, newDim=1) using a
// 3×2 weight matrix.  Result shape should be [2,2,3].
func TestExpandDimension_NewDimEqualsSrcDim(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0},
		[]string{"a", "b"},
		[]int{2, 3},
	)
	matrix := &AggregationMatrix{
		Rows:    3,
		Columns: 2,
		Data: [][]float64{
			{1.0, 2.0},
			{3.0, 4.0},
			{5.0, 6.0},
		},
	}

	result, err := ExpandDimension(src, matrix, 1, 1)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, []int{2, 2, 3}, result.Shape())
	assert.Equal(t, []string{"a", "", "b"}, result.DimensionNames())
	assert.InDeltaSlice(t, []float64{1.0, 6.0, 15.0, 2.0, 8.0, 18.0, 4.0, 15.0, 30.0, 8.0, 20.0, 36.0}, result.Data(), 1e-9)
}

// TestExpandDimension_NewDimLessThanSrcDim verifies expansion when
// newDim < srcDim (2D tensor shape=[2,3], srcDim=1, newDim=0) using a
// 3×2 weight matrix.  Result shape should be [2,2,3].
func TestExpandDimension_NewDimLessThanSrcDim(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0},
		[]string{"a", "b"},
		[]int{2, 3},
	)
	matrix := &AggregationMatrix{
		Rows:    3,
		Columns: 2,
		Data: [][]float64{
			{1.0, 2.0},
			{3.0, 4.0},
			{5.0, 6.0},
		},
	}

	result, err := ExpandDimension(src, matrix, 1, 0)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, []int{2, 2, 3}, result.Shape())
	assert.Equal(t, []string{"", "a", "b"}, result.DimensionNames())
	assert.InDeltaSlice(t, []float64{1.0, 6.0, 15.0, 4.0, 15.0, 30.0, 2.0, 8.0, 18.0, 8.0, 20.0, 36.0}, result.Data(), 1e-9)
}

// TestExpandDimension_ErrorDimOutOfRange verifies that srcDim >= N or
// newDim > N returns an error.
func TestExpandDimension_ErrorDimOutOfRange(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0},
		[]string{"x"},
		[]int{2},
	)
	matrix := &AggregationMatrix{
		Rows:    2,
		Columns: 3,
		Data:    [][]float64{{1.0, 2.0, 3.0}, {4.0, 5.0, 6.0}},
	}

	// srcDim == N
	_, err := ExpandDimension(src, matrix, 1, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "srcDim")

	// srcDim > N
	_, err = ExpandDimension(src, matrix, 2, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "srcDim")

	// srcDim < 0
	_, err = ExpandDimension(src, matrix, -1, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "srcDim")

	// newDim > N (N=1, valid newDim is 0 or 1)
	_, err = ExpandDimension(src, matrix, 0, 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "newDim")

	// newDim < 0
	_, err = ExpandDimension(src, matrix, 0, -1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "newDim")
}

// TestExpandDimension_ErrorScalarSrcDimOutOfRange verifies that a scalar
// tensor (N=0) rejects srcDim != 0.
func TestExpandDimension_ErrorScalarSrcDimOutOfRange(t *testing.T) {
	scalar := calculation.NewChronoconeTensor(
		[]float64{5.0},
		[]string{},
		[]int{},
	)
	matrix := &AggregationMatrix{
		Rows:    1,
		Columns: 3,
		Data:    [][]float64{{1.0, 2.0, 3.0}},
	}

	_, err := ExpandDimension(scalar, matrix, 1, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "srcDim")
}

// TestExpandDimension_ErrorMatrixRowMismatch verifies that
// matrix.Rows != src.Shape()[srcDim] returns an error.
func TestExpandDimension_ErrorMatrixRowMismatch(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0},
		[]string{"x"},
		[]int{2},
	)
	matrix := &AggregationMatrix{
		Rows:    99,
		Columns: 3,
		Data:    [][]float64{{1.0, 2.0, 3.0}},
	}

	_, err := ExpandDimension(src, matrix, 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "matrix rows")
}

// TestExpandDimension_NilTensorAndMatrix verifies nil handling.
func TestExpandDimension_NilTensorAndMatrix(t *testing.T) {
	_, err := ExpandDimension(nil, nil, 0, 0)
	require.Error(t, err)

	src := calculation.NewChronoconeTensor(
		[]float64{1.0},
		[]string{"x"},
		[]int{1},
	)
	_, err = ExpandDimension(src, nil, 0, 0)
	require.Error(t, err)

	matrix := &AggregationMatrix{
		Rows:    1,
		Columns: 1,
		Data:    [][]float64{{1.0}},
	}
	_, err = ExpandDimension(nil, matrix, 0, 0)
	require.Error(t, err)
}

// TestExpandDimension_ErrorMatrixDataLengthMismatch verifies that
// len(matrix.Data) != matrix.Rows returns an error.
func TestExpandDimension_ErrorMatrixDataLengthMismatch(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0, 3.0},
		[]string{"x"},
		[]int{3},
	)
	matrix := &AggregationMatrix{
		Rows:    3,
		Columns: 1,
		Data:    [][]float64{{1.0}}, // only 1 row, not 3
	}

	_, err := ExpandDimension(src, matrix, 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "len(matrix.Data)")
}

// TestExpandDimension_ErrorMatrixRowColumnMismatch verifies that a row
// in matrix.Data with length != matrix.Columns returns an error.
func TestExpandDimension_ErrorMatrixRowColumnMismatch(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0},
		[]string{"x"},
		[]int{2},
	)
	matrix := &AggregationMatrix{
		Rows:    2,
		Columns: 2,
		Data: [][]float64{
			{1.0, 2.0},
			{3.0}, // second row has 1 column, not 2
		},
	}

	_, err := ExpandDimension(src, matrix, 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "matrix.Data[1]")
}

// ---------------------------------------------------------------------------
// Edge case tests
// ---------------------------------------------------------------------------

// TestReduceDimension_IdentityMatrix verifies that an identity matrix
// produces a no-op when matrix.Columns == shape[collapsedDim].
func TestReduceDimension_IdentityMatrix(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0},
		[]string{"row", "col"},
		[]int{2, 3},
	)
	matrix := &AggregationMatrix{
		Rows:    3,
		Columns: 3,
		Data: [][]float64{
			{1.0, 0.0, 0.0},
			{0.0, 1.0, 0.0},
			{0.0, 0.0, 1.0},
		},
	}

	result, err := ReduceDimension(src, matrix, 1, 1)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, []int{2, 3}, result.Shape())
	assert.Equal(t, []string{"row", "col"}, result.DimensionNames())
	assert.InDeltaSlice(t, []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}, result.Data(), 1e-9)
}

// TestExpandDimension_IdentityMatrix verifies expansion with an identity
// matrix.  The source values are projected along the new dimension.
func TestExpandDimension_IdentityMatrix(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0},
		[]string{"row", "col"},
		[]int{2, 3},
	)
	matrix := &AggregationMatrix{
		Rows:    3,
		Columns: 3,
		Data: [][]float64{
			{1.0, 0.0, 0.0},
			{0.0, 1.0, 0.0},
			{0.0, 0.0, 1.0},
		},
	}

	result, err := ExpandDimension(src, matrix, 1, 1)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, []int{2, 3, 3}, result.Shape())
	assert.Equal(t, []string{"row", "", "col"}, result.DimensionNames())
	// For each src[i,j], the value appears only at result[i,j,j].
	expected := []float64{
		1, 0, 0, 0, 2, 0, 0, 0, 3,
		4, 0, 0, 0, 5, 0, 0, 0, 6,
	}
	assert.InDeltaSlice(t, expected, result.Data(), 1e-9)
}

// TestReduceDimension_SingleValueDimensions verifies reduction on a tensor
// where every dimension has size 1.
func TestReduceDimension_SingleValueDimensions(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{42.0},
		[]string{"a", "b", "c"},
		[]int{1, 1, 1},
	)
	matrix := &AggregationMatrix{
		Rows:    1,
		Columns: 1,
		Data:    [][]float64{{1.0}},
	}

	result, err := ReduceDimension(src, matrix, 1, 0)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, []int{1, 1, 1}, result.Shape())
	assert.Equal(t, []string{"b", "a", "c"}, result.DimensionNames())
	assert.InDeltaSlice(t, []float64{42.0}, result.Data(), 1e-9)
}

// TestExpandDimension_SingleValueDimensions verifies expansion on a tensor
// where every dimension has size 1.
func TestExpandDimension_SingleValueDimensions(t *testing.T) {
	src := calculation.NewChronoconeTensor(
		[]float64{42.0},
		[]string{"a", "b"},
		[]int{1, 1},
	)
	matrix := &AggregationMatrix{
		Rows:    1,
		Columns: 1,
		Data:    [][]float64{{1.0}},
	}

	result, err := ExpandDimension(src, matrix, 0, 1)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, []int{1, 1, 1}, result.Shape())
	assert.Equal(t, []string{"a", "", "b"}, result.DimensionNames())
	assert.InDeltaSlice(t, []float64{42.0}, result.Data(), 1e-9)
}

// TestReduceDimension_ScalarError verifies that reducing a rank-zero
// (scalar) tensor is rejected because collapsedDim is out of range.
func TestReduceDimension_ScalarError(t *testing.T) {
	scalar := calculation.NewChronoconeTensor(
		[]float64{5.0},
		[]string{},
		[]int{},
	)
	matrix := &AggregationMatrix{
		Rows:    1,
		Columns: 1,
		Data:    [][]float64{{1.0}},
	}

	_, err := ReduceDimension(scalar, matrix, 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collapsedDim")
}

// TestExpandDimension_ScalarTo1D_Identity verifies scalar expansion with a
// 1×1 identity matrix, producing a 1-element 1D tensor whose value equals
// the source scalar.
func TestExpandDimension_ScalarTo1D_Identity(t *testing.T) {
	scalar := calculation.NewChronoconeTensor(
		[]float64{5.0},
		[]string{},
		[]int{},
	)
	matrix := &AggregationMatrix{
		Rows:    1,
		Columns: 1,
		Data:    [][]float64{{1.0}},
	}

	result, err := ExpandDimension(scalar, matrix, 0, 0)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, []int{1}, result.Shape())
	assert.InDeltaSlice(t, []float64{5.0}, result.Data(), 1e-9)
}
