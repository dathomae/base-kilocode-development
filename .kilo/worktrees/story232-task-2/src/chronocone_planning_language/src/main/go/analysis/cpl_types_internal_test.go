// Package analysis contains internal tests for CPL analysis types.
// These tests access package-private variables and functions.
package analysis

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidInstantConstraintNames verifies the complete set of valid Instant
// constraint names.
func TestValidInstantConstraintNames(t *testing.T) {
	t.Parallel()

	// All 10 valid constraint names must be recognized
	validNames := []string{
		"year",
		"month",
		"day_of_month",
		"day_of_year",
		"week_number",
		"fiscal_quarter",
		"day_of_week",
		"hour",
		"minute",
		"second",
	}

	for _, name := range validNames {
		name := name
		t.Run("valid: "+name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, validInstantConstraintNames[name], "%s should be a valid constraint name", name)
		})
	}

	// Unknown names must not be recognized
	t.Run("unknown name is rejected", func(t *testing.T) {
		t.Parallel()
		assert.False(t, validInstantConstraintNames["invalid_field"], "invalid_field should not be valid")
	})
}

// buildTestConcepts loads standard time concepts from the CPL file for internal testing.
// This ensures tests use the same 0-based ranges defined in standard_concepts.cpl
// (Hour=0..23, Minute=0..59, Second=0..59) rather than hardcoded values.
func buildTestConcepts() map[string]AnalysisConceptDefinition {
	concepts, err := loadStandardTimeConceptsFromFile()
	if err != nil {
		panic(fmt.Sprintf("failed to load standard time concepts: %v", err))
	}
	return concepts
}

// TestCreateInstantConstraintsNewNames verifies that the new constraint names
// are correctly parsed and stored in InstantConstraints.
func TestCreateInstantConstraintsNewNames(t *testing.T) {
	t.Parallel()

	concepts := buildTestConcepts()

	t.Run("day_of_month constraint", func(t *testing.T) {
		t.Parallel()
		result, err := createInstantConstraints([]Constraint{{Name: "day_of_month", Value: "15"}}, concepts)
		requireNoError(t, err)
		assert.NotNil(t, result.DayOfMonth)
		assert.Equal(t, 15, *result.DayOfMonth)
	})

	t.Run("day_of_year constraint", func(t *testing.T) {
		t.Parallel()
		result, err := createInstantConstraints([]Constraint{{Name: "day_of_year", Value: "42"}}, concepts)
		requireNoError(t, err)
		assert.NotNil(t, result.DayOfYear)
		assert.Equal(t, 42, *result.DayOfYear)
	})

	t.Run("fiscal_quarter constraint", func(t *testing.T) {
		t.Parallel()
		result, err := createInstantConstraints([]Constraint{{Name: "fiscal_quarter", Value: "2"}}, concepts)
		requireNoError(t, err)
		assert.NotNil(t, result.FiscalQuarter)
		assert.Equal(t, 2, *result.FiscalQuarter)
	})

}

// TestCreateInstantConstraintsMissingConcept verifies errors when required concepts are missing.
func TestCreateInstantConstraintsMissingConcept(t *testing.T) {
	t.Parallel()

	t.Run("missing Month concept", func(t *testing.T) {
		t.Parallel()
		emptyConcepts := map[string]AnalysisConceptDefinition{}
		_, err := createInstantConstraints([]Constraint{{Name: "month", Value: "January"}}, emptyConcepts)
		requireError(t, err)
		assert.Contains(t, err.Error(), "required concept Month for constraint month is not defined")
	})

	t.Run("missing Hour concept", func(t *testing.T) {
		t.Parallel()
		emptyConcepts := map[string]AnalysisConceptDefinition{}
		_, err := createInstantConstraints([]Constraint{{Name: "hour", Value: "12"}}, emptyConcepts)
		requireError(t, err)
		assert.Contains(t, err.Error(), "required concept Hour for constraint hour is not defined")
	})
}

// TestCreateInstantConstraintsOutOfBoundsInteger verifies errors for out-of-range integer constraints.
func TestCreateInstantConstraintsOutOfBoundsInteger(t *testing.T) {
	t.Parallel()

	concepts := buildTestConcepts()

	t.Run("hour too high", func(t *testing.T) {
		t.Parallel()
		_, err := createInstantConstraints([]Constraint{{Name: "hour", Value: "24"}}, concepts)
		requireError(t, err)
		assert.Contains(t, err.Error(), "hour must be between 0 and 23, got 24")
	})

	t.Run("hour negative", func(t *testing.T) {
		t.Parallel()
		_, err := createInstantConstraints([]Constraint{{Name: "hour", Value: "-1"}}, concepts)
		requireError(t, err)
		assert.Contains(t, err.Error(), "hour must be between 0 and 23, got -1")
	})

	t.Run("minute too high", func(t *testing.T) {
		t.Parallel()
		_, err := createInstantConstraints([]Constraint{{Name: "minute", Value: "60"}}, concepts)
		requireError(t, err)
		assert.Contains(t, err.Error(), "minute must be between 0 and 59, got 60")
	})

	t.Run("second too high", func(t *testing.T) {
		t.Parallel()
		_, err := createInstantConstraints([]Constraint{{Name: "second", Value: "60"}}, concepts)
		requireError(t, err)
		assert.Contains(t, err.Error(), "second must be between 0 and 59, got 60")
	})

	t.Run("day_of_month too high", func(t *testing.T) {
		t.Parallel()
		_, err := createInstantConstraints([]Constraint{{Name: "day_of_month", Value: "32"}}, concepts)
		requireError(t, err)
		assert.Contains(t, err.Error(), "day_of_month must be between 1 and 31, got 32")
	})

	t.Run("week_number too high", func(t *testing.T) {
		t.Parallel()
		_, err := createInstantConstraints([]Constraint{{Name: "week_number", Value: "54"}}, concepts)
		requireError(t, err)
		assert.Contains(t, err.Error(), "week_number must be between 1 and 53, got 54")
	})

	t.Run("fiscal_quarter too high", func(t *testing.T) {
		t.Parallel()
		_, err := createInstantConstraints([]Constraint{{Name: "fiscal_quarter", Value: "5"}}, concepts)
		requireError(t, err)
		assert.Contains(t, err.Error(), "fiscal_quarter must be between 1 and 4, got 5")
	})

	t.Run("year below range", func(t *testing.T) {
		t.Parallel()
		_, err := createInstantConstraints([]Constraint{{Name: "year", Value: "1899"}}, concepts)
		requireError(t, err)
		assert.Contains(t, err.Error(), "year must be between 1900 and 3000, got 1899")
	})
}

// TestCreateInstantConstraintsInvalidStringValue verifies errors for invalid string concept values.
func TestCreateInstantConstraintsInvalidStringValue(t *testing.T) {
	t.Parallel()

	concepts := buildTestConcepts()

	t.Run("invalid month", func(t *testing.T) {
		t.Parallel()
		_, err := createInstantConstraints([]Constraint{{Name: "month", Value: "NotAMonth"}}, concepts)
		requireError(t, err)
		assert.Contains(t, err.Error(), "invalid month: NotAMonth")
	})

	t.Run("invalid day_of_week", func(t *testing.T) {
		t.Parallel()
		_, err := createInstantConstraints([]Constraint{{Name: "day_of_week", Value: "NotADay"}}, concepts)
		requireError(t, err)
		assert.Contains(t, err.Error(), "invalid day_of_week: NotADay")
	})
}

// TestCreateInstantConstraintsValidStringValue verifies valid string concept values are accepted.
func TestCreateInstantConstraintsValidStringValue(t *testing.T) {
	t.Parallel()

	concepts := buildTestConcepts()

	t.Run("valid month", func(t *testing.T) {
		t.Parallel()
		result, err := createInstantConstraints([]Constraint{{Name: "month", Value: "December"}}, concepts)
		requireNoError(t, err)
		assert.NotNil(t, result.Month)
		assert.Equal(t, "December", *result.Month)
	})

	t.Run("valid day_of_week", func(t *testing.T) {
		t.Parallel()
		result, err := createInstantConstraints([]Constraint{{Name: "day_of_week", Value: "Friday"}}, concepts)
		requireNoError(t, err)
		assert.NotNil(t, result.DayOfWeek)
		assert.Equal(t, "Friday", *result.DayOfWeek)
	})
}

// TestCreateInstantConstraintsValidIntegerValue verifies valid integer concept values are accepted.
func TestCreateInstantConstraintsValidIntegerValue(t *testing.T) {
	t.Parallel()

	concepts := buildTestConcepts()

	t.Run("valid hour lower boundary (0-based)", func(t *testing.T) {
		t.Parallel()
		result, err := createInstantConstraints([]Constraint{{Name: "hour", Value: "0"}}, concepts)
		requireNoError(t, err)
		assert.NotNil(t, result.Hour)
		assert.Equal(t, 0, *result.Hour)
	})

	t.Run("valid hour upper boundary (0-based)", func(t *testing.T) {
		t.Parallel()
		result, err := createInstantConstraints([]Constraint{{Name: "hour", Value: "23"}}, concepts)
		requireNoError(t, err)
		assert.NotNil(t, result.Hour)
		assert.Equal(t, 23, *result.Hour)
	})

	t.Run("valid minute lower boundary (0-based)", func(t *testing.T) {
		t.Parallel()
		result, err := createInstantConstraints([]Constraint{{Name: "minute", Value: "0"}}, concepts)
		requireNoError(t, err)
		assert.NotNil(t, result.Minute)
		assert.Equal(t, 0, *result.Minute)
	})

	t.Run("valid minute upper boundary (0-based)", func(t *testing.T) {
		t.Parallel()
		result, err := createInstantConstraints([]Constraint{{Name: "minute", Value: "59"}}, concepts)
		requireNoError(t, err)
		assert.NotNil(t, result.Minute)
		assert.Equal(t, 59, *result.Minute)
	})

	t.Run("valid second lower boundary (0-based)", func(t *testing.T) {
		t.Parallel()
		result, err := createInstantConstraints([]Constraint{{Name: "second", Value: "0"}}, concepts)
		requireNoError(t, err)
		assert.NotNil(t, result.Second)
		assert.Equal(t, 0, *result.Second)
	})

	t.Run("valid second upper boundary (0-based)", func(t *testing.T) {
		t.Parallel()
		result, err := createInstantConstraints([]Constraint{{Name: "second", Value: "59"}}, concepts)
		requireNoError(t, err)
		assert.NotNil(t, result.Second)
		assert.Equal(t, 59, *result.Second)
	})
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func requireError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

// intPtr returns a pointer to the given int value.
func intPtr(v int) *int {
	return &v
}

// TestCreateTypeReferenceString tests createTypeReference with "String".
func TestCreateTypeReferenceString(t *testing.T) {
	t.Parallel()

	ref, err := createTypeReference("String")
	requireNoError(t, err)
	require.NotNil(t, ref)

	builtInRef, ok := ref.(*BuiltInTypeRef)
	require.True(t, ok, "expected *BuiltInTypeRef")

	assert.True(t, builtInRef.IsBuiltIn(), "IsBuiltIn() should be true")
	assert.Equal(t, "String", builtInRef.GetName(), "GetName() should return String")
	assert.Equal(t, StringType, builtInRef.GetBuiltInType(), "GetBuiltInType() should return StringType")
}

// TestCreateTypeReferenceBoolean tests createTypeReference with "Boolean".
func TestCreateTypeReferenceBoolean(t *testing.T) {
	t.Parallel()

	ref, err := createTypeReference("Boolean")
	requireNoError(t, err)
	require.NotNil(t, ref)

	builtInRef, ok := ref.(*BuiltInTypeRef)
	require.True(t, ok, "expected *BuiltInTypeRef")

	assert.True(t, builtInRef.IsBuiltIn(), "IsBuiltIn() should be true")
	assert.Equal(t, "Boolean", builtInRef.GetName(), "GetName() should return Boolean")
	assert.Equal(t, BooleanType, builtInRef.GetBuiltInType(), "GetBuiltInType() should return BooleanType")
}

// TestParseBuiltInTypeNameString tests parseBuiltInTypeName with "String".
func TestParseBuiltInTypeNameString(t *testing.T) {
	t.Parallel()

	bt, err := parseBuiltInTypeName("String")
	requireNoError(t, err)
	assert.Equal(t, StringType, bt, "parseBuiltInTypeName should return StringType")
}

// TestParseBuiltInTypeNameBoolean tests parseBuiltInTypeName with "Boolean".
func TestParseBuiltInTypeNameBoolean(t *testing.T) {
	t.Parallel()

	bt, err := parseBuiltInTypeName("Boolean")
	requireNoError(t, err)
	assert.Equal(t, BooleanType, bt, "parseBuiltInTypeName should return BooleanType")
}

// TestParseBuiltInTypeNameUnknown tests parseBuiltInTypeName with an unknown type name.
func TestParseBuiltInTypeNameUnknown(t *testing.T) {
	t.Parallel()

	bt, err := parseBuiltInTypeName("UnknownType")
	requireError(t, err)
	assert.Contains(t, err.Error(), "unknown built-in type: UnknownType")
	assert.Equal(t, NumberType, bt, "default return value should be NumberType")
}

// TestValidateStringMoreRestrictive tests String constraint narrowing validation.
func TestValidateStringMoreRestrictive(t *testing.T) {
	t.Parallel()

	patternA := regexp.MustCompile("^[a-z]+$")
	patternB := regexp.MustCompile("^[0-9]+$")

	testCases := []struct {
		name        string
		inherited   *StringConstraints
		new         *StringConstraints
		expectError bool
	}{
		{
			name:        "new minimum=10 when inherited minimum=5 → no error",
			inherited:   &StringConstraints{Minimum: intPtr(5)},
			new:         &StringConstraints{Minimum: intPtr(10)},
			expectError: false,
		},
		{
			name:        "new minimum=5 when inherited minimum=10 → error",
			inherited:   &StringConstraints{Minimum: intPtr(10)},
			new:         &StringConstraints{Minimum: intPtr(5)},
			expectError: true,
		},
		{
			name:        "new minimum=nil when inherited minimum=5 → no error",
			inherited:   &StringConstraints{Minimum: intPtr(5)},
			new:         &StringConstraints{},
			expectError: false,
		},
		{
			name:        "new maximum=50 when inherited maximum=100 → no error",
			inherited:   &StringConstraints{Maximum: intPtr(100)},
			new:         &StringConstraints{Maximum: intPtr(50)},
			expectError: false,
		},
		{
			name:        "new maximum=100 when inherited maximum=50 → error",
			inherited:   &StringConstraints{Maximum: intPtr(50)},
			new:         &StringConstraints{Maximum: intPtr(100)},
			expectError: true,
		},
		{
			name:        "new pattern=^[a-z]+$ when inherited pattern=nil → no error",
			inherited:   &StringConstraints{},
			new:         &StringConstraints{Pattern: patternA},
			expectError: false,
		},
		{
			name:        "new pattern=nil when inherited pattern=^[a-z]+$ → no error (inherited)",
			inherited:   &StringConstraints{Pattern: patternA},
			new:         &StringConstraints{},
			expectError: false,
		},
		{
			name:        "same pattern → no error",
			inherited:   &StringConstraints{Pattern: patternA},
			new:         &StringConstraints{Pattern: patternA},
			expectError: false,
		},
		{
			name:        "different pattern → error",
			inherited:   &StringConstraints{Pattern: patternA},
			new:         &StringConstraints{Pattern: patternB},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateStringMoreRestrictive(tc.inherited, tc.new)
			if tc.expectError {
				requireError(t, err)
			} else {
				requireNoError(t, err)
			}
		})
	}
}

// TestValidateBooleanMoreRestrictive tests Boolean constraint narrowing validation.
func TestValidateBooleanMoreRestrictive(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		inherited   *BooleanConstraints
		new         *BooleanConstraints
		expectError bool
	}{
		{
			name:        "empty inherited and new → no error",
			inherited:   &BooleanConstraints{},
			new:         &BooleanConstraints{},
			expectError: false,
		},
		{
			name:        "nil inherited and new → no error",
			inherited:   nil,
			new:         nil,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateBooleanMoreRestrictive(tc.inherited, tc.new)
			if tc.expectError {
				requireError(t, err)
			} else {
				requireNoError(t, err)
			}
		})
	}
}

// TestMergeStringConstraints tests merging two StringConstraints.
func TestMergeStringConstraints(t *testing.T) {
	t.Parallel()

	patternA := regexp.MustCompile("^[a-z]+$")

	testCases := []struct {
		name         string
		current      *StringConstraints
		new          *StringConstraints
		expectError  bool
		expectResult *StringConstraints
	}{
		{
			name:         "merging with current=nil returns new",
			current:      nil,
			new:          &StringConstraints{Minimum: intPtr(10)},
			expectError:  false,
			expectResult: &StringConstraints{Minimum: intPtr(10)},
		},
		{
			name:         "new minimum=10, current minimum=nil → result minimum=10",
			current:      &StringConstraints{},
			new:          &StringConstraints{Minimum: intPtr(10)},
			expectError:  false,
			expectResult: &StringConstraints{Minimum: intPtr(10)},
		},
		{
			name:         "new maximum=nil, current maximum=50 → result maximum=50",
			current:      &StringConstraints{Maximum: intPtr(50)},
			new:          &StringConstraints{},
			expectError:  false,
			expectResult: &StringConstraints{Maximum: intPtr(50)},
		},
		{
			name:         "new pattern set, current pattern nil → result pattern=new",
			current:      &StringConstraints{},
			new:          &StringConstraints{Pattern: patternA},
			expectError:  false,
			expectResult: &StringConstraints{Pattern: patternA},
		},
		{
			name:         "merging with new=nil returns current",
			current:      &StringConstraints{Minimum: intPtr(5)},
			new:          nil,
			expectError:  false,
			expectResult: &StringConstraints{Minimum: intPtr(5)},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := mergeStringConstraints(tc.current, tc.new)
			if tc.expectError {
				requireError(t, err)
			} else {
				requireNoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.expectResult.Minimum, result.Minimum, "Minimum should match")
				assert.Equal(t, tc.expectResult.Maximum, result.Maximum, "Maximum should match")
				if tc.expectResult.Pattern != nil {
					require.NotNil(t, result.Pattern)
					assert.Equal(t, tc.expectResult.Pattern.String(), result.Pattern.String(), "Pattern should match")
				} else {
					assert.Nil(t, result.Pattern, "Pattern should be nil")
				}
			}
		})
	}
}

// TestMergeBooleanConstraints tests merging two BooleanConstraints.
func TestMergeBooleanConstraints(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		current      *BooleanConstraints
		new          *BooleanConstraints
		expectError  bool
		expectResult *BooleanConstraints
	}{
		{
			name:         "merging with current=nil returns new",
			current:      nil,
			new:          &BooleanConstraints{},
			expectError:  false,
			expectResult: &BooleanConstraints{},
		},
		{
			name:         "merging with new=nil returns current",
			current:      &BooleanConstraints{},
			new:          nil,
			expectError:  false,
			expectResult: &BooleanConstraints{},
		},
		{
			name:         "merging two BooleanConstraints returns BooleanConstraints",
			current:      &BooleanConstraints{},
			new:          &BooleanConstraints{},
			expectError:  false,
			expectResult: &BooleanConstraints{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := mergeBooleanConstraints(tc.current, tc.new)
			if tc.expectError {
				requireError(t, err)
			} else {
				requireNoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}
