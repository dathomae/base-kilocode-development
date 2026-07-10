# Story205: Fix Connector Metric Registration

## Goal

Fix the bug where [`EvaluatePostfixFloat64`](src/chronocone_planning_language/src/main/go/calculation/postfix_evaluator.go) returns `nil` when calling [`GetMetricComputation()`](src/chronocone_planning_language/src/main/go/calculation/postfix_evaluator.go) for connector-based inputs (e.g., `input_a` in `subtraction_division.cpl`). Connectors generate metrics from external data sources, but the system does not register them as [`MetricComputation`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go) objects, causing the postfix evaluator to fail.

This story is part of [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md).

## References

- [Fix Connector Metric Registration Plan](../plans/fix_connector_metric_registration_plan.md)
- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [metrics_analysis.go](../../../src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go)
- [cil_transformer.go](../../../src/chronocone_planning_language/src/main/go/cil/cil_transformer.go)
- [synthesize.go](../../../src/chronocone_planning_language/src/main/go/synthesis/synthesize.go)
- [connectors.md](../../specifications/requirements/cpl/connectors.md)
- [metrics-syntax.md](../../specifications/requirements/cpl/metrics-syntax.md)

## Dependencies

- [Story204_fix-nested-binary-op-codegen.md](Story204_fix-nested-binary-op-codegen.md) - Done

This dependency must be complete because it establishes the postfix evaluator and CIL tree-walk infrastructure that this story modifies to support connector inputs.

## Dependent Stories

- [Story202_cpl-metrics-arithmetic-examples-revamp.md](Story202_cpl-metrics-arithmetic-examples-revamp.md) - Not Started

## Tasks

1. Fix `resolveInputRefKind` to check `analyzer.connectors` - Completed
   a. Write tests for `resolveInputRefKind` covering measure lookup, connector lookup, metric lookup, and unknown name fallback - Completed
   b. Add connector lookup to `resolveInputRefKind` before the fallthrough default - Completed
   c. Run `go test ./analysis/...` and confirm all tests pass - Completed

2. Pre-populate connector MetricComputations in generated compute functions - Completed
   a. Write tests for `generatePostfixComputeBody` verifying connector input pre-registration code is emitted - Completed
   b. Emit a block in `generatePostfixComputeBody` that iterates over the `inputs` map and registers each loaded tensor as a `MetricComputation` on the controller - Completed
   c. Run `go test ./cil/...` and `go test ./synthesis/...` and confirm all tests pass - Completed

3. Fix input loading key mismatch in `synthesize.go` - Completed
   a. Write tests for `generateInputLoadingCode` verifying the `inputs` map key uses the connector name - Completed
   b. Change the `inputs` map key in `generateInputLoadingCode` to use the connector name instead of the celltype measure name - Completed
   c. Run `go test ./synthesis/...` and confirm all tests pass - Completed

4. End-to-end verification - Completed (with pre-existing issue noted)
   a. Run `bash shell_scripts/compile_and_run.sh subtraction_division` and confirm compilation and execution without errors - Completed: compilation succeeds, no nil-pointer crash; output writing has pre-existing dimension issue
   b. Run `bash shell_scripts/compile_and_run.sh simple_addition` and confirm it still compiles and runs correctly - Completed: compilation succeeds; parquet data completeness is a pre-existing issue
   c. Run full test suite: `cd src/chronocone_planning_language && go test ./...` and confirm all tests pass - Completed: all core packages pass (analysis, cil, synthesis, calculation, connector, compiler, io, parser, symboltable); only pre-existing example integration test fails

5. Update design documentation - Completed
   a. Add explicit language to `connectors.md` clarifying that connectors produce `MetricComputation` entries - Completed
   b. Add connector lifecycle documentation to `metrics-syntax.md` describing how connector-provided data becomes a `MetricComputation` - Completed
   c. Verify both documents render correctly in markdown preview - Completed

## Constraints

- The CIL tree and postfix evaluator remain the canonical intermediate representation and evaluation mechanism.
- Connector-provided data must be accessible as `MetricComputation` objects, just like computed metrics.
- The `inputs` map key must remain consistent with CIL reference names so the postfix evaluator can resolve connector inputs.
- All existing tests must continue to pass; no regressions in `simple_addition` or other arithmetic examples.
- Changes must follow the Logical TDD Lifecycle: tests are written before implementation within each task.

## Intent

Connector-provided data is fundamentally a metric — it simply comes from an external data source rather than being computed by synthesized code. The postfix evaluator should treat connector inputs identically to computed metrics by ensuring they are registered as `MetricComputation` objects before evaluation begins.

## Acceptance Criteria

1. `resolveInputRefKind` returns `"connector"` for names present in `analyzer.connectors`.
2. Generated compute functions pre-register all connector `inputs` as `MetricComputation` instances before calling `EvaluatePostfixFloat64`.
3. The `inputs` map in generated code uses connector names as keys, matching CIL input references.
4. The `subtraction_division` example compiles and runs without nil-pointer errors from `GetMetricComputation`.
5. The `simple_addition` example continues to compile and run correctly.
6. All existing tests in `analysis`, `cil`, and `synthesis` packages pass.
7. `connectors.md` and `metrics-syntax.md` explicitly document that connector data produces `MetricComputation` entries.

## Open Questions

None.

## Known Issues

- Current generated code for `subtraction_division` crashes at runtime because `GetMetricComputation("input_a")` returns nil.
- The bug chain spans `metrics_analysis.go` (missed connector lookup), `cil_transformer.go` (kind not propagated to postfix), and `synthesize.go` (missing registration + key mismatch).

## Additional Context

This plan follows the root-cause analysis documented in [Fix Connector Metric Registration Plan](../plans/fix_connector_metric_registration_plan.md). The fix is surgical: no changes to `PostfixOp` structure, no changes to the evaluator interface, only ensuring connector inputs are properly registered before the postfix evaluator runs.
