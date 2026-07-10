// Package analysis contains unit tests for the AnalyzeLookupScope function.
//
// These tests verify that the lookup analysis phase correctly:
//   - Resolves dimension references from the Concepts map
//   - Resolves celltype references from the Measures map
//   - Detects duplicate lookup definitions
//   - Reports errors for undefined dimensions or celltypes
//   - Generates CIL trees for valid lookups
//
// Story: Story242 - Lookups Increment 1 - Analysis Phase
package analysis

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Happy Path Tests - Valid Lookup Processing
// ============================================================================

// TestAnalyzeLookupScope_SingleDimension_Valid verifies that a lookup with a
// single dimension is correctly processed and stored in the analyzer.
func TestAnalyzeLookupScope_SingleDimension_Valid(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with pre-loaded concepts and measures
	analyzer := NewCPLAnalyzer()
	analyzer.AddNewConcept(&AnalysisConceptDefinition{ConceptName: "Store"})
	analyzer.measures["Revenue"] = AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
		// UnitName disabled pending completion of story230
		// UnitName: "Dollar",
	}

	// Build a LookupsScope with one lookup definition
	lookupsScope := symboltable.NewSymbolTable(symboltable.LookupsScope)
	lookupName := "SalesLookup"
	lookupsScope.Definitions = []symboltable.SymbolDefinition{
		{
			Name:           &lookupName,
			DefinitionType: symboltable.LookupScopeDefinition,
			Location:       symboltable.SourceLocation{Line: 1, Column: 1},
			Data: &parser.ParserLookupDefinition{
				Name:           "SalesLookup",
				Dimensions:     []string{"Store"},
				Celltype:       "Revenue",
				SourceLocation: symboltable.SourceLocation{Line: 1, Column: 1},
			},
		},
	}

	// Act: Analyze the lookups scope
	err := AnalyzeLookupScope(lookupsScope, NewAnalysisContext(), analyzer)

	// Assert: No error occurred
	require.NoError(t, err, "Should not return error for valid lookup")

	// Assert: Lookup was added to the map
	require.Contains(t, analyzer.lookups, "SalesLookup")
	lookup := analyzer.lookups["SalesLookup"]
	assert.Equal(t, "SalesLookup", lookup.Name)
	require.Len(t, lookup.Dimensions, 1)
	assert.Equal(t, "Store", lookup.Dimensions[0].ConceptName)
	require.NotNil(t, lookup.CelltypeMeasure)
	assert.Equal(t, "Revenue", lookup.CelltypeMeasure.Name)

	// Assert: CIL tree was created with "lookup:" prefix
	cilTrees := analyzer.metricAnalyzer.GetCILTrees()
	require.Contains(t, cilTrees, "lookup:SalesLookup")
	node, ok := cilTrees["lookup:SalesLookup"].(*cil.CILLookupNode)
	require.True(t, ok, "CIL node should be *cil.CILLookupNode")
	assert.Equal(t, "SalesLookup", node.LookupName)
	assert.Equal(t, []string{"Store"}, node.Dimensions)
	assert.Equal(t, "Revenue", node.Celltype)
}

// TestAnalyzeLookupScope_MultipleDimensions_Valid verifies that a lookup with
// multiple dimensions is correctly processed.
func TestAnalyzeLookupScope_MultipleDimensions_Valid(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with pre-loaded concepts and measures
	analyzer := NewCPLAnalyzer()
	analyzer.AddNewConcept(&AnalysisConceptDefinition{ConceptName: "Country"})
	analyzer.AddNewConcept(&AnalysisConceptDefinition{ConceptName: "Product"})
	analyzer.measures["UnitsSold"] = AnalysisMeasureDefinition{
		Name:     "UnitsSold",
		TypeName: "Number",
		// UnitName disabled pending completion of story230
		// UnitName: "Unit",
	}

	// Build a LookupsScope with one lookup definition
	lookupsScope := symboltable.NewSymbolTable(symboltable.LookupsScope)
	lookupName := "CountryProductIndex"
	lookupsScope.Definitions = []symboltable.SymbolDefinition{
		{
			Name:           &lookupName,
			DefinitionType: symboltable.LookupScopeDefinition,
			Location:       symboltable.SourceLocation{Line: 1, Column: 1},
			Data: &parser.ParserLookupDefinition{
				Name:           "CountryProductIndex",
				Dimensions:     []string{"Country", "Product"},
				Celltype:       "UnitsSold",
				SourceLocation: symboltable.SourceLocation{Line: 1, Column: 1},
			},
		},
	}

	// Act: Analyze the lookups scope
	err := AnalyzeLookupScope(lookupsScope, NewAnalysisContext(), analyzer)

	// Assert: No error occurred
	require.NoError(t, err, "Should not return error for valid lookup with multiple dimensions")

	// Assert: Lookup was added with correct dimensions
	require.Contains(t, analyzer.lookups, "CountryProductIndex")
	lookup := analyzer.lookups["CountryProductIndex"]
	require.Len(t, lookup.Dimensions, 2)
	assert.Equal(t, "Country", lookup.Dimensions[0].ConceptName)
	assert.Equal(t, "Product", lookup.Dimensions[1].ConceptName)
	assert.Equal(t, "UnitsSold", lookup.CelltypeMeasure.Name)

	// Assert: CIL tree exists
	cilTrees := analyzer.metricAnalyzer.GetCILTrees()
	require.Contains(t, cilTrees, "lookup:CountryProductIndex")
	node, ok := cilTrees["lookup:CountryProductIndex"].(*cil.CILLookupNode)
	require.True(t, ok)
	assert.Equal(t, []string{"Country", "Product"}, node.Dimensions)
	assert.Equal(t, "UnitsSold", node.Celltype)
}

// ============================================================================
// Error Cases - Validation Failures
// ============================================================================

// TestAnalyzeLookupScope_DuplicateName_ReturnsError verifies that two lookup
// definitions with the same name in the same scope produce an error.
func TestAnalyzeLookupScope_DuplicateName_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with pre-loaded concepts and measures
	analyzer := NewCPLAnalyzer()
	analyzer.AddNewConcept(&AnalysisConceptDefinition{ConceptName: "Store"})
	analyzer.measures["Revenue"] = AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
	}

	// Build a LookupsScope with two duplicate lookup definitions
	lookupsScope := symboltable.NewSymbolTable(symboltable.LookupsScope)
	name1 := "SalesLookup"
	name2 := "SalesLookup"
	lookupsScope.Definitions = []symboltable.SymbolDefinition{
		{
			Name:           &name1,
			DefinitionType: symboltable.LookupScopeDefinition,
			Location:       symboltable.SourceLocation{Line: 1, Column: 1},
			Data: &parser.ParserLookupDefinition{
				Name:           "SalesLookup",
				Dimensions:     []string{"Store"},
				Celltype:       "Revenue",
				SourceLocation: symboltable.SourceLocation{Line: 1, Column: 1},
			},
		},
		{
			Name:           &name2,
			DefinitionType: symboltable.LookupScopeDefinition,
			Location:       symboltable.SourceLocation{Line: 3, Column: 1},
			Data: &parser.ParserLookupDefinition{
				Name:           "SalesLookup",
				Dimensions:     []string{"Store"},
				Celltype:       "Revenue",
				SourceLocation: symboltable.SourceLocation{Line: 3, Column: 1},
			},
		},
	}

	// Act: Analyze the lookups scope
	err := AnalyzeLookupScope(lookupsScope, NewAnalysisContext(), analyzer)

	// Assert: Error mentioning duplicate lookup definition and the name
	require.Error(t, err, "Should return error for duplicate lookup name")
	assert.Contains(t, err.Error(), "duplicate lookup definition")
	assert.Contains(t, err.Error(), "SalesLookup")
}

// TestAnalyzeLookupScope_UndefinedDimension_ReturnsError verifies that a lookup
// referencing an undefined concept dimension produces an error.
func TestAnalyzeLookupScope_UndefinedDimension_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with measure but NO concept
	analyzer := NewCPLAnalyzer()
	analyzer.measures["Revenue"] = AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
	}

	// Build a LookupsScope with one lookup referencing undefined concept
	lookupsScope := symboltable.NewSymbolTable(symboltable.LookupsScope)
	lookupName := "SalesLookup"
	lookupsScope.Definitions = []symboltable.SymbolDefinition{
		{
			Name:           &lookupName,
			DefinitionType: symboltable.LookupScopeDefinition,
			Location:       symboltable.SourceLocation{Line: 1, Column: 1},
			Data: &parser.ParserLookupDefinition{
				Name:           "SalesLookup",
				Dimensions:     []string{"UndefinedStore"},
				Celltype:       "Revenue",
				SourceLocation: symboltable.SourceLocation{Line: 1, Column: 1},
			},
		},
	}

	// Act: Analyze the lookups scope
	err := AnalyzeLookupScope(lookupsScope, NewAnalysisContext(), analyzer)

	// Assert: Error mentioning undefined concept dimension and the dimension name
	require.Error(t, err, "Should return error for undefined dimension")
	assert.Contains(t, err.Error(), "undefined concept dimension")
	assert.Contains(t, err.Error(), "UndefinedStore")
}

// TestAnalyzeLookupScope_UndefinedCelltype_ReturnsError verifies that a lookup
// referencing an undefined celltype measure produces an error.
func TestAnalyzeLookupScope_UndefinedCelltype_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with concept but NO measure
	analyzer := NewCPLAnalyzer()
	analyzer.AddNewConcept(&AnalysisConceptDefinition{ConceptName: "Store"})

	// Build a LookupsScope with one lookup referencing undefined measure
	lookupsScope := symboltable.NewSymbolTable(symboltable.LookupsScope)
	lookupName := "SalesLookup"
	lookupsScope.Definitions = []symboltable.SymbolDefinition{
		{
			Name:           &lookupName,
			DefinitionType: symboltable.LookupScopeDefinition,
			Location:       symboltable.SourceLocation{Line: 1, Column: 1},
			Data: &parser.ParserLookupDefinition{
				Name:           "SalesLookup",
				Dimensions:     []string{"Store"},
				Celltype:       "UndefinedRevenue",
				SourceLocation: symboltable.SourceLocation{Line: 1, Column: 1},
			},
		},
	}

	// Act: Analyze the lookups scope
	err := AnalyzeLookupScope(lookupsScope, NewAnalysisContext(), analyzer)

	// Assert: Error mentioning undefined celltype measure and the measure name
	require.Error(t, err, "Should return error for undefined celltype")
	assert.Contains(t, err.Error(), "undefined celltype measure")
	assert.Contains(t, err.Error(), "UndefinedRevenue")
}
