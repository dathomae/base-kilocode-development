// Package analysis provides the analysis phase types for the Chronocone Planning Language.
//
// This file contains analysis types related to type definitions.
// These types represent the analysis-level interpretation of parsed type definitions.
package analysis

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// BuiltInType represents the base types available in CPL.
// Supports Number, Instant, Interval, String, and Boolean.
type BuiltInType int

const (
	// NumberType represents the numeric base type in CPL.
	// Example: Currency = Number(precision=2)
	NumberType BuiltInType = iota

	// InstantType represents the instant (point-in-time) base type in CPL.
	// Example: Timestamp = Instant(second=0)
	InstantType

	// IntervalType represents the interval (duration) base type in CPL.
	// Example: FiscalMonth = Interval(starts_at=FiscalYearStart, unit=month)
	IntervalType

	// StringType represents the string base type in CPL.
	StringType

	// BooleanType represents the boolean base type in CPL.
	BooleanType
)

// String returns the string representation of BuiltInType.
// This implements the fmt.Stringer interface.
func (b BuiltInType) String() string {
	switch b {
	case NumberType:
		return "Number"
	case InstantType:
		return "Instant"
	case IntervalType:
		return "Interval"
	case StringType:
		return "String"
	case BooleanType:
		return "Boolean"
	default:
		return "Unknown"
	}
}

// TypeReference is an interface that represents a reference to a type.
// It can be either a built-in type (like Number) or a user-declared type.
// This allows the type system to support type derivation where types can
// be based on other types, not just built-in types.
type TypeReference interface {
	// GetName returns the name of the type being referenced.
	GetName() string

	// IsBuiltIn returns true if this is a reference to a built-in type.
	IsBuiltIn() bool

	// GetBuiltInType returns the BuiltInType if IsBuiltIn() is true.
	// Returns NumberType as a default if IsBuiltIn() is false.
	GetBuiltInType() BuiltInType
}

// BuiltInTypeRef is a TypeReference for built-in types like Number.
type BuiltInTypeRef struct {
	builtinType BuiltInType
}

// NewBuiltInTypeRef creates a new BuiltInTypeRef from a BuiltInType.
func NewBuiltInTypeRef(builtinType BuiltInType) *BuiltInTypeRef {
	return &BuiltInTypeRef{builtinType: builtinType}
}

// GetName returns the string name of the built-in type.
func (b *BuiltInTypeRef) GetName() string {
	return b.builtinType.String()
}

// IsBuiltIn returns true since this is a built-in type reference.
func (b *BuiltInTypeRef) IsBuiltIn() bool {
	return true
}

// GetBuiltInType returns the underlying BuiltInType.
func (b *BuiltInTypeRef) GetBuiltInType() BuiltInType {
	return b.builtinType
}

// DeclaredTypeRef is a TypeReference for user-declared types.
// Example: "Currency" in "SmallCurrency = Currency(precision=0)"
type DeclaredTypeRef struct {
	typeName string
}

// NewDeclaredTypeRef creates a new DeclaredTypeRef for a declared type.
func NewDeclaredTypeRef(typeName string) *DeclaredTypeRef {
	return &DeclaredTypeRef{typeName: typeName}
}

// GetName returns the name of the declared type.
func (d *DeclaredTypeRef) GetName() string {
	return d.typeName
}

// IsBuiltIn returns false since this is a declared type reference.
func (d *DeclaredTypeRef) IsBuiltIn() bool {
	return false
}

// GetBuiltInType returns NumberType as a default.
// The actual type should be resolved by walking the derivation chain.
func (d *DeclaredTypeRef) GetBuiltInType() BuiltInType {
	return NumberType
}

// Constraint represents a raw constraint specification from the parser.
// Constraints are captured as name-value pairs during parsing.
// Example: precision=2, min=0, max=100
type Constraint struct {
	Name  string
	Value string
	// Scope holds the symbol table scope for nested constraint values.
	// Set when the constraint value contains child scopes (e.g., Instant(field=value)).
	Scope *symboltable.SymbolTable
}

// NumberConstraints represents optional constraints for Number type definitions.
// All fields are pointers to allow distinguishing between "not specified" and "zero".
type NumberConstraints struct {
	// Precision specifies the number of decimal places.
	// Example: Number(precision=2) allows values like 12.34
	Precision *int

	// Minimum specifies the minimum allowed value (inclusive).
	// Example: Number(min=0) allows values >= 0
	Minimum *int

	// Maximum specifies the maximum allowed value (inclusive).
	// Example: Number(max=100) allows values <= 100
	Maximum *int
}

// InstantConstraints holds field-based constraints for Instant types.
// Every supplied field is a constraint on matching UTC moments.
// Omitted fields are valid and mean underspecification.
//
// The complete set of constraint names is: year, month, day_of_month,
// day_of_year, week_number, fiscal_quarter, day_of_week, hour, minute, second.
//
// Example: Instant(year=2024, month=January, day_of_month=1)
type InstantConstraints struct {
	Year          *int    // e.g., 2025
	Month         *string // e.g., "January"
	DayOfMonth    *int    // e.g., 14
	DayOfYear     *int    // e.g., 42
	FiscalQuarter *int    // e.g., 1
	Hour          *int    // e.g., 9
	Minute        *int    // e.g., 30
	Second        *int    // e.g., 0
	DayOfWeek     *string // e.g., "Monday"
	WeekNumber    *int    // e.g., 1
}

// IntervalConstraints holds constraints for Interval types.
// An Interval type requires a starts_at reference (either an inline
// InstantConstraints or a named Instant-derived type) and a unit reference.
//
// Example: Interval(starts_at=FiscalYearStart, unit=month)
// Example: Interval(starts_at=Instant(month=January, day=1), unit=day)
type IntervalConstraints struct {
	// StartsAt holds inline Instant constraints for the interval start.
	// Used when starts_at is specified with inline Instant constraints.
	StartsAt *InstantConstraints

	// StartTypeName holds the name of a declared Instant-derived type.
	// Used when starts_at references a named type.
	StartTypeName *string

	// UnitName holds the name of the unit for this interval.
	// Example: "month", "day", "year"
	UnitName *string
}

// StringConstraints represents optional constraints for String type definitions.
// All fields are pointers to allow distinguishing between "not specified" and "zero".
type StringConstraints struct {
	// Minimum specifies the minimum length of the string (inclusive).
	// Example: String(minimum=1) requires at least 1 character
	Minimum *int

	// Maximum specifies the maximum length of the string (inclusive).
	// Example: String(maximum=100) allows at most 100 characters
	Maximum *int

	// Pattern specifies a regular expression the string must match.
	// Example: String(pattern="[a-z]+") allows only lowercase letters
	Pattern *regexp.Regexp
}

// BooleanConstraints represents optional constraints for Boolean type definitions.
// Boolean types currently have no configurable constraints.
type BooleanConstraints struct {
}

// isTypeConstraints implements the TypeConstraints interface marker method.
// This method exists only to satisfy the interface and should not be called directly.
func (n NumberConstraints) isTypeConstraints() {}

// isTypeConstraints implements the TypeConstraints interface marker method.
// This method exists only to satisfy the interface and should not be called directly.
func (i InstantConstraints) isTypeConstraints() {}

// isTypeConstraints implements the TypeConstraints interface marker method.
// This method exists only to satisfy the interface and should not be called directly.
func (i IntervalConstraints) isTypeConstraints() {}

// isTypeConstraints implements the TypeConstraints interface marker method.
// This method exists only to satisfy the interface and should not be called directly.
func (s StringConstraints) isTypeConstraints() {}

// isTypeConstraints implements the TypeConstraints interface marker method.
// This method exists only to satisfy the interface and should not be called directly.
func (b BooleanConstraints) isTypeConstraints() {}

// Validate checks Number constraint consistency.
// Returns an error if:
// - Precision is negative
// - Both minimum and maximum are specified and minimum >= maximum
func (n NumberConstraints) Validate() error {
	if n.Precision != nil && *n.Precision < 0 {
		return fmt.Errorf("precision must be non-negative, got %d", *n.Precision)
	}
	if n.Minimum != nil && n.Maximum != nil && *n.Minimum >= *n.Maximum {
		return fmt.Errorf("minimum (%d) must be less than maximum (%d)", *n.Minimum, *n.Maximum)
	}
	return nil
}

// instantConstraintToConceptName maps Instant constraint names to their
// corresponding concept names for concept-driven validation.
var instantConstraintToConceptName = map[string]string{
	"year":           "Year",
	"month":          "Month",
	"day_of_month":   "DayOfMonth",
	"day_of_year":    "DayOfYear",
	"week_number":    "WeekNumber",
	"fiscal_quarter": "FiscalQuarter",
	"day_of_week":    "DayOfWeek",
	"hour":           "Hour",
	"minute":         "Minute",
	"second":         "Second",
}

// Validate checks Interval constraint consistency.
// Returns an error if:
// - Neither starts_at (inline or named) is specified
// - Both inline and named starts_at are specified
func (i IntervalConstraints) Validate() error {
	if i.StartsAt == nil && i.StartTypeName == nil {
		return fmt.Errorf("Interval type must specify starts_at as either an inline Instant constraint or a named Instant-derived type")
	}
	if i.StartsAt != nil && i.StartTypeName != nil {
		return fmt.Errorf("Interval starts_at cannot be both an inline Instant constraint and a named type reference")
	}
	return nil
}

// ConstraintNames returns a list of all constraint names that have values set.
func (i IntervalConstraints) ConstraintNames() []string {
	names := []string{}
	if i.StartsAt != nil || i.StartTypeName != nil {
		names = append(names, "starts_at")
	}
	if i.UnitName != nil {
		names = append(names, "unit")
	}
	return names
}

// GetConstraint returns the value for a given constraint name.
// Returns (value, true) if the constraint exists, or (nil, false) if not.
func (i IntervalConstraints) GetConstraint(name string) (interface{}, bool) {
	switch name {
	case "starts_at":
		if i.StartsAt != nil {
			return i.StartsAt, true
		}
		if i.StartTypeName != nil {
			return i.StartTypeName, true
		}
		return nil, false
	case "unit":
		return i.UnitName, i.UnitName != nil
	default:
		return nil, false
	}
}

// instantConstraintsEqual returns true if two InstantConstraints have the same
// values for all fields. Nil fields are considered equal to nil fields.
func instantConstraintsEqual(a, b *InstantConstraints) bool {
	if a == nil || b == nil {
		return a == b
	}
	if !intPtrEqual(a.Year, b.Year) {
		return false
	}
	if !stringPtrEqual(a.Month, b.Month) {
		return false
	}
	if !intPtrEqual(a.DayOfMonth, b.DayOfMonth) {
		return false
	}
	if !intPtrEqual(a.DayOfYear, b.DayOfYear) {
		return false
	}
	if !intPtrEqual(a.FiscalQuarter, b.FiscalQuarter) {
		return false
	}
	if !intPtrEqual(a.Hour, b.Hour) {
		return false
	}
	if !intPtrEqual(a.Minute, b.Minute) {
		return false
	}
	if !intPtrEqual(a.Second, b.Second) {
		return false
	}
	if !stringPtrEqual(a.DayOfWeek, b.DayOfWeek) {
		return false
	}
	if !intPtrEqual(a.WeekNumber, b.WeekNumber) {
		return false
	}
	return true
}

// intPtrEqual returns true if two *int values are equal (both nil or same value).
func intPtrEqual(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// stringPtrEqual returns true if two *string values are equal (both nil or same value).
func stringPtrEqual(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// IsMoreOrEqualRestrictive compares two constraint values and returns true if
// newValue is more restrictive OR equal to currentValue.
//
// For IntervalConstraints, a constraint is more restrictive when it changes from
// unspecified (nil) to specified (non-nil). Changing a specified value to a
// different value is not allowed (returns false). Removing a constraint
// (non-nil to nil) is less restrictive (returns false).
func (i IntervalConstraints) IsMoreOrEqualRestrictive(constraintName string, currentValue, newValue interface{}) (bool, error) {
	switch constraintName {
	case "starts_at":
		switch current := currentValue.(type) {
		case *string:
			new, ok := newValue.(*string)
			if !ok {
				return false, fmt.Errorf("newValue for starts_at must be *string when currentValue is *string, got %T", newValue)
			}
			if current == nil {
				return true, nil
			}
			if new == nil {
				return false, nil
			}
			return *new == *current, nil
		case *InstantConstraints:
			new, ok := newValue.(*InstantConstraints)
			if !ok {
				return false, fmt.Errorf("newValue for starts_at must be *InstantConstraints when currentValue is *InstantConstraints, got %T", newValue)
			}
			if current == nil {
				return true, nil
			}
			if new == nil {
				return false, nil
			}
			return instantConstraintsEqual(current, new), nil
		default:
			return false, fmt.Errorf("currentValue for starts_at must be *string or *InstantConstraints, got %T", currentValue)
		}
	case "unit":
		current, ok := currentValue.(*string)
		if !ok {
			return false, fmt.Errorf("currentValue for unit must be *string, got %T", currentValue)
		}
		new, ok := newValue.(*string)
		if !ok {
			return false, fmt.Errorf("newValue for unit must be *string, got %T", newValue)
		}
		if current == nil {
			return true, nil
		}
		if new == nil {
			return false, nil
		}
		return *new == *current, nil
	default:
		return false, fmt.Errorf("unknown constraint name: %s", constraintName)
	}
}

// Validate checks Instant constraint consistency.
// Validation is now performed during constraint creation via concept lookup,
// so this method returns nil. It is kept to satisfy the TypeConstraints interface.
func (i InstantConstraints) Validate() error {
	return nil
}

// ConstraintNames returns a list of all constraint names that have values set.
func (i InstantConstraints) ConstraintNames() []string {
	names := []string{}
	if i.Year != nil {
		names = append(names, "year")
	}
	if i.Month != nil {
		names = append(names, "month")
	}
	if i.DayOfMonth != nil {
		names = append(names, "day_of_month")
	}
	if i.DayOfYear != nil {
		names = append(names, "day_of_year")
	}
	if i.FiscalQuarter != nil {
		names = append(names, "fiscal_quarter")
	}
	if i.Hour != nil {
		names = append(names, "hour")
	}
	if i.Minute != nil {
		names = append(names, "minute")
	}
	if i.Second != nil {
		names = append(names, "second")
	}
	if i.DayOfWeek != nil {
		names = append(names, "day_of_week")
	}
	if i.WeekNumber != nil {
		names = append(names, "week_number")
	}
	return names
}

// GetConstraint returns the value for a given constraint name.
// Returns (value, true) if the constraint exists, or (nil, false) if not.
func (i InstantConstraints) GetConstraint(name string) (interface{}, bool) {
	switch name {
	case "year":
		return i.Year, i.Year != nil
	case "month":
		return i.Month, i.Month != nil
	case "day_of_month":
		return i.DayOfMonth, i.DayOfMonth != nil
	case "day_of_year":
		return i.DayOfYear, i.DayOfYear != nil
	case "fiscal_quarter":
		return i.FiscalQuarter, i.FiscalQuarter != nil
	case "hour":
		return i.Hour, i.Hour != nil
	case "minute":
		return i.Minute, i.Minute != nil
	case "second":
		return i.Second, i.Second != nil
	case "day_of_week":
		return i.DayOfWeek, i.DayOfWeek != nil
	case "week_number":
		return i.WeekNumber, i.WeekNumber != nil
	default:
		return nil, false
	}
}

// IsMoreOrEqualRestrictive compares two constraint values and returns true if
// newValue is more restrictive OR equal to currentValue.
//
// For InstantConstraints, a constraint is more restrictive when it changes from
// unspecified (nil) to specified (non-nil). Changing a specified value to a
// different value is not allowed (returns false). Removing a constraint
// (non-nil to nil) is less restrictive (returns false).
//
// Both currentValue and newValue must be *int or *string depending on the
// constraint name, or an error is returned.
func (i InstantConstraints) IsMoreOrEqualRestrictive(constraintName string, currentValue, newValue interface{}) (bool, error) {
	switch constraintName {
	case "year", "day_of_month", "day_of_year", "fiscal_quarter", "hour", "minute", "second", "week_number":
		current, ok := currentValue.(*int)
		if !ok {
			return false, fmt.Errorf("currentValue for %s must be *int", constraintName)
		}
		new, ok := newValue.(*int)
		if !ok {
			return false, fmt.Errorf("newValue for %s must be *int", constraintName)
		}
		if current == nil {
			return true, nil
		}
		if new == nil {
			return false, nil
		}
		return *new == *current, nil
	case "month", "day_of_week":
		current, ok := currentValue.(*string)
		if !ok {
			return false, fmt.Errorf("currentValue for %s must be *string", constraintName)
		}
		new, ok := newValue.(*string)
		if !ok {
			return false, fmt.Errorf("newValue for %s must be *string", constraintName)
		}
		if current == nil {
			return true, nil
		}
		if new == nil {
			return false, nil
		}
		return *new == *current, nil
	default:
		return false, fmt.Errorf("unknown constraint name: %s", constraintName)
	}
}

// IsMoreRestrictive compares two constraint values and returns true if
// newValue is more restrictive than currentValue.
//
// For NumberConstraints:
//   - "precision": lower value is MORE restrictive (e.g., 2 is more restrictive than 4)
//   - "minimum": higher value is MORE restrictive (e.g., 10 is more restrictive than 0)
//   - "maximum": lower value is MORE restrictive (e.g., 100 is more restrictive than 1000)
//
// Both currentValue and newValue must be *int or an error is returned.
func (n NumberConstraints) IsMoreRestrictive(constraintName string, currentValue, newValue interface{}) (bool, error) {
	current, ok := currentValue.(*int)
	if !ok {
		return false, fmt.Errorf("currentValue for %s must be *int", constraintName)
	}
	new, ok := newValue.(*int)
	if !ok {
		return false, fmt.Errorf("newValue for %s must be *int", constraintName)
	}

	// If current is nil, anything is more restrictive
	if current == nil {
		return new != nil, nil
	}
	// If new is nil, nothing is more restrictive
	if new == nil {
		return false, nil
	}

	switch constraintName {
	case "precision":
		// Lower precision is more restrictive (fewer decimal places allowed)
		return *new < *current, nil
	case "minimum":
		// Higher minimum is more restrictive (value must be at least a larger number)
		return *new > *current, nil
	case "maximum":
		// Lower maximum is more restrictive (value must be at most a smaller number)
		return *new < *current, nil
	default:
		return false, fmt.Errorf("unknown constraint name: %s", constraintName)
	}
}

// IsMoreOrEqualRestrictive compares two constraint values and returns true if
// newValue is more restrictive OR equal to currentValue.
//
// This is used to validate that constraints don't expand (become less restrictive).
// A constraint can be declared with the same value as the base type's constraint -
// this is allowed, not an error.
//
// For NumberConstraints:
//   - "precision": lower or equal value is more or equal restrictive
//   - "minimum": higher or equal value is more or equal restrictive
//   - "maximum": lower or equal value is more or equal restrictive
//
// Both currentValue and newValue must be *int or an error is returned.
func (n NumberConstraints) IsMoreOrEqualRestrictive(constraintName string, currentValue, newValue interface{}) (bool, error) {
	current, ok := currentValue.(*int)
	if !ok {
		return false, fmt.Errorf("currentValue for %s must be *int", constraintName)
	}
	new, ok := newValue.(*int)
	if !ok {
		return false, fmt.Errorf("newValue for %s must be *int", constraintName)
	}

	// If current is nil, anything is acceptable (no constraint to check against)
	if current == nil {
		return true, nil
	}
	// If new is nil, it's less restrictive (no constraint specified)
	if new == nil {
		return false, nil
	}

	switch constraintName {
	case "precision":
		// Lower or equal precision is acceptable (0 <= 0, 0 <= 2)
		return *new <= *current, nil
	case "minimum":
		// Higher or equal minimum is acceptable (10 >= 10, 10 >= 0)
		return *new >= *current, nil
	case "maximum":
		// Lower or equal maximum is acceptable (100 <= 100, 100 <= 1000)
		return *new <= *current, nil
	default:
		return false, fmt.Errorf("unknown constraint name: %s", constraintName)
	}
}

// ConstraintNames returns a list of all constraint names that have values set.
func (n NumberConstraints) ConstraintNames() []string {
	names := []string{}
	if n.Precision != nil {
		names = append(names, "precision")
	}
	if n.Minimum != nil {
		names = append(names, "minimum")
	}
	if n.Maximum != nil {
		names = append(names, "maximum")
	}
	return names
}

// GetConstraint returns the value for a given constraint name.
// Returns (value, true) if the constraint exists, or (nil, false) if not.
func (n NumberConstraints) GetConstraint(name string) (interface{}, bool) {
	switch name {
	case "precision":
		return n.Precision, n.Precision != nil
	case "minimum":
		return n.Minimum, n.Minimum != nil
	case "maximum":
		return n.Maximum, n.Maximum != nil
	default:
		return nil, false
	}
}

// Validate checks String constraint consistency.
// Returns an error if:
// - Minimum is negative
// - Maximum is negative
// - Both minimum and maximum are specified and minimum > maximum
func (s StringConstraints) Validate() error {
	if s.Minimum != nil && *s.Minimum < 0 {
		return fmt.Errorf("minimum must be non-negative, got %d", *s.Minimum)
	}
	if s.Maximum != nil && *s.Maximum < 0 {
		return fmt.Errorf("maximum must be non-negative, got %d", *s.Maximum)
	}
	if s.Minimum != nil && s.Maximum != nil && *s.Minimum > *s.Maximum {
		return fmt.Errorf("minimum (%d) must be less than or equal to maximum (%d)", *s.Minimum, *s.Maximum)
	}
	return nil
}

// ConstraintNames returns a list of all constraint names that have values set.
func (s StringConstraints) ConstraintNames() []string {
	names := []string{}
	if s.Minimum != nil {
		names = append(names, "minimum")
	}
	if s.Maximum != nil {
		names = append(names, "maximum")
	}
	if s.Pattern != nil {
		names = append(names, "pattern")
	}
	return names
}

// GetConstraint returns the value for a given constraint name.
// Returns (value, true) if the constraint exists, or (nil, false) if not.
func (s StringConstraints) GetConstraint(name string) (interface{}, bool) {
	switch name {
	case "minimum":
		return s.Minimum, s.Minimum != nil
	case "maximum":
		return s.Maximum, s.Maximum != nil
	case "pattern":
		return s.Pattern, s.Pattern != nil
	default:
		return nil, false
	}
}

// IsMoreOrEqualRestrictive compares two constraint values and returns true if
// newValue is more restrictive OR equal to currentValue.
//
// For StringConstraints:
//   - "minimum": higher value is more restrictive
//   - "maximum": lower value is more restrictive
//   - "pattern": any non-nil pattern is more restrictive than nil; same pattern is equal
//
// Both currentValue and newValue must be *int for minimum/maximum or
// *regexp.Regexp for pattern, or an error is returned.
func (s StringConstraints) IsMoreOrEqualRestrictive(constraintName string, currentValue, newValue interface{}) (bool, error) {
	switch constraintName {
	case "minimum":
		current, ok := currentValue.(*int)
		if !ok {
			return false, fmt.Errorf("currentValue for %s must be *int", constraintName)
		}
		new, ok := newValue.(*int)
		if !ok {
			return false, fmt.Errorf("newValue for %s must be *int", constraintName)
		}
		if current == nil {
			return true, nil
		}
		if new == nil {
			return false, nil
		}
		return *new >= *current, nil
	case "maximum":
		current, ok := currentValue.(*int)
		if !ok {
			return false, fmt.Errorf("currentValue for %s must be *int", constraintName)
		}
		new, ok := newValue.(*int)
		if !ok {
			return false, fmt.Errorf("newValue for %s must be *int", constraintName)
		}
		if current == nil {
			return true, nil
		}
		if new == nil {
			return false, nil
		}
		return *new <= *current, nil
	case "pattern":
		current, ok := currentValue.(*regexp.Regexp)
		if !ok {
			return false, fmt.Errorf("currentValue for %s must be *regexp.Regexp", constraintName)
		}
		new, ok := newValue.(*regexp.Regexp)
		if !ok {
			return false, fmt.Errorf("newValue for %s must be *regexp.Regexp", constraintName)
		}
		if current == nil {
			return true, nil
		}
		if new == nil {
			return false, nil
		}
		return current.String() == new.String(), nil
	default:
		return false, fmt.Errorf("unknown constraint name: %s", constraintName)
	}
}

// Validate checks Boolean constraint consistency.
// Boolean types have no constraints, so this always returns nil.
func (b BooleanConstraints) Validate() error {
	return nil
}

// ConstraintNames returns a list of all constraint names that have values set.
// BooleanConstraints has no constraints, so this returns an empty slice.
func (b BooleanConstraints) ConstraintNames() []string {
	return []string{}
}

// GetConstraint returns the value for a given constraint name.
// BooleanConstraints has no constraints, so this always returns (nil, false).
func (b BooleanConstraints) GetConstraint(name string) (interface{}, bool) {
	return nil, false
}

// IsMoreOrEqualRestrictive compares two constraint values and returns true if
// newValue is more restrictive OR equal to currentValue.
// BooleanConstraints has no constraints, so this always returns an error for any constraint name.
func (b BooleanConstraints) IsMoreOrEqualRestrictive(constraintName string, currentValue, newValue interface{}) (bool, error) {
	return false, fmt.Errorf("unknown constraint name: %s", constraintName)
}

// TypeDefinition represents a type definition in the parser.
// A type definition associates a name with a base type and optional constraints.
// Example: Currency = Number(precision=2)
type TypeDefinition struct {
	Name        string
	BuiltInType BuiltInType
	Constraints []Constraint
}

// TypeConstraints is a sum type interface for different constraint types.
// Each constraint type implements isTypeConstraints() and Validate() methods.
type TypeConstraints interface {
	isTypeConstraints()
	Validate() error
}

// ConstraintComparator defines the interface for comparing constraint values
// to determine if a new constraint is more restrictive than a current one.
//
// This interface is used when validating type derivation chains to ensure
// that each level only makes constraints more restrictive, not less restrictive.
//
// The constraintName parameter identifies which constraint is being compared
// (e.g., "precision", "minimum", "maximum").
//
// The currentValue and newValue parameters are the values being compared.
// For NumberConstraints, these are *int values.
//
// Returns:
//   - true if newValue is more restrictive than currentValue
//   - false if newValue is less restrictive or equal to currentValue
//   - error if the comparison cannot be performed (e.g., incompatible types)
type ConstraintComparator interface {
	// IsMoreRestrictive compares two constraint values and returns true if
	// newValue is more restrictive than currentValue.
	IsMoreRestrictive(constraintName string, currentValue, newValue interface{}) (bool, error)
}

// AnalysisTypeDefinition contains the analyzed type information.
// This structure is built during the analysis phase from parser type definitions.
type AnalysisTypeDefinition struct {
	Name              string
	BaseType          TypeReference   // Can be BuiltInTypeRef or DeclaredTypeRef
	RawConstraints    []Constraint    // Raw constraints as parsed
	ParsedConstraints TypeConstraints // Validated constraints (populated in ComputeDerivationPaths)
	DerivationPath    []string        // Chain from this type to built-in root
}

// ParserTypeDefinition wraps a TypeDefinition with additional metadata
// for the parser-to-analysis handoff.
type ParserTypeDefinition struct {
	TypeDefinition
}

// CreateConstraints converts parser constraints to analysis constraints.
// Performs semantic validation including duplicate constraint name checking and
// value type validation that was removed from the parsing phase.
//
// For Instant and Interval types, the concepts map provides the predefined
// concept definitions used for concept-driven constraint validation.
func CreateConstraints(baseType BuiltInType, parserConstraints []Constraint, concepts map[string]AnalysisConceptDefinition) (TypeConstraints, error) {
	// Check for duplicate constraint names
	seenConstraints := make(map[string]bool)
	for _, pc := range parserConstraints {
		if seenConstraints[pc.Name] {
			return nil, fmt.Errorf("duplicate constraint name: %s", pc.Name)
		}
		seenConstraints[pc.Name] = true
	}

	switch baseType {
	case NumberType:
		return createNumberConstraints(parserConstraints)
	case InstantType:
		return createInstantConstraints(parserConstraints, concepts)
	case IntervalType:
		return createIntervalConstraints(parserConstraints, concepts)
	case StringType:
		return createStringConstraints(parserConstraints)
	case BooleanType:
		return createBooleanConstraints(parserConstraints)
	default:
		return nil, fmt.Errorf("unsupported base type: %v", baseType)
	}
}

// createNumberConstraints converts parser Number constraints to NumberConstraints.
// Validates that constraint values are valid integers.
func createNumberConstraints(parserConstraints []Constraint) (*NumberConstraints, error) {
	nc := &NumberConstraints{}

	for _, pc := range parserConstraints {
		switch pc.Name {
		case "precision":
			// Value type validation: precision must be an integer
			v, err := strconv.Atoi(pc.Value)
			if err != nil {
				return nil, fmt.Errorf("precision constraint must be an integer, got %s", pc.Value)
			}
			nc.Precision = &v
		case "minimum":
			// Value type validation: minimum must be an integer
			v, err := strconv.Atoi(pc.Value)
			if err != nil {
				return nil, fmt.Errorf("minimum constraint must be an integer, got %s", pc.Value)
			}
			nc.Minimum = &v
		case "maximum":
			// Value type validation: maximum must be an integer
			v, err := strconv.Atoi(pc.Value)
			if err != nil {
				return nil, fmt.Errorf("maximum constraint must be an integer, got %s", pc.Value)
			}
			nc.Maximum = &v
		default:
			return nil, fmt.Errorf("unknown constraint for Number: %s", pc.Name)
		}
	}

	// Semantic validation: constraint consistency (e.g., min < max)
	if err := nc.Validate(); err != nil {
		return nil, err
	}

	return nc, nil
}

// validInstantConstraintNames defines the set of valid constraint names for Instant types.
//
// The complete set of 10 recognized names is: year, month, day_of_month,
// day_of_year, week_number, fiscal_quarter, day_of_week, hour, minute, second.
var validInstantConstraintNames = map[string]bool{
	"year":           true,
	"month":          true,
	"day_of_month":   true,
	"day_of_year":    true,
	"week_number":    true,
	"fiscal_quarter": true,
	"day_of_week":    true,
	"hour":           true,
	"minute":         true,
	"second":         true,
}

// createInstantConstraints converts parser Instant constraints to InstantConstraints.
// Validates that constraint names are valid field names and values are correct
// by looking up corresponding concepts in the concepts map.
func createInstantConstraints(parserConstraints []Constraint, concepts map[string]AnalysisConceptDefinition) (*InstantConstraints, error) {
	ic := &InstantConstraints{}

	for _, pc := range parserConstraints {
		if !validInstantConstraintNames[pc.Name] {
			return nil, fmt.Errorf("unknown constraint for Instant: %s", pc.Name)
		}

		conceptName, ok := instantConstraintToConceptName[pc.Name]
		if !ok {
			return nil, fmt.Errorf("unknown constraint for Instant: %s", pc.Name)
		}

		concept, conceptExists := concepts[conceptName]
		if !conceptExists {
			return nil, fmt.Errorf("required concept %s for constraint %s is not defined", conceptName, pc.Name)
		}

		switch pc.Name {
		case "year", "day_of_month", "day_of_year", "fiscal_quarter", "hour", "minute", "second", "week_number":
			v, err := strconv.Atoi(pc.Value)
			if err != nil {
				// Concept names as constraint values mean "unconstrained" (use the full concept range)
				if pc.Value == conceptName {
					continue
				}
				return nil, fmt.Errorf("%s constraint must be an integer, got %s", pc.Name, pc.Value)
			}

			// Validate against concept's Int range bounds
			start, end, hasBounds := concept.GetIntRangeBounds()
			if !hasBounds {
				return nil, fmt.Errorf("concept %s for constraint %s does not have Int range bounds", conceptName, pc.Name)
			}
			if v < start || v > end {
				return nil, fmt.Errorf("%s must be between %d and %d, got %d", pc.Name, start, end, v)
			}

			switch pc.Name {
			case "year":
				ic.Year = &v
			case "day_of_month":
				ic.DayOfMonth = &v
			case "day_of_year":
				ic.DayOfYear = &v
			case "fiscal_quarter":
				ic.FiscalQuarter = &v
			case "hour":
				ic.Hour = &v
			case "minute":
				ic.Minute = &v
			case "second":
				ic.Second = &v
			case "week_number":
				ic.WeekNumber = &v
			}
		case "month", "day_of_week":
			// Concept names as constraint values mean "unconstrained" (use the full concept range)
			if pc.Value == conceptName {
				continue
			}
			// Validate against concept's values
			if !concept.ContainsValue(pc.Value) {
				return nil, fmt.Errorf("invalid %s: %s", pc.Name, pc.Value)
			}
			switch pc.Name {
			case "month":
				ic.Month = &pc.Value
			case "day_of_week":
				ic.DayOfWeek = &pc.Value
			}
		}
	}

	return ic, nil
}

// validIntervalConstraintNames defines the set of valid constraint names for Interval types.
var validIntervalConstraintNames = map[string]bool{
	"starts_at": true,
	"unit":      true,
}

// splitNestedConstraintCall splits a nested constraint call into its callee and inner constraint text.
// The expected syntax is Name(field=value, field2=value2).
func splitNestedConstraintCall(value string) (callee, inner string, ok bool) {
	openParen := strings.Index(value, "(")
	if openParen <= 0 || !strings.HasSuffix(value, ")") {
		return "", "", false
	}

	return value[:openParen], value[openParen+1 : len(value)-1], true
}

// tryParseInlineInstantConstraints attempts to parse a constraint scope as inline Instant constraints.
// It walks the nested constraint's child symbol table scopes instead of re-parsing raw text.
// If the scope does not represent a nested Instant constraint, it falls back to string parsing.
func tryParseInlineInstantConstraints(constraintScope *symboltable.SymbolTable, concepts map[string]AnalysisConceptDefinition) (*InstantConstraints, error) {
	if constraintScope == nil {
		return nil, fmt.Errorf("constraint scope is nil")
	}

	// Verify this is a nested Instant constraint by looking for the callee identifier.
	var callee string
	for _, def := range constraintScope.Definitions {
		if def.Name != nil && *def.Name == "Instant" {
			callee = *def.Name
			break
		}
	}

	if callee == "" {
		// Not an Instant nested constraint - fall back to string parsing.
		var value string
		for _, def := range constraintScope.Definitions {
			if def.Data != nil {
				if strValue, ok := def.Data.(symboltable.StringData); ok {
					value = strValue.Value
					break
				}
			}
		}
		if value == "" {
			return nil, fmt.Errorf("not a nested constraint value")
		}
		return tryParseInlineInstantConstraintsFromString(value, concepts)
	}

	if callee != "Instant" {
		return nil, fmt.Errorf("starts_at inline constraint must use Instant, got %s", callee)
	}

	// Walk child constraint scopes to build constraints.
	if len(constraintScope.Children) == 0 {
		return &InstantConstraints{}, nil
	}

	constraints := make([]Constraint, 0)
	for _, child := range constraintScope.Children {
		if child.ScopeType != symboltable.ConstraintScope {
			continue
		}
		name, value, err := extractConstraint(child)
		if err != nil {
			return nil, fmt.Errorf("failed to extract nested constraint: %w", err)
		}
		constraints = append(constraints, Constraint{
			Name:  name,
			Value: value,
		})
	}

	return createInstantConstraints(constraints, concepts)
}

// tryParseInlineInstantConstraintsFromString parses inline Instant constraints from a raw string.
// This is kept as a fallback for tests and direct constraint construction.
func tryParseInlineInstantConstraintsFromString(value string, concepts map[string]AnalysisConceptDefinition) (*InstantConstraints, error) {
	callee, inner, ok := splitNestedConstraintCall(value)
	if !ok {
		return nil, fmt.Errorf("not a nested constraint value: %s", value)
	}
	if callee != "Instant" {
		return nil, fmt.Errorf("starts_at inline constraint must use Instant, got %s", callee)
	}

	if inner == "" {
		return &InstantConstraints{}, nil
	}

	// Parse "field=value, field2=value2" format
	constraints := make([]Constraint, 0)
	parts := strings.Split(inner, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid constraint format: %s", part)
		}
		constraints = append(constraints, Constraint{
			Name:  strings.TrimSpace(kv[0]),
			Value: strings.TrimSpace(kv[1]),
		})
	}

	return createInstantConstraints(constraints, concepts)
}

// createIntervalConstraints converts parser Interval constraints to IntervalConstraints.
// Handles both named type references and inline Instant constraints.
func createIntervalConstraints(parserConstraints []Constraint, concepts map[string]AnalysisConceptDefinition) (*IntervalConstraints, error) {
	ic := &IntervalConstraints{}

	for _, pc := range parserConstraints {
		if !validIntervalConstraintNames[pc.Name] {
			return nil, fmt.Errorf("unknown constraint for Interval: %s", pc.Name)
		}

		switch pc.Name {
		case "starts_at":
			if pc.Scope != nil && len(pc.Scope.Children) > 0 {
				// Parse as inline Instant constraints from symbol table scope.
				inlineConstraints, err := tryParseInlineInstantConstraints(pc.Scope, concepts)
				if err != nil {
					return nil, fmt.Errorf("invalid inline Instant constraints in starts_at: %w", err)
				}
				ic.StartsAt = inlineConstraints
			} else if _, _, ok := splitNestedConstraintCall(pc.Value); ok {
				// Fallback: parse as inline Instant constraints from string.
				inlineConstraints, err := tryParseInlineInstantConstraintsFromString(pc.Value, concepts)
				if err != nil {
					return nil, fmt.Errorf("invalid inline Instant constraints in starts_at: %w", err)
				}
				ic.StartsAt = inlineConstraints
			} else {
				// Treat as a named type reference.
				ic.StartTypeName = &pc.Value
			}
		case "unit":
			ic.UnitName = &pc.Value
		}
	}

	// Semantic validation: constraint consistency
	if err := ic.Validate(); err != nil {
		return nil, err
	}

	return ic, nil
}

// createStringConstraints converts parser String constraints to StringConstraints.
func createStringConstraints(parserConstraints []Constraint) (*StringConstraints, error) {
	sc := &StringConstraints{}

	for _, pc := range parserConstraints {
		switch pc.Name {
		case "minimum":
			v, err := strconv.Atoi(pc.Value)
			if err != nil {
				return nil, fmt.Errorf("minimum constraint must be an integer, got %s", pc.Value)
			}
			sc.Minimum = &v
		case "maximum":
			v, err := strconv.Atoi(pc.Value)
			if err != nil {
				return nil, fmt.Errorf("maximum constraint must be an integer, got %s", pc.Value)
			}
			sc.Maximum = &v
		case "pattern":
			// Strip surrounding matching quotes from STRING_LITERAL tokens
			raw := pc.Value
			if (strings.HasPrefix(raw, `"`) && strings.HasSuffix(raw, `"`)) ||
				(strings.HasPrefix(raw, `'`) && strings.HasSuffix(raw, `'`)) {
				raw = raw[1 : len(raw)-1]
			}
			if raw == "" {
				return nil, fmt.Errorf("pattern constraint must not be empty")
			}
			compiled, err := regexp.Compile(raw)
			if err != nil {
				return nil, fmt.Errorf("invalid regex pattern: %w", err)
			}
			sc.Pattern = compiled
		default:
			return nil, fmt.Errorf("unknown constraint for String: %s", pc.Name)
		}
	}

	if err := sc.Validate(); err != nil {
		return nil, err
	}

	return sc, nil
}

// createBooleanConstraints converts parser Boolean constraints to BooleanConstraints.
func createBooleanConstraints(parserConstraints []Constraint) (*BooleanConstraints, error) {
	if len(parserConstraints) > 0 {
		return nil, fmt.Errorf("Boolean does not accept constraints")
	}
	return &BooleanConstraints{}, nil
}
