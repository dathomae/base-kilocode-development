# Story 043: Temporal Enforcement

## Status
- **State**: COMPLETED
- **Progress**: 100%

## Objective
Enforce strict transaction-level consistency and identity protection for bitemporal mutations. This includes handling both "effective now" and future-dated assertions to support planned changes and complex timeline management.

## Tasks
1.  **[DONE] Advisory Lock Protocol**
    *   Implement `AdvisoryLockKey` packing (10-bit type / 54-bit ID).
    *   Implement `AcquireLocks` with deterministic sorting to prevent deadlocks.
    *   Add safety checks to prevent recursive locking within the same transaction.

2.  **[DONE] Structural Mutation Orchestrator**
    *   Implement `ExecuteBitemporalMutation` with the 5-phase lifecycle:
        1.  **Group**: Sort entities by table.
        2.  **Lock**: Acquire advisory locks.
        3.  **Staleness Check**: Verify `transaction_at` hasn't diverged.
        4.  **Compute**: Execute user-provided mutation logic.
        5.  **Write**: Atomic termination of old records and insertion of new records.
    *   Integrate PostgreSQL `CURRENT_TIMESTAMP()` as the transaction-stable anchor.
    *   Enforce mandatory positive ID checks for bitemporal updates.

3.  **[DONE] Key Suspension and Reactivation**
    *   Update `LockableEntity` interface to include `Key() string`.
    *   Implement `ResolveEntityID` helper to find stable IDs by key (including tombstoned keys).
    *   Enforce key immutability: Orchestrator errors if a mutation attempts to change an entity's key.
    *   Refactor `Upsert` methods in `rdbms.go` to use the orchestrator and handle reactivations.

4.  **[DONE] Go-Jet Temporal Query Helpers**
    *   Create `temporal_expressions.go` with standard helpers:
        *   `IsCurrent`: Filters for active valid records.
        *   `AsOf`: Filters for records valid at a specific point in time.
        *   `BelievedAt`: Filters for state known at a specific transaction time.
        *   `RecordedAfter`: Filters for records created after a timestamp.
    *   Refactor all `List` and `Get` methods in `rdbms.go` to use these helpers.
5.  **[DONE] Future-Dated Assertions**
    *   Update `ExecuteBitemporalMutation` to accept an optional `effectiveAt`.
    *   Implement guard: `effectiveAt >= CURRENT_TIMESTAMP`.
    *   Generalize truncation logic to handle closing records at `effectiveAt`.
    *   Handle timeline "overwrite" vs "append" scenarios.
    *   Add comprehensive tests for timeline edge cases.

## Verification
- `mutation_test.go`: Comprehensive integration tests covering staleness, key immutability, ID validation, and reactivation.
- Store Layer Build: Verified that `rdbms.go` compiles and correctly uses the new orchestrator and query helpers.
