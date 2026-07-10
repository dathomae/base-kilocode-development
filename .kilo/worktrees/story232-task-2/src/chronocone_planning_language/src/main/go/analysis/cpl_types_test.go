// Package analysis_test contains tests for the CPL analysis types.
package analysis_test

import (
	"regexp"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// intPtr returns a pointer to the given int value.
func intPtr(v int) *int {
	return &v
}

// TestBuiltInType tests the BuiltInType enum and its String method.
func TestBuiltInType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		bt       analysis.BuiltInType
		expected string
	}{
		{
			name:     "NumberType",
			bt:       analysis.NumberType,
			expected: "Number",
		},
		{
			name:     "StringType",
			bt:       analysis.StringType,
			expected: "String",
		},
		{
			name:     "BooleanType",
			bt:       analysis.BooleanType,
			expected: "Boolean",
		},
		{
			name:     "Unknown type (iota after NumberType)",
			bt:       analysis.BuiltInType(999),
			expected: "Unknown",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.bt.String(), "String representation should match")
		})
	}
}

// TestConstraint tests creating Constraint structs.
func TestConstraint(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		constraint    analysis.Constraint
		expectedName  string
		expectedValue string
	}{
		{
			name:          "precision constraint",
			constraint:    analysis.Constraint{Name: "precision", Value: "2"},
			expectedName:  "precision",
			expectedValue: "2",
		},
		{
			name:          "min constraint",
			constraint:    analysis.Constraint{Name: "min", Value: "0"},
			expectedName:  "min",
			expectedValue: "0",
		},
		{
			name:          "max constraint",
			constraint:    analysis.Constraint{Name: "max", Value: "100"},
			expectedName:  "max",
			expectedValue: "100",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expectedName, tc.constraint.Name, "Constraint name should match")
			assert.Equal(t, tc.expectedValue, tc.constraint.Value, "Constraint value should match")
		})
	}
}

// TestNumberConstraints tests creating NumberConstraints structs.
func TestNumberConstraints(t *testing.T) {
	t.Parallel()

	precision := 2
	min := 0
	max := 100

	testCases := []struct {
		name              string
		constraints       analysis.NumberConstraints
		expectPrecision   bool
		expectMin         bool
		expectMax         bool
		expectedPrecision int
		expectedMin       int
		expectedMax       int
	}{
		{
			name: "all constraints specified",
			constraints: analysis.NumberConstraints{
				Precision: &precision,
				Minimum:   &min,
				Maximum:   &max,
			},
			expectPrecision:   true,
			expectMin:         true,
			expectMax:         true,
			expectedPrecision: 2,
			expectedMin:       0,
			expectedMax:       100,
		},
		{
			name: "only precision",
			constraints: analysis.NumberConstraints{
				Precision: &precision,
			},
			expectPrecision:   true,
			expectMin:         false,
			expectMax:         false,
			expectedPrecision: 2,
		},
		{
			name: "min and max only",
			constraints: analysis.NumberConstraints{
				Minimum: &min,
				Maximum: &max,
			},
			expectPrecision: false,
			expectMin:       true,
			expectMax:       true,
			expectedMin:     0,
			expectedMax:     100,
		},
		{
			name: "no constraints",
			constraints: analysis.NumberConstraints{
				Precision: nil,
				Minimum:   nil,
				Maximum:   nil,
			},
			expectPrecision: false,
			expectMin:       false,
			expectMax:       false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.expectPrecision {
				assert.NotNil(t, tc.constraints.Precision, "Precision should be set")
				assert.Equal(t, tc.expectedPrecision, *tc.constraints.Precision, "Precision value should match")
			} else {
				assert.Nil(t, tc.constraints.Precision, "Precision should be nil")
			}

			if tc.expectMin {
				assert.NotNil(t, tc.constraints.Minimum, "Minimum should be set")
				assert.Equal(t, tc.expectedMin, *tc.constraints.Minimum, "Minimum value should match")
			} else {
				assert.Nil(t, tc.constraints.Minimum, "Minimum should be nil")
			}

			if tc.expectMax {
				assert.NotNil(t, tc.constraints.Maximum, "Maximum should be set")
				assert.Equal(t, tc.expectedMax, *tc.constraints.Maximum, "Maximum value should match")
			} else {
				assert.Nil(t, tc.constraints.Maximum, "Maximum should be nil")
			}
		})
	}
}

// TestTypeDefinition tests creating TypeDefinition structs.
func TestTypeDefinition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		typeDef          analysis.TypeDefinition
		expectedName     string
		expectedBuiltIn  analysis.BuiltInType
		expectedConstLen int
	}{
		{
			name: "simple Number type",
			typeDef: analysis.TypeDefinition{
				Name:        "Currency",
				BuiltInType: analysis.NumberType,
				Constraints: []analysis.Constraint{},
			},
			expectedName:     "Currency",
			expectedBuiltIn:  analysis.NumberType,
			expectedConstLen: 0,
		},
		{
			name: "Number type with constraints",
			typeDef: analysis.TypeDefinition{
				Name:        "Percentage",
				BuiltInType: analysis.NumberType,
				Constraints: []analysis.Constraint{
					{Name: "min", Value: "0"},
					{Name: "max", Value: "100"},
				},
			},
			expectedName:     "Percentage",
			expectedBuiltIn:  analysis.NumberType,
			expectedConstLen: 2,
		},
		{
			name: "Currency with precision",
			typeDef: analysis.TypeDefinition{
				Name:        "Currency",
				BuiltInType: analysis.NumberType,
				Constraints: []analysis.Constraint{
					{Name: "precision", Value: "2"},
				},
			},
			expectedName:     "Currency",
			expectedBuiltIn:  analysis.NumberType,
			expectedConstLen: 1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expectedName, tc.typeDef.Name, "Type name should match")
			assert.Equal(t, tc.expectedBuiltIn, tc.typeDef.BuiltInType, "Built-in type should match")
			assert.Equal(t, tc.expectedBuiltIn.String(), "Number", "Built-in type string should be Number")
			assert.Len(t, tc.typeDef.Constraints, tc.expectedConstLen, "Constraint count should match")
		})
	}
}

// TestStringConstraintsValidate tests StringConstraints.Validate.
func TestStringConstraintsValidate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		constraints analysis.StringConstraints
		expectError bool
	}{
		{
			name:        "valid: min=0, max=100",
			constraints: analysis.StringConstraints{Minimum: intPtr(0), Maximum: intPtr(100)},
			expectError: false,
		},
		{
			name:        "valid: only min=5",
			constraints: analysis.StringConstraints{Minimum: intPtr(5)},
			expectError: false,
		},
		{
			name:        "valid: only max=50",
			constraints: analysis.StringConstraints{Maximum: intPtr(50)},
			expectError: false,
		},
		{
			name:        "valid: only pattern",
			constraints: analysis.StringConstraints{Pattern: regexp.MustCompile("^[a-z]+$")},
			expectError: false,
		},
		{
			name:        "invalid: min=-1",
			constraints: analysis.StringConstraints{Minimum: intPtr(-1)},
			expectError: true,
		},
		{
			name:        "invalid: max=-1",
			constraints: analysis.StringConstraints{Maximum: intPtr(-1)},
			expectError: true,
		},
		{
			name:        "invalid: min=5, max=3 (min > max)",
			constraints: analysis.StringConstraints{Minimum: intPtr(5), Maximum: intPtr(3)},
			expectError: true,
		},
		{
			name:        "valid: min=5, max=5 (min == max)",
			constraints: analysis.StringConstraints{Minimum: intPtr(5), Maximum: intPtr(5)},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.constraints.Validate()
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestStringConstraintsConstraintNames tests StringConstraints.ConstraintNames.
func TestStringConstraintsConstraintNames(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		constraints   analysis.StringConstraints
		expectedNames []string
	}{
		{
			name:          "all fields set",
			constraints:   analysis.StringConstraints{Minimum: intPtr(1), Maximum: intPtr(100), Pattern: regexp.MustCompile("^[a-z]+$")},
			expectedNames: []string{"minimum", "maximum", "pattern"},
		},
		{
			name:          "only minimum set",
			constraints:   analysis.StringConstraints{Minimum: intPtr(5)},
			expectedNames: []string{"minimum"},
		},
		{
			name:          "only maximum set",
			constraints:   analysis.StringConstraints{Maximum: intPtr(50)},
			expectedNames: []string{"maximum"},
		},
		{
			name:          "only pattern set",
			constraints:   analysis.StringConstraints{Pattern: regexp.MustCompile(".*")},
			expectedNames: []string{"pattern"},
		},
		{
			name:          "none set",
			constraints:   analysis.StringConstraints{},
			expectedNames: []string{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			names := tc.constraints.ConstraintNames()
			assert.Equal(t, tc.expectedNames, names)
		})
	}
}

// TestStringConstraintsGetConstraint tests StringConstraints.GetConstraint.
func TestStringConstraintsGetConstraint(t *testing.T) {
	t.Parallel()

	minVal := 5
	maxVal := 50
	pattern := regexp.MustCompile("^[a-z]+$")

	testCases := []struct {
		name           string
		constraints    analysis.StringConstraints
		constraintName string
		expectFound    bool
		expectType     string
	}{
		{
			name:           "minimum returns *int when set",
			constraints:    analysis.StringConstraints{Minimum: &minVal},
			constraintName: "minimum",
			expectFound:    true,
			expectType:     "*int",
		},
		{
			name:           "maximum returns *int when set",
			constraints:    analysis.StringConstraints{Maximum: &maxVal},
			constraintName: "maximum",
			expectFound:    true,
			expectType:     "*int",
		},
		{
			name:           "pattern returns *regexp.Regexp when set",
			constraints:    analysis.StringConstraints{Pattern: pattern},
			constraintName: "pattern",
			expectFound:    true,
			expectType:     "*regexp.Regexp",
		},
		{
			name:           "unknown name returns nil, false",
			constraints:    analysis.StringConstraints{Minimum: &minVal},
			constraintName: "unknown",
			expectFound:    false,
			expectType:     "",
		},
		{
			name:           "minimum not set returns nil, false",
			constraints:    analysis.StringConstraints{Maximum: &maxVal},
			constraintName: "minimum",
			expectFound:    false,
			expectType:     "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			val, found := tc.constraints.GetConstraint(tc.constraintName)
			assert.Equal(t, tc.expectFound, found)
			if tc.expectFound {
				assert.NotNil(t, val)
				switch tc.expectType {
				case "*int":
					intVal, ok := val.(*int)
					assert.True(t, ok, "expected *int, got %T", val)
					if ok {
						if tc.constraintName == "minimum" {
							assert.Equal(t, minVal, *intVal)
						} else if tc.constraintName == "maximum" {
							assert.Equal(t, maxVal, *intVal)
						}
					}
				case "*regexp.Regexp":
					regexVal, ok := val.(*regexp.Regexp)
					assert.True(t, ok, "expected *regexp.Regexp, got %T", val)
					assert.Equal(t, pattern, regexVal)
				}
			} else {
				assert.Nil(t, val)
			}
		})
	}
}

// TestStringConstraintsIsMoreOrEqualRestrictive tests StringConstraints.IsMoreOrEqualRestrictive.
func TestStringConstraintsIsMoreOrEqualRestrictive(t *testing.T) {
	t.Parallel()

	patternA := regexp.MustCompile("^[a-z]+$")
	patternB := regexp.MustCompile("^[0-9]+$")

	testCases := []struct {
		name           string
		constraintName string
		currentValue   interface{}
		newValue       interface{}
		expectedResult bool
		expectError    bool
	}{
		{
			name:           "minimum: current=5, new=10 → true",
			constraintName: "minimum",
			currentValue:   intPtr(5),
			newValue:       intPtr(10),
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "minimum: current=10, new=5 → false",
			constraintName: "minimum",
			currentValue:   intPtr(10),
			newValue:       intPtr(5),
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "minimum: current=5, new=5 → true",
			constraintName: "minimum",
			currentValue:   intPtr(5),
			newValue:       intPtr(5),
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "minimum: current=nil, new=5 → true",
			constraintName: "minimum",
			currentValue:   (*int)(nil),
			newValue:       intPtr(5),
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "maximum: current=100, new=50 → true",
			constraintName: "maximum",
			currentValue:   intPtr(100),
			newValue:       intPtr(50),
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "maximum: current=50, new=100 → false",
			constraintName: "maximum",
			currentValue:   intPtr(50),
			newValue:       intPtr(100),
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "pattern: current=nil, new=non-nil → true",
			constraintName: "pattern",
			currentValue:   (*regexp.Regexp)(nil),
			newValue:       patternA,
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "pattern: current=non-nil, new=nil → false",
			constraintName: "pattern",
			currentValue:   patternA,
			newValue:       (*regexp.Regexp)(nil),
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "pattern: current=non-nil, new=same → true",
			constraintName: "pattern",
			currentValue:   patternA,
			newValue:       patternA,
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "pattern: current=non-nil, new=different → false",
			constraintName: "pattern",
			currentValue:   patternA,
			newValue:       patternB,
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "unknown constraint name → error",
			constraintName: "unknown",
			currentValue:   intPtr(5),
			newValue:       intPtr(10),
			expectedResult: false,
			expectError:    true,
		},
	}

	sc := analysis.StringConstraints{}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := sc.IsMoreOrEqualRestrictive(tc.constraintName, tc.currentValue, tc.newValue)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

// Compile-time check that BooleanConstraints satisfies the TypeConstraints interface.
var _ analysis.TypeConstraints = analysis.BooleanConstraints{}

// TestBooleanConstraintsValidate tests BooleanConstraints.Validate.
func TestBooleanConstraintsValidate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		constraints analysis.BooleanConstraints
		expectError bool
	}{
		{
			name:        "empty constraints returns nil",
			constraints: analysis.BooleanConstraints{},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.constraints.Validate()
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBooleanConstraintsConstraintNames tests BooleanConstraints.ConstraintNames.
func TestBooleanConstraintsConstraintNames(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		constraints   analysis.BooleanConstraints
		expectedNames []string
	}{
		{
			name:          "empty constraints returns empty slice",
			constraints:   analysis.BooleanConstraints{},
			expectedNames: []string{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			names := tc.constraints.ConstraintNames()
			assert.Equal(t, tc.expectedNames, names)
		})
	}
}

// TestBooleanConstraintsGetConstraint tests BooleanConstraints.GetConstraint.
func TestBooleanConstraintsGetConstraint(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		constraints    analysis.BooleanConstraints
		constraintName string
		expectFound    bool
	}{
		{
			name:           "unknown name returns nil, false",
			constraints:    analysis.BooleanConstraints{},
			constraintName: "anything",
			expectFound:    false,
		},
		{
			name:           "another unknown name returns nil, false",
			constraints:    analysis.BooleanConstraints{},
			constraintName: "minimum",
			expectFound:    false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			val, found := tc.constraints.GetConstraint(tc.constraintName)
			assert.Equal(t, tc.expectFound, found)
			assert.Nil(t, val)
		})
	}
}

// TestBooleanConstraintsIsMoreOrEqualRestrictive tests BooleanConstraints.IsMoreOrEqualRestrictive.
func TestBooleanConstraintsIsMoreOrEqualRestrictive(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		constraintName string
		currentValue   interface{}
		newValue       interface{}
		expectError    bool
		errorContains  string
	}{
		{
			name:           "unknown constraint name returns error",
			constraintName: "anything",
			currentValue:   nil,
			newValue:       nil,
			expectError:    true,
			errorContains:  "unknown constraint name: anything",
		},
		{
			name:           "minimum constraint name returns error",
			constraintName: "minimum",
			currentValue:   nil,
			newValue:       nil,
			expectError:    true,
			errorContains:  "unknown constraint name: minimum",
		},
		{
			name:           "maximum constraint name returns error",
			constraintName: "maximum",
			currentValue:   nil,
			newValue:       nil,
			expectError:    true,
			errorContains:  "unknown constraint name: maximum",
		},
	}

	bc := analysis.BooleanConstraints{}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := bc.IsMoreOrEqualRestrictive(tc.constraintName, tc.currentValue, tc.newValue)
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, false, result)
			}
		})
	}
}

// TestCreateConstraintsString tests CreateConstraints with StringType.
func TestCreateConstraintsString(t *testing.T) {
	t.Parallel()

	t.Run("valid: minimum=5", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.StringType, []analysis.Constraint{
			{Name: "minimum", Value: "5"},
		}, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		sc, ok := result.(*analysis.StringConstraints)
		require.True(t, ok)
		require.NotNil(t, sc.Minimum)
		assert.Equal(t, 5, *sc.Minimum)
		assert.Nil(t, sc.Maximum)
		assert.Nil(t, sc.Pattern)
	})

	t.Run("valid: maximum=50", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.StringType, []analysis.Constraint{
			{Name: "maximum", Value: "50"},
		}, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		sc, ok := result.(*analysis.StringConstraints)
		require.True(t, ok)
		assert.Nil(t, sc.Minimum)
		require.NotNil(t, sc.Maximum)
		assert.Equal(t, 50, *sc.Maximum)
		assert.Nil(t, sc.Pattern)
	})

	t.Run("valid: pattern=^[a-z]+$", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.StringType, []analysis.Constraint{
			{Name: "pattern", Value: "^[a-z]+$"},
		}, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		sc, ok := result.(*analysis.StringConstraints)
		require.True(t, ok)
		assert.Nil(t, sc.Minimum)
		assert.Nil(t, sc.Maximum)
		require.NotNil(t, sc.Pattern)
		assert.Equal(t, "^[a-z]+$", sc.Pattern.String())
	})

	t.Run("valid: minimum=5, maximum=50", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.StringType, []analysis.Constraint{
			{Name: "minimum", Value: "5"},
			{Name: "maximum", Value: "50"},
		}, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		sc, ok := result.(*analysis.StringConstraints)
		require.True(t, ok)
		require.NotNil(t, sc.Minimum)
		assert.Equal(t, 5, *sc.Minimum)
		require.NotNil(t, sc.Maximum)
		assert.Equal(t, 50, *sc.Maximum)
		assert.Nil(t, sc.Pattern)
	})

	t.Run("valid: minimum=0, maximum=100, pattern=.*", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.StringType, []analysis.Constraint{
			{Name: "minimum", Value: "0"},
			{Name: "maximum", Value: "100"},
			{Name: "pattern", Value: ".*"},
		}, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		sc, ok := result.(*analysis.StringConstraints)
		require.True(t, ok)
		require.NotNil(t, sc.Minimum)
		assert.Equal(t, 0, *sc.Minimum)
		require.NotNil(t, sc.Maximum)
		assert.Equal(t, 100, *sc.Maximum)
		require.NotNil(t, sc.Pattern)
		assert.Equal(t, ".*", sc.Pattern.String())
	})

	t.Run("valid: no constraints", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.StringType, []analysis.Constraint{}, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		sc, ok := result.(*analysis.StringConstraints)
		require.True(t, ok)
		assert.Nil(t, sc.Minimum)
		assert.Nil(t, sc.Maximum)
		assert.Nil(t, sc.Pattern)
	})

	t.Run("invalid: unknown constraint foo=bar", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.StringType, []analysis.Constraint{
			{Name: "foo", Value: "bar"},
		}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown constraint for String")
		assert.Nil(t, result)
	})

	t.Run("invalid: non-integer minimum=abc", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.StringType, []analysis.Constraint{
			{Name: "minimum", Value: "abc"},
		}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be an integer")
		assert.Nil(t, result)
	})

	t.Run("invalid: non-integer maximum=xyz", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.StringType, []analysis.Constraint{
			{Name: "maximum", Value: "xyz"},
		}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be an integer")
		assert.Nil(t, result)
	})

	t.Run("invalid: invalid regex pattern=[", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.StringType, []analysis.Constraint{
			{Name: "pattern", Value: "["},
		}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regex")
		assert.Nil(t, result)
	})

	t.Run("invalid: empty pattern", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.StringType, []analysis.Constraint{
			{Name: "pattern", Value: ""},
		}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
		assert.Nil(t, result)
	})

	t.Run("invalid: duplicate minimum", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.StringType, []analysis.Constraint{
			{Name: "minimum", Value: "5"},
			{Name: "minimum", Value: "10"},
		}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate constraint name")
		assert.Nil(t, result)
	})

	t.Run("invalid: min > max", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.StringType, []analysis.Constraint{
			{Name: "minimum", Value: "50"},
			{Name: "maximum", Value: "10"},
		}, nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("invalid: negative minimum", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.StringType, []analysis.Constraint{
			{Name: "minimum", Value: "-1"},
		}, nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("invalid: negative maximum", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.StringType, []analysis.Constraint{
			{Name: "maximum", Value: "-1"},
		}, nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestCreateConstraintsBoolean tests CreateConstraints with BooleanType.
func TestCreateConstraintsBoolean(t *testing.T) {
	t.Parallel()

	t.Run("any constraint returns error", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.BooleanType, []analysis.Constraint{
			{Name: "foo", Value: "bar"},
		}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Boolean does not accept constraints")
		assert.Nil(t, result)
	})

	t.Run("no constraints returns BooleanConstraints", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.BooleanType, []analysis.Constraint{}, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		_, ok := result.(*analysis.BooleanConstraints)
		require.True(t, ok, "expected *BooleanConstraints type")
	})
}
