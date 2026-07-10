# Plan: Story247 — Lookups Increment 3 (Parse)

## Purpose

This plan provides the technical writer with sufficient detail to create Story247: **lookups-increment3-parse** — extend metrics grammar to support the `lookup` input kind in metric `inputs` blocks (Increment 3 of Epic006).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `lookups-increment3-parse`                                                                   |
| description  | Extend metrics grammar with lookup input kind, add LOOKUP token, modify inputDeclaration rule and MetricsListener |

## Source EROs

- [ERO006: Lookups Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO006_lookups-parsing-phase.md) — Increment 3 metrics grammar changes

## Dependencies

- Story241 (lookups-increment1-parse) — basic lookup parsing must exist
- Epic005 (metrics) — `CPLMetrics.g4` grammar must exist for modification; `ParserInputDefinition` struct must exist
- Note: The `LOOKUP` token (singular) is distinct from the `LOOKUPS` token (plural, section keyword) added in Story241

## Files to Create or Modify

| File                                                | Action | Purpose                                                                                        |
| --------------------------------------------------- | ------ | ---------------------------------------------------------------------------------------------- |
| `src/main/antlr/CPLMetrics.g4`                      | Modify | Add `LOOKUP` alternative to `inputDeclaration` rule                                            |
| `src/main/antlr/ChronoconePlanningLanguageLexer.g4`| Modify | Add `LOOKUP` lexer token (singular; distinct from `LOOKUPS` added in Story241)                 |
| `src/main/go/parser/MetricsListener.go`            | Modify | Extend `ExitInputDeclaration` (or equivalent) to handle `LOOKUP` alternative                   |
| `src/main/go/parser/metrics_listener_test.go`      | Modify | Add tests for lookup input kind parsing                                                        |
| `src/main/go/parser/metric_parser_helpers.go`      | Modify | `ParserInputDefinition` gains `"lookup"` as a valid `InputKind` value (if `InputKind` is a string; if it's an enum, add the enum value) |

**Note:** No changes to `CPLLookups.g4` or `LookupsListener.go` are needed for this increment. Only the metrics grammar and its listener are modified.

### Grammar Changes — CPLMetrics.g4

```antlr
// Before (Increment 1-2 of metrics):
inputDeclaration
    : METRIC identifier
    | DIMENSION identifier
    ;

// After (Increment 3):
inputDeclaration
    : METRIC identifier
    | DIMENSION identifier
    | LOOKUP identifier      // NEW
    ;
```

### Lexer Token

```antlr
LOOKUP : 'lookup';    // Singular, distinct from LOOKUPS : 'lookups' (added in Story241)
```

### MetricsListener Changes

In the handler for `inputDeclaration` (likely `ExitInputDeclaration` or `ExitInputRef`):
- Add `else if ctx.LOOKUP() != nil` branch
- Set `InputKind = "lookup"`
- Set `Name = ctx.Identifier(0).GetText()`

### ParserInputDefinition Extension

If `InputKind` is a string field, the new value `"lookup"` is naturally accommodated. If it's an enum, add `InputKindLookup` to the enum.
- No new struct fields needed — the existing `InputKind` and `Name` fields suffice.

## Task Decomposition (for technical writer)

### Task 1: Extend CPLMetrics.g4 grammar with lookup input kind — Not Started [Small]
   a. Add `LOOKUP` alternative to `inputDeclaration` rule in `CPLMetrics.g4` (`| LOOKUP identifier`) — Not Started
   b. Add `LOOKUP` lexer token to `ChronoconePlanningLanguageLexer.g4` — Not Started
   c. Regenerate ANTLR Go files — Not Started
   d. Write grammar-level test: `lookup <name>` in a metrics `inputs` block parses — Not Started
   e. Verify generated parser compiles — Not Started

### Task 2: Extend MetricsListener for lookup input kind — Not Started [Small]
   a. Add `ctx.LOOKUP()` branch in the input declaration handler — Not Started
   b. Set `InputKind = "lookup"` and `Name` from identifier — Not Started
   c. If `InputKind` is an enum, add `InputKindLookup` value — Not Started
   d. Write listener unit tests: single `lookup` input, `lookup` alongside `metric` and `dimension` inputs, multiple `lookup` inputs — Not Started
   e. Verify listener tests pass — Not Started

### Task 3: Create parse_and_analyze_only examples with lookup inputs — Not Started
   a. Create CPL files with metrics that reference lookups via `lookup` input kind — Not Started
   b. Verify parse round-trip: metrics grammar accepts `lookup` inputs alongside `metric` and `dimension` inputs — Not Started
   c. Verify error cases: `lookup` without identifier (grammar should reject), invalid characters in lookup name — Not Started

### Task 4: Verify backward compatibility — Not Started [Small]
   a. Verify existing metrics with only `metric` and `dimension` inputs still parse correctly — Not Started
   b. Verify existing examples with `inputs` blocks are unaffected — Not Started
   c. Run full parse-phase test suite — Not Started

## Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

All tasks are sequential — each builds on the previous.

## Parallel Execution

None — all tasks have sequential dependencies.

## Reintegration

None — no parallel groups.

## Acceptance Criteria Summary

1. `CPLMetrics.g4` grammar accepts `lookup <name>` as a valid input kind in metric `inputs` blocks
2. `LOOKUP` token (singular) is recognized by the lexer
3. `ParserInputDefinition.InputKind` can be `"lookup"` with the lookup name in `Name`
4. Existing `metric` and `dimension` input kinds continue to work unaffected
5. No cross-reference validation at parser level (lookup name existence checked in analysis)
6. All existing metrics parse tests pass with zero regressions

## Intent Summary for Technical Writer

This story extends the existing metrics grammar to recognize `lookup` as a new input kind in metric `inputs` blocks. The change is minimal: a new alternative in the `inputDeclaration` rule, a new `LOOKUP` lexer token, and an additional branch in the `MetricsListener`. The lookup name is stored as a raw string in `ParserInputDefinition` — cross-reference validation (does the lookup exist? are dimensions compatible?) is deferred to the analysis phase (Story248). No changes to the lookup grammar or listener are needed; only the metrics side is modified.
