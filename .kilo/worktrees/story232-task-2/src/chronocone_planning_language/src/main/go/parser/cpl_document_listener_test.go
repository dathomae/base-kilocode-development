// Package parser_test contains tests for the CPL parser listener.
package parser_test

import (
	"errors"
	"testing"

	antlrruntime "github.com/antlr4-go/antlr/v4"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCPLListener_InitializesEmptyErrorSlice verifies that a newly created
// CPLListener has an empty error slice.
func TestNewCPLListener_InitializesEmptyErrorSlice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
	}{
		{
			name: "listener with nil logger has empty error slice",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			listener := parser.NewCPLListener(nil, nil, nil)
			require.NotNil(t, listener, "Listener should not be nil")

			assert.False(t, listener.HasErrors(), "New listener should not have errors")
			assert.Empty(t, listener.GetErrors(), "New listener should return empty error slice")
		})
	}
}

// TestAddError_AppendsErrors verifies that AddError correctly appends errors
// to the listener's error collection.
func TestAddError_AppendsErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		errorsToAdd   []error
		expectedCount int
	}{
		{
			name:          "add single error",
			errorsToAdd:   []error{errors.New("first error")},
			expectedCount: 1,
		},
		{
			name: "add multiple errors",
			errorsToAdd: []error{
				errors.New("first error"),
				errors.New("second error"),
				errors.New("third error"),
			},
			expectedCount: 3,
		},
		{
			name:          "add no errors",
			errorsToAdd:   []error{},
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			listener := parser.NewCPLListener(nil, nil, nil)

			for _, err := range tc.errorsToAdd {
				listener.AddError(err)
			}

			if tc.expectedCount > 0 {
				assert.True(t, listener.HasErrors(), "Listener should have errors")
			} else {
				assert.False(t, listener.HasErrors(), "Listener should not have errors")
			}

			errors := listener.GetErrors()
			assert.Len(t, errors, tc.expectedCount, "Error count should match")

			// Verify each error is present
			for i, expectedErr := range tc.errorsToAdd {
				assert.Equal(t, expectedErr, errors[i], "Error at index %d should match", i)
			}
		})
	}
}

// TestGetErrors_ReturnsCopyOfSlice verifies that GetErrors returns a copy
// of the internal errors slice, preventing external modifications.
func TestGetErrors_ReturnsCopyOfSlice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
	}{
		{
			name: "modifying returned slice does not affect listener",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			listener := parser.NewCPLListener(nil, nil, nil)

			// Add first error
			firstError := errors.New("first error")
			listener.AddError(firstError)

			// Get errors slice A
			sliceA := listener.GetErrors()
			require.Len(t, sliceA, 1, "Slice A should have 1 error")

			// Add second error
			secondError := errors.New("second error")
			listener.AddError(secondError)

			// Get errors slice B
			sliceB := listener.GetErrors()
			require.Len(t, sliceB, 2, "Slice B should have 2 errors")

			// Verify slice A hasn't changed (it's a copy)
			assert.Len(t, sliceA, 1, "Slice A should still have 1 error (copy behavior)")
			assert.Equal(t, firstError, sliceA[0], "Slice A should still contain first error")

			// Verify slice B contains both errors
			assert.Len(t, sliceB, 2, "Slice B should have both errors")
			assert.Equal(t, firstError, sliceB[0], "Slice B should contain first error")
			assert.Equal(t, secondError, sliceB[1], "Slice B should contain second error")
		})
	}
}

// TestSemanticError_ContainsExpectedContext verifies that SemanticError
// properly formats its context information in the Error() string.
func TestSemanticError_ContainsExpectedContext(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		semanticError  *parser.SemanticError
		expectContains []string
	}{
		{
			name: "error contains all context fields",
			semanticError: &parser.SemanticError{
				Message: "undefined type reference",
				Rule:    "TypeReference",
				Line:    42,
				Column:  15,
			},
			expectContains: []string{
				"undefined type reference",
				"TypeReference",
				"42",
				"15",
			},
		},
		{
			name: "error with different context",
			semanticError: &parser.SemanticError{
				Message: "invalid syntax",
				Rule:    "Statement",
				Line:    1,
				Column:  0,
			},
			expectContains: []string{
				"invalid syntax",
				"Statement",
				"1",
				"0",
			},
		},
		{
			name: "error with empty message",
			semanticError: &parser.SemanticError{
				Message: "",
				Rule:    "Document",
				Line:    100,
				Column:  50,
			},
			expectContains: []string{
				"Document",
				"100",
				"50",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			errString := tc.semanticError.Error()

			for _, expected := range tc.expectContains {
				assert.Contains(t, errString, expected,
					"Error string should contain '%s'", expected)
			}
		})
	}
}

// TestHasErrors_ReturnsTrueWhenErrorsExist verifies the HasErrors method
// correctly reflects the presence of errors.
func TestHasErrors_ReturnsTrueWhenErrorsExist(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		addErrors         []error
		expectedHasErrors bool
	}{
		{
			name:              "no errors returns false",
			addErrors:         []error{},
			expectedHasErrors: false,
		},
		{
			name:              "single error returns true",
			addErrors:         []error{errors.New("error 1")},
			expectedHasErrors: true,
		},
		{
			name: "multiple errors returns true",
			addErrors: []error{
				errors.New("error 1"),
				errors.New("error 2"),
			},
			expectedHasErrors: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			listener := parser.NewCPLListener(nil, nil, nil)

			// Initially should have no errors
			assert.False(t, listener.HasErrors(),
				"New listener should not have errors initially")

			// Add errors
			for _, err := range tc.addErrors {
				listener.AddError(err)
			}

			// Check result
			assert.Equal(t, tc.expectedHasErrors, listener.HasErrors(),
				"HasErrors should return %v after adding %d error(s)",
				tc.expectedHasErrors, len(tc.addErrors))
		})
	}
}

// TestParseDocument_VerifyGlobalScopeStructure verifies that the global scope
// is correctly initialized and has the expected structure.
func TestParseDocument_VerifyGlobalScopeStructure(t *testing.T) {
	t.Parallel()

	listener := parser.NewCPLListener(nil, nil, nil)

	// Verify symbol stack is initialized
	stack := listener.GetSymbolStack()
	require.NotNil(t, stack, "Symbol stack should be initialized")

	// Verify stack starts with global scope
	assert.Equal(t, 1, stack.Size(), "Stack should have one table initially")

	// Verify root is global scope
	root, err := stack.Root()
	require.NoError(t, err, "Should be able to get root")
	assert.Equal(t, symboltable.GlobalScope, root.ScopeType, "Root should be GlobalScope")

	// Verify root has no children initially
	assert.Empty(t, root.Children, "Global scope should have no children initially")

	// Verify root has no definitions initially
	assert.Empty(t, root.Definitions, "Global scope should have no definitions initially")

	// Verify root has no parent
	assert.Nil(t, root.Parent, "Global scope should have no parent")
}

// parseWithListenerDocument parses input with a custom listener for document-level tests.
// This is a helper function that creates a direct ANTLR parse flow.
func parseWithListenerDocument(input string, listener *parser.CPLListener) (*parser.CPLListener, error) {
	// Create input stream and lexer
	inputStream := antlrruntime.NewInputStream(input)
	lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)
	tokenStream := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
	antlrParser := antlr.NewChronoconePlanningLanguageParser(tokenStream)

	// Walk the parse tree with the provided listener
	antlrruntime.ParseTreeWalkerDefault.Walk(listener, antlrParser.Document())

	return listener, nil
}
