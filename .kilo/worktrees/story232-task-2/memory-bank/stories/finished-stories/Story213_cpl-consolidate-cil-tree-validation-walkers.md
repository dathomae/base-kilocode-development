# Story213: Consolidate CIL Tree Validation Walkers

## Goal

Replace the three separate recursive tree-walking validation functions in [`metrics_analysis.go`](../../src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go) with a single multi-purpose CIL tree walk that performs all validation checks in one traversal.

Currently, [`AnalyzeMetricScope`](../../src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:141-154) calls [`validateExpressionTypes`](../../src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:600), [`validateTruthiness`](../../src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:770), and [`validateBranchCompleteness`](../../src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:910) independently on the same CIL tree. Each walker traverses the full tree, which is redundant. Merging them reduces traversal overhead, centralizes validation logic, and makes future validation rules easier to add.

## References

- [Consolidation Plan: CIL Tree Validation Walkers](../plans/walker_consolidation_plan.md)
- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [metrics_analysis.go](../../../src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go)
- [metrics_analysis_test.go](../../../src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go)
- [metrics_analysis_conditional_test.go](../../../src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_conditional_test.go)

## Dependencies

- [Story187_cpl-metrics-conditionals-parse-analyze.md](Story187_cpl-metrics-conditionals-parse-analyze.md) — Conditional parse and analyze must produce valid `CILConditional`, `CILWhenClause`, `CILComparisonOp`, `CILLogicalOp`, and `CILNotOp` nodes (Done)
- [Story188_cpl-metrics-conditionals-synthesize.md](Story188_cpl-metrics-conditionals-synthesize.md) — Conditional synthesis must be in place; this story refactors the analysis layer only (Not Started)

## Dependent Stories

None.

## Tasks

1. Define `validationState` and `validateCILTree` structural shape — Completed
   a. Add the `validationState` struct with `inArithmetic bool` field near the new walker in `metrics_analysis.go` — Completed
   b. Define the `validateCILTree` entry point signature that initializes `validationState{inArithmetic: false}` and delegates to `validateCILTreeWalk` — Completed
   c. Define the `validateCILTreeWalk` recursive walker signature accepting `(cil.CILNode, *validationState) error` — Completed
   d. Verify the structural definitions compile with no implementation logic (placeholder returns) — Completed

2. Implement the consolidated per-node validation walker — Completed
   a. Run existing tests in `metrics_analysis_test.go` and `metrics_analysis_conditional_test.go` to confirm the current contract (Phase B baseline) — Completed
   b. Implement `validateCILTreeWalk` with a type-switch on `cil.CILNode` that applies all three validation rules per node type — Completed
   c. Recurse into children with appropriate `validationState` modifications (`inArithmetic=true` inside `CILBinaryOp` and `CILUnaryOp`) — Completed
   d. Return the first error encountered with preserved category prefixes (`"type error:"`, `"truthiness error:"`, `"completeness error:"`) — Completed
   e. Run existing tests to verify the consolidated walker produces identical behavior to the three separate walkers — Completed

3. Update `AnalyzeMetricScope` call site — Completed
   a. Replace the three separate validation calls in `AnalyzeMetricScope` with a single `validateCILTree(cilRoot.CILNode)` call — Completed
   b. Update error wrapping to use `"validation error in metric %s: %w"` — Completed
   c. Run existing tests to confirm the call site integration works correctly — Completed

4. Remove old walkers and verify zero regressions — Completed
   a. Delete `validateExpressionTypes`, `validateTruthiness`, `validateTruthinessArithmetic`, and `validateBranchCompleteness` from `metrics_analysis.go` — Completed
   b. Delete any now-unused helper code that was only referenced by the old walkers — Completed
   c. Run the full CPL test suite: `cd src/chronocone_planning_language && go test ./src/main/go/...` — Completed
   d. Confirm all tests pass with zero regressions — Completed

5. Update documentation — Completed
   a. Update the doc comment on `AnalyzeMetricScope` to reflect the consolidated validation call — Completed
   b. Add or update doc comments on `validateCILTree` and `validateCILTreeWalk` explaining the per-node validation rules — Completed
   c. Verify documentation accurately describes the `validationState` semantics and error categories — Completed

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests against existing contract), Phase C (implementation) — integrated within each task
2. Error messages must preserve category prefixes so that existing tests continue to pass without modification (or with minimal message-prefix updates)
3. The `isNumericNode` helper remains unchanged — it is still used by the consolidated walker for arithmetic type checks
4. No change to the public API — `validateCILTree` is unexported (package-internal)
5. All existing unit tests in `metrics_analysis_test.go` and `metrics_analysis_conditional_test.go` must pass

## Intent

- **Task 1 Intent**: Establish the structural shape of the consolidated walker before any logic is written, ensuring the API is clean and testable
- **Task 2 Intent**: Merge the three redundant traversals into one, reducing overhead and centralizing validation logic while keeping all existing behavioral contracts intact
- **Task 3 Intent**: Switch the production call site to the new walker with minimal, predictable changes
- **Task 4 Intent**: Eliminate dead code and confirm the refactor introduces no regressions across the entire analysis package
- **Task 5 Intent**: Keep documentation accurate so future maintainers understand the single-walker design and error categories

## Acceptance Criteria

1. `validateCILTree` exists as a single entry point that replaces the three separate walkers
2. `validateCILTreeWalk` traverses the CIL tree once, applying type, truthiness, and completeness checks per node
3. `AnalyzeMetricScope` calls only `validateCILTree` instead of the three legacy functions
4. Old functions (`validateExpressionTypes`, `validateTruthiness`, `validateTruthinessArithmetic`, `validateBranchCompleteness`) are removed
5. All existing tests in `metrics_analysis_test.go` and `metrics_analysis_conditional_test.go` pass without regressions
6. Error messages remain distinguishable by validation category (`type error:`, `truthiness error:`, `completeness error:`)
7. Full CPL test suite (`go test ./src/main/go/...`) passes with zero regressions
8. Doc comments on `AnalyzeMetricScope`, `validateCILTree`, and `validateCILTreeWalk` accurately describe the consolidated validation behavior

## Open Questions

None.

## Known Issues

None.

## Additional Context

This is a pure refactoring story. No new CPL syntax or CIL node types are introduced. The behavioral contract is fully captured by the existing test suite, which serves as the Logical TDD contract (Phase B) for the consolidated walker.
