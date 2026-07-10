# Story158: CPL Parser Connectors Time

## Goal

Implement support for the `grain` keyword in connector declarations. This is Increment 2 of the Connectors implementation, adding time handling. This story is part of [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md).

## References

- [ERO004: Connectors Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO004_connectors-parsing-phase.md)
- [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md)

## Dependencies

- [Story162_cpl-parser-connectors-basic.md](finished-stories/Story162_cpl-parser-connectors-basic.md) - Done
- [Story155_cpl-parser-multiple-files.md](Story155_cpl-parser-multiple-files.md) - Not Started
- [Story119_cpl-concepts-parser-multi-section.md](Story119_cpl-concepts-parser-multi-section.md) - Completed
- [Story120_cpl-concepts-analyzer-multi-section.md](Story120_cpl-concepts-analyzer-multi-section.md) - Completed

## Dependent Stories

- [Story159_cpl-analyzer-connectors-time.md](Story159_cpl-analyzer-connectors-time.md) - Not Started

## Tasks

1. Update `CPLConnectors.g4` grammar - Completed
   a. Write tests for the parser to handle the `grain` keyword in connector declarations - Completed
   b. Update the `CPLConnectors.g4` grammar file to include the `grain` keyword syntax - Completed

2. Update `ChronoconePlanningLanguageLexer.g4` - Completed
   a. Write lexer tests to ensure `GRAIN_KEYWORD` is correctly tokenized - Completed
   b. Update the `ChronoconePlanningLanguageLexer.g4` file to include the `GRAIN_KEYWORD` - Completed

3. Update `ParserConnectorDefinition` - Completed
   a. Write tests to verify `ParserConnectorDefinition` correctly stores the parsed grain - Completed
   b. Update the `ParserConnectorDefinition` struct to include a `Grain` field - Completed

4. Update `ExitConnectorDeclaration` listener method - Completed
   a. Write tests for the listener to ensure it extracts the grain from the parse tree and populates the `ParserConnectorDefinition` - Completed
   b. Update the `ExitConnectorDeclaration` listener method to extract the grain - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- The grammar updates must not break existing connector parsing logic.
- The parser must correctly identify the `grain` keyword and its associated value.
- Grammar for `grain` must remain consistent with the updated `Interval` and `Instant` type definitions (see [ERO002: Types Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md)).

## Intent

- **Task 1 & 2**: To extend the CPL grammar and lexer to recognize the `grain` keyword.
- **Task 3**: To update the internal representation of a parsed connector to include the grain information.
- **Task 4**: To extract the grain information during the parsing process and store it in the internal representation.

## Acceptance Criteria

- The parser successfully parses connector declarations that include the `grain` keyword.
- The parser correctly identifies syntax errors related to the `grain` keyword.
- The `ParserConnectorDefinition` struct correctly contains the parsed grain information.
- All unit tests pass.
