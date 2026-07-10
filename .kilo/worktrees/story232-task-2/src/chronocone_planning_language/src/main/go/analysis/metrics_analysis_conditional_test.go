// Package analysis contains internal tests for the conditional assembly algorithm.
//
// These tests follow Behavioral Test-Driven Development (BTDD) and test
// unexported functions (assembleConditional, symbolDefToPostfixElem,
// buildCILTreeFromSymbolDefs) directly via package-internal access.
//
// Story: Story187 - CPL Metrics Conditionals Parse and Analyze
// Task: Task 3d Phase B - Logical tests for conditional assembly
package analysis

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parsing"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// ============================================================================
// symbolDefToPostfixElem Tests
// ============================================================================

// TestSymbolDefToPostfixElem_Identifier verifies that an Identifier
// SymbolDefinition with a name produces a PostfixIdentifier element.
func TestSymbolDefToPostfixElem_Identifier(t *testing.T) {
	t.Parallel()

	name := "my_input"
	def := symboltable.SymbolDefinition{
		Name:           &name,
		DefinitionType: symboltable.Identifier,
	}

	elem, err := symbolDefToPostfixElem(def, "test_metric")

	require.NoError(t, err)
	id, ok := elem.(parsing.PostfixIdentifier)
	require.True(t, ok, "Identifier should produce PostfixIdentifier element")
	assert.Equal(t, "my_input", id.Name)
}

// TestSymbolDefToPostfixElem_LiteralInt verifies that a LiteralDefinition
// with int Data produces a PostfixLiteral element.
func TestSymbolDefToPostfixElem_LiteralInt(t *testing.T) {
	t.Parallel()

	def := symboltable.SymbolDefinition{
		DefinitionType: symboltable.LiteralDefinition,
		Data:           symboltable.IntData{Value: 42},
	}

	elem, err := symbolDefToPostfixElem(def, "test_metric")

	require.NoError(t, err)
	lit, ok := elem.(parsing.PostfixLiteral)
	require.True(t, ok, "LiteralDefinition with int Data should produce PostfixLiteral element")
	assert.Equal(t, float64(42), lit.Value)
}

// TestSymbolDefToPostfixElem_LiteralFloat64 verifies that a LiteralDefinition
// with float64 Data produces a PostfixLiteral element.
func TestSymbolDefToPostfixElem_LiteralFloat64(t *testing.T) {
	t.Parallel()

	def := symboltable.SymbolDefinition{
		DefinitionType: symboltable.LiteralDefinition,
		Data:           symboltable.FloatData{Value: 3.14},
	}

	elem, err := symbolDefToPostfixElem(def, "test_metric")

	require.NoError(t, err)
	lit, ok := elem.(parsing.PostfixLiteral)
	require.True(t, ok, "LiteralDefinition with float64 Data should produce PostfixLiteral element")
	assert.Equal(t, 3.14, lit.Value)
}

// TestSymbolDefToPostfixElem_Operator verifies that an operator definition
// produces a PostfixOperator with the correct DefType.
func TestSymbolDefToPostfixElem_Operator(t *testing.T) {
	t.Parallel()

	def := symboltable.SymbolDefinition{
		DefinitionType: symboltable.PlusOperatorDefinition,
	}

	elem, err := symbolDefToPostfixElem(def, "test_metric")

	require.NoError(t, err)
	op, ok := elem.(parsing.PostfixOperator)
	require.True(t, ok, "Operator definition should produce PostfixOperator element")
	assert.Equal(t, symboltable.PlusOperatorDefinition, op.DefType)
}

// TestSymbolDefToPostfixElem_BoolLiteral verifies that TrueLiteralDefinition
// produces a PostfixOperator with the correct DefType.
func TestSymbolDefToPostfixElem_BoolLiteral(t *testing.T) {
	t.Parallel()

	def := symboltable.SymbolDefinition{
		DefinitionType: symboltable.TrueLiteralDefinition,
	}

	elem, err := symbolDefToPostfixElem(def, "test_metric")

	require.NoError(t, err)
	op, ok := elem.(parsing.PostfixOperator)
	require.True(t, ok, "Bool literal should produce PostfixOperator element")
	assert.Equal(t, symboltable.TrueLiteralDefinition, op.DefType)
}

// TestSymbolDefToPostfixElem_NilNameIdentifier_ReturnsError verifies that an
// Identifier definition with nil Name returns an error.
func TestSymbolDefToPostfixElem_NilNameIdentifier_ReturnsError(t *testing.T) {
	t.Parallel()

	def := symboltable.SymbolDefinition{
		Name:           nil,
		DefinitionType: symboltable.Identifier,
	}

	_, err := symbolDefToPostfixElem(def, "test_metric")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "neither name nor string data")
}

// TestSymbolDefToPostfixElem_NilDataLiteral_ReturnsError verifies that a
// LiteralDefinition with nil Data returns an error.
func TestSymbolDefToPostfixElem_NilDataLiteral_ReturnsError(t *testing.T) {
	t.Parallel()

	def := symboltable.SymbolDefinition{
		DefinitionType: symboltable.LiteralDefinition,
		Data:           nil,
	}

	_, err := symbolDefToPostfixElem(def, "test_metric")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil data")
}

// ============================================================================
// buildCILTreeFromSymbolDefs Tests
// ============================================================================

// TestBuildCILTreeFromSymbolDefs_Literal verifies that a single literal
// SymbolDefinition produces a CILLiteral node.
func TestBuildCILTreeFromSymbolDefs_Literal(t *testing.T) {
	t.Parallel()

	defs := []symboltable.SymbolDefinition{
		{DefinitionType: symboltable.LiteralDefinition, Data: symboltable.IntData{Value: 42}},
	}

	analyzer := NewCPLAnalyzer()
	node, err := buildCILTreeFromSymbolDefs(defs, "test_metric", analyzer)

	require.NoError(t, err)
	require.NotNil(t, node)
	literal, ok := node.CILNode.(*cil.CILLiteral)
	require.True(t, ok, "CILNode should be CILLiteral")
	assert.Equal(t, float64(42), literal.Value)
	assert.Empty(t, node.InputRefs)
}

// TestBuildCILTreeFromSymbolDefs_Identifier verifies that an identifier
// SymbolDefinition produces a CILInputRef node.
func TestBuildCILTreeFromSymbolDefs_Identifier(t *testing.T) {
	t.Parallel()

	name := "my_measure"
	defs := []symboltable.SymbolDefinition{
		{Name: &name, DefinitionType: symboltable.Identifier},
	}

	analyzer := NewCPLAnalyzer()
	defs_ := analyzer.GetAnalysisDefinitions()
	defs_.Measures["my_measure"] = AnalysisMeasureDefinition{
		Name:     "my_measure",
		TypeName: "Number",
		UnitExpression: &UnitReference{Name: "Unit"},
	}

	node, err := buildCILTreeFromSymbolDefs(defs, "test_metric", analyzer)

	require.NoError(t, err)
	require.NotNil(t, node)
	inputRef, ok := node.CILNode.(*cil.CILInputRef)
	require.True(t, ok, "CILNode should be CILInputRef")
	assert.Equal(t, "connector", inputRef.Kind)
	assert.Equal(t, "my_measure", inputRef.Name)
	require.Len(t, node.InputRefs, 1)
	assert.Equal(t, "my_measure", node.InputRefs[0].Name)
}

// TestBuildCILTreeFromSymbolDefs_Comparison verifies that a comparison
// expression (a > 5) produces a CILComparisonOp node.
func TestBuildCILTreeFromSymbolDefs_Comparison(t *testing.T) {
	t.Parallel()

	aName := "a"
	defs := []symboltable.SymbolDefinition{
		{Name: &aName, DefinitionType: symboltable.Identifier},
		{DefinitionType: symboltable.LiteralDefinition, Data: symboltable.IntData{Value: 5}},
		{DefinitionType: symboltable.GreaterThanOperatorDefinition},
	}

	analyzer := NewCPLAnalyzer()
	defs_ := analyzer.GetAnalysisDefinitions()
	defs_.Measures["a"] = AnalysisMeasureDefinition{
		Name:     "a",
		TypeName: "Number",
		UnitExpression: &UnitReference{Name: "Unit"},
	}

	node, err := buildCILTreeFromSymbolDefs(defs, "test_metric", analyzer)

	require.NoError(t, err)
	require.NotNil(t, node)
	cmpOp, ok := node.CILNode.(*cil.CILComparisonOp)
	require.True(t, ok, "CILNode should be CILComparisonOp")
	assert.Equal(t, operator.GreaterThan, cmpOp.Op)

	// Left should be CILInputRef{a}
	leftRef, ok := cmpOp.Left.(*cil.CILInputRef)
	require.True(t, ok, "Left should be CILInputRef")
	assert.Equal(t, "a", leftRef.Name)

	// Right should be CILLiteral{5}
	rightLit, ok := cmpOp.Right.(*cil.CILLiteral)
	require.True(t, ok, "Right should be CILLiteral")
	assert.Equal(t, float64(5), rightLit.Value)

	require.Len(t, node.InputRefs, 1)
	assert.Equal(t, "a", node.InputRefs[0].Name)
}

// TestBuildCILTreeFromSymbolDefs_EmptySlice_ReturnsError verifies that an
// empty definition slice returns an error (via buildCILTreeFromPostfix).
func TestBuildCILTreeFromSymbolDefs_EmptySlice_ReturnsError(t *testing.T) {
	t.Parallel()

	defs := []symboltable.SymbolDefinition{}

	analyzer := NewCPLAnalyzer()
	_, err := buildCILTreeFromSymbolDefs(defs, "test_metric", analyzer)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "stack")
}

// ============================================================================
// assembleConditional Tests — If/Else Chain
// ============================================================================

// TestAssembleConditional_IfElse verifies that an if/else conditional
// structure produces a CILConditional with two branches.
func TestAssembleConditional_IfElse(t *testing.T) {
	t.Parallel()

	// Build: if (a > 5) then ret1 else ret2
	// Condition postfix: a 5 >
	// If body: ret1 (just identifier)
	// Else body: ret2 (just identifier)
	root := symboltable.NewGlobalSymbolTable()
	metricScope := root.AddChild(symboltable.MetricScope)

	aName := "a"
	ret1Name := "ret1"
	ret2Name := "ret2"

	// If clause: condition = a > 5, body = ret1
	ifBodyScope := metricScope.AddChild(symboltable.IfBodyScope)
	ifBodyScope.Definitions = []symboltable.SymbolDefinition{
		{Name: &ret1Name, DefinitionType: symboltable.Identifier},
	}

	ifClauseData := &symboltable.ClauseData{
		ConditionPostfix: []symboltable.SymbolDefinition{
			{Name: &aName, DefinitionType: symboltable.Identifier},
			{DefinitionType: symboltable.LiteralDefinition, Data: symboltable.IntData{Value: 5}},
			{DefinitionType: symboltable.GreaterThanOperatorDefinition},
		},
		BodyScope: ifBodyScope,
	}
	metricScope.Definitions = append(metricScope.Definitions, symboltable.SymbolDefinition{
		DefinitionType: symboltable.IfClauseDefinition,
		Data:           ifClauseData,
	})

	// Else clause: body = ret2 (no condition)
	elseBodyScope := metricScope.AddChild(symboltable.ElseBodyScope)
	elseBodyScope.Definitions = []symboltable.SymbolDefinition{
		{Name: &ret2Name, DefinitionType: symboltable.Identifier},
	}

	elseClauseData := &symboltable.ClauseData{
		ConditionPostfix: nil,
		BodyScope:        elseBodyScope,
	}
	metricScope.Definitions = append(metricScope.Definitions, symboltable.SymbolDefinition{
		DefinitionType: symboltable.ElseClauseDefinition,
		Data:           elseClauseData,
	})

	analyzer := NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["a"] = AnalysisMeasureDefinition{Name: "a", TypeName: "Number", UnitExpression: &UnitReference{Name: "Unit"}}
	defs.Measures["ret1"] = AnalysisMeasureDefinition{Name: "ret1", TypeName: "Number", UnitExpression: &UnitReference{Name: "Unit"}}
	defs.Measures["ret2"] = AnalysisMeasureDefinition{Name: "ret2", TypeName: "Number", UnitExpression: &UnitReference{Name: "Unit"}}

	node, err := assembleConditional(metricScope, "test_metric", analyzer)

	require.NoError(t, err)
	require.NotNil(t, node)

	cond, ok := node.CILNode.(*cil.CILConditional)
	require.True(t, ok, "CILNode should be CILConditional")
	require.Len(t, cond.Branches, 2, "Should have 2 branches (if + else)")

	// First branch: if (a > 5) then ret1
	branch0 := cond.Branches[0]
	require.NotNil(t, branch0.Condition, "If branch should have a condition")
	cmpOp, ok := branch0.Condition.(*cil.CILComparisonOp)
	require.True(t, ok, "Condition should be CILComparisonOp")
	assert.Equal(t, operator.GreaterThan, cmpOp.Op)

	body0, ok := branch0.Body.(*cil.CILInputRef)
	require.True(t, ok, "If body should be CILInputRef")
	assert.Equal(t, "ret1", body0.Name)

	// Second branch: else → ret2
	branch1 := cond.Branches[1]
	assert.Nil(t, branch1.Condition, "Else branch should have nil condition")
	body1, ok := branch1.Body.(*cil.CILInputRef)
	require.True(t, ok, "Else body should be CILInputRef")
	assert.Equal(t, "ret2", body1.Name)

	// Verify InputRefs collection (a, ret1, ret2)
	require.Len(t, node.InputRefs, 3)
}

// TestAssembleConditional_IfElseIfElse verifies an if/elseif/else chain
// produces a CILConditional with three branches.
func TestAssembleConditional_IfElseIfElse(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricScope := root.AddChild(symboltable.MetricScope)

	aName := "a"
	ret1Name, ret2Name, ret3Name := "ret1", "ret2", "ret3"

	// If: condition = a > 10, body = ret1
	ifBodyScope := metricScope.AddChild(symboltable.IfBodyScope)
	ifBodyScope.Definitions = []symboltable.SymbolDefinition{
		{Name: &ret1Name, DefinitionType: symboltable.Identifier},
	}
	metricScope.Definitions = append(metricScope.Definitions, symboltable.SymbolDefinition{
		DefinitionType: symboltable.IfClauseDefinition,
		Data: &symboltable.ClauseData{
			ConditionPostfix: []symboltable.SymbolDefinition{
				{Name: &aName, DefinitionType: symboltable.Identifier},
				{DefinitionType: symboltable.LiteralDefinition, Data: symboltable.IntData{Value: 10}},
				{DefinitionType: symboltable.GreaterThanOperatorDefinition},
			},
			BodyScope: ifBodyScope,
		},
	})

	// ElseIf: condition = a == 5, body = ret2
	elseIfBodyScope := metricScope.AddChild(symboltable.ElseIfBodyScope)
	elseIfBodyScope.Definitions = []symboltable.SymbolDefinition{
		{Name: &ret2Name, DefinitionType: symboltable.Identifier},
	}
	metricScope.Definitions = append(metricScope.Definitions, symboltable.SymbolDefinition{
		DefinitionType: symboltable.ElseIfClauseDefinition,
		Data: &symboltable.ClauseData{
			ConditionPostfix: []symboltable.SymbolDefinition{
				{Name: &aName, DefinitionType: symboltable.Identifier},
				{DefinitionType: symboltable.LiteralDefinition, Data: symboltable.IntData{Value: 5}},
				{DefinitionType: symboltable.EqOperatorDefinition},
			},
			BodyScope: elseIfBodyScope,
		},
	})

	// Else: body = ret3
	elseBodyScope := metricScope.AddChild(symboltable.ElseBodyScope)
	elseBodyScope.Definitions = []symboltable.SymbolDefinition{
		{Name: &ret3Name, DefinitionType: symboltable.Identifier},
	}
	metricScope.Definitions = append(metricScope.Definitions, symboltable.SymbolDefinition{
		DefinitionType: symboltable.ElseClauseDefinition,
		Data: &symboltable.ClauseData{
			BodyScope: elseBodyScope,
		},
	})

	analyzer := NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	for _, name := range []string{"a", "b", "ret1", "ret2", "ret3"} {
		defs.Measures[name] = AnalysisMeasureDefinition{Name: name, TypeName: "Number", UnitExpression: &UnitReference{Name: "Unit"}}
	}

	node, err := assembleConditional(metricScope, "test_metric", analyzer)

	require.NoError(t, err)
	require.NotNil(t, node)

	cond, ok := node.CILNode.(*cil.CILConditional)
	require.True(t, ok, "CILNode should be CILConditional")
	require.Len(t, cond.Branches, 3)

	// Branch 0: if (a > 10)
	cmp0, ok := cond.Branches[0].Condition.(*cil.CILComparisonOp)
	require.True(t, ok)
	assert.Equal(t, operator.GreaterThan, cmp0.Op)

	// Branch 1: elseif (a == 5)
	cmp1, ok := cond.Branches[1].Condition.(*cil.CILComparisonOp)
	require.True(t, ok)
	assert.Equal(t, operator.Equal, cmp1.Op)

	// Branch 2: else (nil condition)
	assert.Nil(t, cond.Branches[2].Condition)
}

// ============================================================================
// assembleConditional Tests — When/Default Chain
// ============================================================================

// TestAssembleConditional_WhenDefault verifies a when/default conditional
// produces a CILWhenClause with cases and a default body.
func TestAssembleConditional_WhenDefault(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricScope := root.AddChild(symboltable.MetricScope)

	aName := "a"
	ret1Name, ret2Name, retDefaultName := "ret1", "ret2", "ret_default"

	// When case 1: condition = a == 1, body = ret1
	whenBody1 := metricScope.AddChild(symboltable.WhenBodyScope)
	whenBody1.Definitions = []symboltable.SymbolDefinition{
		{Name: &ret1Name, DefinitionType: symboltable.Identifier},
	}
	metricScope.Definitions = append(metricScope.Definitions, symboltable.SymbolDefinition{
		DefinitionType: symboltable.WhenClauseDefinition,
		Data: &symboltable.ClauseData{
			ConditionPostfix: []symboltable.SymbolDefinition{
				{Name: &aName, DefinitionType: symboltable.Identifier},
				{DefinitionType: symboltable.LiteralDefinition, Data: symboltable.IntData{Value: 1}},
				{DefinitionType: symboltable.EqOperatorDefinition},
			},
			BodyScope: whenBody1,
		},
	})

	// When case 2: condition = a == 2, body = ret2
	whenBody2 := metricScope.AddChild(symboltable.WhenBodyScope)
	whenBody2.Definitions = []symboltable.SymbolDefinition{
		{Name: &ret2Name, DefinitionType: symboltable.Identifier},
	}
	metricScope.Definitions = append(metricScope.Definitions, symboltable.SymbolDefinition{
		DefinitionType: symboltable.WhenClauseDefinition,
		Data: &symboltable.ClauseData{
			ConditionPostfix: []symboltable.SymbolDefinition{
				{Name: &aName, DefinitionType: symboltable.Identifier},
				{DefinitionType: symboltable.LiteralDefinition, Data: symboltable.IntData{Value: 2}},
				{DefinitionType: symboltable.EqOperatorDefinition},
			},
			BodyScope: whenBody2,
		},
	})

	// Default: body = ret_default
	defaultBodyScope := metricScope.AddChild(symboltable.DefaultBodyScope)
	defaultBodyScope.Definitions = []symboltable.SymbolDefinition{
		{Name: &retDefaultName, DefinitionType: symboltable.Identifier},
	}
	metricScope.Definitions = append(metricScope.Definitions, symboltable.SymbolDefinition{
		DefinitionType: symboltable.DefaultClauseDefinition,
		Data: &symboltable.ClauseData{
			BodyScope: defaultBodyScope,
		},
	})

	analyzer := NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	for _, name := range []string{"a", "ret1", "ret2", "ret_default"} {
		defs.Measures[name] = AnalysisMeasureDefinition{Name: name, TypeName: "Number", UnitExpression: &UnitReference{Name: "Unit"}}
	}

	node, err := assembleConditional(metricScope, "test_metric", analyzer)

	require.NoError(t, err)
	require.NotNil(t, node)

	when, ok := node.CILNode.(*cil.CILWhenClause)
	require.True(t, ok, "CILNode should be CILWhenClause")
	require.Len(t, when.Clauses, 2, "Should have 2 when cases")

	// Case 0: a == 1 → ret1
	case0 := when.Clauses[0]
	cmp0, ok := case0.Condition.(*cil.CILComparisonOp)
	require.True(t, ok, "Case 0 condition should be CILComparisonOp")
	assert.Equal(t, operator.Equal, cmp0.Op)
	body0, ok := case0.Body.(*cil.CILInputRef)
	require.True(t, ok)
	assert.Equal(t, "ret1", body0.Name)

	// Case 1: a == 2 → ret2
	case1 := when.Clauses[1]
	cmp1, ok := case1.Condition.(*cil.CILComparisonOp)
	require.True(t, ok)
	assert.Equal(t, operator.Equal, cmp1.Op)
	body1, ok := case1.Body.(*cil.CILInputRef)
	require.True(t, ok)
	assert.Equal(t, "ret2", body1.Name)

	// Default: ret_default
	require.NotNil(t, when.Default, "Default should not be nil")
	defaultBody, ok := when.Default.(*cil.CILInputRef)
	require.True(t, ok)
	assert.Equal(t, "ret_default", defaultBody.Name)
}

// TestAssembleConditional_WhenOnly_NoDefault verifies a when clause with
// no default produces a CILWhenClause with nil Default.
func TestAssembleConditional_WhenOnly_NoDefault(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricScope := root.AddChild(symboltable.MetricScope)

	aName := "a"
	ret1Name := "ret1"

	// When case: condition = a == 1, body = ret1
	whenBody := metricScope.AddChild(symboltable.WhenBodyScope)
	whenBody.Definitions = []symboltable.SymbolDefinition{
		{Name: &ret1Name, DefinitionType: symboltable.Identifier},
	}
	metricScope.Definitions = append(metricScope.Definitions, symboltable.SymbolDefinition{
		DefinitionType: symboltable.WhenClauseDefinition,
		Data: &symboltable.ClauseData{
			ConditionPostfix: []symboltable.SymbolDefinition{
				{Name: &aName, DefinitionType: symboltable.Identifier},
				{DefinitionType: symboltable.LiteralDefinition, Data: symboltable.IntData{Value: 1}},
				{DefinitionType: symboltable.EqOperatorDefinition},
			},
			BodyScope: whenBody,
		},
	})

	analyzer := NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["a"] = AnalysisMeasureDefinition{Name: "a", TypeName: "Number", UnitExpression: &UnitReference{Name: "Unit"}}
	defs.Measures["ret1"] = AnalysisMeasureDefinition{Name: "ret1", TypeName: "Number", UnitExpression: &UnitReference{Name: "Unit"}}

	node, err := assembleConditional(metricScope, "test_metric", analyzer)

	require.NoError(t, err)
	require.NotNil(t, node)

	when, ok := node.CILNode.(*cil.CILWhenClause)
	require.True(t, ok, "CILNode should be CILWhenClause")
	require.Len(t, when.Clauses, 1)
	assert.Nil(t, when.Default, "Default should be nil when no DefaultClauseDefinition present")
}

// ============================================================================
// assembleConditional Error Case Tests
// ============================================================================

// TestAssembleConditional_NoConditionalStructure_ReturnsError verifies that
// a scope with no conditional clause definitions returns an error.
func TestAssembleConditional_NoConditionalStructure_ReturnsError(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricScope := root.AddChild(symboltable.MetricScope)

	analyzer := NewCPLAnalyzer()
	_, err := assembleConditional(metricScope, "test_metric", analyzer)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no conditional clause definitions")
}

// TestAssembleConditional_MixedIfAndWhen_ReturnsError verifies that a scope
// with both if and when clause definitions returns an error.
func TestAssembleConditional_MixedIfAndWhen_ReturnsError(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricScope := root.AddChild(symboltable.MetricScope)

	aName := "a"

	// Add an IfClauseDefinition
	ifBodyScope := metricScope.AddChild(symboltable.IfBodyScope)
	metricScope.Definitions = append(metricScope.Definitions, symboltable.SymbolDefinition{
		DefinitionType: symboltable.IfClauseDefinition,
		Data: &symboltable.ClauseData{
			ConditionPostfix: []symboltable.SymbolDefinition{
				{DefinitionType: symboltable.TrueLiteralDefinition},
			},
			BodyScope: ifBodyScope,
		},
	})

	// Add a WhenClauseDefinition
	whenBodyScope := metricScope.AddChild(symboltable.WhenBodyScope)
	metricScope.Definitions = append(metricScope.Definitions, symboltable.SymbolDefinition{
		DefinitionType: symboltable.WhenClauseDefinition,
		Data: &symboltable.ClauseData{
			ConditionPostfix: []symboltable.SymbolDefinition{
				{Name: &aName, DefinitionType: symboltable.Identifier},
			},
			BodyScope: whenBodyScope,
		},
	})

	analyzer := NewCPLAnalyzer()
	_, err := assembleConditional(metricScope, "test_metric", analyzer)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "mixed if and when")
}

// TestAssembleConditional_NilClauseData_ReturnsError verifies that a clause
// definition with nil Data returns an error.
func TestAssembleConditional_NilClauseData_ReturnsError(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricScope := root.AddChild(symboltable.MetricScope)

	// IfClauseDefinition with nil Data
	metricScope.Definitions = append(metricScope.Definitions, symboltable.SymbolDefinition{
		DefinitionType: symboltable.IfClauseDefinition,
		Data:           nil,
	})

	analyzer := NewCPLAnalyzer()
	_, err := assembleConditional(metricScope, "test_metric", analyzer)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil Data")
}

// TestAssembleConditional_WrongData_ReturnsError verifies that a clause
// definition with a non-ClauseData Data value returns an error.
func TestAssembleConditional_WrongDataType_ReturnsError(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricScope := root.AddChild(symboltable.MetricScope)

	// IfClauseDefinition with wrong Data type
	metricScope.Definitions = append(metricScope.Definitions, symboltable.SymbolDefinition{
		DefinitionType: symboltable.IfClauseDefinition,
		Data:           symboltable.StringData{Value: "not clause data"},
	})

	analyzer := NewCPLAnalyzer()
	_, err := assembleConditional(metricScope, "test_metric", analyzer)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected Data type")
}

// TestAssembleConditional_NilBodyScope_ReturnsError verifies that ClauseData
// with nil BodyScope returns an error.
func TestAssembleConditional_NilBodyScope_ReturnsError(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricScope := root.AddChild(symboltable.MetricScope)

	metricScope.Definitions = append(metricScope.Definitions, symboltable.SymbolDefinition{
		DefinitionType: symboltable.IfClauseDefinition,
		Data: &symboltable.ClauseData{
			ConditionPostfix: []symboltable.SymbolDefinition{
				{DefinitionType: symboltable.TrueLiteralDefinition},
			},
			BodyScope: nil,
		},
	})

	analyzer := NewCPLAnalyzer()
	_, err := assembleConditional(metricScope, "test_metric", analyzer)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil BodyScope")
}

// ============================================================================
// Consolidated Validation (validateCILTree) Tests — Story213 Phase B
// ============================================================================
//
// These tests exercise the consolidated walker that replaced
// validateTruthiness, validateExpressionTypes, and validateBranchCompleteness.
// Where the old truthiness tests used nil-in-arithmetic scenarios, the
// consolidated walker catches them via the type check first (nil is
// non-numeric), producing "type error:" prefixed messages.

// TestValidateCILTree_NilLiteralTopLevel verifies that a nil literal
// at the expression root passes validation (root is not arithmetic).
func TestValidateCILTree_NilLiteralTopLevel(t *testing.T) {
	t.Parallel()

	node := &cil.CILNilLiteral{}
	err := validateCILTree(node)
	require.NoError(t, err, "Nil literal at top level should pass truthiness")
}

// TestValidateCILTree_NilInBinaryOpLeft_ReturnsError verifies that nil
// in the left operand of an arithmetic binary op returns an error (caught
// as type error by consolidated walker).
func TestValidateCILTree_NilInBinaryOpLeft_ReturnsError(t *testing.T) {
	t.Parallel()

	node := &cil.CILBinaryOp{
		Op: operator.Add,
		Left:  &cil.CILNilLiteral{},
		Right: &cil.CILLiteral{Value: 5},
	}
	err := validateCILTree(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type error: binary operator + has non-numeric left operand")
}

// TestValidateCILTree_NilInBinaryOpRight_ReturnsError verifies that nil
// in the right operand of an arithmetic binary op returns an error.
func TestValidateCILTree_NilInBinaryOpRight_ReturnsError(t *testing.T) {
	t.Parallel()

	node := &cil.CILBinaryOp{
		Op: operator.Multiply,
		Left:  &cil.CILLiteral{Value: 3},
		Right: &cil.CILNilLiteral{},
	}
	err := validateCILTree(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type error: binary operator * has non-numeric right operand")
}

// TestValidateCILTree_NilInUnaryOp_ReturnsError verifies that nil
// in the operand of a unary negate op returns an error.
func TestValidateCILTree_NilInUnaryOp_ReturnsError(t *testing.T) {
	t.Parallel()

	node := &cil.CILUnaryOp{
		Op: operator.Negate,
		Operand: &cil.CILNilLiteral{},
	}
	err := validateCILTree(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type error: unary operator negate has non-numeric operand")
}

// TestValidateCILTree_NilInComparison_Passes verifies that nil is
// permitted as a comparison operand.
func TestValidateCILTree_NilInComparison_Passes(t *testing.T) {
	t.Parallel()

	node := &cil.CILComparisonOp{
		Op: operator.Equal,
		Left:  &cil.CILNilLiteral{},
		Right: &cil.CILLiteral{Value: 0},
	}
	err := validateCILTree(node)
	require.NoError(t, err, "Nil in comparison should pass truthiness")
}

// TestValidateCILTree_NilDeepInArithmetic_ReturnsError verifies that nil
// anywhere deep in an arithmetic subtree returns an error.
func TestValidateCILTree_NilDeepInArithmetic_ReturnsError(t *testing.T) {
	t.Parallel()

	// Tree: (Nil + 5) * 3
	node := &cil.CILBinaryOp{
		Op: operator.Multiply,
		Left: &cil.CILBinaryOp{
			Op: operator.Add,
			Left:  &cil.CILNilLiteral{},
			Right: &cil.CILLiteral{Value: 5},
		},
		Right: &cil.CILLiteral{Value: 3},
	}
	err := validateCILTree(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type error: binary operator + has non-numeric left operand")
}

// TestValidateCILTree_NilInConditionalCondition_Passes verifies that nil
// is permitted as a conditional branch condition (condition context).
func TestValidateCILTree_NilInConditionalCondition_Passes(t *testing.T) {
	t.Parallel()

	node := &cil.CILConditional{
		Branches: []cil.CILConditionalBranch{
			{Condition: &cil.CILNilLiteral{}, Body: &cil.CILLiteral{Value: 1}},
			{Body: &cil.CILLiteral{Value: 0}}, // else, no condition
		},
	}
	err := validateCILTree(node)
	require.NoError(t, err, "Nil in conditional condition should pass truthiness")
}

// TestValidateCILTree_NilInConditionalBody_Arithmetic_ReturnsError verifies
// that nil in a conditional body that is in arithmetic context returns an error.
func TestValidateCILTree_NilInConditionalBody_Arithmetic_ReturnsError(t *testing.T) {
	t.Parallel()

	// Body contains nil in arithmetic: if True then (nil + 5) else 0
	node := &cil.CILConditional{
		Branches: []cil.CILConditionalBranch{
			{
				Condition: &cil.CILBoolLiteral{Value: true},
				Body: &cil.CILBinaryOp{
					Op: operator.Add,
					Left:  &cil.CILNilLiteral{},
					Right: &cil.CILLiteral{Value: 5},
				},
			},
			{Body: &cil.CILLiteral{Value: 0}},
		},
	}
	err := validateCILTree(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type error: binary operator + has non-numeric left operand")
}

// TestValidateCILTree_NilInWhenCondition_Passes verifies that nil is
// permitted as a when clause condition.
func TestValidateCILTree_NilInWhenCondition_Passes(t *testing.T) {
	t.Parallel()

	node := &cil.CILWhenClause{
		Clauses: []cil.CILWhenCase{
			{Condition: &cil.CILNilLiteral{}, Body: &cil.CILLiteral{Value: 42}},
		},
		Default: &cil.CILLiteral{Value: 0},
	}
	err := validateCILTree(node)
	require.NoError(t, err, "Nil in when condition should pass truthiness")
}

// TestValidateCILTree_NilInLogicalOp_Passes verifies that nil is permitted
// as a logical operand (condition context).
func TestValidateCILTree_NilInLogicalOp_Passes(t *testing.T) {
	t.Parallel()

	node := &cil.CILLogicalOp{
		Op: operator.And,
		Left:  &cil.CILNilLiteral{},
		Right: &cil.CILBoolLiteral{Value: true},
	}
	err := validateCILTree(node)
	require.NoError(t, err, "Nil in logical op should pass truthiness")
}

// TestValidateCILTree_ArithmeticWithValidNodes_Passes verifies that
// a valid arithmetic expression passes validation.
func TestValidateCILTree_ArithmeticWithValidNodes_Passes(t *testing.T) {
	t.Parallel()

	// (5 + 3) * 2
	node := &cil.CILBinaryOp{
		Op: operator.Multiply,
		Left: &cil.CILBinaryOp{
			Op: operator.Add,
			Left:  &cil.CILLiteral{Value: 5},
			Right: &cil.CILLiteral{Value: 3},
		},
		Right: &cil.CILLiteral{Value: 2},
	}
	err := validateCILTree(node)
	require.NoError(t, err, "Valid arithmetic should pass truthiness")
}

// TestValidateCILTree_NilNode_ReturnsError verifies that a nil CILNode
// returns an error.
func TestValidateCILTree_NilNode_ReturnsError(t *testing.T) {
	t.Parallel()

	err := validateCILTree(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil node")
}

// ============================================================================
// Branch Completeness Tests (via Consolidated validateCILTree) — Story213 Phase B
// ============================================================================

// TestValidateCILTree_ConditionalWithBranches_Passes verifies that a
// CILConditional with valid branches passes completeness check.
func TestValidateCILTree_ConditionalWithBranches_Passes(t *testing.T) {
	t.Parallel()

	node := &cil.CILConditional{
		Branches: []cil.CILConditionalBranch{
			{Condition: &cil.CILBoolLiteral{Value: true}, Body: &cil.CILLiteral{Value: 1}},
			{Body: &cil.CILLiteral{Value: 0}}, // else
		},
	}
	err := validateCILTree(node)
	require.NoError(t, err, "Conditional with valid branches should pass")
}

// TestValidateCILTree_ConditionalNoBranches_ReturnsError verifies
// that a CILConditional with an empty Branches slice returns an error.
func TestValidateCILTree_ConditionalNoBranches_ReturnsError(t *testing.T) {
	t.Parallel()

	node := &cil.CILConditional{
		Branches: []cil.CILConditionalBranch{},
	}
	err := validateCILTree(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conditional expression has no branches")
}

// TestValidateCILTree_ConditionalBranchNilBody_ReturnsError verifies
// that a branch with a nil Body returns an error.
func TestValidateCILTree_ConditionalBranchNilBody_ReturnsError(t *testing.T) {
	t.Parallel()

	node := &cil.CILConditional{
		Branches: []cil.CILConditionalBranch{
			{Condition: &cil.CILBoolLiteral{Value: true}, Body: nil},
		},
	}
	err := validateCILTree(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch 0 of conditional expression has nil body")
}

// TestValidateCILTree_ConditionalSecondBranchNilBody_ReturnsError
// verifies that a nil body in a non-first branch is also caught.
func TestValidateCILTree_ConditionalSecondBranchNilBody_ReturnsError(t *testing.T) {
	t.Parallel()

	node := &cil.CILConditional{
		Branches: []cil.CILConditionalBranch{
			{Condition: &cil.CILBoolLiteral{Value: true}, Body: &cil.CILLiteral{Value: 1}},
			{Body: nil}, // else with nil body
		},
	}
	err := validateCILTree(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch 1 of conditional expression has nil body")
}

// TestValidateCILTree_NestedConditional_Passes verifies that a
// nested conditional with all branches complete passes the check.
func TestValidateCILTree_NestedConditional_Passes(t *testing.T) {
	t.Parallel()

	node := &cil.CILConditional{
		Branches: []cil.CILConditionalBranch{
			{
				Condition: &cil.CILBoolLiteral{Value: true},
				Body: &cil.CILConditional{
					Branches: []cil.CILConditionalBranch{
						{Condition: &cil.CILBoolLiteral{Value: false}, Body: &cil.CILLiteral{Value: 2}},
						{Body: &cil.CILLiteral{Value: 3}},
					},
				},
			},
			{Body: &cil.CILLiteral{Value: 0}},
		},
	}
	err := validateCILTree(node)
	require.NoError(t, err, "Nested conditional with complete branches should pass")
}

// TestValidateCILTree_WhenClauseWithCases_Passes verifies that a
// CILWhenClause with valid cases passes completeness check.
func TestValidateCILTree_WhenClauseWithCases_Passes(t *testing.T) {
	t.Parallel()

	node := &cil.CILWhenClause{
		Clauses: []cil.CILWhenCase{
			{Condition: &cil.CILBoolLiteral{Value: true}, Body: &cil.CILLiteral{Value: 1}},
		},
		Default: &cil.CILLiteral{Value: 0},
	}
	err := validateCILTree(node)
	require.NoError(t, err, "When clause with cases should pass")
}

// TestValidateCILTree_WhenClauseNoClauses_ReturnsError verifies
// that a CILWhenClause with an empty Clauses slice returns an error.
func TestValidateCILTree_WhenClauseNoClauses_ReturnsError(t *testing.T) {
	t.Parallel()

	node := &cil.CILWhenClause{
		Clauses: []cil.CILWhenCase{},
	}
	err := validateCILTree(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "when clause must have at least one clause")
}

// TestValidateCILTree_WhenClauseNilBody_ReturnsError verifies
// that a when case with nil Body returns an error.
func TestValidateCILTree_WhenClauseNilBody_ReturnsError(t *testing.T) {
	t.Parallel()

	node := &cil.CILWhenClause{
		Clauses: []cil.CILWhenCase{
			{Condition: &cil.CILBoolLiteral{Value: true}, Body: nil},
		},
	}
	err := validateCILTree(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "when clause 0 has nil body")
}

// TestValidateCILTree_WhenClauseNilDefault_ReturnsError verifies that a
// CILWhenClause with nil Default (required by grammar) returns a completeness error.
func TestValidateCILTree_WhenClauseNilDefault_ReturnsError(t *testing.T) {
	t.Parallel()

	node := &cil.CILWhenClause{
		Clauses: []cil.CILWhenCase{
			{Condition: &cil.CILBoolLiteral{Value: true}, Body: &cil.CILLiteral{Value: 42}},
		},
		Default: nil, // missing required default
	}
	err := validateCILTree(node)
	require.Error(t, err, "When clause with nil default should return error")
	assert.Contains(t, err.Error(), "completeness error: when clause must have a default case")
}

// TestValidateCILTree_NonConditionalNode_Passes verifies that leaf
// nodes (literals, etc.) pass validation.
func TestValidateCILTree_NonConditionalNode_Passes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		node cil.CILNode
	}{
		{"literal", &cil.CILLiteral{Value: 42}},
		{"inputRef", &cil.CILInputRef{Kind: "metric", Name: "x"}},
		{"boolLiteral", &cil.CILBoolLiteral{Value: true}},
		{"nilLiteral", &cil.CILNilLiteral{}},
		{"binaryOp", &cil.CILBinaryOp{Op: operator.Add, Left: &cil.CILLiteral{Value: 1}, Right: &cil.CILLiteral{Value: 2}}},
		{"unaryOp", &cil.CILUnaryOp{Op: operator.Negate, Operand: &cil.CILLiteral{Value: 5}}},
		{"comparisonOp", &cil.CILComparisonOp{Op: operator.Equal, Left: &cil.CILLiteral{Value: 1}, Right: &cil.CILLiteral{Value: 2}}},
		{"logicalOp", &cil.CILLogicalOp{Op: operator.And, Left: &cil.CILBoolLiteral{Value: true}, Right: &cil.CILBoolLiteral{Value: false}}},
		{"notOp", &cil.CILNotOp{Operand: &cil.CILBoolLiteral{Value: true}}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateCILTree(tt.node)
			require.NoError(t, err, "%s should pass branch completeness", tt.name)
		})
	}
}
