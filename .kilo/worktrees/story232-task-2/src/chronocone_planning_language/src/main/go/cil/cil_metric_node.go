// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

// CILMetricNode is the root node for a metric's CIL expression tree.
type CILMetricNode struct {
	CILNode   CILNode       // The root CIL node (for Increment 1: a CILLiteral)
	InputRefs []CILInputRef // References to input metrics/dimensions
}

// CILInputRef represents a reference to an input metric or dimension.
type CILInputRef struct {
	Kind     string // "metric" or "dimension" or "connector"
	Name     string // The name of the referenced entity
	UnitName string // Optional unit name for connector/metric references
}

// cilNode marks CILInputRef as implementing the CILNode interface.
func (c *CILInputRef) cilNode() {}
