# Apache Arrow Serialization

Columnar data serialization for inter-program communication and file export.

## Responsibilities

- Arrow IPC (inter-process communication) format for Flink/Spark data exchange
- Parquet file generation for batch export and analytics
- Unified Model hydration from Arrow arrays
- Memory-efficient columnar representation

## Package Boundaries

_(none documented — depends on Apache Arrow Go libraries.)_

## Use Cases

- Data transfer between Flink and Spark jobs
- Batch export to customer analytics pipelines
- Time-series and aggregation data in columnar format

## When you work here

- Arrow schemas must align with Unified Models
- Document any memory or performance implications of columnar vs. row-oriented formats
- Provide clear examples for FDEs integrating with external analytics tools
