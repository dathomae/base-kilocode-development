package analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDimensionalUnitExpression_Multiply(t *testing.T) {
	// Product × Country produces both in numerator
	product := &DimensionalUnitExpression{Numerator: []string{"Product"}}
	country := &DimensionalUnitExpression{Numerator: []string{"Country"}}
	result := product.Multiply(country)
	assert.Equal(t, []string{"Country", "Product"}, result.Numerator)
	assert.Empty(t, result.Denominator)
}

func TestDimensionalUnitExpression_Multiply_Cancel(t *testing.T) {
	// A/B × B/C → A/C (denominator B cancels with numerator B)
	aOverB := &DimensionalUnitExpression{
		Numerator:   []string{"A"},
		Denominator: []string{"B"},
	}
	bOverC := &DimensionalUnitExpression{
		Numerator:   []string{"B"},
		Denominator: []string{"C"},
	}
	result := aOverB.Multiply(bOverC)
	assert.Equal(t, []string{"A"}, result.Numerator)
	assert.Equal(t, []string{"C"}, result.Denominator)
}

func TestDimensionalUnitExpression_Cancel(t *testing.T) {
	// Common factor Country appears in both and is removed
	expr := &DimensionalUnitExpression{
		Numerator:   []string{"Product", "Country"},
		Denominator: []string{"Country"},
	}
	expr.cancel()
	assert.Equal(t, []string{"Product"}, expr.Numerator)
	assert.Empty(t, expr.Denominator)
}

func TestDimensionalUnitExpression_Cancel_Multiple(t *testing.T) {
	// Multiple common factors
	expr := &DimensionalUnitExpression{
		Numerator:   []string{"A", "A", "B"},
		Denominator: []string{"A", "C"},
	}
	expr.cancel()
	assert.Equal(t, []string{"A", "B"}, expr.Numerator)
	assert.Equal(t, []string{"C"}, expr.Denominator)
}

func TestDimensionalUnitExpression_Equals(t *testing.T) {
	// Same numerators/denominators in different order
	a := &DimensionalUnitExpression{
		Numerator:   []string{"Country", "Product"},
		Denominator: []string{"Status"},
	}
	b := &DimensionalUnitExpression{
		Numerator:   []string{"Product", "Country"},
		Denominator: []string{"Status"},
	}
	assert.True(t, a.Equals(b))
	assert.True(t, b.Equals(a))
}

func TestDimensionalUnitExpression_Equals_Different(t *testing.T) {
	a := &DimensionalUnitExpression{Numerator: []string{"Product"}}
	b := &DimensionalUnitExpression{Numerator: []string{"Country"}}
	assert.False(t, a.Equals(b))
}

func TestDimensionalUnitExpression_IsUnitless(t *testing.T) {
	unitless := &DimensionalUnitExpression{}
	assert.True(t, unitless.IsUnitless())

	notUnitless := &DimensionalUnitExpression{Numerator: []string{"Product"}}
	assert.False(t, notUnitless.IsUnitless())
}

func TestDimensionalUnitExpression_String(t *testing.T) {
	cases := []struct {
		name   string
		expr   *DimensionalUnitExpression
		expect string
	}{
		{"unitless", &DimensionalUnitExpression{}, "Unitless"},
		{"numerator only", &DimensionalUnitExpression{Numerator: []string{"Product"}}, "Product"},
		{"denominator only", &DimensionalUnitExpression{Denominator: []string{"Country"}}, "1/Country"},
		{"both", &DimensionalUnitExpression{Numerator: []string{"Product"}, Denominator: []string{"Country"}}, "Product/Country"},
		{"multiple numerator", &DimensionalUnitExpression{Numerator: []string{"Product", "Status"}}, "Product*Status"},
		{"multiple denominator", &DimensionalUnitExpression{Denominator: []string{"Country", "Region"}}, "1/(Country*Region)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, tc.expr.String())
		})
	}
}

func TestDimensionalUnitsFromConceptSet_Zero(t *testing.T) {
	result := dimensionalUnitsFromConceptSet([]*AnalysisConceptDefinition{})
	assert.True(t, result.IsUnitless())
}

func TestDimensionalUnitsFromConceptSet_One(t *testing.T) {
	concepts := []*AnalysisConceptDefinition{
		{ConceptName: "Product"},
	}
	result := dimensionalUnitsFromConceptSet(concepts)
	assert.Equal(t, []string{"Product"}, result.Numerator)
	assert.Empty(t, result.Denominator)
}

func TestDimensionalUnitsFromConceptSet_Many(t *testing.T) {
	concepts := []*AnalysisConceptDefinition{
		{ConceptName: "Country"},
		{ConceptName: "Product"},
		{ConceptName: "Status"},
	}
	result := dimensionalUnitsFromConceptSet(concepts)
	assert.Equal(t, []string{"Status"}, result.Numerator)
	assert.Equal(t, []string{"Country", "Product"}, result.Denominator)
}

func TestAggregatorMapKey_WithToConcept(t *testing.T) {
	// Sorts from concepts and produces [sorted]->[to]
	key := AggregatorMapKey([]string{"Product", "Country"}, "Product")
	assert.Equal(t, "[Country,Product]->[Product]", key)
}

func TestAggregatorMapKey_EmptyToConcept(t *testing.T) {
	// Scalar case: [sorted]->[]
	key := AggregatorMapKey([]string{"Country"}, "")
	assert.Equal(t, "[Country]->[]", key)
}
