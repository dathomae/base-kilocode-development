# Story162: CPL Parser Connectors Basic

## Goal

Implement the parsing phase for basic connector declarations. This phase recognizes the `connectors` keyword and parses connector declarations, extracting the connector name, dimensions list, and celltype. This story is part of [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md).

## References

- [ERO004: Connectors Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO004_connectors-parsing-phase.md)
- [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md)

## Dependencies

- [Epic001_cpl-concepts-implementation.md](../epics/Epic001_cpl-concepts-implementation.md)
- [Epic003_units-and-measures-implementation.md](../epics/Epic003_units-and-measures-implementation.md)

## Dependent Stories

- [Story156_cpl-analyzer-connectors.md](Story156_cpl-analyzer-connectors.md) - Not Started

## Tasks

1. Update grammar for connectors section - **Completed**
   a. Write tests for parsing the `connectors` section and basic connector declarations - **Completed**
   b. Update `ChronoconePlanningLanguage.g4` to include `connectorsSection` in the `statement` rule - **Completed**
   c. Create `CPLConnectors.g4` with rules for `connectorsSection`, `connectorDeclaration`, and `identifierList` - **Completed**
   d. Update `ChronoconePlanningLanguageLexer.g4` with new keywords (`connectors`, `dimensions`, `celltype`) - **Completed**
   e. Generate updated Go parser code - **Completed**

2. Define scope and definition types - **Completed**
   a. Write tests for the new scope and definition types - **Completed**
   b. Add `ConnectorsScope` and `ConnectorScope` to `ScopeType` - **Completed**
   c. Add `ConnectorsScopeDefinition` and `ConnectorScopeDefinition` to `DefinitionType` - **Completed**

3. Define `ParserConnectorDefinition` struct - **SKIPPED**
   - This task was skipped because the symbol table-driven parsing pattern eliminates the need for parser-level definition structs. The analysis phase extracts data directly from the symbol table.

4. Implement listener methods for connectors - **Completed**
   a. Write tests for the listener methods to ensure they correctly populate the symbol table - **Completed**
   b. Implement `EnterConnectorsSection` and `ExitConnectorsSection` methods - **Completed**
   c. Implement `EnterConnectorDeclaration` and `ExitConnectorDeclaration` methods, extracting dimensions and celltype - **Completed**

5. Remove `ParserConceptDefinition` from AST package - **Completed**
   a. Remove `ParserConceptDefinition` struct from `cpl_concepts.go` - **Completed**
   b. Remove references to `ParserConceptDefinition` from tests and documentation - **Completed**
   c. Verify no functionality is broken (symbol table-driven parsing makes this struct unnecessary) - **Completed**

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- The parser must integrate with the existing ANTLR-based parser infrastructure.
- The parser populates the symbol table with connector definitions; the analysis phase extracts data directly from the symbol table (symbol table-driven parsing pattern).
- The parser only extracts the names (strings) of the dimensions and celltype. Validation is deferred to the analysis phase.

## Intent

- **Task 1**: To define the syntactic rules for connector declarations in the CPL grammar.
- **Task 2**: To provide the necessary scope and definition types for the symbol table.
- **Task 3**: ~~To provide a structured representation of a connector after parsing.~~ **SKIPPED** - Symbol table-driven parsing eliminates the need for parser-level definition structs.
- **Task 4**: To handle the ANTLR callbacks and populate the symbol table with the parsed connector definitions.
- **Task 5**: To remove the unused `ParserConceptDefinition` struct that is no longer needed with symbol table-driven parsing.

## Acceptance Criteria

- The parser successfully processes valid connector declarations.
- The parser reports appropriate syntax errors for invalid connector declarations.
- The symbol table correctly contains the parsed connectors with their dimensions and celltype.
- All unit tests pass.
