// Package parser_test contains tests for the CPL parser metrics section.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parsing"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Helper functions for PostfixElement assertions
// ============================================================================

// assertPostfixLiteral asserts that a PostfixElement is a PostfixLiteral with the given value.
func assertPostfixLiteral(t *testing.T, elem parsing.PostfixElement, expectedValue float64, msgAndArgs ...interface{}) {
	t.Helper()
	lit, ok := elem.(parsing.PostfixLiteral)
	if !ok {
		t.Fatalf("expected PostfixLiteral, got %T", elem)
	}
	assert.Equal(t, expectedValue, lit.Value, msgAndArgs...)
}

// assertPostfixIdentifier asserts that a PostfixElement is a PostfixIdentifier with the given name.
func assertPostfixIdentifier(t *testing.T, elem parsing.PostfixElement, expectedName string, msgAndArgs ...interface{}) {
	t.Helper()
	id, ok := elem.(parsing.PostfixIdentifier)
	if !ok {
		t.Fatalf("expected PostfixIdentifier, got %T", elem)
	}
	assert.Equal(t, expectedName, id.Name, msgAndArgs...)
}

// assertPostfixOperator asserts that a PostfixElement is a PostfixOperator with the given DefType.
func assertPostfixOperator(t *testing.T, elem parsing.PostfixElement, expectedDefType symboltable.DefinitionType, msgAndArgs ...interface{}) {
	t.Helper()
	op, ok := elem.(parsing.PostfixOperator)
	if !ok {
		t.Fatalf("expected PostfixOperator, got %T", elem)
	}
	assert.Equal(t, expectedDefType, op.DefType, msgAndArgs...)
}

// ============================================================================
// Tests
// ============================================================================

// TestParseSingleMetricDefinition tests parsing a single metric with inputs and calculate.
func TestParseSingleMetricDefinition(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric demand_forecast
			dimension Price
		calculate
			0`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify the MetricsScope contains one MetricScope with structured data.
	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	assert.Equal(t, symboltable.MetricsScope, metricsScope.ScopeType, "Expected MetricsScope")
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")
	assert.Equal(t, symboltable.MetricScope, metricsScope.Children[0].ScopeType, "Expected MetricScope")

	// Verify the metric definition was captured
	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	assert.Equal(t, "my_metric", metricDef.Name, "MetricDefinition Name should match metric identifier")
	assert.False(t, metricDef.ShortForm, "MetricDefinition ShortForm should be false for full form syntax")

	// Verify inputs were captured
	require.Len(t, metricDef.Inputs, 2, "MetricDefinition should have 2 inputs")
	assert.Equal(t, "metric", metricDef.Inputs[0].Kind, "First input kind should be 'metric'")
	assert.Equal(t, "demand_forecast", metricDef.Inputs[0].Name, "First input name should be 'demand_forecast'")
	assert.Equal(t, "dimension", metricDef.Inputs[1].Kind, "Second input kind should be 'dimension'")
	assert.Equal(t, "Price", metricDef.Inputs[1].Name, "Second input name should be 'Price'")

	// Verify calculate postfix contains the literal 0
	require.Len(t, metricDef.CalculatePostfix, 1, "CalculatePostfix should have 1 element")
	assertPostfixLiteral(t, metricDef.CalculatePostfix[0], 0, "CalculatePostfix should contain literal 0")
}

// TestParseMultipleMetricDefinitions tests parsing two metrics in one section.
func TestParseMultipleMetricDefinitions(t *testing.T) {
	t.Parallel()

	input := `metrics
	first_metric
		inputs
			metric input1
		calculate
			0
	second_metric
		inputs
			dimension dim1
		calculate
			1`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify the MetricsScope contains two MetricScopes.
	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	assert.Equal(t, symboltable.MetricsScope, metricsScope.ScopeType, "Expected MetricsScope")
	require.Len(t, metricsScope.Children, 2, "Expected two MetricScope children")

	// Verify first metric
	firstDef := requireMetricDefinition(t, metricsScope.Children[0])
	assert.Equal(t, "first_metric", firstDef.Name, "First metric name should be 'first_metric'")
	require.Len(t, firstDef.Inputs, 1, "First metric should have 1 input")
	assert.Equal(t, "metric", firstDef.Inputs[0].Kind, "First metric input kind should be 'metric'")
	assert.Equal(t, "input1", firstDef.Inputs[0].Name, "First metric input name should be 'input1'")
	require.Len(t, firstDef.CalculatePostfix, 1, "First metric CalculatePostfix should have 1 element")
	assertPostfixLiteral(t, firstDef.CalculatePostfix[0], 0, "First metric CalculatePostfix should contain 0")

	// Verify second metric
	secondDef := requireMetricDefinition(t, metricsScope.Children[1])
	assert.Equal(t, "second_metric", secondDef.Name, "Second metric name should be 'second_metric'")
	require.Len(t, secondDef.Inputs, 1, "Second metric should have 1 input")
	assert.Equal(t, "dimension", secondDef.Inputs[0].Kind, "Second metric input kind should be 'dimension'")
	assert.Equal(t, "dim1", secondDef.Inputs[0].Name, "Second metric input name should be 'dim1'")
	require.Len(t, secondDef.CalculatePostfix, 1, "Second metric CalculatePostfix should have 1 element")
	assertPostfixLiteral(t, secondDef.CalculatePostfix[0], 1, "Second metric CalculatePostfix should contain 1")
}

// TestParseMetricInputKinds verifies metric and dimension input kinds are correctly captured.
func TestParseMetricInputKinds(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         string
		expectedKinds []string
		expectedNames []string
	}{
		{
			name: "only metric inputs",
			input: `metrics
	my_metric
		inputs
			metric metric_a
			metric metric_b
		calculate
			0`,
			expectedKinds: []string{"metric", "metric"},
			expectedNames: []string{"metric_a", "metric_b"},
		},
		{
			name: "only dimension inputs",
			input: `metrics
	my_metric
		inputs
			dimension dim_a
			dimension dim_b
		calculate
			0`,
			expectedKinds: []string{"dimension", "dimension"},
			expectedNames: []string{"dim_a", "dim_b"},
		},
		{
			name: "mixed inputs",
			input: `metrics
	my_metric
		inputs
			metric some_metric
			dimension some_dimension
			metric another_metric
		calculate
			0`,
			expectedKinds: []string{"metric", "dimension", "metric"},
			expectedNames: []string{"some_metric", "some_dimension", "another_metric"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := parser.Parse(tc.input)

			require.NoError(t, err, "Expected no error for input: %s", tc.input)
			require.NotNil(t, doc, "Document should not be nil")
			require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
			metricsScope := doc.Children[0]
			require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

			metricDef := requireMetricDefinition(t, metricsScope.Children[0])
			require.Len(t, metricDef.Inputs, len(tc.expectedKinds), "Input count should match expected")

			for i, expectedKind := range tc.expectedKinds {
				assert.Equal(t, expectedKind, metricDef.Inputs[i].Kind, "Input %d kind should match", i)
				assert.Equal(t, tc.expectedNames[i], metricDef.Inputs[i].Name, "Input %d name should match", i)
			}
		})
	}
}

// TestParseMetricCalculateLiteral verifies the literal is captured in the postfix.
func TestParseMetricCalculateLiteral(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         string
		expectedValue float64
	}{
		{
			name: "literal zero",
			input: `metrics
	my_metric
		inputs
			metric input1
		calculate
			0`,
			expectedValue: 0,
		},
		{
			name: "literal one",
			input: `metrics
	my_metric
		inputs
			metric input1
		calculate
			1`,
			expectedValue: 1,
		},
		{
			name: "literal large number",
			input: `metrics
	my_metric
		inputs
			metric input1
		calculate
			12345`,
			expectedValue: 12345,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := parser.Parse(tc.input)

			require.NoError(t, err, "Expected no error for input: %s", tc.input)
			require.NotNil(t, doc, "Document should not be nil")
			require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
			metricsScope := doc.Children[0]
			require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

			metricDef := requireMetricDefinition(t, metricsScope.Children[0])
			require.Len(t, metricDef.CalculatePostfix, 1, "Postfix length should be 1")

			assertPostfixLiteral(t, metricDef.CalculatePostfix[0], tc.expectedValue, "Postfix element should match")
		})
	}
}

// TestParseMetricsWithOtherSections tests parsing metrics alongside other sections.
func TestParseMetricsWithOtherSections(t *testing.T) {
	t.Parallel()

	input := `types
	NumberType = Number()

units
	Unit1 = NumberType

measures
	Measure1 = NumberType(unit=Unit1)

connectors
	Conn1 = dimensions [Dim1] grain Day celltype Measure1

metrics
	my_metric
		inputs
			metric other_metric
		calculate
			0`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have TypesScope, UnitsScope, MeasuresScope, ConnectorsScope, and MetricsScope.
	require.Len(t, doc.Children, 5, "Expected five child scopes")
	metricsScope := doc.Children[4]
	assert.Equal(t, symboltable.MetricsScope, metricsScope.ScopeType, "Expected fifth child to be MetricsScope")
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope child")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	assert.Equal(t, "my_metric", metricDef.Name, "Metric name should be 'my_metric'")
}

// TestParseMetricsScopeStructure verifies the symbol table structure for metrics.
func TestParseMetricsScopeStructure(t *testing.T) {
	t.Parallel()

	input := `metrics
	sales_forecast
		inputs
			metric historical_sales
			dimension Product
			dimension Region
		calculate
			0`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Root scope should be GlobalScope
	assert.Equal(t, symboltable.GlobalScope, doc.ScopeType, "Root scope should be GlobalScope")

	// Should have one child: MetricsScope
	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	assert.Equal(t, symboltable.MetricsScope, metricsScope.ScopeType, "First child should be MetricsScope")
	assert.Equal(t, doc, metricsScope.Parent, "MetricsScope parent should be root scope")

	// MetricsScope should have one child: MetricScope
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope child")
	metricScope := metricsScope.Children[0]
	assert.Equal(t, symboltable.MetricScope, metricScope.ScopeType, "Child should be MetricScope")
	assert.Equal(t, metricsScope, metricScope.Parent, "MetricScope parent should be MetricsScope")

	// Verify the metric definition data
	metricDef := requireMetricDefinition(t, metricScope)
	assert.Equal(t, "sales_forecast", metricDef.Name)
}

// TestParseMetricsSourceLocation tests that source locations are captured for metric definitions.
func TestParseMetricsSourceLocation(t *testing.T) {
	t.Parallel()

	input := `metrics
	metric_one
		inputs
			metric input1
		calculate
			0
	metric_two
		inputs
			dimension dim1
		calculate
			1`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error")
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 2, "Expected two MetricScope children")

	// Verify first metric source location
	firstDef := requireMetricDefinition(t, metricsScope.Children[0])
	assert.Equal(t, "metric_one", firstDef.Name)
	assert.Positive(t, firstDef.SourceLocation.Line, "First metric should capture a source line")

	// Verify second metric source location
	secondDef := requireMetricDefinition(t, metricsScope.Children[1])
	assert.Equal(t, "metric_two", secondDef.Name)
	assert.Positive(t, secondDef.SourceLocation.Line, "Second metric should capture a source line")
}

// ============================================================================
// Short Form Declaration Tests
// ============================================================================

// TestParseShortFormMetric_LiteralExpression tests parsing a short form metric
// with a simple literal expression: revenue = 42
func TestParseShortFormMetric_LiteralExpression(t *testing.T) {
	t.Parallel()

	input := `metrics
	revenue = 42`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify MetricsScope contains one MetricScope
	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	assert.Equal(t, symboltable.MetricsScope, metricsScope.ScopeType)
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")
	assert.Equal(t, symboltable.MetricScope, metricsScope.Children[0].ScopeType)

	// Verify the metric definition
	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	assert.Equal(t, "revenue", metricDef.Name)
	assert.True(t, metricDef.ShortForm, "ShortForm should be true for short form syntax")
	assert.Nil(t, metricDef.Inputs, "Inputs should be nil for short form (derived during analysis)")
	require.Len(t, metricDef.CalculatePostfix, 1, "CalculatePostfix should have 1 element (literal 42)")
	assertPostfixLiteral(t, metricDef.CalculatePostfix[0], 42)
}

// TestParseShortFormMetric_IdentifierExpression tests parsing a short form metric
// with an identifier expression: revenue = demand_forecast
func TestParseShortFormMetric_IdentifierExpression(t *testing.T) {
	t.Parallel()

	input := `metrics
	revenue = demand_forecast`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify MetricsScope contains one MetricScope
	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	assert.Equal(t, "revenue", metricDef.Name)
	assert.True(t, metricDef.ShortForm, "ShortForm should be true")
	assert.Nil(t, metricDef.Inputs)

	require.Len(t, metricDef.CalculatePostfix, 1, "CalculatePostfix should have 1 element (identifier)")
	assertPostfixIdentifier(t, metricDef.CalculatePostfix[0], "demand_forecast")
}

// TestParseShortFormMetric_ArithmeticExpression tests parsing a short form metric
// with an arithmetic expression: revenue = demand_forecast * Price + 5
func TestParseShortFormMetric_ArithmeticExpression(t *testing.T) {
	t.Parallel()

	input := `metrics
	revenue = demand_forecast * Price + 5`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1)
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1)

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	assert.Equal(t, "revenue", metricDef.Name)
	assert.True(t, metricDef.ShortForm)
	assert.Nil(t, metricDef.Inputs)

	// Postfix should be ordered: demand_forecast Price * 5 +
	// (demand_forecast * Price + 5 in infix)
	require.Len(t, metricDef.CalculatePostfix, 5, "Expected 5 postfix elements (2 identifiers, *, 5, +)")
	assertPostfixIdentifier(t, metricDef.CalculatePostfix[0], "demand_forecast")
	assertPostfixIdentifier(t, metricDef.CalculatePostfix[1], "Price")
	assertPostfixOperator(t, metricDef.CalculatePostfix[2], symboltable.MultiplyOperatorDefinition)
	assertPostfixLiteral(t, metricDef.CalculatePostfix[3], 5)
	assertPostfixOperator(t, metricDef.CalculatePostfix[4], symboltable.PlusOperatorDefinition)
}

// TestParseShortFormMetric_ParenthesizedExpression tests parsing a short form
// metric with parenthesized expression: profit = (revenue - cost) * 2
func TestParseShortFormMetric_ParenthesizedExpression(t *testing.T) {
	t.Parallel()

	input := `metrics
	profit = (revenue - cost) * 2`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1)
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1)

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	assert.Equal(t, "profit", metricDef.Name)
	assert.True(t, metricDef.ShortForm)
	assert.Nil(t, metricDef.Inputs)

	// Postfix: revenue cost - 2 *
	// (revenue - cost) * 2 in postfix = revenue cost - 2 *
	require.Len(t, metricDef.CalculatePostfix, 5, "Expected 5 postfix elements")
	assertPostfixIdentifier(t, metricDef.CalculatePostfix[0], "revenue")
	assertPostfixIdentifier(t, metricDef.CalculatePostfix[1], "cost")
	assertPostfixOperator(t, metricDef.CalculatePostfix[2], symboltable.MinusOperatorDefinition)
	assertPostfixLiteral(t, metricDef.CalculatePostfix[3], 2)
	assertPostfixOperator(t, metricDef.CalculatePostfix[4], symboltable.MultiplyOperatorDefinition)
}

// TestParseShortFormMetric_UnaryNegation tests parsing a short form metric
// with unary negation: loss = -profit
func TestParseShortFormMetric_UnaryNegation(t *testing.T) {
	t.Parallel()

	input := `metrics
	loss = -profit`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)

	require.Len(t, doc.Children, 1)
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1)

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	assert.Equal(t, "loss", metricDef.Name)
	assert.True(t, metricDef.ShortForm)
	assert.Nil(t, metricDef.Inputs)

	// Postfix: profit, negate
	require.Len(t, metricDef.CalculatePostfix, 2, "Expected 2 postfix elements (profit, negate)")
	assertPostfixIdentifier(t, metricDef.CalculatePostfix[0], "profit")
	assertPostfixOperator(t, metricDef.CalculatePostfix[1], symboltable.NegateOperatorDefinition)
}

// TestParseShortFormMetric_MixedWithLongForm tests parsing both short form
// and long form metrics in the same metrics section.
func TestParseShortFormMetric_MixedWithLongForm(t *testing.T) {
	t.Parallel()

	input := `metrics
	revenue = demand_forecast * Price
	profit
		inputs
			metric revenue
			dimension Cost
		calculate
			revenue - cost_value`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1)
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 2, "Expected two MetricScope children")

	// First metric (short form)
	firstDef := requireMetricDefinition(t, metricsScope.Children[0])
	assert.Equal(t, "revenue", firstDef.Name)
	assert.True(t, firstDef.ShortForm, "First metric should be short form")
	assert.Nil(t, firstDef.Inputs)
	assert.NotNil(t, firstDef.CalculatePostfix)

	// Second metric (long form)
	secondDef := requireMetricDefinition(t, metricsScope.Children[1])
	assert.Equal(t, "profit", secondDef.Name)
	assert.False(t, secondDef.ShortForm, "Second metric should be long form")
	assert.NotNil(t, secondDef.Inputs)
	assert.Len(t, secondDef.Inputs, 2)
	assert.NotNil(t, secondDef.CalculatePostfix)
}

// TestParseShortFormMetric_MultipleShortFormMetrics tests parsing multiple
// short form metrics in the same metrics section.
func TestParseShortFormMetric_MultipleShortFormMetrics(t *testing.T) {
	t.Parallel()

	input := `metrics
	revenue = demand_forecast * Price
	profit = revenue - cost
	loss = -profit`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1)
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 3, "Expected three MetricScope children")

	for i, name := range []string{"revenue", "profit", "loss"} {
		def := requireMetricDefinition(t, metricsScope.Children[i])
		assert.Equal(t, name, def.Name, "Metric %d name should be %s", i, name)
		assert.True(t, def.ShortForm, "Metric %d should be short form", i)
	}
}

// ============================================================================
// Input Alias and Where Clause Tests
// ============================================================================

// TestParseInputWithAlias tests parsing an input declaration with an alias clause.
func TestParseInputWithAlias(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric demand_forecast as df
		calculate
			0`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	require.Len(t, metricDef.Inputs, 1, "MetricDefinition should have 1 input")

	assert.Equal(t, "metric", metricDef.Inputs[0].Kind, "Input kind should be 'metric'")
	assert.Equal(t, "demand_forecast", metricDef.Inputs[0].Name, "Input name should be 'demand_forecast'")
	assert.Equal(t, "df", metricDef.Inputs[0].Alias, "Input alias should be 'df'")
	assert.Nil(t, metricDef.Inputs[0].WhereClause, "WhereClause should be nil without where clause")
}

// TestParseInputWithWhereClause tests parsing an input declaration with a where clause.
func TestParseInputWithWhereClause(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric demand_forecast where aspect=actuals
		calculate
			0`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	require.Len(t, metricDef.Inputs, 1, "MetricDefinition should have 1 input")

	assert.Equal(t, "metric", metricDef.Inputs[0].Kind, "Input kind should be 'metric'")
	assert.Equal(t, "demand_forecast", metricDef.Inputs[0].Name, "Input name should be 'demand_forecast'")
	assert.Empty(t, metricDef.Inputs[0].Alias, "Alias should be empty without as clause")

	require.NotNil(t, metricDef.Inputs[0].WhereClause, "Should have a where clause")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups, 1, "Should have 1 AndGroup")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions, 1, "Should have 1 condition in first AndGroup")
	assert.Equal(t, "ACTUALS", metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect, "Aspect should be ACTUALS")
	assert.False(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Negated, "Condition should not be negated")
}

// TestParseInputWithAliasAndWhere tests parsing an input declaration with both alias and where clause.
func TestParseInputWithAliasAndWhere(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric demand_forecast as df where aspect=forecast
		calculate
			0`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	require.Len(t, metricDef.Inputs, 1, "MetricDefinition should have 1 input")

	assert.Equal(t, "metric", metricDef.Inputs[0].Kind, "Input kind should be 'metric'")
	assert.Equal(t, "demand_forecast", metricDef.Inputs[0].Name, "Input name should be 'demand_forecast'")
	assert.Equal(t, "df", metricDef.Inputs[0].Alias, "Input alias should be 'df'")

	require.NotNil(t, metricDef.Inputs[0].WhereClause, "Should have a where clause")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups, 1, "Should have 1 AndGroup")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions, 1, "Should have 1 condition in first AndGroup")
	assert.Equal(t, "FORECAST", metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect, "Aspect should be FORECAST")
	assert.False(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Negated, "Condition should not be negated")
}

// TestParseInputWithComplexWhere tests parsing an input declaration with multiple where conditions joined by logical operators.
func TestParseInputWithComplexWhere(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x where aspect=actuals and aspect=forecast
		calculate
			0`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	require.Len(t, metricDef.Inputs, 1, "MetricDefinition should have 1 input")
	assert.Equal(t, "x", metricDef.Inputs[0].Name, "Input name should be 'x'")

	require.NotNil(t, metricDef.Inputs[0].WhereClause, "Should have a where clause")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups, 1, "Should have 1 AndGroup")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions, 2, "Should have 2 conditions in first AndGroup")
	assert.Equal(t, "ACTUALS", metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect, "First aspect should be ACTUALS")
	assert.False(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Negated, "First condition should not be negated")
	assert.Equal(t, "FORECAST", metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[1].Aspect, "Second aspect should be FORECAST")
	assert.False(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[1].Negated, "Second condition should not be negated")
}

// TestParseInputWithNotAspect tests parsing a where clause with the NOT keyword on an aspect expression.
func TestParseInputWithNotAspect(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x where not aspect=forecast
		calculate
			0`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	require.Len(t, metricDef.Inputs, 1, "MetricDefinition should have 1 input")

	require.NotNil(t, metricDef.Inputs[0].WhereClause, "Should have a where clause")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups, 1, "Should have 1 AndGroup")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions, 1, "Should have 1 condition in first AndGroup")
	assert.Equal(t, "FORECAST", metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect, "Aspect should be FORECAST")
	assert.True(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Negated, "Condition should be negated")
}

// TestParseInputWithoutAliasOrWhere tests that inputs without alias or where have empty fields.
func TestParseInputWithoutAliasOrWhere(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric demand_forecast
		calculate
			0`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	require.Len(t, metricDef.Inputs, 1, "MetricDefinition should have 1 input")

	assert.Empty(t, metricDef.Inputs[0].Alias, "Alias should be empty without as clause")
	assert.Nil(t, metricDef.Inputs[0].WhereClause, "WhereClause should be nil without where clause")
}

// TestParseInputWithOr tests parsing a where clause with OR operator.
func TestParseInputWithOr(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x where aspect=actuals or aspect=forecast
		calculate
			0`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	require.Len(t, metricDef.Inputs, 1, "MetricDefinition should have 1 input")

	require.NotNil(t, metricDef.Inputs[0].WhereClause, "Should have a where clause")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups, 2, "Should have 2 AndGroups")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions, 1, "First AndGroup should have 1 condition")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[1].Conditions, 1, "Second AndGroup should have 1 condition")
	assert.Equal(t, "ACTUALS", metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect, "First AndGroup aspect should be ACTUALS")
	assert.Equal(t, "FORECAST", metricDef.Inputs[0].WhereClause.AndGroups[1].Conditions[0].Aspect, "Second AndGroup aspect should be FORECAST")
}

// TestParseInput_AndBindsTighterThanOr_AndFirst tests that AND binds tighter than OR.
// Input: aspect=actuals and aspect=forecast or aspect=actuals
// Expected: (actuals AND forecast) OR actuals
func TestParseInput_AndBindsTighterThanOr_AndFirst(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x where aspect=actuals and aspect=forecast or aspect=actuals
		calculate
			0`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	require.Len(t, metricDef.Inputs, 1, "MetricDefinition should have 1 input")

	require.NotNil(t, metricDef.Inputs[0].WhereClause, "Should have a where clause")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups, 2, "Should have 2 AndGroups")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions, 2, "First AndGroup should have 2 conditions")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[1].Conditions, 1, "Second AndGroup should have 1 condition")
	assert.Equal(t, "ACTUALS", metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect, "First AndGroup first aspect should be ACTUALS")
	assert.Equal(t, "FORECAST", metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[1].Aspect, "First AndGroup second aspect should be FORECAST")
	assert.Equal(t, "ACTUALS", metricDef.Inputs[0].WhereClause.AndGroups[1].Conditions[0].Aspect, "Second AndGroup aspect should be ACTUALS")
}

// TestParseInput_AndBindsTighterThanOr_OrFirst tests that AND binds tighter than OR.
// Input: aspect=actuals or aspect=forecast and aspect=actuals
// Expected: actuals OR (forecast AND actuals)
func TestParseInput_AndBindsTighterThanOr_OrFirst(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x where aspect=actuals or aspect=forecast and aspect=actuals
		calculate
			0`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	require.Len(t, metricDef.Inputs, 1, "MetricDefinition should have 1 input")

	require.NotNil(t, metricDef.Inputs[0].WhereClause, "Should have a where clause")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups, 2, "Should have 2 AndGroups")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions, 1, "First AndGroup should have 1 condition")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[1].Conditions, 2, "Second AndGroup should have 2 conditions")
	assert.Equal(t, "ACTUALS", metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect, "First AndGroup aspect should be ACTUALS")
	assert.Equal(t, "FORECAST", metricDef.Inputs[0].WhereClause.AndGroups[1].Conditions[0].Aspect, "Second AndGroup first aspect should be FORECAST")
	assert.Equal(t, "ACTUALS", metricDef.Inputs[0].WhereClause.AndGroups[1].Conditions[1].Aspect, "Second AndGroup second aspect should be ACTUALS")
}

// TestParseInput_MultipleOrWithAnd tests multiple ORs with AND groups.
// Input: aspect=actuals and aspect=forecast or aspect=actuals and aspect=forecast
// Expected: (actuals AND forecast) OR (actuals AND forecast)
func TestParseInput_MultipleOrWithAnd(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x where aspect=actuals and aspect=forecast or aspect=actuals and aspect=forecast
		calculate
			0`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	require.Len(t, metricDef.Inputs, 1, "MetricDefinition should have 1 input")

	require.NotNil(t, metricDef.Inputs[0].WhereClause, "Should have a where clause")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups, 2, "Should have 2 AndGroups")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions, 2, "First AndGroup should have 2 conditions")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[1].Conditions, 2, "Second AndGroup should have 2 conditions")
	assert.Equal(t, "ACTUALS", metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect, "First AndGroup first aspect should be ACTUALS")
	assert.Equal(t, "FORECAST", metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[1].Aspect, "First AndGroup second aspect should be FORECAST")
	assert.Equal(t, "ACTUALS", metricDef.Inputs[0].WhereClause.AndGroups[1].Conditions[0].Aspect, "Second AndGroup first aspect should be ACTUALS")
	assert.Equal(t, "FORECAST", metricDef.Inputs[0].WhereClause.AndGroups[1].Conditions[1].Aspect, "Second AndGroup second aspect should be FORECAST")
}

// TestParseInput_SingleAndGroup tests a single AND group with two conditions.
// Input: aspect=actuals and aspect=forecast
// Expected: 1 AndGroup with 2 conditions
func TestParseInput_SingleAndGroup(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x where aspect=actuals and aspect=forecast
		calculate
			0`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	require.Len(t, metricDef.Inputs, 1, "MetricDefinition should have 1 input")

	require.NotNil(t, metricDef.Inputs[0].WhereClause, "Should have a where clause")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups, 1, "Should have 1 AndGroup")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions, 2, "AndGroup should have 2 conditions")
	assert.Equal(t, "ACTUALS", metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect, "First aspect should be ACTUALS")
	assert.Equal(t, "FORECAST", metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[1].Aspect, "Second aspect should be FORECAST")
}

// TestParseInput_OrWithNotAnd tests OR with NOT and AND.
// Input: aspect=actuals or not aspect=forecast and aspect=actuals
// Expected: actuals OR (not forecast AND actuals)
func TestParseInput_OrWithNotAnd(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x where aspect=actuals or not aspect=forecast and aspect=actuals
		calculate
			0`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	require.Len(t, metricDef.Inputs, 1, "MetricDefinition should have 1 input")

	require.NotNil(t, metricDef.Inputs[0].WhereClause, "Should have a where clause")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups, 2, "Should have 2 AndGroups")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions, 1, "First AndGroup should have 1 condition")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[1].Conditions, 2, "Second AndGroup should have 2 conditions")
	assert.Equal(t, "ACTUALS", metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect, "First AndGroup aspect should be ACTUALS")
	assert.False(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Negated, "First AndGroup condition should not be negated")
	assert.Equal(t, "FORECAST", metricDef.Inputs[0].WhereClause.AndGroups[1].Conditions[0].Aspect, "Second AndGroup first aspect should be FORECAST")
	assert.True(t, metricDef.Inputs[0].WhereClause.AndGroups[1].Conditions[0].Negated, "Second AndGroup first condition should be negated")
	assert.Equal(t, "ACTUALS", metricDef.Inputs[0].WhereClause.AndGroups[1].Conditions[1].Aspect, "Second AndGroup second aspect should be ACTUALS")
	assert.False(t, metricDef.Inputs[0].WhereClause.AndGroups[1].Conditions[1].Negated, "Second AndGroup second condition should not be negated")
}

// requireMetricDefinition is a helper that extracts a ParserMetricDefinition from a MetricScope.
// It fails the test if the Data field is not populated or is the wrong type.
// It also verifies that the MetricScope.Definitions slice contains a MetricScopeDefinition
// entry with the same metric name and ParserMetricDefinition pointer.
func requireMetricDefinition(t *testing.T, metricScope *symboltable.SymbolTable) *parsing.ParserMetricDefinition {
	t.Helper()

	require.NotNil(t, metricScope.Data, "MetricScope.Data should be populated")
	metricDef, ok := metricScope.Data.(*parsing.ParserMetricDefinition)
	require.True(t, ok, "MetricScope.Data should be a *parsing.ParserMetricDefinition")
	require.NotNil(t, metricDef, "MetricScope.Data should not contain a nil ParserMetricDefinition pointer")

	// Verify Definitions contains a MetricScopeDefinition entry
	found := false
	for _, def := range metricScope.Definitions {
		if def.DefinitionType == symboltable.MetricScopeDefinition {
			require.NotNil(t, def.Name, "MetricScopeDefinition should have a name")
			assert.Equal(t, metricDef.Name, *def.Name, "MetricScopeDefinition name should match metric name")
			assert.Equal(t, metricDef, def.Data, "MetricScopeDefinition Data should be the same *parsing.ParserMetricDefinition pointer")
			found = true
			break
		}
	}
	require.True(t, found, "MetricScope.Definitions should contain a MetricScopeDefinition entry")

	return metricDef
}
