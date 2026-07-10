// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================
// Integration tests: full unit-checking pipeline
// =============================================================

// TestIntegration_ValidPipeline_ConversionBridgesGap verifies that when a
// metric computes a unit different from the IO spec, a supplied conversion
// pair allows CheckMetricOutputUnit to succeed.
func TestIntegration_ValidPipeline_ConversionBridgesGap(t *testing.T) {
	// Metric formula results in Kilometer.
	metricRoot := &CILUnitRef{UnitGroup: "Kilometer"}

	// IO spec expects Meter; a Kilometer↔Meter conversion is provided.
	conversions := []CILConversionPair{
		{SourceName: "Meter", TargetName: "Kilometer"},
	}

	err := CheckMetricOutputUnit(metricRoot, "Meter", conversions)
	require.NoError(t, err, "expected conversion to bridge Kilometer → Meter")
}

// TestIntegration_MissingConversion_ReturnsError verifies that when no
// conversion pair exists between the computed unit and the expected unit,
// CheckMetricOutputUnit returns a descriptive error.
func TestIntegration_MissingConversion_ReturnsError(t *testing.T) {
	// Metric formula results in EUR.
	metricRoot := &CILUnitRef{UnitGroup: "EUR"}

	// IO spec expects USD, but no EUR↔USD conversion is provided.
	err := CheckMetricOutputUnit(metricRoot, "USD", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "EUR")
	assert.Contains(t, err.Error(), "USD")
}

// TestIntegration_DimensionalMismatch_ReturnsError verifies that a
// dimensional mismatch (e.g. USD*Meter cannot be converted to USD) produces
// an error from CheckMetricOutputUnit.
func TestIntegration_DimensionalMismatch_ReturnsError(t *testing.T) {
	// Metric formula: USD * Meter  =>  computed unit is USD*Meter.
	metricRoot := &CILBinaryOp{
		Op:    operator.Multiply,
		Left:  &CILUnitRef{UnitGroup: "USD"},
		Right: &CILUnitRef{UnitGroup: "Meter"},
	}

	// IO spec expects USD, but USD*Meter is dimensionally incompatible.
	err := CheckMetricOutputUnit(metricRoot, "USD", nil)
	require.Error(t, err)
	msg := err.Error()
	assert.Contains(t, msg, "USD")
	assert.Contains(t, msg, "Meter")
}

// TestIntegration_IncompatibleAddSub_ReturnsError verifies that adding or
// subtracting quantities with incompatible units causes ComputeUnitExpression
// to return an error.
func TestIntegration_IncompatibleAddSub_ReturnsError(t *testing.T) {
	// USD + EUR is dimensionally illegal.
	metricRoot := &CILBinaryOp{
		Op:    operator.Add,
		Left:  &CILUnitRef{UnitGroup: "USD"},
		Right: &CILUnitRef{UnitGroup: "EUR"},
	}

	_, err := ComputeUnitExpression(metricRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "incompatible units")
	assert.Contains(t, err.Error(), "USD")
	assert.Contains(t, err.Error(), "EUR")
}

// TestIntegration_UnitCancellation_SimplifiesCorrectly verifies that
// multiplying a unit by a dimensionless ratio (USD/USD) cancels the USD
// and leaves only the original unit.
func TestIntegration_UnitCancellation_SimplifiesCorrectly(t *testing.T) {
	// C * (USD / USD)  =>  after simplification, unit is C.
	metricRoot := &CILBinaryOp{
		Op:   operator.Multiply,
		Left: &CILUnitRef{UnitGroup: "C"},
		Right: &CILBinaryOp{
			Op:    operator.Divide,
			Left:  &CILUnitRef{UnitGroup: "USD"},
			Right: &CILUnitRef{UnitGroup: "USD"},
		},
	}

	ue, err := ComputeUnitExpression(metricRoot)
	require.NoError(t, err)

	simplified := ue.Simplify()
	assert.True(t, simplified.Equals(unitRefExpression("C")),
		"expected unit C after cancellation, got %s", simplified.String())

	// The full pipeline should also accept the simplified unit against "C".
	err = CheckMetricOutputUnit(metricRoot, "C", nil)
	require.NoError(t, err)
}

// TestIntegration_DimensionlessMetric_NoConversions_Passes verifies that
// plain numeric metrics without any unit annotations continue to pass when
// the IO spec is also dimensionless.
func TestIntegration_DimensionlessMetric_NoConversions_Passes(t *testing.T) {
	metricRoot := &CILNumber{Value: 42.0}

	err := CheckMetricOutputUnit(metricRoot, "", nil)
	require.NoError(t, err, "expected dimensionless metric to pass")
}

// TestIntegration_FullPipeline_ComplexExpressionMissingConversion verifies
// that a complex composite unit expression fails when the required conversion
// is absent. The current implementation validates the full computed unit name
// against conversions, so partial unit conversions inside composite
// expressions are not supported.
func TestIntegration_FullPipeline_ComplexExpressionMissingConversion(t *testing.T) {
	metricRoot := &CILBinaryOp{
		Op: operator.Divide,
		Left: &CILBinaryOp{
			Op:    operator.Multiply,
			Left:  &CILUnitRef{UnitGroup: "Kilometer"},
			Right: &CILUnitRef{UnitGroup: "USD"},
		},
		Right: &CILUnitRef{UnitGroup: "Second"},
	}

	err := CheckMetricOutputUnit(metricRoot, "Meter*USD/Second", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Kilometer")
	assert.Contains(t, err.Error(), "Meter")
}

// TestIntegration_SubtractIncompatibleUnits_PropagatesError verifies that
// subtraction with mismatched units also returns an error.
func TestIntegration_SubtractIncompatibleUnits_PropagatesError(t *testing.T) {
	metricRoot := &CILBinaryOp{
		Op:    operator.Subtract,
		Left:  &CILUnitRef{UnitGroup: "Meter"},
		Right: &CILUnitRef{UnitGroup: "Second"},
	}

	_, err := ComputeUnitExpression(metricRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "incompatible units")
}

// TestIntegration_NestedUnaryOp_PreservesUnit verifies that a negation unary
// op does not alter the unit of its operand, and the full pipeline still
// matches.
func TestIntegration_NestedUnaryOp_PreservesUnit(t *testing.T) {
	metricRoot := &CILUnaryOp{
		Op:      operator.Negate,
		Operand: &CILUnitRef{UnitGroup: "Meter"},
	}

	err := CheckMetricOutputUnit(metricRoot, "Meter", nil)
	require.NoError(t, err)
}

// TestIntegration_PowerOp_PropagatesUnit verifies that raising a unit
// reference to a power correctly scales the exponent in the UnitExpression.
func TestIntegration_PowerOp_PropagatesUnit(t *testing.T) {
	// Meter^2
	metricRoot := &CILPower{
		Base:     &CILUnitRef{UnitGroup: "Meter"},
		Exponent: 2,
	}

	ue, err := ComputeUnitExpression(metricRoot)
	require.NoError(t, err)
	assert.Equal(t, 2.0, ue.Numerator["Meter"])

	// Simplification preserves the exponent.
	simplified := ue.Simplify()
	assert.Equal(t, 2.0, simplified.Numerator["Meter"])
}

// TestIntegration_DivisionCancellationViaPipeline verifies that a division
// that results in a dimensionless ratio is accepted by CheckMetricOutputUnit
// when the IO spec is also dimensionless.
func TestIntegration_DivisionCancellationViaPipeline(t *testing.T) {
	// USD / USD
	metricRoot := &CILBinaryOp{
		Op:    operator.Divide,
		Left:  &CILUnitRef{UnitGroup: "USD"},
		Right: &CILUnitRef{UnitGroup: "USD"},
	}

	err := CheckMetricOutputUnit(metricRoot, "", nil)
	require.NoError(t, err, "expected dimensionless result to match empty ioSpecOutputUnit")
}
