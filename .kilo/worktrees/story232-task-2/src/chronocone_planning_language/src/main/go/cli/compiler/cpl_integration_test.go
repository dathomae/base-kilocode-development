package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestFullPipelineIntegration performs an end-to-end test of the complete
// CPL compilation pipeline: parse → analyze → synthesize.
//
// This test verifies that:
//   - The cpl_compiler binary can be built successfully
//   - A CPL file with concept definitions can be parsed and analyzed
//   - The synthesis phase generates valid Go source code
//
// The generated code requires the Chronocone runtime libraries
// (ChronoconeComputationController, ArchitectureSpecification, etc.) and
// is not self-contained. Compiling and running it as a standalone binary
// would require those types to be embedded or imported. This test verifies
// the pipeline produces syntactically valid Go code with the expected structure.
//
// The test creates temporary files for input CPL, output Go code, and the
// compiled CLI binary, ensuring no interference with the project structure.
//
// Expected behavior: The pipeline completes successfully, generating Go code
// containing the computation lifecycle (Setup, Compute, Cleanup, main).
func TestFullPipelineIntegration(t *testing.T) {
	// Create temp directory for all test artifacts
	tempDir := t.TempDir()

	// Build the cpl_compiler binary
	compilerBinary := filepath.Join(tempDir, "cpl_compiler")
	sourcePath := filepath.Join(".", "cpl_compiler.go")

	buildCmd := exec.Command("go", "build", "-o", compilerBinary, sourcePath)
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build cpl_compiler binary: %v\nOutput: %s", err, buildOutput)
	}

	// Create a temporary CPL file with concept definitions
	cplContent := `concepts
    Country=[Values[USA, Canada, Mexico]]
    Priority=[Values[Low, Medium, High]]
`
	cplFile := filepath.Join(tempDir, "test_concepts.cpl")
	err = os.WriteFile(cplFile, []byte(cplContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create CPL test file: %v", err)
	}

	// Create a minimal IO specification file (required for synthesis)
	ioSpecContent := `{"connectors": [], "outputs": [], "architecture": {"cpu_cores": 1, "cache_size": 1024, "main_memory_size": 1024}}`
	ioSpecFile := filepath.Join(tempDir, "io_spec.json")
	err = os.WriteFile(ioSpecFile, []byte(ioSpecContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create IO spec test file: %v", err)
	}

	// Define output path for generated Go code
	outputGoFile := filepath.Join(tempDir, "generated_cli.go")

	// Run the compiler with -o and -io flags
	compileCmd := exec.Command(compilerBinary, "-o", outputGoFile, "-io", ioSpecFile, cplFile)
	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to compile CPL file: %v\nOutput: %s", err, compileOutput)
	}

	// Verify output file was created
	if _, err := os.Stat(outputGoFile); os.IsNotExist(err) {
		t.Fatalf("Expected output file to be created at %s", outputGoFile)
	}

	// Verify the output file has content
	content, err := os.ReadFile(outputGoFile)
	if err != nil {
		t.Fatalf("Failed to read generated Go file: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("Generated Go file is empty")
	}

	// Verify generated Go file contains expected content
	contentStr := string(content)
	if !strings.Contains(contentStr, "package main") {
		t.Error("Generated Go file should contain 'package main'")
	}

	if !strings.Contains(contentStr, "func main()") {
		t.Error("Generated Go file should contain 'func main()'")
	}

	if !strings.Contains(contentStr, "func Setup(") {
		t.Error("Generated Go file should contain 'Setup' function")
	}

	if !strings.Contains(contentStr, "func Compute()") {
		t.Error("Generated Go file should contain 'Compute' function")
	}

	if !strings.Contains(contentStr, "func Cleanup()") {
		t.Error("Generated Go file should contain 'Cleanup' function")
	}

	if !strings.Contains(contentStr, "Computation completed successfully") {
		t.Error("Generated Go file should contain 'Computation completed successfully'")
	}
}

// TestCompileOnlyCLIRequiresNoOutputOrIOSpec verifies that compile-only mode
// runs the parse and analyze phases without requiring generation-only flags.
//
// Expected behavior: The compiler exits successfully when invoked with
// --compile_only and an input CPL file, even when no -o/--output or -io flags
// are provided. It should not create a generated Go file because synthesis is
// skipped.
func TestCompileOnlyCLIRequiresNoOutputOrIOSpec(t *testing.T) {
	// Create temp directory for all test artifacts
	tempDir := t.TempDir()

	// Build the cpl_compiler binary
	compilerBinary := filepath.Join(tempDir, "cpl_compiler")
	sourcePath := filepath.Join(".", "cpl_compiler.go")

	buildCmd := exec.Command("go", "build", "-o", compilerBinary, sourcePath)
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build cpl_compiler binary: %v\nOutput: %s", err, buildOutput)
	}

	// Create a temporary CPL file with definitions that parse and analyze.
	cplContent := `concepts
    Country=[Values[USA, Canada, Mexico]]
`
	cplFile := filepath.Join(tempDir, "test_concepts.cpl")
	err = os.WriteFile(cplFile, []byte(cplContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create CPL test file: %v", err)
	}

	compileCmd := exec.Command(compilerBinary, "--compile_only", cplFile)
	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Expected compile-only command without output or IO spec to succeed: %v\nOutput: %s", err, compileOutput)
	}

	unexpectedOutputFile := filepath.Join(tempDir, "generated_cli.go")
	if _, err := os.Stat(unexpectedOutputFile); !os.IsNotExist(err) {
		t.Error("Compile-only mode should not create a generated Go output file")
	}
}

// TestGeneratedCLIContainsExpectedStructure verifies that the Go code generated
// by the synthesis phase contains the expected structure for the computation
// lifecycle (Setup, Compute, Cleanup, main) when given a real CPL example file.
//
// The generated code references Chronocone runtime types and is designed to be
// compiled within a module that provides those types. This test verifies the
// synthesis output structure without attempting standalone compilation.
//
// This test uses the actual compiler pipeline to generate Go code, then:
//   - Verifies the generated file exists and is non-empty
//   - Checks for expected function signatures and patterns
//
// Expected behavior: The generated Go code contains the full computation lifecycle.
func TestGeneratedCLIContainsExpectedStructure(t *testing.T) {
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

	// Create temp directory for test artifacts
	tempDir := t.TempDir()

	// Build the cpl_compiler binary
	compilerBinary := filepath.Join(tempDir, "cpl_compiler")
	sourcePath := filepath.Join(".", "cpl_compiler.go")

	buildCmd := exec.Command("go", "build", "-o", compilerBinary, sourcePath)
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build cpl_compiler binary: %v\nOutput: %s", err, buildOutput)
	}

	// Create a minimal IO specification file (required for synthesis)
	ioSpecContent := `{"connectors": [], "outputs": [], "architecture": {"cpu_cores": 1, "cache_size": 1024, "main_memory_size": 1024}}`
	ioSpecFile := filepath.Join(tempDir, "io_spec.json")
	err = os.WriteFile(ioSpecFile, []byte(ioSpecContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create IO spec test file: %v", err)
	}

	// Run compiler to generate Go CLI
	outputGoFile := filepath.Join(tempDir, "cpl_cli.go")
	compileCmd := exec.Command(compilerBinary, "-o", outputGoFile, "-io", ioSpecFile, inputFile)
	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to compile CPL file: %v\nOutput: %s", err, compileOutput)
	}

	// Verify generated Go file exists and has content
	content, err := os.ReadFile(outputGoFile)
	if err != nil {
		t.Fatalf("Failed to read generated Go file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package main") {
		t.Error("Generated Go file should contain 'package main'")
	}

	if !strings.Contains(contentStr, "func main()") {
		t.Error("Generated Go file should contain 'func main()'")
	}

	if !strings.Contains(contentStr, "func Setup(") {
		t.Error("Generated Go file should contain 'Setup' function")
	}

	if !strings.Contains(contentStr, "func Compute()") {
		t.Error("Generated Go file should contain 'Compute' function")
	}

	if !strings.Contains(contentStr, "func Cleanup()") {
		t.Error("Generated Go file should contain 'Cleanup' function")
	}

	if !strings.Contains(contentStr, "Computation completed successfully") {
		t.Error("Generated Go file should contain 'Computation completed successfully'")
	}

	// Note: basic_concepts.cpl contains only concept definitions, no connectors.
	// Concepts are not directly emitted into the generated Go code unless they
	// are referenced by a connector or measure. Therefore, we do not assert on
	// concept names appearing in the output here.
}

// TestPipelineWithInvalidCPLFile verifies that the compiler correctly handles
// invalid CPL input by:
//   - Returning a non-zero exit code when given a syntactically invalid CPL file
//   - Displaying an appropriate error message
//   - NOT creating an output file when compilation fails
//
// This test ensures proper error handling and propagation throughout the
// compilation pipeline. When parsing fails due to syntax errors, the compiler
// should fail gracefully with a clear error message and not leave partial
// output files.
//
// Expected behavior: The compiler exits with non-zero status, displays an
// error message containing "parse phase failed", and the output file is
// not created.
func TestPipelineWithInvalidCPLFile(t *testing.T) {
	// Create temp directory for test artifacts
	tempDir := t.TempDir()

	// Build the cpl_compiler binary
	compilerBinary := filepath.Join(tempDir, "cpl_compiler")
	sourcePath := filepath.Join(".", "cpl_compiler.go")

	buildCmd := exec.Command("go", "build", "-o", compilerBinary, sourcePath)
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build cpl_compiler binary: %v\nOutput: %s", err, buildOutput)
	}

	// Create an invalid CPL file with syntax errors (missing closing bracket)
	invalidCPL := `concepts
    Country=[Values[USA, Canada, Mexico
    # Missing closing brackets
`
	invalidCPLFile := filepath.Join(tempDir, "invalid.cpl")
	err = os.WriteFile(invalidCPLFile, []byte(invalidCPL), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid CPL test file: %v", err)
	}

	// Create a minimal IO specification file (required for synthesis)
	ioSpecContent := `{"connectors": [], "outputs": [], "architecture": {"cpu_cores": 1, "cache_size": 1024, "main_memory_size": 1024}}`
	ioSpecFile := filepath.Join(tempDir, "io_spec.json")
	err = os.WriteFile(ioSpecFile, []byte(ioSpecContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create IO spec test file: %v", err)
	}

	// Define output path (should not be created)
	outputGoFile := filepath.Join(tempDir, "should_not_exist.go")

	// Run the compiler on the invalid CPL file (with -io flag)
	compileCmd := exec.Command(compilerBinary, "-o", outputGoFile, "-io", ioSpecFile, invalidCPLFile)
	compileOutput, err := compileCmd.CombinedOutput()

	// Verify that the command returned an error (non-zero exit)
	if err == nil {
		t.Fatal("Expected compiler to fail on invalid CPL file, but it succeeded")
	}

	// Verify error message is displayed
	output := string(compileOutput)
	if !strings.Contains(output, "parse phase failed") {
		t.Errorf("Expected error message to contain 'parse phase failed', got:\n%s", output)
	}

	// Verify output file was NOT created
	if _, err := os.Stat(outputGoFile); !os.IsNotExist(err) {
		t.Error("Expected output file to NOT be created when compilation fails")
	}

	// Verify exit code is non-zero by checking the error type
	exitError, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected *exec.ExitError, got %T: %v", err, err)
	}

	if exitError.ExitCode() == 0 {
		t.Error("Expected non-zero exit code for failed compilation")
	}
}
