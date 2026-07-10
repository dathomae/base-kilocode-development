# Plan: Story246 — Lookups Increment 2 (Synthesize)

## Purpose

This plan provides the technical writer with sufficient detail to create Story246: **lookups-increment2-synthesize** — implement synthesis for temporal lookups with grain (Increment 2 of Epic006).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `lookups-increment2-synthesize`                                                               |
| description  | Generate Go code for temporal lookup loading, tensor indexing with grain dimensions, multi-grain resolution, temporal lookup examples |

## Source EROs

- [ERO006: Lookups Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md) — Increment 2 temporal lookup code generation

## Dependencies

- Story245 (lookups-increment2-analyze) — needs `CILLookupNode.HasGrain` and `CILLookupNode.Grain` fields
- Story243 (lookups-increment1-synthesize) — extends existing lookup loading and registration code

## Files to Create or Modify

| File                                                     | Action | Purpose                                                                               |
| -------------------------------------------------------- | ------ | ------------------------------------------------------------------------------------- |
| `src/main/go/synthesis/lookup_synthesis.go`              | Modify | Extend `generateLookupLoading` to handle grain dimensions and temporal tensor structure |
| `src/main/go/synthesis/cil_tree_walker.go`               | Modify | Update `WalkCILLookupNode` to register temporal lookups separately                    |
| `src/main/go/synthesis/preamble_generator.go`            | Modify | Handle temporal lookups in preamble (grain dimension registration)                    |
| `src/main/go/synthesis/lookup_synthesis_test.go`         | Modify | Add tests for temporal lookup code generation and multi-grain resolution              |
| `.../examples/with_data/lookups_temporal/`               | Create | Example: temporal lookup with currency conversion rates keyed by Week and Month        |

### Synthesis Changes for Grain

#### CIL Tree Walker

When `CILLookupNode.HasGrain` is true:
- Mark the lookup as temporal in the walker scope (`scope.temporalLookups[name] = node`)
- Preamble generator uses this to include grain dimensions in tensor structure

#### Preamble Code Generation — Temporal Lookups

Extended `generateLookupLoading`:
- For each lookup with `HasGrain=true`, include grain dimension names in the tensor's `SetDimensions()` call
- Register grain dimensions on the `MetricComputation` via `computation.SetGrainDimensions(...)`
- Grain dimensions are appended after concept dimensions in the dimension list
- Multi-grain lookups include ALL grain measures as dimensions

Generated code structure:
```go
// For temporal lookup:
tensor.SetDimensions([]string{
    "source_currency", "target_currency",  // concept dimensions
    "Week", "Month",                        // grain dimensions (NEW)
})

computation := NewMetricComputationFromTensor("exchange_rate", tensor)
computation.SetGrainDimensions([]string{"Week", "Month"})  // NEW
controller.AddMetricComputation(computation)
```

#### Multi-Grain Resolution at Runtime

When a lookup has multiple grain measures (e.g., `grain Week, Month`), the runtime framework must select the appropriate row based on the requesting metric cell's time coordinate. The generated code includes a helper `resolveGrainIndex` that:
1. Iterates over the lookup's grain dimensions
2. Checks if the cell time matches each grain's interval
3. Returns the index of the matching grain row

### with_data Example: lookups_temporal

CPL + IO spec + parquet data + `.expected.json` for a temporal lookup:
- Example: `exchange_rate` with dimensions `[Currency, Currency]`, grain `Week, Month`, celltype `ConversionRate`
- Parquet file with currency pairs, Week/Month grain values, and conversion rates
- Expected output verifies temporal lookup loads and indexes correctly by time dimensions

## Task Decomposition (for technical writer)

### Task 1: Update WalkCILLookupNode for temporal lookups — Not Started [Small]
   a. Add temporal lookup detection: if `node.HasGrain`, store in `scope.temporalLookups` — Not Started
   b. Write unit test: walk `CILLookupNode` with grain → verify temporal lookups map populated — Not Started
   c. Verify walker tests pass — Not Started

### Task 2: Extend generateLookupLoading for temporal lookups — Not Started
   a. When `HasGrain` is true, include grain dimensions in tensor `SetDimensions()` call — Not Started
   b. Emit `computation.SetGrainDimensions(...)` for temporal lookups — Not Started
   c. Ensure grain dimensions are appended after concept dimensions — Not Started
   d. Handle mixed timeless/temporal lookups in same compilation unit — Not Started
   e. Write unit tests: verify emitted code for temporal lookups compiles, verify grain dimensions registered — Not Started
   f. Verify synthesis tests pass — Not Started

### Task 3: Implement multi-grain resolution helper generation — Not Started
   a. Emit `resolveGrainIndex` helper function that matches cell time against grain dimensions — Not Started
   b. Generate the helper only when temporal lookups exist in the compilation unit — Not Started
   c. Write unit tests: multi-grain lookup with different time coordinate scenarios — Not Started
   d. Verify tests pass — Not Started

### Task 4: Create with_data/lookups_temporal example — Not Started
   a. Create CPL file with temporal lookup (dimensions, grain, celltype) — Not Started
   b. Create Lookup IO Spec JSON with temporal lookup entry — Not Started
   c. Create parquet data file with concept dimensions + grain columns + celltype column — Not Started
   d. Create `.expected.json` with expected output — Not Started
   e. Verify end-to-end: parse → analyze → synthesize → run → output matches expected — Not Started

### Task 5: Verification and regression — Not Started
   a. Verify timeless lookups (Increment 1) still work with no regressions — Not Started
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

1. CIL tree walker detects temporal lookups via `HasGrain` flag
2. Generated Go code includes grain dimensions in tensor dimension list
3. Generated Go code calls `SetGrainDimensions(...)` on temporal lookup `MetricComputation` objects
4. Multi-grain lookup resolution helper is generated when temporal lookups exist
5. Timeless lookups (no grain) continue to work unchanged
6. `with_data/lookups_temporal` example passes end-to-end
7. Full CPL test suite passes with zero regressions

## Intent Summary for Technical Writer

This story extends the synthesis phase to support temporal lookups. When `CILLookupNode.HasGrain` is true, the generated Go code includes grain dimensions in the tensor's dimensional structure and registers grain dimension metadata on the `MetricComputation`. Multi-grain lookups (e.g., `grain Week, Month`) generate a runtime resolution helper that matches the requesting cell's time coordinate against grain dimensions to select the appropriate row. The synthesis reuses the same parquet reading and controller infrastructure from Increment 1, with grain dimensions appended after concept dimensions in the tensor.
