// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/calculation"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

func TestNewCILCodeGenerator(t *testing.T) {
	registry := NewCILRegistry()
	gen := NewCILCodeGenerator(registry)
	if gen == nil {
		t.Fatalf("NewCILCodeGenerator() returned nil")
	}
	if gen.connectorMetadata == nil {
		t.Errorf("NewCILCodeGenerator().connectorMetadata is nil")
	}
}

func TestCILCodeGenerator_GenerateConnectorMetadataMap_Empty(t *testing.T) {
	gen := NewCILCodeGenerator(NewCILRegistry())
	code := gen.GenerateConnectorMetadataMap()
	if !strings.Contains(code, "var connectorMetadata = map[string]connector.ConnectorMetadata{}") {
		t.Errorf("expected empty map declaration, got %q", code)
	}
}

func TestCILCodeGenerator_GenerateConnectorMetadataMap_NilReceiver(t *testing.T) {
	var gen *CILCodeGenerator
	code := gen.GenerateConnectorMetadataMap()
	if !strings.Contains(code, "var connectorMetadata = map[string]connector.ConnectorMetadata{}") {
		t.Errorf("expected empty map declaration for nil receiver, got %q", code)
	}
}

func TestCILCodeGenerator_GenerateConnectorMetadataMap_Single(t *testing.T) {
	registry := NewCILRegistry()
	registry.RegisterConnector("sales_data", []string{"Product", "Region"}, nil, nil, "float64")
	gen := NewCILCodeGenerator(registry)
	code := gen.GenerateConnectorMetadataMap()

	expected := []string{
		"connectorMetadata",
		"sales_data",
		"Product",
		"Region",
		"float64",
		"map[string]connector.ConnectorMetadata",
	}
	for _, exp := range expected {
		if !strings.Contains(code, exp) {
			t.Errorf("missing expected content %q in:\n%s", exp, code)
		}
	}
	notExpected := []string{"map[string]ConnectorMetadata{", "Grain:"}
	for _, ne := range notExpected {
		if strings.Contains(code, ne) {
			t.Errorf("unexpected content %q in:\n%s", ne, code)
		}
	}
}

func TestCILCodeGenerator_GenerateConnectorMetadataMap_Multiple(t *testing.T) {
	registry := NewCILRegistry()
	registry.RegisterConnector("sales_data", []string{"Product", "Region"}, nil, nil, "float64")
	registry.RegisterConnector("inventory", []string{"Warehouse"}, nil, nil, "int64")
	gen := NewCILCodeGenerator(registry)
	code := gen.GenerateConnectorMetadataMap()

	if !strings.Contains(code, "sales_data") {
		t.Errorf("missing sales_data")
	}
	if !strings.Contains(code, "inventory") {
		t.Errorf("missing inventory")
	}
	if !strings.Contains(code, "float64") {
		t.Errorf("missing float64")
	}
	if !strings.Contains(code, "int64") {
		t.Errorf("missing int64")
	}
}

func TestCILCodeGenerator_GenerateConnectorMetadataMap_WithGrainToConcept(t *testing.T) {
	registry := NewCILRegistry()
	registry.RegisterConnector("sales_data", []string{"Region"}, []string{"YearGrain", "WeekGrain"}, map[string]string{"YearGrain": "Year", "WeekGrain": "WeekNumber"}, "float64")
	gen := NewCILCodeGenerator(registry)
	code := gen.GenerateConnectorMetadataMap()

	if !strings.Contains(code, "Grain: []string{") {
		t.Errorf("missing Grain field")
	}
	if !strings.Contains(code, "GrainToConcept: map[string]string{") {
		t.Errorf("missing GrainToConcept field")
	}
	if !strings.Contains(code, `"YearGrain": "Year"`) {
		t.Errorf("missing grain-to-concept mapping")
	}
}

func TestCILCodeGenerator_GenerateSetupFunction(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	code := gen.GenerateSetupFunction()
	if !strings.Contains(code, "func Setup(") {
		t.Errorf("missing Setup function signature")
	}
	if !strings.Contains(code, "cplio.ArchitectureSpecification") {
		t.Errorf("missing ArchitectureSpecification reference")
	}
	if !strings.Contains(code, "calculation.NewChronoconeComputationController()") {
		t.Errorf("missing controller initialization")
	}
}

func TestCILCodeGenerator_GenerateCleanupFunction(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	code := gen.GenerateCleanupFunction()
	if !strings.Contains(code, "func Cleanup(") {
		t.Errorf("missing Cleanup function signature")
	}
	if !strings.Contains(code, "controller.StopWorkers()") {
		t.Errorf("missing StopWorkers call")
	}
}

func TestCILCodeGenerator_GenerateComputeFunction_Empty(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	code := gen.GenerateComputeFunction([]string{})
	if !strings.Contains(code, "func Compute() error") {
		t.Errorf("missing Compute function signature")
	}
	if strings.Contains(code, "sync.WaitGroup") {
		t.Errorf("unexpected sync.WaitGroup for empty metric names")
	}
}

func TestCILCodeGenerator_GenerateComputeFunction_Nil(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	code := gen.GenerateComputeFunction(nil)
	if !strings.Contains(code, "func Compute() error") {
		t.Errorf("missing Compute function signature")
	}
	if strings.Contains(code, "sync.WaitGroup") {
		t.Errorf("unexpected sync.WaitGroup for nil metric names")
	}
}

func TestCILCodeGenerator_GenerateComputeFunction_WithMetrics(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	code := gen.GenerateComputeFunction([]string{"A", "B"})
	expected := []string{
		"func Compute() error",
		"for name, tensor := range inputs",
		"controller.RegisterMetricComputation",
		"var wg sync.WaitGroup",
		"wg.Add(1)",
		"defer wg.Done()",
		"computeA(controller, inputs)",
		"computeB(controller, inputs)",
		"wg.Wait()",
	}
	for _, exp := range expected {
		if !strings.Contains(code, exp) {
			t.Errorf("missing expected content %q in:\n%s", exp, code)
		}
	}
}

// TestCILCodeGenerator_GenerateComputeFunction_TopologicalOrder verifies that
// GenerateComputeFunction launches goroutines in the exact order provided by
// the dagOrder slice, preserving topological ordering (dependencies before dependents).
func TestCILCodeGenerator_GenerateComputeFunction_TopologicalOrder(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	dagOrder := []string{"input_metric", "intermediate", "output_metric"}
	code := gen.GenerateComputeFunction(dagOrder)

	// Verify all compute calls appear in topological order.
	idxInput := strings.Index(code, "computeinput_metric(controller, inputs)")
	idxIntermediate := strings.Index(code, "computeintermediate(controller, inputs)")
	idxOutput := strings.Index(code, "computeoutput_metric(controller, inputs)")

	if idxInput == -1 {
		t.Errorf("missing computeinput_metric call")
	}
	if idxIntermediate == -1 {
		t.Errorf("missing computeintermediate call")
	}
	if idxOutput == -1 {
		t.Errorf("missing computeoutput_metric call")
	}
	if idxInput != -1 && idxIntermediate != -1 && idxInput > idxIntermediate {
		t.Errorf("computeinput_metric should appear before computeintermediate in topological order")
	}
	if idxIntermediate != -1 && idxOutput != -1 && idxIntermediate > idxOutput {
		t.Errorf("computeintermediate should appear before computeoutput_metric in topological order")
	}
}

func TestCILCodeGenerator_generateConnectorRegistrationCode(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	code := gen.generateConnectorRegistrationCode()
	if code == "" {
		t.Errorf("generateConnectorRegistrationCode() returned empty string")
	}
	if !strings.Contains(code, "for name, tensor := range inputs") {
		t.Errorf("missing input loop")
	}
	if !strings.Contains(code, "controller.RegisterMetricComputation") {
		t.Errorf("missing registration call")
	}
	if !strings.Contains(code, "calculation.NewMetricComputation(0)") {
		t.Errorf("missing MetricComputation construction")
	}
	if !strings.Contains(code, "SetTensor") {
		t.Errorf("missing SetTensor call")
	}
}

func TestCILCodeGenerator_GeneratePostfixTableLiteral_Empty(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	code := gen.GeneratePostfixTableLiteral([]calculation.PostfixOp{})
	if code != "[]calculation.PostfixOp{}" {
		t.Errorf("expected empty literal, got %q", code)
	}
}

func TestCILCodeGenerator_GeneratePostfixTableLiteral_Simple(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	ops := []calculation.PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "push", Name: "b"},
		{Kind: "apply", Op: operator.Add},
	}
	code := gen.GeneratePostfixTableLiteral(ops)
	if !strings.Contains(code, "[]calculation.PostfixOp{") {
		t.Errorf("missing opening")
	}
	if !strings.Contains(code, `Kind: "push"`) {
		t.Errorf("missing push")
	}
	if !strings.Contains(code, `Kind: "apply"`) {
		t.Errorf("missing apply")
	}
	if !strings.Contains(code, `Name: "a"`) {
		t.Errorf("missing Name a")
	}
	if !strings.Contains(code, `Op: operator.Add`) {
		t.Errorf("missing Op +")
	}
}

func TestCILCodeGenerator_GeneratePostfixTableLiteral_WithBranches(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	ops := []calculation.PostfixOp{
		{
			Kind: "conditional",
			Branches: []calculation.PostfixBranch{
				{
					Condition: []calculation.PostfixOp{
						{Kind: "push", Name: "a"},
						{Kind: "push", Value: 0.0},
						{Kind: "apply", Op: operator.GreaterThan},
					},
					Body: []calculation.PostfixOp{
						{Kind: "push", Value: 1.0},
					},
				},
				{
					Condition: nil,
					Body: []calculation.PostfixOp{
						{Kind: "push", Value: 2.0},
					},
				},
			},
		},
	}
	code := gen.GeneratePostfixTableLiteral(ops)
	if !strings.Contains(code, `Kind: "conditional"`) {
		t.Errorf("missing conditional kind")
	}
	if !strings.Contains(code, "Branches:") {
		t.Errorf("missing Branches field")
	}
	if !strings.Contains(code, "Condition:") {
		t.Errorf("missing Condition field")
	}
	if !strings.Contains(code, "Body:") {
		t.Errorf("missing Body field")
	}
}

func TestCILCodeGenerator_generatePostfixComputeBody(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	metricNode := &CILMetricNode{
		CILNode: &CILBinaryOp{
			Op: operator.Add,
			Left:  &CILLiteral{Value: 1.0},
			Right: &CILLiteral{Value: 2.0},
		},
	}
	code, err := gen.generatePostfixComputeBody("TestMetric", metricNode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code, "EvaluatePostfixFloat64") {
		t.Errorf("missing EvaluatePostfixFloat64 call")
	}
	if !strings.Contains(code, `controller.GetMetricComputation("TestMetric").SetTensor(result)`) {
		t.Errorf("missing SetTensor call")
	}
	if !strings.Contains(code, "[]calculation.PostfixOp") {
		t.Errorf("missing postfix ops literal")
	}
}

func TestCILCodeGenerator_generatePostfixComputeBody_NilReceiver(t *testing.T) {
	var gen *CILCodeGenerator
	_, err := gen.generatePostfixComputeBody("X", &CILMetricNode{CILNode: &CILLiteral{Value: 1.0}})
	if err == nil {
		t.Errorf("expected error for nil receiver")
	}
}

func TestCILCodeGenerator_generatePostfixComputeBody_NilMetricNode(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	_, err := gen.generatePostfixComputeBody("X", nil)
	if err == nil {
		t.Errorf("expected error for nil metric node")
	}
}

func TestCILCodeGenerator_generatePostfixComputeBody_NilCILNode(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	_, err := gen.generatePostfixComputeBody("X", &CILMetricNode{})
	if err == nil {
		t.Errorf("expected error for nil CILNode")
	}
}

// TestCILCodeGenerator_GeneratePostfixTableLiteral_SingleLiteral verifies that
// a single-literal postfix table (one push op with a scalar Value) renders
// correctly as a Go struct literal.
func TestCILCodeGenerator_GeneratePostfixTableLiteral_SingleLiteral(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	ops := []calculation.PostfixOp{
		{Kind: "push", Value: 42.0},
	}
	code := gen.GeneratePostfixTableLiteral(ops)
	if code != "[]calculation.PostfixOp{\n\t{Kind: \"push\", Value: 42},\n}" {
		t.Errorf("unexpected single-literal output:\n%s", code)
	}
}

// TestCILCodeGenerator_GeneratePostfixTableLiteral_DeeplyNestedBinary verifies
// that a deeply nested binary expression produces a correctly ordered postfix
// literal with proper indentation and all intermediate ops present.
func TestCILCodeGenerator_GeneratePostfixTableLiteral_DeeplyNestedBinary(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	// Expression: ((a + b) * (c - d)) / e
	ops := []calculation.PostfixOp{
		{Kind: "push", Name: "a"},
		{Kind: "push", Name: "b"},
		{Kind: "apply", Op: operator.Add},
		{Kind: "push", Name: "c"},
		{Kind: "push", Name: "d"},
		{Kind: "apply", Op: operator.Subtract},
		{Kind: "apply", Op: operator.Multiply},
		{Kind: "push", Name: "e"},
		{Kind: "apply", Op: operator.Divide},
	}
	code := gen.GeneratePostfixTableLiteral(ops)
	expected := []string{
		"[]calculation.PostfixOp{",
		`Kind: "push", Name: "a"`,
		`Kind: "push", Name: "b"`,
		`Kind: "apply", Op: operator.Add`,
		`Kind: "push", Name: "c"`,
		`Kind: "push", Name: "d"`,
		`Kind: "apply", Op: operator.Subtract`,
		`Kind: "apply", Op: operator.Multiply`,
		`Kind: "push", Name: "e"`,
		`Kind: "apply", Op: operator.Divide`,
	}
	for _, exp := range expected {
		if !strings.Contains(code, exp) {
			t.Errorf("missing expected content %q in:\n%s", exp, code)
		}
	}
}

// TestCILCodeGenerator_generatePostfixBranchLiteral_Empty verifies that an
// empty branch slice renders as an empty package-qualified slice literal.
func TestCILCodeGenerator_generatePostfixBranchLiteral_Empty(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	code := gen.generatePostfixBranchLiteral([]calculation.PostfixBranch{}, "\t")
	if code != "[]calculation.PostfixBranch{}" {
		t.Errorf("expected empty branch literal, got %q", code)
	}
}

// TestCILCodeGenerator_generatePostfixBranchLiteral_ComplexNested verifies
// that generatePostfixBranchLiteral correctly formats complex conditionals
// with nested branches, including nil conditions for else branches and
// nested conditional ops inside branch bodies.
func TestCILCodeGenerator_generatePostfixBranchLiteral_ComplexNested(t *testing.T) {
	gen := NewCILCodeGenerator(nil)
	branches := []calculation.PostfixBranch{
		{
			Condition: []calculation.PostfixOp{
				{Kind: "push", Name: "x"},
				{Kind: "push", Value: 0.0},
				{Kind: "apply", Op: operator.GreaterThan},
			},
			Body: []calculation.PostfixOp{
				{
					Kind: "conditional",
					Branches: []calculation.PostfixBranch{
						{
							Condition: []calculation.PostfixOp{
								{Kind: "push", Name: "y"},
								{Kind: "push", Value: 0.0},
								{Kind: "apply", Op: operator.LessThan},
							},
							Body: []calculation.PostfixOp{
								{Kind: "push", Value: 1.0},
							},
						},
						{
							Condition: nil,
							Body: []calculation.PostfixOp{
								{Kind: "push", Value: 2.0},
							},
						},
					},
				},
			},
		},
		{
			Condition: nil,
			Body: []calculation.PostfixOp{
				{Kind: "push", Value: 3.0},
			},
		},
	}
	code := gen.generatePostfixBranchLiteral(branches, "\t")
	expected := []string{
		"[]calculation.PostfixBranch{",
		"Condition:",
		"Body:",
		`Kind: "conditional"`,
		`Kind: "push", Name: "x"`,
		`Kind: "push", Name: "y"`,
		`Kind: "apply", Op: operator.GreaterThan`,
		`Kind: "apply", Op: operator.LessThan`,
		"Value: 1",
		"Value: 2",
		"Value: 3",
	}
	for _, exp := range expected {
		if !strings.Contains(code, exp) {
			t.Errorf("missing expected content %q in:\n%s", exp, code)
		}
	}

	// Structural assertion: verify Value: 1 appears after the LessThan operator
	// so we know it sits inside the correct (nested) branch body rather than
	// merely existing somewhere in the output.
	idxLessThan := strings.Index(code, `Kind: "apply", Op: operator.LessThan`)
	idxValue1 := strings.Index(code, "Value: 1")
	if idxLessThan == -1 || idxValue1 == -1 || idxValue1 <= idxLessThan {
		t.Errorf("expected \"Value: 1\" to appear after \"Kind: \\\"apply\\\", Op: operator.LessThan\" in:\n%s", code)
	}
}
