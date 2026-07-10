// Package analysis_test contains tests for the AnalyzeConnectorsScope function.
//
// These tests verify that the connector analysis phase correctly:
//   - Resolves dimension references from the Concepts map
//   - Resolves celltype references from the Measures map
//   - Detects duplicate connector definitions
//   - Reports errors for undefined dimensions or celltypes
//   - Generates CIL trees for valid connectors
//
// Reference: ERO004 Connectors Analysis Phase - Task 4
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
// Happy Path Tests - Valid Connector Processing
// ============================================================================

// TestAnalyzeConnectorsScope_SingleDimension_ValidConnector verifies that a
// connector with a single dimension is correctly processed.
func TestAnalyzeConnectorsScope_SingleDimension_ValidConnector(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with concepts and measures
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()

	// Add a concept for the dimension
	countryConcept := analysis.AnalysisConceptDefinition{
		ConceptName: "Country",
	}
	analyzer.AddNewConcept(&countryConcept)

	// Add a measure for the celltype
	defs.Measures["Amount"] = analysis.AnalysisMeasureDefinition{
		Name:     "Amount",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Dollar"},
	}
	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}

	// Create a ConnectorScope with structured declaration data
	connectorScope := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 5},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Amount",
	}

	context := analysis.NewAnalysisContext()

	// Act: Analyze the connector scope
	err := analysis.AnalyzeConnectorsScope(connectorScope, context, analyzer)

	// Assert: No error occurred
	require.NoError(t, err, "Should not return error for valid connector")

	// Assert: Connector was added to the map
	resultDefs := analyzer.GetAnalysisDefinitions()
	connectorDef, exists := resultDefs.Connectors["SalesData"]
	require.True(t, exists, "SalesData connector should exist in map")

	// Assert: Connector properties are correct
	assert.Equal(t, "SalesData", connectorDef.Name)
	assert.Len(t, connectorDef.Dimensions, 1, "Should have 1 dimension")
	assert.Equal(t, "Country", connectorDef.Dimensions[0].ConceptName)
	assert.Equal(t, "Amount", connectorDef.CellType.Name)
	require.Len(t, connectorDef.Grain, 1, "Grain should be resolved")
	assert.Equal(t, "Day", connectorDef.Grain[0].Name)
	assert.Equal(t, 10, connectorDef.SourceLocation.Line)
	assert.Equal(t, 5, connectorDef.SourceLocation.Column)
}

// TestAnalyzeConnectorsScope_MultipleDimensions_ValidConnector verifies that a
// connector with multiple dimensions is correctly processed.
func TestAnalyzeConnectorsScope_MultipleDimensions_ValidConnector(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with concepts and measures
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()

	// Add concepts for dimensions
	countryConcept := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	productConcept := analysis.AnalysisConceptDefinition{ConceptName: "Product"}
	dateConcept := analysis.AnalysisConceptDefinition{ConceptName: "Date"}
	analyzer.AddNewConcept(&countryConcept)
	analyzer.AddNewConcept(&productConcept)
	analyzer.AddNewConcept(&dateConcept)

	// Add a measure for the celltype
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Dollar"},
	}
	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}

	// Create a ConnectorScope with structured declaration data
	connectorScope := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "InventoryData",
		NameLocation:   symboltable.SourceLocation{Line: 15, Column: 5},
		DimensionNames: []string{"Country", "Product", "Date"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Revenue",
	}

	context := analysis.NewAnalysisContext()

	// Act: Analyze the connector scope
	err := analysis.AnalyzeConnectorsScope(connectorScope, context, analyzer)

	// Assert: No error occurred
	require.NoError(t, err, "Should not return error for valid connector")

	// Assert: Connector was added to the map
	resultDefs := analyzer.GetAnalysisDefinitions()
	connectorDef, exists := resultDefs.Connectors["InventoryData"]
	require.True(t, exists, "InventoryData connector should exist in map")

	// Assert: Connector properties are correct
	assert.Equal(t, "InventoryData", connectorDef.Name)
	require.Len(t, connectorDef.Dimensions, 3, "Should have 3 dimensions")
	assert.Equal(t, "Country", connectorDef.Dimensions[0].ConceptName)
	assert.Equal(t, "Product", connectorDef.Dimensions[1].ConceptName)
	assert.Equal(t, "Date", connectorDef.Dimensions[2].ConceptName)
	assert.Equal(t, "Revenue", connectorDef.CellType.Name)
	require.Len(t, connectorDef.Grain, 1)
	assert.Equal(t, "Day", connectorDef.Grain[0].Name)
}

// TestAnalyzeConnectorsScope_NoDimensions_ValidConnector verifies that a
// connector with no dimensions (only celltype) is correctly processed.
func TestAnalyzeConnectorsScope_NoDimensions_ValidConnector(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with measures
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()

	// Add a measure for the celltype (no concepts needed)
	defs.Measures["TotalCount"] = analysis.AnalysisMeasureDefinition{
		Name:     "TotalCount",
		TypeName: "Integer",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}

	// Create a ConnectorScope with structured declaration data (no dimensions)
	connectorScope := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SummaryData",
		NameLocation:   symboltable.SourceLocation{Line: 20, Column: 5},
		DimensionNames: []string{},
		GrainNames:     []string{"Day"},
		CellTypeName:   "TotalCount",
	}

	context := analysis.NewAnalysisContext()

	// Act: Analyze the connector scope
	err := analysis.AnalyzeConnectorsScope(connectorScope, context, analyzer)

	// Assert: No error occurred
	require.NoError(t, err, "Should not return error for connector without dimensions")

	// Assert: Connector was added to the map
	resultDefs := analyzer.GetAnalysisDefinitions()
	connectorDef, exists := resultDefs.Connectors["SummaryData"]
	require.True(t, exists, "SummaryData connector should exist in map")

	// Assert: Connector has no dimensions
	assert.Equal(t, "SummaryData", connectorDef.Name)
	assert.Empty(t, connectorDef.Dimensions, "Should have no dimensions")
	assert.Equal(t, "TotalCount", connectorDef.CellType.Name)
	require.Len(t, connectorDef.Grain, 1)
	assert.Equal(t, "Day", connectorDef.Grain[0].Name)
}

// TestAnalyzeConnectorsScope_MultipleConnectors_AllValid verifies that multiple
// connectors in the same scope are all processed correctly.
func TestAnalyzeConnectorsScope_MultipleConnectors_AllValid(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with concepts and measures
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()

	// Add concepts
	countryConcept := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	productConcept := analysis.AnalysisConceptDefinition{ConceptName: "Product"}
	analyzer.AddNewConcept(&countryConcept)
	analyzer.AddNewConcept(&productConcept)

	// Add measures
	defs.Measures["Sales"] = analysis.AnalysisMeasureDefinition{Name: "Sales", TypeName: "Number"}
	defs.Measures["Quantity"] = analysis.AnalysisMeasureDefinition{Name: "Quantity", TypeName: "Number"}
	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{Name: "Day", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "day"}}

	// Create parent ConnectorsScope with two child ConnectorScopes
	connectorsScope := symboltable.NewSymbolTable(symboltable.ConnectorsScope)

	// First connector: SalesData, Country, Product, Sales
	connector1 := connectorsScope.AddChild(symboltable.ConnectorScope)
	connector1.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 5, Column: 5},
		DimensionNames: []string{"Country", "Product"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Sales",
	}

	// Second connector: Inventory, Country, Quantity
	connector2 := connectorsScope.AddChild(symboltable.ConnectorScope)
	connector2.Data = &symboltable.ConnectorDeclaration{
		Name:           "Inventory",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 5},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Quantity",
	}

	context := analysis.NewAnalysisContext()

	// Act: Analyze both connector scopes
	err1 := analysis.AnalyzeConnectorsScope(connector1, context, analyzer)
	err2 := analysis.AnalyzeConnectorsScope(connector2, context, analyzer)

	// Assert: No errors occurred
	require.NoError(t, err1, "Should not return error for first connector")
	require.NoError(t, err2, "Should not return error for second connector")

	// Assert: Both connectors exist
	resultDefs := analyzer.GetAnalysisDefinitions()
	assert.Len(t, resultDefs.Connectors, 2, "Should have 2 connectors")
	assert.Contains(t, resultDefs.Connectors, "SalesData")
	assert.Contains(t, resultDefs.Connectors, "Inventory")
}

// ============================================================================
// CIL Tree Generation Tests
// ============================================================================

// TestAnalyzeConnectorsScope_CILTreeGeneration_SingleConnector verifies that
// a CIL tree is generated for a valid connector.
func TestAnalyzeConnectorsScope_CILTreeGeneration_SingleConnector(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with required definitions
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	countryConcept := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConcept)
	defs.Measures["Amount"] = analysis.AnalysisMeasureDefinition{Name: "Amount", TypeName: "Number"}
	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{Name: "Day", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "day"}}

	connectorScope := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 1, Column: 1},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Amount",
	}

	context := analysis.NewAnalysisContext()

	// Act: Analyze the connector
	err := analysis.AnalyzeConnectorsScope(connectorScope, context, analyzer)

	// Assert: No error and CIL tree exists
	require.NoError(t, err)
	resultDefs := analyzer.GetAnalysisDefinitions()

	cilNode, exists := resultDefs.CILTrees["SalesData"]
	require.True(t, exists, "CIL tree should exist for SalesData")
	require.NotNil(t, cilNode)

	// Verify it's a connector reference node with correct name
	connectorNode, ok := cilNode.(*cil.CILConnectorNode)
	require.True(t, ok, "Should be CILConnectorNode")
	assert.Equal(t, "SalesData", connectorNode.ConnectorName)
}

// TestAnalyzeConnectorsScope_CILTreeGeneration_MultipleConnectors verifies that
// CIL trees are generated for multiple connectors.
func TestAnalyzeConnectorsScope_CILTreeGeneration_MultipleConnectors(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with required definitions
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	countryConcept := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConcept)
	defs.Measures["Sales"] = analysis.AnalysisMeasureDefinition{Name: "Sales", TypeName: "Number"}
	defs.Measures["Inventory"] = analysis.AnalysisMeasureDefinition{Name: "Inventory", TypeName: "Number"}
	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{Name: "Day", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "day"}}

	context := analysis.NewAnalysisContext()

	// First connector
	connector1 := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connector1.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesConnector",
		NameLocation:   symboltable.SourceLocation{Line: 1, Column: 1},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Sales",
	}

	// Second connector
	connector2 := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connector2.Data = &symboltable.ConnectorDeclaration{
		Name:           "InventoryConnector",
		NameLocation:   symboltable.SourceLocation{Line: 5, Column: 1},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Inventory",
	}

	// Act: Analyze both connectors
	err1 := analysis.AnalyzeConnectorsScope(connector1, context, analyzer)
	err2 := analysis.AnalyzeConnectorsScope(connector2, context, analyzer)

	// Assert: No errors
	require.NoError(t, err1)
	require.NoError(t, err2)

	// Assert: Both CIL trees exist
	resultDefs := analyzer.GetAnalysisDefinitions()
	assert.Len(t, resultDefs.CILTrees, 2, "Should have 2 CIL trees")

	// Verify first CIL tree
	cilNode1, exists1 := resultDefs.CILTrees["SalesConnector"]
	require.True(t, exists1)
	connectorNode1, ok1 := cilNode1.(*cil.CILConnectorNode)
	require.True(t, ok1)
	assert.Equal(t, "SalesConnector", connectorNode1.ConnectorName)

	// Verify second CIL tree
	cilNode2, exists2 := resultDefs.CILTrees["InventoryConnector"]
	require.True(t, exists2)
	connectorNode2, ok2 := cilNode2.(*cil.CILConnectorNode)
	require.True(t, ok2)
	assert.Equal(t, "InventoryConnector", connectorNode2.ConnectorName)
}

// ============================================================================
// Error Cases - Validation Failures
// ============================================================================

// TestAnalyzeConnectorsScope_UndefinedDimension_ReturnsError verifies that
// an error is returned when a dimension references an undefined concept.
func TestAnalyzeConnectorsScope_UndefinedDimension_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer WITHOUT the dimension concept
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["Amount"] = analysis.AnalysisMeasureDefinition{Name: "Amount"}

	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{Name: "Day", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "day"}}

	connectorScope := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 5},
		DimensionNames: []string{"UndefinedConcept"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Amount",
	}

	context := analysis.NewAnalysisContext()

	// Act: Analyze the connector
	err := analysis.AnalyzeConnectorsScope(connectorScope, context, analyzer)

	// Assert: Error is returned
	require.Error(t, err, "Should return error for undefined dimension")
	assert.Contains(t, err.Error(), "SalesData")
	assert.Contains(t, err.Error(), "UndefinedConcept")
	assert.Contains(t, err.Error(), "undefined concept dimension")
}

// TestAnalyzeConnectorsScope_UndefinedCelltype_ReturnsError verifies that
// an error is returned when the celltype references an undefined measure.
func TestAnalyzeConnectorsScope_UndefinedCelltype_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with dimension but WITHOUT the celltype measure
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	countryConcept := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConcept)

	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{Name: "Day", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "day"}}

	connectorScope := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 5},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "UndefinedMeasure",
	}

	context := analysis.NewAnalysisContext()

	// Act: Analyze the connector
	err := analysis.AnalyzeConnectorsScope(connectorScope, context, analyzer)

	// Assert: Error is returned
	require.Error(t, err, "Should return error for undefined celltype")
	assert.Contains(t, err.Error(), "SalesData")
	assert.Contains(t, err.Error(), "UndefinedMeasure")
	assert.Contains(t, err.Error(), "undefined celltype measure")
}

// TestAnalyzeConnectorsScope_DuplicateConnector_ReturnsError verifies that
// an error is returned when a duplicate connector name is detected.
func TestAnalyzeConnectorsScope_DuplicateConnector_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with required definitions
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	countryConcept := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConcept)
	defs.Measures["Amount"] = analysis.AnalysisMeasureDefinition{Name: "Amount", TypeName: "Number"}
	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{Name: "Day", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "day"}}

	context := analysis.NewAnalysisContext()

	// First connector (valid)
	connector1 := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connector1.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 5, Column: 5},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Amount",
	}

	// Second connector with same name (duplicate)
	connector2 := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connector2.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 5},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Amount",
	}

	// Act: Analyze first connector (should succeed)
	err1 := analysis.AnalyzeConnectorsScope(connector1, context, analyzer)
	require.NoError(t, err1, "First connector should succeed")

	// Act: Analyze second connector (should fail as duplicate)
	err2 := analysis.AnalyzeConnectorsScope(connector2, context, analyzer)

	// Assert: Duplicate error is returned
	require.Error(t, err2, "Should return error for duplicate connector")
	assert.Contains(t, err2.Error(), "SalesData")
	assert.Contains(t, err2.Error(), "duplicate connector definition")
}

// TestAnalyzeConnectorsScope_MultipleUndefinedDimensions_ReturnsError verifies that
// an error is returned for the first undefined dimension encountered.
func TestAnalyzeConnectorsScope_MultipleUndefinedDimensions_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with only one of the dimensions defined
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	countryConcept := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConcept)
	defs.Measures["Amount"] = analysis.AnalysisMeasureDefinition{Name: "Amount", TypeName: "Number"}
	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{Name: "Day", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "day"}}

	connectorScope := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 5},
		DimensionNames: []string{"Country", "UndefinedProduct", "AlsoUndefinedDate"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Amount",
	}

	context := analysis.NewAnalysisContext()

	// Act: Analyze the connector
	err := analysis.AnalyzeConnectorsScope(connectorScope, context, analyzer)

	// Assert: Error for the first undefined dimension
	require.Error(t, err, "Should return error for undefined dimension")
	assert.Contains(t, err.Error(), "UndefinedProduct")
}

// ============================================================================
// Definitions Independence Tests
// ============================================================================

// TestAnalyzeConnectorsScope_IgnoresDefinitionsForRoleInference verifies that
// AnalyzeConnectorsScope resolves connector roles exclusively from the structured
// ConnectorDeclaration in ConnectorScope.Data, regardless of stale, scrambled, or
// incomplete generic Definitions entries left by the parser.
func TestAnalyzeConnectorsScope_IgnoresDefinitionsForRoleInference(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		definitions []symboltable.SymbolDefinition
	}{
		{
			name: "identifiers in scrambled wrong order",
			definitions: []symboltable.SymbolDefinition{
				connectorSymbolDefinition("Revenue", symboltable.Identifier),
				connectorSymbolDefinition("UnrelatedGrain", symboltable.GrainDefinition),
				connectorSymbolDefinition("Product", symboltable.Identifier),
				connectorSymbolDefinition("SalesData", symboltable.Identifier),
				connectorSymbolDefinition("Country", symboltable.Identifier),
			},
		},
		{
			name: "missing identifiers",
			definitions: []symboltable.SymbolDefinition{
				connectorSymbolDefinition("SalesData", symboltable.Identifier),
			},
		},
		{
			name: "extra unrelated identifiers",
			definitions: []symboltable.SymbolDefinition{
				connectorSymbolDefinition("SalesData", symboltable.Identifier),
				connectorSymbolDefinition("Country", symboltable.Identifier),
				connectorSymbolDefinition("Product", symboltable.Identifier),
				connectorSymbolDefinition("Revenue", symboltable.Identifier),
				connectorSymbolDefinition("LegacyExtraDimension", symboltable.Identifier),
				connectorSymbolDefinition("LegacyExtraCellType", symboltable.Identifier),
				connectorSymbolDefinition("LegacyExtraGrain", symboltable.GrainDefinition),
			},
		},
		{
			name:        "no definitions at all",
			definitions: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			analyzer := analysis.NewCPLAnalyzer()
			defs := analyzer.GetAnalysisDefinitions()
			countryConcept := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
			productConcept := analysis.AnalysisConceptDefinition{ConceptName: "Product"}
			analyzer.AddNewConcept(&countryConcept)
			analyzer.AddNewConcept(&productConcept)
			defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{Name: "Revenue", TypeName: "Number"}
			defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{Name: "Day", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "day"}}

			connectorScope := symboltable.NewSymbolTable(symboltable.ConnectorScope)
			connectorScope.Definitions = tc.definitions
			connectorScope.Data = &symboltable.ConnectorDeclaration{
				Name:           "SalesData",
				NameLocation:   symboltable.SourceLocation{Line: 10, Column: 5},
				DimensionNames: []string{"Country", "Product"},
				GrainNames:     []string{"Day"},
				CellTypeName:   "Revenue",
			}

			err := analysis.AnalyzeConnectorsScope(connectorScope, analysis.NewAnalysisContext(), analyzer)

			require.NoError(t, err, "AnalyzeConnectorsScope should ignore ConnectorScope.Definitions for connector role inference")
			resultDefs := analyzer.GetAnalysisDefinitions()
			connectorDef, exists := resultDefs.Connectors["SalesData"]
			require.True(t, exists, "SalesData connector should be created from structured ConnectorDeclaration data")
			require.Len(t, connectorDef.Dimensions, 2, "dimensions should come from ConnectorDeclaration.DimensionNames")
			assert.Equal(t, "Country", connectorDef.Dimensions[0].ConceptName)
			assert.Equal(t, "Product", connectorDef.Dimensions[1].ConceptName)
			require.Len(t, connectorDef.Grain, 1, "grain should come from ConnectorDeclaration.GrainNames")
			assert.Equal(t, "Day", connectorDef.Grain[0].Name)
			require.NotNil(t, connectorDef.CellType, "celltype should come from ConnectorDeclaration.CellTypeName")
			assert.Equal(t, "Revenue", connectorDef.CellType.Name)
		})
	}
}

// TestAnalyzeConnectorsScope_MissingStructuredData_ReturnsError verifies that a
// ConnectorScope with no structured declaration data returns an error.
func TestAnalyzeConnectorsScope_MissingStructuredData_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer and connector scope with nil Data
	analyzer := analysis.NewCPLAnalyzer()
	connectorScope := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	// Data is nil; Definitions content is irrelevant to structured analysis

	context := analysis.NewAnalysisContext()

	// Act: Analyze the connector scope with missing structured data
	err := analysis.AnalyzeConnectorsScope(connectorScope, context, analyzer)

	// Assert: Error because structured declaration data is missing
	require.Error(t, err, "Should return error for missing structured declaration data")
	assert.Contains(t, err.Error(), "connector scope has no structured declaration data")
	resultDefs := analyzer.GetAnalysisDefinitions()
	assert.Empty(t, resultDefs.Connectors, "Should have no connectors")
	assert.Empty(t, resultDefs.CILTrees, "Should have no CIL trees")
}

// TestAnalyzeConnectorsScope_MissingGrain_ReturnsError verifies that a
// connector without a grain measure returns an appropriate error.
func TestAnalyzeConnectorsScope_MissingGrain_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with concepts and measures
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	countryConcept := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConcept)
	defs.Measures["Amount"] = analysis.AnalysisMeasureDefinition{Name: "Amount", TypeName: "Number"}

	connectorScope := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 5},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{},
		CellTypeName:   "Amount",
	}

	context := analysis.NewAnalysisContext()

	// Act: Analyze the connector
	err := analysis.AnalyzeConnectorsScope(connectorScope, context, analyzer)

	// Assert: Error is returned (no grain measure)
	require.Error(t, err, "Should return error for connector without grain")
	assert.Contains(t, err.Error(), "SalesData")
	assert.Contains(t, err.Error(), "has no grain measure")
}

// TestAnalyzeConnectorsScope_SingleDefinition_ReturnsError verifies that a
// connector with a name but no celltype returns an appropriate error.
func TestAnalyzeConnectorsScope_SingleDefinition_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer
	analyzer := analysis.NewCPLAnalyzer()
	analyzer.GetAnalysisDefinitions().Measures["Day"] = analysis.AnalysisMeasureDefinition{Name: "Day", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "day"}}

	connectorScope := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "IncompleteConnector",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 5},
		DimensionNames: []string{},
		GrainNames:     []string{"Day"},
		CellTypeName:   "",
	}

	context := analysis.NewAnalysisContext()

	// Act: Analyze the connector
	err := analysis.AnalyzeConnectorsScope(connectorScope, context, analyzer)

	// Assert: Error is returned (no celltype measure)
	require.Error(t, err, "Should return error for incomplete connector definition")
	assert.Contains(t, err.Error(), "IncompleteConnector")
	assert.Contains(t, err.Error(), "has no celltype measure")
}

// ============================================================================
// Multiple Grains Resolution Tests
// ============================================================================

// TestAnalyzeConnectorsScope_MultipleGrains_Resolved verifies that a connector
// with multiple grain measures resolves all of them into the Grain slice.
func TestAnalyzeConnectorsScope_MultipleGrains_Resolved(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with multi-grain measures
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()

	defs.Measures["Year"] = analysis.AnalysisMeasureDefinition{
		Name:     "Year",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "year"},
	}
	defs.Measures["Month"] = analysis.AnalysisMeasureDefinition{
		Name:     "Month",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "month"},
	}
	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Dollar"},
	}

	connectorScope := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "TimeSeriesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 5},
		DimensionNames: []string{},
		GrainNames:     []string{"Year", "Month", "Day"},
		CellTypeName:   "Revenue",
	}

	context := analysis.NewAnalysisContext()

	// Act: Analyze the connector
	err := analysis.AnalyzeConnectorsScope(connectorScope, context, analyzer)

	// Assert: No error occurred
	require.NoError(t, err, "Should not return error for valid multi-grain connector")

	// Assert: Connector was added with all three grains resolved
	resultDefs := analyzer.GetAnalysisDefinitions()
	connectorDef, exists := resultDefs.Connectors["TimeSeriesData"]
	require.True(t, exists, "TimeSeriesData connector should exist in map")

	require.Len(t, connectorDef.Grain, 3, "Grain should have 3 resolved measures")
	assert.Equal(t, "Year", connectorDef.Grain[0].Name, "First grain should be Year")
	assert.Equal(t, "Month", connectorDef.Grain[1].Name, "Second grain should be Month")
	assert.Equal(t, "Day", connectorDef.Grain[2].Name, "Third grain should be Day")
}

// TestAnalyzeConnectorsScope_MultipleGrains_OrderPreserved verifies that the
// order of grains in the resulting Grain slice matches the declaration order.
func TestAnalyzeConnectorsScope_MultipleGrains_OrderPreserved(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with multi-grain measures
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()

	defs.Measures["Year"] = analysis.AnalysisMeasureDefinition{
		Name:     "Year",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "year"},
	}
	defs.Measures["Month"] = analysis.AnalysisMeasureDefinition{
		Name:     "Month",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "month"},
	}
	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Dollar"},
	}

	connectorScope := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "TimeSeriesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 5},
		DimensionNames: []string{},
		GrainNames:     []string{"Day", "Month", "Year"},
		CellTypeName:   "Revenue",
	}

	context := analysis.NewAnalysisContext()

	// Act: Analyze the connector
	err := analysis.AnalyzeConnectorsScope(connectorScope, context, analyzer)

	// Assert: No error occurred
	require.NoError(t, err, "Should not return error for valid multi-grain connector")

	// Assert: Order is preserved as declared
	resultDefs := analyzer.GetAnalysisDefinitions()
	connectorDef, exists := resultDefs.Connectors["TimeSeriesData"]
	require.True(t, exists, "TimeSeriesData connector should exist in map")

	require.Len(t, connectorDef.Grain, 3, "Grain should have 3 resolved measures")
	assert.Equal(t, "Day", connectorDef.Grain[0].Name, "First grain should match declaration order")
	assert.Equal(t, "Month", connectorDef.Grain[1].Name, "Second grain should match declaration order")
	assert.Equal(t, "Year", connectorDef.Grain[2].Name, "Third grain should match declaration order")
}

// TestAnalyzeConnectorsScope_MultipleGrains_UndefinedSecondGrain_ReturnsError
// verifies that an error is returned when the second grain in a multi-grain
// declaration references an undefined measure.
func TestAnalyzeConnectorsScope_MultipleGrains_UndefinedSecondGrain_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create analyzer with only some grain measures defined
	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()

	defs.Measures["Year"] = analysis.AnalysisMeasureDefinition{
		Name:     "Year",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "year"},
	}
	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Dollar"},
	}

	connectorScope := symboltable.NewSymbolTable(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "TimeSeriesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 5},
		DimensionNames: []string{},
		GrainNames:     []string{"Year", "UndefinedGrain", "Day"},
		CellTypeName:   "Revenue",
	}

	context := analysis.NewAnalysisContext()

	// Act: Analyze the connector
	err := analysis.AnalyzeConnectorsScope(connectorScope, context, analyzer)

	// Assert: Error is returned mentioning the undefined grain measure
	require.Error(t, err, "Should return error when second grain is undefined")
	assert.Contains(t, err.Error(), "UndefinedGrain", "Error should mention the undefined grain measure name")
	assert.Contains(t, err.Error(), "undefined grain measure", "Error should mention undefined grain measure")
}

func connectorSymbolDefinition(name string, defType symboltable.DefinitionType) symboltable.SymbolDefinition {
	return symboltable.SymbolDefinition{
		Name:           strPtr(name),
		DefinitionType: defType,
		Location:       symboltable.SourceLocation{Line: 1, Column: 1},
	}
}

// strPtr is a helper function to get a pointer to a string.
func strPtr(s string) *string {
	return &s
}
