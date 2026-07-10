# Story206: Fix Measure/ConceptValue Confusion

## Goal

Eliminate the naming confusion between grain measure names (e.g., "Daily") and their underlying concept names (e.g., "DayOfWeek") throughout the code generation and runtime pipeline. Ensure measure names are preserved through the entire process until tensors are generated, and that the code tracks which tensor dimensions correspond to grain measures so the relationship to the underlying concept is not lost.

## References

- [Fix Measure/ConceptValue Confusion Plan](../plans/fix_measure_conceptvalue_confusion_plan.md)
- [synthesis/synthesize.go](../../../src/chronocone_planning_language/src/main/go/synthesis/synthesize.go)
- [synthesis/synthesize_test.go](../../../src/chronocone_planning_language/src/main/go/synthesis/synthesize_test.go)
- [connector/connector_metadata.go](../../../src/chronocone_planning_language/src/main/go/connector/connector_metadata.go)
- [io/parquet_reader.go](../../../src/chronocone_planning_language/src/main/go/io/parquet_reader.go)
- [io/write_metrics_parquet.go](../../../src/chronocone_planning_language/src/main/go/io/write_metrics_parquet.go)
- [connector/generate_connector_loaders.go](../../../src/chronocone_planning_language/src/main/go/connector/generate_connector_loaders.go)
- [io/generate_parquet.go](../../../src/chronocone_planning_language/src/main/go/io/generate_parquet.go)

## Dependencies

- [Story203_fix-synthesis-bugs-arithmetic-metrics.md](Story203_fix-synthesis-bugs-arithmetic-metrics.md) - Done
- [Story204_fix-nested-binary-op-codegen.md](Story204_fix-nested-binary-op-codegen.md) - Done

Both dependencies must be complete because they establish the working synthesis and code generation pipeline that this story modifies.

## Dependent Stories

- [Story202_cpl-metrics-arithmetic-examples-revamp.md](Story202_cpl-metrics-arithmetic-examples-revamp.md) - Not Started

## Tasks

1. Add grain measure entries to `BuildConceptValuesMap` in synthesis package - Completed
   a. Write tests for `BuildConceptValuesMap` verifying grain measure names appear as keys when `UnitName` is set on grain measures - Completed
   b. Modify `BuildConceptValuesMap()` to walk connector grain measures and duplicate concept values under the measure name - Completed
   c. Verify existing synthesis tests still pass after the change - Completed

2. Remove grain measure remapping from `GenerateConnectorLoaders` in connector package - Completed
   a. Write tests verifying generated `loadParquetConnector` code does not contain grain measure remapping in `columnToConcept` - Completed
   b. Remove the grain mapping loop from `columnToConcept` construction in `GenerateConnectorLoaders` - Completed
   c. Verify generated code for examples no longer includes `metadata.GrainToConcept` loop - Completed

3. Update `outputSpec()` helper in `iospec_definitions.go` - Completed
   a. Modify `outputSpec()` to accept a `grain []string` parameter instead of hardcoding `"Day"` - Completed
   b. Update all call sites in metrics arithmetic examples to pass the correct grain measure name - Completed
   c. Verify subtraction_division and simple_addition examples use the correct grain measure name - Completed

4. Verify parquet I/O compatibility and run end-to-end validation - Completed
   a. Verify `ReadParquetToTensor` resolves grain measure names via identity mapping and `conceptValues` entries - Completed
   b. Verify `WriteMetricsToParquet` validates dimension names and grain names against `conceptValues` correctly - Completed
   c. Verify `generateOutputWritingCode` uses output grain measure names that match tensor dimension names - Completed
   d. Run `bash shell_scripts/compile_and_run.sh subtraction_division` and confirm compilation, execution, and output parquet validation pass - Completed
   e. Run `bash shell_scripts/compile_and_run.sh simple_addition` and confirm it still compiles and runs correctly - Completed
   f. Run full test suite: `cd src/chronocone_planning_language && go test ./...` and confirm all tests pass - Completed

## Constraints

- The `GrainToConcept` field in `ConnectorMetadata` must be retained for future use (display, debugging, metadata queries) even though it is no longer used for `columnToConcept` mapping.
- The `conceptValues` map may contain duplicate value slices under both measure and concept names; this duplication is intentional and preferred over adding a separate mapping layer.
- Generated code must remain valid Go that compiles without syntax errors.
- No changes to `ReadParquetToTensor` or `WriteMetricsToParquet` logic are expected; compatibility must be verified only.

## Intent

The generated code and runtime should consistently use grain measure names (e.g., "Daily") as tensor dimension names rather than remapping them to underlying concept names (e.g., "DayOfWeek"). This preserves the CPL language semantics where connectors declare grain using measure names, and ensures the output parquet grain field matches the tensor dimensions.

## Acceptance Criteria

1. The generated `conceptValues` map contains entries for grain measure names that resolve to the same values as their underlying concept names.
2. The `columnToConcept` map in generated `loadParquetConnector` does NOT remap grain measure names to concept names (identity mapping only).
3. Tensor dimension names use measure names throughout (no concept names appear in place of grain measures).
4. The generated output `Grain` field uses measure names that match tensor dimension names.
5. `WriteMetricsToParquet` passes validation (dimension names found in `conceptValues`, grain names found in `dimensionNames`).
6. `ReadParquetToTensor` passes validation (dimension names found in `conceptValues`).
7. All existing tests pass (`go test ./...` in the `src/chronocone_planning_language` module).
8. The `compile_and_run.sh subtraction_division` script succeeds end-to-end.

## Open Questions

None.

## Known Issues

- Current generated code remaps grain measure names to concept names in `columnToConcept`, causing dimension name mismatches between tensor generation and parquet I/O.
- The `outputSpec()` helper in `iospec_definitions.go` hardcodes `Grain: []string{"Day"}` which is incorrect for examples using the "Daily" measure.

## Additional Context

This story complements the fixes in Story203 and Story204 by resolving the remaining naming inconsistency in the synthesis pipeline. The duplication of concept values under measure names in `conceptValues` is the simplest approach and avoids adding a separate `dimensionToConceptValues` mapping layer.
