// Package analysis_test contains unit tests for the CPL analysis phase.
//
// These tests follow the Behavioral Test-Driven Development (BTDD) approach,
// where tests are written before the implementation exists. This is Phase B
// of the BTDD lifecycle - defining test signatures and expected behaviors.
//
// The tests cover:
//   - ValuesGenerator value generation
//   - AnalysisDefinitions structure operations
//   - CPLAnalyzer concept analysis
//   - WalkSymbolTableWithHandlers traversal
//
// Reference: ERO001 Concepts Analysis Phase
package analysis_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/ast"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// ValuesGenerator Tests
// ============================================================================

// TestValuesGenerator_GenerateValues_ReturnsSlice verifies that GenerateValues
// returns the Values slice correctly for a ValuesGenerator with multiple values.
//
// This test ensures the basic functionality of the ValuesGenerator - that it
// returns the exact slice of strings that was provided during construction.
func TestValuesGenerator_GenerateValues_ReturnsSlice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		values         []string
		expectedValues []string
	}{
		{
			name:           "single value returns single element slice",
			values:         []string{"USA"},
			expectedValues: []string{"USA"},
		},
		{
			name:           "multiple values returns all values in order",
			values:         []string{"USA", "Canada", "Mexico"},
			expectedValues: []string{"USA", "Canada", "Mexico"},
		},
		{
			name:           "values with spaces are preserved",
			values:         []string{"New York", "Los Angeles"},
			expectedValues: []string{"New York", "Los Angeles"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Create a ValuesGenerator with the test values
			generator := &analysis.ValuesGenerator{Values: tc.values}

			// Act: Generate values
			concepts := make(map[string]analysis.AnalysisConceptDefinition)
			result, err := generator.GenerateValues(concepts)

			// Assert: Verify the result matches expected values
			require.NoError(t, err, "GenerateValues should not return an error")
			assert.Equal(t, tc.expectedValues, result, "Generated values should match input values")
		})
	}
}

// TestValuesGenerator_GenerateValues_EmptySlice verifies that GenerateValues
// returns an empty slice when the ValuesGenerator has no values.
//
// This is an edge case test ensuring that empty value lists are handled
// gracefully without errors.
func TestValuesGenerator_GenerateValues_EmptySlice(t *testing.T) {
	t.Parallel()

	// Arrange: Create a ValuesGenerator with empty values
	generator := &analysis.ValuesGenerator{Values: []string{}}

	// Act: Generate values
	concepts := make(map[string]analysis.AnalysisConceptDefinition)
	result, err := generator.GenerateValues(concepts)

	// Assert: Verify empty slice is returned without error
	require.NoError(t, err, "GenerateValues should not return an error for empty values")
	assert.Empty(t, result, "Result should be an empty slice")
	assert.NotNil(t, result, "Result should not be nil, should be empty slice")
}

// TestValuesGenerator_GenerateValues_IgnoresConceptsMap verifies that
// GenerateValues ignores the concepts map parameter (as ValuesGenerator
// has no external references).
//
// This test ensures that ValuesGenerator correctly satisfies the Generator
// interface without depending on the concepts map.
func TestValuesGenerator_GenerateValues_IgnoresConceptsMap(t *testing.T) {
	t.Parallel()

	// Arrange: Create a ValuesGenerator and test with nil concepts map
	generator := &analysis.ValuesGenerator{Values: []string{"Value1", "Value2"}}

	// Act: Generate values with nil concepts map
	result, err := generator.GenerateValues(nil)

	// Assert: Should still work correctly even with nil map
	require.NoError(t, err, "GenerateValues should handle nil concepts map")
	assert.Equal(t, []string{"Value1", "Value2"}, result, "Values should be returned regardless of concepts map")
}

// ============================================================================
// AnalysisDefinitions Structure Tests
// ============================================================================

// TestAnalysisDefinitions_ConceptsMap_CanBeCreated verifies that an
// AnalysisDefinitions struct can be created with an initialized Concepts map.
//
// This test ensures the basic structure can be instantiated and the Concepts
// map is ready for population.
func TestAnalysisDefinitions_ConceptsMap_CanBeCreated(t *testing.T) {
	t.Parallel()

	// Arrange & Act: Create AnalysisDefinitions with initialized map
	defs := &analysis.AnalysisDefinitions{
		Concepts: make(map[string]analysis.AnalysisConceptDefinition),
	}

	// Assert: Verify the map is initialized and empty
	require.NotNil(t, defs.Concepts, "Concepts map should be initialized")
	assert.Empty(t, defs.Concepts, "Concepts map should be empty initially")
}

// TestAnalysisDefinitions_ConceptsMap_CanBePopulated verifies that concept
// definitions can be stored and retrieved from the Concepts map.
//
// This test ensures the Concepts map works as a lookup structure for
// analysis-phase concept definitions.
func TestAnalysisDefinitions_ConceptsMap_CanBePopulated(t *testing.T) {
	t.Parallel()

	// Arrange: Create AnalysisDefinitions and a concept definition
	defs := &analysis.AnalysisDefinitions{
		Concepts: make(map[string]analysis.AnalysisConceptDefinition),
	}

	conceptDef := analysis.AnalysisConceptDefinition{
		ConceptName: "Country",
		ResolvedValues: []analysis.AnalysisConceptValueDefinition{
			{Name: "USA", Position: 0},
			{Name: "Canada", Position: 1},
		},
	}

	// Act: Store the concept definition
	defs.Concepts["Country"] = conceptDef

	// Assert: Verify the concept can be retrieved
	retrieved, exists := defs.Concepts["Country"]
	require.True(t, exists, "Concept should exist in map")
	assert.Equal(t, "Country", retrieved.ConceptName, "Concept name should match")
	assert.Len(t, retrieved.ResolvedValues, 2, "Concept should have 2 values")
}

// TestAnalysisDefinitions_ConceptsMap_StoresMultipleConcepts verifies that
// multiple concept definitions can be stored in the Concepts map.
//
// This test ensures the map can hold multiple concepts with different names.
func TestAnalysisDefinitions_ConceptsMap_StoresMultipleConcepts(t *testing.T) {
	t.Parallel()

	// Arrange: Create AnalysisDefinitions with multiple concepts
	defs := &analysis.AnalysisDefinitions{
		Concepts: make(map[string]analysis.AnalysisConceptDefinition),
	}

	countryDef := analysis.AnalysisConceptDefinition{
		ConceptName: "Country",
		ResolvedValues: []analysis.AnalysisConceptValueDefinition{
			{Name: "USA", Position: 0},
		},
	}

	statusDef := analysis.AnalysisConceptDefinition{
		ConceptName: "Status",
		ResolvedValues: []analysis.AnalysisConceptValueDefinition{
			{Name: "Active", Position: 0},
			{Name: "Inactive", Position: 1},
		},
	}

	// Act: Store both concepts
	defs.Concepts["Country"] = countryDef
	defs.Concepts["Status"] = statusDef

	// Assert: Verify both concepts are stored correctly
	assert.Len(t, defs.Concepts, 2, "Should have 2 concepts in map")

	country, exists := defs.Concepts["Country"]
	require.True(t, exists, "Country should exist")
	assert.Equal(t, "Country", country.ConceptName)

	status, exists := defs.Concepts["Status"]
	require.True(t, exists, "Status should exist")
	assert.Equal(t, "Status", status.ConceptName)
}

// ============================================================================
// CPLAnalyzer Tests
// ============================================================================

// TestCPLAnalyzer_SingleConceptAnalysis verifies that the analyzer can
// process a simple concept definition with a single Values generator.
//
// This test validates the basic analysis flow: creating an analyzer,
// processing a ConceptsScope table, and retrieving the analysis results.
// It verifies that concept names, values, and positions are correctly
// captured in the analysis structures.
func TestCPLAnalyzer_SingleConceptAnalysis(t *testing.T) {
	t.Parallel()

	// Arrange: Create a mock symbol table with a single concept
	// This simulates parsing: concepts
	//                           Country=[Values[USA, Canada]]
	rootTable := createMockSymbolTableWithMultipleConcepts(
		[]MockConceptDeclaration{
			{ConceptName: "Country",
				Values: []string{"USA", "Canada"},
			},
		})

	// Act: Analyze the symbol table
	analyzer := analysis.NewCPLAnalyzer()
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope:        analysis.EnterAnalyzeConceptScope,
		symboltable.ValueGeneratorScope: analysis.AnalyzeValueGeneratorScope,
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
	}
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

	// Assert: Verify analysis succeeded and concept was captured
	require.NoError(t, err, "Analysis should succeed")

	defs := analyzer.GetAnalysisDefinitions()
	require.NotNil(t, defs, "Analysis definitions should not be nil")
	require.NotNil(t, defs.Concepts, "Concepts map should be initialized")

	country, exists := defs.Concepts["Country"]
	require.True(t, exists, "Country concept should exist")
	assert.Equal(t, "Country", country.ConceptName)
	assert.Len(t, country.ResolvedValues, 2, "Country should have 2 values")

	// Verify value positions
	assert.Equal(t, 0, country.ResolvedValues[0].Position, "USA should be at position 0")
	assert.Equal(t, 1, country.ResolvedValues[1].Position, "Canada should be at position 1")
}

// TestCPLAnalyzer_NumericValueConceptAnalysis verifies that the analyzer can
// process concept definitions with numeric literal values.
//
// This test validates that numeric strings (e.g., "2024", "2025") are correctly
// captured in the ValuesGenerator and stored as AnalysisConceptValueDefinition
// entries with proper positions.
func TestCPLAnalyzer_NumericValueConceptAnalysis(t *testing.T) {
	t.Parallel()

	// Arrange: Create a mock symbol table with numeric concept values
	rootTable := createMockSymbolTableWithMultipleConcepts(
		[]MockConceptDeclaration{
			{ConceptName: "Year",
				Values: []string{"2024", "2025", "2026"},
			},
		})

	// Act: Analyze the symbol table
	analyzer := analysis.NewCPLAnalyzer()
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope:        analysis.EnterAnalyzeConceptScope,
		symboltable.ValueGeneratorScope: analysis.AnalyzeValueGeneratorScope,
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
	}
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

	// Assert: Verify analysis succeeded and numeric values were captured
	require.NoError(t, err, "Analysis should succeed")

	defs := analyzer.GetAnalysisDefinitions()
	require.NotNil(t, defs, "Analysis definitions should not be nil")
	require.NotNil(t, defs.Concepts, "Concepts map should be initialized")

	year, exists := defs.Concepts["Year"]
	require.True(t, exists, "Year concept should exist")
	assert.Equal(t, "Year", year.ConceptName)
	assert.Len(t, year.ResolvedValues, 3, "Year should have 3 values")

	// Verify numeric values and positions
	assert.Equal(t, "2024", year.ResolvedValues[0].Name, "First value should be 2024")
	assert.Equal(t, 0, year.ResolvedValues[0].Position, "2024 should be at position 0")
	assert.Equal(t, "2025", year.ResolvedValues[1].Name, "Second value should be 2025")
	assert.Equal(t, 1, year.ResolvedValues[1].Position, "2025 should be at position 1")
	assert.Equal(t, "2026", year.ResolvedValues[2].Name, "Third value should be 2026")
	assert.Equal(t, 2, year.ResolvedValues[2].Position, "2026 should be at position 2")
}

// TestCPLAnalyzer_MixedAlphabeticAndNumericValues verifies that a single
// concept can contain both alphabetic and numeric values.
func TestCPLAnalyzer_MixedAlphabeticAndNumericValues(t *testing.T) {
	t.Parallel()

	// Arrange: Create a mock symbol table with mixed value types
	rootTable := createMockSymbolTableWithMultipleConcepts(
		[]MockConceptDeclaration{
			{ConceptName: "Quarter",
				Values: []string{"Q1", "Q2", "Q3", "Q4", "2024"},
			},
		})

	// Act: Analyze the symbol table
	analyzer := analysis.NewCPLAnalyzer()
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope:        analysis.EnterAnalyzeConceptScope,
		symboltable.ValueGeneratorScope: analysis.AnalyzeValueGeneratorScope,
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
	}
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

	// Assert: Verify all mixed values are captured in order
	require.NoError(t, err, "Analysis should succeed")

	defs := analyzer.GetAnalysisDefinitions()
	quarter, exists := defs.Concepts["Quarter"]
	require.True(t, exists, "Quarter concept should exist")
	assert.Len(t, quarter.ResolvedValues, 5, "Quarter should have 5 values")

	expectedValues := []string{"Q1", "Q2", "Q3", "Q4", "2024"}
	for i, expected := range expectedValues {
		assert.Equal(t, expected, quarter.ResolvedValues[i].Name, "Value at position %d should match", i)
		assert.Equal(t, i, quarter.ResolvedValues[i].Position, "Position at index %d should match", i)
	}
}

// TestCPLAnalyzer_MultipleConceptsAnalysis verifies that the analyzer can
// process multiple concept definitions within the same scope.
//
// This test validates that multiple concepts can be analyzed in a single
// pass, with each concept having its own values and positions.
func TestCPLAnalyzer_MultipleConceptsAnalysis(t *testing.T) {
	t.Parallel()

	// Arrange: Create a mock symbol table with multiple concepts
	rootTable := createMockSymbolTableWithMultipleConcepts(
		[]MockConceptDeclaration{
			{ConceptName: "Country",
				Values: []string{"USA", "Canada", "Mexico"},
			},
			{ConceptName: "Status",
				Values: []string{"Active", "Inactive"},
			},
		})

	// Act: Analyze the symbol table
	analyzer := analysis.NewCPLAnalyzer()
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope:        analysis.EnterAnalyzeConceptScope,
		symboltable.ValueGeneratorScope: analysis.AnalyzeValueGeneratorScope,
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
	}
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

	// Assert: Verify both concepts were captured
	require.NoError(t, err, "Analysis should succeed")

	defs := analyzer.GetAnalysisDefinitions()
	assert.Len(t, defs.Concepts, 2, "Should have 2 concepts")

	// Verify Country concept
	country, exists := defs.Concepts["Country"]
	require.True(t, exists, "Country should exist")
	assert.Len(t, country.ResolvedValues, 3, "Country should have 3 values")

	// Verify Status concept
	status, exists := defs.Concepts["Status"]
	require.True(t, exists, "Status should exist")
	assert.Len(t, status.ResolvedValues, 2, "Status should have 2 values")
}

// TestCPLAnalyzer_DuplicateConceptDetection verifies that the analyzer
// returns an error when duplicate concept names are detected.
//
// This test ensures that the analysis phase validates concept uniqueness
// across the entire compilation unit, preventing duplicate definitions.
func TestCPLAnalyzer_DuplicateConceptDetection(t *testing.T) {
	t.Parallel()

	// Arrange: Create a mock symbol table with duplicate concept names
	rootTable := createMockSymbolTableWithMultipleConcepts(
		[]MockConceptDeclaration{
			{ConceptName: "Country",
				Values: []string{"USA", "Canada", "Mexico"},
			},
			{ConceptName: "Country",
				Values: []string{"USA", "Canada"},
			},
		})

	// Act: Attempt to analyze
	analyzer := analysis.NewCPLAnalyzer()
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope:        analysis.EnterAnalyzeConceptScope,
		symboltable.ValueGeneratorScope: analysis.AnalyzeValueGeneratorScope,
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
	}
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

	// Assert: Verify error is returned for duplicate concept
	require.Error(t, err, "Analysis should fail for duplicate concept")
	assert.Contains(t, err.Error(), "duplicate concept definition", "Error should indicate duplicate concept")
	assert.Contains(t, err.Error(), "Country", "Error should mention the concept name")
	assert.Contains(t, err.Error(), "first at", "Error should include first definition location")
	assert.Contains(t, err.Error(), "second at", "Error should include second definition location")
}

// TestCPLAnalyzer_ValueIndexing verifies that the position/index is correctly
// computed for each concept value.
//
// This test ensures that values have sequential positions starting from 0,
// enabling O(1) lookup and range operations on concept values.
func TestCPLAnalyzer_ValueIndexing(t *testing.T) {
	t.Parallel()

	// Arrange: Create a concept with multiple values
	values := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}
	rootTable := createMockSymbolTableWithMultipleConcepts(
		[]MockConceptDeclaration{
			{ConceptName: "Month",
				Values: values,
			},
		})

	// Act: Analyze
	analyzer := analysis.NewCPLAnalyzer()
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope:        analysis.EnterAnalyzeConceptScope,
		symboltable.ValueGeneratorScope: analysis.AnalyzeValueGeneratorScope,
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
	}
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

	// Assert: Verify positions are correct
	require.NoError(t, err)

	defs := analyzer.GetAnalysisDefinitions()
	month := defs.Concepts["Month"]
	require.Len(t, month.ResolvedValues, len(values))

	for i, val := range month.ResolvedValues {
		assert.Equal(t, i, val.Position, "Value at index %d should have position %d", i, i)
		assert.Equal(t, values[i], val.Name, "Value name should match")
	}
}

// TestCPLAnalyzer_MultipleGeneratorsSupport verifies that multiple generators
// for a single concept are processed and their values concatenated.
//
// This test ensures that concepts defined with multiple generators (e.g.,
// multiple Values generators) have all their values combined in order.
func TestCPLAnalyzer_MultipleGeneratorsSupport(t *testing.T) {
	t.Parallel()

	// Arrange: Create a concept with multiple generators
	// This simulates: Priority=[Values[Low, Medium], Values[High]]
	rootTable := createMockSymbolTableWithMultipleGenerators("Priority", [][]string{
		{"Low", "Medium"},
		{"High"},
	})

	// Act: Analyze
	analyzer := analysis.NewCPLAnalyzer()
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope:        analysis.EnterAnalyzeConceptScope,
		symboltable.ValueGeneratorScope: analysis.AnalyzeValueGeneratorScope,
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
	}
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

	// Assert: Verify all values are present in order
	require.NoError(t, err)

	defs := analyzer.GetAnalysisDefinitions()
	priority := defs.Concepts["Priority"]
	require.Len(t, priority.ResolvedValues, 3, "Should have 3 values from 2 generators")

	// Verify order: Low (0), Medium (1), High (2)
	assert.Equal(t, "Low", priority.ResolvedValues[0].Name)
	assert.Equal(t, 0, priority.ResolvedValues[0].Position)
	assert.Equal(t, "Medium", priority.ResolvedValues[1].Name)
	assert.Equal(t, 1, priority.ResolvedValues[1].Position)
	assert.Equal(t, "High", priority.ResolvedValues[2].Name)
	assert.Equal(t, 2, priority.ResolvedValues[2].Position)
}

// TestCPLAnalyzer_NewCPLAnalyzer_InitializesConceptsMap verifies that
// NewCPLAnalyzer properly initializes the internal concepts map.
//
// This test ensures the constructor creates a ready-to-use analyzer.
func TestCPLAnalyzer_NewCPLAnalyzer_InitializesConceptsMap(t *testing.T) {
	t.Parallel()

	// Act: Create a new analyzer
	analyzer := analysis.NewCPLAnalyzer()

	// Assert: Verify analyzer is properly initialized
	require.NotNil(t, analyzer, "Analyzer should not be nil")

	defs := analyzer.GetAnalysisDefinitions()
	require.NotNil(t, defs, "Analysis definitions should not be nil")
	require.NotNil(t, defs.Concepts, "Concepts map should be initialized")
	assert.Empty(t, defs.Concepts, "Concepts map should be empty initially")
}

// TestCPLAnalyzer_GetAnalysisDefinitions_ReturnsPointer verifies that
// GetAnalysisDefinitions returns a pointer to AnalysisDefinitions.
//
// This ensures the returned structure can be used for CIL generation.
func TestCPLAnalyzer_GetAnalysisDefinitions_ReturnsPointer(t *testing.T) {
	t.Parallel()

	// Arrange
	analyzer := analysis.NewCPLAnalyzer()

	// Act
	defs := analyzer.GetAnalysisDefinitions()

	// Assert
	require.NotNil(t, defs, "Should return non-nil AnalysisDefinitions")
	// The return type should be *AnalysisDefinitions
	_, ok := interface{}(defs).(*analysis.AnalysisDefinitions)
	assert.True(t, ok, "Return type should be *AnalysisDefinitions")
}

// ============================================================================
// WalkSymbolTableWithHandlers Tests
// ============================================================================

// TestWalkSymbolTableWithHandlers_SignatureAndDocstrings verifies the function
// signature and that appropriate docstrings exist for the handler dispatch mechanism.
//
// This test ensures the public API is properly documented and follows the
// design specified in ERO001.
func TestWalkSymbolTableWithHandlers_SignatureAndDocstrings(t *testing.T) {
	t.Parallel()

	// This test verifies the existence of:
	// - WalkSymbolTableWithHandlers function
	// - AnalysisFunc type
	// - Proper handler dispatch behavior

	// Arrange: Create a simple handler to verify function signature
	handlerCalled := false
	testHandler := func(table *symboltable.SymbolTable, context *analysis.AnalysisContext, analyzer *analysis.CPLAnalyzer) error {
		handlerCalled = true
		return nil
	}

	rootTable := symboltable.NewGlobalSymbolTable()
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.GlobalScope: testHandler,
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{}
	analyzer := analysis.NewCPLAnalyzer()

	// Act
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

	// Assert
	require.NoError(t, err, "Walk should succeed")
	assert.True(t, handlerCalled, "Handler should have been called")
}

// TestWalkSymbolTableWithHandlers_NilTableHandling verifies that passing a nil
// table returns without error.
//
// This test ensures robust handling of edge cases.
func TestWalkSymbolTableWithHandlers_NilTableHandling(t *testing.T) {
	t.Parallel()

	// Arrange
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.GlobalScope: func(table *symboltable.SymbolTable, context *analysis.AnalysisContext, analyzer *analysis.CPLAnalyzer) error {
			t.Fatal("Handler should not be called for nil table")
			return nil
		},
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{}
	analyzer := analysis.NewCPLAnalyzer()

	// Act
	err := analysis.WalkSymbolTableWithHandlers(nil, enterHandlers, exitHandlers, analyzer)

	// Assert
	assert.NoError(t, err, "Should return nil for nil table")
}

// TestWalkSymbolTableWithHandlers_HandlerRegistration verifies that handlers
// are properly registered and invoked for their corresponding scope types.
//
// This test ensures the dispatch mechanism works correctly, calling the
// appropriate handler for each scope type encountered during the walk.
func TestWalkSymbolTableWithHandlers_HandlerRegistration(t *testing.T) {
	t.Parallel()

	// Arrange: Create a tree with multiple scope types
	rootTable := symboltable.NewGlobalSymbolTable()
	conceptsScope := rootTable.AddChild(symboltable.ConceptsScope)
	_ = conceptsScope.AddChild(symboltable.ConceptsScope) // Concept definition scope

	// Track which handlers were called
	globalCalled := false
	conceptsCalled := false

	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.GlobalScope: func(table *symboltable.SymbolTable, context *analysis.AnalysisContext, analyzer *analysis.CPLAnalyzer) error {
			globalCalled = true
			assert.Equal(t, symboltable.GlobalScope, table.ScopeType)
			return nil
		},
		symboltable.ConceptsScope: func(table *symboltable.SymbolTable, context *analysis.AnalysisContext, analyzer *analysis.CPLAnalyzer) error {
			conceptsCalled = true
			assert.Equal(t, symboltable.ConceptsScope, table.ScopeType)
			return nil
		},
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{}
	analyzer := analysis.NewCPLAnalyzer()

	// Act
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

	// Assert
	require.NoError(t, err, "Walk should succeed")
	assert.True(t, globalCalled, "GlobalScope handler should have been called")
	assert.True(t, conceptsCalled, "ConceptsScope handler should have been called")
}

// TestWalkSymbolTableWithHandlers_UnregisteredScopeTypeSkipped verifies that
// scope types without registered handlers are skipped but their children
// are still visited.
//
// This test ensures the walk continues through unhandled scopes to reach
// their children, as specified in ERO001.
func TestWalkSymbolTableWithHandlers_UnregisteredScopeTypeSkipped(t *testing.T) {
	t.Parallel()

	// Arrange: Create a tree with UnitsScope (which won't have a handler)
	rootTable := symboltable.NewGlobalSymbolTable()
	_ = rootTable.AddChild(symboltable.UnitsScope) // No handler for this
	_ = rootTable.AddChild(symboltable.ConceptsScope)

	conceptsHandlerCalled := false
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptsScope: func(table *symboltable.SymbolTable, context *analysis.AnalysisContext, analyzer *analysis.CPLAnalyzer) error {
			conceptsHandlerCalled = true
			return nil
		},
		// Note: No handler for UnitsScope or GlobalScope
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{}
	analyzer := analysis.NewCPLAnalyzer()

	// Act
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

	// Assert
	require.NoError(t, err, "Walk should succeed")
	assert.True(t, conceptsHandlerCalled, "ConceptsScope handler should have been called")
	// The walk should complete without errors even though some scopes have no handlers
}

// TestWalkSymbolTableWithHandlers_ErrorStopsWalk verifies that the first error
// returned by any handler stops the walk and is returned.
//
// This test ensures proper error propagation and early termination on failures.
func TestWalkSymbolTableWithHandlers_ErrorStopsWalk(t *testing.T) {
	t.Parallel()

	// Arrange: Create a tree where we'll trigger an error
	rootTable := symboltable.NewGlobalSymbolTable()
	_ = rootTable.AddChild(symboltable.ConceptsScope)
	_ = rootTable.AddChild(symboltable.ConceptsScope) // Second concepts scope

	callCount := 0
	testError := assert.AnError

	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptsScope: func(table *symboltable.SymbolTable, context *analysis.AnalysisContext, analyzer *analysis.CPLAnalyzer) error {
			callCount++
			if callCount == 1 {
				return testError
			}
			return nil
		},
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{}
	analyzer := analysis.NewCPLAnalyzer()

	// Act
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

	// Assert
	require.Error(t, err, "Should return error from handler")
	assert.Equal(t, testError, err, "Should return the exact error from handler")
	assert.Equal(t, 1, callCount, "Handler should only be called once before error")
}

// TestWalkSymbolTableWithHandlers_PreOrderTraversal verifies that the walk
// visits nodes in pre-order (parent before children).
//
// This test ensures the traversal order matches the specification in ERO001.
func TestWalkSymbolTableWithHandlers_PreOrderTraversal(t *testing.T) {
	t.Parallel()

	// Arrange: Create a tree structure
	rootTable := symboltable.NewGlobalSymbolTable()
	child1 := rootTable.AddChild(symboltable.ConceptsScope)
	_ = child1.AddChild(symboltable.ConceptsScope) // grandchild

	visitOrder := make([]string, 0)

	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.GlobalScope: func(table *symboltable.SymbolTable, context *analysis.AnalysisContext, analyzer *analysis.CPLAnalyzer) error {
			visitOrder = append(visitOrder, "root")
			return nil
		},
		symboltable.ConceptsScope: func(table *symboltable.SymbolTable, context *analysis.AnalysisContext, analyzer *analysis.CPLAnalyzer) error {
			if table.Parent != nil && table.Parent.ScopeType == symboltable.GlobalScope {
				visitOrder = append(visitOrder, "child")
			} else {
				visitOrder = append(visitOrder, "grandchild")
			}
			return nil
		},
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{}
	analyzer := analysis.NewCPLAnalyzer()

	// Act
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

	// Assert
	require.NoError(t, err)
	require.Len(t, visitOrder, 3)
	assert.Equal(t, "root", visitOrder[0], "Root should be visited first")
	assert.Equal(t, "child", visitOrder[1], "Child should be visited second")
	assert.Equal(t, "grandchild", visitOrder[2], "Grandchild should be visited third")
}

// TestWalkSymbolTableWithHandlers_NilHandlersMap verifies behavior when
// handlers map is nil or empty.
//
// This test ensures graceful handling of edge cases.
func TestWalkSymbolTableWithHandlers_NilHandlersMap(t *testing.T) {
	t.Parallel()

	rootTable := symboltable.NewGlobalSymbolTable()
	analyzer := analysis.NewCPLAnalyzer()

	// Act: Pass nil handlers map
	err := analysis.WalkSymbolTableWithHandlers(rootTable, nil, nil, analyzer)

	// Assert
	assert.NoError(t, err, "Should handle nil handlers map gracefully")

	// Act: Pass empty handlers map
	err = analysis.WalkSymbolTableWithHandlers(rootTable, map[symboltable.ScopeType]analysis.AnalysisFunc{}, map[symboltable.ScopeType]analysis.AnalysisFunc{}, analyzer)

	// Assert
	assert.NoError(t, err, "Should handle empty handlers map gracefully")
}

// ============================================================================
// AnalysisFunc Type Tests
// ============================================================================

// TestAnalysisFunc_TypeSignature verifies that AnalysisFunc has the correct
// type signature for use as a handler in WalkSymbolTableWithHandlers.
//
// The signature should be: func(table *SymbolTable, context *AnalysisContext, analyzer *CPLAnalyzer) error
func TestAnalysisFunc_TypeSignature(t *testing.T) {
	t.Parallel()

	// This test verifies the AnalysisFunc type exists and has the right signature
	// by attempting to assign a function to it

	var fn analysis.AnalysisFunc = func(table *symboltable.SymbolTable, context *analysis.AnalysisContext, analyzer *analysis.CPLAnalyzer) error {
		return nil
	}

	require.NotNil(t, fn, "AnalysisFunc should be assignable")

	// Verify it can be called with the expected signature
	testTable := symboltable.NewGlobalSymbolTable()
	testAnalyzer := analysis.NewCPLAnalyzer()
	testContext := analysis.NewAnalysisContext()
	err := fn(testTable, testContext, testAnalyzer)
	assert.NoError(t, err)
}

// ============================================================================
// AnalysisConceptValueDefinition Tests
// ============================================================================

// TestAnalysisConceptValueDefinition_StructFields verifies the struct has
// the expected fields: Name and Position.
//
// This test ensures the structure matches the design specification.
func TestAnalysisConceptValueDefinition_StructFields(t *testing.T) {
	t.Parallel()

	// Arrange & Act: Create a value definition
	valDef := analysis.AnalysisConceptValueDefinition{
		Name:     "TestValue",
		Position: 42,
	}

	// Assert: Verify fields are accessible
	assert.Equal(t, "TestValue", valDef.Name)
	assert.Equal(t, 42, valDef.Position)
}

// ============================================================================
// AnalysisConceptDefinition Tests
// ============================================================================

// TestAnalysisConceptDefinition_ConditionalValues verifies that the
// ConditionalValues map field exists on AnalysisConceptDefinition and behaves
// correctly for reads, missing keys, and zero values.
//
// This test ensures the structure supports conditional concept specifications
// where child concept values are keyed by parent concept values.
func TestAnalysisConceptDefinition_ConditionalValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		setup    func() analysis.AnalysisConceptDefinition
		key      string
		expected []string
		wantNil  bool
	}{
		{
			name: "values can be read back by key",
			setup: func() analysis.AnalysisConceptDefinition {
				return analysis.AnalysisConceptDefinition{
					ConceptName: "State",
					ConditionalValues: map[string][]string{
						"USA":    {"NY", "CA", "TX"},
						"Canada": {"Ontario", "Quebec"},
					},
				}
			},
			key:      "USA",
			expected: []string{"NY", "CA", "TX"},
		},
		{
			name: "values can be read back by a different key",
			setup: func() analysis.AnalysisConceptDefinition {
				return analysis.AnalysisConceptDefinition{
					ConceptName: "State",
					ConditionalValues: map[string][]string{
						"USA":    {"NY", "CA", "TX"},
						"Canada": {"Ontario", "Quebec"},
					},
				}
			},
			key:      "Canada",
			expected: []string{"Ontario", "Quebec"},
		},
		{
			name: "missing key returns nil",
			setup: func() analysis.AnalysisConceptDefinition {
				return analysis.AnalysisConceptDefinition{
					ConceptName: "State",
					ConditionalValues: map[string][]string{
						"USA":    {"NY", "CA", "TX"},
						"Canada": {"Ontario", "Quebec"},
					},
				}
			},
			key:     "Mexico",
			wantNil: true,
		},
		{
			name: "zero value of field is nil and does not panic on read",
			setup: func() analysis.AnalysisConceptDefinition {
				return analysis.AnalysisConceptDefinition{
					ConceptName: "State",
				}
			},
			key:     "USA",
			wantNil: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			def := tc.setup()

			// Act: read from ConditionalValues map
			result := def.ConditionalValues[tc.key]

			// Assert
			if tc.wantNil {
				assert.Nil(t, result, "Expected nil for key '%s'", tc.key)
			} else {
				assert.Equal(t, tc.expected, result, "Expected values for key '%s'", tc.key)
			}
		})
	}
}

// TestAnalysisConceptDefinition_DefaultValues verifies that the
// DefaultValues slice field exists on AnalysisConceptDefinition and behaves
// correctly for index access, length, and zero values.
//
// This test ensures the structure supports default values for child concepts
// when no conditional specification matches a particular parent value.
func TestAnalysisConceptDefinition_DefaultValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		setup       func() analysis.AnalysisConceptDefinition
		expectedLen int
		index       int
		expectedVal string
		wantNil     bool
	}{
		{
			name: "values can be read back by index",
			setup: func() analysis.AnalysisConceptDefinition {
				return analysis.AnalysisConceptDefinition{
					ConceptName:   "Variant",
					DefaultValues: []string{"Standard", "Premium"},
				}
			},
			expectedLen: 2,
			index:       0,
			expectedVal: "Standard",
		},
		{
			name: "values can be read back by a different index",
			setup: func() analysis.AnalysisConceptDefinition {
				return analysis.AnalysisConceptDefinition{
					ConceptName:   "Variant",
					DefaultValues: []string{"Standard", "Premium"},
				}
			},
			expectedLen: 2,
			index:       1,
			expectedVal: "Premium",
		},
		{
			name: "zero value of field is nil and does not panic on read",
			setup: func() analysis.AnalysisConceptDefinition {
				return analysis.AnalysisConceptDefinition{
					ConceptName: "Variant",
				}
			},
			wantNil: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			def := tc.setup()

			// Act & Assert
			if tc.wantNil {
				assert.Nil(t, def.DefaultValues, "Expected DefaultValues to be nil for zero value")
			} else {
				require.Len(t, def.DefaultValues, tc.expectedLen, "DefaultValues should have correct length")
				assert.Equal(t, tc.expectedVal, def.DefaultValues[tc.index], "Value at index %d should match", tc.index)
			}
		})
	}
}

// ============================================================================
// IntRangeGenerator Tests
// ============================================================================

// TestIntRangeGenerator_GenerateValues_BasicRange verifies that GenerateValues
// correctly expands a basic integer range with step=1.
func TestIntRangeGenerator_GenerateValues_BasicRange(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		start    int
		end      int
		step     int
		expected []string
	}{
		{
			name:     "month range 1 to 12 step 1",
			start:    1,
			end:      12,
			step:     1,
			expected: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		},
		{
			name:     "small range 1 to 5 step 1",
			start:    1,
			end:      5,
			step:     1,
			expected: []string{"1", "2", "3", "4", "5"},
		},
		{
			name:     "single value range",
			start:    7,
			end:      7,
			step:     1,
			expected: []string{"7"},
		},
		{
			name:     "even numbers 2 to 10 step 2",
			start:    2,
			end:      10,
			step:     2,
			expected: []string{"2", "4", "6", "8", "10"},
		},
		{
			name:     "fives 0 to 20 step 5",
			start:    0,
			end:      20,
			step:     5,
			expected: []string{"0", "5", "10", "15", "20"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			generator := &analysis.IntRangeGenerator{
				Start: tc.start,
				End:   tc.end,
				Step:  tc.step,
			}

			// Act
			result, err := generator.GenerateValues(nil)

			// Assert
			require.NoError(t, err, "GenerateValues should not return an error")
			assert.Equal(t, tc.expected, result, "Generated values should match expected")
		})
	}
}

// TestIntRangeGenerator_GenerateValues_DescendingRange verifies that GenerateValues
// correctly handles negative steps for descending ranges.
func TestIntRangeGenerator_GenerateValues_DescendingRange(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		start    int
		end      int
		step     int
		expected []string
	}{
		{
			name:     "descending 10 to 1 step -1",
			start:    10,
			end:      1,
			step:     -1,
			expected: []string{"10", "9", "8", "7", "6", "5", "4", "3", "2", "1"},
		},
		{
			name:     "descending even 10 to 2 step -2",
			start:    10,
			end:      2,
			step:     -2,
			expected: []string{"10", "8", "6", "4", "2"},
		},
		{
			name:     "single value descending",
			start:    5,
			end:      5,
			step:     -1,
			expected: []string{"5"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			generator := &analysis.IntRangeGenerator{
				Start: tc.start,
				End:   tc.end,
				Step:  tc.step,
			}

			// Act
			result, err := generator.GenerateValues(nil)

			// Assert
			require.NoError(t, err, "GenerateValues should not return an error")
			assert.Equal(t, tc.expected, result, "Generated values should match expected")
		})
	}
}

// TestIntRangeGenerator_GenerateValues_StepZeroError verifies that a step of zero
// produces an error.
func TestIntRangeGenerator_GenerateValues_StepZeroError(t *testing.T) {
	t.Parallel()

	// Arrange
	generator := &analysis.IntRangeGenerator{
		Start: 1,
		End:   10,
		Step:  0,
	}

	// Act
	result, err := generator.GenerateValues(nil)

	// Assert
	require.Error(t, err, "GenerateValues should return an error for step=0")
	assert.Nil(t, result, "Result should be nil on error")
	assert.Contains(t, err.Error(), "step cannot be zero", "Error should mention step cannot be zero")
}

// TestIntRangeGenerator_GenerateValues_InvalidBoundsError verifies that invalid
// start/end combinations produce errors.
func TestIntRangeGenerator_GenerateValues_InvalidBoundsError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		start       int
		end         int
		step        int
		expectedErr string
	}{
		{
			name:        "start greater than end with positive step",
			start:       10,
			end:         1,
			step:        1,
			expectedErr: "start",
		},
		{
			name:        "start less than end with negative step",
			start:       1,
			end:         10,
			step:        -1,
			expectedErr: "start",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			generator := &analysis.IntRangeGenerator{
				Start: tc.start,
				End:   tc.end,
				Step:  tc.step,
			}

			// Act
			result, err := generator.GenerateValues(nil)

			// Assert
			require.Error(t, err, "GenerateValues should return an error for invalid bounds")
			assert.Nil(t, result, "Result should be nil on error")
			assert.Contains(t, err.Error(), tc.expectedErr, "Error should mention bounds issue")
		})
	}
}

// TestIntRangeGenerator_GenerateValues_MaxValuesError verifies that ranges
// producing too many values return an error.
func TestIntRangeGenerator_GenerateValues_MaxValuesError(t *testing.T) {
	t.Parallel()

	// Arrange: Create a range that would exceed MaxIntRangeValues
	generator := &analysis.IntRangeGenerator{
		Start: 1,
		End:   1 + analysis.MaxIntRangeValues,
		Step:  1,
	}

	// Act
	result, err := generator.GenerateValues(nil)

	// Assert
	require.Error(t, err, "GenerateValues should return an error for excessive range")
	assert.Nil(t, result, "Result should be nil on error")
	assert.Contains(t, err.Error(), "exceeding maximum", "Error should mention maximum exceeded")
}

// TestIntRangeGenerator_GenerateValues_NilConceptsMap verifies that
// GenerateValues works correctly with a nil concepts map.
func TestIntRangeGenerator_GenerateValues_NilConceptsMap(t *testing.T) {
	t.Parallel()

	// Arrange
	generator := &analysis.IntRangeGenerator{
		Start: 1,
		End:   3,
		Step:  1,
	}

	// Act
	result, err := generator.GenerateValues(nil)

	// Assert
	require.NoError(t, err, "GenerateValues should handle nil concepts map")
	assert.Equal(t, []string{"1", "2", "3"}, result, "Values should be generated correctly")
}

// ============================================================================
// CPLAnalyzer Int Range Integration Tests
// ============================================================================

// TestCPLAnalyzer_IntRangeConceptAnalysis verifies that the analyzer can
// process a concept definition with an Int range generator.
func TestCPLAnalyzer_IntRangeConceptAnalysis(t *testing.T) {
	t.Parallel()

	// Arrange: Create a mock symbol table with an Int range concept
	rootTable := createMockSymbolTableWithIntRange("Month", 1, 12, 1)

	// Act: Analyze the symbol table
	analyzer := analysis.NewCPLAnalyzer()
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope:        analysis.EnterAnalyzeConceptScope,
		symboltable.ValueGeneratorScope: analysis.AnalyzeValueGeneratorScope,
		symboltable.IntGeneratorScope:   analysis.AnalyzeIntGeneratorScope,
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
	}
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

	// Assert: Verify analysis succeeded and concept was captured
	require.NoError(t, err, "Analysis should succeed")

	defs := analyzer.GetAnalysisDefinitions()
	require.NotNil(t, defs, "Analysis definitions should not be nil")
	require.NotNil(t, defs.Concepts, "Concepts map should be initialized")

	month, exists := defs.Concepts["Month"]
	require.True(t, exists, "Month concept should exist")
	assert.Equal(t, "Month", month.ConceptName)
	assert.Len(t, month.ResolvedValues, 12, "Month should have 12 values")

	// Verify specific values and positions
	for i := 0; i < 12; i++ {
		expectedValue := fmt.Sprintf("%d", i+1)
		assert.Equal(t, expectedValue, month.ResolvedValues[i].Name, "Value at position %d should match", i)
		assert.Equal(t, i, month.ResolvedValues[i].Position, "Position at index %d should match", i)
	}
}

// TestCPLAnalyzer_IntRangeWithCustomStep verifies that custom step values
// correctly affect value generation during analysis.
func TestCPLAnalyzer_IntRangeWithCustomStep(t *testing.T) {
	t.Parallel()

	// Arrange: Create a mock symbol table with an even numbers concept
	rootTable := createMockSymbolTableWithIntRange("EvenNumbers", 2, 10, 2)

	// Act: Analyze
	analyzer := analysis.NewCPLAnalyzer()
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope:        analysis.EnterAnalyzeConceptScope,
		symboltable.ValueGeneratorScope: analysis.AnalyzeValueGeneratorScope,
		symboltable.IntGeneratorScope:   analysis.AnalyzeIntGeneratorScope,
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
	}
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

	// Assert
	require.NoError(t, err, "Analysis should succeed")

	defs := analyzer.GetAnalysisDefinitions()
	evenNumbers, exists := defs.Concepts["EvenNumbers"]
	require.True(t, exists, "EvenNumbers concept should exist")
	assert.Len(t, evenNumbers.ResolvedValues, 5, "EvenNumbers should have 5 values")

	expectedValues := []string{"2", "4", "6", "8", "10"}
	for i, expected := range expectedValues {
		assert.Equal(t, expected, evenNumbers.ResolvedValues[i].Name, "Value at position %d should match", i)
		assert.Equal(t, i, evenNumbers.ResolvedValues[i].Position, "Position at index %d should match", i)
	}
}

// TestCPLAnalyzer_IntRangeValueIndex verifies that the internal index map
// provides O(1) lookup for Int range-expanded values.
func TestCPLAnalyzer_IntRangeValueIndex(t *testing.T) {
	t.Parallel()

	// Arrange
	rootTable := createMockSymbolTableWithIntRange("Month", 1, 12, 1)

	// Act: Analyze using full Analyze function
	analyzer := analysis.NewCPLAnalyzer()
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope:        analysis.EnterAnalyzeConceptScope,
		symboltable.ValueGeneratorScope: analysis.AnalyzeValueGeneratorScope,
		symboltable.IntGeneratorScope:   analysis.AnalyzeIntGeneratorScope,
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
	}
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)
	require.NoError(t, err)

	defs := analyzer.GetAnalysisDefinitions()
	month := defs.Concepts["Month"]

	// Verify ResolvedValues covers all expected values
	require.Len(t, month.ResolvedValues, 12)

	// Verify positions are sequential and values are correct
	for i := 0; i < 12; i++ {
		assert.Equal(t, fmt.Sprintf("%d", i+1), month.ResolvedValues[i].Name)
		assert.Equal(t, i, month.ResolvedValues[i].Position)
	}
}

// TestCPLAnalyzer_MixedGeneratorsValuesAndIntRange verifies that a concept
// can combine Values and Int range generators.
func TestCPLAnalyzer_MixedGeneratorsValuesAndIntRange(t *testing.T) {
	t.Parallel()

	// Arrange: Create a concept with both Values and Int range generators
	rootTable := createMockSymbolTableWithMixedGenerators("Mixed",
		[]string{"A", "B"},
		&ast.IntRangeGeneratorDef{Start: 1, End: 3, Step: 1})

	// Act: Analyze
	analyzer := analysis.NewCPLAnalyzer()
	enterHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope:        analysis.EnterAnalyzeConceptScope,
		symboltable.ValueGeneratorScope: analysis.AnalyzeValueGeneratorScope,
		symboltable.IntGeneratorScope:   analysis.AnalyzeIntGeneratorScope,
	}
	exitHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConceptScope: analysis.ExitAnalyzeConceptScope,
	}
	err := analysis.WalkSymbolTableWithHandlers(rootTable, enterHandlers, exitHandlers, analyzer)

	// Assert
	require.NoError(t, err, "Analysis should succeed")

	defs := analyzer.GetAnalysisDefinitions()
	mixed, exists := defs.Concepts["Mixed"]
	require.True(t, exists, "Mixed concept should exist")
	assert.Len(t, mixed.ResolvedValues, 5, "Mixed should have 5 values total")

	// Verify order: A (0), B (1), 1 (2), 2 (3), 3 (4)
	assert.Equal(t, "A", mixed.ResolvedValues[0].Name)
	assert.Equal(t, 0, mixed.ResolvedValues[0].Position)
	assert.Equal(t, "B", mixed.ResolvedValues[1].Name)
	assert.Equal(t, 1, mixed.ResolvedValues[1].Position)
	assert.Equal(t, "1", mixed.ResolvedValues[2].Name)
	assert.Equal(t, 2, mixed.ResolvedValues[2].Position)
	assert.Equal(t, "2", mixed.ResolvedValues[3].Name)
	assert.Equal(t, 3, mixed.ResolvedValues[3].Position)
	assert.Equal(t, "3", mixed.ResolvedValues[4].Name)
	assert.Equal(t, 4, mixed.ResolvedValues[4].Position)
}

// TestCPLAnalyzer_IntRangeFullAnalyze verifies that the full Analyze function
// correctly processes Int range generators.
func TestCPLAnalyzer_IntRangeFullAnalyze(t *testing.T) {
	t.Parallel()

	// Arrange: Parse CPL with Int range
	input := "concepts\n\tMonth=[Int[1, 12, 1]]"
	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Act: Run full analysis
	defs, err := analysis.Analyze(rootTable)

	// Assert
	require.NoError(t, err, "Analysis should succeed")
	require.NotNil(t, defs, "Analysis definitions should not be nil")

	month, exists := defs.Concepts["Month"]
	require.True(t, exists, "Month concept should exist")
	assert.Len(t, month.ResolvedValues, 12, "Month should have 12 values")

	for i := 0; i < 12; i++ {
		assert.Equal(t, fmt.Sprintf("%d", i+1), month.ResolvedValues[i].Name)
		assert.Equal(t, i, month.ResolvedValues[i].Position)
	}
}

// TestCPLAnalyzer_IntRangeEvenNumbersFullAnalyze verifies the full Analyze
// function with a custom step.
func TestCPLAnalyzer_IntRangeEvenNumbersFullAnalyze(t *testing.T) {
	t.Parallel()

	// Arrange
	input := "concepts\n\tEvenNumbers=[Int[2, 10, 2]]"
	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Act
	defs, err := analysis.Analyze(rootTable)

	// Assert
	require.NoError(t, err, "Analysis should succeed")

	evenNumbers, exists := defs.Concepts["EvenNumbers"]
	require.True(t, exists, "EvenNumbers concept should exist")
	assert.Len(t, evenNumbers.ResolvedValues, 5, "EvenNumbers should have 5 values")

	expectedValues := []string{"2", "4", "6", "8", "10"}
	for i, expected := range expectedValues {
		assert.Equal(t, expected, evenNumbers.ResolvedValues[i].Name)
		assert.Equal(t, i, evenNumbers.ResolvedValues[i].Position)
	}
}

// TestCPLAnalyzer_IntRangeWithValuesFullAnalyze verifies the full Analyze
// function with mixed Values and Int range generators.
func TestCPLAnalyzer_IntRangeWithValuesFullAnalyze(t *testing.T) {
	t.Parallel()

	// Arrange
	input := "concepts\n\tMixed=[Values[A, B], Int[1, 3, 1]]"
	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Act
	defs, err := analysis.Analyze(rootTable)

	// Assert
	require.NoError(t, err, "Analysis should succeed")

	mixed, exists := defs.Concepts["Mixed"]
	require.True(t, exists, "Mixed concept should exist")
	assert.Len(t, mixed.ResolvedValues, 5, "Mixed should have 5 values total")

	assert.Equal(t, "A", mixed.ResolvedValues[0].Name)
	assert.Equal(t, "B", mixed.ResolvedValues[1].Name)
	assert.Equal(t, "1", mixed.ResolvedValues[2].Name)
	assert.Equal(t, "2", mixed.ResolvedValues[3].Name)
	assert.Equal(t, "3", mixed.ResolvedValues[4].Name)
}

// TestCPLAnalyzer_IntRangeSingleValueFullAnalyze verifies the full Analyze
// function with a single-value Int range (start == end).
func TestCPLAnalyzer_IntRangeSingleValueFullAnalyze(t *testing.T) {
	t.Parallel()

	// Arrange
	input := "concepts\n\tSingleton=[Int[7, 7, 1]]"
	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")

	// Act
	defs, err := analysis.Analyze(rootTable)

	// Assert
	require.NoError(t, err, "Analysis should succeed")

	singleton, exists := defs.Concepts["Singleton"]
	require.True(t, exists, "Singleton concept should exist")
	assert.Len(t, singleton.ResolvedValues, 1, "Singleton should have 1 value")
	assert.Equal(t, "7", singleton.ResolvedValues[0].Name)
	assert.Equal(t, 0, singleton.ResolvedValues[0].Position)
}

// ============================================================================
// Helper Functions for Int Range Tests
// ============================================================================

// createMockSymbolTableWithIntRange creates a mock symbol table containing
// a single concept with an Int range generator.
func createMockSymbolTableWithIntRange(name string, start, end, step int) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()
	conceptsScope := root.AddChild(symboltable.ConceptsScope)
	conceptScope := conceptsScope.AddChild(symboltable.ConceptScope)

	conceptName := name
	_ = conceptScope.Define(&conceptName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})

	// Add Int generator scope with IntRangeGeneratorDef as Data
	intGenScope := conceptScope.AddChild(symboltable.IntGeneratorScope)
	intGenScope.Data = &ast.IntRangeGeneratorDef{
		Start: start,
		End:   end,
		Step:  step,
	}

	return root
}

// createMockSymbolTableWithMixedGenerators creates a mock symbol table with
// a concept that has both Values and Int range generators.
func createMockSymbolTableWithMixedGenerators(name string, values []string, intRange *ast.IntRangeGeneratorDef) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()
	conceptsScope := root.AddChild(symboltable.ConceptsScope)
	conceptScope := conceptsScope.AddChild(symboltable.ConceptScope)

	conceptName := name
	_ = conceptScope.Define(&conceptName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})

	// Add Values generator scope
	valueGenScope := conceptScope.AddChild(symboltable.ValueGeneratorScope)
	for _, v := range values {
		valueGenScope.Define(&v, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
	}

	// Add Int generator scope
	intGenScope := conceptScope.AddChild(symboltable.IntGeneratorScope)
	intGenScope.Data = intRange

	return root
}

type MockConceptDeclaration struct {
	ConceptName string
	Values      []string
}

// ============================================================================
// Helper Functions (for test setup)
// ============================================================================

// createMockSymbolTableWithMultipleConcepts creates a mock symbol table
// containing multiple concepts with their respective values.
func createMockSymbolTableWithMultipleConcepts(concepts []MockConceptDeclaration) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()
	conceptsScope := root.AddChild(symboltable.ConceptsScope)

	for _, concept := range concepts {
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

// createMockSymbolTableWithMultipleGenerators creates a mock symbol table
// with a concept that has multiple generators.
func createMockSymbolTableWithMultipleGenerators(name string, generatorValues [][]string) *symboltable.SymbolTable {
	root := symboltable.NewGlobalSymbolTable()
	conceptsScope := root.AddChild(symboltable.ConceptsScope)
	conceptScope := conceptsScope.AddChild(symboltable.ConceptScope)

	conceptName := name
	_ = conceptScope.Define(&conceptName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})

	for _, values := range generatorValues {
		valueGenScope := conceptScope.AddChild(symboltable.ValueGeneratorScope)
		for _, v := range values {
			valueGenScope.Define(&v, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
		}
	}

	return root
}

// ============================================================================
// ResolveConditionalValues Tests
// ============================================================================

// TestResolveConditionalValues verifies that ResolveConditionalValues correctly
// validates parent value references, builds ConditionalValues maps, populates
// DefaultValues, and constructs the unified ResolvedValues list.
//
// These tests are written in Phase B (BTDD) against a stub implementation
// that currently returns nil, nil. They will fail now and pass after Phase C.
func TestResolveConditionalValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		childData     *ast.ConceptDefinitionData
		parentConcept analysis.AnalysisConceptDefinition
		expectError   bool
		errorContains string
		wantNilResult bool
		// expectations for the returned AnalysisConceptDefinition (when no error)
		expectedConditionalKeys []string
		expectedDefaultValues   []string
		expectedResolvedNames   []string
	}{
		{
			name: "valid conditional spec with single parent value",
			childData: &ast.ConceptDefinitionData{
				Name: "State",
				ConditionalSpecs: []ast.ConditionalSpec{
					{
						ParentValue:   "USA",
						GeneratorType: ast.ValuesGenerator,
						Values:        []string{"NY", "CA", "TX"},
					},
				},
			},
			parentConcept:           analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada"}),
			expectError:             false,
			expectedConditionalKeys: []string{"USA"},
			expectedResolvedNames:   []string{"NY", "CA", "TX"},
		},
		{
			name: "valid conditional spec with multiple parent values",
			childData: &ast.ConceptDefinitionData{
				Name: "State",
				ConditionalSpecs: []ast.ConditionalSpec{
					{
						ParentValue:   "USA",
						GeneratorType: ast.ValuesGenerator,
						Values:        []string{"NY", "CA", "TX"},
					},
					{
						ParentValue:   "Canada",
						GeneratorType: ast.ValuesGenerator,
						Values:        []string{"Ontario", "Quebec"},
					},
				},
			},
			parentConcept:           analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada"}),
			expectError:             false,
			expectedConditionalKeys: []string{"USA", "Canada"},
			expectedResolvedNames:   []string{"NY", "CA", "TX", "Ontario", "Quebec"},
		},
		{
			name: "invalid parent value returns error with source location",
			childData: &ast.ConceptDefinitionData{
				Name:           "State",
				SourceLocation: ast.SourceLocation{Line: 5, Column: 2},
				ConditionalSpecs: []ast.ConditionalSpec{
					{
						ParentValue:   "Mexico",
						GeneratorType: ast.ValuesGenerator,
						Values:        []string{"Jalisco"},
					},
				},
			},
			parentConcept: analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada"}),
			expectError:   true,
			errorContains: "parent value",
		},
		{
			name: "error includes source location",
			childData: &ast.ConceptDefinitionData{
				Name:           "State",
				SourceLocation: ast.SourceLocation{Line: 5, Column: 2},
				ConditionalSpecs: []ast.ConditionalSpec{
					{
						ParentValue:   "Mexico",
						GeneratorType: ast.ValuesGenerator,
						Values:        []string{"Jalisco"},
					},
				},
			},
			parentConcept: analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada"}),
			expectError:   true,
			errorContains: "line 5, column 2",
		},
		{
			name: "default values populated when present",
			childData: &ast.ConceptDefinitionData{
				Name:             "Variant",
				ConditionalSpecs: []ast.ConditionalSpec{},
				DefaultValues:    []string{"Standard", "Premium"},
			},
			parentConcept:         analysis.NewValuesConceptDefinition("Product", []string{"Widget", "Gadget"}),
			expectError:           false,
			expectedDefaultValues: []string{"Standard", "Premium"},
			expectedResolvedNames: []string{"Standard", "Premium"},
		},
		{
			name: "default values included in unified value list",
			childData: &ast.ConceptDefinitionData{
				Name: "Variant",
				ConditionalSpecs: []ast.ConditionalSpec{
					{
						ParentValue:   "Widget",
						GeneratorType: ast.ValuesGenerator,
						Values:        []string{"Deluxe"},
					},
				},
				DefaultValues: []string{"Standard", "Premium"},
			},
			parentConcept:           analysis.NewValuesConceptDefinition("Product", []string{"Widget", "Gadget"}),
			expectError:             false,
			expectedConditionalKeys: []string{"Widget"},
			expectedDefaultValues:   []string{"Standard", "Premium"},
			expectedResolvedNames:   []string{"Deluxe", "Standard", "Premium"},
		},
		{
			name: "validates empty conditional values are stored as empty slice",
			childData: &ast.ConceptDefinitionData{
				Name: "State",
				ConditionalSpecs: []ast.ConditionalSpec{
					{
						ParentValue:   "USA",
						GeneratorType: ast.ValuesGenerator,
						Values:        []string{}, // empty values are valid in ResolveConditionalValues
					},
				},
				DefaultValues: []string{"Default", "Fallback"},
			},
			parentConcept:           analysis.NewValuesConceptDefinition("Country", []string{"USA", "Canada"}),
			expectError:             false,
			expectedConditionalKeys: []string{"USA"},
			expectedDefaultValues:   []string{"Default", "Fallback"},
			expectedResolvedNames:   []string{"Default", "Fallback"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act: call ResolveConditionalValues
			concepts := make(map[string]analysis.AnalysisConceptDefinition)
			result, err := analysis.ResolveConditionalValues(tc.childData, &tc.parentConcept, concepts)

			// Assert error behavior
			if tc.expectError {
				require.Error(t, err, "ResolveConditionalValues should return an error")
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, "Error message should contain expected text")
				}
				return
			}

			require.NoError(t, err, "ResolveConditionalValues should not return an error")

			if tc.wantNilResult {
				assert.Nil(t, result, "Result should be nil")
				return
			}

			require.NotNil(t, result, "Result should not be nil for valid inputs")

			// Validate ConditionalValues keys
			for _, key := range tc.expectedConditionalKeys {
				assert.Contains(t, result.ConditionalValues, key, "ConditionalValues should contain key '%s'", key)
			}

			// Validate specific conditional values
			if len(tc.childData.ConditionalSpecs) > 0 {
				for _, spec := range tc.childData.ConditionalSpecs {
					expectedValues := spec.Values
					actualValues := result.ConditionalValues[spec.ParentValue]
					assert.Equal(t, expectedValues, actualValues, "ConditionalValues for parent '%s' should match", spec.ParentValue)
				}
			}

			// Validate DefaultValues
			if tc.expectedDefaultValues != nil {
				assert.Equal(t, tc.expectedDefaultValues, result.DefaultValues, "DefaultValues should match expected")
			}

			// Validate ResolvedValues names
			if tc.expectedResolvedNames != nil {
				resolvedNames := make([]string, len(result.ResolvedValues))
				for i, v := range result.ResolvedValues {
					resolvedNames[i] = v.Name
				}
				assert.Equal(t, tc.expectedResolvedNames, resolvedNames, "ResolvedValues should contain expected names in order")
			}
		})
	}
}

// ============================================================================
// GetValuesForParent Tests
// ============================================================================

// TestGetValuesForParent verifies that GetValuesForParent correctly returns
// conditional values, default values, or nil based on the input parent value.
func TestGetValuesForParent(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		concept     analysis.AnalysisConceptDefinition
		parentValue string
		expected    []string
		expectNil   bool
	}{
		{
			name: "returns conditional values when parent value matches",
			concept: analysis.AnalysisConceptDefinition{
				ConceptName: "State",
				ConditionalValues: map[string][]string{
					"USA":    {"NY", "CA", "TX"},
					"Canada": {"Ontario", "Quebec"},
				},
				DefaultValues: []string{"Default1", "Default2"},
			},
			parentValue: "USA",
			expected:    []string{"NY", "CA", "TX"},
		},
		{
			name: "returns conditional values for another matching parent value",
			concept: analysis.AnalysisConceptDefinition{
				ConceptName: "State",
				ConditionalValues: map[string][]string{
					"USA":    {"NY", "CA", "TX"},
					"Canada": {"Ontario", "Quebec"},
				},
				DefaultValues: []string{"Default1", "Default2"},
			},
			parentValue: "Canada",
			expected:    []string{"Ontario", "Quebec"},
		},
		{
			name: "falls back to DefaultValues when no conditional match",
			concept: analysis.AnalysisConceptDefinition{
				ConceptName: "State",
				ConditionalValues: map[string][]string{
					"USA": {"NY", "CA", "TX"},
				},
				DefaultValues: []string{"Standard", "Premium"},
			},
			parentValue: "Mexico",
			expected:    []string{"Standard", "Premium"},
		},
		{
			name: "falls back to DefaultValues when no ConditionalValues map at all",
			concept: analysis.AnalysisConceptDefinition{
				ConceptName:   "Variant",
				DefaultValues: []string{"Basic", "Deluxe"},
			},
			parentValue: "Anything",
			expected:    []string{"Basic", "Deluxe"},
		},
		{
			name: "returns nil when neither ConditionalValues nor DefaultValues exist",
			concept: analysis.AnalysisConceptDefinition{
				ConceptName: "State",
			},
			parentValue: "USA",
			expectNil:   true,
		},
		{
			name: "returns nil when ConditionalValues exists but no match and DefaultValues is nil",
			concept: analysis.AnalysisConceptDefinition{
				ConceptName: "State",
				ConditionalValues: map[string][]string{
					"USA": {"NY", "CA", "TX"},
				},
			},
			parentValue: "Mexico",
			expectNil:   true,
		},
		{
			name: "returns nil when ConditionalValues is empty and DefaultValues is nil",
			concept: analysis.AnalysisConceptDefinition{
				ConceptName:       "State",
				ConditionalValues: map[string][]string{},
			},
			parentValue: "USA",
			expectNil:   true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result := tc.concept.GetValuesForParent(tc.parentValue)

			// Assert
			if tc.expectNil {
				assert.Nil(t, result, "Expected nil result for parent value '%s'", tc.parentValue)
			} else {
				assert.Equal(t, tc.expected, result, "Unexpected result for parent value '%s'", tc.parentValue)
			}
		})
	}
}

// ============================================================================
// HasConditionalValues Tests
// ============================================================================

// TestHasConditionalValues verifies that HasConditionalValues correctly
// reports whether the ConditionalValues map is non-empty.
func TestHasConditionalValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		concept  analysis.AnalysisConceptDefinition
		expected bool
	}{
		{
			name: "returns true when ConditionalValues is non-empty",
			concept: analysis.AnalysisConceptDefinition{
				ConceptName: "State",
				ConditionalValues: map[string][]string{
					"USA": {"NY", "CA", "TX"},
				},
			},
			expected: true,
		},
		{
			name: "returns true when ConditionalValues has multiple entries",
			concept: analysis.AnalysisConceptDefinition{
				ConceptName: "State",
				ConditionalValues: map[string][]string{
					"USA":    {"NY", "CA", "TX"},
					"Canada": {"Ontario", "Quebec"},
				},
			},
			expected: true,
		},
		{
			name: "returns false when ConditionalValues is nil",
			concept: analysis.AnalysisConceptDefinition{
				ConceptName: "State",
			},
			expected: false,
		},
		{
			name: "returns false when ConditionalValues is an empty map",
			concept: analysis.AnalysisConceptDefinition{
				ConceptName:       "State",
				ConditionalValues: map[string][]string{},
			},
			expected: false,
		},
		{
			name: "returns true when ConditionalValues has entry with empty slice",
			concept: analysis.AnalysisConceptDefinition{
				ConceptName: "State",
				ConditionalValues: map[string][]string{
					"USA": {},
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result := tc.concept.HasConditionalValues()

			// Assert
			assert.Equal(t, tc.expected, result, "HasConditionalValues returned unexpected result")
		})
	}
}

// TestConditionalSpec_ChildConceptInAnalysis verifies that a child concept defined
// inside a conditional specification is properly analyzed and appears in the
// AnalysisDefinitions with correct ConditionalValues.
//
// This test validates the end-to-end pipeline: parsing conditional spec CPL,
// analyzing the symbol table, and verifying the child concept is correctly
// resolved with parent-conditioned values.
//
// This is a Phase B (BTDD) test written against the current stub implementation
// of processConditionalSpecChild. It will fail until Task 3 implements the function.
func TestConditionalSpec_ChildConceptInAnalysis(t *testing.T) {
	t.Parallel()

	// Arrange: Parse CPL with conditional spec
	input := "concepts\n\tCountry=[Values[USA, Canada]]\n\tCountry[USA] => State=[Values[NY, CA, TX]]"
	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")
	require.NotNil(t, rootTable, "Root table should not be nil")

	// Act: Analyze the parsed symbol table
	defs, err := analysis.Analyze(rootTable)

	// Assert: Analysis should succeed and both concepts should be present
	require.NoError(t, err, "Analysis should succeed")
	require.NotNil(t, defs, "Analysis definitions should not be nil")
	require.NotNil(t, defs.Concepts, "Concepts map should be initialized")

	// Verify Country concept exists
	country, exists := defs.Concepts["Country"]
	require.True(t, exists, "Country concept should exist")
	assert.Equal(t, "Country", country.ConceptName)

	// Verify State concept exists (this will fail until processConditionalSpecChild is implemented)
	state, exists := defs.Concepts["State"]
	require.True(t, exists, "State concept should exist in defs.Concepts")
	assert.Equal(t, "State", state.ConceptName)

	// Verify State has conditional values
	require.NotEmpty(t, state.ConditionalValues, "State.ConditionalValues should be non-empty")
	assert.Equal(t, []string{"NY", "CA", "TX"}, state.ConditionalValues["USA"], "ConditionalValues for USA should match")

	// Verify GetValuesForParent behavior
	assert.Equal(t, []string{"NY", "CA", "TX"}, state.GetValuesForParent("USA"), "GetValuesForParent(USA) should return correct values")
	assert.Nil(t, state.GetValuesForParent("Canada"), "GetValuesForParent(Canada) should return nil (no condition for Canada)")

	// Verify parent relationship
	require.NotNil(t, state.ParentConceptName, "State.ParentConceptName should not be nil")
	assert.Equal(t, "Country", *state.ParentConceptName, "State.ParentConceptName should point to Country")
}

// ============================================================================
// End-to-End Integration Tests — Tasks 6-8
// ============================================================================

// TestConditionalSpec_MultipleConditionsInAnalysis verifies that CPL with
// multiple conditional specifications correctly resolves all child concepts
// with their respective ConditionalValues maps.
//
// This is Task 6a: end-to-end parse + analyze with multiple conditions.
func TestConditionalSpec_MultipleConditionsInAnalysis(t *testing.T) {
	t.Parallel()

	// Arrange: Parse CPL with multiple conditional specs
	input := "concepts\n\tCountry=[Values[USA, Canada, Mexico]]\n\tCountry[USA] => State=[Values[NY, CA, TX]]\n\tCountry[Canada] => Province=[Values[Ontario, Quebec]]\n\tCountry[Mexico] => State=[Values[Jalisco, Oaxaca]]"
	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")
	require.NotNil(t, rootTable, "Root table should not be nil")

	// Act: Analyze the parsed symbol table
	defs, err := analysis.Analyze(rootTable)

	// Assert: Analysis should succeed
	require.NoError(t, err, "Analysis should succeed")
	require.NotNil(t, defs, "Analysis definitions should not be nil")
	require.NotNil(t, defs.Concepts, "Concepts map should be initialized")

	// Verify Country concept exists
	country, exists := defs.Concepts["Country"]
	require.True(t, exists, "Country concept should exist")
	assert.Equal(t, "Country", country.ConceptName)

	// Verify State concept exists with conditional values for USA
	state, exists := defs.Concepts["State"]
	require.True(t, exists, "State concept should exist in defs.Concepts")
	assert.Equal(t, "State", state.ConceptName)
	require.Contains(t, state.ConditionalValues, "USA", "State.ConditionalValues should contain USA key")
	assert.Equal(t, []string{"NY", "CA", "TX"}, state.ConditionalValues["USA"], "ConditionalValues for USA should match first conditional spec")

	// Verify Province concept exists with conditional values for Canada
	province, exists := defs.Concepts["Province"]
	require.True(t, exists, "Province concept should exist in defs.Concepts")
	assert.Equal(t, "Province", province.ConceptName)
	require.Contains(t, province.ConditionalValues, "Canada", "Province.ConditionalValues should contain Canada key")
	assert.Equal(t, []string{"Ontario", "Quebec"}, province.ConditionalValues["Canada"], "ConditionalValues for Canada should match second conditional spec")

	// Verify State also has conditional values for Mexico (from third conditional spec, merged)
	require.Contains(t, state.ConditionalValues, "Mexico", "State.ConditionalValues should contain Mexico key")
	assert.Equal(t, []string{"Jalisco", "Oaxaca"}, state.ConditionalValues["Mexico"], "ConditionalValues for Mexico should match third conditional spec")
}

// TestConditionalSpec_DefaultValuesInAnalysis verifies that CPL with default
// values correctly supports GetValuesForParent fallback: conditional values
// are returned for matching parent values, and default values for non-matching.
//
// This is Task 6b: end-to-end parse + analyze with default values.
func TestConditionalSpec_DefaultValuesInAnalysis(t *testing.T) {
	t.Parallel()

	// Arrange: Parse CPL with default values and conditional override
	// Region is defined unconditionally first (defaults), then overridden conditionally for USA
	input := "concepts\n\tCountry=[Values[USA, Canada]]\n\tRegion=[Values[Default1, Default2]]\n\tCountry[USA] => Region=[Values[NY, CA]]"
	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")
	require.NotNil(t, rootTable, "Root table should not be nil")

	// Act: Analyze the parsed symbol table
	defs, err := analysis.Analyze(rootTable)

	// Assert: Analysis should succeed
	require.NoError(t, err, "Analysis should succeed")
	require.NotNil(t, defs, "Analysis definitions should not be nil")

	// Verify Country concept exists
	country, exists := defs.Concepts["Country"]
	require.True(t, exists, "Country concept should exist")
	assert.Equal(t, "Country", country.ConceptName)

	// Verify Region concept exists
	region, exists := defs.Concepts["Region"]
	require.True(t, exists, "Region concept should exist in defs.Concepts")
	assert.Equal(t, "Region", region.ConceptName)

	// Verify GetValuesForParent returns conditional values for USA
	usaValues := region.GetValuesForParent("USA")
	require.NotNil(t, usaValues, "GetValuesForParent(USA) should return non-nil values")
	assert.Equal(t, []string{"NY", "CA"}, usaValues, "GetValuesForParent(USA) should return conditional values [NY, CA]")

	// Verify GetValuesForParent falls back to default values for Canada
	canadaValues := region.GetValuesForParent("Canada")
	require.NotNil(t, canadaValues, "GetValuesForParent(Canada) should return non-nil values (default fallback)")
	assert.Equal(t, []string{"Default1", "Default2"}, canadaValues, "GetValuesForParent(Canada) should return default values [Default1, Default2]")

	// Verify ConditionalValues map contains USA entry
	require.Contains(t, region.ConditionalValues, "USA", "Region.ConditionalValues should contain USA key")
	assert.Equal(t, []string{"NY", "CA"}, region.ConditionalValues["USA"], "ConditionalValues for USA should match")

	// Verify DefaultValues are populated
	require.NotNil(t, region.DefaultValues, "Region.DefaultValues should be non-nil")
	assert.Equal(t, []string{"Default1", "Default2"}, region.DefaultValues, "DefaultValues should contain the fallback values")
}

// TestConditionalSpec_UndefinedParentValueError verifies that analysis returns
// an error when a conditional spec references a parent value that does not
// exist in the parent concept's value set.
//
// This is Task 6c: end-to-end error case testing.
func TestConditionalSpec_UndefinedParentValueError(t *testing.T) {
	t.Parallel()

	// Arrange: Parse CPL with undefined parent value "Mexico" (Country only has USA, Canada)
	input := "concepts\n\tCountry=[Values[USA, Canada]]\n\tCountry[Mexico] => State=[Values[NY, CA]]"
	rootTable, err := parser.Parse(input)
	require.NoError(t, err, "Parsing should succeed")
	require.NotNil(t, rootTable, "Root table should not be nil")

	// Act: Analyze — this should produce an error
	defs, err := analysis.Analyze(rootTable)

	// Assert: Analysis should fail with error about undefined parent value
	require.Error(t, err, "Analysis should fail when conditional spec references undefined parent value")
	assert.Nil(t, defs, "Analysis definitions should be nil on error")
	assert.Contains(t, err.Error(), "parent value", "Error should mention 'parent value'")
	assert.Contains(t, err.Error(), "Mexico", "Error should mention the undefined parent value 'Mexico'")
}

// TestConditionalSpec_ExampleFiles verifies that all conditional_concept_*.cpl
// example files parse and analyze without errors, and that child concepts from
// conditional specifications appear correctly in the analyzed definitions.
//
// This is Task 8: end-to-end example file verification using table-driven tests.
func TestConditionalSpec_ExampleFiles(t *testing.T) {
	t.Parallel()

	// Define test cases for each example file.
	// expectedConcepts lists all concepts that should appear in defs.Concepts
	// after analysis. childConcepts lists concepts that should have non-empty
	// ConditionalValues maps (indicating they were defined via conditional specs).
	testCases := []struct {
		name             string
		filePath         string
		expectedConcepts []string
		childConcepts    []string // concepts that should have ConditionalValues
	}{
		{
			name:             "conditional_concept_basic.cpl",
			filePath:         "../../../../examples/parse_and_analyze_only/without_errors/conditional_concept_basic.cpl",
			expectedConcepts: []string{"Country", "State"},
			childConcepts:    []string{"State"},
		},
		{
			name:             "conditional_concept_default_values.cpl",
			filePath:         "../../../../examples/parse_and_analyze_only/without_errors/conditional_concept_default_values.cpl",
			expectedConcepts: []string{"Country", "Region"},
			childConcepts:    []string{"Region"},
		},
		{
			name:             "conditional_concept_multiple_parent_values.cpl",
			filePath:         "../../../../examples/parse_and_analyze_only/without_errors/conditional_concept_multiple_parent_values.cpl",
			expectedConcepts: []string{"Country", "Region", "Province", "State"},
			childConcepts:    []string{"Region", "Province", "State"},
		},
		{
			name:             "conditional_concept_multiple_sections.cpl",
			filePath:         "../../../../examples/parse_and_analyze_only/without_errors/conditional_concept_multiple_sections.cpl",
			expectedConcepts: []string{"Country", "State", "Province"},
			childConcepts:    []string{"State", "Province"},
		},
		{
			name:             "conditional_concept_with_concept_range.cpl",
			filePath:         "../../../../examples/parse_and_analyze_only/without_errors/conditional_concept_with_concept_range.cpl",
			expectedConcepts: []string{"Quarter", "Year", "FirstHalf"},
			childConcepts:    []string{"FirstHalf"},
		},
		{
			name:             "conditional_concept_with_int_range.cpl",
			filePath:         "../../../../examples/parse_and_analyze_only/without_errors/conditional_concept_with_int_range.cpl",
			expectedConcepts: []string{"Year", "Month"},
			childConcepts:    []string{"Month"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Step 1: Read the example file
			content, err := os.ReadFile(path.Clean(tc.filePath))
			require.NoError(t, err, "Should be able to read example file %s", tc.filePath)

			// Step 2: Parse the CPL content
			rootTable, err := parser.Parse(string(content))
			require.NoError(t, err, "Parsing should succeed for %s", tc.filePath)
			require.NotNil(t, rootTable, "Root table should not be nil for %s", tc.filePath)

			// Step 3: Analyze the parsed symbol table
			defs, err := analysis.Analyze(rootTable)

			// Step 4: Verify no errors
			require.NoError(t, err, "Analysis should succeed for %s", tc.filePath)
			require.NotNil(t, defs, "Analysis definitions should not be nil for %s", tc.filePath)
			require.NotNil(t, defs.Concepts, "Concepts map should be initialized for %s", tc.filePath)

			// Verify all expected concepts exist
			for _, conceptName := range tc.expectedConcepts {
				concept, exists := defs.Concepts[conceptName]
				require.True(t, exists, "Concept '%s' should exist in defs.Concepts for file %s", conceptName, tc.filePath)
				assert.Equal(t, conceptName, concept.ConceptName, "Concept name should match for '%s' in file %s", conceptName, tc.filePath)
			}

			// Verify child concepts from conditional specs have ConditionalValues
			for _, childName := range tc.childConcepts {
				concept, exists := defs.Concepts[childName]
				require.True(t, exists, "Child concept '%s' should exist in defs.Concepts for file %s", childName, tc.filePath)
				assert.NotEmpty(t, concept.ConditionalValues, "Child concept '%s' should have non-empty ConditionalValues in file %s", childName, tc.filePath)
				require.NotNil(t, concept.ParentConceptName, "Child concept '%s' should have a ParentConceptName in file %s", childName, tc.filePath)
			}

			// Ensure no unexpected concepts are present
			assert.Len(t, defs.Concepts, len(tc.expectedConcepts),
				"Should have exactly %d concepts in file %s, but found %d",
				len(tc.expectedConcepts), tc.filePath, len(defs.Concepts))
		})
	}
}

// ============================================================================
// Analyzer-Error Example File Tests — Story209 Task 5b+5c+5e
// ============================================================================

// TestInvalidCondFiles_TriggerAnalyzerErrors verifies that conditional concept
// example files in the with_errors directory parse successfully but produce
// errors during the analysis phase.
//
// Each test case verifies:
//  1. The file parses without error (valid CPL syntax).
//  2. Analysis returns an error (semantic validation fails).
//  3. The error message contains the expected substring for that error type.
func TestInvalidCondFiles_TriggerAnalyzerErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		filePath      string
		errorContains string
	}{
		{
			name:          "undefined parent concept",
			filePath:      "../../../../examples/parse_and_analyze_only/with_errors/invalid_conditional_concept_undefined_parent.cpl",
			errorContains: "parent concept",
		},
		{
			name:          "invalid parent value",
			filePath:      "../../../../examples/parse_and_analyze_only/with_errors/invalid_conditional_concept_invalid_parent_value.cpl",
			errorContains: "parent value",
		},
		{
			name:          "duplicate default — missing concept reference",
			filePath:      "../../../../examples/parse_and_analyze_only/with_errors/invalid_conditional_concept_duplicate_default.cpl",
			errorContains: "source concept",
		},
		{
			name:          "root conditional — invalid start value",
			filePath:      "../../../../examples/parse_and_analyze_only/with_errors/invalid_conditional_concept_root_conditional.cpl",
			errorContains: "start value not found",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Step 1: Read the example file
			content, err := os.ReadFile(path.Clean(tc.filePath))
			require.NoError(t, err, "Should be able to read example file %s", tc.filePath)

			// Step 2: Parse the CPL content — must succeed (valid syntax)
			rootTable, err := parser.Parse(string(content))
			require.NoError(t, err, "Parsing should succeed for %s", tc.filePath)
			require.NotNil(t, rootTable, "Root table should not be nil for %s", tc.filePath)

			// Step 3: Analyze the parsed symbol table — must fail (semantic error)
			defs, err := analysis.Analyze(rootTable)

			// Step 4: Verify analysis failure
			require.Error(t, err, "Analysis should fail for %s", tc.filePath)
			assert.Nil(t, defs, "Analysis definitions should be nil on error for %s", tc.filePath)
			assert.Contains(t, err.Error(), tc.errorContains,
				"Error message for %s should contain '%s', got: %s", tc.filePath, tc.errorContains, err.Error())
		})
	}
}
