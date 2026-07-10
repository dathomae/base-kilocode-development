# Story 043J: Implement and Correctly Architect the Backend API

## 1. Overview
After Story 043C purged the repository layer, all API handlers were left returning `501 Not Implemented`. This story delivers the working CRUD API for namespace families, namespaces, categories, and elements against a real PostgreSQL database — and does so within the project's strict layered architecture from the start.

Story 043D was written to accomplish this goal but prescribed an approach that violated the architecture: it instructed the implementor to place Go-Jet queries, EBM mutation calls, and generated model types directly inside the `handlers` package. A partial implementation followed that approach. This story supersedes 043D entirely — the functional goal (a working API) and the architectural goal (correct layering) are both requirements here, not alternatives.

## 2. Requirements

### 2.1 Deliver the API Surface
All routes stubbed out in 043C must be implemented and functional against a live PostgreSQL instance. Business logic is intentionally minimal — basic input validity checking only. No cascade analysis, no taxonomy tree validation, no attribute schema enforcement.

**Namespace Families:** list all, get by key, create/update via EBM (PUT), merge non-nil fields via EBM (PATCH).
**Namespaces:** list within a family, get by family+key, PUT, PATCH. Parent family must exist (404 if not).
**Categories:** list within a namespace, get, PUT, PATCH, list attribute shapes, create/update shape via EBM. Parent namespace must exist.
**Elements:** list within a category, get, PUT, PATCH, list attributes, create/update attribute via EBM. Parent category must exist.

All entity keys must be 6–64 characters; return 422 on violation. Localization routes remain 501 stubs.

### 2.2 Re-establish the Domain Layer
The `domain/entities/` package must own all business logic and state orchestration.
- Create or update the necessary domain service functions (e.g., `CreateOrUpdateNamespace`, `GetNamespaceByKey`, `ListNamespaces`) to handle the operations currently embedded in the handlers.
- Move all Go-Jet `SELECT` queries from the handlers into these domain functions.
- Move all `rdbms.ExecuteBitemporalMutation` calls from the handlers into these domain functions.
- Handlers must call exactly one domain function per operation and handle HTTP-specific translation (parsing request bodies, formatting responses).

### 2.2 Purge Database Awareness from Handlers
Handlers must act strictly as a translation layer.
- Remove all imports of `chronocone/backend/serde/rdbms` and `chronocone/backend/serde/rdbms/gen/...` from `src/backend/handlers/...`.
- Handlers must no longer construct `InsertStatement`s or interact with `table.*` or `model.*`.
- Handlers must receive and return Unified Domain Models from the `domain` layer, not `model.*` types.

### 2.3 Eliminate Prohibited Wrapper Patterns
`memory-bank/lessons-learned.md` explicitly outlaws `fooWrapper` structs embedding generated models.
- Remove `namespaceWrapper`, `familyWrapper`, and any other similar wrappers from the handlers.
- Ensure the hand-authored Unified Domain Models (e.g., `NamespaceFamily`, `Namespace`) in `domain/entities/` implement `serde/rdbms.LockableEntity` directly.
- The interface methods must be `EntityID() int64` and `NaturalKey() string` (no `Unwrap() any`).
- Move all `rdbms.RegisterEntity(reflect.TypeOf(DomainModel{}), ...)` calls to the `init()` functions inside `domain/entities/`.

### 2.4 Enforce Temporal Determinism
`memory-bank/lessons-learned.md` prohibits ambient time in the domain layer.
- Remove instances of `time.Now()` or `time.Now().UTC()` in `domain` logic used to drive bitemporal mutations. 
- The `handlers` layer must resolve `"now"` to a concrete `time.Time` and pass it to the domain layer via the `effectiveTime` / `systemTime` parameters. The domain logic must strictly use the provided parameters.

### 2.5 Remove Prohibited Struct Tags
Go-Jet ignores the `sql` struct tag, making it misleading.
- Remove all `sql:"..."` struct tags from domain models (e.g., in `domain/entities/families.go`).
- Rely on the strict Go-Jet naming normalization or use `alias:"<table>.<column>"` as the designated escape hatch if absolutely necessary.

## 3. Acceptance Criteria
* [ ] No file in `src/backend/handlers/` imports `serde/rdbms` or any of its subpackages.
* [ ] No file in `src/backend/handlers/` imports `serde/rdbms/gen/...` (no `table.*` or `model.*`).
* [ ] The prohibited wrapper pattern (`type XWrapper struct { *model.X }`) has been entirely removed from the codebase.
* [ ] All `rdbms.RegisterEntity` calls occur within `domain/entities/` `init()` functions using Unified Domain Models.
* [ ] No `sql:"..."` struct tags exist in the `domain` layer.
* [ ] No `time.Now()` calls exist in `domain` layer business logic (handlers parse ambient time and pass concrete values).
* [ ] All handlers successfully map HTTP requests to exactly one domain service function call.
* [ ] `go test ./...` passes across the backend.
* [ ] The full API surface from §2.1 is functional against a running PostgreSQL instance with no 501 stubs remaining (except localization routes).

## 4. Tasks

Asking for user assistance, assess and define how best to move forward with respect to the architectural violations outlined in the requirements.