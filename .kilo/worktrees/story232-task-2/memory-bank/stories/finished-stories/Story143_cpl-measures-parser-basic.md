# Story143: CPL Measures Parser - Basic Measure Definitions (Increment 2)

## Goal

Implement the parser for basic measure definitions including measuresSection grammar, measureDefinition rule, and listener methods that push ParserMeasureDefinition to the symbol table. This is Increment 2 of Epic003 units and measures parsing, building on Story141.

## References

- [Epic003_units-and-measures-implementation.md](../epics/Epic003_units-and-measures-implementation.md) - Parent epic for units and measures implementation
- [ERO003_units-measures-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO003_units-measures-parsing-phase.md) - Parsing phase data structures for Increment 2
- [ERO003_units-measures-incremental-implementation.md](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md) - Main incremental implementation ERO
- [Story141_cpl-units-parser-basic.md](Story141_cpl-units-parser-basic.md) - Parser increment 1 (dependency)
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md) - Architectural pattern for parsing

### Phase Separation Note

This story implements the parsing phase. Raw syntactic data is captured into the symbol table; semantic structures are built by the analysis phase.

### Migration Reference

This story follows the ERO003 design which replaces the old UnitDecl AST approach with symbol table-driven parsing.

## Dependencies

- [Story141_cpl-units-parser-basic.md](Story141_cpl-units-parser-basic.md) - Must be completed (units parsing infrastructure)
- Epic002 Number type stories - Number type must be implemented (measures combine types with units)

## Dependent Stories

- [Story144_cpl-measures-analyzer-basic.md](Story144_cpl-measures-analyzer-basic.md) - Not Started
- [Story145_cpl-compound-units-parser.md](Story145_cpl-compound-units-parser.md) - Not Started

## Tasks

1. Create CPLMeasures.g4 grammar file with measuresSection and measureDefinition rules - Not Started
   a. Define measuresSection rule matching MEASURES_KEYWORD followed by one or more measureDefinition rules - Not Started
   b. Define measureDefinition rule: identifier EQ identifier LPAREN UNIT EQUALS identifier RPAREN - Not Started

   **Intent**: Establish the grammar foundation for parsing measure group definitions.
   **Acceptance Criteria**: Grammar parses valid measure definitions like `Temperature = Number(unit=TemperatureScale)` and rejects invalid syntax.

2. Add lexer tokens for measures support - Not Started
   a. Add MEASURES_KEYWORD token for 'measures' keyword - Not Started
   b. Add UNIT token for 'unit' keyword - Not Started
   c. Add LPAREN and RPAREN tokens if not already present - Not Started

   **Intent**: Provide lexical analysis for measures section keywords and punctuation.
   **Acceptance Criteria**: Lexer correctly tokenizes measures-related keywords and symbols.

3. Update main grammar to import CPLMeasures and add measuresSection to statement rule - Not Started
   a. Add CPLMeasures to grammar imports - Not Started
   b. Add measuresSection as an alternative in the statement rule - Not Started

   **Intent**: Integrate measures parsing into the main CPL grammar.
   **Acceptance Criteria**: CPL files with measures sections are accepted by the parser alongside existing sections.

4. Implement ParserMeasureDefinition struct - Not Started
   a. Define ParserMeasureDefinition with Name, TypeName, UnitGroupName, and SourceLocation fields - Not Started

   **Intent**: Create Go data structure to represent parsed measure definitions.
   **Acceptance Criteria**: ParserMeasureDefinition can store measure group name, type name, and referenced unit group name.

5. Add MeasuresScope and MeasureScope to ScopeType enum, MeasuresScopeDefinition and MeasureScopeDefinition to DefinitionType enum - Not Started
   a. Add MeasuresScope constant to ScopeType enum - Not Started
   b. Add MeasureScope constant to ScopeType enum - Not Started
   c. Add MeasuresScopeDefinition to DefinitionType enum - Not Started
   d. Add MeasureScopeDefinition to DefinitionType enum - Not Started

   **Intent**: Extend symbol table type system to recognize measures sections and measure definitions.
   **Acceptance Criteria**: Symbol table can create and manage measures scopes.

6. Implement MeasuresListener with EnterMeasuresSection, ExitMeasuresSection, and ExitMeasureDefinition methods - Not Started
   a. Create MeasuresListener struct with symbol table stack and current table fields - Not Started
   b. Implement EnterMeasuresSection to push new MeasuresScope symbol table - Not Started
   c. Implement ExitMeasuresSection to pop the measures scope - Not Started
   d. Implement ExitMeasureDefinition to parse measure name, type name, unit group name, create ParserMeasureDefinition, and add to symbol table - Not Started

   **Intent**: Build the ANTLR listener that processes measure sections and populates the symbol table.
   **Acceptance Criteria**: Listener correctly processes measure definitions and stores them in the symbol table.

7. Implement validation during parsing for Increment 2 - Not Started
   a. Validate measure group name syntax (valid identifier) - Not Started
   b. Validate type name syntax (valid identifier) - Not Started
   c. Validate unit group name syntax (valid identifier) - Not Started
   d. Detect and report duplicate measure group definitions within the same file - Not Started

   **Intent**: Ensure parsed measure definitions are syntactically valid.
   **Acceptance Criteria**: Invalid measure definitions produce clear error messages; valid definitions pass without errors.

8. Write comprehensive unit tests for measures parsing - Not Started
   a. Test parsing of single measure definition - Not Started
   b. Test parsing of multiple measure definitions in one section - Not Started
   c. Test error handling for duplicate measure group names - Not Started
   d. Test error handling for invalid identifiers - Not Started
   e. Test symbol table integration (MeasuresScope pushed/popped correctly) - Not Started

   **Intent**: Verify all parsing functionality works correctly and errors are properly reported.
   **Acceptance Criteria**: All tests pass, covering valid inputs and error conditions.

## Notes

- This story implements Increment 2 of the measures parsing phase as defined in ERO003
- Only basic measure definitions with simple unit group references (no compound units)
- Only Number type and Number-derived types supported (validated in analysis)
- Single measures section per file in this increment
- Type and unit group existence validation is deferred to analysis phase
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
