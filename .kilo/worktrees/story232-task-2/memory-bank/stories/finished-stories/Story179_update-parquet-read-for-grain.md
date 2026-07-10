# Story179: Update Parquet Reader for Grain Measures

## Goal

Fix the architectural mismatch in [`ReadParquetToTensor`](../../src/chronocone_planning_language/src/main/go/io/parquet_reader.go:81) where parquet column names are assumed to be identical to `conceptValues` keys. This assumption fails for grain measures (e.g., column `YearGrain` maps to concept `Year`). Introduce an explicit `columnToConcept` mapping, propagate it through the synthesis and connector loader pipeline, and ensure the end-to-end `random_data_with_grain` example regenerates and compiles successfully.

## References

- [Plan: Update Parquet Reader to Support Grain Measure Columns](../plans/update_parquet_read_for_grain_plan.md)
- [Architecture: CPL General Implementation](../specifications/architecture/cpl/general-implementation.md)
- [Design: CPL Concepts Definitions](../specifications/design/cpl/concepts-definitions.md)
- [Logical TDD Lifecycle](../concepts.md)

## Dependencies

- [Story177_update-parquet-gen-grain.md](../stories/finished-stories/Story177_update-parquet-gen-grain.md) - Done — updated the parquet generator to emit grain columns, producing the test fixture this story reads.
- [Story160_cpl-synthesizer-connectors-time.md](Story160_cpl-synthesizer-connectors-time.md) - Done — established connector grain synthesis in the generated loader.
- [Story157_cpl-synthesizer-connectors.md](../stories/finished-stories/Story157_cpl-synthesizer-connectors.md) - Done — introduced connector metadata and loader generation.

## Dependent Stories

None

## Tasks

1. Update `ReadParquetToTensor` signature and implementation with tests - Completed
   a. Write tests `TestReadParquetToTensor_GrainWithColumnToConcept` and `TestReadParquetToTensor_MissingColumnToConceptMapping` - Completed
   b. Add `columnToConcept map[string]string` parameter to `ReadParquetToTensor`; use it to resolve concept names for `conceptValues` and `valueIndices` lookups while preserving original names for parquet column access - Completed
   c. Update existing `parquet_reader_test.go` calls to pass `nil` for `columnToConcept` (backward-compatibility identity path) - Completed
   d. Run `go test ./src/main/go/io/...` and verify all tests pass - Completed

2. Update `ConnectorMetadata` with `GrainToConcept` and tests - Completed
   a. Write tests `TestNewConnectorMetadataGrainToConcept` and `TestNewConnectorMetadataGrainToConceptImmutability` - Completed
   b. Add `GrainToConcept map[string]string` field to `ConnectorMetadata` and update `NewConnectorMetadata` to accept and deep-copy the map - Completed
   c. Update existing `connector_metadata_test.go` calls to pass `nil` for `grainToConcept` - Completed
   d. Run `go test ./src/main/go/connector/...` and verify metadata tests pass - Completed

3. Update connector loader generated code and tests - Completed
   a. Write/update tests in `connector_loader_test.go` asserting that `GenerateLoadParquetConnectorFunction` emits `columnToConcept` construction and the four-argument `ReadParquetToTensor` call - Completed
   b. Update `GenerateLoadParquetConnectorFunction` in `connector_loader.go` to build a `columnToConcept` map (identity for dimensions, `metadata.GrainToConcept` for grain measures) and pass it to `ReadParquetToTensor` - Completed
   c. Run `go test ./src/main/go/connector/...` and verify loader tests pass - Completed

4. Update CIL transformer to propagate grain-to-concept mapping and tests - Completed
   a. Write tests `TestRegisterConnector_WithGrainToConcept` and `TestGenerateConnectorMetadataMap_WithGrainToConcept` - Completed
   b. Update `CILAbstractTransformer.RegisterConnector` signature to accept `grainToConcept map[string]string` and store it in `ConnectorMetadata` - Completed
   c. Update `GenerateConnectorMetadataMap` to emit deterministic `GrainToConcept` map literals when non-empty - Completed
   d. Update existing `cil_transformer_test.go` calls to pass `nil` for `grainToConcept` - Completed
   e. Run `go test ./src/main/go/cil/...` and verify transformer tests pass - Completed

5. Update synthesis pipeline to compute grain-to-concept mapping and tests - Completed
   a. Write/update tests in `synthesize_test.go` asserting that `BuildTransformerFromAnalysis` populates `GrainToConcept` from `AnalysisMeasureDefinition.UnitName` - Completed
   b. Update `BuildTransformerFromAnalysis` in `synthesize.go` to compute `grainToConcept` for each connector by mapping grain measure names to their `UnitName` concept names - Completed
   c. Pass the computed `grainToConcept` to `transformer.RegisterConnector` - Completed
   d. Run `go test ./src/main/go/synthesis/...` and verify synthesis tests pass - Completed

6. Regenerate example and run full test suite - Completed
   a. Regenerate `random_data_with_grain` example via the synthesizer CLI - Completed
   b. Verify generated `connectorMetadata` contains `GrainToConcept` entries and `loadParquetConnector` builds `columnToConcept` - Completed
   c. Compile generated code to confirm no errors - Completed
   d. Run full test suite across `io`, `connector`, `cil`, and `synthesis` packages - Completed

## Constraints

- `ReadParquetToTensor` must remain backward compatible: when `columnToConcept` is `nil`, column names map to themselves (identity).
- `GrainToConcept` maps in `ConnectorMetadata` must be deep-copied on construction to preserve immutability.
- Generated code in `GenerateConnectorMetadataMap` must emit deterministic output; sort keys when iterating over `GrainToConcept`.
- All changes must follow Go conventions in the existing codebase.
- The test-and-implement cycle must be integrated within each task per the Logical TDD Lifecycle.

## Intent

The goal of each task is to decouple parquet column names from concept lookup keys so that grain measures (whose column names differ from their underlying concept names) can be read correctly into tensors. This requires propagating the measure-to-concept mapping from analysis through synthesis, CIL generation, and finally into the connector loader runtime.

## Acceptance Criteria

- `ReadParquetToTensor` accepts a `columnToConcept map[string]string` parameter and uses it for all `conceptValues` lookups.
- When `columnToConcept` is `nil`, `ReadParquetToTensor` falls back to identity mapping.
- All existing `parquet_reader_test.go` tests pass with `nil` `columnToConcept`.
- New tests cover grain column reading with explicit `columnToConcept` mapping and missing-mapping error cases.
- `ConnectorMetadata` contains a `GrainToConcept` field that is deep-copied on creation.
- `NewConnectorMetadata` accepts and stores `grainToConcept`.
- Generated `loadParquetConnector` builds a `columnToConcept` map from dimensions (identity) and `metadata.GrainToConcept`.
- `CILAbstractTransformer.RegisterConnector` accepts and stores `grainToConcept`.
- `GenerateConnectorMetadataMap` emits `GrainToConcept` map literals into generated code when non-empty.
- `BuildTransformerFromAnalysis` computes `grainToConcept` from `AnalysisMeasureDefinition.UnitName`.
- All existing tests in `connector`, `cil`, and `synthesis` packages are updated and pass.
- The `random_data_with_grain` example regenerates successfully and the generated code compiles.
