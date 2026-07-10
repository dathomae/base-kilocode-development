# Story190: CPL Metrics Aliases and Where Clauses Synthesize

## Goal

Implement the Increment 5 synthesis phase for CPL metric input aliases and where clauses as part of [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md). This story covers:

- Use alias names from `CILInputRef.Alias` as Go variable names in generated compute functions.
- Generate aspect-filtering code that applies slice selection (ACTUALS/FORECAST) on input tensors.
- No new CIL node types; changes limited to variable naming and aspect filtering.

## References

- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [ERO005: Metrics Synthesis Phase — Increment 5](../specifications/design/cpl/go-synthesis-target/eros/ERO005_metrics-synthesis-phase.md)
- [Synthesis Design](../specifications/design/cpl/go-synthesis-target/synthesis.md)
- [IO Specifications Requirements](../specifications/requirements/cpl/io_specs.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story189_cpl-metrics-aliases-where-parse-analyze.md](finished-stories/Story189_cpl-metrics-aliases-where-parse-analyze.md) — parse/analyze phase must produce `AnalysisInputDefinition` with aliases and aspect filters (Done)
- [Story185_cpl-metrics-arithmetic-synthesize.md](finished-stories/Story185_cpl-metrics-arithmetic-synthesize.md) — arithmetic synthesis infrastructure must be in place (Done)

## Dependent Stories

- [Story194_cpl-metrics-dag-io-analyze.md](Story194_cpl-metrics-dag-io-analyze.md) — Increment 8 Analysis (Not Started)
- [Story195_cpl-metrics-dag-io-synthesize.md](Story195_cpl-metrics-dag-io-synthesize.md) — Increment 8 Synthesis (Not Started)

## Tasks

1. Implement alias-based variable naming in generated code — Not Started
   a. Read `CILInputRef.Alias` during code generation; when present, use alias as the Go variable name instead of the original input name — Not Started
   b. Write tests: verify generated code uses alias names for aliased inputs — Not Started
   c. Implement alias variable naming in `VisitCILInputRef`: if alias exists, emit alias as variable name — Not Started
   d. Ensure alias names do not conflict with generated variable names or Go keywords — Not Started

2. Implement aspect-filtering code generation — Not Started
   a. Read aspect filter (`ACTUALS`/`FORECAST`) from input definition during code generation — Not Started
   b. Write tests: verify generated code includes aspect slice selection on input tensors — Not Started
   c. Implement aspect filtering: when aspect filter is present, generate code to select the appropriate slice of the input tensor before use — Not Started
   d. Handle the case where no aspect filter is specified (use full tensor) — Not Started

3. End-to-end verification — Not Started
   a. Create example CPL files with aliased inputs and where clauses — Not Started
   b. Create corresponding IO specs and test parquet data with ACTUALS/FORECAST columns — Not Started
   c. Run parse → analyze → synthesize → compile → execute pipeline and verify alias names and aspect filtering work correctly — Not Started
   d. Run the full CPL test suite to confirm no regressions — Not Started

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. No new CIL node types in this increment — changes are limited to variable naming and aspect filtering logic
3. Alias names must be valid Go identifiers and not conflict with reserved words
4. Generated Go code must be compilable and runnable
5. Go functions that can fail must return `(T, error)`

## Intent

- **Task 1 Intent**: Support user-defined input aliases in generated Go code — improve readability of generated compute functions
- **Task 2 Intent**: Generate aspect filtering code — enable ACTUALS/FORECAST selection on input tensors
- **Task 3 Intent**: Prove alias and aspect filtering pipeline works end-to-end

## Acceptance Criteria

1. Generated Go code uses alias names as variable names when `as` clause is specified
2. Generated code includes aspect slice selection (`ACTUALS`/`FORECAST`) when `where` clause is specified
3. No aspect filtering occurs when no `where` clause is present
4. Alias names do not conflict with Go keywords or generated identifiers
5. Full end-to-end test: aliased inputs and where clauses produce correct output
6. All unit tests pass following Go testing best practices
7. No regressions in existing CPL test suite
