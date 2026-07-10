# ERO006: Lookups Incremental Implementation

## Title

Lookups Incremental Implementation

## Purpose/Goal

Define the incremental approach for implementing lookups functionality across all CPL phases (parsing, analysis, synthesis), detailing the minimal functionality needed to support lookup declarations with dimensions, optional grain, and celltype, followed by integration with metrics, aggregators, and spreaders. Each increment produces a working end-to-end executable.

## Decision Status

Draft

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic006_cpl-lookups-implementation.md](../../../epics/Epic006_cpl-lookups-implementation.md)

## Context

### Situation

Lookups are a fundamental building block of the CPL. They provide "side data" to metric calculations — reference data that maps input dimensions (from concepts and optionally time grains) to values. Lookups are used for items such as conversion factors, weighting tables, and dimension mappings.

Lookups are similar to connectors in that they supply data to the computation framework, but they differ in three important ways: (1) lookups are expected to change more slowly than connector data; (2) lookups often have no time dimension; and (3) lookup dimensions serve as keys into the lookup table rather than describing an output shape.

The lookups implementation must span all four CPL phases: parsing (syntax recognition), analysis (semantic validation and CIL generation), and synthesis (generation of executable Go code that loads and registers lookup data). Beyond the core lookup declaration, lookups also participate in aggregator expressions (as referenced metrics in weighted distributions) and spreader expressions (as distribution data sources).

### Constraints

- The implementation must integrate with the existing ANTLR-based parser infrastructure
- Must follow the [Symbol Table-Driven Parsing Pattern](../../../patterns/symbol-table-driven-parsing.md) for parsing
- Must follow the [Postfix Expression Parsing Pattern](../../../patterns/postfix-expression-parsing.md) for expression handling (since aggregator and spreader expressions that reference lookups use the same expression mechanism)
- Parse/analysis separation must be maintained: the parser captures syntax only; the analyzer performs semantic validation
- Each increment must produce a working executable (parse → analyze → synthesize) with test data
- Lookup data at runtime is registered as `MetricComputation` objects on the `ChronoconeComputationController`, consistent with the existing connector integration model
- Types are initially represented as `float64` (detailed type selection design is deferred to [type-selection.md](../type-selection.md))

### Requirements

- Parse `lookups` sections and lookup declarations with `dimensions`, optional `grain`, and `celltype` clauses (see [lookups-syntax.md](../../requirements/cpl/lookups-syntax.md))
- Build raw `ParserLookupDefinition` structures stored in `SymbolDefinition.Data` during parsing
- Walk the symbol table tree during analysis to assemble `LookupInfo` entries in `AnalysisDefinitions`
- Generate one `CILLookupNode` per lookup (leaf data source node in the CIL tree)
- Walk CIL lookup nodes to generate Go code that loads lookup data from the Lookup IO Specification and registers it as a `MetricComputation`
- Support Lookup IO Specifications that map lookup names to file or database sources
- Support lookup references in metric `inputs` blocks via the `lookup` input kind
- Support lookup references in aggregator and spreader expressions (as referenced data sources for weighted distributions)
- Dimension, grain, and celltype cross-reference validation at analysis time

### Assumptions

- Concepts, types, units, measures, connectors, and metrics are already implemented (Epics 001-005 provide the foundation)
- The ANTLR grammar can be extended with lookup section parsing
- The `MetricComputation` registration mechanism used by connectors can be reused for lookups
- The CIL tree walking infrastructure can be reused from existing connector and metric synthesis
- Aggregator and spreader implementation (to be defined in EROs 007 and 008) will support lookup references in their expressions
- Example generation includes both `parse_and_analyze_only` (CPL files only, with and without errors) and `with_data` (CPL + IO spec + parquet data + expectations)

## Evaluation Criteria

1. **End-to-End Completeness** (High importance): Each increment must produce a working executable that reads CPL, IO spec, and data, then outputs computed metric values that incorporate lookup data.
2. **Incremental Value** (High importance): Each increment must add usable functionality; no increment should be purely scaffolding.
3. **Parse/Analysis Separation** (High importance): The parser must not perform semantic validation; all cross-references and dependency resolution must happen in analysis.
4. **Consistency** (Medium importance): Data structures and patterns should evolve consistently with the existing concepts, types, units, connectors, and metrics implementations.
5. **Testability** (Medium importance): Each increment must include working examples with CPL files, IO specifications, and test data, covering both positive cases and error cases.

## Options Considered

### Option 1: Exclude Aggregator/Spreader Integration from Lookups ERO

**Description:**
Implement lookups as a standalone declaration type with metric integration only. Defer aggregator and spreader integration to the aggregator/spreader EROs (007 and 008).

**Pros:**
- Simpler ERO with fewer increments
- Cleaner separation of concerns — each ERO handles its own feature
- Lookups can be tested and verified independently before being consumed by aggregators/spreaders

**Cons:**
- The most interesting use cases for lookups (weighted distributions in spreaders, indirect aggregator inputs) are deferred
- Risk of building lookup infrastructure that later needs significant refactoring to support aggregator/spreader usage
- May miss early detection of design issues in how lookups interact with dimensional transformations

### Option 2: Include Aggregator/Spreader Integration in this ERO

**Description:**
Define lookups with full integration into metrics, aggregators, and spreaders within a single ERO, covering all use cases the design doc anticipates.

**Pros:**
- Complete picture of how lookups will be used across the CPL
- Early validation of the lookup design against the most demanding use cases
- Avoids later EROs having to back-patch lookup integration decisions

**Cons:**
- More increments, longer timeline before "complete"
- Dependency on aggregator/spreader designs being finalized (which are in-progress)
- Increments 4 and 5 cannot be implemented until aggregator/spreader infrastructure exists

### Option 3: Hybrid — Core lookups in this ERO, integration increments delivered alongside aggregator/spreader EROs

**Description:**
Increments 1-3 (core lookups) defined and delivered in this ERO. Increments 4-5 (aggregator/spreader integration) defined here but scheduled for delivery in coordination with EROs 007 and 008.

**Pros:**
- Core lookup functionality available early for metric-level use cases
- Integration increments are defined and scoped, avoiding discovery work later
- Natural dependency ordering — aggregator/spreader infrastructure must exist before lookups can be used within them

**Cons:**
- Some ERO content is deferred; risk of schedule misalignment between EROs
- More complex coordination across multiple EROs

## Decision

**Chosen Option:** Option 2 — Include Aggregator/Spreader Integration in this ERO

**Rationale:**
The full value of lookups is realized when they participate in aggregator and spreader expressions — conversion factor tables referenced by spreaders, weighting tables used in aggregations, etc. Defining these integration increments now ensures the lookup design is validated end-to-end against all intended use cases and avoids later EROs having to back-patch decisions made in isolation.

While Increments 4 and 5 depend on aggregator/spreader infrastructure (to be defined in EROs 007 and 008), defining them in this ERO establishes the complete scope of lookup implementation and ensures the aggregator/spreader designs account for lookup integration from the start.

## Implications

### Positive Implications

- Complete specification of all lookup usage patterns, preventing future rework
- Early validation of dimensional unit analysis for expressions involving lookups
- Natural pressure on aggregator/spreader designs to support lookup references from day one
- Example generation covers all lookup scenarios, not just metric-level ones

### Negative Implications / Trade-offs

- Increments 4 and 5 are blocked until aggregator/spreader infrastructure exists
- Higher coordination cost across EROs 006, 007, and 008
- ERO is larger than it would be if scoped to core lookups only

### Impact on System Qualities

| Quality Attribute | Impact | Notes |
| ----------------- | ------ | ----- |
| Correctness       | +      | Dimensional unit validation catches lookup misuse at compile time |
| Testability       | +      | Each increment includes both positive and negative test cases |
| Maintainability   | =      | Consistent with existing incremental implementation pattern |
| Complexity        | -      | More increments than a minimal-scope ERO would require |

### Future Considerations

- Lookup IO specification schema may be unified with connector IO spec if they prove structurally identical
- Default value declarations for key-miss handling may be added based on user feedback
- Dynamic lookup data reload for temporal lookups when source data changes

## Notes

### Increment Sketches

Each increment is end-to-end across parse → analyze → CIL generation → CIL walk → synthesize. Every increment produces a working executable with test data. See [Lookups Design](../lookups.md) for full design descriptions.

**Increment 1: Basic Lookup Declarations Without Grain**

- _Parser:_ New `CPLLookups.g4` grammar file with `lookupsSection` and `lookupDeclaration` rules, recognizing `dimensions` and `celltype` clauses. New `LookupsListener` that pushes `ParserLookupDefinition` onto the symbol table. Dimension concept names and celltype measure name stored as raw strings (resolution deferred to analysis).
- _Analyzer:_ New `AnalyzeLookupsScope` function registered with `WalkSymbolTable`. Assembles `LookupInfo` entries from parser definitions. Validates that dimension concept names exist in the concept hierarchy. Validates that celltype measure name exists. Generates one `CILLookupNode` per lookup (leaf node capturing name, dimensions, and celltype).
- _Synthesis:_ CIL tree walker handles `CILLookupNode`, emitting Go code to load lookup data from the Lookup IO Specification and register it as a `MetricComputation` on the `ChronoconeComputationController`.
- _Examples:_
  - `parse_and_analyze_only/without_errors/` — CPL files with valid lookup declarations (timeless, single and multiple dimensions)
  - `parse_and_analyze_only/with_errors/` — CPL files with invalid lookup declarations (missing `dimensions`, missing `celltype`, referencing non-existent concepts/measures)
  - `with_data/lookups_basic/` — CPL + IO spec + parquet data + `.expected.json` for a timeless lookup with product category mapping

**Increment 2: Lookups with Grain**

- _Parser:_ Extended grammar for optional `grain` clause in lookup declarations. Parse one or more grain measure names separated by commas.
- _Analyzer:_ Validate that grain measure names exist and have types derived from `Interval`. Include grain dimensions in `CILLookupNode`. Generate `AggregationMatrix` for temporal lookups with multiple grain measures.
- _Synthesis:_ CIL tree walker emits Go code that loads temporal lookup data and indexes by time as well as concept dimensions. Handles multi-grain lookup resolution at runtime.
- _Examples:_
  - `parse_and_analyze_only/without_errors/` — CPL files with temporal lookups (single and multiple grain measures)
  - `parse_and_analyze_only/with_errors/` — CPL files with invalid grain (non-Interval types, non-existent measures)
  - `with_data/lookups_temporal/` — CPL + IO spec + parquet data + `.expected.json` for a temporal lookup with currency conversion rates keyed by week and month

**Increment 3: Lookup Integration with Metric Calculations**

- _Parser:_ Extended metric grammar to recognize `lookup` input kind in metric `inputs` blocks (alongside existing `metric` and `dimension` input kinds). Parse `lookup <name>` declarations.
- _Analyzer:_ When a metric input references a lookup, validate the lookup exists. Verify dimensional compatibility between the lookup's dimensions and the metric's dimensions. Generate `CILInputRef` nodes for lookup inputs in the metric's CIL tree.
- _Synthesis:_ CIL tree walker emits Go code that calls `controller.GetMetricComputation(lookupName)` for lookup inputs. Verifies unit compatibility between lookup celltypes and the calculations that use them.
- _Examples:_
  - `parse_and_analyze_only/with_errors/` — CPL files with invalid lookup references in metric inputs (non-existent lookup, incompatible dimensions)
  - `with_data/lookups_in_metrics/` — CPL + IO spec + parquet data + `.expected.json` for a metric that uses a lookup (e.g., applying a conversion factor to revenue)

**Increment 4: Lookups in Aggregator Expressions**

- _Parser:_ No changes — aggregator expression parsing already supports concept references and operators via the postfix expression pattern; lookup references appear as identifiers in aggregator expressions.
- _Analyzer:_ When an aggregator expression references a lookup by name, resolve it against the Lookups map. Validate dimensional unit compatibility between the lookup's dimensions and the aggregator's dimensional transformation. The lookup may need to be dimensionally transformed itself (via aggregators or spreaders on the lookup) before use in the aggregator expression. Record the resolved lookup reference in the aggregator's `CILAggregateNode`.
- _Synthesis:_ CIL tree walker emits Go code that accesses lookup data for the aggregator's referenced lookups. Ensures lookups referenced in aggregator expressions are loaded and available at the point of aggregation.
- _Examples:_
  - `with_data/lookups_in_aggregators/` — CPL + IO spec + parquet data + `.expected.json` for an aggregator that uses a lookup (e.g., applying a weighting table from a lookup during aggregation)
- _Prerequisite:_ Aggregator infrastructure (ERO007) must exist at Increment 2 level or above.

**Increment 5: Lookups in Spreader Expressions**

- _Parser:_ No changes — spreader expression parsing already supports metric references and concept references; lookup references appear as identifiers in spreader expressions alongside metric references.
- _Analyzer:_ When a spreader expression references a lookup by name, resolve it against the Lookups map. Apply the same dimensional unit analysis used for metric references in spreaders: determine the required dimensional form of the lookup, apply implicit aggregators/spreaders to the lookup if needed. Record the resolved lookup reference and any implicit transformation nodes in the spreader's `CILSpreadNode`.Validate dimensional and cell unit compatibility.
- _Synthesis:_ CIL tree walker emits Go code that accesses lookup data for the spreader's referenced lookups, including any implicit aggregations. Ensures lookups referenced in spreader expressions are loaded and available at the point of spreading.
- _Examples:_
  - `with_data/lookups_in_spreaders/` — CPL + IO spec + parquet data + `.expected.json` for a spreader that uses a lookup (e.g., spreading a global target across countries using a lookup-based weighting table)
- _Prerequisite:_ Spreader infrastructure (ERO008) must exist at Increment 3 level or above.

### Phase-Specific EROs

- Parsing Phase ERO: [ERO006_lookups-parsing-phase.md](../go-parser-target/eros/ERO006_lookups-parsing-phase.md)
- Analysis Phase ERO: [ERO006_lookups-analysis-phase.md](../go-analysis/eros/ERO006_lookups-analysis-phase.md)
- Synthesis Phase ERO: [ERO006_lookups-synthesis-phase.md](../go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md)

### Related Documents

- [Lookups Design](../lookups.md) — Full lookups design with core rules, CIL node types, and 3 core end-to-end increments
- [CPL Requirements — Lookups](../../requirements/cpl/lookups-syntax.md) — Syntax requirements for lookups
- [Aggregators Design](../aggregators.md) — Aggregator design (lookups referenced in aggregator expressions)
- [Spreaders Design](../spreaders.md) — Spreader design (lookups referenced in spreader expressions)
- [Metrics Design](../metrics-design.md) — Metric design (metrics reference lookups as inputs)
- [Concept Definitions Design](../concepts-definitions.md) — Concept design (lookup dimensions reference concepts)
- [Units and Measures Design](../units-and-measures.md) — Unit/measure design (lookup celltypes are measures)
- [Postfix Expression Parsing Pattern](../../../patterns/postfix-expression-parsing.md) — Pattern for expression handling
- [Symbol Table-Driven Parsing Pattern](../../../patterns/symbol-table-driven-parsing.md) — Architectural pattern for parsing phase design
- [CPL Architecture — Introduction](../../architecture/cpl/introduction.md) — Overall CPL processing lifecycle
