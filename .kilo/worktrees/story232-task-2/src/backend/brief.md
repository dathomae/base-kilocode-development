# Backend Layer

The Go-native backend implements the deterministic calculation engine for ChronoCone.

## Responsibilities

_(see Architecture below — this brief introduces the backend layer's constituent services; per-service responsibilities live in the child briefs.)_

## Package Boundaries

_(none at this level — see child briefs (`handlers`, `domain`, `serde`, `security`, `compute`, `cmd`) for inter-package boundaries.)_

## When you work here

_(see Key Patterns below and Lessons Learned at the end; per-package rules live in child briefs.)_

## Architecture

This layer uses a **Unified Model Pattern** with a **Domain Service Architecture**:

- **Domain Services** (`src/backend/domain/*.go`) — The core business logic layer. Services
  enforce business rules, namespace scoping, temporal integrity, and coordinate with the
  persistence layer. Domain services have database awareness through Go-Jet type-safe queries
  and reflective hydration.

- **Persistence & Bitemporal Orchestration** (`src/backend/serde/rdbms/`) — Handles the
  specialized database work: EBM (Effective Bitemporal Mutations), advisory locking, staleness
  detection, and the Flush/Purge/Assert pipeline. This layer contains complex, test-heavy
  logic that requires deep database expertise.

- **API Handlers** (`src/backend/handlers/`) — Translates HTTP requests into domain service calls.

## Key Patterns

**Unified Model**: A single struct serves as domain model, database-mappable
entity, and API contract. Use `json` tags for the API contract. **Do not use
the `sql` struct tag** — Go-Jet's qrm (read) and `INSERT … MODELS` (write)
both ignore it; presence of `sql` tags creates the false impression that
they drive Go-Jet behaviour. Field names and the struct's type name carry
the persistence mapping — see `domain/brief.md` for the naming discipline.
The `alias:"<table>[.<column>]"` tag is the only escape hatch and should
be used sparingly. No separate DTOs or repository classes.

**Service File Organization**: When a domain service grows, break it into files like:

- `domain/service.go` — Domain models and service interface/public methods
- `domain/service_sql.go` — SQL query construction and database operations for that service
- Additional files like `domain/service_temporal.go` as warranted for logical separation

**EBM Handling**: Domain services reason about data at a single point in time, defined by an
effective timestamp (often implied as "now"). When mutations need to apply across multiple points
in the validity timeline, services delegate that timeline splitting to **EBM** (Effective Bitemporal
Mutations) in `serde/rdbms/`. Services provide the new-assertion INSERT statement; EBM computes
and executes the Flush/Purge/Assert pipeline to maintain bitemporal honesty across the full timeline.

## References

- **Critical**: `memory-bank/specifications/patterns/layered-architecture.md` — Layer responsibilities and protocol
- `memory-bank/specifications/patterns/service-organization.md` — Unified model pattern details (see domain/ and serde/ briefs)

## Lessons Learned

**Check the diff before committing `serde/rdbms/gen/`**

Generated files should only be committed if content actually changed, not just file modes. Always review `git diff serde/rdbms/gen/` before staging — if the diff shows only mode changes (100644 ↔ 100755), revert to avoid noise.

**Run `make test` before pushing anything here**

All backend code changes must pass the test suite. Run `make test` locally before pushing to catch bugs early and avoid CI failures.

**Raw SQL strings are only permitted in `serde/rdbms`**

SQL string construction must happen exclusively in the `serde/rdbms` package. All other packages
(`handlers`, `domain`, `security`) must use rdbms helpers for queries and mutations:

- For queries: Use go-jet's type-safe query builders and `ActiveAt()` / `ActiveAtParams()` temporal helpers
- For mutations: Use `ExecuteBitemporalMutation()` with the domain's `BitemporalMutationBuilder` callback
- Never construct raw SQL strings in domain or handler code; never import database driver types into those layers

This constraint ensures: (1) temporal filtering is consistent across all query sites, (2) mutation semantics
are uniform (EBM pipeline enforcement), (3) the persistence contract stays explicit and auditable,
(4) dependency flow is unidirectional (handlers → domain → rdbms, never reverse).

**Organize test files to mirror source file structure: `setup_test.go`, `foo_test.go` for `foo.go`**

Test suites belong in modular files, not monolithic `_test.go` files. `setup_test.go` contains TestMain and shared test utilities (fixtures, helpers); individual `*_test.go` files contain tests for their corresponding source files. This keeps tests easy to locate and maintain as the package grows.

**Go functions should return value and error**

Functions that can fail should return `(T, error)` with `nil` error on success. This follows Go's "errors are values" philosophy, forces callers to acknowledge failure cases, and is consistent with the standard library (`os.Open`, `io.ReadAll`, etc.). Only deviate when there is an explicit directive to do so.

**Check assumptions or panic; check unlikely conditions or return error**

It is better to fail loudly than to succeed incorrectly when an invariant is violated. Assumptions about internal state that must always hold should be checked with a panic or a clear assertion. Conditions that are possible but unlikely (e.g., unexpected nil from a known-good source) should return an error with a descriptive message.
