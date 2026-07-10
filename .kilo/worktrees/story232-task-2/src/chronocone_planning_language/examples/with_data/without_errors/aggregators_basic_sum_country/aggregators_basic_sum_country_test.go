// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

// Package aggregators_basic_sum_country_test verifies that the example CPL files
// under examples/with_data/without_errors/aggregators_basic_sum_country parse successfully
// when combined, and verifies direct tensorops behavior with the example parquet data.
package aggregators_basic_sum_country

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/calculation"
	cplio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/tensorops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExampleFilesParse verifies that core.cpl and sum_country.cpl can be
// parsed without errors when combined.
func TestExampleFilesParse(t *testing.T) {
	t.Parallel()

	exampleDir := getExampleDir(t)

	corePath := filepath.Join(exampleDir, "core.cpl")
	coreData, err := os.ReadFile(corePath)
	require.NoError(t, err, "Failed to read core.cpl")

	cplPath := filepath.Join(exampleDir, "sum_country.cpl")
	data, err := os.ReadFile(cplPath)
	require.NoError(t, err, "Failed to read sum_country.cpl")

	combined := string(coreData) + "\n" + string(data)

	doc, err := parser.Parse(combined)
	require.NoError(t, err, "Failed to parse combined CPL")
	require.NotNil(t, doc, "Parsed document should not be nil")
}

// TestAggregatorsBasicEndToEnd compiles the CPL files, verifies the generated
// Go source exists, builds the generated program, and checks that the
// aggregation matrix is declared in the generated source.
//
// NOTE: This increment emits the aggregation matrix but does not wire it into
// compute functions, so the test stops short of running the generated program.
func TestAggregatorsBasicEndToEnd(t *testing.T) {
	exampleDir := getExampleDir(t)
	err := compileAndVerify(t, exampleDir)
	require.NoError(t, err, "compileAndVerify failed")
}

// TestReduceDimension_SumCountry reads the parquet data and performs a direct
// tensorops ReduceDimension test, verifying that summing across the Country
// dimension produces the expected aggregated values.
func TestReduceDimension_SumCountry(t *testing.T) {
	exampleDir := getExampleDir(t)
	parquetPath := filepath.Join(exampleDir, "data_source.parquet")

	// Read parquet schema to discover column layout
	schema, numRows, err := cplio.ReadParquetFile(parquetPath)
	require.NoError(t, err, "Failed to read parquet file")
	require.Equal(t, 4, numRows, "Expected 4 rows in parquet")

	// Find column indices
	var valueColIdx int = -1
	for i := 0; i < schema.NumColumns(); i++ {
		if schema.Column(i).Name() == "Value" {
			valueColIdx = i
			break
		}
	}
	require.GreaterOrEqual(t, valueColIdx, 0, "Value column not found in parquet")

	// Read Value column
	rawValues, err := cplio.ReadParquetColumnData(parquetPath, valueColIdx)
	require.NoError(t, err, "Failed to read Value column")
	values, ok := rawValues.([]float64)
	require.True(t, ok, "Value column is not float64")
	require.Equal(t, 4, len(values), "Expected 4 values")

	// Construct a 2D tensor with dimensions [Country, Product].
	// Parquet row order is: USA/widget, USA/gadget, Canada/widget, Canada/gadget.
	// In row-major order with shape [2, 2], the data maps directly:
	//   [0,0]=USA/widget, [0,1]=USA/gadget, [1,0]=Canada/widget, [1,1]=Canada/gadget
	src := calculation.NewChronoconeTensor(
		values,
		[]string{"Country", "Product"},
		[]int{2, 2},
	)

	// Sum aggregation matrix: 2 rows (USA, Canada) × 1 column (sum).
	matrix := &tensorops.AggregationMatrix{
		Rows:    2,
		Columns: 1,
		Data:    [][]float64{{1.0}, {1.0}},
	}

	// Collapse dimension 0 (Country) and place the result at targetDim=1.
	result, err := tensorops.ReduceDimension(src, matrix, 0, 1)
	require.NoError(t, err, "ReduceDimension failed")
	require.NotNil(t, result, "Result should not be nil")

	// Expected result shape: [2, 1] (Product dimension preserved, Country collapsed to size 1)
	assert.Equal(t, []int{2, 1}, result.Shape(), "Result shape mismatch")
	assert.Equal(t, []string{"Product", "Country"}, result.DimensionNames(), "Result dimension names mismatch")

	// Expected values:
	//   widget = 100 + 300 = 400
	//   gadget = 200 + 400 = 600
	assert.InDeltaSlice(t, []float64{400.0, 600.0}, result.Data(), 1e-9, "Result data mismatch")
}

// compileAndVerify compiles the CPL files in exampleDir, verifies the generated
// Go source exists, builds it, and checks that the source contains an
// AggregationMatrix declaration. It does not run the generated program because
// this increment only emits the aggregation matrix without wiring it into the
// compute functions.
func compileAndVerify(t *testing.T, exampleDir string) error {
	t.Helper()

	projectRoot := getProjectRoot(t)
	iospecPath := filepath.Join(exampleDir, "iospec.json")

	// Collect CPL files in the example directory
	var cplFiles []string
	entries, err := os.ReadDir(exampleDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".cpl") {
			cplFiles = append(cplFiles, filepath.Join(exampleDir, entry.Name()))
		}
	}
	if len(cplFiles) == 0 {
		return os.ErrNotExist
	}

	// Generate into a temp directory so we don't pollute the workspace
	genDir := t.TempDir()
	outputFile := filepath.Join(genDir, "main.go")

	compilerArgs := []string{
		"run",
		filepath.Join("src", "main", "go", "cli", "compiler", "cpl_compiler.go"),
		"-io", iospecPath,
		"-o", outputFile,
	}
	compilerArgs = append(compilerArgs, cplFiles...)

	cmd := exec.Command("go", compilerArgs...)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// 1. Verify the generated Go file exists
	if _, err := os.Stat(outputFile); err != nil {
		return err
	}
	info, err := os.Stat(outputFile)
	if err != nil {
		return err
	}
	if info.Size() == 0 {
		return os.ErrInvalid
	}

	// 2. Verify go build succeeds on the generated program
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(genDir, "main"), outputFile)
	buildCmd.Dir = projectRoot
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return err
	}

	// 3. Verify the generated source contains an AggregationMatrix declaration
	src, err := os.ReadFile(outputFile)
	if err != nil {
		return err
	}
	if !strings.Contains(string(src), "AggregationMatrix") {
		return os.ErrInvalid
	}

	return nil
}

// getScriptsDir uses runtime.Caller(0) to find the test file's path, then
// navigates up from examples/with_data/without_errors/aggregators_basic_sum_country/ to the
// project root and returns the path to the shell_scripts/ directory.
func getScriptsDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to get current file path")
	testDir := filepath.Dir(filename)
	// Walk up 4 levels: aggregators_basic_sum_country -> without_errors -> with_data -> examples -> chronocone_planning_language
	cplRoot := filepath.Join(testDir, "..", "..", "..", "..")
	cplRoot, err := filepath.Abs(cplRoot)
	require.NoError(t, err, "failed to resolve absolute path")
	return filepath.Join(cplRoot, "shell_scripts")
}

// getProjectRoot returns the absolute path to the project root directory.
func getProjectRoot(t *testing.T) string {
	t.Helper()
	scriptsDir := getScriptsDir(t)
	projectRoot := filepath.Dir(scriptsDir)
	return projectRoot
}

// getExampleDir returns the absolute path to the example directory.
func getExampleDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to get current file path")
	exampleDir, err := filepath.Abs(filepath.Dir(filename))
	require.NoError(t, err, "failed to resolve example dir path")
	return exampleDir
}
