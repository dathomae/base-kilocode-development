// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package analysis

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- NumberConstraints tests ---

func TestNumberConstraints_ValidateMoreRestrictive_Pass(t *testing.T) {
	t.Run("other is more restrictive - narrower range", func(t *testing.T) {
		inherited := &NumberConstraints{
			Minimum: intPtr(0),
			Maximum: intPtr(100),
		}
		derived := &NumberConstraints{
			Minimum: intPtr(10),
			Maximum: intPtr(50),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("other is equal", func(t *testing.T) {
		inherited := &NumberConstraints{
			Minimum: intPtr(0),
			Maximum: intPtr(100),
		}
		derived := &NumberConstraints{
			Minimum: intPtr(0),
			Maximum: intPtr(100),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("other adds new constraint", func(t *testing.T) {
		inherited := &NumberConstraints{
			Minimum: intPtr(0),
		}
		derived := &NumberConstraints{
			Minimum: intPtr(0),
			Maximum: intPtr(100),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("lower precision is more restrictive", func(t *testing.T) {
		inherited := &NumberConstraints{
			Precision: intPtr(2),
		}
		derived := &NumberConstraints{
			Precision: intPtr(0),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("nil receiver accepts anything", func(t *testing.T) {
		var inherited *NumberConstraints
		derived := &NumberConstraints{
			Minimum: intPtr(0),
			Maximum: intPtr(100),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})
}

func TestNumberConstraints_ValidateMoreRestrictive_Error(t *testing.T) {
	t.Run("other expands minimum", func(t *testing.T) {
		inherited := &NumberConstraints{
			Minimum: intPtr(0),
			Maximum: intPtr(100),
		}
		derived := &NumberConstraints{
			Minimum: intPtr(-10),
			Maximum: intPtr(50),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expands")
	})

	t.Run("other expands maximum", func(t *testing.T) {
		inherited := &NumberConstraints{
			Minimum: intPtr(0),
			Maximum: intPtr(100),
		}
		derived := &NumberConstraints{
			Minimum: intPtr(10),
			Maximum: intPtr(200),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expands")
	})

	t.Run("other increases precision", func(t *testing.T) {
		inherited := &NumberConstraints{
			Precision: intPtr(2),
		}
		derived := &NumberConstraints{
			Precision: intPtr(4),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expands")
	})

	t.Run("type mismatch", func(t *testing.T) {
		inherited := &NumberConstraints{}
		derived := &StringConstraints{}
		err := inherited.ValidateMoreRestrictive(derived)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "*NumberConstraints")
	})
}

func TestNumberConstraints_Merge(t *testing.T) {
	t.Run("new is more restrictive", func(t *testing.T) {
		current := &NumberConstraints{
			Minimum: intPtr(0),
			Maximum: intPtr(100),
		}
		derived := &NumberConstraints{
			Minimum: intPtr(10),
			Maximum: intPtr(50),
		}
		result := current.Merge(derived).(*NumberConstraints)
		assert.Equal(t, 10, *result.Minimum)
		assert.Equal(t, 50, *result.Maximum)
	})

	t.Run("overlapping ranges", func(t *testing.T) {
		current := &NumberConstraints{
			Minimum: intPtr(0),
			Maximum: intPtr(50),
		}
		derived := &NumberConstraints{
			Minimum: intPtr(25),
			Maximum: intPtr(100),
		}
		result := current.Merge(derived).(*NumberConstraints)
		assert.Equal(t, 25, *result.Minimum)
		assert.Equal(t, 100, *result.Maximum)
	})

	t.Run("new adds precision", func(t *testing.T) {
		current := &NumberConstraints{
			Minimum: intPtr(0),
			Maximum: intPtr(100),
		}
		derived := &NumberConstraints{
			Precision: intPtr(2),
		}
		result := current.Merge(derived).(*NumberConstraints)
		assert.Equal(t, 0, *result.Minimum)
		assert.Equal(t, 100, *result.Maximum)
		assert.Equal(t, 2, *result.Precision)
	})

	t.Run("nil current", func(t *testing.T) {
		var current *NumberConstraints
		derived := &NumberConstraints{
			Minimum: intPtr(10),
		}
		result := current.Merge(derived).(*NumberConstraints)
		assert.Equal(t, 10, *result.Minimum)
		assert.Nil(t, result.Maximum)
	})

	t.Run("nil new", func(t *testing.T) {
		current := &NumberConstraints{
			Minimum: intPtr(10),
		}
		var newC *NumberConstraints
		result := current.Merge(newC).(*NumberConstraints)
		assert.Equal(t, 10, *result.Minimum)
	})

	t.Run("type mismatch returns copy of current", func(t *testing.T) {
		current := &NumberConstraints{
			Minimum: intPtr(10),
		}
		other := &StringConstraints{}
		result := current.Merge(other).(*NumberConstraints)
		assert.Equal(t, 10, *result.Minimum)
	})
}

// --- StringConstraints tests ---

func TestStringConstraints_ValidateMoreRestrictive_Pass(t *testing.T) {
	t.Run("other is more restrictive - higher min, lower max", func(t *testing.T) {
		inherited := &StringConstraints{
			Minimum: intPtr(0),
			Maximum: intPtr(100),
		}
		derived := &StringConstraints{
			Minimum: intPtr(10),
			Maximum: intPtr(50),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("other is equal", func(t *testing.T) {
		inherited := &StringConstraints{
			Minimum: intPtr(10),
			Maximum: intPtr(50),
		}
		derived := &StringConstraints{
			Minimum: intPtr(10),
			Maximum: intPtr(50),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("same pattern is allowed", func(t *testing.T) {
		pattern := regexp.MustCompile("[a-z]+")
		inherited := &StringConstraints{
			Pattern: pattern,
		}
		derived := &StringConstraints{
			Pattern: pattern,
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("adding pattern is allowed", func(t *testing.T) {
		inherited := &StringConstraints{
			Minimum: intPtr(0),
		}
		derived := &StringConstraints{
			Minimum: intPtr(0),
			Pattern: regexp.MustCompile("[a-z]+"),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("nil receiver accepts anything", func(t *testing.T) {
		var inherited *StringConstraints
		derived := &StringConstraints{
			Minimum: intPtr(10),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})
}

func TestStringConstraints_ValidateMoreRestrictive_Error(t *testing.T) {
	t.Run("other lowers minimum", func(t *testing.T) {
		inherited := &StringConstraints{
			Minimum: intPtr(10),
		}
		derived := &StringConstraints{
			Minimum: intPtr(5),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expands")
	})

	t.Run("other raises maximum", func(t *testing.T) {
		inherited := &StringConstraints{
			Maximum: intPtr(50),
		}
		derived := &StringConstraints{
			Maximum: intPtr(100),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expands")
	})

	t.Run("other removes pattern - allowed by existing semantics", func(t *testing.T) {
		inherited := &StringConstraints{
			Pattern: regexp.MustCompile("[a-z]+"),
		}
		derived := &StringConstraints{}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("other changes pattern", func(t *testing.T) {
		inherited := &StringConstraints{
			Pattern: regexp.MustCompile("[a-z]+"),
		}
		derived := &StringConstraints{
			Pattern: regexp.MustCompile("[0-9]+"),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expands")
	})

	t.Run("type mismatch", func(t *testing.T) {
		inherited := &StringConstraints{}
		derived := &NumberConstraints{}
		err := inherited.ValidateMoreRestrictive(derived)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "*StringConstraints")
	})
}

func TestStringConstraints_Merge(t *testing.T) {
	t.Run("new is more restrictive", func(t *testing.T) {
		current := &StringConstraints{
			Minimum: intPtr(0),
			Maximum: intPtr(100),
		}
		derived := &StringConstraints{
			Minimum: intPtr(10),
			Maximum: intPtr(50),
		}
		result := current.Merge(derived).(*StringConstraints)
		assert.Equal(t, 10, *result.Minimum)
		assert.Equal(t, 50, *result.Maximum)
	})

	t.Run("overlapping ranges", func(t *testing.T) {
		current := &StringConstraints{
			Minimum: intPtr(0),
			Maximum: intPtr(50),
		}
		derived := &StringConstraints{
			Minimum: intPtr(25),
			Maximum: intPtr(100),
		}
		result := current.Merge(derived).(*StringConstraints)
		assert.Equal(t, 25, *result.Minimum)
		assert.Equal(t, 100, *result.Maximum)
	})

	t.Run("new adds pattern", func(t *testing.T) {
		current := &StringConstraints{
			Minimum: intPtr(0),
		}
		derived := &StringConstraints{
			Pattern: regexp.MustCompile("[a-z]+"),
		}
		result := current.Merge(derived).(*StringConstraints)
		assert.Equal(t, 0, *result.Minimum)
		assert.Equal(t, "[a-z]+", result.Pattern.String())
	})

	t.Run("nil current", func(t *testing.T) {
		var current *StringConstraints
		derived := &StringConstraints{
			Minimum: intPtr(10),
		}
		result := current.Merge(derived).(*StringConstraints)
		assert.Equal(t, 10, *result.Minimum)
	})

	t.Run("nil new", func(t *testing.T) {
		current := &StringConstraints{
			Minimum: intPtr(10),
		}
		var newC *StringConstraints
		result := current.Merge(newC).(*StringConstraints)
		assert.Equal(t, 10, *result.Minimum)
	})

	t.Run("type mismatch returns copy of current", func(t *testing.T) {
		current := &StringConstraints{
			Minimum: intPtr(10),
		}
		other := &NumberConstraints{}
		result := current.Merge(other).(*StringConstraints)
		assert.Equal(t, 10, *result.Minimum)
	})
}

// --- InstantConstraints tests ---

func TestInstantConstraints_ValidateMoreRestrictive_Pass(t *testing.T) {
	t.Run("other adds constraints", func(t *testing.T) {
		inherited := &InstantConstraints{
			Year: intPtr(2024),
		}
		derived := &InstantConstraints{
			Year:  intPtr(2024),
			Month: strPtr("January"),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("other is equal", func(t *testing.T) {
		inherited := &InstantConstraints{
			Year:  intPtr(2024),
			Month: strPtr("January"),
		}
		derived := &InstantConstraints{
			Year:  intPtr(2024),
			Month: strPtr("January"),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("nil receiver accepts anything", func(t *testing.T) {
		var inherited *InstantConstraints
		derived := &InstantConstraints{
			Year: intPtr(2024),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})
}

func TestInstantConstraints_ValidateMoreRestrictive_Error(t *testing.T) {
	t.Run("other changes value", func(t *testing.T) {
		inherited := &InstantConstraints{
			Year: intPtr(2024),
		}
		derived := &InstantConstraints{
			Year: intPtr(2025),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expands")
	})

	t.Run("other removes constraint - allowed by existing semantics", func(t *testing.T) {
		inherited := &InstantConstraints{
			Year: intPtr(2024),
		}
		derived := &InstantConstraints{}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("type mismatch", func(t *testing.T) {
		inherited := &InstantConstraints{}
		derived := &NumberConstraints{}
		err := inherited.ValidateMoreRestrictive(derived)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "*InstantConstraints")
	})
}

func TestInstantConstraints_Merge(t *testing.T) {
	t.Run("new adds fields", func(t *testing.T) {
		current := &InstantConstraints{
			Year: intPtr(2024),
		}
		derived := &InstantConstraints{
			Month: strPtr("January"),
		}
		result := current.Merge(derived).(*InstantConstraints)
		assert.Equal(t, 2024, *result.Year)
		assert.Equal(t, "January", *result.Month)
	})

	t.Run("new overrides with same value", func(t *testing.T) {
		current := &InstantConstraints{
			Year: intPtr(2024),
		}
		derived := &InstantConstraints{
			Year:  intPtr(2024),
			Month: strPtr("January"),
		}
		result := current.Merge(derived).(*InstantConstraints)
		assert.Equal(t, 2024, *result.Year)
		assert.Equal(t, "January", *result.Month)
	})

	t.Run("nil current", func(t *testing.T) {
		var current *InstantConstraints
		derived := &InstantConstraints{
			Year: intPtr(2024),
		}
		result := current.Merge(derived).(*InstantConstraints)
		assert.Equal(t, 2024, *result.Year)
	})

	t.Run("nil new", func(t *testing.T) {
		current := &InstantConstraints{
			Year: intPtr(2024),
		}
		var newC *InstantConstraints
		result := current.Merge(newC).(*InstantConstraints)
		assert.Equal(t, 2024, *result.Year)
	})

	t.Run("type mismatch returns copy of current", func(t *testing.T) {
		current := &InstantConstraints{
			Year: intPtr(2024),
		}
		other := &NumberConstraints{}
		result := current.Merge(other).(*InstantConstraints)
		assert.Equal(t, 2024, *result.Year)
	})
}

// --- IntervalConstraints tests ---

func TestIntervalConstraints_ValidateMoreRestrictive_Pass(t *testing.T) {
	t.Run("other adds constraints", func(t *testing.T) {
		inherited := &IntervalConstraints{
			StartsAt: &InstantConstraints{Year: intPtr(2024)},
		}
		derived := &IntervalConstraints{
			StartsAt: &InstantConstraints{Year: intPtr(2024)},
			UnitName: strPtr("month"),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("other is equal", func(t *testing.T) {
		inherited := &IntervalConstraints{
			StartsAt: &InstantConstraints{Year: intPtr(2024)},
			UnitName: strPtr("month"),
		}
		derived := &IntervalConstraints{
			StartsAt: &InstantConstraints{Year: intPtr(2024)},
			UnitName: strPtr("month"),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("nil receiver accepts anything", func(t *testing.T) {
		var inherited *IntervalConstraints
		derived := &IntervalConstraints{
			UnitName: strPtr("month"),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})
}

func TestIntervalConstraints_ValidateMoreRestrictive_Error(t *testing.T) {
	t.Run("other changes starts_at", func(t *testing.T) {
		inherited := &IntervalConstraints{
			StartsAt: &InstantConstraints{Year: intPtr(2024)},
		}
		derived := &IntervalConstraints{
			StartsAt: &InstantConstraints{Year: intPtr(2025)},
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expands")
	})

	t.Run("other changes unit", func(t *testing.T) {
		inherited := &IntervalConstraints{
			UnitName: strPtr("month"),
		}
		derived := &IntervalConstraints{
			UnitName: strPtr("day"),
		}
		err := inherited.ValidateMoreRestrictive(derived)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expands")
	})

	t.Run("other removes unit - allowed by existing semantics", func(t *testing.T) {
		inherited := &IntervalConstraints{
			UnitName: strPtr("month"),
		}
		derived := &IntervalConstraints{}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("type mismatch", func(t *testing.T) {
		inherited := &IntervalConstraints{}
		derived := &NumberConstraints{}
		err := inherited.ValidateMoreRestrictive(derived)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "*IntervalConstraints")
	})
}

func TestIntervalConstraints_Merge(t *testing.T) {
	t.Run("new adds unit", func(t *testing.T) {
		current := &IntervalConstraints{
			StartsAt: &InstantConstraints{Year: intPtr(2024)},
		}
		derived := &IntervalConstraints{
			UnitName: strPtr("month"),
		}
		result := current.Merge(derived).(*IntervalConstraints)
		assert.Equal(t, 2024, *result.StartsAt.Year)
		assert.Equal(t, "month", *result.UnitName)
	})

	t.Run("new overrides starts_at", func(t *testing.T) {
		current := &IntervalConstraints{
			StartsAt: &InstantConstraints{Year: intPtr(2024)},
		}
		derived := &IntervalConstraints{
			StartsAt: &InstantConstraints{Year: intPtr(2025)},
		}
		result := current.Merge(derived).(*IntervalConstraints)
		assert.Equal(t, 2025, *result.StartsAt.Year)
	})

	t.Run("nil current", func(t *testing.T) {
		var current *IntervalConstraints
		derived := &IntervalConstraints{
			UnitName: strPtr("month"),
		}
		result := current.Merge(derived).(*IntervalConstraints)
		assert.Equal(t, "month", *result.UnitName)
	})

	t.Run("nil new", func(t *testing.T) {
		current := &IntervalConstraints{
			UnitName: strPtr("month"),
		}
		var newC *IntervalConstraints
		result := current.Merge(newC).(*IntervalConstraints)
		assert.Equal(t, "month", *result.UnitName)
	})

	t.Run("type mismatch returns copy of current", func(t *testing.T) {
		current := &IntervalConstraints{
			UnitName: strPtr("month"),
		}
		other := &NumberConstraints{}
		result := current.Merge(other).(*IntervalConstraints)
		assert.Equal(t, "month", *result.UnitName)
	})
}

// --- BooleanConstraints tests ---

func TestBooleanConstraints_ValidateMoreRestrictive(t *testing.T) {
	t.Run("always passes", func(t *testing.T) {
		inherited := &BooleanConstraints{}
		derived := &BooleanConstraints{}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("nil receiver passes", func(t *testing.T) {
		var inherited *BooleanConstraints
		derived := &BooleanConstraints{}
		err := inherited.ValidateMoreRestrictive(derived)
		require.NoError(t, err)
	})

	t.Run("type mismatch", func(t *testing.T) {
		inherited := &BooleanConstraints{}
		derived := &NumberConstraints{}
		err := inherited.ValidateMoreRestrictive(derived)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "*BooleanConstraints")
	})

	t.Run("nil receiver with wrong type", func(t *testing.T) {
		var inherited *BooleanConstraints
		err := inherited.ValidateMoreRestrictive(&NumberConstraints{})
		require.NoError(t, err)
	})
}

func TestBooleanConstraints_Merge(t *testing.T) {
	t.Run("both non-nil returns empty", func(t *testing.T) {
		current := &BooleanConstraints{}
		derived := &BooleanConstraints{}
		result := current.Merge(derived).(*BooleanConstraints)
		assert.NotNil(t, result)
	})

	t.Run("nil current returns empty", func(t *testing.T) {
		var current *BooleanConstraints
		derived := &BooleanConstraints{}
		result := current.Merge(derived).(*BooleanConstraints)
		assert.NotNil(t, result)
	})

	t.Run("nil new returns empty", func(t *testing.T) {
		current := &BooleanConstraints{}
		var newC *BooleanConstraints
		result := current.Merge(newC).(*BooleanConstraints)
		assert.NotNil(t, result)
	})

	t.Run("both nil returns nil", func(t *testing.T) {
		var current *BooleanConstraints
		var newC *BooleanConstraints
		result := current.Merge(newC)
		assert.Nil(t, result)
	})

	t.Run("type mismatch returns empty", func(t *testing.T) {
		current := &BooleanConstraints{}
		other := &NumberConstraints{}
		result := current.Merge(other).(*BooleanConstraints)
		assert.NotNil(t, result)
	})
}

// --- Interface compile-time check ---

func TestConstraintValidator_InterfaceCompliance(t *testing.T) {
	// Verify that all constraint types implement ConstraintValidator
	var _ ConstraintValidator = (*NumberConstraints)(nil)
	var _ ConstraintValidator = (*StringConstraints)(nil)
	var _ ConstraintValidator = (*InstantConstraints)(nil)
	var _ ConstraintValidator = (*IntervalConstraints)(nil)
	var _ ConstraintValidator = (*BooleanConstraints)(nil)
}

// strPtr helper is needed here; intPtr is defined in cpl_types_test.go.
func strPtr(s string) *string {
	return &s
}
