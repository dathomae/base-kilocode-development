# Story 043K: Move Entity Services to Domain Aggregate Roots

## Goal

Migrate the current `domain/entities` public service surface into the parent `domain` package so Go visibility can enforce aggregate-root boundaries. Public callers should interact with complete aggregate roots such as `Category` and `Element`; owned parts such as attrs, localizations, and edge rows should be manipulated only through their owning aggregate.

This story follows [Story043J_remedy_architectural_violations.md](Story043J_remedy_architectural_violations.md) and refines its service placement after the aggregate-root design discussion.

## References

- [Backend brief](../../../src/backend/brief.md)
- [Domain brief](../../../src/backend/domain/brief.md)
- [Service organization pattern](../../specifications/patterns/service-organization.md)
- [Layered architecture pattern](../../specifications/patterns/layered-architecture.md)
- [Database design](../../specifications/architecture/database_design.md)

## Dependencies

- [Story043J_remedy_architectural_violations.md](Story043J_remedy_architectural_violations.md) - Done

## Dependent Stories

- [Story043L_exemplar_aggregate_entity_gaps.md](../Story043L_exemplar_aggregate_entity_gaps.md) - Not Started
- [Story043N_extend_aggregate_pattern.md](../Story043N_extend_aggregate_pattern.md) - Not Started

## Tasks

1. Define the package migration shape - Done
   a. Write or update tests/import checks that fail while handlers or other external packages import `chronocone/backend/domain/entities` directly. - Done
   b. Decide the final file layout under `src/backend/domain/` for aggregate roots, owned parts, SQL helpers, validation helpers, and mutation orchestration. - Done
   c. Present the proposed package migration shape and boundary decisions to the user for review before moving code. - Done
   d. Document the package boundary in `src/backend/domain/brief.md`: public aggregate-root operations live in `domain`; child manipulation helpers are unexported. - Done

2. Move exemplar public types and operations into `domain` - Done
   a. Write tests that import only `chronocone/backend/domain` and exercise family, namespace, category, and element public operations. - Done
   b. Move the current `domain/entities` public models and operations into the parent package without changing their behavior. - Done
   c. Preserve EBM registration through package initialization and verify the registered domain types still hydrate through Go-Jet qrm. - Done

3. Clarify owned-child aggregate boundary for Story 043L - Done
   a. Document that exported Unified Model structs for owned child rows (for example `CategoryAttr`, `CategoryShape`, and `ElementAttr`) are allowed when they are part of an aggregate's read/write data shape. - Done
   b. Document that standalone public child-table CRUD/service helpers are not allowed; owned children must be created, replaced, queried, or filtered through aggregate-root operations such as `GetCategory` / `GetCategories` / `GetElement`. - Done
   c. Remove the pre-043L standalone child read helpers and make dependent handlers return `501 Not Implemented` until Story 043L replaces them with complete aggregate reads. - Done

4. Update callers and documentation - Done
   a. Write handler tests or compile checks that fail until handlers import the parent `domain` package. - Done
   b. Update handlers and any other callers to use `domain` aggregate-root APIs. - Done
   c. Update briefs, references, and story notes that still describe `domain/entities` as the handler-facing service package. - Done

## Test-First Development

Follow the Logical TDD Lifecycle. Each coding task must add failing compile checks or integration tests before moving code, then implement only enough migration to make those tests pass. Do not mark tasks complete until the corresponding tests pass in the current workspace.

## Constraints

- Do not create DTOs to mask package movement. Unified Models remain the handler/domain/API shape unless an API contract genuinely differs.
- Do not expose public CRUD for owned child tables such as attrs or localizations.
- Do not construct raw SQL outside `serde/rdbms`.
- Do not use `sql` struct tags in domain models.
- Preserve temporal parameters in public domain service signatures.
- Use embedded PostgreSQL tests for behavior that depends on Go-Jet hydration or EBM.

## Intent

The goal is structural clarity: Go package visibility should reinforce the aggregate-root design instead of relying on social convention. Future agents should see complete aggregate objects as the service boundary. Owned child structs may be exported as Unified Models when returned inside those aggregates, but child-table CRUD helpers remain private mechanics.

## Acceptance Criteria

- No production handler imports `chronocone/backend/domain/entities`.
- Public handler-facing operations for namespace families, namespaces, categories, and elements are available from `chronocone/backend/domain`.
- Standalone owned-child manipulation helpers are not public APIs; exported child Unified Model structs may appear inside aggregate-root data shapes.
- Domain briefs and service organization references describe aggregate roots in the parent `domain` package.
- `go test ./...` passes for the backend.

## Open Questions

- Should a temporary `domain/entities` compatibility package remain for one story, or should this migration be atomic?
- Story 043L should decide the first exemplar aggregate response shape: nested children under the root, current localization lifted onto the root, or both via explicit read options.
