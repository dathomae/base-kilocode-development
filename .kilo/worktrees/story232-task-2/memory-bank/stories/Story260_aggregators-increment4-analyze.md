# Story260: Aggregators Increment 4 — Analysis Phase

## Goal

Implement aggregator lookup resolution in aggregator expressions, validate output unit compatibility for aggregator-lookup chains, aggregator output dimension derivation. This story (Increment 4 of Epic007) enables automatic aggregator integration into the metric dependency DAG: the analysis phase detects dimension mismatches between metric inputs and outputs after the full DAG is assembled, resolves single unambiguous aggregator sequences from the Aggregators map, and inserts `CILAggregateNode` chains into the metric CIL trees. In this increment, only single unambiguous sequences are supported — multiple possible aggregator paths produce a compile-time error.

## References

- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md)
- [ERO007: Aggregators Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO007_aggregators-analysis-phase.md)
- [Design: CPL Design Document](../specifications/design/cpl/design.md)

## Dependencies

- [Story256_aggregators-increment2-analyze.md](Story256_aggregators-increment2-analyze.md) — provides the populated `Aggregators` map with `AggregatorInfo` entries (needed for `findSingleAggregatorSequence` lookups)
- [Story258_aggregators-increment3-analyze.md](Story258_aggregators-increment3-analyze.md) — provides the full `FunctionRegistry` and typed CIL node types (`CILAggregateSumNode`, `CILAggregateCountNode`, `CILAggregateSizeNode`, `CILAggregateInverseSumNode`, `CILAggregateInverseCountNode`) needed for `buildAggregateNode`
- Epic005 (metrics) — provides the metric DAG analysis infrastructure (`AnalysisMetricDefinition`, input resolution, IO spec dimensions) needed for the second analysis pass

## Dependent Stories

- [Story261_aggregators-increment4-synthesize.md](Story261_aggregators-increment4-synthesize.md) — Not Started — depends on the inserted `CILAggregateNode` chains in the CIL trees to produce computation code

## Tasks

### Task Sizing

Tasks 1, 2, 3a, and 4 are annotated `[Small]` — each modifies at most 2 files (one source + one test), is additive rather than restructuring, and addresses a single narrow concern. Task 3b is not annotated `[Small]` — it modifies `cpl_analyzer.go` (the core analysis pipeline) and requires reasoning about the full analyzer execution flow across multiple interacting structures. Task 3 was decomposed from the original plan (which had 6 subtask steps with a self-contained preamble) into Task 3a (implement the methods with unit tests — purely additive in `analyze_aggregators.go`) and Task 3b (integrate into `cpl_analyzer.go` + integration tests). The architect-for-story-planning identified the natural seam at the boundary between method implementation and pipeline integration.

### Parallel Execution

None — all tasks share at least one modified file. Specifically, Tasks 1, 2, 3a, and 4 all modify `analyze_aggregators.go` and/or `analyze_aggregators_test.go`, making git worktree-based parallelism infeasible. Task 3b modifies `cpl_analyzer.go` and `analyze_aggregators_test.go`, sharing the test file with other tasks. The architect-for-story-planning evaluated Tasks 1 and 2 (both add independent functions to `analyze_aggregators.go`) and confirmed they cannot parallelize because they share a modified file.

### Execution Order

```
Task 1 → Task 2 → Task 3a → Task 3b → Task 4
```

All tasks are sequential. Task 3b depends on Task 3a (methods must exist before pipeline integration). Tasks 1 and 2 are independent of each other (neither calls the other's function) but share `analyze_aggregators.go` and therefore must run sequentially; they are ordered as Task 1 → Task 2 for consistency with the plan. Task 4 is a refinement on the completed pipeline and runs last.

### Reintegration

None — no parallel groups exist.

---

### Task 1: Implement findSingleAggregatorSequence — Not Started [Small]

**Architect background**: This task is appropriately sized for a single sub-agent handoff. It implements one function with 5 subtask steps, touches 2 files (source + test), and has a single verification cycle at the end. No decomposition needed.

1. a. Implement `findSingleAggregatorSequence(inputDims []string, outputDims []string, excessDims []string) ([]AggregatorMapKey, error)` in `src/main/go/analysis/analyze_aggregators.go`. For zero excess dimensions, return a nil sequence. For a single excess dimension, build the `AggregatorMapKey`, query `a.aggregators` map, and return a single-element sequence on success or an error with dimension details on failure. For multiple excess dimensions, return an error deferring to Increment 5. — Not Started
   b. Write tests in `src/main/go/analysis/analyze_aggregators_test.go`: single excess dimension found (returns sequence), single excess not found (returns descriptive error with dimension names), multiple excess dimensions (returns deferred error), zero excess dimensions (returns nil). — Not Started
   c. Run `go test ./src/main/go/analysis/... -run TestFindSingleAggregatorSequence -count=1` and verify all tests pass. — Not Started

**Intent**: Provide a deterministic function that, given a dimensional gap between metric input and output, queries the populated Aggregators map and returns either a single-element aggregator sequence or a descriptive error. This function isolates the aggregator-lookup logic so downstream code (Task 3a) can focus on the iteration and CIL-tree-wrapping logic.

**Acceptance Criteria**:
- `findSingleAggregatorSequence` correctly resolves a single excess dimension by building the AggregatorMapKey and querying the Aggregators map
- A missing aggregator produces an error message that includes the input and output dimension names
- Multiple excess dimensions produce an error stating "multiple excess dimensions require sequence enumeration (Increment 5)"
- Zero excess dimensions return a nil sequence with no error
- All unit tests pass with `go test ./src/main/go/analysis/... -run TestFindSingleAggregatorSequence`

---

### Task 2: Implement buildAggregateNode for typed CIL nodes — Not Started [Small]

**Architect background**: This task is appropriately sized. It implements one factory function with a switch dispatch, touches 2 files, and has a single verification cycle. The `CellOperation` dispatching is a mechanical pattern — each case creates the appropriate typed CIL node with values computed from `AggregatorInfo`. No decomposition needed.

2. a. Implement `buildAggregateNode(info *AggregatorInfo, sourceCIL CILNode) CILNode` in `src/main/go/analysis/analyze_aggregators.go`. Switch on `info.CellOperation`: `"sum"` → `CILAggregateSumNode` (with matrix from `BuildSumAggregationMatrix`), `"count"` → `CILAggregateCountNode`, `"size"` → `CILAggregateSizeNode` (with Size = `len(collapsedValues)`), `"1/sum"` → `CILAggregateInverseSumNode`, `"1/count"` → `CILAggregateInverseCountNode`. Extract collapsed and target concept values via `getCollapsedConceptValues` and `getTargetConceptValues` helpers on `CPLAnalyzer`. — Not Started
   b. Write tests in `src/main/go/analysis/analyze_aggregators_test.go`: `sum` produces `CILAggregateSumNode` with correct matrix dimensions (rows = len(target), cols = len(collapsed)), `size` produces `CILAggregateSizeNode` with Size equal to the collapsed value count, `count` produces `CILAggregateCountNode`, each inverse variant produces the correct node type. — Not Started
   c. Run `go test ./src/main/go/analysis/... -run TestBuildAggregateNode -count=1` and verify all tests pass. — Not Started

**Intent**: Create a factory function that constructs the appropriate typed CIL aggregate node from an `AggregatorInfo` entry. The typed CIL nodes hold pre-built aggregation matrices or constants so the synthesis phase (Story261) can emit computation code without re-deriving aggregation parameters.

**Acceptance Criteria**:
- `buildAggregateNode` dispatches on `CellOperation` and creates the correct typed CIL node for `sum`, `count`, `size`, `1/sum`, and `1/count`
- `CILAggregateSumNode`'s matrix has correct dimensions: rows = number of target concept values, columns = number of collapsed concept values
- `CILAggregateSizeNode`'s `Size` field equals the number of collapsed concept values
- All unit tests pass with `go test ./src/main/go/analysis/... -run TestBuildAggregateNode`

---

### Task 3a: Implement resolveMetricDimensionality and getMetricOutputDimensions with unit tests — Not Started [Small]

**Architect background**: This task was split from the original Task 3 (6 subtask steps, touched 3 files with a self-contained preamble). The preamble — implementing `resolveMetricDimensionality` and `getMetricOutputDimensions` and writing unit tests — forms Task 3a and is a natural seam: the methods are purely additive to `analyze_aggregators.go`, are independently compilable, and can be unit-tested by constructing `AnalysisMetricDefinition` objects without modifying `cpl_analyzer.go`. The remaining integration work and integration tests form Task 3b.

3a. a. Implement `getMetricOutputDimensions(metric *AnalysisMetricDefinition) []string` in `src/main/go/analysis/analyze_aggregators.go`. Extract the required output dimension names from the metric's IO spec (the output concepts defined in the metric's output structure). — Not Started
    b. Implement `resolveMetricDimensionality(metric *AnalysisMetricDefinition) error` in `src/main/go/analysis/analyze_aggregators.go`. For each `metric`-kind input in the metric's input list, compare the input metric's resolved dimensions against the output dimensions. Identify excess dimensions (present in input, absent from output). If no excess dimensions, skip the input. If excess dimensions exist, call `findSingleAggregatorSequence` and, for each aggregator key in the returned sequence, call `buildAggregateNode` to wrap the input's CIL tree. After wrapping, assign the resulting CIL tree chain to `input.CILTree`. Skip inputs whose `InputKind` is not `"metric"`. — Not Started
    c. Write unit tests in `src/main/go/analysis/analyze_aggregators_test.go`:
       - `getMetricOutputDimensions` returns correct dimensions from a metric IO spec
       - `resolveMetricDimensionality` wraps an input's CIL tree with aggregate nodes when a dimension mismatch exists and an aggregator is found
       - `resolveMetricDimensionality` leaves the CIL tree unchanged when dimensions match (no excess dimensions)
       - `resolveMetricDimensionality` skips lookup-kind and dimension-kind inputs (only processes `"metric"` inputs)
       - `resolveMetricDimensionality` returns an error when a dimension mismatch exists but no aggregator is found
       Construct `AnalysisMetricDefinition` objects directly in the test file to avoid depending on the full analyzer pipeline. — Not Started
    d. Run `go test ./src/main/go/analysis/... -run "TestResolveMetricDimensionality|TestGetMetricOutputDimensions" -count=1` and verify all tests pass. — Not Started

**Intent**: Implement the core dimension-mismatch detection and CIL-tree-wrapping logic. This method iterates a metric's inputs, identifies dimensional gaps by comparing input dimensions against output dimensions, and automatically bridges those gaps with aggregator chains. The unit tests verify this logic independently of the analyzer pipeline, using directly constructed test data.

**Acceptance Criteria**:
- `getMetricOutputDimensions` correctly extracts dimension names from the metric's IO spec
- `resolveMetricDimensionality` detects dimension mismatches between input and output and inserts aggregate node chains into the input CIL trees
- Inputs with matching dimensions have unmodified CIL trees
- Non-metric inputs (lookup, dimension) are skipped
- Missing aggregators produce errors with descriptive dimension information
- All unit tests pass with `go test ./src/main/go/analysis/...`

---

### Task 3b: Integrate resolveMetricDimensionality into cpl_analyzer.go second pass — Not Started

**Architect background**: This task is not annotated `[Small]` because it modifies `cpl_analyzer.go` (the core analysis pipeline) and requires understanding the full analyzer execution flow, including when metric DAG assembly completes and where the second-pass iteration fits. The integration test additions in the test file require reasoning about the end-to-end pipeline. The architect considered but rejected a further split: the two files (source + test) form one coordinated change — splitting the test-writing from the pipeline modification would not compile-check independently.

3b. a. In `src/main/go/analysis/cpl_analyzer.go`, find the point after all `AnalyzeMetricDefinitions` have resolved their input metric references and built the DAG. Add a second-pass loop that iterates all metrics and calls `resolveMetricDimensionality(metric)` for each. Accumulate any errors and return them collectively or on first failure. — Not Started
    b. Write integration tests in `src/main/go/analysis/analyze_aggregators_test.go`:
       - Full pipeline test: define a metric with a dimension mismatch and a corresponding aggregator, run the analyzer, and verify the metric's input CIL tree contains an aggregate node wrapping the original input reference
       - Error test: define a metric with a dimension mismatch and no applicable aggregator, run the analyzer, and verify a compile-time error is produced with a descriptive message showing available vs. required dimensions
       - Error test: define a metric with multiple possible aggregator sequences (multiple excess dimensions) and verify the analyzer produces the Increment 5 deferral error — Not Started
    c. Run `go test ./src/main/go/analysis/... -count=1 -timeout 60s` and verify all tests pass, including both the new integration tests and all pre-existing tests. — Not Started

**Intent**: Wire `resolveMetricDimensionality` into the CPL analyzer's second pass, so that after all metric input references are resolved and the DAG is assembled, the analyzer automatically detects dimension mismatches and inserts aggregate nodes. This integration is the step that makes aggregators actionable — without it, `resolveMetricDimensionality` exists but is never called during compilation.

**Acceptance Criteria**:
- The CPL analyzer's second pass iterates all metrics and calls `resolveMetricDimensionality` after the metric DAG is assembled
- The full pipeline produces CIL trees with inserted aggregate nodes when dimension mismatches have matching aggregators
- Missing aggregators produce compile-time errors with descriptive dimension information
- Multiple possible aggregator sequences produce a clear error deferring to Increment 5
- All tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s`

---

### Task 4: Handle scalar reduction edge case — Not Started [Small]

**Architect background**: This task addresses a single edge case — aggregation to an empty `[]` output dimension set (scalar output). It is a refinement on the completed pipeline and touches 2 files at most. Appropriately sized for `[Small]`.

4. a. In `src/main/go/analysis/analyze_aggregators.go`, ensure `findSingleAggregatorSequence` correctly handles the case where the `outputDims` slice is empty (`[]`). The `AggregatorMapKey` should represent the `to` side as `[]`. Verify that `getMetricOutputDimensions` returns an empty slice for scalar-output metrics. — Not Started
   b. Write a test in `src/main/go/analysis/analyze_aggregators_test.go`: metric output is scalar (no dimension concepts), input has `[Country]`, an aggregator `from [Country] to [] = sum(Country)` exists in the map. Run `resolveMetricDimensionality` and verify the sequence is resolved and the aggregate node is inserted. — Not Started
   c. Run `go test ./src/main/go/analysis/... -run TestScalarReduction -count=1` and verify the scalar reduction test passes. — Not Started

**Intent**: Ensure that dimension reduction to a scalar output (no concept dimensions) is correctly handled. Scalar metrics are a common use case (e.g., computing a total across all categories), and this edge case must work end-to-end through the sequence resolution and node-insertion pipeline.

**Acceptance Criteria**:
- `findSingleAggregatorSequence` accepts `outputDims = []` and produces the correct AggregatorMapKey
- `getMetricOutputDimensions` returns `[]` for scalar-output metrics
- A scalar-reduction test passes: metric input `[Country]`, output `[]`, aggregator present → sequence resolved, aggregate node inserted
- All tests pass with `go test ./src/main/go/analysis/... -run TestScalarReduction`

---

## Test-First Development

This story follows the Logical TDD Lifecycle. For each coding task, tests are written before (or alongside) the implementation code. The test-and-implement cycle is integrated within each task, not broken out as separate tasks. See `memory-bank/concepts.md` for details on the Logical TDD Lifecycle.

## Constraints

- **Single-sequence limitation**: In Increment 4, only single unambiguous aggregator sequences are supported. Multiple possible aggregator paths produce a compile-time error. Multi-sequence enumeration and selection is deferred to Increment 5.
- **Metric inputs only**: `resolveMetricDimensionality` only processes inputs with `InputKind == "metric"`. Lookup and dimension inputs do not require dimension adjustment.
- **Existing aggregator infrastructure**: This story depends on the Aggregators map (Story256) and typed CIL node types (Story258) already being in place. It does not create new aggregator types or CIL node definitions.
- **Error message clarity**: All dimension-mismatch errors must include the names of the input dimensions, the output dimensions, and the excess dimensions found, so users can diagnose misconfigurations without reading source code.

## Intent

The aggregator analysis phase transforms the compilation pipeline from "aggregators are declared and validated but not applied" to "aggregators are automatically applied when needed." After this story, when a metric's input carries more dimension concepts than the metric's output requires, the analyzer automatically locates the applicable aggregator, creates a typed aggregate node with the pre-built aggregation matrix, and inserts it into the CIL tree — all without any manual wiring by the metric author. The synthesis phase (Story261) then walks these modified CIL trees to produce computation code.

## Acceptance Criteria

1. `findSingleAggregatorSequence` resolves a single excess dimension by querying the Aggregators map and returns a single-element sequence or a descriptive error
2. Multiple excess dimensions produce a clear deferred error message directing users to Increment 5
3. `buildAggregateNode` creates the correct typed CIL node (sum, count, size, or inverse variants) with the aggregation matrix pre-built from collapsed and target concept values
4. `resolveMetricDimensionality` iterates a metric's inputs, detects dimension mismatches against output dimensions, and inserts aggregate node chains into the input CIL trees
5. `getMetricOutputDimensions` correctly extracts dimension concept names from the metric's IO spec
6. The CPL analyzer's second pass calls `resolveMetricDimensionality` for all metrics after the metric DAG is fully assembled
7. Dimension mismatches with no applicable aggregator produce compile-time errors with descriptive messages showing available vs. required dimensions
8. Metrics with no dimension mismatch have unmodified CIL trees
9. Non-metric inputs (lookup, dimension) are skipped by the dimension-checking logic
10. Scalar reduction (output dimensions = `[]`) is correctly handled end-to-end
11. All tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s`

## Open Questions

- Should `resolveMetricDimensionality` accumulate all dimension-mismatch errors across all metric inputs and report them collectively, or should it fail on the first error? The plan shows `return fmt.Errorf(...)` on first error, which is the simpler approach. If users prefer seeing all dimension issues at once, this can be revised in a follow-up story.
- The plan's description in the header mentions "output unit compatibility for aggregator-lookup chains" which may refer to a future increment's concern (unit type checking across aggregator chains). Confirm whether this story should include any unit-type validation or whether that belongs entirely to the synthesis phase.
