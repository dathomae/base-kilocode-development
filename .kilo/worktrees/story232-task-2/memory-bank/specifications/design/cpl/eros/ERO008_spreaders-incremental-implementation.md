# ERO008: Spreaders Incremental Implementation

## Title

Spreaders Incremental Implementation

## Purpose/Goal

Define the incremental approach for implementing spreaders functionality across all CPL phases (parsing, analysis, synthesis), detailing the minimal functionality needed to support spreader declarations with dimensional transformation patterns, expression-based distribution operations, and dimensional unit validation, with integration into the metric dependency DAG. Each increment produces a working end-to-end executable.

## Decision Status

Accepted

## Implementation Status

Increment 1 Implemented — parser, analysis, CIL generation, and examples complete. Increments 2–4 deferred.

**Related Epics:**

- [Epic008_cpl-spreaders-implementation.md](../../../epics/Epic008_cpl-spreaders-implementation.md)

## Context

### Situation

Spreaders are a fundamental building block of the CPL. They transform metrics from less granular dimensional representations to more granular ones — for example, distributing a metric broken out by `[Product]` to `[Country, Product]` by applying a weighted distribution across countries. Spreaders are the inverse operation of aggregators: where aggregators collapse a dimension, spreaders introduce one.

Spreaders are **unnamed** — they are defined by their `from [<concepts>] to [<concepts>]` dimensional pattern and are matched at analysis time against the dimensional requirements of the metric DAG. Each spreader expands the tensor's dimensionality by exactly one.

The spreader implementation must span all four CPL phases: parsing (syntax recognition), analysis (dimensional unit validation and spreader map assembly), CIL generation (transformation matrix construction), and synthesis (generation of Go code for dimension-expanding matrix multiplication on tensors). Unlike aggregators, spreaders do not require a standalone matrix multiplication infrastructure increment — the tensor matrix multiplication infrastructure built in ERO007 (Aggregators) Increment 1 supports both dimension-reducing and dimension-expanding operations and is directly reusable by spreaders.

Spreaders have additional complexity not present in aggregators: their expressions may reference other metrics (for weighted distributions), requiring the analysis phase to perform implicit aggregator resolution — determining the correct dimensional form of the referenced metric and applying implicit aggregators as needed.

### Constraints

- The implementation must integrate with the existing ANTLR-based parser infrastructure
- Must follow the [Symbol Table-Driven Parsing Pattern](../../../patterns/symbol-table-driven-parsing.md) for parsing
- Must follow the [Postfix Expression Parsing Pattern](../../../patterns/postfix-expression-parsing.md) for spreader expression handling
- Parse/analysis separation must be maintained: the parser captures syntax only; the analyzer performs semantic validation
- Each increment must produce a working executable (parse → analyze → synthesize) with test data
- Spreaders must expand the tensor by exactly one dimension per spreader (`from` may be `[]`, in which case `to` must contain exactly one concept; `to` must contain all concepts from `from` plus exactly one additional concept)
- Multi-dimensional expansions must be composed from sequences of single-dimension spreaders
- Tensor matrix multiplication infrastructure (dimension-reducing and dimension-expanding) is already available from ERO007 Increment 1
- Spreader expressions may reference other metrics; implicit aggregator resolution is required when the referenced metric's dimensional form doesn't match the spreader's requirements
- Types are initially represented as `float64` (detailed type selection design is deferred to [type-selection.md](../type-selection.md))

### Requirements

- Parse `spreaders` sections with `from [<concepts>] to [<concepts plus one more>]` pattern (or `from [] to [<C>]` for scalar expansion) and expression body (see [aggregators-and-spreaders-syntax.md](../../requirements/cpl/aggregators-and-spreaders-syntax.md))
- Build raw `ParserSpreaderDefinition` structures stored in the symbol table under synthetic keys derived from the spreader's dimensional pattern
- Walk the symbol table during analysis to assemble the `SpreaderInfo` map, indexed by `(from_concept_set, to_concept_set)`
- Implement dimensional unit analysis for spreaders: compute the dimensional unit effect of the spreader expression and verify `input_units * expression_units = output_units`. The flow is reversed from aggregators — spreaders introduce a dimension, so the expression must provide the dimensional unit factor for the new dimension.
- Implement cell unit validation alongside dimensional unit validation
- Generate pre-built `SpreadingMatrix` instances from concept value sets during CIL generation
- When dimension mismatches are detected during metric DAG analysis (output has more dimensions than input), query the Spreaders map and insert `CILSpreadNode` nodes into affected metric CIL trees
- Support metric references in spreader expressions: when a spreader expression references a metric, perform implicit aggregator resolution to determine the required dimensional form
- Implement the spreader sequence selection algorithm: least-information-loss → minimal-work → arbitrary
- Support composable spreader sequences and interleaved aggregator/spreader sequences
- Support lookups referenced in spreader expressions (once lookups are implemented — see ERO006)

### Assumptions

- Concepts, types, units, measures, connectors, and metrics are already implemented (Epics 001-005 provide the foundation)
- Aggregator ERO (ERO007) Increment 1 is complete, providing the tensor matrix multiplication infrastructure
- The ANTLR grammar can be extended with spreader section parsing
- Expression parsing reuses the postfix expression pattern established for metrics
- Concept value sets are known at compile time, enabling pre-computation of spreading matrices
- Lookups (ERO006) are available when needed for spreader expression references

## Evaluation Criteria

1. **End-to-End Completeness** (High importance): Each increment must produce a working executable that applies spreading transformations to metric data.
2. **Incremental Value** (High importance): Each increment must add usable functionality; no increment should be purely scaffolding.
3. **Parse/Analysis Separation** (High importance): The parser must not perform semantic validation; all dimensional unit analysis and spreader selection must happen in analysis.
4. **Dimensional Correctness** (High importance): The dimensional unit validation must correctly enforce that spreader expressions produce the required dimensional transformation.
5. **Testability** (Medium importance): Each increment must include working examples with CPL files, IO specifications, and test data, covering both positive cases and error cases.

## Options Considered

### Option 1: Implement Spreaders After Full Aggregator Completion

**Description:**
Delay spreader implementation until aggregators (ERO007) are fully complete (all 4 increments). This ensures the matrix multiplication infrastructure, aggregator map, and selection algorithms are all in place before spreaders begin.

**Pros:**
- Maximum reuse of finished infrastructure
- No risk of coordinating across in-progress EROs
- Spreader implementation can focus on spreader-specific concerns (metric references, implicit aggregators)

**Cons:**
- Spreaders are blocked on all 4 aggregator increments, significantly delaying spreader work
- May discover aggregator design issues during spreader implementation that could have been caught earlier
- Longer total timeline

### Option 2: Start Spreaders After ERO007 Increment 2 (Matrix Multiplication + Aggregator Parsing)

**Description:**
Begin spreader implementation after ERO007 Increment 2 is complete — at this point, matrix multiplication infrastructure exists and aggregator parsing/validation is done, but aggregator selection/integration may still be in progress. Spreaders reuse the matrix infrastructure and follow the same parsing pattern.

**Pros:**
- Matrix multiplication (the key dependency) is available
- Spreader and aggregator work can proceed in parallel after the shared foundation is built
- Earlier detection of cross-cutting design issues
- Follows the natural dependency order: foundations → both aggregators and spreaders can proceed

**Cons:**
- If aggregator selection/integration (ERO007 Increment 3) reveals design changes, spreaders may need to adapt
- Moderate coordination cost between EROs 007 and 008

### Option 3: Start Spreaders After ERO007 Increment 1 (Matrix Multiplication Only)

**Description:**
Begin spreader implementation immediately after matrix multiplication infrastructure is built. Spreaders are the first consumer of the dimension-expanding matrix operations, validating that the matrix infrastructure works for both reduction and expansion.

**Pros:**
- Earliest possible start for spreader work
- Validates the matrix multiplication infrastructure for both use cases early
- Spreader parsing can proceed without waiting for aggregator parsing

**Cons:**
- Aggregator parsing patterns and lessons aren't available to inform spreader parsing design
- Higher risk of diverging from aggregator patterns
- Spreader Increment 3 (metric references with implicit aggregator resolution) requires aggregator infrastructure to exist

## Decision

**Chosen Option:** Option 2 — Start Spreaders After ERO007 Increment 2

**Rationale:**
ERO007 Increment 2 establishes the parsing and validation patterns for dimensional transformers — the `from [C1,C2] to [C2]` pattern, postfix expressions, dimensional unit computation, and the lookup map structure. Starting spreaders after this is complete ensures spreader parsing follows the same established patterns, avoiding divergence.

Matrix multiplication is already in place (from ERO007 Increment 1), so spreader synthesis is immediately possible. The main risk — that aggregator selection/integration might evolve — is mitigated by the fact that spreader Increment 2 (selection and integration) follows the same "single unambiguous sequence" approach as aggregator Increment 3, so any pattern changes can be applied consistently.

Spreader Increment 3 (metric references with implicit aggregator resolution) depends on aggregator infrastructure for the resolution step. This is the only increment that requires aggregator ERO007 to be at Increment 2 or higher, which aligns naturally with the progression.

## Implications

### Positive Implications

- Reuses the tensor matrix multiplication infrastructure from ERO007 Increment 1
- Follows the same parsing and validation patterns established by aggregators
- Spreaders and aggregators can be developed with shared understanding of the dimensional unit model
- Independent testability of the dimension-expanding matrix operations via spreader tests
- No standalone infrastructure increment needed — spreaders start with a CPL pipeline increment

### Negative Implications / Trade-offs

- Spreaders depend on ERO007 Increment 2 being complete before development can begin
- Implicit aggregator resolution (Increment 3) is complex and introduces a dependency on aggregator infrastructure
- Four increments with increasing complexity in later increments
- Testing weighted distributions with metric references requires test data for multiple metrics

### Impact on System Qualities

| Quality Attribute | Impact | Notes |
| ----------------- | ------ | ----- |
| Correctness       | +      | Dimensional unit validation catches spreader misuse at compile time |
| Performance       | =      | Spread operations reuse the same matrix multiplication as aggregators |
| Maintainability   | +      | Consistent patterns with aggregators reduce cognitive load |
| Testability       | +      | Each increment includes both positive and negative test cases |
| Complexity        | -      | Implicit aggregator resolution adds significant analysis-phase complexity |

### Future Considerations

- Additional distribution strategies beyond equal-proportion and metric-weighted (historical proportion, seasonal patterns)
- Spreader expressions that reference lookups in addition to metrics
- Spreader caching for repeated pattern queries during analysis
- Runtime adaptation of spreading matrices when concept values or distribution data change
- Fused spreader/aggregator sequences for optimal performance

## Notes

### Increment Sketches

Each increment is end-to-end across parse → analyze → CIL generation → CIL walk → synthesize. Every increment produces a working executable with test data. See [Spreaders Design](../spreaders.md) for full design descriptions. The tensor matrix multiplication infrastructure (both dimension-reducing and dimension-expanding) is already available from ERO007 Increment 1.

**Increment 1: Basic Spreader Parsing and Validation**

- _Parser:_ New `CPLSpreaders.g4` grammar file with `spreadersSection` and `spreaderDeclaration` rules, recognizing `from [<concepts>] to [<concepts>]` (or `from [] to [<C>]`) pattern and expression body. New `SpreadersListener` that pushes `ParserSpreaderDefinition` onto the symbol table under synthetic keys. Spreader expressions converted to postfix notation using the standard postfix expression parsing pattern. Expression supports concept references and operators (`*`, `/`, `1/size(C)`, etc.).
- _Analyzer:_ New `AnalyzeSpreadersScope` function registered with `WalkSymbolTable`. Assembles `SpreaderInfo` entries into the Spreaders lookup map, keyed by `(from_concept_set, to_concept_set)`. Validates concept references in `from`, `to`, and expression. Implements dimensional unit computation for expression components and validates dimensional unit compatibility for each spreader. Implements cell unit validation.
- _CIL Generation:_ Generate `CILSpreadNode` with pre-computed `SpreadingMatrix` for each valid spreader. The matrix is constructed from concept value sets: rows correspond to source concept values, columns correspond to the new concept values being introduced, entries are distribution weights as determined by the spreader expression.
- _Synthesis:_ CIL tree walker handles `CILSpreadNode`, emitting Go code that applies the spreading matrix to the source tensor using the existing matrix multiplication infrastructure from ERO007 Increment 1 (dimension-expanding mode). Each spreader produces an output tensor with one additional dimension (or rank one for `from []`). Manual application in test code (not yet integrated into metric analysis).
- _Examples:_
  - `parse_and_analyze_only/without_errors/` — CPL files with valid spreader declarations (single-dimension expansion, `from []`, equal-proportion distribution via `1/size(C)`)
  - `parse_and_analyze_only/with_errors/` — CPL files with invalid spreaders (missing `to` concepts, `to` doesn't contain all `from` concepts, `to` doesn't add exactly one concept, non-existent concept references, incompatible dimensional units)
  - `with_data/spreaders_basic/` — CPL + IO spec + parquet data + `.expected.json` for a manually applied spreader (e.g., evenly distributing `[Product]` sales targets to `[Country, Product]` using `1/size(Country) * Country`)

**Increment 2: Spreader Selection and Integration**

- _Parser:_ No changes — spreader declarations already parsed in Increment 1.
- _Analyze:_ After the full metric dependency DAG is assembled (second analysis pass), detect dimension mismatches where input has fewer dimensions than output. Query the Spreaders map for applicable transformations using the dimensional pattern key. Identify a single valid spreader sequence that starts with the available dimensions (ultimately anchored in the dimensions of connector inputs) and results in the target dimensions specified in the IO spec. At this increment, only cases where exactly one possible application sequence exists are supported — if multiple sequences could satisfy the dimensional requirements, this is deferred to Increment 4. Insert `CILSpreadNode` nodes into affected metric CIL trees between the input reference and the calculation body.
- _Synthesis:_ CIL tree walker emits Go code that applies spreader sequences as part of metric computation. The generated code applies each spreader's matrix multiplication in order, producing intermediate tensors at each step.
- _Examples:_
  - `parse_and_analyze_only/without_errors/` — CPL files with metrics that require spreading from coarser inputs to finer outputs, where exactly one spreader sequence exists
  - `parse_and_analyze_only/with_errors/` — CPL files where a dimension mismatch exists but no applicable spreader bridges the gap
  - `with_data/spreaders_integration/` — CPL + IO spec + parquet data + `.expected.json` for metrics composed from inputs at different dimensional granularities, with exactly one valid spreading path (e.g., input `[Product]`, output `[Country, Product]`, one spreader matching `from [Product] to [Country, Product]`)

**Increment 3: Metric References in Spreader Expressions**

- _Parser:_ No changes — the expression parsing infrastructure already supports identifiers; metric references are parsed as identifiers and resolved during analysis.
- _Analyze:_ When a spreader expression references a metric by name, resolve it against the Metrics map. Determine the dimensional form of the referenced metric. If the raw metric's dimensional form doesn't match the spreader's requirements, determine whether an implicit aggregator can bridge the gap (e.g., collapsing a `[Country]` metric to a scalar using `from [Country] to [] = 1/sum(Country)`). Validate that the required aggregator exists in the Aggregators map. Record the resolved metric reference and any implicit `CILAggregateNode` in the spreader's `CILSpreadNode`.
- _Synthesis:_ CIL tree walker emits Go code that computes referenced metrics (including any implicit aggregations) before applying the spread. The generated code ensures referenced metrics are available at the point of spreading. Cell unit compatibility is verified between the referenced metric's cell type and the spreader's cell type.
- _Examples:_
  - `with_data/spreaders_metric_references/` — CPL + IO spec + parquet data + `.expected.json` for a spreader that references a metric for weighted distribution (e.g., spreading `[Product]` sales targets to `[Country, Product]` using `Product / customers`, where `customers` is a `[Country]` metric that requires an implicit `from [Country] to [] = 1/sum(Country)` aggregator)
- _Prerequisite:_ Aggregator ERO (ERO007) must be at Increment 2 or higher to provide the Aggregators map for implicit resolution.

**Increment 4: Spreader/Aggregator Interleaving and Optimization**

- _Parser:_ No changes.
- _Analyze:_ Implement the full selection algorithm for mixed aggregator/spreader sequences:
  1. Least-information-loss principle: prefer sequences that preserve more dimension information (prefer spreading as late as possible and aggregating as early as possible)
  2. Minimal-work principle: if multiple sequences are equally information-preserving, select the one that requires the least computational work
  3. If still equivalent, select arbitrarily but deterministically
  - Implement specificity comparison for ambiguous spreader resolution (more concepts in the `from` set = more specific).
- _Synthesis:_ Optimize multi-step sequences into fused operations where possible (e.g., combine adjacent spreading matrices into a single matrix multiplication, or combine an aggregation followed by a spreading of a different dimension). Handle edge cases: empty `from` (scalar origin), single-dimension metrics, sequences that interleave aggregators and spreaders.
- _Examples:_
  - `with_data/spreaders_optimization/` — CPL + IO spec + parquet data + `.expected.json` for multi-step sequences that exercise the selection algorithm (e.g., input `[Product]`, output `[Country, Region, Product]`, framework selects optimal expansion order and optimizes)
  - Test cases with ambiguous spreader definitions to verify selection correctness
  - Test cases with interleaved aggregator/spreader sequences

### Phase-Specific EROs

- Parsing Phase ERO: [ERO008_spreaders-parsing-phase.md](../go-parser-target/eros/ERO008_spreaders-parsing-phase.md)
- Analysis Phase ERO: [ERO008_spreaders-analysis-phase.md](../go-analysis/eros/ERO008_spreaders-analysis-phase.md)
- Synthesis Phase ERO: [ERO008_spreaders-synthesis-phase.md](../go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md)

### Related Documents

- [Spreaders Design](../spreaders.md) — Full spreaders design with core rules, dimensional unit analysis, CIL node types, and 4 end-to-end increments
- [Aggregators Design](../aggregators.md) — Aggregator design (symmetric inverse operation, shared matrix multiplication infrastructure)
- [CPL Requirements — Aggregators and Spreaders](../../requirements/cpl/aggregators-and-spreaders-syntax.md) — Syntax requirements for aggregators and spreaders
- [CPL Requirements — Concepts, Types, Units, Measures](../../requirements/cpl/concepts-types-units-measures-syntax.md) — Concept hierarchy (spreaders distribute across concept values)
- [Lookups Design](../lookups.md) — Lookup design (spreader expressions may reference lookups)
- [Metrics Design](../metrics-design.md) — Metric design (spreaders are applied to metric inputs during analysis)
- [Concept Definitions Design](../concepts-definitions.md) — Concept design (spreader matrices are built from concept value sets)
- [Units and Measures Design](../units-and-measures.md) — Unit/measure design (dimensional unit algebra)
- [Postfix Expression Parsing Pattern](../../../patterns/postfix-expression-parsing.md) — Pattern for expression handling
- [Symbol Table-Driven Parsing Pattern](../../../patterns/symbol-table-driven-parsing.md) — Architectural pattern for parsing phase design
- [CPL Architecture — Intermediate Language](../../architecture/cpl/analysis/intermediate_language.md) — CIL specification
