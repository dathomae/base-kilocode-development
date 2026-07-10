package parser

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

func TestParserLookupDefinition_Init(t *testing.T) {
	loc := symboltable.SourceLocation{File: "test.cpl", Line: 10, Column: 4}
	pld := ParserLookupDefinition{
		Name:           "TestLookup",
		Dimensions:     []string{"Store", "Product"},
		Celltype:       "RevenueMeasure",
		SourceLocation: loc,
	}

	if pld.Name != "TestLookup" {
		t.Errorf("expected Name 'TestLookup', got '%s'", pld.Name)
	}
	if len(pld.Dimensions) != 2 {
		t.Errorf("expected 2 dimensions, got %d", len(pld.Dimensions))
	}
	if pld.Dimensions[0] != "Store" {
		t.Errorf("expected first dimension 'Store', got '%s'", pld.Dimensions[0])
	}
	if pld.Dimensions[1] != "Product" {
		t.Errorf("expected second dimension 'Product', got '%s'", pld.Dimensions[1])
	}
	if pld.Celltype != "RevenueMeasure" {
		t.Errorf("expected Celltype 'RevenueMeasure', got '%s'", pld.Celltype)
	}
	if pld.SourceLocation.Line != 10 {
		t.Errorf("expected Line 10, got %d", pld.SourceLocation.Line)
	}
	if pld.SourceLocation.Column != 4 {
		t.Errorf("expected Column 4, got %d", pld.SourceLocation.Column)
	}
}

func TestParserLookupDefinition_SourceLocation(t *testing.T) {
	loc := symboltable.SourceLocation{File: "example.cpl", Line: 42, Column: 8}
	pld := ParserLookupDefinition{
		Name:           "SalesLookup",
		Dimensions:     []string{"Region"},
		Celltype:       "SalesData",
		SourceLocation: loc,
	}

	if pld.SourceLocation.File != "example.cpl" {
		t.Errorf("expected File 'example.cpl', got '%s'", pld.SourceLocation.File)
	}
	if pld.SourceLocation.Line != 42 {
		t.Errorf("expected Line 42, got %d", pld.SourceLocation.Line)
	}
	if pld.SourceLocation.Column != 8 {
		t.Errorf("expected Column 8, got %d", pld.SourceLocation.Column)
	}
}
