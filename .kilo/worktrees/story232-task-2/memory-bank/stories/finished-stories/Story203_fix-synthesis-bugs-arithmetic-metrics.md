# Story203: Fix Synthesis Bugs for Arithmetic Metrics

## Goal

Fix two synthesis-layer bugs that prevent the generated Go code from compiling when arithmetic metric expressions are synthesized into `computeX()` functions. This story covers:

- Align generated code types with the `any`-based storage layer (`MetricComputation`, `ChronoconeTensor[any]`) to eliminate generic type invariance compile errors.
- Implement package-level arithmetic functions on `*ChronoconeTensor[any]` since methods cannot be added to the generic `T any` type.
- Correct `buildCILTree` to distinguish measure references (connector-provided inputs) from metric references.
- Fix the `inputs` map keying to use measure (celltype) names instead of connector names so synthesized code can resolve input tensors correctly.

## References

- [Fix Plan: Synthesis Bugs Blocking Arithmetic Metrics Examples](../plans/fix_synthesis_bugs_arithmetic_metrics_plan.md)
- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [ERO005: Metrics Synthesis Phase — Increment 2](../specifications/design/cpl/go-synthesis-target/eros/ERO005_metrics-synthesis-phase.md)
- [Synthesis Design](../specifications/design/cpl/go-synthesis-target/synthesis.md)
- [Tensor Representation Design](../specifications/design/cpl/go-synthesis-target/tensor-representation.md)

## Dependencies

- [Story184_cpl-metrics-arithmetic-parse-analyze.md](Story184_cpl-metrics-arithmetic-parse-analyze.md) — parse/analyze phase must produce valid `CILBinaryOp`, `CILUnaryOp`, and `CILInputRef` nodes with validated IO spec (Done)
- [Story185_cpl-metrics-arithmetic-synthesize.md](Story185_cpl-metrics-arithmetic-synthesize.md) — basic synthesizer infrastructure and arithmetic visitors must be in place (Done)

## Dependent Stories

- [Story186_cpl-metrics-parallelism-synthesize.md](Story186_cpl-metrics-parallelism-synthesize.md) — Increment 3 Synthesis (Not Started)
- [Story202_cpl-metrics-arithmetic-examples-revamp.md](Story202_cpl-metrics-arithmetic-examples-revamp.md) — runnable arithmetic examples depend on these fixes (Not Started)

## Tasks

1. Align generated code types with storage layer — Completed
   a. Define expected type alignment: change `*ChronoconeTensor[float64]` to `*ChronoconeTensor[any]` in generated code signatures (structural definition) — Completed
   b. Write tests for `cil_transformer.go` and `synthesize.go` to verify generated code uses `any` consistently — Completed
   c. Implement type alignment changes in `cil_transformer.go` (`TransformMetric` signature and result declaration) — Completed
   d. Implement type alignment changes in `synthesize.go` (inputs map type declarations) — Completed

2. Implement arithmetic as package-level functions on `*ChronoconeTensor[any]` — Completed
   a. Define function signatures for `TensorAddFloat64`, `TensorSubtractFloat64`, `TensorMultiplyFloat64`, `TensorDivideFloat64`, `TensorNegateFloat64`, `TensorAddScalarFloat64`, `TensorSubtractScalarFloat64`, `TensorMultiplyScalarFloat64`, `TensorDivideScalarFloat64`, and `TensorDimensionNamesMatch` in `calculation/chronocone_tensor.go` (structural definition) — Completed
   b. Write unit tests in `chronocone_tensor_arithmetic_test.go` for all arithmetic functions — Completed
   c. Implement package-level arithmetic functions in `calculation/chronocone_tensor.go` — Completed
   d. Update `VisitCILBinaryOp` in `cil_transformer.go` to emit function-call style code using the new functions — Completed
   e. Update `cil_transformer_test.go` expected snippets for the new function-call style — Completed

3. Fix `buildCILTree` to distinguish measure vs metric references — Completed
   a. Define `resolveInputRefKind` helper and update `buildCILTree` signature to accept `*CPLAnalyzer` (structural definition) — Completed
   b. Write tests for `buildCILTree` to verify measure references get `Kind: "connector"` and metric references get `Kind: "metric"` — Completed
   c. Implement `resolveInputRefKind` and update `buildCILTree` to use it — Completed
   d. Update `AnalyzeMetricScope` caller to pass the analyzer — Completed

4. Fix `inputs` map key to use measure name instead of connector name — Completed
   a. Define updated `generateInputLoadingCode` signature accepting `*analysis.AnalysisDefinitions` (structural definition) — Completed
   b. Write tests for `generateInputLoadingCode` to verify measure name keying — Completed
   c. Implement connector celltype lookup in `generateInputLoadingCode` — Completed
   d. Update `GenerateMainFunction` caller and `synthesize_test.go` for the new parameter — Completed

5. Update tests for code generation changes — Completed
   a. Update `cil_transformer_test.go` expected snippets for function-call style in `TestVisitCILBinaryOp` and `TestVisitCILBinaryOp_Nested` — Completed
   b. Update `synthesize_test.go` for new `generateInputLoadingCode` signature — Completed
   c. Run transformer and synthesis test suites to verify all tests pass — Completed

6. End-to-end verification — Completed
   a. Run the full CPL test suite: `cd src/chronocone_planning_language && go test ./...` — Completed
   b. Compile and run the arithmetic metrics example pipeline (parse → analyze → synthesize → compile → execute) — Completed
   c. Verify no regressions in existing tests — Completed

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. `MetricComputation` and `ChronoconeTensor[T]` type definitions must NOT be changed — the storage layer stays `any`-based
3. Arithmetic functions must support future multi-type extensibility (float64 today, int64/bignum tomorrow) via internal type assertions or sibling functions
4. Generated Go code must compile and run correctly
5. No breaking changes to existing non-arithmetic metric synthesis paths

## Intent

- **Task 1 Intent**: Eliminate the generic type invariance compile error between `*ChronoconeTensor[float64]` (generated code) and `*ChronoconeTensor[any]` (storage layer)
- **Task 2 Intent**: Provide callable arithmetic operations that work on `*ChronoconeTensor[any]` without requiring changes to the tensor generic constraint
- **Task 3 Intent**: Ensure measure references from connectors are correctly identified as `Kind: "connector"` so `VisitCILInputRef` emits `inputs["measure_name"]` instead of `controller.GetMetricComputation("measure_name").GetTensor()`
- **Task 4 Intent**: Ensure the `inputs` map keys match the measure names referenced in generated compute functions
- **Task 5 Intent**: Keep all existing unit tests aligned with the new code generation patterns
- **Task 6 Intent**: Prove the arithmetic pipeline works end-to-end with correct numeric output and no regressions

## Acceptance Criteria

1. Generated code consistently uses `*ChronoconeTensor[any]` — no type mismatch between `GetTensor()` return and variable declarations
2. All arithmetic package-level functions have unit tests and produce correct element-wise float64 results
3. `buildCILTree` assigns `Kind: "connector"` to measure references and `Kind: "metric"` to metric references
4. Generated `main()` stores tensors as `inputs["revenue_amt"]` instead of `inputs["revenue_data"]`
5. `VisitCILInputRef` for `Kind: "connector"` emits `inputs["revenue_amt"]` which matches the map key
6. All updated unit tests pass following Go testing best practices
7. Full end-to-end test: arithmetic expressions produce correct numeric results
8. No regressions in existing CPL test suite
