// Package analysis_test contains integration tests for lookup analysis.
//
// These tests verify that the lookup handler is correctly registered and invoked
// during the symbol table walking process. They test the full integration from
// parsing to analysis, ensuring lookups are properly analyzed when calling
// the top-level Analyze function.
//
// Story: Story242 - Lookups Increment 1 - Analysis Phase
// Task: Task 3 - Write integration test for lookup analysis pipeline
package analysis_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Integration Tests - Handler Registration
// ============================================================================

// TestAnalyze_LookupHandler_Registered verifies that the lookup handler
// is correctly registered and invoked during the Analyze function.
func TestAnalyze_LookupHandler_Registered(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with concepts, measures, and lookups
	input := `concepts
	Country = [Values[USA, UK, Germany]]
	Product = [Values[Widget, Gadget]]

measures
	UnitsSold = Number(unit=Goods)

lookups
	CountryProductIndex
		dimensions [Country, Product]
		celltype UnitsSold`

	// Act: Parse and analyze
	defs, err := parseAndAnalyzeCPL(input)

	// Assert: No error and lookup was analyzed
	require.NoError(t, err, "Should not return error for valid CPL with lookups")
	require.NotNil(t, defs, "AnalysisDefinitions should not be nil")

	// Verify lookup exists
	require.Contains(t, defs.Lookups, "CountryProductIndex", "CountryProductIndex lookup should be analyzed")
	lookup := defs.Lookups["CountryProductIndex"]
	assert.Equal(t, "CountryProductIndex", lookup.Name)
}

// TestAnalyze_LookupWithDimensions_ResolvedCorrectly verifies that
// dimension references are correctly resolved to concept definitions.
func TestAnalyze_LookupWithDimensions_ResolvedCorrectly(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with lookup dimensions
	input := `concepts
	Country = [Values[USA, UK, Germany]]
	Product = [Values[Widget, Gadget, Thingama]]
	Date = [Values[Jan2024, Feb2024]]

measures
	Amount = Number(unit=Dollar)

lookups
	SalesIndex
		dimensions [Country, Product, Date]
		celltype Amount`

	// Act: Parse and analyze
	defs, err := parseAndAnalyzeCPL(input)

	// Assert: Dimensions are correctly resolved
	require.NoError(t, err)
	require.Contains(t, defs.Lookups, "SalesIndex")

	lookup := defs.Lookups["SalesIndex"]
	require.Len(t, lookup.Dimensions, 3, "Should have 3 dimensions")

	// Verify each dimension points to the correct concept
	assert.Equal(t, "Country", lookup.Dimensions[0].ConceptName)
	assert.Equal(t, "Product", lookup.Dimensions[1].ConceptName)
	assert.Equal(t, "Date", lookup.Dimensions[2].ConceptName)
}

// TestAnalyze_LookupCelltypeMeasure_ResolvedCorrectly verifies that
// the celltype measure reference is correctly resolved.
func TestAnalyze_LookupCelltypeMeasure_ResolvedCorrectly(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with a lookup
	input := `concepts
	Country = [Values[USA, UK]]

measures
	Revenue = Number(unit=Dollar)

lookups
	CountryIndex
		dimensions [Country]
		celltype Revenue`

	// Act: Parse and analyze
	defs, err := parseAndAnalyzeCPL(input)

	// Assert: Celltype measure is correctly resolved
	require.NoError(t, err)
	require.Contains(t, defs.Lookups, "CountryIndex")

	lookup := defs.Lookups["CountryIndex"]
	require.NotNil(t, lookup.CelltypeMeasure, "CelltypeMeasure should not be nil")
	assert.Equal(t, "Revenue", lookup.CelltypeMeasure.Name)
	assert.Equal(t, "Number", lookup.CelltypeMeasure.TypeName)
	// UnitName disabled pending completion of story230
	// assert.Equal(t, "Dollar", lookup.CelltypeMeasure.UnitName)
}

// ============================================================================
// CIL Tree Generation Tests
// ============================================================================

// TestAnalyze_LookupCILTree_Generated verifies that a CIL tree is generated
// for each lookup during analysis.
func TestAnalyze_LookupCILTree_Generated(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with a lookup
	input := `concepts
	Country = [Values[USA, UK]]
	Product = [Values[Widget, Gadget]]

measures
	UnitsSold = Number(unit=Goods)

lookups
	CountryProductIndex
		dimensions [Country, Product]
		celltype UnitsSold`

	// Act: Parse and analyze
	defs, err := parseAndAnalyzeCPL(input)

	// Assert: CIL tree was generated with "lookup:" prefix
	require.NoError(t, err)
	require.Contains(t, defs.CILTrees, "lookup:CountryProductIndex", "CIL tree should exist for CountryProductIndex")

	cilNode := defs.CILTrees["lookup:CountryProductIndex"]
	require.NotNil(t, cilNode)

	// Type assert to verify it's a lookup node with correct name
	lookupNode, ok := cilNode.(*cil.CILLookupNode)
	require.True(t, ok, "Should be CILLookupNode")
	assert.Equal(t, "CountryProductIndex", lookupNode.LookupName)
	assert.Equal(t, []string{"Country", "Product"}, lookupNode.Dimensions)
	assert.Equal(t, "UnitsSold", lookupNode.Celltype)
}

// ============================================================================
// Error Propagation Tests
// ============================================================================

// TestAnalyze_LookupUndefinedDimension_ReturnsError verifies that analysis
// fails when a lookup references an undefined concept dimension.
func TestAnalyze_LookupUndefinedDimension_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with a lookup referencing undefined concept
	input := `concepts
	Country = [Values[USA, UK]]

measures
	Revenue = Number(unit=Dollar)

lookups
	SalesIndex
		dimensions [Country, UndefinedConcept]
		celltype Revenue`

	// Act: Parse and analyze
	_, err := parseAndAnalyzeCPL(input)

	// Assert: Error is returned mentioning undefined dimension
	require.Error(t, err, "Should return error for undefined dimension")
	assert.Contains(t, err.Error(), "undefined concept dimension")
	assert.Contains(t, err.Error(), "UndefinedConcept")
}

// TestAnalyze_LookupUndefinedCelltype_ReturnsError verifies that analysis
// fails when a lookup references an undefined measure celltype.
func TestAnalyze_LookupUndefinedCelltype_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with a lookup referencing undefined measure
	input := `concepts
	Country = [Values[USA, UK]]

measures
	Revenue = Number(unit=Dollar)

lookups
	SalesIndex
		dimensions [Country]
		celltype UndefinedMeasure`

	// Act: Parse and analyze
	_, err := parseAndAnalyzeCPL(input)

	// Assert: Error is returned mentioning undefined celltype
	require.Error(t, err, "Should return error for undefined celltype")
	assert.Contains(t, err.Error(), "undefined celltype measure")
	assert.Contains(t, err.Error(), "UndefinedMeasure")
}

// TestAnalyze_LookupDuplicateName_ReturnsError verifies that analysis
// fails when duplicate lookup names are defined.
func TestAnalyze_LookupDuplicateName_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with duplicate lookup names
	input := `concepts
	Country = [Values[USA, UK]]

measures
	Revenue = Number(unit=Dollar)

lookups
	SalesIndex
		dimensions [Country]
		celltype Revenue
	SalesIndex
		dimensions [Country]
		celltype Revenue`

	// Act: Parse and analyze
	_, err := parseAndAnalyzeCPL(input)

	// Assert: Error is returned mentioning duplicate
	require.Error(t, err, "Should return error for duplicate lookup")
	assert.Contains(t, err.Error(), "duplicate lookup definition")
	assert.Contains(t, err.Error(), "SalesIndex")
}

// ============================================================================
// Full Integration Tests - All Sections
// ============================================================================

// TestAnalyze_FullProgram_WithLookups verifies that a complete CPL program
// with concepts, measures, and lookups is correctly analyzed.
func TestAnalyze_FullProgram_WithLookups(t *testing.T) {
	t.Parallel()

	// Arrange: Create a complete CPL program
	input := `concepts
	Country = [Values[USA, UK, Germany, France]]
	Product = [Values[Widget, Gadget, Thing]]

measures
	Revenue = Number(unit=Dollar)
	Quantity = Number(unit=Unit)

lookups
	SalesIndex
		dimensions [Country, Product]
		celltype Revenue
	InventoryIndex
		dimensions [Country, Product]
		celltype Quantity`

	// Act: Parse and analyze
	defs, err := parseAndAnalyzeCPL(input)

	// Assert: All sections were analyzed correctly
	require.NoError(t, err)
	require.NotNil(t, defs)

	// Verify concepts. We don't check the exact total because standard time
	// concepts (Month, DayOfWeek, Hour, etc.) are injected by the analyzer.
	assert.Contains(t, defs.Concepts, "Country")
	assert.Contains(t, defs.Concepts, "Product")

	// Verify measures
	assert.Len(t, defs.Measures, 2, "Should have 2 measures")
	assert.Contains(t, defs.Measures, "Revenue")
	assert.Contains(t, defs.Measures, "Quantity")

	// Verify lookups
	assert.Len(t, defs.Lookups, 2, "Should have 2 lookups")
	assert.Contains(t, defs.Lookups, "SalesIndex")
	assert.Contains(t, defs.Lookups, "InventoryIndex")

	// Verify lookup details
	salesLookup := defs.Lookups["SalesIndex"]
	assert.Equal(t, "SalesIndex", salesLookup.Name)
	assert.Len(t, salesLookup.Dimensions, 2)
	assert.Equal(t, "Country", salesLookup.Dimensions[0].ConceptName)
	assert.Equal(t, "Product", salesLookup.Dimensions[1].ConceptName)
	assert.Equal(t, "Revenue", salesLookup.CelltypeMeasure.Name)

	invLookup := defs.Lookups["InventoryIndex"]
	assert.Equal(t, "InventoryIndex", invLookup.Name)
	assert.Equal(t, "Quantity", invLookup.CelltypeMeasure.Name)

	// Verify CIL trees
	assert.Len(t, defs.CILTrees, 2, "Should have 2 CIL trees")
	assert.Contains(t, defs.CILTrees, "lookup:SalesIndex")
	assert.Contains(t, defs.CILTrees, "lookup:InventoryIndex")
}

// ============================================================================
// Source Location Tests
// ============================================================================

// TestAnalyze_LookupSourceLocation_Captured verifies that source location
// information is correctly captured for lookups.
func TestAnalyze_LookupSourceLocation_Captured(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with a lookup
	input := `concepts
	Country = [Values[USA]]

measures
	Revenue = Number(unit=Dollar)

lookups
	SalesIndex
		dimensions [Country]
		celltype Revenue`

	// Act: Parse and analyze
	defs, err := parseAndAnalyzeCPL(input)

	// Assert: Source location was captured
	require.NoError(t, err)
	require.Contains(t, defs.Lookups, "SalesIndex")

	lookup := defs.Lookups["SalesIndex"]
	assert.Greater(t, lookup.SourceLocation.Line, 0, "Line should be set")
	assert.GreaterOrEqual(t, lookup.SourceLocation.Column, 0, "Column should be set")
}

// ============================================================================
// WalkSymbolTableWithHandlers Integration Tests
// ============================================================================

// TestWalkSymbolTableWithHandlers_LookupsScope_Invoked verifies that the
// lookup handler is correctly invoked when walking a symbol table with a LookupsScope.
// This test sets up the analyzer state directly (as would be done by previous scope handlers)
// and verifies that AnalyzeLookupScope processes the lookup correctly.
func TestWalkSymbolTableWithHandlers_LookupsScope_Invoked(t *testing.T) {
	t.Parallel()

	// Arrange: Build a symbol table tree with a lookups scope
	root := symboltable.NewGlobalSymbolTable()

	// Create lookups scope with a lookup definition
	lookupsScope := root.AddChild(symboltable.LookupsScope)
	lookupsScope.Definitions = []symboltable.SymbolDefinition{
		{
			Name:           strPtr("SalesLookup"),
			DefinitionType: symboltable.LookupScopeDefinition,
			Location:       symboltable.SourceLocation{Line: 10, Column: 1},
			Data: &parser.ParserLookupDefinition{
				Name:           "SalesLookup",
				Dimensions:     []string{"Country"},
				Celltype:       "Revenue",
				SourceLocation: symboltable.SourceLocation{Line: 10, Column: 1},
			},
		},
	}

	// Create analyzer with pre-populated concepts and measures
	// (simulating what would happen after ConceptScope and MeasureScope handlers run)
	analyzer := analysis.NewCPLAnalyzer()
	analyzer.AddNewConcept(&analysis.AnalysisConceptDefinition{ConceptName: "Country"})
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
		// UnitName disabled pending completion of story230
		// UnitName: "Dollar",
	}

	// Register the lookup handler for LookupsScope
	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.LookupsScope: analysis.AnalyzeLookupScope,
	}

	// Act: Walk the symbol table
	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	// Assert: No error and lookup was processed
	require.NoError(t, err, "Walk should not return error")

	resultDefs := analyzer.GetAnalysisDefinitions()
	assert.Len(t, resultDefs.Lookups, 1, "Should have 1 lookup")
	assert.Contains(t, resultDefs.Lookups, "SalesLookup")
	assert.Equal(t, "SalesLookup", resultDefs.Lookups["SalesLookup"].Name)
	assert.Len(t, resultDefs.Lookups["SalesLookup"].Dimensions, 1, "Should have 1 dimension")
	assert.Equal(t, "Revenue", resultDefs.Lookups["SalesLookup"].CelltypeMeasure.Name)
}

// ============================================================================
// Lookup Error Example Files Integration Tests
// ============================================================================

// TestAnalyze_LookupWithErrors_Examples verifies that lookup error example files
// in examples/parse_and_analyze_only/with_errors/ produce expected analysis errors.
// These files serve as regression tests for the lookup analysis phase.
func TestAnalyze_LookupWithErrors_Examples(t *testing.T) {
	t.Parallel()

	examplesDir := getExamplesDir(t)
	withErrorsDir := filepath.Join(examplesDir, "parse_and_analyze_only", "with_errors")

	entries, err := os.ReadDir(withErrorsDir)
	require.NoError(t, err, "Failed to read with_errors directory")

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cpl") {
			continue
		}
		// Only test lookup-specific error examples
		if !strings.HasPrefix(entry.Name(), "lookup_") {
			continue
		}

		entry := entry // capture range variable for closure
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(withErrorsDir, entry.Name())
			_, err := getAnalysisForFile(t, filePath)

			require.Error(t, err, "Expected analysis error for %s", entry.Name())

			switch entry.Name() {
			case "lookup_duplicate_name.cpl":
				assert.Contains(t, err.Error(), "duplicate lookup definition")
			case "lookup_undefined_dimension.cpl":
				assert.Contains(t, err.Error(), "undefined concept dimension")
				assert.Contains(t, err.Error(), "UndefinedConcept")
			case "lookup_undefined_celltype.cpl":
				assert.Contains(t, err.Error(), "undefined celltype measure")
				assert.Contains(t, err.Error(), "UndefinedMeasure")
			}
		})
	}
}


