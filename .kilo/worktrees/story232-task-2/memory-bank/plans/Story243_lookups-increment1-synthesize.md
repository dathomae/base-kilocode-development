# Plan: Story243 — Lookups Increment 1 (Synthesize)

## Purpose

This plan provides the technical writer with sufficient detail to create Story243: **lookups-increment1-synthesize** — implement the synthesis phase for basic lookup declarations without grain (Increment 1 of Epic006).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `lookups-increment1-synthesize`                                                               |
| description  | Implement CIL tree walker for CILLookupNode, Go code generation for MetricComputation registration, parquet loading, timeless lookup examples |

## Source EROs

- [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md) — Increment 1 sketch
- [ERO006: Lookups Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md) — Increment 1 synthesis walker and code generation

## Dependencies

- Story242 (lookups-increment1-analyze) — needs `CILLookupNode` from analysis
- Epic005 (metrics) Increment 1 — CIL tree walking infrastructure must exist

## Files to Create or Modify

| File                                                     | Action | Purpose                                                                           |
| -------------------------------------------------------- | ------ | --------------------------------------------------------------------------------- |
| `src/main/go/synthesis/cil_tree_walker.go`               | Modify | Add `WalkCILLookupNode` method to CIL tree walker                                |
| `src/main/go/synthesis/lookup_synthesis.go`              | Create | `generateLookupLoading` function: emits Go code for parquet reading + registration |
| `src/main/go/synthesis/preamble_generator.go`            | Modify | Wire lookup loading into the generated preamble before metric calculations       |
| `src/main/go/synthesis/main_generator.go`                | Modify | Include lookup loading in the generated `main()` function                        |
| `src/main/go/synthesis/lookup_synthesis_test.go`         | Create | Unit tests for lookup code generation and CIL walker                             |
| `.../examples/with_data/lookups_basic/`                  | Create | Example directory: CPL, IO spec, parquet data, `.expected.json`                  |
| `src/main/go/synthesis/lookup_io_spec.go`                | Create | Lookup IO Specification parsing — maps lookup names to source file paths         |

### WalkCILLookupNode Method

- Register lookup as a data source in `WalkScope`
- Store lookup definition metadata for preamble generator
- No CIL subtree to walk — `CILLookupNode` is a leaf

### Preamble Code Generation — Lookup Loading

The `generateLookupLoading` function emits:

```go
func loadLookupData(controller *ChronoconeComputationController) error {
    // For each lookup:
    //   1. Read parquet file from IO spec path
    //   2. Build tensor with dimension names
    //   3. Register as MetricComputation via controller.AddMetricComputation(...)
}
```

Key details:
- Reuses existing `readParquetFile` from connector synthesis
- Tensor dimensions are the lookup's key concept names (grain not yet applicable)
- `MetricComputation` is registered on the same `ChronoconeComputationController` as connectors
- Cell type is `float64` (type selection is deferred per the ERO)

### Lookup IO Specification

Defines the mapping between lookup names and their data sources. Schema (separate from connector IO spec):

```json
{
  "lookups": [
    {
      "name": "product_category",
      "source": {
        "type": "parquet",
        "path": "product_category.parquet"
      }
    }
  ]
}
```

### Generated Code Structure

```
main()
├── Parse CPL source
├── Parse IO specifications (Lookup IO Spec)
├── Create ChronoconeComputationController
├── Load connector data → register MetricComputations
├── Load lookup data → register MetricComputations    <-- NEW
├── Schedule metric computation order
├── Compute metrics
└── Output results
```

### with_data Example: lookups_basic

CPL + IO spec + parquet data + `.expected.json` for a timeless lookup:
- Example: product_category mapping with dimensions `[Product]` and celltype `CategoryLabel`
- Parquet file with Product values and corresponding CategoryLabel values
- Expected output verifies the lookup loads correctly

## Task Decomposition (for technical writer)

### Task 1: Implement Lookup IO Specification parsing — Not Started
   a. Define Lookup IO Specification JSON schema — Not Started
   b. Create `src/main/go/synthesis/lookup_io_spec.go` with parsing logic — Not Started
   c. Write unit tests for IO spec parsing (valid, malformed, missing file) — Not Started
   d. Verify IO spec tests pass — Not Started

### Task 2: Implement WalkCILLookupNode in the CIL tree walker — Not Started [Small]
   a. Add `WalkCILLookupNode` method to the existing CIL tree walker — Not Started
   b. The method registers the lookup as a data source and stores its metadata — Not Started
   c. Write unit test: walk a `CILLookupNode` → verify walker scope populated — Not Started
   d. Verify walker tests pass — Not Started

### Task 3: Implement lookup code generation (generateLookupLoading) — Not Started
   a. Create `src/main/go/synthesis/lookup_synthesis.go` with `generateLookupLoading` — Not Started
   b. Emit Go code that reads parquet file, builds tensor, registers `MetricComputation` — Not Started
   c. Emit dimension name registration on the tensor — Not Started
   d. Handle empty lookups (no lookups → emit no loading code) — Not Started
   e. Write unit tests: verify emitted Go code compiles and correctly registers lookups — Not Started
   f. Verify synthesis tests pass — Not Started

### Task 4: Wire lookup loading into the preamble and main generator — Not Started
   a. Modify `preamble_generator.go` to call `generateLookupLoading` — Not Started
   b. Modify `main_generator.go` to include `loadLookupData` call before metric calculations — Not Started
   c. Write integration test: full pipeline parse→analyze→synthesize → verify generated Go compiles — Not Started
   d. Verify integration test passes — Not Started

### Task 5: Create with_data/lookups_basic example — Not Started
   a. Create CPL file with timeless lookup (e.g., product_category with dimensions [Product], celltype CategoryLabel) — Not Started
   b. Create Lookup IO Spec JSON mapping lookup name to parquet file — Not Started
   c. Create parquet data file with test values — Not Started
   d. Create `.expected.json` with expected output — Not Started
   e. Verify end-to-end: parse → analyze → synthesize → run → output matches expected — Not Started

### Task 6: Create parse_and_analyze_only synthesis-level examples — Not Started
   a. Verify parse_and_analyze_only CPL files from Stories 241/242 synthesize correctly — Not Started
   b. Run existing CPL test suite to confirm no regressions — Not Started

## Execution Order

```
Task 1 → Task 3 → Task 4 → Task 5 → Task 6
                ↗
Task 2 ─────────┘
```

- Task 1 (IO spec) and Task 2 (CIL walker) can proceed in parallel
- Task 3 (code generation) depends on Task 2 (needs walker metadata)
- Task 4 (preamble wiring) depends on Task 3
- Task 5 (examples) depends on Tasks 3 and 4
- Task 6 (regression) depends on all preceding tasks

## Parallel Execution

- **Group A** (Tasks 1, 2): Disjoint files — `lookup_io_spec.go` vs `cil_tree_walker.go`. No shared compilation dependency.

## Reintegration

**Group A (Tasks 1, 2):**
- Merge order: any order (disjoint files)
- Integration tests: `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` — verify IO spec parsing and CIL walker compile together
- Watch for conflicts in: none expected

## Acceptance Criteria Summary

1. Lookup IO Specification parsing produces a mapping of lookup names to data source paths
2. `WalkCILLookupNode` registers lookup metadata in the walker scope
3. `generateLookupLoading` emits compilable Go code that reads parquet and registers `MetricComputation`
4. Generated `main()` calls `loadLookupData` before any metric calculations
5. Tensor indexing is by concept dimension names (no grain handling yet)
6. `with_data/lookups_basic` example passes end-to-end
7. Existing test suite passes with zero regressions

## Intent Summary for Technical Writer

This story delivers the synthesis phase for basic lookup declarations. The CIL tree walker handles `CILLookupNode` leaf nodes, emitting Go code that reads lookup data from parquet files (specified in the Lookup IO Specification) and registers each lookup as a `MetricComputation` on the `ChronoconeComputationController`. The synthesis reuses the existing parquet reading and controller infrastructure from connector synthesis. In Increment 1, lookups are loaded but not yet accessed by metric calculations — that integration is added in Increment 3. The generated preamble ensures lookup loading occurs before any metric computation runs.
