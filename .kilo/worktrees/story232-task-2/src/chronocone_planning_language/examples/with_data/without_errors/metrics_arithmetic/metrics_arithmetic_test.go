// Package metrics_arithmetic_test verifies that all example CPL files under
// examples/with_data/without_errors parse successfully when combined with the shared core.cpl
// infrastructure in each category directory.
//
// Each example subdirectory contains a metric-only .cpl file that relies on
// core.cpl for shared definitions (concepts, types, units, measures,
// connectors). This test concatenates core.cpl with each example file before
// parsing to verify the combined program is valid.
package metrics_arithmetic_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/stretchr/testify/require"
)

// leafExample describes a discovered example directory and its associated
// category base directory (the directory containing core.cpl).
type leafExample struct {
	name        string // directory name (e.g., "simple_addition")
	categoryDir string // absolute path to the category base directory
}

// TestExampleFilesParse verifies that all example CPL files under
// examples/with_data/without_errors can be parsed without errors when combined with the
// appropriate core.cpl.
func TestExampleFilesParse(t *testing.T) {
	t.Parallel()

	withDataDir := getWithDataDir(t)
	examples := discoverLeafExamples(t, withDataDir)
	require.NotEmpty(t, examples, "No example directories discovered under %s", withDataDir)

	for _, ex := range examples {
		ex := ex
		t.Run(ex.name, func(t *testing.T) {
			t.Parallel()

			corePath := filepath.Join(ex.categoryDir, "core.cpl")
			coreData, err := os.ReadFile(corePath)
			require.NoError(t, err, "Failed to read core.cpl for %s", ex.name)

			exampleDir := filepath.Join(ex.categoryDir, ex.name)
			cplPath := filepath.Join(exampleDir, ex.name+".cpl")
			if _, err := os.Stat(cplPath); err != nil {
				// Fall back to any .cpl file in the example directory
				entries, err := os.ReadDir(exampleDir)
				require.NoError(t, err)
				found := false
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".cpl") {
						cplPath = filepath.Join(exampleDir, entry.Name())
						found = true
						break
					}
				}
				require.True(t, found, "No .cpl file found in %s", exampleDir)
			}

			data, err := os.ReadFile(cplPath)
			require.NoError(t, err, "Failed to read %s", cplPath)

			// Concatenate core.cpl + example.cpl to form a complete program.
			// The CPL compiler accepts multiple .cpl files and merges them;
			// concatenation is equivalent for parsing purposes.
			combined := string(coreData) + "\n" + string(data)

			doc, err := parser.Parse(combined)
			require.NoError(t, err, "Failed to parse %s with core.cpl", ex.name)
			require.NotNil(t, doc, "Parsed document should not be nil for %s", ex.name)
		})
	}
}

// TestMetricsArithmeticEndToEnd below performs full end-to-end synthesis
// verification by running compile_and_run.sh for each example subdirectory,
// which compiles with core.cpl, executes the generated program, and
// verifies each <metric>.parquet against expected_output.xlsx.

// TestMetricsArithmeticEndToEnd discovers all example subdirectories under
// examples/with_data/without_errors that contain an iospec.json file. For each discovered
// example, it runs compile_and_run.sh and reports pass/fail as subtests via
// t.Run().
func TestMetricsArithmeticEndToEnd(t *testing.T) {
	withDataDir := getWithDataDir(t)
	examples := discoverLeafExamples(t, withDataDir)
	require.NotEmpty(t, examples, "No example directories discovered under %s", withDataDir)

	for _, ex := range examples {
		ex := ex
		t.Run(ex.name, func(t *testing.T) {
			err := runCompileAndRun(t, ex.categoryDir, ex.name)
			require.NoError(t, err, "compile_and_run.sh failed for %s", ex.name)
		})
	}
}

// discoverLeafExamples recursively walks baseDir and returns all leaf-level
// example directories. A leaf example directory is one that contains an
// iospec.json file, contains at least one .cpl file, and has no immediate
// subdirectories that also contain an iospec.json file.
func discoverLeafExamples(t *testing.T, baseDir string) []leafExample {
	t.Helper()
	var examples []leafExample
	err := filepath.WalkDir(baseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		iospecPath := filepath.Join(path, "iospec.json")
		if _, err := os.Stat(iospecPath); err != nil {
			return nil // not an example directory
		}
		// Check for child example directories
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		hasChildExample := false
		for _, entry := range entries {
			if entry.IsDir() {
				childIospec := filepath.Join(path, entry.Name(), "iospec.json")
				if _, err := os.Stat(childIospec); err == nil {
					hasChildExample = true
					break
				}
			}
		}
		if hasChildExample {
			return nil // not a leaf, let recursion find children
		}
		// Verify there is at least one .cpl file
		hasCPL := false
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".cpl") {
				hasCPL = true
				break
			}
		}
		if !hasCPL {
			return nil
		}
		// Find the category directory (nearest ancestor with core.cpl)
		categoryDir := findCategoryDir(path, baseDir)
		if categoryDir == "" {
			t.Logf("No core.cpl found for example %s, skipping", path)
			return nil
		}
		examples = append(examples, leafExample{
			name:        filepath.Base(path),
			categoryDir: categoryDir,
		})
		return nil
	})
	require.NoError(t, err, "failed to walk directory %s", baseDir)
	slices.SortFunc(examples, func(a, b leafExample) int {
		return strings.Compare(a.name, b.name)
	})
	return examples
}

// findCategoryDir searches upward from dir (stopping at baseDir) for a
// directory containing core.cpl. It returns the absolute path of that directory
// or an empty string if none is found.
func findCategoryDir(dir, baseDir string) string {
	dir = filepath.Clean(dir)
	baseDir = filepath.Clean(baseDir)
	for {
		corePath := filepath.Join(dir, "core.cpl")
		if _, err := os.Stat(corePath); err == nil {
			return dir
		}
		if dir == baseDir {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// runCompileAndRun locates compile_and_run.sh relative to the test file's
// location, executes the shell script with the example directory name as
// argument, and returns an error if the script exits non-zero.
func runCompileAndRun(t *testing.T, categoryDir, exampleDir string) error {
	t.Helper()
	scriptsDir := getScriptsDir(t)
	scriptPath := filepath.Join(scriptsDir, "compile_and_run.sh")

	cmd := exec.Command("bash", scriptPath, exampleDir, categoryDir)
	cplRoot := filepath.Dir(scriptsDir)
	cmd.Dir = cplRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// getScriptsDir uses runtime.Caller(0) to find the test file's path, then
// navigates up from examples/with_data/without_errors/metrics_arithmetic/ to the
// project root and returns the path to the shell_scripts/ directory.
func getScriptsDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to get current file path")
	testDir := filepath.Dir(filename)
	// Walk up 4 levels: metrics_arithmetic -> without_errors -> with_data -> examples -> chronocone_planning_language
	cplRoot := filepath.Join(testDir, "..", "..", "..", "..")
	cplRoot, err := filepath.Abs(cplRoot)
	require.NoError(t, err, "failed to resolve absolute path")
	return filepath.Join(cplRoot, "shell_scripts")
}

// getWithDataDir returns the absolute path to the examples/with_data directory.
func getWithDataDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to get current file path")
	testDir := filepath.Dir(filename)
	withDataDir := filepath.Join(testDir, "..")
	withDataDir, err := filepath.Abs(withDataDir)
	require.NoError(t, err, "failed to resolve with_data path")
	return withDataDir
}
