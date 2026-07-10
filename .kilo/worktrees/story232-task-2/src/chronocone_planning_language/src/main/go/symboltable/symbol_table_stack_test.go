// Package symboltable provides symbol table data structures for the CPL parser.
//
// This file contains unit tests for the symbol table stack operations.
// Tests follow table-driven patterns with testify assertions and parallel execution.
package symboltable

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSymbolTableStack verifies the NewSymbolTableStack constructor.
func TestNewSymbolTableStack(t *testing.T) {
	t.Parallel()

	stack := NewSymbolTableStack()

	require.NotNil(t, stack)
	assert.True(t, stack.IsEmpty())
	assert.Equal(t, 0, stack.Size())
}

// TestSymbolTableStackPush verifies the Push operation.
func TestSymbolTableStackPush(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		pushCount    int
		expectedSize int
	}{
		{
			name:         "Push single table",
			pushCount:    1,
			expectedSize: 1,
		},
		{
			name:         "Push multiple tables",
			pushCount:    3,
			expectedSize: 3,
		},
		{
			name:         "Push zero tables",
			pushCount:    0,
			expectedSize: 0,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			stack := NewSymbolTableStack()

			for i := 0; i < tc.pushCount; i++ {
				table := NewSymbolTable(GlobalScope)
				stack.Push(table)
			}

			assert.Equal(t, tc.expectedSize, stack.Size())
			if tc.pushCount > 0 {
				assert.False(t, stack.IsEmpty())
			}
		})
	}
}

// TestSymbolTableStackPop verifies the Pop operation.
func TestSymbolTableStackPop(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setup     func() *SymbolTableStack
		expectErr bool
		errCode   ErrorCode
	}{
		{
			name: "Pop from empty stack",
			setup: func() *SymbolTableStack {
				return NewSymbolTableStack()
			},
			expectErr: true,
			errCode:   ErrEmptyStack,
		},
		{
			name: "Pop from single table stack",
			setup: func() *SymbolTableStack {
				stack := NewSymbolTableStack()
				stack.Push(NewSymbolTable(GlobalScope))
				return stack
			},
			expectErr: false,
		},
		{
			name: "Pop from multiple table stack",
			setup: func() *SymbolTableStack {
				stack := NewSymbolTableStack()
				stack.Push(NewSymbolTable(GlobalScope))
				stack.Push(NewSymbolTable(UnitsScope))
				return stack
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			stack := tc.setup()
			initialSize := stack.Size()

			table, err := stack.Pop()

			if tc.expectErr {
				require.Error(t, err)
				assert.Nil(t, table)
				var symErr *SymbolTableError
				require.True(t, errors.As(err, &symErr))
				assert.Equal(t, tc.errCode, symErr.Code)
			} else {
				require.NoError(t, err)
				require.NotNil(t, table)
				assert.Equal(t, initialSize-1, stack.Size())
			}
		})
	}
}

// TestSymbolTableStackPopOrder verifies LIFO order of Pop.
func TestSymbolTableStackPopOrder(t *testing.T) {
	t.Parallel()

	stack := NewSymbolTableStack()

	// Push tables in order
	table1 := NewSymbolTable(GlobalScope)
	table2 := NewSymbolTable(UnitsScope)
	table3 := NewSymbolTable(GlobalScope)

	stack.Push(table1)
	stack.Push(table2)
	stack.Push(table3)

	require.Equal(t, 3, stack.Size())

	// Pop should return in reverse order
	popped3, err := stack.Pop()
	require.NoError(t, err)
	assert.Equal(t, table3, popped3)

	popped2, err := stack.Pop()
	require.NoError(t, err)
	assert.Equal(t, table2, popped2)

	popped1, err := stack.Pop()
	require.NoError(t, err)
	assert.Equal(t, table1, popped1)

	assert.True(t, stack.IsEmpty())
}

// TestSymbolTableStackPeek verifies the Peek operation.
func TestSymbolTableStackPeek(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setup     func() *SymbolTableStack
		expectErr bool
		errCode   ErrorCode
	}{
		{
			name: "Peek from empty stack",
			setup: func() *SymbolTableStack {
				return NewSymbolTableStack()
			},
			expectErr: true,
			errCode:   ErrEmptyStack,
		},
		{
			name: "Peek from single table stack",
			setup: func() *SymbolTableStack {
				stack := NewSymbolTableStack()
				stack.Push(NewSymbolTable(GlobalScope))
				return stack
			},
			expectErr: false,
		},
		{
			name: "Peek from multiple table stack",
			setup: func() *SymbolTableStack {
				stack := NewSymbolTableStack()
				stack.Push(NewSymbolTable(GlobalScope))
				stack.Push(NewSymbolTable(UnitsScope))
				return stack
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			stack := tc.setup()
			initialSize := stack.Size()

			table, err := stack.Peek()

			if tc.expectErr {
				require.Error(t, err)
				assert.Nil(t, table)
				var symErr *SymbolTableError
				require.True(t, errors.As(err, &symErr))
				assert.Equal(t, tc.errCode, symErr.Code)
			} else {
				require.NoError(t, err)
				require.NotNil(t, table)
				// Size should not change
				assert.Equal(t, initialSize, stack.Size())
			}
		})
	}
}

// TestSymbolTableStackPeekDoesNotRemove verifies Peek doesn't remove the top element.
func TestSymbolTableStackPeekDoesNotRemove(t *testing.T) {
	t.Parallel()

	stack := NewSymbolTableStack()
	table := NewSymbolTable(GlobalScope)
	stack.Push(table)

	// Peek multiple times should return same table
	for i := 0; i < 3; i++ {
		peeked, err := stack.Peek()
		require.NoError(t, err)
		assert.Equal(t, table, peeked)
		assert.Equal(t, 1, stack.Size())
	}
}

// TestSymbolTableStackIsEmpty verifies the IsEmpty operation.
func TestSymbolTableStackIsEmpty(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		setup    func() *SymbolTableStack
		expected bool
	}{
		{
			name:     "New stack is empty",
			setup:    func() *SymbolTableStack { return NewSymbolTableStack() },
			expected: true,
		},
		{
			name: "Stack with one table is not empty",
			setup: func() *SymbolTableStack {
				s := NewSymbolTableStack()
				s.Push(NewSymbolTable(GlobalScope))
				return s
			},
			expected: false,
		},
		{
			name: "Stack after push and pop is empty",
			setup: func() *SymbolTableStack {
				s := NewSymbolTableStack()
				s.Push(NewSymbolTable(GlobalScope))
				s.Pop()
				return s
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			stack := tc.setup()
			assert.Equal(t, tc.expected, stack.IsEmpty())
		})
	}
}

// TestSymbolTableStackSize verifies the Size operation.
func TestSymbolTableStackSize(t *testing.T) {
	t.Parallel()

	stack := NewSymbolTableStack()
	assert.Equal(t, 0, stack.Size())

	// Push and verify size increases
	stack.Push(NewSymbolTable(GlobalScope))
	assert.Equal(t, 1, stack.Size())

	stack.Push(NewSymbolTable(UnitsScope))
	assert.Equal(t, 2, stack.Size())

	stack.Push(NewSymbolTable(GlobalScope))
	assert.Equal(t, 3, stack.Size())

	// Pop and verify size decreases
	stack.Pop()
	assert.Equal(t, 2, stack.Size())

	stack.Pop()
	assert.Equal(t, 1, stack.Size())

	stack.Pop()
	assert.Equal(t, 0, stack.Size())
	assert.True(t, stack.IsEmpty())
}

// TestSymbolTableStackRoot verifies the Root operation.
func TestSymbolTableStackRoot(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setup     func() *SymbolTableStack
		expectErr bool
		errCode   ErrorCode
	}{
		{
			name: "Root from empty stack",
			setup: func() *SymbolTableStack {
				return NewSymbolTableStack()
			},
			expectErr: true,
			errCode:   ErrEmptyStack,
		},
		{
			name: "Root from single table stack",
			setup: func() *SymbolTableStack {
				stack := NewSymbolTableStack()
				stack.Push(NewSymbolTable(GlobalScope))
				return stack
			},
			expectErr: false,
		},
		{
			name: "Root from multiple table stack",
			setup: func() *SymbolTableStack {
				stack := NewSymbolTableStack()
				stack.Push(NewSymbolTable(GlobalScope))
				stack.Push(NewSymbolTable(UnitsScope))
				stack.Push(NewSymbolTable(GlobalScope))
				return stack
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			stack := tc.setup()

			root, err := stack.Root()

			if tc.expectErr {
				require.Error(t, err)
				assert.Nil(t, root)
				var symErr *SymbolTableError
				require.True(t, errors.As(err, &symErr))
				assert.Equal(t, tc.errCode, symErr.Code)
			} else {
				require.NoError(t, err)
				require.NotNil(t, root)
				// Root should always be the first pushed table (GlobalScope)
				assert.Equal(t, GlobalScope, root.ScopeType)
			}
		})
	}
}

// TestSymbolTableStackRootReturnsFirstPushed verifies Root returns first pushed table.
func TestSymbolTableStackRootReturnsFirstPushed(t *testing.T) {
	t.Parallel()

	stack := NewSymbolTableStack()

	rootTable := NewSymbolTable(GlobalScope)
	stack.Push(rootTable)
	stack.Push(NewSymbolTable(UnitsScope))
	stack.Push(NewSymbolTable(GlobalScope))
	stack.Push(NewSymbolTable(UnitsScope))

	root, err := stack.Root()
	require.NoError(t, err)
	assert.Equal(t, rootTable, root)
}

// TestSymbolTableStackIntegration verifies the full lifecycle of stack operations.
func TestSymbolTableStackIntegration(t *testing.T) {
	t.Parallel()

	stack := NewSymbolTableStack()

	// Initially empty
	assert.True(t, stack.IsEmpty())
	assert.Equal(t, 0, stack.Size())

	// Push global table (simulating parser start)
	globalTable := NewGlobalSymbolTable()
	stack.Push(globalTable)
	assert.False(t, stack.IsEmpty())
	assert.Equal(t, 1, stack.Size())

	// Peek should return global
	peeked, err := stack.Peek()
	require.NoError(t, err)
	assert.Equal(t, globalTable, peeked)

	// Push units scope (simulating entering units section)
	unitsTable := globalTable.AddChild(UnitsScope)
	stack.Push(unitsTable)
	assert.Equal(t, 2, stack.Size())

	// Peek should return units
	peeked, err = stack.Peek()
	require.NoError(t, err)
	assert.Equal(t, unitsTable, peeked)

	// Root should still be global
	root, err := stack.Root()
	require.NoError(t, err)
	assert.Equal(t, globalTable, root)

	// Pop units scope (simulating exiting units section)
	popped, err := stack.Pop()
	require.NoError(t, err)
	assert.Equal(t, unitsTable, popped)
	assert.Equal(t, 1, stack.Size())

	// Peek should return global again
	peeked, err = stack.Peek()
	require.NoError(t, err)
	assert.Equal(t, globalTable, peeked)

	// Pop global (end of parsing)
	popped, err = stack.Pop()
	require.NoError(t, err)
	assert.Equal(t, globalTable, popped)
	assert.True(t, stack.IsEmpty())

	// Further operations should fail
	_, err = stack.Pop()
	require.Error(t, err)
	_, err = stack.Peek()
	require.Error(t, err)
	_, err = stack.Root()
	require.Error(t, err)
}
