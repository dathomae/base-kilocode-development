// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import "testing"

func TestCILConnectorNodeConstruction(t *testing.T) {
	node := &CILConnectorNode{ConnectorName: "TestConnector"}
	if node.ConnectorName != "TestConnector" {
		t.Errorf("expected ConnectorName to be 'TestConnector', got %q", node.ConnectorName)
	}
}

func TestCILConnectorNodeInterfaceSatisfaction(t *testing.T) {
	node := &CILConnectorNode{ConnectorName: "TestConnector"}
	var _ CILNode = node
}

func TestCILConnectorNodeZeroValue(t *testing.T) {
	node := CILConnectorNode{}
	if node.ConnectorName != "" {
		t.Errorf("expected zero value ConnectorName to be empty, got %q", node.ConnectorName)
	}
}
