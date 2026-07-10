# Epic006: CPL Lookups Implementation

## Title

CPL Lookups Implementation

## Ultimate Goal

Implement the complete lookups functionality for the Chronocone Planning Language across all processing phases (parsing, analysis, synthesis), delivering lookup declarations with dimensions, optional grain, and celltype, integration with metric calculations, and support for lookup references in aggregator and spreader expressions. Each increment produces a working end-to-end executable with test data and examples.

## References

- [Lookups Design](../specifications/design/cpl/lookups.md) — Full lookups design with definition types and data flow
- [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md) — Parent ERO defining the 5-increment strategy
- [ERO006: Lookups Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO006_lookups-parsing-phase.md) — Parsing phase data structures and grammar
- [ERO006: Lookups Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md) — Analysis phase data structures and validation
- [ERO006: Lookups Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md) — Synthesis phase code generation
- [CPL Requirements — Lookups](../specifications/requirements/cpl/lookups-syntax.md) — Syntax requirements for lookups
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md) — Concept design (lookup dimensions reference concepts)
- [Units and Measures Design](../specifications/design/cpl/units-and-measures.md) — Unit/measure design (lookup celltypes are measures)
- [Symbol Table-Driven Parsing Pattern](../patterns/symbol-table-driven-parsing.md) — Architectural pattern for parsing

## Dependencies

- [Epic001_cpl-concepts-implementation.md](Epic001_cpl-concepts-implementation.md) — Concepts must be implemented (lookup dimensions reference concepts)
- [Epic003_units-and-measures-implementation.md](Epic003_units-and-measures-implementation.md) — Units and measures must be implemented (lookup celltypes are measures)
- [Epic005_cpl-metrics-implementation.md](Epic005_cpl-metrics-implementation.md) — Metrics must be implemented for Increment 3 (lookup input kind)
- [Epic007_cpl-aggregators-implementation.md](Epic007_cpl-aggregators-implementation.md) — Aggregators needed for Increment 4 (lookups in aggregator expressions) (not yet defined)
- [Epic008_cpl-spreaders-implementation.md](Epic008_cpl-spreaders-implementation.md) — Spreaders needed for Increment 5 (lookups in spreader expressions) (not yet defined)

## Stories

| Story | Status | Description | ERO |
|-------|--------|-------------|-----|
| StoryXXX_lookups-increment1-parse.md (not yet defined) | Not Started | Implement CPLLookups.g4 grammar, LookupsListener, ParserLookupDefinition, lexer tokens, symbol table storage for timeless lookups | [ERO006_lookups-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO006_lookups-parsing-phase.md) |
| StoryXXX_lookups-increment1-analyze.md (not yet defined) | Not Started | Implement AnalyzeLookupScope, LookupInfo, CILLookupNode, cross-reference resolution, name uniqueness validation for timeless lookups | [ERO006_lookups-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md) |
| StoryXXX_lookups-increment1-synthesize.md (not yet defined) | Not Started | Implement CIL tree walker for CILLookupNode, Go code generation for MetricComputation registration, parquet loading, timeless lookup examples | [ERO006_lookups-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md) |
| StoryXXX_lookups-increment2-parse.md (not yet defined) | Not Started | Extend grammar with optional grain clause, add Grain and HasGrain fields to ParserLookupDefinition, modify listener for grain extraction | [ERO006_lookups-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO006_lookups-parsing-phase.md) |
| StoryXXX_lookups-increment2-analyze.md (not yet defined) | Not Started | Implement grain validation (Interval type constraint, duplicate check, empty check), extend LookupInfo and CILLookupNode with grain fields | [ERO006_lookups-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md) |
| StoryXXX_lookups-increment2-synthesize.md (not yet defined) | Not Started | Generate Go code for temporal lookup loading, tensor indexing with grain dimensions, multi-grain resolution, temporal lookup examples | [ERO006_lookups-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md) |
| StoryXXX_lookups-increment3-parse.md (not yet defined) | Not Started | Extend metrics grammar with lookup input kind, add LOOKUP token, modify inputDeclaration rule and MetricsListener | [ERO006_lookups-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO006_lookups-parsing-phase.md) |
| StoryXXX_lookups-increment3-analyze.md (not yet defined) | Not Started | Implement lookup reference resolution in metric inputs, CILLookupRefNode, dimensional compatibility validation, lookup input ExampleProject | [ERO006_lookups-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md) |
| StoryXXX_lookups-increment3-synthesize.md (not yet defined) | Not Started | Implement CIL tree walker for CILLookupRefNode, GetMetricComputation(lookupName) emission, unit compatibility verification, metric+lookup examples | [ERO006_lookups-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md) |
| StoryXXX_lookups-increment4-analyze.md (not yet defined) | Not Started | Implement lookup reference resolution in aggregator expressions, dimensional unit validation for aggregator-lookup combinations | [ERO006_lookups-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md) |
| StoryXXX_lookups-increment4-synthesize.md (not yet defined) | Not Started | Generate lookup access code within aggregator computation, ensure lookup loading before aggregator execution, aggregator+lookup examples | [ERO006_lookups-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md) |
| StoryXXX_lookups-increment5-analyze.md (not yet defined) | Not Started | Implement lookup reference resolution in spreader expressions, implicit aggregator resolution for lookups, dimensional/cell unit validation | [ERO006_lookups-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md) |
| StoryXXX_lookups-increment5-synthesize.md (not yet defined) | Not Started | Generate lookup access code within spreader computation, implicit aggregator execution for lookups, spreader+lookup examples | [ERO006_lookups-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md) |

## Story Ordering and Dependencies

Stories within this epic are arranged by increment, and within each increment by phase (parse → analyze → synthesize). This ordering is strict:

**Increment 1 (Basic Lookup Declarations):**
1. `lookups-increment1-parse` — No dependencies on other lookup stories. Depends on Epic001 (concepts), Epic003 (units/measures).
2. `lookups-increment1-analyze` — Depends on `lookups-increment1-parse` (needs parser definitions).
3. `lookups-increment1-synthesize` — Depends on `lookups-increment1-analyze` (needs CIL nodes). Depends on Epic005 (metrics) Increment 1 for CIL tree walking infrastructure.

**Increment 2 (Lookups with Grain):**
4. `lookups-increment2-parse` — Depends on `lookups-increment1-parse` (extends existing grammar).
5. `lookups-increment2-analyze` — Depends on `lookups-increment2-parse` and `lookups-increment1-analyze`.
6. `lookups-increment2-synthesize` — Depends on `lookups-increment2-analyze` and `lookups-increment1-synthesize`.

**Increment 3 (Lookup Integration with Metrics):**
7. `lookups-increment3-parse` — Depends on `lookups-increment1-parse`. Depends on Epic005 (metrics) for CPLMetrics.g4 modification.
8. `lookups-increment3-analyze` — Depends on `lookups-increment3-parse` and `lookups-increment1-analyze`.
9. `lookups-increment3-synthesize` — Depends on `lookups-increment3-analyze` and `lookups-increment1-synthesize`.

**Increment 4 (Lookups in Aggregator Expressions):**
10. `lookups-increment4-analyze` — Depends on `lookups-increment1-analyze`. Depends on Epic007 (aggregators) Increment 2 for Aggregators map.
11. `lookups-increment4-synthesize` — Depends on `lookups-increment4-analyze`, `lookups-increment1-synthesize`. Depends on Epic007 Increment 2.

**Increment 5 (Lookups in Spreader Expressions):**
12. `lookups-increment5-analyze` — Depends on `lookups-increment1-analyze`. Depends on Epic008 (spreaders) Increment 3 for spreader analysis infrastructure.
13. `lookups-increment5-synthesize` — Depends on `lookups-increment5-analyze`, `lookups-increment1-synthesize`. Depends on Epic008 Increment 3.

No parser stories exist for Increments 4 and 5 — lookup references in aggregator/spreader expressions are resolved during analysis, not parsed as distinct syntax.

Stories from Epic005 (metrics) and Epic007 (aggregators) may be interleaved with these stories. Increment 3 requires coordination with the metrics epic; Increments 4-5 require coordination with the aggregator and spreader epics (not yet defined).

## Constraints

1. Must follow the [Symbol Table-Driven Parsing Pattern](../patterns/symbol-table-driven-parsing.md) — strict parse/analysis separation
2. Must follow the [Postfix Expression Parsing Pattern](../patterns/postfix-expression-parsing.md) for expression handling (relevant for metric expressions that reference lookups)
3. Each increment must produce a working executable with parse → analyze → synthesize pipeline
4. Each increment must include examples: `parse_and_analyze_only/` (with and without errors) and `with_data/` (CPL + IO spec + parquet + expectations)
5. Lookup data uses the same `MetricComputation` interface as connectors — reuse existing controller infrastructure
6. Types initially represented as `float64` (detailed type selection deferred to [type-selection.md](../specifications/design/cpl/type-selection.md))
7. The `indent` token enforces readable indentation but does not define scope boundaries (consistent with existing CPL grammar)
8. Increments 4 and 5 are blocked until aggregator/spreader epics provide the necessary infrastructure

## Acceptance Criteria

1. All 13 stories in the epic are marked as Completed
2. Timeless lookups (Increment 1) can be declared, parsed, analyzed, and synthesized into a working executable that loads lookup data
3. Temporal lookups (Increment 2) support optional grain with multiple measures and correct time-based indexing
4. Metrics (Increment 3) can reference lookups via the `lookup` input kind and access lookup values via `GetMetricComputation`
5. Aggregators (Increment 4) can reference lookups in their expressions with dimensional unit validation
6. Spreaders (Increment 5) can reference lookups in their expressions with implicit aggregator resolution
7. All examples in `parse_and_analyze_only/` and `with_data/` directories pass validation
8. Error cases produce meaningful compile-time or runtime errors (no silent failures)

## Notes

- The lookup IO specification schema is defined separately from the connector IO spec; unification may occur later if they prove structurally identical
- Default value declarations for key-miss handling are deferred — missing keys are runtime errors
- Dynamic reload of temporal lookups when source data changes is a future consideration
- This epic depends on Epics 001-005 being complete or sufficiently advanced; Increments 4-5 additionally depend on Epics 007-008
