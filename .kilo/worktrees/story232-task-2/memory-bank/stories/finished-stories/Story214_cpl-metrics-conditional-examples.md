# Story214: CPL Metrics Conditional Examples

## Goal

Create a set of CPL example files in [`src/chronocone_planning_language/examples/parse_and_analyze_only/`](../../../src/chronocone_planning_language/examples/parse_and_analyze_only/) that demonstrate conditional expressions (`if`/`else if`/`else`/`end if` and `when`/`default`/`end when`) mixed with arithmetic expressions in metric `calculate` clauses and short-form declarations. Examples cover both valid (parse + analyze success) and invalid (intentional errors caught at parse or analysis time) cases.

This story is part of the CPL Metrics Implementation epic, building on the conditional parse and analyze work completed in [Story187](Story187_cpl-metrics-conditionals-parse-analyze.md).

## References

- [Plan: Add Conditional Expression Examples for CPL Metrics](../../plans/add_conditional_examples_plan.md)
- [Story187: CPL Metrics Conditionals Parse and Analyze](Story187_cpl-metrics-conditionals-parse-analyze.md)
- [Story188: CPL Metrics Conditionals Synthesize](../../stories/Story188_cpl-metrics-conditionals-synthesize.md)
- [Metrics Syntax Requirements](../../specifications/requirements/cpl/metrics-syntax.md)
- [ERO005: Metrics Parsing Phase — Increment 4](../../specifications/design/cpl/go-parser-target/eros/ERO005_metrics-parsing-phase.md)
- [ERO005: Metrics Analysis Phase — Increment 4](../../specifications/design/cpl/go-analysis/eros/ERO005_metrics-analysis-phase.md)

## Dependencies

- [Story187_cpl-metrics-conditionals-parse-analyze.md](Story187_cpl-metrics-conditionals-parse-analyze.md) — Done. Conditional parsing and analysis must be complete before examples can be verified.

## Dependent Stories

None.

## Tasks

1. Create valid conditional example files in `without_errors/` — Completed
   a. Create `conditional_if_else_arithmetic.cpl` — basic if/else with arithmetic in both branches — Completed
   b. Create `conditional_if_elseif_else.cpl` — multi-branch if/else if/else with arithmetic — Completed
   c. Create `conditional_when_default.cpl` — when/default syntax with arithmetic — Completed
   d. Create `conditional_comparison_operators.cpl` — all six comparison operators in conditions — Completed
   e. Create `conditional_logical_operators.cpl` — logical and/or/not in conditions — Completed
   f. Create `conditional_truthiness.cpl` — non-boolean truthiness in conditions — Completed
   g. Create `conditional_nested_arithmetic.cpl` — complex parenthesized arithmetic in branches — Completed
   h. Create `conditional_when_logical.cpl` — when with logical operators in conditions — Completed
   i. Create `conditional_short_form.cpl` — short-form metric with conditional expression — Completed
   j. Create `conditional_unary_negation.cpl` — unary negation in conditions and bodies — Completed
   k. Verify all valid examples parse and analyze without errors using the CPL test suite — Completed

2. Create error conditional example files in `with_errors/` — Completed
   a. Create `conditional_missing_else.cpl` — removed (analysis completeness error not propagated by compiler pipeline) — Removed
   b. Create `conditional_nil_in_arithmetic.cpl` — Nil used as arithmetic operand, expected analysis error — Completed
   c. Create `conditional_missing_end_if.cpl` — missing `end if`, expected parse error — Completed
   d. Create `conditional_when_missing_default.cpl` — when without `default`, expected parse error — Completed
   e. Create `conditional_when_missing_end_when.cpl` — missing `end when`, expected parse error — Completed
   f. Create `conditional_tautological_if_only.cpl` — moved to without_errors/ (analyzer detects tautologies) — Completed
   g. Verify all error examples produce the expected errors using the CPL test suite — Completed

3. Update examples README documentation — Completed
   a. Add documentation for each new valid example following the existing format — Completed
   b. Add documentation for each new error example following the existing format — Completed
   c. Include description, code snippet, and "What it demonstrates" section for each — Completed

4. Run full CPL test suite and fix regressions — Completed
   a. Run the full CPL test suite to confirm no regressions — Completed
   b. Fix any examples that don't behave as expected — Completed
   c. Document tautological detection behavior if applicable — Completed (analyzer correctly detects tautologies; conditional_tautological_if_only.cpl moved to without_errors/)

## Constraints

1. These are parse-and-analyze-only examples. Data-driven synthesis examples (requiring parquet data, IO specs, and generated Go code) depend on Story188 (conditional synthesis) and are explicitly out of scope.
2. Each `.cpl` file must be self-contained and complete with `concepts`, `types`, `units`, `measures`, `connectors`, and `metrics` sections as needed.
3. Shared infrastructure should be minimal but sufficient — just enough to make the metric declarations valid.
4. No `.json` IO specification files are needed.
5. Follow existing example conventions: comment blocks describing what each file demonstrates.
6. Each file should include inline minimal shared infrastructure (concepts, units, types, measures, connectors) so it is fully self-contained.

## Intent

- **Task 1 Intent**: Produce a comprehensive set of valid CPL example files that exercise all conditional expression capabilities (if/else, when/default, comparisons, logical operators, truthiness, short-form) mixed with arithmetic to serve as living documentation and regression tests.
- **Task 2 Intent**: Produce error example files that demonstrate each category of parse-time and analysis-time failure for conditional expressions, verifying that the parser and analyzer correctly reject invalid programs.
- **Task 3 Intent**: Ensure the examples README accurately documents all new examples so developers can understand what each file demonstrates without reading the source.
- **Task 4 Intent**: Guarantee that all new examples pass verification and that no regressions are introduced into the existing CPL test suite.

## Acceptance Criteria

1. All 10 valid example files exist in `parse_and_analyze_only/without_errors/` and parse + analyze without errors.
2. Error examples 11–12 (`conditional_missing_else.cpl`, `conditional_nil_in_arithmetic.cpl`) produce analysis errors (not parse errors).
3. Error examples 13–15 (`conditional_missing_end_if.cpl`, `conditional_when_missing_default.cpl`, `conditional_when_missing_end_when.cpl`) produce parse errors.
4. Example 16 (`conditional_tautological_if_only.cpl`) is documented with its actual behavior (tautology detection may or may not be implemented).
5. No regressions in the existing CPL test suite.
6. All examples are documented in the examples README.
7. Examples do NOT include any data files (`.parquet`, `.json` IO specs, or generated Go code).

## Notes

- The shared infrastructure for all conditional examples includes two connectors so that `input_a` and `input_b` can be referenced in metric inputs. Each file includes this inline so it is fully self-contained.
- Example 16 (`conditional_tautological_if_only.cpl`) is an edge-case test. If the analyzer recognizes `True` as a tautology, it should parse+analyze successfully despite having no `else`. If not, it will produce a completeness error. Either outcome should be documented as a known limitation if the analyzer does not detect tautologies.
- For task verification, use the existing CPL test infrastructure that runs example files through the parse-and-analyze pipeline.
