// Package parser_test contains tests for the CPL concepts parser.
package parser_test

import (
	"testing"

	antlrruntime "github.com/antlr4-go/antlr/v4"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseConceptsSection tests parsing concept definitions with Values generator.
func TestParseConceptsSection(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name: "simple concept with Values generator",
			input: `concepts
		Country=[Values[USA, Canada, Mexico]]`,
			expectError: false,
		},
		{
			name: "concept with multiple values",
			input: `concepts
		Priority=[Values[Low, Medium, High]]`,
			expectError: false,
		},
		{
			name: "multiple concepts in section",
			input: `concepts
		Country=[Values[USA, Canada]]
		Priority=[Values[Low, High]]`,
			expectError: false,
		},
		{
			name: "single value in Values generator",
			input: `concepts
		Status=[Values[Active]]`,
			expectError: false,
		},
		{
			name: "concept with many values",
			input: `concepts
		Month=[Values[Jan, Feb, Mar, Apr, May, Jun, Jul, Aug, Sep, Oct, Nov, Dec]]`,
			expectError: false,
		},
		{
			name:          "empty concepts section",
			input:         `concepts`,
			expectError:   true,
			errorContains: "mismatched input",
		},
		{
			name: "concept with empty values",
			input: `concepts
		Country=[Values[]]`,
			expectError: true,
		},
		{
			name: "concept missing equals sign",
			input: `concepts
		Country[Values[USA]]`,
			expectError: true,
		},
		{
			name: "concept missing Values keyword",
			input: `concepts
		Country=[[USA, Canada]]`,
			expectError: true,
		},
		{
			name: "concept with missing brackets",
			input: `concepts
		Country=Values[USA, Canada]`,
			expectError: true,
		},
		{
			name: "concept with trailing comma",
			input: `concepts
		Country=[Values[USA, Canada,]]`,
			expectError: true,
		},
		{
			name: "concept with integer value",
			input: `concepts
		Year=[Values[2024]]`,
			expectError: false,
		},
		{
			name: "concept with multiple integer values",
			input: `concepts
		Year=[Values[2024, 2025, 2026]]`,
			expectError: false,
		},
		{
			name: "concept with mixed identifier and integer values",
			input: `concepts
		Quarter=[Values[Q1, Q2, Q3, Q4, 2024]]`,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Parse the input and verify it succeeds or fails as expected
			_, err := parser.Parse(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected an error for input: %s", tc.input)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, "Error message should contain expected text")
				}
				return
			}

			require.NoError(t, err, "Expected no error for input: %s", tc.input)
		})
	}
}

// TestValueContext_Accessors verifies that the generated ValueContext exposes
// accessors for IDENTIFIER, VALUE, NUMBER, and INTEGER token types, and that parsing
// correctly sets the appropriate accessor based on the input token.
func TestValueContext_Accessors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		input             string
		expectIdentifier  bool
		expectValue       bool
		expectNumber      bool
		expectInteger     bool
		expectedTokenText string
	}{
		{
			name:              "identifier token",
			input:             "concepts\n\tCountry=[Values[USA]]",
			expectIdentifier:  true,
			expectedTokenText: "USA",
		},
		{
			name:              "integer token (now NUMBER due to NUMBER lexer priority)",
			input:             "concepts\n\tYear=[Values[2024]]",
			expectNumber:      true,
			expectedTokenText: "2024",
		},
		{
			name:              "value token (alphanumeric starting with digit)",
			input:             "concepts\n\tCode=[Values[2024A]]",
			expectValue:       true,
			expectedTokenText: "2024A",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create parser directly to inspect ValueContext
			inputStream := antlrruntime.NewInputStream(tc.input)
			lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)
			tokenStream := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
			p := antlr.NewChronoconePlanningLanguageParser(tokenStream)

			doc := p.Document()
			require.NotNil(t, doc, "Document should parse successfully")

			// Navigate to the first ValueContext in the document
			// document -> statement -> conceptsSection -> conceptDefinition -> generatorPatternList
			// -> generatorPattern -> valuesPatternGenerator -> conceptValueList -> value
			statements := doc.(*antlr.DocumentContext).AllStatement()
			require.NotEmpty(t, statements, "Document should have at least one statement")

			conceptsSection := statements[0].(*antlr.StatementContext).ConceptsSection()
			require.NotNil(t, conceptsSection, "Statement should be a concepts section")

			conceptDefs := conceptsSection.AllConceptDefinition()
			require.NotEmpty(t, conceptDefs, "Concepts section should have at least one concept definition")

			conceptLines := conceptDefs[0].AllConceptLine()
			require.NotEmpty(t, conceptLines, "Concept definition should have a concept line")

			genPatternLists := conceptLines[0].GeneratorPatternList()
			require.NotNil(t, genPatternLists, "Concept line should have a generator pattern list")

			genPatterns := genPatternLists.AllGeneratorPattern()
			require.NotEmpty(t, genPatterns, "Generator pattern list should have patterns")

			valuesGen := genPatterns[0].ValuesPatternGenerator()
			require.NotNil(t, valuesGen, "Generator pattern should be a Values generator")

			valueList := valuesGen.ConceptValueList()
			require.NotNil(t, valueList, "Values generator should have a concept value list")

			values := valueList.AllValue()
			require.NotEmpty(t, values, "Concept value list should have at least one value")

			valueCtx := values[0].(*antlr.ValueContext)

			// Verify the appropriate accessor returns the expected terminal node
			if tc.expectIdentifier {
				node := valueCtx.IDENTIFIER()
				require.NotNil(t, node, "IDENTIFIER() should return a terminal node")
				assert.Equal(t, tc.expectedTokenText, node.GetText(), "Token text should match")
			}
			if tc.expectValue {
				node := valueCtx.VALUE()
				require.NotNil(t, node, "VALUE() should return a terminal node")
				assert.Equal(t, tc.expectedTokenText, node.GetText(), "Token text should match")
			}
			if tc.expectNumber {
				node := valueCtx.NUMBER()
				require.NotNil(t, node, "NUMBER() should return a terminal node")
				assert.Equal(t, tc.expectedTokenText, node.GetText(), "Token text should match")
			}
			if tc.expectInteger {
				node := valueCtx.INTEGER()
				require.NotNil(t, node, "INTEGER() should return a terminal node")
				assert.Equal(t, tc.expectedTokenText, node.GetText(), "Token text should match")
			}
		})
	}
}
