# Story194: CPL Metrics DAG and IO Integration Analyze

## Goal

Implement the Increment 8 analysis phase for CPL metric dependency DAG and IO specification integration as part of [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md). **This is an analysis-only increment — no parser changes.** This story covers:

- Build complete metric dependency DAG in second pass after all definitions are assembled.
- Walk DAG for topological evaluation order.
- Detect circular dependencies (excluding time-shifted self-references).
- Implement dead-code elimination — only include metrics reachable from IO spec output metrics.
- Full IO specification integration: validate output metrics' dimensions, celltype, grain, and connector-to-data-source mappings.

## References

- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [ERO005: Metrics Analysis Phase — Increment 8](../specifications/design/cpl/go-analysis/eros/ERO005_metrics-analysis-phase.md)
- [Metrics Design](../specifications/design/cpl/metrics-design.md)
- [IO Specifications Requirements](../specifications/requirements/cpl/io_specs.md)
- [CPL Architecture — Intermediate Language](../specifications/architecture/cpl/analysis/intermediate_language.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story182_cpl-metrics-basic-parse-analyze.md](finished-stories/Story182_cpl-metrics-basic-parse-analyze.md) — Increment 1 Parse+Analyze (Done)
- [Story184_cpl-metrics-arithmetic-parse-analyze.md](finished-stories/Story184_cpl-metrics-arithmetic-parse-analyze.md) — Increment 2 Parse+Analyze (Done)
- [Story187_cpl-metrics-conditionals-parse-analyze.md](finished-stories/Story187_cpl-metrics-conditionals-parse-analyze.md) — Increment 4 Parse+Analyze (Done)
- [Story189_cpl-metrics-aliases-where-parse-analyze.md](finished-stories/Story189_cpl-metrics-aliases-where-parse-analyze.md) — Increment 5 Parse+Analyze (Done)
- [Story191_cpl-metrics-time-shift-parse-analyze.md](finished-stories/Story191_cpl-metrics-time-shift-parse-analyze.md) — Increment 6 Parse+Analyze (Done)
- [Story193_cpl-metrics-short-form-parse-analyze.md](finished-stories/Story193_cpl-metrics-short-form-parse-analyze.md) — Increment 7 Parse+Analyze (Done)

## Dependent Stories

- [Story195_cpl-metrics-dag-io-synthesize.md](Story195_cpl-metrics-dag-io-synthesize.md) — Increment 8 Synthesis (Not Started)

## Tasks

1. Implement metric dependency DAG construction — Completed
   a. Define DAG data structures: graph with nodes (metrics) and edges (dependencies) — Completed
   b. Implement DAG builder: second pass after all `AnalysisMetricDefinition` entries are assembled — Completed
   c. Walk each metric's CIL tree to discover `CILInputRef` nodes that reference other metrics — Completed
   d. Add edges for each inter-metric reference — Completed
   e. Write tests: verify DAG correctly captures metric dependencies — Completed

2. Implement topological sort and cycle detection — Completed
   a. Implement topological sort algorithm on the DAG — Completed
   b. Detect circular dependencies: cycle in DAG that is not a time-shifted self-reference — Completed
   c. Allow time-shifted self-references: these are not cycles — Completed
   d. Write tests: topological order, cycle detection, time-shifted self-reference exemption — Completed

3. Implement dead-code elimination — Completed
   a. Identify metrics reachable from IO spec output metrics by traversing DAG in reverse — Completed
   b. Mark unreachable metrics as dead code — exclude from synthesis — Completed
   c. Write tests: verify only reachable metrics pass through to synthesis — Completed

4. Implement full IO specification validation — Completed
   a. Validate output metrics' dimensions against connector declarations — Completed
   b. Validate output metrics' celltype against measure declarations — Completed
   c. Validate output metrics' grain against connector grain declarations — Completed
   d. Validate connector-to-data-source mappings (parquet file paths) — Completed
   e. Write tests: IO spec validation rejects dimension/celltype/grain mismatches — Completed

5. Integration verification — Completed
   a. Create example CPL files with multiple metrics and dependencies — Completed
   b. Create IO specs with output metric selection that triggers dead-code elimination — Completed
   c. Verify analysis produces correct DAG, topological order, and prunes dead code — Completed
   d. Test edge cases: empty IO spec (all metrics dead), multi-level dependency chains, circular dependencies — Completed
   e. Run the full CPL test suite to confirm no regressions — Completed

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. This increment is analysis-only — no parser changes permitted
3. DAG construction is a second pass after all metric definitions are assembled
4. Time-shifted self-references are not cycles — they are valid time-series dependencies
5. Dead-code elimination: only metrics reachable from IO spec output metrics are included
6. Generated Go tests must follow the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Intent

- **Task 1 Intent**: Build the dependency graph between metrics — understand which metrics depend on which others
- **Task 2 Intent**: Determine correct evaluation order via topological sort and detect invalid circular dependencies
- **Task 3 Intent**: Prune metrics not needed for the requested output — reduce computation and code size
- **Task 4 Intent**: Validate IO specification against declared entities — ensure consistency before synthesis
- **Task 5 Intent**: Prove DAG and IO integration analysis works correctly

## Acceptance Criteria

1. DAG correctly captures all inter-metric dependencies from CIL tree analysis
2. Topological sort produces valid evaluation order
3. Circular dependencies (excluding time-shifted self-references) are detected and reported
4. Dead-code elimination removes metrics unreachable from IO spec outputs
5. IO spec validation: dimension, celltype, grain, and data source mappings verified
6. All unit tests pass following Go testing best practices
7. No regressions in existing CPL test suite
