# CPL Execution Model

## Overview

**CPL (Chronocone Planning Language)** is a domain-specific language for expressing deterministic 
planning computations. See `memory-bank/specifications/design/cpl/` for the language specification.

The CPL compiler synthesizes planning programs to executable binaries. Currently, it targets Go 
binaries; the compiler is designed for multiple targets.

## Execution Paths

### Path 1: Beam (Distributed Execution)

CPL programs compile to Beam IR, which executes on Flink or Spark for horizontally scalable 
distributed computation. Not yet in use; implementation pending.

### Path 2: Kubernetes (Batch Execution)

CPL programs compile to Go executables. Each execution runs as a Kubernetes Job that:
1. Loads input data from the database
2. Executes the planning computation
3. Writes results back to the database

The executable owns result collection. The Kubernetes batch runner focuses on:
- Job orchestration and submission
- Observability (metrics, logs, tracing)
- Retry and failure handling

**TBD Decision**: Batch runner implementation strategy. Options to evaluate:
- Kubernetes Jobs API directly
- Argo Workflows
- Apache Airflow
- Other orchestration frameworks

## Current Constraints

### RAM-Resident Only

All CPL executions must fit their working data in pod memory. This is a hard constraint for now.

**Problem Categories**:

- **Amazon-scale planning** — Demand forecasting, capacity planning for large retailers. Fits 
  comfortably in RAM. Current focus: single-node CPU parallelism.

- **Scheduling & Inventory Placement** — Large-scale scheduling and inventory problems. Exceed 
  RAM constraints; deferred to future scaling work.

- **Metric Optimization** — Optimizing a metric within a plan. Fits in RAM but requires 
  significant CPU. Future: GPU acceleration or horizontal scaling.

- **Whole Plan Optimization** — Searching across alternative plans. Treats each plan execution 
  as one step in the search. Requires K8s fleet for horizontal search parallelism.

## Future Directions

### GPU Acceleration (TBD)

GPU support is under consideration for metric optimization and other CPU-intensive workloads.
Strategy TBD: synthesizer support, Kubernetes node selection, or both.

### Scaling Beyond RAM (TBD)

For scheduling and inventory problems exceeding RAM, exploration of:
- Partitioning strategies
- Intermediate result spilling
- Multi-stage execution

### Distributed Planning (TBD)

Beam path for horizontally scalable distributed execution remains planned but not yet implemented.

## References

- `memory-bank/specifications/design/cpl/` — CPL language specification
- `src/chronocone_planning_language/` — CPL compiler implementation
- `src/backend/compute/beam/` — Beam orchestration (planned)
- `src/backend/compute/kubernetes/` — Kubernetes batch execution (implementation in progress)
