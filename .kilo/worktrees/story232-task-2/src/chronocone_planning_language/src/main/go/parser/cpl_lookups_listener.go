// Package parser provides CPL parsing functionality.
//
// This file contains listener callbacks related to the CPLLookups.g4 grammar file,
// which defines the lookups section and lookup declarations.
//
// Grammar rules covered:
//   - lookupsSection, lookupDeclaration, dimensionsClause, celltypeClause
//
// Symbol Table Driven Parsing:
// The parser builds a symbol table stack during parsing. Lookup definitions are
// stored as SymbolDefinition entries in a LookupsScope symbol table. Unlike
// metrics and connectors, lookup scopes are flat — there are no per-lookup child
// scopes. Detailed extraction and validation of lookup definitions is deferred
// to the analysis phase.
//
// Story: Story241_lookups-increment1-parse.md
package parser

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// EnterLookupsSection is called when entering the lookups section.
// It pushes a new LookupsScope symbol table onto the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) EnterLookupsSection(ctx *antlr.LookupsSectionContext) {
	l.logEnter("LookupsSection", &ctx.BaseParserRuleContext)

	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering Lookups section: %w", err))
		return
	}

	lookupsTable := current.AddChild(symboltable.LookupsScope)
	l.symbolStack.Push(lookupsTable)
}

// ExitLookupsSection is called when exiting the lookups section.
// It pops the LookupsScope symbol table from the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) ExitLookupsSection(ctx *antlr.LookupsSectionContext) {
	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting Lookups section: %w", err))
	}
	l.logExit("LookupsSection", &ctx.BaseParserRuleContext)
}

// EnterLookupDeclaration is called when entering a lookup declaration.
// Lookups use flat scopes — no child scope is pushed here.
func (l *CPLListener) EnterLookupDeclaration(ctx *antlr.LookupDeclarationContext) {
	l.logEnter("LookupDeclaration", &ctx.BaseParserRuleContext)
}

// ExitLookupDeclaration is called when exiting a lookup declaration.
// It extracts the lookup name, dimension names, and celltype name from the
// parse tree, creates a ParserLookupDefinition, and stores it as a
// SymbolDefinition in the current LookupsScope's Definitions list.
//
// Symbol Table Driven Parsing:
// Lookup scopes are flat (no per-lookup child scopes). The definition is
// stored directly in the LookupsScope's Definitions list with
// DefinitionType set to LookupScopeDefinition.
//
// Example lookup declaration:
//
//	SalesLookup
//	    dimensions[Store]
//	    celltype RevenueMeasure
func (l *CPLListener) ExitLookupDeclaration(ctx *antlr.LookupDeclarationContext) {
	l.logExit("LookupDeclaration", &ctx.BaseParserRuleContext)

	currentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting Lookup declaration: %w", err))
		return
	}

	// Extract the lookup name from the direct identifier child
	nameIdent := ctx.Identifier()
	if nameIdent == nil {
		l.AddError(fmt.Errorf("lookup declaration missing identifier"))
		return
	}
	lookupName := nameIdent.GetText()
	if lookupName == "" {
		l.AddError(fmt.Errorf("lookup declaration has empty identifier"))
		return
	}
	nameLocation := l.makeSourceLocation(nameIdent.GetStart())

	// Extract dimension names from dimensionsClause
	dimClause := ctx.DimensionsClause()
	if dimClause == nil {
		l.AddError(fmt.Errorf("lookup declaration '%s' missing dimensions clause", lookupName))
		return
	}
	dimIdentifiers := dimClause.AllIdentifier()
	if len(dimIdentifiers) == 0 {
		l.AddError(fmt.Errorf("lookup declaration '%s' has no dimensions", lookupName))
		return
	}
	dimensionNames := make([]string, 0, len(dimIdentifiers))
	for _, di := range dimIdentifiers {
		dimensionNames = append(dimensionNames, di.GetText())
	}

	// Extract celltype name from celltypeClause
	cellClause := ctx.CelltypeClause()
	if cellClause == nil {
		l.AddError(fmt.Errorf("lookup declaration '%s' missing celltype clause", lookupName))
		return
	}
	cellIdent := cellClause.Identifier()
	if cellIdent == nil {
		l.AddError(fmt.Errorf("lookup declaration '%s' celltype clause missing identifier", lookupName))
		return
	}
	celltypeName := cellIdent.GetText()

	pld := &ParserLookupDefinition{
		Name:           lookupName,
		Dimensions:     dimensionNames,
		Celltype:       celltypeName,
		SourceLocation: nameLocation,
	}

	nameCopy := lookupName
	if err := currentScope.DefineWithData(&nameCopy, symboltable.LookupScopeDefinition, nameLocation, pld); err != nil {
		l.AddError(fmt.Errorf("error defining lookup '%s': %w", lookupName, err))
	}
}
