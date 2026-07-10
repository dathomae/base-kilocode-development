# Metrics Design

## Overview

Metrics are a fundamental building block of the Chronocone Planning Language (CPL). They define calculated time series with inputs, formulas, and conditional logic. This document outlines the design for how metric definitions are captured, stored, and used throughout the CPL processing pipeline.

A metric has no inherent cell type, dimensions, forecast or actuals length, or time grain associated with the declaration itself. Those properties are determined by the IO specification supplied to the CPL compiler at synthesis time (see [io_specs.md](../../requirements/cpl/io_specs.md)).

## Key Design Principles

1. **Declarative over Imperative**: Metrics specify _what_ to compute, not _how_. The calculation body is a single functional expression with no variables or side effects.
2. **Explicit Inputs**: All dependencies on other metrics, lookups, and dimensions are declared in an `inputs` block.
3. **Dimensional Awareness**: Every value carries dimensional context. Metrics reference concepts for their dimensional structure.
4. **Composability**: Metrics can reference other metrics, enabling complex calculation chains that form a Directed Acyclic Graph (DAG).
5. **Compile-Time Resolution**: Metric definitions (structure, expressions, dependencies) are known at compile, analysis, synthesize, and runtime. This enables static validation, optimization, and dead-code elimination.

## Core Rules for Metrics

### 1. Metrics Sections

Metrics are defined within `metrics` sections. A CPL file may contain one or more `metrics` sections, and metric definitions may be spread across multiple files within the same compilation unit.

The general form of a metrics section is:

```
metrics
    <list of metric definitions>
```

### 2. Metric Identity

A metric name in one metrics section always refers to the same metric as the same name in a different metrics section. Metric names are globally unique within a namespace.

### 3. Metric Declaration Structure (Long Form)

A metric declaration has the following structure:

```
<metric name>
    inputs
        <list of input declarations>
    calculate
        <calculation expression>
```

Both `inputs` and `calculate` must be indented one tab deeper than the metric name. The `inputs` sub-declaration lists all dependencies. The `calculate` sub-declaration contains the business logic.

### 4. Short Form Metrics

Metric definitions may be supplied in an abbreviated short form:

```
<metric name> = <calculation>
```

Where the calculation uses unaliased metric, lookup, and dimension names. No `where` or `backward`/`forward` clauses are permitted in short form. For example:

```
revenue = demand_forecast * Price
```

Is equivalent to:

```
revenue
    inputs
        metric demand_forecast
        dimension Price
    calculate
        demand_forecast * Price
```

Name resolution for short form:

1. If the name matches a metric declared in the CPL specification, it is treated as a `metric` input.
2. If the name matches a lookup declared in the CPL specification, it is treated as a `lookup` input.
3. If the name matches one of the dimensions of the metric (as specified in the IO specification), it is treated as a `dimension` input.
4. A name collision between a lookup, a dimension, or a metric in a short form calculation is a compile-time error.

### 5. Metric Inputs

The `inputs` sub-declaration lists all dependencies. The general form of an input declaration is:

```
<metric|lookup|dimension> <name> [as <alias>] [where <boolean expression>] [backward(<constraint>=<expression>) | forward(<constraint>=<expression>)]
```

#### Input Kinds

- **`metric`** — A reference to another metric defined in the CPL specification.
- **`lookup`** — A reference to a lookup defined in the CPL specification.
- **`dimension`** — A reference to one of the dimensions of the metric being calculated. Dimension inputs make that dimension compulsory for the metric (it cannot be collapsed by aggregators prior to this metric's calculation).

#### Aliases

The `as` clause is optional. If absent, the name of the input is its own alias. No two aliases may be the same within a single metric's input block.

#### Where Clauses

The `where` clause filters or constrains the input. The only select-style operator currently defined is `aspect=<aspect>`, where `<aspect>` is one of:

- `ACTUALS` — Filter the input to actual (historical) data points only.
- `FORECAST` — Filter the input to forecast data points only.

Multiple `where` conditions may be joined using the boolean operators `and`, `or`, and `not`.

#### Backward and Forward (Time Shifting)

The optional `backward(<constraint>=<expression>)` or `forward(<constraint>=<expression>)` clause shifts the input metric in time. The `<constraint>` identifies the time measure (e.g., `weeks`, `days`) and the expression provides the shift amount. Without a `backward` or `forward` clause, each input refers to the "current" cell (the cell at the same time coordinate as the cell being calculated).

Example:

```
inventory
    inputs
        metric order_arrivals
        metric inventory as prior_inventory backward(weeks=1)
    calculate
        ordered_arrivals + prior_inventory
```

### 6. Self-References

Metrics may reference themselves in calculations, but only through an alias that uses a `backward` or `forward` clause. Direct self-reference in a `calculate` clause (naming the metric itself without an alias) is prohibited because an unaliased input always refers to the current cell, which would create a circular dependency at the cell level. The only valid self-reference is to an earlier or later cell, obtained via an alias.

### 7. Metric Calculation

The `calculate` clause specifies the business logic. The body consists of a single expression. The result of that expression is the value of the metric for the cell being calculated. This is a purely functional notation: there are no variables or side effects.

The `calculate` clause supports:

- `if` / `else if` / `else` / `end if` conditional expressions
- `when` / `end when` alternative conditional syntax
- Common arithmetic operators (`+`, `-`, `*`, `/`) for their respective types
- Comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`)
- Logical operators (`and`, `or`, `not`)
- Signed numeric constants and string constants
- `Nil` literal
- Parenthesized expressions
- References to input aliases (from the `inputs` block) and dimension names

The `calculate` clause must produce exactly one result under all possible conditions; otherwise it is a compile-time error.

### 8. Truthiness Rules

In boolean contexts (conditions of `if`, `else if`, `when`, and operands of logical operators):

- `Nil` is equivalent to `False`.
- The boolean value `False` is falsy.
- For numeric values, zero is falsy; any non-zero value is truthy.
- For strings, the empty string is falsy; any non-empty string is truthy.
- All other non-Nil values are truthy.

### 9. Nil in Arithmetic

`Nil` cannot participate in arithmetic expressions. Any attempt to use `Nil` as an operand of `+`, `-`, `*`, or `/` is a compile-time or runtime error.

### 10. When Syntax

An alternative conditional form reduces visual clutter:

```
when
    <boolean condition> => <expression>
    <boolean condition> => <expression>
    ...
    default => <expression>
end when
```

`when` is interpreted identically to a chain of `if` / `else if` / `else`: the first line is an `if`, subsequent lines are `else if`, and the `default` line is `else`. Fall-through is not allowed, and the result must be fully specified. `default` is a reserved keyword in CPL and cannot be used as an identifier.

## Relationship with IO Specification

Metrics defined in CPL are partial specifications. The IO specification provides:

- **Inputs**: Connector-to-data-source mappings (format, driver, parameters)
- **Outputs**: For each desired output metric — its dimensions (concept names), celltype (measure name), and grain (interval measure names)
- **Architecture**: Compute configuration (CPU cores, cache size, main memory size)

The IO specification is supplied at synthesis time and determines:

- Which metrics are output targets
- The dimensional and temporal shape of each output metric
- Which connectors and data sources supply input data
- How the generated executable will be configured for its execution environment

## Compilation Scope

The CPL compiler generates code only for the calculations required to produce the outputs specified in the IO specification. Metrics that are not named in the IO specification and are not direct or indirect inputs to any metric named in the IO specification are excluded from compilation.

## Metric Definition Types

The following definition types capture metric information at different phases of processing:

### 1. Parser Metric Definition

During parsing, metric definitions are captured from the parse tree and stored in the symbol table:

- Raw metric name, inputs, and calculation expression as parsed
- Input declarations with aliases, where clauses, and backward/forward clauses
- Source location information for error reporting

These definitions are stored in `SymbolDefinition.Data` during parsing.

**Expression Handling via Postfix Notation:** The parser does **not** build an expression AST tree. Instead, it follows the [Postfix Expression Parsing Pattern](../../patterns/postfix-expression-parsing.md): as ANTLR `Exit*` listener methods are called in precedence order for expression rules, the parser pushes `SymbolDefinition` entries for operands (identifiers, literals) and operators (arithmetic, comparison, logical) onto the current scope's `Definitions` slice. This naturally converts the infix expression to postfix (reverse Polish) notation through the listener call order.

For example, the expression `a + b * c` produces Definitions in the order: `["a", "b", "c", MULT, PLUS]` — postfix notation for `a + (b * c)`.

The analysis phase then performs an iterative pop/evaluate loop on this postfix sequence to generate the CIL expression tree.

### 2. AnalysisDefinitions — Metric Lookup Map

After parsing completes, the symbol table is walked to assemble the **Metrics lookup map** as part of AnalysisDefinitions:

- **MetricInfo** — Core metric information including name, resolved input references, and validated expression tree
- **InputResolution** — Verified references to metrics and dimensions
- **ExpressionValidation** — Type-checked and validated expression tree

The Metrics lookup map enables the analyzer to:

- Validate metric references in calculations
- Check dimensional compatibility
- Resolve all metric inputs for CIL generation

**Dependency DAG construction is deferred to the second analysis pass**, after all definitions from all sources (all CPL files, database concepts, etc.) have been assembled. Only once the complete set of definitions is known can the dependency graph be reliably built and circular dependencies detected (excluding valid self-references with time shifting).

### 3. CILDefinition — Metric Nodes

In the Chronocone Intermediate Language (CIL), metrics are represented as:

- **CILMetricNode** — A node in the CIL tree representing a metric computation
- **CILExpressionNode** — Expression nodes representing arithmetic, comparison, logical, and conditional operations
- **CILInputRef** — References to input metrics or dimensions
- **CILTimeShift** — Representation of backward/forward time-shifting operations

Since concept values and metric structures are known at synthesis time, the CIL can include:

- Pre-computed expression trees
- Resolved dependency chains
- Known tensor shapes based on IO specification dimensions

### 4. Synthesized Code — Metric Computation

In the generated executable code:

- Each metric becomes a computation function or method
- Dependencies are resolved in topological order
- Expressions are translated to target language operations
- Conditional logic maps to target language control flow
- Time-shifting operations map to array index offset calculations

## Data Flow

The CPL compiler processes metrics through a strict progression of representations:

```
CPL Source
    ↓
[Parser] → ParserMetricDefinition structures stored in SymbolDefinition.Data
    ↓
[Symbol Table Walk] → AnalysisDefinitions with MetricInfo entries
    ↓
[CIL Generation] → One CIL tree per metric (nodes represent operations)
                    One single-node CIL tree per connector (leaf data source)
    ↓
[CIL Tree Walk]  → Traverse each CIL tree to generate Go code
    ↓
[Generated Code] → Executable Go program
    ↓
[Runtime]        → Execute: process cells through metric calculations
```

### CIL Production (Analysis → CIL)

After the symbol table is walked and AnalysisDefinitions are assembled, the CIL generation process produces a **Chronocone Intermediate Language (CIL) tree** for each metric and each connector:

- **Connector CIL trees**: Single-node trees representing the data source. The node captures the connector name, dimensions, celltype, and grain. These are leaf nodes that provide the raw input data for the computation.
- **Metric CIL trees**: Trees whose nodes represent operations in the calculation. A metric that simply returns `0` produces a single-node tree (a literal). A metric with arithmetic produces a tree with binary operation nodes. A metric referencing other metrics has child CIL tree references.

CIL trees are the language-agnostic intermediate representation — they capture what to compute without being tied to a specific target language.

### CIL Tree Walking (CIL → Synthesis)

Synthesis generates target language code by **walking each CIL tree**. The walker visits each node and emits the corresponding target language construct:

- A `CILInputRef` node emits code to access the input metric's data for the current cell
- A `CILLiteral` node (e.g., `0`) emits a constant
- A `CILBinaryOp` node emits the target language operator applied to its walked children
- A `CILConditional` node emits if/else control flow

**This same walking mechanism is used for all increments** — even Increment 1 where the CIL tree is just a single literal node. The walker processes it identically to any other tree, ensuring the full pipeline is exercised from the start.

## Expression Handling

### Parser Phase: Postfix via Symbol Definitions

The parser does **not** build an expression tree. Instead, it follows the [Postfix Expression Parsing Pattern](../../patterns/postfix-expression-parsing.md): as ANTLR `Exit*` listener methods visit expression components in operator precedence order, the parser pushes `SymbolDefinition` entries onto the current scope's `Definitions` slice:

- **Operands**: Identifiers (input references, dimension names), numeric/string/boolean/Nil literals
- **Operators**: Arithmetic (`+`, `-`, `*`, `/`), comparison (`==`, `!=`, `<`, `>`, `<=`, `>=`), logical (`and`, `or`, `not`)

For example, `a + b * c` results in Definitions: `[a, b, c, *, +]` — postfix for `a + (b * c)`.

For conditional expressions (`if`/`else`/`end if`, `when`/`end when`), the condition uses the same postfix mechanism in the parent scope, while each branch body gets its own child scope in the symbol table — the tree structure encodes the conditional hierarchy rather than using flat marker tokens.

### Analysis Phase: Postfix → CIL Tree

The analysis phase walks the Definitions slice in order and performs a stack-based postfix evaluation to construct the CIL expression tree:

```
for each def in Definitions:
    if def is operand: push onto stack
    if def is operator: pop operands, create CIL node, push result

The final CIL tree is a single root node on the stack.
```

The resulting CIL tree nodes represent:

```
CILLiteral        — numeric constant, string constant, Nil, boolean
CILInputRef       — alias pointing to a metric or dimension input
CILBinaryOp       — left CILExpr, operator, right CILExpr
CILUnaryOp        — operator, operand CILExpr
CILConditional    — condition, trueBranch, falseBranch CILExprs
CILWhenClause     — list of branches, default CILExpr
```

### Synthesis Phase: CIL Tree → Target Code

The CIL tree is walked to generate target language (Go) code. Each CIL node type emits the corresponding Go construct: literals become constants, binary ops become Go operators, conditionals become `if`/`else` blocks.

## Metric Dependency DAG

Metrics form a Directed Acyclic Graph (DAG) based on input dependencies. Key properties:

- **Nodes**: Each metric is a node in the graph
- **Edges**: If metric A references metric B as an input, there is a directed edge from B to A
- **Topological Order**: The DAG determines evaluation order — inputs are calculated before dependent metrics
- **Cycle Detection**: Circular dependencies (excluding valid self-references with backward/forward) are compile-time errors
- **Self-References**: A metric referencing itself via a backward/forward clause does not create a cycle; the time shift breaks the circularity at the cell level

```
┌──────────────┐     ┌──────────────┐
│ demand_forecast │───▶│  inventory   │◀── backward(weeks=1)
└──────────────┘     └──────┬───────┘
                            │
                     ┌──────▼───────┐
                     │ revenue      │
                     └──────────────┘
```

## Incremental Implementation Strategy

The implementation follows an incremental approach where **each increment is end-to-end** across all three CPL phases: parse → analyze → synthesize. Every increment produces a working executable that can process CPL files with accompanying IO specifications and test data.

### Increment 1: Basic Metric Declarations with Identity Calculation

**Parse:**

- Parse `metrics` sections with metric names
- Parse `inputs` block with simple `metric` and `dimension` name references (no `as` alias, no `where` clause, no `backward`/`forward`)
- The `calculate` clause contains the literal `0` (identity/trivial calculation)

**Analyze:**

- Walk symbol table to assemble AnalysisDefinitions with MetricInfo entries
- Validate that input names refer to declared metrics and dimensions

**CIL Generation:**

- Generate one CIL tree per metric. For Increment 1, each metric's CIL tree is a single `CILLiteral` node with value `0`
- Generate one single-node CIL tree per connector (leaf data source node capturing connector name, dimensions, celltype, grain)

**Synthesize:**

- Walk each CIL tree using the standard CIL tree walker to generate Go code
- The literal `0` node emits a constant `0` in the generated Go code — the walker processes it identically to any other node type
- Connector CIL trees emit Go code to read parquet data via the Arrow library
- Include a working end-to-end example with a CPL file, IO specification, and test data

### Increment 2: Simple Arithmetic in Calculate with IO Specification Acceptance

**Parse:**

- Parse `calculate` clause with basic arithmetic (`+`, `-`, `*`, `/`), parenthesized expressions, numeric and string literals
- Input references from Increment 1 used as operands in arithmetic expressions

**Analyze:**

- Validate that all referenced input names exist
- Type-check arithmetic operands
- Accept and validate full IO specification — verify output metric names exist, store IO spec for synthesis use

**Synthesize:**

- Generate Go code that evaluates arithmetic expressions per cell
- Read IO spec to determine output metrics
- Include working examples with simple arithmetic metrics

### Increment 3: Tensor Calculation Parallelism

**Parse/Analyze:** (no changes — IO spec already accepted in Increment 2)

**Synthesize:**

- Use compute architecture from IO spec (CPU cores, cache size, memory) to implement tensor calculation parallelism
- Partition tensor computation across multiple goroutines or CPU cores
- Include working examples demonstrating parallel tensor computation

### Increment 4: Conditional Expressions

**Parse:**

- Parse `if` / `else if` / `else` / `end if` conditional expressions
- Parse `when` / `default` / `end when` alternative syntax
- Support boolean conditions with comparison and logical operators
- Each branch body gets its own child scope in the symbol table

**Analyze:**

- Walk child body scopes to assemble conditional CIL nodes
- Implement truthiness rules for Nil, zero, empty strings
- Validate that conditional expressions are fully specified (all branches produce a result)

**Synthesize:**

- Generate Go code with conditional branching per cell
- Include working examples with conditional metric calculations

### Increment 5: Input Aliases and Where Clauses

**Parse:**

- Parse `as` alias clause on metric and dimension inputs
- Parse `where` clause with `aspect=ACTUALS` / `aspect=FORECAST`
- Support boolean combinations in `where` clauses (`and`, `or`, `not`)

**Analyze:**

- Validate alias uniqueness within a metric's input block
- Resolve aliases to their underlying metric/dimension references

**Synthesize:**

- Generate Go code that uses aliases in expressions and applies aspect filtering
- Include working examples with aliased inputs and where clauses

### Increment 6: Time Shifting (Backward/Forward)

**Parse:**

- Parse `backward()` and `forward()` clauses on metric inputs
- Support expressions as shift amounts (numeric constants, references)

**Analyze:**

- Implement self-reference support (metric referencing earlier/later cells of itself)
- Validate backward/forward time measure references

**Synthesize:**

- Generate Go code with array index offset calculations for time shifting
- Include working examples with time-shifted self-references

### Increment 7: Short Form Metrics

**Parse:**

- Parse short form syntax (`<name> = <calculation>`)
- Convert short form to equivalent long form internally

**Analyze:**

- Scan CIL tree for identifier references and resolve to metric or dimension
- Detect name collisions in short form expressions

**Synthesize:**

- Generate identical Go code to the equivalent long form
- Include working examples using short form syntax

### Increment 8: Full Synthesis with IO Specification Integration

**Parse/Analyze:** (no changes — all syntax already supported)

**Synthesize:**

- Use IO spec for output metrics' dimensions, celltype, grain, and compute architecture
- Generate the complete metric dependency DAG from the assembled definition set
- Implement dead-code elimination (only synthesize reachable metrics)
- Generate multi-metric Go executables with correct topological evaluation order
- Integration testing with compiled multi-metric executables

## Validation Rules

The following validation rules must be enforced during parsing and analysis:

1. **Name Uniqueness**: Metric names must be unique within a compilation unit.
2. **Input Existence**: All referenced metric and dimension names must exist.
3. **Alias Uniqueness**: No two aliases within a single metric's input block may be the same.
4. **Expression Completeness**: The `calculate` clause must produce a result under all possible conditions.
5. **Nil in Arithmetic**: Nil cannot be used as an operand of arithmetic operators.
6. **Circular Dependencies**: Direct or indirect circular metric references (excluding time-shifted self-references) are errors. Dependency cycle detection is deferred to the second analysis pass once all definitions are assembled.
7. **Dimension Compatibility**: Dimension references must be valid for the metric's dimensional structure (as specified in the IO specification).
8. **Short Form Name Collision**: A name that matches multiple entity types (metric, dimension) in short form is an error.
9. **Backward/Forward Constraints**: Time shift constraints must reference valid time measures.
10. **Self-Reference Validity**: Self-references must use a `backward` or `forward` clause; direct self-reference is prohibited.

## Parse/Analysis Separation

Following the [Symbol Table-Driven Parsing Pattern](../../patterns/symbol-table-driven-parsing.md), metric processing maintains a strict separation between parsing and analysis phases:

### Parser Responsibilities (Phase 1)

- Recognize `metrics` sections and metric declaration syntax
- Build raw [`ParserMetricDefinition`](memory-bank/specifications/design/cpl/metrics-design.md:196) structures from the parse tree
- Store definitions in `SymbolDefinition.Data` within the symbol table tree
- The parser does **not** validate cross-references, resolve dependencies, or check semantic rules beyond syntax

### Analyzer Responsibilities (Phase 2)

- Walk the symbol table tree to assemble [`AnalysisMetricDefinition`](memory-bank/specifications/design/cpl/metrics-design.md:207) structures
- Resolve input references against the complete set of definitions
- Validate semantic rules (name uniqueness, input existence, expression completeness)
- Build the metric dependency DAG in a **second pass** after all definitions across all sources are assembled
- Detect circular dependencies only after the full dependency graph is known

This separation ensures that parsing remains a syntactic operation and all semantic decisions are deferred until the full definition set is available.

## Design Benefits

Knowing metric definitions at all phases provides several benefits:

1. **Early Error Detection** — Invalid references, type mismatches, and circular dependencies are caught during analysis, not at runtime.
2. **Static Optimization** — Dependency ordering and dead-code elimination happen at compile time.
3. **Dimensional Validation** — Input/output dimension compatibility can be verified before execution.
4. **Expression Compilation** — Expression trees can be optimized and compiled to efficient target code.
5. **Better Tooling** — IDE support can provide autocomplete for metric names, inputs, and dimensions.

## Future Considerations

This design assumes metric definitions are static within a CPL compilation unit. Future enhancements may include:

- Aggregator integration for dimensional transformation in metric inputs
- Spreader integration for inverse dimensional transformation
- Lookup table integration for reference data in calculations
- Dynamic metric definitions loaded from external sources
- Metric versioning for evolving calculation logic
- Incremental recalculation for changed inputs
- Alternative calculation strategies (vectorized, GPU-accelerated)

## Deferred: Numeric Type Selection

The detailed design for numeric type selection — including fixed-point integer types for specified precision and overflow detection with fallback to `math/big` — is deferred to a separate design document. See [Type Selection for Numeric Calculations](type-selection.md).

In the current implementation, the constraints for types are ignored and all values are represented as floating point (`float64`). The initial increments for metrics will continue to use this approach. The type selection design will address:

1. **Fixed-point representation**: When precision is specified (e.g., `Number(precision=2)`), values will use an integer fixed-point type with the implied scale factor, avoiding floating-point rounding errors.
2. **Overflow detection via DAG analysis**: Once the metric dependency DAG is assembled, the range of possible values can be statically analyzed to determine whether native integer types suffice or whether arbitrary-precision arithmetic via Go's `math/big` package is required. When inputs fall within the safe range, native operations proceed; when outside, `math/big` is used.
3. **Sibling-function arithmetic dispatch**: Arithmetic is implemented as sibling functions keyed by numeric type (e.g., `TensorAddFloat64`, `TensorAddInt`, `TensorAddRational`, `TensorAddBigInt`, `TensorAddBigFloat`). The synthesis phase selects the correct variant based on the `measureTypes` map tracked during analysis. All functions accept `*ChronoconeTensor[any]` and internally type-assert elements, keeping the tensor type constraint at `T any` and avoiding coupling between the storage layer and numeric semantics. See [ERO005 — Increment 2](go-synthesis-target/eros/ERO005_metrics-synthesis-phase.md#increment-2-simple-arithmetic-in-calculate-with-io-specification-acceptance) for the synthesis-side design.

## Related Documents

- [CPL Architecture — Introduction](../../architecture/cpl/introduction.md) — Overall CPL processing lifecycle
- [CPL Requirements — Metrics Syntax](../../requirements/cpl/metrics-syntax.md) — Syntax requirements for metrics
- [CPL Requirements — IO Specifications](../../requirements/cpl/io_specs.md) — IO specification requirements
- [CPL Requirements — Aggregators and Spreaders](../../requirements/cpl/aggregators-and-spreaders-syntax.md) — Related dimensional transformation syntax
- [Concept Definitions Design](concepts-definitions.md) — Concept design (concepts are referenced as metric dimensions)
- [CPL Architecture — Intermediate Language](../../architecture/cpl/analysis/intermediate_language.md) — CIL specification
- [Symbol Table-Driven Parsing Pattern](../../patterns/symbol-table-driven-parsing.md) — Architectural pattern for parsing phase design
