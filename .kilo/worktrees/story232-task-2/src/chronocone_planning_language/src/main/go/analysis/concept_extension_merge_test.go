// Package analysis_test contains tests for the CPL analysis phase.
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/ast"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeExtensionConcept creates a concept definition representing a concept extension
// with the specified name, insertion point, and new values.
func makeExtensionConcept(name string, ip ast.InsertionPoint, values []string) *analysis.AnalysisConceptDefinition {
	gen := &analysis.ValuesGenerator{Values: values}
	concept := analysis.NewValuesConceptDefinition(name, values)
	concept.InsertionPoint = ip
	concept.Generators = []analysis.Generator{gen}
	return &concept
}

// makeInitialConcept creates a concept definition representing an initial (non-extension)
// concept with the specified name and values.
func makeInitialConcept(name string, values []string) *analysis.AnalysisConceptDefinition {
	concept := analysis.NewValuesConceptDefinition(name, values)
	return &concept
}

// TestMergeConceptValues_BeforeInsertion verifies that merging an extension
// with a "before" insertion point places the extension's values before the
// specified anchor value in the existing concept's ResolvedValues.
func TestMergeConceptValues_BeforeInsertion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		existingValues []string
		extensionIP    ast.InsertionPoint
		extensionVals  []string
		expectedValues []string
	}{
		{
			name:           "insert before first value",
			existingValues: []string{"Low", "Medium", "High"},
			extensionIP:    ast.InsertionPoint{BeforeValue: "Low"},
			extensionVals:  []string{"Critical"},
			expectedValues: []string{"Critical", "Low", "Medium", "High"},
		},
		{
			name:           "insert before middle value",
			existingValues: []string{"Low", "Medium", "High"},
			extensionIP:    ast.InsertionPoint{BeforeValue: "Medium"},
			extensionVals:  []string{"Normal"},
			expectedValues: []string{"Low", "Normal", "Medium", "High"},
		},
		{
			name:           "insert before last value",
			existingValues: []string{"A", "B", "C"},
			extensionIP:    ast.InsertionPoint{BeforeValue: "C"},
			extensionVals:  []string{"X"},
			expectedValues: []string{"A", "B", "X", "C"},
		},
		{
			name:           "insert multiple values before anchor",
			existingValues: []string{"A", "B", "C"},
			extensionIP:    ast.InsertionPoint{BeforeValue: "B"},
			extensionVals:  []string{"X", "Y"},
			expectedValues: []string{"A", "X", "Y", "B", "C"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			analyzer := analysis.NewCPLAnalyzer()
			existing := makeInitialConcept("Priority", tc.existingValues)
			analyzer.AddConcept(existing)

			extension := makeExtensionConcept("Priority", tc.extensionIP, tc.extensionVals)
			err := analyzer.AddOrMergeConcept(extension)

			require.NoError(t, err, "AddOrMergeConcept should not error for valid before insertion")
			merged := analyzer.GetConcept("Priority")
			require.Len(t, merged.ResolvedValues, len(tc.expectedValues), "merged values length should match expected")
			for i, expected := range tc.expectedValues {
				assert.Equal(t, expected, merged.ResolvedValues[i].Name, "value at position %d", i)
				assert.Equal(t, i, merged.ResolvedValues[i].Position, "Position at index %d", i)
			}
		})
	}
}

// TestMergeConceptValues_AfterInsertion verifies that merging an extension
// with an "after" insertion point places the extension's values after the
// specified anchor value in the existing concept's ResolvedValues.
func TestMergeConceptValues_AfterInsertion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		existingValues []string
		extensionIP    ast.InsertionPoint
		extensionVals  []string
		expectedValues []string
	}{
		{
			name:           "insert after first value",
			existingValues: []string{"Low", "Medium", "High"},
			extensionIP:    ast.InsertionPoint{AfterValue: "Low"},
			extensionVals:  []string{"Urgent"},
			expectedValues: []string{"Low", "Urgent", "Medium", "High"},
		},
		{
			name:           "insert after middle value",
			existingValues: []string{"Low", "Medium", "High"},
			extensionIP:    ast.InsertionPoint{AfterValue: "Medium"},
			extensionVals:  []string{"Normal"},
			expectedValues: []string{"Low", "Medium", "Normal", "High"},
		},
		{
			name:           "insert after last value",
			existingValues: []string{"Low", "Medium", "High"},
			extensionIP:    ast.InsertionPoint{AfterValue: "High"},
			extensionVals:  []string{"Critical"},
			expectedValues: []string{"Low", "Medium", "High", "Critical"},
		},
		{
			name:           "insert multiple values after anchor",
			existingValues: []string{"A", "B", "C"},
			extensionIP:    ast.InsertionPoint{AfterValue: "A"},
			extensionVals:  []string{"X", "Y"},
			expectedValues: []string{"A", "X", "Y", "B", "C"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			analyzer := analysis.NewCPLAnalyzer()
			existing := makeInitialConcept("Priority", tc.existingValues)
			analyzer.AddConcept(existing)

			extension := makeExtensionConcept("Priority", tc.extensionIP, tc.extensionVals)
			err := analyzer.AddOrMergeConcept(extension)

			require.NoError(t, err, "AddOrMergeConcept should not error for valid after insertion")
			merged := analyzer.GetConcept("Priority")
			require.Len(t, merged.ResolvedValues, len(tc.expectedValues), "merged values length should match expected")
			for i, expected := range tc.expectedValues {
				assert.Equal(t, expected, merged.ResolvedValues[i].Name, "value at position %d", i)
				assert.Equal(t, i, merged.ResolvedValues[i].Position, "Position at index %d", i)
			}
		})
	}
}

// TestAddOrMergeConcept_AddNewConcept verifies that calling AddOrMergeConcept
// with a concept that has no existing definition adds it as a new concept.
func TestAddOrMergeConcept_AddNewConcept(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	concept := makeInitialConcept("Country", []string{"USA", "Canada", "Mexico"})
	err := analyzer.AddOrMergeConcept(concept)

	require.NoError(t, err, "AddOrMergeConcept should not error for new concept")
	retrieved := analyzer.GetConcept("Country")
	assert.Equal(t, "Country", retrieved.ConceptName, "concept name should match")
	assert.Len(t, retrieved.ResolvedValues, 3, "should have 3 values")
}

// TestAddOrMergeConcept_MultipleExtensions verifies that multiple extensions
// on the same concept are processed in order, with each extension's values
// inserted at the correct position relative to the current merged state.
func TestAddOrMergeConcept_MultipleExtensions(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()

	// Initial definition: Priority = [Low, Medium, High]
	initial := makeInitialConcept("Priority", []string{"Low", "Medium", "High"})
	err := analyzer.AddOrMergeConcept(initial)
	require.NoError(t, err)

	// Extension 1: Priority = [Critical] before Low → [Critical, Low, Medium, High]
	ext1 := makeExtensionConcept("Priority", ast.InsertionPoint{BeforeValue: "Low"}, []string{"Critical"})
	err = analyzer.AddOrMergeConcept(ext1)
	require.NoError(t, err)

	merged1 := analyzer.GetConcept("Priority")
	require.Len(t, merged1.ResolvedValues, 4)
	assert.Equal(t, "Critical", merged1.ResolvedValues[0].Name)

	// Extension 2: Priority = [Urgent] after Medium → [Critical, Low, Medium, Urgent, High]
	ext2 := makeExtensionConcept("Priority", ast.InsertionPoint{AfterValue: "Medium"}, []string{"Urgent"})
	err = analyzer.AddOrMergeConcept(ext2)
	require.NoError(t, err)

	merged2 := analyzer.GetConcept("Priority")
	expected := []string{"Critical", "Low", "Medium", "Urgent", "High"}
	require.Len(t, merged2.ResolvedValues, len(expected))
	for i, exp := range expected {
		assert.Equal(t, exp, merged2.ResolvedValues[i].Name, "value at position %d", i)
		assert.Equal(t, i, merged2.ResolvedValues[i].Position, "Position at index %d", i)
	}
}

// TestAddOrMergeConcept_DuplicateWithoutInsertionPoint verifies that
// calling AddOrMergeConcept with a concept that has the same name as an
// existing concept but a zero InsertionPoint (no before/after) returns
// a duplicate definition error.
func TestAddOrMergeConcept_DuplicateWithoutInsertionPoint(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	initial := makeInitialConcept("Country", []string{"USA", "Canada"})
	err := analyzer.AddOrMergeConcept(initial)
	require.NoError(t, err)

	// Second definition with same name but no insertion point → error
	duplicate := makeInitialConcept("Country", []string{"USA", "Canada", "Mexico"})
	err = analyzer.AddOrMergeConcept(duplicate)
	assert.Error(t, err, "should error on duplicate concept without insertion point")
}

// TestAddOrMergeConcept_ExtensionWithNoExistingConcept verifies that
// calling AddOrMergeConcept with an extension concept (non-zero InsertionPoint)
// when no existing concept of that name exists returns an error.
func TestAddOrMergeConcept_ExtensionWithNoExistingConcept(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		extIP   ast.InsertionPoint
		errText string
	}{
		{
			name:    "extension with before when no initial concept exists",
			extIP:   ast.InsertionPoint{BeforeValue: "Low"},
			errText: "extension with before insertion",
		},
		{
			name:    "extension with after when no initial concept exists",
			extIP:   ast.InsertionPoint{AfterValue: "Low"},
			errText: "extension with after insertion",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			analyzer := analysis.NewCPLAnalyzer()
			extension := makeExtensionConcept("NonExistent", tc.extIP, []string{"New"})
			err := analyzer.AddOrMergeConcept(extension)
			assert.Error(t, err, "should error when extension has no existing concept to extend")
		})
	}
}

// TestAddOrMergeConcept_InvalidInsertionPoint verifies that an extension
// with an invalid insertion point (both before and after set, or neither)
// returns an error.
func TestAddOrMergeConcept_InvalidInsertionPoint(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		extIP ast.InsertionPoint
	}{
		{
			name:  "both before and after set",
			extIP: ast.InsertionPoint{BeforeValue: "A", AfterValue: "B"},
		},
		{
			name:  "neither before nor after set (zero value on extension)",
			extIP: ast.InsertionPoint{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			analyzer := analysis.NewCPLAnalyzer()
			// Create the initial concept first
			initial := makeInitialConcept("Test", []string{"A", "B", "C"})
			analyzer.AddConcept(initial)

			extension := makeExtensionConcept("Test", tc.extIP, []string{"New"})
			err := analyzer.AddOrMergeConcept(extension)
			assert.Error(t, err, "should error on invalid insertion point")
		})
	}
}

// TestAddOrMergeConcept_BeforeValueNotFound verifies that an extension
// with a before/after anchor value that doesn't exist in the concept
// returns an error.
func TestAddOrMergeConcept_BeforeValueNotFound(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		existingVals  []string
		extensionIP   ast.InsertionPoint
		extensionVals []string
		desc          string
	}{
		{
			name:          "before non-existent value",
			existingVals:  []string{"Low", "Medium", "High"},
			extensionIP:   ast.InsertionPoint{BeforeValue: "NonExistent"},
			extensionVals: []string{"New"},
			desc:          "should error when before value not found",
		},
		{
			name:          "after non-existent value",
			existingVals:  []string{"Low", "Medium", "High"},
			extensionIP:   ast.InsertionPoint{AfterValue: "NonExistent"},
			extensionVals: []string{"New"},
			desc:          "should error when after value not found",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			analyzer := analysis.NewCPLAnalyzer()
			initial := makeInitialConcept("Priority", tc.existingVals)
			analyzer.AddConcept(initial)

			extension := makeExtensionConcept("Priority", tc.extensionIP, tc.extensionVals)
			err := analyzer.AddOrMergeConcept(extension)
			assert.Error(t, err, tc.desc)
		})
	}
}

// TestAddOrMergeConcept_OriginalOrderPreserved verifies that the relative
// order of original values is preserved when extensions are merged.
func TestAddOrMergeConcept_OriginalOrderPreserved(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		existingValues []string
		extensions     []struct {
			ip     ast.InsertionPoint
			values []string
		}
		expectedValues []string
	}{
		{
			name:           "preserve order with before and after insertions",
			existingValues: []string{"A", "B", "C", "D"},
			extensions: []struct {
				ip     ast.InsertionPoint
				values []string
			}{
				{ip: ast.InsertionPoint{BeforeValue: "A"}, values: []string{"X"}},
				{ip: ast.InsertionPoint{AfterValue: "D"}, values: []string{"Y"}},
			},
			expectedValues: []string{"X", "A", "B", "C", "D", "Y"},
		},
		{
			name:           "preserve order with multiple before insertions",
			existingValues: []string{"A", "B", "C"},
			extensions: []struct {
				ip     ast.InsertionPoint
				values []string
			}{
				{ip: ast.InsertionPoint{BeforeValue: "A"}, values: []string{"X"}},
				{ip: ast.InsertionPoint{BeforeValue: "C"}, values: []string{"Y"}},
			},
			expectedValues: []string{"X", "A", "B", "Y", "C"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			analyzer := analysis.NewCPLAnalyzer()
			initial := makeInitialConcept("Test", tc.existingValues)
			err := analyzer.AddOrMergeConcept(initial)
			require.NoError(t, err)

			for _, ext := range tc.extensions {
				extension := makeExtensionConcept("Test", ext.ip, ext.values)
				err = analyzer.AddOrMergeConcept(extension)
				require.NoError(t, err, "should not error on valid extension")
			}

			merged := analyzer.GetConcept("Test")
			require.Len(t, merged.ResolvedValues, len(tc.expectedValues))
			for i, exp := range tc.expectedValues {
				assert.Equal(t, exp, merged.ResolvedValues[i].Name, "value at position %d", i)
				assert.Equal(t, i, merged.ResolvedValues[i].Position, "Position at index %d", i)
			}
		})
	}
}

// TestAddOrMergeConcept_GeneratorsMerged verifies that the extension's
// generators are appended to the existing concept's Generators list.
func TestAddOrMergeConcept_GeneratorsMerged(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()

	// Initial concept with one generator
	initial := makeInitialConcept("Priority", []string{"Low", "Medium", "High"})
	initial.SourceLocation = symboltable.SourceLocation{Line: 1, Column: 1}
	err := analyzer.AddOrMergeConcept(initial)
	require.NoError(t, err)

	// Extension with before clause adds one more generator
	ext1 := makeExtensionConcept("Priority", ast.InsertionPoint{BeforeValue: "Low"}, []string{"Critical"})
	ext1.SourceLocation = symboltable.SourceLocation{Line: 5, Column: 1}
	err = analyzer.AddOrMergeConcept(ext1)
	require.NoError(t, err)

	merged := analyzer.GetConcept("Priority")
	// The initial generator + extension generator should both be present
	assert.GreaterOrEqual(t, len(merged.Generators), 2,
		"merged concept should have generators from both initial and extension definitions")
}

// TestMergeConceptValues_EmptyExtensionValues verifies that merging an extension
// with an empty values slice (valid anchor, no values) is a no-op.
func TestMergeConceptValues_EmptyExtensionValues(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	initial := makeInitialConcept("Test", []string{"A", "B", "C"})
	analyzer.AddConcept(initial)

	// Extension with valid anchor but empty values → should succeed as no-op
	extension := makeExtensionConcept("Test", ast.InsertionPoint{BeforeValue: "B"}, []string{})
	err := analyzer.AddOrMergeConcept(extension)
	require.NoError(t, err, "should not error on extension with empty values")

	merged := analyzer.GetConcept("Test")
	require.Len(t, merged.ResolvedValues, 3, "should still have original 3 values")
	assert.Equal(t, "A", merged.ResolvedValues[0].Name)
	assert.Equal(t, "B", merged.ResolvedValues[1].Name)
	assert.Equal(t, "C", merged.ResolvedValues[2].Name)
}

// TestAddOrMergeConcept_DuplicateInsertionAtSamePosition verifies that
// two extensions targeting the same anchor+direction pair return an error
// on the second extension to prevent ambiguous insertions.
func TestAddOrMergeConcept_DuplicateInsertionAtSamePosition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		ext1IP   ast.InsertionPoint
		ext1Vals []string
		ext2IP   ast.InsertionPoint
		ext2Vals []string
		desc     string
	}{
		{
			name:     "duplicate before same anchor",
			ext1IP:   ast.InsertionPoint{BeforeValue: "Low"},
			ext1Vals: []string{"Critical"},
			ext2IP:   ast.InsertionPoint{BeforeValue: "Low"},
			ext2Vals: []string{"Urgent"},
			desc:     "second 'before Low' should error",
		},
		{
			name:     "duplicate after same anchor",
			ext1IP:   ast.InsertionPoint{AfterValue: "Medium"},
			ext1Vals: []string{"Normal"},
			ext2IP:   ast.InsertionPoint{AfterValue: "Medium"},
			ext2Vals: []string{"Urgent"},
			desc:     "second 'after Medium' should error",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			analyzer := analysis.NewCPLAnalyzer()

			// Add initial concept
			initial := makeInitialConcept("Priority", []string{"Low", "Medium", "High"})
			err := analyzer.AddOrMergeConcept(initial)
			require.NoError(t, err)

			// First extension should succeed
			ext1 := makeExtensionConcept("Priority", tc.ext1IP, tc.ext1Vals)
			err = analyzer.AddOrMergeConcept(ext1)
			require.NoError(t, err, "first extension should succeed")

			// Second extension targeting same anchor+direction should error
			ext2 := makeExtensionConcept("Priority", tc.ext2IP, tc.ext2Vals)
			err = analyzer.AddOrMergeConcept(ext2)
			assert.Error(t, err, tc.desc)
			assert.Contains(t, err.Error(), "duplicate insertion at same position",
				"error should mention duplicate insertion")
		})
	}
}

// TestAddOrMergeConcept_DifferentDirectionSameAnchorAllowed verifies that
// two extensions targeting the same anchor value but with different
// directions (before vs after) are allowed.
func TestAddOrMergeConcept_DifferentDirectionSameAnchorAllowed(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()

	// Add initial concept
	initial := makeInitialConcept("Priority", []string{"Low", "Medium", "High"})
	err := analyzer.AddOrMergeConcept(initial)
	require.NoError(t, err)

	// Extension: before Low → should succeed
	ext1 := makeExtensionConcept("Priority", ast.InsertionPoint{BeforeValue: "Low"}, []string{"Critical"})
	err = analyzer.AddOrMergeConcept(ext1)
	require.NoError(t, err)

	// Extension: after Low → different direction, should also succeed
	ext2 := makeExtensionConcept("Priority", ast.InsertionPoint{AfterValue: "Low"}, []string{"Urgent"})
	err = analyzer.AddOrMergeConcept(ext2)
	require.NoError(t, err, "'before Low' followed by 'after Low' should be allowed")

	merged := analyzer.GetConcept("Priority")
	expected := []string{"Critical", "Low", "Urgent", "Medium", "High"}
	require.Len(t, merged.ResolvedValues, len(expected))
	for i, exp := range expected {
		assert.Equal(t, exp, merged.ResolvedValues[i].Name, "value at position %d", i)
	}
}

// TestAddOrMergeConcept_DuplicateValueInExtension verifies that a concept
// extension that attempts to add values already present in the base concept
// returns an error. This prevents silent duplicate entries in ResolvedValues
// and overwrites in the index map.
func TestAddOrMergeConcept_DuplicateValueInExtension(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		existingVals  []string
		extensionIP   ast.InsertionPoint
		extensionVals []string
		desc          string
	}{
		{
			name:          "duplicate value in before extension",
			existingVals:  []string{"Low", "Medium", "High"},
			extensionIP:   ast.InsertionPoint{BeforeValue: "Medium"},
			extensionVals: []string{"Low"},
			desc:          "should error when extension value already exists in base concept",
		},
		{
			name:          "duplicate value in after extension",
			existingVals:  []string{"Low", "Medium", "High"},
			extensionIP:   ast.InsertionPoint{AfterValue: "Low"},
			extensionVals: []string{"Medium"},
			desc:          "should error when extension value already exists in base concept",
		},
		{
			name:          "multiple values with one duplicate",
			existingVals:  []string{"A", "B", "C"},
			extensionIP:   ast.InsertionPoint{BeforeValue: "C"},
			extensionVals: []string{"X", "B", "Y"},
			desc:          "should error when any extension value already exists in base concept",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			analyzer := analysis.NewCPLAnalyzer()
			initial := makeInitialConcept("Priority", tc.existingVals)
			analyzer.AddConcept(initial)

			extension := makeExtensionConcept("Priority", tc.extensionIP, tc.extensionVals)
			err := analyzer.AddOrMergeConcept(extension)
			assert.Error(t, err, tc.desc)
			assert.Contains(t, err.Error(), "duplicate value",
				"error should mention 'duplicate value'")
		})
	}
}
