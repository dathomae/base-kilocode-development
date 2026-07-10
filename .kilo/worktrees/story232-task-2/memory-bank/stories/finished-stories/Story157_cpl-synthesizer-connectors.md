# Story157: CPL Synthesizer Connectors

## Goal

Implement the synthesis phase for basic connector declarations. This generates Go code that uses the `go-arrow` library to read data from a parquet file specified in the IO Specification, mapping the parquet columns to the connector's dimensions and celltype. This story is part of [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md).

## References

- [ERO004: Connectors Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO004_connectors-synthesis-phase.md)
- [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md)

## Dependencies

- [Story156_cpl-analyzer-connectors.md](finished-stories/Story156_cpl-analyzer-connectors.md) - Done

## Dependent Stories

None

## Tasks

1. Define `IOSpecification` and related structs - Completed
   a. Write tests for parsing the IO specification JSON into the `IOSpecification` struct - Completed
   b. Implement the `IOSpecification` struct and related parsing logic - Completed

2. Define basic `ChronoconeComputationController` and `ChronoconeTensor` structs - Completed
   a. Write tests for the basic allocation logic of `ChronoconeTensor` - Completed
   b. Implement the `ChronoconeComputationController` and `ChronoconeTensor` structs with basic allocation logic - Completed

3. Define `ConnectorSynthesisInfo` struct - Completed
   a. Write tests for the creation and validation of `ConnectorSynthesisInfo` - Completed
   b. Implement the `ConnectorSynthesisInfo` struct - Completed

4. Create `CILAbstractTransformer` struct and iteration logic - Completed
   a. Write tests for `CILAbstractTransformer` to ensure it correctly iterates through the `CILTrees` map - Completed
   b. Implement the `CILAbstractTransformer` struct and its iteration logic - Completed

5. Update `CILAbstractTransformer` to generate connector metadata - Completed
   a. Write tests to verify the generation of the connector metadata map from the `CILTrees` map - Completed
   b. Implement the logic in `CILAbstractTransformer` to generate the connector metadata map - Completed

6. Implement `GenerateConnectorLoaders` function - Completed
   a. Write tests for `GenerateConnectorLoaders` to ensure it correctly orchestrates the generation of individual loaders - Completed
   b. Implement the `GenerateConnectorLoaders` function - Completed

7. Implement `GenerateSingleConnectorLoader` function - Completed
   a. Write tests for `GenerateSingleConnectorLoader` to verify it generates correct Go code for loading a specific connector using `go-arrow` - Completed
   b. Implement the `GenerateSingleConnectorLoader` function - Completed

8. Update main function generation - Completed
   a. Write tests to ensure the generated main function correctly dispatches to loaders based on the IO specification - Completed
   b. Update the main function generation logic to include the dispatch mechanism - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- The generated Go code must correctly use the `go-arrow` library to read parquet files.
- The generated code must accurately map parquet columns to the connector's dimensions and celltype.

## Intent

- **Task 1**: To parse and represent the IO specification JSON, which defines the data sources for connectors.
- **Task 2**: To provide the foundational runtime structures for managing computations and tensor data.
- **Task 3**: To hold the necessary information for synthesizing a specific connector.
- **Task 4 & 5**: To transform the CIL representation of connectors into synthesis-ready metadata.
- **Task 6 & 7**: To generate the actual Go code responsible for loading data from parquet files into tensors.
- **Task 8**: To integrate the generated loaders into the main execution flow of the synthesized program.

## Acceptance Criteria

- The synthesizer successfully generates Go code for valid connector declarations.
- The generated Go code compiles without errors.
- The generated Go code correctly reads data from a specified parquet file and populates a `ChronoconeTensor`.
- All unit tests pass.
