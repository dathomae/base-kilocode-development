// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file contains the aggregator analysis phase implementation.
// It processes aggregator definitions from the symbol table and builds
// the aggregators map and CIL trees.
package analysis

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// AnalyzeAggregatorScope processes aggregator definitions from an AggregatorScope
// symbol table and builds the analysis-level aggregators map and CIL trees.
//
// This function is called during the symbol table walk for each AggregatorScope
// encountered. It reads the aggregator definition from the Definitions slice,
// resolves concept references, evaluates the postfix expression, validates
// dimensional units, and populates the analyzer-wide aggregators map.
//
// Symbol Table Structure:
//
//	The parser stores one AggregatorScopeDefinition in the scope's Definitions
//	slice alongside the postfix expression definitions. The AggregatorScopeDefinition
//	carries a *parser.ParserAggregatorDefinition with:
//	- Key: synthetic identifier
//	- FromConcepts: ordered list of source concept names
//	- ToConcept: target concept name (empty if ToEmpty)
//	- ToEmpty: true for scalar reduction (to [])
//
// Algorithm:
//  1. Find the AggregatorScopeDefinition entry; extract parser data
//  2. Resolve from-concept names against analyzer concepts
//  3. Resolve to-concept name (unless ToEmpty); validate to-concept is in from list
//     and that at least one from-concept differs from to-concept (collapsible dimension)
//  4. Evaluate the aggregator postfix expression (excluding AggregatorScopeDefinition)
//  5. Build AggregatorInfo, validate dimensional units
//  6. Construct compound key, store in analyzer.aggregators[key]
//  7. Build AggregationMatrix using BuildSumAggregationMatrix
//  8. Create CILAggregateNode and store in CIL trees under "aggregator:" prefix
//
// Parameters:
//   - table: The AggregatorScope symbol table whose Definitions contain the
//     aggregator declaration and postfix expression.
//   - context: The analysis context (unused for aggregators but required for handler signature).
//   - analyzer: The CPLAnalyzer containing concepts, aggregators map, and function registry.
//
// Returns:
//   - nil if the aggregator is successfully processed
//   - error if any validation fails
func AnalyzeAggregatorScope(table *symboltable.SymbolTable, context *AnalysisContext, analyzer *CPLAnalyzer) error {
	// Step 1: Find the AggregatorScopeDefinition and extract parser data.
	// Also collect expression-only definitions (excluding AggregatorScopeDefinition).
	var aggDef *parser.ParserAggregatorDefinition
	var exprDefs []symboltable.SymbolDefinition

	for _, def := range table.Definitions {
		if def.DefinitionType == symboltable.AggregatorScopeDefinition {
			if def.Data == nil {
				return fmt.Errorf("aggregator at line %d, column %d has no structured data",
					def.Location.Line, def.Location.Column)
			}
			parsed, ok := def.Data.(*parser.ParserAggregatorDefinition)
			if !ok {
				return fmt.Errorf("aggregator at line %d, column %d has unexpected data type %T",
					def.Location.Line, def.Location.Column, def.Data)
			}
			aggDef = parsed
		} else {
			exprDefs = append(exprDefs, def)
		}
	}

	if aggDef == nil {
		return fmt.Errorf("aggregator scope has no AggregatorScopeDefinition")
	}

	// Step 2: Resolve from-concept names against the analyzer's concepts.
	concepts := analyzer.GetConcepts()
	var fromConcepts []*AnalysisConceptDefinition
	for _, name := range aggDef.FromConcepts {
		conceptDef, exists := concepts[name]
		if !exists {
			return fmt.Errorf("aggregator references undefined from-concept: %s at line %d, column %d",
				name, aggDef.SourceLocation.Line, aggDef.SourceLocation.Column)
		}
		fromConcepts = append(fromConcepts, &conceptDef)
	}

	// Step 3: Resolve to-concept name (unless ToEmpty).
	var toConcept *AnalysisConceptDefinition
	if !aggDef.ToEmpty {
		toConceptDef, exists := concepts[aggDef.ToConcept]
		if !exists {
			return fmt.Errorf("aggregator references undefined to-concept: %s at line %d, column %d",
				aggDef.ToConcept, aggDef.SourceLocation.Line, aggDef.SourceLocation.Column)
		}
		// Check that to-concept is in the from list.
		found := false
		for _, name := range aggDef.FromConcepts {
			if name == aggDef.ToConcept {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("aggregator to-concept %s must be in the from-concept list at line %d, column %d",
				aggDef.ToConcept, aggDef.SourceLocation.Line, aggDef.SourceLocation.Column)
		}
		// At least one from-concept must differ from the to-concept,
		// otherwise there is no dimension to collapse.
		// e.g., from [Country] to [Country] is invalid because
		// both from and to are the same with no other dimensions.
		collapsibleFound := false
		for _, name := range aggDef.FromConcepts {
			if name != aggDef.ToConcept {
				collapsibleFound = true
				break
			}
		}
		if !collapsibleFound {
			return fmt.Errorf("aggregator from-concepts must contain at least one concept different from to-concept %s for collapsing at line %d, column %d",
				aggDef.ToConcept, aggDef.SourceLocation.Line, aggDef.SourceLocation.Column)
		}
		toConcept = &toConceptDef
	} else {
		// ToEmpty requires exactly 1 from-concept.
		if len(aggDef.FromConcepts) != 1 {
			return fmt.Errorf("aggregator with to-empty (scalar reduction) requires exactly 1 from-concept, got %d at line %d, column %d",
				len(aggDef.FromConcepts), aggDef.SourceLocation.Line, aggDef.SourceLocation.Column)
		}
	}

	// Step 4: Evaluate the aggregator postfix expression.
	exprResult, cellOp, err := EvaluateAggregatorExpression(exprDefs, fromConcepts, analyzer.functionRegistry)
	if err != nil {
		return fmt.Errorf("aggregator expression evaluation failed at line %d, column %d: %w",
			aggDef.SourceLocation.Line, aggDef.SourceLocation.Column, err)
	}

	// Step 5: Build AggregatorInfo and validate dimensional units.
	info := &AggregatorInfo{
		FromConcepts:       fromConcepts,
		ToConcept:          toConcept,
		ToEmpty:            aggDef.ToEmpty,
		ExpressionDimUnits: exprResult,
		CellOperation:      cellOp,
		SourceLocation:     aggDef.SourceLocation,
	}

	if err := ValidateAggregatorDimensionalUnits(info); err != nil {
		return fmt.Errorf("aggregator dimensional unit validation failed at line %d, column %d: %w",
			aggDef.SourceLocation.Line, aggDef.SourceLocation.Column, err)
	}

	// Step 6: Construct compound key and check for duplicates.
	key := AggregatorMapKey(aggDef.FromConcepts, aggDef.ToConcept)
	if _, exists := analyzer.aggregators[key]; exists {
		return fmt.Errorf("duplicate aggregator definition: %s at line %d, column %d",
			key, aggDef.SourceLocation.Line, aggDef.SourceLocation.Column)
	}

	// Store in aggregators map.
	analyzer.aggregators[key] = info

	// Step 7: Build AggregationMatrix.
	// Determine which from-concepts are collapsed (appear in expression numerator)
	// and which are preserved (the to-concept).
	collapsedValues := extractCollapsedConceptValues(fromConcepts, aggDef.ToConcept, exprResult)
	var targetValues []string
	if !aggDef.ToEmpty && toConcept != nil {
		targetValues = extractConceptValueNames(toConcept)
	}

	aggMatrix := BuildSumAggregationMatrix(collapsedValues, targetValues)

	// Step 8: Create CILAggregateNode and store in CIL trees.
	cilNode := &cil.CILAggregateNode{
		FromDimensions:       aggDef.FromConcepts,
		ToDimension:          aggDef.ToConcept,
		AggregationMatrix:    AnalysisMatrixToCIL(aggMatrix),
		CellOperation:         cellOp,
		DimensionalUnitEffect: exprResult.String(),
		SourceLocation:        aggDef.SourceLocation,
	}

	analyzer.metricAnalyzer.AddCILTree("aggregator:"+key, cilNode)

	return nil
}

// extractCollapsedConceptValues extracts the value names from the from-concepts
// that are being collapsed (those that appear in the expression).
// The to-concept is preserved (not collapsed), so its values are excluded.
func extractCollapsedConceptValues(
	fromConcepts []*AnalysisConceptDefinition,
	toConceptName string,
	exprResult *DimensionalUnitExpression,
) []string {
	// Build a set of concept names that appear in the expression.
	exprConcepts := make(map[string]bool, len(exprResult.Numerator)+len(exprResult.Denominator))
	for _, n := range exprResult.Numerator {
		exprConcepts[n] = true
	}
	for _, n := range exprResult.Denominator {
		exprConcepts[n] = true
	}

	var values []string
	for _, c := range fromConcepts {
		// Skip the to-concept (preserved, not collapsed).
		if c.ConceptName == toConceptName {
			continue
		}
		// Include concepts that appear in the expression
		// (these are the ones being operated on / collapsed).
		if exprConcepts[c.ConceptName] {
			values = append(values, extractConceptValueNames(c)...)
		}
	}

	return values
}

// extractConceptValueNames extracts the Name field from each ResolvedValues entry.
func extractConceptValueNames(concept *AnalysisConceptDefinition) []string {
	names := make([]string, len(concept.ResolvedValues))
	for i, v := range concept.ResolvedValues {
		names[i] = v.Name
	}
	return names
}
