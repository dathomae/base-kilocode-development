// Package analysis_test contains integration tests for the metric dependency DAG,
// topological sort, cycle detection, and dead-code elimination through the full
// CPL parse-analyze pipeline.
//
// Story: Story194 - CPL Metrics DAG and IO Integration Analyze
// Task: Task 5 - Integration verification
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	cpio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// 5a) Multiple metrics with inter-metric dependencies
// ============================================================================

// TestDAGIntegration_MultipleMetricsWithDependencies parses CPL with multiple
// metrics that have inter-metric dependencies and verifies the DAG correctly
// captures all metric-to-metric dependencies.
func TestDAGIntegration_MultipleMetricsWithDependencies(t *testing.T) {
	t.Parallel()

	input := `concepts
	Price = [Values[Low, High]]
	Region = [Values[East, West]]
	Day = [Values[Monday]]

measures
	revenue = Number(unit=Day)
	Daily = Interval(unit=Day)

connectors
	SalesData = dimensions [Price, Region] grain Daily celltype revenue

metrics
	base_metric
		inputs
			metric revenue
			dimension Price
		calculate
			100
	revenue_forecast = base_metric * Price + 50
	profit_margin = revenue_forecast / base_metric - 10`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed")
	require.NotNil(t, defs)

	// Verify all 3 metrics were analyzed.
	require.Len(t, defs.Metrics, 3, "Should have 3 metrics")
	require.Contains(t, defs.Metrics, "base_metric")
	require.Contains(t, defs.Metrics, "revenue_forecast")
	require.Contains(t, defs.Metrics, "profit_margin")

	// Build DAG from analyzed metrics to verify dependencies.
	ma := analysis.NewMetricAnalyzer()
	for name, metric := range defs.Metrics {
		ma.AddMetric(name, metric)
	}
	for name, tree := range defs.CILTrees {
		ma.AddCILTree(name, tree)
	}

	err = ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	require.NotNil(t, dag)
	require.Len(t, dag.Nodes, 3)

	// base_metric has no metric dependencies (revenue is a measure, Price is a dimension).
	assert.Empty(t, dag.Nodes["base_metric"].Dependencies)

	// revenue_forecast depends on base_metric (via short-form expression).
	assert.Equal(t, []string{"base_metric"}, dag.Nodes["revenue_forecast"].Dependencies)

	// profit_margin depends on revenue_forecast and base_metric.
	assert.ElementsMatch(t, []string{"revenue_forecast", "base_metric"}, dag.Nodes["profit_margin"].Dependencies)

	// Verify root nodes: base_metric has no metric dependencies.
	assert.Contains(t, dag.RootNodes, "base_metric")

	// Verify topological order: dependencies before dependents.
	order, err := dag.TopologicalSort()
	require.NoError(t, err)
	require.Len(t, order, 3)

	baseIdx := indexOf(order, "base_metric")
	forecastIdx := indexOf(order, "revenue_forecast")
	marginIdx := indexOf(order, "profit_margin")

	assert.Less(t, baseIdx, forecastIdx, "base_metric should come before revenue_forecast")
	assert.Less(t, forecastIdx, marginIdx, "revenue_forecast should come before profit_margin")
	assert.Less(t, baseIdx, marginIdx, "base_metric should come before profit_margin")
}

// ============================================================================
// 5b & 5c) IO spec with output metric selection triggers dead-code elimination
// ============================================================================

// TestDAGIntegration_DeadCodeEliminationWithIOSpec verifies that when an IO spec
// only outputs profit_margin, only reachable metrics survive dead-code elimination
// in the full analysis pipeline.
func TestDAGIntegration_DeadCodeEliminationWithIOSpec(t *testing.T) {
	t.Parallel()

	input := `concepts
	Price = [Values[Low, High]]
	Region = [Values[East, West]]
	Day = [Values[Monday]]

measures
	revenue = Number(unit=Day)
	Daily = Interval(unit=Day)

connectors
	SalesData = dimensions [Price, Region] grain Daily celltype revenue

metrics
	base_metric
		inputs
			metric revenue
			dimension Price
		calculate
			100
	revenue_forecast = base_metric * Price + 50
	profit_margin = revenue_forecast / base_metric - 10`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{MetricName: "profit_margin"},
		},
	}

	defs, err := analysis.AnalyzeWithIO(rootTable, ioSpec)
	require.NoError(t, err, "Analysis with IO spec should succeed")
	require.NotNil(t, defs)

	// Verify only reachable metrics remain after dead-code elimination.
	require.Len(t, defs.Metrics, 3, "Only profit_margin, revenue_forecast, and base_metric should be reachable")
	assert.Contains(t, defs.Metrics, "profit_margin")
	assert.Contains(t, defs.Metrics, "revenue_forecast")
	assert.Contains(t, defs.Metrics, "base_metric")

	// Verify CIL trees for surviving metrics are present via metric definitions.
	assert.NotNil(t, defs.Metrics["profit_margin"].CILRoot)
	assert.NotNil(t, defs.Metrics["revenue_forecast"].CILRoot)
	assert.NotNil(t, defs.Metrics["base_metric"].CILRoot)

	// Verify DAG is returned with valid topological order after pruning.
	require.NotNil(t, defs.DAG, "DAG should be present after analysis")
	require.Len(t, defs.DAG.Nodes, 3)
	require.NotNil(t, defs.DAG.Order, "DAG.Order should be populated after dead-code elimination")
	require.Len(t, defs.DAG.Order, 3)

	// Verify topological order: dependencies before dependents.
	baseIdx := indexOf(defs.DAG.Order, "base_metric")
	forecastIdx := indexOf(defs.DAG.Order, "revenue_forecast")
	marginIdx := indexOf(defs.DAG.Order, "profit_margin")

	assert.Less(t, baseIdx, forecastIdx)
	assert.Less(t, forecastIdx, marginIdx)
}

// ============================================================================
// 5d) Edge cases
// ============================================================================

// TestDAGIntegration_EmptyIOSpec verifies that when the IO spec has no outputs,
// all metrics are eliminated as dead code.
func TestDAGIntegration_EmptyIOSpec(t *testing.T) {
	t.Parallel()

	input := `concepts
	Price = [Values[Low, High]]
	Day = [Values[Monday]]

measures
	m = Number(unit=Day)
	Daily = Interval(unit=Day)

connectors
	SalesData = dimensions [Price] grain Daily celltype m

metrics
	metric_a
		inputs
			metric m
			dimension Price
		calculate
			0
	metric_b = metric_a + 1`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{},
	}

	defs, err := analysis.AnalyzeWithIO(rootTable, ioSpec)
	require.NoError(t, err, "Analysis with empty IO spec should succeed")
	require.NotNil(t, defs)

	// All metrics are dead when no outputs are specified.
	assert.Empty(t, defs.Metrics, "All metrics should be eliminated with empty IO spec outputs")
}

// TestDAGIntegration_MultiLevelDependencyChain verifies a chain A -> B -> C -> D
// produces the correct topological order and all intermediate metrics are reachable.
func TestDAGIntegration_MultiLevelDependencyChain(t *testing.T) {
	t.Parallel()

	input := `concepts
	Dim = [Values[D1]]
	Day = [Values[Monday]]

measures
	m = Number(unit=Day)
	Daily = Interval(unit=Day)

connectors
	Data = dimensions [Dim] grain Daily celltype m

metrics
	metric_d
		inputs
			metric m
			dimension Dim
		calculate
			0
	metric_c = metric_d + 1
	metric_b = metric_c + 1
	metric_a = metric_b + 1`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed")
	require.Len(t, defs.Metrics, 4)

	// Build DAG and verify chain.
	ma := analysis.NewMetricAnalyzer()
	for name, metric := range defs.Metrics {
		ma.AddMetric(name, metric)
	}
	for name, tree := range defs.CILTrees {
		ma.AddCILTree(name, tree)
	}

	err = ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	require.Len(t, dag.Nodes, 4)

	// Verify dependency chain.
	assert.Empty(t, dag.Nodes["metric_d"].Dependencies)
	assert.Equal(t, []string{"metric_d"}, dag.Nodes["metric_c"].Dependencies)
	assert.Equal(t, []string{"metric_c"}, dag.Nodes["metric_b"].Dependencies)
	assert.Equal(t, []string{"metric_b"}, dag.Nodes["metric_a"].Dependencies)

	// Verify topological order.
	order, err := dag.TopologicalSort()
	require.NoError(t, err)
	require.Len(t, order, 4)
	assert.Equal(t, []string{"metric_d", "metric_c", "metric_b", "metric_a"}, order)

	// Verify all are reachable from metric_a.
	reachable := dag.MarkReachable([]string{"metric_a"})
	assert.True(t, reachable["metric_a"])
	assert.True(t, reachable["metric_b"])
	assert.True(t, reachable["metric_c"])
	assert.True(t, reachable["metric_d"])
	assert.Len(t, reachable, 4)
}

// TestDAGIntegration_CircularDependencyDetected verifies that circular
// dependencies between metrics are detected and reported as an error when
// building the DAG.
func TestDAGIntegration_CircularDependencyDetected(t *testing.T) {
	t.Parallel()

	// Circular dependencies cannot be expressed through the CPL parser because
	// metric references are validated at analysis time. We construct the
	// metrics manually and build the DAG to verify cycle detection.
	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name: "A",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "B"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "B"}},
		},
	})
	ma.AddMetric("B", analysis.AnalysisMetricDefinition{
		Name: "B",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "A"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "A"}},
		},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	err = ma.GetDAG().DetectCycles(ma.GetMetrics())
	require.Error(t, err, "Circular dependency should be detected")
	assert.Contains(t, err.Error(), "circular dependency")
}

// TestDAGIntegration_TimeShiftedSelfReferenceAllowed verifies that a metric
// with a backward(day) self-reference passes cycle detection in the DAG.
func TestDAGIntegration_TimeShiftedSelfReferenceAllowed(t *testing.T) {
	t.Parallel()

	input := `concepts
	Dim = [Values[D1]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]

measures
	Daily = Interval(unit=Day)

metrics
	revenue
		inputs
			metric revenue backward(day)
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Full analysis should succeed because time-shifted self-reference is permitted.
	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed for time-shifted self-reference")
	require.NotNil(t, defs)
	require.Contains(t, defs.Metrics, "revenue")

	// Build DAG and verify no cycle is detected.
	ma := analysis.NewMetricAnalyzer()
	for name, metric := range defs.Metrics {
		ma.AddMetric(name, metric)
	}
	for name, tree := range defs.CILTrees {
		ma.AddCILTree(name, tree)
	}

	err = ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	// The DAG should have a single node with no dependencies (self-reference
	// with time shift is excluded from the same-period dependency DAG).
	dag := ma.GetDAG()
	require.Len(t, dag.Nodes, 1)
	assert.Empty(t, dag.Nodes["revenue"].Dependencies)

	err = dag.DetectCycles(ma.GetMetrics())
	require.NoError(t, err, "Time-shifted self-reference should not be detected as a cycle")

	// Topological sort should succeed.
	order, err := dag.TopologicalSort()
	require.NoError(t, err)
	assert.Equal(t, []string{"revenue"}, order)
}

// ============================================================================
// Helpers
// ============================================================================

// indexOf returns the index of s in slice, or -1 if not found.
func indexOf(slice []string, s string) int {
	for i, v := range slice {
		if v == s {
			return i
		}
	}
	return -1
}
