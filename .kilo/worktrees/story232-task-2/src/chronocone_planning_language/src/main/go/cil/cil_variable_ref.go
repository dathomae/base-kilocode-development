// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

// CILVariableRef represents a reference to a pattern-bound variable by name
// in the CIL expression tree. It is distinct from CILUnitRef, which references
// a unit group.
type CILVariableRef struct {
	Name string
}

// cilNode marks CILVariableRef as implementing the CILNode interface.
func (c *CILVariableRef) cilNode() {}
