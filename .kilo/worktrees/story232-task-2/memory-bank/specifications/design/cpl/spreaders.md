# Spreaders Design

## Overview

Spreaders are a fundamental building block of the Chronocone Planning Language (CPL). They transform metrics from less granular dimensional representations to more granular ones — for example, distributing a metric broken out by `[Product]` to `[Country, Product]` by applying a weighted distribution across countries. Spreaders are the inverse operation of aggregators: where aggregators collapse a dimension, spreaders introduce one.

Spreaders are the mechanism by which the CPL calculation framework makes the dimensions of an input metric match the dimensions of the metric that references it when the output requires a finer level of detail.

The form of a spreader declaration may be found in [CPL Requirements — Aggregators and Spreaders](../../requirements/cpl/aggregators-and-spreaders-syntax.md).

Spreaders span all four phases of CPL compilation: parsing, analysis, CIL generation, and synthesis. Like aggregators, spreaders are **unnamed** — they are defined by their `from [<concepts>] to [<concepts>]` dimensional pattern and are never explicitly referenced by name in other CPL declarations. The calculation framework selects the appropriate spreader at compile time based on the dimensional transformation required.

## Key Design Principles

1. **Unnamed, Pattern-Matched**: Spreaders have no user-visible name. They are indexed and matched by their dimensional transformation pattern (`from` concept set → `to` concept set). The analyzer queries the spreader index when it detects a dimension mismatch between an input metric and the metric that references it.

2. **Declarative, Not Explicit**: Users declare the set of possible spreaders. They do not specify when or in what order spreaders are applied. The CPL analysis process determines the optimal sequence to produce the required dimensional alignment.

3. **One Dimension Per Spreader**: Each spreader expands the dimensional representation by exactly one concept. The `to` concept set contains all concepts from the `from` set plus exactly one additional concept. Conversely, the `from` set may be `[]` (a scalar with no dimensions), in which case the `to` set contains exactly one concept.

4. **Dimensional Unit Analysis**: Concepts take on the effective role of units during spreading. The spreader expression is validated through dimensional unit algebra — the product of the input metric's dimensional units and the spreader expression's dimensional unit effect must equal the output's dimensional units. This is the same mechanism used for aggregators and for unit/measure conversion.

5. **Vector Multiplication Model**: Conceptually, a spread operation multiplies the aggregated value `A` by a vector `[b0, b1, ..., bn]` (one entry per value of the added dimension) to produce `[A*b0, A*b1, ..., A*bn]`. The expression defines how to construct this vector.

6. **Metric References in Expressions**: Unlike aggregators, spreader expressions may reference other metrics (via their names). This enables weighted distributions — for example, spreading sales targets across countries proportionally to customer counts. When a metric is referenced, the dimensional unit analysis determines whether the raw metric or an aggregated/dimensionally transformed version of it is required to satisfy the dimensional unit constraints.

## Core Rules for Spreaders

### 1. Spreader Sections

Spreaders are defined within `spreaders` sections. A CPL file may contain one or more `spreaders` sections, and spreader definitions may be spread across multiple files within the same compilation unit.

The general form of a spreaders section is:

```
spreaders
    <list of spreader definitions>
```

### 2. Spreader Declaration Structure

A spreader declaration has the form:

```
from [<list of concepts>] to [<list of concepts plus one more>] = <expression involving concepts>
```

Where `from` may also be the empty list `[]`.

- The `from` clause lists the concepts present in the input metric's dimensional representation. This list may be a superset of the concepts actually involved in the transformation.
- The `to` clause lists the output concepts. It must contain all concepts from the `from` list, plus exactly one additional concept. The additional concept is the one being introduced by the spread.
- The `from` list may be empty (`[]`), in which case the spreader distributes a scalar metric (no dimensions) across a single dimension.
- The expression on the right-hand side defines the distribution logic for the new dimension.

```
spreaders
    from [Product] to [Country, Product] = Product * 1/size(Country) * Country
```

This declares a spreader that introduces the `Country` dimension into a `Product`-dimensioned metric. The expression `1/size(Country) * Country` constructs a vector of size equal to the number of countries, with each element being `1/size(Country)` — an equal-proportion distribution. The `Product` concept reference in the expression means the distribution is applied per product value.

A scalar-spreading example:

```
spreaders
    from [] to [Country] = 1/size(Country) * Country
```

This spreads a single scalar value evenly across all countries.

### 3. Spreader Indexing and Matching

Spreaders are indexed by their dimensional pattern during analysis. The index key is the pair `(from_concept_set, to_concept_set)`.

When the analyzer determines that an input metric's dimensions need to be expanded to match the output metric's dimensions, it queries the spreader index:

1. Identify the concept(s) present in the output dimensions but absent from the input dimensions
2. Find all spreaders whose `from` set is a subset of the input dimensions and whose `to` set includes the input dimensions plus an additional concept that matches one of the concepts being added
3. If multiple spreaders could apply, sort candidates by specificity (more concepts in the `from` set = more specific)
4. If multiple equally specific spreaders match, select the one whose dimensional unit expression best matches the required transformation
5. If multiple spreaders are still equivalent, select the one with the minimal computational cost

### 4. Spreader Composability

A single metric may require multiple spreaders applied in sequence to reach the desired output dimensions. For example, going from `[Product]` to `[Country, Region, Product]` would require a sequence of two spreaders. The framework enumerates valid sequences and selects the optimal one.

A sequence may involve spreaders from different `spreaders` sections and different CPL files. The full set of spreaders across all files in the compilation unit is pooled before sequence enumeration.

Spreaders and aggregators may be interleaved in a transformation sequence. The framework selects the sequence that loses the least information.

### 5. Metric References in Spreader Expressions

Spreader expressions may reference other metrics by name. This is the mechanism for weighted distributions based on external data:

```
spreaders
    from [Product] to [Country, Product] = Product / customers
```

Here `customers` is a metric with dimensions `[Country]` that holds customer counts per country. The expression `Product / customers` means "divide the sales for each product by the customer proportion for each country."

When a spreader expression references a metric, the dimensional unit analysis determines the correct dimensional form of that metric to use:

- The raw `customers` metric has dimensional units `(Country)`
- The spreader must produce output dimensional units `(Country * Product)` from input `(Product)`
- The required dimensional unit factor is `(1/Country)`, not `(Country)`
- The analyzer therefore needs `customers` in a form with dimensional units `(1/Country)` — which requires an aggregator of the form `from [Country] to [] = 1/sum(Country)` to be available
- The analyzer detects this requirement, applies the aggregator to `customers`, and uses the result in the spreader expression

This implicit aggregator application is a key design feature: the user writes the spreader expression in a natural way (`Product / customers`) and the framework determines which dimensional transformations of referenced metrics are required to satisfy the unit algebra.

## Dimensional Unit Analysis

Dimensional unit analysis for spreaders follows the same model as for aggregators, with the flow reversed: spreaders introduce a dimension rather than removing one; the dimensional unit algebra validates that the expression's effect on dimensional units bridges the gap between input and output.

### Spreader Unit Expression

For the spreader:

```
from [Product] to [Country, Product] = Product * 1/size(Country) * Country
```

- The input metric has dimensional units `(Product)` — one time series per product
- The spreader must produce output dimensional units `(Country * Product)` — one time series per (country, product) pair
- The dimensional unit gap is `Country` — the expression must provide dimensional units of `Country`
- `1/size(Country)` is `Unitless` (a dimensionless number)
- `Product` has dimensional units `(Product)` (applied as a multiplier per product value)
- `Country` has dimensional units `(Country)` (constructing a vector across countries)
- The expression's dimensional unit effect: `Unitless * Product * Country = Product * Country`
- Verification: `(Product) * (Product * Country) = ?` — here the `Product` in the expression is a multiplier, not a dimensional unit factor in the aggregator sense. The correct frame is: the expression constructs a vector over `Country` that is applied to each `Product` value, so the dimensional effect is `Country`, giving `(Product) * (Country) = (Country * Product)` ✓

### Spreader Expression Unit Table

The same operation table used for aggregators applies, with concept references in spreader expressions interpreted as construction operations rather than reduction operations:

| Operation | Dimensional Units | Description |
|-----------|------------------|-------------|
| `C` (concept reference) | `C` | Per-value reference to dimension C (constructs a vector when C is the new dimension) |
| `size(C)` | `Unitless` | Number of values in concept C |
| `1/size(C)` | `Unitless` | Inverse, often used as a distribution weight |
| `sum(C)` | `C` | Sum across dimension C (used when the spreader references an aggregated metric) |
| metric reference | Depends on usage | Dimensional units of the referenced metric, possibly transformed by implicit aggregators |

### Metric Reference Resolution

When a spreader expression references a metric `M`, the resolution process is:

1. Determine the dimensional units of the raw metric `M` from its declared dimensions
2. Determine the dimensional unit factor required for `M` to satisfy the spreader's unit equation
3. If the raw dimensional units match the requirement, use `M` directly
4. If a transformation is needed, search the aggregator index for a transformation that bridges the gap
5. If no applicable aggregator exists, emit a compile-time error with the required dimensional unit factor, so the user can define one

This resolution is performed during the analysis phase and results in the appropriate `CILAggregateNode` being inserted into the CIL tree for `M` before it is consumed by the spreader.

### Consistency with Aggregator Unit Model

The dimensional unit algebra for spreaders uses the same `*`, `/`, and `^` rules as aggregators and regular unit/measure conversions. Compound dimensional units are preserved as symbolic expressions. This ensures a single, consistent validation framework across all dimensional transformations.

## Spreader Definition Types

The following definition types capture spreader information at different phases of processing:

### 1. Parser Spreader Definition

During parsing, spreader definitions are captured from the AST and stored in parser data structures:

- Raw `from` concept list, `to` concept list, and expression as parsed
- Expression components (operators, concept references, metric references) stored in postfix order within the symbol table's `Definitions` slice, following the same [Postfix Expression Parsing Pattern](../../patterns/postfix-expression-parsing.md) used for metrics
- Source location information for error reporting

These definitions are pushed onto symbol table entries during parsing. Since spreaders are unnamed, they are stored under a synthetic scope key derived from their pattern.

### 2. AnalysisDefinitions — Spreader Map

After parsing completes, the symbol table is walked to assemble the **Spreaders lookup map** as part of AnalysisDefinitions:

- **SpreaderInfo** — Spreader information including `from` concept set, `to` concept set, and the validated expression tree
- **Dimensional Unit Validation** — Computed dimensional unit effect of the expression, verified against the `from → to` transformation
- **Cell Unit Validation** — Checked that expression operators are compatible with the cell types involved
- **Metric Reference Resolution** — All metric names referenced in the expression are resolved against the metric declarations, and any required aggregator transformations are identified
- **Concept Reference Resolution** — All concept names in `from`, `to`, and the expression are resolved against the concept hierarchy

The Spreaders lookup map is keyed by `(from_concept_set, to_concept_set)`. The map enables the analyzer to:
- Query for applicable spreaders when a dimension mismatch is detected (output has more dimensions than input)
- Rank spreaders by specificity and computational cost
- Perform dimensional and cell unit validation at query time

### 3. CILDefinition — Spreader Nodes

In the Chronocone Intermediate Language (CIL), spreaders are represented as:

- **CILSpreadNode** — A node in the CIL tree that represents the application of a spreader to a metric. It contains:
  - `SourceCIL` — Reference to the CIL subtree for the metric being spread
  - `FromDimensions` — The concept dimensions present in the source
  - `ToDimensions` — The concept dimensions of the output (source dimensions plus the new dimension)
  - `NewDimension` — The concept being introduced by this spreader
  - `SpreadingMatrix` — A pre-computed transformation matrix that maps the less granular concept values to the more granular set. For a spreader introducing `Country`, the matrix has columns for the new country values, with entries indicating the distribution weight for each
  - `DistributionExpression` — The validated expression tree for the distribution logic
  - `ReferencedMetrics` — Any metric CIL subtrees referenced by the spreader expression, including implicitly applied aggregator nodes as determined by dimensional unit analysis
  - `DimensionalUnitEffect` — The validated dimensional unit factor of this spreader

The `SpreadingMatrix` is pre-built during CIL generation using the concept value sets from the AnalysisDefinitions.

### 4. Synthesized Code — Spreader Computation

In the generated executable code:

- Each `CILSpreadNode` emits Go code that applies the spreading matrix to the source tensor
- The matrix multiplication is performed using the target language's tensor operations
- The distribution expression determines the weights at each position
- Referenced metrics are computed first (including any implicit aggregators), then used as inputs to the spread
- The resulting tensor has the expanded dimensional shape specified by the `to` concept set

**Tensor Dimensionality Requirement:** Spreader operations change the dimensionality of the tensor — they produce an output tensor with more dimensions than the input tensor. This is a fundamental shift from the element-wise, dimension-preserving tensor operations currently supported (where a scalar computation is applied independently to each cell in the result tensor). Synthesis of spreaders will require implementing full matrix multiplication operations for tensors that map a lower-dimensional tensor to a higher-dimensional one, increasing the tensor's rank by one (or from rank zero for a `from []` spreader). Implicit aggregators applied to referenced metrics within spreader expressions also require the same dimension-reducing matrix multiplication capabilities. These are dependencies that must be addressed before spreader synthesis can be completed.

## Data Flow

The CPL compiler processes spreaders through the standard progression:

```
CPL Source
    ↓
[Parser] → ParserSpreaderDefinition structures in SymbolDefinition.Data
    ↓
[Symbol Table Walk] → AnalysisDefinitions with SpreaderInfo map
    ↓
[Metric DAG Analysis] → Detect dimension mismatches, query spreader map
    ↓
[CIL Generation] → Insert CILSpreadNode(s) into affected metric CIL trees;
                    also insert implicit CILAggregateNode(s) for referenced metrics
    ↓
[CIL Tree Walk]  → Generate Go code for spreading matrix multiplication
    ↓
[Generated Code] → Go program with tensor spreading operations
    ↓
[Runtime]        → Execute: spreading matrices applied to tensor cells
```

### Spreader Application During Metric Analysis

Spreader application, like aggregator application, happens during the second pass of metric analysis:

1. For each metric, compare its input dimensions to its own declared dimensions
2. When the output has more dimensions than an input, query the Spreaders map for applicable transformations
3. For each spreader candidate that references other metrics, resolve the dimensional form of those metrics (applying implicit aggregators as needed)
4. Enumerate all valid sequences of spreaders (and aggregators, if interleaved) that bridge the gap
5. Select the optimal sequence using the least-information-loss → minimal-work → arbitrary selection principles
6. Insert `CILSpreadNode` (and any implicit `CILAggregateNode`) nodes into the metric's CIL tree between the input reference and the calculation body

## Validation Rules

The following validation rules must be enforced during parsing and analysis:

1. **From-To Containment**: The `to` concept set must contain all concepts from the `from` concept set.
2. **Exactly One Dimension Added**: The size of the `to` concept set must be exactly one greater than the size of the `from` concept set. This means `from` may be `[]` (empty), in which case `to` must contain exactly one concept.
3. **Concept Existence**: All concept names in `from`, `to`, and the expression must refer to existing concepts.
4. **Expression Operands**: All concept references in the expression must be to concepts that appear in either the `from` list or the `to` list.
5. **Metric Reference Existence**: Any metric name referenced in the expression must refer to an existing metric.
6. **Dimensional Unit Compatibility**: The dimensional units of the expression must combine with the input dimensional units to produce the output dimensional units. This is a compile-time error if violated.
7. **Cell Unit Compatibility**: The cell units of the values being spread must be compatible with the spreader expression. The expression must produce result cell units that match the output metric's cell type.
8. **Required Aggregator Existence**: If a referenced metric requires an implicit aggregator to satisfy dimensional unit constraints, that aggregator must exist. If not, this is a compile-time error.
9. **Ambiguous Spreaders**: If multiple spreaders share the same pattern `(from, to)` and have identical specificity, this is an error.
10. **No Applicable Spreader**: If a dimension mismatch is detected and no spreader can bridge the gap, this is a compile-time error.

## Incremental Implementation Strategy

### Increment 1: Basic Spreader Parsing and Validation

**Parse:**
- Parse `spreaders` sections with `from`, `to`, and expression
- Convert expressions to postfix notation using the standard postfix expression parsing pattern
- Store raw definitions in the symbol table under synthetic keys

**Analyze:**
- Walk symbol table to assemble the SpreaderInfo map
- Validate concept references in `from`, `to`, and expression
- Implement dimensional unit computation for expression components
- Validate dimensional unit compatibility for each spreader

**CIL Generation:**
- Generate `CILSpreadNode` with `SpreadingMatrix` for each valid spreader
- The matrix maps source concept values to target concept values based on concept value sets

**Synthesize:**
- Generate Go code for a single spreader applied manually (not yet integrated into metric analysis)
- Verify the spreading matrix produces correct tensor shapes in test cases

### Increment 2: Spreader Selection and Integration

**Analyze:**
- After the full metric dependency DAG is assembled, detect dimension mismatches where input has fewer dimensions than output
- Query the Spreaders map for applicable transformations
- Enumerate and rank valid spreader sequences
- Insert `CILSpreadNode` nodes into affected metric CIL trees

**Synthesize:**
- Generate Go code that applies spreader sequences as part of metric computation
- Include end-to-end examples with spread operations

### Increment 3: Metric References in Spreader Expressions

**Analyze:**
- When a spreader expression references a metric, resolve its dimensional form
- Determine whether implicit aggregators are required
- Validate that required aggregators exist
- Insert implicit `CILAggregateNode` nodes for referenced metrics

**Synthesize:**
- Generate Go code that computes referenced metrics (including aggregations) before applying the spread
- Include end-to-end examples with weighted distributions via referenced metrics

### Increment 4: Spreader/Aggregator Interleaving and Optimization

**Analyze:**
- Implement the full selection algorithm for mixed aggregator/spreader sequences
- Implement specificity comparison for ambiguous spreader resolution

**Synthesize:**
- Optimize multi-step sequences into fused operations where possible
- Handle edge cases: empty `from`, single-dimension metrics, sequences that interleave aggregators and spreaders

## Design Benefits

1. **Natural Expression of Inverse Operations**: Spreaders are the symmetric inverse of aggregators, allowing users to think about dimensional transformation as a pair of complementary operations.
2. **Compile-Time Verification via Unit Algebra**: The dimensional unit analysis for spreaders reuses the same `*`, `/`, `^` machinery as aggregators and unit conversions, providing a unified verification framework that catches errors before runtime.
3. **Implicit Aggregator Resolution**: Users write spreader expressions that reference metrics in their natural dimensional form; the framework determines the required transformations, reducing cognitive load.
4. **Metric-Driven Distributions**: Spreader expressions can reference real metrics (like customer counts) to define weighted distributions, enabling realistic business logic without hardcoded weights.

## Relationship with Aggregators

Spreaders and aggregators are inverse operations, but they are not exact inverses — a round-trip through aggregation then spreading generally loses information. The framework's selection algorithm prefers sequences that minimize information loss, which in practice means preferring to spread as late as possible and aggregate as early as possible in the metric DAG.

The two declaration types share:
- The same dimensional unit algebra model
- The same pattern-matching and indexing mechanism
- The same sequence selection algorithm
- Symmetric constraints: aggregators reduce by one dimension, spreaders add one dimension

## Related Documents

- [CPL Requirements — Aggregators and Spreaders](../../requirements/cpl/aggregators-and-spreaders-syntax.md) — Syntax requirements for aggregators and spreaders
- [CPL Requirements — Concepts, Types, Units, Measures](../../requirements/cpl/concepts-types-units-measures-syntax.md) — Concept hierarchy (spreaders distribute across concept values)
- [Aggregators Design](aggregators.md) — Aggregator design (shared dimensional unit model, inverse operation)
- [Units and Measures Design](units-and-measures.md) — Unit/measure design (dimensional unit algebra)
- [Metrics Design](metrics-design.md) — Metric design (spreaders are applied to metric inputs during analysis)
- [Concept Definitions Design](concepts-definitions.md) — Concept design (spreader matrices are built from concept value sets)
- [CPL Architecture — Intermediate Language](../../architecture/cpl/analysis/intermediate_language.md) — CIL specification

## Future Considerations

- Additional distribution strategies beyond equal-proportion and metric-weighted (e.g., historical proportion, seasonal patterns)
- Spreader caching for repeated pattern queries during analysis
- Spreader expressions that reference lookups (in addition to metrics) as distribution data sources
- Runtime adaptation of spreading matrices when concept values or distribution data change
