// Package symboltable provides symbol table data structures for the CPL parser.
//
// This file contains the SymbolTableStack implementation for managing
// a stack of symbol tables during parsing. The stack follows LIFO semantics
// and supports scope entry/exit operations.
package symboltable

// SymbolTableStack manages a stack of symbol tables during parsing.
// It provides LIFO operations for scope management and maintains the
// hierarchical relationship between symbol tables.
type SymbolTableStack struct {
	// tables is the internal slice storing the stack of symbol tables.
	// The last element is the top of the stack.
	tables []*SymbolTable
}

// NewSymbolTableStack creates a new empty stack.
func NewSymbolTableStack() *SymbolTableStack {
	return &SymbolTableStack{
		tables: make([]*SymbolTable, 0),
	}
}

// Push adds a symbol table to the top of the stack.
func (sts *SymbolTableStack) Push(st *SymbolTable) {
	sts.tables = append(sts.tables, st)
}

// Pop removes and returns the top symbol table from the stack.
// Returns (nil, *SymbolTableError) with code ErrEmptyStack if the stack is empty.
func (sts *SymbolTableStack) Pop() (*SymbolTable, error) {
	if sts.IsEmpty() {
		return nil, &SymbolTableError{
			Code:    ErrEmptyStack,
			Message: "cannot pop from empty stack",
		}
	}

	table := sts.tables[len(sts.tables)-1]

	if len(sts.tables) > 0 {
		// Remove the table from the slice.
		sts.tables = sts.tables[:len(sts.tables)-1]
	}

	return table, nil
}

// Peek returns the top symbol table without removing it.
// Returns (nil, *SymbolTableError) with code ErrEmptyStack if the stack is empty.
func (sts *SymbolTableStack) Peek() (*SymbolTable, error) {
	if sts.IsEmpty() {
		return nil, &SymbolTableError{
			Code:    ErrEmptyStack,
			Message: "cannot peek empty stack",
		}
	}

	return sts.tables[len(sts.tables)-1], nil
}

// IsEmpty returns true if the stack has no tables.
func (sts *SymbolTableStack) IsEmpty() bool {
	return len(sts.tables) == 0
}

// Size returns the number of tables on the stack.
func (sts *SymbolTableStack) Size() int {
	return len(sts.tables)
}

// Root returns the root (global) symbol table.
// This is the first table pushed onto the stack.
// Returns (nil, *SymbolTableError) with code ErrEmptyStack if the stack is empty.
func (sts *SymbolTableStack) Root() (*SymbolTable, error) {
	if sts.IsEmpty() {
		return nil, &SymbolTableError{
			Code:    ErrEmptyStack,
			Message: "cannot get root of empty stack",
		}
	}

	return sts.tables[0], nil
}
