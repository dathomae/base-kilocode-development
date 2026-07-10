# Story173: CPL Mandatory Connector Grain

## Goal

Make the `grain` keyword mandatory in CPL connector declarations across the parser, analyzer, tests, and example files. This removes the temporary backward-compatibility support that allowed connectors to omit grain, ensuring all connectors explicitly declare their temporal grain as required by the CPL specification. This story is part of [Epic004: CPL Connectors Implementation](../../epics/Epic004_cpl-connectors-implementation.md).

## References

- [Plan: Make grain Mandatory in Connector Declarations](../../plans/mandatory_connector_grain_plan.md)
- [CPL Requirements for Connectors](../../specifications/requirements/cpl/connectors.md)
- [ERO004: Connectors Analysis Phase](../../specifications/design/cpl/go-analysis/eros/ERO004_connectors-analysis-phase.md)
- [Epic004: CPL Connectors Implementation](../../epics/Epic004_cpl-connectors-implementation.md)

## Dependencies

- [Story158_cpl-parser-connectors-time.md](../finished-stories/Story158_cpl-parser-connectors-time.md) - Done
- [Story159_cpl-analyzer-connectors-time.md](../finished-stories/Story159_cpl-analyzer-connectors-time.md) - Done
- [Story171_cpl-connector-grain-decouple-parser-analyzer.md](../finished-stories/Story171_cpl-connector-grain-decouple-parser-analyzer.md) - Done

## Dependent Stories

None

## Tasks

1. Update ANTLR Grammar and Regenerate Parser - Completed
   a. Change `connectorDeclaration` in `CPLConnectors.g4` to make `GRAIN identifier` required instead of optional - Completed
   b. Regenerate ANTLR parser/lexer Go code - Completed
   c. Verify the generated parser compiles - Completed

2. Update Parser Listener and Tests for Mandatory Grain - Completed
   a. Write parser tests asserting a connector without grain produces a parse error - Completed
   b. Remove backward-compatibility tests that test connectors without grain - Completed
   c. Update all remaining parser tests to include `grain <Measure>` in connector declarations - Completed
   d. Update all remaining parser tests asserting `ConnectorScope` to expect a `GrainDefinition` in `Definitions` - Completed
   e. Simplify `ExitConnectorDeclaration` in `cpl_connectors_listener.go` to remove the no-grain conditional branch - Completed
   f. Run `go test ./src/main/go/parser/...` and verify all tests pass - Completed

3. Update Analysis to Enforce Mandatory Grain - Completed
   a. Write analysis tests asserting a connector with valid grain resolves correctly - Completed
   b. Write analysis tests asserting a connector missing grain returns an error - Completed
   c. Write analysis tests asserting an undefined grain measure returns an error - Completed
   d. Write analysis tests asserting a grain measure whose type is not Interval returns an error - Completed
   e. Update `AnalyzeConnectorsScope` to require at least one `GrainDefinition` in `Definitions` - Completed
   f. Update `AnalyzeConnectorsScope` to resolve and validate the grain measure (exists, is Interval or derived from Interval) - Completed
   g. Update `AnalyzeConnectorsScope` to store the resolved grain in `AnalysisConnectorDefinition.Grain` - Completed
   h. Run `go test ./src/main/go/analysis/...` and verify all tests pass - Completed

4. Update Example CPL Files and Compiler Tests - Completed
   a. Add `grain <Measure>` clause to every connector declaration in example files under `examples/without_errors/` - Completed
   b. Add `grain <Measure>` clause to every connector declaration in example files under `examples/with_errors/` (where the error is not the missing grain itself) - Completed
   c. Update compiler and CLI tests to include grain in all connector declarations - Completed
   d. Update analysis integration tests to include grain and assert `AnalysisConnectorDefinition.Grain` is populated - Completed
   e. Run `go test ./src/main/go/cli/compiler/...` and verify tests pass - Completed

5. Update Epic Documentation and Run Full Test Suite - Completed
   a. Update `Epic004_cpl-connectors-implementation.md` to state grain is mandatory - Completed
   b. Run `go test ./...` from `src/chronocone_planning_language` - Completed
   c. Verify all parser, analysis, compiler, and integration tests pass - Completed
   d. Verify all example files parse and analyze successfully - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- The `grain` keyword must be syntactically mandatory in the grammar; omitting it must produce a parse error.
- The analysis phase must reject connectors that parse successfully but lack a `GrainDefinition` in `Definitions` (defense in depth).
- All existing example files with connectors must continue to be valid after adding grain.
- The `standard_definitions/standard_measures.cpl` file already defines Interval measures (`Daily`, `Weekly`) that can be referenced as grain values.

## Intent

- **Task 1**: To make grain a required syntactic element in the connector grammar.
- **Task 2**: To update the parser and its tests so that missing grain is a parse error and the listener no longer handles optional grain.
- **Task 3**: To update the analyzer and its tests so that grain is validated semantically (exists, is Interval-derived) and stored in the analysis result.
- **Task 4**: To ensure all examples and compiler tests reflect the mandatory grain requirement.
- **Task 5**: To confirm the entire system compiles, passes all tests, and documentation is accurate.

## Acceptance Criteria

- A connector declaration without `grain` produces a parse error.
- The parser no longer treats grain as optional in the grammar or listener.
- The analysis phase validates the grain measure exists and is `Interval` or derived from `Interval`.
- All parser and analysis tests use connectors with grain and assert correct behavior.
- All `.cpl` example files include grain in every connector declaration.
- All integration and compiler tests that construct CPL strings with connectors include grain.
- `go test ./...` passes with no failures.
