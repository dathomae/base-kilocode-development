package analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAggregatorFunctionRegistry_HasSumEntry(t *testing.T) {
	reg := NewAggregatorFunctionRegistry()

	assert.NotNil(t, reg)
	assert.NotNil(t, reg.Functions)

	sumInfo, exists := reg.Functions["sum"]
	assert.True(t, exists, "registry should contain 'sum' function")
	assert.NotNil(t, sumInfo)
	assert.Equal(t, "sum", sumInfo.Name)
	assert.Equal(t, "CILAggregateNode", sumInfo.CILNodeType)
	assert.Equal(t, "Sums values across a concept dimension", sumInfo.Description)
}

func TestAggregatorFunctionRegistry_SumDimUnits(t *testing.T) {
	reg := NewAggregatorFunctionRegistry()
	sumInfo := reg.Functions["sum"]

	dimUnits := sumInfo.DimUnits([]string{"Country"})
	assert.NotNil(t, dimUnits)
	assert.Equal(t, []string{"Country"}, dimUnits.Numerator)
	assert.Empty(t, dimUnits.Denominator)
}

func TestAggregatorFunctionRegistry_SumDimUnits_DifferentConcept(t *testing.T) {
	reg := NewAggregatorFunctionRegistry()
	sumInfo := reg.Functions["sum"]

	dimUnits := sumInfo.DimUnits([]string{"Product"})
	assert.NotNil(t, dimUnits)
	assert.Equal(t, []string{"Product"}, dimUnits.Numerator)
	assert.Empty(t, dimUnits.Denominator)
}

func TestAggregatorFunctionRegistry_UnknownFunction(t *testing.T) {
	reg := NewAggregatorFunctionRegistry()

	unknownInfo, exists := reg.Functions["count"]
	assert.False(t, exists, "registry should not contain 'count' in Increment 2")
	assert.Nil(t, unknownInfo)
}

func TestAggregatorFunctionRegistry_InvalidParamCount(t *testing.T) {
	reg := NewAggregatorFunctionRegistry()
	sumInfo := reg.Functions["sum"]

	// No params
	dimUnits := sumInfo.DimUnits([]string{})
	assert.Nil(t, dimUnits)

	// Too many params
	dimUnits = sumInfo.DimUnits([]string{"Country", "Product"})
	assert.Nil(t, dimUnits)
}
