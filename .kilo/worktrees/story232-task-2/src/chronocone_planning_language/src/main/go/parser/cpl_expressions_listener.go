// Package parser provides CPL parsing functionality.
//
// This file contains listener callbacks for expression rules defined in
// the root grammar ChronoconePlanningLanguage.g4:
//
//   - primaryExpr: handles parenthesized sub-expressions via child TempScope
//     and pushes Identifier definitions for identifier references
//   - expression: detects condition→body and body→complete transitions
//     during conditional parsing
//   - additionExpr / multiplicationExpr / unaryExpr / relationalExpr / equalityExpr /
//     logicalAndExpr / notExpr / logicalOrExpr: push the corresponding operator
//     definitions onto the current expression-accumulating scope in postfix order
//
// These listener methods follow the symbol table-driven parsing pattern:
// expression operators and operands are accumulated as ordered definitions
// in a scope, producing a postfix sequence that is consumed by the analysis phase.
package parser

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// EnterPrimaryExpr is called when entering a primary expression.
// For parenthesized expressions (LPAREN expression RPAREN), pushes a child
// TempScope to isolate the sub-expression's postfix definitions. This gives
// the symbol table stack a depth that tracks expression nesting, which
// ExitExpression uses to detect when the outermost condition or body
// expression has completed (by comparing the current scope against the
// scope that was pushed at branch entry).
func (l *CPLListener) EnterPrimaryExpr(ctx *antlr.PrimaryExprContext) {
	if ctx.LPAREN() == nil {
		return // Only parenthesized expressions push a child scope
	}
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering parenthesized expr: %w", err))
		return
	}
	l.symbolStack.Push(current.AddChild(symboltable.TempScope))
}

// ExitPrimaryExpr is called when exiting a primary expression.
// If this is a parenthesized expression, pops the child TempScope pushed
// by EnterPrimaryExpr and merges its definitions back into the parent scope
// so the sub-expression's postfix operators and operands are included in
// the surrounding expression's postfix sequence.
// If the current scope is an expression-accumulating scope and the primaryExpr
// contains an identifier, pushes an Identifier definition.
func (l *CPLListener) ExitPrimaryExpr(ctx *antlr.PrimaryExprContext) {
	// For parenthesized expressions: pop the child scope and merge defs back
	if ctx.LPAREN() != nil {
		childScope, err := l.symbolStack.Pop()
		if err != nil {
			l.AddError(fmt.Errorf("symbol stack error popping parenthesized expr scope: %w", err))
			return
		}
		parentScope, err := l.symbolStack.Peek()
		if err != nil {
			l.AddError(fmt.Errorf("symbol stack error peeking after parenthesized expr: %w", err))
			return
		}
		// Merge child scope definitions into parent, maintaining postfix order
		parentScope.Definitions = append(parentScope.Definitions, childScope.Definitions...)
		l.logExit("PrimaryExpr (parenthesized)", &ctx.BaseParserRuleContext)
		return
	}

	if !l.shouldPushExpressionDefs() {
		return
	}

	// If this primaryExpr contains an identifier, push it as an Identifier definition
	if identifier := ctx.Identifier(); identifier != nil {
		identifierText := identifier.GetText()
		currentScope, err := l.symbolStack.Peek()
		if err != nil {
			l.AddError(fmt.Errorf("symbol stack error in ExitPrimaryExpr: %w", err))
			return
		}
		err = currentScope.DefineWithData(
			nil,
			symboltable.Identifier,
			l.makeSourceLocation(identifier.GetStart()),
			symboltable.StringData{Value: identifierText},
		)
		if err != nil {
			l.AddError(fmt.Errorf("error defining identifier in ExitPrimaryExpr: %w", err))
		}
	}
}

// ExitAdditionExpr is called when exiting an addition expression.
// If the current scope is MetricScope, pushes a PlusOperatorDefinition
// or MinusOperatorDefinition based on the operator token present.
func (l *CPLListener) ExitAdditionExpr(ctx *antlr.AdditionExprContext) {
	if !l.shouldPushExpressionDefs() {
		return
	}

	currentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error in ExitAdditionExpr: %w", err))
		return
	}

	var defType symboltable.DefinitionType
	switch {
	case ctx.PLUS() != nil:
		defType = symboltable.PlusOperatorDefinition
	case ctx.MINUS() != nil:
		defType = symboltable.MinusOperatorDefinition
	default:
		return
	}

	err = currentScope.DefineWithData(
		nil,
		defType,
		l.makeSourceLocation(ctx.GetStart()),
		nil,
	)
	if err != nil {
		l.AddError(fmt.Errorf("error defining operator in ExitAdditionExpr: %w", err))
	}
}

// ExitMultiplicationExpr is called when exiting a multiplication expression.
// If the current scope is MetricScope, pushes a MultiplyOperatorDefinition
// or DivideOperatorDefinition based on the operator token present.
func (l *CPLListener) ExitMultiplicationExpr(ctx *antlr.MultiplicationExprContext) {
	if !l.shouldPushExpressionDefs() {
		return
	}

	currentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error in ExitMultiplicationExpr: %w", err))
		return
	}

	var defType symboltable.DefinitionType
	switch {
	case ctx.MULTIPLY() != nil:
		defType = symboltable.MultiplyOperatorDefinition
	case ctx.DIVIDE() != nil:
		defType = symboltable.DivideOperatorDefinition
	default:
		return
	}

	err = currentScope.DefineWithData(
		nil,
		defType,
		l.makeSourceLocation(ctx.GetStart()),
		nil,
	)
	if err != nil {
		l.AddError(fmt.Errorf("error defining operator in ExitMultiplicationExpr: %w", err))
	}
}

// ExitUnaryExpr is called when exiting a unary expression.
// If the current scope is an expression-accumulating scope and a MINUS token
// is present, pushes a NegateOperatorDefinition. Logical NOT is handled in
// ExitNotExpr (lower precedence than comparisons).
func (l *CPLListener) ExitUnaryExpr(ctx *antlr.UnaryExprContext) {
	if !l.shouldPushExpressionDefs() {
		return
	}

	currentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error in ExitUnaryExpr: %w", err))
		return
	}

	if ctx.MINUS() != nil {
		err = currentScope.DefineWithData(
			nil,
			symboltable.NegateOperatorDefinition,
			l.makeSourceLocation(ctx.GetStart()),
			nil,
		)
		if err != nil {
			l.AddError(fmt.Errorf("error defining negate operator in ExitUnaryExpr: %w", err))
		}
	}
}

// ExitExpression is called when exiting an expression production.
//
// During conditional parsing (when conditionalStates is non-empty), this handler
// detects the transition from condition expression to body expression:
//
//   - condition → body: The condition expression just finished.
//     Pop the TempScope, save its Definitions as condition postfix,
//     push the body scope (type determined by state.GetBodyScopeType()) for
//     the body expression.
//
//   - body → complete: The body expression just finished.
//     No scope manipulation needed — the branch's exit handler will pop
//     the body scope and define the clause definition.
func (l *CPLListener) ExitExpression(ctx *antlr.ExpressionContext) {
	if len(l.conditionalStates) == 0 {
		return
	}

	state := l.conditionalStates[len(l.conditionalStates)-1]

	if state.IsInCondition() {
		// Only transition when the condition scope is on top of the symbol table
		// stack. Parenthesized sub-expressions (e.g., the "(data_a > data_b)" in
		// "not (data_a > data_b)") push their own child scopes, so when those fire
		// ExitExpression the condition scope is not on top yet.
		currentScope, err := l.symbolStack.Peek()
		if err != nil {
			l.AddError(fmt.Errorf("symbol stack error peeking in ExitExpression: %w", err))
			return
		}
		if currentScope != state.GetConditionScope() {
			return
		}
		// Condition expression just finished. Pop TempScope and save its
		// definitions as the condition postfix.
		tempScope, err := l.symbolStack.Pop()
		if err != nil {
			l.AddError(fmt.Errorf("symbol stack error popping TempScope in ExitExpression: %w", err))
			return
		}
		state.EnterBody(tempScope.Definitions)

		// Push body scope as a child of the parent scope for the body expression.
		parentScope, err := l.symbolStack.Peek()
		if err != nil {
			l.AddError(fmt.Errorf("symbol stack error peeking parent in ExitExpression: %w", err))
			return
		}
		l.symbolStack.Push(parentScope.AddChild(state.GetBodyScopeType()))
	} else if state.IsInBody() {
		// Body expression just finished. No scope action needed; the branch's
		// exit handler will pop the body scope and define the clause definition.
		state.MarkBodyComplete()
	}
}

// ExitRelationalExpr is called when exiting a relational expression.
// Pushes the appropriate comparison operator definition when in an
// expression-accumulating scope.
func (l *CPLListener) ExitRelationalExpr(ctx *antlr.RelationalExprContext) {
	if !l.shouldPushExpressionDefs() {
		return
	}

	currentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error in ExitRelationalExpr: %w", err))
		return
	}

	var defType symboltable.DefinitionType
	switch {
	case ctx.LT(0) != nil:
		defType = symboltable.LessThanOperatorDefinition
	case ctx.GT(0) != nil:
		defType = symboltable.GreaterThanOperatorDefinition
	case ctx.LTEQ(0) != nil:
		defType = symboltable.LessThanEqOperatorDefinition
	case ctx.GTEQ(0) != nil:
		defType = symboltable.GreaterThanEqOperatorDefinition
	default:
		return
	}

	err = currentScope.DefineWithData(
		nil,
		defType,
		l.makeSourceLocation(ctx.GetStart()),
		nil,
	)
	if err != nil {
		l.AddError(fmt.Errorf("error defining relational operator: %w", err))
	}
}

// ExitEqualityExpr is called when exiting an equality expression.
// Pushes equality or inequality operator definitions when in an
// expression-accumulating scope.
func (l *CPLListener) ExitEqualityExpr(ctx *antlr.EqualityExprContext) {
	if !l.shouldPushExpressionDefs() {
		return
	}

	currentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error in ExitEqualityExpr: %w", err))
		return
	}

	var defType symboltable.DefinitionType
	switch {
	case ctx.EQEQ(0) != nil:
		defType = symboltable.EqOperatorDefinition
	case ctx.NOTEQ(0) != nil:
		defType = symboltable.NotEqOperatorDefinition
	default:
		return
	}

	err = currentScope.DefineWithData(
		nil,
		defType,
		l.makeSourceLocation(ctx.GetStart()),
		nil,
	)
	if err != nil {
		l.AddError(fmt.Errorf("error defining equality operator: %w", err))
	}
}

// ExitLogicalAndExpr is called when exiting a logical AND expression.
// Pushes an AndOperatorDefinition when in an expression-accumulating scope.
func (l *CPLListener) ExitLogicalAndExpr(ctx *antlr.LogicalAndExprContext) {
	if !l.shouldPushExpressionDefs() {
		return
	}

	if ctx.AND_KEYWORD(0) == nil {
		return
	}

	currentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error in ExitLogicalAndExpr: %w", err))
		return
	}

	err = currentScope.DefineWithData(
		nil,
		symboltable.AndOperatorDefinition,
		l.makeSourceLocation(ctx.GetStart()),
		nil,
	)
	if err != nil {
		l.AddError(fmt.Errorf("error defining and operator: %w", err))
	}
}

// ExitNotExpr is called when exiting a logical NOT expression.
// Pushes a NotOperatorDefinition when NOT_KEYWORD is present.
// When NOT_KEYWORD is absent, this is just a fall-through to equalityExpr
// and no operator is pushed.
func (l *CPLListener) ExitNotExpr(ctx *antlr.NotExprContext) {
	if !l.shouldPushExpressionDefs() {
		return
	}

	if ctx.NOT_KEYWORD() == nil {
		return
	}

	currentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error in ExitNotExpr: %w", err))
		return
	}

	err = currentScope.DefineWithData(
		nil,
		symboltable.NotOperatorDefinition,
		l.makeSourceLocation(ctx.GetStart()),
		nil,
	)
	if err != nil {
		l.AddError(fmt.Errorf("error defining not operator in ExitNotExpr: %w", err))
	}
}

// ExitLogicalOrExpr is called when exiting a logical OR expression.
// Pushes an OrOperatorDefinition when in an expression-accumulating scope.
func (l *CPLListener) ExitLogicalOrExpr(ctx *antlr.LogicalOrExprContext) {
	if !l.shouldPushExpressionDefs() {
		return
	}

	if ctx.OR_KEYWORD(0) == nil {
		return
	}

	currentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error in ExitLogicalOrExpr: %w", err))
		return
	}

	err = currentScope.DefineWithData(
		nil,
		symboltable.OrOperatorDefinition,
		l.makeSourceLocation(ctx.GetStart()),
		nil,
	)
	if err != nil {
		l.AddError(fmt.Errorf("error defining or operator: %w", err))
	}
}
