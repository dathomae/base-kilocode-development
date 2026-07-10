# Story 043G: Test-Suite Integrity (Deep Assertions, Random Inputs, No No-Ops)

## Goal

Repair the bitemporal test suite so it actually defends the invariants it
claims to defend. Today, two tests in `domain/entities/` assert nothing
real, one schema-defaults test downgrades a wrong-value check to a log
statement, and every EBM/read-path test uses hard-coded magic keys and IDs
that the lessons-learned guidance "Enforcing TDD Integrity via Deep
Assertions" explicitly forbids. The shared suite-wide transaction in
`domain/entities/setup_test.go` also makes every row share a single
`transaction_timestamp()`, which makes bitemporal staleness tests
structurally unable to fail.

This story rewrites the tests to use randomized inputs and assert
post-condition database state rather than surface-level error codes. It
also removes two no-op test files outright (code-reviewer confirmed they
provide zero coverage, and the user has explicitly approved their
deletion).

## References

- [lessons-learned.md](../lessons-learned.md) — "Enforcing TDD Integrity via Deep Assertions"; "Truth in Documentation: Synchronize Story State with Verified Reality"
- [resources/go/testing-guide.md](../../resources/go/testing-guide.md)
- [database_design.md](specifications/architecture/database_design.md) — §3, §7

## Dependencies

- [Story043E_ebm_splinter_correctness.md](Story043E_ebm_splinter_correctness.md) — Not Started (EBM tests must be rewritten against the corrected pipeline)
- [Story043F_read_path_correctness.md](Story043F_read_path_correctness.md) — Not Started (read-path tests must be rewritten against the corrected signatures)

## Dependent Stories

None.

## Tasks

1. **Delete no-op tests** — Not Started
   a. Delete `src/backend/domain/entities/category_test.go` (a five-line
      file that constructs `Category{}` and asserts nothing). - Not Started
   b. Delete `src/backend/domain/entities/category_sql_test.go` (calls a
      stub that returns `(nil, nil)` and asserts the resulting slice is
      empty — structurally guaranteed to pass). - Not Started
   c. Confirm `go test ./...` still passes for `src/backend/`. - Not Started

2. **Promote the `transaction_end` check in `schema_test.go` to a hard assertion** — Not Started
   a. In `TestBitemporalDefaultsPopulation` (`serde/rdbms/schema_test.go`),
      change `if txEnd != "3000-01-01 00:00:00+00" { t.Logf(...) }` to
      `if txEnd != "3000-01-01 00:00:00+00" { t.Errorf(...) }`. - Not Started
   b. Confirm the test still passes against the current schema. - Not Started

3. **Replace all hard-coded magic strings and IDs in `mutation_test.go`** — Not Started
   a. Introduce a per-test helper, e.g., `randomTestKey(t *testing.T)
      string` that returns a UUID-suffixed key of 6–64 characters, and
      `randomTestID(t *testing.T) int64` that returns an ID in a reserved
      range (e.g., `[1_000_000, 1_999_999]` to avoid collision with seed
      fixtures). - Not Started
   b. Replace every literal `"test-fam"`, `"original-key"`, `"changed-key"`,
      `"future-fam"`, `"supersede-fam"`, `"past-fam"`, and numeric IDs like
      `1000`, `1100`, `4000`, `5000`, `6000` with calls to these helpers. - Not Started
   c. For each test, capture the random key/ID into a local variable at the
      top of the test, then use that variable in every assertion. - Not Started
   d. Where a test asserts that an entity was inserted, the assertion must
      read the row back from the database and verify the persisted key/ID
      matches the random value (deep persistence verification, per
      lessons-learned). - Not Started

4. **Replace all hard-coded magic strings and IDs in `read_path_test.go`** — Not Started
   a. Apply the same random-input pattern to `TestReadPathHelpers` and
      `TestHistoricalExclusionConstraint`. - Not Started
   b. The `EXPLAIN`-based `GetActiveFactAt_PruningProof` test should
      embed the random ID in the EXPLAIN query string (still safe because
      it is a generated integer, not user input). - Not Started

5. **Replace all hard-coded magic strings and IDs in `batch_test.go`** — Not Started
   a. Replace the `id := int64(10000 + j)` and `key := fmt.Sprintf("batch-fam-%d", j)` pattern with a random per-row pair drawn from the helpers in Task 3. - Not Started
   b. Keep the batch size (10,000) deterministic for performance verification. - Not Started

6. **Refactor `domain/entities/setup_test.go` away from the suite-wide transaction** — Not Started
   a. Remove the package-level `suiteTx` and the savepoint-based `RunTx`
      wrapper. - Not Started
   b. Replace with a `RunTx(t, func(t, ctx))` that opens a **fresh
      `pgx.Tx` per test** from the shared connection (or pool), applies
      the schema once at `TestMain` against the database (not against a
      transaction), and rolls back per test. - Not Started
   c. Add a regression test, `TestSetupTx_TransactionStartIsFresh`, that
      runs `RunTx` twice in sequence and asserts the two recorded
      `transaction_timestamp()` values differ. - Not Started

7. **Document the `EnableLockSafetyChecks` correctness dependency** — Not Started
   a. In `TestAcquireLocks/PanicOnSecondCall` (`locks_test.go`), add
      `require.True(t, EnableLockSafetyChecks)` as the first line so the
      test fails loudly if the const is ever removed or flipped. - Not Started
   b. (Coordinate with Story 043I, which will remove the const entirely
      and make the safety check unconditional; once that lands, this
      assertion is no longer needed and the require line should be
      dropped.) - Not Started

8. **Add a new lessons-learned entry: `ORDER BY` is mandatory with `LIMIT`** — Not Started
   a. Append to `memory-bank/lessons-learned.md`:
      "**`ORDER BY` is mandatory whenever a query uses `LIMIT`**: A `LIMIT
      N` without an `ORDER BY` returns a planner-dependent subset of rows,
      not the rows the author intended. This produces non-deterministic
      results that depend on physical row order, statistics, or partition
      pruning. The canonical example is `ResolveEntityID` (Story 043H),
      which returned a wrong tombstoned entity ID on key reuse. Even when
      the result is logically single-valued, write the `ORDER BY` to
      document intent and prevent planner regressions." - Not Started

## Test-First Development

This story is primarily a test-rewriting effort. The new tests added in
Tasks 6 (c) follow test-first: the regression test should be written and
shown to fail against the old suite-wide-transaction setup before the
refactor in 6 (b) is applied.

## Constraints

- **No hard-coded primary keys, foreign keys, or business-string literals
  in any test under `serde/rdbms/` or `domain/entities/`.** The only
  allowed literals are infrastructure constants (port numbers, schema
  paths, `bitemporal_mode` enum values).
- **Per-test transaction isolation.** No test may depend on side effects
  from a previous test in the same package.
- **Random-input helpers live in a single shared file** (e.g.,
  `serde/rdbms/test_random.go`) so they can be reused by Stories 043H and
  any future story that touches these tests.
- **Test names should describe the invariant being verified**, not the
  code path being exercised (e.g., `TestEBM_BeliefAtUVPointIsUnique` not
  `TestExecuteBitemporalMutation_HappyPath`).

## Intent

- Tasks 1 and 2 close the trivially-passing-tests loophole.
- Tasks 3–5 are the largest payoff: random inputs prevent both the agents
  who write code and the agents who write tests from gaming via
  hard-coded special cases.
- Task 6 makes the entire `domain/entities/` test package capable of
  detecting bitemporal staleness, which it currently cannot because
  every row shares a single `transaction_timestamp()`.
- Tasks 7 and 8 are documentation that prevents the same class of bugs
  in future stories.

## Acceptance Criteria

- [ ] `src/backend/domain/entities/category_test.go` and
  `category_sql_test.go` no longer exist.
- [ ] `grep -rE '"(test-|original-|changed-|supersede-|future-|past-|batch-|read-path-)[a-z-]+"' src/backend/serde/rdbms/*_test.go src/backend/domain/entities/*_test.go` returns zero hits.
- [ ] `grep -nE '\bint64\(1[0-9]{3,4}\)' src/backend/serde/rdbms/*_test.go` returns zero hits (no magic numeric IDs).
- [ ] `TestBitemporalDefaultsPopulation` uses `Errorf`, not `Logf`, for the
  `transaction_end` check.
- [ ] `TestSetupTx_TransactionStartIsFresh` passes (two sequential
  `RunTx` invocations observe distinct `transaction_timestamp()` values).
- [ ] A new lessons-learned entry "`ORDER BY` is mandatory with `LIMIT`"
  is present in `memory-bank/lessons-learned.md`.
- [ ] All tests pass on `make test` from `src/backend/`.

## Branch

`claude/story043G-test-integrity`
