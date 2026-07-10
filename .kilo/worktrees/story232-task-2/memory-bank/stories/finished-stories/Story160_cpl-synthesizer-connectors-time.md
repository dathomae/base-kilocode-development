# Story160: CPL Synthesizer Connectors Time

## Goal

Implement the synthesis phase for connector declarations with time handling. This generates Go code that maps the parquet time column to the connector's grain. This story is part of [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md).

## References

- [ERO004: Connectors Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO004_connectors-synthesis-phase.md)
- [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md)

## Dependencies

- [Story159_cpl-analyzer-connectors-time.md](Story159_cpl-analyzer-connectors-time.md) - Done
- [Story157_cpl-synthesizer-connectors.md](Story157_cpl-synthesizer-connectors.md) - Done

## Dependent Stories

None

## Tasks

1. Update `ConnectorSynthesisInfo` - Done
   a. Write tests to verify `ConnectorSynthesisInfo` correctly stores the grain information - Done
   b. Update the `ConnectorSynthesisInfo` struct to include a `Grain` field - Done

2. Update `GenerateSingleConnectorLoader` - Done
   a. Write tests for `GenerateSingleConnectorLoader` to verify it generates correct Go code for mapping the parquet time column to the grain - Done
   b. Update the `GenerateSingleConnectorLoader` function to include the time column mapping logic - Done

3. Update tensor allocation - Done
   a. Write tests to ensure the generated tensor allocation code correctly includes the time dimension - Done
   b. Update the tensor allocation logic to account for the time dimension based on the grain - Done

4. Update flat index calculation - Done
   a. Write tests to verify the generated flat index calculation code correctly incorporates the time dimension - Done
   b. Update the flat index calculation logic to include the time dimension - Done

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- The generated Go code must correctly map the parquet time column to the connector's grain.
- The generated tensor allocation and index calculation code must accurately reflect the addition of the time dimension.
- Time dimension handling must remain consistent with the updated `Instant` and `Interval` type definitions (see [ERO002: Types Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO002_types-synthesis-phase.md)).

## Intent

- **Task 1**: To update the synthesis metadata to include the grain information.
- **Task 2**: To generate the Go code responsible for reading the time column from the parquet file.
- **Task 3**: To ensure the generated tensor has the correct dimensions, including time.
- **Task 4**: To ensure data is loaded into the correct locations within the tensor, accounting for the time dimension.

## Acceptance Criteria

- The synthesizer successfully generates Go code for valid connector declarations with a `grain`.
- The generated Go code compiles without errors.
- The generated Go code correctly reads data, including the time column, from a specified parquet file and populates a `ChronoconeTensor` with the correct dimensions.
- All unit tests pass.
