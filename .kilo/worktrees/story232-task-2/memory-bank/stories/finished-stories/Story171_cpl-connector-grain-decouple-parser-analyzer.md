# Story171: CPL Connector Grain Decouple Parser Analyzer

## Goal

Remove the `analysis -> parser` import in production code by eliminating `ParserConnectorDefinition` and expressing connector grain purely through the symbol table's `Definitions` slice. This restores compliance with the [Symbol Table-Driven Parsing Pattern](../../specifications/patterns/symbol-table-driven-parsing.md). This story is part of [Epic004: CPL Connectors Implementation](../../epics/Epic004_cpl-connectors-implementation.md).

## References

- [Plan: Fix Parser/Analyzer Coupling for Connector Grain](../../plans/fix_parser_analyzer_coupling_plan.md)
- [Symbol Table-Driven Parsing Pattern](../../specifications/patterns/symbol-table-driven-parsing.md)
- [ERO004: Connectors Analysis Phase](../../specifications/design/cpl/go-analysis/eros/ERO004_connectors-analysis-phase.md)
- [Review Comment Plan](../../plans/review_comment_plan.md)

## Dependencies

- [Story158_cpl-parser-connectors-time.md](../finished-stories/Story158_cpl-parser-connectors-time.md) - Done
- [Story159_cpl-analyzer-connectors-time.md](../finished-stories/Story159_cpl-analyzer-connectors-time.md) - Done

## Dependent Stories

None

## Tasks

1. Add `GrainDefinition` to the symbol table definition types - Completed
   a. Write tests in `symbol_table_test.go` that define a symbol with `GrainDefinition` and assert its `String()` representation is `"GrainDefinition"` - Completed
   b. Write tests that verify `DefineWithData` accepts `GrainDefinition` and stores it correctly in `Definitions` - Completed
   c. Add `GrainDefinition DefinitionType = iota` constant after `ConnectorScopeDefinition` in the `DefinitionType` enumeration - Completed
   d. Update the `String()` method on `DefinitionType` to return `"GrainDefinition"` for the new constant - Completed
   e. Run `go test ./src/main/go/symboltable/...` and verify all tests pass - Completed

2. Update the parser listener to register grain as a Definition - Completed
   a. Write tests in `cpl_parser_connectors_test.go` that assert `connectorScope.Definitions` contains a `SymbolDefinition` with `DefinitionType == symboltable.GrainDefinition` and the correct `Name` - Completed
   b. Write tests that assert a connector without grain has no `GrainDefinition` in `Definitions` - Completed
   c. Modify `ExitConnectorDeclaration` in `cpl_connectors_listener.go` to call `connectorScope.Define` with `symboltable.GrainDefinition` for the grain identifier instead of building `ParserConnectorDefinition` - Completed
   d. Remove the `connectorScope.Data = def` assignment - Completed
   e. Run `go test ./src/main/go/parser/...` and verify all tests pass - Completed

3. Remove `ParserConnectorDefinition` and update analysis - Completed
   a. Write tests in `connectors_analysis_test.go` that build test symbol tables using `symboltable.GrainDefinition` in `Definitions` instead of setting `table.Data` - Completed
   b. Write tests ensuring existing grain validation tests (undefined grain, wrong type, valid grain) exercise the same behavior with the new structure - Completed
   c. Remove the `"github.com/chronocone/chronocone_planning_language/src/main/go/parser"` import from `connectors_analysis.go` - Completed
   d. Modify `AnalyzeConnectorsScope` to scan `table.Definitions` for a `DefinitionType == symboltable.GrainDefinition` entry instead of type-asserting `table.Data` - Completed
   e. Delete `parser_connector_definition.go` - Completed
   f. Run `go test ./src/main/go/analysis/...` and `go test ./src/main/go/parser/...` to confirm the import cycle is resolved - Completed

4. Add integration test for full parse-to-analyze grain pipeline - Completed
   a. Write the `TestAnalyzeConnectorWithGrainEndToEnd` test in `analyze_connectors_integration_test.go` that calls `parser.Parse` on CPL with a connector containing `grain <Measure>`, then `analysis.Analyze`, and asserts `AnalysisConnectorDefinition.Grain` is non-nil and resolved to the correct measure - Completed
   b. Run `go test ./src/main/go/analysis/...` and verify the test passes - Completed

5. Verify import cycle is fully resolved and all tests pass - Completed
   a. Run `go test ./src/main/go/...` from `src/chronocone_planning_language` - Completed
   b. Confirm `grep -r "ParserConnectorDefinition" src/main/go/` returns zero results - Completed
   c. Confirm `grep -r '"github.com/chronocone/chronocone_planning_language/src/main/go/parser"' src/main/go/analysis/*.go` (excluding `_test.go`) returns zero results - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- The parser must only populate the symbol table; analysis must only walk it without importing parser-specific types.
- `ParserConnectorDefinition` must be completely removed from the codebase.
- The grain identifier must be registered as a `SymbolDefinition` with `DefinitionType: symboltable.GrainDefinition` in the `ConnectorScope.Definitions` slice.
- All existing connector grain behavior must be preserved.
- The import cycle between `analysis` and `parser` must be eliminated.

## Intent

- **Task 1**: To extend the symbol table with a new definition type for grain so it can be expressed declaratively in the symbol table.
- **Task 2**: To shift the parser from attaching a parser-specific struct to `Data` to registering grain as a standard symbol table definition.
- **Task 3**: To remove the `ParserConnectorDefinition` type and update the analyzer to read grain from `Definitions`, eliminating the import cycle.
- **Task 4**: To verify the end-to-end parse-to-analyze pipeline for connectors with grain works correctly after the refactor.
- **Task 5**: To confirm the architectural fix is complete and no remnants of the old coupling remain.

## Acceptance Criteria

- `connectors_analysis.go` compiles without importing `parser`.
- `go test ./src/main/go/parser/...` passes with no import cycle.
- `go test ./src/main/go/analysis/...` passes.
- `ParserConnectorDefinition` is removed from the codebase.
- At least one integration test exercises the full parse-to-analyze pipeline for a connector with `grain`.
- All tests under `./src/main/go/...` pass.
