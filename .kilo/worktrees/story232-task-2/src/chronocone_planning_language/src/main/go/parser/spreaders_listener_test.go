package parser

import (
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// findSpreadersScope traverses the symbol table tree to find the SpreadersScope child.
func findSpreadersScope(root *symboltable.SymbolTable) *symboltable.SymbolTable {
	for _, child := range root.Children {
		if child.ScopeType == symboltable.SpreadersScope {
			return child
		}
	}
	return nil
}

// requireSpreaderDefinition is a helper that extracts a ParserSpreaderDefinition from a SpreaderScope.
func requireSpreaderDefinition(t *testing.T, spreaderScope *symboltable.SymbolTable) *ParserSpreaderDefinition {
	t.Helper()

	if spreaderScope.Data == nil {
		t.Fatal("SpreaderScope.Data should be populated")
	}
	def, ok := spreaderScope.Data.(*ParserSpreaderDefinition)
	if !ok {
		t.Fatalf("SpreaderScope.Data should be a *ParserSpreaderDefinition, got %T", spreaderScope.Data)
	}
	if def == nil {
		t.Fatal("SpreaderScope.Data should not contain a nil ParserSpreaderDefinition pointer")
	}
	return def
}

// requireSpreaderScopeDefinition verifies that the SpreaderScope's Definitions
// contains a SpreaderScopeDefinition entry with the expected synthetic key and
// the same *ParserSpreaderDefinition pointer as scope.Data.
func requireSpreaderScopeDefinition(t *testing.T, spreaderScope *symboltable.SymbolTable, expectedDef *ParserSpreaderDefinition) {
	t.Helper()

	count := 0
	for _, def := range spreaderScope.Definitions {
		if def.DefinitionType == symboltable.SpreaderScopeDefinition {
			count++
			if def.Name == nil || *def.Name != expectedDef.Key {
				t.Errorf("expected SpreaderScopeDefinition name %q, got %v", expectedDef.Key, def.Name)
			}
			psd, ok := def.Data.(*ParserSpreaderDefinition)
			if !ok {
				t.Fatalf("expected SpreaderScopeDefinition Data to be *ParserSpreaderDefinition, got %T", def.Data)
			}
			if psd != expectedDef {
				t.Fatalf("expected SpreaderScopeDefinition Data to be same pointer as scope.Data")
			}
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 SpreaderScopeDefinition in Definitions, got %d", count)
	}
}

func TestSpreadersListener_SingleConceptFromMultiTo(t *testing.T) {
	input := `concepts
	Country=[Values[USA]]
	Product=[Values[A, B]]

spreaders
	from [Product] to [Country, Product] = Demand`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	spreadersScope := findSpreadersScope(rootTable)
	if spreadersScope == nil {
		t.Fatal("expected SpreadersScope child of root table")
	}

	if len(spreadersScope.Children) != 1 {
		t.Fatalf("expected 1 SpreaderScope child, got %d", len(spreadersScope.Children))
	}

	spreaderScope := spreadersScope.Children[0]
	if spreaderScope.ScopeType != symboltable.SpreaderScope {
		t.Fatalf("expected SpreaderScope, got %s", spreaderScope.ScopeType)
	}

	// Verify definition data
	def := requireSpreaderDefinition(t, spreaderScope)
	if len(def.FromConcepts) != 1 || def.FromConcepts[0] != "Product" {
		t.Errorf("expected FromConcepts [Product], got %v", def.FromConcepts)
	}
	if len(def.ToConcepts) != 2 || def.ToConcepts[0] != "Country" || def.ToConcepts[1] != "Product" {
		t.Errorf("expected ToConcepts [Country, Product], got %v", def.ToConcepts)
	}
	if def.FromEmpty {
		t.Error("expected FromEmpty false")
	}

	// Verify synthetic key
	expectedKey := "__spreader__[Product]->[Country,Product]"
	if def.Key != expectedKey {
		t.Errorf("expected synthetic key %q, got %q", expectedKey, def.Key)
	}

	requireSpreaderScopeDefinition(t, spreaderScope, def)

	// Verify postfix definitions: Demand
	if len(spreaderScope.Definitions) != 2 {
		t.Fatalf("expected 2 definitions (1 SpreaderScopeDefinition + 1 postfix), got %d", len(spreaderScope.Definitions))
	}
	requireIdentifierDef(t, spreaderScope.Definitions, 0, "Demand")
}

func TestSpreadersListener_ScalarSpreading(t *testing.T) {
	input := `concepts
	Country=[Values[USA]]

spreaders
	from [] to [Country] = Value`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	spreadersScope := findSpreadersScope(rootTable)
	if spreadersScope == nil {
		t.Fatal("expected SpreadersScope child of root table")
	}

	if len(spreadersScope.Children) != 1 {
		t.Fatalf("expected 1 SpreaderScope child, got %d", len(spreadersScope.Children))
	}

	spreaderScope := spreadersScope.Children[0]
	def := requireSpreaderDefinition(t, spreaderScope)

	if def.FromConcepts != nil {
		t.Errorf("expected FromConcepts nil, got %v", def.FromConcepts)
	}
	if len(def.ToConcepts) != 1 || def.ToConcepts[0] != "Country" {
		t.Errorf("expected ToConcepts [Country], got %v", def.ToConcepts)
	}
	if !def.FromEmpty {
		t.Error("expected FromEmpty true")
	}

	expectedKey := "__spreader__[]->[Country]"
	if def.Key != expectedKey {
		t.Errorf("expected synthetic key %q, got %q", expectedKey, def.Key)
	}

	requireSpreaderScopeDefinition(t, spreaderScope, def)
}

func TestSpreadersListener_MultiConceptFromTo(t *testing.T) {
	input := `concepts
	Country=[Values[USA]]
	Product=[Values[A, B]]
	Region=[Values[North]]

spreaders
	from [Country, Product] to [Region, Country, Product] = Demand`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	spreadersScope := findSpreadersScope(rootTable)
	if spreadersScope == nil {
		t.Fatal("expected SpreadersScope child of root table")
	}

	if len(spreadersScope.Children) != 1 {
		t.Fatalf("expected 1 SpreaderScope child, got %d", len(spreadersScope.Children))
	}

	spreaderScope := spreadersScope.Children[0]
	def := requireSpreaderDefinition(t, spreaderScope)

	if len(def.FromConcepts) != 2 || def.FromConcepts[0] != "Country" || def.FromConcepts[1] != "Product" {
		t.Errorf("expected FromConcepts [Country, Product], got %v", def.FromConcepts)
	}
	if len(def.ToConcepts) != 3 || def.ToConcepts[0] != "Region" || def.ToConcepts[1] != "Country" || def.ToConcepts[2] != "Product" {
		t.Errorf("expected ToConcepts [Region, Country, Product], got %v", def.ToConcepts)
	}
	if def.FromEmpty {
		t.Error("expected FromEmpty false")
	}

	expectedKey := "__spreader__[Country,Product]->[Region,Country,Product]"
	if def.Key != expectedKey {
		t.Errorf("expected synthetic key %q, got %q", expectedKey, def.Key)
	}

	requireSpreaderScopeDefinition(t, spreaderScope, def)
}

func TestSpreadersListener_EmptySection(t *testing.T) {
	input := `spreaders
`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	spreadersScope := findSpreadersScope(rootTable)
	if spreadersScope == nil {
		t.Fatal("expected SpreadersScope child of root table even for empty section")
	}

	if len(spreadersScope.Children) != 0 {
		t.Errorf("expected 0 SpreaderScope children for empty section, got %d", len(spreadersScope.Children))
	}

	if len(spreadersScope.Definitions) != 0 {
		t.Errorf("expected 0 definitions for empty section, got %d", len(spreadersScope.Definitions))
	}
}

func TestSpreadersListener_MultipleDeclarations(t *testing.T) {
	input := `concepts
	Country=[Values[USA]]
	Product=[Values[A, B]]

spreaders
	from [Product] to [Country, Product] = Demand
	from [] to [Country] = Value`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	spreadersScope := findSpreadersScope(rootTable)
	if spreadersScope == nil {
		t.Fatal("expected SpreadersScope child of root table")
	}

	if len(spreadersScope.Children) != 2 {
		t.Fatalf("expected 2 SpreaderScope children, got %d", len(spreadersScope.Children))
	}

	// First spreader
	firstScope := spreadersScope.Children[0]
	firstDef := requireSpreaderDefinition(t, firstScope)
	if firstDef.Key != "__spreader__[Product]->[Country,Product]" {
		t.Errorf("unexpected first key: %s", firstDef.Key)
	}
	if firstDef.FromEmpty {
		t.Error("expected first spreader FromEmpty false")
	}
	requireSpreaderScopeDefinition(t, firstScope, firstDef)

	// Second spreader
	secondScope := spreadersScope.Children[1]
	secondDef := requireSpreaderDefinition(t, secondScope)
	if secondDef.Key != "__spreader__[]->[Country]" {
		t.Errorf("unexpected second key: %s", secondDef.Key)
	}
	if !secondDef.FromEmpty {
		t.Error("expected second spreader FromEmpty true")
	}
	if secondDef.FromConcepts != nil {
		t.Errorf("expected second FromConcepts nil, got %v", secondDef.FromConcepts)
	}
	requireSpreaderScopeDefinition(t, secondScope, secondDef)
}

func TestSpreaderSyntheticKey_EmptyFrom(t *testing.T) {
	key := spreaderSyntheticKey([]string{}, []string{"Country"})
	expected := "__spreader__[]->[Country]"
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestSpreaderSyntheticKey_SingleFromMultiTo(t *testing.T) {
	key := spreaderSyntheticKey([]string{"Product"}, []string{"Country", "Product"})
	expected := "__spreader__[Product]->[Country,Product]"
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestSpreaderSyntheticKey_MultiFromMultiTo(t *testing.T) {
	key := spreaderSyntheticKey([]string{"Country", "Product"}, []string{"Region", "Country"})
	expected := "__spreader__[Country,Product]->[Region,Country]"
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

// TestSpreadersListener_ArithmeticExpression verifies parsing a spreader
// with an arithmetic expression (not just a function call).
func TestSpreadersListener_ArithmeticExpression(t *testing.T) {
	input := `concepts
	Country=[Values[USA]]
	Product=[Values[A, B]]

spreaders
	from [Product] to [Country, Product] = Country + Product`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	spreadersScope := findSpreadersScope(rootTable)
	if spreadersScope == nil {
		t.Fatal("expected SpreadersScope child of root table")
	}

	if len(spreadersScope.Children) != 1 {
		t.Fatalf("expected 1 SpreaderScope child, got %d", len(spreadersScope.Children))
	}

	spreaderScope := spreadersScope.Children[0]
	def := requireSpreaderDefinition(t, spreaderScope)

	if len(def.FromConcepts) != 1 || def.FromConcepts[0] != "Product" {
		t.Errorf("expected FromConcepts [Product], got %v", def.FromConcepts)
	}
	if len(def.ToConcepts) != 2 || def.ToConcepts[0] != "Country" || def.ToConcepts[1] != "Product" {
		t.Errorf("expected ToConcepts [Country, Product], got %v", def.ToConcepts)
	}

	expectedKey := "__spreader__[Product]->[Country,Product]"
	if def.Key != expectedKey {
		t.Errorf("expected synthetic key %q, got %q", expectedKey, def.Key)
	}

	requireSpreaderScopeDefinition(t, spreaderScope, def)

	// Postfix: Country Product +
	if len(spreaderScope.Definitions) != 4 {
		t.Fatalf("expected 4 definitions (1 SpreaderScopeDefinition + 3 postfix), got %d", len(spreaderScope.Definitions))
	}
	requireIdentifierDef(t, spreaderScope.Definitions, 0, "Country")
	requireIdentifierDef(t, spreaderScope.Definitions, 1, "Product")
	if spreaderScope.Definitions[2].DefinitionType != symboltable.PlusOperatorDefinition {
		t.Fatalf("expected definition 2 to be PlusOperatorDefinition, got %s", spreaderScope.Definitions[2].DefinitionType)
	}
}

// TestSpreadersListener_MultipleFnArgs verifies parsing a spreader
// with multiple arguments in the function call.
func TestSpreadersListener_MultipleFnArgs(t *testing.T) {
	input := `concepts
	Country=[Values[USA]]
	Product=[Values[A, B]]

spreaders
	from [Product] to [Country, Product] = weighted_sum(Demand, Weight)`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	spreadersScope := findSpreadersScope(rootTable)
	if spreadersScope == nil {
		t.Fatal("expected SpreadersScope child of root table")
	}

	if len(spreadersScope.Children) != 1 {
		t.Fatalf("expected 1 SpreaderScope child, got %d", len(spreadersScope.Children))
	}

	spreaderScope := spreadersScope.Children[0]
	def := requireSpreaderDefinition(t, spreaderScope)

	if len(def.FromConcepts) != 1 || def.FromConcepts[0] != "Product" {
		t.Errorf("expected FromConcepts [Product], got %v", def.FromConcepts)
	}
	if len(def.ToConcepts) != 2 || def.ToConcepts[0] != "Country" || def.ToConcepts[1] != "Product" {
		t.Errorf("expected ToConcepts [Country, Product], got %v", def.ToConcepts)
	}

	expectedKey := "__spreader__[Product]->[Country,Product]"
	if def.Key != expectedKey {
		t.Errorf("expected synthetic key %q, got %q", expectedKey, def.Key)
	}

	requireSpreaderScopeDefinition(t, spreaderScope, def)

	// Postfix: Demand Weight weighted_sum
	if len(spreaderScope.Definitions) != 4 {
		t.Fatalf("expected 4 definitions (1 SpreaderScopeDefinition + 3 postfix), got %d", len(spreaderScope.Definitions))
	}
	requireIdentifierDef(t, spreaderScope.Definitions, 0, "Demand")
	requireIdentifierDef(t, spreaderScope.Definitions, 1, "Weight")
	requireIdentifierDef(t, spreaderScope.Definitions, 2, "weighted_sum")
}

// TestSpreadersListener_CombinedWithOtherSections verifies that spreaders
// parse correctly alongside other CPL sections.
func TestSpreadersListener_CombinedWithOtherSections(t *testing.T) {
	input := `concepts
	Country=[Values[USA]]
	Product=[Values[A, B]]

spreaders
	from [Product] to [Country, Product] = Demand

measures
	Demand = Number()`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	var spreadersScope *symboltable.SymbolTable
	for _, child := range rootTable.Children {
		if child.ScopeType == symboltable.SpreadersScope {
			spreadersScope = child
			break
		}
	}
	if spreadersScope == nil {
		t.Fatal("expected SpreadersScope child of root table")
	}

	if len(spreadersScope.Children) != 1 {
		t.Fatalf("expected 1 SpreaderScope child, got %d", len(spreadersScope.Children))
	}

	spreaderScope := spreadersScope.Children[0]
	def := requireSpreaderDefinition(t, spreaderScope)

	if def.Key != "__spreader__[Product]->[Country,Product]" {
		t.Errorf("unexpected key: %s", def.Key)
	}

	requireSpreaderScopeDefinition(t, spreaderScope, def)

	// Postfix: Demand
	if len(spreaderScope.Definitions) != 2 {
		t.Fatalf("expected 2 definitions (1 SpreaderScopeDefinition + 1 postfix), got %d", len(spreaderScope.Definitions))
	}
	requireIdentifierDef(t, spreaderScope.Definitions, 0, "Demand")
}

// TestSpreadersListener_SingleArgFnCall verifies parsing a spreader with a
// single-argument function call expression.
func TestSpreadersListener_SingleArgFnCall(t *testing.T) {
	input := `concepts
	Country=[Values[USA]]
	Product=[Values[A, B]]

spreaders
	from [Product] to [Country, Product] = sum(Demand)`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	spreadersScope := findSpreadersScope(rootTable)
	if spreadersScope == nil {
		t.Fatal("expected SpreadersScope child of root table")
	}

	if len(spreadersScope.Children) != 1 {
		t.Fatalf("expected 1 SpreaderScope child, got %d", len(spreadersScope.Children))
	}

	spreaderScope := spreadersScope.Children[0]
	def := requireSpreaderDefinition(t, spreaderScope)

	if len(def.FromConcepts) != 1 || def.FromConcepts[0] != "Product" {
		t.Errorf("expected FromConcepts [Product], got %v", def.FromConcepts)
	}
	if len(def.ToConcepts) != 2 || def.ToConcepts[0] != "Country" || def.ToConcepts[1] != "Product" {
		t.Errorf("expected ToConcepts [Country, Product], got %v", def.ToConcepts)
	}

	expectedKey := "__spreader__[Product]->[Country,Product]"
	if def.Key != expectedKey {
		t.Errorf("expected synthetic key %q, got %q", expectedKey, def.Key)
	}

	requireSpreaderScopeDefinition(t, spreaderScope, def)

	// Postfix: Demand sum
	if len(spreaderScope.Definitions) != 3 {
		t.Fatalf("expected 3 definitions (1 SpreaderScopeDefinition + 2 postfix), got %d", len(spreaderScope.Definitions))
	}
	requireIdentifierDef(t, spreaderScope.Definitions, 0, "Demand")
	requireIdentifierDef(t, spreaderScope.Definitions, 1, "sum")
}

// TestSpreadersListener_GenericIdentifier verifies parsing a spreader with a
// generic identifier expression (not a concept reference, deferred to analysis).
func TestSpreadersListener_GenericIdentifier(t *testing.T) {
	input := `concepts
	Country=[Values[USA]]
	Product=[Values[A, B]]

spreaders
	from [Product] to [Country, Product] = my_metric`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	spreadersScope := findSpreadersScope(rootTable)
	if spreadersScope == nil {
		t.Fatal("expected SpreadersScope child of root table")
	}

	if len(spreadersScope.Children) != 1 {
		t.Fatalf("expected 1 SpreaderScope child, got %d", len(spreadersScope.Children))
	}

	spreaderScope := spreadersScope.Children[0]
	def := requireSpreaderDefinition(t, spreaderScope)

	expectedKey := "__spreader__[Product]->[Country,Product]"
	if def.Key != expectedKey {
		t.Errorf("expected synthetic key %q, got %q", expectedKey, def.Key)
	}

	requireSpreaderScopeDefinition(t, spreaderScope, def)

	// Postfix: my_metric (as Identifier/StringData)
	if len(spreaderScope.Definitions) != 2 {
		t.Fatalf("expected 2 definitions (1 SpreaderScopeDefinition + 1 postfix), got %d", len(spreaderScope.Definitions))
	}
	requireIdentifierDef(t, spreaderScope.Definitions, 0, "my_metric")
}

// TestSpreadersListener_FullDocument verifies that a complete CPL document
// with concepts, measures, and spreaders sections parses end-to-end correctly.
func TestSpreadersListener_FullDocument(t *testing.T) {
	input := `concepts
	Country=[Values[USA]]
	Product=[Values[A, B]]

measures
	Demand = Number()

spreaders
	from [Product] to [Country, Product] = Demand
	from [] to [Country] = Value`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	spreadersScope := findSpreadersScope(rootTable)
	if spreadersScope == nil {
		t.Fatal("expected SpreadersScope child of root table")
	}

	if len(spreadersScope.Children) != 2 {
		t.Fatalf("expected 2 SpreaderScope children, got %d", len(spreadersScope.Children))
	}

	// First spreader
	firstScope := spreadersScope.Children[0]
	firstDef := requireSpreaderDefinition(t, firstScope)
	if firstDef.Key != "__spreader__[Product]->[Country,Product]" {
		t.Errorf("unexpected first key: %s", firstDef.Key)
	}
	requireSpreaderScopeDefinition(t, firstScope, firstDef)
	if len(firstScope.Definitions) != 2 {
		t.Fatalf("expected 2 definitions for first spreader, got %d", len(firstScope.Definitions))
	}
	requireIdentifierDef(t, firstScope.Definitions, 0, "Demand")

	// Second spreader
	secondScope := spreadersScope.Children[1]
	secondDef := requireSpreaderDefinition(t, secondScope)
	if secondDef.Key != "__spreader__[]->[Country]" {
		t.Errorf("unexpected second key: %s", secondDef.Key)
	}
	if !secondDef.FromEmpty {
		t.Error("expected second spreader FromEmpty true")
	}
	requireSpreaderScopeDefinition(t, secondScope, secondDef)
	if len(secondScope.Definitions) != 2 {
		t.Fatalf("expected 2 definitions for second spreader, got %d", len(secondScope.Definitions))
	}
	requireIdentifierDef(t, secondScope.Definitions, 0, "Value")
}

// TestSpreadersListener_MalformedSyntax verifies that malformed spreader
// declarations produce parse errors and do not leave partial
// ParserSpreaderDefinition entries in SpreaderScope.Data.
func TestSpreadersListener_MalformedSyntax(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectError    bool
		errorContains  string
		expectSpdScope int // expected number of SpreaderScope children
	}{
		{
			name: "missing to clause",
			input: `spreaders
		from [Product] = Demand`,
			expectError:    true,
			errorContains:  "mismatched input '='",
			expectSpdScope: 0,
		},
		{
			name: "missing from clause",
			input: `spreaders
		to [Country] = Demand`,
			expectError:    true,
			errorContains:  "extraneous input 'to'",
			expectSpdScope: 0,
		},
		{
			name: "missing expression after equals",
			input: `spreaders
		from [Product] to [Country] =`,
			expectError:    true,
			errorContains:  "mismatched input '<EOF>'",
			expectSpdScope: 0,
		},
		{
			name: "empty to list",
			input: `spreaders
		from [Product] to [] = Demand`,
			expectError:    true,
			errorContains:  "missing target concepts",
			expectSpdScope: 0,
		},
		{
			name: "empty spreader declaration",
			input: `spreaders
		from [] to [] = Demand`,
			expectError:    true,
			errorContains:  "missing target concepts",
			expectSpdScope: 0,
		},
		{
			name: "multiple declarations with one malformed",
			input: `spreaders
		from [Product] to [Country] = Demand
		from [Product] = Demand`,
			expectError:    true,
			errorContains:  "mismatched input '='",
			expectSpdScope: 1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rootTable, err := Parse(tc.input)

			if !tc.expectError {
				if err != nil {
					t.Fatalf("unexpected parse error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected parse error for input:\n%s", tc.input)
			}

			if tc.errorContains != "" {
				if !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("expected error to contain %q, got: %v", tc.errorContains, err)
				}
			}

			spreadersScope := findSpreadersScope(rootTable)
			if spreadersScope == nil {
				if tc.expectSpdScope > 0 {
					t.Fatal("expected SpreadersScope child of root table")
				}
				return
			}

			if len(spreadersScope.Children) != tc.expectSpdScope {
				t.Fatalf("expected %d SpreaderScope children, got %d", tc.expectSpdScope, len(spreadersScope.Children))
			}

			// When no children are expected, ensure no partial ParserSpreaderDefinition
			// entries are left behind by malformed declarations.
			if tc.expectSpdScope == 0 {
				for i, child := range spreadersScope.Children {
					if child.Data != nil {
						t.Errorf("child %d should not have partial Data, got %T", i, child.Data)
					}
				}
			}
		})
	}
}
