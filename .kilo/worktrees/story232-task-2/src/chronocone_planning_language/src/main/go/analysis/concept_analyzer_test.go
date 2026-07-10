// Package analysis_test contains unit tests for the CPL analysis phase.
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/ast"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// ConceptAnalyzer Construction
// ============================================================================

// TestNewConceptAnalyzer_InitializesEmptyMaps verifies that NewConceptAnalyzer
// creates a ConceptAnalyzer with an empty concepts map and a nil lastConceptName.
func TestNewConceptAnalyzer_InitializesEmptyMaps(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()
	require.NotNil(t, ca, "NewConceptAnalyzer should not return nil")

	concepts := ca.GetConcepts()
	assert.Empty(t, concepts, "concepts map should be empty")
}

// ============================================================================
// AddConcept Tests
// ============================================================================

// TestConceptAnalyzer_AddConcept_AddsConcept verifies that AddConcept adds a
// concept to the concepts map.
func TestConceptAnalyzer_AddConcept_AddsConcept(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()
	concept := analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada"})

	err := ca.AddConcept(&concept)
	require.NoError(t, err, "AddConcept should not error for new concept")

	retrieved := ca.GetConcept("Country")
	assert.Equal(t, "Country", retrieved.ConceptName, "concept name should match")
	assert.Len(t, retrieved.ResolvedValues, 2, "should have 2 values")
}

// TestConceptAnalyzer_AddConcept_DuplicateReturnsError verifies that AddConcept
// returns an error when adding a duplicate concept.
func TestConceptAnalyzer_AddConcept_DuplicateReturnsError(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()
	concept1 := analysis.AnalysisConceptDefinition{
		ConceptName:    "Country",
		ResolvedValues: []analysis.AnalysisConceptValueDefinition{{Name: "USA", Position: 0}},
		SourceLocation: symboltable.SourceLocation{Line: 1, Column: 1},
	}
	concept2 := analysis.AnalysisConceptDefinition{
		ConceptName:    "Country",
		ResolvedValues: []analysis.AnalysisConceptValueDefinition{{Name: "Canada", Position: 0}},
		SourceLocation: symboltable.SourceLocation{Line: 2, Column: 1},
	}

	err := ca.AddConcept(&concept1)
	require.NoError(t, err)

	err = ca.AddConcept(&concept2)
	require.Error(t, err, "AddConcept should error for duplicate concept")
	assert.Contains(t, err.Error(), "duplicate concept definition", "error should indicate duplicate")
}

// TestConceptAnalyzer_AddConcept_SetsParentByIndent verifies that AddConcept
// sets the ParentConceptName when a concept has greater indentation than the previous.
func TestConceptAnalyzer_AddConcept_SetsParentByIndent(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()
	parent := analysis.AnalysisConceptDefinition{
		ConceptName:    "Country",
		IndentAsSpaces: 0,
		ResolvedValues: []analysis.AnalysisConceptValueDefinition{{Name: "USA", Position: 0}},
		SourceLocation: symboltable.SourceLocation{Line: 1, Column: 1},
	}
	child := analysis.AnalysisConceptDefinition{
		ConceptName:    "State",
		IndentAsSpaces: 4,
		ResolvedValues: []analysis.AnalysisConceptValueDefinition{{Name: "NY", Position: 0}},
		SourceLocation: symboltable.SourceLocation{Line: 2, Column: 1},
	}

	err := ca.AddConcept(&parent)
	require.NoError(t, err)

	err = ca.AddConcept(&child)
	require.NoError(t, err)

	retrieved := ca.GetConcept("State")
	require.NotNil(t, retrieved.ParentConceptName, "child should have parent set")
	assert.Equal(t, "Country", *retrieved.ParentConceptName, "parent should be Country")
}

// ============================================================================
// AddOrMergeConcept Tests
// ============================================================================

// TestConceptAnalyzer_AddOrMergeConcept_AddsNewConcept verifies that
// AddOrMergeConcept adds a new concept when no existing concept exists.
func TestConceptAnalyzer_AddOrMergeConcept_AddsNewConcept(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()
	concept := analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada", "Mexico"})

	err := ca.AddOrMergeConcept(&concept)
	require.NoError(t, err, "AddOrMergeConcept should not error for new concept")

	retrieved := ca.GetConcept("Country")
	assert.Equal(t, "Country", retrieved.ConceptName)
	assert.Len(t, retrieved.ResolvedValues, 3)
}

// TestConceptAnalyzer_AddOrMergeConcept_MergesExtension verifies that
// AddOrMergeConcept merges an extension into an existing concept.
func TestConceptAnalyzer_AddOrMergeConcept_MergesExtension(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()
	existing := analysis.NewValuesConceptDefinition("Priority", []string{"Low", "Medium", "High"})
	ca.AddConcept(&existing)

	extension := analysis.NewValuesConceptDefinition("Priority", []string{"Critical"})
	extension.InsertionPoint = ast.InsertionPoint{BeforeValue: "Low"}
	extension.Generators = []analysis.Generator{&analysis.ValuesGenerator{Values: []string{"Critical"}}}
	extension.SourceLocation = symboltable.SourceLocation{Line: 5, Column: 1}

	err := ca.AddOrMergeConcept(&extension)
	require.NoError(t, err, "AddOrMergeConcept should merge extension")

	merged := ca.GetConcept("Priority")
	require.Len(t, merged.ResolvedValues, 4, "should have 4 values after merge")
	assert.Equal(t, "Critical", merged.ResolvedValues[0].Name)
	assert.Equal(t, "Low", merged.ResolvedValues[1].Name)
	assert.Equal(t, "Medium", merged.ResolvedValues[2].Name)
	assert.Equal(t, "High", merged.ResolvedValues[3].Name)
}

// TestConceptAnalyzer_AddOrMergeConcept_DuplicateWithoutInsertionPoint verifies
// that AddOrMergeConcept returns an error for a true duplicate.
func TestConceptAnalyzer_AddOrMergeConcept_DuplicateWithoutInsertionPoint(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()
	initial := analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada"})
	initial.SourceLocation = symboltable.SourceLocation{Line: 1, Column: 1}
	err := ca.AddOrMergeConcept(&initial)
	require.NoError(t, err)

	duplicate := analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada", "Mexico"})
	duplicate.SourceLocation = symboltable.SourceLocation{Line: 3, Column: 1}
	err = ca.AddOrMergeConcept(&duplicate)
	require.Error(t, err, "should error on duplicate without insertion point")
	assert.Contains(t, err.Error(), "duplicate concept definition")
}

// TestConceptAnalyzer_AddOrMergeConcept_ExtensionWithNoBase verifies that
// AddOrMergeConcept returns an error when an extension has no base concept.
func TestConceptAnalyzer_AddOrMergeConcept_ExtensionWithNoBase(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()
	extension := analysis.NewValuesConceptDefinition("NonExistent", []string{"New"})
	extension.InsertionPoint = ast.InsertionPoint{BeforeValue: "X"}
	extension.SourceLocation = symboltable.SourceLocation{Line: 1, Column: 1}

	err := ca.AddOrMergeConcept(&extension)
	require.Error(t, err, "should error when extension has no base concept")
	assert.Contains(t, err.Error(), "no initial concept definition to extend")
}

// ============================================================================
// ResolveConceptRanges Tests
// ============================================================================

// TestConceptAnalyzer_ResolveConceptRanges_NoCRGs_ReturnsNil verifies that
// ResolveConceptRanges is a no-op when no ConceptRangeGenerators exist.
func TestConceptAnalyzer_ResolveConceptRanges_NoCRGs_ReturnsNil(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()
	source := analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada"})
	ca.AddConcept(&source)

	err := ca.ResolveConceptRanges()
	require.NoError(t, err, "ResolveConceptRanges should return nil when there are no CRGs")
}

// TestConceptAnalyzer_ResolveConceptRanges_SingleCRG_ResolvesCorrectly verifies
// that a single ConceptRangeGenerator referencing a resolved source concept works.
func TestConceptAnalyzer_ResolveConceptRanges_SingleCRG_ResolvesCorrectly(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()
	source := analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada", "Mexico", "UK", "France"})
	ca.AddConcept(&source)

	target := analysis.AnalysisConceptDefinition{
		ConceptName: "Subset",
		Generators: []analysis.Generator{
			&analysis.ConceptRangeGenerator{ConceptName: "Country", StartValue: "Canada", EndValue: "UK"},
		},
		SourceLocation: symboltable.SourceLocation{},
	}
	ca.AddConcept(&target)

	err := ca.ResolveConceptRanges()
	require.NoError(t, err, "ResolveConceptRanges should succeed")

	resolved := ca.GetConcept("Subset")
	require.Len(t, resolved.ResolvedValues, 3, "Subset should have 3 resolved values")
	assert.Equal(t, "Canada", resolved.ResolvedValues[0].Name)
	assert.Equal(t, "Mexico", resolved.ResolvedValues[1].Name)
	assert.Equal(t, "UK", resolved.ResolvedValues[2].Name)
}

// TestConceptAnalyzer_ResolveConceptRanges_SelfReference_ReturnsError verifies
// that a self-referencing ConceptRangeGenerator produces a circular dependency error.
func TestConceptAnalyzer_ResolveConceptRanges_SelfReference_ReturnsError(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()
	self := analysis.AnalysisConceptDefinition{
		ConceptName: "Self",
		Generators: []analysis.Generator{
			&analysis.ConceptRangeGenerator{ConceptName: "Self", StartValue: "A", EndValue: "B"},
		},
		SourceLocation: symboltable.SourceLocation{},
	}
	ca.AddConcept(&self)

	err := ca.ResolveConceptRanges()
	require.Error(t, err, "should return error for self-reference")
	assert.Contains(t, err.Error(), "circular dependency")
}

// TestConceptAnalyzer_ResolveConceptRanges_MissingSource_ReturnsError verifies
// that referencing a non-existent source concept produces an error.
func TestConceptAnalyzer_ResolveConceptRanges_MissingSource_ReturnsError(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()
	target := analysis.AnalysisConceptDefinition{
		ConceptName: "Target",
		Generators: []analysis.Generator{
			&analysis.ConceptRangeGenerator{ConceptName: "Missing", StartValue: "A", EndValue: "B"},
		},
		SourceLocation: symboltable.SourceLocation{},
	}
	ca.AddConcept(&target)

	err := ca.ResolveConceptRanges()
	require.Error(t, err, "should return error for missing source concept")
	assert.Contains(t, err.Error(), "unknown concept")
}

// TestConceptAnalyzer_ResolveConceptRanges_ChainOfReferences_ResolvesInOrder verifies
// that chained concept references resolve correctly in dependency order.
func TestConceptAnalyzer_ResolveConceptRanges_ChainOfReferences_ResolvesInOrder(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()

	c := analysis.NewValuesConceptDefinition("C", []string{"1", "2", "3"})
	ca.AddConcept(&c)

	b := analysis.AnalysisConceptDefinition{
		ConceptName: "B",
		Generators: []analysis.Generator{
			&analysis.ConceptRangeGenerator{ConceptName: "C", StartValue: "1", EndValue: "2"},
		},
		SourceLocation: symboltable.SourceLocation{},
	}
	ca.AddConcept(&b)

	a := analysis.AnalysisConceptDefinition{
		ConceptName: "A",
		Generators: []analysis.Generator{
			&analysis.ConceptRangeGenerator{ConceptName: "B", StartValue: "1", EndValue: "2"},
		},
		SourceLocation: symboltable.SourceLocation{},
	}
	ca.AddConcept(&a)

	err := ca.ResolveConceptRanges()
	require.NoError(t, err, "should succeed for chain of references")

	resolvedA := ca.GetConcept("A")
	require.Len(t, resolvedA.ResolvedValues, 2, "A should have 2 resolved values")
	assert.Equal(t, "1", resolvedA.ResolvedValues[0].Name)
	assert.Equal(t, "2", resolvedA.ResolvedValues[1].Name)
}

// ============================================================================
// GetConcept / GetConcepts Tests
// ============================================================================

// TestConceptAnalyzer_GetConcept_ExistingConcept verifies that GetConcept returns
// the correct concept for an existing name.
func TestConceptAnalyzer_GetConcept_ExistingConcept(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()
	concept := analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada"})
	ca.AddNewConcept(&concept)

	retrieved := ca.GetConcept("Country")
	assert.Equal(t, "Country", retrieved.ConceptName)
	assert.Len(t, retrieved.ResolvedValues, 2)
}

// TestConceptAnalyzer_GetConcept_MissingConcept verifies that GetConcept returns
// a zero-value concept for a missing name.
func TestConceptAnalyzer_GetConcept_MissingConcept(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()
	retrieved := ca.GetConcept("Missing")
	assert.Equal(t, "", retrieved.ConceptName, "missing concept should return zero value")
}

// TestConceptAnalyzer_GetConcepts_ReturnsMap verifies that GetConcepts
// returns the internal concepts map and modifications are visible through it.
func TestConceptAnalyzer_GetConcepts_ReturnsMap(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConceptAnalyzer()
	concept := analysis.NewValuesConceptDefinition("Country", []string{"USA"})
	ca.AddNewConcept(&concept)

	concepts := ca.GetConcepts()
	assert.Len(t, concepts, 1)
	assert.Contains(t, concepts, "Country")
}
