// Package analysis_test contains tests for the metric dependency DAG.
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
// DAG Data Structure Tests
// ============================================================================

// TestMetricDAG_StructFields verifies that MetricDAG and DAGNode structs have
// the expected fields and can be initialized.
func TestMetricDAG_StructFields(t *testing.T) {
	t.Parallel()

	dag := &analysis.MetricDAG{
		Nodes: map[string]*analysis.DAGNode{
			"m1": {
				MetricName:   "m1",
				Dependencies: []string{"m2"},
				Dependents:   []string{"m3"},
				Visited:      true,
				InProgress:   false,
			},
		},
		RootNodes: []string{"m1"},
		Order:     []string{"m2", "m1"},
	}

	require.NotNil(t, dag)
	require.Contains(t, dag.Nodes, "m1")
	node := dag.Nodes["m1"]
	assert.Equal(t, "m1", node.MetricName)
	assert.Equal(t, []string{"m2"}, node.Dependencies)
	assert.Equal(t, []string{"m3"}, node.Dependents)
	assert.True(t, node.Visited)
	assert.False(t, node.InProgress)
	assert.Equal(t, []string{"m1"}, dag.RootNodes)
	assert.Equal(t, []string{"m2", "m1"}, dag.Order)
}

// ============================================================================
// CollectMetricDependencies Tests
// ============================================================================

// TestCollectMetricDependencies_NilMetric returns nil for nil metric.
func TestCollectMetricDependencies_NilMetric(t *testing.T) {
	t.Parallel()

	deps := analysis.CollectMetricDependencies(nil)
	assert.Nil(t, deps)
}

// TestCollectMetricDependencies_NilCILRoot returns nil when CILRoot is nil.
func TestCollectMetricDependencies_NilCILRoot(t *testing.T) {
	t.Parallel()

	metric := analysis.AnalysisMetricDefinition{
		Name:   "m1",
		Inputs: []analysis.AnalysisInputDefinition{},
	}
	deps := analysis.CollectMetricDependencies(&metric)
	assert.Nil(t, deps)
}

// TestCollectMetricDependencies_EmptyInputRefs returns nil when InputRefs is empty.
func TestCollectMetricDependencies_EmptyInputRefs(t *testing.T) {
	t.Parallel()

	metric := analysis.AnalysisMetricDefinition{
		Name: "m1",
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{},
		},
	}
	deps := analysis.CollectMetricDependencies(&metric)
	assert.Nil(t, deps)
}

// TestCollectMetricDependencies_FiltersMetricKind returns only metric-type refs.
func TestCollectMetricDependencies_FiltersMetricKind(t *testing.T) {
	t.Parallel()

	metric := analysis.AnalysisMetricDefinition{
		Name: "m1",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{
				{Kind: "metric", Name: "dep1"},
				{Kind: "dimension", Name: "Dim1"},
				{Kind: "metric", Name: "dep2"},
			},
		},
	}
	deps := analysis.CollectMetricDependencies(&metric)
	assert.Equal(t, []string{"dep1", "dep2"}, deps)
}

// TestCollectMetricDependencies_Deduplicates returns unique metric names only.
func TestCollectMetricDependencies_Deduplicates(t *testing.T) {
	t.Parallel()

	metric := analysis.AnalysisMetricDefinition{
		Name: "m1",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{
				{Kind: "metric", Name: "dep1"},
				{Kind: "metric", Name: "dep1"},
				{Kind: "metric", Name: "dep2"},
			},
		},
	}
	deps := analysis.CollectMetricDependencies(&metric)
	assert.Equal(t, []string{"dep1", "dep2"}, deps)
}

// ============================================================================
// BuildMetricDependencyDAG Tests
// ============================================================================

// TestBuildMetricDependencyDAG_SingleMetricNoDeps verifies a single metric with
// no dependencies produces a DAG with one node and one root node.
func TestBuildMetricDependencyDAG_SingleMetricNoDeps(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("m1", analysis.AnalysisMetricDefinition{
		Name:    "m1",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	require.NotNil(t, dag)
	require.Len(t, dag.Nodes, 1)
	require.Contains(t, dag.Nodes, "m1")
	assert.Empty(t, dag.Nodes["m1"].Dependencies)
	assert.Empty(t, dag.Nodes["m1"].Dependents)
	assert.Equal(t, []string{"m1"}, dag.RootNodes)
}

// TestBuildMetricDependencyDAG_TwoMetricsChain verifies A depends on B.
func TestBuildMetricDependencyDAG_TwoMetricsChain(t *testing.T) {
	t.Parallel()

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
		Name:    "B",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	require.Len(t, dag.Nodes, 2)

	// A depends on B
	assert.Equal(t, []string{"B"}, dag.Nodes["A"].Dependencies)
	assert.Empty(t, dag.Nodes["A"].Dependents)

	// B has no dependencies, but A depends on it
	assert.Empty(t, dag.Nodes["B"].Dependencies)
	assert.Equal(t, []string{"A"}, dag.Nodes["B"].Dependents)

	// Root node is B
	assert.Equal(t, []string{"B"}, dag.RootNodes)
}

// TestBuildMetricDependencyDAG_MultiLevelChain verifies A -> B -> C chain.
func TestBuildMetricDependencyDAG_MultiLevelChain(t *testing.T) {
	t.Parallel()

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
			{Kind: "metric", Name: "C"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "C"}},
		},
	})
	ma.AddMetric("C", analysis.AnalysisMetricDefinition{
		Name:    "C",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	require.Len(t, dag.Nodes, 3)

	assert.Equal(t, []string{"B"}, dag.Nodes["A"].Dependencies)
	assert.Empty(t, dag.Nodes["A"].Dependents)

	assert.Equal(t, []string{"C"}, dag.Nodes["B"].Dependencies)
	assert.Equal(t, []string{"A"}, dag.Nodes["B"].Dependents)

	assert.Empty(t, dag.Nodes["C"].Dependencies)
	assert.Equal(t, []string{"B"}, dag.Nodes["C"].Dependents)

	assert.Equal(t, []string{"C"}, dag.RootNodes)
}

// TestBuildMetricDependencyDAG_IndependentMetrics verifies no edges between
// metrics that do not depend on each other.
func TestBuildMetricDependencyDAG_IndependentMetrics(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("m1", analysis.AnalysisMetricDefinition{
		Name:    "m1",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddMetric("m2", analysis.AnalysisMetricDefinition{
		Name:    "m2",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	require.Len(t, dag.Nodes, 2)

	assert.Empty(t, dag.Nodes["m1"].Dependencies)
	assert.Empty(t, dag.Nodes["m1"].Dependents)
	assert.Empty(t, dag.Nodes["m2"].Dependencies)
	assert.Empty(t, dag.Nodes["m2"].Dependents)

	// Root nodes should be sorted deterministically.
	require.Len(t, dag.RootNodes, 2)
	assert.Equal(t, []string{"m1", "m2"}, dag.RootNodes)
}

// TestBuildMetricDependencyDAG_DiamondPattern verifies A depends on B and C,
// and B and C both depend on D.
func TestBuildMetricDependencyDAG_DiamondPattern(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name: "A",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "B"},
			{Kind: "metric", Name: "C"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{
				{Kind: "metric", Name: "B"},
				{Kind: "metric", Name: "C"},
			},
		},
	})
	ma.AddMetric("B", analysis.AnalysisMetricDefinition{
		Name: "B",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "D"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "D"}},
		},
	})
	ma.AddMetric("C", analysis.AnalysisMetricDefinition{
		Name: "C",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "D"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "D"}},
		},
	})
	ma.AddMetric("D", analysis.AnalysisMetricDefinition{
		Name:    "D",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	require.Len(t, dag.Nodes, 4)

	// A depends on B and C
	assert.ElementsMatch(t, []string{"B", "C"}, dag.Nodes["A"].Dependencies)
	assert.Empty(t, dag.Nodes["A"].Dependents)

	// B depends on D, A depends on B
	assert.Equal(t, []string{"D"}, dag.Nodes["B"].Dependencies)
	assert.Equal(t, []string{"A"}, dag.Nodes["B"].Dependents)

	// C depends on D, A depends on C
	assert.Equal(t, []string{"D"}, dag.Nodes["C"].Dependencies)
	assert.Equal(t, []string{"A"}, dag.Nodes["C"].Dependents)

	// D has no dependencies, B and C depend on D
	assert.Empty(t, dag.Nodes["D"].Dependencies)
	assert.ElementsMatch(t, []string{"B", "C"}, dag.Nodes["D"].Dependents)

	assert.Equal(t, []string{"D"}, dag.RootNodes)
}

// TestBuildMetricDependencyDAG_RootNodesCorrectlyIdentified verifies that root
// nodes are metrics with zero dependencies.
func TestBuildMetricDependencyDAG_RootNodesCorrectlyIdentified(t *testing.T) {
	t.Parallel()

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
		Name:    "B",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddMetric("C", analysis.AnalysisMetricDefinition{
		Name:    "C",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	require.Len(t, dag.RootNodes, 2)
	assert.Equal(t, []string{"B", "C"}, dag.RootNodes)
	assert.NotContains(t, dag.RootNodes, "A")
}

// TestBuildMetricDependencyDAG_DependentsCorrectlyIdentified verifies that
// Dependents lists are populated correctly.
func TestBuildMetricDependencyDAG_DependentsCorrectlyIdentified(t *testing.T) {
	t.Parallel()

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
		Name:    "B",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	assert.Equal(t, []string{"A"}, dag.Nodes["B"].Dependents)
	assert.Empty(t, dag.Nodes["A"].Dependents)
}

// TestBuildMetricDependencyDAG_UnknownDependency skips non-metric dependencies
// (e.g., measures or connectors) without error.
func TestBuildMetricDependencyDAG_UnknownDependency(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name: "A",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "Unknown"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{},
		},
	})

	// Should succeed: "Unknown" is not a metric, so it's skipped (treated as measure/connector ref)
	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	assert.Empty(t, dag.Nodes["A"].Dependencies)
}

// ============================================================================
// TopologicalSort Tests
// ============================================================================

// TestTopologicalSort_LinearChain verifies topological ordering of A→B→C.
func TestTopologicalSort_LinearChain(t *testing.T) {
	t.Parallel()

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
			{Kind: "metric", Name: "C"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "C"}},
		},
	})
	ma.AddMetric("C", analysis.AnalysisMetricDefinition{
		Name:    "C",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	order, err := ma.GetDAG().TopologicalSort()
	require.NoError(t, err)
	assert.Equal(t, []string{"C", "B", "A"}, order)
}

// TestTopologicalSort_Diamond verifies topological ordering of diamond DAG.
func TestTopologicalSort_Diamond(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name: "A",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "B"},
			{Kind: "metric", Name: "D"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{
				{Kind: "metric", Name: "B"},
				{Kind: "metric", Name: "D"},
			},
		},
	})
	ma.AddMetric("B", analysis.AnalysisMetricDefinition{
		Name: "B",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "C"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "C"}},
		},
	})
	ma.AddMetric("C", analysis.AnalysisMetricDefinition{
		Name:    "C",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddMetric("D", analysis.AnalysisMetricDefinition{
		Name: "D",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "C"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "C"}},
		},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	order, err := ma.GetDAG().TopologicalSort()
	require.NoError(t, err)

	// C must come first, A must come last. B and D order is deterministic (B before D).
	assert.Equal(t, []string{"C", "B", "D", "A"}, order)
}

// TestTopologicalSort_IndependentMetrics verifies ordering of metrics with no dependencies.
func TestTopologicalSort_IndependentMetrics(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("m1", analysis.AnalysisMetricDefinition{
		Name:    "m1",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddMetric("m2", analysis.AnalysisMetricDefinition{
		Name:    "m2",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	order, err := ma.GetDAG().TopologicalSort()
	require.NoError(t, err)
	assert.Equal(t, []string{"m1", "m2"}, order)
}

// TestTopologicalSort_SingleMetric verifies ordering of a single metric.
func TestTopologicalSort_SingleMetric(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("m1", analysis.AnalysisMetricDefinition{
		Name:    "m1",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	order, err := ma.GetDAG().TopologicalSort()
	require.NoError(t, err)
	assert.Equal(t, []string{"m1"}, order)
}

// TestTopologicalSort_CycleError verifies that topological sort returns an error for a cyclic DAG.
func TestTopologicalSort_CycleError(t *testing.T) {
	t.Parallel()

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

	_, err = ma.GetDAG().TopologicalSort()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
}

// ============================================================================
// DetectCycles Tests
// ============================================================================

// TestDetectCycles_SimpleCycle detects A→B→A.
func TestDetectCycles_SimpleCycle(t *testing.T) {
	t.Parallel()

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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

// TestDetectCycles_MultiNodeCycle detects A→B→C→A.
func TestDetectCycles_MultiNodeCycle(t *testing.T) {
	t.Parallel()

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
			{Kind: "metric", Name: "C"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "C"}},
		},
	})
	ma.AddMetric("C", analysis.AnalysisMetricDefinition{
		Name: "C",
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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

// TestDetectCycles_NoCycle verifies no error on acyclic DAG.
func TestDetectCycles_NoCycle(t *testing.T) {
	t.Parallel()

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
		Name:    "B",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	err = ma.GetDAG().DetectCycles(ma.GetMetrics())
	require.NoError(t, err)
}

// TestDetectCycles_SelfLoopWithoutTimeShift rejects a single-node self-loop
// that does not have a time-shifted input.
func TestDetectCycles_SelfLoopWithoutTimeShift(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name: "A",
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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

// TestDetectCycles_TimeShiftedSelfReference permits A→A with backward shift.
func TestDetectCycles_TimeShiftedSelfReference(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name: "A",
		Inputs: []analysis.AnalysisInputDefinition{
			{
				Kind:      "metric",
				Name:      "A",
				TimeShift: &analysis.AnalysisTimeShift{Direction: "backward", Field: "day", Amount: 1},
			},
		},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	err = ma.GetDAG().DetectCycles(ma.GetMetrics())
	require.NoError(t, err)
}

// TestDetectCycles_MixedSelfReferenceAndCycle detects B cycle even when A has permitted self-reference.
func TestDetectCycles_MixedSelfReferenceAndCycle(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name: "A",
		Inputs: []analysis.AnalysisInputDefinition{
			{
				Kind:      "metric",
				Name:      "A",
				TimeShift: &analysis.AnalysisTimeShift{Direction: "backward", Field: "day", Amount: 1},
			},
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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

// TestDetectCycles_TimeShiftedSelfReferenceWithNonShiftedEdge errors when self-reference
// is permitted but another edge in the cycle is not time-shifted.
func TestDetectCycles_TimeShiftedSelfReferenceWithNonShiftedEdge(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name: "A",
		Inputs: []analysis.AnalysisInputDefinition{
			{
				Kind:      "metric",
				Name:      "A",
				TimeShift: &analysis.AnalysisTimeShift{Direction: "backward", Field: "day", Amount: 1},
			},
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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}


// TestBuildMetricDependencyDAG_InputsOnly verifies DAG is built from Inputs
// even when CILRoot has no metric InputRefs.
func TestBuildMetricDependencyDAG_InputsOnly(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name: "A",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "B"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{},
		},
	})
	ma.AddMetric("B", analysis.AnalysisMetricDefinition{
		Name:    "B",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	assert.Equal(t, []string{"B"}, dag.Nodes["A"].Dependencies)
}

// TestBuildMetricDependencyDAG_CILRefsOnly verifies DAG is built from CIL
// InputRefs even when Inputs has no metric entries.
func TestBuildMetricDependencyDAG_CILRefsOnly(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name:   "A",
		Inputs: []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "B"}},
		},
	})
	ma.AddMetric("B", analysis.AnalysisMetricDefinition{
		Name:    "B",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	assert.Equal(t, []string{"B"}, dag.Nodes["A"].Dependencies)
}

// ============================================================================
// MarkReachable Tests
// ============================================================================

// TestMarkReachable_AllReachable verifies all nodes are reachable when
// starting from all output metrics.
func TestMarkReachable_AllReachable(t *testing.T) {
	t.Parallel()

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
		Name:    "B",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	reachable := dag.MarkReachable([]string{"A"})

	assert.True(t, reachable["A"])
	assert.True(t, reachable["B"])
	assert.Len(t, reachable, 2)
}

// TestMarkReachable_SubsetReachable verifies that only dependencies of the
// output metric are reachable; unrelated metrics are not.
func TestMarkReachable_SubsetReachable(t *testing.T) {
	t.Parallel()

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
		Name:    "B",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddMetric("C", analysis.AnalysisMetricDefinition{
		Name:    "C",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	reachable := dag.MarkReachable([]string{"A"})

	assert.True(t, reachable["A"])
	assert.True(t, reachable["B"])
	assert.False(t, reachable["C"])
	assert.Len(t, reachable, 2)
}

// TestMarkReachable_EmptyOutputList verifies that an empty output list
// produces an empty reachable set.
func TestMarkReachable_EmptyOutputList(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name:    "A",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddMetric("B", analysis.AnalysisMetricDefinition{
		Name:    "B",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	reachable := dag.MarkReachable([]string{})

	assert.Len(t, reachable, 0)
}

// TestMarkReachable_MultiLevelChain verifies that all ancestors in a
// multi-level dependency chain are reachable.
func TestMarkReachable_MultiLevelChain(t *testing.T) {
	t.Parallel()

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
			{Kind: "metric", Name: "C"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "C"}},
		},
	})
	ma.AddMetric("C", analysis.AnalysisMetricDefinition{
		Name:    "C",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	reachable := dag.MarkReachable([]string{"A"})

	assert.True(t, reachable["A"])
	assert.True(t, reachable["B"])
	assert.True(t, reachable["C"])
	assert.Len(t, reachable, 3)
}

// TestMarkReachable_DiamondPattern verifies that both paths to a common
// dependency are traversed.
func TestMarkReachable_DiamondPattern(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name: "A",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "B"},
			{Kind: "metric", Name: "C"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{
				{Kind: "metric", Name: "B"},
				{Kind: "metric", Name: "C"},
			},
		},
	})
	ma.AddMetric("B", analysis.AnalysisMetricDefinition{
		Name: "B",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "D"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "D"}},
		},
	})
	ma.AddMetric("C", analysis.AnalysisMetricDefinition{
		Name: "C",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "D"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "D"}},
		},
	})
	ma.AddMetric("D", analysis.AnalysisMetricDefinition{
		Name:    "D",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	reachable := dag.MarkReachable([]string{"A"})

	assert.True(t, reachable["A"])
	assert.True(t, reachable["B"])
	assert.True(t, reachable["C"])
	assert.True(t, reachable["D"])
	assert.Len(t, reachable, 4)
}

// TestMarkReachable_UnknownOutputMetric is handled gracefully — skipped
// when the name doesn't exist in the DAG.
func TestMarkReachable_UnknownOutputMetric(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name:    "A",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	// "Unknown" does not exist in the DAG — should be silently skipped.
	reachable := dag.MarkReachable([]string{"Unknown", "A"})

	assert.True(t, reachable["A"])
	assert.False(t, reachable["Unknown"])
	assert.Len(t, reachable, 1)
}

// TestMarkReachable_MultipleOutputs verifies that starting from multiple
// output metrics merges their reachable sets.
func TestMarkReachable_MultipleOutputs(t *testing.T) {
	t.Parallel()

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
		Name:    "B",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddMetric("X", analysis.AnalysisMetricDefinition{
		Name: "X",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "Y"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "Y"}},
		},
	})
	ma.AddMetric("Y", analysis.AnalysisMetricDefinition{
		Name:    "Y",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	reachable := dag.MarkReachable([]string{"A", "X"})

	assert.True(t, reachable["A"])
	assert.True(t, reachable["B"])
	assert.True(t, reachable["X"])
	assert.True(t, reachable["Y"])
	assert.Len(t, reachable, 4)
}

// TestMarkReachable_EmptyDAG verifies that MarkReachable returns an empty set
// when the DAG has no nodes.
func TestMarkReachable_EmptyDAG(t *testing.T) {
	t.Parallel()

	dag := &analysis.MetricDAG{
		Nodes: map[string]*analysis.DAGNode{},
	}
	reachable := dag.MarkReachable([]string{"A", "B"})

	assert.Empty(t, reachable)
}

// ============================================================================
// EliminateDeadCode Tests
// ============================================================================

// TestEliminateDeadCode_AllReachable verifies no metrics are pruned when
// all metrics are reachable from outputs.
func TestEliminateDeadCode_AllReachable(t *testing.T) {
	t.Parallel()

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
		Name:    "B",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddCILTree("A", &cil.CILLiteral{Value: 0})
	ma.AddCILTree("B", &cil.CILLiteral{Value: 0})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	err = ma.EliminateDeadCode([]string{"A"})
	require.NoError(t, err)

	assert.True(t, ma.HasMetric("A"))
	assert.True(t, ma.HasMetric("B"))
	assert.True(t, ma.HasCILTree("A"))
	assert.True(t, ma.HasCILTree("B"))
	assert.Len(t, ma.GetMetrics(), 2)
	assert.Len(t, ma.GetCILTrees(), 2)
}

// TestEliminateDeadCode_SubsetPruned verifies that unreachable metrics are
// removed from both ma.metrics and ma.cilTrees.
func TestEliminateDeadCode_SubsetPruned(t *testing.T) {
	t.Parallel()

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
		Name:    "B",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddMetric("C", analysis.AnalysisMetricDefinition{
		Name:    "C",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddCILTree("A", &cil.CILLiteral{Value: 0})
	ma.AddCILTree("B", &cil.CILLiteral{Value: 0})
	ma.AddCILTree("C", &cil.CILLiteral{Value: 0})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	err = ma.EliminateDeadCode([]string{"A"})
	require.NoError(t, err)

	// A and B are reachable; C is not.
	assert.True(t, ma.HasMetric("A"))
	assert.True(t, ma.HasMetric("B"))
	assert.False(t, ma.HasMetric("C"))
	assert.True(t, ma.HasCILTree("A"))
	assert.True(t, ma.HasCILTree("B"))
	assert.False(t, ma.HasCILTree("C"))
	assert.Len(t, ma.GetMetrics(), 2)
	assert.Len(t, ma.GetCILTrees(), 2)
}

// TestEliminateDeadCode_EmptyOutputList verifies that all metrics are
// eliminated when the output list is empty.
func TestEliminateDeadCode_EmptyOutputList(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name:    "A",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddMetric("B", analysis.AnalysisMetricDefinition{
		Name:    "B",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddCILTree("A", &cil.CILLiteral{Value: 0})
	ma.AddCILTree("B", &cil.CILLiteral{Value: 0})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	err = ma.EliminateDeadCode([]string{})
	require.NoError(t, err)

	assert.Len(t, ma.GetMetrics(), 0)
	assert.Len(t, ma.GetCILTrees(), 0)
}

// TestEliminateDeadCode_SingleOutputMultiLevelChain verifies that all
// ancestors in a multi-level chain are preserved, but siblings are pruned.
func TestEliminateDeadCode_SingleOutputMultiLevelChain(t *testing.T) {
	t.Parallel()

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
			{Kind: "metric", Name: "C"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "C"}},
		},
	})
	ma.AddMetric("C", analysis.AnalysisMetricDefinition{
		Name:    "C",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	// Sibling: unrelated metric
	ma.AddMetric("D", analysis.AnalysisMetricDefinition{
		Name:    "D",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddCILTree("A", &cil.CILLiteral{Value: 0})
	ma.AddCILTree("B", &cil.CILLiteral{Value: 0})
	ma.AddCILTree("C", &cil.CILLiteral{Value: 0})
	ma.AddCILTree("D", &cil.CILLiteral{Value: 0})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	err = ma.EliminateDeadCode([]string{"A"})
	require.NoError(t, err)

	// A, B, C are reachable; D is not.
	assert.True(t, ma.HasMetric("A"))
	assert.True(t, ma.HasMetric("B"))
	assert.True(t, ma.HasMetric("C"))
	assert.False(t, ma.HasMetric("D"))
	assert.True(t, ma.HasCILTree("A"))
	assert.True(t, ma.HasCILTree("B"))
	assert.True(t, ma.HasCILTree("C"))
	assert.False(t, ma.HasCILTree("D"))
	assert.Len(t, ma.GetMetrics(), 3)
	assert.Len(t, ma.GetCILTrees(), 3)
}

// TestEliminateDeadCode_DiamondPattern verifies both paths to a common
// dependency are preserved.
func TestEliminateDeadCode_DiamondPattern(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name: "A",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "B"},
			{Kind: "metric", Name: "C"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{
				{Kind: "metric", Name: "B"},
				{Kind: "metric", Name: "C"},
			},
		},
	})
	ma.AddMetric("B", analysis.AnalysisMetricDefinition{
		Name: "B",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "D"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "D"}},
		},
	})
	ma.AddMetric("C", analysis.AnalysisMetricDefinition{
		Name: "C",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "D"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "D"}},
		},
	})
	ma.AddMetric("D", analysis.AnalysisMetricDefinition{
		Name:    "D",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddCILTree("A", &cil.CILLiteral{Value: 0})
	ma.AddCILTree("B", &cil.CILLiteral{Value: 0})
	ma.AddCILTree("C", &cil.CILLiteral{Value: 0})
	ma.AddCILTree("D", &cil.CILLiteral{Value: 0})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	err = ma.EliminateDeadCode([]string{"A"})
	require.NoError(t, err)

	// All four are reachable via diamond paths.
	assert.True(t, ma.HasMetric("A"))
	assert.True(t, ma.HasMetric("B"))
	assert.True(t, ma.HasMetric("C"))
	assert.True(t, ma.HasMetric("D"))
	assert.Len(t, ma.GetMetrics(), 4)
	assert.Len(t, ma.GetCILTrees(), 4)
}

// TestEliminateDeadCode_UnknownOutputMetric verifies that unknown output
// metric names are handled gracefully — metrics are still eliminated if
// no known outputs are specified.
func TestEliminateDeadCode_UnknownOutputMetric(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name:    "A",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddCILTree("A", &cil.CILLiteral{Value: 0})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	// "Unknown" does not exist — skip it. No known outputs → all metrics eliminated.
	err = ma.EliminateDeadCode([]string{"Unknown"})
	require.NoError(t, err)

	assert.Len(t, ma.GetMetrics(), 0)
	assert.Len(t, ma.GetCILTrees(), 0)
}

// TestEliminateDeadCode_UnknownAndKnownOutputs verifies that unknown output
// metrics are skipped while known outputs preserve their dependencies.
func TestEliminateDeadCode_UnknownAndKnownOutputs(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name:    "A",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddMetric("B", analysis.AnalysisMetricDefinition{
		Name:    "B",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddCILTree("A", &cil.CILLiteral{Value: 0})
	ma.AddCILTree("B", &cil.CILLiteral{Value: 0})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	// "Unknown" is skipped, but "A" is a valid output — so A is reachable.
	err = ma.EliminateDeadCode([]string{"Unknown", "A"})
	require.NoError(t, err)

	assert.True(t, ma.HasMetric("A"))
	assert.False(t, ma.HasMetric("B"))
	assert.Len(t, ma.GetMetrics(), 1)
	assert.Len(t, ma.GetCILTrees(), 1)
}

// TestEliminateDeadCode_NilDAG returns an error when the DAG has not been built.
func TestEliminateDeadCode_NilDAG(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name:    "A",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.EliminateDeadCode([]string{"A"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DAG has not been built")
}

// TestEliminateDeadCode_Integration_ParsedMetrics verifies dead-code elimination
// on metrics produced by parsing and analyzing CPL.
func TestEliminateDeadCode_Integration_ParsedMetrics(t *testing.T) {
	t.Parallel()

	input := `concepts
	Dim = [Values[D1]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]

measures
	base_measure = Number(unit=Unit)
	Daily = Interval(unit=Day)

metrics
	metric_a
		inputs
			metric base_measure
		calculate
			0
	metric_b = metric_a + 1
	metric_c
		inputs
			metric base_measure
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed")

	// Set up a MetricAnalyzer with the analyzed metrics for dead-code elimination.
	ma := analysis.NewMetricAnalyzer()
	for name, metric := range defs.Metrics {
		ma.AddMetric(name, metric)
	}
	for name, tree := range defs.CILTrees {
		ma.AddCILTree(name, tree)
	}

	err = ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	// Store IO spec with only metric_b as output.
	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{MetricName: "metric_b"},
		},
	}
	err = ma.ValidateAndStoreIOSpec(ioSpec)
	require.NoError(t, err)

	// Eliminate dead code.
	err = ma.EliminateDeadCode([]string{"metric_b"})
	require.NoError(t, err)

	// Verify metric_a and metric_b are reachable; metric_c is dead code.
	assert.True(t, ma.HasMetric("metric_a"))
	assert.True(t, ma.HasMetric("metric_b"))
	assert.False(t, ma.HasMetric("metric_c"))

	// Verify DAG was also pruned.
	dag := ma.GetDAG()
	assert.NotContains(t, dag.Nodes, "metric_c")

	// Verify topological sort works correctly on the pruned/rebuilt DAG
	// without needing an explicit rebuild call.
	order, err := dag.TopologicalSort()
	require.NoError(t, err, "TopologicalSort should succeed after dead-code elimination")
	require.Len(t, order, 2)
	assert.Equal(t, "metric_a", order[0], "base dependency should come first")
	assert.Equal(t, "metric_b", order[1], "dependent metric should come last")
}

// ============================================================================
// Regression Tests for Review Issues
// ============================================================================

// TestTopologicalSortAfterEliminateDeadCode verifies that TopologicalSort
// works correctly after EliminateDeadCode has pruned the DAG, without
// any intermediate rebuild. The DAG rebuild happens inside EliminateDeadCode.
func TestTopologicalSortAfterEliminateDeadCode(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	// A depends on B and C; D is an unrelated independent metric.
	// Outputting only B makes A, C, and D unreachable — only B survives.
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name: "A",
		Inputs: []analysis.AnalysisInputDefinition{
			{Kind: "metric", Name: "B"},
			{Kind: "metric", Name: "C"},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{
				{Kind: "metric", Name: "B"},
				{Kind: "metric", Name: "C"},
			},
		},
	})
	ma.AddMetric("B", analysis.AnalysisMetricDefinition{
		Name:    "B",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddMetric("C", analysis.AnalysisMetricDefinition{
		Name:    "C",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})
	ma.AddMetric("D", analysis.AnalysisMetricDefinition{
		Name:    "D",
		Inputs:  []analysis.AnalysisInputDefinition{},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	// Eliminate dead code: only B is in the output list.
	// Reachable: {B}. A, C, and D are pruned.
	err = ma.EliminateDeadCode([]string{"B"})
	require.NoError(t, err)

	// Only B survives — A depended on B (and C) but A is not in reachable set.
	// TopologicalSort should succeed without false cycle detection.
	dag := ma.GetDAG()
	require.NotNil(t, dag)
	require.Len(t, dag.Nodes, 1)
	assert.Contains(t, dag.Nodes, "B")
	assert.NotContains(t, dag.Nodes, "A")
	assert.NotContains(t, dag.Nodes, "C")
	assert.NotContains(t, dag.Nodes, "D")

	order, err := dag.TopologicalSort()
	require.NoError(t, err, "TopologicalSort should succeed after dead-code elimination without rebuilding")
	assert.Equal(t, []string{"B"}, order)
}

// TestDetectCycles_MixedTimeShiftedAndNormalSelfReference verifies that a
// metric with both a time-shifted self-reference and a normal self-reference
// is correctly detected as a cycle. Only pure time-shifted self-loops are
// permitted.
func TestDetectCycles_MixedTimeShiftedAndNormalSelfReference(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name: "A",
		Inputs: []analysis.AnalysisInputDefinition{
			{
				Kind:      "metric",
				Name:      "A",
				TimeShift: &analysis.AnalysisTimeShift{Direction: "backward", Field: "day", Amount: 1},
			},
			{Kind: "metric", Name: "A"}, // Normal self-reference — should cause a cycle
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "A"}},
		},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	// Should detect a cycle because there's a non-time-shifted self-reference.
	err = ma.GetDAG().DetectCycles(ma.GetMetrics())
	require.Error(t, err, "Mixed self-references should be detected as a cycle")
	assert.Contains(t, err.Error(), "circular dependency")
}

// TestBuildMetricDependencyDAG_TimeShiftedSelfReferenceInCILInputRefs verifies
// that a time-shifted self-reference present in CIL InputRefs (not just Inputs)
// is correctly skipped during DAG construction. This exercises the
// hasTimeShiftedSelfReference code path.
func TestBuildMetricDependencyDAG_TimeShiftedSelfReferenceInCILInputRefs(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	// Metric A has a time-shifted self-reference in both Inputs and CIL InputRefs.
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name: "A",
		Inputs: []analysis.AnalysisInputDefinition{
			{
				Kind:      "metric",
				Name:      "A",
				TimeShift: &analysis.AnalysisTimeShift{Direction: "backward", Field: "day", Amount: 1},
			},
		},
		CILRoot: &cil.CILMetricNode{
			CILNode:   &cil.CILLiteral{Value: 0},
			InputRefs: []cil.CILInputRef{{Kind: "metric", Name: "A"}},
		},
	})

	err := ma.BuildMetricDependencyDAG()
	require.NoError(t, err)

	dag := ma.GetDAG()
	require.Len(t, dag.Nodes, 1)
	// The self-reference should have been skipped (time-shifted).
	assert.Empty(t, dag.Nodes["A"].Dependencies,
		"Time-shifted self-reference in CIL InputRefs should not create a dependency edge")
	assert.Empty(t, dag.Nodes["A"].Dependents)

	// Cycle detection should pass.
	err = dag.DetectCycles(ma.GetMetrics())
	require.NoError(t, err, "Time-shifted self-reference should not be detected as a cycle")
}

// TestDetectCycles_PermitsSelfLoopEdge_Defensive verifies the defensive code path
// in DetectCycles: when a self-loop edge is manually injected into a DAG node
// (bypassing BuildMetricDependencyDAG which normally skips time-shifted edges),
// DetectCycles still permits it via isPermittedSelfReference. This ensures the
// defensive check works correctly even when the DAG is programmatically
// constructed or modified outside the normal build flow.
func TestDetectCycles_PermitsSelfLoopEdge_Defensive(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("A", analysis.AnalysisMetricDefinition{
		Name: "A",
		Inputs: []analysis.AnalysisInputDefinition{
			{
				Kind:      "metric",
				Name:      "A",
				TimeShift: &analysis.AnalysisTimeShift{Direction: "backward", Field: "day", Amount: 1},
			},
		},
		CILRoot: &cil.CILMetricNode{CILNode: &cil.CILLiteral{Value: 0}},
	})

	// Manually construct a DAG with a self-loop edge.
	// BuildMetricDependencyDAG normally skips time-shifted self-references,
	// so we bypass it to directly test the defensive isPermittedSelfReference check.
	dag := &analysis.MetricDAG{
		Nodes: map[string]*analysis.DAGNode{
			"A": {
				MetricName:   "A",
				Dependencies: []string{"A"}, // Manually injected self-loop
			},
		},
	}

	// DetectCycles should permit this because the metric has a time-shifted
	// self-reference in its Inputs and no conflicting non-time-shifted self-ref.
	err := dag.DetectCycles(ma.GetMetrics())
	require.NoError(t, err, "Time-shifted self-loop edge should be permitted by isPermittedSelfReference")
}
