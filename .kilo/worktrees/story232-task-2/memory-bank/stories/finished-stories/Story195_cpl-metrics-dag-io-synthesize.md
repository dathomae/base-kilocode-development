# Story195: CPL Metrics DAG and IO Integration Synthesize

## Goal

Implement the Increment 8 synthesis phase for CPL metric multi-metric executables with full IO specification integration as part of [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md). This is the final increment, producing complete multi-metric Go executables. Output writing stubs already exist in [`generateOutputWritingCode()`](src/chronocone_planning_language/src/main/go/synthesis/synthesize.go:399). This story covers:

- Generate multi-metric Go executables with topological evaluation order — compute functions launched in dependency order (inputs before dependents).
- Register [`MetricComputation`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go) objects with correct consumer counts.
- Dead-code elimination: only generate code for reachable metrics.
- Full IO spec integration for output metrics' dimensions, celltype, grain, and compute architecture.
- Replace output writing stubs with actual tensor-to-parquet writing code that produces parquet files with specified dimensional/temporal configuration.

## References

- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [ERO005: Metrics Synthesis Phase — Increment 8](../specifications/design/cpl/go-synthesis-target/eros/ERO005_metrics-synthesis-phase.md)
- [Synthesis Design](../specifications/design/cpl/go-synthesis-target/synthesis.md)
- [Tensor Representation Design](../specifications/design/cpl/go-synthesis-target/tensor-representation.md)
- [IO Specifications Requirements](../specifications/requirements/cpl/io_specs.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story194_cpl-metrics-dag-io-analyze.md](Story194_cpl-metrics-dag-io-analyze.md) — analysis must produce validated DAG, topological order, pruned metrics, and fully validated IO spec (Not Started)
- All prior synthesis stories (183, 185, 186, 188, 190, 192) — their CIL visitors and code generation infrastructure must be in place

## Dependent Stories

None — this is the final story in Epic005.

## Tasks

1. Implement topological evaluation order generation — Completed
   a. Read topological order from analysis phase results (DAG walk) — Completed
   b. Generate `Compute()` function that launches metric computations in dependency order — Completed
   c. Write tests: verify generated code launches metrics in correct topological order — Completed
   d. Implement consumer count tracking: each `MetricComputation` knows how many downstream metrics depend on it — Completed

2. Implement dead-code elimination in synthesis — Completed
   a. Only generate compute functions for metrics marked as reachable by analysis — Completed
   b. Register only reachable `MetricComputation` objects in `Setup()` — Completed
   c. Write tests: verify unreachable metrics are excluded from generated code — Completed
   d. Ensure connector reading code is only generated for connectors needed by reachable metrics — Completed

3. Implement full IO spec output generation — Completed
   a. Generate output code that respects output metric dimensions from IO spec — Completed
   b. Generate output code that respects celltype from IO spec — Completed
   c. Generate output code that respects grain from IO spec — Completed
   d. Write tests: verify generated output parquet files match IO spec dimensional/temporal configuration — Completed

4. Implement actual tensor-to-parquet writing code — Completed
   a. Replace the stub in `generateOutputWritingCode()` with real parquet writing code — Completed
   b. Generate code to write computed metric tensors to parquet files using go-arrow — Completed
   c. Write output files following IO spec naming and path conventions — Completed
   d. Write tests: verify generated parquet output code is compilable and produces correct files — Completed

5. End-to-end verification — Completed
   a. Create comprehensive example CPL files with multiple metrics, dependencies, and full IO spec — Completed
   b. Create corresponding test parquet data — Completed
   c. Run complete pipeline: parse → analyze → synthesize → compile → execute → verify output parquet files — Completed
   d. Verify topological ordering produces correct results for dependency chains — Completed
   e. Verify dead-code elimination doesn't affect output correctness — Completed
   f. Run the full CPL test suite to confirm no regressions — Completed

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Generated code must follow topological evaluation order from DAG analysis
3. `MetricComputation` objects must have correct consumer counts for resource management
4. Only reachable metrics are included in generated code (dead-code elimination)
5. Output parquet files must match IO spec dimensional/temporal configuration
6. Generated Go code must be compilable and runnable
7. Go functions that can fail must return `(T, error)`
8. Output writing stubs exist — this story replaces them with real implementations

## Intent

- **Task 1 Intent**: Generate orchestration code that evaluates metrics in correct dependency order — inputs before dependents
- **Task 2 Intent**: Prune unreachable metrics from generated code — minimize executable size and computation
- **Task 3 Intent**: Generate output configuration that respects IO spec dimensional and temporal settings
- **Task 4 Intent**: Replace output writing stubs with real parquet writing code — produce final result files
- **Task 5 Intent**: Prove the complete, end-to-end multi-metric pipeline works correctly

## Acceptance Criteria

1. Generated code evaluates metrics in topological dependency order (inputs before dependents)
2. `MetricComputation` objects registered with correct consumer counts
3. Dead-code elimination: only reachable metrics generate code
4. Output parquet files match IO spec dimensional, celltype, grain, and temporal configuration
5. Multi-metric executable runs successfully and produces correct output
6. Full end-to-end test: CPL with multiple metrics, dependencies, and IO spec produces correct multi-file output
7. All unit tests pass following Go testing best practices
8. No regressions in existing CPL test suite
