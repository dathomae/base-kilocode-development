# Story177: Update Parquet Generator to Support Hierarchical Grain Measures

## Goal

Extend the `parquetgen` CLI tool and the `io.GenerateParquetFile` library so that test parquet files can include both **dimension columns** and **grain measure columns**. Grain measure values are specified hierarchically in the JSON spec, with array position defining nesting level (outermost = highest in hierarchy, varies slowest; innermost = lowest, varies fastest). Create a new example based on `examples/with_data/random_data_no_time` that demonstrates a connector with a grain of `YearGrain, WeekGrain, DayGrain` derived from the standard concepts hierarchy.

## References

- [Plan: Update Parquet Generator to Support Hierarchical Grain Measures](../../plans/update_parquet_gen_plan.md)
- [Requirements: CPL Connectors](../../specifications/requirements/cpl/connectors.md)
- [Requirements: CPL IO Specifications](../../specifications/requirements/cpl/io_specs.md)
- [Design: ERO004 Connectors Synthesis Phase](../../specifications/design/cpl/go-synthesis-target/eros/ERO004_connectors-synthesis-phase.md)

## Dependencies

- [Story161_parquet-test-utility.md](Story161_parquet-test-utility.md) - Done
- [Story176_cpl-connectors-multiple-grain-dimensions.md](Story176_cpl-connectors-multiple-grain-dimensions.md) - Done

## Dependent Stories

None

## Tasks

1. Update `io.GenerateParquetFile` with grain-aware signature and tests - Completed
   a. Add `GrainMeasure` struct definition - Completed
   b. Write tests for grain functionality (schema, cartesian product, ordering, grain-only) - Completed
   c. Update `GenerateParquetFile` signature to accept `grain []GrainMeasure` - Completed
   d. Implement grain column generation in parquet output - Completed
   e. Update existing tests to use new signature with `nil` grain - Completed
   f. Verify all `io` package tests pass - Completed

2. Update `parquetgen` CLI to parse new JSON format and tests - Completed
   a. Define `ParquetGenSpec` struct with `dimensions` and `grain` fields - Completed
   b. Write tests for `parseSpec` function - Completed
   c. Replace `parseDimensions` with `parseSpec` that unmarshals the new format - Completed
   d. Update `run` to pass both dimensions and grain to `io.GenerateParquetFile` - Completed
   e. Verify CLI tests pass - Completed

3. Update existing example spec to new format - Completed
   a. Convert `examples/with_data/random_data_no_time/parquet_gen_spec.json` to `dimensions`/`grain` structure - Completed
   b. Verify the updated spec generates a valid parquet file - Completed

4. Create new example with hierarchical grain - Completed
   a. Create `examples/with_data/random_data_with_grain/test.cpl` with Region dimension and YearGrain/WeekGrain/DayGrain grain - Completed
   b. Create `examples/with_data/random_data_with_grain/parquet_gen_spec.json` - Completed
   c. Create `examples/with_data/random_data_with_grain/iospec.json` - Completed
   d. Run updated `parquetgen` tool to generate `testdata.parquet` - Completed
   e. Verify generated parquet has columns: Region, YearGrain, WeekGrain, DayGrain, Value - Completed

5. Run full test suite and fix regressions - Completed
   a. Execute `./gradlew test` in `src/chronocone_planning_language` - Completed
   b. Fix any compilation or test failures - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks.

For Task 1, write the grain-specific tests (correct schema, cartesian product row count, grain column ordering, grain-only generation) before modifying `GenerateParquetFile`. The failing tests define the expected behavior. Then update the signature, implement the grain column logic, and update existing tests to pass `nil` for grain.

For Task 2, write tests for `parseSpec` before replacing `parseDimensions`. The tests should verify correct unmarshaling of both `dimensions` and `grain` fields, and proper conversion to `[]io.GrainMeasure`.

## Constraints

- There is **no backward compatibility** with the old flat JSON format. Every input JSON must use the `dimensions`/`grain` structure.
- Grain measure columns must appear in the parquet file in the order specified by the `grain` array (preserving hierarchy).
- Dimension columns continue to be sorted alphabetically by name.
- When `grain` is `nil` or empty, behavior must be equivalent to the old dimensions-only logic.
- Row count equals the Cartesian product of all dimension value counts and all grain measure value counts.
- The new example must use grain measures (`YearGrain`, `WeekGrain`, `DayGrain`) based on the standard concepts hierarchy.

## Intent

The parquet generator was originally designed for dimension-only connectors. Now that the CPL compiler supports multi-grain connectors, the test data generator must also support grain measures as explicit columns. This ensures that synthesized connector loaders can read parquet files that match the full connector schema, including both dimensions and grain.

## Acceptance Criteria

- `io.GenerateParquetFile` accepts a `grain []GrainMeasure` parameter and generates parquet files with both dimension and grain measure columns.
- Grain columns appear in the order specified by the JSON `grain` array, after alphabetically sorted dimension columns.
- Row count equals the Cartesian product of all dimensions and all grain measures.
- `parquetgen` CLI parses the new JSON format (`dimensions` + `grain`) correctly.
- New tests cover schema correctness, Cartesian product row counts, grain ordering, and grain-only generation.
- All existing tests in the `io` package are updated to the new signature and continue to pass.
- Existing example `random_data_no_time/parquet_gen_spec.json` is updated to the new format.
- New example directory `random_data_with_grain` exists with `test.cpl`, `parquet_gen_spec.json`, `iospec.json`, and generated `testdata.parquet`.
- The new `test.cpl` uses grain measures based on `YearGrain`, `WeekGrain`, and `DayGrain`.
- The generated `testdata.parquet` has columns: `Region`, `YearGrain`, `WeekGrain`, `DayGrain`, `Value`.

## Open Questions

None

## Known Issues

None

## Additional Context

The existing `random_data_no_time` example uses `dimensions [Year, Region] grain Day`. When updating its `parquet_gen_spec.json`, `Year` should be moved into the `grain` array to align with the new tool behavior, while `Region` remains as a dimension. The example's original intent is preserved.
