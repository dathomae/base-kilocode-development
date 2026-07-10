// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import "testing"

// Compile-time assertion: CILUnitRef implements CILNode
var _ CILNode = (*CILUnitRef)(nil)

func TestCILUnitRefConstruction(t *testing.T) {
	ref := &CILUnitRef{UnitGroup: "some_unit"}
	if ref.UnitGroup != "some_unit" {
		t.Errorf("expected UnitGroup 'some_unit', got %q", ref.UnitGroup)
	}
}

func TestCILUnitRefZeroValue(t *testing.T) {
	ref := &CILUnitRef{}
	if ref.UnitGroup != "" {
		t.Errorf("expected zero UnitGroup, got %q", ref.UnitGroup)
	}
}
