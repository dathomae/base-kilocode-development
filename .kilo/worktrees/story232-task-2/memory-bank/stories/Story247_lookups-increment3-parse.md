# Story247: Lookups Increment 3 — Parse

## Goal

Extend the metrics grammar to support the `lookup` input kind in metric `inputs` blocks. Add a `LOOKUP` lexer token (singular, distinct from `LOOKUPS` added in Story241), modify the `inputDeclaration` rule in `CPLMetrics.g4`, and extend the `MetricsListener` to handle the new alternative. The lookup name is stored as a raw string in `ParserInputDefinition` — cross-reference validation is deferred to the analysis phase (Story248). This is Increment 3 of [Epic006: Lookups Implementation](../epics/Epic006_cpl-lookups-implementation.md).

## References

- [Epic006: Lookups Implementation](../epics/Epic006_cpl-lookups-implementation.md)
- [ERO006: Lookups Parsing Phase — Increment 3](../specifications/design/cpl/go-parser-target/eros/ERO006_lookups-parsing-phase.md)
- [CPLMetrics.g4 grammar](../../src/main/antlr/CPLMetrics.g4)
- [ChronoconePlanningLanguageLexer.g4](../../src/main/antlr/ChronoconePlanningLanguageLexer.g4)
- [MetricsListener.go](../../src/main/go/parser/MetricsListener.go)

## Dependencies

- Story241 (lookups-increment1-parse) — _Not Started_ — basic lookup parsing must exist (`LOOKUPS` section keyword in `CPLLookups.g4`)
- Epic005 (metrics) — _Partially Complete_ — `CPLMetrics.g4` grammar and `ParserInputDefinition` struct must exist for modification

## Dependent Stories

- Story248 (lookups-increment3-analyze) — _Not Started_ — cross-reference validation of lookup names requires the parser to recognize `lookup` input kind

## Tasks

### Task Sizing

No tasks qualify for `[Small]` annotation:

- **Task 1**: Touches 3 files (CPLMetrics.g4, ChronoconePlanningLanguageLexer.g4, test file) — exceeds the 2-file threshold. While the changes are mechanical additions to existing grammar rules following established patterns, the 3-file span rules out `[Small]`.
- **Task 2**: Touches 2–3 files (MetricsListener.go, metrics_listener_test.go, and possibly metric_parser_helpers.go if `InputKind` requires an enum extension). The enum ambiguity pushes this past the straightforwardness test — a limited-context agent would need to determine whether `InputKind` is a string or enum and reason about the correct modification. Route to `code-for-story-implementor`.
- **Task 3**: Creates example CPL files (data artifacts). Not a traditional coding task — involves writing test-data files and verifying parser behavior across multiple scenarios. The example creation requires understanding the full CPL metrics syntax; route to `code-for-story-implementor`.
- **Task 4**: Verification-only — runs existing test suites to confirm zero regressions. No source code changes. Not applicable for `[Small]` annotation; route to `code-for-story-implementor`.

### Parallel Execution

None — all tasks have sequential dependencies. Task 2's listener changes require Task 1's grammar modifications to compile. Task 3's examples exercise both grammar and listener. Task 4's regression suite covers all prior work.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

- Task 1 must complete before Task 2 (listener compiles against generated parser code from updated grammar).
- Task 2 must complete before Task 3 (examples exercise the full grammar + listener pipeline).
- Tasks 1–3 must complete before Task 4 (regression suite covers all changes).

### Reintegration

None — no parallel groups. All work flows through a single sequential worktree chain; no worktree merges are required within this story.

---

### Task 1: Extend CPLMetrics.g4 grammar with lookup input kind — Not Started

   a. Add `LOOKUP` alternative to `inputDeclaration` rule in `src/main/antlr/CPLMetrics.g4` — the rule should accept `| LOOKUP identifier` alongside existing `METRIC identifier` and `DIMENSION identifier` alternatives — Not Started
   b. Add `LOOKUP` lexer token to `src/main/antlr/ChronoconePlanningLanguageLexer.g4` — define `LOOKUP : 'lookup';` (singular, distinct from the `LOOKUPS` token added in Story241) — Not Started
   c. Regenerate ANTLR Go files from the updated grammars — Not Started
   d. Write grammar-level test: verify that `lookup <name>` in a metrics `inputs` block parses successfully — Not Started
   e. Verify generated parser compiles: `go build ./src/main/go/parsing/...` — Not Started

### Task 2: Extend MetricsListener for lookup input kind — Not Started

   a. Add `ctx.LOOKUP()` branch in the input declaration handler in `src/main/go/parser/MetricsListener.go` — the branch should follow the same pattern as the existing `METRIC` and `DIMENSION` branches — Not Started
   b. Set `InputKind = "lookup"` and `Name` from the identifier text in the new branch — Not Started
   c. If `InputKind` is an enum type in `src/main/go/parser/metric_parser_helpers.go`, add `InputKindLookup` enum value; if `InputKind` is a string field, no change to helpers is needed — Not Started
   d. Write listener unit tests in `src/main/go/parser/metrics_listener_test.go` covering: single `lookup` input, `lookup` alongside `metric` and `dimension` inputs in the same `inputs` block, and multiple `lookup` inputs — Not Started
   e. Verify listener tests pass: `go test ./src/main/go/parser/... -count=1 -run TestMetrics` — Not Started

### Task 3: Create parse_and_analyze_only examples with lookup inputs — Not Started

   a. Create CPL files in the `parse_and_analyze_only` examples directory that define metrics with `lookup` input kind references — Not Started
   b. Verify parse round-trip: the metrics grammar accepts `lookup` inputs alongside `metric` and `dimension` inputs in the same `inputs` block — Not Started
   c. Verify error cases: `lookup` without an identifier should be rejected by the grammar; invalid characters in the lookup name should be rejected — Not Started

### Task 4: Verify backward compatibility — Not Started

   a. Verify existing metrics with only `metric` and `dimension` inputs still parse correctly — Not Started
   b. Verify existing examples with `inputs` blocks are unaffected by the grammar changes — Not Started
   c. Run the full parse-phase test suite: `go test ./src/main/go/parsing/... -count=1 -timeout 120s` — Not Started

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks (Tasks 1 and 2), write tests before implementing the code. The test-and-implement cycle is integrated within each task in the correct order:

- **Task 1**: Write the grammar-level test (subtask d) before or alongside the grammar changes (subtasks a–b). The grammar test confirms parsing behavior independently of the listener.
- **Task 2**: Write the listener unit tests (subtask d) before the listener implementation (subtasks a–c). The tests define expected behavior first.

See `memory-bank/concepts.md` for the Logical TDD Lifecycle.

## Constraints

- The `LOOKUP` token (singular, for `'lookup'`) must NOT conflict with the `LOOKUPS` token (plural, for `'lookups'`) defined in Story241. These are separate lexer tokens with distinct string values.
- The `lookup` input kind stores only the raw lookup name string — no cross-reference validation (does the lookup exist? are dimensions compatible?) is performed at the parser level. Cross-reference validation is deferred to the analysis phase (Story248).
- No changes to `CPLLookups.g4` or `LookupsListener.go` are needed for this increment. Only the metrics grammar and listener are modified.
- Existing `metric` and `dimension` input kinds must continue to work without regression.
- All existing metrics parse tests must pass after these changes.

## Intent

- **Task 1**: Extend the ANTLR grammar so the lexer recognizes `lookup` as a keyword token and the `inputDeclaration` parser rule accepts `lookup <name>` as a valid alternative. This is the minimal grammar surface change needed for the `lookup` input kind.
- **Task 2**: Extend the `MetricsListener` to handle the new `LOOKUP` parse-tree alternative when walking `inputDeclaration` nodes. Store the lookup name in the existing `ParserInputDefinition` structure using the existing `InputKind`/`Name` fields. This keeps the listener consistent with the grammar.
- **Task 3**: Exercise the full grammar + listener pipeline with realistic CPL files containing `lookup` inputs, verifying both happy-path parsing and error rejection.
- **Task 4**: Confirm that the grammar and listener changes introduce zero regressions in the existing metrics parsing behavior.

## Acceptance Criteria

1. `CPLMetrics.g4` grammar accepts `lookup <name>` as a valid input kind in metric `inputs` blocks.
2. `LOOKUP` token (singular, string value `'lookup'`) is recognized by the lexer and is distinct from the `LOOKUPS` token (string value `'lookups'`).
3. `ParserInputDefinition.InputKind` can be `"lookup"` with the lookup name stored in `Name`.
4. Existing `metric` and `dimension` input kinds continue to work unaffected.
5. No cross-reference validation at parser level (lookup name existence is checked in analysis, Story248).
6. All existing metrics parse tests pass with zero regressions.
