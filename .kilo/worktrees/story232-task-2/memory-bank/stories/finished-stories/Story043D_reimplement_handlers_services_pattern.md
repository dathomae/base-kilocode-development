# Story 043D: Reimplement API Handlers with Services Pattern

> [!WARNING]
> **ARCHITECTURAL VIOLATION**
> This story explicitly instructed the implementor to bypass the `domain` layer and place database queries, mutations, and generated models directly into the HTTP `handlers` package. This violates the core architectural constraints defined in `brief.md` files and `memory-bank/lessons-learned.md`.
> **This story is considered flawed and has been superseded by [Story 043J](Story043J_remedy_architectural_violations.md).**

## 1. Overview
Story 043C stubbed all API handlers with `501 Not Implemented` after purging the repository-pattern `store/` layer. This story replaces those stubs with working implementations that follow the Services Pattern: each handler queries PostgreSQL directly via go-jet and calls `ExecuteBitemporalMutation` (EBM) for writes. Business logic is intentionally minimal — basic input validity checking only. The goal is a working, low-value demo backed by the real database.

**Prerequisite:** Story 043C must be completed first. This story assumes `store/` no longer exists, the bitemporal infrastructure lives in `serde/rdbms/`, and the handlers are 501 stubs.

## 2. Requirements

### 2.1 Handler Architecture
Each handler file follows the Services Pattern:

1. **Receives** a `*sql.DB` (not a `Store` interface) via the handler struct or closure.
2. **Reads** use go-jet `SELECT` statements constructed inline, using `serde/rdbms/gen/postgres/public/table` for table references and the temporal expression helpers from `serde/rdbms/temporal_expressions.go` (e.g., `IsCurrent`) for filtering.
3. **Writes** construct a `[]postgres.InsertStatement` and pass it to `ExecuteBitemporalMutation` via the builder callback. The handler is responsible for:
   - Fetching the current state into a `[]LockableEntity` wrapper.
   - Providing the builder function that returns only `InsertStatement`s.
   - Passing `maxTransactionAt` (from the caller's known state) for staleness detection.
4. **Validation** is limited to:
   - Key format (6–64 characters, per schema constraint).
   - Timestamp parsing (RFC3339 or `"now"`).
   - Parent existence (e.g., namespace family must exist before creating a namespace).
5. **No complex business logic** — no cascade analysis, no taxonomy tree validation, no attribute schema enforcement. Those are future service-layer concerns.

### 2.2 Response Models
Replace all `store.*` types in handler output structs with go-jet generated model types from `serde/rdbms/gen/postgres/public/model`. The go-jet models already carry the correct column names and types. Handler output structs should embed or reference these directly.

If the go-jet model's JSON serialization is inadequate (e.g., internal IDs leaking), create minimal response structs in the `handlers/` package that project only the API-safe fields. Do not create a separate DTO package.

### 2.3 API Surface
Reimplement the following routes (preserving the existing route structure from the 043C stubs):

**Namespace Families:**
| Route | Method | Operation |
|---|---|---|
| `/families` | GET | List all current namespace families |
| `/family/{family_key}` | GET | Get a namespace family by key |
| `/family/{family_key}` | PUT | Create or update a namespace family via EBM |
| `/family/{family_key}` | PATCH | Fetch current, merge non-nil fields, update via EBM |

**Namespaces:**
| Route | Method | Operation |
|---|---|---|
| `/namespaces/{family_key}` | GET | List current namespaces within a family |
| `/namespace/{family_key}/{namespace_key}` | GET | Get a namespace by family + key |
| `/namespace/{family_key}/{namespace_key}` | PUT | Create or update a namespace via EBM |
| `/namespace/{family_key}/{namespace_key}` | PATCH | Fetch current, merge non-nil fields, update via EBM |

**Categories:**
| Route | Method | Operation |
|---|---|---|
| `/categories/{family}/{namespace}` | GET | List current categories within a namespace |
| `/category/{family}/{namespace}/{category_key}` | GET | Get a category |
| `/category/{family}/{namespace}/{category_key}` | PUT | Create or update a category via EBM |
| `/category/{family}/{namespace}/{category_key}` | PATCH | Fetch current, merge non-nil fields, update via EBM |
| `/category-shapes/{family}/{namespace}/{category_key}` | GET | List attribute shapes for a category |
| `/category-shape/{family}/{namespace}/{category_key}/{asserting_family}/{asserting_namespace}/{shape_key}` | PUT | Create or update a category shape via EBM |

**Elements:**
| Route | Method | Operation |
|---|---|---|
| `/elements/{family}/{namespace}/{category_key}` | GET | List current elements within a category |
| `/element/{family}/{namespace}/{category_key}/{element_key}` | GET | Get an element |
| `/element/{family}/{namespace}/{category_key}/{element_key}` | PUT | Create or update an element via EBM |
| `/element/{family}/{namespace}/{category_key}/{element_key}` | PATCH | Fetch current, merge non-nil fields, update via EBM |
| `/attributes/{family}/{namespace}/{category_key}/{element_key}` | GET | List attributes for an element |
| `/attribute/{family}/{namespace}/{category_key}/{element_key}/{asserting_family}/{asserting_namespace}/{shape_key}` | PUT | Create or update an element attribute via EBM |

**PATCH semantics:** All PATCH handlers follow the same pattern — fetch the current active record, overlay non-nil fields from the request body, pass the complete merged entity to EBM as an `InsertStatement`. EBM handles truncating the old record and inserting the new one. The handler does not compute splinters or manage intervals.

**Deferred from this story:**
- Localization routes — separate concern.
- Attribute JSON Schema validation — future service-layer work. Shapes are stored but not enforced against element attribute values.

### 2.4 Main Entry Point
Update `main.go` to:
1. Require `DATABASE_URL` environment variable (no more in-memory fallback).
2. Open a `*sql.DB` connection using `pgx`.
3. Pass `*sql.DB` to each handler registration function.

### 2.5 Parent Existence Checks
For write operations, the handler must verify parent entity existence before calling EBM:
- Creating a **namespace** requires the **namespace family** to exist (current).
- Creating a **category** requires the **namespace** to exist (current).
- Creating an **element** requires the **category** to exist (current).

Use go-jet queries with `IsCurrent` to check. Return `404` if the parent is missing.

### 2.6 Key Validation
All entity keys must satisfy the schema constraint: `length(key) BETWEEN 6 AND 64`. Validate in the handler before calling EBM. Return `422` on violation.

## 3. Acceptance Criteria
* [ ] `go build ./...` succeeds from `src/backend/`.
* [ ] `main.go` requires `DATABASE_URL` and initializes a `*sql.DB` — no in-memory fallback.
* [ ] No handler file imports `chronocone/backend/store`.
* [ ] All handler files import `serde/rdbms/gen/postgres/public/table` and/or `serde/rdbms/gen/postgres/public/model`.
* [ ] GET routes return current entities from PostgreSQL (using `IsCurrent` temporal filter).
* [ ] PUT routes create new entities via EBM, returning the created entity.
* [ ] PATCH routes fetch the current record, merge non-nil request fields, and update via EBM.
* [ ] PUT routes for child entities return `404` if the parent entity does not exist.
* [ ] PUT routes return `422` if the key fails the 6–64 character length check.
* [ ] Category shape and element attribute PUT routes function via EBM (no JSON Schema enforcement).
* [ ] The API is functional against a running PostgreSQL instance with the deployed schema.
* [ ] Localization routes remain as 501 stubs or are absent from route registration.
