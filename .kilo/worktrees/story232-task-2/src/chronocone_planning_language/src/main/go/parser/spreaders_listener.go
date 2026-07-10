// Package parser provides CPL parsing functionality.
//
// This file contains listener callbacks related to the CPLSpreaders.g4 grammar file,
// which defines the spreaders section and spreader declarations.
//
// Grammar rules covered:
//   - spreadersSection, spreaderDeclaration
//
// Symbol Table Driven Parsing:
// The parser builds a symbol table stack during parsing. Spreader definitions are
// stored in SpreaderScope symbol tables under synthetic keys. Detailed extraction
// and validation of spreader definitions is deferred to the analysis phase.
//
// Story: Story264_spreaders-increment1-parse.md
package parser

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// EnterSpreadersSection is called when entering the spreaders section.
// It pushes a new SpreadersScope symbol table onto the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) EnterSpreadersSection(ctx *antlr.SpreadersSectionContext) {
	l.logEnter("SpreadersSection", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the Spreaders scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering Spreaders section: %w", err))
		return
	}

	spreadersTable := current.AddChild(symboltable.SpreadersScope)
	l.symbolStack.Push(spreadersTable)
}

// ExitSpreadersSection is called when exiting the spreaders section.
// It pops the Spreaders scope symbol table from the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) ExitSpreadersSection(ctx *antlr.SpreadersSectionContext) {
	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting Spreaders section: %w", err))
	}
	l.logExit("SpreadersSection", &ctx.BaseParserRuleContext)
}

// EnterSpreaderDeclaration is called when entering a spreader declaration.
// It pushes a new SpreaderScope onto the symbol stack so that expression
// definitions (concept references, fnCall names, operators) accumulate in
// postfix order.
//
// Example spreader declaration:
//
//	from [Product] to [Country, Product] = Demand
//
// This corresponds to the 'spreaderDeclaration' rule in CPLSpreaders.g4.
func (l *CPLListener) EnterSpreaderDeclaration(ctx *antlr.SpreaderDeclarationContext) {
	l.logEnter("SpreaderDeclaration", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the Spreader scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering Spreader declaration: %w", err))
		return
	}

	// Create a child scope for this spreader definition and push it on the stack
	l.symbolStack.Push(current.AddChild(symboltable.SpreaderScope))
}

// extractSpreaderFromToLists extracts the from and to identifier lists from a
// spreader declaration, handling the case where from [] has no identifierList.
// When both from and to have identifierLists, the first is from and the second
// is to. When only one exists, its position relative to the TO token determines
// which clause it belongs to.
func extractSpreaderFromToLists(ctx *antlr.SpreaderDeclarationContext) (fromList, toList antlr.IIdentifierListContext) {
	allLists := ctx.AllIdentifierList()
	if len(allLists) == 0 {
		return nil, nil
	}
	if len(allLists) == 2 {
		return allLists[0], allLists[1]
	}
	// Only one list - determine if it's from or to by position relative to TO token
	singleList := allLists[0]
	toToken := ctx.TO()
	if toToken == nil {
		return nil, singleList // Shouldn't happen, but be safe
	}
	if singleList.GetStart().GetTokenIndex() < toToken.GetSymbol().GetTokenIndex() {
		return singleList, nil // From list only
	}
	return nil, singleList // To list only
}

// ExitSpreaderDeclaration is called when exiting a spreader declaration.
// It extracts the from concept names, to concept names, creates a
// ParserSpreaderDefinition, computes the synthetic key, stores the definition
// in the SpreaderScope's Data field, and pops the spreader scope.
//
// Symbol Table Driven Parsing:
// The from/to identifiers are extracted directly from the parse tree context.
// The fnCall expression identifiers are accumulated in the SpreaderScope's
// Definitions slice by the root listener's expression handlers (ExitFnCall).
func (l *CPLListener) ExitSpreaderDeclaration(ctx *antlr.SpreaderDeclarationContext) {
	l.logExit("SpreaderDeclaration", &ctx.BaseParserRuleContext)

	// Get the current SpreaderScope before popping
	spreaderScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting Spreader declaration: %w", err))
		return
	}

	// Extract from and to identifier lists using position-based disambiguation
	fromList, toList := extractSpreaderFromToLists(ctx)

	// Extract from concepts from the identifierList after 'from'
	var fromConcepts []string
	fromEmpty := false
	if fromList != nil {
		identifiers := fromList.AllIdentifier()
		fromConcepts = make([]string, 0, len(identifiers))
		for _, ident := range identifiers {
			text := ident.GetText()
			if text == "" {
				l.AddError(fmt.Errorf("spreader declaration has empty source concept in 'from' clause"))
				removeChildSymbolTable(spreaderScope.Parent, spreaderScope)
				_, _ = l.symbolStack.Pop()
				return
			}
			fromConcepts = append(fromConcepts, text)
		}
	}
	if len(fromConcepts) == 0 {
		fromConcepts = nil
		fromEmpty = true
	}

	// Extract to concepts from the identifierList after 'to'
	var toConcepts []string
	if toList != nil {
		identifiers := toList.AllIdentifier()
		toConcepts = make([]string, 0, len(identifiers))
		for _, ident := range identifiers {
			text := ident.GetText()
			if text == "" {
				l.AddError(fmt.Errorf("spreader declaration has empty target concept in 'to' clause"))
				removeChildSymbolTable(spreaderScope.Parent, spreaderScope)
				_, _ = l.symbolStack.Pop()
				return
			}
			toConcepts = append(toConcepts, text)
		}
	}
	if len(toConcepts) == 0 {
		l.AddError(fmt.Errorf("spreader declaration missing target concepts in 'to' clause"))
		removeChildSymbolTable(spreaderScope.Parent, spreaderScope)
		_, _ = l.symbolStack.Pop()
		return
	}

	// Validate that either fnCall or expression exists after '='
	if ctx.FnCall() == nil && ctx.Expression() == nil {
		l.AddError(fmt.Errorf("spreader declaration missing expression after '='"))
		removeChildSymbolTable(spreaderScope.Parent, spreaderScope)
		_, _ = l.symbolStack.Pop()
		return
	}
	// If fnCall is present, validate it has a non-empty function identifier
	if ctx.FnCall() != nil {
		if fnIdent := ctx.FnCall().Identifier(); fnIdent == nil || fnIdent.GetText() == "" {
			l.AddError(fmt.Errorf("spreader declaration has empty function call"))
			removeChildSymbolTable(spreaderScope.Parent, spreaderScope)
			_, _ = l.symbolStack.Pop()
			return
		}
	}

	// Get source location from the start of the declaration
	location := l.makeSourceLocation(ctx.GetStart())

	// Create the ParserSpreaderDefinition
	def := &ParserSpreaderDefinition{
		FromConcepts:   fromConcepts,
		ToConcepts:     toConcepts,
		FromEmpty:      fromEmpty,
		SourceLocation: location,
	}

	// Compute synthetic key
	def.Key = spreaderSyntheticKey(fromConcepts, toConcepts)

	// Store the ParserSpreaderDefinition in the spreader scope's Data field
	spreaderScope.Data = def

	// Also push a SpreaderScopeDefinition entry into the scope's Definitions slice
	keyCopy := def.Key
	_ = spreaderScope.DefineWithData(&keyCopy, symboltable.SpreaderScopeDefinition, location, def)

	// Pop the Spreader scope from the symbol stack
	_, err = l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting Spreader declaration: %w", err))
		return
	}
}
