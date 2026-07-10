package parser

import (
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// findAggregatorsScope traverses the symbol table tree to find the AggregatorsScope child.
func findAggregatorsScope(root *symboltable.SymbolTable) *symboltable.SymbolTable {
	for _, child := range root.Children {
		if child.ScopeType == symboltable.AggregatorsScope {
			return child
		}
	}
	return nil
}

// requireAggregatorDefinition is a helper that extracts a ParserAggregatorDefinition from an AggregatorScope.
func requireAggregatorDefinition(t *testing.T, aggregatorScope *symboltable.SymbolTable) *ParserAggregatorDefinition {
	t.Helper()

	if aggregatorScope.Data == nil {
		t.Fatal("AggregatorScope.Data should be populated")
	}
	def, ok := aggregatorScope.Data.(*ParserAggregatorDefinition)
	if !ok {
		t.Fatalf("AggregatorScope.Data should be a *ParserAggregatorDefinition, got %T", aggregatorScope.Data)
	}
	if def == nil {
		t.Fatal("AggregatorScope.Data should not contain a nil ParserAggregatorDefinition pointer")
	}
	return def
}

// requireAggregatorScopeDefinition verifies that the AggregatorScope's Definitions
// contains an AggregatorScopeDefinition entry with the expected synthetic key and
// the same *ParserAggregatorDefinition pointer as scope.Data.
func requireAggregatorScopeDefinition(t *testing.T, aggregatorScope *symboltable.SymbolTable, expectedDef *ParserAggregatorDefinition) {
	t.Helper()

	found := false
	for _, def := range aggregatorScope.Definitions {
		if def.DefinitionType == symboltable.AggregatorScopeDefinition {
			found = true
			if def.Name == nil || *def.Name != expectedDef.Key {
				t.Errorf("expected AggregatorScopeDefinition name %q, got %v", expectedDef.Key, def.Name)
			}
			pad, ok := def.Data.(*ParserAggregatorDefinition)
			if !ok {
				t.Fatalf("expected AggregatorScopeDefinition Data to be *ParserAggregatorDefinition, got %T", def.Data)
			}
			if pad != expectedDef {
				t.Fatalf("expected AggregatorScopeDefinition Data to be same pointer as scope.Data")
			}
			break
		}
	}
	if !found {
		t.Fatal("expected AggregatorScopeDefinition in Definitions, got none")
	}
}

func TestAggregatorsListener_SingleConceptToSameConcept(t *testing.T) {
	input := `aggregators
	from [Country] to [Country] = sum(Country)`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	aggregatorsScope := findAggregatorsScope(rootTable)
	if aggregatorsScope == nil {
		t.Fatal("expected AggregatorsScope child of root table")
	}

	if len(aggregatorsScope.Children) != 1 {
		t.Fatalf("expected 1 AggregatorScope child, got %d", len(aggregatorsScope.Children))
	}

	aggregatorScope := aggregatorsScope.Children[0]
	if aggregatorScope.ScopeType != symboltable.AggregatorScope {
		t.Fatalf("expected AggregatorScope, got %s", aggregatorScope.ScopeType)
	}

	// Verify definition data
	def := requireAggregatorDefinition(t, aggregatorScope)
	if len(def.FromConcepts) != 1 || def.FromConcepts[0] != "Country" {
		t.Errorf("expected FromConcepts [Country], got %v", def.FromConcepts)
	}
	if def.ToConcept != "Country" {
		t.Errorf("expected ToConcept 'Country', got '%s'", def.ToConcept)
	}
	if def.ToEmpty {
		t.Error("expected ToEmpty false")
	}

	// Verify synthetic key
	expectedKey := "__aggregator__[Country]->[Country]"
	if def.Key != expectedKey {
		t.Errorf("expected synthetic key %q, got %q", expectedKey, def.Key)
	}

	requireAggregatorScopeDefinition(t, aggregatorScope, def)

	// Verify postfix definitions: Country sum
	if len(aggregatorScope.Definitions) != 3 {
		t.Fatalf("expected 3 definitions (1 AggregatorScopeDefinition + 2 postfix), got %d", len(aggregatorScope.Definitions))
	}
	requireIdentifierDef(t, aggregatorScope.Definitions, 0, "Country")
	requireIdentifierDef(t, aggregatorScope.Definitions, 1, "sum")
}

func TestAggregatorsListener_MultiConceptFromTo(t *testing.T) {
	input := `aggregators
	from [Country, Product] to [Product] = sum(Demand)`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	aggregatorsScope := findAggregatorsScope(rootTable)
	if aggregatorsScope == nil {
		t.Fatal("expected AggregatorsScope child of root table")
	}

	if len(aggregatorsScope.Children) != 1 {
		t.Fatalf("expected 1 AggregatorScope child, got %d", len(aggregatorsScope.Children))
	}

	aggregatorScope := aggregatorsScope.Children[0]
	if aggregatorScope.ScopeType != symboltable.AggregatorScope {
		t.Fatalf("expected AggregatorScope, got %s", aggregatorScope.ScopeType)
	}

	def := requireAggregatorDefinition(t, aggregatorScope)
	if len(def.FromConcepts) != 2 || def.FromConcepts[0] != "Country" || def.FromConcepts[1] != "Product" {
		t.Errorf("expected FromConcepts [Country, Product], got %v", def.FromConcepts)
	}
	if def.ToConcept != "Product" {
		t.Errorf("expected ToConcept 'Product', got '%s'", def.ToConcept)
	}
	if def.ToEmpty {
		t.Error("expected ToEmpty false")
	}

	expectedKey := "__aggregator__[Country,Product]->[Product]"
	if def.Key != expectedKey {
		t.Errorf("expected synthetic key %q, got %q", expectedKey, def.Key)
	}

	requireAggregatorScopeDefinition(t, aggregatorScope, def)

	// Postfix: Demand sum
	if len(aggregatorScope.Definitions) != 3 {
		t.Fatalf("expected 3 definitions (1 AggregatorScopeDefinition + 2 postfix), got %d", len(aggregatorScope.Definitions))
	}
	requireIdentifierDef(t, aggregatorScope.Definitions, 0, "Demand")
	requireIdentifierDef(t, aggregatorScope.Definitions, 1, "sum")
}

func TestAggregatorsListener_ScalarReduction(t *testing.T) {
	input := `aggregators
	from [Country] to [] = sum(Country)`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	aggregatorsScope := findAggregatorsScope(rootTable)
	if aggregatorsScope == nil {
		t.Fatal("expected AggregatorsScope child of root table")
	}

	if len(aggregatorsScope.Children) != 1 {
		t.Fatalf("expected 1 AggregatorScope child, got %d", len(aggregatorsScope.Children))
	}

	aggregatorScope := aggregatorsScope.Children[0]
	def := requireAggregatorDefinition(t, aggregatorScope)

	if len(def.FromConcepts) != 1 || def.FromConcepts[0] != "Country" {
		t.Errorf("expected FromConcepts [Country], got %v", def.FromConcepts)
	}
	if def.ToConcept != "" {
		t.Errorf("expected ToConcept '', got '%s'", def.ToConcept)
	}
	if !def.ToEmpty {
		t.Error("expected ToEmpty true")
	}

	expectedKey := "__aggregator__[Country]->[]"
	if def.Key != expectedKey {
		t.Errorf("expected synthetic key %q, got %q", expectedKey, def.Key)
	}

	requireAggregatorScopeDefinition(t, aggregatorScope, def)
}

func TestAggregatorsListener_EmptySection(t *testing.T) {
	input := `aggregators
`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	aggregatorsScope := findAggregatorsScope(rootTable)
	if aggregatorsScope == nil {
		t.Fatal("expected AggregatorsScope child of root table even for empty section")
	}

	if len(aggregatorsScope.Children) != 0 {
		t.Errorf("expected 0 AggregatorScope children for empty section, got %d", len(aggregatorsScope.Children))
	}

	if len(aggregatorsScope.Definitions) != 0 {
		t.Errorf("expected 0 definitions for empty section, got %d", len(aggregatorsScope.Definitions))
	}
}

func TestAggregatorsListener_MultipleDeclarations(t *testing.T) {
	input := `aggregators
	from [Country] to [Country] = sum(Country)
	from [Product] to [] = count(Orders)`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	aggregatorsScope := findAggregatorsScope(rootTable)
	if aggregatorsScope == nil {
		t.Fatal("expected AggregatorsScope child of root table")
	}

	if len(aggregatorsScope.Children) != 2 {
		t.Fatalf("expected 2 AggregatorScope children, got %d", len(aggregatorsScope.Children))
	}

	// First aggregator
	firstScope := aggregatorsScope.Children[0]
	firstDef := requireAggregatorDefinition(t, firstScope)
	if firstDef.Key != "__aggregator__[Country]->[Country]" {
		t.Errorf("unexpected first key: %s", firstDef.Key)
	}
	requireAggregatorScopeDefinition(t, firstScope, firstDef)

	// Second aggregator
	secondScope := aggregatorsScope.Children[1]
	secondDef := requireAggregatorDefinition(t, secondScope)
	if secondDef.Key != "__aggregator__[Product]->[]" {
		t.Errorf("unexpected second key: %s", secondDef.Key)
	}
	if !secondDef.ToEmpty {
		t.Error("expected second aggregator ToEmpty true")
	}
	if len(secondDef.FromConcepts) != 1 || secondDef.FromConcepts[0] != "Product" {
		t.Errorf("expected FromConcepts [Product], got %v", secondDef.FromConcepts)
	}
	requireAggregatorScopeDefinition(t, secondScope, secondDef)

	// Verify postfix of second aggregator: Orders count
	if len(secondScope.Definitions) != 3 {
		t.Fatalf("expected 3 definitions (1 AggregatorScopeDefinition + 2 postfix) in second aggregator, got %d", len(secondScope.Definitions))
	}
	requireIdentifierDef(t, secondScope.Definitions, 0, "Orders")
	requireIdentifierDef(t, secondScope.Definitions, 1, "count")
}

func TestAggregatorSyntheticKey_EmptyFrom(t *testing.T) {
	key := aggregatorSyntheticKey([]string{}, "Target")
	expected := "__aggregator__[]->[Target]"
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

// TestAggregatorsListener_ArithmeticExpression verifies parsing an aggregator
// with an arithmetic expression (not just a function call).
func TestAggregatorsListener_ArithmeticExpression(t *testing.T) {
	input := `aggregators
	from [Country, Product] to [Product] = Country + Product`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	aggregatorsScope := findAggregatorsScope(rootTable)
	if aggregatorsScope == nil {
		t.Fatal("expected AggregatorsScope child of root table")
	}

	if len(aggregatorsScope.Children) != 1 {
		t.Fatalf("expected 1 AggregatorScope child, got %d", len(aggregatorsScope.Children))
	}

	aggregatorScope := aggregatorsScope.Children[0]
	def := requireAggregatorDefinition(t, aggregatorScope)

	if len(def.FromConcepts) != 2 || def.FromConcepts[0] != "Country" || def.FromConcepts[1] != "Product" {
		t.Errorf("expected FromConcepts [Country, Product], got %v", def.FromConcepts)
	}
	if def.ToConcept != "Product" {
		t.Errorf("expected ToConcept 'Product', got '%s'", def.ToConcept)
	}

	expectedKey := "__aggregator__[Country,Product]->[Product]"
	if def.Key != expectedKey {
		t.Errorf("expected synthetic key %q, got %q", expectedKey, def.Key)
	}

	requireAggregatorScopeDefinition(t, aggregatorScope, def)

	// Postfix: Country Product +
	if len(aggregatorScope.Definitions) != 4 {
		t.Fatalf("expected 4 definitions (1 AggregatorScopeDefinition + 3 postfix), got %d", len(aggregatorScope.Definitions))
	}
	requireIdentifierDef(t, aggregatorScope.Definitions, 0, "Country")
	requireIdentifierDef(t, aggregatorScope.Definitions, 1, "Product")
	if aggregatorScope.Definitions[2].DefinitionType != symboltable.PlusOperatorDefinition {
		t.Fatalf("expected definition 2 to be PlusOperatorDefinition, got %s", aggregatorScope.Definitions[2].DefinitionType)
	}
}

// TestAggregatorsListener_MultipleFnArgs verifies parsing an aggregator
// with multiple arguments in the function call.
func TestAggregatorsListener_MultipleFnArgs(t *testing.T) {
	input := `aggregators
	from [Country, Product] to [Product] = weighted_sum(Demand, Weight)`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	aggregatorsScope := findAggregatorsScope(rootTable)
	if aggregatorsScope == nil {
		t.Fatal("expected AggregatorsScope child of root table")
	}

	if len(aggregatorsScope.Children) != 1 {
		t.Fatalf("expected 1 AggregatorScope child, got %d", len(aggregatorsScope.Children))
	}

	aggregatorScope := aggregatorsScope.Children[0]
	def := requireAggregatorDefinition(t, aggregatorScope)

	if len(def.FromConcepts) != 2 || def.FromConcepts[0] != "Country" || def.FromConcepts[1] != "Product" {
		t.Errorf("expected FromConcepts [Country, Product], got %v", def.FromConcepts)
	}
	if def.ToConcept != "Product" {
		t.Errorf("expected ToConcept 'Product', got '%s'", def.ToConcept)
	}

	expectedKey := "__aggregator__[Country,Product]->[Product]"
	if def.Key != expectedKey {
		t.Errorf("expected synthetic key %q, got %q", expectedKey, def.Key)
	}

	requireAggregatorScopeDefinition(t, aggregatorScope, def)

	// Postfix: Demand, Weight, weighted_sum
	if len(aggregatorScope.Definitions) != 4 {
		t.Fatalf("expected 4 definitions (1 AggregatorScopeDefinition + 3 postfix), got %d", len(aggregatorScope.Definitions))
	}
	requireIdentifierDef(t, aggregatorScope.Definitions, 0, "Demand")
	requireIdentifierDef(t, aggregatorScope.Definitions, 1, "Weight")
	requireIdentifierDef(t, aggregatorScope.Definitions, 2, "weighted_sum")
}

// TestAggregatorsListener_CombinedWithOtherSections verifies that aggregators
// parse correctly alongside other CPL sections.
func TestAggregatorsListener_CombinedWithOtherSections(t *testing.T) {
	input := `concepts
	Country=[Values[USA]]
	Product=[Values[A, B]]

aggregators
	from [Country, Product] to [Product] = sum(Demand)

measures
	Demand = Number()`

	rootTable, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	var aggregatorsScope *symboltable.SymbolTable
	for _, child := range rootTable.Children {
		if child.ScopeType == symboltable.AggregatorsScope {
			aggregatorsScope = child
			break
		}
	}
	if aggregatorsScope == nil {
		t.Fatal("expected AggregatorsScope child of root table")
	}

	if len(aggregatorsScope.Children) != 1 {
		t.Fatalf("expected 1 AggregatorScope child, got %d", len(aggregatorsScope.Children))
	}

	aggregatorScope := aggregatorsScope.Children[0]
	def := requireAggregatorDefinition(t, aggregatorScope)

	if def.Key != "__aggregator__[Country,Product]->[Product]" {
		t.Errorf("unexpected key: %s", def.Key)
	}

	requireAggregatorScopeDefinition(t, aggregatorScope, def)

	// Postfix: Demand sum
	if len(aggregatorScope.Definitions) != 3 {
		t.Fatalf("expected 3 definitions (1 AggregatorScopeDefinition + 2 postfix), got %d", len(aggregatorScope.Definitions))
	}
	requireIdentifierDef(t, aggregatorScope.Definitions, 0, "Demand")
	requireIdentifierDef(t, aggregatorScope.Definitions, 1, "sum")
}

// TestAggregatorsListener_MalformedSyntax verifies that malformed aggregator
// declarations produce parse errors and do not leave partial
// ParserAggregatorDefinition entries in AggregatorScope.Data.
func TestAggregatorsListener_MalformedSyntax(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectError    bool
		errorContains  string
		expectAggScope int // expected number of AggregatorScope children
	}{
		{
			name: "missing from clause",
			input: `aggregators
	to [Product] = sum(Demand)`,
			expectError:    true,
			errorContains:  "extraneous input 'to'",
			expectAggScope: 0,
		},
		{
			name: "missing to clause",
			input: `aggregators
	from [Demand] = sum(Demand)`,
			expectError:    true,
			errorContains:  "mismatched input '='",
			expectAggScope: 0,
		},
		{
			name: "empty from list",
			input: `aggregators
	from [] to [Product] = sum(Demand)`,
			expectError:    true,
			errorContains:  "missing",
			expectAggScope: 0,
		},
		{
			name: "missing expression after equals",
			input: `aggregators
	from [Demand] to [Product] =`,
			expectError:    true,
			errorContains:  "mismatched input '<EOF>'",
			expectAggScope: 0,
		},
		{
			name: "multiple declarations with one malformed",
			input: `aggregators
	from [Country] to [Country] = sum(Country)
	from [Demand] = sum(Demand)`,
			expectError:    true,
			errorContains:  "mismatched input '='",
			expectAggScope: 1,
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

			aggregatorsScope := findAggregatorsScope(rootTable)
			if aggregatorsScope == nil {
				if tc.expectAggScope > 0 {
					t.Fatal("expected AggregatorsScope child of root table")
				}
				return
			}

			if len(aggregatorsScope.Children) != tc.expectAggScope {
				t.Fatalf("expected %d AggregatorScope children, got %d", tc.expectAggScope, len(aggregatorsScope.Children))
			}
		})
	}
}
