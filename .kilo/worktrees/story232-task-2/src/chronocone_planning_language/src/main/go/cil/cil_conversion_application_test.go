// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import "testing"

var _ CILNode = (*CILConversionApplication)(nil)

// TestCILConversionApplicationConstruction verifies all fields are stored correctly.
func TestCILConversionApplicationConstruction(t *testing.T) {
	operand := &CILNumber{Value: 100.0, AnnotatedUnit: "m"}
	conv := CILConversionPair{
		SourceName:  "meters",
		TargetName:  "kilometers",
		TargetVar:   "km",
		SourceVar:   "m",
		Calculation: "m / 1000.0",
		SourceGroup: "DistanceUnits",
		TargetGroup: "DistanceUnits",
	}
	node := &CILConversionApplication{
		Operand:    operand,
		Conversion: conv,
		FromUnit:   "meters",
		ToUnit:     "kilometers",
	}

	if node.Operand == nil {
		t.Errorf("CILConversionApplication.Operand is nil, expected non-nil")
	}
	if node.Conversion.SourceName != "meters" {
		t.Errorf("CILConversionApplication.Conversion.SourceName = %q, expected %q", node.Conversion.SourceName, "meters")
	}
	if node.Conversion.TargetName != "kilometers" {
		t.Errorf("CILConversionApplication.Conversion.TargetName = %q, expected %q", node.Conversion.TargetName, "kilometers")
	}
	if node.FromUnit != "meters" {
		t.Errorf("CILConversionApplication.FromUnit = %q, expected %q", node.FromUnit, "meters")
	}
	if node.ToUnit != "kilometers" {
		t.Errorf("CILConversionApplication.ToUnit = %q, expected %q", node.ToUnit, "kilometers")
	}
}

// TestCILConversionApplicationNilOperand verifies that a CILConversionApplication
// with a nil Operand does not panic on construction or field access.
func TestCILConversionApplicationNilOperand(t *testing.T) {
	node := &CILConversionApplication{
		Operand:    nil,
		Conversion: CILConversionPair{},
		FromUnit:   "",
		ToUnit:     "",
	}

	if node.Operand != nil {
		t.Errorf("CILConversionApplication.Operand expected nil, got non-nil")
	}
	if node.FromUnit != "" {
		t.Errorf("CILConversionApplication.FromUnit = %q, expected empty string", node.FromUnit)
	}
	if node.ToUnit != "" {
		t.Errorf("CILConversionApplication.ToUnit = %q, expected empty string", node.ToUnit)
	}
}

// TestCILConversionApplicationNestedOperand verifies that the Operand field can
// hold an arbitrary CILNode, including nested expressions.
func TestCILConversionApplicationNestedOperand(t *testing.T) {
	inner := &CILBinaryOp{
		Op:    0, // Unknown operator, acceptable for structural test
		Left:  &CILNumber{Value: 10.0, AnnotatedUnit: "m"},
		Right: &CILNumber{Value: 2.0, AnnotatedUnit: "s"},
	}
	node := &CILConversionApplication{
		Operand: inner,
		Conversion: CILConversionPair{
			SourceName: "m/s",
			TargetName: "km/h",
		},
		FromUnit: "m/s",
		ToUnit:   "km/h",
	}

	if node.Operand == nil {
		t.Fatalf("CILConversionApplication.Operand is nil, expected non-nil")
	}
	_, ok := node.Operand.(*CILBinaryOp)
	if !ok {
		t.Fatalf("CILConversionApplication.Operand type = %T, expected *CILBinaryOp", node.Operand)
	}
}
