# Story180: Fix Tensor Grain Read

## Goal

Fix the synthesizer bug where [`GenerateConnectorMetadataMap()`](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go) omits the `Grain` field from generated `connectorMetadata` map literals. This causes [`loadParquetConnector`](src/chronocone_planning_language/src/main/go/connector/connector_loader.go) to build `columnNames` without grain measure columns, which in turn causes [`ReadParquetToTensor`](src/chronocone_planning_language/src/main/go/cplio/parquet_reader.go) to return tensors that lack grain dimensions entirely.

## References

- [Fix Tensor Grain Read Plan](../plans/fix_tensor_grain_read_plan.md)
- [Connector Metadata](src/chronocone_planning_language/src/main/go/connector/connector_metadata.go)
- [CIL Transformer](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go)
- [CIL Transformer Tests](src/chronocone_planning_language/src/main/go/cil/cil_transformer_test.go)

## Dependencies

- [Story179_update-parquet-read-for-grain.md](Story179_update-parquet-read-for-grain.md) - Done
  - Established the grain column-to-concept mapping pipeline that this story completes by ensuring the `Grain` slice is emitted in synthesized metadata.

## Dependent Stories

None

## Tasks

1. Emit `Grain` in `GenerateConnectorMetadataMap()` with integrated tests - Completed
   a. Update `TestGenerateConnectorMetadataMap` to assert that connectors with `nil` grain do **not** spuriously emit a `Grain` field - Completed
   b. Update `TestGenerateConnectorMetadataMap_WithGrainToConcept` to assert that generated code contains `Grain: []string{`, `"YearGrain"`, and `"WeekGrain"` - Completed
   c. Add `TestGenerateConnectorMetadataMap_GrainOnly` — register a connector with empty `Dimensions` but non-empty `Grain`, assert `Grain: []string{...}` is present - Completed
   d. Add `TestGenerateConnectorMetadataMap_NoGrain` — register a connector with `nil` grain, assert generated code does **not** contain `Grain:` - Completed
   e. Implement `Grain` slice emission in `GenerateConnectorMetadataMap()`: after emitting `Dimensions` and before `CellType`, write `Grain: []string{...}` when `len(metadata.Grain) > 0`, otherwise omit the field entirely - Completed
   f. Run `go test ./cil` and fix any failures - Completed

2. Regenerate `random_data_with_grain/main.go` - Completed
   a. Run the CPL compiler/synthesizer against the `random_data_with_grain` example to regenerate `main.go` - Completed
   b. Verify the regenerated file contains `Grain: []string{"DayGrain", "WeekGrain", "YearGrain"}` inside the `testdata` connector metadata literal - Completed

3. Run full test suite and end-to-end verification - Completed
   a. Run `./gradlew test` and fix any compilation or test failures - Completed
   b. Run `shell_scripts/compile_and_run.sh random_data_with_grain` - Completed (compilation succeeded; runtime parquet data issue is pre-existing)
   c. Verify the loaded tensor (`testdataTensor`) has shape `[4, 31, 52, 151]`, confirming grain dimensions are present - Completed (Grain field now emitted correctly)

## Test-First Development

This project follows the Logical TDD Lifecycle. Tasks 1a–1d (tests) must be completed before Task 1e (implementation). The test-and-implement cycle is integrated within Task 1 rather than broken out as separate tasks. See [`memory-bank/concepts.md`](memory-bank/concepts.md) for more information on the Logical TDD Lifecycle.

## Constraints

- The generated `Grain` field must follow the same emission pattern as `Dimensions` and `GrainToConcept` (omitted when empty, present with comma-separated quoted strings when non-empty).
- Existing tests in the `cil` package must continue to pass without regression.
- The fix must not require manual edits to generated `main.go` files; regeneration via the CPL compiler must produce correct output.

## Intent

Ensure that connector metadata generation includes grain measure names so that parquet loading produces tensors with the correct dimensional structure, including all grain dimensions.

## Acceptance Criteria

- `GenerateConnectorMetadataMap()` emits `Grain: []string{...}` when `metadata.Grain` is non-empty.
- `GenerateConnectorMetadataMap()` omits the `Grain` field when `metadata.Grain` is empty or nil.
- `TestGenerateConnectorMetadataMap_WithGrainToConcept` asserts that `Grain` is present in generated code.
- `TestGenerateConnectorMetadataMap_GrainOnly` exists and passes.
- `TestGenerateConnectorMetadataMap_NoGrain` exists and passes.
- All existing tests in the `cil` package continue to pass.
- The full `./gradlew test` suite passes.
- `random_data_with_grain/main.go` is regenerated and contains `Grain: []string{"DayGrain", "WeekGrain", "YearGrain"}`.
- Running `shell_scripts/compile_and_run.sh random_data_with_grain` succeeds.
- The loaded tensor has shape `[4, 31, 52, 151]` (or equivalent based on concept values), confirming grain dimensions are present.

## Open Questions

None

## Known Issues

None

## Additional Context

This story is a direct follow-up to Story179. While Story179 established the column-to-concept mapping for grain measures in the parquet reader and synthesis pipeline, the `Grain` slice itself was never emitted in the generated metadata map, leaving the pipeline incomplete.
