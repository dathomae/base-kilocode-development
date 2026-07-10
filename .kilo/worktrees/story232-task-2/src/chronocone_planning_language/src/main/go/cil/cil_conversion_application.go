// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

// CILConversionApplication wraps a sub-expression with a conversion to be
// applied at runtime. It represents the application of a unit or measure
// conversion to an operand in the CIL expression tree.
type CILConversionApplication struct {
	Operand    CILNode
	Conversion CILConversionPair
	FromUnit   string
	ToUnit     string
}

// cilNode marks CILConversionApplication as implementing the CILNode interface.
func (c *CILConversionApplication) cilNode() {}
