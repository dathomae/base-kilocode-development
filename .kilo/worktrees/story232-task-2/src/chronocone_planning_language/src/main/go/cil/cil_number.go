// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

// CILNumber represents a numeric literal with unit annotation in the CIL expression tree.
type CILNumber struct {
	Value         float64
	AnnotatedUnit string
}

// cilNode marks CILNumber as implementing the CILNode interface.
func (c *CILNumber) cilNode() {}
