# Story212: CPL Short Form Metrics Examples

## Goal

Add CPL example files that exercise short form metric syntax (`<metric name> = <calculation>`) across both the `parse_and_analyze_only` and `with_data` example hierarchies. The examples must be automatically discoverable and testable by the existing Go test infrastructure without requiring test code changes (except for hardcoded definition files that enumerate examples).

This story is part of the CPL Metrics epic (Epic005). It depends on the short form parse+analyze work completed in Story193.

## References

- [Plan: Short Form Metric Examples](../plans/short_form_examples_plan.md)
- [Story193: CPL Metrics Short Form Parse+Analyze](Story193_cpl-metrics-short-form-parse-analyze.md)
- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [CPL Metrics Syntax Requirements](../specifications/requirements/cpl/metrics-syntax.md)

## Dependencies

- [Story193_cpl-metrics-short-form-parse-analyze.md](Story193_cpl-metrics-short-form-parse-analyze.md) — Done
- [Story202_cpl-metrics-arithmetic-examples-revamp.md](Story202_cpl-metrics-arithmetic-examples-revamp.md) — Done

## Dependent Stories

None

## Tasks

1. Create `parse_and_analyze_only/without_errors/short_form_metrics.cpl` — Completed
   a. Write the CPL file with concepts, units, types, measures, connectors, and four short form metrics (literal, connector ref, arithmetic, metric ref) — Completed
   b. Verify `TestExamplesWithoutErrors` passes for the new file — Completed

2. Create `parse_and_analyze_only/with_errors/short_form_undefined_reference.cpl` — Completed
   a. Write the CPL file with minimal infrastructure and one short form metric referencing an undefined identifier — Completed
   b. Verify `TestExamplesWithErrors` passes (compilation is rejected) — Completed

3. Create `parse_and_analyze_only/with_errors/short_form_name_collision.cpl` — Completed
   a. Write the CPL file with a concept and metric both named `Price`, plus a short form metric referencing the ambiguous name — Completed
   b. Verify `TestExamplesWithErrors` passes (compilation is rejected with collision error) — Completed

4. Create `with_data/metrics_arithmetic/short_form_addition/` end-to-end example — Completed
   a. Create `short_form_addition.cpl` with short form syntax (`sum_metric = input_a + input_b`) — Completed
   b. Create `iospec.json` identical to `simple_addition/iospec.json` — Completed
   c. Create `parquet_gen_spec_a.json` and `parquet_gen_spec_b.json` identical to `simple_addition` specs — Completed
   d. Run `./shell_scripts/regenerate_data.sh short_form_addition` to generate data files — Completed
   e. Create `expected_output.json` matching `simple_addition/expected_output.json` — Completed
   f. Verify `TestExampleFilesParse` and `TestMetricsArithmeticEndToEnd` pass — Completed

5. Update Go definition files for the new with_data example — Completed
   a. Add `ShortFormAdditionCPL` constant and documentation block to `example_definitions.go` — Completed
   b. Add `ShortFormAdditionIOSpecJSON` constant and `ShortFormAdditionIOSpec()` function to `iospec_definitions.go` — Completed
   c. Add test case entry to `iospec_examples_test.go` — Completed
   d. Verify all three Go files compile and `TestIOSpecExamples` passes — Completed

6. Update `examples/README.md` — Completed
   a. Add entries documenting the three parse_and_analyze_only examples and the short_form_addition end-to-end example — Completed

7. Run full example test suite and verify no regressions — Completed
   a. Run `TestExamplesWithoutErrors` and `TestExamplesWithErrors` — Completed
   b. Run `TestExampleFilesParse`, `TestMetricsArithmeticEndToEnd`, and `TestIOSpecExamples` — Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, tests are written before implementing the code. The test-and-implement cycle is integrated within each task in the correct order. Do not break tests out as separate tasks.

For this story, the "tests" are the existing Go test infrastructure that auto-discovers example files:

- `TestExamplesWithoutErrors` and `TestExamplesWithErrors` in `cpl_example_test.go` auto-discover `.cpl` files in `parse_and_analyze_only/` directories.
- `TestExampleFilesParse` and `TestMetricsArithmeticEndToEnd` in `metrics_arithmetic_test.go` auto-discover leaf directories containing `iospec.json` and `.cpl` files.
- `TestIOSpecExamples` in `iospec_examples_test.go` verifies hardcoded IO spec definitions.

The workflow for each task is:

1. Create the example file(s).
2. Run the relevant test(s) to verify they pass (or fail as expected for error examples).
3. If tests fail, fix the example file(s) until tests pass.

## Constraints

- Example files must follow the conventions established by Story202 and Story207.
- The `with_data` example must reuse the existing `metrics_arithmetic/core.cpl` infrastructure.
- Go definition files must follow the existing pattern for constants and functions.
- All generated data files must be produced by the `regenerate_data.sh` script, not hand-crafted.
- The README must accurately describe each example's purpose.

## Intent

The intent is to provide comprehensive, discoverable, and testable examples that demonstrate the short form metric syntax in both valid and invalid forms, as well as a full end-to-end example that proves the syntax works through synthesis and data evaluation.

## Acceptance Criteria

- `parse_and_analyze_only/without_errors/short_form_metrics.cpl` exists and `TestExamplesWithoutErrors` passes for it.
- `parse_and_analyze_only/with_errors/short_form_undefined_reference.cpl` exists and `TestExamplesWithErrors` passes for it (compilation rejected).
- `parse_and_analyze_only/with_errors/short_form_name_collision.cpl` exists and `TestExamplesWithErrors` passes for it (compilation rejected with collision error).
- `with_data/metrics_arithmetic/short_form_addition/` exists with all required files and `TestMetricsArithmeticEndToEnd` passes.
- Go definition files compile without errors and `TestIOSpecExamples` includes the new case.
- `examples/README.md` documents all new examples.
- Full example test suite passes with no regressions.

## Notes

- Short form metrics are defined as `<metric name> = <calculation>`.
- Inputs are auto-derived from identifier references using resolution precedence: measures/connectors → other metrics → concepts.
- The `short_form_addition` example is functionally equivalent to `simple_addition` but uses short form syntax.
