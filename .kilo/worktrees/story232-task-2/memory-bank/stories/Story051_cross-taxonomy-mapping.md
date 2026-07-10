# Story051: Cross-Taxonomy Mapping (Service Pattern)

## Goal

Implement the `MappingService` for cross-taxonomy mapping: the ability for a namespace owner to import and reorganize elements from outside taxonomies into their own structure.

The service manages the `category_mapping` and `element_mapping` tables, ensuring that source entities are valid across namespace family boundaries and maintaining bitemporal integrity.

**Deferred past demo.**

## References

- [Data Stores Requirements](../specifications/requirements/data-stores.md)
- [ChronoCone Concepts](../specifications/chronocone-concepts.md) - Taxonomy and namespace concepts
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Mutation Orchestrator

## Dependencies

- [Story041_relational-schema-implementation.md](finished-stories/Story041_relational-schema-implementation.md) - Completed (new schema tables required)
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Not Started
- [Story044_localization-services.md](Story044_localization-services.md) - Not Started
- [Story046_taxonomy-services.md](Story046_taxonomy-services.md) - Not Started
- [go-jet](https://github.com/go-jet/jet) - Type-safe query construction

## Tasks

1. **Structural Design & Refinement** - Not Started
   a. Create a **Class Diagram** (Mermaid) defining the `MappingService`, the `category_mapping` -> `element_mapping` relationship, and the cross-namespace family validation model.
   b. Create a **Sequence Diagram** (Mermaid) visualizing a **Source Entity Validation** during mapping creation, showing the cross-family read check.
   c. Define the internal structure of the service and its interaction with the Policy Service for cross-family access validation.

2. **Define Service & Base Operations** - Not Started
   a. Create `src/backend/domain/mapping.go`.
   b. Implement CRUD for category and element mappings.
   c. **Validation**: Ensure source categories/elements exist and are active in their respective namespaces.
   d. Implement cross-namespace-family read access validation.
   e. Use the Story 043 orchestrator for all mutations.

3. **Implement Mapping Query Patterns** - Not Started
   a. Implement queries to find all source mappings for a local category.
   b. Implement impact analysis: find all consuming mappings for a given source entity.
   c. Use Go-Jet for complex cross-namespace joins.

4. **Migrate Store Logic to Service** - Not Started
   a. Identify any mapping-related methods in `src/backend/store/rdbms.go` (if any exist).
   b. Refactor the API layer to call `MappingService` directly.
   c. Purge any legacy mapping logic from the old Store implementation.

## Constraints

- All writes must use **Advisory Locks** via the Story 043 engine.
- Mappings cross namespace family boundaries; read access must be validated.
- Mapping relationships must remain within their bitemporal validity ranges.
- The service must be the **sole entry point** for mapping data.
- No ORM usage; all queries must use **Go-Jet**.
- **Domain Mapping**: Use custom domain structs with `sql` and `json` tags for scanning query results. Do not use Go-Jet generated models in the service or API layers.

## Acceptance Criteria

1. New schema tables created with full bitemporal support.
2. `MappingService` handles full CRUD for mappings with correct temporal semantics.
3. Cross-namespace-family source validation works correctly.
4. Impact analysis correctly identifies all affected mappings during source updates.
5. All tests in `domain/mapping_test.go` pass.
