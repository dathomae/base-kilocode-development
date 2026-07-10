# Story041: Relational Schema Implementation

## Goal

Implement schema changes to the PostgreSQL relational store as identified during the Story027 requirements exploration. This story begins with consolidating the existing five v0 schema files into a single ordered schema file and will accumulate additional schema tasks as requirements are defined.

## References

- [Story027: Relational Store Requirements](Story027_relational-store-requirements.md) - Parent requirements story
- [Data Stores Requirements](../specifications/requirements/data-stores.md) - Requirements being implemented
- [Database Design: Bitemporal Model](../specifications/architecture/database_design.md) - Bitemporal pattern documentation
- [schema.sql](../../src/backend/serde/rdbms/schema.sql) - Consolidated master schema file

## Dependencies

- [Story027_relational-store-requirements.md](Story027_relational-store-requirements.md) - In Progress (requirements must be defined before implementation proceeds beyond consolidation)
- [go-jet](https://github.com/go-jet/jet) - Used for type-safe query construction in place of the canceled SQL preprocessor (Story 042)

## Dependent Stories

None (to be determined as requirements exploration continues)

## Tasks

1. Consolidate the five v0 schema files into a single `schema.sql` file with dependency-correct ordering and clear section markers - Completed
   a. Phase A: Structural Definition - Determine correct table creation order based on foreign key and application constraint dependencies - Completed
   b. Phase B: Logical Tests - Identify required test changes for the consolidated schema file - Completed
   c. Phase C: Implementation - Completed
      i. Create single `schema.sql` with section comment blocks for navigability - Completed
      ii. Remove duplicate `CREATE EXTENSION IF NOT EXISTS citext;` statements (keep one at the top) - Completed
      iii. Remove the five individual files - Completed
      iv. Update any references to the individual files (Story027, Story014, database CLI, etc.) - Completed
   d. Phase D: Documentation - Create walkthrough and post-mortem - Not Started

2. Add `transaction_at` column to all bitemporal tables (REL-003) - Completed
   a. Phase A: Structural Definition - Define schema changes and method signatures in `repository.py` - Not Started
   b. Phase B: Logical Tests - Update tests to verify `transaction_at` is populated correctly - Not Started
   c. Phase C: Implementation - Not Started
      i. Add `transaction_at timestamptz NOT NULL DEFAULT transaction_timestamp()` to every bitemporal table - Not Started
      ii. Remove `created_at` from `insertion_metadata` JSONB population in `repository.py` - Not Started
      iii. Update `_build_insert_query` and `_build_insert_query_with_id` to stop writing temporal data into `insertion_metadata` - Not Started
      iv. Update any queries that extract `created_at` from JSONB to use the `transaction_at` column instead - Not Started
   d. Phase D: Documentation - Create walkthrough and post-mortem - Not Started

3. Fix namespace-scoped logical keys for `metric`, `transformer`, and `shifter` (REL-008) - Completed
   a. Phase A: Structural Definition - Define application code method signatures for `ns_id` inclusion - Completed
   b. Phase B: Logical Tests - Update tests for namespace-scoped lookups - Completed
   c. Phase C: Implementation - Completed
      i. Add `ns_id` column to `metric`, `transformer`, and `shifter` tables - Completed
      ii. Change logical key constraints from `UNIQUE (name, validity_end)` to `UNIQUE (ns_id, name, validity_end)` - Completed
      iii. Update physical key covering indices to include `INCLUDE (validity_start)` pattern - Completed
      iv. Update application code (repository, ontology service) to pass `ns_id` when creating/querying these entities - Completed
   d. Phase D: Documentation - Create walkthrough and post-mortem - Not Started

4. Implement strict, schema-driven attribute model (Task 4) - Completed
   a. Phase A: Structural Definition & RDBMS Implementation - Completed
      i. Replaced "stuffing" plan with dedicated `category_shape` table for element-level attribute definitions - Completed
      ii. Added `asserting_ns_id` to `category_shape` to support decentralized, cross-namespace assertions - Completed
      iii. Implemented covering indexes `INCLUDE (validity_start)` for all bitemporal `PRIMARY KEY` and `UNIQUE` constraints - Completed
      iv. Refactored `element_attr` to enforce strict schema adherence via `category_shape_id` reference - Completed
      v. Promoted all `physical_pk` constraints to `PRIMARY KEY` to support `go-jet` - Completed
   b. Phase B: Go Store Layer - Completed
      i. Refactored `store.go` to an interface-based design supporting both `MemoryStore` and `RdbmsStore` - Completed
      ii. Implemented `RdbmsStore` using `go-jet` for type-safe CRUD of `category_shape` and `element_attr` - Completed
      iii. Integrated `RdbmsStore` into `main.go` via `DATABASE_URL` environment variable - Completed
   c. Automated Tooling and Type-Safe Model Generation - Completed
      i. Created `cmd/dbtool` to manage embedded PostgreSQL lifecycle for schema application and generation - Completed
      ii. Implemented `scripts/generate-models.sh` to automate `go-jet` model generation - Completed
      iii. Updated `src/backend/Makefile` to make `serde/rdbms/gen` depend on `schema.sql` - Completed
      iv. Integrated `generate` target into `test` and `all` workflows for automatic synchronization - Completed
   d. Phase C: API Layer - Completed
      i. Add CRUD handlers for `category_shape` - Completed
      ii. Update element attribute handlers to enforce schema validation (existence check) - Completed
   e. Phase D: Documentation - Completed (Updated database_design.md, data-stores.md, and concepts-definitions.md)


*(Additional tasks will be added as Story027 requirements exploration continues)*

## Constraints

- The consolidated schema must produce an identical database when executed against a fresh PostgreSQL instance
- All existing tests and the database initialization CLI must continue to work after consolidation
- The single file should use clear section comment blocks for navigability

## Intent

This story acts as the implementation companion to Story027's requirements exploration. As requirements are confirmed and recorded, corresponding schema changes are added as tasks here. Starting with consolidation gives us a clean, single-file baseline before making structural changes.

## Acceptance Criteria

1. A single `schema.sql` file exists and can be executed with `psql -f schema.sql` to create the full database schema
2. The five individual `.sql` files are removed
3. All existing tests pass without modification (beyond file path changes)
4. The database initialization CLI works with the new file structure

## Notes

- The table creation order must respect: `language` → `storage_config` → `namespace_family_type` → `namespace_type` → `subject_gateway` → `namespace_family_namespace_type` → `namespace_family` → `namespace` → `category` → `element` → `metric` → `transformer` → `shifter` → `ts_doc_type` → `ts_doc` (approximate, to be verified)
- Additional schema tasks from Story027 will be appended to the Tasks section as they are identified
