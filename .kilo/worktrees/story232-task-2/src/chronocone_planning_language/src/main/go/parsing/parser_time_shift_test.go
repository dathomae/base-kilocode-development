// Package parsing_test contains tests for the parser time shift data structure.
package parsing_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parsing"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// ParserTimeShift Construction Tests
// ============================================================================

// TestParserTimeShift_Construction verifies ParserTimeShift construction
// and field access for various configurations.
func TestParserTimeShift_Construction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		shift     parsing.ParserTimeShift
		wantDir   string
		wantField string
		wantAmt   float64
	}{
		{
			name:      "backward year amount 2",
			shift:     parsing.ParserTimeShift{Direction: "backward", Field: "year", Amount: 2.0},
			wantDir:   "backward",
			wantField: "year",
			wantAmt:   2.0,
		},
		{
			name:      "default amount is 1.0",
			shift:     parsing.ParserTimeShift{Direction: "backward", Field: "day", Amount: 1.0},
			wantDir:   "backward",
			wantField: "day",
			wantAmt:   1.0,
		},
		{
			name:      "forward month",
			shift:     parsing.ParserTimeShift{Direction: "forward", Field: "month", Amount: 1.0},
			wantDir:   "forward",
			wantField: "month",
			wantAmt:   1.0,
		},
		{
			name:      "zero value constructor",
			shift:     parsing.ParserTimeShift{},
			wantDir:   "",
			wantField: "",
			wantAmt:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantDir, tt.shift.Direction)
			assert.Equal(t, tt.wantField, tt.shift.Field)
			assert.Equal(t, tt.wantAmt, tt.shift.Amount)
		})
	}
}

// ============================================================================
// ParserInputDefinition TimeShift Field Tests
// ============================================================================

// TestParserInputDefinition_TimeShift verifies that ParserInputDefinition
// correctly carries an optional TimeShift field.
func TestParserInputDefinition_TimeShift(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		def           parsing.ParserInputDefinition
		wantNilShift  bool
		wantDirection string
		wantField     string
		wantAmount    float64
	}{
		{
			name:         "default TimeShift is nil",
			def:          parsing.ParserInputDefinition{},
			wantNilShift: true,
		},
		{
			name: "full input definition with populated TimeShift",
			def: parsing.ParserInputDefinition{
				Kind:  "metric",
				Name:  "sales",
				Alias: "s",
				WhereClause: &parsing.ParserWhereClause{
					AndGroups: []parsing.ParserWhereAndGroup{
						{Conditions: []parsing.ParserWhereCondition{
							{Aspect: "ACTUALS", Negated: false},
						}},
					},
				},
				TimeShift: &parsing.ParserTimeShift{
					Direction: "backward",
					Field:     "year",
					Amount:    2.0,
				},
				NameLocation: symboltable.SourceLocation{Line: 10, Column: 5},
			},
			wantNilShift:  false,
			wantDirection: "backward",
			wantField:     "year",
			wantAmount:    2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.wantNilShift {
				assert.Nil(t, tt.def.TimeShift)
			} else {
				assert.NotNil(t, tt.def.TimeShift)
				assert.Equal(t, tt.wantDirection, tt.def.TimeShift.Direction)
				assert.Equal(t, tt.wantField, tt.def.TimeShift.Field)
				assert.Equal(t, tt.wantAmount, tt.def.TimeShift.Amount)
			}
		})
	}
}
