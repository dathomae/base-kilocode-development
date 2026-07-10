# Story245: Lookups Increment 2 — Grain Analysis

## Goal

Extend the analysis phase for lookup declarations to validate grain clauses: resolve grain measure names against the `Measures` map, validate that each grain measure's type is `Interval` or derived from `Interval`, reject empty and duplicate grain clauses, and populate the extended `LookupInfo` and `CILLookupNode` structures with grain fields. This story is Increment 2 of [Epic006](../../epics/Epic006_cpl-lookups-implementation.md).

## References

- [ERO006: Lookups Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md) — Increment 2 grain validation and structure extensions
- [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md)
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md)

## Dependencies

- [Story244_lookups-increment2-parse.md](Story244_lookups-increment2-parse.md) — Needs `ParserLookupDefinition.HasGrain` and `Grain` fields from the parsing phase
- [Story242_lookups-increment1-analyze.md](Story242_lookups-increment1-analyze.md) — Extends `AnalyzeLookupScope`, `LookupInfo`, and `CILLookupNode` created in Increment 1

## Dependent Stories

- Story246 (increment 3 — lookup integration with metrics) — Needs the `LookupInfo.Grain` and `CILLookupNode.Grain` fields
- Story247 (increment 4 — lookups in aggregator expressions) — Needs the Lookups map with populated grain fields
- Story248 (increment 5 — lookups in spreader expressions) — Needs the Lookups map with populated grain fields

## Tasks

### Task Sizing

- **Task 3** is annotated **[Small]** — it adds test functions exclusively in a single test file, following existing test patterns. The limited scope and single-file footprint make it suitable for the lightweight `small-code-for-story-implementor` agent.
- **Task 1** is not Small — it touches 3 files (two source struct files and a shared test file). While the field additions are mechanically simple, the 3-file coordinated change exceeds the 2-file threshold.
- **Task 2** is not Small — it spans 9 subtasks across potentially 3+ files (analysis function, type helper, test file), introduces a recursive helper (`isIntervalType`), and requires reasoning about the type system derivation chain. The architect-for-story-planning flagged this as borderline (9 subtasks) but kept as a single task due to the lack of a natural decomposition seam — grain validation is one cohesive concern.
- **Task 4** is not Small — it creates multiple CPL example files across two subdirectories.

### Parallel Execution

None — all tasks have sequential dependencies.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

### Reintegration

None — no parallel groups.

### Task 1: Extend LookupInfo and CILLookupNode with grain fields — Not Started

Add `Grain` and `HasGrain` fields to both `LookupInfo` and `CILLookupNode`, and verify existing tests continue to pass with zero-value defaults.

1. a. In `src/main/go/analysis/lookup_info.go`, add fields to the `LookupInfo` struct:
   ```go
   Grain    []*AnalysisMeasureDefinition  // Resolved grain measure references (nil if !HasGrain)
   HasGrain bool                         // True when lookup has a temporal grain dimension
   ```
   — Not Started
2. b. In `src/main/go/analysis/cil_lookup_node.go`, add fields to the `CILLookupNode` struct:
   ```go
   Grain    []string  // Grain measure names (empty if no grain)
   HasGrain bool      // Whether grain clause was present
   ```
   — Not Started
3. c. Write unit tests in `src/main/go/analysis/analyze_lookups_test.go` that verify:
   - `LookupInfo.Grain` is `nil` and `LookupInfo.HasGrain` is `false` for a lookup without grain
   - `CILLookupNode.Grain` is `nil` (or empty) and `CILLookupNode.HasGrain` is `false` for a lookup without grain
   — Not Started
4. d. Run `go test ./src/main/go/analysis/...` and verify the package compiles and all existing tests pass.
   — Not Started

**Acceptance Criteria:**
- `LookupInfo` has `Grain []*AnalysisMeasureDefinition` and `HasGrain bool` fields
- `CILLookupNode` has `Grain []string` and `HasGrain bool` fields
- Existing lookup tests (from Story242) pass without modification
- Zero-value defaults are correct: `Grain` is nil, `HasGrain` is false

**Intent:** These fields carry grain information from the parser definition through analysis to CIL generation. Increment 1 did not include grain support; Increment 2 adds the fields so the grain validation logic (Task 2) has a place to store its results.

### Task 2: Implement grain validation in AnalyzeLookupScope — Not Started

Implement the grain clause validation, measure resolution, and CIL node grain population within `AnalyzeLookupScope`. This task follows Test-First Development — write a test, implement the corresponding logic, verify, and repeat. Grain validation activates only when `parserDef.HasGrain` is `true`. Timeless lookups (no grain) must pass through unchanged.

1. a. Write a test for the **empty grain clause** error: a `ParserLookupDefinition` with `HasGrain=true` and `Grain=[]` (empty slice) must return an error "grain clause present but no grain measures specified". Implement the check in `AnalyzeLookupScope`. Verify the test passes.
   — Not Started

2. b. Write a test for the **duplicate grain measure names** error: a `ParserLookupDefinition` with `HasGrain=true` and `Grain=["Daily", "Daily"]` (duplicate name) must return an error indicating duplicate grain measure names. Implement a `uniqueStrings` helper and the duplicate check. Verify the test passes.
   — Not Started

3. c. Write a test for **grain measure resolution**: a `ParserLookupDefinition` with `HasGrain=true` and a `Grain` name that does not exist in the `Measures` map must return an error "grain measure ... not found". Implement the resolution logic that looks up each grain name in `a.measures`. Verify the test passes.
   — Not Started

4. d. Write a test for the **`isIntervalType` helper** in `src/main/go/analysis/types_analysis.go`: a type whose `DerivationPath` contains `"Interval"` must return `true`; a type whose derivation path does NOT contain `"Interval"` must return `false`; a `nil` type must return `false`. Implement `isIntervalType` as a helper that checks the `DerivationPath` slice for `"Interval"`. Verify the test passes.
   — Not Started

5. e. Write a test for **non-Interval grain type** error: a `ParserLookupDefinition` with `HasGrain=true` and a grain measure whose type is `Number` or derived from `Number` (not `Interval`) must return an error indicating the measure's type is not Interval or derived from Interval. Implement the type check using `isIntervalType` from subtask d. Verify the test passes.
   — Not Started

6. f. Write a test for **valid grain population**: a `ParserLookupDefinition` with `HasGrain=true` and valid grain measure names whose types derive from `Interval` must produce a `LookupInfo` with `HasGrain=true` and `Grain` containing the resolved `*AnalysisMeasureDefinition` references. Implement the population of `info.Grain` and `info.HasGrain`. Verify the test passes.
   — Not Started

7. g. Write a test for **CILLookupNode grain population**: verify that when `HasGrain=true` and grain measures are resolved, the generated `CILLookupNode` has `HasGrain=true` and `Grain` containing the resolved measure names (strings). Implement the population of `cilNode.Grain` and `cilNode.HasGrain`. Verify the test passes.
   — Not Started

8. h. Write tests for **multi-grain lookups**: a lookup with two grain measures (e.g., `Daily`, `Monthly`) must resolve both, populate both in `LookupInfo.Grain` and `CILLookupNode.Grain`, and validate both for Interval-derived types. Verify the tests pass with the existing implementation.
   — Not Started

9. i. Run `go test ./src/main/go/analysis/... -count=1` and confirm all new and existing tests pass.
   — Not Started

**Acceptance Criteria:**
- Empty grain clause (`HasGrain=true`, empty `Grain`) produces an error with the lookup name
- Duplicate grain measure names produce an error with the lookup name
- Non-existent grain measure names produce an error with the lookup name and the missing name
- Grain measures with non-Interval types produce an error with the measure name and its actual type
- Valid grain measures (Interval-derived) resolve correctly and populate both `LookupInfo.Grain` and `CILLookupNode.Grain`
- `isIntervalType` correctly walks the type derivation chain
- Timeless lookups pass through `AnalyzeLookupScope` without modifications

**Intent:** Grain validation ensures that when a lookup declares a temporal grain dimension, all grain measures exist, have Interval-derived types, and are free of duplicates. The `isIntervalType` helper enables reuse by later increments that may need to validate Interval-derived types in other contexts.

**Sizing note:** This task has 9 subtasks across potentially 3 source files. The architect-for-story-planning flagged it as borderline but found no natural decomposition seam — grain validation is a single concern where splitting would produce artificially incomplete validation with no independent verification benefit. Route to `code-for-story-implementor`.

### Task 3: Verify backward compatibility and edge cases — Not Started [Small]

Add tests to confirm that timeless lookups (no grain) and mixed-compilation units (some lookups with grain, some without) continue to behave correctly.

1. a. Write tests in `src/main/go/analysis/analyze_lookups_test.go` that verify timeless lookups (no grain clause) still analyze successfully and produce `LookupInfo` instances with `HasGrain=false` and `Grain=nil`.
   — Not Started
2. b. Write tests that verify lookups with grain and lookups without grain can coexist in the same compilation unit — both types of `LookupInfo` and `CILLookupNode` appear correctly in the analysis results.
   — Not Started
3. c. Write tests that verify all grain-related error messages include the lookup name and source location, consistent with existing error message conventions in the analysis package.
   — Not Started
4. d. Run `go test ./src/main/go/analysis/... -count=1` and confirm the full test suite passes with no regressions.
   — Not Started

**Acceptance Criteria:**
- Timeless lookups produce identical analysis results as before the grain changes
- Mixed compilation units produce both grain-bearing and grain-free lookup results correctly
- All error messages follow existing conventions (lookup name, source location, descriptive message)
- Full analysis test suite passes

**Intent:** This task ensures that Increment 2 grain validation does not break any existing behavior. Timeless lookups must remain indistinguishable from their Increment 1 behavior. Error messages must be consistent with the established analysis package conventions.

### Task 4: Create parse_and_analyze_only examples with grain — Not Started

Create CPL example files that exercise grain-bearing lookup analysis for both successful and error cases in the `parse_and_analyze_only` test directory.

1. a. In `examples/parse_and_analyze_only/without_errors/`, create CPL files that define lookups with grain and parse+analyze successfully:
   - A lookup with a single grain measure (simplest temporal lookup)
   - A lookup with multiple grain measures (e.g., both daily and monthly resolution)
   - Each file must include any prerequisite declarations (types, measures, concepts) needed for a self-contained parse-and-analyze test
   — Not Started
2. b. In `examples/parse_and_analyze_only/with_errors/`, create CPL files that produce specific analysis errors:
   - A lookup whose grain measure references a non-existent measure name
   - A lookup whose grain measure has a non-Interval type (e.g., `Number` type)
   - A lookup with an empty grain clause (`grain()`)
   - A lookup with duplicate grain measure names in the same grain clause
   — Not Started
3. c. Verify the examples using the existing `parse_and_analyze_only` test infrastructure: run the standard test command for parse-and-analyze examples and confirm all new examples produce their expected results (success for `without_errors`, specific error for `with_errors`).
   — Not Started

**Acceptance Criteria:**
- At least one successful multi-grain lookup example in `without_errors/`
- Error examples covering all four grain validation rules (missing measure, non-Interval type, empty clause, duplicates) in `with_errors/`
- All examples are self-contained (include prerequisite type, measure, and concept declarations)
- Existing parse-and-analyze tests pass alongside the new examples

**Intent:** Example files serve as living documentation of the grain validation behavior and ensure the full parse→analyze pipeline works end to end. The error examples validate each grain validation rule from the perspective of CPL source text.

## Test-First Development

This story follows the Logical TDD Lifecycle. For each coding task, tests are written before the corresponding implementation. Each subtask in Task 2 pairs a test with its implementation so that every validation rule is verified before the next rule is added. Do not write tests after implementation is complete — interleave test and implementation within each subtask.

## Constraints

- Must not modify the `ParserLookupDefinition` struct (owned by the parsing phase, Story244)
- Must follow existing analysis package conventions: error messages include entity name and source location, helper functions follow established naming patterns
- The `isIntervalType` helper must handle the type derivation chain correctly: a type derived from `Interval` (e.g., `Daily = Interval(...)`, or `SmallDaily = Daily(...)`) must be recognized as an Interval-derived type
- Zero-value defaults (`Grain=nil`, `HasGrain=false`) for timeless lookups must remain correct
- All Go test and build commands must use fully-qualified package paths from `source_root`: `go test ./src/main/go/analysis/...`
- Increment 1 behavior must be fully preserved — no regressions in existing tests

## Intent

This story extends the analysis phase to validate the grain clause in lookup declarations. When the parser produces a `ParserLookupDefinition` with `HasGrain=true`, the analysis phase resolves each grain measure name to an `AnalysisMeasureDefinition`, validates that each has an Interval-derived type, checks for empty and duplicate grain measures, and populates the `LookupInfo` and `CILLookupNode` structures. The `isIntervalType` helper provides a reusable way to walk the type derivation chain. Timeless lookups are unaffected.

## Acceptance Criteria

1. `LookupInfo.Grain` contains resolved `*AnalysisMeasureDefinition` references when `HasGrain` is true
2. `LookupInfo.HasGrain` correctly reflects the presence of a grain clause
3. `CILLookupNode.Grain` contains the resolved grain measure names (strings) when `HasGrain` is true
4. `CILLookupNode.HasGrain` correctly reflects the presence of a grain clause
5. Grain measure names must exist in the `Measures` map (unknown measures produce an error)
6. Grain measure types must be `Interval` or derived from `Interval` (non-Interval types produce an error)
7. An empty grain clause (`HasGrain=true`, `Grain` is empty) produces an error with a clear message
8. Duplicate grain measure names in a single lookup's grain clause produce an error
9. Timeless lookups (no grain clause) continue to work without any change in behavior
10. Error messages include the lookup name and the relevant measure or type name for diagnostic clarity
