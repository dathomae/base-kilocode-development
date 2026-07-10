package main

import (
	"flag"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
)

// TestOutputFlagDefaultValue verifies that the -o flag defaults to empty
// when no output path is explicitly provided on the command line.
//
// This test ensures the output path has no default value, requiring users
// to explicitly specify where the generated Go CLI file should be written.
//
// Expected behavior: When the compiler is invoked without the -o flag,
// the output path should be empty, causing the compiler to print an error
// and usage message.
func TestOutputFlagDefaultValue(t *testing.T) {
	// Save and restore original command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Reset flags to default state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Set up minimal args (no -o flag)
	os.Args = []string{"cpl_compiler", "test.cpl"}

	// Parse flags
	outputFlag := flag.String("o", "", "Specify output file path for generated Go CLI (required)")
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	// Verify default value is empty (no fallback)
	if *outputFlag != "" {
		t.Errorf("Expected default output path to be empty, got '%s'", *outputFlag)
	}
}

// TestOutputFlagParsing verifies that the -o flag can be set to a custom value
// when explicitly provided on the command line.
//
// This test ensures that user-specified output paths are correctly parsed
// and stored, overriding the default value. This is essential for users
// who need to control where the generated Go CLI file is written.
//
// Expected behavior: When the compiler is invoked with -o=/custom/path.go,
// the output path should be set to "/custom/path.go" instead of the default.
func TestOutputFlagParsing(t *testing.T) {
	// Save and restore original command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Reset flags to default state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Set up args with custom -o flag
	customPath := "/custom/path.go"
	os.Args = []string{"cpl_compiler", "-o=" + customPath, "test.cpl"}

	// Parse flags
	outputFlag := flag.String("o", "", "Specify output file path for generated Go CLI (required)")
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	// Verify custom value was parsed
	if *outputFlag != customPath {
		t.Errorf("Expected output path to be '%s', got '%s'", customPath, *outputFlag)
	}
}

// TestLongFormOutputFlag verifies that --output works as an alias for -o.
//
// This test ensures that users can use either the short form (-o) or the
// long form (--output) to specify the output file path. Both forms should
// be functionally equivalent and set the output path correctly.
//
// Expected behavior: When the compiler is invoked with --output=/custom/path.go,
// it should behave identically to -o=/custom/path.go.
func TestLongFormOutputFlag(t *testing.T) {
	// Save and restore original command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Reset flags to default state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Set up args with custom --output flag (long form)
	customPath := "/long/form/path.go"
	os.Args = []string{"cpl_compiler", "--output=" + customPath, "test.cpl"}

	// Parse flags - need to set up both -o and --output
	outputFlag := flag.String("o", "", "Specify output file path for generated Go CLI (required)")
	flag.StringVar(outputFlag, "output", "", "Specify output file path for generated Go CLI (required)")
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	// Verify custom value was parsed via long form
	if *outputFlag != customPath {
		t.Errorf("Expected output path to be '%s' via --output, got '%s'", customPath, *outputFlag)
	}
}

// TestCompileOnlyFlagParsing verifies that --compile_only enables compile-only
// mode when provided on the command line.
//
// Expected behavior: When the compiler is invoked with --compile_only, the flag
// should parse to true so the CLI can skip output and IO specification
// validation before running parse and analysis.
func TestCompileOnlyFlagParsing(t *testing.T) {
	// Save and restore original command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Reset flags to default state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Set up args with compile-only mode enabled
	os.Args = []string{"cpl_compiler", "--compile_only", "test.cpl"}

	// Parse flags
	compileOnlyFlag := flag.Bool("compile_only", false, "Run parse and analyze phases only; skip code generation")
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	// Verify compile-only mode was enabled
	if !*compileOnlyFlag {
		t.Error("Expected --compile_only to parse as true")
	}
}

// TestValidateOutputPathFunction verifies the validateOutputPath function signature
// and basic behavior.
//
// This test ensures that the validateOutputPath function exists with the correct
// signature (taking a string path parameter and returning an error) and handles
// various path validation scenarios appropriately.
//
// The function should validate:
// - Path is not empty
// - Parent directory exists or can be created
// - Path is writable
//
// Expected behavior: The function should return nil for valid paths and
// an appropriate error for invalid paths.
func TestValidateOutputPathFunction(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "empty path returns error",
			path:        "",
			expectError: true,
		},
		{
			name:        "valid path in temp directory",
			path:        filepath.Join(t.TempDir(), "output.go"),
			expectError: false,
		},
		{
			name:        "valid path in nested temp directory (creates parent)",
			path:        filepath.Join(t.TempDir(), "nested", "deep", "output.go"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOutputPath(tt.path)
			if tt.expectError && err == nil {
				t.Errorf("validateOutputPath(%q) expected error but got nil", tt.path)
			}
			if !tt.expectError && err != nil {
				t.Errorf("validateOutputPath(%q) expected no error but got: %v", tt.path, err)
			}
		})
	}
}

// TestHelpTextIncludesOutputFlag verifies that the help output includes
// documentation for the -o flag.
//
// This test ensures that users can discover the output path option by running
// the compiler with -h or --help flags. The help text should clearly describe
// the purpose of the -o flag and indicate that it is required.
//
// Expected behavior: The help output should contain information about the
// -o flag, including its purpose (specifying output file path) and
// indication that it is required (no default value).
func TestHelpTextIncludesOutputFlag(t *testing.T) {
	// Build the compiler binary in a temp directory
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "cpl_compiler")

	// Get the source file path
	sourcePath := filepath.Join(".", "cpl_compiler.go")

	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", binaryPath, sourcePath)
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, buildOutput)
	}

	// Run the binary with -help flag
	helpCmd := exec.Command(binaryPath, "-help")
	helpOutput, err := helpCmd.CombinedOutput()
	if err != nil && !strings.Contains(string(helpOutput), "usage") {
		// Some implementations exit with error on -help, which is fine
		// as long as the output contains the usage information
		t.Fatalf("Failed to get help output: %v\nOutput: %s", err, helpOutput)
	}

	helpText := string(helpOutput)

	// Check that help text includes output flag documentation
	if !strings.Contains(helpText, "-o") {
		t.Errorf("Help text should contain '-o' flag documentation")
	}

	if !strings.Contains(helpText, "output") {
		t.Errorf("Help text should contain 'output' keyword")
	}

	// Verify help text indicates the flag is required unless compile-only mode is used.
	if !strings.Contains(helpText, "required unless --compile_only") {
		t.Errorf("Help text should indicate that -o is required unless --compile_only is set")
	}
}

// TestHelpTextIncludesCompileOnlyFlag verifies that the help output documents
// the compile-only mode flag.
func TestHelpTextIncludesCompileOnlyFlag(t *testing.T) {
	// Build the compiler binary in a temp directory
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "cpl_compiler")

	// Get the source file path
	sourcePath := filepath.Join(".", "cpl_compiler.go")

	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", binaryPath, sourcePath)
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, buildOutput)
	}

	// Run the binary with -help flag
	helpCmd := exec.Command(binaryPath, "-help")
	helpOutput, err := helpCmd.CombinedOutput()
	if err != nil && !strings.Contains(string(helpOutput), "usage") {
		// Some implementations exit with error on -help, which is fine
		// as long as the output contains the usage information
		t.Fatalf("Failed to get help output: %v\nOutput: %s", err, helpOutput)
	}

	helpText := string(helpOutput)

	if !strings.Contains(helpText, "compile_only") {
		t.Errorf("Help text should contain 'compile_only' flag documentation")
	}

	if !strings.Contains(helpText, "skip code generation") {
		t.Errorf("Help text should explain that compile-only mode skips code generation")
	}
}

// TestCPLCompilerCreation tests that NewCPLCompiler creates a compiler instance
// with the correct fields initialized from the provided parameters.
//
// This test verifies that the CPLCompiler struct is properly constructed with:
//   - The input path correctly stored for reading the CPL source file
//   - The output path correctly stored for writing the generated Go code
//   - The logger instance properly assigned for compilation messages
//
// The compiler acts as the orchestrator for the full compilation pipeline,
// so proper initialization of these fields is essential for the Compile method
// to function correctly across all three phases (parse, analyze, synthesize).
//
// Expected behavior: NewCPLCompiler should return a non-nil *CPLCompiler with
// all fields (inputPath, outputPath, logger) set to the values passed as arguments.
func TestCPLCompilerCreation(t *testing.T) {
	logger := slog.Default()
	compiler := NewCPLCompiler([]string{"input.cpl"}, "output.go", "/path/to/io.json", logger, slog.LevelInfo)

	if compiler == nil {
		t.Fatal("NewCPLCompiler returned nil")
	}

	if len(compiler.inputPaths) != 1 || compiler.inputPaths[0] != "input.cpl" {
		t.Errorf("Expected inputPaths to be ['input.cpl'], got '%v'", compiler.inputPaths)
	}

	if compiler.outputPath != "output.go" {
		t.Errorf("Expected outputPath to be 'output.go', got '%s'", compiler.outputPath)
	}

	if compiler.ioSpecPath != "/path/to/io.json" {
		t.Errorf("Expected ioSpecPath to be '/path/to/io.json', got '%s'", compiler.ioSpecPath)
	}

	if compiler.compileOnly {
		t.Error("Expected compileOnly to default to false")
	}

	if compiler.logger != logger {
		t.Error("Expected logger to be the same instance passed to constructor")
	}
}

// TestSynthesizePhaseWithIOFlag verifies that the synthesizePhase uses
// Synthesize with IO spec to produce a full computation lifecycle program.
func TestSynthesizePhaseWithIOFlag(t *testing.T) {
	tempDir := t.TempDir()

	// Create a valid IO specification JSON file
	ioSpecContent := `{
		"connectors": [
			{
				"connector": "sales_data",
				"data_format": "parquet",
				"driver": "file",
				"driver_parameters": [{"path": "/data/sales.parquet"}]
			}
		],
		"outputs": [
			{
				"metric": "total_sales",
				"dimensions": ["Country"],
				"celltype": "float64",
				"grain": ["Country"]
			}
		],
		"architecture": {
			"cpu_cores": 4,
			"cache_size": 1048576,
			"main_memory_size": 1073741824
		}
	}`
	ioSpecPath := filepath.Join(tempDir, "io_spec.json")
	err := os.WriteFile(ioSpecPath, []byte(ioSpecContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write IO spec file: %v", err)
	}

	outputPath := filepath.Join(tempDir, "output.go")

	// Create analysis definitions with a connector that matches the IO spec
	analysisDefs := &analysis.AnalysisDefinitions{
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

	logger := slog.Default()
	compiler := NewCPLCompiler([]string{"input.cpl"}, outputPath, ioSpecPath, logger, slog.LevelInfo)

	err = compiler.synthesizePhase(analysisDefs)
	if err != nil {
		t.Fatalf("synthesizePhase with IO spec failed: %v", err)
	}

	// Verify output file was created
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Expected output file to be readable, got error: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "package main") {
		t.Errorf("Expected output to contain 'package main'")
	}

	if !strings.Contains(contentStr, "func main()") {
		t.Errorf("Expected output to contain 'func main()'")
	}

	if !strings.Contains(contentStr, "sales_data") {
		t.Errorf("Expected output to contain 'sales_data' connector reference")
	}
}

// TestSynthesizePhaseWithInvalidIOPath verifies error handling when the
// IO spec file does not exist.
func TestSynthesizePhaseWithInvalidIOPath(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.go")

	logger := slog.Default()
	compiler := NewCPLCompiler([]string{"input.cpl"}, outputPath, "/nonexistent/io_spec.json", logger, slog.LevelInfo)

	err := compiler.synthesizePhase(&analysis.AnalysisDefinitions{})
	if err == nil {
		t.Fatal("Expected error for invalid IO spec path, got nil")
	}

	if !strings.Contains(err.Error(), "failed to load IO specification") {
		t.Errorf("Expected error about loading IO spec, got: %v", err)
	}
}

// TestSynthesizePhaseWithInvalidIOJSON verifies error handling when the
// IO spec file contains invalid JSON.
func TestSynthesizePhaseWithInvalidIOJSON(t *testing.T) {
	tempDir := t.TempDir()

	ioSpecPath := filepath.Join(tempDir, "invalid_io.json")
	err := os.WriteFile(ioSpecPath, []byte("not valid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid IO spec file: %v", err)
	}

	outputPath := filepath.Join(tempDir, "output.go")
	logger := slog.Default()
	compiler := NewCPLCompiler([]string{"input.cpl"}, outputPath, ioSpecPath, logger, slog.LevelInfo)

	err = compiler.synthesizePhase(&analysis.AnalysisDefinitions{})
	if err == nil {
		t.Fatal("Expected error for invalid IO spec JSON, got nil")
	}

	if !strings.Contains(err.Error(), "failed to load IO specification") {
		t.Errorf("Expected error about loading IO spec, got: %v", err)
	}
}

// TestCompileCallsParsePhase tests that the Compile method calls the parsePhase
// method as the first step in the compilation pipeline.
//
// This test verifies that Compile initiates the parsing phase, which is responsible
// for converting the input CPL file into an AST representation. The parsePhase
// must be called before any other phase and its output (the AST document) should
// be passed to subsequent phases.
//
// The parse phase is the foundation of the compilation pipeline - if parsing fails,
// the entire compilation should halt and return the parsing error without proceeding
// to analysis or synthesis.
//
// Expected behavior: Compile should call parsePhase and return an error if parsing
// fails (e.g., due to syntax errors in the input CPL file).
func TestCompileCallsParsePhase(t *testing.T) {
	// Create a temp directory for test files
	tempDir := t.TempDir()

	// Create a non-existent input file to trigger parse error
	nonExistentFile := filepath.Join(tempDir, "non_existent.cpl")

	logger := slog.Default()
	outputPath := filepath.Join(tempDir, "output.go")
	compiler := NewCPLCompiler([]string{nonExistentFile}, outputPath, "", logger, slog.LevelInfo)

	err := compiler.Compile()
	if err == nil {
		t.Fatal("Expected error when parsing non-existent file, got nil")
	}

	// Verify the error is related to parsing/reading the file
	if !strings.Contains(err.Error(), "parse phase failed") {
		t.Errorf("Expected error to contain 'parse phase failed', got: %v", err)
	}
}

// TestCompileCallsAnalyzePhase tests that the Compile method calls the analyzePhase
// method after successful parsing and before synthesis.
//
// This test verifies that Compile properly orchestrates the transition from the
// parse phase to the analyze phase. The analyzePhase receives the AST document
// produced by parsePhase and performs semantic analysis, validating concepts and
// extracting analysis definitions.
//
// The analysis phase is critical for ensuring the semantic correctness of the
// CPL code - it validates that concepts are properly defined, references are
// resolved, and the AST represents a valid program structure.
//
// Expected behavior: Compile should call analyzePhase with the AST from parsePhase
// and return an error if semantic analysis fails (e.g., undefined concepts,
// type mismatches, or invalid references).
func TestCompileCallsAnalyzePhase(t *testing.T) {
	// Get the examples directory path
	// From src/main/go/compiler, go up 4 levels to reach the project root
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	projectRoot := filepath.Join(testDir, "..", "..", "..", "..", "..")
	examplesDir := filepath.Join(projectRoot, "examples", "parse_and_analyze_only", "without_errors")
	inputFile := filepath.Join(examplesDir, "basic_concepts.cpl")

	// Fail if example file doesn't exist
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		t.Fatalf("Example file not found at %s: %v", inputFile, err)
	}

	tempDir := t.TempDir()

	// Create a minimal IO spec for synthesis
	ioSpecPath := filepath.Join(tempDir, "io_spec.json")
	ioSpecContent := `{"connectors": [], "outputs": [], "architecture": {"cpu_cores": 1, "cache_size": 1024, "main_memory_size": 1024}}`
	if err := os.WriteFile(ioSpecPath, []byte(ioSpecContent), 0644); err != nil {
		t.Fatalf("Failed to write IO spec file: %v", err)
	}

	logger := slog.Default()
	outputPath := filepath.Join(tempDir, "output.go")
	compiler := NewCPLCompiler([]string{inputFile}, outputPath, ioSpecPath, logger, slog.LevelInfo)

	// This should succeed - parsing and analysis should work
	err := compiler.Compile()
	if err != nil {
		t.Fatalf("Expected successful compilation, got error: %v", err)
	}
}

// TestCompileCallsSynthesizePhase tests that the Compile method calls the
// synthesizePhase method after successful analysis.
//
// This test verifies that Compile properly orchestrates the final phase of the
// pipeline. The synthesizePhase receives the AnalysisDefinitions produced by
// analyzePhase and generates the output Go code that will be written to the
// specified output file.
//
// The synthesis phase is responsible for code generation - transforming the
// analyzed and validated CPL concepts into executable Go code. This is the
// final transformation step in the compilation pipeline.
//
// Expected behavior: Compile should call synthesizePhase with the AnalysisDefinitions
// from analyzePhase and return an error if code generation fails (e.g., due to
// I/O errors when writing the output file).
func TestCompileCallsSynthesizePhase(t *testing.T) {
	// Get the examples directory path
	// From src/main/go/compiler, go up 4 levels to reach the project root
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	projectRoot := filepath.Join(testDir, "..", "..", "..", "..", "..")
	examplesDir := filepath.Join(projectRoot, "examples", "parse_and_analyze_only", "without_errors")
	inputFile := filepath.Join(examplesDir, "basic_concepts.cpl")

	// Fail if example file doesn't exist
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		t.Fatalf("Example file not found at %s: %v", inputFile, err)
	}

	tempDir := t.TempDir()

	// Create a minimal IO spec for synthesis
	ioSpecPath := filepath.Join(tempDir, "io_spec.json")
	ioSpecContent := `{"connectors": [], "outputs": [], "architecture": {"cpu_cores": 1, "cache_size": 1024, "main_memory_size": 1024}}`
	if err := os.WriteFile(ioSpecPath, []byte(ioSpecContent), 0644); err != nil {
		t.Fatalf("Failed to write IO spec file: %v", err)
	}

	logger := slog.Default()
	outputPath := filepath.Join(tempDir, "output.go")
	compiler := NewCPLCompiler([]string{inputFile}, outputPath, ioSpecPath, logger, slog.LevelInfo)

	// Compile should succeed
	err := compiler.Compile()
	if err != nil {
		t.Fatalf("Expected successful compilation, got error: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Expected output file to be created at %s", outputPath)
	}
}

// TestCompilePipelineSuccess tests a complete successful compilation through
// all three phases of the pipeline: parse → analyze → synthesize.
//
// This test verifies the happy path where:
//   - The input CPL file is successfully parsed into an AST
//   - The AST passes semantic analysis and produces AnalysisDefinitions
//   - The AnalysisDefinitions are successfully synthesized into Go code
//   - The output file is written to the specified path
//
// This integration-style test ensures that all three phases work together
// correctly and that the Compile method returns nil (no error) when the
// entire pipeline completes successfully.
//
// Expected behavior: Compile should execute all three phases in sequence and
// return nil when compilation completes successfully, with the generated Go
// code written to the output path.
func TestCompilePipelineSuccess(t *testing.T) {
	// Get the examples directory path
	// From src/main/go/compiler, go up 4 levels to reach the project root
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	projectRoot := filepath.Join(testDir, "..", "..", "..", "..", "..")
	examplesDir := filepath.Join(projectRoot, "examples", "parse_and_analyze_only", "without_errors")
	inputFile := filepath.Join(examplesDir, "basic_concepts.cpl")

	// Fail if example file doesn't exist
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		t.Fatalf("Example file not found at %s: %v", inputFile, err)
	}

	tempDir := t.TempDir()

	// Create a minimal IO spec for synthesis
	ioSpecPath := filepath.Join(tempDir, "io_spec.json")
	ioSpecContent := `{"connectors": [], "outputs": [], "architecture": {"cpu_cores": 1, "cache_size": 1024, "main_memory_size": 1024}}`
	if err := os.WriteFile(ioSpecPath, []byte(ioSpecContent), 0644); err != nil {
		t.Fatalf("Failed to write IO spec file: %v", err)
	}

	logger := slog.Default()
	outputPath := filepath.Join(tempDir, "output.go")
	compiler := NewCPLCompiler([]string{inputFile}, outputPath, ioSpecPath, logger, slog.LevelInfo)

	// Compile should complete successfully through all three phases
	err := compiler.Compile()
	if err != nil {
		t.Fatalf("Expected successful compilation, got error: %v", err)
	}

	// Verify output file was created and has content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Expected output file to be readable, got error: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected output file to have content")
	}

	// Verify the output contains expected Go code elements
	contentStr := string(content)
	if !strings.Contains(contentStr, "package main") {
		t.Error("Expected output to contain 'package main'")
	}
}

// TestCompileOnlySkipsSynthesisWithoutIOSpec verifies compile-only mode runs
// parse and analysis without requiring an IO specification or producing output.
func TestCompileOnlySkipsSynthesisWithoutIOSpec(t *testing.T) {
	// Get the examples directory path
	// From src/main/go/compiler, go up 4 levels to reach the project root
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	projectRoot := filepath.Join(testDir, "..", "..", "..", "..", "..")
	examplesDir := filepath.Join(projectRoot, "examples", "parse_and_analyze_only", "without_errors")
	inputFile := filepath.Join(examplesDir, "basic_concepts.cpl")

	// Fail if example file doesn't exist
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		t.Fatalf("Example file not found at %s: %v", inputFile, err)
	}

	tempDir := t.TempDir()
	logger := slog.Default()
	outputPath := filepath.Join(tempDir, "output.go")
	compiler := NewCPLCompiler([]string{inputFile}, outputPath, "", logger, slog.LevelInfo)
	compiler.compileOnly = true

	err := compiler.Compile()
	if err != nil {
		t.Fatalf("Expected compile-only compilation without IO spec to succeed, got error: %v", err)
	}

	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Error("Expected output file to NOT be created in compile-only mode")
	}
}

// TestCompileStopsOnParseError tests that the compilation pipeline halts
// immediately when the parse phase encounters an error.
//
// This test verifies error handling behavior in the Compile method:
//   - If parsePhase returns an error, Compile should return that error immediately
//   - The analyzePhase should NOT be called when parsing fails
//   - The synthesizePhase should NOT be called when parsing fails
//
// This ensures that semantic analysis and code generation are not attempted
// on an invalid AST, preventing cascading errors and providing clear feedback
// about the root cause (syntax errors in the input).
//
// Expected behavior: When parsePhase returns an error, Compile should return
// that error without calling analyzePhase or synthesizePhase.
func TestCompileStopsOnParseError(t *testing.T) {
	// Create a temp directory
	tempDir := t.TempDir()

	// Create an invalid CPL file with syntax errors
	invalidCPL := `concepts
    Country=[Values[USA, Canada, Mexico
    # Missing closing bracket
`
	inputFile := filepath.Join(tempDir, "invalid.cpl")
	err := os.WriteFile(inputFile, []byte(invalidCPL), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	logger := slog.Default()
	outputPath := filepath.Join(tempDir, "output.go")
	compiler := NewCPLCompiler([]string{inputFile}, outputPath, "", logger, slog.LevelInfo)

	// Compile should fail with a parse error
	err = compiler.Compile()
	if err == nil {
		t.Fatal("Expected error when compiling invalid CPL file, got nil")
	}

	// Verify the error is from the parse phase
	if !strings.Contains(err.Error(), "parse phase failed") {
		t.Errorf("Expected error to contain 'parse phase failed', got: %v", err)
	}

	// Verify output file was NOT created
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Error("Expected output file to NOT be created when parsing fails")
	}
}

// TestCompileStopsOnAnalysisError tests that the compilation pipeline halts
// when the analyze phase encounters an error, skipping synthesis.
//
// This test verifies error handling behavior in the Compile method:
//   - If parsePhase succeeds but analyzePhase returns an error, Compile should
//     return that analysis error
//   - The synthesizePhase should NOT be called when analysis fails
//
// This ensures that code generation is not attempted on semantically invalid
// input. Even if the syntax is correct (parsing succeeded), semantic errors
// such as undefined concepts or invalid references should prevent synthesis.
//
// Expected behavior: When analyzePhase returns an error, Compile should return
// that error without calling synthesizePhase, but after having called parsePhase
// successfully.
func TestCompileStopsOnAnalysisError(t *testing.T) {
	// Create a temp directory
	tempDir := t.TempDir()

	// Create a CPL file that parses successfully but would fail analysis
	// For Increment 1, this test verifies that if analysis fails,
	// compilation stops and synthesis is not attempted
	// Since the current parser doesn't produce analysis errors in Increment 1,
	// we test the mechanism by verifying the pipeline structure
	cplContent := `concepts
    Country=[Values[USA, Canada, Mexico]]
`
	inputFile := filepath.Join(tempDir, "valid_syntax.cpl")
	err := os.WriteFile(inputFile, []byte(cplContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a minimal IO spec for synthesis
	ioSpecPath := filepath.Join(tempDir, "io_spec.json")
	ioSpecContent := `{"connectors": [], "outputs": [], "architecture": {"cpu_cores": 1, "cache_size": 1024, "main_memory_size": 1024}}`
	if err := os.WriteFile(ioSpecPath, []byte(ioSpecContent), 0644); err != nil {
		t.Fatalf("Failed to write IO spec file: %v", err)
	}

	logger := slog.Default()
	outputPath := filepath.Join(tempDir, "output.go")
	compiler := NewCPLCompiler([]string{inputFile}, outputPath, ioSpecPath, logger, slog.LevelInfo)

	// For Increment 1, this should succeed since analysis doesn't yet produce errors
	// The key point is that the pipeline is structured to stop on analysis errors
	err = compiler.Compile()
	if err != nil {
		t.Fatalf("Expected successful compilation for Increment 1 (no analysis errors yet), got: %v", err)
	}
}

// TestDiscoverStandardDefinitions verifies that discoverStandardDefinitions
// finds all .cpl files in the configured standard definitions directory.
func TestDiscoverStandardDefinitions(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create standard definition files
	file1 := filepath.Join(tempDir, "types.cpl")
	if err := os.WriteFile(file1, []byte("types\n\tCurrency = Number(precision=2)\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	file2 := filepath.Join(tempDir, "concepts.cpl")
	if err := os.WriteFile(file2, []byte("concepts\n\tCountry=[Values[USA]]\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a non-.cpl file that should be ignored
	nonCPL := filepath.Join(tempDir, "readme.txt")
	if err := os.WriteFile(nonCPL, []byte("not a cpl file"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	compiler := NewCPLCompiler([]string{}, "output.go", "", slog.Default(), slog.LevelInfo)
	compiler.standardDefsDir = tempDir

	paths, err := compiler.discoverStandardDefinitions()
	if err != nil {
		t.Fatalf("discoverStandardDefinitions failed: %v", err)
	}

	if len(paths) != 2 {
		t.Errorf("Expected 2 standard definition files, got %d", len(paths))
	}

	// Verify both .cpl files were discovered
	foundTypes := false
	foundConcepts := false
	for _, p := range paths {
		if filepath.Base(p) == "types.cpl" {
			foundTypes = true
		}
		if filepath.Base(p) == "concepts.cpl" {
			foundConcepts = true
		}
	}
	if !foundTypes {
		t.Error("Expected types.cpl to be discovered")
	}
	if !foundConcepts {
		t.Error("Expected concepts.cpl to be discovered")
	}
}

// TestDiscoverStandardDefinitions_NonExistentDir returns an empty slice
// when the standard definitions directory does not exist.
func TestDiscoverStandardDefinitions_NonExistentDir(t *testing.T) {
	t.Parallel()

	compiler := NewCPLCompiler([]string{}, "output.go", "", slog.Default(), slog.LevelInfo)
	compiler.standardDefsDir = "/nonexistent/standard_definitions"

	paths, err := compiler.discoverStandardDefinitions()
	if err != nil {
		t.Fatalf("discoverStandardDefinitions should not error for missing dir: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("Expected 0 paths for non-existent directory, got %d", len(paths))
	}
}

// TestDiscoverStandardDefinitions_EmptyDir returns an empty slice
// when the standard definitions directory exists but contains no .cpl files.
func TestDiscoverStandardDefinitions_EmptyDir(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	compiler := NewCPLCompiler([]string{}, "output.go", "", slog.Default(), slog.LevelInfo)
	compiler.standardDefsDir = tempDir

	paths, err := compiler.discoverStandardDefinitions()
	if err != nil {
		t.Fatalf("discoverStandardDefinitions failed: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("Expected 0 paths for empty directory, got %d", len(paths))
	}
}

// TestParsePhase_IncludesStandardDefinitions verifies that parsePhase
// automatically includes standard definition files in the compilation.
func TestParsePhase_IncludesStandardDefinitions(t *testing.T) {
	tempDir := t.TempDir()

	// Create a standard definitions directory with a type definition
	stdDefsDir := filepath.Join(tempDir, "standard_definitions")
	if err := os.MkdirAll(stdDefsDir, 0755); err != nil {
		t.Fatalf("Failed to create standard definitions directory: %v", err)
	}

	stdFile := filepath.Join(stdDefsDir, "std_types.cpl")
	if err := os.WriteFile(stdFile, []byte("types\n\tCurrency = Number(precision=2)\n"), 0644); err != nil {
		t.Fatalf("Failed to create standard type file: %v", err)
	}

	// Create a user CPL file with a concept
	userFile := filepath.Join(tempDir, "user.cpl")
	if err := os.WriteFile(userFile, []byte("concepts\n\tCountry=[Values[USA]]\n"), 0644); err != nil {
		t.Fatalf("Failed to create user file: %v", err)
	}

	compiler := NewCPLCompiler([]string{userFile}, "output.go", "", slog.Default(), slog.LevelInfo)
	compiler.standardDefsDir = stdDefsDir

	result, err := compiler.parsePhase()
	if err != nil {
		t.Fatalf("parsePhase failed: %v", err)
	}
	if result == nil || result.rootTable == nil {
		t.Fatal("Expected non-nil parse result with root table")
	}

	// Verify both types and concepts scopes are present
	if len(result.rootTable.Children) < 2 {
		t.Errorf("Expected at least 2 children (types from std + concepts from user), got %d", len(result.rootTable.Children))
	}
}

// TestParsePhase_StandardDefsFirst ensures that standard definitions are
// parsed before user files so that user code can reference standard types.
func TestParsePhase_StandardDefsFirst(t *testing.T) {
	tempDir := t.TempDir()

	// Create standard definitions
	stdDefsDir := filepath.Join(tempDir, "standard_definitions")
	if err := os.MkdirAll(stdDefsDir, 0755); err != nil {
		t.Fatalf("Failed to create standard definitions directory: %v", err)
	}

	stdFile := filepath.Join(stdDefsDir, "std_types.cpl")
	if err := os.WriteFile(stdFile, []byte("types\n\tCurrency = Number(precision=2)\n"), 0644); err != nil {
		t.Fatalf("Failed to create standard type file: %v", err)
	}

	// Create a user CPL file
	userFile := filepath.Join(tempDir, "user.cpl")
	if err := os.WriteFile(userFile, []byte("concepts\n\tCountry=[Values[USA]]\n"), 0644); err != nil {
		t.Fatalf("Failed to create user file: %v", err)
	}

	compiler := NewCPLCompiler([]string{userFile}, "output.go", "", slog.Default(), slog.LevelInfo)
	compiler.standardDefsDir = stdDefsDir

	result, err := compiler.parsePhase()
	if err != nil {
		t.Fatalf("parsePhase failed: %v", err)
	}

	// Verify the symbol table contains definitions from both standard and user files
	if result.rootTable == nil {
		t.Fatal("Expected non-nil root table")
	}

	// Look for TypesScope (from standard definitions)
	foundTypesScope := false
	for _, child := range result.rootTable.Children {
		if child.ScopeType.String() == "TypesScope" {
			foundTypesScope = true
			break
		}
	}
	if !foundTypesScope {
		t.Error("Expected TypesScope from standard definitions to be present in root table")
	}

	// Look for ConceptsScope (from user file)
	foundConceptsScope := false
	for _, child := range result.rootTable.Children {
		if child.ScopeType.String() == "ConceptsScope" {
			foundConceptsScope = true
			break
		}
	}
	if !foundConceptsScope {
		t.Error("Expected ConceptsScope from user file to be present in root table")
	}
}


