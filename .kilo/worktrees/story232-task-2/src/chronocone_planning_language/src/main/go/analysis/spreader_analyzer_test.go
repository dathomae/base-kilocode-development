package analysis

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSpreaderFunctionRegistry_HasAllEntries(t *testing.T) {
	reg := NewSpreaderFunctionRegistry()

	assert.NotNil(t, reg)
	assert.NotNil(t, reg.Functions)
	assert.Len(t, reg.Functions, 6)

	expectedNames := []string{"size", "1/size", "sum", "count", "1/sum", "1/count"}
	for _, name := range expectedNames {
		info, exists := reg.Functions[name]
		assert.True(t, exists, "registry should contain '%s' function", name)
		assert.NotNil(t, info)
		assert.Equal(t, name, info.Name)
		assert.Equal(t, 1, info.Arity, "%s should have arity 1", name)
	}
}

func TestSpreaderFunctionRegistry_SizeDimUnits(t *testing.T) {
	reg := NewSpreaderFunctionRegistry()
	info := reg.Functions["size"]

	dimUnits := info.DimUnits([]string{"Country"})
	assert.NotNil(t, dimUnits)
	assert.Empty(t, dimUnits.Numerator)
	assert.Empty(t, dimUnits.Denominator)
	assert.True(t, dimUnits.IsUnitless())
}

func TestSpreaderFunctionRegistry_InverseSizeDimUnits(t *testing.T) {
	reg := NewSpreaderFunctionRegistry()
	info := reg.Functions["1/size"]

	dimUnits := info.DimUnits([]string{"Country"})
	assert.NotNil(t, dimUnits)
	assert.Empty(t, dimUnits.Numerator)
	assert.Empty(t, dimUnits.Denominator)
	assert.True(t, dimUnits.IsUnitless())
}

func TestSpreaderFunctionRegistry_SumDimUnits(t *testing.T) {
	reg := NewSpreaderFunctionRegistry()
	info := reg.Functions["sum"]

	dimUnits := info.DimUnits([]string{"Country"})
	assert.NotNil(t, dimUnits)
	assert.Equal(t, []string{"Country"}, dimUnits.Numerator)
	assert.Empty(t, dimUnits.Denominator)
}

func TestSpreaderFunctionRegistry_CountDimUnits(t *testing.T) {
	reg := NewSpreaderFunctionRegistry()
	info := reg.Functions["count"]

	dimUnits := info.DimUnits([]string{"Country"})
	assert.NotNil(t, dimUnits)
	assert.Equal(t, []string{"Country"}, dimUnits.Numerator)
	assert.Empty(t, dimUnits.Denominator)
}

func TestSpreaderFunctionRegistry_InverseSumDimUnits(t *testing.T) {
	reg := NewSpreaderFunctionRegistry()
	info := reg.Functions["1/sum"]

	dimUnits := info.DimUnits([]string{"Country"})
	assert.NotNil(t, dimUnits)
	assert.Empty(t, dimUnits.Numerator)
	assert.Equal(t, []string{"Country"}, dimUnits.Denominator)
}

func TestSpreaderFunctionRegistry_InverseCountDimUnits(t *testing.T) {
	reg := NewSpreaderFunctionRegistry()
	info := reg.Functions["1/count"]

	dimUnits := info.DimUnits([]string{"Country"})
	assert.NotNil(t, dimUnits)
	assert.Empty(t, dimUnits.Numerator)
	assert.Equal(t, []string{"Country"}, dimUnits.Denominator)
}

func TestSpreaderFunctionRegistry_DimUnitsDifferentConcept(t *testing.T) {
	reg := NewSpreaderFunctionRegistry()

	// Test with a different concept name for functions that place param in numerator
	sumInfo := reg.Functions["sum"]
	dimUnits := sumInfo.DimUnits([]string{"Product"})
	assert.NotNil(t, dimUnits)
	assert.Equal(t, []string{"Product"}, dimUnits.Numerator)
	assert.Empty(t, dimUnits.Denominator)

	// Test with a different concept name for functions that place param in denominator
	invSumInfo := reg.Functions["1/sum"]
	dimUnits = invSumInfo.DimUnits([]string{"Product"})
	assert.NotNil(t, dimUnits)
	assert.Empty(t, dimUnits.Numerator)
	assert.Equal(t, []string{"Product"}, dimUnits.Denominator)
}

func TestSpreaderFunctionRegistry_UnknownFunction(t *testing.T) {
	reg := NewSpreaderFunctionRegistry()

	unknownInfo, exists := reg.Functions["unknown"]
	assert.False(t, exists, "registry should not contain 'unknown' function")
	assert.Nil(t, unknownInfo)
}

func TestSpreaderFunctionRegistry_InvalidParamCount(t *testing.T) {
	reg := NewSpreaderFunctionRegistry()

	// Test each function with invalid param counts
	for _, name := range []string{"size", "1/size", "sum", "count", "1/sum", "1/count"} {
		info := reg.Functions[name]

		// No params
		dimUnits := info.DimUnits([]string{})
		assert.Nil(t, dimUnits, "%s should return nil with zero params", name)

		// Too many params
		dimUnits = info.DimUnits([]string{"Country", "Product"})
		assert.Nil(t, dimUnits, "%s should return nil with two params", name)
	}
}

func TestSpreaderFunctionRegistry_CILNodeTypes(t *testing.T) {
	reg := NewSpreaderFunctionRegistry()

	cases := []struct {
		name         string
		expectedType string
	}{
		{"size", "CILSpreadSizeNode"},
		{"1/size", "CILSpreadInverseSizeNode"},
		{"sum", "CILSpreadSumNode"},
		{"count", "CILSpreadCountNode"},
		{"1/sum", "CILSpreadInverseSumNode"},
		{"1/count", "CILSpreadInverseCountNode"},
	}

	for _, tc := range cases {
		info := reg.Functions[tc.name]
		assert.Equal(t, tc.expectedType, info.CILNodeType, "%s CILNodeType mismatch", tc.name)
	}
}

func TestSpreaderFunctionRegistry_Descriptions(t *testing.T) {
	reg := NewSpreaderFunctionRegistry()

	cases := []struct {
		name         string
		expectedDesc string
	}{
		{"size", "Number of distinct values in concept"},
		{"1/size", "1/N where N is number of values in concept"},
		{"sum", "Sums values across a dimension (implicit aggs)"},
		{"count", "Counts values in a dimension"},
		{"1/sum", "Inverse of sum across a dimension"},
		{"1/count", "Inverse of count across a dimension"},
	}

	for _, tc := range cases {
		info := reg.Functions[tc.name]
		assert.Equal(t, tc.expectedDesc, info.Description, "%s Description mismatch", tc.name)
	}
}

// ============================================================================
// Test Helpers
// ============================================================================

func makeTestAnalyzerWithConcepts(concepts ...string) *CPLAnalyzer {
	analyzer := NewCPLAnalyzer()
	for _, name := range concepts {
		analyzer.AddNewConcept(&AnalysisConceptDefinition{ConceptName: name})
	}
	return analyzer
}

func makeSpreaderIdentifier(name string) symboltable.SymbolDefinition {
	return symboltable.SymbolDefinition{
		DefinitionType: symboltable.Identifier,
		Data:           symboltable.StringData{Value: name},
	}
}

func makeSpreaderOperator(defType symboltable.DefinitionType) symboltable.SymbolDefinition {
	return symboltable.SymbolDefinition{
		DefinitionType: defType,
	}
}

// ============================================================================
// resolveConceptList Tests
// ============================================================================

func TestResolveConceptList_Success(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")

	concepts, err := analyzer.resolveConceptList([]string{"Product", "Country"})
	require.NoError(t, err)
	require.Len(t, concepts, 2)
	assert.Equal(t, "Product", concepts[0].ConceptName)
	assert.Equal(t, "Country", concepts[1].ConceptName)
}

func TestResolveConceptList_UnknownConcept(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country")

	_, err := analyzer.resolveConceptList([]string{"Country", "Unknown"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown concept")
	assert.Contains(t, err.Error(), "Unknown")
}

// ============================================================================
// validateSpreaderDimensionalRules Tests
// ============================================================================

func TestValidateSpreaderDimensionalRules_Valid(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}, {ConceptName: "Product"}}

	err := analyzer.validateSpreaderDimensionalRules(from, to)
	assert.NoError(t, err)
}

func TestValidateSpreaderDimensionalRules_TooManyNewDims(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product", "Status")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{
		{ConceptName: "Country"},
		{ConceptName: "Product"},
		{ConceptName: "Status"},
	}

	err := analyzer.validateSpreaderDimensionalRules(from, to)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one dimension")
}

func TestValidateSpreaderDimensionalRules_FromNotInTo(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product", "Status")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}, {ConceptName: "Status"}}

	err := analyzer.validateSpreaderDimensionalRules(from, to)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in to concepts")
	assert.Contains(t, err.Error(), "Product")
}

// ============================================================================
// findNewDimension Tests
// ============================================================================

func TestFindNewDimension_Success(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}, {ConceptName: "Product"}}

	newDim, err := analyzer.findNewDimension(from, to)
	require.NoError(t, err)
	assert.Equal(t, "Country", newDim.ConceptName)
}

func TestFindNewDimension_ScalarSpreading(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country")
	from := []*AnalysisConceptDefinition{}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}}

	newDim, err := analyzer.findNewDimension(from, to)
	require.NoError(t, err)
	assert.Equal(t, "Country", newDim.ConceptName)
}

func TestFindNewDimension_NoNewDimension(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("A", "B")
	from := []*AnalysisConceptDefinition{{ConceptName: "A"}, {ConceptName: "B"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "A"}, {ConceptName: "B"}}

	_, err := analyzer.findNewDimension(from, to)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not find new dimension")
}

// ============================================================================
// evaluateSpreaderExpression Tests
// ============================================================================

func TestEvaluateSpreaderExpression_UnitlessInverseSize(t *testing.T) {
	// 1/size(Country) postfix: Country, 1/size
	analyzer := makeTestAnalyzerWithConcepts("Country")
	from := []*AnalysisConceptDefinition{{ConceptName: "Country"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}}

	defs := []symboltable.SymbolDefinition{
		makeSpreaderIdentifier("Country"),
		makeSpreaderIdentifier("1/size"),
	}

	result, err := analyzer.evaluateSpreaderExpression(defs, from, to)
	require.NoError(t, err)
	assert.True(t, result.IsUnitless(), "expected unitless result for 1/size(Country)")
}

func TestEvaluateSpreaderExpression_SizeFunction(t *testing.T) {
	// size(Country) postfix: Country, size
	// size() is unitless
	analyzer := makeTestAnalyzerWithConcepts("Country")
	from := []*AnalysisConceptDefinition{{ConceptName: "Country"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}}

	defs := []symboltable.SymbolDefinition{
		makeSpreaderIdentifier("Country"),
		makeSpreaderIdentifier("size"),
	}

	result, err := analyzer.evaluateSpreaderExpression(defs, from, to)
	require.NoError(t, err)
	assert.True(t, result.IsUnitless(), "expected unitless result for size(Country)")
	assert.Empty(t, result.Numerator)
	assert.Empty(t, result.Denominator)
}

func TestEvaluateSpreaderExpression_ProductTimesInverseSize(t *testing.T) {
	// Product * 1/size(Country) postfix: Product, Country, 1/size, *
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}, {ConceptName: "Product"}}

	defs := []symboltable.SymbolDefinition{
		makeSpreaderIdentifier("Product"),
		makeSpreaderIdentifier("Country"),
		makeSpreaderIdentifier("1/size"),
		makeSpreaderOperator(symboltable.MultiplyOperatorDefinition),
	}

	result, err := analyzer.evaluateSpreaderExpression(defs, from, to)
	require.NoError(t, err)
	assert.Equal(t, []string{"Product"}, result.Numerator)
	assert.Empty(t, result.Denominator)
}

func TestEvaluateSpreaderExpression_MultiConceptMultiply(t *testing.T) {
	// Country * Product postfix: Country, Product, *
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}, {ConceptName: "Product"}}

	defs := []symboltable.SymbolDefinition{
		makeSpreaderIdentifier("Country"),
		makeSpreaderIdentifier("Product"),
		makeSpreaderOperator(symboltable.MultiplyOperatorDefinition),
	}

	result, err := analyzer.evaluateSpreaderExpression(defs, from, to)
	require.NoError(t, err)
	assert.Equal(t, []string{"Country", "Product"}, result.Numerator)
	assert.Empty(t, result.Denominator)
}

func TestEvaluateSpreaderExpression_IdentifierResolution(t *testing.T) {
	// Verify that concepts in from/to resolve correctly
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}, {ConceptName: "Product"}}

	// Just Country
	defs := []symboltable.SymbolDefinition{
		makeSpreaderIdentifier("Country"),
	}

	result, err := analyzer.evaluateSpreaderExpression(defs, from, to)
	require.NoError(t, err)
	assert.Equal(t, []string{"Country"}, result.Numerator)
}

func TestEvaluateSpreaderExpression_UnknownIdentifierError(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}, {ConceptName: "Product"}}

	defs := []symboltable.SymbolDefinition{
		makeSpreaderIdentifier("UnknownMetric"),
	}

	_, err := analyzer.evaluateSpreaderExpression(defs, from, to)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown identifier")
	assert.Contains(t, err.Error(), "UnknownMetric")
	assert.Contains(t, err.Error(), "Increment 1 supports concepts only")
}

func TestEvaluateSpreaderExpression_DivideOperator(t *testing.T) {
	// Product / Country postfix: Product, Country, /
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}, {ConceptName: "Product"}}

	defs := []symboltable.SymbolDefinition{
		makeSpreaderIdentifier("Product"),
		makeSpreaderIdentifier("Country"),
		makeSpreaderOperator(symboltable.DivideOperatorDefinition),
	}

	result, err := analyzer.evaluateSpreaderExpression(defs, from, to)
	require.NoError(t, err)
	assert.Equal(t, []string{"Product"}, result.Numerator)
	assert.Equal(t, []string{"Country"}, result.Denominator)
}

func TestEvaluateSpreaderExpression_PlusOperator(t *testing.T) {
	// Country + Product postfix: Country, Product, +
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}, {ConceptName: "Product"}}

	defs := []symboltable.SymbolDefinition{
		makeSpreaderIdentifier("Country"),
		makeSpreaderIdentifier("Product"),
		makeSpreaderOperator(symboltable.PlusOperatorDefinition),
	}

	result, err := analyzer.evaluateSpreaderExpression(defs, from, to)
	require.NoError(t, err)
	assert.Equal(t, []string{"Country"}, result.Numerator)
	assert.Empty(t, result.Denominator)
}

func TestEvaluateSpreaderExpression_MinusOperator(t *testing.T) {
	// Country - Product postfix: Country, Product, -
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}, {ConceptName: "Product"}}

	defs := []symboltable.SymbolDefinition{
		makeSpreaderIdentifier("Country"),
		makeSpreaderIdentifier("Product"),
		makeSpreaderOperator(symboltable.MinusOperatorDefinition),
	}

	result, err := analyzer.evaluateSpreaderExpression(defs, from, to)
	require.NoError(t, err)
	assert.Equal(t, []string{"Country"}, result.Numerator)
	assert.Empty(t, result.Denominator)
}

func TestEvaluateSpreaderExpression_EmptyExpression(t *testing.T) {
	// Zero definitions should result in "malformed expression" error
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}, {ConceptName: "Product"}}

	defs := []symboltable.SymbolDefinition{}

	_, err := analyzer.evaluateSpreaderExpression(defs, from, to)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "malformed expression")
}

func TestEvaluateSpreaderExpression_InsufficientOperandsMultiply(t *testing.T) {
	// * with only one operand on stack: Product, *
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}, {ConceptName: "Product"}}

	defs := []symboltable.SymbolDefinition{
		makeSpreaderIdentifier("Product"),
		makeSpreaderOperator(symboltable.MultiplyOperatorDefinition),
	}

	_, err := analyzer.evaluateSpreaderExpression(defs, from, to)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient operands")
}

func TestEvaluateSpreaderExpression_SpreaderScopeDefinitionSkip(t *testing.T) {
	// SpreaderScopeDefinition should be skipped and not affect the stack
	analyzer := makeTestAnalyzerWithConcepts("Country")
	from := []*AnalysisConceptDefinition{{ConceptName: "Country"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}}

	defs := []symboltable.SymbolDefinition{
		makeSpreaderIdentifier("Country"),
		{DefinitionType: symboltable.SpreaderScopeDefinition},
		makeSpreaderIdentifier("1/size"),
	}

	result, err := analyzer.evaluateSpreaderExpression(defs, from, to)
	require.NoError(t, err)
	assert.True(t, result.IsUnitless(), "expected unitless result for 1/size(Country) with skipped scope def")
}

func TestEvaluateSpreaderExpression_UnitReferenceError(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}, {ConceptName: "Product"}}

	defs := []symboltable.SymbolDefinition{
		{
			DefinitionType: symboltable.UnitReferenceDefinition,
			Data:           parser.ParserUnitReference{Name: "USD"},
		},
	}

	_, err := analyzer.evaluateSpreaderExpression(defs, from, to)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unit references not supported in spreader expressions")
	assert.Contains(t, err.Error(), "USD")
}

func TestEvaluateSpreaderExpression_MalformedExpression(t *testing.T) {
	// Too many items on stack: Product, Country
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}, {ConceptName: "Product"}}

	defs := []symboltable.SymbolDefinition{
		makeSpreaderIdentifier("Product"),
		makeSpreaderIdentifier("Country"),
	}

	_, err := analyzer.evaluateSpreaderExpression(defs, from, to)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "malformed expression")
}

func TestEvaluateSpreaderExpression_FunctionInvalidParams(t *testing.T) {
	// 1/size with a non-simple-concept param: Product, Country, *, 1/size
	// This would try to pass {Country*Product} as param to 1/size
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")
	from := []*AnalysisConceptDefinition{{ConceptName: "Product"}}
	to := []*AnalysisConceptDefinition{{ConceptName: "Country"}, {ConceptName: "Product"}}

	defs := []symboltable.SymbolDefinition{
		makeSpreaderIdentifier("Product"),
		makeSpreaderIdentifier("Country"),
		makeSpreaderOperator(symboltable.MultiplyOperatorDefinition),
		makeSpreaderIdentifier("1/size"),
	}

	_, err := analyzer.evaluateSpreaderExpression(defs, from, to)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid params")
}

// ============================================================================
// AnalyzeSpreaderScope Tests
// ============================================================================

func TestAnalyzeSpreaderScope_Success(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")

	key := "__spreader__[Product]->[Country,Product]"
	pDef := &parser.ParserSpreaderDefinition{
		Key:            key,
		FromConcepts:   []string{"Product"},
		ToConcepts:     []string{"Country", "Product"},
		FromEmpty:      false,
		SourceLocation: symboltable.SourceLocation{Line: 1, Column: 1},
	}

	scope := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	_ = scope.DefineWithData(&key, symboltable.SpreaderScopeDefinition, pDef.SourceLocation, pDef)
	// Valid expression: 1/count(Country) in postfix -> Country, 1/count
	scope.Definitions = append(scope.Definitions, makeSpreaderIdentifier("Country"))
	scope.Definitions = append(scope.Definitions, makeSpreaderIdentifier("1/count"))

	err := AnalyzeSpreaderScope(scope, nil, analyzer)
	assert.NoError(t, err)

	// Verify SpreaderInfo was stored in the analyzer's spreaders map
	mapKey := SpreaderMapKey([]string{"Product"}, []string{"Country", "Product"})
	info, exists := analyzer.spreaders[mapKey]
	require.True(t, exists, "SpreaderInfo should be stored in spreaders map")
	require.NotNil(t, info)

	// Verify all fields are populated
	require.Len(t, info.FromConcepts, 1)
	assert.Equal(t, "Product", info.FromConcepts[0].ConceptName)
	require.Len(t, info.ToConcepts, 2)
	assert.Equal(t, "Country", info.ToConcepts[0].ConceptName)
	assert.Equal(t, "Product", info.ToConcepts[1].ConceptName)
	assert.False(t, info.FromEmpty)
	require.NotNil(t, info.NewDimension)
	assert.Equal(t, "Country", info.NewDimension.ConceptName)
	require.NotNil(t, info.FromDimensionalUnits)
	assert.Equal(t, []string{"Product"}, info.FromDimensionalUnits.Denominator)
	require.NotNil(t, info.ToDimensionalUnits)
	assert.Equal(t, []string{"Country", "Product"}, info.ToDimensionalUnits.Denominator)
	require.NotNil(t, info.ExpressionDimUnits)
	assert.Equal(t, []string{"Country"}, info.ExpressionDimUnits.Denominator)
	assert.Equal(t, 1, info.SourceLocation.Line)
	assert.Equal(t, 1, info.SourceLocation.Column)

	// Verify GetAnalysisDefinitions() exposes the Spreaders map
	defs := analyzer.GetAnalysisDefinitions()
	require.NotNil(t, defs)
	require.Contains(t, defs.Spreaders, mapKey)
	assert.Equal(t, info, defs.Spreaders[mapKey])
}

func TestAnalyzeSpreaderScope_ScalarSpread(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country")

	key := "__spreader__[]->[Country]"
	pDef := &parser.ParserSpreaderDefinition{
		Key:            key,
		FromConcepts:   []string{},
		ToConcepts:     []string{"Country"},
		FromEmpty:      true,
		SourceLocation: symboltable.SourceLocation{Line: 5, Column: 10},
	}

	scope := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	_ = scope.DefineWithData(&key, symboltable.SpreaderScopeDefinition, pDef.SourceLocation, pDef)
	// Valid expression: 1/count(Country) -> gives 1/Country
	scope.Definitions = append(scope.Definitions, makeSpreaderIdentifier("Country"))
	scope.Definitions = append(scope.Definitions, makeSpreaderIdentifier("1/count"))

	err := AnalyzeSpreaderScope(scope, nil, analyzer)
	assert.NoError(t, err)

	mapKey := SpreaderMapKey([]string{}, []string{"Country"})
	info, exists := analyzer.spreaders[mapKey]
	require.True(t, exists, "SpreaderInfo should be stored in spreaders map")
	require.NotNil(t, info)
	assert.True(t, info.FromEmpty)
	assert.Equal(t, 5, info.SourceLocation.Line)
	assert.Equal(t, 10, info.SourceLocation.Column)
	require.NotNil(t, info.NewDimension)
	assert.Equal(t, "Country", info.NewDimension.ConceptName)
	require.NotNil(t, info.FromDimensionalUnits)
	assert.True(t, info.FromDimensionalUnits.IsUnitless())
	require.NotNil(t, info.ToDimensionalUnits)
	assert.Equal(t, []string{"Country"}, info.ToDimensionalUnits.Denominator)
	require.NotNil(t, info.ExpressionDimUnits)
	assert.Equal(t, []string{"Country"}, info.ExpressionDimUnits.Denominator)

	// Verify GetAnalysisDefinitions() exposes the Spreaders map
	defs := analyzer.GetAnalysisDefinitions()
	require.NotNil(t, defs)
	require.Contains(t, defs.Spreaders, mapKey)
	assert.Equal(t, info, defs.Spreaders[mapKey])
}

func TestAnalyzeSpreaderScope_DimensionalUnitMismatch(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")

	key := "__spreader__[Product]->[Country,Product]"
	pDef := &parser.ParserSpreaderDefinition{
		Key:          key,
		FromConcepts: []string{"Product"},
		ToConcepts:   []string{"Country", "Product"},
	}

	scope := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	_ = scope.DefineWithData(&key, symboltable.SpreaderScopeDefinition, symboltable.SourceLocation{}, pDef)
	// Invalid expression: just Country -> gives Country (numerator)
	// Needed: 1/Country (denominator) to satisfy 1/Product * 1/Country = 1/(Country*Product)
	scope.Definitions = append(scope.Definitions, makeSpreaderIdentifier("Country"))

	err := AnalyzeSpreaderScope(scope, nil, analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dimensional unit mismatch")
}

func TestAnalyzeSpreaderScope_MissingSpreaderScopeDefinition(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")

	scope := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	scope.Definitions = append(scope.Definitions, makeSpreaderIdentifier("Country"))

	err := AnalyzeSpreaderScope(scope, nil, analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no SpreaderScopeDefinition")
}

func TestAnalyzeSpreaderScope_UnknownFromConcept(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Product")

	key := "__spreader__[Unknown]->[Product]"
	pDef := &parser.ParserSpreaderDefinition{
		Key:          key,
		FromConcepts: []string{"Unknown"},
		ToConcepts:   []string{"Product"},
	}

	scope := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	_ = scope.DefineWithData(&key, symboltable.SpreaderScopeDefinition, symboltable.SourceLocation{}, pDef)

	err := AnalyzeSpreaderScope(scope, nil, analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown concept")
	assert.Contains(t, err.Error(), "Unknown")
}

func TestAnalyzeSpreaderScope_UnknownToConcept(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Product")

	key := "__spreader__[Product]->[Product,Unknown]"
	pDef := &parser.ParserSpreaderDefinition{
		Key:          key,
		FromConcepts: []string{"Product"},
		ToConcepts:   []string{"Product", "Unknown"},
	}

	scope := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	_ = scope.DefineWithData(&key, symboltable.SpreaderScopeDefinition, symboltable.SourceLocation{}, pDef)

	err := AnalyzeSpreaderScope(scope, nil, analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown concept")
	assert.Contains(t, err.Error(), "Unknown")
}

func TestAnalyzeSpreaderScope_InvalidDimensionalRules_TooManyNewDims(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product", "Status")

	key := "__spreader__[Product]->[Country,Product,Status]"
	pDef := &parser.ParserSpreaderDefinition{
		Key:          key,
		FromConcepts: []string{"Product"},
		ToConcepts:   []string{"Country", "Product", "Status"},
	}

	scope := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	_ = scope.DefineWithData(&key, symboltable.SpreaderScopeDefinition, symboltable.SourceLocation{}, pDef)

	err := AnalyzeSpreaderScope(scope, nil, analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one dimension")
}

func TestAnalyzeSpreaderScope_InvalidDimensionalRules_FromNotInTo(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product", "Status")

	key := "__spreader__[Product]->[Country,Status]"
	pDef := &parser.ParserSpreaderDefinition{
		Key:          key,
		FromConcepts: []string{"Product"},
		ToConcepts:   []string{"Country", "Status"},
	}

	scope := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	_ = scope.DefineWithData(&key, symboltable.SpreaderScopeDefinition, symboltable.SourceLocation{}, pDef)

	err := AnalyzeSpreaderScope(scope, nil, analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in to concepts")
	assert.Contains(t, err.Error(), "Product")
}

func TestAnalyzeSpreaderScope_UnknownIdentifierInExpression(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")

	key := "__spreader__[Product]->[Country,Product]"
	pDef := &parser.ParserSpreaderDefinition{
		Key:          key,
		FromConcepts: []string{"Product"},
		ToConcepts:   []string{"Country", "Product"},
	}

	scope := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	_ = scope.DefineWithData(&key, symboltable.SpreaderScopeDefinition, symboltable.SourceLocation{}, pDef)
	// Unknown identifier in expression
	scope.Definitions = append(scope.Definitions, makeSpreaderIdentifier("UnknownMetric"))

	err := AnalyzeSpreaderScope(scope, nil, analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown identifier")
	assert.Contains(t, err.Error(), "UnknownMetric")
}

func TestAnalyzeSpreaderScope_UnknownFunctionInExpression(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")

	key := "__spreader__[Product]->[Country,Product]"
	pDef := &parser.ParserSpreaderDefinition{
		Key:          key,
		FromConcepts: []string{"Product"},
		ToConcepts:   []string{"Country", "Product"},
	}

	scope := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	_ = scope.DefineWithData(&key, symboltable.SpreaderScopeDefinition, symboltable.SourceLocation{}, pDef)
	// Treat unknown function as identifier that is neither a known function nor a concept
	scope.Definitions = append(scope.Definitions, makeSpreaderIdentifier("unknownFunc"))

	err := AnalyzeSpreaderScope(scope, nil, analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown identifier")
	assert.Contains(t, err.Error(), "unknownFunc")
}

func TestAnalyzeSpreaderScope_DuplicatePattern(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product")

	// First spreader
	key1 := "__spreader__[Product]->[Country,Product]"
	pDef1 := &parser.ParserSpreaderDefinition{
		Key:            key1,
		FromConcepts:   []string{"Product"},
		ToConcepts:     []string{"Country", "Product"},
		FromEmpty:      false,
		SourceLocation: symboltable.SourceLocation{Line: 1, Column: 1},
	}
	scope1 := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	_ = scope1.DefineWithData(&key1, symboltable.SpreaderScopeDefinition, pDef1.SourceLocation, pDef1)
	scope1.Definitions = append(scope1.Definitions, makeSpreaderIdentifier("Country"))
	scope1.Definitions = append(scope1.Definitions, makeSpreaderIdentifier("1/count"))

	err := AnalyzeSpreaderScope(scope1, nil, analyzer)
	require.NoError(t, err)

	// Second spreader with same (from, to) pattern — should fail as duplicate
	key2 := "__spreader__[Product]->[Country,Product]"
	pDef2 := &parser.ParserSpreaderDefinition{
		Key:            key2,
		FromConcepts:   []string{"Product"},
		ToConcepts:     []string{"Country", "Product"},
		FromEmpty:      false,
		SourceLocation: symboltable.SourceLocation{Line: 10, Column: 5},
	}
	scope2 := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	_ = scope2.DefineWithData(&key2, symboltable.SpreaderScopeDefinition, pDef2.SourceLocation, pDef2)
	scope2.Definitions = append(scope2.Definitions, makeSpreaderIdentifier("Country"))
	scope2.Definitions = append(scope2.Definitions, makeSpreaderIdentifier("1/count"))

	err = AnalyzeSpreaderScope(scope2, nil, analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate spreader definition")
	assert.Contains(t, err.Error(), "line 1, column 1")  // first source location
	assert.Contains(t, err.Error(), "line 10, column 5") // second source location
}

func TestAnalyzeSpreaderScope_DifferentFromToOrderNotDuplicate(t *testing.T) {
	analyzer := makeTestAnalyzerWithConcepts("Country", "Product", "Status")

	// First spreader: from [Product] to [Country,Product]
	key1 := "__spreader__[Product]->[Country,Product]"
	pDef1 := &parser.ParserSpreaderDefinition{
		Key:            key1,
		FromConcepts:   []string{"Product"},
		ToConcepts:     []string{"Country", "Product"},
		FromEmpty:      false,
		SourceLocation: symboltable.SourceLocation{Line: 1, Column: 1},
	}
	scope1 := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	_ = scope1.DefineWithData(&key1, symboltable.SpreaderScopeDefinition, pDef1.SourceLocation, pDef1)
	scope1.Definitions = append(scope1.Definitions, makeSpreaderIdentifier("Country"))
	scope1.Definitions = append(scope1.Definitions, makeSpreaderIdentifier("1/count"))

	err := AnalyzeSpreaderScope(scope1, nil, analyzer)
	require.NoError(t, err)

	// Second spreader: from [Product] to [Status,Product] — different to, should succeed
	key2 := "__spreader__[Product]->[Status,Product]"
	pDef2 := &parser.ParserSpreaderDefinition{
		Key:            key2,
		FromConcepts:   []string{"Product"},
		ToConcepts:     []string{"Status", "Product"},
		FromEmpty:      false,
		SourceLocation: symboltable.SourceLocation{Line: 10, Column: 5},
	}
	scope2 := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	_ = scope2.DefineWithData(&key2, symboltable.SpreaderScopeDefinition, pDef2.SourceLocation, pDef2)
	scope2.Definitions = append(scope2.Definitions, makeSpreaderIdentifier("Status"))
	scope2.Definitions = append(scope2.Definitions, makeSpreaderIdentifier("1/count"))

	err = AnalyzeSpreaderScope(scope2, nil, analyzer)
	require.NoError(t, err)

	// Verify both spreaders are stored
	mapKey1 := SpreaderMapKey([]string{"Product"}, []string{"Country", "Product"})
	mapKey2 := SpreaderMapKey([]string{"Product"}, []string{"Status", "Product"})
	_, exists1 := analyzer.spreaders[mapKey1]
	_, exists2 := analyzer.spreaders[mapKey2]
	assert.True(t, exists1)
	assert.True(t, exists2)
}

// ============================================================================
// validateSpreaderDimUnits Tests
// ============================================================================

func TestValidateSpreaderDimUnits_Valid(t *testing.T) {
	analyzer := NewCPLAnalyzer()
	from := &DimensionalUnitExpression{Denominator: []string{"Product"}}
	expr := &DimensionalUnitExpression{Denominator: []string{"Country"}}
	to := &DimensionalUnitExpression{Denominator: []string{"Country", "Product"}}

	err := analyzer.validateSpreaderDimUnits(from, expr, to)
	assert.NoError(t, err)
}

func TestValidateSpreaderDimUnits_UnitlessValid(t *testing.T) {
	analyzer := NewCPLAnalyzer()
	from := &DimensionalUnitExpression{}
	expr := &DimensionalUnitExpression{Denominator: []string{"Country"}}
	to := &DimensionalUnitExpression{Denominator: []string{"Country"}}

	err := analyzer.validateSpreaderDimUnits(from, expr, to)
	assert.NoError(t, err)
}

func TestValidateSpreaderDimUnits_Mismatch(t *testing.T) {
	analyzer := NewCPLAnalyzer()
	from := &DimensionalUnitExpression{Denominator: []string{"Product"}}
	expr := &DimensionalUnitExpression{Numerator: []string{"Country"}}
	to := &DimensionalUnitExpression{Denominator: []string{"Country", "Product"}}

	err := analyzer.validateSpreaderDimUnits(from, expr, to)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dimensional unit mismatch")
	assert.Contains(t, err.Error(), "input")
	assert.Contains(t, err.Error(), "expression")
	assert.Contains(t, err.Error(), "expected")
}

// ============================================================================
// dimUnitsFromConcepts Tests
// ============================================================================

func TestDimUnitsFromConcepts_Empty(t *testing.T) {
	analyzer := NewCPLAnalyzer()
	result := analyzer.dimUnitsFromConcepts([]*AnalysisConceptDefinition{})
	require.NotNil(t, result)
	assert.True(t, result.IsUnitless())
}

func TestDimUnitsFromConcepts_SingleConcept(t *testing.T) {
	analyzer := NewCPLAnalyzer()
	result := analyzer.dimUnitsFromConcepts([]*AnalysisConceptDefinition{{ConceptName: "Country"}})
	require.NotNil(t, result)
	assert.Empty(t, result.Numerator)
	assert.Equal(t, []string{"Country"}, result.Denominator)
}

func TestDimUnitsFromConcepts_MultipleConcepts(t *testing.T) {
	analyzer := NewCPLAnalyzer()
	result := analyzer.dimUnitsFromConcepts([]*AnalysisConceptDefinition{
		{ConceptName: "Country"},
		{ConceptName: "Product"},
	})
	require.NotNil(t, result)
	assert.Empty(t, result.Numerator)
	assert.Equal(t, []string{"Country", "Product"}, result.Denominator)
}

// ============================================================================
// SpreaderMapKey Tests
// ============================================================================

func TestSpreaderMapKey(t *testing.T) {
	assert.Equal(t, "[A,B]->[C,D]", SpreaderMapKey([]string{"B", "A"}, []string{"D", "C"}))
	assert.Equal(t, "[Product]->[Country,Product]", SpreaderMapKey([]string{"Product"}, []string{"Country", "Product"}))
	assert.Equal(t, "[]->[Country]", SpreaderMapKey([]string{}, []string{"Country"}))
}

// ============================================================================
// BuildSpreadingMatrix Tests
// ============================================================================

func TestBuildSpreadingMatrix_EqProportional(t *testing.T) {
	values := []string{"USA", "Canada", "Mexico"}
	weights := map[string]float64{
		"USA":    1.0 / 3.0,
		"Canada": 1.0 / 3.0,
		"Mexico": 1.0 / 3.0,
	}

	matrix := BuildSpreadingMatrix(values, weights)

	require.NotNil(t, matrix)
	assert.Equal(t, 1, matrix.Rows)
	assert.Equal(t, 3, matrix.Columns)
	assert.Equal(t, []string{"source"}, matrix.RowLabels)
	assert.Equal(t, values, matrix.ColLabels)
	require.Len(t, matrix.Data, 1)
	require.Len(t, matrix.Data[0], 3)
	assert.InDelta(t, 1.0/3.0, matrix.Data[0][0], 1e-9)
	assert.InDelta(t, 1.0/3.0, matrix.Data[0][1], 1e-9)
	assert.InDelta(t, 1.0/3.0, matrix.Data[0][2], 1e-9)
}

func TestBuildSpreadingMatrix_SingleValue(t *testing.T) {
	values := []string{"Only"}
	weights := map[string]float64{"Only": 1.0}

	matrix := BuildSpreadingMatrix(values, weights)

	require.NotNil(t, matrix)
	assert.Equal(t, 1, matrix.Rows)
	assert.Equal(t, 1, matrix.Columns)
	assert.Equal(t, []float64{1.0}, matrix.Data[0])
}

// ============================================================================
// CILSpreadNode Tests
// ============================================================================

func TestCILSpreadNode_Structure(t *testing.T) {
	analyzer := NewCPLAnalyzer()
	country := NewValuesConceptDefinition("Country", []string{"USA", "Canada", "Mexico"})
	product := NewValuesConceptDefinition("Product", []string{"Widget", "Gadget"})
	analyzer.AddNewConcept(&country)
	analyzer.AddNewConcept(&product)

	key := "__spreader__[Product]->[Country,Product]"
	pDef := &parser.ParserSpreaderDefinition{
		Key:            key,
		FromConcepts:   []string{"Product"},
		ToConcepts:     []string{"Country", "Product"},
		FromEmpty:      false,
		SourceLocation: symboltable.SourceLocation{Line: 1, Column: 1},
	}

	scope := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	_ = scope.DefineWithData(&key, symboltable.SpreaderScopeDefinition, pDef.SourceLocation, pDef)
	// Valid expression: 1/count(Country) in postfix -> Country, 1/count
	scope.Definitions = append(scope.Definitions, makeSpreaderIdentifier("Country"))
	scope.Definitions = append(scope.Definitions, makeSpreaderIdentifier("1/count"))

	err := AnalyzeSpreaderScope(scope, nil, analyzer)
	require.NoError(t, err)

	// Verify the CIL node was stored
	cilTrees := analyzer.metricAnalyzer.GetCILTrees()
	mapKey := SpreaderMapKey([]string{"Product"}, []string{"Country", "Product"})
	cilKey := "spreader:" + mapKey
	require.Contains(t, cilTrees, cilKey)

	node, ok := cilTrees[cilKey].(*cil.CILSpreadNode)
	require.True(t, ok, "CIL node should be *cil.CILSpreadNode")

	// Verify FromDimensions (sorted)
	assert.Equal(t, []string{"Product"}, node.FromDimensions)
	// Verify ToDimensions (sorted)
	assert.Equal(t, []string{"Country", "Product"}, node.ToDimensions)
	// Verify NewDimension
	assert.Equal(t, "Country", node.NewDimension)
	// Verify matrix dimensions
	require.NotNil(t, node.SpreadingMatrix)
	assert.Equal(t, 1, node.SpreadingMatrix.Rows)
	assert.Equal(t, 3, node.SpreadingMatrix.Columns)
	// Verify matrix data is not nil
	require.NotNil(t, node.SpreadingMatrix.Data)
	// Verify row and column labels
	assert.Equal(t, []string{"source"}, node.SpreadingMatrix.RowLabels)
	assert.Equal(t, []string{"USA", "Canada", "Mexico"}, node.SpreadingMatrix.ColLabels)
	// Verify equal-proportion weights
	for j := 0; j < 3; j++ {
		assert.InDelta(t, 1.0/3.0, node.SpreadingMatrix.Data[0][j], 1e-9)
	}
	// Verify DimensionalUnitEffect
	assert.Equal(t, "1/Country", node.DimensionalUnitEffect)
	// Verify source location
	assert.Equal(t, 1, node.SourceLocation.Line)
	assert.Equal(t, 1, node.SourceLocation.Column)
	// Verify nil fields for Increment 1
	assert.Nil(t, node.DistributionExpression)
	assert.Nil(t, node.ReferencedMetrics)
	assert.Nil(t, node.SourceCIL)
}

func TestCILSpreadNode_InCILTrees(t *testing.T) {
	analyzer := NewCPLAnalyzer()
	country := NewValuesConceptDefinition("Country", []string{"USA", "Canada"})
	analyzer.AddNewConcept(&country)

	key := "__spreader__[]->[Country]"
	pDef := &parser.ParserSpreaderDefinition{
		Key:            key,
		FromConcepts:   []string{},
		ToConcepts:     []string{"Country"},
		FromEmpty:      true,
		SourceLocation: symboltable.SourceLocation{Line: 5, Column: 10},
	}

	scope := symboltable.NewSymbolTable(symboltable.SpreaderScope)
	_ = scope.DefineWithData(&key, symboltable.SpreaderScopeDefinition, pDef.SourceLocation, pDef)
	// Valid expression: 1/count(Country) -> gives 1/Country
	scope.Definitions = append(scope.Definitions, makeSpreaderIdentifier("Country"))
	scope.Definitions = append(scope.Definitions, makeSpreaderIdentifier("1/count"))

	err := AnalyzeSpreaderScope(scope, nil, analyzer)
	require.NoError(t, err)

	mapKey := SpreaderMapKey([]string{}, []string{"Country"})
	cilKey := "spreader:" + mapKey

	// Verify CIL node is retrievable from cilTrees
	cilTrees := analyzer.metricAnalyzer.GetCILTrees()
	require.Contains(t, cilTrees, cilKey)

	node, ok := cilTrees[cilKey].(*cil.CILSpreadNode)
	require.True(t, ok, "CIL node should be *cil.CILSpreadNode")
	assert.Equal(t, "Country", node.NewDimension)
	require.NotNil(t, node.SpreadingMatrix)
	assert.Equal(t, 1, node.SpreadingMatrix.Rows)
	assert.Equal(t, 2, node.SpreadingMatrix.Columns)
	assert.Equal(t, []string{"USA", "Canada"}, node.SpreadingMatrix.ColLabels)
	// Verify nil fields for Increment 1
	assert.Nil(t, node.DistributionExpression)
	assert.Nil(t, node.ReferencedMetrics)
	assert.Nil(t, node.SourceCIL)
}
