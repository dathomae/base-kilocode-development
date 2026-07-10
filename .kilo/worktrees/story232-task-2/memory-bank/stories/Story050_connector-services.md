# Story050: Connector Services (Service Pattern)

## Goal

Implement the `ConnectorService` for managing the connector domain: storage configuration. These operational entities define how and where blob/object data is stored across the system.

The service enforces global uniqueness for storage configuration keys and ensures bitemporal consistency using the Story 043 engine.

## References

- [Data Stores Requirements](../specifications/requirements/data-stores.md) - REL-001 through REL-009
- [schema.sql](../../src/backend/serde/rdbms/schema.sql) - Database schema
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Mutation Orchestrator

## Dependencies

- [Story041_relational-schema-implementation.md](finished-stories/Story041_relational-schema-implementation.md) - Completed
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Not Started
- [go-jet](https://github.com/go-jet/jet) - Type-safe query construction

## Tasks

1. **Structural Design & Refinement** - Not Started
   a. Create a **Class Diagram** (Mermaid) defining the `ConnectorService` and the `storage_config` model, including driver-specific JSON attribute management.
   b. Create a **Sequence Diagram** (Mermaid) visualizing the **Global Uniqueness Validation** for storage configuration keys across all namespaces.
   c. Define the internal structure of the service and how it manages driver-specific configuration schema validation.

2. **Define Service & Base Operations** - Not Started
   a. Create `src/backend/domain/connector.go`.
   b. Implement CRUD for storage configurations.
   c. **Uniqueness**: Enforce that storage configuration keys are globally unique.
   d. Use the Story 043 orchestrator for all mutations.
   e. Manage driver assignment and driver-specific attributes (JSON).

3. **Migrate Store Logic to Service** - Not Started
   a. Identify all connector-related methods in `src/backend/store/rdbms.go`.
   b. Refactor the API layer to call `ConnectorService` directly.
   c. Purge the connector "dumb" logic from the old Store implementation.

## Constraints

- All writes must use **Advisory Locks** via the Story 043 engine.
- Key immutability and suspension/reactivation rules (REL-007) must be enforced.
- The service must be the **sole entry point** for connector data.
- Storage configuration keys must be globally unique across all namespaces.
- No ORM usage; all queries must use **Go-Jet**.
- **Domain Mapping**: Use custom domain structs with `sql` and `json` tags for scanning query results. Do not use Go-Jet generated models in the service or API layers.

## Acceptance Criteria

1. `ConnectorService` handles full CRUD for `storage_config` with correct bitemporal behavior.
2. Storage config globally unique key enforcement is verified.
3. All operations use the mutation orchestrator with advisory locks.
4. All tests in `domain/connector_test.go` pass.

