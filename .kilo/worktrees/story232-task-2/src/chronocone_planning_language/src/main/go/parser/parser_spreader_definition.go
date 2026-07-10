// Package parser provides CPL parsing functionality.
//
// This file contains the ParserSpreaderDefinition data structure that carries
// parsed spreader information through the pipeline. The struct stores raw
// concept names as strings; all cross-reference resolution is deferred to the
// analysis phase.
package parser

import (
	"fmt"
	"strings"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// ParserSpreaderDefinition represents a single spreader declaration extracted during parsing.
// It stores raw identifier references for later resolution during analysis.
type ParserSpreaderDefinition struct {
	// Key is the synthetic key assigned by the parser to uniquely identify
	// this unnamed spreader declaration.
	Key string

	// FromConcepts is the ordered list of source concept names declared after 'from'.
	// Nil when FromEmpty is true (scalar spreading).
	FromConcepts []string

	// ToConcepts is the ordered list of target concept names declared after 'to'.
	// Slice of target concept names (spreaders expand to multiple concepts).
	ToConcepts []string

	// FromEmpty is true when the from-source is explicitly empty (e.g., 'from []').
	// This indicates scalar spreading.
	FromEmpty bool

	// SourceLocation captures where the spreader declaration begins in source.
	SourceLocation symboltable.SourceLocation
}

// IsDefinitionData marks ParserSpreaderDefinition as implementing symboltable.DefinitionData.
func (ParserSpreaderDefinition) IsDefinitionData() {}

// spreaderSyntheticKey generates a deterministic synthetic key for a
// spreader declaration based on its from and to concepts.
//
// Format: __spreader__[A,B]->[C,D] for multi-concept from with multi-concept to,
//
//	__spreader__[A]->[B,C] for single from/multi to,
//	__spreader__[]->[A] for scalar spreading.
func spreaderSyntheticKey(from []string, to []string) string {
	return fmt.Sprintf("__spreader__[%s]->[%s]", strings.Join(from, ","), strings.Join(to, ","))
}
