# Story233: Fix AND/OR Precedence in Where Clause Grammar

## Goal

Fix the where clause grammar so that `and` binds tighter than `or`, matching standard boolean operator precedence. Currently `whereExpression` produces a flat list of conditions with interleaved logical operators, meaning `aspect=A and aspect=B or aspect=C` is indistinguishable from `aspect=A or aspect=B and aspect=C`.

This story restructures both the grammar and the downstream parser/analysis data structures from flat slices to an OR-of-AND tree (a restricted disjunctive normal form). The `LogicalOp` field is removed entirely — operator type becomes implicit in the tree structure.

## References

- [Plan: Fix AND/OR Precedence in Where Clause Grammar](../plans/fix-precedence-bug-plan.md)
- [Brief](../brief.md)
- [Requirements](../specifications/requirements/requirements.md)
- [Design](../specifications/design/design.md)

## Dependencies

- [Story189_cpl-metrics-aliases-where-parse-analyze.md](finished-stories/Story189_cpl-metrics-aliases-where-parse-analyze.md) — _Done_: This story built the original flat where clause parsing and analysis that this bug-fix restructures.

## Dependent Stories

- [Story190_cpl-metrics-aliases-where-synthesize.md](Story190_cpl-metrics-aliases-where-synthesize.md) — _Not Started_: Synthesis of where clause filtering depends on the final where clause data structures produced by this story.

## Tasks

### Task Sizing

- **Task 1** is not small: it spans 4 source files across 2 packages (parsing, parser) and requires reasoning about ANTLR context hierarchy changes. Grammar, struct, and listener changes form one indivisible compilation unit.
- **Task 2** is `[Small]`: single test file, mechanical assertion path updates plus additive test functions following existing patterns.
- **Task 3** is not small: 3 files across analysis and analysis_test, introducing new structs and updating resolution logic.
- **Task 4** is an operational verification task with no code changes.

### Parallel Execution

- **Group A** (Tasks 2, 3): Task 2 modifies `cpl_metrics_parser_test.go` only; Task 3 modifies `analysis_where_clause.go`, `metrics_analysis.go`, and `metrics_analysis_test.go` only. Different packages (parser_test vs analysis + analysis_test). Both depend on Task 1's output (new structs and listener) but do not depend on each other's changes.

### Execution Order

```
Task 1 → [Task 2 ‖ Task 3] → Task 4
```

### Reintegration

**Group A (Tasks 2, 3):**
- Merge order: any order (fully disjoint files and packages, no shared infrastructure)
- Integration tests: `cd src/chronocone_planning_language && go test ./src/main/go/... -count=1` — verify combined parse→analyze pipeline with new OR-of-AND where clause structures
- Watch for conflicts in: none expected

### Task 1: Restructure grammar, parser data structures, listener, and regenerate ANTLR — Completed

Rewrite the where clause grammar from a flat list to a precedence hierarchy, update parser data structures to an OR-of-AND tree, rewrite the listener extraction logic, and regenerate ANTLR code. All changes must compile together as a unit.

a. Modify `src/main/antlr/CPLMetrics.g4` — replace flat `whereExpression` / `logicalOp` rules with precedence hierarchy:
   ```
   whereExpression : orExpression ;
   orExpression : andExpression (OR_KEYWORD andExpression)* ;
   andExpression : aspectExpression (AND_KEYWORD aspectExpression)* ;
   aspectExpression : NOT_KEYWORD? ASPECT_KEYWORD EQUALS (ACTUALS_KEYWORD | FORECAST_KEYWORD) ;
   ```
   Remove the `logicalOp` rule entirely.

b. Modify `src/main/go/parsing/parser_where_condition.go` — replace the flat `ParserWhereCondition` struct with OR-of-AND tree types:
   ```go
   type ParserWhereClause struct {
       AndGroups []ParserWhereAndGroup
   }
   type ParserWhereAndGroup struct {
       Conditions []ParserWhereCondition
   }
   type ParserWhereCondition struct {
       Aspect  string
       Negated bool
   }
   ```
   Remove the `LogicalOp` field entirely.

c. Modify `src/main/go/parsing/parser_input_definition.go` — change the field from a flat slice to the new tree type:
   ```go
   // Before:
   WhereConditions []ParserWhereCondition
   // After:
   WhereClause *ParserWhereClause
   ```

d. Modify `src/main/go/parser/cpl_metrics_listener.go` — rewrite `ExitInputDeclaration` where clause extraction to traverse the new ANTLR context hierarchy. Build a `ParserWhereClause` with `AndGroups` populated from `orExpression` → `andExpression` → `aspectExpression`.

e. Run `./gradlew generateAntlr` to regenerate ANTLR Go files.

f. Verify the parser package compiles: `cd src/chronocone_planning_language && go build ./src/main/go/parser/... ./src/main/go/parsing/...`

### Task 2: Update parser tests for new where clause structure — Completed

Update existing parser where clause tests to assert via the new `WhereClause.AndGroups` tree structure, and add new precedence tests verifying that `and` binds tighter than `or`.

a. Update existing where clause tests in `src/main/go/parser/cpl_metrics_parser_test.go`:
   - `TestParseInputWithAlias` — assert `WhereClause` is nil (no where clause)
   - `TestParseInputWithWhereClause` — assert via `WhereClause.AndGroups[0].Conditions[0]`
   - `TestParseInputWithAliasAndWhere` — same pattern
   - `TestParseInputWithComplexWhere` — now asserts one AND group with 2 conditions: `AndGroups[0].Conditions[0]` and `AndGroups[0].Conditions[1]`
   - `TestParseInputWithNotAspect` — assert `WhereClause.AndGroups[0].Conditions[0].Negated`
   - `TestParseInputWithoutAliasOrWhere` — assert `WhereClause` is nil

b. Add new precedence tests:
   - `TestParseInputWithOr` — `where aspect=actuals or aspect=forecast` → 2 AndGroups, each with 1 condition
   - `TestParseInput_AndBindsTighterThanOr_AndFirst` — `where aspect=A and aspect=B or aspect=C` → AndGroup 0: [A, B], AndGroup 1: [C]
   - `TestParseInput_AndBindsTighterThanOr_OrFirst` — `where aspect=A or aspect=B and aspect=C` → AndGroup 0: [A], AndGroup 1: [B, C]
   - `TestParseInput_MultipleOrWithAnd` — `where aspect=A and aspect=B or aspect=C and aspect=D` → AndGroup 0: [A, B], AndGroup 1: [C, D]
   - `TestParseInput_SingleAndGroup` — `where aspect=A and aspect=B` → 1 AndGroup with 2 conditions
   - `TestParseInput_OrWithNotAnd` — `where aspect=A or not aspect=B and aspect=C` → AndGroup 0: [A], AndGroup 1: [not B, C]

c. Run parser tests: `cd src/chronocone_planning_language && go test ./src/main/go/parser/... -count=1 -timeout 60s`

### Task 3: Update analysis data structures, resolution, and tests — Completed

Mirror the parser restructuring in the analysis package: replace flat where clause structures with the OR-of-AND tree, update resolution logic, and update/add tests.

a. Modify `src/main/go/analysis/analysis_where_clause.go` — replace flat structures with OR-of-AND tree:
   ```go
   type AnalysisWhereClause struct {
       AndGroups []AnalysisWhereAndGroup
   }
   type AnalysisWhereAndGroup struct {
       Conditions []AnalysisWhereCondition
   }
   type AnalysisWhereCondition struct {
       Negated bool
       Aspect  string
   }
   ```
   Remove the `LogicalOp` field.

b. Modify `src/main/go/analysis/metrics_analysis.go`:
   - Update `resolveWhereClause` to accept `*parsing.ParserWhereClause` and produce `*AnalysisWhereClause` by walking the parser tree's `AndGroups` and `Conditions`.
   - Update the call site in `resolveMetricInputs` to pass `input.WhereClause` instead of `input.WhereConditions`.

c. Update existing analysis tests in `src/main/go/analysis/metrics_analysis_test.go`:
   - `TestAnalyzeMetricScope_WithWhereClause` — assert via `WhereClause.AndGroups[0].Conditions[0]`
   - `TestAnalyzeMetricScope_WithWhereNotClause` — same pattern
   - `TestAnalyzeMetricScope_WithComplexWhere` — now asserts one AND group with 2 conditions
   - `TestAnalyzeMetricScope_EndToEnd_AliasAndWhere` — update assertions

d. Add new analysis precedence tests mirroring parser tests:
   - `TestAnalyzeMetricScope_OrPrecedence` — `where aspect=A or aspect=B` → 2 AndGroups, 1 condition each
   - `TestAnalyzeMetricScope_AndBindsTighter` — `where aspect=A and aspect=B or aspect=C` → AndGroup 0: [A,B], AndGroup 1: [C]
   - `TestAnalyzeMetricScope_AndBindsTighterReversed` — `where aspect=A or aspect=B and aspect=C` → AndGroup 0: [A], AndGroup 1: [B,C]

e. Run analysis tests: `cd src/chronocone_planning_language && go test ./src/main/go/analysis/... -count=1 -timeout 60s`

### Task 4: Integration verification — Completed

Run the full test suite and verify example files still parse and analyze correctly.

a. Run full test suite: `cd src/chronocone_planning_language && go test ./src/main/go/... -count=1 -timeout 120s`

b. Verify example files parse and analyze correctly:
   - `src/chronocone_planning_language/src/test/resources/cpl/parse_and_analyze_only/without_errors/metrics_aliases_where.cpl`
   - `src/chronocone_planning_language/src/test/resources/cpl/parse_and_analyze_only/with_errors/metrics_duplicate_alias.cpl`

c. Confirm no regressions in any package.

## Test-First Development

This story follows the Logical TDD Lifecycle. Tests are integrated within each task rather than broken out as separate tasks:

- **Task 1**: No new tests (structural compilation-only change), but compilation verification acts as the acceptance gate.
- **Task 2**: Existing parser tests are updated to assert the new structure before the story is considered complete; new precedence tests are added to verify the corrected behavior.
- **Task 3**: Existing analysis tests are updated to assert the new structure; new precedence tests verify the analysis tree preserves the same AND/OR grouping.
- **Task 4**: Full integration test suite acts as the final acceptance gate.

## Constraints

1. Do not break existing short-form metric behavior (short form inputs have no where clauses).
2. Maintain parse/analysis separation: parser captures the tree from grammar; analysis copies it without additional interpretation (synthesis will interpret).
3. All existing tests must continue to pass after restructuring assertions.
4. Regenerated ANTLR files must not be hand-edited.
5. Do not introduce parentheses for explicit grouping — OR-of-AND covers all meaningful combinations for this domain.

## Intent

For each task, the overall goal is:

- **Task 1**: Establish the correct grammatical precedence (AND tighter than OR) and the corresponding parser-side tree representation so the parse tree unambiguously encodes operator precedence.
- **Task 2**: Verify that the parser produces the correct OR-of-AND tree for all combinations of `and`, `or`, and `not`.
- **Task 3**: Ensure the analysis phase faithfully mirrors the parser tree without semantic distortion.
- **Task 4**: Guarantee that the restructured where clause handling integrates cleanly with the rest of the CPL pipeline.

## Acceptance Criteria

- [ ] `where aspect=A and aspect=B or aspect=C` parses as `(A AND B) OR C` (two AndGroups: [A,B] and [C]).
- [ ] `where aspect=A or aspect=B and aspect=C` parses as `A OR (B AND C)` (two AndGroups: [A] and [B,C]).
- [ ] `where aspect=A and aspect=B` parses as a single AndGroup with two conditions.
- [ ] `where aspect=A or aspect=B` parses as two AndGroups, each with one condition.
- [ ] Negation (`not aspect=forecast`) is preserved per-condition in the tree.
- [ ] The `LogicalOp` field does not exist in any parser or analysis struct.
- [ ] All parser tests pass.
- [ ] All analysis tests pass.
- [ ] Full `go test ./src/main/go/...` passes with no regressions.
- [ ] Example CPL files continue to parse and analyze correctly.

## Notes

- The OR-of-AND structure (disjunctive normal form) is sufficient for the CPL where clause domain because only two primitive values exist (`ACTUALS`, `FORECAST`). A full boolean expression tree would be overengineered.
- This structure maps naturally to the synthesis phase (Story190), where each input will be resolved separately per aspect.
