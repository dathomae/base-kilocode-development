# Story133: CPL Types Parser - String and Boolean Types (Increment 2)

## Goal

Extend the CPL parser to support String and Boolean base types, including String constraints (min/max length, pattern) and Boolean constraints (none allowed). This is Increment 2 of Epic002 types parsing, building on the Number type foundation from Story131.

## References

- [Epic002_cpl-types-implementation.md](../epics/Epic002_cpl-types-implementation.md) - Parent epic for types implementation
- [ERO002_types-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md) - Parsing phase data structures for Increment 2
- [ERO002_types-incremental-implementation.md](../specifications/design/cpl/eros/ERO002_types-incremental-implementation.md) - Main incremental implementation ERO
- [Story131_cpl-types-parser-number.md](Story131_cpl-types-parser-number.md) - Parser increment 1 (dependency)

## Dependencies

- [Story131_cpl-types-parser-number.md](Story131_cpl-types-parser-number.md) - Must be completed (base types infrastructure)

## Dependent Stories

- [Story134_cpl-types-analyzer-string-boolean.md](Story134_cpl-types-analyzer-string-boolean.md) - Not Started
- [Story135_cpl-types-parser-multiple-sections.md](Story135_cpl-types-parser-multiple-sections.md) - Not Started

## Tasks

1. Update CPLTypes.g4 grammar to add String and Boolean base types - Not Started
   a. Add STRING_TYPE and BOOLEAN_TYPE alternatives to baseType rule - Not Started
   b. Add STRING_TYPE and BOOLEAN_TYPE lexer tokens - Not Started

   **Intent**: Extend grammar to recognize String and Boolean type keywords.
   **Acceptance Criteria**: Grammar parses type definitions using String and Boolean base types.

2. Extend BuiltInType enum with StringType and BooleanType constants - Not Started
   a. Add StringType and BooleanType to BuiltInType enum - Not Started
   b. Update String() method to return "String" and "Boolean" for new types - Not Started

   **Intent**: Add new base type constants for String and Boolean.
   **Acceptance Criteria**: BuiltInType enum correctly represents all three base types.

3. Implement StringConstraints and BooleanConstraints structs - Not Started
   a. Define StringConstraints with Minimum, Maximum (*int), and Pattern (*string) fields - Not Started
   b. Define BooleanConstraints as empty struct (no constraints allowed) - Not Started
   c. Add GetStringConstraints() method to ParserTypeDefinition - Not Started

   **Intent**: Create data structures for String and Boolean constraint storage.
   **Acceptance Criteria**: StringConstraints can store length and pattern constraints; BooleanConstraints is empty.

4. Update types listener methods in cpl_types_listener.go for String and Boolean - Not Started
   a. Add "String" and "Boolean" cases to base type parsing in EnterTypeDefinition - Not Started
   b. Update parseConstraint to handle String-specific constraints (minimum, maximum, pattern) - Not Started
   c. Syntactically validate that Boolean types have no constraints - Not Started

   **Intent**: Process String and Boolean type definitions in the parser listener.
   **Acceptance Criteria**: String and Boolean type definitions are correctly parsed and stored in symbol table.

5. Update syntactic validation during parsing for String and Boolean types - Not Started
   a. Validate String constraint names: minimum, maximum, pattern - Not Started
   b. Syntactically validate Boolean types have no constraints (report error if constraints present) - Not Started

   **Intent**: Ensure String and Boolean type definitions are syntactically valid.
   **Acceptance Criteria**: Invalid String and Boolean definitions produce clear error messages.

   **Note**: Semantic validation (constraint value type checking, duplicate detection) happens in Story134 (analysis phase).

6. Write comprehensive unit tests for String and Boolean parsing - Not Started
   a. Test String type with minimum constraint only - Not Started
   b. Test String type with maximum constraint only - Not Started
   c. Test String type with pattern constraint only - Not Started
   d. Test String type with all constraints combined - Not Started
   e. Test Boolean type with no constraints - Not Started
   f. Test error handling for Boolean with constraints - Not Started
   g. Test error handling for invalid String constraint names - Not Started

   **Intent**: Verify String and Boolean parsing works correctly and errors are properly reported.
   **Acceptance Criteria**: All tests pass, covering valid inputs and error conditions.

## Notes

- This story implements Increment 2 of the types parsing phase as defined in ERO002
- String constraints include: minimum (length), maximum (length), pattern (regex)
- Boolean type has no constraints allowed
- No Instant or Interval type support in this increment
- No multiple type sections support in this increment
- **Parsing/Analysis Separation**: Constraint value validation (integers for min/max, regex compilation for pattern) and duplicate detection happen in Story134 (analysis phase), not during parsing
- Follow Logical TDD Lifecycle: write tests before implementing code for each task

**SUPERSEDED by [Story196](Story196_cpl-types-string-boolean.md)** - This story was written against an earlier parser-centric design that was superseded by the symbol-table-driven parsing pattern. The parser already accepts String/Boolean via the generic `baseType: identifier;` rule. The remaining work has been consolidated into Story196.
