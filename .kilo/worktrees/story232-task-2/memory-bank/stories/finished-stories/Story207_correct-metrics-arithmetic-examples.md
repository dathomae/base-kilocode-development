# Story207: Correct Metrics Arithmetic Examples

## Goal

Correct all `metrics_arithmetic` examples (except `subtraction_division`, which is the reference) so they follow the same structural patterns as the reference. The reference example defines its own connectors (`input_a`, `input_b`, `input_c`) with `celltype packages` (the measure from `core.cpl`), uses full `Daily` grain (all 7 days), and has no unused declarations. Update Go test expectations and supporting code to match the corrected examples.

## References

- [Plan](../plans/correct_examples_plan.md)
- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [Story202: CPL Metrics Arithmetic Examples Revamp](Story202_cpl-metrics-arithmetic-examples-revamp.md)

## Dependencies

- [Story202_cpl-metrics-arithmetic-examples-revamp.md](Story202_cpl-metrics-arithmetic-examples-revamp.md) — In Progress
  - The examples directory structure, `core.cpl`, and initial `parquetgen`/`parquet_verify` tooling must exist before corrections can be applied.
- [Story203_fix-synthesis-bugs-arithmetic-metrics.md](Story203_fix-synthesis-bugs-arithmetic-metrics.md) — Done
  - Synthesis-layer type mismatch fixes must be in place for generated Go code to compile.
- [Story204_fix-nested-binary-op-codegen.md](Story204_fix-nested-binary-op-codegen.md) — Done
  - Nested binary operation code generation must work for complex expressions to synthesize correctly.

## Dependent Stories

None

## Tasks

1. Update Go test expectations and supporting code to match the target reference pattern — Completed
   a. Update `example_definitions.go` comments for all examples: remove references to connectors being defined in `core.cpl`, update connector names to `input_a`/`input_b`/`input_c`, remove references to extra measures in example files — Completed
   b. Update `iospec_definitions.go` builders: rename connector names from `data_*` to `input_*`, change celltype to `packages`, change grain to `["Daily"]`, fix `SubtractionDivisionIOSpec` celltype from `package` to `packages` — Completed
   c. Update `iospec_examples_test.go` expectations: `expectedConnectors` to `input_*`, `expectedCellType` to `packages`, `expectedGrain` to `["Daily"]` — Completed
   d. Run `go test ./...` in the `metrics_arithmetic` package to confirm tests fail with the new expectations — Completed

2. Correct `simple_addition` example files — Completed
   a. Update `simple_addition.cpl`: add `connectors` block declaring `input_a` and `input_b` with `celltype packages` — Completed
   b. Update `iospec.json`: rename connectors to `input_a`/`input_b`, change celltype to `packages` — Completed
   c. Regenerate `data_a.json` and `data_a.parquet` from `parquet_gen_spec_a.json` using `parquetgen` — Completed
   d. Regenerate `data_b.json` and `data_b.parquet` from `parquet_gen_spec_b.json` using `parquetgen` — Completed
   e. Run `go test ./...` to verify `simple_addition` passes — Completed

3. Correct `unary_negation` example files — Completed
   a. Update `unary_negation.cpl`: add `connectors` block declaring `input_a` and `input_b` with `celltype packages` — Completed
   b. Update `iospec.json`: rename connectors to `input_a`/`input_b`, change celltype to `packages`, change grain to `["Daily"]` — Completed
   c. Update `parquet_gen_spec_a.json` and `parquet_gen_spec_b.json`: expand `grain.values` to all 7 days, expand `values` to 14 entries — Completed
   d. Regenerate `data_a.*` and `data_b.*` from updated specs using `parquetgen` — Completed
   e. Update `expected_output.json`: expand to 14 rows with `Daily` grain, same calculated values for every day — Completed
   f. Run `go test ./...` to verify `unary_negation` passes — Completed

4. Correct `multiplication_precedence` example files — Completed
   a. Update `multiplication_precedence.cpl`: remove unused `measures` block, rename connector to `input_c` with `celltype packages`, change metric input to `metric input_c` — Completed
   b. Update `iospec.json`: rename connectors to `input_a`/`input_b`/`input_c`, change celltype to `packages`, change grain to `["Daily"]` — Completed
   c. Update all `parquet_gen_spec_*.json`: expand to all 7 days with 14 values each — Completed
   d. Regenerate all `data_*.*` files from updated specs using `parquetgen` — Completed
   e. Update `expected_output.json`: expand to 14 rows with `Daily` grain — Completed
   f. Run `go test ./...` to verify `multiplication_precedence` passes — Completed

5. Correct `parenthesized` example files — Completed
   a. Update `parenthesized.cpl`: remove unused `measures` block, rename connectors to `input_c`/`input_d` with `celltype packages`, change metric inputs to `metric input_c`/`metric input_d` — Completed
   b. Update `iospec.json`: rename connectors to `input_a` through `input_d`, change celltype to `packages`, change grain to `["Daily"]` — Completed
   c. Update all `parquet_gen_spec_*.json`: expand to all 7 days with 14 values each — Completed
   d. Regenerate all `data_*.*` files from updated specs using `parquetgen` — Completed
   e. Update `expected_output.json`: expand to 14 rows with `Daily` grain — Completed
   f. Run `go test ./...` to verify `parenthesized` passes — Completed

6. Correct `complex_expression` example files — Completed
   a. Update `complex_expression.cpl`: remove unused `measures` block, rename connectors to `input_c`/`input_d`/`input_e` with `celltype packages`, change metric inputs accordingly — Completed
   b. Update `iospec.json`: rename connectors to `input_a` through `input_e`, change celltype to `packages`, change grain to `["Daily"]` — Completed
   c. Update all `parquet_gen_spec_*.json`: expand to all 7 days with 14 values each — Completed
   d. Regenerate all `data_*.*` files from updated specs using `parquetgen` — Completed
   e. Update `expected_output.json`: expand to 14 rows with `Daily` grain — Completed
   f. Run `go test ./...` to verify `complex_expression` passes — Completed

7. Apply minor `subtraction_division` fix and run full verification — Completed
   a. Update `subtraction_division/iospec.json`: change output celltype from `package` to `packages` — Completed
   b. Run `go test ./...` in the `metrics_arithmetic` package to verify all examples pass end-to-end — Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. Task 1 updates the Go test expectations and supporting code first (Phase A: Structural Definition), defining the target pattern that all examples must match. Tasks 2–6 then correct each example's CPL, JSON, and data files to satisfy those expectations (Phase B: Logical Tests, Phase C: Implementation). Task 7 performs the final full-suite verification. The test-and-implement cycle is integrated within each task rather than broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- All `parquet_gen_spec_*.json` files must contain exactly 14 entries in their `values` array when fully specified (2 regions × 7 days).
- Connector names must be `input_a`, `input_b`, `input_c`, `input_d`, `input_e` across all examples to match the reference pattern.
- Grain must be `["Daily"]` in all `iospec.json` and `expected_output.json` files.
- Celltype must be `packages` (the measure name from `core.cpl`) in all connector declarations and output specs.
- No example `.cpl` file may declare unused measures or types.
- `parquetgen` must be used to regenerate all `.parquet` and `.json` data files from updated specs.

## Intent

The primary intent is to bring consistency to the `metrics_arithmetic` examples by making them all follow the same reference pattern. This eliminates confusion about where connectors are declared, ensures all examples use the same `Daily` grain with full 7-day coverage, and removes unused declarations that clutter the CPL files. The corrected examples will serve as reliable reference material for understanding how arithmetic metrics are defined and verified in the Chronocone system.

## Acceptance Criteria

1. All six `metrics_arithmetic` example subdirectories contain `.cpl` files that declare their own connectors with `celltype packages` and no unused measures.
2. All `iospec.json` files use connector names `input_a`/`input_b` (and `input_c`/`input_d`/`input_e` where needed), celltype `packages`, and grain `["Daily"]`.
3. All `parquet_gen_spec_*.json` files specify all 7 days with 14 values each.
4. All `data_*.json` and `data_*.parquet` files are regenerated from the updated specs and contain 14 rows.
5. All `expected_output.json` files contain 14 rows with `Daily` grain and correct calculated values.
6. `subtraction_division/iospec.json` output celltype is `packages`.
7. `example_definitions.go`, `iospec_definitions.go`, and `iospec_examples_test.go` reflect the corrected patterns.
8. `go test ./...` in the `metrics_arithmetic` package passes for all examples.

## Notes

- The `subtraction_division` example is the reference pattern; it requires only a minor celltype fix in its `iospec.json`.
- The `values` arrays in expanded specs should repeat the existing Monday-only values for all 7 days. For example, if the old spec had `values: [10.0, 20.0]`, the expanded spec should have `values: [10.0, 10.0, 10.0, 10.0, 10.0, 10.0, 10.0, 20.0, 20.0, 20.0, 20.0, 20.0, 20.0, 20.0]` (7 days × 2 regions).
- Run `parquetgen` with: `parquetgen -input=parquet_gen_spec_<letter>.json -output=data_<letter>.parquet -json_output=data_<letter>.json`
