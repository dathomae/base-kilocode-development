# Story176: CPL Connectors Multiple Grain Dimensions

## Goal

Extend the CPL grammar, parser, analysis, and synthesis phases so that connector `grain` declarations accept **multiple** Interval-derived measure dimensions (minimum one). This enables compound temporal granularities such as `grain YearGrain, WeekGrain, DayGrain`. This story is part of [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md).

## References

- [Plan: Extend Grain Declarations to Accept Multiple Interval-Derived Dimensions](../plans/grain_multiple_dimensions_plan.md)
- [ERO004: Connectors Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO004_connectors-parsing-phase.md)
- [ERO004: Connectors Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO004_connectors-analysis-phase.md)
- [ERO004: Connectors Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO004_connectors-synthesis-phase.md)
- [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md)

## Dependencies

- [Story158_cpl-parser-connectors-time.md](Story158_cpl-parser-connectors-time.md) - Done
- [Story159_cpl-analyzer-connectors-time.md](Story159_cpl-analyzer-connectors-time.md) - Done
- [Story160_cpl-synthesizer-connectors-time.md](../Story160_cpl-synthesizer-connectors-time.md) - Not Started

## Dependent Stories

None

## Tasks

1. Update grammar, symbol table, and parser listener for multiple grains - Completed
   a. Write parser tests for multiple grain identifiers and symbol table extraction - Completed
   b. Update `CPLConnectors.g4` grammar to accept `GRAIN identifierList` instead of `GRAIN identifier` - Completed
   c. Regenerate ANTLR Go parser code via `./gradlew generateGrammarSource` - Completed
   d. Update `symboltable.ConnectorDeclaration` to store `GrainNames []string` and `GrainLocations []SourceLocation` - Completed
   e. Update parser listener `ExitConnectorDeclaration` to extract multiple grain identifiers from `ctx.IdentifierList(1)` - Completed
   f. Verify all parser tests pass - Completed

2. Update analysis for multiple grain resolution and validation - Completed
   a. Write analysis tests for multiple grain resolution, empty grain lists, and Interval-derived validation - Completed
   b. Update `AnalysisConnectorDefinition.Grain` from `*AnalysisMeasureDefinition` to `[]*AnalysisMeasureDefinition` - Completed
   c. Update `AnalyzeConnectorsScope` to resolve all grain measures and reject empty grain lists - Completed
   d. Update `ValidateConnectorGrainTypes` to validate that every grain measure is Interval-derived - Completed
   e. Verify all analysis unit and integration tests pass - Completed

3. Update IO specification and synthesis for grain slices - Completed
   a. Write tests for IO spec JSON serialization with grain arrays - Completed
   b. Update `io.OutputSpecification.Grain` from `string` to `[]string` - Completed
   c. Update synthesis code in `generateOutputWritingCode` to handle a slice of grain strings - Deferred to Story160
   d. Verify all IO tests pass - Completed
   e. Synthesis multi-grain output handling deferred to Story160; minimal compilation fix in place with TODO

4. Update compiler and remaining tests - Completed
   a. Update compiler tests (`cpl_compiler_test.go`, `cpl_example_test.go`) to use grain arrays - Completed
   b. Update remaining parser multi-file tests and integration tests for grain slice assertions - Completed
   c. Verify all tests pass - Completed

5. Update CPL examples and JSON files - Completed
   a. Update all `.cpl` example files to use Interval-derived measures for grain declarations - Completed (all existing examples already valid)
   b. Add new example `multi_grain_connector.cpl` demonstrating multiple grain dimensions - Completed
   c. Update all accompanying `.json` IO spec files so `"grain": "..."` becomes `"grain": ["..."]` - Completed
   d. Verify examples compile and tests pass - Completed

6. Run full test suite and fix failures - Completed
   a. Run complete test suite via `./gradlew test` - Completed
   b. Fix any compilation or test failures arising from struct type changes - Completed (no fixes required)

7. Update requirements, design documents, and Epic004 - Completed
   a. Update `memory-bank/specifications/requirements/cpl/connectors.md` to describe multiple grain measures - Completed
   b. Update `memory-bank/specifications/requirements/cpl/io_specs.md` to describe grain arrays - Completed
   c. Update ERO004 documents (parsing, analysis, synthesis phases) to describe ordered lists of grain measures - Completed
   d. Update `memory-bank/epics/Epic004_cpl-connectors-implementation.md` to reflect multiple grain measures in constraints and acceptance criteria - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- The grammar must enforce at least one grain identifier via `identifierList` (which already requires one or more identifiers).
- Each grain measure's type must be `Interval` or derived from `Interval`.
- The order of grain measures in the declaration must be preserved in the analysis and synthesis phases.
- All existing CPL examples must continue to use valid Interval-derived measures for grain.
- Changes must be backward-compatible for single-grain declarations (they simply become single-element arrays).

## Intent

- **Task 1**: To extend the parser infrastructure so that the grammar, generated code, symbol table, and listener all support an ordered list of grain measure names.
- **Task 2**: To extend the analysis phase so that it resolves multiple grain measures, validates each is Interval-derived, and stores them as an ordered slice.
- **Task 3**: To extend the IO specification and synthesis phase so that grain is emitted and consumed as an array of strings.
- **Task 4**: To ensure all compiler-level and integration tests are updated to reflect the new grain slice types.
- **Task 5**: To ensure all CPL examples and accompanying JSON files correctly demonstrate and specify multiple grain dimensions.
- **Task 6**: To verify the entire system compiles and passes all tests after the cross-cutting changes.
- **Task 7**: To keep all requirements, design, and epic documentation consistent with the new multi-grain capability.

## Acceptance Criteria

- Grammar accepts `GRAIN identifierList` with at least one identifier.
- ANTLR Go code is regenerated and committed.
- `symboltable.ConnectorDeclaration` stores `GrainNames []string` and `GrainLocations []SourceLocation`.
- Parser listener correctly extracts multiple grain identifiers and registers each as a `GrainDefinition`.
- `analysis.AnalysisConnectorDefinition.Grain` is `[]*AnalysisMeasureDefinition`.
- `AnalyzeConnectorsScope` resolves all grain measures and rejects empty grain lists.
- `ValidateConnectorGrainTypes` validates that every grain measure is Interval-derived.
- `io.OutputSpecification.Grain` is `[]string`.
- Synthesis code handles the grain slice when generating output writing code.
- All parser, analysis, synthesis, IO, and compiler tests pass.
- All CPL examples use Interval-derived measures for grain.
- A new example demonstrates multiple grain dimensions.
- All accompanying JSON IO spec files use grain arrays.
- Requirements documents (`connectors.md`, `io_specs.md`) describe multiple grain measures.
- Design documents and EROs describe multiple grain measures.
- Epic004 is consistent with multiple grain measures.
