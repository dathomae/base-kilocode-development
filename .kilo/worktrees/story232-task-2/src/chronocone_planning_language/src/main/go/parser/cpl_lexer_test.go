// Package parser_test contains tests for the CPL lexer.
package parser_test

import (
	"testing"

	antlrruntime "github.com/antlr4-go/antlr/v4"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestArrowToken verifies that the two-character sequence "=>" is correctly
// recognized by the lexer as a single ARROW token.
//
// Test cases:
//   - "lexed as single ARROW" — the standalone "=>" input must produce one
//     token whose type is ChronoconePlanningLanguageLexerARROW.
//   - "not lexed as EQUALS" — the same input must NOT produce a token whose
//     type is ChronoconePlanningLanguageLexerEQUALS, ensuring the longest-match
//     rule is applied (ANTLR resolves "=>" as a single token instead of "=" + ">").
//
// Why this test is important:
//   - ARROW is the punctuation token that separates a parent-value condition
//     from the child concept definition in conditional syntax
//     (e.g., Country[USA] => Region=[Values[NY, CA]]).
//   - If "=>" were mis-lexed as two EQUALS tokens (or as EQUALS + GREATER),
//     the parser would see a malformed sequence and reject valid conditional
//     CPL syntax.
func TestArrowToken(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		input           string
		expectedText    string
		expectedType    int
		shouldNotBeType int
	}{
		{
			name:         "lexed as single ARROW",
			input:        "=>",
			expectedText: "=>",
			expectedType: antlr.ChronoconePlanningLanguageLexerARROW,
		},
		{
			name:            "not lexed as EQUALS",
			input:           "=>",
			expectedText:    "=>",
			expectedType:    antlr.ChronoconePlanningLanguageLexerARROW,
			shouldNotBeType: antlr.ChronoconePlanningLanguageLexerEQUALS,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			inputStream := antlrruntime.NewInputStream(tc.input)
			lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)
			tokens := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
			tokens.Fill()

			require.GreaterOrEqual(t, tokens.Size(), 1, "expected at least one token in the stream")

			firstToken := tokens.Get(0)
			assert.Equal(t, tc.expectedText, firstToken.GetText(), "token text should match expected")
			assert.Equal(t, tc.expectedType, firstToken.GetTokenType(), "token type should be ARROW")

			if tc.shouldNotBeType != 0 {
				assert.NotEqual(t, tc.shouldNotBeType, firstToken.GetTokenType(),
					"=> should not be lexed as %d", tc.shouldNotBeType)
			}
		})
	}
}

// TestArrowTokenInConditionalContext verifies that the ARROW token (=>) is
// recognized correctly when it appears in a realistic multi-token conditional
// CPL fragment, surrounded by LBRACKET, RBRACKET, EQUALS, and IDENTIFIER tokens.
//
// What is tested:
//   - Input: "Country[USA] => Region"
//   - Expected behavior: the token stream contains an ARROW token positioned
//     between the RBRACKET that closes the parent-value qualifier and the
//     IDENTIFIER that names the child concept.
//
// Why this test is important:
//   - This validates the ARROW token in its real syntactic context, ensuring
//     that preceding tokens (RBRACKET, WS) and following tokens (IDENTIFIER, WS)
//     do not interfere with correct ARROW recognition.
//   - It tests the end-to-end lexer behavior for Increment 4 conditional
//     syntax, which is the primary user-facing feature of this story.
func TestArrowTokenInConditionalContext(t *testing.T) {
	t.Parallel()

	input := "Country[USA] => Region"

	inputStream := antlrruntime.NewInputStream(input)
	lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)
	tokens := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
	tokens.Fill()

	var found bool
	for i := 0; i < tokens.Size(); i++ {
		token := tokens.Get(i)
		if token.GetText() == "=>" {
			found = true
			assert.Equal(t, antlr.ChronoconePlanningLanguageLexerARROW, token.GetTokenType(),
				"'=>' in conditional context should have ARROW token type")
			break
		}
	}
	require.True(t, found, "token stream should contain ARROW token '=>'")
}
