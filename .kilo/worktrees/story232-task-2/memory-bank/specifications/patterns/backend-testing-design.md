# Backend Testing Design

Testing strategy for the backend layer, organized by responsibility and test audience.

## Overview

The backend uses a **layered testing approach** where each layer has distinct testing responsibilities and serves different audiences:

- **Persistence & Bitemporal Orchestration** (`serde/rdbms/`) — Comprehensive edge case testing
- **Domain Services** (`domain/`) — Business logic and temporal correctness testing
- **API Handlers** (`handlers/`) — User-focused examples and documentation

Each layer tests against embedded PostgreSQL (never mocks). The tests serve dual purposes:
1. **Verify correctness** — Catch runtime bugs and integration issues
2. **Document intent** — Show users, FDEs, and new developers how the system is meant to be used

## Persistence & Bitemporal Orchestration (`serde/rdbms/`)

**What to test:** The complete Flush/Purge/Assert (EBM) pipeline, advisory locking, staleness detection, and bitemporal invariants.

**Why:** This layer implements complex, safety-critical bitemporal semantics. Errors propagate to all workloads. A mocked test can pass while the real database query is broken. The EBM pipeline is tightly coupled; breaking one step produces incorrect results silently.

**Test style:** Implementation-focused edge cases covering:
- Staleness conflicts (concurrent mutations on the same entity)
- Validity range overlaps and boundaries
- Advisory lock timeouts and deadlock scenarios
- Large batch mutations
- Invalid input that violates schema constraints
- Verify database state after each operation (active/historical partitions, mode markers, transaction metadata)

**Audience:** Developers maintaining bitemporal logic. These tests document the safety guarantees the layer provides to domain services.

## Domain Services (`domain/`)

**What to test:** Business rule enforcement, namespace scoping, temporal integrity, and orchestration with persistence.

**Why:** Domain services use go-jet queries, so they have database dependencies. Mocking those dependencies lets logical bugs hide. A service function can "pass" a mock test while actually constructing invalid SQL or missing critical WHERE clauses. Real database tests catch these immediately.

**Test style:** Edge cases in business logic covering:
- Boundary conditions (empty result sets, single rows, large batches)
- Temporal edge cases (validity overlaps, transitions at exact boundaries)
- Multi-namespace scenarios (foreign assertions, scope violations)
- Error paths (missing parent entities, constraint violations, staleness conflicts)
- Verify both return values AND database state after each operation

**Audience:** Developers building domain services. These tests document what the service guarantees and which edge cases are handled.

## API Handlers (`handlers/`)

**What to test:** HTTP mechanics (parsing, routing, serialization) and the integration of handler + domain service + database.

**Why:** Code that compiles and structurally follows patterns can have subtle runtime bugs undetectable at compile time: type mismatches (int32 vs int64), missing temporal filters, serialization errors, failed parent entity checks.

**Test style:** User-focused example scenarios showing:
- A realistic user action (e.g., "planner creates a family")
- Clear BEFORE state (what's in the database or system beforehand)
- The HTTP request the user makes
- Clear AFTER state (what the API returns, what's now in the database)
- Use realistic entity names and workflows, not test suffixes
- Verify both HTTP response AND resulting database state

**Example structure:**
```go
t.Run("Example_PlannerCreatesAndRetrievesFamily", func(t *testing.T) {
    // SCENARIO: A planner sets up a new planning hierarchy root
    // BEFORE: No "acmecorp" family in the system
    
    // The planner creates the family
    createResp, createBody := doRequest(router, "PUT", "/family/acmecorp", ...)
    
    // AFTER: The family is created and returned
    assert.Equal(t, http.StatusOK, createResp.StatusCode)
    var created familyOutput
    json.Unmarshal(createBody, &created)
    assert.Equal(t, "acmecorp", created.Key)
    
    // The planner can retrieve it
    getResp, getBody := doRequest(router, "GET", "/family/acmecorp", nil)
    // verify...
})
```

**Audience:** Customers, FDEs, and new team members learning how to use the API. These tests serve as living documentation of the intended user workflows. A customer reading these tests understands "here's how I interact with the system."

## Overlap and Duplication

Some functionality is tested at multiple layers (e.g., "create entity, verify state" appears in both domain and handler tests). This is healthy, not wasteful:

- **Domain test** for "create entity, verify state" catches business logic bugs
- **Handler test** for the same action catches HTTP mechanics, serialization, and response-building bugs
- They test different things at different layers

The handler test duplication is acceptable because it serves a different purpose: documentation. A handler test answers "how do I use this API?" in a way that a domain test (focused on edge cases) does not.

## When to Use Mocks

**Never use mocks for:**
- Database queries (test against embedded PostgreSQL instead)
- Any layer that has database dependencies

**OK to mock:**
- External services at system boundaries (if the system integrates with external APIs)
- Only if the external service is already documented as unreliable or unavailable during testing

## References

- `src/backend/serde/rdbms/brief.md` — Bitemporal testing specifics
- `src/backend/domain/brief.md` — Domain service testing specifics
- `src/backend/handlers/brief.md` — Handler testing and user-focused examples
- `memory-bank/specifications/architecture/database-design.md` — Bitemporal honesty invariants
