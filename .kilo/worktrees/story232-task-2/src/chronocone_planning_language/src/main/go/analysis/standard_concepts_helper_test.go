// Package analysis provides test helpers for loading standard CPL concepts.
//
// This file contains helpers that load standard time concepts from the
// standard_concepts.cpl file instead of using hardcoded Go values.
// These helpers are usable from both internal tests (package analysis)
// and external tests (package analysis_test).
package analysis

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loadStandardTimeConceptsFromFile loads standard time concepts from the
// standard_concepts.cpl file located in the standard_definitions directory.
//
// It locates the CPL file relative to this test file using runtime.Caller,
// parses the file content with parser.Parse(), and analyzes the resulting
// symbol table with analysis.Analyze() to produce a concepts map.
//
// Returns:
//   - A map from concept name to AnalysisConceptDefinition containing the
//     parsed concept definitions from the CPL file.
//   - An error if the file cannot be read, parsed, or analyzed.
func loadStandardTimeConceptsFromFile() (map[string]AnalysisConceptDefinition, error) {
	_, testFile, _, _ := runtime.Caller(0)
	// From src/main/go/analysis/ go up 4 levels to reach the module root,
	// then into standard_definitions/standard_concepts.cpl.
	projectRoot := filepath.Join(filepath.Dir(testFile), "..", "..", "..", "..")
	cplPath := filepath.Join(projectRoot, "standard_definitions", "standard_concepts.cpl")

	content, err := os.ReadFile(cplPath)
	if err != nil {
		return nil, err
	}

	doc, err := parser.Parse(string(content))
	if err != nil {
		return nil, err
	}

	defs, err := Analyze(doc)
	if err != nil {
		return nil, err
	}

	return defs.Concepts, nil
}

// LoadStandardTimeConcepts loads standard time concepts from the
// standard_concepts.cpl file for use in external tests.
//
// This helper is exported for use by tests in the analysis_test package.
// It calls loadStandardTimeConceptsFromFile() and fails the test if
// loading or parsing encounters an error.
//
// Parameters:
//   - t: The testing.T instance used to report failures.
//
// Returns:
//   - A map from concept name to AnalysisConceptDefinition containing the
//     parsed concept definitions from the CPL file.
func LoadStandardTimeConcepts(t *testing.T) map[string]AnalysisConceptDefinition {
	t.Helper()
	concepts, err := loadStandardTimeConceptsFromFile()
	if err != nil {
		t.Fatalf("Failed to load standard time concepts: %v", err)
	}
	return concepts
}

// TestLoadStandardTimeConcepts verifies that loadStandardTimeConceptsFromFile
// successfully loads all standard time concepts from the CPL file.
func TestLoadStandardTimeConcepts(t *testing.T) {
	t.Parallel()

	concepts, err := loadStandardTimeConceptsFromFile()
	require.NoError(t, err, "loadStandardTimeConceptsFromFile should not error")
	require.NotNil(t, concepts, "concepts map should not be nil")

	// Verify all expected time-related concepts are present.
	expectedConcepts := []string{
		"Year", "DayOfYear", "FiscalQuarter", "WeekNumber",
		"DayOfWeek", "Month", "DayOfMonth", "Hour", "Minute", "Second",
	}
	for _, name := range expectedConcepts {
		name := name
		t.Run("contains "+name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, concepts, name, "Expected concept %s to be loaded", name)
		})
	}

	// Verify Month is a Values concept with the correct values.
	t.Run("Month values", func(t *testing.T) {
		t.Parallel()
		monthConcept, ok := concepts["Month"]
		require.True(t, ok, "Month concept should exist")
		expectedMonths := []string{
			"January", "February", "March", "April", "May", "June",
			"July", "August", "September", "October", "November", "December",
		}
		for _, month := range expectedMonths {
			assert.True(t, monthConcept.ContainsValue(month), "Month should contain %s", month)
		}
	})

	// Verify DayOfWeek is a Values concept with the correct values.
	t.Run("DayOfWeek values", func(t *testing.T) {
		t.Parallel()
		dowConcept, ok := concepts["DayOfWeek"]
		require.True(t, ok, "DayOfWeek concept should exist")
		expectedDays := []string{
			"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday",
		}
		for _, day := range expectedDays {
			assert.True(t, dowConcept.ContainsValue(day), "DayOfWeek should contain %s", day)
		}
	})

	// Verify Year IntRange bounds.
	t.Run("Year bounds", func(t *testing.T) {
		t.Parallel()
		yearConcept, ok := concepts["Year"]
		require.True(t, ok, "Year concept should exist")
		start, end, hasBounds := yearConcept.GetIntRangeBounds()
		require.True(t, hasBounds, "Year should have Int range bounds")
		assert.Equal(t, 1900, start, "Year start bound")
		assert.Equal(t, 3000, end, "Year end bound")
	})

	// Verify DayOfMonth IntRange bounds.
	t.Run("DayOfMonth bounds", func(t *testing.T) {
		t.Parallel()
		domConcept, ok := concepts["DayOfMonth"]
		require.True(t, ok, "DayOfMonth concept should exist")
		start, end, hasBounds := domConcept.GetIntRangeBounds()
		require.True(t, hasBounds, "DayOfMonth should have Int range bounds")
		assert.Equal(t, 1, start, "DayOfMonth start bound")
		assert.Equal(t, 31, end, "DayOfMonth end bound")
	})

	// Verify DayOfYear IntRange bounds.
	t.Run("DayOfYear bounds", func(t *testing.T) {
		t.Parallel()
		doyConcept, ok := concepts["DayOfYear"]
		require.True(t, ok, "DayOfYear concept should exist")
		start, end, hasBounds := doyConcept.GetIntRangeBounds()
		require.True(t, hasBounds, "DayOfYear should have Int range bounds")
		assert.Equal(t, 1, start, "DayOfYear start bound")
		assert.Equal(t, 366, end, "DayOfYear end bound")
	})

	// Verify WeekNumber IntRange bounds.
	t.Run("WeekNumber bounds", func(t *testing.T) {
		t.Parallel()
		wnConcept, ok := concepts["WeekNumber"]
		require.True(t, ok, "WeekNumber concept should exist")
		start, end, hasBounds := wnConcept.GetIntRangeBounds()
		require.True(t, hasBounds, "WeekNumber should have Int range bounds")
		assert.Equal(t, 1, start, "WeekNumber start bound")
		assert.Equal(t, 53, end, "WeekNumber end bound")
	})

	// Verify FiscalQuarter IntRange bounds.
	t.Run("FiscalQuarter bounds", func(t *testing.T) {
		t.Parallel()
		fqConcept, ok := concepts["FiscalQuarter"]
		require.True(t, ok, "FiscalQuarter concept should exist")
		start, end, hasBounds := fqConcept.GetIntRangeBounds()
		require.True(t, hasBounds, "FiscalQuarter should have Int range bounds")
		assert.Equal(t, 1, start, "FiscalQuarter start bound")
		assert.Equal(t, 4, end, "FiscalQuarter end bound")
	})

	// Verify Hour IntRange bounds from CPL file (0-based).
	t.Run("Hour bounds", func(t *testing.T) {
		t.Parallel()
		hourConcept, ok := concepts["Hour"]
		require.True(t, ok, "Hour concept should exist")
		start, end, hasBounds := hourConcept.GetIntRangeBounds()
		require.True(t, hasBounds, "Hour should have Int range bounds")
		assert.Equal(t, 0, start, "Hour start bound")
		assert.Equal(t, 23, end, "Hour end bound")
	})

	// Verify Minute IntRange bounds from CPL file (0-based).
	t.Run("Minute bounds", func(t *testing.T) {
		t.Parallel()
		minuteConcept, ok := concepts["Minute"]
		require.True(t, ok, "Minute concept should exist")
		start, end, hasBounds := minuteConcept.GetIntRangeBounds()
		require.True(t, hasBounds, "Minute should have Int range bounds")
		assert.Equal(t, 0, start, "Minute start bound")
		assert.Equal(t, 59, end, "Minute end bound")
	})

	// Verify Second IntRange bounds from CPL file (0-based).
	t.Run("Second bounds", func(t *testing.T) {
		t.Parallel()
		secondConcept, ok := concepts["Second"]
		require.True(t, ok, "Second concept should exist")
		start, end, hasBounds := secondConcept.GetIntRangeBounds()
		require.True(t, hasBounds, "Second should have Int range bounds")
		assert.Equal(t, 0, start, "Second start bound")
		assert.Equal(t, 59, end, "Second end bound")
	})
}

// TestLoadStandardTimeConceptsExported verifies the exported helper
// function works correctly for external test packages.
func TestLoadStandardTimeConceptsExported(t *testing.T) {
	t.Parallel()

	// The exported helper should load concepts without failing the test.
	concepts := LoadStandardTimeConcepts(t)
	require.NotNil(t, concepts, "LoadStandardTimeConcepts should return a non-nil map")

	// Spot-check that key concepts are present.
	assert.Contains(t, concepts, "Year")
	assert.Contains(t, concepts, "Month")
	assert.Contains(t, concepts, "Hour")
}
