# Story046: Taxonomy Services (Service Pattern)

## Goal

Implement the `TaxonomyService` for managing the taxonomy domain: categories, elements, and hierarchical taxonomy structures. This service handles complex cascade logic for tombstoning and batch validation of structural integrity.

The service encapsulates recursive taxonomy traversal and ensures that all structural changes conform to bitemporal constraints using the Story 043 engine.

## References

- [Data Stores Requirements](../specifications/requirements/data-stores.md) - REL-001 through REL-009
- [schema.sql](../../src/backend/serde/rdbms/schema.sql) - Database schema
- [ChronoCone Concepts](../specifications/chronocone-concepts.md) - Taxonomy structure
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Mutation Orchestrator

## Dependencies

- [Story041_relational-schema-implementation.md](finished-stories/Story041_relational-schema-implementation.md) - Completed
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Not Started
- [Story044_localization-services.md](Story044_localization-services.md) - Not Started
- [go-jet](https://github.com/go-jet/jet) - Type-safe query construction

## Tasks

1. **Structural Design & Refinement** - Not Started
   a. Create a **Class Diagram** (Mermaid) defining the `TaxonomyService`, its entity model (Category/Element), and the recursive relationship structure.
   b. Create a **Sequence Diagram** (Mermaid) visualizing a **Batch Tombstone Cascade** operation, showing how it identifies and locks downstream participants.
   c. Define the internal structure of the service and any helper types for tree representation.

2. **Define Service & Base Operations** - Not Started
   a. Create `src/backend/domain/taxonomy.go`.
   b. Implement CRUD for categories within a namespace (scoped to `ns_id`).
   c. Implement CRUD for elements within a category (scoped to `category_id`).
   d. Use the Story 043 orchestrator for all mutations.
   e. Enforce Key Immutability and suspension rules.

3. **Implement Hierarchy & Tree Traversal** - Not Started
   a. Implement taxonomy parent assignment and hierarchy validation (no circular references).
   b. Implement recursive query logic (Go-Jet + CTE or equivalent) for taxonomy tree traversal.
   c. Provide high-level "Get Tree" methods that return domain-oriented tree structures.

4. **Implement Batch Taxonomy Operations (REL-005)** - Not Started
   a. Implement in-situ validation for batch updates.
   b. Implement cascade impact analysis (e.g., tombstone element → track invalidated metric dimensions).
   c. Use the "Protocol" (orchestrated set of mutations) pattern to ensure atomic batch consistency with advisory locks.

5. **Migrate Store Logic to Service** - Not Started
   a. Identify all taxonomy-related methods in `src/backend/store/rdbms.go`.
   b. Refactor the API layer to call `TaxonomyService` directly.
   c. Purge the taxonomy "dumb" logic from the old Store implementation.

## Constraints

- All writes must use **Advisory Locks** via the Story 043 engine.
- Batch updates must use deterministic ordering for lock acquisition to prevent deadlocks.
- In-memory validation is preferred over row-by-row DB logic for structural integrity.
- The service must be the **sole entry point** for taxonomy data.
- No ORM usage; all queries must use **Go-Jet**.
- **Domain Mapping**: Use custom domain structs with `sql` and `json` tags for scanning query results. Do not use Go-Jet generated models in the service or API layers.

## Acceptance Criteria

1. `TaxonomyService` handles full CRUD for categories and elements with correct bitemporal behavior.
2. Taxonomy hierarchy operations work correctly (add/move/tombstone with structural validation).
3. Batch operations are atomic and correctly acquisition locks.
4. Cascade impact analysis correctly identifies all affected downstream entities.
5. All tests in `domain/taxonomy_test.go` pass.

