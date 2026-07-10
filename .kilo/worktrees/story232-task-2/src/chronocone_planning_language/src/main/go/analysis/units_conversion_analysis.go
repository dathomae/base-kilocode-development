// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file contains unit conversion analysis: extracting pattern/calculation
// pairs from ConversionScope entries and validating them.
package analysis

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// patternPair holds a unit name and its bound variable extracted from a PatternScope.
type patternPair struct {
	unitName string
	variable string
}

// AnalyzeUnitConversions walks the children of a UnitScope to find ConversionScope
// entries and delegates to the shared analyzeConversions core.
//
// In addition to the core validation, it ensures that the target and source units
// belong to the same concept (unit group).
func AnalyzeUnitConversions(table *symboltable.SymbolTable, units map[string]AnalysisUnitDefinition, unitDef *AnalysisUnitDefinition) error {
	lookup := make(map[string]*AnalysisUnitDefinition, len(units))
	for name, u := range units {
		uCopy := u
		lookup[name] = &uCopy
	}

	validatePair := func(targetDef, sourceDef *AnalysisUnitDefinition) error {
		if targetDef.ConceptName != sourceDef.ConceptName {
			return fmt.Errorf("conversion pattern units %s and %s belong to different groups", targetDef.Name, sourceDef.Name)
		}
		return nil
	}

	return analyzeConversions(table, symboltable.ConversionScope, lookup, unitDef.Name, unitDef, units, "unit", validatePair)
}

// extractPatternPairs reads a PatternScope's Definitions in order and extracts
// unit-variable pairs. The expected order is: UnitReference, Identifier,
// UnitReference, Identifier, etc.
func extractPatternPairs(patternScope *symboltable.SymbolTable) ([]patternPair, error) {
	var pairs []patternPair
	var pendingUnit string

	for _, def := range patternScope.Definitions {
		switch def.DefinitionType {
		case symboltable.UnitReferenceDefinition:
			ref, ok := def.Data.(parser.ParserUnitReference)
			if !ok {
				return nil, fmt.Errorf("invalid unit reference data in pattern")
			}
			pendingUnit = ref.Name
		case symboltable.Identifier:
			if pendingUnit == "" {
				return nil, fmt.Errorf("pattern has identifier without preceding unit reference")
			}
			varName := ""
			if def.Name != nil {
				varName = *def.Name
			}
			pairs = append(pairs, patternPair{unitName: pendingUnit, variable: varName})
			pendingUnit = ""
		}
	}

	if pendingUnit != "" {
		return nil, fmt.Errorf("pattern has unit reference %q without bound variable", pendingUnit)
	}

	return pairs, nil
}

// validateNoBareNumbers ensures that a calculation containing numeric literals
// also contains at least one unit reference. A calculation composed entirely of
// bare numbers and operators is rejected.
func validateNoBareNumbers(defs []symboltable.SymbolDefinition) error {
	hasNumber := false
	hasUnitRef := false
	for _, def := range defs {
		switch def.DefinitionType {
		case symboltable.NumberDefinition:
			hasNumber = true
		case symboltable.UnitReferenceDefinition:
			hasUnitRef = true
		}
	}
	if hasNumber && !hasUnitRef {
		return fmt.Errorf("calculation contains bare numbers without unit annotations")
	}
	return nil
}
