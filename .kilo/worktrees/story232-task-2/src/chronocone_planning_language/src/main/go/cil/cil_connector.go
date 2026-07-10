// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

// CILConnectorNode represents a reference to a connector by name.
// It implements the CILNode interface and serves as a leaf node
// in CIL expression trees, representing an input data source.
type CILConnectorNode struct {
	ConnectorName string
}

func (c *CILConnectorNode) cilNode() {}
