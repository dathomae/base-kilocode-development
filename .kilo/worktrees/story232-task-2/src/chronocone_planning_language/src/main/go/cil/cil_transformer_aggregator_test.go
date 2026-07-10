// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// moduleRoot returns the absolute path to the chronocone_planning_language module root.
func moduleRoot() string {
	_, currentFile, _, _ := runtime.Caller(0)
	// This file is at src/main/go/cil/cil_transformer_aggregator_test.go
	// Module root is four levels up from this file's directory.
	return filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "..")
}

func TestTransformMetric_AggregatorNode(t *testing.T) {
	tests := []struct {
		name              string
		metricName        string
		root              interface{}
		expectContains    []string
		expectNotContains []string
		expectError       bool
		expectErrorContains string
	}{
		{
			name:       "sum from [Country, Product] to [Product]",
			metricName: "AggSales",
			root: &CILMetricNode{
				CILNode: &CILAggregateNode{
					SourceCIL:      &CILInputRef{Kind: "connector", Name: "sales_data"},
					FromDimensions: []string{"Country", "Product"},
					ToDimension:    "Product",
				},
				InputRefs: []CILInputRef{
					{Kind: "connector", Name: "sales_data"},
				},
			},
			expectContains: []string{
				"func computeAggSales(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"ops := []calculation.PostfixOp",
				`Kind: "push"`,
				`Name: "sales_data"`,
				"sourceResult, err := calculation.EvaluatePostfixFloat64(controller, ops)",
				"check(err, \"Failed to evaluate source for aggregation\")",
				"sourceData := sourceResult.Data()",
				"sourceFloat64 := make([]float64, len(sourceData))",
				"sourceFloat64[i] = v.(float64)",
				"sourceTensor := calculation.NewChronoconeTensor(sourceFloat64, sourceResult.DimensionNames(), sourceResult.Shape())",
				"tensorops.ReduceDimension(sourceTensor, aggregationMatrix_Country, 0, 1)",
				"if err != nil {",
				`check(fmt.Errorf("aggregation failed: %w", err)`,
				"controller.GetMetricComputation(\"AggSales\").SetTensor(anyResult)",
				"check(err, \"Failed to set tensor for AggSales\")",
			},
			expectNotContains: []string{
				"// TODO:",
				"unsupported CIL node type",
			},
		},
		{
			name:       "scalar aggregation from [Country] to []",
			metricName: "TotalSales",
			root: &CILMetricNode{
				CILNode: &CILAggregateNode{
					SourceCIL:      &CILLiteral{Value: 42.0},
					FromDimensions: []string{"Country"},
					ToDimension:    "",
				},
			},
			expectContains: []string{
				"func computeTotalSales(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any])",
				"tensorops.ReduceDimension(sourceTensor, aggregationMatrix_Country, 0, -1)",
				"_ = inputs",
			},
		},
		{
			name:       "collapse middle dimension [A, B, C] to [A, C]",
			metricName: "MidCollapse",
			root: &CILMetricNode{
				CILNode: &CILAggregateNode{
					SourceCIL:      &CILBinaryOp{Op: operator.Add, Left: &CILInputRef{Kind: "metric", Name: "m1"}, Right: &CILInputRef{Kind: "metric", Name: "m2"}},
					FromDimensions: []string{"A", "B", "C"},
					ToDimension:    "C",
				},
				InputRefs: []CILInputRef{
					{Kind: "metric", Name: "m1"},
					{Kind: "metric", Name: "m2"},
				},
			},
			expectContains: []string{
				"aggregationMatrix_A",
				"tensorops.ReduceDimension(sourceTensor, aggregationMatrix_A, 0, 2)",
				`Op: operator.Add`,
			},
		},
		{
			name:                "nil aggregate node",
			metricName:          "AnyMetric",
			root:                &CILMetricNode{CILNode: (*CILAggregateNode)(nil)},
			expectError:         true,
			expectErrorContains: "aggregate node is nil",
		},
		{
			name:       "nil SourceCIL",
			metricName: "NoSource",
			root: &CILMetricNode{
				CILNode: &CILAggregateNode{
					SourceCIL:      nil,
					FromDimensions: []string{"Country"},
					ToDimension:    "",
				},
			},
			expectError:         true,
			expectErrorContains: "aggregate node SourceCIL is nil",
		},
		{
			name:       "invalid dimensions no collapsed dim",
			metricName: "BadDims",
			root: &CILMetricNode{
				CILNode: &CILAggregateNode{
					SourceCIL:      &CILLiteral{Value: 1.0},
					FromDimensions: []string{"Country"},
					ToDimension:    "Country",
				},
			},
			expectError:         true,
			expectErrorContains: "could not determine collapsed dimension",
		},
		{
			name:       "unsupported CILAggregateNode in postfix walk",
			metricName: "AnyMetric",
			root: &CILMetricNode{
				CILNode: &CILAggregateNode{
					SourceCIL:      &CILLiteral{Value: 1.0},
					FromDimensions: []string{"D"},
					ToDimension:    "",
				},
			},
			// When generating through TransformMetric, the aggregate node IS handled.
			// This test case verifies that walkToPostfix itself rejects CILAggregateNode.
			expectContains: []string{
				"func computeAnyMetric",
				"tensorops.ReduceDimension",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := NewCILAbstractTransformer()
			if transformer == nil {
				t.Fatalf("NewCILAbstractTransformer() returned nil")
			}

			code, err := transformer.TransformMetric(tt.metricName, tt.root)

			if tt.expectError {
				if err == nil {
					t.Errorf("TransformMetric() returned nil error, expected error")
					return
				}
				if tt.expectErrorContains != "" && !strings.Contains(err.Error(), tt.expectErrorContains) {
					t.Errorf("TransformMetric() error = %q, expected to contain %q", err.Error(), tt.expectErrorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("TransformMetric() returned unexpected error: %v", err)
				return
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(code, expected) {
					t.Errorf("TransformMetric() output missing expected content: %q\nGot:\n%s", expected, code)
				}
			}

			for _, notExpected := range tt.expectNotContains {
				if strings.Contains(code, notExpected) {
					t.Errorf("TransformMetric() output contains unexpected content: %q\nGot:\n%s", notExpected, code)
				}
			}
		})
	}
}

// TestTransformMetric_AggregateNode_Compiles verifies that the generated
// compute function for a CILAggregateNode compiles as part of a complete
// Go program.
func TestTransformMetric_AggregatorNode_Compiles(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("'go' command not available in PATH")
	}

	transformer := NewCILAbstractTransformer()
	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}

	metricName := "AggSales"
	root := &CILMetricNode{
		CILNode: &CILAggregateNode{
			SourceCIL:      &CILInputRef{Kind: "connector", Name: "sales_data"},
			FromDimensions: []string{"Country", "Product"},
			ToDimension:    "Product",
		},
		InputRefs: []CILInputRef{
			{Kind: "connector", Name: "sales_data"},
		},
	}

	generatedCode, err := transformer.TransformMetric(metricName, root)
	if err != nil {
		t.Fatalf("TransformMetric() returned unexpected error: %v", err)
	}

	// Wrap the generated compute function in a minimal compilable Go program.
	program := fmt.Sprintf(`package main

import (
	"fmt"
	"os"

	"github.com/chronocone/chronocone_planning_language/src/main/go/calculation"
	"github.com/chronocone/chronocone_planning_language/src/main/go/tensorops"
)

func check(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%%s: %%v\n", msg, err)
		os.Exit(1)
	}
}

var aggregationMatrix_Country = &tensorops.AggregationMatrix{
	Rows:      2,
	Columns:   1,
	Data:      [][]float64{{1}, {1}},
	RowLabels: []string{"USA", "Canada"},
	ColLabels: []string{"Total"},
}

%s

func main() {
	controller := calculation.NewChronoconeComputationController()
	inputs := map[string]*calculation.ChronoconeTensor[any]{}

	// Register the source metric with a 2D tensor.
	mc := calculation.NewMetricComputation(0)
	if err := controller.RegisterMetricComputation("sales_data", mc); err != nil {
		check(err, "Failed to register sales_data")
	}
	srcData := make([]any, 4)
	for i := range srcData {
		srcData[i] = float64(i + 1)
	}
	srcTensor := calculation.NewChronoconeTensor(srcData, []string{"Country", "Product"}, []int{2, 2})
	if err := mc.SetTensor(srcTensor); err != nil {
		check(err, "Failed to set tensor for sales_data")
	}

	// Register the output metric computation.
	mc2 := calculation.NewMetricComputation(0)
	if err := controller.RegisterMetricComputation("AggSales", mc2); err != nil {
		check(err, "Failed to register AggSales")
	}

	computeAggSales(controller, inputs)
}
`, generatedCode)

	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(srcPath, []byte(program), 0644); err != nil {
		t.Fatalf("failed to write test program: %v", err)
	}

	cmd := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "testbin"), srcPath)
	cmd.Dir = moduleRoot()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build generated program: %v\nOutput:\n%s", err, string(out))
	}
}

// TestWalkCILTreeToPostfix_AggregateNodeError verifies that walkToPostfix
// returns an explicit error when encountering a CILAggregateNode.
func TestWalkCILTreeToPostfix_AggregatorNodeError(t *testing.T) {
	agg := &CILAggregateNode{
		SourceCIL:      &CILLiteral{Value: 1.0},
		FromDimensions: []string{"Country"},
		ToDimension:    "",
	}

	_, err := WalkCILTreeToPostfix(agg)
	if err == nil {
		t.Fatalf("WalkCILTreeToPostfix(CILAggregateNode) returned nil error, expected error")
	}

	expectedMsg := "CILAggregateNode must be handled by TransformMetric"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("WalkCILTreeToPostfix() error = %q, expected to contain %q", err.Error(), expectedMsg)
	}
}
