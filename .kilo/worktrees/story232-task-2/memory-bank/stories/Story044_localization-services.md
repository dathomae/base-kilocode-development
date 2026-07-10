# Story044: Localization Services (Service Pattern)

## Goal

Implement the `LocalizationService` for managing bitemporal human-facing strings across all domain entities. This service provides a clean API for `display_name`, `tooltip`, and `description` management with built-in locale fallback logic.

By moving to a Service Pattern, we encapsulate the complex join logic and temporal filtering required for localization, providing a simple interface to the API layer.

## References

- [Database Design: Bitemporal Model](../specifications/architecture/database_design.md) - Section 5: Localization Pattern
- [Data Stores Requirements](../specifications/requirements/data-stores.md) - REL-003, REL-007
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Mutation Orchestrator

## Dependencies

- [Story041_relational-schema-implementation.md](finished-stories/Story041_relational-schema-implementation.md) - Completed
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Not Started (Required for mutation orchestration)
- [go-jet](https://github.com/go-jet/jet) - Type-safe query construction

## Dependent Stories

- [Story045_policy-services.md](Story045_policy-services.md) - Not Started
- [Story046_taxonomy-services.md](Story046_taxonomy-services.md) - Not Started
- [Story047_taxonomy-extension-services.md](Story047_taxonomy-extension-services.md) - Not Started
- [Story048_plan-type-services.md](Story048_plan-type-services.md) - Not Started

## Tasks

1. **Structural Design & Refinement** - Not Started
   a. Create a **Class Diagram** (Mermaid) in this story or a dedicated design doc defining the `LocalizationService` and its relationship to the `languages` reference data.
   b. Create a **Sequence Diagram** (Mermaid) visualizing the **Locale Fallback Chain** (`lang+country` → `lang` → `default`).
   c. Define the internal structure of the service (e.g., does it require a cache for active languages?).
   d. Define the generic interface/struct for localized data transfers.

2. **Define Service & Reference Data** - Not Started
   a. Create `src/backend/domain/localization.go`.
   b. Implement language/locale reference methods (e.g., `ListLanguages`, `AddLanguage`).

3. **Implement Bitemporal Localization Mutations** - Not Started
   a. Implement `func (s *LocalizationService) UpsertLocalization[T any](ctx context.Context, entityID int64, locale string, data T) error`.
   b. Use the **Mutation Orchestrator** from Story 043 to handle the Close-and-Insert logic for localization rows.
   c. Ensure that updating localization does **not** close or version the base entity (REL-007).

4. **Implement Query & Fallback Logic** - Not Started
   a. Implement `func (s *LocalizationService) GetLocalized[T any](ctx context.Context, entityID int64, locale string) (T, error)`.
   b. Implement the fallback chain: `lang+country` → `lang` → `default (en)`.
   c. Use Go-Jet to build a reusable query fragment for localized entity retrieval.

5. **Migrate Store Logic to Service** - Not Started
   a. Identify all localization-related methods in `src/backend/store/rdbms.go`.
   b. Refactor the API layer to call `LocalizationService` directly.
   c. Purge the localization "dumb" logic from the old Store implementation.

## Constraints

- All writes must use **Advisory Locks** via the Story 043 engine.
- Localization versioning must remain independent of base entity versioning.
- The service must be the **sole entry point** for localization data; no direct DB access from other services or handlers.
- No ORM usage; all queries must use **Go-Jet**.
- **Domain Mapping**: Use custom domain structs with `sql` and `json` tags for scanning query results. Do not use Go-Jet generated models in the service or API layers.

## Acceptance Criteria

1. `LocalizationService` successfully handles CRUD for all entity types.
2. Locale fallback chain works correctly in unit tests.
3. Updating localized strings does not trigger versioning on the base entity.
4. API handlers can consume the service without knowing about bitemporal mechanics.
5. All tests in `domain/localization_test.go` pass.
