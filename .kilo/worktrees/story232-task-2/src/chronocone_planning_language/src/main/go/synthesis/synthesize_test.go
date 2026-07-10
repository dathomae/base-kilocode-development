package synthesis

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/connector"
	cio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
)

// TestSynthesizeCreatesOutputFile verifies that the Synthesize function
// creates an output file at the specified path with a full computation
// lifecycle program.
func TestSynthesizeCreatesOutputFile(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.go")

	ioSpec := &cio.IOSpecification{
		Connectors:       []cio.ConnectorSpecification{},
		Outputs:      []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{},
	}

	err := Synthesize(nil, ioSpec, outputPath)
	if err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package main") {
		t.Errorf("Generated file missing 'package main'")
	}
	if !strings.Contains(contentStr, "func main()") {
		t.Errorf("Generated file missing 'func main()'")
	}
	if !strings.Contains(contentStr, "func check(err error, msg string)") {
		t.Errorf("Generated file missing 'check' helper function")
	}
}

// TestSynthesizeWithAnalysisDefinitions verifies that the Synthesize function
// correctly accepts and uses AnalysisDefinitions when generating the CLI.
func TestSynthesizeWithAnalysisDefinitions(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.go")

	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Country": {ConceptName: "Country"},
			"Status":  {ConceptName: "Status"},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors:       []cio.ConnectorSpecification{},
		Outputs:      []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{},
	}

	err := Synthesize(analysisDefs, ioSpec, outputPath)
	if err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package main") {
		t.Errorf("Generated file missing 'package main'")
	}
	if !strings.Contains(contentStr, "func main()") {
		t.Errorf("Generated file missing 'func main()'")
	}
}

// TestSynthesizeHandlesFileErrors verifies that the Synthesize function
// properly handles and returns errors when file operations fail.
func TestSynthesizeHandlesFileErrors(t *testing.T) {
	invalidPath := "/nonexistent/directory/output.go"

	ioSpec := &cio.IOSpecification{
		Connectors:       []cio.ConnectorSpecification{},
		Outputs:      []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{},
	}

	err := Synthesize(nil, ioSpec, invalidPath)
	if err == nil {
		t.Errorf("Expected error for invalid path, but got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "failed to create output file") {
		t.Errorf("Error message missing expected context: %v", err)
	}
}

// TestGenerateMainFunction verifies that GenerateMainFunction generates valid Go main function.
func TestGenerateMainFunction(t *testing.T) {
	tests := []struct {
		name           string
		ioSpec         *cio.IOSpecification
		connectors     map[string]*connector.ConnectorMetadata
		expectContains []string
		expectError    bool
	}{
		{
			name: "basic main function",
			ioSpec: &cio.IOSpecification{
				Connectors: []cio.ConnectorSpecification{
					{
						ConnectorName: "sales_data",
						DataFormat:    "parquet",
						Driver:        "file",
					},
				},
				Outputs: []cio.OutputSpecification{
					{
						MetricName: "total_sales",
						Dimensions: []string{"Region"},
						CellType:   "float64",
						Grain:      []string{"Region"},
					},
				},
				Architecture: cio.ArchitectureSpecification{
					CPUCoreCount:    4,
					CacheSizeBytes:  1024 * 1024,
					MainMemoryBytes: 1024 * 1024 * 1024,
				},
			},
			connectors: map[string]*connector.ConnectorMetadata{
				"sales_data": {
					Name:       "sales_data",
					Dimensions: []string{"Product", "Region"},
					CellType:   "float64",
				},
			},
			expectContains: []string{
				"func main()",
				"package main",
				"func main",
				"Setup",
				"Compute",
				"Cleanup",
			},
			expectError: false,
		},
		{
			name: "main with multiple inputs",
			ioSpec: &cio.IOSpecification{
				Connectors: []cio.ConnectorSpecification{
					{
						ConnectorName: "sales_data",
						DataFormat:    "parquet",
						Driver:        "file",
					},
					{
						ConnectorName: "inventory",
						DataFormat:    "parquet",
						Driver:        "file",
					},
				},
				Outputs: []cio.OutputSpecification{
					{
						MetricName: "stock_status",
						Dimensions: []string{"Product"},
						CellType:   "int64",
						Grain:      []string{"Product"},
					},
				},
				Architecture: cio.ArchitectureSpecification{
					CPUCoreCount:    8,
					CacheSizeBytes:  2 * 1024 * 1024,
					MainMemoryBytes: 2 * 1024 * 1024 * 1024,
				},
			},
			connectors: map[string]*connector.ConnectorMetadata{
				"sales_data": {
					Name:       "sales_data",
					Dimensions: []string{"Product", "Region"},
					CellType:   "float64",
				},
				"inventory": {
					Name:       "inventory",
					Dimensions: []string{"Warehouse", "Product"},
					CellType:   "int64",
				},
			},
			expectContains: []string{
				"func main()",
				"sales_data",
				"inventory",
			},
			expectError: false,
		},
		{
			name:           "nil io spec",
			ioSpec:         nil,
			connectors:     map[string]*connector.ConnectorMetadata{},
			expectContains: []string{},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := cil.NewCILAbstractTransformer()

			if transformer == nil {
				t.Fatalf("NewCILAbstractTransformer() returned nil")
			}

			// Register connectors
			for _, metadata := range tt.connectors {
				transformer.RegisterConnector(metadata.Name, metadata.Dimensions, nil, nil, metadata.CellType)
			}

			code, err := GenerateMainFunction(&SynthesisInput{Transformer: transformer, IOSpec: tt.ioSpec})

			if tt.expectError {
				if err == nil {
					t.Errorf("GenerateMainFunction() returned nil error, expected error")
				}
			} else {
				if err != nil {
					t.Errorf("GenerateMainFunction() returned error: %v", err)
				}

				for _, expected := range tt.expectContains {
					if !strings.Contains(code, expected) {
						t.Errorf("GenerateMainFunction() output missing expected content: %q", expected)
					}
				}
			}
		})
	}
}

// TestGenerateMainFunctionImports verifies that generated code includes proper imports.
// The "sync" import is conditional — it is only included when the IO spec has metric
// outputs that require sync.WaitGroup for goroutine synchronization.
func TestGenerateMainFunctionImports(t *testing.T) {
	tests := []struct {
		name          string
		ioSpec        *cio.IOSpecification
		expectSync    bool
		expectNotSync bool
	}{
		{
			name: "no outputs omits sync import",
			ioSpec: &cio.IOSpecification{
				Connectors:  []cio.ConnectorSpecification{},
				Outputs: []cio.OutputSpecification{},
				Architecture: cio.ArchitectureSpecification{
					CPUCoreCount:    4,
					CacheSizeBytes:  1024 * 1024,
					MainMemoryBytes: 1024 * 1024 * 1024,
				},
			},
			expectSync:    false,
			expectNotSync: true,
		},
		{
			name: "with outputs includes sync import",
			ioSpec: &cio.IOSpecification{
				Connectors: []cio.ConnectorSpecification{},
				Outputs: []cio.OutputSpecification{
					{
						MetricName: "total_sales",
						Dimensions: []string{"Region"},
						CellType:   "float64",
						Grain:      []string{"Region"},
					},
				},
				Architecture: cio.ArchitectureSpecification{
					CPUCoreCount:    4,
					CacheSizeBytes:  1024 * 1024,
					MainMemoryBytes: 1024 * 1024 * 1024,
				},
			},
			expectSync:    true,
			expectNotSync: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := cil.NewCILAbstractTransformer()

			if transformer == nil {
				t.Fatalf("NewCILAbstractTransformer() returned nil")
			}

			code, err := GenerateMainFunction(&SynthesisInput{Transformer: transformer, IOSpec: tt.ioSpec})

			if err != nil {
				t.Fatalf("GenerateMainFunction() returned error: %v", err)
			}

			// Verify import statement
			if !strings.Contains(code, "import") {
				t.Errorf("GenerateMainFunction() output missing 'import' statement")
			}

			// Always-required imports
			alwaysRequired := []string{
				"\"fmt\"",
				"\"os\"",
				"\"github.com/chronocone/chronocone_planning_language/src/main/go/calculation\"",
				"\"github.com/chronocone/chronocone_planning_language/src/main/go/connector\"",
				"cplio \"github.com/chronocone/chronocone_planning_language/src/main/go/io\"",
			}

			for _, imp := range alwaysRequired {
				if !strings.Contains(code, imp) {
					t.Errorf("GenerateMainFunction() output missing required import: %q", imp)
				}
			}

			// Conditional sync import
			if tt.expectSync && !strings.Contains(code, "\"sync\"") {
				t.Errorf("GenerateMainFunction() output missing conditional 'sync' import when outputs are present")
			}
			if tt.expectNotSync && strings.Contains(code, "\"sync\"") {
				t.Errorf("GenerateMainFunction() output contains 'sync' import when no outputs are present (would cause compile error)")
			}
		})
	}
}

// TestGenerateMainFunctionQualifiedNames verifies that generated code uses
// package-qualified type names and does not use unqualified type references.
func TestGenerateMainFunctionQualifiedNames(t *testing.T) {
	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "sales_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs: []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	transformer := cil.NewCILAbstractTransformer()
	transformer.RegisterConnector("sales_data", []string{"Product", "Region"}, nil, nil, "float64")

	code, err := GenerateMainFunction(&SynthesisInput{Transformer: transformer, IOSpec: ioSpec})
	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	// Verify package-qualified type references are present
	qualifiedChecks := []string{
		"calculation.ChronoconeComputationController",
		"calculation.NewChronoconeComputationController()",
		"cplio.ArchitectureSpecification",
		"cplio.ConnectorSpecification",
		"connector.ConnectorMetadata",
	}

	for _, expected := range qualifiedChecks {
		if !strings.Contains(code, expected) {
			t.Errorf("GenerateMainFunction() output missing qualified reference: %q", expected)
		}
	}

	// Verify unqualified type references are NOT present (when used as types)
	unqualifiedChecks := []string{
		"(architecture *ArchitectureSpecification)",
		" := ArchitectureSpecification{",
		" := ConnectorSpecification{",
	}

	for _, unexpected := range unqualifiedChecks {
		if strings.Contains(code, unexpected) {
			t.Errorf("GenerateMainFunction() output contains unqualified type reference %q", unexpected)
		}
	}
}

// TestGenerateMainFunctionDispatchesToConnectorLoaders verifies dispatch to connector loaders.
func TestGenerateMainFunctionDispatchesToConnectorLoaders(t *testing.T) {
	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "sales_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
			{
				ConnectorName: "inventory",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "total_sales",
				Dimensions: []string{"Region"},
				CellType:   "float64",
				Grain:      []string{"Region"},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	transformer := cil.NewCILAbstractTransformer()

	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}

	// Register connectors
	transformer.RegisterConnector("sales_data", []string{"Product", "Region"}, nil, nil, "float64")
	transformer.RegisterConnector("inventory", []string{"Warehouse", "Product"}, nil, nil, "int64")

	code, err := GenerateMainFunction(&SynthesisInput{Transformer: transformer, IOSpec: ioSpec})

	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	// Verify connector names are referenced
	if !strings.Contains(code, "sales_data") {
		t.Errorf("GenerateMainFunction() output missing 'sales_data' reference")
	}

	if !strings.Contains(code, "inventory") {
		t.Errorf("GenerateMainFunction() output missing 'inventory' reference")
	}

	// Verify loader function names might be present (implementation dependent)
	// These might be in the generated code depending on implementation
}

// TestGenerateMainFunctionWithOutputWriting verifies output writing code generation.
func TestGenerateMainFunctionWithOutputWriting(t *testing.T) {
	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "total_sales",
				Dimensions: []string{"Region"},
				CellType:   "float64",
				Grain:      []string{"Region"},
			},
			{
				MetricName: "avg_price",
				Dimensions: []string{"Product"},
				CellType:   "float64",
				Grain:      []string{"Product"},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	transformer := cil.NewCILAbstractTransformer()

	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}

	code, err := GenerateMainFunction(&SynthesisInput{Transformer: transformer, IOSpec: ioSpec})

	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	// Verify output metrics are referenced
	if !strings.Contains(code, "total_sales") {
		t.Errorf("GenerateMainFunction() output missing 'total_sales' reference")
	}

	if !strings.Contains(code, "avg_price") {
		t.Errorf("GenerateMainFunction() output missing 'avg_price' reference")
	}

	// Verify cplio.WriteMetricsToParquet call is present
	if !strings.Contains(code, "cplio.WriteMetricsToParquet") {
		t.Errorf("GenerateMainFunction() output missing 'cplio.WriteMetricsToParquet' reference")
	}

	// Verify single output parquet file path
	if !strings.Contains(code, "total_sales.parquet") {
		t.Errorf("GenerateMainFunction() output missing 'total_sales.parquet' reference")
	}

	// Verify cplio.WriteMetricsToParquet call is present
	if !strings.Contains(code, "cplio.WriteMetricsToParquet") {
		t.Errorf("GenerateMainFunction() output missing 'cplio.WriteMetricsToParquet' reference")
	}

	// Verify old stdout placeholder is removed
	if strings.Contains(code, "fmt.Printf(\"Writing output:") {
		t.Errorf("GenerateMainFunction() output still contains old stdout placeholder")
	}

	// Verify metric-specific variable names are present
	if !strings.Contains(code, "total_salesComp") {
		t.Errorf("GenerateMainFunction() output missing 'total_salesComp' variable")
	}

	if !strings.Contains(code, "avg_priceTensor") {
		t.Errorf("GenerateMainFunction() output missing 'avg_priceTensor' variable")
	}

	// Verify error handling with os.Stderr and check() helper
	if !strings.Contains(code, "os.Stderr") {
		t.Errorf("GenerateMainFunction() output missing 'os.Stderr' reference")
	}

	if !strings.Contains(code, "check(err, \"") {
		t.Errorf("GenerateMainFunction() output missing 'check(err, ...)' error handling call")
	}
}

// TestGenerateOutputWritingCode_NoStdoutPlaceholder verifies that when outputs
// are present, the generated code does NOT contain fmt.Printf placeholder
// statements for metric writing.
func TestGenerateOutputWritingCode_NoStdoutPlaceholder(t *testing.T) {
	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "total_sales",
				Dimensions: []string{"Region"},
				CellType:   "float64",
				Grain:      []string{"Region"},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	transformer := cil.NewCILAbstractTransformer()

	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}

	code, err := GenerateMainFunction(&SynthesisInput{Transformer: transformer, IOSpec: ioSpec})

	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	// Verify no stdout placeholder for output writing
	if strings.Contains(code, "fmt.Printf(\"Writing output:") {
		t.Errorf("Generated code contains stdout placeholder 'fmt.Printf(\"Writing output:\")'")
	}

	// Verify real parquet writing code is present instead
	if !strings.Contains(code, "cplio.WriteMetricsToParquet") {
		t.Errorf("Generated code missing 'cplio.WriteMetricsToParquet'")
	}
}

// TestGenerateMainFunctionArchitectureConfig verifies architecture configuration in generated code.
func TestGenerateMainFunctionArchitectureConfig(t *testing.T) {
	ioSpec := &cio.IOSpecification{
		Connectors:  []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    8,
			CacheSizeBytes:  16 * 1024 * 1024,
			MainMemoryBytes: 32 * 1024 * 1024 * 1024,
		},
	}

	transformer := cil.NewCILAbstractTransformer()

	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}

	code, err := GenerateMainFunction(&SynthesisInput{Transformer: transformer, IOSpec: ioSpec})

	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	// Verify architecture-related code is present
	if !strings.Contains(code, "Architecture") && !strings.Contains(code, "architecture") {
		t.Errorf("GenerateMainFunction() output missing architecture configuration")
	}
}

// TestGenerateMainFunctionErrorHandling verifies error handling in generated code.
func TestGenerateMainFunctionErrorHandling(t *testing.T) {
	ioSpec := &cio.IOSpecification{
		Connectors:  []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	transformer := cil.NewCILAbstractTransformer()

	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}

	code, err := GenerateMainFunction(&SynthesisInput{Transformer: transformer, IOSpec: ioSpec})

	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	// Verify error handling is present
	if !strings.Contains(code, "error") {
		t.Errorf("GenerateMainFunction() output missing error handling")
	}
}

// TestGenerateMainFunction_CheckHelper verifies that the generated program
// contains exactly one check() function definition and that all error handling
// sites call check() instead of using inline fmt.Fprintf + os.Exit patterns.
//
// This test exercises a scenario with inputs, outputs, and metrics to ensure
// all code generation sites (main, input loading, output writing, postfix
// compute) use the check helper.
func TestGenerateMainFunction_CheckHelper(t *testing.T) {
	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "test_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "test_metric",
				Dimensions: []string{"Region"},
				CellType:   "float64",
				Grain:      []string{"Region"},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    2,
			CacheSizeBytes:  1024,
			MainMemoryBytes: 4096,
		},
	}

	transformer := cil.NewCILAbstractTransformer()
	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}
	transformer.RegisterConnector("test_data", []string{"Region"}, nil, nil, "float64")

	input := &SynthesisInput{Transformer: transformer, IOSpec: ioSpec}
	code, err := GenerateMainFunction(input)
	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	// 1. Verify exactly one check function definition
	checkFuncCount := countOccurrences(code, "func check(err error, msg string)")
	if checkFuncCount != 1 {
		t.Errorf("Generated code contains %d check function definitions, expected exactly 1", checkFuncCount)
	}

	// 2. Verify check is called at error sites (not just defined)
	//    check( should appear more than once (definition + at least one call site)
	checkCallCount := countOccurrences(code, "check(")
	if checkCallCount < 2 {
		t.Errorf("Generated code contains %d occurrences of 'check(', expected at least 2 (definition + call sites)", checkCallCount)
	}

	// 3. Verify no os.Exit(1) outside the check function.
	//    The check function body itself contains os.Exit(1).
	exitCount := countOccurrences(code, "os.Exit(1)")
	if exitCount != 1 {
		t.Errorf("Generated code contains %d occurrences of 'os.Exit(1)', expected exactly 1 (inside check function)", exitCount)
	}

	// 4. Verify the old exitCode pattern is removed
	if strings.Contains(code, "exitCode") {
		t.Errorf("Generated code still contains 'exitCode' — should use check() helper instead")
	}

}

// countOccurrences counts the number of non-overlapping occurrences of substr in s.
func countOccurrences(s, substr string) int {
	count := 0
	for i := 0; i <= len(s)-len(substr); {
		idx := strings.Index(s[i:], substr)
		if idx < 0 {
			break
		}
		count++
		i += idx + len(substr)
	}
	return count
}

// TestGenerateMainFunctionWithNoInputs verifies main generation with no inputs.
func TestGenerateMainFunctionWithNoInputs(t *testing.T) {
	ioSpec := &cio.IOSpecification{
		Connectors:  []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	transformer := cil.NewCILAbstractTransformer()

	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}

	code, err := GenerateMainFunction(&SynthesisInput{Transformer: transformer, IOSpec: ioSpec})

	if err != nil {
		t.Errorf("GenerateMainFunction() with no inputs returned error: %v", err)
	}

	// Verify main function is still generated
	if !strings.Contains(code, "func main()") {
		t.Errorf("GenerateMainFunction() output missing 'func main()' with no inputs")
	}
}

// TestGenerateMainFunctionWithNoOutputs verifies main generation with no outputs.
func TestGenerateMainFunctionWithNoOutputs(t *testing.T) {
	ioSpec := &cio.IOSpecification{
		Connectors:  []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	transformer := cil.NewCILAbstractTransformer()

	if transformer == nil {
		t.Fatalf("NewCILAbstractTransformer() returned nil")
	}

	code, err := GenerateMainFunction(&SynthesisInput{Transformer: transformer, IOSpec: ioSpec})

	if err != nil {
		t.Errorf("GenerateMainFunction() with no outputs returned error: %v", err)
	}

	// Verify main function is still generated
	if !strings.Contains(code, "func main()") {
		t.Errorf("GenerateMainFunction() output missing 'func main()' with no outputs")
	}
}

// TestBuildTransformerFromAnalysis verifies that BuildTransformerFromAnalysis
// correctly populates a CILAbstractTransformer from AnalysisDefinitions.
func TestBuildTransformerFromAnalysis(t *testing.T) {
	tests := []struct {
		name             string
		analysisDefs     *analysis.AnalysisDefinitions
		expectConnectors int
		expectMeasures   int
		expectConnector  string
		expectDimensions []string
		expectCellType   string
	}{
		{
			name:             "nil analysis defs returns empty transformer",
			analysisDefs:     nil,
			expectConnectors: 0,
			expectMeasures:   0,
		},
		{
			name: "connector and measure registration",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"Country": {ConceptName: "Country"},
					"Product": {ConceptName: "Product"},
				},
				Types: map[string]analysis.AnalysisTypeDefinition{
					"Currency": {
						Name:           "Currency",
						DerivationPath: []string{"Currency", "Number"},
					},
				},
				Measures: map[string]analysis.AnalysisMeasureDefinition{
					"Amount": {
						Name:     "Amount",
						TypeName: "Currency",
					},
				},
				Connectors: map[string]analysis.AnalysisConnectorDefinition{
					"SalesData": {
						Name: "SalesData",
						Dimensions: []*analysis.AnalysisConceptDefinition{
							{ConceptName: "Country"},
							{ConceptName: "Product"},
						},
						CellType: &analysis.AnalysisMeasureDefinition{
							Name:     "Amount",
							TypeName: "Currency",
						},
					},
				},
			},
			expectConnectors: 1,
			expectMeasures:   1,
			expectConnector:  "SalesData",
			expectDimensions: []string{"Country", "Product"},
			expectCellType:   "float64",
		},
		{
			name: "multiple connectors",
			analysisDefs: &analysis.AnalysisDefinitions{
				Connectors: map[string]analysis.AnalysisConnectorDefinition{
					"SalesData": {
						Name: "SalesData",
						Dimensions: []*analysis.AnalysisConceptDefinition{
							{ConceptName: "Region"},
						},
						CellType: &analysis.AnalysisMeasureDefinition{
							Name:     "Amount",
							TypeName: "Number",
						},
					},
					"Inventory": {
						Name: "Inventory",
						Dimensions: []*analysis.AnalysisConceptDefinition{
							{ConceptName: "Warehouse"},
							{ConceptName: "Product"},
						},
						CellType: &analysis.AnalysisMeasureDefinition{
							Name:     "Count",
							TypeName: "Number",
						},
					},
				},
			},
			expectConnectors: 2,
			expectMeasures:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := BuildTransformerFromAnalysis(tt.analysisDefs)
			if err != nil {
				t.Fatalf("BuildTransformerFromAnalysis() returned error: %v", err)
			}
			transformer := input.Transformer

			if transformer == nil {
				t.Fatal("BuildTransformerFromAnalysis() returned nil")
			}

			// Count registered connectors
			connectorCount := 0
			for _, name := range []string{"SalesData", "Inventory"} {
				if transformer.GetConnectorMetadata(name) != nil {
					connectorCount++
				}
			}
			// For the nil case, just verify no panics and empty transformer
			if tt.analysisDefs == nil {
				if connectorCount != 0 {
					t.Errorf("Expected 0 connectors for nil defs, got %d", connectorCount)
				}
				return
			}

			if tt.expectConnector != "" {
				meta := transformer.GetConnectorMetadata(tt.expectConnector)
				if meta == nil {
					t.Errorf("Expected connector %q to be registered", tt.expectConnector)
					return
				}
				if len(meta.Dimensions) != len(tt.expectDimensions) {
					t.Errorf("Expected dimensions %v, got %v", tt.expectDimensions, meta.Dimensions)
				}
				for i, dim := range tt.expectDimensions {
					if meta.Dimensions[i] != dim {
						t.Errorf("Expected dimension[%d]=%q, got %q", i, dim, meta.Dimensions[i])
					}
				}
				if meta.CellType != tt.expectCellType {
					t.Errorf("Expected cell type %q, got %q", tt.expectCellType, meta.CellType)
				}
			}
		})
	}
}

// TestBuildTransformerFromAnalysis_WithGrainToConcept verifies that
// BuildTransformerFromAnalysis populates GrainToConcept when connector
// grain measures have UnitName set.
func TestBuildTransformerFromAnalysis_WithGrainToConcept(t *testing.T) {
	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Year":   {ConceptName: "Year"},
			"Region": {ConceptName: "Region"},
		},
		Measures: map[string]analysis.AnalysisMeasureDefinition{
			"YearGrain": {
				Name:     "YearGrain",
				TypeName: "Interval",
				UnitExpression: &analysis.UnitReference{Name: "Year"},
			},
		},
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"SalesData": {
				Name: "SalesData",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Region"},
				},
				CellType: &analysis.AnalysisMeasureDefinition{
					Name:     "Amount",
					TypeName: "Number",
				},
				Grain: []*analysis.AnalysisMeasureDefinition{
					{
						Name:     "YearGrain",
						TypeName: "Interval",
						UnitExpression: &analysis.UnitReference{Name: "Year"},
					},
				},
			},
		},
	}

	input, err := BuildTransformerFromAnalysis(analysisDefs)
	if err != nil {
		t.Fatalf("BuildTransformerFromAnalysis() returned error: %v", err)
	}
	transformer := input.Transformer

	meta := transformer.GetConnectorMetadata("SalesData")
	if meta == nil {
		t.Fatalf("Expected connector %q to be registered", "SalesData")
	}

	if meta.GrainToConcept == nil {
		t.Fatalf("Expected GrainToConcept to be populated, got nil")
	}

	if meta.GrainToConcept["YearGrain"] != "Year" {
		t.Errorf("Expected GrainToConcept[%q] = %q, got %q", "YearGrain", "Year", meta.GrainToConcept["YearGrain"])
	}
}

// TestBuildTransformerFromAnalysis_Lookups verifies that BuildTransformerFromAnalysis
// copies lookups from analysis definitions and walks CILLookupNode entries to
// register lookup metadata in the transformer's CILRegistry.
func TestBuildTransformerFromAnalysis_Lookups(t *testing.T) {
	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Country": {ConceptName: "Country"},
			"Product": {ConceptName: "Product"},
		},
		Measures: map[string]analysis.AnalysisMeasureDefinition{
			"UnitsSold": {
				Name:     "UnitsSold",
				TypeName: "Number",
			},
		},
		Lookups: map[string]*analysis.LookupInfo{
			"CountryProductIndex": {
				Name: "CountryProductIndex",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Country"},
					{ConceptName: "Product"},
				},
				CelltypeMeasure: &analysis.AnalysisMeasureDefinition{
					Name:     "UnitsSold",
					TypeName: "Number",
				},
			},
		},
		CILTrees: map[string]cil.CILNode{
			"lookup:CountryProductIndex": &cil.CILLookupNode{
				LookupName: "CountryProductIndex",
				Dimensions: []string{"Country", "Product"},
				Celltype:   "UnitsSold",
			},
		},
	}

	input, err := BuildTransformerFromAnalysis(analysisDefs)
	if err != nil {
		t.Fatalf("BuildTransformerFromAnalysis() returned error: %v", err)
	}

	// Verify SynthesisInput carries lookups
	if input.Lookups == nil {
		t.Fatalf("input.Lookups is nil, expected populated")
	}
	if len(input.Lookups) != 1 {
		t.Errorf("input.Lookups has %d entries, expected 1", len(input.Lookups))
	}
	lookupInfo, ok := input.Lookups["CountryProductIndex"]
	if !ok {
		t.Errorf("input.Lookups missing %q", "CountryProductIndex")
	}
	if lookupInfo != nil && lookupInfo.Name != "CountryProductIndex" {
		t.Errorf("input.Lookups[%q].Name = %q, expected %q", "CountryProductIndex", lookupInfo.Name, "CountryProductIndex")
	}

	// Verify transformer registry has lookup metadata registered
	transformer := input.Transformer
	meta := transformer.GetLookupMetadata("CountryProductIndex")
	if meta == nil {
		t.Fatalf("GetLookupMetadata(%q) returned nil after registration", "CountryProductIndex")
	}
	if meta.Name != "CountryProductIndex" {
		t.Errorf("lookup metadata Name = %q, expected %q", meta.Name, "CountryProductIndex")
	}
	if len(meta.Dimensions) != 2 {
		t.Errorf("lookup metadata Dimensions length = %d, expected 2", len(meta.Dimensions))
	}
	if meta.Dimensions[0] != "Country" {
		t.Errorf("lookup metadata Dimensions[0] = %q, expected %q", meta.Dimensions[0], "Country")
	}
	if meta.Dimensions[1] != "Product" {
		t.Errorf("lookup metadata Dimensions[1] = %q, expected %q", meta.Dimensions[1], "Product")
	}
	if meta.CellType != "UnitsSold" {
		t.Errorf("lookup metadata CellType = %q, expected %q", meta.CellType, "UnitsSold")
	}

	// Verify AllLookupMetadata returns the map
	all := transformer.AllLookupMetadata()
	if all == nil {
		t.Fatalf("AllLookupMetadata() returned nil")
	}
	if len(all) != 1 {
		t.Errorf("AllLookupMetadata() returned %d entries, expected 1", len(all))
	}
}

// TestBuildTransformerFromAnalysis_Lookups_NilCILTrees verifies that
// BuildTransformerFromAnalysis handles nil CILTrees gracefully when lookups
// are present in the analysis definitions.
func TestBuildTransformerFromAnalysis_Lookups_NilCILTrees(t *testing.T) {
	analysisDefs := &analysis.AnalysisDefinitions{
		Lookups: map[string]*analysis.LookupInfo{
			"SimpleLookup": {
				Name: "SimpleLookup",
			},
		},
		CILTrees: nil,
	}

	input, err := BuildTransformerFromAnalysis(analysisDefs)
	if err != nil {
		t.Fatalf("BuildTransformerFromAnalysis() returned error: %v", err)
	}

	if input.Lookups == nil {
		t.Fatalf("input.Lookups is nil, expected populated")
	}
	if len(input.Lookups) != 1 {
		t.Errorf("input.Lookups has %d entries, expected 1", len(input.Lookups))
	}

	transformer := input.Transformer
	all := transformer.AllLookupMetadata()
	if len(all) != 0 {
		t.Errorf("AllLookupMetadata() returned %d entries, expected 0", len(all))
	}
}

// TestSynthesizeWithConnectors verifies that Synthesize generates a full
// computation lifecycle program when given analysis definitions with connectors.
func TestSynthesizeWithConnectors(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.go")

	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Country": {ConceptName: "Country"},
			"Product": {ConceptName: "Product"},
		},
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"sales_data": {
				Name: "sales_data",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Country"},
					{ConceptName: "Product"},
				},
				CellType: &analysis.AnalysisMeasureDefinition{
					Name:     "Amount",
					TypeName: "Number",
				},
			},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "sales_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "total_sales",
				Dimensions: []string{"Country"},
				CellType:   "float64",
				Grain:      []string{"Country"},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	err := Synthesize(analysisDefs, ioSpec, outputPath)
	if err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	expectedElements := []string{
		"package main",
		"func main()",
		"func Setup",
		"func Compute",
		"func Cleanup",
		"ArchitectureSpecification",
		"sales_data",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Synthesize output missing expected content: %q", expected)
		}
	}
}

// TestBuildConditionalConceptValuesMap verifies conditional parent→child value extraction.
func strPtr(s string) *string {
	return &s
}

func TestBuildConditionalConceptValuesMap(t *testing.T) {
	tests := []struct {
		name         string
		analysisDefs *analysis.AnalysisDefinitions
		expectMap    map[string]ConditionalConceptEntry
	}{
		{
			name:         "nil analysis defs",
			analysisDefs: nil,
			expectMap:    map[string]ConditionalConceptEntry{},
		},
		{
			name: "empty concepts",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{},
			},
			expectMap: map[string]ConditionalConceptEntry{},
		},
		{
			name: "concept with conditional values",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"DayInMonth": {
						ConceptName:       "DayInMonth",
						ParentConceptName: strPtr("Month"),
						ConditionalValues: map[string][]string{
							"Jan": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31"},
							"Feb": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28"},
						},
					},
				},
			},
			expectMap: map[string]ConditionalConceptEntry{
				"DayInMonth": {
					ParentConcept: "Month",
					Values: map[string][]string{
						"Jan": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31"},
						"Feb": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28"},
					},
				},
			},
		},
		{
			name: "multiple parent values with different child sets",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"State": {
						ConceptName:       "State",
						ParentConceptName: strPtr("Country"),
						ConditionalValues: map[string][]string{
							"USA":    {"NY", "CA", "TX"},
							"Canada": {"ON", "BC", "AB"},
						},
					},
				},
			},
			expectMap: map[string]ConditionalConceptEntry{
				"State": {
					ParentConcept: "Country",
					Values: map[string][]string{
						"USA":    {"NY", "CA", "TX"},
						"Canada": {"ON", "BC", "AB"},
					},
				},
			},
		},
		{
			name: "concept with empty conditional values should not appear",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"Month": {
						ConceptName:       "Month",
						ConditionalValues: map[string][]string{},
					},
				},
			},
			expectMap: map[string]ConditionalConceptEntry{},
		},
		{
			name: "concept without conditionals should not appear",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"NonConditional": {
						ConceptName: "NonConditional",
					},
				},
			},
			expectMap: map[string]ConditionalConceptEntry{},
		},
		{
			name: "conditional concept with nil parent concept name is skipped",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"Orphan": {
						ConceptName: "Orphan",
						ConditionalValues: map[string][]string{
							"A": {"1", "2"},
						},
					},
				},
			},
			expectMap: map[string]ConditionalConceptEntry{},
		},
		{
			name: "mixed concepts — only conditional ones included",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"DayInMonth": {
						ConceptName:       "DayInMonth",
						ParentConceptName: strPtr("Month"),
						ConditionalValues: map[string][]string{
							"Jan": {"1", "2", "3"},
							"Feb": {"1", "2"},
						},
					},
					"Month": {
						ConceptName:       "Month",
						ConditionalValues: map[string][]string{},
					},
					"NonConditional": {
						ConceptName: "NonConditional",
					},
				},
			},
			expectMap: map[string]ConditionalConceptEntry{
				"DayInMonth": {
					ParentConcept: "Month",
					Values: map[string][]string{
						"Jan": {"1", "2", "3"},
						"Feb": {"1", "2"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := BuildTransformerFromAnalysis(tt.analysisDefs)
			if err != nil {
				t.Fatalf("BuildTransformerFromAnalysis() returned error: %v", err)
			}
			result := input.BuildConditionalConceptValuesMap(nil)
			if len(result) != len(tt.expectMap) {
				t.Errorf("BuildConditionalConceptValuesMap() returned %d entries, expected %d", len(result), len(tt.expectMap))
			}
			for conceptName, expectedEntry := range tt.expectMap {
				actualEntry, ok := result[conceptName]
				if !ok {
					t.Errorf("BuildConditionalConceptValuesMap() missing key %q", conceptName)
					continue
				}
				if actualEntry.ParentConcept != expectedEntry.ParentConcept {
					t.Errorf("BuildConditionalConceptValuesMap()[%q].ParentConcept = %q, expected %q", conceptName, actualEntry.ParentConcept, expectedEntry.ParentConcept)
					continue
				}
				if len(actualEntry.Values) != len(expectedEntry.Values) {
					t.Errorf("BuildConditionalConceptValuesMap()[%q].Values has %d parent entries, expected %d", conceptName, len(actualEntry.Values), len(expectedEntry.Values))
					continue
				}
				for parentValue, expectedChildValues := range expectedEntry.Values {
					actualChildValues, ok := actualEntry.Values[parentValue]
					if !ok {
						t.Errorf("BuildConditionalConceptValuesMap()[%q].Values missing parent value %q", conceptName, parentValue)
						continue
					}
					if len(actualChildValues) != len(expectedChildValues) {
						t.Errorf("BuildConditionalConceptValuesMap()[%q].Values[%q] has %d child values, expected %d", conceptName, parentValue, len(actualChildValues), len(expectedChildValues))
						continue
					}
					for i, v := range expectedChildValues {
						if actualChildValues[i] != v {
							t.Errorf("BuildConditionalConceptValuesMap()[%q].Values[%q][%d] = %q, expected %q", conceptName, parentValue, i, actualChildValues[i], v)
						}
					}
				}
			}
			// Verify no unexpected keys
			for conceptName := range result {
				if _, ok := tt.expectMap[conceptName]; !ok {
					t.Errorf("BuildConditionalConceptValuesMap() contains unexpected key %q", conceptName)
				}
			}
		})
	}
}

// TestBuildConditionalConceptValuesMap_WarningForOrphanConcept verifies that
// when a concept has conditional values but no parent concept name, a warning
// is emitted via the provided logger.
func TestBuildConditionalConceptValuesMap_WarningForOrphanConcept(t *testing.T) {
	// A simple logger that captures output via Printf
	var warnings []string
	captureLog := &captureLogger{
		printf: func(format string, v ...interface{}) {
			warnings = append(warnings, fmt.Sprintf(format, v...))
		},
	}

	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Orphan": {
				ConceptName: "Orphan",
				ConditionalValues: map[string][]string{
					"A": {"1", "2"},
				},
			},
		},
	}

	input, err := BuildTransformerFromAnalysis(analysisDefs)
	if err != nil {
		t.Fatalf("BuildTransformerFromAnalysis() returned error: %v", err)
	}
	result := input.BuildConditionalConceptValuesMap(captureLog)
	if len(result) != 0 {
		t.Errorf("BuildConditionalConceptValuesMap() returned %d entries, expected 0", len(result))
	}

	if len(warnings) == 0 {
		t.Errorf("Expected a warning for orphan concept, but none was emitted")
	}
	if len(warnings) > 0 {
		hasExpected := false
		for _, w := range warnings {
			if strings.Contains(w, "Orphan") && strings.Contains(w, "no parent") {
				hasExpected = true
				break
			}
		}
		if !hasExpected {
			t.Errorf("Warning does not mention orphan concept: %v", warnings)
		}
	}
}

// captureLogger implements Logger by delegating to a configurable Printf callback.
type captureLogger struct {
	printf func(format string, v ...interface{})
}

func (c *captureLogger) Printf(format string, v ...interface{}) {
	if c != nil && c.printf != nil {
		c.printf(format, v...)
	}
}

// TestBuildConnectorValidityMasks verifies correctness of connector-level mask
// computation, covering connectors with conditional grain dimensions, connectors
// without conditionals, and mixed scenarios.
//
// NOTE: BuildConnectorValidityMasks collects grain measure names and passes them
// to ComputeValidityMask. For a mask to be produced, the grain measure names must
// be present in both conceptValues and conditionalConceptValues. In practice,
// BuildConceptValuesMap adds grain measure names to conceptValues when their
// UnitName references a concept with resolved values, so both names coexist.
func TestBuildConnectorValidityMasks(t *testing.T) {
	tests := []struct {
		name                     string
		analysisDefs             *analysis.AnalysisDefinitions
		conditionalConceptValues map[string]ConditionalConceptEntry
		conceptValues            map[string][]string
		expectMaskFor            string
		expectNoMaskFor          string
		expectMaskCount          int
		expectError              bool
		expectErrorContains      string
	}{
		{
			name:            "nil analysis defs returns empty map",
			analysisDefs:    nil,
			expectMaskCount: 0,
		},
		{
			name: "connector with no grain produces no mask",
			analysisDefs: &analysis.AnalysisDefinitions{
				Connectors: map[string]analysis.AnalysisConnectorDefinition{
					"no_grain": {
						Name: "no_grain",
					},
				},
			},
			expectMaskCount: 0,
		},
		{
			name: "connector with grain measures matching conditional concepts produces mask",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"DayInMonth": {
						ConceptName: "DayInMonth",
						ConditionalValues: map[string][]string{
							"Jan": {"1", "2", "3"},
							"Feb": {"1", "2"},
						},
					},
					"Month": {ConceptName: "Month"},
				},
				Measures: map[string]analysis.AnalysisMeasureDefinition{
					"Monthly":  {Name: "Monthly", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Month"}},
					"DailyDim": {Name: "DailyDim", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "DayInMonth"}},
				},
				Connectors: map[string]analysis.AnalysisConnectorDefinition{
					"mixed_grain": {
						Name: "mixed_grain",
						Grain: []*analysis.AnalysisMeasureDefinition{
							{Name: "Month", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Month"}},
							{Name: "DayInMonth", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "DayInMonth"}},
						},
					},
				},
			},
			conceptValues: map[string][]string{
				"Month":      {"Jan", "Feb"},
				"DayInMonth": {"1", "2", "3"},
			},
			conditionalConceptValues: map[string]ConditionalConceptEntry{
				"DayInMonth": {
					ParentConcept: "Month",
					Values: map[string][]string{
						"Jan": {"1", "2", "3"},
						"Feb": {"1", "2"},
					},
				},
			},
			expectMaskFor:   "mixed_grain",
			expectMaskCount: 1,
		},
		{
			name: "grain without conditional concepts produces no mask",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"Region": {ConceptName: "Region"},
				},
				Measures: map[string]analysis.AnalysisMeasureDefinition{
					"RegionGrain": {Name: "RegionGrain", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Region"}},
				},
				Connectors: map[string]analysis.AnalysisConnectorDefinition{
					"sales_data": {
						Name: "sales_data",
						Grain: []*analysis.AnalysisMeasureDefinition{
							{Name: "RegionGrain", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Region"}},
						},
					},
				},
			},
			conceptValues: map[string][]string{
				"Region":      {"North", "South"},
				"RegionGrain": {"North", "South"},
			},
			conditionalConceptValues: map[string]ConditionalConceptEntry{},
			expectNoMaskFor:          "sales_data",
			expectMaskCount:          0,
		},
		{
			name: "multiple connectors with mixed conditionals",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"DayInMonth": {
						ConceptName: "DayInMonth",
						ConditionalValues: map[string][]string{
							"Jan": {"1", "2"},
						},
					},
					"Month":  {ConceptName: "Month"},
					"Region": {ConceptName: "Region"},
				},
				Measures: map[string]analysis.AnalysisMeasureDefinition{
					"Monthly":  {Name: "Monthly", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Month"}},
					"Regional": {Name: "Regional", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Region"}},
				},
				Connectors: map[string]analysis.AnalysisConnectorDefinition{
					"monthly_data": {
						Name: "monthly_data",
						Grain: []*analysis.AnalysisMeasureDefinition{
							{Name: "Month", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Month"}},
							{Name: "DayInMonth", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "DayInMonth"}},
						},
					},
					"regional_data": {
						Name: "regional_data",
						Grain: []*analysis.AnalysisMeasureDefinition{
							{Name: "Regional", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Region"}},
						},
					},
				},
			},
			conceptValues: map[string][]string{
				"Month":      {"Jan", "Feb"},
				"DayInMonth": {"1", "2"},
				"Region":     {"North", "South"},
				"Regional":   {"North", "South"},
			},
			conditionalConceptValues: map[string]ConditionalConceptEntry{
				"DayInMonth": {
					ParentConcept: "Month",
					Values: map[string][]string{
						"Jan": {"1", "2"},
						"Feb": {"1"},
					},
				},
			},
			expectMaskFor:   "monthly_data",
			expectNoMaskFor: "regional_data",
			expectMaskCount: 1,
		},
		{
			name: "grain measure name differs from concept name — UnitName resolves mask",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"DayInMonth": {
						ConceptName: "DayInMonth",
						ConditionalValues: map[string][]string{
							"Jan": {"1", "2", "3"},
							"Feb": {"1", "2"},
						},
					},
					"Month": {ConceptName: "Month"},
				},
				Measures: map[string]analysis.AnalysisMeasureDefinition{
					"Monthly":  {Name: "Monthly", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Month"}},
					"DailyDim": {Name: "DailyDim", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "DayInMonth"}},
				},
				Connectors: map[string]analysis.AnalysisConnectorDefinition{
					"diff_names": {
						Name: "diff_names",
						Grain: []*analysis.AnalysisMeasureDefinition{
							{Name: "Monthly", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Month"}},
							{Name: "DailyDim", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "DayInMonth"}},
						},
					},
				},
			},
			conceptValues: map[string][]string{
				"Month":      {"Jan", "Feb"},
				"DayInMonth": {"1", "2", "3"},
				"Monthly":    {"Jan", "Feb"},
				"DailyDim":   {"1", "2", "3"},
			},
			conditionalConceptValues: map[string]ConditionalConceptEntry{
				"DayInMonth": {
					ParentConcept: "Month",
					Values: map[string][]string{
						"Jan": {"1", "2", "3"},
						"Feb": {"1", "2"},
					},
				},
			},
			expectMaskFor:   "diff_names",
			expectMaskCount: 1,
		},
		{
			name: "grain measure with empty UnitName returns error",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"Month":      {ConceptName: "Month"},
					"DayInMonth": {ConceptName: "DayInMonth"},
				},
				Measures: map[string]analysis.AnalysisMeasureDefinition{
					"Month": {Name: "Month", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Month"}},
				},
				Connectors: map[string]analysis.AnalysisConnectorDefinition{
					"bad_grain": {
						Name: "bad_grain",
						Grain: []*analysis.AnalysisMeasureDefinition{
							{Name: "Month", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Month"}},
							{Name: "DayInMonth", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: ""}},
						},
					},
				},
			},
			expectError:         true,
			expectErrorContains: "no UnitName specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := BuildTransformerFromAnalysis(tt.analysisDefs)
			if err != nil {
				t.Fatalf("BuildTransformerFromAnalysis() returned error: %v", err)
			}
			result, err := input.BuildConnectorValidityMasks(tt.conditionalConceptValues, tt.conceptValues)
			if tt.expectError {
				if err == nil {
					t.Errorf("BuildConnectorValidityMasks() returned nil error, expected error")
				} else if tt.expectErrorContains != "" && !strings.Contains(err.Error(), tt.expectErrorContains) {
					t.Errorf("BuildConnectorValidityMasks() error = %q, expected to contain %q", err.Error(), tt.expectErrorContains)
				}
				return
			}
			if err != nil {
				t.Errorf("BuildConnectorValidityMasks() returned unexpected error: %v", err)
				return
			}
			if len(result) != tt.expectMaskCount {
				t.Errorf("BuildConnectorValidityMasks() returned %d entries, expected %d", len(result), tt.expectMaskCount)
			}
			if tt.expectMaskFor != "" {
				mask, ok := result[tt.expectMaskFor]
				if !ok {
					t.Errorf("BuildConnectorValidityMasks() missing expected mask for connector %q", tt.expectMaskFor)
				} else if mask == nil {
					t.Errorf("BuildConnectorValidityMasks() mask for connector %q is nil", tt.expectMaskFor)
				}
			}
			if tt.expectNoMaskFor != "" {
				if _, ok := result[tt.expectNoMaskFor]; ok {
					t.Errorf("BuildConnectorValidityMasks() has unexpected mask for connector %q", tt.expectNoMaskFor)
				}
			}
		})
	}
}

// TestComputeValidityMask_MonthDayInMonth verifies conditional validity with Month/DayInMonth.
func TestComputeValidityMask_MonthDayInMonth(t *testing.T) {
	dimNames := []string{"Month", "DayInMonth"}
	conceptValues := map[string][]string{
		"Month":      {"Jan", "Feb", "Mar"},
		"DayInMonth": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31"},
	}
	conditional := map[string]ConditionalConceptEntry{
		"DayInMonth": {
			ParentConcept: "Month",
			Values: map[string][]string{
				"Jan": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31"},
				"Feb": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28"},
				"Mar": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31"},
			},
		},
	}

	mask := ComputeValidityMask(dimNames, conditional, conceptValues)
	if mask == nil {
		t.Fatalf("ComputeValidityMask returned nil, expected non-nil mask")
	}

	// Helper to check a specific cell's validity
	cellIndex := func(monthIdx, dayIdx int) int {
		return monthIdx*31 + dayIdx
	}

	isInvalid := func(cell int) bool {
		word := cell >> 6
		bit := uint(cell & 63)
		return (mask[word] & (1 << bit)) != 0
	}

	// Feb-30 should be invalid (Feb index=1, DayInMonth "30" index=29)
	feb30 := cellIndex(1, 29)
	if !isInvalid(feb30) {
		t.Errorf("Feb-30 (cell %d) should be invalid", feb30)
	}

	// Jan-15 should be valid (Jan index=0, DayInMonth "15" index=14)
	jan15 := cellIndex(0, 14)
	if isInvalid(jan15) {
		t.Errorf("Jan-15 (cell %d) should be valid", jan15)
	}

	// Feb days 29-31 should all be invalid (indices 28, 29, 30)
	for _, dayIdx := range []int{28, 29, 30} {
		cell := cellIndex(1, dayIdx)
		if !isInvalid(cell) {
			t.Errorf("Feb day %d (cell %d) should be invalid", dayIdx+1, cell)
		}
	}
}

// TestComputeValidityMask_OverlappingDimensionValues verifies that when two
// dimensions share the same value names, ComputeValidityMask still resolves
// the correct parent dimension. Month and Season both have values "Jan" and
// "Feb", but DayInMonth is conditional on Month, not Season. A buggy
// implementation that guesses the parent by value overlap would incorrectly
// use Season.
func TestComputeValidityMask_OverlappingDimensionValues(t *testing.T) {
	dimNames := []string{"Month", "Season", "DayInMonth"}
	conceptValues := map[string][]string{
		"Month":      {"Jan", "Feb"},
		"Season":     {"Jan", "Feb"},
		"DayInMonth": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31"},
	}
	conditional := map[string]ConditionalConceptEntry{
		"DayInMonth": {
			ParentConcept: "Month",
			Values: map[string][]string{
				"Jan": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31"},
				"Feb": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28"},
			},
		},
	}

	mask := ComputeValidityMask(dimNames, conditional, conceptValues)
	if mask == nil {
		t.Fatalf("ComputeValidityMask returned nil, expected non-nil mask")
	}

	// Helper to compute flat index from dimension indices
	// Strides: Month=62, Season=31, DayInMonth=1
	cellIndex := func(monthIdx, seasonIdx, dayIdx int) int {
		return monthIdx*62 + seasonIdx*31 + dayIdx
	}

	isInvalid := func(cell int) bool {
		word := cell >> 6
		bit := uint(cell & 63)
		return (mask[word] & (1 << bit)) != 0
	}

	// Month=Jan, Season=Feb, DayInMonth=29 should be VALID because
	// DayInMonth=29 is valid for Month=Jan. A buggy implementation that
	// used Season as parent would see Season=Feb and mark it invalid.
	cellJanFeb29 := cellIndex(0, 1, 28)
	if isInvalid(cellJanFeb29) {
		t.Errorf("Month=Jan, Season=Feb, DayInMonth=29 (cell %d) should be valid", cellJanFeb29)
	}

	// Month=Feb, Season=Jan, DayInMonth=29 should be INVALID because
	// DayInMonth=29 is NOT valid for Month=Feb. A buggy implementation that
	// used Season as parent would see Season=Jan and incorrectly mark it valid.
	cellFebJan29 := cellIndex(1, 0, 28)
	if !isInvalid(cellFebJan29) {
		t.Errorf("Month=Feb, Season=Jan, DayInMonth=29 (cell %d) should be invalid", cellFebJan29)
	}

	// Sanity checks for cells where Month and Season align
	// Month=Jan, Season=Jan, DayInMonth=29 should be VALID
	cellJanJan29 := cellIndex(0, 0, 28)
	if isInvalid(cellJanJan29) {
		t.Errorf("Month=Jan, Season=Jan, DayInMonth=29 (cell %d) should be valid", cellJanJan29)
	}

	// Month=Feb, Season=Feb, DayInMonth=29 should be INVALID
	cellFebFeb29 := cellIndex(1, 1, 28)
	if !isInvalid(cellFebFeb29) {
		t.Errorf("Month=Feb, Season=Feb, DayInMonth=29 (cell %d) should be invalid", cellFebFeb29)
	}

	// DayInMonth=28 should be valid for all Month/Season combinations
	for monthIdx := 0; monthIdx < 2; monthIdx++ {
		for seasonIdx := 0; seasonIdx < 2; seasonIdx++ {
			cell := cellIndex(monthIdx, seasonIdx, 27)
			if isInvalid(cell) {
				t.Errorf("Month=%d, Season=%d, DayInMonth=28 (cell %d) should be valid", monthIdx, seasonIdx, cell)
			}
		}
	}
}

// TestComputeValidityMask_NoConditionals verifies nil mask when no conditionals exist.
func TestComputeValidityMask_NoConditionals(t *testing.T) {
	dimNames := []string{"X", "Y"}
	conceptValues := map[string][]string{
		"X": {"x1", "x2"},
		"Y": {"y1", "y2"},
	}
	conditional := map[string]ConditionalConceptEntry{}

	mask := ComputeValidityMask(dimNames, conditional, conceptValues)
	if mask != nil {
		t.Errorf("ComputeValidityMask returned %v, expected nil", mask)
	}
}

// TestComputeValidityMask_EmptyDimensions verifies nil mask for empty dimensions.
func TestComputeValidityMask_EmptyDimensions(t *testing.T) {
	mask := ComputeValidityMask([]string{}, nil, nil)
	if mask != nil {
		t.Errorf("ComputeValidityMask returned %v, expected nil", mask)
	}
}

// TestComputeValidityMask_BitPackingAcrossUint64Boundary verifies mask spans multiple uint64 words.
func TestComputeValidityMask_BitPackingAcrossUint64Boundary(t *testing.T) {
	// Create >64 cells: 10 x 10 = 100 cells
	dimNames := []string{"A", "B"}
	conceptValues := map[string][]string{
		"A": {"a0", "a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9"},
		"B": {"b0", "b1", "b2", "b3", "b4", "b5", "b6", "b7", "b8", "b9"},
	}
	// Make cell (a7, b3) invalid: a7=index7, b3=index3 => flat index = 7*10+3 = 73
	conditional := map[string]ConditionalConceptEntry{
		"B": {
			ParentConcept: "A",
			Values: map[string][]string{
				"a0": {"b0", "b1", "b2", "b3", "b4", "b5", "b6", "b7", "b8", "b9"},
				"a1": {"b0", "b1", "b2", "b3", "b4", "b5", "b6", "b7", "b8", "b9"},
				"a2": {"b0", "b1", "b2", "b3", "b4", "b5", "b6", "b7", "b8", "b9"},
				"a3": {"b0", "b1", "b2", "b3", "b4", "b5", "b6", "b7", "b8", "b9"},
				"a4": {"b0", "b1", "b2", "b3", "b4", "b5", "b6", "b7", "b8", "b9"},
				"a5": {"b0", "b1", "b2", "b3", "b4", "b5", "b6", "b7", "b8", "b9"},
				"a6": {"b0", "b1", "b2", "b3", "b4", "b5", "b6", "b7", "b8", "b9"},
				"a7": {"b0", "b1", "b2", "b4", "b5", "b6", "b7", "b8", "b9"}, // missing b3
				"a8": {"b0", "b1", "b2", "b3", "b4", "b5", "b6", "b7", "b8", "b9"},
				"a9": {"b0", "b1", "b2", "b3", "b4", "b5", "b6", "b7", "b8", "b9"},
			},
		},
	}

	mask := ComputeValidityMask(dimNames, conditional, conceptValues)
	if mask == nil {
		t.Fatalf("ComputeValidityMask returned nil, expected non-nil mask")
	}

	// 100 cells => ceil(100/64) = 2 uint64s
	if len(mask) != 2 {
		t.Fatalf("expected mask length 2, got %d", len(mask))
	}

	// Cell 73 should be in word 1 (73>>6 = 1), bit 9 (73&63 = 9)
	word := 1
	bit := uint(9)
	if (mask[word] & (1 << bit)) == 0 {
		t.Errorf("cell 73 (word %d, bit %d) should be invalid", word, bit)
	}

	// Verify all other cells are valid
	invalidCount := 0
	for cell := 0; cell < 100; cell++ {
		w := cell >> 6
		b := uint(cell & 63)
		if (mask[w] & (1 << b)) != 0 {
			invalidCount++
		}
	}
	if invalidCount != 1 {
		t.Errorf("expected exactly 1 invalid cell, got %d", invalidCount)
	}

	// Verify bits beyond totalCells (100..127) are set to invalid.
	// These occupy mask[1] at positions 36..63.
	for bit := uint(36); bit < 64; bit++ {
		cell := 64 + int(bit)
		if (mask[1] & (1 << bit)) == 0 {
			t.Errorf("trailing bit %d (cell %d) should be invalid (set to 1)", bit, cell)
		}
	}
}

// TestComputeValidityMask_ParentDimensionAbsent verifies that when a conditional
// child dimension's parent is not present in dimNames, all cells are marked
// invalid because the parent value cannot be resolved.
func TestComputeValidityMask_ParentDimensionAbsent(t *testing.T) {
	dimNames := []string{"DayInMonth"}
	conceptValues := map[string][]string{
		"DayInMonth": {"1", "2", "3"},
	}
	conditional := map[string]ConditionalConceptEntry{
		"DayInMonth": {
			ParentConcept: "Month", // Month is absent from dimNames
			Values: map[string][]string{
				"Jan": {"1", "2", "3"},
			},
		},
	}

	mask := ComputeValidityMask(dimNames, conditional, conceptValues)
	if mask == nil {
		t.Fatalf("ComputeValidityMask returned nil, expected non-nil mask")
	}

	// All 3 cells should be invalid
	for cell := 0; cell < 3; cell++ {
		word := cell >> 6
		bit := uint(cell & 63)
		if (mask[word] & (1 << bit)) == 0 {
			t.Errorf("cell %d should be invalid when parent dimension is absent", cell)
		}
	}
}

// TestBuildConceptValuesMap verifies concept value extraction from analysis definitions.
func TestBuildConceptValuesMap(t *testing.T) {
	tests := []struct {
		name         string
		analysisDefs *analysis.AnalysisDefinitions
		expectMap    map[string][]string
	}{
		{
			name:         "nil analysis defs",
			analysisDefs: nil,
			expectMap:    map[string][]string{},
		},
		{
			name: "concepts with resolved values",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"Year": {
						ConceptName: "Year",
						ResolvedValues: []analysis.AnalysisConceptValueDefinition{
							{Name: "2023", Position: 0},
							{Name: "2024", Position: 1},
						},
					},
					"Region": {
						ConceptName: "Region",
						ResolvedValues: []analysis.AnalysisConceptValueDefinition{
							{Name: "North", Position: 0},
							{Name: "South", Position: 1},
							{Name: "East", Position: 2},
							{Name: "West", Position: 3},
						},
					},
				},
			},
			expectMap: map[string][]string{
				"Year":   {"2023", "2024"},
				"Region": {"North", "South", "East", "West"},
			},
		},
		{
			name: "concepts without resolved values",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"Empty": {
						ConceptName:    "Empty",
						ResolvedValues: []analysis.AnalysisConceptValueDefinition{},
					},
				},
			},
			expectMap: map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := BuildTransformerFromAnalysis(tt.analysisDefs)
			if err != nil {
				t.Fatalf("BuildTransformerFromAnalysis() returned error: %v", err)
			}
			result := input.BuildConceptValuesMap()
			if len(result) != len(tt.expectMap) {
				t.Errorf("BuildConceptValuesMap() returned %d entries, expected %d", len(result), len(tt.expectMap))
			}
			for key, expectedValues := range tt.expectMap {
				actualValues, ok := result[key]
				if !ok {
					t.Errorf("BuildConceptValuesMap() missing key %q", key)
					continue
				}
				if len(actualValues) != len(expectedValues) {
					t.Errorf("BuildConceptValuesMap()[%q] has %d values, expected %d", key, len(actualValues), len(expectedValues))
					continue
				}
				for i, v := range expectedValues {
					if actualValues[i] != v {
						t.Errorf("BuildConceptValuesMap()[%q][%d] = %q, expected %q", key, i, actualValues[i], v)
					}
				}
			}
		})
	}
}

// TestBuildConceptValuesMap_WithGrainMeasures verifies that BuildConceptValuesMap
// includes grain measure names as keys when their UnitName references a concept
// that has resolved values, duplicating the concept values under the measure name.
func TestBuildConceptValuesMap_WithGrainMeasures(t *testing.T) {
	tests := []struct {
		name         string
		analysisDefs *analysis.AnalysisDefinitions
		expectMap    map[string][]string
	}{
		{
			name: "grain measure with UnitName duplicates concept values",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"DayOfWeek": {
						ConceptName: "DayOfWeek",
						ResolvedValues: []analysis.AnalysisConceptValueDefinition{
							{Name: "Monday", Position: 0},
							{Name: "Tuesday", Position: 1},
						},
					},
				},
				Measures: map[string]analysis.AnalysisMeasureDefinition{
					"Daily": {
						Name:     "Daily",
						TypeName: "Interval",
						UnitExpression: &analysis.UnitReference{Name: "DayOfWeek"},
					},
				},
				Connectors: map[string]analysis.AnalysisConnectorDefinition{
					"SalesData": {
						Name: "SalesData",
						Dimensions: []*analysis.AnalysisConceptDefinition{
							{ConceptName: "Region"},
						},
						Grain: []*analysis.AnalysisMeasureDefinition{
							{
								Name:     "Daily",
								TypeName: "Interval",
								UnitExpression: &analysis.UnitReference{Name: "DayOfWeek"},
							},
						},
					},
				},
			},
			expectMap: map[string][]string{
				"DayOfWeek": {"Monday", "Tuesday"},
				"Daily":     {"Monday", "Tuesday"},
			},
		},
		{
			name: "grain measure without UnitName is not added",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"Region": {
						ConceptName: "Region",
						ResolvedValues: []analysis.AnalysisConceptValueDefinition{
							{Name: "North", Position: 0},
						},
					},
				},
				Measures: map[string]analysis.AnalysisMeasureDefinition{
					"SimpleGrain": {
						Name:     "SimpleGrain",
						TypeName: "Interval",
						UnitExpression: &analysis.UnitReference{Name: ""},
					},
				},
				Connectors: map[string]analysis.AnalysisConnectorDefinition{
					"Data": {
						Name: "Data",
						Grain: []*analysis.AnalysisMeasureDefinition{
							{
								Name:     "SimpleGrain",
								TypeName: "Interval",
								UnitExpression: &analysis.UnitReference{Name: ""},
							},
						},
					},
				},
			},
			expectMap: map[string][]string{
				"Region": {"North"},
			},
		},
		{
			name: "multiple connectors with grain measures",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"Year": {
						ConceptName: "Year",
						ResolvedValues: []analysis.AnalysisConceptValueDefinition{
							{Name: "2023", Position: 0},
							{Name: "2024", Position: 1},
						},
					},
					"Month": {
						ConceptName: "Month",
						ResolvedValues: []analysis.AnalysisConceptValueDefinition{
							{Name: "Jan", Position: 0},
							{Name: "Feb", Position: 1},
						},
					},
				},
				Measures: map[string]analysis.AnalysisMeasureDefinition{
					"Yearly": {
						Name:     "Yearly",
						TypeName: "Interval",
						UnitExpression: &analysis.UnitReference{Name: "Year"},
					},
					"Monthly": {
						Name:     "Monthly",
						TypeName: "Interval",
						UnitExpression: &analysis.UnitReference{Name: "Month"},
					},
				},
				Connectors: map[string]analysis.AnalysisConnectorDefinition{
					"Sales": {
						Name: "Sales",
						Grain: []*analysis.AnalysisMeasureDefinition{
							{
								Name:     "Yearly",
								TypeName: "Interval",
								UnitExpression: &analysis.UnitReference{Name: "Year"},
							},
						},
					},
					"Inventory": {
						Name: "Inventory",
						Grain: []*analysis.AnalysisMeasureDefinition{
							{
								Name:     "Monthly",
								TypeName: "Interval",
								UnitExpression: &analysis.UnitReference{Name: "Month"},
							},
						},
					},
				},
			},
			expectMap: map[string][]string{
				"Year":    {"2023", "2024"},
				"Month":   {"Jan", "Feb"},
				"Yearly":  {"2023", "2024"},
				"Monthly": {"Jan", "Feb"},
			},
		},
		{
			name: "grain measure referencing non-existent concept is ignored",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"Day": {
						ConceptName: "Day",
						ResolvedValues: []analysis.AnalysisConceptValueDefinition{
							{Name: "Monday", Position: 0},
						},
					},
				},
				Measures: map[string]analysis.AnalysisMeasureDefinition{
					"Weekly": {
						Name:     "Weekly",
						TypeName: "Interval",
						UnitExpression: &analysis.UnitReference{Name: "Week"}, // Week concept does not exist
					},
				},
				Connectors: map[string]analysis.AnalysisConnectorDefinition{
					"Data": {
						Name: "Data",
						Grain: []*analysis.AnalysisMeasureDefinition{
							{
								Name:     "Weekly",
								TypeName: "Interval",
								UnitExpression: &analysis.UnitReference{Name: "Week"},
							},
						},
					},
				},
			},
			expectMap: map[string][]string{
				"Day": {"Monday"},
			},
		},
		{
			name:         "nil analysis defs",
			analysisDefs: nil,
			expectMap:    map[string][]string{},
		},
		{
			name: "connector with no grain does not add entries",
			analysisDefs: &analysis.AnalysisDefinitions{
				Concepts: map[string]analysis.AnalysisConceptDefinition{
					"Region": {
						ConceptName: "Region",
						ResolvedValues: []analysis.AnalysisConceptValueDefinition{
							{Name: "North", Position: 0},
						},
					},
				},
				Connectors: map[string]analysis.AnalysisConnectorDefinition{
					"NoGrainData": {
						Name:       "NoGrainData",
						Dimensions: []*analysis.AnalysisConceptDefinition{{ConceptName: "Region"}},
						Grain:      []*analysis.AnalysisMeasureDefinition{},
					},
				},
			},
			expectMap: map[string][]string{
				"Region": {"North"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := BuildTransformerFromAnalysis(tt.analysisDefs)
			if err != nil {
				t.Fatalf("BuildTransformerFromAnalysis() returned error: %v", err)
			}
			result := input.BuildConceptValuesMap()
			if len(result) != len(tt.expectMap) {
				t.Errorf("BuildConceptValuesMap() returned %d entries, expected %d", len(result), len(tt.expectMap))
			}
			for key, expectedValues := range tt.expectMap {
				actualValues, ok := result[key]
				if !ok {
					t.Errorf("BuildConceptValuesMap() missing key %q", key)
					continue
				}
				if len(actualValues) != len(expectedValues) {
					t.Errorf("BuildConceptValuesMap()[%q] has %d values, expected %d", key, len(actualValues), len(expectedValues))
					continue
				}
				for i, v := range expectedValues {
					if actualValues[i] != v {
						t.Errorf("BuildConceptValuesMap()[%q][%d] = %q, expected %q", key, i, actualValues[i], v)
					}
				}
			}
			// Verify no unexpected keys
			for key := range result {
				if _, ok := tt.expectMap[key]; !ok {
					t.Errorf("BuildConceptValuesMap() contains unexpected key %q", key)
				}
			}
		})
	}
}

// TestGenerateMainFunctionWithConceptValues verifies concept values map generation in output.
func TestGenerateMainFunctionWithConceptValues(t *testing.T) {
	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Year": {
				ConceptName: "Year",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "2023", Position: 0},
					{Name: "2024", Position: 1},
				},
			},
			"Region": {
				ConceptName: "Region",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "North", Position: 0},
					{Name: "South", Position: 1},
					{Name: "East", Position: 2},
					{Name: "West", Position: 3},
				},
			},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "sales_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs:      []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{},
	}

	transformer := cil.NewCILAbstractTransformer()
	transformer.RegisterConnector("sales_data", []string{"Region"}, nil, nil, "float64")

	input := &SynthesisInput{
		Transformer: transformer,
		IOSpec:      ioSpec,
		Concepts:    analysisDefs.Concepts,
	}
	code, err := GenerateMainFunction(input)
	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	if !strings.Contains(code, "var conceptValues = map[string][]string") {
		t.Errorf("GenerateMainFunction() output missing conceptValues declaration")
	}
	if !strings.Contains(code, `"Year"`) {
		t.Errorf("GenerateMainFunction() output missing Year concept")
	}
	if !strings.Contains(code, `"Region"`) {
		t.Errorf("GenerateMainFunction() output missing Region concept")
	}
	if !strings.Contains(code, `"2023"`) {
		t.Errorf("GenerateMainFunction() output missing 2023 value")
	}
	if !strings.Contains(code, `"North"`) {
		t.Errorf("GenerateMainFunction() output missing North value")
	}
}

// TestGenerateMainFunctionWithConditionalConceptValues verifies that when
// conditional concept values exist AND a connector's grain uses them,
// connectorValidityMasks is emitted and the unused conditionalConceptValues
// variable is not.
func TestGenerateMainFunctionWithConditionalConceptValues(t *testing.T) {
	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"DayInMonth": {
				ConceptName:       "DayInMonth",
				ParentConceptName: strPtr("Month"),
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "1", Position: 0},
					{Name: "2", Position: 1},
					{Name: "3", Position: 2},
				},
				ConditionalValues: map[string][]string{
					"Jan": {"1", "2", "3"},
					"Feb": {"1", "2"},
				},
			},
			"Month": {
				ConceptName: "Month",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "Jan", Position: 0},
					{Name: "Feb", Position: 1},
				},
			},
			"Year": {
				ConceptName: "Year",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "2023", Position: 0},
					{Name: "2024", Position: 1},
				},
			},
		},
		Measures: map[string]analysis.AnalysisMeasureDefinition{
			"Monthly": {Name: "Monthly", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Month"}},
			"DayDim":  {Name: "DayDim", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "DayInMonth"}},
		},
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"sales_data": {
				Name: "sales_data",
				Grain: []*analysis.AnalysisMeasureDefinition{
					{Name: "Monthly", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Month"}},
					{Name: "DayDim", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "DayInMonth"}},
				},
			},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "sales_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs:      []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{},
	}

	transformer := cil.NewCILAbstractTransformer()
	transformer.RegisterConnector("sales_data", []string{"Year"}, []string{"Monthly", "DayDim"}, nil, "float64")

	input := &SynthesisInput{
		Transformer: transformer,
		IOSpec:      ioSpec,
		Concepts:    analysisDefs.Concepts,
		Measures:    analysisDefs.Measures,
		Connectors:  analysisDefs.Connectors,
	}
	code, err := GenerateMainFunction(input)
	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	if strings.Contains(code, "conditionalConceptValues") {
		t.Errorf("GenerateMainFunction() output should NOT emit unused conditionalConceptValues")
	}
	if !strings.Contains(code, "var connectorValidityMasks = map[string][]uint64") {
		t.Errorf("GenerateMainFunction() output missing connectorValidityMasks declaration")
	}
}

// TestGenerateMainFunctionWithConditionalConceptValuesEmpty verifies that when
// no concepts have conditional values, the conditionalConceptValues variable
// is NOT emitted.
func TestGenerateMainFunctionWithConditionalConceptValuesEmpty(t *testing.T) {
	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Year": {
				ConceptName: "Year",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "2023", Position: 0},
				},
			},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors:       []cio.ConnectorSpecification{},
		Outputs:      []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{},
	}

	transformer := cil.NewCILAbstractTransformer()

	input := &SynthesisInput{
		Transformer: transformer,
		IOSpec:      ioSpec,
		Concepts:    analysisDefs.Concepts,
		Measures:    analysisDefs.Measures,
		Connectors:  analysisDefs.Connectors,
	}
	code, err := GenerateMainFunction(input)
	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	if strings.Contains(code, "conditionalConceptValues") {
		t.Errorf("GenerateMainFunction() output should NOT contain conditionalConceptValues when no conditionals exist")
	}
}

// TestGenerateMainFunctionWithDriverParameters verifies driver parameters generation in output.
func TestGenerateMainFunctionWithDriverParameters(t *testing.T) {
	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "testdata",
				DataFormat:    "parquet",
				Driver:        "file",
				DriverParameters: []map[string]interface{}{
					{"path": "testdata.parquet"},
				},
			},
		},
		Outputs:      []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{},
	}

	transformer := cil.NewCILAbstractTransformer()
	transformer.RegisterConnector("testdata", []string{"Year", "Region"}, nil, nil, "float64")

	code, err := GenerateMainFunction(&SynthesisInput{Transformer: transformer, IOSpec: ioSpec})
	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	if !strings.Contains(code, "DriverParameters: []map[string]interface{}") {
		t.Errorf("GenerateMainFunction() output missing DriverParameters")
	}
	if !strings.Contains(code, `"path"`) {
		t.Errorf("GenerateMainFunction() output missing path key")
	}
	if !strings.Contains(code, `"testdata.parquet"`) {
		t.Errorf("GenerateMainFunction() output missing testdata.parquet value")
	}
}

// TestGenerateMainFunctionPassesConceptValuesToLoaders verifies conceptValues is passed to loaders.
func TestGenerateMainFunctionPassesConceptValuesToLoaders(t *testing.T) {
	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "sales_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs:      []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{},
	}

	transformer := cil.NewCILAbstractTransformer()
	transformer.RegisterConnector("sales_data", []string{"Region"}, nil, nil, "float64")

	code, err := GenerateMainFunction(&SynthesisInput{Transformer: transformer, IOSpec: ioSpec})
	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	if !strings.Contains(code, "loadSalesData(sales_dataConnectorSpec, conceptValues)") {
		t.Errorf("GenerateMainFunction() output missing conceptValues passed to loader")
	}
}

// TestSynthesizeErrorHandling verifies that Synthesize returns errors for invalid inputs.
func TestSynthesizeErrorHandling(t *testing.T) {
	t.Run("nil io spec", func(t *testing.T) {
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "output.go")

		err := Synthesize(&analysis.AnalysisDefinitions{}, nil, outputPath)
		if err == nil {
			t.Errorf("Expected error for nil ioSpec, got nil")
		}
	})

	t.Run("invalid output path", func(t *testing.T) {
		ioSpec := &cio.IOSpecification{
			Connectors:       []cio.ConnectorSpecification{},
			Outputs:      []cio.OutputSpecification{},
			Architecture: cio.ArchitectureSpecification{},
		}

		err := Synthesize(&analysis.AnalysisDefinitions{}, ioSpec, "/nonexistent/dir/output.go")
		if err == nil {
			t.Errorf("Expected error for invalid output path, got nil")
		}
	})
}

// TestSynthesizeGeneratesParquetOutputWriting is an integration test that verifies
// the full Synthesize() pipeline produces generated code containing Parquet-writing
// calls driven by ioSpec.Outputs. It creates analysis definitions with concepts,
// connectors, and an IO spec with a single metric output, then validates the
// generated code includes cplio.WriteTensorToParquet, metric-specific variable
// names, error handling, and the parquet file path.
func TestSynthesizeGeneratesParquetOutputWriting(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.go")

	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Region": {
				ConceptName: "Region",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "North", Position: 0},
					{Name: "South", Position: 1},
				},
			},
		},
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"sales_data": {
				Name: "sales_data",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Region"},
				},
				CellType: &analysis.AnalysisMeasureDefinition{
					Name:     "Amount",
					TypeName: "Number",
				},
			},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "sales_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "total_sales",
				Dimensions: []string{"Region"},
				CellType:   "float64",
				Grain:      []string{"Region"},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	err := Synthesize(analysisDefs, ioSpec, outputPath)
	if err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Verify generated code contains expected Parquet output writing elements
	expectedElements := []string{
		"package main",
		"func main()",
		"cplio.WriteMetricsToParquet",
		"total_sales.parquet",
		"total_salesComp",
		"total_salesTensor",
		`controller.GetMetricComputation("total_sales")`,
		"check(err, \"",
		"os.Stderr",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Generated file missing expected content: %q", expected)
		}
	}

	// Verify old stdout placeholder is removed
	if strings.Contains(contentStr, "fmt.Printf(\"Writing output:") {
		t.Errorf("Generated file still contains old stdout placeholder")
	}
}

// TestSynthesizeMultipleOutputsGenerateParquetCalls verifies that when the IO spec
// contains multiple metric outputs, the generated code includes separate
// cplio.WriteTensorToParquet calls for each metric with their respective
// .parquet file paths.
func TestSynthesizeMultipleOutputsGenerateParquetCalls(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.go")

	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Region": {
				ConceptName: "Region",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "North", Position: 0},
					{Name: "South", Position: 1},
				},
			},
		},
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"sales_data": {
				Name: "sales_data",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Region"},
				},
				CellType: &analysis.AnalysisMeasureDefinition{
					Name:     "Amount",
					TypeName: "Number",
				},
			},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "sales_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "revenue",
				Dimensions: []string{"Region"},
				CellType:   "float64",
				Grain:      []string{"Region"},
			},
			{
				MetricName: "costs",
				Dimensions: []string{"Region"},
				CellType:   "float64",
				Grain:      []string{"Region"},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	err := Synthesize(analysisDefs, ioSpec, outputPath)
	if err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Verify per-metric parquet file paths
	if !strings.Contains(contentStr, "revenue.parquet") {
		t.Errorf("Generated file missing 'revenue.parquet'")
	}
	if !strings.Contains(contentStr, "costs.parquet") {
		t.Errorf("Generated file missing 'costs.parquet'")
	}

	// Count cplio.WriteMetricsToParquet occurrences - should be exactly 2
	count := strings.Count(contentStr, "cplio.WriteMetricsToParquet")
	if count != 2 {
		t.Errorf("Expected exactly 2 'cplio.WriteMetricsToParquet' calls, got %d", count)
	}

	// Verify both metrics have variable name generation
	if !strings.Contains(contentStr, "revenueComp") {
		t.Errorf("Generated file missing 'revenueComp' variable")
	}
	if !strings.Contains(contentStr, "costsTensor") {
		t.Errorf("Generated file missing 'costsTensor' variable")
	}
}

// TestSynthesizeNoOutputsNoParquetCalls verifies that when the IO spec has no
// outputs, the generated code does NOT contain any cplio.WriteTensorToParquet
// calls but still contains the main function and other expected structural elements.
func TestSynthesizeNoOutputsNoParquetCalls(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.go")

	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Region": {
				ConceptName: "Region",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "North", Position: 0},
					{Name: "South", Position: 1},
				},
			},
		},
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"sales_data": {
				Name: "sales_data",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Region"},
				},
				CellType: &analysis.AnalysisMeasureDefinition{
					Name:     "Amount",
					TypeName: "Number",
				},
			},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "sales_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs: []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	err := Synthesize(analysisDefs, ioSpec, outputPath)
	if err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Verify no parquet write calls are present
	if strings.Contains(contentStr, "cplio.WriteMetricsToParquet") {
		t.Errorf("Generated file should NOT contain 'cplio.WriteMetricsToParquet' when no outputs are specified")
	}

	// Verify main function and other expected structural elements still present
	expectedElements := []string{
		"package main",
		"func main()",
		"func Setup",
		"func Compute",
		"func Cleanup",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Generated file missing expected content: %q", expected)
		}
	}
}

// TestSynthesizeWithMetrics is an integration test that verifies the full
// Synthesize() pipeline produces correct code when analysis definitions include
// metrics. It creates AnalysisDefinitions with concepts, connectors, and metrics,
// an IO spec with matching metric outputs, and validates all expected metric
// lifecycle elements are present in the generated code.
func TestSynthesizeWithMetrics(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.go")

	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Region": {
				ConceptName: "Region",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "North", Position: 0},
					{Name: "South", Position: 1},
					{Name: "East", Position: 2},
					{Name: "West", Position: 3},
				},
			},
			"Price": {
				ConceptName: "Price",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "Low", Position: 0},
					{Name: "Medium", Position: 1},
					{Name: "High", Position: 2},
				},
			},
		},
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"demand_data": {
				Name: "demand_data",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Region"},
				},
				CellType: &analysis.AnalysisMeasureDefinition{
					Name:     "demand_forecast",
					TypeName: "Number",
				},
			},
		},
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"revenue": {
				Name: "revenue",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "demand_forecast"},
					{Kind: "dimension", Name: "Price"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode:   &cil.CILLiteral{Value: 0},
					InputRefs: []cil.CILInputRef{{Kind: "dimension", Name: "Price"}},
				},
			},
			"profit": {
				Name: "profit",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "cost_forecast"},
					{Kind: "dimension", Name: "Price"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode:   &cil.CILLiteral{Value: 1},
					InputRefs: []cil.CILInputRef{{Kind: "dimension", Name: "Price"}},
				},
			},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "demand_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "revenue",
				Dimensions: []string{"Price"},
				CellType:   "float64",
				Grain:      []string{"Price"},
			},
			{
				MetricName: "profit",
				Dimensions: []string{"Price"},
				CellType:   "float64",
				Grain:      []string{"Price"},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	err := Synthesize(analysisDefs, ioSpec, outputPath)
	if err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	// 1. Verify metric names variable declaration
	if !strings.Contains(contentStr, `var metricNames = []string{`) {
		t.Errorf("Generated file missing 'var metricNames = []string{'")
	}
	if !strings.Contains(contentStr, `"revenue"`) {
		t.Errorf("Generated file missing '\"revenue\"' in metricNames")
	}
	if !strings.Contains(contentStr, `"profit"`) {
		t.Errorf("Generated file missing '\"profit\"' in metricNames")
	}

	// 2. Verify metric registration in Setup (now with error handling)
	if !strings.Contains(contentStr, "controller.RegisterMetricComputation(metricName, mc)") {
		t.Errorf("Generated file missing 'controller.RegisterMetricComputation(metricName, mc)'")
	}

	// 3. Verify goroutine launches in Compute function with WaitGroup sync
	if !strings.Contains(contentStr, "computerevenue(controller, inputs)") {
		t.Errorf("Generated file missing 'computerevenue(controller, inputs)'")
	}
	if !strings.Contains(contentStr, "computeprofit(controller, inputs)") {
		t.Errorf("Generated file missing 'computeprofit(controller, inputs)'")
	}
	if !strings.Contains(contentStr, "sync.WaitGroup") {
		t.Errorf("Generated file missing 'sync.WaitGroup'")
	}
	if !strings.Contains(contentStr, "wg.Wait()") {
		t.Errorf("Generated file missing 'wg.Wait()'")
	}

	// 4. Verify Parquet output writing calls
	if !strings.Contains(contentStr, "cplio.WriteMetricsToParquet") {
		t.Errorf("Generated file missing 'cplio.WriteMetricsToParquet'")
	}

	// 5. Verify core lifecycle functions
	expectedFunctions := []string{
		"func main()",
		"func Setup",
		"func Compute",
		"func Cleanup",
	}
	for _, expected := range expectedFunctions {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Generated file missing expected function: %q", expected)
		}
	}

	// 6. Verify metric-specific variable names
	if !strings.Contains(contentStr, "revenueComp") {
		t.Errorf("Generated file missing 'revenueComp' variable")
	}
	if !strings.Contains(contentStr, "revenueTensor") {
		t.Errorf("Generated file missing 'revenueTensor' variable")
	}
	if !strings.Contains(contentStr, "profitComp") {
		t.Errorf("Generated file missing 'profitComp' variable")
	}
	if !strings.Contains(contentStr, "profitTensor") {
		t.Errorf("Generated file missing 'profitTensor' variable")
	}

	// 7. Verify controller.GetMetricComputation calls
	if !strings.Contains(contentStr, `controller.GetMetricComputation("revenue")`) {
		t.Errorf("Generated file missing 'controller.GetMetricComputation(\"revenue\")'")
	}
	if !strings.Contains(contentStr, `controller.GetMetricComputation("profit")`) {
		t.Errorf("Generated file missing 'controller.GetMetricComputation(\"profit\")'")
	}

	// 8. Verify Parquet file paths (per-metric)
	if !strings.Contains(contentStr, "revenue.parquet") {
		t.Errorf("Generated file missing 'revenue.parquet'")
	}
	if !strings.Contains(contentStr, "profit.parquet") {
		t.Errorf("Generated file missing 'profit.parquet'")
	}

	// 9. Count WriteMetricsToParquet occurrences (should be exactly 2)
	count := strings.Count(contentStr, "cplio.WriteMetricsToParquet")
	if count != 2 {
		t.Errorf("Expected exactly 2 'cplio.WriteMetricsToParquet' calls, got %d", count)
	}
}

// TestCollectMetricNamesFromOutputs verifies that collectMetricNamesFromOutputs
// correctly extracts metric names from the IO spec's Outputs slice. It handles
// nil input, empty outputs, and populated outputs with expected metric names.
func TestCollectMetricNamesFromOutputs(t *testing.T) {
	tests := []struct {
		name     string
		ioSpec   *cio.IOSpecification
		expected []string
	}{
		{
			name:     "nil ioSpec returns nil",
			ioSpec:   nil,
			expected: nil,
		},
		{
			name: "empty outputs returns nil",
			ioSpec: &cio.IOSpecification{
				Connectors:       []cio.ConnectorSpecification{},
				Outputs:      []cio.OutputSpecification{},
				Architecture: cio.ArchitectureSpecification{},
			},
			expected: nil,
		},
		{
			name: "populated outputs returns metric names",
			ioSpec: &cio.IOSpecification{
				Connectors: []cio.ConnectorSpecification{},
				Outputs: []cio.OutputSpecification{
					{
						MetricName: "total_sales",
						Dimensions: []string{"Region"},
						CellType:   "float64",
						Grain:      []string{"Region"},
					},
					{
						MetricName: "avg_price",
						Dimensions: []string{"Product"},
						CellType:   "float64",
						Grain:      []string{"Product"},
					},
				},
				Architecture: cio.ArchitectureSpecification{},
			},
			expected: []string{"total_sales", "avg_price"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectMetricNamesFromOutputs(tt.ioSpec)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("collectMetricNamesFromOutputs() = %v, expected nil", result)
				}
			} else {
				if len(result) != len(tt.expected) {
					t.Errorf("collectMetricNamesFromOutputs() returned %d entries, expected %d", len(result), len(tt.expected))
				}
				for i, v := range tt.expected {
					if result[i] != v {
						t.Errorf("collectMetricNamesFromOutputs()[%d] = %q, expected %q", i, result[i], v)
					}
				}
			}
		})
	}
}

// TestEndToEndSynthesizeCompileRun verifies the full end-to-end pipeline:
// parse → analyze → synthesize → compile → execute. It programmatically
// constructs AnalysisDefinitions matching a minimal CPL structure with one
// metric (literal 0), generates test parquet input data, synthesizes Go code,
// compiles and runs it with `go run`, then reads the output parquet and
// verifies all metric values are 0.0.
func TestEndToEndSynthesizeCompileRun(t *testing.T) {
	// Check that `go` is available in PATH
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("'go' command not available in PATH: %v", err)
	}

	// Create a temp directory within the module so imports resolve correctly.
	// The synthesized code imports packages from this module, so it must
	// live under the module root.
	moduleRoot, err := findModuleRoot()
	if err != nil {
		t.Fatalf("cannot find module root: %v", err)
	}

	testDir, err := os.MkdirTemp(filepath.Join(moduleRoot, "generated", "cmd"), "e2e_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp test dir under module root: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(testDir) })

	// Create data directory for test parquet input
	dataDir := filepath.Join(testDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("Failed to create data dir: %v", err)
	}

	// Generate test parquet input file with dimensions matching the CPL structure.
	// The CPL defines: concepts Region[North, South], Day[Monday]
	// connector demand_data = dimensions [Region] grain Daily celltype demand_forecast
	// The connector loader expects parquet columns for dimension names AND grain
	// measure names. So we include "Daily" (the grain measure name) as a column.
	parquetPath := filepath.Join(dataDir, "demand_data.parquet")
	inputDimensions := map[string][]string{
		"Region": {"North", "South"},
		"Daily":  {"Monday"},
	}
	if err := cio.GenerateParquetFile(parquetPath, inputDimensions, nil, nil); err != nil {
		t.Fatalf("Failed to generate test parquet file: %v", err)
	}

	// Build AnalysisDefinitions matching the minimal CPL file structure
	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Region": {
				ConceptName: "Region",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "North", Position: 0},
					{Name: "South", Position: 1},
				},
			},
			"Day": {
				ConceptName: "Day",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "Monday", Position: 0},
				},
			},
		},
		Measures: map[string]analysis.AnalysisMeasureDefinition{
			"demand_forecast": {
				Name:     "demand_forecast",
				TypeName: "Number",
				UnitExpression: &analysis.UnitReference{Name: "Region"},
			},
			"Daily": {
				Name:     "Daily",
				TypeName: "Interval",
				UnitExpression: &analysis.UnitReference{Name: "Day"},
			},
		},
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"demand_data": {
				Name: "demand_data",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Region"},
				},
				CellType: &analysis.AnalysisMeasureDefinition{
					Name:     "demand_forecast",
					TypeName: "Number",
				},
				Grain: []*analysis.AnalysisMeasureDefinition{
					{
						Name:     "Daily",
						TypeName: "Interval",
						UnitExpression: &analysis.UnitReference{Name: "Day"},
					},
				},
			},
		},
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"test_metric": {
				Name: "test_metric",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "demand_forecast"},
					{Kind: "dimension", Name: "Region"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode:   &cil.CILLiteral{Value: 0},
					InputRefs: []cil.CILInputRef{{Kind: "dimension", Name: "Region"}},
				},
			},
		},
	}

	// Build IO spec with test_metric as output driving the parquet path to point
	// at the generated test input data.
	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "demand_data",
				DataFormat:    "parquet",
				Driver:        "file",
				DriverParameters: []map[string]interface{}{
					{"path": parquetPath},
				},
			},
		},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "test_metric",
				Dimensions: []string{"Region"},
				CellType:   "float64",
				Grain:      []string{}, // No grain; metric tensor only has Region dimension
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  8388608,
			MainMemoryBytes: 17179869184,
		},
	}

	// Synthesize Go code
	outputPath := filepath.Join(testDir, "main.go")
	if err := Synthesize(analysisDefs, ioSpec, outputPath); err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	// Run the generated program with `go run`. The generated code writes
	// output parquet files relative to its working directory, so set it
	// to testDir so we can find the output.
	cmd := exec.Command("go", "run", outputPath)
	cmd.Dir = testDir
	runOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Generated program failed to run: %v\nOutput:\n%s", err, string(runOutput))
	}

	// Verify the output parquet file was created (now test_metric.parquet)
	resultParquetPath := filepath.Join(testDir, "test_metric.parquet")
	resultParquet, numRows, err := cio.ReadParquetFile(resultParquetPath)
	if err != nil {
		t.Fatalf("Failed to read output parquet file: %v", err)
	}

	// Expected rows: Region=[North,South] × Day=[Monday] = 2 rows
	expectedRows := 2
	if numRows != expectedRows {
		t.Errorf("Expected %d rows in output parquet, got %d", expectedRows, numRows)
	}

	// Verify schema has expected columns: Region, Value (no grain)
	colNames := cio.GetColumnNames(resultParquet)
	hasRegion := false
	hasValue := false
	for _, name := range colNames {
		switch name {
		case "Region":
			hasRegion = true
		case "Value":
			hasValue = true
		}
	}
	if !hasRegion {
		t.Errorf("Output parquet missing 'Region' column. Columns: %v", colNames)
	}
	if !hasValue {
		t.Errorf("Output parquet missing 'Value' column. Columns: %v", colNames)
	}

	// Read the Value column and verify all values are 0.0
	valueColIndex := -1
	for i, name := range colNames {
		if name == "Value" {
			valueColIndex = i
			break
		}
	}
	if valueColIndex < 0 {
		t.Fatalf("Could not find Value column in output parquet")
	}

	rawValues, err := cio.ReadParquetColumnData(resultParquetPath, valueColIndex)
	if err != nil {
		t.Fatalf("Failed to read Value column: %v", err)
	}

	values, ok := rawValues.([]float64)
	if !ok {
		t.Fatalf("Expected Value column to be []float64, got %T", rawValues)
	}

	if len(values) != expectedRows {
		t.Errorf("Expected %d values, got %d", expectedRows, len(values))
	}

	for i, v := range values {
		if v != 0.0 {
			t.Errorf("Value[%d] = %f, expected 0.0", i, v)
		}
	}
}

// findModuleRoot locates the Go module root directory by searching upward
// from the source file's location for a go.mod file. This is needed so
// the end-to-end test can place generated code inside the module for proper
// import resolution. Using runtime.Caller(0) ensures the search starts from
// the test file's directory, independent of the current working directory.
func findModuleRoot() (string, error) {
	// Locate this source file to start the search from a known location
	// rather than relying on os.Getwd() which varies by test invocation.
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("cannot determine source file location")
	}
	dir := filepath.Dir(filename)

	// Walk up from the source file's directory looking for go.mod
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if info, err := os.Stat(goModPath); err == nil && !info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// makeDayRange returns a slice of day strings from start to end inclusive.
func makeDayRange(start, end int) []string {
	result := make([]string, 0, end-start+1)
	for i := start; i <= end; i++ {
		result = append(result, fmt.Sprintf("%d", i))
	}
	return result
}

// TestEndToEndSynthesizeCompileRun_ConditionalGrain verifies the full
// end-to-end pipeline with conditional concept hierarchies. It constructs
// AnalysisDefinitions with Month and conditional DayInMonth, generates a
// conditional parquet file with 90 valid rows, synthesizes Go code, compiles
// and runs it, then verifies the output parquet contains the expected 90
// metric values.
func TestEndToEndSynthesizeCompileRun_ConditionalGrain(t *testing.T) {
	// Check that `go` is available in PATH
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("'go' command not available in PATH: %v", err)
	}

	moduleRoot, err := findModuleRoot()
	if err != nil {
		t.Fatalf("cannot find module root: %v", err)
	}

	testDir, err := os.MkdirTemp(filepath.Join(moduleRoot, "generated", "cmd"), "e2e_conditional_*")
	if err != nil {
		t.Fatalf("Failed to create temp test dir under module root: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(testDir) })

	dataDir := filepath.Join(testDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("Failed to create data dir: %v", err)
	}

	// Generate conditional parquet input with 90 rows (31+28+31).
	// DayInMonth and Monthly are both conditional on Month so that each row's
	// Monthly value matches its Month value.
	parquetPath := filepath.Join(dataDir, "monthly_data.parquet")
	dimensions := map[string][]string{
		"Month":      {"Jan", "Feb", "Mar"},
		"DayInMonth": makeDayRange(1, 31),
	}
	grain := []cio.GrainMeasure{
		{Name: "Monthly", Values: []string{"Jan", "Feb", "Mar"}},
	}
	values := make([]float64, 90)
	for i := range values {
		values[i] = float64(i + 1)
	}
	conditionalDims := map[string]cio.ConditionalDimMapping{
		"DayInMonth": {
			Parent: "Month",
			Values: map[string][]string{
				"Jan": makeDayRange(1, 31),
				"Feb": makeDayRange(1, 28),
				"Mar": makeDayRange(1, 31),
			},
		},
		"Monthly": {
			Parent: "Month",
			Values: map[string][]string{
				"Jan": {"Jan"},
				"Feb": {"Feb"},
				"Mar": {"Mar"},
			},
		},
	}
	if err := cio.GenerateParquetFileWithConditionals(parquetPath, dimensions, grain, values, conditionalDims); err != nil {
		t.Fatalf("Failed to generate conditional parquet file: %v", err)
	}

	// Build AnalysisDefinitions
	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Month": {
				ConceptName: "Month",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "Jan", Position: 0},
					{Name: "Feb", Position: 1},
					{Name: "Mar", Position: 2},
				},
			},
			"DayInMonth": {
				ConceptName:       "DayInMonth",
				ParentConceptName: strPtr("Month"),
				ConditionalValues: map[string][]string{
					"Jan": makeDayRange(1, 31),
					"Feb": makeDayRange(1, 28),
					"Mar": makeDayRange(1, 31),
				},
				ResolvedValues: func() []analysis.AnalysisConceptValueDefinition {
					vals := makeDayRange(1, 31)
					defs := make([]analysis.AnalysisConceptValueDefinition, len(vals))
					for i, v := range vals {
						defs[i] = analysis.AnalysisConceptValueDefinition{Name: v, Position: i}
					}
					return defs
				}(),
			},
		},
		Measures: map[string]analysis.AnalysisMeasureDefinition{
			"Monthly": {
				Name:     "Monthly",
				TypeName: "Interval",
				UnitExpression: &analysis.UnitReference{Name: "Month"},
			},
			"data_count": {
				Name:     "data_count",
				TypeName: "Number",
				UnitExpression: &analysis.UnitReference{Name: "Month"},
			},
		},
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"monthly_data": {
				Name: "monthly_data",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Month"},
					{ConceptName: "DayInMonth"},
				},
				CellType: &analysis.AnalysisMeasureDefinition{
					Name:     "data_count",
					TypeName: "Number",
				},
				Grain: []*analysis.AnalysisMeasureDefinition{
					{
						Name:     "Monthly",
						TypeName: "Interval",
						UnitExpression: &analysis.UnitReference{Name: "Month"},
					},
				},
			},
		},
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"data_count": {
				Name: "data_count",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "data_count"},
					{Kind: "dimension", Name: "Month"},
					{Kind: "dimension", Name: "DayInMonth"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode:   &cil.CILInputRef{Kind: "connector", Name: "monthly_data"},
					InputRefs: []cil.CILInputRef{{Kind: "connector", Name: "monthly_data"}},
				},
			},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "monthly_data",
				DataFormat:    "parquet",
				Driver:        "file",
				DriverParameters: []map[string]interface{}{
					{"path": parquetPath},
				},
			},
		},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "data_count",
				Dimensions: []string{"Month", "DayInMonth"},
				CellType:   "float64",
				Grain:      []string{"Monthly"},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  8388608,
			MainMemoryBytes: 17179869184,
		},
	}

	outputPath := filepath.Join(testDir, "main.go")
	if err := Synthesize(analysisDefs, ioSpec, outputPath); err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	cmd := exec.Command("go", "run", outputPath)
	cmd.Dir = testDir
	runOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Generated program failed to run: %v\nOutput:\n%s", err, string(runOutput))
	}

	resultParquetPath := filepath.Join(testDir, "data_count.parquet")
	resultParquet, numRows, err := cio.ReadParquetFile(resultParquetPath)
	if err != nil {
		t.Fatalf("Failed to read output parquet file: %v", err)
	}

	// The dense tensor output includes the grain dimension (Monthly) alongside
	// the metric dimensions (Month × DayInMonth), producing 3×31×3 = 279 rows.
	// Of these, 3 cells per invalid (Month, DayInMonth) combination are zero.
	if numRows != 279 {
		t.Errorf("Expected 279 rows (dense 3×31 tensor with Monthly grain), got %d", numRows)
	}

	colNames := cio.GetColumnNames(resultParquet)
	hasMonth := false
	hasDayInMonth := false
	hasMonthly := false
	hasValue := false
	for _, name := range colNames {
		switch name {
		case "Month":
			hasMonth = true
		case "DayInMonth":
			hasDayInMonth = true
		case "Monthly":
			hasMonthly = true
		case "Value":
			hasValue = true
		}
	}
	if !hasMonth {
		t.Errorf("Output parquet missing 'Month' column. Columns: %v", colNames)
	}
	if !hasDayInMonth {
		t.Errorf("Output parquet missing 'DayInMonth' column. Columns: %v", colNames)
	}
	if !hasMonthly {
		t.Errorf("Output parquet missing 'Monthly' column. Columns: %v", colNames)
	}
	if !hasValue {
		t.Errorf("Output parquet missing 'Value' column. Columns: %v", colNames)
	}

	valueColIndex := -1
	for i, name := range colNames {
		if name == "Value" {
			valueColIndex = i
			break
		}
	}
	if valueColIndex < 0 {
		t.Fatalf("Could not find Value column in output parquet")
	}

	rawValues, err := cio.ReadParquetColumnData(resultParquetPath, valueColIndex)
	if err != nil {
		t.Fatalf("Failed to read Value column: %v", err)
	}

	outputValues, ok := rawValues.([]float64)
	if !ok {
		t.Fatalf("Expected Value column to be []float64, got %T", rawValues)
	}

	// The dense tensor representation may include extra zero rows for invalid
	// (Month, DayInMonth, Monthly) combinations. Collect all non-zero values
	// and verify they are exactly 1.0..90.0 in order.
	var nonZeroValues []float64
	for _, v := range outputValues {
		if v != 0.0 {
			nonZeroValues = append(nonZeroValues, v)
		}
	}

	if len(nonZeroValues) != 90 {
		t.Errorf("Expected 90 non-zero metric values, got %d (total rows: %d)", len(nonZeroValues), numRows)
	}

	for i, v := range nonZeroValues {
		expected := float64(i + 1)
		if v != expected {
			t.Errorf("Non-zero value[%d] = %f, expected %f", i, v, expected)
		}
	}
}

// TestGenerateMainFunction_MetricWiring verifies that GenerateMainFunction
// wires metric names into the generated lifecycle functions correctly.
// These tests confirm metric registration, goroutine launches, and lifecycle
// function presence in the generated code.
func TestGenerateMainFunction_MetricWiring(t *testing.T) {
	tests := []struct {
		name           string
		ioSpec         *cio.IOSpecification
		expectContains []string
		expectMissing  []string
	}{
		{
			name: "outputs with metrics contain registration loop",
			ioSpec: &cio.IOSpecification{
				Connectors: []cio.ConnectorSpecification{},
				Outputs: []cio.OutputSpecification{
					{
						MetricName: "total_sales",
						Dimensions: []string{"Region"},
						CellType:   "float64",
						Grain:      []string{"Region"},
					},
					{
						MetricName: "avg_price",
						Dimensions: []string{"Product"},
						CellType:   "float64",
						Grain:      []string{"Product"},
					},
				},
				Architecture: cio.ArchitectureSpecification{
					CPUCoreCount:    4,
					CacheSizeBytes:  1024 * 1024,
					MainMemoryBytes: 1024 * 1024 * 1024,
				},
			},
			expectContains: []string{
				"var metricNames = []string{",
				"\"total_sales\"",
				"\"avg_price\"",
				"controller.RegisterMetricComputation(metricName, mc)",
				"func Setup",
				"func Compute",
				"func Cleanup",
			},
			expectMissing: []string{
				"// TODO: Metric registration will be implemented in Phase C",
				"_ = metricNames",
			},
		},
		{
			name: "no outputs still contains lifecycle functions",
			ioSpec: &cio.IOSpecification{
				Connectors:  []cio.ConnectorSpecification{},
				Outputs: []cio.OutputSpecification{},
				Architecture: cio.ArchitectureSpecification{
					CPUCoreCount:    4,
					CacheSizeBytes:  1024 * 1024,
					MainMemoryBytes: 1024 * 1024 * 1024,
				},
			},
			expectContains: []string{
				"func Setup",
				"func Compute",
				"func Cleanup",
			},
			expectMissing: []string{
				"RegisterMetricComputation",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := cil.NewCILAbstractTransformer()
			if transformer == nil {
				t.Fatalf("NewCILAbstractTransformer() returned nil")
			}

			code, err := GenerateMainFunction(&SynthesisInput{Transformer: transformer, IOSpec: tt.ioSpec})
			if err != nil {
				t.Fatalf("GenerateMainFunction() returned error: %v", err)
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(code, expected) {
					t.Errorf("GenerateMainFunction() output missing expected content: %q", expected)
				}
			}

			for _, unexpected := range tt.expectMissing {
				if strings.Contains(code, unexpected) {
					t.Errorf("GenerateMainFunction() output contains unexpected content: %q", unexpected)
				}
			}
		})
	}
}

// TestGenerateMainFunction_OutputMetricsOnly verifies that compute functions
// are only generated for metrics that appear in the IO spec's output section.
// Metrics defined in analysisDefs but not listed as outputs should NOT have
// compute functions generated.
func TestGenerateMainFunction_OutputMetricsOnly(t *testing.T) {
	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Region": {
				ConceptName: "Region",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "North", Position: 0},
					{Name: "South", Position: 1},
				},
			},
		},
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"revenue": {
				Name: "revenue",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "dimension", Name: "Region"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode:   &cil.CILLiteral{Value: 100},
					InputRefs: []cil.CILInputRef{{Kind: "dimension", Name: "Region"}},
				},
			},
			"profit": {
				Name: "profit",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "dimension", Name: "Region"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode:   &cil.CILLiteral{Value: 50},
					InputRefs: []cil.CILInputRef{{Kind: "dimension", Name: "Region"}},
				},
			},
			"costs": {
				Name: "costs",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "dimension", Name: "Region"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode:   &cil.CILLiteral{Value: 30},
					InputRefs: []cil.CILInputRef{{Kind: "dimension", Name: "Region"}},
				},
			},
		},
	}

	// IO spec only lists revenue and profit as outputs; costs is NOT an output
	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "revenue",
				Dimensions: []string{"Region"},
				CellType:   "float64",
				Grain:      []string{"Region"},
			},
			{
				MetricName: "profit",
				Dimensions: []string{"Region"},
				CellType:   "float64",
				Grain:      []string{"Region"},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	input, err := BuildTransformerFromAnalysis(analysisDefs)
	if err != nil {
		t.Fatalf("BuildTransformerFromAnalysis() returned error: %v", err)
	}
	input.IOSpec = ioSpec
	code, err := GenerateMainFunction(input)
	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	// Output metrics should have compute functions
	if !strings.Contains(code, "computerevenue(controller, inputs)") {
		t.Errorf("Generated code missing 'computerevenue(controller, inputs)' for output metric")
	}
	if !strings.Contains(code, "computeprofit(controller, inputs)") {
		t.Errorf("Generated code missing 'computeprofit(controller, inputs)' for output metric")
	}

	// Non-output metric should NOT have a compute function
	if strings.Contains(code, "computecosts(controller, inputs)") {
		t.Errorf("Generated code contains 'computecosts(controller, inputs)' for non-output metric")
	}

	// Verify metricNames only contains output metrics
	if !strings.Contains(code, `"revenue"`) {
		t.Errorf("Generated code missing '\"revenue\"' in metricNames")
	}
	if !strings.Contains(code, `"profit"`) {
		t.Errorf("Generated code missing '\"profit\"' in metricNames")
	}
	if strings.Contains(code, `"costs"`) {
		t.Errorf("Generated code contains '\"costs\"' in metricNames for non-output metric")
	}
}

// TestGenerateMainFunction_NoOutputsNoComputeFunctions verifies that when the
// IO spec has no outputs, no compute functions are generated even if
// analysisDefs contains metrics with CIL roots.
func TestGenerateMainFunction_NoOutputsNoComputeFunctions(t *testing.T) {
	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Region": {
				ConceptName: "Region",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "North", Position: 0},
					{Name: "South", Position: 1},
				},
			},
		},
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"revenue": {
				Name: "revenue",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "dimension", Name: "Region"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode:   &cil.CILLiteral{Value: 100},
					InputRefs: []cil.CILInputRef{{Kind: "dimension", Name: "Region"}},
				},
			},
			"profit": {
				Name: "profit",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "dimension", Name: "Region"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode:   &cil.CILLiteral{Value: 50},
					InputRefs: []cil.CILInputRef{{Kind: "dimension", Name: "Region"}},
				},
			},
		},
	}

	// IO spec has NO outputs
	ioSpec := &cio.IOSpecification{
		Connectors:  []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	input, err := BuildTransformerFromAnalysis(analysisDefs)
	if err != nil {
		t.Fatalf("BuildTransformerFromAnalysis() returned error: %v", err)
	}
	input.IOSpec = ioSpec
	code, err := GenerateMainFunction(input)
	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	// No compute functions should be generated
	if strings.Contains(code, "computerevenue(controller, inputs)") {
		t.Errorf("Generated code contains 'computerevenue(controller, inputs)' when no outputs specified")
	}
	if strings.Contains(code, "computeprofit(controller, inputs)") {
		t.Errorf("Generated code contains 'computeprofit(controller, inputs)' when no outputs specified")
	}

	// metricNames variable should not be present
	if strings.Contains(code, "var metricNames = []string{") {
		t.Errorf("Generated code contains 'var metricNames' when no outputs specified")
	}

	// Core lifecycle functions should still be present
	if !strings.Contains(code, "func Setup") {
		t.Errorf("Generated code missing 'func Setup'")
	}
	if !strings.Contains(code, "func Compute") {
		t.Errorf("Generated code missing 'func Compute'")
	}
	if !strings.Contains(code, "func Cleanup") {
		t.Errorf("Generated code missing 'func Cleanup'")
	}
}

// TestSynthesize_OutputMetricsFiltering is an integration test that verifies
// the full Synthesize() pipeline only generates compute functions for metrics
// listed in the IO spec outputs. It creates analysis definitions with three
// metrics but only two in the IO spec outputs, then validates the generated
// code includes compute functions only for the output metrics.
func TestSynthesize_OutputMetricsFiltering(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.go")

	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Region": {
				ConceptName: "Region",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "North", Position: 0},
					{Name: "South", Position: 1},
				},
			},
		},
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"sales_data": {
				Name: "sales_data",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Region"},
				},
				CellType: &analysis.AnalysisMeasureDefinition{
					Name:     "Amount",
					TypeName: "Number",
				},
			},
		},
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"revenue": {
				Name: "revenue",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "Amount"},
					{Kind: "dimension", Name: "Region"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode:   &cil.CILLiteral{Value: 100},
					InputRefs: []cil.CILInputRef{{Kind: "dimension", Name: "Region"}},
				},
			},
			"profit": {
				Name: "profit",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "metric", Name: "Amount"},
					{Kind: "dimension", Name: "Region"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode:   &cil.CILLiteral{Value: 50},
					InputRefs: []cil.CILInputRef{{Kind: "dimension", Name: "Region"}},
				},
			},
			"internal_calc": {
				Name: "internal_calc",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "dimension", Name: "Region"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode:   &cil.CILLiteral{Value: 0},
					InputRefs: []cil.CILInputRef{{Kind: "dimension", Name: "Region"}},
				},
			},
		},
	}

	// Only revenue and profit are outputs; internal_calc is not
	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "sales_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "revenue",
				Dimensions: []string{"Region"},
				CellType:   "float64",
				Grain:      []string{"Region"},
			},
			{
				MetricName: "profit",
				Dimensions: []string{"Region"},
				CellType:   "float64",
				Grain:      []string{"Region"},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	err := Synthesize(analysisDefs, ioSpec, outputPath)
	if err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Output metrics should have compute functions
	if !strings.Contains(contentStr, "computerevenue(controller, inputs)") {
		t.Errorf("Generated file missing 'computerevenue(controller, inputs)' for output metric")
	}
	if !strings.Contains(contentStr, "computeprofit(controller, inputs)") {
		t.Errorf("Generated file missing 'computeprofit(controller, inputs)' for output metric")
	}

	// Non-output metric should NOT have a compute function
	if strings.Contains(contentStr, "computeinternal_calc(controller, inputs)") {
		t.Errorf("Generated file contains 'computeinternal_calc(controller, inputs)' for non-output metric")
	}

	// metricNames should only contain output metrics
	if !strings.Contains(contentStr, `"revenue"`) {
		t.Errorf("Generated file missing '\"revenue\"' in metricNames")
	}
	if !strings.Contains(contentStr, `"profit"`) {
		t.Errorf("Generated file missing '\"profit\"' in metricNames")
	}
	if strings.Contains(contentStr, `"internal_calc"`) {
		t.Errorf("Generated file contains '\"internal_calc\"' in metricNames for non-output metric")
	}

	// Core lifecycle functions should be present
	expectedFunctions := []string{
		"func main()",
		"func Setup",
		"func Compute",
		"func Cleanup",
	}
	for _, expected := range expectedFunctions {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Generated file missing expected function: %q", expected)
		}
	}
}

// TestGenerateInputLoadingCode_ConnectorNameKey verifies that the inputs map
// key uses the connector name (e.g., "revenue_amt") rather than the CellType
// measure name (e.g., "Revenue"). CIL trees and the postfix evaluator reference
// connectors by their connector name, so the key must match for runtime lookup
// to succeed.
func TestGenerateInputLoadingCode_ConnectorNameKey(t *testing.T) {
	// Create IO spec with an input referencing this connector
	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "revenue_amt",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs:      []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{},
	}

	// Generate the input loading code
	code := generateConnectorLoadingCode(ioSpec)

	// Verify the inputs map key uses the connector name, not the CellType measure name
	if !strings.Contains(code, `inputs["revenue_amt"]`) {
		t.Errorf("Generated code missing expected key inputs[%q]; got:\n%s", "revenue_amt", code)
	}
	if strings.Contains(code, `inputs["Revenue"]`) {
		t.Errorf("Generated code incorrectly uses CellType measure name as key (inputs[%q]); got:\n%s", "Revenue", code)
	}
}

// TestGenerateMainFunction_EmitsConnectorValidityMasks verifies that when
// analysis definitions contain conditional concepts and connectors with grain
// measures matching those concepts, GenerateMainFunction emits the
// connectorValidityMasks variable, hex-encoded mask values, and tensor.SetMask
// calls inside loader functions.
func TestGenerateMainFunction_EmitsConnectorValidityMasks(t *testing.T) {
	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Month": {
				ConceptName: "Month",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "Jan", Position: 0},
					{Name: "Feb", Position: 1},
				},
			},
			"DayInMonth": {
				ConceptName:       "DayInMonth",
				ParentConceptName: strPtr("Month"),
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "1", Position: 0},
					{Name: "2", Position: 1},
					{Name: "3", Position: 2},
				},
				ConditionalValues: map[string][]string{
					"Jan": {"1", "2", "3"},
					"Feb": {"1", "2"},
				},
			},
		},
		Measures: map[string]analysis.AnalysisMeasureDefinition{
			"Month":      {Name: "Month", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Month"}},
			"DayInMonth": {Name: "DayInMonth", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "DayInMonth"}},
		},
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"monthly_data": {
				Name: "monthly_data",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Month"},
				},
				Grain: []*analysis.AnalysisMeasureDefinition{
					{Name: "Month", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "Month"}},
					{Name: "DayInMonth", TypeName: "Interval", UnitExpression: &analysis.UnitReference{Name: "DayInMonth"}},
				},
			},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "monthly_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs:      []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{},
	}

	transformer := cil.NewCILAbstractTransformer()
	transformer.RegisterConnector("monthly_data", []string{"Month"}, []string{"Month", "DayInMonth"}, map[string]string{"Month": "Month", "DayInMonth": "DayInMonth"}, "float64")

	input := &SynthesisInput{
		Transformer: transformer,
		IOSpec:      ioSpec,
		Concepts:    analysisDefs.Concepts,
		Measures:    analysisDefs.Measures,
		Connectors:  analysisDefs.Connectors,
	}
	code, err := GenerateMainFunction(input)
	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	if !strings.Contains(code, "var connectorValidityMasks = map[string][]uint64{") {
		t.Errorf("Generated code missing connectorValidityMasks declaration")
	}
	if !strings.Contains(code, "0x") {
		t.Errorf("Generated code missing hex-encoded mask values")
	}
	if !strings.Contains(code, "tensor.SetMask(mask)") {
		t.Errorf("Generated code missing tensor.SetMask(mask)")
	}
}

// TestSynthesisInputContract verifies that SynthesisInput serves as the stable
// contract between analysis and synthesis. It tests that:
//  1. BuildTransformerFromAnalysis returns a *SynthesisInput with all expected
//     fields populated from AnalysisDefinitions.
//  2. GenerateMainFunction accepts *SynthesisInput and extracts data from it.
//  3. BuildConceptValuesMap works as a method on SynthesisInput.
func TestSynthesisInputContract(t *testing.T) {
	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Year": {
				ConceptName: "Year",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "2023", Position: 0},
					{Name: "2024", Position: 1},
				},
			},
		},
		Measures: map[string]analysis.AnalysisMeasureDefinition{
			"Amount": {Name: "Amount", TypeName: "Number"},
		},
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"SalesData": {
				Name: "SalesData",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Year"},
				},
				CellType: &analysis.AnalysisMeasureDefinition{
					Name:     "Amount",
					TypeName: "Number",
				},
			},
		},
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"total_sales": {
				Name: "total_sales",
				CILRoot: &cil.CILMetricNode{
					CILNode: &cil.CILLiteral{Value: 0},
				},
			},
		},
	}

	// Step 1: BuildTransformerFromAnalysis returns *SynthesisInput
	input, err := BuildTransformerFromAnalysis(analysisDefs)
	if err != nil {
		t.Fatalf("BuildTransformerFromAnalysis() returned error: %v", err)
	}
	if input == nil {
		t.Fatal("BuildTransformerFromAnalysis() returned nil")
	}
	if input.Transformer == nil {
		t.Error("BuildTransformerFromAnalysis() returned SynthesisInput with nil Transformer")
	}
	if len(input.Connectors) != 1 {
		t.Errorf("Expected 1 connector, got %d", len(input.Connectors))
	}
	if len(input.Measures) != 1 {
		t.Errorf("Expected 1 measure, got %d", len(input.Measures))
	}
	if len(input.Metrics) != 1 {
		t.Errorf("Expected 1 metric, got %d", len(input.Metrics))
	}
	if len(input.Concepts) != 1 {
		t.Errorf("Expected 1 concept, got %d", len(input.Concepts))
	}

	// Verify transformer was populated
	meta := input.Transformer.GetConnectorMetadata("SalesData")
	if meta == nil {
		t.Error("Expected SalesData connector to be registered on transformer")
	}

	// Step 2: GenerateMainFunction accepts *SynthesisInput directly
	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "total_sales",
				Dimensions: []string{"Year"},
				CellType:   "float64",
				Grain:      []string{"Year"},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}
	input.IOSpec = ioSpec

	code, err := GenerateMainFunction(input)
	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	// Verify synthesis extracted metric data from SynthesisInput
	if !strings.Contains(code, `"total_sales"`) {
		t.Errorf("GenerateMainFunction() output missing total_sales metric")
	}
	if !strings.Contains(code, "var conceptValues = map[string][]string") {
		t.Errorf("GenerateMainFunction() output missing conceptValues declaration")
	}
	if !strings.Contains(code, `"2023"`) {
		t.Errorf("GenerateMainFunction() output missing concept value from SynthesisInput")
	}

	// Step 3: BuildConceptValuesMap works as a method on SynthesisInput
	conceptValues := input.BuildConceptValuesMap()
	if len(conceptValues) != 1 {
		t.Errorf("BuildConceptValuesMap() returned %d entries, expected 1", len(conceptValues))
	}
	if _, ok := conceptValues["Year"]; !ok {
		t.Errorf("BuildConceptValuesMap() missing Year concept")
	}
}

// TestSynthesisInputNilSafety verifies that methods on SynthesisInput handle
// nil input gracefully without panicking.
func TestSynthesisInputNilSafety(t *testing.T) {
	var input *SynthesisInput

	conceptValues := input.BuildConceptValuesMap()
	if conceptValues == nil {
		t.Error("BuildConceptValuesMap() on nil input returned nil instead of empty map")
	}
	if len(conceptValues) != 0 {
		t.Errorf("BuildConceptValuesMap() on nil input returned %d entries, expected 0", len(conceptValues))
	}

	conditionalValues := input.BuildConditionalConceptValuesMap(nil)
	if conditionalValues == nil {
		t.Error("BuildConditionalConceptValuesMap() on nil input returned nil instead of empty map")
	}
	if len(conditionalValues) != 0 {
		t.Errorf("BuildConditionalConceptValuesMap() on nil input returned %d entries, expected 0", len(conditionalValues))
	}

	masks, err := input.BuildConnectorValidityMasks(nil, nil)
	if err != nil {
		t.Errorf("BuildConnectorValidityMasks() on nil input returned error: %v", err)
	}
	if masks == nil {
		t.Error("BuildConnectorValidityMasks() on nil input returned nil instead of empty map")
	}
	if len(masks) != 0 {
		t.Errorf("BuildConnectorValidityMasks() on nil input returned %d entries, expected 0", len(masks))
	}
}

// TestGenerateInputLoadingCode verifies that generateConnectorLoadingCode produces
// the correct input loading code for various IO specifications.
func TestGenerateInputLoadingCode(t *testing.T) {
	tests := []struct {
		name           string
		ioSpec         *cio.IOSpecification
		expectContains []string
		expectMissing  []string
	}{
		{
			name:           "nil ioSpec returns empty string",
			ioSpec:         nil,
			expectContains: []string{},
			expectMissing:  []string{},
		},
		{
			name: "empty inputs returns empty string",
			ioSpec: &cio.IOSpecification{
				Connectors:  []cio.ConnectorSpecification{},
				Outputs: []cio.OutputSpecification{},
			},
			expectContains: []string{},
			expectMissing:  []string{},
		},
		{
			name: "single input generates loader dispatch code",
			ioSpec: &cio.IOSpecification{
				Connectors: []cio.ConnectorSpecification{
					{
						ConnectorName: "sales_data",
						DataFormat:    "parquet",
						Driver:        "file",
					},
				},
			},
			expectContains: []string{
				"cplio.ConnectorSpecification",
				"loadSalesData",
				`inputs["sales_data"]`,
				"check(err, \"Failed to load sales_data\")",
			},
			expectMissing: []string{
				"inventoryInputSpec",
				"os.Stderr",
				"exitCode",
			},
		},
		{
			name: "multiple inputs generate separate loader calls",
			ioSpec: &cio.IOSpecification{
				Connectors: []cio.ConnectorSpecification{
					{
						ConnectorName: "sales_data",
						DataFormat:    "parquet",
						Driver:        "file",
					},
					{
						ConnectorName: "inventory",
						DataFormat:    "parquet",
						Driver:        "file",
					},
				},
			},
			expectContains: []string{
				"sales_dataConnectorSpec",
				"loadSalesData",
				`inputs["sales_data"]`,
				"inventoryConnectorSpec",
				"loadInventory",
				`inputs["inventory"]`,
			},
			expectMissing: []string{},
		},
		{
			name: "input with driver parameters generates parameter literals",
			ioSpec: &cio.IOSpecification{
				Connectors: []cio.ConnectorSpecification{
					{
						ConnectorName: "sales_data",
						DataFormat:    "parquet",
						Driver:        "file",
						DriverParameters: []map[string]interface{}{
							{
								"host":    "localhost",
								"timeout": 30.5,
								"workers": 4,
								"async":   true,
							},
						},
					},
				},
			},
			expectContains: []string{
				"DriverParameters: []map[string]interface{}",
				`"host"`,
				`"localhost"`,
				`"timeout"`,
				"30.5",
				`"workers"`,
				"4",
				`"async"`,
				"true",
			},
			expectMissing: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := generateConnectorLoadingCode(tt.ioSpec)
			if tt.ioSpec == nil || len(tt.ioSpec.Connectors) == 0 {
				if code != "" {
					t.Errorf("generateConnectorLoadingCode() = %q, expected empty string", code)
				}
			}
			for _, expected := range tt.expectContains {
				if !strings.Contains(code, expected) {
					t.Errorf("generateConnectorLoadingCode() output missing expected content: %q", expected)
				}
			}
			for _, unexpected := range tt.expectMissing {
				if strings.Contains(code, unexpected) {
					t.Errorf("generateConnectorLoadingCode() output contains unexpected content: %q", unexpected)
				}
			}
		})
	}
}

// TestGenerateOutputWritingCode verifies that generateOutputWritingCode produces
// the correct output writing code for various IO specifications.
func TestGenerateOutputWritingCode(t *testing.T) {
	tests := []struct {
		name           string
		ioSpec         *cio.IOSpecification
		expectContains []string
		expectMissing  []string
	}{
		{
			name:           "nil ioSpec returns empty string",
			ioSpec:         nil,
			expectContains: []string{},
			expectMissing:  []string{},
		},
		{
			name: "empty outputs returns empty string",
			ioSpec: &cio.IOSpecification{
				Outputs: []cio.OutputSpecification{},
			},
			expectContains: []string{},
			expectMissing:  []string{},
		},
		{
			name: "single output generates writing code",
			ioSpec: &cio.IOSpecification{
				Outputs: []cio.OutputSpecification{
					{
						MetricName: "total_sales",
						Dimensions: []string{"Region"},
						CellType:   "float64",
						Grain:      []string{"Region"},
					},
				},
			},
			expectContains: []string{
				"total_salesComp",
				"total_salesTensor",
				"total_salesRaw",
				"total_salesData",
				`controller.GetMetricComputation("total_sales")`,
				"cplio.WriteMetricsToParquet",
				"total_sales.parquet",
				`check(err, "Failed to write output metric total_sales")`,
				"Type assertion failed",
			},
			expectMissing: []string{
				"exitCode",
			},
		},
		{
			name: "multiple outputs generate separate parquet writes",
			ioSpec: &cio.IOSpecification{
				Outputs: []cio.OutputSpecification{
					{
						MetricName: "revenue",
						Dimensions: []string{"Region"},
						CellType:   "float64",
						Grain:      []string{"Region"},
					},
					{
						MetricName: "costs",
						Dimensions: []string{"Product"},
						CellType:   "float64",
						Grain:      []string{"Product"},
					},
				},
			},
			expectContains: []string{
				"revenueComp",
				"revenueTensor",
				"costsComp",
				"costsTensor",
				"cplio.WriteMetricsToParquet",
				"revenue.parquet",
				"costs.parquet",
			},
			expectMissing: []string{},
		},
		{
			name: "output without grain uses nil",
			ioSpec: &cio.IOSpecification{
				Outputs: []cio.OutputSpecification{
					{
						MetricName: "total_sales",
						Dimensions: []string{"Region"},
						CellType:   "float64",
						Grain:      nil,
					},
				},
			},
			expectContains: []string{
				"Grain: nil",
			},
			expectMissing: []string{
				"[]string{}",
			},
		},
		{
			name: "output with grain generates grain slice",
			ioSpec: &cio.IOSpecification{
				Outputs: []cio.OutputSpecification{
					{
						MetricName: "total_sales",
						Dimensions: []string{"Region"},
						CellType:   "float64",
						Grain:      []string{"Region", "Product"},
					},
				},
			},
			expectContains: []string{
				`Grain: []string{"Region", "Product"}`,
			},
			expectMissing: []string{
				"Grain: nil",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := generateOutputWritingCode(tt.ioSpec)
			if tt.ioSpec == nil || len(tt.ioSpec.Outputs) == 0 {
				if code != "" {
					t.Errorf("generateOutputWritingCode() = %q, expected empty string", code)
				}
			}
			for _, expected := range tt.expectContains {
				if !strings.Contains(code, expected) {
					t.Errorf("generateOutputWritingCode() output missing expected content: %q", expected)
				}
			}
			for _, unexpected := range tt.expectMissing {
				if strings.Contains(code, unexpected) {
					t.Errorf("generateOutputWritingCode() output contains unexpected content: %q", unexpected)
				}
			}
		})
	}
}

// TestGenerateMainFunctionRejectsNilInput verifies that GenerateMainFunction
// returns an error when given a nil SynthesisInput.
func TestGenerateMainFunctionRejectsNilInput(t *testing.T) {
	_, err := GenerateMainFunction(nil)
	if err == nil {
		t.Error("GenerateMainFunction(nil) returned nil error, expected error")
	}
}

// TestGenerateMainFunctionRejectsNilIOSpec verifies that GenerateMainFunction
// returns an error when SynthesisInput has a nil IOSpec.
func TestGenerateMainFunctionRejectsNilIOSpec(t *testing.T) {
	input := &SynthesisInput{
		Transformer: cil.NewCILAbstractTransformer(),
	}
	_, err := GenerateMainFunction(input)
	if err == nil {
		t.Error("GenerateMainFunction(input with nil IOSpec) returned nil error, expected error")
	}
}

// TestGenerateMainFunctionRejectsNilTransformer verifies that GenerateMainFunction
// returns an error when SynthesisInput has a nil Transformer.
func TestGenerateMainFunctionRejectsNilTransformer(t *testing.T) {
	input := &SynthesisInput{
		IOSpec: &cio.IOSpecification{
			Connectors:       []cio.ConnectorSpecification{},
			Outputs:      []cio.OutputSpecification{},
			Architecture: cio.ArchitectureSpecification{},
		},
	}
	_, err := GenerateMainFunction(input)
	if err == nil {
		t.Error("GenerateMainFunction(input with nil Transformer) returned nil error, expected error")
	}
}

// TestBuildMainImports_ByteIdentical verifies that buildMainImports
// rendered via GeneratedProgram.RenderMainSource produces the expected
// import output for all 8 combinations of needsSync/needsInputs/needsOperator
// plus a tensorops case.
func TestBuildMainImports_ByteIdentical(t *testing.T) {
	tests := []struct {
		name             string
		needsSync        bool
		needsInputs      bool
		needsOperator    bool
		needsTensorops   bool
		golden           string
	}{
		{
			name: "no imports beyond base", needsSync: false, needsInputs: false, needsOperator: false, needsTensorops: false,
			golden: "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/calculation\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/connector\"\n\tcplio \"github.com/chronocone/chronocone_planning_language/src/main/go/io\"\n)",
		},
		{
			name: "sync only", needsSync: true, needsInputs: false, needsOperator: false, needsTensorops: false,
			golden: "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n\t\"sync\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/calculation\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/connector\"\n\tcplio \"github.com/chronocone/chronocone_planning_language/src/main/go/io\"\n)",
		},
		{
			name: "inputs only", needsSync: false, needsInputs: true, needsOperator: false, needsTensorops: false,
			golden: "package main\n\nimport (\n\t\"flag\"\n\t\"fmt\"\n\t\"os\"\n\t\"path/filepath\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/calculation\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/connector\"\n\tcplio \"github.com/chronocone/chronocone_planning_language/src/main/go/io\"\n)",
		},
		{
			name: "operator only", needsSync: false, needsInputs: false, needsOperator: true, needsTensorops: false,
			golden: "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/calculation\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/connector\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/operator\"\n\tcplio \"github.com/chronocone/chronocone_planning_language/src/main/go/io\"\n)",
		},
		{
			name: "sync+inputs", needsSync: true, needsInputs: true, needsOperator: false, needsTensorops: false,
			golden: "package main\n\nimport (\n\t\"flag\"\n\t\"fmt\"\n\t\"os\"\n\t\"path/filepath\"\n\t\"sync\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/calculation\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/connector\"\n\tcplio \"github.com/chronocone/chronocone_planning_language/src/main/go/io\"\n)",
		},
		{
			name: "sync+operator", needsSync: true, needsInputs: false, needsOperator: true, needsTensorops: false,
			golden: "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n\t\"sync\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/calculation\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/connector\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/operator\"\n\tcplio \"github.com/chronocone/chronocone_planning_language/src/main/go/io\"\n)",
		},
		{
			name: "inputs+operator", needsSync: false, needsInputs: true, needsOperator: true, needsTensorops: false,
			golden: "package main\n\nimport (\n\t\"flag\"\n\t\"fmt\"\n\t\"os\"\n\t\"path/filepath\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/calculation\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/connector\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/operator\"\n\tcplio \"github.com/chronocone/chronocone_planning_language/src/main/go/io\"\n)",
		},
		{
			name: "all imports", needsSync: true, needsInputs: true, needsOperator: true, needsTensorops: false,
			golden: "package main\n\nimport (\n\t\"flag\"\n\t\"fmt\"\n\t\"os\"\n\t\"path/filepath\"\n\t\"sync\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/calculation\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/connector\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/operator\"\n\tcplio \"github.com/chronocone/chronocone_planning_language/src/main/go/io\"\n)",
		},
		{
			name: "tensorops only", needsSync: false, needsInputs: false, needsOperator: false, needsTensorops: true,
			golden: "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/calculation\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/connector\"\n\t\"github.com/chronocone/chronocone_planning_language/src/main/go/tensorops\"\n\tcplio \"github.com/chronocone/chronocone_planning_language/src/main/go/io\"\n)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := buildMainImports(tt.needsSync, tt.needsInputs, tt.needsOperator, tt.needsTensorops)
			program := &GeneratedProgram{
				PackageName: "main",
				Imports:     imports,
			}
			// Trim trailing newlines: RenderMainSource appends "\n\n" after imports,
			// but the golden strings (derived from the pre-refactoring import
			// generation) omit trailing newlines.
			actual := strings.TrimRight(program.RenderMainSource(), "\n")

			if tt.golden != actual {
				t.Errorf("buildMainImports + RenderMainSource output differs from expected.\n--- Expected (%d bytes) ---\n%s\n--- Actual (%d bytes) ---\n%s",
					len(tt.golden), tt.golden, len(actual), actual)
			}
		})
	}
}

// TestRenderMainSource_WithAllFields verifies that RenderMainSource correctly
// renders a GeneratedProgram with all fields populated (imports, variables,
// functions, and raw source before main).
func TestRenderMainSource_WithAllFields(t *testing.T) {
	program := &GeneratedProgram{
		PackageName: "main",
		Imports: []ImportDecl{
			{Path: "fmt"},
			{Path: "os"},
			{Path: "github.com/chronocone/chronocone_planning_language/src/main/go/calculation"},
			{Path: "github.com/chronocone/chronocone_planning_language/src/main/go/connector"},
			{Alias: "cplio", Path: "github.com/chronocone/chronocone_planning_language/src/main/go/io"},
		},
		Variables: []VariableDecl{
			{
				Name:    "testVar",
				Type:    "string",
				Value:   `"hello"`,
				Comment: "testVar is a test variable.",
			},
		},
		Functions: []FunctionDecl{
			{
				Name:    "Setup",
				Params:  "cfg *Config",
				Returns: "error",
				Body:    "\tfmt.Println(\"setup\")\n\treturn nil\n",
				Comment: "Setup initializes the program.",
			},
		},
		RawSourceBeforeMain: []string{
			"func Compute() error {\n\treturn nil\n}\n\n",
		},
	}

	source := program.RenderMainSource()

	// Verify package declaration
	if !strings.Contains(source, "package main") {
		t.Errorf("RenderMainSource() missing 'package main'")
	}

	// Verify imports in declaration order (fmt, os, calculation, connector, cplio)
	if !strings.Contains(source, `"fmt"`) {
		t.Errorf("RenderMainSource() missing 'fmt' import")
	}
	if !strings.Contains(source, `"os"`) {
		t.Errorf("RenderMainSource() missing 'os' import")
	}

	// Verify structured variable with comment
	if !strings.Contains(source, "// testVar is a test variable.") {
		t.Errorf("RenderMainSource() missing variable comment")
	}
	if !strings.Contains(source, `var testVar string = "hello"`) {
		t.Errorf("RenderMainSource() missing variable declaration")
	}

	// Verify structured function with comment
	if !strings.Contains(source, "// Setup initializes the program.") {
		t.Errorf("RenderMainSource() missing function comment")
	}
	if !strings.Contains(source, "func Setup(cfg *Config) error {") {
		t.Errorf("RenderMainSource() missing Setup function signature")
	}

	// Verify raw source is rendered after structured functions
	if !strings.Contains(source, "func Compute() error {") {
		t.Errorf("RenderMainSource() missing Compute function from RawSourceBeforeMain")
	}

	// Verify ordering: Setup comes before Compute (structured functions before raw source)
	setupIdx := strings.Index(source, "func Setup(")
	computeIdx := strings.Index(source, "func Compute(")
	if setupIdx < 0 || computeIdx < 0 {
		t.Fatal("Cannot determine function ordering")
	}
	if setupIdx > computeIdx {
		t.Errorf("Setup function should appear before Compute function (structured before raw source)")
	}
}

// TestGenerateMainFunction_UsesGeneratedProgram verifies that the refactored
// GenerateMainFunction produces output byte-identical to its expected format
// when using the GeneratedProgram-based approach.
func TestGenerateMainFunction_UsesGeneratedProgram(t *testing.T) {
	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "testdata",
				DataFormat:    "parquet",
				Driver:        "file",
				DriverParameters: []map[string]interface{}{
					{"path": "testdata.parquet"},
				},
			},
		},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "test_metric",
				Dimensions: []string{"Region"},
				CellType:   "float64",
				Grain:      []string{"Region"},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	transformer := cil.NewCILAbstractTransformer()
	transformer.RegisterConnector("testdata", []string{"Region"}, nil, nil, "float64")

	input := &SynthesisInput{
		Transformer: transformer,
		IOSpec:      ioSpec,
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"test_metric": {
				Name: "test_metric",
				CILRoot: &cil.CILMetricNode{
					CILNode: &cil.CILLiteral{Value: 42},
				},
			},
		},
	}

	code, err := GenerateMainFunction(input)
	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	// Verify the output has the expected structure from GeneratedProgram-based generation
	expectedElements := []string{
		"package main",
		"import (",
		`"flag"`,
		`"fmt"`,
		`"os"`,
		`"path/filepath"`,
		`"sync"`,
		"calculation\"",
		"connector\"",
		"cplio \"github.com/chronocone/chronocone_planning_language/src/main/go/io\"",
		"var controller *calculation.ChronoconeComputationController",
		"var inputs map[string]*calculation.ChronoconeTensor[any]",
		"func Setup(architecture *cplio.ArchitectureSpecification) error {",
		"func main() {",
		"cplio.WriteMetricsToParquet",
		"func check(err error, msg string)",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(code, expected) {
			t.Errorf("GenerateMainFunction() output missing expected content: %q", expected)
		}
	}
}

// TestResolveGoType_BuiltInTypes verifies that resolveGoType correctly maps
// all CPL built-in types to their Go equivalents.
func TestResolveGoType_BuiltInTypes(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		want     string
	}{
		{"Number", "Number", "float64"},
		{"String", "String", "string"},
		{"Boolean", "Boolean", "bool"},
		{"Instant", "Instant", "time.Time"},
		{"Interval", "Interval", "time.Duration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveGoType(tt.typeName, nil)
			if err != nil {
				t.Fatalf("resolveGoType(%q) returned unexpected error: %v", tt.typeName, err)
			}
			if got != tt.want {
				t.Errorf("resolveGoType(%q) = %q, want %q", tt.typeName, got, tt.want)
			}
		})
	}
}

// TestResolveGoType_DerivedTypes verifies that resolveGoType resolves derived
// types by following their derivation path to the built-in root.
func TestResolveGoType_DerivedTypes(t *testing.T) {
	tests := []struct {
		name  string
		types map[string]analysis.AnalysisTypeDefinition
		want  string
	}{
		{
			name: "Currency derives from Number",
			types: map[string]analysis.AnalysisTypeDefinition{
				"Currency": {
					Name:           "Currency",
					DerivationPath: []string{"Currency", "Number"},
				},
			},
			want: "float64",
		},
		{
			name: "SmallCurrency derives from Currency derives from Number",
			types: map[string]analysis.AnalysisTypeDefinition{
				"SmallCurrency": {
					Name:           "SmallCurrency",
					DerivationPath: []string{"SmallCurrency", "Currency", "Number"},
				},
				"Currency": {
					Name:           "Currency",
					DerivationPath: []string{"Currency", "Number"},
				},
			},
			want: "float64",
		},
		{
			name: "CustomString derives from String",
			types: map[string]analysis.AnalysisTypeDefinition{
				"CustomString": {
					Name:           "CustomString",
					DerivationPath: []string{"CustomString", "String"},
				},
			},
			want: "string",
		},
		{
			name: "CustomBool derives from Boolean",
			types: map[string]analysis.AnalysisTypeDefinition{
				"CustomBool": {
					Name:           "CustomBool",
					DerivationPath: []string{"CustomBool", "Boolean"},
				},
			},
			want: "bool",
		},
		{
			name: "CustomInstant derives from Instant",
			types: map[string]analysis.AnalysisTypeDefinition{
				"CustomInstant": {
					Name:           "CustomInstant",
					DerivationPath: []string{"CustomInstant", "Instant"},
				},
			},
			want: "time.Time",
		},
		{
			name: "CustomInterval derives from Interval",
			types: map[string]analysis.AnalysisTypeDefinition{
				"CustomInterval": {
					Name:           "CustomInterval",
					DerivationPath: []string{"CustomInterval", "Interval"},
				},
			},
			want: "time.Duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Determine the type name from the types map (single entry tests)
			var typeName string
			for k := range tt.types {
				typeName = k
				break
			}
			got, err := resolveGoType(typeName, tt.types)
			if err != nil {
				t.Fatalf("resolveGoType(%q) returned unexpected error: %v", typeName, err)
			}
			if got != tt.want {
				t.Errorf("resolveGoType(%q) = %q, want %q", typeName, got, tt.want)
			}
		})
	}
}

// TestResolveGoType_Errors verifies that resolveGoType returns errors for
// unresolvable types instead of silently falling back to float64.
func TestResolveGoType_Errors(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		types    map[string]analysis.AnalysisTypeDefinition
		wantErr  string
	}{
		{
			name:     "unknown type not in map",
			typeName: "NonExistent",
			types:    map[string]analysis.AnalysisTypeDefinition{},
			wantErr:  "unknown type: NonExistent",
		},
		{
			name:     "empty derivation path",
			typeName: "BadType",
			types: map[string]analysis.AnalysisTypeDefinition{
				"BadType": {
					Name:           "BadType",
					DerivationPath: []string{},
				},
			},
			wantErr: "type BadType has empty derivation path",
		},
		{
			name:     "circular derivation path",
			typeName: "CircularA",
			types: map[string]analysis.AnalysisTypeDefinition{
				"CircularA": {
					Name:           "CircularA",
					DerivationPath: []string{"CircularA", "CircularB", "CircularA"},
				},
			},
			wantErr: "circular derivation path detected for type CircularA",
		},
		{
			name:     "derivation path ending in non-built-in",
			typeName: "Unresolvable",
			types: map[string]analysis.AnalysisTypeDefinition{
				"Unresolvable": {
					Name:           "Unresolvable",
					DerivationPath: []string{"Unresolvable", "AlsoUnknown"},
				},
			},
			wantErr: "does not resolve to a built-in type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resolveGoType(tt.typeName, tt.types)
			if err == nil {
				t.Fatalf("resolveGoType(%q) expected error, got nil", tt.typeName)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("resolveGoType(%q) error = %q, want to contain %q", tt.typeName, err.Error(), tt.wantErr)
			}
		})
	}
}

// TestGenerateMainFunction_LookupWiring verifies that GenerateMainFunction wires
// lookup loading into both the preamble (loadLookupData function) and the main
// function body (call to loadLookupData before Compute).
func TestGenerateMainFunction_LookupWiring(t *testing.T) {
	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Country": {ConceptName: "Country"},
			"Product": {ConceptName: "Product"},
		},
		Measures: map[string]analysis.AnalysisMeasureDefinition{
			"UnitsSold": {Name: "UnitsSold", TypeName: "Number"},
		},
		Lookups: map[string]*analysis.LookupInfo{
			"CountryProductIndex": {
				Name: "CountryProductIndex",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Country"},
					{ConceptName: "Product"},
				},
				CelltypeMeasure: &analysis.AnalysisMeasureDefinition{
					Name:     "UnitsSold",
					TypeName: "Number",
				},
			},
		},
		CILTrees: map[string]cil.CILNode{
			"lookup:CountryProductIndex": &cil.CILLookupNode{
				LookupName: "CountryProductIndex",
				Dimensions: []string{"Country", "Product"},
				Celltype:   "UnitsSold",
			},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors:  []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{},
		Lookups: []cio.LookupSpecification{
			{
				LookupName:       "CountryProductIndex",
				DataFormat:       "parquet",
				Driver:           "parquet",
				DriverParameters: []map[string]interface{}{{"path": "/data/lookups/country_product.parquet"}},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	input, err := BuildTransformerFromAnalysis(analysisDefs)
	if err != nil {
		t.Fatalf("BuildTransformerFromAnalysis() returned error: %v", err)
	}
	input.IOSpec = ioSpec

	code, err := GenerateMainFunction(input)
	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	// Verify preamble contains loadLookupData function
	if !strings.Contains(code, "func loadLookupData(controller *calculation.ChronoconeComputationController, dataDir *string) error") {
		t.Errorf("Generated code missing 'func loadLookupData'")
	}

	// Verify main calls loadLookupData(controller, dataDir)
	if !strings.Contains(code, "loadLookupData(controller, dataDir)") {
		t.Errorf("Generated code missing 'loadLookupData(controller, dataDir)' call in main")
	}

	// Verify loadLookupData is called before Compute() in main.
	// Use the specific call pattern to avoid matching the func Compute() declaration.
	lookupCallIdx := strings.Index(code, "loadLookupData(controller, dataDir)")
	computeCallIdx := strings.Index(code, "if err := Compute(); err != nil {")
	if lookupCallIdx < 0 || computeCallIdx < 0 {
		t.Fatalf("Cannot determine ordering: lookupCallIdx=%d, computeCallIdx=%d", lookupCallIdx, computeCallIdx)
	}
	if lookupCallIdx > computeCallIdx {
		t.Errorf("loadLookupData(controller, dataDir) should be called before Compute()")
	}

	// Verify lookup-specific code generation
	if !strings.Contains(code, "cplio.ReadParquetToTensor(filepath.Join(*dataDir, \"/data/lookups/country_product.parquet\"") {
		t.Errorf("Generated code missing ReadParquetToTensor call with lookup path")
	}
	if !strings.Contains(code, `controller.RegisterMetricComputation("CountryProductIndex", mc)`) {
		t.Errorf("Generated code missing lookup RegisterMetricComputation call")
	}
}

// TestCollectConnectorsForReachableMetrics verifies that the helper correctly
// identifies connectors referenced by reachable metrics' CIL trees.
func TestCollectConnectorsForReachableMetrics(t *testing.T) {
	tests := []struct {
		name     string
		input    *SynthesisInput
		dagOrder []string
		expected map[string]bool
	}{
		{
			name:     "nil input returns empty set",
			input:    nil,
			dagOrder: []string{"m1"},
			expected: map[string]bool{},
		},
		{
			name:     "empty dagOrder returns empty set",
			input:    &SynthesisInput{},
			dagOrder: []string{},
			expected: map[string]bool{},
		},
		{
			name: "metric with CILConnectorNode",
			input: &SynthesisInput{
				Metrics: map[string]analysis.AnalysisMetricDefinition{
					"revenue": {
						Name: "revenue",
						CILRoot: &cil.CILMetricNode{
							CILNode: &cil.CILConnectorNode{ConnectorName: "sales_data"},
						},
					},
				},
			},
			dagOrder: []string{"revenue"},
			expected: map[string]bool{"sales_data": true},
		},
		{
			name: "metric with CILInputRef connector",
			input: &SynthesisInput{
				Metrics: map[string]analysis.AnalysisMetricDefinition{
					"profit": {
						Name: "profit",
						CILRoot: &cil.CILMetricNode{
							CILNode:   &cil.CILInputRef{Kind: "connector", Name: "cost_data"},
							InputRefs: []cil.CILInputRef{{Kind: "connector", Name: "cost_data"}},
						},
					},
				},
			},
			dagOrder: []string{"profit"},
			expected: map[string]bool{"cost_data": true},
		},
		{
			name: "multiple metrics, only dagOrder metrics considered",
			input: &SynthesisInput{
				Metrics: map[string]analysis.AnalysisMetricDefinition{
					"reachable": {
						Name: "reachable",
						CILRoot: &cil.CILMetricNode{
							CILNode: &cil.CILConnectorNode{ConnectorName: "needed_connector"},
						},
					},
					"unreachable": {
						Name: "unreachable",
						CILRoot: &cil.CILMetricNode{
							CILNode: &cil.CILConnectorNode{ConnectorName: "unused_connector"},
						},
					},
				},
			},
			dagOrder: []string{"reachable"},
			expected: map[string]bool{"needed_connector": true},
		},
		{
			name: "metric without CILRoot is skipped",
			input: &SynthesisInput{
				Metrics: map[string]analysis.AnalysisMetricDefinition{
					"no_cil": {Name: "no_cil"},
				},
			},
			dagOrder: []string{"no_cil"},
			expected: map[string]bool{},
		},
		{
			name: "deeply nested connector reference",
			input: &SynthesisInput{
				Metrics: map[string]analysis.AnalysisMetricDefinition{
					"deep": {
						Name: "deep",
						CILRoot: &cil.CILMetricNode{
							CILNode: &cil.CILBinaryOp{
								Left:  &cil.CILConnectorNode{ConnectorName: "a"},
								Right: &cil.CILBinaryOp{
									Left:  &cil.CILLiteral{Value: 1},
									Right: &cil.CILConnectorNode{ConnectorName: "b"},
								},
							},
						},
					},
				},
			},
			dagOrder: []string{"deep"},
			expected: map[string]bool{"a": true, "b": true},
		},
		{
			name: "aggregator node with connector source",
			input: &SynthesisInput{
				Metrics: map[string]analysis.AnalysisMetricDefinition{
					"agg": {
						Name: "agg",
						CILRoot: &cil.CILMetricNode{
							CILNode: &cil.CILAggregateNode{
								SourceCIL:      &cil.CILConnectorNode{ConnectorName: "sales_data"},
								FromDimensions: []string{"Country", "Product"},
								ToDimension:    "Product",
							},
						},
					},
				},
			},
			dagOrder: []string{"agg"},
			expected: map[string]bool{"sales_data": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectConnectorsForReachableMetrics(tt.input, tt.dagOrder)
			if len(result) != len(tt.expected) {
				t.Errorf("collectConnectorsForReachableMetrics() returned %d entries, expected %d: got %v", len(result), len(tt.expected), result)
			}
			for name := range tt.expected {
				if !result[name] {
					t.Errorf("collectConnectorsForReachableMetrics() missing expected connector %q", name)
				}
			}
			for name := range result {
				if !tt.expected[name] {
					t.Errorf("collectConnectorsForReachableMetrics() contains unexpected connector %q", name)
				}
			}
		})
	}
}

// TestGenerateMainFunction_NoLookupsOmitsLoadLookupData verifies that when no
// lookups are registered, the generated code does NOT contain loadLookupData
// function or call.
func TestGenerateMainFunction_NoLookupsOmitsLoadLookupData(t *testing.T) {
	ioSpec := &cio.IOSpecification{
		Connectors:  []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	transformer := cil.NewCILAbstractTransformer()
	input := &SynthesisInput{Transformer: transformer, IOSpec: ioSpec}

	code, err := GenerateMainFunction(input)
	if err != nil {
		t.Fatalf("GenerateMainFunction() returned error: %v", err)
	}

	if strings.Contains(code, "func loadLookupData") {
		t.Errorf("Generated code should NOT contain 'func loadLookupData' when no lookups are registered")
	}
	if strings.Contains(code, "loadLookupData(controller, dataDir)") {
		t.Errorf("Generated code should NOT contain 'loadLookupData(controller, dataDir)' when no lookups are registered")
	}
}

// TestGenerateMainFunction_MissingLookupsWithRegisteredLookups verifies that
// GenerateMainFunction returns an error when lookups are registered but
// no lookup entries are present in the IO spec.
func TestGenerateMainFunction_MissingLookupsWithRegisteredLookups(t *testing.T) {
	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Country": {ConceptName: "Country"},
			"Product": {ConceptName: "Product"},
		},
		Measures: map[string]analysis.AnalysisMeasureDefinition{
			"UnitsSold": {Name: "UnitsSold", TypeName: "Number"},
		},
		Lookups: map[string]*analysis.LookupInfo{
			"CountryProductIndex": {
				Name: "CountryProductIndex",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Country"},
					{ConceptName: "Product"},
				},
				CelltypeMeasure: &analysis.AnalysisMeasureDefinition{
					Name:     "UnitsSold",
					TypeName: "Number",
				},
			},
		},
		CILTrees: map[string]cil.CILNode{
			"lookup:CountryProductIndex": &cil.CILLookupNode{
				LookupName: "CountryProductIndex",
				Dimensions: []string{"Country", "Product"},
				Celltype:   "UnitsSold",
			},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors:  []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	input, err := BuildTransformerFromAnalysis(analysisDefs)
	if err != nil {
		t.Fatalf("BuildTransformerFromAnalysis() returned error: %v", err)
	}
	input.IOSpec = ioSpec

	code, err := GenerateMainFunction(input)
	if err == nil {
		t.Fatal("GenerateMainFunction() expected error for missing lookups with registered lookups, got nil")
	}
	if code != "" {
		t.Errorf("GenerateMainFunction() should return empty code on error, got: %s", code)
	}
	if !strings.Contains(err.Error(), "lookup IO spec") {
		t.Errorf("Error should mention missing lookup IO spec, got: %v", err)
	}
}

// TestGenerateMainFunction_LookupCompile verifies that code generated with
// lookups compiles successfully with `go build`.
func TestGenerateMainFunction_LookupCompile(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("'go' command not available in PATH")
	}

	moduleRoot, err := findModuleRoot()
	if err != nil {
		t.Fatalf("cannot find module root: %v", err)
	}

	generatedCmdDir := filepath.Join(moduleRoot, "generated", "cmd")
	if err := os.MkdirAll(generatedCmdDir, 0755); err != nil {
		t.Fatalf("Failed to create generated/cmd directory: %v", err)
	}
	testDir, err := os.MkdirTemp(generatedCmdDir, "lookup_compile_*")
	if err != nil {
		t.Fatalf("Failed to create temp test dir under module root: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(testDir) })

	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Country": {
				ConceptName: "Country",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "USA", Position: 0},
					{Name: "Canada", Position: 1},
				},
			},
			"Product": {
				ConceptName: "Product",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "Widget", Position: 0},
					{Name: "Gadget", Position: 1},
				},
			},
		},
		Measures: map[string]analysis.AnalysisMeasureDefinition{
			"UnitsSold": {Name: "UnitsSold", TypeName: "Number"},
		},
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"sales_data": {
				Name: "sales_data",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Country"},
				},
				CellType: &analysis.AnalysisMeasureDefinition{
					Name:     "UnitsSold",
					TypeName: "Number",
				},
			},
		},
		Lookups: map[string]*analysis.LookupInfo{
			"CountryProductIndex": {
				Name: "CountryProductIndex",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Country"},
					{ConceptName: "Product"},
				},
				CelltypeMeasure: &analysis.AnalysisMeasureDefinition{
					Name:     "UnitsSold",
					TypeName: "Number",
				},
			},
		},
		CILTrees: map[string]cil.CILNode{
			"lookup:CountryProductIndex": &cil.CILLookupNode{
				LookupName: "CountryProductIndex",
				Dimensions: []string{"Country", "Product"},
				Celltype:   "UnitsSold",
			},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors:  []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{},
		Lookups: []cio.LookupSpecification{
			{
				LookupName:       "CountryProductIndex",
				DataFormat:       "parquet",
				Driver:           "parquet",
				DriverParameters: []map[string]interface{}{{"path": "/data/lookups/country_product.parquet"}},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	outputPath := filepath.Join(testDir, "main.go")
	if err := Synthesize(analysisDefs, ioSpec, outputPath); err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	// Verify compilation succeeds
	binaryPath := filepath.Join(testDir, "testbin")
	cmd := exec.Command("go", "build", "-o", binaryPath, outputPath)
	cmd.Dir = testDir
	buildOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Generated program failed to compile: %v\nOutput:\n%s", err, string(buildOutput))
	}

	// Verify binary was created
	if _, err := os.Stat(binaryPath); err != nil {
		t.Errorf("Compiled binary not found at %s: %v", binaryPath, err)
	}
}

// TestSynthesize_LookupWiringIntegration verifies that Synthesize passes
// ioSpec.Lookups through to the generated code.
func TestSynthesize_LookupWiringIntegration(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.go")

	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Region": {ConceptName: "Region"},
		},
		Measures: map[string]analysis.AnalysisMeasureDefinition{
			"IndexValue": {Name: "IndexValue", TypeName: "Number"},
		},
		Lookups: map[string]*analysis.LookupInfo{
			"RegionMapping": {
				Name: "RegionMapping",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Region"},
				},
				CelltypeMeasure: &analysis.AnalysisMeasureDefinition{
					Name:     "IndexValue",
					TypeName: "Number",
				},
			},
		},
		CILTrees: map[string]cil.CILNode{
			"lookup:RegionMapping": &cil.CILLookupNode{
				LookupName: "RegionMapping",
				Dimensions: []string{"Region"},
				Celltype:   "IndexValue",
			},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors:  []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{},
		Lookups: []cio.LookupSpecification{
			{
				LookupName:       "RegionMapping",
				DataFormat:       "parquet",
				Driver:           "parquet",
				DriverParameters: []map[string]interface{}{{"path": "/data/lookups/regions.parquet"}},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}


	err := Synthesize(analysisDefs, ioSpec, outputPath)
	if err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "func loadLookupData") {
		t.Errorf("Generated file missing 'func loadLookupData'")
	}
	if !strings.Contains(contentStr, "loadLookupData(controller, dataDir)") {
		t.Errorf("Generated file missing 'loadLookupData(controller, dataDir)' call")
	}
	if !strings.Contains(contentStr, "/data/lookups/regions.parquet") {
		t.Errorf("Generated file missing lookup parquet path")
	}
}

// TestSynthesize_NoLookupsOmitsLookupData verifies that passing an ioSpec with no lookups produces
// valid code with no loadLookupData function when no lookups are present.
func TestSynthesize_NoLookupsOmitsLookupData(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.go")

	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Region": {ConceptName: "Region"},
		},
	}

	ioSpec := &cio.IOSpecification{
		Connectors:  []cio.ConnectorSpecification{},
		Outputs: []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	err := Synthesize(analysisDefs, ioSpec, outputPath)
	if err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if strings.Contains(contentStr, "func loadLookupData") {
		t.Errorf("Generated file should NOT contain 'func loadLookupData' when no lookups are present")
	}
	if !strings.Contains(contentStr, "func main()") {
		t.Errorf("Generated file missing 'func main()'")
	}
}

// TestEndToEndSynthesizeLookups_Basic verifies the full end-to-end pipeline
// for a basic lookup scenario matching the lookups_basic example:
// parse → analyze → synthesize → compile → execute → verify.
//
// The test models: a connector dummy (Category × Daily), a lookup
// product_category (Product → values), and a metric result = dummy.
// The generated program loads the connector, loads the lookup (registered
// as a MetricComputation on the controller), computes the metric (pass-through
// from connector), and writes output.parquet.
//
// In Increment 1, lookups are loaded but not consumed by metric calculations,
// so the output reflects the connector data.
func TestEndToEndSynthesizeLookups_Basic(t *testing.T) {
	// Check that `go` is available in PATH
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("'go' command not available in PATH: %v", err)
	}

	moduleRoot, err := findModuleRoot()
	if err != nil {
		t.Fatalf("cannot find module root: %v", err)
	}

	testDir, err := os.MkdirTemp(filepath.Join(moduleRoot, "generated", "cmd"), "e2e_lookup_*")
	if err != nil {
		t.Fatalf("Failed to create temp test dir under module root: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(testDir) })

	// Create data directory for test parquet input files
	dataDir := filepath.Join(testDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("Failed to create data dir: %v", err)
	}

	// --- Generate dummy connector parquet ---
	// Categories: Metal, Plastic, Wood (3 values) × Days: Sunday..Saturday (7 values) = 21 rows
	categories := []string{"Metal", "Plastic", "Wood"}
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	dummyParquetPath := filepath.Join(dataDir, "dummy.parquet")
	dummyDimensions := map[string][]string{
		"Category": categories,
		"Daily":    days,
	}
	dummyValues := make([]float64, 21)
	for i := range dummyValues {
		dummyValues[i] = float64((i / 7) + 1) // Category index + 1
	}
	if err := cio.GenerateParquetFile(dummyParquetPath, dummyDimensions, nil, dummyValues); err != nil {
		t.Fatalf("Failed to generate dummy.parquet: %v", err)
	}

	// --- Generate lookup parquet ---
	// Product values: Steel, Copper, Aluminum (3 rows) with 1.0 values
	lookupParquetPath := filepath.Join(dataDir, "lookups_basic_product_category.parquet")
	lookupRows := []map[string]string{
		{"Product": "Steel"},
		{"Product": "Copper"},
		{"Product": "Aluminum"},
	}
	lookupValues := []float64{1.0, 1.0, 1.0}
	if err := cio.WriteParquetFromRows(lookupParquetPath, lookupRows, lookupValues); err != nil {
		t.Fatalf("Failed to generate lookup parquet: %v", err)
	}

	// --- Build AnalysisDefinitions matching the CPL structure ---
	analysisDefs := &analysis.AnalysisDefinitions{
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Product": {
				ConceptName: "Product",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "Steel", Position: 0},
					{Name: "Copper", Position: 1},
					{Name: "Aluminum", Position: 2},
				},
			},
			"Category": {
				ConceptName: "Category",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "Metal", Position: 0},
					{Name: "Plastic", Position: 1},
					{Name: "Wood", Position: 2},
				},
			},
			"Day": {
				ConceptName: "Day",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "Sunday", Position: 0},
					{Name: "Monday", Position: 1},
					{Name: "Tuesday", Position: 2},
					{Name: "Wednesday", Position: 3},
					{Name: "Thursday", Position: 4},
					{Name: "Friday", Position: 5},
					{Name: "Saturday", Position: 6},
				},
			},
		},
		Measures: map[string]analysis.AnalysisMeasureDefinition{
			"CategoryLabel": {
				Name:     "CategoryLabel",
				TypeName: "Number",
				UnitExpression: &analysis.UnitReference{Name: "Category"},
			},
			"Daily": {
				Name:     "Daily",
				TypeName: "Interval",
				UnitExpression: &analysis.UnitReference{Name: "Day"},
			},
		},
		Connectors: map[string]analysis.AnalysisConnectorDefinition{
			"dummy": {
				Name: "dummy",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Category"},
				},
				CellType: &analysis.AnalysisMeasureDefinition{
					Name:     "CategoryLabel",
					TypeName: "Number",
				},
				Grain: []*analysis.AnalysisMeasureDefinition{
					{
						Name:     "Daily",
						TypeName: "Interval",
						UnitExpression: &analysis.UnitReference{Name: "Day"},
					},
				},
			},
		},
		Lookups: map[string]*analysis.LookupInfo{
			"product_category": {
				Name: "product_category",
				Dimensions: []*analysis.AnalysisConceptDefinition{
					{ConceptName: "Product"},
				},
				CelltypeMeasure: &analysis.AnalysisMeasureDefinition{
					Name:     "CategoryLabel",
					TypeName: "Number",
				},
			},
		},
		CILTrees: map[string]cil.CILNode{
			"lookup:product_category": &cil.CILLookupNode{
				LookupName: "product_category",
				Dimensions: []string{"Product"},
				Celltype:   "CategoryLabel",
			},
		},
		Metrics: map[string]analysis.AnalysisMetricDefinition{
			"result": {
				Name: "result",
				Inputs: []analysis.AnalysisInputDefinition{
					{Kind: "connector", Name: "dummy"},
					{Kind: "dimension", Name: "Category"},
				},
				CILRoot: &cil.CILMetricNode{
					CILNode:   &cil.CILInputRef{Kind: "connector", Name: "dummy"},
					InputRefs: []cil.CILInputRef{{Kind: "connector", Name: "dummy"}},
				},
			},
		},
	}

	// --- Build IO specs ---
	ioSpec := &cio.IOSpecification{
		Connectors: []cio.ConnectorSpecification{
			{
				ConnectorName: "dummy",
				DataFormat:    "parquet",
				Driver:        "file",
				DriverParameters: []map[string]interface{}{
					{"path": dummyParquetPath},
				},
			},
		},
		Outputs: []cio.OutputSpecification{
			{
				MetricName: "result",
				Dimensions: []string{"Category"},
				CellType:   "float64",
				Grain:      []string{"Daily"},
			},
		},
		Architecture: cio.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  8388608,
			MainMemoryBytes: 17179869184,
		},
		Lookups: []cio.LookupSpecification{
			{
				LookupName:       "product_category",
				DataFormat:       "parquet",
				Driver:           "parquet",
				DriverParameters: []map[string]interface{}{{"path": lookupParquetPath}},
			},
		},
	}

	// --- Synthesize Go code ---
	outputPath := filepath.Join(testDir, "main.go")
	if err := Synthesize(analysisDefs, ioSpec, outputPath); err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	// Verify the generated code contains loadLookupData wiring
	generatedCode, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated code: %v", err)
	}
	codeStr := string(generatedCode)
	if !strings.Contains(codeStr, "func loadLookupData") {
		t.Errorf("Generated code missing 'func loadLookupData'")
	}
	if !strings.Contains(codeStr, "loadLookupData(controller, dataDir)") {
		t.Errorf("Generated code missing 'loadLookupData(controller, dataDir)' call in main")
	}

	// --- Run the generated program ---
	cmd := exec.Command("go", "run", outputPath)
	cmd.Dir = testDir
	runOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Generated program failed to run: %v\nOutput:\n%s", err, string(runOutput))
	}

	// --- Verify output parquet ---
	resultParquetPath := filepath.Join(testDir, "result.parquet")
	resultParquet, numRows, err := cio.ReadParquetFile(resultParquetPath)
	if err != nil {
		t.Fatalf("Failed to read output parquet: %v", err)
	}

	// Expected: 3 Categories × 7 Days × 1 Daily value per (Category, Day) = 21 rows
	expectedRows := 21
	if numRows != expectedRows {
		t.Errorf("Expected %d rows in output parquet, got %d", expectedRows, numRows)
	}

	// Verify column names: Category, Daily, Value
	colNames := cio.GetColumnNames(resultParquet)
	hasCategory := false
	hasDaily := false
	hasValue := false
	for _, name := range colNames {
		switch name {
		case "Category":
			hasCategory = true
		case "Daily":
			hasDaily = true
		case "Value":
			hasValue = true
		}
	}
	if !hasCategory {
		t.Errorf("Output parquet missing 'Category' column. Columns: %v", colNames)
	}
	if !hasDaily {
		t.Errorf("Output parquet missing 'Daily' column. Columns: %v", colNames)
	}
	if !hasValue {
		t.Errorf("Output parquet missing 'Value' column. Columns: %v", colNames)
	}

	// Verify the CommandCompletedSuccessfully message
	if !strings.Contains(string(runOutput), "Computation completed successfully") {
		t.Errorf("Generated program should output 'Computation completed successfully'. Got: %s", string(runOutput))
	}
}

// TestGenerateMainFunction_WithAggregator verifies that GenerateMainFunction
// correctly emits aggregation matrix declarations and the tensorops import
// when aggregator CIL trees are present, and that the generated program
// compiles successfully.
func TestGenerateMainFunction_WithAggregator(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("'go' command not available in PATH")
	}

	moduleRoot, err := findModuleRoot()
	if err != nil {
		t.Fatalf("cannot find module root: %v", err)
	}

	transformer := cil.NewCILAbstractTransformer()

	ioSpec := &cio.IOSpecification{
		Connectors:   []cio.ConnectorSpecification{},
		Outputs:      []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{},
	}

	input := &SynthesisInput{
		Transformer: transformer,
		IOSpec:      ioSpec,
		Concepts: map[string]analysis.AnalysisConceptDefinition{
			"Country": {
				ConceptName: "Country",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "USA", Position: 0},
					{Name: "Canada", Position: 1},
				},
			},
			"Region": {
				ConceptName: "Region",
				ResolvedValues: []analysis.AnalysisConceptValueDefinition{
					{Name: "North", Position: 0},
					{Name: "South", Position: 1},
				},
			},
		},
		CILTrees: map[string]cil.CILNode{
			"aggregator:RegionToCountry": &cil.CILAggregateNode{
				FromDimensions:             []string{"Country", "Region"},
				ToDimension:       "Country",
				AggregationMatrix: &cil.AggregationMatrix{
					Rows:      2,
					Columns:   2,
					Data:      [][]float64{{1, 0}, {0, 1}},
					RowLabels: []string{"North", "South"},
					ColLabels: []string{"North", "South"},
				},
			},
		},
	}

	code, err := GenerateMainFunction(input)
	if err != nil {
		t.Fatalf("GenerateMainFunction failed: %v", err)
	}

	// Verify tensorops import is present.
	if !strings.Contains(code, `"github.com/chronocone/chronocone_planning_language/src/main/go/tensorops"`) {
		t.Errorf("Generated code missing tensorops import")
	}

	// Verify aggregation matrix variable declaration is present.
	if !strings.Contains(code, "var aggregationMatrix_Region") {
		t.Errorf("Generated code missing aggregation matrix variable declaration")
	}

	// Verify the matrix appears before compute/cleanup functions by checking
	// that the matrix declaration precedes the Compute function declaration.
	matrixIdx := strings.Index(code, "var aggregationMatrix_Region")
	computeIdx := strings.Index(code, "func Compute()")
	if matrixIdx == -1 || computeIdx == -1 || matrixIdx > computeIdx {
		t.Errorf("Aggregation matrix declaration should appear before Compute function")
	}

	// Write generated code to a temp file inside the module for compilation.
	testDir, err := os.MkdirTemp(filepath.Join(moduleRoot, "generated", "cmd"), "aggregator_compile_*")
	if err != nil {
		t.Fatalf("Failed to create temp test dir under module root: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(testDir) })

	outputPath := filepath.Join(testDir, "main.go")
	if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write generated code: %v", err)
	}

	cmd := exec.Command("go", "build", outputPath)
	cmd.Dir = moduleRoot
	buildOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Generated program failed to compile: %v\nOutput:\n%s", err, string(buildOutput))
	}
}

// TestGenerateMainFunction_WithoutAggregator verifies that no tensorops
// import or aggregation matrix declarations are emitted when no aggregators
// are present.
func TestGenerateMainFunction_WithoutAggregator(t *testing.T) {
	transformer := cil.NewCILAbstractTransformer()

	ioSpec := &cio.IOSpecification{
		Connectors:   []cio.ConnectorSpecification{},
		Outputs:      []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{},
	}

	input := &SynthesisInput{
		Transformer: transformer,
		IOSpec:      ioSpec,
		CILTrees:    map[string]cil.CILNode{},
	}

	code, err := GenerateMainFunction(input)
	if err != nil {
		t.Fatalf("GenerateMainFunction failed: %v", err)
	}

	if strings.Contains(code, `"github.com/chronocone/chronocone_planning_language/src/main/go/tensorops"`) {
		t.Errorf("Generated code should not contain tensorops import when no aggregators are present")
	}

	if strings.Contains(code, "aggregationMatrix_") {
		t.Errorf("Generated code should not contain aggregation matrix declarations when no aggregators are present")
	}
}
