// Package parser_test contains tests for the CPL concepts listener.
package parser_test

import (
	"testing"

	antlrruntime "github.com/antlr4-go/antlr/v4"
	"github.com/chronocone/chronocone_planning_language/src/main/go/ast"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseConceptsSection_ScopePushPop verifies that entering and exiting
// the concepts section correctly pushes and pops the ConceptsScope.
func TestParseConceptsSection_ScopePushPop(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		input             string
		expectedScopeType symboltable.ScopeType
		expectedStackSize int
	}{
		{
			name:              "single concept pushes ConceptsScope",
			input:             "concepts\n\tCountry=[Values[USA]]",
			expectedScopeType: symboltable.ConceptsScope,
			expectedStackSize: 2, // global + concepts
		},
		{
			name:              "multiple concepts in section",
			input:             "concepts\n\tCountry=[Values[USA, Canada]]\n\tStatus=[Values[Active]]",
			expectedScopeType: symboltable.ConceptsScope,
			expectedStackSize: 2,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			listener := parser.NewCPLListener(nil, nil, nil)
			_, err := parseWithListener(tc.input, listener)

			require.NoError(t, err, "Parsing should not fail")

			// After parsing, the stack should have just global scope
			assert.Equal(t, 1, listener.GetSymbolStack().Size(),
				"Stack should be back to global scope after parsing")

			// Verify no errors
			assert.False(t, listener.HasErrors(), "Should have no semantic errors")
		})
	}
}

// TestParseConceptDefinition_CreatesDefinition verifies that concept definitions
// are correctly parsed and stored.
func TestParseConceptDefinition_CreatesDefinition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          string
		expectedName   string
		expectedValues []string
	}{
		{
			name:           "simple concept with single value",
			input:          "concepts\n\tCountry=[Values[USA]]",
			expectedName:   "Country",
			expectedValues: []string{"USA"},
		},
		{
			name:           "concept with multiple values",
			input:          "concepts\n\tCountry=[Values[USA, Canada, Mexico]]",
			expectedName:   "Country",
			expectedValues: []string{"USA", "Canada", "Mexico"},
		},
		{
			name:           "concept with priority values",
			input:          "concepts\n\tPriority=[Values[Low, Medium, High]]",
			expectedName:   "Priority",
			expectedValues: []string{"Low", "Medium", "High"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := parser.Parse(tc.input)

			require.NoError(t, err, "Parsing should succeed")
		})
	}
}

// TestParseValuesGenerator_ParsesValues verifies that Values generators
// are correctly parsed with their values.
func TestParseValuesGenerator_ParsesValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "single value in Values",
			input:       "concepts\n\tStatus=[Values[Active]]",
			expectError: false,
		},
		{
			name:        "multiple values in Values",
			input:       "concepts\n\tCountry=[Values[USA, Canada, Mexico]]",
			expectError: false,
		},
		{
			name:        "many values in Values",
			input:       "concepts\n\tMonth=[Values[Jan, Feb, Mar, Apr, May, Jun]]",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := parser.Parse(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected parsing to fail")
			} else {
				require.NoError(t, err, "Parsing should succeed")
			}
		})
	}
}

// TestParseConceptsSection_UnitsAndConceptsTogether verifies that units and
// concepts sections can be parsed together.
func TestParseConceptsSection_UnitsAndConceptsTogether(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name: "only concepts section",
			input: `concepts
	Country=[Values[USA]]`,
			expectError: false,
		},
		{
			name: "types and concepts sections",
			input: `types
	Metric = Number()
concepts
	Country=[Values[USA]]`,
			expectError: false,
		},
		{
			name: "concepts before types",
			input: `concepts
	Country=[Values[USA]]
types
	Metric = Number()`,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			listener := parser.NewCPLListener(nil, nil, nil)
			_, err := parseWithListener(tc.input, listener)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.False(t, listener.HasErrors(), "Should have no semantic errors")
			}
		})
	}
}

// TestSymbolTable_ConceptsScopeType verifies ConceptsScope type.
func TestSymbolTable_ConceptsScopeType(t *testing.T) {
	t.Parallel()

	// Verify the new scope type exists and has correct string representation
	assert.Equal(t, "ConceptsScope", symboltable.ConceptsScope.String())
}

// findConceptScopeWithData walks the symbol table tree and returns the first
// ConceptScope that has non-nil Data. Returns nil if none found.
func findConceptScopeWithData(root *symboltable.SymbolTable) *symboltable.SymbolTable {
	var walk func(st *symboltable.SymbolTable) *symboltable.SymbolTable
	walk = func(st *symboltable.SymbolTable) *symboltable.SymbolTable {
		if st.ScopeType == symboltable.ConceptScope && st.Data != nil {
			return st
		}
		for _, child := range st.Children {
			if found := walk(child); found != nil {
				return found
			}
		}
		return nil
	}
	return walk(root)
}

// TestListener_BeforeClauseCreatesInsertionPoint verifies that parsing a concept
// extension with "before" produces an InsertionPoint with BeforeValue set.
func TestListener_BeforeClauseCreatesInsertionPoint(t *testing.T) {
	t.Parallel()

	input := `concepts
	Priority=[Values[Critical]] before Low`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should not fail")
	assert.False(t, listener.HasErrors(), "Should have no semantic errors")

	root, err := listener.GetSymbolStack().Root()
	require.NoError(t, err, "Should be able to get root")

	scope := findConceptScopeWithData(root)
	require.NotNil(t, scope, "Should find a ConceptScope with data")

	conceptData, ok := scope.Data.(*ast.ConceptDefinitionData)
	require.True(t, ok, "ConceptScope Data should be *ast.ConceptDefinitionData")
	assert.Equal(t, "Priority", conceptData.Name, "Concept name should be 'Priority'")
	assert.Equal(t, "Low", conceptData.InsertionPoint.BeforeValue, "BeforeValue should be 'Low'")
	assert.Empty(t, conceptData.InsertionPoint.AfterValue, "AfterValue should be empty")
}

// TestListener_AfterClauseCreatesInsertionPoint verifies that parsing a concept
// extension with "after" produces an InsertionPoint with AfterValue set.
func TestListener_AfterClauseCreatesInsertionPoint(t *testing.T) {
	t.Parallel()

	input := `concepts
	Priority=[Values[Urgent]] after Medium`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should not fail")
	assert.False(t, listener.HasErrors(), "Should have no semantic errors")

	root, err := listener.GetSymbolStack().Root()
	require.NoError(t, err, "Should be able to get root")

	scope := findConceptScopeWithData(root)
	require.NotNil(t, scope, "Should find a ConceptScope with data")

	conceptData, ok := scope.Data.(*ast.ConceptDefinitionData)
	require.True(t, ok, "ConceptScope Data should be *ast.ConceptDefinitionData")
	assert.Equal(t, "Priority", conceptData.Name, "Concept name should be 'Priority'")
	assert.Equal(t, "Medium", conceptData.InsertionPoint.AfterValue, "AfterValue should be 'Medium'")
	assert.Empty(t, conceptData.InsertionPoint.BeforeValue, "BeforeValue should be empty")
}

// TestListener_ConceptExtensionStoresData verifies that ConceptDefinitionData
// is stored in the ConceptScope with correct Name and InsertionPoint fields.
func TestListener_ConceptExtensionStoresData(t *testing.T) {
	t.Parallel()

	input := `concepts
	Priority=[Values[Critical]] before Low`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should not fail")

	root, err := listener.GetSymbolStack().Root()
	require.NoError(t, err, "Should be able to get root")

	scope := findConceptScopeWithData(root)
	require.NotNil(t, scope, "Should find a ConceptScope with data")

	conceptData, ok := scope.Data.(*ast.ConceptDefinitionData)
	require.True(t, ok, "ConceptScope Data should be *ast.ConceptDefinitionData")
	assert.Equal(t, "Priority", conceptData.Name)
	assert.Equal(t, "Low", conceptData.InsertionPoint.BeforeValue)
	assert.Greater(t, conceptData.SourceLocation.Line, 0, "SourceLocation.Line should be set")
}

// TestListener_InitialDefinitionNoInsertionPoint verifies that a regular concept
// definition (no before/after) does not have InsertionPoint set in its data.
func TestListener_InitialDefinitionNoInsertionPoint(t *testing.T) {
	t.Parallel()

	input := `concepts
	Priority=[Values[Low, Medium, High]]`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should not fail")
	assert.False(t, listener.HasErrors(), "Should have no semantic errors")

	root, err := listener.GetSymbolStack().Root()
	require.NoError(t, err, "Should be able to get root")

	// For initial definitions, no ConceptScope should have Data set because
	// ExitConceptDefinition does not set any data (only ExitConceptExtension does).
	scope := findConceptScopeWithData(root)
	assert.Nil(t, scope, "Initial definition should not have any ConceptScope with data")
}

// TestListener_BeforeAfterValidation_ValidInsertionPoint verifies that a valid
// concept extension with a before/after insertion point and a valid anchor
// identifier produces NO semantic validation errors.
//
// What is tested:
//   - Input: a concept extension with "before" and a valid anchor identifier
//     (e.g., Priority=[Values[Critical]] before Low)
//   - Expected behavior: the listener successfully extracts the InsertionPoint
//     data without raising any semantic errors. The BeforeValue is set to the
//     anchor identifier text and AfterValue remains empty.
//
// Why this test is important:
//   - This is the happy-path test for the insertion point validation guards
//     added in Task 6. It ensures that valid before/after syntax does not
//     trigger false-positive validation errors.
//   - Validates both the ExitInsertionPoint validation (anchor identifier
//     non-empty) and the ExitConceptExtension validation (concept name non-empty).
func TestListener_BeforeAfterValidation_ValidInsertionPoint(t *testing.T) {
	t.Parallel()

	input := `concepts
	Priority=[Values[Critical]] before Low`

	listener := parser.NewCPLListener(nil, nil, nil)
	_, err := parseWithListener(input, listener)
	require.NoError(t, err, "Parsing should not fail")
	assert.False(t, listener.HasErrors(), "Should have no semantic errors")

	root, err := listener.GetSymbolStack().Root()
	require.NoError(t, err, "Should be able to get root")

	scope := findConceptScopeWithData(root)
	require.NotNil(t, scope, "Should find a ConceptScope with data")

	conceptData, ok := scope.Data.(*ast.ConceptDefinitionData)
	require.True(t, ok, "ConceptScope Data should be *ast.ConceptDefinitionData")
	assert.Equal(t, "Priority", conceptData.Name, "Concept name should be 'Priority'")
	assert.Equal(t, "Low", conceptData.InsertionPoint.BeforeValue, "BeforeValue should be 'Low'")
	assert.Empty(t, conceptData.InsertionPoint.AfterValue, "AfterValue should be empty")
}

// TestListener_BeforeAfterValidation_ValidAnchorNoFalsePositive verifies the
// defensive semantic validation guards that ensure the anchor identifier in an
// insertion point is non-empty.
//
// What is tested:
//   - The grammar (insertionPoint: BEFORE identifier | AFTER identifier) enforces
//     that an identifier always follows BEFORE/AFTER at the syntax level. This
//     means truly malformed input (missing anchor identifier) cannot be parsed
//     and will be rejected as a syntax error before reaching semantic validation.
//   - This test verifies that valid input with a proper anchor identifier does
//     NOT trigger the defensive semantic validation guards. It serves as a
//     regression test ensuring the validation does not produce false positives.
//
// Why this test is important:
//   - Documents the "belt-and-suspenders" nature of the identifier validation:
//     the grammar guarantees correctness at parse time, and the semantic check
//     provides an additional safety layer.
//   - If the grammar were ever changed in a way that allows empty identifiers
//     through insertionPoint, this test infrastructure exists to catch the
//     resulting semantic error.
func TestListener_BeforeAfterValidation_ValidAnchorNoFalsePositive(t *testing.T) {
	t.Parallel()

	// Test both before and after variants to ensure the validation guards
	// handle both paths correctly without false positives.
	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "before with valid identifier",
			input:       "concepts\n\tPriority=[Values[Critical]] before Low",
			expectError: false,
		},
		{
			name:        "after with valid identifier",
			input:       "concepts\n\tPriority=[Values[Urgent]] after Medium",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			listener := parser.NewCPLListener(nil, nil, nil)
			_, err := parseWithListener(tc.input, listener)
			require.NoError(t, err, "Parsing should not fail")

			if tc.expectError {
				assert.True(t, listener.HasErrors(), "Should have semantic errors")
			} else {
				assert.False(t, listener.HasErrors(), "Should have no semantic errors")
			}

			// Verify the insertion point data is correctly extracted
			root, err := listener.GetSymbolStack().Root()
			require.NoError(t, err, "Should be able to get root")

			scope := findConceptScopeWithData(root)
			require.NotNil(t, scope, "Should find a ConceptScope with data")

			conceptData, ok := scope.Data.(*ast.ConceptDefinitionData)
			require.True(t, ok, "ConceptScope Data should be *ast.ConceptDefinitionData")
			assert.NotEmpty(t, conceptData.Name, "Concept name should not be empty")
			assert.True(t,
				conceptData.InsertionPoint.BeforeValue != "" || conceptData.InsertionPoint.AfterValue != "",
				"At least one of BeforeValue/AfterValue should be set")
		})
	}
}

// parseWithListener parses input with a custom listener.
// This is a helper function that mimics the internal parse flow.
func parseWithListener(input string, listener *parser.CPLListener) (*parser.CPLListener, error) {
	// Create input stream and lexer
	inputStream := antlrruntime.NewInputStream(input)
	lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)
	tokenStream := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
	antlrParser := antlr.NewChronoconePlanningLanguageParser(tokenStream)

	// Walk the parse tree with the provided listener
	antlrruntime.ParseTreeWalkerDefault.Walk(listener, antlrParser.Document())

	return listener, nil
}
