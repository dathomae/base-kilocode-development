# Story215: CPL Metrics Parallel Branch Execution

## Goal

Convert the sequential branch evaluation in the conditional postfix evaluator to parallel goroutine dispatch. Each conditional branch body is computed concurrently, and results are combined when all branches complete. Additionally, add mask-aware computation skipping within each branch so only elements whose mask bit is `0` (true/selected) are computed.

This story is part of Epic005 (CPL Metrics Implementation). It builds on the sequential conditional evaluator from Story188 and may leverage patterns from Story186 (parallelism infrastructure).

## References

- [ERO005 Metrics Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO005_metrics-synthesis-phase.md)
- [CPL Metrics Design](../specifications/design/cpl/metrics-design.md)
- [CPL Evaluation Model](../specifications/architecture/cpl/execution-model.md)
- [`postfix_op.go`](../../src/chronocone_planning_language/src/main/go/calculation/postfix_op.go:1) — conditional evaluator added in Story188
- [`chronocone_tensor.go`](../../src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:1) — tensor types and mask infrastructure

## Dependencies

- [Story188_cpl-metrics-conditionals-synthesize.md](Story188_cpl-metrics-conditionals-synthesize.md) — _Done_ — Provides the sequential conditional evaluator with mask infrastructure (condition mask, TensorWhereFloat64, branch body evaluation)
- [Story186_cpl-metrics-parallelism-synthesize.md](Story186_cpl-metrics-parallelism-synthesize.md) — _Not Started_ — May provide useful patterns for worker pool integration, though branch parallelism is a different problem (few branches, each doing significant work) vs tensor-chunk parallelism (many chunks, each doing small work)
- [Story227_cpl-conditional-concept-values-synthesize.md](finished-stories/Story227_cpl-conditional-concept-values-synthesize.md) — _Done_ — Conditional concept tensors will have pre-computed validity masks set by synthesis, so parallel branch evaluation can skip invalid cells without additional work

## Dependent Stories

- [Story227_cpl-conditional-concept-values-synthesize.md](Story227_cpl-conditional-concept-values-synthesize.md) — _Not Started_

## Tasks

1. Add mask-aware element filtering to arithmetic ops — Not Started
   a. Write tests for mask-aware TensorAddFloat64, TensorMultiplyFloat64, etc. — Not Started
   b. Modify existing tensor arithmetic functions so that when a mask is present, only elements with mask bit `0` are computed — Not Started
   c. Verify elements with mask bit `1` retain their existing value in the output — Not Started
   d. Run tests and benchmarks to confirm correctness and measure improvement — Not Started

2. Refactor conditional evaluation to launch goroutines per branch — Not Started
   a. Write tests for concurrent branch evaluation (2-branch, 4-branch, 8-branch conditionals) — Not Started
   b. Replace the sequential for-loop over branches in EvaluatePostfixFloat64 conditional handler with goroutine dispatch — Not Started
   c. Use sync.WaitGroup to track branch goroutines and pre-allocated result slice — Not Started
   d. Have the main goroutine wait for all branches, then call TensorWhereFloat64 — Not Started
   e. Run integration tests to verify identical results to sequential execution — Not Started

3. Add branch-level error collection and cancellation — Not Started
   a. Write tests for error propagation when one branch fails — Not Started
   b. Add error channels or errgroup integration so that if any branch goroutine returns an error, the parent collects and returns it — Not Started
   c. Use context.Context for cancellation — if one branch fails, cancel the others — Not Started
   d. Verify no goroutine leaks under error conditions — Not Started

4. Performance tests — Not Started
   a. Write benchmarks comparing sequential vs parallel branch evaluation — Not Started
   b. Benchmark 2-branch (if/else), 4-branch, and 8-branch conditionals — Not Started
   c. Benchmark various tensor sizes (100 to 1M elements) — Not Started
   d. Benchmark dense masks (all-true) vs sparse masks (10% true) — Not Started
   e. Analyze results and document speedup characteristics — Not Started

5. Integration tests — Not Started
   a. Write tests verifying parallel branch execution produces identical results to sequential execution — Not Started
   b. Run the full set of conditional example CPL files through the parallel path — Not Started
   c. Verify backward compatibility: sequential path remains as fallback — Not Started
   d. Confirm no data races with `go test -race` — Not Started

## Test-First Development

This project follows the Logical TDD Lifecycle (BTDD). For each coding task above, tests are written before implementation. The test-and-implement cycle is integrated within each task rather than broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- Must not introduce data races — branch body evaluations read from shared input tensors but write to separate result tensors
- Must be backward compatible — sequential path remains as fallback
- Follow Logical TDD Lifecycle (BTDD)
- Follow Go testing best practices
- Condition evaluation stays sequential — the condition mask must be computed before branches can evaluate
- The parallel path is a drop-in optimization behind the same `EvaluatePostfixFloat64` interface

## Intent

For each task:

1. **Task 1 (Mask-aware arithmetic)**: The foundational optimization that makes branch evaluation efficient by skipping computation for masked-out elements.
2. **Task 2 (Goroutine dispatch)**: The core parallelism improvement that converts sequential branch evaluation to concurrent execution.
3. **Task 3 (Error handling)**: Ensure robust error propagation and cancellation so that partial failures don't leave dangling goroutines or incomplete results.
4. **Task 4 (Performance tests)**: Quantify the improvement and identify the tensor sizes and mask densities where parallel execution is beneficial.
5. **Task 5 (Integration tests)**: Guarantee correctness and backward compatibility before the parallel path is considered production-ready.

## Acceptance Criteria

- All existing conditional metric examples produce identical results under both sequential and parallel evaluation
- `go test -race` passes without any detected races
- Benchmarks demonstrate measurable speedup for multi-branch conditionals on sufficiently large tensors
- Error conditions in any branch cancel all other branches and return the error to the caller
- The sequential evaluation path remains available and functional as a fallback

## Open Questions

None

## Known Issues

None

## Additional Context

This story addresses two performance opportunities left on the table by Story188's sequential conditional evaluator:

1. **Parallelism opportunity**: Each branch body is independent — they only depend on the condition mask, not on each other. They can be computed concurrently on separate goroutines.
2. **Mask-aware skipping**: Currently all positions are computed in every branch, even though `Where` only selects data from positions where that branch's mask is `0`. For large tensors with sparse masks, this wastes significant computation.
