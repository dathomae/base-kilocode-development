# Story192: CPL Metrics Time Shifting — Synthesize

## Goal

Implement the Increment 6 synthesis phase for CPL metric time shifting (backward/forward) as part of [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md). This story covers:

- Define `CILTimeShift` node type representing a time-shifted tensor access in the CIL expression tree.
- Extend the CIL tree walker (`walkToPostfix`) and visitor (`VisitCILTimeShift`) to generate Go code for array index offset calculations (`currentCellIndex ± shiftAmount`).
- Handle self-reference boundary conditions (before start of time → zero value).
- Generate code that respects topological ordering for self-referencing metrics (earlier time cells computed before later ones), so that backward references access already-computed values.

## References

- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [ERO005: Metrics Synthesis Phase — Increment 6](../specifications/design/cpl/go-synthesis-target/eros/ERO005_metrics-synthesis-phase.md)
- [Synthesis Design](../specifications/design/cpl/go-synthesis-target/synthesis.md)
- [Tensor Representation Design](../specifications/design/cpl/go-synthesis-target/tensor-representation.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story191_cpl-metrics-time-shift-parse-analyze.md](finished-stories/Story191_cpl-metrics-time-shift-parse-analyze.md) — _Done_ — parse/analyze phase must produce validated `AnalysisTimeShift` data
- [Story185_cpl-metrics-arithmetic-synthesize.md](finished-stories/Story185_cpl-metrics-arithmetic-synthesize.md) — _Done_ — arithmetic synthesis infrastructure (`CILAbstractTransformer`, `WalkCILTreeToPostfix`, `EvaluatePostfixFloat64`) must be in place

## Dependent Stories

- [Story194_cpl-metrics-dag-io-analyze.md](Story194_cpl-metrics-dag-io-analyze.md) — _Not Started_ — Increment 8 Analysis (DAG construction must account for time-shifted self-references as dependencies, not cycles)
- [Story195_cpl-metrics-dag-io-synthesize.md](Story195_cpl-metrics-dag-io-synthesize.md) — _Not Started_ — Increment 8 Synthesis

## Tasks

### Task Sizing

- **Task 1** (CILTimeShift node, walker, visitor, and boundary handling): Medium — introduces a new CIL node type requiring coordinated changes across `cil/` package files and the postfix evaluator. Touches 3–4 files: node definition, walker case, visitor method, and evaluator time-index logic. Includes boundary condition handling (index < 0 → zero value).
- **Task 2** (Self-referencing computation ordering): Medium — modifies the generated computation loop to iterate time cells in ascending order for self-referencing metrics. Requires understanding of how `EvaluatePostfixFloat64` and tensor arithmetic functions operate on time-indexed cells.
- **Task 3** (End-to-end verification): Medium — creates example CPL files with self-referencing time-shifted metrics, generates parquet test data with a time dimension, runs the full pipeline, and verifies correctness of time-shifted calculations.

### Parallel Execution

None — all tasks have sequential dependencies. Task 1 must complete first (CILTimeShift node and walker must exist before the computation loop can reference them). Task 2 depends on Task 1 (the computation ordering loop references the CILTimeShift node type and time-index access pattern). Task 3 depends on both.

### Execution Order

```
Task 1 → Task 2 → Task 3
```

### Reintegration

None — no parallel groups.

### Task 1: Define CILTimeShift node and implement VisitCILTimeShift with boundary condition handling — Not Started
   a. Define `CILTimeShift` struct in `src/main/go/cil/cil_nodes.go` (or new file `src/main/go/cil/cil_time_shift.go`):
      - `Direction string` — "backward" or "forward"
      - `Field string` — constraint field name (e.g., "year", "month")
      - `Amount float64` — shift amount (default 1)
      - `Operand CILNode` — the operand being time-shifted (typically a `CILInputRef`) — Not Started
   b. Add `cilNode()` marker method so `CILTimeShift` implements `CILNode` — Not Started
   c. Add `case *CILTimeShift` to `walkToPostfix` in `src/main/go/cil/postfix_walk.go`: recursively walk the operand, then append a "time_shift" postfix op with direction, field, and amount — Not Started
   d. Add `time_shift` op kind handling to `EvaluatePostfixFloat64` in `src/main/go/calculation/postfix_op.go`: pop the operand tensor and apply time-index offset — Not Started
   e. Implement boundary condition guard: when the shifted time index would be < 0 (backward past start) or >= tensor size (forward past end), return the zero value (`0.0` for float64) — Not Started
   f. Implement `VisitCILTimeShift` method on `CILAbstractTransformer` in `src/main/go/cil/cil_transformer_visitors.go` for the code-gen path (used when the transformer directly walks CIL trees rather than going through postfix evaluation) — Not Started
   g. Add `case *CILTimeShift` to `TransformMetric` in `src/main/go/cil/cil_transformer_visitors.go` to dispatch to `generatePostfixComputeBody` — Not Started
   h. Write tests for:
      - `walkToPostfix` produces correct "time_shift" postfix op with direction, field, amount
      - `EvaluatePostfixFloat64` applies backward shift (cell at index `t - shiftAmount`)
      - `EvaluatePostfixFloat64` applies forward shift (cell at index `t + shiftAmount`)
      - Boundary condition: backward shift before start returns 0.0
      - Boundary condition: forward shift past end returns 0.0 — Not Started
   i. Verify all CIL and calculation package tests pass: `go test ./src/main/go/cil/... ./src/main/go/calculation/...` — Not Started

### Task 2: Implement self-referencing metric computation ordering (time-ascending loop) — Not Started
   a. Identify the computation loop that iterates over time cells — currently the `EvaluatePostfixFloat64` function and tensor arithmetic operations process tensors element-by-element in the order returned by `Data()`. For self-referencing metrics (where a metric input references itself via backward/forward), the time-iteration must be ascending (`t = 0 → N-1`) so that backward references access already-computed cells — Not Started
   b. Design the ordering mechanism: when a metric's CIL tree contains a self-referencing `CILTimeShift` node (operand name matches metric name), the generated compute function must iterate time cells in ascending order and compute each cell individually rather than using bulk tensor operations — Not Started
   c. Update `generatePostfixComputeBody` or `TransformMetric` in `src/main/go/cil/cil_code_generator.go` (or `cil_transformer_visitors.go`) to detect self-referencing CILTimeShift nodes and emit an ascending cell-by-cell loop — Not Started
   d. The generated loop must:
      - Iterate `cellIndex` from `0` to `totalCells-1`
      - For each cell, evaluate the postfix table with `cellIndex` as context
      - Access time-shifted cells at `cellIndex ± shiftAmount * timeStride` where `timeStride` is the product of non-time dimension sizes — Not Started
   e. Write tests for:
      - Self-referencing metric (e.g., `calculate revenue + revenue.backward(year)`) generates ascending cell-by-cell loop
      - Non-self-referencing metric generates normal bulk tensor evaluation (no regression)
      - Backward reference at cell 0 returns 0.0 (boundary condition in ordered context)
      - Backward reference at cell t>0 returns the value computed at cell t-1 — Not Started
   f. Verify synthesis tests pass — Not Started

### Task 3: End-to-end verification — Not Started
   a. Create example CPL file with a time-shifted metric calculation:
      - Connector providing input data with a time dimension (e.g., monthly revenue data)
      - A simple metric that uses backward shift: `calculate revenue + revenue.backward(year)` — Not Started
   b. Create example CPL file with forward shift: `calculate revenue - revenue.forward(month)` — Not Started
   c. Create corresponding IO specifications and test parquet data with time dimension values — Not Started
   d. Run parse → analyze → synthesize → compile → execute pipeline:
      - Verify backward-shifted values match the value from the previous time cell
      - Verify forward-shifted values match the value from the next time cell — Not Started
   e. Test edge cases:
      - Shift at time boundary (first cell with backward shift → zero value)
      - Multiple time-shifted inputs on the same metric
      - Self-referencing metric with both time-shifted and non-time-shifted inputs — Not Started
   f. Verify boundary values: out-of-bounds backward shift returns `0.0`; out-of-bounds forward shift returns `0.0` — Not Started
   g. Run full CPL test suite to confirm no regressions: `go test ./... -count=1` from `src/chronocone_planning_language/src/main/go/` — Not Started

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task above, tests are written before implementation and the test-and-implement cycle is integrated within each task rather than broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Time shift index calculation must handle boundary conditions (index < 0 or >= tensor size)
3. Self-referencing metrics must compute in time-ascending order to ensure backward references access already-computed values
4. Zero value for out-of-bounds shifts: the value is `0.0` (float64 default)
5. Non-self-referencing metrics must continue to use bulk tensor operations with no ordering change (no regression)
6. Generated Go code must be compilable and runnable
7. Go functions that can fail must return `(T, error)`
8. Generated Go tests must follow the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Intent

- **Task 1 Intent**: Introduce the CILTimeShift node type, integrate it into the postfix walker and evaluator, and handle time-index boundary conditions — this makes time-shifted tensor access a first-class operation in the CIL expression tree
- **Task 2 Intent**: Ensure correct computation ordering for metrics that reference themselves through time shifts — the generated code must iterate time ascending so backward references find already-computed values
- **Task 3 Intent**: Prove the time shifting pipeline works end-to-end with correct temporal calculations, boundary handling, and no regressions

## Acceptance Criteria

1. `CILTimeShift` implements `CILNode` and carries direction, constraint field, shift amount, and operand
2. `walkToPostfix` produces correct `time_shift` postfix op for backward/forward shifts
3. `EvaluatePostfixFloat64` applies time-index offset and handles boundary conditions (out-of-bounds → 0.0)
4. `VisitCILTimeShift` generates correct Go code for time-shifted tensor access
5. Self-referencing metrics generate ascending cell-by-cell computation loops
6. Non-self-referencing metrics continue to use bulk tensor evaluation (no regression)
7. Full end-to-end test: time-shifted calculations produce correct temporal results
8. All unit tests pass following Go testing best practices
9. No regressions in existing CPL test suite
