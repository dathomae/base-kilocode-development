package metrics_conditionals_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/stretchr/testify/require"
)

// TestExampleFilesParse verifies that all conditional example CPL files under
// examples/with_data/without_errors/metrics_conditionals can be parsed without errors when
// combined with core.cpl.
func TestExampleFilesParse(t *testing.T) {
	t.Parallel()

	withDataDir := getWithDataDir(t)
	conditionalsDir := filepath.Join(withDataDir, "metrics_conditionals")
	corePath := filepath.Join(conditionalsDir, "core.cpl")

	coreData, err := os.ReadFile(corePath)
	require.NoError(t, err, "Failed to read core.cpl")

	examples := []string{
		"conditional_if_else_arithmetic",
		"conditional_if_elseif_else",
		"conditional_when_default",
		"conditional_comparison_operators",
		"conditional_logical_operators",
		"conditional_unary_negation",
		"conditional_short_form",
		"conditional_truthiness",
		"conditional_tautological_if_only",
		"conditional_nested_arithmetic",
	}

	for _, ex := range examples {
		ex := ex
		t.Run(ex, func(t *testing.T) {
			t.Parallel()

			cplPath := filepath.Join(conditionalsDir, ex, ex+".cpl")
			data, err := os.ReadFile(cplPath)
			require.NoError(t, err, "Failed to read %s", cplPath)

			combined := string(coreData) + "\n" + string(data)
			doc, err := parser.Parse(combined)
			require.NoError(t, err, "Failed to parse %s with core.cpl", ex)
			require.NotNil(t, doc, "Parsed document should not be nil for %s", ex)
		})
	}
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
