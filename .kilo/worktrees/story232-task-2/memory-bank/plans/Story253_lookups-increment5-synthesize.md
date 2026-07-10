# Plan: Story253 — Lookups Increment 5 (Synthesize)

## Purpose

This plan provides the technical writer with sufficient detail to create Story253: **lookups-increment5-synthesize** — implement synthesis for lookup access within spreader computations (Increment 5 of Epic006).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `lookups-increment5-synthesize`                                                               |
| description  | Generate lookup access code within spreader computation, implicit aggregator execution for lookups, spreader+lookup examples |

## Source EROs

- [ERO006: Lookups Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md) — Increment 5 (documented for completeness; implementation details in ERO008)
- [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md) — Increment 5 sketch

## Prerequisites

- **Epic008 (Spreaders) must be at Increment 3 or above** — spreader CIL tree walker and code generation infrastructure must exist
- Story252 (lookups-increment5-analyze) — needs resolved lookup references and implicit transformation nodes in spreader CIL nodes
- Story243 (lookups-increment1-synthesize) — needs lookup loading infrastructure

## Dependencies

- Epic008 Increment 3 — needs spreader CIL tree walker and synthesis infrastructure
- Story252 (analyze) — needs lookup references and implicit transformations in spreader CIL nodes
- Story243 (synthesize) — needs lookup loading infrastructure

## Context — No New Synthesis Infrastructure

Increment 5 requires **no new synthesis infrastructure** at the lookup level. When a spreader references a lookup, the spreader's CIL tree walker emits calls to `controller.GetMetricComputation(lookupName)` as part of the spreading code. The lookup data is already loaded and registered.

The synthesis work involves:
1. The spreader CIL tree walker emitting lookup access code inline
2. Emitting implicit aggregator code for lookups that need dimensional alignment
3. Ensuring lookup data is loaded before any spreader that references it
4. Ensuring lookups are not deregistered until all consuming metrics complete

## Files to Create or Modify

| File                                                     | Action | Purpose                                                                                          |
| -------------------------------------------------------- | ------ | ------------------------------------------------------------------------------------------------ |
| `src/main/go/synthesis/spreader_synthesis.go`            | Modify | Emit `controller.GetMetricComputation(lookupName)` in spreader code generation; emit implicit aggregator code for lookups |
| `src/main/go/synthesis/preamble_generator.go`            | Modify | Ensure lookup loading precedes spreader execution; schedule lookup→aggregator→spreader ordering   |
| `src/main/go/synthesis/spreader_synthesis_test.go`       | Modify | Add tests for lookup access and implicit aggregator code in spreader synthesis                   |
| `.../examples/with_data/lookups_in_spreaders/`           | Create | Example: spreader using a lookup (e.g., distributing a global target using a lookup-based weighting table) |

**Note:** File paths are indicative — the technical writer should align with the actual spreader package organization from ERO008.

## Synthesis Logic

### Spreader CIL Tree Walker — Lookup Access

When the spreader's CIL tree walker encounters a lookup reference node, it emits:
```go
lookupVal := controller.GetMetricComputation("lookupName").GetCell(dimCoords)
```

### Implicit Aggregator Execution for Lookups

If the analysis phase (Story252) recorded implicit transformation nodes for the lookup (to align its dimensions with the spreader context), the spreader synthesizer must:
1. Generate aggregation code that transforms the lookup's data to the required dimensional form
2. Register the transformed lookup result as a temporary `MetricComputation` or use it inline
3. The transformed data is then used in the spreader expression

### Execution Ordering — Preamble

The preamble generator must ensure:
1. `loadLookupData(controller)` runs first
2. Any implicit aggregators on lookups run
3. Spreaders that reference lookups run
4. Metrics that depend on spreaders run

### with_data Example: lookups_in_spreaders

- CPL: spreader that distributes a global target across countries using a lookup-based weighting table
- Example: `country_weight` lookup keyed by `[Country]` with celltype `WeightFactor`; spreader distributes `global_target` across countries proportionally to lookup weights
- Parquet data: lookup data + connector data + metric data
- `.expected.json`: expected output after spreading with lookup-based weighting

## Task Decomposition (for technical writer)

### Task 1: Emit lookup access code in spreader synthesis — Not Started
   a. In the spreader CIL tree walker, when a lookup reference is encountered, emit `controller.GetMetricComputation(lookupName).GetCell(...)` — Not Started
   b. Construct coordinate arguments from the spreader's dimensional context — Not Started
   c. Write unit tests: verify emitted Go code includes correct lookup access — Not Started
   d. Verify synthesis tests pass — Not Started

### Task 2: Generate implicit aggregator code for lookups — Not Started
   a. When the spreader CIL node has implicit transformation nodes for a lookup, emit aggregation code — Not Started
   b. Emit code that transforms the lookup's dimensions to match the spreader context — Not Started
   c. Handle the transformed lookup data in the spreader expression — Not Started
   d. Write unit tests: lookup needs aggregation before spreading, lookup needs no transformation, lookup needs spreading before spreading (nested) — Not Started
   e. Verify tests pass — Not Started

### Task 3: Ensure correct execution ordering — Not Started [Small]
   a. Verify `loadLookupData` is called before any spreader that references a lookup — Not Started
   b. Verify implicit aggregators on lookups run before the spreader — Not Started
   c. Ensure lookups are not deregistered until all consuming spreader-dependent metrics complete — Not Started
   d. Write integration test: verify generated preamble ordering — Not Started
   e. Verify test passes — Not Started

### Task 4: Create with_data/lookups_in_spreaders example — Not Started
   a. Create CPL file with spreader referencing a lookup in its expression — Not Started
   b. Create IO spec and parquet data files — Not Started
   c. Create `.expected.json` with expected output — Not Started
   d. Verify end-to-end: parse → analyze → synthesize → run → output matches expected — Not Started

### Task 5: Verification and regression — Not Started
   a. Verify existing examples (Increments 1-4) still work — Not Started
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

1. Spreader CIL tree walker emits `controller.GetMetricComputation(lookupName)` for lookup references
2. Implicit aggregator code is generated for lookups that need dimensional alignment
3. Lookup data is loaded before any spreader that references it
4. Implicit aggregators on lookups run before the spreader
5. Lookups referenced in spreaders are not deregistered until all consuming metrics complete
6. `with_data/lookups_in_spreaders` example passes end-to-end
7. All existing tests pass with zero regressions

## Intent Summary for Technical Writer

This story enables lookup data access within spreader computations. Since no new synthesis infrastructure is needed at the lookup level, the work focuses on the spreader's CIL tree walker emitting `controller.GetMetricComputation(lookupName)` calls, generating implicit aggregator code when lookups need dimensional alignment, and the preamble generator ensuring correct execution ordering. Lookup access in spreaders reuses the same `GetMetricComputation` pattern established in Stories 243 and 249. This story is blocked until Epic008's spreader synthesis infrastructure exists at Increment 3 or above.
