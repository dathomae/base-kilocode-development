// Package parser provides CPL parsing functionality.
//
// This file contains the ParserAggregatorDefinition data structure that carries
// parsed aggregator information through the pipeline. The struct stores raw
// concept names as strings; all cross-reference resolution is deferred to the
// analysis phase (Story256).
package parser

import (
	"fmt"
	"strings"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// ParserAggregatorDefinition represents a single aggregator declaration extracted during parsing.
// It stores raw identifier references for later resolution during analysis.
type ParserAggregatorDefinition struct {
	// Key is the synthetic key assigned by the parser to uniquely identify
	// this unnamed aggregator declaration.
	Key string

	// FromConcepts is the ordered list of source concept names declared after 'from'.
	FromConcepts []string

	// ToConcept is the target concept name declared after 'to'.
	// Empty when ToEmpty is true (scalar reduction).
	ToConcept string

	// ToEmpty is true when the to-target is explicitly empty (e.g., 'to []').
	ToEmpty bool

	// SourceLocation captures where the aggregator declaration begins in source.
	SourceLocation symboltable.SourceLocation
}

// IsDefinitionData marks ParserAggregatorDefinition as implementing symboltable.DefinitionData.
func (ParserAggregatorDefinition) IsDefinitionData() {}

// aggregatorSyntheticKey generates a deterministic synthetic key for an
// aggregator declaration based on its from and to concepts.
//
// Format: __aggregator__[A,B]->[C] for multi-concept from with target,
//
//	__aggregator__[A]->[B] for single from/to,
//	__aggregator__[A]->[] for scalar reduction.
func aggregatorSyntheticKey(from []string, to string) string {
	return fmt.Sprintf("__aggregator__[%s]->[%s]", strings.Join(from, ","), to)
}
