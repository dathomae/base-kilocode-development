// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import "testing"

var _ CILNode = (*CILNumber)(nil)

func TestCILNumberConstruction(t *testing.T) {
	n := &CILNumber{Value: 3.14, AnnotatedUnit: "test_unit"}
	if n.Value != 3.14 {
		t.Errorf("expected Value 3.14, got %f", n.Value)
	}
	if n.AnnotatedUnit != "test_unit" {
		t.Errorf("expected AnnotatedUnit 'test_unit', got %q", n.AnnotatedUnit)
	}
}

func TestCILNumberZeroValue(t *testing.T) {
	n := &CILNumber{}
	if n.Value != 0.0 {
		t.Errorf("expected zero Value, got %f", n.Value)
	}
	if n.AnnotatedUnit != "" {
		t.Errorf("expected zero AnnotatedUnit, got %q", n.AnnotatedUnit)
	}
}
