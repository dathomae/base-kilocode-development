# Symbol Table Design for CPL Parser

## Overview

The symbol table is a core data structure used during the **parsing phase** of CPL (Chronocone Planning Language) processing. As the parser walks the AST, it maintains a stack of symbol tables to track names and their corresponding definitions across different scopes. The resulting symbol table tree is passed to the **analysis phase** where it is walked to accumulate declarations and validate references.

### Integration with CPL Listener Error Handling

The symbol table follows the same error handling pattern as the CPL listener (implemented in Story113). Symbol table operations return errors using Go's idiomatic `(result, error)` pattern. These errors are then captured by the listener's `AddError()` method and accumulated in the listener's `errors []error` slice. This allows semantic errors from symbol table operations (duplicate definitions, undefined symbols, stack errors) to be reported alongside other semantic errors.

## Architecture

### Symbol Table Stack

The parser maintains a symbol table stack during parsing:

1. **Global Scope Initialization**: Before parsing begins, a global symbol table is pushed onto the stack
2. **Scope Entry**: When entering a new scope (e.g., units section), a new symbol table is:
   - Pushed onto the stack
   - Linked to its parent (the table previously on top)
   - Added to the parent's list of child tables
3. **Scope Exit**: When exiting a scope, the symbol table is popped from the stack
4. **Final State**: At the end of parsing, exactly one symbol table remains - the global table

### Symbol Table Structure

Each symbol table contains:

| Field              | Type                 | Description                                                 |
| ------------------ | -------------------- | ----------------------------------------------------------- |
| `ScopeType`        | `ScopeType`          | Enumeration indicating the type of scope this table defines |
| `Parent`           | `*SymbolTable`       | Pointer to parent symbol table (nil for root/global)        |
| `Children`         | `[]*SymbolTable`     | List of pointers to child symbol tables                     |
| `Definitions`      | `[]SymbolDefinition` | List of symbol definitions in this scope                    |
| `IndentationLevel` | `int`                | Nesting level for indentation-based scoping (0 = root)      |

The `IndentationLevel` field is used for CPL features that rely on indentation-based scoping, such as concept hierarchies (see [Concepts Parsing ERO](../eros/ERO001_concepts-parsing-phase.md)). It tracks the indentation depth at which a scope was created, enabling the parser and analyzer to establish parent-child relationships based on source code indentation.

### Scope Types

The following scope types are defined:

| Scope Type    | Description                        |
| ------------- | ---------------------------------- |
| `GlobalScope` | The root scope of the CPL document |
| `UnitsScope`  | Scope containing unit declarations |

Additional scope types for other CPL constructs (connectors, metrics, aggregators, etc.) can be added as the grammar evolves.

### Symbol Definition

A symbol definition captures information about a named entity:

| Field            | Type             | Description                                             |
| ---------------- | ---------------- | ------------------------------------------------------- |
| `Name`           | `*string`        | The identifier name (nil if the definition has no name) |
| `DefinitionType` | `DefinitionType` | The kind of definition (unit, connector, metric, etc.)  |
| `Location`       | `SourceLocation` | Where the symbol was defined                            |
| `Data`           | `interface{}`    | Type-specific data (to be populated during analysis)    |

## Go Implementation

### Types

```go
// ScopeType enumerates the different kinds of scopes in CPL.
type ScopeType int

const (
    GlobalScope ScopeType = iota
    UnitsScope
    // Additional scope types for connectors, metrics, aggregators, etc.
    // can be added here as the grammar evolves
)

// DefinitionType enumerates the kinds of definitions that can appear
// in a symbol table.
type DefinitionType int

const (
    UnitDefinition DefinitionType = iota
    // Additional definition types for connectors, metrics, aggregators, etc.
    // can be added here as the grammar evolves
)

// SourceLocation captures where a symbol is defined in source.
type SourceLocation struct {
    Line   int
    Column int
}

// SymbolDefinition captures a name and its definition details.
// Name is a pointer to allow for optional names (nil if no name).
type SymbolDefinition struct {
    Name           *string
    DefinitionType DefinitionType
    Location       SourceLocation
    Data           interface{}
}

// SymbolTable captures a set of names and their corresponding definitions.
type SymbolTable struct {
    ScopeType        ScopeType
    Parent           *SymbolTable
    Children         []*SymbolTable
    Definitions      []SymbolDefinition
    IndentationLevel int  // Nesting level for indentation-based scoping (0 = root)
}
```

### Symbol Table Operations

```go
// NewSymbolTable creates a new symbol table with the given scope type.
func NewSymbolTable(scopeType ScopeType) *SymbolTable

// NewGlobalSymbolTable creates a new global (root) symbol table.
func NewGlobalSymbolTable() *SymbolTable

// AddChild creates a new symbol table as a child of this one.
// The new table is added to this table's Children list.
func (st *SymbolTable) AddChild(scopeType ScopeType) *SymbolTable

// Define adds a new symbol definition to this table.
// Returns an error if the name is already defined in this scope.
// If name is nil, the definition is added without a name.
func (st *SymbolTable) Define(name *string, defType DefinitionType,
    location SourceLocation) error

// Lookup searches for a symbol definition by name starting from this table
// and walking up through parent tables.
// Returns an error if the symbol is not found.
func (st *SymbolTable) Lookup(name string) (*SymbolDefinition, error)

// LookupLocal searches for a symbol definition by name in this table only.
// Returns an error if the symbol is not found.
func (st *SymbolTable) LookupLocal(name string) (*SymbolDefinition, error)
```

**Go Error Handling Pattern**

Following Go best practices, symbol table operations return both results and errors:

- Functions return `(*SymbolDefinition, error)` where the caller checks for errors
- A `nil` error indicates success
- Non-nil errors provide context about what went wrong
- Callers must explicitly check and handle errors

Example usage:

```go
def, err := table.Lookup("mySymbol")
if err != nil {
    // Handle error - symbol not found
    return err
}
// Use def...
```

### Symbol Table Stack

```go
// SymbolTableStack manages a stack of symbol tables during parsing.
type SymbolTableStack struct {
    tables []*SymbolTable
}

// NewSymbolTableStack creates a new empty stack.
func NewSymbolTableStack() *SymbolTableStack

// Push adds a symbol table to the top of the stack.
func (sts *SymbolTableStack) Push(st *SymbolTable)

// Pop removes and returns the top symbol table from the stack.
// Returns an error if the stack is empty.
func (sts *SymbolTableStack) Pop() (*SymbolTable, error)

// Peek returns the top symbol table without removing it.
// Returns an error if the stack is empty.
func (sts *SymbolTableStack) Peek() (*SymbolTable, error)

// IsEmpty returns true if the stack has no tables.
func (sts *SymbolTableStack) IsEmpty() bool

// Size returns the number of tables on the stack.
func (sts *SymbolTableStack) Size() int

// Root returns the root (global) symbol table.
// Returns an error if the stack is empty.
func (sts *SymbolTableStack) Root() (*SymbolTable, error)
```

## Parser Integration

### Initialization

Before parsing begins:

1. Create a new symbol table stack
2. Create the global symbol table
3. Push the global table onto the stack

```go
stack := NewSymbolTableStack()
global := NewGlobalSymbolTable()
stack.Push(global)
```

### Scope Management

During parsing, the parser uses Enter/Exit callbacks to manage scope:

**Entering Units Section:**

```go
func (listener *CPLListener) EnterUnitsSection(ctx *antlr.UnitsSectionContext) {
    current, err := listener.symbolStack.Peek()
    if err != nil {
        listener.AddError(fmt.Errorf("symbol stack error: %w", err))
        return
    }
    unitsTable := current.AddChild(UnitsScope)
    listener.symbolStack.Push(unitsTable)
}
```

**Exiting Units Section:**

```go
func (listener *CPLListener) ExitUnitsSection(ctx *antlr.UnitsSectionContext) {
    _, err := listener.symbolStack.Pop()
    if err != nil {
        listener.AddError(fmt.Errorf("symbol stack error: %w", err))
    }
}
```

### Symbol Definition

When parsing declarations, symbols are added to the current scope:

**Unit Declaration:**

```go
func (listener *CPLListener) ExitUnitDeclaration(ctx *antlr.UnitDeclarationContext) {
    // ... existing AST building code ...

    current, err := listener.symbolStack.Peek()
    if err != nil {
        listener.AddError(fmt.Errorf("symbol stack error: %w", err))
        return
    }

    err = current.Define(
        &listener.currentUnit.Name,  // Pass pointer to name
        UnitDefinition,
        SourceLocation{Line: ctx.GetStart().GetLine(), Column: ctx.GetStart().GetColumn()},
    )
    if err != nil {
        listener.AddError(fmt.Errorf("unit definition error: %w", err))
    }
}
```

**Anonymous Definitions:**

For definitions without names, pass `nil` as the name:

```go
func (listener *CPLListener) ExitAnonymousDefinition(ctx *antlr.AnonymousDefinitionContext) {
    current, err := listener.symbolStack.Peek()
    if err != nil {
        listener.AddError(fmt.Errorf("symbol stack error: %w", err))
        return
    }

    err = current.Define(
        nil,  // No name for this definition
        SomeDefinitionType,
        SourceLocation{Line: ctx.GetStart().GetLine(), Column: ctx.GetStart().GetColumn()},
    )
    if err != nil {
        listener.AddError(fmt.Errorf("definition error: %w", err))
    }
}
```

## Symbol Table Tree

After parsing completes, the symbol tables form a tree structure rooted at the global table:

```
GlobalSymbolTable (GlobalScope)
├── UnitsSymbolTable (UnitsScope)
├── ConnectorsSymbolTable (ConnectorsScope)
├── MetricsSymbolTable (MetricsScope)
└── ... additional scope tables as grammar evolves
```

The tree structure is built through parent-child relationships:

- Each child table maintains a `Parent` pointer to its containing scope
- Each parent table maintains a `Children` list of its nested scopes
- This enables tree-walking during the analysis phase

## Analysis Phase Usage

During analysis, the symbol table tree is walked to:

1. **Accumulate Declarations**: Walk the tree to find all scope tables and extract definitions
2. **Validate References**: Use the tree structure to resolve symbol references and detect undefined names
3. **Detect Duplicates**: Check for duplicate definitions within and across scopes

### Tree Walking Pattern

```go
// WalkSymbolTable recursively walks the symbol table tree,
// calling the provided function for each table.
func WalkSymbolTable(table *SymbolTable, fn func(*SymbolTable)) {
    fn(table)
    for _, child := range table.Children {
        WalkSymbolTable(child, fn)
    }
}
```

## Validation

The symbol table enforces the following invariants:

1. **Unique Named Definitions Within Scope**: A symbol table cannot contain duplicate names among its named definitions (definitions where `Name != nil`). Unnamed definitions (where `Name == nil`) can appear multiple times.
2. **Single Root**: After parsing, exactly one table (the global table) should remain on the stack
3. **Proper Nesting**: Child tables must have their parent as the table that was on top of the stack when they were created
4. **Balanced Push/Pop**: Every scope entry must have a corresponding exit

## Error Handling

The symbol table reports the following errors during parsing:

| Error                 | Description                                         |
| --------------------- | --------------------------------------------------- |
| `DuplicateDefinition` | A named definition already exists in the same scope |
| `SymbolNotFound`      | A lookup failed to find the requested symbol        |
| `UnbalancedStack`     | Parsing ended with more than one table on the stack |
| `EmptyStack`          | An operation was attempted on an empty stack        |

### Error Types

Symbol table errors implement the standard Go `error` interface and can be captured by the CPL listener's `AddError()` method. Errors include context about the operation that failed.

```go
// SymbolTableError represents an error from symbol table operations.
type SymbolTableError struct {
    Code    ErrorCode
    Message string
    Name    string  // The symbol name involved, if applicable
}

func (e *SymbolTableError) Error() string {
    return e.Message
}

// ErrorCode identifies the type of symbol table error.
type ErrorCode int

const (
    ErrDuplicateDefinition ErrorCode = iota
    ErrSymbolNotFound
    ErrUnbalancedStack
    ErrEmptyStack
)
```

### Integration with Listener Error Slice

Symbol table errors are captured by the CPL listener and added to its error slice:

```go
// In CPLListener method
current, err := listener.symbolStack.Peek()
if err != nil {
    // Wrap and add to listener's error slice
    listener.AddError(fmt.Errorf("symbol table error: %w", err))
    return
}

err = current.Define(&name, UnitDefinition, location)
if err != nil {
    // Wrap and add to listener's error slice
    listener.AddError(fmt.Errorf("semantic error: %w", err))
}
```

The wrapped errors follow Go's error wrapping conventions (`fmt.Errorf` with `%w` verb), allowing error inspection using `errors.Is()` and `errors.As()`.

### Caller Responsibility

Following Go conventions, callers must check for errors returned by symbol table operations. In the context of the CPL listener, errors are wrapped and added to the listener's error slice:

```go
// Example: Defining a symbol in a listener method
name := "mySymbol"
err := table.Define(&name, UnitDefinition, location)
if err != nil {
    // Add to listener's error slice for comprehensive error reporting
    listener.AddError(fmt.Errorf("failed to define unit: %w", err))
    return
}

// Example: Looking up a symbol in a listener method
def, err := table.Lookup("mySymbol")
if err != nil {
    var symErr *SymbolTableError
    if errors.As(err, &symErr) && symErr.Code == ErrSymbolNotFound {
        listener.AddError(fmt.Errorf("undefined symbol: %w", err))
    } else {
        listener.AddError(fmt.Errorf("lookup error: %w", err))
    }
    return
}
// Use def...
```

Error wrapping with `%w` allows error inspection using `errors.Is()` and `errors.As()` when processing the listener's error slice after parsing.

## Future Extensions

The symbol table design supports future CPL features:

- **Nested Scopes**: The tree structure naturally supports nested scopes (e.g., within metric expressions)
- **Additional Definition Types**: New `DefinitionType` values can be added for connectors, metrics, aggregators
- **Symbol Attributes**: The `Data` field allows attaching type-specific information during analysis
