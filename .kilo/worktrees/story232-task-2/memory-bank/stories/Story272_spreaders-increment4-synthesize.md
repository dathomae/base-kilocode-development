# Story272: Spreaders Increment 4 (Synthesize)

## Goal

Implement synthesis-phase optimization for mixed aggregator/spreader interleaved sequences. When the Increment 4 analyzer produces interleaved `CILSpreadNode`/`CILAggregateNode` chains, the synthesizer analyzes adjacent operations for fusability: operations on different dimensions can be combined into single matrix multiplications, reducing intermediate tensor allocations and improving runtime performance. Two spreads on different dimensions fuse into one expand; an aggregate and spread on different dimensions fuse into a single transform. Operations on the same dimension cannot fuse (data dependency). Scalar spreading (`from []`) is handled as a special case within interleaved sequences. Fused matrices are pre-computed at synthesis time and emitted in the preamble. For non-fusable sequences, the Increment 2/3 sequential emission path is preserved as a fallback.

Part of Epic008 (Spreaders Incremental Implementation). This is Increment 4 of the spreaders epic.

## References

- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md) — Increment 4 sketch
- [ERO008: Spreaders Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md) — Increment 4 data structures, fused operations, generated code examples
- [Plan: Story272](../plans/Story272_spreaders-increment4-synthesize.md)

## Dependencies

- Story271 (spreaders-increment4-analyze) — analysis must produce interleaved `CILSpreadNode`/`CILAggregateNode` chains selected by the optimization algorithm (Not Started)
- Story270 (spreaders-increment3-synthesize) — metric reference walking and dynamic matrix construction infrastructure must exist
- Story268 (spreaders-increment2-synthesize) — spreader chain walking must exist
- Story261 (aggregators-increment4-synthesize) — aggregator chain walking and metric function generation infrastructure must exist; fused operations combine both aggregator and spreader patterns
- Story254 (aggregators-increment1-matrix-mult) — `tensorops.ReduceDimension` and `tensorops.ExpandDimension` are the synthesis primitives

## Dependent Stories

None — this is the final synthesis story in the spreaders epic.

## Tasks

### Task Sizing

No tasks qualify as `[Small]`. All tasks introduce new abstractions (Task 1: `FusedGroup` struct and new file infrastructure), add new methods requiring walker scope understanding (Task 2: `EmitFusedSpreaderSequence`), handle multiple edge cases requiring rank-transition reasoning (Task 3), touch 4+ files across packages with coordinated integration changes (Task 4), or execute broad cross-package regression verification (Task 5). Architect-for-story-planning confirmed all tasks are at their natural granularity — no decomposition seams exist that would produce isolated units without artificial coordination overhead.

### Parallel Execution

None — all tasks have sequential dependencies. Each task builds on the output of the preceding task. File analysis confirms: Tasks 1-2 share `fused_sequence_emitter.go`; Tasks 3-4 modify walker files that depend on the fusion infrastructure from Tasks 1-2; Task 5 verifies everything.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4 → Task 5
```

- Task 1 must complete first — fusion grouping is the foundation for all subsequent work
- Task 2 depends on Task 1 — fusion uses grouped operations
- Task 3 depends on Tasks 1 and 2 — edge cases build on fused emission infrastructure
- Task 4 depends on all preceding tasks — integration uses the complete fusion + scalar pipeline
- Task 5 is final verification

### Reintegration

None — no parallel groups.

### Task 1: Implement fusion grouping and fusable condition analysis — Not Started

Implement the foundational fusion grouping logic: define the `FusedGroup` struct, implement operation grouping, and implement the fusability predicate (`canFuse`).

**Intent**: Provide the data structure and grouping logic that identifies which adjacent operations in an interleaved spreader/aggregator sequence can be fused into single matrix operations. This is the analysis phase of fusion — it determines *whether* two operations can fuse and groups fusable operations together, but does not yet perform the matrix fusion or code emission.

**Acceptance Criteria**:
- `FusedGroup` struct correctly represents a group of adjacent operations with their combined dimension information
- `groupFusableOperations` correctly scans a sequence and partitions it into maximal groups of pairwise-fusable operations
- `canFuse` returns true for: spread+spread on different dims, aggregate+spread on different dims, spread+aggregate on different dims
- `canFuse` returns false for: aggregate+aggregate (matrices have different collapsed dims), aggregate+spread on same dim, spread+spread on same dim
- All grouping unit tests pass: `go test ./src/main/go/synthesis/... -run Fused -count=1`

1. Create `src/main/go/synthesis/fused_sequence_emitter.go` — define the `FusedGroup` struct holding a slice of `*DimensionalOp` references, the combined operation kind, and the union of affected dimensions — Not Started
2. Write unit tests in `src/main/go/synthesis/fused_sequence_emitter_test.go` — test cases: adjacent spreads on different dims are fusable; aggregate+spread on different dims are fusable; aggregate+spread on same dim are NOT fusable; two aggregates are NOT fusable; three adjacent spreads on different dims group into one FusedGroup — Not Started
3. Implement `canFuse(op1, op2 *FusedGroup) bool` — checks dimension independence (no data dependency conflict) and operation kind compatibility per the fusable conditions from the plan — Not Started
4. Implement `groupFusableOperations(ops []*DimensionalOp) []*FusedGroup` — scans the sequence left-to-right, greedily groups adjacent fusable operations, returns maximal groups — Not Started
5. Run `go test ./src/main/go/synthesis/... -run Fused -count=1` and verify all tests pass — Not Started

### Task 2: Implement matrix fusion and fused code emission — Not Started

Implement the core fusion execution: combine matrices from fusable groups into single `AggregationMatrix` values and emit the fused operations as `ExpandDimension`/`ReduceDimension`/`Transform` calls.

**Intent**: Transform the groups produced by Task 1 into executable code. For each FusedGroup, compute the fused matrix (multiplying individual operation matrices), then emit a single tensor operation call rather than sequential calls. This is where the performance optimization materializes — 3 sequential tensor ops become 1.

**Acceptance Criteria**:
- `fuseGroups` produces a single `AggregationMatrix` whose dimensions match the combined effect of sequential individual matrices
- `EmitFusedSpreaderSequence` walks the ops list, calls grouping/fusion, and emits One `ExpandDimension` per fused expand group, one `ReduceDimension` per fused reduce group
- Variable naming follows walker conventions (scope.freshVar), intermediate variables are tracked in scope
- Fused matrix preamble variable names are generated correctly
- All fusion emission tests pass: `go test ./src/main/go/synthesis/... -run Fused -count=1`

1. Write unit tests in `src/main/go/synthesis/fused_sequence_emitter_test.go` — test cases: fused matrix for two adjacent spreads has correct output dimensions; fused matrix for aggregate+spread produces same result as sequential; EmitFusedSpreaderSequence emits correct Go code for a two-spread fused sequence; emitted code for aggregate+spread fused sequence uses the correct tensor operation; three-spread fused sequence emits one ExpandDimension call — Not Started
2. Implement `fuseGroups(g1, g2 *FusedGroup) *FusedGroup` — combines matrices from two fusable groups into a single `AggregationMatrix` by matrix multiplication; the combined group preserves the union of affected dimensions — Not Started
3. Implement `EmitFusedSpreaderSequence` as a method on `*CILTreeWalker` — accepts `[]*DimensionalOp`, source variable name, and `*WalkScope`; calls `groupFusableOperations`, iterates groups fusing adjacent ones via `fuseGroups` where possible, emits one tensor operation per resulting group, returns the final result variable — Not Started
4. Implement emission helpers: emit `ExpandDimension` for fused expand groups using `scope.emitExpandDimensionCall`; emit `ReduceDimension` for fused reduce groups using `scope.emitReduceDimensionCall`; emit preamble variable declarations for fused matrices — Not Started
5. Run `go test ./src/main/go/synthesis/... -run Fused -count=1` and verify all tests pass — Not Started

### Task 3: Implement scalar spreading and edge case handling — Not Started

Handle the `from []` (scalar) spreader case within interleaved sequences and edge cases where fused sequences begin or end at rank-zero.

**Intent**: Scalar spreading (`from []`) produces a rank-zero to rank-one expansion. Within an interleaved fused sequence, this must be handled correctly — the source for a fused sequence may be a scalar (rank-zero input), or the result may collapse back to scalar. This task ensures the synthesizer correctly tracks tensor ranks through fused operations that include scalar transitions.

**Acceptance Criteria**:
- `from []` spreaders within interleaved sequences emit `ExpandDimension` with `dim=0, pos=0` (scalar→1D expansion)
- Fused sequence with scalar source (rank-zero input, multi-dim output) emits correct expansion chain
- Fused sequence ending in scalar (multi-dim input, rank-zero output) emits correct reduction chain
- Scalar→scalar fused sequence (no-op) is handled without emitting unnecessary operations
- All scalar edge case tests pass: `go test ./src/main/go/synthesis/... -run "Scalar|Fused" -count=1`

1. Write unit tests in `src/main/go/synthesis/fused_sequence_emitter_test.go` — test cases: scalar→1D→2D fused sequence emits two ExpandDimension calls; 2D→scalar fused sequence emits correct ReduceDimension; scalar→scalar no-op emits no tensor operations; `from []` spreader in middle of interleaved sequence is handled correctly — Not Started
2. Implement `handleScalarSpreadInSequence(node *CILSpreadNode, sourceVar string, scope *WalkScope)` — detects `node.FromEmpty`, emits `ExpandDimension(sourceVar, matrix, 0, 0)` for rank-zero to rank-one expansion, registers the result variable in scope — Not Started
3. Handle scalar source edge case: when the entire fused sequence begins with a scalar input, the first operation must be a rank-0→rank-N expansion; track input rank through the fusion pipeline — Not Started
4. Handle scalar result edge case: when the fused sequence ends with reduction to scalar, the final operation must collapse to rank-zero; verify the last fused operation is a ReduceDimension to scalar — Not Started
5. Run `go test ./src/main/go/synthesis/... -run "Scalar|Fused" -count=1` and verify all tests pass — Not Started

### Task 4: Integrate into metric computation pipeline and create examples — Not Started

Wire the fused sequence emission into the metric computation code generator and create end-to-end examples that demonstrate the optimization.

**Intent**: Connect the fusion infrastructure (Tasks 1-3) to the actual synthesis pipeline. Update the walker to call `EmitFusedSpreaderSequence` instead of simple chain walking when interleaved sequences are present. Ensure fused matrices appear in the preamble. Create CPL examples that exercise the optimization and verify correctness.

**Acceptance Criteria**:
- Metric computation generator calls `EmitFusedSpreaderSequence` for interleaved spreader/aggregator chains
- All fused matrices are emitted as Go `var` declarations in the preamble
- Full generated program with fused sequences compiles and produces correct output
- `with_data/spreaders_optimization/` examples are created and passed: interleaved aggregate-spread, ambiguous spreader selection, fused vs non-fused comparison, non-fusable sequence (same dim)
- All integration tests pass: `go test ./src/main/go/synthesis/... -count=1`
- Example verification: `go test ./src/main/go/... -run SpreadersOpt -count=1`

1. Update `src/main/go/synthesis/cil_spread_walker.go` — modify the spreader chain walking path to detect interleaved `CILSpreadNode`/`CILAggregateNode` chains and dispatch to `EmitFusedSpreaderSequence` instead of the simple chain walker — Not Started
2. Update `src/main/go/synthesis/cil_aggregate_walker.go` — ensure the aggregator walker integrates with the fused sequence path; when an aggregate node appears in a chain with spread nodes, the combined walker path produces interleaved ops for fusion — Not Started
3. Ensure fused matrix preamble variables are emitted — update the preamble generator to call `EmitFusedMatrix` for each fused matrix produced by `EmitFusedSpreaderSequence`; verify all matrices appear as `var` declarations in the generated Go preamble — Not Started
4. Write integration tests in `src/main/go/synthesis/fused_sequence_emitter_test.go` — test cases: full generated program with fused sequence compiles; fused program output matches non-fused program output (correctness invariant); preamble contains all fused matrix declarations; error from fused operation is wrapped correctly; non-fused Increment 2/3 sequences still produce correct output — Not Started
5. Create `with_data/spreaders_optimization/` examples:
   a. Create `with_data/spreaders_optimization/spreaders_optimization.cpl` — CPL definition with mixed aggregator+spreader metric producing interleaved sequences — Not Started
   b. Create `with_data/spreaders_optimization/spreaders_optimization.io.yaml` — IO spec defining input/output dimensions — Not Started
   c. Create `with_data/spreaders_optimization/spreaders_optimization_input.xlsx` — input data exercising all fusion scenarios — Not Started
   d. Create `with_data/spreaders_optimization/spreaders_optimization.expected.json` — expected output for verification — Not Started
6. Run `go test ./src/main/go/... -run SpreadersOpt -count=1 -timeout 120s` and verify all examples produce output matching `.expected.json` — Not Started

### Task 5: Regression verification — Not Started

Verify that the fusion optimization does not break existing behavior from earlier spreader increments.

**Intent**: The fusion path is a synthesis-phase optimization — it must produce identical output to the sequential (non-fused) path for all valid inputs. This task confirms the correctness invariant and ensures no regressions in the Increment 2 and 3 synthesis paths.

**Acceptance Criteria**:
- Increment 2 single-spreader sequences continue to work (non-fused path)
- Increment 3 metric reference walking works correctly with fused sequences
- Full test suite passes with no regressions
- Fused path produces fewer intermediate tensor allocations than sequential path (verified by shape trace in tests)

1. Verify Increment 2 single-spreader sequences — run existing spreader tests and confirm output is unchanged: `go test ./src/main/go/synthesis/... -run Spreader -count=1` — Not Started
2. Verify Increment 3 metric references — run existing metric reference tests with fused sequences enabled: `go test ./src/main/go/synthesis/... -run MetricRef -count=1` — Not Started
3. Run full test suite: `go test ./src/main/go/... -count=1 -timeout 120s` — confirm zero test failures and no new compilation errors — Not Started

## Test-First Development

This story follows the Logical TDD Lifecycle. Each coding task includes:
1. Write tests that define the expected behavior (the subtask labeled "Write unit tests")
2. Implement the code to make the tests pass (the implementation subtasks)
3. Run the tests to verify correctness (the verification subtask)

Tests and implementation are integrated within each task rather than broken out as separate tasks.

## Constraints

- **Correctness invariant**: Fused output must be identical to sequential (non-fused) output for all valid inputs. Fusion is purely a performance optimization.
- **Non-fusable fallback**: When operations cannot be fused (same dimension, incompatible shapes), the Increment 2/3 sequential emission path is preserved. Do not remove or bypass the existing chain-walking code.
- **Tensor operations**: All fused outputs use `tensorops.ExpandDimension`, `tensorops.ReduceDimension`, or a combined `tensorops.ApplyFusedTransform`. No new tensor operation primitives are introduced.
- **Preamble correctness**: All fused matrices must be declared in the preamble before any computation function references them. Missing declarations cause compilation failures.
- **Scope discipline**: Variable names generated by `EmitFusedSpreaderSequence` must follow the walker's `scope.freshVar` convention and be tracked in scope for downstream consumers.
- **Source root**: All Go paths are relative to `src/chronocone_planning_language`. Build and test commands run from `src/chronocone_planning_language`.

## Intent

This story delivers the synthesis-phase optimization for mixed aggregator/spreader sequences. When the Increment 4 analyzer (Story271) produces interleaved `CILSpreadNode`/`CILAggregateNode` chains, the synthesizer analyzes adjacent operations for fusability: operations on different dimensions can be combined into single matrix multiplications, reducing intermediate tensor allocations and improving runtime performance. Two spreads on different dimensions fuse into one expand; an aggregate and spread on different dimensions fuse into a single transform. Operations on the same dimension cannot fuse (they have a data dependency). Scalar spreading (`from []`) is handled as a special case within interleaved sequences. The fused matrices are pre-computed at synthesis time and emitted in the preamble. The `EmitFusedSpreaderSequence` function in the walker replaces the simple chain walking from Increment 2 — it groups, fuses, and emits optimized operations. For non-fusable sequences, the Increment 2/3 sequential emission path is preserved as a fallback.

Note: The fusion optimization is purely a synthesis-phase optimization — it does not change the analysis results. The analysis selects the optimal sequence; the synthesis optimizes how that sequence is executed.

## Notes

- The `fused_sequence_emitter.go` file is new infrastructure. Tasks 1 and 2 both modify this file and its test — they run sequentially in the same worktree or sequentially across worktrees.
- The example directory `with_data/spreaders_optimization/` follows the existing pattern from earlier spreader increments (e.g., `with_data/spreaders_standard/`).
- The `EmitFusedSpreaderSequence` function signature follows the existing walker pattern: `(ops []*DimensionalOp, sourceVar string, scope *WalkScope) (string, error)`.
- Fusable conditions are defined in the plan: two operations fuse when they operate on different dimensions, intermediate tensor shapes are compatible, and the combined operation produces the same result as sequential execution.
