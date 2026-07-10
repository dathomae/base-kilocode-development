# Domain Services

Core business logic and planning domain models. This is the "brain" of the system.

## Responsibilities

- Business rule enforcement and validation
- Namespace scoping and multi-tenancy
- Temporal integrity and bitemporal coordination
- Orchestration of persistence operations

## Package Boundaries

| Package                  | Status         | Purpose                                                                                                                                   |
| ------------------------ | -------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `../serde/rdbms`         | Required       | Mutation orchestration via ExecuteBitemporalMutation and query helpers                                                                    |
| `../security`            | Required       | User authentication and authorization checks                                                                                              |
| `../serde/json_schema`   | Available      | Schema validation for foreign namespace assertions                                                                                        |
| `../handlers`            | **Prohibited** | Domain services provide values to handlers, not the reverse                                                                               |
| `../serde/rdbms/gen/...` | **Restricted** | Whatever is imported may only be used to call RegisterDomainModel or inside go-jet statements. Never use generated models as field types. |

## Handler-Facing Operations

Entry points available for external callers live in the parent `domain` package.
Handlers must import `chronocone/backend/domain`, never child entity packages.
Public operations are aggregate-root oriented; owned child manipulation helpers
remain unexported unless the child is itself a true aggregate root.

- **NamespaceFamilies** (`namespace_family.go`, `namespace_family_sql.go`)
  - `CreateOrUpdateFamily`
  - `GetFamilyByKey`
  - `ListFamilies`
- **Namespaces** (`namespace.go`, `namespace_sql.go`)
  - `CreateOrUpdateNamespace`
  - `GetNamespaceByKey`
  - `ListNamespaces`
- **Categories** (`category.go`, `category_sql.go`)
  - `CreateOrUpdateCategory`
  - `GetCategoryByKey`
  - `ListCategories`
- **Elements** (`element.go`, `element_sql.go`)
  - `CreateOrUpdateElement`
  - `GetElementByKey`
  - `ListElements`

## Architecture

Each aggregate root is organized as:

- `aggregate.go` — Unified aggregate-root model and public surface for that root
- `aggregate_sql.go` — SQL query construction, EBM registration, and database operations
- `aggregate_attr.go` — owned child-table models/helpers when currently needed
- Additional files (e.g., `aggregate_temporal.go`) for logical separation as needed

Services use **Unified Models** — one hand-authored struct per persisted entity that
serves as domain type, persistence record, and API contract simultaneously.
Use `json` tags for the API contract. The `alias:"<table_name>"` tag is the
documented escape hatch for cases where Go-Jet's qrm needs to be redirected
(see "Naming Discipline" below); most fields and entities need no tag at
all. **Never use the `sql` struct tag** — Go-Jet ignores it for both read
(qrm) and write (`INSERT … MODELS`) paths, so its presence is misleading.

See `memory-bank/specifications/patterns/service-organization.md` for full details.

## Domain Models

Domain models are defined entirely within the domain package. They must not use
the generated models from `serde/rdbms/gen/postgres/public/model` as field types or in business logic.
Generated models and the `table.*` package may only be imported for two restricted purposes:

1. Calling `RegisterDomainModel(domainType, generatedType)` at package initialization.
2. Using column metadata for SQL construction (e.g., `table.NamespaceFamily.AllColumns`, `table.NamespaceFamily.Key.EQ(...)`).

Domain models embed persistence mixins from `domain/versioning.go`:

- `Bounded` — ValidityStart and ValidityEnd for temporal validity
- `Keyed` — Natural key for immutability verification
- `Audit` — Temporal audit metadata

Domain models implement `serde/rdbms.LockableEntity` directly (no wrappers).
`LockableEntity.NaturalKey()` returns the entity's natural key; for entities
whose natural-key column is literally `key`, the embedded `Keyed.Key` field
collides with a method named `Key()` — hence the `NaturalKey()` accessor
returning `e.Key`. No `Unwrap` indirection: the domain struct _is_ the
persistence record.

### Naming Discipline (Go-Jet qrm requirement)

Go-Jet's `qrm` resolves SELECT projections by matching the destination
struct's **Go type name** to the column-alias prefix (`<table>.`) and the
**Go field name** to the column name, after normalizing both (lowercase,
strip `_`/`-`/space). For a unified model to hydrate correctly:

1. The domain type's Go name must normalize to its table name
   (e.g., `NamespaceFamily` ↔ `namespace_family`).
2. Each direct field's Go name must normalize to its column name
   (e.g., `NsFamID` ↔ `ns_fam_id`, `CustomerSubjectID` ↔ `customer_subject_id`).
3. Embedded persistence mixins (`Keyed`, `Bounded`, `Audit`) carry an
   `alias:"<table_name>"` tag so qrm rebinds the column-prefix lookup to
   the embedding entity's table instead of the mixin's own type name.

When a Go field name genuinely cannot match its column (keyword
collision, ambiguity from multiple FK columns to the same parent table),
add `alias:"<table>.<column>"` directly on the field — this overrides
both halves of the lookup. This should be rare.

## Database Awareness

Domain services are not database-agnostic. They use Go-Jet for type-safe queries and
reflective hydration of Unified Models. All state mutations must delegate to EBM
(`serde/rdbms/ExecuteBitemporalMutation`) to maintain bitemporal consistency.

## Foreign Namespace Assertions

External namespaces may assert statements against elements in this namespace. To do so,
they must first assert on the "schema" category: "my statements look like [schema]".

Domain services enforce this by:

1. Requiring a schema assertion before accepting any other assertions from a foreign namespace
2. Using `serde/json_schema` utilities to validate incoming statements against the asserted schema

This ensures external namespaces make only valid assertions. See `schema.sql` and
`memory-bank/specifications/architecture/database-design.md` for context.

## Temporal Parameters in Domain Service Signatures

Domain services accept both `effectiveTime` and `systemTime` parameters even when
callers pass `nil` for one or both. This enables forward compatibility with future
UI features (e.g., "what would the budget be on date X using data known at time Y?").

- `effectiveTime` (ValidTime): When a fact is true in the world. Nil means "now" in the real world.
- `systemTime` (SystemTime): When the fact was recorded in the database. Nil means CURRENT_TIMESTAMP.

Services pass these parameters directly to `ActiveAt()` and `ActiveAtParams()` rdbms helpers
for temporal filtering. Do not compute over ambient time; accept it as a parameter. Do not
skip parameters even if they seem unnecessary today.

## When you work here

- Accept temporal parameters in public service functions; pass them to rdbms query helpers
- Keep services focused on a single logical point in time (the effective timestamp provided)
- Delegate timeline splitting and mutation orchestration to EBM
- Write clear, testable business logic
- Document non-obvious constraints or temporal assumptions

## References

- **Critical**: `memory-bank/specifications/patterns/service-organization.md` — Unified model pattern
- **Critical**: `memory-bank/specifications/patterns/layered-architecture.md` — Domain responsibilities
- `memory-bank/specifications/architecture/database_design.md` — Bitemporal semantics (reference as needed)

## Lessons Learned

**Test all logical edge cases against embedded PostgreSQL; never mock the database**

Domain services use go-jet queries, so they have database dependencies. Do not mock those dependencies. Instead:

1. Test against an embedded PostgreSQL instance (see `serde/rdbms/setup_test.go` for the fixture)
2. Seed realistic test data that exercises edge cases:
   - Boundary conditions (empty result sets, single rows, large batches)
   - Temporal edge cases (validity overlaps, transitions at exact boundaries)
   - Multi-namespace scenarios (foreign assertions, scope violations)
   - Error paths (missing parent entities, constraint violations, staleness conflicts)
3. Assert both return values AND database state after each operation

Why: Mocking database queries lets logical bugs hide. A service function can "pass" a mock test while actually constructing invalid SQL or missing critical WHERE clauses. Real database tests catch these immediately. Bitemporal logic in particular has subtle temporal correctness requirements that only a real database can fully validate.

See `memory-bank/specifications/patterns/backend-testing-design.md` for the complete testing strategy across all backend layers and why domain tests focus on edge cases rather than user examples.

**Never manufacture or default missing input data**

When reading external data, the domain layer must not fill in missing values with zeros or any other default, nor silently accept duplicates or extraneous rows. If input deviates from the specification in any way — wrong row count, missing combinations, extra rows — the operation must fail with a descriptive error. Silently manufacturing data destroys trust in deterministic outputs.

**Encapsulate private domain information**

Internal registries, lookup tables, or configuration constants that live in the domain layer must not be directly accessed by other layers (e.g., handlers or serialization). Expose only what callers need through domain-layer functions that return the result in a controlled form. This preserves encapsulation and lets the internal representation change without affecting callers.

**`ORDER BY` is mandatory whenever a query uses `LIMIT`**

A `LIMIT N` without `ORDER BY` returns a planner-dependent subset of rows, not the intended rows. Even when the result is logically single-valued, always write the `ORDER BY` to document intent, prevent planner regressions, and ensure deterministic results under concurrent inserts or schema changes.
