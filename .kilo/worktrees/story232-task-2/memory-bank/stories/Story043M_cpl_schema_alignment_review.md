# Story 043M: Review CPL Drift and Align the Database Schema

## Goal

Review `src/chronocone_planning_language` against the current bitemporal database schema and domain aggregate direction, then adjust `src/backend/serde/rdbms/schema.sql` and related generated metadata so persisted entities match the concepts CPL now expresses.

This story exists because CPL has continued evolving since the earlier database stories and may no longer map cleanly onto the schema's taxonomy, metric, connector, and morphism tables.

## References

- [Backend brief](../../src/backend/brief.md)
- [RDBMS brief](../../src/backend/serde/rdbms/brief.md)
- [Database design](../specifications/architecture/database_design.md)
- [Service organization pattern](../specifications/patterns/service-organization.md)
- [Story043L_exemplar_aggregate_entity_gaps.md](Story043L_exemplar_aggregate_entity_gaps.md)
- [Story046_taxonomy-services.md](Story046_taxonomy-services.md)
- [Story048_plan-type-services.md](Story048_plan-type-services.md)
- [Story052_morphism-services.md](Story052_morphism-services.md)

## Dependencies

- [Story043K_domain_package_aggregate_roots.md](finished-stories/Story043K_domain_package_aggregate_roots.md) - Done

## Dependent Stories

- [Story043N_extend_aggregate_pattern.md](Story043N_extend_aggregate_pattern.md) - Not Started

## Tasks

1. Establish CPL review context - Not Started
   a. Read `src/chronocone_planning_language/README.md` before inspecting or changing CPL-related files. - Not Started
   b. Write an inventory test or review artifact that lists CPL declarations and generated concepts that require persistence. - Not Started
   c. Compare CPL concepts against current `schema.sql` sections for taxonomy, morphisms, timeseries, storage, and policy. - Not Started
   d. Present the CPL-to-schema scope assessment to the user for review before classifying drift or proposing schema changes. - Not Started

2. Identify schema drift - Not Started
   a. Write failing schema or generation tests that capture each confirmed mismatch between CPL and the database schema. - Not Started
   b. Classify each mismatch as missing table, obsolete table, missing column, wrong ownership, wrong edge direction, or naming mismatch. - Not Started
   c. Present the drift classification and proposed schema-change scope to the user for review before editing `schema.sql`. - Not Started
   d. Record any ambiguity as open questions rather than guessing at architectural intent. - Not Started

3. Adjust schema and generated models - Not Started
   a. Update `schema.sql` for confirmed schema changes while preserving bitemporal partitioning, localization, key, and insertion metadata rules. - Not Started
   b. Regenerate Go-Jet models using the project script and inspect generated diffs for content changes rather than mode-only churn. - Not Started
   c. Update rdbms registry constants and table registrations to match any new or renamed bitemporal tables. - Not Started

4. Reconcile domain aggregate plans - Not Started
   a. Update exemplar aggregate stories or docs if CPL review changes which tables are owned parts versus aggregate roots. - Not Started
   b. Add tests proving the adjusted schema supports the exemplar aggregate reads and writes needed by Story 043L. - Not Started
   c. Update the copy-forward guidance for Story 043N so future entities copy the corrected schema shape. - Not Started

## Test-First Development

Follow the Logical TDD Lifecycle. Schema changes must be preceded by failing schema, generation, or embedded PostgreSQL tests that demonstrate the drift. Do not mark documentation or schema status complete until generated code and backend tests verify the change.

## Constraints

- `src/chronocone_planning_language` has its own ownership context; read its README before work and avoid opportunistic refactors.
- Raw SQL belongs in `serde/rdbms`; domain and handlers must continue using Go-Jet/rdbms helpers.
- Bitemporal tables must preserve validity, transaction, mode, insertion metadata, active/historical partitioning, and EBM compatibility.
- Localization tables must continue following the language/country/display text pattern unless a deliberate schema decision changes it.
- Ask for guidance when CPL semantics imply an architectural schema change rather than a local correction.

## Intent

The goal is to prevent agents from copying a polished aggregate pattern over a stale schema. CPL should be treated as a source of semantic pressure on the database model before the pattern is generalized.

## Acceptance Criteria

- A documented CPL-to-schema gap inventory exists in the story implementation notes or an appropriate memory-bank artifact.
- Confirmed schema mismatches are corrected in `schema.sql`.
- Go-Jet generated models and rdbms registry metadata match the adjusted schema.
- Backend tests verify the adjusted schema against embedded PostgreSQL.
- Any unresolved CPL/schema questions are documented explicitly and not hidden behind partial implementation.

## Open Questions

- Which CPL concepts are authoritative enough to drive schema changes now, and which remain exploratory?
- Should any current morphism tables be renamed to align with CPL vocabulary before the pattern is copied broadly?
