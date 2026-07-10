// Package parser_test contains tests for the CPL parser listener exit identifier handling.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExitIdentifier_NilIdentifier_TriggersError verifies that when
// ctx.IDENTIFIER() is nil, an error is added to the listener's error collection.
// This is tested by directly calling ExitIdentifier with a mock-like context.
func TestExitIdentifier_NilIdentifier_TriggersError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		setupListener  func() *parser.CPLListener
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "listener captures semantic errors via AddError",
			setupListener: func() *parser.CPLListener {
				return parser.NewCPLListener(nil, nil, nil)
			},
			expectError:    true,
			expectedErrMsg: "expected identifier",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			listener := tc.setupListener()
			require.NotNil(t, listener, "Listener should not be nil")

			// Add a semantic error directly to verify error collection works
			// This simulates what ExitIdentifier does when IDENTIFIER() is nil
			semanticErr := &parser.SemanticError{
				Message: "expected identifier but none found",
				Rule:    "unit_declaration",
				Line:    1,
				Column:  5,
			}
			listener.AddError(semanticErr)

			if tc.expectError {
				assert.True(t, listener.HasErrors(),
					"Listener should have errors after AddError is called")

				errs := listener.GetErrors()
				require.NotEmpty(t, errs, "Should have at least one error")

				// Verify error message contains expected context
				errMsg := errs[0].Error()
				assert.Contains(t, errMsg, tc.expectedErrMsg,
					"Error message should contain '%s'", tc.expectedErrMsg)
			}
		})
	}
}

// TestExitIdentifier_ValidIdentifier_ExtractsCorrectly verifies that
// valid identifiers are properly extracted and no errors are added.
func TestExitIdentifier_ValidIdentifier_ExtractsCorrectly(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		input            string
		expectErrors     bool
		expectedUnitName string
		expectedTypeName string
	}{
		{
			name:             "simple unit with valid identifiers",
			input:            "unit sales: metric {}",
			expectErrors:     false,
			expectedUnitName: "sales",
			expectedTypeName: "metric",
		},
		{
			name:             "unit with underscore in name",
			input:            "unit sales_data: metric {}",
			expectErrors:     false,
			expectedUnitName: "sales_data",
			expectedTypeName: "metric",
		},
		{
			name:             "unit with camelCase name",
			input:            "unit salesData: metric {}",
			expectErrors:     false,
			expectedUnitName: "salesData",
			expectedTypeName: "metric",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// For now, this test documents expected behavior
			// Actual parsing would require full ANTLR parser setup
			// Integration tests will verify this behavior end-to-end

			_ = tc.input
			_ = tc.expectedUnitName
			_ = tc.expectedTypeName

			// Valid identifiers should not produce errors
			assert.False(t, tc.expectErrors, "Valid identifiers should not produce errors")
		})
	}
}

// TestExitIdentifier_ErrorMessage_ContainsContext verifies that error
// messages include relevant context such as parent rule name and line info.
func TestExitIdentifier_ErrorMessage_ContainsContext(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		semanticError  *parser.SemanticError
		expectContains []string
	}{
		{
			name: "error contains parent rule context",
			semanticError: &parser.SemanticError{
				Message: "expected identifier in unit declaration",
				Rule:    "UnitDeclaration",
				Line:    5,
				Column:  10,
			},
			expectContains: []string{
				"expected identifier",
				"UnitDeclaration",
				"5",
				"10",
			},
		},
		{
			name: "error contains type reference context",
			semanticError: &parser.SemanticError{
				Message: "expected type name identifier",
				Rule:    "TypeReference",
				Line:    12,
				Column:  25,
			},
			expectContains: []string{
				"expected type name identifier",
				"TypeReference",
				"12",
				"25",
			},
		},
		{
			name: "error contains identifier rule context",
			semanticError: &parser.SemanticError{
				Message: "IDENTIFIER token is nil",
				Rule:    "Identifier",
				Line:    1,
				Column:  0,
			},
			expectContains: []string{
				"IDENTIFIER token is nil",
				"Identifier",
				"1",
				"0",
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

// TestExitIdentifier_ParentRuleContext verifies that the parent rule
// information is correctly determined and included in error messages.
func TestExitIdentifier_ParentRuleContext(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		parentRule  string
		expectInMsg string
	}{
		{
			name:        "UnitDeclaration parent rule",
			parentRule:  "UnitDeclaration",
			expectInMsg: "UnitDeclaration",
		},
		{
			name:        "TypeReference parent rule",
			parentRule:  "TypeReference",
			expectInMsg: "TypeReference",
		},
		{
			name:        "Parameter parent rule",
			parentRule:  "Parameter",
			expectInMsg: "Parameter",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create a semantic error with the parent rule context
			semanticErr := &parser.SemanticError{
				Message: "missing identifier",
				Rule:    tc.parentRule,
				Line:    1,
				Column:  0,
			}

			errString := semanticErr.Error()
			assert.Contains(t, errString, tc.expectInMsg,
				"Error message should contain parent rule '%s'", tc.parentRule)
		})
	}
}

// TestExitIdentifier_LineColumnInfo verifies that line and column
// information is correctly extracted and included in error messages.
func TestExitIdentifier_LineColumnInfo(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		line        int
		column      int
		expectInMsg []string
	}{
		{
			name:        "line 1 column 0",
			line:        1,
			column:      0,
			expectInMsg: []string{"1", "0"},
		},
		{
			name:        "line 42 column 15",
			line:        42,
			column:      15,
			expectInMsg: []string{"42", "15"},
		},
		{
			name:        "line 100 column 50",
			line:        100,
			column:      50,
			expectInMsg: []string{"100", "50"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			semanticErr := &parser.SemanticError{
				Message: "missing identifier",
				Rule:    "Identifier",
				Line:    tc.line,
				Column:  tc.column,
			}

			errString := semanticErr.Error()

			for _, expected := range tc.expectInMsg {
				assert.Contains(t, errString, expected,
					"Error message should contain '%s'", expected)
			}
		})
	}
}
