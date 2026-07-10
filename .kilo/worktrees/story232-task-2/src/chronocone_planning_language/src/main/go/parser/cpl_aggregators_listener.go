// Package parser provides CPL parsing functionality.
//
// This file contains listener callbacks related to the CPLAggregators.g4 grammar file,
// which defines the aggregators section and aggregator declarations.
//
// Grammar rules covered:
//   - aggregatorsSection, aggregatorDeclaration
//
// Symbol Table Driven Parsing:
// The parser builds a symbol table stack during parsing. Aggregator definitions are
// stored in AggregatorScope symbol tables under synthetic keys. Detailed extraction
// and validation of aggregator definitions is deferred to the analysis phase.
//
// Story: Story255_aggregators-increment2-parse.md
package parser

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// EnterAggregatorsSection is called when entering the aggregators section.
// It pushes a new AggregatorsScope symbol table onto the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) EnterAggregatorsSection(ctx *antlr.AggregatorsSectionContext) {
	l.logEnter("AggregatorsSection", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the Aggregators scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering Aggregators section: %w", err))
		return
	}

	aggregatorsTable := current.AddChild(symboltable.AggregatorsScope)
	l.symbolStack.Push(aggregatorsTable)
}

// ExitAggregatorsSection is called when exiting the aggregators section.
// It pops the Aggregators scope symbol table from the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) ExitAggregatorsSection(ctx *antlr.AggregatorsSectionContext) {
	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting Aggregators section: %w", err))
	}
	l.logExit("AggregatorsSection", &ctx.BaseParserRuleContext)
}

// EnterAggregatorDeclaration is called when entering an aggregator declaration.
// It pushes a new AggregatorScope onto the symbol stack so that expression
// definitions (concept references, fnCall names, operators) accumulate in
// postfix order.
//
// Example aggregator declaration:
//
//	from [Country, Product] to [Product] = sum(Demand)
//
// This corresponds to the 'aggregatorDeclaration' rule in CPLAggregators.g4.
func (l *CPLListener) EnterAggregatorDeclaration(ctx *antlr.AggregatorDeclarationContext) {
	l.logEnter("AggregatorDeclaration", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the Aggregator scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering Aggregator declaration: %w", err))
		return
	}

	// Create a child scope for this aggregator definition and push it on the stack
	l.symbolStack.Push(current.AddChild(symboltable.AggregatorScope))
}

// ExitAggregatorDeclaration is called when exiting an aggregator declaration.
// It extracts the from concept names, to target, creates a ParserAggregatorDefinition,
// computes the synthetic key, stores the definition in the AggregatorScope's Data
// field, and pops the aggregator scope.
//
// Symbol Table Driven Parsing:
// The from/to identifiers are extracted directly from the parse tree context.
// The fnCall expression identifiers are accumulated in the AggregatorScope's
// Definitions slice by the root listener's expression handlers (ExitFnCall).
func (l *CPLListener) ExitAggregatorDeclaration(ctx *antlr.AggregatorDeclarationContext) {
	l.logExit("AggregatorDeclaration", &ctx.BaseParserRuleContext)

	// Get the current AggregatorScope before popping
	aggregatorScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting Aggregator declaration: %w", err))
		return
	}

	// Extract from concepts from the identifierList after 'from'
	var fromConcepts []string
	if identifierList := ctx.IdentifierList(); identifierList != nil {
		identifiers := identifierList.AllIdentifier()
		fromConcepts = make([]string, 0, len(identifiers))
		for _, ident := range identifiers {
			text := ident.GetText()
			if text == "" {
				l.AddError(fmt.Errorf("aggregator declaration has empty source concept in 'from' clause"))
				removeChildSymbolTable(aggregatorScope.Parent, aggregatorScope)
				_, _ = l.symbolStack.Pop()
				return
			}
			fromConcepts = append(fromConcepts, text)
		}
	}
	if len(fromConcepts) == 0 {
		l.AddError(fmt.Errorf("aggregator declaration missing source concepts in 'from' clause"))
		removeChildSymbolTable(aggregatorScope.Parent, aggregatorScope)
		_, _ = l.symbolStack.Pop()
		return
	}

	// Validate to target exists (to [] is valid; missing to is not)
	if ctx.ToTarget() == nil {
		l.AddError(fmt.Errorf("aggregator declaration missing 'to' clause"))
		removeChildSymbolTable(aggregatorScope.Parent, aggregatorScope)
		_, _ = l.symbolStack.Pop()
		return
	}

	// Extract to target
	var toConcept string
	toEmpty := false
	if toTarget := ctx.ToTarget(); toTarget != nil {
		if ident := toTarget.Identifier(); ident != nil {
			toConcept = ident.GetText()
		} else {
			// Empty alternative: to []
			toEmpty = true
			toConcept = ""
		}
	}

	// Validate that either fnCall or expression exists after '='
	if ctx.FnCall() == nil && ctx.Expression() == nil {
		l.AddError(fmt.Errorf("aggregator declaration missing expression after '='"))
		removeChildSymbolTable(aggregatorScope.Parent, aggregatorScope)
		_, _ = l.symbolStack.Pop()
		return
	}
	// If fnCall is present, validate it has a non-empty function identifier
	if ctx.FnCall() != nil {
		if fnIdent := ctx.FnCall().Identifier(); fnIdent == nil || fnIdent.GetText() == "" {
			l.AddError(fmt.Errorf("aggregator declaration has empty function call"))
			removeChildSymbolTable(aggregatorScope.Parent, aggregatorScope)
			_, _ = l.symbolStack.Pop()
			return
		}
	}

	// Get source location from the start of the declaration
	location := l.makeSourceLocation(ctx.GetStart())

	// Create the ParserAggregatorDefinition
	def := &ParserAggregatorDefinition{
		FromConcepts:   fromConcepts,
		ToConcept:      toConcept,
		ToEmpty:        toEmpty,
		SourceLocation: location,
	}

	// Compute synthetic key
	def.Key = aggregatorSyntheticKey(fromConcepts, toConcept)

	// Store the ParserAggregatorDefinition in the aggregator scope's Data field
	aggregatorScope.Data = def

	// Also push an AggregatorScopeDefinition entry into the scope's Definitions slice
	keyCopy := def.Key
	_ = aggregatorScope.DefineWithData(&keyCopy, symboltable.AggregatorScopeDefinition, location, def)

	// Pop the Aggregator scope from the symbol stack
	_, err = l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting Aggregator declaration: %w", err))
		return
	}
}
