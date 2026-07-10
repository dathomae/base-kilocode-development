# Story049: Plan Storage Services (Service Pattern)

## Goal

Implement the `PlanStorageService` for managing plan storage: the bridge between relational metadata (`ts_doc`) and blob/object storage. This service coordinates document versioning with physical data persistence.

The service ensures that relational metadata updates are correctly synchronized with object storage writes, handling failure scenarios and maintaining bitemporal consistency.

## References

- [Data Stores Requirements](../specifications/requirements/data-stores.md) - REL-001 through REL-011
- [schema.sql](../../src/backend/serde/rdbms/schema.sql) - Database schema
- [Story029_object-store-requirements.md](Story029_object-store-requirements.md) - Object store requirements
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Mutation Orchestrator

## Dependencies

- [Story041_relational-schema-implementation.md](finished-stories/Story041_relational-schema-implementation.md) - Completed
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Not Started
- [Story044_localization-services.md](Story044_localization-services.md) - Not Started
- [Story048_plan-type-services.md](Story048_plan-type-services.md) - Not Started
- [Story029_object-store-requirements.md](Story029_object-store-requirements.md) - Not Started (Blocked by storage driver definition)
- [go-jet](https://github.com/go-jet/jet) - Type-safe query construction

## Tasks

1. **Structural Design & Refinement** - Not Started
   a. Create a **Class Diagram** (Mermaid) defining the `PlanStorageService`, its relationship to the `ObjectStore` interface, and the `ts_doc` metadata bridge.
   b. Create a **Sequence Diagram** (Mermaid) visualizing the **Synchronous Coordination** between the RDBMS metadata write and the Object Store blob write.
   c. Define the failure handling strategy (split-brain) and how the service ensures atomicity (or eventual consistency) across two disparate storage systems.

2. **Define Service & Base Operations** - Not Started
   a. Create `src/backend/domain/plan_storage.go`.
   b. Implement CRUD for timeseries documents within a namespace.
   c. **Validation**: Ensure `ts_doc_type` references are active and valid.
   d. Use the Story 043 orchestrator for all metadata mutations.

3. **Implement Storage Coordination** - Not Started
   a. Implement the coordination logic: update relational metadata AND write to object storage.
   b. Handle "split-brain" scenarios: what happens if the relational write succeeds but the blob write fails?
   c. Implement cleanup/rollback logic for failed storage operations.

4. **Migrate Store Logic to Service** - Not Started
   a. Identify all plan-storage-related methods in `src/backend/store/rdbms.go`.
   b. Refactor the API layer to call `PlanStorageService` directly.
   c. Purge the plan storage "dumb" logic from the old Store implementation.

## Constraints

- All writes must use **Advisory Locks** via the Story 043 engine.
- Relational metadata and object storage must remain in sync.
- The service must be the **sole entry point** for plan storage metadata.
- No ORM usage; all queries must use **Go-Jet**.
- **Domain Mapping**: Use custom domain structs with `sql` and `json` tags for scanning query results. Do not use Go-Jet generated models in the service or API layers.

## Acceptance Criteria

1. `PlanStorageService` handles full CRUD for `ts_doc` with correct bitemporal behavior.
2. Relational and object storage writes are correctly coordinated.
3. Failure handling for split-brain scenarios is verified.
4. All tests in `domain/plan_storage_test.go` pass.
