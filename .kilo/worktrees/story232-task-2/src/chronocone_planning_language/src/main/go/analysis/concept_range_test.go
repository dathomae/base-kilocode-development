// Package analysis_test contains unit tests for the CPL analysis phase.
//
// These tests follow the Behavioral Test-Driven Development (BTDD) approach,
// where tests are written before the implementation exists. This is Phase B
// of the BTDD lifecycle - defining test signatures and expected behaviors.
//
// The tests in this file cover the ConceptRangeGenerator struct and its
// behavior as a Generator implementation.
//
// Reference: Story129 CPL Concepts Analyzer - Concept Range
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/ast"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// ConceptRangeGenerator Tests
// ============================================================================

// TestConceptRangeGenerator_StructFields verifies that the ConceptRangeGenerator
// struct has the expected fields: ConceptName, StartValue, and EndValue, each
// with the correct string type.
//
// This test ensures the structure matches the design specification and that
// all fields are accessible and correctly typed.
func TestConceptRangeGenerator_StructFields(t *testing.T) {
	t.Parallel()

	gen := analysis.ConceptRangeGenerator{
		ConceptName: "Country",
		StartValue:  "Canada",
		EndValue:    "UK",
	}

	assert.Equal(t, "Country", gen.ConceptName, "ConceptName field should match the set value")
	assert.Equal(t, "Canada", gen.StartValue, "StartValue field should match the set value")
	assert.Equal(t, "UK", gen.EndValue, "EndValue field should match the set value")
}

// TestConceptRangeGenerator_ImplementsGeneratorInterface verifies that
// ConceptRangeGenerator satisfies the Generator interface.
//
// This test ensures that a *ConceptRangeGenerator can be assigned to a variable
// of type Generator, confirming the GenerateValues method has the correct
// signature: func(concepts *map[string]AnalysisConceptDefinition) ([]string, error).
func TestConceptRangeGenerator_ImplementsGeneratorInterface(t *testing.T) {
	t.Parallel()

	// Compile-time check: if *ConceptRangeGenerator does not implement Generator,
	// this line will fail to compile.
	var _ analysis.Generator = (*analysis.ConceptRangeGenerator)(nil)
}

// TestConceptRangeGenerator_GenerateValues_NilMapError verifies that
// GenerateValues returns an error (rather than panicking) when passed
// a nil or empty concepts map — the referenced concept will simply
// not be found.
func TestConceptRangeGenerator_GenerateValues_NilMapError(t *testing.T) {
	t.Parallel()

	gen := analysis.ConceptRangeGenerator{
		ConceptName: "Country",
		StartValue:  "Canada",
		EndValue:    "UK",
	}

	values, err := gen.GenerateValues(nil)

	require.Error(t, err, "GenerateValues should return an error for a nil concepts map")
	assert.Nil(t, values, "GenerateValues should return nil values when an error occurs")
	assert.Contains(t, err.Error(), "referenced concept not found", "error should indicate the referenced concept was not found")
}

// TestConceptRangeGenerator_GenerateValues_BasicRange verifies that
// GenerateValues extracts a sub-range from a Values-based source concept.
//
// Given a Country concept with values [USA, Canada, Mexico, UK, France],
// a range from "Canada" to "UK" should return ["Canada", "Mexico", "UK"].
func TestConceptRangeGenerator_GenerateValues_BasicRange(t *testing.T) {
	t.Parallel()

	concepts := map[string]analysis.AnalysisConceptDefinition{
		"Country": analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada", "Mexico", "UK", "France"}),
	}

	gen := analysis.ConceptRangeGenerator{
		ConceptName: "Country",
		StartValue:  "Canada",
		EndValue:    "UK",
	}

	values, err := gen.GenerateValues(concepts)

	require.NoError(t, err, "GenerateValues should succeed for a valid range")
	assert.Equal(t, []string{"Canada", "Mexico", "UK"}, values, "values should be the sub-range from Canada to UK")
}

// TestConceptRangeGenerator_GenerateValues_SingleValue verifies that
// GenerateValues returns a single-element slice when start and end are the same value.
func TestConceptRangeGenerator_GenerateValues_SingleValue(t *testing.T) {
	t.Parallel()

	concepts := map[string]analysis.AnalysisConceptDefinition{
		"Country": analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada", "Mexico", "UK", "France"}),
	}

	gen := analysis.ConceptRangeGenerator{
		ConceptName: "Country",
		StartValue:  "Canada",
		EndValue:    "Canada",
	}

	values, err := gen.GenerateValues(concepts)

	require.NoError(t, err, "GenerateValues should succeed when start equals end")
	assert.Equal(t, []string{"Canada"}, values, "values should contain a single element")
}

// TestConceptRangeGenerator_GenerateValues_MissingConcept verifies that
// GenerateValues returns an error when the referenced concept does not exist in the map.
func TestConceptRangeGenerator_GenerateValues_MissingConcept(t *testing.T) {
	t.Parallel()

	concepts := map[string]analysis.AnalysisConceptDefinition{}

	gen := analysis.ConceptRangeGenerator{
		ConceptName: "MissingConcept",
		StartValue:  "A",
		EndValue:    "B",
	}

	values, err := gen.GenerateValues(concepts)

	require.Error(t, err, "GenerateValues should return an error for a missing concept")
	assert.Contains(t, err.Error(), "referenced concept not found", "error should indicate the referenced concept was not found")
	assert.Nil(t, values, "values should be nil when an error occurs")
}

// TestConceptRangeGenerator_GenerateValues_MissingStartValue verifies that
// GenerateValues returns an error when the start value is not present in the source concept.
func TestConceptRangeGenerator_GenerateValues_MissingStartValue(t *testing.T) {
	t.Parallel()

	concepts := map[string]analysis.AnalysisConceptDefinition{
		"Country": analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada", "Mexico", "UK", "France"}),
	}

	gen := analysis.ConceptRangeGenerator{
		ConceptName: "Country",
		StartValue:  "Germany",
		EndValue:    "UK",
	}

	values, err := gen.GenerateValues(concepts)

	require.Error(t, err, "GenerateValues should return an error for a missing start value")
	assert.Contains(t, err.Error(), "start value not found", "error should indicate the start value was not found")
	assert.Nil(t, values, "values should be nil when an error occurs")
}

// TestConceptRangeGenerator_GenerateValues_MissingEndValue verifies that
// GenerateValues returns an error when the end value is not present in the source concept.
func TestConceptRangeGenerator_GenerateValues_MissingEndValue(t *testing.T) {
	t.Parallel()

	concepts := map[string]analysis.AnalysisConceptDefinition{
		"Country": analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada", "Mexico", "UK", "France"}),
	}

	gen := analysis.ConceptRangeGenerator{
		ConceptName: "Country",
		StartValue:  "Canada",
		EndValue:    "Germany",
	}

	values, err := gen.GenerateValues(concepts)

	require.Error(t, err, "GenerateValues should return an error for a missing end value")
	assert.Contains(t, err.Error(), "end value not found", "error should indicate the end value was not found")
	assert.Nil(t, values, "values should be nil when an error occurs")
}

// TestConceptRangeGenerator_GenerateValues_StartAfterEnd verifies that
// GenerateValues returns an error when the start value comes after the end value
// in the source concept's order.
func TestConceptRangeGenerator_GenerateValues_StartAfterEnd(t *testing.T) {
	t.Parallel()

	concepts := map[string]analysis.AnalysisConceptDefinition{
		"Country": analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada", "Mexico", "UK", "France"}),
	}

	gen := analysis.ConceptRangeGenerator{
		ConceptName: "Country",
		StartValue:  "UK",
		EndValue:    "Canada",
	}

	values, err := gen.GenerateValues(concepts)

	require.Error(t, err, "GenerateValues should return an error when start comes after end")
	assert.Contains(t, err.Error(), "comes after", "error should indicate the start value comes after the end value")
	assert.Nil(t, values, "values should be nil when an error occurs")
}

// TestConceptRangeGenerator_GenerateValues_FromIntRangeConcept verifies that
// GenerateValues works correctly when the source concept was built with an IntRange generator.
//
// Given a Month concept with Int[1, 12, 1], a range from "3" to "6" should return
// ["3", "4", "5", "6"].
func TestConceptRangeGenerator_GenerateValues_FromIntRangeConcept(t *testing.T) {
	t.Parallel()

	monthConcept, err := analysis.NewIntRangeConceptDefinition("Month", 1, 12, 1)
	require.NoError(t, err, "setup: NewIntRangeConceptDefinition should succeed")

	concepts := map[string]analysis.AnalysisConceptDefinition{
		"Month": monthConcept,
	}

	gen := analysis.ConceptRangeGenerator{
		ConceptName: "Month",
		StartValue:  "3",
		EndValue:    "6",
	}

	values, err := gen.GenerateValues(concepts)

	require.NoError(t, err, "GenerateValues should succeed for an IntRange-based concept")
	assert.Equal(t, []string{"3", "4", "5", "6"}, values, "values should be the sub-range from 3 to 6")
}

// TestConceptRangeGenerator_GenerateValues_EntireConcept verifies that
// GenerateValues returns the full concept value list when start is the first value
// and end is the last value.
func TestConceptRangeGenerator_GenerateValues_EntireConcept(t *testing.T) {
	t.Parallel()

	concepts := map[string]analysis.AnalysisConceptDefinition{
		"Country": analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada", "Mexico", "UK", "France"}),
	}

	gen := analysis.ConceptRangeGenerator{
		ConceptName: "Country",
		StartValue:  "USA",
		EndValue:    "France",
	}

	values, err := gen.GenerateValues(concepts)

	require.NoError(t, err, "GenerateValues should succeed for the entire concept range")
	assert.Equal(t, []string{"USA", "Canada", "Mexico", "UK", "France"}, values, "values should contain the full concept list")
}

// TestConceptRangeGenerator_GenerateValues_PreservesOrder verifies that
// GenerateValues returns values in the source concept's declaration order,
// not in alphabetical or any other order.
func TestConceptRangeGenerator_GenerateValues_PreservesOrder(t *testing.T) {
	t.Parallel()

	concepts := map[string]analysis.AnalysisConceptDefinition{
		"Priority": analysis.NewValuesConceptDefinition("Priority", []string{"Low", "Medium", "High", "Critical"}),
	}

	gen := analysis.ConceptRangeGenerator{
		ConceptName: "Priority",
		StartValue:  "Medium",
		EndValue:    "Critical",
	}

	values, err := gen.GenerateValues(concepts)

	require.NoError(t, err, "GenerateValues should succeed and preserve source order")
	assert.Equal(t, []string{"Medium", "High", "Critical"}, values, "values should preserve the source concept's order")
}

// ============================================================================
// CreateGenerator Factory Tests — ConceptRangeGeneratorDef
// ============================================================================

// TestCreateGenerator_ConceptRangeGeneratorDef_ReturnsCorrectType verifies that
// CreateGenerator returns a *ConceptRangeGenerator when given an ast.ConceptRangeGeneratorDef.
func TestCreateGenerator_ConceptRangeGeneratorDef_ReturnsCorrectType(t *testing.T) {
	t.Parallel()

	parserGen := ast.ConceptRangeGeneratorDef{
		ConceptName: "Country",
		StartValue:  "Canada",
		EndValue:    "UK",
	}

	gen, err := analysis.CreateGenerator(parserGen)

	require.NoError(t, err, "CreateGenerator should not return an error for a valid ConceptRangeGeneratorDef")
	require.NotNil(t, gen, "CreateGenerator should return a non-nil generator")

	// Type assertion: must be *ConceptRangeGenerator
	conceptRangeGen, ok := gen.(*analysis.ConceptRangeGenerator)
	require.True(t, ok, "CreateGenerator should return a *ConceptRangeGenerator, got %T", gen)

	// Verify fields are correctly mapped
	assert.Equal(t, "Country", conceptRangeGen.ConceptName, "ConceptName should match the parser definition")
	assert.Equal(t, "Canada", conceptRangeGen.StartValue, "StartValue should match the parser definition")
	assert.Equal(t, "UK", conceptRangeGen.EndValue, "EndValue should match the parser definition")
}

// TestCreateGenerator_ConceptRangeGeneratorDef_SatisfiesGeneratorInterface verifies that
// the Generator returned by CreateGenerator for a ConceptRangeGeneratorDef can be used
// as a Generator interface value.
func TestCreateGenerator_ConceptRangeGeneratorDef_SatisfiesGeneratorInterface(t *testing.T) {
	t.Parallel()

	parserGen := ast.ConceptRangeGeneratorDef{
		ConceptName: "Priority",
		StartValue:  "Medium",
		EndValue:    "Critical",
	}

	gen, err := analysis.CreateGenerator(parserGen)

	require.NoError(t, err, "CreateGenerator should not return an error")
	require.NotNil(t, gen, "CreateGenerator should return a non-nil generator")

	// Compile-time and runtime check: the returned value satisfies the Generator interface
	var _ analysis.Generator = gen
}

// TestCreateGenerator_ConceptRangeGeneratorDef_GenerateValuesWorks verifies that
// the ConceptRangeGenerator created by the factory can successfully generate values
// when provided with a valid concepts map.
func TestCreateGenerator_ConceptRangeGeneratorDef_GenerateValuesWorks(t *testing.T) {
	t.Parallel()

	parserGen := ast.ConceptRangeGeneratorDef{
		ConceptName: "Country",
		StartValue:  "Canada",
		EndValue:    "UK",
	}

	gen, err := analysis.CreateGenerator(parserGen)
	require.NoError(t, err, "CreateGenerator should not return an error")

	concepts := map[string]analysis.AnalysisConceptDefinition{
		"Country": analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada", "Mexico", "UK", "France"}),
	}

	values, err := gen.GenerateValues(concepts)

	require.NoError(t, err, "GenerateValues should succeed for a factory-created ConceptRangeGenerator")
	assert.Equal(t, []string{"Canada", "Mexico", "UK"}, values, "values should be the sub-range from Canada to UK")
}

// TestCreateGenerator_ConceptRangeGeneratorDef_EmptyStringsAllowed verifies that
// CreateGenerator accepts empty strings for ConceptName, StartValue, and EndValue
// without error (validation is deferred to GenerateValues).
func TestCreateGenerator_ConceptRangeGeneratorDef_EmptyStringsAllowed(t *testing.T) {
	t.Parallel()

	parserGen := ast.ConceptRangeGeneratorDef{
		ConceptName: "",
		StartValue:  "",
		EndValue:    "",
	}

	gen, err := analysis.CreateGenerator(parserGen)

	require.NoError(t, err, "CreateGenerator should not error on empty strings; validation is deferred")
	require.NotNil(t, gen, "CreateGenerator should return a non-nil generator even with empty strings")

	conceptRangeGen := gen.(*analysis.ConceptRangeGenerator)
	assert.Equal(t, "", conceptRangeGen.ConceptName)
	assert.Equal(t, "", conceptRangeGen.StartValue)
	assert.Equal(t, "", conceptRangeGen.EndValue)
}

// TestCreateGenerator_UnknownType_ReturnsError verifies that CreateGenerator returns
// an error for an unsupported generator type, preserving the existing behavior.
func TestCreateGenerator_UnknownType_ReturnsError(t *testing.T) {
	t.Parallel()

	gen, err := analysis.CreateGenerator("not-a-generator")

	require.Error(t, err, "CreateGenerator should return an error for an unknown type")
	assert.Nil(t, gen, "generator should be nil when an error occurs")
	assert.Contains(t, err.Error(), "unknown generator type", "error message should indicate unknown type")
}

// ============================================================================
// ResolveConceptRanges Tests
// ============================================================================

// TestResolveConceptRanges_NoConceptRanges_ReturnsNil verifies that
// ResolveConceptRanges is a no-op when no concepts have ConceptRangeGenerators.
func TestResolveConceptRanges_NoConceptRanges_ReturnsNil(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	source := analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada"})
	analyzer.AddConcept(&source)

	err := analyzer.ResolveConceptRanges()
	require.NoError(t, err, "ResolveConceptRanges should return nil when there are no CRGs")
}

// TestResolveConceptRanges_SingleConceptRange_ResolvesCorrectly verifies that
// a concept with a single CRG referencing a Values-based source concept resolves
// to the expected sub-range.
func TestResolveConceptRanges_SingleConceptRange_ResolvesCorrectly(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	source := analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada", "Mexico", "UK", "France"})
	analyzer.AddConcept(&source)

	target := analysis.AnalysisConceptDefinition{
		ConceptName: "Subset",
		Generators: []analysis.Generator{
			&analysis.ConceptRangeGenerator{ConceptName: "Country", StartValue: "Canada", EndValue: "UK"},
		},
		SourceLocation: symboltable.SourceLocation{},
	}
	analyzer.AddConcept(&target)

	err := analyzer.ResolveConceptRanges()
	require.NoError(t, err, "ResolveConceptRanges should succeed for a valid single CRG")

	defs := analyzer.GetAnalysisDefinitions()
	resolvedTarget, exists := defs.Concepts["Subset"]
	require.True(t, exists, "Subset concept should exist")
	require.Len(t, resolvedTarget.ResolvedValues, 3, "Subset should have 3 resolved values")
	assert.Equal(t, "Canada", resolvedTarget.ResolvedValues[0].Name)
	assert.Equal(t, "Mexico", resolvedTarget.ResolvedValues[1].Name)
	assert.Equal(t, "UK", resolvedTarget.ResolvedValues[2].Name)
	assert.True(t, resolvedTarget.ContainsValue("Canada"))
	assert.True(t, resolvedTarget.ContainsValue("Mexico"))
	assert.True(t, resolvedTarget.ContainsValue("UK"))
}

// TestResolveConceptRanges_ChainOfReferences_ResolvesInOrder verifies that
// a chain of concept references (A refs B, B refs C) resolves correctly when
// C is fully resolved and the algorithm processes in dependency order.
func TestResolveConceptRanges_ChainOfReferences_ResolvesInOrder(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()

	c := analysis.NewValuesConceptDefinition("C", []string{"1", "2", "3"})
	analyzer.AddConcept(&c)

	b := analysis.AnalysisConceptDefinition{
		ConceptName: "B",
		Generators: []analysis.Generator{
			&analysis.ConceptRangeGenerator{ConceptName: "C", StartValue: "1", EndValue: "2"},
		},
		SourceLocation: symboltable.SourceLocation{},
	}
	analyzer.AddConcept(&b)

	a := analysis.AnalysisConceptDefinition{
		ConceptName: "A",
		Generators: []analysis.Generator{
			&analysis.ConceptRangeGenerator{ConceptName: "B", StartValue: "1", EndValue: "2"},
		},
		SourceLocation: symboltable.SourceLocation{},
	}
	analyzer.AddConcept(&a)

	err := analyzer.ResolveConceptRanges()
	require.NoError(t, err, "ResolveConceptRanges should succeed for a chain of references")

	defs := analyzer.GetAnalysisDefinitions()
	resolvedA := defs.Concepts["A"]
	require.Len(t, resolvedA.ResolvedValues, 2, "A should have 2 resolved values")
	assert.Equal(t, "1", resolvedA.ResolvedValues[0].Name)
	assert.Equal(t, "2", resolvedA.ResolvedValues[1].Name)
}

// TestResolveConceptRanges_SelfReference_ReturnsCircularDependencyError verifies
// that a concept referencing itself via a CRG produces a circular dependency error.
func TestResolveConceptRanges_SelfReference_ReturnsCircularDependencyError(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()

	self := analysis.AnalysisConceptDefinition{
		ConceptName: "Self",
		Generators: []analysis.Generator{
			&analysis.ConceptRangeGenerator{ConceptName: "Self", StartValue: "A", EndValue: "B"},
		},
		SourceLocation: symboltable.SourceLocation{},
	}
	analyzer.AddConcept(&self)

	err := analyzer.ResolveConceptRanges()
	require.Error(t, err, "ResolveConceptRanges should return an error for self-reference")
	assert.Contains(t, err.Error(), "circular dependency", "error should indicate circular dependency")
}

// TestResolveConceptRanges_MutualReference_ReturnsCircularDependencyError verifies
// that two concepts referencing each other (A→B, B→A) produces a circular dependency error.
func TestResolveConceptRanges_MutualReference_ReturnsCircularDependencyError(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()

	a := analysis.AnalysisConceptDefinition{
		ConceptName: "A",
		Generators: []analysis.Generator{
			&analysis.ConceptRangeGenerator{ConceptName: "B", StartValue: "X", EndValue: "Y"},
		},
		SourceLocation: symboltable.SourceLocation{},
	}
	analyzer.AddConcept(&a)

	b := analysis.AnalysisConceptDefinition{
		ConceptName: "B",
		Generators: []analysis.Generator{
			&analysis.ConceptRangeGenerator{ConceptName: "A", StartValue: "X", EndValue: "Y"},
		},
		SourceLocation: symboltable.SourceLocation{},
	}
	analyzer.AddConcept(&b)

	err := analyzer.ResolveConceptRanges()
	require.Error(t, err, "ResolveConceptRanges should return an error for mutual reference")
	assert.Contains(t, err.Error(), "circular dependency", "error should indicate circular dependency")
}

// TestResolveConceptRanges_SourceConceptNotFound_ReturnsError verifies that
// referencing a non-existent source concept produces a descriptive error.
func TestResolveConceptRanges_SourceConceptNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()

	target := analysis.AnalysisConceptDefinition{
		ConceptName: "Target",
		Generators: []analysis.Generator{
			&analysis.ConceptRangeGenerator{ConceptName: "Missing", StartValue: "A", EndValue: "B"},
		},
		SourceLocation: symboltable.SourceLocation{},
	}
	analyzer.AddConcept(&target)

	err := analyzer.ResolveConceptRanges()
	require.Error(t, err, "ResolveConceptRanges should return an error for missing source concept")
	assert.Contains(t, err.Error(), "unknown concept", "error should indicate unknown source concept")
}

// TestResolveConceptRanges_DuplicateValueFromRange_ReturnsError verifies that
// if a concept already has a resolved value and a CRG generates the same value,
// a duplicate value error is returned.
func TestResolveConceptRanges_DuplicateValueFromRange_ReturnsError(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()

	source := analysis.NewValuesConceptDefinition("Source", []string{"A", "B", "C"})
	analyzer.AddConcept(&source)

	target := analysis.AnalysisConceptDefinition{
		ConceptName: "Target",
		Generators: []analysis.Generator{
			&analysis.ValuesGenerator{Values: []string{"A"}},
			&analysis.ConceptRangeGenerator{ConceptName: "Source", StartValue: "A", EndValue: "C"},
		},
		ResolvedValues: []analysis.AnalysisConceptValueDefinition{
			{Name: "A", Position: 0},
		},
		SourceLocation: symboltable.SourceLocation{},
	}
	target.BuildIndex()
	analyzer.AddConcept(&target)

	err := analyzer.ResolveConceptRanges()
	require.Error(t, err, "ResolveConceptRanges should return an error for duplicate values")
	assert.Contains(t, err.Error(), "duplicate value", "error should indicate duplicate value")
}

// TestResolveConceptRanges_MultipleCRGsInOneConcept_AllResolved verifies that
// a concept with multiple ConceptRangeGenerators referencing different source
// concepts resolves all of them correctly.
func TestResolveConceptRanges_MultipleCRGsInOneConcept_AllResolved(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()

	source1 := analysis.NewValuesConceptDefinition("Source1", []string{"A", "B", "C"})
	source2 := analysis.NewValuesConceptDefinition("Source2", []string{"X", "Y", "Z"})
	analyzer.AddConcept(&source1)
	analyzer.AddConcept(&source2)

	target := analysis.AnalysisConceptDefinition{
		ConceptName: "Target",
		Generators: []analysis.Generator{
			&analysis.ConceptRangeGenerator{ConceptName: "Source1", StartValue: "A", EndValue: "B"},
			&analysis.ConceptRangeGenerator{ConceptName: "Source2", StartValue: "Y", EndValue: "Z"},
		},
		SourceLocation: symboltable.SourceLocation{},
	}
	analyzer.AddConcept(&target)

	err := analyzer.ResolveConceptRanges()
	require.NoError(t, err, "ResolveConceptRanges should succeed with multiple CRGs in one concept")

	defs := analyzer.GetAnalysisDefinitions()
	resolvedTarget := defs.Concepts["Target"]
	require.Len(t, resolvedTarget.ResolvedValues, 4, "Target should have 4 resolved values")
	assert.Equal(t, "A", resolvedTarget.ResolvedValues[0].Name)
	assert.Equal(t, "B", resolvedTarget.ResolvedValues[1].Name)
	assert.Equal(t, "Y", resolvedTarget.ResolvedValues[2].Name)
	assert.Equal(t, "Z", resolvedTarget.ResolvedValues[3].Name)
}

// ============================================================================
// AnalyzeConceptRangeGeneratorScope Handler Tests (Task 6)
// ============================================================================

// TestAnalyzeConceptRangeGeneratorScope_NilData_ReturnsError verifies that
// the handler returns an error when table.Data is nil (parsing validation failed).
func TestAnalyzeConceptRangeGeneratorScope_NilData_ReturnsError(t *testing.T) {
	t.Parallel()

	table := &symboltable.SymbolTable{
		ScopeType: symboltable.ConceptGeneratorScope,
		Data:      nil,
	}
	context := analysis.NewAnalysisContext()
	analyzer := analysis.NewCPLAnalyzer()

	err := analysis.AnalyzeConceptRangeGeneratorScope(table, context, analyzer)

	require.Error(t, err, "should return error when Data is nil")
	assert.Contains(t, err.Error(), "no data", "error should indicate missing data")
}

// TestAnalyzeConceptRangeGeneratorScope_WrongDataType_ReturnsError verifies that
// the handler returns an error when table.Data is not *ast.ConceptRangeGeneratorDef.
func TestAnalyzeConceptRangeGeneratorScope_WrongDataType_ReturnsError(t *testing.T) {
	t.Parallel()

	table := &symboltable.SymbolTable{
		ScopeType: symboltable.ConceptGeneratorScope,
		Data:      "not a ConceptRangeGeneratorDef",
	}
	context := analysis.NewAnalysisContext()
	analyzer := analysis.NewCPLAnalyzer()

	err := analysis.AnalyzeConceptRangeGeneratorScope(table, context, analyzer)

	require.Error(t, err, "should return error for wrong data type")
	assert.Contains(t, err.Error(), "not *ast.ConceptRangeGeneratorDef", "error should indicate wrong type")
}

// TestAnalyzeConceptRangeGeneratorScope_AddsGeneratorToConcept verifies that
// the handler creates a ConceptRangeGenerator with the correct fields and adds
// it to the current concept on the context stack.
func TestAnalyzeConceptRangeGeneratorScope_AddsGeneratorToConcept(t *testing.T) {
	t.Parallel()

	// Arrange: Set up a symbol table with ConceptRangeGeneratorDef data
	table := &symboltable.SymbolTable{
		ScopeType: symboltable.ConceptGeneratorScope,
		Data: &ast.ConceptRangeGeneratorDef{
			ConceptName: "Country",
			StartValue:  "Canada",
			EndValue:    "UK",
		},
	}

	// Create context with a concept pushed onto it
	analyzer := analysis.NewCPLAnalyzer()
	context := analysis.NewAnalysisContext()
	conceptDef := analysis.NewAnalysisConceptDefinition("Subset", 1, analyzer, symboltable.SourceLocation{Line: 1, Column: 1})
	err := context.Push(conceptDef)
	require.NoError(t, err, "push should succeed")

	// Act: Call the handler
	err = analysis.AnalyzeConceptRangeGeneratorScope(table, context, analyzer)

	// Assert
	require.NoError(t, err, "handler should succeed with valid data")

	// Pop the concept and verify the generator was added
	popped, err := context.Pop()
	require.NoError(t, err, "pop should succeed")
	require.NotNil(t, popped.ConceptDefinition, "concept definition should not be nil")

	concept := popped.ConceptDefinition
	require.Len(t, concept.Generators, 1, "concept should have 1 generator added")

	crg, ok := concept.Generators[0].(*analysis.ConceptRangeGenerator)
	require.True(t, ok, "generator should be *ConceptRangeGenerator, got %T", concept.Generators[0])
	assert.Equal(t, "Country", crg.ConceptName, "ConceptName should match")
	assert.Equal(t, "Canada", crg.StartValue, "StartValue should match")
	assert.Equal(t, "UK", crg.EndValue, "EndValue should match")

	// Verify no values were generated yet
	assert.Len(t, concept.ResolvedValues, 0, "ResolvedValues should be empty; generation is deferred")
}

// TestAnalyzeConceptRangeGeneratorScope_DeferredGeneration verifies that
// the ConceptRangeGenerator does not generate values immediately (unlike
// ValuesGenerator or IntRangeGenerator) because the source concept may not
// be resolved yet.
func TestAnalyzeConceptRangeGeneratorScope_DeferredGeneration(t *testing.T) {
	t.Parallel()

	// Arrange
	table := &symboltable.SymbolTable{
		ScopeType: symboltable.ConceptGeneratorScope,
		Data: &ast.ConceptRangeGeneratorDef{
			ConceptName: "Source",
			StartValue:  "A",
			EndValue:    "B",
		},
	}

	analyzer := analysis.NewCPLAnalyzer()
	context := analysis.NewAnalysisContext()
	conceptDef := analysis.NewAnalysisConceptDefinition("Target", 1, analyzer, symboltable.SourceLocation{Line: 1, Column: 1})
	context.Push(conceptDef)

	// Act
	err := analysis.AnalyzeConceptRangeGeneratorScope(table, context, analyzer)

	// Assert: Should NOT return error even though Source concept doesn't exist yet
	require.NoError(t, err, "handler should succeed without source concept; resolution is deferred")

	popped, _ := context.Pop()
	assert.Len(t, popped.ConceptDefinition.ResolvedValues, 0, "no values should be generated yet")
}

// ============================================================================
// Concept Range Integration Tests via WalkSymbolTableWithHandlers (Task 6)
// ============================================================================

// TestConceptRangeIntegration_WalkAndResolve_BasicRange verifies the full
// analysis pipeline: handler is invoked via WalkSymbolTableWithHandlers,
// then ResolveConceptRanges resolves the deferred values.
func TestConceptRangeIntegration_WalkAndResolve_BasicRange(t *testing.T) {
	t.Parallel()

	// Arrange: Build a symbol table resembling:
	//   concepts
	//     Country=[Values[USA, Canada, Mexico, UK, France]]
	//     Subset=[Country[Canada..UK]]
	//
	// Both concepts are added to the analyzer; Subset has a ConceptRangeGenerator
	// that references Country.
	analyzer := analysis.NewCPLAnalyzer()

	// Add source concept directly (simulating that it was processed first)
	source := analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada", "Mexico", "UK", "France"})
	analyzer.AddNewConcept(&source)

	// Build symbol table for Subset with a ConceptRangeGenerator
	root := symboltable.NewGlobalSymbolTable()
	conceptsScope := root.AddChild(symboltable.ConceptsScope)
	conceptScope := conceptsScope.AddChild(symboltable.ConceptScope)
	conceptName := "Subset"
	conceptScope.Define(&conceptName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
	crgScope := conceptScope.AddChild(symboltable.ConceptGeneratorScope)
	crgScope.Data = &ast.ConceptRangeGeneratorDef{
		ConceptName: "Country",
		StartValue:  "Canada",
		EndValue:    "UK",
	}

	// Act: Walk the symbol table with handlers including ConceptGeneratorScope
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptsScope:         analysis.EnterAnalyzeConceptsScope,
		symboltable.ConceptScope:          analysis.EnterAnalyzeConceptScope,
		symboltable.ConceptGeneratorScope: analysis.AnalyzeConceptRangeGeneratorScope,
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
	}

	err := analysis.WalkSymbolTableWithHandlers(root, enterHandlers, exitHandlers, analyzer)
	require.NoError(t, err, "walk should succeed")

	// Resolve concept ranges
	err = analyzer.ResolveConceptRanges()
	require.NoError(t, err, "resolution should succeed")

	// Assert: Verify Subset has the expected values
	defs := analyzer.GetAnalysisDefinitions()
	subset, exists := defs.Concepts["Subset"]
	require.True(t, exists, "Subset concept should exist")
	require.Len(t, subset.ResolvedValues, 3, "Subset should have 3 resolved values")
	assert.Equal(t, "Canada", subset.ResolvedValues[0].Name)
	assert.Equal(t, "Mexico", subset.ResolvedValues[1].Name)
	assert.Equal(t, "UK", subset.ResolvedValues[2].Name)
}

// TestConceptRangeIntegration_WalkAndResolve_NoSourceConcept_ReturnsError verifies
// that ResolveConceptRanges returns an error when the referenced source concept
// does not exist in the analyzer.
func TestConceptRangeIntegration_WalkAndResolve_NoSourceConcept_ReturnsError(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()

	// Build symbol table for Target referencing "Missing"
	root := symboltable.NewGlobalSymbolTable()
	conceptsScope := root.AddChild(symboltable.ConceptsScope)
	conceptScope := conceptsScope.AddChild(symboltable.ConceptScope)
	conceptName := "Target"
	conceptScope.Define(&conceptName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
	crgScope := conceptScope.AddChild(symboltable.ConceptGeneratorScope)
	crgScope.Data = &ast.ConceptRangeGeneratorDef{
		ConceptName: "Missing",
		StartValue:  "A",
		EndValue:    "B",
	}

	// Act: Walk the symbol table
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptsScope:         analysis.EnterAnalyzeConceptsScope,
		symboltable.ConceptScope:          analysis.EnterAnalyzeConceptScope,
		symboltable.ConceptGeneratorScope: analysis.AnalyzeConceptRangeGeneratorScope,
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
	}

	err := analysis.WalkSymbolTableWithHandlers(root, enterHandlers, exitHandlers, analyzer)
	require.NoError(t, err, "walk should succeed (deferred resolution)")

	// Resolve concept ranges — should fail because "Missing" doesn't exist
	err = analyzer.ResolveConceptRanges()
	require.Error(t, err, "resolution should fail for missing source concept")
	assert.Contains(t, err.Error(), "unknown concept", "error should indicate unknown source concept")
}

// TestConceptRangeIntegration_FullAnalyze_BasicRange verifies the complete
// analysis pipeline via the Analyze() function, which includes the handler
// via entryHandlers and ResolveConceptRanges in the correct order.
func TestConceptRangeIntegration_FullAnalyze_BasicRange(t *testing.T) {
	t.Parallel()

	// Parse CPL with concept range
	input := "concepts\n\tQuarter=[Values[Q1, Q2, Q3, Q4]]\n\tFirstHalf=[Quarter[Q1..Q2]]"
	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "parsing should succeed")

	// Act: Full analysis
	defs, err := analysis.Analyze(rootTable)

	// Assert
	require.NoError(t, err, "full analysis should succeed")

	source, exists := defs.Concepts["Quarter"]
	require.True(t, exists, "Quarter concept should exist")
	assert.Len(t, source.ResolvedValues, 4, "Quarter should have 4 values")

	firstHalf, exists := defs.Concepts["FirstHalf"]
	require.True(t, exists, "FirstHalf concept should exist")
	require.Len(t, firstHalf.ResolvedValues, 2, "FirstHalf should have 2 values")
	assert.Equal(t, "Q1", firstHalf.ResolvedValues[0].Name)
	assert.Equal(t, "Q2", firstHalf.ResolvedValues[1].Name)
}

// TestConceptRangeIntegration_FullAnalyze_ChainedReference verifies that
// chained concept range references resolve correctly via the full Analyze pipeline.
func TestConceptRangeIntegration_FullAnalyze_ChainedReference(t *testing.T) {
	t.Parallel()

	// Parse CPL with chained concept range references
	// Use named values since Concept Range expects identifiers, not numeric literals
	input := "concepts\n\tBase=[Values[V1, V2, V3, V4, V5]]\n\tMiddle=[Base[V1..V3]]\n\tTop=[Middle[V1..V2]]"
	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "parsing should succeed")

	// Act: Full analysis
	defs, err := analysis.Analyze(rootTable)

	// Assert
	require.NoError(t, err, "full analysis should succeed for chained references")

	middle, exists := defs.Concepts["Middle"]
	require.True(t, exists, "Middle concept should exist")
	require.Len(t, middle.ResolvedValues, 3, "Middle should have 3 values")
	assert.Equal(t, "V1", middle.ResolvedValues[0].Name)
	assert.Equal(t, "V2", middle.ResolvedValues[1].Name)
	assert.Equal(t, "V3", middle.ResolvedValues[2].Name)

	top, exists := defs.Concepts["Top"]
	require.True(t, exists, "Top concept should exist")
	require.Len(t, top.ResolvedValues, 2, "Top should have 2 values")
	assert.Equal(t, "V1", top.ResolvedValues[0].Name)
	assert.Equal(t, "V2", top.ResolvedValues[1].Name)
}

// TestConceptRangeIntegration_FullAnalyze_CircularDependency_ReturnsError verifies
// that circular concept range references produce an error during full analysis.
func TestConceptRangeIntegration_FullAnalyze_CircularDependency_ReturnsError(t *testing.T) {
	t.Parallel()

	// Parse CPL with circular references
	input := "concepts\n\tA=[B[Value1..Value2]]\n\tB=[A[Value1..Value2]]"
	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "parsing should succeed")

	// Act: Full analysis
	defs, err := analysis.Analyze(rootTable)

	// Assert
	require.Error(t, err, "analysis should fail for circular dependencies")
	assert.Contains(t, err.Error(), "circular dependency", "error should indicate circular dependency")
	assert.Nil(t, defs, "definitions should be nil on error")
}

// TestConceptRangeIntegration_FullAnalyze_WithValuesAndIntRange verifies that
// ConceptRange works with concepts built from Values and IntRange generators.
func TestConceptRangeIntegration_FullAnalyze_WithValuesAndIntRange(t *testing.T) {
	t.Parallel()

	// Parse CPL mixing Values and ConceptRange from two source concepts
	// Use named values since Concept Range expects identifiers
	input := "concepts\n\tCountry=[Values[USA, Canada, Mexico, UK, France]]\n\tStatus=[Values[Active, Inactive, Pending, Done]]\n\tSubset=[Country[Canada..UK], Status[Active..Pending]]"
	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "parsing should succeed")

	// Act: Full analysis
	defs, err := analysis.Analyze(rootTable)

	// Assert
	require.NoError(t, err, "full analysis should succeed")

	subset, exists := defs.Concepts["Subset"]
	require.True(t, exists, "Subset concept should exist")
	require.Len(t, subset.ResolvedValues, 6, "Subset should have 6 values (3 from Country + 3 from Status)")

	// First 3 values from Country range
	assert.Equal(t, "Canada", subset.ResolvedValues[0].Name)
	assert.Equal(t, "Mexico", subset.ResolvedValues[1].Name)
	assert.Equal(t, "UK", subset.ResolvedValues[2].Name)

	// Next 3 values from Status range
	assert.Equal(t, "Active", subset.ResolvedValues[3].Name)
	assert.Equal(t, "Inactive", subset.ResolvedValues[4].Name)
	assert.Equal(t, "Pending", subset.ResolvedValues[5].Name)
}

// ============================================================================
// Helper Functions for Concept Range Tests
// ============================================================================

// createMockSymbolTableWithConceptRange builds a symbol table with a single
// concept that has a ConceptRangeGenerator.
func createMockSymbolTableWithConceptRange(name, sourceConcept, startValue, endValue string) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()
	conceptsScope := root.AddChild(symboltable.ConceptsScope)
	conceptScope := conceptsScope.AddChild(symboltable.ConceptScope)

	cn := name
	conceptScope.Define(&cn, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})

	crgScope := conceptScope.AddChild(symboltable.ConceptGeneratorScope)
	crgScope.Data = &ast.ConceptRangeGeneratorDef{
		ConceptName: sourceConcept,
		StartValue:  startValue,
		EndValue:    endValue,
	}

	return root
}
