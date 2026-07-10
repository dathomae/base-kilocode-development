// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

// CILNode is the interface for all nodes in the CIL expression tree.
// This interface is implemented by concrete node types such as CILLiteral.
type CILNode interface {
	// cilNode is a marker method to prevent external types from implementing CILNode.
	cilNode()
}
