# Story263: Aggregators Increment 5 — Synthesize (Epic007)

## Goal

Implement fused matrix operations for multi-step aggregator sequences, scalar reduction edge case handling, and end-to-end sequence optimization examples. This is Increment 5 of Epic007 (Aggregators), delivering the final optimization layer for aggregator synthesis. When multi-step aggregator sequences are selected (e.g., collapse Region then Country to reach Product), adjacent aggregation matrices are fused into single matrix multiplications where possible, reducing the number of tensor operations and intermediate tensors. The `EmitFusedAggregatorSequence` walker method extracts matrices from typed aggregate nodes, fuses compatible pairs, and emits optimized `ReduceDimension` calls. Scalar reduction (`to []`) producing rank-zero tensors is handled as a special edge case. An end-to-end `with_data` example exercises multi-step sequence selection and verifies correct numerical results.

## References

- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md)
- [ERO007: Aggregators Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md)
- [Epic007: CPL Aggregators Implementation](../epics/Epic007_cpl-aggregators-implementation.md)
- [Source root: src/chronocone_planning_language](../../src/chronocone_planning_language/)
- [Plan: Story263 aggregators-increment5-synthesize](../plans/Story263_aggregators-increment5-synthesize.md)

## Dependencies

- [Story254_aggregators-increment1-matrix-mult.md](Story254_aggregators-increment1-matrix-mult.md) — `tensorops.ReduceDimension`, `tensorops.AggregationMatrix`, `composeIndex`/`decomposeIndex` helpers must exist — **Not Started**
- [Story259_aggregators-increment3-synthesize.md](Story259_aggregators-increment3-synthesize.md) — walk methods for all typed CIL aggregate nodes (`CILAggregateSumNode`, `CILAggregateCountNode`, `CILAggregateSizeNode`, `CILAggregateInverseSumNode`, `CILAggregateInverseCountNode`) must exist in the CIL tree walker — **Not Yet Authored** (plan file pending)
- [Story261_aggregators-increment4-synthesize.md](Story261_aggregators-increment4-synthesize.md) — metric code generation with aggregate nodes (emitting aggregation matrices in preamble, emitting ReduceDimension calls for single-step aggregators during metric synthesis) must be functional — **Not Yet Authored** (plan file pending)
- [Story262_aggregators-increment5-analyze.md](Story262_aggregators-increment5-analyze.md) — `enumerateAggregatorSequences` and `selectOptimalSequence` must produce multi-step sequences with dimensions chains and intermediate aggregate nodes — **Not Yet Authored** (plan file pending)

## Dependent Stories

None — this is the final synthesis increment for the Aggregators epic. Future work includes spreader synthesis (ERO008) and incremental refinements.

## Tasks

### Task Sizing

- **Task 1** is not `[Small]` — it introduces new abstractions (`fuseAggregationMatrices`, `fuseMatrixChain`) with matrix multiplication logic, dimension validation, and a greedy fusion algorithm. Matrix multiplication requires reasoning about 2D data structures, index alignment, and compatibility checks. Touches `tensorops.go` + `tensorops_test.go` (2 files, but complex math).
- **Task 2a** is not `[Small]` — it introduces a new walker method (`EmitFusedAggregatorSequence`) that must dispatch on multiple typed CIL aggregate node types, extract matrices, call the tensorops fusion functions, emit ReduceDimension code with correct dimension indices, manage scope-based variable naming, and register fused matrix variables in the preamble generator. This requires multi-structure reasoning across the CIL node hierarchy, the walker scope model, and the `CILCodeGenerator` preamble. Touches `cil_transformer.go` + `cil_transformer_aggregator_test.go`.
- **Task 2b** is not `[Small]` — it integrates `EmitFusedAggregatorSequence` into the walker's handling of chained aggregate nodes in metric inputs. This requires understanding how `TransformMetric` dispatches on metric input types, detecting when a metric input is a multi-step aggregate chain (from the analysis phase's aggregate node sequencing), and routing through the fused emission path. Touches `cil_transformer_metric.go`; may involve coordination with the postfix walk path.
- **Task 3** is `[Small]` — it adds edge case tests for scalar reduction across two test files (`tensorops_test.go` + `cil_transformer_aggregator_test.go`). The scope is narrow (rank-zero tensor handling, empty shape edge cases), following existing test patterns in both packages. **Warning**: If source code changes beyond test additions are needed (e.g., fixing `ReduceDimension` or `composeIndex`/`decomposeIndex` for rank-zero tensors), the task is not small — route to `code-for-story-implementor`.
- **Task 4** is not `[Small]` — it creates a complete `with_data` example directory with 5+ coordinated file types (CPL source, JSON IO spec, parquet test data, XLSX expected output, Go test). All files must be internally consistent with the multi-step aggregation scenario.
- **Task 5** is `[Small]` — it adds a single integration test to the existing `synthesize_test.go`, following the established `TestSynthesizeWithAnalysisDefinitions` pattern. Purely additive, one concern, one file.

Tasks 1, 2a, 2b, and 4 route to `code-for-story-implementor`. Tasks 3 and 5 route to `small-code-for-story-implementor`.

### Parallel Execution

- **Group A** (Tasks 2b, 3, 4, 5): Task 2b modifies `cil_transformer_metric.go`; Task 3 modifies `tensorops_test.go` (tensorops package) and `cil_transformer_aggregator_test.go` (cil package); Task 4 creates files under `examples/with_data/aggregators_optimization/`; Task 5 modifies `synthesis/synthesize_test.go` (synthesis package). All file sets are fully disjoint across four distinct directories.

### Execution Order

```
Task 1 → Task 2a → [Task 2b ‖ Task 3 ‖ Task 4 ‖ Task 5]
```

- Task 1 must complete first: matrix fusion is used by Task 2a's `EmitFusedAggregatorSequence`
- Task 2a depends on Task 1: the `EmitFusedAggregatorSequence` method calls `fuseMatrixChain` and `fuseAggregationMatrices`
- Tasks 2b, 3, 4, 5 are parallel after Task 2a: Task 2b integrates the method into metric inputs; Task 3 is a tensorops + cil edge case; Task 4 is examples; Task 5 is synthesis integration tests

### Reintegration

**Group A (Tasks 2b, 3, 4, 5):**
- Merge order: any order (fully disjoint file paths across four different directories/packages)
- Integration tests: `go test ./src/main/go/tensorops/... ./src/main/go/cil/... ./src/main/go/synthesis/... -count=1 -timeout 60s` — verify all tests pass together after merging the story branch with all worktree branches
- Watch for conflicts in: none expected (all tasks touch different files in different directories)

### Task 1: Implement matrix fusion — Not Started

Implement `fuseAggregationMatrices` and `fuseMatrixChain` in the `tensorops` package for combining adjacent aggregation matrices in multi-step sequences.

   a. Implement `fuseAggregationMatrices(a, b *AggregationMatrix) *AggregationMatrix` in `src/main/go/tensorops/tensorops.go` — multiply two compatible matrices A (M×N) × B (N×P) → C (M×P) using triple-nested loops; validate that `a.Columns == b.Rows` (return nil on mismatch); assign row labels from `a.RowLabels` and column labels from `b.ColLabels` to the fused matrix — Not Started
   b. Implement `fuseMatrixChain(matrices []*AggregationMatrix) []*AggregationMatrix` in `src/main/go/tensorops/tensorops.go` — greedily fuse adjacent compatible matrices in a chain: iterate through the list, accumulate compatible pairs via `fuseAggregationMatrices`, emit the accumulated result when a non-compatible pair is encountered (or at end of chain) — Not Started
   c. Write tests in `src/main/go/tensorops/tensorops_test.go`: 2×3 × 3×2 → 2×2 with correct numerical values computed by hand; identity matrix fusion (no-op — same labels); dimension mismatch returns nil; chain of 3 matrices where first two are compatible but third is not → 2 fused matrices; chain of 3 matrices where all are compatible → 1 fused matrix — Not Started
   d. Verify tests pass: `go test ./src/main/go/tensorops/... -run Fuse -count=1` — Not Started

### Task 2a: Add EmitFusedAggregatorSequence method to CIL walker — Not Started

Add the `EmitFusedAggregatorSequence` method to the CIL tree walker (`CILAbstractTransformer`) that extracts matrices from a sequence of typed aggregate nodes, fuses compatible adjacent matrices using the tensorops fusion functions, and emits optimized `ReduceDimension` calls.

   a. Implement `EmitFusedAggregatorSequence` in `src/main/go/cil/cil_transformer.go` — accept a slice of typed aggregate CIL nodes (`[]CILNode`), extract `AggregationMatrix` from each node via type-switch on `*CILAggregateSumNode`, `*CILAggregateCountNode`, `*CILAggregateInverseSumNode`, `*CILAggregateInverseCountNode` (skip `*CILAggregateSizeNode` — no matrix); call `fuseMatrixChain` on the collected matrices; for each fused matrix, emit a `ReduceDimension` call with a unique `fused_<N>` variable name and error handling (`if err != nil { return nil, err }`); register each fused matrix in the `CILCodeGenerator` preamble via `scope.registerMatrixVar`; push the final result variable onto the operand stack — Not Started
   b. Write tests in `src/main/go/cil/cil_transformer_aggregator_test.go`: 2-step compatible sequence emits a single fused `ReduceDimension` call with a fused matrix variable; 2-step incompatible sequence (dimensions don't align) emits two separate `ReduceDimension` calls; verify fused matrix variables are named `fusedMatrix_<name>_<N>` in the preamble — Not Started
   c. Verify tests pass: `go test ./src/main/go/cil/... -run FusedAggregator -count=1` — Not Started

### Task 2b: Integrate fused emission into metric input handling — Not Started

Wire `EmitFusedAggregatorSequence` into the walker's handling of chained aggregate nodes in metric inputs, so that when the analysis phase produces multi-step aggregate sequences, the synthesis phase generates fused code.

   a. In `src/main/go/cil/cil_transformer_metric.go`, extend the `TransformMetric` function (or the postfix walk path) to detect when a metric input is a chain of aggregate nodes (from `enumerateAggregatorSequences`/`selectOptimalSequence` analysis output): if an `InputRef` of kind "metric" is backed by a multi-step aggregate chain (detected via the CIL node hierarchy), route through `EmitFusedAggregatorSequence` instead of emitting individual aggregate nodes — Not Started
   b. Verify compilation: `go build ./src/main/go/cil/...` — Not Started
   c. Write integration test in `src/main/go/cil/cil_transformer_aggregator_test.go`: a metric DAG with a 2-step aggregator chain → fused emission is triggered; the generated compute function contains fused ReduceDimension calls (not individual aggregate calls); preamble contains fused matrix definitions — Not Started
   d. Verify tests pass: `go test ./src/main/go/cil/... -run FusedIntegr -count=1` — Not Started

### Task 3: Handle scalar reduction edge case — Not Started [Small]

Verify that `tensorops.ReduceDimension` and `composeIndex`/`decomposeIndex` handle rank-zero tensor results correctly (scalar reduction where `to` is `[]`). Add targeted tests for the edge case. Route to `small-code-for-story-implementor` for test additions only. **If source code changes beyond test additions are needed** (e.g., fixing `ReduceDimension` or the index helpers), escalate to `code-for-story-implementor`.

   a. Verify `tensorops.ReduceDimension` handles rank-zero result tensors (shape `[]` after scalar reduction) — examine the result shape construction logic for the case where the source tensor is collapsed to a scalar (check for off-by-one errors when `targetDim` is `[]` or `Columns == 0`) — Not Started
   b. Verify `composeIndex`/`decomposeIndex` handle empty shape (`[]int{}`) — iterate over zero-length shape slices without panicking; `composeIndex([], [])` should return 0; `decomposeIndex(0, [])` should return `[]` — Not Started
   c. Write test in `src/main/go/tensorops/tensorops_test.go`: 1D tensor [10, 20] → scalar reduction (aggregation matrix [[1.0], [1.0]]) → 0D tensor [30]; verify shape is `[]`, data is `[30.0]` — Not Started
   d. Write walker test in `src/main/go/cil/cil_transformer_aggregator_test.go`: a scalar reduction aggregate node emits correct `ReduceDimension` call with `collapsedDim=0, targetDim=0` and the result variable has shape `[]` — Not Started
   e. Verify tests pass: `go test ./src/main/go/tensorops/... ./src/main/go/cil/... -run Scalar -count=1` — Not Started

### Task 4: Create with_data sequence optimization example — Not Started

Create a complete `with_data` example exercising multi-step aggregator sequence selection and optimization. The example demonstrates that the framework enumerates both valid ordering paths through a dimension chain, selects the optimal one (least information loss), and produces correct aggregated results.

   a. Create directory `examples/with_data/aggregators_optimization/` — Not Started
   b. Write CPL file (`aggregators.cpl`): define concepts Country={USA, Canada} (2 values), Region={North, South, East, West} (4 values), Product={widget, gadget} (2 values); define connector with dimensions [Country, Region, Product]; define two aggregators: `from [Country, Region, Product] to [Region, Product] = sum(Country)` and `from [Region, Product] to [Product] = sum(Region)`; define a metric with dimensions [Product] that references the connector — Not Started
   c. Write IO spec (`io_spec.json`): input connector pointing to parquet data, output metric with dimension [Product] — Not Started
   d. Generate parquet test data with known values for all 2×4×2 = 16 combinations of Country, Region, Product — Not Started
   e. Compute ground-truth `.expected.xlsx`: aggregate the parquet data by hand (sum over Country then Region → 2 product values) — Not Started
   f. Verify framework selects the optimal aggregation order: Country-first (2→4) has less information loss than Region-first (4→2); confirm the generated code uses the `enumerateAggregatorSequences` + `selectOptimalSequence` path — Not Started
   g. Verify example test passes as part of the full test suite: `go test ./examples/with_data/... -run aggregators_optimization -count=1` — Not Started

### Task 5: End-to-end multi-step synthesis test — Not Started [Small]

Add an integration test to the synthesis test suite that validates the full pipeline: analysis produces multi-step aggregate sequences, synthesis generates fused (or sequential) ReduceDimension code, and the generated program produces correct numerical results.

   a. Write test in `src/main/go/synthesis/synthesize_test.go`: construct an `AnalysisDefinitions` with concepts (Country=2 values, Region=4 values) and a metric requiring 2-step aggregation; call `Synthesize` with the analysis definitions and a minimal IO spec; verify the generated code contains `tensorops.ReduceDimension` calls with fused matrix variable names (`fusedMatrix_*`); verify the generated code compiles via `go build` on the output file; verify that executing the compiled program produces correct aggregated numerical results — Not Started
   b. Verify test passes: `go test ./src/main/go/synthesis/... -run MultiStepAggregator -count=1` — Not Started

## Test-First Development

This story follows the Logical TDD Lifecycle. For each coding task, write tests before implementing the code. The test-and-implement cycle is integrated within each task in the correct order:

- **Task 1**: Write matrix fusion tests (subtask c) before the fusion functions pass them (subtask d).
- **Task 2a**: Write CIL walker tests (subtask b) before `EmitFusedAggregatorSequence` passes them (subtask c).
- **Task 2b**: Write integration test (subtask c) before the integrated code passes it (subtask d).
- **Task 3**: Write scalar reduction tests (subtasks c–d) before verifying they pass (subtask e).
- **Task 4**: All example files must be internally consistent — verify the example test passes as acceptance.
- **Task 5**: Write the synthesis integration test (subtask a) before verifying it passes (subtask b).

Tests are NOT broken out as separate tasks. See `memory-bank/concepts.md` for details on the Logical TDD Lifecycle.

## Constraints

- The `tensorops.ReduceDimension` function, `tensorops.AggregationMatrix` type, and `composeIndex`/`decomposeIndex` helpers must exist (built in Story254)
- The full set of typed CIL aggregate nodes must exist with walk methods in the CIL tree walker (built in Story259):
  - `CILAggregateSumNode` — carries aggregation matrix and dimension metadata, walked to emit ReduceDimension calls
  - `CILAggregateCountNode` — same structure as sum, uses count matrix
  - `CILAggregateSizeNode` — emits compile-time constant (no matrix)
  - `CILAggregateInverseSumNode` / `CILAggregateInverseCountNode` — same structure, inverse matrices
- The `CILCodeGenerator` preamble generator must support emitting `AggregationMatrix` as Go struct literals (built in Story261)
- The `CILAbstractTransformer` must support `registerMatrixVar` for registering fused matrices in the preamble (built in Story261)
- `enumerateAggregatorSequences` and `selectOptimalSequence` must produce multi-step sequences of typed aggregate nodes (built in Story262)
- Aggregation matrices are pre-built from concept value sets at compile time and emitted as Go struct literals
- Matrix fusion is mathematically equivalent to sequential application — it must produce identical numerical results
- All generated code must compile under `source_root` (`src/chronocone_planning_language`)
- Generated code follows existing patterns: `check()` helper for error handling, `ChronoconeTensor[any]` for tensor types
- Types are represented as `float64`

## Intent

- **Task 1**: Provide low-level matrix fusion infrastructure so adjacent aggregation matrices can be combined into single multiplications, reducing the number of tensor operations in multi-step sequences.
- **Task 2a**: Add the walker method that orchestrates matrix fusion during code generation — extracts matrices from typed nodes, fuses them, and emits optimized `ReduceDimension` calls with fused matrix variable names.
- **Task 2b**: Wire the fused emission path into the metric computation pipeline so that when the analysis phase detects a multi-step aggregate sequence, the synthesis phase automatically generates fused (optimized) code instead of individual aggregate calls.
- **Task 3**: Close the coverage gap for scalar reduction — ensure that collapsing all dimensions to `[]` produces a correct rank-zero tensor and that the index decomposition helpers don't panic on empty shapes.
- **Task 4**: Provide a concrete end-to-end example that exercises the full sequence optimization pipeline (enumeration, selection, fusion, code generation, execution) and serves as a regression test.
- **Task 5**: Validate that the synthesis pipeline correctly handles multi-step aggregate sequences end-to-end, producing compilable, correct Go code.

## Acceptance Criteria

1. `fuseAggregationMatrices` correctly multiplies two compatible matrices A (M×N) × B (N×P) → C (M×P) with correct numerical values; returns nil for incompatible dimensions
2. `fuseMatrixChain` greedily fuses adjacent compatible matrices in a chain; incompatible matrices remain unfused and are emitted as separate steps
3. `EmitFusedAggregatorSequence` extracts matrices from typed aggregate CIL nodes, fuses compatible pairs, and emits `ReduceDimension` calls for each fused step with unique variable names
4. Fused matrix variables are registered in the `CILCodeGenerator` preamble and named `fusedMatrix_<name>_<N>`
5. The walker detects multi-step aggregate chains in metric inputs and routes them through the fused emission path
6. Generated code for a 2-step compatible sequence contains a single fused `ReduceDimension` call with a fused matrix; a 2-step incompatible sequence contains two separate calls
7. Scalar reduction (`to []`) produces rank-zero tensors (shape `[]`) with correct numerical results
8. `composeIndex`/`decomposeIndex` handle empty shapes without panicking
9. `with_data/aggregators_optimization/` example: multi-step sequence exercised, optimal order selected, correct aggregated results in expected XLSX
10. End-to-end synthesis test produces compilable Go code that executes and returns correct aggregated numerical results
11. All tests pass across `tensorops`, `cil`, and `synthesis` packages
