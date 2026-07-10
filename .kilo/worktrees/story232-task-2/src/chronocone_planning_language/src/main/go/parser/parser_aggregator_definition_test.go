// Package parser provides CPL parsing functionality.
//
// This file contains unit tests for ParserAggregatorDefinition and the
// aggregatorSyntheticKey helper.
package parser

import (
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParserAggregatorDefinitionInitialization verifies that
// ParserAggregatorDefinition can be initialized with all fields populated.
func TestParserAggregatorDefinitionInitialization(t *testing.T) {
	t.Parallel()

	loc := symboltable.SourceLocation{
		File:   "test.cpl",
		Line:   42,
		Column: 8,
	}

	def := ParserAggregatorDefinition{
		FromConcepts:   []string{"Country", "State"},
		ToConcept:      "City",
		ToEmpty:        false,
		SourceLocation: loc,
	}

	require.NotNil(t, def.FromConcepts)
	assert.Equal(t, []string{"Country", "State"}, def.FromConcepts)
	assert.Equal(t, "City", def.ToConcept)
	assert.False(t, def.ToEmpty)
	assert.Equal(t, loc, def.SourceLocation)
}

// TestAggregatorSyntheticKey verifies the synthetic key generation for
// various aggregator patterns.
func TestAggregatorSyntheticKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		from     []string
		to       string
		expected string
	}{
		{
			name:     "multi-concept from with single to",
			from:     []string{"A", "B"},
			to:       "C",
			expected: "__aggregator__[A,B]->[C]",
		},
		{
			name:     "single-concept from with single to",
			from:     []string{"X"},
			to:       "Y",
			expected: "__aggregator__[X]->[Y]",
		},
		{
			name:     "scalar reduction (to empty)",
			from:     []string{"A", "B"},
			to:       "",
			expected: "__aggregator__[A,B]->[]",
		},
		{
			name:     "empty from list with to",
			from:     []string{},
			to:       "C",
			expected: "__aggregator__[]->[C]",
		},
		{
			name:     "empty from list scalar reduction",
			from:     []string{},
			to:       "",
			expected: "__aggregator__[]->[]",
		},
		{
			name:     "three-concept from with to",
			from:     []string{"A", "B", "C"},
			to:       "D",
			expected: "__aggregator__[A,B,C]->[D]",
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := aggregatorSyntheticKey(tc.from, tc.to)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestAggregatorSyntheticKeyPrefix verifies that the __aggregator__ prefix
// is always present in generated keys.
func TestAggregatorSyntheticKeyPrefix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		from []string
		to   string
	}{
		{
			name: "multi-concept",
			from: []string{"A", "B"},
			to:   "C",
		},
		{
			name: "single-concept",
			from: []string{"X"},
			to:   "Y",
		},
		{
			name: "scalar reduction",
			from: []string{"A"},
			to:   "",
		},
		{
			name: "empty from",
			from: []string{},
			to:   "Z",
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := aggregatorSyntheticKey(tc.from, tc.to)
			assert.True(t, strings.HasPrefix(result, "__aggregator__"),
				"Expected key %q to start with __aggregator__", result)
		})
	}
}

// TestParserAggregatorDefinitionWithToEmpty verifies that a scalar reduction
// definition sets ToEmpty to true and leaves ToConcept empty.
func TestParserAggregatorDefinitionWithToEmpty(t *testing.T) {
	t.Parallel()

	loc := symboltable.SourceLocation{
		File:   "aggregators.cpl",
		Line:   10,
		Column: 4,
	}

	def := ParserAggregatorDefinition{
		FromConcepts:   []string{"Revenue"},
		ToConcept:      "",
		ToEmpty:        true,
		SourceLocation: loc,
	}

	assert.Empty(t, def.ToConcept)
	assert.True(t, def.ToEmpty)
	assert.Equal(t, loc, def.SourceLocation)
}
