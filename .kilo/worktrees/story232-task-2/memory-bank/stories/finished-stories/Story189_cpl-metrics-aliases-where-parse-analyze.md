# Story189: CPL Metrics Aliases and Where Clauses Parse and Analyze

## Goal

Implement the Increment 5 parsing and analysis phases for CPL metric input aliases and where clauses as part of [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md). This story covers:

- Extend `inputDeclaration` grammar with optional `as` alias clause and `where` clause with `aspect=ACTUALS`/`aspect=FORECAST` filtering and boolean combinations (`and`, `or`, `not`).
- Store alias mapping and where conditions in [`ParserInputDefinition`](src/chronocone_planning_language/src/main/go/parsing/parser_input_definition.go).
- Extend analysis input resolution to validate alias uniqueness within each metric's input block, resolve where-clause aspect references, and record aspect filters on resolved input references.

## References

- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [Metrics Syntax Requirements](../specifications/requirements/cpl/metrics-syntax.md)
- [ERO005: Metrics Parsing Phase — Increment 5](../specifications/design/cpl/go-parser-target/eros/ERO005_metrics-parsing-phase.md)
- [ERO005: Metrics Analysis Phase — Increment 5](../specifications/design/cpl/go-analysis/eros/ERO005_metrics-analysis-phase.md)
- [CPL Architecture — Introduction](../specifications/architecture/cpl/introduction.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story184_cpl-metrics-arithmetic-parse-analyze.md](Story184_cpl-metrics-arithmetic-parse-analyze.md) — arithmetic parse/analyze with IO spec acceptance must be complete (Not Started)

## Dependent Stories

- [Story190_cpl-metrics-aliases-where-synthesize.md](Story190_cpl-metrics-aliases-where-synthesize.md) — Increment 5 Synthesis (Not Started)

## Tasks

1. Extend input declaration grammar with alias and where clauses — Completed
   a. Add optional `as` alias clause to `inputDeclaration`: `AS identifier` — Completed
   b. Add optional `where` clause: `WHERE_KEYWORD whereExpression` — Completed
   c. Add `whereExpression` rule: `aspectExpression (logicalOp aspectExpression)*` — Completed
   d. Add `aspectExpression` rule: `ASPECT_KEYWORD ASSIGN (ACTUALS_KEYWORD | FORECAST_KEYWORD)` — Completed
   e. Add `AS_KEYWORD`, `WHERE_KEYWORD`, `ASPECT_KEYWORD`, `ACTUALS_KEYWORD`, `FORECAST_KEYWORD` lexer tokens — Completed
   f. Regenerate ANTLR parser/lexer Go files — Completed

2. Update parser data structures and listener — Completed
   a. Extend [`ParserInputDefinition`](src/chronocone_planning_language/src/main/go/parsing/parser_input_definition.go) with `Alias string` field and `WhereConditions []ParserWhereCondition` — Completed
   b. Define [`ParserWhereCondition`](src/chronocone_planning_language/src/main/go/parsing/parser_where_condition.go) with fields: `Aspect string` (ACTUALS/FORECAST), `LogicalOp string` (and/or/not) — Completed
   c. Implement `ExitInputAsClause` — set alias on current `ParserInputDefinition` — Completed
   d. Implement `ExitWhereExpression` / `ExitAspectExpression` — push where condition onto current input definition — Completed
   e. Verify parser tests: parse input declarations with aliases and where clauses, assert correct `ParserInputDefinition` values — Completed

3. Implement analysis phase for alias and where clause validation — Completed
   a. Extend `AnalysisInputDefinition` with `Alias string` and `AspectFilter string` fields — Completed
   b. Implement alias uniqueness validation: each alias within a metric's input block must be unique — Completed
   c. Implement where-clause aspect resolution: resolve `ACTUALS`/`FORECAST` references — Completed
   d. Record aspect filter on resolved input references for synthesis use — Completed
   e. Verify analyzer tests: alias uniqueness enforced, aspect filters correctly recorded — Completed

4. Integration verification — Completed
   a. Create example CPL files with aliased inputs and where clauses — Completed
   b. Verify parse-analyze pipeline produces correct analysis data with aliases and aspect filters — Completed
   c. Test edge cases: duplicate aliases, complex where expressions with and/or/not — Completed
   d. Run the full CPL test suite to confirm no regressions — Completed

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Aliases must be unique within each metric's input block
3. Where clause aspect filtering supports only `ACTUALS` and `FORECAST` for this increment
4. Parse/analysis separation maintained: parser captures syntax; analysis validates uniqueness and resolves references
5. Generated Go tests must follow the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Intent

- **Task 1 Intent**: Extend grammar with alias and where clause syntax to support input renaming and aspect-based filtering
- **Task 2 Intent**: Capture alias names and where conditions in parser data structures during parsing
- **Task 3 Intent**: Validate alias uniqueness and resolve aspect references during analysis — prepare aspect filters for synthesis
- **Task 4 Intent**: Verify alias and where clause pipeline works end-to-end

## Acceptance Criteria

1. `inputDeclaration` grammar correctly parses optional `as` alias and `where` clauses
2. `ParserInputDefinition` correctly stores alias and where condition data
3. Alias uniqueness is validated within each metric's input block
4. Aspect references (`ACTUALS`/`FORECAST`) are correctly resolved and recorded
5. All unit tests pass following Go testing best practices
6. No regressions in existing CPL test suite
7. Parse/analysis separation maintained
