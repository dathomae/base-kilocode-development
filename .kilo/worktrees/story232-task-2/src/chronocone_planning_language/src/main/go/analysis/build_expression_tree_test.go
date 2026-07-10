// Package analysis_test contains unit tests for the CPL expression tree builder.
//
// These tests follow the Behavioral Test-Driven Development (BTDD) approach,
// where tests are written before the implementation exists.
//
// The tests cover:
//   - Simple unit reference resolution
//   - Binary operators (multiply, divide, power, plus, minus)
//   - Parentheses/precedence in postfix
//   - Nested operators
//   - Number literals
//   - Error cases (malformed, undefined unit, insufficient operands, too many operands)
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Test Helpers
// ============================================================================

func makeUnitRef(name string) symboltable.SymbolDefinition {
	return symboltable.SymbolDefinition{
		DefinitionType: symboltable.UnitReferenceDefinition,
		Data:           parser.ParserUnitReference{Name: name},
	}
}

func makeOperator(defType symboltable.DefinitionType, opType parser.OperatorType) symboltable.SymbolDefinition {
	return symboltable.SymbolDefinition{
		DefinitionType: defType,
		Data:           parser.ParserOperator{Operator: opType},
	}
}

func makeNumber(value float64) symboltable.SymbolDefinition {
	return symboltable.SymbolDefinition{
		DefinitionType: symboltable.NumberDefinition,
		Data:           symboltable.FloatData{Value: value},
	}
}

// ============================================================================
// BuildExpressionTree Tests
// ============================================================================

// TestBuildExpressionTree_SimpleUnitReference verifies that a single unit
// reference resolves to a UnitReference expression with UnitGroup set.
func TestBuildExpressionTree_SimpleUnitReference(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"Length": {Name: "Length", ConceptName: "DistanceUnits"},
	}
	defs := []symboltable.SymbolDefinition{
		makeUnitRef("Length"),
	}
	expr, err := analysis.BuildExpressionTree(defs, units, nil)
	require.NoError(t, err, "unexpected error")

	ref, ok := expr.(*analysis.UnitReference)
	require.True(t, ok, "expected UnitReference, got %T", expr)
	assert.Equal(t, "Length", ref.Name, "expected Name=Length")
	require.NotNil(t, ref.UnitGroup, "expected UnitGroup to be set")
	assert.Equal(t, "Length", ref.UnitGroup.Name, "expected UnitGroup.Name=Length")
	assert.Equal(t, "DistanceUnits", ref.UnitGroup.ConceptName, "expected UnitGroup.ConceptName=DistanceUnits")
}

// TestBuildExpressionTree_Multiplication verifies multiplication of two units.
func TestBuildExpressionTree_Multiplication(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"a": {Name: "a", ConceptName: "A"},
		"b": {Name: "b", ConceptName: "B"},
	}
	defs := []symboltable.SymbolDefinition{
		makeUnitRef("a"),
		makeUnitRef("b"),
		makeOperator(symboltable.MultiplyOperatorDefinition, parser.OpMultiply),
	}
	expr, err := analysis.BuildExpressionTree(defs, units, nil)
	require.NoError(t, err)

	node, ok := expr.(*analysis.MultiplicationNode)
	require.True(t, ok, "expected MultiplicationNode, got %T", expr)

	left, ok := node.Left.(*analysis.UnitReference)
	require.True(t, ok, "expected left to be UnitReference, got %T", node.Left)
	assert.Equal(t, "a", left.Name)

	right, ok := node.Right.(*analysis.UnitReference)
	require.True(t, ok, "expected right to be UnitReference, got %T", node.Right)
	assert.Equal(t, "b", right.Name)
}

// TestBuildExpressionTree_Division verifies division of two units.
func TestBuildExpressionTree_Division(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"a": {Name: "a", ConceptName: "A"},
		"b": {Name: "b", ConceptName: "B"},
	}
	defs := []symboltable.SymbolDefinition{
		makeUnitRef("a"),
		makeUnitRef("b"),
		makeOperator(symboltable.DivideOperatorDefinition, parser.OpDivide),
	}
	expr, err := analysis.BuildExpressionTree(defs, units, nil)
	require.NoError(t, err)

	node, ok := expr.(*analysis.DivisionNode)
	require.True(t, ok, "expected DivisionNode, got %T", expr)

	left, ok := node.Left.(*analysis.UnitReference)
	require.True(t, ok, "expected left to be UnitReference, got %T", node.Left)
	assert.Equal(t, "a", left.Name)

	right, ok := node.Right.(*analysis.UnitReference)
	require.True(t, ok, "expected right to be UnitReference, got %T", node.Right)
	assert.Equal(t, "b", right.Name)
}

// TestBuildExpressionTree_Power verifies exponentiation with a numeric literal.
func TestBuildExpressionTree_Power(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"a": {Name: "a", ConceptName: "A"},
	}
	defs := []symboltable.SymbolDefinition{
		makeUnitRef("a"),
		makeNumber(2),
		makeOperator(symboltable.PowerOperatorDefinition, parser.OpPower),
	}
	expr, err := analysis.BuildExpressionTree(defs, units, nil)
	require.NoError(t, err)

	node, ok := expr.(*analysis.PowerNode)
	require.True(t, ok, "expected PowerNode, got %T", expr)

	base, ok := node.Base.(*analysis.UnitReference)
	require.True(t, ok, "expected base to be UnitReference, got %T", node.Base)
	assert.Equal(t, "a", base.Name)
	assert.Equal(t, 2, node.Exponent, "expected exponent=2")
}

// TestBuildExpressionTree_Addition verifies addition of two units.
func TestBuildExpressionTree_Addition(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"a": {Name: "a", ConceptName: "A"},
		"b": {Name: "b", ConceptName: "B"},
	}
	defs := []symboltable.SymbolDefinition{
		makeUnitRef("a"),
		makeUnitRef("b"),
		makeOperator(symboltable.PlusOperatorDefinition, parser.OpPlus),
	}
	expr, err := analysis.BuildExpressionTree(defs, units, nil)
	require.NoError(t, err)

	node, ok := expr.(*analysis.AdditionNode)
	require.True(t, ok, "expected AdditionNode, got %T", expr)

	left, ok := node.Left.(*analysis.UnitReference)
	require.True(t, ok, "expected left to be UnitReference, got %T", node.Left)
	assert.Equal(t, "a", left.Name)

	right, ok := node.Right.(*analysis.UnitReference)
	require.True(t, ok, "expected right to be UnitReference, got %T", node.Right)
	assert.Equal(t, "b", right.Name)
}

// TestBuildExpressionTree_Subtraction verifies subtraction of two units.
func TestBuildExpressionTree_Subtraction(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"a": {Name: "a", ConceptName: "A"},
		"b": {Name: "b", ConceptName: "B"},
	}
	defs := []symboltable.SymbolDefinition{
		makeUnitRef("a"),
		makeUnitRef("b"),
		makeOperator(symboltable.MinusOperatorDefinition, parser.OpMinus),
	}
	expr, err := analysis.BuildExpressionTree(defs, units, nil)
	require.NoError(t, err)

	node, ok := expr.(*analysis.SubtractionNode)
	require.True(t, ok, "expected SubtractionNode, got %T", expr)

	left, ok := node.Left.(*analysis.UnitReference)
	require.True(t, ok, "expected left to be UnitReference, got %T", node.Left)
	assert.Equal(t, "a", left.Name)

	right, ok := node.Right.(*analysis.UnitReference)
	require.True(t, ok, "expected right to be UnitReference, got %T", node.Right)
	assert.Equal(t, "b", right.Name)
}

// TestBuildExpressionTree_Parentheses verifies that postfix already encodes
// precedence correctly: (a + b) * c -> [a, b, +, c, *].
func TestBuildExpressionTree_Parentheses(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"a": {Name: "a", ConceptName: "A"},
		"b": {Name: "b", ConceptName: "B"},
		"c": {Name: "c", ConceptName: "C"},
	}
	defs := []symboltable.SymbolDefinition{
		makeUnitRef("a"),
		makeUnitRef("b"),
		makeOperator(symboltable.PlusOperatorDefinition, parser.OpPlus),
		makeUnitRef("c"),
		makeOperator(symboltable.MultiplyOperatorDefinition, parser.OpMultiply),
	}
	expr, err := analysis.BuildExpressionTree(defs, units, nil)
	require.NoError(t, err)

	// Expected tree: MultiplicationNode{Left: AdditionNode{a,b}, Right: c}
	mulNode, ok := expr.(*analysis.MultiplicationNode)
	require.True(t, ok, "expected MultiplicationNode, got %T", expr)

	addNode, ok := mulNode.Left.(*analysis.AdditionNode)
	require.True(t, ok, "expected left to be AdditionNode, got %T", mulNode.Left)

	leftA, ok := addNode.Left.(*analysis.UnitReference)
	require.True(t, ok, "expected addNode.Left to be UnitReference, got %T", addNode.Left)
	assert.Equal(t, "a", leftA.Name)

	leftB, ok := addNode.Right.(*analysis.UnitReference)
	require.True(t, ok, "expected addNode.Right to be UnitReference, got %T", addNode.Right)
	assert.Equal(t, "b", leftB.Name)

	rightC, ok := mulNode.Right.(*analysis.UnitReference)
	require.True(t, ok, "expected mulNode.Right to be UnitReference, got %T", mulNode.Right)
	assert.Equal(t, "c", rightC.Name)
}

// TestBuildExpressionTree_NestedOperators verifies nested binary operators:
// a + (b * c) -> [a, b, c, *, +].
func TestBuildExpressionTree_NestedOperators(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"a": {Name: "a", ConceptName: "A"},
		"b": {Name: "b", ConceptName: "B"},
		"c": {Name: "c", ConceptName: "C"},
	}
	defs := []symboltable.SymbolDefinition{
		makeUnitRef("a"),
		makeUnitRef("b"),
		makeUnitRef("c"),
		makeOperator(symboltable.MultiplyOperatorDefinition, parser.OpMultiply),
		makeOperator(symboltable.PlusOperatorDefinition, parser.OpPlus),
	}
	expr, err := analysis.BuildExpressionTree(defs, units, nil)
	require.NoError(t, err)

	// Expected tree: AdditionNode{Left: a, Right: MultiplicationNode{b,c}}
	addNode, ok := expr.(*analysis.AdditionNode)
	require.True(t, ok, "expected AdditionNode, got %T", expr)

	leftA, ok := addNode.Left.(*analysis.UnitReference)
	require.True(t, ok, "expected addNode.Left to be UnitReference, got %T", addNode.Left)
	assert.Equal(t, "a", leftA.Name)

	mulNode, ok := addNode.Right.(*analysis.MultiplicationNode)
	require.True(t, ok, "expected addNode.Right to be MultiplicationNode, got %T", addNode.Right)

	leftB, ok := mulNode.Left.(*analysis.UnitReference)
	require.True(t, ok, "expected mulNode.Left to be UnitReference, got %T", mulNode.Left)
	assert.Equal(t, "b", leftB.Name)

	leftC, ok := mulNode.Right.(*analysis.UnitReference)
	require.True(t, ok, "expected mulNode.Right to be UnitReference, got %T", mulNode.Right)
	assert.Equal(t, "c", leftC.Name)
}

// TestBuildExpressionTree_NumberNode verifies that a numeric literal alone
// produces a NumberNode.
func TestBuildExpressionTree_NumberNode(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{}
	defs := []symboltable.SymbolDefinition{
		makeNumber(3.14),
	}
	expr, err := analysis.BuildExpressionTree(defs, units, nil)
	require.NoError(t, err)

	node, ok := expr.(*analysis.NumberNode)
	require.True(t, ok, "expected NumberNode, got %T", expr)
	assert.InDelta(t, 3.14, node.Value, 0.0001, "expected Value=3.14")
}

// TestBuildExpressionTree_MalformedEmptyInput verifies that an empty input
// returns an error.
func TestBuildExpressionTree_MalformedEmptyInput(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{}
	defs := []symboltable.SymbolDefinition{}
	_, err := analysis.BuildExpressionTree(defs, units, nil)
	require.Error(t, err, "expected error for empty input")
	assert.Contains(t, err.Error(), "malformed expression", "error should indicate malformed expression")
}

// TestBuildExpressionTree_UndefinedUnit verifies that referencing an undefined
// unit creates a UnitReference with a nil UnitGroup (validation is deferred).
func TestBuildExpressionTree_UndefinedUnit(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{}
	defs := []symboltable.SymbolDefinition{
		makeUnitRef("Bogus"),
	}
	expr, err := analysis.BuildExpressionTree(defs, units, nil)
	require.NoError(t, err, "undefined units are allowed during tree building")
	ref, ok := expr.(*analysis.UnitReference)
	require.True(t, ok, "result should be a UnitReference")
	assert.Equal(t, "Bogus", ref.Name)
	assert.Nil(t, ref.UnitGroup, "UnitGroup should be nil for undefined unit")
}

// TestBuildExpressionTree_InsufficientOperands verifies that an operator with
// too few operands returns an error.
func TestBuildExpressionTree_InsufficientOperands(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{}
	defs := []symboltable.SymbolDefinition{
		makeOperator(symboltable.MultiplyOperatorDefinition, parser.OpMultiply),
	}
	_, err := analysis.BuildExpressionTree(defs, units, nil)
	require.Error(t, err, "expected error for insufficient operands")
	assert.Contains(t, err.Error(), "insufficient operands", "error should indicate insufficient operands")
}

// TestBuildExpressionTree_TooManyOperands verifies that too many values on the
// stack after evaluation returns an error.
func TestBuildExpressionTree_TooManyOperands(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"a": {Name: "a", ConceptName: "A"},
		"b": {Name: "b", ConceptName: "B"},
		"c": {Name: "c", ConceptName: "C"},
	}
	defs := []symboltable.SymbolDefinition{
		makeUnitRef("a"),
		makeUnitRef("b"),
		makeUnitRef("c"),
	}
	_, err := analysis.BuildExpressionTree(defs, units, nil)
	require.Error(t, err, "expected error for too many operands")
	assert.Contains(t, err.Error(), "malformed expression", "error should indicate malformed expression")
}

// TestBuildExpressionTree_ComplexExpressionWithNumbers verifies a complex
// expression mixing numbers and unit references.
// Expression: (5 / 9) * (C / F)
// Postfix: [5, 9, /, C, F, /, *]
func TestBuildExpressionTree_ComplexExpressionWithNumbers(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"C": {Name: "C", ConceptName: "Temp"},
		"F": {Name: "F", ConceptName: "Temp"},
	}
	defs := []symboltable.SymbolDefinition{
		makeNumber(5),
		makeNumber(9),
		makeOperator(symboltable.DivideOperatorDefinition, parser.OpDivide),
		makeUnitRef("C"),
		makeUnitRef("F"),
		makeOperator(symboltable.DivideOperatorDefinition, parser.OpDivide),
		makeOperator(symboltable.MultiplyOperatorDefinition, parser.OpMultiply),
	}
	expr, err := analysis.BuildExpressionTree(defs, units, nil)
	require.NoError(t, err)

	// Expected tree: MultiplicationNode{Left: DivisionNode{5,9}, Right: DivisionNode{C,F}}
	mulNode, ok := expr.(*analysis.MultiplicationNode)
	require.True(t, ok, "expected MultiplicationNode, got %T", expr)

	leftDiv, ok := mulNode.Left.(*analysis.DivisionNode)
	require.True(t, ok, "expected mulNode.Left to be DivisionNode, got %T", mulNode.Left)

	leftNum, ok := leftDiv.Left.(*analysis.NumberNode)
	require.True(t, ok, "expected leftDiv.Left to be NumberNode, got %T", leftDiv.Left)
	assert.InDelta(t, 5.0, leftNum.Value, 0.0001)

	rightNum, ok := leftDiv.Right.(*analysis.NumberNode)
	require.True(t, ok, "expected leftDiv.Right to be NumberNode, got %T", leftDiv.Right)
	assert.InDelta(t, 9.0, rightNum.Value, 0.0001)

	rightDiv, ok := mulNode.Right.(*analysis.DivisionNode)
	require.True(t, ok, "expected mulNode.Right to be DivisionNode, got %T", mulNode.Right)

	leftC, ok := rightDiv.Left.(*analysis.UnitReference)
	require.True(t, ok, "expected rightDiv.Left to be UnitReference, got %T", rightDiv.Left)
	assert.Equal(t, "C", leftC.Name)

	rightF, ok := rightDiv.Right.(*analysis.UnitReference)
	require.True(t, ok, "expected rightDiv.Right to be UnitReference, got %T", rightDiv.Right)
	assert.Equal(t, "F", rightF.Name)
}

// TestBuildExpressionTree_NilVarsPreservesExistingBehavior verifies that passing
// nil for vars preserves the existing behavior: undefined units create UnitReference
// nodes with a nil UnitGroup.
func TestBuildExpressionTree_NilVarsPreservesExistingBehavior(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{}
	defs := []symboltable.SymbolDefinition{
		makeUnitRef("Bogus"),
	}
	expr, err := analysis.BuildExpressionTree(defs, units, nil)
	require.NoError(t, err, "undefined units are allowed when vars is nil")

	ref, ok := expr.(*analysis.UnitReference)
	require.True(t, ok, "result should be a UnitReference")
	assert.Equal(t, "Bogus", ref.Name)
	assert.Nil(t, ref.UnitGroup, "UnitGroup should be nil for undefined unit")
}

// TestBuildExpressionTree_VarsLookupCreatesVariableRef verifies that when a name
// appears in the vars map, a VariableRef node is created instead of a UnitReference.
func TestBuildExpressionTree_VarsLookupCreatesVariableRef(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"Length": {Name: "Length", ConceptName: "DistanceUnits"},
	}
	vars := map[string]bool{"x": true}
	defs := []symboltable.SymbolDefinition{
		makeUnitRef("x"),
	}
	expr, err := analysis.BuildExpressionTree(defs, units, vars)
	require.NoError(t, err)

	varRef, ok := expr.(*analysis.VariableRef)
	require.True(t, ok, "expected VariableRef, got %T", expr)
	assert.Equal(t, "x", varRef.Name)
}

// TestBuildExpressionTree_VarsLookupMixedWithUnitRef verifies a mixed expression
// where some names are variables and others are units.
func TestBuildExpressionTree_VarsLookupMixedWithUnitRef(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"Length": {Name: "Length", ConceptName: "DistanceUnits"},
	}
	vars := map[string]bool{"x": true}
	defs := []symboltable.SymbolDefinition{
		makeUnitRef("x"),
		makeUnitRef("Length"),
		makeOperator(symboltable.MultiplyOperatorDefinition, parser.OpMultiply),
	}
	expr, err := analysis.BuildExpressionTree(defs, units, vars)
	require.NoError(t, err)

	node, ok := expr.(*analysis.MultiplicationNode)
	require.True(t, ok, "expected MultiplicationNode, got %T", expr)

	left, ok := node.Left.(*analysis.VariableRef)
	require.True(t, ok, "expected left to be VariableRef, got %T", node.Left)
	assert.Equal(t, "x", left.Name)

	right, ok := node.Right.(*analysis.UnitReference)
	require.True(t, ok, "expected right to be UnitReference, got %T", node.Right)
	assert.Equal(t, "Length", right.Name)
	require.NotNil(t, right.UnitGroup)
	assert.Equal(t, "DistanceUnits", right.UnitGroup.ConceptName)
}

// TestBuildExpressionTree_UndefinedIdentifierReturnsError verifies that when vars
// is non-nil and a name is in neither vars nor units, an error is returned.
func TestBuildExpressionTree_UndefinedIdentifierReturnsError(t *testing.T) {
	t.Parallel()

	units := map[string]analysis.AnalysisUnitDefinition{
		"Length": {Name: "Length", ConceptName: "DistanceUnits"},
	}
	vars := map[string]bool{"x": true}
	defs := []symboltable.SymbolDefinition{
		makeUnitRef("Unknown"),
	}
	_, err := analysis.BuildExpressionTree(defs, units, vars)
	require.Error(t, err, "expected error for undefined identifier when vars is provided")
	assert.Contains(t, err.Error(), "undefined identifier", "error should indicate undefined identifier")
	assert.Contains(t, err.Error(), "Unknown", "error should mention the undefined identifier")
}
