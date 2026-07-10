// Package parser_test contains tests for the CPL parser.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/stretchr/testify/require"
)

// TestParseEmptyDocument tests parsing an empty CPL document.
func TestParseEmptyDocument(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "completely empty document",
			input:       "",
			expectError: false,
		},
		{
			name:        "document with only whitespace",
			input:       "   \n\t  ",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := parser.Parse(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected an error")
				return
			}

			require.NoError(t, err, "Expected no error for empty document")
			require.NotNil(t, doc, "Symbol table should not be nil")
		})
	}
}
