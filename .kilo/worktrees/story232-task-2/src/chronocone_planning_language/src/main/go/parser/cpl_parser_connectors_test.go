// Package parser_test contains tests for the CPL parser connectors section.
package parser_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseSingleConnectorDefinition tests parsing a single connector with dimensions, grain, and celltype.
func TestParseSingleConnectorDefinition(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tSalesData = dimensions [Store, Product] grain Day celltype SalesMeasure"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify the ConnectorsScope contains one ConnectorScope with structured data.
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")
	connectorsScope := doc.Children[0]
	assert.Equal(t, symboltable.ConnectorsScope, connectorsScope.ScopeType, "Expected ConnectorsScope")
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")
	assert.Equal(t, symboltable.ConnectorScope, connectorsScope.Children[0].ScopeType, "Expected ConnectorScope")

	decl := requireConnectorDeclaration(t, connectorsScope.Children[0])
	assert.Equal(t, "SalesData", decl.Name, "ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"Store", "Product"}, decl.DimensionNames, "ConnectorDeclaration DimensionNames should preserve parsed dimensions in order")
	require.NotEmpty(t, decl.GrainNames, "ConnectorDeclaration should have at least one grain")
	assert.Equal(t, "Day", decl.GrainNames[0], "ConnectorDeclaration GrainNames[0] should match grain identifier")
	assert.Equal(t, "SalesMeasure", decl.CellTypeName, "ConnectorDeclaration CellTypeName should match celltype identifier")
}

// TestParseMultipleConnectorDefinitions tests parsing multiple connectors in one section.
func TestParseMultipleConnectorDefinitions(t *testing.T) {
	t.Parallel()

	input := `connectors
	SalesData = dimensions [Store, Product] grain Day celltype SalesMeasure
	InventoryData = dimensions [Warehouse, SKU] grain Week celltype CountMeasure`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify the ConnectorsScope contains two ConnectorScopes with structured data.
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")
	connectorsScope := doc.Children[0]
	assert.Equal(t, symboltable.ConnectorsScope, connectorsScope.ScopeType, "Expected ConnectorsScope")
	require.Len(t, connectorsScope.Children, 2, "Expected two ConnectorScope children")

	firstDecl := requireConnectorDeclaration(t, connectorsScope.Children[0])
	assert.Equal(t, "SalesData", firstDecl.Name, "First ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"Store", "Product"}, firstDecl.DimensionNames, "First ConnectorDeclaration DimensionNames should preserve parsed dimensions in order")
	require.NotEmpty(t, firstDecl.GrainNames, "First ConnectorDeclaration should have at least one grain")
	assert.Equal(t, "Day", firstDecl.GrainNames[0], "First ConnectorDeclaration GrainNames[0] should match grain identifier")
	assert.Equal(t, "SalesMeasure", firstDecl.CellTypeName, "First ConnectorDeclaration CellTypeName should match celltype identifier")

	secondDecl := requireConnectorDeclaration(t, connectorsScope.Children[1])
	assert.Equal(t, "InventoryData", secondDecl.Name, "Second ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"Warehouse", "SKU"}, secondDecl.DimensionNames, "Second ConnectorDeclaration DimensionNames should preserve parsed dimensions in order")
	require.NotEmpty(t, secondDecl.GrainNames, "Second ConnectorDeclaration should have at least one grain")
	assert.Equal(t, "Week", secondDecl.GrainNames[0], "Second ConnectorDeclaration GrainNames[0] should match grain identifier")
	assert.Equal(t, "CountMeasure", secondDecl.CellTypeName, "Second ConnectorDeclaration CellTypeName should match celltype identifier")
}

// TestParseConnectorWithSingleDimension tests connector with only one dimension.
func TestParseConnectorWithSingleDimension(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tSimple = dimensions [Region] grain Day celltype Count"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")
	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")
	decl := requireConnectorDeclaration(t, connectorsScope.Children[0])
	assert.Equal(t, "Simple", decl.Name, "ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"Region"}, decl.DimensionNames, "ConnectorDeclaration DimensionNames should preserve the single parsed dimension")
	require.NotEmpty(t, decl.GrainNames, "ConnectorDeclaration should have at least one grain")
	assert.Equal(t, "Day", decl.GrainNames[0], "ConnectorDeclaration GrainNames[0] should match grain identifier")
	assert.Equal(t, "Count", decl.CellTypeName, "ConnectorDeclaration CellTypeName should match celltype identifier")
}

// TestParseConnectorWithManyDimensions tests connector with many dimensions.
func TestParseConnectorWithManyDimensions(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tComplex = dimensions [Dim1, Dim2, Dim3, Dim4, Dim5, Dim6] grain Day celltype Value"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")
	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")
	decl := requireConnectorDeclaration(t, connectorsScope.Children[0])
	assert.Equal(t, "Complex", decl.Name, "ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"Dim1", "Dim2", "Dim3", "Dim4", "Dim5", "Dim6"}, decl.DimensionNames, "ConnectorDeclaration DimensionNames should preserve all parsed dimensions in order")
	require.NotEmpty(t, decl.GrainNames, "ConnectorDeclaration should have at least one grain")
	assert.Equal(t, "Day", decl.GrainNames[0], "ConnectorDeclaration GrainNames[0] should match grain identifier")
	assert.Equal(t, "Value", decl.CellTypeName, "ConnectorDeclaration CellTypeName should match celltype identifier")
}

// TestParseConnectorsSourceLocation tests that source locations are captured for connector definitions.
func TestParseConnectorsSourceLocation(t *testing.T) {
	t.Parallel()

	input := `connectors
	First = dimensions [A, B] grain Day celltype M1
	Second = dimensions [C, D] grain Week celltype M2`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error")
	require.NotNil(t, doc, "Document should not be nil")

	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")
	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 2, "Expected two ConnectorScope children")

	firstDecl := requireConnectorDeclaration(t, connectorsScope.Children[0])
	assert.Equal(t, "First", firstDecl.Name, "First ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"A", "B"}, firstDecl.DimensionNames, "First ConnectorDeclaration DimensionNames should preserve parsed dimensions in order")
	assert.Equal(t, "M1", firstDecl.CellTypeName, "First ConnectorDeclaration CellTypeName should match celltype identifier")
	assert.Positive(t, firstDecl.NameLocation.Line, "First connector name should capture a source line")
	require.Len(t, firstDecl.DimensionLocations, 2, "First ConnectorDeclaration DimensionLocations should match dimension count")
	assert.Positive(t, firstDecl.DimensionLocations[0].Line, "First dimension should capture a source line")
	require.NotEmpty(t, firstDecl.GrainNames, "First ConnectorDeclaration should have at least one grain")
	assert.Equal(t, "Day", firstDecl.GrainNames[0], "First ConnectorDeclaration GrainNames[0] should match grain identifier")
	assert.Positive(t, firstDecl.GrainLocations[0].Line, "First grain should capture a source line")
	assert.Positive(t, firstDecl.CellTypeLocation.Line, "First celltype should capture a source line")

	secondDecl := requireConnectorDeclaration(t, connectorsScope.Children[1])
	assert.Equal(t, "Second", secondDecl.Name, "Second ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"C", "D"}, secondDecl.DimensionNames, "Second ConnectorDeclaration DimensionNames should preserve parsed dimensions in order")
	assert.Equal(t, "M2", secondDecl.CellTypeName, "Second ConnectorDeclaration CellTypeName should match celltype identifier")
	assert.Positive(t, secondDecl.NameLocation.Line, "Second connector name should capture a source line")
	require.Len(t, secondDecl.DimensionLocations, 2, "Second ConnectorDeclaration DimensionLocations should match dimension count")
	assert.Positive(t, secondDecl.DimensionLocations[0].Line, "Second dimension should capture a source line")
	require.NotEmpty(t, secondDecl.GrainNames, "Second ConnectorDeclaration should have at least one grain")
	assert.Equal(t, "Week", secondDecl.GrainNames[0], "Second ConnectorDeclaration GrainNames[0] should match grain identifier")
	assert.Positive(t, secondDecl.GrainLocations[0].Line, "Second grain should capture a source line")
	assert.Positive(t, secondDecl.CellTypeLocation.Line, "Second celltype should capture a source line")
}

// TestParseConnectorsInvalidSyntax tests error handling for invalid syntax patterns in connectors section.
func TestParseConnectorsInvalidSyntax(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name:          "missing connectors keyword",
			input:         "\tSalesData = dimensions [Store] grain Day celltype Measure",
			expectError:   true,
			errorContains: "mismatched input",
		},
		{
			name:          "missing dimensions keyword",
			input:         "connectors\n\tSalesData = [Store] grain Day celltype Measure",
			expectError:   true,
			errorContains: "dimensions",
		},
		{
			name:          "missing grain keyword",
			input:         "connectors\n\tSalesData = dimensions [Store] celltype Measure",
			expectError:   true,
			errorContains: "grain",
		},
		{
			name:          "missing celltype keyword",
			input:         "connectors\n\tSalesData = dimensions [Store] grain Day",
			expectError:   true,
			errorContains: "celltype",
		},
		{
			name:          "missing opening bracket for dimensions",
			input:         "connectors\n\tSalesData = dimensions Store] grain Day celltype Measure",
			expectError:   true,
			errorContains: "[",
		},
		{
			name:          "missing closing bracket for dimensions",
			input:         "connectors\n\tSalesData = dimensions [Store grain Day celltype Measure",
			expectError:   true,
			errorContains: "]",
		},
		{
			name:          "empty dimensions list",
			input:         "connectors\n\tSalesData = dimensions [] grain Day celltype Measure",
			expectError:   true,
			errorContains: "IDENTIFIER",
		},
		{
			name:          "invalid identifier in dimensions",
			input:         "connectors\n\tSalesData = dimensions [123invalid] grain Day celltype Measure",
			expectError:   true,
			errorContains: "mismatched input '123invalid' expecting",
		},
		{
			name:          "missing comma between dimensions",
			input:         "connectors\n\tSalesData = dimensions [A B] grain Day celltype Measure",
			expectError:   true,
			errorContains: "B",
		},
		{
			name:          "trailing comma in dimensions",
			input:         "connectors\n\tSalesData = dimensions [A, B,] grain Day celltype Measure",
			expectError:   true,
			errorContains: "IDENTIFIER",
		},
		{
			name:          "invalid connector name",
			input:         "connectors\n\t123Bad = dimensions [Store] grain Day celltype Measure",
			expectError:   true,
			errorContains: "extraneous input",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := parser.Parse(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected an error for input: %s", tc.input)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, "Error message should contain expected text")
				}
				return
			}

			require.NoError(t, err, "Expected no error for input: %s", tc.input)
		})
	}
}

// TestParseConnectorWithoutGrainProducesError tests that a connector declaration without grain produces a parse error.
func TestParseConnectorWithoutGrainProducesError(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tSalesData = dimensions [Store, Product] celltype SalesMeasure"

	_, err := parser.Parse(input)

	require.Error(t, err, "Expected an error for connector without grain: %s", input)
	assert.Contains(t, err.Error(), "grain", "Error message should mention missing grain")
}

// TestParseConnectorsRequiresAtLeastOneDefinition tests that connectors section requires at least one definition.
func TestParseConnectorsRequiresAtLeastOneDefinition(t *testing.T) {
	t.Parallel()

	// The grammar requires connectorDeclaration+, so at least one definition is required
	input := "connectors"

	_, err := parser.Parse(input)

	// Should fail because connectorDeclaration+ requires at least one connector declaration
	require.Error(t, err, "Expected an error for connectors section without definitions")
}

// TestParseConnectorsWithOtherSections tests parsing connectors alongside types, units, measures sections.
func TestParseConnectorsWithOtherSections(t *testing.T) {
	t.Parallel()

	input := `types
	Length = Number(precision=2)
	Mass = Number()

units
	Meter = Length
	Kilogram = Mass

measures
	Distance = Length(unit=Meter)
	Weight = Mass(unit=Kilogram)

connectors
	SensorData = dimensions [Location, Time] grain Day celltype Distance
	WeightData = dimensions [Scale, Batch] grain Week celltype Weight`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")

	// Verify we have TypesScope, UnitsScope, MeasuresScope, and ConnectorsScope.
	require.Len(t, doc.Children, 4, "Expected four child scopes (TypesScope, UnitsScope, MeasuresScope, ConnectorsScope)")
	connectorsScope := doc.Children[3]
	assert.Equal(t, symboltable.ConnectorsScope, connectorsScope.ScopeType, "Expected fourth child to be ConnectorsScope")
	require.Len(t, connectorsScope.Children, 2, "Expected two ConnectorScope children")

	firstDecl := requireConnectorDeclaration(t, connectorsScope.Children[0])
	assert.Equal(t, "SensorData", firstDecl.Name, "First ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"Location", "Time"}, firstDecl.DimensionNames, "First ConnectorDeclaration DimensionNames should preserve parsed dimensions in order")
	require.NotEmpty(t, firstDecl.GrainNames, "First ConnectorDeclaration should have at least one grain")
	assert.Equal(t, "Day", firstDecl.GrainNames[0], "First ConnectorDeclaration GrainNames[0] should match grain identifier")
	assert.Equal(t, "Distance", firstDecl.CellTypeName, "First ConnectorDeclaration CellTypeName should match celltype identifier")

	secondDecl := requireConnectorDeclaration(t, connectorsScope.Children[1])
	assert.Equal(t, "WeightData", secondDecl.Name, "Second ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"Scale", "Batch"}, secondDecl.DimensionNames, "Second ConnectorDeclaration DimensionNames should preserve parsed dimensions in order")
	require.NotEmpty(t, secondDecl.GrainNames, "Second ConnectorDeclaration should have at least one grain")
	assert.Equal(t, "Week", secondDecl.GrainNames[0], "Second ConnectorDeclaration GrainNames[0] should match grain identifier")
	assert.Equal(t, "Weight", secondDecl.CellTypeName, "Second ConnectorDeclaration CellTypeName should match celltype identifier")
}

// TestParseConnectorsWithWhitespaceVariations tests parsing with various whitespace configurations.
func TestParseConnectorsWithWhitespaceVariations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "tabs for indentation",
			input: "connectors\n\tConn = dimensions [A, B] grain Day celltype M",
		},
		{
			name:  "spaces for indentation",
			input: "connectors\n    Conn = dimensions [A, B] grain Day celltype M",
		},
		{
			name:  "multiple spaces around equals",
			input: "connectors\n\tConn  =  dimensions [A, B] grain Day celltype M",
		},
		{
			name:  "spaces after opening bracket",
			input: "connectors\n\tConn = dimensions [ A, B ] grain Day celltype M",
		},
		{
			name:  "no spaces in dimensions list",
			input: "connectors\n\tConn = dimensions [A,B] grain Day celltype M",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := parser.Parse(tc.input)

			require.NoError(t, err, "Expected no error for input: %s", tc.input)
			require.NotNil(t, doc, "Document should not be nil")

			// Verify ConnectorsScope was created
			require.Len(t, doc.Children, 1, "Expected one child scope")
		})
	}
}

// TestParseConnectorsScopeIntegration tests that ConnectorsScope is properly pushed/popped.
func TestParseConnectorsScopeIntegration(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tConn = dimensions [A] grain Day celltype M"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error")
	require.NotNil(t, doc, "Document should not be nil")

	// Verify ConnectorsScope was created with structured ConnectorDeclaration data.
	require.Len(t, doc.Children, 1, "Expected one child scope")
	connectorsScope := doc.Children[0]
	assert.Equal(t, symboltable.ConnectorsScope, connectorsScope.ScopeType, "Expected ConnectorsScope")
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")

	decl := requireConnectorDeclaration(t, connectorsScope.Children[0])
	assert.Equal(t, "Conn", decl.Name, "ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"A"}, decl.DimensionNames, "ConnectorDeclaration DimensionNames should preserve parsed dimensions in order")
	require.NotEmpty(t, decl.GrainNames, "ConnectorDeclaration should have at least one grain")
	assert.Equal(t, "Day", decl.GrainNames[0], "ConnectorDeclaration GrainNames[0] should match grain identifier")
	assert.Equal(t, "M", decl.CellTypeName, "ConnectorDeclaration CellTypeName should match celltype identifier")
}

// TestParseConnectorWithGrainExtractsScopeStructure tests that a
// connector with grain parses successfully and produces the expected scope
// structure.
func TestParseConnectorWithGrainExtractsScopeStructure(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tSalesData = dimensions [Store, Product] grain Day celltype SalesMeasure"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")

	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")
}

// TestParseConnectorSingleDimensionExtractsScopeStructure tests that
// a single-dimension connector parses successfully and produces the expected
// scope structure.
func TestParseConnectorSingleDimensionExtractsScopeStructure(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tSimple = dimensions [Region] grain Day celltype Count"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")

	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")
}

// TestParseConnectorManyDimensionsExtractsScopeStructure tests that
// a connector with many dimensions parses successfully and produces the expected
// scope structure.
func TestParseConnectorManyDimensionsExtractsScopeStructure(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tComplex = dimensions [Dim1, Dim2, Dim3, Dim4, Dim5, Dim6] grain Day celltype Value"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")

	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")
}

// TestParseMultipleConnectorsWithGrainExtractsDefinitions tests that
// multiple connectors with grain parse successfully and produce
// the expected scope structure.
func TestParseMultipleConnectorsWithGrainExtractsDefinitions(t *testing.T) {
	t.Parallel()

	input := `connectors
	SalesData = dimensions [Store, Product] grain Day celltype SalesMeasure
	InventoryData = dimensions [Warehouse, SKU] grain Week celltype CountMeasure
	OrderData = dimensions [Customer, Product] grain Week celltype OrderMeasure`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")

	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 3, "Expected three ConnectorScopes")
}

// TestParseConnectorBackwardCompatibilityDefinitionsSlice tests that identifiers
// continue to be captured in the ConnectorScope's Definitions slice as parse-symbol
// artifacts even though connector analysis now consumes ConnectorScope.Data.
func TestParseConnectorBackwardCompatibilityDefinitionsSlice(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tSalesData = dimensions [Store, Product] grain Day celltype SalesMeasure"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")

	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")

	connectorScope := connectorsScope.Children[0]
	require.NotEmpty(t, connectorScope.Definitions, "ConnectorScope Definitions should not be empty for backward compatibility")

	// Collect all identifier names from Definitions
	var names []string
	for _, def := range connectorScope.Definitions {
		if def.Name != nil {
			names = append(names, *def.Name)
		}
	}

	// Should contain connector name, dimensions, grain, and celltype
	assert.Contains(t, names, "SalesData", "Definitions should contain connector name")
	assert.Contains(t, names, "Store", "Definitions should contain dimension Store")
	assert.Contains(t, names, "Product", "Definitions should contain dimension Product")
	assert.Contains(t, names, "Day", "Definitions should contain grain Day")
	assert.Contains(t, names, "SalesMeasure", "Definitions should contain celltype SalesMeasure")
}

// TestParseConnectorInvalidGrainSyntax tests error handling for invalid grain syntax patterns.
func TestParseConnectorInvalidGrainSyntax(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name:          "missing grain value",
			input:         "connectors\n\tSalesData = dimensions [Store, Product] grain celltype SalesMeasure",
			expectError:   true,
			errorContains: "IDENTIFIER",
		},
		{
			name:          "grain with numeric value",
			input:         "connectors\n\tSalesData = dimensions [Store, Product] grain 123 celltype SalesMeasure",
			expectError:   true,
			errorContains: "mismatched input",
		},
		{
			name:          "grain before dimensions",
			input:         "connectors\n\tSalesData = grain Day dimensions [Store, Product] celltype SalesMeasure",
			expectError:   true,
			errorContains: "dimensions",
		},
		{
			name:          "grain after celltype",
			input:         "connectors\n\tSalesData = dimensions [Store, Product] celltype SalesMeasure grain Day",
			expectError:   true,
			errorContains: "mismatched input 'celltype' expecting 'grain'",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := parser.Parse(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected an error for input: %s", tc.input)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, "Error message should contain expected text")
				}
				return
			}

			require.NoError(t, err, "Expected no error for input: %s", tc.input)
		})
	}
}

// TestParseConnectorWithGrainRegistersGrainDefinition tests that when a connector
// includes a grain keyword, the parser registers a SymbolDefinition with
// DefinitionType GrainDefinition in the ConnectorScope's Definitions slice.
func TestParseConnectorWithGrainRegistersGrainDefinition(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tSalesData = dimensions [Store, Product] grain Day celltype SalesMeasure"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")

	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")

	connectorScope := connectorsScope.Children[0]

	var grainDef *symboltable.SymbolDefinition
	for i := range connectorScope.Definitions {
		if connectorScope.Definitions[i].DefinitionType == symboltable.GrainDefinition {
			grainDef = &connectorScope.Definitions[i]
			break
		}
	}

	require.NotNil(t, grainDef, "ConnectorScope.Definitions should contain a SymbolDefinition with DefinitionType GrainDefinition")
	require.NotNil(t, grainDef.Name, "GrainDefinition Name should not be nil")
	assert.Equal(t, "Day", *grainDef.Name, "GrainDefinition Name should be 'Day'")
}

// TestParseConnectorWithGrain_StructuredRecord verifies that parser output attaches
// a structured ConnectorDeclaration record to ConnectorScope.Data for a connector
// declaration that includes a grain.
func TestParseConnectorWithGrain_StructuredRecord(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tSalesData = dimensions [Store, Product] grain Day celltype SalesMeasure"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")

	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")

	connectorScope := connectorsScope.Children[0]
	decl := requireConnectorDeclaration(t, connectorScope)

	assert.Equal(t, "SalesData", decl.Name, "ConnectorDeclaration Name should match connector identifier")
	assert.Positive(t, decl.NameLocation.Line, "ConnectorDeclaration NameLocation should capture a source line")
	assert.Equal(t, []string{"Store", "Product"}, decl.DimensionNames, "ConnectorDeclaration DimensionNames should preserve parsed dimensions in order")
	require.Len(t, decl.DimensionLocations, 2, "ConnectorDeclaration DimensionLocations should match dimension count")
	assert.Positive(t, decl.DimensionLocations[0].Line, "First dimension should capture a source line")
	assert.Positive(t, decl.DimensionLocations[1].Line, "Second dimension should capture a source line")
	require.NotEmpty(t, decl.GrainNames, "ConnectorDeclaration should have at least one grain")
	assert.Equal(t, "Day", decl.GrainNames[0], "ConnectorDeclaration GrainNames[0] should match grain identifier")
	assert.Positive(t, decl.GrainLocations[0].Line, "ConnectorDeclaration GrainLocations[0] should capture a source line")
	assert.Equal(t, "SalesMeasure", decl.CellTypeName, "ConnectorDeclaration CellTypeName should match celltype identifier")
	assert.Positive(t, decl.CellTypeLocation.Line, "ConnectorDeclaration CellTypeLocation should capture a source line")
}

// TestParseMultipleConnectors_StructuredRecords verifies that each connector in a
// connectors section receives its own structured ConnectorDeclaration record.
func TestParseMultipleConnectors_StructuredRecords(t *testing.T) {
	t.Parallel()

	input := `connectors
	SalesData = dimensions [Store, Product] grain Day celltype SalesMeasure
	InventoryData = dimensions [Warehouse, SKU] grain Week celltype CountMeasure`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")

	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 2, "Expected two ConnectorScope children")

	firstDecl := requireConnectorDeclaration(t, connectorsScope.Children[0])
	assert.Equal(t, "SalesData", firstDecl.Name, "First ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"Store", "Product"}, firstDecl.DimensionNames, "First ConnectorDeclaration DimensionNames should preserve parsed dimensions in order")
	require.NotEmpty(t, firstDecl.GrainNames, "First ConnectorDeclaration should have at least one grain")
	assert.Equal(t, "Day", firstDecl.GrainNames[0], "First ConnectorDeclaration GrainNames[0] should match grain identifier")
	assert.Equal(t, "SalesMeasure", firstDecl.CellTypeName, "First ConnectorDeclaration CellTypeName should match celltype identifier")

	secondDecl := requireConnectorDeclaration(t, connectorsScope.Children[1])
	assert.Equal(t, "InventoryData", secondDecl.Name, "Second ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"Warehouse", "SKU"}, secondDecl.DimensionNames, "Second ConnectorDeclaration DimensionNames should preserve parsed dimensions in order")
	require.NotEmpty(t, secondDecl.GrainNames, "Second ConnectorDeclaration should have at least one grain")
	assert.Equal(t, "Week", secondDecl.GrainNames[0], "Second ConnectorDeclaration GrainNames[0] should match grain identifier")
	assert.Equal(t, "CountMeasure", secondDecl.CellTypeName, "Second ConnectorDeclaration CellTypeName should match celltype identifier")
}

// TestParseConnectorSameNameDimensionAndGrain_StructuredRecord reproduces the
// same-name dimension and grain bug and verifies that matching names do not cause
// dimension loss in the structured ConnectorDeclaration record.
func TestParseConnectorSameNameDimensionAndGrain_StructuredRecord(t *testing.T) {
	t.Parallel()

	input := `connectors
	SalesData = dimensions [Store, Day] grain Day celltype SalesMeasure`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")

	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")

	connectorScope := connectorsScope.Children[0]
	decl := requireConnectorDeclaration(t, connectorScope)

	assert.Equal(t, []string{"Store", "Day"}, decl.DimensionNames, "ConnectorDeclaration DimensionNames should preserve all dimensions even when one matches the grain")
	require.Len(t, decl.DimensionLocations, 2, "ConnectorDeclaration DimensionLocations should match dimension count")
	require.NotEmpty(t, decl.GrainNames, "ConnectorDeclaration should have at least one grain")
	assert.Equal(t, "Day", decl.GrainNames[0], "ConnectorDeclaration GrainNames[0] should still capture the grain identifier")
	assert.Equal(t, "SalesMeasure", decl.CellTypeName, "ConnectorDeclaration CellTypeName should match celltype identifier")
}

// TestParseConnectorSingleDimension_StructuredRecord verifies structured record
// population for a connector with exactly one declared dimension.
func TestParseConnectorSingleDimension_StructuredRecord(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tSimple = dimensions [Region] grain Day celltype Count"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")

	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")

	connectorScope := connectorsScope.Children[0]
	decl := requireConnectorDeclaration(t, connectorScope)

	assert.Equal(t, "Simple", decl.Name, "ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"Region"}, decl.DimensionNames, "ConnectorDeclaration DimensionNames should preserve the single parsed dimension")
	require.Len(t, decl.DimensionLocations, 1, "ConnectorDeclaration DimensionLocations should contain one location")
	assert.Positive(t, decl.DimensionLocations[0].Line, "Single dimension should capture a source line")
	require.NotEmpty(t, decl.GrainNames, "ConnectorDeclaration should have at least one grain")
	assert.Equal(t, "Day", decl.GrainNames[0], "ConnectorDeclaration GrainNames[0] should match grain identifier")
	assert.Equal(t, "Count", decl.CellTypeName, "ConnectorDeclaration CellTypeName should match celltype identifier")
}

// TestParseConnectorWithMultipleGrains parses a connector with 3 grain identifiers and verifies
// GrainNames and GrainLocations are correctly populated.
func TestParseConnectorWithMultipleGrains(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tSalesData = dimensions [Store, Product] grain Day, Week, Month celltype SalesMeasure"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")

	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")

	decl := requireConnectorDeclaration(t, connectorsScope.Children[0])
	assert.Equal(t, "SalesData", decl.Name, "ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"Store", "Product"}, decl.DimensionNames, "ConnectorDeclaration DimensionNames should preserve parsed dimensions in order")
	assert.Equal(t, []string{"Day", "Week", "Month"}, decl.GrainNames, "ConnectorDeclaration GrainNames should contain all 3 grain identifiers in order")
	require.Len(t, decl.GrainLocations, 3, "ConnectorDeclaration GrainLocations should have 3 entries")
	for i, loc := range decl.GrainLocations {
		assert.Positive(t, loc.Line, "Grain location %d should capture a valid line number", i)
	}
	assert.Equal(t, "Day", decl.GrainNames[0], "GrainNames[0] should equal the first grain identifier")
	assert.Positive(t, decl.GrainLocations[0].Line, "GrainLocations[0] should capture a valid line number")
	assert.Equal(t, "SalesMeasure", decl.CellTypeName, "ConnectorDeclaration CellTypeName should match celltype identifier")
}

// TestParseConnectorWithTwoGrains parses a connector with 2 grain identifiers.
func TestParseConnectorWithTwoGrains(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tSalesData = dimensions [Store, Product] grain Day, Week celltype SalesMeasure"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")

	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")

	decl := requireConnectorDeclaration(t, connectorsScope.Children[0])
	assert.Equal(t, []string{"Day", "Week"}, decl.GrainNames, "ConnectorDeclaration GrainNames should contain both grain identifiers in order")
	require.Len(t, decl.GrainLocations, 2, "ConnectorDeclaration GrainLocations should have 2 entries")
	assert.Positive(t, decl.GrainLocations[0].Line, "First grain should capture a source line")
	assert.Positive(t, decl.GrainLocations[1].Line, "Second grain should capture a source line")
	assert.Equal(t, "Day", decl.GrainNames[0], "GrainNames[0] should equal the first grain identifier")
}

// TestParseConnectorSingleGrain parses a connector with 1 grain and verifies
// that the slice fields contain a single element.
func TestParseConnectorSingleGrain(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tSalesData = dimensions [Store, Product] grain Day celltype SalesMeasure"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")

	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")

	decl := requireConnectorDeclaration(t, connectorsScope.Children[0])
	assert.Equal(t, []string{"Day"}, decl.GrainNames, "GrainNames should be a single-element array")
	require.Len(t, decl.GrainLocations, 1, "GrainLocations should have exactly 1 entry for single grain")
	assert.Positive(t, decl.GrainLocations[0].Line, "Single grain location should capture a valid line number")
	assert.Equal(t, "Day", decl.GrainNames[0], "GrainNames[0] should equal 'Day'")
}

// TestParseMultipleConnectorsWithMixedGrainCounts parses multiple connectors where some have 1 grain
// and some have multiple grains, verifying each scope retains the correct grain data.
func TestParseMultipleConnectorsWithMixedGrainCounts(t *testing.T) {
	t.Parallel()

	input := `connectors
	SalesData = dimensions [Store, Product] grain Day celltype SalesMeasure
	InventoryData = dimensions [Warehouse, SKU] grain Week, Month celltype CountMeasure`

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")

	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 2, "Expected two ConnectorScope children")

	firstDecl := requireConnectorDeclaration(t, connectorsScope.Children[0])
	assert.Equal(t, "SalesData", firstDecl.Name, "First ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"Day"}, firstDecl.GrainNames, "First connector should have single grain")
	assert.Equal(t, "Day", firstDecl.GrainNames[0], "First connector GrainNames[0] should be 'Day'")

	secondDecl := requireConnectorDeclaration(t, connectorsScope.Children[1])
	assert.Equal(t, "InventoryData", secondDecl.Name, "Second ConnectorDeclaration Name should match connector identifier")
	assert.Equal(t, []string{"Week", "Month"}, secondDecl.GrainNames, "Second connector should have two grains in order")
	assert.Equal(t, "Week", secondDecl.GrainNames[0], "Second connector GrainNames[0] should be first grain 'Week'")
}

// TestParseConnectorMultipleGrains_RegisterAllGrainDefinitions verifies that each grain identifier
// gets registered as a GrainDefinition in the ConnectorScope's Definitions slice.
func TestParseConnectorMultipleGrains_RegisterAllGrainDefinitions(t *testing.T) {
	t.Parallel()

	input := "connectors\n\tSalesData = dimensions [Store, Product] grain Day, Week, Month celltype SalesMeasure"

	doc, err := parser.Parse(input)

	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, doc, "Document should not be nil")
	require.Len(t, doc.Children, 1, "Expected one child scope (ConnectorsScope)")

	connectorsScope := doc.Children[0]
	require.Len(t, connectorsScope.Children, 1, "Expected one ConnectorScope")

	connectorScope := connectorsScope.Children[0]

	var grainDefs []*symboltable.SymbolDefinition
	for i := range connectorScope.Definitions {
		if connectorScope.Definitions[i].DefinitionType == symboltable.GrainDefinition {
			grainDefs = append(grainDefs, &connectorScope.Definitions[i])
		}
	}

	require.Len(t, grainDefs, 3, "ConnectorScope.Definitions should contain 3 GrainDefinitions")
	require.NotNil(t, grainDefs[0].Name, "First GrainDefinition Name should not be nil")
	assert.Equal(t, "Day", *grainDefs[0].Name, "First GrainDefinition Name should be 'Day'")
	require.NotNil(t, grainDefs[1].Name, "Second GrainDefinition Name should not be nil")
	assert.Equal(t, "Week", *grainDefs[1].Name, "Second GrainDefinition Name should be 'Week'")
	require.NotNil(t, grainDefs[2].Name, "Third GrainDefinition Name should not be nil")
	assert.Equal(t, "Month", *grainDefs[2].Name, "Third GrainDefinition Name should be 'Month'")
}

func requireConnectorDeclaration(t *testing.T, connectorScope *symboltable.SymbolTable) *symboltable.ConnectorDeclaration {
	t.Helper()

	require.NotNil(t, connectorScope.Data, "ConnectorScope.Data should be populated")
	decl, ok := connectorScope.Data.(*symboltable.ConnectorDeclaration)
	require.True(t, ok, "ConnectorScope.Data should be a *symboltable.ConnectorDeclaration")
	require.NotNil(t, decl, "ConnectorScope.Data should not contain a nil ConnectorDeclaration pointer")

	return decl
}
