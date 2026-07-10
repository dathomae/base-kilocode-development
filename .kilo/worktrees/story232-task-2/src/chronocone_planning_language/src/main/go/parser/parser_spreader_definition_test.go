// Package parser provides CPL parsing functionality.
//
// This file contains unit tests for ParserSpreaderDefinition and the
// spreaderSyntheticKey helper.
package parser

import (
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParserSpreaderDefinitionInitialization verifies that
// ParserSpreaderDefinition can be initialized with all fields populated.
func TestParserSpreaderDefinitionInitialization(t *testing.T) {
	t.Parallel()

	loc := symboltable.SourceLocation{
		File:   "test.cpl",
		Line:   42,
		Column: 8,
	}

	def := ParserSpreaderDefinition{
		Key:            "__spreader__[Product]->[Country,Product]",
		FromConcepts:   []string{"Product"},
		ToConcepts:     []string{"Country", "Product"},
		FromEmpty:      false,
		SourceLocation: loc,
	}

	require.NotNil(t, def.FromConcepts)
	assert.Equal(t, "__spreader__[Product]->[Country,Product]", def.Key)
	assert.Equal(t, []string{"Product"}, def.FromConcepts)
	assert.Equal(t, []string{"Country", "Product"}, def.ToConcepts)
	assert.False(t, def.FromEmpty)
	assert.Equal(t, loc, def.SourceLocation)
}

// TestSpreaderSyntheticKey verifies the synthetic key generation for
// various spreader patterns.
func TestSpreaderSyntheticKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		from     []string
		to       []string
		expected string
	}{
		{
			name:     "multi-concept from with multi-concept to",
			from:     []string{"Product"},
			to:       []string{"Country", "Product"},
			expected: "__spreader__[Product]->[Country,Product]",
		},
		{
			name:     "single-concept from with single-concept to",
			from:     []string{"X"},
			to:       []string{"Y"},
			expected: "__spreader__[X]->[Y]",
		},
		{
			name:     "scalar spreading (from empty to single)",
			from:     []string{},
			to:       []string{"Country"},
			expected: "__spreader__[]->[Country]",
		},
		{
			name:     "both empty",
			from:     []string{},
			to:       []string{},
			expected: "__spreader__[]->[]",
		},
		{
			name:     "multi-concept from with single-concept to",
			from:     []string{"A", "B", "C"},
			to:       []string{"D"},
			expected: "__spreader__[A,B,C]->[D]",
		},
		{
			name:     "single-concept from with multi-concept to",
			from:     []string{"A"},
			to:       []string{"B", "C", "D"},
			expected: "__spreader__[A]->[B,C,D]",
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := spreaderSyntheticKey(tc.from, tc.to)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestSpreaderSyntheticKeyPrefix verifies that the __spreader__ prefix
// is always present in generated keys.
func TestSpreaderSyntheticKeyPrefix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		from []string
		to   []string
	}{
		{
			name: "multi-concept from with multi-concept to",
			from: []string{"Product"},
			to:   []string{"Country", "Product"},
		},
		{
			name: "single-concept from with single-concept to",
			from: []string{"X"},
			to:   []string{"Y"},
		},
		{
			name: "scalar spreading",
			from: []string{},
			to:   []string{"Country"},
		},
		{
			name: "both empty",
			from: []string{},
			to:   []string{},
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := spreaderSyntheticKey(tc.from, tc.to)
			assert.True(t, strings.HasPrefix(result, "__spreader__"),
				"Expected key %q to start with __spreader__", result)
		})
	}
}

// TestParserSpreaderDefinitionWithFromEmpty verifies that a scalar spreading
// definition sets FromEmpty to true and leaves FromConcepts empty.
func TestParserSpreaderDefinitionWithFromEmpty(t *testing.T) {
	t.Parallel()

	loc := symboltable.SourceLocation{
		File:   "spreaders.cpl",
		Line:   10,
		Column: 4,
	}

	def := ParserSpreaderDefinition{
		FromConcepts:   []string{},
		ToConcepts:     []string{"Country"},
		FromEmpty:      true,
		SourceLocation: loc,
	}

	assert.Empty(t, def.FromConcepts)
	assert.Equal(t, []string{"Country"}, def.ToConcepts)
	assert.True(t, def.FromEmpty)
	assert.Equal(t, loc, def.SourceLocation)
}
