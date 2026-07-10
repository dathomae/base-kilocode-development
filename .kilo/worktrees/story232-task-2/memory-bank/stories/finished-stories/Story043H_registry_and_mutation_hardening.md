# Story 043H: Bitemporal Registry Honesty and EBM Hardening

## Goal

Remove three sources of latent bugs from the bitemporal infrastructure:

1. `BitemporalTableRegistry` derives the primary-key column name from
   `tbl.Alias() + "_id"`, which silently misnames the identity columns of
   junction tables (`metric_dim`, `transformer_input`,
   `transformer_output`, `shifter_input`, `shifter_output`, and the
   localization junction tables) that have composite keys.
2. `ResolveEntityID` uses `LIMIT 1` with no `ORDER BY`. For a key that has
   been tombstoned and then reactivated for a different entity, the
   returned ID is planner-dependent and frequently wrong.
3. EBM's raw-SQL `flushSQL` (via `fmt.Sprintf`) and the `getTableSpecificCols`
   reflection-then-string-join pattern both bypass the Go-Jet type system,
   leaving the column lists silently coupled to schema state via reflection
   conventions that are not enforced anywhere.

Additionally, this story documents the role of the registry as a
**boundary helper** that legitimately reaches into both the generated
go-jet types and the EBM machinery — clarifying for future agents that
its impurity is intentional.

## References

- [database_design.md](specifications/architecture/database_design.md) — §3, §4, §7, §8
- [layered-architecture.md](specifications/patterns/layered-architecture.md)
- [service-organization.md](specifications/patterns/service-organization.md)
- [Story043_temporal-enforcement.md](finished-stories/Story043_temporal-enforcement.md) — Task 3 (Key Suspension and Reactivation)
- [lessons-learned.md](../lessons-learned.md) — "`ORDER BY` is mandatory with `LIMIT`" (added in Story 043G)

## Dependencies

- [Story043E_ebm_splinter_correctness.md](Story043E_ebm_splinter_correctness.md) — Not Started (EBM's mutation pipeline is rewritten in 043E; the column-list refactor in this story builds on the rewritten file)

## Dependent Stories

None.

## Tasks

1. **Make the bitemporal registry explicit about identity columns** — Done
   a. Changed `BitemporalRegistryEntry` (in `serde/rdbms/constants.go`) to add
      `IDColumns []postgres.Column` and `NonMetaColumns postgres.ColumnList` fields.
      - `IDColumns` is the ordered list of columns that uniquely identify an
        entity instance (single column for most tables; two columns for
        junction tables; three columns for localization tables).
      - `NonMetaColumns` is the precomputed column list excluding `IDColumns`
        and the bitemporal-meta columns (`validity_*`, `transaction_*`,
        `mode`, `insertion_metadata`). It replaces the runtime reflection
        in `getTableSpecificCols`. - Done
   b. Populated `IDColumns` and `NonMetaColumns` for every registered model
      type in `BitemporalTableRegistry` using a `registryEntry()` helper. - Done
   c. Added a doc comment to the top of `constants.go` explaining the
      registry's role as a **boundary helper**. - Done

2. **Replace `getIDColumnName` with registry lookups** — Done
   a. Deleted `getIDColumnName` from `mutation.go`. - Done
   b. Replaced every callsite (Phase 1, Phase 3, Phase 5, `ResolveEntityID`)
      with registry lookups returning the correct `IDColumns` slice. - Done
   c. Added regression test `TestEBM_CompositeKeyJunction` in `story043h_test.go`:
      seeds a `metric_dim` row pair (metric_id=90001, dim=1 and dim=2), mutates
      only dim=1 via EBM, and asserts that dim=2 is completely untouched
      (1 active row, 0 historical). - Done

3. **Rewrite `ResolveEntityID` to be two-pass with explicit ordering** — Done
   a. Changed the return signature to `(id int64, status ResolveStatus, err error)`
      where `ResolveStatus` is a typed enum: `ResolveActive`, `ResolveTombstoned`,
      or `ResolveNotFound`. - Done
   b. Pass 1: queries the active partition only, ORDER BY `validity_start DESC`,
      LIMIT 1. If a row is found, returns `(id, ResolveActive, nil)`. - Done
   c. Pass 2: queries the parent table filtered to `mode = 'historical'`,
      ORDER BY `transaction_end DESC, validity_end DESC`, LIMIT 1. If a
      row is found, returns `(id, ResolveTombstoned, nil)`. - Done
   d. If neither pass finds a row, returns `(0, ResolveNotFound, nil)`. - Done
   e. Tests written in `story043h_test.go`: - Done
      i. `TestResolveEntityID_ActiveHit` — passes.
      ii. `TestResolveEntityID_TombstonedOnly` — passes.
      iii. `TestResolveEntityID_NotFound` — passes.
      iv. `TestResolveEntityID_KeyReuse_PrefersActive` — passes.

4. **Replace `getTableSpecificCols` with the precomputed registry list** — Done
   a. Deleted `getTableSpecificCols` from `mutation.go`. - Done
   b. FLUSH statement reads column list from `entry.NonMetaColumns`. - Done
   c. Added startup assertion `validateRegistry()` in `mutation.go` that every
      registered entry has a non-empty `NonMetaColumns`. - Done

5. **Evaluate Go-Jet `INSERT … SELECT` for the FLUSH step** — Done
   Evaluated and documented in a comment in `mutation.go`. Conclusion: the Go-Jet
   form requires programmatic column-list construction and type assertions for the
   `transaction_end → CURRENT_TIMESTAMP` and `mode → 'historical'` projections,
   making it materially less readable than the raw SQL form. Raw SQL retained with
   a comment citing this story and the specific tradeoffs. - Done

6. **Tighten `overridingInsert` against future identifier collisions** — Done
   a. Added a FORBIDDEN IDENTIFIERS banner block near the top of `schema.sql`. - Done
   b. Changed `overridingInsert.Sql`'s pivot from `strings.LastIndex` to
      `valuesClauseRE.FindStringIndex` using
      `regexp.MustCompile(`(?i)\sVALUES\s*\(`)`. - Done
   c. Added `TestOverridingInsert_ValuesSubstringCollision` in `story043h_test.go`:
      synthesizes an INSERT whose key value contains the substring "values (",
      verifies "OVERRIDING SYSTEM VALUE" appears exactly once and before the
      actual `VALUES` keyword. - Done

7. **Reduce unexplained `time.Now()` calls in `mutation.go`** — Done
   Added one-line comments at every `time.Now()` call explaining why wall-clock
   time is correct (past-dated guard uses `time.Now()` because transaction time
   is unknown at validation time; explicit documentation deters future
   over-engineering). - Done

## Test-First Development

Tasks 2 (c), 3 (e), 5 (c — performance/readability evaluation has no
automated test, only documented decision), 6 (c) are explicit
test-before-implementation. The other tasks are mechanical refactors
covered by the existing EBM/read-path suites (after Stories 043E, 043F,
043G have run).

## Constraints

- **Composite-key tables must use both columns** in the staleness check,
  PURGE filter, and FLUSH `WHERE` clause. The registry change is a means;
  this end is the test.
- **`ResolveStatus` is a typed enum**, not an untyped constant or a `bool`
  pair. Callers must `switch` on it; do not let it degrade into "found /
  not found" before reactivation logic is implemented in 043D and later
  stories.
- **No new `time.Now()` calls** outside `mutation.go` and the past-dated
  guard.

## Intent

- Task 1 + 2 + 4 make the registry the single source of truth for
  per-table metadata, replacing today's brittle reflection +
  string-concatenation conventions.
- Task 3 fixes a real correctness bug (key-reuse → wrong entity ID) that
  would silently surface during the reactivation work in Story 043D and
  Story 045.
- Task 5 is the user's "use Jet if it adds type safety without
  obfuscation; otherwise document and keep raw SQL" guidance, written
  out as an explicit task with a documented outcome.
- Task 6 makes the identifier-collision risk explicit at the schema and
  enforces the contract programmatically.
- Task 7 closes the wall-clock determinism gap surfaced by the Code
  Reviewer audit.

## Acceptance Criteria

- [x] `BitemporalRegistryEntry` has `IDColumns` and `NonMetaColumns` fields,
  populated for every registered model.
- [x] `getIDColumnName` and `getTableSpecificCols` no longer exist.
- [x] `TestEBM_CompositeKeyJunction` passes (junction tables mutate
  correctly under EBM).
- [x] `ResolveEntityID` returns `(int64, ResolveStatus, error)`; all four
  tests in Task 3 (e) pass.
- [x] `schema.sql` contains the FORBIDDEN IDENTIFIERS banner.
- [x] `overridingInsert.Sql` uses a regex pivot; the synthesized-collision
  test passes.
- [x] Every `time.Now()` call in `mutation.go` is either commented or
  delegated to an injected clock.

## Branch

`claude/story-43-h-9VmSL`

## Status

**COMPLETED** — 2026-05-13
