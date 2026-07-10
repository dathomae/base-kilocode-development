package parser

import (
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

func TestLookupsIntegration_WithConcepts(t *testing.T) {
	// Test that lookups can coexist with other sections
	input := `concepts
    Store=[Values[NY, CA]]
measures
    RevenueMeasure = Number(unit=Currency)

lookups
    SalesLookup
        dimensions[Store]
        celltype RevenueMeasure
`
	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("expected parse to succeed, got error: %v", err)
	}

	// Verify concepts section
	var foundConcepts bool
	for _, child := range rootTable.Children {
		if child.ScopeType == symboltable.ConceptsScope {
			foundConcepts = true
		}
	}
	if !foundConcepts {
		t.Error("expected ConceptsScope child of root table")
	}

	// Verify lookups section
	lookupsScope := findLookupsScope(rootTable)
	if lookupsScope == nil {
		t.Fatal("expected LookupsScope child of root table")
	}
}

func TestLookupsIntegration_MissingDimensions(t *testing.T) {
	input := `lookups
    BadLookup
        celltype SomeMeasure
`
	_, err := Parse(input)
	if err == nil {
		t.Fatal("expected parse error for missing dimensions clause")
	}
	if !strings.Contains(err.Error(), "dimensions") {
		t.Errorf("expected error to mention 'dimensions', got: %v", err)
	}
}

func TestLookupsIntegration_MissingCelltype(t *testing.T) {
	input := `lookups
    BadLookup
        dimensions[Store]
`
	_, err := Parse(input)
	if err == nil {
		t.Fatal("expected parse error for missing celltype clause")
	}
	if !strings.Contains(err.Error(), "celltype") {
		t.Errorf("expected error to mention 'celltype', got: %v", err)
	}
}

func TestLookupsIntegration_EmptyDimensions(t *testing.T) {
	input := `lookups
    BadLookup
        dimensions[]
        celltype SomeMeasure
`
	_, err := Parse(input)
	if err == nil {
		t.Fatal("expected parse error for empty dimension list")
	}
	if !strings.Contains(err.Error(), "IDENTIFIER") {
		t.Errorf("expected error to mention missing identifier, got: %v", err)
	}
}

func TestLookupsIntegration_MalformedSyntax(t *testing.T) {
	input := `lookups
    BadLookup
        dimensions[Store
        celltype SomeMeasure
`
	_, err := Parse(input)
	if err == nil {
		t.Fatal("expected parse error for malformed syntax (missing bracket)")
	}
	if !strings.Contains(err.Error(), "']'") && !strings.Contains(err.Error(), "','") {
		t.Errorf("expected error to mention missing bracket or comma, got: %v", err)
	}
}
