package analysis_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Spreader Error Example Files Integration Tests
// ============================================================================

// TestAnalyze_SpreaderWithErrors_Examples verifies that spreader error example
// files in examples/parse_and_analyze_only/with_errors/ produce expected
// analysis errors. These files serve as regression tests for both the parse-level
// and analysis-level spreader handling.
//
// Note: Some examples (e.g., spreaders-missing-to, spreaders-malformed-syntax)
// produce parse errors rather than analysis errors. The test still validates
// them because getAnalysisForFile returns an error during parsing, which
// satisfies the require.Error assertion.
func TestAnalyze_SpreaderWithErrors_Examples(t *testing.T) {
	t.Parallel()

	examplesDir := getExamplesDir(t)
	withErrorsDir := filepath.Join(examplesDir, "parse_and_analyze_only", "with_errors")

	entries, err := os.ReadDir(withErrorsDir)
	require.NoError(t, err, "Failed to read with_errors directory")

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cpl") {
			continue
		}
		// Only test spreader-specific error examples
		if !strings.HasPrefix(entry.Name(), "spreaders-") {
			continue
		}

		entry := entry // capture range variable for closure
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(withErrorsDir, entry.Name())
			_, err := getAnalysisForFile(t, filePath)

			require.Error(t, err, "Expected analysis error for %s", entry.Name())

			switch entry.Name() {
			case "spreaders-dimensional-mismatch.cpl":
				assert.Contains(t, err.Error(), "dimensional unit mismatch")
			case "spreaders-missing-concept.cpl":
				assert.Contains(t, err.Error(), "unknown concept")
				assert.Contains(t, err.Error(), "Product")
			case "spreaders-duplicate-pattern.cpl":
				assert.Contains(t, err.Error(), "duplicate spreader definition")
			}
		})
	}
}
