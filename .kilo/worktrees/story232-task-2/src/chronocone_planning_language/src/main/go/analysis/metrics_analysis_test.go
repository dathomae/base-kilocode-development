// Package analysis_test contains unit tests for the metrics analysis phase.
//
// These tests follow the Behavioral Test-Driven Development (BTDD) approach,
// where tests are written before the implementation exists. This is Phase B
// of the BTDD lifecycle - defining test signatures and expected behaviors.
//
// The tests cover:
//   - AnalyzeMetricsScope assembles AnalysisMetricDefinition
//   - CIL tree construction for Increment 1 (single CILLiteral(0))
//   - Multiple metrics in one section
//   - Input validation against declared entities
//   - Connector CIL tree generation
//
// Story: Story182 - CPL Metrics Basic Parse and Analyze
// Task: Task 3 Phase B - Write logical tests for metrics analysis
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	cpio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parsing"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// ============================================================================
// AnalyzeMetricScope Basic Assembly Tests
// ============================================================================

// TestAnalyzeMetricScope_AssemblesDefinition verifies that AnalyzeMetricScope
// correctly assembles an AnalysisMetricDefinition with Name, Inputs, and CILRoot
// from a parsed metric scope.
func TestAnalyzeMetricScope_AssemblesDefinition(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a basic metric with inputs and calculate clause
	input := `metrics
	my_metric
		inputs
			metric demand_forecast
			dimension Price
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")
	require.NotNil(t, rootTable, "Root table should not be nil")

	// Create analyzer with pre-populated concepts and measures for validation
	analyzer := analysis.NewCPLAnalyzer()

	// Add required concept for dimension validation
	priceConcept := analysis.AnalysisConceptDefinition{ConceptName: "Price"}
	analyzer.AddNewConcept(&priceConcept)

	// Add required measure for metric input validation
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["demand_forecast"] = analysis.AnalysisMeasureDefinition{
		Name:     "demand_forecast",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	// Set up handlers for metrics analysis
	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Walk the symbol table with metrics handlers
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Analysis should succeed and metric definition should be assembled
	require.NoError(t, err, "Walk should not return error for valid metric")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "my_metric", "Metric should be in results")

	metric := resultDefs.Metrics["my_metric"]
	assert.Equal(t, "my_metric", metric.Name, "Metric name should match")
	require.Len(t, metric.Inputs, 2, "Metric should have 2 inputs")

	// Verify first input (metric reference)
	assert.Equal(t, "metric", metric.Inputs[0].Kind, "First input kind should be 'metric'")
	assert.Equal(t, "demand_forecast", metric.Inputs[0].Name, "First input name should match")

	// Verify second input (dimension reference)
	assert.Equal(t, "dimension", metric.Inputs[1].Kind, "Second input kind should be 'dimension'")
	assert.Equal(t, "Price", metric.Inputs[1].Name, "Second input name should match")

	// Verify CILRoot is present
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")
}

// TestAnalyzeMetricScope_CILTreeConstruction verifies that for Increment 1,
// the CIL tree contains a single CILLiteral(0) node.
func TestAnalyzeMetricScope_CILTreeConstruction(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with calculate 0
	input := `metrics
	simple_metric
		inputs
			metric input1
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer with required measure
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["input1"] = analysis.AnalysisMeasureDefinition{
		Name:     "input1",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze the metric
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: CIL tree should be correctly constructed
	require.NoError(t, err, "Analysis should succeed")

	resultDefs := analyzer.GetAnalysisDefinitions()
	metric := resultDefs.Metrics["simple_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// For Increment 1, CILNode should be a CILLiteral with value 0
	literal, ok := metric.CILRoot.CILNode.(*cil.CILLiteral)
	require.True(t, ok, "CILNode should be a CILLiteral for Increment 1")
	assert.Equal(t, float64(0), literal.Value, "CILLiteral value should be 0")

	// Verify no input references in the CIL tree for a literal-only metric
	assert.Empty(t, metric.CILRoot.InputRefs, "InputRefs should be empty for literal-only metric")
}

// TestAnalyzeMetricScope_LiteralOneCILTree verifies CIL tree construction
// for calculate 1 produces CILLiteral(1).
func TestAnalyzeMetricScope_LiteralOneCILTree(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with calculate 1
	input := `metrics
	metric_one
		inputs
			metric input1
		calculate
			1`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer with required measure
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["input1"] = analysis.AnalysisMeasureDefinition{
		Name:     "input1",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze the metric
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: CIL tree should contain CILLiteral(1)
	require.NoError(t, err, "Analysis should succeed")

	resultDefs := analyzer.GetAnalysisDefinitions()
	metric := resultDefs.Metrics["metric_one"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	literal, ok := metric.CILRoot.CILNode.(*cil.CILLiteral)
	require.True(t, ok, "CILNode should be a CILLiteral")
	assert.Equal(t, float64(1), literal.Value, "CILLiteral value should be 1")
}

// ============================================================================
// Multiple Metrics Tests
// ============================================================================

// TestAnalyzeMetricsScope_MultipleMetrics verifies that analysis produces
// one AnalysisMetricDefinition per metric in a metrics section.
func TestAnalyzeMetricsScope_MultipleMetrics(t *testing.T) {
	t.Parallel()

	// Arrange: Parse multiple metrics in one section
	input := `metrics
	first_metric
		inputs
			metric input1
		calculate
			0
	second_metric
		inputs
			dimension Dim1
		calculate
			1
	third_metric
		inputs
			metric input2
			dimension Dim2
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer with required concepts and measures
	analyzer := analysis.NewCPLAnalyzer()

	// Add concepts for dimension validation
	dim1Concept := analysis.AnalysisConceptDefinition{ConceptName: "Dim1"}
	dim2Concept := analysis.AnalysisConceptDefinition{ConceptName: "Dim2"}
	analyzer.AddNewConcept(&dim1Concept)
	analyzer.AddNewConcept(&dim2Concept)

	// Add measures for metric input validation
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["input1"] = analysis.AnalysisMeasureDefinition{
		Name:     "input1",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}
	defs.Measures["input2"] = analysis.AnalysisMeasureDefinition{
		Name:     "input2",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze the metrics
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: All three metrics should be present
	require.NoError(t, err, "Analysis should succeed")

	resultDefs := analyzer.GetAnalysisDefinitions()
	assert.Len(t, resultDefs.Metrics, 3, "Should have 3 metrics")

	// Verify first metric
	require.Contains(t, resultDefs.Metrics, "first_metric")
	first := resultDefs.Metrics["first_metric"]
	assert.Equal(t, "first_metric", first.Name)
	require.Len(t, first.Inputs, 1)
	assert.Equal(t, "metric", first.Inputs[0].Kind)

	// Verify second metric
	require.Contains(t, resultDefs.Metrics, "second_metric")
	second := resultDefs.Metrics["second_metric"]
	assert.Equal(t, "second_metric", second.Name)
	require.Len(t, second.Inputs, 1)
	assert.Equal(t, "dimension", second.Inputs[0].Kind)

	// Verify third metric
	require.Contains(t, resultDefs.Metrics, "third_metric")
	third := resultDefs.Metrics["third_metric"]
	assert.Equal(t, "third_metric", third.Name)
	require.Len(t, third.Inputs, 2)
}

// TestAnalyzeMetricsScope_DuplicateMetricNames_ReturnsError verifies that
// duplicate metric names in the same section are detected and rejected.
func TestAnalyzeMetricsScope_DuplicateMetricNames_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with duplicate metric names
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)

	// First metric scope
	firstMetricScope := metricsScope.AddChild(symboltable.MetricScope)
	firstMetricDef := &parsing.ParserMetricDefinition{
		Name:             "duplicate_metric",
		Inputs:           []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 0}},
		SourceLocation:   symboltable.SourceLocation{Line: 2, Column: 1},
	}
	firstMetricScope.Data = firstMetricDef
	nameCopy1 := firstMetricDef.Name
	_ = firstMetricScope.DefineWithData(&nameCopy1, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, firstMetricDef)

	// Second metric scope with same name
	secondMetricScope := metricsScope.AddChild(symboltable.MetricScope)
	secondMetricDef := &parsing.ParserMetricDefinition{
		Name:             "duplicate_metric",
		Inputs:           []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 1}},
		SourceLocation:   symboltable.SourceLocation{Line: 8, Column: 1},
	}
	secondMetricScope.Data = secondMetricDef
	nameCopy2 := secondMetricDef.Name
	_ = secondMetricScope.DefineWithData(&nameCopy2, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, secondMetricDef)

	analyzer := analysis.NewCPLAnalyzer()
	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze the metrics
	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	// Assert: Should return error for duplicate metric name
	require.Error(t, err, "Should return error for duplicate metric name")
	assert.Contains(t, err.Error(), "duplicate", "Error should mention duplicate")
	assert.Contains(t, err.Error(), "duplicate_metric", "Error should mention the metric name")
}

// ============================================================================
// Input Validation Tests
// ============================================================================

// TestAnalyzeMetricScope_UndefinedMetricInput_ReturnsError verifies that
// referencing an undefined metric in inputs produces an error.
func TestAnalyzeMetricScope_UndefinedMetricInput_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric referencing an undefined metric input
	input := `metrics
	my_metric
		inputs
			metric undefined_metric
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer WITHOUT adding the referenced metric
	analyzer := analysis.NewCPLAnalyzer()
	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze without the required measure defined
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Should return error for undefined metric input
	require.Error(t, err, "Should return error for undefined metric input")
	assert.Contains(t, err.Error(), "undefined", "Error should mention undefined")
	assert.Contains(t, err.Error(), "undefined_metric", "Error should mention the metric name")
}

// TestAnalyzeMetricScope_UndefinedDimensionInput_ReturnsError verifies that
// referencing an undefined dimension in inputs produces an error.
func TestAnalyzeMetricScope_UndefinedDimensionInput_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric referencing an undefined dimension
	input := `metrics
	my_metric
		inputs
			dimension UndefinedDimension
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer WITHOUT adding the referenced concept
	analyzer := analysis.NewCPLAnalyzer()
	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze without the required concept defined
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Should return error for undefined dimension
	require.Error(t, err, "Should return error for undefined dimension")
	assert.Contains(t, err.Error(), "undefined", "Error should mention undefined")
	assert.Contains(t, err.Error(), "UndefinedDimension", "Error should mention the dimension name")
}

// TestAnalyzeMetricScope_ValidMetricInput_Accepted verifies that referencing
// a declared metric in inputs is accepted.
func TestAnalyzeMetricScope_ValidMetricInput_Accepted(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with valid metric input reference
	input := `metrics
	my_metric
		inputs
			metric valid_input
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer WITH the referenced measure defined
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["valid_input"] = analysis.AnalysisMeasureDefinition{
		Name:     "valid_input",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze with the required measure defined
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Should succeed
	require.NoError(t, err, "Analysis should succeed for valid metric input")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "my_metric")
	metric := resultDefs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 1)
	assert.Equal(t, "metric", metric.Inputs[0].Kind)
	assert.Equal(t, "valid_input", metric.Inputs[0].Name)
}

// TestAnalyzeMetricScope_ValidDimensionInput_Accepted verifies that referencing
// a declared dimension in inputs is accepted.
func TestAnalyzeMetricScope_ValidDimensionInput_Accepted(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with valid dimension input reference
	input := `metrics
	my_metric
		inputs
			dimension ValidDimension
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer WITH the referenced concept defined
	analyzer := analysis.NewCPLAnalyzer()
	validConcept := analysis.AnalysisConceptDefinition{ConceptName: "ValidDimension"}
	analyzer.AddNewConcept(&validConcept)

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze with the required concept defined
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Should succeed
	require.NoError(t, err, "Analysis should succeed for valid dimension input")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "my_metric")
	metric := resultDefs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 1)
	assert.Equal(t, "dimension", metric.Inputs[0].Kind)
	assert.Equal(t, "ValidDimension", metric.Inputs[0].Name)
}

// TestAnalyzeMetricScope_MixedInputs_Validation verifies that a mix of valid
// metric and dimension inputs are all validated correctly.
func TestAnalyzeMetricScope_MixedInputs_Validation(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with mixed inputs
	input := `metrics
	my_metric
		inputs
			metric metric_a
			dimension concept_x
			metric metric_b
			dimension concept_y
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer with all referenced entities defined
	analyzer := analysis.NewCPLAnalyzer()

	// Add concepts for dimensions
	conceptX := analysis.AnalysisConceptDefinition{ConceptName: "concept_x"}
	conceptY := analysis.AnalysisConceptDefinition{ConceptName: "concept_y"}
	analyzer.AddNewConcept(&conceptX)
	analyzer.AddNewConcept(&conceptY)

	// Add measures for metric inputs
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["metric_a"] = analysis.AnalysisMeasureDefinition{
		Name:     "metric_a",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}
	defs.Measures["metric_b"] = analysis.AnalysisMeasureDefinition{
		Name:     "metric_b",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Should succeed with all 4 inputs validated
	require.NoError(t, err, "Analysis should succeed for valid mixed inputs")

	resultDefs := analyzer.GetAnalysisDefinitions()
	metric := resultDefs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 4)

	// Verify input order and types
	assert.Equal(t, "metric", metric.Inputs[0].Kind)
	assert.Equal(t, "metric_a", metric.Inputs[0].Name)
	assert.Equal(t, "dimension", metric.Inputs[1].Kind)
	assert.Equal(t, "concept_x", metric.Inputs[1].Name)
	assert.Equal(t, "metric", metric.Inputs[2].Kind)
	assert.Equal(t, "metric_b", metric.Inputs[2].Name)
	assert.Equal(t, "dimension", metric.Inputs[3].Kind)
	assert.Equal(t, "concept_y", metric.Inputs[3].Name)
}

// ============================================================================
// Connector CIL Tree Tests
// ============================================================================

// TestAnalyzeMetricsScope_ConnectorCILTree verifies that a single-node CIL tree
// is generated for each connector declaration.
func TestAnalyzeMetricsScope_ConnectorCILTree(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a CPL document with connectors and metrics
	input := `connectors
	SalesData = dimensions [Country] grain Day celltype Revenue

metrics
	my_metric
		inputs
			metric SalesData
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer with required concepts and measures
	analyzer := analysis.NewCPLAnalyzer()

	// Add Country concept
	countryConcept := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConcept)

	// Add Day and Revenue measures
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Dollar"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
		symboltable.MetricsScope:   analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:    analysis.AnalyzeMetricScope,
	}

	// Act: Analyze both connectors and metrics
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Should succeed and connector CIL tree should be present
	require.NoError(t, err, "Analysis should succeed")

	resultDefs := analyzer.GetAnalysisDefinitions()

	// Verify connector was analyzed
	require.Contains(t, resultDefs.Connectors, "SalesData", "Connector should be in results")

	// Verify connector CIL tree exists
	require.Contains(t, resultDefs.CILTrees, "SalesData", "Connector should have CIL tree")
	connectorCIL := resultDefs.CILTrees["SalesData"]
	require.NotNil(t, connectorCIL, "Connector CIL tree should not be nil")
}

// TestAnalyzeMetricsScope_MultipleConnectors_CILTrees verifies that each
// connector gets its own single-node CIL tree.
func TestAnalyzeMetricsScope_MultipleConnectors_CILTrees(t *testing.T) {
	t.Parallel()

	// Arrange: Parse multiple connectors
	input := `connectors
	Conn1 = dimensions [Dim1] grain Day celltype Measure1
	Conn2 = dimensions [Dim2] grain Day celltype Measure2

metrics
	my_metric
		inputs
			metric Conn1
			metric Conn2
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer with required concepts and measures
	analyzer := analysis.NewCPLAnalyzer()

	// Add concepts
	dim1Concept := analysis.AnalysisConceptDefinition{ConceptName: "Dim1"}
	dim2Concept := analysis.AnalysisConceptDefinition{ConceptName: "Dim2"}
	analyzer.AddNewConcept(&dim1Concept)
	analyzer.AddNewConcept(&dim2Concept)

	// Add measures
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}
	defs.Measures["Measure1"] = analysis.AnalysisMeasureDefinition{
		Name:     "Measure1",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}
	defs.Measures["Measure2"] = analysis.AnalysisMeasureDefinition{
		Name:     "Measure2",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
		symboltable.MetricsScope:   analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:    analysis.AnalyzeMetricScope,
	}

	// Act: Analyze
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Both connectors should have CIL trees
	require.NoError(t, err, "Analysis should succeed")

	resultDefs := analyzer.GetAnalysisDefinitions()

	// Verify both connectors exist
	require.Contains(t, resultDefs.Connectors, "Conn1")
	require.Contains(t, resultDefs.Connectors, "Conn2")

	// Verify both have CIL trees
	require.Contains(t, resultDefs.CILTrees, "Conn1")
	require.Contains(t, resultDefs.CILTrees, "Conn2")

	// Verify CIL trees are distinct
	assert.NotNil(t, resultDefs.CILTrees["Conn1"])
	assert.NotNil(t, resultDefs.CILTrees["Conn2"])
}

// ============================================================================
// Full Analysis Integration Tests
// ============================================================================

// TestAnalyze_FullMetricsSection uses the complete Analyze function to verify
// that metrics are correctly processed in a full compilation context.
func TestAnalyze_FullMetricsSection(t *testing.T) {
	t.Parallel()

	// Arrange: Full CPL document with concepts, measures, connectors, and metrics
	input := `concepts
	Country = [Values[USA, Canada]]

types
	NumberType = Number()

units
	Currency = Country

measures
	Revenue = NumberType(unit=Currency)
	Day = Interval(unit=Currency)

connectors
	SalesData = dimensions [Country] grain Day celltype Revenue

metrics
	total_revenue
		inputs
			metric SalesData
		calculate
			0`

	// Act: Parse and analyze the full document
	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)

	// Assert: Full analysis should succeed
	require.NoError(t, err, "Full analysis should succeed")
	require.NotNil(t, defs, "Analysis definitions should not be nil")

	// Verify metric was analyzed
	require.Contains(t, defs.Metrics, "total_revenue", "Metric should be in results")
	metric := defs.Metrics["total_revenue"]
	assert.Equal(t, "total_revenue", metric.Name)
	require.Len(t, metric.Inputs, 1)
	assert.Equal(t, "metric", metric.Inputs[0].Kind)
	assert.Equal(t, "SalesData", metric.Inputs[0].Name)

	// Verify CIL tree exists
	require.NotNil(t, metric.CILRoot, "Metric CILRoot should not be nil")
	literal, ok := metric.CILRoot.CILNode.(*cil.CILLiteral)
	require.True(t, ok, "CILNode should be CILLiteral")
	assert.Equal(t, float64(0), literal.Value)
}

// TestAnalyze_MetricsOnlySection verifies analysis of a standalone metrics section
// when all referenced entities are properly declared.
func TestAnalyze_MetricsOnlySection(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table programmatically for precise control
	root := symboltable.NewGlobalSymbolTable()

	// Add concepts scope with Country concept
	conceptsScope := root.AddChild(symboltable.ConceptsScope)
	conceptScope := conceptsScope.AddChild(symboltable.ConceptScope)
	conceptName := "Country"
	conceptScope.Define(&conceptName, symboltable.Identifier, symboltable.SourceLocation{Line: 2, Column: 1})
	valueGenScope := conceptScope.AddChild(symboltable.ValueGeneratorScope)
	valueUSA := "USA"
	valueCanada := "Canada"
	valueGenScope.Define(&valueUSA, symboltable.Identifier, symboltable.SourceLocation{Line: 2, Column: 20})
	valueGenScope.Define(&valueCanada, symboltable.Identifier, symboltable.SourceLocation{Line: 2, Column: 25})

	// Add measures scope with required measures
	measuresScope := root.AddChild(symboltable.MeasuresScope)

	// Revenue measure: Revenue = Number(unit=Currency)
	revenueScope := measuresScope.AddChild(symboltable.MeasureScope)
	revenueName := "Revenue"
	revenueType := "Number"
	revenueScope.Define(&revenueName, symboltable.Identifier, symboltable.SourceLocation{Line: 5, Column: 1})
	revenueScope.Define(&revenueType, symboltable.Identifier, symboltable.SourceLocation{Line: 5, Column: 10})
	_ = revenueScope.DefineWithData(nil, symboltable.UnitReferenceDefinition, symboltable.SourceLocation{Line: 5, Column: 22}, parser.ParserUnitReference{Name: "Currency"})

	// Day measure: Day = Interval(unit=day)
	dayScope := measuresScope.AddChild(symboltable.MeasureScope)
	dayName := "Day"
	dayType := "Interval"
	dayScope.Define(&dayName, symboltable.Identifier, symboltable.SourceLocation{Line: 6, Column: 1})
	dayScope.Define(&dayType, symboltable.Identifier, symboltable.SourceLocation{Line: 6, Column: 6})
	_ = dayScope.DefineWithData(nil, symboltable.UnitReferenceDefinition, symboltable.SourceLocation{Line: 6, Column: 20}, parser.ParserUnitReference{Name: "day"})

	// Add connectors scope with SalesData connector
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)
	connectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 8, Column: 1},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Revenue",
	}

	// Add metrics scope with a metric
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name: "forecast",
		Inputs: []parsing.ParserInputDefinition{
			{Kind: "metric", Name: "SalesData", NameLocation: symboltable.SourceLocation{Line: 12, Column: 3}},
		},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 0}},
		SourceLocation:   symboltable.SourceLocation{Line: 11, Column: 1},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	// Act: Analyze the constructed symbol table
	defs, err := analysis.Analyze(root)

	// Assert: Analysis should succeed
	require.NoError(t, err, "Analysis should succeed for valid metrics section")
	require.NotNil(t, defs, "Definitions should not be nil")

	// Verify metric exists
	require.Contains(t, defs.Metrics, "forecast", "Metric should be in results")
	metric := defs.Metrics["forecast"]
	assert.Equal(t, "forecast", metric.Name)
	require.Len(t, metric.Inputs, 1)
	assert.Equal(t, "metric", metric.Inputs[0].Kind)
	assert.Equal(t, "SalesData", metric.Inputs[0].Name)
}

// ============================================================================
// Arithmetic Expression Integration Tests (Parse → Analyze Pipeline)
// ============================================================================
//
// These tests exercise the full parse→analyze pipeline for metrics with
// arithmetic expressions in their calculate clauses. They verify that the
// parser correctly produces postfix sequences and the analyzer consumes them
// to build correct CIL expression trees.
//
// Story: Story184 - CPL Metrics Arithmetic Parse and Analyze
// Task: Task 6c - Integration verification

// TestParseAndAnalyze_SimpleAddition verifies the full pipeline for a simple
// binary addition expression: calculate a + b.
func TestParseAndAnalyze_SimpleAddition(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with a simple addition expression
	input := `metrics
	arith_metric
		inputs
			metric input_a
			metric input_b
		calculate
			input_a + input_b`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer with required measures for input validation
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["input_a"] = analysis.AnalysisMeasureDefinition{
		Name:     "input_a",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}
	defs.Measures["input_b"] = analysis.AnalysisMeasureDefinition{
		Name:     "input_b",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze the parsed metric
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Analysis should succeed and produce correct CIL tree
	require.NoError(t, err, "Analysis should succeed for simple addition")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "arith_metric")
	metric := resultDefs.Metrics["arith_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILBinaryOp with "+" operator
	binOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "Root CILNode should be CILBinaryOp")
	assert.Equal(t, operator.Add, binOp.Op, "Operator should be '+'")

	// Verify left operand is CILInputRef referencing input_a
	leftRef, ok := binOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Left operand should be CILInputRef")
	assert.Equal(t, "input_a", leftRef.Name)

	// Verify right operand is CILInputRef referencing input_b
	rightRef, ok := binOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Right operand should be CILInputRef")
	assert.Equal(t, "input_b", rightRef.Name)

	// Verify InputRefs collection
	require.Len(t, metric.CILRoot.InputRefs, 2, "InputRefs should contain both identifiers")
}

// TestParseAndAnalyze_MultiplicationWithPrecedence verifies the full pipeline
// for an expression with operator precedence: calculate a * b + c
// which should parse as (a * b) + c.
func TestParseAndAnalyze_MultiplicationWithPrecedence(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with multiplication and addition
	input := `metrics
	prec_metric
		inputs
			metric a
			metric b
			metric c
		calculate
			a * b + c`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer with required measures
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	for _, name := range []string{"a", "b", "c"} {
		defs.Measures[name] = analysis.AnalysisMeasureDefinition{
			Name:     name,
			TypeName: "Number",
			UnitExpression: &analysis.UnitReference{Name: "Unit"},
		}
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze the parsed metric
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Analysis should succeed and produce correct CIL tree with precedence
	require.NoError(t, err, "Analysis should succeed for precedence expression")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "prec_metric")
	metric := resultDefs.Metrics["prec_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Expression: a * b + c should parse as (a * b) + c
	// Root should be CILBinaryOp{Op: operator.Add}
	rootOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "Root CILNode should be CILBinaryOp")
	assert.Equal(t, operator.Add, rootOp.Op, "Root operator should be '+'")

	// Right child should be CILInputRef{c} (addition has lower precedence)
	rightRef, ok := rootOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Right child should be CILInputRef (c)")
	assert.Equal(t, "c", rightRef.Name)

	// Left child should be CILBinaryOp{Op: operator.Multiply} representing (a * b)
	leftOp, ok := rootOp.Left.(*cil.CILBinaryOp)
	require.True(t, ok, "Left child should be CILBinaryOp (multiplication)")
	assert.Equal(t, operator.Multiply, leftOp.Op, "Left child operator should be '*'")

	// Left-left: CILInputRef{a}
	leftLeft, ok := leftOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Left-left should be CILInputRef (a)")
	assert.Equal(t, "a", leftLeft.Name)

	// Left-right: CILInputRef{b}
	leftRight, ok := leftOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Left-right should be CILInputRef (b)")
	assert.Equal(t, "b", leftRight.Name)

	// Verify InputRefs collection contains all 3 identifiers
	require.Len(t, metric.CILRoot.InputRefs, 3, "InputRefs should contain a, b, and c")
}

// TestParseAndAnalyze_SimpleSubtraction verifies the full pipeline for a simple
// binary subtraction expression: calculate a - b.
// Expected CIL tree: CILBinaryOp{Op: operator.Subtract} with Left: CILInputRef{a}, Right: CILInputRef{b}.
func TestParseAndAnalyze_SimpleSubtraction(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with a simple subtraction expression
	input := `metrics
	sub_metric
		inputs
			metric a
			metric b
		calculate
			a - b`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer with required measures
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	for _, name := range []string{"a", "b"} {
		defs.Measures[name] = analysis.AnalysisMeasureDefinition{
			Name:     name,
			TypeName: "Number",
			UnitExpression: &analysis.UnitReference{Name: "Unit"},
		}
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze the parsed metric
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Analysis should succeed and produce correct CIL tree
	require.NoError(t, err, "Analysis should succeed for simple subtraction")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "sub_metric")
	metric := resultDefs.Metrics["sub_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILBinaryOp with "-" operator
	binOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "Root CILNode should be CILBinaryOp")
	assert.Equal(t, operator.Subtract, binOp.Op, "Operator should be '-'")

	// Verify left operand is CILInputRef referencing a
	leftRef, ok := binOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Left operand should be CILInputRef")
	assert.Equal(t, "a", leftRef.Name)

	// Verify right operand is CILInputRef referencing b
	rightRef, ok := binOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Right operand should be CILInputRef")
	assert.Equal(t, "b", rightRef.Name)

	// Verify InputRefs collection
	require.Len(t, metric.CILRoot.InputRefs, 2, "InputRefs should contain both identifiers")
}

// TestParseAndAnalyze_SimpleDivision verifies the full pipeline for a simple
// binary division expression: calculate a / b.
// Expected CIL tree: CILBinaryOp{Op: operator.Divide} with Left: CILInputRef{a}, Right: CILInputRef{b}.
func TestParseAndAnalyze_SimpleDivision(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with a simple division expression
	input := `metrics
	div_metric
		inputs
			metric a
			metric b
		calculate
			a / b`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer with required measures
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	for _, name := range []string{"a", "b"} {
		defs.Measures[name] = analysis.AnalysisMeasureDefinition{
			Name:     name,
			TypeName: "Number",
			UnitExpression: &analysis.UnitReference{Name: "Unit"},
		}
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze the parsed metric
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Analysis should succeed and produce correct CIL tree
	require.NoError(t, err, "Analysis should succeed for simple division")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "div_metric")
	metric := resultDefs.Metrics["div_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILBinaryOp with "/" operator
	binOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "Root CILNode should be CILBinaryOp")
	assert.Equal(t, operator.Divide, binOp.Op, "Operator should be '/'")

	// Verify left operand is CILInputRef referencing a
	leftRef, ok := binOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Left operand should be CILInputRef")
	assert.Equal(t, "a", leftRef.Name)

	// Verify right operand is CILInputRef referencing b
	rightRef, ok := binOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Right operand should be CILInputRef")
	assert.Equal(t, "b", rightRef.Name)

	// Verify InputRefs collection
	require.Len(t, metric.CILRoot.InputRefs, 2, "InputRefs should contain both identifiers")
}

// TestParseAndAnalyze_UnaryNegation verifies the full pipeline for a unary
// negation expression: calculate -a.
// Expected CIL tree: CILUnaryOp{Op: operator.Negate} with Operand: CILInputRef{a}.
func TestParseAndAnalyze_UnaryNegation(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with a unary negation expression
	input := `metrics
	neg_metric
		inputs
			metric a
		calculate
			-a`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer with required measures
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["a"] = analysis.AnalysisMeasureDefinition{
		Name:     "a",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze the parsed metric
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Analysis should succeed and produce correct CIL tree
	require.NoError(t, err, "Analysis should succeed for unary negation")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "neg_metric")
	metric := resultDefs.Metrics["neg_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILUnaryOp with "negate" operator
	unaryOp, ok := metric.CILRoot.CILNode.(*cil.CILUnaryOp)
	require.True(t, ok, "Root CILNode should be CILUnaryOp")
	assert.Equal(t, operator.Negate, unaryOp.Op, "Operator should be 'negate'")

	// Verify operand is CILInputRef referencing a
	operandRef, ok := unaryOp.Operand.(*cil.CILInputRef)
	require.True(t, ok, "Operand should be CILInputRef")
	assert.Equal(t, "a", operandRef.Name)

	// Verify InputRefs collection
	require.Len(t, metric.CILRoot.InputRefs, 1, "InputRefs should contain one identifier")
}

// TestParseAndAnalyze_Parenthesized verifies the full pipeline for a parenthesized
// expression: calculate (a + b) * (c - d).
// Expected CIL tree: CILBinaryOp{Op: operator.Multiply} whose Left is CILBinaryOp{Op: operator.Add}
// with CILInputRef{a}, CILInputRef{b} and whose Right is CILBinaryOp{Op: operator.Subtract}
// with CILInputRef{c}, CILInputRef{d}.
func TestParseAndAnalyze_Parenthesized(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with a parenthesized expression
	input := `metrics
	paren_metric
		inputs
			metric a
			metric b
			metric c
			metric d
		calculate
			(a + b) * (c - d)`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer with required measures
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	for _, name := range []string{"a", "b", "c", "d"} {
		defs.Measures[name] = analysis.AnalysisMeasureDefinition{
			Name:     name,
			TypeName: "Number",
			UnitExpression: &analysis.UnitReference{Name: "Unit"},
		}
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze the parsed metric
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Analysis should succeed and produce correct CIL tree
	require.NoError(t, err, "Analysis should succeed for parenthesized expression")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "paren_metric")
	metric := resultDefs.Metrics["paren_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Root should be CILBinaryOp{Op: operator.Multiply}
	rootOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "Root CILNode should be CILBinaryOp")
	assert.Equal(t, operator.Multiply, rootOp.Op, "Root operator should be '*'")

	// Left child should be CILBinaryOp{Op: operator.Add} representing (a + b)
	leftOp, ok := rootOp.Left.(*cil.CILBinaryOp)
	require.True(t, ok, "Left child should be CILBinaryOp (addition)")
	assert.Equal(t, operator.Add, leftOp.Op, "Left child operator should be '+'")

	// Left-left: CILInputRef{a}
	leftLeft, ok := leftOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Left-left should be CILInputRef (a)")
	assert.Equal(t, "a", leftLeft.Name)

	// Left-right: CILInputRef{b}
	leftRight, ok := leftOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Left-right should be CILInputRef (b)")
	assert.Equal(t, "b", leftRight.Name)

	// Right child should be CILBinaryOp{Op: operator.Subtract} representing (c - d)
	rightOp, ok := rootOp.Right.(*cil.CILBinaryOp)
	require.True(t, ok, "Right child should be CILBinaryOp (subtraction)")
	assert.Equal(t, operator.Subtract, rightOp.Op, "Right child operator should be '-'")

	// Right-left: CILInputRef{c}
	rightLeft, ok := rightOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Right-left should be CILInputRef (c)")
	assert.Equal(t, "c", rightLeft.Name)

	// Right-right: CILInputRef{d}
	rightRight, ok := rightOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Right-right should be CILInputRef (d)")
	assert.Equal(t, "d", rightRight.Name)

	// Verify InputRefs collection contains all 4 identifiers
	require.Len(t, metric.CILRoot.InputRefs, 4, "InputRefs should contain a, b, c, and d")
}

// TestParseAndAnalyze_ComplexExpression verifies the full pipeline for a complex
// expression combining multiple operators: calculate a * b + c / d - e.
// Expected CIL tree has 3 binary ops with correct precedence nesting.
func TestParseAndAnalyze_ComplexExpression(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with a complex expression
	input := `metrics
	complex_metric
		inputs
			metric a
			metric b
			metric c
			metric d
			metric e
		calculate
			a * b + c / d - e`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer with required measures
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	for _, name := range []string{"a", "b", "c", "d", "e"} {
		defs.Measures[name] = analysis.AnalysisMeasureDefinition{
			Name:     name,
			TypeName: "Number",
			UnitExpression: &analysis.UnitReference{Name: "Unit"},
		}
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze the parsed metric
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Analysis should succeed and produce correct CIL tree
	require.NoError(t, err, "Analysis should succeed for complex expression")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "complex_metric")
	metric := resultDefs.Metrics["complex_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Expression: a * b + c / d - e should parse as ((a * b) + (c / d)) - e
	// Root should be CILBinaryOp{Op: operator.Subtract}
	rootOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "Root CILNode should be CILBinaryOp")
	assert.Equal(t, operator.Subtract, rootOp.Op, "Root operator should be '-'")

	// Right child should be CILInputRef{e}
	rightRef, ok := rootOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Right child should be CILInputRef (e)")
	assert.Equal(t, "e", rightRef.Name)

	// Left child should be CILBinaryOp{Op: operator.Add} representing (a * b) + (c / d)
	leftOp, ok := rootOp.Left.(*cil.CILBinaryOp)
	require.True(t, ok, "Left child should be CILBinaryOp (addition)")
	assert.Equal(t, operator.Add, leftOp.Op, "Left child operator should be '+'")

	// Left-left child should be CILBinaryOp{Op: operator.Multiply} representing (a * b)
	leftLeftOp, ok := leftOp.Left.(*cil.CILBinaryOp)
	require.True(t, ok, "Left-left child should be CILBinaryOp (multiplication)")
	assert.Equal(t, operator.Multiply, leftLeftOp.Op, "Left-left operator should be '*'")

	// Left-left-left: CILInputRef{a}
	leftLeftLeft, ok := leftLeftOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Left-left-left should be CILInputRef (a)")
	assert.Equal(t, "a", leftLeftLeft.Name)

	// Left-left-right: CILInputRef{b}
	leftLeftRight, ok := leftLeftOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Left-left-right should be CILInputRef (b)")
	assert.Equal(t, "b", leftLeftRight.Name)

	// Left-right child should be CILBinaryOp{Op: operator.Divide} representing (c / d)
	leftRightOp, ok := leftOp.Right.(*cil.CILBinaryOp)
	require.True(t, ok, "Left-right child should be CILBinaryOp (division)")
	assert.Equal(t, operator.Divide, leftRightOp.Op, "Left-right operator should be '/'")

	// Left-right-left: CILInputRef{c}
	leftRightLeft, ok := leftRightOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Left-right-left should be CILInputRef (c)")
	assert.Equal(t, "c", leftRightLeft.Name)

	// Left-right-right: CILInputRef{d}
	leftRightRight, ok := leftRightOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Left-right-right should be CILInputRef (d)")
	assert.Equal(t, "d", leftRightRight.Name)

	// Verify InputRefs collection contains all 5 identifiers
	require.Len(t, metric.CILRoot.InputRefs, 5, "InputRefs should contain a, b, c, d, and e")
}

// TestParseAndAnalyze_MixedLiteralAndIdentifier verifies the full pipeline for an
// expression mixing literals and identifiers: calculate a + 5 * b.
// Expected CIL tree mixes CILLiteral and CILInputRef nodes correctly.
func TestParseAndAnalyze_MixedLiteralAndIdentifier(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with a mixed literal and identifier expression
	input := `metrics
	mixed_metric
		inputs
			metric a
			metric b
		calculate
			a + 5 * b`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer with required measures
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	for _, name := range []string{"a", "b"} {
		defs.Measures[name] = analysis.AnalysisMeasureDefinition{
			Name:     name,
			TypeName: "Number",
			UnitExpression: &analysis.UnitReference{Name: "Unit"},
		}
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze the parsed metric
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Analysis should succeed and produce correct CIL tree
	require.NoError(t, err, "Analysis should succeed for mixed literal and identifier expression")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "mixed_metric")
	metric := resultDefs.Metrics["mixed_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Expression: a + 5 * b should parse as a + (5 * b)
	// Root should be CILBinaryOp{Op: operator.Add}
	rootOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "Root CILNode should be CILBinaryOp")
	assert.Equal(t, operator.Add, rootOp.Op, "Root operator should be '+'")

	// Left child should be CILInputRef{a}
	leftRef, ok := rootOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Left child should be CILInputRef (a)")
	assert.Equal(t, "a", leftRef.Name)

	// Right child should be CILBinaryOp{Op: operator.Multiply} representing (5 * b)
	rightOp, ok := rootOp.Right.(*cil.CILBinaryOp)
	require.True(t, ok, "Right child should be CILBinaryOp (multiplication)")
	assert.Equal(t, operator.Multiply, rightOp.Op, "Right child operator should be '*'")

	// Right-left: CILLiteral{5}
	rightLeftLit, ok := rightOp.Left.(*cil.CILLiteral)
	require.True(t, ok, "Right-left should be CILLiteral (5)")
	assert.Equal(t, float64(5), rightLeftLit.Value)

	// Right-right: CILInputRef{b}
	rightRightRef, ok := rightOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Right-right should be CILInputRef (b)")
	assert.Equal(t, "b", rightRightRef.Name)

	// Verify InputRefs collection contains only identifiers (not literals)
	require.Len(t, metric.CILRoot.InputRefs, 2, "InputRefs should contain a and b")
}

// ============================================================================
// Postfix Evaluator (buildCILTree) Tests — Phase B: Logical Tests
// ============================================================================
//
// These tests verify the stack-based postfix evaluator that processes
// ParserMetricDefinition.CalculatePostfix to build CIL expression trees.
// Tests use direct symbol table construction to isolate the postfix
// evaluator from the parser.
//
// Story: Story184 - CPL Metrics Arithmetic Parse and Analyze
// Task: Task 4d Phase B - Logical tests for postfix evaluator

// TestBuildCILTree_LiteralValue verifies that a postfix with a single numeric
// literal produces a CILLiteral with the correct value. Tests multiple values
// including zero, positive, and negative integers.
func TestBuildCILTree_LiteralValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		postfix   []parsing.PostfixElement
		wantValue float64
	}{
		{name: "zero", postfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 0}}, wantValue: 0},
		{name: "one", postfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 1}}, wantValue: 1},
		{name: "fortyTwo", postfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 42}}, wantValue: 42},
		{name: "negativeFive", postfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: -5}}, wantValue: -5},
		{name: "float64", postfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 3.14}}, wantValue: 3.14},
		{name: "int64", postfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 100}}, wantValue: 100},
		{name: "negativeFloat64", postfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: -2.5}}, wantValue: -2.5},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Build symbol table with a metric containing the literal postfix
			root := symboltable.NewGlobalSymbolTable()
			metricsScope := root.AddChild(symboltable.MetricsScope)
			metricScope := metricsScope.AddChild(symboltable.MetricScope)
			metricDef := &parsing.ParserMetricDefinition{
				Name:             "test_metric",
				Inputs:           []parsing.ParserInputDefinition{},
				CalculatePostfix: tt.postfix,
			}
			metricScope.Data = metricDef
			nameCopy := metricDef.Name
			_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

			analyzer := analysis.NewCPLAnalyzer()

			// Act: Analyze the metric scope
			err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

			// Assert: Should succeed and produce correct CIL tree
			require.NoError(t, err, "Analysis should succeed for literal postfix")

			resultDefs := analyzer.GetAnalysisDefinitions()
			require.Contains(t, resultDefs.Metrics, "test_metric")
			metric := resultDefs.Metrics["test_metric"]
			require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

			literal, ok := metric.CILRoot.CILNode.(*cil.CILLiteral)
			require.True(t, ok, "CILNode should be a CILLiteral")
			assert.Equal(t, tt.wantValue, literal.Value, "CILLiteral value should match input")
			assert.Empty(t, metric.CILRoot.InputRefs, "InputRefs should be empty for literal-only postfix")
		})
	}
}

// TestAnalyzeMetricScope_DefinitionsOnlyPath verifies that AnalyzeMetricScope
// succeeds when the MetricScope has no Data field but contains a
// MetricScopeDefinition entry in Definitions. This ensures the primary
// Definitions-iteration path works without falling back to the legacy Data field.
func TestAnalyzeMetricScope_DefinitionsOnlyPath(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric whose Data is nil but Definitions
	// contains a MetricScopeDefinition entry.
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:             "definitions_only_metric",
		Inputs:           []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 42}},
	}
	metricScope.Data = nil
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed without falling back to Data
	require.NoError(t, err, "Analysis should succeed when metric definition is only in Definitions")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "definitions_only_metric")
	metric := resultDefs.Metrics["definitions_only_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	literal, ok := metric.CILRoot.CILNode.(*cil.CILLiteral)
	require.True(t, ok, "CILNode should be a CILLiteral")
	assert.Equal(t, float64(42), literal.Value, "CILLiteral value should be 42")
}

// TestBuildCILTree_IdentifierInputRef verifies that a postfix with a single
// string identifier produces a CILInputRef with Kind "metric" and the correct
// Name. Also verifies the InputRefs slice on CILMetricNode collects identifiers.
func TestBuildCILTree_IdentifierInputRef(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing an identifier postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:             "test_metric",
		Inputs:           []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixIdentifier{Name: "my_input"}},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["my_input"] = analysis.AnalysisMeasureDefinition{
		Name:     "my_input",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct CIL tree with identifier reference
	require.NoError(t, err, "Analysis should succeed for identifier postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify the root CILNode is a CILInputRef with correct Kind and Name
	inputRef, ok := metric.CILRoot.CILNode.(*cil.CILInputRef)
	require.True(t, ok, "CILNode should be a CILInputRef for identifier postfix")
	assert.Equal(t, "connector", inputRef.Kind, "CILInputRef Kind should be 'connector' for measure lookup")
	assert.Equal(t, "my_input", inputRef.Name, "CILInputRef Name should match identifier")

	// Verify InputRefs collects the identifier
	require.Len(t, metric.CILRoot.InputRefs, 1, "InputRefs should contain one entry")
	assert.Equal(t, "connector", metric.CILRoot.InputRefs[0].Kind)
	assert.Equal(t, "my_input", metric.CILRoot.InputRefs[0].Name)
}

// TestBuildCILTree_BinaryAddition verifies that postfix [42, "identifier", PlusOperatorDefinition]
// produces CILBinaryOp{Op: operator.Add, Left: CILLiteral{42}, Right: CILInputRef{...}}.
func TestBuildCILTree_BinaryAddition(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing binary addition postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixLiteral{Value: 42},
			parsing.PostfixIdentifier{Name: "input_a"},
			parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["input_a"] = analysis.AnalysisMeasureDefinition{
		Name:     "input_a",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct binary addition CIL tree
	require.NoError(t, err, "Analysis should succeed for binary addition postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILBinaryOp with "+" operator
	binOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "CILNode should be a CILBinaryOp")
	assert.Equal(t, operator.Add, binOp.Op, "Operator should be '+'")

	// Verify left operand is CILLiteral{42}
	leftLit, ok := binOp.Left.(*cil.CILLiteral)
	require.True(t, ok, "Left operand should be CILLiteral")
	assert.Equal(t, float64(42), leftLit.Value)

	// Verify right operand is CILInputRef with correct name
	rightRef, ok := binOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Right operand should be CILInputRef")
	assert.Equal(t, "connector", rightRef.Kind)
	assert.Equal(t, "input_a", rightRef.Name)
}

// TestBuildCILTree_BinarySubtraction verifies that postfix [lit1, lit2, MinusOperatorDefinition]
// produces CILBinaryOp{Op: operator.Subtract, Left: CILLiteral{lit1}, Right: CILLiteral{lit2}}.
func TestBuildCILTree_BinarySubtraction(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing binary subtraction postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixLiteral{Value: 10},
			parsing.PostfixLiteral{Value: 3},
			parsing.PostfixOperator{DefType: symboltable.MinusOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct binary subtraction CIL tree
	require.NoError(t, err, "Analysis should succeed for binary subtraction postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILBinaryOp with "-" operator
	binOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "CILNode should be a CILBinaryOp")
	assert.Equal(t, operator.Subtract, binOp.Op, "Operator should be '-'")

	// Verify left operand is CILLiteral{10}
	leftLit, ok := binOp.Left.(*cil.CILLiteral)
	require.True(t, ok, "Left operand should be CILLiteral")
	assert.Equal(t, float64(10), leftLit.Value)

	// Verify right operand is CILLiteral{3}
	rightLit, ok := binOp.Right.(*cil.CILLiteral)
	require.True(t, ok, "Right operand should be CILLiteral")
	assert.Equal(t, float64(3), rightLit.Value)
}

// TestBuildCILTree_BinaryMultiplication verifies that postfix [id_a, id_b, MultiplyOperatorDefinition]
// produces CILBinaryOp{Op: operator.Multiply, Left: CILInputRef{id_a}, Right: CILInputRef{id_b}}.
func TestBuildCILTree_BinaryMultiplication(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing binary multiplication postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixIdentifier{Name: "id_a"},
			parsing.PostfixIdentifier{Name: "id_b"},
			parsing.PostfixOperator{DefType: symboltable.MultiplyOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["id_a"] = analysis.AnalysisMeasureDefinition{
		Name:     "id_a",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}
	defs.Measures["id_b"] = analysis.AnalysisMeasureDefinition{
		Name:     "id_b",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct binary multiplication CIL tree
	require.NoError(t, err, "Analysis should succeed for binary multiplication postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILBinaryOp with "*" operator
	binOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "CILNode should be a CILBinaryOp")
	assert.Equal(t, operator.Multiply, binOp.Op, "Operator should be '*'")

	// Verify left operand is CILInputRef{id_a}
	leftRef, ok := binOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Left operand should be CILInputRef")
	assert.Equal(t, "connector", leftRef.Kind)
	assert.Equal(t, "id_a", leftRef.Name)

	// Verify right operand is CILInputRef{id_b}
	rightRef, ok := binOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Right operand should be CILInputRef")
	assert.Equal(t, "connector", rightRef.Kind)
	assert.Equal(t, "id_b", rightRef.Name)
}

// TestBuildCILTree_BinaryDivision verifies that postfix [lit, id, DivideOperatorDefinition]
// produces CILBinaryOp{Op: operator.Divide, Left: CILLiteral{lit}, Right: CILInputRef{id}}.
func TestBuildCILTree_BinaryDivision(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing binary division postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixLiteral{Value: 100},
			parsing.PostfixIdentifier{Name: "divisor"},
			parsing.PostfixOperator{DefType: symboltable.DivideOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["divisor"] = analysis.AnalysisMeasureDefinition{
		Name:     "divisor",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct binary division CIL tree
	require.NoError(t, err, "Analysis should succeed for binary division postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILBinaryOp with "/" operator
	binOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "CILNode should be a CILBinaryOp")
	assert.Equal(t, operator.Divide, binOp.Op, "Operator should be '/'")

	// Verify left operand is CILLiteral{100}
	leftLit, ok := binOp.Left.(*cil.CILLiteral)
	require.True(t, ok, "Left operand should be CILLiteral")
	assert.Equal(t, float64(100), leftLit.Value)

	// Verify right operand is CILInputRef{divisor}
	rightRef, ok := binOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Right operand should be CILInputRef")
	assert.Equal(t, "connector", rightRef.Kind)
	assert.Equal(t, "divisor", rightRef.Name)
}

// TestBuildCILTree_UnaryNegation verifies that postfix [literal, NegateOperatorDefinition]
// produces CILUnaryOp{Op: operator.Negate, Operand: CILLiteral{...}}.
func TestBuildCILTree_UnaryNegation(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing unary negation postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixLiteral{Value: 5},
			parsing.PostfixOperator{DefType: symboltable.NegateOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct unary negation CIL tree
	require.NoError(t, err, "Analysis should succeed for unary negation postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILUnaryOp with "negate" operator
	unaryOp, ok := metric.CILRoot.CILNode.(*cil.CILUnaryOp)
	require.True(t, ok, "CILNode should be a CILUnaryOp")
	assert.Equal(t, operator.Negate, unaryOp.Op, "Operator should be 'negate'")

	// Verify operand is CILLiteral{5}
	operand, ok := unaryOp.Operand.(*cil.CILLiteral)
	require.True(t, ok, "Operand should be CILLiteral")
	assert.Equal(t, float64(5), operand.Value)
}

// TestBuildCILTree_UnaryNegationOfIdentifier verifies that postfix
// [identifier, NegateOperatorDefinition] produces CILUnaryOp{Op: operator.Negate,
// Operand: CILInputRef{...}}, covering negation of a named input.
func TestBuildCILTree_UnaryNegationOfIdentifier(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing identifier negation postfix
	// Postfix: my_input - (unary)  →  -my_input
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixIdentifier{Name: "my_input"},
			parsing.PostfixOperator{DefType: symboltable.NegateOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["my_input"] = analysis.AnalysisMeasureDefinition{
		Name:     "my_input",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct unary negation of identifier CIL tree
	require.NoError(t, err, "Analysis should succeed for identifier negation postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILUnaryOp with "negate" operator
	unaryOp, ok := metric.CILRoot.CILNode.(*cil.CILUnaryOp)
	require.True(t, ok, "CILNode should be a CILUnaryOp")
	assert.Equal(t, operator.Negate, unaryOp.Op, "Operator should be 'negate'")

	// Verify operand is CILInputRef{my_input}
	operand, ok := unaryOp.Operand.(*cil.CILInputRef)
	require.True(t, ok, "Operand should be CILInputRef")
	assert.Equal(t, "connector", operand.Kind)
	assert.Equal(t, "my_input", operand.Name)
}

// TestBuildCILTree_NegationInExpression verifies that postfix
// ["a", "b", NegateOperatorDefinition, PlusOperatorDefinition]
// (representing a + (-b)) produces correct nested CIL tree where
// the left operand of + is a CILInputRef and the right operand is
// a CILUnaryOp wrapping CILInputRef.
func TestBuildCILTree_NegationInExpression(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing negation in expression postfix
	// Postfix: a b Negate Plus  →  a + (-b)
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixIdentifier{Name: "id_a"},
			parsing.PostfixIdentifier{Name: "id_b"},
			parsing.PostfixOperator{DefType: symboltable.NegateOperatorDefinition},
			parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["id_a"] = analysis.AnalysisMeasureDefinition{
		Name:     "id_a",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}
	defs.Measures["id_b"] = analysis.AnalysisMeasureDefinition{
		Name:     "id_b",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct CIL tree with negation as sub-expression
	require.NoError(t, err, "Analysis should succeed for negation-in-expression postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Level 1: Root is CILBinaryOp{Op: operator.Add}
	rootOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "Root CILNode should be CILBinaryOp")
	assert.Equal(t, operator.Add, rootOp.Op, "Root operator should be '+'")

	// Level 2 (left child): CILInputRef{id_a}
	leftRef, ok := rootOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Left child should be CILInputRef")
	assert.Equal(t, "id_a", leftRef.Name)

	// Level 2 (right child): CILUnaryOp{Op: operator.Negate}
	rightUnary, ok := rootOp.Right.(*cil.CILUnaryOp)
	require.True(t, ok, "Right child should be CILUnaryOp")
	assert.Equal(t, operator.Negate, rightUnary.Op)

	// Level 3 (right operand): CILInputRef{id_b}
	negatedRef, ok := rightUnary.Operand.(*cil.CILInputRef)
	require.True(t, ok, "Negated operand should be CILInputRef")
	assert.Equal(t, "id_b", negatedRef.Name)
}

// TestBuildCILTree_ArithmeticWithPrecedence verifies that postfix
// [id_a, id_b, MultiplyOperatorDefinition, lit, PlusOperatorDefinition]
// (representing a * b + 0) produces a 3-level tree: binary + at root,
// binary * as left child, CILLiteral as right child.
func TestBuildCILTree_ArithmeticWithPrecedence(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing precedence postfix
	// Postfix: id_a id_b * 0 +  →  (id_a * id_b) + 0
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixIdentifier{Name: "id_a"},
			parsing.PostfixIdentifier{Name: "id_b"},
			parsing.PostfixOperator{DefType: symboltable.MultiplyOperatorDefinition},
			parsing.PostfixLiteral{Value: 0},
			parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["id_a"] = analysis.AnalysisMeasureDefinition{
		Name:     "id_a",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}
	defs.Measures["id_b"] = analysis.AnalysisMeasureDefinition{
		Name:     "id_b",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct 3-level CIL tree
	require.NoError(t, err, "Analysis should succeed for precedence postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Level 1: Root is CILBinaryOp{Op: operator.Add}
	rootOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "Root CILNode should be CILBinaryOp")
	assert.Equal(t, operator.Add, rootOp.Op, "Root operator should be '+'")

	// Level 2 (right child): CILLiteral{0}
	rightLit, ok := rootOp.Right.(*cil.CILLiteral)
	require.True(t, ok, "Right child should be CILLiteral")
	assert.Equal(t, float64(0), rightLit.Value)

	// Level 2 (left child): CILBinaryOp{Op: operator.Multiply}
	leftOp, ok := rootOp.Left.(*cil.CILBinaryOp)
	require.True(t, ok, "Left child should be CILBinaryOp")
	assert.Equal(t, operator.Multiply, leftOp.Op, "Left child operator should be '*'")

	// Level 3 (left-left): CILInputRef{id_a}
	leftLeft, ok := leftOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Left-left child should be CILInputRef")
	assert.Equal(t, "id_a", leftLeft.Name)

	// Level 3 (left-right): CILInputRef{id_b}
	leftRight, ok := leftOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Left-right child should be CILInputRef")
	assert.Equal(t, "id_b", leftRight.Name)
}

// TestBuildCILTree_ComplexExpression verifies that postfix
// [id_a, id_b, PlusOperatorDefinition, id_c, id_d, DivideOperatorDefinition, MinusOperatorDefinition]
// (representing (a + b) - (c / d)) produces correct nested CIL tree.
func TestBuildCILTree_ComplexExpression(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing complex expression postfix
	// Postfix: id_a id_b + id_c id_d / -  →  (id_a + id_b) - (id_c / id_d)
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixIdentifier{Name: "id_a"},
			parsing.PostfixIdentifier{Name: "id_b"},
			parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition},
			parsing.PostfixIdentifier{Name: "id_c"},
			parsing.PostfixIdentifier{Name: "id_d"},
			parsing.PostfixOperator{DefType: symboltable.DivideOperatorDefinition},
			parsing.PostfixOperator{DefType: symboltable.MinusOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	for _, name := range []string{"id_a", "id_b", "id_c", "id_d"} {
		defs.Measures[name] = analysis.AnalysisMeasureDefinition{
			Name:     name,
			TypeName: "Number",
			UnitExpression: &analysis.UnitReference{Name: "Unit"},
		}
	}

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct nested CIL tree
	require.NoError(t, err, "Analysis should succeed for complex expression postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Level 1: Root is CILBinaryOp{Op: operator.Subtract}
	rootOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "Root CILNode should be CILBinaryOp")
	assert.Equal(t, operator.Subtract, rootOp.Op, "Root operator should be '-'")

	// Level 2 (left child): CILBinaryOp{Op: operator.Add} representing (id_a + id_b)
	leftOp, ok := rootOp.Left.(*cil.CILBinaryOp)
	require.True(t, ok, "Left child should be CILBinaryOp")
	assert.Equal(t, operator.Add, leftOp.Op, "Left child operator should be '+'")

	// Level 3 (left-left): CILInputRef{id_a}
	leftLeft, ok := leftOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Left-left child should be CILInputRef")
	assert.Equal(t, "id_a", leftLeft.Name)

	// Level 3 (left-right): CILInputRef{id_b}
	leftRight, ok := leftOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Left-right child should be CILInputRef")
	assert.Equal(t, "id_b", leftRight.Name)

	// Level 2 (right child): CILBinaryOp{Op: operator.Divide} representing (id_c / id_d)
	rightOp, ok := rootOp.Right.(*cil.CILBinaryOp)
	require.True(t, ok, "Right child should be CILBinaryOp")
	assert.Equal(t, operator.Divide, rightOp.Op, "Right child operator should be '/'")

	// Level 3 (right-left): CILInputRef{id_c}
	rightLeft, ok := rightOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Right-left child should be CILInputRef")
	assert.Equal(t, "id_c", rightLeft.Name)

	// Level 3 (right-right): CILInputRef{id_d}
	rightRight, ok := rightOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Right-right child should be CILInputRef")
	assert.Equal(t, "id_d", rightRight.Name)
}

// TestBuildCILTree_InputRefsCollection verifies that CILMetricNode.InputRefs
// collects all unique identifier names from the postfix expression.
func TestBuildCILTree_InputRefsCollection(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing multiple identifiers
	// Postfix: id_a id_b + id_a *  →  (id_a + id_b) * id_a
	// Should collect id_a and id_b (id_a appears twice but should be unique)
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixIdentifier{Name: "id_a"},
			parsing.PostfixIdentifier{Name: "id_b"},
			parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition},
			parsing.PostfixIdentifier{Name: "id_a"},
			parsing.PostfixOperator{DefType: symboltable.MultiplyOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["id_a"] = analysis.AnalysisMeasureDefinition{
		Name:     "id_a",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}
	defs.Measures["id_b"] = analysis.AnalysisMeasureDefinition{
		Name:     "id_b",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and collect unique InputRefs
	require.NoError(t, err, "Analysis should succeed for InputRefs collection postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify InputRefs contains both unique identifiers
	require.Len(t, metric.CILRoot.InputRefs, 2, "InputRefs should contain 2 unique identifiers")

	// Collect names for verification
	names := make(map[string]bool)
	for _, ref := range metric.CILRoot.InputRefs {
		assert.Equal(t, "connector", ref.Kind, "InputRef Kind should be 'connector' for measure lookup")
		names[ref.Name] = true
	}
	assert.True(t, names["id_a"], "InputRefs should contain id_a")
	assert.True(t, names["id_b"], "InputRefs should contain id_b")
}

// ============================================================================
// Postfix Evaluator Error Case Tests
// ============================================================================

// TestBuildCILTree_EmptyPostfix verifies that an empty postfix returns an error.
func TestBuildCILTree_EmptyPostfix(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing empty postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:             "test_metric",
		Inputs:           []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should return error for empty postfix
	require.Error(t, err, "Should return error for empty postfix")
	assert.Contains(t, err.Error(), "postfix", "Error should mention postfix")
}

// TestBuildCILTree_UnbalancedStack_BinaryOp verifies that postfix with a single
// operand followed by a binary operator (e.g., [literal, PlusOperatorDefinition])
// returns an error due to stack underflow.
func TestBuildCILTree_UnbalancedStack_BinaryOp(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing underflow postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixLiteral{Value: 42},
			parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should return error for stack underflow on binary operator
	require.Error(t, err, "Should return error for binary operator stack underflow")
	assert.Contains(t, err.Error(), "underflow", "Error should mention underflow")
}

// TestBuildCILTree_UnbalancedStack_UnaryOp verifies that postfix with just
// NegateOperatorDefinition and no operands returns an error.
func TestBuildCILTree_UnbalancedStack_UnaryOp(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing unary underflow postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixOperator{DefType: symboltable.NegateOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should return error for stack underflow on unary operator
	require.Error(t, err, "Should return error for unary operator stack underflow")
	assert.Contains(t, err.Error(), "underflow", "Error should mention underflow")
}

// TestBuildCILTree_ExcessElementsOnStack verifies that postfix with multiple
// operands and no operators returns an error (more than 1 element on stack at end).
func TestBuildCILTree_ExcessElementsOnStack(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing excess elements postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixLiteral{Value: 1},
			parsing.PostfixLiteral{Value: 2},
			parsing.PostfixLiteral{Value: 3},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should return error for excess elements on stack
	require.Error(t, err, "Should return error for excess elements on stack")
	assert.Contains(t, err.Error(), "stack", "Error should mention stack")
}

// ============================================================================
// Postfix Evaluator — Comparison, Logical, and Literal Tests — Phase B
// ============================================================================
//
// These tests verify that buildCILTree correctly handles comparison operators,
// logical operators, and non-integer literals (Nil, True, False) added in
// Story187 Increment 4. Tests use direct symbol table construction following
// the same pattern as the existing arithmetic postfix evaluator tests.
//
// Story: Story187 - CPL Metrics Conditionals Parse and Analyze
// Task: Task 3c Phase B - Logical tests for conditional operators and literals

// TestBuildCILTree_ComparisonEq verifies that postfix [42, 5, EqOperatorDefinition]
// produces CILComparisonOp{Op: operator.Equal, Left: CILLiteral{42}, Right: CILLiteral{5}}.
func TestBuildCILTree_ComparisonEq(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing equality comparison postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixLiteral{Value: 42},
			parsing.PostfixLiteral{Value: 5},
			parsing.PostfixOperator{DefType: symboltable.EqOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct CILComparisonOp tree
	require.NoError(t, err, "Analysis should succeed for equality comparison postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILComparisonOp with "==" operator
	cmpOp, ok := metric.CILRoot.CILNode.(*cil.CILComparisonOp)
	require.True(t, ok, "CILNode should be a CILComparisonOp")
	assert.Equal(t, operator.Equal, cmpOp.Op, "Operator should be '=='")

	// Verify left operand is CILLiteral{42}
	leftLit, ok := cmpOp.Left.(*cil.CILLiteral)
	require.True(t, ok, "Left operand should be CILLiteral")
	assert.Equal(t, float64(42), leftLit.Value)

	// Verify right operand is CILLiteral{5}
	rightLit, ok := cmpOp.Right.(*cil.CILLiteral)
	require.True(t, ok, "Right operand should be CILLiteral")
	assert.Equal(t, float64(5), rightLit.Value)

	// Verify InputRefs is empty (no identifiers in postfix)
	assert.Empty(t, metric.CILRoot.InputRefs, "InputRefs should be empty for literal-only comparison")
}

// TestBuildCILTree_ComparisonLessThan verifies that postfix
// ["id_a", "id_b", LessThanOperatorDefinition] produces
// CILComparisonOp{Op: operator.LessThan, Left: CILInputRef{id_a}, Right: CILInputRef{id_b}}
// and correctly collects both identifiers in InputRefs.
func TestBuildCILTree_ComparisonLessThan(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing less-than comparison postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixIdentifier{Name: "id_a"},
			parsing.PostfixIdentifier{Name: "id_b"},
			parsing.PostfixOperator{DefType: symboltable.LessThanOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["id_a"] = analysis.AnalysisMeasureDefinition{
		Name:     "id_a",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}
	defs.Measures["id_b"] = analysis.AnalysisMeasureDefinition{
		Name:     "id_b",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct CILComparisonOp tree
	require.NoError(t, err, "Analysis should succeed for less-than comparison postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILComparisonOp with "<" operator
	cmpOp, ok := metric.CILRoot.CILNode.(*cil.CILComparisonOp)
	require.True(t, ok, "CILNode should be a CILComparisonOp")
	assert.Equal(t, operator.LessThan, cmpOp.Op, "Operator should be '<'")

	// Verify left operand is CILInputRef{id_a}
	leftRef, ok := cmpOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Left operand should be CILInputRef")
	assert.Equal(t, "id_a", leftRef.Name)

	// Verify right operand is CILInputRef{id_b}
	rightRef, ok := cmpOp.Right.(*cil.CILInputRef)
	require.True(t, ok, "Right operand should be CILInputRef")
	assert.Equal(t, "id_b", rightRef.Name)

	// Verify InputRefs collects both identifiers
	require.Len(t, metric.CILRoot.InputRefs, 2, "InputRefs should contain both identifiers")
}

// TestBuildCILTree_LogicalAnd verifies that postfix
// [True, False, AndOperatorDefinition] produces
// CILLogicalOp{Op: operator.And, Left: CILBoolLiteral{true}, Right: CILBoolLiteral{false}}.
func TestBuildCILTree_LogicalAnd(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing logical AND postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixOperator{DefType: symboltable.TrueLiteralDefinition},
			parsing.PostfixOperator{DefType: symboltable.FalseLiteralDefinition},
			parsing.PostfixOperator{DefType: symboltable.AndOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct CILLogicalOp tree
	require.NoError(t, err, "Analysis should succeed for logical AND postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILLogicalOp with "and" operator
	logOp, ok := metric.CILRoot.CILNode.(*cil.CILLogicalOp)
	require.True(t, ok, "CILNode should be a CILLogicalOp")
	assert.Equal(t, operator.And, logOp.Op, "Operator should be 'and'")

	// Verify left operand is CILBoolLiteral{true}
	leftBool, ok := logOp.Left.(*cil.CILBoolLiteral)
	require.True(t, ok, "Left operand should be CILBoolLiteral")
	assert.True(t, leftBool.Value, "Left value should be true")

	// Verify right operand is CILBoolLiteral{false}
	rightBool, ok := logOp.Right.(*cil.CILBoolLiteral)
	require.True(t, ok, "Right operand should be CILBoolLiteral")
	assert.False(t, rightBool.Value, "Right value should be false")
}

// TestBuildCILTree_LogicalNot verifies that postfix
// [TrueLiteralDefinition, NotOperatorDefinition] produces
// CILNotOp{Operand: CILBoolLiteral{true}}.
func TestBuildCILTree_LogicalNot(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing logical NOT postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixOperator{DefType: symboltable.TrueLiteralDefinition},
			parsing.PostfixOperator{DefType: symboltable.NotOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct CILNotOp tree
	require.NoError(t, err, "Analysis should succeed for logical NOT postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILNotOp
	notOp, ok := metric.CILRoot.CILNode.(*cil.CILNotOp)
	require.True(t, ok, "CILNode should be a CILNotOp")

	// Verify operand is CILBoolLiteral{true}
	operand, ok := notOp.Operand.(*cil.CILBoolLiteral)
	require.True(t, ok, "Operand should be CILBoolLiteral")
	assert.True(t, operand.Value, "Operand value should be true")
}

// TestBuildCILTree_BoolLiteralTrue verifies that postfix [TrueLiteralDefinition]
// produces CILBoolLiteral{Value: true}.
func TestBuildCILTree_BoolLiteralTrue(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing True literal postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:             "test_metric",
		Inputs:           []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixOperator{DefType: symboltable.TrueLiteralDefinition}},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct CIL tree
	require.NoError(t, err, "Analysis should succeed for True literal postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	boolLit, ok := metric.CILRoot.CILNode.(*cil.CILBoolLiteral)
	require.True(t, ok, "CILNode should be a CILBoolLiteral")
	assert.True(t, boolLit.Value, "Bool value should be true")
	assert.Empty(t, metric.CILRoot.InputRefs, "InputRefs should be empty for literal-only postfix")
}

// TestBuildCILTree_BoolLiteralFalse verifies that postfix [FalseLiteralDefinition]
// produces CILBoolLiteral{Value: false}.
func TestBuildCILTree_BoolLiteralFalse(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing False literal postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:             "test_metric",
		Inputs:           []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixOperator{DefType: symboltable.FalseLiteralDefinition}},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct CIL tree
	require.NoError(t, err, "Analysis should succeed for False literal postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	boolLit, ok := metric.CILRoot.CILNode.(*cil.CILBoolLiteral)
	require.True(t, ok, "CILNode should be a CILBoolLiteral")
	assert.False(t, boolLit.Value, "Bool value should be false")
	assert.Empty(t, metric.CILRoot.InputRefs, "InputRefs should be empty for literal-only postfix")
}

// TestBuildCILTree_NilLiteral verifies that postfix [NilLiteralDefinition]
// produces CILNilLiteral{}.
func TestBuildCILTree_NilLiteral(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing Nil literal postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:             "test_metric",
		Inputs:           []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixOperator{DefType: symboltable.NilLiteralDefinition}},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and produce correct CIL tree
	require.NoError(t, err, "Analysis should succeed for Nil literal postfix")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	_, ok := metric.CILRoot.CILNode.(*cil.CILNilLiteral)
	require.True(t, ok, "CILNode should be a CILNilLiteral")
	assert.Empty(t, metric.CILRoot.InputRefs, "InputRefs should be empty for literal-only postfix")
}

// TestBuildCILTree_ComparisonStackUnderflow verifies that postfix with a single
// operand followed by a comparison operator (e.g., [literal, EqOperatorDefinition])
// returns an error due to stack underflow.
func TestBuildCILTree_ComparisonStackUnderflow(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing comparison underflow postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixLiteral{Value: 42},
			parsing.PostfixOperator{DefType: symboltable.EqOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should return error for stack underflow on comparison operator
	require.Error(t, err, "Should return error for comparison operator stack underflow")
	assert.Contains(t, err.Error(), "underflow", "Error should mention underflow")
}

// TestBuildCILTree_LogicalNotStackUnderflow verifies that postfix with just
// NotOperatorDefinition and no operands returns an error.
func TestBuildCILTree_LogicalNotStackUnderflow(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing logical NOT underflow postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixOperator{DefType: symboltable.NotOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should return error for stack underflow on logical NOT operator
	require.Error(t, err, "Should return error for logical NOT operator stack underflow")
	assert.Contains(t, err.Error(), "underflow", "Error should mention underflow")
}

// ============================================================================
// Type Validation (validateExpressionTypes) Integration Tests — Phase B+C
// ============================================================================
//
// These tests verify that type checking of CIL expression trees is integrated
// into AnalyzeMetricScope. Since all existing CIL node types (CILLiteral,
// CILInputRef, CILBinaryOp, CILUnaryOp) are numeric-producing, all valid
// arithmetic expressions should pass type validation.
//
// Tests use programmatic symbol table construction to isolate the type
// validation logic from the parser.
//
// Story: Story184 - CPL Metrics Arithmetic Parse and Analyze
// Task: Task 4e Phase B - Logical tests for type validation

// TestValidateExpressionTypes_LiteralNode verifies that a metric with a
// single CILLiteral in its expression tree passes type validation.
func TestValidateExpressionTypes_LiteralNode(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing a literal postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:             "test_metric",
		Inputs:           []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 42}},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Literal nodes should pass type validation (always numeric)
	require.NoError(t, err, "Analysis should succeed for literal node expression")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
}

// TestValidateExpressionTypes_InputRefNode verifies that a metric with a
// single CILInputRef in its expression tree passes type validation.
func TestValidateExpressionTypes_InputRefNode(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric containing an identifier postfix
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:             "test_metric",
		Inputs:           []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixIdentifier{Name: "my_input"}},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["my_input"] = analysis.AnalysisMeasureDefinition{
		Name:     "my_input",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: InputRef nodes should pass type validation (assumed numeric for Increment 2)
	require.NoError(t, err, "Analysis should succeed for identifier expression")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
}

// TestValidateExpressionTypes_BinaryOp_NumericOperands verifies that a
// binary operation with numeric (literal) operands passes type validation.
func TestValidateExpressionTypes_BinaryOp_NumericOperands(t *testing.T) {
	t.Parallel()

	// Arrange: Postfix: 1 2 + (both operands are CILLiteral, both numeric)
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixLiteral{Value: 1},
			parsing.PostfixLiteral{Value: 2},
			parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Binary op with numeric operands should pass type validation
	require.NoError(t, err, "Analysis should succeed for binary op with numeric operands")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")

	// Verify CIL tree structure
	metric := resultDefs.Metrics["test_metric"]
	binOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "CILNode should be a CILBinaryOp")
	assert.Equal(t, operator.Add, binOp.Op)
}

// TestValidateExpressionTypes_UnaryOp_NumericOperand verifies that a
// unary negation with a numeric (literal) operand passes type validation.
func TestValidateExpressionTypes_UnaryOp_NumericOperand(t *testing.T) {
	t.Parallel()

	// Arrange: Postfix: 5 negate (operand is CILLiteral, numeric)
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixLiteral{Value: 5},
			parsing.PostfixOperator{DefType: symboltable.NegateOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Unary op with numeric operand should pass type validation
	require.NoError(t, err, "Analysis should succeed for unary op with numeric operand")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")

	// Verify CIL tree structure
	metric := resultDefs.Metrics["test_metric"]
	unaryOp, ok := metric.CILRoot.CILNode.(*cil.CILUnaryOp)
	require.True(t, ok, "CILNode should be a CILUnaryOp")
	assert.Equal(t, operator.Negate, unaryOp.Op)
}

// TestValidateExpressionTypes_NestedExpression verifies that a deeply
// nested expression with all numeric nodes passes type validation.
// Postfix: id_a 10 * id_b negate +  →  (id_a * 10) + (-id_b)
func TestValidateExpressionTypes_NestedExpression(t *testing.T) {
	t.Parallel()

	// Arrange: Build a deeply nested expression: (id_a * 10) + (-id_b)
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:   "test_metric",
		Inputs: []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{
			parsing.PostfixIdentifier{Name: "id_a"},
			parsing.PostfixLiteral{Value: 10},
			parsing.PostfixOperator{DefType: symboltable.MultiplyOperatorDefinition},
			parsing.PostfixIdentifier{Name: "id_b"},
			parsing.PostfixOperator{DefType: symboltable.NegateOperatorDefinition},
			parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition},
		},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["id_a"] = analysis.AnalysisMeasureDefinition{
		Name:     "id_a",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}
	defs.Measures["id_b"] = analysis.AnalysisMeasureDefinition{
		Name:     "id_b",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Deeply nested expression with all numeric nodes should pass type validation
	require.NoError(t, err, "Analysis should succeed for nested expression")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")

	// Verify CIL tree structure: root should be +, left should be *, right should be negate
	metric := resultDefs.Metrics["test_metric"]
	rootOp, ok := metric.CILRoot.CILNode.(*cil.CILBinaryOp)
	require.True(t, ok, "Root should be CILBinaryOp")
	assert.Equal(t, operator.Add, rootOp.Op)

	// Left child: (id_a * 10)
	leftOp, ok := rootOp.Left.(*cil.CILBinaryOp)
	require.True(t, ok, "Left child should be CILBinaryOp (multiplication)")
	assert.Equal(t, operator.Multiply, leftOp.Op)

	// Right child: (-id_b)
	rightUnary, ok := rootOp.Right.(*cil.CILUnaryOp)
	require.True(t, ok, "Right child should be CILUnaryOp (negation)")
	assert.Equal(t, operator.Negate, rightUnary.Op)
}

// ============================================================================
// IO Specification Validation Tests
// ============================================================================

// TestValidateAndStoreIOSpec_ValidOutputs verifies that an IO spec whose output
// metric names all exist in declared metrics is accepted and stored on the analyzer.
func TestValidateAndStoreIOSpec_ValidOutputs(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with declared metrics
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Metrics["revenue_forecast"] = analysis.AnalysisMetricDefinition{Name: "revenue_forecast"}
	defs.Metrics["profit_margin"] = analysis.AnalysisMetricDefinition{Name: "profit_margin"}

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{MetricName: "revenue_forecast"},
			{MetricName: "profit_margin"},
		},
	}

	// Act: Validate and store the IO spec
	err := analyzer.ValidateAndStoreIOSpec(ioSpec)

	// Assert: Should succeed and store the IO spec
	require.NoError(t, err, "Validation should succeed when all output metrics exist")
	stored := analyzer.GetIOSpec()
	require.NotNil(t, stored, "IO spec should be stored after successful validation")
	assert.Len(t, stored.Outputs, 2, "Stored IO spec should have 2 outputs")
	assert.Equal(t, "revenue_forecast", stored.Outputs[0].MetricName)
	assert.Equal(t, "profit_margin", stored.Outputs[1].MetricName)
}

// TestValidateAndStoreIOSpec_UnknownMetric verifies that referencing an output
// metric that does not exist in declared metrics returns an error.
func TestValidateAndStoreIOSpec_UnknownMetric(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with only one declared metric
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Metrics["revenue_forecast"] = analysis.AnalysisMetricDefinition{Name: "revenue_forecast"}

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{MetricName: "revenue_forecast"},
			{MetricName: "unknown_metric"},
		},
	}

	// Act: Validate the IO spec
	err := analyzer.ValidateAndStoreIOSpec(ioSpec)

	// Assert: Should return error referencing the unknown metric
	require.Error(t, err, "Validation should fail for unknown output metric")
	assert.Contains(t, err.Error(), "unknown metric", "Error should mention unknown metric")
	assert.Contains(t, err.Error(), "unknown_metric", "Error should contain the metric name")

	// Verify IO spec was NOT stored
	stored := analyzer.GetIOSpec()
	assert.Nil(t, stored, "IO spec should not be stored after validation failure")
}

// TestValidateAndStoreIOSpec_EmptyOutputs verifies that an IO spec with an
// empty outputs list is accepted (all zero outputs are trivially valid).
func TestValidateAndStoreIOSpec_EmptyOutputs(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with no declared metrics
	analyzer := analysis.NewCPLAnalyzer()

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{},
	}

	// Act: Validate and store the empty IO spec
	err := analyzer.ValidateAndStoreIOSpec(ioSpec)

	// Assert: Empty outputs list should pass validation
	require.NoError(t, err, "Validation should succeed for empty outputs list")
	stored := analyzer.GetIOSpec()
	require.NotNil(t, stored, "IO spec should be stored after successful validation")
	assert.Empty(t, stored.Outputs, "Stored IO spec should have empty outputs")
}

// ============================================================================
// resolveInputRefKind Tests
// ============================================================================
//
// These tests verify that resolveInputRefKind correctly classifies identifier
// references in metric calculate clauses into their appropriate input kinds.
// The function is exercised indirectly through buildCILTree/AnalyzeMetricScope
// since it is unexported.
//
// Story: Story205 - Fix resolveInputRefKind to check analyzer.connectors
// Task: Task 1 Phase B - Logical tests for resolveInputRefKind

// TestResolveInputRefKind_MeasureLookup verifies that when an identifier name
// exists in analyzer.measures, resolveInputRefKind returns "connector".
func TestResolveInputRefKind_MeasureLookup(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric referencing a measure name
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:             "test_metric",
		Inputs:           []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixIdentifier{Name: "measure_input"}},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["measure_input"] = analysis.AnalysisMeasureDefinition{
		Name:     "measure_input",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and CILInputRef Kind should be "connector"
	require.NoError(t, err, "Analysis should succeed for measure lookup")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	inputRef, ok := metric.CILRoot.CILNode.(*cil.CILInputRef)
	require.True(t, ok, "CILNode should be a CILInputRef")
	assert.Equal(t, "connector", inputRef.Kind, "CILInputRef Kind should be 'connector' for measure lookup")
	assert.Equal(t, "measure_input", inputRef.Name, "CILInputRef Name should match")
}

// TestResolveInputRefKind_ConnectorLookup verifies that when an identifier name
// exists in analyzer.connectors, resolveInputRefKind returns "connector".
func TestResolveInputRefKind_ConnectorLookup(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric referencing a connector name
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:             "test_metric",
		Inputs:           []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixIdentifier{Name: "connector_input"}},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Connectors["connector_input"] = analysis.AnalysisConnectorDefinition{Name: "connector_input"}

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and CILInputRef Kind should be "connector"
	require.NoError(t, err, "Analysis should succeed for connector lookup")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	inputRef, ok := metric.CILRoot.CILNode.(*cil.CILInputRef)
	require.True(t, ok, "CILNode should be a CILInputRef")
	assert.Equal(t, "connector", inputRef.Kind, "CILInputRef Kind should be 'connector' for connector lookup")
	assert.Equal(t, "connector_input", inputRef.Name, "CILInputRef Name should match")
}

// TestResolveInputRefKind_MetricLookup verifies that when an identifier name
// exists in analyzer.metrics, resolveInputRefKind returns "metric".
func TestResolveInputRefKind_MetricLookup(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric referencing another metric name
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:             "test_metric",
		Inputs:           []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixIdentifier{Name: "metric_input"}},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Metrics["metric_input"] = analysis.AnalysisMetricDefinition{Name: "metric_input"}

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should succeed and CILInputRef Kind should be "metric"
	require.NoError(t, err, "Analysis should succeed for metric lookup")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "test_metric")
	metric := resultDefs.Metrics["test_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	inputRef, ok := metric.CILRoot.CILNode.(*cil.CILInputRef)
	require.True(t, ok, "CILNode should be a CILInputRef")
	assert.Equal(t, "metric", inputRef.Kind, "CILInputRef Kind should be 'metric' for metric lookup")
	assert.Equal(t, "metric_input", inputRef.Name, "CILInputRef Name should match")
}

// ============================================================================
// Conditional Expression Integration Tests (Parse → Analyze Pipeline)
// ============================================================================
//
// These tests exercise the full parse→analyze pipeline for metrics with
// conditional expressions (if/else/when/default) in their calculate clauses.
// They verify that the parser correctly produces clause definitions in child
// scopes and the analyzer consumes them via assembleConditional to build
// correct CILConditional or CILWhenClause trees.
//
// Story: Story187 - CPL Metrics Conditionals Parse and Analyze
// Task: Integration verification (end-to-end pipeline)

// TestParseAndAnalyze_IfElseConditional verifies the full pipeline for a
// simple if/else conditional expression: calculate if x > 0 then 1 else 2.
func TestParseAndAnalyze_IfElseConditional(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with a simple if/else conditional expression
	input := `metrics
	cond_metric
		inputs
			metric x
		calculate
			if
				x > 0
					1 else
						2 end if`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Create analyzer with required measures for input validation
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["x"] = analysis.AnalysisMeasureDefinition{
		Name:     "x",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze the parsed metric
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Analysis should succeed and produce correct CILConditional tree
	require.NoError(t, err, "Analysis should succeed for if/else conditional")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "cond_metric")
	metric := resultDefs.Metrics["cond_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILConditional
	cond, ok := metric.CILRoot.CILNode.(*cil.CILConditional)
	require.True(t, ok, "Root CILNode should be CILConditional")

	// Should have 2 branches (if + else)
	require.Len(t, cond.Branches, 2, "CILConditional should have 2 branches")

	// First branch (if x > 0): condition should be CILComparisonOp{x > 0}
	require.NotNil(t, cond.Branches[0].Condition, "First branch should have condition")
	cmpOp, ok := cond.Branches[0].Condition.(*cil.CILComparisonOp)
	require.True(t, ok, "Condition should be CILComparisonOp")
	assert.Equal(t, operator.GreaterThan, cmpOp.Op, "Comparison operator should be '>'")

	// First branch body should be CILLiteral{1}
	require.NotNil(t, cond.Branches[0].Body, "First branch should have body")
	bodyLit, ok := cond.Branches[0].Body.(*cil.CILLiteral)
	require.True(t, ok, "First branch body should be CILLiteral")
	assert.Equal(t, float64(1), bodyLit.Value)

	// Second branch (else): should have nil condition
	assert.Nil(t, cond.Branches[1].Condition, "Else branch should have nil condition")

	// Second branch body should be CILLiteral{2}
	require.NotNil(t, cond.Branches[1].Body, "Else branch should have body")
	elseLit, ok := cond.Branches[1].Body.(*cil.CILLiteral)
	require.True(t, ok, "Else branch body should be CILLiteral")
	assert.Equal(t, float64(2), elseLit.Value)

	// Verify InputRefs collection (should contain "x")
	require.Len(t, metric.CILRoot.InputRefs, 1, "InputRefs should contain x")
	assert.Equal(t, "x", metric.CILRoot.InputRefs[0].Name)
}

// TestParseAndAnalyze_IfElseIfElseConditional verifies the full pipeline for
// an if/else-if/else conditional with multiple comparison operators.
func TestParseAndAnalyze_IfElseIfElseConditional(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with if/else-if/else conditional
	input := `metrics
	multi_metric
		inputs
			metric x
			metric y
		calculate
			if
				x > 10
					1 else if y > 0
						2 else
							3 end if`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["x"] = analysis.AnalysisMeasureDefinition{
		Name: "x", TypeName: "Number", UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}
	defs.Measures["y"] = analysis.AnalysisMeasureDefinition{
		Name: "y", TypeName: "Number", UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert
	require.NoError(t, err, "Analysis should succeed for if/else-if/else conditional")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "multi_metric")
	metric := resultDefs.Metrics["multi_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	cond, ok := metric.CILRoot.CILNode.(*cil.CILConditional)
	require.True(t, ok, "Root CILNode should be CILConditional")

	// Should have 3 branches (if + else-if + else)
	require.Len(t, cond.Branches, 3, "CILConditional should have 3 branches")

	// First branch: if x > 10 → 1
	require.NotNil(t, cond.Branches[0].Condition, "If branch should have condition")
	require.NotNil(t, cond.Branches[0].Body, "If branch should have body")

	// Second branch: else if y > 0 → 2
	require.NotNil(t, cond.Branches[1].Condition, "Else-if branch should have condition")
	require.NotNil(t, cond.Branches[1].Body, "Else-if branch should have body")

	// Third branch: else → 3
	assert.Nil(t, cond.Branches[2].Condition, "Else branch should have nil condition")
	require.NotNil(t, cond.Branches[2].Body, "Else branch should have body")

	// Verify InputRefs contains both x and y
	require.Len(t, metric.CILRoot.InputRefs, 2, "InputRefs should contain x and y")
}

// TestParseAndAnalyze_WhenDefault verifies the full pipeline for a
// simple when/default conditional expression.
func TestParseAndAnalyze_WhenDefault(t *testing.T) {
	t.Parallel()

	// Arrange: Parse a metric with when/default conditional
	input := `metrics
	when_metric
		inputs
			metric x
		calculate
			when
				x > 0 =>
					1 default =>
						0 end when`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["x"] = analysis.AnalysisMeasureDefinition{
		Name: "x", TypeName: "Number", UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)

	// Assert: Analysis should succeed and produce CILWhenClause tree
	require.NoError(t, err, "Analysis should succeed for when/default conditional")
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "when_metric")
	metric := resultDefs.Metrics["when_metric"]
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")

	// Verify root is CILWhenClause
	whenClause, ok := metric.CILRoot.CILNode.(*cil.CILWhenClause)
	require.True(t, ok, "Root CILNode should be CILWhenClause")

	// Should have 1 when case + default
	require.Len(t, whenClause.Clauses, 1, "CILWhenClause should have 1 clause")

	// Verify when clause has condition x > 0
	require.NotNil(t, whenClause.Clauses[0].Condition, "When clause should have condition")
	cmpOp, ok := whenClause.Clauses[0].Condition.(*cil.CILComparisonOp)
	require.True(t, ok, "When condition should be CILComparisonOp")
	assert.Equal(t, operator.GreaterThan, cmpOp.Op)

	// Verify when clause body is CILLiteral{1}
	require.NotNil(t, whenClause.Clauses[0].Body, "When clause should have body")
	bodyLit, ok := whenClause.Clauses[0].Body.(*cil.CILLiteral)
	require.True(t, ok, "When body should be CILLiteral")
	assert.Equal(t, float64(1), bodyLit.Value)

	// Verify default body is CILLiteral{0}
	require.NotNil(t, whenClause.Default, "Default body should not be nil")
	defaultLit, ok := whenClause.Default.(*cil.CILLiteral)
	require.True(t, ok, "Default body should be CILLiteral")
	assert.Equal(t, float64(0), defaultLit.Value)

	// Verify InputRefs collection
	require.Len(t, metric.CILRoot.InputRefs, 1, "InputRefs should contain x")
	assert.Equal(t, "x", metric.CILRoot.InputRefs[0].Name)
}

// TestResolveInputRefKind_UnknownNameFallback verifies that when an identifier
// name is not found in measures, connectors, or metrics, resolveInputRefKind
// returns an error instead of silently defaulting to "metric".
func TestResolveInputRefKind_UnknownNameFallback(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with a metric referencing an unknown name
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name:             "test_metric",
		Inputs:           []parsing.ParserInputDefinition{},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixIdentifier{Name: "unknown_input"}},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()

	// Act: Analyze the metric scope
	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

	// Assert: Should return an error for undefined reference
	require.Error(t, err, "Analysis should return error for undefined reference")
	assert.Contains(t, err.Error(), "undefined reference", "Error should mention undefined reference")
	assert.Contains(t, err.Error(), "unknown_input", "Error should contain the undefined name")
}

// ============================================================================
// Alias and Where Clause Analysis Tests
// ============================================================================
//
// These tests verify that the analysis phase correctly resolves aliases and
// where clauses from parser input definitions into analysis input definitions.
//
// Story: Story189 - CPL Metrics Aliases and Where Clause Parse and Analyze
// Task: Task 3 - Implement analysis phase for alias and where clause validation

// TestAnalyzeMetricScope_WithAlias verifies that a metric input with an 'as'
// alias produces an AnalysisInputDefinition with the Alias field set.
func TestAnalyzeMetricScope_WithAlias(t *testing.T) {
	t.Parallel()
	input := `metrics
	my_metric
		inputs
			metric demand_forecast as df
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["demand_forecast"] = analysis.AnalysisMeasureDefinition{
		Name:     "demand_forecast",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)
	require.NoError(t, err)

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "my_metric")
	metric := resultDefs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 1)
	assert.Equal(t, "df", metric.Inputs[0].Alias)
	assert.Equal(t, "demand_forecast", metric.Inputs[0].Name)
}

// TestAnalyzeMetricScope_DuplicateAlias verifies that duplicate aliases in
// a metric's inputs produce an error.
func TestAnalyzeMetricScope_DuplicateAlias(t *testing.T) {
	t.Parallel()

	// Arrange: Build symbol table with duplicate aliases programmatically
	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name: "my_metric",
		Inputs: []parsing.ParserInputDefinition{
			{Kind: "metric", Name: "demand_forecast", Alias: "df"},
			{Kind: "metric", Name: "actuals", Alias: "df"},
		},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 0}},
		SourceLocation:   symboltable.SourceLocation{Line: 2, Column: 1},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["demand_forecast"] = analysis.AnalysisMeasureDefinition{
		Name:     "demand_forecast",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}
	defs.Measures["actuals"] = analysis.AnalysisMeasureDefinition{
		Name:     "actuals",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}

	// Act: Analyze the metric
	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	// Assert: Should return error for duplicate alias
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate alias 'df'")
	assert.Contains(t, err.Error(), "my_metric")
}

// TestAnalyzeMetricScope_WithWhereClause verifies that a metric input with a
// where clause produces an AnalysisInputDefinition with a non-nil WhereClause.
func TestAnalyzeMetricScope_WithWhereClause(t *testing.T) {
	t.Parallel()
	input := `metrics
	my_metric
		inputs
			metric demand_forecast where aspect=actuals
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["demand_forecast"] = analysis.AnalysisMeasureDefinition{
		Name:     "demand_forecast",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)
	require.NoError(t, err)

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "my_metric")
	metric := resultDefs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 1)
	require.NotNil(t, metric.Inputs[0].WhereClause)
	require.Len(t, metric.Inputs[0].WhereClause.AndGroups, 1)
	require.Len(t, metric.Inputs[0].WhereClause.AndGroups[0].Conditions, 1)
	assert.Equal(t, "ACTUALS", metric.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect)
	assert.False(t, metric.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Negated)
}

// TestAnalyzeMetricScope_WithWhereNotClause verifies that a metric input with
// a 'where not' clause produces an AnalysisWhereCondition with Negated=true.
func TestAnalyzeMetricScope_WithWhereNotClause(t *testing.T) {
	t.Parallel()
	input := `metrics
	my_metric
		inputs
			metric demand_forecast where not aspect=forecast
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["demand_forecast"] = analysis.AnalysisMeasureDefinition{
		Name:     "demand_forecast",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)
	require.NoError(t, err)

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "my_metric")
	metric := resultDefs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 1)
	require.NotNil(t, metric.Inputs[0].WhereClause)
	require.Len(t, metric.Inputs[0].WhereClause.AndGroups, 1)
	require.Len(t, metric.Inputs[0].WhereClause.AndGroups[0].Conditions, 1)
	assert.True(t, metric.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Negated)
	assert.Equal(t, "FORECAST", metric.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect)
}

// TestAnalyzeMetricScope_WithComplexWhere verifies that a metric input with
// multiple where conditions connected by logical operators is parsed correctly.
func TestAnalyzeMetricScope_WithComplexWhere(t *testing.T) {
	t.Parallel()
	input := `metrics
	my_metric
		inputs
			metric demand_forecast where aspect=actuals and aspect=forecast
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["demand_forecast"] = analysis.AnalysisMeasureDefinition{
		Name:     "demand_forecast",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)
	require.NoError(t, err)

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "my_metric")
	metric := resultDefs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 1)
	require.NotNil(t, metric.Inputs[0].WhereClause)
	require.Len(t, metric.Inputs[0].WhereClause.AndGroups, 1)
	require.Len(t, metric.Inputs[0].WhereClause.AndGroups[0].Conditions, 2)
	assert.Equal(t, "ACTUALS", metric.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect)
	assert.False(t, metric.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Negated)
	assert.Equal(t, "FORECAST", metric.Inputs[0].WhereClause.AndGroups[0].Conditions[1].Aspect)
	assert.False(t, metric.Inputs[0].WhereClause.AndGroups[0].Conditions[1].Negated)
}

// TestAnalyzeMetricScope_AliasDefaultsToName verifies that a metric input
// without an alias has its Alias field set to the input Name.
func TestAnalyzeMetricScope_AliasDefaultsToName(t *testing.T) {
	t.Parallel()
	input := `metrics
	my_metric
		inputs
			metric demand_forecast
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["demand_forecast"] = analysis.AnalysisMeasureDefinition{
		Name:     "demand_forecast",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)
	require.NoError(t, err)

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "my_metric")
	metric := resultDefs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 1)
	assert.Equal(t, "demand_forecast", metric.Inputs[0].Alias)
	assert.Equal(t, "demand_forecast", metric.Inputs[0].Name)
}

// ============================================================================
// End-to-End Alias and Where Clause Integration Test
// ============================================================================
//
// This test verifies the full parse→analyze pipeline for metrics using
// aliases and where clauses. It ensures that:
//   - Parser correctly produces ParserInputDefinition with Alias and WhereClause
//   - Analyzer correctly resolves them into AnalysisInputDefinition
//   - Both Alias and WhereClause fields are populated in the final analysis result
//
// Story: Story189 - CPL Metrics Aliases and Where Clause Parse and Analyze
// Task: Task 4 - Integration verification

// TestAnalyzeMetricScope_EndToEnd_AliasAndWhere verifies the full parse-analyze
// pipeline for a metric with aliased inputs and where clauses.
func TestAnalyzeMetricScope_EndToEnd_AliasAndWhere(t *testing.T) {
	t.Parallel()

	input := `concepts
	Region = [Values[North, South]]
	Price = [Values[Low, High]]
	Package = [Values[Each]]
	Day = [Values[Monday]]

units
	Package=Package
	Region=Region
	Day=Day

types
	Count = Number(precision=0)

measures
	packages = Count(unit=Package)
	Daily = Interval(unit=Day)

connectors
	demand_fc = dimensions [Region] grain Daily celltype packages
	price_dim = dimensions [Region] grain Daily celltype packages

metrics
	weighted_demand
		inputs
			metric demand_fc as demand where aspect=actuals
			dimension Price as price where aspect=forecast
		calculate
			0`

	// Act: Parse and analyze the full document
	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Full analysis should succeed")
	require.NotNil(t, defs, "Analysis definitions should not be nil")

	// Assert: Metric should be present with correct inputs
	require.Contains(t, defs.Metrics, "weighted_demand", "Metric should be in results")
	metric := defs.Metrics["weighted_demand"]
	assert.Equal(t, "weighted_demand", metric.Name)
	require.Len(t, metric.Inputs, 2, "Metric should have 2 inputs")

	// Verify first input: metric with alias and where clause
	firstInput := metric.Inputs[0]
	assert.Equal(t, "metric", firstInput.Kind, "First input kind should be 'metric'")
	assert.Equal(t, "demand_fc", firstInput.Name, "First input name should be 'demand_fc'")
	assert.Equal(t, "demand", firstInput.Alias, "First input alias should be 'demand'")
	require.NotNil(t, firstInput.WhereClause, "First input should have a WhereClause")
	require.Len(t, firstInput.WhereClause.AndGroups, 1, "First input should have 1 and-group")
	require.Len(t, firstInput.WhereClause.AndGroups[0].Conditions, 1, "First input should have 1 where condition")
	assert.Equal(t, "ACTUALS", firstInput.WhereClause.AndGroups[0].Conditions[0].Aspect)
	assert.False(t, firstInput.WhereClause.AndGroups[0].Conditions[0].Negated)

	// Verify second input: dimension with alias and where clause
	secondInput := metric.Inputs[1]
	assert.Equal(t, "dimension", secondInput.Kind, "Second input kind should be 'dimension'")
	assert.Equal(t, "Price", secondInput.Name, "Second input name should be 'Price'")
	assert.Equal(t, "price", secondInput.Alias, "Second input alias should be 'price'")
	require.NotNil(t, secondInput.WhereClause, "Second input should have a WhereClause")
	require.Len(t, secondInput.WhereClause.AndGroups, 1, "Second input should have 1 and-group")
	require.Len(t, secondInput.WhereClause.AndGroups[0].Conditions, 1, "Second input should have 1 where condition")
	assert.Equal(t, "FORECAST", secondInput.WhereClause.AndGroups[0].Conditions[0].Aspect)
	assert.False(t, secondInput.WhereClause.AndGroups[0].Conditions[0].Negated)

	// Verify CILRoot is present
	require.NotNil(t, metric.CILRoot, "CILRoot should not be nil")
	literal, ok := metric.CILRoot.CILNode.(*cil.CILLiteral)
	require.True(t, ok, "CILNode should be CILLiteral for literal-only metric")
	assert.Equal(t, float64(0), literal.Value)
}

// ============================================================================
// Where Clause Precedence Tests
// ============================================================================

// TestAnalyzeMetricScope_OrPrecedence verifies that `where aspect=A or aspect=B`
// produces two AndGroups, each with one condition.
func TestAnalyzeMetricScope_OrPrecedence(t *testing.T) {
	t.Parallel()
	input := `metrics
	my_metric
		inputs
			metric demand_forecast where aspect=actuals or aspect=forecast
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["demand_forecast"] = analysis.AnalysisMeasureDefinition{
		Name:     "demand_forecast",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)
	require.NoError(t, err)

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "my_metric")
	metric := resultDefs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 1)
	require.NotNil(t, metric.Inputs[0].WhereClause)
	require.Len(t, metric.Inputs[0].WhereClause.AndGroups, 2, "OR should produce 2 AndGroups")
	require.Len(t, metric.Inputs[0].WhereClause.AndGroups[0].Conditions, 1)
	assert.Equal(t, "ACTUALS", metric.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect)
	require.Len(t, metric.Inputs[0].WhereClause.AndGroups[1].Conditions, 1)
	assert.Equal(t, "FORECAST", metric.Inputs[0].WhereClause.AndGroups[1].Conditions[0].Aspect)
}

// TestAnalyzeMetricScope_AndBindsTighter verifies that `where aspect=A and aspect=B or aspect=C`
// produces AndGroup 0 with [A, B] and AndGroup 1 with [C], because AND binds tighter than OR.
func TestAnalyzeMetricScope_AndBindsTighter(t *testing.T) {
	t.Parallel()
	input := `metrics
	my_metric
		inputs
			metric demand_forecast where aspect=actuals and aspect=forecast or aspect=actuals
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["demand_forecast"] = analysis.AnalysisMeasureDefinition{
		Name:     "demand_forecast",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)
	require.NoError(t, err)

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "my_metric")
	metric := resultDefs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 1)
	require.NotNil(t, metric.Inputs[0].WhereClause)
	require.Len(t, metric.Inputs[0].WhereClause.AndGroups, 2)
	require.Len(t, metric.Inputs[0].WhereClause.AndGroups[0].Conditions, 2)
	assert.Equal(t, "ACTUALS", metric.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect)
	assert.Equal(t, "FORECAST", metric.Inputs[0].WhereClause.AndGroups[0].Conditions[1].Aspect)
	require.Len(t, metric.Inputs[0].WhereClause.AndGroups[1].Conditions, 1)
	assert.Equal(t, "ACTUALS", metric.Inputs[0].WhereClause.AndGroups[1].Conditions[0].Aspect)
}

// TestAnalyzeMetricScope_AndBindsTighterReversed verifies that `where aspect=A or aspect=B and aspect=C`
// produces AndGroup 0 with [A] and AndGroup 1 with [B, C], because AND binds tighter than OR.
func TestAnalyzeMetricScope_AndBindsTighterReversed(t *testing.T) {
	t.Parallel()
	input := `metrics
	my_metric
		inputs
			metric demand_forecast where aspect=actuals or aspect=forecast and aspect=actuals
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["demand_forecast"] = analysis.AnalysisMeasureDefinition{
		Name:     "demand_forecast",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
		symboltable.MetricScope:  analysis.AnalyzeMetricScope,
	}
	err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)
	require.NoError(t, err)

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "my_metric")
	metric := resultDefs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 1)
	require.NotNil(t, metric.Inputs[0].WhereClause)
	require.Len(t, metric.Inputs[0].WhereClause.AndGroups, 2)
	require.Len(t, metric.Inputs[0].WhereClause.AndGroups[0].Conditions, 1)
	assert.Equal(t, "ACTUALS", metric.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect)
	require.Len(t, metric.Inputs[0].WhereClause.AndGroups[1].Conditions, 2)
	assert.Equal(t, "FORECAST", metric.Inputs[0].WhereClause.AndGroups[1].Conditions[0].Aspect)
	assert.Equal(t, "ACTUALS", metric.Inputs[0].WhereClause.AndGroups[1].Conditions[1].Aspect)
}
