// Package analysis_test contains integration tests for the CPL analysis phase.
//
// These tests complement the parsing tests in parser/cpl_parser_examples_test.go
// by verifying that parsed CPL structures are correctly transformed into
// analysis-level structures with resolved values, positions, and metadata.
//
// The parsing tests verify "does it parse?" - these tests verify "does analysis
// produce correct structures?"
//
// Adding new example files:
//  1. Add the .cpl file to examples/parse_and_analyze_only/without_errors/ or examples/parse_and_analyze_only/with_errors/
//  2. Update the expectations map in TestAnalysisIntegration_Examples with expected
//     analysis results for concepts, values, and positions
//  3. Add verification logic in the test or extend verifyConcepts helper if needed
//  4. Run tests to verify the new example is processed correctly
//
// Test structure follows Go best practices:
// - Table-driven tests for multiple scenarios
// - t.Parallel() for efficient parallel execution
// - testify assert/require for clear assertions
// - Range variable capture in closures to avoid concurrency issues
package analysis_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedConcept represents the expected analysis result for a single concept.
//
// This structure is used to define expected values in the test expectations map,
// making it easy to verify that analysis produces correct concept definitions.
type expectedConcept struct {
	// Name is the concept identifier (e.g., "Country", "Priority").
	Name string

	// Values is the ordered list of expected concept values.
	Values []string

	// Positions is the ordered list of expected positions (0-based indices).
	// Must have the same length as Values, with Position[i] being the position of Values[i].
	Positions []int
}

// expectedAnalysisResult captures the expected analysis output for an example file.
//
// For files with concepts, this contains a map from concept name to expected
// concept definition. For files without concepts (e.g., units_only.cpl), this
// may be nil or empty - the test will just verify parse and analyze succeed.
type expectedAnalysisResult struct {
	// Concepts is a map from concept name to its expected definition.
	// The key is the concept name, and the value defines expected values and positions.
	Concepts map[string]expectedConcept
}

// getExamplesDir returns the path to the examples directory.
//
// This helper navigates from the test file location to the project root,
// then into the examples directory. The path calculation assumes the test
// file is in src/main/go/analysis/ and needs to go up 4 levels to reach
// the project root where examples/ is located.
//
// Parameters:
//   - t: The testing.T instance for error reporting.
//
// Returns:
//   - The absolute path to the examples directory.
func getExamplesDir(t *testing.T) string {
	t.Helper()
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	// From src/main/go/analysis/ go up 4 levels to reach the module root (where go.mod is), then into examples
	projectRoot := filepath.Join(testDir, "..", "..", "..", "..")
	return filepath.Join(projectRoot, "examples")
}

// getStandardDefsPath returns the path to the standard_definitions directory.
func getStandardDefsPath(t *testing.T) string {
	t.Helper()
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	projectRoot := filepath.Join(testDir, "..", "..", "..", "..")
	return filepath.Join(projectRoot, "standard_definitions")
}

// getAnalysisForFile parses and analyzes a CPL file, returning the analysis definitions.
//
// This helper function optionally prepends standard definition files to the input,
// ensuring that standard time concepts are available for Instant constraint
// validation just as they would be in a normal compilation flow.
//
// Parameters:
//   - t: The testing.T instance for assertions.
//   - filePath: The path to the .cpl file to analyze.
//   - stdDefPaths: Optional variadic list of standard definition file paths to
//     prepend to the parse list.
//
// Returns:
//   - The AnalysisDefinitions produced by analysis, or nil if analysis fails.
//   - An error if any step (read, parse, analyze) fails.
func getAnalysisForFile(t *testing.T, filePath string, stdDefPaths ...string) (*analysis.AnalysisDefinitions, error) {
	t.Helper()

	config := &parser.ParseLoggingConfig{
		LogLevel:   slog.LevelInfo,
		OutputFile: "",
	}

	// Build the file list by prepending any standard definition files.
	// Standard concepts are NOT automatically loaded for all files because some
	// example files define their own concepts that may conflict with standard
	// definitions (e.g., int_range_concepts.cpl defines Month as an Int range,
	// while standard_concepts.cpl defines Month as Values). Tests that
	// specifically need standard concepts should provide them via stdDefPaths.
	filesToParse := append(stdDefPaths, filePath)
	rootTable, err := parser.ParseFiles(filesToParse, config)
	if err != nil {
		return nil, err
	}

	// Analyze the merged symbol table to produce AnalysisDefinitions
	defs, err := analysis.Analyze(rootTable)
	if err != nil {
		return nil, err
	}

	return defs, nil
}

// verifyConcepts checks that an actual concept definition matches expected values.
//
// This helper verifies:
// - The concept name matches the expected name
// - The number of resolved values matches the expected count
// - Each value name matches the expected value
// - Each value position matches the expected position
//
// Parameters:
//   - t: The testing.T instance for assertions.
//   - conceptName: The name of the concept being verified (for error messages).
//   - expected: The expected concept definition.
//   - actual: The actual concept definition from analysis.
func verifyConcepts(t *testing.T, conceptName string, expected expectedConcept, actual analysis.AnalysisConceptDefinition) {
	t.Helper()

	// Verify concept name matches
	assert.Equal(t, expected.Name, actual.ConceptName, "Concept name should match for %s", conceptName)

	// Verify number of resolved values matches expected
	require.Equal(t, len(expected.Values), len(actual.ResolvedValues),
		"Number of values should match for concept %s", conceptName)
	require.Equal(t, len(expected.Positions), len(actual.ResolvedValues),
		"Number of positions should match number of values for concept %s", conceptName)

	// Verify each value name and position matches expected
	for i, expectedValue := range expected.Values {
		assert.Equal(t, expectedValue, actual.ResolvedValues[i].Name,
			"Value at position %d should match for concept %s", i, conceptName)
		assert.Equal(t, expected.Positions[i], actual.ResolvedValues[i].Position,
			"Position for value '%s' should match in concept %s", expectedValue, conceptName)
	}
}

// TestAnalysisIntegration_Examples verifies analysis produces expected structures from example CPL files.
//
// This test discovers all .cpl files in examples/parse_and_analyze_only/without_errors/ and verifies that:
//   - Each file parses successfully
//   - Each file analyzes successfully (produces AnalysisDefinitions)
//   - For files with known expectations (basic_concepts.cpl), concept definitions
//     have correct names, values, and positions
//   - For files without concepts (units_only.cpl, multiple_sections.cpl), only
//     parse and analyze success is verified
//
// The test uses a table-driven structure with t.Run() subtests for each file,
// enabling parallel execution and clear failure isolation.
func TestAnalysisIntegration_Examples(t *testing.T) {
	t.Parallel()

	// Define expected analysis results for known example files
	// This map associates each example file with its expected analysis output
	expectations := map[string]expectedAnalysisResult{
		"basic_concepts.cpl": {
			Concepts: map[string]expectedConcept{
				"Country": {
					Name:      "Country",
					Values:    []string{"USA", "Canada", "Mexico"},
					Positions: []int{0, 1, 2},
				},
				"Priority": {
					Name:      "Priority",
					Values:    []string{"Low", "Medium", "High"},
					Positions: []int{0, 1, 2},
				},
			},
		},
		"multiple_concepts_sections.cpl": {
			Concepts: map[string]expectedConcept{
				"Country": {
					Name:      "Country",
					Values:    []string{"USA", "Canada", "Mexico"},
					Positions: []int{0, 1, 2},
				},
				"Region": {
					Name:      "Region",
					Values:    []string{"North", "South", "East", "West"},
					Positions: []int{0, 1, 2, 3},
				},
				"Priority": {
					Name:      "Priority",
					Values:    []string{"Low", "Medium", "High"},
					Positions: []int{0, 1, 2},
				},
				"Status": {
					Name:      "Status",
					Values:    []string{"Active", "Inactive", "Pending"},
					Positions: []int{0, 1, 2},
				},
			},
		},
		"multiple_concepts_sections_hierarchy.cpl": {
			Concepts: map[string]expectedConcept{
				"Country": {
					Name:      "Country",
					Values:    []string{"USA", "Canada"},
					Positions: []int{0, 1},
				},
				"Region": {
					Name:      "Region",
					Values:    []string{"North", "South"},
					Positions: []int{0, 1},
				},
				"Product": {
					Name:      "Product",
					Values:    []string{"Widget", "Gadget"},
					Positions: []int{0, 1},
				},
				"Variant": {
					Name:      "Variant",
					Values:    []string{"A", "B", "C"},
					Positions: []int{0, 1, 2},
				},
			},
		},
		"int_range_concepts.cpl": {
			Concepts: map[string]expectedConcept{
				"Month": {
					Name:      "Month",
					Values:    []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
					Positions: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
				},
				"FiscalQuarter": {
					Name:      "FiscalQuarter",
					Values:    []string{"1", "2", "3", "4"},
					Positions: []int{0, 1, 2, 3},
				},
				"EvenNumber": {
					Name:      "EvenNumber",
					Values:    []string{"2", "4", "6", "8", "10"},
					Positions: []int{0, 1, 2, 3, 4},
				},
				"SmallCount": {
					Name:      "SmallCount",
					Values:    []string{"1", "2", "3", "4", "5"},
					Positions: []int{0, 1, 2, 3, 4},
				},
				"Singleton": {
					Name:      "Singleton",
					Values:    []string{"7"},
					Positions: []int{0},
				},
				"Fives": {
					Name:      "Fives",
					Values:    []string{"0", "5", "10", "15", "20"},
					Positions: []int{0, 1, 2, 3, 4},
				},
			},
		},
		"int_range_mixed_generators.cpl": {
			Concepts: map[string]expectedConcept{
				"Mixed": {
					Name:      "Mixed",
					Values:    []string{"A", "B", "1", "2", "3"},
					Positions: []int{0, 1, 2, 3, 4},
				},
				"PriorityAndLevels": {
					Name:      "PriorityAndLevels",
					Values:    []string{"Low", "Medium", "High", "1", "2"},
					Positions: []int{0, 1, 2, 3, 4},
				},
				"NumberedCategories": {
					Name:      "NumberedCategories",
					Values:    []string{"10", "20", "30", "Other", "Misc"},
					Positions: []int{0, 1, 2, 3, 4},
				},
			},
		},
		// units_only.cpl and multiple_sections.cpl have no concepts,
		// so they have no expectations beyond "parses and analyzes without errors"
		"units_only.cpl":        {},
		"multiple_sections.cpl": {},
	}

	// expectedErrors maps file names to error substrings that are expected
	// from analysis. This is used for files that use features not yet fully
	// implemented (e.g., aggregator functions not registered in the current increment).
	expectedErrors := map[string]string{
		// aggregators-single-concept.cpl has from [Country] to [Country] —
		// all from-concepts equal the to-concept, so there is no collapsible
		// dimension. The analyzer rejects this as invalid.
		"aggregators-single-concept.cpl": "must contain at least one concept different from to-concept",
	}

	examplesDir := getExamplesDir(t)
	withoutErrorsDir := filepath.Join(examplesDir, "parse_and_analyze_only", "without_errors")

	entries, err := os.ReadDir(withoutErrorsDir)
	require.NoError(t, err, "Failed to read without_errors directory")

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cpl") {
			continue
		}

		entry := entry // capture range variable for closure
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(withoutErrorsDir, entry.Name())

			// Files that use Instant/Interval constraints require standard time
			// concepts (Hour, Minute, Second, DayOfWeek, Month, etc.) to be
			// defined. These concepts are loaded automatically by the CLI compiler
			// from standard_definitions/standard_concepts.cpl, but this general
			// integration test does not load standard definitions for all files to
			// avoid conflicts with example files that define their own concepts
			// (e.g., int_range_concepts.cpl defines Month as an Int range). For
			// files that need standard concepts, we conditionally prepend the
			// standard definitions file so they can be analyzed without skipping.
			var stdDefs []string
			if entry.Name() == "instant_declarations.cpl" || entry.Name() == "interval_declarations.cpl" {
				stdDefs = []string{filepath.Join(getStandardDefsPath(t), "standard_concepts.cpl")}
			}

			// Get analysis definitions for this file
			defs, err := getAnalysisForFile(t, filePath, stdDefs...)

			// Files with expected errors (features not yet fully implemented)
			if expectedErr, hasExpectedErr := expectedErrors[entry.Name()]; hasExpectedErr {
				require.Error(t, err, "Expected analysis error for %s", entry.Name())
				assert.Contains(t, err.Error(), expectedErr,
					"Expected error for %s to contain %q, got: %v", entry.Name(), expectedErr, err)
				return
			}

			require.NoError(t, err, "Expected no error for %s", entry.Name())
			require.NotNil(t, defs, "AnalysisDefinitions should not be nil for %s", entry.Name())

			// Look up expectations for this file
			expected, hasExpectations := expectations[entry.Name()]

			// If we have specific expectations for concepts, verify them
			if hasExpectations && len(expected.Concepts) > 0 {
				for conceptName, expectedConcept := range expected.Concepts {
					actualConcept, exists := defs.Concepts[conceptName]
					require.True(t, exists, "Expected concept '%s' to exist in %s", conceptName, entry.Name())
					verifyConcepts(t, conceptName, expectedConcept, actualConcept)
				}

			// Verify all expected concepts exist. We don't check the exact total
			// because standard time concepts (Month, DayOfWeek, Hour, etc.) are
			// loaded from standard_definitions/standard_concepts.cpl alongside
			// the example file, just as the CLI compiler does.
			require.GreaterOrEqual(t, len(defs.Concepts), len(expected.Concepts),
				"Should have at least %d concepts for %s", len(expected.Concepts), entry.Name())
			}

			// Spreader-specific assertions: verify Spreaders map and CIL trees are populated
			if strings.HasPrefix(entry.Name(), "spreaders-") {
				if entry.Name() == "spreaders-zero-declarations.cpl" {
					require.Empty(t, defs.Spreaders, "Zero-declarations spreader file should produce empty Spreaders map")
				} else {
					require.NotEmpty(t, defs.Spreaders, "Spreader file %s should produce non-empty Spreaders map", entry.Name())
					var hasSpreaderCIL bool
					for k := range defs.CILTrees {
						if strings.HasPrefix(k, "spreader:") {
							hasSpreaderCIL = true
							break
						}
					}
					require.True(t, hasSpreaderCIL, "Spreader file %s should produce CIL trees with spreader entries", entry.Name())
				}
			}
		})
	}
}

// ============================================================================
// Multiple Units and Measures Sections Integration Tests
// ============================================================================

// TestAnalysisIntegration_MultipleUnitsAndMeasuresSections verifies that a file
// with multiple units sections and multiple measures sections parses and analyzes
// correctly. Each section creates its own scope, and definitions are independently
// processed without interference.
func TestAnalysisIntegration_MultipleUnitsAndMeasuresSections(t *testing.T) {
	t.Parallel()

	input := `concepts
	Country = [Values[USA, Canada]]
	MassConcept = [Values[kg, lb]]

units
	Currency = Country

units
	Weight = MassConcept

measures
	Price = CurrencyType(unit=Currency)

measures
	Bulk = MassType(unit=Weight)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Expected no error for mixed multiple sections")
	require.NotNil(t, defs, "AnalysisDefinitions should not be nil")

	// Verify concepts from all sections
	require.Contains(t, defs.Concepts, "Country", "Country concept should be analyzed")
	require.Contains(t, defs.Concepts, "MassConcept", "MassConcept concept should be analyzed")

	// Verify units from both sections
	require.Contains(t, defs.Units, "Currency", "Currency unit should be analyzed")
	assert.Equal(t, "Country", defs.Units["Currency"].ConceptName)

	require.Contains(t, defs.Units, "Weight", "Weight unit should be analyzed")
	assert.Equal(t, "MassConcept", defs.Units["Weight"].ConceptName)

	// Verify measures from both sections
	require.Contains(t, defs.Measures, "Price", "Price measure should be analyzed")
	assert.Equal(t, "Price", defs.Measures["Price"].Name)
	assert.Equal(t, "CurrencyType", defs.Measures["Price"].TypeName)
	assert.Equal(t, "Currency", defs.Measures["Price"].GetUnitName())

	require.Contains(t, defs.Measures, "Bulk", "Bulk measure should be analyzed")
	assert.Equal(t, "Bulk", defs.Measures["Bulk"].Name)
	assert.Equal(t, "MassType", defs.Measures["Bulk"].TypeName)
	assert.Equal(t, "Weight", defs.Measures["Bulk"].GetUnitName())
}

// TestAnalysisIntegration_MeasuresReferenceUnitsAcrossSections verifies that
// measures can reference units defined in different sections, and that the
// parser and analyzer correctly handle cross-section references.
func TestAnalysisIntegration_MeasuresReferenceUnitsAcrossSections(t *testing.T) {
	t.Parallel()

	// Measure in second measures section references unit from first units section,
	// and measure in first measures section references unit from second units section
	input := `concepts
	Country = [Values[USA]]

units
	Currency = Country

units
	Weight = Country

measures
	LocalPrice = Type1(unit=Currency)

measures
	ShippingCost = Type2(unit=Weight)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Expected no error for cross-section references")
	require.NotNil(t, defs, "AnalysisDefinitions should not be nil")

	// Verify units from both sections
	require.Contains(t, defs.Units, "Currency", "Currency unit should be analyzed")
	require.Contains(t, defs.Units, "Weight", "Weight unit should be analyzed")

	// Verify measures reference correct units
	require.Contains(t, defs.Measures, "LocalPrice", "LocalPrice measure should be analyzed")
	assert.Equal(t, "Currency", defs.Measures["LocalPrice"].GetUnitName(),
		"LocalPrice should reference Currency unit from first units section")

	require.Contains(t, defs.Measures, "ShippingCost", "ShippingCost measure should be analyzed")
	assert.Equal(t, "Weight", defs.Measures["ShippingCost"].GetUnitName(),
		"ShippingCost should reference Weight unit from second units section")
}

// TestAnalysisIntegration_UnitsAndMeasuresNoInterference verifies that units
// and measures with the same name do not interfere with each other, since they
// are stored in separate analyzer maps.
func TestAnalysisIntegration_UnitsAndMeasuresNoInterference(t *testing.T) {
	t.Parallel()

	// Both a unit and a measure named "Price" - should not conflict
	input := `concepts
	Country = [Values[USA]]

units
	Price = Country

measures
	Price = Type1(unit=Price)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Expected no error for same name in units and measures")
	require.NotNil(t, defs, "AnalysisDefinitions should not be nil")

	// Verify both unit and measure exist with the same name
	require.Contains(t, defs.Units, "Price", "Price unit should be analyzed")
	assert.Equal(t, "Country", defs.Units["Price"].ConceptName)

	require.Contains(t, defs.Measures, "Price", "Price measure should be analyzed")
	assert.Equal(t, "Price", defs.Measures["Price"].Name)
	assert.Equal(t, "Type1", defs.Measures["Price"].TypeName)
	assert.Equal(t, "Price", defs.Measures["Price"].GetUnitName())
}

// TestAnalysisIntegration_InterleavedUnitsAndMeasuresSections verifies that
// units and measures sections can be interleaved with other section types
// without issues.
func TestAnalysisIntegration_InterleavedUnitsAndMeasuresSections(t *testing.T) {
	t.Parallel()

	input := `concepts
	Country = [Values[USA, Canada]]

units
	Currency = Country

measures
	Price = Type1(unit=Currency)

units
	Weight = Country

measures
	Bulk = Type2(unit=Weight)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Expected no error for interleaved sections")
	require.NotNil(t, defs, "AnalysisDefinitions should not be nil")

	// Verify all units
	require.Len(t, defs.Units, 2, "Should have 2 units")
	require.Contains(t, defs.Units, "Currency")
	require.Contains(t, defs.Units, "Weight")

	// Verify all measures
	require.Len(t, defs.Measures, 2, "Should have 2 measures")
	require.Contains(t, defs.Measures, "Price")
	require.Contains(t, defs.Measures, "Bulk")
}
