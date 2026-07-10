// Package parser_test contains tests for the CPL metrics listener.
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
// Time Shift Listener Tests
// ============================================================================

// TestParseInputWithBackwardTimeShift verifies parsing an input with backward(year).
func TestParseInputWithBackwardTimeShift(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x backward(year)
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
	assert.Equal(t, "x", metricDef.Inputs[0].Name, "Input name should be 'x'")

	require.NotNil(t, metricDef.Inputs[0].TimeShift, "TimeShift should not be nil")
	assert.Equal(t, "backward", metricDef.Inputs[0].TimeShift.Direction, "Direction should be 'backward'")
	assert.Equal(t, "year", metricDef.Inputs[0].TimeShift.Field, "Field should be 'year'")
	assert.Equal(t, 1.0, metricDef.Inputs[0].TimeShift.Amount, "Amount should default to 1.0")
}

// TestParseInputWithForwardTimeShift verifies parsing an input with forward(month).
func TestParseInputWithForwardTimeShift(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x forward(month)
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
	assert.Equal(t, "x", metricDef.Inputs[0].Name, "Input name should be 'x'")

	require.NotNil(t, metricDef.Inputs[0].TimeShift, "TimeShift should not be nil")
	assert.Equal(t, "forward", metricDef.Inputs[0].TimeShift.Direction, "Direction should be 'forward'")
	assert.Equal(t, "month", metricDef.Inputs[0].TimeShift.Field, "Field should be 'month'")
	assert.Equal(t, 1.0, metricDef.Inputs[0].TimeShift.Amount, "Amount should default to 1.0")
}

// TestParseInputWithBackwardTimeShiftAndAmount verifies parsing an input with backward(day, 2).
func TestParseInputWithBackwardTimeShiftAndAmount(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x backward(day, 2)
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
	assert.Equal(t, "x", metricDef.Inputs[0].Name, "Input name should be 'x'")

	require.NotNil(t, metricDef.Inputs[0].TimeShift, "TimeShift should not be nil")
	assert.Equal(t, "backward", metricDef.Inputs[0].TimeShift.Direction, "Direction should be 'backward'")
	assert.Equal(t, "day", metricDef.Inputs[0].TimeShift.Field, "Field should be 'day'")
	require.NotNil(t, metricDef.Inputs[0].TimeShift.Expression, "Expression should not be nil for backward(day, 2)")
	assert.Len(t, metricDef.Inputs[0].TimeShift.Expression, 1, "Expression should have 1 element")
	if lit, ok := metricDef.Inputs[0].TimeShift.Expression[0].(parsing.PostfixLiteral); ok {
		assert.InDelta(t, 2.0, lit.Value, 0.001, "Expression literal value should be 2.0")
	} else {
		t.Fatalf("Expected PostfixLiteral, got %T", metricDef.Inputs[0].TimeShift.Expression[0])
	}
	assert.Equal(t, 1.0, metricDef.Inputs[0].TimeShift.Amount, "Amount should default to 1.0")
}

// TestParseInputWithBackwardTimeShiftAndArithmeticExpression verifies parsing
// backward(day, 1 + 1) captures the arithmetic expression as a multi-element
// postfix sequence in ParserTimeShift.Expression.
func TestParseInputWithBackwardTimeShiftAndArithmeticExpression(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x backward(day, 1 + 1)
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
	assert.Equal(t, "x", metricDef.Inputs[0].Name, "Input name should be 'x'")

	require.NotNil(t, metricDef.Inputs[0].TimeShift, "TimeShift should not be nil")
	assert.Equal(t, "backward", metricDef.Inputs[0].TimeShift.Direction, "Direction should be 'backward'")
	assert.Equal(t, "day", metricDef.Inputs[0].TimeShift.Field, "Field should be 'day'")

	// Verify postfix structure: [Literal(1), Literal(1), Operator(+)]
	require.NotNil(t, metricDef.Inputs[0].TimeShift.Expression, "Expression should not be nil")
	assert.Len(t, metricDef.Inputs[0].TimeShift.Expression, 3, "Expression should have 3 elements (1, 1, +)")

	// First element: literal 1
	if lit, ok := metricDef.Inputs[0].TimeShift.Expression[0].(parsing.PostfixLiteral); ok {
		assert.InDelta(t, 1.0, lit.Value, 0.001, "First expression element should be 1.0")
	} else {
		t.Fatalf("Expected PostfixLiteral at index 0, got %T", metricDef.Inputs[0].TimeShift.Expression[0])
	}

	// Second element: literal 1
	if lit, ok := metricDef.Inputs[0].TimeShift.Expression[1].(parsing.PostfixLiteral); ok {
		assert.InDelta(t, 1.0, lit.Value, 0.001, "Second expression element should be 1.0")
	} else {
		t.Fatalf("Expected PostfixLiteral at index 1, got %T", metricDef.Inputs[0].TimeShift.Expression[1])
	}

	// Third element: plus operator
	if op, ok := metricDef.Inputs[0].TimeShift.Expression[2].(parsing.PostfixOperator); ok {
		assert.Equal(t, symboltable.PlusOperatorDefinition, op.DefType,
			"Third expression element should be a plus operator")
	} else {
		t.Fatalf("Expected PostfixOperator at index 2, got %T", metricDef.Inputs[0].TimeShift.Expression[2])
	}

	assert.Equal(t, 1.0, metricDef.Inputs[0].TimeShift.Amount, "Amount should default to 1.0")
}

// TestParseInputWithoutTimeShift verifies that TimeShift is nil when not specified.
func TestParseInputWithoutTimeShift(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x
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
	assert.Equal(t, "x", metricDef.Inputs[0].Name, "Input name should be 'x'")
	assert.Nil(t, metricDef.Inputs[0].TimeShift, "TimeShift should be nil when not specified")
}

// TestParseInputWithAliasAndTimeShift verifies alias and time shift together.
func TestParseInputWithAliasAndTimeShift(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			dimension d as my_dim backward(hour, 3)
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

	assert.Equal(t, "dimension", metricDef.Inputs[0].Kind, "Input kind should be 'dimension'")
	assert.Equal(t, "d", metricDef.Inputs[0].Name, "Input name should be 'd'")
	assert.Equal(t, "my_dim", metricDef.Inputs[0].Alias, "Input alias should be 'my_dim'")

	require.NotNil(t, metricDef.Inputs[0].TimeShift, "TimeShift should not be nil")
	assert.Equal(t, "backward", metricDef.Inputs[0].TimeShift.Direction, "Direction should be 'backward'")
	assert.Equal(t, "hour", metricDef.Inputs[0].TimeShift.Field, "Field should be 'hour'")
	require.NotNil(t, metricDef.Inputs[0].TimeShift.Expression, "Expression should not be nil for backward(hour, 3)")
	assert.Len(t, metricDef.Inputs[0].TimeShift.Expression, 1, "Expression should have 1 element")
	if lit, ok := metricDef.Inputs[0].TimeShift.Expression[0].(parsing.PostfixLiteral); ok {
		assert.InDelta(t, 3.0, lit.Value, 0.001, "Expression literal value should be 3.0")
	} else {
		t.Fatalf("Expected PostfixLiteral, got %T", metricDef.Inputs[0].TimeShift.Expression[0])
	}
	assert.Equal(t, 1.0, metricDef.Inputs[0].TimeShift.Amount, "Amount should default to 1.0")
}

// TestParseInputWithWhereAndTimeShift verifies where clause and time shift together.
func TestParseInputWithWhereAndTimeShift(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x where aspect=actuals backward(week_number)
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
	assert.Equal(t, "x", metricDef.Inputs[0].Name, "Input name should be 'x'")

	require.NotNil(t, metricDef.Inputs[0].WhereClause, "WhereClause should not be nil")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups, 1, "Should have 1 AndGroup")
	require.Len(t, metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions, 1, "Should have 1 condition")
	assert.Equal(t, "ACTUALS", metricDef.Inputs[0].WhereClause.AndGroups[0].Conditions[0].Aspect, "Aspect should be ACTUALS")

	require.NotNil(t, metricDef.Inputs[0].TimeShift, "TimeShift should not be nil")
	assert.Equal(t, "backward", metricDef.Inputs[0].TimeShift.Direction, "Direction should be 'backward'")
	assert.Equal(t, "week_number", metricDef.Inputs[0].TimeShift.Field, "Field should be 'week_number'")
	assert.Equal(t, 1.0, metricDef.Inputs[0].TimeShift.Amount, "Amount should default to 1.0")
}

// TestParseMultipleInputsWithMixedTimeShifts verifies time shift doesn't leak between inputs.
func TestParseMultipleInputsWithMixedTimeShifts(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric a backward(year)
			metric b
			metric c forward(month, 2)
		calculate
			0`

	doc, err := parser.Parse(input)
	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	require.Len(t, metricDef.Inputs, 3, "MetricDefinition should have 3 inputs")

	// First input has backward time shift
	assert.Equal(t, "a", metricDef.Inputs[0].Name)
	require.NotNil(t, metricDef.Inputs[0].TimeShift)
	assert.Equal(t, "backward", metricDef.Inputs[0].TimeShift.Direction)
	assert.Equal(t, "year", metricDef.Inputs[0].TimeShift.Field)
	assert.Equal(t, 1.0, metricDef.Inputs[0].TimeShift.Amount)

	// Second input has no time shift
	assert.Equal(t, "b", metricDef.Inputs[1].Name)
	assert.Nil(t, metricDef.Inputs[1].TimeShift, "TimeShift should be nil for second input")

	// Third input has forward time shift with amount
	assert.Equal(t, "c", metricDef.Inputs[2].Name)
	require.NotNil(t, metricDef.Inputs[2].TimeShift)
	assert.Equal(t, "forward", metricDef.Inputs[2].TimeShift.Direction)
	assert.Equal(t, "month", metricDef.Inputs[2].TimeShift.Field)
	require.NotNil(t, metricDef.Inputs[2].TimeShift.Expression, "Expression should not be nil for forward(month, 2)")
	assert.Len(t, metricDef.Inputs[2].TimeShift.Expression, 1, "Expression should have 1 element")
	if lit, ok := metricDef.Inputs[2].TimeShift.Expression[0].(parsing.PostfixLiteral); ok {
		assert.InDelta(t, 2.0, lit.Value, 0.001, "Expression literal value should be 2.0")
	} else {
		t.Fatalf("Expected PostfixLiteral, got %T", metricDef.Inputs[2].TimeShift.Expression[0])
	}
	assert.Equal(t, 1.0, metricDef.Inputs[2].TimeShift.Amount, "Amount should default to 1.0")
}

// TestParseInputWithBothBackwardAndForward_ReturnsError verifies that specifying
// both backward() and forward() on the same input declaration produces an error.
// The grammar requires backward() before forward(); the listener rejects both.
func TestParseInputWithBothBackwardAndForward_ReturnsError(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x backward(year) forward(month)
		calculate
			0`

	_, err := parser.Parse(input)
	require.Error(t, err, "Expected error when both backward() and forward() are specified")
	assert.Contains(t, err.Error(), "cannot have both backward() and forward()")
	assert.Contains(t, err.Error(), "line")
	assert.Contains(t, err.Error(), "column")
}

// TestParseInputWithTrailingComma_ReturnsError verifies that a trailing comma
// inside a backward() clause (e.g., backward(year,)) produces a parse error.
// The grammar rule shiftExpression requires an expression after the comma.
func TestParseInputWithTrailingComma_ReturnsError(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x backward(year,)
		calculate
			0`

	_, err := parser.Parse(input)
	require.Error(t, err, "Expected error for trailing comma in backward(year,)")
}

// TestParseInputWithForwardBeforeBackward_ReturnsError verifies that specifying
// forward() before backward() on the same input declaration produces a parse error.
// The grammar rule inputDeclaration has backwardClause? before forwardClause?,
// so forward(x) backward(y) is syntactically invalid.
func TestParseInputWithForwardBeforeBackward_ReturnsError(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric x forward(month) backward(year)
		calculate
			0`

	_, err := parser.Parse(input)
	require.Error(t, err, "Expected error when forward() is placed before backward()")
	assert.NotContains(t, err.Error(), "cannot have both backward() and forward()",
		"Error should be a grammar/parse error, not a listener semantic error")
}
