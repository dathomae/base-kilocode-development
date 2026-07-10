# CPL Architecture - Introduction

## Overview

The Chronocone Planning Language (CPL) is a declarative language for defining metrics, data sources, calculations, and dimensional relationships. CPL files use the `.cpl` extension and are processed through a four-phase lifecycle that transforms declarative definitions into executable planning logic.

## CPL File Structure

A `.cpl` file is organized as a series of statements that may include the following sections:

### Concepts Declaration Section

The `concepts` section declares conceptual categories and their hierarchical relationships. Concepts define the dimensional structure of metrics and may include:

- A set of ordered values (e.g., `Country=[USA, Canada, Mexico]`)
- Parent-child relationships with other concepts (e.g., `Region` as a child of `Country`)

This section may appear multiple times within a file. When included multiple times, the declared concepts are cumulative and need not be defined before they are used. All concept names within a file must be unique.

### Types Declaration Section

The `types` section declares data types derived from Chronocone's ground types (`Number`, `String`, `Instant`, `Interval`, `Boolean`, `Nil`). Derived types may specify constraints such as precision for numbers or maximum length for strings. This section may appear multiple times within a file.

### Units Declaration Section

The `units` section declares named unit types and conversion rules between units. This section may appear multiple times within a file. When included multiple times, the declared units are cumulative and need not be defined before they are used. All unit names within a file must be unique.

### Measures Declaration Section

The `measures` section declares measure groups, which combine a type with units. Measures define the logic for converting between different units of the same conceptual type. This section may appear multiple times within a file.

### Connector Declarations

Zero or more `connector` declarations specify how to import data from external sources. Connectors define the mapping between external APIs or data sources and internal metrics or lookups.

### Metric Declarations

Zero or more `metric` declarations define calculated values. Metrics specify formulas, input dependencies, grain, and dimensional constraints.

### Aggregator Declarations

Zero or more `aggregator` declarations define transformations that align data across different dimensional granularities. Aggregators specify how to roll up or summarize values from one set of dimensions to another.

### Spreader Declarations

Zero or more `spreader` declarations define transformations that distribute values from a coarser granularity to a finer granularity. Spreaders are the inverse of aggregators.

### Lookup Declarations

Zero or more `lookup` declarations define reference tables that map between different concept spaces. Lookups provide static or slowly-changing reference data.

## Processing Lifecycle

The Chronocone Planning Language is processed through a four-phase lifecycle that transforms declarative `.cpl` files into executable planning logic. Each phase builds upon the output of the previous phase, gradually refining the representation from source code to runnable form.

## The Four Phases

### 1. Parsing / Syntax Analysis Time

During the parsing phase, `.cpl` files are read and validated for syntactic correctness. The parser processes the source text and produces:

- An **Abstract Syntax Tree (AST)** representing the structure of declarations
- **Symbol tables** capturing names and their scopes across the file
- **Parser data structures** that capture individual elements associated with parse tree nodes and are pushed onto symbol table entries for the symbol table on top of the stack
- Syntax error reports for malformed constructs

The output of this phase is a decorated AST with complete symbol resolution, ready for semantic analysis.

### 2. Analysis Time

During the analysis phase, relationships declared in the CPL source are validated against both the source itself and the taxonomy database. In addition to the definitions read from the database and derived from parsing
CPL files, the compiler takes in an output specification that lists the desired output metric(s), their dimensions, types, time grain and forecast and actuals length. Part of analysis is determining which metrics and inputs need to be included to produce the desired output and which aggregators, spreaders,
unit or measure conversions and so on must be included to produce the specified output metric(s). If it is not possible to produce the desired
metric(s) the analyzer will give advice about why it isn't possible and
recommendations for possible additions to the CPL definitions that would make it possible. These are intended for both humans and AI agents.

This phase ensures that plans make semantic sense by checking:

- References to taxonomy terms exist and are used correctly
- Built-in types are applied appropriately
- Concept relationships are consistent
- No cycles are introduced in metric calculations
- All required metrics can be derived with consistent dimensions
- Units and measures are compatible in calculations

**AnalysisDefinitions:** After parsing is complete, the symbol table tree is walked to gather up definitions for each entity type (concepts, units, types, measures, metrics, aggregators, spreaders, and lookups) and stored in a more natural form. There will be:

- A **concepts lookup map** for the analyzer to look up concept definitions
- A **types lookup map** for type definitions
- A **units lookup map** for unit definitions
- A **measures lookup map** for measure definitions
- And so on for metrics, aggregators, spreaders, and lookups

Collectively these are known as **AnalysisDefinitions**.

The output of this phase is a **Data and Control Flow Graph** oriented around a set of internal operators. This graph represents the intermediate form of the plan in an abstract operations language used by the synthesis phase.

### 3. Synthesis Time

During the synthesis phase, the validated intermediate representation is transformed into an executable form. The synthesis process:

- Walks the Data and Control Flow Graph
- Maps abstract operators to concrete implementation constructs
- Generates **CILDefinition trees** - data structures that use the Chronocone Intermediate Language (CIL) defined in the architecture/cpl/analysis directory
- Generates code in the target language

**CILDefinition Types:** After the AnalysisDefinition objects are assembled, data trees are generated using the CIL intermediate language. One tree is generated for each metric. These CILDefinition trees capture the computation in a language-agnostic form.

**Target Language Generation:** The CILDefinition trees are walked to generate the executable program. There will eventually be synthesizers for multiple target languages; the first will be for Go.

While the synthesized form could theoretically be any representation, the first version of the system will generate Go code suitable for compilation and execution.

### 4. Evaluation Time

During the evaluation phase, the synthesized form is executed to produce planning outputs by running the compiled executable code. The synthesize program will read inputs

**Key architectural point:** Since concept definitions (both relationships AND values) are known at compile, analysis, synthesize, and runtime, the synthesized executable can make optimizations based on the full concept structure without needing to defer concept value resolution to runtime.

The output includes calculated metric values, any runtime errors encountered, and potentially intermediate results for debugging or auditing purposes.

## Phase Relationships

```
.cpl Files → [Parsing] → AST + Symbols + Parser Data Structures
                   ↓
           [Symbol Table Walk] → AnalysisDefinitions
                   ↓
           [Analysis] → Data/Control Flow Graph
                   ↓
           [CIL Generation] → CILDefinition Trees
                   ↓
           [Synthesis] → Generated Code (Go)
                   ↓
           [Evaluation] → Results
```

## Data Structure Organization

The CPL architecture organizes data structures across the processing phases as follows:

### Phase 1: Parser Data Structures

- Associated with the parser
- Capture individual elements from parse trees
- Pushed onto symbol table entries for the symbol table on top of the stack
- Tied to specific AST nodes and parse contexts

### Phase 2: AnalysisDefinitions

After parsing completes, the symbol table tree is walked to assemble:

- **Concepts lookup map** - concept definitions with relationships and values
- **Types lookup map** - type definitions and constraints
- **Units lookup map** - unit definitions and conversion rules
- **Measures lookup map** - measure group definitions
- **Metrics lookup map** - metric definitions
- **Aggregators lookup map** - aggregator definitions
- **Spreaders lookup map** - spreader definitions
- **Lookups lookup map** - lookup table definitions

### Phase 3: CILDefinition Types

After AnalysisDefinitions are assembled:

- One CIL tree generated per metric. When multiple metrics are involved, a CIL operation tree may reference the CIL tree of another metric.
- Uses the Chronocone Intermediate Language
- Language-agnostic representation of calculations
- Captures operations, dimensions, and data flow

### Phase 4: Synthesized Code

- Target language specific (Go initially)
- Executable form of the calculation
- Optimized based on known concept values and relationships

Each phase is designed to be independently testable and debuggable. Errors are detected as early as possible, with clear messaging to help users correct issues in their CPL definitions.

## Concept Definition Availability

A key architectural principle is that **concept definitions (both relationships and values) are known at compile, analysis, synthesize, and runtime**. This enables:

- Static validation of concept references at analysis time
- Optimization of data structures based on known concept value sets
- Pre-computation of concept hierarchies and relationships
- More efficient synthesized code that doesn't need to dynamically discover concept values

This is a change from earlier designs where concept values were only known at runtime.
