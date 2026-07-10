package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSpreadersExamples_Valid(t *testing.T) {
	baseDir := filepath.Join("..", "..", "..", "..", "examples", "parse_and_analyze_only", "without_errors")
	validExamples := []string{
		"spreaders-simple.cpl",
		"spreaders-scalar.cpl",
		"spreaders-arithmetic-expression.cpl",
		"spreaders-zero-declarations.cpl",
	}
	for _, example := range validExamples {
		t.Run(example, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(baseDir, example))
			if err != nil {
				t.Fatalf("failed to read example file %s: %v", example, err)
			}
			_, err = Parse(string(content))
			if err != nil {
				t.Fatalf("expected %s to parse successfully, got error: %v", example, err)
			}
		})
	}
}

func TestSpreadersExamples_Errors(t *testing.T) {
	baseDir := filepath.Join("..", "..", "..", "..", "examples", "parse_and_analyze_only", "with_errors")
	errorExamples := []string{
		"spreaders-missing-from.cpl",
		"spreaders-missing-to.cpl",
		"spreaders-malformed-syntax.cpl",
		"spreaders-invalid-expression.cpl",
	}
	for _, example := range errorExamples {
		t.Run(example, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(baseDir, example))
			if err != nil {
				t.Fatalf("failed to read example file %s: %v", example, err)
			}
			_, err = Parse(string(content))
			if err == nil {
				t.Fatalf("expected %s to produce a parse error, but it succeeded", example)
			}
		})
	}
}
