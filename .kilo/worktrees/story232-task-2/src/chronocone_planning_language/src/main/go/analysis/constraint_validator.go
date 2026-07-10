// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file defines the ConstraintValidator interface and its implementations
// for all constraint types. It replaces the five near-identical constraint merge
// functions with an interface-driven approach while preserving exact narrowing semantics.
package analysis

import "fmt"

// ConstraintValidator defines the interface for validating and merging constraints.
//
// It replaces the five near-identical constraint merge/validate functions with a
// single interface-driven approach. Each constraint type (Number, String, Instant,
// Interval, Boolean) implements this interface.
type ConstraintValidator interface {
	// ValidateMoreRestrictive returns an error if other is NOT at least as
	// restrictive as the receiver. In other words, it validates that other
	// does not expand (become less restrictive than) the receiver.
	//
	// This preserves the exact semantics of the existing validateMoreRestrictive
	// function: a derived type's constraints must be equal to or narrower than
	// its base type's constraints.
	ValidateMoreRestrictive(other ConstraintValidator) error

	// Merge returns a new ConstraintValidator that combines the receiver and
	// other, taking the more restrictive value for each constraint field.
	//
	// This preserves the exact semantics of the existing merge functions:
	// for each constraint field, the value that narrows the allowed set most
	// is chosen. If one side does not specify a constraint, the other's value
	// is used.
	Merge(other ConstraintValidator) ConstraintValidator
}

// --- NumberConstraints implementation ---

// ValidateMoreRestrictive validates that other is at least as restrictive as
// the receiver. Returns an error if other expands any constraint.
func (n *NumberConstraints) ValidateMoreRestrictive(other ConstraintValidator) error {
	if n == nil {
		// A nil receiver has no constraints, so anything is acceptable.
		return nil
	}
	o, ok := other.(*NumberConstraints)
	if !ok {
		return fmt.Errorf("other must be *NumberConstraints, got %T", other)
	}
	return validateNumberMoreRestrictive(n, o)
}

// Merge combines two NumberConstraints, taking the more restrictive value for
// each field. The result is a new NumberConstraints; neither input is modified.
func (n *NumberConstraints) Merge(other ConstraintValidator) ConstraintValidator {
	if n == nil {
		if other == nil {
			return nil
		}
		o, ok := other.(*NumberConstraints)
		if !ok || o == nil {
			return nil
		}
		result := *o
		return &result
	}
	if other == nil {
		result := *n
		return &result
	}
	o, ok := other.(*NumberConstraints)
	if !ok || o == nil {
		result := *n
		return &result
	}
	merged, _ := mergeNumberConstraints(n, o)
	return merged
}

// --- InstantConstraints implementation ---

// ValidateMoreRestrictive validates that other is at least as restrictive as
// the receiver. Returns an error if other removes or changes any constraint.
func (i *InstantConstraints) ValidateMoreRestrictive(other ConstraintValidator) error {
	if i == nil {
		return nil
	}
	o, ok := other.(*InstantConstraints)
	if !ok {
		return fmt.Errorf("other must be *InstantConstraints, got %T", other)
	}
	return validateInstantMoreRestrictive(i, o)
}

// Merge combines two InstantConstraints, taking the union of all specified
// fields (new takes precedence when present). The result is a new
// InstantConstraints; neither input is modified.
func (i *InstantConstraints) Merge(other ConstraintValidator) ConstraintValidator {
	if i == nil {
		if other == nil {
			return nil
		}
		o, ok := other.(*InstantConstraints)
		if !ok || o == nil {
			return nil
		}
		result := *o
		return &result
	}
	if other == nil {
		result := *i
		return &result
	}
	o, ok := other.(*InstantConstraints)
	if !ok || o == nil {
		result := *i
		return &result
	}
	merged, _ := mergeInstantConstraints(i, o)
	return merged
}

// --- IntervalConstraints implementation ---

// ValidateMoreRestrictive validates that other is at least as restrictive as
// the receiver. Returns an error if other removes or changes any constraint.
func (i *IntervalConstraints) ValidateMoreRestrictive(other ConstraintValidator) error {
	if i == nil {
		return nil
	}
	o, ok := other.(*IntervalConstraints)
	if !ok {
		return fmt.Errorf("other must be *IntervalConstraints, got %T", other)
	}
	return validateIntervalMoreRestrictive(i, o)
}

// Merge combines two IntervalConstraints, taking the union of all specified
// fields (new takes precedence when present). The result is a new
// IntervalConstraints; neither input is modified.
func (i *IntervalConstraints) Merge(other ConstraintValidator) ConstraintValidator {
	if i == nil {
		if other == nil {
			return nil
		}
		o, ok := other.(*IntervalConstraints)
		if !ok || o == nil {
			return nil
		}
		result := *o
		return &result
	}
	if other == nil {
		result := *i
		return &result
	}
	o, ok := other.(*IntervalConstraints)
	if !ok || o == nil {
		result := *i
		return &result
	}
	merged, _ := mergeIntervalConstraints(i, o)
	return merged
}

// --- StringConstraints implementation ---

// ValidateMoreRestrictive validates that other is at least as restrictive as
// the receiver. Returns an error if other expands any constraint.
func (s *StringConstraints) ValidateMoreRestrictive(other ConstraintValidator) error {
	if s == nil {
		return nil
	}
	o, ok := other.(*StringConstraints)
	if !ok {
		return fmt.Errorf("other must be *StringConstraints, got %T", other)
	}
	return validateStringMoreRestrictive(s, o)
}

// Merge combines two StringConstraints, taking the more restrictive value for
// each field. The result is a new StringConstraints; neither input is modified.
func (s *StringConstraints) Merge(other ConstraintValidator) ConstraintValidator {
	if s == nil {
		if other == nil {
			return nil
		}
		o, ok := other.(*StringConstraints)
		if !ok || o == nil {
			return nil
		}
		result := *o
		return &result
	}
	if other == nil {
		result := *s
		return &result
	}
	o, ok := other.(*StringConstraints)
	if !ok || o == nil {
		result := *s
		return &result
	}
	merged, _ := mergeStringConstraints(s, o)
	return merged
}

// --- BooleanConstraints implementation ---

// ValidateMoreRestrictive always returns nil because BooleanConstraints has no
// configurable constraints.
func (b *BooleanConstraints) ValidateMoreRestrictive(other ConstraintValidator) error {
	if b == nil {
		return nil
	}
	if other == nil {
		return nil
	}
	_, ok := other.(*BooleanConstraints)
	if !ok {
		return fmt.Errorf("other must be *BooleanConstraints, got %T", other)
	}
	return nil
}

// Merge combines two BooleanConstraints. Since there are no fields, it returns
// a new empty BooleanConstraints.
func (b *BooleanConstraints) Merge(other ConstraintValidator) ConstraintValidator {
	if b == nil {
		if other == nil {
			return nil
		}
		o, ok := other.(*BooleanConstraints)
		if !ok || o == nil {
			return nil
		}
		result := *o
		return &result
	}
	if other == nil {
		result := *b
		return &result
	}
	o, ok := other.(*BooleanConstraints)
	if !ok {
		result := *b
		return &result
	}
	merged, _ := mergeBooleanConstraints(b, o)
	return merged
}
