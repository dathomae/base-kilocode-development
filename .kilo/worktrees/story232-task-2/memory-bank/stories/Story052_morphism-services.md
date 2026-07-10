# Story052: Morphism Services (Service Pattern)

## Goal

Implement the `MorphismService` for managing morphisms: **transformer**, **shifter**, **filter**, and **lookup**. These define the computation DAG (how metrics are derived from other metrics within a namespace).

The service ensures that all morphism definitions maintain bitemporal consistency and that the overall computation DAG remains acyclic.

## References

- [Data Stores Requirements](../specifications/requirements/data-stores.md) - REL-001 through REL-009
- [schema.sql](../../src/backend/serde/rdbms/schema.sql) - Database schema
- [ChronoCone Concepts](../specifications/chronocone-concepts.md) - Transformers, shifters, and the computation DAG
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Mutation Orchestrator

## Dependencies

- [Story041_relational-schema-implementation.md](finished-stories/Story041_relational-schema-implementation.md) - Completed (new tables for filter/lookup required)
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Not Started
- [Story044_localization-services.md](Story044_localization-services.md) - Not Started
- [Story048_plan-type-services.md](Story048_plan-type-services.md) - Not Started
- [Story032_metrics-and-calculation-requirements.md](Story032_metrics-and-calculation-requirements.md) - Not Started (Defines morphism types)
- [go-jet](https://github.com/go-jet/jet) - Type-safe query construction

## Tasks

1. **Structural Design & Refinement** - Not Started
   a. Create a **Class Diagram** (Mermaid) defining the `MorphismService` and the different morphism models (Transformer, Shifter, Filter, Lookup) within the computation DAG.
   b. Create a **Sequence Diagram** (Mermaid) visualizing the **Cycle Detection** logic across all morphism types within a namespace.
   c. Define the internal structure of the service and how it implements recursive Go-Jet joins for DAG traversal.

2. **Define Service & Base Operations** - Not Started
   a. Create `src/backend/domain/morphism.go`.
   b. Implement CRUD for each morphism type within a namespace (scoped to `ns_id`).
   c. **Validation**: Ensure input and output metric references point to active metrics.
   d. Use the Story 043 orchestrator for all mutations.
   e. Manage input ordering and type-specific attributes.

3. **Implement DAG & Cycle Detection** - Not Started
   a. Implement cycle detection logic to prevent circular references within a namespace.
   b. Implement upstream/downstream traversal logic for a given metric across all morphism types.
   c. Use Go-Jet for complex recursive joins across multiple morphism tables.

4. **Migrate Store Logic to Service** - Not Started
   a. Identify all morphism-related methods in `src/backend/store/rdbms.go`.
   b. Refactor the API layer to call `MorphismService` directly.
   c. Purge the morphism "dumb" logic from the old Store implementation.

## Constraints

- All writes must use **Advisory Locks** via the Story 043 engine.
- DAGs must be strictly contained within a single namespace.
- Key immutability and suspension/reactivation rules (REL-007) must be enforced.
- No ORM usage; all queries must use **Go-Jet**.
- **Domain Mapping**: Use custom domain structs with `sql` and `json` tags for scanning query results. Do not use Go-Jet generated models in the service or API layers.

## Acceptance Criteria

1. Morphism types are correctly implemented as per exploration results.
2. `MorphismService` handles full CRUD for all types with correct bitemporal behavior.
3. DAG relationships are maintained with cycle detection and impact analysis.
4. All entities are properly namespace-scoped.
5. All tests in `domain/morphism_test.go` pass.
