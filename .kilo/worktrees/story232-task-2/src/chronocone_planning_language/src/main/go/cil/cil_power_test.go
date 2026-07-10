// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import "testing"

var _ CILNode = (*CILPower)(nil)

func TestCILPowerConstruction(t *testing.T) {
	base := &CILUnitRef{UnitGroup: "test_unit"}
	p := &CILPower{Base: base, Exponent: 2, DerivedUnit: "test_derived"}
	if p.Base != base {
		t.Error("Base not set correctly")
	}
	if p.Exponent != 2 {
		t.Errorf("expected Exponent 2, got %d", p.Exponent)
	}
	if p.DerivedUnit != "test_derived" {
		t.Errorf("expected DerivedUnit 'test_derived', got %q", p.DerivedUnit)
	}
}

func TestCILPowerZeroValue(t *testing.T) {
	p := &CILPower{}
	if p.Base != nil {
		t.Error("expected nil Base")
	}
	if p.Exponent != 0 {
		t.Errorf("expected zero Exponent, got %d", p.Exponent)
	}
	if p.DerivedUnit != "" {
		t.Errorf("expected zero DerivedUnit, got %q", p.DerivedUnit)
	}
}
