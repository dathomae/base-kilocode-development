// Package parser_test contains integration tests for CPL parser error collection.
//
// These tests verify that error collection through the public ParseWithLogLevel API
// works correctly for various error scenarios including syntax errors, semantic errors,
// and combinations of both.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_ParameterOutsideTypesContext_ReturnsError verifies that
// parameter declarations outside of a types context produce errors.
func TestIntegration_ParameterOutsideTypesContext_ReturnsError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		input        string
		expectError  bool
		expectedErrs []string
	}{
		{
			name:        "parameter at document level produces syntax error",
			input:       "precision: '2'",
			expectError: true,
			expectedErrs: []string{
				"parse error",
			},
		},
		{
			name:        "multiple parameters outside types produces syntax error",
			input:       "a: '1' b: '2'",
			expectError: true,
			expectedErrs: []string{
				"parse error",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := &parser.ParseLoggingConfig{
				LogLevel:   0, // INFO level
				OutputFile: "",
			}

			_, err := parser.ParseWithLogLevel(tc.input, config)

			if tc.expectError {
				require.Error(t, err, "Expected an error for input: %s", tc.input)
				errMsg := err.Error()
				for _, expected := range tc.expectedErrs {
					assert.Contains(t, errMsg, expected,
						"Error message should contain '%s'", expected)
				}
			} else {
				require.NoError(t, err, "Expected no error for input: %s", tc.input)
			}
		})
	}
}

// TestIntegration_ValidDocuments_ReturnsNoErrors verifies that valid CPL documents
// parse successfully and return a valid document without errors.
func TestIntegration_ValidDocuments_ReturnsNoErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
	}{
		{
			name: "simple valid types section",
			input: `types
	Metric = Number()`,
		},
		{
			name: "simple valid concepts section",
			input: `concepts
	Country=[Values[USA, Canada]]`,
		},
		{
			name: "types and concepts sections",
			input: `types
	Metric = Number()
concepts
	Country=[Values[USA, Canada]]`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := &parser.ParseLoggingConfig{
				LogLevel:   0, // INFO level
				OutputFile: "",
			}

			doc, err := parser.ParseWithLogLevel(tc.input, config)

			require.NoError(t, err, "Expected no error for input: %s", tc.input)
			require.NotNil(t, doc, "Expected a valid document")
		})
	}
}

// TestIntegration_ParseAPI works correctly with the simplified Parse() API.
func TestIntegration_ParseAPI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		input        string
		expectError  bool
		expectedErrs []string
	}{
		{
			name:        "valid document using Parse API",
			input:       "types\n\tMetric = Number()",
			expectError: false,
		},
		{
			name:         "syntax error using Parse API",
			input:        "types\n\t= Number()",
			expectError:  true,
			expectedErrs: []string{"parse error"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := parser.Parse(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected an error for input: %s", tc.input)
				errMsg := err.Error()
				for _, expected := range tc.expectedErrs {
					assert.Contains(t, errMsg, expected,
						"Error message should contain '%s'", expected)
				}
			} else {
				require.NoError(t, err, "Expected no error for input: %s", tc.input)
				require.NotNil(t, doc, "Expected a valid document")
			}
		})
	}
}

// TestIntegration_NilConfig_ReturnsError verifies that ParseWithLogLevel
// returns an error when passed a nil configuration.
func TestIntegration_NilConfig_ReturnsError(t *testing.T) {
	t.Parallel()

	_, err := parser.ParseWithLogLevel("types\n\tMetric = Number()", nil)
	require.Error(t, err, "Expected an error when config is nil")
	assert.Contains(t, err.Error(), "nil", "Error should mention nil configuration")
}

// containsSemanticError is a helper function that checks if an error message
// contains semantic error indicators.
func containsSemanticError(errMsg string) bool {
	return len(errMsg) > 0 && (containsSubstring(errMsg, "semantic error") ||
		containsSubstring(errMsg, "type_declaration") ||
		containsSubstring(errMsg, "type_reference") ||
		containsSubstring(errMsg, "parameter"))
}

// containsSubstring is a helper function to check if a string contains a substring.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

// containsHelper is a helper that actually performs the substring check.
func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
