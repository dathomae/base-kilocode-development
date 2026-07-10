// Package main contains the CPL compiler.
package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// extractLookupNames parses a CPL file content and returns the names of all
// lookups declared in the lookups section.
func extractLookupNames(content string) []string {
	var names []string
	lines := strings.Split(content, "\n")
	inLookups := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "lookups" {
			inLookups = true
			continue
		}
		if !inLookups {
			continue
		}
		if line == "" {
			continue
		}
		// Not indented means we've left the lookups section.
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			break
		}
		// Skip property lines within a lookup declaration.
		if strings.HasPrefix(trimmed, "dimensions") || strings.HasPrefix(trimmed, "celltype") {
			continue
		}
		if trimmed != "" {
			names = append(names, trimmed)
		}
	}
	return names
}

// buildIOSpecWithLookups returns a JSON IO spec string that includes the
// given lookup names so synthesis does not fail on missing lookup data.
func buildIOSpecWithLookups(lookupNames []string) string {
	type driverParam struct {
		Path string `json:"path"`
	}
	type lookupEntry struct {
		LookupName       string        `json:"lookup"`
		DataFormat       string        `json:"data_format"`
		Driver           string        `json:"driver"`
		DriverParameters []driverParam `json:"driver_parameters"`
	}
	type spec struct {
		Connectors   []interface{} `json:"connectors"`
		Lookups      []lookupEntry `json:"lookups"`
		Outputs      []interface{} `json:"outputs"`
		Architecture struct {
			CPUCores       int `json:"cpu_cores"`
			CacheSize      int `json:"cache_size"`
			MainMemorySize int `json:"main_memory_size"`
		} `json:"architecture"`
	}
	s := spec{
		Connectors: []interface{}{},
		Outputs:    []interface{}{},
	}
	s.Architecture.CPUCores = 1
	s.Architecture.CacheSize = 1024
	s.Architecture.MainMemorySize = 1024
	for _, name := range lookupNames {
		s.Lookups = append(s.Lookups, lookupEntry{
			LookupName: name,
			DataFormat: "parquet",
			Driver:     "file",
			DriverParameters: []driverParam{
				{Path: "dummy.parquet"},
			},
		})
	}
	b, _ := json.Marshal(s)
	return string(b)
}

// getExamplesDir returns the path to the examples directory.
func getExamplesDir(t *testing.T) string {
	t.Helper()
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	// From src/main/go/cli/compiler go up 5 levels to reach the module root (where go.mod is), then into examples
	projectRoot := filepath.Join(testDir, "..", "..", "..", "..", "..")
	return filepath.Join(projectRoot, "examples")
}

// getStandardDefsDir returns the path to the standard_definitions directory.
func getStandardDefsDir(t *testing.T) string {
	t.Helper()
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	// From src/main/go/cli/compiler go up 5 levels to reach the module root, then into standard_definitions
	projectRoot := filepath.Join(testDir, "..", "..", "..", "..", "..")
	return filepath.Join(projectRoot, "standard_definitions")
}

// TestExamplesWithoutErrors tests that all files in examples/without_errors compile successfully.
func TestExamplesWithoutErrors(t *testing.T) {
	t.Parallel()

	examplesDir := getExamplesDir(t)
	withoutErrorsDir := filepath.Join(examplesDir, "parse_and_analyze_only", "without_errors")

	entries, err := os.ReadDir(withoutErrorsDir)
	require.NoError(t, err, "Failed to read without_errors directory")

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cpl") {
			continue
		}

		entry := entry // capture range variable
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(withoutErrorsDir, entry.Name())

			tempDir := t.TempDir()
			outputPath := filepath.Join(tempDir, "output.go")

			// Create a minimal IO spec for synthesis, embedding lookup data when
			// the CPL file declares lookups.
			cplData, err := os.ReadFile(filePath)
			require.NoError(t, err, "Failed to read CPL file")
			lookupNames := extractLookupNames(string(cplData))
			ioSpecContent := buildIOSpecWithLookups(lookupNames)
			ioSpecPath := filepath.Join(tempDir, "io_spec.json")
			require.NoError(t, os.WriteFile(ioSpecPath, []byte(ioSpecContent), 0644), "Failed to write IO spec")

			logger := slog.Default()
			cplCompiler := NewCPLCompiler([]string{filePath}, outputPath, ioSpecPath, logger, slog.LevelError)

			// Files that use Instant/Interval constraints require standard time
			// concepts (Hour, Minute, Second, DayOfWeek, Month, etc.) to be
			// defined. We create a temporary standard definitions directory
			// containing only standard_concepts.cpl to avoid conflicts with
			// standard_types.cpl and standard_measures.cpl which define types
			// and measures that duplicate or are incompatible with these examples.
			if entry.Name() == "instant_declarations.cpl" || entry.Name() == "interval_declarations.cpl" {
				stdDefsDir := filepath.Join(t.TempDir(), "standard_defs")
				require.NoError(t, os.MkdirAll(stdDefsDir, 0755))
				srcPath := filepath.Join(getStandardDefsDir(t), "standard_concepts.cpl")
				data, err := os.ReadFile(srcPath)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(filepath.Join(stdDefsDir, "standard_concepts.cpl"), data, 0644))
				cplCompiler.standardDefsDir = stdDefsDir
			}

			err = cplCompiler.Compile()

			// aggregators-single-concept.cpl has from [Country] to [Country] —
			// all from-concepts equal the to-concept, so there is no collapsible
			// dimension. This is caught at analysis time.
			if entry.Name() == "aggregators-single-concept.cpl" {
				assert.Error(t, err, "Expected analysis error for %s (no collapsible dimension)", entry.Name())
				return
			}

			assert.NoError(t, err, "Expected no error for %s", entry.Name())
		})
	}
}

// TestExamplesWithErrors tests that all files in examples/with_errors produce compile errors.
func TestExamplesWithErrors(t *testing.T) {
	t.Parallel()

	examplesDir := getExamplesDir(t)
	withErrorsDir := filepath.Join(examplesDir, "parse_and_analyze_only", "with_errors")

	entries, err := os.ReadDir(withErrorsDir)
	require.NoError(t, err, "Failed to read with_errors directory")

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cpl") {
			continue
		}

		entry := entry // capture range variable
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(withErrorsDir, entry.Name())
			tempDir := t.TempDir()
			outputPath := filepath.Join(tempDir, "output.go")

			// Create a minimal IO spec for synthesis
			ioSpecPath := filepath.Join(tempDir, "io_spec.json")
			ioSpecContent := `{"connectors": [], "outputs": [], "architecture": {"cpu_cores": 1, "cache_size": 1024, "main_memory_size": 1024}}`
			require.NoError(t, os.WriteFile(ioSpecPath, []byte(ioSpecContent), 0644), "Failed to write IO spec")

			logger := slog.Default()
			cplCompiler := NewCPLCompiler([]string{filePath}, outputPath, ioSpecPath, logger, slog.LevelError)

			err := cplCompiler.Compile()
			assert.Error(t, err, "Expected compile error for %s", entry.Name())
		})
	}
}
