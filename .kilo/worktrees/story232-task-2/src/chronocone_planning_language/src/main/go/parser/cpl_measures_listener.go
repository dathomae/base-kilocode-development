// Package parser provides CPL parsing functionality.
//
// This file contains listener callbacks related to the CPLMeasures.g4 grammar file,
// which defines the Measures section and Measure definitions.
//
// Grammar rules covered:
//   - MeasuresSection, MeasureDeclaration
//
// Note: The identifier rule is defined in ChronoconePlanningLanguage.g4 and its
// listener methods are in cpl_document_listener.go
//
// Symbol Table Driven Parsing:
// The parser builds a symbol table stack during parsing. Identifiers are added
// to the symbol table by EnterIdentifier (defined in cpl_document_listener.go).
// Detailed extraction and validation of Measure definitions is deferred to the
// analysis phase (Story154 and subsequent stories).
package parser

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// EnterMeasuresSection is called when entering the Measures section.
// It pushes a new MeasuresScope symbol table onto the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) EnterMeasuresSection(ctx *antlr.MeasuresSectionContext) {
	l.logEnter("MeasuresSection", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the Measures scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering Measures section: %w", err))
		return
	}

	MeasuresTable := current.AddChild(symboltable.MeasuresScope)
	l.symbolStack.Push(MeasuresTable)
}

// ExitMeasuresSection is called when exiting the Measures section.
// It pops the Measures scope symbol table from the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) ExitMeasuresSection(ctx *antlr.MeasuresSectionContext) {
	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting Measures section: %w", err))
	}
	l.logExit("MeasuresSection", &ctx.BaseParserRuleContext)
}

// EnterMeasureDeclaration is called when entering a Measure declaration.
// It pushes a new MeasureScope onto the symbol stack.
//
// Example Measure declaration: Currency = Country
// This corresponds to the 'MeasureDeclaration' rule in CPLMeasures.g4.
//
// Note: The Measure identifier and concept identifier are added to the symbol
// table by EnterIdentifier (defined in cpl_document_listener.go) when those
// nodes are visited.
func (l *CPLListener) EnterMeasureDeclaration(ctx *antlr.MeasureDeclarationContext) {
	l.logEnter("MeasureDeclaration", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the Measure scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering Measure declaration: %w", err))
		return
	}

	// Create a child scope for this Measure definition and push it on the stack
	l.symbolStack.Push(current.AddChild(symboltable.MeasureScope))
}

// ExitMeasureDeclaration is called when exiting a Measure declaration.
// It prepends the measure name and type name to the MeasureScope's Definitions
// (since MeasureScope is now in shouldPushExpressionDefs, EnterIdentifier no
// longer adds them automatically), then pops the Measure scope from the symbol stack.
//
// Symbol Table Driven Parsing:
// The MeasureScope persists in the symbol table tree as a child of MeasuresScope.
// Detailed extraction and validation of Measure definitions is deferred to the
// analysis phase (handled by AnalyzeMeasureScope in Measures_analysis.go).
func (l *CPLListener) ExitMeasureDeclaration(ctx *antlr.MeasureDeclarationContext) {
	l.logExit("MeasureDeclaration", &ctx.BaseParserRuleContext)

	// Get the current MeasureScope before popping
	measureScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting Measure declaration: %w", err))
		return
	}

	// MeasureScope is now in shouldPushExpressionDefs(), so EnterIdentifier
	// does not add identifiers automatically. We manually prepend measure name
	// and type name so analysis can find them at Definitions[0] and Definitions[1].
	measureNameCtx := ctx.Identifier(0)
	typeNameCtx := ctx.Identifier(1)

	var nameDefs []symboltable.SymbolDefinition
	if measureNameCtx != nil && measureNameCtx.GetStart() != nil {
		name := measureNameCtx.GetText()
		nameDefs = append(nameDefs, symboltable.SymbolDefinition{
			Name:           &name,
			DefinitionType: symboltable.Identifier,
			Location:       l.makeSourceLocation(measureNameCtx.GetStart()),
		})
	}
	if typeNameCtx != nil && typeNameCtx.GetStart() != nil {
		name := typeNameCtx.GetText()
		nameDefs = append(nameDefs, symboltable.SymbolDefinition{
			Name:           &name,
			DefinitionType: symboltable.Identifier,
			Location:       l.makeSourceLocation(typeNameCtx.GetStart()),
		})
	}

	// Prepend name definitions before any postfix expression definitions
	measureScope.Definitions = append(nameDefs, measureScope.Definitions...)

	// Pop the Measure scope from the symbol stack
	_, err = l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting Measure declaration: %w", err))
		return
	}
}
