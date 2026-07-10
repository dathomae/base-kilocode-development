// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import "testing"

// Compile-time assertion: CILVariableRef implements CILNode
var _ CILNode = (*CILVariableRef)(nil)

func TestCILVariableRefConstruction(t *testing.T) {
	ref := &CILVariableRef{Name: "m"}
	if ref.Name != "m" {
		t.Errorf("expected Name 'm', got %q", ref.Name)
	}
}

func TestCILVariableRefZeroValue(t *testing.T) {
	ref := &CILVariableRef{}
	if ref.Name != "" {
		t.Errorf("expected zero Name, got %q", ref.Name)
	}
}
