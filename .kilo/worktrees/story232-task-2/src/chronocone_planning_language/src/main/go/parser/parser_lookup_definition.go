// Package parser provides CPL parsing functionality.
//
// This file contains the ParserLookupDefinition data structure that carries
// parsed lookup information through the pipeline. The struct stores raw
// concept and measure names as strings; all cross-reference resolution is
// deferred to the analysis phase (Story242).
//
// Design: ERO006 - Lookups Parsing Phase
package parser

import "github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"

// ParserLookupDefinition represents a single lookup declaration extracted during parsing.
// It stores raw identifier references for later resolution during analysis.
type ParserLookupDefinition struct {
	Name           string                     // Lookup identifier
	Dimensions     []string                   // Raw concept names (no grain in Increment 1)
	Celltype       string                     // Raw measure name
	SourceLocation symboltable.SourceLocation // For error reporting
}

// IsDefinitionData marks ParserLookupDefinition as valid DefinitionData
// for storage in SymbolDefinition.Data during parsing.
func (*ParserLookupDefinition) IsDefinitionData() {}
