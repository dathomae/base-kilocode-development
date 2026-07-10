package parser

import (
	"testing"

	antlrruntime "github.com/antlr4-go/antlr/v4"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// parseArithmeticExpr parses an arithmetic expression string and captures
// the postfix definitions on a TempScope.
func parseArithmeticExpr(input string) (*symboltable.SymbolTable, []error) {
	inputStream := antlrruntime.NewInputStream(input)
	lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)
	tokenStream := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
	parser := antlr.NewChronoconePlanningLanguageParser(tokenStream)

	listener := NewCPLListener(nil, lexer, parser)

	// Push a TempScope to simulate an expression-accumulating context
	global, _ := listener.symbolStack.Peek()
	tempScope := global.AddChild(symboltable.TempScope)
	listener.symbolStack.Push(tempScope)

	// Parse as arithmetic expression
	tree := parser.ArithmeticExpr()
	antlrruntime.ParseTreeWalkerDefault.Walk(listener, tree)

	return tempScope, listener.GetErrors()
}

// requireUnitRef asserts that the definition at index i is a UnitReferenceDefinition
// with the expected name.
func requireUnitRef(t *testing.T, defs []symboltable.SymbolDefinition, i int, expectedName string) {
	t.Helper()
	if i >= len(defs) {
		t.Fatalf("expected at least %d definitions, got %d", i+1, len(defs))
	}
	def := defs[i]
	if def.DefinitionType != symboltable.UnitReferenceDefinition {
		t.Fatalf("expected definition %d to be UnitReferenceDefinition, got %s", i, def.DefinitionType)
	}
	unitRef, ok := def.Data.(ParserUnitReference)
	if !ok {
		t.Fatalf("expected definition %d data to be ParserUnitReference, got %T", i, def.Data)
	}
	if unitRef.Name != expectedName {
		t.Fatalf("expected unit reference name %q at %d, got %q", expectedName, i, unitRef.Name)
	}
}

// requireNumber asserts that the definition at index i is a NumberDefinition
// with the expected value.
func requireNumber(t *testing.T, defs []symboltable.SymbolDefinition, i int, expectedValue float64) {
	t.Helper()
	if i >= len(defs) {
		t.Fatalf("expected at least %d definitions, got %d", i+1, len(defs))
	}
	def := defs[i]
	if def.DefinitionType != symboltable.NumberDefinition {
		t.Fatalf("expected definition %d to be NumberDefinition, got %s", i, def.DefinitionType)
	}
	floatData, ok := def.Data.(symboltable.FloatData)
	if !ok {
		t.Fatalf("expected definition %d data to be FloatData, got %T", i, def.Data)
	}
	if floatData.Value != expectedValue {
		t.Fatalf("expected number value %v at %d, got %v", expectedValue, i, floatData.Value)
	}
}

// requireOperator asserts that the definition at index i is an operator definition
// with the expected OperatorType.
func requireOperator(t *testing.T, defs []symboltable.SymbolDefinition, i int, expectedOp OperatorType) {
	t.Helper()
	if i >= len(defs) {
		t.Fatalf("expected at least %d definitions, got %d", i+1, len(defs))
	}
	def := defs[i]
	op, ok := def.Data.(ParserOperator)
	if !ok {
		t.Fatalf("expected definition %d data to be ParserOperator, got %T", i, def.Data)
	}
	if op.Operator != expectedOp {
		t.Fatalf("expected operator %d at %d, got %d", expectedOp, i, op.Operator)
	}
}

func TestPostfixOrder_Multiplication(t *testing.T) {
	scope, errs := parseArithmeticExpr("a*b")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	defs := scope.Definitions
	if len(defs) != 3 {
		t.Fatalf("expected 3 definitions, got %d", len(defs))
	}
	requireUnitRef(t, defs, 0, "a")
	requireUnitRef(t, defs, 1, "b")
	requireOperator(t, defs, 2, OpMultiply)
}

func TestPostfixOrder_Division(t *testing.T) {
	scope, errs := parseArithmeticExpr("a/b")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	defs := scope.Definitions
	if len(defs) != 3 {
		t.Fatalf("expected 3 definitions, got %d", len(defs))
	}
	requireUnitRef(t, defs, 0, "a")
	requireUnitRef(t, defs, 1, "b")
	requireOperator(t, defs, 2, OpDivide)
}

func TestPostfixOrder_Power(t *testing.T) {
	scope, errs := parseArithmeticExpr("a^2")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	defs := scope.Definitions
	if len(defs) != 3 {
		t.Fatalf("expected 3 definitions, got %d", len(defs))
	}
	requireUnitRef(t, defs, 0, "a")
	requireNumber(t, defs, 1, 2)
	requireOperator(t, defs, 2, OpPower)
}

func TestPostfixOrder_Addition(t *testing.T) {
	scope, errs := parseArithmeticExpr("a+b")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	defs := scope.Definitions
	if len(defs) != 3 {
		t.Fatalf("expected 3 definitions, got %d", len(defs))
	}
	requireUnitRef(t, defs, 0, "a")
	requireUnitRef(t, defs, 1, "b")
	requireOperator(t, defs, 2, OpPlus)
}

func TestPostfixOrder_Parenthesized(t *testing.T) {
	scope, errs := parseArithmeticExpr("(a+b)*c")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	defs := scope.Definitions
	if len(defs) != 5 {
		t.Fatalf("expected 5 definitions, got %d", len(defs))
	}
	requireUnitRef(t, defs, 0, "a")
	requireUnitRef(t, defs, 1, "b")
	requireOperator(t, defs, 2, OpPlus)
	requireUnitRef(t, defs, 3, "c")
	requireOperator(t, defs, 4, OpMultiply)
}
