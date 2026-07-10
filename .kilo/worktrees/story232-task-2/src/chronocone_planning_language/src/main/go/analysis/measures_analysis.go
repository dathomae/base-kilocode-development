// Package analysis provides the analysis phase data structures for the Chronocone Planning Language.
//
// This file contains the measures analysis phase implementation.
// It processes measure definitions from the symbol table and builds the Units lookup map.
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
type AnalysisMeasureDefinition struct {
	// Name is the name of the unit being defined.
	Name string

	// TypeName is the name of the type for the measure.
	TypeName string

	// UnitExpression is the parsed expression tree for compound units.
	// For simple units, this will be a *UnitReference.
	// For compound units, this will be a tree of Expression nodes.
	UnitExpression Expression

	// StartsAt is the optional named Instant-derived type that anchors the
	// start of an Interval-based measure (e.g., Midnight for Daily).
	StartsAt *string

	// Conversions contains the measure conversion pairs derived from
	// MeasureConversionScope children during analysis. Both forward and
	// inverse pairs are stored so the synthesizer sees explicit pairs.
	Conversions []ConversionPair[AnalysisMeasureDefinition]

	// Location is where the unit definition appears in the source.
	Location symboltable.SourceLocation
}

// AnalyzeMeasureScope processes a single measure definition when entering a MeasureScope.
// This function is registered with WalkSymbolTableWithHandlers and called for each
// MeasureScope encountered during the walk.
//
// It extracts the measure name and type name from the MeasureScope's identifier
// definitions, builds the unit expression tree from postfix definitions (starting at
// index 2), and extracts the optional starts_at from ConstraintScope children.
//
// Example measure declaration: Currency = CurrencyType(unit=CurrencyUnit)
//   - Measure name: Currency (first identifier)
//   - Type name: CurrencyType (second identifier)
//   - Unit expression: CurrencyUnit (as a UnitReference)
//
// Example interval measure: Daily = Interval(starts_at=Midnight, unit=DurationUnit)
//   - Measure name: Daily (first identifier)
//   - Type name: Interval (second identifier)
//   - Unit expression: DurationUnit (as a UnitReference)
//   - Constraints: starts_at=Midnight
func AnalyzeMeasureScope(table *symboltable.SymbolTable, context *AnalysisContext, analyzer *CPLAnalyzer) error {
	if len(table.Definitions) == 0 {
		// A parser rule matched nothing.
		return nil
	}

	if len(table.Definitions) < 2 {
		return fmt.Errorf("measure definition must have at least two identifiers: measure_name = type_name(...), got %d", len(table.Definitions))
	}

	measureName := *table.Definitions[0].Name
	typeName := *table.Definitions[1].Name
	var location symboltable.SourceLocation = table.Definitions[0].Location

	var startsAt *string

	// Build expression tree from postfix definitions (Story229)
	var unitExpression Expression
	if len(table.Definitions) > 2 {
		var buildErr error
		unitExpression, buildErr = BuildExpressionTree(table.Definitions[2:], analyzer.units, nil)
		if buildErr != nil {
			return fmt.Errorf("failed to build unit expression for measure %q: %w", measureName, buildErr)
		}
	}

	// Extract starts_at constraint and conversion clauses from MeasureScope children
	for _, child := range table.Children {
		switch child.ScopeType {
		case symboltable.ConstraintScope:
			constraintName, constraintValue, err := extractConstraint(child)
			if err != nil {
				return fmt.Errorf("failed to extract constraint in measure %s: %w", measureName, err)
			}

			switch constraintName {
			case "starts_at":
				startsAt = &constraintValue
			default:
				return fmt.Errorf("unknown constraint '%s' in measure definition '%s'", constraintName, measureName)
			}
		case symboltable.MeasureConversionScope:
			// Conversion analysis is deferred to the post-processing step in cpl_analyzer.go
			// (analyzeAllMeasureConversions), which runs after ALL measures are registered.
			// This avoids ordering issues where a conversion references a measure
			// that hasn't been processed yet during the initial walk.
		default:
			return fmt.Errorf("measure %q has unexpected child scope type %v", measureName, child.ScopeType)
		}
	}

	if unitExpression == nil {
		return fmt.Errorf("measure definition '%s' must specify a unit expression", measureName)
	}

	// Cross-section duplicate detection: Check for duplicate measure names
	if _, exists := analyzer.measures[measureName]; exists {
		return fmt.Errorf("duplicate measure definition: %s", measureName)
	}

	// Build analysis measure definition
	analysisDef := AnalysisMeasureDefinition{
		Name:           measureName,
		TypeName:       typeName,
		UnitExpression: unitExpression,
		StartsAt:       startsAt,
		Location:       location,
	}

	// Store in the measures map. Conversion analysis is deferred to the
	// analyzeAllMeasureConversions post-processing step in cpl_analyzer.go,
	// which runs after all measures from all sections are registered.
	analyzer.measures[measureName] = analysisDef

	return nil
}

// GetUnitNameFromExpression extracts the unit name from a UnitReference expression.
// Returns "" for nil or compound (non-UnitReference) expressions.
func GetUnitNameFromExpression(expr Expression) string {
	if ref, ok := expr.(*UnitReference); ok {
		return ref.Name
	}
	return ""
}

// GetUnitName returns the unit name for this measure definition.
// It extracts the name from the UnitExpression tree when it is a simple
// UnitReference, or returns "" for compound expressions.
func (m AnalysisMeasureDefinition) GetUnitName() string {
	return GetUnitNameFromExpression(m.UnitExpression)
}

// VerifyMeasureUnitReferences validates that all unit references in measure expression
// trees resolve to declared units. Skips validation when no units are declared.
func (analyzer *CPLAnalyzer) VerifyMeasureUnitReferences() error {
	if len(analyzer.units) == 0 {
		return nil
	}
	for name, measure := range analyzer.measures {
		if measure.UnitExpression != nil {
			if err := verifyExpressionRefs(measure.UnitExpression, analyzer.units, name); err != nil {
				return err
			}
		}
	}
	return nil
}

func verifyExpressionRefs(expr Expression, units map[string]AnalysisUnitDefinition, measureName string) error {
	switch e := expr.(type) {
	case *UnitReference:
		if e.UnitGroup == nil {
			// Attempt deferred resolution against the complete units map
			if unit, exists := units[e.Name]; exists {
				e.UnitGroup = &unit
			} else {
				return fmt.Errorf("measure %q references undefined unit %q", measureName, e.Name)
			}
		}
	case *DivisionNode:
		if err := verifyExpressionRefs(e.Left, units, measureName); err != nil {
			return err
		}
		return verifyExpressionRefs(e.Right, units, measureName)
	case *MultiplicationNode:
		if err := verifyExpressionRefs(e.Left, units, measureName); err != nil {
			return err
		}
		return verifyExpressionRefs(e.Right, units, measureName)
	case *AdditionNode:
		if err := verifyExpressionRefs(e.Left, units, measureName); err != nil {
			return err
		}
		return verifyExpressionRefs(e.Right, units, measureName)
	case *SubtractionNode:
		if err := verifyExpressionRefs(e.Left, units, measureName); err != nil {
			return err
		}
		return verifyExpressionRefs(e.Right, units, measureName)
	case *PowerNode:
		return verifyExpressionRefs(e.Base, units, measureName)
	}
	return nil
}
