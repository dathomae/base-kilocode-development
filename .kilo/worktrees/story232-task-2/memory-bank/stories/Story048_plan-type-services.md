# Story048: Plan Type Services (Service Pattern)

## Goal

Implement the `PlanTypeService` for managing plan type definitions: `metric`, `metric_dim`, `ts_doc_type`, and `ts_doc_row_type`. These define the dimensional model for metrics and the structure of timeseries documents within a namespace.

The service ensures that all dimensional references are valid and active, maintaining the integrity of the computation DAG.

## References

- [Data Stores Requirements](../specifications/requirements/data-stores.md) - REL-001 through REL-009
- [schema.sql](../../src/backend/serde/rdbms/schema.sql) - Database schema
- [ChronoCone Concepts](../specifications/chronocone-concepts.md) - Metrics and the computation DAG
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Mutation Orchestrator

## Dependencies

- [Story041_relational-schema-implementation.md](finished-stories/Story041_relational-schema-implementation.md) - Completed
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Not Started
- [Story044_localization-services.md](Story044_localization-services.md) - Not Started
- [Story046_taxonomy-services.md](Story046_taxonomy-services.md) - Not Started
- [go-jet](https://github.com/go-jet/jet) - Type-safe query construction

## Dependent Stories

- [Story049_plan-storage-services.md](Story049_plan-storage-services.md) - Not Started
- [Story052_morphism-services.md](Story052_morphism-services.md) - Not Started 

## Tasks

1. **Structural Design & Refinement** - Not Started
   a. Create a **Class Diagram** (Mermaid) defining the `PlanTypeService`, the `metric` -> `metric_dim` -> `category` relationship, and the `ts_doc` metadata model.
   b. Create a **Sequence Diagram** (Mermaid) visualizing a **Metric Creation** operation, showing the validation of dimensions against the Taxonomy Service.
   c. Define the internal structure of the service and its interaction with the Taxonomy Service for dimensional validation.

2. **Define Service & Base Operations** - Not Started
   a. Create `src/backend/domain/plan_type.go`.
   b. Implement CRUD for metrics within a namespace (scoped to `ns_id`).
   c. **Validation**: Ensure that metric dimensions (`metric_dim`) reference active taxonomy categories.
   d. Use the Story 043 orchestrator for all mutations.

3. **Implement Timeseries Document Type Operations** - Not Started
   a. Implement CRUD for `ts_doc_type` and `ts_doc_row_type` within a namespace.
   b. Ensure correct parent-child scoping between doc types and row types.
   c. Use the Story 043 orchestrator for all mutations.

4. **Implement Impact Analysis Query Patterns** - Not Started
   a. Implement queries to determine which metrics are affected by a taxonomy change.
   b. Implement queries to list metrics with their complete dimensional structure.
   c. Use Go-Jet for complex joins between metrics, dimensions, and categories.

5. **Migrate Store Logic to Service** - Not Started
   a. Identify all plan-type-related methods in `src/backend/store/rdbms.go`.
   b. Refactor the API layer to call `PlanTypeService` directly.
   c. Purge the plan type "dumb" logic from the old Store implementation.

## Constraints

- All writes must use **Advisory Locks** via the Story 043 engine.
- Metrics and plan types must be correctly namespace-scoped.
- The service must be the **sole entry point** for plan type data.
- No ORM usage; all queries must use **Go-Jet**.
- **Domain Mapping**: Use custom domain structs with `sql` and `json` tags for scanning query results. Do not use Go-Jet generated models in the service or API layers.

## Acceptance Criteria

1. `PlanTypeService` handles full CRUD for metrics, dimensions, and document types with correct bitemporal behavior.
2. Metric-to-dimension references are validated against active taxonomy categories.
3. Impact analysis correctly identifies affected metrics during taxonomy updates.
4. All entities are properly namespace-scoped.
5. All tests in `domain/plan_type_test.go` pass.
