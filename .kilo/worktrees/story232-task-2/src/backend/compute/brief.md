# Compute Layer

ChronoCone supports multiple compute fabrics for distributed execution of deterministic DAGs.

## Responsibilities

_(see Architecture below — this brief describes the two compute paths supported by ChronoCone; path-specific responsibilities live in `beam/` and `kubernetes/`.)_

## Package Boundaries

_(none at this level — see child briefs (`beam/`, `kubernetes/`) for fabric-specific boundaries.)_

## Architecture

**Beam Path** (`beam/`) — For horizontal scalability via CPL synthesis. A single CPL program
compiles to Beam, which can then execute on Flink or Spark. Ideal when you want fabric-agnostic
scalability and the CPL synthesizer can express your computation.

**Kubernetes Path** (`kubernetes/`) — Direct deployment without Beam intermediation. Run multiple
CPL programs simultaneously on a Kubernetes fleet. Useful when Beam abstraction doesn't apply or
when customers prefer direct k8s control.

## When you work here

- Keep compute fabric implementations independent and testable in isolation
- Beam orchestration is fabric-agnostic; fabric-specific code lives in flink/ and spark/ subdirectories
- Customer-facing documentation should explain when to choose each path
