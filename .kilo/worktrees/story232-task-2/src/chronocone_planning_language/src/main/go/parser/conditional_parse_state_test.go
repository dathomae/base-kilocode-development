// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

// Package parser_test contains unit tests for the ConditionalParseState type.
//
// NOTE: There is no dedicated unit test verifying multiple ConditionalParseState
// instances on the CPLListener.conditionalStates slice (i.e., nested conditionals
// created by if/else-if/else chains). That behavior is covered implicitly by the
// integration tests in cpl_conditional_test.go which parse full conditional
// programs and exercise the slice growth in EnterConditionalLikeBranch and
// pop in ExitConditionalLikeBranch.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
)

// TestConditionalParseStateExpressionCountTransitions verifies the core state
// machine transitions: 0 (condition) → 1 (body) → 2 (complete).
func TestConditionalParseStateExpressionCountTransitions(t *testing.T) {
	t.Parallel()

	conditionScope := symboltable.NewSymbolTable(symboltable.TempScope)
	state := parser.NewConditionalParseState(symboltable.IfBodyScope, conditionScope)

	// Initial state: in condition, not complete
	assert.True(t, state.IsInCondition(), "new state should be in condition phase")
	assert.False(t, state.IsInBody(), "new state should not be in body phase")
	assert.False(t, state.IsComplete(), "new state should not be complete")

	// Transition 0 → 1 (condition → body)
	postfix := []symboltable.SymbolDefinition{
		{DefinitionType: symboltable.Identifier},
		{DefinitionType: symboltable.GreaterThanOperatorDefinition},
	}
	ok := state.EnterBody(postfix)
	assert.True(t, ok, "EnterBody should succeed from condition phase")
	assert.False(t, state.IsInCondition(), "state should no longer be in condition phase")
	assert.True(t, state.IsInBody(), "state should be in body phase")
	assert.False(t, state.IsComplete(), "state should not be complete after EnterBody")

	// Transition 1 → 2 (body → complete)
	state.MarkBodyComplete()
	assert.False(t, state.IsInCondition(), "complete state should not be in condition phase")
	assert.False(t, state.IsInBody(), "complete state should not be in body phase")
	assert.True(t, state.IsComplete(), "state should be complete after MarkBodyComplete")
}

// TestConditionalParseStateEnterConditionResetsState verifies that
// EnterCondition resets the state back to the condition phase.
func TestConditionalParseStateEnterConditionResetsState(t *testing.T) {
	t.Parallel()

	conditionScope := symboltable.NewSymbolTable(symboltable.TempScope)
	state := parser.NewConditionalParseState(symboltable.IfBodyScope, conditionScope)

	// Move through full lifecycle
	state.EnterBody([]symboltable.SymbolDefinition{
		{DefinitionType: symboltable.Identifier},
	})
	state.MarkBodyComplete()
	assert.True(t, state.IsComplete(), "state should be complete")

	// Reset
	state.EnterCondition()
	assert.True(t, state.IsInCondition(), "after EnterCondition, state should be in condition phase")
	assert.False(t, state.IsInBody(), "after EnterCondition, state should not be in body phase")
	assert.False(t, state.IsComplete(), "after EnterCondition, state should not be complete")
	assert.Nil(t, state.GetConditionPostfix(), "EnterCondition should clear saved postfix")
}

// TestConditionalParseStateScopeNesting verifies that the state machine does
// not conflate nested parenthesized sub-expressions (which push their own
// TempScope child scopes) with the outermost condition-to-body transition.
//
// When a parenthesized sub-expression is parsed, EnterPrimaryExpr pushes a
// child TempScope. When ExitExpression fires for that nested expression, the
// current scope is the child TempScope, NOT the conditionScope stored in the
// state. The caller uses GetConditionScope() to detect this mismatch and skip
// the transition.
func TestConditionalParseStateScopeNesting(t *testing.T) {
	t.Parallel()

	conditionScope := symboltable.NewSymbolTable(symboltable.TempScope)
	nestedScope := conditionScope.AddChild(symboltable.TempScope)
	state := parser.NewConditionalParseState(symboltable.IfBodyScope, conditionScope)

	// The conditionScope should match itself
	assert.Equal(t, conditionScope, state.GetConditionScope(),
		"GetConditionScope should return the condition scope")

	// The nested scope should NOT match the condition scope
	assert.NotEqual(t, nestedScope, state.GetConditionScope(),
		"nested scope should not match condition scope")

	// Verify that the caller can distinguish nested scopes from the condition scope
	// by comparing pointers. This is the mechanism that prevents premature
	// condition-to-body transition when parenthesized sub-expressions exit.
	assert.Same(t, conditionScope, state.GetConditionScope(),
		"condition scope should be same pointer")
	assert.NotSame(t, nestedScope, state.GetConditionScope(),
		"nested scope should not be same pointer as condition scope")
}

// TestConditionalParseStateConditionPostfixCapture verifies that EnterBody
// correctly captures and GetConditionPostfix correctly retrieves the condition's
// postfix definitions.
func TestConditionalParseStateConditionPostfixCapture(t *testing.T) {
	t.Parallel()

	conditionScope := symboltable.NewSymbolTable(symboltable.TempScope)
	state := parser.NewConditionalParseState(symboltable.WhenBodyScope, conditionScope)

	// Before EnterBody, postfix should be nil
	assert.Nil(t, state.GetConditionPostfix(), "postfix should be nil before EnterBody")

	// Capture postfix during condition-to-body transition
	expectedPostfix := []symboltable.SymbolDefinition{
		{DefinitionType: symboltable.Identifier},
		{DefinitionType: symboltable.LiteralDefinition},
		{DefinitionType: symboltable.GreaterThanOperatorDefinition},
	}
	state.EnterBody(expectedPostfix)

	actual := state.GetConditionPostfix()
	assert.Equal(t, expectedPostfix, actual, "GetConditionPostfix should return captured postfix")
	assert.Len(t, actual, 3, "postfix should have 3 elements")
	assert.Equal(t, symboltable.Identifier, actual[0].DefinitionType)
	assert.Equal(t, symboltable.LiteralDefinition, actual[1].DefinitionType)
	assert.Equal(t, symboltable.GreaterThanOperatorDefinition, actual[2].DefinitionType)
}

// TestConditionalParseStateEnterBodyIdempotency verifies that EnterBody only
// transitions once from condition to body. Subsequent calls have no effect.
func TestConditionalParseStateEnterBodyIdempotency(t *testing.T) {
	t.Parallel()

	conditionScope := symboltable.NewSymbolTable(symboltable.TempScope)
	state := parser.NewConditionalParseState(symboltable.ElseIfBodyScope, conditionScope)

	postfix1 := []symboltable.SymbolDefinition{
		{DefinitionType: symboltable.Identifier},
	}
	postfix2 := []symboltable.SymbolDefinition{
		{DefinitionType: symboltable.LiteralDefinition},
	}

	// First call succeeds
	ok := state.EnterBody(postfix1)
	assert.True(t, ok, "first EnterBody should succeed")
	assert.Equal(t, postfix1, state.GetConditionPostfix(), "postfix should be postfix1")

	// Second call from body phase fails
	ok = state.EnterBody(postfix2)
	assert.False(t, ok, "second EnterBody should fail")
	assert.Equal(t, postfix1, state.GetConditionPostfix(), "postfix should still be postfix1")
	assert.True(t, state.IsInBody(), "state should still be in body phase")
}

// TestConditionalParseStateMarkBodyCompleteIdempotency verifies that
// MarkBodyComplete only transitions once from body to complete.
func TestConditionalParseStateMarkBodyCompleteIdempotency(t *testing.T) {
	t.Parallel()

	conditionScope := symboltable.NewSymbolTable(symboltable.TempScope)
	state := parser.NewConditionalParseState(symboltable.IfBodyScope, conditionScope)

	// MarkBodyComplete from condition phase should have no effect
	state.MarkBodyComplete()
	assert.False(t, state.IsComplete(), "MarkBodyComplete from condition should not complete")
	assert.True(t, state.IsInCondition(), "should still be in condition phase")

	// Move to body phase
	state.EnterBody([]symboltable.SymbolDefinition{})
	assert.True(t, state.IsInBody(), "should be in body phase")

	// First MarkBodyComplete succeeds
	state.MarkBodyComplete()
	assert.True(t, state.IsComplete(), "should be complete after MarkBodyComplete")

	// Second MarkBodyComplete should not change state
	state.MarkBodyComplete()
	assert.True(t, state.IsComplete(), "should still be complete")
}

// TestConditionalParseStateEnterBodyFromCompleteFails verifies that EnterBody
// returns false when called from the complete state.
func TestConditionalParseStateEnterBodyFromCompleteFails(t *testing.T) {
	t.Parallel()

	conditionScope := symboltable.NewSymbolTable(symboltable.TempScope)
	state := parser.NewConditionalParseState(symboltable.IfBodyScope, conditionScope)

	// Move through full lifecycle
	state.EnterBody([]symboltable.SymbolDefinition{{DefinitionType: symboltable.Identifier}})
	state.MarkBodyComplete()
	assert.True(t, state.IsComplete(), "state should be complete")

	// EnterBody from complete state should fail
	ok := state.EnterBody([]symboltable.SymbolDefinition{{DefinitionType: symboltable.LiteralDefinition}})
	assert.False(t, ok, "EnterBody from complete state should fail")
	assert.True(t, state.IsComplete(), "state should still be complete")
	assert.Equal(t, symboltable.Identifier, state.GetConditionPostfix()[0].DefinitionType,
		"postfix should remain unchanged")
}

// TestConditionalParseStateNilAndEmptyPostfix verifies that nil and empty
// postfix slices are captured correctly.
func TestConditionalParseStateNilAndEmptyPostfix(t *testing.T) {
	t.Parallel()

	conditionScope := symboltable.NewSymbolTable(symboltable.TempScope)

	// Nil postfix
	state1 := parser.NewConditionalParseState(symboltable.IfBodyScope, conditionScope)
	state1.EnterBody(nil)
	assert.Nil(t, state1.GetConditionPostfix(), "nil postfix should be captured as nil")
	assert.True(t, state1.IsInBody(), "state should be in body phase after EnterBody with nil")

	// Empty postfix
	state2 := parser.NewConditionalParseState(symboltable.ElseIfBodyScope, conditionScope)
	state2.EnterBody([]symboltable.SymbolDefinition{})
	assert.NotNil(t, state2.GetConditionPostfix(), "empty postfix should be captured as non-nil empty slice")
	assert.Empty(t, state2.GetConditionPostfix(), "empty postfix should have length 0")
	assert.True(t, state2.IsInBody(), "state should be in body phase after EnterBody with empty slice")
}

// TestConditionalParseStateDeepScopeNesting verifies that deeply nested scopes
// (grandchildren of conditionScope) are still correctly distinguished.
func TestConditionalParseStateDeepScopeNesting(t *testing.T) {
	t.Parallel()

	conditionScope := symboltable.NewSymbolTable(symboltable.TempScope)
	childScope := conditionScope.AddChild(symboltable.TempScope)
	grandchildScope := childScope.AddChild(symboltable.TempScope)
	state := parser.NewConditionalParseState(symboltable.WhenBodyScope, conditionScope)

	assert.Same(t, conditionScope, state.GetConditionScope(),
		"condition scope should be the original scope")
	assert.NotSame(t, childScope, state.GetConditionScope(),
		"child scope should not match condition scope")
	assert.NotSame(t, grandchildScope, state.GetConditionScope(),
		"grandchild scope should not match condition scope")
}

// TestConditionalParseStateMultipleLifecycleCycles verifies that a single
// ConditionalParseState instance can be reused across multiple condition/body
// cycles via EnterCondition.
func TestConditionalParseStateMultipleLifecycleCycles(t *testing.T) {
	t.Parallel()

	conditionScope := symboltable.NewSymbolTable(symboltable.TempScope)
	state := parser.NewConditionalParseState(symboltable.IfBodyScope, conditionScope)

	// First cycle
	postfix1 := []symboltable.SymbolDefinition{
		{DefinitionType: symboltable.Identifier},
		{DefinitionType: symboltable.GreaterThanOperatorDefinition},
	}
	state.EnterBody(postfix1)
	state.MarkBodyComplete()
	assert.True(t, state.IsComplete(), "should be complete after first cycle")
	assert.Equal(t, postfix1, state.GetConditionPostfix(),
		"postfix should be captured from first cycle")

	// Reset and second cycle
	state.EnterCondition()
	assert.True(t, state.IsInCondition(),
		"should be in condition phase after EnterCondition")

	postfix2 := []symboltable.SymbolDefinition{
		{DefinitionType: symboltable.LiteralDefinition},
		{DefinitionType: symboltable.LiteralDefinition},
		{DefinitionType: symboltable.PlusOperatorDefinition},
	}
	state.EnterBody(postfix2)
	assert.True(t, state.IsInBody(),
		"should be in body phase during second cycle")
	assert.Equal(t, postfix2, state.GetConditionPostfix(),
		"postfix should be captured from second cycle")

	state.MarkBodyComplete()
	assert.True(t, state.IsComplete(), "should be complete after second cycle")
}

// TestConditionalParseStateConstructorSetsFields verifies that the constructor
// correctly initializes all fields.
func TestConditionalParseStateConstructorSetsFields(t *testing.T) {
	t.Parallel()

	conditionScope := symboltable.NewSymbolTable(symboltable.TempScope)

	testCases := []struct {
		name            string
		bodyScopeType   symboltable.ScopeType
		expectScopeType symboltable.ScopeType
	}{
		{"IfBodyScope", symboltable.IfBodyScope, symboltable.IfBodyScope},
		{"ElseIfBodyScope", symboltable.ElseIfBodyScope, symboltable.ElseIfBodyScope},
		{"ElseBodyScope", symboltable.ElseBodyScope, symboltable.ElseBodyScope},
		{"WhenBodyScope", symboltable.WhenBodyScope, symboltable.WhenBodyScope},
		{"DefaultBodyScope", symboltable.DefaultBodyScope, symboltable.DefaultBodyScope},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			state := parser.NewConditionalParseState(tc.bodyScopeType, conditionScope)
			assert.Equal(t, tc.expectScopeType, state.GetBodyScopeType(),
				"GetBodyScopeType should match constructor value")
			assert.Same(t, conditionScope, state.GetConditionScope(),
				"GetConditionScope should match constructor value")
			assert.True(t, state.IsInCondition(), "new state should be in condition phase")
		})
	}
}
