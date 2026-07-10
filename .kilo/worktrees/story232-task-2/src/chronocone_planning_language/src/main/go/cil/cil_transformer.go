// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

// CILAbstractTransformer transforms CIL abstract nodes into executable Go code.
// It orchestrates the synthesis process by composing a CILRegistry (for type
// registrations) and a CILCodeGenerator (for code generation), and by walking
// CIL trees to generate code that uses the Chronocone tensor computation library.
type CILAbstractTransformer struct {
	*CILRegistry
	*CILCodeGenerator
}

// NewCILAbstractTransformer creates a new transformer with empty type mappings.
//
// Returns:
//   - *CILAbstractTransformer: A new transformer instance.
func NewCILAbstractTransformer() *CILAbstractTransformer {
	registry := NewCILRegistry()
	return &CILAbstractTransformer{
		CILRegistry:      registry,
		CILCodeGenerator: NewCILCodeGenerator(registry),
	}
}
