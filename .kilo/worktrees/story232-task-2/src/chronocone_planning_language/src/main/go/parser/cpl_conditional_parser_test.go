// Package parser_test contains tests for CPL conditional expression parsing.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requireClauseData extracts ClauseData from a SymbolDefinition.
// Fails the test if the Data field is nil or has the wrong type.
func requireClauseData(t *testing.T, def symboltable.SymbolDefinition) symboltable.ClauseData {
	t.Helper()
	require.NotNil(t, def.Data, "Definition Data should not be nil")
	clauseData, ok := def.Data.(symboltable.ClauseData)
	require.True(t, ok, "Definition Data should be ClauseData, got %T", def.Data)
	return clauseData
}

// findDefByType finds the first SymbolDefinition matching the given DefinitionType
// in the scope's Definitions. Returns nil if not found.
func findDefByType(defs []symboltable.SymbolDefinition, defType symboltable.DefinitionType) *symboltable.SymbolDefinition {
	for i := range defs {
		if defs[i].DefinitionType == defType {
			return &defs[i]
		}
	}
	return nil
}

// assertPostfixElement asserts a postfix element matches an expected SymbolDefinition.
func assertPostfixElement(t *testing.T, expected symboltable.SymbolDefinition, actual symboltable.SymbolDefinition, index int) {
	t.Helper()
	assert.Equal(t, expected.DefinitionType, actual.DefinitionType,
		"Postfix element %d: DefinitionType mismatch", index)
	if expected.Name != nil {
		require.NotNil(t, actual.Name, "Postfix element %d: expected Name not nil", index)
		assert.Equal(t, *expected.Name, *actual.Name,
			"Postfix element %d: Name mismatch", index)
	} else {
		assert.Nil(t, actual.Name, "Postfix element %d: expected Name nil", index)
	}
}

// TestIfBranchCreatesChildScope verifies that parsing a simple if expression
// with a single if-branch produces an IfBodyScope child scope. The condition
// expression is parsed in the parent scope (MetricScope for calculate clauses),
// and the body expression is parsed in the child IfBodyScope.
func TestIfBranchCreatesChildScope(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                   string
		input                  string
		expectIfBodyScope      bool
		expectedBodyChildCount int
	}{
		{
			name: "simple if with literal body",
			input: `metrics
	test_metric
		inputs
			metric x
		calculate
			if
				1 > 0
					42 end if`,
			expectIfBodyScope:      true,
			expectedBodyChildCount: 2, // TempScope + IfBodyScope
		},
		{
			name: "if with arithmetic condition",
			input: `metrics
	test_metric
		inputs
			metric a
			metric b
		calculate
			if
				a + b > 0
					a end if`,
			expectIfBodyScope:      true,
			expectedBodyChildCount: 2, // TempScope + IfBodyScope
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
			require.Equal(t, symboltable.MetricsScope, metricsScope.ScopeType)
			require.Len(t, metricsScope.Children, 1, "Expected one MetricScope child")
			metricScope := metricsScope.Children[0]
			require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

			// Check for IfClauseDefinition in MetricScope.Definitions
			ifDef := findDefByType(metricScope.Definitions, symboltable.IfClauseDefinition)
			require.NotNil(t, ifDef, "Expected IfClauseDefinition in MetricScope.Definitions")
			_ = requireClauseData(t, *ifDef)

			// Check for IfBodyScope child
			require.Len(t, metricScope.Children, tc.expectedBodyChildCount,
				"Expected %d child scope(s) in MetricScope", tc.expectedBodyChildCount)
			if tc.expectIfBodyScope {
				var bodyScope *symboltable.SymbolTable
				for _, child := range metricScope.Children {
					if child.ScopeType == symboltable.IfBodyScope {
						bodyScope = child
						break
					}
				}
				require.NotNil(t, bodyScope, "Expected IfBodyScope in MetricScope.Children")
				assert.Equal(t, symboltable.IfBodyScope, bodyScope.ScopeType,
					"Should find IfBodyScope child")
			}
		})
	}
}

// TestIfBranchStoresClauseData verifies that after parsing an if expression,
// the parent scope (MetricScope) contains an IfClauseDefinition with ClauseData
// carrying the condition's postfix definitions.
func TestIfBranchStoresClauseData(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		input               string
		expectedClauseTypes []symboltable.DefinitionType // expected clause def types in order
	}{
		{
			name: "simple if with identifier condition",
			input: `metrics
	test_metric
		inputs
			metric x
		calculate
			if
				x
					42 end if`,
			expectedClauseTypes: []symboltable.DefinitionType{symboltable.IfClauseDefinition},
		},
		{
			name: "if with comparison condition",
			input: `metrics
	test_metric
		inputs
			metric x
		calculate
			if
				x > 0
					x end if`,
			expectedClauseTypes: []symboltable.DefinitionType{symboltable.IfClauseDefinition},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := parser.Parse(tc.input)
			require.NoError(t, err, "Expected no error for input: %s", tc.input)
			require.NotNil(t, doc, "Document should not be nil")
			metricsScope := doc.Children[0]
			metricScope := metricsScope.Children[0]

			// Verify IfClauseDefinition exists with ClauseData
			ifDef := findDefByType(metricScope.Definitions, symboltable.IfClauseDefinition)
			require.NotNil(t, ifDef, "Expected IfClauseDefinition in MetricScope.Definitions")
			clauseData := requireClauseData(t, *ifDef)
			require.NotNil(t, clauseData.ConditionPostfix,
				"ConditionPostfix should not be nil")
		})
	}
}

// TestIfBranchConditionPostfix verifies that condition expressions in if-branches
// produce correct postfix sequences in the ClauseData stored on the
// IfClauseDefinition.
func TestIfBranchConditionPostfix(t *testing.T) {
	t.Parallel()

	// Helper to create expected SymbolDefinition
	makeDef := func(defType symboltable.DefinitionType, name string) symboltable.SymbolDefinition {
		var namePtr *string
		if name != "" {
			namePtr = &name
		}
		return symboltable.SymbolDefinition{
			DefinitionType: defType,
			Name:           namePtr,
		}
	}

	testCases := []struct {
		name            string
		input           string
		expectedPostfix []symboltable.SymbolDefinition
	}{
		{
			name: "simple identifier condition",
			input: `metrics
	test_metric
		inputs
			metric x
		calculate
			if
				x
					42 end if`,
			expectedPostfix: []symboltable.SymbolDefinition{
				makeDef(symboltable.Identifier, ""),
			},
		},
		{
			name: "greater than comparison",
			input: `metrics
	test_metric
		inputs
			metric x
		calculate
			if
				x > 0
					42 end if`,
			expectedPostfix: []symboltable.SymbolDefinition{
				makeDef(symboltable.Identifier, ""),
				makeDef(symboltable.LiteralDefinition, ""),
				makeDef(symboltable.GreaterThanOperatorDefinition, ""),
			},
		},
		{
			name: "less than or equal comparison",
			input: `metrics
	test_metric
		inputs
			metric x
		calculate
			if
				x <= 5
					42 end if`,
			expectedPostfix: []symboltable.SymbolDefinition{
				makeDef(symboltable.Identifier, ""),
				makeDef(symboltable.LiteralDefinition, ""),
				makeDef(symboltable.LessThanEqOperatorDefinition, ""),
			},
		},
		{
			name: "equality comparison",
			input: `metrics
	test_metric
		inputs
			metric x
			metric y
		calculate
			if
				x == y
					42 end if`,
			expectedPostfix: []symboltable.SymbolDefinition{
				makeDef(symboltable.Identifier, ""),
				makeDef(symboltable.Identifier, ""),
				makeDef(symboltable.EqOperatorDefinition, ""),
			},
		},
		{
			name: "inequality comparison",
			input: `metrics
	test_metric
		inputs
			metric x
		calculate
			if
				x != 0
					42 end if`,
			expectedPostfix: []symboltable.SymbolDefinition{
				makeDef(symboltable.Identifier, ""),
				makeDef(symboltable.LiteralDefinition, ""),
				makeDef(symboltable.NotEqOperatorDefinition, ""),
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
			metricsScope := doc.Children[0]
			metricScope := metricsScope.Children[0]

			// Get the IfClauseDefinition
			ifDef := findDefByType(metricScope.Definitions, symboltable.IfClauseDefinition)
			require.NotNil(t, ifDef, "Expected IfClauseDefinition")
			clauseData := requireClauseData(t, *ifDef)

			// Verify postfix length
			require.Len(t, clauseData.ConditionPostfix, len(tc.expectedPostfix),
				"ConditionPostfix length mismatch")

			// Verify each postfix element
			for i, expected := range tc.expectedPostfix {
				assertPostfixElement(t, expected, clauseData.ConditionPostfix[i], i)
			}
		})
	}
}

// TestNestedIfBranch verifies that nested if expressions produce the correct
// multi-level scope hierarchy. Each level of nesting creates an IfBodyScope
// child of the parent scope, with its own IfClauseDefinition in the respective
// parent.
func TestNestedIfBranch(t *testing.T) {
	t.Parallel()

	t.Run("single nested if", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric a
			metric b
		calculate
			if
				a
					if
						b
							42 end if end if`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]
		require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

		// MetricScope should have IfClauseDefinition for outer condition "a"
		ifDef := findDefByType(metricScope.Definitions, symboltable.IfClauseDefinition)
		require.NotNil(t, ifDef, "Expected outer IfClauseDefinition in MetricScope")
		outerClauseData := requireClauseData(t, *ifDef)
		require.Len(t, outerClauseData.ConditionPostfix, 1,
			"Outer condition should have 1 element (identifier 'a')")

		// MetricScope should have 2 children: TempScope + outer IfBodyScope
		require.Len(t, metricScope.Children, 2, "MetricScope should have 2 children (TempScope + outer IfBodyScope)")
		var outerBodyScope *symboltable.SymbolTable
		for _, child := range metricScope.Children {
			if child.ScopeType == symboltable.IfBodyScope {
				outerBodyScope = child
				break
			}
		}
		require.NotNil(t, outerBodyScope, "Expected outer IfBodyScope in MetricScope.Children")
		assert.Equal(t, symboltable.IfBodyScope, outerBodyScope.ScopeType)

		// Outer IfBodyScope should have IfClauseDefinition for inner condition "b"
		innerIfDef := findDefByType(outerBodyScope.Definitions, symboltable.IfClauseDefinition)
		require.NotNil(t, innerIfDef, "Expected inner IfClauseDefinition in outer body scope")
		innerClauseData := requireClauseData(t, *innerIfDef)
		require.Len(t, innerClauseData.ConditionPostfix, 1,
			"Inner condition should have 1 element (identifier 'b')")

		// Outer IfBodyScope should have 2 children: TempScope + inner IfBodyScope
		require.Len(t, outerBodyScope.Children, 2, "Outer body scope should have 2 children (TempScope + inner IfBodyScope)")
		var innerBodyScope *symboltable.SymbolTable
		for _, child := range outerBodyScope.Children {
			if child.ScopeType == symboltable.IfBodyScope {
				innerBodyScope = child
				break
			}
		}
		require.NotNil(t, innerBodyScope, "Expected inner IfBodyScope in outer body scope Children")
		assert.Equal(t, symboltable.IfBodyScope, innerBodyScope.ScopeType)
	})

	t.Run("deeply nested 3 levels", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric a
			metric b
			metric c
		calculate
			if
				a
					if
						b
							if
								c
									42 end if end if end if`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]

		// Level 1: MetricScope → TempScope + IfBodyScope
		require.Len(t, metricScope.Children, 2)
		level1 := metricScope.Children[1] // IfBodyScope is second child
		assert.Equal(t, symboltable.IfBodyScope, level1.ScopeType)
		require.NotNil(t, findDefByType(metricScope.Definitions, symboltable.IfClauseDefinition))

		// Level 2: IfBodyScope → TempScope + IfBodyScope
		require.Len(t, level1.Children, 2)
		level2 := level1.Children[1]
		assert.Equal(t, symboltable.IfBodyScope, level2.ScopeType)
		require.NotNil(t, findDefByType(level1.Definitions, symboltable.IfClauseDefinition))

		// Level 3: IfBodyScope → TempScope + IfBodyScope (body of innermost if)
		require.Len(t, level2.Children, 2)
		level3 := level2.Children[1]
		assert.Equal(t, symboltable.IfBodyScope, level3.ScopeType)
		require.NotNil(t, findDefByType(level2.Definitions, symboltable.IfClauseDefinition))
	})
}

// TestIfBranchClauseDataBodyScope verifies that ClauseData.BodyScope correctly
// references the body scope child created during parsing of if-branch expressions.
func TestIfBranchClauseDataBodyScope(t *testing.T) {
	t.Parallel()

	t.Run("simple if with literal body", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric x
		calculate
			if
				x > 0
					42 end if`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]
		require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

		// The body scope should be among the children of MetricScope
		// (TempScope for condition + IfBodyScope for body)
		require.Len(t, metricScope.Children, 2, "MetricScope should have 2 children (TempScope + IfBodyScope)")
		var expectedBodyScope *symboltable.SymbolTable
		for _, child := range metricScope.Children {
			if child.ScopeType == symboltable.IfBodyScope {
				expectedBodyScope = child
				break
			}
		}
		require.NotNil(t, expectedBodyScope, "Expected IfBodyScope in MetricScope.Children")
		require.Equal(t, symboltable.IfBodyScope, expectedBodyScope.ScopeType)

		// Verify IfClauseDefinition has ClauseData with BodyScope set
		ifDef := findDefByType(metricScope.Definitions, symboltable.IfClauseDefinition)
		require.NotNil(t, ifDef, "Expected IfClauseDefinition in MetricScope.Definitions")
		clauseData := requireClauseData(t, *ifDef)

		require.NotNil(t, clauseData.BodyScope, "ClauseData.BodyScope should not be nil")
		require.Equal(t, symboltable.IfBodyScope, clauseData.BodyScope.ScopeType,
			"BodyScope should be IfBodyScope")
		require.Same(t, expectedBodyScope, clauseData.BodyScope,
			"BodyScope should be the same pointer as MetricScope's child IfBodyScope")

		// Body scope should contain the literal 42
		require.Len(t, clauseData.BodyScope.Definitions, 1,
			"Body scope should have 1 definition (literal 42)")
		assert.Equal(t, symboltable.LiteralDefinition, clauseData.BodyScope.Definitions[0].DefinitionType)
	})

	t.Run("if with arithmetic body", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric x
			metric a
			metric b
		calculate
			if
				x > 0
					a + b end if`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]
		require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

		require.Len(t, metricScope.Children, 2, "MetricScope should have 2 children (TempScope + IfBodyScope)")
		var expectedBodyScope *symboltable.SymbolTable
		for _, child := range metricScope.Children {
			if child.ScopeType == symboltable.IfBodyScope {
				expectedBodyScope = child
				break
			}
		}
		require.NotNil(t, expectedBodyScope, "Expected IfBodyScope in MetricScope.Children")
		require.Equal(t, symboltable.IfBodyScope, expectedBodyScope.ScopeType)

		ifDef := findDefByType(metricScope.Definitions, symboltable.IfClauseDefinition)
		require.NotNil(t, ifDef, "Expected IfClauseDefinition in MetricScope.Definitions")
		clauseData := requireClauseData(t, *ifDef)

		require.NotNil(t, clauseData.BodyScope, "ClauseData.BodyScope should not be nil")
		require.Equal(t, symboltable.IfBodyScope, clauseData.BodyScope.ScopeType,
			"BodyScope should be IfBodyScope")
		require.Same(t, expectedBodyScope, clauseData.BodyScope,
			"BodyScope should be the same pointer as MetricScope's child IfBodyScope")

		// Body scope should contain postfix for a + b: a, b, PlusOperator
		require.Len(t, clauseData.BodyScope.Definitions, 3,
			"Body scope should have 3 definitions (a, b, +)")
		expectedPostfix := []symboltable.SymbolDefinition{
			{DefinitionType: symboltable.Identifier},
			{DefinitionType: symboltable.Identifier},
			{DefinitionType: symboltable.PlusOperatorDefinition},
		}
		for i, expected := range expectedPostfix {
			assertPostfixElement(t, expected, clauseData.BodyScope.Definitions[i], i)
		}
		// Verify identifier data contains the correct names
		assert.Equal(t, symboltable.StringData{Value: "a"}, clauseData.BodyScope.Definitions[0].Data)
		assert.Equal(t, symboltable.StringData{Value: "b"}, clauseData.BodyScope.Definitions[1].Data)
	})

	t.Run("nested if", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric a
			metric b
		calculate
			if
				a
					if
						b
							42 end if end if`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]
		require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

		// Outer if
		require.Len(t, metricScope.Children, 2, "MetricScope should have 2 children (TempScope + outer IfBodyScope)")
		var outerBodyScope *symboltable.SymbolTable
		for _, child := range metricScope.Children {
			if child.ScopeType == symboltable.IfBodyScope {
				outerBodyScope = child
				break
			}
		}
		require.NotNil(t, outerBodyScope, "Expected outer IfBodyScope in MetricScope.Children")
		require.Equal(t, symboltable.IfBodyScope, outerBodyScope.ScopeType)

		outerIfDef := findDefByType(metricScope.Definitions, symboltable.IfClauseDefinition)
		require.NotNil(t, outerIfDef, "Expected outer IfClauseDefinition in MetricScope.Definitions")
		outerClauseData := requireClauseData(t, *outerIfDef)

		require.NotNil(t, outerClauseData.BodyScope, "Outer ClauseData.BodyScope should not be nil")
		require.Equal(t, symboltable.IfBodyScope, outerClauseData.BodyScope.ScopeType,
			"Outer BodyScope should be IfBodyScope")
		require.Same(t, outerBodyScope, outerClauseData.BodyScope,
			"Outer BodyScope should be the same pointer as MetricScope's child")

		// Inner if (inside outer body scope)
		require.Len(t, outerBodyScope.Children, 2,
			"Outer body scope should have 2 children (TempScope + inner IfBodyScope)")
		var innerBodyScope *symboltable.SymbolTable
		for _, child := range outerBodyScope.Children {
			if child.ScopeType == symboltable.IfBodyScope {
				innerBodyScope = child
				break
			}
		}
		require.NotNil(t, innerBodyScope, "Expected inner IfBodyScope in outer body scope Children")
		require.Equal(t, symboltable.IfBodyScope, innerBodyScope.ScopeType)

		innerIfDef := findDefByType(outerBodyScope.Definitions, symboltable.IfClauseDefinition)
		require.NotNil(t, innerIfDef, "Expected inner IfClauseDefinition in outer body scope")
		innerClauseData := requireClauseData(t, *innerIfDef)

		require.NotNil(t, innerClauseData.BodyScope, "Inner ClauseData.BodyScope should not be nil")
		require.Equal(t, symboltable.IfBodyScope, innerClauseData.BodyScope.ScopeType,
			"Inner BodyScope should be IfBodyScope")
		require.Same(t, innerBodyScope, innerClauseData.BodyScope,
			"Inner BodyScope should be the same pointer as outer body scope's child")
	})
}

// TestElseIfBranchCreatesChildScope verifies that parsing an if/else-if expression
// produces an ElseIfBodyScope child scope alongside the IfBodyScope. Both the
// IfClauseDefinition and ElseIfClauseDefinition are stored in the parent MetricScope.
func TestElseIfBranchCreatesChildScope(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		input              string
		expectedChildCount int
	}{
		{
			name: "if with else-if using comparison conditions",
			input: `metrics
	test_metric
		inputs
			metric x
			metric y
		calculate
			if
				x > 0
					1 else if y > 0
						2 end if`,
			expectedChildCount: 4, // TempScope(for if) + IfBodyScope + TempScope(for else-if) + ElseIfBodyScope
		},
		{
			name: "if with else-if literal bodies",
			input: `metrics
	test_metric
		inputs
			metric a
			metric b
		calculate
			if
				a
					42 else if b
						99 end if`,
			expectedChildCount: 4, // TempScope(for if) + IfBodyScope + TempScope(for else-if) + ElseIfBodyScope
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
			require.Equal(t, symboltable.MetricsScope, metricsScope.ScopeType)
			require.Len(t, metricsScope.Children, 1, "Expected one MetricScope child")
			metricScope := metricsScope.Children[0]
			require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

			// Check for IfClauseDefinition in MetricScope.Definitions
			ifDef := findDefByType(metricScope.Definitions, symboltable.IfClauseDefinition)
			require.NotNil(t, ifDef, "Expected IfClauseDefinition in MetricScope.Definitions")
			_ = requireClauseData(t, *ifDef)

			// Check for ElseIfClauseDefinition in MetricScope.Definitions
			elseIfDef := findDefByType(metricScope.Definitions, symboltable.ElseIfClauseDefinition)
			require.NotNil(t, elseIfDef, "Expected ElseIfClauseDefinition in MetricScope.Definitions")
			_ = requireClauseData(t, *elseIfDef)

			// Check that both IfBodyScope and ElseIfBodyScope exist in children
			require.Len(t, metricScope.Children, tc.expectedChildCount,
				"Expected %d children in MetricScope (TempScope(for if) + IfBodyScope + TempScope(for else-if) + ElseIfBodyScope)",
				tc.expectedChildCount)

			var ifBodyScope *symboltable.SymbolTable
			var elseIfBodyScope *symboltable.SymbolTable
			for _, child := range metricScope.Children {
				switch child.ScopeType {
				case symboltable.IfBodyScope:
					ifBodyScope = child
				case symboltable.ElseIfBodyScope:
					elseIfBodyScope = child
				}
			}
			require.NotNil(t, ifBodyScope, "Expected IfBodyScope in MetricScope.Children")
			require.NotNil(t, elseIfBodyScope, "Expected ElseIfBodyScope in MetricScope.Children")
			assert.Equal(t, symboltable.IfBodyScope, ifBodyScope.ScopeType)
			assert.Equal(t, symboltable.ElseIfBodyScope, elseIfBodyScope.ScopeType)
		})
	}
}

// TestElseIfBranchStoresClauseData verifies that after parsing an if/else-if expression,
// the parent scope contains an ElseIfClauseDefinition with ClauseData carrying
// the condition's postfix definitions.
func TestElseIfBranchStoresClauseData(t *testing.T) {
	t.Parallel()

	t.Run("else-if with comparison condition", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric x
			metric y
		calculate
			if
				x > 0
					1 else if y <= 5
						2 end if`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]

		// Verify ElseIfClauseDefinition exists with ClauseData
		elseIfDef := findDefByType(metricScope.Definitions, symboltable.ElseIfClauseDefinition)
		require.NotNil(t, elseIfDef, "Expected ElseIfClauseDefinition in MetricScope.Definitions")
		clauseData := requireClauseData(t, *elseIfDef)

		require.NotNil(t, clauseData.ConditionPostfix,
			"ConditionPostfix should not be nil")

		// Verify postfix contains the expected sequence for y <= 5:
		// Identifier (y), LiteralDefinition (5), LessThanEqOperatorDefinition
		require.Len(t, clauseData.ConditionPostfix, 3,
			"ConditionPostfix should have 3 elements (identifier, literal, operator)")

		assert.Equal(t, symboltable.Identifier, clauseData.ConditionPostfix[0].DefinitionType,
			"First postfix element should be Identifier (y)")
		assert.Equal(t, symboltable.LiteralDefinition, clauseData.ConditionPostfix[1].DefinitionType,
			"Second postfix element should be LiteralDefinition (5)")
		assert.Equal(t, symboltable.LessThanEqOperatorDefinition, clauseData.ConditionPostfix[2].DefinitionType,
			"Third postfix element should be LessThanEqOperatorDefinition (<=)")
	})
}

// TestElseIfBranchBodyScope verifies that ClauseData.BodyScope on an
// ElseIfClauseDefinition correctly references the ElseIfBodyScope child scope
// created during parsing. Follows the same pattern as TestIfBranchClauseDataBodyScope.
func TestElseIfBranchBodyScope(t *testing.T) {
	t.Parallel()

	t.Run("else-if with literal body", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric x
			metric y
		calculate
			if
				x > 0
					1 else if y > 0
						2 end if`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]
		require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

		// MetricScope should have 3 children: TempScope + IfBodyScope + ElseIfBodyScope
		require.Len(t, metricScope.Children, 4,
			"MetricScope should have 4 children (TempScope(for if) + IfBodyScope + TempScope(for else-if) + ElseIfBodyScope)")

		var expectedBodyScope *symboltable.SymbolTable
		for _, child := range metricScope.Children {
			if child.ScopeType == symboltable.ElseIfBodyScope {
				expectedBodyScope = child
				break
			}
		}
		require.NotNil(t, expectedBodyScope, "Expected ElseIfBodyScope in MetricScope.Children")
		require.Equal(t, symboltable.ElseIfBodyScope, expectedBodyScope.ScopeType)

		// Verify ElseIfClauseDefinition has ClauseData with BodyScope set
		elseIfDef := findDefByType(metricScope.Definitions, symboltable.ElseIfClauseDefinition)
		require.NotNil(t, elseIfDef, "Expected ElseIfClauseDefinition in MetricScope.Definitions")
		clauseData := requireClauseData(t, *elseIfDef)

		require.NotNil(t, clauseData.BodyScope, "ClauseData.BodyScope should not be nil")
		require.Equal(t, symboltable.ElseIfBodyScope, clauseData.BodyScope.ScopeType,
			"BodyScope should be ElseIfBodyScope")
		require.Same(t, expectedBodyScope, clauseData.BodyScope,
			"BodyScope should be the same pointer as MetricScope's child ElseIfBodyScope")

		// Body scope should contain the literal 2
		require.Len(t, clauseData.BodyScope.Definitions, 1,
			"Body scope should have 1 definition (literal 2)")
		assert.Equal(t, symboltable.LiteralDefinition, clauseData.BodyScope.Definitions[0].DefinitionType)
	})

	t.Run("else-if with arithmetic body", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric x
			metric a
			metric b
		calculate
			if
				x > 0
					1 else if x <= 0
						a + b end if`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]
		require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

		require.Len(t, metricScope.Children, 4,
			"MetricScope should have 4 children (TempScope(for if) + IfBodyScope + TempScope(for else-if) + ElseIfBodyScope)")

		var expectedBodyScope *symboltable.SymbolTable
		for _, child := range metricScope.Children {
			if child.ScopeType == symboltable.ElseIfBodyScope {
				expectedBodyScope = child
				break
			}
		}
		require.NotNil(t, expectedBodyScope, "Expected ElseIfBodyScope in MetricScope.Children")
		require.Equal(t, symboltable.ElseIfBodyScope, expectedBodyScope.ScopeType)

		elseIfDef := findDefByType(metricScope.Definitions, symboltable.ElseIfClauseDefinition)
		require.NotNil(t, elseIfDef, "Expected ElseIfClauseDefinition in MetricScope.Definitions")
		clauseData := requireClauseData(t, *elseIfDef)

		require.NotNil(t, clauseData.BodyScope, "ClauseData.BodyScope should not be nil")
		require.Equal(t, symboltable.ElseIfBodyScope, clauseData.BodyScope.ScopeType,
			"BodyScope should be ElseIfBodyScope")
		require.Same(t, expectedBodyScope, clauseData.BodyScope,
			"BodyScope should be the same pointer as MetricScope's child ElseIfBodyScope")

		// Body scope should contain postfix for a + b: a, b, PlusOperator
		require.Len(t, clauseData.BodyScope.Definitions, 3,
			"Body scope should have 3 definitions (a, b, +)")
		expectedPostfix := []symboltable.SymbolDefinition{
			{DefinitionType: symboltable.Identifier},
			{DefinitionType: symboltable.Identifier},
			{DefinitionType: symboltable.PlusOperatorDefinition},
		}
		for i, expected := range expectedPostfix {
			assertPostfixElement(t, expected, clauseData.BodyScope.Definitions[i], i)
		}
	})
}

// TestElseIfBranchConditionPostfix verifies that condition expressions in
// else-if branches produce correct postfix sequences in the ClauseData stored
// on the ElseIfClauseDefinition. Table-driven over various operators.
func TestElseIfBranchConditionPostfix(t *testing.T) {
	t.Parallel()

	// Helper to create expected SymbolDefinition
	makeDef := func(defType symboltable.DefinitionType, name string) symboltable.SymbolDefinition {
		var namePtr *string
		if name != "" {
			namePtr = &name
		}
		return symboltable.SymbolDefinition{
			DefinitionType: defType,
			Name:           namePtr,
		}
	}

	testCases := []struct {
		name            string
		input           string
		expectedPostfix []symboltable.SymbolDefinition
	}{
		{
			name: "greater than",
			input: `metrics
	test_metric
		inputs
			metric x
			metric y
		calculate
			if
				x > 10
					1 else if y > 0
						2 end if`,
			expectedPostfix: []symboltable.SymbolDefinition{
				makeDef(symboltable.Identifier, ""),
				makeDef(symboltable.LiteralDefinition, ""),
				makeDef(symboltable.GreaterThanOperatorDefinition, ""),
			},
		},
		{
			name: "less than",
			input: `metrics
	test_metric
		inputs
			metric x
			metric y
		calculate
			if
				x > 10
					1 else if y < 5
						2 end if`,
			expectedPostfix: []symboltable.SymbolDefinition{
				makeDef(symboltable.Identifier, ""),
				makeDef(symboltable.LiteralDefinition, ""),
				makeDef(symboltable.LessThanOperatorDefinition, ""),
			},
		},
		{
			name: "equality",
			input: `metrics
	test_metric
		inputs
			metric x
			metric y
		calculate
			if
				x > 10
					1 else if y == 0
						2 end if`,
			expectedPostfix: []symboltable.SymbolDefinition{
				makeDef(symboltable.Identifier, ""),
				makeDef(symboltable.LiteralDefinition, ""),
				makeDef(symboltable.EqOperatorDefinition, ""),
			},
		},
		{
			name: "not equal",
			input: `metrics
	test_metric
		inputs
			metric x
			metric y
		calculate
			if
				x > 10
					1 else if y != 0
						2 end if`,
			expectedPostfix: []symboltable.SymbolDefinition{
				makeDef(symboltable.Identifier, ""),
				makeDef(symboltable.LiteralDefinition, ""),
				makeDef(symboltable.NotEqOperatorDefinition, ""),
			},
		},
		{
			name: "greater than or equal",
			input: `metrics
	test_metric
		inputs
			metric x
			metric y
		calculate
			if
				x > 10
					1 else if y >= 0
						2 end if`,
			expectedPostfix: []symboltable.SymbolDefinition{
				makeDef(symboltable.Identifier, ""),
				makeDef(symboltable.LiteralDefinition, ""),
				makeDef(symboltable.GreaterThanEqOperatorDefinition, ""),
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
			metricsScope := doc.Children[0]
			metricScope := metricsScope.Children[0]

			// Get the ElseIfClauseDefinition
			elseIfDef := findDefByType(metricScope.Definitions, symboltable.ElseIfClauseDefinition)
			require.NotNil(t, elseIfDef, "Expected ElseIfClauseDefinition")
			clauseData := requireClauseData(t, *elseIfDef)

			// Verify postfix length
			require.Len(t, clauseData.ConditionPostfix, len(tc.expectedPostfix),
				"ConditionPostfix length mismatch")

			// Verify each postfix element
			for i, expected := range tc.expectedPostfix {
				assertPostfixElement(t, expected, clauseData.ConditionPostfix[i], i)
			}
		})
	}
}

// TestElseBranchCreatesChildScope verifies that parsing an if/else expression
// produces an ElseBodyScope child scope alongside the IfBodyScope. Unlike else-if,
// the else-branch has NO TempScope since there is no condition expression.
// Expected children: TempScope(for if) + IfBodyScope + ElseBodyScope = 3
func TestElseBranchCreatesChildScope(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		input              string
		expectedChildCount int
	}{
		{
			name: "if with else literal bodies",
			input: `metrics
	test_metric
		inputs
			metric x
		calculate
			if
				x > 0
					1 else
						2 end if`,
			expectedChildCount: 3, // TempScope(for if) + IfBodyScope + ElseBodyScope
		},
		{
			name: "if with else arithmetic body",
			input: `metrics
	test_metric
		inputs
			metric a
			metric b
		calculate
			if
				a > 0
					a else
						b end if`,
			expectedChildCount: 3, // TempScope(for if) + IfBodyScope + ElseBodyScope
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
			require.Equal(t, symboltable.MetricsScope, metricsScope.ScopeType)
			require.Len(t, metricsScope.Children, 1, "Expected one MetricScope child")
			metricScope := metricsScope.Children[0]
			require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

			// Check for IfClauseDefinition in MetricScope.Definitions
			ifDef := findDefByType(metricScope.Definitions, symboltable.IfClauseDefinition)
			require.NotNil(t, ifDef, "Expected IfClauseDefinition in MetricScope.Definitions")
			_ = requireClauseData(t, *ifDef)

			// Check for ElseClauseDefinition in MetricScope.Definitions
			elseDef := findDefByType(metricScope.Definitions, symboltable.ElseClauseDefinition)
			require.NotNil(t, elseDef, "Expected ElseClauseDefinition in MetricScope.Definitions")
			_ = requireClauseData(t, *elseDef)

			// Check that both IfBodyScope and ElseBodyScope exist in children
			require.Len(t, metricScope.Children, tc.expectedChildCount,
				"Expected %d children in MetricScope (TempScope(for if) + IfBodyScope + ElseBodyScope)",
				tc.expectedChildCount)

			var ifBodyScope *symboltable.SymbolTable
			var elseBodyScope *symboltable.SymbolTable
			for _, child := range metricScope.Children {
				switch child.ScopeType {
				case symboltable.IfBodyScope:
					ifBodyScope = child
				case symboltable.ElseBodyScope:
					elseBodyScope = child
				}
			}
			require.NotNil(t, ifBodyScope, "Expected IfBodyScope in MetricScope.Children")
			require.NotNil(t, elseBodyScope, "Expected ElseBodyScope in MetricScope.Children")
			assert.Equal(t, symboltable.IfBodyScope, ifBodyScope.ScopeType)
			assert.Equal(t, symboltable.ElseBodyScope, elseBodyScope.ScopeType)
		})
	}
}

// TestElseBranchStoresClauseData verifies that after parsing an if/else expression,
// the parent scope contains an ElseClauseDefinition with ClauseData carrying
// BodyScope but NO ConditionPostfix (else has no condition).
func TestElseBranchStoresClauseData(t *testing.T) {
	t.Parallel()

	t.Run("else with literal body", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric x
		calculate
			if
				x > 0
					1 else
						2 end if`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]

		// Verify ElseClauseDefinition exists with ClauseData
		elseDef := findDefByType(metricScope.Definitions, symboltable.ElseClauseDefinition)
		require.NotNil(t, elseDef, "Expected ElseClauseDefinition in MetricScope.Definitions")
		clauseData := requireClauseData(t, *elseDef)

		// ElseClauseDefinition should have NO ConditionPostfix
		require.Nil(t, clauseData.ConditionPostfix,
			"ConditionPostfix should be nil for else-branch (no condition)")

		// But BodyScope should be set
		require.NotNil(t, clauseData.BodyScope,
			"BodyScope should not be nil")
		require.Equal(t, symboltable.ElseBodyScope, clauseData.BodyScope.ScopeType,
			"BodyScope should be ElseBodyScope")
	})

	t.Run("else with arithmetic body", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric x
			metric a
			metric b
		calculate
			if
				x > 0
					x else
						a + b end if`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]

		elseDef := findDefByType(metricScope.Definitions, symboltable.ElseClauseDefinition)
		require.NotNil(t, elseDef, "Expected ElseClauseDefinition in MetricScope.Definitions")
		clauseData := requireClauseData(t, *elseDef)

		// No condition postfix
		require.Nil(t, clauseData.ConditionPostfix,
			"ConditionPostfix should be nil for else-branch")

		// BodyScope should contain postfix for a + b: a, b, PlusOperator
		require.NotNil(t, clauseData.BodyScope, "BodyScope should not be nil")
		require.Len(t, clauseData.BodyScope.Definitions, 3,
			"Body scope should have 3 definitions (a, b, +)")
		expectedPostfix := []symboltable.SymbolDefinition{
			{DefinitionType: symboltable.Identifier},
			{DefinitionType: symboltable.Identifier},
			{DefinitionType: symboltable.PlusOperatorDefinition},
		}
		for i, expected := range expectedPostfix {
			assertPostfixElement(t, expected, clauseData.BodyScope.Definitions[i], i)
		}
	})
}

// TestElseBranchBodyScope verifies that ClauseData.BodyScope on an
// ElseClauseDefinition correctly references the ElseBodyScope child scope
// created during parsing. Follows the same pattern as TestIfBranchClauseDataBodyScope.
func TestElseBranchBodyScope(t *testing.T) {
	t.Parallel()

	t.Run("else with literal body", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric x
		calculate
			if
				x > 0
					1 else
						42 end if`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]
		require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

		// MetricScope should have 3 children: TempScope + IfBodyScope + ElseBodyScope
		require.Len(t, metricScope.Children, 3,
			"MetricScope should have 3 children (TempScope(for if) + IfBodyScope + ElseBodyScope)")

		var expectedBodyScope *symboltable.SymbolTable
		for _, child := range metricScope.Children {
			if child.ScopeType == symboltable.ElseBodyScope {
				expectedBodyScope = child
				break
			}
		}
		require.NotNil(t, expectedBodyScope, "Expected ElseBodyScope in MetricScope.Children")
		require.Equal(t, symboltable.ElseBodyScope, expectedBodyScope.ScopeType)

		// Verify ElseClauseDefinition has ClauseData with BodyScope set
		elseDef := findDefByType(metricScope.Definitions, symboltable.ElseClauseDefinition)
		require.NotNil(t, elseDef, "Expected ElseClauseDefinition in MetricScope.Definitions")
		clauseData := requireClauseData(t, *elseDef)

		require.NotNil(t, clauseData.BodyScope, "ClauseData.BodyScope should not be nil")
		require.Equal(t, symboltable.ElseBodyScope, clauseData.BodyScope.ScopeType,
			"BodyScope should be ElseBodyScope")
		require.Same(t, expectedBodyScope, clauseData.BodyScope,
			"BodyScope should be the same pointer as MetricScope's child ElseBodyScope")

		// Body scope should contain the literal 42
		require.Len(t, clauseData.BodyScope.Definitions, 1,
			"Body scope should have 1 definition (literal 42)")
		assert.Equal(t, symboltable.LiteralDefinition, clauseData.BodyScope.Definitions[0].DefinitionType)

		// No condition postfix for else-branch
		require.Nil(t, clauseData.ConditionPostfix, "Else-branch should have no ConditionPostfix")
	})

	t.Run("else with arithmetic body", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric x
			metric a
			metric b
		calculate
			if
				x > 0
					x else
						a + b end if`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]
		require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

		require.Len(t, metricScope.Children, 3,
			"MetricScope should have 3 children (TempScope(for if) + IfBodyScope + ElseBodyScope)")

		var expectedBodyScope *symboltable.SymbolTable
		for _, child := range metricScope.Children {
			if child.ScopeType == symboltable.ElseBodyScope {
				expectedBodyScope = child
				break
			}
		}
		require.NotNil(t, expectedBodyScope, "Expected ElseBodyScope in MetricScope.Children")
		require.Equal(t, symboltable.ElseBodyScope, expectedBodyScope.ScopeType)

		elseDef := findDefByType(metricScope.Definitions, symboltable.ElseClauseDefinition)
		require.NotNil(t, elseDef, "Expected ElseClauseDefinition in MetricScope.Definitions")
		clauseData := requireClauseData(t, *elseDef)

		require.NotNil(t, clauseData.BodyScope, "ClauseData.BodyScope should not be nil")
		require.Equal(t, symboltable.ElseBodyScope, clauseData.BodyScope.ScopeType,
			"BodyScope should be ElseBodyScope")
		require.Same(t, expectedBodyScope, clauseData.BodyScope,
			"BodyScope should be the same pointer as MetricScope's child ElseBodyScope")

		// Body scope should contain postfix for a + b: a, b, PlusOperator
		require.Len(t, clauseData.BodyScope.Definitions, 3,
			"Body scope should have 3 definitions (a, b, +)")
		expectedPostfix := []symboltable.SymbolDefinition{
			{DefinitionType: symboltable.Identifier},
			{DefinitionType: symboltable.Identifier},
			{DefinitionType: symboltable.PlusOperatorDefinition},
		}
		for i, expected := range expectedPostfix {
			assertPostfixElement(t, expected, clauseData.BodyScope.Definitions[i], i)
		}

		// No condition postfix
		require.Nil(t, clauseData.ConditionPostfix, "Else-branch should have no ConditionPostfix")
	})
}

// TestElseBranchWithElseIf verifies that an if/else-if/else expression produces
// all three clause types with correct scopes and ClauseData.
// Expected children: TempScope(for if) + IfBodyScope + TempScope(for else-if) + ElseIfBodyScope + ElseBodyScope = 5
func TestElseBranchWithElseIf(t *testing.T) {
	t.Parallel()

	t.Run("if else-if else with literals", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric x
			metric y
		calculate
			if
				x > 10
					1 else if y > 0
						2 else
							3 end if`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]
		require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

		// Should have 5 children: 2 TempScopes + IfBody + ElseIfBody + ElseBody
		require.Len(t, metricScope.Children, 5,
			"MetricScope should have 5 children (TempScope+IfBodyScope+TempScope+ElseIfBodyScope+ElseBodyScope)")

		// Verify all three clause definitions exist
		ifDef := findDefByType(metricScope.Definitions, symboltable.IfClauseDefinition)
		require.NotNil(t, ifDef, "Expected IfClauseDefinition")
		ifClause := requireClauseData(t, *ifDef)
		require.NotNil(t, ifClause.ConditionPostfix, "If should have ConditionPostfix")
		require.NotNil(t, ifClause.BodyScope, "If should have BodyScope")

		elseIfDef := findDefByType(metricScope.Definitions, symboltable.ElseIfClauseDefinition)
		require.NotNil(t, elseIfDef, "Expected ElseIfClauseDefinition")
		elseIfClause := requireClauseData(t, *elseIfDef)
		require.NotNil(t, elseIfClause.ConditionPostfix, "Else-if should have ConditionPostfix")
		require.NotNil(t, elseIfClause.BodyScope, "Else-if should have BodyScope")

		elseDef := findDefByType(metricScope.Definitions, symboltable.ElseClauseDefinition)
		require.NotNil(t, elseDef, "Expected ElseClauseDefinition")
		elseClause := requireClauseData(t, *elseDef)
		require.Nil(t, elseClause.ConditionPostfix, "Else should NOT have ConditionPostfix")
		require.NotNil(t, elseClause.BodyScope, "Else should have BodyScope")

		// Verify all three body scope types exist in children
		var ifBodyScope, elseIfBodyScope, elseBodyScope *symboltable.SymbolTable
		for _, child := range metricScope.Children {
			switch child.ScopeType {
			case symboltable.IfBodyScope:
				ifBodyScope = child
			case symboltable.ElseIfBodyScope:
				elseIfBodyScope = child
			case symboltable.ElseBodyScope:
				elseBodyScope = child
			}
		}
		require.NotNil(t, ifBodyScope, "Expected IfBodyScope")
		require.NotNil(t, elseIfBodyScope, "Expected ElseIfBodyScope")
		require.NotNil(t, elseBodyScope, "Expected ElseBodyScope")

		// BodyScope pointers should match
		require.Same(t, ifBodyScope, ifClause.BodyScope, "IfClause.BodyScope should match child IfBodyScope")
		require.Same(t, elseIfBodyScope, elseIfClause.BodyScope, "ElseIfClause.BodyScope should match child ElseIfBodyScope")
		require.Same(t, elseBodyScope, elseClause.BodyScope, "ElseClause.BodyScope should match child ElseBodyScope")

		// Verify body scopes have the expected literals
		require.Len(t, ifBodyScope.Definitions, 1, "If body should have 1 definition (literal 1)")
		assert.Equal(t, symboltable.LiteralDefinition, ifBodyScope.Definitions[0].DefinitionType)

		require.Len(t, elseIfBodyScope.Definitions, 1, "Else-if body should have 1 definition (literal 2)")
		assert.Equal(t, symboltable.LiteralDefinition, elseIfBodyScope.Definitions[0].DefinitionType)

		require.Len(t, elseBodyScope.Definitions, 1, "Else body should have 1 definition (literal 3)")
		assert.Equal(t, symboltable.LiteralDefinition, elseBodyScope.Definitions[0].DefinitionType)
	})
}

// ================================================================================
// WHEN-CLAUSE AND DEFAULT-CLAUSE TESTS (Task 2f)
// ================================================================================
//
// CPL when-expression syntax (inline continuation, same as if/else-if pattern):
//
//	when
//	    condition =>
//	        body default =>
//	            default_body end when
//
// For multiple when-clauses, the next when-clause continues on the same line:
//
//	when
//	    cond1 =>
//	        body1 cond2 =>
//	            body2 default =>
//	                default_body end when
//
// Key grammar constraint: no INDENT between consecutive clauses or before
// END_KEYWORD WHEN_KEYWORD — these must appear on the same line.

// TestWhenClauseCreatesChildScope verifies that parsing a simple when expression
// with a single when-clause produces a WhenBodyScope child scope. The condition
// expression is parsed in the parent scope (MetricScope for calculate clauses),
// and the body expression is parsed in the child WhenBodyScope.
func TestWhenClauseCreatesChildScope(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		input              string
		expectedChildCount int
	}{
		{
			name: "single when with literal body",
			input: `metrics
	test_metric
		inputs
			metric x
		calculate
			when
				x > 0 =>
					1 default =>
						0 end when`,
			expectedChildCount: 3, // TempScope(for when) + WhenBodyScope + DefaultBodyScope
		},
		{
			name: "single when with arithmetic body",
			input: `metrics
	test_metric
		inputs
			metric a
			metric b
		calculate
			when
				a > b =>
					a + b default =>
						0 end when`,
			expectedChildCount: 3, // TempScope(for when) + WhenBodyScope + DefaultBodyScope
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
			require.Equal(t, symboltable.MetricsScope, metricsScope.ScopeType)
			require.Len(t, metricsScope.Children, 1, "Expected one MetricScope child")
			metricScope := metricsScope.Children[0]
			require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

			// Check for WhenClauseDefinition in MetricScope.Definitions
			whenDef := findDefByType(metricScope.Definitions, symboltable.WhenClauseDefinition)
			require.NotNil(t, whenDef, "Expected WhenClauseDefinition in MetricScope.Definitions")
			_ = requireClauseData(t, *whenDef)

			// Check for DefaultClauseDefinition in MetricScope.Definitions
			defaultDef := findDefByType(metricScope.Definitions, symboltable.DefaultClauseDefinition)
			require.NotNil(t, defaultDef, "Expected DefaultClauseDefinition in MetricScope.Definitions")
			_ = requireClauseData(t, *defaultDef)

			// Check that both WhenBodyScope and DefaultBodyScope exist in children
			require.Len(t, metricScope.Children, tc.expectedChildCount,
				"Expected %d children in MetricScope (TempScope + WhenBodyScope + DefaultBodyScope)",
				tc.expectedChildCount)

			var whenBodyScope *symboltable.SymbolTable
			var defaultBodyScope *symboltable.SymbolTable
			for _, child := range metricScope.Children {
				switch child.ScopeType {
				case symboltable.WhenBodyScope:
					whenBodyScope = child
				case symboltable.DefaultBodyScope:
					defaultBodyScope = child
				}
			}
			require.NotNil(t, whenBodyScope, "Expected WhenBodyScope in MetricScope.Children")
			require.NotNil(t, defaultBodyScope, "Expected DefaultBodyScope in MetricScope.Children")
			assert.Equal(t, symboltable.WhenBodyScope, whenBodyScope.ScopeType)
			assert.Equal(t, symboltable.DefaultBodyScope, defaultBodyScope.ScopeType)
		})
	}
}

// TestWhenClauseConditionPostfix verifies that condition expressions in when-clauses
// produce correct postfix sequences in the ClauseData stored on the
// WhenClauseDefinition.
func TestWhenClauseConditionPostfix(t *testing.T) {
	t.Parallel()

	makeDef := func(defType symboltable.DefinitionType, name string) symboltable.SymbolDefinition {
		var namePtr *string
		if name != "" {
			namePtr = &name
		}
		return symboltable.SymbolDefinition{
			DefinitionType: defType,
			Name:           namePtr,
		}
	}

	testCases := []struct {
		name            string
		input           string
		expectedPostfix []symboltable.SymbolDefinition
	}{
		{
			name: "greater than comparison",
			input: `metrics
	test_metric
		inputs
			metric x
		calculate
			when
				x > 0 =>
					1 default =>
						0 end when`,
			expectedPostfix: []symboltable.SymbolDefinition{
				makeDef(symboltable.Identifier, ""),
				makeDef(symboltable.LiteralDefinition, ""),
				makeDef(symboltable.GreaterThanOperatorDefinition, ""),
			},
		},
		{
			name: "equality comparison",
			input: `metrics
	test_metric
		inputs
			metric x
			metric y
		calculate
			when
				x == y =>
					1 default =>
						0 end when`,
			expectedPostfix: []symboltable.SymbolDefinition{
				makeDef(symboltable.Identifier, ""),
				makeDef(symboltable.Identifier, ""),
				makeDef(symboltable.EqOperatorDefinition, ""),
			},
		},
		{
			name: "simple identifier condition",
			input: `metrics
	test_metric
		inputs
			metric x
		calculate
			when
				x =>
					42 default =>
						0 end when`,
			expectedPostfix: []symboltable.SymbolDefinition{
				makeDef(symboltable.Identifier, ""),
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
			metricsScope := doc.Children[0]
			metricScope := metricsScope.Children[0]

			// Get the WhenClauseDefinition
			whenDef := findDefByType(metricScope.Definitions, symboltable.WhenClauseDefinition)
			require.NotNil(t, whenDef, "Expected WhenClauseDefinition")
			clauseData := requireClauseData(t, *whenDef)

			// Verify postfix length
			require.Len(t, clauseData.ConditionPostfix, len(tc.expectedPostfix),
				"ConditionPostfix length mismatch")

			// Verify each postfix element
			for i, expected := range tc.expectedPostfix {
				assertPostfixElement(t, expected, clauseData.ConditionPostfix[i], i)
			}
		})
	}
}

// TestDefaultClauseStoresClauseData verifies that after parsing a when/default expression,
// the parent scope contains a DefaultClauseDefinition with ClauseData carrying
// BodyScope but NO ConditionPostfix (default has no condition).
func TestDefaultClauseStoresClauseData(t *testing.T) {
	t.Parallel()

	t.Run("default with literal body", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric x
		calculate
			when
				x > 0 =>
					1 default =>
						0 end when`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]

		// Verify DefaultClauseDefinition exists with ClauseData
		defaultDef := findDefByType(metricScope.Definitions, symboltable.DefaultClauseDefinition)
		require.NotNil(t, defaultDef, "Expected DefaultClauseDefinition in MetricScope.Definitions")
		clauseData := requireClauseData(t, *defaultDef)

		// DefaultClauseDefinition should have NO ConditionPostfix
		require.Nil(t, clauseData.ConditionPostfix,
			"ConditionPostfix should be nil for default-clause (no condition)")

		// But BodyScope should be set
		require.NotNil(t, clauseData.BodyScope,
			"BodyScope should not be nil")
		require.Equal(t, symboltable.DefaultBodyScope, clauseData.BodyScope.ScopeType,
			"BodyScope should be DefaultBodyScope")
	})

	t.Run("default with arithmetic body", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric a
			metric b
		calculate
			when
				a > 0 =>
					a default =>
						a + b end when`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]

		defaultDef := findDefByType(metricScope.Definitions, symboltable.DefaultClauseDefinition)
		require.NotNil(t, defaultDef, "Expected DefaultClauseDefinition in MetricScope.Definitions")
		clauseData := requireClauseData(t, *defaultDef)

		// No condition postfix
		require.Nil(t, clauseData.ConditionPostfix,
			"ConditionPostfix should be nil for default-clause")

		// BodyScope should contain postfix for a + b: a, b, PlusOperator
		require.NotNil(t, clauseData.BodyScope, "BodyScope should not be nil")
		require.Len(t, clauseData.BodyScope.Definitions, 3,
			"Body scope should have 3 definitions (a, b, +)")
		expectedPostfix := []symboltable.SymbolDefinition{
			{DefinitionType: symboltable.Identifier},
			{DefinitionType: symboltable.Identifier},
			{DefinitionType: symboltable.PlusOperatorDefinition},
		}
		for i, expected := range expectedPostfix {
			assertPostfixElement(t, expected, clauseData.BodyScope.Definitions[i], i)
		}
	})
}

// TestWhenClauseBodyScope verifies that ClauseData.BodyScope on a
// WhenClauseDefinition correctly references the WhenBodyScope child scope
// created during parsing.
func TestWhenClauseBodyScope(t *testing.T) {
	t.Parallel()

	t.Run("when with literal body", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric x
		calculate
			when
				x > 0 =>
					42 default =>
						0 end when`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]
		require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

		// MetricScope should have 3 children: TempScope + WhenBodyScope + DefaultBodyScope
		require.Len(t, metricScope.Children, 3,
			"MetricScope should have 3 children (TempScope + WhenBodyScope + DefaultBodyScope)")

		var expectedBodyScope *symboltable.SymbolTable
		for _, child := range metricScope.Children {
			if child.ScopeType == symboltable.WhenBodyScope {
				expectedBodyScope = child
				break
			}
		}
		require.NotNil(t, expectedBodyScope, "Expected WhenBodyScope in MetricScope.Children")
		require.Equal(t, symboltable.WhenBodyScope, expectedBodyScope.ScopeType)

		// Verify WhenClauseDefinition has ClauseData with BodyScope set
		whenDef := findDefByType(metricScope.Definitions, symboltable.WhenClauseDefinition)
		require.NotNil(t, whenDef, "Expected WhenClauseDefinition in MetricScope.Definitions")
		clauseData := requireClauseData(t, *whenDef)

		require.NotNil(t, clauseData.BodyScope, "ClauseData.BodyScope should not be nil")
		require.Equal(t, symboltable.WhenBodyScope, clauseData.BodyScope.ScopeType,
			"BodyScope should be WhenBodyScope")
		require.Same(t, expectedBodyScope, clauseData.BodyScope,
			"BodyScope should be the same pointer as MetricScope's child WhenBodyScope")

		// Body scope should contain the literal 42
		require.Len(t, clauseData.BodyScope.Definitions, 1,
			"Body scope should have 1 definition (literal 42)")
		assert.Equal(t, symboltable.LiteralDefinition, clauseData.BodyScope.Definitions[0].DefinitionType)

		// Condition postfix should be present
		require.NotNil(t, clauseData.ConditionPostfix, "When-clause should have ConditionPostfix")
	})

	t.Run("when with arithmetic body", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric a
			metric b
		calculate
			when
				a > 0 =>
					a + b default =>
						0 end when`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]
		require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

		require.Len(t, metricScope.Children, 3,
			"MetricScope should have 3 children (TempScope + WhenBodyScope + DefaultBodyScope)")

		var expectedBodyScope *symboltable.SymbolTable
		for _, child := range metricScope.Children {
			if child.ScopeType == symboltable.WhenBodyScope {
				expectedBodyScope = child
				break
			}
		}
		require.NotNil(t, expectedBodyScope, "Expected WhenBodyScope in MetricScope.Children")
		require.Equal(t, symboltable.WhenBodyScope, expectedBodyScope.ScopeType)

		whenDef := findDefByType(metricScope.Definitions, symboltable.WhenClauseDefinition)
		require.NotNil(t, whenDef, "Expected WhenClauseDefinition in MetricScope.Definitions")
		clauseData := requireClauseData(t, *whenDef)

		require.NotNil(t, clauseData.BodyScope, "ClauseData.BodyScope should not be nil")
		require.Equal(t, symboltable.WhenBodyScope, clauseData.BodyScope.ScopeType,
			"BodyScope should be WhenBodyScope")
		require.Same(t, expectedBodyScope, clauseData.BodyScope,
			"BodyScope should be the same pointer as MetricScope's child WhenBodyScope")

		// Body scope should contain postfix for a + b: a, b, PlusOperator
		require.Len(t, clauseData.BodyScope.Definitions, 3,
			"Body scope should have 3 definitions (a, b, +)")
		expectedPostfix := []symboltable.SymbolDefinition{
			{DefinitionType: symboltable.Identifier},
			{DefinitionType: symboltable.Identifier},
			{DefinitionType: symboltable.PlusOperatorDefinition},
		}
		for i, expected := range expectedPostfix {
			assertPostfixElement(t, expected, clauseData.BodyScope.Definitions[i], i)
		}
	})
}

// TestDefaultClauseBodyScope verifies that ClauseData.BodyScope on a
// DefaultClauseDefinition correctly references the DefaultBodyScope child scope
// created during parsing.
func TestDefaultClauseBodyScope(t *testing.T) {
	t.Parallel()

	t.Run("default with literal body", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric x
		calculate
			when
				x > 0 =>
					1 default =>
						42 end when`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]
		require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

		require.Len(t, metricScope.Children, 3,
			"MetricScope should have 3 children (TempScope + WhenBodyScope + DefaultBodyScope)")

		var expectedBodyScope *symboltable.SymbolTable
		for _, child := range metricScope.Children {
			if child.ScopeType == symboltable.DefaultBodyScope {
				expectedBodyScope = child
				break
			}
		}
		require.NotNil(t, expectedBodyScope, "Expected DefaultBodyScope in MetricScope.Children")
		require.Equal(t, symboltable.DefaultBodyScope, expectedBodyScope.ScopeType)

		// Verify DefaultClauseDefinition has ClauseData with BodyScope set
		defaultDef := findDefByType(metricScope.Definitions, symboltable.DefaultClauseDefinition)
		require.NotNil(t, defaultDef, "Expected DefaultClauseDefinition in MetricScope.Definitions")
		clauseData := requireClauseData(t, *defaultDef)

		require.NotNil(t, clauseData.BodyScope, "ClauseData.BodyScope should not be nil")
		require.Equal(t, symboltable.DefaultBodyScope, clauseData.BodyScope.ScopeType,
			"BodyScope should be DefaultBodyScope")
		require.Same(t, expectedBodyScope, clauseData.BodyScope,
			"BodyScope should be the same pointer as MetricScope's child DefaultBodyScope")

		// Body scope should contain the literal 42
		require.Len(t, clauseData.BodyScope.Definitions, 1,
			"Body scope should have 1 definition (literal 42)")
		assert.Equal(t, symboltable.LiteralDefinition, clauseData.BodyScope.Definitions[0].DefinitionType)

		// No condition postfix for default-clause
		require.Nil(t, clauseData.ConditionPostfix, "Default-clause should have no ConditionPostfix")
	})

	t.Run("default with arithmetic body", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric a
			metric b
		calculate
			when
				a > 0 =>
					a default =>
						a + b end when`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]
		require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

		require.Len(t, metricScope.Children, 3,
			"MetricScope should have 3 children (TempScope + WhenBodyScope + DefaultBodyScope)")

		var expectedBodyScope *symboltable.SymbolTable
		for _, child := range metricScope.Children {
			if child.ScopeType == symboltable.DefaultBodyScope {
				expectedBodyScope = child
				break
			}
		}
		require.NotNil(t, expectedBodyScope, "Expected DefaultBodyScope in MetricScope.Children")
		require.Equal(t, symboltable.DefaultBodyScope, expectedBodyScope.ScopeType)

		defaultDef := findDefByType(metricScope.Definitions, symboltable.DefaultClauseDefinition)
		require.NotNil(t, defaultDef, "Expected DefaultClauseDefinition in MetricScope.Definitions")
		clauseData := requireClauseData(t, *defaultDef)

		require.NotNil(t, clauseData.BodyScope, "ClauseData.BodyScope should not be nil")
		require.Equal(t, symboltable.DefaultBodyScope, clauseData.BodyScope.ScopeType,
			"BodyScope should be DefaultBodyScope")
		require.Same(t, expectedBodyScope, clauseData.BodyScope,
			"BodyScope should be the same pointer as MetricScope's child DefaultBodyScope")

		// Body scope should contain postfix for a + b: a, b, PlusOperator
		require.Len(t, clauseData.BodyScope.Definitions, 3,
			"Body scope should have 3 definitions (a, b, +)")
		expectedPostfix := []symboltable.SymbolDefinition{
			{DefinitionType: symboltable.Identifier},
			{DefinitionType: symboltable.Identifier},
			{DefinitionType: symboltable.PlusOperatorDefinition},
		}
		for i, expected := range expectedPostfix {
			assertPostfixElement(t, expected, clauseData.BodyScope.Definitions[i], i)
		}

		// No condition postfix
		require.Nil(t, clauseData.ConditionPostfix, "Default-clause should have no ConditionPostfix")
	})
}

// TestWhenWithMultipleWhenClauses verifies that multiple when-clauses produce
// correct definitions with correct scopes and ClauseData.
// Uses inline continuation: body1 cond2 => body2 default => default_body end when
// Expected children: TempScope(for when1) + WhenBodyScope1 + TempScope(for when2) + WhenBodyScope2 + DefaultBodyScope = 5
func TestWhenWithMultipleWhenClauses(t *testing.T) {
	t.Parallel()

	t.Run("two when clauses with default", func(t *testing.T) {
		t.Parallel()

		input := `metrics
	test_metric
		inputs
			metric x
		calculate
			when
				x > 10 =>
					1 x > 0 =>
						2 default =>
							0 end when`

		doc, err := parser.Parse(input)
		require.NoError(t, err, "Expected no error")
		require.NotNil(t, doc)

		metricScope := doc.Children[0].Children[0]
		require.Equal(t, symboltable.MetricScope, metricScope.ScopeType)

		// Should have 5 children: 2 TempScopes + 2 WhenBodyScopes + DefaultBodyScope
		require.Len(t, metricScope.Children, 5,
			"MetricScope should have 5 children (TempScope+WhenBodyScope+TempScope+WhenBodyScope+DefaultBodyScope)")

		// Count WhenClauseDefinitions in MetricScope.Definitions
		whenDefCount := countDefsByType(metricScope.Definitions, symboltable.WhenClauseDefinition)
		assert.Equal(t, 2, whenDefCount, "Expected 2 WhenClauseDefinitions")

		// Verify DefaultClauseDefinition exists
		defaultDef := findDefByType(metricScope.Definitions, symboltable.DefaultClauseDefinition)
		require.NotNil(t, defaultDef, "Expected DefaultClauseDefinition")
		defaultClause := requireClauseData(t, *defaultDef)
		require.Nil(t, defaultClause.ConditionPostfix, "Default should have no ConditionPostfix")
		require.NotNil(t, defaultClause.BodyScope, "Default should have BodyScope")

		// Find all WhenBodyScopes and DefaultBodyScope
		var whenBodyScopes []*symboltable.SymbolTable
		var defaultBodyScope *symboltable.SymbolTable
		for _, child := range metricScope.Children {
			switch child.ScopeType {
			case symboltable.WhenBodyScope:
				whenBodyScopes = append(whenBodyScopes, child)
			case symboltable.DefaultBodyScope:
				defaultBodyScope = child
			}
		}
		require.Len(t, whenBodyScopes, 2, "Expected 2 WhenBodyScopes")
		require.NotNil(t, defaultBodyScope, "Expected DefaultBodyScope")

		// Verify each WhenBodyScope has the expected literal
		assert.Len(t, whenBodyScopes[0].Definitions, 1, "First when body should have 1 definition")
		assert.Equal(t, symboltable.LiteralDefinition, whenBodyScopes[0].Definitions[0].DefinitionType)

		assert.Len(t, whenBodyScopes[1].Definitions, 1, "Second when body should have 1 definition")
		assert.Equal(t, symboltable.LiteralDefinition, whenBodyScopes[1].Definitions[0].DefinitionType)

		// Default body should have literal 0
		assert.Len(t, defaultBodyScope.Definitions, 1, "Default body should have 1 definition")
		assert.Equal(t, symboltable.LiteralDefinition, defaultBodyScope.Definitions[0].DefinitionType)
	})
}

// countDefsByType counts the number of SymbolDefinitions of the given type
// in the definitions slice.
func countDefsByType(defs []symboltable.SymbolDefinition, defType symboltable.DefinitionType) int {
	count := 0
	for _, def := range defs {
		if def.DefinitionType == defType {
			count++
		}
	}
	return count
}

// ================================================================================
// BOOLEAN AND NIL LITERAL PARSING TESTS
// ================================================================================
//
// These tests verify that True, False, and Nil literal keywords are correctly
// recognized by the lexer and mapped to the appropriate DefinitionType
// (TrueLiteralDefinition, FalseLiteralDefinition, NilLiteralDefinition) in the
// condition postfix. This closes the test gap that allowed the lexer/listener
// case-mismatch bug (Story187 review) to go unnoticed.

// TestBooleanLiteralConditionPostfix verifies that True and False literals
// used as conditions in if/else-if branches produce the correct postfix
// definition types (TrueLiteralDefinition, FalseLiteralDefinition).
func TestBooleanLiteralConditionPostfix(t *testing.T) {
	t.Parallel()

	makeDef := func(defType symboltable.DefinitionType) symboltable.SymbolDefinition {
		return symboltable.SymbolDefinition{DefinitionType: defType}
	}

	testCases := []struct {
		name            string
		input           string
		expectedPostfix []symboltable.SymbolDefinition
	}{
		{
			name: "True literal as if condition",
			input: `metrics
	test_metric
		inputs
			metric x
		calculate
			if
				True
					42 end if`,
			expectedPostfix: []symboltable.SymbolDefinition{
				makeDef(symboltable.TrueLiteralDefinition),
			},
		},
		{
			name: "False literal as else-if condition",
			input: `metrics
	test_metric
		inputs
			metric x
		calculate
			if
				True
					1 else if False
						2 end if`,
			expectedPostfix: []symboltable.SymbolDefinition{
				makeDef(symboltable.FalseLiteralDefinition),
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
			metricScope := doc.Children[0].Children[0]

			var clauseDef *symboltable.SymbolDefinition
			if tc.name == "True literal as if condition" {
				clauseDef = findDefByType(metricScope.Definitions, symboltable.IfClauseDefinition)
			} else {
				clauseDef = findDefByType(metricScope.Definitions, symboltable.ElseIfClauseDefinition)
			}
			require.NotNil(t, clauseDef, "Expected clause definition")
			clauseData := requireClauseData(t, *clauseDef)

			require.Len(t, clauseData.ConditionPostfix, len(tc.expectedPostfix),
				"ConditionPostfix length mismatch")
			for i, expected := range tc.expectedPostfix {
				assert.Equal(t, expected.DefinitionType, clauseData.ConditionPostfix[i].DefinitionType,
					"Postfix element %d: DefinitionType mismatch", i)
			}
		})
	}
}

// TestNilLiteralConditionPostfix verifies that the Nil literal used as a
// condition produces a NilLiteralDefinition in the postfix.
func TestNilLiteralConditionPostfix(t *testing.T) {
	t.Parallel()

	input := `metrics
	test_metric
		inputs
			metric x
		calculate
			when
				Nil =>
					0 default =>
						42 end when`

	doc, err := parser.Parse(input)
	require.NoError(t, err, "Expected no error")
	require.NotNil(t, doc)

	metricScope := doc.Children[0].Children[0]
	whenDef := findDefByType(metricScope.Definitions, symboltable.WhenClauseDefinition)
	require.NotNil(t, whenDef, "Expected WhenClauseDefinition")

	clauseData := requireClauseData(t, *whenDef)
	require.Len(t, clauseData.ConditionPostfix, 1, "ConditionPostfix should have 1 element")
	assert.Equal(t, symboltable.NilLiteralDefinition, clauseData.ConditionPostfix[0].DefinitionType,
		"Condition should be NilLiteralDefinition")
}

// _symbolTableTypes is a compile-time check that the required symbol table types exist.
// If these types are renamed or removed, this file will fail to compile, alerting
// the developer to update the conditional parser tests.
var _ = []symboltable.ScopeType{
	symboltable.IfBodyScope,
	symboltable.ElseIfBodyScope,
	symboltable.ElseBodyScope,
	symboltable.WhenBodyScope,
	symboltable.DefaultBodyScope,
}

var _ = []symboltable.DefinitionType{
	symboltable.IfClauseDefinition,
	symboltable.ElseIfClauseDefinition,
	symboltable.ElseClauseDefinition,
	symboltable.WhenClauseDefinition,
	symboltable.DefaultClauseDefinition,
	symboltable.EqOperatorDefinition,
	symboltable.NotEqOperatorDefinition,
	symboltable.LessThanOperatorDefinition,
	symboltable.GreaterThanOperatorDefinition,
	symboltable.LessThanEqOperatorDefinition,
	symboltable.GreaterThanEqOperatorDefinition,
	symboltable.AndOperatorDefinition,
	symboltable.OrOperatorDefinition,
	symboltable.NotOperatorDefinition,
}
