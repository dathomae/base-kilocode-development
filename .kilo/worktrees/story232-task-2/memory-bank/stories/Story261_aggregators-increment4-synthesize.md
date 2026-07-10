# Story261: Aggregators Increment 4 — Synthesize (Epic007)

## Goal

Extend synthesis for aggregator chains: generate runtime aggregation matrix calls with lookup tensor inputs, unit-correct aggregation codegen, and output dimension adjustment in preamble. This story (Increment 4 of Epic007) connects the analysis-phase aggregator injection (Story260) with the synthesis-phase code generation (Stories 257, 259). No new CIL walk methods are required — the walk methods added in Stories 257 and 259 already handle all aggregate node types. The synthesis phase naturally walks the modified CIL trees produced by the analysis phase, emitting `ReduceDimension` calls as part of metric computation. This story verifies and hardens that integration through tests and a concrete end-to-end example.

## References

- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md)
- [ERO007: Aggregators Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md)
- [Design: CPL Design Document](../specifications/design/cpl/design.md)
- [Plan: Story261 Aggregators Increment 4 Synthesize](../plans/Story261_aggregators-increment4-synthesize.md)
- [Source root: src/chronocone_planning_language](../../src/chronocone_planning_language/)

## Dependencies

- [Story260_aggregators-increment4-analyze.md](Story260_aggregators-increment4-analyze.md) — produces CIL trees with inserted aggregate node chains from `resolveMetricDimensionality`
- [Story259_aggregators-increment3-synthesize.md](Story259_aggregators-increment3-synthesize.md) — provides walk methods for all five typed CIL node types
- [Story257_aggregators-increment2-synthesize.md](Story257_aggregators-increment2-synthesize.md) — provides matrix emission infrastructure (`GenerateAggregationMatrix`) and preamble integration
- [Story254_aggregators-increment1-matrix-mult.md](Story254_aggregators-increment1-matrix-mult.md) — provides `tensorops.ReduceDimension` and `tensorops.AggregationMatrix` types

## Dependent Stories

None — this is the final story of Increment 4. Subsequent aggregator increments (Increment 5) will depend on the verified end-to-end pipeline.

## Tasks

### Task Sizing

Task 2 is annotated `[Small]` — it modifies at most 2 files (one source review + one test addition), is additive, and addresses a single narrow concern (scheduling order verification). Task 1 is not `[Small]` — it requires understanding existing metric computation code generation patterns, the CIL tree structure with aggregate nodes inserted by the analysis phase, and potentially modifying `cil_transformer_metric.go` based on test results. Task 3 is not `[Small]` — it creates a complete with_data example with multiple coordinated file types (CPL, IO spec, parquet, expected output) that must be internally consistent (consistent with Story257 Task 4 precedent). Task 4 is not `[Small]` — it writes end-to-end synthesis integration tests that require constructing AnalysisDefinitions with aggregate-wrapped inputs, synthesizing, verifying generated code compilation and correctness, and testing error propagation across the full synthesis pipeline.

### Parallel Execution

- **Group A** (Tasks 3, 4): Task 3 creates files in `examples/with_data/aggregators_integration/`; Task 4 modifies `src/main/go/synthesis/synthesize_test.go`. Fully disjoint file sets, different packages. Neither task depends on the other's output — Task 3 verifies through example pipeline execution, Task 4 verifies through programmatic Go tests.

### Execution Order

```
Task 1 → Task 2 → [Task 3 ‖ Task 4]
```

Task 1 must complete first: metric computation code generation must be verified before scheduling can be meaningfully tested. Task 2 depends on Task 1 because scheduling tests use the same metric code generation patterns that Task 1 validates. Tasks 3 and 4 are parallel: the with_data example and the programmatic integration test are independent verification paths.

### Reintegration

**Group A (Tasks 3, 4):**
- Merge order: any order (fully disjoint directories, no shared infrastructure)
- Integration tests: `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` — verify Task 4's end-to-end synthesis tests pass after merge. The example directory's test is verified independently during Task 3.
- Watch for conflicts in: none expected (disjoint file paths: `examples/` vs `src/main/go/synthesis/`)

---

### Task 1: Verify and test metric computation with aggregate nodes in CIL trees — Not Started

**Architect background**: This task verifies that the existing metric computation code generator (`cil_transformer_metric.go`) correctly handles CIL trees containing aggregate node chains inserted by the analysis phase. No new walk methods are needed — the walk methods from Stories 257 and 259 already handle all aggregate node types. The task writes tests that construct CIL trees with aggregate-wrapped inputs and verifies the generated code structure, then makes any minor adjustments needed (e.g., ensuring the `tensorops` package reference is imported).

1. a. Review `src/main/go/cil/cil_transformer_metric.go` to confirm that existing metric computation code generation correctly handles CIL trees containing aggregate nodes (e.g., `CILAggregateSumNode`) inserted by the analysis phase. The walker must walk through aggregate nodes to reach input references, emit `ReduceDimension` calls before cell-wise operations, and register aggregated intermediate tensors for subsequent use. — Not Started
   b. Write a test in `src/main/go/cil/cil_transformer_aggregator_test.go`: construct a CIL tree where a metric input reference is wrapped in a `CILAggregateSumNode` → run through `cil_transformer_metric` → verify the generated Go code contains a `tensorops.ReduceDimension` call with the correct variable names, dimension indices, and aggregation matrix parameter before the cell-wise computation body. — Not Started
   c. Write a test: verify the complete generated code structure for a metric with an aggregate-wrapped input — the structure must be: input metric loading → aggregate application (`ReduceDimension` with error handling) → cell-wise computation → result return. Confirm that the error handling for `ReduceDimension` uses descriptive `fmt.Errorf` wrapping. — Not Started
   d. If adjustments are needed to `src/main/go/cil/cil_transformer_metric.go` based on test failures, make them. Expected adjustments are minor (e.g., ensuring the `tensorops` package import is present in the generated code's import block). — Not Started
   e. Run `go test ./src/main/go/cil/... -run Aggregator -count=1` and verify all aggregator-related tests pass. — Not Started

**Intent**: Confirm that the metric computation code generator produces correct `ReduceDimension` calls when CIL trees contain aggregate nodes. This task validates that the synthesis phase naturally handles the modified CIL tree structure produced by the analysis phase in Story260, without requiring new walk methods. The tests serve as regression protection for this critical integration point.

**Acceptance Criteria**:
- A CIL tree with `CILAggregateSumNode` wrapping a metric input produces generated code containing `tensorops.ReduceDimension` before cell-wise operations
- The generated code's error handling for `ReduceDimension` wraps errors with descriptive messages
- The generated code structure follows the correct order: input loading → aggregation → cell-wise computation → return
- Any necessary minor adjustments to `cil_transformer_metric.go` are made and verified
- All aggregator tests pass with `go test ./src/main/go/cil/... -run Aggregator`

---

### Task 2: Verify topological scheduling with aggregated dependencies — Not Started [Small]

**Architect background**: This task verifies that the topological scheduler in `synthesize.go` correctly orders metrics whose inputs have aggregate-wrapped dependencies. The existing topological sort should naturally handle this since aggregate nodes wrap CIL references — the metric still depends on the input metric conceptually. The task writes a test to confirm this and makes minor adjustments if needed.

2. a. Review `src/main/go/synthesis/synthesize.go` scheduling logic to confirm that metrics with aggregate-wrapped inputs are scheduled after their input metrics. The topological sort dependency tracking should not be affected by aggregate node insertion — the dependency is on the metric, not on the aggregate node. — Not Started
   b. Write a test in `src/main/go/synthesis/synthesize_test.go`: construct a two-metric DAG where Metric B has an aggregate-wrapped input referencing Metric A's output → run scheduling → verify that Metric B is scheduled after Metric A in the generated main function's evaluation order. — Not Started
   c. If adjustments to `src/main/go/synthesis/synthesize.go` are needed (e.g., if the dependency tracker needs to unwrap aggregate nodes to find the true metric dependency), make them. — Not Started
   d. Run `go test ./src/main/go/synthesis/... -run "Schedule|DAG" -count=1` and verify all scheduling tests pass. — Not Started

**Intent**: Confirm that the topological scheduler correctly orders metrics with aggregate-wrapped dependencies. The scheduler must produce an evaluation order where aggregated intermediate tensors are computed before their consumers, ensuring that generated programs execute without nil-reference errors at runtime.

**Acceptance Criteria**:
- A metric whose input is another metric wrapped with aggregate nodes is scheduled after its input metric
- The generated main function's evaluation order respects the dependency DAG with aggregate-wrapped edges
- Any necessary adjustments to the scheduler are minor and verified
- All scheduling tests pass with `go test ./src/main/go/synthesis/...`

---

### Task 3: Create with_data aggregator integration example — Not Started

**Architect background**: This task creates a concrete end-to-end example validating the full aggregator pipeline: CPL → parse → analyze → CIL → synthesize → execute → verify. The example uses known data and hand-computed ground truth to confirm that the synthesis phase produces correct aggregation results when aggregate nodes are present in CIL trees.

3. a. Create the `examples/with_data/aggregators_integration/` directory structure. — Not Started
   b. Write a CPL file defining the concepts (`Country`, `Product`), a connector (`source_data`) producing `[Country, Product]` data, an aggregator declaration (`from [Country, Product] to [Product] = sum(Country)`), and a metric (`revenue_by_product`) that takes the connector as input and applies implicit aggregation to produce `[Product]` output. — Not Started
   c. Write an IO specification file: input parquet connector with `[Country, Product]` dimensions, output metric spec with `[Product]` dimensions. — Not Started
   d. Generate parquet test data with known numeric values: Country={USA, Canada} × Product={widget, gadget} with specific per-cell values that produce unambiguous sums. — Not Started
   e. Create a `.expected.xlsx` file with hand-computed ground-truth aggregated values: each product's value is the sum across all countries. — Not Started
   f. Run the full pipeline on the example (parse → analyze → CIL transform → synthesize → execute) and verify that the output matches the expected aggregated values in `.expected.xlsx`. — Not Started

**Intent**: This example provides a concrete, reproducible validation of the Increment 4 synthesis pipeline. It confirms that when the analysis phase inserts aggregate nodes into CIL trees (Story260), the synthesis phase (this story) produces correct executable code that performs the aggregation and computes the metric. The example serves as both a development verification tool and a template for future aggregator scenarios.

**Acceptance Criteria**:
- CPL file parses and analyzes without errors, with aggregate nodes inserted into the metric's CIL tree
- Synthesized program compiles and executes without runtime errors
- Execution output matches the hand-computed ground truth in `.expected.xlsx`
- Full pipeline runs end-to-end on the example

---

### Task 4: End-to-end synthesis integration test with aggregators — Not Started

**Architect background**: This task writes programmatic integration tests that exercise the full synthesis pipeline with aggregate nodes in CIL trees. Unlike Task 3's file-based example, these tests construct AnalysisDefinitions directly in Go code and verify the generated code's structure, compilation, and runtime behavior. Error propagation from `ReduceDimension` is explicitly tested.

4. a. Write a synthesis integration test in `src/main/go/synthesis/synthesize_test.go`: construct a full `AnalysisDefinitions` with a metric that has an aggregate-wrapped input → run synthesis → verify the generated Go code compiles with `go build` and produces correct results when executed with test input data. The test should validate that the generated function structure matches the expected pattern: input loading → aggregate application → cell-wise computation → return. — Not Started
   b. Write a test for error propagation: construct an `AnalysisDefinitions` where a `ReduceDimension` call would fail at runtime (e.g., mismatched matrix dimensions) → verify that the generated code's error handling produces a descriptive error message that includes the metric name and the aggregation step that failed. — Not Started
   c. Run `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` and verify all tests pass, including both the new integration tests and all pre-existing synthesis tests. — Not Started

**Intent**: Programmatically validate the synthesis pipeline's handling of aggregate nodes through Go-level integration tests. These tests complement the file-based example in Task 3 by testing edge cases (error propagation, code structure verification) that are harder to exercise through file-based examples. Together, Tasks 3 and 4 provide confidence that the synthesis phase correctly handles the modified CIL trees produced by the analysis phase.

**Acceptance Criteria**:
- An `AnalysisDefinitions` with aggregate-wrapped inputs produces compilable Go code
- The generated code, when executed, produces correct aggregated results matching expected values
- Generated error handling for `ReduceDimension` failures produces descriptive messages including the metric name
- All synthesis tests pass with `go test ./src/main/go/synthesis/... -count=1 -timeout 60s`

---

## Test-First Development

This story follows the Logical TDD Lifecycle. For each coding task, tests are written before (or alongside) the implementation code. The test-and-implement cycle is integrated within each task, not broken out as separate tasks. See `memory-bank/concepts.md` for details on the Logical TDD Lifecycle.

## Constraints

- **No new walk methods**: Increment 4 does not add new CIL walk methods. The walk methods added in Stories 257 and 259 already handle all aggregate node types. This story verifies and hardens the integration; any code changes are minor adjustments, not new infrastructure.
- **Existing aggregator infrastructure**: This story depends on the typed CIL node types (Story258), the Aggregators map (Story256), and the analysis-phase aggregate node insertion (Story260) already being in place.
- **Single-sequence limitation**: In Increment 4, only single unambiguous aggregator sequences are supported. Multi-sequence enumeration is deferred to Increment 5. The synthesis phase does not need to handle multi-step aggregator chains.
- **Generated code compiles**: All generated Go code must compile under `source_root` (`src/chronocone_planning_language`).
- **Error message clarity**: `ReduceDimension` errors in generated code must be wrapped with descriptive messages that include the metric name and aggregation step, so users can diagnose failures without reading generated source.
- **File paths**: All file paths in tasks are relative to `source_root` (`src/chronocone_planning_language`). Go test commands use fully-qualified package paths from `source_root`.

## Intent

This story completes Increment 4 by verifying that the synthesis phase correctly handles the CIL tree structure produced by the analysis phase. When Story260 inserted aggregate node chains into metric CIL trees during the second analysis pass, no new synthesis code was required — the walk methods from Stories 257 and 259 already handle aggregate node types. However, this integration point is critical: if the metric computation code generator fails to walk through aggregate nodes, or if the topological scheduler misorders aggregated dependencies, the generated programs will fail at runtime. This story validates the integration through tests, verifies that the scheduling order respects aggregated dependencies, and provides concrete end-to-end evidence (file-based example + programmatic tests) that the full aggregator pipeline works correctly.

## Acceptance Criteria

1. Metric computation code generation handles CIL trees containing aggregate node chains: emits `ReduceDimension` calls before cell-wise operations
2. Generated metric functions properly load inputs, apply aggregators, compute cell-wise, and return results
3. Error propagation from `ReduceDimension` is properly wired with descriptive messages
4. Topological scheduling ensures aggregated metrics are computed before their consumers
5. `examples/with_data/aggregators_integration/` example: CPL + IO spec + parquet → synthesized executable produces correct aggregated results matching expected output
6. End-to-end synthesis test with aggregate nodes in CIL tree produces compilable code with correct results
7. All tests pass with `go test ./src/main/go/cil/... ./src/main/go/synthesis/... -count=1 -timeout 60s`
