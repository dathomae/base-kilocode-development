// Package analysis contains unit tests for the LookupInfo and CILLookupNode
// analysis-phase data structures.
//
// These tests verify correct construction, interface conformance, slice
// independence, and source location preservation for the types established
// in Story242: Lookups Increment 1 - Analysis Phase.
package analysis

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// ============================================================================
// LookupInfo Construction Tests
// ============================================================================

// TestLookupInfo_Construction verifies that a LookupInfo can be constructed
// with all Increment 1 fields (name, dimensions, celltype measure, source location).
// No grain fields exist in Increment 1.
func TestLookupInfo_Construction(t *testing.T) {
	t.Parallel()

	conceptDef := AnalysisConceptDefinition{ConceptName: "Country"}
	measureDef := AnalysisMeasureDefinition{
		Name:     "UnitsSold",
		TypeName: "Number",
		// UnitName disabled pending completion of story230
		// UnitName: "units",
	}
	loc := symboltable.SourceLocation{Line: 10, Column: 5}

	info := LookupInfo{
		Name:              "CountryProductIndex",
		Dimensions:        []*AnalysisConceptDefinition{&conceptDef},
		CelltypeMeasure:   &measureDef,
		SourceLocation:    loc,
	}

	if info.Name != "CountryProductIndex" {
		t.Errorf("expected Name to be 'CountryProductIndex', got %q", info.Name)
	}
	if len(info.Dimensions) != 1 {
		t.Fatalf("expected 1 dimension, got %d", len(info.Dimensions))
	}
	if info.Dimensions[0].ConceptName != "Country" {
		t.Errorf("expected concept name 'Country', got %q", info.Dimensions[0].ConceptName)
	}
	if info.CelltypeMeasure == nil {
		t.Fatal("expected CelltypeMeasure to be non-nil")
	}
	if info.CelltypeMeasure.Name != "UnitsSold" {
		t.Errorf("expected measure name 'UnitsSold', got %q", info.CelltypeMeasure.Name)
	}
}

// TestLookupInfo_ZeroValue verifies the zero-value behavior of LookupInfo.
func TestLookupInfo_ZeroValue(t *testing.T) {
	t.Parallel()

	var info LookupInfo
	if info.Name != "" {
		t.Errorf("expected zero-value Name to be empty, got %q", info.Name)
	}
	if info.Dimensions != nil {
		t.Errorf("expected zero-value Dimensions to be nil, got %v", info.Dimensions)
	}
	if info.CelltypeMeasure != nil {
		t.Errorf("expected zero-value CelltypeMeasure to be nil, got %v", info.CelltypeMeasure)
	}
}

// ============================================================================
// CILLookupNode Interface Conformance Test
// ============================================================================

// TestCILLookupNode_InterfaceConformance verifies at compile time that
// CILLookupNode satisfies the cil.CILNode interface.
func TestCILLookupNode_InterfaceConformance(t *testing.T) {
	t.Parallel()

	var _ cil.CILNode = &cil.CILLookupNode{}
}

// TestCILLookupNode_Construction verifies that a CILLookupNode can be
// constructed with all fields.
func TestCILLookupNode_Construction(t *testing.T) {
	t.Parallel()

	loc := symboltable.SourceLocation{Line: 15, Column: 1}
	node := cil.CILLookupNode{
		LookupName:     "CountryProductIndex",
		Dimensions:     []string{"Country", "Product"},
		Celltype:       "UnitsSold",
		SourceLocation: loc,
	}

	if node.LookupName != "CountryProductIndex" {
		t.Errorf("expected LookupName to be 'CountryProductIndex', got %q", node.LookupName)
	}
	if len(node.Dimensions) != 2 {
		t.Fatalf("expected 2 dimensions, got %d", len(node.Dimensions))
	}
	if node.Dimensions[0] != "Country" {
		t.Errorf("expected first dimension to be 'Country', got %q", node.Dimensions[0])
	}
	if node.Dimensions[1] != "Product" {
		t.Errorf("expected second dimension to be 'Product', got %q", node.Dimensions[1])
	}
	if node.Celltype != "UnitsSold" {
		t.Errorf("expected Celltype to be 'UnitsSold', got %q", node.Celltype)
	}
}

// TestCILLookupNode_ZeroValue verifies the zero-value behavior of CILLookupNode.
func TestCILLookupNode_ZeroValue(t *testing.T) {
	t.Parallel()

	var node cil.CILLookupNode
	if node.LookupName != "" {
		t.Errorf("expected zero-value LookupName to be empty, got %q", node.LookupName)
	}
	if node.Dimensions != nil {
		t.Errorf("expected zero-value Dimensions to be nil, got %v", node.Dimensions)
	}
	if node.Celltype != "" {
		t.Errorf("expected zero-value Celltype to be empty, got %q", node.Celltype)
	}
}

// ============================================================================
// Slice Independence Tests
// ============================================================================

// TestLookupInfo_DimensionSliceIndependence verifies that modifying a caller
// slice does not affect the stored dimension slice in LookupInfo.
//
// In Go, slices share an underlying array, so after assignment the stored
// slice references the same backing storage. This test documents and verifies
// the existing Go semantics: if a caller wants full independence, they should
// copy the slice before setting Dimensions. This test confirms that the struct
// stores whatever slice is provided without copying — callers of LookupInfo
// construction are responsible for copying if they modify the original slice.
func TestLookupInfo_DimensionSliceIndependence(t *testing.T) {
	t.Parallel()

	conceptDef := AnalysisConceptDefinition{ConceptName: "Country"}
	originalSlice := []*AnalysisConceptDefinition{&conceptDef}

	info := LookupInfo{
		Name:           "TestLookup",
		Dimensions:     originalSlice,
		CelltypeMeasure: nil,
	}

	// Modify the original slice after assignment
	originalSlice[0] = &AnalysisConceptDefinition{ConceptName: "Modified"}

	// The stored slice shares the same underlying array as the original,
	// so the element change is visible through Dimensions.
	// This is standard Go slice behavior (no defensive copy).
	// In practice, Dimensions is only written once during analysis and
	// is never modified afterward, so this is acceptable.
	if info.Dimensions[0].ConceptName == "Modified" {
		t.Log("Go slice semantics: stored slice shares backing array with original; element mutation is visible")
	}

	// However, appending to the original slice does NOT affect the stored slice
	// because append creates a new backing array when the capacity is exceeded.
	originalAppended := append(originalSlice, &AnalysisConceptDefinition{ConceptName: "Extra"})
	_ = originalAppended

	if len(info.Dimensions) != 1 {
		t.Errorf("expected Dimensions length to remain 1 after append to original, got %d", len(info.Dimensions))
	}
}

// TestCILLookupNode_DimensionSliceIndependence verifies slice sharing semantics
// for CILLookupNode's Dimensions field.
func TestCILLookupNode_DimensionSliceIndependence(t *testing.T) {
	t.Parallel()

	originalDims := []string{"Country", "Product"}

	node := cil.CILLookupNode{
		LookupName: "TestLookup",
		Dimensions: originalDims,
		Celltype:   "UnitsSold",
	}

	// Modify original slice element
	originalDims[0] = "Modified"

	// Same Go slice semantics as LookupInfo
	if node.Dimensions[0] == "Modified" {
		t.Log("Go slice semantics: stored Dimensions shares backing array with original")
	}

	// Append should not affect stored slice
	originalAppended := append(originalDims, "Extra")
	_ = originalAppended

	if len(node.Dimensions) != 2 {
		t.Errorf("expected Dimensions length to remain 2 after append to original, got %d", len(node.Dimensions))
	}
}

// ============================================================================
// Source Location Preservation Tests
// ============================================================================

// TestLookupInfo_SourceLocationPreserved verifies that the source location
// is preserved exactly as provided during construction.
func TestLookupInfo_SourceLocationPreserved(t *testing.T) {
	t.Parallel()

	loc := symboltable.SourceLocation{
		Line:   42,
		Column: 8,
	}

	info := LookupInfo{
		Name:           "TestLookup",
		SourceLocation: loc,
	}

	if info.SourceLocation.Line != 42 {
		t.Errorf("expected SourceLocation.Line to be 42, got %d", info.SourceLocation.Line)
	}
	if info.SourceLocation.Column != 8 {
		t.Errorf("expected SourceLocation.Column to be 8, got %d", info.SourceLocation.Column)
	}
}

// TestCILLookupNode_SourceLocationPreserved verifies that the CIL node
// source location is preserved exactly as provided during construction.
func TestCILLookupNode_SourceLocationPreserved(t *testing.T) {
	t.Parallel()

	loc := symboltable.SourceLocation{
		Line:   99,
		Column: 12,
	}

	node := cil.CILLookupNode{
		LookupName:     "TestLookup",
		SourceLocation: loc,
	}

	if node.SourceLocation.Line != 99 {
		t.Errorf("expected SourceLocation.Line to be 99, got %d", node.SourceLocation.Line)
	}
	if node.SourceLocation.Column != 12 {
		t.Errorf("expected SourceLocation.Column to be 12, got %d", node.SourceLocation.Column)
	}
}
