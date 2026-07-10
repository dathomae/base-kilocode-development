// Package analysis_test contains unit tests for the connector analysis data structures.
//
// These tests verify that AnalysisConnectorDefinition correctly stores and
// exposes the analyzed grain as a measure reference. The grain represents the
// finest level of detail at which data in the connector is recorded, and is
// expressed as a reference to an AnalysisMeasureDefinition.
//
// Story: Story159 - CPL Analyzer Connectors Time
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// AnalysisConnectorDefinition Grain Field Tests
// ============================================================================

// TestAnalysisConnectorDefinition_WithGrain verifies that an AnalysisConnectorDefinition
// can be constructed with a Grain field set to a slice containing a pointer to an AnalysisMeasureDefinition.
//
// The grain is a measure reference that indicates the finest level of detail for
// the connector's data. It must be stored as a []*AnalysisMeasureDefinition so that
// the synthesis phase can access the measure's type and unit information directly.
func TestAnalysisConnectorDefinition_WithGrain(t *testing.T) {
	t.Parallel()

	// Arrange: Create a measure definition to use as the grain
	grainMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "DailyGrain",
		TypeName: "Instant",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}

	// Act: Construct an AnalysisConnectorDefinition with the Grain field set
	connector := analysis.AnalysisConnectorDefinition{
		Name:           "SalesData",
		Dimensions:     nil,
		CellType:       nil,
		Grain:          []*analysis.AnalysisMeasureDefinition{&grainMeasure},
		SourceLocation: symboltable.SourceLocation{Line: 1, Column: 1},
	}

	// Assert: The Grain field is set and contains the correct measure
	require.Len(t, connector.Grain, 1, "Grain field should have one element when explicitly set")
	assert.Equal(t, "DailyGrain", connector.Grain[0].Name, "Grain measure name should match")
	assert.Equal(t, "Instant", connector.Grain[0].TypeName, "Grain measure type name should match")
	assert.Equal(t, "day", connector.Grain[0].GetUnitName(), "Grain measure unit name should match")
}

// TestAnalysisConnectorDefinition_GrainIsSliceOfPointers verifies that
// the Grain field stores a slice of pointers to AnalysisMeasureDefinition values.
//
// This is important because the synthesis phase needs to access the original
// measure definitions (including any resolved type information) without copying.
func TestAnalysisConnectorDefinition_GrainIsSliceOfPointers(t *testing.T) {
	t.Parallel()

	// Arrange: Create a measure definition
	grainMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "MonthlyGrain",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "month"},
	}

	// Act: Construct connector with a slice containing a pointer to the measure
	connector := analysis.AnalysisConnectorDefinition{
		Name:  "Inventory",
		Grain: []*analysis.AnalysisMeasureDefinition{&grainMeasure},
	}

	// Assert: Modifying the original measure is reflected through the pointer
	require.Len(t, connector.Grain, 1)
	assert.Equal(t, "MonthlyGrain", connector.Grain[0].Name)

	// Verify the field type is []*AnalysisMeasureDefinition by indexing
	var measureRef *analysis.AnalysisMeasureDefinition = connector.Grain[0]
	require.NotNil(t, measureRef)
	assert.Equal(t, connector.Grain[0].Name, measureRef.Name)
}

// TestAnalysisConnectorDefinition_WithNilGrain verifies that an AnalysisConnectorDefinition
// can be constructed with a nil Grain field and behaves correctly.
//
// A nil grain indicates that no explicit grain has been specified for the connector.
// This is a valid state that the analysis and synthesis phases must handle gracefully.
func TestAnalysisConnectorDefinition_WithNilGrain(t *testing.T) {
	t.Parallel()

	// Arrange: Create a celltype measure
	cellTypeMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Dollar"},
	}

	// Act: Construct an AnalysisConnectorDefinition with Grain left nil
	connector := analysis.AnalysisConnectorDefinition{
		Name:           "SalesData",
		Dimensions:     nil,
		CellType:       &cellTypeMeasure,
		Grain:          nil,
		SourceLocation: symboltable.SourceLocation{Line: 5, Column: 1},
	}

	// Assert: The connector is valid even with a nil Grain
	assert.Equal(t, "SalesData", connector.Name)
	assert.Nil(t, connector.Grain, "Grain field should be nil when not set")
	assert.NotNil(t, connector.CellType, "CellType should still be set independently of Grain")
	assert.Equal(t, "Revenue", connector.CellType.Name)
}

// TestAnalysisConnectorDefinition_GrainFieldType verifies that the Grain field
// on AnalysisConnectorDefinition has the expected type []*AnalysisMeasureDefinition.
//
// This compile-time test ensures the field declaration matches the architectural
// requirement that grain is stored as a slice of measure references.
func TestAnalysisConnectorDefinition_GrainFieldType(t *testing.T) {
	t.Parallel()

	// This test verifies at compile time that the Grain field can be assigned
	// a value of type []*AnalysisMeasureDefinition and that the zero value is nil.
	var connector analysis.AnalysisConnectorDefinition

	// The zero value of a slice field should be nil
	assert.Nil(t, connector.Grain, "Zero-value Grain field should be nil")

	// Assignment of []*AnalysisMeasureDefinition should compile
	measure := analysis.AnalysisMeasureDefinition{Name: "TestGrain"}
	connector.Grain = []*analysis.AnalysisMeasureDefinition{&measure}
	assert.Len(t, connector.Grain, 1)
}

// TestAnalysisConnectorDefinition_GrainAndCellTypeIndependence verifies that
// the Grain and CellType fields are independent and can hold different measures.
//
// In CPL, the celltype defines the data type of values in the connector, while
// the grain defines the temporal (or other) granularity. These are semantically
// distinct and must be stored independently.
func TestAnalysisConnectorDefinition_GrainAndCellTypeIndependence(t *testing.T) {
	t.Parallel()

	// Arrange: Create distinct measures for celltype and grain
	cellTypeMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Dollar"},
	}
	grainMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "Daily",
		TypeName: "Instant",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}

	// Act: Construct connector with both CellType and Grain set
	connector := analysis.AnalysisConnectorDefinition{
		Name:     "SalesData",
		CellType: &cellTypeMeasure,
		Grain:    []*analysis.AnalysisMeasureDefinition{&grainMeasure},
	}

	// Assert: Both fields are independent
	require.NotNil(t, connector.CellType)
	require.Len(t, connector.Grain, 1)
	assert.Equal(t, "Revenue", connector.CellType.Name, "CellType should be Revenue")
	assert.Equal(t, "Daily", connector.Grain[0].Name, "Grain should be Daily")
	assert.NotEqual(t, connector.CellType.Name, connector.Grain[0].Name,
		"CellType and Grain should be different measures")
}

// ============================================================================
// AnalyzeConnectorsScope Grain Resolution Tests
// ============================================================================

// TestAnalyzeConnectorsScope_ValidGrain_Resolved verifies that when a
// ConnectorScope.Data structured ConnectorDeclaration contains GrainNames,
// AnalyzeConnectorsScope resolves the grain measure names against the analyzer's
// measures map and sets the Grain field on the resulting
// AnalysisConnectorDefinition.
//
// This test uses direct symbol table construction, setting up a ConnectorScope
// with structured declaration data for name, dimensions, grain, and celltype.
func TestAnalyzeConnectorsScope_ValidGrain_Resolved(t *testing.T) {
	t.Parallel()

	// Arrange: Build a symbol table tree with a connectors scope
	root := symboltable.NewGlobalSymbolTable()
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)
	connectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)

	// Set up structured declaration in Data (the analyzer no longer reconstructs from Definitions)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 1},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Revenue",
	}

	// Create analyzer with pre-populated concepts and measures
	analyzer := analysis.NewCPLAnalyzer()

	countryConceptDef := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConceptDef)

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

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
	}

	// Act: Walk the symbol table
	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	// Assert: No error and grain was resolved
	require.NoError(t, err, "Walk should not return error for valid grain")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Connectors, "SalesData")
	connector := resultDefs.Connectors["SalesData"]
	require.Len(t, connector.Grain, 1, "Grain should be resolved for connector with grain")
	assert.Equal(t, "Day", connector.Grain[0].Name, "Grain measure name should match")
	assert.Equal(t, "Interval", connector.Grain[0].TypeName, "Grain measure type should match")
	assert.Equal(t, "day", connector.Grain[0].GetUnitName(), "Grain measure unit should match")
}

// TestAnalyzeConnectorsScope_NoGrain_ReturnsError verifies that when a ConnectorScope.Data
// structured ConnectorDeclaration has empty GrainNames, AnalyzeConnectorsScope rejects
// the connector because grain is mandatory.
func TestAnalyzeConnectorsScope_NoGrain_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Build a symbol table tree with a connectors scope
	root := symboltable.NewGlobalSymbolTable()
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)
	connectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)

	// Set up structured declaration in Data with empty GrainNames
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 1},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{},
		CellTypeName:   "Revenue",
	}

	// Create analyzer with pre-populated concepts and measures
	analyzer := analysis.NewCPLAnalyzer()

	countryConceptDef := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConceptDef)

	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Dollar"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
	}

	// Act: Walk the symbol table
	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	// Assert: Missing grain is rejected because connector grain is mandatory
	require.Error(t, err, "Walk should return error when grain is absent")
	assert.Contains(t, err.Error(), "SalesData")
	assert.Contains(t, err.Error(), "has no grain measure")
}

// TestAnalyzeConnectorsScope_UndefinedGrain_ReturnsError verifies that when a
// ConnectorScope.Data structured ConnectorDeclaration has GrainNames referencing
// a measure that does not exist in the analyzer's measures map,
// AnalyzeConnectorsScope returns an appropriate error.
func TestAnalyzeConnectorsScope_UndefinedGrain_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange: Build a symbol table tree with a connectors scope
	root := symboltable.NewGlobalSymbolTable()
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)
	connectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)

	// Set up structured declaration in Data with an undefined grain measure
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 1},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"UndefinedGrain"},
		CellTypeName:   "Revenue",
	}

	// Create analyzer with pre-populated concepts and measures
	analyzer := analysis.NewCPLAnalyzer()

	countryConceptDef := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConceptDef)

	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Dollar"},
	}
	// Note: "UndefinedGrain" is intentionally NOT added to measures

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
	}

	// Act: Walk the symbol table
	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	// Assert: Error is returned mentioning the undefined grain measure
	require.Error(t, err, "Should return error for undefined grain measure")
	assert.Contains(t, err.Error(), "UndefinedGrain", "Error should mention the undefined grain measure name")
}

// ============================================================================
// AnalyzeConnectorsScope Grain Type Validation Tests
// ============================================================================

// TestAnalyzeConnectorsScope_GrainBuiltInInterval_Valid verifies that when a
// connector's grain measure has the built-in Interval type, the connector is
// created successfully.
func TestAnalyzeConnectorsScope_GrainBuiltInInterval_Valid(t *testing.T) {
	t.Parallel()

	// Arrange: Build a symbol table tree with a connectors scope
	root := symboltable.NewGlobalSymbolTable()
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)
	connectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)

	// Set up structured declaration in Data
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 1},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Revenue",
	}

	// Create analyzer with pre-populated concepts, measures, and types
	analyzer := analysis.NewCPLAnalyzer()

	countryConceptDef := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConceptDef)

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

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
	}

	// Act: Walk the symbol table
	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	// Assert: No error and grain was resolved
	require.NoError(t, err, "Walk should not return error for grain with built-in Interval type")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Connectors, "SalesData")
	connector := resultDefs.Connectors["SalesData"]
	require.Len(t, connector.Grain, 1, "Grain should be resolved for connector with Interval grain")
	assert.Equal(t, "Day", connector.Grain[0].Name, "Grain measure name should match")
	assert.Equal(t, "Interval", connector.Grain[0].TypeName, "Grain measure type should be Interval")
}

// TestAnalyzeConnectorsScope_GrainDerivedInterval_Valid verifies that when a
// connector's grain measure has a type that derives from Interval, the connector
// is created successfully.
func TestAnalyzeConnectorsScope_GrainDerivedInterval_Valid(t *testing.T) {
	t.Parallel()

	// Arrange: Build a symbol table tree with a connectors scope
	root := symboltable.NewGlobalSymbolTable()
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)
	connectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)

	// Set up structured declaration in Data
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 1},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"DailyGrain"},
		CellTypeName:   "Revenue",
	}

	// Create analyzer with pre-populated concepts, measures, and types
	analyzer := analysis.NewCPLAnalyzer()

	countryConceptDef := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConceptDef)

	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Dollar"},
	}
	defs.Measures["DailyGrain"] = analysis.AnalysisMeasureDefinition{
		Name:     "DailyGrain",
		TypeName: "Daily",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}

	// Add the "Daily" type derived from Interval
	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Daily",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.IntervalType),
		DerivationPath: []string{"Daily", "Interval"},
	})

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
	}

	// Act: Walk the symbol table
	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	// Assert: No error and grain was resolved
	require.NoError(t, err, "Walk should not return error for grain with Interval-derived type")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Connectors, "SalesData")
	connector := resultDefs.Connectors["SalesData"]
	require.Len(t, connector.Grain, 1, "Grain should be resolved for connector with derived Interval grain")
	assert.Equal(t, "DailyGrain", connector.Grain[0].Name, "Grain measure name should match")
	assert.Equal(t, "Daily", connector.Grain[0].TypeName, "Grain measure type should be Daily")
}

// TestAnalyzeConnectorsScope_GrainNumberType_Deferred verifies that when a
// connector's grain measure has the built-in Number type, AnalyzeConnectorsScope
// accepts and resolves the grain during the first pass, then the deferred
// post-derivation validation pass rejects it (Story172 Task 3).
func TestAnalyzeConnectorsScope_GrainNumberType_Deferred(t *testing.T) {
	t.Parallel()

	// Arrange: Build a symbol table tree with a connectors scope
	root := symboltable.NewGlobalSymbolTable()
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)
	connectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)

	// Set up structured declaration in Data
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 1},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"Count"},
		CellTypeName:   "Revenue",
	}

	// Create analyzer with pre-populated concepts and measures
	analyzer := analysis.NewCPLAnalyzer()

	countryConceptDef := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConceptDef)

	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Dollar"},
	}
	defs.Measures["Count"] = analysis.AnalysisMeasureDefinition{
		Name:     "Count",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Unit"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
	}

	// Act: Walk the symbol table
	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	// Assert: No error because grain type validation is deferred
	require.NoError(t, err, "Should not return error for grain measure with Number type in first pass")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Connectors, "SalesData")
	connector := resultDefs.Connectors["SalesData"]
	require.Len(t, connector.Grain, 1, "Grain should be resolved even when not Interval-derived")
	assert.Equal(t, "Count", connector.Grain[0].Name)
	assert.Equal(t, "Number", connector.Grain[0].TypeName)

	err = analyzer.ValidateConnectorGrainTypes()
	require.Error(t, err, "Deferred validation should reject Number grain type")
	assert.Contains(t, err.Error(), "SalesData")
	assert.Contains(t, err.Error(), "Count")
	assert.Contains(t, err.Error(), "Number")
}

// TestAnalyzeConnectorsScope_GrainNumberDerivedType_Deferred verifies that when a
// connector's grain measure has a type derived from Number, AnalyzeConnectorsScope
// accepts and resolves the grain during the first pass, then the deferred
// post-derivation validation pass rejects it (Story172 Task 3).
func TestAnalyzeConnectorsScope_GrainNumberDerivedType_Deferred(t *testing.T) {
	t.Parallel()

	// Arrange: Build a symbol table tree with a connectors scope
	root := symboltable.NewGlobalSymbolTable()
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)
	connectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)

	// Set up structured declaration in Data
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 1},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"CurrencyGrain"},
		CellTypeName:   "Revenue",
	}

	// Create analyzer with pre-populated concepts, measures, and types
	analyzer := analysis.NewCPLAnalyzer()

	countryConceptDef := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConceptDef)

	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Dollar"},
	}
	defs.Measures["CurrencyGrain"] = analysis.AnalysisMeasureDefinition{
		Name:     "CurrencyGrain",
		TypeName: "Currency",
		UnitExpression: &analysis.UnitReference{Name: "Dollar"},
	}

	// Add the "Currency" type derived from Number
	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Currency",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
		DerivationPath: []string{"Currency", "Number"},
	})

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
	}

	// Act: Walk the symbol table
	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	// Assert: No error because grain type validation is deferred
	require.NoError(t, err, "Should not return error for grain measure with Number-derived type in first pass")

	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Connectors, "SalesData")
	connector := resultDefs.Connectors["SalesData"]
	require.Len(t, connector.Grain, 1, "Grain should be resolved even when not Interval-derived")
	assert.Equal(t, "CurrencyGrain", connector.Grain[0].Name)
	assert.Equal(t, "Currency", connector.Grain[0].TypeName)

	err = analyzer.ValidateConnectorGrainTypes()
	require.Error(t, err, "Deferred validation should reject Number-derived grain type")
	assert.Contains(t, err.Error(), "SalesData")
	assert.Contains(t, err.Error(), "CurrencyGrain")
	assert.Contains(t, err.Error(), "Currency")
}

// TestAnalyzeConnectorsScope_NoGrain_FailsBeforeValidation verifies that when a
// ConnectorScope.Data structured ConnectorDeclaration has empty GrainNames,
// AnalyzeConnectorsScope rejects the connector before deferred grain type validation.
func TestAnalyzeConnectorsScope_NoGrain_FailsBeforeValidation(t *testing.T) {
	t.Parallel()

	// Arrange: Build a symbol table tree with a connectors scope
	root := symboltable.NewGlobalSymbolTable()
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)
	connectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)

	// Set up structured declaration in Data with empty GrainNames
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		NameLocation:   symboltable.SourceLocation{Line: 10, Column: 1},
		DimensionNames: []string{"Country"},
		GrainNames:     []string{},
		CellTypeName:   "Revenue",
	}

	// Create analyzer with pre-populated concepts and measures
	analyzer := analysis.NewCPLAnalyzer()

	countryConceptDef := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConceptDef)

	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "Dollar"},
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
	}

	// Act: Walk the symbol table
	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	// Assert: Missing grain is rejected before type validation can be skipped
	require.Error(t, err, "Walk should return error when grain is absent")
	assert.Contains(t, err.Error(), "SalesData")
	assert.Contains(t, err.Error(), "has no grain measure")
}

// ============================================================================
// ValidateConnectorGrainTypes Tests
// ============================================================================

// TestValidateConnectorGrainTypes_NilGrain_Skipped verifies that connectors with
// no explicit grain are ignored by the deferred grain type validation pass.
func TestValidateConnectorGrainTypes_NilGrain_Skipped(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Connectors["NoGrainConnector"] = analysis.AnalysisConnectorDefinition{
		Name:  "NoGrainConnector",
		Grain: nil,
	}

	err := analyzer.ValidateConnectorGrainTypes()

	require.NoError(t, err, "connector without explicit grain should be skipped")
}

// TestValidateConnectorGrainTypes_BuiltInInterval_Accepted verifies that a
// connector grain measure whose TypeName is the built-in Interval type is valid.
func TestValidateConnectorGrainTypes_BuiltInInterval_Accepted(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	grainMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "DailyGrain",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}
	defs.Connectors["IntervalConnector"] = analysis.AnalysisConnectorDefinition{
		Name:  "IntervalConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&grainMeasure},
	}

	err := analyzer.ValidateConnectorGrainTypes()

	require.NoError(t, err, "built-in Interval grain type should be accepted")
}

// TestValidateConnectorGrainTypes_DerivedInterval_Accepted verifies that a
// connector grain measure whose declared type derives from Interval is valid.
func TestValidateConnectorGrainTypes_DerivedInterval_Accepted(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "DailyInterval",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.IntervalType),
		DerivationPath: []string{"DailyInterval", "Interval"},
	})
	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "BusinessDayInterval",
		BaseType:       analysis.NewDeclaredTypeRef("DailyInterval"),
		DerivationPath: []string{"BusinessDayInterval", "DailyInterval", "Interval"},
	})

	defs := analyzer.GetAnalysisDefinitions()
	grainMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "BusinessDayGrain",
		TypeName: "BusinessDayInterval",
		UnitExpression: &analysis.UnitReference{Name: "business_day"},
	}
	defs.Connectors["DerivedIntervalConnector"] = analysis.AnalysisConnectorDefinition{
		Name:  "DerivedIntervalConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&grainMeasure},
	}

	err := analyzer.ValidateConnectorGrainTypes()

	require.NoError(t, err, "Interval-derived grain type should be accepted")
}

// TestValidateConnectorGrainTypes_BuiltInNumber_Rejected verifies that a
// connector grain measure whose TypeName is the built-in Number type is invalid.
func TestValidateConnectorGrainTypes_BuiltInNumber_Rejected(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	grainMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "CountGrain",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "count"},
	}
	defs.Connectors["NumberConnector"] = analysis.AnalysisConnectorDefinition{
		Name:  "NumberConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&grainMeasure},
	}

	err := analyzer.ValidateConnectorGrainTypes()

	require.Error(t, err, "built-in Number grain type should be rejected")
	assert.Contains(t, err.Error(), "NumberConnector", "error should identify the connector")
	assert.Contains(t, err.Error(), "CountGrain", "error should identify the grain measure")
	assert.Contains(t, err.Error(), "Number", "error should identify the invalid grain type")
}

// TestValidateConnectorGrainTypes_DerivedNumber_Rejected verifies that a
// connector grain measure whose declared type derives from Number is invalid.
func TestValidateConnectorGrainTypes_DerivedNumber_Rejected(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Currency",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
		DerivationPath: []string{"Currency", "Number"},
	})
	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "RetailCurrency",
		BaseType:       analysis.NewDeclaredTypeRef("Currency"),
		DerivationPath: []string{"RetailCurrency", "Currency", "Number"},
	})

	defs := analyzer.GetAnalysisDefinitions()
	grainMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "CurrencyGrain",
		TypeName: "RetailCurrency",
		UnitExpression: &analysis.UnitReference{Name: "dollar"},
	}
	defs.Connectors["DerivedNumberConnector"] = analysis.AnalysisConnectorDefinition{
		Name:  "DerivedNumberConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&grainMeasure},
	}

	err := analyzer.ValidateConnectorGrainTypes()

	require.Error(t, err, "Number-derived grain type should be rejected")
	assert.Contains(t, err.Error(), "DerivedNumberConnector", "error should identify the connector")
	assert.Contains(t, err.Error(), "CurrencyGrain", "error should identify the grain measure")
	assert.Contains(t, err.Error(), "RetailCurrency", "error should identify the invalid declared grain type")
}

// TestValidateConnectorGrainTypes_UndeclaredType_Rejected verifies that a
// connector grain measure whose TypeName is not a built-in or declared type is invalid.
func TestValidateConnectorGrainTypes_UndeclaredType_Rejected(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	grainMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "MysteryGrain",
		TypeName: "UndeclaredType",
		UnitExpression: &analysis.UnitReference{Name: "mystery_unit"},
	}
	defs.Connectors["UndeclaredTypeConnector"] = analysis.AnalysisConnectorDefinition{
		Name:  "UndeclaredTypeConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&grainMeasure},
	}

	err := analyzer.ValidateConnectorGrainTypes()

	require.Error(t, err, "undeclared grain type should be rejected")
	assert.Contains(t, err.Error(), "UndeclaredTypeConnector", "error should identify the connector")
	assert.Contains(t, err.Error(), "MysteryGrain", "error should identify the grain measure")
	assert.Contains(t, err.Error(), "UndeclaredType", "error should identify the undeclared grain type")
}

// ============================================================================
// AnalyzeConnectorsScope Structured ConnectorDeclaration Tests
// ============================================================================

// TestAnalyzeConnectorsScope_SameNameDimensionAndGrain_ResolvesBoth verifies
// that AnalyzeConnectorsScope consumes ConnectorScope.Data as a structured
// ConnectorDeclaration and preserves a dimension whose concept name matches the
// grain measure name.
func TestAnalyzeConnectorsScope_SameNameDimensionAndGrain_ResolvesBoth(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)
	connectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		DimensionNames: []string{"Country", "Day"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Revenue",
	}

	analyzer := analysis.NewCPLAnalyzer()
	countryConceptDef := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	dayConceptDef := analysis.AnalysisConceptDefinition{ConceptName: "Day"}
	analyzer.AddNewConcept(&countryConceptDef)
	analyzer.AddNewConcept(&dayConceptDef)

	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
	}
	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
	}

	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	require.NoError(t, err)
	resultDefs := analyzer.GetAnalysisDefinitions()
	require.Contains(t, resultDefs.Connectors, "SalesData")

	connector := resultDefs.Connectors["SalesData"]
	require.Len(t, connector.Dimensions, 2)
	assert.Equal(t, "Country", connector.Dimensions[0].ConceptName)
	assert.Equal(t, "Day", connector.Dimensions[1].ConceptName, "dimension concept named Day should be preserved even when grain is also Day")
	require.Len(t, connector.Grain, 1)
	assert.Equal(t, "Day", connector.Grain[0].Name)
	assert.Equal(t, "Interval", connector.Grain[0].TypeName)
	require.NotNil(t, connector.CellType)
	assert.Equal(t, "Revenue", connector.CellType.Name)
}

// TestAnalyzeConnectorsScope_StructuredRecordNoGrain_Error verifies that
// empty GrainNames in ConnectorScope.Data are rejected because grain is mandatory.
func TestAnalyzeConnectorsScope_StructuredRecordNoGrain_Error(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)
	connectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		DimensionNames: []string{"Country"},
		GrainNames:     []string{},
		CellTypeName:   "Revenue",
	}

	analyzer := analysis.NewCPLAnalyzer()
	countryConceptDef := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConceptDef)

	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
	}

	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SalesData")
	assert.Contains(t, err.Error(), "has no grain measure")
}

// TestAnalyzeConnectorsScope_StructuredRecordDuplicateName_Error verifies that
// duplicate connector names are detected when connector names come from
// ConnectorScope.Data structured records.
func TestAnalyzeConnectorsScope_StructuredRecordDuplicateName_Error(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)

	firstConnectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)
	firstConnectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Revenue",
	}

	secondConnectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)
	secondConnectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"Day"},
		CellTypeName:   "Revenue",
	}

	analyzer := analysis.NewCPLAnalyzer()
	countryConceptDef := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConceptDef)

	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
	}
	defs.Measures["Day"] = analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
	}

	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate connector definition")
	assert.Contains(t, err.Error(), "SalesData")
}

// TestAnalyzeConnectorsScope_StructuredRecordUndefinedDimension_Error verifies
// that undefined dimension concepts are reported when dimensions come from a
// ConnectorDeclaration structured record.
func TestAnalyzeConnectorsScope_StructuredRecordUndefinedDimension_Error(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)
	connectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		DimensionNames: []string{"UndefinedConcept"},
		CellTypeName:   "Revenue",
	}

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
	}

	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "undefined concept dimension")
	assert.Contains(t, err.Error(), "UndefinedConcept")
}

// TestAnalyzeConnectorsScope_StructuredRecordUndefinedCelltype_Error verifies
// that undefined celltype measures are reported when the celltype name comes
// from a ConnectorDeclaration structured record.
func TestAnalyzeConnectorsScope_StructuredRecordUndefinedCelltype_Error(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)
	connectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		DimensionNames: []string{"Country"},
		CellTypeName:   "UndefinedMeasure",
	}

	analyzer := analysis.NewCPLAnalyzer()
	countryConceptDef := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConceptDef)

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
	}

	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "undefined celltype measure")
	assert.Contains(t, err.Error(), "UndefinedMeasure")
}

// TestAnalyzeConnectorsScope_StructuredRecordUndefinedGrain_Error verifies that
// undefined grain measures are reported when the grain name comes from a
// ConnectorDeclaration structured record.
func TestAnalyzeConnectorsScope_StructuredRecordUndefinedGrain_Error(t *testing.T) {
	t.Parallel()

	root := symboltable.NewGlobalSymbolTable()
	connectorsScope := root.AddChild(symboltable.ConnectorsScope)
	connectorScope := connectorsScope.AddChild(symboltable.ConnectorScope)
	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:           "SalesData",
		DimensionNames: []string{"Country"},
		GrainNames:     []string{"UndefinedGrain"},
		CellTypeName:   "Revenue",
	}

	analyzer := analysis.NewCPLAnalyzer()
	countryConceptDef := analysis.AnalysisConceptDefinition{ConceptName: "Country"}
	analyzer.AddNewConcept(&countryConceptDef)

	defs := analyzer.GetAnalysisDefinitions()
	defs.Measures["Revenue"] = analysis.AnalysisMeasureDefinition{
		Name:     "Revenue",
		TypeName: "Number",
	}

	entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.ConnectorScope: analysis.AnalyzeConnectorsScope,
	}

	err := analysis.WalkSymbolTableWithHandlers(root, entryHandlers, nil, analyzer)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "undefined grain measure")
	assert.Contains(t, err.Error(), "UndefinedGrain")
}

// ============================================================================
// ValidateConnectorGrainTypes Multiple Grains Tests
// ============================================================================

// TestValidateConnectorGrainTypes_MultipleIntervalGrains_Accepted verifies that
// a connector with multiple grains, all of which have the built-in Interval type,
// passes deferred grain type validation.
func TestValidateConnectorGrainTypes_MultipleIntervalGrains_Accepted(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()

	yearMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "Year",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "year"},
	}
	monthMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "Month",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "month"},
	}
	dayMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}

	defs.Connectors["MultiGrainConnector"] = analysis.AnalysisConnectorDefinition{
		Name:  "MultiGrainConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&yearMeasure, &monthMeasure, &dayMeasure},
	}

	err := analyzer.ValidateConnectorGrainTypes()

	require.NoError(t, err, "Multiple Interval grains should be accepted")
}

// TestValidateConnectorGrainTypes_MultipleGrains_FirstInvalid_Rejected verifies
// that when the first grain in a multi-grain connector has an invalid type, the
// deferred validation returns an error identifying the connector, grain index [0],
// grain measure name, and invalid type.
func TestValidateConnectorGrainTypes_MultipleGrains_FirstInvalid_Rejected(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()

	numberMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "CountGrain",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "count"},
	}
	intervalMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}

	defs.Connectors["MixedGrainConnector"] = analysis.AnalysisConnectorDefinition{
		Name:  "MixedGrainConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&numberMeasure, &intervalMeasure},
	}

	err := analyzer.ValidateConnectorGrainTypes()

	require.Error(t, err, "Should reject connector when first grain has invalid type")
	assert.Contains(t, err.Error(), "MixedGrainConnector", "Error should identify the connector")
	assert.Contains(t, err.Error(), "[0]", "Error should identify grain index [0]")
	assert.Contains(t, err.Error(), "CountGrain", "Error should identify the grain measure")
	assert.Contains(t, err.Error(), "Number", "Error should identify the invalid type")
}

// TestValidateConnectorGrainTypes_MultipleGrains_SecondInvalid_Rejected verifies
// that when the second grain in a multi-grain connector has an invalid type, the
// deferred validation returns an error identifying grain index [1].
func TestValidateConnectorGrainTypes_MultipleGrains_SecondInvalid_Rejected(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()

	intervalMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}
	numberMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "CountGrain",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "count"},
	}

	defs.Connectors["MixedGrainConnector"] = analysis.AnalysisConnectorDefinition{
		Name:  "MixedGrainConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&intervalMeasure, &numberMeasure},
	}

	err := analyzer.ValidateConnectorGrainTypes()

	require.Error(t, err, "Should reject connector when second grain has invalid type")
	assert.Contains(t, err.Error(), "[1]", "Error should identify grain index [1]")
}

// TestValidateConnectorGrainTypes_MultipleGrains_DerivedInterval_Accepted verifies
// that a connector with multiple grains where one derives from Interval and
// another is the built-in Interval type passes deferred validation.
func TestValidateConnectorGrainTypes_MultipleGrains_DerivedInterval_Accepted(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Daily",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.IntervalType),
		DerivationPath: []string{"Daily", "Interval"},
	})

	defs := analyzer.GetAnalysisDefinitions()

	derivedMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "DailyGrain",
		TypeName: "Daily",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}
	intervalMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}

	defs.Connectors["DerivedMultiGrainConnector"] = analysis.AnalysisConnectorDefinition{
		Name:  "DerivedMultiGrainConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&derivedMeasure, &intervalMeasure},
	}

	err := analyzer.ValidateConnectorGrainTypes()

	require.NoError(t, err, "Derived Interval and built-in Interval grains should both be accepted")
}
