# Epic007: CPL Aggregators Implementation

## Title

CPL Aggregators Implementation

## Ultimate Goal

Implement the complete aggregators functionality for the Chronocone Planning Language across all processing phases (parsing, analysis, synthesis), delivering tensor matrix multiplication infrastructure, aggregator declarations with generic function call syntax, a FunctionRegistry with operation-specific CIL node types, dimensional unit analysis, and integration into the metric dependency DAG with sequence optimization. Each increment produces a working end-to-end executable with test data and examples.

## References

- [Aggregators Design](../specifications/design/cpl/aggregators.md) — Full aggregators design with definition types, dimensional unit analysis, and data flow
- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md) — Parent ERO defining the 4-increment strategy
- [ERO007: Aggregators Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO007_aggregators-parsing-phase.md) — Parsing phase data structures, grammar, synthetic keys
- [ERO007: Aggregators Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO007_aggregators-analysis-phase.md) — Analysis phase data structures, dimensional unit analysis, sequence selection
- [ERO007: Aggregators Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md) — Tensor matrix multiplication, CIL tree walking, optimization
- [CPL Requirements — Aggregators and Spreaders](../specifications/requirements/cpl/aggregators-and-spreaders-syntax.md) — Syntax requirements
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md) — Concept design (aggregator matrices from concept value sets)
- [Metrics Design](../specifications/design/cpl/metrics-design.md) — Metric design (aggregators applied during metric DAG analysis)
- [Units and Measures Design](../specifications/design/cpl/units-and-measures.md) — Unit/measure design (dimensional unit algebra)
- [Postfix Expression Parsing Pattern](../patterns/postfix-expression-parsing.md) — Pattern for expression handling
- [Symbol Table-Driven Parsing Pattern](../patterns/symbol-table-driven-parsing.md) — Architectural pattern for parsing

## Dependencies

- [Epic001_cpl-concepts-implementation.md](Epic001_cpl-concepts-implementation.md) — Concepts must be implemented (aggregator from/to reference concepts, matrices built from concept values)
- [Epic005_cpl-metrics-implementation.md](Epic005_cpl-metrics-implementation.md) — Metrics must be implemented for Increment 3 (aggregators applied during metric DAG analysis)
- [Epic006_cpl-lookups-implementation.md](Epic006_cpl-lookups-implementation.md) — Lookups needed for Increment 4 (lookups referenced in aggregator expressions)
- [Epic008_cpl-spreaders-implementation.md](Epic008_cpl-spreaders-implementation.md) — Spreaders reuse the same matrix multiplication infrastructure (Increment 1)

## Stories

| Story | Status | Description | ERO |
|-------|--------|-------------|-----|
| StoryXXX_aggregators-increment1-matrix-mult.md (not yet defined) | Not Started | Implement full tensor matrix multiplication (dimension-reducing and dimension-expanding), composeIndex/decomposeIndex helpers, unit tests with hand-constructed tensors | [ERO007_aggregators-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md) |
| StoryXXX_aggregators-increment2-parse.md (not yet defined) | Not Started | Implement CPLAggregators.g4 grammar, AggregatorsListener, ParserAggregatorDefinition, synthetic key assignment, postfix expression storage, lexer tokens | [ERO007_aggregators-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO007_aggregators-parsing-phase.md) |
| StoryXXX_aggregators-increment2-analyze.md (not yet defined) | Not Started | Implement AnalyzeAggregatorScope, AggregatorInfo, DimensionalUnitExpression, postfix evaluator, dimensional unit validation, CILAggregateNode, AggregationMatrix building | [ERO007_aggregators-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO007_aggregators-analysis-phase.md) |
| StoryXXX_aggregators-increment2-synthesize.md (not yet defined) | Not Started | Implement CIL tree walker for CILAggregateNode, matrix emission, manual sum aggregation test code, end-to-end aggregator examples | [ERO007_aggregators-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md) |
| StoryXXX_aggregators-increment3-analyze.md (not yet defined) | Not Started | Define operation-specific CIL node types (sum, count, size, 1/sum, 1/count), extend FunctionRegistry, implement resolveAggregatorFnCall, CIL node insertion in expressions | [ERO007_aggregators-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO007_aggregators-analysis-phase.md) |
| StoryXXX_aggregators-increment3-synthesize.md (not yet defined) | Not Started | Implement CIL tree walker for all function CIL node types, size(C) constant emission, expression-integration examples | [ERO007_aggregators-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md) |
| StoryXXX_aggregators-increment4-analyze.md (not yet defined) | Not Started | Implement dimension mismatch detection during metric DAG analysis, Aggregators map query, single unambiguous sequence resolution, CIL aggregate node insertion into metric CIL trees | [ERO007_aggregators-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO007_aggregators-analysis-phase.md) |
| StoryXXX_aggregators-increment4-synthesize.md (not yet defined) | Not Started | Generate metric computation code with aggregator integration, scheduled topological computation order, aggregator integration examples | [ERO007_aggregators-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md) |
| StoryXXX_aggregators-increment5-analyze.md (not yet defined) | Not Started | Implement full sequence enumeration, selection algorithm (least-info-loss → minimal-work → arbitrary), ambiguous aggregator resolution, specificity comparison | [ERO007_aggregators-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO007_aggregators-analysis-phase.md) |
| StoryXXX_aggregators-increment5-synthesize.md (not yet defined) | Not Started | Implement fused matrix operations, scalar reduction edge cases, multi-step sequence optimization, sequence optimization examples | [ERO007_aggregators-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md) |

## Story Ordering and Dependencies

Stories within this epic are arranged by increment, and within each increment by phase (parse → analyze → synthesize). This ordering is strict:

**Increment 1 (Tensor Matrix Multiplication Infrastructure):**
1. `aggregators-increment1-matrix-mult` — Standalone infrastructure with no CPL dependencies. Pure Go tensor math, independently testable. No parse/analyze stories in this increment.

**Increment 2 (Basic Aggregator Parsing and Validation):**
2. `aggregators-increment2-parse` — Depends on Increment 1 (matrix multiplication) being complete for synthesis to succeed. Depends on Epic001 (concepts), Epic005 (metrics) for postfix expression pattern.
3. `aggregators-increment2-analyze` — Depends on `aggregators-increment2-parse` (needs parser definitions).
4. `aggregators-increment2-synthesize` — Depends on `aggregators-increment2-analyze` (needs CIL nodes and pre-built matrices). Depends on `aggregators-increment1-matrix-mult` for ReduceDimension.

**Increment 3 (Additional Aggregator Function CIL Node Types):**
5. `aggregators-increment3-analyze` — Depends on `aggregators-increment2-analyze` and `aggregators-increment2-synthesize` (needs working synthesis to validate CIL node types). Extends FunctionRegistry with `count`, `size`, `1/sum`, `1/count`.
6. `aggregators-increment3-synthesize` — Depends on `aggregators-increment3-analyze` (needs new CIL node type definitions).

**Increment 4 (Aggregator Selection and Integration):**
7. `aggregators-increment4-analyze` — Depends on `aggregators-increment3-analyze` (needs full FunctionRegistry and CIL node types). Depends on Epic005 (metrics) for metric DAG analysis infrastructure.
8. `aggregators-increment4-synthesize` — Depends on `aggregators-increment4-analyze` and `aggregators-increment3-synthesize`.

**Increment 5 (Aggregator Sequence Optimization):**
9. `aggregators-increment5-analyze` — Depends on `aggregators-increment4-analyze` (extends sequence finding to multiple).
10. `aggregators-increment5-synthesize` — Depends on `aggregators-increment5-analyze` and `aggregators-increment4-synthesize`.

No parser stories exist for Increments 3, 4, and 5 — function CIL node types, aggregator selection, and sequence optimization are analysis/synthesis operations.

Stories from Epic005 (metrics) and Epic006 (lookups) may be interleaved with these stories. Increment 4 requires coordination with Epic006 for lookup references in aggregator expressions.

## Constraints

1. Must follow the [Symbol Table-Driven Parsing Pattern](../patterns/symbol-table-driven-parsing.md) — strict parse/analysis separation
2. Must follow the [Postfix Expression Parsing Pattern](../patterns/postfix-expression-parsing.md) for aggregator expression handling
3. Each increment must produce a working executable with parse → analyze → synthesize pipeline (except Increment 1 which is synthesis-only)
4. Each increment must include examples: `parse_and_analyze_only/` (with and without errors) and `with_data/` (CPL + IO spec + parquet + expectations)
5. Aggregators are unnamed — stored under synthetic keys (`__aggregator__[from]->[to]`); keys must never leak into user-visible output
6. Each aggregator reduces by exactly one dimension (or to scalar via `to []`); multi-dimension reductions are composed from sequences
7. Matrix multiplication must support both dimension reduction (aggregators) and expansion (spreaders in Epic008)
8. Types initially represented as `float64`
9. The `indent` token enforces readable indentation but does not define scope boundaries

## Acceptance Criteria

1. All 10 stories in the epic are marked as Completed
2. Matrix multiplication infrastructure (Increment 1) passes all unit tests for reduction and expansion operations
3. Aggregators (Increment 2) can be declared, parsed, analyzed, and synthesized into a working executable with manual aggregation
4. Metrics (Increment 3) automatically apply aggregators when dimension mismatches are detected, producing correct aggregated results
5. Multi-step aggregator sequences (Increment 4) are correctly selected via the optimization algorithm and produce identical results to any valid ordering
6. Aggregators referenced in lookup expressions (cross-Epic006) produce correct results with dimensional unit validation
7. All examples in `parse_and_analyze_only/` and `with_data/` directories pass validation
8. Error cases produce meaningful compile-time or runtime errors

## Notes

- Increment 1 (matrix multiplication) is a reusable foundation — it is also a dependency for Epic008 (spreaders)
- The dimensional unit algebra (DimensionalUnitExpression) reuses the same `*`, `/`, `^` machinery as the unit/measure conversion system
- Scalar reduction (`to []`) produces rank-zero tensors — the synthesis must handle this edge case
- Fused matrix operations in Increment 4 provide performance optimization but produce identical results to sequential operations
- This epic depends on Epics 001 and 005 being sufficiently advanced; Increment 4 additionally depends on Epic006
