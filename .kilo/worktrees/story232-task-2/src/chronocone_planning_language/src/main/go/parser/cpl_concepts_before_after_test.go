// Package parser_test contains tests for the CPL concepts parser before/after lexer tokens.
package parser_test

import (
	"testing"

	antlrruntime "github.com/antlr4-go/antlr/v4"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLexerRecognizesBeforeKeyword verifies that the lexer produces a BEFORE token
// for the keyword 'before' when encountered in CPL source text.
//
// What is tested:
//   - Input: the standalone keyword "before" fed directly to the lexer.
//   - Expected behavior: the lexer emits a single token whose token type equals
//     ChronoconePlanningLanguageLexerBEFORE (the ANTLR-generated constant for the
//     BEFORE keyword), rather than IDENTIFIER or any other token type.
//
// Why this test is important:
//   - BEFORE is a reserved keyword used in concept extension syntax to insert new
//     values before an existing value (e.g., Priority=[Values[Critical]] before Low).
//   - If the lexer fails to recognize 'before' as a keyword, it will fall back to
//     IDENTIFIER, causing downstream parser rules (insertionPoint, conceptExtension)
//     to reject valid CPL syntax.
//   - This is the foundational lexical test that must pass before any parser-level
//     before/after tests can succeed.
func TestLexerRecognizesBeforeKeyword(t *testing.T) {
	t.Parallel()

	inputStream := antlrruntime.NewInputStream("before")
	lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)
	tokens := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
	tokens.Fill()

	require.GreaterOrEqual(t, tokens.Size(), 1, "expected at least one token in the stream")

	firstToken := tokens.Get(0)
	assert.Equal(t, "before", firstToken.GetText(), "token text should be 'before'")
	assert.Equal(t, antlr.ChronoconePlanningLanguageLexerBEFORE, firstToken.GetTokenType(), "token type should be BEFORE")
}

// TestLexerRecognizesAfterKeyword verifies that the lexer produces an AFTER token
// for the keyword 'after' when encountered in CPL source text.
//
// What is tested:
//   - Input: the standalone keyword "after" fed directly to the lexer.
//   - Expected behavior: the lexer emits a single token whose token type equals
//     ChronoconePlanningLanguageLexerAFTER (the ANTLR-generated constant for the
//     AFTER keyword), rather than IDENTIFIER or any other token type.
//
// Why this test is important:
//   - AFTER is a reserved keyword used in concept extension syntax to insert new
//     values after an existing value (e.g., Priority=[Values[Urgent]] after Medium).
//   - If the lexer fails to recognize 'after' as a keyword, it will fall back to
//     IDENTIFIER, causing downstream parser rules (insertionPoint, conceptExtension)
//     to reject valid CPL syntax.
//   - This is the foundational lexical test that must pass before any parser-level
//     before/after tests can succeed.
func TestLexerRecognizesAfterKeyword(t *testing.T) {
	t.Parallel()

	inputStream := antlrruntime.NewInputStream("after")
	lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)
	tokens := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
	tokens.Fill()

	require.GreaterOrEqual(t, tokens.Size(), 1, "expected at least one token in the stream")

	firstToken := tokens.Get(0)
	assert.Equal(t, "after", firstToken.GetText(), "token text should be 'after'")
	assert.Equal(t, antlr.ChronoconePlanningLanguageLexerAFTER, firstToken.GetTokenType(), "token type should be AFTER")
}

// TestLexerBeforeAfterAreNotIdentifiers verifies that the keywords 'before' and
// 'after' are recognized as reserved keywords and not as generic IDENTIFIER tokens.
//
// What is tested:
//   - Input: the strings "before" and "after" lexed individually.
//   - Expected behavior: each string yields a token whose type is BEFORE or AFTER
//     respectively, and explicitly NOT ChronoconePlanningLanguageLexerIDENTIFIER.
//
// Why this test is important:
//   - The IDENTIFIER rule in the lexer ([a-zA-Z_][a-zA-Z0-9_]*) would otherwise
//     match 'before' and 'after' because they consist entirely of letters.
//   - ANTLR resolves keyword-vs-identifier ambiguity by prioritizing the rule that
//     appears first in the grammar; this test guards against future grammar
//     reordering that could accidentally demote BEFORE/AFTER below IDENTIFIER.
//   - If these words become identifiers, existing concept names or value names
//     named "before" or "after" could conflict with the keyword syntax, leading
//     to subtle parsing bugs in concept extensions.
func TestLexerBeforeAfterAreNotIdentifiers(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		input        string
		expectedType int
	}{
		{
			name:         "before is not identifier",
			input:        "before",
			expectedType: antlr.ChronoconePlanningLanguageLexerBEFORE,
		},
		{
			name:         "after is not identifier",
			input:        "after",
			expectedType: antlr.ChronoconePlanningLanguageLexerAFTER,
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
			assert.Equal(t, tc.expectedType, firstToken.GetTokenType(), "token type should match expected keyword type")
			assert.NotEqual(t, antlr.ChronoconePlanningLanguageLexerIDENTIFIER, firstToken.GetTokenType(), "token should not be an IDENTIFIER")
		})
	}
}

// TestLexerBeforeAfterInConceptContext verifies that 'before' and 'after' tokenize
// correctly when they appear inside a full concept extension CPL fragment.
//
// What is tested:
//   - Input: a complete concept extension line such as
//     "concepts\n\tPriority=[Values[Critical]] before Low"
//     and
//     "concepts\n\tPriority=[Values[Urgent]] after Medium"
//   - Expected behavior: the token stream contains a BEFORE (or AFTER) token
//     positioned between the RBRACKET that closes the generator pattern list and
//     the IDENTIFIER token that names the existing anchor value.
//
// Why this test is important:
//   - This tests the keywords in their real syntactic context, ensuring that
//     preceding tokens (RBRACKET, WS) and following tokens (IDENTIFIER, INDENT)
//     do not interfere with correct BEFORE/AFTER recognition.
//   - It validates the end-to-end lexer behavior for Increment 7 concept
//     extension syntax, which is the primary user-facing feature of this story.
//   - A failure here indicates that either the lexer grammar or the generated
//     lexer code does not correctly handle the BEFORE/AFTER tokens in multi-token
//     streams, which would block parser and listener implementation in later tasks.
func TestLexerBeforeAfterInConceptContext(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		input        string
		expectedType int
		keywordText  string
	}{
		{
			name:         "before in concept extension context",
			input:        "Priority=[Values[Critical]] before Low",
			expectedType: antlr.ChronoconePlanningLanguageLexerBEFORE,
			keywordText:  "before",
		},
		{
			name:         "after in concept extension context",
			input:        "Priority=[Values[Urgent]] after Medium",
			expectedType: antlr.ChronoconePlanningLanguageLexerAFTER,
			keywordText:  "after",
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

			var found bool
			for i := 0; i < tokens.Size(); i++ {
				token := tokens.Get(i)
				if token.GetText() == tc.keywordText {
					found = true
					assert.Equal(t, tc.expectedType, token.GetTokenType(), "keyword '%s' should have expected token type", tc.keywordText)
					break
				}
			}
			assert.True(t, found, "token stream should contain keyword '%s'", tc.keywordText)
		})
	}
}

// TestLexerBeforeAfterCaseSensitivity verifies that mixed-case and uppercase
// variants of 'before' and 'after' are NOT recognized as reserved keywords.
//
// What is tested:
//   - Inputs: "Before", "BEFORE", "After", "AFTER"
//   - Expected behavior: each variant yields an IDENTIFIER token, not BEFORE or AFTER.
//
// Why this test is important:
//   - ANTLR keyword literals are case-sensitive by default. Other CPL keywords
//     such as 'Values' and 'Int' are capitalized, so the case rules for 'before'
//     and 'after' are not obvious without an explicit test.
//   - If a future grammar change accidentally makes keyword matching case-insensitive,
//     concept or value names such as "Before" or "After" would be rejected as
//     invalid identifiers, breaking backward compatibility.
//   - This guards against regressions in keyword-vs-identifier resolution.
func TestLexerBeforeAfterCaseSensitivity(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Before mixed case is identifier",
			input: "Before",
		},
		{
			name:  "BEFORE uppercase is identifier",
			input: "BEFORE",
		},
		{
			name:  "After mixed case is identifier",
			input: "After",
		},
		{
			name:  "AFTER uppercase is identifier",
			input: "AFTER",
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
			assert.Equal(t, tc.input, firstToken.GetText(), "token text should preserve original casing")
			assert.Equal(t, antlr.ChronoconePlanningLanguageLexerIDENTIFIER, firstToken.GetTokenType(), "mixed-case or uppercase variant should be an IDENTIFIER")
			assert.NotEqual(t, antlr.ChronoconePlanningLanguageLexerBEFORE, firstToken.GetTokenType(), "token should not be BEFORE")
			assert.NotEqual(t, antlr.ChronoconePlanningLanguageLexerAFTER, firstToken.GetTokenType(), "token should not be AFTER")
		})
	}
}

// TestParserBeforeClause verifies that the parser recognizes a conceptExtension
// production when the BEFORE keyword is used to specify an insertion point.
//
// What is tested:
//   - Input: a complete concept extension line such as
//     "Priority=[Values[Critical]] before Low" inside a concepts section.
//   - Expected behavior: the parser produces a parse tree that contains a
//     conceptExtension node with an insertionPoint child whose BEFORE token
//     references the anchor identifier "Low".
//
// Why this test is important:
//   - BEFORE is the primary syntactic mechanism for inserting new concept values
//     before an existing value without reordering the original set.
//   - A failure here indicates that the CPLConcepts.g4 grammar change
//     (conceptExtension / insertionPoint) was not correctly integrated into the
//     generated parser, or that the lexer is not emitting BEFORE tokens in
//     context.
//   - This is the foundational parser test that must pass before listener logic
//     (Task 5) can extract InsertionPoint data and store it in the symbol table.
func TestParserBeforeClause(t *testing.T) {
	t.Parallel()

	input := `concepts
	Priority=[Values[Critical]] before Low`

	inputStream := antlrruntime.NewInputStream(input)
	lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)
	tokenStream := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
	p := antlr.NewChronoconePlanningLanguageParser(tokenStream)

	doc := p.Document()
	require.NotNil(t, doc, "Document should parse successfully")

	statements := doc.(*antlr.DocumentContext).AllStatement()
	require.NotEmpty(t, statements, "Document should have at least one statement")

	conceptsSection := statements[0].(*antlr.StatementContext).ConceptsSection()
	require.NotNil(t, conceptsSection, "Statement should be a concepts section")

	conceptDefs := conceptsSection.AllConceptDefinition()
	require.NotEmpty(t, conceptDefs, "Concepts section should have at least one concept definition")

	conceptExtensions := conceptDefs[0].AllConceptExtension()
	require.NotEmpty(t, conceptExtensions, "Concept definition should contain a conceptExtension")

	insertionPoint := conceptExtensions[0].InsertionPoint()
	require.NotNil(t, insertionPoint, "ConceptExtension should have an insertionPoint")
	assert.NotNil(t, insertionPoint.BEFORE(), "insertionPoint should have a BEFORE terminal")
	assert.Equal(t, "Low", insertionPoint.Identifier().GetText(), "anchor identifier should be 'Low'")
}

// TestParserAfterClause verifies that the parser recognizes a conceptExtension
// production when the AFTER keyword is used to specify an insertion point.
//
// What is tested:
//   - Input: a complete concept extension line such as
//     "Priority=[Values[Urgent]] after Medium" inside a concepts section.
//   - Expected behavior: the parser produces a parse tree that contains a
//     conceptExtension node with an insertionPoint child whose AFTER token
//     references the anchor identifier "Medium".
//
// Why this test is important:
//   - AFTER is the complementary syntactic mechanism for inserting new concept
//     values after an existing value.
//   - It ensures that the grammar disjunction in insertionPoint (BEFORE | AFTER)
//     resolves correctly for the AFTER branch.
//   - Without this test, a regression that breaks ONLY the AFTER branch could go
//     unnoticed because BEFORE tests might still pass.
func TestParserAfterClause(t *testing.T) {
	t.Parallel()

	input := `concepts
	Priority=[Values[Urgent]] after Medium`

	inputStream := antlrruntime.NewInputStream(input)
	lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)
	tokenStream := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
	p := antlr.NewChronoconePlanningLanguageParser(tokenStream)

	doc := p.Document()
	require.NotNil(t, doc, "Document should parse successfully")

	statements := doc.(*antlr.DocumentContext).AllStatement()
	require.NotEmpty(t, statements, "Document should have at least one statement")

	conceptsSection := statements[0].(*antlr.StatementContext).ConceptsSection()
	require.NotNil(t, conceptsSection, "Statement should be a concepts section")

	conceptDefs := conceptsSection.AllConceptDefinition()
	require.NotEmpty(t, conceptDefs, "Concepts section should have at least one concept definition")

	conceptExtensions := conceptDefs[0].AllConceptExtension()
	require.NotEmpty(t, conceptExtensions, "Concept definition should contain a conceptExtension")

	insertionPoint := conceptExtensions[0].InsertionPoint()
	require.NotNil(t, insertionPoint, "ConceptExtension should have an insertionPoint")
	assert.NotNil(t, insertionPoint.AFTER(), "insertionPoint should have an AFTER terminal")
	assert.Equal(t, "Medium", insertionPoint.Identifier().GetText(), "anchor identifier should be 'Medium'")
}

// TestParserConceptExtensionFullSection verifies that the parser handles
// concept extensions when they appear inside a full concepts section alongside
// initial concept definitions.
//
// What is tested:
//   - Input: a multi-line CPL document containing a concepts section with both
//     an initial definition (e.g., Priority=[Values[Low, Medium, High]]) and
//     an extension line (e.g., Priority=[Values[Critical]] before Low).
//   - Expected behavior: the conceptsSection rule parses successfully, yielding
//     a parse tree with multiple conceptDefinition children where at least one
//     is an initial definition and at least one is a conceptExtension.
//
// Why this test is important:
//   - In real CPL documents, extensions always appear in a later concepts
//     section that references a concept declared earlier; the parser must handle
//     the mixed syntax within the same section.
//   - It validates that the indent handling and the repetition inside
//     conceptDefinition work correctly when both alternatives
//     (initial definition and conceptExtension) are present.
//   - A failure here would block the listener from building a complete
//     ConceptsScope that contains both original and extended values.
func TestParserConceptExtensionFullSection(t *testing.T) {
	t.Parallel()

	input := `concepts
	Priority=[Values[Low, Medium, High]]
	Priority=[Values[Critical]] before Low`

	inputStream := antlrruntime.NewInputStream(input)
	lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)
	tokenStream := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
	p := antlr.NewChronoconePlanningLanguageParser(tokenStream)

	doc := p.Document()
	require.NotNil(t, doc, "Document should parse successfully")

	statements := doc.(*antlr.DocumentContext).AllStatement()
	require.NotEmpty(t, statements, "Document should have at least one statement")

	conceptsSection := statements[0].(*antlr.StatementContext).ConceptsSection()
	require.NotNil(t, conceptsSection, "Statement should be a concepts section")

	conceptDefs := conceptsSection.AllConceptDefinition()
	require.GreaterOrEqual(t, len(conceptDefs), 2, "Concepts section should have at least two concept definitions (one initial, one extension)")

	// Find which conceptDefinition has the conceptExtension
	var foundExtension bool
	var foundInitial bool
	for _, cd := range conceptDefs {
		exts := cd.AllConceptExtension()
		if len(exts) > 0 {
			foundExtension = true
			insertionPoint := exts[0].InsertionPoint()
			require.NotNil(t, insertionPoint, "ConceptExtension should have an insertionPoint")
			assert.NotNil(t, insertionPoint.BEFORE(), "insertionPoint should have a BEFORE terminal")
			assert.Equal(t, "Low", insertionPoint.Identifier().GetText(), "anchor identifier should be 'Low'")
		}
		lines := cd.AllConceptLine()
		if len(lines) > 0 {
			foundInitial = true
		}
	}
	assert.True(t, foundExtension, "One conceptDefinition should contain a conceptExtension")
	assert.True(t, foundInitial, "One conceptDefinition should contain a generatorPatternList (initial definition)")
}

// TestParserInitialDefinitionStillWorks is a regression test that verifies
// initial concept definitions (without before/after) continue to parse correctly
// after the grammar was modified to support conceptExtension.
//
// What is tested:
//   - Input: a canonical initial definition such as
//     "Priority=[Values[Low, Medium, High]]" inside a concepts section.
//   - Expected behavior: the parser matches the first alternative of
//     conceptDefinition (identifier EQUALS LBRACKET generatorPatternList RBRACKET)
//     and does NOT produce a conceptExtension or insertionPoint node.
//
// Why this test is important:
//   - The grammar change added a second alternative (conceptExtension) that
//     shares the same prefix as the initial definition; ANTLR must resolve this
//     ambiguity correctly via lookahead.
//   - If the parser mistakenly requires an insertion point for every concept
//     line, all existing CPL documents using Increment 1-5 syntax would break.
//   - This guards against regressions in the core concepts parsing path that
//     could affect every existing concept definition in the system.
func TestParserInitialDefinitionStillWorks(t *testing.T) {
	t.Parallel()

	input := `concepts
	Priority=[Values[Low, Medium, High]]`

	inputStream := antlrruntime.NewInputStream(input)
	lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)
	tokenStream := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
	p := antlr.NewChronoconePlanningLanguageParser(tokenStream)

	doc := p.Document()
	require.NotNil(t, doc, "Document should parse successfully")

	statements := doc.(*antlr.DocumentContext).AllStatement()
	require.NotEmpty(t, statements, "Document should have at least one statement")

	conceptsSection := statements[0].(*antlr.StatementContext).ConceptsSection()
	require.NotNil(t, conceptsSection, "Statement should be a concepts section")

	conceptDefs := conceptsSection.AllConceptDefinition()
	require.NotEmpty(t, conceptDefs, "Concepts section should have at least one concept definition")

	// Should have no ConceptExtension children (it's a plain initial definition)
	conceptExtensions := conceptDefs[0].AllConceptExtension()
	assert.Empty(t, conceptExtensions, "Initial definition should not contain a conceptExtension")

	// Should have a ConceptLine (the initial definition's value)
	conceptLines := conceptDefs[0].AllConceptLine()
	require.NotEmpty(t, conceptLines, "Initial definition should have a conceptLine")

	// Verify the generatorPatternList within the conceptLine
	genPatternLists := conceptLines[0].GeneratorPatternList()
	require.NotNil(t, genPatternLists, "Concept line should have a generatorPatternList")

	// Verify the identifier is correct
	assert.Equal(t, "Priority", conceptLines[0].Identifier().GetText(), "Concept name should be 'Priority'")
}

// TestParserBeforeKeywordNotIdentifier verifies that the parser treats
// BEFORE as a reserved keyword token rather than a generic IDENTIFIER
// token when it appears in the insertion point position.
//
// What is tested:
//   - Input: a concept extension line where "before" appears between
//     the closing RBRACKET of the generator pattern list and the anchor value.
//   - Expected behavior: the parser's parse tree shows BEFORE as a distinct
//     keyword terminal; it is not absorbed into an IDENTIFIER token, and the
//     insertionPoint rule matches successfully.
//
// Why this test is important:
//   - If the lexer demotes BEFORE to IDENTIFIER (e.g., due to grammar
//     reordering), the parser would see a sequence like RBRACKET IDENTIFIER
//     IDENTIFIER, which does not match the insertionPoint rule.
//   - This test bridges lexer-level and parser-level verification: the lexer
//     tests confirm token types in isolation, while this test confirms those
//     token types are respected during rule matching.
//   - A failure here indicates a mismatch between the lexer token stream and
//     the parser's expected token types, which would silently reject all valid
//     concept extension syntax.
func TestParserBeforeKeywordNotIdentifier(t *testing.T) {
	t.Parallel()

	input := `concepts
	Priority=[Values[Critical]] before Low`

	inputStream := antlrruntime.NewInputStream(input)
	lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)
	tokenStream := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
	p := antlr.NewChronoconePlanningLanguageParser(tokenStream)

	doc := p.Document()
	require.NotNil(t, doc, "Document should parse successfully")

	statements := doc.(*antlr.DocumentContext).AllStatement()
	require.NotEmpty(t, statements, "Document should have at least one statement")

	conceptsSection := statements[0].(*antlr.StatementContext).ConceptsSection()
	require.NotNil(t, conceptsSection, "Statement should be a concepts section")

	conceptDefs := conceptsSection.AllConceptDefinition()
	require.NotEmpty(t, conceptDefs, "Concepts section should have at least one concept definition")

	conceptExtensions := conceptDefs[0].AllConceptExtension()
	require.NotEmpty(t, conceptExtensions, "Concept definition should contain a conceptExtension")

	insertionPoint := conceptExtensions[0].InsertionPoint()
	require.NotNil(t, insertionPoint, "ConceptExtension should have an insertionPoint")

	// Verify BEFORE is present as a keyword terminal
	beforeToken := insertionPoint.BEFORE()
	assert.NotNil(t, beforeToken, "BEFORE should be present as a keyword terminal")
	assert.Equal(t, "before", beforeToken.GetText(), "token text should be 'before'")

	// The anchor identifier should be "Low" and not the BEFORE token
	anchorId := insertionPoint.Identifier()
	assert.Equal(t, "Low", anchorId.GetText(), "anchor identifier should be 'Low'")
}

// TestParserFallbackWithoutInsertionPoint verifies the parser's behavior when a
// concept line lacks an insertion point and therefore cannot match the
// conceptExtension alternative.
//
// What is tested:
//   - Input: a concept line that structurally resembles an extension but omits
//     both before and after, such as "Priority=[Values[X]]" inside a concepts
//     section.
//   - Expected behavior: the parser falls back to the initial definition
//     alternative and parses the line successfully without error; it does NOT
//     attempt to force a conceptExtension match and then fail.
//
// Why this test is important:
//   - Because conceptExtension and the initial definition share the same prefix,
//     ANTLR must use lookahead to decide between them. If the decision logic is
//     flawed, a line without before/after could trigger a parse error instead of
//     correctly matching the initial definition.
//   - This ensures that the grammar ambiguity was resolved correctly and that
//     existing syntax remains backward compatible.
//   - It also documents the expected fallback behavior, making future grammar
//     changes safer by establishing the baseline for how missing insertion points
//     are handled.
func TestParserFallbackWithoutInsertionPoint(t *testing.T) {
	t.Parallel()

	input := `concepts
	Priority=[Values[X]]`

	inputStream := antlrruntime.NewInputStream(input)
	lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)
	tokenStream := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
	p := antlr.NewChronoconePlanningLanguageParser(tokenStream)

	doc := p.Document()
	require.NotNil(t, doc, "Document should parse successfully (falls back to initial definition)")

	statements := doc.(*antlr.DocumentContext).AllStatement()
	require.NotEmpty(t, statements, "Document should have at least one statement")

	conceptsSection := statements[0].(*antlr.StatementContext).ConceptsSection()
	require.NotNil(t, conceptsSection, "Statement should be a concepts section")

	conceptDefs := conceptsSection.AllConceptDefinition()
	require.NotEmpty(t, conceptDefs, "Concepts section should have at least one concept definition")

	// Should have no ConceptExtension children — parsed as initial definition
	conceptExtensions := conceptDefs[0].AllConceptExtension()
	assert.Empty(t, conceptExtensions, "Line without before/after should not produce a conceptExtension")

	// Should have a ConceptLine (initial definition fallback)
	conceptLines := conceptDefs[0].AllConceptLine()
	require.NotEmpty(t, conceptLines, "Fallback initial definition should have a conceptLine")

	// Verify the conceptLine has a GeneratorPatternList
	genPatternLists := conceptLines[0].GeneratorPatternList()
	require.NotNil(t, genPatternLists, "Fallback concept line should have a generatorPatternList")
}
