package analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSumAggregationMatrix_MultipleColumns(t *testing.T) {
	collapsed := []string{"USA", "Canada", "Mexico"}
	target := []string{"Widget", "Gadget"}

	matrix := BuildSumAggregationMatrix(collapsed, target)

	assert.Equal(t, 3, matrix.Rows)
	assert.Equal(t, 2, matrix.Columns)
	assert.Equal(t, collapsed, matrix.RowLabels)
	assert.Equal(t, target, matrix.ColLabels)

	// All entries should be 1.0
	for i := 0; i < matrix.Rows; i++ {
		assert.Len(t, matrix.Data[i], matrix.Columns)
		for j := 0; j < matrix.Columns; j++ {
			assert.InDelta(t, 1.0, matrix.Data[i][j], 1e-9, "entry [%d][%d] should be 1.0", i, j)
		}
	}
}

func TestBuildSumAggregationMatrix_SingleColumn(t *testing.T) {
	collapsed := []string{"USA", "Canada"}
	target := []string{"Widget"}

	matrix := BuildSumAggregationMatrix(collapsed, target)

	assert.Equal(t, 2, matrix.Rows)
	assert.Equal(t, 1, matrix.Columns)
	assert.Equal(t, collapsed, matrix.RowLabels)
	assert.Equal(t, target, matrix.ColLabels)

	assert.InDelta(t, 1.0, matrix.Data[0][0], 1e-9)
	assert.InDelta(t, 1.0, matrix.Data[1][0], 1e-9)
}

func TestBuildSumAggregationMatrix_NilTargetValues(t *testing.T) {
	collapsed := []string{"USA", "Canada", "Mexico"}

	matrix := BuildSumAggregationMatrix(collapsed, nil)

	assert.Equal(t, 3, matrix.Rows)
	assert.Equal(t, 1, matrix.Columns)
	assert.Equal(t, collapsed, matrix.RowLabels)
	assert.Nil(t, matrix.ColLabels)

	for i := 0; i < matrix.Rows; i++ {
		assert.InDelta(t, 1.0, matrix.Data[i][0], 1e-9)
	}
}

func TestBuildSumAggregationMatrix_EmptyTargetValues(t *testing.T) {
	collapsed := []string{"USA", "Canada"}

	matrix := BuildSumAggregationMatrix(collapsed, []string{})

	assert.Equal(t, 2, matrix.Rows)
	assert.Equal(t, 1, matrix.Columns)
	assert.Equal(t, collapsed, matrix.RowLabels)
	assert.Empty(t, matrix.ColLabels)

	for i := 0; i < matrix.Rows; i++ {
		assert.InDelta(t, 1.0, matrix.Data[i][0], 1e-9)
	}
}

func TestBuildSumAggregationMatrix_EmptyCollapsed(t *testing.T) {
	matrix := BuildSumAggregationMatrix([]string{}, []string{"Widget"})

	assert.Equal(t, 0, matrix.Rows)
	assert.Equal(t, 1, matrix.Columns)
	assert.Empty(t, matrix.Data)
	assert.Empty(t, matrix.RowLabels)
}

func TestBuildSizeAggregationMatrix(t *testing.T) {
	collapsed := []string{"USA", "Canada", "Mexico"}

	matrix := BuildSizeAggregationMatrix(collapsed)

	assert.Equal(t, 3, matrix.Rows)
	assert.Equal(t, 1, matrix.Columns)
	assert.Equal(t, collapsed, matrix.RowLabels)
	assert.Equal(t, []string{"count"}, matrix.ColLabels)

	for i := 0; i < matrix.Rows; i++ {
		assert.InDelta(t, 1.0, matrix.Data[i][0], 1e-9)
	}
}

func TestBuildSizeAggregationMatrix_Empty(t *testing.T) {
	matrix := BuildSizeAggregationMatrix([]string{})

	assert.Equal(t, 0, matrix.Rows)
	assert.Equal(t, 1, matrix.Columns)
	assert.Empty(t, matrix.Data)
	assert.Empty(t, matrix.RowLabels)
	assert.Equal(t, []string{"count"}, matrix.ColLabels)
}

func TestBuildInverseSumAggregationMatrix_MultipleColumns(t *testing.T) {
	collapsed := []string{"USA", "Canada", "Mexico"}
	target := []string{"Widget", "Gadget"}

	matrix := BuildInverseSumAggregationMatrix(collapsed, target)

	assert.Equal(t, 3, matrix.Rows)
	assert.Equal(t, 2, matrix.Columns)
	assert.Equal(t, collapsed, matrix.RowLabels)
	assert.Equal(t, target, matrix.ColLabels)

	expectedWeight := 1.0 / 3.0
	for i := 0; i < matrix.Rows; i++ {
		assert.Len(t, matrix.Data[i], matrix.Columns)
		for j := 0; j < matrix.Columns; j++ {
			assert.InDelta(t, expectedWeight, matrix.Data[i][j], 1e-9, "entry [%d][%d] should be 1/3", i, j)
		}
	}
}

func TestBuildInverseSumAggregationMatrix_NilTarget(t *testing.T) {
	collapsed := []string{"USA", "Canada", "Mexico", "UK"}

	matrix := BuildInverseSumAggregationMatrix(collapsed, nil)

	assert.Equal(t, 4, matrix.Rows)
	assert.Equal(t, 1, matrix.Columns)
	assert.Equal(t, collapsed, matrix.RowLabels)
	assert.Nil(t, matrix.ColLabels)

	expectedWeight := 1.0 / 4.0
	for i := 0; i < matrix.Rows; i++ {
		assert.InDelta(t, expectedWeight, matrix.Data[i][0], 1e-9)
	}
}

func TestBuildInverseSumAggregationMatrix_EmptyTarget(t *testing.T) {
	collapsed := []string{"USA", "Canada"}

	matrix := BuildInverseSumAggregationMatrix(collapsed, []string{})

	assert.Equal(t, 2, matrix.Rows)
	assert.Equal(t, 1, matrix.Columns)
	assert.Equal(t, collapsed, matrix.RowLabels)
	assert.Empty(t, matrix.ColLabels)

	expectedWeight := 1.0 / 2.0
	for i := 0; i < matrix.Rows; i++ {
		assert.InDelta(t, expectedWeight, matrix.Data[i][0], 1e-9)
	}
}
