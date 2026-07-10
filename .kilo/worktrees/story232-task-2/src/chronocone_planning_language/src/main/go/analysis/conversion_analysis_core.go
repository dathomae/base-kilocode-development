// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file contains the shared generic conversion analysis core that is used
// by both unit and measure conversion analysis.
package analysis

import (
	"fmt"
	"reflect"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// analyzeConversions is the shared generic core for conversion analysis.
// It walks the children of the given scope table looking for targetScopeType
// entries, pairs their PatternScope and CalculationScope children, and builds
// ConversionPair[Def] values that are appended to def.Conversions.
//
// The type parameter Def is either AnalysisUnitDefinition or AnalysisMeasureDefinition.
// If validatePair is non-nil, it is called after target and source definitions
// are resolved and before conversions are built, allowing callers to enforce
// definition-specific constraints (e.g. concept name matching for units or
// self-conversion checks for measures).
func analyzeConversions[Def any](
	table *symboltable.SymbolTable,
	targetScopeType symboltable.ScopeType,
	lookup map[string]*Def,
	defName string,
	def *Def,
	units map[string]AnalysisUnitDefinition,
	entityLabel string,
	validatePair func(targetDef, sourceDef *Def) error,
) error {
	for _, child := range table.Children {
		if child.ScopeType != targetScopeType {
			continue
		}

		var patternScopes []*symboltable.SymbolTable
		var calcScopes []*symboltable.SymbolTable
		for _, c := range child.Children {
			switch c.ScopeType {
			case symboltable.PatternScope:
				patternScopes = append(patternScopes, c)
			case symboltable.CalculationScope:
				calcScopes = append(calcScopes, c)
			}
		}

		if len(patternScopes) == 0 {
			return fmt.Errorf("conversion missing pattern scope in %s %s", entityLabel, defName)
		}
		if len(calcScopes) == 0 {
			return fmt.Errorf("conversion missing calculation scope in %s %s", entityLabel, defName)
		}
		if len(patternScopes) != len(calcScopes) {
			return fmt.Errorf("%s %s: mismatch between pattern scopes (%d) and calculation scopes (%d)", entityLabel, defName, len(patternScopes), len(calcScopes))
		}

		for i := range patternScopes {
			patternScope := patternScopes[i]
			calcScope := calcScopes[i]

			pairs, err := extractPatternPairs(patternScope)
			if err != nil {
				return fmt.Errorf("%s %s: %w", entityLabel, defName, err)
			}
			if len(pairs) < 2 {
				return fmt.Errorf("%s %s: pattern must have at least two unit-variable pairs, got %d", entityLabel, defName, len(pairs))
			}

			target := pairs[0]
			source := pairs[1]

			targetDef, targetExists := lookup[target.unitName]
			sourceDef, sourceExists := lookup[source.unitName]
			if !targetExists {
				return fmt.Errorf("%s %s: pattern references undefined %s %s", entityLabel, defName, entityLabel, target.unitName)
			}
			if !sourceExists {
				return fmt.Errorf("%s %s: pattern references undefined %s %s", entityLabel, defName, entityLabel, source.unitName)
			}

			if validatePair != nil {
				if err := validatePair(targetDef, sourceDef); err != nil {
					return fmt.Errorf("%s %s: %w", entityLabel, defName, err)
				}
			}

			boundVars := map[string]bool{
				target.variable: true,
				source.variable: true,
			}

			for v := range boundVars {
				if _, exists := units[v]; exists {
					return fmt.Errorf("%s %s: pattern-bound variable %q conflicts with unit name", entityLabel, defName, v)
				}
			}

			if err := validateNoBareNumbers(calcScope.Definitions); err != nil {
				return fmt.Errorf("%s %s: %w", entityLabel, defName, err)
			}

			calcExpr, err := BuildExpressionTree(calcScope.Definitions, units, boundVars)
			if err != nil {
				return fmt.Errorf("%s %s: failed to build calculation expression: %w", entityLabel, defName, err)
			}

			forwardPair := ConversionPair[Def]{
				SourceDef:   sourceDef,
				TargetDef:   targetDef,
				TargetVar:   target.variable,
				SourceVar:   source.variable,
				Calculation: calcExpr,
			}

			inverseCalc, err := invertExpression(calcExpr, source.variable, target.variable)
			if err != nil {
				return fmt.Errorf("%s %s: failed to invert expression: %w", entityLabel, defName, err)
			}

			inversePair := ConversionPair[Def]{
				SourceDef:   targetDef,
				TargetDef:   sourceDef,
				TargetVar:   source.variable,
				SourceVar:   target.variable,
				Calculation: inverseCalc,
			}

			if err := appendConversionsToField(def, forwardPair, inversePair); err != nil {
				return fmt.Errorf("%s %s: %w", entityLabel, defName, err)
			}
		}
	}

	return nil
}

// appendConversionsToField appends conversion pairs to the Conversions field
// of def using reflection. The def parameter must be a pointer to a struct with
// a Conversions field of type []ConversionPair[Def].
//
// Reflection is required here because Go generics do not support accessing
// struct fields through type parameters (e.g., def.Conversions is not allowed
// when Def is constrained by `any`). The Conversions field is accessed by name
// via reflect.Value.FieldByName, then asserted to the concrete slice type
// []ConversionPair[Def] before appending.
func appendConversionsToField[Def any](def *Def, pairs ...ConversionPair[Def]) error {
	v := reflect.ValueOf(def)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("def must be a non-nil pointer")
	}
	elem := v.Elem()
	conversionsField := elem.FieldByName("Conversions")
	if !conversionsField.IsValid() {
		return fmt.Errorf("def has no Conversions field")
	}
	existing, ok := conversionsField.Interface().([]ConversionPair[Def])
	if !ok {
		return fmt.Errorf("Conversions field has incompatible type")
	}
	existing = append(existing, pairs...)
	conversionsField.Set(reflect.ValueOf(existing))
	return nil
}
