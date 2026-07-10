# Story110: CPL Symbol Table Implementation

## Goal

Implement the symbol table data structures and stack management for the CPL parser according to the design document, enabling scope tracking and symbol definition during parsing. The symbol table integrates with the listener's error slice pattern established in Story113.

## References

- [Symbol Table Design](../specifications/design/cpl/go-target/symbol-table-design.md)
- [Slog Logging Infrastructure](../specifications/design/cpl/go-target/slog-logging-infrastructure.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md) - Testing patterns and organization
- [Story109_cpl-parser-slog-logging.md](finished-stories/Story109_cpl-parser-slog-logging.md) - Completed
- [Story111_remove-dimensions-from-cpl.md](Story111_remove-dimensions-from-cpl.md) - Removes dimensions (dependency)
- [Story112_refactor-cpl-tests-to-go-best-practices.md](Story112_refactor-cpl-tests-to-go-best-practices.md) - Go testing patterns (dependency)
- [Story113_cpl-listener-error-slice.md](Story113_cpl-listener-error-slice.md) - Error handling pattern (dependency)

## Dependencies

- [Story109_cpl-parser-slog-logging.md](finished-stories/Story109_cpl-parser-slog-logging.md) - Completed (provides the base parser structure)
- [Story111_remove-dimensions-from-cpl.md](Story111_remove-dimensions-from-cpl.md) - Must complete first (removes dimensions-related code)
- [Story112_refactor-cpl-tests-to-go-best-practices.md](Story112_refactor-cpl-tests-to-go-best-practices.md) - Must complete first (establishes testify and testing patterns)
- [Story113_cpl-listener-error-slice.md](Story113_cpl-listener-error-slice.md) - Must complete first (provides error slice infrastructure)

## Dependent Stories

- [Story104_cpl-parser-connectors.md](Story104_cpl-parser-connectors.md) - Not Started
- [Story106_cpl-parser-metrics.md](Story106_cpl-parser-metrics.md) - Not Started

## Tasks

1. Create symbol table type definitions - Completed
   a. Define ErrorCode enumeration (ErrDuplicateDefinition, ErrSymbolNotFound, ErrUnbalancedStack, ErrEmptyStack) - Completed
   b. Define ScopeType enumeration (GlobalScope, UnitsScope) with String() method - Completed
   c. Define DefinitionType enumeration (UnitDefinition) with String() method - Completed
   d. Create SourceLocation struct with Line and Column fields - Completed
   e. Create SymbolDefinition struct with Name, DefinitionType, Location, and Data fields - Completed
   f. Create SymbolTableError struct implementing error interface - Completed
   g. Create SymbolTable struct with ScopeType, Parent, Children, and Definitions fields - Completed
   h. Write table-driven unit tests with testify assertions and parallel execution - Completed
   i. Implement type definitions to make tests pass - Completed

2. Implement SymbolTable operations - Completed
   a. Define NewSymbolTable(scopeType ScopeType) *SymbolTable constructor signature - Completed
   b. Define NewGlobalSymbolTable() *SymbolTable convenience constructor signature - Completed
   c. Define AddChild(scopeType ScopeType) *SymbolTable method signature - Completed
   d. Define Define(name *string, defType DefinitionType, location SourceLocation) error method signature - Completed
   e. Define Lookup(name string) (*SymbolDefinition, error) method signature - Completed
   f. Define LookupLocal(name string) (*SymbolDefinition, error) method signature - Completed
   g. Write table-driven unit tests with testify assertions and parallel execution - Completed
   h. Implement SymbolTable methods to make tests pass - Completed

3. Implement SymbolTableStack for parser integration - Completed
   a. Create SymbolTableStack struct with tables slice - Completed
   b. Define NewSymbolTableStack() *SymbolTableStack constructor signature - Completed
   c. Define Push(st *SymbolTable) method signature - Completed
   d. Define Pop() (*SymbolTable, error) method signature - Completed
   e. Define Peek() (*SymbolTable, error) method signature - Completed
   f. Define IsEmpty() bool method signature - Completed
   g. Define Size() int method signature - Completed
   h. Define Root() (\*SymbolTable, error) method signature - Completed
   i. Write table-driven unit tests with testify assertions and parallel execution - Completed
   j. Implement SymbolTableStack methods to make tests pass - Completed

4. Add tree walking utilities - Completed
   a. Define WalkSymbolTable function signature for recursive tree traversal - Completed
   b. Write table-driven tests with testify assertions for WalkSymbolTable - Completed
   c. Implement WalkSymbolTable function - Completed

5. Integrate symbol table with CPL parser listener - Completed
   a. Add symbolStack field to CPLListener struct - Completed
   b. Initialize symbol table stack in NewCPLListener constructor - Completed
   c. Define EnterUnitsSection callback signature to push UnitsScope table - Completed
   d. Define ExitUnitsSection callback signature to pop scope with error handling - Completed
   e. Define ExitUnitDeclaration callback signature to define unit symbols - Completed
   f. Write integration tests with testify assertions and parallel execution - Completed
   g. Implement parser integration using listener AddError() for error reporting - Completed

6. Add validation and error handling - Completed
   a. Write tests for duplicate definition detection using table-driven pattern - Completed
   b. Write tests for stack validation (empty stack, unbalanced) - Completed
   c. Verify integration with listener error slice works correctly - Completed

## Test-First Development

This story follows the Logical TDD Lifecycle as defined in [memory-bank/concepts.md](../concepts.md). Each coding task integrates the full TDD cycle:

1. **Phase A - Structural Definition**: Define type and method signatures before writing tests or implementation
2. **Phase B - Logical Contract**: Write failing tests using table-driven pattern with testify assertions
3. **Phase C - Implementation**: Implement code to make tests pass
4. **Phase D - Verification**: Verify all tests pass with parallel execution

### Go Testing Best Practices

**All tests must follow the patterns defined in the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md).** This guide provides essential patterns for:

- Code organization (test file naming, package structure)
- Test construction (table-driven tests, parallel tests, subtests)
- Testing utilities (`t.Cleanup()`, `t.Helper()`, `t.TempDir()`)
- Assertion patterns using testify

All tests must follow the patterns established in Story112:

**1. Table-Driven Tests**

```go
testCases := []struct {
    name        string
    input       string
    expectError bool
    errorCode   ErrorCode
}{
    // test cases
}
```

**2. Parallel Execution**

Use `t.Parallel()` at both test function and subtest levels with proper variable capture:

```go
func TestXxx(t *testing.T) {
    t.Parallel()
    for _, tc := range testCases {
        tc := tc  // Variable capture for parallel execution
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            // test code
        })
    }
}
```

**3. Testify Assertions**

Use `require` for fail-fast assertions and `assert` for multiple checks:

```go
require.NoError(t, err)
assert.Equal(t, expected, actual)
```

**4. Error Inspection**

Use `errors.As()` to verify specific error types and codes:

```go
var symErr *SymbolTableError
require.Error(t, err)
require.True(t, errors.As(err, &symErr))
assert.Equal(t, ErrSymbolNotFound, symErr.Code)
```

## Constraints

- Follow the design documented in symbol-table-design.md exactly
- Use Go idioms and naming conventions (e.g., NewXXX constructors, String() methods for enums)
- All operations must maintain proper parent-child relationships in the tree
- The Data field in SymbolDefinition should use interface{} for future extensibility
- Symbol table must be thread-safe for single-parser use (no concurrent access required)
- Stack operations must handle edge cases (empty stack, unbalanced push/pop)
- Duplicate definitions within the same scope must be detected and reported with location
- Lookup methods return errors instead of boolean (following Go best practices)
- Integration must work with existing ANTLR-generated parser listeners and Story113 error slice
- All tests must use table-driven pattern, testify assertions, and parallel execution
- Errors must be wrapped and added to listener's error slice using AddError()

## Intent

The symbol table serves as the bridge between parsing and analysis phases. During parsing, it tracks which names are defined in which scopes, enabling detection of duplicates and providing a structure for later analysis to accumulate full declaration information. The tree structure supports future nested scopes (e.g., within metric expressions).

The symbol table follows the same error handling pattern as the CPL listener: operations return errors using Go's idiomatic `(result, error)` pattern. These errors are then captured by the listener's `AddError()` method and accumulated in the listener's `errors []error` slice. This allows semantic errors from symbol table operations to be reported alongside other semantic errors.

## Acceptance Criteria

- All type definitions match the design document specifications
- ScopeType only includes GlobalScope and UnitsScope (no dimensions)
- DefinitionType only includes UnitDefinition (no dimensions)
- SymbolTable supports creating parent-child relationships and hierarchical lookups
- Lookup and LookupLocal return `(*SymbolDefinition, error)` following Go conventions
- SymbolTableStack provides proper LIFO semantics for parser scope management
- SymbolTableError implements the error interface with ErrorCode for error inspection
- Parser listener correctly enters/exits scopes and defines symbols during parsing
- Duplicate definitions in the same scope are detected and reported with accurate line/column
- Tree walking utilities can traverse the entire symbol table structure
- All tests pass with >90% coverage of symbol table code
- All tests use table-driven pattern with testify assertions and parallel execution
- Integration with existing parser and Story113 error slice works correctly
- Documentation comments follow Go conventions for all exported types and functions
- Symbol table errors are properly wrapped and added to listener's error slice

## Notes

### Symbol Table Error Types

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
func (listener *CPLListener) ExitUnitDeclaration(ctx *antlr.UnitDeclarationContext) {
    // ... existing AST building code ...

    current, err := listener.symbolStack.Peek()
    if err != nil {
        listener.AddError(fmt.Errorf("symbol stack error: %w", err))
        return
    }

    err = current.Define(
        &listener.currentUnit.Name,
        UnitDefinition,
        SourceLocation{Line: ctx.GetStart().GetLine(), Column: ctx.GetStart().GetColumn()},
    )
    if err != nil {
        listener.AddError(fmt.Errorf("unit definition error: %w", err))
    }
}
```

### Go Error Handling Pattern

Following Go best practices, symbol table operations return both results and errors:

```go
// Lookup returns error instead of boolean
def, err := table.Lookup("mySymbol")
if err != nil {
    var symErr *SymbolTableError
    if errors.As(err, &symErr) && symErr.Code == ErrSymbolNotFound {
        // Handle not found
    }
    return err
}
// Use def...
```

### Future Extensions

The symbol table design supports future CPL features:

- **Nested Scopes**: The tree structure naturally supports nested scopes (e.g., within metric expressions)
- **Additional Definition Types**: New `DefinitionType` values can be added for connectors, metrics, aggregators
- **Symbol Attributes**: The `Data` field allows attaching type-specific information during analysis

### Files Created

| File                                                                                   | Purpose                                |
| -------------------------------------------------------------------------------------- | -------------------------------------- |
| `src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go`             | Core symbol table types and operations |
| `src/chronocone_planning_language/src/main/go/symboltable/symbol_table_stack.go`       | Stack management for parser            |
| `src/chronocone_planning_language/src/main/go/symboltable/symbol_table_walk.go`        | Tree walking utilities                 |
| `src/chronocone_planning_language/src/main/go/symboltable/symbol_table_test.go`        | Unit tests for symbol table            |
| `src/chronocone_planning_language/src/main/go/symboltable/symbol_table_stack_test.go`  | Unit tests for stack operations        |
| `src/chronocone_planning_language/src/main/go/symboltable/symbol_table_walk_test.go`   | Unit tests for tree walking            |
| `src/chronocone_planning_language/src/main/go/parser/symbol_table_integration_test.go` | Integration tests with parser          |

### Files Modified

| File                                                                           | Changes                                        |
| ------------------------------------------------------------------------------ | ---------------------------------------------- |
| `src/chronocone_planning_language/src/main/go/parser/parser_support.go`        | Added symbolStack field and initialization     |
| `src/chronocone_planning_language/src/main/go/parser/cpl_document_listener.go` | Added Enter/Exit UnitsSection callbacks        |
| `src/chronocone_planning_language/src/main/go/parser/cpl_units_listener.go`    | Added symbol definition in ExitUnitDeclaration |
