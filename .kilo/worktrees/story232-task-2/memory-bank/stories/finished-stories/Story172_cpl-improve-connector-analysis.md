# Story172: CPL Improve Connector Analysis

## Goal

Refactor connector parsing and analysis to use a structured per-connector declaration record attached to each connector scope, eliminating the fragile mixed-identifier reconstruction that currently infers connector roles by position and name equality. Move interval-family grain validation to a post-derivation pass so grain measure types are checked only after full type derivation paths are available. This story is part of [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md).

## References

- [Improve Connector Analysis Plan](../plans/improve_connector_analysis_plan.md)
- [ERO004: Connectors Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO004_connectors-analysis-phase.md)
- [Requirements: Connectors](../specifications/requirements/cpl/connectors.md)
- [Connectors Analysis](../../src/chronocone_planning_language/src/main/go/analysis/connectors_analysis.go)
- [Connectors Analysis Test](../../src/chronocone_planning_language/src/main/go/analysis/connectors_analysis_test.go)
- [Connectors Integration Test](../../src/chronocone_planning_language/src/main/go/analysis/analyze_connectors_integration_test.go)
- [Connector Parser Listener](../../src/chronocone_planning_language/src/main/go/parser/cpl_connectors_listener.go)
- [Connector Parser Test](../../src/chronocone_planning_language/src/main/go/parser/cpl_parser_connectors_test.go)

## Dependencies

- [Story171_cpl-connector-grain-decouple-parser-analyzer.md](finished-stories/Story171_cpl-connector-grain-decouple-parser-analyzer.md) - Done

## Dependent Stories

None

## Tasks

1. Create parser-level connector declaration data structure and listener population - Completed
   a. Write parser tests for structured per-connector declaration capture - Completed
   b. Define the structured record type and attach it to the connector scope without import cycles - Completed
   c. Populate the structured record in the connector listener before popping the connector scope - Completed

2. Refactor connector analysis to consume structured records - Completed
   a. Write analyzer tests for same-name dimension concept and grain measure - Completed
   b. Refactor first-pass connector analysis to resolve dimensions, grain, and celltype from the structured record only - Completed
   c. Verify duplicate connector detection and basic error handling remain correct - Completed

3. Implement deferred interval-family grain validation - Completed
   a. Write tests for deferred interval-family grain validation using finalized derivation paths - Completed
   b. Implement post-derivation connector grain validation pass that checks grain measure type against interval family - Completed
   c. Invoke the validation pass from `Analyze()` after `ComputeDerivationPaths()` and related type/unit validation - Completed

4. Remove legacy mixed-identifier connector reconstruction - Completed
   a. Write tests confirming analyzer no longer depends on generic identifier ordering for connector roles - Completed
   b. Remove or quarantine legacy mixed-identifier reconstruction code from connector analysis - Completed
   c. Update any remaining tests to assert structured record fields instead of generic identifier lists - Completed

5. Update connector analysis documentation - Completed
   a. Update ERO004 to reflect structured per-connector records attached to connector scopes and deferred grain validation - Completed
   b. Update connectors.md requirements to clarify that connectors sections are syntactic containers and grain names are measure names - Completed
   c. Verify documentation aligns with implementation - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- The structured record must contain connector name, dimension names, optional grain measure name, celltype measure name, and source locations for diagnostics.
- The record must be scoped to a single connector declaration and attached to its individual `ConnectorScope`, not to the enclosing `ConnectorsScope`.
- Compilation-wide uniqueness and aggregation must remain in the analyzer-wide connectors map maintained by `CPLAnalyzer`.
- Grain names must be resolved against the measures map, never against concepts.
- A connector must be able to legally use a dimension concept and grain measure with the same textual name.
- The interval-family grain type validation must run only after `ComputeDerivationPaths()` and related type/unit validation have completed.
- Existing valid connector programs must continue to analyze successfully.
- Invalid connector programs must still report useful errors for undefined dimensions, undefined celltype measures, undefined grain measures, duplicate connector names, and non-interval-family grain measure types.
- Avoid introducing package coupling when placing the structured declaration type; prefer the lowest package that avoids import cycles.

## Intent

- **Task 1**: To establish a clean, structured data model for connector declarations at the parser level so the analyzer no longer needs to infer roles from generic identifiers.
- **Task 2**: To eliminate the bug where a dimension concept and grain measure with the same name causes the dimension to be lost, and to make connector analysis role-explicit.
- **Task 3**: To ensure grain measure type validation uses finalized derivation paths so derived interval types are accepted correctly.
- **Task 4**: To prevent future regression to mixed-identifier reconstruction by removing or quarantining the legacy code path.
- **Task 5**: To keep design documentation aligned with the new structured-record and deferred-validation architecture.

## Acceptance Criteria

- Connector dimensions are resolved only from the structured dimensions field, never inferred by excluding identifiers whose names match the grain.
- A connector can legally use a dimension concept and grain measure with the same textual name without losing the dimension.
- Grain names are resolved against measures, not concepts.
- Grain interval-family checks use finalized derivation paths.
- Multiple connectors sections and multiple parsed inputs can contribute connector declaration scopes without requiring an aggregate parser-level connectors object.
- Existing valid connector programs continue to analyze successfully.
- Invalid connector programs still report useful errors for undefined dimensions, undefined celltype measures, undefined grain measures, duplicate connector names, and non-interval-family grain measure types.
- The analyzer no longer depends on the connector scope's generic identifier definitions for dimensions, grain, or celltype.

## Notes

- If complete removal of legacy symbol entries is too disruptive, leave them in place as parse-symbol artifacts but make the structured record the only supported analyzer input.
- The existing `GrainDefinition()` marker in the symbol table may be removed later if no longer needed, but this story can leave it in place until tests prove the structured record path covers all use cases.
- The check must validate the grain measure's type, not the grain name. Do not consult concepts for grain legality.
