// Package analysis_test contains unit tests for the CPL aggregator evaluator.
//
// These tests follow the Behavioral Test-Driven Development (BTDD) approach,
// where tests are written before the implementation exists.
//
// The tests cover:
//   - Simple function calls (sum)
//   - Concept reference arithmetic (divide, multiply)
//   - Mixed expressions with function calls and arithmetic
//   - Error cases (unknown function, insufficient operands, invalid params)
//   - Dimensional unit validation
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Test Helpers
// ============================================================================

func makeConceptRef(name string) symboltable.SymbolDefinition {
	return symboltable.SymbolDefinition{
		DefinitionType: symboltable.Identifier,
		Data:           symboltable.StringData{Value: name},
	}
}

func makeAggOperator(defType symboltable.DefinitionType) symboltable.SymbolDefinition {
	return symboltable.SymbolDefinition{
		DefinitionType: defType,
	}
}

func makeFromConcepts(names ...string) []*analysis.AnalysisConceptDefinition {
	concepts := make([]*analysis.AnalysisConceptDefinition, len(names))
	for i, name := range names {
		concepts[i] = &analysis.AnalysisConceptDefinition{ConceptName: name}
	}
	return concepts
}

// ============================================================================
// evaluateAggregatorExpression Tests
// ============================================================================

// TestEvaluateAggregatorExpression_SumFunction verifies that sum(Country)
// produces a DimensionalUnitExpression with Numerator [Country].
func TestEvaluateAggregatorExpression_SumFunction(t *testing.T) {
	t.Parallel()

	registry := analysis.NewAggregatorFunctionRegistry()
	defs := []symboltable.SymbolDefinition{
		makeConceptRef("Country"),
		makeConceptRef("sum"),
	}
	fromConcepts := makeFromConcepts("Country")

	result, cellOp, err := analysis.EvaluateAggregatorExpression(defs, fromConcepts, registry)
	require.NoError(t, err)
	assert.Equal(t, "sum", cellOp)
	assert.Equal(t, []string{"Country"}, result.Numerator)
	assert.Empty(t, result.Denominator)
}

// TestEvaluateAggregatorExpression_ConceptDivide verifies that Country/Country
// produces a unitless DimensionalUnitExpression.
func TestEvaluateAggregatorExpression_ConceptDivide(t *testing.T) {
	t.Parallel()

	registry := analysis.NewAggregatorFunctionRegistry()
	defs := []symboltable.SymbolDefinition{
		makeConceptRef("Country"),
		makeConceptRef("Country"),
		makeAggOperator(symboltable.DivideOperatorDefinition),
	}
	fromConcepts := makeFromConcepts("Country")

	result, cellOp, err := analysis.EvaluateAggregatorExpression(defs, fromConcepts, registry)
	require.NoError(t, err)
	assert.Empty(t, cellOp)
	assert.True(t, result.IsUnitless(), "expected unitless result, got %s", result)
}

// TestEvaluateAggregatorExpression_MixedExpression verifies that
// (Country/Country)*sum(Country) produces {Numerator: [Country]}.
func TestEvaluateAggregatorExpression_MixedExpression(t *testing.T) {
	t.Parallel()

	registry := analysis.NewAggregatorFunctionRegistry()
	defs := []symboltable.SymbolDefinition{
		makeConceptRef("Country"),
		makeConceptRef("Country"),
		makeAggOperator(symboltable.DivideOperatorDefinition),
		makeConceptRef("Country"),
		makeConceptRef("sum"),
		makeAggOperator(symboltable.MultiplyOperatorDefinition),
	}
	fromConcepts := makeFromConcepts("Country")

	result, cellOp, err := analysis.EvaluateAggregatorExpression(defs, fromConcepts, registry)
	require.NoError(t, err)
	assert.Equal(t, "sum", cellOp)
	assert.Equal(t, []string{"Country"}, result.Numerator)
	assert.Empty(t, result.Denominator)
}

// TestEvaluateAggregatorExpression_UnknownFunction verifies that an unknown
// function name returns an error.
func TestEvaluateAggregatorExpression_UnknownFunction(t *testing.T) {
	t.Parallel()

	registry := analysis.NewAggregatorFunctionRegistry()
	defs := []symboltable.SymbolDefinition{
		makeConceptRef("Country"),
		makeConceptRef("unknown"),
	}
	fromConcepts := makeFromConcepts("Country")

	_, _, err := analysis.EvaluateAggregatorExpression(defs, fromConcepts, registry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown", "error should indicate unknown function or concept")
}

// TestEvaluateAggregatorExpression_InsufficientOperands verifies that an
// operator with too few operands returns an error.
func TestEvaluateAggregatorExpression_InsufficientOperands(t *testing.T) {
	t.Parallel()

	registry := analysis.NewAggregatorFunctionRegistry()
	defs := []symboltable.SymbolDefinition{
		makeAggOperator(symboltable.MultiplyOperatorDefinition),
	}
	fromConcepts := makeFromConcepts("Country")

	_, _, err := analysis.EvaluateAggregatorExpression(defs, fromConcepts, registry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient operands", "error should indicate insufficient operands")
}

// TestEvaluateAggregatorExpression_InvalidParams verifies that passing
// invalid params to a function returns an error.
func TestEvaluateAggregatorExpression_InvalidParams(t *testing.T) {
	t.Parallel()

	registry := analysis.NewAggregatorFunctionRegistry()
	// Build postfix: Country Country / sum
	// sum receives a unitless expression (not a simple concept ref)
	defs := []symboltable.SymbolDefinition{
		makeConceptRef("Country"),
		makeConceptRef("Country"),
		makeAggOperator(symboltable.DivideOperatorDefinition),
		makeConceptRef("sum"),
	}
	fromConcepts := makeFromConcepts("Country")

	_, _, err := analysis.EvaluateAggregatorExpression(defs, fromConcepts, registry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid params", "error should indicate invalid params to function")
}

// ============================================================================
// validateAggregatorDimensionalUnits Tests
// ============================================================================

// TestValidateAggregatorDimensionalUnits_Valid verifies that valid dimensional
// units pass validation.
func TestValidateAggregatorDimensionalUnits_Valid(t *testing.T) {
	t.Parallel()

	// from [Country, Product] to [Product] with expression sum(Country)
	// From dims: {Numerator: [Product], Denominator: [Country]}
	// Expression dims: {Numerator: [Country]}
	// Expected to dims: {Numerator: [Product]}
	// From * Expression = {Numerator: [Product], Denominator: [Country]} * {Numerator: [Country]}
	//                  = {Numerator: [Product]}
	// Which equals to dims. Valid!

	info := &analysis.AggregatorInfo{
		FromConcepts: makeFromConcepts("Country", "Product"),
		ToConcept:    &analysis.AnalysisConceptDefinition{ConceptName: "Product"},
		ToEmpty:      false,
		ExpressionDimUnits: &analysis.DimensionalUnitExpression{
			Numerator: []string{"Country"},
		},
	}

	err := analysis.ValidateAggregatorDimensionalUnits(info)
	require.NoError(t, err)
	assert.Equal(t, []string{"Product"}, info.FromDimensionalUnits.Numerator)
	assert.Equal(t, []string{"Country"}, info.FromDimensionalUnits.Denominator)
	assert.Equal(t, []string{"Product"}, info.ToDimensionalUnits.Numerator)
}

// TestValidateAggregatorDimensionalUnits_Mismatch verifies that a dimensional
// unit mismatch returns a descriptive error.
func TestValidateAggregatorDimensionalUnits_Mismatch(t *testing.T) {
	t.Parallel()

	// from [Country] to [Product] with expression sum(Country)
	// From dims: {Numerator: [Country]}
	// Expression dims: {Numerator: [Country]}
	// Expected to dims: {Numerator: [Product]}
	// From * Expression = {Numerator: [Country]} * {Numerator: [Country]}
	//                  = {Numerator: [Country, Country]}
	// Which does NOT equal to dims {Numerator: [Product]}. Mismatch!

	info := &analysis.AggregatorInfo{
		FromConcepts: makeFromConcepts("Country"),
		ToConcept:    &analysis.AnalysisConceptDefinition{ConceptName: "Product"},
		ToEmpty:      false,
		ExpressionDimUnits: &analysis.DimensionalUnitExpression{
			Numerator: []string{"Country"},
		},
	}

	err := analysis.ValidateAggregatorDimensionalUnits(info)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mismatch", "error should indicate dimensional unit mismatch")
	assert.Contains(t, err.Error(), "from", "error should mention from units")
	assert.Contains(t, err.Error(), "expression", "error should mention expression units")
	assert.Contains(t, err.Error(), "expected to", "error should mention expected to units")
}

// TestValidateAggregatorDimensionalUnits_EmptyTo verifies that scalar reduction
// (to []) passes validation when the output is unitless.
//
// Dimensional check:
//
//	from [Country, Product] to [] = Country/Product
//	From = {Numerator: [Product], Denominator: [Country]}
//	Expression = {Numerator: [Country], Denominator: [Product]}
//	To = unitless
//	From * Expression = unitless (everything cancels) — match!
func TestValidateAggregatorDimensionalUnits_EmptyTo(t *testing.T) {
	t.Parallel()

	info := &analysis.AggregatorInfo{
		FromConcepts: makeFromConcepts("Country", "Product"),
		ToEmpty:      true,
		ExpressionDimUnits: &analysis.DimensionalUnitExpression{
			Numerator:   []string{"Country"},
			Denominator: []string{"Product"},
		},
	}

	err := analysis.ValidateAggregatorDimensionalUnits(info)
	require.NoError(t, err)
	assert.True(t, info.ToDimensionalUnits.IsUnitless())
}
