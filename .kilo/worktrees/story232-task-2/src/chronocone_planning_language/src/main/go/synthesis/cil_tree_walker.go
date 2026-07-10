// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package synthesis

import (
	"fmt"
	"strings"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
)

// CILTreeWalker walks the CIL trees in analysis definitions and registers
// discovered entities (connectors, lookups, metrics) in the CIL abstract
// transformer's registry. It is the single point of type-aware dispatch for
// CIL node types, replacing ad-hoc inline loops in BuildTransformerFromAnalysis.
//
// Connector metadata is sourced from the structured analysis-level connector
// definitions (analysisDefs.Connectors) because this carries the full
// set of connector metadata (grain, grain-to-concept mappings, resolved
// concept references). The CILConnectorNode in CILTrees carries only the
// connector name and serves as a leaf node for CIL expression trees.
//
// Lookup registration walks CILTrees looking for CILLookupNode entries
// keyed with the "lookup:" prefix.
//
// Usage:
//
//	walker := NewCILTreeWalker(transformer)
//	if err := walker.WalkCILTrees(analysisDefs); err != nil {
//	    // handle error
//	}
//
// After walking, the transformer's CILRegistry is populated with metadata
// for all discovered connectors, lookups, and measure types, ready for code
// generation.
type CILTreeWalker struct {
	transformer *cil.CILAbstractTransformer
}

// NewCILTreeWalker creates a new CIL tree walker that registers discovered
// entities into the given transformer's registry.
//
// Parameters:
//   - transformer: The CIL abstract transformer whose registry will be
//     populated with discovered entities.
//
// Returns:
//   - *CILTreeWalker: A new walker instance.
func NewCILTreeWalker(transformer *cil.CILAbstractTransformer) *CILTreeWalker {
	return &CILTreeWalker{
		transformer: transformer,
	}
}

// WalkCILTrees walks the analysis definitions and registers all discovered
// entities in the transformer's CILRegistry. It processes three sources of
// registration data:
//
//  1. Connector definitions from analysisDefs.Connectors (structured metadata
//     with full dimension, grain, and cell-type information).
//  2. Lookup nodes from analysisDefs.CILTrees (CILLookupNode entries keyed
//     with the "lookup:" prefix).
//  3. Measure type registrations from analysisDefs.Measures.
//
// Parameters:
//   - analysisDefs: The analysis definitions containing connector definitions,
//     CIL trees, measure definitions, and type definitions.
//
// Returns:
//   - error: An error if connector or measure type resolution fails.
func (w *CILTreeWalker) WalkCILTrees(analysisDefs *analysis.AnalysisDefinitions) error {
	if analysisDefs == nil || w.transformer == nil {
		return nil
	}

	// 1. Register connectors from structured analysis definitions.
	//    Connector metadata (dimensions, grain, grain→concept mappings)
	//    lives in the analysis-level connector definition, not in
	//    CILConnectorNode (which carries only the name).
	if err := w.registerConnectors(analysisDefs); err != nil {
		return err
	}

	// 2. Register lookups from CILTrees by type-dispatching on
	//    CILLookupNode entries.
	if err := w.registerLookups(analysisDefs.CILTrees); err != nil {
		return err
	}

	// 3. Register measure types from analysis definitions.
	if err := w.registerMeasureTypes(analysisDefs); err != nil {
		return err
	}

	return nil
}

// registerConnectors processes all connector definitions and registers
// their metadata in the transformer's CILRegistry.
//
// Parameters:
//   - analysisDefs: The analysis definitions containing connector and type
//     definitions for cell-type resolution.
//
// Returns:
//   - error: An error if cell type resolution fails.
func (w *CILTreeWalker) registerConnectors(analysisDefs *analysis.AnalysisDefinitions) error {
	for _, connectorDef := range analysisDefs.Connectors {
		if err := w.registerOneConnector(connectorDef, analysisDefs.Types); err != nil {
			return err
		}
	}
	return nil
}

// registerOneConnector registers a single connector definition's metadata.
//
// Parameters:
//   - connectorDef: The analysis-level connector definition.
//   - types: The map of type definitions for cell-type resolution.
//
// Returns:
//   - error: An error if cell type resolution fails.
func (w *CILTreeWalker) registerOneConnector(connectorDef analysis.AnalysisConnectorDefinition, types map[string]analysis.AnalysisTypeDefinition) error {
	// Extract dimension concept names.
	dimensions := make([]string, 0, len(connectorDef.Dimensions))
	for _, dim := range connectorDef.Dimensions {
		dimensions = append(dimensions, dim.ConceptName)
	}

	// Resolve cell type through type derivation chain.
	cellType := "float64" // Default
	if connectorDef.CellType != nil {
		var err error
		cellType, err = resolveGoType(connectorDef.CellType.TypeName, types)
		if err != nil {
			return fmt.Errorf("failed to resolve connector %q cell type: %w", connectorDef.Name, err)
		}
	}

	// Build grain and grain-to-concept mapping.
	var grain []string
	var grainToConcept map[string]string
	if len(connectorDef.Grain) > 0 {
		grain = make([]string, 0, len(connectorDef.Grain))
		grainToConcept = make(map[string]string, len(connectorDef.Grain))
		for _, g := range connectorDef.Grain {
			grain = append(grain, g.Name)
			if g.GetUnitName() != "" {
				grainToConcept[g.Name] = g.GetUnitName()
			}
		}
	}

	w.transformer.RegisterConnector(connectorDef.Name, dimensions, grain, grainToConcept, cellType)
	return nil
}

// registerLookups processes CILTrees entries and registers CILLookupNode
// metadata in the transformer's CILRegistry.
//
// Parameters:
//   - cilTrees: The map of CIL tree entries.
//
// Returns:
//   - error: Always nil; lookup registration is not expected to fail.
func (w *CILTreeWalker) registerLookups(cilTrees map[string]cil.CILNode) error {
	if cilTrees == nil {
		return nil
	}
	for key, node := range cilTrees {
		lookupNode, ok := node.(*cil.CILLookupNode)
		if !ok {
			continue
		}
		if !strings.HasPrefix(key, "lookup:") {
			continue
		}
		lookupName := strings.TrimPrefix(key, "lookup:")
		w.transformer.RegisterLookup(lookupName, lookupNode.Dimensions, lookupNode.Celltype)
	}
	return nil
}

// registerMeasureTypes processes measure definitions and registers their
// resolved Go types in the transformer's CILRegistry.
//
// Parameters:
//   - analysisDefs: The analysis definitions containing measure and type
//     definitions.
//
// Returns:
//   - error: An error if a measure's type cannot be resolved.
func (w *CILTreeWalker) registerMeasureTypes(analysisDefs *analysis.AnalysisDefinitions) error {
	for measureName, measureDef := range analysisDefs.Measures {
		goType, err := resolveGoType(measureDef.TypeName, analysisDefs.Types)
		if err != nil {
			return fmt.Errorf("failed to resolve measure %q type: %w", measureName, err)
		}
		w.transformer.RegisterMeasureType(measureName, goType)
	}
	return nil
}
