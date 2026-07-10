# Story 107: CPL Metric Data Structure Design

## Overview

Design the in-memory data structure for representing metrics, including its translation to and from metadata definitions in the database and Parquet files for metric storage. This is a design-only story that produces documentation and specification artifacts.

## References

- [CPL Language Examples](../specifications/requirements/metrics-definition-examples-2a.md) - Metric examples and use cases
- [Story106_cpl-parser-metrics.md](Story106_cpl-parser-metrics.md) - Metric parsing
- [Story015_timeseries-parser-go.md](Story015_timeseries-parser-go.md) - Related time series object parsing

## Dependencies

### Prerequisites

- [Story106_cpl-parser-metrics.md](Story106_cpl-parser-metrics.md) - Not Started (requires understanding of metric AST types)

### Dependents

- [Story108_cpl-metric-evaluation.md](Story108_cpl-metric-evaluation.md) - Not Started

## Tasks

1. Analyze metric requirements - Not Started
   a. Document dimensional structure requirements (dimension order, values)
   b. Document time series requirements (timestamps, values, grains)
   c. Document actual vs forecast data representation
   d. Document multi-metric storage requirements
   e. Review existing time series implementations (Story015)

2. Design in-memory metric structure - Not Started
   a. Define core data structures for metric representation
   b. Design dimension indexing mechanism
   c. Design time series data layout
   d. Design metadata attachment approach
   e. Document design rationale and trade-offs

3. Design database metadata mapping - Not Started
   a. Map metric structure to relational schema tables
   b. Define serialization format for dimension definitions
   c. Define serialization format for calculation expressions
   d. Document foreign key relationships
   e. Document versioning considerations

4. Design Parquet file mapping - Not Started
   a. Define Parquet schema for metric storage
   b. Design column layout for efficient querying
   c. Define custom metadata encoding
   d. Document compression and encoding choices
   e. Document partitioning strategy

5. Design DAG representation for metric dependencies - Not Started
   a. Define node types for the metric dependency graph
   b. Design edge representation for metric references
   c. Define cycle detection approach for invalid circular dependencies
   d. Document topological sorting for evaluation order
   e. Design handling of self-referential metrics (time-series DAG)

6. Design DAG calculation approach - Not Started
   a. Define evaluation strategy for dependency-ordered metrics
   b. Design approach for incremental recalculation
   c. Document handling of missing or null input values
   d. Define approach for parallel vs sequential evaluation
   e. Document caching strategy for intermediate results

7. Design transformation interfaces - Not Started
   a. Define interface for AST to in-memory conversion
   b. Define interface for in-memory to database conversion
   c. Define interface for in-memory to Parquet conversion
   d. Document error handling approach
   e. Define interface for DAG construction from metric declarations

8. Produce design documentation - Not Started
   a. Create `docs/metric-data-structure.md` with complete design
   b. Include diagrams for data flow and relationships
   c. Include diagrams for DAG structure and evaluation flow
   d. Include example code showing usage patterns

## Decisions

1. **Design-only scope**: This story produces documentation and specifications, not implementation code. Implementation occurs in Story108 and subsequent stories.

2. **Multi-format support**: The design must support three representations: in-memory (Go structs), database (relational tables), and file (Parquet format).

3. **Performance considerations**: The design prioritizes efficient access patterns for planning use cases, including dimension-based filtering and time range queries.

## Notes

- This story is a placeholder for design work that must be completed before metric evaluation can be implemented.
- The design should consider alignment with existing time series parsing work in Story015.
- Database schema design should reference the relational store requirements from Story027.
- Parquet design should consider the fastparquet API capabilities documented in the resources.
- The DAG design must handle both inter-metric dependencies (one metric referencing another) and self-referential metrics (time-series dependencies like `inventory = lag(inventory,1) + arrivals`).
- Cycle detection is required to prevent invalid circular dependencies between metrics.
