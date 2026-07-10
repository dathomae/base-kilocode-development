# Story 043I: Documentation, Schema Header, and Scaffold Cleanup

## Goal

Close out the 043-series corrective work by:

1. Rewriting the `schema.sql` header so it no longer advertises the
   deprecated "Close and Insert" pattern as canonical.
2. Reconciling minor wording inconsistencies among `database_design.md`,
   `layered-architecture.md`, and `service-organization.md` (canonical
   naming, scope of "current belief", staleness predicate wording,
   `domain/` tree allowance, JSON-tag-to-column-alignment guidance).
3. Re-opening Story 043B and recording its actual state per
   lessons-learned "Truth in Documentation".
4. Removing the `EnableLockSafetyChecks` constant — it was correctly
   identified by the user as a correctness invariant, not a tunable, so
   the safety check should always run and the toggle should not exist.
5. Cleaning up domain and security scaffolding (idiomatic field names,
   removal of unused empty structs, story-pointer comments for stub
   files).
6. Replacing `cmd/dbtool/main.go`'s runtime path-walking schema
   discovery with `//go:embed`, with documented reasoning so the
   bootstrap-cycle rationale survives.
7. Removing `src/backend/tools.go` (a one-line agent shim the user
   confirmed should be deleted).

This is the last story in the 043 corrective series. Once it lands, the
"Close and Insert" pattern has no remaining citations in the codebase or
the memory bank.

## References

- [database_design.md](specifications/architecture/database_design.md)
- [layered-architecture.md](specifications/patterns/layered-architecture.md)
- [service-organization.md](specifications/patterns/service-organization.md)
- [Story043B_active_historical_partitioning.md](Story043B_active_historical_partitioning.md)
- [Story043C_purge_store_abstraction.md](finished-stories/Story043C_purge_store_abstraction.md)
- [lessons-learned.md](../lessons-learned.md) — "Truth in Documentation: Synchronize Story State with Verified Reality"

## Dependencies

- [Story043E_ebm_splinter_correctness.md](Story043E_ebm_splinter_correctness.md) — Not Started (this story's wording reconciliation in `database_design.md` §7 builds on the rewrite 043E performs)
- [Story043F_read_path_correctness.md](Story043F_read_path_correctness.md) — Not Started (this story may further tighten §8 wording once §8's signatures are known to match the code)
- [Story043H_registry_and_mutation_hardening.md](Story043H_registry_and_mutation_hardening.md) — Not Started (this story finalizes the "EBM" canonical naming everywhere)

## Dependent Stories

None.

## Tasks

1. **Rewrite the `schema.sql` header** — Not Started
   a. Replace lines 1–14 of `src/backend/serde/rdbms/schema.sql` with:
      ```sql
      /*
        CHRONOCONE CONSOLIDATED SCHEMA

        ARCHITECTURAL REFERENCE:
        This schema is bitemporal. Mutations flow exclusively through
        ExecuteBitemporalMutation (EBM) in serde/rdbms/mutation.go,
        which implements the Flush/Purge/Assert pipeline and owns all
        interval algebra. Application code outside that file may emit
        only INSERT statements against bitemporal tables.

        Full documentation:
          memory-bank/specifications/architecture/database_design.md

        QUICK RULES:
          1. No UPDATEs or DELETEs against bitemporal tables outside
             serde/rdbms/mutation.go.
          2. Identity: override identity columns to reuse stable IDs
             across versions (OVERRIDING SYSTEM VALUE).
          3. Localization: human-facing strings reside in *_localization
             tables.
          4. Keys: technical keys (CITEXT) are 6-64 chars and stable
             across all versions of an entity.
          5. Audit: every row must include insertion_metadata
             (AuditContext).

        FORBIDDEN IDENTIFIERS (do not name a table, column, or alias
        these):
          values   -- collides with overridingInsert's SQL pivot
                      (see serde/rdbms/mutation.go)
      */
      ```
      The "FORBIDDEN IDENTIFIERS" banner is shared with Story 043H; if
      043H landed first, the banner already exists and this task only
      replaces the upper portion. - Not Started
   b. Grep the repository to confirm no other `"Close and Insert"` or
      `"Close-and-Insert"` references survive in any `*.go`, `*.sql`, or
      `*.md` file: - Not Started
      ```sh
      grep -rIE "(Close[- ]and[- ]Insert|Close,? and Insert)" \
        src/backend memory-bank
      ```
      Expected: zero hits.

2. **Allow a `domain/` tree in pattern documents** — Not Started
   a. In `service-organization.md` §"Directory Structure", expand the
      example tree to show that `domain/` may itself contain sub-packages
      (e.g., `domain/entities/`, `domain/timeseries/`) when grouping
      improves cohesion. - Not Started
   b. Add a one-sentence rule:
      "`domain/` MAY be subdivided into sub-packages by entity grouping.
      Flat is preferred when the entity count is small; group when a flat
      list would exceed roughly a dozen files or when cohesive entity
      families benefit from a shared internal vocabulary." - Not Started
   c. Apply the same expansion to `layered-architecture.md` §"Architecture
      Overview" if its directory description constrains layout. - Not Started

3. **Soften the JSON-tag-to-SQL-column alignment rule** — Not Started
   a. In `service-organization.md` §"API Serialization and JSON Tags",
      change rule 1 from "Use `snake_case` for all JSON tags **to
      maintain consistency with SQL column names**" to:
      "Use `snake_case` for all JSON tags. A 1:1 correspondence between
      JSON tag and SQL column name is the **default bias** because it
      minimizes mental mapping, but the chosen technology stack (Go-Jet
      reflection driven by `sql` tags, plus separate `json` tags)
      explicitly supports deviating when the API contract has a
      legitimate reason to differ from the storage column name." - Not Started

4. **Reconcile EBM canonical naming everywhere** — Not Started
   a. Pick one canonical form: **`ExecuteBitemporalMutation` (EBM)** on
      first reference per document, **EBM** thereafter. - Not Started
   b. In every "how" document
      (`database_design.md`, `layered-architecture.md`,
      `service-organization.md`), confirm the first reference uses the
      full name with `(EBM)` and subsequent references use `EBM`. Rewrite
      stragglers like "the Mutation Orchestrator", "Bitemporal Mutation
      Orchestrator", or "MutationOrchestrator" to match the canonical
      form. - Not Started

5. **Tighten "current belief" wording in `database_design.md` §3** — Not Started
   a. Replace "only the current belief exists" with "only rows with
      `transaction_end = 'infinity'` exist". The current wording is
      imprecise because the active partition holds multiple validity
      segments per entity (preserved-prefix splinters), which is fine
      and intentional. - Not Started

6. **Tighten staleness predicate wording in `service-organization.md`** — Not Started
   a. In §"Bitemporal Mutation Protocol" step 2, replace "If they differ,
      the record was modified by another process" with "If the DB's max
      `transaction_start` is **greater than** the caller's known value,
      the record was modified by another process". (This duplicates work
      already done in Story 043E; if 043E landed first, this task is a
      no-op verification step.) - Not Started

7. **Cross-reference `database_design.md` §7 ↔ `layered-architecture.md` §2** — Not Started
   a. Add a one-line bridge in `database_design.md` §7: "Executed under
      advisory lock after staleness check; see `layered-architecture.md`
      §2 for the full orchestrator lifecycle." - Not Started
   b. Add the reverse bridge in `layered-architecture.md` §2: "The Flush,
      Purge, and Assert sub-pipeline is specified in
      `database_design.md` §7." - Not Started

8. **Re-open Story 043B and align its status** — Not Started
   a. Move `Story043B_active_historical_partitioning.md` from
      `memory-bank/stories/` if necessary (it currently sits at the
      stories root). - Not Started
   b. Change the status block from `COMPLETED ✅` to `IN PROGRESS —
      pending 043E and 043F to satisfy AC #3 (multi-interval splinter)
      and AC #6 (1D read helpers return flat 1D arrays)`. - Not Started
   c. Uncheck the affected acceptance criteria (AC #3 and AC #6). - Not Started
   d. Once 043E and 043F are merged to main and their tests pass, the
      next work-planner pass should re-check the criteria and either
      re-close 043B or split out the still-failing items into a new
      sub-story. - Not Started

9. **Remove `EnableLockSafetyChecks`** — Not Started
   a. Delete the `const EnableLockSafetyChecks = true` from
      `serde/rdbms/constants.go`. - Not Started
   b. Delete the `if EnableLockSafetyChecks { ... }` guard in
      `serde/rdbms/locks.go`; the safety check becomes unconditional. - Not Started
   c. Remove the `require.True(t, EnableLockSafetyChecks)` line that
      Story 043G added to `locks_test.go` (no longer applicable). - Not Started
   d. Update any documentation that references the const. - Not Started

10. **Idiomatic field names and removal of unused empty structs in `domain/localization.go`** — Not Started
    a. Rename `cur_name`, `cur_desc`, `cur_tooltip`, `cur_index` to
       `curName`, `curDesc`, `curTooltip`, `curIndex`. - Not Started
    b. Delete the `NumberFormat` and `DateTimeFormat` empty structs and
       their commented intent — they will be re-introduced when their
       requirements are real (likely Story 044 or 100). - Not Started
    c. Keep the entity stubs (`Category`, `Element`, `Hierarchy`,
       `Morphism`, etc.) in place per user direction; they are
       scaffolding for future stories. - Not Started

11. **Add story-pointer comments to `security/*` stubs** — Not Started
    a. Add a one-line file comment at the top of each empty `security/`
       file pointing to the story that will fill it in: - Not Started
       - `security/authentication.go`: `// Stub. Implementation pending Story 025 (identity-authentication-authorization-requirements) and Story 045 (policy-services).`
       - `security/authorization.go`: `// Stub. Implementation pending Story 025 and Story 045.`
       - `security/workflow.go`: `// Stub. Implementation pending Story 037 (agents-and-agentic-interaction-requirements) and Story 045.`
       - `security/connection/blob.go`: `// Stub. Implementation pending Story 050 (connector-services).`
       - `security/connection/database.go`: `// Stub. Implementation pending Story 050.`
    b. Add the same kind of one-line comment to each `domain/entities/*.go`
       file that currently has empty struct fields, e.g., `// Field shape
       pending Story 046 (taxonomy-services).` for `category.go` and
       `element.go`. - Not Started

12. **Replace `cmd/dbtool/main.go`'s schema discovery with `//go:embed`** — Not Started
    a. Add a sibling file `cmd/dbtool/schema_embed.go`:
       ```go
       package main

       import _ "embed"

       // schemaSQL is embedded at compile time. dbtool must NOT import the
       // serde/rdbms package because dbtool runs *before* go-jet has
       // generated the model files in serde/rdbms/gen/. Embedding the
       // schema file directly (rather than walking up the filesystem)
       // makes the dependency explicit to the Go toolchain and survives
       // packaging into a single binary.
       //
       //go:embed all:../../serde/rdbms/schema.sql
       var schemaSQL string
       ```
       (Adjust the embed path to whatever Go's `embed` directive accepts
       for traversing up two directories; if Go forbids `..` in embed
       paths, instead symlink `schema.sql` into `cmd/dbtool/` or copy as
       part of the generate step. Document whichever approach is used.) - Not Started
    b. Replace `applySchemaInternal`'s `os.ReadFile` + `runtime.Caller`
       logic with `conn.Exec(ctx, schemaSQL)`. - Not Started
    c. Verify the dbtool binary still applies the schema correctly. - Not Started

13. **Delete `src/backend/tools.go`** — Not Started
    a. Remove the file outright. - Not Started
    b. Verify `go build ./...` still succeeds. - Not Started

## Test-First Development

This story is predominantly documentation and small mechanical refactors;
test-first is only practical for Task 12 (`go:embed` swap). For that
task, the test is the existing `TestSchemaExecution` in
`serde/rdbms/schema_test.go` plus a smoke run of
`go run ./cmd/dbtool --port 5439 --apply-schema --keep-alive=false`
asserting non-zero exit on failure.

## Constraints

- **No "Close and Insert" references anywhere** in `src/backend/` or
  `memory-bank/` after this story.
- **Story 043B re-open is normative, not optional.** Future agents
  reading the memory bank must see 043B's true state.
- **Embed path must work in both `go test` and `go run` contexts**, and
  must not depend on the developer's current working directory.

## Intent

- Tasks 1, 4, 5, 6, 7: complete the documentation cleanup begun in
  Stories 043C and 043E.
- Tasks 2, 3: codify the user's preference for tree-shaped `domain/`
  layouts and a softer JSON-to-column alignment rule.
- Task 8: enforce the truth-in-documentation lesson.
- Task 9: remove a "tunable" that should never have been one.
- Tasks 10, 11: clean up scaffolding so future agents read intent, not
  empty structs.
- Task 12: replace fragile runtime path discovery with a compile-time
  guarantee, with the reasoning preserved next to the code.
- Task 13: remove the agent shim with no replacement needed.

## Acceptance Criteria

- [ ] `grep -rIE "(Close[- ]and[- ]Insert|Close,? and Insert)" src/backend
  memory-bank` returns zero hits.
- [ ] `grep -nE "(the )?Mutation Orchestrator|MutationOrchestrator"
  memory-bank/specifications/` returns zero hits (all replaced by
  "EBM" or "ExecuteBitemporalMutation (EBM)").
- [ ] `Story043B_active_historical_partitioning.md` status reads
  `IN PROGRESS` and acceptance criteria #3 and #6 are unchecked.
- [ ] `EnableLockSafetyChecks` does not appear in `src/backend/`.
- [ ] `domain/localization.go` uses idiomatic Go field names; `NumberFormat`
  and `DateTimeFormat` are deleted.
- [ ] Every `security/*.go` and `domain/entities/*.go` stub file has a
  one-line comment pointing to the story that will fill it in.
- [ ] `src/backend/tools.go` does not exist.
- [ ] `cmd/dbtool/main.go` uses `//go:embed` (or the documented
  alternative) to load `schema.sql`; the `runtime.Caller` /
  `filepath.Join` discovery code is gone.
- [ ] `make test` from `src/backend/` passes.

## Branch

`claude/story043I-docs-and-scaffold-cleanup`
