// Package analysis provides the analysis phase types for the Chronocone Planning Language.
//
// This file contains the types analysis phase implementation.
// It processes type definitions from the symbol table and builds the Types lookup map.
package analysis

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// EnterAnalyzeTypeScope processes a single type definition when entering a TypeScope.
// This function is registered with WalkSymbolTableWithHandlers and called for each
// TypeScope encountered during the walk.
//
// It extracts and stores the type name, its immediate base type, and raw constraints.
// Note: Constraint validation requires knowledge of the built-in type, which isn't
// available until ComputeDerivationPaths runs at the end of Analyze.
func EnterAnalyzeTypeScope(table *symboltable.SymbolTable, context *AnalysisContext, analyzer *CPLAnalyzer) error {
	// Extract type definition from TypeScope
	// The TypeScope should contain:
	// - An Identifier with the type name
	// - An Identifier with the base type name (e.g., "Number")
	// - Optionally, TypeScope children for constraints

	if len(table.Definitions) == 0 {
		// The parser matched nothing, just ignore it.
		return nil
	}

	typeName, baseTypeName, constraints, err := extractTypeDefinition(table)
	if err != nil {
		return fmt.Errorf("invalid type definition: %w", err)
	}

	// Cross-section duplicate detection: Check for duplicate type names
	if analyzer.typeAnalyzer.HasType(typeName) {
		return fmt.Errorf("duplicate type definition: %s", typeName)
	}

	// Convert base type name to TypeReference
	// The base type can be either a built-in type (like "Number") or a declared type
	baseType, err := createTypeReference(baseTypeName)
	if err != nil {
		return fmt.Errorf("invalid base type for type %s: %w", typeName, err)
	}

	// Build analysis type definition
	// Constraints are stored as raw []Constraint - validation happens in ComputeDerivationPaths
	// Note: DerivationPath is left empty here - it's computed later by ComputeDerivationPaths
	analysisDef := AnalysisTypeDefinition{
		Name:           typeName,
		BaseType:       baseType,
		RawConstraints: constraints, // Raw constraints - validated later when built-in type is known
	}

	analyzer.typeAnalyzer.AddType(analysisDef)

	return nil
}

// extractTypeDefinition extracts type definition information from a TypeScope.
// It returns the type name, base type name, and raw constraints.
func extractTypeDefinition(typeScope *symboltable.SymbolTable) (typeName, baseTypeName string, constraints []Constraint, err error) {
	if len(typeScope.Definitions) != 2 {
		return "", "", nil, fmt.Errorf("Expected type name and base type name in TypeScope, got definitions of length %d", len(typeScope.Definitions))
	}

	typeName = *typeScope.Definitions[0].Name
	baseTypeName = *typeScope.Definitions[1].Name

	// Extract constraints from TypeScope children
	for _, child := range typeScope.Children {
		if child.ScopeType != symboltable.ConstraintScope {
			return "", "", nil, fmt.Errorf("Improper symbol table structure detected in extractTypeDefinition")
		}

		constraintName, constraintValue, err := extractConstraint(child)
		if err != nil {
			return "", "", nil, fmt.Errorf("failed to extract constraint: %w", err)
		}
		constraints = append(constraints, Constraint{
			Name:  constraintName,
			Value: constraintValue,
			Scope: child,
		})
	}

	return typeName, baseTypeName, constraints, nil
}

// extractConstraint extracts a constraint name-value pair from a constraint TypeScope.
// The constraint scope may contain more than 2 definitions when the constraint value
// is an identifier (which adds both an identifier definition and a constraint value definition)
// or when the constraint value is nested (e.g., Instant(day_of_week=Monday)), which adds
// the nested identifier and child constraint scopes.
func extractConstraint(constraintScope *symboltable.SymbolTable) (name, value string, err error) {
	if len(constraintScope.Definitions) < 2 {
		return "", "", fmt.Errorf("expected at least constraint name and value, got %d definitions", len(constraintScope.Definitions))
	}

	// Find the constraint name: the first definition with a non-nil Name
	for _, def := range constraintScope.Definitions {
		if def.Name != nil && name == "" {
			name = *def.Name
		}
		if def.Data != nil {
			strValue, ok := def.Data.(symboltable.StringData)
			if !ok {
				return "", "", fmt.Errorf("constraint value data is not a string, got %T", def.Data)
			}
			value = strValue.Value
		}
	}

	if name == "" {
		return "", "", fmt.Errorf("constraint name not found")
	}

	return name, value, nil
}

// ValidateIntervalTypes checks that Interval types reference valid Instant-derived types and units.
// This function operates on the analyzed types map after ComputeDerivationPaths has run.
// For each Interval type, it verifies that:
// - Parsed interval constraints are present
// - Named starts_at references exist in the types map
// - Named starts_at references derive from Instant (built-in)
// - Unit references exist in the units map
func ValidateIntervalTypes(types map[string]AnalysisTypeDefinition, units map[string]AnalysisUnitDefinition) error {
	for typeName, typeDef := range types {
		if len(typeDef.DerivationPath) == 0 {
			continue
		}

		// Check if this type derives from Interval
		builtInType := typeDef.DerivationPath[len(typeDef.DerivationPath)-1]
		if builtInType != "Interval" {
			continue
		}

		intervalConstraints, ok := typeDef.ParsedConstraints.(*IntervalConstraints)
		if !ok || intervalConstraints == nil {
			return fmt.Errorf("Interval type %s must specify interval constraints", typeName)
		}

		// If starts_at is a named type reference, verify it exists and is Instant-derived
		if intervalConstraints.StartTypeName != nil {
			startTypeName := *intervalConstraints.StartTypeName
			startTypeDef, exists := types[startTypeName]
			if !exists {
				return fmt.Errorf("Interval type %s references unknown starts_at type %s", typeName, startTypeName)
			}
			if len(startTypeDef.DerivationPath) == 0 {
				return fmt.Errorf("starts_at type %s for Interval type %s has no derivation path", startTypeName, typeName)
			}
			startBuiltIn := startTypeDef.DerivationPath[len(startTypeDef.DerivationPath)-1]
			if startBuiltIn != "Instant" {
				return fmt.Errorf("starts_at type %s for Interval type %s must derive from Instant, got %s", startTypeName, typeName, startBuiltIn)
			}
		}

		if intervalConstraints.UnitName != nil {
			unitName := *intervalConstraints.UnitName
			if _, exists := units[unitName]; !exists {
				return fmt.Errorf("Interval type %s references unknown unit %s", typeName, unitName)
			}
		}
	}
	return nil
}

// createTypeReference converts a base type name string to a TypeReference.
// It determines whether the base type is a built-in type or a declared type.
func createTypeReference(baseTypeName string) (TypeReference, error) {
	switch baseTypeName {
	case "Number":
		return NewBuiltInTypeRef(NumberType), nil
	case "Instant":
		return NewBuiltInTypeRef(InstantType), nil
	case "Interval":
		return NewBuiltInTypeRef(IntervalType), nil
	case "String":
		return NewBuiltInTypeRef(StringType), nil
	case "Boolean":
		return NewBuiltInTypeRef(BooleanType), nil
	default:
		// If it's not a known built-in type, treat it as a declared type
		return NewDeclaredTypeRef(baseTypeName), nil
	}
}
