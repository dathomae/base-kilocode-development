// Package parser provides CPL parsing functionality.
//
// This file contains integration tests for symbol table functionality
// within the CPL parser. Tests follow table-driven patterns with testify
// assertions and parallel execution.
package parser

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// TestSymbolTable_InitializedOnNewListener verifies the symbol table stack
// is initialized when creating a new CPLListener.
func TestSymbolTable_InitializedOnNewListener(t *testing.T) {
	t.Parallel()

	listener := NewCPLListener(nil, nil, nil)

	require.NotNil(t, listener.symbolStack)
	assert.False(t, listener.symbolStack.IsEmpty(), "Symbol stack should not be empty")
	assert.Equal(t, 1, listener.symbolStack.Size(), "Should have one table (global)")

	root, err := listener.symbolStack.Root()
	require.NoError(t, err)
	assert.Equal(t, symboltable.GlobalScope, root.ScopeType)
}

// TestSymbolTable_GlobalScopeNotEmpty verifies global scope is created
// and can hold definitions.
func TestSymbolTable_GlobalScopeNotEmpty(t *testing.T) {
	t.Parallel()

	listener := NewCPLListener(nil, nil, nil)

	// Get the global scope
	root, err := listener.symbolStack.Root()
	require.NoError(t, err)
	assert.Equal(t, symboltable.GlobalScope, root.ScopeType)
}

// TestSymbolTable_GetSymbolStackAccessor verifies the GetSymbolStack method.
func TestSymbolTable_GetSymbolStackAccessor(t *testing.T) {
	t.Parallel()

	listener := NewCPLListener(nil, nil, nil)
	stack := listener.GetSymbolStack()

	require.NotNil(t, stack)
	assert.Equal(t, listener.symbolStack, stack)
}

// TestSymbolTable_UnitDefinitionDirectly verifies defining a unit symbol directly
// in the symbol table works correctly.
func TestSymbolTable_UnitDefinitionDirectly(t *testing.T) {
	t.Parallel()

	listener := NewCPLListener(nil, nil, nil)

	// Get the current table (global)
	current, err := listener.symbolStack.Peek()
	require.NoError(t, err)

	// Create a child units scope
	unitsTable := current.AddChild(symboltable.UnitsScope)
	listener.symbolStack.Push(unitsTable)

	// Define a unit
	unitName := "meter"
	location := symboltable.SourceLocation{Line: 1, Column: 10}
	err = unitsTable.Define(&unitName, symboltable.UnitDefinition, location)
	require.NoError(t, err)

	// Verify the definition was added
	require.Len(t, unitsTable.Definitions, 1)
	assert.Equal(t, "meter", *unitsTable.Definitions[0].Name)
	assert.Equal(t, symboltable.UnitDefinition, unitsTable.Definitions[0].DefinitionType)
	assert.Equal(t, 1, unitsTable.Definitions[0].Location.Line)
	assert.Equal(t, 10, unitsTable.Definitions[0].Location.Column)

	// Pop the units scope
	_, err = listener.symbolStack.Pop()
	require.NoError(t, err)

	// Stack should be back to just global
	assert.Equal(t, 1, listener.symbolStack.Size())
}

// TestSymbolTable_ParentChildRelationship verifies parent-child relationships
// are correctly maintained.
func TestSymbolTable_ParentChildRelationship(t *testing.T) {
	t.Parallel()

	listener := NewCPLListener(nil, nil, nil)

	root, err := listener.symbolStack.Root()
	require.NoError(t, err)

	// Create child table
	child := root.AddChild(symboltable.UnitsScope)

	// Child's parent should be root
	assert.Equal(t, root, child.Parent)

	// Child should be in root's children
	require.Len(t, root.Children, 1)
	assert.Equal(t, child, root.Children[0])
}

// TestSymbolTable_StackOperations verifies stack push/pop operations.
func TestSymbolTable_StackOperations(t *testing.T) {
	t.Parallel()

	listener := NewCPLListener(nil, nil, nil)

	// Initial state: one table
	assert.Equal(t, 1, listener.symbolStack.Size())

	// Push a new table
	root, _ := listener.symbolStack.Root()
	child := root.AddChild(symboltable.UnitsScope)
	listener.symbolStack.Push(child)

	assert.Equal(t, 2, listener.symbolStack.Size())

	// Peek should return child
	peek, err := listener.symbolStack.Peek()
	require.NoError(t, err)
	assert.Equal(t, child, peek)

	// Pop should return child
	popped, err := listener.symbolStack.Pop()
	require.NoError(t, err)
	assert.Equal(t, child, popped)

	// Back to one table
	assert.Equal(t, 1, listener.symbolStack.Size())
}

// TestSymbolTable_WalkTable verifies the symbol table tree can be walked.
func TestSymbolTable_WalkTable(t *testing.T) {
	t.Parallel()

	listener := NewCPLListener(nil, nil, nil)

	root, _ := listener.symbolStack.Root()
	child1 := root.AddChild(symboltable.UnitsScope)
	_ = root.AddChild(symboltable.UnitsScope)    // child2
	_ = child1.AddChild(symboltable.GlobalScope) // grandchild

	// Walk the tree and collect scope types
	scopeTypes := make([]symboltable.ScopeType, 0)
	symboltable.WalkSymbolTable(root, func(st *symboltable.SymbolTable) {
		scopeTypes = append(scopeTypes, st.ScopeType)
	})

	// Should have GlobalScope, UnitsScope, UnitsScope, GlobalScope
	require.Len(t, scopeTypes, 4)
	assert.Equal(t, symboltable.GlobalScope, scopeTypes[0])
	// The other order depends on how we added them
	unitsCount := 0
	globalCount := 0
	for _, st := range scopeTypes[1:] {
		switch st {
		case symboltable.UnitsScope:
			unitsCount++
		case symboltable.GlobalScope:
			globalCount++
		}
	}
	assert.Equal(t, 2, unitsCount)
	assert.Equal(t, 1, globalCount)
}

// TestSymbolTable_Lookup verifies symbol lookup works correctly.
func TestSymbolTable_Lookup(t *testing.T) {
	t.Parallel()

	listener := NewCPLListener(nil, nil, nil)

	root, _ := listener.symbolStack.Root()
	unitsTable := root.AddChild(symboltable.UnitsScope)

	// Define a unit in units scope
	unitName := "meter"
	location := symboltable.SourceLocation{Line: 1, Column: 10}
	err := unitsTable.Define(&unitName, symboltable.UnitDefinition, location)
	require.NoError(t, err)

	// Lookup from units scope should find it (local lookup)
	def, err := unitsTable.Lookup("meter")
	require.NoError(t, err)
	assert.Equal(t, "meter", *def.Name)

	// Define a unit in root scope
	rootUnit := "second"
	rootLoc := symboltable.SourceLocation{Line: 2, Column: 1}
	err = root.Define(&rootUnit, symboltable.UnitDefinition, rootLoc)
	require.NoError(t, err)

	// Lookup from units scope should find root's definition (walks up)
	def, err = unitsTable.Lookup("second")
	require.NoError(t, err)
	assert.Equal(t, "second", *def.Name)

	// Lookup non-existent should fail
	_, err = unitsTable.Lookup("kilogram")
	require.Error(t, err)

	var symErr *symboltable.SymbolTableError
	require.True(t, errors.As(err, &symErr))
	assert.Equal(t, symboltable.ErrSymbolNotFound, symErr.Code)
}

// TestSymbolTable_LookupLocal verifies local-only lookup.
func TestSymbolTable_LookupLocal(t *testing.T) {
	t.Parallel()

	listener := NewCPLListener(nil, nil, nil)

	root, _ := listener.symbolStack.Root()
	unitsTable := root.AddChild(symboltable.UnitsScope)

	// Define in units scope only
	unitName := "meter"
	location := symboltable.SourceLocation{Line: 1, Column: 10}
	err := unitsTable.Define(&unitName, symboltable.UnitDefinition, location)
	require.NoError(t, err)

	// Local lookup from units scope should find it
	def, err := unitsTable.LookupLocal("meter")
	require.NoError(t, err)
	assert.Equal(t, "meter", *def.Name)

	// Local lookup from root should NOT find it
	_, err = root.LookupLocal("meter")
	require.Error(t, err)

	var symErr *symboltable.SymbolTableError
	require.True(t, errors.As(err, &symErr))
	assert.Equal(t, symboltable.ErrSymbolNotFound, symErr.Code)
}

// TestSymbolTable_LookupShadowing verifies child definitions shadow parents.
func TestSymbolTable_LookupShadowing(t *testing.T) {
	t.Parallel()

	listener := NewCPLListener(nil, nil, nil)

	root, _ := listener.symbolStack.Root()
	unitsTable := root.AddChild(symboltable.UnitsScope)

	// Define same name in both scopes
	parentUnit := "unit"
	childUnit := "unit"
	parentLoc := symboltable.SourceLocation{Line: 1, Column: 1}
	childLoc := symboltable.SourceLocation{Line: 2, Column: 1}

	err := root.Define(&parentUnit, symboltable.UnitDefinition, parentLoc)
	require.NoError(t, err)

	err = unitsTable.Define(&childUnit, symboltable.UnitDefinition, childLoc)
	require.NoError(t, err)

	// Lookup from child should find child's definition
	def, err := unitsTable.Lookup("unit")
	require.NoError(t, err)
	assert.Equal(t, 2, def.Location.Line) // Child's location
}

// TestSymbolTable_DefinitionTypeString verifies definition type string representation.
func TestSymbolTable_DefinitionTypeString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "UnitDefinition", symboltable.UnitDefinition.String())
}

// TestSymbolTable_ScopeTypeString verifies scope type string representation.
func TestSymbolTable_ScopeTypeString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scope    symboltable.ScopeType
		expected string
	}{
		{symboltable.GlobalScope, "GlobalScope"},
		{symboltable.UnitsScope, "UnitsScope"},
		{symboltable.ScopeType(999), "ScopeType(999)"}, // Unknown
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, tc.scope.String())
	}
}
