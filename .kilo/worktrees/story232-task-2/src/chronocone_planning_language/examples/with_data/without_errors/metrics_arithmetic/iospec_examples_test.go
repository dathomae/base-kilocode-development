// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package metrics_arithmetic

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/io"
)

// TestIOSpecExamples verifies that each arithmetic example IO specification JSON
// file can be loaded successfully and contains the expected output metric name
// and correct connector inputs.
func TestIOSpecExamples(t *testing.T) {
	tests := []struct {
		filename           string
		expectedMetric     string
		expectedInputCnt   int
		expectedConnectors []string
		expectedCellType   string
		expectedGrain      []string
	}{
		{
			filename:           SimpleAdditionIOSpecJSON,
			expectedMetric:     "sum_metric",
			expectedInputCnt:   2,
			expectedConnectors: []string{"input_a", "input_b"},
			expectedCellType:   "packages",
			expectedGrain:      []string{"Daily"},
		},
		{
			filename:           MultiplicationPrecedenceIOSpecJSON,
			expectedMetric:     "precedence_metric",
			expectedInputCnt:   3,
			expectedConnectors: []string{"input_a", "input_b", "input_c"},
			expectedCellType:   "packages",
			expectedGrain:      []string{"Daily"},
		},
		{
			filename:           SubtractionDivisionIOSpecJSON,
			expectedMetric:     "sub_div_metric",
			expectedInputCnt:   3,
			expectedConnectors: []string{"input_a", "input_b", "input_c"},
			expectedCellType:   "packages",
			expectedGrain:      []string{"Daily"},
		},
		{
			filename:           UnaryNegationIOSpecJSON,
			expectedMetric:     "negate_metric",
			expectedInputCnt:   2,
			expectedConnectors: []string{"input_a", "input_b"},
			expectedCellType:   "packages",
			expectedGrain:      []string{"Daily"},
		},
		{
			filename:           ParenthesizedIOSpecJSON,
			expectedMetric:     "paren_metric",
			expectedInputCnt:   4,
			expectedConnectors: []string{"input_a", "input_b", "input_c", "input_d"},
			expectedCellType:   "packages",
			expectedGrain:      []string{"Daily"},
		},
		{
			filename:           ComplexExpressionIOSpecJSON,
			expectedMetric:     "complex_metric",
			expectedInputCnt:   5,
			expectedConnectors: []string{"input_a", "input_b", "input_c", "input_d", "input_e"},
			expectedCellType:   "packages",
			expectedGrain:      []string{"Daily"},
		},
		{
			filename:           ShortFormAdditionIOSpecJSON,
			expectedMetric:     "sum_metric",
			expectedInputCnt:   2,
			expectedConnectors: []string{"input_a", "input_b"},
			expectedCellType:   "packages",
			expectedGrain:      []string{"Daily"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			spec, err := io.LoadIOSpecification(tt.filename)
			if err != nil {
				t.Fatalf("LoadIOSpecification(%q) returned error: %v", tt.filename, err)
			}

			if spec == nil {
				t.Fatalf("LoadIOSpecification(%q) returned nil", tt.filename)
			}

			if len(spec.Connectors) != tt.expectedInputCnt {
				t.Errorf("expected %d inputs, got %d", tt.expectedInputCnt, len(spec.Connectors))
			}

			// Verify connector names
			for i, expectedConn := range tt.expectedConnectors {
				if i >= len(spec.Connectors) {
					t.Errorf("missing input connector %q at index %d", expectedConn, i)
					continue
				}
				if spec.Connectors[i].ConnectorName != expectedConn {
					t.Errorf("input[%d]: expected connector %q, got %q",
						i, expectedConn, spec.Connectors[i].ConnectorName)
				}
				if spec.Connectors[i].DataFormat != "parquet" {
					t.Errorf("input[%d] %q: expected data_format 'parquet', got %q",
						i, expectedConn, spec.Connectors[i].DataFormat)
				}
				if spec.Connectors[i].Driver != "file" {
					t.Errorf("input[%d] %q: expected driver 'file', got %q",
						i, expectedConn, spec.Connectors[i].Driver)
				}
			}

			if len(spec.Outputs) != 1 {
				t.Fatalf("expected 1 output, got %d", len(spec.Outputs))
			}

			out := spec.Outputs[0]
			if out.MetricName != tt.expectedMetric {
				t.Errorf("expected metric %q, got %q", tt.expectedMetric, out.MetricName)
			}

			if len(out.Dimensions) != 1 || out.Dimensions[0] != "Region" {
				t.Errorf("expected dimensions [Region], got %v", out.Dimensions)
			}

			if out.CellType != tt.expectedCellType {
				t.Errorf("expected celltype %q, got %q", tt.expectedCellType, out.CellType)
			}

			if len(out.Grain) != len(tt.expectedGrain) {
				t.Errorf("expected grain %v, got %v", tt.expectedGrain, out.Grain)
			} else {
				for i, g := range tt.expectedGrain {
					if out.Grain[i] != g {
						t.Errorf("expected grain %v, got %v", tt.expectedGrain, out.Grain)
						break
					}
				}
			}

			if spec.Architecture.CPUCoreCount != 4 {
				t.Errorf("expected cpu_cores=4, got %d", spec.Architecture.CPUCoreCount)
			}

			if spec.Architecture.CacheSizeBytes != 8388608 {
				t.Errorf("expected cache_size=8388608, got %d", spec.Architecture.CacheSizeBytes)
			}

			if spec.Architecture.MainMemoryBytes != 8589934592 {
				t.Errorf("expected main_memory_size=8589934592, got %d", spec.Architecture.MainMemoryBytes)
			}
		})
	}
}
