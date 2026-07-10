# ERO004: Connectors Incremental Implementation

## Title

Connectors Incremental Implementation

## Purpose/Goal

Define the incremental approach for implementing connectors functionality across all CPL phases (parsing, analysis, synthesis, evaluation), detailing the minimal functionality needed to support a connector declaration and reading data from a parquet file in the synthesized program.

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic004_cpl-connectors-implementation.md](../../../../epics/Epic004_cpl-connectors-implementation.md) (To be created)

## Context

### Situation

Connectors define the shape of input data coming into the Chronocone system. A connector declaration specifies the dimensions, grain, and cell type of the input data. The actual connection details (e.g., reading from a parquet file) are provided via an IO Specification passed to the CPL compiler.

To avoid overwhelming complexity, we need an incremental approach that focuses on the minimal functionality required to parse a connector declaration, analyze it, generate its CIL representation, and synthesize Go code that reads data from a parquet file using Apache Arrow.

### Constraints

- The implementation must support reading data from parquet files using Arrow in the synthesized code.
- The design must align with the existing CPL phase architecture (parse → analyze → synthesize).
- The CIL representation must treat connectors as metrics with fixed dimensions.
- Other parts of CIL generation and synthesis support library (metrics, aggregators, spreaders) will be built out later.

### Requirements

- Parse `connectors` sections and declarations.
- Validate connector dimensions, grain, and cell type during analysis.
- Generate CIL representation for connectors.
- Synthesize Go code to read parquet files based on the connector and IO specification.

## Evaluation Criteria

1. **Minimal Viable Product** (High importance): The first increment must provide end-to-end functionality for reading parquet data based on a connector declaration.
2. **Extensibility** (Medium importance): The design must allow for future additions like other data sources or complex metric calculations.
3. **Consistency** (High importance): Data structures should evolve consistently across all phases.

## Decision

**Chosen Option:** Incremental Implementation by Feature

**Rationale:**
We will focus on the minimal functionality needed to support a connector declaration and reading data from a parquet file. Future increments will build out other parts of the CIL generation and synthesis support library as metric, aggregator, and spreader functionality is built out.

### Increment 1: Basic Connector Declaration and Parquet Reading

**Features Implemented:**

- **Parsing:** Support for the `connectors` keyword and basic connector declarations (dimensions, celltype).
- **Analysis:** Validation that referenced dimensions (concepts) and celltype (measure) exist in the symbol table.
- **CIL Generation:** Representation of the connector as a base metric node in the CIL tree.
- **Synthesis:** Generation of Go code that uses the `go-arrow` library to read data from a parquet file specified in the IO Specification, mapping the parquet columns to the connector's dimensions and celltype.

**Limitations:**

- Only supports reading from parquet files.
- No complex metric calculations, aggregators, or spreaders (to be added in future increments).
- Assumes the parquet schema matches the connector declaration.
- No time dimension handling.

**Value Provided:**

End-to-end ability to define an input data shape in CPL and synthesize code that reads matching data from a parquet file into Chronocone's internal tensor representation.

### Increment 2: Time Handling

**Features Implemented:**

- **Parsing:** Support for the `grain` keyword in connector declarations.
- **Analysis:** Validation that the grain refers to an existing measure, and that the measure's underlying type is `Interval` or derived from `Interval`. Validation that `Interval` types reference a valid unit (a `Duration` concept value) via the regular unit and measure mechanisms.
- **CIL Generation:** Representation of the time dimension in the CIL tree.
- **Synthesis:** Generation of Go code that maps the parquet time column to the connector's grain.

**Limitations:**

- Assumes the parquet schema matches the connector declaration including the time column.

**Value Provided:**

Ability to handle time-series data in connectors, mapping time columns to Chronocone's internal time representation.

## Implications

### Positive Implications

- Rapid delivery of core data ingestion capability.
- Validates the end-to-end pipeline from CPL parsing to synthesized Go code execution.
- Establishes the pattern for integrating external data sources.

### Negative Implications / Trade-offs

- Initial implementation is tightly coupled to parquet/arrow.
- Error handling for schema mismatches between parquet and CPL might be basic initially.

## Notes

**Phase-Specific EROs:**

- Parsing Phase ERO: [ERO004_connectors-parsing-phase.md](../../go-parser-target/eros/ERO004_connectors-parsing-phase.md)
- Analysis Phase ERO: [ERO004_connectors-analysis-phase.md](../../go-analysis/eros/ERO004_connectors-analysis-phase.md)
- Synthesis Phase ERO: [ERO004_connectors-synthesis-phase.md](../../go-synthesis-target/eros/ERO004_connectors-synthesis-phase.md)
