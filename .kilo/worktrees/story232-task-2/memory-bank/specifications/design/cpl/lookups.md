# Lookups Design

## Overview

Lookups are a fundamental building block of the Chronocone Planning Language (CPL). They provide "side data" to metric calculations — reference data that maps input dimensions (from concepts and optionally time grains) to values. Lookups are used for items such as conversion factors, weighting tables, and dimension mappings.

Lookups are similar to connectors in that they supply data to the computation framework, but they differ in three important ways: (1) lookups are expected to change more slowly than connector data and are typically read from files or databases rather than external services; (2) lookups often have no time dimension, representing data that does not vary with time; and (3) lookup dimensions serve as keys into the lookup table, whereas connector dimensions describe the output shape of the data.

The form of a lookup declaration may be found in [CPL Requirements — Lookups](../../requirements/cpl/lookups-syntax.md). The form of a Lookup IO Specification is described in that same document.

Lookups span all four phases of CPL compilation: parsing, analysis, CIL generation, and synthesis.

## Key Design Principles

1. **Key-Value Semantics**: Lookup dimensions function as keys; the lookup returns a single value per key combination. This is distinct from connectors, whose dimensions describe the full cross-product shape of a tensor.
2. **Optional Temporality**: Lookups may or may not have a time dimension. Timeless lookups are purely dimensionally indexed; temporal lookups add grain dimensions to the key.
3. **Unit-Carrying Cell Types**: The `celltype` measure of a lookup carries compound units that reflect the dimensional relationship of the key concepts (e.g., `Currency/Currency` for a conversion matrix). This enables the calculation framework to verify dimensional correctness at compile time.
4. **MetricComputation Integration**: At runtime, lookup data is registered as `MetricComputation` objects, allowing it to participate in metric calculations identically to connector inputs and computed metrics.

## Core Rules for Lookups

### 1. Lookup Sections

Lookups are defined within `lookups` sections. A CPL file may contain one or more `lookups` sections, and lookup definitions may be spread across multiple files within the same compilation unit.

The general form of a lookups section is:

```
lookups
    <list of lookup definitions>
```

### 2. Lookup Identity

A lookup name in one lookups section always refers to the same lookup as the same name in a different lookups section. Lookup names are globally unique within a compilation unit and namespace.

### 3. Lookup Declaration Structure

A lookup declaration consists of a name, a `dimensions` clause listing the key concepts, an optional `grain` clause listing time measures, and a `celltype` clause specifying the measure that defines the type and units of the returned value:

```
<lookup name>
    dimensions [<list of concept names>]
    [grain <measure name>[, <measure name>...]]
    celltype <measure name>
```

The `dimensions` clause lists the concept names that serve as keys. The order of concept names in the `dimensions` list is significant: the first dimension listed is interpreted as the outermost key, and position in the list determines how the lookup tensor is indexed.

The `grain` clause is optional. When absent, the lookup has no time dimension. When present, it references one or more measure names whose types must be `Interval` or derived from `Interval`. Multiple grain measures allow the lookup to satisfy requests at different time granularities.

The `celltype` clause references a single measure name. The measure's unit reflects the dimensional relationship of the key concepts. For a currency conversion lookup with dimensions `[Currency, Currency]`, the celltype measure would have units `Currency/Currency`.

### 4. Key-Value Shape vs. Tensor Shape

A lookup's logical shape is a key-value mapping: for each combination of values of the key dimensions (concept dimensions plus any grain dimensions), there is a single value. This is conceptually distinct from a connector's shape, which is a full tensor of all concept-value combinations.

However, at runtime, lookup data is loaded into tensors indexed by the key dimensions, with cell values of the type specified by `celltype`. The tensor representation is used for uniformity — it allows the same `MetricComputation` interface to serve both connectors and lookups.

### 5. Key-Miss Behavior

A request for a lookup value with a key (dimensional combination) that does not exist in the source data is a runtime error. There is no default value mechanism. The caller is responsible for ensuring that the lookup covers the required key space. This constraint may be relaxed in the future if use cases demonstrate a need for default values.

### 6. Unit Analysis for Lookups

The units of a lookup's `celltype` measure must be consistent with the dimensional relationship of the key concepts. The calculation framework uses these units during compile-time dimensional analysis to verify that lookup values are used correctly in metric calculations.

For example, a conversion lookup with dimensions `[Currency, Currency]` and celltype units `Currency/Currency` composes to specific compound units like `EURO/USD` for each row. When a metric multiplies a value denominated in USD by this lookup, the unit algebra `USD * (EURO/USD) = EURO` confirms the result is in the expected target currency.

### 7. Lookup Integration in Metric Inputs

Metrics reference lookups via the `lookup` input kind in their `inputs` block:

```
metrics
    converted_revenue
        inputs
            metric revenue
            lookup currency_conversion
            dimension Currency
        calculate
            revenue / currency_conversion
```

The `lookup` input is distinct from `metric` and `dimension` inputs. The calculation framework resolves lookup references against the set of lookup declarations in the compilation unit.

## Lookup Definition Types

The following definition types capture lookup information at different phases of processing:

### 1. Parser Lookup Definition

During parsing, lookup definitions are captured from the AST and stored in parser data structures:

- Raw lookup name, dimension list, optional grain list, and celltype reference as parsed
- Source location information for error reporting

These definitions are pushed onto symbol table entries during parsing. The parser does not validate cross-references, resolve dimension names, or check semantic rules beyond syntax.

### 2. AnalysisDefinitions — Lookup Map

After parsing completes, the symbol table is walked to assemble the **Lookups lookup map** as part of AnalysisDefinitions:

- **LookupInfo** — Core lookup information including name, resolved dimension concept references, resolved grain measure references (if any), and resolved celltype measure reference
- **Dimension Validation** — Verification that all dimension concept names exist in the concept hierarchy
- **Grain Validation** — Verification that grain measure names exist and have types derived from `Interval`
- **Celltype Validation** — Verification that the celltype measure name exists

The Lookups lookup map enables the analyzer to:
- Validate lookup references in metric inputs
- Check dimensionality and grain compatibility between lookups and requesting metrics
- Resolve lookup tensor shapes for CIL generation

### 3. CILDefinition — Lookup Nodes

In the Chronocone Intermediate Language (CIL), lookups are represented as:

- **CILLookupNode** — A leaf node in the CIL tree representing a lookup data source, analogous to a connector CIL node. It captures the lookup name, dimensions, grain (if any), and celltype. It produces a tensor that the calculation framework can index for metric computations.

Since lookup structures are known at synthesis time, the CIL can include:
- Pre-computed dimension value sets for indexing
- Known tensor shapes based on declared dimensions and grain

### 4. Synthesized Code — Lookup Integration

In the generated executable code:

- Each lookup is registered as a `MetricComputation` object on the `ChronoconeComputationController` before the postfix evaluator runs
- Lookup data is loaded from the source specified in the Lookup IO Specification
- The generated compute function calls `controller.GetMetricComputation(lookupName)` to access lookup data at runtime
- Lookup tensors are indexed by the key dimensions to retrieve values for metric calculations

## Data Flow

The CPL compiler processes lookups through the same strict progression as other declaration types:

```
CPL Source
    ↓
[Parser] → ParserLookupDefinition structures stored in SymbolDefinition.Data
    ↓
[Symbol Table Walk] → AnalysisDefinitions with LookupInfo entries
    ↓
[CIL Generation] → One single-node CIL tree per lookup (leaf data source)
    ↓
[CIL Tree Walk]  → Traverse each CIL tree to generate Go code
    ↓
[Generated Code] → Go code that loads and registers lookup data
    ↓
[Runtime]        → Execute: metric calculations access lookup tensors by key
```

### CIL Production (Analysis → CIL)

After the symbol table is walked and AnalysisDefinitions are assembled, the CIL generation process produces a **CILLookupNode** for each lookup. These are single-node trees (leaf data source nodes), structurally similar to connector CIL nodes:

- The node captures the lookup name, dimensions, grain (if any), and celltype
- The node serves as a data source for metric CIL trees that reference the lookup

### CIL Tree Walking (CIL → Synthesis)

Synthesis walks each CIL lookup node and emits the corresponding Go code to load the lookup data from the source specified in the Lookup IO Specification and register it as a `MetricComputation`. Metric CIL trees that reference the lookup will emit `controller.GetMetricComputation(lookupName)` to access the data at runtime.

## Validation Rules

The following validation rules must be enforced during parsing and analysis:

1. **Name Uniqueness**: Lookup names must be unique within a compilation unit.
2. **Dimension Existence**: All concept names in the `dimensions` clause must refer to existing concepts.
3. **Grain Validity**: When a `grain` clause is present, every measure name must refer to an existing measure whose type is `Interval` or derived from `Interval`.
4. **Celltype Existence**: The measure name in the `celltype` clause must refer to an existing measure.
5. **Lookup Reference Existence**: When a metric declares a `lookup` input, the lookup name must refer to an existing lookup.
6. **Dimensional Compatibility**: The lookup's dimensions must be compatible with the dimensions of the metric that references it, as determined by the dimensional analysis phase.

## Parse/Analysis Separation

Following the [Symbol Table-Driven Parsing Pattern](../../patterns/symbol-table-driven-parsing.md), lookup processing maintains a strict separation between parsing and analysis phases:

### Parser Responsibilities (Phase 1)

- Recognize `lookups` sections and lookup declaration syntax
- Build raw parser lookup definition structures from the parse tree
- Store definitions in `SymbolDefinition.Data` within the symbol table tree
- The parser does **not** validate cross-references, resolve dimension names, or check semantic rules beyond syntax

### Analyzer Responsibilities (Phase 2)

- Walk the symbol table tree to assemble `AnalysisLookupDefinition` structures
- Resolve dimension, grain, and celltype references against the complete set of definitions
- Validate semantic rules (name uniqueness, reference existence, grain type constraints)
- Verify dimensional compatibility with referencing metrics

## Incremental Implementation Strategy

The implementation follows the same incremental approach as other CPL declarations: each increment is end-to-end across parse → analyze → synthesize.

### Increment 1: Basic Lookup Declarations Without Grain

**Parse:**
- Parse `lookups` sections with lookup names and `dimensions` clause
- Parse `celltype` clause referencing a measure name
- Store raw definitions in the symbol table

**Analyze:**
- Walk symbol table to assemble AnalysisDefinitions with LookupInfo entries
- Validate that dimension concept names exist
- Validate that celltype measure name exists

**CIL Generation:**
- Generate one `CILLookupNode` leaf tree per lookup (no time dimension in this increment)

**Synthesize:**
- Generate Go code to load lookup data from the Lookup IO Specification
- Register each lookup as a `MetricComputation` on the controller
- Include a working end-to-end example with a timeless lookup

### Increment 2: Lookups with Grain

**Parse:**
- Parse optional `grain` clause with one or more measure names

**Analyze:**
- Validate that grain measure names exist and have `Interval`-derived types

**CIL Generation:**
- Include grain dimensions in `CILLookupNode`

**Synthesize:**
- Generate Go code that loads temporal lookup data and indexes by time as well as concept dimensions

### Increment 3: Lookup Integration with Metric Calculations

**Parse/Analyze:** (no new syntax — lookup declarations already supported)

**Synthesize:**
- Generate metric calculation code that references lookups via `controller.GetMetricComputation`
- Verify unit compatibility between lookup celltypes and the calculations that use them

## Design Benefits

1. **Reuse of MetricComputation Interface**: Lookups use the same `MetricComputation` mechanism as connectors and computed metrics, reducing implementation complexity.
2. **Compile-Time Unit Verification**: The celltype measure's compound units allow the framework to verify dimensional correctness before runtime.
3. **Clear Semantic Distinction**: The keys-vs-shape distinction between lookups and connectors prevents confusion about how each declaration type should be used.
4. **Symmetric Grain Support**: Supporting multiple grain measures in lookups (as in connectors) preserves language symmetry and eases learning.

## Related Documents

- [CPL Requirements — Lookups](../../requirements/cpl/lookups-syntax.md) — Syntax requirements for lookups
- [CPL Requirements — Connectors](../../requirements/cpl/connectors.md) — Syntax requirements for connectors (structurally similar)
- [CPL Requirements — IO Specifications](../../requirements/cpl/io_specs.md) — IO specification requirements (separate lookup IO spec)
- [Metrics Design](metrics-design.md) — Metric design (metrics reference lookups as inputs)
- [Concept Definitions Design](concepts-definitions.md) — Concept design (lookup dimensions reference concepts)
- [Units and Measures Design](units-and-measures.md) — Unit/measure design (lookup celltypes are measures with compound units)

## Future Considerations

- Dynamic lookup data loaded from external sources at synthesis time
- Default value declarations for key-miss handling
- Unification of lookup and connector IO specifications if they prove structurally identical
- Incremental reload of temporal lookups when source data changes
