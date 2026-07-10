# Story254: Aggregators Increment 1 — Tensor Matrix Multiplication Infrastructure

## Goal

Implement standalone tensor matrix multiplication infrastructure for dimension-reducing and dimension-expanding operations. This is Increment 1 of Epic007 (Aggregators), providing the pure Go tensor math foundation that both aggregators (dimension reduction) and spreaders (dimension expansion) will consume in later increments. The tensorops package is independently testable with hand-constructed tensors and has no CPL parsing, analysis, or CIL dependencies.

## References

- [Plan: Story254](../../memory-bank/plans/Story254_aggregators-increment1-matrix-mult.md)
- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md)
- [ERO007: Aggregators Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md) — Matrix multiplication algorithms, `ReduceDimension`, `ExpandDimension`, helper functions
- [Epic007: CPL Aggregators Implementation](../epics/Epic007_cpl-aggregators-implementation.md)

## Dependencies

None — This is a standalone infrastructure story with no CPL parsing, analysis, or CIL dependencies. Pure Go tensor math. This is the foundational story for Epic007.

## Dependent Stories

- Story for Increment 2 (Aggregator Parsing and Validation) — Not Started (consumes tensorops.ReduceDimension)
- Story for Increment 3 (Additional Aggregator Function CIL Node Types) — Not Started (consumes tensorops operations)
- Story for Increment 4 (Aggregator Selection and Integration) — Not Started (consumes tensorops operations)
- Story for Increment 5 (Aggregator Sequence Optimization) — Not Started (consumes tensorops operations)

## Tasks

### Task Sizing

Tasks 1 and 4 are annotated `[Small]` — they touch at most 2 files and are purely additive with narrow scope.

- **Task 1** (`[Small]`): Creates one source file (`index_helpers.go`) and adds tests to the shared test file. Both helpers (`composeIndex`, `decomposeIndex`) follow a single established pattern (row-major index conversion). Straightforward.
- **Task 2** (not small): Introduces a new struct (`AggregationMatrix`) and a new function (`ReduceDimension`) with complex multi-dimensional coordinate logic, nested loops, matrix multiplication, and error handling. Lays down new package infrastructure.
- **Task 3** (not small): Introduces a new function (`ExpandDimension`) with complex dimension-expansion logic, nested coordinate mapping, and matrix distribution. Requires understanding of the multi-dimensional coordinate system established in Task 2 and the shared `AggregationMatrix` struct.
- **Task 4** (`[Small]`): Adds a single test file (`tensorops_test.go` only). Follows existing test patterns for edge cases (identity matrix, rank-zero tensors, single-value dimensions). Purely additive.

### Parallel Execution

None — all tasks have sequential dependencies. Tasks 2 and 3 both modify `tensorops.go` and `tensorops_test.go` (shared files), preventing parallel worktree execution.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

- Task 1 must complete first (index helpers are used by Tasks 2 and 3).
- Task 2 must complete before Task 3 (the `AggregationMatrix` struct and `ReduceDimension` function must exist in `tensorops.go` before `ExpandDimension` can be added to the same file without merge conflicts).
- Task 3 must complete before Task 4 (edge case tests exercise both `ReduceDimension` and `ExpandDimension`).

### Reintegration

None — no parallel groups. All tasks execute sequentially in the same worktree branch.

### Task 1: Create index helper functions — Completed [Small]

Create `composeIndex` and `decomposeIndex` helper functions that convert between flat row-major indices and N-dimensional coordinates, following Test-First Development.

a. Write unit tests for `composeIndex` and `decomposeIndex` in `src/main/go/tensorops/tensorops_test.go`. Include round-trip identity tests (`decompose → compose` returns the original flat index), various shapes (1D, 2D, 3D), and boundary values (index 0, last index). Tests must fail initially since the functions do not exist. — Not Started
b. Create `src/main/go/tensorops/index_helpers.go` with `composeIndex(coords []int, shape []int) int` and `decomposeIndex(index int, shape []int) []int` using row-major order (last dimension varies fastest). — Not Started
c. Verify tests pass: `go test ./src/main/go/tensorops/...` — Not Started

### Task 2: Implement AggregationMatrix struct and ReduceDimension — Completed

Define the `AggregationMatrix` struct and implement the `ReduceDimension` function that contracts an N-dimensional tensor to (N-1) dimensions along a specified collapsed dimension using a weight matrix. Follows Test-First Development.

a. Write unit tests for `ReduceDimension` in `src/main/go/tensorops/tensorops_test.go`. Include: 1D→scalar reduction with a known sum matrix, 2D→1D reduction, and 3D→2D reduction with ground-truth expected values. Also write error case tests: dimension index out of range and matrix row count mismatch. Tests must fail initially. — Not Started
b. Create `src/main/go/tensorops/tensorops.go` with:
   - `AggregationMatrix` struct (`Rows`, `Columns`, `Data [][]float64`, `RowLabels []string`, `ColLabels []string`)
   - `ReduceDimension(src *calculation.ChronoconeTensor[float64], matrix *AggregationMatrix, collapsedDim int, targetDim int) (*calculation.ChronoconeTensor[float64], error)` — builds result shape by removing `collapsedDim`, iterates over all result cells, sums over collapsed dimension: `result[...j...] = Σₖ (src[...k...] × matrix[k][j])`. Validates `src.Shape()[collapsedDim] == matrix.Rows`. Uses `calculation.NewChronoconeTensor`, `calculation.ChronoconeTensor.Shape()`, `calculation.ChronoconeTensor.Data()`, and the index helpers from Task 1. — Not Started
c. Verify tests pass: `go test ./src/main/go/tensorops/...` — Not Started

### Task 3: Implement ExpandDimension — Completed

Implement the `ExpandDimension` function that expands an N-dimensional tensor to (N+1) dimensions by distributing source values through a weight matrix. Follows Test-First Development.

a. Write unit tests for `ExpandDimension` in `src/main/go/tensorops/tensorops_test.go`. Include: scalar→1D expansion, 1D→2D expansion, and 2D→3D expansion with known distribution matrices and ground-truth expected values. Also write error case tests: dimension index out of range and matrix row count mismatch. Tests must fail initially. — Not Started
b. Implement `ExpandDimension(src *calculation.ChronoconeTensor[float64], matrix *AggregationMatrix, srcDim int, newDim int) (*calculation.ChronoconeTensor[float64], error)` in `src/main/go/tensorops/tensorops.go`. Builds expanded result shape with `matrix.Columns` inserted at `newDim`. Each result cell: `result[...j...] = src[...i...] × matrix[i][j]`. Validates `src.Shape()[srcDim] == matrix.Rows`. Reuses the `AggregationMatrix` struct and index helpers from Tasks 1 and 2. — Not Started
c. Verify tests pass: `go test ./src/main/go/tensorops/...` — Not Started

### Task 4: Edge case tests and validation — Completed [Small]

Add edge case tests that exercise both `ReduceDimension` and `ExpandDimension` with boundary conditions to complete test coverage. Follows Test-First Development (write tests, then verify they pass against the already-implemented functions).

a. Write edge case tests in `src/main/go/tensorops/tensorops_test.go`: identity matrix no-op (matrix with 1.0 on diagonal, result equals source), rank-zero tensors (scalar input to `ExpandDimension`), single-value dimensions, and verify all tests exercise both `ReduceDimension` and `ExpandDimension`. — Not Started
b. Verify tests pass: `go test ./src/main/go/tensorops/...` — Not Started

## Test-First Development

This story follows the Logical TDD Lifecycle. For each coding task, write tests before implementing the code. The test-and-implement cycle is integrated within each task — tests are subtask (a), implementation is subtask (b), and verification is the final subtask. Do not separate tests into standalone tasks.

## Constraints

- This is pure Go tensor math — no CPL parsing, analysis, or CIL dependencies.
- The `tensorops` package lives at `src/main/go/tensorops/` (a separate package from `calculation`). It MUST use the exported accessor methods of `calculation.ChronoconeTensor` (`Shape()`, `Data()`, `DimensionNames()`, `IsScalar()`) and the `calculation.NewChronoconeTensor` constructor — it cannot access private fields.
- Index conversion follows row-major order (last dimension varies fastest).
- `ReduceDimension` removes exactly one dimension; `ExpandDimension` adds exactly one dimension.
- All arithmetic uses `float64`. No type parameterization needed in this increment.
- The `AggregationMatrix.RowLabels` and `AggregationMatrix.ColLabels` fields store labels for debugging but are not used in math operations in this increment.
- Test data must be hand-constructed with known expected values — no CPL files, IO specs, or parquet data.

## Intent

### Task 1 Intent

Provide the fundamental coordinate↔index conversion utilities needed by both `ReduceDimension` and `ExpandDimension`. These helpers are the building blocks for all tensor traversal in this story.

### Task 2 Intent

Deliver the `AggregationMatrix` data structure and the dimension-reducing operation. `ReduceDimension` is the core operation for aggregators — it contracts a tensor along one dimension, summing weighted contributions into a lower-dimensional result.

### Task 3 Intent

Deliver the dimension-expanding operation. `ExpandDimension` is the core operation for spreaders — it expands a tensor by one dimension, distributing source values through a weight matrix into a higher-dimensional result.

### Task 4 Intent

Validate both operations with boundary conditions (identity matrices, rank-zero tensors, single-value dimensions) to ensure correctness across edge cases that later increments will exercise.

## Acceptance Criteria

1. `composeIndex` and `decomposeIndex` round-trip is identity for all tested shapes
2. `ReduceDimension` correctly reduces 1D→1D, 2D→2D, 3D→3D with ground-truth expected values
3. `ExpandDimension` correctly expands scalar→1D, 1D→2D, 2D→3D with ground-truth expected values
4. Dimension index out-of-range and matrix size mismatch errors are returned with clear messages
5. Identity matrix is a no-op (result equals source)
6. All tests pass with `go test ./src/main/go/tensorops/...`
7. No CPL files, grammars, or parsing infrastructure needed
8. The tensorops package is independently compilable and testable (no dependency on the rest of the CPL compiler)

## Notes

- The `tensorops` package imports only `calculation` (for `ChronoconeTensor`) and `fmt` (for error messages). No other project packages are required.
- Run tests from `source_root` (`src/chronocone_planning_language`): `cd src/chronocone_planning_language && go test ./src/main/go/tensorops/... -count=1 -timeout 60s`
- The `src/main/go/tensorops/` directory does not exist yet — it will be created by Task 1 (for `index_helpers.go`) and Task 2 (for `tensorops.go`).
- Task 2 creates `tensorops.go` first in the sequential execution order; Task 3 only adds the `ExpandDimension` function to the existing file.
