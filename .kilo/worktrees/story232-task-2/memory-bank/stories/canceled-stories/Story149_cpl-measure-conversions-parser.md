# Story149: CPL Measure Conversions Parser - Measure Conversion Clauses (Increment 5)

## Goal

Implement the parser for measure conversions including measureConversionClause grammar rule, MeasureConversionScope symbol table scopes, and pattern/calculation expression parsing. This is the final increment (5) of Epic003 units and measures parsing.

## References

- [Epic003_units-and-measures-implementation.md](../epics/Epic003_units-and-measures-implementation.md) - Parent epic for units and measures implementation
- [ERO003_units-measures-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO003_units-measures-parsing-phase.md) - Parsing phase data structures for Increment 5
- [ERO003_units-measures-incremental-implementation.md](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md) - Main incremental implementation ERO
- [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md) - Pattern for expression parsing
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md) - Architectural pattern for parsing
- [Story147_cpl-unit-conversions-parser.md](Story147_cpl-unit-conversions-parser.md) - Parser increment 4 (dependency)

### Phase Separation Note

This story implements the parsing phase. Raw syntactic data is captured into the symbol table; semantic structures are built by the analysis phase.

### Migration Reference

This story follows the ERO003 design which replaces the old UnitDecl AST approach with symbol table-driven parsing.

## Dependencies

- [Story147_cpl-unit-conversions-parser.md](Story147_cpl-unit-conversions-parser.md) - Must be completed (conversion parsing infrastructure)

## Dependent Stories

- [Story150_cpl-measure-conversions-analyzer.md](Story150_cpl-measure-conversions-analyzer.md) - Not Started

## Tasks

1. Update CPLMeasures.g4 to support measure conversion clauses - Not Started
   a. Extend measureDefinition rule to include optional measureConversionClause\* - Not Started
   b. Define measureConversionClause rule: CONVERSION_KEYWORD measureConversionPair+ - Not Started
   c. Define measureConversionPair rule: patternClause measureCalculationClause - Not Started
   d. Define measureCalculationClause rule: CALCULATION_KEYWORD expression - Not Started

   **Intent**: Add grammar for measure conversion pattern/calculation pairs.
   **Acceptance Criteria**: Grammar parses measure conversion clauses with pattern/calculation pairs.

2. Add MeasureConversionScope to ScopeType enum and MeasureConversionScopeDefinition to DefinitionType enum - Not Started
   a. Add MeasureConversionScope constant to ScopeType enum - Not Started
   b. Add MeasureConversionScopeDefinition to DefinitionType enum - Not Started

   **Intent**: Extend symbol table to recognize measure conversion scopes.
   **Acceptance Criteria**: Symbol table can create and manage measure conversion scopes.

3. Implement measure conversion listener methods in MeasuresListener - Not Started
   a. Implement EnterMeasureConversionClause to push MeasureConversionScope - Not Started
   b. Implement ExitMeasureConversionClause to pop MeasureConversionScope - Not Started
   c. Reuse pattern and calculation clause handling from unit conversions - Not Started

   **Intent**: Build listener methods that process measure conversion clauses into symbol table scopes.
   **Acceptance Criteria**: Measure conversion clauses create proper nested scopes.

4. Implement validation during parsing for Increment 5 - Not Started
   a. Validate measure conversion clause syntax - Not Started
   b. Validate pattern syntax for measure references - Not Started
   c. Validate calculation expression syntax - Not Started

   **Intent**: Ensure measure conversion clauses are syntactically valid.
   **Acceptance Criteria**: Invalid measure conversion clauses produce clear error messages.

5. Write comprehensive unit tests for measure conversions parsing - Not Started
   a. Test parsing of measure conversion with simple pattern - Not Started
   b. Test parsing of measure conversion with context-dependent pattern - Not Started
   c. Test parsing of multiple measure conversion pairs - Not Started
   d. Test error handling for invalid measure conversion syntax - Not Started
   e. Test symbol table structure for measure conversions - Not Started

   **Intent**: Verify measure conversion parsing works correctly and errors are properly reported.
   **Acceptance Criteria**: All tests pass, covering valid inputs and error conditions.

## Notes

- This story implements Increment 5 (final) of the units and measures parsing phase as defined in ERO003
- Measure conversions use same pattern/calculation pair structure as unit conversions
- Context-dependent conversions (e.g., daylight savings time) are supported
- Formula evaluation depends on metrics implementation
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
