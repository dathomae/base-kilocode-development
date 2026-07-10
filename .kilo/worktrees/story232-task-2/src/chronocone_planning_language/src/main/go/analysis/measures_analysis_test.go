// Package analysis_test contains tests for the measures analysis phase.
package analysis_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getParserExamplesDir returns the path to the parser examples directory.
func getParserExamplesDir(t *testing.T) string {
	t.Helper()
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	// From src/main/go/analysis/ go up 4 levels to reach the module root (where go.mod is), then into examples
	projectRoot := filepath.Join(testDir, "..", "..", "..", "..")
	return filepath.Join(projectRoot, "examples")
}

// parseAndAnalyze parses a CPL input string and returns analysis definitions.
func parseAndAnalyze(input string) (*analysis.AnalysisDefinitions, error) {
	doc, err := parser.Parse(input)
	if err != nil {
		return nil, err
	}
	return analysis.Analyze(doc)
}

// buildMeasuresSymbolTable builds a symbol table with measures for testing.
func buildMeasuresSymbolTable(measureDefs []struct {
	measureName string
	typeName    string
	unitName    string
}) *symboltable.SymbolTable {
	// Build a symbol table tree for measures
	root := symboltable.NewGlobalSymbolTable()
	measuresScope := root.AddChild(symboltable.MeasuresScope)

	for _, def := range measureDefs {
		measureScope := measuresScope.AddChild(symboltable.MeasureScope)

		measureName := def.measureName
		typeName := def.typeName
		unitName := def.unitName

		measureScope.Define(&measureName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
		measureScope.Define(&typeName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 2})

		// Push unit reference as postfix definition
		_ = measureScope.DefineWithData(nil, symboltable.UnitReferenceDefinition, symboltable.SourceLocation{Line: 1, Column: 4}, parser.ParserUnitReference{Name: unitName})
	}

	return root
}

// TestAnalyzeMeasureScope_SingleMeasure tests analyzing a single measure definition.
func TestAnalyzeMeasureScope_SingleMeasure(t *testing.T) {
	t.Parallel()

	input := "measures\n\tDistance = Length(unit=Meter)"

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, defs, "AnalysisDefinitions should not be nil")

	// Verify the measure was analyzed
	require.Contains(t, defs.Measures, "Distance", "Distance measure should be in measures map")
	measure := defs.Measures["Distance"]
	assert.Equal(t, "Distance", measure.Name, "Measure name should be Distance")
	assert.Equal(t, "Length", measure.TypeName, "Type name should be Length")
	assert.Equal(t, "Meter", measure.GetUnitName(), "Unit name should be Meter")
}

// TestAnalyzeMeasureScope_MultipleMeasures tests analyzing multiple measure definitions.
func TestAnalyzeMeasureScope_MultipleMeasures(t *testing.T) {
	t.Parallel()

	input := `measures
	Distance = Length(unit=Meter)
	Weight = Mass(unit=Kilogram)
	Price = Currency(unit=Dollar)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Expected no error for input: %s", input)
	require.NotNil(t, defs, "AnalysisDefinitions should not be nil")

	// Verify all measures were analyzed
	require.Len(t, defs.Measures, 3, "Should have 3 measures")

	// Check Distance measure
	distance, exists := defs.Measures["Distance"]
	require.True(t, exists, "Distance measure should exist")
	assert.Equal(t, "Distance", distance.Name)
	assert.Equal(t, "Length", distance.TypeName)
	assert.Equal(t, "Meter", distance.GetUnitName())

	// Check Weight measure
	weight, exists := defs.Measures["Weight"]
	require.True(t, exists, "Weight measure should exist")
	assert.Equal(t, "Weight", weight.Name)
	assert.Equal(t, "Mass", weight.TypeName)
	assert.Equal(t, "Kilogram", weight.GetUnitName())

	// Check Price measure
	price, exists := defs.Measures["Price"]
	require.True(t, exists, "Price measure should exist")
	assert.Equal(t, "Price", price.Name)
	assert.Equal(t, "Currency", price.TypeName)
	assert.Equal(t, "Dollar", price.GetUnitName())
}

// TestAnalyzeMeasureScope_UnitReferenceResolution tests that measure references are stored correctly.
func TestAnalyzeMeasureScope_UnitReferenceResolution(t *testing.T) {
	t.Parallel()

	// This test verifies that the unit reference in the measure definition is captured correctly
	input := "measures\n\tTemperature = NumericType(unit=CelsiusScale)"

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Expected no error for input: %s", input)

	temperature, exists := defs.Measures["Temperature"]
	require.True(t, exists, "Temperature measure should exist")
	assert.Equal(t, "NumericType", temperature.TypeName, "Type name should be NumericType")
	assert.Equal(t, "CelsiusScale", temperature.GetUnitName(), "Unit name should be CelsiusScale")
}

// TestAnalyzeMeasureScope_DuplicateMeasure tests that duplicate measure names are detected.
func TestAnalyzeMeasureScope_DuplicateMeasure(t *testing.T) {
	t.Parallel()

	input := `measures
	Distance = Length(unit=Meter)
	Distance = Mass(unit=Kilogram)`

	_, err := parseAndAnalyze(input)
	require.Error(t, err, "Expected error for duplicate measure names")
	assert.Contains(t, err.Error(), "duplicate", "Error should mention duplicate")
}

// TestAnalyzeMeasureScope_WrongIdentifierCount tests error handling for wrong number of identifiers.
func TestAnalyzeMeasureScope_WrongIdentifierCount(t *testing.T) {
	t.Parallel()

	// Build a symbol table with only 1 identifier - should error
	root := symboltable.NewGlobalSymbolTable()
	measuresScope := root.AddChild(symboltable.MeasuresScope)
	measureScope := measuresScope.AddChild(symboltable.MeasureScope)

	measureName := "Distance"
	measureScope.Define(&measureName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})

	// Try to analyze
	analyzer := analysis.NewCPLAnalyzer()
	handlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MeasureScope: analysis.AnalyzeMeasureScope,
	}

	err := analysis.WalkSymbolTableWithHandlers(root, handlers, nil, analyzer)
	require.Error(t, err, "Expected error for wrong identifier count")
	assert.Contains(t, err.Error(), "at least two identifiers", "Error should mention at least two identifiers required")
}

// TestAnalyzeMeasureScope_EmptyDefinitions tests error handling for empty definitions.
func TestAnalyzeMeasureScope_EmptyDefinitions(t *testing.T) {
	t.Parallel()

	// Build a symbol table with empty MeasureScope
	root := symboltable.NewGlobalSymbolTable()
	measuresScope := root.AddChild(symboltable.MeasuresScope)
	_ = measuresScope.AddChild(symboltable.MeasureScope)
	// No definitions added

	// Try to analyze - should not error on empty definitions
	analyzer := analysis.NewCPLAnalyzer()
	handlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MeasureScope: analysis.AnalyzeMeasureScope,
	}

	err := analysis.WalkSymbolTableWithHandlers(root, handlers, nil, analyzer)
	require.NoError(t, err, "Should not error on empty definitions")
}

// TestIntegration_WithWalkSymbolTableWithHandlers tests integration with the symbol table walker.
func TestIntegration_WithWalkSymbolTableWithHandlers(t *testing.T) {
	t.Parallel()

	// Build a complete symbol table with measures
	root := symboltable.NewGlobalSymbolTable()
	measuresScope := root.AddChild(symboltable.MeasuresScope)

	measureDefs := []struct {
		measureName string
		typeName    string
		unitName    string
	}{
		{"Distance", "Length", "Meter"},
		{"Weight", "Mass", "Kilogram"},
	}

	for _, def := range measureDefs {
		measureScope := measuresScope.AddChild(symboltable.MeasureScope)

		measureName := def.measureName
		typeName := def.typeName
		unitName := def.unitName

		measureScope.Define(&measureName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 1})
		measureScope.Define(&typeName, symboltable.Identifier, symboltable.SourceLocation{Line: 1, Column: 2})

		// Push unit reference as postfix definition
		_ = measureScope.DefineWithData(nil, symboltable.UnitReferenceDefinition, symboltable.SourceLocation{Line: 1, Column: 4}, parser.ParserUnitReference{Name: unitName})
	}

	// Walk with handlers
	analyzer := analysis.NewCPLAnalyzer()
	handlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
		symboltable.MeasureScope: analysis.AnalyzeMeasureScope,
	}

	err := analysis.WalkSymbolTableWithHandlers(root, handlers, nil, analyzer)
	require.NoError(t, err, "Expected no error during walk")

	defs := analyzer.GetAnalysisDefinitions()
	require.Len(t, defs.Measures, 2, "Should have 2 measures")
	assert.Contains(t, defs.Measures, "Distance")
	assert.Contains(t, defs.Measures, "Weight")
}

// TestEndToEnd_ParseAndAnalyze tests complete parse and analyze workflow.
func TestEndToEnd_ParseAndAnalyze(t *testing.T) {
	t.Parallel()

	// Test with the measures-example.cpl file from examples
	examplesDir := getParserExamplesDir(t)
	examplePath := filepath.Join(examplesDir, "parse_and_analyze_only", "without_errors", "measures-example.cpl")

	// Check if the file exists
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Fatalf("measures-example.cpl not found at %s: %v", examplePath, err)
	}

	// Read the file
	content, err := os.ReadFile(examplePath)
	require.NoError(t, err, "Should be able to read example file")

	// Parse and analyze
	defs, err := parseAndAnalyze(string(content))
	require.NoError(t, err, "Should parse and analyze without error")
	require.NotNil(t, defs, "AnalysisDefinitions should not be nil")

	// Verify measures from the example file
	// From measures-example.cpl:
	// CurrencyMeasure = CurrencyType(unit=CurrencyUnit)
	// Currency = Currency(unit=Currency)
	require.Contains(t, defs.Measures, "CurrencyMeasure", "CurrencyMeasure should be analyzed")
	require.Contains(t, defs.Measures, "Currency", "Currency measure should be analyzed")

	// Verify CurrencyMeasure
	cm := defs.Measures["CurrencyMeasure"]
	assert.Equal(t, "CurrencyMeasure", cm.Name)
	assert.Equal(t, "CurrencyType", cm.TypeName)
	assert.Equal(t, "CurrencyUnit", cm.GetUnitName())

	// Verify Currency
	c := defs.Measures["Currency"]
	assert.Equal(t, "Currency", c.Name)
	assert.Equal(t, "Currency", c.TypeName)
	assert.Equal(t, "Currency", c.GetUnitName())
}

// TestAnalyzeMeasures_WithConceptsAndTypes tests measures analysis in context with other sections.
func TestAnalyzeMeasures_WithConceptsAndTypes(t *testing.T) {
	t.Parallel()

	input := `concepts
	Currency = [Values[USD, Euro, Yen]]

types
	CurrencyType = Number(precision=2)

units
	CurrencyUnit = Currency

measures
	Price = CurrencyType(unit=CurrencyUnit)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Expected no error for input: %s", input)

	// Verify concepts
	require.Contains(t, defs.Concepts, "Currency", "Currency concept should be analyzed")

	// Verify types
	require.Contains(t, defs.Types, "CurrencyType", "CurrencyType should be analyzed")

	// Verify units
	require.Contains(t, defs.Units, "CurrencyUnit", "CurrencyUnit should be analyzed")

	// Verify measures
	require.Contains(t, defs.Measures, "Price", "Price measure should be analyzed")
	price := defs.Measures["Price"]
	assert.Equal(t, "Price", price.Name)
	assert.Equal(t, "CurrencyType", price.TypeName)
	assert.Equal(t, "CurrencyUnit", price.GetUnitName())
}

// TestAnalyzeMeasures_LocationCaptured tests that source locations are captured in analysis.
func TestAnalyzeMeasures_LocationCaptured(t *testing.T) {
	t.Parallel()

	input := `measures
	Distance = Length(unit=Meter)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Expected no error for input: %s", input)

	distance := defs.Measures["Distance"]
	assert.Greater(t, distance.Location.Line, 0, "Location line should be set")
	assert.GreaterOrEqual(t, distance.Location.Column, 0, "Location column should be set")
}

// TestAnalyzeMeasures_ExampleFilesIntegration tests that example files are correctly analyzed.
func TestAnalyzeMeasures_ExampleFilesIntegration(t *testing.T) {
	t.Parallel()

	examplesDir := getParserExamplesDir(t)
	withoutErrorsDir := filepath.Join(examplesDir, "parse_and_analyze_only", "without_errors")

	entries, err := os.ReadDir(withoutErrorsDir)
	require.NoError(t, err, "Failed to read without_errors directory")

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cpl") {
			continue
		}

		// Only process files that might contain measures
		content, err := os.ReadFile(filepath.Join(withoutErrorsDir, entry.Name()))
		if err != nil {
			continue
		}

		if !strings.Contains(string(content), "measures") {
			continue
		}

		entry := entry // capture range variable
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(withoutErrorsDir, entry.Name())

			var defs *analysis.AnalysisDefinitions
			var err error

			if entry.Name() == "interval_declarations.cpl" {
				// interval_declarations.cpl uses Instant/Interval constraints that
				// require standard time concepts (Hour, Minute, Second, DayOfWeek,
				// Month, etc.). Prepend standard_concepts.cpl so they are available.
				stdDefsPath := filepath.Join(getParserExamplesDir(t), "..", "standard_definitions", "standard_concepts.cpl")
				config := &parser.ParseLoggingConfig{
					LogLevel:   slog.LevelInfo,
					OutputFile: "",
				}
				doc, parseErr := parser.ParseFiles([]string{stdDefsPath, filePath}, config)
				require.NoError(t, parseErr, "Expected no parse error for %s", entry.Name())
				defs, err = analysis.Analyze(doc)
			} else {
				// Read and parse
				content, readErr := os.ReadFile(filePath)
				require.NoError(t, readErr, "Should be able to read file")
				defs, err = parseAndAnalyze(string(content))
			}
			require.NoError(t, err, "Expected no error for %s", entry.Name())
			require.NotNil(t, defs, "AnalysisDefinitions should not be nil for %s", entry.Name())

			// If the file contains measures, verify measures were analyzed
			// (empty measures map is valid if there are no measures sections)
		})
	}
}

// TestAnalyzeMeasureScope_WithSimpleUnitExpression tests a simple unit reference expression.
func TestAnalyzeMeasureScope_WithSimpleUnitExpression(t *testing.T) {
	t.Parallel()
	// Need a units section so BuildExpressionTree can resolve the unit
	input := "concepts\n\tLength = [Values[m]]\nunits\n\tLength = Length\n\nmeasures\n\tSpeed = Number(unit=Length)"
	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)

	measure, exists := defs.Measures["Speed"]
	require.True(t, exists)
	assert.Equal(t, "Length", measure.GetUnitName())
	require.NotNil(t, measure.UnitExpression)
	ref, ok := measure.UnitExpression.(*analysis.UnitReference)
	require.True(t, ok, "Expected UnitReference, got %T", measure.UnitExpression)
	assert.Equal(t, "Length", ref.Name)
}

// TestAnalyzeMeasureScope_WithDivisionUnitExpression tests a division unit expression.
func TestAnalyzeMeasureScope_WithDivisionUnitExpression(t *testing.T) {
	t.Parallel()
	input := "concepts\n\tLength = [Values[m]]\n\tTime = [Values[s]]\nunits\n\tLength = Length\n\tTime = Time\n\nmeasures\n\tSpeed = Number(unit=Length/Time)"
	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)

	measure, exists := defs.Measures["Speed"]
	require.True(t, exists)
	assert.Equal(t, "", measure.GetUnitName()) // compound expressions return ""
	require.NotNil(t, measure.UnitExpression)
	_, ok := measure.UnitExpression.(*analysis.DivisionNode)
	require.True(t, ok, "Expected DivisionNode, got %T", measure.UnitExpression)
}

// TestAnalyzeMeasureScope_WithPowerUnitExpression tests a power unit expression.
func TestAnalyzeMeasureScope_WithPowerUnitExpression(t *testing.T) {
	t.Parallel()
	input := "concepts\n\tLength = [Values[m]]\n\tTime = [Values[s]]\nunits\n\tLength = Length\n\tTime = Time\n\nmeasures\n\tAccel = Number(unit=Length/Time^2)"
	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)

	measure, exists := defs.Measures["Accel"]
	require.True(t, exists)
	div, ok := measure.UnitExpression.(*analysis.DivisionNode)
	require.True(t, ok, "Expected DivisionNode")
	pow, ok := div.Right.(*analysis.PowerNode)
	require.True(t, ok, "Expected PowerNode as right child")
	assert.Equal(t, 2, pow.Exponent)
}

// TestAnalyzeMeasureScope_WithMultiplyDivideExpression tests multiplication and division combined.
func TestAnalyzeMeasureScope_WithMultiplyDivideExpression(t *testing.T) {
	t.Parallel()
	input := "concepts\n\tMass = [Values[kg]]\n\tLength = [Values[m]]\n\tTime = [Values[s]]\nunits\n\tMass = Mass\n\tLength = Length\n\tTime = Time\n\nmeasures\n\tMomentum = Number(unit=Mass*Length/Time)"
	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)

	measure, exists := defs.Measures["Momentum"]
	require.True(t, exists)
	require.NotNil(t, measure.UnitExpression)
	div, ok := measure.UnitExpression.(*analysis.DivisionNode)
	require.True(t, ok, "Expected DivisionNode, got %T", measure.UnitExpression)
	_, ok = div.Left.(*analysis.MultiplicationNode)
	require.True(t, ok, "Expected MultiplicationNode as left child of DivisionNode")
}

// TestVerifyMeasureUnitReferences_UndefinedUnit tests that undefined units are caught.
func TestVerifyMeasureUnitReferences_UndefinedUnit(t *testing.T) {
	t.Parallel()
	input := "concepts\n\tMeter = [Values[m]]\nunits\n\tMeter = Meter\n\nmeasures\n\tSpeed = Number(unit=BadUnit)"
	_, err := parseAndAnalyze(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "BadUnit")
}

// TestAnalyzeMeasures_CompoundUnits_Simple tests a simple unit reference.
func TestAnalyzeMeasures_CompoundUnits_Simple(t *testing.T) {
	t.Parallel()
	input := "concepts\n\tLength = [Values[Meter]]\n\tSpeed = [Values[Fast]]\n\nunits\n\tLength = Length\n\ntypes\n\tNumber = Number()\n\nmeasures\n\tSpeed = Number(unit=Length)"
	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	measure := defs.Measures["Speed"]
	assert.Equal(t, "Length", measure.GetUnitName())
	require.NotNil(t, measure.UnitExpression)
	_, ok := measure.UnitExpression.(*analysis.UnitReference)
	assert.True(t, ok, "Expected UnitReference for simple unit")
}

// TestAnalyzeMeasures_CompoundUnits_Division tests a division unit expression.
func TestAnalyzeMeasures_CompoundUnits_Division(t *testing.T) {
	t.Parallel()
	input := "concepts\n\tLength = [Values[Meter]]\n\tTime = [Values[Second]]\n\tSpeed = [Values[Fast]]\n\nunits\n\tLength = Length\n\tTime = Time\n\ntypes\n\tNumber = Number()\n\nmeasures\n\tSpeed = Number(unit=Length/Time)"
	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	measure := defs.Measures["Speed"]
	require.NotNil(t, measure.UnitExpression)
	_, ok := measure.UnitExpression.(*analysis.DivisionNode)
	assert.True(t, ok, "Expected DivisionNode for Length/Time")
}

// TestAnalyzeMeasures_CompoundUnits_Power tests a power unit expression.
func TestAnalyzeMeasures_CompoundUnits_Power(t *testing.T) {
	t.Parallel()
	input := "concepts\n\tLength = [Values[Meter]]\n\tTime = [Values[Second]]\n\tAccel = [Values[Gravity]]\n\nunits\n\tLength = Length\n\tTime = Time\n\ntypes\n\tNumber = Number()\n\nmeasures\n\tAcceleration = Number(unit=Length/Time^2)"
	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	measure := defs.Measures["Acceleration"]
	require.NotNil(t, measure.UnitExpression)
	div, ok := measure.UnitExpression.(*analysis.DivisionNode)
	require.True(t, ok, "Expected DivisionNode")
	pow, ok := div.Right.(*analysis.PowerNode)
	require.True(t, ok, "Expected PowerNode as right child of DivisionNode")
	assert.Equal(t, 2, pow.Exponent)
}

// TestAnalyzeMeasures_CompoundUnits_MultiplyDivide tests multiplication and division combined.
func TestAnalyzeMeasures_CompoundUnits_MultiplyDivide(t *testing.T) {
	t.Parallel()
	input := "concepts\n\tMass = [Values[Kilogram]]\n\tLength = [Values[Meter]]\n\tTime = [Values[Second]]\n\tMomentum = [Values[Default]]\n\nunits\n\tMass = Mass\n\tLength = Length\n\tTime = Time\n\ntypes\n\tNumber = Number()\n\nmeasures\n\tMomentum = Number(unit=Mass*Length/Time)"
	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	measure := defs.Measures["Momentum"]
	require.NotNil(t, measure.UnitExpression)
	div, ok := measure.UnitExpression.(*analysis.DivisionNode)
	require.True(t, ok, "Expected DivisionNode")
	_, ok = div.Left.(*analysis.MultiplicationNode)
	require.True(t, ok, "Expected MultiplicationNode as left child of DivisionNode")
}

// TestAnalyzeMeasures_CompoundUnits_Parentheses tests parenthesized power expression.
func TestAnalyzeMeasures_CompoundUnits_Parentheses(t *testing.T) {
	t.Parallel()
	input := "concepts\n\tMass = [Values[Kilogram]]\n\tLength = [Values[Meter]]\n\tDensity = [Values[Default]]\n\nunits\n\tMass = Mass\n\tLength = Length\n\ntypes\n\tNumber = Number()\n\nmeasures\n\tAreaDensity = Number(unit=Mass/(Length^2))"
	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	measure := defs.Measures["AreaDensity"]
	require.NotNil(t, measure.UnitExpression)
	div, ok := measure.UnitExpression.(*analysis.DivisionNode)
	require.True(t, ok, "Expected DivisionNode")
	pow, ok := div.Right.(*analysis.PowerNode)
	require.True(t, ok, "Expected PowerNode as right child of DivisionNode")
	ref, ok := pow.Base.(*analysis.UnitReference)
	require.True(t, ok, "Expected UnitReference as base of PowerNode")
	assert.Equal(t, "Length", ref.Name)
	assert.Equal(t, 2, pow.Exponent)
}
