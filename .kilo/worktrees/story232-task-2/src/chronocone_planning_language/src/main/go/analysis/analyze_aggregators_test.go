// Package analysis contains unit tests for the AnalyzeAggregatorScope function.
//
// These tests verify that the aggregator analysis phase correctly:
//   - Resolves from-concept and to-concept references from the Concepts map
//   - Evaluates postfix expressions via EvaluateAggregatorExpression
//   - Validates dimensional units via ValidateAggregatorDimensionalUnits
//   - Builds aggregation matrices for sum operations
//   - Generates CIL trees under the "aggregator:" prefix
//   - Detects and reports errors for invalid configurations
//
// Story: Story256 - Aggregators Increment 2 - Analysis Phase
package analysis

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Test Helpers
// ============================================================================

// buildAggregatorScope creates a symbol table representing an AggregatorScope
// with the given from/to concepts and postfix expression definition types.
// The parser stores one AggregatorScopeDefinition + postfix expression defs.
func buildAggregatorScope(fromConcepts []string, toConcept string, toEmpty bool, exprDefs []symboltable.SymbolDefinition) *symboltable.SymbolTable {
	scope := symboltable.NewSymbolTable(symboltable.AggregatorScope)

	// Build the synthetic key
	syntheticKey := AggregatorMapKey(fromConcepts, toConcept)
	storeToConcept := toConcept
	if toEmpty {
		storeToConcept = ""
	}

	// Create the ParserAggregatorDefinition (mimics what parser creates).
	// We use a pointer stored both in scope.Data and the AggregatorScopeDefinition's Data.
	aggDef := &parser.ParserAggregatorDefinition{
		Key:            syntheticKey,
		FromConcepts:   fromConcepts,
		ToConcept:      storeToConcept,
		ToEmpty:        toEmpty,
		SourceLocation: symboltable.SourceLocation{Line: 1, Column: 1},
	}

	scope.Data = aggDef

	// Build definitions: AggregatorScopeDefinition + expression defs
	scopeKey := syntheticKey
	scope.Definitions = []symboltable.SymbolDefinition{
		{
			Name:           &scopeKey,
			DefinitionType: symboltable.AggregatorScopeDefinition,
			Location:       symboltable.SourceLocation{Line: 1, Column: 1},
			Data:           aggDef,
		},
	}
	scope.Definitions = append(scope.Definitions, exprDefs...)

	return scope
}

// makeAggIdentifierDef creates a SymbolDefinition for an Identifier with the given name.
func makeAggIdentifierDef(name string) symboltable.SymbolDefinition {
	return symboltable.SymbolDefinition{
		DefinitionType: symboltable.Identifier,
		Data:           symboltable.StringData{Value: name},
		Location:       symboltable.SourceLocation{Line: 1, Column: 1},
	}
}

// makeAggOperatorDef creates a SymbolDefinition for an operator.
func makeAggOperatorDef(defType symboltable.DefinitionType) symboltable.SymbolDefinition {
	return symboltable.SymbolDefinition{
		DefinitionType: defType,
		Location:       symboltable.SourceLocation{Line: 1, Column: 1},
	}
}

// ============================================================================
// Happy Path Tests - Valid Aggregator Processing
// ============================================================================

// TestAnalyzeAggregatorScope_MultiConceptFromTo_Valid verifies that an aggregator
// from [Country, Product] to [Product] = sum(Country) is correctly processed.
//
// Dimensional check:
//
//	From dims: {Numerator: [Product], Denominator: [Country]}
//	Expression dims (sum(Country)): {Numerator: [Country]}
//	From * Expression = {Numerator: [Product]} (Country cancels)
//	To dims: {Numerator: [Product]}  — match!
func TestAnalyzeAggregatorScope_MultiConceptFromTo_Valid(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with pre-loaded concepts that have values.
	analyzer := NewCPLAnalyzer()
	country := NewValuesConceptDefinition("Country", []string{"USA", "Canada", "Mexico"})
	product := NewValuesConceptDefinition("Product", []string{"Widget", "Gadget"})
	analyzer.AddNewConcept(&country)
	analyzer.AddNewConcept(&product)

	// Expression: sum(Country) — postfix: Country, sum
	exprDefs := []symboltable.SymbolDefinition{
		makeAggIdentifierDef("Country"),
		makeAggIdentifierDef("sum"),
	}

	// Build scope: from [Country, Product] to [Product]
	scope := buildAggregatorScope(
		[]string{"Country", "Product"}, "Product", false, exprDefs,
	)

	// Act: Analyze the aggregator scope
	err := AnalyzeAggregatorScope(scope, NewAnalysisContext(), analyzer)

	// Assert: No error
	require.NoError(t, err, "Should not return error for valid aggregator")

	// Assert: Aggregator was stored in the map
	expectedKey := AggregatorMapKey([]string{"Country", "Product"}, "Product")
	require.Contains(t, analyzer.aggregators, expectedKey)

	info := analyzer.aggregators[expectedKey]
	require.NotNil(t, info)
	assert.False(t, info.ToEmpty)
	assert.Equal(t, "sum", info.CellOperation)

	// Verify from concepts
	require.Len(t, info.FromConcepts, 2)
	assert.Equal(t, "Country", info.FromConcepts[0].ConceptName)
	assert.Equal(t, "Product", info.FromConcepts[1].ConceptName)

	// Verify to concept
	require.NotNil(t, info.ToConcept)
	assert.Equal(t, "Product", info.ToConcept.ConceptName)

	// Verify dimensional units
	assert.Equal(t, []string{"Product"}, info.FromDimensionalUnits.Numerator)
	assert.Equal(t, []string{"Country"}, info.FromDimensionalUnits.Denominator)
	assert.Equal(t, []string{"Country"}, info.ExpressionDimUnits.Numerator)
	assert.Empty(t, info.ExpressionDimUnits.Denominator)
	assert.Equal(t, []string{"Product"}, info.ToDimensionalUnits.Numerator)

	// Assert: CIL tree was created with "aggregator:" prefix
	cilTrees := analyzer.metricAnalyzer.GetCILTrees()
	cilKey := "aggregator:" + expectedKey
	require.Contains(t, cilTrees, cilKey)

	node, ok := cilTrees[cilKey].(*cil.CILAggregateNode)
	require.True(t, ok, "CIL node should be *cil.CILAggregateNode")
	assert.Equal(t, []string{"Country", "Product"}, node.FromDimensions)
	assert.Equal(t, "Product", node.ToDimension)
	assert.Equal(t, "sum", node.CellOperation)

	// Verify aggregation matrix
	require.NotNil(t, node.AggregationMatrix)
	assert.Equal(t, 3, node.AggregationMatrix.Rows)    // USA, Canada, Mexico
	assert.Equal(t, 2, node.AggregationMatrix.Columns) // Widget, Gadget
	assert.Equal(t, []string{"USA", "Canada", "Mexico"}, node.AggregationMatrix.RowLabels)
	assert.Equal(t, []string{"Widget", "Gadget"}, node.AggregationMatrix.ColLabels)

	// Verify all matrix entries are 1.0
	for i := range node.AggregationMatrix.Data {
		for j := range node.AggregationMatrix.Data[i] {
			assert.Equal(t, 1.0, node.AggregationMatrix.Data[i][j])
		}
	}
}

// TestAnalyzeAggregatorScope_SingleConceptRefExpression_Valid verifies an aggregator
// with a simple concept reference expression (no function call).
//
// from [Country, Product] to [Product] = Country
// Postfix: Country
//
//	Dimensional check:
//	  Expr: {Numerator: [Country]}
//	  From: {Numerator: [Product], Denominator: [Country]}
//	  From*Expr: {Numerator: [Product]} ✓
//	  To: {Numerator: [Product]} ✓
func TestAnalyzeAggregatorScope_SingleConceptRefExpression_Valid(t *testing.T) {
	t.Parallel()

	analyzer := NewCPLAnalyzer()
	country := NewValuesConceptDefinition("Country", []string{"USA", "Canada"})
	product := NewValuesConceptDefinition("Product", []string{"Widget", "Gadget"})
	analyzer.AddNewConcept(&country)
	analyzer.AddNewConcept(&product)

	// Expression: just Country (concept reference, no function call)
	exprDefs := []symboltable.SymbolDefinition{
		makeAggIdentifierDef("Country"),
	}

	scope := buildAggregatorScope(
		[]string{"Country", "Product"}, "Product", false, exprDefs,
	)

	err := AnalyzeAggregatorScope(scope, NewAnalysisContext(), analyzer)
	require.NoError(t, err)

	expectedKey := AggregatorMapKey([]string{"Country", "Product"}, "Product")
	info := analyzer.aggregators[expectedKey]
	require.NotNil(t, info)
	assert.Empty(t, info.CellOperation, "CellOperation should be empty when no function is called")

	// Verify CIL tree
	cilTrees := analyzer.metricAnalyzer.GetCILTrees()
	cilKey := "aggregator:" + expectedKey
	node, ok := cilTrees[cilKey].(*cil.CILAggregateNode)
	require.True(t, ok)
	assert.Equal(t, 2, node.AggregationMatrix.Rows)    // USA, Canada
	assert.Equal(t, 2, node.AggregationMatrix.Columns) // Widget, Gadget
}

// TestAnalyzeAggregatorScope_ToEmptySingleFromConcept_Valid verifies that a
// scalar reduction (to []) with exactly one from-concept is successfully processed.
//
// from [Country] to [] = Country/Country/Country
// Postfix: Country, Country, /, Country, /
//
//	Dimensional check:
//	  Country/Country → unitless
//	  unitless/Country → {Denominator: [Country]}
//	  From: {Numerator: [Country]}
//	  From*Expr: {Numerator: [Country]} * {Denominator: [Country]} = unitless ✓
//	  To: unitless ✓
func TestAnalyzeAggregatorScope_ToEmptySingleFromConcept_Valid(t *testing.T) {
	t.Parallel()

	analyzer := NewCPLAnalyzer()
	country := NewValuesConceptDefinition("Country", []string{"USA", "Canada"})
	analyzer.AddNewConcept(&country)

	// Expression: Country/Country/Country — postfix: Country, Country, /, Country, /
	exprDefs := []symboltable.SymbolDefinition{
		makeAggIdentifierDef("Country"),
		makeAggIdentifierDef("Country"),
		makeAggOperatorDef(symboltable.DivideOperatorDefinition),
		makeAggIdentifierDef("Country"),
		makeAggOperatorDef(symboltable.DivideOperatorDefinition),
	}

	// Build scope: from [Country] to []
	scope := buildAggregatorScope(
		[]string{"Country"}, "", true, exprDefs,
	)

	err := AnalyzeAggregatorScope(scope, NewAnalysisContext(), analyzer)
	require.NoError(t, err, "Should not return error for valid scalar reduction")

	expectedKey := AggregatorMapKey([]string{"Country"}, "")
	require.Contains(t, analyzer.aggregators, expectedKey)

	info := analyzer.aggregators[expectedKey]
	require.NotNil(t, info)
	assert.True(t, info.ToEmpty)
	assert.Empty(t, info.CellOperation, "CellOperation should be empty when no function is called")

	// Verify dimensional units
	require.NotNil(t, info.FromDimensionalUnits)
	assert.Equal(t, []string{"Country"}, info.FromDimensionalUnits.Numerator)
	assert.Empty(t, info.FromDimensionalUnits.Denominator)

	require.NotNil(t, info.ToDimensionalUnits)
	assert.True(t, info.ToDimensionalUnits.IsUnitless(), "ToDimensionalUnits should be unitless for scalar reduction")

	// Verify CIL tree
	cilTrees := analyzer.metricAnalyzer.GetCILTrees()
	cilKey := "aggregator:" + expectedKey
	require.Contains(t, cilTrees, cilKey)

	node, ok := cilTrees[cilKey].(*cil.CILAggregateNode)
	require.True(t, ok)
	assert.Equal(t, []string{"Country"}, node.FromDimensions)
	assert.Empty(t, node.ToDimension)
	assert.Equal(t, "1/Country", node.DimensionalUnitEffect)

	// For scalar reduction, matrix should have 1 column
	require.NotNil(t, node.AggregationMatrix)
	assert.Equal(t, 2, node.AggregationMatrix.Rows)      // USA, Canada
	assert.Equal(t, 1, node.AggregationMatrix.Columns)   // scalar
	assert.Equal(t, []string{"USA", "Canada"}, node.AggregationMatrix.RowLabels)
}

// ============================================================================
// Error Cases - Validation Failures
// ============================================================================

// TestAnalyzeAggregatorScope_UndefinedFromConcept_ReturnsError verifies that an
// aggregator referencing an undefined from-concept produces an error.
func TestAnalyzeAggregatorScope_UndefinedFromConcept_ReturnsError(t *testing.T) {
	t.Parallel()

	analyzer := NewCPLAnalyzer()
	// Only add Country, not Product
	country := NewValuesConceptDefinition("Country", []string{"USA"})
	analyzer.AddNewConcept(&country)

	exprDefs := []symboltable.SymbolDefinition{
		makeAggIdentifierDef("Country"),
		makeAggIdentifierDef("sum"),
	}

	scope := buildAggregatorScope(
		[]string{"Country", "UndefinedProduct"}, "UndefinedProduct", false, exprDefs,
	)

	err := AnalyzeAggregatorScope(scope, NewAnalysisContext(), analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "undefined from-concept")
	assert.Contains(t, err.Error(), "UndefinedProduct")
}

// TestAnalyzeAggregatorScope_UndefinedToConcept_ReturnsError verifies that an
// aggregator referencing an undefined to-concept produces an error.
func TestAnalyzeAggregatorScope_UndefinedToConcept_ReturnsError(t *testing.T) {
	t.Parallel()

	analyzer := NewCPLAnalyzer()
	country := NewValuesConceptDefinition("Country", []string{"USA"})
	analyzer.AddNewConcept(&country)
	// Product is NOT added — UndefinedProduct doesn't exist

	exprDefs := []symboltable.SymbolDefinition{
		makeAggIdentifierDef("Country"),
		makeAggIdentifierDef("sum"),
	}

	scope := buildAggregatorScope(
		[]string{"Country", "UndefinedProduct"}, "UndefinedProduct", false, exprDefs,
	)

	err := AnalyzeAggregatorScope(scope, NewAnalysisContext(), analyzer)
	require.Error(t, err)
	// Error should mention the undefined concept (caught as undefined from-concept first)
	assert.Contains(t, err.Error(), "undefined from-concept")
}

// TestAnalyzeAggregatorScope_UndefinedToConceptOnly_ReturnsError verifies that an
// aggregator referencing an undefined to-concept (where from-concepts are valid)
// produces an error specifically for the to-concept.
func TestAnalyzeAggregatorScope_UndefinedToConceptOnly_ReturnsError(t *testing.T) {
	t.Parallel()

	analyzer := NewCPLAnalyzer()
	country := NewValuesConceptDefinition("Country", []string{"USA"})
	analyzer.AddNewConcept(&country)
	// Product is NOT added — UndefinedProduct doesn't exist

	exprDefs := []symboltable.SymbolDefinition{
		makeAggIdentifierDef("Country"),
		makeAggIdentifierDef("sum"),
	}

	scope := buildAggregatorScope(
		[]string{"Country"}, "UndefinedProduct", false, exprDefs,
	)

	err := AnalyzeAggregatorScope(scope, NewAnalysisContext(), analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "aggregator references undefined to-concept")
	assert.Contains(t, err.Error(), "UndefinedProduct")
}

// TestAnalyzeAggregatorScope_ToConceptNotInFromList_ReturnsError verifies that a
// to-concept not in the from list produces an error.
func TestAnalyzeAggregatorScope_ToConceptNotInFromList_ReturnsError(t *testing.T) {
	t.Parallel()

	analyzer := NewCPLAnalyzer()
	country := NewValuesConceptDefinition("Country", []string{"USA"})
	product := NewValuesConceptDefinition("Product", []string{"A"})
	analyzer.AddNewConcept(&country)
	analyzer.AddNewConcept(&product)

	// Expression: sum(Country) with from [Country] to [Product]
	// Product is not in the from list!
	exprDefs := []symboltable.SymbolDefinition{
		makeAggIdentifierDef("Country"),
		makeAggIdentifierDef("sum"),
	}

	scope := buildAggregatorScope(
		[]string{"Country"}, "Product", false, exprDefs,
	)

	err := AnalyzeAggregatorScope(scope, NewAnalysisContext(), analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be in the from-concept list")
	assert.Contains(t, err.Error(), "Product")
}

// TestAnalyzeAggregatorScope_NoCollapsibleDimension_ReturnsError verifies that an
// aggregator with from == to (no dimension to collapse) produces an error.
// e.g., from [Country] to [Country] — all dimensions are preserved, nothing collapses.
func TestAnalyzeAggregatorScope_NoCollapsibleDimension_ReturnsError(t *testing.T) {
	t.Parallel()

	analyzer := NewCPLAnalyzer()
	country := NewValuesConceptDefinition("Country", []string{"USA", "Canada"})
	analyzer.AddNewConcept(&country)

	// Expression: Country/Country — postfix: Country, Country, /
	exprDefs := []symboltable.SymbolDefinition{
		makeAggIdentifierDef("Country"),
		makeAggIdentifierDef("Country"),
		makeAggOperatorDef(symboltable.DivideOperatorDefinition),
	}

	// from [Country] to [Country] — no collapsed dimension
	scope := buildAggregatorScope(
		[]string{"Country"}, "Country", false, exprDefs,
	)

	err := AnalyzeAggregatorScope(scope, NewAnalysisContext(), analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must contain at least one concept different from to-concept")
}

// TestAnalyzeAggregatorScope_MultiConceptFromDifferentTo_Valid verifies that an
// aggregator with from [Country, Product] to [Product] is valid because
// Country is a collapsible dimension (different from to-concept Product).
func TestAnalyzeAggregatorScope_MultiConceptFromDifferentTo_Valid(t *testing.T) {
	t.Parallel()

	analyzer := NewCPLAnalyzer()
	country := NewValuesConceptDefinition("Country", []string{"USA", "Canada"})
	product := NewValuesConceptDefinition("Product", []string{"Widget", "Gadget"})
	analyzer.AddNewConcept(&country)
	analyzer.AddNewConcept(&product)

	// Expression: sum(Country) — postfix: Country, sum
	exprDefs := []symboltable.SymbolDefinition{
		makeAggIdentifierDef("Country"),
		makeAggIdentifierDef("sum"),
	}

	// from [Country, Product] to [Product] — Country differs from Product, valid
	scope := buildAggregatorScope(
		[]string{"Country", "Product"}, "Product", false, exprDefs,
	)

	err := AnalyzeAggregatorScope(scope, NewAnalysisContext(), analyzer)
	require.NoError(t, err, "Multi-concept from with different to should be valid")
}

// TestAnalyzeAggregatorScope_ToEmptyMultipleFromConcepts_ReturnsError verifies that
// scalar reduction (to []) with multiple from-concepts produces an error.
func TestAnalyzeAggregatorScope_ToEmptyMultipleFromConcepts_ReturnsError(t *testing.T) {
	t.Parallel()

	analyzer := NewCPLAnalyzer()
	country := NewValuesConceptDefinition("Country", []string{"USA"})
	product := NewValuesConceptDefinition("Product", []string{"A"})
	analyzer.AddNewConcept(&country)
	analyzer.AddNewConcept(&product)

	exprDefs := []symboltable.SymbolDefinition{
		makeAggIdentifierDef("Country"),
		makeAggIdentifierDef("sum"),
	}

	// to-empty with 2 from-concepts should error
	scope := buildAggregatorScope(
		[]string{"Country", "Product"}, "", true, exprDefs,
	)

	err := AnalyzeAggregatorScope(scope, NewAnalysisContext(), analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires exactly 1 from-concept")
}

// TestAnalyzeAggregatorScope_DuplicateAggregator_ReturnsError verifies that
// analyzing the same aggregator key twice produces an error.
func TestAnalyzeAggregatorScope_DuplicateAggregator_ReturnsError(t *testing.T) {
	t.Parallel()

	analyzer := NewCPLAnalyzer()
	country := NewValuesConceptDefinition("Country", []string{"USA", "Canada"})
	product := NewValuesConceptDefinition("Product", []string{"Widget"})
	analyzer.AddNewConcept(&country)
	analyzer.AddNewConcept(&product)

	exprDefs := []symboltable.SymbolDefinition{
		makeAggIdentifierDef("Country"),
		makeAggIdentifierDef("sum"),
	}

	scope := buildAggregatorScope(
		[]string{"Country", "Product"}, "Product", false, exprDefs,
	)

	// First analysis should succeed
	err := AnalyzeAggregatorScope(scope, NewAnalysisContext(), analyzer)
	require.NoError(t, err)

	// Second analysis of the same scope should fail (duplicate key)
	err = AnalyzeAggregatorScope(scope, NewAnalysisContext(), analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate aggregator definition")
}

// TestAnalyzeAggregatorScope_UnknownExpressionFunction_ReturnsError verifies that
// an expression using an unknown function name produces an error.
func TestAnalyzeAggregatorScope_UnknownExpressionFunction_ReturnsError(t *testing.T) {
	t.Parallel()

	analyzer := NewCPLAnalyzer()
	country := NewValuesConceptDefinition("Country", []string{"USA"})
	product := NewValuesConceptDefinition("Product", []string{"A"})
	analyzer.AddNewConcept(&country)
	analyzer.AddNewConcept(&product)

	// Expression: unknown_function(Country) — postfix: Country, unknown_function
	exprDefs := []symboltable.SymbolDefinition{
		makeAggIdentifierDef("Country"),
		makeAggIdentifierDef("unknown_function"),
	}

	scope := buildAggregatorScope(
		[]string{"Country", "Product"}, "Product", false, exprDefs,
	)

	err := AnalyzeAggregatorScope(scope, NewAnalysisContext(), analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown function or concept")
}

// TestAnalyzeAggregatorScope_NoAggregatorScopeDefinition_ReturnsError verifies that
// a scope without an AggregatorScopeDefinition produces an error.
func TestAnalyzeAggregatorScope_NoAggregatorScopeDefinition_ReturnsError(t *testing.T) {
	t.Parallel()

	analyzer := NewCPLAnalyzer()

	// Create a scope with only expression defs, no AggregatorScopeDefinition
	scope := symboltable.NewSymbolTable(symboltable.AggregatorScope)
	scope.Definitions = []symboltable.SymbolDefinition{
		makeAggIdentifierDef("Country"),
		makeAggIdentifierDef("sum"),
	}

	err := AnalyzeAggregatorScope(scope, NewAnalysisContext(), analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no AggregatorScopeDefinition")
}

// TestAnalyzeAggregatorScope_NilData_ReturnsError verifies that an
// AggregatorScopeDefinition with nil Data produces an error.
func TestAnalyzeAggregatorScope_NilData_ReturnsError(t *testing.T) {
	t.Parallel()

	analyzer := NewCPLAnalyzer()

	scope := symboltable.NewSymbolTable(symboltable.AggregatorScope)
	key := "testKey"
	scope.Definitions = []symboltable.SymbolDefinition{
		{
			Name:           &key,
			DefinitionType: symboltable.AggregatorScopeDefinition,
			Location:       symboltable.SourceLocation{Line: 1, Column: 1},
			Data:           nil, // nil data should be caught
		},
		makeAggIdentifierDef("Country"),
	}

	err := AnalyzeAggregatorScope(scope, NewAnalysisContext(), analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no structured data")
}

// ============================================================================
// CIL Node Interface Conformance Test
// ============================================================================

// TestCILAggregateNode_SatisfiesCILNodeInterface verifies at compile time that
// CILAggregateNode satisfies the cil.CILNode interface.
func TestCILAggregateNode_SatisfiesCILNodeInterface(t *testing.T) {
	t.Parallel()
	var _ cil.CILNode = &cil.CILAggregateNode{}
}

// TestCILAggregateNode_Construction verifies that a CILAggregateNode can be
// constructed with all fields populated.
func TestCILAggregateNode_Construction(t *testing.T) {
	t.Parallel()

	matrixData := [][]float64{
		{1.0, 1.0},
		{1.0, 1.0},
	}

	node := cil.CILAggregateNode{
		FromDimensions: []string{"Country", "Product"},
		ToDimension:    "Product",
		AggregationMatrix: &cil.AggregationMatrix{
			Rows:      2,
			Columns:   2,
			Data:      matrixData,
			RowLabels: []string{"USA", "Canada"},
			ColLabels: []string{"Widget", "Gadget"},
		},
		CellOperation:         "sum",
		DimensionalUnitEffect: "Country",
		SourceLocation:        symboltable.SourceLocation{Line: 1, Column: 1},
	}

	assert.Equal(t, []string{"Country", "Product"}, node.FromDimensions)
	assert.Equal(t, "Product", node.ToDimension)
	require.NotNil(t, node.AggregationMatrix)
	assert.Equal(t, 2, node.AggregationMatrix.Rows)
	assert.Equal(t, 2, node.AggregationMatrix.Columns)
	assert.Equal(t, "sum", node.CellOperation)
	assert.Equal(t, "Country", node.DimensionalUnitEffect)

	// Assert that a CIL tree can store this node
	var n cil.CILNode = &node
	assert.NotNil(t, n)
}
