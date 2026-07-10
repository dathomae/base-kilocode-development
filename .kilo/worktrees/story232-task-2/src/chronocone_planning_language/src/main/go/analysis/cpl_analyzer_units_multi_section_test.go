// Package analysis_test contains tests for the CPL analyzer units multi-section functionality.
//
// This file tests that the analyzer correctly handles multiple units sections,
// including duplicate detection across sections and symbol table walk behavior.
//
// The key behaviors verified:
//   - Duplicate unit definitions across multiple sections are detected as errors
//   - Valid multi-section scenarios (different units in each section) work correctly
//   - The symbol table walk visits all UnitScope tables regardless of section
//
// Reference: Story163 - CPL Units and Measures Multiple Sections
package analysis_test

import (
	"fmt"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Cross-Section Duplicate Detection Tests
// ============================================================================

// TestAnalyzer_DuplicateUnitsAcrossSections verifies that duplicate unit
// definitions across multiple units sections are detected as errors during analysis.
//
// This test ensures that the analyzer's single units map correctly catches
// duplicates even when they originate from different UnitsScope tables.
func TestAnalyzer_DuplicateUnitsAcrossSections(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		units         []MockUnitDeclaration
		expectError   bool
		errorContains string
	}{
		{
			name: "same unit name in two sections",
			units: []MockUnitDeclaration{
				{UnitName: "Currency", ConceptName: "Country"},
				{UnitName: "Currency", ConceptName: "Mass"},
			},
			expectError:   true,
			errorContains: "Currency",
		},
		{
			name: "duplicate Weight unit across sections",
			units: []MockUnitDeclaration{
				{UnitName: "Weight", ConceptName: "Mass"},
				{UnitName: "Length", ConceptName: "Distance"},
				{UnitName: "Weight", ConceptName: "Force"},
			},
			expectError:   true,
			errorContains: "Weight",
		},
		{
			name: "duplicate in second and third section",
			units: []MockUnitDeclaration{
				{UnitName: "Currency", ConceptName: "Country"},
				{UnitName: "Length", ConceptName: "Distance"},
				{UnitName: "Length", ConceptName: "Time"},
			},
			expectError:   true,
			errorContains: "Length",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Create mock symbol table with units across multiple sections
			rootTable := createMockSymbolTableWithMultipleUnitsSections(tc.units)

			// Act: Analyze the symbol table
			analyzer := analysis.NewCPLAnalyzer()
			enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
				symboltable.UnitScope: analysis.AnalyzeUnitScope,
			}
			err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, nil, analyzer)

			// Assert
			if tc.expectError {
				require.Error(t, err, "Analysis should fail for duplicate units across sections")
				assert.Contains(t, err.Error(), "duplicate unit definition", "Error should indicate duplicate")
				assert.Contains(t, err.Error(), tc.errorContains, "Error should mention the unit name")
			} else {
				require.NoError(t, err, "Analysis should succeed")
			}
		})
	}
}

// TestAnalyzer_ValidMultiSectionUnits verifies that valid multi-section
// scenarios work correctly - different units can be defined in separate
// units sections without errors.
func TestAnalyzer_ValidMultiSectionUnits(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		section1Units []MockUnitDeclaration
		section2Units []MockUnitDeclaration
		expectedUnits map[string]string // unit name -> expected concept name
	}{
		{
			name: "different units in two sections",
			section1Units: []MockUnitDeclaration{
				{UnitName: "Currency", ConceptName: "Country"},
			},
			section2Units: []MockUnitDeclaration{
				{UnitName: "Weight", ConceptName: "Mass"},
			},
			expectedUnits: map[string]string{
				"Currency": "Country",
				"Weight":   "Mass",
			},
		},
		{
			name: "multiple units in each section",
			section1Units: []MockUnitDeclaration{
				{UnitName: "Currency", ConceptName: "Country"},
				{UnitName: "Length", ConceptName: "Distance"},
			},
			section2Units: []MockUnitDeclaration{
				{UnitName: "Weight", ConceptName: "Mass"},
				{UnitName: "Time", ConceptName: "Duration"},
			},
			expectedUnits: map[string]string{
				"Currency": "Country",
				"Length":   "Distance",
				"Weight":   "Mass",
				"Time":     "Duration",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Create mock symbol table with multiple sections
			rootTable := createMockSymbolTableWithTwoUnitsSections(
				tc.section1Units,
				tc.section2Units,
			)

			// Act: Analyze the symbol table
			analyzer := analysis.NewCPLAnalyzer()
			enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
				symboltable.UnitScope: analysis.AnalyzeUnitScope,
			}
			err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, nil, analyzer)

			// Assert: Analysis should succeed
			require.NoError(t, err, "Analysis should succeed for valid multi-section units")

			// Verify all expected units exist with correct concept names
			defs := analyzer.GetAnalysisDefinitions()
			require.NotNil(t, defs, "Analysis definitions should not be nil")
			require.NotNil(t, defs.Units, "Units map should be initialized")

			for unitName, expectedConceptName := range tc.expectedUnits {
				actualUnit, exists := defs.Units[unitName]
				require.True(t, exists, "Unit '%s' should exist in analysis results", unitName)
				assert.Equal(t, unitName, actualUnit.Name, "Unit name should match")
				assert.Equal(t, expectedConceptName, actualUnit.ConceptName, "Concept name should match for unit '%s'", unitName)
			}

			// Verify total unit count
			assert.Len(t, defs.Units, len(tc.expectedUnits),
				"Should have exactly %d units", len(tc.expectedUnits))
		})
	}
}

// TestAnalyzer_ThreeSectionsWithOneDuplicate verifies that when three units
// sections exist and one unit is duplicated, the analysis correctly detects
// the duplicate.
func TestAnalyzer_ThreeSectionsWithOneDuplicate(t *testing.T) {
	t.Parallel()

	// Arrange: Create three units sections with one duplicate
	rootTable := createMockSymbolTableWithNUnitsSections([][]MockUnitDeclaration{
		{{UnitName: "Currency", ConceptName: "Country"}},
		{{UnitName: "Weight", ConceptName: "Mass"}},
		{{UnitName: "Currency", ConceptName: "Force"}}, // Duplicate!
	})

	// Act: Analyze the symbol table
	analyzer := analysis.NewCPLAnalyzer()
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.UnitScope: analysis.AnalyzeUnitScope,
	}
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, nil, analyzer)

	// Assert: Should fail with duplicate error
	require.Error(t, err, "Analysis should fail for duplicate unit across three sections")
	assert.Contains(t, err.Error(), "duplicate unit definition", "Error should indicate duplicate")
	assert.Contains(t, err.Error(), "Currency", "Error should mention the duplicated unit name")
}

// TestAnalyzer_WalkSymbolTableVisitsAllUnitScopes verifies that the
// WalkSymbolTableWithHandlers function visits all UnitScope tables,
// regardless of how many units sections exist in the document.
func TestAnalyzer_WalkSymbolTableVisitsAllUnitScopes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		numUnitsSections   int
		unitsPerSection    int
		expectedTotalUnits int
	}{
		{
			name:               "two sections with two units each",
			numUnitsSections:   2,
			unitsPerSection:    2,
			expectedTotalUnits: 4,
		},
		{
			name:               "three sections with varying units",
			numUnitsSections:   3,
			unitsPerSection:    1,
			expectedTotalUnits: 3,
		},
		{
			name:               "single section",
			numUnitsSections:   1,
			unitsPerSection:    3,
			expectedTotalUnits: 3,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Create mock symbol table with multiple sections
			rootTable := createMockSymbolTableWithNUniformUnitsSections(
				tc.numUnitsSections,
				tc.unitsPerSection,
			)

			// Track visited scopes
			visitedUnitScopes := 0

			trackingEnterHandler := map[symboltable.ScopeType]analysis.AnalysisFunc{
				symboltable.UnitScope: func(table *symboltable.SymbolTable, context *analysis.AnalysisContext, analyzer *analysis.CPLAnalyzer) error {
					visitedUnitScopes++
					return analysis.AnalyzeUnitScope(table, context, analyzer)
				},
			}

			analyzer := analysis.NewCPLAnalyzer()

			// Act: Walk the symbol table
			err := analysis.WalkSymbolTableWithHandlers(rootTable, trackingEnterHandler, nil, analyzer)

			// Assert
			require.NoError(t, err, "Walk should succeed")
			assert.Equal(t, tc.expectedTotalUnits, visitedUnitScopes,
				"Should visit all %d UnitScope tables", tc.expectedTotalUnits)

			// Verify all units were analyzed
			defs := analyzer.GetAnalysisDefinitions()
			assert.Len(t, defs.Units, tc.expectedTotalUnits,
				"Should have %d units in analysis results", tc.expectedTotalUnits)
		})
	}
}

// ============================================================================
// Helper Types and Functions
// ============================================================================

// MockUnitDeclaration represents a unit declaration for testing.
type MockUnitDeclaration struct {
	UnitName    string
	ConceptName string
}

// createMockSymbolTableWithMultipleUnitsSections creates a mock symbol table
// with multiple UnitsScope tables, each containing one unit declaration.
// Each unit in the slice gets its own UnitsScope for cross-section testing.
func createMockSymbolTableWithMultipleUnitsSections(units []MockUnitDeclaration) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()

	for _, unit := range units {
		unitsScope := root.AddChild(symboltable.UnitsScope)
		unitScope := unitsScope.AddChild(symboltable.UnitScope)

		unitName := unit.UnitName
		conceptName := unit.ConceptName

		_ = unitScope.Define(&unitName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
		_ = unitScope.Define(&conceptName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 2})
	}

	return root
}

// createMockSymbolTableWithTwoUnitsSections creates a mock symbol table with
// exactly two UnitsScope tables, each containing the specified units.
func createMockSymbolTableWithTwoUnitsSections(
	section1Units []MockUnitDeclaration,
	section2Units []MockUnitDeclaration,
) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()

	// First section
	unitsScope1 := root.AddChild(symboltable.UnitsScope)
	for _, unit := range section1Units {
		unitScope := unitsScope1.AddChild(symboltable.UnitScope)
		unitName := unit.UnitName
		conceptName := unit.ConceptName
		_ = unitScope.Define(&unitName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
		_ = unitScope.Define(&conceptName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 2})
	}

	// Second section
	unitsScope2 := root.AddChild(symboltable.UnitsScope)
	for _, unit := range section2Units {
		unitScope := unitsScope2.AddChild(symboltable.UnitScope)
		unitName := unit.UnitName
		conceptName := unit.ConceptName
		_ = unitScope.Define(&unitName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
		_ = unitScope.Define(&conceptName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 2})
	}

	return root
}

// createMockSymbolTableWithNUnitsSections creates a mock symbol table with
// N UnitsScope tables, each containing the units from the provided sections.
func createMockSymbolTableWithNUnitsSections(sections [][]MockUnitDeclaration) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()

	for _, sectionUnits := range sections {
		unitsScope := root.AddChild(symboltable.UnitsScope)
		for _, unit := range sectionUnits {
			unitScope := unitsScope.AddChild(symboltable.UnitScope)
			unitName := unit.UnitName
			conceptName := unit.ConceptName
			_ = unitScope.Define(&unitName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
			_ = unitScope.Define(&conceptName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 2})
		}
	}

	return root
}

// createMockSymbolTableWithNUniformUnitsSections creates a mock symbol table with
// N UnitsScope tables, each containing M units.
func createMockSymbolTableWithNUniformUnitsSections(numSections, unitsPerSection int) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()

	unitCounter := 0
	for s := 0; s < numSections; s++ {
		unitsScope := root.AddChild(symboltable.UnitsScope)

		for u := 0; u < unitsPerSection; u++ {
			unitCounter++
			unitName := fmt.Sprintf("Unit%d", unitCounter)
			conceptName := fmt.Sprintf("Concept%d", unitCounter)

			unitScope := unitsScope.AddChild(symboltable.UnitScope)
			_ = unitScope.Define(&unitName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
			_ = unitScope.Define(&conceptName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 2})
		}
	}

	return root
}
