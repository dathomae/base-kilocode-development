# Story 043E: EBM Splinter Correctness and History Preservation

## Goal

Fix the bitemporal mutation orchestrator (`ExecuteBitemporalMutation`, "EBM")
so that **every** superseded active row is preserved in the historical
partition before being purged, and so that no active row carries a
`transaction_start` value that lies about when the system formed the belief
it currently expresses. This is the foundational correction that brings the
implementation into agreement with the philosophy: **EBM owns time;
callers own business logic and only ever emit `INSERT` statements for the
new assertion they want recorded.**

This story also corrects two design documents whose literal text contradicts
that philosophy (`database_design.md` §7 step 3 and Rule 10), so that future
agents read a single, consistent specification.

## References

- [database_design.md](specifications/architecture/database_design.md) — §3, §6, §7, §8
- [layered-architecture.md](specifications/patterns/layered-architecture.md) — §2 Domain Service Layer
- [service-organization.md](specifications/patterns/service-organization.md) — §"Bitemporal Mutation Protocol"
- [Story043_temporal-enforcement.md](finished-stories/Story043_temporal-enforcement.md)
- [Story043B_active_historical_partitioning.md](Story043B_active_historical_partitioning.md)

## Dependencies

None — this story is the foundational fix and unblocks the rest of the
043-corrective series.

## Dependent Stories

- [Story043F_read_path_correctness.md](Story043F_read_path_correctness.md) — Not Started (benefits from EBM's new invariants when constructing read-path tests)
- [Story043G_test_integrity.md](Story043G_test_integrity.md) — Not Started (rewrites EBM tests; must follow this story's contract changes)
- [Story043H_registry_and_mutation_hardening.md](Story043H_registry_and_mutation_hardening.md) — Not Started (extends the registry; benefits from EBM owning all temporal SQL)
- [Story043I_docs_and_scaffold_cleanup.md](Story043I_docs_and_scaffold_cleanup.md) — Not Started (the doc rewrites in this story unblock further wording reconciliation in 043I)

## Tasks

1. **Drop `LockableEntity.MaxTransactionStart()` from the interface** — Completed
   a. Write a failing compile-time check: a no-op test file in `serde/rdbms/` that satisfies `LockableEntity` without implementing `MaxTransactionStart`. - Completed
   b. Remove `MaxTransactionStart() time.Time` from the `LockableEntity` interface in `serde/rdbms/mutation.go`. - Completed
   c. Remove the corresponding method from `nfWrapper` in `mutation_test.go` and any other wrapper implementations introduced during this story. - Completed
   d. Verify `go build ./...` from `src/backend/` succeeds. - Completed

2. **Add three failing integration tests that pin down bitemporal honesty** — Completed
   a. `TestEBM_FutureRecordIsFlushedBeforePurge`: seed a future-dated active row `[t2, ∞)@tx0`, run a mutation effective at `t1` with `t0 < t1 < t2`. Assert that after the mutation, exactly one historical row exists with `validity_start = t2`, `validity_end ≈ original_validity_end`, and `transaction_end ≈ now()`. - Completed
   b. `TestEBM_ActiveRowsDoNotLieAboutTransactionStart`: after any future-dated mutation, assert that no active row for the affected entity has `transaction_start < now() - epsilon`, where `epsilon` is generous enough to absorb clock skew but small enough to reject a stale `transaction_start` from before the mutation. - Completed
   c. `TestEBM_BeliefAtUVPointIsUnique`: for a fixed `(entity_id, validity_T, transaction_U)` covering a point that was historically known and then revised, assert that exactly **one** row across the active + historical partitions matches the §3 predicate `validity_start <= T AND validity_end > T AND transaction_start <= U AND U < transaction_end`. - Completed
   d. All three tests **must fail** against the current implementation. Run the suite and record the failure modes in the task notes for verification on the way back. - Completed (all three fail: (a) historical=0 instead of 1 — future row purged without being flushed; (b) truncated active row carries the seeded stale transaction_start instead of a fresh one; (c) two rows match the UV predicate — the historical row and the active row that lies about its transaction_start.)

3. **Rewrite the EBM mutation pipeline so all three tests pass** — Completed
   a. In `serde/rdbms/mutation.go`, change the Phase 5 ordering and predicates:
      - **FLUSH** must capture **every** active row that will be affected, including rows starting at or after `effExpr` (the current code's `validity_start < $2 AND validity_end > $2` filter is too narrow). The correct filter is "any active row that will be modified, deleted, or shortened by this mutation". - Completed (filter is now `mode='active' AND id = ANY($1) AND validity_end > $2`)
      - **PURGE** must `DELETE` only rows that have been FLUSHED in this transaction. - Completed (added `mode='active'` guard so historical rows just inserted by FLUSH are not re-deleted)
      - **TRUNCATE** is permitted to remain inside EBM as an `UPDATE` (callers still can't emit it; the type system continues to enforce that), **but** it MUST refresh `transaction_start = CURRENT_TIMESTAMP` on every row whose `validity_end` it modifies. A truncation without a refreshed `transaction_start` is forbidden by the bitemporal-honesty invariant. - Completed (UPDATE now sets both `validity_end = effExpr` and `transaction_start = CURRENT_TIMESTAMP`)
   b. Verify that the historical partition's GiST exclusion constraint never fires under the new pipeline (the FLUSH `transaction_end = CURRENT_TIMESTAMP` makes the historical interval disjoint from `[now, ∞)`). - Completed (`TestHistoricalExclusionConstraint` continues to pass alongside the three new tests)
   c. Verify the three tests from Task 2 now pass. - Completed (all pass; full `./serde/rdbms/` suite green)

4. **Audit and update the existing EBM test suite** — Completed
   a. Update `TestExecuteBitemporalMutation/SupersedeFutureRecord` to additionally assert that R2 (the pre-existing future row that gets superseded) is present in the historical partition with a capped `transaction_end`. - Completed
   b. Update `TestExecuteBitemporalMutation/FutureDatedMutation` to additionally assert that the truncated original row's `transaction_start` is fresh (≥ test start time), not the seeded `now`. - Completed (seed now uses an explicitly-stale value so the freshness comparison is robust against `transaction_timestamp` returning the moment of `BEGIN`)
   c. Confirm `TestBatchBitemporalImport` still passes. If it does not, adjust the test to match the new historical-preservation behavior (10,000 historical rows distributed across hash partitions remains the goal). - Completed (passes unchanged: 5003+4997 rows across the two hash partitions)

5. **Reconcile `database_design.md` §7 with the EBM-owns-time philosophy** — Completed
   a. Rewrite §7 step 3 to make EBM the splinter computer and constrain callers to `[]postgres.InsertStatement`. - Completed
   b. Rewrite Rule 10 to scope interval-algebra-in-Go to EBM and explicitly forbid `UPDATE`/`DELETE` outside `serde/rdbms/mutation.go`. - Completed
   c. Replace §7's "Rationale" paragraph (which previously framed in-memory splinter computation as an end in itself) with one that names the two bitemporal-honesty invariants — preservation and honest active state — directly. - Completed

6. **Reconcile `layered-architecture.md` §2 step 3 and `service-organization.md` §"Bitemporal Mutation Protocol" step 3** — Completed
   a. In both documents, change "Compute New Splinters: Calculate the prefix, new assertion, and suffix versions" to "Compute New Splinters: **EBM** calculates the prefix, new assertion, and suffix versions in Go and emits whichever SQL statements most efficiently express them. Callers provide only the new-assertion `INSERT` statements via the builder callback." - Completed
   b. Add a cross-reference from `database_design.md` §7 to `layered-architecture.md` §2 ("executed under advisory lock after staleness check; see `layered-architecture.md` §2") and vice versa. - Completed (forward link added in §7 step 3 during Task 5; back-links added from both `layered-architecture.md` §2 step 3 and `service-organization.md` step 3 to `database_design.md` §7)

7. **Tighten staleness wording in `service-organization.md`** — Completed
   a. Replace "If they differ, the record was modified by another process" with "If the DB's max `transaction_start` is **greater than** the caller's known value, the record was modified by another process". - Completed

## Test-First Development

Tasks 1 and 2 are explicit test-first: the interface change (Task 1) has a
compile-time check, and the three pin-down tests (Task 2) must fail before
the rewrite in Task 3. Tasks 5–7 are pure documentation and have no
automated test, but each doc passage should be re-read end-to-end after
editing to confirm it is consistent with the rest of the document.

## Constraints

- **No callers may emit `UPDATE` or `DELETE`** against bitemporal tables.
  The `BitemporalMutationBuilder[T]` signature must continue to return only
  `[]postgres.InsertStatement`. This is the structural constraint that
  protects the philosophy regardless of how EBM is implemented internally.
- **`transaction_start` may never be supplied by application code.**
  Inserts must rely on the column's `DEFAULT transaction_timestamp()`;
  EBM's internal `UPDATE` (where it persists) must `SET transaction_start =
  CURRENT_TIMESTAMP` so the DB clock is the single source of truth.
- **No new external dependencies.** Stick to `lib/pq`, `pgx`, `go-jet/v2`,
  and the standard library.
- **Tests use the embedded postgres harness already in `setup_test.go`.**
  Do not introduce a parallel test infrastructure.

## Intent

- Task 1: remove the dead-code interface obligation that misleads readers
  about how staleness is checked.
- Task 2: pin down, in executable form, the three invariants that distinguish
  a correct bitemporal pipeline from one that merely happens to pass today's
  weak tests.
- Task 3: bring the implementation into agreement with those invariants
  with the least architectural disruption — keeping the in-place `UPDATE`
  as an internal optimization but making it honest.
- Tasks 4 and beyond: ensure that the published specifications match the
  newly correct implementation, so future agents do not regenerate the
  old bugs from old documentation.

## Acceptance Criteria

- [x] `LockableEntity` no longer declares `MaxTransactionStart()`.
- [x] `TestEBM_FutureRecordIsFlushedBeforePurge`,
  `TestEBM_ActiveRowsDoNotLieAboutTransactionStart`, and
  `TestEBM_BeliefAtUVPointIsUnique` all pass against the new EBM.
- [x] After any EBM mutation, the SQL predicate
  `SELECT count(*) FROM <table> WHERE entity_id = $1 AND mode = 'active'
   AND transaction_start < $tx_start_of_mutation` returns zero across all
  affected entities in the test suite. (Verified in `TestEBM_ActiveRowsDoNotLieAboutTransactionStart` and the augmented `TestExecuteBitemporalMutation/FutureDatedMutation`.)
- [x] `database_design.md` §7 step 3, Rule 10, and the §7 Rationale
  paragraph have been rewritten per Task 5; `layered-architecture.md` and
  `service-organization.md` have matching language per Task 6.
- [x] `grep -nE "Computing splinters in-memory \(rather than via SQL UPDATE\)"
  memory-bank/` returns zero hits.
- [x] All previously-passing EBM tests continue to pass after their updates
  in Task 4.

## Branch

`claude/story043E-ebm-splinter-correctness`
