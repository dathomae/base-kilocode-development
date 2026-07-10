# Plan: Story244 — Lookups Increment 2 (Parse)

## Purpose

This plan provides the technical writer with sufficient detail to create Story244: **lookups-increment2-parse** — extend the parsing phase to support the optional `grain` clause in lookup declarations (Increment 2 of Epic006).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `lookups-increment2-parse`                                                                   |
| description  | Extend grammar with optional grain clause, add Grain and HasGrain fields to ParserLookupDefinition, modify listener for grain extraction |

## Source EROs

- [ERO006: Lookups Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO006_lookups-parsing-phase.md) — Increment 2 grammar and data structure changes

## Dependencies

- Story241 (lookups-increment1-parse) — extends existing `CPLLookups.g4` grammar and listener

## Files to Create or Modify

| File                                                | Action | Purpose                                                                                    |
| --------------------------------------------------- | ------ | ------------------------------------------------------------------------------------------ |
| `src/main/antlr/CPLLookups.g4`                      | Modify | Add optional `grainClause?` to `lookupDeclaration`; add `grainClause` rule                  |
| `src/main/antlr/ChronoconePlanningLanguageLexer.g4`| Modify | Add `GRAIN` lexer token                                                                    |
| `src/main/go/parser/parser_lookup_definition.go`   | Modify | Add `Grain []string` and `HasGrain bool` fields                                             |
| `src/main/go/parser/LookupsListener.go`            | Modify | Update `ExitLookupDeclaration` to extract optional grain measure names                     |
| `src/main/go/parser/lookups_listener_test.go`      | Modify | Add tests for grain parsing (present, absent, multiple measures)                           |

### Grammar Changes

```antlr
// lookupDeclaration: add optional grainClause before celltypeClause
lookupDeclaration
    : indent identifier dimensionsClause grainClause? celltypeClause
    ;

// New rule
grainClause
    : indent GRAIN identifier (COMMA identifier)*
    ;
```

### Lexer Token

```antlr
GRAIN : 'grain';
```

### ParserLookupDefinition Changes (Increment 2 — fields added)

```go
type ParserLookupDefinition struct {
    Name           string
    Dimensions     []string
    Grain          []string     // NEW: raw measure names (nil if no grain clause)
    HasGrain       bool         // NEW: true if grain clause was present
    Celltype       string
    SourceLocation SourceLocation
}
```

### LookupsListener Modifications

In `ExitLookupDeclaration`, after extracting dimensions and before extracting celltype:
- Check `ctx.GrainClause() != nil`
- If present: set `HasGrain = true`, extract all identifiers from grain clause
- If absent: `HasGrain` remains false, `Grain` is nil
- Assemble `ParserLookupDefinition` with all fields

### Important Semantics

- `HasGrain` is distinct from `len(Grain) > 0`: `HasGrain` is true when the user wrote a `grain` clause, even with zero measures (which is a validation error caught in analysis)
- A lookup with no grain clause is timeless; `HasGrain = false`, `Grain = nil`
- Grain measure names are stored as raw strings; type validation (Interval-constraint) is deferred to analysis

## Task Decomposition (for technical writer)

### Task 1: Extend grammar with grain clause and add GRAIN token — Not Started [Small]
   a. Add `grainClause` rule to `CPLLookups.g4`: `indent GRAIN identifier (COMMA identifier)*` — Not Started
   b. Modify `lookupDeclaration` rule to include `grainClause?` between `dimensionsClause` and `celltypeClause` — Not Started
   c. Add `GRAIN` lexer token to `ChronoconePlanningLanguageLexer.g4` — Not Started
   d. Regenerate ANTLR Go files — Not Started
   e. Write grammar-level tests: valid grain (single measure), valid grain (multiple measures), no grain clause, empty grain clause — Not Started
   f. Verify generated parser compiles — Not Started

### Task 2: Extend ParserLookupDefinition with grain fields — Not Started [Small]
   a. Add `Grain []string` and `HasGrain bool` fields to `ParserLookupDefinition` — Not Started
   b. Update existing unit tests to verify zero values for grain fields when no grain clause — Not Started
   c. Verify package compiles — Not Started

### Task 3: Update LookupsListener for grain extraction — Not Started
   a. Modify `ExitLookupDeclaration` to check for optional `GrainClause` — Not Started
   b. Extract grain measure names from `GrainClauseContext.AllIdentifier()` when present — Not Started
   c. Set `HasGrain = true` when grain clause is present, even if no measures listed — Not Started
   d. Write listener unit tests: grain clause present (single measure), grain clause present (multiple measures), no grain clause (verify HasGrain=false, Grain=nil), empty grain clause (HasGrain=true, Grain=[]) — Not Started
   e. Verify listener tests pass — Not Started

### Task 4: Create parse_and_analyze_only examples with grain — Not Started
   a. In `parse_and_analyze_only/without_errors/`, add CPL files with temporal lookups: single grain, multiple grains — Not Started
   b. In `parse_and_analyze_only/with_errors/`, add CPL files: empty grain clause (zero measures), non-existent grain measure (no parser error, but included for analysis phase) — Not Started
   c. Verify all examples parse correctly — Not Started

## Execution Order

```
Task 1 → Task 3 → Task 4
            ↗
Task 2 ──────┘
```

- Task 1 must complete first (grammar rules define what the listener extracts)
- Task 2 (struct fields) and Task 1 can overlap, but Task 2 can run independently
- Task 3 depends on Tasks 1 and 2 (needs grammar rules AND struct fields)
- Task 4 depends on all preceding tasks

## Parallel Execution

- **Group A** (Tasks 1, 2): Task 1 modifies grammar and lexer files; Task 2 modifies `parser_lookup_definition.go`. Disjoint files, different concerns. Worktree isolation required.

## Reintegration

**Group A (Tasks 1, 2):**
- Merge order: Task 1 first (grammar generation may produce new types that Task 2's struct references)
- Integration tests: `go test ./src/main/go/parser/... -count=1 -timeout 60s` — verify grammar and structs compile together
- Watch for conflicts in: none expected (disjoint files)

## Acceptance Criteria Summary

1. `CPLLookups.g4` grammar accepts optional `grain Clause` with one or more comma-separated measure names
2. `GRAIN` token recognized by the lexer
3. `ParserLookupDefinition.Grain` populated with raw measure names when grain clause is present
4. `ParserLookupDefinition.HasGrain` is true when grain clause is present (including empty), false when absent
5. Lookups without grain clauses continue to parse correctly (backward compatibility)
6. No grain type validation at parser level (deferred to analysis)
7. `parse_and_analyze_only/` examples with grain parse correctly

## Intent Summary for Technical Writer

This story extends the parsing phase to support the optional `grain` clause in lookup declarations. The grammar is modified to accept an optional `grainClause` between `dimensionsClause` and `celltypeClause`. The listener extracts grain measure names as raw strings; the analysis phase (Story245) validates that grain measures exist and have Interval-derived types. The `HasGrain` boolean flag distinguishes "grain not present" (timeless lookup) from "grain present" (temporal lookup), even when the grain clause contains zero measures (an error caught in analysis).
