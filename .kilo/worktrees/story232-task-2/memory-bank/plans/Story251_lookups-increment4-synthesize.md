# Plan: Story251 — Lookups Increment 4 (Synthesize)

## Purpose

This plan provides the technical writer with sufficient detail to create Story251: **lookups-increment4-synthesize** — implement synthesis for lookup access within aggregator computations (Increment 4 of Epic006).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `lookups-increment4-synthesize`                                                               |
| description  | Generate lookup access code within aggregator computation, ensure lookup loading before aggregator execution, aggregator+lookup examples |

## Source EROs

- [ERO006: Lookups Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md) — Increment 4 (documented for completeness; implementation details in ERO007)
- [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md) — Increment 4 sketch

## Prerequisites

- **Epic007 (Aggregators) must be at Increment 2 or above** — aggregator CIL tree walker and code generation infrastructure must exist
- Story251 (lookups-increment4-analyze) — needs resolved lookup references in aggrecator CIL nodes
- Story243 (lookups-increment1-synthesize) — needs lookup loading infrastructure

## Dependencies

- Epic007 Increment 2 — needs aggregator CIL tree walker and synthesis infrastructure
- Story251 (analyze) — needs lookup references in aggregator CIL nodes
- Story243 (synthesize) — needs lookup loading infrastructure

## Context — No New Synthesis Infrastructure

Increment 4 requires **no new synthesis infrastructure** at the lookup level. When an aggregator references a lookup, the aggregator's CIL tree walker emits calls to `controller.GetMetricComputation(lookupName)` as part of the aggregation code. The lookup data is already loaded and registered before aggregators run.

The synthesis work involves:
1. Ensuring lookup data is loaded before any aggregator that references it
2. The aggregator CIL tree walker emitting lookup access code inline
3. Unit compatibility verification at synthesis time

## Files to Create or Modify

| File                                                     | Action | Purpose                                                                                          |
| -------------------------------------------------------- | ------ | ------------------------------------------------------------------------------------------------ |
| `src/main/go/synthesis/aggregator_synthesis.go`          | Modify | Emit `controller.GetMetricComputation(lookupName)` in aggregator code generation                 |
| `src/main/go/synthesis/preamble_generator.go`            | Modify | Ensure lookup loading precedes aggregator execution (schedule lookup→aggregator ordering)         |
| `src/main/go/synthesis/aggregator_synthesis_test.go`     | Modify | Add tests for lookup access code in aggregator synthesis                                         |
| `.../examples/with_data/lookups_in_aggregators/`         | Create | Example: aggregator using a lookup (e.g., weighting table during aggregation)                    |

**Note:** File paths are indicative — the technical writer should align with the actual aggregator package organization from ERO007.

## Synthesis Logic

### Aggregator CIL Tree Walker — Lookup Access

When the aggregator's CIL tree walker encounters a lookup reference node (as recorded in the aggregator CIL node from analysis), it emits:
```go
lookupVal := controller.GetMetricComputation("lookupName").GetCell(dimCoords)
```

This is similar to how `WalkCILLookupRefNode` works in metric synthesis (Story249), but within the aggregator's dimensional context.

### Execution Ordering — Preamble

The preamble generator must ensure:
1. `loadLookupData(controller)` runs first
2. Metrics that don't depend on aggregators run
3. Aggregators that reference lookups run (with lookup data available)
4. Metrics that depend on aggregators run

Lookups referenced in aggregator expressions must not be deregistered (garbage collected) until all metrics that use the aggregator have completed.

### with_data Example: lookups_in_aggregators

- CPL: aggregator that applies a weighting table from a lookup during aggregation
- Example use case: `product_weight` lookup keyed by `[Product]` with celltype `WeightFactor`; aggregator weights product-level metrics by category
- Parquet data: lookup data + connector data + metric data
- `.expected.json`: expected output after aggregation with lookup-based weighting

## Task Decomposition (for technical writer)

### Task 1: Emit lookup access code in aggregator synthesis — Not Started
   a. In the aggregator CIL tree walker, when a lookup reference is encountered, emit `controller.GetMetricComputation(lookupName).GetCell(...)` — Not Started
   b. Construct coordinate arguments from the aggregator's dimensional context — Not Started
   c. Write unit tests: verify emitted Go code includes correct lookup access — Not Started
   d. Verify synthesis tests pass — Not Started

### Task 2: Ensure lookups are loaded before aggregators — Not Started [Small]
   a. In the preamble generator, verify `loadLookupData` is called before any aggregator that references a lookup — Not Started
   b. Handle the case where a lookup is referenced by both metrics and aggregators (loaded once, used by both) — Not Started
   c. Write unit test: verify generated code ordering (lookup load → metric → aggregator → metric consumer) — Not Started
   d. Verify test passes — Not Started

### Task 3: Verify lookup deregistration safety — Not Started [Small]
   a. Ensure lookups referenced in aggregators are not deregistered until all consuming metrics finish — Not Started
   b. Write unit test: verify scheduling respects lookup→aggregator→metric dependency chain — Not Started
   c. Verify test passes — Not Started

### Task 4: Create with_data/lookups_in_aggregators example — Not Started
   a. Create CPL file with aggregator referencing a lookup in its expression — Not Started
   b. Create IO spec and parquet data files — Not Started
   c. Create `.expected.json` with expected output — Not Started
   d. Verify end-to-end: parse → analyze → synthesize → run → output matches expected — Not Started

### Task 5: Verification and regression — Not Started
   a. Verify existing examples (Increments 1-3) still work — Not Started
   b. Run full CPL test suite to confirm no regressions — Not Started

## Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4 → Task 5
```

All tasks are sequential — each builds on the previous.

## Parallel Execution

None — all tasks have sequential dependencies.

## Reintegration

None — no parallel groups.

## Acceptance Criteria Summary

1. Aggregator CIL tree walker emits `controller.GetMetricComputation(lookupName)` for lookup references
2. Lookup data is loaded before any aggregator that references it
3. Lookups referenced in aggregators are not deregistered until all consuming metrics complete
4. Generated aggregator code correctly accesses lookup data by dimensional coordinates
5. `with_data/lookups_in_aggregators` example passes end-to-end
6. All existing tests pass with zero regressions

## Intent Summary for Technical Writer

This story enables lookup data access within aggregator computations. Since no new synthesis infrastructure is needed at the lookup level, the work focuses on the aggregator's CIL tree walker emitting `controller.GetMetricComputation(lookupName)` calls and the preamble generator ensuring correct execution ordering. Lookup data is already loaded via the infrastructure built in Story243; this story ensures aggregators can access it. The story is blocked until Epic007's aggregator synthesis infrastructure exists at Increment 2 or above.
