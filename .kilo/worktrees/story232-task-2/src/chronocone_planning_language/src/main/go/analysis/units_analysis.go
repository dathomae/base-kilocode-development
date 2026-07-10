// Package analysis provides the analysis phase data structures for the Chronocone Planning Language.
//
// This file contains the units analysis phase implementation.
// It processes unit definitions from the symbol table and builds the Units lookup map.
package analysis

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// AnalysisUnitDefinition contains the analyzed unit information.
//
// A unit definition links a unit name to a concept name (the unit group).
// This structure represents a fully-processed unit definition, ready for
// validation and code generation.
//
// Example: Currency = Country
//   - Name: "Currency"
//   - ConceptName: "Country"
type AnalysisUnitDefinition struct {
	// Name is the name of the unit being defined.
	Name string

	// ConceptName is the name of the concept that serves as the unit group.
	// The concept must be defined in the concepts section.
	ConceptName string

	// Location is where the unit definition appears in the source.
	Location symboltable.SourceLocation

	// Conversions holds the analyzed conversion pairs for this unit.
	// Each pair describes a bidirectional conversion between this unit
	// and another unit within the same unit group.
	Conversions []ConversionPair[AnalysisUnitDefinition]
}

// AnalyzeUnitScope processes a single unit definition when entering a UnitScope.
// This function is registered with WalkSymbolTableWithHandlers and called for each
// UnitScope encountered during the walk.
//
// It extracts the unit name and concept name from the UnitScope's identifier definitions.
// The first identifier is the unit name, the second identifier is the concept name.
//
// Example unit declaration: Currency = Country
//   - Unit name: Currency (first identifier)
//   - Concept name: Country (second identifier)
func AnalyzeUnitScope(table *symboltable.SymbolTable, context *AnalysisContext, analyzer *CPLAnalyzer) error {
	// The UnitScope should contain two Identifiers:
	// - First: unit name (e.g., "Currency")
	// - Second: concept name (e.g., "Country")
	//
	// Identifiers are added to the symbol table by the parser's EnterIdentifier method
	// (defined in cpl_document_listener.go) when those nodes are visited.
	if len(table.Definitions) == 0 {
		// A parser rule matched nothing.
		return nil
	}

	if len(table.Definitions) != 2 {
		return fmt.Errorf("unit definition must have two identifiers: unit_name = concept_name, got %d", len(table.Definitions))
	}

	unitName := *table.Definitions[0].Name
	conceptName := *table.Definitions[1].Name
	var location symboltable.SourceLocation = table.Definitions[0].Location

	// Cross-section duplicate detection: Check for duplicate unit names
	if _, exists := analyzer.units[unitName]; exists {
		return fmt.Errorf("duplicate unit definition: %s", unitName)
	}

	// Build analysis unit definition
	analysisDef := AnalysisUnitDefinition{
		Name:        unitName,
		ConceptName: conceptName,
		Location:    location,
	}

	analyzer.units[unitName] = analysisDef

	return nil
}

// VerifyUnitConceptReferences validates that all concept references in unit definitions exist.
// This function should be called after all units and concepts have been collected during analysis.
//
// For each unit definition, it checks that the referenced concept name exists in the concepts
// map. Units that reference undefined concepts will cause this function to return an error.
//
// Parameters:
//   - units: A map of unit name to AnalysisUnitDefinition containing the concept references
//   - concepts: A map of concept name to AnalysisConceptDefinition for validation
//
// Returns:
//   - An error if any unit references a concept that does not exist
//   - Nil if all concept references are valid
//
// Example error:
//
//	"unit 'Currency' references undefined concept 'Country'"
//	"unit 'Weight' references undefined concept 'WeightUnit'"
func VerifyUnitConceptReferences(units map[string]AnalysisUnitDefinition, concepts map[string]AnalysisConceptDefinition) error {
	for unitName, unitDef := range units {
		if _, exists := concepts[unitDef.ConceptName]; !exists {
			return fmt.Errorf("unit %q references undefined concept %q", unitName, unitDef.ConceptName)
		}
	}
	return nil
}
