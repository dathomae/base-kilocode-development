// Package parser_test contains tests for CPL arithmetic expression parsing.
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
// Postfix Element Assertion Helpers
// ============================================================================

// assertPostfixSequence verifies the CalculatePostfix elements match the
// expected typed values. Expected values can be:
//   - string: asserted as PostfixIdentifier with that name
//   - float64: asserted as PostfixLiteral with that value
//   - symboltable.DefinitionType: asserted as PostfixOperator with that DefType
func assertPostfixSequence(t *testing.T, postfix []parsing.PostfixElement, expected []parsing.PostfixElement, msgAndArgs ...interface{}) {
	t.Helper()
	require.Len(t, postfix, len(expected), "Postfix length should match expected")

	for i, exp := range expected {
		switch expVal := exp.(type) {
		case parsing.PostfixIdentifier:
			id, ok := postfix[i].(parsing.PostfixIdentifier)
			if !ok {
				t.Fatalf("Postfix element %d: expected PostfixIdentifier, got %T", i, postfix[i])
			}
			assert.Equal(t, expVal.Name, id.Name, "Postfix element %d name should match", i)
		case parsing.PostfixLiteral:
			lit, ok := postfix[i].(parsing.PostfixLiteral)
			if !ok {
				t.Fatalf("Postfix element %d: expected PostfixLiteral, got %T", i, postfix[i])
			}
			assert.Equal(t, expVal.Value, lit.Value, "Postfix element %d value should match", i)
		case parsing.PostfixOperator:
			op, ok := postfix[i].(parsing.PostfixOperator)
			if !ok {
				t.Fatalf("Postfix element %d: expected PostfixOperator, got %T", i, postfix[i])
			}
			assert.Equal(t, expVal.DefType, op.DefType, "Postfix element %d DefType should match", i)
		default:
			t.Fatalf("Postfix element %d: unsupported expected type %T", i, exp)
		}
	}
}

// TestParseExpression_BasicCases verifies postfix generation for simple literals
// and single binary operators using a table-driven pattern. Each case parses a
// metric whose calculate clause contains a basic expression and asserts the
// resulting postfix sequence on ParserMetricDefinition.CalculatePostfix.
//
// Covered cases:
//   - Simple literal: calculate 42 → [42]
//   - Single addition: calculate a + b → ["a", "b", PlusOperatorDefinition]
//   - Single subtraction: calculate a - b → ["a", "b", MinusOperatorDefinition]
//   - Single multiplication: calculate a * b → ["a", "b", MultiplyOperatorDefinition]
//   - Single division: calculate a / b → ["a", "b", DivideOperatorDefinition]
func TestParseExpression_BasicCases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		input           string
		expectedPostfix []parsing.PostfixElement
	}{
		{
			name: "simple literal",
			input: `metrics
	my_metric
		inputs
			metric a
		calculate
			42`,
			expectedPostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 42}},
		},
		{
			name: "single addition",
			input: `metrics
	my_metric
		inputs
			metric a
			metric b
		calculate
			a + b`,
			expectedPostfix: []parsing.PostfixElement{parsing.PostfixIdentifier{Name: "a"}, parsing.PostfixIdentifier{Name: "b"}, parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition}},
		},
		{
			name: "single subtraction",
			input: `metrics
	my_metric
		inputs
			metric a
			metric b
		calculate
			a - b`,
			expectedPostfix: []parsing.PostfixElement{parsing.PostfixIdentifier{Name: "a"}, parsing.PostfixIdentifier{Name: "b"}, parsing.PostfixOperator{DefType: symboltable.MinusOperatorDefinition}},
		},
		{
			name: "single multiplication",
			input: `metrics
	my_metric
		inputs
			metric a
			metric b
		calculate
			a * b`,
			expectedPostfix: []parsing.PostfixElement{parsing.PostfixIdentifier{Name: "a"}, parsing.PostfixIdentifier{Name: "b"}, parsing.PostfixOperator{DefType: symboltable.MultiplyOperatorDefinition}},
		},
		{
			name: "single division",
			input: `metrics
	my_metric
		inputs
			metric a
			metric b
		calculate
			a / b`,
			expectedPostfix: []parsing.PostfixElement{parsing.PostfixIdentifier{Name: "a"}, parsing.PostfixIdentifier{Name: "b"}, parsing.PostfixOperator{DefType: symboltable.DivideOperatorDefinition}},
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
			assertPostfixSequence(t, metricDef.CalculatePostfix, tc.expectedPostfix)
		})
	}
}

// TestParseExpression_OperatorPrecedence verifies that multiplication has higher
// precedence than addition, so the multiplication is evaluated first in postfix.
//
// Input:
//
//	metrics
//	    my_metric
//	        inputs
//	            metric a
//	            metric b
//	            metric c
//	        calculate
//	            a + b * c
//
// Expected postfix: ["a", "b", "c", MultiplyOperatorDefinition, PlusOperatorDefinition]
func TestParseExpression_OperatorPrecedence(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric a
			metric b
			metric c
		calculate
			a + b * c`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	expectedPostfix := []parsing.PostfixElement{
		parsing.PostfixIdentifier{Name: "a"}, parsing.PostfixIdentifier{Name: "b"}, parsing.PostfixIdentifier{Name: "c"}, parsing.PostfixOperator{DefType: symboltable.MultiplyOperatorDefinition}, parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition},
	}
	assertPostfixSequence(t, metricDef.CalculatePostfix, expectedPostfix)
}

// TestParseExpression_ParenthesizedSubexpression verifies that parentheses
// override default precedence, forcing the addition to be evaluated before
// the multiplication.
//
// Input:
//
//	metrics
//	    my_metric
//	        inputs
//	            metric a
//	            metric b
//	            metric c
//	        calculate
//	            (a + b) * c
//
// Expected postfix: ["a", "b", PlusOperatorDefinition, "c", MultiplyOperatorDefinition]
func TestParseExpression_ParenthesizedSubexpression(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric a
			metric b
			metric c
		calculate
			(a + b) * c`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	expectedPostfix := []parsing.PostfixElement{
		parsing.PostfixIdentifier{Name: "a"}, parsing.PostfixIdentifier{Name: "b"}, parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition}, parsing.PostfixIdentifier{Name: "c"}, parsing.PostfixOperator{DefType: symboltable.MultiplyOperatorDefinition},
	}
	assertPostfixSequence(t, metricDef.CalculatePostfix, expectedPostfix)
}

// TestParseExpression_UnaryNegation verifies that a unary minus prefix on an
// identifier produces a postfix sequence with the operand followed by the
// negate operator.
//
// Input:
//
//	metrics
//	    my_metric
//	        inputs
//	            metric a
//	        calculate
//	            -a
//
// Expected postfix: ["a", NegateOperatorDefinition]
func TestParseExpression_UnaryNegation(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric a
		calculate
			-a`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	expectedPostfix := []parsing.PostfixElement{
		parsing.PostfixIdentifier{Name: "a"}, parsing.PostfixOperator{DefType: symboltable.NegateOperatorDefinition},
	}
	assertPostfixSequence(t, metricDef.CalculatePostfix, expectedPostfix)
}

// TestParseExpression_UnaryNegationWithOperators verifies that unary negation
// interacts correctly with binary operators, with the negation applied to
// the first operand before the addition.
//
// Input:
//
//	metrics
//	    my_metric
//	        inputs
//	            metric a
//	            metric b
//	        calculate
//	            -a + b
//
// Expected postfix: ["a", NegateOperatorDefinition, "b", PlusOperatorDefinition]
func TestParseExpression_UnaryNegationWithOperators(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric a
			metric b
		calculate
			-a + b`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	expectedPostfix := []parsing.PostfixElement{
		parsing.PostfixIdentifier{Name: "a"}, parsing.PostfixOperator{DefType: symboltable.NegateOperatorDefinition}, parsing.PostfixIdentifier{Name: "b"}, parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition},
	}
	assertPostfixSequence(t, metricDef.CalculatePostfix, expectedPostfix)
}

// TestParseExpression_ComplexExpression verifies that a complex expression
// combining addition, subtraction, multiplication, and division with correct
// precedence produces the expected postfix sequence.
//
// Input:
//
//	metrics
//	    my_metric
//	        inputs
//	            metric a
//	            metric b
//	            metric c
//	            metric d
//	            metric e
//	        calculate
//	            a + b * c - d / e
//
// Expected postfix:
//
//	["a", "b", "c", MultiplyOperatorDefinition, PlusOperatorDefinition,
//	 "d", "e", DivideOperatorDefinition, MinusOperatorDefinition]
func TestParseExpression_ComplexExpression(t *testing.T) {
	t.Parallel()

	input := `metrics
	my_metric
		inputs
			metric a
			metric b
			metric c
			metric d
			metric e
		calculate
			a + b * c - d / e`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (MetricsScope)")
	metricsScope := doc.Children[0]
	require.Len(t, metricsScope.Children, 1, "Expected one MetricScope")

	metricDef := requireMetricDefinition(t, metricsScope.Children[0])
	expectedPostfix := []parsing.PostfixElement{
		parsing.PostfixIdentifier{Name: "a"}, parsing.PostfixIdentifier{Name: "b"}, parsing.PostfixIdentifier{Name: "c"}, parsing.PostfixOperator{DefType: symboltable.MultiplyOperatorDefinition}, parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition},
		parsing.PostfixIdentifier{Name: "d"}, parsing.PostfixIdentifier{Name: "e"}, parsing.PostfixOperator{DefType: symboltable.DivideOperatorDefinition}, parsing.PostfixOperator{DefType: symboltable.MinusOperatorDefinition},
	}
	assertPostfixSequence(t, metricDef.CalculatePostfix, expectedPostfix)
}

// TestParseExpression_ConditionalExprDelegation verifies that the expression
// rule correctly delegates through the full conditional expression chain
// (expression → conditionalExpr → whenExpr → logicalOrExpr → ... → additionExpr)
// for arithmetic expressions. This is a combined structural smoke test for the
// Task 1a conditionalExpr and Task 1b whenExpr skeletons — it confirms that
// adding these rules to the delegation chain did not break existing arithmetic
// parsing and that the grammar remains compilable with the new rules in place.
//
// Covered cases:
//   - Simple literal falls through conditionalExpr: calculate 42 → [42]
//   - Binary addition falls through conditionalExpr:
//     calculate a + b → ["a", "b", PlusOperatorDefinition]
//   - Parenthesized subexpression with mixed precedence falls through
//     conditionalExpr:
//     calculate (a + b) * c → ["a", "b", PlusOperatorDefinition, "c", MultiplyOperatorDefinition]
func TestParseExpression_ConditionalExprDelegation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		input           string
		expectedPostfix []parsing.PostfixElement
	}{
		{
			name: "simple literal through conditionalExpr delegation",
			input: `metrics
	my_metric
		inputs
			metric a
		calculate
			42`,
			expectedPostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 42}},
		},
		{
			name: "addition through conditionalExpr delegation",
			input: `metrics
	my_metric
		inputs
			metric a
			metric b
		calculate
			a + b`,
			expectedPostfix: []parsing.PostfixElement{parsing.PostfixIdentifier{Name: "a"}, parsing.PostfixIdentifier{Name: "b"}, parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition}},
		},
		{
			name: "parenthesized expression through conditionalExpr delegation",
			input: `metrics
	my_metric
		inputs
			metric a
			metric b
			metric c
		calculate
			(a + b) * c`,
			expectedPostfix: []parsing.PostfixElement{
				parsing.PostfixIdentifier{Name: "a"}, parsing.PostfixIdentifier{Name: "b"}, parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition}, parsing.PostfixIdentifier{Name: "c"}, parsing.PostfixOperator{DefType: symboltable.MultiplyOperatorDefinition},
			},
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
			assertPostfixSequence(t, metricDef.CalculatePostfix, tc.expectedPostfix)
		})
	}
}
