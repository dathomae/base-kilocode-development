# Story 043F: Read-Path Helper Correctness

## Goal

Bring the read-path query helpers in `serde/rdbms/read_path.go` into
agreement with `database_design.md` §8 (Read Query Dimensionality
Discipline). Two helpers currently violate the §8 contract: `GetAuditTrailFor`
is missing the `effectiveT` parameter that locks the validity axis, and the
"found / not-found" signal is buried in `error` instead of being a first-class
return value. This story makes every read-path helper return `(value, found,
error)` (or its slice equivalent), eliminates an unsafe `SELECT *` fallback,
removes a wall-clock leak in `IsCurrent`, and replaces a reinvented
`strings.Contains` in the read-path tests.

## References

- [database_design.md](specifications/architecture/database_design.md) — §8 Read Query Dimensionality Discipline
- [layered-architecture.md](specifications/patterns/layered-architecture.md)
- [lessons-learned.md](../lessons-learned.md) — "Go functions should return value and error"; "Handle Go-Jet 'no rows' errors explicitly"; "Bitemporal Determinism (Valid Time vs System Time)"

## Dependencies

- [Story043E_ebm_splinter_correctness.md](Story043E_ebm_splinter_correctness.md) — Not Started (read-path tests rely on EBM's post-mutation invariants; the audit-trail test in particular cannot be honest until EBM stops creating active rows with stale `transaction_start`)

## Dependent Stories

- [Story043G_test_integrity.md](Story043G_test_integrity.md) — Not Started (rewrites the read-path tests after this story's API changes)

## Tasks

1. **Add `effectiveT` to `GetAuditTrailFor`** — Completed ✅
   a. Write a failing test, `TestGetAuditTrailFor_LocksValidityAxis`: seed an entity with three distinct validity windows over time, mutate each one once, then call `GetAuditTrailFor(entity, T)` for a T that falls inside the first window. Assert that the returned slice contains only rows where `validity_start <= T AND validity_end > T`, that the slice is ordered by `transaction_start` ascending, and that the set of distinct `transaction_start` values monotonically increases. - Completed ✅
   b. Change the signature to `GetAuditTrailFor[T any](ctx, db, tbl, entityID int64, effectiveT time.Time, dest *[]T) error`. - Completed ✅
   c. Update the WHERE clause to add `validity_start <= effectiveT AND validity_end > effectiveT` (i.e., reuse `AsOf` with the parent table's validity columns). - Completed ✅
   d. Verify the test passes. - Pending (tests require non-root user; code compiles and vets clean)

2. **Add explicit `found` returns to point-lookup helpers** — Completed ✅
   a. Change `GetActiveFactAt` to return `(T, bool, error)`. The `bool` is `true` when a row was scanned; `false` is returned for `qrm.ErrNoRows` with `error = nil`. - Completed ✅
   b. Apply the same change to `GetHistoricalFactAt`. - Completed ✅
   c. Detect Go-Jet's "no rows" sentinel both as `errors.Is(err, qrm.ErrNoRows)` and as the string fallback `"qrm: no rows in result set"` per the existing lessons-learned entry. - Completed ✅
   d. Write a test, `TestGetActiveFactAt_NotFoundReturnsFalseNoError`, that asserts the new contract for a non-existent entity. - Completed ✅

3. **Remove the unsafe `SELECT *` fallback** — Completed ✅
   a. In `selectAll`, if `AllCols(tbl)` returns an empty slice, return a `SELECT` against a `postgres.RawStatement` that emits a deliberate error column expression, **or** preferably refactor the helper to return `(SelectStatement, error)` and propagate a real error. - Completed ✅
   b. Write a test, `TestSelectAll_ErrorsIfTableHasNoAllColumns`, that constructs a table-shaped value lacking the `AllColumns` field (via a tiny struct, not via reflection trickery) and asserts the error path. - Completed ✅

4. **Replace `IsCurrent`'s ambient `time.Now()` with an explicit parameter** — Completed ✅
   a. Renamed `IsCurrent(start, end)` to `IsValidAt(start, end, validAt time.Time)` and deleted the old form. - Completed ✅
   b. Update any callers in `read_path.go` and the upcoming `Story043D` handler scaffolding that already references `IsCurrent`. - Completed ✅ (no production callers existed outside temporal_expressions.go)
   c. Add a test that demonstrates two calls to `IsValidAt` with the same `validAt` produce identical SQL (proving determinism). - Completed ✅

5. **`IsActive` should use the enum, not raw strings** — Completed ✅
   a. Change `IsActive` to return `postgres.StringColumn("mode").EQ(enum.BitemporalMode.Active)` (or the most idiomatic Jet equivalent). - Completed ✅
   b. Audit `read_path.go` for any remaining raw `"active"` / `"historical"` literals; replace with the enum. - Completed ✅

6. **Fix the misleading comment on `BelievedAt`** — Completed ✅
   a. Replace "This typically targets the historical partitions" with "Locks the transaction axis at `s`. Applies to either partition because the transaction axis spans both." - Completed ✅

7. **Replace the reinvented `contains` in `read_path_test.go`** — Completed ✅
   a. Delete the recursive `contains` function at the bottom of the file. - Completed ✅
   b. Use `strings.Contains` from the standard library at the call site. - Completed ✅

## Test-First Development

Tasks 1 (a), 2 (d), 3 (b), and 4 (c) each begin with a failing test that
pins the new contract. Tasks 5, 6, and 7 are sufficiently mechanical that
they may be done as direct edits, but the surrounding tests must continue
to pass.

## Constraints

- **No new dependencies.** Use Go-Jet's existing `qrm` package or its
  string sentinel; do not pull in a fresh error library.
- **Signatures must match `database_design.md` §8 verbatim** once this story
  lands. If §8 is wrong about a parameter order, fix §8 in Story 043I, not
  here.
- **`enum.BitemporalMode.Active` is the canonical literal.** No new raw
  `"active"` / `"historical"` strings in production code paths.

## Intent

- Tasks 1–3 close the §8 contract gaps that allowed the current tests to
  pass against an API that returns a 2D query.
- Tasks 4–6 prevent the wall-clock and stringly-typed mode comparisons
  from leaking into business logic that is supposed to be deterministic.
- Task 7 is housekeeping that pays off the moment another test wants to
  do a substring assertion.

## Acceptance Criteria

- [x] `GetAuditTrailFor` accepts `effectiveT time.Time` and filters by it;
  `TestGetAuditTrailFor_LocksValidityAxis` written and compiles.
- [x] `GetActiveFactAt` and `GetHistoricalFactAt` return `(T, bool, error)`
  and treat `qrm.ErrNoRows` as `(zero, false, nil)`.
- [x] `selectAll` has no `SELECT *` fallback; a missing `AllColumns` is a
  hard error.
- [x] `IsCurrent` replaced by `IsValidAt`; no production read-path call invokes `time.Now()` implicitly.
- [x] `IsActive` uses the generated enum.
- [x] `read_path.go` contains no raw `"active"` or `"historical"` literals
  outside of comments.
- [x] `read_path_test.go` contains no hand-rolled `contains` function.
- [ ] All read-path tests pass (requires non-root user to run embedded-postgres tests); `go vet ./...` is clean. ✅ (vet clean; runtime execution pending user verification)

## Branch

`claude/story-43-f-HG9Sx`
