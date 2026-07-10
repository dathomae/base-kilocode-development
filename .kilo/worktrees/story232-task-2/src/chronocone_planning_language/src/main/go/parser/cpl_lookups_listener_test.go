package parser

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// findLookupsScope traverses the symbol table tree to find the LookupsScope child.
func findLookupsScope(root *symboltable.SymbolTable) *symboltable.SymbolTable {
	for _, child := range root.Children {
		if child.ScopeType == symboltable.LookupsScope {
			return child
		}
	}
	return nil
}

func TestLookupsListener_SingleLookup(t *testing.T) {
	input := `lookups
    SalesLookup
        dimensions[Store]
        celltype RevenueMeasure
`
	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	lookupsScope := findLookupsScope(rootTable)
	if lookupsScope == nil {
		t.Fatal("expected LookupsScope child of root table")
	}

	if len(lookupsScope.Definitions) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(lookupsScope.Definitions))
	}

	def := lookupsScope.Definitions[0]
	if def.DefinitionType != symboltable.LookupScopeDefinition {
		t.Errorf("expected LookupScopeDefinition, got %s", def.DefinitionType)
	}
	if def.Name == nil || *def.Name != "SalesLookup" {
		t.Errorf("expected name 'SalesLookup', got %v", def.Name)
	}

	pld, ok := def.Data.(*ParserLookupDefinition)
	if !ok {
		t.Fatalf("expected *ParserLookupDefinition in Data, got %T", def.Data)
	}
	if len(pld.Dimensions) != 1 || pld.Dimensions[0] != "Store" {
		t.Errorf("expected dimensions [Store], got %v", pld.Dimensions)
	}
	if pld.Celltype != "RevenueMeasure" {
		t.Errorf("expected celltype 'RevenueMeasure', got '%s'", pld.Celltype)
	}
}

func TestLookupsListener_MultipleLookups(t *testing.T) {
	input := `lookups
    SalesLookup
        dimensions[Store]
        celltype RevenueMeasure
    CostLookup
        dimensions[Product, Region]
        celltype CostMeasure
`
	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	lookupsScope := findLookupsScope(rootTable)
	if lookupsScope == nil {
		t.Fatal("expected LookupsScope child of root table")
	}

	if len(lookupsScope.Definitions) != 2 {
		t.Fatalf("expected 2 definitions, got %d", len(lookupsScope.Definitions))
	}

	// First definition: SalesLookup
	def1 := lookupsScope.Definitions[0]
	if def1.Name == nil || *def1.Name != "SalesLookup" {
		t.Errorf("expected first definition name 'SalesLookup', got %v", def1.Name)
	}
	pld1, ok := def1.Data.(*ParserLookupDefinition)
	if !ok {
		t.Fatalf("expected *ParserLookupDefinition in Data, got %T", def1.Data)
	}
	if len(pld1.Dimensions) != 1 || pld1.Dimensions[0] != "Store" {
		t.Errorf("expected dimensions [Store], got %v", pld1.Dimensions)
	}

	// Second definition: CostLookup
	def2 := lookupsScope.Definitions[1]
	if def2.Name == nil || *def2.Name != "CostLookup" {
		t.Errorf("expected second definition name 'CostLookup', got %v", def2.Name)
	}
	pld2, ok := def2.Data.(*ParserLookupDefinition)
	if !ok {
		t.Fatalf("expected *ParserLookupDefinition in Data, got %T", def2.Data)
	}
	if len(pld2.Dimensions) != 2 || pld2.Dimensions[0] != "Product" || pld2.Dimensions[1] != "Region" {
		t.Errorf("expected dimensions [Product, Region], got %v", pld2.Dimensions)
	}
	if pld2.Celltype != "CostMeasure" {
		t.Errorf("expected celltype 'CostMeasure', got '%s'", pld2.Celltype)
	}
}

func TestLookupsListener_EmptySection(t *testing.T) {
	input := `lookups
`
	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	lookupsScope := findLookupsScope(rootTable)
	if lookupsScope == nil {
		t.Fatal("expected LookupsScope child of root table even for empty section")
	}

	if len(lookupsScope.Definitions) != 0 {
		t.Errorf("expected 0 definitions for empty section, got %d", len(lookupsScope.Definitions))
	}
}
