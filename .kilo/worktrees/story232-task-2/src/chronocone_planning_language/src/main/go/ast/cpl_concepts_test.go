// Package ast_test contains tests for the CPL AST types.
package ast_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/ast"
	"github.com/stretchr/testify/assert"
)

// TestValuesGeneratorDef tests creating Values generator definitions.
func TestValuesGeneratorDef(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		values         []string
		expectedLength int
		expectedFirst  string
	}{
		{
			name:           "single value",
			values:         []string{"Active"},
			expectedLength: 1,
			expectedFirst:  "Active",
		},
		{
			name:           "multiple values",
			values:         []string{"USA", "Canada", "Mexico"},
			expectedLength: 3,
			expectedFirst:  "USA",
		},
		{
			name:           "many values",
			values:         []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"},
			expectedLength: 6,
			expectedFirst:  "Jan",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gen := ast.ValuesGeneratorDef{Values: tc.values}

			assert.Len(t, gen.Values, tc.expectedLength, "Values count should match")
			if tc.expectedLength > 0 {
				assert.Equal(t, tc.expectedFirst, gen.Values[0], "First value should match")
			}
		})
	}
}

// TestSourceLocation tests source location struct.
func TestSourceLocation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		location ast.SourceLocation
	}{
		{
			name:     "line 1, column 0",
			location: ast.SourceLocation{Line: 1, Column: 0},
		},
		{
			name:     "line 5, column 10",
			location: ast.SourceLocation{Line: 5, Column: 10},
		},
		{
			name:     "line 100, column 50",
			location: ast.SourceLocation{Line: 100, Column: 50},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.location.Line, tc.location.Line, "Line should match")
			assert.Equal(t, tc.location.Column, tc.location.Column, "Column should match")
		})
	}
}

// TestInsertionPoint_BeforeOnly verifies creating an InsertionPoint with only BeforeValue set.
func TestInsertionPoint_BeforeOnly(t *testing.T) {
	t.Parallel()

	ip := ast.InsertionPoint{BeforeValue: "Low"}

	assert.Equal(t, "Low", ip.BeforeValue, "BeforeValue should match")
	assert.Equal(t, "", ip.AfterValue, "AfterValue should be empty")
}

// TestInsertionPoint_AfterOnly verifies creating an InsertionPoint with only AfterValue set.
func TestInsertionPoint_AfterOnly(t *testing.T) {
	t.Parallel()

	ip := ast.InsertionPoint{AfterValue: "Medium"}

	assert.Equal(t, "Medium", ip.AfterValue, "AfterValue should match")
	assert.Equal(t, "", ip.BeforeValue, "BeforeValue should be empty")
}

// TestInsertionPoint_BothFields verifies creating an InsertionPoint with both BeforeValue and AfterValue set.
func TestInsertionPoint_BothFields(t *testing.T) {
	t.Parallel()

	ip := ast.InsertionPoint{BeforeValue: "Low", AfterValue: "High"}

	assert.Equal(t, "Low", ip.BeforeValue, "BeforeValue should match")
	assert.Equal(t, "High", ip.AfterValue, "AfterValue should match")
}

// TestInsertionPoint_ZeroValue verifies the zero value of InsertionPoint has empty strings for both fields.
func TestInsertionPoint_ZeroValue(t *testing.T) {
	t.Parallel()

	ip := ast.InsertionPoint{}

	assert.Equal(t, "", ip.BeforeValue, "BeforeValue should be empty")
	assert.Equal(t, "", ip.AfterValue, "AfterValue should be empty")
}

// TestConceptDefinitionData_InitialDefinition verifies a ConceptDefinitionData with zero-value InsertionPoint.
func TestConceptDefinitionData_InitialDefinition(t *testing.T) {
	t.Parallel()

	cd := ast.ConceptDefinitionData{Name: "Priority"}

	assert.Equal(t, "Priority", cd.Name, "Name should match")
	assert.Equal(t, "", cd.InsertionPoint.BeforeValue, "BeforeValue should be empty for initial definition")
	assert.Equal(t, "", cd.InsertionPoint.AfterValue, "AfterValue should be empty for initial definition")
}

// TestConceptDefinitionData_ExtensionWithBefore verifies a ConceptDefinitionData with BeforeValue set.
func TestConceptDefinitionData_ExtensionWithBefore(t *testing.T) {
	t.Parallel()

	cd := ast.ConceptDefinitionData{
		Name:           "Priority",
		InsertionPoint: ast.InsertionPoint{BeforeValue: "Low"},
	}

	assert.Equal(t, "Priority", cd.Name, "Name should match")
	assert.Equal(t, "Low", cd.InsertionPoint.BeforeValue, "BeforeValue should match")
	assert.Equal(t, "", cd.InsertionPoint.AfterValue, "AfterValue should be empty")
}

// TestConceptDefinitionData_ExtensionWithAfter verifies a ConceptDefinitionData with AfterValue set.
func TestConceptDefinitionData_ExtensionWithAfter(t *testing.T) {
	t.Parallel()

	cd := ast.ConceptDefinitionData{
		Name:           "Priority",
		InsertionPoint: ast.InsertionPoint{AfterValue: "Medium"},
	}

	assert.Equal(t, "Priority", cd.Name, "Name should match")
	assert.Equal(t, "Medium", cd.InsertionPoint.AfterValue, "AfterValue should match")
	assert.Equal(t, "", cd.InsertionPoint.BeforeValue, "BeforeValue should be empty")
}

// TestConceptDefinitionData_SourceLocation verifies SourceLocation fields are accessible.
func TestConceptDefinitionData_SourceLocation(t *testing.T) {
	t.Parallel()

	cd := ast.ConceptDefinitionData{
		Name:           "Priority",
		SourceLocation: ast.SourceLocation{Line: 42, Column: 7},
	}

	assert.Equal(t, 42, cd.SourceLocation.Line, "Line should match")
	assert.Equal(t, 7, cd.SourceLocation.Column, "Column should match")
}
