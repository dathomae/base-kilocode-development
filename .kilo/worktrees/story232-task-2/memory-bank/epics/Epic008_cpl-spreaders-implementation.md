# Epic008: CPL Spreaders Implementation

## Title

CPL Spreaders Implementation

## Ultimate Goal

Implement the complete spreaders functionality for the Chronocone Planning Language across all processing phases (parsing, analysis, synthesis), delivering spreader declarations with dimensional transformation patterns, generic function call expressions, a spreader-specific FunctionRegistry, metric reference resolution with implicit aggregator determination, integration into the metric dependency DAG, and mixed aggregator/spreader sequence optimization. Each increment produces a working end-to-end executable with test data and examples.

## References

- [Spreaders Design](../specifications/design/cpl/spreaders.md) — Full spreaders design with definition types, dimensional unit analysis, and data flow
- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md) — Parent ERO defining the 4-increment strategy
- [ERO008: Spreaders Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO008_spreaders-parsing-phase.md) — Parsing phase data structures, grammar, synthetic keys
- [ERO008: Spreaders Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO008_spreaders-analysis-phase.md) — Analysis phase data structures, FunctionRegistry, metric reference resolution
- [ERO008: Spreaders Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md) — CIL tree walking, dimension expansion, implicit aggregator execution
- [CPL Requirements — Aggregators and Spreaders](../specifications/requirements/cpl/aggregators-and-spreaders-syntax.md) — Syntax requirements
- [Aggregators Design](../specifications/design/cpl/aggregators.md) — Aggregator design (symmetric inverse, shared matrix multiplication)
- [Metrics Design](../specifications/design/cpl/metrics-design.md) — Metric design (spreaders applied during metric DAG analysis)
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md) — Concept design (spreader matrices from concept value sets)
- [Units and Measures Design](../specifications/design/cpl/units-and-measures.md) — Unit/measure design (dimensional unit algebra)
- [Postfix Expression Parsing Pattern](../patterns/postfix-expression-parsing.md) — Pattern for expression handling
- [Symbol Table-Driven Parsing Pattern](../patterns/symbol-table-driven-parsing.md) — Architectural pattern for parsing

## Dependencies

- [Epic001_cpl-concepts-implementation.md](Epic001_cpl-concepts-implementation.md) — Concepts must be implemented (spreader from/to reference concepts, matrices built from concept values)
- [Epic005_cpl-metrics-implementation.md](Epic005_cpl-metrics-implementation.md) — Metrics must be implemented for Increment 2 (spreaders applied during metric DAG analysis)
- [Epic007_cpl-aggregators-implementation.md](Epic007_cpl-aggregators-implementation.md) — Aggregators must be at Increment 2+ for matrix multiplication infrastructure and aggregate map for implicit resolution (Increment 3)
- [Epic006_cpl-lookups-implementation.md](Epic006_cpl-lookups-implementation.md) — Lookups needed when spreader expressions reference lookups (future integration)

## Stories

| Story | Status | Description | ERO |
|-------|--------|-------------|-----|
| StoryXXX_spreaders-increment1-parse.md (not yet defined) | Not Started | Implement CPLSpreaders.g4 grammar, SpreadersListener, ParserSpreaderDefinition, synthetic key assignment, generic fnCall + identifier expression support, lexer tokens | [ERO008_spreaders-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO008_spreaders-parsing-phase.md) |
| StoryXXX_spreaders-increment1-analyze.md (not yet defined) | Not Started | Implement AnalyzeSpreaderScope, SpreaderInfo, SpreaderFunctionRegistry (all functions), expression evaluator with identifier resolution, dimensional unit validation, CILSpreadNode, SpreadingMatrix building | [ERO008_spreaders-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO008_spreaders-analysis-phase.md) |
| StoryXXX_spreaders-increment1-synthesize.md (not yet defined) | Not Started | Implement CIL tree walker for CILSpreadNode, spreading matrix emission, ExpandDimension integration, manual spreader test code, end-to-end spreader examples | [ERO008_spreaders-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md) |
| StoryXXX_spreaders-increment2-analyze.md (not yet defined) | Not Started | Implement dimension mismatch detection during metric DAG analysis, Spreaders map query, single unambiguous sequence resolution, CILSpreadNode insertion into metric CIL trees | [ERO008_spreaders-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO008_spreaders-analysis-phase.md) |
| StoryXXX_spreaders-increment2-synthesize.md (not yet defined) | Not Started | Generate metric computation code with spreader integration, scheduled topological computation order, spreader integration examples with single unambiguous paths | [ERO008_spreaders-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md) |
| StoryXXX_spreaders-increment3-analyze.md (not yet defined) | Not Started | Implement metric reference resolution in spreader expressions, implicit aggregator determination via dimensional unit analysis, required aggregator validation, cell unit compatibility checks | [ERO008_spreaders-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO008_spreaders-analysis-phase.md) |
| StoryXXX_spreaders-increment3-synthesize.md (not yet defined) | Not Started | Generate code for referenced metric computation with implicit aggregators, dynamic spreading matrix construction from metric data, weighted distribution examples | [ERO008_spreaders-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md) |
| StoryXXX_spreaders-increment4-analyze.md (not yet defined) | Not Started | Implement mixed aggregator/spreader sequence enumeration, full selection algorithm (least-info-loss → minimal-work → arbitrary), ambiguous spreader resolution, specificity comparison | [ERO008_spreaders-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO008_spreaders-analysis-phase.md) |
| StoryXXX_spreaders-increment4-synthesize.md (not yet defined) | Not Started | Implement fused spreader/aggregator operations, scalar spreading edge cases, multi-step interleaved sequence optimization, optimization examples | [ERO008_spreaders-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md) |

## Story Ordering and Dependencies

Stories within this epic are arranged by increment, and within each increment by phase (parse → analyze → synthesize). This ordering is strict:

**Increment 1 (Basic Spreader Parsing and Validation):**
1. `spreaders-increment1-parse` — No dependencies on other spreader stories. Depends on Epic001 (concepts), Epic005 (metrics) for postfix expression pattern. Depends on Epic007 Increment 1 (matrix multiplication infrastructure).
2. `spreaders-increment1-analyze` — Depends on `spreaders-increment1-parse` (needs parser definitions). Depends on Epic007 Increment 2 for dimensional unit analysis patterns.
3. `spreaders-increment1-synthesize` — Depends on `spreaders-increment1-analyze` (needs CIL nodes and pre-built matrices). Depends on Epic007 Increment 1 for ExpandDimension.

**Increment 2 (Spreader Selection and Integration):**
4. `spreaders-increment2-analyze` — Depends on `spreaders-increment1-analyze` (needs Spreaders map). Depends on Epic005 (metrics) for metric DAG analysis infrastructure. Only single unambiguous sequences supported.
5. `spreaders-increment2-synthesize` — Depends on `spreaders-increment2-analyze` and `spreaders-increment1-synthesize`.

**Increment 3 (Metric References in Spreader Expressions):**
6. `spreaders-increment3-analyze` — Depends on `spreaders-increment1-analyze` (needs Spreaders map and FunctionRegistry). Depends on Epic007 Increment 2+ (needs Aggregators map for implicit resolution).
7. `spreaders-increment3-synthesize` — Depends on `spreaders-increment3-analyze` and `spreaders-increment1-synthesize`. Depends on Epic007 Increment 2+ for ReduceDimension (implicit aggregators).

**Increment 4 (Spreader/Aggregator Interleaving and Optimization):**
8. `spreaders-increment4-analyze` — Depends on `spreaders-increment3-analyze` and `spreaders-increment2-analyze` (extends sequence finding to mixed interleaved sequences).
9. `spreaders-increment4-synthesize` — Depends on `spreaders-increment4-analyze` and `spreaders-increment3-synthesize`.

No parser stories exist for Increments 2, 3, and 4 — spreader selection, metric reference resolution, and interleaving optimization are analysis/synthesis operations.

Stories from Epic005 (metrics), Epic006 (lookups), and Epic007 (aggregators) may be interleaved with these stories. Increment 3 requires coordination with Epic007 for implicit aggregator resolution. Increment 4 requires both Epic007 and Epic005 to be sufficiently advanced.

## Constraints

1. Must follow the [Symbol Table-Driven Parsing Pattern](../patterns/symbol-table-driven-parsing.md) — strict parse/analysis separation
2. Must follow the [Postfix Expression Parsing Pattern](../patterns/postfix-expression-parsing.md) for spreader expression handling
3. Each increment must produce a working executable with parse → analyze → synthesize pipeline
4. Each increment must include examples: `parse_and_analyze_only/` (with and without errors) and `with_data/` (CPL + IO spec + parquet + expectations)
5. Spreaders are unnamed — stored under synthetic keys (`__spreader__[from]->[to]`); keys must never leak into user-visible output
6. Each spreader expands by exactly one dimension (or from scalar `from []` to a single dimension); multi-dimension expansions are composed from sequences
7. Spreader functions use generic `<identifier>(<paramList>)` syntax — no reserved function name tokens
8. Tensor matrix multiplication infrastructure is reused from Epic007 Increment 1 — no new matrix math needed
9. Types initially represented as `float64`
10. The `indent` token enforces readable indentation but does not define scope boundaries

## Acceptance Criteria

1. All 9 stories in the epic are marked as Completed
2. Spreaders (Increment 1) can be declared, parsed, analyzed, and synthesized into a working executable with manual spreading using equal-proportion distribution
3. Metrics (Increment 2) automatically apply spreaders when dimension mismatches are detected, producing correct spread results for single unambiguous sequences
4. Spreader expressions (Increment 3) can reference metrics by name; implicit aggregator resolution produces correct weighted distributions
5. Mixed aggregator/spreader sequences (Increment 4) are correctly selected by the optimization algorithm and produce identical results to any valid ordering
6. All examples in `parse_and_analyze_only/` and `with_data/` directories pass validation
7. Error cases produce meaningful compile-time or runtime errors

## Notes

- Spreaders reuse the `tensorops.ExpandDimension` function from Epic007 Increment 1 — no new matrix math infrastructure
- The spreader FunctionRegistry is separate from the aggregator FunctionRegistry despite sharing many function names, because the dimensional unit semantics differ in the spreading context
- Metric reference resolution (Increment 3) is the most complex analysis-phase feature — implicit aggregator determination requires solving `raw_units * factor = required_units` for `factor`
- The `from []` (scalar spreading) case is a special edge case handled in both analysis and synthesis
- This epic depends heavily on Epic007 (aggregators) being sufficiently advanced; coordination across epics is critical
