// Package symboltable provides symbol table data structures for the CPL parser.
//
// This file contains tree walking utilities for traversing the symbol table
// hierarchy during the analysis phase.
package analysis

import (
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// AnalysisFunc is the signature for scope-specific analysis functions.
// It receives the symbol table for the scope being visited and returns
// an error if analysis fails.
type AnalysisFunc func(table *symboltable.SymbolTable, context *AnalysisContext, analyzer *CPLAnalyzer) error

// WalkSymbolTableWithHandlers recursively walks the symbol table tree,
// calling the appropriate analysis function from the handlers map
// for each table based on its ScopeType.
//
// The walk is performed in pre-order (parent before children).
// If a scope type has no registered handler, that scope is skipped
// but its children are still visited.
// The first error returned by any handler stops the walk and is returned.
//
// Parameters:
//   - table: The root symbol table to start walking from. If nil, returns nil immediately.
//   - handlers: A map from ScopeType to analysis functions. Each handler is called
//     when a scope of its registered type is encountered.
//
// Returns:
//   - nil if the walk completes successfully or if table is nil.
//   - The first error returned by any handler, stopping the walk.

func WalkSymbolTableWithHandlers(
	table *symboltable.SymbolTable,
	enterHandlers map[symboltable.ScopeType]AnalysisFunc,
	exitHandlers map[symboltable.ScopeType]AnalysisFunc,
	analyzer *CPLAnalyzer) error {

	context := NewAnalysisContext()

	return WalkSymbolTableWithHandlersInternal(table, enterHandlers, exitHandlers, context, analyzer)
}

func WalkSymbolTableWithHandlersInternal(
	table *symboltable.SymbolTable,
	enterHandlers map[symboltable.ScopeType]AnalysisFunc,
	exitHandlers map[symboltable.ScopeType]AnalysisFunc,
	context *AnalysisContext,
	analyzer *CPLAnalyzer) error {
	if table == nil {
		return nil
	}

	// Call the enter handler for this scope type if one is registered
	if handler, exists := enterHandlers[table.ScopeType]; exists && handler != nil {
		err := handler(table, context, analyzer)

		if err != nil {
			return err
		}
	}

	// Recursively visit all children
	for _, child := range table.Children {
		if err := WalkSymbolTableWithHandlersInternal(child, enterHandlers, exitHandlers, context, analyzer); err != nil {
			return err
		}
	}

	// Call the exit handler for this scope type if one is registered
	if handler, exists := exitHandlers[table.ScopeType]; exists && handler != nil {
		err := handler(table, context, analyzer)

		if err != nil {
			return err
		}
	}

	return nil
}
