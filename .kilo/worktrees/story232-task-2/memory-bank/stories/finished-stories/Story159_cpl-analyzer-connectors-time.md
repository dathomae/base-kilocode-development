# Story159: CPL Analyzer Connectors Time

## Goal

Implement the analysis phase for connector declarations with time handling. This validates that the grain refers to an existing measure, and that the measure's underlying type is `Interval` or derived from `Interval`. This story is part of [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md).

## References

- [ERO004: Connectors Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO004_connectors-analysis-phase.md)
- [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md)

## Dependencies

- [Story158_cpl-parser-connectors-time.md](Story158_cpl-parser-connectors-time.md) - Not Started
- [Story156_cpl-analyzer-connectors.md](Story156_cpl-analyzer-connectors.md) - Not Started

## Dependent Stories

- [Story160_cpl-synthesizer-connectors-time.md](Story160_cpl-synthesizer-connectors-time.md) - Not Started

## Tasks

1. Update `AnalysisConnectorDefinition` - Completed
   a. Write tests to verify `AnalysisConnectorDefinition` correctly stores the analyzed grain - Completed
   b. Update the `AnalysisConnectorDefinition` struct to include a `Grain` field - Completed

2. Update `AnalyzeConnectorsScope` to resolve grain measure - Completed
   a. Write tests for `AnalyzeConnectorsScope` to verify it correctly resolves the grain measure from the symbol table - Completed
   b. Update the `AnalyzeConnectorsScope` function to resolve the grain measure - Completed

3. Validate grain measure type - Completed
   a. Write tests to ensure `AnalyzeConnectorsScope` reports an error if the grain measure's underlying type is not `Interval` or derived from `Interval` - Completed
   b. Implement the validation logic in `AnalyzeConnectorsScope` to check the grain measure's underlying type is `Interval` or derived from `Interval`, referencing the updated `Interval` and `Instant` type definitions - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- The analyzer must correctly identify and report errors if the referenced grain measure does not exist in the symbol table.
- The analyzer must correctly identify and report errors if the referenced grain measure's underlying type is not `Interval` or derived from `Interval`.
- Grain measure type resolution depends on the updated `Instant`/`Interval` type definitions and concept-driven constraint validation (see [ERO002: Types Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md)).

## Intent

- **Task 1**: To update the internal representation of an analyzed connector to include the grain information.
- **Task 2**: To resolve the grain measure reference against the symbol table.
- **Task 3**: To enforce the semantic rule that a connector's grain must be an `Interval` or derived from `Interval` measure.

## Acceptance Criteria

- The analyzer successfully processes valid connector declarations with a `grain`.
- The analyzer reports appropriate errors for invalid grain references (e.g., missing measure or incorrect type).
- The `AnalysisConnectorDefinition` struct correctly contains the analyzed grain information.
- All unit tests pass.
