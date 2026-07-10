// Package analysis_test contains tests for the CPL analyzer connectors integration.
//
// These tests verify that the AnalysisDefinitions struct correctly includes
// and initializes the Connectors and CILTrees maps, and that GetAnalysisDefinitions()
// returns a properly populated struct.
//
// Reference: ERO004 Connectors Analysis Phase - Task 2
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// AnalysisDefinitions Connectors Map Tests
// ============================================================================

// TestAnalysisDefinitions_ConnectorsMap_Initialized verifies that the Connectors
// map in AnalysisDefinitions is properly initialized and can store connector definitions.
func TestAnalysisDefinitions_ConnectorsMap_Initialized(t *testing.T) {
	t.Parallel()

	// Arrange: Create an AnalysisDefinitions with initialized Connectors map
	defs := &analysis.AnalysisDefinitions{
		Concepts:   make(map[string]analysis.AnalysisConceptDefinition),
		Types:      make(map[string]analysis.AnalysisTypeDefinition),
		Units:      make(map[string]analysis.AnalysisUnitDefinition),
		Measures:   make(map[string]analysis.AnalysisMeasureDefinition),
		Connectors: make(map[string]analysis.AnalysisConnectorDefinition),
		CILTrees:   make(map[string]cil.CILNode),
	}

	// Assert: Verify Connectors map is initialized and empty
	require.NotNil(t, defs.Connectors, "Connectors map should be initialized")
	assert.Empty(t, defs.Connectors, "Connectors map should be empty initially")
}

// TestAnalysisDefinitions_ConnectorsMap_StoreAndRetrieve verifies that connector
// definitions can be stored and retrieved from the Connectors map.
func TestAnalysisDefinitions_ConnectorsMap_StoreAndRetrieve(t *testing.T) {
	t.Parallel()

	// Arrange: Create AnalysisDefinitions with initialized maps
	defs := &analysis.AnalysisDefinitions{
		Concepts:   make(map[string]analysis.AnalysisConceptDefinition),
		Types:      make(map[string]analysis.AnalysisTypeDefinition),
		Units:      make(map[string]analysis.AnalysisUnitDefinition),
		Measures:   make(map[string]analysis.AnalysisMeasureDefinition),
		Connectors: make(map[string]analysis.AnalysisConnectorDefinition),
		CILTrees:   make(map[string]cil.CILNode),
	}

	// Arrange: Create a connector definition
	connectorDef := analysis.AnalysisConnectorDefinition{
		Name:           "SalesData",
		Dimensions:     []*analysis.AnalysisConceptDefinition{},
		CellType:       &analysis.AnalysisMeasureDefinition{Name: "Amount"},
		SourceLocation: symboltable.SourceLocation{Line: 10, Column: 5},
	}

	// Act: Store the connector definition
	defs.Connectors["SalesData"] = connectorDef

	// Assert: Verify the connector can be retrieved
	retrieved, exists := defs.Connectors["SalesData"]
	require.True(t, exists, "Connector should exist in map")
	assert.Equal(t, "SalesData", retrieved.Name)
	assert.Equal(t, 10, retrieved.SourceLocation.Line)
}

// TestAnalysisDefinitions_ConnectorsMap_StoreMultiple verifies that multiple
// connector definitions can be stored in the Connectors map.
func TestAnalysisDefinitions_ConnectorsMap_StoreMultiple(t *testing.T) {
	t.Parallel()

	// Arrange: Create AnalysisDefinitions with initialized maps
	defs := &analysis.AnalysisDefinitions{
		Concepts:   make(map[string]analysis.AnalysisConceptDefinition),
		Types:      make(map[string]analysis.AnalysisTypeDefinition),
		Units:      make(map[string]analysis.AnalysisUnitDefinition),
		Measures:   make(map[string]analysis.AnalysisMeasureDefinition),
		Connectors: make(map[string]analysis.AnalysisConnectorDefinition),
		CILTrees:   make(map[string]cil.CILNode),
	}

	// Arrange: Create multiple connector definitions
	connectors := []analysis.AnalysisConnectorDefinition{
		{
			Name:           "SalesData",
			Dimensions:     []*analysis.AnalysisConceptDefinition{},
			CellType:       &analysis.AnalysisMeasureDefinition{Name: "Amount"},
			SourceLocation: symboltable.SourceLocation{Line: 1, Column: 1},
		},
		{
			Name:           "Inventory",
			Dimensions:     []*analysis.AnalysisConceptDefinition{},
			CellType:       &analysis.AnalysisMeasureDefinition{Name: "Quantity"},
			SourceLocation: symboltable.SourceLocation{Line: 5, Column: 1},
		},
		{
			Name:           "Customers",
			Dimensions:     []*analysis.AnalysisConceptDefinition{},
			CellType:       &analysis.AnalysisMeasureDefinition{Name: "Count"},
			SourceLocation: symboltable.SourceLocation{Line: 10, Column: 1},
		},
	}

	// Act: Store all connectors
	for _, conn := range connectors {
		defs.Connectors[conn.Name] = conn
	}

	// Assert: Verify all connectors are stored
	assert.Len(t, defs.Connectors, 3, "Should have 3 connectors")
	assert.Contains(t, defs.Connectors, "SalesData")
	assert.Contains(t, defs.Connectors, "Inventory")
	assert.Contains(t, defs.Connectors, "Customers")
}

// TestAnalysisDefinitions_ConnectorsMap_NonExistentKey verifies behavior
// when accessing a non-existent connector key.
func TestAnalysisDefinitions_ConnectorsMap_NonExistentKey(t *testing.T) {
	t.Parallel()

	// Arrange: Create AnalysisDefinitions with initialized maps
	defs := &analysis.AnalysisDefinitions{
		Concepts:   make(map[string]analysis.AnalysisConceptDefinition),
		Types:      make(map[string]analysis.AnalysisTypeDefinition),
		Units:      make(map[string]analysis.AnalysisUnitDefinition),
		Measures:   make(map[string]analysis.AnalysisMeasureDefinition),
		Connectors: make(map[string]analysis.AnalysisConnectorDefinition),
		CILTrees:   make(map[string]cil.CILNode),
	}

	// Act: Try to retrieve non-existent connector
	_, exists := defs.Connectors["NonExistent"]

	// Assert: Verify connector does not exist
	assert.False(t, exists, "Non-existent connector should not be found")
}

// ============================================================================
// AnalysisDefinitions CILTrees Map Tests
// ============================================================================

// TestAnalysisDefinitions_CILTreesMap_Initialized verifies that the CILTrees
// map in AnalysisDefinitions is properly initialized and can store CIL nodes.
func TestAnalysisDefinitions_CILTreesMap_Initialized(t *testing.T) {
	t.Parallel()

	// Arrange: Create an AnalysisDefinitions with initialized CILTrees map
	defs := &analysis.AnalysisDefinitions{
		Concepts:   make(map[string]analysis.AnalysisConceptDefinition),
		Types:      make(map[string]analysis.AnalysisTypeDefinition),
		Units:      make(map[string]analysis.AnalysisUnitDefinition),
		Measures:   make(map[string]analysis.AnalysisMeasureDefinition),
		Connectors: make(map[string]analysis.AnalysisConnectorDefinition),
		CILTrees:   make(map[string]cil.CILNode),
	}

	// Assert: Verify CILTrees map is initialized and empty
	require.NotNil(t, defs.CILTrees, "CILTrees map should be initialized")
	assert.Empty(t, defs.CILTrees, "CILTrees map should be empty initially")
}

// TestAnalysisDefinitions_CILTreesMap_StoreAndRetrieve verifies that CIL nodes
// can be stored and retrieved from the CILTrees map.
func TestAnalysisDefinitions_CILTreesMap_StoreAndRetrieve(t *testing.T) {
	t.Parallel()

	// Arrange: Create AnalysisDefinitions with initialized maps
	defs := &analysis.AnalysisDefinitions{
		Concepts:   make(map[string]analysis.AnalysisConceptDefinition),
		Types:      make(map[string]analysis.AnalysisTypeDefinition),
		Units:      make(map[string]analysis.AnalysisUnitDefinition),
		Measures:   make(map[string]analysis.AnalysisMeasureDefinition),
		Connectors: make(map[string]analysis.AnalysisConnectorDefinition),
		CILTrees:   make(map[string]cil.CILNode),
	}

	// Arrange: Create a CIL connector reference node
	cilNode := &cil.CILConnectorNode{ConnectorName: "SalesData"}

	// Act: Store the CIL node
	defs.CILTrees["SalesData"] = cilNode

	// Assert: Verify the CIL node can be retrieved
	retrieved, exists := defs.CILTrees["SalesData"]
	require.True(t, exists, "CIL node should exist in map")
	require.NotNil(t, retrieved)

	// Type assert to verify it's a connector reference node
	connectorNode, ok := retrieved.(*cil.CILConnectorNode)
	require.True(t, ok, "Retrieved node should be CILConnectorNode")
	assert.Equal(t, "SalesData", connectorNode.ConnectorName)
}

// TestAnalysisDefinitions_CILTreesMap_StoreMultiple verifies that multiple
// CIL nodes can be stored in the CILTrees map.
func TestAnalysisDefinitions_CILTreesMap_StoreMultiple(t *testing.T) {
	t.Parallel()

	// Arrange: Create AnalysisDefinitions with initialized maps
	defs := &analysis.AnalysisDefinitions{
		Concepts:   make(map[string]analysis.AnalysisConceptDefinition),
		Types:      make(map[string]analysis.AnalysisTypeDefinition),
		Units:      make(map[string]analysis.AnalysisUnitDefinition),
		Measures:   make(map[string]analysis.AnalysisMeasureDefinition),
		Connectors: make(map[string]analysis.AnalysisConnectorDefinition),
		CILTrees:   make(map[string]cil.CILNode),
	}

	// Arrange: Create multiple CIL nodes
	nodes := map[string]cil.CILNode{
		"SalesData": &cil.CILConnectorNode{ConnectorName: "SalesData"},
		"Inventory": &cil.CILConnectorNode{ConnectorName: "Inventory"},
		"Customers": &cil.CILConnectorNode{ConnectorName: "Customers"},
	}

	// Act: Store all CIL nodes
	for name, node := range nodes {
		defs.CILTrees[name] = node
	}

	// Assert: Verify all CIL nodes are stored
	assert.Len(t, defs.CILTrees, 3, "Should have 3 CIL nodes")
	assert.Contains(t, defs.CILTrees, "SalesData")
	assert.Contains(t, defs.CILTrees, "Inventory")
	assert.Contains(t, defs.CILTrees, "Customers")
}

// TestAnalysisDefinitions_CILTreesMap_NonExistentKey verifies behavior
// when accessing a non-existent CIL tree key.
func TestAnalysisDefinitions_CILTreesMap_NonExistentKey(t *testing.T) {
	t.Parallel()

	// Arrange: Create AnalysisDefinitions with initialized maps
	defs := &analysis.AnalysisDefinitions{
		Concepts:   make(map[string]analysis.AnalysisConceptDefinition),
		Types:      make(map[string]analysis.AnalysisTypeDefinition),
		Units:      make(map[string]analysis.AnalysisUnitDefinition),
		Measures:   make(map[string]analysis.AnalysisMeasureDefinition),
		Connectors: make(map[string]analysis.AnalysisConnectorDefinition),
		CILTrees:   make(map[string]cil.CILNode),
	}

	// Act: Try to retrieve non-existent CIL node
	_, exists := defs.CILTrees["NonExistent"]

	// Assert: Verify CIL node does not exist
	assert.False(t, exists, "Non-existent CIL node should not be found")
}

// ============================================================================
// CPLAnalyzer GetAnalysisDefinitions Tests
// ============================================================================

// TestCPLAnalyzer_GetAnalysisDefinitions_ReturnsInitializedConnectorsMap verifies
// that GetAnalysisDefinitions returns an AnalysisDefinitions with an initialized
// Connectors map.
func TestCPLAnalyzer_GetAnalysisDefinitions_ReturnsInitializedConnectorsMap(t *testing.T) {
	t.Parallel()

	// Arrange: Create a new analyzer
	analyzer := analysis.NewCPLAnalyzer()

	// Act: Get analysis definitions
	defs := analyzer.GetAnalysisDefinitions()

	// Assert: Verify Connectors map is initialized
	require.NotNil(t, defs, "GetAnalysisDefinitions should not return nil")
	require.NotNil(t, defs.Connectors, "Connectors map should be initialized")
	assert.Empty(t, defs.Connectors, "Connectors map should be empty for new analyzer")
}

// TestCPLAnalyzer_GetAnalysisDefinitions_ReturnsInitializedCILTreesMap verifies
// that GetAnalysisDefinitions returns an AnalysisDefinitions with an initialized
// CILTrees map.
func TestCPLAnalyzer_GetAnalysisDefinitions_ReturnsInitializedCILTreesMap(t *testing.T) {
	t.Parallel()

	// Arrange: Create a new analyzer
	analyzer := analysis.NewCPLAnalyzer()

	// Act: Get analysis definitions
	defs := analyzer.GetAnalysisDefinitions()

	// Assert: Verify CILTrees map is initialized
	require.NotNil(t, defs, "GetAnalysisDefinitions should not return nil")
	require.NotNil(t, defs.CILTrees, "CILTrees map should be initialized")
	assert.Empty(t, defs.CILTrees, "CILTrees map should be empty for new analyzer")
}

// TestCPLAnalyzer_GetAnalysisDefinitions_ReturnsAllInitializedMaps verifies
// that GetAnalysisDefinitions returns an AnalysisDefinitions with all maps
// (including new Connectors and CILTrees) properly initialized.
func TestCPLAnalyzer_GetAnalysisDefinitions_ReturnsAllInitializedMaps(t *testing.T) {
	t.Parallel()

	// Arrange: Create a new analyzer
	analyzer := analysis.NewCPLAnalyzer()

	// Act: Get analysis definitions
	defs := analyzer.GetAnalysisDefinitions()

	// Assert: Verify all maps are initialized
	require.NotNil(t, defs, "GetAnalysisDefinitions should not return nil")
	require.NotNil(t, defs.Concepts, "Concepts map should be initialized")
	require.NotNil(t, defs.Types, "Types map should be initialized")
	require.NotNil(t, defs.Units, "Units map should be initialized")
	require.NotNil(t, defs.Measures, "Measures map should be initialized")
	require.NotNil(t, defs.Connectors, "Connectors map should be initialized")
	require.NotNil(t, defs.CILTrees, "CILTrees map should be initialized")

	// Verify all maps are empty for a new analyzer
	assert.Empty(t, defs.Concepts, "Concepts map should be empty")
	assert.Empty(t, defs.Types, "Types map should be empty")
	assert.Empty(t, defs.Units, "Units map should be empty")
	assert.Empty(t, defs.Measures, "Measures map should be empty")
	assert.Empty(t, defs.Connectors, "Connectors map should be empty")
	assert.Empty(t, defs.CILTrees, "CILTrees map should be empty")
}

// TestCPLAnalyzer_GetAnalysisDefinitions_PreservesConnectorData verifies
// that connector data added to the analyzer is preserved in the returned
// AnalysisDefinitions.
func TestCPLAnalyzer_GetAnalysisDefinitions_PreservesConnectorData(t *testing.T) {
	t.Parallel()

	// Arrange: Create a new analyzer
	analyzer := analysis.NewCPLAnalyzer()

	// Arrange: Create and add a connector definition
	connectorDef := analysis.AnalysisConnectorDefinition{
		Name:           "SalesData",
		Dimensions:     []*analysis.AnalysisConceptDefinition{},
		CellType:       &analysis.AnalysisMeasureDefinition{Name: "Amount"},
		SourceLocation: symboltable.SourceLocation{Line: 10, Column: 5},
	}

	// Access the internal connectors map through the returned definitions
	// and add the connector (simulating what analysis would do)
	defs := analyzer.GetAnalysisDefinitions()
	defs.Connectors["SalesData"] = connectorDef

	// Act: Get definitions again
	defs2 := analyzer.GetAnalysisDefinitions()

	// Assert: Verify connector data is preserved
	retrieved, exists := defs2.Connectors["SalesData"]
	require.True(t, exists, "Connector should exist")
	assert.Equal(t, "SalesData", retrieved.Name)
}

// TestCPLAnalyzer_GetAnalysisDefinitions_PreservesCILTreeData verifies
// that CIL tree data added to the analyzer is preserved in the returned
// AnalysisDefinitions.
func TestCPLAnalyzer_GetAnalysisDefinitions_PreservesCILTreeData(t *testing.T) {
	t.Parallel()

	// Arrange: Create a new analyzer
	analyzer := analysis.NewCPLAnalyzer()

	// Arrange: Create and add a CIL node
	cilNode := &cil.CILConnectorNode{ConnectorName: "SalesData"}

	// Access the internal CILTrees map through the returned definitions
	defs := analyzer.GetAnalysisDefinitions()
	defs.CILTrees["SalesData"] = cilNode

	// Act: Get definitions again
	defs2 := analyzer.GetAnalysisDefinitions()

	// Assert: Verify CIL node is preserved
	retrieved, exists := defs2.CILTrees["SalesData"]
	require.True(t, exists, "CIL node should exist")
	require.NotNil(t, retrieved)
}

// TestCPLAnalyzer_NewCPLAnalyzer_InitializesAllMaps verifies that NewCPLAnalyzer
// properly initializes all internal maps including connectors and cilTrees.
func TestCPLAnalyzer_NewCPLAnalyzer_InitializesAllMaps(t *testing.T) {
	t.Parallel()

	// Act: Create a new analyzer
	analyzer := analysis.NewCPLAnalyzer()

	// Assert: Verify analyzer was created (non-nil)
	require.NotNil(t, analyzer, "NewCPLAnalyzer should return non-nil analyzer")

	// Get definitions to verify internal state
	defs := analyzer.GetAnalysisDefinitions()
	require.NotNil(t, defs, "GetAnalysisDefinitions should not return nil")

	// Verify all maps are initialized and empty
	assert.NotNil(t, defs.Concepts, "Concepts map should be initialized")
	assert.NotNil(t, defs.Types, "Types map should be initialized")
	assert.NotNil(t, defs.Units, "Units map should be initialized")
	assert.NotNil(t, defs.Measures, "Measures map should be initialized")
	assert.NotNil(t, defs.Connectors, "Connectors map should be initialized")
	assert.NotNil(t, defs.CILTrees, "CILTrees map should be initialized")
}
