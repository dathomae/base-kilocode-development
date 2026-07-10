// Package parser_test contains tests for the CPL parser error handling.
package parser_test

import (
	"os"
	"path"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseWithLogLevel_SyntaxErrors_ReturnsErrors verifies that syntax errors
// from the lexer and parser are correctly returned by ParseWithLogLevel.
func TestParseWithLogLevel_SyntaxErrors_ReturnsErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:           "missing equals sign in types",
			input:          "types\n\tSales Metric",
			expectError:    true,
			expectedErrMsg: "mismatched",
		},
		{
			name:           "invalid token at start",
			input:          "@#$ Sales = Number",
			expectError:    true,
			expectedErrMsg: "token",
		},
		{
			name:           "incomplete types declaration",
			input:          "types",
			expectError:    true,
			expectedErrMsg: "mismatched",
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
				assert.Contains(t, err.Error(), tc.expectedErrMsg,
					"Error message should contain '%s'", tc.expectedErrMsg)
			} else {
				assert.NoError(t, err, "Expected no error for input: %s", tc.input)
			}
		})
	}
}

// TestParseWithLogLevel_NoErrors_ReturnsDocument verifies that valid CPL documents
// parse successfully without errors and return a valid document.
func TestParseWithLogLevel_NoErrors_ReturnsDocument(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "simple valid types section",
			input:       "types\n\tSales = Number()",
			expectError: false,
		},
		{
			name:        "simple valid concepts section",
			input:       "concepts\n\tCountry=[Values[USA, Canada]]",
			expectError: false,
		},
		{
			name:        "multiple concepts",
			input:       "concepts\n\tCountry=[Values[USA, Canada]]\n\tPriority=[Values[High, Low]]",
			expectError: false,
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

			if tc.expectError {
				require.Error(t, err, "Expected an error for input: %s", tc.input)
			} else {
				require.NoError(t, err, "Expected no error for input: %s", tc.input)
				require.NotNil(t, doc, "Expected a valid document")
			}
		})
	}
}

// TestParseWithLogLevel_ErrorMessageFormat_IsReadable verifies that combined
// error messages have a readable format.
func TestParseWithLogLevel_ErrorMessageFormat_IsReadable(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		input            string
		expectError      bool
		expectedPrefix   string
		expectedContains []string
	}{
		{
			name:           "syntax error format includes parse error prefix",
			input:          "types\n\t= Number",
			expectError:    true,
			expectedPrefix: "parse error:",
			expectedContains: []string{
				"extraneous input", // Syntax error component
			},
		},
		{
			name:           "syntax error format for incomplete type reference",
			input:          "types\n\tSales =",
			expectError:    true,
			expectedPrefix: "parse error:",
			expectedContains: []string{
				"mismatched", // Syntax error marker
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
				assert.Contains(t, errMsg, tc.expectedPrefix,
					"Error message should start with '%s'", tc.expectedPrefix)
				for _, expected := range tc.expectedContains {
					assert.Contains(t, errMsg, expected,
						"Error message should contain '%s'", expected)
				}
			}
		})
	}
}

// TestParseWithLogLevel_NilConfig_ReturnsError verifies that ParseWithLogLevel
// returns an error when passed a nil configuration.
func TestParseWithLogLevel_NilConfig_ReturnsError(t *testing.T) {
	t.Parallel()

	_, err := parser.ParseWithLogLevel("types\n\tMetric = Number()", nil)
	require.Error(t, err, "Expected an error when config is nil")
	assert.Contains(t, err.Error(), "nil", "Error should mention nil configuration")
}

// TestInvalidCondFiles_TriggerParseErrors verifies that each of the 5 invalid
// conditional concept example files triggers a parse error.
// This is Task 5d of Story209.
func TestInvalidCondFiles_TriggerParseErrors(t *testing.T) {
	t.Parallel()

	baseDir := "../../../../examples/parse_and_analyze_only/with_errors"
	files := []string{
		"invalid_conditional_concept_missing_arrow.cpl",
		"invalid_conditional_concept_missing_parent_value.cpl",
		"invalid_conditional_concept_missing_equals.cpl",
		"invalid_conditional_concept_empty_concept_line.cpl",
		"invalid_conditional_concept_missing_close_bracket.cpl",
	}

	for _, f := range files {
		f := f
		t.Run(f, func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(path.Clean(path.Join(baseDir, f)))
			require.NoError(t, err, "Should be able to read %s", f)
			_, err = parser.Parse(string(content))
			require.Error(t, err, "Expected parse error for %s, but got none", f)
		})
	}
}

// TestParse_ValidInput_ParsesSuccessfully verifies that the Parse function
// works correctly with default settings for valid input.
func TestParse_ValidInput_ParsesSuccessfully(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid types parses correctly",
			input:       "types\n\tMetric = Number()",
			expectError: false,
		},
		{
			name:        "syntax error returns error",
			input:       "types\n\t= Number()",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := parser.Parse(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected an error for input: %s", tc.input)
			} else {
				require.NoError(t, err, "Expected no error for input: %s", tc.input)
				require.NotNil(t, doc, "Expected a valid document")
			}
		})
	}
}
