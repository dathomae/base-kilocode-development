# Story113: CPL Listener Error Slice

## Goal

Add an error[] slice to CPLListener to capture semantic errors during parsing. Currently, the listener silently ignores or only logs errors that occur during AST construction. This story implements proper error collection so that semantic errors (like missing identifiers or parameters outside unit context) are captured and returned alongside syntax errors, providing comprehensive error reporting for CPL document parsing.

## References

- [CPL Parser README](../../src/chronocone_planning_language/README.md)
- [Story109_cpl-parser-slog-logging.md](finished-stories/Story109_cpl-parser-slog-logging.md) - Structured logging infrastructure
- [Story111_remove-dimensions-from-cpl.md](Story111_remove-dimensions-from-cpl.md) - Removes dimensions code that affects files being modified
- [Logical TDD Lifecycle](../concepts.md) - Development methodology for this story
- [CPL Listener](src/chronocone_planning_language/src/main/go/parser/chronocone_planning_language_listener.go)
- [CPL Units Listener](src/chronocone_planning_language/src/main/go/parser/cpl_units_listener.go)
- [CPL Parser](src/chronocone_planning_language/src/main/go/parser/cpl_parser.go)

## Dependencies

- [Story112_refactor-cpl-tests-to-go-best-practices.md](Story112_refactor-cpl-tests-to-go-best-practices.md) - Not Started
  - Must be completed first to establish testify dependency and Go testing patterns
  - Provides the table-driven test patterns, testify assertions, and parallel execution examples
- [Story111_remove-dimensions-from-cpl.md](Story111_remove-dimensions-from-cpl.md) - Not Started
  - This story removes dimensions-related code from the files that need modification
  - The error slice implementation should be added after dimensions removal to avoid merge conflicts

## Dependent Stories

None

## Tasks

1. Add error slice infrastructure to CPLListener - Completed
   a. Define the CPLListener error slice field and methods (Phase A: Structural Definition) - Completed
   b. Write tests for error infrastructure (Phase B: Logical Contract) - Completed
   c. Implement error infrastructure (Phase C: Implementation) - Completed
   d. Verify error infrastructure tests pass (Phase D: Verification) - Completed

2. Update ExitIdentifier to capture semantic errors - Completed
   a. Define error handling for ExitIdentifier (Phase A: Structural Definition) - Completed
   b. Write tests for ExitIdentifier error cases (Phase B: Logical Contract) - Completed
   c. Replace silent return with AddError call in ExitIdentifier (Phase C: Implementation) - Completed
   d. Verify ExitIdentifier error handling works correctly (Phase D: Verification) - Completed

3. Update ExitParameter to capture semantic errors - Completed
   a. Define error handling for ExitParameter (Phase A: Structural Definition) - Completed
   b. Write tests for ExitParameter error cases (Phase B: Logical Contract) - Completed
   c. Add error handling to ExitParameter (Phase C: Implementation) - Completed
   d. Verify ExitParameter error handling works correctly (Phase D: Verification) - Completed

4. Update ExitUnitDeclaration to capture semantic errors - Completed
   a. Define error handling for ExitUnitDeclaration (Phase A: Structural Definition) - Completed
   b. Write tests for ExitUnitDeclaration error cases (Phase B: Logical Contract) - Completed
   c. Add error handling to ExitUnitDeclaration (Phase C: Implementation) - Completed
   d. Verify ExitUnitDeclaration error handling works correctly (Phase D: Verification) - Completed

5. Update ParseWithLogLevel to check and return listener errors - Completed
   a. Define error combination approach (Phase A: Structural Definition) - Completed
   b. Write tests for combined error handling (Phase B: Logical Contract) - Completed
   c. Implement combined error handling in ParseWithLogLevel (Phase C: Implementation) - Completed
   d. Verify ParseWithLogLevel error handling works correctly (Phase D: Verification) - Completed

6. Add integration tests for error collection - Completed
   a. Define integration test scenarios (Phase A: Structural Definition) - Completed
   b. Write integration tests (Phase B: Logical Contract) - Completed
   c. Ensure existing tests pass with error infrastructure (Phase C: Implementation) - Completed
   d. Verify all tests pass (Phase D: Verification) - Completed

## Test-First Development

This story follows the Logical TDD Lifecycle as defined in [memory-bank/concepts.md](../concepts.md). Each coding task integrates the full TDD cycle:

1. **Phase A - Structural Definition**: Define the method signatures, error types, and interfaces before writing any tests or implementation
2. **Phase B - Logical Contract**: Write failing tests that define the expected behavior and error conditions
3. **Phase C - Implementation**: Implement the code to make tests pass
4. **Phase D - Verification**: Verify all tests pass and no regressions occur

Key principles for this story:

- Tests are written **before** implementation in each task
- Error cases are tested explicitly with clear failure conditions
- Tests verify both error conditions and successful operation
- Integration tests ensure semantic errors surface through the public ParseWithLogLevel API

## Constraints

- Must not break existing test functionality - all existing tests must continue to pass
- Error messages must be descriptive and include context (rule name, expected identifier, etc.)
- Error slice must be thread-safe if parsing could be concurrent (document if not thread-safe)
- Must use existing slog logger for consistent logging
- Must follow Go conventions for error handling (error interface, error wrapping with fmt.Errorf)
- Must maintain backward compatibility with existing Parse() function behavior
- All new tests must follow Go testing best practices (table-driven, parallel, testify assertions)
- Test files must be created in `src/test/go/parser/` directory

## Intent

The intent is to provide comprehensive error reporting for CPL parsing that captures both syntax errors (from ANTLR) and semantic errors (from listener validation). Currently, semantic errors like missing identifiers or parameters outside unit context are silently ignored, making it difficult for users to understand why their CPL documents fail to parse correctly. By collecting these errors in a slice and returning them alongside syntax errors, users get complete visibility into all parsing issues.

## Acceptance Criteria

1. CPLListener has an `errors []error` field that accumulates semantic errors
2. `AddError(err error)` method appends errors to the slice and logs them
3. `GetErrors() []error` method returns all accumulated errors
4. `HasErrors() bool` method returns true when errors exist
5. `NewCPLListener` initializes an empty error slice
6. `ExitIdentifier` adds an error when ctx.IDENTIFIER() is nil instead of silently returning
7. `ExitParameter` adds an error when l.currentUnit is nil
8. `ExitUnitDeclaration` validates units and adds errors for missing required fields
9. `ParseWithLogLevel` combines syntax errors and semantic errors into a single comprehensive error
10. All existing tests continue to pass
11. New tests verify error collection for all identified error scenarios
12. Error messages are descriptive and include relevant context

## Notes

### Go Testing Best Practices

All tests in this story must follow the Go testing best practices established in Story112:

**1. Table-Driven Tests**

Structure tests as a slice of struct test cases:

```go
testCases := []struct {
    name        string
    input       string
    expectError bool
    errorMsg    string
}{
    // test cases
}
```

**2. Parallel Execution**

Use `t.Parallel()` at both the test function and subtest levels:

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

**4. Helper Functions**

Use `t.Helper()` in helper functions for proper error reporting:

```go
func parseDocument(t *testing.T, input string) *CPLDocument {
    t.Helper()
    // helper code
}
```

### Error Types to Define

```go
// SemanticError represents a semantic error during parsing
type SemanticError struct {
    Message string
    Rule    string
    Line    int
    Column  int
}

func (e *SemanticError) Error() string {
    return fmt.Sprintf("semantic error at line %d:%d in %s: %s", e.Line, e.Column, e.Rule, e.Message)
}
```

### Files to Modify

| File                                                                                           | Changes                                                           |
| ---------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| `src/chronocone_planning_language/src/main/go/parser/chronocone_planning_language_listener.go` | Add error slice and methods, update ExitIdentifier, ExitParameter |
| `src/chronocone_planning_language/src/main/go/parser/cpl_units_listener.go`                    | Update ExitUnitDeclaration with validation                        |
| `src/chronocone_planning_language/src/main/go/parser/cpl_parser.go`                            | Update ParseWithLogLevel to combine and return errors             |

### ANTLR Error Handling Context

ANTLR was originally written in Java and used exception throws to indicate errors. This mechanism doesn't work well for Go since Go doesn't have checked exceptions. The current implementation uses error listeners (like `syntaxErrorListener`) to capture syntax errors. This story extends that pattern to semantic errors by collecting them in the CPLListener and surfacing them through the ParseWithLogLevel function.
