// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

// CILUnitRef represents a reference to a unit group by name in the CIL expression tree.
type CILUnitRef struct {
	UnitGroup string
}

// cilNode marks CILUnitRef as implementing the CILNode interface.
func (c *CILUnitRef) cilNode() {}
