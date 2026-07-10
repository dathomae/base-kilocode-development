# RDBMS Persistence

PostgreSQL persistence implementation with bitemporal support and consistency guarantees.

## Responsibilities

- Go-Jet integration for type-safe query construction
- Reflective data hydration into Unified Models
- EBM (Effective Bitemporal Mutations) and the Flush/Purge/Assert pipeline
- Advisory locking for serialization
- Staleness detection for optimistic concurrency control
- Partition management (active/historical)

## Package Boundaries

| Package                 | Status         | Purpose                                     |
| ----------------------- | -------------- | ------------------------------------------- |
| `github.com/jackc/pgx`  | Required       | PostgreSQL driver                           |
| `github.com/go-jet/jet` | Required       | Type-safe query construction                |
| `../../domain`          | **Prohibited** | Domain services call rdbms, not the reverse |
| `../../handlers`        | **Prohibited** | Handlers call domain, not rdbms directly    |

## Generated Models Usage

The go-jet-generated models in `serde/rdbms/gen/postgres/public/model/` are used **only for column metadata**.
This allows type-safe query construction without coupling domain logic to the database schema.

Examples of proper usage:

- ✓ `table.NamespaceFamily.AllColumns` — Column list for SELECT
- ✓ `table.NamespaceFamily.Key.EQ(...)` — Type-safe column expression
- ✗ Using generated `model.NamespaceFamily` as a domain model
- ✗ Passing generated models to domain services

Domain models (defined in the parent `domain` package) are the Unified Models used throughout the system. They implement `LockableEntity` and flow through handlers → domain services → serde/rdbms. The generated models never escape the persistence layer.

## Key Patterns

**Mutation Protocol: ExecuteBitemporalMutation** — All state changes flow through `ExecuteBitemporalMutation()`.
This is a general-purpose upsert orchestrator, not update-only. It supports both creating new entities (zero ID)
and mutating existing ones (populated ID). The protocol ensures:

1. Lock acquisition on logical entity ID (new entities get ID from query results; zero-ID entities signal "new")
2. For existing entities: staleness check (re-read under lock, compare transaction metadata)
3. EBM computation (prefix/assertion/suffix splinters via builder callback)
4. Flush (superseded records to historical partition)
5. Purge (remove from active partition)
6. Assert (insert new splinters to active partition)
7. Commit and lock release

Although it only supports mutating one table at a time, it handles large inserts by mutating the full scope at
once. If the default inserter is used, it batches into groups of 1000 rows at a time.

**Bitemporal Honesty** — Validity ranges and transactional metadata must satisfy invariants
documented in `memory-bank/specifications/architecture/database-design.md`.

**Synchronized Black-Boxes Pattern** — Temporal filtering logic is split across two paired helper functions
that must be called together to maintain correctness:

- `ActiveAt(validTime, systemTime)` returns a `postgres.BoolExpression` for WHERE clauses
- `ActiveAtParams(validTime, systemTime)` returns `[]interface{}` for query parameter binding

These two functions maintain synchronized parameter order and semantics. **They must be used together:**
callers construct a single query using `ActiveAt()` in the WHERE clause and pass the same
`validTime` and `systemTime` values to `ActiveAtParams()` for parameter binding. Any call site
that uses these helpers should include a comment referencing its nearby call site and linking back
to this paragraph. This pattern ensures that changes to temporal filtering logic propagate correctly
across all query sites.

## Mutation Entry Point

Domain services orchestrate all state changes through `ExecuteBitemporalMutation()`. This function:

- Accepts entities with zero ID (new) or populated ID (existing)
- Acquires advisory locks on entity IDs
- For existing entities: performs staleness check
- Executes the Flush/Purge/Assert pipeline
- Returns mutated entities with IDs populated

**Transaction time is always set by the database.** EBM queries `SELECT CURRENT_TIMESTAMP` once at the start of every mutation and uses that value for all `transaction_start` boundaries written during that call. Domain services must never call `time.Now()` or set `TransactionStart` on entity structs — the database clock is authoritative.

**`maxTransactionStart` and staleness detection.** The first time parameter to `ExecuteBitemporalMutation` is the caller's last-known maximum `transaction_start` for the entity, used to detect concurrent mutations. Pass `time.Time{}` (zero value) to skip staleness detection — EBM treats a zero value as "caller has no prior read." This is correct for domain services that do not yet implement optimistic concurrency control. When a non-zero value is passed and the database has a newer `transaction_start`, EBM returns a staleness error.

See `../../domain/brief.md` for how domain services use this entry point.

## Caller-Facing Operations

Domain services call these functions from serde/rdbms:

- **Mutations** (`mutation.go`)
  - `ExecuteBitemporalMutation`
  - `DefaultBitemporalInserter`
- **Queries** (`read_path.go`)
  - `IsActive`
  - `IsNoRows`
  - `ResolveEntityID`
- **Types** (`types.go`)
  - `LockableEntity`
  - `BitemporalMutationBuilder`
- **Constants** — use `c.InfinityTime` via `import c "chronocone/backend/constants"`. The sentinel is no longer re-exported from this package; see `constants/brief.md`.

## When you work here

- This is where complex, test-heavy database logic lives. Errors here propagate to all workloads.
- Write comprehensive tests for mutation scenarios, edge cases, and concurrent access
- Document any PostgreSQL-specific assumptions or version constraints
- Error messages must help FDEs diagnose data consistency issues in the field

## References

- **Critical**: `memory-bank/specifications/architecture/database_design.md` — Bitemporal honesty invariants

## Lessons Learned

**All tests must run against embedded PostgreSQL. Never mock the database layer.**

This layer implements complex, safety-critical bitemporal semantics. Tests here are the last line of defense against data consistency violations. Do not use mocks or test doubles:

1. **Always use the embedded PostgreSQL fixture** (see `setup_test.go`). It provides a clean, isolated database for each test.
2. **Test the full Flush/Purge/Assert pipeline**, not just individual SQL statements. The invariants depend on the orchestration working correctly under concurrency and edge cases.
3. **Test edge cases and failure paths**:
   - Staleness conflicts (concurrent mutations on the same entity)
   - Validity range overlaps and boundaries
   - Advisory lock timeouts and deadlock scenarios
   - Large batch mutations (many entities in one EBM call)
   - Invalid input that violates schema constraints
4. **Verify database state, not just return values**. Assert that the active/historical partitions contain exactly what they should, with correct mode markers and transaction metadata.

Why: Bitemporal mutations are difficult to reason about in isolation. A mocked test can pass while the real database query is broken. The EBM pipeline is a tightly coupled sequence; breaking one step (e.g., forgetting to update `transaction_start` in the FLUSH) silently produces incorrect results. Only a real database test can catch these bugs.

See `memory-bank/specifications/patterns/backend-testing-design.md` for the complete testing strategy across all backend layers and why persistence tests focus on comprehensive edge cases and the full EBM pipeline.

**Handle Go-Jet 'no rows' errors explicitly**

When `QueryContext` hydrates into a struct (not a slice), Go-Jet's `qrm` returns an error if no rows are found. The error string is `"qrm: no rows in result set"`. Always check for both `sql.ErrNoRows` and `qrm.ErrNoRows` (or the string) so a missing record is handled as a domain concern rather than a fatal database failure. Never let a "not found" path propagate as an unexpected error.

**`ORDER BY` is mandatory whenever a query uses `LIMIT`**

A `LIMIT N` without `ORDER BY` returns a planner-dependent subset of rows, not the intended rows. This produces non-deterministic results that depend on physical row order and planner statistics. Even when the result is logically single-valued, always write the `ORDER BY` to document intent and prevent planner regressions across PostgreSQL versions or schema changes.
