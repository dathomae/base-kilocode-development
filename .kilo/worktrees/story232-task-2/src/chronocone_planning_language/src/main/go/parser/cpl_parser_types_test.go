// Package parser_test contains tests for the CPL parser.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseNumberTypeWithPrecisionConstraint tests parsing Number type with precision constraint only.
// Note: In the symbol table driven parsing approach, type definitions are stored as identifiers
// in TypeScope children of the TypesScope. The detailed extraction of type data
// (BuiltInType, Constraints) is deferred to the analysis phase (Story132).
func TestParseNumberTypeWithPrecisionConstraint(t *testing.T) {
	t.Parallel()

	input := "types\n\tCurrency = Number(precision=2)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have a TypesScope child
	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	assert.Equal(t, symboltable.TypesScope, typesScope.ScopeType, "Child scope should be TypesScope")

	// Verify we have a TypeScope child containing the type definition
	require.Len(t, typesScope.Children, 1, "Expected one TypeScope child")
	typeScope := typesScope.Children[0]
	assert.Equal(t, symboltable.TypeScope, typeScope.ScopeType, "Child should be TypeScope")

	// Verify identifiers are present in the TypeScope
	// The type name and base type name are added by EnterIdentifier
	require.GreaterOrEqual(t, len(typeScope.Definitions), 1, "Expected at least one identifier definition")

	// Look up the type by name in the TypeScope
	typeDef, err := typeScope.Lookup("Currency")
	require.NoError(t, err, "Currency should be found in TypeScope")
	assert.NotNil(t, typeDef.Name)
	assert.Equal(t, "Currency", *typeDef.Name, "Type name should be Currency")

	// Data field is nil at parser phase - analysis phase will populate it
	assert.Nil(t, typeDef.Data, "Type definition data should be nil at parser phase")
}

// TestParseNumberTypeWithMultipleConstraints tests parsing Number type with precision, min, and max constraints.
// Note: In the symbol table driven parsing approach, type definitions are stored as identifiers
// in TypeScope children of the TypesScope. The detailed extraction and validation of type data
// is deferred to the analysis phase.
func TestParseNumberTypeWithMultipleConstraints(t *testing.T) {
	t.Parallel()

	input := "types\n\tPercentage = Number(precision=2, min=0, max=100)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have a TypesScope child
	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	assert.Equal(t, symboltable.TypesScope, typesScope.ScopeType, "Child scope should be TypesScope")

	// Verify we have a TypeScope child containing the type definition
	require.Len(t, typesScope.Children, 1, "Expected one TypeScope child")
	typeScope := typesScope.Children[0]
	assert.Equal(t, symboltable.TypeScope, typeScope.ScopeType, "Child should be TypeScope")

	// Look up the type by name in the TypeScope
	typeDef, err := typeScope.Lookup("Percentage")
	require.NoError(t, err, "Percentage should be found in TypeScope")
	assert.NotNil(t, typeDef.Name)
	assert.Equal(t, "Percentage", *typeDef.Name, "Type name should be Percentage")

	// Data field is nil at parser phase - analysis phase will populate it
	assert.Nil(t, typeDef.Data, "Type definition data should be nil at parser phase")
}

// TestParseNumberTypeWithNoConstraints tests parsing Number type with no constraints.
// Note: In the symbol table driven parsing approach, type definitions are stored as identifiers
// in TypeScope children of the TypesScope. The detailed extraction of type data
// is deferred to the analysis phase.
func TestParseNumberTypeWithNoConstraints(t *testing.T) {
	t.Parallel()

	input := "types\n\tAmount = Number()"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have a TypesScope child
	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	assert.Equal(t, symboltable.TypesScope, typesScope.ScopeType, "Child scope should be TypesScope")

	// Verify we have a TypeScope child containing the type definition
	require.Len(t, typesScope.Children, 1, "Expected one TypeScope child")
	typeScope := typesScope.Children[0]
	assert.Equal(t, symboltable.TypeScope, typeScope.ScopeType, "Child should be TypeScope")

	// Look up the type by name in the TypeScope
	typeDef, err := typeScope.Lookup("Amount")
	require.NoError(t, err, "Amount should be found in TypeScope")
	assert.NotNil(t, typeDef.Name)
	assert.Equal(t, "Amount", *typeDef.Name, "Type name should be Amount")

	// Data field is nil at parser phase - analysis phase will populate it
	assert.Nil(t, typeDef.Data, "Type definition data should be nil at parser phase")
}

// TestParseMultipleTypeDefinitions tests parsing multiple type definitions in one section.
// Note: In the symbol table driven parsing approach, each type definition has its own
// TypeScope child of the TypesScope.
func TestParseMultipleTypeDefinitions(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)
	Percentage = Number(precision=2, min=0, max=100)
	Amount = Number()`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have a TypesScope child
	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	assert.Equal(t, symboltable.TypesScope, typesScope.ScopeType, "Child scope should be TypesScope")

	// Verify we have three TypeScope children (one per type definition)
	require.Len(t, typesScope.Children, 3, "Expected three TypeScope children")

	// Verify each type definition
	typeNames := make(map[string]bool)
	for _, typeScope := range typesScope.Children {
		assert.Equal(t, symboltable.TypeScope, typeScope.ScopeType, "Each child should be TypeScope")
		for _, def := range typeScope.Definitions {
			if def.Name != nil {
				typeNames[*def.Name] = true
			}
		}
	}

	assert.True(t, typeNames["Currency"], "Should have Currency type")
	assert.True(t, typeNames["Percentage"], "Should have Percentage type")
	assert.True(t, typeNames["Amount"], "Should have Amount type")
}

// TestParseInvalidBaseType tests that unknown base types parse successfully.
//
// Note: This test verifies the PARSER behavior, not analysis phase behavior.
// The grammar accepts identifiers as base types (for type derivation), so
// "Integer" parses successfully. However, the ANALYSIS phase will detect that
// "Integer" doesn't terminate in a built-in type and report an error there.
// This separation is intentional - the parser handles syntax, analysis handles semantics.
func TestParseInvalidBaseType(t *testing.T) {
	t.Parallel()

	// Test with unknown base type - parser accepts identifiers as base types
	// The grammar allows: baseType : identifier
	// So "Integer" is a valid identifier and parses successfully.
	// The analysis phase will detect that "Integer" doesn't derive from a built-in type.
	input := "types\n\tCurrency = Integer(precision=2)"

	doc, err := parser.Parse(input)

	// Parser accepts identifier as base type, so this succeeds at parse level
	require.NoError(t, err, "Parser should accept identifier as base type (analysis will detect semantic issues)")
	require.NotNil(t, doc, "Document should not be nil")
}

// TestParseTypesInvalidSyntax tests error handling for invalid syntax patterns in types section.
func TestParseTypesInvalidSyntax(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name:          "missing parentheses",
			input:         "types\n\tCurrency = Number",
			expectError:   true,
			errorContains: "expecting '('",
		},
		{
			name:          "missing closing parenthesis",
			input:         "types\n\tCurrency = Number(precision=2",
			expectError:   true,
			errorContains: "missing ')'",
		},
		{
			name:          "missing opening parenthesis",
			input:         "types\n\tCurrency = Numberprecision=2)",
			expectError:   true,
			errorContains: "mismatched input",
		},
		{
			name:          "invalid identifier",
			input:         "types\n\t123invalid = Number()",
			expectError:   true,
			errorContains: "extraneous input '123invalid'",
		},
		{
			name:          "missing base type",
			input:         "types\n\tCurrency = ()",
			expectError:   true,
			errorContains: "missing",
		},
		{
			name:          "empty constraint name",
			input:         "types\n\tCurrency = Number(=2)",
			expectError:   true,
			errorContains: "mismatched",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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

// TestParseTypesScopeIntegration tests that TypesScope is properly pushed/popped.
func TestParseTypesScopeIntegration(t *testing.T) {
	t.Parallel()

	// Parse a simple types section with proper indentation
	input := "types\n\tCurrency = Number(precision=2)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error")
	require.NotNil(t, doc, "Document should not be nil")

	// Verify TypesScope was created
	require.Len(t, doc.Children, 1, "Expected one child scope")
	typesScope := doc.Children[0]
	assert.Equal(t, symboltable.TypesScope, typesScope.ScopeType, "Child scope should be TypesScope")

	// Verify the typesScope parent is the GlobalScope
	assert.Equal(t, doc, typesScope.Parent, "TypesScope parent should be GlobalScope")

	// Verify the TypeScope child exists with the type definition
	require.Len(t, typesScope.Children, 1, "Should have one TypeScope child")
	typeScope := typesScope.Children[0]
	assert.Equal(t, symboltable.TypeScope, typeScope.ScopeType, "Child scope should be TypeScope")
}

// TestParseTypesSymbolTableIntegration tests that type definitions are stored properly in symbol table.
// Note: Type definitions are stored in TypeScope children, not directly in TypesScope.
func TestParseTypesSymbolTableIntegration(t *testing.T) {
	t.Parallel()

	// Parse multiple type definitions
	input := `types
	Currency = Number(precision=2)
	Percentage = Number(precision=2, min=0, max=100)`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error")
	require.NotNil(t, doc, "Document should not be nil")

	// Get the TypesScope
	typesScope := doc.Children[0]
	require.Equal(t, symboltable.TypesScope, typesScope.ScopeType)

	// Verify both TypeScope children exist
	require.Len(t, typesScope.Children, 2, "Should have two TypeScope children")

	// Find the type scopes and verify types can be looked up
	currencyFound := false
	percentageFound := false

	for _, typeScope := range typesScope.Children {
		currencyDef, err := typeScope.Lookup("Currency")
		if err == nil && currencyDef != nil {
			currencyFound = true
			assert.NotNil(t, currencyDef.Name)
			assert.Equal(t, "Currency", *currencyDef.Name)
		}

		percentageDef, err := typeScope.Lookup("Percentage")
		if err == nil && percentageDef != nil {
			percentageFound = true
			assert.NotNil(t, percentageDef.Name)
			assert.Equal(t, "Percentage", *percentageDef.Name)
		}
	}

	assert.True(t, currencyFound, "Currency should be found in one of the TypeScope children")
	assert.True(t, percentageFound, "Percentage should be found in one of the TypeScope children")

	// Verify lookup fails for undefined types in each TypeScope
	for _, typeScope := range typesScope.Children {
		_, err := typeScope.Lookup("UndefinedType")
		assert.Error(t, err, "UndefinedType should not be found")
	}
}

// TestParseTypesWithWhitespaceVariations tests parsing with various whitespace configurations.
func TestParseTypesWithWhitespaceVariations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "tabs for indentation",
			input: "types\n\tCurrency = Number(precision=2)",
		},
		{
			name:  "spaces for indentation",
			input: "types\n    Currency = Number(precision=2)",
		},
		{
			name:  "multiple spaces between constraints",
			input: "types\n\tCurrency = Number(precision=2,  min=0,  max=100)",
		},
		{
			name:  "spaces around equals",
			input: "types\n\tCurrency = Number(precision=2)",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := parser.Parse(tc.input)

			require.NoError(t, err, "Expected no error for input: %s", tc.input)
			require.NotNil(t, doc, "Document should not be nil")

			// Verify TypesScope was created
			require.Len(t, doc.Children, 1, "Expected one child scope")
		})
	}
}

// TestParseTypesSourceLocation tests that source locations are captured correctly.
func TestParseTypesSourceLocation(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)
	Percentage = Number(precision=2, min=0, max=100)`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error")
	require.NotNil(t, doc, "Document should not be nil")

	// Get the TypesScope
	typesScope := doc.Children[0]

	// Verify source locations are present
	for _, def := range typesScope.Definitions {
		assert.Greater(t, def.Location.Line, 0, "Source location line should be greater than 0")
		assert.GreaterOrEqual(t, def.Location.Column, 0, "Source location column should be >= 0")
	}
}

// TestParseInstantTypeWithSecondConstraint tests parsing Instant type with second=0 constraint.
func TestParseInstantTypeWithSecondConstraint(t *testing.T) {
	t.Parallel()

	input := "types\n\tTimestamp = Instant(second=0)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have a TypesScope child
	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	assert.Equal(t, symboltable.TypesScope, typesScope.ScopeType, "Child scope should be TypesScope")

	// Verify we have a TypeScope child containing the type definition
	require.Len(t, typesScope.Children, 1, "Expected one TypeScope child")
	typeScope := typesScope.Children[0]
	assert.Equal(t, symboltable.TypeScope, typeScope.ScopeType, "Child should be TypeScope")

	// Look up the type by name in the TypeScope
	typeDef, err := typeScope.Lookup("Timestamp")
	require.NoError(t, err, "Timestamp should be found in TypeScope")
	assert.NotNil(t, typeDef.Name)
	assert.Equal(t, "Timestamp", *typeDef.Name, "Type name should be Timestamp")
}

// TestParseInstantTypeWithYearConstraint tests parsing Instant type with year=2024 constraint.
func TestParseInstantTypeWithYearConstraint(t *testing.T) {
	t.Parallel()

	input := "types\n\tYear2024 = Instant(year=2024)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	require.Len(t, typesScope.Children, 1, "Expected one TypeScope child")

	typeScope := typesScope.Children[0]
	typeDef, err := typeScope.Lookup("Year2024")
	require.NoError(t, err, "Year2024 should be found in TypeScope")
	assert.Equal(t, "Year2024", *typeDef.Name, "Type name should be Year2024")
}

// TestParseInstantTypeWithDayOfWeekConstraint tests parsing Instant type with day_of_week=Monday constraint.
func TestParseInstantTypeWithDayOfWeekConstraint(t *testing.T) {
	t.Parallel()

	input := "types\n\tMondayStart = Instant(day_of_week=Monday)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	require.Len(t, typesScope.Children, 1, "Expected one TypeScope child")

	typeScope := typesScope.Children[0]
	typeDef, err := typeScope.Lookup("MondayStart")
	require.NoError(t, err, "MondayStart should be found in TypeScope")
	assert.Equal(t, "MondayStart", *typeDef.Name, "Type name should be MondayStart")
}

// TestParseInstantTypeWithMultipleConstraints tests parsing Instant type with multiple field constraints.
func TestParseInstantTypeWithMultipleConstraints(t *testing.T) {
	t.Parallel()

	input := "types\n\tNewYear = Instant(month=January, day=1)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	require.Len(t, typesScope.Children, 1, "Expected one TypeScope child")

	typeScope := typesScope.Children[0]
	require.Len(t, typeScope.Children, 2, "Expected two ConstraintScope children")

	typeDef, err := typeScope.Lookup("NewYear")
	require.NoError(t, err, "NewYear should be found in TypeScope")
	assert.Equal(t, "NewYear", *typeDef.Name, "Type name should be NewYear")
}

// TestParseInstantTypeWithInvalidConstraint verifies that invalid constraint names parse successfully.
// Constraint name validation is deferred to the analysis phase.
func TestParseInstantTypeWithInvalidConstraint(t *testing.T) {
	t.Parallel()

	input := "types\n\tBadInstant = Instant(invalid_field=123)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Parser should accept any constraint name (analysis will detect semantic issues)")
	require.NotNil(t, doc, "Document should not be nil")
}

// TestParseInstantTypeWithUnknownConstraintName verifies that unknown constraint names parse successfully.
// Constraint name validation is deferred to the analysis phase.
func TestParseInstantTypeWithUnknownConstraintName(t *testing.T) {
	t.Parallel()

	input := "types\n\tBadInstant = Instant(foobar=42)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Parser should accept any constraint name (analysis will detect semantic issues)")
	require.NotNil(t, doc, "Document should not be nil")
}

// TestParseInstantTypeWithNoConstraints tests parsing Instant type with no constraints.
func TestParseInstantTypeWithNoConstraints(t *testing.T) {
	t.Parallel()

	input := "types\n\tAnyInstant = Instant()"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	require.Len(t, typesScope.Children, 1, "Expected one TypeScope child")

	typeScope := typesScope.Children[0]
	typeDef, err := typeScope.Lookup("AnyInstant")
	require.NoError(t, err, "AnyInstant should be found in TypeScope")
	assert.Equal(t, "AnyInstant", *typeDef.Name, "Type name should be AnyInstant")
}

// TestParseInstantTypeMixedWithNumberType tests parsing both Instant and Number types in the same section.
func TestParseInstantTypeMixedWithNumberType(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)
	Timestamp = Instant(second=0)`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	require.Len(t, typesScope.Children, 2, "Expected two TypeScope children")

	typeNames := make(map[string]bool)
	for _, typeScope := range typesScope.Children {
		for _, def := range typeScope.Definitions {
			if def.Name != nil {
				typeNames[*def.Name] = true
			}
		}
	}

	assert.True(t, typeNames["Currency"], "Should have Currency type")
	assert.True(t, typeNames["Timestamp"], "Should have Timestamp type")
}

// TestParseIntervalTypeWithNamedStartsAtAndUnit tests parsing Interval type with named starts_at type and unit constraints.
func TestParseIntervalTypeWithNamedStartsAtAndUnit(t *testing.T) {
	t.Parallel()

	input := "types\n\tFiscalMonth = Interval(starts_at=FiscalYearStart, unit=month)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	assert.Equal(t, symboltable.TypesScope, typesScope.ScopeType, "Child scope should be TypesScope")

	require.Len(t, typesScope.Children, 1, "Expected one TypeScope child")
	typeScope := typesScope.Children[0]
	assert.Equal(t, symboltable.TypeScope, typeScope.ScopeType, "Child should be TypeScope")

	typeDef, err := typeScope.Lookup("FiscalMonth")
	require.NoError(t, err, "FiscalMonth should be found in TypeScope")
	assert.NotNil(t, typeDef.Name)
	assert.Equal(t, "FiscalMonth", *typeDef.Name, "Type name should be FiscalMonth")

	// Data field is nil at parser phase - analysis phase will populate it
	assert.Nil(t, typeDef.Data, "Type definition data should be nil at parser phase")
}

// TestParseIntervalTypeWithInlineStartsAt tests parsing Interval type with inline Instant constraints.
func TestParseIntervalTypeWithInlineStartsAt(t *testing.T) {
	t.Parallel()

	input := "types\n\tWorkWeek = Interval(starts_at=Instant(day_of_week=Monday), unit=day)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	require.Len(t, typesScope.Children, 1, "Expected one TypeScope child")

	typeScope := typesScope.Children[0]
	typeDef, err := typeScope.Lookup("WorkWeek")
	require.NoError(t, err, "WorkWeek should be found in TypeScope")
	assert.Equal(t, "WorkWeek", *typeDef.Name, "Type name should be WorkWeek")
	require.Len(t, typeScope.Children, 2, "Expected two top-level ConstraintScope children")
}

// TestParseIntervalTypeWithNonInstantInlineStartsAt tests parsing accepts generic nested starts_at constraint values.
func TestParseIntervalTypeWithNonInstantInlineStartsAt(t *testing.T) {
	t.Parallel()

	input := "types\n\tWorkWeek = Interval(starts_at=NotInstant(day_of_week=Monday), unit=day)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Parser should accept generic nested constraint syntax for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	require.Len(t, typesScope.Children, 1, "Expected one TypeScope child")

	typeScope := typesScope.Children[0]
	typeDef, err := typeScope.Lookup("WorkWeek")
	require.NoError(t, err, "WorkWeek should be found in TypeScope")
	assert.Equal(t, "WorkWeek", *typeDef.Name, "Type name should be WorkWeek")
	require.Len(t, typeScope.Children, 2, "Expected two top-level ConstraintScope children")
}

// TestParseIntervalTypeWithNamedStartsAtOnly tests parsing Interval type with only starts_at constraint.
func TestParseIntervalTypeWithNamedStartsAtOnly(t *testing.T) {
	t.Parallel()

	input := "types\n\tWorkDay = Interval(starts_at=FiscalYearStart)"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	require.Len(t, typesScope.Children, 1, "Expected one TypeScope child")

	typeScope := typesScope.Children[0]
	typeDef, err := typeScope.Lookup("WorkDay")
	require.NoError(t, err, "WorkDay should be found in TypeScope")
	assert.Equal(t, "WorkDay", *typeDef.Name, "Type name should be WorkDay")
}

// TestParseIntervalTypeWithNoConstraints tests parsing Interval type with no constraints.
// Constraint presence validation is deferred to the analysis phase (Story140).
func TestParseIntervalTypeWithNoConstraints(t *testing.T) {
	t.Parallel()

	input := "types\n\tAnyInterval = Interval()"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	require.Len(t, typesScope.Children, 1, "Expected one TypeScope child")

	typeScope := typesScope.Children[0]
	typeDef, err := typeScope.Lookup("AnyInterval")
	require.NoError(t, err, "AnyInterval should be found in TypeScope")
	assert.Equal(t, "AnyInterval", *typeDef.Name, "Type name should be AnyInterval")
}

// TestParseIntervalTypeMixedWithOtherTypes tests parsing Interval mixed with Number and Instant in the same section.
func TestParseIntervalTypeMixedWithOtherTypes(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)
	Timestamp = Instant(second=0)
	FiscalMonth = Interval(starts_at=FiscalYearStart, unit=month)`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (TypesScope)")
	typesScope := doc.Children[0]
	require.Len(t, typesScope.Children, 3, "Expected three TypeScope children")

	typeNames := make(map[string]bool)
	for _, typeScope := range typesScope.Children {
		for _, def := range typeScope.Definitions {
			if def.Name != nil {
				typeNames[*def.Name] = true
			}
		}
	}

	assert.True(t, typeNames["Currency"], "Should have Currency type")
	assert.True(t, typeNames["Timestamp"], "Should have Timestamp type")
	assert.True(t, typeNames["FiscalMonth"], "Should have FiscalMonth type")
}

// TestParseTwoTypesSections verifies parsing a file with two separate `types`
// sections produces two TypesScope children under the root document scope,
// each containing their own TypeScope children.
func TestParseTwoTypesSections(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)

types
	Percentage = Number(precision=2, minimum=0, maximum=100)`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input with two types sections")
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have exactly two TypesScope children
	require.Len(t, doc.Children, 2, "Expected two children (two TypesScope)")

	firstTypesScope := doc.Children[0]
	secondTypesScope := doc.Children[1]

	// Verify both are TypesScope
	assert.Equal(t, symboltable.TypesScope, firstTypesScope.ScopeType, "First child should be TypesScope")
	assert.Equal(t, symboltable.TypesScope, secondTypesScope.ScopeType, "Second child should be TypesScope")

	// Verify each TypesScope's parent is the document scope (GlobalScope)
	assert.Equal(t, doc, firstTypesScope.Parent, "First TypesScope parent should be GlobalScope")
	assert.Equal(t, doc, secondTypesScope.Parent, "Second TypesScope parent should be GlobalScope")

	// Verify first TypesScope has at least 1 TypeScope child containing Currency
	// Note: there may be extra empty TypeScope children due to newline handling in the lexer
	firstTypeScopes := filterScopesByType(firstTypesScope.Children, symboltable.TypeScope)
	require.GreaterOrEqual(t, len(firstTypeScopes), 1, "First TypesScope should have at least one TypeScope child")

	// Find the TypeScope containing Currency
	var currencyTypeScope *symboltable.SymbolTable
	for _, ts := range firstTypeScopes {
		if _, err := ts.Lookup("Currency"); err == nil {
			currencyTypeScope = ts
			break
		}
	}
	require.NotNil(t, currencyTypeScope, "Should find a TypeScope containing Currency")
	currencyDef, err := currencyTypeScope.Lookup("Currency")
	require.NoError(t, err, "Currency should be found in first TypeScope")
	assert.NotNil(t, currencyDef.Name)
	assert.Equal(t, "Currency", *currencyDef.Name, "Type name should be Currency")

	// Verify second TypesScope has at least 1 TypeScope child containing Percentage
	secondTypeScopes := filterScopesByType(secondTypesScope.Children, symboltable.TypeScope)
	require.GreaterOrEqual(t, len(secondTypeScopes), 1, "Second TypesScope should have at least one TypeScope child")

	// Find the TypeScope containing Percentage
	var percentageTypeScope *symboltable.SymbolTable
	for _, ts := range secondTypeScopes {
		if _, err := ts.Lookup("Percentage"); err == nil {
			percentageTypeScope = ts
			break
		}
	}
	require.NotNil(t, percentageTypeScope, "Should find a TypeScope containing Percentage")
	percentageDef, err := percentageTypeScope.Lookup("Percentage")
	require.NoError(t, err, "Percentage should be found in second TypeScope")
	assert.NotNil(t, percentageDef.Name)
	assert.Equal(t, "Percentage", *percentageDef.Name, "Type name should be Percentage")
}

// TestParseThreeTypesSections verifies that parsing a file with three separate
// `types` sections produces three TypesScope children under the root document
// scope, each containing their own TypeScope children.
func TestParseThreeTypesSections(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)

types
	Percentage = Number(precision=2, minimum=0, maximum=100)

types
	Amount = Number()`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input with three types sections")
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have exactly three TypesScope children
	require.Len(t, doc.Children, 3, "Expected three children (three TypesScope)")

	firstTypesScope := doc.Children[0]
	secondTypesScope := doc.Children[1]
	thirdTypesScope := doc.Children[2]

	// Verify all are TypesScope
	assert.Equal(t, symboltable.TypesScope, firstTypesScope.ScopeType, "First child should be TypesScope")
	assert.Equal(t, symboltable.TypesScope, secondTypesScope.ScopeType, "Second child should be TypesScope")
	assert.Equal(t, symboltable.TypesScope, thirdTypesScope.ScopeType, "Third child should be TypesScope")

	// Verify each TypesScope's parent is the document scope (GlobalScope)
	assert.Equal(t, doc, firstTypesScope.Parent, "First TypesScope parent should be GlobalScope")
	assert.Equal(t, doc, secondTypesScope.Parent, "Second TypesScope parent should be GlobalScope")
	assert.Equal(t, doc, thirdTypesScope.Parent, "Third TypesScope parent should be GlobalScope")

	// Verify first TypesScope has at least 1 TypeScope child containing Currency
	firstTypeScopes := filterScopesByType(firstTypesScope.Children, symboltable.TypeScope)
	require.GreaterOrEqual(t, len(firstTypeScopes), 1, "First TypesScope should have at least one TypeScope child")

	var currencyTypeScope *symboltable.SymbolTable
	for _, ts := range firstTypeScopes {
		if _, err := ts.Lookup("Currency"); err == nil {
			currencyTypeScope = ts
			break
		}
	}
	require.NotNil(t, currencyTypeScope, "Should find a TypeScope containing Currency")
	currencyDef, err := currencyTypeScope.Lookup("Currency")
	require.NoError(t, err, "Currency should be found in first TypeScope")
	assert.NotNil(t, currencyDef.Name)
	assert.Equal(t, "Currency", *currencyDef.Name, "Type name should be Currency")

	// Verify second TypesScope has at least 1 TypeScope child containing Percentage
	secondTypeScopes := filterScopesByType(secondTypesScope.Children, symboltable.TypeScope)
	require.GreaterOrEqual(t, len(secondTypeScopes), 1, "Second TypesScope should have at least one TypeScope child")

	var percentageTypeScope *symboltable.SymbolTable
	for _, ts := range secondTypeScopes {
		if _, err := ts.Lookup("Percentage"); err == nil {
			percentageTypeScope = ts
			break
		}
	}
	require.NotNil(t, percentageTypeScope, "Should find a TypeScope containing Percentage")
	percentageDef, err := percentageTypeScope.Lookup("Percentage")
	require.NoError(t, err, "Percentage should be found in second TypeScope")
	assert.NotNil(t, percentageDef.Name)
	assert.Equal(t, "Percentage", *percentageDef.Name, "Type name should be Percentage")

	// Verify third TypesScope has at least 1 TypeScope child containing Amount
	thirdTypeScopes := filterScopesByType(thirdTypesScope.Children, symboltable.TypeScope)
	require.GreaterOrEqual(t, len(thirdTypeScopes), 1, "Third TypesScope should have at least one TypeScope child")

	var amountTypeScope *symboltable.SymbolTable
	for _, ts := range thirdTypeScopes {
		if _, err := ts.Lookup("Amount"); err == nil {
			amountTypeScope = ts
			break
		}
	}
	require.NotNil(t, amountTypeScope, "Should find a TypeScope containing Amount")
	amountDef, err := amountTypeScope.Lookup("Amount")
	require.NoError(t, err, "Amount should be found in third TypeScope")
	assert.NotNil(t, amountDef.Name)
	assert.Equal(t, "Amount", *amountDef.Name, "Type name should be Amount")
}

// TestParseTypesSectionsIsolation verifies each TypesScope's TypeScope children are isolated;
// a type defined in section A is not visible in section B.
func TestParseTypesSectionsIsolation(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)

types
	Percentage = Number(precision=2, minimum=0, maximum=100)`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input with two types sections")
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have exactly two TypesScope children
	require.Len(t, doc.Children, 2, "Expected two children (two TypesScope)")

	firstTypesScope := doc.Children[0]
	secondTypesScope := doc.Children[1]

	// Verify both are TypesScope
	assert.Equal(t, symboltable.TypesScope, firstTypesScope.ScopeType, "First child should be TypesScope")
	assert.Equal(t, symboltable.TypesScope, secondTypesScope.ScopeType, "Second child should be TypesScope")

	// Verify each TypesScope contains its own TypeScope children
	firstTypeScopes := filterScopesByType(firstTypesScope.Children, symboltable.TypeScope)
	require.GreaterOrEqual(t, len(firstTypeScopes), 1, "First TypesScope should have at least one TypeScope child")

	secondTypeScopes := filterScopesByType(secondTypesScope.Children, symboltable.TypeScope)
	require.GreaterOrEqual(t, len(secondTypeScopes), 1, "Second TypesScope should have at least one TypeScope child")

	// Verify Currency is found in the first TypesScope's TypeScope children
	currencyFound := false
	for _, ts := range firstTypeScopes {
		_, err := ts.Lookup("Currency")
		if err == nil {
			currencyFound = true
			break
		}
	}
	assert.True(t, currencyFound, "Currency should be found in first TypesScope's TypeScope children")

	// Verify Percentage is found in the second TypesScope's TypeScope children
	percentageFound := false
	for _, ts := range secondTypeScopes {
		_, err := ts.Lookup("Percentage")
		if err == nil {
			percentageFound = true
			break
		}
	}
	assert.True(t, percentageFound, "Percentage should be found in second TypesScope's TypeScope children")

	// Key assertion: Currency should NOT be found in the second TypesScope's TypeScope children
	for _, ts := range secondTypeScopes {
		_, err := ts.Lookup("Currency")
		assert.Error(t, err, "Currency should NOT be found in second TypesScope's TypeScope children")
	}

	// Key assertion: Percentage should NOT be found in the first TypesScope's TypeScope children
	for _, ts := range firstTypeScopes {
		_, err := ts.Lookup("Percentage")
		assert.Error(t, err, "Percentage should NOT be found in first TypesScope's TypeScope children")
	}
}

// TestParseMixedTypesAcrossSections verifies that parsing a file with three separate
// types sections containing different base types (Number, String, Boolean) produces
// three TypesScope children, each with the correct type definition.
func TestParseMixedTypesAcrossSections(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)

types
	NamePattern = String(maximum=50)

types
	ActiveFlag = Boolean()`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input with mixed types across sections")
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have exactly three TypesScope children
	require.Len(t, doc.Children, 3, "Expected three children (three TypesScope)")

	firstTypesScope := doc.Children[0]
	secondTypesScope := doc.Children[1]
	thirdTypesScope := doc.Children[2]

	// Verify all are TypesScope
	assert.Equal(t, symboltable.TypesScope, firstTypesScope.ScopeType, "First child should be TypesScope")
	assert.Equal(t, symboltable.TypesScope, secondTypesScope.ScopeType, "Second child should be TypesScope")
	assert.Equal(t, symboltable.TypesScope, thirdTypesScope.ScopeType, "Third child should be TypesScope")

	// Verify first TypesScope has at least 1 TypeScope child containing Currency
	firstTypeScopes := filterScopesByType(firstTypesScope.Children, symboltable.TypeScope)
	require.GreaterOrEqual(t, len(firstTypeScopes), 1, "First TypesScope should have at least one TypeScope child")

	var currencyTypeScope *symboltable.SymbolTable
	for _, ts := range firstTypeScopes {
		if _, err := ts.Lookup("Currency"); err == nil {
			currencyTypeScope = ts
			break
		}
	}
	require.NotNil(t, currencyTypeScope, "Should find a TypeScope containing Currency")
	currencyDef, err := currencyTypeScope.Lookup("Currency")
	require.NoError(t, err, "Currency should be found in first TypeScope")
	assert.NotNil(t, currencyDef.Name)
	assert.Equal(t, "Currency", *currencyDef.Name, "Type name should be Currency")

	// Verify second TypesScope has at least 1 TypeScope child containing NamePattern
	secondTypeScopes := filterScopesByType(secondTypesScope.Children, symboltable.TypeScope)
	require.GreaterOrEqual(t, len(secondTypeScopes), 1, "Second TypesScope should have at least one TypeScope child")

	var namePatternTypeScope *symboltable.SymbolTable
	for _, ts := range secondTypeScopes {
		if _, err := ts.Lookup("NamePattern"); err == nil {
			namePatternTypeScope = ts
			break
		}
	}
	require.NotNil(t, namePatternTypeScope, "Should find a TypeScope containing NamePattern")
	namePatternDef, err := namePatternTypeScope.Lookup("NamePattern")
	require.NoError(t, err, "NamePattern should be found in second TypeScope")
	assert.NotNil(t, namePatternDef.Name)
	assert.Equal(t, "NamePattern", *namePatternDef.Name, "Type name should be NamePattern")

	// Verify third TypesScope has at least 1 TypeScope child containing ActiveFlag
	thirdTypeScopes := filterScopesByType(thirdTypesScope.Children, symboltable.TypeScope)
	require.GreaterOrEqual(t, len(thirdTypeScopes), 1, "Third TypesScope should have at least one TypeScope child")

	var activeFlagTypeScope *symboltable.SymbolTable
	for _, ts := range thirdTypeScopes {
		if _, err := ts.Lookup("ActiveFlag"); err == nil {
			activeFlagTypeScope = ts
			break
		}
	}
	require.NotNil(t, activeFlagTypeScope, "Should find a TypeScope containing ActiveFlag")
	activeFlagDef, err := activeFlagTypeScope.Lookup("ActiveFlag")
	require.NoError(t, err, "ActiveFlag should be found in third TypeScope")
	assert.NotNil(t, activeFlagDef.Name)
	assert.Equal(t, "ActiveFlag", *activeFlagDef.Name, "Type name should be ActiveFlag")
}
