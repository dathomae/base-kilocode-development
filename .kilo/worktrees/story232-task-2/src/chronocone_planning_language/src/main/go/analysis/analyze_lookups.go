// Package analysis provides the analysis phase data structures for the Chronocone Planning Language.
//
// This file contains the lookup analysis phase implementation.
// It processes lookup definitions from the symbol table and builds the Lookups lookup map.
// See Story242: Lookups Increment 1 - Analysis Phase for the incremental implementation plan.
package analysis

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// AnalyzeLookupScope processes lookup definitions from a LookupsScope symbol table
// and builds the analysis-level lookups map.
//
// This function is called during the symbol table walk for each LookupsScope
// encountered. It reads lookup information from the Definitions slice, validates
// name-based references, and populates the analyzer-wide lookups map with
// fully-resolved definitions.
//
// Symbol Table Structure:
//
//	The parser stores lookup definitions as SymbolDefinition entries in the
//	LookupsScope's Definitions slice. Each SymbolDefinition has:
//	- Name: the lookup identifier
//	- DefinitionType: LookupScopeDefinition
//	- Data: *parser.ParserLookupDefinition with Name, Dimensions, Celltype, SourceLocation
//
// Validation Performed:
//   - Data type validation: each SymbolDefinition.Data must be a *parser.ParserLookupDefinition
//   - Duplicate lookup detection: fails if the lookup name already exists in the analyzer-wide lookups map
//   - Dimension resolution: each dimension name must exist in the concepts map
//   - Celltype resolution: the celltype measure name must exist in the measures map
//
// CIL Tree Generation:
//
// For each valid lookup, a one-node CIL tree is created consisting of a
// CILLookupNode. The tree key uses the "lookup:" prefix to avoid collisions
// with connector and metric CIL trees.
//
// Parameters:
//   - table: The LookupsScope symbol table whose Definitions contain lookup declarations
//   - context: The analysis context (unused for lookups but required for handler signature)
//   - analyzer: The CPLAnalyzer containing the concepts and measures maps for resolution
//
// Returns:
//   - nil if all lookup declarations are successfully processed
//   - error if data type validation or reference validation fails
func AnalyzeLookupScope(table *symboltable.SymbolTable, context *AnalysisContext, analyzer *CPLAnalyzer) error {
	for _, def := range table.Definitions {
		if def.DefinitionType != symboltable.LookupScopeDefinition {
			continue
		}

		if def.Data == nil {
			return fmt.Errorf("lookup definition at line %d, column %d has no structured data",
				def.Location.Line, def.Location.Column)
		}

		lookupDef, ok := def.Data.(*parser.ParserLookupDefinition)
		if !ok {
			return fmt.Errorf("lookup definition at line %d, column %d has unexpected data type %T",
				def.Location.Line, def.Location.Column, def.Data)
		}

		lookupName := lookupDef.Name
		if lookupName == "" {
			return fmt.Errorf("lookup declaration has no name at line %d, column %d",
				def.Location.Line, def.Location.Column)
		}

		// Check for duplicate lookup definitions
		if _, exists := analyzer.lookups[lookupName]; exists {
			return fmt.Errorf("duplicate lookup definition: %s at line %d, column %d",
				lookupName, def.Location.Line, def.Location.Column)
		}

		// Resolve celltype measure
		if lookupDef.Celltype == "" {
			return fmt.Errorf("lookup %s has no celltype measure", lookupName)
		}

		cellTypeMeasure, exists := analyzer.measures[lookupDef.Celltype]
		if !exists {
			return fmt.Errorf("lookup %s references undefined celltype measure: %s at line %d, column %d",
				lookupName, lookupDef.Celltype, def.Location.Line, def.Location.Column)
		}

		// Resolve dimension concepts
		var dimensions []*AnalysisConceptDefinition
		for _, dimName := range lookupDef.Dimensions {
			if !analyzer.conceptAnalyzer.HasConcept(dimName) {
				return fmt.Errorf("lookup %s references undefined concept dimension: %s at line %d, column %d",
					lookupName, dimName, def.Location.Line, def.Location.Column)
			}
			conceptCopy := analyzer.conceptAnalyzer.GetConcept(dimName)
			dimensions = append(dimensions, &conceptCopy)
		}

		// Create analysis-level lookup definition
		lookupInfo := &LookupInfo{
			Name:            lookupName,
			Dimensions:      dimensions,
			CelltypeMeasure: &cellTypeMeasure,
			SourceLocation:  lookupDef.SourceLocation,
		}

		// Store the lookup definition
		analyzer.lookups[lookupName] = lookupInfo

		// Create a one-node CIL tree for the lookup
		analyzer.metricAnalyzer.AddCILTree("lookup:"+lookupName, &cil.CILLookupNode{
			LookupName:     lookupName,
			Dimensions:     lookupDef.Dimensions,
			Celltype:       lookupDef.Celltype,
			SourceLocation: lookupDef.SourceLocation,
		})
	}

	return nil
}
