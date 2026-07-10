// Package main contains integration tests that exercise all examples under
// examples/with_data by dynamically discovering them and running the full
// compile → run → verify pipeline for each one.
package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	cplio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getWithDataDir returns the path to the examples/with_data/without_errors directory.
func getWithDataDir(t *testing.T) string {
	t.Helper()
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	projectRoot := filepath.Join(testDir, "..", "..", "..", "..", "..")
	return filepath.Join(projectRoot, "examples", "with_data", "without_errors")
}

// getWithDataErrorsDir returns the path to the examples/with_data/with_errors directory.
func getWithDataErrorsDir(t *testing.T) string {
	t.Helper()
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	projectRoot := filepath.Join(testDir, "..", "..", "..", "..", "..")
	return filepath.Join(projectRoot, "examples", "with_data", "with_errors")
}

// findModuleRoot locates the Go module root by searching upward from the
// test file for a go.mod file.
func findModuleRootFromTest(t *testing.T) string {
	t.Helper()
	_, testFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(testFile)
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if info, err := os.Stat(goModPath); err == nil && !info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("cannot find module root")
		}
		dir = parent
	}
}

// getOutputParquetPaths reads the iospec.json in the given directory and returns
// the expected output parquet file paths based on the output metric names.
// Each output metric produces a <metricName>.parquet file.
func getOutputParquetPaths(iospecPath string) ([]string, error) {
	data, err := os.ReadFile(iospecPath)
	if err != nil {
		return nil, err
	}

	var spec struct {
		Outputs []struct {
			Metric string `json:"metric"`
		} `json:"outputs"`
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	var paths []string
	for _, output := range spec.Outputs {
		paths = append(paths, output.Metric+".parquet")
	}
	return paths, nil
}

// exampleCase represents a single discovered example directory.
type exampleCase struct {
	name         string
	dir          string
	parentDir    string
	cplFiles     []string
	hasCoreCPL   bool
	hasIOSpec    bool
	hasParquet   bool
	hasExpected  bool
	isCompileOnly bool
}

// discoverWithDataExamples walks the examples/with_data/without_errors tree and returns all
// leaf directories that contain iospec.json, at least one .cpl file, and at
// least one .parquet file.
func discoverWithDataExamples(t *testing.T) []exampleCase {
	t.Helper()
	withDataDir := getWithDataDir(t)

	entries, err := os.ReadDir(withDataDir)
	require.NoError(t, err, "Failed to read without_errors directory")

	var examples []exampleCase

	for _, categoryEntry := range entries {
		if !categoryEntry.IsDir() {
			continue
		}
		categoryDir := filepath.Join(withDataDir, categoryEntry.Name())

		// Check for core.cpl in the category directory
		coreCPLPath := filepath.Join(categoryDir, "core.cpl")
		hasCoreCPL := false
		if _, err := os.Stat(coreCPLPath); err == nil {
			hasCoreCPL = true
		}

		// Walk example subdirectories
		subEntries, err := os.ReadDir(categoryDir)
		require.NoError(t, err, "Failed to read category directory %s", categoryDir)

		for _, subEntry := range subEntries {
			if !subEntry.IsDir() {
				continue
			}
			exampleDir := filepath.Join(categoryDir, subEntry.Name())

			// Check for required files
			iospecPath := filepath.Join(exampleDir, "iospec.json")
			_, iospecErr := os.Stat(iospecPath)
			hasIOSpec := (iospecErr == nil)

			// Find .cpl files
			var cplFiles []string
			cplEntries, _ := os.ReadDir(exampleDir)
			for _, cplEntry := range cplEntries {
				if !cplEntry.IsDir() && strings.HasSuffix(cplEntry.Name(), ".cpl") {
					cplFiles = append(cplFiles, filepath.Join(exampleDir, cplEntry.Name()))
				}
			}

			// Check for parquet files
			hasParquet := false
			for _, entry := range cplEntries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".parquet") {
					hasParquet = true
					break
				}
			}

			// Check for expected_output.xlsx
			expectedPath := filepath.Join(exampleDir, "expected_output.xlsx")
			_, expectedErr := os.Stat(expectedPath)
			hasExpected := (expectedErr == nil)

			// Check for compile-only marker
			compileOnlyPath := filepath.Join(exampleDir, ".compile_only")
			_, compileOnlyErr := os.Stat(compileOnlyPath)
			isCompileOnly := (compileOnlyErr == nil)

			// Only include directories that have the minimum required files
			if hasIOSpec && len(cplFiles) > 0 && hasParquet {
				examples = append(examples, exampleCase{
					name:         subEntry.Name(),
					dir:          exampleDir,
					parentDir:    categoryDir,
					cplFiles:     cplFiles,
					hasCoreCPL:   hasCoreCPL,
					hasIOSpec:    hasIOSpec,
					hasParquet:   hasParquet,
					hasExpected:  hasExpected,
					isCompileOnly: isCompileOnly,
				})
			}
		}
	}

	return examples
}

// TestWithDataExamples dynamically discovers all examples under
// examples/with_data and runs the full compile → run → verify pipeline for
// each one. This test mirrors the behavior of shell_scripts/compile_and_run.sh.
//
// For each discovered example directory:
//   1. Collect all .cpl files in the example directory
//   2. Prepend core.cpl from the parent category directory if it exists
//   3. Compile the CPL files using the CPL compiler with the iospec.json
//   4. Run the generated Go program with -data-dir pointing at the example dir
//   5. If expected_output.xlsx exists, verify output.parquet against it
func TestWithDataExamples(t *testing.T) {
	examples := discoverWithDataExamples(t)

	moduleRoot := findModuleRootFromTest(t)

	for _, ex := range examples {
		ex := ex // capture range variable
		t.Run(ex.name, func(t *testing.T) {
			t.Parallel()

			// Build the list of CPL files: core.cpl first (if present), then example CPLs
			var cplPaths []string
			if ex.hasCoreCPL {
				cplPaths = append(cplPaths, filepath.Join(ex.parentDir, "core.cpl"))
			}
			cplPaths = append(cplPaths, ex.cplFiles...)

			// Create a temp directory for generated output under the module root
			// so that imports resolve correctly.
			outputDir := filepath.Join(moduleRoot, "generated", "cmd", ex.name)
			require.NoError(t, os.MkdirAll(outputDir, 0755), "Failed to create output directory")
			t.Cleanup(func() { os.RemoveAll(outputDir) })

			outputPath := filepath.Join(outputDir, "main.go")
			iospecPath := filepath.Join(ex.dir, "iospec.json")

			// Phase 1: Compile CPL files
			logger := slog.Default()
			compiler := NewCPLCompiler(cplPaths, outputPath, iospecPath, logger, slog.LevelError)
			// Point standardDefsDir at the absolute path so it resolves regardless of CWD
			compiler.standardDefsDir = filepath.Join(moduleRoot, "standard_definitions")
			err := compiler.Compile()
			require.NoError(t, err, "Compilation failed for %s", ex.name)

			// Compile-only examples are verified by dedicated tests and compile-only
			// integration tests; skip runtime execution here.
			if ex.isCompileOnly {
				t.Skipf("Example %s is marked compile-only (.compile_only), skipping runtime verification", ex.name)
			}

			// Phase 2: Run the generated program
			cmd := exec.Command("go", "run", outputPath)
			cmd.Dir = ex.dir // Run from example dir so relative parquet paths resolve
			runOutput, err := cmd.CombinedOutput()
			require.NoError(t, err, "Generated program failed for %s\nOutput:\n%s", ex.name, string(runOutput))

			// Phase 3: Verify output against expected_output.xlsx
			expectedPath := filepath.Join(ex.dir, "expected_output.xlsx")
			require.True(t, ex.hasExpected, "expected_output.xlsx is missing for example %s", ex.name)

			outputParquets, err := getOutputParquetPaths(iospecPath)
			require.NoError(t, err, "Failed to read iospec for %s", ex.name)
			require.NotEmpty(t, outputParquets, "iospec has no outputs for example %s", ex.name)

			// Verify each output metric has its corresponding parquet file.
			// For single-output examples, this is the metric-named parquet file.
			// For multi-output examples, pass -metric to filter the xlsx to the
			// specific metric being verified.
			for _, parquetName := range outputParquets {
				outputParquet := filepath.Join(ex.dir, parquetName)
				metricName := strings.TrimSuffix(parquetName, ".parquet")

				verifyCmd := exec.Command("go", "run",
					filepath.Join(moduleRoot, "src", "main", "go", "cli", "xls_verify", "main.go"),
					"-parquet="+outputParquet,
					"-xlsx="+expectedPath,
					"-metric="+metricName,
				)
				verifyOutput, err := verifyCmd.CombinedOutput()
				assert.NoError(t, err, "Verification failed for %s (parquet: %s)\nOutput:\n%s", ex.name, parquetName, string(verifyOutput))
			}
		})
	}
}

// TestWithDataExamples_Discovery verifies that the example discovery mechanism
// finds at least the known example directories. This serves as a sanity check
// that the discovery logic itself is working.
func TestWithDataExamples_Discovery(t *testing.T) {
	examples := discoverWithDataExamples(t)
	require.NotEmpty(t, examples, "Expected to discover at least one with_data example")

	// Build a set of discovered names
	discovered := make(map[string]bool)
	for _, ex := range examples {
		discovered[ex.name] = true
	}

	// Verify some known examples are present
	knownExamples := []string{
		"simple_addition",
		"conditional_comparison_operators",
		"conditional_if_else_arithmetic",
		"conditional_if_elseif_else",
	}

	for _, name := range knownExamples {
		assert.True(t, discovered[name], "Expected to discover example %q", name)
	}
}

// TestWithDataExamples_ExpectedOutputExists verifies that every discovered
// example that is intended for full end-to-end execution has an
// expected_output.xlsx file. Compile-only examples (marked with a
// .compile_only file in the example directory) are excluded from this check.
func TestWithDataExamples_ExpectedOutputExists(t *testing.T) {
	examples := discoverWithDataExamples(t)
	require.NotEmpty(t, examples, "Expected to discover at least one with_data example")

	var missing []string
	for _, ex := range examples {
		if ex.isCompileOnly {
			continue
		}
		if !ex.hasExpected {
			missing = append(missing, ex.name)
		}
	}

	if len(missing) > 0 {
		t.Errorf("The following %d example(s) are missing expected_output.xlsx: %v", len(missing), missing)
	}
}

// TestWithDataExamples_Structure verifies structural properties of discovered
// examples (each has iospec.json, at least one .cpl, at least one .parquet).
func TestWithDataExamples_Structure(t *testing.T) {
	examples := discoverWithDataExamples(t)
	require.NotEmpty(t, examples, "Expected to discover at least one with_data example")

	for _, ex := range examples {
		assert.True(t, ex.hasIOSpec, "Example %q missing iospec.json", ex.name)
		assert.NotEmpty(t, ex.cplFiles, "Example %q has no .cpl files", ex.name)
		assert.True(t, ex.hasParquet, "Example %q has no .parquet files", ex.name)
	}
}

// TestWithDataExamples_CompileOnly verifies that all discovered examples can
// at least compile successfully (parse + analyze) without running them. This
// provides faster feedback when only compilation correctness is in question.
func TestWithDataExamples_CompileOnly(t *testing.T) {
	examples := discoverWithDataExamples(t)

	for _, ex := range examples {
		ex := ex // capture range variable
		t.Run(ex.name, func(t *testing.T) {
			t.Parallel()

			var cplPaths []string
			if ex.hasCoreCPL {
				cplPaths = append(cplPaths, filepath.Join(ex.parentDir, "core.cpl"))
			}
			cplPaths = append(cplPaths, ex.cplFiles...)

			logger := slog.Default()
			compiler := NewCPLCompiler(cplPaths, "", "", logger, slog.LevelError)
			compiler.compileOnly = true
			// Point standardDefsDir at the absolute path so it resolves regardless of CWD
			moduleRoot := findModuleRootFromTest(t)
			compiler.standardDefsDir = filepath.Join(moduleRoot, "standard_definitions")

			err := compiler.Compile()
			assert.NoError(t, err, "Compile-only failed for %s", ex.name)
		})
	}
}

// TestWithDataExamples_WithErrors walks examples/with_data/with_errors/ and
// verifies that each discovered error example fails compilation with an error
// message containing the substring specified in expected_error.txt.
//
// Error examples are organized using the same two-level directory pattern as
// without_errors: category directories containing leaf example subdirectories.
// Each leaf must contain at least one .cpl file, an iospec.json, and an
// expected_error.txt file. This test does not attempt to run generated programs.
//
// If the with_errors/ directory is empty, the test passes trivially.
func TestWithDataExamples_WithErrors(t *testing.T) {
	withErrorsDir := getWithDataErrorsDir(t)

	entries, err := os.ReadDir(withErrorsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return // Directory doesn't exist — pass trivially
		}
		require.NoError(t, err, "Failed to read with_errors directory")
	}

	moduleRoot := findModuleRootFromTest(t)
	var testedExamples []string

	for _, categoryEntry := range entries {
		if !categoryEntry.IsDir() {
			continue
		}
		categoryDir := filepath.Join(withErrorsDir, categoryEntry.Name())

		// Check for core.cpl in the category directory
		coreCPLPath := filepath.Join(categoryDir, "core.cpl")
		hasCoreCPL := false
		if _, err := os.Stat(coreCPLPath); err == nil {
			hasCoreCPL = true
		}

		// Walk example subdirectories
		subEntries, err := os.ReadDir(categoryDir)
		require.NoError(t, err, "Failed to read category directory %s", categoryDir)

		for _, subEntry := range subEntries {
			if !subEntry.IsDir() {
				continue
			}
			exampleDir := filepath.Join(categoryDir, subEntry.Name())
			exampleName := subEntry.Name()

			t.Run(exampleName, func(t *testing.T) {
				// Verify iospec.json exists
				iospecPath := filepath.Join(exampleDir, "iospec.json")
				if _, err := os.Stat(iospecPath); err != nil {
					t.Fatalf("iospec.json is missing for error example %s", exampleName)
				}

				// Collect .cpl files
				var cplFiles []string
				cplEntries, _ := os.ReadDir(exampleDir)
				for _, cplEntry := range cplEntries {
					if !cplEntry.IsDir() && strings.HasSuffix(cplEntry.Name(), ".cpl") {
						cplFiles = append(cplFiles, filepath.Join(exampleDir, cplEntry.Name()))
					}
				}
				require.NotEmpty(t, cplFiles, "Error example %q has no .cpl files", exampleName)

				// Read expected_error.txt
				expectedErrorPath := filepath.Join(exampleDir, "expected_error.txt")
				expectedErrorData, err := os.ReadFile(expectedErrorPath)
				if err != nil {
					t.Fatalf("expected_error.txt is missing for error example %s", exampleName)
				}
				expectedSubstring := strings.TrimSpace(string(expectedErrorData))

				// Build CPL paths
				var cplPaths []string
				if hasCoreCPL {
					cplPaths = append(cplPaths, coreCPLPath)
				}
				cplPaths = append(cplPaths, cplFiles...)

				// Compile — should fail during synthesis
				outputDir := filepath.Join(moduleRoot, "generated", "cmd", "errors", exampleName)
				require.NoError(t, os.MkdirAll(outputDir, 0755), "Failed to create output directory")
				t.Cleanup(func() { os.RemoveAll(outputDir) })
				outputPath := filepath.Join(outputDir, "main.go")

				logger := slog.Default()
				compiler := NewCPLCompiler(cplPaths, outputPath, iospecPath, logger, slog.LevelError)
				compiler.standardDefsDir = filepath.Join(moduleRoot, "standard_definitions")
				err = compiler.Compile()

				if err == nil {
					t.Fatalf("expected compilation error but got success for example %s", exampleName)
				}

				// Assert error message contains expected substring
				errMsg := err.Error()
				assert.Contains(t, errMsg, expectedSubstring,
					"Compilation error for %s did not contain expected substring from expected_error.txt", exampleName)

			testedExamples = append(testedExamples, exampleName)
			})
		}
	}

	if len(testedExamples) == 0 {
		t.Skip("No error examples found in with_errors/")
	}
}

// TestMultiMetricDependencyChainEndToEnd verifies the complete multi-metric
// dependency chain example end-to-end. It compiles the CPL, runs the generated
// program, and performs deep assertions on:
//   1. Output parquet files exist and contain correct computed values
//   2. Topological ordering (base_revenue before tax before total)
//   3. Dead-code elimination (unreachable_metric is not generated)
func TestMultiMetricDependencyChainEndToEnd(t *testing.T) {
	moduleRoot := findModuleRootFromTest(t)
	exampleDir := filepath.Join(moduleRoot, "examples", "with_data", "without_errors", "multi_metric", "dependency_chain")
	parentDir := filepath.Join(moduleRoot, "examples", "with_data", "without_errors", "multi_metric")

	cplPaths := []string{
		filepath.Join(parentDir, "core.cpl"),
		filepath.Join(exampleDir, "dependency_chain.cpl"),
	}
	iospecPath := filepath.Join(exampleDir, "iospec.json")

	outputDir := filepath.Join(moduleRoot, "generated", "cmd", "dependency_chain_e2e")
	require.NoError(t, os.MkdirAll(outputDir, 0755), "Failed to create output directory")
	defer os.RemoveAll(outputDir)

	outputPath := filepath.Join(outputDir, "main.go")

	// Phase 1: Compile CPL files
	logger := slog.Default()
	compiler := NewCPLCompiler(cplPaths, outputPath, iospecPath, logger, slog.LevelError)
	compiler.standardDefsDir = filepath.Join(moduleRoot, "standard_definitions")
	err := compiler.Compile()
	require.NoError(t, err, "Compilation failed for dependency_chain")

	// Phase 2: Run the generated program
	cmd := exec.Command("go", "run", outputPath)
	cmd.Dir = exampleDir
	runOutput, err := cmd.CombinedOutput()
	require.NoError(t, err, "Generated program failed for dependency_chain\nOutput:\n%s", string(runOutput))

	// Phase 3: Verify output parquet files exist
	baseRevenueParquet := filepath.Join(exampleDir, "base_revenue.parquet")
	totalParquet := filepath.Join(exampleDir, "total.parquet")
	require.FileExists(t, baseRevenueParquet, "base_revenue.parquet should exist")
	require.FileExists(t, totalParquet, "total.parquet should exist")

	// Phase 4: Verify base_revenue values (identity from connector)
	baseRevenueMap := readParquetMetricValues(t, baseRevenueParquet)
	require.Len(t, baseRevenueMap, 4, "base_revenue should have 4 cells")
	assert.InDelta(t, 100.0, baseRevenueMap["Widget|North|Monday"], 1e-9, "base_revenue Widget/North should be 100")
	assert.InDelta(t, 200.0, baseRevenueMap["Widget|South|Monday"], 1e-9, "base_revenue Widget/South should be 200")
	assert.InDelta(t, 300.0, baseRevenueMap["Gadget|North|Monday"], 1e-9, "base_revenue Gadget/North should be 300")
	assert.InDelta(t, 400.0, baseRevenueMap["Gadget|South|Monday"], 1e-9, "base_revenue Gadget/South should be 400")

	// Phase 5: Verify total values (base_revenue + tax = base_revenue + base_revenue/10)
	totalMap := readParquetMetricValues(t, totalParquet)
	require.Len(t, totalMap, 4, "total should have 4 cells")
	assert.InDelta(t, 110.0, totalMap["Widget|North|Monday"], 1e-9, "total Widget/North should be 110")
	assert.InDelta(t, 220.0, totalMap["Widget|South|Monday"], 1e-9, "total Widget/South should be 220")
	assert.InDelta(t, 330.0, totalMap["Gadget|North|Monday"], 1e-9, "total Gadget/North should be 330")
	assert.InDelta(t, 440.0, totalMap["Gadget|South|Monday"], 1e-9, "total Gadget/South should be 440")

	// Phase 6: Verify topological ordering in generated Compute() function
	generatedCode, err := os.ReadFile(outputPath)
	require.NoError(t, err, "Should be able to read generated code")
	generatedStr := string(generatedCode)

	baseIdx := strings.Index(generatedStr, "computebase_revenue(controller, inputs)")
	taxIdx := strings.Index(generatedStr, "computetax(controller, inputs)")
	totalIdx := strings.Index(generatedStr, "computetotal(controller, inputs)")

	require.GreaterOrEqual(t, baseIdx, 0, "computebase_revenue call should exist in generated Compute()")
	require.GreaterOrEqual(t, taxIdx, 0, "computetax call should exist in generated Compute()")
	require.GreaterOrEqual(t, totalIdx, 0, "computetotal call should exist in generated Compute()")

	assert.Less(t, baseIdx, taxIdx, "base_revenue should be computed before tax")
	assert.Less(t, taxIdx, totalIdx, "tax should be computed before total")

	// Phase 7: Verify dead-code elimination — unreachable_metric should NOT be generated
	assert.NotContains(t, generatedStr, "computeunreachable_metric", "Unreachable metric compute function should NOT be generated")
	assert.NotContains(t, generatedStr, "\"unreachable_metric\"", "Unreachable metric should not appear in generated code")
}

// readParquetMetricValues reads a metric output parquet file and returns a map
// from composite key (Product|Region|SingleDay) to float64 Value.
func readParquetMetricValues(t *testing.T, parquetPath string) map[string]float64 {
	t.Helper()
	schema, numRows, err := cplio.ReadParquetFile(parquetPath)
	require.NoError(t, err, "Failed to read parquet %s", parquetPath)

	// Find column indices
	colIdx := make(map[string]int)
	for i := 0; i < schema.NumColumns(); i++ {
		colIdx[schema.Column(i).Name()] = i
	}

	require.Contains(t, colIdx, "Value", "Value column not found")
	require.Contains(t, colIdx, "Product", "Product column not found")
	require.Contains(t, colIdx, "Region", "Region column not found")
	require.Contains(t, colIdx, "SingleDay", "SingleDay column not found")

	// Read all columns
	cols := make(map[string]interface{})
	for name, idx := range colIdx {
		data, err := cplio.ReadParquetColumnData(parquetPath, idx)
		require.NoError(t, err, "Failed to read column %s from %s", name, parquetPath)
		cols[name] = data
	}

	productVals := cols["Product"].([]string)
	regionVals := cols["Region"].([]string)
	singleDayVals := cols["SingleDay"].([]string)
	valueVals := cols["Value"].([]float64)

	result := make(map[string]float64, numRows)
	for i := 0; i < numRows; i++ {
		key := productVals[i] + "|" + regionVals[i] + "|" + singleDayVals[i]
		result[key] = valueVals[i]
	}
	return result
}
