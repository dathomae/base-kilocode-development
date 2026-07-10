// Package analysis_test contains tests for the CPL analyzer measures multi-section functionality.
//
// This file tests that the analyzer correctly handles multiple measures sections,
// including duplicate detection across sections and symbol table walk behavior.
//
// The key behaviors verified:
//   - Duplicate measure definitions across multiple sections are detected as errors
//   - Valid multi-section scenarios (different measures in each section) work correctly
//   - The symbol table walk visits all MeasureScope tables regardless of section
//
// Reference: Story163 - CPL Units and Measures Multiple Sections
package analysis_test

import (
	"fmt"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Cross-Section Duplicate Detection Tests
// ============================================================================

// TestAnalyzer_DuplicateMeasuresAcrossSections verifies that duplicate measure
// definitions across multiple measures sections are detected as errors during analysis.
//
// This test ensures that the analyzer's single measures map correctly catches
// duplicates even when they originate from different MeasuresScope tables.
func TestAnalyzer_DuplicateMeasuresAcrossSections(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		measures      []MockMeasureDeclaration
		expectError   bool
		errorContains string
	}{
		{
			name: "same measure name in two sections",
			measures: []MockMeasureDeclaration{
				{MeasureName: "Distance", TypeName: "Length", UnitName: "Meter"},
				{MeasureName: "Distance", TypeName: "Mass", UnitName: "Kilogram"},
			},
			expectError:   true,
			errorContains: "Distance",
		},
		{
			name: "duplicate Weight measure across sections",
			measures: []MockMeasureDeclaration{
				{MeasureName: "Weight", TypeName: "Mass", UnitName: "Kilogram"},
				{MeasureName: "Price", TypeName: "Currency", UnitName: "Dollar"},
				{MeasureName: "Weight", TypeName: "Force", UnitName: "Newton"},
			},
			expectError:   true,
			errorContains: "Weight",
		},
		{
			name: "duplicate in second and third section",
			measures: []MockMeasureDeclaration{
				{MeasureName: "Distance", TypeName: "Length", UnitName: "Meter"},
				{MeasureName: "Time", TypeName: "Duration", UnitName: "Second"},
				{MeasureName: "Time", TypeName: "Instant", UnitName: "Epoch"},
			},
			expectError:   true,
			errorContains: "Time",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Create mock symbol table with measures across multiple sections
			rootTable := createMockSymbolTableWithMultipleMeasuresSections(tc.measures)

			// Act: Analyze the symbol table
			analyzer := analysis.NewCPLAnalyzer()
			enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
				symboltable.MeasureScope: analysis.AnalyzeMeasureScope,
			}
			err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, nil, analyzer)

			// Assert
			if tc.expectError {
				require.Error(t, err, "Analysis should fail for duplicate measures across sections")
				assert.Contains(t, err.Error(), "duplicate measure definition", "Error should indicate duplicate")
				assert.Contains(t, err.Error(), tc.errorContains, "Error should mention the measure name")
			} else {
				require.NoError(t, err, "Analysis should succeed")
			}
		})
	}
}

// TestAnalyzer_ValidMultiSectionMeasures verifies that valid multi-section
// scenarios work correctly - different measures can be defined in separate
// measures sections without errors.
func TestAnalyzer_ValidMultiSectionMeasures(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		section1Measures []MockMeasureDeclaration
		section2Measures []MockMeasureDeclaration
		expectedMeasures map[string]MockMeasureDeclaration // measure name -> expected declaration
	}{
		{
			name: "different measures in two sections",
			section1Measures: []MockMeasureDeclaration{
				{MeasureName: "Distance", TypeName: "Length", UnitName: "Meter"},
			},
			section2Measures: []MockMeasureDeclaration{
				{MeasureName: "Weight", TypeName: "Mass", UnitName: "Kilogram"},
			},
			expectedMeasures: map[string]MockMeasureDeclaration{
				"Distance": {MeasureName: "Distance", TypeName: "Length", UnitName: "Meter"},
				"Weight":   {MeasureName: "Weight", TypeName: "Mass", UnitName: "Kilogram"},
			},
		},
		{
			name: "multiple measures in each section",
			section1Measures: []MockMeasureDeclaration{
				{MeasureName: "Distance", TypeName: "Length", UnitName: "Meter"},
				{MeasureName: "Time", TypeName: "Duration", UnitName: "Second"},
			},
			section2Measures: []MockMeasureDeclaration{
				{MeasureName: "Weight", TypeName: "Mass", UnitName: "Kilogram"},
				{MeasureName: "Price", TypeName: "Currency", UnitName: "Dollar"},
			},
			expectedMeasures: map[string]MockMeasureDeclaration{
				"Distance": {MeasureName: "Distance", TypeName: "Length", UnitName: "Meter"},
				"Time":     {MeasureName: "Time", TypeName: "Duration", UnitName: "Second"},
				"Weight":   {MeasureName: "Weight", TypeName: "Mass", UnitName: "Kilogram"},
				"Price":    {MeasureName: "Price", TypeName: "Currency", UnitName: "Dollar"},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Create mock symbol table with multiple sections
			rootTable := createMockSymbolTableWithTwoMeasuresSections(
				tc.section1Measures,
				tc.section2Measures,
			)

			// Act: Analyze the symbol table
			analyzer := analysis.NewCPLAnalyzer()
			enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
				symboltable.MeasureScope: analysis.AnalyzeMeasureScope,
			}
			err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, nil, analyzer)

			// Assert: Analysis should succeed
			require.NoError(t, err, "Analysis should succeed for valid multi-section measures")

			// Verify all expected measures exist with correct type and unit names
			defs := analyzer.GetAnalysisDefinitions()
			require.NotNil(t, defs, "Analysis definitions should not be nil")
			require.NotNil(t, defs.Measures, "Measures map should be initialized")

			for measureName, expected := range tc.expectedMeasures {
				actualMeasure, exists := defs.Measures[measureName]
				require.True(t, exists, "Measure '%s' should exist in analysis results", measureName)
				assert.Equal(t, expected.MeasureName, actualMeasure.Name, "Measure name should match")
				assert.Equal(t, expected.TypeName, actualMeasure.TypeName, "Type name should match for measure '%s'", measureName)
				assert.Equal(t, expected.UnitName, actualMeasure.GetUnitName(), "Unit name should match for measure '%s'", measureName)
			}

			// Verify total measure count
			assert.Len(t, defs.Measures, len(tc.expectedMeasures),
				"Should have exactly %d measures", len(tc.expectedMeasures))
		})
	}
}

// TestAnalyzer_ThreeMeasuresSectionsWithOneDuplicate verifies that when three measures
// sections exist and one measure is duplicated, the analysis correctly detects
// the duplicate.
func TestAnalyzer_ThreeMeasuresSectionsWithOneDuplicate(t *testing.T) {
	t.Parallel()

	// Arrange: Create three measures sections with one duplicate
	rootTable := createMockSymbolTableWithNMeasuresSections([][]MockMeasureDeclaration{
		{{MeasureName: "Distance", TypeName: "Length", UnitName: "Meter"}},
		{{MeasureName: "Weight", TypeName: "Mass", UnitName: "Kilogram"}},
		{{MeasureName: "Distance", TypeName: "Time", UnitName: "Second"}}, // Duplicate!
	})

	// Act: Analyze the symbol table
	analyzer := analysis.NewCPLAnalyzer()
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MeasureScope: analysis.AnalyzeMeasureScope,
	}
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, nil, analyzer)

	// Assert: Should fail with duplicate error
	require.Error(t, err, "Analysis should fail for duplicate measure across three sections")
	assert.Contains(t, err.Error(), "duplicate measure definition", "Error should indicate duplicate")
	assert.Contains(t, err.Error(), "Distance", "Error should mention the duplicated measure name")
}

// TestAnalyzer_WalkSymbolTableVisitsAllMeasureScopes verifies that the
// WalkSymbolTableWithHandlers function visits all MeasureScope tables,
// regardless of how many measures sections exist in the document.
func TestAnalyzer_WalkSymbolTableVisitsAllMeasureScopes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                  string
		numMeasuresSections   int
		measuresPerSection    int
		expectedTotalMeasures int
	}{
		{
			name:                  "two sections with two measures each",
			numMeasuresSections:   2,
			measuresPerSection:    2,
			expectedTotalMeasures: 4,
		},
		{
			name:                  "three sections with varying measures",
			numMeasuresSections:   3,
			measuresPerSection:    1,
			expectedTotalMeasures: 3,
		},
		{
			name:                  "single section",
			numMeasuresSections:   1,
			measuresPerSection:    3,
			expectedTotalMeasures: 3,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Create mock symbol table with multiple sections
			rootTable := createMockSymbolTableWithNUniformMeasuresSections(
				tc.numMeasuresSections,
				tc.measuresPerSection,
			)

			// Track visited scopes
			visitedMeasureScopes := 0

			trackingEnterHandler := map[symboltable.ScopeType]analysis.AnalysisFunc{
				symboltable.MeasureScope: func(table *symboltable.SymbolTable, context *analysis.AnalysisContext, analyzer *analysis.CPLAnalyzer) error {
					visitedMeasureScopes++
					return analysis.AnalyzeMeasureScope(table, context, analyzer)
				},
			}

			analyzer := analysis.NewCPLAnalyzer()

			// Act: Walk the symbol table
			err := analysis.WalkSymbolTableWithHandlers(rootTable, trackingEnterHandler, nil, analyzer)

			// Assert
			require.NoError(t, err, "Walk should succeed")
			assert.Equal(t, tc.expectedTotalMeasures, visitedMeasureScopes,
				"Should visit all %d MeasureScope tables", tc.expectedTotalMeasures)

			// Verify all measures were analyzed
			defs := analyzer.GetAnalysisDefinitions()
			assert.Len(t, defs.Measures, tc.expectedTotalMeasures,
				"Should have %d measures in analysis results", tc.expectedTotalMeasures)
		})
	}
}

// ============================================================================
// Helper Types and Functions
// ============================================================================

// MockMeasureDeclaration represents a measure declaration for testing.
type MockMeasureDeclaration struct {
	MeasureName string
	TypeName    string
	UnitName    string
}

// createMockSymbolTableWithMultipleMeasuresSections creates a mock symbol table
// with multiple MeasuresScope tables, each containing one measure declaration.
// Each measure in the slice gets its own MeasuresScope for cross-section testing.
func createMockSymbolTableWithMultipleMeasuresSections(measures []MockMeasureDeclaration) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()

	for _, measure := range measures {
		measuresScope := root.AddChild(symboltable.MeasuresScope)
		measureScope := measuresScope.AddChild(symboltable.MeasureScope)

		measureName := measure.MeasureName
		typeName := measure.TypeName
		unitName := measure.UnitName

		_ = measureScope.Define(&measureName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
		_ = measureScope.Define(&typeName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 2})

		// Push unit reference as postfix definition
		_ = measureScope.DefineWithData(nil, symboltable.UnitReferenceDefinition, symboltable.SourceLocation{Line: 1, Column: 4}, parser.ParserUnitReference{Name: unitName})
	}

	return root
}

// createMockSymbolTableWithTwoMeasuresSections creates a mock symbol table with
// exactly two MeasuresScope tables, each containing the specified measures.
func createMockSymbolTableWithTwoMeasuresSections(
	section1Measures []MockMeasureDeclaration,
	section2Measures []MockMeasureDeclaration,
) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()

	// First section
	measuresScope1 := root.AddChild(symboltable.MeasuresScope)
	for _, measure := range section1Measures {
		measureScope := measuresScope1.AddChild(symboltable.MeasureScope)
		measureName := measure.MeasureName
		typeName := measure.TypeName
		unitName := measure.UnitName
		_ = measureScope.Define(&measureName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
		_ = measureScope.Define(&typeName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 2})

		_ = measureScope.DefineWithData(nil, symboltable.UnitReferenceDefinition, symboltable.SourceLocation{Line: 1, Column: 4}, parser.ParserUnitReference{Name: unitName})
	}

	// Second section
	measuresScope2 := root.AddChild(symboltable.MeasuresScope)
	for _, measure := range section2Measures {
		measureScope := measuresScope2.AddChild(symboltable.MeasureScope)
		measureName := measure.MeasureName
		typeName := measure.TypeName
		unitName := measure.UnitName
		_ = measureScope.Define(&measureName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
		_ = measureScope.Define(&typeName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 2})

		_ = measureScope.DefineWithData(nil, symboltable.UnitReferenceDefinition, symboltable.SourceLocation{Line: 1, Column: 4}, parser.ParserUnitReference{Name: unitName})
	}

	return root
}

// createMockSymbolTableWithNMeasuresSections creates a mock symbol table with
// N MeasuresScope tables, each containing the measures from the provided sections.
func createMockSymbolTableWithNMeasuresSections(sections [][]MockMeasureDeclaration) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()

	for _, sectionMeasures := range sections {
		measuresScope := root.AddChild(symboltable.MeasuresScope)
		for _, measure := range sectionMeasures {
			measureScope := measuresScope.AddChild(symboltable.MeasureScope)
			measureName := measure.MeasureName
			typeName := measure.TypeName
			unitName := measure.UnitName
			_ = measureScope.Define(&measureName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
			_ = measureScope.Define(&typeName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 2})

			_ = measureScope.DefineWithData(nil, symboltable.UnitReferenceDefinition, symboltable.SourceLocation{Line: 1, Column: 4}, parser.ParserUnitReference{Name: unitName})
		}
	}

	return root
}

// createMockSymbolTableWithNUniformMeasuresSections creates a mock symbol table with
// N MeasuresScope tables, each containing M measures.
func createMockSymbolTableWithNUniformMeasuresSections(numSections, measuresPerSection int) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()

	measureCounter := 0
	for s := 0; s < numSections; s++ {
		measuresScope := root.AddChild(symboltable.MeasuresScope)

		for m := 0; m < measuresPerSection; m++ {
			measureCounter++
			measureName := fmt.Sprintf("Measure%d", measureCounter)
			typeName := fmt.Sprintf("Type%d", measureCounter)
			unitName := fmt.Sprintf("Unit%d", measureCounter)

			measureScope := measuresScope.AddChild(symboltable.MeasureScope)
			_ = measureScope.Define(&measureName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
			_ = measureScope.Define(&typeName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 2})

			_ = measureScope.DefineWithData(nil, symboltable.UnitReferenceDefinition, symboltable.SourceLocation{Line: 1, Column: 4}, parser.ParserUnitReference{Name: unitName})
		}
	}

	return root
}
