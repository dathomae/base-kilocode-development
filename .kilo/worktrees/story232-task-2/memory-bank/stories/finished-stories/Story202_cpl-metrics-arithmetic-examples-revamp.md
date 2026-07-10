# Story202: CPL Metrics Arithmetic Examples Revamp

## Goal

Transform the `src/chronocone_planning_language/examples/metrics_arithmetic/` directory from isolated, non-synthesizable CPL stubs into fully synthesizable, runnable, and verifiable examples. A shared [`core.cpl`](src/chronocone_planning_language/examples/metrics_arithmetic/core.cpl) defines common infrastructure (concepts, types, units, measures, and base connectors), while each example lives in its own subdirectory with a metric-only CPL file, IO spec, generated input data, and an expected-output specification for automated verification. This story also delivers reusable CLI tooling: an extended `parquetgen` with fixed-value and JSON output support, a new `parquet_verify` utility, and synthesis changes to write all output metrics to a single parquet file with a `Metric` column.

## References

- [Plan](../plans/revamp_metric_test_examples_plan.md)
- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [Requirements: CPL Metrics Syntax](../specifications/requirements/cpl/metrics-syntax.md)
- [Requirements: CPL IO Specs](../specifications/requirements/cpl/io_specs.md)

## Dependencies

- [Story185_cpl-metrics-arithmetic-synthesize.md](Story185_cpl-metrics-arithmetic-synthesize.md) — Done
  - Arithmetic metric synthesis (VisitCILBinaryOp, VisitCILUnaryOp, VisitCILInputRef) must be working before examples can be end-to-end verified.
- [Story157_cpl-synthesizer-connectors.md](Story157_cpl-synthesizer-connectors.md) — Done
  - Connector-based parquet input reading must be synthesized for the examples to consume input data.

## Dependent Stories

None

## Tasks

1. Extend `parquetgen` CLI with fixed values and JSON output — Completed
   a. Write tests for `ParquetGenSpec` parsing with optional `Values` field and validation that length matches expected row count — Completed
   b. Write tests for JSON output generation format and content correctness — Completed
   c. Implement `Values []float64` field in `ParquetGenSpec`; when present, use fixed values instead of random generation — Completed
   d. Add `-json_output` flag to `parquetgen`; write dimension columns, grain columns, and Value as a JSON array of objects — Completed
   e. Verify all `parquetgen` tests pass — Completed

2. Create `parquet_verify` CLI utility — Completed
   a. Write tests for expected JSON parsing, schema verification, row matching, and value comparison with tolerance — Completed
   b. Implement `ExpectedOutput` and `ExpectedRow` structs with JSON parsing — Completed
   c. Implement parquet schema verification (Metric column, dimensions, grain, Value) — Completed
   d. Implement row matching by Metric, dimension values, and grain values — Completed
   e. Implement float64 value comparison within tolerance (`1e-9`) with clear diff output — Completed
   f. Define exit codes: `0` match, `1` differences, `2` error — Completed
   g. Verify all `parquet_verify` tests pass — Completed

3. Modify synthesis to write all metrics to a single parquet file with `Metric` column — Completed
   a. Write tests for `WriteMetricsToParquet` validating schema, row count, and Metric column population — Completed
   b. Implement `MetricTensorData` struct and `WriteMetricsToParquet` in `cplio` package — Completed
   c. Update `generateOutputWritingCode` in `synthesize.go` to collect all output metrics and call `WriteMetricsToParquet("output.parquet", ...)` — Completed
   d. Retain `WriteTensorToParquet` for backward compatibility — Completed
   e. Verify synthesis tests pass — Completed

4. Create `core.cpl` shared infrastructure and example subdirectories — Completed
   a. Create `core.cpl` with shared concepts, types, units, measures, and two base connectors (`data_a`, `data_b`) — Completed
   b. Create `parquet_gen_spec_a.json` and `parquet_gen_spec_b.json` for generating 2-row input data — Completed
   c. Create six subdirectories (`simple_addition`, `multiplication_precedence`, `subtraction_division`, `unary_negation`, `parenthesized`, `complex_expression`) — Completed
   d. Write metric-only CPL files for each example; for examples with >2 inputs, define additional measures and connectors inline — Completed
   e. Write `iospec.json` for each example mapping connectors to parquet files and declaring output metrics — Completed
   f. Write `expected_output.json` for each example describing the single output parquet with `Metric` column — Completed

5. Update `compile_and_run.sh` and supporting Go files — Completed
   a. Modify `compile_and_run.sh` to locate and pass `core.cpl` alongside example-specific CPL files — Completed
   b. Modify `compile_and_run.sh` to run `parquet_verify` after execution when `expected_output.json` exists — Completed
   c. Update `example_definitions.go` with subdirectory paths and `CoreCPL` constant — Completed
   d. Update `iospec_definitions.go` and `iospec_examples_test.go` for new IO spec structure and paths — Completed
   e. Update `metrics_arithmetic_test.go` to include `core.cpl` in compiler invocations — Completed
   f. Remove old flat `.cpl` and `_iospec.json` files after migration — Completed

6. Generate input data, update documentation, and perform end-to-end verification — Completed
   a. Generate `data_a.parquet`, `data_a.json`, `data_b.parquet`, `data_b.json` (and additional inputs for multi-input examples) using extended `parquetgen` with fixed values — Completed (42 files: 21 .parquet + 21 .json across 6 examples)
   b. Compute and populate `expected_output.json` values for each example based on fixed input values — Completed (values verified against fixed inputs for all 6 examples)
   c. Update `src/chronocone_planning_language/examples/README.md` with new structure, multi-file compilation, `parquetgen`, and `parquet_verify` usage — Completed
   d. Create a generic Go test that discovers and iterates over all subdirectories under `examples/with_data/metrics_arithmetic/` (excluding `core.cpl`), compiles each example with `core.cpl`, executes it, and verifies output with `parquet_verify` — Completed
   e. Run the generic directory-tree test and ensure all six examples pass end-to-end verification — Completed

   **Bug fixes applied during Task 6 verification:**
   - `core.cpl`: Removed duplicate `Daily` measure (already in standard definitions); removed `Day` concept/unit; updated grain references from `Day` to `DayOfWeek` to match standard `Daily = Interval(starts_at=Midnight, unit=DayOfWeek)`.
   - All 6 `iospec.json`: Updated `"grain": ["Day"]` → `"grain": ["DayOfWeek"]`.
   - All 6 `expected_output.json`: Updated `"grain": ["Day"]` → `"grain": ["DayOfWeek"]` and `"Day": "Monday"` → `"DayOfWeek": "Monday"`.
   - `cil_transformer.go` `GenerateConnectorMetadataMap()`: Fixed double-comma bug; added deterministic key ordering; use explicit `connector.ConnectorMetadata{...}` and consistent trailing commas.
   - `cil_transformer.go` `VisitCILBinaryOp`/`VisitCILUnaryOp`: Changed to emit `result = <op>(...)` instead of bare expression, so generated code uses `var result; result = ...` pattern.
   - `cil_transformer.go` `TransformMetric`: Changed non-literal path from `result := <expr>` to `var result; <expr>` pattern.
   - `cil_transformer_test.go`: Updated 4 test expectations to match new `var result`/`result =` pattern.
   - All CIL tests pass.

     **Resolved blocking issues (pre-existing synthesis bugs from Story 185):**
   1. ~~`GetTensor()` returns `*ChronoconeTensor[any]`~~ — **RESOLVED**: The synthesis now uses package-level tensor functions (e.g., `calculation.TensorAddFloat64`) that accept `*ChronoconeTensor[any]` directly, and the postfix table evaluator handles all tensor resolution internally without calling methods on the return type.
   2. ~~Metric input references emitted as `controller.GetMetricComputation("input_a").GetTensor()` (kind "metric") instead of `inputs["data_a"]` (kind "connector")~~ — **RESOLVED**: `resolveInputRefKind()` now correctly maps measure references to kind `"connector"`, and the generated `Compute()` function pre-registers all `inputs` entries as `MetricComputation` instances, ensuring the postfix evaluator resolves names correctly regardless of kind.

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks (Tasks 1, 2, and 3), tests are written before implementation to define the logical contract. The test-and-implement cycle is integrated within each task rather than broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- All examples must use deterministic (fixed) input values to ensure reproducible expected outputs.
- The CPL compiler accepts multiple `.cpl` files as positional arguments; `core.cpl` must always be passed alongside each example-specific file.
- Synthesis output writing must remain backward compatible: `WriteTensorToParquet` must be retained for other callers.
- `parquet_verify` must reuse the existing `cplio.ReadParquetFile` function rather than implementing a new parquet reader.
- The `parquetgen` JSON output format must match the structure used for manual verification and expected-output computation.

## Intent

The primary intent is to make the `metrics_arithmetic` examples fully runnable and verifiable, establishing a reusable pattern for future CPL example suites. By factoring shared infrastructure into `core.cpl`, each example file focuses on its metric expression. By creating standalone CLI tools (`parquetgen` extension, `parquet_verify`), the verification workflow becomes composable and usable in CI pipelines. By writing all outputs to a single parquet file, verification and maintenance are simplified.

## Acceptance Criteria

1. `core.cpl` exists in `examples/metrics_arithmetic/` and compiles successfully alongside each example subdirectory's CPL file.
2. Six subdirectories exist, each containing: a metric-only `.cpl` file, `iospec.json`, `parquet_gen_spec_*.json`, generated `.parquet` and `.json` data files, and `expected_output.json`.
3. `parquetgen` supports the `Values` field for fixed-value generation and the `-json_output` flag for JSON dump output.
4. `parquet_verify` reads a parquet file and an expected JSON file, compares schema, row count, row matching, and values within tolerance, and exits with code `0` on match, `1` on differences, `2` on error.
5. Synthesis writes all output metrics to a single `output.parquet` with a `Metric` column (string), dimension columns, grain columns, and a `Value` column (float64).
6. `compile_and_run.sh` includes `core.cpl` on the compiler command line and invokes `parquet_verify` after execution when `expected_output.json` is present.
7. A generic Go test discovers all subdirectories under `examples/with_data/metrics_arithmetic/`, compiles each with `core.cpl`, executes, and verifies output against `expected_output.json` using `parquet_verify` — all six examples pass.

## Notes

- For examples with more than two inputs (`parenthesized` with 4 inputs, `complex_expression` with 5 inputs), the example-specific CPL file must declare additional measures (`c`, `d`, `e`) and connectors (`data_c`, `data_d`, `data_e`) since `core.cpl` only defines `input_a`/`input_b` and `data_a`/`data_b`.
- The `expected_output.json` `value` fields should be computed deterministically from the fixed input values; the JSON dumps from `parquetgen` make this computation straightforward without reading parquet programmatically.
