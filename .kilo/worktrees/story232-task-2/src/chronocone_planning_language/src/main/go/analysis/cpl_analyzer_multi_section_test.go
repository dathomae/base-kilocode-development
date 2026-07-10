// Package analysis_test contains tests for the CPL analyzer multi-section functionality.
//
// This file tests that the analyzer correctly handles multiple concepts sections,
// including duplicate detection across sections and hierarchy preservation.
//
// The key behaviors verified:
//   - Duplicate concept definitions across multiple sections are detected as errors
//   - Valid multi-section scenarios (different concepts in each section) work correctly
//   - Hierarchy is preserved when concepts span multiple sections
//   - The symbol table walk visits all ConceptsScope tables regardless of section
//
// Reference: Story120 - CPL Concepts Analyzer Multi-Section
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

// TestAnalyzer_DuplicateConceptsAcrossSections verifies that duplicate concept
// definitions across multiple concepts sections are detected as errors during analysis.
//
// This test ensures that the analyzer's single concepts map correctly catches
// duplicates even when they originate from different ConceptsScope tables.
func TestAnalyzer_DuplicateConceptsAcrossSections(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		concepts      []MockConceptDeclaration
		expectError   bool
		errorContains string
	}{
		{
			name: "same concept name in two sections",
			concepts: []MockConceptDeclaration{
				{ConceptName: "Status", Values: []string{"Active", "Inactive"}},
				{ConceptName: "Status", Values: []string{"Pending", "Completed"}},
			},
			expectError:   true,
			errorContains: "Status",
		},
		{
			name: "duplicate Country concept across sections",
			concepts: []MockConceptDeclaration{
				{ConceptName: "Country", Values: []string{"USA", "Canada"}},
				{ConceptName: "Priority", Values: []string{"Low", "High"}},
				{ConceptName: "Country", Values: []string{"Mexico"}},
			},
			expectError:   true,
			errorContains: "Country",
		},
		{
			name: "duplicate in second and third section",
			concepts: []MockConceptDeclaration{
				{ConceptName: "Type", Values: []string{"A", "B"}},
				{ConceptName: "Category", Values: []string{"X", "Y"}},
				{ConceptName: "Type", Values: []string{"C"}},
			},
			expectError:   true,
			errorContains: "Type",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Create mock symbol table with concepts across multiple sections
			rootTable := createMockSymbolTableWithMultipleConceptsSections(tc.concepts)

			// Act: Analyze the symbol table
			analyzer := analysis.NewCPLAnalyzer()
			enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
				symboltable.ConceptsScope:       analysis.EnterAnalyzeConceptsScope,
				symboltable.ConceptScope:        analysis.EnterAnalyzeConceptScope,
				symboltable.ValueGeneratorScope: analysis.AnalyzeValueGeneratorScope,
			}
			exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
				symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
			}
			err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

			// Assert
			if tc.expectError {
				require.Error(t, err, "Analysis should fail for duplicate concepts across sections")
				assert.Contains(t, err.Error(), "duplicate concept definition", "Error should indicate duplicate")
				assert.Contains(t, err.Error(), tc.errorContains, "Error should mention the concept name")
				assert.Contains(t, err.Error(), "first at", "Error should include first definition location")
				assert.Contains(t, err.Error(), "second at", "Error should include second definition location")
			} else {
				require.NoError(t, err, "Analysis should succeed")
			}
		})
	}
}

// TestAnalyzer_ValidMultiSectionConcepts verifies that valid multi-section
// scenarios work correctly - different concepts can be defined in separate
// concepts sections without errors.
func TestAnalyzer_ValidMultiSectionConcepts(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		section1Concepts []MockConceptDeclaration
		section2Concepts []MockConceptDeclaration
		expectedConcepts map[string][]string // concept name -> expected values
	}{
		{
			name: "different concepts in two sections",
			section1Concepts: []MockConceptDeclaration{
				{ConceptName: "Country", Values: []string{"USA", "Canada", "Mexico"}},
			},
			section2Concepts: []MockConceptDeclaration{
				{ConceptName: "Priority", Values: []string{"Low", "Medium", "High"}},
			},
			expectedConcepts: map[string][]string{
				"Country":  {"USA", "Canada", "Mexico"},
				"Priority": {"Low", "Medium", "High"},
			},
		},
		{
			name: "multiple concepts in each section",
			section1Concepts: []MockConceptDeclaration{
				{ConceptName: "Status", Values: []string{"Active", "Inactive"}},
				{ConceptName: "Type", Values: []string{"A", "B", "C"}},
			},
			section2Concepts: []MockConceptDeclaration{
				{ConceptName: "Category", Values: []string{"X", "Y"}},
				{ConceptName: "Region", Values: []string{"North", "South", "East", "West"}},
			},
			expectedConcepts: map[string][]string{
				"Status":   {"Active", "Inactive"},
				"Type":     {"A", "B", "C"},
				"Category": {"X", "Y"},
				"Region":   {"North", "South", "East", "West"},
			},
		},
		{
			name: "three sections with unique concepts",
			section1Concepts: []MockConceptDeclaration{
				{ConceptName: "Color", Values: []string{"Red", "Blue"}},
			},
			section2Concepts: []MockConceptDeclaration{
				{ConceptName: "Size", Values: []string{"Small", "Large"}},
			},
			expectedConcepts: map[string][]string{
				"Color": {"Red", "Blue"},
				"Size":  {"Small", "Large"},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Create mock symbol table with multiple sections
			rootTable := createMockSymbolTableWithTwoConceptsSections(
				tc.section1Concepts,
				tc.section2Concepts,
			)

			// Act: Analyze the symbol table
			analyzer := analysis.NewCPLAnalyzer()
			enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
				symboltable.ConceptsScope:       analysis.EnterAnalyzeConceptsScope,
				symboltable.ConceptScope:        analysis.EnterAnalyzeConceptScope,
				symboltable.ValueGeneratorScope: analysis.AnalyzeValueGeneratorScope,
			}
			exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
				symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
			}
			err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

			// Assert: Analysis should succeed
			require.NoError(t, err, "Analysis should succeed for valid multi-section concepts")

			// Verify all expected concepts exist with correct values
			defs := analyzer.GetAnalysisDefinitions()
			require.NotNil(t, defs, "Analysis definitions should not be nil")
			require.NotNil(t, defs.Concepts, "Concepts map should be initialized")

			for conceptName, expectedValues := range tc.expectedConcepts {
				actualConcept, exists := defs.Concepts[conceptName]
				require.True(t, exists, "Concept '%s' should exist in analysis results", conceptName)
				assert.Equal(t, conceptName, actualConcept.ConceptName, "Concept name should match")

				// Verify values
				require.Len(t, actualConcept.ResolvedValues, len(expectedValues),
					"Concept '%s' should have %d values", conceptName, len(expectedValues))

				for i, expectedValue := range expectedValues {
					assert.Equal(t, expectedValue, actualConcept.ResolvedValues[i].Name,
						"Value at position %d for concept '%s' should match", i, conceptName)
					assert.Equal(t, i, actualConcept.ResolvedValues[i].Position,
						"Position for value '%s' in concept '%s' should be %d", expectedValue, conceptName, i)
				}
			}

			// Verify total concept count
			assert.Len(t, defs.Concepts, len(tc.expectedConcepts),
				"Should have exactly %d concepts", len(tc.expectedConcepts))
		})
	}
}

// TestAnalyzer_MultiSectionHierarchyPreservation verifies that parent-child
// hierarchy is correctly preserved when concepts span multiple sections.
//
// Each section maintains its own internal hierarchy, and the analyzer correctly
// processes all of them during the symbol table walk.
func TestAnalyzer_MultiSectionHierarchyPreservation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		section1Hierarchy map[string]MockHierarchyConcept
		section2Hierarchy map[string]MockHierarchyConcept
		expectedParents   map[string]*string // concept name -> expected parent name (nil for root)
	}{
		{
			name: "hierarchy in first section only",
			section1Hierarchy: map[string]MockHierarchyConcept{
				"Country": {Name: "Country", Values: []string{"USA", "Canada"}, Parent: nil, Indent: 0},
				"Region":  {Name: "Region", Values: []string{"North", "South"}, Parent: stringPtr("Country"), Indent: 4},
			},
			section2Hierarchy: map[string]MockHierarchyConcept{
				"Priority": {Name: "Priority", Values: []string{"Low", "High"}, Parent: nil, Indent: 0},
			},
			expectedParents: map[string]*string{
				"Country":  nil,
				"Region":   stringPtr("Country"),
				"Priority": nil,
			},
		},
		{
			name: "hierarchy in both sections",
			section1Hierarchy: map[string]MockHierarchyConcept{
				"Product": {Name: "Product", Values: []string{"Widget", "Gadget"}, Parent: nil, Indent: 0},
				"Variant": {Name: "Variant", Values: []string{"A", "B"}, Parent: stringPtr("Product"), Indent: 4},
			},
			section2Hierarchy: map[string]MockHierarchyConcept{
				"Department": {Name: "Department", Values: []string{"Sales", "Engineering"}, Parent: nil, Indent: 0},
				"Team":       {Name: "Team", Values: []string{"Alpha", "Beta"}, Parent: stringPtr("Department"), Indent: 4},
			},
			expectedParents: map[string]*string{
				"Product":    nil,
				"Variant":    stringPtr("Product"),
				"Department": nil,
				"Team":       stringPtr("Department"),
			},
		},
		{
			name: "deep hierarchy in one section, flat in another",
			section1Hierarchy: map[string]MockHierarchyConcept{
				"A": {Name: "A", Values: []string{"A1"}, Parent: nil, Indent: 0},
				"B": {Name: "B", Values: []string{"B1"}, Parent: stringPtr("A"), Indent: 4},
				"C": {Name: "C", Values: []string{"C1"}, Parent: stringPtr("B"), Indent: 8},
			},
			section2Hierarchy: map[string]MockHierarchyConcept{
				"X": {Name: "X", Values: []string{"X1", "X2"}, Parent: nil, Indent: 0},
				"Y": {Name: "Y", Values: []string{"Y1", "Y2"}, Parent: nil, Indent: 0},
			},
			expectedParents: map[string]*string{
				"A": nil,
				"B": stringPtr("A"),
				"C": stringPtr("B"),
				"X": nil,
				"Y": nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Create mock symbol table with hierarchical concepts
			rootTable := createMockSymbolTableWithHierarchicalSections(
				tc.section1Hierarchy,
				tc.section2Hierarchy,
			)

			// Act: Analyze the symbol table
			analyzer := analysis.NewCPLAnalyzer()
			enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
				symboltable.ConceptsScope:       analysis.EnterAnalyzeConceptsScope,
				symboltable.ConceptScope:        analysis.EnterAnalyzeConceptScope,
				symboltable.ValueGeneratorScope: analysis.AnalyzeValueGeneratorScope,
			}
			exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
				symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
			}
			err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

			// Assert: Analysis should succeed
			require.NoError(t, err, "Analysis should succeed for hierarchical multi-section concepts")

			// Verify parent relationships
			defs := analyzer.GetAnalysisDefinitions()
			require.NotNil(t, defs, "Analysis definitions should not be nil")

			for conceptName, expectedParent := range tc.expectedParents {
				actualConcept, exists := defs.Concepts[conceptName]
				require.True(t, exists, "Concept '%s' should exist", conceptName)

				if expectedParent == nil {
					assert.Nil(t, actualConcept.ParentConceptName,
						"Concept '%s' should have no parent", conceptName)
				} else {
					require.NotNil(t, actualConcept.ParentConceptName,
						"Concept '%s' should have a parent", conceptName)
					assert.Equal(t, *expectedParent, *actualConcept.ParentConceptName,
						"Concept '%s' should have parent '%s'", conceptName, *expectedParent)
				}
			}
		})
	}
}

// TestAnalyzer_WalkSymbolTableVisitsAllConceptsScopes verifies that the
// WalkSymbolTableWithHandlers function visits all ConceptsScope tables,
// regardless of how many sections exist in the document.
func TestAnalyzer_WalkSymbolTableVisitsAllConceptsScopes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                  string
		numConceptsSections   int
		conceptsPerSection    int
		expectedTotalConcepts int
	}{
		{
			name:                  "two sections with two concepts each",
			numConceptsSections:   2,
			conceptsPerSection:    2,
			expectedTotalConcepts: 4,
		},
		{
			name:                  "three sections with varying concepts",
			numConceptsSections:   3,
			conceptsPerSection:    1,
			expectedTotalConcepts: 3,
		},
		{
			name:                  "single section",
			numConceptsSections:   1,
			conceptsPerSection:    3,
			expectedTotalConcepts: 3,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Create mock symbol table with multiple sections
			rootTable := createMockSymbolTableWithNConceptsSections(
				tc.numConceptsSections,
				tc.conceptsPerSection,
			)

			// Track visited scopes
			visitedConceptsScopes := 0
			visitedConceptScopes := 0

			trackingEnterHandler := map[symboltable.ScopeType]analysis.AnalysisFunc{
				symboltable.ConceptsScope: func(table *symboltable.SymbolTable, context *analysis.AnalysisContext, analyzer *analysis.CPLAnalyzer) error {
					visitedConceptsScopes++
					return analysis.EnterAnalyzeConceptsScope(table, context, analyzer)
				},
				symboltable.ConceptScope: func(table *symboltable.SymbolTable, context *analysis.AnalysisContext, analyzer *analysis.CPLAnalyzer) error {
					visitedConceptScopes++
					return analysis.EnterAnalyzeConceptScope(table, context, analyzer)
				},
				symboltable.ValueGeneratorScope: analysis.AnalyzeValueGeneratorScope,
			}
			exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
				symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
			}

			analyzer := analysis.NewCPLAnalyzer()

			// Act: Walk the symbol table
			err := analysis.WalkSymbolTableWithHandlers(rootTable, trackingEnterHandler, exitHandlers, analyzer)

			// Assert
			require.NoError(t, err, "Walk should succeed")
			assert.Equal(t, tc.numConceptsSections, visitedConceptsScopes,
				"Should visit all %d ConceptsScope tables", tc.numConceptsSections)
			assert.Equal(t, tc.expectedTotalConcepts, visitedConceptScopes,
				"Should visit all %d ConceptScope tables", tc.expectedTotalConcepts)

			// Verify all concepts were analyzed
			defs := analyzer.GetAnalysisDefinitions()
			assert.Len(t, defs.Concepts, tc.expectedTotalConcepts,
				"Should have %d concepts in analysis results", tc.expectedTotalConcepts)
		})
	}
}

// TestAnalyzer_AddConceptDuplicateDetection verifies the AddConcept method
// correctly detects duplicates when called directly.
func TestAnalyzer_AddConceptDuplicateDetection(t *testing.T) {
	t.Parallel()

	// Arrange
	analyzer := analysis.NewCPLAnalyzer()

	concept1 := &analysis.AnalysisConceptDefinition{
		ConceptName:    "Country",
		ResolvedValues: []analysis.AnalysisConceptValueDefinition{{Name: "USA", Position: 0}},
	}

	concept2 := &analysis.AnalysisConceptDefinition{
		ConceptName:    "Country",
		ResolvedValues: []analysis.AnalysisConceptValueDefinition{{Name: "Canada", Position: 0}},
	}

	concept3 := &analysis.AnalysisConceptDefinition{
		ConceptName:    "Priority",
		ResolvedValues: []analysis.AnalysisConceptValueDefinition{{Name: "High", Position: 0}},
	}

	// Act & Assert: First concept should succeed
	err := analyzer.AddConcept(concept1)
	require.NoError(t, err, "First AddConcept should succeed")

	// Duplicate should fail
	err = analyzer.AddConcept(concept2)
	require.Error(t, err, "AddConcept with duplicate name should fail")
	assert.Contains(t, err.Error(), "duplicate concept definition")
	assert.Contains(t, err.Error(), "Country")
	assert.Contains(t, err.Error(), "first at")
	assert.Contains(t, err.Error(), "second at")

	// Different concept should succeed
	err = analyzer.AddConcept(concept3)
	require.NoError(t, err, "AddConcept with different name should succeed")

	// Verify final state
	defs := analyzer.GetAnalysisDefinitions()
	assert.Len(t, defs.Concepts, 2, "Should have exactly 2 concepts")
	assert.Contains(t, defs.Concepts, "Country")
	assert.Contains(t, defs.Concepts, "Priority")
}

// ============================================================================
// Helper Types and Functions
// ============================================================================

// MockHierarchyConcept represents a concept with hierarchy information for testing.
type MockHierarchyConcept struct {
	Name   string
	Values []string
	Parent *string
	Indent int
}

// stringPtr returns a pointer to a string.
func stringPtr(s string) *string {
	return &s
}

// createMockSymbolTableWithMultipleConceptsSections creates a mock symbol table
// with multiple ConceptsScope tables, each containing the specified concepts.
// All concepts are placed in a single ConceptsScope per section for simplicity.
func createMockSymbolTableWithMultipleConceptsSections(concepts []MockConceptDeclaration) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()

	// Group concepts into sections (each concept in its own section for this helper)
	for _, concept := range concepts {
		conceptsScope := root.AddChild(symboltable.ConceptsScope)
		conceptScope := conceptsScope.AddChild(symboltable.ConceptScope)

		_ = conceptScope.Define(&concept.ConceptName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})

		// Add value generator scope
		valueGenScope := conceptScope.AddChild(symboltable.ValueGeneratorScope)
		for _, v := range concept.Values {
			valueGenScope.Define(&v, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
		}
	}

	return root
}

// createMockSymbolTableWithTwoConceptsSections creates a mock symbol table with
// exactly two ConceptsScope tables, each containing the specified concepts.
func createMockSymbolTableWithTwoConceptsSections(
	section1Concepts []MockConceptDeclaration,
	section2Concepts []MockConceptDeclaration,
) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()

	// First section
	conceptsScope1 := root.AddChild(symboltable.ConceptsScope)
	for _, concept := range section1Concepts {
		conceptScope := conceptsScope1.AddChild(symboltable.ConceptScope)
		_ = conceptScope.Define(&concept.ConceptName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})

		valueGenScope := conceptScope.AddChild(symboltable.ValueGeneratorScope)
		for _, v := range concept.Values {
			valueGenScope.Define(&v, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
		}
	}

	// Second section
	conceptsScope2 := root.AddChild(symboltable.ConceptsScope)
	for _, concept := range section2Concepts {
		conceptScope := conceptsScope2.AddChild(symboltable.ConceptScope)
		_ = conceptScope.Define(&concept.ConceptName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})

		valueGenScope := conceptScope.AddChild(symboltable.ValueGeneratorScope)
		for _, v := range concept.Values {
			valueGenScope.Define(&v, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
		}
	}

	return root
}

// createMockSymbolTableWithHierarchicalSections creates a mock symbol table with
// two ConceptsScope tables containing hierarchical concept definitions.
func createMockSymbolTableWithHierarchicalSections(
	section1Hierarchy map[string]MockHierarchyConcept,
	section2Hierarchy map[string]MockHierarchyConcept,
) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()

	// First section with hierarchy
	conceptsScope1 := root.AddChild(symboltable.ConceptsScope)
	var lastConceptScope *symboltable.SymbolTable
	lastIndent := -1

	// Process concepts in order to establish proper parent-child relationships
	// We need to ensure concepts are processed from parent to child
	processHierarchyConcepts(conceptsScope1, section1Hierarchy, &lastConceptScope, &lastIndent)

	// Second section with hierarchy
	conceptsScope2 := root.AddChild(symboltable.ConceptsScope)
	lastConceptScope = nil
	lastIndent = -1
	processHierarchyConcepts(conceptsScope2, section2Hierarchy, &lastConceptScope, &lastIndent)

	return root
}

// processHierarchyConcepts processes hierarchical concepts in order, creating
// the proper symbol table structure with indentation-based parent relationships.
// Concepts are processed in order of increasing indentation to ensure proper
// parent-child relationships are established.
func processHierarchyConcepts(
	conceptsScope *symboltable.SymbolTable,
	hierarchy map[string]MockHierarchyConcept,
	lastConceptScope **symboltable.SymbolTable,
	lastIndent *int,
) {
	// Collect concepts from map into slice for sorting
	type conceptWithIndent struct {
		concept MockHierarchyConcept
		scope   *symboltable.SymbolTable
	}
	var concepts []conceptWithIndent
	for _, hc := range hierarchy {
		concepts = append(concepts, conceptWithIndent{concept: hc})
	}

	// Sort concepts by indentation (parents first, then children)
	for i := 0; i < len(concepts); i++ {
		for j := i + 1; j < len(concepts); j++ {
			if concepts[j].concept.Indent < concepts[i].concept.Indent {
				concepts[i], concepts[j] = concepts[j], concepts[i]
			}
		}
	}

	// Create concept scopes in order
	for i := range concepts {
		hc := &concepts[i].concept
		conceptScope := conceptsScope.AddChild(symboltable.ConceptScope)
		conceptScope.IndentedSpaceEquivalent = hc.Indent
		_ = conceptScope.Define(&hc.Name, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})

		// Add value generator scope
		valueGenScope := conceptScope.AddChild(symboltable.ValueGeneratorScope)
		for _, v := range hc.Values {
			valueGenScope.Define(&v, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
		}

		concepts[i].scope = conceptScope
	}
}

// createMockSymbolTableWithNConceptsSections creates a mock symbol table with
// N ConceptsScope tables, each containing M concepts.
func createMockSymbolTableWithNConceptsSections(numSections, conceptsPerSection int) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()

	conceptCounter := 0
	for s := 0; s < numSections; s++ {
		conceptsScope := root.AddChild(symboltable.ConceptsScope)

		for c := 0; c < conceptsPerSection; c++ {
			conceptCounter++
			conceptName := fmt.Sprintf("Concept%d", conceptCounter)

			conceptScope := conceptsScope.AddChild(symboltable.ConceptScope)
			_ = conceptScope.Define(&conceptName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})

			// Add value generator with one value
			valueGenScope := conceptScope.AddChild(symboltable.ValueGeneratorScope)
			valueName := fmt.Sprintf("Value%d", conceptCounter)
			valueGenScope.Define(&valueName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
		}
	}

	return root
}
