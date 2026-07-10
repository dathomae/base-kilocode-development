# Story271: Spreaders Increment 4 — Analyze

## Goal

Implement the full selection algorithm for mixed aggregator/spreader sequences, ambiguous spreader resolution via specificity comparison, and multi-step interleaved sequence optimization (Increment 4 of Epic008). When multiple spreader/aggregator sequences can bridge a dimensional gap between metric inputs and outputs, the analyzer enumerates all valid interleaved sequences, ranks them by information loss (primary criterion: prefer spreading late, aggregating early) and computational work (secondary criterion: smaller matrix operations), and selects the optimal sequence deterministically. When multiple spreaders match the same dimensional pattern, specificity comparison (number of `from` concepts) selects the best fit; equal-specificity ties produce analysis errors. The selected sequence is inserted into metric CIL trees as interleaved `CILSpreadNode`/`CILAggregateNode` chains. Increment 2's single-unambiguous-sequence behavior is preserved as a regression path.

## References

- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md) — Increment 4 sketch
- [ERO008: Spreaders Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO008_spreaders-analysis-phase.md) — Increment 4 data structures and algorithms
- [Plan: Story271](../plans/Story271_spreaders-increment4-analyze.md)

## Dependencies

- Story269 (spreaders-increment3-analyze) — metric reference resolution must be complete; spreader analysis infrastructure fully built; `spreader_dag_integration.go` with `findSingleSpreaderSequence` exists
- Story267 (spreaders-increment2-analyze) — single-unambiguous sequence finding must exist; Increment 4 extends this to handle multiple candidates and interleaving
- Story260 (aggregators-increment4-analyze) — aggregator selection algorithm and aggregator integration must be complete; mixed sequences require both aggregator and spreader resolution
- Epic005 (metrics) — metric DAG analysis infrastructure must be fully functional for both aggregator and spreader insertion

## Dependent Stories

- [Story272_spreaders-increment4-synthesize.md](Story272_spreaders-increment4-synthesize.md) — Not Started — Depends on `DimensionalOp` struct and interleaved `CILSpreadNode`/`CILAggregateNode` chains produced by the Increment 4 analyzer for fused operation synthesis

## Tasks

### Task Sizing

Task 5 qualifies as `[Small]` — it is a verification-only task that adds regression test functions to existing test files, following established patterns, with no new production code or abstractions. Tasks 1 through 4 are **not Small**: Task 1 creates a new file with new package-level abstractions (`DimensionalOp` struct, recursive enumeration); Task 2 implements four coordinated methods forming a scoring/selection algorithm; Task 3 implements grouping and specificity comparison logic; Task 4 touches 3+ files with coordinated integration changes across the analysis pipeline.

Architect-for-story-planning analysis confirmed that all five tasks are at their natural granularity. No natural decomposition seams exist that would produce isolated units without artificial coordination overhead. Tasks 1–3 share the same source file (`mixed_sequence_selector.go`) and represent incremental feature additions to the same subsystem — splitting any one would create a partial, non-compilable intermediate state. Task 4 is borderline (6 subtasks, multiple files) but the core integration work (subtasks a–c) forms one tightly-coupled unit — the `resolveMetricDimensionalitySpreaders` update cannot be split from the ambiguous-resolution wiring without creating an isolated half-integration. The examples in subtask e are follow-on data work, not a separable coding concern.

### Parallel Execution

None — all tasks have sequential dependencies. Tasks 1–3 share the same source file (`mixed_sequence_selector.go`) and must execute in order. Task 4 depends on the complete selection algorithm (Tasks 1–3) and ambiguous resolution (Task 3). Task 5 depends on full integration (Task 4) for regression verification. File analysis confirms no disjoint file sets exist.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4 → Task 5
```

- Task 1 must complete first (enumeration is the foundation for all subsequent work)
- Task 2 depends on Task 1 (selection ranks enumerated sequences)
- Task 3 depends on Tasks 1 and 2 (ambiguous resolution may change spreader map before enumeration is re-run; the `DimensionalOp` struct is needed for consistency)
- Task 4 depends on all preceding tasks (integration uses the full selection + resolution pipeline)
- Task 5 is final verification

### Reintegration

None — no parallel groups.

### Task 1: Implement mixed sequence enumeration — Not Started

Create `src/main/go/analysis/mixed_sequence_selector.go` with `DimensionalOp` struct, `setDifference` helper, `enumerateMixedRecursive`, and `enumerateMixedSequences`. Write tests in `src/main/go/analysis/mixed_sequence_selector_test.go` following the Logical TDD Lifecycle (write test first, then implement).

   a. Create `src/main/go/analysis/mixed_sequence_selector.go` — define `DimensionalOp` struct with fields: `OpType` (string: "aggregate" or "spread"), `OpKey` (string: map key for the operation), `FromDimensions` ([]string), `ToDimensions` ([]string), `Matrix` (*AggregationMatrix), `CILNode` (CILNode: CILAggregateNode or CILSpreadNode), `NewDimension` (string: for spread, the dimension being added), `CollapsedDim` (string: for aggregate, the dimension being collapsed) — Not Started
   b. Implement `setDifference(a, b []string) []string` — returns elements in a not in b; used to determine dimensions to remove (via aggregators) and dimensions to add (via spreaders) — Not Started
   c. Implement `enumerateMixedRecursive` — depth-first enumeration of valid interleaved aggregator/spreader sequences; respects the constraint that aggregators can only collapse dimensions that currently exist and spreaders can only add dimensions that don't currently exist — Not Started
   d. Implement `enumerateMixedSequences(fromDims, toDims []*AnalysisConceptDefinition) ([][]DimensionalOp, error)` — public entry point that computes toRemove/toAdd via setDifference, calls enumerateMixedRecursive, and returns all valid sequences — Not Started
   e. Write unit tests in `src/main/go/analysis/mixed_sequence_selector_test.go`: enumerate simple 1-step cases (single aggregator, single spreader), multi-step interleaved (aggregate A, spread to C, aggregate B, spread to D), verify all returned sequences are valid (respect dimensional constraints at each step), verify no valid sequences produces empty result — Not Started
   f. Verify enumeration tests pass: `go test ./src/main/go/analysis/... -run MixedSequence -count=1` — Not Started

### Task 2: Implement full selection algorithm — Not Started

Add `computeInformationLoss`, `computeWorkCost`, `sequenceScore`, and `selectOptimalSequence` methods to `src/main/go/analysis/mixed_sequence_selector.go`. Write tests in `src/main/go/analysis/mixed_sequence_selector_test.go` following the Logical TDD Lifecycle (write test, then implement).

   a. Implement `computeInformationLoss(seq []DimensionalOp) int` — scores sequences by how much dimension information is preserved: aggregating as early as possible preserves information (fewer dimensions lost early); spreading as late as possible preserves information (more dimensions in intermediate results). For each operation: aggregation adds `currentDims` to loss (earlier aggregation = higher loss because more dimensions exist at that point), spreading increments `currentDims` — Not Started
   b. Implement `computeWorkCost(seq []DimensionalOp) int` — estimates computational cost using matrix dimension products as proxy; spread operations multiply cost by 2 (expansion is slightly more work than reduction). For each operation: `cost += op.Matrix.Rows * op.Matrix.Columns` — Not Started
   c. Implement `sequenceScore(seq []DimensionalOp) int` — combines information loss and work cost with weighted priority: `infoLoss*1000000 + workCost`. Information loss is the primary criterion; work cost only differentiates when information loss is equal — Not Started
   d. Implement `selectOptimalSequence(sequences [][]DimensionalOp) ([]DimensionalOp, error)` — sorts sequences by score (ascending), returns the best. Returns error if sequences slice is empty — Not Started
   e. Write unit tests in `src/main/go/analysis/mixed_sequence_selector_test.go`: verify info loss ranking (late spread beats early spread), verify work cost tie-breaking (equal info loss, lower work cost selected), verify deterministic selection (same inputs produce same output), verify optimal sequence for mixed aggregator/spreader (e.g., aggregate country then spread to region vs spread to region then aggregate country) — Not Started
   f. Verify selection tests pass: `go test ./src/main/go/analysis/... -run "Selection|Optimal" -count=1` — Not Started

### Task 3: Implement ambiguous spreader resolution — Not Started

Add `resolveAmbiguousSpreaders` method to `src/main/go/analysis/mixed_sequence_selector.go`. Write tests in `src/main/go/analysis/mixed_sequence_selector_test.go` following the Logical TDD Lifecycle (write test, then implement).

   a. Implement `resolveAmbiguousSpreaders() error` — groups spreaders by map key, selects the most specific spreader per key; replaces the spreader map entry with the winner — Not Started
   b. Specificity is defined as `len(info.FromConcepts)` — more from concepts means the spreader requires more known dimensions and is therefore a better dimensional fit. A spreader `from [Country, Product] to [Country, Product, Region]` is more specific (specificity=2) than `from [Country] to [Country, Product, Region]` (specificity=1) — Not Started
   c. When multiple spreaders share the same pattern AND have equal specificity, return error: `"ambiguous spreaders for pattern %s: multiple spreaders with equal specificity (%d)"` — Not Started
   d. Write unit tests in `src/main/go/analysis/mixed_sequence_selector_test.go`: two spreaders with same pattern and different specificity → most specific wins; two spreaders with same pattern and equal specificity → error; single spreader → no change; spreaders with different patterns → no interference — Not Started
   e. Verify ambiguity resolution tests pass: `go test ./src/main/go/analysis/... -run Ambiguous -count=1` — Not Started

### Task 4: Integrate Increment 4 into DAG analysis — Not Started

Update `src/main/go/analysis/spreader_dag_integration.go` to use Increment 4 selection when multiple sequences exist, wire ambiguous spreader resolution into the analysis pipeline, add integration tests, and create `parse_and_analyze_only/` examples.

   a. Update `resolveMetricDimensionalitySpreaders` in `src/main/go/analysis/spreader_dag_integration.go` (the function formerly named `findSingleSpreaderSequence` in Increment 2, updated for Increment 3) to: (1) call `enumerateMixedSequences` to generate all valid sequences, (2) call `selectOptimalSequence` to pick the best, (3) chain CIL nodes (`CILSpreadNode`/`CILAggregateNode`) for the optimal sequence — Not Started
   b. Preserve Increment 2 single-sequence path as fallback: when `enumerateMixedSequences` returns exactly one sequence, use it directly without scoring (matches Increment 2 behavior exactly). When zero sequences exist, return error — Not Started
   c. Wire `resolveAmbiguousSpreaders` into the analysis pipeline: call it once before DAG analysis (in `CPLAnalyzer.Analyze` or equivalent entry point) so that the spreader map is resolved before any metric DAG resolution queries it — Not Started
   d. Write integration tests: metric with multiple valid spreader sequences (verifies optimal selection), metric with mixed aggregator/spreader requirement (verifies interleaved chains), metric with interleaved operations (verifies multi-step sequence correctness). Run with: `go test ./src/main/go/analysis/... -run "Integration|Interleave" -count=1` — Not Started
   e. Create `parse_and_analyze_only/` examples: mixed sequences (aggregator + spreader interleaving), ambiguous spreaders (specificity resolution), optimal selection scenarios (multiple sequences → best picked). Add expected outputs verifying `parse_and_analyze_only` success — Not Started
   f. Verify full analyze pipeline with Increment 4 features: `go test ./src/main/go/analysis/... -count=1 -timeout 120s` — Not Started

### Task 5: Regression tests — Not Started [Small]

   a. Verify Increment 2 behavior unchanged: single unambiguous spreader sequences still work correctly. Add a regression test confirming that when only one sequence exists, Increment 4 selects it without error. Run with: `go test ./src/main/go/analysis/... -run "SingleSpread" -count=1` — Not Started
   b. Verify Increment 3 behavior unchanged: metric references in spreader expressions still resolve correctly. Add a regression test confirming spreader expressions with metric references work end-to-end. Run with: `go test ./src/main/go/analysis/... -run "MetricRef" -count=1` — Not Started
   c. Run full test suite: `go test ./src/main/go/analysis/... -count=1 -timeout 120s` — confirm no regressions — Not Started

## Test-First Development

This story follows the Logical TDD Lifecycle. For coding tasks (Tasks 1–4), write tests before implementing the code. Each task includes the test-and-implement cycle in the correct order — tests are written first (subtask e in each task), then verified to fail, then the implementation (subtasks a–d) is completed, and finally all tests pass (subtask f). Do not break tests out as separate tasks.

## Constraints

- **No parsing changes**: This story is analysis-phase only. No grammar modifications, no ANTLR changes, no parsing listener modifications.
- **Enumeration correctness**: `enumerateMixedRecursive` must respect the constraint that aggregators can only collapse dimensions that currently exist and spreaders can only add dimensions that don't currently exist. Each intermediate state in an enumerated sequence must be dimensionally valid.
- **Selection determinism**: The selection algorithm must be deterministic — same inputs must always produce the same optimal sequence. The deterministic tiebreaker (Criterion 3: sorted operation key) ensures this when information loss and work cost are equal.
- **Information loss scoring**: Information loss uses the formula described in the plan. Aggregation at step `i` adds `currentDims` to the loss score (earlier aggregation loses more information). Spreading increments `currentDims` by 1 (delays information loss). The `*1000000` weight ensures information loss dominates work cost as the primary criterion.
- **Work cost**: Uses matrix dimension products (`Rows * Columns`) as a proxy for computational expense. Spread operations multiply cost by 2. No heuristic tuning — simple matrix dimension product only.
- **Specificity definition**: Specificity equals `len(info.FromConcepts)`. A spreader requiring more known dimensions is more specific. Equal specificity with the same pattern is an error.
- **State space**: For typical CPL metrics (3–5 dimensions), enumeration is fast. No pruning needed for initial implementation.
- **Fused operations**: Handled in synthesis (Story272). This story produces individual `CILSpreadNode`/`CILAggregateNode` chains — the synthesizer may fuse them later.
- **Compile guard**: All new Go files must pass `go build ./src/main/go/analysis/...` before tests can compile.

## Intent

### Task 1: Mixed sequence enumeration

The goal is to create the foundational data structure (`DimensionalOp`) and enumeration algorithm. When a metric's input dimensions don't match its output dimensions, the analyzer must find sequences of aggregators and spreaders that transform one to the other. The enumeration must generate all valid sequences, respecting the constraint that operations can only use dimensions that exist at that point. The `DimensionalOp` struct bridges analysis and synthesis — it carries both dimensional metadata and CIL nodes for code generation. The `setDifference` helper isolates the dimensions-to-remove and dimensions-to-add sets. The recursive enumeration explores all valid interleavings.

### Task 2: Full selection algorithm

The goal is to rank enumerated sequences by quality and select the best one. The three-tier criteria (information loss → work cost → deterministic tiebreak) ensure that the selected sequence preserves the most dimensional information while minimizing computational work. Information loss scoring formalizes the intuition that spreading late (keeping more dimensions in intermediate results) and aggregating early (removing dimensions only when needed) preserves information. Work cost uses matrix sizes as a simple proxy — smaller matrices mean less computation. The `sequenceScore` function combines these with a large weight on information loss to ensure it dominates.

### Task 3: Ambiguous spreader resolution

The goal is to handle the case where multiple spreaders match the same dimensional pattern. Rather than erroring on ambiguity immediately, the analyzer attempts to resolve via specificity: a spreader that requires more known `from` concepts is more specific and therefore a better fit. If two spreaders are equally specific for the same pattern, the ambiguity is genuine and the analyzer must error — there is no basis to choose between them. This resolution runs once before DAG analysis so the rest of the pipeline sees a clean spreader map.

### Task 4: DAG integration

The goal is to wire the new Increment 4 capabilities into the existing metric DAG analysis pipeline. `resolveMetricDimensionalitySpreaders` (the function that bridges dimensional gaps in metric inputs) is updated to use the full enumeration and selection algorithm when multiple sequences exist, while preserving Increment 2's single-sequence fast path. Ambiguous spreader resolution runs as a pre-pass before DAG analysis. Integration tests verify the complete pipeline and `parse_and_analyze_only/` examples provide standalone verification scenarios.

### Task 5: Regression tests

The goal is to confirm that Increment 2 and Increment 3 behaviors are preserved — single unambiguous sequences still work, metric references in spreader expressions still resolve correctly, and no existing tests fail. The full test suite serves as a final guard against regressions.

## Acceptance Criteria

1. `enumerateMixedSequences` generates all valid interleaved aggregator/spreader sequences that transform `fromDims` to `toDims`
2. Each enumerated sequence respects dimensional constraints at every intermediate step (only existing dimensions are aggregated, only missing dimensions are spread to)
3. Selection algorithm picks the least-information-loss sequence when multiple exist
4. When information loss is equal, the minimal-work sequence is selected
5. Selection is deterministic — same inputs always produce the same optimal sequence
6. Single unambiguous sequences (Increment 2 behavior) still work as a regression path
7. Ambiguous spreaders are resolved by specificity — more `from` concepts means more specific, and the most specific spreader is selected
8. Equal-specificity ambiguous spreaders produce a clear error with the pattern key and specificity value
9. Mixed aggregator/spreader sequences are correctly chained in metric CIL trees as interleaved `CILSpreadNode` and `CILAggregateNode` chains
10. Interleaved operations maintain dimensional correctness at each step in the chain
11. All `parse_and_analyze_only/` examples pass verification without errors
12. No regressions in Increment 1, 2, or 3 behavior — existing tests continue to pass
