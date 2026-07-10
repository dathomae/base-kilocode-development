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
// Phase B: CheckMetricOutputUnit tests
// =============================================================

func TestCheckMetricOutputUnit_MatchingUnitsPass(t *testing.T) {
	// Metric computes Meter, IO spec expects Meter → nil error
	metricRoot := &CILUnitRef{UnitGroup: "Meter"}
	err := CheckMetricOutputUnit(metricRoot, "Meter", nil)
	require.NoError(t, err)
}

func TestCheckMetricOutputUnit_MismatchedWithConversionPass(t *testing.T) {
	// Metric computes Kilometer, IO spec expects Meter, Kilometer→Meter conversion exists → nil error
	metricRoot := &CILUnitRef{UnitGroup: "Kilometer"}
	conversions := []CILConversionPair{
		{SourceName: "Kilometer", TargetName: "Meter"},
	}
	err := CheckMetricOutputUnit(metricRoot, "Meter", conversions)
	require.NoError(t, err)
}

func TestCheckMetricOutputUnit_MismatchedWithoutConversionFail(t *testing.T) {
	// Metric computes EUR, IO spec expects USD, no conversion → error
	metricRoot := &CILUnitRef{UnitGroup: "EUR"}
	conversions := []CILConversionPair{
		{SourceName: "Celsius", TargetName: "Fahrenheit"},
	}
	err := CheckMetricOutputUnit(metricRoot, "USD", conversions)
	require.Error(t, err)
	msg := err.Error()
	assert.Contains(t, msg, "EUR")
	assert.Contains(t, msg, "USD")
}

func TestCheckMetricOutputUnit_IncompatibleAdditionFail(t *testing.T) {
	// Metric has USD + EUR in its formula → error from ComputeUnitExpression propagated
	metricRoot := &CILBinaryOp{
		Op:    operator.Add,
		Left:  &CILUnitRef{UnitGroup: "USD"},
		Right: &CILUnitRef{UnitGroup: "EUR"},
	}
	err := CheckMetricOutputUnit(metricRoot, "USD", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "incompatible units")
}

func TestCheckMetricOutputUnit_SimpleDimensionlessPass(t *testing.T) {
	// Metric has a plain number, IO spec expects dimensionless → nil
	metricRoot := &CILNumber{Value: 42.0}
	err := CheckMetricOutputUnit(metricRoot, "", nil)
	require.NoError(t, err)
}

func TestCheckMetricOutputUnit_DifferentSimplifiedFormsPass(t *testing.T) {
	// Metric computes Meter*Second/Second, IO spec expects Meter → passes after simplification
	metricRoot := &CILBinaryOp{
		Op: operator.Divide,
		Left: &CILBinaryOp{
			Op:    operator.Multiply,
			Left:  &CILUnitRef{UnitGroup: "Meter"},
			Right: &CILUnitRef{UnitGroup: "Second"},
		},
		Right: &CILUnitRef{UnitGroup: "Second"},
	}
	err := CheckMetricOutputUnit(metricRoot, "Meter", nil)
	require.NoError(t, err)
}

func TestCheckMetricOutputUnit_NilCILRootFail(t *testing.T) {
	err := CheckMetricOutputUnit(nil, "Meter", nil)
	require.Error(t, err)
}

func TestCheckMetricOutputUnit_ConnectorWithUnitNameAndConversion(t *testing.T) {
	metricRoot := &CILInputRef{Kind: "connector", Name: "connector_a", UnitName: "Kilometer"}
	conversions := []CILConversionPair{{SourceName: "Kilometer", TargetName: "Meter"}}
	err := CheckMetricOutputUnit(metricRoot, "Meter", conversions)
	require.NoError(t, err)
}

func TestCheckMetricOutputUnit_ConnectorWithUnitNameNoConversion(t *testing.T) {
	metricRoot := &CILInputRef{Kind: "connector", Name: "connector_a", UnitName: "Kilometer"}
	err := CheckMetricOutputUnit(metricRoot, "Meter", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no conversion available")
}
