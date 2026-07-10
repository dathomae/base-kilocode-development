# Story047: Taxonomy Extension Services (Service Pattern)

## Goal

Implement the `TaxonomyExtensionService` for managing taxonomy extension tables: `category_shape` (schema declarations) and `element_attr` (individual statements). This service implements a **schema-and-statement** pattern for cross-namespace assertions.

The service provides validation logic to ensure that `element_attr` records conform to the JSON schemas and regex patterns defined in their parent `category_shape`.

## References

- [Data Stores Requirements](../specifications/requirements/data-stores.md) - REL-001 through REL-009
- [schema.sql](../../src/backend/serde/rdbms/schema.sql) - Database schema
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Mutation Orchestrator

## Dependencies

- [Story041_relational-schema-implementation.md](finished-stories/Story041_relational-schema-implementation.md) - Completed (schema correction required)
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Not Started
- [Story044_localization-services.md](Story044_localization-services.md) - Not Started
- [Story046_taxonomy-services.md](Story046_taxonomy-services.md) - Not Started
- [go-jet](https://github.com/go-jet/jet) - Type-safe query construction

## Tasks

1. **Structural Design & Refinement** - Not Started
   a. Create a **Class Diagram** (Mermaid) defining the `TaxonomyExtensionService`, the `category_shape` -> `element_attr` relationship, and the cross-namespace assertion model.
   b. Create a **Sequence Diagram** (Mermaid) visualizing the **JSON Schema & Regex Validation** flow during an attribute upsert.
   c. Define the internal structure of the service and how it manages external schema validation libraries.

2. **Define Service & Base Operations** - Not Started
   a. Create `src/backend/domain/taxonomy_extension.go`.
   b. Implement CRUD for `category_shape` records.
   c. Support both "Placeholder" (naming only) and "Schema Declaration" (JSON + Regex) variants.
   d. Use the Story 043 orchestrator for all mutations.
   e. Ensure cross-namespace ownership logic (a namespace declaring a shape on another's category).

3. **Implement Element Attribute Operations (Statements)** - Not Started
   a. Implement CRUD for `element_attr` records.
   b. **Validation**: Enforce that `element_attr.value` conforms to the parent `category_shape` JSON schema and regex patterns.
   c. Use the Story 043 orchestrator for all mutations.
   d. Handle cascade behavior (tombstoning a category or shape affects its attributes).

4. **Implement Extension Query Patterns** - Not Started
   a. Implement queries to fetch all shapes for a category across all namespaces.
   b. Implement queries to fetch all attributes for an element, grouped by shape.
   c. Use Go-Jet for complex joins between categories, elements, shapes, and attributes.

5. **Migrate Store Logic to Service** - Not Started
   a. Identify all extension-related methods in `src/backend/store/rdbms.go`.
   b. Refactor the API layer to call `TaxonomyExtensionService` directly.
   c. Purge the extension "dumb" logic from the old Store implementation.

## Constraints

- All writes must use **Advisory Locks** via the Story 043 engine.
- JSON schema validation must be performed at the application level (within the service).
- The service must be the **sole entry point** for taxonomy extension data.
- No ORM usage; all queries must use **Go-Jet**.
- **Domain Mapping**: Use custom domain structs with `sql` and `json` tags for scanning query results. Do not use Go-Jet generated models in the service or API layers.
- Key immutability for shapes must be enforced.

## Acceptance Criteria

1. Corrected schema with `element_attr` referencing parent `category_shape`.
2. Full CRUD for `category_shape` and `element_attr` with correct bitemporal behavior.
3. JSON schema and regex validation correctly enforced by the service.
4. Cross-namespace ownership model works correctly.
5. All tests in `domain/taxonomy_extension_test.go` pass.
