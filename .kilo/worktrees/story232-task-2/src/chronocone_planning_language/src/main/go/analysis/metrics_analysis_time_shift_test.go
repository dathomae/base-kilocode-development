// Package analysis_test contains unit tests for the time shift validation logic
// in the metrics analysis phase.
//
// These tests verify that:
//   - Valid time-shift constraint field keywords are accepted
//   - Invalid/unrecognized field keywords produce errors
//   - Self-reference with a time shift is allowed (valid time-series dependency)
//   - Self-reference without a time shift is prohibited
//   - Multiple inputs with mixed time-shift configurations are handled correctly
//
// Story: Story191 - CPL Metrics Time Shift Parse and Analyze
// Task: Task 5 - Implement time shift validation logic
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parsing"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Valid Time-Shift Constraint Field Tests
// ============================================================================

// TestAnalyzeMetricScope_ValidTimeShiftFields verifies that each valid Instant
// constraint field keyword is accepted in a time-shift clause.
func TestAnalyzeMetricScope_ValidTimeShiftFields(t *testing.T) {
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

			// Arrange: Build symbol table with a metric input that has a time shift
			root := symboltable.NewGlobalSymbolTable()
			metricsScope := root.AddChild(symboltable.MetricsScope)
			metricScope := metricsScope.AddChild(symboltable.MetricScope)
			metricDef := &parsing.ParserMetricDefinition{
				Name: "my_metric",
				Inputs: []parsing.ParserInputDefinition{
					{
						Kind: "metric",
						Name: "demand_forecast",
						TimeShift: &parsing.ParserTimeShift{
							Direction: "backward",
							Field:     field,
							Amount:    1.0,
						},
						NameLocation: symboltable.SourceLocation{Line: 3, Column: 3},
					},
				},
				CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 0}},
				SourceLocation:   symboltable.SourceLocation{Line: 2, Column: 1},
			}
			metricScope.Data = metricDef
			nameCopy := metricDef.Name
			_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)
			analyzer := analysis.NewCPLAnalyzer()
			defs := analyzer.GetAnalysisDefinitions()
			defs.Measures["demand_forecast"] = analysis.AnalysisMeasureDefinition{
				Name:     "demand_forecast",
				TypeName: "Number",
				UnitExpression: &analysis.UnitReference{Name: "Unit"},
			}

			// Act: Analyze the metric scope
			err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)

			// Assert: Should succeed and preserve the time shift
			require.NoError(t, err, "Analysis should succeed for valid time-shift field %q", field)

			resultDefs := analyzer.GetAnalysisDefinitions()
			require.Contains(t, resultDefs.Metrics, "my_metric")
			metric := resultDefs.Metrics["my_metric"]
			require.Len(t, metric.Inputs, 1)
			require.NotNil(t, metric.Inputs[0].TimeShift)
			assert.Equal(t, "backward", metric.Inputs[0].TimeShift.Direction)
			assert.Equal(t, field, metric.Inputs[0].TimeShift.Field)
			assert.Equal(t, 1.0, metric.Inputs[0].TimeShift.Amount)
		})
	}
}

// TestAnalyzeMetricScope_TimeShiftBackwardYear verifies backward(year) is accepted.
func TestAnalyzeMetricScope_TimeShiftBackwardYear(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name: "my_metric",
		Inputs: []parsing.ParserInputDefinition{
			{
				Kind: "metric",
				Name: "input_a",
				TimeShift: &parsing.ParserTimeShift{
					Direction: "backward",
					Field:     "year",
					Amount:    2.0,
				},
				NameLocation: symboltable.SourceLocation{Line: 3, Column: 3},
			},
		},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 0}},
		SourceLocation:   symboltable.SourceLocation{Line: 2, Column: 1},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["input_a"] = analysis.AnalysisMeasureDefinition{
		Name:     "input_a",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)
	require.NoError(t, err)

	resultDefs := analyzer.GetAnalysisDefinitions()
	metric := resultDefs.Metrics["my_metric"]
	require.NotNil(t, metric.Inputs[0].TimeShift)
	assert.Equal(t, "backward", metric.Inputs[0].TimeShift.Direction)
	assert.Equal(t, "year", metric.Inputs[0].TimeShift.Field)
	assert.Equal(t, 2.0, metric.Inputs[0].TimeShift.Amount)
}

// TestAnalyzeMetricScope_TimeShiftForwardMonth verifies forward(month) is accepted.
func TestAnalyzeMetricScope_TimeShiftForwardMonth(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name: "my_metric",
		Inputs: []parsing.ParserInputDefinition{
			{
				Kind: "metric",
				Name: "input_a",
				TimeShift: &parsing.ParserTimeShift{
					Direction: "forward",
					Field:     "month",
					Amount:    3.0,
				},
				NameLocation: symboltable.SourceLocation{Line: 3, Column: 3},
			},
		},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 0}},
		SourceLocation:   symboltable.SourceLocation{Line: 2, Column: 1},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["input_a"] = analysis.AnalysisMeasureDefinition{
		Name:     "input_a",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)
	require.NoError(t, err)

	resultDefs := analyzer.GetAnalysisDefinitions()
	metric := resultDefs.Metrics["my_metric"]
	require.NotNil(t, metric.Inputs[0].TimeShift)
	assert.Equal(t, "forward", metric.Inputs[0].TimeShift.Direction)
	assert.Equal(t, "month", metric.Inputs[0].TimeShift.Field)
	assert.Equal(t, 3.0, metric.Inputs[0].TimeShift.Amount)
}

// ============================================================================
// Invalid Time-Shift Field Error Tests
// ============================================================================

// TestAnalyzeMetricScope_InvalidTimeShiftField_ReturnsError verifies that an
// unrecognized time-shift field keyword produces an error.
func TestAnalyzeMetricScope_InvalidTimeShiftField_ReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		direction string
		field     string
	}{
		{name: "backward with foo", direction: "backward", field: "foo"},
		{name: "forward with empty", direction: "forward", field: ""},
		{name: "backward with day_of_month", direction: "backward", field: "day_of_month"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := symboltable.NewGlobalSymbolTable()
			metricsScope := root.AddChild(symboltable.MetricsScope)
			metricScope := metricsScope.AddChild(symboltable.MetricScope)
			metricDef := &parsing.ParserMetricDefinition{
				Name: "my_metric",
				Inputs: []parsing.ParserInputDefinition{
					{
						Kind: "metric",
						Name: "input_a",
						TimeShift: &parsing.ParserTimeShift{
							Direction: tt.direction,
							Field:     tt.field,
							Amount:    1.0,
						},
						NameLocation: symboltable.SourceLocation{Line: 3, Column: 3},
					},
				},
				CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 0}},
				SourceLocation:   symboltable.SourceLocation{Line: 2, Column: 1},
			}
			metricScope.Data = metricDef
			nameCopy := metricDef.Name
			_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)
			analyzer := analysis.NewCPLAnalyzer()
			defs := analyzer.GetAnalysisDefinitions()
			defs.Measures["input_a"] = analysis.AnalysisMeasureDefinition{
				Name:     "input_a",
				TypeName: "Number",
				UnitExpression: &analysis.UnitReference{Name: "Unit"},
			}

			err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)
			require.Error(t, err, "Should return error for invalid time-shift field %q", tt.field)
			assert.Contains(t, err.Error(), "invalid time-shift field")
			assert.Contains(t, err.Error(), tt.field)
			assert.Contains(t, err.Error(), "my_metric")
		})
	}
}

// ============================================================================
// Self-Reference Tests
// ============================================================================

// TestAnalyzeMetricScope_SelfReferenceWithTimeShift_Allowed verifies that a
// metric referencing itself with a backward time shift is accepted as a valid
// time-series dependency.
func TestAnalyzeMetricScope_SelfReferenceWithTimeShift_Allowed(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name: "revenue",
		Inputs: []parsing.ParserInputDefinition{
			{
				Kind: "metric",
				Name: "revenue",
				TimeShift: &parsing.ParserTimeShift{
					Direction: "backward",
					Field:     "year",
					Amount:    1.0,
				},
				NameLocation: symboltable.SourceLocation{Line: 3, Column: 3},
			},
		},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 0}},
		SourceLocation:   symboltable.SourceLocation{Line: 2, Column: 1},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)
	analyzer := analysis.NewCPLAnalyzer()

	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)
	require.NoError(t, err, "Self-reference with time shift should be allowed")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Metrics, "revenue")
	metric := resultDefs.Metrics["revenue"]
	require.Len(t, metric.Inputs, 1)
	require.NotNil(t, metric.Inputs[0].TimeShift)
	assert.Equal(t, "backward", metric.Inputs[0].TimeShift.Direction)
	assert.Equal(t, "year", metric.Inputs[0].TimeShift.Field)
}

// TestAnalyzeMetricScope_SelfReferenceWithForwardTimeShift_Allowed verifies that
// a metric referencing itself with a forward time shift is accepted.
func TestAnalyzeMetricScope_SelfReferenceWithForwardTimeShift_Allowed(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name: "revenue",
		Inputs: []parsing.ParserInputDefinition{
			{
				Kind: "metric",
				Name: "revenue",
				TimeShift: &parsing.ParserTimeShift{
					Direction: "forward",
					Field:     "month",
					Amount:    3.0,
				},
				NameLocation: symboltable.SourceLocation{Line: 3, Column: 3},
			},
		},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 0}},
		SourceLocation:   symboltable.SourceLocation{Line: 2, Column: 1},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)
	analyzer := analysis.NewCPLAnalyzer()

	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)
	require.NoError(t, err, "Self-reference with forward time shift should be allowed")

	resultDefs := analyzer.GetAnalysisDefinitions()
	metric := resultDefs.Metrics["revenue"]
	require.NotNil(t, metric.Inputs[0].TimeShift)
	assert.Equal(t, "forward", metric.Inputs[0].TimeShift.Direction)
	assert.Equal(t, "month", metric.Inputs[0].TimeShift.Field)
	assert.Equal(t, 3.0, metric.Inputs[0].TimeShift.Amount)
}

// TestAnalyzeMetricScope_SelfReferenceWithoutTimeShift_ReturnsError verifies that
// a metric referencing itself without a time shift is prohibited.
func TestAnalyzeMetricScope_SelfReferenceWithoutTimeShift_ReturnsError(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name: "revenue",
		Inputs: []parsing.ParserInputDefinition{
			{
				Kind:         "metric",
				Name:         "revenue",
				NameLocation: symboltable.SourceLocation{Line: 3, Column: 3},
			},
		},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 0}},
		SourceLocation:   symboltable.SourceLocation{Line: 2, Column: 1},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)
	analyzer := analysis.NewCPLAnalyzer()

	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)
	require.Error(t, err, "Self-reference without time shift should return error")
	assert.Contains(t, err.Error(), "references itself")
	assert.Contains(t, err.Error(), "revenue")
	assert.Contains(t, err.Error(), "line")
	assert.Contains(t, err.Error(), "column")
}

// ============================================================================
// Multiple Inputs with Mixed Time Shifts
// ============================================================================

// TestAnalyzeMetricScope_MultipleInputs_MixedTimeShifts verifies that a metric
// with multiple inputs — some time-shifted, some not — is handled correctly,
// including self-reference detection.
func TestAnalyzeMetricScope_MultipleInputs_MixedTimeShifts(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name: "revenue",
		Inputs: []parsing.ParserInputDefinition{
			{
				Kind:         "metric",
				Name:         "costs",
				NameLocation: symboltable.SourceLocation{Line: 3, Column: 3},
			},
			{
				Kind: "metric",
				Name: "revenue",
				TimeShift: &parsing.ParserTimeShift{
					Direction: "backward",
					Field:     "year",
					Amount:    1.0,
				},
				NameLocation: symboltable.SourceLocation{Line: 4, Column: 3},
			},
			{
				Kind: "metric",
				Name: "demand_forecast",
				TimeShift: &parsing.ParserTimeShift{
					Direction: "forward",
					Field:     "month",
					Amount:    2.0,
				},
				NameLocation: symboltable.SourceLocation{Line: 5, Column: 3},
			},
		},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 0}},
		SourceLocation:   symboltable.SourceLocation{Line: 2, Column: 1},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["costs"] = analysis.AnalysisMeasureDefinition{
		Name:     "costs",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}
	defs.Measures["demand_forecast"] = analysis.AnalysisMeasureDefinition{
		Name:     "demand_forecast",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)
	require.NoError(t, err, "Analysis should succeed for mixed time-shift inputs")

	resultDefs := analyzer.GetAnalysisDefinitions()
	metric := resultDefs.Metrics["revenue"]
	require.Len(t, metric.Inputs, 3)

	// First input: no time shift
	assert.Nil(t, metric.Inputs[0].TimeShift)
	assert.Equal(t, "costs", metric.Inputs[0].Name)

	// Second input: self-reference with time shift
	require.NotNil(t, metric.Inputs[1].TimeShift)
	assert.Equal(t, "backward", metric.Inputs[1].TimeShift.Direction)
	assert.Equal(t, "year", metric.Inputs[1].TimeShift.Field)

	// Third input: external reference with time shift
	require.NotNil(t, metric.Inputs[2].TimeShift)
	assert.Equal(t, "forward", metric.Inputs[2].TimeShift.Direction)
	assert.Equal(t, "month", metric.Inputs[2].TimeShift.Field)
	assert.Equal(t, 2.0, metric.Inputs[2].TimeShift.Amount)
}

// TestAnalyzeMetricScope_MultipleInputs_SelfReferenceWithoutShift_ReturnsError
// verifies that self-reference detection works when there are multiple inputs.
func TestAnalyzeMetricScope_MultipleInputs_SelfReferenceWithoutShift_ReturnsError(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name: "revenue",
		Inputs: []parsing.ParserInputDefinition{
			{
				Kind:         "metric",
				Name:         "costs",
				NameLocation: symboltable.SourceLocation{Line: 3, Column: 3},
			},
			{
				Kind:         "metric",
				Name:         "revenue",
				NameLocation: symboltable.SourceLocation{Line: 4, Column: 3},
			},
		},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 0}},
		SourceLocation:   symboltable.SourceLocation{Line: 2, Column: 1},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["costs"] = analysis.AnalysisMeasureDefinition{
		Name:     "costs",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)
	require.Error(t, err, "Should return error when one of multiple inputs is a self-reference without time shift")
	assert.Contains(t, err.Error(), "references itself")
	assert.Contains(t, err.Error(), "revenue")
	assert.Contains(t, err.Error(), "line")
	assert.Contains(t, err.Error(), "column")
}

// ============================================================================
// Nil Time-Shift Tests
// ============================================================================

// TestAnalyzeMetricScope_NilTimeShift_Passes verifies that an input without
// a time shift (nil TimeShift) is handled correctly and produces a nil
// AnalysisTimeShift.
func TestAnalyzeMetricScope_NilTimeShift_Passes(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	metricsScope := root.AddChild(symboltable.MetricsScope)
	metricScope := metricsScope.AddChild(symboltable.MetricScope)
	metricDef := &parsing.ParserMetricDefinition{
		Name: "my_metric",
		Inputs: []parsing.ParserInputDefinition{
			{
				Kind:         "metric",
				Name:         "input_a",
				NameLocation: symboltable.SourceLocation{Line: 3, Column: 3},
			},
		},
		CalculatePostfix: []parsing.PostfixElement{parsing.PostfixLiteral{Value: 0}},
		SourceLocation:   symboltable.SourceLocation{Line: 2, Column: 1},
	}
	metricScope.Data = metricDef
	nameCopy := metricDef.Name
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, metricDef)
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["input_a"] = analysis.AnalysisMeasureDefinition{
		Name:     "input_a",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	err := analysis.AnalyzeMetricScope(metricScope, nil, analyzer)
	require.NoError(t, err)

	resultDefs := analyzer.GetAnalysisDefinitions()
	metric := resultDefs.Metrics["my_metric"]
	require.Len(t, metric.Inputs, 1)
	assert.Nil(t, metric.Inputs[0].TimeShift)
}
