# ERO007: Aggregators Incremental Implementation

## Title

Aggregators Incremental Implementation

## Purpose/Goal

Define the incremental approach for implementing aggregators functionality across all CPL phases (parsing, analysis, synthesis), detailing the minimal functionality needed to support aggregator declarations with dimensional transformation patterns, expression-based aggregation operations, and dimensional unit validation, with integration into the metric dependency DAG. Each increment produces a working end-to-end executable.

## Decision Status

Draft

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic007_cpl-aggregators-implementation.md](../../../epics/Epic007_cpl-aggregators-implementation.md)

## Context

### Situation

Aggregators are a fundamental building block of the CPL. They transform metrics from more granular dimensional representations to less granular ones — for example, collapsing a metric broken out by `[Country, Product, Warehouse]` to `[Country, Product]` by summing data across the `Warehouse` dimension. Aggregators are the mechanism by which the CPL calculation framework makes the dimensions of an input metric match the dimensions of the metric that references it when the output is at a coarser level of detail.

Aggregators are **unnamed** — they are defined by their `from [<concepts>] to [<concept>]` dimensional pattern and are matched at analysis time against the dimensional requirements of the metric DAG. Each aggregator reduces the tensor's dimensionality by exactly one.

The aggregator implementation must span all four CPL phases: parsing (syntax recognition), analysis (dimensional unit validation and aggregator map assembly), CIL generation (transformation matrix construction), and synthesis (generation of Go code for full matrix multiplication on tensors). Aggregator synthesis requires a fundamental shift from element-wise tensor operations to dimension-changing matrix multiplication. However, the matrix multiplication infrastructure itself (Increment 1) is a standalone, CPL-independent piece of work — pure Go tensor math with no parsing, analysis, or CIL components.

### Constraints

- The implementation must integrate with the existing ANTLR-based parser infrastructure
- Must follow the [Symbol Table-Driven Parsing Pattern](../../../patterns/symbol-table-driven-parsing.md) for parsing
- Must follow the [Postfix Expression Parsing Pattern](../../../patterns/postfix-expression-parsing.md) for aggregator expression handling
- Parse/analysis separation must be maintained: the parser captures syntax only; the analyzer performs semantic validation
- Each increment must produce a working executable (parse → analyze → synthesize) with test data
- Aggregators must reduce the tensor by exactly one dimension per aggregator (`to` may be a concept name or `[]`, but when `to` is `[]` the `from` list must contain exactly one concept)
- Multi-dimensional reductions must be composed from sequences of single-dimension aggregators
- Synthesis requires full tensor matrix multiplication operations (dimension-reducing), which is a new capability beyond the currently supported element-wise, dimension-preserving tensor operations
- Types are initially represented as `float64` (detailed type selection design is deferred to [type-selection.md](../type-selection.md))

### Requirements

- Parse `aggregators` sections with `from [<concepts>] to [<concept>]` pattern (or `to []` for scalar reduction) and expression body (see [aggregators-and-spreaders-syntax.md](../../requirements/cpl/aggregators-and-spreaders-syntax.md))
- Build raw `ParserAggregatorDefinition` structures stored in the symbol table under synthetic keys derived from the aggregator's dimensional pattern
- Walk the symbol table during analysis to assemble the `AggregatorInfo` map, indexed by `(from_concept_set, to_concept)`
- Implement dimensional unit analysis: compute the dimensional unit effect of the aggregator expression and verify `input_units * expression_units = output_units`
- Implement cell unit validation alongside dimensional unit validation
- Generate pre-built `AggregationMatrix` instances from concept value sets during CIL generation
- When dimimension mismatches are detected during metric DAG analysis, query the Aggregators map and insert `CILAggregateNode` nodes into affected metric CIL trees
- Implement the aggregator sequence selection algorithm: least-information-loss → minimal-work → arbitrary
- Support composable aggregator sequences (multiple single-dimension aggregators applied in series)
- Support lookups referenced in aggregator expressions (once lookups are implemented — see ERO006)

### Assumptions

- Concepts, types, units, measures, connectors, and metrics are already implemented (Epics 001-005 provide the foundation)
- The ANTLR grammar can be extended with aggregator section parsing
- Expression parsing reuses the postfix expression pattern established for metrics
- Concept value sets are known at compile time, enabling pre-computation of aggregation matrices
- Full tensor matrix multiplication infrastructure will be implemented as part of aggregator synthesis
- Lookups (ERO006) are available when needed for aggregator expression references

## Evaluation Criteria

1. **End-to-End Completeness** (High importance): Each increment must produce a working executable that applies aggregation transformations to metric data.
2. **Incremental Value** (High importance): Each increment must add usable functionality; no increment should be purely scaffolding.
3. **Parse/Analysis Separation** (High importance): The parser must not perform semantic validation; all dimensional unit analysis and aggregator selection must happen in analysis.
4. **Dimensional Correctness** (High importance): The dimensional unit validation must correctly enforce that aggregator expressions produce the required dimensional transformation.
5. **Testability** (Medium importance): Each increment must include working examples with CPL files, IO specifications, and test data, covering both positive cases and error cases.

## Options Considered

### Option 1: Implement Aggregation as Element-Wise Operations

**Description:**
Implement aggregators using the existing element-wise tensor operation infrastructure, applying aggregation one cell at a time rather than using full matrix multiplication.

**Pros:**
- Reuses existing tensor operation infrastructure
- Faster time to first working increment
- Avoids implementing matrix multiplication

**Cons:**
- Violates the design requirement that aggregators reduce tensor dimensionality — the tensor's shape must change, not just cell values
- Would require workarounds that diverge from the design and would need to be rewritten later
- Element-wise approach fundamentally cannot express dimension reduction
- Would not support sequences of aggregators (chained dimension reductions)

### Option 2: Matrix Multiplication as Standalone Increment, Then Aggregator Parsing

**Description:**
Increment 1 builds the tensor matrix multiplication infrastructure as a standalone, testable piece of work with no CPL parsing dependencies. Increment 2 adds aggregator parsing and validation on top of the matrix foundation. Subsequent increments integrate aggregators into the metric DAG and optimize sequences.

**Pros:**
- Matrix multiplication is independently testable and reusable for both aggregators and spreaders
- Increment 1 is a clean, bounded piece of work with well-defined inputs/outputs (pure tensor math)
- Aggregator parsing (Increment 2) benefits from having the runtime infrastructure already in place
- Lower risk: matrix multiplication can be validated with simple unit tests before being consumed by the CPL pipeline
- The matrix multiplication infrastructure is a dependency that can be built and merged independently

**Cons:**
- Increment 1 produces no CPL-level functionality (pure Go tensor math)
- Four increments instead of three
- Increment 1 has no parse/analyze/CIL component (synthesis-only infrastructure)

### Option 3: Increment 1 with Manual Aggregation, Then Full Matrix Multiplication

**Description:**
Increment 1 parses and validates aggregators but applies them through manually constructed Go code (hardcoded aggregation for test cases). Increment 2 replaces the manual approach with full matrix multiplication infrastructure.

**Pros:**
- Increment 1 produces a working end-to-end result quickly, validating the parse/analyze pipeline
- Matrix multiplication is deferred to Increment 2 where it can be built with confidence in the surrounding infrastructure
- Balances early results with architectural correctness

**Cons:**
- Increment 1's manually constructed code is throwaway — it must be replaced in Increment 2
- Two distinct implementation approaches to maintain in sequence

## Decision

**Chosen Option:** Option 2 — Matrix Multiplication as Standalone Increment, Then Aggregator Parsing

**Rationale:**
Aggregators fundamentally change tensor dimensionality. Implementing them using element-wise operations would be architecturally incorrect. The matrix multiplication infrastructure is a necessary dependency for both aggregators and spreaders — building it first as a standalone increment (Increment 1) gives it clear boundaries, independent testability, and reusability.

This structure has several advantages:
1. **Independent validation**: Matrix multiplication can be tested exhaustively with hand-constructed tensors before being consumed by the CPL pipeline.
2. **Reusability**: The same matrix multiplication infrastructure is directly usable by both aggregators (dimension-reducing) and spreaders (dimension-expanding).
3. **Risk reduction**: If matrix multiplication is harder than expected, the issue is isolated and doesn't block CPL parsing/analysis work. If it's straightforward, Increment 1 completes quickly and Increment 2 can begin immediately.
4. **Clean dependency**: Aggregator parsing (Increment 2) targets a working runtime, making synthesis simpler — it just needs to emit matrix multiplication calls rather than implementing them inline.

The cost is one additional increment (4 vs. 3) and a synthesis-only first increment with no CPL parsing component. Both are acceptable trade-offs for the architectural clarity and risk reduction.

## Implications

### Positive Implications

- Architecturally correct tensor operations from day one — matrix multiplication built and tested before aggregator parsing begins
- Matrix multiplication infrastructure is independently testable and reusable for spreaders and other dimensional transformations
- No throwaway code or rework required
- Each aggregator naturally produces the correct output tensor shape
- Clean separation between infrastructure work (Increment 1) and CPL feature work (Increments 2-4)

### Negative Implications / Trade-offs

- Four increments instead of three; Increment 1 has no parse/analyze/CIL component
- Increment 1 is pure infrastructure with no CPL-level functionality
- Delayed first demonstrable end-to-end aggregator result compared to a parse-first approach
- Matrix multiplication implementation complexity may reveal additional design decisions that need ERO documentation
- Performance characteristics of matrix multiplication may need optimization in later increments

### Impact on System Qualities

| Quality Attribute | Impact | Notes |
| ----------------- | ------ | ----- |
| Correctness       | +      | Dimensional unit validation and proper tensor shapes prevent subtle errors |
| Performance       | -      | Matrix multiplication may be expensive for large tensors; optimization deferred |
| Maintainability   | +      | Clean architecture with no workarounds or throwaway code |
| Testability       | +      | Each increment includes both positive and negative test cases |
| Complexity        | -      | Matrix multiplication is a significant new capability; 4 increments vs. 3 |

### Future Considerations

- Tensor matrix multiplication performance optimization (parallelism, sparse matrices)
- Runtime adaptation of aggregation matrices when concept values change
- Aggregator caching for repeated pattern queries during analysis
- Aggregators that reference metric data within their expressions (beyond concept-based operations)
- Fused aggregator sequences (applying multiple aggregators in a single optimized pass)

## Notes

### Increment Sketches

Each increment is end-to-end across parse → analyze → CIL generation → CIL walk → synthesize. Every increment produces a working executable with test data. See [Aggregators Design](../aggregators.md) for full design descriptions.

**Increment 1: Tensor Matrix Multiplication Infrastructure**

This is a standalone infrastructure increment that does not involve any aggregator parsing or CPL syntax. It implements full matrix multiplication operations on tensors, which is a prerequisite for both aggregators (dimension-reducing multiplication) and spreaders (dimension-expanding multiplication). This is a fundamental new capability beyond the currently supported element-wise, dimension-preserving tensor operations.

- _Implementation:_
  - New tensor matrix multiplication operations that can change tensor dimensionality: reducing a tensor by one dimension (for aggregators) and expanding a tensor by one dimension (for spreaders).
  - For dimension reduction: given an N-dimensional source tensor and a 2D aggregation matrix (rows = source concept values being collapsed, columns = target concept values), produce an (N-1)-dimensional result tensor. Each cell in the result is computed by contracting the source tensor along the collapsed dimension using the matrix weights.
  - For dimension expansion: given an N-dimensional source tensor and a 2D spreading matrix (rows = source concept values, columns = new concept values), produce an (N+1)-dimensional result tensor. Each cell in the result is the source cell value multiplied by the corresponding matrix weight.
  - Both operations must handle arbitrary tensor ranks (including rank zero for scalar tensors) and arbitrary dimension positions for the affected dimension.
  - Matrix multiplication operations must be type-generic, working with the existing `*ChronoconeTensor[any]` interface.
- _Testing:_
  - Unit tests for dimension reduction: 1D → scalar, 2D → 1D, 3D → 2D, with known matrices and expected outputs.
  - Unit tests for dimension expansion: scalar → 1D, 1D → 2D, 2D → 3D, with known matrices and expected outputs.
  - Edge cases: rank zero tensors, single-value dimensions, identity matrices (no-op multiplication).
  - No specific concepts or CPL files are needed — this is pure Go tensor math tested with hand-constructed tensors.
- _No CPL parsing, analysis, or CIL changes in this increment._ The matrix multiplication infrastructure is the foundation that Increments 2 and later will use.

**Increment 2: Basic Aggregator Parsing and Validation**

- _Parser:_ New `CPLAggregators.g4` grammar file with `aggregatorsSection` and `aggregatorDeclaration` rules, recognizing `from [<concepts>] to [<concept>]` (or `to []`) pattern and expression body. New `AggregatorsListener` that pushes `ParserAggregatorDefinition` onto the symbol table under synthetic keys. Aggregator expressions converted to postfix notation using the standard postfix expression parsing pattern.
- _Analyzer:_ New `AnalyzeAggregatorsScope` function registered with `WalkSymbolTable`. Assembles `AggregatorInfo` entries into the Aggregators lookup map, keyed by `(from_concept_set, to_concept)`. Validates concept references in `from`, `to`, and expression. Implements dimensional unit computation for expression components (`sum`, `count`, `size`, concept references) and validates dimensional unit compatibility for each aggregator. Implements cell unit validation.
- _CIL Generation:_ Generate `CILAggregateNode` with pre-computed `AggregationMatrix` for each valid aggregator. The matrix is constructed from concept value sets: rows correspond to source concept values being collapsed, columns correspond to target concept values, entries are aggregation weights.
- _Synthesis:_ CIL tree walker handles `CILAggregateNode`, emitting Go code that applies the aggregation matrix to the source tensor using the matrix multiplication infrastructure from Increment 1. Each aggregator produces an output tensor with one fewer dimension (or rank zero for `to []`). Manual application in test code (not yet integrated into metric analysis).
- _Examples:_
  - `parse_and_analyze_only/without_errors/` — CPL files with valid aggregator declarations (single-dimension reduction, multi-concept `from` sets)
  - `parse_and_analyze_only/with_errors/` — CPL files with invalid aggregators (missing `to` concept, `to` concept not in `from`, `to []` with multiple concepts in `from`, non-existent concept references, incompatible dimensional units)
  - `with_data/aggregators_basic/` — CPL + IO spec + parquet data + `.expected.json` for a manually applied aggregator (e.g., summing `Country` in a `[Country, Product]` metric to produce a `[Product]` result)

**Increment 3: Additional Aggregator Function CIL Node Types**

This increment defines operation-specific CIL node types (`CILAggregateSumNode`, `CILAggregateCountNode`, `CILAggregateSizeNode`, `CILAggregateInverseSumNode`, `CILAggregateInverseCountNode`) for the remaining aggregator functions. The FunctionRegistry is extended to map all function names to their corresponding CIL node types. The generic `CILAggregateNode` from Increment 2 is replaced with typed nodes.

- _Parser:_ No changes — function calls use the generic `fnCall` rule already parsed in Increment 2.
- _Analyzer:_ Register all functions (`count`, `size`, `1/sum`, `1/count`) in the FunctionRegistry. Implement `resolveAggregatorFnCall` to create the appropriate CIL node type based on registry lookup. Each CIL node type carries its own aggregation matrix (or concept value count for `size`). CIL nodes fit into expressions — e.g., `(a-b)*sum(Country)` produces a CIL tree with `CILBinaryOp("*", ...)` wrapping a `CILAggregateSumNode`.
- _Synthesis:_ CIL tree walker handles each new node type, emitting the corresponding matrix multiplication or scalar constant. `CILAggregateSizeNode` emits a constant integer (the size value) rather than a tensor operation.
- _Examples:_
  - `parse_and_analyze_only/without_errors/` — CPL files with aggregators using `count`, `size`, `1/sum`, `1/count` functions
  - `parse_and_analyze_only/with_errors/` — CPL files with unknown function names (not in registry)
  - `with_data/aggregators_functions/` — CPL + IO spec + parquet data + `.expected.json` for aggregators using multiple function types

**Increment 4: Aggregator Selection and Integration**

- _Parser:_ No changes — aggregator declarations already parsed in Increment 2.
- _Analyze:_ After the full metric dependency DAG is assembled (second analysis pass), detect dimension mismatches between metric inputs and outputs. Query the Aggregators map for applicable transformations using the dimensional pattern key. Identify a single valid aggregator sequence that starts with the available dimensions (ultimately anchored in the dimensions of connector inputs) and results in the target dimensions specified in the IO spec. At this increment, only cases where exactly one possible application sequence exists are supported — if multiple sequences could satisfy the dimensional requirements, this is deferred to Increment 4. Insert `CILAggregateNode` nodes into affected metric CIL trees between the input reference and the calculation body.
- _Synthesis:_ CIL tree walker emits Go code that applies aggregator sequences as part of metric computation. The generated code applies each aggregator's matrix multiplication in order, producing intermediate tensors at each step.
- _Examples:_
  - `parse_and_analyze_only/without_errors/` — CPL files with metrics that have different dimensional granularities from their inputs, requiring implicit aggregator application where exactly one aggregator sequence exists
  - `parse_and_analyze_only/with_errors/` — CPL files where a dimension mismatch exists but no applicable aggregator bridges the gap
  - `with_data/aggregators_integration/` — CPL + IO spec + parquet data + `.expected.json` for metrics composed from inputs at different dimensional granularities, with exactly one valid aggregation path (e.g., input `[Country, Product]`, output `[Product]`, one aggregator matching `from [Country, Product] to [Product]`)

**Increment 5: Aggregator Sequence Optimization**

- _Parser:_ No changes.
- _Analyze:_ Implement the full selection algorithm:
  1. Least-information-loss principle: prefer aggregator sequences that preserve more dimension information (e.g., prefer aggregating a dimension with fewer distinct values first)
  2. Minimal-work principle: if multiple sequences are equally information-preserving, select the one that requires the least computational work
  3. If still equivalent, select arbitrarily but deterministically (e.g., lexicographic by concept name)
  - Implement specificity comparison for ambiguous aggregator resolution (more concepts in `from` set = more specific).
- _Synthesis:_ Optimize multi-step aggregator sequences into fused operations where possible (e.g., combine adjacent aggregation matrices into a single matrix multiplication). Handle edge cases: empty `to` (scalar reduction), single-dimension metrics, metrics with many extra dimensions that are unaffected by the aggregation.
- _Examples:_
  - `with_data/aggregators_optimization/` — CPL + IO spec + parquet data + `.expected.json` for multi-step aggregation sequences that exercise the selection algorithm (e.g., input `[Country, Region, Product]`, output `[Product]`, framework selects optimal collapse order)
  - Test cases with ambiguous aggregator definitions to verify selection correctness

### Phase-Specific EROs

- Parsing Phase ERO: [ERO007_aggregators-parsing-phase.md](../go-parser-target/eros/ERO007_aggregators-parsing-phase.md)
- Analysis Phase ERO: [ERO007_aggregators-analysis-phase.md](../go-analysis/eros/ERO007_aggregators-analysis-phase.md)
- Synthesis Phase ERO: [ERO007_aggregators-synthesis-phase.md](../go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md)

### Related Documents

- [Aggregators Design](../aggregators.md) — Full aggregators design with core rules, dimensional unit analysis, CIL node types, and 3 end-to-end increments
- [Spreaders Design](../spreaders.md) — Spreader design (symmetric inverse operation, shared matrix multiplication infrastructure)
- [CPL Requirements — Aggregators and Spreaders](../../requirements/cpl/aggregators-and-spreaders-syntax.md) — Syntax requirements for aggregators and spreaders
- [CPL Requirements — Concepts, Types, Units, Measures](../../requirements/cpl/concepts-types-units-measures-syntax.md) — Concept hierarchy (aggregators use parent-child relationships)
- [Lookups Design](../lookups.md) — Lookup design (aggregator expressions may reference lookups)
- [Metrics Design](../metrics-design.md) — Metric design (aggregators are applied to metric inputs during analysis)
- [Concept Definitions Design](../concepts-definitions.md) — Concept design (aggregator matrices are built from concept value sets)
- [Units and Measures Design](../units-and-measures.md) — Unit/measure design (dimensional unit algebra)
- [Postfix Expression Parsing Pattern](../../../patterns/postfix-expression-parsing.md) — Pattern for expression handling
- [Symbol Table-Driven Parsing Pattern](../../../patterns/symbol-table-driven-parsing.md) — Architectural pattern for parsing phase design
- [CPL Architecture — Intermediate Language](../../architecture/cpl/analysis/intermediate_language.md) — CIL specification
