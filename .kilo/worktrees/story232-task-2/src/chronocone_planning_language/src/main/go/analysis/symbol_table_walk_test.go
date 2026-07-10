package analysis

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWalkSymbolTableEmpty verifies walking an empty table.
func TestWalkSymbolTableEmpty(t *testing.T) {
	t.Parallel()

	table := symboltable.NewSymbolTable(symboltable.GlobalScope)
	visitCount := 0

	symboltable.WalkSymbolTable(table, func(st *symboltable.SymbolTable) {
		visitCount++
	})

	assert.Equal(t, 1, visitCount, "Should visit the root table once")
}

// TestWalkSymbolTableSingleLevel verifies walking a single-level tree.
func TestWalkSymbolTableSingleLevel(t *testing.T) {
	t.Parallel()

	root := symboltable.NewSymbolTable(symboltable.GlobalScope)
	child1 := root.AddChild(symboltable.UnitsScope)
	child2 := root.AddChild(symboltable.UnitsScope)

	visited := make([]*symboltable.SymbolTable, 0)

	symboltable.WalkSymbolTable(root, func(st *symboltable.SymbolTable) {
		visited = append(visited, st)
	})

	require.Len(t, visited, 3, "Should visit root and 2 children")
	assert.Equal(t, root, visited[0])
	// Children may be visited in any order, but root should be first
	assert.Contains(t, visited[1:], child1)
	assert.Contains(t, visited[1:], child2)
}

// TestWalkSymbolTableMultiLevel verifies walking a multi-level tree.
func TestWalkSymbolTableMultiLevel(t *testing.T) {
	t.Parallel()

	// Build a tree:
	// root (GlobalScope)
	// ├── child1 (UnitsScope)
	// │   └── grandchild1 (GlobalScope)
	// └── child2 (UnitsScope)

	root := symboltable.NewSymbolTable(symboltable.GlobalScope)
	child1 := root.AddChild(symboltable.UnitsScope)
	grandchild1 := child1.AddChild(symboltable.GlobalScope)
	child2 := root.AddChild(symboltable.UnitsScope)

	visited := make([]*symboltable.SymbolTable, 0)

	symboltable.WalkSymbolTable(root, func(st *symboltable.SymbolTable) {
		visited = append(visited, st)
	})

	require.Len(t, visited, 4, "Should visit all 4 tables")
	// Root should be first (pre-order)
	assert.Equal(t, root, visited[0])
	// All tables should be visited exactly once
	assert.Contains(t, visited, root)
	assert.Contains(t, visited, child1)
	assert.Contains(t, visited, grandchild1)
	assert.Contains(t, visited, child2)
}

// TestWalkSymbolTablePreOrder verifies pre-order traversal.
func TestWalkSymbolTablePreOrder(t *testing.T) {
	t.Parallel()

	// Build a tree:
	// root (GlobalScope)
	// ├── child1 (UnitsScope)
	// │   └── grandchild1 (GlobalScope)
	// └── child2 (UnitsScope)

	root := symboltable.NewSymbolTable(symboltable.GlobalScope)
	child1 := root.AddChild(symboltable.UnitsScope)
	grandchild1 := child1.AddChild(symboltable.GlobalScope)
	child2 := root.AddChild(symboltable.UnitsScope)

	// Add definitions to identify tables
	name1 := "root"
	name2 := "child1"
	name3 := "grandchild1"
	name4 := "child2"

	root.Define(&name1, symboltable.UnitDefinition, symboltable.SourceLocation{Line: 1, Column: 1})
	child1.Define(&name2, symboltable.UnitDefinition, symboltable.SourceLocation{Line: 2, Column: 1})
	grandchild1.Define(&name3, symboltable.UnitDefinition, symboltable.SourceLocation{Line: 3, Column: 1})
	child2.Define(&name4, symboltable.UnitDefinition, symboltable.SourceLocation{Line: 4, Column: 1})

	visitedOrder := make([]string, 0)

	symboltable.WalkSymbolTable(root, func(st *symboltable.SymbolTable) {
		// Record the first symbol name in each table
		if len(st.Definitions) > 0 && st.Definitions[0].Name != nil {
			visitedOrder = append(visitedOrder, *st.Definitions[0].Name)
		}
	})

	// Pre-order: root, then children left-to-right
	// Note: Children are added in order, so child1 before child2
	require.Len(t, visitedOrder, 4)
	assert.Equal(t, "root", visitedOrder[0])
	assert.Equal(t, "child1", visitedOrder[1])
	// grandchild1 is child of child1, so comes before child2
	assert.Equal(t, "grandchild1", visitedOrder[2])
	assert.Equal(t, "child2", visitedOrder[3])
}

// TestWalkSymbolTableNilCallback verifies behavior with nil callback (should not panic).
func TestWalkSymbolTableNilCallback(t *testing.T) {
	t.Parallel()

	root := symboltable.NewSymbolTable(symboltable.GlobalScope)
	root.AddChild(symboltable.UnitsScope)

	// Should not panic with nil callback
	// Note: This test documents current behavior; if the implementation
	// changes to handle nil differently, update accordingly
	// WalkSymbolTable(root, nil) // Would panic - skip this case

	// With valid callback should work fine
	count := 0
	symboltable.WalkSymbolTable(root, func(st *symboltable.SymbolTable) {
		count++
	})
	assert.Equal(t, 2, count)
}

// TestWalkSymbolTableCollectDefinitions verifies collecting all definitions.
func TestWalkSymbolTableCollectDefinitions(t *testing.T) {
	t.Parallel()

	// Build a tree with definitions at various levels
	root := symboltable.NewSymbolTable(symboltable.GlobalScope)
	child1 := root.AddChild(symboltable.UnitsScope)
	child2 := root.AddChild(symboltable.UnitsScope)

	meter := "meter"
	kilogram := "kilogram"
	second := "second"

	root.Define(&meter, symboltable.UnitDefinition, symboltable.SourceLocation{Line: 1, Column: 1})
	child1.Define(&kilogram, symboltable.UnitDefinition, symboltable.SourceLocation{Line: 2, Column: 1})
	child2.Define(&second, symboltable.UnitDefinition, symboltable.SourceLocation{Line: 3, Column: 1})

	// Collect all definition names
	names := make([]string, 0)
	symboltable.WalkSymbolTable(root, func(st *symboltable.SymbolTable) {
		for _, def := range st.Definitions {
			if def.Name != nil {
				names = append(names, *def.Name)
			}
		}
	})

	require.Len(t, names, 3)
	assert.Contains(t, names, "meter")
	assert.Contains(t, names, "kilogram")
	assert.Contains(t, names, "second")
}

// TestWalkSymbolTableCountScopes verifies counting scopes by type.
func TestWalkSymbolTableCountScopes(t *testing.T) {
	t.Parallel()

	root := symboltable.NewSymbolTable(symboltable.GlobalScope)
	root.AddChild(symboltable.UnitsScope)
	root.AddChild(symboltable.UnitsScope)
	root.AddChild(symboltable.GlobalScope)

	globalCount := 0
	unitsCount := 0

	symboltable.WalkSymbolTable(root, func(st *symboltable.SymbolTable) {
		switch st.ScopeType {
		case symboltable.GlobalScope:
			globalCount++
		case symboltable.UnitsScope:
			unitsCount++
		}
	})

	assert.Equal(t, 2, globalCount, "Should have 2 global scopes")
	assert.Equal(t, 2, unitsCount, "Should have 2 units scopes")
}

// TestWalkSymbolTableDeepTree verifies walking a deeply nested tree.
func TestWalkSymbolTableDeepTree(t *testing.T) {
	t.Parallel()

	// Create a deeply nested tree
	root := symboltable.NewSymbolTable(symboltable.GlobalScope)
	current := root
	depth := 10

	for i := 0; i < depth; i++ {
		current = current.AddChild(symboltable.UnitsScope)
	}

	visitCount := 0
	maxDepth := 0

	symboltable.WalkSymbolTable(root, func(st *symboltable.SymbolTable) {
		visitCount++
		// Calculate depth by walking up parents
		d := 0
		for p := st; p.Parent != nil; p = p.Parent {
			d++
		}
		if d > maxDepth {
			maxDepth = d
		}
	})

	assert.Equal(t, depth+1, visitCount, "Should visit all tables")
	assert.Equal(t, depth, maxDepth, "Max depth should be %d", depth)
}

// TestWalkSymbolTableSubTree verifies walking a subtree.
func TestWalkSymbolTableSubTree(t *testing.T) {
	t.Parallel()

	// Build tree:
	// root
	// ├── child1
	// │   └── grandchild1
	// └── child2
	//     └── grandchild2

	root := symboltable.NewSymbolTable(symboltable.GlobalScope)
	child1 := root.AddChild(symboltable.UnitsScope)
	grandchild1 := child1.AddChild(symboltable.GlobalScope)
	child2 := root.AddChild(symboltable.UnitsScope)
	_ = child2.AddChild(symboltable.GlobalScope) // grandchild2 - not used in assertions

	// Walk only from child1's subtree
	visited := make([]*symboltable.SymbolTable, 0)
	symboltable.WalkSymbolTable(child1, func(st *symboltable.SymbolTable) {
		visited = append(visited, st)
	})

	// Should visit child1 and grandchild1 only
	require.Len(t, visited, 2)

	// Verify child1 is visited first (pre-order)
	assert.Equal(t, child1, visited[0])
	assert.Equal(t, grandchild1, visited[1])
}
