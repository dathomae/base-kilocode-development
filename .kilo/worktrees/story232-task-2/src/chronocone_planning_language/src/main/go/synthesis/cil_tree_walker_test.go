// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package synthesis

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
)

// TestCILTreeWalker_WalkCILTrees_Connectors verifies that the walker
// registers connector metadata in the transformer's CILRegistry from
// analysisDefs.Connectors (the structured analysis data).
func TestCILTreeWalker_WalkCILTrees_Connectors(t *testing.T) {
	transformer := cil.NewCILAbstractTransformer()
	walker := NewCILTreeWalker(transformer)

	analysisDefs := &analysis.AnalysisDefinitions{
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"sales_data": {
				Name: "sales_data",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Product"},
					{ConceptName: "Region"},
				},
				CellType: &analysis.AnalysisMeasureDefinition{
					Name:     "Revenue",
					TypeName: "Number",
				},
				Grain: []*analysis.AnalysisMeasureDefinition{
					{Name: "Monthly"},
				},
			},
		},
	}

	err := walker.WalkCILTrees(analysisDefs)
	if err != nil {
		t.Fatalf("WalkCILTrees() returned error: %v", err)
	}

	// Verify connector metadata was registered.
	meta := transformer.GetConnectorMetadata("sales_data")
	if meta == nil {
		t.Fatalf("GetConnectorMetadata(%q) returned nil", "sales_data")
	}
	if meta.Name != "sales_data" {
		t.Errorf("connector metadata Name = %q, expected %q", meta.Name, "sales_data")
	}
	if len(meta.Dimensions) != 2 {
		t.Errorf("connector metadata Dimensions length = %d, expected 2", len(meta.Dimensions))
	}
	if meta.Dimensions[0] != "Product" {
		t.Errorf("connector metadata Dimensions[0] = %q, expected %q", meta.Dimensions[0], "Product")
	}
	if meta.Dimensions[1] != "Region" {
		t.Errorf("connector metadata Dimensions[1] = %q, expected %q", meta.Dimensions[1], "Region")
	}
	if meta.CellType != "float64" {
		t.Errorf("connector metadata CellType = %q, expected %q", meta.CellType, "float64")
	}
	if len(meta.Grain) != 1 {
		t.Errorf("connector metadata Grain length = %d, expected 1", len(meta.Grain))
	}
	if meta.Grain[0] != "Monthly" {
		t.Errorf("connector metadata Grain[0] = %q, expected %q", meta.Grain[0], "Monthly")
	}
}

// TestCILTreeWalker_WalkCILTrees_Lookups verifies that the walker
// registers lookup metadata in the transformer's CILRegistry when
// walking CIL trees containing CILLookupNode entries with "lookup:"
// prefixed keys.
func TestCILTreeWalker_WalkCILTrees_Lookups(t *testing.T) {
	transformer := cil.NewCILAbstractTransformer()
	walker := NewCILTreeWalker(transformer)

	analysisDefs := &analysis.AnalysisDefinitions{
		CILTrees: map[string]cil.CILNode{
			"lookup:CountryProductIndex": &cil.CILLookupNode{
				LookupName: "CountryProductIndex",
				Dimensions: []string{"Country", "Product"},
				Celltype:   "UnitsSold",
			},
			"lookup:RegionMapping": &cil.CILLookupNode{
				LookupName: "RegionMapping",
				Dimensions: []string{"Region"},
				Celltype:   "RegionCode",
			},
		},
	}

	err := walker.WalkCILTrees(analysisDefs)
	if err != nil {
		t.Fatalf("WalkCILTrees() returned error: %v", err)
	}

	// Verify first lookup metadata.
	meta := transformer.GetLookupMetadata("CountryProductIndex")
	if meta == nil {
		t.Fatalf("GetLookupMetadata(%q) returned nil", "CountryProductIndex")
	}
	if meta.Name != "CountryProductIndex" {
		t.Errorf("lookup metadata Name = %q, expected %q", meta.Name, "CountryProductIndex")
	}
	if len(meta.Dimensions) != 2 {
		t.Errorf("lookup metadata Dimensions length = %d, expected 2", len(meta.Dimensions))
	}
	if meta.Dimensions[0] != "Country" {
		t.Errorf("lookup metadata Dimensions[0] = %q, expected %q", meta.Dimensions[0], "Country")
	}
	if meta.Dimensions[1] != "Product" {
		t.Errorf("lookup metadata Dimensions[1] = %q, expected %q", meta.Dimensions[1], "Product")
	}
	if meta.CellType != "UnitsSold" {
		t.Errorf("lookup metadata CellType = %q, expected %q", meta.CellType, "UnitsSold")
	}

	// Verify second lookup metadata.
	meta2 := transformer.GetLookupMetadata("RegionMapping")
	if meta2 == nil {
		t.Fatalf("GetLookupMetadata(%q) returned nil", "RegionMapping")
	}
	if len(meta2.Dimensions) != 1 {
		t.Errorf("lookup metadata Dimensions length = %d, expected 1", len(meta2.Dimensions))
	}
	if meta2.Dimensions[0] != "Region" {
		t.Errorf("lookup metadata Dimensions[0] = %q, expected %q", meta2.Dimensions[0], "Region")
	}

	// Verify AllLookupMetadata returns all entries.
	all := transformer.AllLookupMetadata()
	if all == nil {
		t.Fatalf("AllLookupMetadata() returned nil")
	}
	if len(all) != 2 {
		t.Errorf("AllLookupMetadata() returned %d entries, expected 2", len(all))
	}
}

// TestCILTreeWalker_WalkCILTrees_Measures verifies that the walker
// registers measure types from analysis definitions.
func TestCILTreeWalker_WalkCILTrees_Measures(t *testing.T) {
	transformer := cil.NewCILAbstractTransformer()
	walker := NewCILTreeWalker(transformer)

	analysisDefs := &analysis.AnalysisDefinitions{
		Measures: map[string]analysis.AnalysisMeasureDefinition{
			"Revenue": {
				Name:     "Revenue",
				TypeName: "Number",
			},
			"Count": {
				Name:     "Count",
				TypeName: "Number",
			},
		},
	}

	err := walker.WalkCILTrees(analysisDefs)
	if err != nil {
		t.Fatalf("WalkCILTrees() returned error: %v", err)
	}
	// No direct assertion — success is "no error". The registry doesn't
	// expose a direct GetMeasureType method; registration is verified
	// indirectly through synthesis tests.
}

// TestCILTreeWalker_WalkCILTrees_NilAnalysisDefs verifies graceful handling
// of nil analysis definitions.
func TestCILTreeWalker_WalkCILTrees_NilAnalysisDefs(t *testing.T) {
	transformer := cil.NewCILAbstractTransformer()
	walker := NewCILTreeWalker(transformer)

	err := walker.WalkCILTrees(nil)
	if err != nil {
		t.Fatalf("WalkCILTrees(nil) returned error: %v", err)
	}
}

// TestCILTreeWalker_WalkCILTrees_NilTransformer verifies graceful handling
// of a nil transformer inside the walker.
func TestCILTreeWalker_WalkCILTrees_NilTransformer(t *testing.T) {
	walker := &CILTreeWalker{transformer: nil}

	analysisDefs := &analysis.AnalysisDefinitions{
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"test_conn": {
				Name: "test_conn",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Dim"},
				},
			},
		},
		CILTrees: map[string]cil.CILNode{
			"lookup:test": &cil.CILLookupNode{
				LookupName: "test",
				Dimensions: []string{"Dim"},
				Celltype:   "Value",
			},
		},
	}

	err := walker.WalkCILTrees(analysisDefs)
	if err != nil {
		t.Fatalf("WalkCILTrees() with nil transformer returned error: %v", err)
	}
}

// TestCILTreeWalker_WalkCILTrees_ConnectorAndLookup verifies that the
// walker correctly handles connector and lookup registration in a
// single walk. Non-lookup CIL tree nodes (connectors, literals) in
// CILTrees are silently skipped since connector registration sources
// from analysisDefs.Connectors, not from CIL trees.
func TestCILTreeWalker_WalkCILTrees_ConnectorAndLookup(t *testing.T) {
	transformer := cil.NewCILAbstractTransformer()
	walker := NewCILTreeWalker(transformer)

	analysisDefs := &analysis.AnalysisDefinitions{
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"inventory": {
				Name: "inventory",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Warehouse"},
				},
				CellType: &analysis.AnalysisMeasureDefinition{
					Name:     "Quantity",
					TypeName: "Number",
				},
				Grain: []*analysis.AnalysisMeasureDefinition{
					{Name: "Daily"},
				},
			},
		},
		CILTrees: map[string]cil.CILNode{
			// CILConnectorNode in the tree is ignored during registration —
			// connector registration comes from analysisDefs.Connectors.
			"inventory": &cil.CILConnectorNode{ConnectorName: "inventory"},
			"lookup:product_categories": &cil.CILLookupNode{
				LookupName: "product_categories",
				Dimensions: []string{"Product"},
				Celltype:   "CategoryLabel",
			},
			// Non-registerable leaf node — should be silently skipped.
			"revenue_literal": &cil.CILLiteral{Value: 100.0},
		},
	}

	err := walker.WalkCILTrees(analysisDefs)
	if err != nil {
		t.Fatalf("WalkCILTrees() returned error: %v", err)
	}

	// Verify connector was registered from analysisDefs.Connectors.
	if meta := transformer.GetConnectorMetadata("inventory"); meta == nil {
		t.Error("GetConnectorMetadata('inventory') returned nil after walk")
	}

	// Verify lookup was registered from CILTrees.
	if meta := transformer.GetLookupMetadata("product_categories"); meta == nil {
		t.Error("GetLookupMetadata('product_categories') returned nil after walk")
	}
}

// TestCILTreeWalker_WalkCILTrees_LookupWithoutPrefix verifies that
// a CILLookupNode without the "lookup:" key prefix is silently skipped.
func TestCILTreeWalker_WalkCILTrees_LookupWithoutPrefix(t *testing.T) {
	transformer := cil.NewCILAbstractTransformer()
	walker := NewCILTreeWalker(transformer)

	analysisDefs := &analysis.AnalysisDefinitions{
		CILTrees: map[string]cil.CILNode{
			"not_a_lookup_key": &cil.CILLookupNode{
				LookupName: "orphan_lookup",
				Dimensions: []string{"Dim"},
				Celltype:   "Value",
			},
		},
	}

	err := walker.WalkCILTrees(analysisDefs)
	if err != nil {
		t.Fatalf("WalkCILTrees() returned error: %v", err)
	}

	// The orphan lookup should not have been registered because the key
	// lacks the "lookup:" prefix.
	if meta := transformer.GetLookupMetadata("orphan_lookup"); meta != nil {
		t.Error("orphan_lookup should not have been registered without 'lookup:' prefix")
	}
}

// TestCILTreeWalker_AllowsMeasureRegistration verifies that measure
// registration produces no error for valid measures.
func TestCILTreeWalker_AllowsMeasureRegistration(t *testing.T) {
	transformer := cil.NewCILAbstractTransformer()
	walker := NewCILTreeWalker(transformer)

	analysisDefs := &analysis.AnalysisDefinitions{
		Measures: map[string]analysis.AnalysisMeasureDefinition{
			"Revenue": {
				Name:     "Revenue",
				TypeName: "Number",
			},
		},
	}

	err := walker.WalkCILTrees(analysisDefs)
	if err != nil {
		t.Fatalf("WalkCILTrees() returned error: %v", err)
	}
}

// TestCILTreeWalker_MeasureWithUnknownType verifies that the walker
// returns an error when a measure references an unresolvable type.
func TestCILTreeWalker_MeasureWithUnknownType(t *testing.T) {
	transformer := cil.NewCILAbstractTransformer()
	walker := NewCILTreeWalker(transformer)

	analysisDefs := &analysis.AnalysisDefinitions{
		Measures: map[string]analysis.AnalysisMeasureDefinition{
			"MysteryMeasure": {
				Name:     "MysteryMeasure",
				TypeName: "NonexistentType",
			},
		},
	}

	err := walker.WalkCILTrees(analysisDefs)
	if err == nil {
		t.Fatal("WalkCILTrees() expected error for unknown measure type, got nil")
	}
}

// TestCILTreeWalker_ConnectorWithUnresolvableCellType verifies that the
// walker returns an error when a connector's cell type cannot be resolved.
func TestCILTreeWalker_ConnectorWithUnresolvableCellType(t *testing.T) {
	transformer := cil.NewCILAbstractTransformer()
	walker := NewCILTreeWalker(transformer)

	analysisDefs := &analysis.AnalysisDefinitions{
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"bad_connector": {
				Name: "bad_connector",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Dim"},
				},
				CellType: &analysis.AnalysisMeasureDefinition{
					Name:     "GhostMeasure",
					TypeName: "GhostType",
				},
			},
		},
	}

	err := walker.WalkCILTrees(analysisDefs)
	if err == nil {
		t.Fatal("WalkCILTrees() expected error for unresolvable cell type, got nil")
	}
}
