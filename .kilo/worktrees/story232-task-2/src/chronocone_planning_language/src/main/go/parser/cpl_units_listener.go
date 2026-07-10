// Package parser provides CPL parsing functionality.
//
// This file contains listener callbacks related to the CPLUnits.g4 grammar file,
// which defines the units section and unit definitions.
//
// Grammar rules covered:
//   - unitsSection, unitDeclaration
//
// Note: The identifier rule is defined in ChronoconePlanningLanguage.g4 and its
// listener methods are in cpl_document_listener.go
//
// Symbol Table Driven Parsing:
// The parser builds a symbol table stack during parsing. Identifiers are added
// to the symbol table by EnterIdentifier (defined in cpl_document_listener.go).
// Detailed extraction and validation of unit definitions is deferred to the
// analysis phase (Story154 and subsequent stories).
package parser

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// EnterUnitsSection is called when entering the units section.
// It pushes a new UnitsScope symbol table onto the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) EnterUnitsSection(ctx *antlr.UnitsSectionContext) {
	l.logEnter("UnitsSection", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the units scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering units section: %w", err))
		return
	}

	unitsTable := current.AddChild(symboltable.UnitsScope)
	l.symbolStack.Push(unitsTable)
}

// ExitUnitsSection is called when exiting the units section.
// It pops the units scope symbol table from the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) ExitUnitsSection(ctx *antlr.UnitsSectionContext) {
	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting units section: %w", err))
	}
	l.logExit("UnitsSection", &ctx.BaseParserRuleContext)
}

// EnterUnitDeclaration is called when entering a unit declaration.
// It pushes a new UnitScope onto the symbol stack.
//
// Example unit declaration: Currency = Country
// This corresponds to the 'unitDeclaration' rule in CPLUnits.g4.
//
// Note: The unit identifier and concept identifier are added to the symbol
// table by EnterIdentifier (defined in cpl_document_listener.go) when those
// nodes are visited.
func (l *CPLListener) EnterUnitDeclaration(ctx *antlr.UnitDeclarationContext) {
	l.logEnter("UnitDeclaration", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the unit scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering unit declaration: %w", err))
		return
	}

	// Create a child scope for this unit definition and push it on the stack
	l.symbolStack.Push(current.AddChild(symboltable.UnitScope))
}

// ExitUnitDeclaration is called when exiting a unit declaration.
// It pops the unit scope from the symbol stack.
//
// Symbol Table Driven Parsing:
// The identifiers for the unit definition (unit_name and concept_name) are added
// to the symbol table by EnterIdentifier (defined in cpl_document_listener.go)
// when those nodes are visited. The UnitScope persists in the symbol table tree
// as a child of UnitsScope.
//
// Detailed extraction and validation of unit definitions is deferred to the
// analysis phase (handled by AnalyzeUnitScope in units_analysis.go).
func (l *CPLListener) ExitUnitDeclaration(ctx *antlr.UnitDeclarationContext) {
	l.logExit("UnitDeclaration", &ctx.BaseParserRuleContext)

	// Pop the unit scope from the symbol stack
	// The UnitScope and its identifier definitions persist in the symbol table tree
	// (as a child of UnitsScope). Analysis phase will extract the unit definition
	// by walking the symbol table.
	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting unit declaration: %w", err))
		return
	}
}
