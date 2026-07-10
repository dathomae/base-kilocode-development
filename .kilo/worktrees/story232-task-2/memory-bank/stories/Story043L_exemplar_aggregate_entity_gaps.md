# Story 043L: Fill Exemplar Aggregate Entity Gaps

## Goal

Implement the exemplar aggregate-root pattern for namespace family, namespace, category, and element. Public service operations should deal in complete aggregate objects. Exported Unified Model structs for owned rows are allowed when they are part of the aggregate's data shape, but owned parts such as localization rows, category shapes, category attrs, and element attrs must be created, replaced, queried, or filtered only through their owning aggregate.

This story builds on [Story043K_domain_package_aggregate_roots.md](Story043K_domain_package_aggregate_roots.md).

## References

- [Backend brief](../../src/backend/brief.md)
- [Domain brief](../../src/backend/domain/brief.md)
- [Service organization pattern](../specifications/patterns/service-organization.md)
- [Database design](../specifications/architecture/database_design.md)
- [Story043J_remedy_architectural_violations.md](Story043J_remedy_architectural_violations.md)
- [Story044_localization-services.md](Story044_localization-services.md)
- [Story046_taxonomy-services.md](Story046_taxonomy-services.md)
- [Story047_taxonomy-extension-services.md](Story047_taxonomy-extension-services.md)

## Dependencies

- [Story043K_domain_package_aggregate_roots.md](finished-stories/Story043K_domain_package_aggregate_roots.md) - Done

## Dependent Stories

- [Story043N_extend_aggregate_pattern.md](Story043N_extend_aggregate_pattern.md) - Not Started

## Tasks

1. Define exemplar aggregate read shapes - Not Started
   a. Write compile-time and JSON contract tests for `NamespaceFamily`, `Namespace`, `Category`, and `Element` aggregate shapes, including owned child data where the exemplar API exposes it. - Not Started
   b. Decide whether exemplar reads return nested child collections, lift the current localization onto root fields, or support both through explicit read options. - Not Started
   c. Add aggregate fields for owned localizations and exemplar owned parts as Unified Model data, without treating exported child structs as standalone service APIs. - Not Started
   d. Present the proposed exemplar aggregate shapes, localization lifting behavior, and collection semantics to the user for review before implementing persistence behavior. - Not Started
   e. Document authoritative collection semantics: whether missing children mean unchanged, replace, close validity, or not loaded. - Not Started

2. Add owned localization support for exemplars - Not Started
   a. Write embedded PostgreSQL tests for creating, reading, and replacing localizations through each aggregate root. - Not Started
   b. Add domain models and EBM registration for `namespace_family_localization`, `namespace_localization`, `category_localization`, and `element_localization`. - Not Started
   c. Add unexported helpers that hydrate and mutate localization rows only through aggregate operations. - Not Started

3. Add category-owned shape and attr support - Not Started
   a. Write embedded PostgreSQL tests showing `GetCategory` and `GetCategories` return categories with the authorized point-in-time owned shape/attr data required by the exemplar API. - Not Started
   b. Write embedded PostgreSQL tests showing category mutations create, replace, or close `category_shape`, `category_shape_localization`, `category_attr`, and `category_attr_localization` only through category aggregate operations. - Not Started
   c. Register domain models for the category owned-part tables and correct any JSONB field shapes needed for Go-Jet hydration. - Not Started
   d. Implement aggregate-level read and create/update/replace behavior for category shapes and attrs. - Not Started

4. Add element-owned attr support - Not Started
   a. Write embedded PostgreSQL tests showing `GetElement` and element list reads return elements with authorized point-in-time owned attr data required by the exemplar API. - Not Started
   b. Write embedded PostgreSQL tests showing element mutations create, replace, or close `element_attr` and `element_attr_localization` only through element aggregate operations. - Not Started
   c. Register domain models for element attr tables and correct JSONB field shapes needed for Go-Jet hydration. - Not Started
   d. Implement aggregate-level read and create/update/replace behavior for element attrs. - Not Started

5. Remove stale scaffold and align handlers - Not Started
   a. Write tests that fail while stale no-op functions such as `FindCategories` remain as accepted paths. - Not Started
   b. Remove or replace stale skeleton code and route handlers through the aggregate-root APIs. - Not Started
   c. Update story and brief references so "CRUD" means aggregate-level operations, not table-level public operations. - Not Started

## Test-First Development

Follow the Logical TDD Lifecycle. Each behavior task must start with embedded PostgreSQL tests that verify both returned aggregate values and database state. Avoid mock database tests for Go-Jet or EBM behavior.

## Constraints

- Services expose complete aggregate objects; owned parts do not have independent public service CRUD.
- Exported child Unified Model structs are allowed as fields or collection elements inside aggregate read/write shapes; the boundary is service access, not type visibility.
- Each edge or owned collection must have explicit ownership and replacement semantics.
- Use EBM for all bitemporal mutations.
- Do not use `time.Now()` in domain logic.
- Do not use `sql` struct tags in domain models.
- JSONB fields should use typed wrappers or established local JSONB patterns rather than ad hoc maps unless the story explicitly records a reason.

## Intent

The goal is to create a high-quality exemplar that future agents can copy: complete aggregate reads and writes, private child-table mechanics, clear temporal behavior, and tests that prove persistence state instead of only surface responses.

## Acceptance Criteria

- Namespace family, namespace, category, and element aggregate roots can be read and mutated through public domain operations.
- `GetCategory` / `GetCategories` and `GetElement` / element list reads return the exemplar aggregate-owned child data according to the approved read shape.
- Their owned localization rows are available only through aggregate-root operations, with current-localization lifting behavior documented where supported.
- Category shape/category attr and element attr gaps are filled for the exemplar taxonomy aggregates.
- All newly mutable bitemporal domain models are registered for EBM.
- No independent public CRUD exists for owned child tables.
- Embedded PostgreSQL tests verify returned aggregates and underlying database state.
- `go test ./...` passes for the backend.

## Open Questions

- Should aggregate mutation APIs be named `CreateOrUpdate*`, `Replace*`, or split into explicit `Put*` and `Patch*` semantics?
- Should localization replacement be full aggregate replacement by default, or a named sub-operation on the aggregate root?
- Should exemplar reads default to nested child collections, lifted current-localization fields, or an explicit option that chooses between them?
