// Package parser provides CPL parsing functionality.
//
// This file contains listener callbacks for conditional expression rules
// defined in the root grammar ChronoconePlanningLanguage.g4:
//
//   - ifBranch: pushes an IfBodyScope child scope, captures condition postfix
//     as ClauseData on an IfClauseDefinition in the parent scope
//   - elseIfBranch: pushes an ElseIfBodyScope child scope, captures condition postfix
//     as ClauseData on an ElseIfClauseDefinition in the parent scope
//   - elseBranch: pushes an ElseBodyScope child scope (no condition)
//   - whenClause: pushes a WhenBodyScope child scope, captures condition postfix
//     as ClauseData on a WhenClauseDefinition in the parent scope
//   - defaultClause: pushes a DefaultBodyScope child scope (no condition)
//
// These listener methods follow the symbol table-driven parsing pattern:
// each branch body gets its own child scope for proper scope isolation.
// Conditions use postfix in the parent scope; body expressions use postfix
// in the child scope.
package parser

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// EnterIfBranch is called when entering an ifBranch production.
//
// The grammar rule is:
//
//	ifBranch
//	    : expression          // Condition — in parent scope
//	      indent expression   // Body — in child IfBodyScope
//	    ;
//
// During parsing, the condition expression is evaluated in the parent scope
// (where definitions accumulate as postfix). After the condition exits, the
// listener switches to a child IfBodyScope for the body expression.
func (l *CPLListener) EnterIfBranch(ctx *antlr.IfBranchContext) {
	l.logEnter("IfBranch", &ctx.BaseParserRuleContext)

	// Push a TempScope as a child of the current scope to isolate the
	// condition expression's postfix definitions. These will be captured
	// when ExitExpression fires (expressionCount 0→1), then stored as
	// ClauseData on an IfClauseDefinition by ExitIfBranch.
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering if branch: %w", err))
		return
	}
	conditionScope := current.AddChild(symboltable.TempScope)
	l.symbolStack.Push(conditionScope)

	// Initialize conditional parse state to track which expression
	// (condition vs body) is being parsed. expressionCount 0 = condition.
	// bodyScopeType tells ExitExpression which child scope type to push
	// when transitioning from condition to body expression.
	// conditionScope is stored so ExitExpression can distinguish the
	// outermost condition expression from nested parenthesized sub-expressions
	// (which push their own child scopes).
	l.conditionalStates = append(l.conditionalStates, NewConditionalParseState(
		symboltable.IfBodyScope,
		conditionScope,
	))
}

// ExitIfBranch is called when exiting an ifBranch production.
//
// On exit, the listener:
//  1. Pops the IfBodyScope (child scope for the body expression)
//  2. Extracts the condition's postfix definitions (saved from parent scope)
//  3. Defines an IfClauseDefinition in the parent scope with ClauseData
//     containing the condition's postfix and stores the body scope reference
//     in ClauseData.BodyScope so the analysis phase can locate the body
//     expression's postfix definitions.
func (l *CPLListener) ExitIfBranch(ctx *antlr.IfBranchContext) {
	// Pop IfBodyScope (pushed by ExitExpression during condition→body switch).
	// The body expression's postfix definitions have accumulated in this scope.
	bodyScope, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting if branch (pop IfBodyScope): %w", err))
		if len(l.conditionalStates) > 0 {
			l.conditionalStates = l.conditionalStates[:len(l.conditionalStates)-1]
		}
		l.logExit("IfBranch", &ctx.BaseParserRuleContext)
		return
	}

	// Peek the parent scope (MetricScope or another conditional body scope)
	parentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting if branch (peek parent): %w", err))
		if len(l.conditionalStates) > 0 {
			l.conditionalStates = l.conditionalStates[:len(l.conditionalStates)-1]
		}
		l.logExit("IfBranch", &ctx.BaseParserRuleContext)
		return
	}

	// Retrieve and pop the conditional state for this if branch
	if len(l.conditionalStates) == 0 {
		l.AddError(fmt.Errorf("no conditional state when exiting if branch"))
		l.logExit("IfBranch", &ctx.BaseParserRuleContext)
		return
	}
	state := l.conditionalStates[len(l.conditionalStates)-1]
	l.conditionalStates = l.conditionalStates[:len(l.conditionalStates)-1]

	// Define an IfClauseDefinition in the parent scope with ClauseData
	// containing the condition's postfix definitions (saved by ExitExpression)
	// and the body scope reference so the analysis phase can locate the body
	// expression's postfix definitions.
	clauseData := symboltable.ClauseData{
		ConditionPostfix: state.GetConditionPostfix(),
		BodyScope:        bodyScope,
	}
	err = parentScope.DefineWithData(
		nil,
		symboltable.IfClauseDefinition,
		symboltable.SourceLocation{},
		clauseData,
	)
	if err != nil {
		l.AddError(fmt.Errorf("error defining IfClauseDefinition: %w", err))
	}

	l.logExit("IfBranch", &ctx.BaseParserRuleContext)
}

// EnterElseIfBranch is called when entering an elseIfBranch production.
//
// The grammar rule is:
//
//	elseIfBranch
//	    : ELSE_KEYWORD IF_KEYWORD expression
//	      indent expression   // Body — in child ElseIfBodyScope
//	    ;
//
// The pattern is analogous to EnterIfBranch: pushes a TempScope for condition
// postfix isolation, initializes ConditionalParseState with bodyScopeType
// set to ElseIfBodyScope.
func (l *CPLListener) EnterElseIfBranch(ctx *antlr.ElseIfBranchContext) {
	l.logEnter("ElseIfBranch", &ctx.BaseParserRuleContext)

	// Push a TempScope as a child of the current scope to isolate the
	// condition expression's postfix definitions. These will be captured
	// when ExitExpression fires (expressionCount 0→1), then stored as
	// ClauseData on an ElseIfClauseDefinition by ExitElseIfBranch.
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering else-if branch: %w", err))
		return
	}
	conditionScope := current.AddChild(symboltable.TempScope)
	l.symbolStack.Push(conditionScope)

	// Initialize conditional parse state to track which expression
	// (condition vs body) is being parsed. expressionCount 0 = condition.
	// bodyScopeType tells ExitExpression which child scope type to push
	// when transitioning from condition to body expression.
	// conditionScope is stored so ExitExpression can distinguish the
	// outermost condition expression from nested parenthesized sub-expressions
	// (which push their own child scopes).
	l.conditionalStates = append(l.conditionalStates, NewConditionalParseState(
		symboltable.ElseIfBodyScope,
		conditionScope,
	))
}

// ExitElseIfBranch is called when exiting an elseIfBranch production.
//
// On exit, the listener:
//  1. Pops the ElseIfBodyScope (child scope for the body expression)
//  2. Extracts the condition's postfix definitions (saved from parent scope)
//  3. Defines an ElseIfClauseDefinition in the parent scope with ClauseData
//     containing the condition's postfix and stores the body scope reference
//     in ClauseData.BodyScope so the analysis phase can locate the body
//     expression's postfix definitions.
func (l *CPLListener) ExitElseIfBranch(ctx *antlr.ElseIfBranchContext) {
	// Pop ElseIfBodyScope (pushed by ExitExpression during condition→body switch).
	// The body expression's postfix definitions have accumulated in this scope.
	bodyScope, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting else-if branch (pop ElseIfBodyScope): %w", err))
		if len(l.conditionalStates) > 0 {
			l.conditionalStates = l.conditionalStates[:len(l.conditionalStates)-1]
		}
		l.logExit("ElseIfBranch", &ctx.BaseParserRuleContext)
		return
	}

	// Peek the parent scope (MetricScope or another conditional body scope)
	parentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting else-if branch (peek parent): %w", err))
		if len(l.conditionalStates) > 0 {
			l.conditionalStates = l.conditionalStates[:len(l.conditionalStates)-1]
		}
		l.logExit("ElseIfBranch", &ctx.BaseParserRuleContext)
		return
	}

	// Retrieve and pop the conditional state for this else-if branch
	if len(l.conditionalStates) == 0 {
		l.AddError(fmt.Errorf("no conditional state when exiting else-if branch"))
		l.logExit("ElseIfBranch", &ctx.BaseParserRuleContext)
		return
	}
	state := l.conditionalStates[len(l.conditionalStates)-1]
	l.conditionalStates = l.conditionalStates[:len(l.conditionalStates)-1]

	// Define an ElseIfClauseDefinition in the parent scope with ClauseData
	// containing the condition's postfix definitions (saved by ExitExpression)
	// and the body scope reference so the analysis phase can locate the body
	// expression's postfix definitions.
	clauseData := symboltable.ClauseData{
		ConditionPostfix: state.GetConditionPostfix(),
		BodyScope:        bodyScope,
	}
	err = parentScope.DefineWithData(
		nil,
		symboltable.ElseIfClauseDefinition,
		symboltable.SourceLocation{},
		clauseData,
	)
	if err != nil {
		l.AddError(fmt.Errorf("error defining ElseIfClauseDefinition: %w", err))
	}

	l.logExit("ElseIfBranch", &ctx.BaseParserRuleContext)
}

// EnterElseBranch is called when entering an elseBranch production.
//
// The grammar rule is:
//
//	elseBranch
//	    : ELSE_KEYWORD
//	      indent expression   // Body — in child ElseBodyScope
//	    ;
//
// Unlike if/else-if branches, the else-branch has NO condition expression.
// The listener pushes an ElseBodyScope directly as a child of the current
// scope for the body expression. No TempScope or ConditionalParseState is
// needed since there is no condition to isolate.
func (l *CPLListener) EnterElseBranch(ctx *antlr.ElseBranchContext) {
	l.logEnter("ElseBranch", &ctx.BaseParserRuleContext)

	// Push an ElseBodyScope as a child of the current scope for the body
	// expression. No TempScope needed (no condition to isolate).
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering else branch: %w", err))
		return
	}
	l.symbolStack.Push(current.AddChild(symboltable.ElseBodyScope))

	// No ConditionalParseState push — else-branch has no condition, so
	// ExitExpression will not see a state for this branch. The body
	// expression accumulates postfix in the ElseBodyScope directly.
}

// ExitElseBranch is called when exiting an elseBranch production.
//
// On exit, the listener:
//  1. Pops the ElseBodyScope (child scope for the body expression)
//  2. Defines an ElseClauseDefinition in the parent scope with ClauseData
//     containing only BodyScope (no ConditionPostfix — else has no condition)
func (l *CPLListener) ExitElseBranch(ctx *antlr.ElseBranchContext) {
	// Pop ElseBodyScope (pushed by EnterElseBranch).
	// The body expression's postfix definitions have accumulated in this scope.
	bodyScope, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting else branch (pop ElseBodyScope): %w", err))
		l.logExit("ElseBranch", &ctx.BaseParserRuleContext)
		return
	}

	// Peek the parent scope (MetricScope or another conditional body scope)
	parentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting else branch (peek parent): %w", err))
		l.logExit("ElseBranch", &ctx.BaseParserRuleContext)
		return
	}

	// Define an ElseClauseDefinition in the parent scope with ClauseData
	// containing only the body scope reference. No ConditionPostfix since
	// else-branch has no condition.
	clauseData := symboltable.ClauseData{
		BodyScope: bodyScope,
	}
	err = parentScope.DefineWithData(
		nil,
		symboltable.ElseClauseDefinition,
		symboltable.SourceLocation{},
		clauseData,
	)
	if err != nil {
		l.AddError(fmt.Errorf("error defining ElseClauseDefinition: %w", err))
	}

	l.logExit("ElseBranch", &ctx.BaseParserRuleContext)
}

// ================================================================================
// WHEN-CLAUSE AND DEFAULT-CLAUSE LISTENER METHODS (Task 2f)
// ================================================================================

// EnterWhenClause is called when entering a whenClause production.
//
// The grammar rule is:
//
//	whenClause
//	    : expression ARROW     // Condition — in parent scope
//	      indent expression    // Body — in child WhenBodyScope
//	    ;
//
// The pattern is analogous to EnterIfBranch: pushes a TempScope for condition
// postfix isolation, initializes ConditionalParseState with bodyScopeType
// set to WhenBodyScope.
func (l *CPLListener) EnterWhenClause(ctx *antlr.WhenClauseContext) {
	l.logEnter("WhenClause", &ctx.BaseParserRuleContext)

	// Push a TempScope as a child of the current scope to isolate the
	// condition expression's postfix definitions. These will be captured
	// when ExitExpression fires (expressionCount 0→1), then stored as
	// ClauseData on a WhenClauseDefinition by ExitWhenClause.
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering when clause: %w", err))
		return
	}
	conditionScope := current.AddChild(symboltable.TempScope)
	l.symbolStack.Push(conditionScope)

	// Initialize conditional parse state to track which expression
	// (condition vs body) is being parsed. expressionCount 0 = condition.
	// bodyScopeType tells ExitExpression which child scope type to push
	// when transitioning from condition to body expression.
	// conditionScope is stored so ExitExpression can distinguish the
	// outermost condition expression from nested parenthesized sub-expressions
	// (which push their own child scopes).
	l.conditionalStates = append(l.conditionalStates, NewConditionalParseState(
		symboltable.WhenBodyScope,
		conditionScope,
	))
}

// ExitWhenClause is called when exiting a whenClause production.
//
// On exit, the listener:
//  1. Pops the WhenBodyScope (child scope for the body expression)
//  2. Extracts the condition's postfix definitions (saved from parent scope)
//  3. Defines a WhenClauseDefinition in the parent scope with ClauseData
//     containing the condition's postfix and stores the body scope reference
//     in ClauseData.BodyScope so the analysis phase can locate the body
//     expression's postfix definitions.
func (l *CPLListener) ExitWhenClause(ctx *antlr.WhenClauseContext) {
	// Pop WhenBodyScope (pushed by ExitExpression during condition→body switch).
	// The body expression's postfix definitions have accumulated in this scope.
	bodyScope, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting when clause (pop WhenBodyScope): %w", err))
		if len(l.conditionalStates) > 0 {
			l.conditionalStates = l.conditionalStates[:len(l.conditionalStates)-1]
		}
		l.logExit("WhenClause", &ctx.BaseParserRuleContext)
		return
	}

	// Peek the parent scope (MetricScope or another conditional body scope)
	parentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting when clause (peek parent): %w", err))
		if len(l.conditionalStates) > 0 {
			l.conditionalStates = l.conditionalStates[:len(l.conditionalStates)-1]
		}
		l.logExit("WhenClause", &ctx.BaseParserRuleContext)
		return
	}

	// Retrieve and pop the conditional state for this when clause
	if len(l.conditionalStates) == 0 {
		l.AddError(fmt.Errorf("no conditional state when exiting when clause"))
		l.logExit("WhenClause", &ctx.BaseParserRuleContext)
		return
	}
	state := l.conditionalStates[len(l.conditionalStates)-1]
	l.conditionalStates = l.conditionalStates[:len(l.conditionalStates)-1]

	// Define a WhenClauseDefinition in the parent scope with ClauseData
	// containing the condition's postfix definitions (saved by ExitExpression)
	// and the body scope reference so the analysis phase can locate the body
	// expression's postfix definitions.
	clauseData := symboltable.ClauseData{
		ConditionPostfix: state.GetConditionPostfix(),
		BodyScope:        bodyScope,
	}
	err = parentScope.DefineWithData(
		nil,
		symboltable.WhenClauseDefinition,
		symboltable.SourceLocation{},
		clauseData,
	)
	if err != nil {
		l.AddError(fmt.Errorf("error defining WhenClauseDefinition: %w", err))
	}

	l.logExit("WhenClause", &ctx.BaseParserRuleContext)
}

// EnterDefaultClause is called when entering a defaultClause production.
//
// The grammar rule is:
//
//	defaultClause
//	    : DEFAULT_KEYWORD ARROW
//	      indent expression   // Body — in child DefaultBodyScope
//	    ;
//
// Unlike when-clauses, the default-clause has NO condition expression.
// The listener pushes a DefaultBodyScope directly as a child of the current
// scope for the body expression. No TempScope or ConditionalParseState is
// needed since there is no condition to isolate.
func (l *CPLListener) EnterDefaultClause(ctx *antlr.DefaultClauseContext) {
	l.logEnter("DefaultClause", &ctx.BaseParserRuleContext)

	// Push a DefaultBodyScope as a child of the current scope for the body
	// expression. No TempScope needed (no condition to isolate).
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering default clause: %w", err))
		return
	}
	l.symbolStack.Push(current.AddChild(symboltable.DefaultBodyScope))

	// No ConditionalParseState push — default-clause has no condition, so
	// ExitExpression will not see a state for this clause. The body
	// expression accumulates postfix in the DefaultBodyScope directly.
}

// ExitDefaultClause is called when exiting a defaultClause production.
//
// On exit, the listener:
//  1. Pops the DefaultBodyScope (child scope for the body expression)
//  2. Defines a DefaultClauseDefinition in the parent scope with ClauseData
//     containing only BodyScope (no ConditionPostfix — default has no condition)
func (l *CPLListener) ExitDefaultClause(ctx *antlr.DefaultClauseContext) {
	// Pop DefaultBodyScope (pushed by EnterDefaultClause).
	// The body expression's postfix definitions have accumulated in this scope.
	bodyScope, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting default clause (pop DefaultBodyScope): %w", err))
		l.logExit("DefaultClause", &ctx.BaseParserRuleContext)
		return
	}

	// Peek the parent scope (MetricScope or another conditional body scope)
	parentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting default clause (peek parent): %w", err))
		l.logExit("DefaultClause", &ctx.BaseParserRuleContext)
		return
	}

	// Define a DefaultClauseDefinition in the parent scope with ClauseData
	// containing only the body scope reference. No ConditionPostfix since
	// default-clause has no condition.
	clauseData := symboltable.ClauseData{
		BodyScope: bodyScope,
	}
	err = parentScope.DefineWithData(
		nil,
		symboltable.DefaultClauseDefinition,
		symboltable.SourceLocation{},
		clauseData,
	)
	if err != nil {
		l.AddError(fmt.Errorf("error defining DefaultClauseDefinition: %w", err))
	}

	l.logExit("DefaultClause", &ctx.BaseParserRuleContext)
}
