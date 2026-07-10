// Package parsing_test contains tests for the typed postfix expression elements.
package parsing_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parsing"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Interface Satisfaction Tests
// ============================================================================

// TestPostfixLiteral_Interface verifies PostfixLiteral satisfies PostfixElement.
func TestPostfixLiteral_Interface(t *testing.T) {
	t.Parallel()

	var _ parsing.PostfixElement = parsing.PostfixLiteral{}
}

// TestPostfixIdentifier_Interface verifies PostfixIdentifier satisfies PostfixElement.
func TestPostfixIdentifier_Interface(t *testing.T) {
	t.Parallel()

	var _ parsing.PostfixElement = parsing.PostfixIdentifier{}
}

// TestPostfixOperator_Interface verifies PostfixOperator satisfies PostfixElement.
func TestPostfixOperator_Interface(t *testing.T) {
	t.Parallel()

	var _ parsing.PostfixElement = parsing.PostfixOperator{}
}

// ============================================================================
// Construction and Field Access Tests
// ============================================================================

// TestPostfixLiteral_Construction verifies construction and field access.
func TestPostfixLiteral_Construction(t *testing.T) {
	t.Parallel()

	lit := parsing.PostfixLiteral{Value: 42.5}
	assert.Equal(t, float64(42.5), lit.Value)

	// Zero value
	zero := parsing.PostfixLiteral{}
	assert.Equal(t, float64(0), zero.Value)
}

// TestPostfixIdentifier_Construction verifies construction and field access.
func TestPostfixIdentifier_Construction(t *testing.T) {
	t.Parallel()

	id := parsing.PostfixIdentifier{Name: "demand_forecast"}
	assert.Equal(t, "demand_forecast", id.Name)

	// Zero value
	zero := parsing.PostfixIdentifier{}
	assert.Equal(t, "", zero.Name)
}

// TestPostfixOperator_Construction verifies construction and field access.
func TestPostfixOperator_Construction(t *testing.T) {
	t.Parallel()

	op := parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition}
	assert.Equal(t, symboltable.PlusOperatorDefinition, op.DefType)

	op2 := parsing.PostfixOperator{DefType: symboltable.TrueLiteralDefinition}
	assert.Equal(t, symboltable.TrueLiteralDefinition, op2.DefType)

	// Zero value
	zero := parsing.PostfixOperator{}
	assert.Equal(t, symboltable.DefinitionType(0), zero.DefType)
}

// ============================================================================
// Sample Postfix Expression Construction Test
// ============================================================================

// TestPostfixExpression_Construction verifies that a mix of all three
// element types can be assembled into a []PostfixElement slice.
func TestPostfixExpression_Construction(t *testing.T) {
	t.Parallel()

	// Postfix for: demand_forecast * Price + 5
	// Which is: demand_forecast Price * 5 +
	postfix := []parsing.PostfixElement{
		parsing.PostfixIdentifier{Name: "demand_forecast"},
		parsing.PostfixIdentifier{Name: "Price"},
		parsing.PostfixOperator{DefType: symboltable.MultiplyOperatorDefinition},
		parsing.PostfixLiteral{Value: 5},
		parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition},
	}

	assert.Len(t, postfix, 5)

	// Verify element 0 is identifier
	id0, ok := postfix[0].(parsing.PostfixIdentifier)
	assert.True(t, ok)
	assert.Equal(t, "demand_forecast", id0.Name)

	// Verify element 1 is identifier
	id1, ok := postfix[1].(parsing.PostfixIdentifier)
	assert.True(t, ok)
	assert.Equal(t, "Price", id1.Name)

	// Verify element 2 is multiply operator
	op2, ok := postfix[2].(parsing.PostfixOperator)
	assert.True(t, ok)
	assert.Equal(t, symboltable.MultiplyOperatorDefinition, op2.DefType)

	// Verify element 3 is literal
	lit3, ok := postfix[3].(parsing.PostfixLiteral)
	assert.True(t, ok)
	assert.Equal(t, float64(5), lit3.Value)

	// Verify element 4 is plus operator
	op4, ok := postfix[4].(parsing.PostfixOperator)
	assert.True(t, ok)
	assert.Equal(t, symboltable.PlusOperatorDefinition, op4.DefType)
}
