// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file contains the TypeAnalyzer which handles type-related analysis
// including type registration, derivation path computation, and constraint validation.
package analysis

import "fmt"

// TypeAnalyzer handles type-related analysis operations.
//
// It maintains the types map and provides methods for type registration,
// derivation path computation, and constraint accumulation/validation.
type TypeAnalyzer struct {
	// types is a map from type name to its analyzed definition.
	// This map is built during analysis by processing type definitions
	// from the symbol table. It provides O(1) access to type information
	// for validation and code generation.
	types map[string]AnalysisTypeDefinition
}

// NewTypeAnalyzer creates a new TypeAnalyzer with initialized data structures.
//
// Returns:
//   - A pointer to a new, initialized TypeAnalyzer.
func NewTypeAnalyzer() *TypeAnalyzer {
	return &TypeAnalyzer{
		types: make(map[string]AnalysisTypeDefinition),
	}
}

// AddType adds a type definition to the analyzer's types map.
// This is primarily useful for testing, allowing direct population of the types map.
func (ta *TypeAnalyzer) AddType(typeDef AnalysisTypeDefinition) {
	ta.types[typeDef.Name] = typeDef
}

// GetTypes returns the types map.
func (ta *TypeAnalyzer) GetTypes() map[string]AnalysisTypeDefinition {
	return ta.types
}

// GetType returns the type definition with the given name.
// Returns a zero-value AnalysisTypeDefinition if the type is not found.
func (ta *TypeAnalyzer) GetType(name string) AnalysisTypeDefinition {
	return ta.types[name]
}

// HasType returns true if a type with the given name exists.
func (ta *TypeAnalyzer) HasType(name string) bool {
	_, exists := ta.types[name]
	return exists
}

// ComputeDerivationPaths computes the full derivation path for each type definition.
// This must be called after all type definitions have been collected, since types may
// be derived from other types that haven't been processed yet when EnterAnalyzeTypeScope
// is called.
//
// For each type, the path traces back through its base type chain to the built-in
// type it ultimately derives from. For example:
//   - Currency -> Number: ["Currency", "Number"]
//   - SmallCurrency -> Currency -> Number: ["SmallCurrency", "Currency", "Number"]
//
// Constraint accumulation and validation is also performed here. Constraints from
// the derivation chain are accumulated, validating that each type's constraints are
// only more restrictive than its base type's constraints.
func (ta *TypeAnalyzer) ComputeDerivationPaths(concepts map[string]AnalysisConceptDefinition) error {
	for typeName, typeDef := range ta.types {
		// First, compute the full derivation path - the terminal end is always a built-in type
		path, err := ta.computeTypeDerivationPath(typeName)
		if err != nil {
			// If we can't compute the derivation path, the terminal type isn't a built-in type
			return fmt.Errorf("type %s does not derive from a built-in type: %w", typeName, err)
		}

		// The built-in type is the last element of the derivation path
		builtInType, err := parseBuiltInTypeName(path[len(path)-1])
		if err != nil {
			return fmt.Errorf("invalid built-in type %q at terminal of derivation path for %s: %w",
				path[len(path)-1], typeName, err)
		}

		// Accumulate and validate constraints from the derivation path
		mergedConstraints, err := ta.accumulateConstraints(path, builtInType, concepts)
		if err != nil {
			return fmt.Errorf("constraint validation failed for type %s: %w", typeName, err)
		}

		typeDef.DerivationPath = path
		typeDef.ParsedConstraints = mergedConstraints
		ta.types[typeName] = typeDef
	}
	return nil
}

// accumulateConstraints walks the derivation path from most derived to built-in type,
// accumulating constraints and validating that each type's constraints are only
// more restrictive than its base type's constraints.
//
// The path goes from the derived type toward the built-in type (e.g., ["SmallCurrency", "Currency", "Number"]).
// For each type in the path, we:
// 1. Parse its raw constraints
// 2. Validate that each constraint is more restrictive than the inherited constraint
// 3. Merge by taking the more restrictive value
//
// Returns the final merged constraints after validating the entire chain.
func (ta *TypeAnalyzer) accumulateConstraints(path []string, builtInType BuiltInType, concepts map[string]AnalysisConceptDefinition) (TypeConstraints, error) {
	// Start with nil constraints (will be replaced as we traverse)
	var currentConstraints TypeConstraints = nil

	// Walk the path from the most derived type to the built-in type
	// The path is: [mostDerived, ..., baseType, builtInType]
	// We skip the last element (builtInType) since it has no constraints to parse
	for i := len(path) - 2; i >= 0; i-- {
		typeName := path[i]
		typeDef, exists := ta.types[typeName]
		if !exists {
			return nil, fmt.Errorf("type %s not found in types map", typeName)
		}

		// Parse this type's constraints. Interval roots require starts_at,
		// so they must be validated even when the raw constraint list is empty.
		var parsedConstraints TypeConstraints
		if len(typeDef.RawConstraints) > 0 || (builtInType == IntervalType && currentConstraints == nil) {
			var err error
			parsedConstraints, err = CreateConstraints(builtInType, typeDef.RawConstraints, concepts)
			if err != nil {
				return nil, fmt.Errorf("invalid constraints for type %s: %w", typeName, err)
			}
		}

		// If this type has constraints, validate against inherited constraints
		if parsedConstraints != nil {
			if currentConstraints != nil {
				// Validate that this type's constraints are more restrictive than inherited
				if err := validateMoreRestrictive(currentConstraints, parsedConstraints); err != nil {
					return nil, fmt.Errorf("type %s: %w", typeName, err)
				}
			}

			// Merge constraints, taking the more restrictive values
			var err error
			currentConstraints, err = mergeConstraints(currentConstraints, parsedConstraints)
			if err != nil {
				return nil, fmt.Errorf("type %s: %w", typeName, err)
			}
		}
	}

	return currentConstraints, nil
}

// computeTypeDerivationPath computes the derivation path for a single type by
// recursively tracing its base type chain until a built-in type is reached.
func (ta *TypeAnalyzer) computeTypeDerivationPath(typeName string) ([]string, error) {
	typeDef, exists := ta.types[typeName]
	if !exists {
		return nil, fmt.Errorf("type %s not found", typeName)
	}

	// Start the path with this type
	path := []string{typeName}

	// If the base type is built-in, we're done
	if typeDef.BaseType.IsBuiltIn() {
		path = append(path, typeDef.BaseType.GetName())
		return path, nil
	}

	// Otherwise, it's a declared type - recursively get its derivation path
	declaredRef, ok := typeDef.BaseType.(*DeclaredTypeRef)
	if !ok {
		return nil, fmt.Errorf("base type for %s is not a DeclaredTypeRef", typeName)
	}

	baseTypeName := declaredRef.GetName()
	basePath, err := ta.computeTypeDerivationPath(baseTypeName)
	if err != nil {
		return nil, err
	}

	// Add the base type name, then extend with the rest of the derivation chain
	path = append(path, baseTypeName)
	path = append(path, basePath[1:]...)
	return path, nil
}

// validateMoreRestrictive checks that newConstraints are not less restrictive than inheritedConstraints.
// For each constraint in newConstraints, it must be at least as restrictive as the corresponding
// constraint in inheritedConstraints. Equal values are allowed (constraints can be repeated), but
// less restrictive values (expansion) are not allowed.
//
// This uses the ConstraintNames() and GetConstraint() methods to iterate over constraints
// without hardcoding specific constraint names.
func validateMoreRestrictive(inheritedConstraints, newConstraints TypeConstraints) error {
	switch inherited := inheritedConstraints.(type) {
	case *NumberConstraints:
		new, ok := newConstraints.(*NumberConstraints)
		if !ok {
			return fmt.Errorf("new constraints must be NumberConstraints")
		}
		return validateNumberMoreRestrictive(inherited, new)
	case *InstantConstraints:
		new, ok := newConstraints.(*InstantConstraints)
		if !ok {
			return fmt.Errorf("new constraints must be InstantConstraints")
		}
		return validateInstantMoreRestrictive(inherited, new)
	case *IntervalConstraints:
		new, ok := newConstraints.(*IntervalConstraints)
		if !ok {
			return fmt.Errorf("new constraints must be IntervalConstraints")
		}
		return validateIntervalMoreRestrictive(inherited, new)
	case *StringConstraints:
		new, ok := newConstraints.(*StringConstraints)
		if !ok {
			return fmt.Errorf("new constraints must be StringConstraints")
		}
		return validateStringMoreRestrictive(inherited, new)
	case *BooleanConstraints:
		new, ok := newConstraints.(*BooleanConstraints)
		if !ok {
			return fmt.Errorf("new constraints must be BooleanConstraints")
		}
		return validateBooleanMoreRestrictive(inherited, new)
	default:
		return fmt.Errorf("unsupported constraint type for validation: %T", inheritedConstraints)
	}
}

// validateNumberMoreRestrictive validates Number constraint narrowing.
func validateNumberMoreRestrictive(inherited, new *NumberConstraints) error {
	for _, constraintName := range new.ConstraintNames() {
		newValue, _ := new.GetConstraint(constraintName)
		inheritedValue, exists := inherited.GetConstraint(constraintName)

		// Only check if the inherited constraint exists
		if exists {
			isMoreOrEqualRestrictive, err := inherited.IsMoreOrEqualRestrictive(constraintName, inheritedValue, newValue)
			if err != nil {
				return err
			}
			if !isMoreOrEqualRestrictive {
				// Format the values for the error message
				newInt := newValue.(*int)
				inheritedInt := inheritedValue.(*int)
				return fmt.Errorf("constraint '%s=%d' expands inherited constraint '%s=%d' (expansion not allowed)",
					constraintName, *newInt, constraintName, *inheritedInt)
			}
		}
	}
	return nil
}

// validateInstantMoreRestrictive validates Instant constraint narrowing.
// For Instant constraints, a derived type can only add constraints (nil -> value),
// not change existing ones (value -> different value) or remove them.
func validateInstantMoreRestrictive(inherited, new *InstantConstraints) error {
	for _, constraintName := range new.ConstraintNames() {
		newValue, _ := new.GetConstraint(constraintName)
		inheritedValue, exists := inherited.GetConstraint(constraintName)

		// Only check if the inherited constraint exists
		if exists {
			isMoreOrEqualRestrictive, err := inherited.IsMoreOrEqualRestrictive(constraintName, inheritedValue, newValue)
			if err != nil {
				return err
			}
			if !isMoreOrEqualRestrictive {
				return fmt.Errorf("constraint '%s' expands inherited constraint (expansion not allowed)", constraintName)
			}
		}
	}
	return nil
}

// validateIntervalMoreRestrictive validates Interval constraint narrowing.
// For Interval constraints, a derived type can only add constraints (nil -> value),
// not change existing ones (value -> different value) or remove them.
func validateIntervalMoreRestrictive(inherited, new *IntervalConstraints) error {
	for _, constraintName := range new.ConstraintNames() {
		newValue, _ := new.GetConstraint(constraintName)
		inheritedValue, exists := inherited.GetConstraint(constraintName)

		// Only check if the inherited constraint exists
		if exists {
			isMoreOrEqualRestrictive, err := inherited.IsMoreOrEqualRestrictive(constraintName, inheritedValue, newValue)
			if err != nil {
				return err
			}
			if !isMoreOrEqualRestrictive {
				return fmt.Errorf("constraint '%s' expands inherited constraint (expansion not allowed)", constraintName)
			}
		}
	}
	return nil
}

// mergeConstraints merges newConstraints into currentConstraints, returning the result.
// For each constraint in newConstraints, the more restrictive value is used.
func mergeConstraints(currentConstraints, newConstraints TypeConstraints) (TypeConstraints, error) {
	if currentConstraints == nil {
		return newConstraints, nil
	}
	if newConstraints == nil {
		return currentConstraints, nil
	}

	switch current := currentConstraints.(type) {
	case *NumberConstraints:
		new, ok := newConstraints.(*NumberConstraints)
		if !ok {
			return nil, fmt.Errorf("new constraints must be NumberConstraints")
		}
		return mergeNumberConstraints(current, new)
	case *InstantConstraints:
		new, ok := newConstraints.(*InstantConstraints)
		if !ok {
			return nil, fmt.Errorf("new constraints must be InstantConstraints")
		}
		return mergeInstantConstraints(current, new)
	case *IntervalConstraints:
		new, ok := newConstraints.(*IntervalConstraints)
		if !ok {
			return nil, fmt.Errorf("new constraints must be IntervalConstraints")
		}
		return mergeIntervalConstraints(current, new)
	case *StringConstraints:
		new, ok := newConstraints.(*StringConstraints)
		if !ok {
			return nil, fmt.Errorf("new constraints must be StringConstraints")
		}
		return mergeStringConstraints(current, new)
	case *BooleanConstraints:
		new, ok := newConstraints.(*BooleanConstraints)
		if !ok {
			return nil, fmt.Errorf("new constraints must be BooleanConstraints")
		}
		return mergeBooleanConstraints(current, new)
	default:
		return nil, fmt.Errorf("unsupported constraint type for merge: %T", currentConstraints)
	}
}

// mergeNumberConstraints merges two NumberConstraints, taking the more restrictive value for each field.
func mergeNumberConstraints(current, new *NumberConstraints) (*NumberConstraints, error) {
	result := &NumberConstraints{}

	// For each constraint, take the more restrictive value
	if new.Precision != nil {
		result.Precision = new.Precision
	} else {
		result.Precision = current.Precision
	}

	if new.Minimum != nil {
		result.Minimum = new.Minimum
	} else {
		result.Minimum = current.Minimum
	}

	if new.Maximum != nil {
		result.Maximum = new.Maximum
	} else {
		result.Maximum = current.Maximum
	}

	return result, nil
}

// mergeInstantConstraints merges two InstantConstraints, taking the union of all specified fields.
// Since validation ensures values don't change, we simply take the new value if present, else current.
func mergeInstantConstraints(current, new *InstantConstraints) (*InstantConstraints, error) {
	result := &InstantConstraints{}

	if new.Year != nil {
		result.Year = new.Year
	} else {
		result.Year = current.Year
	}

	if new.Month != nil {
		result.Month = new.Month
	} else {
		result.Month = current.Month
	}

	if new.DayOfMonth != nil {
		result.DayOfMonth = new.DayOfMonth
	} else {
		result.DayOfMonth = current.DayOfMonth
	}

	if new.DayOfYear != nil {
		result.DayOfYear = new.DayOfYear
	} else {
		result.DayOfYear = current.DayOfYear
	}

	if new.FiscalQuarter != nil {
		result.FiscalQuarter = new.FiscalQuarter
	} else {
		result.FiscalQuarter = current.FiscalQuarter
	}

	if new.Hour != nil {
		result.Hour = new.Hour
	} else {
		result.Hour = current.Hour
	}

	if new.Minute != nil {
		result.Minute = new.Minute
	} else {
		result.Minute = current.Minute
	}

	if new.Second != nil {
		result.Second = new.Second
	} else {
		result.Second = current.Second
	}

	if new.DayOfWeek != nil {
		result.DayOfWeek = new.DayOfWeek
	} else {
		result.DayOfWeek = current.DayOfWeek
	}

	if new.WeekNumber != nil {
		result.WeekNumber = new.WeekNumber
	} else {
		result.WeekNumber = current.WeekNumber
	}

	return result, nil
}

// mergeIntervalConstraints merges two IntervalConstraints.
func mergeIntervalConstraints(current, new *IntervalConstraints) (*IntervalConstraints, error) {
	result := &IntervalConstraints{}

	// Merge StartsAt: new takes precedence
	if new.StartsAt != nil {
		result.StartsAt = new.StartsAt
	} else {
		result.StartsAt = current.StartsAt
	}

	// Merge StartTypeName: new takes precedence
	if new.StartTypeName != nil {
		result.StartTypeName = new.StartTypeName
	} else {
		result.StartTypeName = current.StartTypeName
	}

	// Merge UnitName: new takes precedence
	if new.UnitName != nil {
		result.UnitName = new.UnitName
	} else {
		result.UnitName = current.UnitName
	}

	return result, nil
}

// validateStringMoreRestrictive validates String constraint narrowing.
func validateStringMoreRestrictive(inherited, new *StringConstraints) error {
	for _, constraintName := range new.ConstraintNames() {
		newValue, _ := new.GetConstraint(constraintName)
		inheritedValue, exists := inherited.GetConstraint(constraintName)
		if exists {
			isMoreOrEqualRestrictive, err := inherited.IsMoreOrEqualRestrictive(constraintName, inheritedValue, newValue)
			if err != nil {
				return err
			}
			if !isMoreOrEqualRestrictive {
				return fmt.Errorf("constraint '%s' expands inherited constraint (expansion not allowed)", constraintName)
			}
		}
	}
	return nil
}

// validateBooleanMoreRestrictive validates Boolean constraint narrowing.
func validateBooleanMoreRestrictive(inherited, new *BooleanConstraints) error {
	return nil
}

// mergeStringConstraints merges two StringConstraints, taking the more restrictive value for each field.
func mergeStringConstraints(current, new *StringConstraints) (*StringConstraints, error) {
	if current == nil {
		return new, nil
	}
	if new == nil {
		return current, nil
	}

	result := &StringConstraints{}

	if new.Minimum != nil {
		result.Minimum = new.Minimum
	} else {
		result.Minimum = current.Minimum
	}

	if new.Maximum != nil {
		result.Maximum = new.Maximum
	} else {
		result.Maximum = current.Maximum
	}

	if new.Pattern != nil {
		result.Pattern = new.Pattern
	} else {
		result.Pattern = current.Pattern
	}

	return result, nil
}

// mergeBooleanConstraints merges two BooleanConstraints, taking the more restrictive value for each field.
func mergeBooleanConstraints(current, new *BooleanConstraints) (*BooleanConstraints, error) {
	if current == nil {
		return new, nil
	}
	if new == nil {
		return current, nil
	}
	return &BooleanConstraints{}, nil
}

// parseBuiltInTypeName converts a built-in type name string to a BuiltInType.
// Returns an error if the name doesn't match any known built-in type.
func parseBuiltInTypeName(name string) (BuiltInType, error) {
	switch name {
	case "Number":
		return NumberType, nil
	case "Instant":
		return InstantType, nil
	case "Interval":
		return IntervalType, nil
	case "String":
		return StringType, nil
	case "Boolean":
		return BooleanType, nil
	default:
		return NumberType, fmt.Errorf("unknown built-in type: %s", name)
	}
}
