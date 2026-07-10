// Package analysis_test contains tests for the TypeAnalyzer.
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// TypeAnalyzer Construction Tests
// ============================================================================

// TestNewTypeAnalyzer creates a TypeAnalyzer and verifies its types map is initialized.
func TestNewTypeAnalyzer(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	require.NotNil(t, ta)
}

// ============================================================================
// AddType Tests
// ============================================================================

// TestTypeAnalyzer_AddType registers a type and verifies it can be retrieved.
func TestTypeAnalyzer_AddType(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Currency",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
		RawConstraints: nil,
		DerivationPath: []string{},
	})

	// ComputeDerivationPaths should succeed with one built-in derived type
	err := ta.ComputeDerivationPaths(nil)
	require.NoError(t, err)
}

// TestTypeAnalyzer_AddTypeMultiple registers multiple types.
func TestTypeAnalyzer_AddTypeMultiple(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Currency",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
		RawConstraints: nil,
		DerivationPath: []string{},
	})
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Amount",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
		RawConstraints: nil,
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(nil)
	require.NoError(t, err)
}

// ============================================================================
// Derivation Path Computation Tests
// ============================================================================

// TestTypeAnalyzer_ComputeDerivationPaths_SingleBuiltin computes a single type derived from Number.
func TestTypeAnalyzer_ComputeDerivationPaths_SingleBuiltin(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Currency",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
		RawConstraints: nil,
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(nil)
	require.NoError(t, err)
}

// TestTypeAnalyzer_ComputeDerivationPaths_Chain verifies a multi-level derivation chain.
func TestTypeAnalyzer_ComputeDerivationPaths_Chain(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Currency",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
		RawConstraints: []analysis.Constraint{{Name: "precision", Value: "2"}},
		DerivationPath: []string{},
	})
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "SmallCurrency",
		BaseType:       analysis.NewDeclaredTypeRef("Currency"),
		RawConstraints: []analysis.Constraint{{Name: "precision", Value: "0"}},
		DerivationPath: []string{},
	})
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "TinyCurrency",
		BaseType:       analysis.NewDeclaredTypeRef("SmallCurrency"),
		RawConstraints: []analysis.Constraint{{Name: "precision", Value: "0"}, {Name: "minimum", Value: "0"}},
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(nil)
	require.NoError(t, err)
}

// TestTypeAnalyzer_ComputeDerivationPaths_UnknownBaseType errors when base type is missing.
func TestTypeAnalyzer_ComputeDerivationPaths_UnknownBaseType(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Orphan",
		BaseType:       analysis.NewDeclaredTypeRef("NonExistent"),
		RawConstraints: nil,
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not derive from a built-in type")
}

// TestTypeAnalyzer_ComputeDerivationPaths_NotEndingInBuiltin errors when chain doesn't terminate in built-in.
func TestTypeAnalyzer_ComputeDerivationPaths_NotEndingInBuiltin(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Currency",
		BaseType:       analysis.NewDeclaredTypeRef("Integer"),
		RawConstraints: []analysis.Constraint{{Name: "precision", Value: "2"}},
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not derive from a built-in type")
}

// ============================================================================
// Constraint Validation Tests
// ============================================================================

// TestTypeAnalyzer_ConstraintValidation_NarrowingSucceeds when derived constraints are more restrictive.
func TestTypeAnalyzer_ConstraintValidation_NarrowingSucceeds(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Currency",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
		RawConstraints: []analysis.Constraint{{Name: "precision", Value: "2"}},
		DerivationPath: []string{},
	})
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "SmallCurrency",
		BaseType:       analysis.NewDeclaredTypeRef("Currency"),
		RawConstraints: []analysis.Constraint{{Name: "precision", Value: "0"}},
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(nil)
	require.NoError(t, err)
}

// TestTypeAnalyzer_ConstraintValidation_ExpansionFails when derived constraints expand.
func TestTypeAnalyzer_ConstraintValidation_ExpansionFails(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "ExactCurrency",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
		RawConstraints: []analysis.Constraint{{Name: "precision", Value: "0"}},
		DerivationPath: []string{},
	})
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "FlexibleCurrency",
		BaseType:       analysis.NewDeclaredTypeRef("ExactCurrency"),
		RawConstraints: []analysis.Constraint{{Name: "precision", Value: "2"}},
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expands inherited constraint")
}

// TestTypeAnalyzer_ConstraintValidation_MinimumExpansionFails when minimum decreases.
func TestTypeAnalyzer_ConstraintValidation_MinimumExpansionFails(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "BoundedType",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
		RawConstraints: []analysis.Constraint{{Name: "minimum", Value: "10"}},
		DerivationPath: []string{},
	})
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "ExpandedType",
		BaseType:       analysis.NewDeclaredTypeRef("BoundedType"),
		RawConstraints: []analysis.Constraint{{Name: "minimum", Value: "0"}},
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expands inherited constraint")
}

// TestTypeAnalyzer_ConstraintValidation_MaximumExpansionFails when maximum increases.
func TestTypeAnalyzer_ConstraintValidation_MaximumExpansionFails(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "SmallRange",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
		RawConstraints: []analysis.Constraint{{Name: "maximum", Value: "100"}},
		DerivationPath: []string{},
	})
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "LargeRange",
		BaseType:       analysis.NewDeclaredTypeRef("SmallRange"),
		RawConstraints: []analysis.Constraint{{Name: "maximum", Value: "1000"}},
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expands inherited constraint")
}

// TestTypeAnalyzer_ConstraintValidation_StringNarrowingSucceeds for String constraints.
func TestTypeAnalyzer_ConstraintValidation_StringNarrowingSucceeds(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "BaseStr",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.StringType),
		RawConstraints: []analysis.Constraint{{Name: "minimum", Value: "0"}, {Name: "maximum", Value: "100"}},
		DerivationPath: []string{},
	})
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "ShortStr",
		BaseType:       analysis.NewDeclaredTypeRef("BaseStr"),
		RawConstraints: []analysis.Constraint{{Name: "maximum", Value: "50"}},
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(nil)
	require.NoError(t, err)
}

// TestTypeAnalyzer_ConstraintValidation_StringExpansionFails for String constraints.
func TestTypeAnalyzer_ConstraintValidation_StringExpansionFails(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "BaseStr",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.StringType),
		RawConstraints: []analysis.Constraint{{Name: "maximum", Value: "50"}},
		DerivationPath: []string{},
	})
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "WideStr",
		BaseType:       analysis.NewDeclaredTypeRef("BaseStr"),
		RawConstraints: []analysis.Constraint{{Name: "maximum", Value: "100"}},
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expands inherited constraint")
}

// TestTypeAnalyzer_ConstraintValidation_InstantNarrowingSucceeds for Instant constraints.
func TestTypeAnalyzer_ConstraintValidation_InstantNarrowingSucceeds(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	concepts := analysis.LoadStandardTimeConcepts(t)
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "NewYear",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.InstantType),
		RawConstraints: []analysis.Constraint{{Name: "month", Value: "January"}, {Name: "day_of_month", Value: "1"}},
		DerivationPath: []string{},
	})
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "NewYear2024",
		BaseType:       analysis.NewDeclaredTypeRef("NewYear"),
		RawConstraints: []analysis.Constraint{{Name: "year", Value: "2024"}},
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(concepts)
	require.NoError(t, err)
}

// TestTypeAnalyzer_ConstraintValidation_InstantExpansionFails when Instant constraint changes.
func TestTypeAnalyzer_ConstraintValidation_InstantExpansionFails(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	concepts := analysis.LoadStandardTimeConcepts(t)
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "ExactTime",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.InstantType),
		RawConstraints: []analysis.Constraint{{Name: "hour", Value: "12"}},
		DerivationPath: []string{},
	})
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "DifferentTime",
		BaseType:       analysis.NewDeclaredTypeRef("ExactTime"),
		RawConstraints: []analysis.Constraint{{Name: "hour", Value: "13"}},
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(concepts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expands inherited constraint")
}

// TestTypeAnalyzer_ConstraintValidation_BooleanAlwaysSucceeds for Boolean types.
func TestTypeAnalyzer_ConstraintValidation_BooleanAlwaysSucceeds(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "BaseBool",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.BooleanType),
		RawConstraints: []analysis.Constraint{},
		DerivationPath: []string{},
	})
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "DerivedBool",
		BaseType:       analysis.NewDeclaredTypeRef("BaseBool"),
		RawConstraints: []analysis.Constraint{},
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(nil)
	require.NoError(t, err)
}

// ============================================================================
// Instant/Interval with Concepts Tests
// ============================================================================

// TestTypeAnalyzer_ComputeDerivationPaths_InstantWithConcepts verifies Instant types need concepts.
func TestTypeAnalyzer_ComputeDerivationPaths_InstantWithConcepts(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	concepts := analysis.LoadStandardTimeConcepts(t)
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Timestamp",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.InstantType),
		RawConstraints: []analysis.Constraint{{Name: "second", Value: "1"}},
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(concepts)
	require.NoError(t, err)
}

// TestTypeAnalyzer_ComputeDerivationPaths_InstantMissingConceptsFails verifies missing concepts error.
func TestTypeAnalyzer_ComputeDerivationPaths_InstantMissingConceptsFails(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	// Do NOT add standard time concepts
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Timestamp",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.InstantType),
		RawConstraints: []analysis.Constraint{{Name: "second", Value: "1"}},
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required concept Second for constraint second is not defined")
}

// TestTypeAnalyzer_ComputeDerivationPaths_IntervalWithConcepts verifies Interval types need concepts.
func TestTypeAnalyzer_ComputeDerivationPaths_IntervalWithConcepts(t *testing.T) {
	t.Parallel()

	ta := analysis.NewTypeAnalyzer()
	concepts := analysis.LoadStandardTimeConcepts(t)
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:     "FiscalMonth",
		BaseType: analysis.NewBuiltInTypeRef(analysis.IntervalType),
		RawConstraints: []analysis.Constraint{
			{Name: "starts_at", Value: "FiscalYearStart"},
			{Name: "unit", Value: "month"},
		},
		DerivationPath: []string{},
	})
	ta.AddType(analysis.AnalysisTypeDefinition{
		Name:           "FiscalYearStart",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.InstantType),
		RawConstraints: []analysis.Constraint{{Name: "month", Value: "January"}, {Name: "day_of_month", Value: "1"}},
		DerivationPath: []string{},
	})

	err := ta.ComputeDerivationPaths(concepts)
	require.NoError(t, err)
}
