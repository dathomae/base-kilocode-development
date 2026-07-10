# Story186: CPL Metrics Parallelism Synthesize

## Goal

Implement the Increment 3 synthesis phase for CPL metric tensor calculation parallelism as part of [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md). This is a **synthesis-only increment** — no parser or analyzer changes. The architecture configuration embedding in generated code already exists in [`GenerateMainFunction()`](src/chronocone_planning_language/src/main/go/synthesis/synthesize.go:262-266) and [`GenerateSetupFunction()`](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go:298). This story covers:

- Implement the TODO methods in [`ChronoconeComputationController`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go): `SetArchitectureConfig()`, `StartWorkers()`, `StopWorkers()`, `SubmitChunk()`, `GetWorkChannel()`, `RegisterMetricComputation()`, `GetMetricComputation()`.
- Determine worker goroutine count from CPU cores and configure chunk size optimization from cache size.
- Generated code already uses high-level tensor operation methods which internally handle chunking and parallel execution via the work channel and worker pool.

## References

- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [ERO005: Metrics Synthesis Phase — Increment 3](../specifications/design/cpl/go-synthesis-target/eros/ERO005_metrics-synthesis-phase.md)
- [Synthesis Design](../specifications/design/cpl/go-synthesis-target/synthesis.md)
- [Tensor Representation Design](../specifications/design/cpl/go-synthesis-target/tensor-representation.md)
- [IO Specifications Requirements](../specifications/requirements/cpl/io_specs.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story185_cpl-metrics-arithmetic-synthesize.md](finished-stories/Story185_cpl-metrics-arithmetic-synthesize.md) — arithmetic synthesis must be complete, providing high-level tensor operation methods (Done)

## Dependent Stories

- [Story194_cpl-metrics-dag-io-analyze.md](Story194_cpl-metrics-dag-io-analyze.md) — Increment 8 Analysis (Not Started)
- [Story195_cpl-metrics-dag-io-synthesize.md](Story195_cpl-metrics-dag-io-synthesize.md) — Increment 8 Synthesis (Not Started)

## Tasks

1. Implement ChronoconeComputationController methods — Not Started
   a. Implement `SetArchitectureConfig()` (currently TODO at [`chronocone_tensor.go:169`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:169)): store CPU cores, cache size, main memory; initialize work channel with buffer size = core count — Not Started
   b. Implement `RegisterMetricComputation()` (TODO at line 181): add `MetricComputation` to `metricComputations` map — Not Started
   c. Implement `GetMetricComputation()` (TODO at line 194): return `MetricComputation` from map — Not Started
   d. Implement `GetWorkChannel()` (TODO at line 205): return the initialized work channel — Not Started
   e. Write tests for controller initialization: verify worker count derived from CPU cores, metric registration/lookup — Not Started

2. Implement worker pool lifecycle — Not Started
   a. Implement `StartWorkers()` (TODO at line 214): launch `numWorkers` goroutines that read from work channel and execute chunks — Not Started
   b. Implement `StopWorkers()` (TODO at line 221): close work channel, wait for worker goroutines via `workerWaitGroup` — Not Started
   c. Implement `SubmitChunk()` (TODO at line 231): send chunk to work channel with backpressure — Not Started
   d. Calculate optimal chunk size from cache size and tensor element size — Not Started
   e. Write tests for worker pool: verify correct goroutine count, chunk processing, graceful shutdown — Not Started

3. Verify parallel execution behavior — Not Started
   a. Create test CPL file with multiple metrics and large enough data to observe parallel execution — Not Started
   b. Write integration tests: verify generated code uses multiple goroutines — Not Started
   c. Verify generated code respects architecture configuration (correct worker count, chunk size) — Not Started
   d. Verify deterministic results — parallel execution must produce same output regardless of scheduling — Not Started

4. End-to-end verification — Not Started
   a. Create example IO spec with varying compute architecture configurations — Not Started
   b. Run parse → analyze → synthesize → compile → execute pipeline with different architecture configs — Not Started
   c. Verify parallel execution produces correct results matching sequential execution — Not Started
   d. Run the full CPL test suite to confirm no regressions — Not Started

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. This increment is synthesis-only — no parser or analyzer changes permitted
3. Generated code already embeds architecture config — this story implements the runtime controller
4. High-level tensor operation methods (`Add`, `Subtract`, `Multiply`, etc.) must internally handle chunking and parallel dispatch
5. Deterministic results are required — same input must always produce same output
6. Generated Go code must be compilable and runnable
7. Go functions that can fail must return `(T, error)`

## Intent

- **Task 1 Intent**: Implement the controller methods that are currently TODO placeholders — make the computation controller functional
- **Task 2 Intent**: Implement the worker pool lifecycle — start, stop, and submit work to parallel goroutines
- **Task 3 Intent**: Prove parallel execution works correctly and deterministically
- **Task 4 Intent**: Demonstrate configurable parallelism with end-to-end verification

## Acceptance Criteria

1. `ChronoconeComputationController` methods are fully implemented (no more TODOs)
2. Worker pool size is derived from `cpu_cores` with minimum of 1
3. Chunk size is optimized based on `cache_size`
4. Generated code uses work channel and worker pool for parallel tensor operations
5. Parallel execution produces deterministic results identical to sequential execution
6. All unit tests pass following Go testing best practices
7. No regressions in existing CPL test suite
