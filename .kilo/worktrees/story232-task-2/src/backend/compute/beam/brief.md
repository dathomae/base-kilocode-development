# Beam Orchestration

Beam is the abstraction layer enabling a single CPL synthesizer to target multiple execution fabrics.

## Responsibilities

- CPL-to-Beam compilation and intermediate representation
- Horizontal scalability abstraction for Flink and Spark
- Common orchestration logic shared between fabric implementations
- Beam execution monitoring and resource coordination

## Package Boundaries

_(none documented at this level — fabric-specific boundaries live in `flink/` and `spark/`.)_

## Architecture / Supported Fabrics

- `flink/` — Execute Beam-compiled programs on Apache Flink
- `spark/` — Execute Beam-compiled programs on Apache Spark

## When you work here

- Beam code must remain fabric-agnostic; defer fabric-specific optimizations to flink/ and spark/
- Changes to the Beam IR should be coordinated with all fabric implementations
- Document how your changes affect CPL synthesis and execution guarantees
