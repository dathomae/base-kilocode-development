# Story181: Reject Insufficient Parquet Data

## Goal

Modify the synthesized program's parquet reading logic so that it terminates with a descriptive error when the input parquet file does not exactly match the structure described by the CPL connector definitions. The input must contain precisely the expected dimensions and grains, with exactly one row per valid combination and no extraneous rows, columns, or duplicates. The system must never manufacture data by filling missing values with zeros or any other default.

## References

- [Plan](../plans/reject_insufficient_data_plan.md)
- [Concepts - No Manufactured Data](../concepts.md)
- [Concepts - Logical TDD Lifecycle](../concepts.md)

## Dependencies

- [Story157_cpl-synthesizer-connectors.md](finished-stories/Story157_cpl-synthesizer-connectors.md) - Done
- [Story165_cpl-generated-code-compilation.md](finished-stories/Story165_cpl-generated-code-compilation.md) - Done
- [Story179_update-parquet-read-for-grain.md](finished-stories/Story179_update-parquet-read-for-grain.md) - Done
- [Story180_fix-tensor-grain-read.md](finished-stories/Story180_fix-tensor-grain-read.md) - Done

## Dependent Stories

None

## Tasks

1. Add strict exact-match validation to `ReadParquetToTensor` - Done
   a. Structural definition: Add column-count validation, row-count validation, duplicate detection via `seen []bool` tracker, and missing-combination validation to `ReadParquetToTensor`. Replace the zero-initialized `data` slice with the `seen` tracker. Update the function docstring and algorithm comments to remove zero-default references and document the strict exact-match requirement. - Done
   b. Logical tests: Write failing tests for exact-match enforcement: `TestReadParquetToTensor_ExactMatchSucceeds`, `TestReadParquetToTensor_MissingDataReturnsError`, `TestReadParquetToTensor_DuplicateRowsReturnsError`, `TestReadParquetToTensor_ExtraRowsReturnsError`, `TestReadParquetToTensor_ExtraColumnsReturnsError`, `TestReadParquetToTensor_ExplicitZeroValueSucceeds`, and `TestReadParquetToTensor_ErrorListsMissingCombinations`. Run the new tests and confirm they fail against the current zero-filling implementation. - Done
   c. Implementation: Implement the validation logic and helper to reverse-compute dimension values from a flat index for error messages. Make all new tests pass. Run the full `io` package test suite and confirm all existing tests continue to pass. - Done

2. Verify integration with broader test suites - Done
   a. Run the `connector` package tests to ensure generated loader code still integrates correctly. - Done
   b. Run the `synthesis` package tests to ensure synthesized code generation is unaffected. - Done
   c. Run the compiler integration tests (`cpl_integration_test.go`) to verify end-to-end compilation behavior is unchanged. - Done

## Constraints

- The error message for missing data must describe: (a) the total number of expected rows, (b) the number actually found, and (c) at least the first few missing combinations to aid debugging.
- Explicit `0.0` values in the input file are valid as long as the row is part of the exact expected set.
- No extra columns are permitted beyond the dimensions plus a single `"Value"` column.
- The total number of rows must exactly equal the Cartesian product of the concept value counts for all dimensions and grains.
- This story follows the [Logical TDD Lifecycle](../concepts.md). Tests are written before implementation within each coding task.

## Intent

The intent is to enforce the [No Manufactured Data](../concepts.md) principle at the data loading boundary. By rejecting parquet files that do not exactly match the CPL connector specification, we prevent silent data corruption that could lead to incorrect planning results.

## Acceptance Criteria

- `ReadParquetToTensor` returns an error when the parquet file does not exactly match the CPL specification.
- Column exactness is enforced: extra columns produce an error listing unexpected column names.
- Row count exactness is enforced: incorrect row counts produce an error stating expected and actual counts.
- Combination exactness is enforced: missing combinations, duplicate combinations, and extraneous rows all produce errors.
- Error messages for missing data include expected count, found count, and specific missing dimension combinations.
- Explicit `0.0` values in the input file load successfully when all exact-match conditions are met.
- All docstrings and comments referencing zero-default behavior are removed or corrected.
- All new unit tests pass.
- All existing tests in the `io`, `connector`, `synthesis` packages and compiler integration tests continue to pass.

## Notes

- The current implementation in `src/chronocone_planning_language/src/main/go/io/parquet_reader.go` allocates a dense `[]float64` slice initialized to zero and writes values only for rows present in the parquet file. This behavior is documented in the function docstring as "Unspecified combinations default to 0.0" and must be changed to a strict error.
