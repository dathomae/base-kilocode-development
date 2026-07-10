// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package synthesis

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	cio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// buildMetricCILTree creates AnalysisDefinitions with a single metric
// whose CILExpression is the given root node.
func buildMetricCILTree(metricName string, rootExpr cil.CILNode) *analysis.AnalysisDefinitions {
	return &analysis.AnalysisDefinitions{
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			metricName: {
				Name: metricName,
				CILRoot: &cil.CILMetricNode{
					CILNode:   rootExpr,
					InputRefs: nil,
				},
			},
		},
	}
}

// buildSynthesisInput is a test helper that returns a SynthesisInput with the
// given conversion pairs, populated from BuildTransformerFromAnalysis.
func buildSynthesisInput(analysisDefs *analysis.AnalysisDefinitions) *SynthesisInput {
	input, _ := BuildTransformerFromAnalysis(analysisDefs)
	if input == nil {
		input = &SynthesisInput{}
	}
	return input
}

// TestUnitChecking_ValidPipeline_ConversionBridgesGap verifies that when a
// metric computes Kilometer and the IO spec expects Meter, a Kilometer->Meter
// conversion allows validation to pass.
func TestUnitChecking_ValidPipeline_ConversionBridgesGap(t *testing.T) {
	// Set up Kilometer->Meter conversion
	kmDef := analysis.AnalysisUnitDefinition{Name: "Kilometer"}
	mDef := analysis.AnalysisUnitDefinition{Name: "Meter"}
	kmDef.Conversions = []analysis.ConversionPair[analysis.AnalysisUnitDefinition]{
		{
			SourceDef:   &kmDef,
			TargetDef:   &mDef,
			SourceVar:   "km",
			TargetVar:   "m",
			Calculation: &analysis.NumberNode{Value: 1000},
		},
	}

	analysisDefs := buildMetricCILTree("distance", &cil.CILUnitRef{UnitGroup: "Kilometer"})
	analysisDefs.Units = map[string]analysis.AnalysisUnitDefinition{
		"Kilometer": kmDef,
	}

	input := buildSynthesisInput(analysisDefs)

	ioSpec := &cio.IOSpecification{
		Outputs: []cio.OutputSpecification{{MetricName: "distance", Unit: "Meter"}},
	}

	err := validateOutputUnits(input, ioSpec)
	if err != nil {
		t.Fatalf("validateOutputUnits failed unexpectedly: %v", err)
	}
}

// TestUnitChecking_MissingConversion_ReturnsError verifies that when a metric
// computes EUR and the IO spec expects USD, validation returns an error when
// no EUR->USD conversion exists.
func TestUnitChecking_MissingConversion_ReturnsError(t *testing.T) {
	// Set up a conversion that is NOT EUR->USD
	celsiusDef := analysis.AnalysisUnitDefinition{Name: "Celsius"}
	fahrenheitDef := analysis.AnalysisUnitDefinition{Name: "Fahrenheit"}
	celsiusDef.Conversions = []analysis.ConversionPair[analysis.AnalysisUnitDefinition]{
		{
			SourceDef:   &celsiusDef,
			TargetDef:   &fahrenheitDef,
			SourceVar:   "c",
			TargetVar:   "f",
			Calculation: &analysis.NumberNode{Value: 1.8},
		},
	}

	analysisDefs := buildMetricCILTree("revenue", &cil.CILUnitRef{UnitGroup: "EUR"})
	analysisDefs.Units = map[string]analysis.AnalysisUnitDefinition{
		"Celsius": celsiusDef,
	}

	input := buildSynthesisInput(analysisDefs)

	ioSpec := &cio.IOSpecification{
		Outputs: []cio.OutputSpecification{{MetricName: "revenue", Unit: "USD"}},
	}

	err := validateOutputUnits(input, ioSpec)
	if err == nil {
		t.Fatal("Expected validateOutputUnits to return error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "unit check failed") {
		t.Errorf("Error missing 'unit check failed': %v", err)
	}
	if !strings.Contains(msg, "EUR") {
		t.Errorf("Error missing 'EUR': %v", err)
	}
	if !strings.Contains(msg, "USD") {
		t.Errorf("Error missing 'USD': %v", err)
	}
}

// TestUnitChecking_IncompatibleAddition_ReturnsError verifies that adding
// incompatible units (USD + EUR) causes a unit check error even when
// conversions are provided.
func TestUnitChecking_IncompatibleAddition_ReturnsError(t *testing.T) {
	// Provide some conversions (doesn't matter — addition itself should fail)
	usdDef := analysis.AnalysisUnitDefinition{Name: "USD"}
	eurDef := analysis.AnalysisUnitDefinition{Name: "EUR"}
	usdDef.Conversions = []analysis.ConversionPair[analysis.AnalysisUnitDefinition]{
		{
			SourceDef:   &usdDef,
			TargetDef:   &eurDef,
			SourceVar:   "usd",
			TargetVar:   "eur",
			Calculation: &analysis.NumberNode{Value: 0.9},
		},
	}

	metricRoot := &cil.CILBinaryOp{
		Op:    operator.Add,
		Left:  &cil.CILUnitRef{UnitGroup: "USD"},
		Right: &cil.CILUnitRef{UnitGroup: "EUR"},
	}

	analysisDefs := buildMetricCILTree("total", metricRoot)
	analysisDefs.Units = map[string]analysis.AnalysisUnitDefinition{
		"USD": usdDef,
	}

	input := buildSynthesisInput(analysisDefs)

	ioSpec := &cio.IOSpecification{
		Outputs: []cio.OutputSpecification{{MetricName: "total", Unit: "USD"}},
	}

	err := validateOutputUnits(input, ioSpec)
	if err == nil {
		t.Fatal("Expected validateOutputUnits to return error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "incompatible") && !strings.Contains(msg, "differ") {
		t.Errorf("Error missing 'incompatible' or 'differ': %v", err)
	}
}

// TestUnitChecking_DimensionlessPassThrough verifies that a dimensionless
// computation with an empty Unit field succeeds.
func TestUnitChecking_DimensionlessPassThrough(t *testing.T) {
	analysisDefs := buildMetricCILTree("count", &cil.CILNumber{Value: 42.0})

	input := buildSynthesisInput(analysisDefs)

	ioSpec := &cio.IOSpecification{
		Outputs: []cio.OutputSpecification{{MetricName: "count"}},
	}

	err := validateOutputUnits(input, ioSpec)
	if err != nil {
		t.Fatalf("validateOutputUnits failed unexpectedly: %v", err)
	}
}

// TestUnitChecking_EmptyUnitFieldPasses verifies that an explicitly empty
// Unit field ("" ) on the output spec passes for a simple computation.
func TestUnitChecking_EmptyUnitFieldPasses(t *testing.T) {
	analysisDefs := buildMetricCILTree("ratio", &cil.CILNumber{Value: 3.14})

	input := buildSynthesisInput(analysisDefs)

	ioSpec := &cio.IOSpecification{
		Outputs: []cio.OutputSpecification{{MetricName: "ratio", Unit: ""}},
	}

	err := validateOutputUnits(input, ioSpec)
	if err != nil {
		t.Fatalf("validateOutputUnits failed unexpectedly: %v", err)
	}
}

// TestUnitChecking_Regression_NoConversions_ExistingMetricsPass verifies that
// a basic metric with a simple CIL tree compiles successfully through Synthesize
// when there are NO unit or measure conversions defined. This ensures existing
// programs without unit conversions continue to compile.
func TestUnitChecking_Regression_NoConversions_ExistingMetricsPass(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.go")

	metricCILRoot := &cil.CILLiteral{Value: 42}

	analysisDefs := buildMetricCILTree("Revenue", metricCILRoot)
	// Ensure all maps are explicitly empty (not nil) to reflect a clean program
	// that has never declared units, measures, concepts, types, or connectors.
	analysisDefs.Concepts = map[string]analysis.AnalysisConceptDefinition{}
	analysisDefs.Types = map[string]analysis.AnalysisTypeDefinition{}
	analysisDefs.Connectors = map[string]analysis.AnalysisConnectorDefinition{}
	analysisDefs.Units = map[string]analysis.AnalysisUnitDefinition{}
	analysisDefs.Measures = map[string]analysis.AnalysisMeasureDefinition{}

	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "Revenue",
				Dimensions: []string{},
				CellType:   "Number",
				Grain:      []string{},
				Unit:       "", // empty = dimensionless / no unit constraint
			},
		},
		Architecture: cio.ArchitectureSpecification{},
	}

	err := Synthesize(analysisDefs, ioSpec, outputPath)
	if err != nil {
		t.Errorf("Synthesize failed for regression case (no conversions): %v", err)
	}
}
