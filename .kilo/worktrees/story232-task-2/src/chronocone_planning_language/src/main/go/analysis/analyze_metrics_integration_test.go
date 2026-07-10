// Package analysis_test contains integration tests for metrics analysis.
//
// These tests verify that the metrics handler is correctly registered and invoked
// during the symbol table walking process. They test the full integration from
// parsing to analysis, ensuring metrics are properly analyzed when calling
// the top-level Analyze function.
//
// Story: Story182 - CPL Metrics Basic Parse and Analyze
// Task: Task 4c - Verify end-to-end parse-analyze pipeline
package analysis_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getMetricsExamplesDir returns the path to the examples directory.
func getMetricsExamplesDir(t *testing.T) string {
	t.Helper()
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	projectRoot := filepath.Join(testDir, "..", "..", "..", "..")
	return filepath.Join(projectRoot, "examples")
}

// ============================================================================
// Integration Tests - Basic Metrics File
// ============================================================================

// TestAnalyzeMetrics_BasicMetricsFile_EndToEnd verifies the end-to-end
// parse-analyze pipeline for the basic_metrics.cpl example file.
func TestAnalyzeMetrics_BasicMetricsFile_EndToEnd(t *testing.T) {
	t.Parallel()

	examplesDir := getMetricsExamplesDir(t)
	filePath := filepath.Join(examplesDir, "parse_and_analyze_only", "without_errors", "basic_metrics.cpl")

	_, err := os.Stat(filePath)
	require.NoError(t, err, "basic_metrics.cpl should exist")

	config := &parser.ParseLoggingConfig{
		LogLevel:   slog.LevelInfo,
		OutputFile: "",
	}
	rootTable, err := parser.ParseFiles([]string{filePath}, config)
	require.NoError(t, err, "Parsing basic_metrics.cpl should succeed")
	require.NotNil(t, rootTable, "Root table should not be nil")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed for basic_metrics.cpl")
	require.NotNil(t, defs, "AnalysisDefinitions should not be nil")

	// Verify 2 metrics were analyzed
	require.Len(t, defs.Metrics, 2, "Should have 2 metrics defined")

	// Verify revenue metric
	require.Contains(t, defs.Metrics, "revenue", "revenue metric should be analyzed")
	revenue := defs.Metrics["revenue"]
	assert.Equal(t, "revenue", revenue.Name)
	require.Len(t, revenue.Inputs, 2, "revenue should have 2 inputs")
	assert.Equal(t, "metric", revenue.Inputs[0].Kind)
	assert.Equal(t, "demand_forecast", revenue.Inputs[0].Name)
	assert.Equal(t, "dimension", revenue.Inputs[1].Kind)
	assert.Equal(t, "Price", revenue.Inputs[1].Name)

	// Verify revenue CILRoot
	require.NotNil(t, revenue.CILRoot, "revenue CILRoot should not be nil")
	revenueLiteral, ok := revenue.CILRoot.CILNode.(*cil.CILLiteral)
	require.True(t, ok, "revenue CILNode should be a CILLiteral")
	assert.Equal(t, float64(0), revenueLiteral.Value)

	// Verify profit metric
	require.Contains(t, defs.Metrics, "profit", "profit metric should be analyzed")
	profit := defs.Metrics["profit"]
	assert.Equal(t, "profit", profit.Name)
	require.Len(t, profit.Inputs, 2, "profit should have 2 inputs")
	assert.Equal(t, "metric", profit.Inputs[0].Kind)
	assert.Equal(t, "cost_forecast", profit.Inputs[0].Name)
	assert.Equal(t, "dimension", profit.Inputs[1].Kind)
	assert.Equal(t, "Cost", profit.Inputs[1].Name)

	// Verify profit CILRoot
	require.NotNil(t, profit.CILRoot, "profit CILRoot should not be nil")
	profitLiteral, ok := profit.CILRoot.CILNode.(*cil.CILLiteral)
	require.True(t, ok, "profit CILNode should be a CILLiteral")
	assert.Equal(t, float64(1), profitLiteral.Value)
}

// ============================================================================
// Integration Tests - Inline CPL with Measures
// ============================================================================

// TestAnalyzeMetrics_SingleMetric_WithInputs verifies a single metric
// with metric (measure) and dimension (concept) inputs.
func TestAnalyzeMetrics_SingleMetric_WithInputs(t *testing.T) {
	t.Parallel()

	input := `concepts
	Price = [Values[Low, Medium, High]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]

measures
	demand_forecast = Number(unit=Unit)
	Daily = Interval(unit=Day)

metrics
	revenue
		inputs
			metric demand_forecast
			dimension Price
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed")
	require.NotNil(t, defs, "AnalysisDefinitions should not be nil")

	require.Contains(t, defs.Metrics, "revenue")
	metric := defs.Metrics["revenue"]
	assert.Equal(t, "revenue", metric.Name)
	require.Len(t, metric.Inputs, 2)
	assert.Equal(t, "metric", metric.Inputs[0].Kind)
	assert.Equal(t, "demand_forecast", metric.Inputs[0].Name)
	assert.Equal(t, "dimension", metric.Inputs[1].Kind)
	assert.Equal(t, "Price", metric.Inputs[1].Name)

	require.NotNil(t, metric.CILRoot)
	literal, ok := metric.CILRoot.CILNode.(*cil.CILLiteral)
	require.True(t, ok, "CILNode should be CILLiteral")
	assert.Equal(t, float64(0), literal.Value)
}

// TestAnalyzeMetrics_TwoMetrics_TwoMeasures verifies two metrics
// referencing different measures.
func TestAnalyzeMetrics_TwoMetrics_TwoMeasures(t *testing.T) {
	t.Parallel()

	input := `concepts
	Price = [Values[Low, Medium, High]]
	Cost = [Values[Low, Medium, High]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]

measures
	demand_forecast = Number(unit=Unit)
	cost_forecast = Number(unit=Unit)
	Daily = Interval(unit=Day)

metrics
	revenue
		inputs
			metric demand_forecast
			dimension Price
		calculate
			0
	profit
		inputs
			metric cost_forecast
			dimension Cost
		calculate
			1`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed")
	require.Len(t, defs.Metrics, 2)

	require.Contains(t, defs.Metrics, "revenue")
	revenue := defs.Metrics["revenue"]
	assert.Equal(t, "revenue", revenue.Name)
	require.Len(t, revenue.Inputs, 2)
	assert.Equal(t, "demand_forecast", revenue.Inputs[0].Name)
	assert.Equal(t, "Price", revenue.Inputs[1].Name)

	require.Contains(t, defs.Metrics, "profit")
	profit := defs.Metrics["profit"]
	assert.Equal(t, "profit", profit.Name)
	require.Len(t, profit.Inputs, 2)
	assert.Equal(t, "cost_forecast", profit.Inputs[0].Name)
	assert.Equal(t, "Cost", profit.Inputs[1].Name)
}

// TestAnalyzeMetrics_MultipleMetrics_WithDifferentInputs verifies
// multiple metrics with different input combinations.
func TestAnalyzeMetrics_MultipleMetrics_WithDifferentInputs(t *testing.T) {
	t.Parallel()

	input := `concepts
	DimA = [Values[A1, A2]]
	DimB = [Values[B1, B2]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]

measures
	measure1 = Number(unit=Unit)
	measure2 = Number(unit=Unit)
	Daily = Interval(unit=Day)

metrics
	metric_a
		inputs
			metric measure1
		calculate
			0
	metric_b
		inputs
			dimension DimA
		calculate
			1
	metric_c
		inputs
			metric measure2
			dimension DimA
			dimension DimB
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed")
	require.Len(t, defs.Metrics, 3)

	require.Contains(t, defs.Metrics, "metric_a")
	assert.Len(t, defs.Metrics["metric_a"].Inputs, 1)

	require.Contains(t, defs.Metrics, "metric_b")
	assert.Len(t, defs.Metrics["metric_b"].Inputs, 1)

	require.Contains(t, defs.Metrics, "metric_c")
	assert.Len(t, defs.Metrics["metric_c"].Inputs, 3)
}

// TestAnalyzeMetrics_CILTreeForLiteralZero verifies CILLiteral(0) in CIL tree.
func TestAnalyzeMetrics_CILTreeForLiteralZero(t *testing.T) {
	t.Parallel()

	input := `concepts
	Dim = [Values[D1]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]

measures
	m = Number(unit=Unit)
	Daily = Interval(unit=Day)

metrics
	test_metric
		inputs
			metric m
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed")
	require.Contains(t, defs.Metrics, "test_metric")

	metric := defs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot)

	literal, ok := metric.CILRoot.CILNode.(*cil.CILLiteral)
	require.True(t, ok, "CILNode should be CILLiteral")
	assert.Equal(t, float64(0), literal.Value)
}

// TestAnalyzeMetrics_CILTreeForLiteralOne verifies CILLiteral(1) in CIL tree.
func TestAnalyzeMetrics_CILTreeForLiteralOne(t *testing.T) {
	t.Parallel()

	input := `concepts
	Dim = [Values[D1]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]

measures
	m = Number(unit=Unit)
	Daily = Interval(unit=Day)

metrics
	test_metric
		inputs
			metric m
		calculate
			1`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed")
	require.Contains(t, defs.Metrics, "test_metric")

	metric := defs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot)

	literal, ok := metric.CILRoot.CILNode.(*cil.CILLiteral)
	require.True(t, ok, "CILNode should be CILLiteral")
	assert.Equal(t, float64(1), literal.Value)
}

// ============================================================================
// Short Form Analysis Tests (Story193 - Increment 7)
// ============================================================================

// TestAnalyzeMetrics_ShortForm_LiteralExpression verifies analysis of a short form
// metric with a literal expression: my_metric = 42
func TestAnalyzeMetrics_ShortForm_LiteralExpression(t *testing.T) {
	t.Parallel()

	input := `concepts
	Price = [Values[Low, Medium, High]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]
measures
	MyMeasure = Number(unit=Unit)
	Daily = Interval(unit=Day)
metrics
	my_metric = 42`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")
	require.NotNil(t, rootTable)

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed")
	require.NotNil(t, defs)

	require.Contains(t, defs.Metrics, "my_metric")
	metric := defs.Metrics["my_metric"]
	assert.Equal(t, "my_metric", metric.Name)
	assert.Empty(t, metric.Inputs, "Short form metric with only literal should have no inputs")
	require.NotNil(t, metric.CILRoot)
	literal, ok := metric.CILRoot.CILNode.(*cil.CILLiteral)
	require.True(t, ok, "CILNode should be CILLiteral")
	assert.Equal(t, float64(42), literal.Value)
}

// TestAnalyzeMetrics_ShortForm_IdentifierAsMetricInput verifies analysis of a short form
// metric referencing another metric: my_metric = base_metric
func TestAnalyzeMetrics_ShortForm_IdentifierAsMetricInput(t *testing.T) {
	t.Parallel()

	input := `concepts
	Price = [Values[Low, High]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]
measures
	MyMeasure = Number(unit=Unit)
	Daily = Interval(unit=Day)
metrics
	base_metric
		inputs
			metric MyMeasure
		calculate
			10
	my_metric = base_metric`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed")

	require.Contains(t, defs.Metrics, "my_metric")
	metric := defs.Metrics["my_metric"]
	assert.Equal(t, "my_metric", metric.Name)
	require.Len(t, metric.Inputs, 1, "Should have 1 input (base_metric)")
	assert.Equal(t, "metric", metric.Inputs[0].Kind, "Input kind should be metric")
	assert.Equal(t, "base_metric", metric.Inputs[0].Name)
}

// TestAnalyzeMetrics_ShortForm_DimensionInput verifies analysis of a short form
// metric referencing a dimension: my_metric = Price
func TestAnalyzeMetrics_ShortForm_DimensionInput(t *testing.T) {
	t.Parallel()

	input := `concepts
	Price = [Values[Low, High]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]
measures
	MyMeasure = Number(unit=Unit)
	Daily = Interval(unit=Day)
metrics
	my_metric = Price`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed")

	require.Contains(t, defs.Metrics, "my_metric")
	metric := defs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 1, "Should have 1 input (Price)")
	assert.Equal(t, "dimension", metric.Inputs[0].Kind, "Input kind should be dimension")
	assert.Equal(t, "Price", metric.Inputs[0].Name)
}

// TestAnalyzeMetrics_ShortForm_ArithmeticMixedInputs verifies analysis of a short form
// metric with both metric and dimension references in arithmetic expression.
func TestAnalyzeMetrics_ShortForm_ArithmeticMixedInputs(t *testing.T) {
	t.Parallel()

	input := `concepts
	Price = [Values[Low, High]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]
measures
	MyMeasure = Number(unit=Unit)
	Daily = Interval(unit=Day)
metrics
	base_metric
		inputs
			metric MyMeasure
		calculate
			10
	my_metric = base_metric * Price + 5`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed")

	require.Contains(t, defs.Metrics, "my_metric")
	metric := defs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 2, "Should have 2 inputs (base_metric and Price)")

	// Inputs are collected deduplicated by name
	inputMap := make(map[string]string)
	for _, input := range metric.Inputs {
		inputMap[input.Name] = input.Kind
	}
	assert.Equal(t, "metric", inputMap["base_metric"], "base_metric should be metric kind")
	assert.Equal(t, "dimension", inputMap["Price"], "Price should be dimension kind")
}

// TestAnalyzeMetrics_ShortForm_UndefinedReference verifies that an undefined reference
// in a short form metric produces an analysis error.
func TestAnalyzeMetrics_ShortForm_UndefinedReference(t *testing.T) {
	t.Parallel()

	input := `concepts
	Price = [Values[Low, High]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]
measures
	MyMeasure = Number(unit=Unit)
	Daily = Interval(unit=Day)
metrics
	my_metric = undefined_name`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	_, err = analysis.Analyze(rootTable)
	require.Error(t, err, "Analysis should fail for undefined reference")
	assert.Contains(t, err.Error(), "undefined reference")
	assert.Contains(t, err.Error(), "undefined_name")
}

// TestAnalyzeMetrics_ShortForm_NameCollision verifies that a name matching both a metric
// and a dimension produces a collision error in short form analysis.
func TestAnalyzeMetrics_ShortForm_NameCollision(t *testing.T) {
	t.Parallel()

	input := `concepts
	Price = [Values[Low, High]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]
measures
	MyMeasure = Number(unit=Unit)
	Daily = Interval(unit=Day)
metrics
	Price
		inputs
			metric MyMeasure
		calculate
			10
	collision_metric = Price`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	_, err = analysis.Analyze(rootTable)
	require.Error(t, err, "Analysis should fail for name collision")
	assert.Contains(t, err.Error(), "collision")
	assert.Contains(t, err.Error(), "Price")
}

// TestAnalyzeMetrics_ShortForm_EquivalentToLongForm verifies that short form and
// equivalent long form produce identical CIL trees.
func TestAnalyzeMetrics_ShortForm_EquivalentToLongForm(t *testing.T) {
	t.Parallel()

	// Short form version: base_measure is a measure (leaf data source)
	shortInput := `concepts
	Price = [Values[Low, High]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]
measures
	MyMeasure = Number(unit=Unit)
	base_measure = Number(unit=Unit)
	Daily = Interval(unit=Day)
metrics
	revenue = base_measure * Price + 100`

	// Equivalent long form version
	longInput := `concepts
	Price = [Values[Low, High]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]
measures
	MyMeasure = Number(unit=Unit)
	base_measure = Number(unit=Unit)
	Daily = Interval(unit=Day)
metrics
	revenue
		inputs
			metric base_measure
			dimension Price
		calculate
			base_measure * Price + 100`

	shortTable, err := parser.Parse(shortInput)
	require.NoError(t, err, "Parsing short form should succeed")
	shortDefs, err := analysis.Analyze(shortTable)
	require.NoError(t, err, "Analysis of short form should succeed")

	longTable, err := parser.Parse(longInput)
	require.NoError(t, err, "Parsing long form should succeed")
	longDefs, err := analysis.Analyze(longTable)
	require.NoError(t, err, "Analysis of long form should succeed")

	// Both should have revenue metric
	require.Contains(t, shortDefs.Metrics, "revenue")
	require.Contains(t, longDefs.Metrics, "revenue")

	// Compare revenue metric inputs — short form and long form must produce
	// identical AnalysisInputDefinition structures (same names, same kinds).
	// Kind normalization in the analysis phase maps CIL-internal "connector"
	// to the user-facing "metric" keyword, matching long form.
	shortRevenue := shortDefs.Metrics["revenue"]
	longRevenue := longDefs.Metrics["revenue"]
	require.Len(t, shortRevenue.Inputs, 2, "Short form revenue should have 2 inputs")
	require.Len(t, longRevenue.Inputs, 2, "Long form revenue should have 2 inputs")

	// Build name→kind maps for order-independent comparison
	shortInputs := make(map[string]string)
	for _, in := range shortRevenue.Inputs {
		shortInputs[in.Name] = in.Kind
	}
	longInputs := make(map[string]string)
	for _, in := range longRevenue.Inputs {
		longInputs[in.Name] = in.Kind
	}
	assert.Equal(t, longInputs, shortInputs,
		"Short form and long form must produce identical input definitions (same names and kinds)")

	// Both should have CIL trees and they should be structurally identical
	require.NotNil(t, shortRevenue.CILRoot)
	require.NotNil(t, longRevenue.CILRoot)
	assertCILTreeEqual(t, shortRevenue.CILRoot.CILNode, longRevenue.CILRoot.CILNode,
		"Short form and long form must produce identical CIL trees")

	// InputRefs collected from the CIL tree must also match (after kind normalization)
	shortRefs := make(map[string]string)
	for _, ref := range shortRevenue.CILRoot.InputRefs {
		shortRefs[ref.Name] = ref.Kind
	}
	longRefs := make(map[string]string)
	for _, ref := range longRevenue.CILRoot.InputRefs {
		longRefs[ref.Name] = ref.Kind
	}
	assert.Equal(t, longRefs, shortRefs,
		"Short form and long form CIL InputRefs must match (after kind normalization)")
}

// assertCILTreeEqual recursively compares two CIL expression trees for structural
// equality. It checks that both trees have the same node types, operators, and
// operand values. This provides deep structural verification that short form
// and long form parsing produce identical CIL trees.
func assertCILTreeEqual(t *testing.T, expected, actual cil.CILNode, msg string) {
	t.Helper()

	if expected == nil && actual == nil {
		return
	}
	if expected == nil || actual == nil {
		t.Errorf("%s: one node is nil (expected type=%T, actual type=%T)", msg, expected, actual)
		return
	}

	switch expectedNode := expected.(type) {
	case *cil.CILLiteral:
		actualNode, ok := actual.(*cil.CILLiteral)
		if !ok {
			t.Errorf("%s: expected CILLiteral but got %T", msg, actual)
			return
		}
		assert.Equal(t, expectedNode.Value, actualNode.Value,
			"%s: CILLiteral values differ", msg)

	case *cil.CILInputRef:
		actualNode, ok := actual.(*cil.CILInputRef)
		if !ok {
			t.Errorf("%s: expected CILInputRef but got %T", msg, actual)
			return
		}
		assert.Equal(t, expectedNode.Kind, actualNode.Kind,
			"%s: CILInputRef kind differs for %q", msg, expectedNode.Name)
		assert.Equal(t, expectedNode.Name, actualNode.Name,
			"%s: CILInputRef name differs", msg)

	case *cil.CILBinaryOp:
		actualNode, ok := actual.(*cil.CILBinaryOp)
		if !ok {
			t.Errorf("%s: expected CILBinaryOp but got %T", msg, actual)
			return
		}
		assert.Equal(t, expectedNode.Op, actualNode.Op,
			"%s: CILBinaryOp operator differs", msg)
		assertCILTreeEqual(t, expectedNode.Left, actualNode.Left,
			msg+" (binary op "+expectedNode.Op.String()+" left)")
		assertCILTreeEqual(t, expectedNode.Right, actualNode.Right,
			msg+" (binary op "+expectedNode.Op.String()+" right)")

	case *cil.CILUnaryOp:
		actualNode, ok := actual.(*cil.CILUnaryOp)
		if !ok {
			t.Errorf("%s: expected CILUnaryOp but got %T", msg, actual)
			return
		}
		assert.Equal(t, expectedNode.Op, actualNode.Op,
			"%s: CILUnaryOp operator differs", msg)
		assertCILTreeEqual(t, expectedNode.Operand, actualNode.Operand,
			msg+" (unary op "+expectedNode.Op.String()+" operand)")

	default:
		t.Errorf("%s: unexpected CILNode type %T", msg, expected)
	}
}
