// Package analysis_test contains integration tests for connector analysis.
//
// These tests verify that the connector handler is correctly registered and invoked
// during the symbol table walking process. They test the full integration from
// parsing to analysis, ensuring connectors are properly analyzed when calling
// the top-level Analyze function.
//
// Reference: ERO004 Connectors Analysis Phase - Task 5
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseAndAnalyzeCPL parses a CPL input string and returns analysis definitions.
func parseAndAnalyzeCPL(input string) (*analysis.AnalysisDefinitions, error) {
	doc, err := parser.Parse(input)
	if err != nil {
		return nil, err
	}
	return analysis.Analyze(doc)
}

// ============================================================================
// Integration Tests - Handler Registration
// ============================================================================

// TestAnalyze_ConnectorHandler_Registered verifies that the connector handler
// is correctly registered and invoked during the Analyze function.
func TestAnalyze_ConnectorHandler_Registered(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with concepts, measures, and connectors
	input := `concepts
	Country = [Values[USA, UK, Germany]]
	Product = [Values[Widget, Gadget]]

measures
	Revenue = Number(unit=Dollar)
	Day = Interval(unit=Day)

connectors
	SalesData = dimensions [Country, Product] grain Day celltype Revenue`

	// Act: Parse and analyze
	defs, err := parseAndAnalyzeCPL(input)

	// Assert: No error and connector was analyzed
	require.NoError(t, err, "Should not return error for valid CPL with connectors")
	require.NotNil(t, defs, "AnalysisDefinitions should not be nil")

	// Verify connector exists
	require.Contains(t, defs.Connectors, "SalesData", "SalesData connector should be analyzed")
	connector := defs.Connectors["SalesData"]
	assert.Equal(t, "SalesData", connector.Name)
}

// TestAnalyze_ConnectorWithMultipleConnectors_AllAnalyzed verifies that
// multiple connectors are all analyzed correctly.
func TestAnalyze_ConnectorWithMultipleConnectors_AllAnalyzed(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with multiple connectors
	input := `concepts
	Country = [Values[USA, UK]]
	Product = [Values[Widget, Gadget]]

measures
	Revenue = Number(unit=Dollar)
	Quantity = Number(unit=Unit)
	Day = Interval(unit=Day)

connectors
	SalesData = dimensions [Country, Product] grain Day celltype Revenue
	Inventory = dimensions [Country, Product] grain Day celltype Quantity`

	// Act: Parse and analyze
	defs, err := parseAndAnalyzeCPL(input)

	// Assert: Both connectors were analyzed
	require.NoError(t, err, "Should not return error for valid CPL with multiple connectors")
	require.Len(t, defs.Connectors, 2, "Should have 2 connectors")

	// Verify first connector
	require.Contains(t, defs.Connectors, "SalesData")
	salesConn := defs.Connectors["SalesData"]
	assert.Equal(t, "SalesData", salesConn.Name)
	assert.Len(t, salesConn.Dimensions, 2)
	assert.Equal(t, "Revenue", salesConn.CellType.Name)

	// Verify second connector
	require.Contains(t, defs.Connectors, "Inventory")
	invConn := defs.Connectors["Inventory"]
	assert.Equal(t, "Inventory", invConn.Name)
	assert.Len(t, invConn.Dimensions, 2)
	assert.Equal(t, "Quantity", invConn.CellType.Name)
}

// TestAnalyze_ConnectorWithDimensions_ResolvedCorrectly verifies that
// dimension references are correctly resolved to concept definitions.
func TestAnalyze_ConnectorWithDimensions_ResolvedCorrectly(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with connector dimensions
	input := `concepts
	Country = [Values[USA, UK, Germany]]
	Product = [Values[Widget, Gadget, Thingama]]
	Date = [Values[Jan2024, Feb2024]]

measures
	Amount = Number(unit=Dollar)
	Day = Interval(unit=Day)

connectors
	SalesData = dimensions [Country, Product, Date] grain Day celltype Amount`

	// Act: Parse and analyze
	defs, err := parseAndAnalyzeCPL(input)

	// Assert: Dimensions are correctly resolved
	require.NoError(t, err)
	require.Contains(t, defs.Connectors, "SalesData")

	connector := defs.Connectors["SalesData"]
	require.Len(t, connector.Dimensions, 3, "Should have 3 dimensions")

	// Verify each dimension points to the correct concept
	assert.Equal(t, "Country", connector.Dimensions[0].ConceptName)
	assert.Equal(t, "Product", connector.Dimensions[1].ConceptName)
	assert.Equal(t, "Date", connector.Dimensions[2].ConceptName)
}

// TestAnalyze_ConnectorCILTree_Generated verifies that a CIL tree is generated
// for each connector during analysis.
func TestAnalyze_ConnectorCILTree_Generated(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with a connector
	input := `concepts
	Country = [Values[USA, UK]]

measures
	Revenue = Number(unit=Dollar)
	Day = Interval(unit=Day)

connectors
	SalesData = dimensions [Country] grain Day celltype Revenue`

	// Act: Parse and analyze
	defs, err := parseAndAnalyzeCPL(input)

	// Assert: CIL tree was generated
	require.NoError(t, err)
	require.Contains(t, defs.CILTrees, "SalesData", "CIL tree should exist for SalesData")

	cilNode := defs.CILTrees["SalesData"]
	require.NotNil(t, cilNode)

	// Type assert to verify it's a connector reference with correct name
	connectorNode, ok := cilNode.(*cil.CILConnectorNode)
	require.True(t, ok, "Should be CILConnectorNode")
	assert.Equal(t, "SalesData", connectorNode.ConnectorName)
}

// TestAnalyze_ConnectorEmptyDimensions_ValidConnector verifies that a connector
// with empty dimensions list is correctly analyzed.
// Note: The grammar requires at least one dimension, so this tests the minimum case.
func TestAnalyze_ConnectorEmptyDimensions_ValidConnector(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with a connector with one dimension
	input := `concepts
	Country = [Values[USA, UK]]

measures
	TotalRevenue = Number(unit=Dollar)
	Day = Interval(unit=Day)

connectors
	Summary = dimensions [Country] grain Day celltype TotalRevenue`

	// Act: Parse and analyze
	defs, err := parseAndAnalyzeCPL(input)

	// Assert: Connector was analyzed with one dimension
	require.NoError(t, err)
	require.Contains(t, defs.Connectors, "Summary")

	connector := defs.Connectors["Summary"]
	assert.Equal(t, "Summary", connector.Name)
	assert.Len(t, connector.Dimensions, 1, "Should have 1 dimension")
	assert.Equal(t, "TotalRevenue", connector.CellType.Name)
}

// ============================================================================
// Error Propagation Tests
// ============================================================================

// TestAnalyze_ConnectorUndefinedDimension_ReturnsError verifies that analysis
// fails when a connector references an undefined concept dimension.
func TestAnalyze_ConnectorUndefinedDimension_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with a connector referencing undefined concept
	input := `concepts
	Country = [Values[USA, UK]]

measures
	Revenue = Number(unit=Dollar)
	Day = Interval(unit=Day)

connectors
	SalesData = dimensions [Country, UndefinedConcept] grain Day celltype Revenue`

	// Act: Parse and analyze
	_, err := parseAndAnalyzeCPL(input)

	// Assert: Error is returned mentioning undefined dimension
	require.Error(t, err, "Should return error for undefined dimension")
	assert.Contains(t, err.Error(), "undefined concept dimension")
	assert.Contains(t, err.Error(), "UndefinedConcept")
}

// TestAnalyze_ConnectorUndefinedCelltype_ReturnsError verifies that analysis
// fails when a connector references an undefined measure celltype.
func TestAnalyze_ConnectorUndefinedCelltype_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with a connector referencing undefined measure
	input := `concepts
	Country = [Values[USA, UK]]

measures
	Revenue = Number(unit=Dollar)
	Day = Interval(unit=Day)

connectors
	SalesData = dimensions [Country] grain Day celltype UndefinedMeasure`

	// Act: Parse and analyze
	_, err := parseAndAnalyzeCPL(input)

	// Assert: Error is returned mentioning undefined celltype
	require.Error(t, err, "Should return error for undefined celltype")
	assert.Contains(t, err.Error(), "undefined celltype measure")
	assert.Contains(t, err.Error(), "UndefinedMeasure")
}

// TestAnalyze_ConnectorDuplicateName_ReturnsError verifies that analysis
// fails when duplicate connector names are defined.
func TestAnalyze_ConnectorDuplicateName_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with duplicate connector names
	input := `concepts
	Country = [Values[USA, UK]]

measures
	Revenue = Number(unit=Dollar)
	Day = Interval(unit=Day)

connectors
	SalesData = dimensions [Country] grain Day celltype Revenue
	SalesData = dimensions [Country] grain Day celltype Revenue`

	// Act: Parse and analyze
	_, err := parseAndAnalyzeCPL(input)

	// Assert: Error is returned mentioning duplicate
	require.Error(t, err, "Should return error for duplicate connector")
	assert.Contains(t, err.Error(), "duplicate connector definition")
	assert.Contains(t, err.Error(), "SalesData")
}

// ============================================================================
// Full Integration Tests - All Sections
// ============================================================================

// TestAnalyze_FullProgram_WithAllSections verifies that a complete CPL program
// with concepts, types, units, measures, and connectors is correctly analyzed.
func TestAnalyze_FullProgram_WithAllSections(t *testing.T) {
	t.Parallel()

	// Arrange: Create a complete CPL program
	input := `concepts
	Country = [Values[USA, UK, Germany, France]]
	Product = [Values[Widget, Gadget, Thing]]

measures
	Revenue = Number(unit=Dollar)
	Quantity = Number(unit=Unit)
	Day = Interval(unit=Day)

connectors
	Sales = dimensions [Country, Product] grain Day celltype Revenue
	Inventory = dimensions [Country, Product] grain Day celltype Quantity`

	// Act: Parse and analyze
	defs, err := parseAndAnalyzeCPL(input)

	// Assert: All sections were analyzed correctly
	require.NoError(t, err)
	require.NotNil(t, defs)

	// Verify concepts. We don't check the exact total because standard time
	// concepts (Month, DayOfWeek, Hour, etc.) are injected by the analyzer.
	assert.Contains(t, defs.Concepts, "Country")
	assert.Contains(t, defs.Concepts, "Product")

	// Verify measures
	assert.Len(t, defs.Measures, 3, "Should have 3 measures")
	assert.Contains(t, defs.Measures, "Revenue")
	assert.Contains(t, defs.Measures, "Quantity")
	assert.Contains(t, defs.Measures, "Day")

	// Verify connectors
	assert.Len(t, defs.Connectors, 2, "Should have 2 connectors")
	assert.Contains(t, defs.Connectors, "Sales")
	assert.Contains(t, defs.Connectors, "Inventory")

	// Verify CIL trees
	assert.Len(t, defs.CILTrees, 2, "Should have 2 CIL trees")
	assert.Contains(t, defs.CILTrees, "Sales")
	assert.Contains(t, defs.CILTrees, "Inventory")
}

// TestAnalyzeConnectorWithGrainEndToEnd verifies the full parser-to-analyzer
// pipeline for a connector with a grain clause.
func TestAnalyzeConnectorWithGrainEndToEnd(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program where the parser captures a structured
	// ConnectorDeclaration and the analyzer resolves its GrainNames to an Interval
	// measure.
	input := `concepts
	Country = [Values[USA, UK]]

measures
	Revenue = Number(unit=Dollar)
	Month = Interval(unit=Month)

connectors
	SalesData = dimensions [Country] grain Month celltype Revenue`

	// Act: Parse and analyze through the full pipeline.
	defs, err := parseAndAnalyzeCPL(input)

	// Assert: Grain is resolved from the structured ConnectorDeclaration field
	// through analyzer definitions.
	require.NoError(t, err)
	require.Contains(t, defs.Connectors, "SalesData")
	connector := defs.Connectors["SalesData"]
	require.Len(t, connector.Grain, 1, "Grain should be resolved for connector with grain clause")
	assert.Equal(t, "Month", connector.Grain[0].Name)
	assert.Equal(t, "Interval", connector.Grain[0].TypeName)
}

// TestAnalyze_DeferredGrainValidation_BuiltInInterval_Accepts verifies that the
// full parseAndAnalyzeCPL pipeline accepts an explicit connector grain measure
// whose type is the built-in Interval type.
func TestAnalyze_DeferredGrainValidation_BuiltInInterval_Accepts(t *testing.T) {
	t.Parallel()

	input := `concepts
	Country = [Values[USA, UK]]

measures
	Revenue = Number(unit=Dollar)
	Day = Interval(unit=Day)

connectors
	SalesData = dimensions [Country] grain Day celltype Revenue`

	defs, err := parseAndAnalyzeCPL(input)

	require.NoError(t, err, "built-in Interval grain should analyze successfully")
	require.NotNil(t, defs, "analysis definitions should be returned")
	require.Contains(t, defs.Connectors, "SalesData", "SalesData connector should exist")
	connector := defs.Connectors["SalesData"]
	require.Len(t, connector.Grain, 1, "grain measure should be resolved")
	assert.Equal(t, "Day", connector.Grain[0].Name)
	assert.Equal(t, "Interval", connector.Grain[0].TypeName)
}

// TestAnalyze_DeferredGrainValidation_DerivedInterval_Accepts verifies that the
// full parseAndAnalyzeCPL pipeline accepts an explicit connector grain measure
// whose declared type ultimately derives from Interval.
func TestAnalyze_DeferredGrainValidation_DerivedInterval_Accepts(t *testing.T) {
	t.Parallel()

	input := `concepts
	Country = [Values[USA, UK]]
	Hour = [Int[0, 23, 1]]
	DurationUnit = [Values[Day]]
	Dollar = [Values[USD]]

units
	Day = DurationUnit
	Dollar = Dollar

types
	DayStart = Instant(hour=1)
	Daily = Interval(starts_at=DayStart, unit=Day)

measures
	Revenue = Number(unit=Dollar)
	DayGrain = Daily(unit=Day)

connectors
	SalesData = dimensions [Country] grain DayGrain celltype Revenue`

	defs, err := parseAndAnalyzeCPL(input)

	require.NoError(t, err, "Interval-derived grain should analyze successfully")
	require.NotNil(t, defs, "analysis definitions should be returned")
	require.Contains(t, defs.Connectors, "SalesData", "SalesData connector should exist")
	connector := defs.Connectors["SalesData"]
	require.Len(t, connector.Grain, 1, "grain measure should be resolved")
	assert.Equal(t, "DayGrain", connector.Grain[0].Name)
	assert.Equal(t, "Daily", connector.Grain[0].TypeName)
	require.Contains(t, defs.Types, "Daily", "derived Interval type should be analyzed")
	assert.Equal(t, []string{"Daily", "Interval"}, defs.Types["Daily"].DerivationPath)
}

// TestAnalyze_DeferredGrainValidation_BuiltInNumber_Rejects verifies that the
// full parseAndAnalyzeCPL pipeline rejects an explicit connector grain measure
// whose type is the built-in Number type.
func TestAnalyze_DeferredGrainValidation_BuiltInNumber_Rejects(t *testing.T) {
	t.Parallel()

	input := `concepts
	Country = [Values[USA, UK]]

measures
	Revenue = Number(unit=Dollar)
	Count = Number(unit=Unit)

connectors
	SalesData = dimensions [Country] grain Count celltype Revenue`

	_, err := parseAndAnalyzeCPL(input)

	require.Error(t, err, "built-in Number grain should be rejected")
	assert.Contains(t, err.Error(), "SalesData", "error should identify the connector")
	assert.Contains(t, err.Error(), "Count", "error should identify the grain measure")
	assert.Contains(t, err.Error(), "Number", "error should identify the invalid grain type")
}

// TestAnalyze_DeferredGrainValidation_DerivedNumber_Rejects verifies that the
// full parseAndAnalyzeCPL pipeline rejects an explicit connector grain measure
// whose declared type ultimately derives from Number.
func TestAnalyze_DeferredGrainValidation_DerivedNumber_Rejects(t *testing.T) {
	t.Parallel()

	input := `concepts
	Country = [Values[USA, UK]]

types
	Currency = Number()

measures
	Revenue = Number(unit=Dollar)
	MoneyGrain = Currency(unit=Dollar)

connectors
	SalesData = dimensions [Country] grain MoneyGrain celltype Revenue`

	_, err := parseAndAnalyzeCPL(input)

	require.Error(t, err, "Number-derived grain should be rejected")
	assert.Contains(t, err.Error(), "SalesData", "error should identify the connector")
	assert.Contains(t, err.Error(), "MoneyGrain", "error should identify the grain measure")
	assert.Contains(t, err.Error(), "Currency", "error should identify the invalid declared grain type")
}

// TestAnalyze_ConnectorSourceLocation_Captured verifies that source location
// information is correctly captured for connectors.
func TestAnalyze_ConnectorSourceLocation_Captured(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with a connector
	input := `concepts
	Country = [Values[USA]]

measures
	Revenue = Number(unit=Dollar)
	Day = Interval(unit=Day)

connectors
	SalesData = dimensions [Country] grain Day celltype Revenue`

	// Act: Parse and analyze
	defs, err := parseAndAnalyzeCPL(input)

	// Assert: Source location was captured
	require.NoError(t, err)
	require.Contains(t, defs.Connectors, "SalesData")

	connector := defs.Connectors["SalesData"]
	assert.Greater(t, connector.SourceLocation.Line, 0, "Line should be set")
	assert.GreaterOrEqual(t, connector.SourceLocation.Column, 0, "Column should be set")
}

// ============================================================================
// WalkSymbolTableWithHandlers Integration Tests
// ============================================================================

// TestWalkSymbolTableWithHandlers_ConnectorScope_Invoked verifies that the
// connector handler is correctly invoked when walking a symbol table with a ConnectorScope.
// This test sets up the analyzer state directly (as would be done by previous scope handlers)
// and verifies that AnalyzeConnectorsScope processes the connector correctly.
func TestWalkSymbolTableWithHandlers_ConnectorScope_Invoked(t *testing.T) {
	t.Parallel()

	// Arrange: Build a symbol table tree with a connectors scope
	root := symboltable.NewGlobalSymbolTable()

	// Create connectors scope with a connector
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)
	connectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 1},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Revenue",
	}

	// Create analyzer with pre-populated concepts and measures
	// (simulating what would happen after ConceptScope and MeasureScope handlers run)
	analyzer := analysis.NewCPLAnalyzer()

	// Add the concept that the connector references
	countryConceptDef := analysis.AnalysisConceptDefinition{
		ConceptName: "Country",
	}
	analyzer.AddNewConcept(&countryConceptDef)

	// Add the measure that the connector references
	defs := analyzer.GetAnalysisDefinitions()
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

	// Register the connector handler for ConnectorScope (not ConnectorsScope)
	// ConnectorScope represents an individual connector, while ConnectorsScope is the parent container
	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
	}

	// Act: Walk the symbol table
	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	// Assert: No error and connector was processed
	require.NoError(t, err, "Walk should not return error")

	resultDefs := analyzer.GetAnalysisDefinitions()
	assert.Len(t, resultDefs.Connectors, 1, "Should have 1 connector")
	assert.Contains(t, resultDefs.Connectors, "SalesData")
	assert.Equal(t, "SalesData", resultDefs.Connectors["SalesData"].Name)
	assert.Len(t, resultDefs.Connectors["SalesData"].Dimensions, 1, "Should have 1 dimension")
	assert.Equal(t, "Revenue", resultDefs.Connectors["SalesData"].CellType.Name)
}

// ============================================================================
// Multiple Grains End-to-End Tests
// ============================================================================

// TestAnalyzeConnectorWithMultipleGrainsEndToEnd verifies the full
// parser-to-analyzer pipeline for a connector declaring multiple grains.
func TestAnalyzeConnectorWithMultipleGrainsEndToEnd(t *testing.T) {
	t.Parallel()

	// Arrange: Create a CPL program with a connector declaring three grains.
	input := `concepts
	Country = [Values[USA, UK]]

measures
	Revenue = Number(unit=Dollar)
	Year = Interval(unit=Year)
	Month = Interval(unit=Month)
	Day = Interval(unit=Day)

connectors
	SalesData = dimensions [Country] grain Year, Month, Day celltype Revenue`

	// Act: Parse and analyze through the full pipeline.
	defs, err := parseAndAnalyzeCPL(input)

	// Assert: No error and all grains are resolved in declaration order.
	require.NoError(t, err, "Should not return error for valid multi-grain connector")
	require.NotNil(t, defs, "AnalysisDefinitions should not be nil")
	require.Contains(t, defs.Connectors, "SalesData", "SalesData connector should be analyzed")

	connector := defs.Connectors["SalesData"]
	require.Len(t, connector.Grain, 3, "Grain should have 3 resolved measures")
	assert.Equal(t, "Year", connector.Grain[0].Name, "First grain should be Year")
	assert.Equal(t, "Interval", connector.Grain[0].TypeName, "First grain should have Interval type")
	assert.Equal(t, "Month", connector.Grain[1].Name, "Second grain should be Month")
	assert.Equal(t, "Interval", connector.Grain[1].TypeName, "Second grain should have Interval type")
	assert.Equal(t, "Day", connector.Grain[2].Name, "Third grain should be Day")
	assert.Equal(t, "Interval", connector.Grain[2].TypeName, "Third grain should have Interval type")
}
