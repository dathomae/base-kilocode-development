// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"strings"
	"testing"
)

// TestValidateConversion_ValidForward verifies that a direct forward
// conversion pair is accepted.
func TestValidateConversion_ValidForward(t *testing.T) {
	conversions := []CILConversionPair{
		{SourceName: "Meter", TargetName: "Kilometer"},
	}
	err := ValidateConversion("Meter", "Kilometer", conversions, nil)
	if err != nil {
		t.Errorf("ValidateConversion(Meter, Kilometer) = %v, want nil", err)
	}
}

// TestValidateConversion_ReverseDirection verifies that a conversion pair
// matches when the requested direction is the reverse of the stored pair.
func TestValidateConversion_ReverseDirection(t *testing.T) {
	conversions := []CILConversionPair{
		{SourceName: "Meter", TargetName: "Kilometer"},
	}
	err := ValidateConversion("Kilometer", "Meter", conversions, nil)
	if err != nil {
		t.Errorf("ValidateConversion(Kilometer, Meter) = %v, want nil", err)
	}
}

// TestValidateConversion_MissingConversion verifies that an error is returned
// when no conversion pair bridges the two units.
func TestValidateConversion_MissingConversion(t *testing.T) {
	conversions := []CILConversionPair{
		{SourceName: "Celsius", TargetName: "Fahrenheit"},
	}
	err := ValidateConversion("Celsius", "Meter", conversions, nil)
	if err == nil {
		t.Fatal("ValidateConversion(Celsius, Meter) = nil, want error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "no conversion available") {
		t.Errorf("error message %q does not contain 'no conversion available'", msg)
	}
}

// TestValidateConversion_SameUnit verifies that identical source and target
// units always succeed without requiring a conversion pair.
func TestValidateConversion_SameUnit(t *testing.T) {
	err := ValidateConversion("Meter", "Meter", nil, nil)
	if err != nil {
		t.Errorf("ValidateConversion(Meter, Meter) = %v, want nil", err)
	}

	err = ValidateConversion("Meter", "Meter", []CILConversionPair{}, nil)
	if err != nil {
		t.Errorf("ValidateConversion(Meter, Meter, empty) = %v, want nil", err)
	}
}

// TestValidateConversion_EmptyConversions verifies that an empty or nil
// conversion list produces an error when the units differ.
func TestValidateConversion_EmptyConversions(t *testing.T) {
	err := ValidateConversion("Meter", "Kilometer", nil, nil)
	if err == nil {
		t.Fatal("ValidateConversion(Meter, Kilometer, nil) = nil, want error")
	}

	err = ValidateConversion("Meter", "Kilometer", []CILConversionPair{}, nil)
	if err == nil {
		t.Fatal("ValidateConversion(Meter, Kilometer, empty) = nil, want error")
	}
}

// TestValidateConversion_MultipleConversions verifies that the correct pair
// is found among many unrelated conversions.
func TestValidateConversion_MultipleConversions(t *testing.T) {
	conversions := []CILConversionPair{
		{SourceName: "Celsius", TargetName: "Fahrenheit"},
		{SourceName: "Meter", TargetName: "Kilometer"},
		{SourceName: "Gram", TargetName: "Kilogram"},
	}
	err := ValidateConversion("Meter", "Kilometer", conversions, nil)
	if err != nil {
		t.Errorf("ValidateConversion(Meter, Kilometer) = %v, want nil", err)
	}
}

// TestValidateConversionClosure_CompleteGroup verifies that closure succeeds
// when all unit pairs in a group have a conversion pair.
func TestValidateConversionClosure_CompleteGroup(t *testing.T) {
	unitGroups := map[string]string{
		"Meter":      "DistanceUnits",
		"Kilometer":  "DistanceUnits",
		"Centimeter": "DistanceUnits",
	}
	pairs := []CILConversionPair{
		{SourceName: "Meter", TargetName: "Kilometer"},
		{SourceName: "Kilometer", TargetName: "Centimeter"},
		{SourceName: "Meter", TargetName: "Centimeter"},
	}
	err := ValidateConversionClosure(pairs, unitGroups)
	if err != nil {
		t.Errorf("ValidateConversionClosure() = %v, want nil", err)
	}
}

// TestValidateConversionClosure_MissingPair verifies that an error is returned
// when some pairs in a unit group lack conversion pairs.
func TestValidateConversionClosure_MissingPair(t *testing.T) {
	unitGroups := map[string]string{
		"Meter":      "DistanceUnits",
		"Kilometer":  "DistanceUnits",
		"Centimeter": "DistanceUnits",
	}
	pairs := []CILConversionPair{
		{SourceName: "Meter", TargetName: "Kilometer"},
	}
	err := ValidateConversionClosure(pairs, unitGroups)
	if err == nil {
		t.Fatal("ValidateConversionClosure() = nil, want error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "no conversion path") {
		t.Errorf("error message %q does not contain 'no conversion path'", msg)
	}
	if !strings.Contains(msg, "DistanceUnits") {
		t.Errorf("error message %q does not contain 'DistanceUnits'", msg)
	}
}

// TestValidateConversionClosure_EmptyConversions verifies that an empty
// conversion slice produces an error when the unit group has 2+ units.
func TestValidateConversionClosure_EmptyConversions(t *testing.T) {
	unitGroups := map[string]string{
		"Meter":     "DistanceUnits",
		"Kilometer": "DistanceUnits",
	}
	err := ValidateConversionClosure([]CILConversionPair{}, unitGroups)
	if err == nil {
		t.Fatal("ValidateConversionClosure(empty) = nil, want error")
	}
}

// TestValidateConversionClosure_SingleUnit verifies that a unit group with a
// single unit trivially has closure.
func TestValidateConversionClosure_SingleUnit(t *testing.T) {
	unitGroups := map[string]string{
		"Meter": "DistanceUnits",
	}
	err := ValidateConversionClosure([]CILConversionPair{}, unitGroups)
	if err != nil {
		t.Errorf("ValidateConversionClosure(single) = %v, want nil", err)
	}
}

// TestValidateConversionClosure_MultipleGroups verifies that multiple unit
// groups are validated independently and closure succeeds when each is complete.
func TestValidateConversionClosure_MultipleGroups(t *testing.T) {
	unitGroups := map[string]string{
		"Meter":     "DistanceUnits",
		"Kilometer": "DistanceUnits",
		"USD":       "CurrencyUnits",
		"EUR":       "CurrencyUnits",
	}
	pairs := []CILConversionPair{
		{SourceName: "Meter", TargetName: "Kilometer"},
		{SourceName: "USD", TargetName: "EUR"},
	}
	err := ValidateConversionClosure(pairs, unitGroups)
	if err != nil {
		t.Errorf("ValidateConversionClosure() = %v, want nil", err)
	}
}

// TestValidateConversionClosure_MissingInOneGroup_SuccessfulOther verifies that
// an error is reported only for the incomplete group.
func TestValidateConversionClosure_MissingInOneGroup_SuccessfulOther(t *testing.T) {
	unitGroups := map[string]string{
		"Meter":      "DistanceUnits",
		"Kilometer":  "DistanceUnits",
		"Centimeter": "DistanceUnits",
		"USD":        "CurrencyUnits",
		"EUR":        "CurrencyUnits",
	}
	pairs := []CILConversionPair{
		{SourceName: "Meter", TargetName: "Kilometer"},
		{SourceName: "USD", TargetName: "EUR"},
	}
	err := ValidateConversionClosure(pairs, unitGroups)
	if err == nil {
		t.Fatal("ValidateConversionClosure() = nil, want error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "DistanceUnits") {
		t.Errorf("error message %q does not contain 'DistanceUnits'", msg)
	}
	if strings.Contains(msg, "CurrencyUnits") {
		t.Errorf("error message %q should not mention 'CurrencyUnits'", msg)
	}
}

// TestValidateConversionWithClosure_PassesWhenClosureOK verifies that
// ValidateConversionWithClosure succeeds when closure is satisfied.
func TestValidateConversionWithClosure_PassesWhenClosureOK(t *testing.T) {
	unitGroups := map[string]string{
		"Meter":     "DistanceUnits",
		"Kilometer": "DistanceUnits",
	}
	pairs := []CILConversionPair{
		{SourceName: "Meter", TargetName: "Kilometer"},
	}
	err := ValidateConversionWithClosure("Meter", "Kilometer", pairs, unitGroups)
	if err != nil {
		t.Errorf("ValidateConversionWithClosure() = %v, want nil", err)
	}
}

// TestValidateConversionWithClosure_FailsWhenClosureBroken verifies that
// ValidateConversionWithClosure fails when the unit group is not closed.
func TestValidateConversionWithClosure_FailsWhenClosureBroken(t *testing.T) {
	unitGroups := map[string]string{
		"Meter":      "DistanceUnits",
		"Kilometer":  "DistanceUnits",
		"Centimeter": "DistanceUnits",
	}
	pairs := []CILConversionPair{
		{SourceName: "Meter", TargetName: "Kilometer"},
	}
	err := ValidateConversionWithClosure("Meter", "Kilometer", pairs, unitGroups)
	if err == nil {
		t.Fatal("ValidateConversionWithClosure() = nil, want error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "no conversion path") {
		t.Errorf("error message %q does not contain 'no conversion path'", msg)
	}
	if !strings.Contains(msg, "DistanceUnits") {
		t.Errorf("error message %q does not contain 'DistanceUnits'", msg)
	}
}

// TestValidateConversionWithClosure_EmptyUnitGroupsOK verifies that nil or
// empty unitGroups skips closure checking.
func TestValidateConversionWithClosure_EmptyUnitGroupsOK(t *testing.T) {
	pairs := []CILConversionPair{
		{SourceName: "Meter", TargetName: "Kilometer"},
	}
	err := ValidateConversionWithClosure("Meter", "Kilometer", pairs, nil)
	if err != nil {
		t.Errorf("ValidateConversionWithClosure(nil groups) = %v, want nil", err)
	}
}

// TestValidateConversionWithClosure_EmptyUnitGroupsStillValidatesPair verifies
// that even with nil unitGroups, the direct pair validation still runs.
func TestValidateConversionWithClosure_EmptyUnitGroupsStillValidatesPair(t *testing.T) {
	pairs := []CILConversionPair{
		{SourceName: "Meter", TargetName: "Kilometer"},
	}
	err := ValidateConversionWithClosure("Meter", "Centimeter", pairs, nil)
	if err == nil {
		t.Fatal("ValidateConversionWithClosure() = nil, want error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "no conversion available") {
		t.Errorf("error message %q does not contain 'no conversion available'", msg)
	}
}
