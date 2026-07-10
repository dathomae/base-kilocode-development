// Package parser provides CPL parsing functionality.
//
// This file contains listener callbacks related to the CPLTypes.g4 grammar file,
// which defines the types section and type definitions.
//
// Grammar rules covered:
//   - typesSection, typeDefinition, baseType, constraintList, constraint, constraintValue
//
// Note: The identifier rule is defined in ChronoconePlanningLanguage.g4 and its
// listener methods are in chronocone_planning_language_listener.go
//
// Symbol Table Driven Parsing:
// The parser builds a symbol table stack during parsing. Identifiers are added
// to the symbol table by EnterIdentifier (defined in cpl_document_listener.go).
// Detailed extraction and validation of type definitions is deferred to the
// analysis phase (Story132 and subsequent stories).
package parser

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// EnterTypesSection is called when entering the types section.
// It pushes a new TypesScope symbol table onto the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) EnterTypesSection(ctx *antlr.TypesSectionContext) {
	l.logEnter("TypesSection", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the types scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering types section: %w", err))
		return
	}

	typesTable := current.AddChild(symboltable.TypesScope)
	l.symbolStack.Push(typesTable)
}

// ExitTypesSection is called when exiting the types section.
// It pops the types scope symbol table from the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) ExitTypesSection(ctx *antlr.TypesSectionContext) {
	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting types section: %w", err))
	}
	l.logExit("TypesSection", &ctx.BaseParserRuleContext)
}

// EnterTypeDefinition is called when entering a type definition.
// It pushes a new TypeScope onto the symbol stack.
//
// Example type definition: Currency = Number(precision=2)
// This corresponds to the 'typeDefinition' rule in CPLTypes.g4.
//
// Note: The type identifier and base type identifier are added to the symbol
// table by EnterIdentifier (defined in cpl_document_listener.go) when those
// nodes are visited. Detailed extraction of type definition data is deferred
// to the analysis phase.
func (l *CPLListener) EnterTypeDefinition(ctx *antlr.TypeDefinitionContext) {
	l.logEnter("TypeDefinition", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the type scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering type definition: %w", err))
		return
	}

	// Create a child scope for this type definition and push it on the stack
	// Note: TypeScope is used for the type's internal scope (similar to ConceptScope)
	l.symbolStack.Push(current.AddChild(symboltable.TypeScope))
}

// ExitTypeDefinition is called when exiting a type definition.
// It pops the type scope from the symbol stack.
//
// Note: The type identifier and base type identifier are added to the TypeScope
// by EnterIdentifier when those nodes are visited. The TypeScope and its
// definitions persist in the symbol table tree (as a child of TypesScope).
// The analysis phase will walk the tree and extract type definitions from
// the TypeScope children.
func (l *CPLListener) ExitTypeDefinition(ctx *antlr.TypeDefinitionContext) {
	// Pop the type scope from the symbol stack
	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting type definition: %w", err))
	}
	l.logExit("TypeDefinition", &ctx.BaseParserRuleContext)
}

// EnterBaseType is called when entering a base type.
//
// Note: The identifier is matched in the baseType rule, and EnterIdentifier
// (defined in cpl_document_listener.go) will add it to the symbol table.
// Base type interpretation (built-in vs declared) is deferred to the analysis phase.
func (l *CPLListener) EnterBaseType(ctx *antlr.BaseTypeContext) {
	l.logEnter("BaseType", &ctx.BaseParserRuleContext)
}

// ExitBaseType is called when exiting a base type.
// The base type identifier is added to the symbol table by EnterIdentifier.
func (l *CPLListener) ExitBaseType(ctx *antlr.BaseTypeContext) {
	l.logExit("BaseType", &ctx.BaseParserRuleContext)
}
