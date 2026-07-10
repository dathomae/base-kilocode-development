// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

// CILPower represents exponentiation with derived unit in the CIL expression tree.
type CILPower struct {
	Base        CILNode
	Exponent    int
	DerivedUnit string
}

// cilNode marks CILPower as implementing the CILNode interface.
func (c *CILPower) cilNode() {}
