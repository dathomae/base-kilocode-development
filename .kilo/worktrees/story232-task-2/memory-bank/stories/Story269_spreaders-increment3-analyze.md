# Story269: Spreaders Increment 3 — Analyze

## Goal

Implement metric reference resolution in spreader expressions, implicit aggregator determination via dimensional unit analysis, required aggregator validation, and cell unit compatibility checks (Increment 3 of Epic008 — Spreaders Incremental Implementation). When a spreader expression references a metric by name (e.g., `Product / customers` for weighted distribution), the analyzer resolves the identifier at analysis time through a three-tier priority system (metrics → concepts → lookups), determines whether the metric's dimensional form matches the requirement, and if not, finds an implicit aggregator from the Aggregators map to bridge the gap. Resolved metric subtrees (possibly wrapped in `CILAggregateNode`) are stored in `CILSpreadNode.ReferencedMetrics`. This story also detects self-referencing and circular metric dependencies via spreader expressions.

## References

- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md)
- [ERO008: Spreaders Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO008_spreaders-analysis-phase.md)
- [Design: CPL Design Document](../specifications/design/cpl/design.md)

## Dependencies

- Story265 (spreaders-increment1-analyze) — `SpreaderInfo`, `SpreaderFunctionRegistry`, postfix expression evaluator, and `CILSpreadNode` must exist
- [Story267_spreaders-increment2-analyze.md](Story267_spreaders-increment2-analyze.md) - Not Started — Spreader integration into metric DAG must exist; spreader expressions with metric references are used in the same DAG context
- [Story256_aggregators-increment2-analyze.md](Story256_aggregators-increment2-analyze.md) - Not Started — Aggregators map must exist for implicit aggregator resolution
- [Story260_aggregators-increment4-analyze.md](Story260_aggregators-increment4-analyze.md) - Not Started — Aggregator selection/integration must be complete; implicit aggregator resolution queries the Aggregators map
- Epic005 (metrics) — Metrics map must exist for resolution; metric CIL trees must be available for `CILMetricRefNode` generation

## Dependent Stories

- Story270 (spreaders-increment3-synthesize) — depends on `ReferencedMetrics` in `CILSpreadNode` and the resolved CIL subtrees for code generation that computes referenced metrics (with implicit aggregators) before the spread operation

## Tasks

### Task Sizing

All four tasks are **not Small**. Task 1 creates a new file with new package-level abstractions (three-tier identifier resolution returning dual values). Task 2 implements complex algorithmic logic (dimensional unit algebra with aggregator factor matching and inverse-effect fallback). Task 3 spans two source files with coordinated changes — adding a struct field to `CILSpreadNode` in `cil_nodes.go` and updating the expression evaluator in `spreader_analyzer.go` to populate it, along with self-reference and circular-reference detection. Task 4 modifies the analysis pipeline, creates example directories, and adds validation logic. The architect-for-story-planning confirmed that Task 3 is at its natural granularity — splitting the `ReferencedMetrics` field addition from the expression evaluator update would create an isolated field-with-no-consumer task with no isolation benefit, since both halves must compile together in the same package.

### Parallel Execution

None — all tasks have sequential dependencies. Each task builds on the work of the previous task:
- Task 2 requires the identifier resolution functions from Task 1
- Task 3 requires both identifier resolution (Task 1) and implicit aggregation (Task 2) for the expression evaluator
- Task 4 requires all three preceding tasks for end-to-end validation

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

No parallel groups exist. All tasks execute sequentially.

### Task 1: Implement three-tier identifier resolution — Not Started

Create `src/main/go/analysis/spreader_metric_resolver.go` with `resolveSpreaderIdentifier`, implementing the priority-ordered resolution: metrics → concepts → lookups. Each resolution tier returns both a `*DimensionalUnitExpression` (for dimensional unit validation) and an optional `CILNode` (for code generation — a `CILMetricRefNode` wrapping the metric's CIL tree). Write tests in `src/main/go/analysis/spreader_metric_resolver_test.go` following the Logical TDD Lifecycle (write test first, then implement).

   a. Create `src/main/go/analysis/spreader_metric_resolver.go` — implement `resolveSpreaderIdentifier` with priority order: metrics → concepts → lookups — Not Started
   b. Each resolution tier returns both `*DimensionalUnitExpression` (for dim unit validation) and optional `CILNode` (for code generation) — Not Started
   c. Write unit tests in `src/main/go/analysis/spreader_metric_resolver_test.go`: metric name resolves to `CILMetricRefNode`, concept name resolves to concept even if a metric with the same name exists (concepts checked only after metrics), unknown identifier returns error — Not Started
   d. Verify resolution tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s` — Not Started

### Task 2: Implement implicit aggregator determination — Not Started

Extend `src/main/go/analysis/spreader_metric_resolver.go` with `resolveSpreaderMetricReference`, `findAggregatorByDimUnitEffect`, and `findAggregatorByInverseEffect`. When a metric's raw dimensional units do not match the required units, the analyzer computes `requiredFactor = requiredDimUnits.Divide(rawDimUnits)` and searches the Aggregators map for a matching transformation. If found, builds a `CILAggregateNode` wrapping the `CILMetricRefNode`. Write tests in `src/main/go/analysis/spreader_metric_resolver_test.go` following the Logical TDD Lifecycle.

   a. Implement `resolveSpreaderMetricReference` — checks if metric's raw dim units match required, if not searches for aggregator — Not Started
   b. Implement `findAggregatorByDimUnitEffect` — searches Aggregators map for transformation matching the required dim unit factor (`aggEffect.Equals(factor)`) — Not Started
   c. Implement `findAggregatorByInverseEffect` — fallback search for aggregators whose inverse transformation matches — Not Started
   d. Build `CILAggregateNode` wrapping `CILMetricRefNode` when implicit aggregation needed, using `AggregatorInfo.Matrix` and `AggregatorInfo.CellOperation` — Not Started
   e. Write unit tests in `src/main/go/analysis/spreader_metric_resolver_test.go`: direct match (no aggregator needed), implicit aggregator found, implicit aggregator not found (error), multiple candidate aggregators (first match wins), verify CIL node structure after implicit aggregation (`CILAggregateNode` wrapping `CILMetricRefNode` with correct `AggregationMatrix`) — Not Started
   f. Verify implicit aggregation tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s` — Not Started

### Task 3: Update expression evaluation and CILSpreadNode — Not Started

Update `evaluateSpreaderExpression` in `src/main/go/analysis/spreader_analyzer.go` to call `resolveSpreaderIdentifier` instead of concept-only resolution for identifier operands. Record resolved metric references (possibly wrapped in `CILAggregateNode`) for `CILSpreadNode.ReferencedMetrics` population. Add the `ReferencedMetrics []CILNode` field to `CILSpreadNode` in `src/main/go/analysis/cil_nodes.go`. Implement self-reference detection (spreader expression cannot reference the output metric) and circular reference detection (spreader's metric references must not create cycles in the dependency graph). Write tests in `src/main/go/analysis/spreader_analyzer_test.go` following the Logical TDD Lifecycle. The architect-for-story-planning confirmed this task is at its natural granularity — adding the `ReferencedMetrics` field to `CILSpreadNode` and updating the expression evaluator to populate it form a tightly coupled unit; splitting would create an unused-field task with no isolation benefit.

   a. Update `evaluateSpreaderExpression` in `src/main/go/analysis/spreader_analyzer.go` to call `resolveSpreaderIdentifier` instead of concept-only resolution for identifier operands — Not Started
   b. Record resolved metric references (with implicit aggregator wrappers if applicable) in the spreader's context for `CILSpreadNode.ReferencedMetrics` population — Not Started
   c. Update `CILSpreadNode` struct in `src/main/go/analysis/cil_nodes.go` to include `ReferencedMetrics []CILNode` field — Not Started
   d. Implement self-reference detection: spreader expression cannot reference the output metric (prevents infinite recursion) — Not Started
   e. Implement circular reference detection: spreader's metric references must not create cycles in the dependency graph — Not Started
   f. Write unit tests in `src/main/go/analysis/spreader_analyzer_test.go`: expression evaluation with metric reference produces correct dim units, `ReferencedMetrics` populated correctly for both direct and implicit-aggregator cases, self-reference produces error, circular reference produces error — Not Started
   g. Verify expression evaluation tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s` — Not Started

### Task 4: Cell unit compatibility and end-to-end analysis — Not Started

Implement cell unit compatibility validation between the referenced metric's cell type and the spreader's cell type. Wire the full Increment 3 analysis into the main analysis pipeline. Create `parse_and_analyze_only/` example directories for Increment 3 scenarios. Write integration tests following the Logical TDD Lifecycle.

   a. Implement cell unit compatibility validation in `src/main/go/analysis/spreader_analyzer.go`: `validateCellUnitCompatibility` compares spreader cell units against referenced metric cell units after any implicit aggregation — Not Started
   b. Wire complete Increment 3 analysis into the main analysis pipeline (ensure `resolveSpreaderIdentifier` and `resolveSpreaderMetricReference` are called during full parse→analyze flow) — Not Started
   c. Create `parse_and_analyze_only/` examples for Increment 3 in `src/main/go/analysis/parse_and_analyze_only/spreader_inc3/`: valid metric references, implicit aggregation cases, self-reference errors, circular reference errors — Not Started
   d. Verify full parse→analyze pipeline produces correct results and errors with `go test ./src/main/go/analysis/... -count=1 -timeout 60s` — Not Started

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task, write the test before implementing the code that satisfies it. The test-and-implement cycle is integrated within each task — tests are not broken out as separate tasks. See `memory-bank/concepts.md` for details on the Logical TDD Lifecycle.

## Constraints

- Only single metric references (not expressions combining multiple metrics) in Increment 3
- Implicit aggregators must be found in the Aggregators map — no cross-epic aggregation discovery
- Cell unit checking relies on `float64` type (detailed type selection deferred)
- Dynamic spreading matrix from metric data must be re-computed if concept values change
- Resolution priority is fixed at: metrics → concepts → lookups (consistent with language semantics)
- Dimensional unit algebra: `raw_dim_units * aggregator_factor = required_dim_units` — the analyzer solves for the aggregator factor

## Intent

This story introduces the most complex analysis-phase feature in the spreaders epic: metric reference resolution with implicit aggregator determination. When a spreader expression references a metric by name (e.g., `Product / customers` for weighted distribution), the analyzer resolves the identifier through a three-tier priority system, determines the metric's dimensional form, and if it does not match what the expression needs, finds an implicit aggregator from the Aggregators map to bridge the gap. The dimensional algebra solves for the required aggregator factor. When an implicit aggregator is found, a `CILAggregateNode` wrapping a `CILMetricRefNode` is stored in `CILSpreadNode.ReferencedMetrics`. The synthesizer (Increment 3 synthesis, Story270) walks these subtrees to emit code that computes referenced metrics (with implicit aggregators) before the spread operation. Self-referencing and circular dependencies are detected and rejected during analysis.

## Acceptance Criteria

1. Identifier resolution follows priority: metrics → concepts → lookups
2. Metric references in spreader expressions resolve to `CILMetricRefNode` with the metric's CIL tree
3. When a metric's dimensional form does not match requirements, implicit aggregator resolution finds the right aggregator
4. `CILAggregateNode` correctly wraps `CILMetricRefNode` when implicit aggregation is applied
5. Missing aggregator for required transformation produces error
6. Self-referencing spreader expressions produce error
7. Circular metric dependencies via spreader expressions produce error
8. `CILSpreadNode.ReferencedMetrics` populated correctly with resolved metric subtrees
9. Cell unit compatibility validated between referenced metric and spreader context
10. Concept resolution in spreader expressions still works (non-metric identifiers resolve to concepts)
11. All `parse_and_analyze_only` examples pass
12. All tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s`
