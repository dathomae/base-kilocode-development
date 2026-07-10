# Story191: CPL Metrics Time Shifting — Parse and Analyze

## Goal

Implement the Increment 6 parsing and analysis phases for CPL metric time shifting (backward/forward) as part of [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md). This story covers:

- Extend `inputDeclaration` grammar with `backward()`/`forward()` clauses accepting Instant constraint field keywords (`year`, `month`, `day`, `hour`, `minute`, `second`, `day_of_week`, `week_number`) with shift expressions.
- Store `ParserTimeShift` in `ParserInputDefinition`.
- Extend analysis to validate time-shift constraint field references, evaluate shift amount expressions, and handle self-references (metric referencing itself via backward/forward alias creates a time-series dependency, not a cycle). Detect prohibited direct self-references without time shift.

## References

- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [Metrics Syntax Requirements](../specifications/requirements/cpl/metrics-syntax.md)
- [CPL Architecture — Talking About Time](../specifications/requirements/cpl/talking-about-time.md)
- [ERO005: Metrics Parsing Phase — Increment 6](../specifications/design/cpl/go-parser-target/eros/ERO005_metrics-parsing-phase.md)
- [ERO005: Metrics Analysis Phase — Increment 6](../specifications/design/cpl/go-analysis/eros/ERO005_metrics-analysis-phase.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story184_cpl-metrics-arithmetic-parse-analyze.md](finished-stories/Story184_cpl-metrics-arithmetic-parse-analyze.md) — _Done_ — arithmetic parse/analyze infrastructure must be in place
- [Story189_cpl-metrics-aliases-where-parse-analyze.md](finished-stories/Story189_cpl-metrics-aliases-where-parse-analyze.md) — _Done_ — `ParserInputDefinition` and `AnalysisInputDefinition` with alias and where clause support must exist

## Dependent Stories

- [Story192_cpl-metrics-time-shift-synthesize.md](Story192_cpl-metrics-time-shift-synthesize.md) — _Not Started_ — Increment 6 Synthesis

## Tasks

### Task Sizing

- **Task 1** (Extend input declaration grammar): `[Small]` — single grammar file change (`CPLMetrics.g4`) plus ANTLR regeneration. Mechanical transformation following established grammar patterns from Increments 1–5.
- **Task 2** (Parser data structures): `[Small]` — new file (`parser_time_shift.go`) plus single field addition to existing struct (`parser_input_definition.go`). At most 2 files, purely additive, compiles independently.
- **Task 3** (Parser listener methods): `[Small]` — adds `ExitBackwardClause`/`ExitForwardClause` to existing listener file. Follows established listener method patterns for other clause types. 1 file.
- **Task 4** (Analysis data structures): `[Small]` — extends `AnalysisInputDefinition` with `TimeShift` field (1 existing file). Purely additive, compiles independently.
- **Task 5** (Time shift validation logic): Medium — multiple coordinated validations across `metrics_analysis.go` and related analysis files. Requires reasoning about self-reference semantics (time-shifted = dependency, not cycle; no time shift = true cycle error). Touches 2–3 files.
- **Task 6** (Integration verification): Medium — creates example CPL files, verifies full pipeline, tests edge cases, runs regression. 4+ subtasks.

### Parallel Execution

- **Group A** (Tasks 1, 2, 4): All three tasks are file-disjoint and introduce purely additive changes:
  - Task 1: `src/main/antlr/CPLMetrics.g4` (grammar file) — not Go source
  - Task 2: `src/main/go/parsing/parser_time_shift.go` (new) + `src/main/go/parsing/parser_input_definition.go` (modify) — parsing package
  - Task 4: `src/main/go/analysis/analysis_input_definition.go` (modify) — analysis package
  - Different packages, no compile-time coupling, no shared files.

### Execution Order

```
[Task 1 ‖ Task 2 ‖ Task 4] → Task 3 → Task 5 → Task 6
```

- Tasks 1, 2, 4 run in parallel (file-disjoint, no dependencies).
- Task 3 depends on Task 1 (listener methods reference new grammar rules) and Task 2 (listener creates `ParserTimeShift` instances).
- Task 5 depends on Task 4 (analysis structs) and Task 3 (parser produces time shift data that analysis translates and validates). By the time Task 3 completes, Tasks 1, 2, and 4 are already done.
- Task 6 depends on all preceding tasks (full pipeline verification).

### Reintegration

**Group A (Tasks 1, 2, 4):**
- Merge order: any order (fully disjoint files: grammar file, parsing package, analysis package)
- Integration tests: `go build ./src/main/go/parsing/... ./src/main/go/analysis/...` — verify parser and analyzer packages compile together after merge
- Watch for conflicts in: none expected (fully disjoint file sets across different packages)

### Task 1: Extend input declaration grammar with backward/forward time shift clauses — Completed
   a. Add `backwardClause` rule: `BACKWARD_KEYWORD LPAREN shiftExpression RPAREN` — Not Started
   b. Add `forwardClause` rule: `FORWARD_KEYWORD LPAREN shiftExpression RPAREN` — Not Started
   c. Add `shiftExpression` rule: Instant constraint field keyword (`year`, `month`, `day`, `hour`, `minute`, `second`, `day_of_week`, `week_number`) with optional numeric shift amount — Not Started
   d. Add `BACKWARD_KEYWORD` and `FORWARD_KEYWORD` lexer tokens — Not Started
   e. Regenerate ANTLR parser/lexer Go files from updated `src/main/antlr/CPLMetrics.g4` — Not Started
   f. Verify generated files compile: `go build ./src/main/go/parsing/...` — Not Started

### Task 2: Define ParserTimeShift struct and extend ParserInputDefinition — Completed
   a. Create `src/main/go/parsing/parser_time_shift.go` with `ParserTimeShift` struct: `Direction string` (backward/forward), `Field string` (constraint field name), `Amount float64` (shift amount, default 1) — Not Started
   b. ~~Add `IsDefinitionData()` marker method so `ParserTimeShift` can be stored as `DefinitionData`~~ — **Removed**: `ParserTimeShift` is stored as a pointer field on `ParserInputDefinition` (which implements `DefinitionData`), never as a standalone `DefinitionData` value. No marker method needed.
   c. Extend `ParserInputDefinition` in `src/main/go/parsing/parser_input_definition.go` with `TimeShift *ParserTimeShift` field — Not Started
   d. Write unit tests verifying `ParserTimeShift` struct construction and `ParserInputDefinition` field defaults — Not Started
   e. Verify parser package compiles and tests pass: `go test ./src/main/go/parsing/...` — Not Started

### Task 3: Implement ExitBackwardClause/ExitForwardClause listener methods — Completed
   a. Implement `ExitBackwardClause` — create `ParserTimeShift` with Direction "backward" and attach to current `ParserInputDefinition` via scope context — Not Started
   b. Implement `ExitForwardClause` — create `ParserTimeShift` with Direction "forward" and attach to current `ParserInputDefinition` — Not Started
   c. Handle `shiftExpression` parsing: extract constraint field keyword and optional numeric amount into `ParserTimeShift.Field` and `ParserTimeShift.Amount` — Not Started
   d. Write listener unit tests: parse input declarations with `backward(year)`, `forward(month)`, `backward(day, 2)` and assert correct `ParserTimeShift` values on the input definition — Not Started
   e. Verify listener tests pass — Not Started

### Task 4: Define AnalysisTimeShift struct and extend AnalysisInputDefinition — Completed
   a. Define `AnalysisTimeShift` struct in `src/main/go/analysis/analysis_input_definition.go` with fields: `Direction string`, `Field string`, `Amount float64` — Not Started
   b. Extend `AnalysisInputDefinition` struct with `TimeShift *AnalysisTimeShift` field — Not Started
   c. Write unit tests verifying `AnalysisTimeShift` struct construction and `AnalysisInputDefinition` with nil vs populated `TimeShift` — Not Started
   d. Verify analysis package compiles: `go build ./src/main/go/analysis/...` — Not Started

### Task 5: Implement time shift validation logic — Completed
   a. Translate `ParserTimeShift` to `AnalysisTimeShift` in `resolveMetricInputs` within `src/main/go/analysis/metrics_analysis.go` — Not Started
   b. Validate time-shift constraint field references: field keyword must be a valid Instant constraint field (`year`, `month`, `day`, `hour`, `minute`, `second`, `day_of_week`, `week_number`) — report error with source location for unrecognized fields — Not Started
   c. Evaluate shift amount expressions: resolve numeric values (support simple numeric literals for this increment) — retain as `float64` in `AnalysisTimeShift.Amount` — Not Started
   d. Handle self-references with time shift: a metric referencing itself via `backward()`/`forward()` on an aliased input creates a time-series dependency (allowed). The self-reference is detected by input name matching the enclosing metric name. When a time shift is present, treat it as a dependency, not a cycle — Not Started
   e. Detect prohibited direct self-references: a metric referencing itself with no time shift (`input MyMetric`) is a true cycle and must produce an error — Not Started
   f. Write analyzer unit tests covering:
      - Valid backward/forward shifts with recognized constraint fields
      - Error for unrecognized field keywords
      - Self-reference with time shift (allowed)
      - Self-reference without time shift (error)
      - Multiple time-shifted inputs on a single metric — Not Started
   g. Verify analyzer tests pass — Not Started

### Task 6: Integration verification — Completed
   a. Create example CPL file with time-shifted inputs: metric A references metric B with `backward(year)` — Not Started
   b. Create example CPL file with self-referencing metric: metric Revenue calculates `revenue + revenue.backward(year)` — Not Started
   c. Verify parse stage produces correct `ParserTimeShift` values for both examples — Not Started
   d. Verify analysis stage correctly validates field references and handles self-reference semantics — Not Started
   e. Test edge cases: shift amount expressions (positive and negative), multiple time-shifted inputs, forward shift — Not Started
   f. Verify self-reference without time shift is rejected with a clear error — Not Started
   g. Run full CPL test suite to confirm no regressions: `go test ./... -count=1` from `src/chronocone_planning_language/src/main/go/` — Not Started

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task above, tests are written before implementation and the test-and-implement cycle is integrated within each task rather than broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Time-shift constraint field references must use valid Instant constraint field keywords
3. Self-references with time shift (backward/forward) are allowed — these create time-series dependencies
4. Self-references without time shift are prohibited — these are true cycles; no existing infrastructure for self-reference detection exists, so this is entirely new analysis logic
5. Parse/analysis separation maintained: parser captures syntax; analysis validates field references and shift amounts
6. Generated Go tests must follow the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Intent

- **Task 1 Intent**: Extend grammar with backward/forward time shift syntax for input declarations
- **Task 2 Intent**: Define parser data structures to capture time shift information during parsing
- **Task 3 Intent**: Implement listener methods that parse time shift syntax into structured data
- **Task 4 Intent**: Define analysis data structures to store validated time shift information
- **Task 5 Intent**: Validate time shift constraint fields, evaluate shift amounts, and correctly handle self-reference semantics during analysis
- **Task 6 Intent**: Verify time shift parsing and analysis pipeline works end-to-end with no regressions

## Acceptance Criteria

1. Input declaration grammar correctly parses `backward(year)` and `forward(month)` time shift clauses
2. `ParserTimeShift` correctly stores direction, constraint field, and shift amount
3. `AnalysisTimeShift` correctly carries validated time shift data through analysis
4. Time-shift constraint field references are validated against valid Instant constraint fields
5. Self-references with time shift are permitted (time-series dependency)
6. Self-references without time shift are detected and rejected with clear error messages
7. All unit tests pass following Go testing best practices
8. No regressions in existing CPL test suite
9. Parse/analysis separation maintained throughout
