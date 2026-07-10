# Story246: Lookups Increment 2 — Synthesize

## Goal

Generate Go synthesis code for temporal lookups (lookups with `grain` clauses). When `CILLookupNode.HasGrain` is true, the synthesized code includes grain dimension names in the tensor's dimensional structure via `SetDimensions()`, registers grain dimension metadata on the `MetricComputation` via `SetGrainDimensions()`, and emits a `resolveGrainIndex` runtime helper for multi-grain lookups (e.g., `grain Week, Month`). Includes a `with_data/lookups_temporal` example demonstrating end-to-end temporal lookup behavior. This is Increment 2 of lookups synthesis, extending the Increment 1 synthesis infrastructure for timeless lookups.

## References

- [ERO006: Lookups Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md) — Increment 2 temporal lookup code generation
- [Story246 plan](../plans/Story246_lookups-increment2-synthesize.md)
- [Lookups Design](../specifications/design/lookups.md)
- [CPL Requirements — Lookups](../specifications/requirements/cpl/lookups-syntax.md)

## Dependencies

- [Story245_lookups-increment2-analyze.md](Story245_lookups-increment2-analyze.md) — _Not Started_ — provides `CILLookupNode.HasGrain` and `CILLookupNode.Grain` fields
- [Story243_lookups-increment1-synthesize.md](Story243_lookups-increment1-synthesize.md) — _Not Started_ — provides base lookup loading code, `WalkCILLookupNode` walker, and `generateLookupLoading` function

## Dependent Stories

None — this is the final Increment 2 story for lookups.

## Tasks

### Task Sizing

- **Task 1** (walker temporal detection): `[Small]` — modifies at most 2 files (walker source + test), adds a single `if node.HasGrain` block to an existing walker method, follows the existing walker switch pattern, single concern.
- **Task 2** (generateLookupLoading extension): Not `[Small]` — 6 subtasks extending a single function with multiple code-generation concerns (SetDimensions, SetGrainDimensions, ordering, mixed timeless/temporal handling). Requires reasoning about grain dimension semantics and the distinction between temporal and timeless lookups. Decomposition review confirmed no natural seam — all changes are to one function and splitting would create artificial intermediate artifacts.
- **Task 3** (resolveGrainIndex helper): Not `[Small]` — introduces a new runtime resolution helper that must match cell time coordinates against grain dimension intervals. Requires understanding time coordinate semantics, grain interval matching, and multi-grain selection logic.
- **Task 4** (with_data example): Not `[Small]` — creates multiple file types (CPL, IO spec, parquet data, expected output) and requires integration with the full parse→analyze→synthesize→run pipeline. While a single example, the breadth across file formats and tools exceeds the narrow-scope threshold.
- **Task 5** (regression): `[Small]` — runs existing test suites, no code changes.

### Parallel Execution

None — all tasks have sequential dependencies.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4 → Task 5
```

- Task 1 (walker detection) must complete first — subsequent tasks reference `scope.temporalLookups` populated by the walker.
- Task 2 (generateLookupLoading) depends on Task 1 — it consumes the temporal lookups map.
- Task 3 (resolveGrainIndex) depends on Task 2 — the grain infrastructure in `generateLookupLoading` must be in place before the resolution helper is generated.
- Task 4 (example) depends on Tasks 1–3 — the end-to-end example requires the full synthesis pipeline for temporal lookups.
- Task 5 (regression) runs last to confirm no existing functionality is broken.

### Reintegration

None — no parallel groups.

### Task 1: Update WalkCILLookupNode for temporal lookups — Not Started [Small]
   a. Add temporal lookup detection in `WalkCILLookupNode`: if `node.HasGrain` is true, store the node in `scope.temporalLookups[node.LookupName]`. This is a single conditional block added after the existing lookup data source registration logic. — Not Started
   b. Write unit tests in `src/main/go/synthesis/cil_tree_walker_test.go`: walk a `CILLookupNode` with `HasGrain = true` and verify the temporal lookups map contains the node; walk a `CILLookupNode` with `HasGrain = false` and verify the temporal lookups map does not contain the node. — Not Started
   c. Run `go test ./src/main/go/synthesis/... -count=1 -run TestWalkCILLookupNode` and verify all walker tests pass. — Not Started

### Task 2: Extend generateLookupLoading for temporal lookups — Not Started
   a. When `HasGrain` is true for a lookup, include grain dimension names in the tensor's `SetDimensions()` call. Grain dimensions are appended after concept dimensions in the dimension list. — Not Started
   b. Emit `computation.SetGrainDimensions(...)` for temporal lookups: generate a call that passes the grain dimension names to `MetricComputation.SetGrainDimensions()`. This call is emitted inside the `if .HasGrain` template branch, immediately after the `NewMetricComputationFromTensor` call. — Not Started
   c. Ensure that when a compilation unit contains both temporal and timeless lookups, the generated code correctly handles both: timeless lookups omit the grain dimension code paths (no `SetGrainDimensions` call, no grain names in `SetDimensions`). — Not Started
   d. Verify that grain dimensions appear after concept dimensions, not interleaved or before. — Not Started
   e. Write unit tests in `src/main/go/synthesis/lookup_synthesis_test.go`:
      - Test with a single temporal lookup (one grain dimension): verify emitted Go code includes the grain dimension in `SetDimensions` and emits `SetGrainDimensions`.
      - Test with a multi-grain temporal lookup (e.g., `grain Week, Month`): verify both grain names appear in `SetDimensions` and `SetGrainDimensions`.
      - Test with a mixed compilation unit (one timeless lookup + one temporal lookup): verify the timeless lookup's code omits grain references while the temporal lookup's code includes them.
      - Test with only timeless lookups: verify no grain-related code is emitted (no regression from Increment 1). — Not Started
   f. Run `go test ./src/main/go/synthesis/... -count=1` and verify all synthesis tests pass. — Not Started

### Task 3: Implement multi-grain resolution helper generation — Not Started
   a. Emit a `resolveGrainIndex` helper function in the generated Go code when temporal lookups exist in the compilation unit. The helper iterates over a lookup's grain dimensions, matches the requesting cell's time coordinate against each grain dimension's interval, and returns the index of the matching grain row. — Not Started
   b. The helper is generated once per generated program (not per lookup). If no temporal lookups exist in the compilation unit, the helper is not emitted. — Not Started
   c. Write unit tests in `src/main/go/synthesis/lookup_synthesis_test.go`:
      - Test with a multi-grain lookup (Week + Month): verify `resolveGrainIndex` is emitted and contains iteration over grain dimensions.
      - Test with a single-grain temporal lookup: verify `resolveGrainIndex` is emitted and handles the single grain case.
      - Test with only timeless lookups: verify `resolveGrainIndex` is not emitted.
      - Test time coordinate matching logic: verify the generated helper correctly selects the matching grain index for different cell time values. — Not Started
   d. Run `go test ./src/main/go/synthesis/... -count=1` and verify all synthesis tests pass. — Not Started

### Task 4: Create with_data/lookups_temporal example — Not Started
   a. Create the directory `examples/with_data/lookups_temporal/` with a `core.cpl` file declaring: an `exchange_rate` lookup with dimensions `[Currency, Currency]`, grain `Week, Month`, and celltype `ConversionRate`. Include the necessary concept and measure declarations. — Not Started
   b. Create a Lookup IO Spec JSON file that maps the `exchange_rate` lookup name to a parquet file path, specifying the column schema (source currency, target currency, Week grain value, Month grain value, conversion rate value). — Not Started
   c. Create a parquet data file with sample exchange rates for multiple currency pairs, each with Week and Month grain values. Use the existing parquet generation tooling. — Not Started
   d. Create an `.expected.json` file with the expected output when the example is compiled and run. Verify that temporal lookup data loads correctly and is indexed by time dimensions. — Not Started
   e. Run the full pipeline end-to-end: `./gradlew run` (or equivalent) to parse → analyze → synthesize → compile → execute the generated program. Verify the output matches `.expected.json`. — Not Started

### Task 5: Verification and regression — Not Started [Small]
   a. Verify that timeless lookups (no grain) from Increment 1 continue to synthesize and run correctly with no regressions. — Not Started
   b. Run `go test ./src/main/go/... -count=1` from `src/chronocone_planning_language/` and confirm all packages pass. — Not Started

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task above, tests are written before implementation and the test-and-implement cycle is integrated within each task rather than broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- Must follow the Logical TDD Lifecycle — write tests before implementation changes.
- Must not break existing timeless lookup functionality from Increment 1.
- Grain dimensions must appear after concept dimensions in the tensor dimension list.
- The `SetGrainDimensions()` call must use the same grain dimension name strings declared in CPL.
- Multi-grain resolution must iterate over grain dimensions and match the requesting cell's time coordinate.
- The `resolveGrainIndex` helper must only be emitted when temporal lookups exist in the compilation unit.
- Must use the existing `MetricComputation` interface and `ChronoconeComputationController` infrastructure from connector synthesis.
- All Go source paths are relative to `src/chronocone_planning_language/` (the `source_root`). Go commands use fully-qualified package paths from `source_root`.

## Intent

1. **Task 1 (Walker temporal detection)**: Enable the CIL tree walker to differentiate temporal lookups from timeless lookups by populating a dedicated `scope.temporalLookups` map. This map is consumed by subsequent synthesis tasks to determine whether grain code paths are needed.

2. **Task 2 (generateLookupLoading extension)**: Extend the code generation function to produce grain-dimensional tensor structures and `SetGrainDimensions()` registration calls for temporal lookups. Ensure the generated code correctly handles mixed compilation units containing both temporal and timeless lookups.

3. **Task 3 (Multi-grain resolution helper)**: Generate a runtime helper function that resolves which grain row to use when a lookup has multiple grain measures. The helper matches the requesting metric cell's time coordinate against each grain dimension's interval and returns the correct index.

4. **Task 4 (with_data example)**: Demonstrate and validate the full temporal lookup pipeline — from CPL declaration through parquet data loading, tensor construction with grain dimensions, and runtime grain resolution — with a concrete, verifiable example.

5. **Task 5 (Regression)**: Guarantee that the Increment 2 changes do not break Increment 1 (timeless lookup) functionality or any other existing CPL features.

## Acceptance Criteria

1. `WalkCILLookupNode` populates `scope.temporalLookups` when `CILLookupNode.HasGrain` is true.
2. Generated Go code for temporal lookups includes grain dimensions in `tensor.SetDimensions()` after concept dimensions.
3. Generated Go code emits `computation.SetGrainDimensions(...)` for temporal lookups.
4. Mixed compilation units (temporal + timeless lookups) generate correct code for both lookup types.
5. `resolveGrainIndex` helper is emitted when temporal lookups exist and omitted when they do not.
6. Multi-grain lookups (e.g., `grain Week, Month`) generate correct resolution code that iterates over grain dimensions.
7. The `with_data/lookups_temporal` example parses, analyzes, synthesizes, compiles, runs, and produces output matching `.expected.json`.
8. All existing tests pass with zero regressions.
9. Timeless lookups from Increment 1 continue to work unchanged.

## Footer

### Open Questions

None

### Known Issues

None
