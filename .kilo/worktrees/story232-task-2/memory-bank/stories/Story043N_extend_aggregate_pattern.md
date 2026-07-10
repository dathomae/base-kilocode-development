# Story 043N: Extend the Aggregate Pattern to Similar Entities

## Goal

Copy the matured aggregate-root pattern from the exemplar entities to all other similarly shaped schema entities. Public domain services should expose complete aggregate roots; owned localizations, attrs, and edge/intersection rows should be mutated only through one write owner while remaining readable from appropriate projections.

This story depends on the exemplar implementation and the CPL/schema alignment review so agents copy a verified pattern instead of early scaffolding.

## References

- [Backend brief](../../src/backend/brief.md)
- [Domain brief](../../src/backend/domain/brief.md)
- [RDBMS brief](../../src/backend/serde/rdbms/brief.md)
- [Service organization pattern](../specifications/patterns/service-organization.md)
- [Database design](../specifications/architecture/database_design.md)
- [Story043L_exemplar_aggregate_entity_gaps.md](Story043L_exemplar_aggregate_entity_gaps.md)
- [Story043M_cpl_schema_alignment_review.md](Story043M_cpl_schema_alignment_review.md)

## Legacy Context Barrier

Legacy stories 044-052 discuss related service areas but predate the aggregate-root pattern and may need to be rewritten. Do not treat them as authoritative for this story. Use them only as historical context after reading the current briefs, Story 043K, Story 043L, and Story 043M.

## Dependencies

- [Story043L_exemplar_aggregate_entity_gaps.md](Story043L_exemplar_aggregate_entity_gaps.md) - Not Started
- [Story043M_cpl_schema_alignment_review.md](Story043M_cpl_schema_alignment_review.md) - Not Started

## Dependent Stories

None - This story creates the broad scaffold future feature stories will consume.

## Tasks

1. Inventory similarly shaped entities - Not Started
   a. Write a schema-driven inventory test or review artifact listing base tables, localization tables, attr-like tables, and edge/intersection tables. - Not Started
   b. Classify each table as aggregate root, owned component, read-only reference, or edge/intersection. - Not Started
   c. Assign exactly one write owner for each mutable edge/intersection table and document read projections from the other side. - Not Started
   d. Present the entity classification, edge ownership decisions, and proposed implementation scope to the user for review before adding domain scaffolds. - Not Started

2. Implement aggregate roots for plan and morphism entities - Not Started
   a. Write embedded PostgreSQL tests for aggregate reads and writes for metrics, metric dimensions, transformers, transformer inputs/outputs, shifters, and shifter inputs/outputs. - Not Started
   b. Add domain aggregate models, EBM registration, hydration helpers, validation helpers, and mutation orchestration following the exemplar pattern. - Not Started
   c. Ensure edge rows are writable only through their owner aggregate while readable from related aggregate projections where useful. - Not Started

3. Implement aggregate roots for storage and timeseries entities - Not Started
   a. Write tests for storage configs, ts doc types, ts doc row types, ts docs, localizations, tags, and lineage according to their bitemporal or non-bitemporal lifecycle. - Not Started
   b. Add domain aggregate models and helpers following the exemplar pattern, noting where `ts_doc` is not itself bitemporal. - Not Started
   c. Define write ownership for tags and lineage and expose read projections without independent public CRUD. - Not Started

4. Implement or confirm policy/reference treatment - Not Started
   a. Write tests confirming read-only or admin-only behavior for language, namespace family type, namespace type, subject gateway, and namespace-family namespace-type capability rows. - Not Started
   b. Add read models or aggregate-owned helpers only where they support real domain operations. - Not Started
   c. Avoid inventing public mutation APIs for reference data unless a policy story explicitly requires them. - Not Started

5. Update handlers, docs, and guardrails - Not Started
   a. Add compile checks or tests that prevent handlers from importing implementation-only packages or child-table mutation helpers. - Not Started
   b. Update handlers to call aggregate-root APIs and update OpenAPI/MCP contracts as needed. - Not Started
   c. Update briefs and memory-bank guidance with the one-write-owner, multi-read-projection rule for edges. - Not Started

## Test-First Development

Follow the Logical TDD Lifecycle. Every new aggregate must begin with tests that verify returned aggregate shape, database state, EBM registration, and ownership boundaries. Do not generate broad boilerplate without tests proving the copied pattern still matches the entity's semantics.

## Constraints

- Public services expose complete aggregate roots, not table-level child CRUD.
- Each mutable edge/intersection table has one write owner and may have multiple read projections.
- Read-only/reference tables should remain read-only unless another story establishes a mutation use case.
- Preserve bitemporal invariants and EBM for bitemporal tables.
- Use real embedded PostgreSQL tests for persistence and temporal behavior.
- Do not use `sql` struct tags in domain models.

## Intent

The goal is broad scaffold consistency. Future advanced stories should find every major entity represented by an aggregate-root service with private owned-part mechanics, clear edge ownership, temporal parameters, and tests that prove the pattern works.

## Acceptance Criteria

- Every similarly shaped schema entity is classified and either implemented as an aggregate root, implemented as an owned component, or documented as reference/read-only.
- Metrics, transformers, shifters, storage configs, and timeseries metadata have aggregate-root domain scaffolds where appropriate.
- Edge/intersection tables have exactly one documented write owner and read projections from other useful sides.
- No independent public CRUD exists for owned components or edge rows.
- Backend tests cover the copied aggregate pattern across all implemented entity families.
- `go test ./...` passes for the backend.

## Open Questions

- Which non-bitemporal tables should be brought under aggregate-root services versus left as persistence details?
- Should generated OpenAPI/MCP contracts represent aggregate roots immediately, or should handler contract changes be staged after backend domain scaffolding?
