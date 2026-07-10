package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLookupsExamples_Valid(t *testing.T) {
	baseDir := filepath.Join("..", "..", "..", "..", "examples", "parse_and_analyze_only", "without_errors")
	validExamples := []string{
		"lookups_single_dimension.cpl",
		"lookups_multiple_dimensions.cpl",
		"lookups_multiple_declarations.cpl",
		"lookups_empty_section.cpl",
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

func TestLookupsExamples_Errors(t *testing.T) {
	baseDir := filepath.Join("..", "..", "..", "..", "examples", "parse_and_analyze_only", "with_errors")
	errorExamples := []string{
		"lookups_missing_dimensions.cpl",
		"lookups_missing_celltype.cpl",
		"lookups_empty_dimension_list.cpl",
		"lookups_malformed_syntax.cpl",
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
