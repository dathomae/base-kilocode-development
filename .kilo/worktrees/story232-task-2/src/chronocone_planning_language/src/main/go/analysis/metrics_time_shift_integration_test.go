// Package analysis_test contains integration tests for CPL metric time shifting.
//
// These tests verify the complete parse→analysis pipeline for time-shifted
// metric inputs, including:
//   - Parser produces correct ParserTimeShift values
//   - Analysis validates constraint field references
//   - Self-reference semantics (allowed with time shift, prohibited without)
//   - Edge cases (multiple inputs, mixed configurations, negative amounts)
//
// Story: Story191 - CPL Metrics Time Shift Parse and Analyze
// Task: Task 6 - Integration verification
package analysis_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parsing"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTimeShiftExamplesDir returns the path to the metrics_time_shift examples directory.
func getTimeShiftExamplesDir(t *testing.T) string {
	t.Helper()
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	projectRoot := filepath.Join(testDir, "..", "..", "..", "..")
	return filepath.Join(projectRoot, "examples", "parse_and_analyze_only", "without_errors", "metrics_time_shift")
}

// requireParsedMetricDefinition extracts a ParserMetricDefinition from a parsed document's MetricsScope.
func requireParsedMetricDefinition(t *testing.T, doc *symboltable.SymbolTable) *parsing.ParserMetricDefinition {
	t.Helper()
	require.NotNil(t, doc, "Document should not be nil")

	var metricsScope *symboltable.SymbolTable
	for _, child := range doc.Children {
		if child.ScopeType == symboltable.MetricsScope {
			metricsScope = child
			break
		}
	}
	require.NotNil(t, metricsScope, "Expected to find a MetricsScope among root children")
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")
	metricScope := metricsScope.Children[0]
	require.NotNil(t, metricScope.Data, "MetricScope.Data should be populated")
	metricDef, ok := metricScope.Data.(*parsing.ParserMetricDefinition)
	require.True(t, ok, "MetricScope.Data should be a *parsing.ParserMetricDefinition")
	return metricDef
}

// ============================================================================
// 6b) Verify parse stage produces correct ParserTimeShift values
// ============================================================================

// TestParseTimeShift_BackwardYear_DefaultAmount verifies parsing backward(year)
// produces ParserTimeShift with Direction="backward", Field="year", Amount=1.0.
func TestParseTimeShift_BackwardYear_DefaultAmount(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x backward(year)
		calculate
			0`

	doc, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	metricDef := requireParsedMetricDefinition(t, doc)
	require.Len(t, metricDef.Inputs, 1, "Should have 1 input")

	require.NotNil(t, metricDef.Inputs[0].TimeShift, "TimeShift should not be nil")
	assert.Equal(t, "backward", metricDef.Inputs[0].TimeShift.Direction)
	assert.Equal(t, "year", metricDef.Inputs[0].TimeShift.Field)
	assert.Equal(t, 1.0, metricDef.Inputs[0].TimeShift.Amount, "Default amount should be 1.0")
}

// TestParseTimeShift_BackwardDay_ExplicitAmount verifies parsing backward(day, 2)
// produces ParserTimeShift with Amount=2.0.
func TestParseTimeShift_BackwardDay_ExplicitAmount(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x backward(day, 2)
		calculate
			0`

	doc, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	metricDef := requireParsedMetricDefinition(t, doc)
	require.Len(t, metricDef.Inputs, 1)

	require.NotNil(t, metricDef.Inputs[0].TimeShift)
	assert.Equal(t, "backward", metricDef.Inputs[0].TimeShift.Direction)
	assert.Equal(t, "day", metricDef.Inputs[0].TimeShift.Field)
	require.NotNil(t, metricDef.Inputs[0].TimeShift.Expression, "Expression should not be nil for backward(day, 2)")
	assert.Len(t, metricDef.Inputs[0].TimeShift.Expression, 1, "Expression should have 1 element")
	if lit, ok := metricDef.Inputs[0].TimeShift.Expression[0].(parsing.PostfixLiteral); ok {
		assert.InDelta(t, 2.0, lit.Value, 0.001, "Expression literal value should be 2.0")
	} else {
		t.Fatalf("Expected PostfixLiteral, got %T", metricDef.Inputs[0].TimeShift.Expression[0])
	}
	assert.Equal(t, 1.0, metricDef.Inputs[0].TimeShift.Amount, "Amount should default to 1.0")
}

// TestAnalyzeTimeShift_ArithmeticExpressionInShiftAmount verifies that an
// arithmetic expression in the shift amount (e.g., backward(day, 1 + 1))
// is correctly parsed as a multi-element postfix sequence and evaluated
// to the correct float64 value during analysis.
func TestAnalyzeTimeShift_ArithmeticExpressionInShiftAmount(t *testing.T) {
	t.Parallel()

	input := `concepts
	Dim = [Values[D1]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]
measures
	daily_metric = Number(unit=Unit)
	Daily = Interval(unit=Day)
metrics
	shifted
		inputs
			metric daily_metric backward(day, 1 + 1)
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed for arithmetic expression in shift amount")
	require.NotNil(t, defs)

	require.Contains(t, defs.Metrics, "shifted")
	metric := defs.Metrics["shifted"]
	require.Len(t, metric.Inputs, 1)

	require.NotNil(t, metric.Inputs[0].TimeShift)
	assert.Equal(t, "backward", metric.Inputs[0].TimeShift.Direction)
	assert.Equal(t, "day", metric.Inputs[0].TimeShift.Field)
	assert.Equal(t, 2.0, metric.Inputs[0].TimeShift.Amount,
		"Arithmetic expression 1 + 1 should evaluate to 2.0")
}

// TestParseTimeShift_ForwardMonth verifies parsing forward(month) produces
// ParserTimeShift with Direction="forward", Field="month", Amount=1.0.
func TestParseTimeShift_ForwardMonth(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x forward(month)
		calculate
			0`

	doc, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	metricDef := requireParsedMetricDefinition(t, doc)
	require.Len(t, metricDef.Inputs, 1)

	require.NotNil(t, metricDef.Inputs[0].TimeShift)
	assert.Equal(t, "forward", metricDef.Inputs[0].TimeShift.Direction)
	assert.Equal(t, "month", metricDef.Inputs[0].TimeShift.Field)
	assert.Equal(t, 1.0, metricDef.Inputs[0].TimeShift.Amount)
}

// TestParseTimeShift_FromExampleFile verifies parsing the time_shift_basic.cpl example
// produces the expected ParserTimeShift values.
func TestParseTimeShift_FromExampleFile(t *testing.T) {
	t.Parallel()

	examplesDir := getTimeShiftExamplesDir(t)
	filePath := filepath.Join(examplesDir, "time_shift_basic.cpl")

	config := &parser.ParseLoggingConfig{}
	doc, err := parser.ParseFiles([]string{filePath}, config)
	require.NoError(t, err, "Parsing example file should succeed")

	metricDef := requireParsedMetricDefinition(t, doc)
	require.Len(t, metricDef.Inputs, 1)

	assert.Equal(t, "metric_b", metricDef.Inputs[0].Name)
	require.NotNil(t, metricDef.Inputs[0].TimeShift)
	assert.Equal(t, "backward", metricDef.Inputs[0].TimeShift.Direction)
	assert.Equal(t, "year", metricDef.Inputs[0].TimeShift.Field)
	assert.Equal(t, 1.0, metricDef.Inputs[0].TimeShift.Amount)
}

// ============================================================================
// 6c) Verify analysis stage correctly validates field references
// ============================================================================

// TestAnalyzeTimeShift_AllValidFields_Pass verifies that each valid Instant
// constraint field keyword passes analysis when used in a time-shift clause.
func TestAnalyzeTimeShift_AllValidFields_Pass(t *testing.T) {
	t.Parallel()

	fields := []string{
		"year", "month", "day", "hour", "minute", "second", "day_of_week", "week_number",
	}

	for _, field := range fields {
		field := field // capture range variable
		t.Run(field, func(t *testing.T) {
			t.Parallel()

			input := `concepts
			Dim = [Values[D1]]
			Unit = [Values[Each]]
			Day = [Values[Monday]]
		measures
			m = Number(unit=Unit)
			Daily = Interval(unit=Day)
		metrics
			my_metric
				inputs
					metric m backward(` + field + `)
				calculate
					0`

			rootTable, err := parser.Parse(input)
			require.NoError(t, err, "Parsing should succeed for field %q", field)

			defs, err := analysis.Analyze(rootTable)
			require.NoError(t, err, "Analysis should succeed for valid field %q", field)
			require.NotNil(t, defs)

			require.Contains(t, defs.Metrics, "my_metric")
			metric := defs.Metrics["my_metric"]
			require.Len(t, metric.Inputs, 1)
			require.NotNil(t, metric.Inputs[0].TimeShift)
			assert.Equal(t, field, metric.Inputs[0].TimeShift.Field)
		})
	}
}

// TestAnalyzeTimeShift_InvalidField_Foo_Error verifies that an invalid field
// keyword "foo" produces an analysis error.
func TestAnalyzeTimeShift_InvalidField_Foo_Error(t *testing.T) {
	t.Parallel()

	input := `concepts
	Dim = [Values[D1]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]
measures
	m = Number(unit=Unit)
	Daily = Interval(unit=Day)
metrics
	my_metric
		inputs
			metric m backward(foo)
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	_, err = analysis.Analyze(rootTable)
	require.Error(t, err, "Analysis should fail for invalid field 'foo'")
	assert.Contains(t, err.Error(), "invalid time-shift field")
	assert.Contains(t, err.Error(), "foo")
}

// TestAnalyzeTimeShift_InvalidField_Empty_Error verifies that an empty field
// in a backward clause produces a parse error (the grammar requires an identifier).
func TestAnalyzeTimeShift_InvalidField_Empty_Error(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric m backward()
		calculate
			0`

	_, err := parser.Parse(input)
	require.Error(t, err, "Parsing should fail for empty field in backward()")
}

// ============================================================================
// 6d) Verify self-reference semantics
// ============================================================================

// TestAnalyzeTimeShift_SelfReference_Backward_Allowed verifies that a metric
// referencing itself with backward(year) passes analysis as a time-series dependency.
func TestAnalyzeTimeShift_SelfReference_Backward_Allowed(t *testing.T) {
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
			metric revenue backward(year)
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Self-reference with backward time shift should be allowed")
	require.NotNil(t, defs)

	require.Contains(t, defs.Metrics, "revenue")
	metric := defs.Metrics["revenue"]
	require.Len(t, metric.Inputs, 1)
	require.NotNil(t, metric.Inputs[0].TimeShift)
	assert.Equal(t, "backward", metric.Inputs[0].TimeShift.Direction)
	assert.Equal(t, "year", metric.Inputs[0].TimeShift.Field)
}

// TestAnalyzeTimeShift_SelfReference_Forward_Allowed verifies that a metric
// referencing itself with forward(month) passes analysis.
func TestAnalyzeTimeShift_SelfReference_Forward_Allowed(t *testing.T) {
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
			metric revenue forward(month)
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Self-reference with forward time shift should be allowed")
	require.NotNil(t, defs)

	require.Contains(t, defs.Metrics, "revenue")
	metric := defs.Metrics["revenue"]
	require.Len(t, metric.Inputs, 1)
	require.NotNil(t, metric.Inputs[0].TimeShift)
	assert.Equal(t, "forward", metric.Inputs[0].TimeShift.Direction)
	assert.Equal(t, "month", metric.Inputs[0].TimeShift.Field)
}

// TestAnalyzeTimeShift_SelfReference_NoShift_Error verifies that a metric
// referencing itself without a time shift returns an error containing "references itself".
func TestAnalyzeTimeShift_SelfReference_NoShift_Error(t *testing.T) {
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
			metric revenue
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	_, err = analysis.Analyze(rootTable)
	require.Error(t, err, "Self-reference without time shift should return error")
	assert.Contains(t, err.Error(), "references itself")
	assert.Contains(t, err.Error(), "revenue")
}

// TestAnalyzeTimeShift_SelfReference_InvalidField_Error verifies that a
// self-reference with an invalid time-shift field still returns a field validation error.
func TestAnalyzeTimeShift_SelfReference_InvalidField_Error(t *testing.T) {
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
			metric revenue backward(invalid_field)
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	_, err = analysis.Analyze(rootTable)
	require.Error(t, err, "Self-reference with invalid field should return error")
	assert.Contains(t, err.Error(), "invalid time-shift field")
	assert.Contains(t, err.Error(), "invalid_field")
}

// ============================================================================
// 6e) Test edge cases
// ============================================================================

// TestAnalyzeTimeShift_MultipleShiftedInputs verifies a metric with multiple
// time-shifted inputs is handled correctly.
func TestAnalyzeTimeShift_MultipleShiftedInputs(t *testing.T) {
	t.Parallel()

	input := `concepts
	Dim = [Values[D1]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]
measures
	input_a = Number(unit=Unit)
	input_b = Number(unit=Unit)
	Daily = Interval(unit=Day)
metrics
	my_metric
		inputs
			metric input_a backward(year)
			metric input_b forward(month, 3)
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed for multiple shifted inputs")
	require.NotNil(t, defs)

	metric := defs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 2)

	require.NotNil(t, metric.Inputs[0].TimeShift)
	assert.Equal(t, "backward", metric.Inputs[0].TimeShift.Direction)
	assert.Equal(t, "year", metric.Inputs[0].TimeShift.Field)

	require.NotNil(t, metric.Inputs[1].TimeShift)
	assert.Equal(t, "forward", metric.Inputs[1].TimeShift.Direction)
	assert.Equal(t, "month", metric.Inputs[1].TimeShift.Field)
	assert.Equal(t, 3.0, metric.Inputs[1].TimeShift.Amount)
}

// TestAnalyzeTimeShift_MixedShiftedAndNonShifted verifies a mix of time-shifted
// and non-time-shifted inputs on a single metric.
func TestAnalyzeTimeShift_MixedShiftedAndNonShifted(t *testing.T) {
	t.Parallel()

	input := `concepts
	Dim = [Values[D1]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]
measures
	input_a = Number(unit=Unit)
	input_b = Number(unit=Unit)
	input_c = Number(unit=Unit)
	Daily = Interval(unit=Day)
metrics
	my_metric
		inputs
			metric input_a backward(year)
			metric input_b
			metric input_c forward(day)
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed for mixed inputs")
	require.NotNil(t, defs)

	metric := defs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 3)

	require.NotNil(t, metric.Inputs[0].TimeShift)
	assert.Equal(t, "backward", metric.Inputs[0].TimeShift.Direction)
	assert.Equal(t, "year", metric.Inputs[0].TimeShift.Field)

	assert.Nil(t, metric.Inputs[1].TimeShift, "Second input should have no time shift")

	require.NotNil(t, metric.Inputs[2].TimeShift)
	assert.Equal(t, "forward", metric.Inputs[2].TimeShift.Direction)
	assert.Equal(t, "day", metric.Inputs[2].TimeShift.Field)
}

// TestAnalyzeTimeShift_BackwardAndForwardDifferentInputs verifies both backward
// and forward on different inputs of the same metric.
func TestAnalyzeTimeShift_BackwardAndForwardDifferentInputs(t *testing.T) {
	t.Parallel()

	input := `concepts
	Dim = [Values[D1]]
	Unit = [Values[Each]]
	Day = [Values[Monday]]
measures
	past = Number(unit=Unit)
	future = Number(unit=Unit)
	Daily = Interval(unit=Day)
metrics
	my_metric
		inputs
			metric past backward(month, 6)
			metric future forward(week_number, 2)
		calculate
			0`

	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	defs, err := analysis.Analyze(rootTable)
	require.NoError(t, err, "Analysis should succeed")
	require.NotNil(t, defs)

	metric := defs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 2)

	assert.Equal(t, "backward", metric.Inputs[0].TimeShift.Direction)
	assert.Equal(t, "month", metric.Inputs[0].TimeShift.Field)
	assert.Equal(t, 6.0, metric.Inputs[0].TimeShift.Amount)

	assert.Equal(t, "forward", metric.Inputs[1].TimeShift.Direction)
	assert.Equal(t, "week_number", metric.Inputs[1].TimeShift.Field)
	assert.Equal(t, 2.0, metric.Inputs[1].TimeShift.Amount)
}

// TestAnalyzeTimeShift_NegativeAmount verifies that a negative amount in a
// time shift is preserved correctly during analysis. The grammar does not
// currently support negative literals in shift expressions, so this test
// constructs the symbol table programmatically to verify analysis behavior.
func TestAnalyzeTimeShift_NegativeAmount(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name: "my_metric",
		Inputs: []parsing.ParserInputDefinition{
			{
				Kind: "metric",
				Name: "input_a",
				TimeShift: &parsing.ParserTimeShift{
					Direction: "backward",
					Field:     "day",
					Amount:    -3.0,
				},
				NameLocation: symboltable.SourceLocation{Line: 3, Column: 3},
			},
		},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 0}},
		SourceLocation:   symboltable.SourceLocation{Line: 2, Column: 1},
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

	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)
	require.NoError(t, err, "Analysis should succeed for negative amount")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "my_metric")
	metric := resultDefs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 1)
	require.NotNil(t, metric.Inputs[0].TimeShift)
	assert.Equal(t, -3.0, metric.Inputs[0].TimeShift.Amount)
}

// ============================================================================
// 6f) Verify self-reference without time shift is rejected
// ============================================================================

// TestAnalyzeTimeShift_SelfRefErrorFile verifies that the time_shift_self_ref_error.cpl
// example file produces the expected self-reference error during analysis.
func TestAnalyzeTimeShift_SelfRefErrorFile(t *testing.T) {
	t.Parallel()

	examplesDir := getTimeShiftExamplesDir(t)
	filePath := filepath.Join(examplesDir, "time_shift_self_ref_error.cpl")

	config := &parser.ParseLoggingConfig{}
	rootTable, err := parser.ParseFiles([]string{filePath}, config)
	require.NoError(t, err, "Parsing example file should succeed")

	_, err = analysis.Analyze(rootTable)
	require.Error(t, err, "Analysis should fail for self-reference without time shift")
	assert.Contains(t, err.Error(), "references itself")
	assert.Contains(t, err.Error(), "revenue")
}

// ============================================================================
// Example file integration tests
// ============================================================================

// TestAnalyzeTimeShift_ExampleFiles verify that all time-shift example files
// parse and analyze correctly (except the intentionally error file).
func TestAnalyzeTimeShift_ExampleFiles(t *testing.T) {
	t.Parallel()

	examplesDir := getTimeShiftExamplesDir(t)

	// Files expected to parse and analyze successfully
	validExamples := []struct {
		fileName      string
		metricName    string
		expectedInput string
		expectedField string
		expectedDir   string
		expectedAmt   float64
	}{
		{
			fileName:      "time_shift_basic.cpl",
			metricName:    "metric_a",
			expectedInput: "metric_b",
			expectedField: "year",
			expectedDir:   "backward",
			expectedAmt:   1.0,
		},
		{
			fileName:      "time_shift_self_ref.cpl",
			metricName:    "revenue",
			expectedInput: "revenue",
			expectedField: "year",
			expectedDir:   "backward",
			expectedAmt:   1.0,
		},
		{
			fileName:      "time_shift_forward.cpl",
			metricName:    "my_forecast",
			expectedInput: "base",
			expectedField: "month",
			expectedDir:   "forward",
			expectedAmt:   1.0,
		},
		{
			fileName:      "time_shift_with_where.cpl",
			metricName:    "adjusted_sales",
			expectedInput: "sales",
			expectedField: "week_number",
			expectedDir:   "backward",
			expectedAmt:   1.0,
		},
		{
			fileName:      "time_shift_amount.cpl",
			metricName:    "shifted_metric",
			expectedInput: "daily_metric",
			expectedField: "day",
			expectedDir:   "backward",
			expectedAmt:   2.0,
		},
	}

	for _, ex := range validExamples {
		ex := ex // capture range variable
		t.Run(ex.fileName, func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(examplesDir, ex.fileName)
			config := &parser.ParseLoggingConfig{}
			rootTable, err := parser.ParseFiles([]string{filePath}, config)
			require.NoError(t, err, "Parsing %s should succeed", ex.fileName)

			defs, err := analysis.Analyze(rootTable)
			require.NoError(t, err, "Analysis of %s should succeed", ex.fileName)
			require.NotNil(t, defs)

			// Find the single metric in the file
			require.Len(t, defs.Metrics, 1, "Example should contain exactly one metric")
			metric, ok := defs.Metrics[ex.metricName]
			require.True(t, ok, "Expected metric %q to be present", ex.metricName)

			require.Len(t, metric.Inputs, 1, "Example metric should have exactly one input")
			assert.Equal(t, ex.expectedInput, metric.Inputs[0].Name)
			require.NotNil(t, metric.Inputs[0].TimeShift, "TimeShift should not be nil")
			assert.Equal(t, ex.expectedDir, metric.Inputs[0].TimeShift.Direction)
			assert.Equal(t, ex.expectedField, metric.Inputs[0].TimeShift.Field)
			assert.Equal(t, ex.expectedAmt, metric.Inputs[0].TimeShift.Amount)
		})
	}
}
