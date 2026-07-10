# Kubernetes Compute Fabric

Kubernetes-native execution of CPL-synthesized programs for RAM-resident and GPU-accelerated workloads.

## Responsibilities

_(see Architecture, Execution Model, and Use Cases below — this brief describes Kubernetes-native execution of CPL programs.)_

## Package Boundaries

_(none documented — depends on Kubernetes client libraries and the synthesized CPL executables.)_

## Architecture

CPL programs synthesize to native Go executable programs. The Kubernetes fabric hosts a batch runner
that orchestrates execution of these executables across a Kubernetes cluster.

## Execution Model

Each CPL program execution runs as a Kubernetes Job that:

1. Loads input data from the database
2. Executes the planning computation
3. Writes results back to the database

The executable owns result collection. The batch runner handles job orchestration, observability,
retries, and failure handling. See `memory-bank/specifications/architecture/cpl/execution-model.md`
for the full execution model and current constraints.

## Use Cases

- **RAM-Resident Problems**: CPL programs whose data fits in pod memory (e.g., Amazon-scale planning)
- **GPU Acceleration**: CPL programs targeting GPU-enabled Kubernetes nodes (planned)
- **Whole Plan Optimization**: Horizontal search parallelism via K8s fleet scaling

## Supported Deployment Methods

- `helm/` — Helm charts for parameterized customer deployments

## TBD Decisions

- Batch runner implementation strategy (see execution-model.md for options)
- GPU acceleration support and scheduling
- Scaling strategies for beyond-RAM problems

## When you work here

- Focus on operational clarity for FDEs deploying in customer environments
- Error messages and logs must help diagnose issues in customer Kubernetes clusters
- Document deployment scenarios, resource requirements, and GPU integration

## References

- **Critical**: `memory-bank/specifications/architecture/cpl/execution-model.md` — CPL execution semantics and constraints
