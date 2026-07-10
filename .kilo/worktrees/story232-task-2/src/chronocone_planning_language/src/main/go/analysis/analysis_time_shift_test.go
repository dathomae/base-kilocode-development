package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/stretchr/testify/assert"
)

// TestAnalysisTimeShift_BackwardYear tests constructing an AnalysisTimeShift
// with Direction="backward", Field="year", Amount=2.0.
func TestAnalysisTimeShift_BackwardYear(t *testing.T) {
	t.Parallel()

	shift := analysis.AnalysisTimeShift{
		Direction: "backward",
		Field:     "year",
		Amount:    2.0,
	}

	assert.Equal(t, "backward", shift.Direction)
	assert.Equal(t, "year", shift.Field)
	assert.Equal(t, 2.0, shift.Amount)
}

// TestAnalysisTimeShift_ForwardMonth tests constructing an AnalysisTimeShift
// with Direction="forward", Field="month", and default Amount.
func TestAnalysisTimeShift_ForwardMonth(t *testing.T) {
	t.Parallel()

	shift := analysis.AnalysisTimeShift{
		Direction: "forward",
		Field:     "month",
		Amount:    1.0,
	}

	assert.Equal(t, "forward", shift.Direction)
	assert.Equal(t, "month", shift.Field)
	assert.Equal(t, 1.0, shift.Amount)
}

// TestAnalysisInputDefinition_NilTimeShift tests that AnalysisInputDefinition
// has a nil TimeShift by default.
func TestAnalysisInputDefinition_NilTimeShift(t *testing.T) {
	t.Parallel()

	def := analysis.AnalysisInputDefinition{
		Kind: "metric",
		Name: "Revenue",
	}

	assert.Nil(t, def.TimeShift)
	assert.Equal(t, "metric", def.Kind)
	assert.Equal(t, "Revenue", def.Name)
}

// TestAnalysisInputDefinition_WithTimeShift tests AnalysisInputDefinition
// with a populated AnalysisTimeShift.
func TestAnalysisInputDefinition_WithTimeShift(t *testing.T) {
	t.Parallel()

	def := analysis.AnalysisInputDefinition{
		Kind: "metric",
		Name: "Revenue",
		TimeShift: &analysis.AnalysisTimeShift{
			Direction: "backward",
			Field:     "day",
			Amount:    7.0,
		},
	}

	assert.NotNil(t, def.TimeShift)
	assert.Equal(t, "backward", def.TimeShift.Direction)
	assert.Equal(t, "day", def.TimeShift.Field)
	assert.Equal(t, 7.0, def.TimeShift.Amount)
}

// TestAnalysisTimeShift_ValidConstraintFields tests all valid Instant constraint
// field keywords as AnalysisTimeShift Field values.
func TestAnalysisTimeShift_ValidConstraintFields(t *testing.T) {
	t.Parallel()

	fields := []string{
		"year",
		"month",
		"day",
		"hour",
		"minute",
		"second",
		"day_of_week",
		"week_number",
	}

	for _, field := range fields {
		field := field // capture range variable
		t.Run(field, func(t *testing.T) {
			t.Parallel()

			shift := analysis.AnalysisTimeShift{
				Direction: "backward",
				Field:     field,
				Amount:    1.0,
			}

			assert.Equal(t, field, shift.Field)
			assert.Equal(t, "backward", shift.Direction)
			assert.Equal(t, 1.0, shift.Amount)
		})
	}
}
