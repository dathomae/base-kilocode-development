package parser

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// requireIdentifierDef is a helper that asserts the i-th definition is an
// Identifier with the expected name. It is shared across multiple parser test
// files (e.g. aggregators and spreaders listener tests).
func requireIdentifierDef(t *testing.T, defs []symboltable.SymbolDefinition, i int, expectedName string) {
	t.Helper()

	if i >= len(defs) {
		t.Fatalf("expected at least %d definitions, got %d", i+1, len(defs))
	}
	def := defs[i]
	if def.DefinitionType != symboltable.Identifier {
		t.Fatalf("expected definition %d to be Identifier, got %s", i, def.DefinitionType)
	}
	strData, ok := def.Data.(symboltable.StringData)
	if !ok {
		t.Fatalf("expected definition %d Data to be StringData, got %T", i, def.Data)
	}
	if strData.Value != expectedName {
		t.Fatalf("expected definition %d identifier name %q, got %q", i, expectedName, strData.Value)
	}
}
