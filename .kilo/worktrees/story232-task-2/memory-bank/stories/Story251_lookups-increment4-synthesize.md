# Story251: Lookups Increment 4 — Synthesize

## Goal

Generate lookup access code within aggregator computations, ensuring lookup data is loaded before aggregator execution and that lookup data is not deregistered until all consuming metrics complete. This story enables the synthesis phase to emit `controller.GetMetricComputation(lookupName).GetCell(...)` calls inside aggregator compute functions, orders lookup loading before aggregator execution in the generated main function, and verifies correctness through a `with_data` example demonstrating an aggregator that applies a weighting table from a lookup. No new synthesis infrastructure is added at the lookup level — lookup loading infrastructure already exists from Story243; this story extends the aggregator synthesis to use it.

This story is part of the lookups incremental implementation plan. See [ERO006: Lookups Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md) and [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md) for the full plan.

## References

- [requirements](../specifications/requirements/requirements.md)
- [design](../specifications/design/design.md)
- [ERO006: Lookups Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md)
- [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md)
- [Source root: src/chronocone_planning_language/](../../src/chronocone_planning_language/)

## Dependencies

- **Epic007 (Aggregators) at Increment 2 or above** — aggregator CIL tree walker and synthesis infrastructure including `GenerateComputeFunction` support for aggregator CIL nodes and aggregator-specific compute function generation must exist. The exact file paths for aggregator synthesis depend on Epic007's implementation.
- [Story250_lookups-increment4-analyze.md](Story250_lookups-increment4-analyze.md) — Not Started (resolved lookup references must be recorded in aggregator CIL nodes)
- [Story243_lookups-increment1-synthesize.md](Story243_lookups-increment1-synthesize.md) — Not Started (lookup loading infrastructure `loadLookupData` and `controller.RegisterMetricComputation` for lookups must be available)

## Dependent Stories

- [Story252_lookups-increment5-analyze.md](Story252_lookups-increment5-analyze.md) — Not Started
- [Story253_lookups-increment5-synthesize.md](Story253_lookups-increment5-synthesize.md) — Not Started

## Tasks

### Task Sizing

All tasks below are not annotated `[Small]` except Task 2 and Task 3:

- **Task 1** (not small) modifies the aggregator CIL tree walker to emit lookup access code. While it touches at most 2 files (source + test), it requires understanding the walker architecture, constructing coordinate arguments from the aggregator's dimensional context, and mapping between lookup key dimensions and the aggregator's dimensional space — more than a mechanical switch-case addition.
- **Task 2** (small) modifies the preamble generator to ensure lookup loading precedes aggregator execution. Single concern (ordering), at most 2 files, additive change to an existing scheduling function. The task follows the established pattern of execution ordering in the main function generator.
- **Task 3** (small) extends lookup deregistration logic to account for aggregator consumers. Narrow scope, 1–2 files, additive change to existing deregistration guard.
- **Task 4** (not small) creates an end-to-end `with_data` example directory. While it follows the existing example template pattern, it requires creating multiple files (CPL, IO spec, parquet data, expected output, test integration) and verifying the full pipeline end-to-end.
- **Task 5** (not applicable) runs existing tests to verify no regressions. No source file modifications.

### Parallel Execution

None — all tasks have sequential dependencies. Each task builds on the output of the previous task.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4 → Task 5
```

### Reintegration

None — no parallel groups.

---

### Task 1: Emit lookup access code in aggregator synthesis — Not Started

When the aggregator's CIL tree walker encounters a lookup reference node (as recorded in the aggregator CIL node by Story250's analysis), emit Go code that calls `controller.GetMetricComputation(lookupName).GetCell(dimCoords)` to retrieve the lookup value at runtime. The coordinate arguments must be constructed from the aggregator's dimensional context, mapping the aggregator iteration coordinates to the lookup's key dimensions. This is analogous to how `WalkCILLookupRefNode` works in metric synthesis (Story249) but operates within the aggregator's dimensional context rather than a metric's.

The synthesis logic: when the aggregator CIL tree walker encounters a `CILAggregateNode` (or its equivalent from Epic007) that contains lookup reference metadata, it generates:

```go
lookupVal := controller.GetMetricComputation("lookupName").GetCell(dimCoords)
```

where `dimCoords` is a `[]int` mapping aggregator dimension positions to the lookup's key dimensions. The coordinate construction must correctly map from the aggregator's dimension order to the lookup's key dimension order, accounting for any implicit dimensional transformations recorded by Story250.

   a. In the aggregator CIL tree walker (exact file path depends on Epic007 Increment 2 implementation; expected path: `src/main/go/synthesis/aggregator_synthesis.go`), add a walker case for lookup reference CIL nodes encountered during aggregator tree traversal — Not Started
   b. Construct `[]int` coordinate arguments from the aggregator's dimensional context by mapping aggregator dimension indices to lookup key dimension indices — Not Started
   c. Emit `controller.GetMetricComputation(lookupName).GetCell(dimCoords)` as a Go expression in the generated aggregator compute function — Not Started
   d. Write unit tests in the aggregator synthesis test file (expected path: `src/main/go/synthesis/aggregator_synthesis_test.go`): verify that the emitted Go code includes correct `GetMetricComputation` and `GetCell` calls with properly constructed coordinate arguments — Not Started
   e. Compile and run tests: `go test ./src/main/go/synthesis/... -run AggregatorLookup -count=1 -timeout 60s` — Not Started

**Files**: aggregator synthesis file (modify, per Epic007), aggregator synthesis test file (modify, per Epic007)

**Acceptance criteria**:
- Aggregator CIL tree walker emits `controller.GetMetricComputation(lookupName).GetCell(...)` for lookup references
- Coordinate arguments correctly map aggregator dimensions to lookup key dimensions
- Generated Go code compiles and produces correct runtime results
- Unit tests verify the emitted code structure and coordinate construction

### Task 2: Ensure lookups are loaded before aggregators — Not Started [Small]

In the preamble generator (the code that produces the `Compute()` and `Setup()` functions in the synthesized main program), ensure that `loadLookupData` is called before any aggregator compute function that references a lookup. The execution ordering must be: lookup data loading → metrics that do not depend on aggregators → aggregators that reference lookups → metrics that depend on aggregator output.

When a lookup is referenced by both a metric and an aggregator, it must be loaded only once and used by both. The preamble generator must detect this shared-reference pattern and emit a single `loadLookupData` call positioned before all consumers.

   a. In the preamble generator (expected location: within `src/main/go/synthesis/synthesize.go` in the `GenerateMainFunction` or preamble helper), inspect aggregator CIL nodes for lookup references and ensure `loadLookupData` calls precede the aggregator's compute invocation in the generated `main()` function — Not Started
   b. Handle shared lookup references: when a lookup is referenced by both a metric and an aggregator, emit a single `loadLookupData` call before both consumers — Not Started
   c. Write unit test in the synthesis test file: verify the generated `main()` function code places `loadLookupData` calls before aggregator compute calls, and that shared lookups are loaded once — Not Started
   d. Compile and run tests: `go test ./src/main/go/synthesis/... -run PreambleLookupOrdering -count=1 -timeout 60s` — Not Started

**Files**: `src/main/go/synthesis/synthesize.go` (modify), `src/main/go/synthesis/synthesize_test.go` (modify)

**Acceptance criteria**:
- Generated code loads lookup data before any aggregator that references the lookup
- Shared lookups (metric + aggregator consumer) are loaded exactly once
- Execution order in generated `main()` is: lookup load → independent metrics → aggregators → dependent metrics
- Unit test verifies the ordering in generated code

### Task 3: Verify lookup deregistration safety — Not Started [Small]

Lookups referenced in aggregator expressions must not be deregistered (garbage collected via `controller.DeregisterMetricComputation`) until all metrics that consume the aggregator's output have completed. The deregistration logic from Story243 must be extended to track aggregator dependencies: a lookup is deregistered only after the last metric that consumes it — either directly or indirectly through an aggregator — has finished.

The dependency chain is: lookup → aggregator (consumes lookup) → metric (consumes aggregator output). The lookup must survive until the final metric completes.

   a. In the deregistration/scheduling logic (within `src/main/go/synthesis/synthesize.go` or a dedicated cleanup generator), extend the dependency tracking to include aggregator consumers: a lookup marked for deregistration must wait for all downstream aggregator consumers to complete — Not Started
   b. Write unit test: verify the generated cleanup code does not deregister a lookup used by an aggregator until after the aggregator's consuming metrics have finished — Not Started
   c. Compile and run tests: `go test ./src/main/go/synthesis/... -run LookupDeregistrationAggregator -count=1 -timeout 60s` — Not Started

**Files**: `src/main/go/synthesis/synthesize.go` (modify), `src/main/go/synthesis/synthesize_test.go` (modify)

**Acceptance criteria**:
- Lookups referenced in aggregators are not deregistered until all consuming metrics finish
- Dependency chain lookup → aggregator → metric is respected in the generated cleanup code
- Unit test verifies the generated cleanup/deregistration ordering

### Task 4: Create with_data/lookups_in_aggregators example — Not Started

Create an end-to-end `with_data` example demonstrating an aggregator that applies a weighting table from a lookup during aggregation. The example simulates the use case: a `product_weight` lookup keyed by `[Product]` with celltype `WeightFactor`; an aggregator weights product-level metrics by category using the lookup values.

The example directory follows the existing `with_data` pattern (see `examples/with_data/metrics_arithmetic/` for the template) and includes CPL definitions, parquet input data for the lookup and connectors, an IO specification, and expected output.

   a. Create example directory `examples/with_data/lookups_in_aggregators/` — Not Started
   b. Create `core.cpl` with: lookup definition (`product_weight` keyed by `[Product]`), connector definitions for product metrics and lookup data, concept definitions for Product, Category, and WeightFactor, and an aggregator that uses the lookup in its expression — Not Started
   c. Create parquet data files for the lookup data, connector inputs, and any supporting data — Not Started
   d. Create `iospec.json` defining inputs (connectors + lookup data) and outputs (the aggregated metric) — Not Started
   e. Create expected output file (`.expected.xlsx` or `.expected.json`) with the correct aggregation results — Not Started
   f. Integrate the example into the existing `with_data` test driver so that `go test ./examples/with_data/... -run LookupsInAggregators -count=1` verifies end-to-end: parse → analyze → synthesize → run → output matches expected — Not Started

**Files**: `examples/with_data/lookups_in_aggregators/core.cpl` (create), `examples/with_data/lookups_in_aggregators/iospec.json` (create), parquet data files (create), expected output file (create), test integration (modify existing test file or create new)

**Acceptance criteria**:
- Example CPL parses and analyzes without errors
- Synthesized code runs and produces output matching the expected file
- End-to-end test passes in CI
- Example serves as documentation for the aggregator-lookup integration pattern

### Task 5: Verification and regression — Not Started

Run the full CPL test suite to confirm that the changes introduced in Tasks 1–4 do not break any existing functionality. Pay special attention to existing lookup examples (Increments 1–3) and existing aggregator examples (Epic007 Increments 1–2).

   a. Verify existing lookup `with_data` examples (Increments 1–3) still pass: `go test ./examples/with_data/... -count=1 -timeout 120s` — Not Started
   b. Verify existing aggregator `with_data` examples (Epic007) still pass: `go test ./examples/with_data/... -count=1 -timeout 120s` — Not Started
   c. Run the complete synthesis test suite: `go test ./src/main/go/synthesis/... -count=1 -timeout 120s` — Not Started
   d. Run the full CPL test suite (all packages): `go test ./src/main/go/... -count=1 -timeout 180s` — Not Started

**Files**: No source file modifications — verification only.

**Acceptance criteria**:
- All existing lookup examples produce correct results
- All existing aggregator examples produce correct results
- All synthesis tests pass with zero failures
- Full CPL test suite passes with zero regressions

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- No new synthesis infrastructure is added at the lookup level — lookup loading and registration already exist from Story243
- The aggregator CIL tree walker and synthesis infrastructure must exist from Epic007 Increment 2 before this story can execute
- Lookup reference metadata in aggregator CIL nodes must be populated by Story250 before this story can execute
- The emitted lookup access code uses `controller.GetMetricComputation(lookupName).GetCell(dimCoords)` — the same API used by metric-level lookup access in Story249
- Lookup data must not be deregistered until all metrics (direct and indirect consumers through aggregators) have completed
- The `with_data` example must follow existing directory structure and test conventions

## Intent

### Task 1 Intent

Enable the aggregator CIL tree walker to emit lookup access code. When an aggregator references a lookup (as identified during analysis by Story250), the synthesizer must generate Go code that retrieves the lookup value at runtime. The coordinate construction must correctly map the aggregator's dimensional iteration space to the lookup's key dimensions so that the correct cell is accessed.

### Task 2 Intent

Ensure correct execution ordering in the synthesized program. Lookup data must be loaded before any aggregator that references it, and the ordering must respect the dependency chain: lookup loading → aggregator execution → metric consumption of aggregator output. This prevents runtime errors from accessing unloaded lookup data.

### Task 3 Intent

Prevent premature cleanup of lookup data. Lookups that feed into aggregators have an indirect dependency chain — the aggregator consumes the lookup, and downstream metrics consume the aggregator. The lookup must not be freed until the final downstream metric has completed. This extends the existing deregistration safety mechanism from Story243.

### Task 4 Intent

Provide an end-to-end example that demonstrates the aggregator-lookup integration and serves as both documentation and an automated regression test. The example simulates a realistic use case (weighted aggregation using a lookup table) and verifies the complete pipeline.

### Task 5 Intent

Confirm that the changes introduced in this story do not break existing functionality. Since lookup and aggregator synthesis share infrastructure (the preamble generator, the main function generator, deregistration logic), it is essential to verify that existing features continue to work correctly.

## Acceptance Criteria

1. Aggregator CIL tree walker emits `controller.GetMetricComputation(lookupName).GetCell(dimCoords)` for lookup references
2. Coordinate arguments correctly map aggregator dimensions to lookup key dimensions
3. Generated code orders lookup loading before aggregator execution
4. Shared lookups (metric + aggregator consumers) are loaded exactly once
5. Lookups referenced in aggregators are not deregistered until all consuming metrics complete
6. The dependency chain lookup → aggregator → metric is respected in generated code
7. `with_data/lookups_in_aggregators` example passes end-to-end (parse → analyze → synthesize → run → output matches expected)
8. All existing tests pass with zero regressions
