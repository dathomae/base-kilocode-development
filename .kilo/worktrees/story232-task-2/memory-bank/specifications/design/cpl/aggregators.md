# Aggregators Design

## Overview

Aggregators are a fundamental building block of the Chronocone Planning Language (CPL). They transform metrics from more granular dimensional representations to less granular ones — for example, collapsing a metric broken out by `[Country, Product, Warehouse]` to `[Country, Product]` by combining data across the `Warehouse` dimension. Aggregators are the mechanism by which the CPL calculation framework makes the dimensions of an input metric match the dimensions of the metric that references it when the output is at a coarser level of detail.

The form of an aggregator declaration may be found in [CPL Requirements — Aggregators and Spreaders](../../requirements/cpl/aggregators-and-spreaders-syntax.md).

Aggregators span all four phases of CPL compilation: parsing, analysis, CIL generation, and synthesis. Unlike metrics, connectors, and lookups, aggregators are **unnamed** — they are defined by their `from [<concepts>] to [<concepts>]` dimensional pattern and are never explicitly referenced by name in other CPL declarations. The calculation framework selects the appropriate aggregator at compile time based on the dimensional transformation required.

## Key Design Principles

1. **Unnamed, Pattern-Matched**: Aggregators have no user-visible name. They are indexed and matched by their dimensional transformation pattern (`from` concept set → `to` concept set). The analyzer queries the aggregator index when it detects a dimension mismatch between an input metric and the metric that references it.

2. **Declarative, Not Explicit**: Users declare the set of possible aggregators. They do not specify when or in what order aggregators are applied. The CPL analysis process determines the optimal sequence to produce the required dimensional alignment.

3. **Dimensional Unit Analysis**: Concepts take on the effective role of units during aggregation. The aggregator expression is validated through dimensional unit algebra — the product of the input metric's dimensional units and the aggregator's dimensional unit effect must equal the output's dimensional units. This check is structurally identical to the unit and measure conversion resolution model, using the same `*`, `/`, and `^` algebra.

4. **Cell Unit Preservation**: Aggregation operations (`sum`, etc.) require that the cell values being combined have identical units. The cell unit analysis runs in parallel with the dimensional unit analysis and can independently fail.

5. **Least-Information-Loss Selection**: When multiple aggregator sequences could accomplish the required dimensional transformation, the framework selects the sequence that loses the least information. If multiple sequences are equivalent after applying the least-information-loss principle, the framework selects the one that does the least computational work. If multiple sequences are still equivalent, the framework may select arbitrarily.

## Core Rules for Aggregators

### 1. Aggregator Sections

Aggregators are defined within `aggregators` sections. A CPL file may contain one or more `aggregators` sections, and aggregator definitions may be spread across multiple files within the same compilation unit.

The general form of an aggregators section is:

```
aggregators
    <list of aggregator definitions>
```

### 2. Aggregator Declaration Structure

An aggregator declaration has the form:

```
from <list of concepts> to [<concept name>] = <expression involving concepts>
```

Where `to` may also be the empty list `[]`.

- The `from` clause lists the concepts present in the input metric's dimensional representation. This list may be a superset of the concepts actually involved in the transformation — the aggregator applies to any metric whose dimensions include at least these concepts.
- The `to` clause specifies the target concept, or the empty list `[]` to reduce to a scalar (no concept dimensions). When `to` is `[]`, the `from` list must contain exactly one concept — an aggregator can only reduce to a scalar when collapsing a single dimension.
- Each aggregator reduces by exactly one dimension: the target concept must be one of the concepts in the `from` list, and it is the one that is collapsed. When `to` is `[]`, the single concept in the `from` list is collapsed to a scalar.
- The expression on the right-hand side defines the aggregation operation (e.g., `sum(Country)`) applied to the concept being collapsed.

```
aggregators
    from [Country, Product] to [Product] = sum(Country)
```

This declares an aggregator that collapses the `Country` dimension, summing the values across all country-specific time series into a single time series for each `Product` value. The aggregator will match any metric whose dimensions include both `Country` and `Product`, regardless of whether the metric also includes other dimensions such as `Warehouse` or `Region`.

A scalar-reducing aggregator collapses a single dimension to no dimensions:

```
aggregators
    from [Country] to [] = sum(Country)
```

This aggregator sums the values across all country-specific time series, producing a single scalar time series. Since `to` is `[]`, the `from` list must contain exactly one concept.

To reduce by multiple dimensions (e.g., collapsing both `Country` and `Region` to `Product`), the framework composes a sequence of single-dimension aggregators. In the example `from [Country, Region, Product] to [Product]`, the framework would apply two aggregators in sequence: one collapsing `Country`, another collapsing `Region` (or vice versa). Since each aggregator reduces by exactly one dimension, the dimensional unit validation is always a simplification of a single concept.

### 3. Aggregator Indexing and Matching

Since aggregators are unnamed, they must be indexed by their dimensional pattern during analysis. The index key is the pair `(from_concept_set, to_concept)`.

When the analyzer determines that an input metric's dimensions do not match the output metric's required dimensions, it queries the aggregator index:

1. Identify the concept(s) present in the input dimensions but absent from the output dimensions
2. Find all aggregators whose `from` set is a subset of the input dimensions and whose `to` concept is one of the concepts being removed
3. If multiple aggregators could apply, sort candidates by specificity (more concepts in the `from` set = more specific)
4. If multiple equally specific aggregators match, select the one whose dimensional unit expression best matches the required transformation
5. If multiple aggregators are still equivalent, select the one with the minimal computational cost

### 4. Aggregator Composability

A single metric may require multiple aggregators applied in sequence to reach the desired output dimensions. For example, going from `[Country, Region, Product]` to `[Product]` might require aggregating `Region` then `Country`, or `Country` then `Region`. The framework enumerates valid sequences and selects the optimal one.

A sequence may involve aggregators from different `aggregators` sections and different CPL files. The full set of aggregators across all files in the compilation unit is pooled before sequence enumeration.

## Dimensional Unit Analysis

Dimensional unit analysis is the core validation mechanism for aggregators. It ensures that the aggregator expression produces the correct dimensional units for the output. This section describes the mathematical model; implementation across phases is covered in subsequent sections.

### Dimensional Unit Model

When a metric has dimensions `[C1, C2, ..., Cn]`, it implicitly has **dimensional units** derived from those concepts. The exact form depends on the context:

- For a metric with dimensions `[Country, Product]`, the dimensional units are `Product/Country` — one time series per product, with many country values contributing to each product time series.
- For a metric with dimensions `[Product]`, the dimensional units are `Product` — one time series per product value.

Dimensional units are distinct from cell units (the units of the individual numeric values in the time series, such as `USD` or `Count`). Dimensional units are never visible to the user; they are an internal compiler concept used for validation.

### Aggregator Unit Expression

An aggregator's expression defines its effect on dimensional units. For the aggregator:

```
from [Country, Product] to [Product] = sum(Country)
```

- The input metric has dimensional units `Product/Country`
- The aggregator expression `sum(Country)` has dimensional units `Country` (it sums across the Country dimension)
- The output dimensional units are `Product/Country * Country = Product`

The result `Product` matches the output metric's dimensional units, so the aggregator is valid.

### Aggregator Expression Units

Each operation in an aggregator expression contributes a specific dimensional unit factor:

| Operation | Dimensional Units | Description |
|-----------|------------------|-------------|
| `sum(C)` | `C` | Sums values across dimension C |
| `count(C)` | `C` | Counts the number of values in dimension C |
| `size(C)` | `Unitless` | Number of values in concept C |
| `1/sum(C)` | `1/C` | Inverse of sum across C |
| `1/count(C)` | `1/C` | Inverse of count across C |
| `C` (concept reference) | `C` | The concept itself as a dimensional unit |

Concept references in expressions are interpreted as being applied per-value: `sum(Country)` means "for each Country value, sum the corresponding time series."

### Validation Algorithm

The validation of an aggregator proceeds as:

1. Compute the dimensional units of the output metric from the `to` concept set
2. Compute the dimensional units of the input metric from the `from` concept set
3. Compute the dimensional unit effect of the aggregator expression by evaluating the expression with dimensional unit algebra
4. Verify that `input_units * expression_units = output_units` (or equivalent rearrangement using `*`, `/` operations)

### Cell Unit Validation

In addition to the dimensional unit check, cell unit validation ensures that the aggregation operation is meaningful for the cell values:

- `sum(C)` is valid when all cell values being summed have identical units
- Operations that combine multiple metrics within the expression may introduce cell unit incompatibilities

Cell unit validation is performed alongside dimensional unit validation. Failures in either produce a compile-time error.

### Consistency with Unit/Measure Conversions

The dimensional unit algebra follows the same rules as regular unit and measure conversion algebra: only `*`, `/`, and `^` operations are permitted; `+` and `-` implicitly require identical units; compound dimensional units are preserved as symbolic expressions. This ensures that the dimensional unit model and the cell unit model use the same machinery and produce consistent results.

## Aggregator Definition Types

The following definition types capture aggregator information at different phases of processing:

### 1. Parser Aggregator Definition

During parsing, aggregator definitions are captured from the AST and stored in parser data structures:

- Raw `from` concept list, `to` concept (or empty), and expression as parsed
- Expression components (operators, concept references) stored in postfix order within the symbol table's `Definitions` slice, following the same [Postfix Expression Parsing Pattern](../../patterns/postfix-expression-parsing.md) used for metrics
- Source location information for error reporting

These definitions are pushed onto symbol table entries during parsing. Since aggregators are unnamed, they are stored under a synthetic scope key derived from their pattern.

### 2. AnalysisDefinitions — Aggregator Map

After parsing completes, the symbol table is walked to assemble the **Aggregators lookup map** as part of AnalysisDefinitions:

- **AggregatorInfo** — Aggregator information including `from` concept set, `to` concept (or empty), and the validated expression tree
- **Dimensional Unit Validation** — Computed dimensional unit effect of the expression, verified against the `from → to` transformation
- **Cell Unit Validation** — Checked that expression operators are compatible with the cell types involved
- **Concept Reference Resolution** — All concept names in `from`, `to`, and the expression are resolved against the concept hierarchy

The Aggregators lookup map is keyed by `(from_concept_set, to_concept)`. Multiple aggregators may share the same key (same pattern, different expressions). The map enables the analyzer to:
- Query for applicable aggregators when a dimension mismatch is detected
- Rank aggregators by specificity and computational cost
- Perform dimensional and cell unit validation at query time

### 3. CILDefinition — Aggregator Nodes

In the Chronocone Intermediate Language (CIL), aggregators are represented as:

- **CILAggregateNode** — A node in the CIL tree that represents the application of an aggregator to a metric. It contains:
  - `SourceCIL` — Reference to the CIL subtree for the metric being aggregated
  - `FromDimensions` — The concept dimensions present in the source
  - `ToDimension` — The concept dimension being collapsed (or empty for scalar)
  - `AggregationMatrix` — A pre-computed transformation matrix that maps the more granular concept values to the less granular set. For `sum(Country)`, the matrix has rows for source country values and columns for target product values, with entries indicating the aggregation weight
  - `CellOperation` — The operation applied at each cell (e.g., sum, count)
  - `DimensionalUnitEffect` — The validated dimensional unit factor of this aggregator

The `AggregationMatrix` is pre-built during CIL generation using the concept value sets from the AnalysisDefinitions. Since concept values are known at compile time, the matrix can be fully materialized before synthesis.

### 4. Synthesized Code — Aggregator Computation

In the generated executable code:

- Each `CILAggregateNode` emits Go code that applies the aggregation matrix to the source tensor
- The matrix multiplication is performed using the target language's tensor operations
- The cell operation (sum, count, etc.) determines the arithmetic applied at each position
- The resulting tensor has the reduced dimensional shape specified by the `to` concept set

**Tensor Dimensionality Requirement:** Aggregator operations change the dimensionality of the tensor — they produce an output tensor with fewer dimensions than the input tensor. This is a fundamental shift from the element-wise, dimension-preserving tensor operations currently supported (where a scalar computation is applied independently to each cell in the result tensor). Synthesis of aggregators will require implementing full matrix multiplication operations for tensors that map a higher-dimensional tensor to a lower-dimensional one, reducing the tensor's rank by one (or to rank zero for a `to []` aggregator). This is a dependency that must be addressed before aggregator synthesis can be completed.

## Data Flow

The CPL compiler processes aggregators through the standard progression:

```
CPL Source
    ↓
[Parser] → ParserAggregatorDefinition structures in SymbolDefinition.Data
    ↓
[Symbol Table Walk] → AnalysisDefinitions with AggregatorInfo map
    ↓
[Metric DAG Analysis] → Detect dimension mismatches, query aggregator map
    ↓
[CIL Generation] → Insert CILAggregateNode(s) into affected metric CIL trees
    ↓
[CIL Tree Walk]  → Generate Go code for aggregation matrix multiplication
    ↓
[Generated Code] → Go program with tensor aggregation operations
    ↓
[Runtime]        → Execute: aggregation matrices applied to tensor cells
```

### Aggregator Application During Metric Analysis

The most significant design element is when and how aggregators are applied — this happens during the second pass of metric analysis, after the full metric dependency DAG is assembled:

1. For each metric, compare its input dimensions to its own declared dimensions (from the IO specification or from the inputs that supply dimensions)
2. When a dimension mismatch is detected, query the Aggregators map for applicable transformations
3. Enumerate all valid sequences of aggregators that bridge the gap
4. Select the optimal sequence using the least-information-loss → minimal-work → arbitrary selection principles
5. Insert `CILAggregateNode` nodes into the metric's CIL tree between the input reference and the calculation body

## Validation Rules

The following validation rules must be enforced during parsing and analysis:

1. **From-To Overlap**: The concept named in the `to` clause must be present in the `from` list. If `to` is `[]`, the `from` list must contain exactly one concept.
2. **Concept Existence**: All concept names in the `from` list and the expression must refer to existing concepts.
3. **Expression Operands**: All concept references in the expression must be to concepts that appear in the `from` list.
4. **Dimensional Unit Compatibility**: The dimensional units of the expression must combine with the input dimensional units to produce the output dimensional units. This is a compile-time error if violated.
5. **Cell Unit Compatibility**: The cell units of the values being aggregated must be compatible with the aggregation operation. This is a compile-time error if violated.
6. **Ambiguous Aggregators**: If multiple aggregators share the same pattern `(from, to)` and have identical specificity, this is an error. Specificity is determined by the number of concepts in the `from` list — more concepts = more specific.
7. **No Applicable Aggregator**: If a dimension mismatch is detected and no aggregator can bridge the gap, this is a compile-time error.

## Incremental Implementation Strategy

### Increment 1: Basic Aggregator Parsing and Validation

**Parse:**
- Parse `aggregators` sections with `from`, `to`, and expression
- Convert expressions to postfix notation using the standard postfix expression parsing pattern
- Store raw definitions in the symbol table under synthetic keys

**Analyze:**
- Walk symbol table to assemble the AggregatorInfo map
- Validate concept references in `from`, `to`, and expression
- Implement dimensional unit computation for expression components (sum, count, size, concept references)
- Validate dimensional unit compatibility for each aggregator

**CIL Generation:**
- Generate `CILAggregateNode` with `AggregationMatrix` for each valid aggregator
- The matrix maps source concept values to target concept values based on the concept hierarchy

**Synthesize:**
- Generate Go code for a single aggregator applied manually (not yet integrated into metric analysis)
- Verify the aggregation matrix produces correct tensor shapes in test cases

### Increment 2: Aggregator Selection and Integration

**Analyze:**
- After the full metric dependency DAG is assembled, detect dimension mismatches between metric inputs and outputs
- Query the Aggregators map for applicable transformations
- Enumerate and rank valid aggregator sequences
- Insert `CILAggregateNode` nodes into affected metric CIL trees

**Synthesize:**
- Generate Go code that applies aggregator sequences as part of metric computation
- Include end-to-end examples where metrics with different dimensional granularities are composed

### Increment 3: Aggregator Sequence Optimization

**Analyze:**
- Implement the full selection algorithm: least-information-loss → minimal-work → arbitrary
- Implement specificity comparison for ambiguous aggregator resolution

**Synthesize:**
- Optimize multi-step aggregator sequences into fused operations where possible
- Handle edge cases: empty `to`, single-dimension metrics, metrics with many extra dimensions

## Design Benefits

1. **Declarative Safety**: Users declare possible transformations without specifying control flow; the framework applies them only where dimensionally correct.
2. **Compile-Time Verification**: Dimensional unit analysis catches misalignments before runtime, preventing subtle errors from dimension mismatches.
3. **Unified Unit Model**: The dimensional unit algebra reuses the same `*`, `/`, `^` machinery as cell unit conversion, reducing implementation complexity and cognitive load.
4. **Deterministic Resolution**: The least-information-loss → minimal-work → arbitrary selection rules ensure reproducible compilation results.

## Related Documents

- [CPL Requirements — Aggregators and Spreaders](../../requirements/cpl/aggregators-and-spreaders-syntax.md) — Syntax requirements for aggregators and spreaders
- [CPL Requirements — Concepts, Types, Units, Measures](../../requirements/cpl/concepts-types-units-measures-syntax.md) — Concept hierarchy (aggregators use concept parent-child relationships)
- [Units and Measures Design](units-and-measures.md) — Unit/measure design (dimensional units reuse the same algebra)
- [Metrics Design](metrics-design.md) — Metric design (aggregators are applied to metric inputs during analysis)
- [Concept Definitions Design](concepts-definitions.md) — Concept design (aggregator matrices are built from concept value sets)
- [CPL Architecture — Intermediate Language](../../architecture/cpl/analysis/intermediate_language.md) — CIL specification

## Future Considerations

- Additional aggregation operations beyond `sum`, `count`, and `size` (e.g., `avg`, `min`, `max`, weighted sum)
- Aggregators that reference metric data within their expressions (similar to how spreaders can reference metrics)
- Aggregator caching for repeated pattern queries during analysis
- Runtime adaptation of aggregation matrices when concept values change
