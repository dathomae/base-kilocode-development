# Story270: Spreaders Increment 3 — Synthesis

## Goal

Extend the CPL synthesis pipeline so that spreaders referencing metrics (with implicit aggregators where needed) generate correct runtime code. When the Increment 3 analyzer populates `CILSpreadNode.ReferencedMetrics` with `CILMetricRefNode` and `CILAggregateNode` subtrees, the walker first walks these subtrees to emit code that computes the referenced metrics before the spread operation. Dynamic spreading matrices are constructed at runtime from the computed metric data (for example, `product_value / total_customers` as the per-country weight), replacing Increment 1's static pre-built matrices when metric references are present. The synthesis phase integrates metric reference computation, dynamic matrix construction, and preamble variable emission into the full code generation pipeline.

This story is Increment 3 of Epic008 (Spreaders).

## References

- [Plan: Story270](../plans/Story270_spreaders-increment3-synthesize.md)
- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md)
- [ERO008: Spreaders Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md)
- [ERO007: Aggregators Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md) — `ReduceDimension` call pattern reused for implicit aggregators

## Dependencies

- Story269 (spreaders-increment3-analyze) — Not Started — must produce `CILSpreadNode` with populated `ReferencedMetrics` slice (including `CILMetricRefNode` and `CILAggregateNode` for implicit aggregators)
- Story268 (spreaders-increment2-synthesize) — Not Started — CIL tree walking infrastructure for spreader-integrated metrics must exist
- Story261 (aggregators-increment4-synthesize) — Not Started — `ChronoconeComputationController`, topological scheduling, and implicit aggregator code generation pattern established
- [Story254_aggregators-increment1-matrix-mult.md](Story254_aggregators-increment1-matrix-mult.md) — Not Started — provides `tensorops.ReduceDimension` (for implicit aggregators) and `tensorops.ExpandDimension` (for the spread itself) synthesis primitives

## Dependent Stories

- Story272 (spreaders-increment4-synthesize) — Not Started — depends on the Increment 3 synthesis pipeline to produce spreader+aggregator chain computation code
- Story271 (spreaders-increment4-analyze) — Not Started — depends on the CIL tree structure established by Increment 3 analysis and synthesis

## Tasks

### Task Sizing

Tasks 1, 2, and 3 are not annotated `[Small]` — each requires reasoning about complex code generation infrastructure or coordinated multi-structure changes. Task 4 is annotated `[Small]` — it creates example files following existing templates and addresses a single narrow concern.

- **Task 1** (not small): Modifies a recursive CIL tree walker (`WalkCILSpreadNode`) to handle two new node types (`CILMetricRefNode`, `CILAggregateNode` wrapping `CILMetricRefNode`). Changes walker control flow by introducing a pre-walk phase for `ReferencedMetrics` subtrees before the `SourceCIL` walk. Requires reasoning about the CIL node hierarchy, walker scope management, and variable registration across the spread synthesis context.
- **Task 2** (not small): Creates a new file with new package-level code generation abstractions (`emitDynamicMatrixConstruction`). Introduces logic for constructing `AggregationMatrix` at runtime from computed metric data, with two distinct code paths (equal-proportion uniform weights versus metric-derived weighted distribution). Lays down new code infrastructure in the synthesis package.
- **Task 3** (not small): Integration work touching the code generator pipeline entry point. Wires Increment 3 synthesis into the existing metric computation code generator, verifies topological ordering respects metric references, and ensures spreading and aggregation matrix preamble variables are emitted. Coordinated changes across the code generator entry point and preamble emission logic.
- **Task 4** (`[Small]`): Creates an example directory (`examples/spreaders/metric_references/`) with CPL, IO spec, input XLSX, and expected JSON files. Follows existing example patterns and addresses a single concern (weighted distribution end-to-end verification). Purely additive.

### Parallel Execution

None — all tasks have sequential dependencies. Task 2 depends on Task 1 (dynamic matrix construction uses walked result variables and scope state produced by the walker). Task 3 depends on Tasks 1 and 2 (integration wires the completed walker logic and matrix construction into the code generator pipeline). Task 4 depends on all preceding tasks (examples verify the full pipeline). No file-disjoint task pairs exist.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

- Task 1 must complete first (the walker must emit `GetMetricComputation` and `ReduceDimension` calls before dynamic matrix construction can reference the resulting variables).
- Task 2 depends on Task 1 (dynamic matrix construction function receives walked result variables from the scope populated by the walker).
- Task 3 depends on Tasks 1 and 2 (integration wires the completed walker and matrix construction into the code generator entry point).
- Task 4 depends on all preceding tasks (examples verify the full synthesis pipeline produces correct weighted distribution output).

### Reintegration

None — no parallel groups exist. All tasks execute sequentially in the same worktree branch.

---

### Task 1: Implement ReferencedMetrics subtree walking — Not Started

Update `WalkCILSpreadNode` in the CIL tree walker to walk `ReferencedMetrics` subtrees before walking `SourceCIL`. When the walker encounters `CILMetricRefNode` nodes within the `ReferencedMetrics` slice, emit a `controller.GetMetricComputation(name).Tensor()` call. When it encounters `CILAggregateNode` nodes wrapping `CILMetricRefNode` (implicit aggregators), emit a `tensorops.ReduceDimension` call using the pre-existing aggregation matrix variable from the preamble. Register all walked result variables in the walk scope so Task 2 can reference them during dynamic matrix construction.

Follow Test-First Development: write tests for the walker extension before implementing the walker changes. Tests verify that the walker emits correct code for each of the three cases: direct metric reference, metric reference with implicit aggregation, and static spreading matrix fallback when `ReferencedMetrics` is empty.

a. Write unit tests in `src/main/go/synthesis/cil_spread_walker_test.go` for the updated `WalkCILSpreadNode`:
   - Test `walkSpreadNodeWithReferencedMetric`: construct a `CILSpreadNode` with a `ReferencedMetrics` slice containing a single `CILMetricRefNode` referencing a `[Country]` metric named `customers`. Verify the walker emits `controller.GetMetricComputation("customers").Tensor()` in the generated code. Verify the variable is registered in scope before the `SourceCIL` walk begins.
   - Test `walkSpreadNodeWithImplicitAggregator`: construct a `CILSpreadNode` with a `ReferencedMetrics` slice containing a `CILAggregateNode` wrapping a `CILMetricRefNode` for a metric requiring aggregation from `[Country]` to `[]`. Verify the walker emits a `ReduceDimension` call referencing the correct aggregation matrix variable name.
   - Test `walkSpreadNodeNoReferencedMetrics`: construct a `CILSpreadNode` with an empty `ReferencedMetrics` slice. Verify the walker does not emit any `GetMetricComputation` or `ReduceDimension` calls and proceeds directly to walking `SourceCIL` with static matrix fallback.
   - Tests must fail initially since the walker does not yet handle `ReferencedMetrics`. — Not Started

b. Update `WalkCILSpreadNode` in `src/main/go/synthesis/cil_spread_walker.go` to iterate `node.ReferencedMetrics` before walking `node.SourceCIL`:
   - For each `refNode` in `node.ReferencedMetrics`:
     - If `refNode` is a `CILMetricRefNode`: emit `refVar := scope.freshVar("refMetric")`, emit `controller.GetMetricComputation(name).Tensor()` assignment, and register the variable in the walk scope.
     - If `refNode` is a `CILAggregateNode` wrapping a `CILMetricRefNode`: walk the inner `CILMetricRefNode` to get the metric tensor reference, then emit a `ReduceDimension` call using the aggregation matrix variable name from the aggregator synthesis preamble (`aggregationMatrix_<index>_<fn>_<dim>` pattern per Story257/Story261). Register the resulting scalar tensor in scope. — Not Started

c. Verify walker tests pass: `go test ./src/main/go/synthesis/... -run "TestWalkCILSpreadNode|TestWalkSpreadNode" -count=1`. All tests from step (a) must pass. — Not Started

**Intent**: Enable the CIL tree walker to process `ReferencedMetrics` subtrees before the spread operation, emitting `GetMetricComputation` calls for direct metric references and `ReduceDimension` calls for metric references requiring implicit aggregation. The walked result variables are registered in scope so downstream code (Task 2) can construct dynamic spreading matrices from the computed metric data.

**Acceptance Criteria**:
- `WalkCILSpreadNode` walks `ReferencedMetrics` before `SourceCIL`
- `CILMetricRefNode` emits `controller.GetMetricComputation(name).Tensor()` call with the correct metric name
- `CILAggregateNode` wrapping `CILMetricRefNode` emits `tensorops.ReduceDimension` call referencing the correct aggregation matrix variable
- Empty `ReferencedMetrics` slice produces no extra code emission (static spreader behavior preserved)
- Walked result variables are registered in scope before `SourceCIL` walking begins
- All walker tests pass

---

### Task 2: Implement dynamic spreading matrix code generation — Not Started

Create `dynamic_spread_matrix.go` with the `emitDynamicMatrixConstruction` function that generates runtime Go code constructing an `AggregationMatrix` from computed metric data. When `ReferencedMetrics` is non-nil, the walker calls this function to emit code that builds a dynamic matrix with weights derived from the referenced metric values (for example, `weight = product_value / customers_scalar` for weighted distribution). When no metric references exist, static pre-built matrices from Increment 1 are used instead.

Follow Test-First Development: write tests that verify the generated code constructs matrices with correct dimensions, weight computations, and column labels before implementing the code generation function.

a. Write unit tests in `src/main/go/synthesis/dynamic_spread_matrix_test.go`:
   - Test `emitDynamicMatrixWeighted`: verify that `emitDynamicMatrixConstruction` generates code constructing an `AggregationMatrix` where `Data[i][j]` is computed as `sourceValue / refScalarValue`. Verify the generated code includes `Rows`, `Columns`, `ColLabels`, and `Data` initialization.
   - Test `emitDynamicMatrixEqualProportion`: verify that when no metric reference provides weights (equal-proportion case), the generated code produces uniform `1/N` weights where `N` is the number of target dimension values.
   - Test `emitDynamicMatrixColLabels`: verify the generated code populates `ColLabels` from the target dimension's category values.
   - Test `emitDynamicMatrixShape`: verify the generated matrix has `Rows: 1` (expands during iteration) and `Columns` matching the target dimension cardinality.
   - Tests must fail initially since `dynamic_spread_matrix.go` does not exist. — Not Started

b. Create `src/main/go/synthesis/dynamic_spread_matrix.go` with the `emitDynamicMatrixConstruction` function:
   - Accept parameters: `matrixVar` (variable name for the generated matrix), `sourceVar` (variable name for the source tensor), and the `CILSpreadNode` containing the target dimension and referenced metric information.
   - Generate Go code that:
     - Initializes an `AggregationMatrix` with `Rows: 1`, `Columns` set to the target dimension cardinality, and `ColLabels` populated from dimension category values.
     - For weighted distribution (referenced metrics present): emits a loop computing `weight = sourceValue / refScalarValue` for each source cell and target dimension value, assigning to `matrix.Data[0][j]`.
     - For equal-proportion (no metric reference or uniform weights): emits `matrix.Data[0][j] = 1.0 / float64(numTargetValues)` for each target dimension value.
   - Handle the distinction between weighted and equal-proportion cases based on the presence and type of referenced metrics in the `CILSpreadNode`. — Not Started

c. Verify dynamic matrix tests pass: `go test ./src/main/go/synthesis/... -run "TestDynamicSpreadMatrix|TestEmitDynamicMatrix" -count=1`. All tests from step (a) must pass. — Not Started

**Intent**: Provide a code generation function that emits runtime Go code to construct a dynamic `AggregationMatrix` from computed metric data. This replaces Increment 1's static pre-built matrices when spreaders reference metrics. The generated code computes weights (for example, per-country weight = product_value / total_customers) and builds the matrix structure at runtime.

**Acceptance Criteria**:
- `emitDynamicMatrixConstruction` generates valid Go code constructing an `AggregationMatrix` struct
- The generated `AggregationMatrix` has correct `Rows`, `Columns`, and `ColLabels`
- Weighted distribution case computes weights from source and referenced metric values
- Equal-proportion case uses uniform `1/N` weights
- Generated code compiles without syntax errors
- All dynamic matrix tests pass

---

### Task 3: Integrate metric reference synthesis into full pipeline — Not Started

Wire the Increment 3 synthesis components (ReferencedMetrics subtree walking from Task 1 and dynamic matrix construction from Task 2) into the metric computation code generator. Ensure the topological computation order naturally respects metric references (the analysis-phase dependency DAG already includes metric references as edges, so the scheduler requires no additional changes). Ensure all spreading and aggregation matrix preamble variables needed by the generated code are emitted before any computation functions.

Follow Test-First Development: write integration tests that exercise the full synthesis pipeline with a weighted distribution scenario before implementing the integration wiring.

a. Write integration tests in `src/main/go/synthesis/spreader_metric_ref_synth_test.go` (create the file if it does not exist):
   - Test `synthesizeWeightedDistributionFullPipeline`: construct a minimal CIL tree with a spreader that references a `[Country]` metric with an implicit aggregator (`from [Country] to [] = 1/sum(Country)`). Run the full synthesis pipeline. Verify the generated output contains, in order: (1) referenced metric `GetMetricComputation` call, (2) `ReduceDimension` call for implicit aggregation, (3) dynamic matrix construction code, (4) `ExpandDimension` call for the spread. Verify the generated code compiles with `go build ./src/main/go/...`.
   - Test `synthesizeDirectMetricRefNoImplicitAgg`: construct a CIL tree where the referenced metric dimensions already match the spreader's needs. Verify no `ReduceDimension` call is emitted — only `GetMetricComputation`.
   - Test `topologicalOrderRespectsReferences`: verify that in the generated computation function list, the referenced metric computation function appears before the spreader metric computation function.
   - Test `preambleVariablesEmitted`: verify that the generated preamble includes aggregation matrix variables needed by `ReduceDimension` calls in the computation functions.
   - Tests must fail initially since integration wiring has not been implemented. — Not Started

b. Wire Increment 3 synthesis into the metric computation code generator in `src/main/go/synthesis/cil_spread_walker.go` (and any related code generator entry-point files):
   - After walking `ReferencedMetrics` (Task 1) and before the spread operation, call `emitDynamicMatrixConstruction` from Task 2 when `node.ReferencedMetrics` is non-empty.
   - For the static fallback path (empty `ReferencedMetrics`), preserve the existing Increment 1 behavior using pre-built spreading matrix variables.
   - Ensure the walker emits `ExpandDimension` using either the dynamically constructed matrix variable or the static pre-built matrix variable, depending on which path was taken. — Not Started

c. Verify integration tests pass: `go test ./src/main/go/synthesis/... -run "TestSynthesizeWeighted|TestSynthesizeDirectMetricRef|TestTopologicalOrder|TestPreambleVariables" -count=1`. All tests from step (a) must pass, and the generated code must compile. — Not Started

**Intent**: Integrate the Increment 3 synthesis components into the full code generation pipeline so that spreaders referencing metrics produce correct, compilable Go computation code. The integration preserves Increment 1's static matrix behavior as a fallback when no metric references are present. Topological ordering requires no additional code — the analysis-phase DAG already includes metric references as dependency edges.

**Acceptance Criteria**:
- Full synthesis pipeline produces compilable Go code for weighted distribution spreaders
- Referenced metric `GetMetricComputation` calls appear before spreader computation in generated code
- `ReduceDimension` calls appear only when implicit aggregation is needed (dimension mismatch)
- Dynamic matrix construction code appears between metric computation and `ExpandDimension` call
- Static matrix fallback preserved when `ReferencedMetrics` is empty
- Preamble includes all aggregation matrix variables needed by computation functions
- All integration tests pass
- Generated code compiles with `go build ./src/main/go/...`

---

### Task 4: Create end-to-end weighted distribution examples — Not Started [Small]

Create example files in `examples/spreaders/metric_references/` that provide end-to-end test scenarios for weighted distribution spreaders. Each example includes a CPL definition, an IO specification, input data in XLSX format, and expected output in JSON format. The examples follow the existing pattern used by other `with_data/` example directories.

a. Create the directory structure under the appropriate `with_data` root (within `src/chronocone_planning_language/`):
   ```
   examples/spreaders/metric_references/
   ├── spreaders_metric_refs.cpl
   ├── spreaders_metric_refs.io.yaml
   ├── spreaders_metric_refs_input.xlsx
   ├── spreaders_metric_refs.expected.json
   ├── spreaders_direct_metric_ref.cpl
   ├── spreaders_direct_metric_ref.io.yaml
   ├── spreaders_direct_metric_ref_input.xlsx
   ├── spreaders_direct_metric_ref.expected.json
   ├── spreaders_equal_proportion_metric.cpl
   ├── spreaders_equal_proportion_metric.io.yaml
   ├── spreaders_equal_proportion_metric_input.xlsx
   └── spreaders_equal_proportion_metric.expected.json
   ```
   — Not Started

b. Example 1 — Weighted distribution with implicit aggregator (`spreaders_metric_refs`):
   - Define a `[Country]` connector metric `customers` with values `USA=1000, Canada=500`.
   - Define a `[Product]` metric `source_product` with values `widget=300, gadget=600`.
   - Define a spreader `from [Product] to [Country, Product] = Product / customers` requiring implicit aggregation of `customers` (`from [Country] to [] = 1/sum(Country)`) to produce a scalar `customersScalar = 1500`.
   - Expected output: `widget` at `USA=60` (300/1500×300), `widget` at `Canada=30` (300/1500×150), `gadget` at `USA=80` (600/1500×200), `gadget` at `Canada=40` (600/1500×100). Verify with `expected.json`. — Not Started

c. Example 2 — Direct metric reference with matching dimensions (`spreaders_direct_metric_ref`):
   - Define a `[Country, Product]` metric `weight_factors` where the dimensions already match the spreader target.
   - Define a spreader referencing `weight_factors` directly — no implicit aggregation needed because dimensions match.
   - Verify that the generated output uses `GetMetricComputation` without `ReduceDimension`, and produces correct spread results. — Not Started

d. Example 3 — Metric reference with equal-proportion fallback (`spreaders_equal_proportion_metric`):
   - Define a `[Country]` metric that contributes to but does not dominate the spread weight calculation (the spreader expression divides by the metric value, producing a proportional distribution).
   - Verify the spreader uses dynamic matrix construction (not static pre-built) and produces correct proportional results matching `expected.json`. — Not Started

e. Verify all examples produce output matching their respective `.expected.json` files using the project's XLSX verification test infrastructure. — Not Started

**Intent**: Provide end-to-end examples that verify the full synthesis pipeline produces correct weighted distribution output across three scenarios: implicit aggregation, direct metric reference with matching dimensions, and equal-proportion fallback. These examples serve as regression tests for the Increment 3 synthesis pipeline and documentation for the weighted distribution spreader feature.

**Acceptance Criteria**:
- All three example directories created with CPL, IO spec, input XLSX, and expected JSON files
- Example 1 produces correct weighted distribution with implicit aggregation
- Example 2 produces correct distribution without `ReduceDimension` (direct reference, matching dims)
- Example 3 produces correct proportional distribution with dynamic matrix construction
- All examples pass XLSX verification against their expected output

## Constraints

- `ReduceDimension` calls for implicit aggregators must reuse the aggregation matrix preamble variables emitted by the aggregator synthesis stories (Story257/Story261). The spreader synthesizer references these variables by name rather than re-emitting them.
- The topological scheduling requires no additional synthesis code — the analysis-phase dependency DAG already includes metric references as dependency edges, so the existing `ChronoconeComputationController` scheduler produces correct computation order.
- Static spreading matrix fallback (Increment 1 behavior) must be preserved when `ReferencedMetrics` is empty — existing spreaders without metric references must continue to generate code using pre-built matrix variables.
- Error propagation must be maintained: metric computation failures, implicit aggregation failures, and spread failures must all produce meaningful error messages in generated code.

## Notes

- The `CILSpreadNode.ReferencedMetrics` field is populated by the Increment 3 analyzer (Story269). If this field is `nil` or empty, the walker falls through to Increment 1 static matrix behavior.
- The `CILAggregateNode` nodes wrapping `CILMetricRefNode` represent implicit aggregators — the analyzer inserts these when a referenced metric's dimensions do not match what the spreader expression expects. The synthesizer emits `ReduceDimension` calls to perform these aggregations at runtime.
- Aggregation matrix variable names follow the pattern `aggregationMatrix_<index>_<fn>_<dim>` established by Story257/Story261. The spreader synthesizer uses the same pattern to reference these variables in generated `ReduceDimension` calls.
- Dynamic spreading matrix construction differs from Increment 1's static matrices: static matrices are pre-built at compile time and stored as package-level variables; dynamic matrices are constructed at runtime in the computation function body using metric data values.
