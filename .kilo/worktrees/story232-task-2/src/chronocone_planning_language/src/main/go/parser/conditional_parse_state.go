// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

// Package parser provides CPL parsing functionality.
//
// This file contains the ConditionalParseState type, which encapsulates the
// state machine for parsing conditional expressions (ifBranch, elseIfBranch,
// whenClause). It tracks transitions between the condition expression (parsed
// in a parent scope) and the body expression (parsed in a child scope), and
// captures the condition's postfix definitions for later storage as ClauseData.
package parser

import "github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"

// ConditionalParseState tracks parsing state for a single conditional branch.
//
// During conditional parsing, the parser switches between:
//   - Condition expression: evaluated in a TempScope pushed at branch entry
//   - Body expression: evaluated in a child body scope (IfBodyScope, etc.)
//
// Because ANTLR walks children depth-first with no callback between the two
// expressions, this type tracks which expression is being parsed and saves the
// condition's postfix definitions for later storage as ClauseData.
//
// The type also guards against conflating nested parenthesized sub-expressions
// (which push their own child TempScopes) with the outermost condition-to-body
// transition by comparing the current scope against the stored conditionScope.
type ConditionalParseState struct {
	// expressionCount tracks which expression within the current branch is being parsed.
	// 0 = condition expression (evaluated in parent scope)
	// 1 = body expression (evaluated in child scope)
	// 2 = body expression complete
	expressionCount int

	// savedConditionPostfix holds the condition's postfix definitions captured from
	// the condition scope before switching to the body scope.
	savedConditionPostfix []symboltable.SymbolDefinition

	// bodyScopeType specifies the scope type to push for the body expression
	// when transitioning from condition to body.
	bodyScopeType symboltable.ScopeType

	// conditionScope is the symbol table pushed at branch entry to accumulate
	// the condition expression's postfix definitions. It is used to distinguish
	// the outermost condition expression from nested parenthesized sub-expressions
	// (which push their own child scopes).
	conditionScope *symboltable.SymbolTable
}

// NewConditionalParseState creates a new ConditionalParseState for the given
// body scope type and condition scope. The state begins in the condition phase
// (expressionCount == 0).
func NewConditionalParseState(bodyScopeType symboltable.ScopeType, conditionScope *symboltable.SymbolTable) *ConditionalParseState {
	return &ConditionalParseState{
		expressionCount: 0,
		bodyScopeType:   bodyScopeType,
		conditionScope:  conditionScope,
	}
}

// EnterCondition resets the state to the condition phase, clearing any
// previously captured postfix definitions.
// This is typically called when entering a new conditional branch.
//
// NOTE: EnterCondition is currently only exercised by unit tests and is
// reserved for future state reuse (e.g., resetting a pooled instance for
// a subsequent branch). It is not called by production parsing code today.
func (s *ConditionalParseState) EnterCondition() {
	s.expressionCount = 0
	s.savedConditionPostfix = nil
}

// EnterBody transitions the state from condition (0) to body (1) and stores
// the condition's postfix definitions. It returns true if the transition
// occurred, or false if the state was not in the condition phase.
func (s *ConditionalParseState) EnterBody(conditionPostfix []symboltable.SymbolDefinition) bool {
	if s.expressionCount != 0 {
		return false
	}
	s.savedConditionPostfix = conditionPostfix
	s.expressionCount = 1
	return true
}

// IsComplete returns true when the body expression has finished (state == 2).
func (s *ConditionalParseState) IsComplete() bool {
	return s.expressionCount == 2
}

// GetConditionPostfix returns the condition's postfix definitions captured
// during the condition-to-body transition.
func (s *ConditionalParseState) GetConditionPostfix() []symboltable.SymbolDefinition {
	return s.savedConditionPostfix
}

// GetBodyScopeType returns the scope type that should be pushed for the body
// expression (e.g., IfBodyScope, ElseIfBodyScope, WhenBodyScope).
func (s *ConditionalParseState) GetBodyScopeType() symboltable.ScopeType {
	return s.bodyScopeType
}

// GetConditionScope returns the symbol table that was pushed for the condition
// expression. This is used to distinguish the outermost condition expression
// from nested parenthesized sub-expressions.
func (s *ConditionalParseState) GetConditionScope() *symboltable.SymbolTable {
	return s.conditionScope
}

// IsInCondition returns true when the state is in the condition phase (0).
func (s *ConditionalParseState) IsInCondition() bool {
	return s.expressionCount == 0
}

// IsInBody returns true when the state is in the body phase (1).
func (s *ConditionalParseState) IsInBody() bool {
	return s.expressionCount == 1
}

// MarkBodyComplete transitions the state from body (1) to complete (2).
func (s *ConditionalParseState) MarkBodyComplete() {
	if s.expressionCount == 1 {
		s.expressionCount = 2
	}
}
