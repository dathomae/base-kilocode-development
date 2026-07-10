# Story226: CPL Refactoring — Phase 6 Test Hardening

## Goal

Add dedicated unit tests for previously untested or poorly tested areas, fix hardcoded defaults, implicit contracts, and generated error handling patterns.

## References

- [CPL Refactoring Plan](../plans/cpl_refactoring_plan.md)
- [Story225: Phase 5 Address Fragility](Story225_cpl-refactoring-phase5-address-fragility.md)

## Dependencies

- [Story225_cpl-refactoring-phase5-address-fragility.md](Story225_cpl-refactoring-phase5-address-fragility.md) — must be Done. Phase 5 stabilizes contracts and removes fragile patterns; tests added now will not need rewriting when those structures change.

## Dependent Stories

None.

## Tasks

### Task Sizing

Tasks 1, 2, and 3 each add tests to specific files and qualify as `[Small]`. Task 4 fixes hardcoded defaults in synthesis. Task 5 addresses generated error handling.

### Parallel Execution

- **Group A** (Tasks 1, 2, 3): These tasks add tests to disjoint files in different packages (`synthesis/`, `cil/`, `parser/`).
- **Group B** (Tasks 4, 5): Both modify `synthesis/synthesize.go` and `cil/` codegen; same-package risk.

### Worktree Suitability

- **Group A (Tasks 1, 2, 3):** — fully disjoint packages (`synthesis/` vs `cil/` vs `parser/`), safe for subagents
- **Group B (Tasks 4, 5): [Worktree]** — both tasks modify `synthesis/` and `cil/` files; same-package compilation risk

### Execution Order

```
[Task 1 ‖ Task 2 ‖ Task 3] → [Task 4 ‖ Task 5]
```

### Task 1: Add tests for generateInputLoadingCode and generateOutputWritingCode — Completed

**Goal**: Add dedicated unit tests for IO code generation paths.

**Test-First Development**:

a. Identify the exact functions `generateInputLoadingCode` and `generateOutputWritingCode` in `synthesis/synthesize.go`.

b. Write unit tests that invoke these functions with representative `IOSpecification` values and verify the generated code strings contain expected patterns (import statements, function calls, error handling).

c. Run `go test ./synthesis/...` — new tests must pass.

**Acceptance Criteria**:
- Unit tests exist for `generateInputLoadingCode` and `generateOutputWritingCode`
- Tests cover at least one happy path and one error path each
- All tests pass

### Task 2: Add edge-case tests for postfix codegen — Completed [Small]

**Goal**: Add dedicated tests for `GeneratePostfixTableLiteral` and `generatePostfixBranchLiteral` edge cases.

**Test-First Development**:

a. Identify `GeneratePostfixTableLiteral` and `generatePostfixBranchLiteral` in `cil/postfix_codegen.go`.

b. Write unit tests covering:
   - Deeply nested binary expressions
   - Complex conditionals with nested branches
   - Empty postfix tables
   - Single-literal tables

c. Run `go test ./cil/...` — new tests must pass.

**Acceptance Criteria**:
- Edge-case tests exist for both functions
- Tests cover deep nesting, complex conditionals, empty and single-literal tables
- All tests pass

### Task 3: Add tests for conditional parse state machine — Completed [Small]

**Goal**: Add isolated unit tests for the conditional parsing state machine.

**Test-First Development**:

a. If Task 2 of Story225 has already encapsulated `ConditionalParseState`, write tests against the public methods.

b. If not yet encapsulated, write tests against the existing `conditionalParseState` struct that verify:
   - State transitions on expression entry/exit
   - Correct postfix capture for conditions vs bodies
   - Parenthesized sub-expression handling

c. Run `go test ./parser/...` — new tests must pass.

**Acceptance Criteria**:
- Isolated unit tests for conditional parsing state machine exist
- All state transitions and edge cases covered
- All tests pass

### Task 4: Fix hardcoded float64 default type — Completed

**Goal**: Replace silent `float64` fallback with explicit error for unresolvable types.

**Test-First Development**:

a. Define a canonical type mapping registry (CPL type → Go type) in `synthesis/` or `cil/`.

b. Update `resolveGoType` to return `(string, error)` instead of `string`.

c. Update all callers of `resolveGoType` to handle the error.

d. Add tests for `resolveGoType` covering:
   - All built-in types (`Number`, `String`, `Boolean`, `Instant`, `Interval`)
   - Derived types with valid derivation paths
   - Circular derivation paths (should error)
   - Empty derivation path (should error)
   - Unknown type (should error, not default to `float64`)

e. Run `go test ./synthesis/... ./cil/...` — must pass.

**Acceptance Criteria**:
- `resolveGoType` returns an error for unresolvable types
- No silent `float64` fallback remains
- Tests cover all built-in types and error cases
- All tests pass

### Task 5: Extract consistent error handling pattern in generated code — Completed

**Goal**: Replace scattered inline error handling in generated Go with a reusable function.

**Test-First Development**:

a. Audit all generated error handling patterns in `synthesis/synthesize.go` and `cil/` (~15 instances).

b. Define a helper function string that is emitted into every generated program:
   ```go
   func check(err error, msg string) {
       if err != nil {
           fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
           os.Exit(1)
       }
   }
   ```

c. Update all code generation sites to emit `check(err, "...")` instead of inline `if err != nil` blocks.

d. Add a test that verifies the generated program contains exactly one `check` function definition and that all error sites call it.

e. Run `go test ./synthesis/... ./cil/...` — must pass.

**Acceptance Criteria**:
- Generated programs contain a single `check` helper
- All error handling sites use the helper
- Test verifies helper presence and usage count
- All tests pass

## Constraints

1. New tests must follow existing test patterns in the codebase (table-driven where appropriate)
2. `resolveGoType` must not silently default — every unresolvable type must produce an error
3. Generated code must still compile and pass integration tests
4. All existing tests must pass after each task

## Intent

Close test coverage gaps and eliminate silent failure modes so that future changes to type resolution, code generation, or parsing are caught by tests rather than producing incorrect generated programs.

## Acceptance Criteria

1. Unit tests exist for input/output code generation
2. Edge-case tests exist for postfix table/branch literals
3. Unit tests exist for conditional parsing state machine
4. `resolveGoType` returns errors for unresolvable types with no silent fallback
5. Generated programs use a consistent `check` helper for error handling
6. `go test ./...` passes with zero failures

## Notes

- Task 4 may require updating golden-file integration tests if changing `resolveGoType` to return an error affects generated code for edge cases.
- Task 5 reduces generated code size and improves readability, but the generated `check` function must not conflict with any user-defined identifier.
