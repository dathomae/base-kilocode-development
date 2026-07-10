// Package parser provides CPL parsing functionality.
//
// This file contains listener callbacks related to the CPLConnectors.g4 grammar file,
// which defines the Connectors section and Connector declarations.
//
// Grammar rules covered:
//   - ConnectorsSection, ConnectorDeclaration
//
// Note: The identifier rule is defined in ChronoconePlanningLanguage.g4 and its
// listener methods are in cpl_document_listener.go
//
// Symbol Table Driven Parsing:
// The parser builds a symbol table stack during parsing. Identifiers are added
// to the symbol table by EnterIdentifier (defined in cpl_document_listener.go).
// Detailed extraction and validation of Connector definitions is deferred to the
// analysis phase.
package parser

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// EnterConnectorsSection is called when entering the Connectors section.
// It pushes a new ConnectorsScope symbol table onto the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) EnterConnectorsSection(ctx *antlr.ConnectorsSectionContext) {
	l.logEnter("ConnectorsSection", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the Connectors scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering Connectors section: %w", err))
		return
	}

	connectorsTable := current.AddChild(symboltable.ConnectorsScope)
	l.symbolStack.Push(connectorsTable)
}

// ExitConnectorsSection is called when exiting the Connectors section.
// It pops the Connectors scope symbol table from the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) ExitConnectorsSection(ctx *antlr.ConnectorsSectionContext) {
	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting Connectors section: %w", err))
	}
	l.logExit("ConnectorsSection", &ctx.BaseParserRuleContext)
}

// EnterConnectorDeclaration is called when entering a Connector declaration.
// It pushes a new ConnectorScope onto the symbol stack.
//
// Example Connector declaration: SalesData = dimensions [Store, Product] grain Day celltype SalesMeasure
// This corresponds to the 'ConnectorDeclaration' rule in CPLConnectors.g4.
//
// Note: The Connector identifier and dimension/celltype identifiers are added to the symbol
// table by EnterIdentifier (defined in cpl_document_listener.go) when those
// nodes are visited.
func (l *CPLListener) EnterConnectorDeclaration(ctx *antlr.ConnectorDeclarationContext) {
	l.logEnter("ConnectorDeclaration", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the Connector scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering Connector declaration: %w", err))
		return
	}

	// Create a child scope for this Connector definition and push it on the stack
	l.symbolStack.Push(current.AddChild(symboltable.ConnectorScope))
}

// ExitConnectorDeclaration is called when exiting a Connector declaration.
// It extracts the grain from the parse tree, registers a GrainDefinition
// in the ConnectorScope, and then pops the scope from the symbol stack.
//
// Symbol Table Driven Parsing:
// The identifiers for the Connector definition (connector name, dimensions, grain, celltype) are added
// to the symbol table by EnterIdentifier (defined in cpl_document_listener.go)
// when those nodes are visited. The ConnectorScope persists in the symbol table tree
// as a child of ConnectorsScope.
//
// Grain Extraction:
// The grammar supports multiple grain identifiers via an IdentifierList after the GRAIN keyword.
// ctx.IdentifierList(0) returns the dimensions list, and ctx.IdentifierList(1) returns the grain list.
//
// Detailed extraction and validation of Connector definitions is deferred to the
// analysis phase.
func (l *CPLListener) ExitConnectorDeclaration(ctx *antlr.ConnectorDeclarationContext) {
	l.logExit("ConnectorDeclaration", &ctx.BaseParserRuleContext)

	// Get the current ConnectorScope before popping
	connectorScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting Connector declaration: %w", err))
		return
	}

	// Extract identifiers from grammar-specific contexts. Connector name and celltype
	// are direct children of ConnectorDeclarationContext, while dimensions and grain
	// are nested in IdentifierListContext.
	// Direct identifier order: [0] connector name, [last] celltype.
	allIdentifiers := ctx.AllIdentifier()

	// The grammar allows zero or more connector declarations per
	// ConnectorDeclarationContext (the loop may match zero times, e.g. when
	// only a newline follows the indent). When there are no identifiers we
	// remove the parser-created empty ConnectorScope before popping so the
	// analyzer only visits scopes with structured connector declaration data.
	if len(allIdentifiers) == 0 {
		removeChildSymbolTable(connectorScope.Parent, connectorScope)
		_, err = l.symbolStack.Pop()
		if err != nil {
			l.AddError(fmt.Errorf("symbol stack error exiting Connector declaration: %w", err))
		}
		return
	}

	// A valid connector declaration requires at least 2 direct identifiers:
	// connector name and celltype. Grain is now inside IdentifierList(1).
	if len(allIdentifiers) < 2 {
		l.AddError(fmt.Errorf("connector declaration has insufficient identifiers"))
		removeChildSymbolTable(connectorScope.Parent, connectorScope)
		_, _ = l.symbolStack.Pop()
		return
	}

	connectorName := allIdentifiers[0].GetText()
	connectorLocation := l.makeSourceLocation(allIdentifiers[0].GetStart())

	// Extract grain identifiers from IdentifierList(1).
	grainList := ctx.IdentifierList(1)
	if grainList == nil || len(grainList.AllIdentifier()) == 0 {
		l.AddError(fmt.Errorf("connector declaration missing grain"))
		removeChildSymbolTable(connectorScope.Parent, connectorScope)
		_, _ = l.symbolStack.Pop()
		return
	}

	grainIdentifiers := grainList.AllIdentifier()

	// Build GrainNames and GrainLocations for multi-grain support.
	grainNames := make([]string, 0, len(grainIdentifiers))
	grainLocations := make([]symboltable.SourceLocation, 0, len(grainIdentifiers))
	for _, gi := range grainIdentifiers {
		grainNames = append(grainNames, gi.GetText())
		grainLocations = append(grainLocations, l.makeSourceLocation(gi.GetStart()))
	}

	cellTypeIdentifier := allIdentifiers[len(allIdentifiers)-1]
	cellTypeName := cellTypeIdentifier.GetText()
	cellTypeLocation := l.makeSourceLocation(cellTypeIdentifier.GetStart())

	var dimNames []string
	var dimLocations []symboltable.SourceLocation
	if identifierList := ctx.IdentifierList(0); identifierList != nil {
		dimensionIdentifiers := identifierList.AllIdentifier()
		dimNames = make([]string, 0, len(dimensionIdentifiers))
		dimLocations = make([]symboltable.SourceLocation, 0, len(dimensionIdentifiers))
		for _, dimensionIdentifier := range dimensionIdentifiers {
			dimNames = append(dimNames, dimensionIdentifier.GetText())
			dimLocations = append(dimLocations, l.makeSourceLocation(dimensionIdentifier.GetStart()))
		}
	}

	connectorScope.Data = &symboltable.ConnectorDeclaration{
		Name:               connectorName,
		NameLocation:       connectorLocation,
		DimensionNames:     dimNames,
		DimensionLocations: dimLocations,
		GrainNames:         grainNames,
		GrainLocations:     grainLocations,
		CellTypeName:       cellTypeName,
		CellTypeLocation:   cellTypeLocation,
	}

	// Register each grain as a GrainDefinition in the symbol table
	for i, gi := range grainIdentifiers {
		grainName := gi.GetText()
		_ = connectorScope.Define(&grainName, symboltable.GrainDefinition, grainLocations[i])
	}

	// Pop the Connector scope from the symbol stack
	// The ConnectorScope and its identifier definitions persist in the symbol table tree
	// (as a child of ConnectorsScope). Analysis phase will extract the Connector definition
	// by walking the symbol table.
	_, err = l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting Connector declaration: %w", err))
		return
	}
}

func removeChildSymbolTable(parent *symboltable.SymbolTable, child *symboltable.SymbolTable) {
	if parent == nil || child == nil {
		return
	}

	for i, candidate := range parent.Children {
		if candidate == child {
			parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
			child.Parent = nil
			return
		}
	}
}
