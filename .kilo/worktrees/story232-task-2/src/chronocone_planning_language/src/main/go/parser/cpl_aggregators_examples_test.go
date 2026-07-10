package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAggregatorsExamples_Valid(t *testing.T) {
	baseDir := filepath.Join("..", "..", "..", "..", "examples", "parse_and_analyze_only", "without_errors")
	validExamples := []string{
		"aggregators-single-concept.cpl",
		"aggregators-multi-concept.cpl",
		"aggregators-scalar-reduction.cpl",
		"aggregators-expression-arithmetic.cpl",
		"aggregators-zero-declarations.cpl",
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

func TestAggregatorsExamples_Errors(t *testing.T) {
	baseDir := filepath.Join("..", "..", "..", "..", "examples", "parse_and_analyze_only", "with_errors")
	errorExamples := []string{
		"aggregators-missing-from.cpl",
		"aggregators-missing-to.cpl",
		"aggregators-empty-from-list.cpl",
		"aggregators-malformed-syntax.cpl",
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
