# Story112: Refactor CPL Tests to Go Best Practices

## Goal

Refactor the Go test code in the CPL parser to follow Go testing best practices as defined in the [resources/go/testing-guide.md](../resources/go/testing-guide.md). This includes reorganizing test files to be co-located with source code, introducing table-driven tests, enabling parallel execution, adding testify assertions, and fixing helper function conventions.

## References

- [Go Testing Best Practices Guide](../resources/go/testing-guide.md)
- [Lessons Learned - Consult Go Testing Guide](../lessons-learned.md)
- [plans/go-testing-refactoring-analysis.md](../../plans/go-testing-refactoring-analysis.md) - Detailed analysis of deviations
- [CPL Parser README](../../src/chronocone_planning_language/README.md)

## Dependencies

- [Story111_remove-dimensions-from-cpl.md](Story111_remove-dimensions-from-cpl.md) - Not Started
  - This story removes dimensions-related test files that would otherwise need refactoring
  - Tests in general_test.go that reference dimensions will need updating/removing

## Dependent Stories

None

## Tasks

1. Add testify dependency to go.mod - Completed
   a. Add `github.com/stretchr/testify` to go.mod dependencies - Completed
   b. Run `go mod tidy` to update go.sum - Completed
   c. Verify dependency is correctly resolved - Completed

2. Refactor AST tests to use table-driven pattern - Completed
   a. Refactor `src/test/go/ast/units_test.go` - Completed
   - Convert `TestAddUnit` and `TestUnitDeclWithParameters` to table-driven format - Completed
   - Use testify assertions (require/assert) - Completed
   - Enable parallel execution with `t.Parallel()` - Completed
     b. Refactor `src/test/go/ast/general_test.go` - Completed
   - Convert `TestNewCPLDocument` and `TestMultipleUnits` to table-driven format - Completed
   - Remove dimension-related assertions (after Story111) - Completed
   - Use testify assertions - Completed
   - Enable parallel execution - Completed

3. Refactor parser tests to use table-driven pattern - Completed
   a. Refactor `src/test/go/parser/units_test.go` - Completed
   - Consolidate unit tests into single table-driven test `TestParseUnits` - Completed
   - Use testify assertions - Completed
   - Enable parallel execution - Completed
     b. Refactor `src/test/go/parser/general_test.go` - Completed
   - Convert `TestParseEmptyDocument`, `TestParseMultipleSections`, and `TestParseInvalidSyntax` to table-driven format - Completed
   - Remove dimension-related tests (after Story111) - Completed
   - Use testify assertions - Completed
   - Enable parallel execution - Completed
   - Fix `TestParseWithLogLevel_FullIntegration` to not use dimensions - Completed
     c. Refactor `src/test/go/parser/examples_test.go` - Completed
   - Add `t.Helper()` to `getExamplesDir` function - Completed
   - Enable parallel execution for example tests - Completed
   - Use testify assertions where appropriate - Completed

4. Refactor logging tests - Completed
   a. Refactor `src/test/go/logging/parser_logging_test.go` - Completed
   - Complete the tests (`TestParseLoggingConfig_ValidLogLevels`, `TestParseLoggingConfig_InvalidLogLevel`) - Completed
   - Convert to table-driven pattern where applicable - Completed
   - Use testify assertions - Completed
   - Enable parallel execution - Completed

5. Reorganize test files according to Go best practices - Completed
   a. Move `src/test/go/ast/units_test.go` to `src/main/go/ast/cpl_units_test.go` - Completed
   b. Move `src/test/go/ast/general_test.go` to `src/main/go/ast/cpl_ast_test.go` - Completed
   c. Move `src/test/go/parser/units_test.go` to `src/main/go/parser/cpl_parser_units_test.go` - Completed
   d. Move `src/test/go/parser/general_test.go` to `src/main/go/parser/cpl_parser_test.go` - Completed
   e. Move `src/test/go/parser/examples_test.go` to `src/main/go/parser/cpl_parser_examples_test.go` - Completed
   f. Move `src/test/go/logging/parser_logging_test.go` to `src/main/go/parser/cpl_parser_logging_test.go` - Completed
   g. Update import paths in all moved test files - Completed
   h. Remove empty `src/test/go` directory structure - Completed
   i. Update `README.md` to reflect new test file organization - Completed
   j. Update `toc.md` to reflect new file structure - Completed

6. Verify all tests pass - Completed
   a. Run `./gradlew test` to execute full test suite - Completed
   b. Verify no regressions in test coverage - Completed
   c. Verify parallel execution works correctly - Completed

## Test-First Development

This story follows the Logical TDD Lifecycle. Since we are refactoring existing tests:

1. **Phase A (Structural Definition)**: The existing tests define the baseline structure
2. **Phase B (Logical Tests)**: Refactored tests must continue to pass with the same behavior
3. **Phase C (Implementation)**: Apply table-driven pattern, testify, and parallel execution
4. **Phase D (Verification)**: Verify all tests pass and behavior is preserved

The key principle is that we refactor the tests without changing their behavior. Each refactored test should validate the same functionality as before.

## Constraints

- All existing tests must continue to pass after refactoring
- Test behavior must remain unchanged (same assertions, same validation)
- Must use Gradle build (`./gradlew test`) not `go test` directly
- Parallel tests must use proper variable capture (`tc := tc`)
- Test files must follow Go best practices by being co-located with source code
- Test files must be named `<source_file>_test.go` (e.g., `cpl_ast.go` → `cpl_ast_test.go`)

## Intent

The intent is to bring the CPL parser test code into alignment with Go community best practices. This improves:

1. **Maintainability**: Table-driven tests are easier to extend and modify
2. **Performance**: Parallel test execution reduces test runtime
3. **Readability**: Testify assertions are more expressive and produce better error messages
4. **Consistency**: Following standard Go patterns makes the codebase more approachable for Go developers

## Acceptance Criteria

1. All test files use table-driven patterns where applicable
2. All tests use testify assertions (require for fail-fast, assert for multiple checks)
3. All independent tests use `t.Parallel()` for parallel execution
4. Helper functions use `t.Helper()` for proper error reporting
5. All tests pass with `./gradlew test`
6. No regression in test coverage
7. testify is properly declared as a dependency in go.mod
8. All test files are co-located with the source code they test
9. Test files follow the `<source>_test.go` naming convention
10. No test files remain in `src/test/go/` directory structure

## Notes

### Before/After Example

**Before:**

```go
func TestParseUnitDeclaration(t *testing.T) {
    input := `units
\t\tunit = Count`

    doc, err := parser.Parse(input)
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
    if len(doc.Units) != 1 {
        t.Fatalf("Expected 1 unit, got %d", len(doc.Units))
    }
    // ... more assertions
}
```

**After:**

```go
func TestParseUnits(t *testing.T) {
    t.Parallel()

    testCases := []struct {
        name          string
        input         string
        expectedUnits []ast.UnitDecl
        expectError   bool
    }{
        {
            name:          "single unit without parameters",
            input:         "units\n\t\tunit = Count",
            expectedUnits: []ast.UnitDecl{{Name: "unit", TypeName: "Count"}},
        },
        // ... more cases
    }

    for _, tc := range testCases {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            doc, err := parser.Parse(tc.input)
            require.NoError(t, err)
            assert.Equal(t, tc.expectedUnits, doc.Units)
        })
    }
}
```

### Files to Modify

#### Refactoring (Already Completed)

| File (Current Location)                                                       | Status    | Changes                                            |
| ----------------------------------------------------------------------------- | --------- | -------------------------------------------------- |
| `src/chronocone_planning_language/go.mod`                                     | Completed | Add testify dependency                             |
| `src/chronocone_planning_language/src/test/go/ast/units_test.go`              | Completed | Table-driven, testify, parallel                    |
| `src/chronocone_planning_language/src/test/go/ast/general_test.go`            | Completed | Table-driven, testify, parallel, remove dimensions |
| `src/chronocone_planning_language/src/test/go/parser/units_test.go`           | Completed | Table-driven, testify, parallel                    |
| `src/chronocone_planning_language/src/test/go/parser/general_test.go`         | Completed | Table-driven, testify, parallel, remove dimensions |
| `src/chronocone_planning_language/src/test/go/parser/examples_test.go`        | Completed | Add t.Helper(), testify, parallel                  |
| `src/chronocone_planning_language/src/test/go/logging/parser_logging_test.go` | Completed | Complete tests, table-driven, testify, parallel    |

#### File Reorganization (Pending)

| Current Location                             | New Location                                     | Action |
| -------------------------------------------- | ------------------------------------------------ | ------ |
| `src/test/go/ast/units_test.go`              | `src/main/go/ast/cpl_units_test.go`              | Move   |
| `src/test/go/ast/general_test.go`            | `src/main/go/ast/cpl_ast_test.go`                | Move   |
| `src/test/go/parser/units_test.go`           | `src/main/go/parser/cpl_parser_units_test.go`    | Move   |
| `src/test/go/parser/general_test.go`         | `src/main/go/parser/cpl_parser_test.go`          | Move   |
| `src/test/go/parser/examples_test.go`        | `src/main/go/parser/cpl_parser_examples_test.go` | Move   |
| `src/test/go/logging/parser_logging_test.go` | `src/main/go/parser/cpl_parser_logging_test.go`  | Move   |
| `src/test/go/`                               | (remove empty directories)                       | Delete |
| `src/chronocone_planning_language/README.md` | -                                                | Update |
| `src/chronocone_planning_language/toc.md`    | -                                                | Update |

### Import Path Changes Required

When moving test files from `src/test/go/<package>/` to `src/main/go/<package>/`, the import paths stay the same since they reference the full module path. However, some tests import from other packages within the module that may now be at different relative positions.

**Key import updates needed:**

1. **In ast tests** (`cpl_ast_test.go`, `cpl_units_test.go`):
   - Import `"github.com/chronocone/chronocone_planning_language/src/main/go/ast"` stays the same
   - These should use `package ast_test` to maintain black-box testing

2. **In parser tests** (all `cpl_parser_*_test.go`):
   - Import `"github.com/chronocone/chronocone_planning_language/src/main/go/ast"` stays the same
   - Import `"github.com/chronocone/chronocone_planning_language/src/main/go/parser"` stays the same
   - These should use `package parser_test` to maintain black-box testing

3. **Example file path resolution** (in `cpl_parser_examples_test.go`):
   - The `getExamplesDir` function uses `runtime.Caller(0)` to find the test file location
   - Currently navigates: `testDir → .. → .. → .. → ..` to reach project root
   - After move: needs to navigate from `src/main/go/parser/` to project root
   - New path: `filepath.Join(testDir, "..", "..", "..", "..", "examples")` becomes `filepath.Join(testDir, "..", "..", "..", "examples")`

### New Directory Structure After Reorganization

```
src/
├── main/
│   ├── antlr/                    # ANTLR4 grammar files (.g4)
│   └── go/
│       ├── ast/                  # AST types AND their tests (co-located)
│       │   ├── cpl_ast.go
│       │   ├── cpl_ast_test.go           # was: src/test/go/ast/general_test.go
│       │   ├── cpl_units.go
│       │   └── cpl_units_test.go         # was: src/test/go/ast/units_test.go
│       └── parser/               # Parser AND its tests (co-located)
│           ├── antlr/            # ANTLR-generated Go code
│           ├── cpl_parser.go
│           ├── cpl_parser_test.go        # was: src/test/go/parser/general_test.go
│           ├── cpl_parser_units_test.go  # was: src/test/go/parser/units_test.go
│           ├── cpl_parser_examples_test.go # was: src/test/go/parser/examples_test.go
│           ├── cpl_parser_logging_test.go  # was: src/test/go/logging/parser_logging_test.go
│           ├── cpl_listener.go
│           └── cpl_units_listener.go
└── examples/                     # Sample CPL files
```
