# Story139: CPL Types Parser - Interval Type (Increment 5)

## Goal

Extend the CPL parser to support the Interval base type with starts_at reference and unit constraints. This is the final increment (5) of Epic002 types parsing, completing all five built-in type support.

## References

- [Epic002_cpl-types-implementation.md](../epics/Epic002_cpl-types-implementation.md) - Parent epic for types implementation
- [ERO002_types-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md) - Parsing phase data structures for Increment 5
- [ERO002_types-incremental-implementation.md](../specifications/design/cpl/eros/ERO002_types-incremental-implementation.md) - Main incremental implementation ERO
- [Story137_cpl-types-parser-instant.md](Story137_cpl-types-parser-instant.md) - Parser increment 4 (dependency)

## Dependencies

- [Story137_cpl-types-parser-instant.md](Story137_cpl-types-parser-instant.md) - Must be completed (Instant type)

## Dependent Stories

- [Story140_cpl-types-analyzer-interval.md](Story140_cpl-types-analyzer-interval.md) - Not Started

## Tasks

1. Update CPLTypes.g4 grammar to add Interval base type - Completed
   a. Add INTERVAL_TYPE alternative to baseType rule - Completed
   b. Add INTERVAL_TYPE lexer token for 'Interval' keyword - Completed
   c. Added NUMBER_TYPE and INSTANT_TYPE tokens for consistency with existing built-in types - Completed

   **Intent**: Extend grammar to recognize Interval type keyword.
   **Acceptance Criteria**: Grammar parses type definitions using Interval base type.

2. Extend BuiltInType enum with IntervalType constant - Completed
   a. Add IntervalType to BuiltInType enum - Completed
   b. Update String() method to return "Interval" - Completed

   **Intent**: Add Interval base type constant.
   **Acceptance Criteria**: BuiltInType enum correctly represents all five built-in types.

3. Implement IntervalConstraints struct - Completed
   a. Define IntervalConstraints with StartsAt (*InstantConstraints, for inline), StartTypeName (*string, for named type reference), and UnitName (\*string) fields - Completed
   b. Add GetIntervalConstraints() method to ParserTypeDefinition - Not Applicable
   _Note: ParserTypeDefinition is not used in the current symbol-table-driven architecture. Constraint extraction belongs in the analysis phase (Story140) per ERO002._

   **Intent**: Create data structures for Interval constraint storage.
   **Acceptance Criteria**: IntervalConstraints can store inline Instant constraints, named starts_at type references, and unit references.

4. Update types listener methods in cpl_types_listener.go for Interval - Completed
   a. Add "Interval" case to base type parsing in EnterTypeDefinition - Completed
   b. Update parseConstraint to handle Interval-specific constraints (starts_at, unit) - Completed
   c. Parse starts_at constraint value as either an identifier (named type name) or inline Instant constraints - Completed
   d. Parse unit constraint value as identifier (unit name) - Completed

   **Intent**: Process Interval type definitions in the parser listener.
   **Acceptance Criteria**: Interval type definitions are correctly parsed and stored in symbol table.

5. Update syntactic validation during parsing for Interval types - Completed
   a. Syntactically validate constraint names are valid for Interval (starts_at, unit) - Completed
   b. Syntactically validate starts_at and unit constraint formats - Completed

   **Intent**: Ensure Interval type definitions are syntactically valid.
   **Acceptance Criteria**: Invalid Interval definitions produce clear error messages.

   **Note**: Semantic validation (starts_at type must be Instant-derived, unit reference validation, constraint presence checking) happens in Story140 (analysis phase).

6. Write comprehensive unit tests for Interval parsing - Completed
   a. Test Interval type with named starts_at type and unit constraints - Completed
   b. Test Interval type with inline Instant constraints for starts_at and unit constraint - Completed
   c. Test error handling for missing starts_at constraint - Completed
   d. Test error handling for missing unit constraint - Completed
   e. Test error handling for invalid constraint names - Completed

   **Intent**: Verify Interval parsing works correctly and errors are properly reported.
   **Acceptance Criteria**: All tests pass, covering valid inputs and error conditions.

## Notes

- This story implements Increment 5 (final) of the types parsing phase as defined in ERO002
- Interval type requires both starts_at (Instant-derived type name) and unit (unit name) constraints
- **Parsing/Analysis Separation**: Cross-reference validation (starts_at type must be Instant-derived), unit reference validation, and duplicate detection happen in Story140 (analysis phase), not during parsing
- This completes all five built-in type parsing support
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
