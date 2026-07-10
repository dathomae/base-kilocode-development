# Story137: CPL Types Parser - Instant Type (Increment 4)

## Goal

Extend the CPL parser to support the Instant base type with field-based constraints (year through second). This is Increment 4 of Epic002 types parsing, adding time-related type support.

## References

- [Epic002_cpl-types-implementation.md](../epics/Epic002_cpl-types-implementation.md) - Parent epic for types implementation
- [ERO002_types-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md) - Parsing phase data structures for Increment 4
- [ERO002_types-incremental-implementation.md](../specifications/design/cpl/eros/ERO002_types-incremental-implementation.md) - Main incremental implementation ERO
- [Story133_cpl-types-parser-string-boolean.md](Story133_cpl-types-parser-string-boolean.md) - Parser increment 2 (dependency)

## Dependencies

- [Story133_cpl-types-parser-string-boolean.md](Story133_cpl-types-parser-string-boolean.md) - Must be completed (String/Boolean base types)
- [Story135_cpl-types-parser-multiple-sections.md](Story135_cpl-types-parser-multiple-sections.md) - Should be completed (multiple sections)

## Dependent Stories

- [Story138_cpl-types-analyzer-instant.md](Story138_cpl-types-analyzer-instant.md) - Not Started
- [Story139_cpl-types-parser-interval.md](Story139_cpl-types-parser-interval.md) - Not Started

## Tasks

1. Update CPLTypes.g4 grammar to add Instant base type - Completed
   a. Add INSTANT_TYPE alternative to baseType rule - Completed
   b. Add INSTANT_TYPE lexer token for 'Instant' keyword - Completed

   **Intent**: Extend grammar to recognize Instant type keyword.
   **Acceptance Criteria**: Grammar parses type definitions using Instant base type.

2. Extend BuiltInType enum with InstantType constant - Completed
   a. Add InstantType to BuiltInType enum - Completed
   b. Update String() method to return "Instant" - Completed

   **Intent**: Add Instant base type constant.
   **Acceptance Criteria**: BuiltInType enum correctly represents Instant type.

3. Implement field-based InstantConstraints struct - Completed
   a. Define InstantConstraints with Year (*int), Month (*string), Day (*int), Hour (*int), Minute (*int), Second (*int), DayOfWeek (*string), WeekNumber (*int) fields - Completed
   b. Add GetInstantConstraints() method to ParserTypeDefinition - Completed

   **Intent**: Create data structures for Instant constraint storage.
   **Acceptance Criteria**: InstantConstraints can store all field-based constraints.

4. Update types listener methods in cpl_types_listener.go for Instant - Completed
   a. Add "Instant" case to base type parsing in EnterTypeDefinition - Completed
   b. Update parseConstraint to handle Instant-specific constraints (year, month, day, hour, minute, second, day_of_week, week_number) - Completed
   c. Parse field constraint values as integer literals or identifiers - Completed

   **Intent**: Process Instant type definitions in the parser listener.
   **Acceptance Criteria**: Instant type definitions are correctly parsed and stored in symbol table.

5. Update syntactic validation during parsing for Instant types - Completed
   a. Validate constraint names are valid for Instant (year, month, day, hour, minute, second, day_of_week, week_number) - Completed
   b. Syntactically validate field constraint formats - Completed

   **Intent**: Ensure Instant type definitions are syntactically valid.
   **Acceptance Criteria**: Invalid Instant definitions produce clear error messages.

   **Note**: Semantic validation (field constraint range validation) happens in Story138 (analysis phase).

6. Write comprehensive unit tests for Instant parsing - Completed
   a. Test Instant type with second=0 - Completed
   b. Test Instant type with year=2024 - Completed
   c. Test Instant type with day_of_week=Monday - Completed
   d. Test Instant type with multiple field constraints (e.g., month=January, day=1) - Completed
   e. Test error handling for invalid field constraint values - Completed
   f. Test error handling for unknown constraint names - Completed

   **Intent**: Verify Instant parsing works correctly and errors are properly reported.
   **Acceptance Criteria**: All tests pass, covering valid inputs and error conditions.

## Notes

- This story implements Increment 4 of the types parsing phase as defined in ERO002
- Instant type uses field-based constraints (year through second)
- No Interval type support in this increment
- **Parsing/Analysis Separation**: Field constraint range validation and duplicate detection happen in Story138 (analysis phase), not during parsing
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
